/**
 * agent-pool.mjs — Shared Ephemeral Agent Pool
 *
 * WHY THIS EXISTS:
 * ────────────────
 * The primary Codex agent in monitor.mjs is a long-lived singleton thread.
 * Every operation that calls `execPrimaryPrompt` serialises behind that single
 * thread — task attempts, conflict resolution, follow-ups, and health-checks
 * all compete for the same lock.  Under load (or when a single prompt is
 * slow) this creates a bottleneck that stalls the entire monitor pipeline.
 *
 * This module provides **ephemeral, per-operation SDK threads** that spin up
 * on demand and tear down after a single prompt completes.  Each call gets its
 * own isolated Codex thread, so multiple operations can run concurrently
 * without blocking each other.
 *
 * The pattern is extracted from `sdk-conflict-resolver.mjs`'s
 * `launchFreshCodexThread`, but wrapped to match the `execPrimaryPrompt`
 * signature so it can be used as a drop-in replacement in monitor.mjs.
 *
 * EXPORTS:
 *   launchEphemeralThread(prompt, cwd, timeoutMs)
 *     → Low-level: spawns a fresh Codex SDK thread, runs one prompt,
 *       returns { success, output, error }.
 *
 *   execPooledPrompt(userMessage, options?)
 *     → High-level: matches the execPrimaryPrompt signature
 *       ({ finalResponse, items, usage }) so callers in monitor.mjs can
 *       swap in without changing surrounding code.
 */

import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

/** Repository root (two levels up from scripts/codex-monitor/) */
const REPO_ROOT = resolve(__dirname, "..", "..");

/** Default timeout: 15 minutes */
const DEFAULT_TIMEOUT_MS = 15 * 60 * 1000;

// ---------------------------------------------------------------------------
// Low-level: ephemeral thread
// ---------------------------------------------------------------------------

/**
 * Spin up a fresh, isolated Codex SDK thread, execute a single prompt, and
 * return the result.  The thread is not reused — it exists only for this one
 * operation, which means it cannot block (or be blocked by) any other thread.
 *
 * @param {string}  prompt     The prompt to send to the agent.
 * @param {string}  [cwd]      Working directory for the thread (defaults to REPO_ROOT).
 * @param {number}  [timeoutMs] Abort after this many milliseconds (default 15 min).
 * @param {object}  [extra]    Optional extras: { onEvent, abortController }.
 * @returns {Promise<{ success: boolean, output: string, items: Array, error: string|null }>}
 */
export async function launchEphemeralThread(
  prompt,
  cwd = REPO_ROOT,
  timeoutMs = DEFAULT_TIMEOUT_MS,
  extra = {},
) {
  const tag = "[agent-pool:ephemeral]";
  const { onEvent, abortController: externalAC } = extra;

  // ── 1. Load the SDK ──────────────────────────────────────────────────────
  let CodexClass;
  try {
    const mod = await import("@openai/codex-sdk");
    CodexClass = mod.Codex;
  } catch (err) {
    return {
      success: false,
      output: "",
      items: [],
      error: `Codex SDK not available: ${err.message}`,
    };
  }

  // ── 2. Create an ephemeral thread ────────────────────────────────────────
  const codex = new CodexClass();
  const threadOptions = {
    sandboxMode: "danger-full-access",
    workingDirectory: cwd,
    skipGitRepoCheck: true,
    approvalPolicy: "never",
  };
  const thread = codex.startThread(threadOptions);

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
      // Forward raw events to the caller when requested
      if (typeof onEvent === "function") {
        try {
          onEvent(event);
        } catch (_) {
          /* caller callback errors should not kill the stream */
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

    return { success: true, output, items: allItems, error: null };
  } catch (err) {
    clearTimeout(timer);

    if (err.name === "AbortError" || String(err) === "timeout") {
      return {
        success: false,
        output: "",
        items: [],
        error: `${tag} timeout after ${timeoutMs}ms`,
      };
    }

    return { success: false, output: "", items: [], error: err.message };
  }
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
 * @returns {Promise<{ finalResponse: string, items: Array, usage: object|null }>}
 */
export async function execPooledPrompt(userMessage, options = {}) {
  const {
    onEvent,
    timeoutMs = DEFAULT_TIMEOUT_MS,
    abortController,
    cwd = REPO_ROOT,
    // statusData and sendRawEvents are accepted but not used — keeps the
    // call-site compatible with execPrimaryPrompt without modification.
  } = options;

  const result = await launchEphemeralThread(userMessage, cwd, timeoutMs, {
    onEvent,
    abortController,
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
