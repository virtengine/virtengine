/**
 * agent-pool.mjs — Universal SDK-Aware Ephemeral Agent Pool
 *
 * WHY THIS EXISTS:
 * ────────────────
 * The primary agent in monitor.mjs is a long-lived singleton thread.
 * Every operation that calls `execPrimaryPrompt` serialises behind that single
 * thread — task attempts, conflict resolution, follow-ups, and health-checks
 * all compete for the same lock.  Under load (or when a single prompt is
 * slow) this creates a bottleneck that stalls the entire monitor pipeline.
 *
 * This module provides **ephemeral, per-operation SDK threads** that spin up
 * on demand and tear down after a single prompt completes.  Each call gets its
 * own isolated thread, so multiple operations can run concurrently without
 * blocking each other.
 *
 * MULTI-SDK SUPPORT:
 * ──────────────────
 * The pool dynamically selects the correct SDK adapter (Codex, Copilot, or
 * Claude) based on configuration.  Resolution order:
 *   1. `AGENT_POOL_SDK` env var (explicit override)
 *   2. `PRIMARY_AGENT` env var → maps to SDK
 *   3. `loadConfig().agentPool.sdk` from `codex-monitor.config.json`
 *   4. Fallback chain through available SDKs
 *
 * EXPORTS:
 *   launchEphemeralThread(prompt, cwd, timeoutMs, extra?)
 *     → Low-level: spawns a fresh SDK thread, runs one prompt,
 *       returns { success, output, items, error, sdk }.
 *
 *   execPooledPrompt(userMessage, options?)
 *     → High-level: matches the execPrimaryPrompt signature
 *       ({ finalResponse, items, usage }) so callers in monitor.mjs can
 *       swap in without changing surrounding code.
 *
 *   getPoolSdkName()     → returns current pool SDK name
 *   setPoolSdk(name)     → override pool SDK at runtime
 *   resetPoolSdkCache()  → force re-resolution
 *   getAvailableSdks()   → returns list of non-disabled SDKs
 */

import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import { loadConfig } from "./config.mjs";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

/** Repository root (two levels up from scripts/codex-monitor/) */
const REPO_ROOT = resolve(__dirname, "..", "..");

/** Default timeout: 90 minutes */
const DEFAULT_TIMEOUT_MS = 90 * 60 * 1000;

/** Tag for console logging */
const TAG = "[agent-pool]";

// ---------------------------------------------------------------------------
// SDK Adapter Registry
// ---------------------------------------------------------------------------

/**
 * @typedef {Object} SdkAdapter
 * @property {string}   name           Human-readable SDK name.
 * @property {Function} load           Async loader returning the launcher fn.
 * @property {string}   envDisableKey  Env var name that disables this SDK.
 */

/**
 * Registry of supported SDK adapters.
 * Each entry maps a canonical name to its loader and disable-check env var.
 * @type {Record<string, SdkAdapter>}
 */
const SDK_ADAPTERS = {
  codex: {
    name: "codex",
    load: loadCodexAdapter,
    envDisableKey: "CODEX_SDK_DISABLED",
  },
  copilot: {
    name: "copilot",
    load: loadCopilotAdapter,
    envDisableKey: "COPILOT_SDK_DISABLED",
  },
  claude: {
    name: "claude",
    load: loadClaudeAdapter,
    envDisableKey: "CLAUDE_SDK_DISABLED",
  },
};

/** Ordered fallback chain for SDK resolution */
const SDK_FALLBACK_ORDER = ["codex", "copilot", "claude"];

// ---------------------------------------------------------------------------
// SDK Resolution & Cache
// ---------------------------------------------------------------------------

/** @type {string|null} Cached resolved SDK name */
let resolvedSdkName = null;

/** @type {boolean} Whether initial resolution has been logged */
let resolutionLogged = false;

/**
 * Check whether an SDK is disabled via its env var.
 * @param {string} name SDK canonical name.
 * @returns {boolean}
 */
