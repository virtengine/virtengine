/**
 * agent-endpoint.mjs — Lightweight HTTP server for agent self-reporting
 *
 * Agents running in worktrees use this REST API to tell the orchestrator
 * "I'm done" / "I hit an error" / "I'm still alive" without polling.
 *
 * Features:
 *   - Node.js built-in `http.createServer` — zero external dependencies
 *   - Binds to 127.0.0.1 on configurable port (AGENT_ENDPOINT_PORT or 18432)
 *   - JSON request/response, CORS for localhost
 *   - 30s request timeout, 1MB max body
 *   - Callback hooks for monitor integration
 *
 * EXPORTS:
 *   AgentEndpoint          — Main class
 *   createAgentEndpoint()  — Factory function
 */

import { createServer } from "node:http";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import { writeFileSync, mkdirSync, unlinkSync } from "node:fs";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const TAG = "[agent-endpoint]";

const DEFAULT_PORT = 18432;
const MAX_BODY_SIZE = 1024 * 1024; // 1 MB
const REQUEST_TIMEOUT_MS = 30_000; // 30 seconds

// Valid status transitions when an agent self-reports
const VALID_TRANSITIONS = {
  inprogress: ["inreview", "blocked", "done"],
  inreview: ["done"],
};

// ── Helpers ─────────────────────────────────────────────────────────────────

/**
 * Parse JSON body from an incoming request with size limit.
 * @param {import("node:http").IncomingMessage} req
 * @returns {Promise<object>}
 */
function parseBody(req) {
  return new Promise((resolve, reject) => {
    const chunks = [];
    let size = 0;

    req.on("data", (chunk) => {
      size += chunk.length;
      if (size > MAX_BODY_SIZE) {
        req.destroy();
        reject(new Error("Request body too large"));
        return;
      }
      chunks.push(chunk);
    });

    req.on("end", () => {
      const raw = Buffer.concat(chunks).toString("utf8");
      if (!raw || raw.trim() === "") {
        resolve({});
        return;
      }
      try {
        resolve(JSON.parse(raw));
      } catch (err) {
        // Include preview of malformed JSON for debugging (truncate to 200 chars)
        const preview = raw.length > 200 ? raw.slice(0, 200) + "..." : raw;
        reject(new Error(`Invalid JSON body: ${err.message} — Preview: ${preview}`));
      }
    });

    req.on("error", reject);
  });
}

/**
 * Send a JSON response.
 * @param {import("node:http").ServerResponse} res
 * @param {number} status
 * @param {object} body
 */
function sendJson(res, status, body) {
  const payload = JSON.stringify(body);
  res.writeHead(status, {
    "Content-Type": "application/json",
    "Access-Control-Allow-Origin": "http://localhost",
    "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
    "Access-Control-Allow-Headers": "Content-Type",
    "Cache-Control": "no-store",
  });
  res.end(payload);
}

/**
 * Extract a task ID from a URL pathname like /api/tasks/:id/...
 * @param {string} pathname
 * @returns {string|null}
 */
function extractTaskId(pathname) {
  const match = pathname.match(/^\/api\/tasks\/([^/]+)/);
  return match ? match[1] : null;
}

// ── AgentEndpoint Class ─────────────────────────────────────────────────────

export class AgentEndpoint {
  /**
   * @param {object} options
   * @param {number}   [options.port]            — Listen port (default: env or 18432)
   * @param {object}   [options.taskStore]        — Task store instance (kanban adapter)
   * @param {Function} [options.onTaskComplete]   — (taskId, data) => void
   * @param {Function} [options.onTaskError]      — (taskId, data) => void
   * @param {Function} [options.onStatusChange]   — (taskId, newStatus, source) => void
   */
  constructor(options = {}) {
    this._port =
      options.port ||
      (process.env.AGENT_ENDPOINT_PORT
        ? Number(process.env.AGENT_ENDPOINT_PORT)
        : DEFAULT_PORT);
    this._taskStore = options.taskStore || null;
    this._onTaskComplete = options.onTaskComplete || null;
    this._onTaskError = options.onTaskError || null;
    this._onStatusChange = options.onStatusChange || null;
    this._server = null;
    this._running = false;
    this._startedAt = null;
    this._portFilePath = resolve(__dirname, ".cache", "agent-endpoint-port");
  }

