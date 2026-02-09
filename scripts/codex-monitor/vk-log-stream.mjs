/**
 * vk-log-stream.mjs — Real-time agent log capture from Vibe-Kanban WebSocket.
 *
 * Connects to VK's execution-process raw-logs WebSocket endpoints to capture
 * agent stdout/stderr that would otherwise be invisible to the monitor.
 *
 * VK Architecture (reverse-engineered from BloopAI/vibe-kanban):
 *   - Each agent session creates an "execution process" inside VK
 *   - Agent stdout/stderr → MsgStore (in-memory broadcast channel)
 *   - MsgStore → raw-logs WebSocket endpoint for live streaming
 *   - MsgStore → SQLite (JSONL) for persistence
 *
 * VK API endpoints used:
 *   GET /api/execution-processes/{id}                     — get single process (REST)
 *   GET /api/execution-processes/{id}/raw-logs/ws         — raw log stream (WebSocket)
 *   GET /api/execution-processes/stream/session/ws?session_id=X — process discovery per session (WebSocket, JSON Patch)
 *
 * Note: There is NO list-all endpoint (GET /api/execution-processes). Discovery
 * must go through the session-based WebSocket stream or direct connectToProcess() calls.
 *
 * WebSocket message format (LogMsg enum variants):
 *   {"Stdout": "line of text"}
 *   {"Stderr": "line of text"}
 *   {"JsonPatch": [{"op": "add", "path": "...", "value": {...}}]}
 *   {"Finished": ""}
 *
 * Usage:
 *   import { VkLogStream } from "./vk-log-stream.mjs";
 *   const stream = new VkLogStream(vkEndpointUrl, { logDir, onLine });
 *   stream.start();
 *   stream.connectToSession(sessionId);     // discover processes via session stream
 *   stream.connectToProcess(processId);     // direct connection to known process
 *   // ... later
 *   stream.stop();
 */

import { appendFileSync, existsSync, mkdirSync } from "node:fs";
import { resolve } from "node:path";

// ── Configuration defaults ──────────────────────────────────────────────────
const RECONNECT_DELAY_MS = 3000; // Delay before reconnecting a dropped WebSocket
const MAX_RECONNECT_ATTEMPTS = 10; // Max consecutive reconnect failures per process
const SESSION_RECONNECT_DELAY_MS = 5000; // Delay before reconnecting a dropped session stream
const MAX_SESSION_RECONNECT_ATTEMPTS = 15; // Max consecutive reconnect failures per session

/**
 * VkLogStream - Captures real-time agent logs from VK execution processes.
 *
 * Discovery model:
 *   - No REST list endpoint exists; discovery uses the session-based WebSocket
 *     (stream/session/ws) which pushes JSON Patch updates as processes start/stop.
 *   - Monitor calls connectToSession(sessionId) when sessions are created.
 *   - Monitor calls connectToProcess(processId) when a specific process ID is known.
 */
export class VkLogStream {
  /** @type {string} VK API base URL (e.g. http://192.168.0.161:54089) */
  #baseUrl;

  /** @type {string} WebSocket base URL (ws:// or wss://) */
  #wsBaseUrl;

  /** @type {string} Directory to write per-process log files */
  #logDir;

  /** @type {boolean} Whether to echo log lines to console */
  #echo;

  /** @type {((line: string, meta: {processId: string, stream: string}) => void)|null} */
  #onLine;

  /** @type {Map<string, WebSocket>} Active raw-log WebSocket connections by process ID */
  #connections = new Map();

  /** @type {Map<string, number>} Reconnect attempt counts by process ID */
  #reconnectCounts = new Map();

  /** @type {Set<string>} Process IDs that have sent Finished */
  #finished = new Set();

  /** @type {boolean} Whether the stream is running */
  #running = false;

  /** @type {Set<string>} Known process IDs (to avoid re-connecting) */
  #knownProcessIds = new Set();

  /** @type {Map<string, WebSocket>} Session stream WebSocket connections by session ID */
  #sessionStreams = new Map();

  /** @type {Map<string, number>} Session stream reconnect attempt counts */
  #sessionReconnectCounts = new Map();