function isDisabled(name) {
  const adapter = SDK_ADAPTERS[name];
  if (!adapter) return true;
  return process.env[adapter.envDisableKey] === "1";
}

/**
 * Log which SDK was selected (only on first resolution).
 * @param {string} name SDK name.
 * @param {string} source How it was determined.
 */
function logResolution(name, source) {
  if (!resolutionLogged) {
    console.log(`${TAG} SDK selected: ${name} (via ${source})`);
    resolutionLogged = true;
  }
}

/**
 * Resolve which SDK the pool should use.
 *
 * Resolution order:
 *   1. Runtime override via `setPoolSdk()` (already cached)
 *   2. `AGENT_POOL_SDK` env var
 *   3. `PRIMARY_AGENT` env var
 *   4. `loadConfig().agentPool.sdk` from codex-monitor.config.json
 *   5. First non-disabled SDK in fallback chain
 *
 * @returns {string} Canonical SDK name (e.g. "codex", "copilot", "claude").
 */
function resolvePoolSdkName() {
  if (resolvedSdkName) return resolvedSdkName;

  // 1. AGENT_POOL_SDK env var (explicit override)
  const envPoolSdk = (process.env.AGENT_POOL_SDK || "").trim().toLowerCase();
  if (envPoolSdk && SDK_ADAPTERS[envPoolSdk] && !isDisabled(envPoolSdk)) {
    resolvedSdkName = envPoolSdk;
    logResolution(envPoolSdk, "AGENT_POOL_SDK env");
    return resolvedSdkName;
  }

  // 2. PRIMARY_AGENT env var
  const envPrimaryRaw = (process.env.PRIMARY_AGENT || "").trim().toLowerCase();
  // Normalize: "copilot-sdk" → "copilot", "codex-sdk" → "codex", etc.
  const envPrimary = envPrimaryRaw.replace(/-sdk$/, "");
  if (envPrimary && SDK_ADAPTERS[envPrimary] && !isDisabled(envPrimary)) {
    resolvedSdkName = envPrimary;
    logResolution(envPrimary, "PRIMARY_AGENT env");
    return resolvedSdkName;
  }

  // 3. codex-monitor.config.json → agentPool.sdk
  try {
    const config = loadConfig();
    const configSdk = (
      config?.agentPool?.sdk ||
      config?.primaryAgent ||
      ""
    ).toLowerCase();
    if (configSdk && SDK_ADAPTERS[configSdk] && !isDisabled(configSdk)) {
      resolvedSdkName = configSdk;
      logResolution(configSdk, "codex-monitor.config.json");
      return resolvedSdkName;
    }
  } catch {
    // config.mjs not available — continue with fallback
  }

  // 4. Fallback chain: first non-disabled SDK
  for (const name of SDK_FALLBACK_ORDER) {
    if (!isDisabled(name)) {
      resolvedSdkName = name;
      logResolution(name, "fallback chain");
      return resolvedSdkName;
    }
  }

  // All disabled — default to codex anyway (will fail at load time)
  resolvedSdkName = "codex";
  logResolution("codex", "last resort (all SDKs disabled)");
  return resolvedSdkName;
}

// ---------------------------------------------------------------------------
// Public SDK management API
// ---------------------------------------------------------------------------

/**
 * Get the name of the currently resolved pool SDK.
 * @returns {string} SDK name ("codex", "copilot", or "claude").
 */
export function getPoolSdkName() {
  return resolvePoolSdkName();
}

/**
 * Override the pool SDK at runtime.
 * @param {string} name SDK name ("codex", "copilot", or "claude").
 * @throws {Error} If the name is not a recognised SDK.
 */
export function setPoolSdk(name) {
  const normalised = (name || "").trim().toLowerCase();
  if (!SDK_ADAPTERS[normalised]) {
    throw new Error(
      `${TAG} unknown SDK "${name}". Valid: ${Object.keys(SDK_ADAPTERS).join(", ")}`,
    );
  }
  resolvedSdkName = normalised;
  resolutionLogged = false;
  logResolution(normalised, "setPoolSdk() runtime override");
}

