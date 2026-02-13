/**
 * @module agent-hooks
 * @description Comprehensive agent lifecycle hooks system for the codex-monitor
 * orchestrator. Provides a configurable hook pipeline that fires at key points
 * in the agent task lifecycle (session start/stop, tool use, git operations,
 * PR creation, task completion).
 *
 * Hooks can be loaded from config files (.codex/hooks.json, .vscode/hooks.json,
 * codex-monitor.config.json) or registered programmatically. Each hook targets
 * one or more SDKs (codex, copilot, claude) and can be blocking or fire-and-forget.
 *
 * @example
 * import { loadHooks, executeHooks, registerBuiltinHooks } from "./agent-hooks.mjs";
 *
 * await loadHooks();          // Load from config files
 * registerBuiltinHooks();     // Register built-in quality gates
 *
 * const ctx = { taskId: "abc", branch: "ve/abc-fix-bug", sdk: "codex" };
 * await executeHooks("SessionStart", ctx);
 *
 * const result = await executeBlockingHooks("PrePush", ctx);
 * if (!result.passed) {
 *   console.error("Quality gates failed:", result.failures);
 * }
 */

import { spawnSync, spawn } from "node:child_process";
import { readFileSync, existsSync } from "node:fs";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import { randomUUID } from "node:crypto";

// ── Constants ───────────────────────────────────────────────────────────────

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

/** Repository root (two levels up from scripts/codex-monitor/) */
const REPO_ROOT = resolve(__dirname, "..", "..");

/** Console log prefix */
export const TAG = "[agent-hooks]";

/** Default timeout for hook execution (60 seconds) */
const DEFAULT_TIMEOUT_MS = 60_000;

/** Maximum output captured per hook (64 KB) */
const MAX_OUTPUT_BYTES = 64 * 1024;

/** Whether we're running on Windows */
const IS_WINDOWS = process.platform === "win32";

/**
 * Valid hook event names matching VS Code / Claude Code naming conventions.
 * @type {readonly string[]}
 */
export const HOOK_EVENTS = Object.freeze([
  "SessionStart",
  "SessionStop",
  "PreToolUse",
  "PostToolUse",
  "SubagentStart",
  "SubagentStop",
  "PrePush",
  "PostPush",
  "PreCommit",
  "PostCommit",
  "PrePR",
  "PostPR",
  "TaskComplete",
]);

/**
 * Canonical SDK names.
 * @type {readonly string[]}
 */
const VALID_SDKS = Object.freeze(["codex", "copilot", "claude"]);

/**
 * Wildcard indicating a hook applies to all SDKs.
 * @type {string}
 */
const SDK_WILDCARD = "*";

// ── Types (JSDoc) ───────────────────────────────────────────────────────────

/**
 * @typedef {Object} HookDefinition
 * @property {string}   id          - Unique identifier (auto-generated if omitted)
 * @property {string}   command     - Shell command to execute
 * @property {string}   [description] - Human-readable description
 * @property {number}   [timeout]   - Timeout in milliseconds (default: 60000)
 * @property {boolean}  [blocking]  - If true, failure stops the pipeline (default: false)
 * @property {string[]} [sdks]      - SDK filter: ["codex"], ["copilot","claude"], or ["*"] (default: ["*"])
 * @property {Record<string,string>} [env] - Additional environment variables
 * @property {boolean}  [builtin]   - Whether this is a built-in hook (not from config)
 */

/**
 * @typedef {Object} HookContext
 * @property {string}  [taskId]       - Current task ID
 * @property {string}  [taskTitle]    - Current task title
 * @property {string}  [branch]       - Branch name
 * @property {string}  [worktreePath] - Worktree path
 * @property {string}  [sdk]          - Active SDK name (codex/copilot/claude)
 * @property {string}  [event]        - Hook event name (set automatically)
 * @property {number}  [timestamp]    - Execution timestamp (set automatically)
 * @property {string}  [repoRoot]     - Repository root path
 * @property {Record<string,string>} [extra] - Additional context values
 */

