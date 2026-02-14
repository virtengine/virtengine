import { execSync } from "node:child_process";
import { createHmac } from "node:crypto";
import { existsSync } from "node:fs";
import { open, readFile, readdir, stat } from "node:fs/promises";
import { createServer } from "node:http";
import { resolve, extname } from "node:path";
import { fileURLToPath } from "node:url";
import { getKanbanAdapter } from "./kanban-adapter.mjs";
import { getActiveThreads } from "./agent-pool.mjs";
import {
  listActiveWorktrees,
  getWorktreeStats,
  pruneStaleWorktrees,
  releaseWorktree,
  releaseWorktreeByBranch,
} from "./worktree-manager.mjs";
import {
  loadSharedWorkspaceRegistry,
  sweepExpiredLeases,
  getSharedAvailabilityMap,
  claimSharedWorkspace,
  releaseSharedWorkspace,
  renewSharedWorkspaceLease,
} from "./shared-workspace-registry.mjs";
import { initPresence, listActiveInstances, selectCoordinator } from "./presence.mjs";
import { loadWorkspaceRegistry, getLocalWorkspace } from "./workspace-registry.mjs";

const __dirname = resolve(fileURLToPath(new URL(".", import.meta.url)));
const repoRoot = resolve(__dirname, "..", "..");
const uiRoot = resolve(__dirname, "ui");
const statusPath = resolve(repoRoot, ".cache", "ve-orchestrator-status.json");
const logsDir = resolve(__dirname, "logs");
const agentLogsDir = resolve(repoRoot, ".cache", "agent-logs");

const DEFAULT_PORT = Number(process.env.TELEGRAM_UI_PORT || "0") || 0;
const DEFAULT_HOST = process.env.TELEGRAM_UI_HOST || "0.0.0.0";
const ALLOW_UNSAFE = ["1", "true", "yes"].includes(
  String(process.env.TELEGRAM_UI_ALLOW_UNSAFE || "").toLowerCase(),
);
const AUTH_MAX_AGE_SEC = Number(
  process.env.TELEGRAM_UI_AUTH_MAX_AGE_SEC || "86400",
);
const PRESENCE_TTL_MS = Number(
  process.env.TELEGRAM_PRESENCE_TTL_SEC || "180",
) * 1000;

const MIME_TYPES = {
  ".html": "text/html; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".js": "application/javascript; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".svg": "image/svg+xml",
  ".ico": "image/x-icon",
};

let uiServer = null;
let uiServerUrl = null;
let uiDeps = {};

export function injectUiDependencies(deps = {}) {
  uiDeps = { ...uiDeps, ...deps };
}

export function getTelegramUiUrl() {
  const explicit = process.env.TELEGRAM_UI_BASE_URL || process.env.TELEGRAM_WEBAPP_URL;
  if (explicit) return explicit.replace(/\/+$/, "");
  return uiServerUrl;
}

function jsonResponse(res, statusCode, payload) {
  const body = JSON.stringify(payload, null, 2);
  res.writeHead(statusCode, {
    "Content-Type": "application/json; charset=utf-8",
    "Access-Control-Allow-Origin": "*",
  });
  res.end(body);
}

function textResponse(res, statusCode, body, contentType = "text/plain") {
  res.writeHead(statusCode, {
    "Content-Type": `${contentType}; charset=utf-8`,
    "Access-Control-Allow-Origin": "*",
  });
  res.end(body);
}

function parseInitData(initData) {
  const params = new URLSearchParams(initData);
  const data = {};
  for (const [key, value] of params.entries()) {
    data[key] = value;
  }
  return data;
}

function validateInitData(initData, botToken) {
  if (!initData || !botToken) return false;
  const params = new URLSearchParams(initData);
  const hash = params.get("hash");
  if (!hash) return false;
  params.delete("hash");
  const entries = Array.from(params.entries()).sort(([a], [b]) =>
    a.localeCompare(b),
  );
  const dataCheckString = entries.map(([k, v]) => `${k}=${v}`).join("\n");
  const secret = createHmac("sha256", "WebAppData").update(botToken).digest();
  const signature = createHmac("sha256", secret)
    .update(dataCheckString)
    .digest("hex");
  if (signature !== hash) return false;
  const authDate = Number(params.get("auth_date") || 0);
  if (Number.isFinite(authDate) && authDate > 0 && AUTH_MAX_AGE_SEC > 0) {
    const ageSec = Math.max(0, Math.floor(Date.now() / 1000) - authDate);
    if (ageSec > AUTH_MAX_AGE_SEC) return false;
  }
  return true;
}