/**
 * Force re-resolution of the pool SDK on next use.
 * Useful after environment changes.
 */
export function resetPoolSdkCache() {
  resolvedSdkName = null;
  resolutionLogged = false;
}

/**
 * Returns the list of SDK names that are not disabled.
 * @returns {string[]}
 */
export function getAvailableSdks() {
  return Object.keys(SDK_ADAPTERS).filter((name) => !isDisabled(name));
}

// ---------------------------------------------------------------------------
// Per-SDK Ephemeral Thread Launchers
// ---------------------------------------------------------------------------

/**
 * Launch a single ephemeral prompt via the **Codex SDK**.
 *
 * Creates a fresh `Codex` instance + thread, streams one turn, tears down.
 *
 * @param {string}  prompt     Prompt text.
 * @param {string}  cwd        Working directory.
 * @param {number}  timeoutMs  Abort timeout in ms.
 * @param {object}  extra      Optional { onEvent, abortController }.
 * @returns {Promise<{ success: boolean, output: string, items: Array, error: string|null, sdk: string }>}
 */
async function launchCodexThread(prompt, cwd, timeoutMs, extra = {}) {
  const { onEvent, abortController: externalAC } = extra;

  // ── 1. Load the SDK ──────────────────────────────────────────────────────
  let CodexClass;
  try {
    const mod = await import("@openai/codex-sdk");
    CodexClass = mod.Codex;
    if (!CodexClass) throw new Error("Codex export not found in SDK module");
  } catch (err) {
    return {
      success: false,
      output: "",
      items: [],
      error: `Codex SDK not available: ${err.message}`,
      sdk: "codex",
    };
  }

  // ── 2. Create an ephemeral thread ────────────────────────────────────────
  const codex = new CodexClass();
  const thread = codex.startThread({
    sandboxMode: "danger-full-access",
    workingDirectory: cwd,
    skipGitRepoCheck: true,
    approvalPolicy: "never",
  });

  // ── 3. Timeout / abort wiring ────────────────────────────────────────────
  const controller = externalAC || new AbortController();
  const timer = setTimeout(() => controller.abort("timeout"), timeoutMs);

  // ── 4. Stream the turn ───────────────────────────────────────────────────
  try {
    const turn = await thread.runStreamed(prompt, {
      signal: controller.signal,
    });

    let finalResponse = "";
    const allItems = [];

    for await (const event of turn.events) {
      if (typeof onEvent === "function") {
        try {
          onEvent(event);
        } catch {
          /* caller errors must not kill stream */
        }
      }
      if (event.type === "item.completed") {
        allItems.push(event.item);
        if (event.item.type === "agent_message" && event.item.text) {
          finalResponse += event.item.text + "\n";
        }
      }
    }

    clearTimeout(timer);
    const output =
      finalResponse.trim() || "(Agent completed with no text output)";
    return {
      success: true,
      output,
      items: allItems,
      error: null,
      sdk: "codex",
    };
  } catch (err) {
    clearTimeout(timer);
    if (err.name === "AbortError" || String(err) === "timeout") {
      return {
        success: false,
        output: "",
        items: [],
        error: `${TAG} codex timeout after ${timeoutMs}ms`,
        sdk: "codex",
      };
    }
    return {
      success: false,
      output: "",
      items: [],
      error: err.message,
      sdk: "codex",
    };
  }
}

/**
 * Launch a single ephemeral prompt via the **Copilot SDK**.
 *
 * Creates a fresh `CopilotClient`, starts it, opens an ephemeral session
 * (no reuse), sends the prompt, collects the response, and tears down.
 *
 * @param {string}  prompt     Prompt text.
 * @param {string}  cwd        Working directory.
 * @param {number}  timeoutMs  Abort timeout in ms.
 * @param {object}  extra      Optional { onEvent, abortController }.
 * @returns {Promise<{ success: boolean, output: string, items: Array, error: string|null, sdk: string }>}
 */
