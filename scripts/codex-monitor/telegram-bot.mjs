/**
 * telegram-bot.mjs — Two-way Telegram ↔ Codex shell for VirtEngine monitor.
 *
 * Polls Telegram Bot API for incoming messages, routes slash commands to
 * built-in handlers, and forwards free-text to the persistent Codex shell.
 *
 * Architecture:
 *   Telegram → getUpdates long-poll → handleUpdate()
 *     ├─ /command → built-in handler (fast, no Codex)
 *     └─ free-text → CodexShell.exec() → response back to Telegram
 *
 * Security: Only accepts messages from the configured TELEGRAM_CHAT_ID.
 */

import { execSync, spawnSync } from "node:child_process";
import { existsSync } from "node:fs";
import { readFile, readdir, stat, unlink, writeFile } from "node:fs/promises";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";
import {
  execCodexPrompt,
  isCodexBusy,
  getThreadInfo,
  resetThread,
  initCodexShell,
  steerCodexPrompt,
} from "./codex-shell.mjs";
import {
  loadWorkspaceRegistry,
  formatRegistryDiagnostics,
  getDefaultModelPriority,
} from "./workspace-registry.mjs";

const __dirname = resolve(fileURLToPath(new URL(".", import.meta.url)));
const repoRoot = resolve(__dirname, "..", "..");
const statusPath = resolve(repoRoot, ".cache", "ve-orchestrator-status.json");
const telegramPollLockPath = resolve(repoRoot, ".cache", "telegram-getupdates.lock");

// ── Configuration ────────────────────────────────────────────────────────────

const telegramToken = process.env.TELEGRAM_BOT_TOKEN;
const telegramChatId = process.env.TELEGRAM_CHAT_ID;
const POLL_TIMEOUT_S = 30; // long-poll timeout
const MAX_MESSAGE_LEN = 4000; // Telegram max is 4096, leave margin
const POLL_ERROR_BACKOFF_MS = 5000;
const CODEX_TIMEOUT_MS = 15 * 60 * 1000; // 15 min for agentic tasks
let telegramPollLockHeld = false;

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
        const data = JSON.parse(raw);
        const pid = Number(data?.pid);
        if (!canSignalProcess(pid)) {
          await unlink(telegramPollLockPath);
          return await acquireTelegramPollLock(owner);
        }
      } catch {
        /* best effort */
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

// ── External refs (injected by monitor.mjs) ──────────────────────────────────

let _sendTelegramMessage = null; // injected from monitor.mjs
let _readStatusData = null;
let _readStatusSummary = null;
let _getCurrentChild = null;
let _startProcess = null;
let _getVibeKanbanUrl = null;
let _fetchVk = null;
let _getRepoRoot = null;

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
}) {
  _sendTelegramMessage = sendTelegramMessage;
  _readStatusData = readStatusData;
  _readStatusSummary = readStatusSummary;
  _getCurrentChild = getCurrentChild;
  _startProcess = startProcess;
  _getVibeKanbanUrl = getVibeKanbanUrl;
  _fetchVk = fetchVk;
  _getRepoRoot = getRepoRoot;
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
  } catch { /* best effort */ }
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
  } catch { /* best effort */ }
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