/**
 * @typedef {Object} HookResult
 * @property {string}  id          - Hook ID
 * @property {string}  command     - Command that was executed
 * @property {boolean} success     - Whether the hook succeeded (exit code 0)
 * @property {number}  exitCode    - Process exit code (-1 on timeout/error)
 * @property {string}  stdout      - Captured stdout (truncated)
 * @property {string}  stderr      - Captured stderr (truncated)
 * @property {number}  durationMs  - Execution duration in milliseconds
 * @property {string}  [error]     - Error message if hook failed to execute
 */

/**
 * @typedef {Object} BlockingHookResult
 * @property {boolean}      passed   - True if all blocking hooks succeeded
 * @property {HookResult[]} results  - Results from all executed hooks
 * @property {HookResult[]} failures - Only the hooks that failed
 */

// ── Hook Registry ───────────────────────────────────────────────────────────

/**
 * Internal registry: event name → array of hook definitions.
 * @type {Map<string, HookDefinition[]>}
 */
const _registry = new Map();

/**
 * Initialise registry with empty arrays for each valid event.
 */
function _initRegistry() {
  for (const event of HOOK_EVENTS) {
    if (!_registry.has(event)) {
      _registry.set(event, []);
    }
  }
}

/**
 * Reset the hook registry to empty state. Useful for testing.
 */
export function resetHooks() {
  _registry.clear();
  _initRegistry();
}

// Ensure the registry is initialised on module load.
_initRegistry();

// ── Config Loading ──────────────────────────────────────────────────────────

/**
 * Default config file search paths, resolved relative to the repo root.
 * Searched in order; first existing file wins.
 * @type {string[]}
 */
const CONFIG_SEARCH_PATHS = [
  ".codex/hooks.json",
  ".vscode/hooks.json",
  "scripts/codex-monitor/codex-monitor.config.json",
];

/**
 * Load hook definitions from a JSON config file.
 *
 * The file should contain a top-level `hooks` object whose keys are event names
 * and values are arrays of {@link HookDefinition} objects.
 *
 * If no `configPath` is provided, the function searches the default locations
 * (`.codex/hooks.json`, `.vscode/hooks.json`, `codex-monitor.config.json`).
 * Hooks loaded from config are merged with (appended to) any programmatically
 * registered hooks.
 *
 * @param {string} [configPath] - Absolute or repo-relative path to a hooks config file
 * @returns {number} Number of hooks loaded
 *
 * @example
 * loadHooks();                           // Search default paths
 * loadHooks(".codex/hooks.json");        // Explicit path
 */
export function loadHooks(configPath) {
  /** @type {string|null} */
  let resolvedPath = null;

  if (configPath) {
    resolvedPath = resolve(REPO_ROOT, configPath);
    if (!existsSync(resolvedPath)) {
      console.warn(`${TAG} config file not found: ${resolvedPath}`);
      return 0;
    }
  } else {
    // Search default paths
    for (const relPath of CONFIG_SEARCH_PATHS) {
      const candidate = resolve(REPO_ROOT, relPath);
      if (existsSync(candidate)) {
        resolvedPath = candidate;
        break;
      }
    }
    if (!resolvedPath) {
      console.log(
        `${TAG} no hook config file found — using built-in hooks only`,
      );
      return 0;
    }
  }

  let raw;
  try {
    raw = readFileSync(resolvedPath, "utf8");
  } catch (err) {
    console.error(
      `${TAG} failed to read config file: ${resolvedPath}`,
      err.message,
    );
    return 0;
  }

  let config;
  try {
    config = JSON.parse(raw);
  } catch (err) {
    console.error(
      `${TAG} invalid JSON in config file: ${resolvedPath}`,
      err.message,
    );
    return 0;
  }

  // Support both top-level { hooks: { ... } } and nested inside codex-monitor config
  const hooksDef = config.hooks ?? config.agentHooks ?? null;
  if (!hooksDef || typeof hooksDef !== "object") {
    console.log(`${TAG} no "hooks" or "agentHooks" key in ${resolvedPath}`);
    return 0;
  }

  let loaded = 0;

  for (const [event, defs] of Object.entries(hooksDef)) {
    if (!HOOK_EVENTS.includes(event)) {
      console.warn(`${TAG} ignoring unknown hook event "${event}" in config`);
      continue;
    }

    const hookArray = Array.isArray(defs) ? defs : [defs];
    for (const def of hookArray) {
      if (!def.command) {
        console.warn(
          `${TAG} skipping hook for "${event}" — missing "command" field`,
        );
        continue;
      }

      const hookDef = _normalizeHookDef(def);
      registerHook(event, hookDef);
      loaded++;
    }
  }

  console.log(`${TAG} loaded ${loaded} hook(s) from ${resolvedPath}`);
  return loaded;
}

