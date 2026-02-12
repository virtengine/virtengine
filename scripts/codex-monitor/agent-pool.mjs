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

/**
 * Hard timeout buffer: added on top of the soft timeout.
 * If the SDK's async iterator ignores the AbortSignal, this hard timeout
 * forcibly breaks the Promise.race to prevent infinite hangs.
 */
const HARD_TIMEOUT_BUFFER_MS = 60_000; // 60 seconds

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
      threadId: null,
    };
  }

  // ── 2. Create an ephemeral thread ────────────────────────────────────────
  // Sandbox policy: configurable via CODEX_SANDBOX env var or config
  // Options: "danger-full-access" (default — full write access for worktree workflows),
  //          "workspace-write" (restricted — breaks with worktrees), "read-only"
  const sandboxPolicy = process.env.CODEX_SANDBOX || "danger-full-access";

  const codex = new CodexClass();
  const thread = codex.startThread({
    sandboxMode: sandboxPolicy,
    workingDirectory: cwd,
    skipGitRepoCheck: true,
    approvalPolicy: "never",
  });

  if (!thread) {
    return {
      success: false,
      output: "",
      items: [],
      error: "Codex SDK startThread() returned null — SDK may be misconfigured or API unreachable",
      sdk: "codex",
      threadId: null,
    };
  }

  // ── 3. Timeout / abort wiring ────────────────────────────────────────────
  const controller = externalAC || new AbortController();
  const timer = setTimeout(() => controller.abort("timeout"), timeoutMs);

  // Hard timeout: safety net if the SDK's async iterator ignores AbortSignal.
  // Fires HARD_TIMEOUT_BUFFER_MS after the soft timeout to forcibly break the loop.
  let hardTimer;

  // ── 4. Stream the turn ───────────────────────────────────────────────────
  try {
    const turn = await thread.runStreamed(prompt, {
      signal: controller.signal,
    });

    let finalResponse = "";
    const allItems = [];

    // Race the event iterator against a hard timeout.
    // The soft timeout fires controller.abort() which the SDK should honor.
    // The hard timeout is a safety net in case the SDK iterator ignores the abort.
    const hardTimeoutPromise = new Promise((_, reject) => {
      hardTimer = setTimeout(
        () => reject(new Error("hard_timeout")),
        timeoutMs + HARD_TIMEOUT_BUFFER_MS,
      );
    });

    const iterateEvents = async () => {
      for await (const event of turn.events) {
        if (controller.signal.aborted) break;
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
    };

    await Promise.race([iterateEvents(), hardTimeoutPromise]);
    clearTimeout(hardTimer);
    clearTimeout(timer);

    const output =
      finalResponse.trim() || "(Agent completed with no text output)";
    return {
      success: true,
      output,
      items: allItems,
      error: null,
      sdk: "codex",
      threadId: thread.id || null,
    };
  } catch (err) {
    clearTimeout(timer);
    if (hardTimer) clearTimeout(hardTimer);
    const isTimeout =
      err.name === "AbortError" ||
      String(err) === "timeout" ||
      err.message === "hard_timeout";
    if (isTimeout) {
      return {
        success: false,
        output: "",
        items: [],
        error: `${TAG} codex timeout after ${timeoutMs}ms${err.message === "hard_timeout" ? " (hard timeout — SDK iterator unresponsive)" : ""}`,
        sdk: "codex",
        threadId: null,
      };
    }
    return {
      success: false,
      output: "",
      items: [],
      error: err.message,
      sdk: "codex",
      threadId: null,
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
      threadId: null,
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
      threadId: null,
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
        // Wire abort signal into this inner promise
        if (controller.signal) {
          const onAbort = () => {
            if (typeof off === "function") off();
            rejectP(new Error("timeout"));
          };
          if (controller.signal.aborted) {
            onAbort();
          } else {
            controller.signal.addEventListener("abort", onAbort, { once: true });
          }
        }
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
      threadId: null,
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
        threadId: null,
      };
    }
    return {
      success: false,
      output: "",
      items: [],
      error: err.message,
      sdk: "copilot",
      threadId: null,
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
      threadId: null,
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
        // Wire abort signal: if the controller fires while we're waiting
        // for the next message, break out of the wait instead of hanging forever.
        await new Promise((r) => {
          resolver = r;
          if (controller.signal) {
            const onAbort = () => {
              closed = true;
              r();
            };
            if (controller.signal.aborted) {
              closed = true;
              r();
              return;
            }
            controller.signal.addEventListener("abort", onAbort, { once: true });
          }
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
      threadId: null,
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
        threadId: null,
      };
    }
    return {
      success: false,
      output: "",
      items: [],
      error: err.message,
      sdk: "claude",
      threadId: null,
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
    threadId: null,
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

// ---------------------------------------------------------------------------
// Thread Persistence & Resume Registry
// ---------------------------------------------------------------------------

/**
 * @typedef {Object} ThreadRecord
 * @property {string}      threadId   SDK-specific thread/session ID.
 * @property {string}      sdk        Which SDK owns this thread.
 * @property {string}      taskKey    Caller-defined key (task ID, PR#, etc.).
 * @property {string}      cwd        Working directory used.
 * @property {number}      turnCount  How many turns have been run.
 * @property {number}      createdAt  Unix ms when first created.
 * @property {number}      lastUsedAt Unix ms of most recent run.
 * @property {string|null} lastError  Last error message if any.
 * @property {boolean}     alive      Whether this thread is still usable.
 */

/** @type {Map<string, ThreadRecord>} In-memory registry keyed by taskKey */
const threadRegistry = new Map();

const THREAD_REGISTRY_FILE = resolve(__dirname, "logs", "thread-registry.json");
const THREAD_MAX_AGE_MS = 4 * 60 * 60 * 1000; // 4 hours

/** Maximum turns before a thread is considered exhausted and must be replaced */
const MAX_THREAD_TURNS = 30;

/** Maximum absolute age for a thread (regardless of lastUsedAt) */
const THREAD_MAX_ABSOLUTE_AGE_MS = 8 * 60 * 60 * 1000; // 8 hours

/**
 * Load thread registry from disk (best-effort).
 */
async function loadThreadRegistry() {
  try {
    const { readFile } = await import("node:fs/promises");
    const raw = await readFile(THREAD_REGISTRY_FILE, "utf8");
    const entries = JSON.parse(raw);
    const now = Date.now();
    let pruned = 0;
    for (const [key, record] of Object.entries(entries)) {
      // Expire old threads (by lastUsedAt)
      if (now - record.lastUsedAt > THREAD_MAX_AGE_MS) { pruned++; continue; }
      // Expire threads that have been alive too long (absolute age)
      if (now - record.createdAt > THREAD_MAX_ABSOLUTE_AGE_MS) { pruned++; continue; }
      // Expire high-turn threads (context exhaustion)
      if (record.turnCount >= MAX_THREAD_TURNS) {
        console.log(`${TAG} expiring exhausted thread for task "${key}" (${record.turnCount} turns)`);
        pruned++;
        continue;
      }
      if (!record.alive) { pruned++; continue; }
      threadRegistry.set(key, record);
    }
    // Persist the cleaned registry back to disk so stale entries don't linger
    if (pruned > 0) {
      saveThreadRegistry().catch(() => {});
    }
  } catch {
    // No registry file yet — that's fine
  }
}

/**
 * Persist thread registry to disk (best-effort).
 */
async function saveThreadRegistry() {
  try {
    const { writeFile, mkdir } = await import("node:fs/promises");
    await mkdir(resolve(__dirname, "logs"), { recursive: true });
    const obj = Object.fromEntries(threadRegistry);
    await writeFile(THREAD_REGISTRY_FILE, JSON.stringify(obj, null, 2), "utf8");
  } catch {
    // Non-critical — registry is an optimisation, not a requirement
  }
}

// Load registry at module init
loadThreadRegistry().catch(() => {});

// ---------------------------------------------------------------------------
// Per-SDK Resume Launchers
// ---------------------------------------------------------------------------

/**
 * Resume an existing Codex thread and run a follow-up prompt.
 * Uses `codex.resumeThread(threadId)` from @openai/codex-sdk.
 *
 * @param {string} threadId  Thread ID from a previous launchCodexThread.
 * @param {string} prompt    Follow-up prompt.
 * @param {string} cwd       Working directory.
 * @param {number} timeoutMs Abort timeout in ms.
 * @param {object} extra     Optional { onEvent, abortController }.
 * @returns {Promise<{ success: boolean, output: string, items: Array, error: string|null, sdk: string, threadId: string|null }>}
 */
async function resumeCodexThread(threadId, prompt, cwd, timeoutMs, extra = {}) {
  const { onEvent, abortController: externalAC } = extra;

  let CodexClass;
  try {
    const mod = await import("@openai/codex-sdk");
    CodexClass = mod.Codex;
    if (!CodexClass) throw new Error("Codex export not found");
  } catch (err) {
    return {
      success: false,
      output: "",
      items: [],
      error: `Codex SDK not available: ${err.message}`,
      sdk: "codex",
      threadId: null,
    };
  }

  const codex = new CodexClass();

  let thread;
  try {
    const sandboxPolicy = process.env.CODEX_SANDBOX || "danger-full-access";
    thread = codex.resumeThread(threadId, {
      sandboxMode: sandboxPolicy,
      workingDirectory: cwd,
      skipGitRepoCheck: true,
      approvalPolicy: "never",
    });
  } catch (err) {
    // Resume failed (thread expired, not found, etc.) — signal caller to start fresh
    return {
      success: false,
      output: "",
      items: [],
      error: `Thread resume failed: ${err.message}`,
      sdk: "codex",
      threadId: null,
    };
  }

  if (!thread) {
    return {
      success: false,
      output: "",
      items: [],
      error: "Codex SDK resumeThread() returned null — thread may have expired",
      sdk: "codex",
      threadId: null,
    };
  }

  const controller = externalAC || new AbortController();
  const timer = setTimeout(() => controller.abort("timeout"), timeoutMs);
  let hardTimer;

  try {
    const turn = await thread.runStreamed(prompt, {
      signal: controller.signal,
    });
    let finalResponse = "";
    const allItems = [];

    // Hard timeout safety net (same as launchCodexThread)
    const hardTimeoutPromise = new Promise((_, reject) => {
      hardTimer = setTimeout(
        () => reject(new Error("hard_timeout")),
        timeoutMs + HARD_TIMEOUT_BUFFER_MS,
      );
    });

    const iterateEvents = async () => {
      for await (const event of turn.events) {
        if (controller.signal.aborted) break;
        if (typeof onEvent === "function")
          try {
            onEvent(event);
          } catch {
            /* */
          }
        if (event.type === "item.completed") {
          allItems.push(event.item);
          if (event.item.type === "agent_message" && event.item.text) {
            finalResponse += event.item.text + "\n";
          }
        }
      }
    };

    await Promise.race([iterateEvents(), hardTimeoutPromise]);
    clearTimeout(hardTimer);
    clearTimeout(timer);

    const newThreadId = thread.id || threadId;
    return {
      success: true,
      output: finalResponse.trim() || "(resumed — no text output)",
      items: allItems,
      error: null,
      sdk: "codex",
      threadId: newThreadId,
    };
  } catch (err) {
    clearTimeout(timer);
    if (hardTimer) clearTimeout(hardTimer);
    const isTimeout =
      err.name === "AbortError" ||
      String(err) === "timeout" ||
      err.message === "hard_timeout";
    return {
      success: false,
      output: "",
      items: [],
      error: isTimeout
        ? `${TAG} codex resume timeout after ${timeoutMs}ms${err.message === "hard_timeout" ? " (hard timeout)" : ""}`
        : `Thread resume error: ${err.message}`,
      sdk: "codex",
      threadId: null,
    };
  }
}

/**
 * "Resume" for SDKs without native thread persistence.
 * Falls back to starting a fresh thread with a context-carrying preamble.
 *
 * @param {string} _threadId  Ignored — no native resume available.
 * @param {string} prompt     Follow-up prompt.
 * @param {string} cwd        Working directory.
 * @param {number} timeoutMs  Abort timeout.
 * @param {object} extra      Optional extras.
 * @param {string} sdkName    "copilot" or "claude".
 * @returns {Promise<Object>}
 */
async function resumeGenericThread(
  _threadId,
  prompt,
  cwd,
  timeoutMs,
  extra = {},
  sdkName = "copilot",
) {
  // No native resume — launch fresh with context preamble
  const contextPrompt = `# CONTINUATION — Resuming Prior Context\n\nYou are continuing work from a previous session. Pick up where you left off.\n\n---\n\n${prompt}`;
  const launcher =
    sdkName === "claude" ? launchClaudeThread : launchCopilotThread;
  const result = await launcher(contextPrompt, cwd, timeoutMs, extra);
  return { ...result, threadId: null }; // No persistent ID available
}

// ---------------------------------------------------------------------------
// Thread-Persistent Launcher
// ---------------------------------------------------------------------------

/**
 * Launch a new thread OR resume an existing one for the given task key.
 *
 * When a `taskKey` is provided:
 *   1. Check the thread registry for an existing, alive thread.
 *   2. If found and the same SDK — attempt resume (Codex) or context-carry (others).
 *   3. If resume fails or no prior thread — start fresh.
 *   4. Register the new thread for future resume.
 *
 * Without `taskKey`, behaves identically to `launchEphemeralThread`.
 *
 * @param {string}  prompt      Prompt to run.
 * @param {string}  [cwd]       Working directory.
 * @param {number}  [timeoutMs] Timeout in ms.
 * @param {object}  [extra]     Options:
 * @param {string}  [extra.taskKey]    Key for thread registry (task ID, PR number, etc.)
 * @param {string}  [extra.sdk]        Force a specific SDK.
 * @param {Function} [extra.onEvent]   Event callback.
 * @param {AbortController} [extra.abortController]
 * @returns {Promise<{ success: boolean, output: string, items: Array, error: string|null, sdk: string, threadId: string|null, resumed: boolean }>}
 */
export async function launchOrResumeThread(
  prompt,
  cwd = REPO_ROOT,
  timeoutMs = DEFAULT_TIMEOUT_MS,
  extra = {},
) {
  const { taskKey, ...restExtra } = extra;

  // No taskKey — pure ephemeral (backward compatible)
  if (!taskKey) {
    const result = await launchEphemeralThread(
      prompt,
      cwd,
      timeoutMs,
      restExtra,
    );
    return { ...result, threadId: result.threadId || null, resumed: false };
  }

  // Check registry for existing thread
  const existing = threadRegistry.get(taskKey);
  if (existing && existing.alive && existing.threadId) {
    // Check if thread has exceeded max turns — force fresh start
    if (existing.turnCount >= MAX_THREAD_TURNS) {
      console.warn(
        `${TAG} thread for task "${taskKey}" exceeded ${MAX_THREAD_TURNS} turns (has ${existing.turnCount}) — invalidating and starting fresh`,
      );
      existing.alive = false;
      threadRegistry.set(taskKey, existing);
      saveThreadRegistry().catch(() => {});
      // Fall through to fresh launch below
    } else if (Date.now() - existing.createdAt > THREAD_MAX_ABSOLUTE_AGE_MS) {
      console.warn(
        `${TAG} thread for task "${taskKey}" exceeded absolute age limit — invalidating and starting fresh`,
      );
      existing.alive = false;
      threadRegistry.set(taskKey, existing);
      saveThreadRegistry().catch(() => {});
      // Fall through to fresh launch below
    } else {
    const sdkName = restExtra.sdk || existing.sdk || resolvePoolSdkName();

    // Only attempt native resume for Codex (it has resumeThread API)
    if (sdkName === "codex" && existing.sdk === "codex") {
      console.log(
        `${TAG} resuming Codex thread ${existing.threadId} for task "${taskKey}" (turn ${existing.turnCount + 1})`,
      );
      const result = await resumeCodexThread(
        existing.threadId,
        prompt,
        cwd,
        timeoutMs,
        restExtra,
      );

      if (result.success) {
        // Update registry
        existing.turnCount += 1;
        existing.lastUsedAt = Date.now();
        existing.lastError = null;
        if (result.threadId) existing.threadId = result.threadId;
        threadRegistry.set(taskKey, existing);
        saveThreadRegistry().catch(() => {});
        return { ...result, resumed: true };
      }

      // Resume failed — fall through to fresh launch
      console.warn(
        `${TAG} resume failed for task "${taskKey}": ${result.error}. Starting fresh.`,
      );
      existing.alive = false;
      threadRegistry.set(taskKey, existing);
    } else if (existing.sdk !== sdkName) {
      // SDK changed — invalidate old thread
      console.log(
        `${TAG} SDK changed from ${existing.sdk} to ${sdkName} for task "${taskKey}", starting fresh`,
      );
      existing.alive = false;
    } else {
      // Non-Codex SDK: use context-carry resume
      console.log(
        `${TAG} context-carry resume for ${sdkName} thread, task "${taskKey}"`,
      );
      const result = await resumeGenericThread(
        existing.threadId,
        prompt,
        cwd,
        timeoutMs,
        restExtra,
        sdkName,
      );

      if (result.success) {
        existing.turnCount += 1;
        existing.lastUsedAt = Date.now();
        existing.lastError = null;
        threadRegistry.set(taskKey, existing);
        saveThreadRegistry().catch(() => {});
        return { ...result, resumed: true };
      }

      console.warn(
        `${TAG} context-carry resume failed for task "${taskKey}": ${result.error}`,
      );
      existing.alive = false;
    }
    } // close else for turn-count / absolute-age guard
  }

  // Fresh launch — register the new thread
  const result = await launchEphemeralThread(prompt, cwd, timeoutMs, restExtra);

  // Register thread for future resume
  const record = {
    threadId: result.threadId || null,
    sdk: result.sdk,
    taskKey,
    cwd,
    turnCount: 1,
    createdAt: Date.now(),
    lastUsedAt: Date.now(),
    lastError: result.success ? null : result.error,
    alive: result.success && !!result.threadId,
  };
  threadRegistry.set(taskKey, record);
  saveThreadRegistry().catch(() => {});

  return { ...result, threadId: result.threadId || null, resumed: false };
}

// ---------------------------------------------------------------------------
// Error Recovery Wrapper
// ---------------------------------------------------------------------------

/**
 * Execute a prompt with automatic error recovery via thread resume.
 *
 * If the initial run fails, this will:
 *   1. Resume the same thread with the error context
 *   2. Ask the agent to diagnose and fix the issue
 *   3. Retry up to `maxRetries` times
 *
 * @param {string}  prompt      Initial prompt.
 * @param {object}  options     Options:
 * @param {string}  options.taskKey       Required — identifies the thread.
 * @param {string}  [options.cwd]         Working directory.
 * @param {number}  [options.timeoutMs]   Per-attempt timeout.
 * @param {number}  [options.maxRetries]  Max follow-up attempts (default: 2).
 * @param {Function} [options.shouldRetry] Custom predicate: (result) => boolean.
 * @param {Function} [options.buildRetryPrompt] Custom retry prompt builder: (result, attempt) => string.
 * @param {string}  [options.sdk]         Force SDK.
 * @param {Function} [options.onEvent]    Event callback.
 * @returns {Promise<{ success: boolean, output: string, items: Array, error: string|null, sdk: string, attempts: number, resumed: boolean }>}
 */
export async function execWithRetry(prompt, options = {}) {
  const {
    taskKey,
    cwd = REPO_ROOT,
    timeoutMs = DEFAULT_TIMEOUT_MS,
    maxRetries = 2,
    shouldRetry,
    buildRetryPrompt,
    sdk,
    onEvent,
    abortController,
  } = options;

  if (!taskKey) {
    throw new Error(
      `${TAG} execWithRetry requires a taskKey for thread persistence`,
    );
  }

  let lastResult = null;
  const totalAttempts = 1 + maxRetries;

  for (let attempt = 1; attempt <= totalAttempts; attempt++) {
    const currentPrompt =
      attempt === 1
        ? prompt
        : typeof buildRetryPrompt === "function"
          ? buildRetryPrompt(lastResult, attempt)
          : `# ERROR RECOVERY — Attempt ${attempt}/${totalAttempts}\n\nYour previous attempt failed with:\n\`\`\`\n${lastResult?.error || lastResult?.output || "(unknown error)"}\n\`\`\`\n\nPlease diagnose the issue, fix it, and try again. Here was the original task:\n\n${prompt}`;

    console.log(
      `${TAG} execWithRetry: attempt ${attempt}/${totalAttempts} for task "${taskKey}"${attempt > 1 ? " (resume)" : ""}`,
    );

    // Check if externally aborted (e.g. watchdog killed this slot)
    if (abortController?.signal?.aborted) {
      lastResult = {
        success: false,
        output: "",
        items: [],
        error: "Externally aborted (watchdog or manual kill)",
        sdk: sdk || "unknown",
        threadId: null,
      };
      break;
    }

    lastResult = await launchOrResumeThread(currentPrompt, cwd, timeoutMs, {
      taskKey,
      sdk,
      onEvent,
      abortController,
    });

    // Check if we should retry
    if (lastResult.success) {
      // If caller has custom shouldRetry (e.g. "output must contain 'PASS'"), check it
      if (typeof shouldRetry === "function" && shouldRetry(lastResult)) {
        console.log(
          `${TAG} attempt ${attempt} succeeded but shouldRetry returned true`,
        );
        continue;
      }
      return { ...lastResult, attempts: attempt };
    }

    // Failed — should we retry?
    if (attempt < totalAttempts) {
      if (typeof shouldRetry === "function" && !shouldRetry(lastResult)) {
        // Custom predicate says don't retry
        console.log(`${TAG} shouldRetry returned false — not retrying`);
        return { ...lastResult, attempts: attempt };
      }
      console.warn(
        `${TAG} attempt ${attempt} failed, will retry: ${lastResult.error}`,
      );
    }
  }

  return { ...lastResult, attempts: totalAttempts };
}

// ---------------------------------------------------------------------------
// Thread Management Exports
// ---------------------------------------------------------------------------

/**
 * Get the thread record for a task key.
 * @param {string} taskKey
 * @returns {ThreadRecord|null}
 */
export function getThreadRecord(taskKey) {
  return threadRegistry.get(taskKey) || null;
}

/**
 * Invalidate (kill) a thread record so it won't be resumed.
 * @param {string} taskKey
 */
export function invalidateThread(taskKey) {
  const record = threadRegistry.get(taskKey);
  if (record) {
    record.alive = false;
    threadRegistry.set(taskKey, record);
    saveThreadRegistry().catch(() => {});
  }
}

/**
 * Invalidate a thread and force a fresh start on next attempt.
 * Unlike invalidateThread which just sets alive=false, this also logs the reason.
 * @param {string} taskKey
 * @param {string} reason
 */
export function forceNewThread(taskKey, reason = "manual") {
  const record = threadRegistry.get(taskKey);
  if (record) {
    console.log(`${TAG} force-invalidating thread for task "${taskKey}": ${reason} (was turn ${record.turnCount})`);
    record.alive = false;
    threadRegistry.set(taskKey, record);
    saveThreadRegistry().catch(() => {});
  }
}

/**
 * Clear all thread records (e.g. on monitor restart).
 */
export function clearThreadRegistry() {
  threadRegistry.clear();
  saveThreadRegistry().catch(() => {});
}

/**
 * Prune all threads that have exceeded MAX_THREAD_TURNS or are older than THREAD_MAX_ABSOLUTE_AGE_MS.
 * Call on startup to clean up zombie threads from prior runs.
 * @returns {number} Number of threads pruned
 */
export function pruneAllExhaustedThreads() {
  let pruned = 0;
  const now = Date.now();
  for (const [key, record] of threadRegistry) {
    let reason = null;
    if (record.turnCount >= MAX_THREAD_TURNS) {
      reason = `${record.turnCount} turns (max ${MAX_THREAD_TURNS})`;
    } else if (now - record.createdAt > THREAD_MAX_ABSOLUTE_AGE_MS) {
      reason = `absolute age ${Math.round((now - record.createdAt) / 3600000)}h`;
    } else if (!record.alive) {
      reason = "already dead";
    }
    if (reason) {
      console.log(`${TAG} pruning thread for task "${key}": ${reason}`);
      record.alive = false;
      threadRegistry.set(key, record);
      pruned++;
    }
  }
  if (pruned > 0) {
    saveThreadRegistry().catch(() => {});
    console.log(`${TAG} pruned ${pruned} exhausted/stale threads`);
  }
  return pruned;
}

/**
 * Get summary of all active threads.
 * @returns {Array<{ taskKey: string, sdk: string, threadId: string|null, turnCount: number, age: number }>}
 */
export function getActiveThreads() {
  const now = Date.now();
  const result = [];
  for (const [key, record] of threadRegistry) {
    if (!record.alive) continue;
    if (now - record.lastUsedAt > THREAD_MAX_AGE_MS) continue;
    if (now - record.createdAt > THREAD_MAX_ABSOLUTE_AGE_MS) continue;
    if (record.turnCount >= MAX_THREAD_TURNS) continue;
    result.push({
      taskKey: key,
      sdk: record.sdk,
      threadId: record.threadId,
      turnCount: record.turnCount,
      age: now - record.createdAt,
    });
  }
  return result;
}
