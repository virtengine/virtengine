/**
 * telegram-bot.mjs — Two-way Telegram ↔ primary agent for VirtEngine monitor.
 *
 * Polls Telegram Bot API for incoming messages, routes slash commands to
 * built-in handlers, and forwards free-text to the persistent primary agent.
 *
 * Architecture:
 *   Telegram → getUpdates long-poll → handleUpdate()
 *     ├─ /command → built-in handler (fast, no agent)
 *     └─ free-text → PrimaryAgent.exec() → response back to Telegram
 *
 * Security: Only accepts messages from the configured TELEGRAM_CHAT_ID.
 */

import { execSync, spawnSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import { readFile, readdir, stat, unlink, writeFile } from "node:fs/promises";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";
import {
  execPrimaryPrompt,
  isPrimaryBusy,
  getPrimaryAgentInfo,
  resetPrimaryAgent,
  initPrimaryAgent,
  steerPrimaryPrompt,
  getPrimaryAgentName,
  switchPrimaryAgent,
} from "./primary-agent.mjs";
import {
  getPoolSdkName,
  setPoolSdk,
  resetPoolSdkCache,
  getAvailableSdks,
  getActiveThreads,
  clearThreadRegistry,
  invalidateThread,
} from "./agent-pool.mjs";
import {
  getKanbanAdapter,
  setKanbanBackend,
  getAvailableBackends,
  getKanbanBackendName,
} from "./kanban-adapter.mjs";
import {
  getWorktreeManager,
  listActiveWorktrees as listManagedWorktrees,
  pruneStaleWorktrees,
  getWorktreeStats,
} from "./worktree-manager.mjs";
import { loadExecutorConfig } from "./config.mjs";
import {
  loadWorkspaceRegistry,
  formatRegistryDiagnostics,
  getDefaultModelPriority,
  getLocalWorkspace,
} from "./workspace-registry.mjs";
import {
  claimSharedWorkspace,
  formatSharedWorkspaceDetail,
  formatSharedWorkspaceSummary,
  getSharedAvailabilityMap,
  loadSharedWorkspaceRegistry as loadSharedRegistry,
  releaseSharedWorkspace,
  resolveSharedWorkspace,
  sweepExpiredLeases as sweepSharedLeases,
} from "./shared-workspace-registry.mjs";
import {
  buildLocalPresence,
  formatCoordinatorSummary,
  formatPresenceMessage,
  formatPresenceSummary,
  initPresence,
  notePresence,
  parsePresenceMessage,
} from "./presence.mjs";

const __dirname = resolve(fileURLToPath(new URL(".", import.meta.url)));
const repoRoot = resolve(__dirname, "..", "..");
const statusPath = resolve(repoRoot, ".cache", "ve-orchestrator-status.json");
const telegramPollLockPath = resolve(
  repoRoot,
  ".cache",
  "telegram-getupdates.lock",
);
const liveDigestStatePath = resolve(repoRoot, ".cache", "ve-live-digest.json");

// ── Configuration ────────────────────────────────────────────────────────────

const telegramToken = process.env.TELEGRAM_BOT_TOKEN;
const telegramChatId = process.env.TELEGRAM_CHAT_ID;
const POLL_TIMEOUT_S = 30; // long-poll timeout
const MAX_MESSAGE_LEN = 4000; // Telegram max is 4096, leave margin
const POLL_ERROR_BACKOFF_MS = 5000;
const AGENT_TIMEOUT_MS = (() => {
  const minRaw = Number(process.env.TELEGRAM_AGENT_TIMEOUT_MIN || "");
  if (Number.isFinite(minRaw) && minRaw > 0) return minRaw * 60 * 1000;
  const raw = Number(
    process.env.TELEGRAM_AGENT_TIMEOUT_MS ||
      process.env.PRIMARY_AGENT_TIMEOUT_MS ||
      process.env.INTERNAL_EXECUTOR_TIMEOUT_MS ||
      "",
  );
  if (Number.isFinite(raw) && raw > 0) return raw;
  return 90 * 60 * 1000; // 90 min default
})();
let telegramPollLockHeld = false;
const presenceIntervalSec = Number(
  process.env.TELEGRAM_PRESENCE_INTERVAL_SEC || "60",
);
const presenceTtlSec = Number(
  process.env.TELEGRAM_PRESENCE_TTL_SEC || String(presenceIntervalSec * 3),
);
const presenceDisabled = ["1", "true", "yes"].includes(
  String(process.env.TELEGRAM_PRESENCE_DISABLED || "").toLowerCase(),
);
const presenceSilent = ["1", "true", "yes"].includes(
  String(process.env.TELEGRAM_PRESENCE_SILENT || "").toLowerCase(),
);
const presenceOnlyOnChange = ["1", "true", "yes"].includes(
  String(process.env.TELEGRAM_PRESENCE_ONLY_ON_CHANGE || "true").toLowerCase(),
);
const presenceChatId = process.env.TELEGRAM_PRESENCE_CHAT_ID;
const presenceTtlMs = Number.isFinite(presenceTtlSec)
  ? Math.max(0, presenceTtlSec * 1000)
  : 0;

console.log(`[telegram-bot] agent timeout set to ${Math.round(AGENT_TIMEOUT_MS / 60000)} min`);

// ── Message Batching Configuration ──────────────────────────────────────────
const batchingEnabled = !["0", "false", "no"].includes(
  String(process.env.TELEGRAM_BATCH_NOTIFICATIONS || "true").toLowerCase(),
);
const batchIntervalSec = Number(
  process.env.TELEGRAM_BATCH_INTERVAL_SEC || "300",
); // 5 minutes default
const batchMaxSize = Number(process.env.TELEGRAM_BATCH_MAX_SIZE || "50");
// Priority threshold: only messages >= this priority bypass batching (1=critical, 2=error, 3=warning, 4=info, 5=debug)
const immediateThreshold = Number(
  process.env.TELEGRAM_IMMEDIATE_PRIORITY || "1",
);

// ── Live Digest Configuration ──────────────────────────────────────────────
// Instead of batching and flushing, we maintain a single "live" Telegram message
// per digest window that gets continuously edited as new events arrive.
const liveDigestEnabled = !["0", "false", "no"].includes(
  String(process.env.TELEGRAM_LIVE_DIGEST || "true").toLowerCase(),
);
const liveDigestWindowSec = Number(
  process.env.TELEGRAM_LIVE_DIGEST_WINDOW_SEC || "1200",
); // 20 minutes default
const liveDigestEditDebounceMs = Number(
  process.env.TELEGRAM_LIVE_DIGEST_DEBOUNCE_MS || "3000",
); // 3 second debounce on edits

function canSignalProcess(pid) {
  if (!Number.isFinite(pid) || pid <= 0) return false;
  try {
    process.kill(pid, 0);
    return true;
  } catch {
    return false;
  }
}

async function acquireTelegramPollLock(owner) {
  if (telegramPollLockHeld) return true;
  try {
    const payload = JSON.stringify(
      { owner, pid: process.pid, started_at: new Date().toISOString() },
      null,
      2,
    );
    await writeFile(telegramPollLockPath, payload, { flag: "wx" });
    telegramPollLockHeld = true;
    return true;
  } catch (err) {
    if (err && err.code === "EEXIST") {
      try {
        const raw = await readFile(telegramPollLockPath, "utf8");
        if (!raw || !raw.trim()) {
          // Empty/corrupt lock file — treat as stale
          await unlink(telegramPollLockPath);
          return await acquireTelegramPollLock(owner);
        }
        const data = JSON.parse(raw);
        const pid = Number(data?.pid);
        if (!canSignalProcess(pid)) {
          await unlink(telegramPollLockPath);
          return await acquireTelegramPollLock(owner);
        }
      } catch {
        // Lock file is corrupt/unparseable — remove and retry
        try {
          await unlink(telegramPollLockPath);
        } catch {
          /* ignore */
        }
        return await acquireTelegramPollLock(owner);
      }
    }
    return false;
  }
}

async function releaseTelegramPollLock() {
  if (!telegramPollLockHeld) return;
  telegramPollLockHeld = false;
  try {
    await unlink(telegramPollLockPath);
  } catch {
    /* best effort */
  }
}

// ── State ────────────────────────────────────────────────────────────────────

let lastUpdateId = 0;
let polling = false;
let pollAbort = null;
let presenceReady = false;
let workspaceRegistryPromise = null;
let localWorkspaceCache = null;

// ── Agent session state (for follow-up steering & bottom-pinning) ────────────

let activeAgentSession = null; // { chatId, messageId, taskPreview, abortController, followUpQueue, ... }
let agentMessageId = null; // current agent streaming message ID
let agentChatId = null; // chat where agent is running

// ── Queues ──────────────────────────────────────────────────────────────────

let fastCommandQueue = Promise.resolve();
let commandQueue = Promise.resolve();
let agentQueue = Promise.resolve();

function enqueueFastCommand(task) {
  fastCommandQueue = fastCommandQueue.then(task).catch((err) => {
    console.error(`[telegram-bot] fast command error: ${err.message || err}`);
  });
}

function enqueueCommand(task) {
  commandQueue = commandQueue.then(task).catch((err) => {
    console.error(`[telegram-bot] command error: ${err.message || err}`);
  });
}

function enqueueAgentTask(task) {
  agentQueue = agentQueue.then(task).catch((err) => {
    console.error(`[telegram-bot] agent error: ${err.message || err}`);
  });
}

async function getWorkspaceRegistryCached() {
  if (!workspaceRegistryPromise) {
    workspaceRegistryPromise = loadWorkspaceRegistry();
  }
  return workspaceRegistryPromise;
}

async function getLocalWorkspaceContext() {
  if (localWorkspaceCache) return localWorkspaceCache;
  const loaded = await getWorkspaceRegistryCached();
  const registry = loaded.registry || loaded;
  localWorkspaceCache = getLocalWorkspace(
    registry,
    process.env.VE_WORKSPACE_ID || "",
  );
  return localWorkspaceCache;
}

// ── External refs (injected by monitor.mjs) ──────────────────────────────────

let _sendTelegramMessage = null; // injected from monitor.mjs
let _readStatusData = null;
let _readStatusSummary = null;
let _getCurrentChild = null;
let _startProcess = null;
let _getVibeKanbanUrl = null;
let _fetchVk = null;
let _getRepoRoot = null;
let _startFreshSession = null;
let _attemptFreshSessionRetry = null;
let _buildRetryPrompt = null;
let _getActiveAttemptInfo = null;
let _triggerTaskPlanner = null;
let _reconcileTaskStatuses = null;
let _onDigestSealed = null;
let _getAnomalyReport = null;
let _getInternalExecutor = null;
let _getExecutorMode = null;

/**
 * Inject monitor.mjs functions so the bot can send messages and read status.
 * Call this BEFORE startTelegramBot().
 */
export function injectMonitorFunctions({
  sendTelegramMessage,
  readStatusData,
  readStatusSummary,
  getCurrentChild,
  startProcess,
  getVibeKanbanUrl,
  fetchVk,
  getRepoRoot,
  startFreshSession,
  attemptFreshSessionRetry,
  buildRetryPrompt,
  getActiveAttemptInfo,
  triggerTaskPlanner,
  reconcileTaskStatuses,
  onDigestSealed,
  getAnomalyReport,
  getInternalExecutor,
  getExecutorMode,
}) {
  _sendTelegramMessage = sendTelegramMessage;
  _readStatusData = readStatusData;
  _readStatusSummary = readStatusSummary;
  _getCurrentChild = getCurrentChild;
  _startProcess = startProcess;
  _getVibeKanbanUrl = getVibeKanbanUrl;
  _fetchVk = fetchVk;
  _getRepoRoot = getRepoRoot;
  _startFreshSession = startFreshSession;
  _attemptFreshSessionRetry = attemptFreshSessionRetry;
  _buildRetryPrompt = buildRetryPrompt;
  _getActiveAttemptInfo = getActiveAttemptInfo;
  _triggerTaskPlanner = triggerTaskPlanner;
  _reconcileTaskStatuses = reconcileTaskStatuses;
  _onDigestSealed = onDigestSealed || null;
  _getAnomalyReport = getAnomalyReport || null;
  _getInternalExecutor = getInternalExecutor || null;
  _getExecutorMode = getExecutorMode || null;
}

/**
 * Called by monitor.mjs when a notification is sent while the agent is streaming.
 * Re-sends the agent message so it stays at the bottom of the chat.
 */
export async function bumpAgentMessage() {
  if (!activeAgentSession || activeAgentSession.background) return;
  if (!agentMessageId || !agentChatId) return;
  try {
    // Delete the old message
    await deleteDirect(agentChatId, agentMessageId);
  } catch {
    /* best effort */
  }
  // Re-send at bottom
  const session = activeAgentSession;
  const msg = buildStreamMessage({
    taskPreview: session.taskPreview,
    actionLog: session.actionLog,
    currentThought: session.currentThought,
    totalActions: session.totalActions,
    phase: session.phase,
    finalResponse: null,
  });
  const newId = await sendDirect(agentChatId, msg);
  if (newId) {
    agentMessageId = newId;
    session.messageId = newId;
  }
}

/**
 * Check if agent is active (for external callers like monitor.mjs).
 */
export function isAgentActive() {
  return !!activeAgentSession;
}

// ── Telegram API Helpers ─────────────────────────────────────────────────────

async function sendReply(chatId, text, options = {}) {
  // If monitor's sendTelegramMessage is available, use it (handles dedup & history)
  if (_sendTelegramMessage) {
    // Bypass dedup for direct replies
    await sendDirect(chatId, text, options);
    return;
  }
  await sendDirect(chatId, text, options);
}

async function sendDirect(chatId, text, options = {}) {
  if (!telegramToken) return null;

  // Split long messages
  const chunks = splitMessage(text, MAX_MESSAGE_LEN);
  let lastMessageId = null;
  for (const chunk of chunks) {
    const url = `https://api.telegram.org/bot${telegramToken}/sendMessage`;
    const payload = {
      chat_id: chatId,
      text: chunk,
    };
    if (options.parseMode) {
      payload.parse_mode = options.parseMode;
    }
    if (options.silent) {
      payload.disable_notification = true;
    }
    payload.disable_web_page_preview = true;

    try {
      const res = await fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!res.ok) {
        const body = await res.text();
        console.warn(`[telegram-bot] send failed: ${res.status} ${body}`);
        // If HTML parse mode fails, retry as plain text
        if (options.parseMode && res.status === 400) {
          return sendDirect(chatId, chunk, {
            ...options,
            parseMode: undefined,
          });
        }
      } else {
        const data = await res.json();
        if (data.ok && data.result?.message_id) {
          lastMessageId = data.result.message_id;
        }
      }
    } catch (err) {
      console.warn(`[telegram-bot] send error: ${err.message}`);
    }
  }
  return lastMessageId;
}

/**
 * Edit an existing Telegram message in-place.
 * Falls back to sending a new message if the edit fails (message too old, etc.).
 */
async function editDirect(chatId, messageId, text, options = {}) {
  if (!telegramToken || !messageId) return messageId;

  // Telegram editMessageText has 4096 char limit — truncate if needed
  const truncated =
    text.length > MAX_MESSAGE_LEN
      ? text.slice(0, MAX_MESSAGE_LEN - 20) + "\n\n…(truncated)"
      : text;

  const url = `https://api.telegram.org/bot${telegramToken}/editMessageText`;
  const payload = {
    chat_id: chatId,
    message_id: messageId,
    text: truncated,
    disable_web_page_preview: true,
  };
  if (options.parseMode) {
    payload.parse_mode = options.parseMode;
  }

  try {
    const res = await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const body = await res.text();
      // "message is not modified" is fine — content didn't change
      if (body.includes("message is not modified")) return messageId;
      // "message can't be edited" — send new message instead
      if (
        body.includes("can't be edited") ||
        body.includes("MESSAGE_ID_INVALID")
      ) {
        console.warn(`[telegram-bot] edit failed, sending new message`);
        return await sendDirect(chatId, truncated, options);
      }
      console.warn(`[telegram-bot] edit failed: ${res.status} ${body}`);
      // For HTML parse errors, retry without parse mode
      if (options.parseMode && res.status === 400) {
        return editDirect(chatId, messageId, truncated, {
          ...options,
          parseMode: undefined,
        });
      }
    }
    return messageId;
  } catch (err) {
    console.warn(`[telegram-bot] edit error: ${err.message}`);
    return messageId;
  }
}

// ── Action Summarizer ────────────────────────────────────────────────────────

/**
 * Delete a Telegram message. Best-effort, failures are silently ignored.
 */
async function deleteDirect(chatId, messageId) {
  if (!telegramToken || !messageId) return;
  const url = `https://api.telegram.org/bot${telegramToken}/deleteMessage`;
  try {
    await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ chat_id: chatId, message_id: messageId }),
    });
  } catch {
    /* best effort */
  }
}

/**
 * Extract a short filename from a path (last 2 segments for context).
 */
function shortPath(p) {
  if (!p) return "";
  const parts = p.replace(/\\/g, "/").split("/").filter(Boolean);
  return parts.length > 2 ? parts.slice(-2).join("/") : parts.join("/");
}

/**
 * Extract a file path target from a command string (first path-like argument).
 */
