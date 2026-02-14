/**
 * copilot-shell.mjs — Persistent Copilot SDK agent for the VirtEngine monitor.
 *
 * Uses the GitHub Copilot SDK (@github/copilot-sdk) to maintain a persistent
 * session with multi-turn conversation, tool use (shell, file I/O, MCP), and
 * streaming. Designed as a drop-in primary agent when Copilot is configured
 * as the primary executor.
 */

import { existsSync, readFileSync, appendFileSync, mkdirSync } from "node:fs";
import { readFile, writeFile, mkdir } from "node:fs/promises";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { execSync } from "node:child_process";
import { resolveRepoRoot } from "./repo-root.mjs";

const __dirname = resolve(fileURLToPath(new URL(".", import.meta.url)));

// ── Configuration ────────────────────────────────────────────────────────────

const DEFAULT_TIMEOUT_MS = 60 * 60 * 1000; // 60 min for agentic tasks
const STATE_FILE = resolve(__dirname, "logs", "copilot-shell-state.json");
const SESSION_LOG_DIR = resolve(__dirname, "logs", "copilot-sessions");
const REPO_ROOT = resolveRepoRoot();

// ── State ────────────────────────────────────────────────────────────────────

let CopilotClientClass = null; // CopilotClient class from SDK
let copilotClient = null;
let clientStarted = false;
let activeSession = null;
let activeSessionId = null;
let activeTurn = false;
let turnCount = 0;
let workspacePath = null;

function envFlagEnabled(value) {
  const raw = String(value ?? "")
    .trim()
    .toLowerCase();
  return ["1", "true", "yes", "on", "y"].includes(raw);
}

function resolveCopilotTransport() {
  const raw = String(process.env.COPILOT_TRANSPORT || "auto")
    .trim()
    .toLowerCase();
  if (["auto", "sdk", "cli", "url"].includes(raw)) {
    return raw;
  }
  console.warn(
    `[copilot-shell] invalid COPILOT_TRANSPORT='${raw}', defaulting to 'auto'`,
  );
  return "auto";
}

// ── Helpers ──────────────────────────────────────────────────────────────────

function timestamp() {
  return new Date().toISOString();
}

