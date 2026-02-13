#!/usr/bin/env node

/**
 * telegram-sentinel.mjs — Always-on Telegram command listener for VirtEngine.
 *
 * Runs independently of the main codex-monitor process, ensuring Telegram
 * commands are always handled even when codex-monitor is down.
 *
 * Architecture:
 *   ┌─────────────────┐
 *   │ telegram-sentinel│──── always running ────────────────────────────────┐
 *   │  (this file)     │                                                   │
 *   └────────┬─────────┘                                                   │
 *            │                                                             │
 *            ├─ Standalone Mode (codex-monitor DOWN)                       │
 *            │   ├─ Polls Telegram directly                                │
 *            │   ├─ Handles simple commands (/ping, /status, /sentinel)    │
 *            │   └─ Auto-starts codex-monitor for complex commands         │
 *            │                                                             │
 *            └─ Companion Mode (codex-monitor UP)                          │
 *                ├─ Does NOT poll (lets telegram-bot.mjs handle it)        │
 *                ├─ Monitors codex-monitor health via PID file             │
 *                └─ Transitions to Standalone if codex-monitor dies        │
 *                                                                          │
 *   ┌─────────────────┐                                                    │
 *   │  codex-monitor   │ ← started/stopped by sentinel as needed ─────────┘
 *   │  (cli.mjs fork)  │
 *   └─────────────────┘
 *
 * Usage:
 *   node telegram-sentinel.mjs          # start sentinel
 *   node telegram-sentinel.mjs --stop   # stop sentinel
 *   node telegram-sentinel.mjs --status # check sentinel status
 */

import {
  existsSync,
  readFileSync,
  mkdirSync,
  unlinkSync,
  writeFileSync,
} from "node:fs";
import { readFile, writeFile, unlink } from "node:fs/promises";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import { spawn } from "node:child_process";
import os from "node:os";

// ── Paths ────────────────────────────────────────────────────────────────────

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const repoRoot = resolve(__dirname, "..", "..");
const cacheDir = resolve(repoRoot, ".cache");

const MONITOR_PID_FILE = resolve(__dirname, ".cache", "codex-monitor.pid");
const SENTINEL_PID_FILE = resolve(cacheDir, "telegram-sentinel.pid");
const SENTINEL_HEARTBEAT_FILE = resolve(cacheDir, "sentinel-heartbeat.json");
const SENTINEL_LOCK_FILE = resolve(cacheDir, "telegram-sentinel.lock");
const SENTINEL_COMMAND_QUEUE_FILE = resolve(
  cacheDir,
  "sentinel-command-queue.json",
);
const MONITOR_POLL_LOCK_FILE = resolve(cacheDir, "telegram-getupdates.lock");
const STATUS_FILE = resolve(cacheDir, "ve-orchestrator-status.json");

const TAG = "[sentinel]";
const POLL_TIMEOUT_S = 30;
const MAX_MESSAGE_LEN = 4000;
const HEALTH_CHECK_INTERVAL_MS = 30_000;
const POLL_ERROR_BACKOFF_BASE_MS = 5_000;
const POLL_ERROR_BACKOFF_MAX_MS = 120_000;
const COMMAND_QUEUE_MAX_SIZE = 50;
const COMMAND_QUEUE_TTL_MS = 10 * 60 * 1000; // 10 minutes
const MONITOR_START_TIMEOUT_MS = 60_000; // 60s to wait for monitor to become healthy
const MONITOR_HEALTH_POLL_MS = 2_000; // check every 2s during startup

// ── State ────────────────────────────────────────────────────────────────────

/** @type {"standalone" | "companion"} */
let mode = "standalone";
let running = false;
let polling = false;
/** @type {AbortController | null} */
let pollAbort = null;
let lastUpdateId = 0;
let healthCheckTimer = null;
let heartbeatTimer = null;
let consecutivePollErrors = 0;
let commandsProcessed = 0;
let startedAt = new Date().toISOString();
/** @type {Array<{ chatId: string|number, text: string, timestamp: number }>} */
let commandQueue = [];
/** @type {Promise<void> | null} */
let monitorStartPromise = null;
let sentinelPollLockHeld = false;

// ── Environment ──────────────────────────────────────────────────────────────

/** @type {string} */
let telegramToken = "";
/** @type {string} */
let telegramChatId = "";
/** @type {string} */
let projectName = "";

/**
 * Parse the .env file for Telegram credentials and project name.
 * Uses a simple line-by-line parser — no external dependencies.
 * @returns {{ TELEGRAM_BOT_TOKEN?: string, TELEGRAM_CHAT_ID?: string, PROJECT_NAME?: string }}
 */
function loadEnvCredentials() {
  const envPath = resolve(__dirname, ".env");
  /** @type {Record<string, string>} */
  const vars = {};

  if (!existsSync(envPath)) return vars;

  try {
    const lines = readFileSync(envPath, "utf8").split("\n");
    for (const line of lines) {
      const trimmed = line.trim();
      if (!trimmed || trimmed.startsWith("#")) continue;
      const eqIdx = trimmed.indexOf("=");
      if (eqIdx === -1) continue;
      const key = trimmed.slice(0, eqIdx).trim();
      let val = trimmed.slice(eqIdx + 1).trim();
      // Strip surrounding quotes
      if (
        (val.startsWith('"') && val.endsWith('"')) ||
        (val.startsWith("'") && val.endsWith("'"))
      ) {
        val = val.slice(1, -1);
      }
      vars[key] = val;
    }
  } catch {
    // best effort
  }

  return vars;
}

/**
 * Initialize environment variables from .env and process.env.
 * Process.env takes precedence over .env file values.
 */