function extractTarget(cmd) {
  if (!cmd) return "";
  // Match file paths (containing / or \ and a file extension, or known directories)
  const m = cmd.match(
    /(?:['"])?([\w.\-/\\]+\.(?:ps1|mjs|js|ts|go|json|yaml|yml|md|log|txt|toml|sh))(?:['"])?/i,
  );
  if (m) return shortPath(m[1]);
  // Match directory paths
  const d = cmd.match(
    /(?:(?:Get-Content|cat|head|tail|type|Select-String)\s+(?:-Path\s+)?['"]?)([\w.\-/\\]+)/i,
  );
  if (d) return shortPath(d[1]);
  return "";
}

function normalizeToolName(value) {
  return String(value || "")
    .trim()
    .toLowerCase()
    .replace(/\s+/g, "_");
}

function getCopilotToolInfo(event) {
  const data = event?.data || {};
  const toolName =
    data.toolName ||
    data.name ||
    data.tool?.name ||
    event?.toolName ||
    event?.tool ||
    "";
  const input =
    data.input ||
    data.args ||
    data.parameters ||
    data.toolInput ||
    data.payload ||
    null;
  const output = data.output || data.result || data.toolOutput || null;
  const status = data.status || event?.status || "";
  return { toolName: String(toolName || ""), input, output, status };
}

function extractCopilotCommand(input) {
  if (!input) return null;
  if (typeof input === "string") return input;
  const candidates = [
    "command",
    "cmd",
    "shell",
    "script",
    "execute",
    "args",
    "run",
  ];
  for (const key of candidates) {
    const value = input[key];
    if (typeof value === "string" && value.trim()) {
      return value.trim();
    }
  }
  return null;
}

function extractCopilotPath(input) {
  if (!input) return null;
  if (typeof input === "string") return input;
  const candidates = [
    "path",
    "file",
    "filename",
    "filepath",
    "filePath",
    "fullPath",
    "target",
  ];
  for (const key of candidates) {
    const value = input[key];
    if (typeof value === "string" && value.trim()) {
      return value.trim();
    }
  }
  return null;
}

function isCopilotReadTool(name) {
  const tool = normalizeToolName(name);
  return /read|open|view|get_file|read_file/.test(tool);
}

function isCopilotWriteTool(name) {
  const tool = normalizeToolName(name);
  return /write|edit|apply|patch|create|update|save/.test(tool);
}

function isCopilotSearchTool(name) {
  const tool = normalizeToolName(name);
  return /search|grep|rg|find|query/.test(tool);
}

function summarizeCopilotTool(toolName, input) {
  const tool = normalizeToolName(toolName);
  const command = extractCopilotCommand(input);
  const target = extractCopilotPath(input);
  if (!toolName) return "running tool";
  if (/command|shell|execute|run/.test(tool)) {
    return command ? `running ${shortSnippet(command, 80)}` : "running command";
  }
  if (isCopilotReadTool(tool)) {
    return target ? `reading ${shortPath(target)}` : "reading file";
  }
  if (isCopilotWriteTool(tool)) {
    return target ? `updating ${shortPath(target)}` : "updating files";
  }
  if (isCopilotSearchTool(tool)) {
    return target ? `searching ${shortPath(target)}` : "searching";
  }
  if (/mcp/.test(tool)) return `MCP tool: ${toolName}`;
  return `tool: ${toolName}`;
}

/**
 * Convert a raw agent event into a concise human-readable action description.
 * Shows which files are being read/written, line counts for changes, and
 * concise command summaries with targets.
 */
function summarizeAction(event) {
  if (!event) return null;

  switch (event.type) {
    case "item.started": {
      const item = event.item;
      switch (item.type) {
        case "command_execution": {
          const desc = summarizeCommand(item.command);
          const target = extractTarget(item.command);
          return {
            icon: "⚡",
            text: target ? `${desc} → ${target}` : desc,
            phase: "running",
          };
        }
        case "mcp_tool_call":
          return {
            icon: "🔌",
            text: `MCP: ${item.server}/${item.tool}`,
            phase: "running",
          };
        case "reasoning":
          return item.text
            ? { icon: "💭", text: item.text.slice(0, 200), phase: "thinking" }
            : null;
        case "web_search":
          return {
            icon: "🔍",
            text: `Searching: ${item.query?.slice(0, 80)}`,
            phase: "searching",
          };
        case "todo_list":
          return item.items?.length
            ? {
                icon: "📋",
                text: `Planning ${item.items.length} steps`,
                phase: "planning",
              }
            : null;
        default:
          return null;
      }
    }

    case "item.completed": {
      const item = event.item;
      switch (item.type) {
        case "command_execution": {
          const ok = item.exit_code === 0;
          const desc = summarizeCommand(item.command);
          const target = extractTarget(item.command);
          const label = target ? `${desc} → ${target}` : desc;
          return {
            icon: ok ? "✅" : "❌",
            text: label + (ok ? "" : ` (exit ${item.exit_code})`),
            phase: "done",
          };
        }
        case "file_change": {
          if (item.changes?.length) {
            const fileDescs = item.changes.map((c) => {
              const name = shortPath(c.path);
              const kind =
                c.kind === "add" ? "➕" : c.kind === "delete" ? "🗑️" : "✏️";
              // Show line counts if available
              const adds = c.additions ?? c.lines_added ?? 0;
              const dels = c.deletions ?? c.lines_deleted ?? 0;
              const stats = adds || dels ? ` (+${adds} -${dels})` : "";
              return `${kind} ${name}${stats}`;
            });
            return {
              icon: "📁",
              text: fileDescs.join(", "),
              phase: "done",
              detail: "file_change",
              files: item.changes.map((c) => ({
                path: c.path,
                kind: c.kind,
                adds: c.additions ?? c.lines_added ?? 0,
                dels: c.deletions ?? c.lines_deleted ?? 0,
              })),
            };
          }
          return null;
        }
        case "mcp_tool_call": {
          const ok = item.status === "completed";
          return {
            icon: ok ? "✅" : "❌",
            text: `MCP ${item.server}/${item.tool}: ${ok ? "done" : "failed"}`,
            phase: "done",
          };
        }
        case "agent_message":
          return null; // final response handled separately
        default:
          return null;
      }
    }

    case "assistant.reasoning":
    case "assistant.reasoning_delta": {
      const text = event.data?.content || event.data?.deltaContent || "";
      return text
        ? { icon: "💭", text: text.slice(0, 200), phase: "thinking" }
        : null;
    }

    case "tool.execution_start": {
      const { toolName, input } = getCopilotToolInfo(event);
      return {
        icon: "🛠️",
        text: summarizeCopilotTool(toolName, input),
        phase: "running",
      };
    }

    case "tool.execution_complete": {
      const { toolName, input, status } = getCopilotToolInfo(event);
      const ok =
        !status ||
        ["ok", "success", "completed", "done"].includes(
          String(status).toLowerCase(),
        );
      return {
        icon: ok ? "✅" : "❌",
        text: summarizeCopilotTool(toolName, input) + (ok ? "" : " (failed)"),
        phase: "done",
      };
    }

    case "session.error":
      return {
        icon: "❌",
        text: `Failed: ${event.data?.message || "unknown"}`,
        phase: "error",
      };

    case "item.updated": {
      const item = event.item;
      if (item.type === "reasoning" && item.text) {
        return { icon: "💭", text: item.text.slice(0, 200), phase: "thinking" };
      }
      if (item.type === "todo_list" && item.items) {
        const done = item.items.filter((t) => t.completed).length;
        return {
          icon: "📋",
          text: `Progress: ${done}/${item.items.length} steps`,
          phase: "planning",
        };
      }
      return null;
    }

    case "turn.failed":
      return {
        icon: "❌",
        text: `Failed: ${event.error?.message || "unknown"}`,
        phase: "error",
      };
    default:
      return null;
  }
}

/**
 * Convert raw command strings into concise human-readable descriptions.
 */
function summarizeCommand(cmd) {
  if (!cmd) return "(unknown command)";
  const c = cmd.trim();
  const cmdLower = c.toLowerCase();
  const clean = c.replace(/\s+/g, " ").trim();

  // Git commands
  if (/^git\s+diff/i.test(c)) return "checking git diff";
  if (/^git\s+log/i.test(c)) return "reading git log";
  if (/^git\s+branch/i.test(c)) return "listing git branches";
  if (/^git\s+status/i.test(c)) return "checking git status";
  if (/^git\s+add/i.test(c)) return "staging files";
  if (/^git\s+commit/i.test(c)) return "committing changes";
  if (/^git\s+push/i.test(c)) return "pushing to remote";
  if (/^git\s+pull/i.test(c)) return "pulling from remote";
  if (/^git\s+checkout/i.test(c)) return "switching branch";
  if (/^git\s+merge/i.test(c)) return "merging branches";
  if (/^git\s+stash/i.test(c)) return "stashing changes";

  // PowerShell / search patterns
  if (/pwsh.*-file/i.test(c)) {
    const target = extractTarget(c);
    return target
      ? `running PowerShell file ${target}`
      : "running PowerShell file";
  }
  if (/pwsh.*Get-Content/i.test(c)) {
    const target = extractTarget(c);
    return target ? `reading ${target}` : "reading file contents";
  }
  if (/pwsh.*Select-String/i.test(c)) {
    const target = extractTarget(c);
    const patternMatch = c.match(/-Pattern\s+["']?([^"']+)["']?/i);
    const pattern = patternMatch ? patternMatch[1].slice(0, 40) : null;
    if (target && pattern) return `searching "${pattern}" in ${target}`;
    return target ? `searching in ${target}` : "searching in files";
  }
  if (/pwsh.*Get-ChildItem.*Select-String/i.test(c))
    return "searching across files";
  if (/pwsh/i.test(c) || /powershell/i.test(c))
    return describePowerShell(clean);

  // Node/npm/pnpm
  if (/^node\s+-[ec]/i.test(c))
    return `running Node.js script: ${shortSnippet(clean, 60)}`;
  if (/^npm\s+/i.test(c)) return `running npm: ${shortSnippet(clean, 60)}`;
  if (/^pnpm\s+/i.test(c)) return `running pnpm: ${shortSnippet(clean, 60)}`;

  // Go
  if (/^go\s+test/i.test(c)) {
    const pkgs = extractGoPackages(c);
    return pkgs ? `running Go tests: ${pkgs}` : "running Go tests";
  }
  if (/^go\s+build/i.test(c)) {
    const pkgs = extractGoPackages(c);
    return pkgs ? `building Go: ${pkgs}` : "building Go project";
  }
  if (/^go\s+vet/i.test(c)) {
    const pkgs = extractGoPackages(c);
    return pkgs ? `vetting Go: ${pkgs}` : "vetting Go code";
  }
  if (/^go\s+/i.test(c)) return `running Go: ${shortSnippet(clean, 60)}`;

  // Make
  if (/^make\s+/i.test(c))
    return `running make ${c.split(/\s+/)[1] || ""}`.trim();

  // gh CLI
  if (/^gh\s+pr/i.test(c)) return "managing GitHub PR";
  if (/^gh\s+issue/i.test(c)) return "managing GitHub issue";
  if (/^gh\s+/i.test(c)) return `running gh: ${shortSnippet(clean, 60)}`;

  // cat/head/tail/grep/find/ls
  if (/^(cat|head|tail|type)\s+/i.test(c)) {
    const target = extractTarget(c);
    return target ? `reading ${target}` : "reading file";
  }
  if (/^(grep|findstr|rg)\s+/i.test(c)) {
    const target = extractTarget(c);
    const pat = c.match(/(['"])(.*?)\1/);
    const pattern = pat ? pat[2].slice(0, 40) : null;
    if (target && pattern) return `searching "${pattern}" in ${target}`;
    return target ? `searching in ${target}` : "searching in files";
  }
  if (/^(find|fd)\s+/i.test(c)) return "finding files";
  if (/^(ls|dir|Get-ChildItem)\s*/i.test(c)) return "listing directory";

  // Docker
  if (/^docker\s+/i.test(c))
    return `running docker: ${shortSnippet(clean, 60)}`;

  // Fallback: first word + truncated
  const firstWord = c.split(/\s+/)[0];
  if (firstWord.length < 20)
    return `running ${firstWord}: ${shortSnippet(clean, 60)}`;
  return shortSnippet(clean, 80);
}

function shortSnippet(text, maxLen = 80) {
  if (!text) return "";
  if (text.length <= maxLen) return text;
  return text.slice(0, maxLen - 1) + "…";
}

function describePowerShell(command) {
  const cmd = command;
  const cmdMatch = cmd.match(/-Command\s+(.+)/i);
  const fileMatch = cmd.match(/-File\s+([^\s]+)/i);
  if (fileMatch) {
    const target = shortPath(fileMatch[1]);
    return `running PowerShell file ${target}`;
  }
  if (cmdMatch) {
    const inner = cmdMatch[1].replace(/^['"]|['"]$/g, "");
    const target = extractTarget(inner);
    const snippet = shortSnippet(inner, 70);
    return target
      ? `running PowerShell: ${snippet} → ${target}`
      : `running PowerShell: ${snippet}`;
  }
  return "running PowerShell command";
}

function extractGoPackages(command) {
  const parts = command.split(/\s+/).filter(Boolean);
  const pkgs = parts.filter((p) => p.startsWith("./") || p.includes("/"));
  if (!pkgs.length) return "";
  const unique = [...new Set(pkgs)];
  return unique.slice(0, 3).join(" ") + (unique.length > 3 ? " …" : "");
}

function splitMessage(text, maxLen) {
  if (!text) return ["(empty response)"];
  if (text.length <= maxLen) return [text];

  const chunks = [];
  let remaining = text;
  while (remaining.length > 0) {
    if (remaining.length <= maxLen) {
      chunks.push(remaining);
      break;
    }
    // Try to split at newline
    let splitIdx = remaining.lastIndexOf("\n", maxLen);
    if (splitIdx < maxLen * 0.3) {
      splitIdx = maxLen; // no good newline, hard split
    }
    chunks.push(remaining.slice(0, splitIdx));
    remaining = remaining.slice(splitIdx);
  }
  return chunks;
}

// ── Polling ──────────────────────────────────────────────────────────────────

async function pollUpdates() {
  if (!telegramToken) return [];

  const url = `https://api.telegram.org/bot${telegramToken}/getUpdates`;
  const params = new URLSearchParams({
    offset: String(lastUpdateId + 1),
    timeout: String(POLL_TIMEOUT_S),
    allowed_updates: JSON.stringify(["message"]),
  });

  pollAbort = new AbortController();
  try {
    const res = await fetch(`${url}?${params}`, {
      signal: pollAbort.signal,
      // No explicit timeout — the Telegram API long-poll handles timing
    });
    if (!res.ok) {
      const body = await res.text();
      console.warn(`[telegram-bot] getUpdates failed: ${res.status} ${body}`);
      if (res.status === 409) {
        polling = false;
        await releaseTelegramPollLock();
      }
      return [];
    }
    const data = await res.json();
    return data.ok ? data.result || [] : [];
  } catch (err) {
    if (err.name === "AbortError") return [];
    console.warn(`[telegram-bot] poll error: ${err.message}`);
    return [];
  } finally {
    pollAbort = null;
  }
}

// ── Update Handling ──────────────────────────────────────────────────────────

async function handleUpdate(update) {
  if (!update.message) return;

  const msg = update.message;
  const chatId = String(msg.chat?.id);
  const text = (msg.text || "").trim();

  const presencePayload = text ? parsePresenceMessage(text) : null;
  if (presencePayload) {
    if (!presenceChatId || chatId !== String(presenceChatId)) {
      console.warn(
        `[telegram-bot] ignored presence from chat ${chatId} (expected ${presenceChatId || "unset"})`,
      );
      return;
    }
    await ensurePresenceReady();
    const receivedAt = Number.isFinite(msg.date)
      ? new Date(msg.date * 1000).toISOString()
      : new Date().toISOString();
    await notePresence(presencePayload, {
      source: "telegram",
      receivedAt,
    });
    return;
  }

  // Security: only accept from configured chat
  if (telegramChatId && chatId !== String(telegramChatId)) {
    console.warn(
      `[telegram-bot] rejected message from chat ${chatId} (expected ${telegramChatId})`,
    );
    return;
  }

  if (!text) return;

  console.log(
    `[telegram-bot] received: "${text.slice(0, 80)}${text.length > 80 ? "..." : ""}" from chat ${chatId}`,
  );

  // Route: slash command or free-text
  if (text.startsWith("/")) {
    const cmd = text.split(/\s+/)[0].toLowerCase().replace(/@\w+/, "");
    if (FAST_COMMANDS.has(cmd)) {
      enqueueFastCommand(() => handleCommand(text, chatId));
      return;
    }
    enqueueCommand(() => handleCommand(text, chatId));
    return;
  }

  // Free-text agent task runs in a separate queue so polling isn't blocked.
  // If agent is already busy, handle immediately so follow-ups can be queued.
  if (isPrimaryBusy()) {
    void handleFreeText(text, chatId);
    return;
  }
  enqueueAgentTask(() => handleFreeText(text, chatId));
}

// ── Command Router ────────────────────────────────────────────────────────────

// ── Task Pause / Resume / Repos ──────────────────────────────────────────────

/**
 * /pausetasks — Pause the task executor (running agents continue, no new dispatch)
 */
async function cmdPauseTasks(chatId) {
  const executor = _getInternalExecutor?.();
  if (!executor) {
    return sendDirect(chatId, "⚠️ Internal executor not enabled — nothing to pause.");
  }
  if (executor.isPaused()) {
    const info = executor.getPauseInfo();
    const dur = info.pauseDuration;
    return sendDirect(chatId, `⏸ Already paused (${dur >= 60 ? Math.round(dur / 60) + "m" : dur + "s"} ago).\nUse /resumetasks to resume.`);
  }
  executor.pause();
  const status = executor.getStatus();
  const lines = [`⏸ *Task executor paused*`];
  if (status.activeSlots > 0) {
    lines.push(`\n${status.activeSlots} running task(s) will continue to completion.`);
    lines.push(`No new tasks will be dispatched until /resumetasks.`);
  } else {
    lines.push(`No tasks running. Use /resumetasks when ready.`);
  }
  return sendDirect(chatId, lines.join("\n"), { parse_mode: "Markdown" });
}

/**
 * /resumetasks — Resume the task executor
 */
async function cmdResumeTasks(chatId) {
  const executor = _getInternalExecutor?.();
  if (!executor) {
    return sendDirect(chatId, "⚠️ Internal executor not enabled — nothing to resume.");
  }
  if (!executor.isPaused()) {
    return sendDirect(chatId, "▶️ Executor is already running — not paused.");
  }
  const info = executor.getPauseInfo();
  const dur = info.pauseDuration;
  executor.resume();
  const durStr = dur >= 60 ? Math.round(dur / 60) + "m" : dur + "s";
  return sendDirect(chatId, `▶️ *Task executor resumed* (was paused for ${durStr}).\nWill pick up tasks on next poll cycle.`, { parse_mode: "Markdown" });
}

/**
 * /repos — View configured repositories and their status
 */
async function cmdRepos(chatId, _text) {
  try {
    const config = (await import("./config.mjs")).default;
    const repos = config.repositories || [];
    const selected = config.selectedRepository || config.repoSlug || "(default)";
    
    if (repos.length === 0) {
      return sendDirect(chatId, [
        "📁 *Repositories*",
        "",
        `Active: \`${config.repoSlug || config.repoRoot || "current directory"}\``,
        "",
        "_Single-repo mode. Add repositories in codex-monitor.config.json:_",
        "\`\`\`json",
        JSON.stringify({
          repositories: [
            { name: "backend", path: "./", slug: "org/backend", primary: true },
            { name: "frontend", path: "../frontend", slug: "org/frontend" }
          ]
        }, null, 2),
        "\`\`\`",
      ].join("\n"), { parse_mode: "Markdown" });
    }
    
    const lines = ["📁 *Repositories*", ""];
    for (const repo of repos) {
      const isCurrent = repo.name === selected || repo.slug === selected || repo.primary;
      const icon = isCurrent ? "🟢" : "⚪";
      const primary = repo.primary ? " _(primary)_" : "";
      lines.push(`${icon} \`${repo.name}\` — ${repo.slug || repo.path || "?"}${primary}`);
    }
    lines.push("");
    lines.push(`Selected: \`${selected}\``);
    lines.push("Switch: \`/repos set <name>\`");
    
    return sendDirect(chatId, lines.join("\n"), { parse_mode: "Markdown" });
  } catch (err) {
    return sendDirect(chatId, `❌ Failed to read repo config: ${err.message}`);
  }
}

/**
 * /maxparallel — View or set max parallel task slots
 */
async function cmdMaxParallel(chatId, text) {
  const executor = _getInternalExecutor?.();
  if (!executor) {
    return sendDirect(chatId, "⚠️ Internal executor not enabled.");
  }
  const arg = (text || "").replace("/maxparallel", "").trim();
  if (arg) {
    const n = parseInt(arg, 10);
    if (isNaN(n) || n < 0 || n > 20) {
      return sendDirect(chatId, "⚠️ Provide a number between 0 and 20.");
    }
    const old = executor.maxParallel;
    executor.maxParallel = n;
    if (n === 0) {
      executor.pause();
      return sendDirect(chatId, `⏸ Max parallel set to 0 — executor paused. Use /maxparallel <n> to resume.`);
    }
    if (executor.isPaused() && n > 0) {
      executor.resume();
    }
    return sendDirect(chatId, `✅ Max parallel: ${old} → ${n}`);
  }
  const status = executor.getStatus();
  return sendDirect(chatId, `📊 Max parallel: ${status.maxParallel} (active: ${status.activeSlots})`);
}

const COMMANDS = {
  "/help": { handler: cmdHelp, desc: "Show available commands" },
  "/ask": { handler: cmdAsk, desc: "Send prompt to agent: /ask <prompt>" },
  "/status": { handler: cmdStatus, desc: "Detailed orchestrator status" },
  "/tasks": {
    handler: cmdTasks,
    desc: "Active tasks, workspace metrics & retries",
  },
  "/logs": { handler: cmdLogs, desc: "Recent monitor logs" },
  "/agentlogs": { handler: cmdAgentLogs, desc: "Agent output for branch: /agentlogs <branch>" },
  "/log": { handler: cmdLogs, desc: "Alias for /logs" },
  "/branches": { handler: cmdBranches, desc: "Recent git branches" },
  "/diff": { handler: cmdDiff, desc: "Git diff summary (staged)" },
  "/restart": { handler: cmdRestart, desc: "Restart orchestrator process" },
  "/retry": {
    handler: cmdRetry,
    desc: "Start fresh session for stuck task: /retry [reason]",
  },
  "/plan": {
    handler: cmdPlan,
    desc: "Trigger task planner: /plan [count] (default 5)",
  },
  "/cleanup": {
    handler: cmdCleanupMerged,
    desc: "Reconcile VK tasks with merged PRs/branches",
  },
  "/reconcile": {
    handler: cmdCleanupMerged,
    desc: "Alias for /cleanup",
  },
  "/history": {
    handler: cmdHistory,
    desc: "Primary agent session history",
  },
  "/clear": {
    handler: cmdClear,
    desc: "Clear primary agent session context",
  },
  "/reset_thread": {
    handler: cmdClear,
    desc: "Alias for /clear (reset thread)",
  },
  "/git": { handler: cmdGit, desc: "Run a git command: /git log --oneline -5" },
  "/shell": { handler: cmdShell, desc: "Run a shell command: /shell ls -la" },
  "/background": {
    handler: cmdBackground,
    desc: "Run a task in background or background the active agent",
  },
  "/region": {
    handler: cmdRegion,
    desc: "View/switch Codex region: /region [us|sweden|auto]",
  },
  "/health": {
    handler: cmdHealth,
    desc: "Executor health status & model routing",
  },
  "/anomalies": {
    handler: cmdAnomalies,
    desc: "Agent anomaly detector status & active concerns",
  },
  "/model": {
    handler: cmdModel,
    desc: "Override executor for next task: /model gpt-5.2-codex",
  },
  "/sdk": {
    handler: cmdSdk,
    desc: "View/switch agent pool SDK: /sdk [codex|copilot|claude]",
  },
  "/kanban": {
    handler: cmdKanban,
    desc: "View/switch kanban backend: /kanban [vk|github|jira]",
  },
  "/threads": {
    handler: cmdThreads,
    desc: "View active agent threads: /threads [clear]",
  },
  "/worktrees": {
    handler: cmdWorktrees,
    desc: "View/manage worktrees: /worktrees [prune|release <taskKey>]",
  },
  "/executor": {
    handler: cmdExecutor,
    desc: "View/manage executor mode: /executor [status|mode <vk|internal|hybrid>|slots]",
  },
  "/shared_workspaces": {
    handler: cmdSharedWorkspaces,
    desc: "List shared cloud workspace availability",
  },
  "/claim": {
    handler: cmdSharedWorkspaceClaim,
    desc: "Claim a shared workspace: /claim <id> [--owner <id>] [--ttl <minutes>]",
  },
  "/release": {
    handler: cmdSharedWorkspaceRelease,
    desc: "Release a shared workspace: /release <id> [--owner <id>] [--force]",
  },
  "/agent": {
    handler: cmdAgent,
    desc: "Route a task to a workspace: /agent --workspace <id> <task>",
  },
  "/stop": {
    handler: cmdStop,
    desc: "Stop the running agent and wait for new instructions",
  },
  "/steer": {
    handler: cmdSteer,
    desc: "Steer a running agent: /steer focus on X",
  },
  "/context": {
    handler: cmdSteer,
    desc: "Alias for /steer — update in-flight agent context",
  },
  "/presence": {
    handler: cmdPresence,
    desc: "Show active codex-monitor instances",
  },
  "/instances": {
    handler: cmdPresence,
    desc: "Alias for /presence",
  },
  "/coordinator": {
    handler: cmdCoordinator,
    desc: "Show current coordinator selection",
  },
  "/pausetasks": {
    handler: cmdPauseTasks,
    desc: "Pause task dispatch (running tasks continue)",
  },
  "/resumetasks": {
    handler: cmdResumeTasks,
    desc: "Resume task dispatch after pause",
  },
  "/pause": {
    handler: cmdPauseTasks,
    desc: "Alias for /pausetasks",
  },
  "/resume": {
    handler: cmdResumeTasks,
    desc: "Alias for /resumetasks",
  },
  "/repos": {
    handler: cmdRepos,
    desc: "View configured repositories",
  },
  "/maxparallel": {
    handler: cmdMaxParallel,
    desc: "View/set max parallel slots: /maxparallel [n]",
  },
};

/**
 * Delete all existing bot commands from every scope to clear stale/old entries.
 * Telegram stores commands per-scope, so we must clear each one explicitly.
 */
async function clearAllBotCommands() {
  const scopes = [
    { scope: { type: "default" } },
    { scope: { type: "all_private_chats" } },
    { scope: { type: "all_group_chats" } },
    { scope: { type: "all_chat_administrators" } },
  ];

  for (const body of scopes) {
    try {
      await fetch(
        `https://api.telegram.org/bot${telegramToken}/deleteMyCommands`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(body),
        },
      );
    } catch {
      /* best effort — scope may not have had commands */
    }
  }
}

/**
 * Sync the COMMANDS object to Telegram's bot command menu via setMyCommands.
 * First clears ALL existing commands from every scope to remove stale entries
 * (e.g. leftover commands from a previous project or bot configuration).
 * Then sets the current command list.
 */
async function registerBotCommands() {
  // Step 1: Clear all old commands from every scope
  await clearAllBotCommands();

  // Step 2: Build and register current commands
  const seen = new Set();
  const commands = [];
  for (const [cmd, entry] of Object.entries(COMMANDS)) {
    const command = cmd.replace(/^\//, ""); // strip leading /
    if (seen.has(command)) continue; // skip duplicate keys (e.g. /agent appears twice)
    // Telegram only allows lowercase letters, digits, underscores (1-32 chars)
    if (!/^[a-z0-9_]{1,32}$/.test(command)) {
      console.warn(`[telegram-bot] skipping invalid command name: /${command}`);
      continue;
    }
    seen.add(command);
    // Telegram limits description to 256 chars
    const description = (entry.desc || command).slice(0, 256);
    commands.push({ command, description });
  }

  try {
    const res = await fetch(
      `https://api.telegram.org/bot${telegramToken}/setMyCommands`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ commands }),
      },
    );
    const data = await res.json();
    if (data.ok) {
      console.log(
        `[telegram-bot] registered ${commands.length} commands with Telegram`,
      );
    } else {
      console.warn(
        `[telegram-bot] setMyCommands failed: ${data.description || JSON.stringify(data)}`,
      );
    }
  } catch (err) {
    console.warn(`[telegram-bot] setMyCommands error: ${err.message}`);
  }
}

const FAST_COMMANDS = new Set([
  "/status",
  "/tasks",
  "/sdk",
  "/kanban",
  "/threads",
  "/worktrees",
  "/executor",
  "/pausetasks",
  "/resumetasks",
  "/pause",
  "/resume",
  "/maxparallel",
  "/repos",
]);

async function handleCommand(text, chatId) {
  const parts = text.split(/\s+/);
  const cmd = parts[0].toLowerCase().replace(/@\w+/, ""); // strip @botname
  const cmdArgs = parts.slice(1).join(" ");

  const entry = COMMANDS[cmd] || COMMANDS[cmd.replace(/-/g, "_")];
  if (entry) {
    try {
      await entry.handler(chatId, cmdArgs);
    } catch (err) {
      await sendReply(chatId, `❌ Command error: ${err.message}`);
    }
  } else {
    await sendReply(
      chatId,
      `Unknown command: ${cmd}\nType /help for available commands.`,
    );
  }
}

// ── Built-in Command Handlers ────────────────────────────────────────────────

function splitArgs(input) {
  if (!input) return [];
  const tokens = [];
  const re = /"([^"]*)"|'([^']*)'|(\S+)/g;
  let match;
  while ((match = re.exec(input)) !== null) {
    tokens.push(match[1] ?? match[2] ?? match[3]);
  }
  return tokens;
}

function parseSharedWorkspaceArgs(args) {
  const tokens = splitArgs(args);
  const parsed = {
    workspaceId: null,
    owner: null,
    ttlMinutes: null,
    note: "",
    reason: "",
    force: false,
  };
  for (let i = 0; i < tokens.length; i++) {
    const token = tokens[i];
    if (token === "--owner") {
      parsed.owner = tokens[i + 1];
      i++;
      continue;
    }
    if (token === "--ttl") {
      parsed.ttlMinutes = Number(tokens[i + 1]);
      i++;
      continue;
    }
    if (token === "--note") {
      parsed.note = tokens.slice(i + 1).join(" ");
      break;
    }
    if (token === "--reason") {
      parsed.reason = tokens.slice(i + 1).join(" ");
      break;
    }
    if (token === "--force") {
      parsed.force = true;
      continue;
    }
    if (!parsed.workspaceId) {
      parsed.workspaceId = token;
    }
  }
  return parsed;
}

function parseAgentArgs(args) {
  const tokens = splitArgs(args);
  let workspaceId = null;
  let role = null;
  let model = null;
  let queue = false;
  let newSession = false;
  let dryRun = false;
  let taskTokens = [];
  const remaining = [];

  for (let i = 0; i < tokens.length; i++) {
    const token = tokens[i];
    if (token === "--workspace" || token === "-w") {
      workspaceId = tokens[i + 1] || null;
      i++;
      continue;
    }
    if (token.startsWith("--workspace=")) {
      workspaceId = token.slice("--workspace=".length) || null;
      continue;
    }
    if (token === "--role" || token === "-r") {
      role = tokens[i + 1] || null;
      i++;
      continue;
    }
    if (token.startsWith("--role=")) {
      role = token.slice("--role=".length) || null;
      continue;
    }
    if (token === "--model" || token === "-m") {
      model = tokens[i + 1] || null;
      i++;
      continue;
    }
    if (token.startsWith("--model=")) {
      model = token.slice("--model=".length) || null;
      continue;
    }
    if (token === "--queue") {
      queue = true;
      continue;
    }
    if (token === "--new-session") {
      newSession = true;
      continue;
    }
    if (token === "--dry-run") {
      dryRun = true;
      continue;
    }
    if (token === "--task" || token === "-t") {
      taskTokens = tokens.slice(i + 1);
      break;
    }
    if (token.startsWith("--task=")) {
      taskTokens = [token.slice("--task=".length)];
      break;
    }
    remaining.push(token);
  }

  if (taskTokens.length === 0) {
    taskTokens = remaining;
  }

  return {
    workspaceId: workspaceId ? workspaceId.trim() : null,
    role: role ? role.trim() : null,
    model: model ? model.trim() : null,
    queue,
    newSession,
    dryRun,
    message: taskTokens.join(" ").trim(),
  };
}

function isPathLike(value) {
  return /[\\/]|^[A-Za-z]:/.test(value);
}

function findRepoPath(basePath) {
  if (!basePath) return null;
  const resolved = resolve(basePath);
  if (existsSync(resolve(resolved, "go.mod"))) {
    return resolved;
  }
  const nested = resolve(resolved, "virtengine");
  if (existsSync(resolve(nested, "go.mod"))) {
    return nested;
  }
  return null;
}

async function listWorkspaceIds(worktreesRoot) {
  const entries = await readdir(worktreesRoot, { withFileTypes: true }).catch(
    () => [],
  );
  const ids = [];
  for (const entry of entries) {
    if (!entry.isDirectory()) continue;
    const repoPath = findRepoPath(resolve(worktreesRoot, entry.name));
    if (repoPath) ids.push(entry.name);
  }
  return ids;
}

async function resolveWorkspaceRepo(workspaceId) {
  const trimmed = (workspaceId || "").trim();
  if (!trimmed) {
    return {
      repoPath: repoRoot,
      label: "primary coordinator",
      isPrimary: true,
    };
  }

  const lower = trimmed.toLowerCase();
  if (["primary", "coordinator", "default"].includes(lower)) {
    return {
      repoPath: repoRoot,
      label: "primary coordinator",
      isPrimary: true,
    };
  }

  const worktreesRoot = resolve(repoRoot, "..", "..");
  const candidates = [];

  if (isPathLike(trimmed)) {
    candidates.push(trimmed);
    candidates.push(resolve(worktreesRoot, trimmed));
  } else {
    candidates.push(resolve(worktreesRoot, trimmed));
  }

  for (const candidate of candidates) {
    const repoPath = findRepoPath(candidate);
    if (repoPath) {
      return {
        repoPath,
        label: trimmed,
        isPrimary: repoPath === repoRoot,
      };
    }
  }

  const suggestions = await listWorkspaceIds(worktreesRoot);
  return {
    error: `Workspace "${trimmed}" not found.`,
    suggestions,
    worktreesRoot,
  };
}

async function loadWorkspaceStatusData(workspacePath) {
  try {
    const workspaceStatusPath = resolve(
      workspacePath,
      ".cache",
      "ve-orchestrator-status.json",
    );
    const raw = await readFile(workspaceStatusPath, "utf8").catch(() => null);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

async function cmdHelp(chatId) {
  const lines = ["🤖 VirtEngine Primary Agent Commands:\n"];
  for (const [cmd, { desc }] of Object.entries(COMMANDS)) {
    lines.push(`${cmd} — ${desc}`);
  }
  lines.push(
    "",
    "Any other text → sent to the primary agent (full repo + MCP access)",
  );
  await sendReply(chatId, lines.join("\n"));
}

async function cmdAsk(chatId, args) {
  const prompt = String(args || "").trim();
  if (!prompt) {
    await sendReply(chatId, "Usage: /ask <prompt>");
    return;
  }
  enqueueAgentTask(() => handleFreeText(prompt, chatId));
}

async function cmdStatus(chatId) {
  await sendReply(chatId, "⏳ Reading orchestrator status...");

  let statusText = "Status unavailable";

  // Try the formatted summary first
  if (_readStatusSummary) {
    try {
      const summary = await _readStatusSummary();
      if (summary?.text) {
        await sendReply(chatId, summary.text, {
          parseMode: summary.parseMode || undefined,
        });
        return;
      }
    } catch {
      /* fallback to raw */
    }
  }

  // Fallback: read raw status file
  try {
    const raw = await readFile(statusPath, "utf8");
    const data = JSON.parse(raw);

    const counts = data.counts || {};
    const sm = data.success_metrics || {};
    const backlog = data.backlog_remaining ?? "?";
    const submitted = Array.isArray(data.submitted_tasks)
      ? data.submitted_tasks.length
      : 0;
    const completed = Array.isArray(data.completed_tasks)
      ? data.completed_tasks.length
      : 0;
    const errors = Array.isArray(data.error_tasks) ? data.error_tasks : [];
    const reviews = Array.isArray(data.review_tasks) ? data.review_tasks : [];
    const manualReviews = Array.isArray(data.manual_review_tasks)
      ? data.manual_review_tasks
      : [];

    const lines = [
      "📊 VirtEngine Orchestrator Status",
      "",
      `Running: ${counts.running ?? 0}`,
      `Review: ${counts.review ?? 0}`,
      `Error: ${counts.error ?? 0}`,
      `Manual Review: ${counts.manual_review ?? 0}`,
      `Backlog: ${backlog}`,
      "",
      `Submitted: ${submitted} | Completed: ${completed}`,
      `First-shot: ${sm.first_shot_rate ?? 0}% (${sm.first_shot_success ?? 0}/${(sm.first_shot_success ?? 0) + (sm.needed_fix ?? 0) + (sm.failed ?? 0)})`,
      `Needed fix: ${sm.needed_fix ?? 0} | Failed: ${sm.failed ?? 0}`,
    ];

    if (errors.length > 0) {
      lines.push(
        "",
        "⚠️ Error tasks:",
        ...errors.slice(0, 5).map((t) => `  - ${t}`),
      );
    }
    if (manualReviews.length > 0) {
      lines.push(
        "",
        "👀 Manual review:",
        ...manualReviews.slice(0, 5).map((t) => `  - ${t}`),
      );
    }

    statusText = lines.join("\n");
  } catch (err) {
    statusText = `Status file error: ${err.message}`;
  }

  await sendReply(chatId, statusText);
}

async function cmdAnomalies(chatId) {
  if (!_getAnomalyReport) {
    await sendReply(chatId, "Anomaly detector not initialized.");
    return;
  }
  try {
    const report = _getAnomalyReport();
    await sendReply(chatId, report, { parseMode: "HTML" });
  } catch (err) {
    await sendReply(chatId, `Error getting anomaly report: ${err.message}`);
  }
}

async function cmdTasks(chatId) {
  try {
    // ── Prefer live executor slots over stale status file ──
    const executor = _getInternalExecutor?.();
    const executorStatus = executor?.getStatus?.();

    if (executorStatus) {
      const lines = [];

      // Show pause state prominently at top
      if (executorStatus.paused) {
        const dur = executorStatus.pauseDuration || 0;
        const durStr = dur >= 3600 ? Math.round(dur / 3600) + "h" : dur >= 60 ? Math.round(dur / 60) + "m" : dur + "s";
        lines.push(`⏸ PAUSED (for ${durStr}) — /resumetasks to resume`);
        lines.push("");
      }

      if (executorStatus.slots.length > 0) {
        lines.push(`📋 Active Agents (${executorStatus.activeSlots}/${executorStatus.maxParallel} slots)\n`);

      for (const slot of executorStatus.slots) {
        const emoji = slot.status === "running" ? "🟢" : slot.status === "error" ? "❌" : "🔵";
        const runMin = Math.round(slot.runningFor / 60);
        const runStr = runMin >= 60 ? `${Math.floor(runMin / 60)}h${runMin % 60}m` : `${runMin}m`;

        // Branch name is the agent ID — show it prominently
        const branch = slot.branch || slot.taskId.substring(0, 8);
        const shortBranch = branch.replace(/^ve\//, "");
        lines.push(`${emoji} ${shortBranch}`);
        lines.push(`   ${slot.taskTitle}`);
        lines.push(`   SDK: ${slot.sdk} | ⏱️ ${runStr} | Attempt #${slot.attempt}`);

        // Git diff stats
        if (slot.branch) {
          try {
            const diffStat = execSync(
              `git diff --shortstat main...${slot.branch} 2>nul || echo ""`,
              { cwd: repoRoot, encoding: "utf8", timeout: 8000 },
            ).trim();
            if (diffStat) {
              const insMatch = diffStat.match(/(\d+) insertion/);
              const delMatch = diffStat.match(/(\d+) deletion/);
              const filesMatch = diffStat.match(/(\d+) file/);
              lines.push(`   📊 ${filesMatch?.[1] || 0} files | +${insMatch?.[1] || 0} -${delMatch?.[1] || 0}`);
            }
          } catch { /* branch not pushed yet */ }
        }
        lines.push(""); // spacing
      }

      lines.push("────────────────────────────");
      lines.push(`Use /agentlogs <branch> for agent output`);

      await sendReply(chatId, lines.join("\n"));
      return;
      } else {
        // No active slots — show status summary
        lines.push(`📋 No active agents (0/${executorStatus.maxParallel} slots)`);
        if (executorStatus.blockedTasks?.length > 0) {
          lines.push(`\n⛔ ${executorStatus.blockedTasks.length} task(s) blocked (exceeded retry limit)`);
        }
        lines.push("");
        lines.push(executorStatus.paused
          ? `Use /resumetasks to start accepting tasks`
          : `Waiting for todo tasks in kanban...`);
      }

      await sendReply(chatId, lines.join("\n"));
      return;
    }

    // ── Fallback: read status file (legacy/VK mode) ──
    const raw = await readFile(statusPath, "utf8");
    const data = JSON.parse(raw);
    const attempts = data.attempts || {};

    if (Object.keys(attempts).length === 0) {
      await sendReply(chatId, "No active task attempts tracked.");
      return;
    }

    const lines = ["📋 Active Task Attempts\n"];

    for (const [id, attempt] of Object.entries(attempts)) {
      if (!attempt) continue;
      const status = attempt.status || "unknown";
      const emoji = status === "running" ? "🟢" : status === "review" ? "👀" : status === "error" ? "❌" : status === "completed" ? "✅" : "⏸️";
      const branch = attempt.branch || "";
      const pr = attempt.pr_number ? ` PR#${attempt.pr_number}` : "";
      const title = attempt.task_title || attempt.task_id || id;
      const shortBranch = branch ? branch.replace(/^ve\//, "") : title;

      lines.push(`${emoji} ${shortBranch}${pr}`);
      lines.push(`   ${title}`);
      lines.push(`   Status: ${status} | Agent: ${attempt.executor || "?"}`);

      const started = attempt.started_at || attempt.created_at || attempt.updated_at;
      if (started) {
        const dur = Date.now() - Date.parse(started);
        const mins = Math.floor(dur / 60000);
        const hrs = Math.floor(mins / 60);
        const remMin = mins % 60;
        const durStr = hrs > 0 ? `${hrs}h ${remMin}m` : `${mins}m`;
        lines.push(`   ⏱️ Active: ${durStr}`);
      }

      if (branch) {
        try {
          const diffStat = execSync(
            `git diff --shortstat main...${branch} 2>nul || echo ""`,
            { cwd: repoRoot, encoding: "utf8", timeout: 8000 },
          ).trim();
          if (diffStat) {
            const insMatch = diffStat.match(/(\d+) insertion/);
            const delMatch = diffStat.match(/(\d+) deletion/);
            const filesMatch = diffStat.match(/(\d+) file/);
            lines.push(`   📊 ${filesMatch?.[1] || 0} files | +${insMatch?.[1] || 0} -${delMatch?.[1] || 0}`);
          }
        } catch { /* git diff not available */ }
      }
      lines.push("");
    }

    const running = Object.values(attempts).filter((a) => a?.status === "running").length;
    const errors = Object.values(attempts).filter((a) => a?.status === "error").length;
    const reviews = Object.values(attempts).filter(
      (a) => a?.status === "review" || a?.status === "manual_review",
    ).length;
    lines.push("────────────────────────────");
    lines.push(
      `Total: ${Object.keys(attempts).length} | Running: ${running} | Review: ${reviews} | Error: ${errors}`,
    );

    await sendReply(chatId, lines.join("\n"));
  } catch (err) {
    await sendReply(chatId, `Error reading tasks: ${err.message}`);
  }
}

/**
 * /agentlogs {branch} — Show agent output for a specific branch/worktree.
 * The branch can be partial (e.g. "hpc-topology" matches "ve/73ea9114-xl-p1-feat-hpc-topology-aware-scheduling").
 * Shows: last git log, last commit diff stat, worktree status.
 */
async function cmdAgentLogs(chatId, args) {
  const query = (args || "").trim();
  if (!query) {
    await sendReply(chatId, "Usage: /agentlogs <branch-fragment>\n\nExample: /agentlogs hpc-topology");
    return;
  }

  try {
    // Find matching worktree
    const worktreeDir = resolve(repoRoot, ".cache", "worktrees");
    let dirs;
    try {
      dirs = await readdir(worktreeDir);
    } catch {
      await sendReply(chatId, "No worktrees directory found.");
      return;
    }

    const matches = dirs.filter((d) => d.toLowerCase().includes(query.toLowerCase()));
    if (matches.length === 0) {
      await sendReply(chatId, `No worktree matching "${query}".\n\nAvailable:\n${dirs.slice(0, 15).join("\n")}`);
      return;
    }

    const wtName = matches[0]; // Best match
    const wtPath = resolve(worktreeDir, wtName);
    const lines = [`📂 Agent: ${wtName}\n`];

    // Git log (last 5 commits)
    try {
      const gitLog = execSync(
        `git log --oneline -5 2>&1`,
        { cwd: wtPath, encoding: "utf8", timeout: 10000 },
      ).trim();
      if (gitLog) {
        lines.push("📝 Recent commits:");
        lines.push(gitLog);
      } else {
        lines.push("📝 No commits yet");
      }
    } catch { lines.push("📝 Git log unavailable"); }

    lines.push("");

    // Git status
    try {
      const gitStatus = execSync(
        `git status --short 2>&1`,
        { cwd: wtPath, encoding: "utf8", timeout: 10000 },
      ).trim();
      if (gitStatus) {
        const statusLines = gitStatus.split("\n");
        lines.push(`📄 Working tree: ${statusLines.length} changed files`);
        lines.push(statusLines.slice(0, 15).join("\n"));
        if (statusLines.length > 15) lines.push(`... +${statusLines.length - 15} more`);
      } else {
        lines.push("📄 Working tree: clean");
      }
    } catch { lines.push("📄 Git status unavailable"); }

    lines.push("");

    // Diff stat vs main
    try {
      const branchName = execSync(`git branch --show-current 2>&1`, { cwd: wtPath, encoding: "utf8", timeout: 5000 }).trim();
      const diffStat = execSync(
        `git diff --stat main...${branchName} 2>&1`,
        { cwd: wtPath, encoding: "utf8", timeout: 10000 },
      ).trim();
      if (diffStat) {
        const statLines = diffStat.split("\n");
        lines.push("📊 Diff vs main:");
        // Show only summary line (last line)
        lines.push(statLines[statLines.length - 1] || "(none)");
      }
    } catch { /* no diff available */ }

    lines.push("");

    // Check for active executor slot matching this branch
    const executor = _getInternalExecutor?.();
    if (executor) {
      const executorStatus = executor.getStatus?.();
      const slot = executorStatus?.slots?.find(
        (s) => s.branch && wtName.includes(s.branch.replace("ve/", "").replace(/\//g, "-"))
      );
      if (slot) {
        const runMin = Math.round(slot.runningFor / 60);
        const runStr = runMin >= 60 ? `${Math.floor(runMin / 60)}h${runMin % 60}m` : `${runMin}m`;
        lines.push(`🤖 Active agent: ${slot.sdk} | Running: ${runStr} | Attempt #${slot.attempt}`);
      } else {
        lines.push("🤖 No active agent on this branch");
      }
    }

    await sendReply(chatId, lines.join("\n"));
  } catch (err) {
    await sendReply(chatId, `Error: ${err.message}`);
  }
}


async function cmdLogs(chatId, _args) {
  const numLines = parseInt(_args, 10) || 30;
  try {
    const logFiles = await readdir(resolve(__dirname, "logs")).catch(() => []);
    const logFile = logFiles
      .filter((f) => f.endsWith(".log"))
      .sort()
      .pop(); // most recent

    if (!logFile) {
      await sendReply(chatId, "No log files found.");
      return;
    }

    const logPath = resolve(__dirname, "logs", logFile);
    const content = await readFile(logPath, "utf8");
    const lines = content.split("\n").filter(Boolean);
    const tail = lines.slice(-numLines).join("\n");

    await sendReply(
      chatId,
      `📄 Last ${numLines} lines of ${logFile}:\n\n${tail || "(empty)"}`,
    );
  } catch (err) {
    await sendReply(chatId, `Error reading logs: ${err.message}`);
  }
}

async function cmdBranches(chatId, _args) {
  try {
    const result = execSync("git branch -a --sort=-committerdate", {
      cwd: repoRoot,
      encoding: "utf8",
      timeout: 10000,
    });
    const lines = result.split("\n").filter(Boolean).slice(0, 20);
    await sendReply(
      chatId,
      `🌿 Recent branches (top 20):\n\n${lines.join("\n")}`,
    );
  } catch (err) {
    await sendReply(chatId, `Error listing branches: ${err.message}`);
  }
}

async function cmdDiff(chatId, _args) {
  try {
    const diffStat = execSync("git diff --stat HEAD", {
      cwd: repoRoot,
      encoding: "utf8",
      timeout: 10000,
    });
    if (!diffStat.trim()) {
      await sendReply(chatId, "No uncommitted changes.");
      return;
    }
    await sendReply(
      chatId,
      `📝 Working tree changes:\n\n${diffStat.slice(0, 3500)}`,
    );
  } catch (err) {
    await sendReply(chatId, `Error reading diff: ${err.message}`);
  }
}

async function cmdRestart(chatId) {
  await sendReply(chatId, "🔄 Restarting orchestrator process...");
  try {
    if (_getCurrentChild) {
      const child = _getCurrentChild();
      if (child && child.pid) {
        try {
          child.kill("SIGTERM");
        } catch {
          /* best effort */
        }
      }
    }
    // The monitor's handleExit will auto-restart the process
    await sendReply(
      chatId,
      "✅ Restart signal sent. Monitor will auto-restart the orchestrator.",
    );
  } catch (err) {
    await sendReply(chatId, `❌ Restart failed: ${err.message}`);
  }
}

async function cmdRetry(chatId, args) {
  if (!_attemptFreshSessionRetry) {
    await sendReply(
      chatId,
      "❌ Fresh session retry not available (not injected from monitor).",
    );
    return;
  }

  const reason = args?.trim() || "manual_retry_via_telegram";
  await sendReply(chatId, `🔄 Attempting fresh session retry (${reason})...`);

  try {
    const started = await _attemptFreshSessionRetry(reason);
    if (started) {
      await sendReply(
        chatId,
        "✅ Fresh session started. New agent will pick up the task.",
      );
    } else {
      await sendReply(
        chatId,
        "⚠️ Fresh session retry failed. Check logs for details (rate limit, no active attempt, or VK endpoint unavailable).",
      );
    }
  } catch (err) {
    await sendReply(chatId, `❌ Retry error: ${err.message || err}`);
  }
}

async function cmdPlan(chatId, args) {
  if (!_triggerTaskPlanner) {
    await sendReply(
      chatId,
      "❌ Task planner not available (not injected from monitor).",
    );
    return;
  }

  // Parse optional task count: /plan 5 or /plan 10
  const parsed = parseInt(args?.trim(), 10);
  const taskCount = Number.isFinite(parsed) && parsed > 0 ? parsed : 5;

  await sendReply(chatId, `📋 Triggering task planner (${taskCount} tasks)...`);

  try {
    const result = await _triggerTaskPlanner(
      "manual-telegram",
      { source: "telegram /plan command" },
      { taskCount, notify: false },
    );
    if (result?.status === "skipped") {
      if (result.reason === "planner_disabled") {
        await sendReply(
          chatId,
          "⚠️ Task planner disabled. Set TASK_PLANNER_MODE=kanban or codex-sdk.",
        );
        return;
      }
      if (result.reason === "planner_busy") {
        await sendReply(
          chatId,
          "⚠️ Task planner already running. Try again in a moment.",
        );
        return;
      }
      const lines = [
        "⚠️ Task planner skipped — a planning task already exists.",
      ];
      if (result.taskTitle) {
        lines.push(`Title: ${result.taskTitle}`);
      }
      if (result.taskId) {
        lines.push(`Task ID: ${result.taskId}`);
      }
      if (result.taskUrl) {
        lines.push(result.taskUrl);
      }
      await sendReply(chatId, lines.join("\n"));
      return;
    }
    if (result?.status === "created") {
      const lines = [
        "✅ Task planner task created.",
        result.taskTitle ? `Title: ${result.taskTitle}` : null,
        result.taskId ? `Task ID: ${result.taskId}` : null,
        result.taskUrl || null,
      ].filter(Boolean);
      await sendReply(chatId, lines.join("\n"));
      return;
    }
    if (result?.status === "completed") {
      await sendReply(
        chatId,
        `✅ Task planner completed. Output saved to ${result.outputPath}`,
      );
      return;
    }
    await sendReply(
      chatId,
      `✅ Task planner triggered for ${taskCount} tasks. Check backlog shortly.`,
    );
  } catch (err) {
    await sendReply(chatId, `❌ Task planner error: ${err.message || err}`);
  }
}

async function cmdCleanupMerged(chatId) {
  if (!_reconcileTaskStatuses) {
    await sendReply(
      chatId,
      "❌ Cleanup not available (not injected from monitor).",
    );
    return;
  }
  await sendReply(
    chatId,
    "🧹 Reconciling VK task statuses with PR/branch state…",
  );
  try {
    const result = await _reconcileTaskStatuses("manual-telegram");
    const lines = [
      "✅ Cleanup complete.",
      `Checked: ${result?.checked ?? 0}`,
      `Moved to done: ${result?.movedDone ?? 0}`,
      `Moved to inreview: ${result?.movedReview ?? 0}`,
    ];
    await sendReply(chatId, lines.join("\n"));
  } catch (err) {
    await sendReply(chatId, `❌ Cleanup error: ${err.message || err}`);
  }
}

async function cmdHistory(chatId) {
  const info = getPrimaryAgentInfo();
  const sessionLabel = info.sessionId || info.threadId || "(none)";
  const agentLabel = info.adapter || info.provider || getPrimaryAgentName();
  const lines = [
    `🧠 Primary Agent (${agentLabel})`,
    "",
    `Session: ${sessionLabel}`,
    `Turns: ${info.turnCount}`,
    `Active: ${info.isActive ? "yes" : "no"}`,
    `Busy: ${info.isBusy ? "yes" : "no"}`,
    info.workspacePath ? `Workspace: ${info.workspacePath}` : "",
    info.fallbackReason ? `Fallback: ${info.fallbackReason}` : "",
    "",
    "The session persists across messages.",
    "Use /clear to start a fresh session.",
  ];
  await sendReply(chatId, lines.filter(Boolean).join("\n"));
}

async function cmdClear(chatId) {
  await resetPrimaryAgent();
  await sendReply(
    chatId,
    "🧹 Agent session reset. Next message starts a fresh conversation.",
  );
}

async function cmdGit(chatId, gitArgs) {
  if (!gitArgs) {
    await sendReply(
      chatId,
      "Usage: /git <command>\nExample: /git log --oneline -10",
    );
    return;
  }

  // Safety: block destructive commands
  const dangerous = ["push", "reset --hard", "clean -fd", "checkout -f"];
  const lower = gitArgs.toLowerCase();
  if (dangerous.some((d) => lower.startsWith(d))) {
    await sendReply(
      chatId,
      `⚠️ Blocked: 'git ${gitArgs}' is a destructive command. Use the agent shell for that.`,
    );
    return;
  }

  try {
    const result = execSync(`git ${gitArgs}`, {
      cwd: repoRoot,
      encoding: "utf8",
      timeout: 15000,
    });
    await sendReply(
      chatId,
      `$ git ${gitArgs}\n\n${result.slice(0, 3800) || "(no output)"}`,
    );
  } catch (err) {
    await sendReply(
      chatId,
      `$ git ${gitArgs}\n\n❌ ${err.message?.slice(0, 1500) || err}`,
    );
  }
}

async function cmdShell(chatId, shellArgs) {
  if (!shellArgs) {
    await sendReply(
      chatId,
      "Usage: /shell <command>\nExample: /shell ls -la scripts/",
    );
    return;
  }

  // Safety: block very destructive patterns
  const dangerous = ["rm -rf /", "format", "del /f /s", "shutdown", "reboot"];
  const lower = shellArgs.toLowerCase();
  if (dangerous.some((d) => lower.includes(d))) {
    await sendReply(chatId, `⚠️ Blocked: '${shellArgs}' looks destructive.`);
    return;
  }

  try {
    const isWin = process.platform === "win32";
    const result = execSync(shellArgs, {
      cwd: repoRoot,
      encoding: "utf8",
      timeout: 30000,
      shell: isWin ? "cmd.exe" : "/bin/sh",
    });
    await sendReply(
      chatId,
      `$ ${shellArgs}\n\n${result.slice(0, 3800) || "(no output)"}`,
    );
  } catch (err) {
    const stderr = err.stderr ? err.stderr.toString().slice(0, 1000) : "";
    const stdout = err.stdout ? err.stdout.toString().slice(0, 1000) : "";
    await sendReply(
      chatId,
      `$ ${shellArgs}\n\n❌ ${stderr || stdout || err.message}`,
    );
  }
}

// ── Region / Health / Model Override Commands ─────────────────────────────────

function runPwsh(psScript, timeoutMs = 15000) {
  const isWin = process.platform === "win32";
  const pwsh = isWin ? "powershell.exe" : "pwsh";
  const script = `& { ${psScript} }`;
  const result = spawnSync(pwsh, ["-NoProfile", "-Command", script], {
    cwd: repoRoot,
    encoding: "utf8",
    timeout: timeoutMs,
  });
  if (result.error) {
    throw new Error(result.error.message);
  }
  if (result.status !== 0) {
    throw new Error(
      (result.stderr || result.stdout || "").trim() ||
        `powershell command failed (exit ${result.status})`,
    );
  }
  return result.stdout;
}

async function readStatusSnapshot() {
  try {
    const raw = await readFile(statusPath, "utf8");
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

function buildExecutorKey(executor, variant) {
  const exec = (executor || "unknown").toString().trim().toUpperCase();
  const varNorm = variant ? String(variant).trim().toUpperCase() : "";
  return varNorm ? `${exec}:${varNorm}` : exec;
}

function buildExecutorHealthFromStatus(statusData) {
  const metrics = new Map();
  const attempts = statusData?.attempts ?? {};
  for (const info of Object.values(attempts)) {
    if (!info) continue;
    const key = buildExecutorKey(info.executor, info.executor_variant);
    const entry = metrics.get(key) || {
      active: 0,
      failures: 0,
      successes: 0,
      timeouts: 0,
      rate_limits: 0,
    };

    const processStatus = String(
      info.last_process_status || info.status || "",
    ).toLowerCase();
    const trackedStatus = String(info.status || "").toLowerCase();

    if (
      ["running", "queued", "in_progress", "active"].includes(processStatus) ||
      trackedStatus === "running"
    ) {
      entry.active += 1;
    } else if (
      ["failed", "error", "killed", "aborted"].includes(processStatus) ||
      trackedStatus === "error"
    ) {
      entry.failures += 1;
    } else if (
      ["completed", "success", "review", "done"].includes(processStatus) ||
      ["review", "completed", "done"].includes(trackedStatus)
    ) {
      entry.successes += 1;
    }

    if (processStatus.includes("timeout")) entry.timeouts += 1;
    if (processStatus.includes("rate") || processStatus.includes("limit")) {
      entry.rate_limits += 1;
    }

    metrics.set(key, entry);
  }
  return metrics;
}

function deriveExecutorStatus(stats) {
  if (!stats) return "unknown";
  if ((stats.rate_limits ?? 0) > 0) return "cooldown";
  if ((stats.failures ?? 0) > 0 || (stats.timeouts ?? 0) > 0) {
    return "degraded";
  }
  if ((stats.active ?? 0) > 0 || (stats.successes ?? 0) > 0) {
    return "healthy";
  }
  return "unknown";
}

function buildExecutorHealthEntries(executorConfig, metrics) {
  const entries = [];
  const usedKeys = new Set();
  const executors = executorConfig?.executors ?? [];

  for (const exec of executors) {
    const key = buildExecutorKey(exec.executor, exec.variant);
    const stats =
      metrics.get(key) ||
      metrics.get(buildExecutorKey(exec.executor, null)) ||
      null;
    usedKeys.add(key);
    entries.push({
      label:
        exec.executor && exec.variant
          ? `${exec.executor}/${exec.variant}`
          : exec.executor || exec.name || key,
      tier: exec.tier || exec.role || "default",
      region: exec.region || exec.variant || "default",
      status: deriveExecutorStatus(stats),
      stats: stats || {
        active: 0,
        failures: 0,
        successes: 0,
        timeouts: 0,
        rate_limits: 0,
      },
    });
  }

  for (const [key, stats] of metrics.entries()) {
    if (usedKeys.has(key)) continue;
    entries.push({
      label: key.replace(":", "/"),
      tier: "default",
      region: "default",
      status: deriveExecutorStatus(stats),
      stats,
    });
  }

  return entries;
}

async function cmdRegion(chatId, regionArg) {
  if (!regionArg || regionArg.trim() === "") {
    // Show current region status
    try {
      const result = runPwsh(
        `. '${resolve(repoRoot, "scripts", "codex-monitor", "ve-kanban.ps1")}'; Initialize-CodexRegionTracking; Get-RegionStatus | ConvertTo-Json -Depth 3`,
      );
      const status = JSON.parse(result);
      const lines = [
        "🌍 Codex Region Status",
        "",
        `Active: ${status.active_region?.toUpperCase() || "unknown"}`,
        `Override: ${status.override || "auto"}`,
        `Sweden available: ${status.sweden_available ? "✅" : "❌"}`,
        `Cooldown: ${status.cooldown_min}min`,
      ];
      if (status.switched_ago_min !== null) {
        lines.push(`Switched: ${status.switched_ago_min}min ago`);
      }
      if (status.active_region === "sweden") {
        lines.push(
          `Auto-restore to US: ${status.cooldown_expired ? "ready" : `in ${Math.round(status.cooldown_min - (status.switched_ago_min || 0))}min`}`,
        );
      }
      lines.push("", "Usage: /region us | /region sweden | /region auto");
      await sendReply(chatId, lines.join("\n"));
    } catch (err) {
      await sendReply(chatId, `Error reading region: ${err.message}`);
    }
    return;
  }

  const target = regionArg.trim().toLowerCase();
  if (!["us", "sweden", "auto"].includes(target)) {
    await sendReply(
      chatId,
      "Usage: /region us | /region sweden | /region auto",
    );
    return;
  }

  try {
    const psCmd =
      target === "auto"
        ? `. '${resolve(repoRoot, "scripts", "codex-monitor", "ve-kanban.ps1")}'; Set-RegionOverride -Region $null | ConvertTo-Json`
        : `. '${resolve(repoRoot, "scripts", "codex-monitor", "ve-kanban.ps1")}'; Set-RegionOverride -Region '${target}' | ConvertTo-Json`;
    const result = runPwsh(psCmd);
    const info = JSON.parse(result);
    const icon = info.changed ? "✅" : "ℹ️";
    await sendReply(
      chatId,
      `${icon} Region: ${info.region?.toUpperCase()}\nReason: ${info.reason}`,
    );
  } catch (err) {
    await sendReply(chatId, `❌ Region switch failed: ${err.message}`);
  }
}

async function cmdHealth(chatId) {
  try {
    const statusData = await readStatusSnapshot();
    let executorConfig = null;
    try {
      executorConfig = loadExecutorConfig(__dirname, null);
    } catch {
      executorConfig = null;
    }
    const metrics = buildExecutorHealthFromStatus(statusData);
    const arr = buildExecutorHealthEntries(executorConfig, metrics);

    const iconMap = {
      healthy: "✅",
      degraded: "⚠️",
      cooldown: "⏸️",
      disabled: "❌",
    };
    const lines = ["🏥 Executor Health Dashboard\n"];

    if (!arr.length) {
      lines.push("No executor data available.");
    }

    for (const e of arr) {
      const icon = iconMap[e.status] || "❓";
      lines.push(
        `${icon} ${e.label} (${e.tier}/${e.region})\n` +
          `   Status: ${e.status} | Active: ${e.stats.active}\n` +
          `   ✓${e.stats.successes} ✗${e.stats.failures} ⏱${e.stats.timeouts} 🚫${e.stats.rate_limits}`,
      );
    }

    // Add region info
    try {
      const regionScript = [
        `. '${resolve(repoRoot, "scripts", "codex-monitor", "ve-kanban.ps1")}';`,
        "Initialize-CodexRegionTracking;",
        "Get-RegionStatus | ConvertTo-Json",
      ].join(" ");
      const regionResult = runPwsh(regionScript, 10000);
      const region = JSON.parse(regionResult);
      lines.push(
        "",
        `🌍 Region: ${region.active_region?.toUpperCase()} ${region.override ? `(override: ${region.override})` : "(auto)"}`,
        `Sweden backup: ${region.sweden_available ? "available" : "not configured"}`,
      );
    } catch {
      lines.push("", "🌍 Region: unavailable");
    }

    await sendReply(chatId, lines.join("\n"));
  } catch (err) {
    await sendReply(chatId, `Error reading health: ${err.message}`);
  }
}

async function cmdPresence(chatId) {
  await ensurePresenceReady();
  const child = _getCurrentChild ? _getCurrentChild() : null;
  const payload = buildLocalPresence({
    orchestrator_running: !!child,
    orchestrator_pid: child?.pid ?? null,
    vk_url: _getVibeKanbanUrl ? _getVibeKanbanUrl() : null,
  });
  await notePresence(payload, {
    source: "local",
    receivedAt: payload.updated_at,
  });
  const nowMs = Date.now();
  const summary = formatPresenceSummary({ nowMs, ttlMs: presenceTtlMs });
  await sendReply(chatId, summary);
}

async function cmdCoordinator(chatId) {
  await ensurePresenceReady();
  const nowMs = Date.now();
  const summary = formatCoordinatorSummary({ nowMs, ttlMs: presenceTtlMs });
  await sendReply(chatId, summary);
}

/** State for model override — write a file that orchestrator reads */
const modelOverridePath = resolve(repoRoot, ".cache", "executor-override.json");

async function cmdModel(chatId, modelArg) {
  if (!modelArg || modelArg.trim() === "") {
    // Show current model routing info
    try {
      const exists = existsSync(modelOverridePath);
      let overrideText = "none (auto routing)";
      if (exists) {
        const raw = await readFile(modelOverridePath, "utf8");
        const data = JSON.parse(raw);
        if (
          data.model &&
          (!data.expires_at || new Date(data.expires_at) > new Date())
        ) {
          overrideText = `${data.model} (until ${data.expires_at || "cleared"})`;
        }
      }
      const lines = [
        "🤖 Model Routing",
        "",
        `Override: ${overrideText}`,
        "",
        "Available models:",
        "  gpt-5.2-codex      — Primary, best speed/quality",
        "  gpt-5.1-codex-max  — Large tasks, extra capacity",
        "  gpt-5.1-codex-mini — Small tasks, subagent-optimized",
        "  claude-opus-4.6    — Supreme quality, complex refactors",
        "  claude-code        — Claude Code CLI fallback",
        "",
        "Usage:",
        "  /model gpt-5.2-codex     Set override for next 3 tasks",
        "  /model auto              Clear override (smart routing)",
      ];
      await sendReply(chatId, lines.join("\n"));
    } catch (err) {
      await sendReply(chatId, `Error: ${err.message}`);
    }
    return;
  }

  const target = modelArg.trim().toLowerCase();

  if (target === "auto" || target === "clear") {
    try {
      if (existsSync(modelOverridePath)) {
        await unlink(modelOverridePath);
      }
      await sendReply(
        chatId,
        "✅ Model override cleared. Smart routing active.",
      );
    } catch (err) {
      await sendReply(chatId, `❌ Error: ${err.message}`);
    }
    return;
  }

  const validModels = [
    "gpt-5.2-codex",
    "gpt-5.1-codex-max",
    "gpt-5.1-codex-mini",
    "claude-opus-4.6",
    "claude-code",
  ];
  if (!validModels.includes(target)) {
    await sendReply(
      chatId,
      `Unknown model: ${target}\nValid: ${validModels.join(", ")}`,
    );
    return;
  }

  try {
    const override = {
      model: target,
      remaining_tasks: 3,
      set_at: new Date().toISOString(),
      expires_at: new Date(Date.now() + 60 * 60 * 1000).toISOString(), // 1 hour
    };
    await writeFile(
      modelOverridePath,
      JSON.stringify(override, null, 2),
      "utf8",
    );
    await sendReply(
      chatId,
      `✅ Model override set: ${target}\nApplies to next 3 tasks (or 1 hour)`,
    );
  } catch (err) {
    await sendReply(chatId, `❌ Error: ${err.message}`);
  }
}

async function cmdKanban(chatId, backendArg) {
  if (!backendArg || backendArg.trim() === "") {
    const current = getKanbanBackendName();
    const available = getAvailableBackends();
    const lines = [
      "📋 Kanban Backend Status",
      "",
      `Active: ${current}`,
      `Available: ${available.join(", ")}`,
      "",
      "Switch backend:",
      "  /kanban vk        Vibe-Kanban (default)",
      "  /kanban github     GitHub Issues",
      "  /kanban jira       Jira (stub)",
    ];
    await sendReply(chatId, lines.join("\n"));
    return;
  }

  const target = backendArg.trim().toLowerCase();
  const validBackends = getAvailableBackends();

  if (!validBackends.includes(target)) {
    await sendReply(
      chatId,
      `Unknown backend: ${target}\nValid: ${validBackends.join(", ")}`,
    );
    return;
  }

  try {
    setKanbanBackend(target);
    await sendReply(
      chatId,
      `✅ Kanban backend switched to: ${target}\nActive: ${getKanbanBackendName()}`,
    );
  } catch (err) {
    await sendReply(chatId, `❌ Error switching backend: ${err.message}`);
  }
}

async function cmdThreads(chatId, subArg) {
  if (subArg && subArg.trim().toLowerCase() === "clear") {
    clearThreadRegistry();
    await sendReply(chatId, "✅ Thread registry cleared.");
    return;
  }

  if (subArg && subArg.trim().toLowerCase().startsWith("kill ")) {
    const taskKey = subArg.trim().substring(5).trim();
    if (!taskKey) {
      await sendReply(chatId, "Usage: /threads kill <taskKey>");
      return;
    }
    invalidateThread(taskKey);
    await sendReply(chatId, `✅ Thread for "${taskKey}" invalidated.`);
    return;
  }

  const threads = getActiveThreads();
  if (threads.length === 0) {
    await sendReply(
      chatId,
      "🧵 No active agent threads.\n\nThreads are created when tasks run via the agent pool with thread persistence.",
    );
    return;
  }

  const lines = [`🧵 Active Agent Threads (${threads.length})`, ""];

  for (const t of threads) {
    const ageMin = Math.round(t.age / 60_000);
    lines.push(
      `• ${t.taskKey}`,
      `  SDK: ${t.sdk} | Turns: ${t.turnCount} | Age: ${ageMin}m`,
      `  Thread: ${t.threadId ? t.threadId.slice(0, 12) + "…" : "(none)"}`,
      "",
    );
  }

  lines.push(
    "Commands:",
    "  /threads clear          Clear all thread records",
    "  /threads kill <taskKey>  Invalidate a specific thread",
  );

  await sendReply(chatId, lines.join("\n"));
}

/**
 * /worktrees — View and manage git worktrees.
 *
 * Subcommands:
 *   /worktrees           — Show all active worktrees with branch, task, age
 *   /worktrees stats     — Show aggregate statistics
 *   /worktrees prune     — Prune stale/orphaned worktrees
 *   /worktrees release <taskKey> — Release a specific worktree by task key
 */
async function cmdWorktrees(chatId, args) {
  const parts = args ? args.trim().split(/\s+/) : [];
  const sub = parts[0]?.toLowerCase();

  if (sub === "prune") {
    // Prune stale worktrees
    try {
      const result = await pruneStaleWorktrees();
      const lines = [`🧹 Worktree prune complete:`];
      lines.push(`  Pruned: ${result.pruned}`);
      lines.push(`  Registry evicted: ${result.evicted}`);
      await sendReply(chatId, lines.join("\n"));
    } catch (err) {
      await sendReply(chatId, `❌ Prune failed: ${err.message}`);
    }
    return;
  }

  if (sub === "release") {
    const taskKey = parts[1];
    if (!taskKey) {
      await sendReply(chatId, "Usage: /worktrees release <taskKey>");
      return;
    }
    try {
      const wm = getWorktreeManager();
      const result = wm.releaseWorktree(taskKey);
      if (result.success) {
        await sendReply(
          chatId,
          `✅ Released worktree for "${taskKey}": ${result.path}`,
        );
      } else {
        await sendReply(
          chatId,
          `⚠️ No worktree found for task key "${taskKey}"`,
        );
      }
    } catch (err) {
      await sendReply(chatId, `❌ Release failed: ${err.message}`);
    }
    return;
  }

  if (sub === "stats") {
    try {
      const stats = getWorktreeStats();
      const lines = [`📊 Worktree Stats:`];
      lines.push(`  Total tracked: ${stats.total}`);
      lines.push(`  Active: ${stats.active}`);
      lines.push(`  Stale: ${stats.stale}`);
      if (Object.keys(stats.byOwner).length > 0) {
        lines.push(`  By owner:`);
        for (const [owner, count] of Object.entries(stats.byOwner)) {
          lines.push(`    ${owner}: ${count}`);
        }
      }
      await sendReply(chatId, lines.join("\n"));
    } catch (err) {
      await sendReply(chatId, `❌ Stats failed: ${err.message}`);
    }
    return;
  }

  // Default: list all active worktrees
  try {
    const worktrees = listManagedWorktrees();
    if (!worktrees || worktrees.length === 0) {
      await sendReply(chatId, "🌳 No active worktrees tracked.");
      return;
    }

    const lines = [`🌳 Active Worktrees (${worktrees.length}):\n`];
    for (const wt of worktrees) {
      const ageMin = Math.round((wt.age || 0) / 60000);
      const ageStr =
        ageMin >= 60 ? `${Math.round(ageMin / 60)}h` : `${ageMin}m`;
      const branch = wt.branch || "(detached)";
      const taskKey = wt.taskKey ? ` [${wt.taskKey}]` : "";
      const owner = wt.owner ? ` (${wt.owner})` : "";
      const status = wt.status || "active";
      lines.push(`• ${branch}${taskKey}${owner}`);
      lines.push(`  Status: ${status} | Age: ${ageStr}`);
      lines.push(`  Path: ${wt.path}`);
    }

    lines.push(
      `\nCommands: /worktrees prune | /worktrees release <key> | /worktrees stats`,
    );
    await sendReply(chatId, lines.join("\n"));
  } catch (err) {
    await sendReply(chatId, `❌ Worktree list failed: ${err.message}`);
  }
}

/**
 * /executor — View and manage the internal task executor.
 *
 * Subcommands:
 *   /executor           — Show status (mode, active slots, SDK, etc.)
 *   /executor status    — Same as above
 *   /executor slots     — Show active task slots with details
 *   /executor mode <vk|internal|hybrid> — Show current mode (runtime switch not supported)
 */
async function cmdExecutor(chatId, args) {
  const parts = args ? args.trim().split(/\s+/) : [];
  const sub = parts[0]?.toLowerCase();

  // Get monitor functions for executor access
  const executor = _getInternalExecutor?.();
  const mode = _getExecutorMode?.() || "vk";

  if (sub === "slots") {
    if (!executor) {
      await sendReply(
        chatId,
        `⚙️ Internal executor not active (mode: ${mode})`,
      );
      return;
    }
    const status = executor.getStatus();
    if (status.slots.length === 0) {
      await sendReply(
        chatId,
        `⚙️ No active task slots (${status.activeSlots}/${status.maxParallel} used)`,
      );
      return;
    }
    const lines = [
      `⚙️ Active Task Slots (${status.activeSlots}/${status.maxParallel}):\n`,
    ];
    for (const slot of status.slots) {
      const runMin = Math.round(slot.runningFor / 60);
      const runStr =
        runMin >= 60
          ? `${Math.round(runMin / 60)}h${runMin % 60}m`
          : `${runMin}m`;
      lines.push(`• ${slot.taskTitle}`);
      lines.push(`  ID: ${slot.taskId.substring(0, 8)} | SDK: ${slot.sdk}`);
      lines.push(`  Branch: ${slot.branch}`);
      lines.push(
        `  Running: ${runStr} | Attempt: ${slot.attempt} | Status: ${slot.status}`,
      );
    }
    await sendReply(chatId, lines.join("\n"));
    return;
  }

  if (sub === "mode") {
    const target = parts[1]?.toLowerCase();
    if (target && ["vk", "internal", "hybrid"].includes(target)) {
      await sendReply(
        chatId,
        `⚙️ Current mode: ${mode}\n` +
          `ℹ️ Mode can be changed via EXECUTOR_MODE env var or config.\n` +
          `Restart the monitor after changing to apply.`,
      );
    } else {
      await sendReply(
        chatId,
        `⚙️ Current executor mode: ${mode}\n\nValid modes: vk, internal, hybrid`,
      );
    }
    return;
  }

  // Default: show status
  const lines = [`⚙️ Executor Status\n`];
  lines.push(`Mode: ${mode}`);

  if (executor) {
    const status = executor.getStatus();
    lines.push(`Running: ${status.running ? "✅ Yes" : "❌ No"}`);
    lines.push(`SDK: ${status.sdk}`);
    lines.push(`Active Slots: ${status.activeSlots}/${status.maxParallel}`);
    lines.push(`Poll Interval: ${status.pollIntervalMs / 1000}s`);
    lines.push(`Task Timeout: ${Math.round(status.taskTimeoutMs / 60000)}min`);
    lines.push(`Max Retries: ${status.maxRetries}`);
    lines.push(`Cooldowns: ${status.cooldowns}`);
    if (status.projectId) {
      lines.push(`Project ID: ${status.projectId.substring(0, 8)}...`);
    }
  } else {
    lines.push(`Internal executor: not active`);
    if (mode === "vk") {
      lines.push(
        `\nℹ️ Using VK executor only. Set EXECUTOR_MODE=internal or hybrid to enable.`,
      );
    }
  }

  lines.push(`\nCommands: /executor slots | /executor mode`);
  await sendReply(chatId, lines.join("\n"));
}

async function cmdSdk(chatId, sdkArg) {
  if (!sdkArg || sdkArg.trim() === "") {
    // Show current SDK info
    const poolSdk = getPoolSdkName();
    const primaryAgent = getPrimaryAgentName();
    const available = getAvailableSdks();
    const lines = [
      "🔌 Agent SDK Status",
      "",
      `Pool SDK: ${poolSdk}`,
      `Primary Agent: ${primaryAgent}`,
      `Available: ${available.join(", ") || "(none)"}`,
      "",
      "Switch SDK:",
      "  /sdk copilot    Use Copilot SDK",
      "  /sdk codex      Use Codex SDK",
      "  /sdk claude     Use Claude SDK",
      "  /sdk auto       Reset to config default",
    ];
    await sendReply(chatId, lines.join("\n"));
    return;
  }

  const target = sdkArg.trim().toLowerCase().replace(/-sdk$/, "");

  if (target === "auto" || target === "reset") {
    resetPoolSdkCache();
    await sendReply(
      chatId,
      "✅ Agent pool SDK reset to config default.\nCurrent: " +
        getPoolSdkName(),
    );
    return;
  }

  const validSdks = ["codex", "copilot", "claude"];
  if (!validSdks.includes(target)) {
    await sendReply(
      chatId,
      `Unknown SDK: ${target}\nValid: ${validSdks.join(", ")}, auto`,
    );
    return;
  }

  try {
    // Switch pool SDK
    setPoolSdk(target);

    // Also switch primary agent to match
    const switchResult = await switchPrimaryAgent(`${target}-sdk`);
    const primaryStatus = switchResult.ok
      ? `Primary agent: ${switchResult.name}`
      : `Primary agent switch failed: ${switchResult.reason}`;

    await sendReply(
      chatId,
      `✅ SDK switched to: ${target}\nPool SDK: ${getPoolSdkName()}\n${primaryStatus}`,
    );
  } catch (err) {
    await sendReply(chatId, `❌ Error switching SDK: ${err.message}`);
  }
}

async function cmdSharedWorkspaces(chatId, rawArgs) {
  const registry = await loadSharedRegistry();
  const sweep = await sweepSharedLeases({
    registry,
    actor: `telegram:${chatId}`,
  });
  const tokens = splitArgs(rawArgs);
  if (tokens.length > 0) {
    const workspace = resolveSharedWorkspace(sweep.registry, tokens[0]);
    if (!workspace) {
      await sendReply(chatId, `Unknown shared workspace '${tokens[0]}'.`);
      return;
    }
    await sendReply(chatId, formatSharedWorkspaceDetail(workspace));
    return;
  }
  await sendReply(chatId, formatSharedWorkspaceSummary(sweep.registry));
}

async function cmdSharedWorkspaceClaim(chatId, rawArgs) {
  const parsed = parseSharedWorkspaceArgs(rawArgs);
  if (!parsed.workspaceId) {
    await sendReply(
      chatId,
      "Usage: /claim <id> [--owner <id>] [--ttl <minutes>] [--note <text>]",
    );
    return;
  }
  const actor = `telegram:${chatId}`;
  const owner = parsed.owner || actor;
  const result = await claimSharedWorkspace({
    workspaceId: parsed.workspaceId,
    owner,
    ttlMinutes: parsed.ttlMinutes,
    note: parsed.note,
    force: parsed.force,
    actor,
  });
  if (result.error) {
    await sendReply(chatId, `❌ ${result.error}`);
    return;
  }
  await sendReply(
    chatId,
    `✅ Claimed ${result.workspace.id} for ${result.lease.owner} (expires ${result.lease.lease_expires_at})`,
  );
}

async function cmdSharedWorkspaceRelease(chatId, rawArgs) {
  const parsed = parseSharedWorkspaceArgs(rawArgs);
  if (!parsed.workspaceId) {
    await sendReply(
      chatId,
      "Usage: /release <id> [--owner <id>] [--reason <text>] [--force]",
    );
    return;
  }
  const actor = `telegram:${chatId}`;
  const result = await releaseSharedWorkspace({
    workspaceId: parsed.workspaceId,
    owner: parsed.owner,
    reason: parsed.reason,
    force: parsed.force,
    actor,
  });
  if (result.error) {
    await sendReply(chatId, `❌ ${result.error}`);
    return;
  }
  await sendReply(chatId, `✅ Released ${result.workspace.id}`);
}

// ── /agent — route to workspace registry ────────────────────────────────────

const MODEL_PROFILE_MAP = {
  "gpt-5.2-codex": {
    executor: "CODEX",
    variant: "DEFAULT",
    model: "gpt-5.2-codex",
  },
  "gpt-5.1-codex-max": {
    executor: "CODEX",
    variant: "DEFAULT",
    model: "gpt-5.1-codex-max",
  },
  "gpt-5.1-codex-mini": {
    executor: "CODEX",
    variant: "DEFAULT",
    model: "gpt-5.1-codex-mini",
  },
  "claude-opus-4.6": {
    executor: "COPILOT",
    variant: "CLAUDE_OPUS_4_6",
    model: "claude-opus-4.6",
  },
  "claude-code": {
    executor: "COPILOT",
    variant: "CLAUDE_CODE",
    model: "claude-code",
  },
};

function normalizeHost(host) {
  if (!host) return null;
  const trimmed = String(host).trim();
  if (!trimmed) return null;
  if (/^https?:\/\//i.test(trimmed)) return trimmed;
  return `http://${trimmed}`;
}

function buildExecutorProfile(model, customProfile) {
  if (customProfile && typeof customProfile === "object") {
    const profile = { ...customProfile };
    if (model && !profile.model) profile.model = model;
    return profile;
  }
  if (!model) return null;
  return MODEL_PROFILE_MAP[model] || { model };
}

function resolveModelSelection(workspace, preferredModel) {
  const priorities = Array.isArray(workspace.model_priority)
    ? workspace.model_priority
    : getDefaultModelPriority();
  const candidates = [];
  if (preferredModel) candidates.push(preferredModel);
  candidates.push(...priorities);

  for (const entry of candidates) {
    if (!entry) continue;
    if (typeof entry === "string") {
      const model = entry.trim();
      if (!model) continue;
      return { model, profile: buildExecutorProfile(model) };
    }
    if (typeof entry === "object") {
      const model = entry.model || entry.name || null;
      const profile = buildExecutorProfile(model, entry);
      if (profile) {
        return { model: model || profile.model || null, profile };
      }
    }
  }
  return { model: null, profile: null };
}

async function vkRequest(host, path, options = {}) {
  const { method = "GET", body, timeoutMs = 15000 } = options;
  const base = normalizeHost(host);
  if (!base) {
    throw new Error("Workspace host missing");
  }
  const url = new URL(path, base);
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort("timeout"), timeoutMs);
  try {
    const res = await fetch(url.toString(), {
      method,
      headers: { "Content-Type": "application/json" },
      body: body ? JSON.stringify(body) : undefined,
      signal: controller.signal,
    });
    const text = await res.text();
    if (!res.ok) {
      throw new Error(
        `VK ${res.status}: ${text.slice(0, 200) || res.statusText}`,
      );
    }
    let data = null;
    if (text) {
      try {
        data = JSON.parse(text);
      } catch (err) {
        throw new Error(`VK response parse error: ${err.message}`);
      }
    }
    if (data && data.success === false) {
      throw new Error(data.message || "VK API error");
    }
    return data?.data ?? data;
  } finally {
    clearTimeout(timer);
  }
}

async function getWorkspaceSummaries(host) {
  const data = await vkRequest(host, "/api/task-attempts/summary", {
    method: "POST",
    body: { archived: false },
  });
  if (!data) return [];
  if (Array.isArray(data.summaries)) return data.summaries;
  return Array.isArray(data) ? data : [data];
}

function scoreWorkspace(summary) {
  const status = String(
    summary?.latest_process_status ?? summary?.status ?? summary?.state ?? "",
  ).toLowerCase();
  const busy = ["running", "queued", "in_progress", "active"];
  const idle = ["completed", "idle", "success", "done"];
  const failed = ["failed", "error", "crashed", "killed", "aborted"];
  if (busy.includes(status)) {
    return { available: false, score: 0, status: "busy" };
  }
  if (idle.includes(status)) {
    return { available: true, score: 3, status: "healthy" };
  }
  if (failed.includes(status)) {
    return { available: true, score: 1, status: "degraded" };
  }
  return { available: true, score: 2, status: status || "unknown" };
}

async function getWorkspaceHealth(workspaces) {
  const health = new Map();
  const hostMap = new Map();
  for (const ws of workspaces) {
    const host = normalizeHost(ws.host);
    if (!host) {
      health.set(ws.id, { available: true, score: 1, status: "unknown" });
      continue;
    }
    if (!hostMap.has(host)) hostMap.set(host, []);
    hostMap.get(host).push(ws);
  }

  for (const [host, wsList] of hostMap.entries()) {
    let summaries = [];
    try {
      summaries = await getWorkspaceSummaries(host);
    } catch {
      for (const ws of wsList) {
        health.set(ws.id, { available: true, score: 1, status: "unknown" });
      }
      continue;
    }

    const summaryMap = new Map();
    for (const summary of summaries) {
      if (summary?.workspace_id) {
        summaryMap.set(summary.workspace_id, summary);
      }
    }
    for (const ws of wsList) {
      const summary = summaryMap.get(ws.id);
      if (summary) {
        const scored = scoreWorkspace(summary);
        const last = Date.parse(
          summary.latest_process_completed_at ||
            summary.latest_process_started_at ||
            summary.updated_at ||
            "",
        );
        health.set(ws.id, {
          ...scored,
          lastCompletedAt: Number.isFinite(last) ? last : null,
        });
      } else {
        health.set(ws.id, { available: true, score: 1, status: "unknown" });
      }
    }
  }

  return health;
}

function selectWorkspace(candidates, healthMap, options = {}) {
  const { preferredId } = options;
  const scored = candidates.map((ws) => {
    const h = healthMap.get(ws.id) || {
      available: true,
      score: 1,
      status: "unknown",
    };
    return { ws, health: h };
  });

  const sortFn = (a, b) => {
    const scoreDiff = (b.health.score ?? 0) - (a.health.score ?? 0);
    if (scoreDiff !== 0) return scoreDiff;
    const lastA = a.health.lastCompletedAt || 0;
    const lastB = b.health.lastCompletedAt || 0;
    return lastB - lastA;
  };

  if (preferredId) {
    const target = scored.find((item) => item.ws.id === preferredId);
    if (target && target.health.available) {
      return { workspace: target.ws, health: target.health };
    }
    const fallback = scored.sort(sortFn)[0];
    return {
      workspace: fallback?.ws || null,
      health: fallback?.health || null,
      fallbackFrom: target?.ws || null,
    };
  }

  const best = scored.sort(sortFn)[0];
  return { workspace: best?.ws || null, health: best?.health || null };
}

function rankWorkspaceCandidates(candidates, healthMap, options = {}) {
  const { preferredId } = options;
  const scored = candidates.map((ws) => {
    const h = healthMap.get(ws.id) || {
      available: true,
      score: 1,
      status: "unknown",
    };
    return { ws, health: h };
  });

  const sortFn = (a, b) => {
    const scoreDiff = (b.health.score ?? 0) - (a.health.score ?? 0);
    if (scoreDiff !== 0) return scoreDiff;
    const lastA = a.health.lastCompletedAt || 0;
    const lastB = b.health.lastCompletedAt || 0;
    return lastB - lastA;
  };

  const sorted = scored.sort(sortFn).map((item) => item.ws);
  if (!preferredId) return sorted;
  const preferred = scored.find((item) => item.ws.id === preferredId);
  if (!preferred) return sorted;
  return [
    preferred.ws,
    ...sorted.filter((candidate) => candidate.id !== preferred.ws.id),
  ];
}

function pickLatestSession(sessions) {
  if (!Array.isArray(sessions) || sessions.length === 0) return null;
  return [...sessions].sort((a, b) => {
    const ta = Date.parse(a.updated_at || a.created_at || 0) || 0;
    const tb = Date.parse(b.updated_at || b.created_at || 0) || 0;
    return tb - ta;
  })[0];
}

async function dispatchAgentMessage(workspace, message, options = {}) {
  const host = normalizeHost(workspace.host);
  const executorProfile = options.executorProfile || null;
  const sessions = await vkRequest(
    host,
    `/api/sessions?workspace_id=${encodeURIComponent(workspace.id)}`,
  );
  let session = pickLatestSession(sessions);
  let created = false;

  if (!session || options.newSession) {
    session = await vkRequest(host, "/api/sessions", {
      method: "POST",
      body: { workspace_id: workspace.id },
    });
    created = true;
  }
  if (!session?.id) {
    throw new Error("Failed to acquire workspace session.");
  }

  if (options.queue) {
    await vkRequest(host, `/api/sessions/${session.id}/queue`, {
      method: "POST",
      body: {
        message,
        executor_profile_id: executorProfile || undefined,
      },
    });
    return { sessionId: session.id, created, action: "queued" };
  }

  await vkRequest(host, `/api/sessions/${session.id}/follow-up`, {
    method: "POST",
    body: {
      prompt: message,
      executor_profile_id: executorProfile || undefined,
    },
  });
  return { sessionId: session.id, created, action: "follow-up" };
}

async function cmdAgent(chatId, rawArgs) {
  const parsed = parseAgentArgs(rawArgs || "");
  const { message, workspaceId, role, model, queue, newSession, dryRun } =
    parsed;

  const { registry, errors, warnings } = await loadWorkspaceRegistry();
  const diagnostics = formatRegistryDiagnostics(errors, warnings);
  if (diagnostics) {
    await sendReply(chatId, diagnostics);
  }

  if (!message) {
    const list = registry.workspaces
      .map((ws) => `  - ${ws.id} (${ws.role})`)
      .join("\n");
    const usage = [
      "Usage: /agent --workspace <id> <task>",
      "       /agent --role <role> <task>",
      "Options: --model <name> --queue --new-session --dry-run",
      "",
      "Available workspaces:",
      list || "  (none)",
    ];
    await sendReply(chatId, usage.join("\n"));
    return;
  }

  if (!registry.workspaces.length) {
    await sendReply(chatId, "No workspaces available to route.");
    return;
  }

  let candidates = registry.workspaces;
  let preferredId = null;

  if (workspaceId) {
    const match = registry.workspaces.find(
      (ws) => ws.id.toLowerCase() === workspaceId.toLowerCase(),
    );
    if (!match) {
      const ids = registry.workspaces.map((ws) => ws.id).join(", ");
      await sendReply(
        chatId,
        `Unknown workspace: ${workspaceId}\nAvailable: ${ids}`,
      );
      return;
    }
    candidates = [match];
    preferredId = match.id;
  } else if (role) {
    const roleLower = role.toLowerCase();
    candidates = registry.workspaces.filter(
      (ws) => ws.role && ws.role.toLowerCase() === roleLower,
    );
    if (candidates.length === 0) {
      await sendReply(chatId, `No workspaces found with role: ${role}`);
      return;
    }
  } else {
    const primary = registry.workspaces.filter(
      (ws) => (ws.role || "").toLowerCase() === "primary",
    );
    if (primary.length > 0) {
      candidates = primary;
    }
  }

  const leaseOwner =
    process.env.VE_WORKSPACE_OWNER ||
    process.env.USER ||
    process.env.USERNAME ||
    `telegram:${chatId}`;
  const leaseTtlSecRaw = Number(process.env.VE_WORKSPACE_LEASE_TTL_SEC || "");
  const leaseTtlMinRaw = Number(process.env.VE_WORKSPACE_LEASE_TTL_MIN || "");
  const leaseTtlMinutes = Number.isFinite(leaseTtlMinRaw)
    ? leaseTtlMinRaw
    : Number.isFinite(leaseTtlSecRaw)
      ? Math.ceil(leaseTtlSecRaw / 60)
      : null;

  let availabilityMap = new Map();
  try {
    const registry = await loadSharedRegistry();
    const sweep = await sweepSharedLeases({ registry, actor: leaseOwner });
    availabilityMap = getSharedAvailabilityMap(sweep.registry);
  } catch {
    availabilityMap = new Map();
  }

  if (availabilityMap.size > 0) {
    const filtered = candidates.filter((ws) => {
      const entry = availabilityMap.get(ws.id);
      if (!entry) return true;
      const state = String(entry.state || "available").toLowerCase();
      return state === "available";
    });
    if (filtered.length === 0) {
      await sendReply(
        chatId,
        "No available workspaces found (all leased or unavailable).",
      );
      return;
    }
    candidates = filtered;
  }

  const healthMap = await getWorkspaceHealth(candidates);
  const ranked = rankWorkspaceCandidates(candidates, healthMap, {
    preferredId,
  });
  const preferredMatch = preferredId
    ? candidates.find((ws) => ws.id === preferredId)
    : null;

  let selectedWorkspace = null;
  let selectedHealth = null;
  let leaseError = null;

  for (const candidate of ranked) {
    if (dryRun) {
      selectedWorkspace = candidate;
      selectedHealth = healthMap.get(candidate.id) || null;
      break;
    }
    try {
      const claimResult = await claimSharedWorkspace({
        workspaceId: candidate.id,
        owner: leaseOwner,
        ttlMinutes: leaseTtlMinutes,
        note: `telegram:${chatId}`,
        actor: leaseOwner,
      });
      if (claimResult?.error) {
        throw new Error(claimResult.error);
      }
      selectedWorkspace = candidate;
      selectedHealth = healthMap.get(candidate.id) || null;
      break;
    } catch (err) {
      leaseError = err;
    }
  }

  if (!selectedWorkspace) {
    await sendReply(
      chatId,
      leaseError?.message || "No available workspace found for routing.",
    );
    return;
  }

  const selection = {
    workspace: selectedWorkspace,
    health: selectedHealth,
    fallbackFrom:
      preferredMatch && preferredMatch.id !== selectedWorkspace.id
        ? preferredMatch
        : null,
  };

  const modelSelection = resolveModelSelection(selection.workspace, model);
  const selectedModel = modelSelection.model || model || "auto";

  const infoLines = [
    `Routing → ${selection.workspace.name} (${selection.workspace.id})`,
    `Role: ${selection.workspace.role || "n/a"}`,
    `Host: ${normalizeHost(selection.workspace.host) || "n/a"}`,
    `Model: ${selectedModel}`,
  ];
  if (selection.fallbackFrom) {
    infoLines.push(`Fallback: ${selection.fallbackFrom.id} unavailable`);
  }

  if (dryRun) {
    infoLines.push("Dry-run only. No message sent.");
    await sendReply(chatId, infoLines.join("\\n"));
    return;
  }

  try {
    const result = await dispatchAgentMessage(selection.workspace, message, {
      executorProfile: modelSelection.profile,
      queue,
      newSession,
    });
    infoLines.push(`Action: ${result.action}`);
    infoLines.push(
      `Session: ${result.sessionId}${result.created ? " (new)" : ""}`,
    );
    await sendReply(chatId, infoLines.join("\\n"));
  } catch (err) {
    await sendReply(
      chatId,
      `❌ /agent failed: ${err.message || err}\n${infoLines.join("\\n")}`,
    );
  }
}

// ── /background — run task silently or background active agent ────────────────

async function cmdBackground(chatId, args) {
  const task = (args || "").trim();
  if (task) {
    await sendReply(
      chatId,
      `🛰️ Background task queued: "${task.slice(0, 80)}${task.length > 80 ? "…" : ""}"`,
    );
    await handleFreeText(task, chatId, { background: true });
    return;
  }

  if (!activeAgentSession) {
    await sendReply(
      chatId,
      "No active agent. Usage:\n/background <task>\n(background current agent with /background)",
    );
    return;
  }

  activeAgentSession.background = true;
  activeAgentSession.suppressEdits = true;

  if (agentMessageId && agentChatId) {
    try {
      await deleteDirect(agentChatId, agentMessageId);
    } catch {
      /* best effort */
    }
  }
  agentMessageId = null;
  if (activeAgentSession) {
    activeAgentSession.messageId = null;
  }

  await sendReply(
    chatId,
    "🛰️ Background mode enabled for the active agent. I will post a final summary when it completes. Use /stop to cancel or /steer to adjust context.",
  );
}

// ── /stop — Stop Running Agent ───────────────────────────────────────────────

async function cmdStop(chatId) {
  if (!activeAgentSession) {
    await sendReply(chatId, "No agent is currently running.");
    return;
  }
  activeAgentSession.aborted = true;
  if (activeAgentSession.abortController) {
    try {
      activeAgentSession.abortController.abort("user_stop");
    } catch {
      /* best effort */
    }
  }
  if (activeAgentSession.actionLog) {
    activeAgentSession.actionLog.push({
      icon: "🛑",
      text: "Stop requested by user (will halt after current step)",
    });
    if (activeAgentSession.scheduleEdit) {
      activeAgentSession.scheduleEdit();
    }
  }
  await sendReply(chatId, "🛑 Stop signal sent. Agent will halt and wait.");
}

// ── /steer — Steering update for running agent ───────────────────────────────

async function cmdSteer(chatId, steerArgs) {
  if (!steerArgs || !steerArgs.trim()) {
    await sendReply(chatId, "Usage: /steer <update or correction>");
    return;
  }
  const message = steerArgs.trim();

  if (!activeAgentSession || !isPrimaryBusy()) {
    await sendReply(chatId, "No active agent. Sending as a new task.");
    await handleFreeText(message, chatId);
    return;
  }

  const result = await steerPrimaryPrompt(message);
  if (result.ok) {
    if (activeAgentSession.actionLog) {
      activeAgentSession.actionLog.push({
        icon: "🧭",
        text: `Steering update delivered (${result.mode})`,
      });
      if (activeAgentSession.scheduleEdit) {
        activeAgentSession.scheduleEdit();
      }
    }
    await sendReply(chatId, `🧭 Steering sent (${result.mode}).`);
    return;
  }

  if (!activeAgentSession.followUpQueue) {
    activeAgentSession.followUpQueue = [];
  }
  activeAgentSession.followUpQueue.push(message);
  const qLen = activeAgentSession.followUpQueue.length;
  if (activeAgentSession.actionLog) {
    const steerStatus = result.reason || "failed";
    activeAgentSession.actionLog.push({
      icon: "🧭",
      text: `Steering queued (#${qLen}; steer failed: ${steerStatus})`,
      kind: "followup_queued",
      steerStatus,
    });
    if (activeAgentSession.scheduleEdit) {
      activeAgentSession.scheduleEdit();
    }
  }
  await sendReply(chatId, `🧭 Steering queued (#${qLen}).`);
}

// ── Free-text → Primary Agent Dispatch ───────────────────────────────────────

/**
 * Build the rolling summary message text from accumulated action log.
 * This is the single message that gets continuously edited in Telegram.
 */
function suppressSteerFailedLines(actionLog) {
  if (!Array.isArray(actionLog)) return;
  for (let i = actionLog.length - 1; i >= 0; i -= 1) {
    const entry = actionLog[i];
    if (!entry || entry.kind !== "followup_queued") continue;
    if (entry.steerStatus && entry.steerStatus !== "ok") {
      actionLog.splice(i, 1);
    }
  }
}

function buildStreamMessage({
  taskPreview,
  actionLog,
  currentThought,
  totalActions,
  phase,
  finalResponse,
  filesRead,
  filesWritten,
  searchesDone,
  statusIcon,
}) {
  const header = `🔧 Agent: ${taskPreview}`;
  const counter = `📊 Actions: ${totalActions} | ${phase}`;
  const separator = "────────────────────────────";

  // Show last N actions (keep message compact)
  const MAX_VISIBLE_ACTIONS = 20;
  const visibleActions = actionLog.slice(-MAX_VISIBLE_ACTIONS);
  const hiddenCount = actionLog.length - visibleActions.length;

  const lines = [header, counter, separator];

  if (hiddenCount > 0) {
    lines.push(`… ${hiddenCount} earlier action${hiddenCount > 1 ? "s" : ""}`);
  }

  for (const action of visibleActions) {
    lines.push(`${action.icon} ${action.text}`);
  }

  if (currentThought) {
    lines.push("", `💭 ${currentThought}`);
  }

  if (!finalResponse) {
    if (filesWritten?.size) {
      lines.push("", "✍️ Files modified so far:");
      const recent = Array.from(filesWritten.entries()).slice(-6);
      for (const [fpath, info] of recent) {
        const name = shortPath(fpath);
        if (info.adds || info.dels) {
          lines.push(`  ✏️ ${name} (+${info.adds} -${info.dels})`);
        } else {
          lines.push(`  ✏️ ${name}`);
        }
      }
    }
    if (filesRead?.size) {
      lines.push("", "📖 Files read so far:");
      const recent = Array.from(filesRead.values()).slice(-6);
      for (const fpath of recent) {
        lines.push(`  📄 ${shortPath(fpath)}`);
      }
    }
    if (searchesDone) {
      lines.push("", `🔎 Searches: ${searchesDone}`);
    }
  }

  if (finalResponse) {
    // ── Final summary block ──────────────────────────────────────
    const icon = statusIcon || "✅";
    lines.push("", separator);
    lines.push(`${icon} ${phase}`);
    lines.push("");

    // Stats line
    const stats = [];
    if (filesRead?.size) stats.push(`${filesRead.size} files read`);
    if (filesWritten?.size) stats.push(`${filesWritten.size} files modified`);
    if (searchesDone) stats.push(`${searchesDone} searches`);
    if (stats.length) {
      lines.push(`📈 ${stats.join(" · ")}`);
    }

    // Files modified detail
    if (filesWritten?.size) {
      lines.push("");
      lines.push("📁 Files modified:");
      for (const [fpath, info] of filesWritten) {
        const name = shortPath(fpath);
        if (info.adds || info.dels) {
          lines.push(`  ✏️ ${name} (+${info.adds} -${info.dels})`);
        } else {
          const kindIcon =
            info.kind === "add" ? "➕" : info.kind === "delete" ? "🗑️" : "✏️";
          lines.push(`  ${kindIcon} ${name}`);
        }
      }
    }

    lines.push("");
    lines.push(finalResponse.slice(0, 1200));
  }

  return lines.join("\n");
}

async function handleFreeText(text, chatId, options = {}) {
  const backgroundMode = !!options.background;
  // ── Follow-up steering: if agent is busy, queue message as follow-up ──
  if (isPrimaryBusy() && activeAgentSession) {
    if (!activeAgentSession.followUpQueue) {
      activeAgentSession.followUpQueue = [];
    }
    activeAgentSession.followUpQueue.push(text);
    const qLen = activeAgentSession.followUpQueue.length;

    // Try immediate steering so the in-flight run can adapt ASAP.
    const steerResult = await steerPrimaryPrompt(text);
    const steerStatus = steerResult.ok ? "ok" : steerResult.reason || "failed";
    const steerNote = steerResult.ok
      ? `Steer ${steerResult.mode}.`
      : `Steer failed (${steerStatus}).`;

    // Acknowledge the follow-up in both the user's chat and update the agent message
    await sendDirect(
      chatId,
      `📌 Follow-up queued (#${qLen}). Agent will process it after current action. ${steerNote}`,
    );

    // Add follow-up indicator to the streaming message
    if (activeAgentSession.actionLog) {
      activeAgentSession.actionLog.push({
        icon: "📌",
        text: `Follow-up: "${text.length > 60 ? text.slice(0, 60) + "…" : text}" (${steerNote})`,
        kind: "followup_queued",
        steerStatus,
      });
      // Trigger an edit to show the follow-up in the streaming message
      if (activeAgentSession.scheduleEdit) {
        activeAgentSession.scheduleEdit();
      }
    }
    return;
  }

  // ── Block if agent is busy but no session (shouldn't happen normally) ──
  if (isPrimaryBusy()) {
    await sendReply(
      chatId,
      "⏳ Agent is executing a task. Please wait for it to finish...",
    );
    return;
  }

  const taskPreview = text.length > 60 ? text.slice(0, 60) + "…" : text;

  // Send the initial message and capture its ID for editing (unless background)
  let messageId = null;
  if (!backgroundMode) {
    messageId = await sendDirect(
      chatId,
      buildStreamMessage({
        taskPreview,
        actionLog: [],
        currentThought: null,
        totalActions: 0,
        phase: "starting…",
        finalResponse: null,
      }),
    );
  }

  // Load current status for context
  let statusData = null;
  try {
    if (_readStatusData) {
      statusData = await _readStatusData();
    } else {
      const raw = await readFile(statusPath, "utf8").catch(() => null);
      statusData = raw ? JSON.parse(raw) : null;
    }
  } catch {
    /* best effort */
  }

  // ── Single-message streaming state ──────────────────────────────────
  const actionLog = []; // { icon, text } entries
  let currentThought = null;
  let totalActions = 0;
  let phase = "working…";
  let lastEditAt = 0;
  const EDIT_THROTTLE_MS = 2000; // edit at most every 2s (Telegram rate limit)
  let editPending = false;
  let editTimer = null;

  // ── Tracking for final summary ──────────────────────────────────────
  const filesRead = new Set(); // file paths read
  const filesWritten = new Map(); // path → { kind, adds, dels }
  let searchCount = 0;
  let hadError = false;

  const doEdit = async () => {
    if (backgroundMode || activeAgentSession?.background) return;
    editPending = false;
    const msg = buildStreamMessage({
      taskPreview,
      actionLog,
      currentThought,
      totalActions,
      phase,
      finalResponse: null,
      filesRead,
      filesWritten,
      searchesDone: searchCount,
    });
    if (messageId) {
      messageId = await editDirect(chatId, messageId, msg);
      agentMessageId = messageId;
    }
    lastEditAt = Date.now();
  };

  const scheduleEdit = () => {
    if (backgroundMode || activeAgentSession?.background) return;
    if (editPending) return;
    const now = Date.now();
    const elapsed = now - lastEditAt;
    if (elapsed >= EDIT_THROTTLE_MS) {
      editPending = true;
      void doEdit();
    } else {
      editPending = true;
      if (editTimer) clearTimeout(editTimer);
      editTimer = setTimeout(() => void doEdit(), EDIT_THROTTLE_MS - elapsed);
    }
  };

  // ── Set up agent session (enables follow-up steering & bottom-pinning) ──
  const abortController = new AbortController();
  activeAgentSession = {
    chatId,
    messageId,
    taskPreview,
    actionLog,
    currentThought: null,
    totalActions: 0,
    phase: "working…",
    followUpQueue: [],
    scheduleEdit,
    aborted: false,
    abortController,
    background: backgroundMode,
    suppressEdits: backgroundMode,
  };
  agentMessageId = messageId;
  agentChatId = chatId;

  const onEvent = async (_formatted, rawEvent) => {
    const action = rawEvent ? summarizeAction(rawEvent) : null;
    if (!action) return;

    // ── Track files read & written for final summary ──────────────
    if (rawEvent.type === "item.completed") {
      const item = rawEvent.item;
      if (item.type === "command_execution" && item.command) {
        const target = extractTarget(item.command);
        if (target) {
          // Determine if this is a read or search command
          if (
            /^(cat|head|tail|type|Get-Content)/i.test(item.command.trim()) ||
            /pwsh.*Get-Content/i.test(item.command)
          ) {
            filesRead.add(target);
          }
          if (
            /^(grep|findstr|rg|Select-String)/i.test(item.command.trim()) ||
            /pwsh.*Select-String/i.test(item.command)
          ) {
            searchCount++;
          }
        }
      }
      if (item.type === "file_change" && item.changes?.length) {
        for (const c of item.changes) {
          filesWritten.set(c.path, {
            kind: c.kind || "modify",
            adds: c.additions ?? c.lines_added ?? 0,
            dels: c.deletions ?? c.lines_deleted ?? 0,
          });
        }
      }
    }

    if (
      rawEvent.type === "tool.execution_start" ||
      rawEvent.type === "tool.execution_complete"
    ) {
      const { toolName, input } = getCopilotToolInfo(rawEvent);
      const command = extractCopilotCommand(input);
      const target = extractCopilotPath(input);

      if (command) {
        const cmdTarget = extractTarget(command);
        if (
          cmdTarget &&
          (/^(cat|head|tail|type|Get-Content)/i.test(command.trim()) ||
            /pwsh.*Get-Content/i.test(command))
        ) {
          filesRead.add(cmdTarget);
        }
        if (
          /^(grep|findstr|rg|Select-String)/i.test(command.trim()) ||
          /pwsh.*Select-String/i.test(command)
        ) {
          searchCount++;
        }
      }

      if (isCopilotReadTool(toolName) && target) {
        filesRead.add(target);
      }
      if (isCopilotSearchTool(toolName)) {
        searchCount++;
      }
      if (isCopilotWriteTool(toolName) && target) {
        filesWritten.set(target, {
          kind: "modify",
          adds: 0,
          dels: 0,
        });
      }
    }

    // ── Track file changes from action detail ─────────────────────
    if (action.detail === "file_change" && action.files) {
      for (const f of action.files) {
        filesWritten.set(f.path, {
          kind: f.kind || "modify",
          adds: f.adds || 0,
          dels: f.dels || 0,
        });
      }
    }

    if (action.phase === "thinking") {
      currentThought = action.text;
      if (activeAgentSession) activeAgentSession.currentThought = action.text;
    } else {
      if (action.phase === "done" || action.phase === "running") {
        totalActions++;
        if (activeAgentSession) activeAgentSession.totalActions = totalActions;
      }
      actionLog.push(action);
      // Keep thought visible while actions proceed (only clear on new non-thinking action)
      if (action.phase !== "thinking") {
        currentThought = null;
        if (activeAgentSession) activeAgentSession.currentThought = null;
      }
    }

    if (action.phase === "error") {
      phase = "error";
      hadError = true;
    } else if (action.phase === "planning") {
      phase = "planning…";
    } else {
      phase = "working…";
    }
    if (activeAgentSession) activeAgentSession.phase = phase;

    scheduleEdit();
  };

  try {
    const result = await execPrimaryPrompt(text, {
      statusData,
      timeoutMs: AGENT_TIMEOUT_MS,
      onEvent,
      sendRawEvents: true, // request raw events alongside formatted ones
      abortController,
    });

    if (editTimer) clearTimeout(editTimer);

    // ── Process follow-up queue ───────────────────────────────────
    // If user sent follow-up messages while agent was working, process them now
    const followUps = activeAgentSession?.followUpQueue || [];
    if (followUps.length > 0 && !activeAgentSession?.aborted) {
      for (const followUp of followUps) {
        actionLog.push({
          icon: "📌",
          text: `Processing follow-up: "${followUp.slice(0, 60)}"`,
        });
        phase = "processing follow-up…";
        scheduleEdit();

        try {
          const followUpResult = await execPrimaryPrompt(followUp, {
            statusData,
            timeoutMs: AGENT_TIMEOUT_MS,
            onEvent,
            sendRawEvents: true,
          });

          // Merge follow-up results
          if (followUpResult.finalResponse) {
            result.finalResponse =
              (result.finalResponse || "") +
              `\n\n📌 Follow-up result:\n${followUpResult.finalResponse}`;
            suppressSteerFailedLines(actionLog);
          }
        } catch (err) {
          actionLog.push({
            icon: "❌",
            text: `Follow-up error: ${err.message}`,
          });
        }
      }
    }

    // Final edit with the complete summary
    const itemSummary = result.items.filter(
      (i) =>
        i.type === "command_execution" ||
        i.type === "file_change" ||
        i.type === "mcp_tool_call",
    ).length;

    totalActions = Math.max(totalActions, itemSummary);

    // Determine final status icon
    const hasChanges = filesWritten.size > 0;
    let statusIcon;
    if (hadError) {
      statusIcon = "❌";
      phase = "Failed — needs manual review";
    } else if (hasChanges) {
      statusIcon = "✅";
      phase = "Completed successfully";
    } else {
      // No files changed — might be informational or might need user input
      statusIcon = "❓";
      phase = "Completed — no files changed";
    }

    const finalMsg = buildStreamMessage({
      taskPreview,
      actionLog,
      currentThought: null,
      totalActions,
      phase,
      finalResponse: result.finalResponse || null,
      filesRead,
      filesWritten,
      searchesDone: searchCount,
      statusIcon,
    });
    if (backgroundMode || activeAgentSession?.background) {
      await sendReply(chatId, finalMsg);
    } else {
      await editDirect(chatId, messageId, finalMsg);
    }
  } catch (err) {
    if (editTimer) clearTimeout(editTimer);
    const finalMsg = buildStreamMessage({
      taskPreview,
      actionLog,
      currentThought: null,
      totalActions,
      phase: "Failed — error during execution",
      finalResponse: `Error: ${err.message}`,
      filesRead,
      filesWritten,
      searchesDone: searchCount,
      statusIcon: "❌",
    });
    if (backgroundMode || activeAgentSession?.background) {
      await sendReply(chatId, finalMsg);
    } else {
      await editDirect(chatId, messageId, finalMsg);
    }
  } finally {
    // ── Clean up agent session ────────────────────────────────────
    activeAgentSession = null;
    agentMessageId = null;
    agentChatId = null;
  }
}

// ── Main Polling Loop ────────────────────────────────────────────────────────

async function pollLoop() {
  while (polling) {
    try {
      const updates = await pollUpdates();
      for (const update of updates) {
        lastUpdateId = Math.max(lastUpdateId, update.update_id);
        try {
          await handleUpdate(update);
        } catch (err) {
          console.error(
            `[telegram-bot] error handling update ${update.update_id}: ${err.message}`,
          );
        }
      }
    } catch (err) {
      console.error(`[telegram-bot] poll loop error: ${err.message}`);
      // Backoff before retrying
      await new Promise((r) => setTimeout(r, POLL_ERROR_BACKOFF_MS));
    }
  }
}

async function ensurePresenceReady() {
  if (presenceReady) return;
  const localWorkspace = await getLocalWorkspaceContext();
  await initPresence({
    repoRoot: _getRepoRoot ? _getRepoRoot() : repoRoot,
    localWorkspace,
  });
  presenceReady = true;
}

function startPresenceLoop() {
  if (presenceDisabled) return;
  if (!telegramToken || !presenceChatId) return;
  if (!Number.isFinite(presenceIntervalSec) || presenceIntervalSec <= 0) {
    return;
  }
  const intervalMs = presenceIntervalSec * 1000;
  let lastSentPayload = null;

  const sendPresence = async () => {
    try {
      await ensurePresenceReady();
      const child = _getCurrentChild ? _getCurrentChild() : null;
      const payload = buildLocalPresence({
        orchestrator_running: !!child,
        orchestrator_pid: child?.pid ?? null,
        vk_url: _getVibeKanbanUrl ? _getVibeKanbanUrl() : null,
      });

      // Check if state changed significantly (ignore updated_at for comparison)
      const shouldSend =
        !presenceOnlyOnChange ||
        !lastSentPayload ||
        hasPresenceChanged(lastSentPayload, payload);

      // Always update local registry
      await notePresence(payload, {
        source: "local",
        receivedAt: payload.updated_at,
      });

      // Only send to Telegram if state changed or not configured to only-on-change
      if (shouldSend) {
        await sendDirect(presenceChatId, formatPresenceMessage(payload), {
          silent: presenceSilent,
        });
        lastSentPayload = payload;
      }
    } catch (err) {
      console.warn(
        `[telegram-bot] presence heartbeat error: ${err.message || err}`,
      );
    }
  };
  setTimeout(() => void sendPresence(), intervalMs);
  setInterval(() => void sendPresence(), intervalMs);
}

function hasPresenceChanged(prev, curr) {
  if (!prev || !curr) return true;
  // Compare meaningful fields (ignore timestamps)
  const significantFields = [
    "instance_id",
    "workspace_id",
    "workspace_role",
    "orchestrator_running",
    "orchestrator_pid",
    "git_branch",
    "git_sha",
    "coordinator_priority",
    "coordinator_eligible",
  ];
  for (const field of significantFields) {
    if (prev[field] !== curr[field]) {
      return true;
    }
  }
  return false;
}

// ── Notification Batching System ─────────────────────────────────────────────

const messageQueue = {
  critical: [], // priority 1 - immediate
  errors: [], // priority 2
  warnings: [], // priority 3
  info: [], // priority 4
  debug: [], // priority 5
};

let batchFlushTimer = null;

// ── Live Digest State ───────────────────────────────────────────────────────
// A single Telegram message that gets continuously edited as events happen.
// When the digest window expires, the message is sealed and the next event
// starts a fresh one.
let liveDigest = {
  messageId: null, // Telegram message_id of the current live digest
  chatId: null, // chat_id it was sent to
  startedAt: 0, // timestamp when this digest window started
  entries: [], // { emoji, text, time } — events in this digest window
  sealTimer: null, // timer to seal the digest after the window expires
  editTimer: null, // debounce timer for edits
  editPending: false, // whether an edit is pending
  sealed: false, // true once the window has expired and message is finalized
};

const PRIORITY_EMOJI = {
  1: "🔴",
  2: "❌",
  3: "⚠️",
  4: "ℹ️",
  5: "🔹",
};

/**
 * Build the live digest message text from accumulated entries.
 */
function buildLiveDigestText() {
  const d = liveDigest;
  const startTime = new Date(d.startedAt).toISOString().slice(11, 19);
  const now = new Date().toISOString().slice(11, 19);

  // Count by severity
  const counts = { 1: 0, 2: 0, 3: 0, 4: 0, 5: 0 };
  for (const e of d.entries) {
    counts[e.priority] = (counts[e.priority] || 0) + 1;
  }

  const countParts = [];
  if (counts[1] > 0) countParts.push(`🔴 ${counts[1]}`);
  if (counts[2] > 0) countParts.push(`❌ ${counts[2]}`);
  if (counts[3] > 0) countParts.push(`⚠️ ${counts[3]}`);
  if (counts[4] > 0) countParts.push(`ℹ️ ${counts[4]}`);

  const statusLine = d.sealed
    ? `📊 Digest (${startTime} → ${now}) — sealed`
    : `📊 Live Digest (since ${startTime}) — updating...`;
  const headerLine =
    countParts.length > 0
      ? `${statusLine}\n${countParts.join(" • ")}`
      : statusLine;

  // Build event lines (most recent at bottom, like a log)
  // Telegram 4096 char limit — keep recent events, trim old ones
  const MAX_LEN = 3800; // leave room for header
  const lines = [];
  let totalLen = headerLine.length + 2; // +2 for \n\n separator

  // Add entries from newest to oldest, then reverse for chronological order
  for (let i = d.entries.length - 1; i >= 0; i--) {
    const e = d.entries[i];
    const line = `${e.time} ${e.emoji} ${e.text}`;
    if (totalLen + line.length + 1 > MAX_LEN) {
      const trimmed = d.entries.length - lines.length;
      if (trimmed > 0) {
        lines.push(`  …${trimmed} earlier event(s) trimmed`);
      }
      break;
    }
    lines.push(line);
    totalLen += line.length + 1;
  }

  lines.reverse(); // chronological order

  return [headerLine, "", ...lines].join("\n");
}

/**
 * Schedule a debounced edit of the live digest message.
 */
function scheduleLiveDigestEdit() {
  const d = liveDigest;
  if (d.editTimer) {
    clearTimeout(d.editTimer);
  }
  d.editPending = true;
  d.editTimer = setTimeout(async () => {
    d.editPending = false;
    d.editTimer = null;
    if (!d.messageId || !d.chatId) return;
    const text = buildLiveDigestText();
    try {
      const newId = await editDirect(d.chatId, d.messageId, text);
      if (newId && newId !== d.messageId) {
        d.messageId = newId; // editDirect fell back to sendDirect
      }
    } catch (err) {
      console.warn(`[telegram-bot] live digest edit failed: ${err.message}`);
    }
  }, liveDigestEditDebounceMs);
}

/**
 * Seal the current live digest window — mark final, clear state for next window.
 */
function sealLiveDigest() {
  const d = liveDigest;
  if (d.entries.length === 0) {
    // Nothing happened in this window — just reset
    resetLiveDigest();
    return;
  }

  // Snapshot entries before sealing (for devmode auto code fix callback)
  const sealedEntries = [...d.entries];
  const sealedText = buildLiveDigestText();

  d.sealed = true;
  // Flush one last edit to mark it sealed
  if (d.editTimer) clearTimeout(d.editTimer);
  const text = buildLiveDigestText();
  if (d.messageId && d.chatId) {
    editDirect(d.chatId, d.messageId, text).catch(() => {});
  }

  // Fire digest sealed callback (used by devmode auto code fix)
  if (_onDigestSealed) {
    try {
      _onDigestSealed({ entries: sealedEntries, text: sealedText });
    } catch (err) {
      console.warn(
        `[telegram-bot] onDigestSealed callback error: ${err.message}`,
      );
    }
  }

  // Reset for next window
  resetLiveDigest();
}

/**
 * Reset live digest state for a new window.
 */
function resetLiveDigest() {
  if (liveDigest.sealTimer) clearTimeout(liveDigest.sealTimer);
  if (liveDigest.editTimer) clearTimeout(liveDigest.editTimer);
  liveDigest = {
    messageId: null,
    chatId: null,
    startedAt: 0,
    entries: [],
    sealTimer: null,
    editTimer: null,
    editPending: false,
    sealed: false,
  };
  // Clear persisted state
  writeFile(liveDigestStatePath, "{}").catch(() => {});
}

/**
 * Persist live digest state to disk so restarts can resume the same message.
 */
function persistLiveDigest() {
  const d = liveDigest;
  if (!d.messageId) return;
  const state = {
    messageId: d.messageId,
    chatId: d.chatId,
    startedAt: d.startedAt,
    entries: d.entries,
  };
  writeFile(liveDigestStatePath, JSON.stringify(state)).catch(() => {});
}

/**
 * Restore live digest state from disk after a restart.
 * Returns true if a valid digest was restored (still within window).
 */
export async function restoreLiveDigest() {
  // Already restored or active — skip
  if (liveDigest.messageId) return true;
  try {
    const raw = await readFile(liveDigestStatePath, "utf8");
    const state = JSON.parse(raw);
    if (!state.messageId || !state.startedAt) return false;
    const now = Date.now();
    const windowMs = liveDigestWindowSec * 1000;
    // Only restore if the window hasn't expired
    if (now - state.startedAt >= windowMs) {
      await writeFile(liveDigestStatePath, "{}").catch(() => {});
      return false;
    }
    liveDigest.messageId = state.messageId;
    liveDigest.chatId = state.chatId || telegramChatId;
    liveDigest.startedAt = state.startedAt;
    liveDigest.entries = state.entries || [];
    liveDigest.sealed = false;
    // Re-schedule the seal timer for remaining window time
    const remaining = windowMs - (now - state.startedAt);
    liveDigest.sealTimer = setTimeout(() => sealLiveDigest(), remaining);
    console.log(
      `[telegram-bot] restored live digest (${liveDigest.entries.length} entries, ${Math.round(remaining / 1000)}s remaining)`,
    );
    return true;
  } catch {
    return false;
  }
}

/**
 * Add an event to the live digest. Creates the message on first event,
 * then edits it for subsequent events in the same window.
 */
async function addToLiveDigest(text, priority, category) {
  const d = liveDigest;
  const now = Date.now();
  const timeStr = new Date(now).toISOString().slice(11, 19);
  const emoji = PRIORITY_EMOJI[priority] || "ℹ️";

  // Check if we need a new digest window
  const windowMs = liveDigestWindowSec * 1000;
  const windowExpired = d.startedAt > 0 && now - d.startedAt >= windowMs;

  if (windowExpired || !d.startedAt) {
    // Seal old digest if it had entries
    if (d.entries.length > 0) {
      sealLiveDigest();
    } else {
      resetLiveDigest();
    }
  }

  // Add entry
  liveDigest.entries.push({ emoji, text, time: timeStr, priority, category });

  if (!liveDigest.startedAt) {
    // First event — create the digest message
    liveDigest.startedAt = now;
    liveDigest.chatId = telegramChatId;

    const messageText = buildLiveDigestText();
    const msgId = await sendDirect(telegramChatId, messageText, {
      silent: true,
    });
    if (msgId) {
      liveDigest.messageId = msgId;
    }
    persistLiveDigest();

    // Schedule seal
    liveDigest.sealTimer = setTimeout(() => sealLiveDigest(), windowMs);
  } else {
    // Subsequent event — debounced edit
    scheduleLiveDigestEdit();
    persistLiveDigest();
  }
}

/**
 * Queue a notification for batched delivery.
 * Routes through Live Digest when enabled, falls back to batch queues.
 * @param {string} text - Message text
 * @param {number} priority - 1=critical(immediate), 2=error, 3=warning, 4=info, 5=debug
 * @param {object} options - { category, data, silent }
 */
function queueNotification(text, priority = 4, options = {}) {
  // Critical messages always go immediately
  if (priority <= immediateThreshold) {
    return sendDirect(telegramChatId, text, { silent: options.silent });
  }

  // Live Digest mode: append to the continuously-edited message
  if (liveDigestEnabled && batchingEnabled) {
    return addToLiveDigest(text, priority, options.category || "general");
  }

  // Legacy batch mode fallback
  if (!batchingEnabled) {
    return sendDirect(telegramChatId, text, { silent: options.silent });
  }

  const category = options.category || "info";
  const entry = {
    text,
    priority,
    category,
    timestamp: new Date().toISOString(),
    data: options.data || {},
  };

  // Route to appropriate queue
  if (priority === 1) {
    messageQueue.critical.push(entry);
  } else if (priority === 2) {
    messageQueue.errors.push(entry);
  } else if (priority === 3) {
    messageQueue.warnings.push(entry);
  } else if (priority === 5) {
    messageQueue.debug.push(entry);
  } else {
    messageQueue.info.push(entry);
  }

  // Flush if queue is getting too large
  const totalSize =
    messageQueue.critical.length +
    messageQueue.errors.length +
    messageQueue.warnings.length +
    messageQueue.info.length +
    messageQueue.debug.length;

  if (totalSize >= batchMaxSize) {
    flushNotificationQueue();
  }
}

/**
 * Format and send batched notifications as a summary (legacy fallback).
 */
async function flushNotificationQueue() {
  const sections = [];
  const counts = {
    critical: messageQueue.critical.length,
    errors: messageQueue.errors.length,
    warnings: messageQueue.warnings.length,
    info: messageQueue.info.length,
    debug: messageQueue.debug.length,
  };

  const totalMessages =
    counts.critical +
    counts.errors +
    counts.warnings +
    counts.info +
    counts.debug;

  if (totalMessages === 0) return; // Nothing to send

  // Build summary header
  const timestamp = new Date().toISOString().slice(11, 19);
  let header = `📊 Update Summary (${timestamp})`;
  if (totalMessages > 0) {
    const parts = [];
    if (counts.critical > 0) parts.push(`🔴 ${counts.critical}`);
    if (counts.errors > 0) parts.push(`❌ ${counts.errors}`);
    if (counts.warnings > 0) parts.push(`⚠️ ${counts.warnings}`);
    if (counts.info > 0) parts.push(`ℹ️ ${counts.info}`);
    header += `\n${parts.join(" • ")}`;
  }

  // Critical messages (show all)
  if (counts.critical > 0) {
    sections.push(
      `🔴 Critical:\n${messageQueue.critical.map((m) => `  • ${m.text}`).join("\n")}`,
    );
  }

  // Errors (show up to 5, then summarize)
  if (counts.errors > 0) {
    const errorTexts = messageQueue.errors
      .slice(0, 5)
      .map((m) => `  • ${m.text}`);
    if (counts.errors > 5) {
      errorTexts.push(`  • ... and ${counts.errors - 5} more errors`);
    }
    sections.push(`❌ Errors:\n${errorTexts.join("\n")}`);
  }

  // Warnings (show up to 3, then summarize)
  if (counts.warnings > 0) {
    const warnTexts = messageQueue.warnings
      .slice(0, 3)
      .map((m) => `  • ${m.text}`);
    if (counts.warnings > 3) {
      warnTexts.push(`  • ... and ${counts.warnings - 3} more warnings`);
    }
    sections.push(`⚠️ Warnings:\n${warnTexts.join("\n")}`);
  }

  // Info messages (aggregate by category)
  if (counts.info > 0) {
    const categories = {};
    for (const msg of messageQueue.info) {
      const cat = msg.category || "general";
      categories[cat] = (categories[cat] || 0) + 1;
    }
    const summary = Object.entries(categories)
      .map(([cat, count]) => `  • ${cat}: ${count}`)
      .join("\n");
    sections.push(`ℹ️ Info:\n${summary}`);
  }

  // Build final message
  const message = [header, ...sections].join("\n\n");

  // Send the summary
  await sendDirect(telegramChatId, message, { silent: true });

  // Clear queues
  messageQueue.critical.length = 0;
  messageQueue.errors.length = 0;
  messageQueue.warnings.length = 0;
  messageQueue.info.length = 0;
  messageQueue.debug.length = 0;
}

/**
 * Start periodic flushing of the notification queue.
 * In live-digest mode, the flush loop is only used as a fallback seal timer.
 */
async function startBatchFlushLoop() {
  if (!batchingEnabled || batchFlushTimer) return;
  // In live digest mode, restore persisted state (if not already restored) and skip flush loop
  if (liveDigestEnabled) {
    if (!liveDigest.messageId) await restoreLiveDigest();
    return;
  }
  const intervalMs = batchIntervalSec * 1000;
  batchFlushTimer = setInterval(() => {
    flushNotificationQueue().catch((err) =>
      console.warn(`[telegram-bot] batch flush error: ${err.message}`),
    );
  }, intervalMs);
}

/**
 * Stop the batch flush loop and seal any active live digest.
 */
function stopBatchFlushLoop() {
  if (batchFlushTimer) {
    clearInterval(batchFlushTimer);
    batchFlushTimer = null;
  }
  // Seal any active live digest
  if (liveDigest.entries.length > 0) {
    sealLiveDigest();
  } else {
    resetLiveDigest();
  }
}

// ── Public API ───────────────────────────────────────────────────────────────

/**
 * Start the two-way Telegram bot.
 * Call injectMonitorFunctions() first if you want full integration.
 */
export async function startTelegramBot() {
  if (!telegramToken || !telegramChatId) {
    console.warn(
      "[telegram-bot] disabled (missing TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID)",
    );
    return;
  }

  const lockOk = await acquireTelegramPollLock("telegram-bot");
  if (!lockOk) {
    console.warn(
      "[telegram-bot] polling disabled (another getUpdates poller is active)",
    );
    return;
  }

  // Initialize the primary agent context
  await initPrimaryAgent();

  // Register bot commands with Telegram (updates the / menu)
  await registerBotCommands();

  // Start presence announcements for multi-workstation discovery
  startPresenceLoop();

  // Start batched notification / live digest system
  await startBatchFlushLoop();

  // Clear any pending updates that arrived while we were offline
  try {
    const stale = await pollUpdates();
    for (const u of stale) {
      lastUpdateId = Math.max(lastUpdateId, u.update_id);
    }
    if (stale.length > 0) {
      console.log(`[telegram-bot] skipped ${stale.length} stale updates`);
    }
  } catch {
    /* best effort */
  }

  polling = true;

  // Only send "online" notification on truly fresh starts, not code-change restarts.
  // Check the self-restart marker file first, then fall back to rapid-restart heuristic.
  const botStartPath = resolve(repoRoot, ".cache", "ve-last-bot-start.txt");
  const selfRestartPath = resolve(repoRoot, ".cache", "ve-self-restart.marker");
  let suppressOnline = false;
  try {
    if (existsSync(selfRestartPath)) {
      const ts = Number(readFileSync(selfRestartPath, "utf8"));
      if (Date.now() - ts < 30_000) suppressOnline = true;
    }
  } catch {
    /* best effort */
  }
  if (!suppressOnline) {
    try {
      const prev = await readFile(botStartPath, "utf8");
      const elapsed = Date.now() - Number(prev);
      if (elapsed < 60_000) suppressOnline = true;
    } catch {
      /* first start or missing file */
    }
  }
  await writeFile(botStartPath, String(Date.now())).catch(() => {});

  if (suppressOnline) {
    console.log(
      "[telegram-bot] restarted (suppressed online notification — rapid restart)",
    );
  } else {
    await sendDirect(
      telegramChatId,
      `🤖 VirtEngine primary agent online (${getPrimaryAgentName()}).\n\nType /help for commands or send any message to chat with the agent.`,
    );
  }

  console.log("[telegram-bot] started — listening for messages");

  // Start the polling loop (non-blocking)
  pollLoop().catch((err) => {
    console.error(`[telegram-bot] fatal poll loop error: ${err.message}`);
    polling = false;
  });
}

/**
 * Stop the Telegram bot polling.
 */
export function stopTelegramBot(options = {}) {
  polling = false;
  if (pollAbort) {
    try {
      pollAbort.abort();
    } catch {
      /* best effort */
    }
  }
  if (options.preserveDigest) {
    // Self-restart: persist live digest state for the next process to resume.
    // Don't seal or reset — the new process will pick up where we left off.
    persistLiveDigest();
    if (liveDigest.sealTimer) clearTimeout(liveDigest.sealTimer);
    if (liveDigest.editTimer) clearTimeout(liveDigest.editTimer);
    if (batchFlushTimer) {
      clearInterval(batchFlushTimer);
      batchFlushTimer = null;
    }
  } else {
    // Normal shutdown: seal any active live digest and flush legacy queues
    if (liveDigestEnabled && liveDigest.entries.length > 0) {
      sealLiveDigest();
    } else {
      flushNotificationQueue().catch(() => {});
    }
    stopBatchFlushLoop();
  }
  void releaseTelegramPollLock();
  console.log("[telegram-bot] stopped");
}

/**
 * Queue a notification for batched delivery (exported for monitor.mjs).
 * @param {string} text - Message text
 * @param {number} priority - 1=critical(immediate), 2=error, 3=warning, 4=info, 5=debug
 * @param {object} options - { category, data, silent }
 */
export function notify(text, priority = 4, options = {}) {
  return queueNotification(text, priority, options);
}

/**
 * Get a snapshot of the current live digest entries.
 * Useful for external consumers that need to read digest state.
 */
export function getDigestSnapshot() {
  return {
    entries: [...liveDigest.entries],
    startedAt: liveDigest.startedAt,
    sealed: liveDigest.sealed,
  };
}


// ── Periodic status file writer ─────────────────────────────────
// Called by monitor to keep the status file in sync with live executor state
let _statusWriterTimer = null;

export function startStatusFileWriter(intervalMs = 30000) {
  if (_statusWriterTimer) return;
  _statusWriterTimer = setInterval(async () => {
    try {
      const executor = _getInternalExecutor?.();
      if (!executor) return;
      const status = executor.getStatus?.();
      if (!status) return;

      let data = {};
      try {
        const raw = await readFile(statusPath, "utf8");
        data = JSON.parse(raw);
      } catch { /* fresh file */ }

      // Convert executor slots to the attempts format
      const attempts = {};
      for (const slot of status.slots) {
        attempts[slot.taskId] = {
          task_id: slot.taskId,
          task_title: slot.taskTitle,
          branch: slot.branch,
          status: slot.status,
          executor: slot.sdk,
          started_at: new Date(Date.now() - slot.runningFor * 1000).toISOString(),
          updated_at: new Date().toISOString(),
          attempt: slot.attempt,
        };
      }

      data.attempts = attempts;
      data.last_executor_sync = new Date().toISOString();
      data.executor_mode = status.mode || "unknown";
      data.active_slots = `${status.activeSlots}/${status.maxParallel}`;

      const { writeFile } = await import("node:fs/promises");
      await writeFile(statusPath, JSON.stringify(data, null, 2));
    } catch (err) {
      console.warn("[telegram-bot] Status file write error:", err.message);
    }
  }, intervalMs);

  if (_statusWriterTimer.unref) _statusWriterTimer.unref();
}

export function stopStatusFileWriter() {
  if (_statusWriterTimer) {
    clearInterval(_statusWriterTimer);
    _statusWriterTimer = null;
  }
}