// ── Registration ────────────────────────────────────────────────────────────

/**
 * Register a hook for a specific event.
 *
 * @param {string} event - One of {@link HOOK_EVENTS}
 * @param {HookDefinition} hookDef - Hook definition
 * @returns {string} The hook's unique ID
 * @throws {Error} If the event name is invalid
 *
 * @example
 * const id = registerHook("PrePush", {
 *   command: "scripts/agent-preflight.ps1",
 *   blocking: true,
 *   timeout: 300000,
 * });
 */
export function registerHook(event, hookDef) {
  if (!HOOK_EVENTS.includes(event)) {
    throw new Error(
      `${TAG} invalid hook event: "${event}". Valid events: ${HOOK_EVENTS.join(", ")}`,
    );
  }

  const normalized = _normalizeHookDef(hookDef);

  if (!_registry.has(event)) {
    _registry.set(event, []);
  }

  // Prevent duplicate registration by ID
  const existing = _registry.get(event);
  const idx = existing.findIndex((h) => h.id === normalized.id);
  if (idx >= 0) {
    existing[idx] = normalized;
    console.log(`${TAG} updated hook "${normalized.id}" for event "${event}"`);
  } else {
    existing.push(normalized);
    console.log(
      `${TAG} registered hook "${normalized.id}" for event "${event}"` +
        (normalized.blocking ? " (blocking)" : ""),
    );
  }

  return normalized.id;
}

/**
 * Remove a previously registered hook by event and ID.
 *
 * @param {string} event - Hook event name
 * @param {string} id - Hook ID to remove
 * @returns {boolean} True if the hook was found and removed
 */
export function unregisterHook(event, id) {
  if (!_registry.has(event)) return false;

  const hooks = _registry.get(event);
  const idx = hooks.findIndex((h) => h.id === id);
  if (idx < 0) return false;

  hooks.splice(idx, 1);
  console.log(`${TAG} unregistered hook "${id}" from event "${event}"`);
  return true;
}

/**
 * Get all registered hooks, optionally filtered by event.
 *
 * @param {string} [event] - If provided, only return hooks for this event
 * @returns {Record<string, HookDefinition[]>|HookDefinition[]} All hooks or hooks for one event
 */
export function getRegisteredHooks(event) {
  if (event) {
    if (!HOOK_EVENTS.includes(event)) {
      throw new Error(`${TAG} invalid hook event: "${event}"`);
    }
    return [...(_registry.get(event) ?? [])];
  }

  /** @type {Record<string, HookDefinition[]>} */
  const result = {};
  for (const [ev, hooks] of _registry.entries()) {
    if (hooks.length > 0) {
      result[ev] = [...hooks];
    }
  }
  return result;
}

// ── Hook Execution ──────────────────────────────────────────────────────────

/**
 * Execute all hooks registered for an event (blocking and non-blocking).
 *
 * Blocking hooks run sequentially and their results are awaited.
 * Non-blocking hooks run in parallel (fire-and-forget) — errors are logged but
 * do not affect the return value.
 *
 * @param {string} event - Hook event name
 * @param {HookContext} context - Execution context
 * @returns {Promise<HookResult[]>} Results from all executed hooks
 */