function initEnv() {
  const fileVars = loadEnvCredentials();
  telegramToken =
    process.env.TELEGRAM_BOT_TOKEN || fileVars.TELEGRAM_BOT_TOKEN || "";
  telegramChatId =
    process.env.TELEGRAM_CHAT_ID || fileVars.TELEGRAM_CHAT_ID || "";
  projectName =
    process.env.PROJECT_NAME || fileVars.PROJECT_NAME || "virtengine";
}

// ── Process Utilities ────────────────────────────────────────────────────────

/**
 * Check if a process with the given PID is alive.
 * @param {number} pid
 * @returns {boolean}
 */
function isProcessAlive(pid) {
  if (!Number.isFinite(pid) || pid <= 0) return false;
  try {
    process.kill(pid, 0);
    return true;
  } catch {
    return false;
  }
}

/**
 * Read a PID from a file and check if the process is alive.
 * @param {string} pidPath
 * @returns {number | null} The PID if alive, null otherwise.
 */
function readAlivePid(pidPath) {
  try {
    if (!existsSync(pidPath)) return null;
    const pid = parseInt(readFileSync(pidPath, "utf8").trim(), 10);
    if (isNaN(pid)) return null;
    return isProcessAlive(pid) ? pid : null;
  } catch {
    return null;
  }
}

/**
 * Write a PID file atomically (best effort).
 * @param {string} pidPath
 * @param {number} pid
 */
function writePidFile(pidPath, pid) {
  try {
    mkdirSync(dirname(pidPath), { recursive: true });
    writeFileSync(pidPath, String(pid), "utf8");
  } catch {
    /* best effort */
  }
}

/**
 * Remove a PID file.
 * @param {string} pidPath
 */
function removePidFile(pidPath) {
  try {
    if (existsSync(pidPath)) unlinkSync(pidPath);
  } catch {
    /* best effort */
  }
}

// ── Sentinel Lock ────────────────────────────────────────────────────────────

/**
 * Acquire the sentinel poll lock file. Uses exclusive create (wx) to prevent
 * races between multiple sentinel instances.
 * @returns {Promise<boolean>}
 */
async function acquireSentinelPollLock() {
  if (sentinelPollLockHeld) return true;
  try {
    const payload = JSON.stringify(
      {
        owner: "sentinel",
        pid: process.pid,
        started_at: new Date().toISOString(),
      },
      null,
      2,
    );
    await writeFile(SENTINEL_LOCK_FILE, payload, { flag: "wx" });
    sentinelPollLockHeld = true;
    return true;
  } catch (err) {
    if (err && err.code === "EEXIST") {
      // Check if the existing lock holder is still alive
      try {
        const raw = await readFile(SENTINEL_LOCK_FILE, "utf8");
        if (!raw || !raw.trim()) {
          await unlink(SENTINEL_LOCK_FILE).catch(() => {});
          return acquireSentinelPollLock();
        }
        const data = JSON.parse(raw);
        const pid = Number(data?.pid);
        if (!isProcessAlive(pid)) {
          // Stale lock — reclaim
          await unlink(SENTINEL_LOCK_FILE).catch(() => {});
          return acquireSentinelPollLock();
        }
        // Another live sentinel holds the lock
        return false;
      } catch {
        // Corrupt lock file — remove and retry
        await unlink(SENTINEL_LOCK_FILE).catch(() => {});
        return acquireSentinelPollLock();
      }
    }
    return false;
  }
}

/**
 * Release the sentinel poll lock file.
 * @returns {Promise<void>}
 */
async function releaseSentinelPollLock() {
  if (!sentinelPollLockHeld) return;
  sentinelPollLockHeld = false;
  try {
    await unlink(SENTINEL_LOCK_FILE).catch(() => {});
  } catch {
    /* best effort */
  }
}

// ── Telegram API ─────────────────────────────────────────────────────────────

/**
 * Send a text message to a Telegram chat.
 * Handles message splitting for long texts and retries on transient errors.
 * @param {string | number} chatId
 * @param {string} text
 * @param {object} [options]
 * @param {string} [options.parseMode]
 * @param {boolean} [options.silent]
 * @returns {Promise<number | null>} The message_id of the last sent chunk, or null.
 */
async function sendTelegram(chatId, text, options = {}) {
  if (!telegramToken) return null;
  const chunks = splitMessage(text, MAX_MESSAGE_LEN);
  let lastMessageId = null;

  for (const chunk of chunks) {
    const url = `https://api.telegram.org/bot${telegramToken}/sendMessage`;
    /** @type {Record<string, any>} */
    const payload = {
      chat_id: chatId,
      text: chunk,
      disable_web_page_preview: true,
    };
    if (options.parseMode) payload.parse_mode = options.parseMode;
    if (options.silent) payload.disable_notification = true;

    try {
      const res = await fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
        signal: AbortSignal.timeout(15_000),
      });

      if (!res || typeof res.ok === "undefined") {
        log("warn", "send error: invalid response object");
        continue;
      }

      if (!res.ok) {
        const body = await res.text().catch(() => "");
        log("warn", `send failed: ${res.status} ${body}`);
        // If parse_mode caused the error, retry as plain text
        if (options.parseMode && res.status === 400) {
          return sendTelegram(chatId, chunk, {
            ...options,
            parseMode: undefined,
          });
        }
      } else {
        try {
          const data = await res.json();
          if (data.ok && data.result?.message_id) {
            lastMessageId = data.result.message_id;
          }
        } catch {
          /* best effort */
        }
      }
    } catch (err) {
      log("warn", `send error: ${err.message}`);
    }
  }
  return lastMessageId;
}

/**
 * Split a text into chunks that fit within Telegram's message limit.
 * @param {string} text
 * @param {number} maxLen
 * @returns {string[]}
 */