  /** @type {Set<string>} Session IDs we're tracking */
  #trackedSessions = new Set();

  /**
   * @param {string} vkEndpointUrl - VK API base URL (e.g. http://127.0.0.1:54089)
   * @param {object} [opts]
   * @param {string} [opts.logDir] - Directory for per-process log files
   * @param {boolean} [opts.echo=false] - Echo log lines to console
   * @param {(line: string, meta: {processId: string, stream: string}) => void} [opts.onLine] - Callback per log line
   */
  constructor(vkEndpointUrl, opts = {}) {
    this.#baseUrl = vkEndpointUrl.replace(/\/+$/, "");
    this.#wsBaseUrl = this.#baseUrl
      .replace(/^http:/, "ws:")
      .replace(/^https:/, "wss:");
    this.#logDir = opts.logDir || null;
    this.#echo = opts.echo || false;
    this.#onLine = opts.onLine || null;

    if (this.#logDir) {
      try {
        mkdirSync(this.#logDir, { recursive: true });
      } catch {
        /* best effort */
      }
    }
  }

  /**
   * Start the log stream manager (enables connections, no automatic polling).
   * Call connectToSession() or connectToProcess() to actually capture logs.
   */
  start() {
    if (this.#running) return;
    this.#running = true;
    console.log(
      `[vk-log-stream] started — ready for session/process connections (${this.#baseUrl})`,
    );
  }

  /**
   * Stop all connections (raw-log streams + session streams).
   */
  stop() {
    if (!this.#running) return;
    this.#running = false;

    // Close session stream WebSockets
    for (const [id, ws] of this.#sessionStreams) {
      try {
        ws.close(1000, "monitor shutdown");
      } catch {
        /* best effort */
      }
    }
    this.#sessionStreams.clear();
    this.#sessionReconnectCounts.clear();
    this.#trackedSessions.clear();

    // Close raw-log WebSockets
    for (const [id, ws] of this.#connections) {
      try {
        ws.close(1000, "monitor shutdown");
      } catch {
        /* best effort */
      }
    }
    this.#connections.clear();
    this.#reconnectCounts.clear();
    console.log("[vk-log-stream] stopped");
  }

  /**
   * Connect to a session's execution-process stream to auto-discover processes.
   * VK endpoint: GET /api/execution-processes/stream/session/ws?session_id=X
   *
   * The server pushes JSON Patch updates as processes are created, updated, or
   * completed. This method parses those patches to automatically connect to
   * each running process's raw-logs WebSocket.
   *
   * @param {string} sessionId - The VK session UUID
   */
  connectToSession(sessionId) {
    if (!sessionId || !this.#running) return;
    if (this.#sessionStreams.has(sessionId)) return; // already connected
    this.#trackedSessions.add(sessionId);
    this.#openSessionStream(sessionId);
  }

  /**
   * Connect to a specific execution process's raw-logs WebSocket.
   * @param {string} processId - The execution process UUID
   * @param {object} [meta] - Optional metadata (task_id, branch, etc.)
   */
  connectToProcess(processId, meta = {}) {
    if (!processId || !this.#running) return;
    if (this.#connections.has(processId) || this.#finished.has(processId)) {
      return;
    }
    this.#knownProcessIds.add(processId);
    this.#connectWebSocket(processId, meta);
  }

  /**
   * Get the set of currently connected process IDs.
   * @returns {Set<string>}
   */
  getActiveConnections() {
    return new Set(this.#connections.keys());
  }

  /**
   * Get stream statistics.
   * @returns {{ active: number, finished: number, known: number, sessions: number }}
   */
  getStats() {
    return {
      active: this.#connections.size,
      finished: this.#finished.size,
      known: this.#knownProcessIds.size,
      sessions: this.#sessionStreams.size,
    };
  }

  // ── Private methods ────────────────────────────────────────────────────────

  /**
   * Open the session-based execution-process stream WebSocket.
   * Receives JSON Patch updates for all processes in the session.
   *
   * Initial snapshot: {"op":"replace","path":"/execution_processes","value":{...}}
   * Live updates:     {"op":"add|replace|remove","path":"/execution_processes/<id>","value":{...}}
   *
   * @param {string} sessionId
   */
  #openSessionStream(sessionId) {
    const shortSid = sessionId.slice(0, 8);
    const params = new URLSearchParams({ session_id: sessionId });
    const wsUrl = `${this.#wsBaseUrl}/api/execution-processes/stream/session/ws?${params}`;

    let ws;
    try {
      ws = new WebSocket(wsUrl);
    } catch (err) {
      console.warn(
        `[vk-log-stream] failed to create session stream WS for ${shortSid}: ${err.message}`,
      );
      return;
    }

    this.#sessionStreams.set(sessionId, ws);
    this.#sessionReconnectCounts.set(sessionId, 0);

    ws.addEventListener("open", () => {
      console.log(`[vk-log-stream] session stream connected (${shortSid})`);
      this.#sessionReconnectCounts.set(sessionId, 0);
    });

    ws.addEventListener("message", (event) => {
      this.#handleSessionStreamMessage(sessionId, event.data);
    });

    ws.addEventListener("close", () => {
      this.#sessionStreams.delete(sessionId);
      if (!this.#running || !this.#trackedSessions.has(sessionId)) return;

      const attempts =
        (this.#sessionReconnectCounts.get(sessionId) || 0) + 1;
      this.#sessionReconnectCounts.set(sessionId, attempts);

      if (attempts > MAX_SESSION_RECONNECT_ATTEMPTS) {
        console.warn(
          `[vk-log-stream] session stream ${shortSid} max reconnects (${MAX_SESSION_RECONNECT_ATTEMPTS}) reached`,
        );
        this.#trackedSessions.delete(sessionId);
        return;
      }

      const delay = Math.min(
        SESSION_RECONNECT_DELAY_MS * Math.pow(1.5, attempts - 1),
        60000,
      );
      setTimeout(() => {
        if (this.#running && this.#trackedSessions.has(sessionId)) {
          this.#openSessionStream(sessionId);
        }
      }, delay);
    });

    ws.addEventListener("error", (event) => {
      const msg = event?.message || event?.error?.message || "";
      if (msg && !msg.includes("ECONNREFUSED")) {
        console.warn(
          `[vk-log-stream] session stream ${shortSid} error: ${msg}`,
        );
      }
    });
  }

  /**
   * Handle a message from the session execution-process stream.
   *
   * VK sends JSON Patch arrays. The initial snapshot replaces /execution_processes
   * with an object keyed by process ID. Live updates add/replace/remove at
   * /execution_processes/<processId>.
   *
   * @param {string} sessionId
   * @param {string|Buffer} rawData
   */
  #handleSessionStreamMessage(sessionId, rawData) {
    const text = typeof rawData === "string" ? rawData : rawData.toString();
    if (!text) return;

    let msg;
    try {
      msg = JSON.parse(text);
    } catch {
      return; // ignore non-JSON
    }

    // ── LogMsg::JsonPatch — the primary format for session streams ──
    if (Array.isArray(msg.JsonPatch)) {
      for (const patch of msg.JsonPatch) {
        this.#processSessionPatch(sessionId, patch);
      }
      return;
    }

    // ── Raw JSON Patch array (some VK versions send this directly) ──
    if (Array.isArray(msg)) {
      for (const patch of msg) {
        this.#processSessionPatch(sessionId, patch);
      }
      return;
    }

    // ── LogMsg::Finished — session is done ──
    if ("Finished" in msg) {
      const shortSid = sessionId.slice(0, 8);
      console.log(
        `[vk-log-stream] session stream ${shortSid} received Finished`,
      );
      this.#trackedSessions.delete(sessionId);
      const ws = this.#sessionStreams.get(sessionId);
      if (ws) {
        try {
          ws.close(1000, "session finished");
        } catch {
          /* best effort */
        }
        this.#sessionStreams.delete(sessionId);
      }
      return;
    }
  }

  /**
   * Process a single JSON Patch operation from the session stream.
   * Extracts execution process IDs and connects to their raw-logs streams.
   *
   * @param {string} sessionId
   * @param {object} patch - { op, path, value }
   */
  #processSessionPatch(sessionId, patch) {
    if (!patch || !patch.path) return;

    const { op, path, value } = patch;

    // Initial snapshot: replace /execution_processes → object keyed by ID
    if (
      path === "/execution_processes" &&
      op === "replace" &&
      value &&
      typeof value === "object"
    ) {
      for (const [processId, proc] of Object.entries(value)) {
        this.#maybeConnectProcess(processId, proc, sessionId);
      }
      return;
    }

    // Live update: add/replace /execution_processes/<processId>
    const match = path.match(/^\/execution_processes\/([^/]+)$/);
    if (match) {
      const processId = match[1];
      if (op === "remove") {
        // Process removed — mark finished
        this.#finished.add(processId);
        return;
      }
      if (value && typeof value === "object") {
        this.#maybeConnectProcess(processId, value, sessionId);
      }
    }
  }

  /**
   * Connect to a process's raw-logs stream if it's running and not already tracked.
   * @param {string} processId
   * @param {object} proc - process data from VK (status, run_reason, etc.)
   * @param {string} sessionId
   */
  #maybeConnectProcess(processId, proc, sessionId) {
    if (!processId) return;

    const status = (proc.status || "").toLowerCase();
    if (
      status === "completed" ||
      status === "killed" ||
      status === "failed"
    ) {
      this.#finished.add(processId);
      return;
    }

    if (!this.#connections.has(processId) && !this.#finished.has(processId)) {
      this.#knownProcessIds.add(processId);
      this.#connectWebSocket(processId, {
        sessionId,
        runReason: proc.run_reason,
        status,
      });
    }
  }

  /**
   * Connect WebSocket to a specific execution process's raw-logs endpoint.
   * @param {string} processId
   * @param {object} meta
   */
  #connectWebSocket(processId, meta = {}) {
    const shortId = processId.slice(0, 8);
    const wsUrl = `${this.#wsBaseUrl}/api/execution-processes/${processId}/raw-logs/ws`;

    let ws;
    try {
      ws = new WebSocket(wsUrl);
    } catch (err) {
      console.warn(
        `[vk-log-stream] failed to create WebSocket for ${shortId}: ${err.message}`,
      );
      return;
    }

    this.#connections.set(processId, ws);
    this.#reconnectCounts.set(processId, 0);

    const logPrefix = `[vk-log-stream:${shortId}]`;

    ws.addEventListener("open", () => {
      console.log(
        `${logPrefix} connected to raw-logs WebSocket` +
          (meta.taskId ? ` (task: ${meta.taskId.slice(0, 8)})` : ""),
      );
      this.#reconnectCounts.set(processId, 0);
    });

    ws.addEventListener("message", (event) => {
      this.#handleMessage(processId, event.data, meta);
    });

    ws.addEventListener("close", (event) => {
      this.#connections.delete(processId);
      if (this.#finished.has(processId)) {
        console.log(`${logPrefix} WebSocket closed (process finished)`);
        return;
      }
      if (!this.#running) return;

      const attempts = (this.#reconnectCounts.get(processId) || 0) + 1;
      this.#reconnectCounts.set(processId, attempts);

      if (attempts > MAX_RECONNECT_ATTEMPTS) {
        console.warn(
          `${logPrefix} max reconnect attempts (${MAX_RECONNECT_ATTEMPTS}) reached, giving up`,
        );
        return;
      }

      const delay = Math.min(
        RECONNECT_DELAY_MS * Math.pow(1.5, attempts - 1),
        30000,
      );
      console.log(
        `${logPrefix} reconnecting in ${Math.round(delay)}ms (attempt ${attempts})`,
      );
      setTimeout(() => {
        if (this.#running && !this.#finished.has(processId)) {
          this.#connectWebSocket(processId, meta);
        }
      }, delay);
    });

    ws.addEventListener("error", (event) => {
      // Errors are followed by close events, so just log
      const msg = event?.message || event?.error?.message || "unknown";
      if (!msg.includes("ECONNREFUSED")) {
        console.warn(`${logPrefix} WebSocket error: ${msg}`);
      }
    });
  }

  /**
   * Handle a WebSocket message from the raw-logs stream.
   *
   * VK sends LogMsg variants serialized as JSON:
   *   {"Stdout": "line"}
   *   {"Stderr": "line"}
   *   {"JsonPatch": [...]}
   *   {"Finished": ""}
   *
   * @param {string} processId
   * @param {string|Buffer} rawData
   * @param {object} meta
   */
  #handleMessage(processId, rawData, meta) {
    const text = typeof rawData === "string" ? rawData : rawData.toString();
    if (!text) return;

    let msg;
    try {
      msg = JSON.parse(text);
    } catch {
      // Not JSON — treat as raw text line
      this.#emitLine(processId, text, "stdout", meta);
      return;
    }

    // ── LogMsg::Stdout ──
    if (typeof msg.Stdout === "string") {
      this.#emitLine(processId, msg.Stdout, "stdout", meta);
      return;
    }

    // ── LogMsg::Stderr ──
    if (typeof msg.Stderr === "string") {
      this.#emitLine(processId, msg.Stderr, "stderr", meta);
      return;
    }

    // ── LogMsg::Finished ──
    if ("Finished" in msg) {
      this.#finished.add(processId);
      const shortId = processId.slice(0, 8);
      console.log(`[vk-log-stream:${shortId}] execution process finished`);
      const ws = this.#connections.get(processId);
      if (ws) {
        try {
          ws.close(1000, "finished");
        } catch {
          /* best effort */
        }
        this.#connections.delete(processId);
      }
      // Write final marker to log file
      this.#writeToFile(
        processId,
        `\n--- [vk-log-stream] Process ${shortId} finished at ${new Date().toISOString()} ---\n`,
      );
      return;
    }

    // ── LogMsg::JsonPatch — extract content from patch operations ──
    if (Array.isArray(msg.JsonPatch)) {
      for (const patch of msg.JsonPatch) {
        if (patch?.value) {
          const type = (patch.value.type || "").toUpperCase();
          const content = patch.value.content || patch.value.text || "";
          if (content) {
            const stream =
              type === "STDERR"
                ? "stderr"
                : type === "STDOUT"
                  ? "stdout"
                  : "stdout";
            this.#emitLine(processId, content, stream, meta);
          }
        }
      }
      return;
    }

    // ── LogMsg::SessionId, LogMsg::MessageId, LogMsg::Ready — informational ──
    if (msg.SessionId || msg.MessageId || msg.Ready !== undefined) {
      return; // Ignore informational messages
    }

    // Unknown format — log raw for debugging
    const shortId = processId.slice(0, 8);
    console.warn(
      `[vk-log-stream:${shortId}] unknown message format: ${text.slice(0, 200)}`,
    );
  }

  /**
   * Emit a parsed log line to all outputs (file, console, callback).
   * @param {string} processId
   * @param {string} content
   * @param {"stdout"|"stderr"} stream
   * @param {object} meta
   */
  #emitLine(processId, content, stream, meta) {
    // Strip trailing newlines since we add our own
    const line = content.replace(/\r?\n$/, "");
    if (!line) return;

    const shortId = processId.slice(0, 8);

    // Write to per-process log file
    this.#writeToFile(processId, `[${stream}] ${line}\n`);

    // Echo to console
    if (this.#echo) {
      const prefix = stream === "stderr" ? "ERR" : "OUT";
      try {
        process.stdout.write(`[vk:${shortId}:${prefix}] ${line}\n`);
      } catch {
        /* EPIPE */
      }
    }

    // Callback
    if (this.#onLine) {
      try {
        this.#onLine(line, { processId, stream, ...meta });
      } catch {
        /* callback error — ignore */
      }
    }
  }

  /**
   * Write text to the per-process log file.
   * @param {string} processId
   * @param {string} text
   */
  #writeToFile(processId, text) {
    if (!this.#logDir) return;
    const shortId = processId.slice(0, 8);
    const logPath = resolve(this.#logDir, `vk-exec-${shortId}.log`);
    try {
      appendFileSync(logPath, text);
    } catch {
      /* best effort */
    }
  }
}

export default VkLogStream;