export async function executeHooks(event, context = {}) {
  if (!HOOK_EVENTS.includes(event)) {
    console.warn(`${TAG} executeHooks called with unknown event: "${event}"`);
    return [];
  }

  const hooks = _getFilteredHooks(event, context.sdk);
  if (hooks.length === 0) return [];

  const enrichedCtx = _enrichContext(event, context);
  const env = _buildEnv(enrichedCtx);

  /** @type {HookResult[]} */
  const results = [];

  // Separate blocking and non-blocking hooks
  const blocking = hooks.filter((h) => h.blocking);
  const nonBlocking = hooks.filter((h) => !h.blocking);

  // Run blocking hooks sequentially
  for (const hook of blocking) {
    const hookEnv = { ...env, ..._normalizeEnvValues(hook.env) };
    const result = _executeHookSync(hook, enrichedCtx, hookEnv);
    results.push(result);

    if (!result.success) {
      console.error(
        `${TAG} blocking hook "${hook.id}" failed for event "${event}" ` +
          `(exit ${result.exitCode}, ${result.durationMs}ms)`,
      );
      if (result.stderr) {
        console.error(`${TAG}   stderr: ${_truncate(result.stderr, 500)}`);
      }
    } else {
      console.log(
        `${TAG} blocking hook "${hook.id}" passed for event "${event}" (${result.durationMs}ms)`,
      );
    }
  }

  // Run non-blocking hooks in parallel (fire-and-forget)
  const nonBlockingPromises = nonBlocking.map(async (hook) => {
    const hookEnv = { ...env, ..._normalizeEnvValues(hook.env) };
    try {
      const result = await _executeHookAsync(hook, enrichedCtx, hookEnv);
      results.push(result);

      if (!result.success) {
        console.warn(
          `${TAG} non-blocking hook "${hook.id}" failed for event "${event}" ` +
            `(exit ${result.exitCode})`,
        );
      }
    } catch (err) {
      console.warn(
        `${TAG} non-blocking hook "${hook.id}" threw: ${err.message}`,
      );
      results.push({
        id: hook.id,
        command: hook.command,
        success: false,
        exitCode: -1,
        stdout: "",
        stderr: "",
        durationMs: 0,
        error: err.message,
      });
    }
  });

  // Wait for non-blocking hooks but don't let them block indefinitely
  await Promise.allSettled(nonBlockingPromises);

  return results;
}

/**
 * Execute only the **blocking** hooks for an event and return a pass/fail result.
 *
 * This is the method to call before critical operations (push, commit, PR) where
 * all quality gates must pass.
 *
 * @param {string} event - Hook event name
 * @param {HookContext} context - Execution context
 * @returns {Promise<BlockingHookResult>} Aggregated pass/fail with details
 *
 * @example
 * const result = await executeBlockingHooks("PrePush", { taskId, branch, sdk: "codex" });
 * if (!result.passed) {
 *   console.error("Blocked:", result.failures.map(f => f.error || f.stderr));
 * }
 */
export async function executeBlockingHooks(event, context = {}) {
  if (!HOOK_EVENTS.includes(event)) {
    console.warn(
      `${TAG} executeBlockingHooks called with unknown event: "${event}"`,
    );
    return { passed: true, results: [], failures: [] };
  }

  const hooks = _getFilteredHooks(event, context.sdk).filter((h) => h.blocking);
  if (hooks.length === 0) {
    return { passed: true, results: [], failures: [] };
  }

  const enrichedCtx = _enrichContext(event, context);
  const env = _buildEnv(enrichedCtx);

  /** @type {HookResult[]} */
  const results = [];
  /** @type {HookResult[]} */
  const failures = [];

  for (const hook of hooks) {
    const hookEnv = { ...env, ..._normalizeEnvValues(hook.env) };
    const result = _executeHookSync(hook, enrichedCtx, hookEnv);
    results.push(result);

    if (!result.success) {
      failures.push(result);
      console.error(
        `${TAG} BLOCKING FAILURE: hook "${hook.id}" for event "${event}" — ` +
          `exit ${result.exitCode} (${result.durationMs}ms)`,
      );
    } else {
      console.log(
        `${TAG} blocking hook "${hook.id}" passed (${result.durationMs}ms)`,
      );
    }
  }

  const passed = failures.length === 0;

  if (passed) {
    console.log(
      `${TAG} all ${results.length} blocking hook(s) passed for "${event}"`,
    );
  } else {
    console.error(
      `${TAG} ${failures.length}/${results.length} blocking hook(s) FAILED for "${event}"`,
    );
  }

  return { passed, results, failures };
}

// ── Built-in Hooks ──────────────────────────────────────────────────────────