function splitMessage(text, maxLen) {
  if (!text) return ["(empty)"];
  if (text.length <= maxLen) return [text];
  const chunks = [];
  let remaining = text;
  while (remaining.length > 0) {
    if (remaining.length <= maxLen) {
      chunks.push(remaining);
      break;
    }
    let splitIdx = remaining.lastIndexOf("\n", maxLen);
    if (splitIdx < maxLen * 0.3) splitIdx = maxLen;
    chunks.push(remaining.slice(0, splitIdx));
    remaining = remaining.slice(splitIdx);
  }
  return chunks;
}

// ── Telegram Polling ─────────────────────────────────────────────────────────

/**
 * Long-poll the Telegram Bot API for new updates.
 * @returns {Promise<Array<object>>}
 */
async function pollUpdates() {
  if (!telegramToken) return [];

  const url = `https://api.telegram.org/bot${telegramToken}/getUpdates`;
  const params = new URLSearchParams({
    offset: String(lastUpdateId + 1),
    timeout: String(POLL_TIMEOUT_S),
    allowed_updates: JSON.stringify(["message"]),
  });

  pollAbort = new AbortController();
  let res;
  try {
    res = await fetch(`${url}?${params}`, {
      signal: pollAbort.signal,
    });
  } catch (err) {
    if (err.name === "AbortError") return [];
    throw err;
  } finally {
    pollAbort = null;
  }

  if (!res || typeof res.ok === "undefined") {
    throw new Error("invalid response object from Telegram");
  }

  if (!res.ok) {
    const body = await res.text().catch(() => "");
    // 409 = conflict — another poller is active
    if (res.status === 409) {
      log(
        "warn",
        "Telegram 409 conflict — another poller is active, backing off",
      );
    }
    throw new Error(`getUpdates failed: ${res.status} ${body}`);
  }

  const data = await res.json();
  return data.ok ? data.result || [] : [];
}

/**
 * Main polling loop. Runs continuously while sentinel is in standalone mode.
 * Implements exponential backoff on errors.
 */
async function pollLoop() {
  log("info", "polling loop started");

  while (running && polling && mode === "standalone") {
    try {
      const updates = await pollUpdates();
      consecutivePollErrors = 0;

      for (const update of updates) {
        lastUpdateId = Math.max(lastUpdateId, update.update_id);
        await handleUpdate(update);
      }
    } catch (err) {
      if (!running) break;
      consecutivePollErrors++;
      const backoff = Math.min(
        POLL_ERROR_BACKOFF_BASE_MS * Math.pow(2, consecutivePollErrors - 1),
        POLL_ERROR_BACKOFF_MAX_MS,
      );
      log(
        "warn",
        `poll error (attempt ${consecutivePollErrors}): ${err.message} — retry in ${Math.round(backoff / 1000)}s`,
      );
      await sleep(backoff);
    }
  }

  log("info", "polling loop stopped");
}

// ── Update Handler ───────────────────────────────────────────────────────────

/** Commands that the sentinel can handle without codex-monitor. */
const STANDALONE_COMMANDS = new Set([
  "/ping",
  "/status",
  "/sentinel",
  "/start",
  "/stop",
  "/help",
]);

/** Commands that require codex-monitor to be running. */
const MONITOR_REQUIRED_COMMANDS = new Set([
  "/resumetask",
  "/resume",
  "/tasks",
  "/task",
  "/sdk",
  "/model",
  "/switch",
  "/worktrees",
  "/prune",
  "/batch",
  "/threads",
  "/rebalance",
  "/logs",
  "/errors",
  "/restart",
  "/config",
]);

/**
 * Handle a single Telegram update.
 * @param {object} update
 */
async function handleUpdate(update) {
  const msg = update.message;
  if (!msg || !msg.text) return;

  const chatId = String(msg.chat?.id);
  // Security: only accept messages from the configured chat
  if (chatId !== String(telegramChatId)) {
    log("warn", `ignoring message from unauthorized chat ${chatId}`);
    return;
  }

  const text = msg.text.trim();
  const command = text.split(/\s+/)[0].toLowerCase();
  // Strip @botname suffix from commands (e.g. /ping@MyBot → /ping)
  const bareCommand = command.includes("@") ? command.split("@")[0] : command;

  commandsProcessed++;

  // ── Standalone-handled commands ──────────────────────────────────────────
  if (STANDALONE_COMMANDS.has(bareCommand)) {
    await handleStandaloneCommand(chatId, bareCommand, text);
    return;
  }

  // ── Commands requiring codex-monitor ─────────────────────────────────────
  // Either a known monitor command, free-text message, or unknown command
  log("info", `command "${bareCommand}" requires codex-monitor`);
  await handleMonitorCommand(chatId, text, bareCommand);
}

// ── Standalone Command Handlers ──────────────────────────────────────────────

/**
 * Handle commands that the sentinel can process without codex-monitor.
 * @param {string} chatId
 * @param {string} command
 * @param {string} fullText
 */
async function handleStandaloneCommand(chatId, command, fullText) {
  switch (command) {
    case "/ping":
      await handlePing(chatId);
      break;
    case "/status":
      await handleStatus(chatId);
      break;
    case "/sentinel":
      await handleSentinelInfo(chatId);
      break;
    case "/start":
      await handleStartMonitor(chatId);
      break;
    case "/stop":
      await handleStopMonitor(chatId);
      break;
    case "/help":
      await handleHelp(chatId);
      break;
    default:
      await sendTelegram(chatId, `Unknown standalone command: ${command}`);
  }
}

/**
 * /ping — Simple liveness check for the sentinel.
 * @param {string} chatId
 */