/**
 * Convert a raw Codex event into a concise human-readable action description.
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
                c.kind === "add"
                  ? "➕"
                  : c.kind === "delete"
                    ? "🗑️"
                    : "✏️";
              // Show line counts if available
              const adds = c.additions ?? c.lines_added ?? 0;
              const dels = c.deletions ?? c.lines_deleted ?? 0;
              const stats =
                adds || dels
                  ? ` (+${adds} -${dels})`
                  : "";
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
    return target ? `running PowerShell file ${target}` : "running PowerShell file";
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
  if (/^node\s+-[ec]/i.test(c)) return `running Node.js script: ${shortSnippet(clean, 60)}`;
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
  if (/^docker\s+/i.test(c)) return `running docker: ${shortSnippet(clean, 60)}`;

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
    return target ? `running PowerShell: ${snippet} → ${target}` : `running PowerShell: ${snippet}`;
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
  // If Codex is already busy, handle immediately so follow-ups can be queued.
  if (isCodexBusy()) {
    void handleFreeText(text, chatId);
    return;
  }
  enqueueAgentTask(() => handleFreeText(text, chatId));
}

// ── Command Router ────────────────────────────────────────────────────────────

const COMMANDS = {
  "/help": { handler: cmdHelp, desc: "Show available commands" },
  "/status": { handler: cmdStatus, desc: "Detailed orchestrator status" },
  "/tasks": { handler: cmdTasks, desc: "Active tasks, workspace metrics & retries" },
  "/logs": { handler: cmdLogs, desc: "Recent monitor logs" },
  "/log": { handler: cmdLogs, desc: "Alias for /logs" },
  "/branches": { handler: cmdBranches, desc: "Recent git branches" },
  "/diff": { handler: cmdDiff, desc: "Git diff summary (staged)" },
  "/restart": { handler: cmdRestart, desc: "Restart orchestrator process" },
  "/history": { handler: cmdHistory, desc: "Codex conversation history" },
  "/clear": { handler: cmdClear, desc: "Clear Codex conversation context" },
  "/reset_thread": { handler: cmdClear, desc: "Alias for /clear (reset thread)" },
  "/git": { handler: cmdGit, desc: "Run a git command: /git log --oneline -5" },
  "/shell": { handler: cmdShell, desc: "Run a shell command: /shell ls -la" },
  "/background": {
    handler: cmdBackground,
    desc: "Run a task in background or background the active agent",
  },
  "/agent": {
    handler: cmdAgent,
    desc: "Dispatch a task to a workspace: /agent --workspace <id> --task <prompt>",
  },
  "/region": {
    handler: cmdRegion,
    desc: "View/switch Codex region: /region [us|sweden|auto]",
  },
  "/health": {
    handler: cmdHealth,
    desc: "Executor health status & model routing",
  },
  "/model": {
    handler: cmdModel,
    desc: "Override executor for next task: /model gpt-5.2-codex",
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
};

const FAST_COMMANDS = new Set(["/status", "/tasks"]);

async function handleCommand(text, chatId) {
  const parts = text.split(/\s+/);
  const cmd = parts[0].toLowerCase().replace(/@\w+/, ""); // strip @botname
  const cmdArgs = parts.slice(1).join(" ");

  const entry = COMMANDS[cmd];
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

function parseAgentArgs(args) {
  const tokens = splitArgs(args);
  let workspaceId = null;
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
    task: taskTokens.join(" ").trim(),
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
  const lines = ["🤖 VirtEngine Codex Shell Commands:\n"];
  for (const [cmd, { desc }] of Object.entries(COMMANDS)) {
    lines.push(`${cmd} — ${desc}`);
  }
  lines.push("", "Any other text → sent to Codex AI (full repo + MCP access)");
  await sendReply(chatId, lines.join("\n"));
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

async function cmdTasks(chatId) {
  try {
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
      const emoji =
        status === "running"
          ? "🔄"
          : status === "review"
            ? "👀"
            : status === "error"
              ? "❌"
              : status === "completed"
                ? "✅"
                : "⏸️";
      const branch = attempt.branch || "";
      const pr = attempt.pr_number ? ` PR#${attempt.pr_number}` : "";
      const title = attempt.task_title || attempt.task_id || id;
      const executor = attempt.executor || "?";

      lines.push(`${emoji} ${title}${pr}`);
      lines.push(`   Status: ${status} | Agent: ${executor}`);

      // ── Workspace duration ──────────────────────────────────────
      const started = attempt.started_at || attempt.created_at || attempt.updated_at;
      if (started) {
        const dur = Date.now() - Date.parse(started);
        const mins = Math.floor(dur / 60000);
        const hrs = Math.floor(mins / 60);
        const remMin = mins % 60;
        const durStr = hrs > 0 ? `${hrs}h ${remMin}m` : `${mins}m`;
        lines.push(`   ⏱️ Active: ${durStr}`);
      }

      // ── Retry count (from failure_counts tracked by auto-reattempt) ──
      const failKey = attempt.task_id || id;
      const failCount = data.task_failure_counts?.[failKey] || 0;
      if (failCount > 0) {
        lines.push(`   🔁 Failures: ${failCount}/3${failCount >= 3 ? " → auto-reattempt" : ""}`);
      }

      // ── Git diff stats for the branch ──────────────────────────
      if (branch) {
        try {
          const diffStat = execSync(
            `git diff --shortstat main...${branch} 2>nul || echo ""`,
            { cwd: repoRoot, encoding: "utf8", timeout: 8000 },
          ).trim();
          if (diffStat) {
            // Extract insertions/deletions from "N files changed, X insertions(+), Y deletions(-)"
            const insMatch = diffStat.match(/(\d+) insertion/);
            const delMatch = diffStat.match(/(\d+) deletion/);
            const filesMatch = diffStat.match(/(\d+) file/);
            const ins = insMatch ? insMatch[1] : "0";
            const del = delMatch ? delMatch[1] : "0";
            const files = filesMatch ? filesMatch[1] : "0";
            lines.push(`   📊 ${files} files | +${ins} -${del}`);
          }
        } catch { /* git diff not available or branch doesn't exist */ }
      }
      lines.push(""); // spacing between tasks
    }

    // ── Overall workspace summary ────────────────────────────────
    const running = Object.values(attempts).filter((a) => a?.status === "running").length;
    const errors = Object.values(attempts).filter((a) => a?.status === "error").length;
    const reviews = Object.values(attempts).filter(
      (a) => a?.status === "review" || a?.status === "manual_review",
    ).length;
    lines.push("────────────────────────────");
    lines.push(`Total: ${Object.keys(attempts).length} | Running: ${running} | Review: ${reviews} | Error: ${errors}`);

    await sendReply(chatId, lines.join("\n"));
  } catch (err) {
    await sendReply(chatId, `Error reading tasks: ${err.message}`);
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

async function cmdHistory(chatId) {
  const info = getThreadInfo();
  const lines = [
    `🧠 Codex Agent Thread`,
    "",
    `Thread ID: ${info.threadId || "(none)"}`,
    `Turns: ${info.turnCount}`,
    `Active: ${info.isActive ? "yes" : "no"}`,
    `Busy: ${info.isBusy ? "yes" : "no"}`,
    "",
    "The thread persists across messages.",
    "Use /clear to start a fresh thread.",
  ];
  await sendReply(chatId, lines.join("\n"));
}

async function cmdClear(chatId) {
  await resetThread();
  await sendReply(
    chatId,
    "🧹 Agent thread reset. Next message starts a fresh conversation.",
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
      `⚠️ Blocked: 'git ${gitArgs}' is a destructive command. Use Codex shell for that.`,
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
  const result = spawnSync(
    pwsh,
    ["-NoProfile", "-Command", script],
    { cwd: repoRoot, encoding: "utf8", timeout: timeoutMs },
  );
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

async function cmdRegion(chatId, regionArg) {
  if (!regionArg || regionArg.trim() === "") {
    // Show current region status
    try {
      const result = runPwsh(
        `. '${resolve(repoRoot, "scripts", "ve-kanban.ps1")}'; Initialize-CodexRegionTracking; Get-RegionStatus | ConvertTo-Json -Depth 3`,
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
        ? `. '${resolve(repoRoot, "scripts", "ve-kanban.ps1")}'; Set-RegionOverride -Region $null | ConvertTo-Json`
        : `. '${resolve(repoRoot, "scripts", "ve-kanban.ps1")}'; Set-RegionOverride -Region '${target}' | ConvertTo-Json`;
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
    const psScript = [
      `. '${resolve(repoRoot, "scripts", "ve-kanban.ps1")}';`,
      "Initialize-ExecutorHealth;",
      "$out = @();",
      "foreach ($exec in $script:VK_EXECUTORS) {",
      "  $p = $exec.provider;",
      "  $h = $script:ExecutorHealth[$p];",
      "  $status = Get-ExecutorHealthStatus -Provider $p;",
      "  $out += @{",
      "    model = $exec.model;",
      "    variant = $exec.variant;",
      "    tier = $exec.tier;",
      "    region = $exec.region;",
      "    status = $status;",
      "    active = if ($h) { $h.active_tasks } else { 0 };",
      "    failures = if ($h) { $h.consecutive_failures } else { 0 };",
      "    successes = if ($h) { $h.total_successes } else { 0 };",
      "    timeouts = if ($h) { $h.total_timeouts } else { 0 };",
      "    rate_limits = if ($h) { $h.total_rate_limits } else { 0 };",
      "  }",
      "}",
      "$out | ConvertTo-Json -Depth 3",
    ].join(" ");

    const result = runPwsh(psScript);

    const executors = JSON.parse(result);
    const arr = Array.isArray(executors) ? executors : [executors];

    const iconMap = {
      healthy: "✅",
      degraded: "⚠️",
      cooldown: "⏸️",
      disabled: "❌",
    };
    const lines = ["🏥 Executor Health Dashboard\n"];

    for (const e of arr) {
      const icon = iconMap[e.status] || "❓";
      lines.push(
        `${icon} ${e.model} (${e.tier}/${e.region})\n` +
          `   Status: ${e.status} | Active: ${e.active}\n` +
          `   ✓${e.successes} ✗${e.failures} ⏱${e.timeouts} 🚫${e.rate_limits}`,
      );
    }

    // Add region info
    const regionScript = [
      `. '${resolve(repoRoot, "scripts", "ve-kanban.ps1")}';`,
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

    await sendReply(chatId, lines.join("\n"));
  } catch (err) {
    await sendReply(chatId, `Error reading health: ${err.message}`);
  }
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

// ── /agent — route to workspace registry ────────────────────────────────────

const MODEL_PROFILE_MAP = {
  "gpt-5.2-codex": { executor: "CODEX", variant: "DEFAULT", model: "gpt-5.2-codex" },
  "gpt-5.1-codex-max": { executor: "CODEX", variant: "DEFAULT", model: "gpt-5.1-codex-max" },
  "gpt-5.1-codex-mini": { executor: "CODEX", variant: "DEFAULT", model: "gpt-5.1-codex-mini" },
  "claude-opus-4.6": { executor: "COPILOT", variant: "CLAUDE_OPUS_4_6", model: "claude-opus-4.6" },
  "claude-code": { executor: "COPILOT", variant: "CLAUDE_CODE", model: "claude-code" },
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
      throw new Error(`VK ${res.status}: ${text.slice(0, 200) || res.statusText}`);
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
    summary?.latest_process_status ??
      summary?.status ??
      summary?.state ??
      "",
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
    const h = healthMap.get(ws.id) || { available: true, score: 1, status: "unknown" };
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
  const { message, workspaceId, role, model, queue, newSession, dryRun } = parsed;

  const { registry, errors, warnings } = await loadWorkspaceRegistry();
  const diagnostics = formatRegistryDiagnostics(errors, warnings);
  if (diagnostics) {
    await sendReply(chatId, diagnostics);
  }

  if (!message) {
    const list = registry.workspaces.map((ws) => `  - ${ws.id} (${ws.role})`).join("\n");
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
      await sendReply(chatId, `Unknown workspace: ${workspaceId}\nAvailable: ${ids}`);
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

  const healthMap = await getWorkspaceHealth(candidates);
  const selection = selectWorkspace(candidates, healthMap, { preferredId });
  if (!selection.workspace) {
    await sendReply(chatId, "No available workspace found for routing.");
    return;
  }

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
    infoLines.push(`Session: ${result.sessionId}${result.created ? " (new)" : ""}`);
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
    } catch { /* best effort */ }
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
    } catch { /* best effort */ }
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

  if (!activeAgentSession || !isCodexBusy()) {
    await sendReply(chatId, "No active agent. Sending as a new task.");
    await handleFreeText(message, chatId);
    return;
  }

  const result = await steerCodexPrompt(message);
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
    activeAgentSession.actionLog.push({
      icon: "🧭",
      text: `Steering queued (#${qLen})`,
    });
    if (activeAgentSession.scheduleEdit) {
      activeAgentSession.scheduleEdit();
    }
  }
  await sendReply(chatId, `🧭 Steering queued (#${qLen}).`);
}

// ── Free-text → Codex Agent Dispatch ──────────────────────────────────────────

/**
 * Build the rolling summary message text from accumulated action log.
 * This is the single message that gets continuously edited in Telegram.
 */
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
  const MAX_VISIBLE_ACTIONS = 12;
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
  if (isCodexBusy() && activeAgentSession) {
    if (!activeAgentSession.followUpQueue) {
      activeAgentSession.followUpQueue = [];
    }
    activeAgentSession.followUpQueue.push(text);
    const qLen = activeAgentSession.followUpQueue.length;

    // Try immediate steering so the in-flight run can adapt ASAP.
    const steerResult = await steerCodexPrompt(text);
    const steerNote = steerResult.ok
      ? `Steer ${steerResult.mode}.`
      : `Steer failed (${steerResult.reason}).`;

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
      });
      // Trigger an edit to show the follow-up in the streaming message
      if (activeAgentSession.scheduleEdit) {
        activeAgentSession.scheduleEdit();
      }
    }
    return;
  }

  // ── Block if Codex is busy but no session (shouldn't happen normally) ──
  if (isCodexBusy()) {
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
    const result = await execCodexPrompt(text, {
      statusData,
      timeoutMs: CODEX_TIMEOUT_MS,
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
        actionLog.push({ icon: "📌", text: `Processing follow-up: "${followUp.slice(0, 60)}"` });
        phase = "processing follow-up…";
        scheduleEdit();

        try {
          const followUpResult = await execCodexPrompt(followUp, {
            statusData,
            timeoutMs: CODEX_TIMEOUT_MS,
            onEvent,
            sendRawEvents: true,
          });

          // Merge follow-up results
          if (followUpResult.finalResponse) {
            result.finalResponse = (result.finalResponse || "") +
              `\n\n📌 Follow-up result:\n${followUpResult.finalResponse}`;
          }
        } catch (err) {
          actionLog.push({ icon: "❌", text: `Follow-up error: ${err.message}` });
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

  // Initialize the Codex shell context
  await initCodexShell();

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

  // Send startup notification
  await sendDirect(
    telegramChatId,
    "🤖 VirtEngine Codex Shell online.\n\nType /help for commands or send any message to chat with Codex.",
  );

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
export function stopTelegramBot() {
  polling = false;
  if (pollAbort) {
    try {
      pollAbort.abort();
    } catch {
      /* best effort */
    }
  }
  void releaseTelegramPollLock();
  console.log("[telegram-bot] stopped");
}