/**
 * Register the default built-in hooks. These provide essential quality gates
 * that run regardless of config file contents.
 *
 * Built-in hooks:
 *   - **PrePush** — Runs `scripts/agent-preflight.ps1` (Windows) or
 *     `scripts/agent-preflight.sh` (Unix) to validate quality gates.
 *   - **TaskComplete** — Runs a basic acceptance-criteria check via git log.
 */
export function registerBuiltinHooks() {
  // ── PrePush: agent preflight quality gate ──
  const preflightScript = IS_WINDOWS
    ? "powershell -NoProfile -ExecutionPolicy Bypass -File scripts/agent-preflight.ps1"
    : "bash scripts/agent-preflight.sh";

  registerHook("PrePush", {
    id: "builtin-prepush-preflight",
    command: preflightScript,
    description: "Run agent preflight quality gates before push",
    timeout: 300_000, // 5 minutes
    blocking: true,
    sdks: [SDK_WILDCARD],
    builtin: true,
  });

  // ── TaskComplete: verify at least one commit exists on the branch ──
  registerHook("TaskComplete", {
    id: "builtin-task-complete-validation",
    command: _buildTaskCompleteCommand(),
    description:
      "Validate task produced at least one commit ahead of base branch",
    timeout: 30_000, // 30 seconds
    blocking: true,
    sdks: [SDK_WILDCARD],
    builtin: true,
  });

  console.log(`${TAG} built-in hooks registered`);
}

/**
 * Build the shell command for the TaskComplete validation hook.
 * Checks that HEAD has at least one commit ahead of the merge-base with main.
 *
 * @returns {string}
 */
function _buildTaskCompleteCommand() {
  if (IS_WINDOWS) {
    // PowerShell one-liner: count commits ahead of main
    return [
      "powershell -NoProfile -Command",
      '"$ahead = git rev-list --count $(git merge-base HEAD origin/main)..HEAD;',
      "if ([int]$ahead -lt 1) { Write-Error 'No commits ahead of origin/main'; exit 1 }",
      'else { Write-Host "OK: $ahead commit(s) ahead of origin/main" }"',
    ].join(" ");
  }

  // Bash one-liner
  return [
    "bash -c",
    "'ahead=$(git rev-list --count $(git merge-base HEAD origin/main)..HEAD);",
    'if [ "$ahead" -lt 1 ]; then echo "No commits ahead of origin/main" >&2; exit 1;',
    'else echo "OK: $ahead commit(s) ahead of origin/main"; fi\'',
  ].join(" ");
}

// ── Internal: Filtering ─────────────────────────────────────────────────────

/**
 * Get hooks for an event, filtered by the active SDK.
 *
 * @param {string} event - Hook event name
 * @param {string} [sdk] - Active SDK name; if omitted, all hooks are returned
 * @returns {HookDefinition[]}
 */
function _getFilteredHooks(event, sdk) {
  const hooks = _registry.get(event) ?? [];
  if (!sdk) return [...hooks];

  const normalizedSdk = sdk.toLowerCase();
  return hooks.filter((hook) => {
    const sdks = hook.sdks ?? [SDK_WILDCARD];
    return sdks.includes(SDK_WILDCARD) || sdks.includes(normalizedSdk);
  });
}

// ── Internal: Context & Environment ─────────────────────────────────────────

/**
 * Enrich a context object with defaults (event, timestamp, repoRoot).
 *
 * @param {string} event
 * @param {HookContext} context
 * @returns {HookContext}
 */
function _enrichContext(event, context) {
  return {
    repoRoot: REPO_ROOT,
    ...context,
    event,
    timestamp: Date.now(),
  };
}

/**
 * Build the environment variables map that is passed to every hook subprocess.
 *
 * @param {HookContext} ctx - Enriched hook context
 * @returns {Record<string, string>}
 */