async function handlePing(chatId) {
  const monPid = readAlivePid(MONITOR_PID_FILE);
  const monStatus = monPid ? `✅ running (PID ${monPid})` : "❌ not running";
  const uptime = formatUptime(Date.now() - new Date(startedAt).getTime());
  await sendTelegram(
    chatId,
    [
      "🏓 *Pong!*",
      "",
      `Sentinel: ✅ alive (${uptime})`,
      `Mode: ${mode}`,
      `Monitor: ${monStatus}`,
      `Host: \`${os.hostname()}\``,
    ].join("\n"),
    { parseMode: "Markdown" },
  );
}

/**
 * /status — Read the cached orchestrator status file.
 * @param {string} chatId
 */
async function handleStatus(chatId) {
  try {
    if (!existsSync(STATUS_FILE)) {
      await sendTelegram(
        chatId,
        "📊 No status file found. codex-monitor may not have run yet.",
      );
      return;
    }
    const raw = await readFile(STATUS_FILE, "utf8");
    const data = JSON.parse(raw);

    const lines = ["📊 *Orchestrator Status*", ""];

    if (data.executor_mode) lines.push(`Mode: \`${data.executor_mode}\``);
    if (data.active_slots) lines.push(`Slots: \`${data.active_slots}\``);
    if (data.last_executor_sync) {
      const ago = formatUptime(
        Date.now() - new Date(data.last_executor_sync).getTime(),
      );
      lines.push(`Last sync: ${ago} ago`);
    }

    // Show active attempts
    if (data.attempts && typeof data.attempts === "object") {
      const active = Object.values(data.attempts).filter(
        (a) => a.status === "running" || a.status === "pending",
      );
      if (active.length > 0) {
        lines.push("", "*Active Tasks:*");
        for (const a of active.slice(0, 10)) {
          const title = a.task_title || a.task_id?.slice(0, 8) || "?";
          lines.push(`• ${title} — ${a.status} (${a.executor || "?"})`);
        }
      } else {
        lines.push("", "No active tasks.");
      }
    }

    await sendTelegram(chatId, lines.join("\n"), { parseMode: "Markdown" });
  } catch (err) {
    await sendTelegram(chatId, `❌ Error reading status: ${err.message}`);
  }
}

/**
 * /sentinel — Show detailed sentinel information.
 * @param {string} chatId
 */
async function handleSentinelInfo(chatId) {
  const status = getSentinelStatus();
  const lines = [
    "🛡️ *Telegram Sentinel*",
    "",
    `PID: \`${process.pid}\``,
    `Mode: ${status.mode}`,
    `Started: ${status.startedAt}`,
    `Uptime: ${formatUptime(Date.now() - new Date(status.startedAt).getTime())}`,
    `Monitor PID: ${status.monitorPid ? `\`${status.monitorPid}\`` : "none"}`,
    `Commands processed: ${status.commandsProcessed}`,
    `Commands queued: ${status.commandsQueued}`,
    `Poll errors: ${consecutivePollErrors}`,
    `Host: \`${os.hostname()}\``,
    `Platform: \`${process.platform} ${process.arch}\``,
    `Node: \`${process.version}\``,
  ];

  await sendTelegram(chatId, lines.join("\n"), { parseMode: "Markdown" });
}

/**
 * /start — Manually start codex-monitor.
 * @param {string} chatId
 */
async function handleStartMonitor(chatId) {
  const monPid = readAlivePid(MONITOR_PID_FILE);
  if (monPid) {
    await sendTelegram(
      chatId,
      `✅ codex-monitor is already running (PID ${monPid}).`,
    );
    return;
  }
  await sendTelegram(chatId, "🚀 Starting codex-monitor...");
  try {
    await ensureMonitorRunning("manual /start command");
    const pid = readAlivePid(MONITOR_PID_FILE);
    await sendTelegram(
      chatId,
      `✅ codex-monitor started${pid ? ` (PID ${pid})` : ""}.`,
    );
  } catch (err) {
    await sendTelegram(
      chatId,
      `❌ Failed to start codex-monitor: ${err.message}`,
    );
  }
}

/**
 * /stop — Manually stop codex-monitor.
 * @param {string} chatId
 */
async function handleStopMonitor(chatId) {
  const monPid = readAlivePid(MONITOR_PID_FILE);
  if (!monPid) {
    await sendTelegram(chatId, "ℹ️ codex-monitor is not running.");
    return;
  }
  await sendTelegram(chatId, `🛑 Stopping codex-monitor (PID ${monPid})...`);
  try {
    process.kill(monPid, "SIGTERM");
    // Wait for process to die
    let gone = false;
    for (let i = 0; i < 20; i++) {
      await sleep(500);
      if (!isProcessAlive(monPid)) {
        gone = true;
        break;
      }
    }
    if (!gone) {
      try {
        process.kill(monPid, "SIGKILL");
      } catch {
        /* best effort */
      }
    }
    removePidFile(MONITOR_PID_FILE);
    await sendTelegram(chatId, "✅ codex-monitor stopped.");
    // Transition to standalone mode after stopping monitor
    await transitionToStandalone("monitor manually stopped");
  } catch (err) {
    await sendTelegram(chatId, `❌ Error stopping monitor: ${err.message}`);
  }
}

/**
 * /help — Show available sentinel commands.
 * @param {string} chatId
 */