async function launchCopilotThread(prompt, cwd, timeoutMs, extra = {}) {
  const { onEvent, abortController: externalAC } = extra;

  // ── 1. Load the SDK ──────────────────────────────────────────────────────
  let CopilotClientClass;
  try {
    const mod = await import("@github/copilot-sdk");
    CopilotClientClass = mod.CopilotClient || mod.default?.CopilotClient;
    if (!CopilotClientClass) throw new Error("CopilotClient export not found");
  } catch (err) {
    return {
      success: false,
      output: "",
      items: [],
      error: `Copilot SDK not available: ${err.message}`,
      sdk: "copilot",
    };
  }

  // ── 2. Detect auth token ─────────────────────────────────────────────────
  const token =
    process.env.COPILOT_CLI_TOKEN ||
    process.env.GITHUB_TOKEN ||
    process.env.GH_TOKEN ||
    process.env.GITHUB_PAT ||
    undefined;

  // ── 3. Create & start ephemeral client ───────────────────────────────────
  const controller = externalAC || new AbortController();
  const timer = setTimeout(() => controller.abort("timeout"), timeoutMs);

  let client;
  try {
    const clientOpts = token ? { token } : undefined;
    client = new CopilotClientClass(clientOpts);
    await client.start();
  } catch (err) {
    clearTimeout(timer);
    return {
      success: false,
      output: "",
      items: [],
      error: `Copilot client start failed: ${err.message}`,
      sdk: "copilot",
    };
  }

  // ── 4. Create ephemeral session ──────────────────────────────────────────
  try {
    const sessionConfig = {
      streaming: true,
      systemMessage: {
        mode: "replace",
        content:
          "You are an ephemeral task agent. Execute the given task immediately. " +
          "Do NOT ask for confirmation. Produce concise, actionable output.",
      },
      infiniteSessions: { enabled: false },
    };

    const session = await client.createSession(sessionConfig);

    // ── 5. Send prompt & collect response ──────────────────────────────────
    let finalResponse = "";
    const allItems = [];

    // Wire up event listener if session supports it
    let unsubscribe = null;
    if (typeof session.on === "function") {
      unsubscribe = session.on((event) => {
        if (!event) return;
        allItems.push(event);
        if (event.type === "assistant.message" && event.data?.content) {
          finalResponse = event.data.content;
        }
        if (
          event.type === "assistant.message_delta" &&
          event.data?.deltaContent
        ) {
          finalResponse += event.data.deltaContent;
        }
        if (typeof onEvent === "function") {
          try {
            onEvent(event);
          } catch {
            /* best effort */
          }
        }
      });
    }

    const formattedPrompt =
      `# YOUR TASK — EXECUTE NOW\n\n${prompt}\n\n---\n` +
      'Do NOT respond with "Ready" or ask what to do. EXECUTE this task.';

    const sendFn = session.sendAndWait || session.send;
    if (typeof sendFn !== "function") {
      throw new Error("Copilot session does not support send");
    }

    const sendPromise = sendFn.call(session, { prompt: formattedPrompt });

    // If only send() (not sendAndWait), wait for idle event
    if (!session.sendAndWait && typeof session.on === "function") {
      await new Promise((resolveP, rejectP) => {
        const idleHandler = (event) => {
          if (event?.type === "session.idle") resolveP();
          if (event?.type === "session.error") {
            rejectP(new Error(event.data?.message || "session error"));
          }
        };
        const off = session.on(idleHandler);
        Promise.resolve(sendPromise).catch(rejectP);
        setTimeout(() => {
          if (typeof off === "function") off();
          resolveP();
        }, timeoutMs + 1000);
      });
    } else {
      await sendPromise;
    }

    clearTimeout(timer);
    if (typeof unsubscribe === "function") unsubscribe();

    const output =
      finalResponse.trim() || "(Agent completed with no text output)";
    return {
      success: true,
      output,
      items: allItems,
      error: null,
      sdk: "copilot",
    };
  } catch (err) {
    clearTimeout(timer);
    if (err.name === "AbortError" || String(err) === "timeout") {
      return {
        success: false,
        output: "",
        items: [],
        error: `${TAG} copilot timeout after ${timeoutMs}ms`,
        sdk: "copilot",
      };
    }
    return {
      success: false,
      output: "",
      items: [],
      error: err.message,
      sdk: "copilot",
    };
  } finally {
    // Best-effort teardown — don't let cleanup errors propagate
    try {
      if (client && typeof client.stop === "function") client.stop();
    } catch {
      /* ignore */
    }
  }
}