function _buildEnv(ctx) {
  /** @type {Record<string, string>} */
  const env = {
    ...process.env,
    VE_HOOK_EVENT: ctx.event ?? "",
    VE_TASK_ID: ctx.taskId ?? "",
    VE_TASK_TITLE: ctx.taskTitle ?? "",
    VE_BRANCH_NAME: ctx.branch ?? "",
    VE_WORKTREE_PATH: ctx.worktreePath ?? "",
    VE_SDK: ctx.sdk ?? "",
    VE_REPO_ROOT: ctx.repoRoot ?? REPO_ROOT,
    VE_HOOK_BLOCKING: "false", // Overridden per-hook in execution
  };

  // Merge any extra context values as env vars
  if (ctx.extra && typeof ctx.extra === "object") {
    for (const [key, val] of Object.entries(ctx.extra)) {
      env[`VE_HOOK_${key.toUpperCase()}`] = String(val ?? "");
    }
  }

  return env;
}

// ── Internal: Synchronous Hook Execution ────────────────────────────────────

/**
 * Execute a hook synchronously using `spawnSync`. Used for blocking hooks.
 *
 * @param {HookDefinition} hook
 * @param {HookContext} ctx
 * @param {Record<string, string>} env
 * @returns {HookResult}
 */
function _executeHookSync(hook, ctx, env) {
  const start = Date.now();
  const timeout = hook.timeout ?? DEFAULT_TIMEOUT_MS;
  const cwd = ctx.worktreePath || ctx.repoRoot || REPO_ROOT;

  const hookEnv = {
    ...env,
    VE_HOOK_BLOCKING: "true",
  };

  try {
    const result = spawnSync(hook.command, {
      cwd,
      env: hookEnv,
      encoding: "utf8",
      timeout,
      shell: true,
      windowsHide: true,
      maxBuffer: MAX_OUTPUT_BYTES,
    });

    const durationMs = Date.now() - start;
    const exitCode = result.status ?? -1;

    // Handle timeout (spawnSync sets signal to SIGTERM on timeout)
    if (result.signal === "SIGTERM" || result.error?.code === "ETIMEDOUT") {
      return {
        id: hook.id,
        command: hook.command,
        success: false,
        exitCode: -1,
        stdout: _truncate(result.stdout ?? "", MAX_OUTPUT_BYTES),
        stderr: _truncate(result.stderr ?? "", MAX_OUTPUT_BYTES),
        durationMs,
        error: `Hook timed out after ${timeout}ms`,
      };
    }

    return {
      id: hook.id,
      command: hook.command,
      success: exitCode === 0,
      exitCode,
      stdout: _truncate(result.stdout ?? "", MAX_OUTPUT_BYTES),
      stderr: _truncate(result.stderr ?? "", MAX_OUTPUT_BYTES),
      durationMs,
      error: result.error ? result.error.message : undefined,
    };
  } catch (err) {
    return {
      id: hook.id,
      command: hook.command,
      success: false,
      exitCode: -1,
      stdout: "",
      stderr: "",
      durationMs: Date.now() - start,
      error: `Failed to spawn hook: ${err.message}`,
    };
  }
}

// ── Internal: Asynchronous Hook Execution ───────────────────────────────────

/**
 * Execute a hook asynchronously using `spawn`. Used for non-blocking hooks.
 * Returns a Promise that resolves when the process exits or times out.
 *
 * @param {HookDefinition} hook
 * @param {HookContext} ctx
 * @param {Record<string, string>} env
 * @returns {Promise<HookResult>}
 */