async function handleHelp(chatId) {
  const monPid = readAlivePid(MONITOR_PID_FILE);
  const monStatus = monPid ? "running" : "stopped";

  const lines = [
    "🛡️ *Sentinel Commands* (always available)",
    "",
    "/ping — Check sentinel + monitor liveness",
    "/status — Show cached orchestrator status",
    "/sentinel — Show sentinel details",
    "/start — Start codex-monitor",
    "/stop — Stop codex-monitor",
    "/help — This message",
    "",
    `Monitor is *${monStatus}*. All other commands will ${monPid ? "be forwarded to" : "auto-start"} codex-monitor.`,
  ];

  await sendTelegram(chatId, lines.join("\n"), { parseMode: "Markdown" });
}

// ── Monitor-Required Command Handling ────────────────────────────────────────

/**
 * Handle commands that need codex-monitor. Starts the monitor if not running
 * and queues the command for replay once it's healthy.
 * @param {string} chatId
 * @param {string} text
 * @param {string} command
 */
async function handleMonitorCommand(chatId, text, command) {
  const monPid = readAlivePid(MONITOR_PID_FILE);

  if (monPid) {
    // Monitor is running but sentinel is somehow in standalone mode — this
    // can happen briefly during transitions. Queue the command for the
    // monitor to pick up.
    queueCommand(chatId, text);
    await writeCommandQueueFile();
    log("info", "monitor running — queued command for replay");
    return;
  }

  // Monitor is not running — notify user and start it
  await sendTelegram(
    chatId,
    "⏳ codex-monitor was not running — starting now...",
  );

  // Queue this command for replay after startup
  queueCommand(chatId, text);

  try {
    await ensureMonitorRunning(`command: ${command}`);
    // Write the command queue to disk so monitor can read it on startup
    await writeCommandQueueFile();
    log(
      "info",
      `monitor started — ${commandQueue.length} command(s) queued for replay`,
    );
  } catch (err) {
    await sendTelegram(
      chatId,
      `❌ Failed to start codex-monitor: ${err.message}\n\nYour command was not processed.`,
    );
    // Clear the failed commands
    commandQueue = [];
  }
}

// ── Command Queue ────────────────────────────────────────────────────────────

/**
 * Add a command to the replay queue.
 * @param {string | number} chatId
 * @param {string} text
 */
function queueCommand(chatId, text) {
  // Evict stale commands
  const now = Date.now();
  commandQueue = commandQueue.filter(
    (c) => now - c.timestamp < COMMAND_QUEUE_TTL_MS,
  );

  // Enforce max queue size
  if (commandQueue.length >= COMMAND_QUEUE_MAX_SIZE) {
    log(
      "warn",
      `command queue full (${COMMAND_QUEUE_MAX_SIZE}), dropping oldest`,
    );
    commandQueue.shift();
  }

  commandQueue.push({ chatId: String(chatId), text, timestamp: now });
}

/**
 * Write the command queue to a JSON file for codex-monitor to read.
 * @returns {Promise<void>}
 */
async function writeCommandQueueFile() {
  try {
    mkdirSync(dirname(SENTINEL_COMMAND_QUEUE_FILE), { recursive: true });
    await writeFile(
      SENTINEL_COMMAND_QUEUE_FILE,
      JSON.stringify(commandQueue, null, 2),
      "utf8",
    );
  } catch (err) {
    log("warn", `failed to write command queue: ${err.message}`);
  }
}

/**
 * Get the current command queue.
 * @returns {Array<{ chatId: string, text: string, timestamp: number }>}
 */
export function getQueuedCommands() {
  return [...commandQueue];
}

// ── Monitor Lifecycle ────────────────────────────────────────────────────────

/**
 * Check if the codex-monitor process is running.
 * @returns {boolean}
 */
export function isMonitorRunning() {
  return readAlivePid(MONITOR_PID_FILE) !== null;
}

/**
 * Ensure codex-monitor is running. If not, start it and wait until it's healthy.
 * Returns immediately if monitor is already running. Coalesces concurrent calls
 * so only one monitor start happens at a time.
 * @param {string} reason - Human-readable reason for starting the monitor.
 * @returns {Promise<void>}
 */
export async function ensureMonitorRunning(reason) {
  // Already running
  if (readAlivePid(MONITOR_PID_FILE)) return;

  // Another call is already starting the monitor — piggyback on it
  if (monitorStartPromise) {
    log("info", `waiting for in-progress monitor start (reason: ${reason})`);
    return monitorStartPromise;
  }

  monitorStartPromise = startAndWaitForMonitor(reason);
  try {
    await monitorStartPromise;
  } finally {
    monitorStartPromise = null;
  }
}

/**
 * Start codex-monitor as a detached background process and wait for it to
 * become healthy (PID file written and process alive).
 * @param {string} reason
 * @returns {Promise<void>}
 */