/**
 * Launch a single ephemeral prompt via the **Claude Agent SDK**.
 *
 * Creates a fresh message queue, pushes the user message, iterates the
 * response stream, and collects text output.  Fully ephemeral — no session
 * reuse.
 *
 * @param {string}  prompt     Prompt text.
 * @param {string}  cwd        Working directory.
 * @param {number}  timeoutMs  Abort timeout in ms.
 * @param {object}  extra      Optional { onEvent, abortController }.
 * @returns {Promise<{ success: boolean, output: string, items: Array, error: string|null, sdk: string }>}
 */
async function launchClaudeThread(prompt, cwd, timeoutMs, extra = {}) {
  const { onEvent, abortController: externalAC } = extra;

  // ── 1. Load the SDK ──────────────────────────────────────────────────────
  let queryFn;
  try {
    const mod = await import("@anthropic-ai/claude-agent-sdk");
    queryFn = mod.query;
    if (!queryFn) throw new Error("query() not found in Claude SDK");
  } catch (err) {
    return {
      success: false,
      output: "",
      items: [],
      error: `Claude SDK not available: ${err.message}`,
      sdk: "claude",
    };
  }

  // ── 2. Detect auth ──────────────────────────────────────────────────────
  const apiKey =
    process.env.ANTHROPIC_API_KEY ||
    process.env.CLAUDE_API_KEY ||
    process.env.CLAUDE_KEY ||
    undefined;

  // ── 3. Build message queue ───────────────────────────────────────────────
  const controller = externalAC || new AbortController();
  const timer = setTimeout(() => controller.abort("timeout"), timeoutMs);

  /**
   * Minimal async message queue for the Claude SDK streaming interface.
   * @returns {{ iterator: Function, push: Function, close: Function }}
   */
  function createMessageQueue() {
    const q = [];
    let resolver = null;
    let closed = false;

    async function* iterator() {
      while (true) {
        if (q.length > 0) {
          yield q.shift();
          continue;
        }
        if (closed) return;
        await new Promise((r) => {
          resolver = r;
        });
        resolver = null;
      }
    }
    function push(msg) {
      if (closed) return false;
      q.push(msg);
      if (resolver) {
        resolver();
        resolver = null;
      }
      return true;
    }
    function close() {
      closed = true;
      if (resolver) {
        resolver();
        resolver = null;
      }
    }
    return { iterator, push, close };
  }

  /**
   * Build a Claude-format user message.
   * @param {string} text
   * @returns {object}
   */
  function makeUserMessage(text) {
    return {
      type: "user",
      session_id: "",
      message: {
        role: "user",
        content: [{ type: "text", text }],
      },
      parent_tool_use_id: null,
    };
  }

  // ── 4. Execute query ─────────────────────────────────────────────────────
  try {
    const msgQueue = createMessageQueue();

    const formattedPrompt =
      `# YOUR TASK — EXECUTE NOW\n\n${prompt}\n\n---\n` +
      'Do NOT respond with "Ready" or ask what to do. EXECUTE this task.';

    msgQueue.push(makeUserMessage(formattedPrompt));

    /** @type {object} */
    const options = {
      cwd,
      settingSources: ["user", "project"],
      permissionMode: process.env.CLAUDE_PERMISSION_MODE || "bypassPermissions",
    };
    if (apiKey) options.apiKey = apiKey;

    const model =
      process.env.CLAUDE_MODEL ||
      process.env.CLAUDE_CODE_MODEL ||
      process.env.ANTHROPIC_MODEL ||
      "";
    if (model) options.model = model;

    const result = queryFn({
      prompt: msgQueue.iterator(),
      options,
    });

    let finalResponse = "";
    const allItems = [];

    for await (const message of result) {
      // Extract text from assistant messages
      const contentBlocks = message?.message?.content || message?.content || [];

      if (message?.type === "assistant" && Array.isArray(contentBlocks)) {
        for (const block of contentBlocks) {
          if (block?.type === "text" && block.text) {
            finalResponse += block.text + "\n";
          }
        }
      }

      // Normalise to item-style events for the onEvent callback
      const syntheticEvent = { type: message?.type || "unknown", message };
      allItems.push(syntheticEvent);
      if (typeof onEvent === "function") {
        try {
          onEvent(syntheticEvent);
        } catch {
          /* best effort */
        }
      }

      // If the SDK signals completion, close the queue
      if (message?.type === "result") {
        msgQueue.close();
      }
    }

    clearTimeout(timer);
    msgQueue.close();

    const output =
      finalResponse.trim() || "(Agent completed with no text output)";
    return {
      success: true,
      output,
      items: allItems,
      error: null,
      sdk: "claude",
    };
  } catch (err) {
    clearTimeout(timer);
    if (err.name === "AbortError" || String(err) === "timeout") {
      return {
        success: false,
        output: "",
        items: [],
        error: `${TAG} claude timeout after ${timeoutMs}ms`,
        sdk: "claude",
      };
    }
    return {
      success: false,
      output: "",
      items: [],
      error: err.message,
      sdk: "claude",
    };
  }
}