  // ── Lifecycle ───────────────────────────────────────────────────────────

  /**
   * Start the HTTP server.
   * @returns {Promise<void>}
   */
  async start() {
    if (this._running) return;

    const MAX_PORT_RETRIES = 5;
    let lastErr;

    for (let attempt = 0; attempt < MAX_PORT_RETRIES; attempt++) {
      const port = this._port + attempt;
      try {
        await this._tryListen(port);
        this._port = port; // update in case we incremented
        return;
      } catch (err) {
        lastErr = err;
        if (err.code === "EADDRINUSE") {
          console.warn(
            `${TAG} Port ${port} in use (attempt ${attempt + 1}/${MAX_PORT_RETRIES}), trying to free it...`,
          );
          // Try to kill the process holding the port (Windows)
          await this._killProcessOnPort(port);
          // Retry same port once after kill
          try {
            await this._tryListen(port);
            this._port = port;
            return;
          } catch (retryErr) {
            if (retryErr.code === "EADDRINUSE") {
              console.warn(
                `${TAG} Port ${port} still in use after kill, trying next port`,
              );
              continue;
            }
            throw retryErr;
          }
        }
        throw err;
      }
    }

    // All retries exhausted — start without endpoint (non-fatal)
    console.error(
      `${TAG} Could not bind to any port after ${MAX_PORT_RETRIES} attempts: ${lastErr?.message}`,
    );
    console.warn(
      `${TAG} Running WITHOUT agent endpoint — agents can still work via poll-based completion`,
    );
  }

  /**
   * Attempt to listen on a specific port. Returns a promise.
   * @param {number} port
   * @returns {Promise<void>}
   */
  _tryListen(port) {
    return new Promise((resolveStart, rejectStart) => {
      const server = createServer((req, res) => this._handleRequest(req, res));
      server.setTimeout(REQUEST_TIMEOUT_MS);

      server.on("timeout", (socket) => {
        console.log(`${TAG} Request timed out, destroying socket`);
        socket.destroy();
      });

      server.on("error", (err) => {
        if (!this._running) {
          rejectStart(err);
        } else {
          console.error(`${TAG} Server error:`, err.message);
        }
      });

      server.listen(port, "127.0.0.1", () => {
        this._server = server;
        this._running = true;
        this._startedAt = Date.now();
        console.log(`${TAG} Listening on 127.0.0.1:${port}`);
        this._writePortFile();
        resolveStart();
      });
    });
  }

  /**
   * Attempt to kill whatever process is holding a port (Windows netstat+taskkill).
   * @param {number} port
   * @returns {Promise<void>}
   */
  async _killProcessOnPort(port) {
    try {
      const { execSync } = await import("node:child_process");
      // Find PID holding the port on Windows
      const output = execSync(`netstat -ano | findstr ":${port}"`, {
        encoding: "utf8",
        timeout: 5000,
      }).trim();
      const lines = output.split("\n").filter((l) => l.includes("LISTENING"));
      const pids = new Set();
      for (const line of lines) {
        const parts = line.trim().split(/\s+/);
        const pid = parts[parts.length - 1];
        if (pid && /^\d+$/.test(pid) && pid !== String(process.pid)) {
          pids.add(pid);
        }
      }
      for (const pid of pids) {
        console.log(`${TAG} Killing stale process PID ${pid} on port ${port}`);
        try {
          execSync(`taskkill /F /PID ${pid}`, {
            encoding: "utf8",
            timeout: 5000,
          });
        } catch (killErr) {
          /* may already be dead — log for diagnostics */
          console.warn(
            `${TAG} taskkill PID ${pid} failed: ${killErr.stderr?.trim() || killErr.message || "unknown error"}`,
          );
        }
      }
      // Give OS time to release the port
      await new Promise((r) => setTimeout(r, 1000));
    } catch (outerErr) {
      // netstat/taskkill may fail on non-Windows or if port already free
      if (outerErr.status !== 1) {
        // status 1 = no matching netstat entries (port already free)
        console.warn(
          `${TAG} _killProcessOnPort(${port}) failed: ${outerErr.message || "unknown error"}`,
        );
      }
    }
  }