function _executeHookAsync(hook, ctx, env) {
  return new Promise((resolvePromise) => {
    const start = Date.now();
    const timeout = hook.timeout ?? DEFAULT_TIMEOUT_MS;
    const cwd = ctx.worktreePath || ctx.repoRoot || REPO_ROOT;

    const hookEnv = {
      ...env,
      VE_HOOK_BLOCKING: "false",
    };

    /** @type {string[]} */
    const stdoutChunks = [];
    /** @type {string[]} */
    const stderrChunks = [];
    let totalBytes = 0;
    let settled = false;

    /** @param {HookResult} result */
    function settle(result) {
      if (settled) return;
      settled = true;
      resolvePromise(result);
    }

    let child;
    try {
      child = spawn(hook.command, {
        cwd,
        env: hookEnv,
        shell: true,
        windowsHide: true,
        stdio: ["ignore", "pipe", "pipe"],
      });
    } catch (err) {
      settle({
        id: hook.id,
        command: hook.command,
        success: false,
        exitCode: -1,
        stdout: "",
        stderr: "",
        durationMs: Date.now() - start,
        error: `Failed to spawn hook: ${err.message}`,
      });
      return;
    }

    // Capture stdout
    child.stdout?.on("data", (chunk) => {
      if (totalBytes < MAX_OUTPUT_BYTES) {
        stdoutChunks.push(chunk.toString("utf8"));
        totalBytes += chunk.length;
      }
    });

    // Capture stderr
    child.stderr?.on("data", (chunk) => {
      if (totalBytes < MAX_OUTPUT_BYTES) {
        stderrChunks.push(chunk.toString("utf8"));
        totalBytes += chunk.length;
      }
    });

    // Timeout handler
    const timer = setTimeout(() => {
      try {
        child.kill("SIGTERM");
      } catch {
        // Process may have already exited
      }
      settle({
        id: hook.id,
        command: hook.command,
        success: false,
        exitCode: -1,
        stdout: stdoutChunks.join(""),
        stderr: stderrChunks.join(""),
        durationMs: Date.now() - start,
        error: `Hook timed out after ${timeout}ms`,
      });
    }, timeout);

    child.on("close", (code) => {
      clearTimeout(timer);
      settle({
        id: hook.id,
        command: hook.command,
        success: code === 0,
        exitCode: code ?? -1,
        stdout: stdoutChunks.join(""),
        stderr: stderrChunks.join(""),
        durationMs: Date.now() - start,
      });
    });

    child.on("error", (err) => {
      clearTimeout(timer);
      settle({
        id: hook.id,
        command: hook.command,
        success: false,
        exitCode: -1,
        stdout: stdoutChunks.join(""),
        stderr: stderrChunks.join(""),
        durationMs: Date.now() - start,
        error: err.message,
      });
    });
  });
}

// ── Internal: Normalisation Helpers ─────────────────────────────────────────

/**
 * Normalise a raw hook definition object, filling in defaults.
 *
 * @param {Partial<HookDefinition>} def
 * @returns {HookDefinition}
 */
function _normalizeHookDef(def) {
  const id = def.id ?? `hook-${randomUUID().slice(0, 8)}`;
  const sdks = _normalizeSdks(def.sdks);

  return {
    id,
    command: String(def.command ?? ""),
    description: def.description ?? "",
    timeout:
      typeof def.timeout === "number" && def.timeout > 0
        ? def.timeout
        : DEFAULT_TIMEOUT_MS,
    blocking: Boolean(def.blocking),
    sdks,
    env: def.env && typeof def.env === "object" ? { ...def.env } : {},
    builtin: Boolean(def.builtin),
  };
}

/**
 * Normalise the SDKs array from a hook definition.
 *
 * @param {unknown} sdks
 * @returns {string[]}
 */
function _normalizeSdks(sdks) {
  if (!sdks || !Array.isArray(sdks) || sdks.length === 0) {
    return [SDK_WILDCARD];
  }

  const normalised = sdks
    .map((s) => String(s).toLowerCase().trim())
    .filter((s) => s === SDK_WILDCARD || VALID_SDKS.includes(s));

  if (normalised.length === 0) return [SDK_WILDCARD];
  if (normalised.includes(SDK_WILDCARD)) return [SDK_WILDCARD];
  return [...new Set(normalised)];
}

/**
 * Ensure all env values are strings (non-string values are coerced).
 *
 * @param {Record<string, unknown>} [env]
 * @returns {Record<string, string>}
 */
function _normalizeEnvValues(env) {
  if (!env || typeof env !== "object") return {};
  /** @type {Record<string, string>} */
  const result = {};
  for (const [key, val] of Object.entries(env)) {
    result[key] = String(val ?? "");
  }
  return result;
}

/**
 * Truncate a string to a maximum length, appending an ellipsis marker if truncated.
 *
 * @param {string} str
 * @param {number} maxLen
 * @returns {string}
 */
function _truncate(str, maxLen) {
  if (!str || str.length <= maxLen) return str ?? "";
  return str.slice(0, maxLen) + "\n... (truncated)";
}