// ---------------------------------------------------------------------------
// Adapter loader functions (return the per-SDK launcher)
// ---------------------------------------------------------------------------

/**
 * @returns {Promise<Function>} The Codex launcher function.
 */
async function loadCodexAdapter() {
  return launchCodexThread;
}

/**
 * @returns {Promise<Function>} The Copilot launcher function.
 */
async function loadCopilotAdapter() {
  return launchCopilotThread;
}

/**
 * @returns {Promise<Function>} The Claude launcher function.
 */
async function loadClaudeAdapter() {
  return launchClaudeThread;
}

// ---------------------------------------------------------------------------
// Unified ephemeral thread launcher
// ---------------------------------------------------------------------------

/**
 * Spin up a fresh, isolated SDK thread, execute a single prompt, and return
 * the result.  The thread is not reused — it exists only for this one
 * operation, which means it cannot block (or be blocked by) any other thread.
 *
 * SDK selection:
 *   - Pass `extra.sdk` to force a specific SDK for this call.
 *   - Otherwise uses the resolved pool SDK (env / config / fallback).
 *   - If the primary SDK fails with "not available", tries the fallback chain.
 *
 * @param {string}  prompt      The prompt to send to the agent.
 * @param {string}  [cwd]       Working directory (defaults to REPO_ROOT).
 * @param {number}  [timeoutMs] Abort after this many ms (default 90 min).
 * @param {object}  [extra]     Optional extras:
 * @param {string}  [extra.sdk]             Force a specific SDK for this call.
 * @param {Function} [extra.onEvent]        Callback for raw SDK events.
 * @param {AbortController} [extra.abortController] External abort controller.
 * @returns {Promise<{ success: boolean, output: string, items: Array, error: string|null, sdk: string }>}
 */