  /**
   * Stop the HTTP server.
   * @returns {Promise<void>}
   */
  stop() {
    if (!this._running || !this._server) return Promise.resolve();

    return new Promise((resolveStop) => {
      this._running = false;
      this._removePortFile();
      this._server.close(() => {
        console.log(`${TAG} Server stopped`);
        resolveStop();
      });
      // Force-close lingering connections after 5s
      setTimeout(() => {
        resolveStop();
      }, 5000);
    });
  }

  /** @returns {number} */
  getPort() {
    return this._port;
  }

  /** @returns {boolean} */
  isRunning() {
    return this._running;
  }

  /**
   * Lightweight status for diagnostics (/agents).
   * @returns {{ running: boolean, port: number, startedAt: number|null, uptimeMs: number }}
   */
  getStatus() {
    return {
      running: this._running,
      port: this._port,
      startedAt: this._startedAt || null,
      uptimeMs:
        this._running && this._startedAt
          ? Math.max(0, Date.now() - this._startedAt)
          : 0,
    };
  }

  // ── Port Discovery File ─────────────────────────────────────────────────

  _writePortFile() {
    try {
      const dir = dirname(this._portFilePath);
      mkdirSync(dir, { recursive: true });
      writeFileSync(this._portFilePath, String(this._port));
      console.log(
        `${TAG} Port file written: ${this._portFilePath} → ${this._port}`,
      );
    } catch (err) {
      console.error(`${TAG} Failed to write port file:`, err.message);
    }
  }

  _removePortFile() {
    try {
      unlinkSync(this._portFilePath);
    } catch {
      // Ignore — file may already be gone
    }
  }

  // ── Request Router ──────────────────────────────────────────────────────

  /**
   * @param {import("node:http").IncomingMessage} req
   * @param {import("node:http").ServerResponse} res
   */
  async _handleRequest(req, res) {
    // Handle CORS preflight
    if (req.method === "OPTIONS") {
      sendJson(res, 204, {});
      return;
    }

    const url = new URL(req.url, `http://${req.headers.host || "localhost"}`);
    const pathname = url.pathname;
    const method = req.method;

    try {
      // ── Static routes ───────────────────────────────────────────────
      if (method === "GET" && pathname === "/health") {
        return this._handleHealth(res);
      }

      if (method === "GET" && pathname === "/api/status") {
        return this._handleStatus(res);
      }

      if (method === "GET" && pathname === "/api/tasks") {
        return await this._handleListTasks(url, res);
      }

      // ── Task-specific routes ────────────────────────────────────────
      const taskId = extractTaskId(pathname);

      if (taskId) {
        if (method === "GET" && pathname === `/api/tasks/${taskId}`) {
          return await this._handleGetTask(taskId, res);
        }

        if (method === "POST" && pathname === `/api/tasks/${taskId}/status`) {
          const body = await parseBody(req);
          return await this._handleStatusChange(taskId, body, res);
        }

        if (
          method === "POST" &&
          pathname === `/api/tasks/${taskId}/heartbeat`
        ) {
          const body = await parseBody(req);
          return await this._handleHeartbeat(taskId, body, res);
        }

        if (method === "POST" && pathname === `/api/tasks/${taskId}/complete`) {
          const body = await parseBody(req);
          return await this._handleComplete(taskId, body, res);
        }

        if (method === "POST" && pathname === `/api/tasks/${taskId}/error`) {
          const body = await parseBody(req);
          return await this._handleError(taskId, body, res);
        }
      }

      // ── 404 ─────────────────────────────────────────────────────────
      sendJson(res, 404, { error: "Not found", path: pathname });
    } catch (err) {
      console.error(`${TAG} ${method} ${pathname} error:`, err.message);
      sendJson(res, 500, { error: err.message || "Internal server error" });
    }
  }

  // ── Route Handlers ──────────────────────────────────────────────────────

  _handleHealth(res) {
    const uptimeSeconds =
      this._startedAt != null
        ? Math.floor((Date.now() - this._startedAt) / 1000)
        : 0;
    sendJson(res, 200, { ok: true, uptime: uptimeSeconds });
  }