async function startAndWaitForMonitor(reason) {
  log("info", `starting codex-monitor (reason: ${reason})`);

  // If sentinel is currently polling, release the sentinel lock.
  // The monitor's telegram-bot.mjs will acquire its own poll lock.
  const wasPolling = polling;
  if (wasPolling) {
    polling = false;
    if (pollAbort) {
      try {
        pollAbort.abort();
      } catch {
        /* ok */
      }
    }
    await releaseSentinelPollLock();
    log("info", "released sentinel poll lock for monitor startup");
  }

  // Ensure log directory exists for daemon output
  const daemonLog = resolve(__dirname, "logs", "daemon.log");
  try {
    mkdirSync(dirname(daemonLog), { recursive: true });
  } catch {
    /* ok */
  }

  // Start cli.mjs as a detached daemon child
  const child = spawn(
    process.execPath,
    [
      "--max-old-space-size=4096",
      resolve(__dirname, "cli.mjs"),
      "--daemon-child",
    ],
    {
      detached: true,
      stdio: "ignore",
      env: { ...process.env, CODEX_MONITOR_DAEMON: "1" },
      cwd: repoRoot,
    },
  );

  child.on("error", (err) => {
    log("error", `monitor spawn error: ${err.message}`);
  });

  child.unref();

  const spawnedPid = child.pid;
  if (!spawnedPid) {
    throw new Error("codex-monitor failed to spawn (no PID)");
  }

  log("info", `monitor spawned (PID ${spawnedPid}), waiting for health...`);

  // Wait for the monitor to become healthy (PID file written + process alive)
  const deadline = Date.now() + MONITOR_START_TIMEOUT_MS;
  while (Date.now() < deadline) {
    await sleep(MONITOR_HEALTH_POLL_MS);

    const alivePid = readAlivePid(MONITOR_PID_FILE);
    if (alivePid) {
      log("info", `monitor is healthy (PID ${alivePid})`);
      // Transition to companion mode
      await transitionToCompanion(alivePid);
      return;
    }

    // Check if spawned process died prematurely
    if (!isProcessAlive(spawnedPid)) {
      throw new Error(
        `codex-monitor process died during startup (PID ${spawnedPid})`,
      );
    }
  }

  throw new Error(
    `codex-monitor did not become healthy within ${MONITOR_START_TIMEOUT_MS / 1000}s`,
  );
}

// ── Mode Transitions ─────────────────────────────────────────────────────────

/**
 * Transition to standalone mode. Starts polling for Telegram updates directly.
 * @param {string} reason
 */
async function transitionToStandalone(reason) {
  if (mode === "standalone" && polling) {
    log("debug", `already in standalone mode (${reason})`);
    return;
  }

  log("info", `transitioning to standalone mode: ${reason}`);
  mode = "standalone";

  // Check if the main bot poll lock is held by a live process
  const mainBotPolling = await isMainBotPolling();
  if (mainBotPolling) {
    log("info", "main bot is still polling — skipping sentinel poll start");
    return;
  }

  // Acquire sentinel poll lock and start polling
  const lockAcquired = await acquireSentinelPollLock();
  if (!lockAcquired) {
    log(
      "warn",
      "failed to acquire sentinel poll lock — another sentinel may be running",
    );
    return;
  }

  // Clear stale updates before starting the loop
  try {
    const stale = await pollUpdates();
    for (const u of stale) {
      lastUpdateId = Math.max(lastUpdateId, u.update_id);
    }
    if (stale.length > 0) {
      log("info", `skipped ${stale.length} stale updates`);
    }
  } catch {
    /* best effort */
  }

  polling = true;
  consecutivePollErrors = 0;

  // Fire polling loop (non-blocking)
  pollLoop().catch((err) => {
    log("error", `poll loop crashed: ${err.message}`);
    polling = false;
  });

  await writeHeartbeat();
}

/**
 * Transition to companion mode. Stops polling and lets telegram-bot.mjs handle it.
 * @param {number} monitorPid
 */
async function transitionToCompanion(monitorPid) {
  log("info", `transitioning to companion mode (monitor PID ${monitorPid})`);
  mode = "companion";

  // Stop polling if active
  polling = false;
  if (pollAbort) {
    try {
      pollAbort.abort();
    } catch {
      /* ok */
    }
  }
  await releaseSentinelPollLock();

  await writeHeartbeat();
}

/**
 * Check if the main telegram-bot.mjs poll lock is held by a live process.
 * @returns {Promise<boolean>}
 */
async function isMainBotPolling() {
  try {
    if (!existsSync(MONITOR_POLL_LOCK_FILE)) return false;
    const raw = await readFile(MONITOR_POLL_LOCK_FILE, "utf8");
    if (!raw || !raw.trim()) return false;
    const data = JSON.parse(raw);
    const pid = Number(data?.pid);
    return isProcessAlive(pid);
  } catch {
    return false;
  }
}

// ── Health Monitoring ────────────────────────────────────────────────────────

/**
 * Periodic health check for codex-monitor. Runs every HEALTH_CHECK_INTERVAL_MS.
 */
async function healthCheck() {
  const monPid = readAlivePid(MONITOR_PID_FILE);

  if (mode === "companion") {
    if (!monPid) {
      // Monitor died while in companion mode — send crash notification and go standalone
      log("warn", "monitor process died — transitioning to standalone");
      removePidFile(MONITOR_PID_FILE);

      // Notify user
      const host = os.hostname();
      const tag = projectName ? `[${projectName}]` : "";
      await sendTelegram(
        telegramChatId,
        [
          `🔥 ${tag} codex-monitor crashed`,
          "",
          `Host: \`${host}\``,
          `Time: ${new Date().toISOString()}`,
          "",
          "Sentinel is switching to standalone mode. Use /start to restart codex-monitor.",
        ].join("\n"),
        { parseMode: "Markdown" },
      );

      await transitionToStandalone("monitor process died");
    }
  } else if (mode === "standalone") {
    if (monPid) {
      // Monitor appeared while in standalone mode (started externally)
      log(
        "info",
        `monitor detected (PID ${monPid}) — switching to companion mode`,
      );
      await transitionToCompanion(monPid);
    } else {
      // Check if main bot has acquired the poll lock (edge case: monitor starting up)
      const mainPolling = await isMainBotPolling();
      if (mainPolling && polling) {
        log("info", "main bot is polling — stopping sentinel polling");
        polling = false;
        if (pollAbort) {
          try {
            pollAbort.abort();
          } catch {
            /* ok */
          }
        }
        await releaseSentinelPollLock();
      } else if (!mainPolling && !polling) {
        // Neither is polling — sentinel should resume
        log("info", "no poller active — resuming sentinel polling");
        await transitionToStandalone("no active poller detected");
      }
    }
  }

  // Clean up stale PID files
  const sentinelPid = readAlivePid(SENTINEL_PID_FILE);
  if (sentinelPid && sentinelPid !== process.pid) {
    // Another sentinel is alive — we shouldn't be running
    log(
      "warn",
      `another sentinel is alive (PID ${sentinelPid}) — stopping this instance`,
    );
    stopSentinel();
    return;
  }

  await writeHeartbeat();
}