export async function launchEphemeralThread(
  prompt,
  cwd = REPO_ROOT,
  timeoutMs = DEFAULT_TIMEOUT_MS,
  extra = {},
) {
  // Determine the primary SDK to try
  const requestedSdk = extra.sdk
    ? String(extra.sdk).trim().toLowerCase()
    : null;

  const primaryName =
    requestedSdk && SDK_ADAPTERS[requestedSdk]
      ? requestedSdk
      : resolvePoolSdkName();

  const primaryAdapter = SDK_ADAPTERS[primaryName];

  // ── Try primary SDK ──────────────────────────────────────────────────────
  if (primaryAdapter && !isDisabled(primaryName)) {
    const launcher = await primaryAdapter.load();
    const result = await launcher(prompt, cwd, timeoutMs, extra);

    // If it succeeded, or if the error isn't "not available", return as-is
    if (result.success || !result.error?.includes("not available")) {
      return result;
    }

    // Primary SDK not installed — fall through to fallback chain
    console.warn(
      `${TAG} primary SDK "${primaryName}" not available, trying fallback chain`,
    );
  }

  // ── Fallback chain ───────────────────────────────────────────────────────
  for (const name of SDK_FALLBACK_ORDER) {
    if (name === primaryName) continue; // already tried
    if (isDisabled(name)) continue;

    const adapter = SDK_ADAPTERS[name];
    if (!adapter) continue;

    console.log(`${TAG} trying fallback SDK: ${name}`);
    const launcher = await adapter.load();
    const result = await launcher(prompt, cwd, timeoutMs, extra);

    if (result.success || !result.error?.includes("not available")) {
      return result;
    }
  }

  // ── All SDKs exhausted ───────────────────────────────────────────────────
  const triedSdks = SDK_FALLBACK_ORDER.filter((n) => !isDisabled(n));
  return {
    success: false,
    output: "",
    items: [],
    error: `${TAG} no SDK available. Tried: ${triedSdks.join(", ") || "(all disabled)"}`,
    sdk: primaryName,
  };
}

// ---------------------------------------------------------------------------
// High-level: drop-in replacement for execPrimaryPrompt
// ---------------------------------------------------------------------------

/**
 * Execute a prompt on a pooled ephemeral thread with the **same signature** as
 * `execPrimaryPrompt` from codex-shell.mjs.  This allows callers in
 * monitor.mjs to swap from the singleton agent to a concurrent pool thread
 * without changing any surrounding code.
 *
 * @param {string} userMessage  The prompt / instruction to execute.
 * @param {object} [options]    Compatible with execPrimaryPrompt options.
 * @param {Function}           [options.onEvent]         Callback for raw SDK events.
 * @param {object}             [options.statusData]      (Unused — accepted for compat.)
 * @param {number}             [options.timeoutMs]       Override default timeout.
 * @param {boolean}            [options.sendRawEvents]   (Unused — accepted for compat.)
 * @param {AbortController}    [options.abortController] External abort controller.
 * @param {string}             [options.cwd]             Working directory override.
 * @param {string}             [options.sdk]             Force a specific SDK.
 * @returns {Promise<{ finalResponse: string, items: Array, usage: object|null }>}
 */
export async function execPooledPrompt(userMessage, options = {}) {
  const {
    onEvent,
    timeoutMs = DEFAULT_TIMEOUT_MS,
    abortController,
    cwd = REPO_ROOT,
    sdk,
    // statusData and sendRawEvents are accepted but not used — keeps the
    // call-site compatible with execPrimaryPrompt without modification.
  } = options;

  const result = await launchEphemeralThread(userMessage, cwd, timeoutMs, {
    onEvent,
    abortController,
    sdk,
  });

  if (!result.success) {
    // Match execPrimaryPrompt behaviour: always return the triple, let the
    // caller inspect finalResponse for error handling.
    return {
      finalResponse: result.error
        ? `[agent-pool error] ${result.error}`
        : "(no output)",
      items: result.items || [],
      usage: null,
    };
  }

  return {
    finalResponse: result.output,
    items: result.items,
    usage: null, // ephemeral threads don't aggregate usage today
  };
}