  _handleStatus(res) {
    const uptimeSeconds =
      this._startedAt != null
        ? Math.floor((Date.now() - this._startedAt) / 1000)
        : 0;
    const storeStats = this._taskStore
      ? { connected: true }
      : { connected: false };

    sendJson(res, 200, {
      executor: { running: this._running, port: this._port },
      store: storeStats,
      uptime: uptimeSeconds,
    });
  }

  async _handleListTasks(url, res) {
    if (!this._taskStore) {
      sendJson(res, 503, { error: "Task store not configured" });
      return;
    }

    const statusFilter = url.searchParams.get("status") || undefined;

    try {
      let tasks;
      if (typeof this._taskStore.listTasks === "function") {
        // kanban adapter-style store
        tasks = await this._taskStore.listTasks(null, { status: statusFilter });
      } else if (typeof this._taskStore.list === "function") {
        tasks = await this._taskStore.list({ status: statusFilter });
      } else {
        sendJson(res, 501, { error: "Task store does not support listing" });
        return;
      }

      const taskList = Array.isArray(tasks) ? tasks : [];
      sendJson(res, 200, { tasks: taskList, count: taskList.length });
    } catch (err) {
      console.error(`${TAG} listTasks error:`, err.message);
      sendJson(res, 500, { error: `Failed to list tasks: ${err.message}` });
    }
  }

  async _handleGetTask(taskId, res) {
    if (!this._taskStore) {
      sendJson(res, 503, { error: "Task store not configured" });
      return;
    }

    try {
      let task;
      if (typeof this._taskStore.getTask === "function") {
        task = await this._taskStore.getTask(taskId);
      } else if (typeof this._taskStore.get === "function") {
        task = await this._taskStore.get(taskId);
      } else {
        sendJson(res, 501, { error: "Task store does not support get" });
        return;
      }

      if (!task) {
        sendJson(res, 404, { error: "Task not found" });
        return;
      }

      sendJson(res, 200, { task });
    } catch (err) {
      console.error(`${TAG} getTask(${taskId}) error:`, err.message);
      sendJson(res, 404, { error: "Task not found" });
    }
  }

  async _handleStatusChange(taskId, body, res) {
    if (!this._taskStore) {
      sendJson(res, 503, { error: "Task store not configured" });
      return;
    }

    const { status, message } = body;
    if (!status) {
      sendJson(res, 400, { error: "Missing 'status' in body" });
      return;
    }

    const allowed = ["inreview", "done", "blocked"];
    if (!allowed.includes(status)) {
      sendJson(res, 400, {
        error: `Invalid status '${status}'. Allowed: ${allowed.join(", ")}`,
      });
      return;
    }

    // Validate transition
    try {
      let currentTask;
      if (typeof this._taskStore.getTask === "function") {
        currentTask = await this._taskStore.getTask(taskId);
      } else if (typeof this._taskStore.get === "function") {
        currentTask = await this._taskStore.get(taskId);
      }

      if (currentTask) {
        const currentStatus = currentTask.status || "unknown";
        const validNext = VALID_TRANSITIONS[currentStatus];
        if (validNext && !validNext.includes(status)) {
          sendJson(res, 409, {
            error: `Invalid transition: ${currentStatus} → ${status}. Allowed: ${validNext.join(", ")}`,
          });
          return;
        }
      }
    } catch {
      // If we can't fetch current task, proceed anyway
    }

    try {
      let updatedTask;
      if (typeof this._taskStore.updateTaskStatus === "function") {
        updatedTask = await this._taskStore.updateTaskStatus(taskId, status);
      } else if (typeof this._taskStore.update === "function") {
        updatedTask = await this._taskStore.update(taskId, { status });
      }

      console.log(
        `${TAG} Task ${taskId} status → ${status} (source=agent)${message ? ` msg="${message}"` : ""}`,
      );

      if (this._onStatusChange) {
        try {
          await this._onStatusChange(taskId, status, "agent");
        } catch (err) {
          console.error(`${TAG} onStatusChange callback error:`, err.message);
        }
      }

      sendJson(res, 200, {
        ok: true,
        task: updatedTask || { id: taskId, status },
      });
    } catch (err) {
      console.error(`${TAG} statusChange(${taskId}) error:`, err.message);
      sendJson(res, 500, { error: `Failed to update status: ${err.message}` });
    }
  }