function requireAuth(req) {
  if (ALLOW_UNSAFE) return true;
  const initData =
    req.headers["x-telegram-initdata"] ||
    req.headers["x-telegram-init-data"] ||
    req.headers["x-telegram-init"] ||
    req.headers["x-telegram-webapp"] ||
    req.headers["x-telegram-webapp-data"] ||
    "";
  const token = process.env.TELEGRAM_BOT_TOKEN || "";
  if (!initData) return false;
  return validateInitData(String(initData), token);
}

async function readStatusSnapshot() {
  try {
    const raw = await readFile(statusPath, "utf8");
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

function runGit(args, timeoutMs = 10000) {
  return execSync(`git ${args}`, {
    cwd: repoRoot,
    encoding: "utf8",
    timeout: timeoutMs,
  }).trim();
}

async function readJsonBody(req) {
  return new Promise((resolveBody, rejectBody) => {
    let data = "";
    req.on("data", (chunk) => {
      data += chunk;
      if (data.length > 1_000_000) {
        rejectBody(new Error("payload too large"));
        req.destroy();
      }
    });
    req.on("end", () => {
      if (!data) return resolveBody(null);
      try {
        resolveBody(JSON.parse(data));
      } catch (err) {
        rejectBody(err);
      }
    });
  });
}

async function getLatestLogTail(lineCount) {
  const files = await readdir(logsDir).catch(() => []);
  const logFile = files.filter((f) => f.endsWith(".log")).sort().pop();
  if (!logFile) return { file: null, lines: [] };
  const logPath = resolve(logsDir, logFile);
  const content = await readFile(logPath, "utf8");
  const lines = content.split("\n").filter(Boolean);
  const tail = lines.slice(-lineCount);
  return { file: logFile, lines: tail };
}

async function tailFile(filePath, lineCount, maxBytes = 1_000_000) {
  const info = await stat(filePath);
  const size = info.size || 0;
  const start = Math.max(0, size - maxBytes);
  const length = Math.max(0, size - start);
  const handle = await open(filePath, "r");
  const buffer = Buffer.alloc(length);
  try {
    if (length > 0) {
      await handle.read(buffer, 0, length, start);
    }
  } finally {
    await handle.close();
  }
  const text = buffer.toString("utf8");
  const lines = text.split("\n").filter(Boolean);
  const tail = lines.slice(-lineCount);
  return {
    file: filePath,
    lines: tail,
    size,
    truncated: size > maxBytes,
  };
}

async function listAgentLogFiles(query = "", limit = 60) {
  const entries = [];
  const files = await readdir(agentLogsDir).catch(() => []);
  for (const name of files) {
    if (!name.endsWith(".log")) continue;
    if (query && !name.toLowerCase().includes(query.toLowerCase())) continue;
    try {
      const info = await stat(resolve(agentLogsDir, name));
      entries.push({
        name,
        size: info.size,
        mtime: info.mtime?.toISOString?.() || new Date(info.mtime).toISOString(),
        mtimeMs: info.mtimeMs,
      });
    } catch {
      // ignore
    }
  }
  entries.sort((a, b) => b.mtimeMs - a.mtimeMs);
  return entries.slice(0, limit);
}

async function ensurePresenceLoaded() {
  const loaded = await loadWorkspaceRegistry().catch(() => null);
  const registry = loaded?.registry || loaded || null;
  const localWorkspace = registry
    ? getLocalWorkspace(registry, process.env.VE_WORKSPACE_ID || "")
    : null;
  await initPresence({ repoRoot, localWorkspace });
}

async function handleApi(req, res, url) {
  if (req.method === "OPTIONS") {
    res.writeHead(204, {
      "Access-Control-Allow-Origin": "*",
      "Access-Control-Allow-Methods": "GET,POST,OPTIONS",
      "Access-Control-Allow-Headers": "Content-Type,X-Telegram-InitData",
    });
    res.end();
    return;
  }

  if (!requireAuth(req)) {
    jsonResponse(res, 401, {
      ok: false,
      error: "Unauthorized. Telegram init data missing or invalid.",
    });
    return;
  }

  const path = url.pathname;
  if (path === "/api/status") {
    const data = await readStatusSnapshot();
    jsonResponse(res, 200, { ok: true, data });
    return;
  }

  if (path === "/api/executor") {
    const executor = uiDeps.getInternalExecutor?.();
    const mode = uiDeps.getExecutorMode?.() || "vk";
    jsonResponse(res, 200, {
      ok: true,
      data: executor?.getStatus?.() || null,
      mode,
      paused: executor?.isPaused?.() || false,
    });
    return;
  }

  if (path === "/api/executor/pause") {
    const executor = uiDeps.getInternalExecutor?.();
    if (!executor) {
      jsonResponse(res, 400, { ok: false, error: "Internal executor not enabled." });
      return;
    }
    executor.pause();
    jsonResponse(res, 200, { ok: true, paused: true });
    return;
  }

  if (path === "/api/executor/resume") {
    const executor = uiDeps.getInternalExecutor?.();
    if (!executor) {
      jsonResponse(res, 400, { ok: false, error: "Internal executor not enabled." });
      return;
    }
    executor.resume();
    jsonResponse(res, 200, { ok: true, paused: false });
    return;
  }

  if (path === "/api/executor/maxparallel") {
    try {
      const executor = uiDeps.getInternalExecutor?.();
      if (!executor) {
        jsonResponse(res, 400, { ok: false, error: "Internal executor not enabled." });
        return;
      }
      const body = await readJsonBody(req);
      const value = Number(body?.value ?? body?.maxParallel);
      if (!Number.isFinite(value) || value < 0 || value > 20) {
        jsonResponse(res, 400, { ok: false, error: "value must be between 0 and 20" });
        return;
      }
      executor.maxParallel = value;
      if (value === 0) {
        executor.pause();
      } else if (executor.isPaused?.()) {
        executor.resume();
      }
      jsonResponse(res, 200, { ok: true, maxParallel: executor.maxParallel });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/projects") {
    try {
      const adapter = getKanbanAdapter();
      const projects = await adapter.listProjects();
      jsonResponse(res, 200, { ok: true, data: projects });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/tasks") {
    const status = url.searchParams.get("status") || "";
    const projectId = url.searchParams.get("project") || "";
    const page = Math.max(0, Number(url.searchParams.get("page") || "0"));
    const pageSize = Math.min(
      50,
      Math.max(5, Number(url.searchParams.get("pageSize") || "15")),
    );
    try {
      const adapter = getKanbanAdapter();
      const projects = await adapter.listProjects();
      const activeProject =
        projectId || projects[0]?.id || projects[0]?.project_id || "";
      if (!activeProject) {
        jsonResponse(res, 200, { ok: true, data: [], page, pageSize, total: 0 });
        return;
      }
      const tasks = await adapter.listTasks(activeProject, status ? { status } : {});
      const total = tasks.length;
      const start = page * pageSize;
      const slice = tasks.slice(start, start + pageSize);
      jsonResponse(res, 200, {
        ok: true,
        data: slice,
        page,
        pageSize,
        total,
        projectId: activeProject,
      });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/tasks/start") {
    try {
      const body = await readJsonBody(req);
      const taskId = body?.taskId || body?.id;
      if (!taskId) {
        jsonResponse(res, 400, { ok: false, error: "taskId is required" });
        return;
      }
      const executor = uiDeps.getInternalExecutor?.();
      if (!executor) {
        jsonResponse(res, 400, {
          ok: false,
          error: "Internal executor not enabled. Set EXECUTOR_MODE=internal or hybrid.",
        });
        return;
      }
      const adapter = getKanbanAdapter();
      const task = await adapter.getTask(taskId);
      if (!task) {
        jsonResponse(res, 404, { ok: false, error: "Task not found." });
        return;
      }
      void executor.executeTask(task);
      jsonResponse(res, 200, { ok: true, taskId });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/tasks/update") {
    try {
      const body = await readJsonBody(req);
      const taskId = body?.taskId || body?.id;
      const status = body?.status;
      if (!taskId || !status) {
        jsonResponse(res, 400, { ok: false, error: "taskId and status required" });
        return;
      }
      const adapter = getKanbanAdapter();
      const updated = await adapter.updateTaskStatus(taskId, status);
      jsonResponse(res, 200, { ok: true, data: updated });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/logs") {
    const lines = Math.min(1000, Math.max(10, Number(url.searchParams.get("lines") || "200")));
    try {
      const tail = await getLatestLogTail(lines);
      jsonResponse(res, 200, { ok: true, data: tail });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/threads") {
    try {
      const threads = getActiveThreads();
      jsonResponse(res, 200, { ok: true, data: threads });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/worktrees") {
    try {
      const worktrees = listActiveWorktrees();
      const stats = await getWorktreeStats();
      jsonResponse(res, 200, { ok: true, data: worktrees, stats });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/worktrees/prune") {
    try {
      const result = await pruneStaleWorktrees({ actor: "telegram-ui" });
      jsonResponse(res, 200, { ok: true, data: result });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/worktrees/release") {
    try {
      const body = await readJsonBody(req);
      const taskKey = body?.taskKey || body?.key;
      const branch = body?.branch;
      let released = null;
      if (taskKey) {
        released = await releaseWorktree(taskKey);
      } else if (branch) {
        released = await releaseWorktreeByBranch(branch);
      } else {
        jsonResponse(res, 400, { ok: false, error: "taskKey or branch required" });
        return;
      }
      jsonResponse(res, 200, { ok: true, data: released });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/presence") {
    try {
      await ensurePresenceLoaded();
      const instances = listActiveInstances({ ttlMs: PRESENCE_TTL_MS });
      const coordinator = selectCoordinator({ ttlMs: PRESENCE_TTL_MS });
      jsonResponse(res, 200, { ok: true, data: { instances, coordinator } });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/shared-workspaces") {
    try {
      const registry = await loadSharedWorkspaceRegistry();
      const sweep = await sweepExpiredLeases({
        registry,
        actor: "telegram-ui",
      });
      const availability = getSharedAvailabilityMap(sweep.registry);
      jsonResponse(res, 200, {
        ok: true,
        data: sweep.registry,
        availability,
        expired: sweep.expired || [],
      });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/shared-workspaces/claim") {
    try {
      const body = await readJsonBody(req);
      const workspaceId = body?.workspaceId || body?.id;
      if (!workspaceId) {
        jsonResponse(res, 400, { ok: false, error: "workspaceId required" });
        return;
      }
      const result = await claimSharedWorkspace({
        workspaceId,
        owner: body?.owner,
        ttlMinutes: body?.ttlMinutes,
        note: body?.note,
        actor: "telegram-ui",
      });
      if (result.error) {
        jsonResponse(res, 400, { ok: false, error: result.error });
        return;
      }
      jsonResponse(res, 200, { ok: true, data: result.workspace, lease: result.lease });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/shared-workspaces/release") {
    try {
      const body = await readJsonBody(req);
      const workspaceId = body?.workspaceId || body?.id;
      if (!workspaceId) {
        jsonResponse(res, 400, { ok: false, error: "workspaceId required" });
        return;
      }
      const result = await releaseSharedWorkspace({
        workspaceId,
        owner: body?.owner,
        force: body?.force,
        reason: body?.reason,
        actor: "telegram-ui",
      });
      if (result.error) {
        jsonResponse(res, 400, { ok: false, error: result.error });
        return;
      }
      jsonResponse(res, 200, { ok: true, data: result.workspace });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/shared-workspaces/renew") {
    try {
      const body = await readJsonBody(req);
      const workspaceId = body?.workspaceId || body?.id;
      if (!workspaceId) {
        jsonResponse(res, 400, { ok: false, error: "workspaceId required" });
        return;
      }
      const result = await renewSharedWorkspaceLease({
        workspaceId,
        owner: body?.owner,
        ttlMinutes: body?.ttlMinutes,
        actor: "telegram-ui",
      });
      if (result.error) {
        jsonResponse(res, 400, { ok: false, error: result.error });
        return;
      }
      jsonResponse(res, 200, { ok: true, data: result.workspace, lease: result.lease });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/agent-logs") {
    try {
      const file = url.searchParams.get("file");
      const query = url.searchParams.get("query") || "";
      const lines = Math.min(
        1000,
        Math.max(20, Number(url.searchParams.get("lines") || "200")),
      );
      if (!file) {
        const files = await listAgentLogFiles(query);
        jsonResponse(res, 200, { ok: true, data: files });
        return;
      }
      const filePath = resolve(agentLogsDir, file);
      if (!filePath.startsWith(agentLogsDir)) {
        jsonResponse(res, 403, { ok: false, error: "Forbidden" });
        return;
      }
      if (!existsSync(filePath)) {
        jsonResponse(res, 404, { ok: false, error: "Log not found" });
        return;
      }
      const tail = await tailFile(filePath, lines);
      jsonResponse(res, 200, { ok: true, data: tail });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/agent-logs/context") {
    try {
      const query = url.searchParams.get("query") || "";
      if (!query) {
        jsonResponse(res, 400, { ok: false, error: "query required" });
        return;
      }
      const worktreeDir = resolve(repoRoot, ".cache", "worktrees");
      const dirs = await readdir(worktreeDir).catch(() => []);
      const matches = dirs.filter((d) =>
        d.toLowerCase().includes(query.toLowerCase()),
      );
      if (matches.length === 0) {
        jsonResponse(res, 200, { ok: true, data: { matches: [] } });
        return;
      }
      const wtName = matches[0];
      const wtPath = resolve(worktreeDir, wtName);
      let gitLog = "";
      let gitStatus = "";
      let diffStat = "";
      try {
        gitLog = execSync("git log --oneline -5 2>&1", {
          cwd: wtPath,
          encoding: "utf8",
          timeout: 10000,
        }).trim();
      } catch {
        gitLog = "";
      }
      try {
        gitStatus = execSync("git status --short 2>&1", {
          cwd: wtPath,
          encoding: "utf8",
          timeout: 10000,
        }).trim();
      } catch {
        gitStatus = "";
      }
      try {
        const branch = execSync("git branch --show-current 2>&1", {
          cwd: wtPath,
          encoding: "utf8",
          timeout: 5000,
        }).trim();
        diffStat = execSync(`git diff --stat main...${branch} 2>&1`, {
          cwd: wtPath,
          encoding: "utf8",
          timeout: 10000,
        }).trim();
      } catch {
        diffStat = "";
      }
      jsonResponse(res, 200, {
        ok: true,
        data: {
          name: wtName,
          path: wtPath,
          gitLog,
          gitStatus,
          diffStat,
        },
      });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/git/branches") {
    try {
      const raw = runGit("branch -a --sort=-committerdate", 15000);
      const lines = raw.split("\n").map((line) => line.trim()).filter(Boolean);
      jsonResponse(res, 200, { ok: true, data: lines.slice(0, 40) });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (path === "/api/git/diff") {
    try {
      const diff = runGit("diff --stat HEAD", 15000);
      jsonResponse(res, 200, { ok: true, data: diff });
    } catch (err) {
      jsonResponse(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  jsonResponse(res, 404, { ok: false, error: "Unknown API endpoint" });
}

async function handleStatic(req, res, url) {
  const pathname = url.pathname === "/" ? "/index.html" : url.pathname;
  const filePath = resolve(uiRoot, `.${pathname}`);

  if (!filePath.startsWith(uiRoot)) {
    textResponse(res, 403, "Forbidden");
    return;
  }

  if (!existsSync(filePath)) {
    textResponse(res, 404, "Not Found");
    return;
  }

  try {
    const ext = extname(filePath).toLowerCase();
    const contentType = MIME_TYPES[ext] || "application/octet-stream";
    const data = await readFile(filePath);
    res.writeHead(200, {
      "Content-Type": contentType,
      "Access-Control-Allow-Origin": "*",
      "Cache-Control": "no-store",
    });
    res.end(data);
  } catch (err) {
    textResponse(res, 500, `Failed to load ${pathname}: ${err.message}`);
  }
}

export async function startTelegramUiServer(options = {}) {
  if (uiServer) return uiServer;

  const port = Number(options.port || DEFAULT_PORT);
  if (!port) return null;

  injectUiDependencies(options.dependencies || {});

  uiServer = createServer(async (req, res) => {
    const url = new URL(req.url || "/", `http://${req.headers.host || "localhost"}`);
    if (url.pathname.startsWith("/api/")) {
      await handleApi(req, res, url);
      return;
    }
    await handleStatic(req, res, url);
  });

  await new Promise((resolveReady, rejectReady) => {
    uiServer.once("error", rejectReady);
    uiServer.listen(port, options.host || DEFAULT_HOST, () => {
      resolveReady();
    });
  });

  const host = options.publicHost || process.env.TELEGRAM_UI_PUBLIC_HOST || "localhost";
  const protocol = host === "localhost" || host.startsWith("127.") ? "http" : "https";
  uiServerUrl = `${protocol}://${host}:${port}`;
  console.log(`[telegram-ui] server listening on ${uiServerUrl}`);

  return uiServer;
}

export function stopTelegramUiServer() {
  if (!uiServer) return;
  try {
    uiServer.close();
  } catch {
    /* best effort */
  }
  uiServer = null;
}