// ── Heartbeat ────────────────────────────────────────────────────────────────

/**
 * Write the sentinel heartbeat file.
 * @returns {Promise<void>}
 */
async function writeHeartbeat() {
  /** @type {import("./telegram-sentinel.mjs").SentinelHeartbeat} */
  const heartbeat = {
    pid: process.pid,
    startedAt,
    mode,
    monitorPid: readAlivePid(MONITOR_PID_FILE),
    lastCheck: new Date().toISOString(),
    commandsQueued: commandQueue.length,
    commandsProcessed,
  };

  try {
    mkdirSync(dirname(SENTINEL_HEARTBEAT_FILE), { recursive: true });
    await writeFile(
      SENTINEL_HEARTBEAT_FILE,
      JSON.stringify(heartbeat, null, 2),
      "utf8",
    );
  } catch (err) {
    log("warn", `heartbeat write failed: ${err.message}`);
  }
}

// ── Public API ───────────────────────────────────────────────────────────────

/**
 * Start the Telegram sentinel. This is the main entry point.
 *
 * @param {object} [options]
 * @param {boolean} [options.skipExistingCheck] - Skip checking for an existing sentinel.
 * @returns {Promise<void>}
 */
export async function startSentinel(options = {}) {
  if (running) {
    log("warn", "sentinel is already running");
    return;
  }

  initEnv();

  if (!telegramToken || !telegramChatId) {
    log(
      "error",
      "cannot start sentinel: TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID not configured",
    );
    console.error(
      `${TAG} Set these in scripts/codex-monitor/.env or as environment variables.`,
    );
    process.exit(1);
  }

  // Ensure cache directory exists
  mkdirSync(cacheDir, { recursive: true });
  mkdirSync(dirname(MONITOR_PID_FILE), { recursive: true });

  // Check for existing sentinel
  if (!options.skipExistingCheck) {
    const existingPid = readAlivePid(SENTINEL_PID_FILE);
    if (existingPid && existingPid !== process.pid) {
      console.error(
        `${TAG} Another sentinel is already running (PID ${existingPid}). Use --stop first.`,
      );
      process.exit(1);
    }
  }

  running = true;
  startedAt = new Date().toISOString();
  writePidFile(SENTINEL_PID_FILE, process.pid);

  log("info", `sentinel started (PID ${process.pid})`);

  // Determine initial mode
  const monPid = readAlivePid(MONITOR_PID_FILE);
  if (monPid) {
    log(
      "info",
      `codex-monitor already running (PID ${monPid}) — starting in companion mode`,
    );
    await transitionToCompanion(monPid);
  } else {
    log("info", "codex-monitor not running — starting in standalone mode");
    await transitionToStandalone("initial startup");
  }

  // Set up periodic health checks
  healthCheckTimer = setInterval(() => {
    healthCheck().catch((err) => {
      log("error", `health check error: ${err.message}`);
    });
  }, HEALTH_CHECK_INTERVAL_MS);
  if (healthCheckTimer.unref) healthCheckTimer.unref();

  // Set up periodic heartbeat writes
  heartbeatTimer = setInterval(() => {
    writeHeartbeat().catch(() => {});
  }, HEALTH_CHECK_INTERVAL_MS);
  if (heartbeatTimer.unref) heartbeatTimer.unref();

  // Initial heartbeat
  await writeHeartbeat();

  // Register shutdown handlers
  const shutdown = () => {
    log("info", "received shutdown signal");
    stopSentinel();
    process.exit(0);
  };
  process.on("SIGINT", shutdown);
  process.on("SIGTERM", shutdown);
  process.on("uncaughtException", (err) => {
    log("error", `uncaught exception: ${err.message}\n${err.stack}`);
    // Attempt crash notification
    sendTelegram(
      telegramChatId,
      `🛡️❌ Sentinel crashed: ${err.message}\nHost: \`${os.hostname()}\``,
      { parseMode: "Markdown" },
    ).catch(() => {});
    stopSentinel();
    process.exit(1);
  });
  process.on("unhandledRejection", (reason) => {
    log("error", `unhandled rejection: ${reason}`);
  });
}

/**
 * Stop the sentinel gracefully. Cleans up timers, locks, and PID files.
 */
export function stopSentinel() {
  if (!running) return;
  running = false;
  polling = false;

  // Abort any pending poll
  if (pollAbort) {
    try {
      pollAbort.abort();
    } catch {
      /* ok */
    }
  }

  // Clear timers
  if (healthCheckTimer) {
    clearInterval(healthCheckTimer);
    healthCheckTimer = null;
  }
  if (heartbeatTimer) {
    clearInterval(heartbeatTimer);
    heartbeatTimer = null;
  }

  // Release locks and PID files
  releaseSentinelPollLock().catch(() => {});
  removePidFile(SENTINEL_PID_FILE);

  // Clean up heartbeat file
  try {
    if (existsSync(SENTINEL_HEARTBEAT_FILE))
      unlinkSync(SENTINEL_HEARTBEAT_FILE);
  } catch {
    /* best effort */
  }

  log("info", "sentinel stopped");
}

/**
 * Get the current sentinel status.
 * @returns {SentinelStatus}
 */