function safeJsonParse(raw) {
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

function safeStringify(value, maxLen = 8000) {
  let text = "";
  try {
    text = JSON.stringify(value);
  } catch {
    text = String(value);
  }
  if (text.length > maxLen) {
    text = text.slice(0, maxLen) + "...";
  }
  return text;
}

function initSessionLog(sessionId, prompt, timeoutMs) {
  if (!sessionId) return null;
  try {
    mkdirSync(SESSION_LOG_DIR, { recursive: true });
    const stamp = new Date().toISOString().replace(/[:.]/g, "-");
    const shortId = sessionId.slice(0, 8);
    const logPath = resolve(
      SESSION_LOG_DIR,
      `copilot-session-${stamp}-${shortId}.log`,
    );
    const header = [
      "# Copilot Session Log",
      `# Timestamp: ${timestamp()}`,
      `# Session ID: ${sessionId}`,
      `# Timeout: ${timeoutMs}ms`,
      `# Prompt: ${(prompt || "").slice(0, 500)}`,
      "",
    ].join("\n");
    appendFileSync(logPath, header + "\n");
    return logPath;
  } catch {
    return null;
  }
}

function logSessionEvent(logPath, event) {
  if (!logPath || !event) return;
  try {
    const payload = safeStringify(event);
    appendFileSync(
      logPath,
      `${timestamp()} ${event.type || "event"} ${payload}\n`,
    );
  } catch {
    /* best effort */
  }
}

// ── SDK Loading ──────────────────────────────────────────────────────────────

async function loadCopilotSdk() {
  if (CopilotClientClass) return CopilotClientClass;
  if (envFlagEnabled(process.env.COPILOT_SDK_DISABLED)) {
    console.warn("[copilot-shell] SDK disabled via COPILOT_SDK_DISABLED");
    return null;
  }
  try {
    const mod = await import("@github/copilot-sdk");
    CopilotClientClass =
      mod.CopilotClient || mod.default?.CopilotClient || null;
    if (!CopilotClientClass) {
      throw new Error("CopilotClient export not found");
    }
    console.log("[copilot-shell] SDK loaded successfully");
    return CopilotClientClass;
  } catch (err) {
    console.error(`[copilot-shell] failed to load SDK: ${err.message}`);
    return null;
  }
}

/**
 * Detect GitHub token from multiple sources (auth passthrough).
 * Priority: ENV > gh CLI > undefined (SDK will use default auth).
 */
function detectGitHubToken() {
  // 1. Direct token env vars (highest priority)
  const envToken =
    process.env.COPILOT_CLI_TOKEN ||
    process.env.GITHUB_TOKEN ||
    process.env.GH_TOKEN ||
    process.env.GITHUB_PAT;
  if (envToken) {
    console.log("[copilot-shell] using token from environment");
    return envToken;
  }

  // 2. Try to read from gh CLI auth
  try {
    execSync("gh auth status", { stdio: "pipe", encoding: "utf8" });
    console.log("[copilot-shell] detected gh CLI authentication");
    // gh CLI is authenticated - SDK will use it automatically
    return undefined;
  } catch {
    // gh not authenticated or not installed
  }

  // 3. VS Code auth detection could be added here
  // For now, return undefined to let SDK use default auth flow
  console.log("[copilot-shell] no pre-auth detected, using SDK default auth");
  return undefined;
}

const OPENAI_ENV_KEYS = [
  "OPENAI_API_KEY",
  "OPENAI_BASE_URL",
  "OPENAI_ORGANIZATION",
  "OPENAI_PROJECT",
];

async function withSanitizedOpenAiEnv(fn) {
  const saved = {};
  for (const key of OPENAI_ENV_KEYS) {
    if (Object.prototype.hasOwnProperty.call(process.env, key)) {
      saved[key] = process.env[key];
      delete process.env[key];
    }
  }
  try {
    return await fn();
  } finally {
    for (const [key, value] of Object.entries(saved)) {
      if (value !== undefined) process.env[key] = value;
    }
  }
}

async function ensureClientStarted() {
  if (clientStarted && copilotClient) return true;
  const Cls = await loadCopilotSdk();
  if (!Cls) return false;

  // Auth passthrough: detect from multiple sources
  const cliPath =
    process.env.COPILOT_CLI_PATH ||
    process.env.GITHUB_COPILOT_CLI_PATH ||
    undefined;
  const cliUrl = process.env.COPILOT_CLI_URL || undefined;
  const token = detectGitHubToken();
  const transport = resolveCopilotTransport();

  let clientOptions;
  if (transport === "url") {
    if (!cliUrl) {
      console.warn(
        "[copilot-shell] COPILOT_TRANSPORT=url requested but COPILOT_CLI_URL is unset; falling back to auto",
      );
      clientOptions = cliPath || token ? { cliPath, token } : undefined;
    } else {
      clientOptions = { cliUrl };
    }
  } else if (transport === "cli") {
    clientOptions = { cliPath: cliPath || "copilot", token };
  } else if (transport === "sdk") {
    clientOptions = token ? { token } : undefined;
  } else {
    clientOptions = cliUrl
      ? { cliUrl }
      : cliPath || token
        ? { cliPath, token }
        : undefined;
  }

  await withSanitizedOpenAiEnv(async () => {
    copilotClient = new Cls(clientOptions);
    await copilotClient.start();
  });
  clientStarted = true;
  console.log("[copilot-shell] client started");
  return true;
}

// ── State Persistence ────────────────────────────────────────────────────────

async function loadState() {
  try {
    const raw = await readFile(STATE_FILE, "utf8");
    const data = JSON.parse(raw);
    activeSessionId = data.sessionId || null;
    turnCount = data.turnCount || 0;
    workspacePath = data.workspacePath || null;
    console.log(
      `[copilot-shell] loaded state: sessionId=${activeSessionId}, turns=${turnCount}`,
    );
  } catch {
    activeSessionId = null;
    turnCount = 0;
    workspacePath = null;
  }
}

async function saveState() {
  try {
    await mkdir(resolve(__dirname, "logs"), { recursive: true });
    await writeFile(
      STATE_FILE,
      JSON.stringify(
        {
          sessionId: activeSessionId,
          turnCount,
          workspacePath,
          updatedAt: timestamp(),
        },
        null,
        2,
      ),
      "utf8",
    );
  } catch (err) {
    console.warn(`[copilot-shell] failed to save state: ${err.message}`);
  }
}

// ── System Prompt ────────────────────────────────────────────────────────────

const SYSTEM_PROMPT = `# AGENT DIRECTIVE — EXECUTE IMMEDIATELY

You are an autonomous AI coding agent deployed inside the VirtEngine orchestrator.
You are NOT a chatbot. You are NOT waiting for input. You EXECUTE tasks.

CRITICAL RULES:
1. NEVER respond with "Ready" or "What would you like me to do?" — you already have your task below.
2. NEVER ask clarifying questions — infer intent and take action.
3. DO the work. Read files, run commands, analyze code, write output.
4. Show your work as you go — print what you're reading, what you found, what you're doing next.
5. Produce DETAILED, STRUCTURED output with your findings and actions taken.
6. If the task involves analysis, actually READ the files and show what you found.
7. If the task involves code changes, actually MAKE the changes.
8. Think step-by-step, show your reasoning, then act.

You have FULL ACCESS to:
- The VirtEngine repository (Cosmos SDK blockchain + provider daemon + ML pipelines)
- Shell: git, gh, node, go, pwsh, make, and all system commands
- File read/write: read any file, create/edit any file
- MCP servers: GitHub, Playwright, Context7, Exa, Vibe-Kanban, Chrome DevTools
- Subagents and VS Code tools when available

Repository layout:
  app/          → Cosmos SDK app wiring
  x/            → Blockchain modules (veid, mfa, encryption, market, escrow, roles, hpc)
  pkg/          → Off-chain services (provider_daemon, inference, workflow)
  ml/           → Python ML pipelines
  scripts/      → Orchestrator, monitor, utilities
  portal/       → Frontend (Next.js)
  _docs/        → Architecture docs, progress tracking
  .github/      → CI, agents, instructions

Key files:
  scripts/codex-monitor/ve-orchestrator.ps1 — Main orchestrator (manages parallel AI agents)
  .cache/ve-orchestrator-status.json — Live status data
  scripts/codex-monitor/logs/ — Monitor logs
  _docs/ralph/progress.md — Project progress tracking
  .github/agents/ — Agent definitions (Task Planner, etc.)
  AGENTS.md — Repo guide for agents
`;

// ── MCP / Tool Config ────────────────────────────────────────────────────────

function loadMcpServersFromFile(path) {
  if (!path || !existsSync(path)) return null;
  const raw = readFileSync(path, "utf8");
  const parsed = safeJsonParse(raw);
  if (!parsed) return null;
  if (parsed.mcpServers && typeof parsed.mcpServers === "object") {
    return parsed.mcpServers;
  }
  if (
    parsed.mcp &&
    parsed.mcp.servers &&
    typeof parsed.mcp.servers === "object"
  ) {
    return parsed.mcp.servers;
  }
  if (parsed["github.copilot.mcpServers"]) {
    return parsed["github.copilot.mcpServers"];
  }
  return null;
}

function loadMcpServers() {
  if (process.env.COPILOT_MCP_SERVERS) {
    const parsed = safeJsonParse(process.env.COPILOT_MCP_SERVERS);
    if (parsed && typeof parsed === "object") {
      return parsed.mcpServers && typeof parsed.mcpServers === "object"
        ? parsed.mcpServers
        : parsed;
    }
  }
  const configPath =
    process.env.COPILOT_MCP_CONFIG || resolve(REPO_ROOT, ".vscode", "mcp.json");
  return loadMcpServersFromFile(configPath);
}

function buildSessionConfig() {
  const config = {
    streaming: true,
    systemMessage: {
      mode: "replace",
      content: SYSTEM_PROMPT,
    },
    infiniteSessions: { enabled: true },
  };
  const model =
    process.env.COPILOT_MODEL || process.env.COPILOT_SDK_MODEL || "";
  if (model) config.model = model;
  const mcpServers = loadMcpServers();
  if (mcpServers) config.mcpServers = mcpServers;
  return config;
}

// ── Session Management ───────────────────────────────────────────────────────

async function getSession() {
  if (activeSession) return activeSession;
  const started = await ensureClientStarted();
  if (!started) throw new Error("Copilot SDK not available");

  const config = buildSessionConfig();

  if (activeSessionId && typeof copilotClient?.resumeSession === "function") {
    try {
      activeSession = await copilotClient.resumeSession(
        activeSessionId,
        config,
      );
      workspacePath = activeSession?.workspacePath || workspacePath;
      console.log(`[copilot-shell] resumed session ${activeSessionId}`);
      return activeSession;
    } catch (err) {
      console.warn(
        `[copilot-shell] failed to resume session ${activeSessionId}: ${err.message} — starting fresh`,
      );
      activeSessionId = null;
    }
  }

  activeSession = await copilotClient.createSession(config);
  activeSessionId =
    activeSession?.sessionId || activeSession?.id || activeSessionId;
  workspacePath = activeSession?.workspacePath || workspacePath;
  await saveState();
  console.log(`[copilot-shell] new session started: ${activeSessionId}`);
  return activeSession;
}

// ── Main Execution ───────────────────────────────────────────────────────────

export async function execCopilotPrompt(userMessage, options = {}) {
  const {
    onEvent = null,
    statusData = null,
    timeoutMs = DEFAULT_TIMEOUT_MS,
    sendRawEvents = false,
    abortController = null,
  } = options;

  if (activeTurn) {
    return {
      finalResponse:
        "⏳ Agent is still executing a previous task. Please wait.",
      items: [],
      usage: null,
    };
  }

  activeTurn = true;

  let unsubscribe = null;
  const session = await getSession();
  const logPath = initSessionLog(activeSessionId, userMessage, timeoutMs);
  const items = [];
  let finalResponse = "";
  let responseFromMessage = false;

  const handleEvent = async (event) => {
    if (!event) return;
    logSessionEvent(logPath, event);
    items.push(event);
    if (event.type === "assistant.message" && event.data?.content) {
      finalResponse = event.data.content;
      responseFromMessage = true;
    }
    if (
      !responseFromMessage &&
      event.type === "assistant.message_delta" &&
      event.data?.deltaContent
    ) {
      finalResponse += event.data.deltaContent;
    }
    if (event.type === "session.idle") {
      turnCount += 1;
      await saveState();
    }
    if (onEvent) {
      try {
        if (sendRawEvents) {
          await onEvent(null, event);
        } else {
          await onEvent(null);
        }
      } catch {
        /* best effort */
      }
    }
  };

  try {
    if (typeof session.on === "function") {
      unsubscribe = session.on(handleEvent);
    }

    const controller = abortController || new AbortController();
    const timer = setTimeout(() => controller.abort("timeout"), timeoutMs);

    const onAbort = () => {
      const reason = controller.signal.reason || "user_stop";
      if (typeof session.abort === "function") session.abort(reason);
      if (typeof session.cancel === "function") session.cancel(reason);
      if (typeof session.stop === "function") session.stop(reason);
    };
    if (controller?.signal) {
      controller.signal.addEventListener("abort", onAbort, { once: true });
    }

    // Build prompt with optional orchestrator status
    let prompt = userMessage;
    if (statusData) {
      const statusSnippet = JSON.stringify(statusData, null, 2).slice(0, 2000);
      prompt = `[Orchestrator Status]\n\`\`\`json\n${statusSnippet}\n\`\`\`\n\n# YOUR TASK — EXECUTE NOW\n\n${userMessage}\n\n---\nDo NOT respond with "Ready" or ask what to do. EXECUTE this task. Read files, run commands, produce detailed output.`;
    } else {
      prompt = `# YOUR TASK — EXECUTE NOW\n\n${userMessage}\n\n---\nDo NOT respond with "Ready" or ask what to do. EXECUTE this task. Read files, run commands, produce detailed output.`;
    }

    const sendFn = session.sendAndWait || session.send;
    if (typeof sendFn !== "function") {
      throw new Error("Copilot SDK session does not support send");
    }

    // Pass timeout parameter to sendAndWait to override 60s SDK default
    const sendPromise = session.sendAndWait
      ? sendFn.call(session, { prompt }, timeoutMs)
      : sendFn.call(session, { prompt });

    // If send() returns before idle, wait for session.idle if available
    if (!session.sendAndWait) {
      await new Promise((resolve, reject) => {
        const idleHandler = (event) => {
          if (!event) return;
          if (event.type === "session.idle") resolve();
          if (event.type === "session.error") {
            reject(new Error(event.data?.message || "session error"));
          }
        };
        const off = session.on ? session.on(idleHandler) : null;
        Promise.resolve(sendPromise).catch(reject);
        setTimeout(resolve, timeoutMs + 1000);
        if (typeof off === "function") {
          setTimeout(() => off(), timeoutMs + 2000);
        }
      });
    } else {
      await sendPromise;
    }

    clearTimeout(timer);
    controller.signal?.removeEventListener("abort", onAbort);

    return {
      finalResponse:
        finalResponse.trim() || "(Agent completed with no text output)",
      items,
      usage: null,
    };
  } catch (err) {
    if (err?.name === "AbortError" || /abort|timeout/i.test(err?.message)) {
      const reason = abortController?.signal?.reason || "timeout";
      const msg =
        reason === "user_stop"
          ? "🛑 Agent stopped by user."
          : `⏱️ Agent timed out after ${timeoutMs / 1000}s`;
      return { finalResponse: msg, items: [], usage: null };
    }
    throw err;
  } finally {
    if (typeof unsubscribe === "function") {
      try {
        unsubscribe();
      } catch {
        /* best effort */
      }
    } else if (typeof session.off === "function") {
      try {
        session.off(handleEvent);
      } catch {
        /* best effort */
      }
    }
    activeTurn = false;
  }
}

/**
 * Copilot SDK does not currently expose steering APIs. We return unsupported.
 */
export async function steerCopilotPrompt() {
  return { ok: false, reason: "unsupported" };
}

export function isCopilotBusy() {
  return !!activeTurn;
}

export function getSessionInfo() {
  return {
    sessionId: activeSessionId,
    turnCount,
    isActive: !!activeSession,
    isBusy: !!activeTurn,
    workspacePath,
  };
}

export async function resetSession() {
  activeSession = null;
  activeSessionId = null;
  workspacePath = null;
  turnCount = 0;
  activeTurn = false;
  await saveState();
  console.log("[copilot-shell] session reset");
}

export async function initCopilotShell() {
  await loadState();
  const started = await ensureClientStarted();
  if (started) {
    console.log("[copilot-shell] initialised with Copilot SDK");
    return true;
  }
  console.warn(
    "[copilot-shell] initialised WITHOUT Copilot SDK — agent will not work",
  );
  return false;
}
