import { execSync } from "node:child_process";
import { existsSync } from "node:fs";
import { readFile, readdir, stat } from "node:fs/promises";
import { createServer } from "node:http";
import { resolve, extname } from "node:path";
import { fileURLToPath } from "node:url";
import { getActiveThreads } from "./agent-pool.mjs";
import {
  listActiveWorktrees,
  getWorktreeStats,
} from "./worktree-manager.mjs";

const __dirname = resolve(fileURLToPath(new URL(".", import.meta.url)));
const repoRoot = resolve(__dirname, "..", "..");
const uiRoot = resolve(__dirname, "ui");
const statusPath = resolve(repoRoot, ".cache", "ve-orchestrator-status.json");
const logsDir = resolve(__dirname, "logs");

const defaultPort = Number(process.env.TELEGRAM_UI_PORT || "0") || 0;
const defaultHost = process.env.TELEGRAM_UI_HOST || "127.0.0.1";

const mimeTypes = {
  ".html": "text/html; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".js": "application/javascript; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".svg": "image/svg+xml",
};

let server = null;
let serverUrl = null;
let deps = {};

export function injectUiDependencies(next = {}) {
  deps = { ...deps, ...next };
}

export function getTelegramUiUrl() {
  const explicit =
    process.env.TELEGRAM_UI_BASE_URL || process.env.TELEGRAM_WEBAPP_URL;
  if (explicit) return String(explicit).replace(/\/+$/, "");
  return serverUrl;
}

function json(res, status, payload) {
  res.writeHead(status, {
    "Content-Type": "application/json; charset=utf-8",
    "Cache-Control": "no-store",
    "Access-Control-Allow-Origin": "*",
    "Access-Control-Allow-Headers": "Content-Type, X-Telegram-InitData",
    "Access-Control-Allow-Methods": "GET,POST,OPTIONS",
  });
  res.end(JSON.stringify(payload, null, 2));
}

function text(res, status, body, contentType = "text/plain; charset=utf-8") {
  res.writeHead(status, {
    "Content-Type": contentType,
    "Cache-Control": "no-store",
  });
  res.end(body);
}