export function getSentinelStatus() {
  return {
    pid: process.pid,
    running,
    startedAt,
    mode,
    monitorPid: readAlivePid(MONITOR_PID_FILE),
    polling,
    commandsQueued: commandQueue.length,
    commandsProcessed,
    consecutivePollErrors,
    uptime: Date.now() - new Date(startedAt).getTime(),
  };
}

// ── Logging ──────────────────────────────────────────────────────────────────

/**
 * Simple structured logger. All output goes to stdout/stderr with a tag prefix.
 * @param {"info" | "warn" | "error" | "debug"} level
 * @param {string} message
 */
function log(level, message) {
  const timestamp = new Date().toISOString();
  const prefix = `${timestamp} ${TAG}`;
  switch (level) {
    case "error":
      console.error(`${prefix} ERROR: ${message}`);
      break;
    case "warn":
      console.warn(`${prefix} WARN: ${message}`);
      break;
    case "debug":
      if (process.env.SENTINEL_DEBUG === "1") {
        console.log(`${prefix} DEBUG: ${message}`);
      }
      break;
    default:
      console.log(`${prefix} ${message}`);
  }
}

// ── Utility ──────────────────────────────────────────────────────────────────

/**
 * Format a duration in milliseconds to a human-readable string.
 * @param {number} ms
 * @returns {string}
 */
function formatUptime(ms) {
  if (ms < 0) ms = 0;
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) return `${days}d ${hours % 24}h ${minutes % 60}m`;
  if (hours > 0) return `${hours}h ${minutes % 60}m`;
  if (minutes > 0) return `${minutes}m ${seconds % 60}s`;
  return `${seconds}s`;
}

/**
 * Sleep for the given number of milliseconds.
 * @param {number} ms
 * @returns {Promise<void>}
 */
function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// ── Type Definitions (JSDoc) ─────────────────────────────────────────────────

/**
 * @typedef {object} SentinelHeartbeat
 * @property {number} pid
 * @property {string} startedAt
 * @property {"standalone" | "companion"} mode
 * @property {number | null} monitorPid
 * @property {string} lastCheck
 * @property {number} commandsQueued
 * @property {number} commandsProcessed
 */

/**
 * @typedef {object} SentinelStatus
 * @property {number} pid
 * @property {boolean} running
 * @property {string} startedAt
 * @property {"standalone" | "companion"} mode
 * @property {number | null} monitorPid
 * @property {boolean} polling
 * @property {number} commandsQueued
 * @property {number} commandsProcessed
 * @property {number} consecutivePollErrors
 * @property {number} uptime
 */

// ── CLI Entry Point ──────────────────────────────────────────────────────────

const isDirectExecution = (() => {
  try {
    const thisFile = fileURLToPath(import.meta.url);
    const argv1 = process.argv[1];
    if (!argv1) return false;
    // Normalize paths for comparison (Windows backslash vs posix)
    const normalizedThis = thisFile.replace(/\\/g, "/").toLowerCase();
    const normalizedArgv = resolve(argv1).replace(/\\/g, "/").toLowerCase();
    return normalizedThis === normalizedArgv;
  } catch {
    return false;
  }
})();

if (isDirectExecution) {
  const args = process.argv.slice(2);

  if (args.includes("--help") || args.includes("-h")) {
    console.log(`
  telegram-sentinel — Always-on Telegram command listener for VirtEngine

  USAGE
    node telegram-sentinel.mjs [options]

  OPTIONS
    --stop       Stop a running sentinel
    --status     Check sentinel status
    --help       Show this help

  ENVIRONMENT
    TELEGRAM_BOT_TOKEN    Telegram bot token (or set in .env)
    TELEGRAM_CHAT_ID      Authorized chat ID (or set in .env)
    SENTINEL_DEBUG=1      Enable debug logging

  The sentinel monitors codex-monitor and handles Telegram commands
  even when the main process is not running.
    `);
    process.exit(0);
  }

  if (args.includes("--stop")) {
    const pid = readAlivePid(SENTINEL_PID_FILE);
    if (!pid) {
      console.log("  No sentinel running.");
      removePidFile(SENTINEL_PID_FILE);
      process.exit(0);
    }
    console.log(`  Stopping sentinel (PID ${pid})...`);
    try {
      process.kill(pid, "SIGTERM");
      let gone = false;
      for (let i = 0; i < 20; i++) {
        await sleep(500);
        if (!isProcessAlive(pid)) {
          gone = true;
          break;
        }
      }
      if (!gone) {
        try {
          process.kill(pid, "SIGKILL");
        } catch {
          /* ok */
        }
      }
      removePidFile(SENTINEL_PID_FILE);
      console.log("  ✓ Sentinel stopped.");
    } catch (err) {
      console.error(`  Failed: ${err.message}`);
      process.exit(1);
    }
    process.exit(0);
  }

  if (args.includes("--status")) {
    const pid = readAlivePid(SENTINEL_PID_FILE);
    if (pid) {
      console.log(`  Sentinel is running (PID ${pid})`);
      try {
        if (existsSync(SENTINEL_HEARTBEAT_FILE)) {
          const hb = JSON.parse(readFileSync(SENTINEL_HEARTBEAT_FILE, "utf8"));
          console.log(`  Mode: ${hb.mode}`);
          console.log(`  Monitor PID: ${hb.monitorPid || "none"}`);
          console.log(`  Last check: ${hb.lastCheck}`);
          console.log(`  Commands processed: ${hb.commandsProcessed}`);
        }
      } catch {
        /* best effort */
      }
    } else {
      console.log("  Sentinel is not running.");
      removePidFile(SENTINEL_PID_FILE);
    }
    process.exit(0);
  }

  // Default: start sentinel
  startSentinel().catch((err) => {
    console.error(`${TAG} Fatal: ${err.message}`);
    process.exit(1);
  });
}