  async _handleHeartbeat(taskId, body, res) {
    const timestamp = new Date().toISOString();
    const { message } = body;

    console.log(
      `${TAG} Heartbeat from task ${taskId}${message ? `: ${message}` : ""}`,
    );

    // Try to update lastActivityAt on the task if the store supports it
    if (this._taskStore) {
      try {
        if (typeof this._taskStore.update === "function") {
          await this._taskStore.update(taskId, { lastActivityAt: timestamp });
        } else if (typeof this._taskStore.updateTaskStatus === "function") {
          // kanban adapter doesn't have a generic update, but heartbeat is still recorded
        }
      } catch {
        // Non-critical — heartbeat is logged regardless
      }
    }

    sendJson(res, 200, { ok: true, timestamp });
  }

  async _handleComplete(taskId, body, res) {
    const { hasCommits, branch, prUrl, output } = body;

    console.log(
      `${TAG} Task ${taskId} complete: hasCommits=${!!hasCommits}, branch=${branch || "none"}, pr=${prUrl || "none"}`,
    );

    let nextAction = "cooldown";

    if (hasCommits) {
      nextAction = "review";

      // Update task status to inreview
      if (this._taskStore) {
        try {
          if (typeof this._taskStore.updateTaskStatus === "function") {
            await this._taskStore.updateTaskStatus(taskId, "inreview");
          } else if (typeof this._taskStore.update === "function") {
            await this._taskStore.update(taskId, { status: "inreview" });
          }
        } catch (err) {
          console.error(
            `${TAG} Failed to set task ${taskId} to inreview:`,
            err.message,
          );
        }
      }
    } else {
      // No commits — record the attempt but don't change status
      console.log(`${TAG} Task ${taskId} completed with no commits`);
    }

    // Fire callback
    if (this._onTaskComplete) {
      try {
        await this._onTaskComplete(taskId, {
          hasCommits,
          branch,
          prUrl,
          output,
        });
      } catch (err) {
        console.error(`${TAG} onTaskComplete callback error:`, err.message);
      }
    }

    // Retrieve updated task for response
    let task = { id: taskId };
    if (this._taskStore) {
      try {
        if (typeof this._taskStore.getTask === "function") {
          task = (await this._taskStore.getTask(taskId)) || task;
        } else if (typeof this._taskStore.get === "function") {
          task = (await this._taskStore.get(taskId)) || task;
        }
      } catch {
        // Use fallback
      }
    }

    sendJson(res, 200, { ok: true, task, nextAction });
  }

  async _handleError(taskId, body, res) {
    const { error: errorMsg, pattern } = body;

    if (!errorMsg) {
      sendJson(res, 400, { error: "Missing 'error' in body" });
      return;
    }

    const validPatterns = [
      "plan_stuck",
      "rate_limit",
      "token_overflow",
      "api_error",
    ];
    if (pattern && !validPatterns.includes(pattern)) {
      console.log(
        `${TAG} Task ${taskId} error with unknown pattern '${pattern}': ${errorMsg}`,
      );
    } else {
      console.log(
        `${TAG} Task ${taskId} error${pattern ? ` (${pattern})` : ""}: ${errorMsg}`,
      );
    }

    // Determine action based on pattern
    let action = "retry";
    if (pattern === "rate_limit") {
      action = "cooldown";
    } else if (pattern === "token_overflow") {
      action = "blocked";
    } else if (pattern === "plan_stuck") {
      action = "retry";
    } else if (pattern === "api_error") {
      action = "cooldown";
    }

    // Fire callback
    if (this._onTaskError) {
      try {
        await this._onTaskError(taskId, { error: errorMsg, pattern });
      } catch (err) {
        console.error(`${TAG} onTaskError callback error:`, err.message);
      }
    }

    sendJson(res, 200, { ok: true, action });
  }
}

// ── Factory ─────────────────────────────────────────────────────────────────

/**
 * Create an AgentEndpoint instance.
 * @param {object} [options] — Same as AgentEndpoint constructor
 * @returns {AgentEndpoint}
 */
export function createAgentEndpoint(options) {
  return new AgentEndpoint(options);
}