async function readStatus() {
  try {
    const raw = await readFile(statusPath, "utf8");
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

function normalizeTasks(status) {
  const attempts = status?.attempts || {};
  return Object.values(attempts || {})
    .filter(Boolean)
    .map((attempt) => ({
      task_id: attempt.task_id || "",
      task_title: attempt.task_title || "",
      status: attempt.status || "unknown",
      branch: attempt.branch || "",
      executor: attempt.executor || "",
      started_at: attempt.started_at || attempt.created_at || null,
    }));
}

function normalizeAgents(executorStatus) {
  const slots = Array.isArray(executorStatus?.slots) ? executorStatus.slots : [];
  return slots.map((slot) => ({
    agentInstanceId: slot.agentInstanceId || null,
    taskId: slot.taskId || "",
    taskTitle: slot.taskTitle || "",
    status: slot.status || "unknown",
    sdk: slot.sdk || "",
    branch: slot.branch || "",
    runningFor: slot.runningFor || 0,
  }));
}

async function tailLatestLog(lineCount = 200) {
  const files = await readdir(logsDir).catch(() => []);
  const latest = files
    .filter((file) => file.endsWith(".log"))
    .sort()
    .pop();
  if (!latest) return { file: null, tail: [] };
  const full = resolve(logsDir, latest);
  const raw = await readFile(full, "utf8").catch(() => "");
  const lines = raw.split("\n").filter(Boolean);
  return { file: latest, tail: lines.slice(-Math.max(20, lineCount)) };
}

async function runCommand(command, chatId = process.env.TELEGRAM_CHAT_ID || "") {
  const value = String(command || "").trim();
  if (!value.startsWith("/")) {
    return { ok: false, error: "Command must start with /" };
  }
  if (typeof deps.executeCommand !== "function") {
    return { ok: false, error: "Command bridge unavailable" };
  }
  await deps.executeCommand(value, chatId);
  return { ok: true };
}

function parseBody(req) {
  return new Promise((resolveBody, rejectBody) => {
    let raw = "";
    req.on("data", (chunk) => {
      raw += chunk;
      if (raw.length > 1_000_000) {
        rejectBody(new Error("payload too large"));
      }
    });
    req.on("end", () => {
      if (!raw.trim()) return resolveBody({});
      try {
        resolveBody(JSON.parse(raw));
      } catch (err) {
        rejectBody(err);
      }
    });
  });
}

async function handleApi(req, res, url) {
  if (req.method === "OPTIONS") {
    res.writeHead(204, {
      "Access-Control-Allow-Origin": "*",
      "Access-Control-Allow-Headers": "Content-Type, X-Telegram-InitData",
      "Access-Control-Allow-Methods": "GET,POST,OPTIONS",
    });
    res.end();
    return;
  }

  if (url.pathname === "/api/ping") {
    json(res, 200, { ok: true, now: new Date().toISOString() });
    return;
  }

  if (url.pathname === "/api/bootstrap") {
    const status = await readStatus();
    const executor = deps.getInternalExecutor?.()?.getStatus?.() || null;
    json(res, 200, {
      ok: true,
      status,
      executor,
      executorMode: deps.getExecutorMode?.() || "vk",
      now: new Date().toISOString(),
    });
    return;
  }

  if (url.pathname === "/api/status") {
    json(res, 200, { ok: true, data: await readStatus() });
    return;
  }

  if (url.pathname === "/api/tasks") {
    const status = await readStatus();
    json(res, 200, { ok: true, tasks: normalizeTasks(status) });
    return;
  }

  if (url.pathname === "/api/agents") {
    const executorStatus = deps.getInternalExecutor?.()?.getStatus?.() || null;
    json(res, 200, { ok: true, agents: normalizeAgents(executorStatus) });
    return;
  }

  if (url.pathname === "/api/worktrees") {
    const worktrees = await listActiveWorktrees().catch(() => []);
    const stats = await getWorktreeStats().catch(() => ({}));
    json(res, 200, { ok: true, worktrees, stats });
    return;
  }

  if (url.pathname === "/api/threads") {
    const threads = getActiveThreads().map((thread) => ({
      taskKey: thread.taskKey,
      sdk: thread.sdk,
      turnCount: thread.turnCount,
    }));
    json(res, 200, { ok: true, threads });
    return;
  }

  if (url.pathname === "/api/logs") {
    const lines = Number(url.searchParams.get("lines") || "200") || 200;
    json(res, 200, { ok: true, ...(await tailLatestLog(lines)) });
    return;
  }

  if (url.pathname === "/api/git/branches") {
    try {
      const output = execSync("git for-each-ref --sort=-committerdate --count=20 --format='%(refname:short)' refs/heads", {
        cwd: repoRoot,
        encoding: "utf8",
      });
      const data = output
        .split("\n")
        .map((line) => line.trim())
        .filter(Boolean);
      json(res, 200, { ok: true, data });
    } catch (err) {
      json(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (url.pathname === "/api/git/diff") {
    try {
      const data = execSync("git status --short", {
        cwd: repoRoot,
        encoding: "utf8",
      });
      json(res, 200, { ok: true, data });
    } catch (err) {
      json(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (url.pathname === "/api/executor/pause" && req.method === "POST") {
    const executor = deps.getInternalExecutor?.();
    if (!executor) {
      json(res, 400, { ok: false, error: "Internal executor unavailable" });
      return;
    }
    executor.pause();
    json(res, 200, { ok: true });
    return;
  }

  if (url.pathname === "/api/executor/resume" && req.method === "POST") {
    const executor = deps.getInternalExecutor?.();
    if (!executor) {
      json(res, 400, { ok: false, error: "Internal executor unavailable" });
      return;
    }
    executor.resume();
    json(res, 200, { ok: true });
    return;
  }

  if (url.pathname === "/api/executor/maxparallel" && req.method === "POST") {
    try {
      const executor = deps.getInternalExecutor?.();
      if (!executor) {
        json(res, 400, { ok: false, error: "Internal executor unavailable" });
        return;
      }
      const body = await parseBody(req);
      const value = Number(body?.value ?? -1);
      if (!Number.isFinite(value) || value < 0 || value > 20) {
        json(res, 400, { ok: false, error: "Value must be between 0 and 20" });
        return;
      }
      executor.maxParallel = value;
      if (value === 0) executor.pause();
      else if (executor.isPaused?.()) executor.resume();
      json(res, 200, { ok: true, maxParallel: executor.maxParallel });
    } catch (err) {
      json(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  if (url.pathname === "/api/command" && req.method === "POST") {
    try {
      const body = await parseBody(req);
      const result = await runCommand(body?.command, body?.chatId);
      json(res, result.ok ? 200 : 400, result);
    } catch (err) {
      json(res, 500, { ok: false, error: err.message });
    }
    return;
  }

  json(res, 404, { ok: false, error: "Unknown API endpoint" });
}

async function serveStatic(res, pathname) {
  const relative = pathname === "/" ? "/index.html" : pathname;
  const clean = relative.replace(/\.\./g, "");
  const filePath = resolve(uiRoot, `.${clean}`);
  if (!filePath.startsWith(uiRoot)) {
    text(res, 403, "Forbidden");
    return;
  }
  if (!existsSync(filePath)) {
    text(res, 404, "Not found");
    return;
  }
  const ext = extname(filePath).toLowerCase();
  const body = await readFile(filePath);
  res.writeHead(200, {
    "Content-Type": mimeTypes[ext] || "application/octet-stream",
    "Cache-Control": "no-store",
  });
  res.end(body);
}

export async function startTelegramUiServer(options = {}) {
  if (options.dependencies) injectUiDependencies(options.dependencies);
  if (server) return { url: getTelegramUiUrl(), alreadyRunning: true };

  server = createServer(async (req, res) => {
    try {
      const url = new URL(req.url || "/", "http://localhost");
      if (url.pathname.startsWith("/api/")) {
        await handleApi(req, res, url);
        return;
      }
      await serveStatic(res, url.pathname);
    } catch (err) {
      json(res, 500, { ok: false, error: err.message || String(err) });
    }
  });

  await new Promise((resolveStart, rejectStart) => {
    server.once("error", rejectStart);
    server.listen(defaultPort, defaultHost, () => resolveStart());
  });

  const address = server.address();
  const port = typeof address === "object" && address ? address.port : defaultPort;
  serverUrl = `http://${defaultHost === "0.0.0.0" ? "127.0.0.1" : defaultHost}:${port}`;
  return { url: getTelegramUiUrl(), localUrl: serverUrl };
}

export async function stopTelegramUiServer() {
  if (!server) return;
  const current = server;
  server = null;
  serverUrl = null;
  await new Promise((resolveStop) => {
    try {
      current.close(() => resolveStop());
    } catch {
      resolveStop();
    }
  });
}
