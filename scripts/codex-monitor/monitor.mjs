import { execSync, spawn, spawnSync } from "node:child_process";
import { existsSync, readFileSync, watch, writeFileSync } from "node:fs";
import {
  copyFile,
  mkdir,
  readFile,
  rename,
  unlink,
  writeFile,
} from "node:fs/promises";
import net from "node:net";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { acquireMonitorLock, runMaintenanceSweep } from "./maintenance.mjs";
import {
  attemptAutoFix,
  fixLoopingError,
  isDevMode,
  runCodexExec,
} from "./autofix.mjs";
import {
  startTelegramBot,
  stopTelegramBot,
  injectMonitorFunctions,
  notify,
  restoreLiveDigest,
} from "./telegram-bot.mjs";
import {
  execPrimaryPrompt,
  isPrimaryBusy,
  initPrimaryAgent,
} from "./primary-agent.mjs";
import { loadConfig } from "./config.mjs";
import { formatPreflightReport, runPreflightChecks } from "./preflight.mjs";
import { startAutoUpdateLoop, stopAutoUpdateLoop } from "./update-check.mjs";
import { ensureCodexConfig, printConfigSummary } from "./codex-config.mjs";
import { RestartController } from "./restart-controller.mjs";
import {
  analyzeMergeStrategy,
  resetMergeStrategyDedup,
} from "./merge-strategy.mjs";
import {
  normalizeDedupKey,
  stripAnsi,
  isErrorLine,
  escapeHtml,
  formatHtmlLink,
  getErrorFingerprint,
  getMaxParallelFromArgs,
  parsePrNumberFromUrl,
} from "./utils.mjs";

const __dirname = resolve(fileURLToPath(new URL(".", import.meta.url)));

// ── Load unified configuration ──────────────────────────────────────────────
let config = loadConfig();

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

let {
  projectName,
  scriptPath,
  scriptArgs,
  restartDelayMs,
  maxRestarts,
  logDir,
  logMaxSizeMb,
  logCleanupIntervalMin,
  watchEnabled,
  watchPath: configWatchPath,
  echoLogs,
  autoFixEnabled,
  preflightEnabled: configPreflightEnabled,
  preflightRetryMs: configPreflightRetryMs,
  repoRoot,
  statusPath,
  telegramPollLockPath,
  telegramToken,
  telegramChatId,
  telegramIntervalMin,
  telegramCommandPollTimeoutSec,
  telegramCommandConcurrency,
  telegramCommandMaxBatch,
  telegramBotEnabled,
  telegramCommandEnabled,
  repoSlug,
  repoUrlBase,
  vkRecoveryPort,
  vkRecoveryHost,
  vkEndpointUrl,
  vkPublicUrl,
  vkTaskUrlTemplate,
  vkRecoveryCooldownMin,
  vkSpawnEnabled,
  vkEnsureIntervalMs,
  plannerPerCapitaThreshold,
  plannerIdleSlotThreshold,
  plannerDedupMs,
  plannerMode: configPlannerMode,
  agentPrompts,
  scheduler: executorScheduler,
  agentSdk,
  envPaths,
  dependabotAutoMerge,
  dependabotAutoMergeIntervalMin,
  dependabotMergeMethod,
  dependabotAuthors,
  telegramVerbosity,
} = config;

void initPrimaryAgent(config);

let watchPath = resolve(configWatchPath);
let codexEnabled = config.codexEnabled;
let plannerMode = configPlannerMode; // "codex-sdk" | "kanban" | "disabled"
console.log(`[monitor] task planner mode: ${plannerMode}`);
let codexDisabledReason = codexEnabled
  ? ""
  : process.env.CODEX_SDK_DISABLED === "1"
    ? "disabled via CODEX_SDK_DISABLED"
    : agentSdk?.primary && agentSdk.primary !== "codex"
      ? `disabled via agent_sdk.primary=${agentSdk.primary}`
      : "disabled via --no-codex";
let preflightEnabled = configPreflightEnabled;
let preflightRetryMs = configPreflightRetryMs;

// Merge strategy: Codex-powered merge decision analysis
// Enabled by default unless CODEX_ANALYZE_MERGE_STRATEGY=false
const codexAnalyzeMergeStrategy =
  codexEnabled &&
  (process.env.CODEX_ANALYZE_MERGE_STRATEGY || "").toLowerCase() !== "false";
const mergeStrategyMode = String(
  process.env.MERGE_STRATEGY_MODE || "smart",
).toLowerCase();
const codexResolveConflictsEnabled =
  codexEnabled && mergeStrategyMode.includes("codexsdk");
const conflictResolutionTimeoutMs = Number(
  process.env.MERGE_CONFLICT_RESOLUTION_TIMEOUT_MS || "600000",
);
// When telegram-bot.mjs is active it owns getUpdates — monitor must NOT poll
// to avoid HTTP 409 "Conflict: terminated by other getUpdates request".
let telegramPollLockHeld = false;
let preflightInProgress = false;
let preflightLastResult = null;
let preflightLastRunAt = 0;
let preflightRetryTimer = null;

let CodexClient = null;

let restartCount = 0;
let shuttingDown = false;
let currentChild = null;
let pendingRestart = false;
let skipNextAnalyze = false;
let skipNextRestartCount = false;

// Cached VK repo ID (lazy loaded on first PR/rebase call)
let cachedRepoId = null;
// Cached VK project ID (lazy loaded)
let cachedProjectId = null;
let watcher = null;
let watcherDebounce = null;
let watchFileName = null;
let envWatchers = [];
let envWatcherDebounce = null;

// ── Self-restart: exit code 75 signals cli.mjs to re-fork with fresh ESM cache
const SELF_RESTART_EXIT_CODE = 75;
let selfWatcher = null;
let selfWatcherDebounce = null;
let pendingSelfRestart = null; // filename that triggered a deferred restart

// ── Self-restart marker: detect if this process was spawned by a code-change restart
const selfRestartMarkerPath = resolve(
  config.cacheDir || resolve(config.repoRoot, ".cache"),
  "ve-self-restart.marker",
);
let isSelfRestart = false;
try {
  if (existsSync(selfRestartMarkerPath)) {
    const ts = Number(
      (await import("node:fs")).readFileSync(selfRestartMarkerPath, "utf8"),
    );
    // Marker is valid if written within the last 30 seconds
    if (Date.now() - ts < 30_000) {
      isSelfRestart = true;
      console.log(
        "[monitor] detected self-restart marker — suppressing startup notifications",
      );
    }
    // Clean up marker regardless
    try {
      (await import("node:fs")).unlinkSync(selfRestartMarkerPath);
    } catch {
      /* best effort */
    }
  }
} catch {
  /* first start or missing file */
}

let telegramNotifierInterval = null;
let telegramNotifierTimeout = null;
let vkRecoveryLastAt = 0;
let vkNonJsonNotifiedAt = 0;
let vibeKanbanProcess = null;
let vibeKanbanStartedAt = 0;
const smartPrAllowRecreateClosed =
  process.env.VE_SMARTPR_ALLOW_RECREATE_CLOSED === "1";
const githubToken =
  process.env.GITHUB_TOKEN ||
  process.env.GH_TOKEN ||
  process.env.GITHUB_PAT ||
  process.env.GITHUB_PAT_TOKEN ||
  "";
let monitorFailureHandling = false;
const monitorFailureTimestamps = [];
const monitorFailureWindowMs = 10 * 60 * 1000;
const monitorRestartCooldownMs = 60 * 1000;
let lastMonitorRestartAt = 0;
const orchestratorRestartTimestamps = [];
const orchestratorRestartWindowMs = 5 * 60 * 1000;
const orchestratorRestartThreshold = 8;
const orchestratorPauseMs = 10 * 60 * 1000;
let orchestratorHaltedUntil = 0;
let orchestratorLoopFixInProgress = false;
let monitorSafeModeUntil = 0;
let orchestratorResumeTimer = null;

// ── Mutex / restart-loop prevention ─────────────────────────────────────────
// When the orchestrator exits because "Another orchestrator instance is already
// running" (mutex held), the monitor must NOT restart immediately — the old
// instance still has the mutex and a tight restart loop will form.
const restartController = new RestartController();

let logRemainder = "";
const mergeNotified = new Set();
const pendingMerges = new Set();
const errorNotified = new Map();
const mergeFailureNotified = new Map();
const vkErrorNotified = new Map();
const telegramDedup = new Map();

// ── Deduplication tracking (utilities imported from utils.mjs) ───────────────

// ── Internal crash loop circuit breaker ──────────────────────────────────────
// Detects rapid failure bursts independently of Telegram dedup.
// When tripped, kills the orchestrator child and pauses everything.
const CIRCUIT_BREAKER_WINDOW_MS = 60_000; // 1 minute
const CIRCUIT_BREAKER_THRESHOLD = 5; // 5 failures in window = circuit trips
const CIRCUIT_BREAKER_PAUSE_MS = 5 * 60_000; // 5-minute hard pause
let circuitBreakerTripped = false;
let circuitBreakerResetAt = 0;
let circuitBreakerNotified = false;
const circuitBreakerTimestamps = [];

function recordCircuitBreakerEvent() {
  const now = Date.now();
  circuitBreakerTimestamps.push(now);
  // Prune events outside window
  while (
    circuitBreakerTimestamps.length &&
    now - circuitBreakerTimestamps[0] > CIRCUIT_BREAKER_WINDOW_MS
  ) {
    circuitBreakerTimestamps.shift();
  }
  return circuitBreakerTimestamps.length;
}

function isCircuitBreakerTripped() {
  const now = Date.now();
  // If paused, check if pause expired
  if (circuitBreakerTripped && now >= circuitBreakerResetAt) {
    circuitBreakerTripped = false;
    circuitBreakerNotified = false;
    circuitBreakerTimestamps.length = 0;
    console.warn("[monitor] circuit breaker reset — resuming normal operation");
    return false;
  }
  return circuitBreakerTripped;
}

function tripCircuitBreaker(failureCount) {
  if (circuitBreakerTripped) return; // already tripped
  circuitBreakerTripped = true;
  circuitBreakerResetAt = Date.now() + CIRCUIT_BREAKER_PAUSE_MS;
  const pauseMin = Math.round(CIRCUIT_BREAKER_PAUSE_MS / 60_000);
  console.error(
    `[monitor] 🔌 CIRCUIT BREAKER TRIPPED: ${failureCount} failures in ${Math.round(CIRCUIT_BREAKER_WINDOW_MS / 1000)}s. ` +
      `Killing orchestrator and pausing all restarts for ${pauseMin} minutes.`,
  );

  // Kill the orchestrator child if running
  if (currentChild) {
    try {
      currentChild.kill("SIGTERM");
    } catch {
      /* best effort */
    }
  }

  // Block orchestrator restarts via safe mode
  monitorSafeModeUntil = circuitBreakerResetAt;

  // Send ONE summary Telegram message (if not already notified)
  if (!circuitBreakerNotified && telegramToken && telegramChatId) {
    circuitBreakerNotified = true;
    const msg =
      `🔌 Circuit breaker tripped: ${failureCount} failures in ${Math.round(CIRCUIT_BREAKER_WINDOW_MS / 1000)}s.\n` +
      `Orchestrator killed. All restarts paused for ${pauseMin} minutes.\n` +
      `Will auto-resume at ${new Date(circuitBreakerResetAt).toLocaleTimeString()}.`;
    // Fire-and-forget with skipDedup to ensure it gets through
    sendTelegramMessage(msg, { skipDedup: true }).catch(() => {});
  }
}

let allCompleteNotified = false;
let backlogLowNotified = false;
let idleAgentsNotified = false;
let plannerTriggered = false;
const plannerStatePath = resolve(logDir, "task-planner-state.json");

// ── Telegram history ring buffer ────────────────────────────────────────────
// Stores the last N sent messages for context enrichment (fed to autofix prompts)
const TELEGRAM_HISTORY_MAX = 25;
const telegramHistory = [];
let telegramUpdateOffset = 0;
const telegramCommandQueue = [];
let telegramCommandActive = 0;
let telegramCommandPolling = false;

function pushTelegramHistory(text) {
  const stamp = new Date().toISOString().slice(11, 19);
  telegramHistory.push(`[${stamp}] ${text.slice(0, 300)}`);
  if (telegramHistory.length > TELEGRAM_HISTORY_MAX) {
    telegramHistory.shift();
  }
}

function recordMonitorFailure() {
  const now = Date.now();
  monitorFailureTimestamps.push(now);
  while (
    monitorFailureTimestamps.length &&
    now - monitorFailureTimestamps[0] > monitorFailureWindowMs
  ) {
    monitorFailureTimestamps.shift();
  }
  return monitorFailureTimestamps.length;
}

function shouldRestartMonitor() {
  const now = Date.now();
  if (now - lastMonitorRestartAt < monitorRestartCooldownMs) {
    return false;
  }
  return monitorFailureTimestamps.length >= 3;
}

function schedulePreflightRetry(waitMs) {
  if (preflightRetryTimer) return;
  const delay = Math.max(30000, waitMs || preflightRetryMs);
  preflightRetryTimer = setTimeout(() => {
    preflightRetryTimer = null;
    startProcess();
  }, delay);
}

async function ensurePreflightReady(reason) {
  if (!preflightEnabled) return true;
  if (preflightInProgress) return false;
  const now = Date.now();
  if (preflightLastResult && !preflightLastResult.ok) {
    const elapsed = now - preflightLastRunAt;
    if (elapsed < preflightRetryMs) {
      schedulePreflightRetry(preflightRetryMs - elapsed);
      return false;
    }
  }
  preflightInProgress = true;
  const result = runPreflightChecks({ repoRoot });
  preflightInProgress = false;
  preflightLastResult = result;
  preflightLastRunAt = Date.now();
  const report = formatPreflightReport(result, {
    retryMs: result.ok ? 0 : preflightRetryMs,
  });
  if (!result.ok) {
    console.error(report);
    console.warn(
      `[monitor] preflight failed (${reason || "startup"}); blocking orchestrator start.`,
    );
    schedulePreflightRetry(preflightRetryMs);
    return false;
  }
  console.log(report);
  return true;
}

function restartSelf(reason) {
  if (shuttingDown) return;
  const now = Date.now();
  lastMonitorRestartAt = now;
  console.warn(`[monitor] restarting self (${reason || "unknown"})`);
  try {
    const child = spawn(process.execPath, process.argv.slice(1), {
      cwd: process.cwd(),
      env: { ...process.env },
      detached: true,
      stdio: "ignore",
    });
    child.unref();
  } catch (err) {
    console.warn(
      `[monitor] failed to spawn replacement monitor: ${err.message || err}`,
    );
  }
  process.exit(1);
}

function recordOrchestratorRestart() {
  const now = Date.now();
  orchestratorRestartTimestamps.push(now);
  while (
    orchestratorRestartTimestamps.length &&
    now - orchestratorRestartTimestamps[0] > orchestratorRestartWindowMs
  ) {
    orchestratorRestartTimestamps.shift();
  }
  return orchestratorRestartTimestamps.length;
}

function shouldHaltOrchestrator() {
  const now = Date.now();
  if (now < orchestratorHaltedUntil) {
    return true;
  }
  return orchestratorRestartTimestamps.length >= orchestratorRestartThreshold;
}

function detectChangedFiles(repoRootPath) {
  try {
    const output = execSync("git diff --name-only", {
      cwd: repoRootPath,
      encoding: "utf8",
      timeout: 10_000,
    });
    return output
      .split(/\r?\n/)
      .map((line) => line.trim())
      .filter(Boolean);
  } catch {
    return [];
  }
}

function getChangeSummary(repoRootPath, files) {
  if (!files.length) return "(no file changes detected)";
  try {
    const diff = execSync("git diff --stat", {
      cwd: repoRootPath,
      encoding: "utf8",
      timeout: 10_000,
    });
    return diff.trim() || files.join(", ");
  } catch {
    return files.join(", ");
  }
}

const monitorFixAttempts = new Map();
const monitorFixMaxAttempts = 2;
const monitorFixCooldownMs = 5 * 60 * 1000;

function canAttemptMonitorFix(signature) {
  const record = monitorFixAttempts.get(signature);
  if (!record) return true;
  if (record.count >= monitorFixMaxAttempts) return false;
  if (Date.now() - record.lastAt < monitorFixCooldownMs) return false;
  return true;
}

function recordMonitorFixAttempt(signature) {
  const record = monitorFixAttempts.get(signature) || { count: 0, lastAt: 0 };
  record.count += 1;
  record.lastAt = Date.now();
  monitorFixAttempts.set(signature, record);
  return record.count;
}

async function attemptMonitorFix({ error, logText }) {
  if (!autoFixEnabled) return { fixed: false, outcome: "autofix-disabled" };
  if (!codexEnabled) return { fixed: false, outcome: "codex-disabled" };

  const signature = error?.message || "monitor-crash";
  if (!canAttemptMonitorFix(signature)) {
    return { fixed: false, outcome: "monitor-fix-exhausted" };
  }

  const attemptNum = recordMonitorFixAttempt(signature);
  const prompt = `You are debugging the ${projectName} codex-monitor.

The monitor process hit an unexpected exception and needs a fix.
Please inspect and fix code in the codex-monitor directory:
- monitor.mjs
- autofix.mjs
- maintenance.mjs

Crash info:
${error?.stack || error?.message || String(error)}

Recent log context:
${logText.slice(-4000)}

Instructions:
1) Identify the root cause of the crash in codex-monitor.
2) Apply a minimal fix.
3) Do not refactor unrelated code.
4) Keep behavior stable and production-safe.`;

  const filesBefore = detectChangedFiles(repoRoot);
  const result = await runCodexExec(prompt, repoRoot);
  const filesAfter = detectChangedFiles(repoRoot);
  const newChanges = filesAfter.filter((f) => !filesBefore.includes(f));
  const changeSummary = getChangeSummary(repoRoot, newChanges);

  const stamp = nowStamp();
  const auditPath = resolve(
    logDir,
    `monitor-fix-${stamp}-attempt${attemptNum}.log`,
  );
  await writeFile(
    auditPath,
    [
      `# Monitor fix attempt #${attemptNum}`,
      `# Signature: ${signature}`,
      `# Timestamp: ${new Date().toISOString()}`,
      "",
      "## Prompt sent to Codex:",
      prompt,
      "",
      "## Codex result:",
      result.output || "(no output)",
      result.error ? `## Error: ${result.error}` : "",
      `## Files changed: ${newChanges.join(", ") || "none"}`,
      "",
      "## Diff summary:",
      changeSummary,
    ].join("\n"),
    "utf8",
  );

  if (result.success && newChanges.length > 0) {
    return { fixed: true, outcome: `changes: ${changeSummary}` };
  }

  return {
    fixed: false,
    outcome: result.error || "no changes written",
  };
}

// Hard cap: if we hit this many failures in the window, actually exit.
const MONITOR_FAILURE_HARD_CAP = 30;
// Minimum interval between handleMonitorFailure executions (prevent Telegram spam).
const MONITOR_FAILURE_COOLDOWN_MS = 5000;
let lastMonitorFailureHandledAt = 0;

async function handleMonitorFailure(reason, err) {
  if (monitorFailureHandling) return;
  const now = Date.now();

  // ── Circuit breaker: if tripped, suppress ALL handling silently ──
  if (isCircuitBreakerTripped()) return;

  // Rate-limit: don't re-enter within cooldown
  if (now - lastMonitorFailureHandledAt < MONITOR_FAILURE_COOLDOWN_MS) return;
  monitorFailureHandling = true;
  lastMonitorFailureHandledAt = now;
  const failureCount = recordMonitorFailure();
  const message = err && err.message ? err.message : String(err || reason);

  // ── Circuit breaker: track rapid failure bursts ──
  const burstCount = recordCircuitBreakerEvent();
  if (burstCount >= CIRCUIT_BREAKER_THRESHOLD) {
    tripCircuitBreaker(burstCount);
    monitorFailureHandling = false;
    return; // circuit breaker sends its own summary message
  }

  // Hard cap: exit the process to break the loop for good
  if (failureCount >= MONITOR_FAILURE_HARD_CAP) {
    const msg = `🛑 codex-monitor hit hard failure cap (${failureCount}). Exiting to break crash loop.`;
    console.error(`[monitor] ${msg}`);
    if (telegramToken && telegramChatId) {
      try {
        await sendTelegramMessage(msg);
      } catch {
        /* best effort */
      }
    }
    process.exit(1);
    return;
  }

  try {
    await ensureLogDir();
    const crashPath = resolve(logDir, `monitor-crash-${nowStamp()}.log`);
    const payload = [
      `# Monitor crash: ${reason}`,
      `# Timestamp: ${new Date().toISOString()}`,
      "",
      "## Error:",
      err?.stack || message,
      "",
      "## Recent logs:",
      logRemainder.slice(-8000),
    ].join("\n");
    await writeFile(crashPath, payload, "utf8");

    if (telegramToken && telegramChatId) {
      try {
        await sendTelegramMessage(
          `⚠️ codex-monitor exception (${reason}). Attempting recovery (count=${failureCount}).`,
        );
      } catch {
        /* suppress Telegram errors during failure handling */
      }
    }

    const fixResult = await attemptMonitorFix({
      error: err || new Error(reason),
      logText: logRemainder,
    });

    if (fixResult.fixed) {
      if (telegramToken && telegramChatId) {
        try {
          await sendTelegramMessage(
            `🛠️ codex-monitor auto-fix applied. Restarting monitor.\n${fixResult.outcome}`,
          );
        } catch {
          /* best effort */
        }
      }
      restartSelf("monitor-fix-applied");
      return;
    }

    if (shouldRestartMonitor()) {
      monitorSafeModeUntil = Date.now() + orchestratorPauseMs;
      const pauseMin = Math.round(orchestratorPauseMs / 60000);
      if (telegramToken && telegramChatId) {
        try {
          await sendTelegramMessage(
            `🛑 codex-monitor entering safe mode after repeated failures (${failureCount} in 10m). Pausing restarts for ${pauseMin} minutes.`,
          );
        } catch {
          /* best effort */
        }
      }
      return;
    }
  } catch (fatal) {
    // Use process.stderr to avoid EPIPE on stdout
    try {
      process.stderr.write(
        `[monitor] failure handler crashed: ${fatal.message || fatal}\n`,
      );
    } catch {
      /* completely give up */
    }
  } finally {
    monitorFailureHandling = false;
  }
}

const crashLoopFixAttempts = new Map();
const crashLoopFixMaxAttempts = 2;
const crashLoopFixCooldownMs = 10 * 60 * 1000;

function canAttemptCrashLoopFix(signature) {
  const record = crashLoopFixAttempts.get(signature);
  if (!record) return true;
  if (record.count >= crashLoopFixMaxAttempts) return false;
  if (Date.now() - record.lastAt < crashLoopFixCooldownMs) return false;
  return true;
}

function recordCrashLoopFixAttempt(signature) {
  const record = crashLoopFixAttempts.get(signature) || { count: 0, lastAt: 0 };
  record.count += 1;
  record.lastAt = Date.now();
  crashLoopFixAttempts.set(signature, record);
  return record.count;
}

async function attemptCrashLoopFix({ reason, logText }) {
  if (!autoFixEnabled || !codexEnabled) {
    return { fixed: false, outcome: "codex-disabled" };
  }
  const signature = `crash-loop:${reason}`;
  if (!canAttemptCrashLoopFix(signature)) {
    return { fixed: false, outcome: "crash-loop-fix-exhausted" };
  }

  const attemptNum = recordCrashLoopFixAttempt(signature);
  const prompt = `You are a reliability engineer debugging a crash loop in ${projectName} automation.

The orchestrator is restarting repeatedly within minutes.
Please diagnose the likely root cause and apply a minimal fix.

Targets (edit only if needed):
- ${scriptPath}
- codex-monitor/monitor.mjs
- codex-monitor/autofix.mjs
- codex-monitor/maintenance.mjs

Recent log excerpt:
${logText.slice(-6000)}

Constraints:
1) Prevent rapid restart loops (introduce backoff or safe-mode).
2) Keep behavior stable and production-safe.
3) Do not refactor unrelated code.
4) Prefer small guardrails over big rewrites.`;

  const filesBefore = detectChangedFiles(repoRoot);
  const result = await runCodexExec(prompt, repoRoot, 180_000);
  const filesAfter = detectChangedFiles(repoRoot);
  const newChanges = filesAfter.filter((f) => !filesBefore.includes(f));
  const changeSummary = getChangeSummary(repoRoot, newChanges);

  const stamp = nowStamp();
  const auditPath = resolve(
    logDir,
    `crash-loop-fix-${stamp}-attempt${attemptNum}.log`,
  );
  await writeFile(
    auditPath,
    [
      `# Crash-loop fix attempt #${attemptNum}`,
      `# Signature: ${signature}`,
      `# Timestamp: ${new Date().toISOString()}`,
      "",
      "## Prompt sent to Codex:",
      prompt,
      "",
      "## Codex result:",
      result.output || "(no output)",
      result.error ? `## Error: ${result.error}` : "",
      `## Files changed: ${newChanges.join(", ") || "none"}`,
      "",
      "## Diff summary:",
      changeSummary,
    ].join("\n"),
    "utf8",
  );

  if (result.success && newChanges.length > 0) {
    return { fixed: true, outcome: `changes: ${changeSummary}` };
  }
  return { fixed: false, outcome: result.error || "no changes written" };
}

export function getTelegramHistory() {
  return [...telegramHistory];
}

// ── Repeating error detection (loop detector) ───────────────────────────────────
// Tracks fingerprints of error lines. When the same error appears
// LOOP_THRESHOLD times within LOOP_WINDOW_MS, triggers Codex autofix.
const LOOP_THRESHOLD = 4;
const LOOP_WINDOW_MS = 10 * 60 * 1000; // 10 minutes
const LOOP_COOLDOWN_MS = 15 * 60 * 1000; // 15 min cooldown per fingerprint

/** @type {Map<string, {timestamps: number[], fixTriggeredAt: number}>} */
const errorFrequency = new Map();
let loopFixInProgress = false;

function trackErrorFrequency(line) {
  const fingerprint = getErrorFingerprint(line);
  if (!fingerprint) return;

  const now = Date.now();
  let record = errorFrequency.get(fingerprint);
  if (!record) {
    record = { timestamps: [], fixTriggeredAt: 0 };
    errorFrequency.set(fingerprint, record);
  }

  record.timestamps.push(now);
  // Trim old entries outside window
  record.timestamps = record.timestamps.filter((t) => now - t < LOOP_WINDOW_MS);

  // Check threshold
  if (
    record.timestamps.length >= LOOP_THRESHOLD &&
    now - record.fixTriggeredAt > LOOP_COOLDOWN_MS &&
    !loopFixInProgress
  ) {
    record.fixTriggeredAt = now;
    console.log(
      `[monitor] repeating error detected (${record.timestamps.length}x): ${fingerprint.slice(0, 80)}`,
    );
    triggerLoopFix(line, record.timestamps.length);
  }
}

async function triggerLoopFix(errorLine, repeatCount) {
  if (!autoFixEnabled) return;
  loopFixInProgress = true;

  const telegramFn =
    telegramToken && telegramChatId
      ? (msg) => void sendTelegramMessage(msg)
      : null;

  try {
    const result = await fixLoopingError({
      errorLine,
      repeatCount,
      repoRoot,
      logDir,
      onTelegram: telegramFn,
      recentMessages: getTelegramHistory(),
    });

    if (result.fixed) {
      console.log(
        "[monitor] loop fix applied — file watcher will restart orchestrator",
      );
    }
  } catch (err) {
    console.warn(`[monitor] loop fix error: ${err.message || err}`);
    if (telegramFn) {
      telegramFn(`🔁 Loop fix crashed: ${err.message || err}`);
    }
  } finally {
    loopFixInProgress = false;
  }
}

const contextPatterns = [
  "ContextWindowExceeded",
  "context window",
  "ran out of room",
];

const errorPatterns = [
  /\bERROR\b/i,
  /Exception/i,
  /Traceback/i,
  /SetValueInvocationException/i,
  /Cannot bind argument/i,
  /Unhandled/i,
  /\bFailed\b/i,
  /Copilot assignment failed/i,
];

const errorNoisePatterns = [
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+Status:/i,
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+Initial sync:/i,
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+SyncCopilotState:/i,
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+CI (pending|failing)/i,
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+PR #\d+ .*CI=/i,
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+Merge failed for PR/i,
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+Merge failure reason:/i,
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+Retry merge failed for PR/i,
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+Auto-merge enable failed:/i,
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+Failed to initialize vibe-kanban configuration/i,
  /HTTP GET http:\/\/127\.0\.0\.1:54089\/api\/projects failed/i,
  // Stats summary line (contains "Failed" as a counter, not an error)
  /First-shot:.*Failed:/i,
  // Box-drawing cycle summary lines
  /^\s*[│┃|]\s*(Elapsed|Submitted|Tracked|First-shot):/i,
  /^\s*[─┄╌═]+/,
  /^\s*[└┗╚][─┄╌═]+/,
  /^\s*[╔╗╚╝║═]+/,
  // "No remote branch" is handled by smartPR, not an error
  /No remote branch for .* — agent must push/i,
  // Telegram 409 conflicts (harmless, handled by auto-disable)
  /telegram getUpdates failed: 409/i,
  /getUpdates failed: 409/i,
];

const vkErrorPatterns = [
  /Failed to initialize vibe-kanban configuration/i,
  /HTTP GET http:\/\/127\.0\.0\.1:54089\/api\/projects failed/i,
];

function notifyErrorLine(line) {
  if (!telegramToken || !telegramChatId) {
    return;
  }
  if (vkErrorPatterns.some((pattern) => pattern.test(line))) {
    notifyVkError(line);
    return;
  }

  // Track error frequency for loop detection (always, even if deduped for Telegram)
  trackErrorFrequency(line);

  const key = line.trim();
  if (!key) {
    return;
  }
  const now = Date.now();
  const last = errorNotified.get(key) || 0;
  if (now - last < 5 * 60 * 1000) {
    return;
  }
  errorNotified.set(key, now);
  queueErrorMessage(line.trim());
}

function notifyVkError(line) {
  const key = "vibe-kanban-unavailable";
  const now = Date.now();
  const last = vkErrorNotified.get(key) || 0;
  if (now - last < 10 * 60 * 1000) {
    return;
  }
  vkErrorNotified.set(key, now);
  const vkLink = formatHtmlLink(vkEndpointUrl, "VK_ENDPOINT_URL");
  const publicLink = vkPublicUrl
    ? formatHtmlLink(vkPublicUrl, "Public URL")
    : null;
  const message = [
    `${projectName} Orchestrator Warning`,
    "Vibe-Kanban API unreachable.",
    `Check ${vkLink} and ensure the service is running.`,
    publicLink ? `Open ${publicLink}.` : null,
  ]
    .filter(Boolean)
    .join("\n");
  void sendTelegramMessage(message, { parseMode: "HTML" });
  triggerVibeKanbanRecovery(line);
}

function notifyCodexTrigger(context) {
  if (!telegramToken || !telegramChatId) {
    return;
  }
  void sendTelegramMessage(`Codex triggered: ${context}`);
}

async function runCodexRecovery(reason) {
  if (!codexEnabled) {
    return null;
  }
  try {
    if (!CodexClient) {
      const ready = await ensureCodexSdkReady();
      if (!ready) {
        throw new Error(codexDisabledReason || "Codex SDK not available");
      }
    }
    const codex = new CodexClient();
    const thread = codex.startThread();
    const prompt = `You are monitoring a Node.js orchestrator.
A local service (vibe-kanban) is unreachable.
Provide a short recovery plan and validate environment assumptions.
Reason: ${reason}`;
    const result = await thread.run(prompt);
    const outPath = resolve(logDir, `codex-recovery-${nowStamp()}.txt`);
    await writeFile(outPath, formatCodexResult(result), "utf8");
    return outPath;
  } catch (err) {
    const message = err && err.message ? err.message : String(err);
    const outPath = resolve(logDir, `codex-recovery-${nowStamp()}.txt`);
    await writeFile(outPath, `Codex recovery failed: ${message}\n`, "utf8");
    return null;
  }
}

let vkRestartCount = 0;
const vkMaxRestarts = 20;
const vkRestartDelayMs = 5000;

async function startVibeKanbanProcess() {
  if (!vkSpawnEnabled) {
    return;
  }
  if (vibeKanbanProcess && !vibeKanbanProcess.killed) {
    return;
  }

  // ── Guard: if the API is already reachable (e.g. detached from a previous
  // monitor instance), adopt it instead of spawning a new copy that will
  // crash with EADDRINUSE/exit-code-1.
  if (await isVibeKanbanOnline()) {
    console.log(
      `[monitor] vibe-kanban already online at ${vkEndpointUrl} — skipping spawn`,
    );
    vkRestartCount = 0;
    return;
  }

  // ── Kill any stale process holding the port ───────────────────────
  try {
    const portCheck = execSync(
      `netstat -aon | findstr ":${vkRecoveryPort}.*LISTENING"`,
      { encoding: "utf8", timeout: 5000, stdio: "pipe" },
    ).trim();
    const pidMatch = portCheck.match(/(\d+)\s*$/);
    if (pidMatch) {
      const stalePid = pidMatch[1];
      console.log(
        `[monitor] killing stale process ${stalePid} on port ${vkRecoveryPort}`,
      );
      try {
        execSync(`taskkill /F /PID ${stalePid}`, {
          timeout: 5000,
          stdio: "pipe",
        });
      } catch {
        /* best effort */
      }
      // Brief delay so the OS releases the port
      await new Promise((r) => setTimeout(r, 1500));
    }
  } catch {
    /* no process on port — fine */
  }

  const env = {
    ...process.env,
    PORT: vkRecoveryPort,
    HOST: vkRecoveryHost,
  };

  // Prefer locally-installed vibe-kanban binary (from npm dependency),
  // fall back to npx for global/remote installs.
  const vkBin = resolve(__dirname, "node_modules", ".bin", "vibe-kanban");
  const useLocal = existsSync(vkBin) || existsSync(vkBin + ".cmd");
  const spawnCmd = useLocal
    ? process.platform === "win32"
      ? vkBin + ".cmd"
      : vkBin
    : "npx";
  const spawnArgs = useLocal ? [] : ["--yes", "vibe-kanban"];

  console.log(
    `[monitor] starting vibe-kanban via ${useLocal ? "local bin" : "npx"} (HOST=${vkRecoveryHost} PORT=${vkRecoveryPort}, endpoint=${vkEndpointUrl})`,
  );

  vibeKanbanProcess = spawn(spawnCmd, spawnArgs, {
    env,
    cwd: repoRoot,
    stdio: "ignore",
    shell: true,
    detached: true,
  });
  vibeKanbanProcess.unref();
  vibeKanbanStartedAt = Date.now();

  vibeKanbanProcess.on("error", (err) => {
    vibeKanbanProcess = null;
    vibeKanbanStartedAt = 0;
    const message = err && err.message ? err.message : String(err);
    console.warn(`[monitor] vibe-kanban spawn error: ${message}`);
    scheduleVibeKanbanRestart();
  });

  vibeKanbanProcess.on("exit", (code, signal) => {
    vibeKanbanProcess = null;
    vibeKanbanStartedAt = 0;
    const reason = signal ? `signal ${signal}` : `exit code ${code}`;
    console.warn(`[monitor] vibe-kanban exited (${reason})`);
    if (!shuttingDown) {
      scheduleVibeKanbanRestart();
    }
  });
}

function scheduleVibeKanbanRestart() {
  if (shuttingDown) return;
  if (!vkSpawnEnabled) return;
  vkRestartCount++;
  if (vkRestartCount > vkMaxRestarts) {
    console.error(
      `[monitor] vibe-kanban exceeded ${vkMaxRestarts} restarts, giving up`,
    );
    if (telegramToken && telegramChatId) {
      void sendTelegramMessage(
        `Vibe-kanban exceeded ${vkMaxRestarts} restart attempts. Manual intervention required.`,
      );
    }
    return;
  }
  const delay = Math.min(vkRestartDelayMs * vkRestartCount, 60000);
  console.log(
    `[monitor] restarting vibe-kanban in ${delay}ms (attempt ${vkRestartCount}/${vkMaxRestarts})`,
  );
  setTimeout(() => void startVibeKanbanProcess(), delay);
}

async function canConnectTcp(host, port, timeoutMs = 1200) {
  return new Promise((resolve) => {
    const socket = net.connect({ host, port: Number(port) });
    const done = (ok) => {
      try {
        socket.destroy();
      } catch {
        /* best effort */
      }
      resolve(ok);
    };
    socket.setTimeout(timeoutMs);
    socket.once("connect", () => done(true));
    socket.once("timeout", () => done(false));
    socket.once("error", () => done(false));
  });
}

async function isVibeKanbanOnline() {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 2000);
  try {
    const res = await fetch(`${vkEndpointUrl}/api/projects`, {
      signal: controller.signal,
    });
    // Any HTTP response means the service is up, even if auth/route fails.
    return true;
  } catch {
    return await canConnectTcp(vkRecoveryHost, vkRecoveryPort);
  } finally {
    clearTimeout(timeout);
  }
}

async function ensureVibeKanbanRunning() {
  if (!vkSpawnEnabled) {
    return;
  }
  if (await isVibeKanbanOnline()) {
    // Reset restart counter on successful health check
    vkRestartCount = 0;
    return;
  }
  // If process is alive, give it 15s grace to start up
  if (vibeKanbanProcess && !vibeKanbanProcess.killed) {
    const graceMs = 15000;
    if (vibeKanbanStartedAt && Date.now() - vibeKanbanStartedAt < graceMs) {
      return;
    }
    // Process alive but API not responding — kill and let auto-restart handle it
    console.warn(
      "[monitor] vibe-kanban process alive but API unresponsive, killing",
    );
    try {
      vibeKanbanProcess.kill();
    } catch {
      /* best effort */
    }
    return;
  }
  // No process running — start fresh
  await startVibeKanbanProcess();
}

function restartVibeKanbanProcess() {
  if (!vkSpawnEnabled) {
    return;
  }
  // Just kill the process — the exit handler will auto-restart it
  if (vibeKanbanProcess && !vibeKanbanProcess.killed) {
    try {
      vibeKanbanProcess.kill();
    } catch {
      /* best effort */
    }
  } else {
    void startVibeKanbanProcess();
  }
}

async function triggerVibeKanbanRecovery(reason) {
  if (!vkSpawnEnabled) {
    return;
  }
  const now = Date.now();
  const cooldownMs = vkRecoveryCooldownMin * 60 * 1000;
  if (now - vkRecoveryLastAt < cooldownMs) {
    return;
  }
  vkRecoveryLastAt = now;

  if (telegramToken && telegramChatId) {
    const link = formatHtmlLink(vkEndpointUrl, "VK_ENDPOINT_URL");
    const notice = codexEnabled
      ? `Codex recovery triggered: vibe-kanban unreachable. Attempting restart. (${link})`
      : `Vibe-kanban recovery triggered (Codex disabled). Attempting restart. (${link})`;
    void sendTelegramMessage(notice, { parseMode: "HTML" });
  }
  await runCodexRecovery(reason || "vibe-kanban unreachable");
  restartVibeKanbanProcess();
}

// ── VK API client ───────────────────────────────────────────────────────────

/**
 * Generic HTTP client for the Vibe-Kanban REST API.
 * @param {string} path  - API path (e.g. "/api/projects")
 * @param {object} [opts] - { method, body, timeoutMs }
 * @returns {Promise<object|null>} Parsed JSON body, or null on failure.
 */
async function fetchVk(path, opts = {}) {
  const url = `${vkEndpointUrl}${path.startsWith("/") ? path : "/" + path}`;
  const method = (opts.method || "GET").toUpperCase();
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), opts.timeoutMs || 15000);
  try {
    const fetchOpts = {
      method,
      signal: controller.signal,
      headers: { "Content-Type": "application/json" },
    };
    if (opts.body && method !== "GET") {
      fetchOpts.body = JSON.stringify(opts.body);
    }
    const res = await fetch(url, fetchOpts);
    if (!res.ok) {
      const text = await res.text().catch(() => "");
      console.warn(
        `[monitor] fetchVk ${method} ${path} failed: ${res.status} ${text.slice(0, 200)}`,
      );
      return null;
    }
    const contentType = res.headers.get("content-type") || "";
    if (!contentType.includes("application/json")) {
      const text = await res.text().catch(() => "");
      console.warn(
        `[monitor] fetchVk ${method} ${path} error: non-JSON response (${contentType || "unknown"})`,
      );
      if (text) {
        console.warn(
          `[monitor] fetchVk ${method} ${path} body: ${text.slice(0, 200)}`,
        );
      }
      const now = Date.now();
      if (now - vkNonJsonNotifiedAt > 10 * 60 * 1000) {
        vkNonJsonNotifiedAt = now;
        notifyVkError(
          "Vibe-Kanban API returned HTML/non-JSON. Check VK_BASE_URL/VK_ENDPOINT_URL.",
        );
      }
      return null;
    }
    return await res.json();
  } catch (err) {
    const msg = err?.message || String(err);
    if (!msg.includes("abort")) {
      console.warn(`[monitor] fetchVk ${method} ${path} error: ${msg}`);
    }
    return null;
  } finally {
    clearTimeout(timeout);
  }
}

/**
 * GET /api/task-attempts/:id/branch-status
 * Returns branch status data for an attempt (commits ahead/behind, conflicts, etc.)
 */
async function fetchBranchStatus(attemptId) {
  const res = await fetchVk(`/api/task-attempts/${attemptId}/branch-status`);
  if (!res?.success || !Array.isArray(res.data)) return null;
  return res.data[0] || null;
}

async function getAttemptInfo(attemptId) {
  try {
    const statusData = await readStatusData();
    const attempts = statusData?.active_attempts || [];
    const match = attempts.find((a) => a.id === attemptId);
    if (match) return match;
  } catch {
    /* best effort */
  }
  const res = await fetchVk(`/api/task-attempts/${attemptId}`);
  if (res?.success && res.data) {
    return res.data;
  }
  return null;
}

function ghAvailable() {
  const res = spawnSync("gh", ["--version"], { stdio: "ignore" });
  return res.status === 0;
}

async function findExistingPrForBranch(branch) {
  if (!branch || !ghAvailable()) return null;
  const res = spawnSync(
    "gh",
    [
      "pr",
      "list",
      "--head",
      branch,
      "--state",
      "all",
      "--limit",
      "5",
      "--json",
      "number,state,title,url,mergedAt,closedAt",
    ],
    { encoding: "utf8" },
  );
  if (res.status !== 0) {
    return null;
  }
  try {
    const items = JSON.parse(res.stdout || "[]");
    return Array.isArray(items) && items.length > 0 ? items[0] : null;
  } catch {
    return null;
  }
}

async function findExistingPrForBranchApi(branch) {
  if (!branch || !githubToken || !repoSlug) return null;
  const [owner, repo] = repoSlug.split("/");
  if (!owner || !repo) return null;
  const head = `${owner}:${branch}`;
  const url = `https://api.github.com/repos/${owner}/${repo}/pulls?state=all&head=${encodeURIComponent(
    head,
  )}`;
  try {
    const res = await fetch(url, {
      headers: {
        Authorization: `Bearer ${githubToken}`,
        Accept: "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28",
        "User-Agent": "codex-monitor",
      },
    });
    if (!res.ok) {
      const text = await res.text().catch(() => "");
      console.warn(
        `[monitor] GitHub API PR lookup failed (${res.status}): ${text.slice(0, 120)}`,
      );
      return null;
    }
    const items = await res.json();
    return Array.isArray(items) && items.length > 0 ? items[0] : null;
  } catch (err) {
    console.warn(
      `[monitor] GitHub API PR lookup error: ${err?.message || err}`,
    );
    return null;
  }
}

async function getPullRequestByNumber(prNumber) {
  if (!Number.isFinite(prNumber) || prNumber <= 0) return null;
  if (ghAvailable()) {
    const res = spawnSync(
      "gh",
      [
        "pr",
        "view",
        String(prNumber),
        "--json",
        "number,state,title,url,mergedAt,closedAt,mergeable,mergeStateStatus",
      ],
      { encoding: "utf8" },
    );
    if (res.status === 0) {
      try {
        return JSON.parse(res.stdout || "{}");
      } catch {
        /* fall through */
      }
    }
  }
  if (!githubToken || !repoSlug) return null;
  const [owner, repo] = repoSlug.split("/");
  if (!owner || !repo) return null;
  const url = `https://api.github.com/repos/${owner}/${repo}/pulls/${prNumber}`;
  try {
    const res = await fetch(url, {
      headers: {
        Authorization: `Bearer ${githubToken}`,
        Accept: "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28",
        "User-Agent": "codex-monitor",
      },
    });
    if (!res.ok) {
      const text = await res.text().catch(() => "");
      console.warn(
        `[monitor] GitHub API PR ${prNumber} lookup failed (${res.status}): ${text.slice(0, 120)}`,
      );
      return null;
    }
    return await res.json();
  } catch (err) {
    console.warn(
      `[monitor] GitHub API PR ${prNumber} lookup error: ${err?.message || err}`,
    );
    return null;
  }
}

/**
 * Find the matching VK project by projectName, with caching.
 * Falls back to the first project if no name match.
 */
async function findVkProjectId() {
  if (cachedProjectId) return cachedProjectId;

  const projectsRes = await fetchVk("/api/projects");
  if (
    !projectsRes?.success ||
    !Array.isArray(projectsRes.data) ||
    projectsRes.data.length === 0
  ) {
    console.warn("[monitor] Failed to fetch projects from VK API");
    return null;
  }

  // Match by projectName (case-insensitive)
  const match = projectsRes.data.find(
    (p) => p.name?.toLowerCase() === projectName?.toLowerCase(),
  );
  const project = match || projectsRes.data[0];
  if (!project?.id) {
    console.warn("[monitor] No projects found in VK API");
    return null;
  }
  if (!match) {
    console.warn(
      `[monitor] No VK project matching "${projectName}" — using "${project.name}" as fallback`,
    );
  }
  cachedProjectId = project.id;
  console.log(
    `[monitor] Cached project_id: ${cachedProjectId.substring(0, 8)}... (${project.name})`,
  );
  return cachedProjectId;
}

/**
 * Fetches and caches the repo_id from VK API.
 * Uses the flat /api/repos endpoint and matches by repoRoot path or projectName.
 */
async function getRepoId() {
  if (cachedRepoId) return cachedRepoId;
  if (process.env.VK_REPO_ID) {
    cachedRepoId = process.env.VK_REPO_ID;
    return cachedRepoId;
  }

  try {
    // Use the flat /api/repos endpoint (not nested under projects)
    const reposRes = await fetchVk("/api/repos");
    if (
      !reposRes?.success ||
      !Array.isArray(reposRes.data) ||
      reposRes.data.length === 0
    ) {
      console.warn("[monitor] Failed to fetch repos from VK API");
      return null;
    }

    // Match by repo path (normalized for comparison)
    const normalPath = (p) =>
      (p || "").replace(/\\/g, "/").replace(/\/+$/, "").toLowerCase();
    const targetPath = normalPath(repoRoot);

    let repo = reposRes.data.find((r) => normalPath(r.path) === targetPath);

    // Fallback: match by name / display_name
    if (!repo) {
      repo = reposRes.data.find(
        (r) =>
          (r.name || r.display_name || "").toLowerCase() ===
          projectName?.toLowerCase(),
      );
    }

    if (!repo) {
      console.warn(
        `[monitor] No VK repo matching path "${repoRoot}" or name "${projectName}" — ` +
          `available: ${reposRes.data.map((r) => r.name).join(", ")}`,
      );
      return null;
    }

    cachedRepoId = repo.id;
    console.log(
      `[monitor] Cached repo_id: ${cachedRepoId.substring(0, 8)}... (${repo.name})`,
    );
    return cachedRepoId;
  } catch (err) {
    console.warn(`[monitor] Error fetching repo_id: ${err.message}`);
    return null;
  }
}

/**
 * POST /api/task-attempts/:id/rebase
 * Rebases the attempt's worktree onto target branch.
 */
async function rebaseAttempt(attemptId, baseBranch) {
  const repoId = await getRepoId();
  if (!repoId) {
    console.warn("[monitor] Cannot rebase: repo_id not available");
    return {
      success: false,
      error: "repo_id_missing",
      message: "repo_id not available",
    };
  }
  const body = { repo_id: repoId };
  if (baseBranch) {
    body.old_base_branch = baseBranch;
    body.new_base_branch = baseBranch;
  }
  const res = await fetchVk(`/api/task-attempts/${attemptId}/rebase`, {
    method: "POST",
    body,
    timeoutMs: 60000,
  });
  return res;
}

/**
 * POST /api/task-attempts/:id/pr
 * Creates a PR via the VK API (triggers prepush hooks in the worktree).
 * Can take up to 15 minutes if prepush hooks run lint/test/build.
 * @param {string} attemptId
 * @param {object} prOpts - { title, description, draft }
 */
async function createPRViaVK(attemptId, prOpts = {}) {
  // Fetch repo_id if not cached
  const repoId = await getRepoId();
  if (!repoId) {
    console.error("[monitor] Cannot create PR: repo_id not available");
    return { success: false, error: "repo_id_missing", _elapsedMs: 0 };
  }

  const body = {
    repo_id: repoId,
    title: prOpts.title || "",
    description: prOpts.description || "",
    draft: prOpts.draft ?? true,
    base: prOpts.base || process.env.VK_TARGET_BRANCH || "origin/main",
  };
  const startMs = Date.now();
  const res = await fetchVk(`/api/task-attempts/${attemptId}/pr`, {
    method: "POST",
    body,
    timeoutMs: 15 * 60 * 1000, // prepush hooks can take up to 15 min
  });
  const elapsed = Date.now() - startMs;
  // Attach timing so callers can distinguish instant vs slow failures
  if (res) res._elapsedMs = elapsed;
  return { ...(res || { success: false }), _elapsedMs: elapsed };
}

/**
 * POST /api/task-attempts/:id/resolve-conflicts
 * Auto-resolves merge conflicts after a failed rebase by accepting "ours" changes.
 */
async function resolveConflicts(attemptId) {
  const res = await fetchVk(
    `/api/task-attempts/${attemptId}/resolve-conflicts`,
    { method: "POST", body: {}, timeoutMs: 60000 },
  );
  return res;
}

/**
 * POST /api/task-attempts/:id/archive
 * Archives a stale attempt (0 commits, many behind).
 */
async function archiveAttempt(attemptId) {
  const res = await fetchVk(`/api/task-attempts/${attemptId}/archive`, {
    method: "POST",
    body: {},
    timeoutMs: 30000,
  });
  return res;
}

// ── Fresh session retry system ──────────────────────────────────────────────
// When an agent gets stuck (context window exhausted, crash loop, repeated
// failures), starting a fresh session in the SAME workspace is often the
// most effective recovery — the new agent gets clean context but inherits
// the existing worktree and file changes.

/**
 * Build a retry prompt that gives the fresh agent full task context.
 * Mirrors the format the user showed: failure notice + task context block.
 *
 * @param {object} attemptInfo - { task_id, task_title, task_description, branch, id }
 * @param {string} reason      - Why we're retrying (e.g., "context_window_exhausted")
 * @param {string} [logTail]   - Last N chars of log for diagnosis
 * @returns {string} The follow-up prompt
 */
function buildRetryPrompt(attemptInfo, reason, logTail) {
  const parts = [
    `Detected a failure (${reason}). Please retry your task. If it fails again, I will start a fresh session.`,
    "",
    "Task context (vibe-kanban):",
    `Branch: ${attemptInfo.branch || "unknown"}`,
    `Title: ${attemptInfo.task_title || attemptInfo.name || "unknown"}`,
  ];

  if (attemptInfo.task_description) {
    parts.push(`Description:\n${attemptInfo.task_description}`);
  }

  parts.push(
    "",
    "If VE_TASK_TITLE/VE_TASK_DESCRIPTION are missing, treat this as a VK task:",
    "Worktree paths often include .git/worktrees/ or vibe-kanban.",
    "VK tasks always map to a ve/<id>-<slug> branch.",
    "Resume with the context above, then commit/push/PR as usual.",
  );

  if (logTail) {
    const trimmed = logTail.slice(-2000).trim();
    if (trimmed) {
      parts.push("", "Recent log output:", "```", trimmed, "```");
    }
  }

  return parts.join("\n");
}

/**
 * Get the currently active attempt info from VK status data.
 * @returns {Promise<object|null>} Attempt info with task context, or null
 */
async function getActiveAttemptInfo() {
  try {
    const statusData = await readStatusData();
    const attempts = statusData?.active_attempts || [];
    // Find the running/most recent attempt
    const running =
      attempts.find((a) => a.status === "running") ||
      attempts.find((a) => a.status === "error") ||
      attempts[0];

    if (!running) return null;

    // Enrich with task description if available
    if (running.task_id && !running.task_description) {
      try {
        const taskRes = await fetchVk(`/api/tasks/${running.task_id}`);
        if (taskRes?.success && taskRes.data) {
          running.task_title = running.task_title || taskRes.data.title;
          running.task_description =
            taskRes.data.description || taskRes.data.body || "";
        }
      } catch {
        /* best effort */
      }
    }

    return running;
  } catch {
    return null;
  }
}

// Rate-limit fresh session creation to avoid spam
const FRESH_SESSION_COOLDOWN_MS = 5 * 60 * 1000; // 5 minutes
let lastFreshSessionAt = 0;
let freshSessionCount = 0;
const FRESH_SESSION_MAX_PER_TASK = 3; // max retries per task before giving up
const freshSessionTaskRetries = new Map();

/**
 * Start a fresh VK session in the same workspace and send a retry prompt.
 * This is the nuclear option when an agent is irrecoverably stuck.
 *
 * @param {string} workspaceId - The workspace/attempt UUID
 * @param {string} prompt      - The follow-up prompt with task context
 * @param {string} taskId      - Task ID for retry tracking
 * @returns {Promise<{success: boolean, sessionId?: string, reason?: string}>}
 */
async function startFreshSession(workspaceId, prompt, taskId) {
  const now = Date.now();

  // Rate limit
  if (now - lastFreshSessionAt < FRESH_SESSION_COOLDOWN_MS) {
    const waitSec = Math.round(
      (FRESH_SESSION_COOLDOWN_MS - (now - lastFreshSessionAt)) / 1000,
    );
    console.warn(`[monitor] fresh session rate-limited, ${waitSec}s remaining`);
    return { success: false, reason: `rate-limited (${waitSec}s)` };
  }

  // Per-task retry limit
  if (taskId) {
    const retries = freshSessionTaskRetries.get(taskId) || 0;
    if (retries >= FRESH_SESSION_MAX_PER_TASK) {
      console.warn(
        `[monitor] fresh session limit reached for task ${taskId.slice(0, 8)} (${retries}/${FRESH_SESSION_MAX_PER_TASK})`,
      );
      return {
        success: false,
        reason: `max retries (${FRESH_SESSION_MAX_PER_TASK}) reached for task`,
      };
    }
    freshSessionTaskRetries.set(taskId, retries + 1);
  }

  lastFreshSessionAt = now;
  freshSessionCount += 1;

  try {
    // Step 1: Create a new session for the workspace
    const session = await fetchVk("/api/sessions", {
      method: "POST",
      body: { workspace_id: workspaceId },
      timeoutMs: 15000,
    });

    if (!session?.id) {
      console.warn("[monitor] failed to create fresh VK session");
      return { success: false, reason: "session creation failed" };
    }

    // Step 2: Send the retry prompt as a follow-up
    const followUp = await fetchVk(`/api/sessions/${session.id}/follow-up`, {
      method: "POST",
      body: { prompt },
      timeoutMs: 15000,
    });

    if (!followUp) {
      console.warn("[monitor] failed to send follow-up to fresh session");
      return { success: false, reason: "follow-up send failed" };
    }

    console.log(
      `[monitor] ✅ Fresh session started: ${session.id} (retry #${freshSessionCount})`,
    );

    return { success: true, sessionId: session.id };
  } catch (err) {
    console.warn(`[monitor] fresh session error: ${err.message || err}`);
    return { success: false, reason: err.message || String(err) };
  }
}

/**
 * High-level: detect a stuck agent, build retry prompt, start fresh session.
 * Call this from handleExit, crash loop detection, or smartPRFlow stale detection.
 *
 * @param {string} reason  - Why we're retrying
 * @param {string} [logTail] - Recent log output for context
 * @returns {Promise<boolean>} true if fresh session started
 */
async function attemptFreshSessionRetry(reason, logTail) {
  if (!vkEndpointUrl) {
    console.log("[monitor] fresh session retry skipped — no VK endpoint");
    return false;
  }

  const attemptInfo = await getActiveAttemptInfo();
  if (!attemptInfo?.id) {
    console.log("[monitor] fresh session retry skipped — no active attempt");
    return false;
  }

  const prompt = buildRetryPrompt(attemptInfo, reason, logTail);
  const result = await startFreshSession(
    attemptInfo.id,
    prompt,
    attemptInfo.task_id,
  );

  if (result.success) {
    if (telegramToken && telegramChatId) {
      const taskLabel =
        attemptInfo.task_title || attemptInfo.branch || "unknown";
      void sendTelegramMessage(
        `🔄 Fresh session started for "${taskLabel}" (${reason}).\nNew session: ${result.sessionId}`,
      );
    }
    return true;
  }

  console.warn(`[monitor] fresh session retry failed: ${result.reason}`);
  if (telegramToken && telegramChatId) {
    void sendTelegramMessage(
      `⚠️ Fresh session retry failed (${reason}): ${result.reason}`,
    );
  }
  return false;
}

/**
 * Calculate how long a task has been in its current state (ms).
 * Uses `updated_at` if available, otherwise `created_at`.
 * @param {object} task - VK task object with `updated_at` / `created_at`
 * @returns {number} Age in milliseconds, or 0 if no timestamp available
 */
function getTaskAgeMs(task) {
  const ts = task?.updated_at || task?.created_at;
  if (!ts) return 0;
  const parsed = new Date(ts).getTime();
  if (Number.isNaN(parsed)) return 0;
  return Math.max(0, Date.now() - parsed);
}

/**
 * GET /api/projects/:project_id/tasks?status=<status>
 * Fetches tasks by status from VK API.
 * @param {string} status - Task status (e.g., "inreview", "todo", "done")
 * @returns {Promise<Array>} Array of task objects, or empty array on failure
 */
async function fetchTasksByStatus(status) {
  try {
    // Find matching VK project
    const projectId = await findVkProjectId();
    if (!projectId) {
      console.warn("[monitor] No VK project found for task query");
      return [];
    }

    // Use flat /api/tasks endpoint with query params
    const tasksRes = await fetchVk(
      `/api/tasks?project_id=${projectId}&status=${status}`,
    );
    if (!tasksRes?.success || !Array.isArray(tasksRes.data)) {
      console.warn(`[monitor] Failed to fetch tasks with status=${status}`);
      return [];
    }

    return tasksRes.data;
  } catch (err) {
    console.warn(
      `[monitor] Error fetching tasks by status: ${err.message || err}`,
    );
    return [];
  }
}

/**
 * PUT /api/tasks/:task_id
 * Updates task status via VK API.
 * @param {string} taskId - Task UUID
 * @param {string} newStatus - New status ("todo", "inprogress", "inreview", "done", "cancelled")
 * @returns {Promise<boolean>} true if successful, false otherwise
 */
async function updateTaskStatus(taskId, newStatus) {
  const res = await fetchVk(`/api/tasks/${taskId}`, {
    method: "PUT",
    body: { status: newStatus },
    timeoutMs: 10000,
  });
  return res?.success === true;
}

/**
 * Checks if a git branch has been merged into the main branch.
 * Uses GitHub CLI + git commands to determine merge status.
 *
 * IMPORTANT: "branch not on remote" does NOT mean merged. The agent may
 * never have pushed, the PR may have been closed without merging, or the
 * branch was manually deleted. We must verify via GitHub PR state.
 *
 * @param {string} branch - Branch name (e.g., "ve/1234-feat-auth")
 * @returns {Promise<boolean>} true if definitively merged, false otherwise
 */
async function isBranchMerged(branch) {
  if (!branch) return false;

  try {
    // ── Strategy 1: Check GitHub for a merged PR with this head branch ──
    // This is the most reliable signal — if GitHub says merged, it's merged.
    if (ghAvailable()) {
      try {
        const ghResult = execSync(
          `gh pr list --head "${branch}" --state merged --json number,mergedAt --limit 1`,
          {
            cwd: repoRoot,
            encoding: "utf8",
            stdio: ["pipe", "pipe", "ignore"],
            timeout: 15000,
          },
        ).trim();
        const mergedPRs = JSON.parse(ghResult || "[]");
        if (mergedPRs.length > 0) {
          console.log(
            `[monitor] Branch ${branch} has merged PR #${mergedPRs[0].number}`,
          );
          return true;
        }
      } catch {
        // gh failed — fall through to git-based checks
      }
    }

    // ── Strategy 2: Check if branch exists on remote ────────────────────
    const branchExistsCmd = `git ls-remote --heads origin ${branch}`;
    const branchExists = execSync(branchExistsCmd, {
      cwd: repoRoot,
      encoding: "utf8",
      stdio: ["pipe", "pipe", "ignore"],
    }).trim();

    // Branch NOT on remote — this does NOT prove it was merged.
    // Without a confirmed merged PR (strategy 1), we must assume NOT merged.
    if (!branchExists) {
      console.log(
        `[monitor] Branch ${branch} not found on remote — no merged PR found, treating as NOT merged`,
      );
      return false;
    }

    // ── Strategy 3: Branch exists on remote — check if ancestor of main ─
    execSync("git fetch origin main --quiet", {
      cwd: repoRoot,
      stdio: "ignore",
      timeout: 15000,
    });

    // Check if the branch is fully merged into origin/main
    // Returns non-zero exit code if not merged
    const mergeCheckCmd = `git merge-base --is-ancestor origin/${branch} origin/main`;
    execSync(mergeCheckCmd, {
      cwd: repoRoot,
      stdio: "ignore",
      timeout: 10000,
    });

    // If we get here, the branch is merged
    console.log(`[monitor] Branch ${branch} is ancestor of main (merged)`);
    return true;
  } catch (err) {
    // Non-zero exit code means not merged, or other error
    return false;
  }
}

/**
 * Persistent cache of task IDs already confirmed as done.
 * Survives monitor restarts by writing to disk.
 * @type {Set<string>}
 */
const mergedTaskCache = new Set();

/**
 * Branch-level dedup cache — VK can have duplicate tasks (different IDs)
 * pointing at the same branch. Once a branch is confirmed merged we skip
 * ALL tasks that reference it, regardless of task ID.
 * @type {Set<string>}
 */
const mergedBranchCache = new Set();

/** Path to the persistent merged-task cache file */
const mergedTaskCachePath = resolve(
  config.cacheDir || resolve(config.repoRoot, ".cache"),
  "ve-merged-tasks.json",
);

/** Load persisted merged-task cache from disk (best-effort) */
function loadMergedTaskCache() {
  try {
    if (existsSync(mergedTaskCachePath)) {
      const raw = readFileSync(mergedTaskCachePath, "utf8");
      const data = JSON.parse(raw);
      // No expiry — merged PRs don't un-merge. Cache is permanent.
      const ids = data.taskIds ?? data; // back-compat: old format was flat {id:ts}
      for (const id of Object.keys(ids)) {
        mergedTaskCache.add(id);
      }
      if (Array.isArray(data.branches)) {
        for (const b of data.branches) {
          mergedBranchCache.add(b);
        }
      }
      const total = mergedTaskCache.size + mergedBranchCache.size;
      if (total > 0) {
        console.log(
          `[monitor] Restored ${mergedTaskCache.size} task IDs + ${mergedBranchCache.size} branches from merged-task cache`,
        );
      }
    }
  } catch {
    /* best-effort — start fresh */
  }
}

/** Persist merged-task cache to disk (best-effort) */
function saveMergedTaskCache() {
  try {
    const taskIds = {};
    const now = Date.now();
    for (const id of mergedTaskCache) {
      taskIds[id] = now;
    }
    const payload = {
      taskIds,
      branches: [...mergedBranchCache],
    };
    writeFileSync(
      mergedTaskCachePath,
      JSON.stringify(payload, null, 2),
      "utf8",
    );
  } catch {
    /* best-effort */
  }
}

// Load cache on startup
loadMergedTaskCache();

/** Maximum number of tasks to process per sweep (0 = unlimited) */
const MERGE_CHECK_BATCH_SIZE = 0;

/** Small delay between GitHub API calls to avoid rate-limiting (ms) */
const MERGE_CHECK_THROTTLE_MS = 1500;

/**
 * Cooldown cache for tasks whose branches are all unresolvable (deleted,
 * no PR, abandoned).  We re-check them every 30 min instead of every cycle.
 * After STALE_MAX_STRIKES consecutive stale checks the task is moved back
 * to "todo" so another agent can pick it up.
 * Key = task ID, Value = { lastCheck: timestamp, strikes: number }.
 * @type {Map<string, {lastCheck: number, strikes: number}>}
 */
const staleBranchCooldown = new Map();
const STALE_COOLDOWN_MS = 30 * 60 * 1000; // 30 minutes
const STALE_MAX_STRIKES = 2; // move to todo after this many stale checks

/**
 * Age-based stale detection: if a task has been in inprogress/inreview for
 * longer than this threshold with no active branch or PR, it is immediately
 * moved back to "todo" on the first check — no strikes needed.
 * Configurable via STALE_TASK_AGE_HOURS env var (default: 3).
 */
const STALE_TASK_AGE_HOURS = Number(process.env.STALE_TASK_AGE_HOURS || "3");
const STALE_TASK_AGE_MS = STALE_TASK_AGE_HOURS * 60 * 60 * 1000;

/**
 * Cooldown cache for tasks whose PRs have merge conflicts.
 * We re-trigger conflict resolution at most every 30 minutes per task.
 * Key = task ID, Value = timestamp of last resolution attempt.
 * @type {Map<string, number>}
 */
const conflictResolutionCooldown = new Map();
const CONFLICT_COOLDOWN_MS = 30 * 60 * 1000; // 30 minutes

/**
 * Periodic check: find tasks in "inreview" status, check if their PRs
 * have been merged, and automatically move them to "done" status.
 * Also detects open PRs with merge conflicts and triggers resolution.
 */
async function checkMergedPRsAndUpdateTasks() {
  try {
    console.log("[monitor] Checking for merged PRs to update task status...");

    const statuses = ["inreview", "inprogress"];
    const tasksByStatus = await Promise.all(
      statuses.map((status) => fetchTasksByStatus(status)),
    );
    const taskMap = new Map();
    statuses.forEach((status, index) => {
      for (const task of tasksByStatus[index]) {
        if (task?.id) {
          taskMap.set(task.id, { task, status });
        }
      }
    });
    const reviewTasks = Array.from(taskMap.values()).filter(
      (entry) => !mergedTaskCache.has(entry.task.id),
    );
    if (reviewTasks.length === 0) {
      console.log(
        "[monitor] No tasks in review/inprogress status (after dedup)",
      );
      return { checked: 0, movedDone: 0, movedReview: 0 };
    }

    const totalCandidates = reviewTasks.length;
    const batch =
      MERGE_CHECK_BATCH_SIZE > 0
        ? reviewTasks.slice(0, MERGE_CHECK_BATCH_SIZE)
        : reviewTasks;
    console.log(
      `[monitor] Found ${totalCandidates} tasks in review/inprogress` +
        (MERGE_CHECK_BATCH_SIZE > 0 && totalCandidates > MERGE_CHECK_BATCH_SIZE
          ? ` (processing first ${MERGE_CHECK_BATCH_SIZE})`
          : ""),
    );

    // For each task, get its workspace/branch and check if merged
    const statusData = await readStatusData();
    const attempts = Array.isArray(statusData?.active_attempts)
      ? statusData.active_attempts
      : Object.values(statusData?.attempts || {});

    // Also fetch VK task-attempts as fallback (covers archived attempts
    // that are no longer in the orchestrator's status file)
    let vkAttempts = [];
    try {
      const vkRes = await fetchVk("/api/task-attempts");
      const vkData = vkRes?.data ?? vkRes;
      if (Array.isArray(vkData)) {
        vkAttempts = vkData;
      }
    } catch {
      /* best-effort fallback */
    }

    let movedCount = 0;
    let movedReviewCount = 0;
    let movedTodoCount = 0;
    let conflictsTriggered = 0;
    /** @type {string[]} */
    const completedTaskNames = [];
    /** @type {string[]} */
    const recoveredTaskNames = [];

    for (const entry of batch) {
      const task = entry.task;
      const taskStatus = entry.status;

      // ── Skip cancelled/done tasks — they should never be recovered ──
      if (taskStatus === "cancelled" || taskStatus === "done") {
        continue;
      }

      // ── Stale cooldown: skip tasks we already checked recently ──
      const staleEntry = staleBranchCooldown.get(task.id);
      if (
        staleEntry &&
        Date.now() - staleEntry.lastCheck < STALE_COOLDOWN_MS
      ) {
        continue;
      }

      // ── Gather ALL attempts for this task (local + VK API) ──
      // VK can have multiple attempts with different branches. An older
      // attempt may have the merged PR while the newest was abandoned.
      const localAttempt = attempts.find((a) => a?.task_id === task.id);
      const allVkAttempts = vkAttempts
        .filter((a) => a?.task_id === task.id)
        .sort(
          (a, b) =>
            new Date(b.created_at).getTime() -
            new Date(a.created_at).getTime(),
        );

      // Build a deduplicated list of all branches + PR numbers to check
      /** @type {Array<{branch?: string, prNumber?: number, attemptId?: string}>} */
      const candidates = [];
      const seenBranches = new Set();

      const addCandidate = (src) => {
        const b = src?.branch;
        const pr =
          src?.pr_number ||
          parsePrNumberFromUrl(src?.pr_url);
        const aid = src?.id; // attempt UUID
        if (b && !seenBranches.has(b)) {
          seenBranches.add(b);
          candidates.push({ branch: b, prNumber: pr || undefined, attemptId: aid });
        } else if (pr && !candidates.some((c) => c.prNumber === pr)) {
          candidates.push({ branch: b, prNumber: pr, attemptId: aid });
        }
      };

      if (localAttempt) addCandidate(localAttempt);
      for (const a of allVkAttempts) addCandidate(a);
      // Also check task-level fields
      addCandidate({
        branch:
          task?.branch || task?.workspace_branch || task?.git_branch,
        pr_number: task?.pr_number,
        pr_url: task?.pr_url,
      });

      if (candidates.length === 0) {
        // ── Only recover idle inprogress tasks — never inreview ──
        // inreview tasks are monitored by merge/conflict checks.
        // inprogress tasks with an active agent should not be touched.
        if (taskStatus !== "inprogress") {
          console.log(
            `[monitor] No attempt found for task "${task.title}" (${task.id.substring(0, 8)}...) in ${taskStatus} — skipping (only idle inprogress tasks are recovered)`,
          );
          continue;
        }

        // Check if an agent is actively working on this task
        const hasActiveAgent =
          task.has_in_progress_attempt === true ||
          !!localAttempt;
        if (hasActiveAgent) {
          console.log(
            `[monitor] Task "${task.title}" (${task.id.substring(0, 8)}...) has active agent — skipping recovery`,
          );
          continue;
        }

        // ── Age-based immediate recovery ──
        // If the task has been stuck for longer than STALE_TASK_AGE_MS
        // with no active agent and no branch/PR, move it to todo immediately.
        const taskAge = getTaskAgeMs(task);
        if (taskAge >= STALE_TASK_AGE_MS) {
          const ageHours = (taskAge / (60 * 60 * 1000)).toFixed(1);
          console.log(
            `[monitor] No attempt found for idle task "${task.title}" (${task.id.substring(0, 8)}...) — stale for ${ageHours}h, moving to todo`,
          );
          const success = await updateTaskStatus(task.id, "todo");
          if (success) {
            movedTodoCount++;
            recoveredTaskNames.push(task.title);
            staleBranchCooldown.delete(task.id);
            console.log(
              `[monitor] ♻️ Recovered idle task "${task.title}" → todo (age-based: ${ageHours}h, no agent, no branch/PR)`,
            );
          }
          continue;
        }

        const prev = staleBranchCooldown.get(task.id);
        const strikes = (prev?.strikes || 0) + 1;
        staleBranchCooldown.set(task.id, { lastCheck: Date.now(), strikes });
        console.log(
          `[monitor] No attempt found for idle task "${task.title}" (${task.id.substring(0, 8)}...) — strike ${strikes}/${STALE_MAX_STRIKES}`,
        );
        if (strikes >= STALE_MAX_STRIKES) {
          const success = await updateTaskStatus(task.id, "todo");
          if (success) {
            movedTodoCount++;
            recoveredTaskNames.push(task.title);
            staleBranchCooldown.delete(task.id);
            console.log(
              `[monitor] ♻️ Recovered idle task "${task.title}" → todo (no branch/PR after ${strikes} checks)`,
            );
          }
        }
        continue;
      }

      if (allVkAttempts.length > 0) {
        const branches = candidates.map((c) => c.branch).filter(Boolean);
        console.log(
          `[monitor] Task "${task.title}": checking ${candidates.length} attempt(s) [${branches.join(", ")}]`,
        );
      }

      // ── Branch-level dedup: skip if ANY branch is already known-merged ──
      const knownBranch = candidates.find(
        (c) => c.branch && mergedBranchCache.has(c.branch),
      );
      if (knownBranch) {
        mergedTaskCache.add(task.id);
        // Cache all branches for this task
        for (const c of candidates) {
          if (c.branch) mergedBranchCache.add(c.branch);
        }
        saveMergedTaskCache();
        void updateTaskStatus(task.id, "done");
        continue;
      }

      // ── Check ALL candidates for a merged PR/branch ──
      let resolved = false;
      let hasOpenPR = false;
      /** @type {Array<{prNumber: number, attemptId?: string, branch?: string}>} */
      const conflictCandidates = [];

      for (const cand of candidates) {
        // Check PR by number first (cheapest)
        if (cand.prNumber) {
          const prInfo = await getPullRequestByNumber(cand.prNumber);
          const isMerged =
            !!prInfo?.mergedAt ||
            (!!prInfo?.merged_at && prInfo.merged_at !== null);
          const prState = prInfo?.state
            ? String(prInfo.state).toUpperCase()
            : "";

          if (isMerged) {
            console.log(
              `[monitor] Task "${task.title}" (${task.id.substring(0, 8)}...) has merged PR #${cand.prNumber}, updating to done`,
            );
            const success = await updateTaskStatus(task.id, "done");
            movedCount++;
            mergedTaskCache.add(task.id);
            for (const c of candidates) {
              if (c.branch) mergedBranchCache.add(c.branch);
            }
            saveMergedTaskCache();
            completedTaskNames.push(task.title);
            if (success) {
              console.log(
                `[monitor] ✅ Moved task "${task.title}" from ${taskStatus} → done`,
              );
            } else {
              console.warn(
                `[monitor] ⚠️ VK update failed for "${task.title}" — cached anyway (PR is merged)`,
              );
            }
            resolved = true;
            break;
          }
          if (prState === "OPEN") {
            hasOpenPR = true;
            // Detect merge conflicts on open PRs
            // gh CLI: mergeable = "CONFLICTING" / "MERGEABLE" / "UNKNOWN"
            // REST API: mergeable = false, mergeable_state = "dirty"
            const isConflicting =
              prInfo?.mergeable === "CONFLICTING" ||
              prInfo?.mergeable === false ||
              prInfo?.mergeable_state === "dirty" ||
              prInfo?.mergeStateStatus === "DIRTY";
            if (isConflicting) {
              conflictCandidates.push({
                prNumber: cand.prNumber,
                attemptId: cand.attemptId,
                branch: cand.branch,
              });
            }
          }
        }

        if (!cand.branch) continue;

        // Throttle between GitHub API calls
        if (MERGE_CHECK_THROTTLE_MS > 0) {
          await new Promise((r) => setTimeout(r, MERGE_CHECK_THROTTLE_MS));
        }

        // Check if the branch has been merged (checks gh + git)
        const merged = await isBranchMerged(cand.branch);
        if (merged) {
          console.log(
            `[monitor] Task "${task.title}" (${task.id.substring(0, 8)}...) has merged branch ${cand.branch}, updating to done`,
          );
          const success = await updateTaskStatus(task.id, "done");
          movedCount++;
          mergedTaskCache.add(task.id);
          for (const c of candidates) {
            if (c.branch) mergedBranchCache.add(c.branch);
          }
          saveMergedTaskCache();
          completedTaskNames.push(task.title);
          if (success) {
            console.log(
              `[monitor] ✅ Moved task "${task.title}" from ${taskStatus} → done`,
            );
          } else {
            console.warn(
              `[monitor] ⚠️ VK update failed for "${task.title}" — cached anyway (branch is merged)`,
            );
          }
          resolved = true;
          break;
        }

        // Branch not merged — look up its open PR and check for conflicts
        if (!cand.prNumber) {
          let branchPr = null;
          if (ghAvailable()) {
            branchPr = await findExistingPrForBranch(cand.branch);
          }
          if (!branchPr) {
            branchPr = await findExistingPrForBranchApi(cand.branch);
          }
          if (branchPr) {
            const bpState = String(branchPr.state).toUpperCase();
            if (bpState === "OPEN") {
              hasOpenPR = true;
              // Fetch full PR info (with mergeable) via number
              const fullPrInfo = await getPullRequestByNumber(branchPr.number);
              const isConflicting =
                fullPrInfo?.mergeable === "CONFLICTING" ||
                fullPrInfo?.mergeable === false ||
                fullPrInfo?.mergeable_state === "dirty" ||
                fullPrInfo?.mergeStateStatus === "DIRTY";
              if (isConflicting) {
                conflictCandidates.push({
                  prNumber: branchPr.number,
                  attemptId: cand.attemptId,
                  branch: cand.branch,
                });
              }
            }
          }
        }
      }

      if (resolved) continue;

      // ── Conflict resolution for open PRs with merge conflicts ──
      if (conflictCandidates.length > 0) {
        const lastConflictCheck = conflictResolutionCooldown.get(task.id);
        const onCooldown =
          lastConflictCheck &&
          Date.now() - lastConflictCheck < CONFLICT_COOLDOWN_MS;
        if (!onCooldown) {
          const cc = conflictCandidates[0];
          // Find the attempt ID — prefer the one on the candidate, else search
          let resolveAttemptId = cc.attemptId;
          if (!resolveAttemptId) {
            const matchAttempt = allVkAttempts.find(
              (a) => a.branch === cc.branch || a.pr_number === cc.prNumber,
            );
            resolveAttemptId = matchAttempt?.id || localAttempt?.id;
          }
          if (resolveAttemptId) {
            const shortId = resolveAttemptId.substring(0, 8);
            console.log(
              `[monitor] ⚠️ Task "${task.title}" PR #${cc.prNumber} has merge conflicts — triggering rebase/resolution`,
            );
            conflictResolutionCooldown.set(task.id, Date.now());
            conflictsTriggered++;
            // Fire-and-forget: let smartPRFlow handle rebase + conflict resolution
            void smartPRFlow(resolveAttemptId, shortId, "conflict");
            if (telegramToken && telegramChatId) {
              void sendTelegramMessage(
                `🔀 PR #${cc.prNumber} for "${task.title}" has merge conflicts — triggering auto-resolution`,
              );
            }
          } else {
            console.warn(
              `[monitor] Task "${task.title}" PR #${cc.prNumber} has conflicts but no attempt ID — cannot trigger resolution`,
            );
          }
        }
      }

      // Task is NOT merged via any attempt — handle accordingly
      if (hasOpenPR && taskStatus !== "inreview") {
        const success = await updateTaskStatus(task.id, "inreview");
        if (success) {
          movedReviewCount++;
          console.log(
            `[monitor] ✅ Moved task "${task.title}" from ${taskStatus} → inreview`,
          );
        }
      } else if (!hasOpenPR) {
        // ── Only recover idle inprogress tasks — never inreview ──
        if (taskStatus !== "inprogress") {
          console.log(
            `[monitor] Task "${task.title}" (${task.id.substring(0, 8)}...): no open PR but status=${taskStatus} — skipping recovery`,
          );
          continue;
        }

        // Check if an agent is actively working on this task
        const hasActiveAgent =
          task.has_in_progress_attempt === true ||
          !!localAttempt;
        if (hasActiveAgent) {
          console.log(
            `[monitor] Task "${task.title}" (${task.id.substring(0, 8)}...): no open PR but agent is active — skipping recovery`,
          );
          continue;
        }

        // Genuinely idle inprogress task with no open PR — recover
        const taskAge = getTaskAgeMs(task);
        if (taskAge >= STALE_TASK_AGE_MS) {
          const ageHours = (taskAge / (60 * 60 * 1000)).toFixed(1);
          console.log(
            `[monitor] Idle task "${task.title}" (${task.id.substring(0, 8)}...): no branch/PR, stale for ${ageHours}h — moving to todo`,
          );
          const success = await updateTaskStatus(task.id, "todo");
          if (success) {
            movedTodoCount++;
            recoveredTaskNames.push(task.title);
            staleBranchCooldown.delete(task.id);
            console.log(
              `[monitor] ♻️ Recovered idle task "${task.title}" → todo (age-based: ${ageHours}h, no agent, no branch/PR)`,
            );
          }
        } else {
          // Not old enough — use the strike-based system
          const prev = staleBranchCooldown.get(task.id);
          const strikes = (prev?.strikes || 0) + 1;
          staleBranchCooldown.set(task.id, { lastCheck: Date.now(), strikes });
          console.log(
            `[monitor] Idle task "${task.title}" (${task.id.substring(0, 8)}...): no branch, no PR (strike ${strikes}/${STALE_MAX_STRIKES})`,
          );
          if (strikes >= STALE_MAX_STRIKES) {
            const success = await updateTaskStatus(task.id, "todo");
            if (success) {
              movedTodoCount++;
              recoveredTaskNames.push(task.title);
              staleBranchCooldown.delete(task.id);
              console.log(
                `[monitor] ♻️ Recovered task "${task.title}" from ${taskStatus} → todo (abandoned — ${strikes} stale checks)`,
              );
            }
          }
        }
      }
    }

    // Send a single aggregated Telegram notification
    if (movedCount > 0 && telegramToken && telegramChatId) {
      if (movedCount <= 3) {
        // Few tasks — list them individually
        for (const name of completedTaskNames) {
          void sendTelegramMessage(`✅ Task completed: "${name}"`);
        }
      } else {
        // Many tasks — send a single summary to avoid spam
        const listed = completedTaskNames
          .slice(0, 5)
          .map((n) => `• ${n}`)
          .join("\n");
        const extra = movedCount > 5 ? `\n…and ${movedCount - 5} more` : "";
        void sendTelegramMessage(
          `✅ ${movedCount} tasks moved to done:\n${listed}${extra}`,
        );
      }
    }

    if (movedCount > 0) {
      console.log(`[monitor] Moved ${movedCount} merged tasks to done status`);
    }
    if (movedReviewCount > 0) {
      console.log(
        `[monitor] Moved ${movedReviewCount} tasks to inreview (PR open)`,
      );
    }
    if (conflictsTriggered > 0) {
      console.log(
        `[monitor] Triggered conflict resolution for ${conflictsTriggered} PR(s)`,
      );
    }
    // Notify about tasks recovered to todo
    if (movedTodoCount > 0) {
      console.log(
        `[monitor] Recovered ${movedTodoCount} abandoned tasks to todo`,
      );
      if (telegramToken && telegramChatId) {
        if (movedTodoCount <= 3) {
          for (const name of recoveredTaskNames) {
            void sendTelegramMessage(
              `♻️ Task recovered to todo (abandoned — no branch/PR): "${name}"`,
            );
          }
        } else {
          const listed = recoveredTaskNames
            .slice(0, 5)
            .map((n) => `• ${n}`)
            .join("\n");
          const extra =
            movedTodoCount > 5 ? `\n…and ${movedTodoCount - 5} more` : "";
          void sendTelegramMessage(
            `♻️ ${movedTodoCount} abandoned tasks recovered to todo:\n${listed}${extra}`,
          );
        }
      }
    }
    return {
      checked: batch.length,
      movedDone: movedCount,
      movedReview: movedReviewCount,
      movedTodo: movedTodoCount,
      conflictsTriggered,
      cached: mergedTaskCache.size,
    };
  } catch (err) {
    console.warn(`[monitor] Error checking merged PRs: ${err.message || err}`);
    return { checked: 0, movedDone: 0, movedReview: 0, movedTodo: 0, error: err };
  }
}

async function reconcileTaskStatuses(reason = "manual") {
  console.log(`[monitor] Reconciling VK tasks (${reason})...`);
  return await checkMergedPRsAndUpdateTasks();
}

// ── Dependabot / Bot PR Auto-Merge ──────────────────────────────────────────

/** Set of PR numbers we've already attempted to merge this session */
const dependabotMergeAttempted = new Set();

/**
 * Check for open Dependabot (or other bot) PRs where all CI checks have passed,
 * and auto-merge them.
 *
 * Flow:
 *   1. `gh pr list` filtered by bot authors
 *   2. For each PR, `gh pr checks` to verify all CI passed
 *   3. `gh pr merge --squash` (or configured method)
 *   4. Notify via Telegram
 */
async function checkAndMergeDependabotPRs() {
  if (!dependabotAutoMerge) return;
  if (!repoSlug) {
    console.warn("[dependabot] auto-merge disabled — no repo slug configured");
    return;
  }

  const authorFilter = dependabotAuthors.map((a) => `author:${a}`).join(" ");

  try {
    // List open PRs by bot authors
    const listCmd = `gh pr list --repo ${repoSlug} --state open --json number,title,author,headRefName,statusCheckRollup --limit 20`;
    const listResult = execSync(listCmd, {
      cwd: repoRoot,
      encoding: "utf8",
      timeout: 30_000,
    }).trim();

    const prs = JSON.parse(listResult || "[]");
    if (prs.length === 0) return;

    // Filter to only bot-authored PRs
    const botPRs = prs.filter((pr) => {
      const login = pr.author?.login || pr.author?.name || "";
      return dependabotAuthors.some(
        (a) =>
          login === a ||
          login === a.replace("app/", "") ||
          a === `app/${login}`,
      );
    });

    if (botPRs.length === 0) return;
    console.log(
      `[dependabot] found ${botPRs.length} bot PR(s): ${botPRs.map((p) => `#${p.number}`).join(", ")}`,
    );

    for (const pr of botPRs) {
      if (dependabotMergeAttempted.has(pr.number)) continue;

      try {
        // Check CI status — all checks must pass
        const checksCmd = `gh pr checks ${pr.number} --repo ${repoSlug} --json name,state,conclusion --required`;
        let checksResult;
        try {
          checksResult = execSync(checksCmd, {
            cwd: repoRoot,
            encoding: "utf8",
            timeout: 15_000,
          }).trim();
        } catch (checksErr) {
          // gh pr checks returns exit code 1 if any check failed/pending
          // Parse the output anyway if available
          checksResult = checksErr.stdout?.trim() || "";
          if (!checksResult) {
            console.log(
              `[dependabot] PR #${pr.number}: checks still pending or failed`,
            );
            continue;
          }
        }

        let checks;
        try {
          checks = JSON.parse(checksResult || "[]");
        } catch {
          // JSON parse failed — might be old gh version, try simpler check
          console.log(
            `[dependabot] PR #${pr.number}: could not parse checks output`,
          );
          continue;
        }

        // All required checks must be in a passing state
        const allPassed =
          checks.length > 0 &&
          checks.every(
            (c) =>
              c.conclusion === "SUCCESS" ||
              c.conclusion === "success" ||
              c.conclusion === "NEUTRAL" ||
              c.conclusion === "neutral" ||
              c.conclusion === "SKIPPED" ||
              c.conclusion === "skipped",
          );

        if (!allPassed) {
          const pending = checks.filter(
            (c) =>
              !c.conclusion ||
              c.state === "PENDING" ||
              c.state === "IN_PROGRESS" ||
              c.state === "QUEUED",
          );
          const failed = checks.filter(
            (c) =>
              c.conclusion === "FAILURE" ||
              c.conclusion === "failure" ||
              c.conclusion === "ERROR" ||
              c.conclusion === "error" ||
              c.conclusion === "TIMED_OUT" ||
              c.conclusion === "timed_out",
          );
          if (failed.length > 0) {
            console.log(
              `[dependabot] PR #${pr.number}: ${failed.length} check(s) failed — skipping`,
            );
            dependabotMergeAttempted.add(pr.number); // don't retry failed
          } else if (pending.length > 0) {
            console.log(
              `[dependabot] PR #${pr.number}: ${pending.length} check(s) still pending`,
            );
          } else if (checks.length === 0) {
            console.log(
              `[dependabot] PR #${pr.number}: no required checks found — waiting`,
            );
          }
          continue;
        }

        // All checks passed — merge!
        console.log(
          `[dependabot] PR #${pr.number}: all ${checks.length} check(s) passed — merging (${dependabotMergeMethod})`,
        );
        dependabotMergeAttempted.add(pr.number);

        const mergeCmd = `gh pr merge ${pr.number} --repo ${repoSlug} --${dependabotMergeMethod} --delete-branch --auto`;
        try {
          execSync(mergeCmd, {
            cwd: repoRoot,
            encoding: "utf8",
            timeout: 30_000,
          });
          console.log(`[dependabot] ✅ PR #${pr.number} merged: ${pr.title}`);
          void sendTelegramMessage(
            `✅ Auto-merged bot PR #${pr.number}: ${pr.title}`,
          );
        } catch (mergeErr) {
          const errMsg = mergeErr.stderr || mergeErr.message || "";
          console.warn(
            `[dependabot] merge failed for PR #${pr.number}: ${errMsg.slice(0, 200)}`,
          );
          // If auto-merge was enabled (queued), that's fine — gh returns success for --auto
          if (errMsg.includes("auto-merge")) {
            console.log(
              `[dependabot] PR #${pr.number}: auto-merge enabled, will merge when protection rules are met`,
            );
            void sendTelegramMessage(
              `🔄 Auto-merge enabled for bot PR #${pr.number}: ${pr.title}`,
            );
          }
        }
      } catch (prErr) {
        console.warn(
          `[dependabot] error processing PR #${pr.number}: ${prErr.message || prErr}`,
        );
      }
    }
  } catch (err) {
    console.warn(`[dependabot] error listing bot PRs: ${err.message || err}`);
  }
}

// ── Merge Strategy Analysis ─────────────────────────────────────────────────

/**
 * Run the Codex-powered merge strategy analysis for a completed task.
 * This is fire-and-forget (void) — it runs async in the background and
 * handles its own errors/notifications.
 *
 * @param {import("./merge-strategy.mjs").MergeContext} ctx
 */
async function runMergeStrategyAnalysis(ctx) {
  const tag = `merge-strategy(${ctx.shortId})`;
  try {
    if (isPrimaryBusy()) {
      console.log(`[${tag}] Codex busy — deferring analysis`);
      // Retry after 60 seconds if Codex frees up
      setTimeout(() => {
        if (!isPrimaryBusy()) void runMergeStrategyAnalysis(ctx);
      }, 60_000);
      return;
    }

    const telegramFn =
      telegramToken && telegramChatId
        ? (msg) => void sendTelegramMessage(msg)
        : null;

    const decision = await analyzeMergeStrategy(ctx, {
      execCodex: execPrimaryPrompt,
      timeoutMs:
        parseInt(process.env.MERGE_STRATEGY_TIMEOUT_MS, 10) || 10 * 60 * 1000,
      logDir,
      onTelegram: telegramFn,
    });

    if (!decision || !decision.success) {
      console.warn(`[${tag}] analysis failed — falling back to manual review`);
      return;
    }

    // ── Act on the decision ──────────────────────────────────────
    switch (decision.action) {
      case "merge_after_ci_pass":
        console.log(`[${tag}] → merge_after_ci_pass`);
        // Nothing extra — the VK cleanup script handles auto-merge after CI
        break;

      case "prompt":
        console.log(
          `[${tag}] → prompt agent: ${(decision.message || "").slice(0, 100)}`,
        );
        if (codexEnabled && !isPrimaryBusy() && decision.message) {
          void execPrimaryPrompt(
            `The merge strategy reviewer has feedback on your work for task "${ctx.taskTitle || ctx.shortId}":\n\n` +
              decision.message +
              `\n\nPlease address this feedback, commit, and push.`,
            { timeoutMs: 15 * 60 * 1000 },
          );
        }
        break;

      case "close_pr":
        console.log(`[${tag}] → close_pr: ${decision.reason || "no reason"}`);
        if (telegramFn) {
          telegramFn(
            `🚫 Merge strategy recommends closing PR for ${ctx.shortId}: ${decision.reason || "no reason given"}`,
          );
        }
        // Don't auto-close — just flag for human. Closing could lose work.
        break;

      case "re_attempt":
        console.log(`[${tag}] → re_attempt: ${decision.reason || "no reason"}`);
        // Trigger a fresh session retry
        if (typeof attemptFreshSessionRetry === "function") {
          const freshStarted = await attemptFreshSessionRetry(
            "merge_strategy_re_attempt",
            decision.reason || "Merge strategy recommended re-attempt",
          );
          if (freshStarted && telegramFn) {
            telegramFn(
              `🔄 Merge strategy recommended re-attempt for ${ctx.shortId}. Fresh session started.`,
            );
          }
        }
        break;

      case "manual_review":
        console.log(
          `[${tag}] → manual_review: ${decision.reason || "no reason"}`,
        );
        // Already notified via Telegram by analyzeMergeStrategy
        break;

      case "wait": {
        const waitSec = decision.seconds || 300;
        console.log(`[${tag}] → wait ${waitSec}s: ${decision.reason || ""}`);
        // Re-run analysis after the wait period
        setTimeout(() => {
          void runMergeStrategyAnalysis({
            ...ctx,
            ciStatus: "re-check",
          });
        }, waitSec * 1000);
        break;
      }

      case "noop":
        console.log(`[${tag}] → noop`);
        break;

      default:
        console.warn(`[${tag}] unknown action: ${decision.action}`);
    }
  } catch (err) {
    console.warn(
      `[${tag}] merge strategy analysis error: ${err.message || err}`,
    );
  }
}

// ── Smart PR creation flow ──────────────────────────────────────────────────

const DEFAULT_TARGET_BRANCH = process.env.VK_TARGET_BRANCH || "origin/main";
const DEFAULT_CODEX_MONITOR_UPSTREAM =
  process.env.CODEX_MONITOR_TASK_UPSTREAM || "origin/ve/codex-monitor-generic";

function normalizeBranchName(value) {
  if (!value) return null;
  const trimmed = String(value).trim();
  return trimmed ? trimmed : null;
}

function extractUpstreamFromText(text) {
  if (!text) return null;
  const match = String(text).match(
    /\b(?:upstream|base|target)(?:_branch| branch)?\s*[:=]\s*([A-Za-z0-9._/-]+)/i,
  );
  if (!match) return null;
  return normalizeBranchName(match[1]);
}

function collectTaskLabels(task) {
  const labels = [];
  if (!task) return labels;
  for (const field of ["labels", "label", "tags", "tag", "categories", "category"]) {
    const value = task[field];
    if (!value) continue;
    if (typeof value === "string") {
      labels.push(value);
      continue;
    }
    if (Array.isArray(value)) {
      for (const item of value) {
        if (!item) continue;
        if (typeof item === "string") labels.push(item);
        else if (item.name) labels.push(item.name);
        else if (item.label) labels.push(item.label);
        else if (item.title) labels.push(item.title);
      }
    }
  }
  if (task.metadata) {
    for (const field of ["labels", "tags"]) {
      const value = task.metadata[field];
      if (!value) continue;
      if (typeof value === "string") labels.push(value);
      else if (Array.isArray(value)) labels.push(...value);
    }
  }
  return labels;
}

function getTaskTextBlob(task) {
  const parts = [];
  if (!task) return "";
  for (const field of ["title", "name", "description", "body", "details", "content"]) {
    const value = task[field];
    if (value) parts.push(value);
  }
  const labels = collectTaskLabels(task);
  if (labels.length) parts.push(labels.join(" "));
  return parts.join("\n");
}

function resolveUpstreamFromTask(task) {
  if (!task) return null;

  const directFields = [
    "target_branch",
    "base_branch",
    "upstream_branch",
    "upstream",
    "target",
    "base",
    "targetBranch",
    "baseBranch",
  ];
  for (const field of directFields) {
    if (task[field]) return normalizeBranchName(task[field]);
  }
  if (task.metadata) {
    for (const field of directFields) {
      if (task.metadata[field]) return normalizeBranchName(task.metadata[field]);
    }
  }

  for (const label of collectTaskLabels(task)) {
    const match = String(label).match(
      /^(?:upstream|base|target)(?:_branch)?[:=]\s*([A-Za-z0-9._/-]+)$/i,
    );
    if (match) return normalizeBranchName(match[1]);
  }

  const fromText = extractUpstreamFromText(getTaskTextBlob(task));
  if (fromText) return fromText;

  const text = getTaskTextBlob(task).toLowerCase();
  if (
    text.includes("codex-monitor") ||
    text.includes("codex monitor") ||
    text.includes("@virtengine/codex-monitor") ||
    text.includes("scripts/codex-monitor")
  ) {
    return DEFAULT_CODEX_MONITOR_UPSTREAM;
  }

  return null;
}

function resolveAttemptTargetBranch(attempt, task) {
  if (attempt) {
    const candidate =
      attempt.target_branch ||
      attempt.targetBranch ||
      attempt.base_branch ||
      attempt.baseBranch ||
      attempt.upstream_branch;
    const normalized = normalizeBranchName(candidate);
    if (normalized) return normalized;
    if (Array.isArray(attempt.repos) && attempt.repos.length) {
      const repoTarget =
        attempt.repos[0].target_branch || attempt.repos[0].targetBranch;
      const repoNorm = normalizeBranchName(repoTarget);
      if (repoNorm) return repoNorm;
    }
  }

  const fromTask = resolveUpstreamFromTask(task);
  if (fromTask) return fromTask;

  return DEFAULT_TARGET_BRANCH;
}

/**
 * Intelligent multi-step PR creation using the VK API:
 *
 *   1. Check branch-status → decide action
 *   2. Stale detection: 0 commits AND far behind → rebase first, archive on error
 *   3. Rebase onto main (resolve conflicts automatically if possible)
 *   4. Create PR via /pr endpoint
 *   5. Distinguish fast-fail (<2s = worktree issue) vs slow-fail (>30s = prepush)
 *   6. On prepush failure → prompt agent to fix lint/test issues and push
 *
 * @param {string} attemptId - Full attempt UUID
 * @param {string} shortId   - Short ID for logging (4-8 chars)
 * @param {string} status    - "completed", "failed", or "no-remote-branch"
 */
async function smartPRFlow(attemptId, shortId, status) {
  const tag = `smartPR(${shortId})`;
  try {
    // ── Step 1: Check branch status ─────────────────────────────
    const branchStatus = await fetchBranchStatus(attemptId);
    if (!branchStatus) {
      console.log(`[monitor] ${tag}: cannot fetch branch-status, skipping`);
      return;
    }

    const { commits_ahead, commits_behind, has_uncommitted_changes } =
      branchStatus;

    // ── Step 2: Stale attempt detection ─────────────────────────
    // 0 commits ahead, 0 uncommitted changes, many behind → stale
    const isStale =
      commits_ahead === 0 && !has_uncommitted_changes && commits_behind > 10;
    if (isStale) {
      console.warn(
        `[monitor] ${tag}: stale attempt — 0 commits, ${commits_behind} behind. Trying rebase first.`,
      );
    }

    // No commits and no changes → archive stale attempt instead of silently skipping
    if (commits_ahead === 0 && !has_uncommitted_changes) {
      console.warn(
        `[monitor] ${tag}: no commits ahead, no changes — archiving stale attempt`,
      );
      await archiveAttempt(attemptId);
      if (telegramToken && telegramChatId) {
        void sendTelegramMessage(
          `🗑️ Archived attempt ${shortId}: no commits, no changes (status=${status}). Task will be reattempted.`,
        );
      }
      return;
    }

    // Uncommitted changes but no commits → agent didn't commit
    if (has_uncommitted_changes && commits_ahead === 0) {
      console.log(
        `[monitor] ${tag}: uncommitted changes but no commits — agent needs to commit first`,
      );
      // Ask the agent to commit via primary agent
      if (!isPrimaryBusy()) {
        void execPrimaryPrompt(
          `Task attempt ${shortId} has uncommitted changes but no commits.\n` +
            `Please navigate to the worktree for this attempt and:\n` +
            `1. Stage all changes: git add -A\n` +
            `2. Create a conventional commit\n` +
            `3. Push and create a PR`,
          { timeoutMs: 10 * 60 * 1000 },
        );
      }
      return;
    }

    // ── Resolve target branch (task-level upstream overrides) ───
    const attempt = await getAttemptInfo(attemptId);
    let taskData = null;
    if (attempt?.task_id) {
      try {
        const taskRes = await fetchVk(`/api/tasks/${attempt.task_id}`);
        if (taskRes?.success && taskRes.data) {
          taskData = taskRes.data;
          attempt.task_title = attempt.task_title || taskRes.data.title;
          attempt.task_description =
            taskRes.data.description || taskRes.data.body || "";
        }
      } catch {
        /* best effort */
      }
    }
    const targetBranch = resolveAttemptTargetBranch(attempt, taskData);

    // ── Step 3: Rebase onto target branch ────────────────────────
    console.log(
      `[monitor] ${tag}: rebasing onto ${targetBranch}...`,
    );
    const rebaseResult = await rebaseAttempt(attemptId, targetBranch);

    if (rebaseResult && !rebaseResult.success) {
      if (isStale) {
        console.warn(
          `[monitor] ${tag}: stale attempt rebase failed — archiving and reattempting next cycle.`,
        );
        await archiveAttempt(attemptId);
        const freshStarted = await attemptFreshSessionRetry(
          "stale_attempt_rebase_failed",
          `Attempt ${shortId} was stale and rebase failed.`,
        );
        if (telegramToken && telegramChatId) {
          const action = freshStarted
            ? "Fresh session started for reattempt."
            : "Will reattempt on next cycle.";
          void sendTelegramMessage(
            `🗑️ Archived stale attempt ${shortId} after failed rebase. ${action}`,
          );
        }
        return;
      }
      const errorData = rebaseResult.error_data;
      // Rebase has conflicts → try auto-resolve
      if (errorData?.type === "merge_conflicts") {
        const files = errorData.conflicted_files || [];
        console.warn(
          `[monitor] ${tag}: rebase conflicts in ${files.join(", ")} — attempting auto-resolve`,
        );
        const resolveResult = await resolveConflicts(attemptId);
        if (resolveResult?.success) {
          console.log(`[monitor] ${tag}: conflicts resolved automatically`);
        } else {
          const attemptInfo = await getAttemptInfo(attemptId);
          const worktreeDir =
            attemptInfo?.worktree_dir || attemptInfo?.worktree || null;
          if (codexResolveConflictsEnabled) {
            console.warn(
              `[monitor] ${tag}: auto-resolve failed — running Codex SDK conflict resolution`,
            );
            const prompt = `You are fixing a git rebase conflict in a Vibe-Kanban worktree.
Worktree: ${worktreeDir || "(unknown)"}
Attempt: ${shortId}
Conflicted files: ${files.join(", ") || "(unknown)"}

Instructions:
1) cd into the worktree directory.
2) Inspect git status and conflicted files.
3) Resolve conflicts, then run: git add -A
4) Continue the rebase (git rebase --continue) if needed.
5) Ensure the branch builds/tests if necessary.
6) Commit if prompted and push the branch.
Return a short summary of what you did.`;
            const codexResult = await runCodexExec(
              prompt,
              worktreeDir || repoRoot,
              conflictResolutionTimeoutMs,
            );
            const logPath = resolve(
              logDir,
              `codex-conflict-${shortId}-${nowStamp()}.log`,
            );
            await writeFile(
              logPath,
              codexResult.output || codexResult.error || "(no output)",
              "utf8",
            );
            if (codexResult.success) {
              console.log(
                `[monitor] ${tag}: Codex conflict resolution succeeded`,
              );
              if (telegramToken && telegramChatId) {
                void sendTelegramMessage(
                  `✅ Codex resolved rebase conflicts for ${shortId}. Log: ${logPath}`,
                );
              }
              return;
            }
            console.warn(
              `[monitor] ${tag}: Codex conflict resolution failed — prompting agent`,
            );
            if (telegramToken && telegramChatId) {
              void sendTelegramMessage(
                `⚠️ Codex failed to resolve conflicts for ${shortId}. Log: ${logPath}`,
              );
            }
          }
          // Auto-resolve failed — ask agent to fix
          console.warn(
            `[monitor] ${tag}: auto-resolve failed — prompting agent`,
          );
          if (telegramToken && telegramChatId) {
            void sendTelegramMessage(
              `⚠️ Attempt ${shortId} has unresolvable rebase conflicts: ${files.join(", ")}`,
            );
          }
          if (!isPrimaryBusy()) {
            void execPrimaryPrompt(
              `Task attempt ${shortId} has rebase conflicts in: ${files.join(", ")}.\n` +
                `Please resolve the conflicts, commit, push, and create a PR.`,
              { timeoutMs: 15 * 60 * 1000 },
            );
          }
          return;
        }
      }
    }

    // ── Step 4: Build PR title & description from VK task ─────

    let prTitle = attempt?.task_title || attempt?.branch || shortId;
    prTitle = prTitle.replace(/\s*\(vibe-kanban\)$/i, "");

    // Build PR description from task description + auto-created footer
    let prDescription = "";
    if (attempt?.task_description) {
      prDescription = attempt.task_description.trim();
      prDescription += `\n\n---\n_Auto-created by codex-monitor (${status})_`;
    } else {
      prDescription = `Auto-created by codex-monitor after ${status} status.`;
    }

    const branchName = attempt?.branch || branchStatus?.branch || null;
    if (attempt?.pr_number || attempt?.pr_url) {
      console.log(
        `[monitor] ${tag}: attempt already linked to PR (${attempt.pr_number || attempt.pr_url}) — skipping`,
      );
      return;
    }
    if (branchName) {
      let existingPr = null;
      if (ghAvailable()) {
        existingPr = await findExistingPrForBranch(branchName);
      }
      if (!existingPr) {
        existingPr = await findExistingPrForBranchApi(branchName);
      }
      if (existingPr) {
        const state = (existingPr.state || "").toUpperCase();
        if (state === "CLOSED" && smartPrAllowRecreateClosed) {
          console.log(
            `[monitor] ${tag}: existing CLOSED PR #${existingPr.number} found, recreating allowed by VE_SMARTPR_ALLOW_RECREATE_CLOSED`,
          );
        } else {
          console.log(
            `[monitor] ${tag}: existing PR #${existingPr.number} (${state}) for ${branchName} — skipping auto-PR`,
          );
          if (telegramToken && telegramChatId) {
            void sendTelegramMessage(
              `⚠️ Auto-PR skipped for ${shortId}: existing PR #${existingPr.number} (${state}) already linked to ${branchName}.`,
            );
          }
          return;
        }
      }
    }

    // ── Step 5: Create PR via VK API ────────────────────────────
    console.log(`[monitor] ${tag}: creating PR "${prTitle}"...`);
    const prResult = await createPRViaVK(attemptId, {
      title: prTitle,
      description: prDescription,
      draft: false,
      base: targetBranch,
    });

    if (prResult?.success) {
      const prUrl = prResult.data?.url || prResult.data?.html_url || "";
      const prNum = prResult.data?.number || null;
      console.log(
        `[monitor] ${tag}: PR created successfully${prUrl ? " — " + prUrl : ""}`,
      );
      if (telegramToken && telegramChatId) {
        void sendTelegramMessage(
          `✅ Auto-created PR for ${shortId}${prUrl ? ": " + prUrl : ""}`,
        );
      }

      // ── Step 5b: Merge strategy analysis (Codex-powered) ─────
      if (codexAnalyzeMergeStrategy && !isPrimaryBusy()) {
        void runMergeStrategyAnalysis({
          attemptId,
          shortId,
          status,
          prTitle,
          prNumber: prNum,
          prUrl,
          prState: "open",
          branch: branchName,
          commitsAhead: branchStatus.commits_ahead,
          commitsBehind: branchStatus.commits_behind,
          taskTitle: attempt?.task_title,
          taskDescription: attempt?.task_description,
          worktreeDir: attempt?.worktree_dir || attempt?.worktree || null,
        });
      }

      return;
    }

    // ── Step 6: Handle PR creation failure ──────────────────────
    const elapsed = prResult._elapsedMs || 0;
    const isFastFail = elapsed < 2000; // < 2s = instant (worktree/config issue)

    if (prResult.error === "repo_id_missing") {
      console.warn(
        `[monitor] ${tag}: PR creation failed — repo_id missing (VK config/API issue)`,
      );
      if (telegramToken && telegramChatId) {
        void sendTelegramMessage(
          `⚠️ Auto-PR for ${shortId} failed: repo_id missing. Check VK_BASE_URL/VK_REPO_ID.`,
        );
      }
      return;
    }

    if (isFastFail) {
      // Instant failure — worktree issue, ask agent to handle everything
      console.warn(
        `[monitor] ${tag}: PR creation fast-failed (${elapsed}ms) — worktree/config issue`,
      );
      if (telegramToken && telegramChatId) {
        void sendTelegramMessage(
          `⚠️ Auto-PR for ${shortId} fast-failed (${elapsed}ms) — likely worktree issue. Prompting agent.`,
        );
      }
      if (!isPrimaryBusy()) {
        void execPrimaryPrompt(
          `Task attempt ${shortId} needs to create a PR but the automated PR creation ` +
            `failed instantly (worktree or config issue).\n` +
            `Branch: ${attempt?.branch || shortId}\n\n` +
            `Please:\n` +
            `1. Navigate to the worktree\n` +
            `2. Ensure git status is clean and commits exist\n` +
            `3. Run: git push --set-upstream origin ${attempt?.branch || shortId}\n` +
            `4. Create a PR targeting main`,
          { timeoutMs: 15 * 60 * 1000 },
        );
      }
    } else {
      // Slow failure — prepush hooks failed (lint/test/build)
      console.warn(
        `[monitor] ${tag}: PR creation slow-failed (${Math.round(elapsed / 1000)}s) — prepush hook failure`,
      );
      if (telegramToken && telegramChatId) {
        void sendTelegramMessage(
          `⚠️ Auto-PR for ${shortId} failed after ${Math.round(elapsed / 1000)}s (prepush hooks). Prompting agent to fix.`,
        );
      }
      if (!isPrimaryBusy()) {
        void execPrimaryPrompt(
          `Task attempt ${shortId}: the prepush hooks (lint/test/build) failed ` +
            `when trying to create a PR.\n` +
            `Branch: ${attempt?.branch || shortId}\n\n` +
            `Please:\n` +
            `1. Navigate to the worktree for this branch\n` +
            `2. Fix any lint, test, or build errors\n` +
            `3. Commit the fixes\n` +
            `4. Rebase onto main: git pull --rebase origin main\n` +
            `5. Push: git push --set-upstream origin ${attempt?.branch || shortId}\n` +
            `6. Create a PR targeting main`,
          { timeoutMs: 15 * 60 * 1000 },
        );
      }
    }
  } catch (err) {
    console.warn(`[monitor] ${tag}: error — ${err.message || err}`);
  }
}

// Tracks attempts we've already tried smartPR for (dedup)
const smartPRAttempted = new Set();

/**
 * Check if a shortId (or a prefix/suffix of it) is already tracked.
 * Handles the case where the orchestrator emits different-length prefixes
 * for the same attempt UUID (e.g., "2f71" and "2f7153e7").
 */
function isSmartPRAttempted(shortId) {
  if (smartPRAttempted.has(shortId)) return true;
  for (const existing of smartPRAttempted) {
    if (existing.startsWith(shortId) || shortId.startsWith(existing)) {
      return true;
    }
  }
  return false;
}

/**
 * Resolve a short (4-8 char) attempt ID prefix to the full UUID and trigger
 * smartPRFlow. De-duplicated so each attempt is only processed once per
 * monitor lifetime.
 */
async function resolveAndTriggerSmartPR(shortId, status) {
  if (isSmartPRAttempted(shortId)) return;
  smartPRAttempted.add(shortId);

  try {
    const statusData = await readStatusData();
    const attempts = statusData?.active_attempts || [];
    const match = attempts.find((a) => a.id?.startsWith(shortId));
    if (!match) {
      // Try the full list via VK API
      const allAttempts = await fetchVk(
        "/api/task-attempts?status=review,error",
      );
      const vkMatch =
        allAttempts?.data?.find((a) => a.id?.startsWith(shortId)) || null;
      if (!vkMatch) {
        console.log(
          `[monitor] smartPR(${shortId}): attempt not found in status or VK data`,
        );
        return;
      }
      await smartPRFlow(vkMatch.id, shortId, status);
      return;
    }
    await smartPRFlow(match.id, shortId, status);
  } catch (err) {
    console.warn(`[monitor] resolveSmartPR(${shortId}): ${err.message || err}`);
  }
}

const errorQueue = [];

function queueErrorMessage(line) {
  errorQueue.push(stripAnsi(line));
  if (errorQueue.length >= 3) {
    void flushErrorQueue();
  }
}

async function flushErrorQueue() {
  if (!telegramToken || !telegramChatId) {
    return;
  }
  if (errorQueue.length === 0) {
    return;
  }
  const lines = errorQueue.splice(0, errorQueue.length);
  const message = [`${projectName} Orchestrator Error`, ...lines].join("\n");
  await sendTelegramMessage(message);
}

function notifyMerge(line) {
  const match = line.match(/PR\s+#(\d+)/i);
  if (!match) {
    return;
  }
  const pr = match[1];
  if (mergeNotified.has(pr)) {
    return;
  }
  mergeNotified.add(pr);
  pendingMerges.add(pr);
}

function notifyMergeFailure(line) {
  if (!telegramToken || !telegramChatId) {
    return;
  }
  const match = line.match(
    /Merge notify: PR #(\d+)\s+stage=([^\s]+)\s+category=([^\s]+)\s+action=([^\s]+)\s+reason=(.+)$/i,
  );
  if (!match) {
    return;
  }
  const pr = match[1];
  const stage = match[2];
  const category = match[3];
  const action = match[4];
  const reason = match[5];
  if (stage !== "manual_review") {
    return;
  }
  if (mergeFailureNotified.has(pr)) {
    return;
  }
  mergeFailureNotified.set(pr, Date.now());
  const message = [
    `Merge failed for PR #${pr} (${stage})`,
    `Category: ${category}`,
    `Action: ${action}`,
    `Reason: ${reason}`,
    `${repoUrlBase}/pull/${pr}`,
  ].join("\n");
  void sendTelegramMessage(message);
}

async function flushMergeNotifications() {
  if (!telegramToken || !telegramChatId) {
    return;
  }
  if (pendingMerges.size === 0) {
    return;
  }
  const merged = Array.from(pendingMerges);
  pendingMerges.clear();
  const formatted = merged
    .map((pr) => `#${pr} ${repoUrlBase}/pull/${pr}`)
    .join(", ");
  const message = `Merged PRs: ${formatted}`;
  await sendTelegramMessage(message);
}

async function readStatusData() {
  try {
    const raw = await readFile(statusPath, "utf8");
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

async function readStatusSummary() {
  try {
    const status = await readStatusData();
    if (!status) {
      return {
        text: `${projectName} Orchestrator Update\nStatus: unavailable (missing status file)`,
        parseMode: null,
      };
    }

    const counts = status.counts || {};
    const submitted = Array.isArray(status.submitted_tasks)
      ? status.submitted_tasks
      : [];
    const completed = Array.isArray(status.completed_tasks)
      ? status.completed_tasks
      : [];
    const followups = Array.isArray(status.followup_events)
      ? status.followup_events
      : [];
    const copilotRequests = Array.isArray(status.copilot_requests)
      ? status.copilot_requests
      : [];
    const attempts = status.attempts || {};
    const manualReviewTasks = Array.isArray(status.manual_review_tasks)
      ? status.manual_review_tasks
      : [];

    const now = Date.now();
    const intervalMs = telegramIntervalMin * 60 * 1000;
    const cutoff = now - intervalMs;

    const recentSubmitted = submitted.filter((item) => {
      if (!item.submitted_at) {
        return false;
      }
      const ts = Date.parse(item.submitted_at);
      return Number.isFinite(ts) && ts >= cutoff;
    });

    const recentCompleted = completed.filter((item) => {
      if (!item.completed_at) {
        return false;
      }
      const ts = Date.parse(item.completed_at);
      return Number.isFinite(ts) && ts >= cutoff;
    });

    const recentFollowups = followups.filter((item) => {
      if (!item.occurred_at) {
        return false;
      }
      const ts = Date.parse(item.occurred_at);
      return Number.isFinite(ts) && ts >= cutoff;
    });

    const recentCopilot = copilotRequests.filter((item) => {
      if (!item.occurred_at) {
        return false;
      }
      const ts = Date.parse(item.occurred_at);
      return Number.isFinite(ts) && ts >= cutoff;
    });

    const manualReviewLines = manualReviewTasks.length
      ? manualReviewTasks.map((taskId) => {
          const attempt = Object.values(attempts).find(
            (item) =>
              item &&
              item.task_id === taskId &&
              item.status === "manual_review",
          );
          if (attempt && attempt.pr_number) {
            const prNumber = `#${attempt.pr_number}`;
            return `- ${formatHtmlLink(
              `${repoUrlBase}/pull/${attempt.pr_number}`,
              prNumber,
            )}`;
          }
          return `- ${escapeHtml(taskId)}`;
        })
      : ["- none"];

    const createdLines = recentSubmitted.length
      ? recentSubmitted.map((item) => {
          const title = item.task_title || item.task_id || "(task)";
          const link = item.task_url
            ? formatHtmlLink(item.task_url, title)
            : escapeHtml(title);
          return `- ${link}`;
        })
      : ["- none"];

    const mergedLines = recentCompleted.length
      ? recentCompleted.map((item) => {
          const prNumber = item.pr_number ? `#${item.pr_number}` : "";
          const title = item.pr_title || prNumber || "(PR)";
          const link = item.pr_url
            ? formatHtmlLink(item.pr_url, title)
            : escapeHtml(title);
          const suffix =
            prNumber && !title.includes(prNumber) ? ` (${prNumber})` : "";
          return `- ${link}${suffix}`;
        })
      : ["- none"];

    const followupLines = recentFollowups.length
      ? recentFollowups.map((item) => {
          const title = item.task_title || item.task_id || "(task)";
          const link = item.task_url
            ? formatHtmlLink(item.task_url, title)
            : escapeHtml(title);
          const reason = item.reason ? `: ${escapeHtml(item.reason)}` : "";
          return `- ${link}${reason}`;
        })
      : ["- none"];

    const copilotLines = recentCopilot.length
      ? recentCopilot.map((item) => {
          const prNumber = item.pr_number ? `#${item.pr_number}` : "";
          const title = item.pr_title || prNumber || "(PR)";
          const link = item.pr_url
            ? formatHtmlLink(item.pr_url, title)
            : escapeHtml(title);
          const reason = item.reason ? `: ${escapeHtml(item.reason)}` : "";
          return `- ${link}${reason}`;
        })
      : ["- none"];

    const running = counts.running ?? 0;
    const review = counts.review ?? 0;
    const error = counts.error ?? 0;
    const manualReview = counts.manual_review ?? 0;

    // Success rate metrics
    const sm = status.success_metrics || {};
    const firstShot = sm.first_shot_success ?? 0;
    const neededFix = sm.needed_fix ?? 0;
    const failed = sm.failed ?? 0;
    const firstShotRate = sm.first_shot_rate ?? 0;
    const totalDecided = firstShot + neededFix + failed;
    const successLine =
      totalDecided > 0
        ? `First-shot: ${firstShotRate}% (${firstShot}/${totalDecided}) | Fix: ${neededFix} | Failed: ${failed}`
        : "No completed tasks yet";

    const message = [
      `${projectName} Orchestrator ${telegramIntervalMin}-min Update`,
      `New tasks created (${recentSubmitted.length}):`,
      ...createdLines,
      `Merged tasks (${recentCompleted.length}):`,
      ...mergedLines,
      `Task follow-ups (${recentFollowups.length}):`,
      ...followupLines,
      `Copilot triggered (${recentCopilot.length}):`,
      ...copilotLines,
      `Manual review (${manualReviewTasks.length}):`,
      ...manualReviewLines,
      `Counts: running=${running}, review=${review}, error=${error}, manual_review=${manualReview}, conflict_resolving=${conflictResolutionCooldown.size}`,
      `Success: ${successLine}`,
    ].join("\n");

    return { text: message, parseMode: "HTML" };
  } catch (err) {
    return {
      text: `${projectName} Orchestrator Update\nStatus: unavailable (missing status file)`,
      parseMode: null,
    };
  }
}

async function readPlannerState() {
  try {
    const raw = await readFile(plannerStatePath, "utf8");
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

async function writePlannerState(nextState) {
  await ensureLogDir();
  await writeFile(plannerStatePath, JSON.stringify(nextState, null, 2), "utf8");
}

async function updatePlannerState(patch) {
  const current = (await readPlannerState()) || {};
  const merged = { ...current, ...patch };
  await writePlannerState(merged);
  return merged;
}

function isPlannerDeduped(state, now) {
  if (!state || !state.last_triggered_at) {
    return false;
  }
  // Only dedup if the last run was successful — failed/skipped runs
  // should not block subsequent attempts
  if (!state.last_success_at) {
    return false;
  }
  const last = Date.parse(state.last_success_at);
  if (!Number.isFinite(last)) {
    return false;
  }
  return now - last < plannerDedupMs;
}

async function maybeTriggerTaskPlanner(reason, details) {
  if (plannerMode === "disabled") {
    console.log(`[monitor] task planner skipped: mode=disabled`);
    return;
  }
  if (plannerMode === "codex-sdk" && !codexEnabled) {
    console.log(`[monitor] task planner skipped: codex-sdk mode but Codex disabled`);
    return;
  }
  if (plannerTriggered) {
    console.log(`[monitor] task planner skipped: already running`);
    return;
  }
  const now = Date.now();
  const state = await readPlannerState();
  if (isPlannerDeduped(state, now)) {
    const lastAt = state?.last_triggered_at || "unknown";
    console.log(`[monitor] task planner skipped: deduped (last triggered ${lastAt})`);
    return;
  }
  try {
    const result = await triggerTaskPlanner(reason, details);
    console.log(`[monitor] task planner result: ${result?.status || "unknown"} (${reason})`);
  } catch (err) {
    // Auto-triggered planner failures are non-fatal — already logged/notified by triggerTaskPlanner
    console.warn(
      `[monitor] auto-triggered planner failed: ${err.message || err}`,
    );
  }
}

async function sendTelegramMessage(text, options = {}) {
  const targetChatId = options.chatId ?? telegramChatId;
  if (!telegramToken || !targetChatId) {
    return;
  }
  const rawDedupKey = options.dedupKey ?? String(text || "").trim();
  // Use fuzzy normalization so structural duplicates with different numbers match
  const dedupKey = options.exactDedup
    ? rawDedupKey
    : normalizeDedupKey(rawDedupKey);
  if (dedupKey && !options.skipDedup) {
    const now = Date.now();
    const last = telegramDedup.get(dedupKey) || 0;
    if (now - last < 5 * 60 * 1000) {
      return;
    }
    telegramDedup.set(dedupKey, now);
  }

  // Always record to history ring buffer (even deduped messages are useful context)
  pushTelegramHistory(String(text || ""));

  // Determine priority based on message content
  const textLower = String(text || "").toLowerCase();
  let priority = 4; // default: info
  let category = "general";

  // Positive signals override negative keyword matches — a "✅ Task completed"
  // message should never be classified as an error even when the task title
  // happens to contain words like "error" or "failed".
  const isPositive =
    textLower.includes("✅") ||
    textLower.includes("task completed") ||
    textLower.includes("branch merged") ||
    textLower.includes("pr merged");

  // Priority 1: Critical/Fatal
  if (
    !isPositive &&
    (textLower.includes("fatal") ||
      textLower.includes("critical") ||
      textLower.includes("🔥"))
  ) {
    priority = 1;
    category = "critical";
  }
  // Priority 2: Errors
  else if (
    !isPositive &&
    (textLower.includes("error") ||
      textLower.includes("failed") ||
      textLower.includes("❌") ||
      textLower.includes("auto-fix gave up"))
  ) {
    priority = 2;
    category = "error";
  }
  // Priority 3: Warnings
  else if (
    !isPositive &&
    (textLower.includes("warning") || textLower.includes("⚠️"))
  ) {
    priority = 3;
    category = "warning";
  }
  // Priority 4: Info (default)
  else {
    // Categorize info messages
    if (textLower.includes("pr") || textLower.includes("pull request")) {
      category = "pr";
    } else if (textLower.includes("task") || textLower.includes("completed")) {
      category = "task";
    } else if (textLower.includes("codex") || textLower.includes("analysis")) {
      category = "analysis";
    } else if (
      textLower.includes("auto-created") ||
      textLower.includes("merged")
    ) {
      category = "git";
    }
  }

  // Route through batching system — apply verbosity filter first.
  // minimal: only priority 1-2 (critical + error)
  // summary: priority 1-4 (everything except debug) — DEFAULT
  // detailed: priority 1-5 (everything)
  const maxPriority =
    telegramVerbosity === "minimal"
      ? 2
      : telegramVerbosity === "detailed"
        ? 5
        : 4;
  if (priority > maxPriority) return; // filtered out by verbosity setting

  return notify(text, priority, {
    category,
    silent: options.silent,
    data: { parseMode: options.parseMode, chatId: targetChatId },
  });
}

function enqueueTelegramCommand(handler) {
  telegramCommandQueue.push(handler);
  void drainTelegramCommandQueue();
}

function drainTelegramCommandQueue() {
  while (
    telegramCommandActive < telegramCommandConcurrency &&
    telegramCommandQueue.length > 0
  ) {
    const job = telegramCommandQueue.shift();
    if (!job) {
      continue;
    }
    telegramCommandActive += 1;
    Promise.resolve()
      .then(job)
      .catch((err) => {
        console.warn(
          `[monitor] telegram command handler failed: ${err?.message || err}`,
        );
      })
      .finally(() => {
        telegramCommandActive -= 1;
        setImmediate(() => drainTelegramCommandQueue());
      });
  }
}

function normalizeTelegramCommand(text) {
  if (!text) {
    return null;
  }
  const trimmed = String(text).trim();
  if (!trimmed.startsWith("/")) {
    return null;
  }
  const [raw, ...rest] = trimmed.split(/\s+/);
  const command = raw.split("@")[0].toLowerCase();
  return { command, args: rest.join(" ") };
}

function isAllowedTelegramChat(chatId) {
  if (!telegramChatId) {
    return true;
  }
  return String(chatId) === String(telegramChatId);
}

function limitLines(lines, limit = 8) {
  if (lines.length <= limit) {
    return lines;
  }
  const remaining = lines.length - limit;
  return [...lines.slice(0, limit), `- ...and ${remaining} more`];
}

function buildVkTaskUrl(taskId, projectId) {
  if (!taskId) {
    return null;
  }
  const template = String(vkTaskUrlTemplate || "").trim();
  if (template) {
    return template
      .replace("{projectId}", projectId || "")
      .replace("{taskId}", taskId);
  }
  const base = String(vkPublicUrl || vkEndpointUrl || "").replace(/\/+$/, "");
  if (!base || !projectId) {
    return null;
  }
  return `${base}/local-projects/${projectId}/tasks/${taskId}`;
}

function formatTaskLink(item) {
  const title = item.task_title || item.task_id || "(task)";
  if (item.task_url) {
    return formatHtmlLink(item.task_url, title);
  }
  return escapeHtml(title);
}

function formatAttemptLine(attempt) {
  if (!attempt) {
    return null;
  }
  const taskId = attempt.task_id ? escapeHtml(attempt.task_id) : "(task)";
  const branch = attempt.branch ? ` (${escapeHtml(attempt.branch)})` : "";
  const status = attempt.status ? ` — ${escapeHtml(attempt.status)}` : "";
  if (attempt.pr_number) {
    const prLabel = `#${attempt.pr_number}`;
    const prLink = formatHtmlLink(
      `${repoUrlBase}/pull/${attempt.pr_number}`,
      prLabel,
    );
    return `- ${taskId} ${prLink}${branch}${status}`;
  }
  return `- ${taskId}${branch}${status}`;
}

async function buildTasksResponse() {
  const status = await readStatusData();
  if (!status) {
    return {
      text: "Status unavailable (missing status file).",
      parseMode: null,
    };
  }

  const counts = status.counts || {};
  const attempts = status.attempts || {};
  const runningAttempts = Object.values(attempts).filter(
    (attempt) => attempt && attempt.status === "running",
  );

  const reviewTasks = Array.isArray(status.review_tasks)
    ? status.review_tasks
    : [];
  const errorTasks = Array.isArray(status.error_tasks)
    ? status.error_tasks
    : [];
  const manualReviewTasks = Array.isArray(status.manual_review_tasks)
    ? status.manual_review_tasks
    : [];
  const submitted = Array.isArray(status.submitted_tasks)
    ? status.submitted_tasks
    : [];

  const runningLines = limitLines(
    runningAttempts
      .map((attempt) => formatAttemptLine(attempt))
      .filter(Boolean),
  );
  const submittedLines = limitLines(
    submitted.map((item) => `- ${formatTaskLink(item)}`),
  );

  const reviewLines = reviewTasks.length
    ? limitLines(reviewTasks.map((taskId) => `- ${escapeHtml(taskId)}`))
    : ["- none"];
  const errorLines = errorTasks.length
    ? limitLines(errorTasks.map((taskId) => `- ${escapeHtml(taskId)}`))
    : ["- none"];
  const manualLines = manualReviewTasks.length
    ? limitLines(manualReviewTasks.map((taskId) => `- ${escapeHtml(taskId)}`))
    : ["- none"];

  const message = [
    `${projectName} Task Snapshot`,
    `Counts: running=${counts.running ?? 0}, review=${counts.review ?? 0}, error=${counts.error ?? 0}, manual_review=${counts.manual_review ?? 0}`,
    `Backlog remaining: ${status.backlog_remaining ?? 0}`,
    "Running attempts:",
    ...(runningLines.length ? runningLines : ["- none"]),
    "Recently submitted:",
    ...(submittedLines.length ? submittedLines : ["- none"]),
    "Needs review:",
    ...reviewLines,
    "Errors:",
    ...errorLines,
    "Manual review:",
    ...manualLines,
  ].join("\n");

  return { text: message, parseMode: "HTML" };
}

async function buildAgentResponse() {
  const status = await readStatusData();
  const attempts = status?.attempts || {};
  const runningAttempts = Object.values(attempts).filter(
    (attempt) => attempt && attempt.status === "running",
  );
  const activeLines = limitLines(
    runningAttempts
      .map((attempt) => formatAttemptLine(attempt))
      .filter(Boolean),
  );
  const orchestratorState = currentChild
    ? `Orchestrator running (pid ${currentChild.pid}).`
    : "Orchestrator not running.";
  const message = [
    `${projectName} Agent Status`,
    orchestratorState,
    `Active attempts: ${runningAttempts.length}`,
    ...(activeLines.length ? activeLines : ["- none"]),
  ].join("\n");
  return { text: message, parseMode: "HTML" };
}

async function buildBackgroundResponse() {
  const vkOnline = await isVibeKanbanOnline();
  const now = Date.now();
  const halted =
    now < orchestratorHaltedUntil
      ? `halted until ${new Date(orchestratorHaltedUntil).toISOString()}`
      : "active";
  const safeMode =
    now < monitorSafeModeUntil
      ? `safe-mode until ${new Date(monitorSafeModeUntil).toISOString()}`
      : "normal";
  const message = [
    `${projectName} Background Status`,
    currentChild
      ? `Orchestrator: running (pid ${currentChild.pid})`
      : "Orchestrator: stopped",
    `Monitor state: ${halted}, ${safeMode}`,
    `Vibe-kanban: ${vkOnline ? "online" : "unreachable"}`,
  ].join("\n");
  return { text: message, parseMode: null };
}

async function buildHealthResponse() {
  const status = await readStatusData();
  const updatedAt = status?.updated_at
    ? new Date(status.updated_at).toISOString()
    : "unknown";
  const vkOnline = await isVibeKanbanOnline();
  const message = [
    `${projectName} Health`,
    `Orchestrator: ${currentChild ? "running" : "stopped"}`,
    `Status updated: ${updatedAt}`,
    `Vibe-kanban: ${vkOnline ? "online" : "unreachable"}`,
  ].join("\n");
  return { text: message, parseMode: null };
}

async function handleTelegramUpdate(update) {
  if (!update) {
    return;
  }
  const message =
    update.message || update.edited_message || update.callback_query?.message;
  if (!message) {
    return;
  }
  const chatId = message.chat?.id;
  if (!chatId || !isAllowedTelegramChat(chatId)) {
    return;
  }
  const parsed = normalizeTelegramCommand(message.text || "");
  if (!parsed) {
    return;
  }

  let response = null;
  switch (parsed.command) {
    case "/status":
      response = await readStatusSummary();
      break;
    case "/tasks":
      response = await buildTasksResponse();
      break;
    case "/agent":
      response = await buildAgentResponse();
      break;
    case "/background":
      response = await buildBackgroundResponse();
      break;
    case "/health":
      response = await buildHealthResponse();
      break;
    case "/help":
    case "/start":
      response = {
        text: [
          `${projectName} Command Help`,
          "/status — summary snapshot",
          "/tasks — task breakdown",
          "/agent — active agent status",
          "/background — monitor status",
          "/health — service health",
        ].join("\n"),
        parseMode: null,
      };
      break;
    default:
      response = {
        text: "Unknown command. Send /help for available commands.",
        parseMode: null,
      };
      break;
  }

  if (!response || !response.text) {
    return;
  }

  await sendTelegramMessage(response.text, {
    chatId,
    parseMode: response.parseMode,
    disablePreview: true,
    skipDedup: true,
  });
}

async function fetchTelegramUpdates() {
  const url = `https://api.telegram.org/bot${telegramToken}/getUpdates`;
  const params = new URLSearchParams({
    offset: String(telegramUpdateOffset),
    timeout: String(Math.max(5, telegramCommandPollTimeoutSec)),
    limit: String(Math.max(1, telegramCommandMaxBatch)),
  });

  const controller = new AbortController();
  const timeoutMs = (telegramCommandPollTimeoutSec + 5) * 1000;
  const timeout = setTimeout(() => controller.abort(), timeoutMs);
  try {
    const res = await fetch(`${url}?${params.toString()}`, {
      signal: controller.signal,
    });
    if (!res.ok) {
      const body = await res.text();
      console.warn(
        `[monitor] telegram getUpdates failed: ${res.status} ${body}`,
      );
      if (res.status === 409) {
        telegramCommandEnabled = false;
        await releaseTelegramPollLock();
      }
      return [];
    }
    const data = await res.json();
    if (!data.ok || !Array.isArray(data.result)) {
      return [];
    }
    return data.result;
  } catch (err) {
    if (err?.name !== "AbortError") {
      console.warn(
        `[monitor] telegram getUpdates error: ${err?.message || err}`,
      );
    }
    return [];
  } finally {
    clearTimeout(timeout);
  }
}

async function pollTelegramCommands() {
  if (shuttingDown) {
    telegramCommandPolling = false;
    return;
  }
  if (!telegramCommandEnabled) {
    telegramCommandPolling = false;
    return;
  }
  try {
    const updates = await fetchTelegramUpdates();
    if (updates.length) {
      for (const update of updates) {
        if (typeof update.update_id === "number") {
          telegramUpdateOffset = update.update_id + 1;
        }
        enqueueTelegramCommand(async () => {
          try {
            await handleTelegramUpdate(update);
          } catch (err) {
            const message =
              err && err.message ? err.message : String(err || "unknown error");
            console.warn(`[monitor] telegram command crashed: ${message}`);
            const chatId = update.message?.chat?.id;
            if (chatId && isAllowedTelegramChat(chatId)) {
              await sendTelegramMessage(`Command failed: ${message}`, {
                chatId,
                skipDedup: true,
              });
            }
          }
        });
      }
    }
    const delayMs = updates.length ? 0 : 1000;
    setTimeout(pollTelegramCommands, delayMs);
  } catch (err) {
    const message = err && err.message ? err.message : String(err);
    console.warn(`[monitor] telegram command poll error: ${message}`);
    setTimeout(pollTelegramCommands, 3000);
  }
}

function startTelegramCommandListener() {
  if (!telegramToken || !telegramCommandEnabled) {
    return;
  }
  if (telegramCommandPolling) {
    return;
  }
  void acquireTelegramPollLock("monitor").then((ok) => {
    if (!ok) {
      telegramCommandEnabled = false;
      return;
    }
    telegramCommandPolling = true;
    void pollTelegramCommands();
  });
}

async function startTelegramNotifier() {
  if (telegramNotifierInterval) {
    clearInterval(telegramNotifierInterval);
    telegramNotifierInterval = null;
  }
  if (telegramNotifierTimeout) {
    clearTimeout(telegramNotifierTimeout);
    telegramNotifierTimeout = null;
  }
  if (!telegramToken || !telegramChatId) {
    console.warn(
      "[monitor] telegram notifier disabled (missing TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID)",
    );
    return;
  }
  if (!Number.isFinite(telegramIntervalMin) || telegramIntervalMin <= 0) {
    console.warn("[monitor] telegram notifier disabled (invalid interval)");
    return;
  }
  const intervalMs = telegramIntervalMin * 60 * 1000;
  const sendUpdate = async () => {
    const summary = await readStatusSummary();
    if (summary && summary.text) {
      await sendTelegramMessage(summary.text, {
        parseMode: summary.parseMode,
        disablePreview: true,
      });
    }
    await flushMergeNotifications();
    await checkStatusMilestones();
  };

  // Suppress "Notifier started" message on rapid restarts (e.g. code-change restarts).
  // If the last start was <60s ago, skip the notification — just log locally.
  const lastStartPath = resolve(
    repoRoot,
    ".cache",
    "ve-last-notifier-start.txt",
  );
  let suppressStartup = isSelfRestart;
  if (!suppressStartup) {
    try {
      const prev = await readFile(lastStartPath, "utf8");
      const elapsed = Date.now() - Number(prev);
      if (elapsed < 60_000) suppressStartup = true;
    } catch {
      /* first start or missing file */
    }
  }
  await writeFile(lastStartPath, String(Date.now())).catch(() => {});

  if (suppressStartup) {
    console.log(
      `[monitor] notifier restarted (suppressed telegram notification — rapid restart)`,
    );
  } else {
    void sendTelegramMessage(`${projectName} Orchestrator Notifier started.`);
  }
  telegramNotifierTimeout = setTimeout(sendUpdate, intervalMs);
  telegramNotifierInterval = setInterval(sendUpdate, intervalMs);
}

async function checkStatusMilestones() {
  const status = await readStatusData();
  if (!status) {
    return;
  }
  const counts = status.counts || {};
  const backlogRemaining = status.backlog_remaining ?? 0;
  const running = counts.running ?? 0;
  const review = counts.review ?? 0;
  const error = counts.error ?? 0;
  const maxParallel = getMaxParallelFromArgs(scriptArgs) || running || 1;
  const backlogPerCapita =
    maxParallel > 0 ? backlogRemaining / maxParallel : backlogRemaining;
  const idleSlots = Math.max(0, maxParallel - running);

  if (
    !allCompleteNotified &&
    backlogRemaining === 0 &&
    running === 0 &&
    review === 0 &&
    error === 0
  ) {
    allCompleteNotified = true;
    await sendTelegramMessage(
      "All tasks completed. Orchestrator backlog is empty.",
    );
    await maybeTriggerTaskPlanner("backlog-empty", {
      backlogRemaining,
      backlogPerCapita,
      running,
      review,
      error,
      idleSlots,
      maxParallel,
    });
    return;
  }

  // Planner triggers: reset notification flags each cycle so we can
  // re-trigger if conditions persist and dedup window has passed.
  // The dedup state file prevents rapid re-triggering (default 6h).
  const plannerConditionsMet =
    backlogRemaining > 0 &&
    Number.isFinite(backlogPerCapita) &&
    backlogPerCapita < plannerPerCapitaThreshold;
  const idleConditionsMet = idleSlots >= plannerIdleSlotThreshold;

  if (plannerConditionsMet) {
    if (!backlogLowNotified) {
      backlogLowNotified = true;
      await sendTelegramMessage(
        `Backlog per-capita low: ${backlogRemaining} tasks for ${maxParallel} slots (${backlogPerCapita.toFixed(
          2,
        )} per slot). Triggering task planner.`,
      );
    }
    await maybeTriggerTaskPlanner("backlog-per-capita", {
      backlogRemaining,
      backlogPerCapita,
      running,
      review,
      error,
      idleSlots,
      maxParallel,
      threshold: plannerPerCapitaThreshold,
    });
    return;
  } else {
    // Conditions no longer met — reset so we re-notify next time
    backlogLowNotified = false;
  }

  if (idleConditionsMet) {
    if (!idleAgentsNotified) {
      idleAgentsNotified = true;
      await sendTelegramMessage(
        `Agents idle: ${idleSlots} slot(s) available (running ${running}/${maxParallel}). Triggering task planner.`,
      );
    }
    await maybeTriggerTaskPlanner("idle-slots", {
      backlogRemaining,
      backlogPerCapita,
      running,
      review,
      error,
      idleSlots,
      maxParallel,
      threshold: plannerIdleSlotThreshold,
    });
  } else {
    idleAgentsNotified = false;
  }
}

async function triggerTaskPlanner(
  reason,
  details,
  { taskCount, notify = true } = {},
) {
  if (plannerMode === "disabled") {
    return { status: "skipped", reason: "planner_disabled" };
  }
  if (plannerTriggered) {
    return { status: "skipped", reason: "planner_busy" };
  }
  plannerTriggered = true;
  await updatePlannerState({
    last_triggered_at: new Date().toISOString(),
    last_trigger_reason: reason || "manual",
    last_trigger_details: details || null,
    last_trigger_mode: plannerMode,
  });

  try {
    if (plannerMode === "kanban") {
      return await triggerTaskPlannerViaKanban(reason, { taskCount, notify });
    }
    return await triggerTaskPlannerViaCodex(reason, { notify });
  } catch (err) {
    const message = err && err.message ? err.message : String(err);
    if (notify) {
      await sendTelegramMessage(
        `Task planner run failed (${plannerMode}): ${message}`,
      );
    }
    throw err; // re-throw so callers (e.g. /plan command) know it failed
  } finally {
    plannerTriggered = false;
  }
}

/**
 * Trigger the task planner by creating a VK task — a real agent will
 * pick it up and plan the next phase of work.
 */
async function triggerTaskPlannerViaKanban(
  reason,
  { taskCount, notify = true } = {},
) {
  const numTasks =
    taskCount && Number.isFinite(taskCount) && taskCount > 0 ? taskCount : 5;
  // Get project ID using the name-matched helper
  const projectId = await findVkProjectId();
  if (!projectId) {
    throw new Error("Cannot reach VK API or no project found");
  }

  // Check for existing planner tasks to avoid duplicates
  // Only block on TODO tasks whose title matches the exact format we create
  const existingTasks = await fetchVk(
    `/api/tasks?project_id=${projectId}&status=todo`,
  );
  const existingPlanner = (existingTasks?.data || []).find((t) => {
    // Double-check status client-side — VK API filter may not work reliably
    if (t.status && t.status !== "todo") return false;
    const title = (t.title || "").toLowerCase();
    // Only match the exact title format we create: "Plan next tasks (...)"
    return (
      title.startsWith("plan next tasks") ||
      title.startsWith("plan next phase")
    );
  });
  if (existingPlanner) {
    console.log(
      `[monitor] task planner VK task already exists in backlog — skipping: "${existingPlanner.title}" (${existingPlanner.id})`,
    );
    const taskUrl = buildVkTaskUrl(existingPlanner.id, projectId);
    if (notify) {
      const suffix = taskUrl ? `\n${taskUrl}` : "";
      await sendTelegramMessage(
        `📋 Task planner skipped — existing planning task found (${projectId.substring(0, 8)}...).${suffix}`,
      );
    }
    return {
      status: "skipped",
      reason: "existing_planner_task",
      taskId: existingPlanner.id,
      taskTitle: existingPlanner.title,
      taskUrl,
      projectId,
    };
  }

  const plannerPrompt = agentPrompts.planner;
  const taskBody = {
    title: `Plan next tasks (${reason || "backlog-empty"})`,
    description: [
      "## Task Planner — Auto-created by codex-monitor",
      "",
      `**Trigger reason:** ${reason || "manual"}`,
      "",
      "### Instructions",
      "",
      plannerPrompt,
      "",
      "### Additional Context",
      "",
      "- Review recently merged PRs on GitHub to understand what was completed",
      "- Check `git log --oneline -20` for the latest changes",
      "- Look at open issues for inspiration",
      `- Create ${numTasks} well-scoped tasks in vibe-kanban`,
      "- Each task should be completable by a single agent in 1-4 hours",
    ].join("\n"),
    status: "todo",
    project_id: projectId,
  };

  const result = await fetchVk(`/api/tasks`, {
    method: "POST",
    body: taskBody,
    timeoutMs: 15000,
  });

  if (result?.success) {
    console.log(
      `[monitor] task planner VK task created: ${result.data?.id || "ok"}`,
    );
    await updatePlannerState({
      last_success_at: new Date().toISOString(),
      last_success_reason: reason || "manual",
    });
    const createdId = result.data?.id || null;
    const createdUrl = buildVkTaskUrl(createdId, projectId);
    if (notify) {
      const suffix = createdUrl ? `\n${createdUrl}` : "";
      await sendTelegramMessage(
        `📋 Task planner: created VK task for next phase planning (${reason}).${suffix}`,
      );
    }
    return {
      status: "created",
      taskId: createdId,
      taskTitle: taskBody.title,
      taskUrl: createdUrl,
      projectId,
    };
  }
  throw new Error("VK task creation failed");
}

/**
 * Trigger the task planner via Codex SDK — runs the planner prompt directly
 * in an in-process Codex thread.
 */
async function triggerTaskPlannerViaCodex(reason, { notify = true } = {}) {
  if (!codexEnabled) {
    throw new Error(
      "Codex SDK disabled — use TASK_PLANNER_MODE=kanban instead",
    );
  }
  notifyCodexTrigger("task planner run");
  if (!CodexClient) {
    CodexClient = await loadCodexSdk();
  }
  if (!CodexClient) {
    throw new Error("Codex SDK not available");
  }
  const agentPrompt = agentPrompts.planner;
  const codex = new CodexClient();
  const thread = codex.startThread();
  const prompt = `${agentPrompt}\n\nPlease execute the task planning instructions above.`;
  const result = await thread.run(prompt);
  const outPath = resolve(logDir, `task-planner-${nowStamp()}.md`);
  const output = formatCodexResult(result);
  await writeFile(outPath, output, "utf8");
  console.log(`[monitor] task planner output saved: ${outPath}`);
  await updatePlannerState({
    last_success_at: new Date().toISOString(),
    last_success_reason: reason || "manual",
  });
  if (notify) {
    await sendTelegramMessage(
      `📋 Task planner run completed (${reason || "manual"}). Output saved: ${outPath}`,
    );
  }
  return { status: "completed", outputPath: outPath };
}

async function ensureLogDir() {
  await mkdir(logDir, { recursive: true });
}

/**
 * Truncate the log directory to stay within logMaxSizeMb.
 * Deletes oldest files first until total size is under the limit.
 * Returns { deletedCount, freedBytes, totalBefore, totalAfter }.
 */
async function truncateOldLogs() {
  if (!logMaxSizeMb || logMaxSizeMb <= 0)
    return { deletedCount: 0, freedBytes: 0 };
  const { readdir, stat: fsStat } = await import("node:fs/promises");
  const maxBytes = logMaxSizeMb * 1024 * 1024;
  let entries;
  try {
    entries = await readdir(logDir);
  } catch {
    return { deletedCount: 0, freedBytes: 0 };
  }
  // Gather file info
  const files = [];
  for (const name of entries) {
    const filePath = resolve(logDir, name);
    try {
      const s = await fsStat(filePath);
      if (s.isFile()) {
        files.push({ name, path: filePath, size: s.size, mtimeMs: s.mtimeMs });
      }
    } catch {
      /* skip inaccessible files */
    }
  }
  const totalBefore = files.reduce((sum, f) => sum + f.size, 0);
  if (totalBefore <= maxBytes) {
    return {
      deletedCount: 0,
      freedBytes: 0,
      totalBefore,
      totalAfter: totalBefore,
    };
  }
  // Sort oldest first
  files.sort((a, b) => a.mtimeMs - b.mtimeMs);
  let currentSize = totalBefore;
  let deletedCount = 0;
  let freedBytes = 0;
  for (const f of files) {
    if (currentSize <= maxBytes) break;
    try {
      await unlink(f.path);
      currentSize -= f.size;
      freedBytes += f.size;
      deletedCount++;
    } catch {
      /* skip locked/active files */
    }
  }
  const totalAfter = currentSize;
  if (deletedCount > 0) {
    const mbFreed = (freedBytes / 1024 / 1024).toFixed(1);
    const mbAfter = (totalAfter / 1024 / 1024).toFixed(1);
    console.log(
      `[monitor] log rotation: deleted ${deletedCount} old log files, freed ${mbFreed} MB (${mbAfter} MB / ${logMaxSizeMb} MB limit)`,
    );
  }
  return { deletedCount, freedBytes, totalBefore, totalAfter };
}

async function finalizeActiveLog(activePath, archivePath) {
  try {
    await rename(activePath, archivePath);
  } catch {
    try {
      await copyFile(activePath, archivePath);
      await unlink(activePath);
    } catch {
      // Best effort only.
    }
  }
}

function nowStamp() {
  const d = new Date();
  const pad = (v) => String(v).padStart(2, "0");
  return `${d.getFullYear()}${pad(d.getMonth() + 1)}${pad(d.getDate())}-${pad(
    d.getHours(),
  )}${pad(d.getMinutes())}${pad(d.getSeconds())}`;
}

function safeStringify(value) {
  const seen = new Set();
  try {
    return JSON.stringify(
      value,
      (key, val) => {
        if (typeof val === "object" && val !== null) {
          if (seen.has(val)) {
            return "[Circular]";
          }
          seen.add(val);
        }
        if (typeof val === "bigint") {
          return val.toString();
        }
        return val;
      },
      2,
    );
  } catch {
    return null;
  }
}

function formatCodexResult(result) {
  if (result === null || result === undefined) {
    return "";
  }
  if (typeof result === "string") {
    return result;
  }
  if (typeof result === "number" || typeof result === "boolean") {
    return String(result);
  }
  if (typeof result === "object") {
    const candidates = [
      result.output,
      result.text,
      result.message,
      result.content,
    ];
    for (const candidate of candidates) {
      if (typeof candidate === "string" && candidate.trim()) {
        return candidate;
      }
    }
    const serialized = safeStringify(result);
    if (serialized) {
      return serialized;
    }
  }
  return String(result);
}

async function analyzeWithCodex(logPath, logText, reason) {
  if (!codexEnabled) {
    return;
  }
  notifyCodexTrigger(`orchestrator analysis (${reason})`);

  // ── Build a workspace-aware prompt ────────────────────────────────────
  // The old approach used CodexClient SDK (chat-only, no file access).
  // The new approach uses `codex exec` with --full-auto so the agent can
  // actually read files, inspect git status, and give a real diagnosis.
  const logTail = logText.slice(-12000);
  const prompt = `You are diagnosing why the VirtEngine orchestrator exited.
You have FULL READ ACCESS to the workspace. Use it.

## Context
- Exit reason: ${reason}
- Orchestrator script: ${scriptPath}
- Repository root: ${repoRoot}
- Active log file: ${logPath}
- Monitor script: scripts/codex-monitor/monitor.mjs
- VK endpoint: ${vkEndpointUrl || "(not set)"}
- Git branch: ${(() => {
    try {
      return execSync("git branch --show-current", {
        cwd: repoRoot,
        encoding: "utf8",
      }).trim();
    } catch {
      return "unknown";
    }
  })()}

## Log tail (last ~12k chars)
\`\`\`
${logTail}
\`\`\`

## Instructions
1. READ the orchestrator script (${scriptPath}) to understand the code flow
2. READ any relevant source files referenced in the log
3. Check git status/diff if relevant
4. Diagnose the ROOT CAUSE — not surface symptoms
5. Do NOT edit or create any files. Analysis only.
6. Common issues:
   - Path errors: worktree paths don't contain the orchestrator script
   - Mutex contention: multiple instances fighting over named mutex
   - VK API failures: wrong HTTP method, endpoint down, auth issues
   - Git rebase conflicts: agent branches conflict with main
   - Exit 64: pwsh can't find the -File target
   - SIGKILL: OOM or external termination
7. Return a SHORT, ACTIONABLE diagnosis with the concrete fix.`;

  try {
    // Use runCodexExec from autofix.mjs — gives Codex workspace access
    const result = await runCodexExec(prompt, repoRoot, 90_000);

    const analysisPath = logPath.replace(/\.log$/, "-analysis.txt");
    const analysisText = result.output || result.error || "(no output)";
    await writeFile(analysisPath, analysisText, "utf8");

    if (telegramToken && telegramChatId) {
      const summary = analysisText.slice(0, 500).replace(/\n{3,}/g, "\n\n");
      void sendTelegramMessage(
        `🔍 Codex Analysis Result (${reason}):\n${summary}${analysisText.length > 500 ? "\n...(truncated)" : ""}`,
      );
    }
  } catch (err) {
    // Fallback: try the SDK chat approach if exec is unavailable
    try {
      if (!CodexClient) {
        const ready = await ensureCodexSdkReady();
        if (!ready) throw new Error(codexDisabledReason || "Codex SDK N/A");
      }
      const codex = new CodexClient();
      const thread = codex.startThread();
      const result = await thread.run(prompt);
      const analysisPath = logPath.replace(/\.log$/, "-analysis.txt");
      const analysisText = formatCodexResult(result);
      await writeFile(analysisPath, analysisText, "utf8");
      if (telegramToken && telegramChatId) {
        const summary = analysisText.slice(0, 500).replace(/\n{3,}/g, "\n\n");
        void sendTelegramMessage(
          `🔍 Codex Analysis Result (${reason}):\n${summary}${analysisText.length > 500 ? "\n...(truncated)" : ""}`,
        );
      }
    } catch (fallbackErr) {
      const analysisPath = logPath.replace(/\.log$/, "-analysis.txt");
      const message = fallbackErr?.message || String(fallbackErr);
      await writeFile(
        analysisPath,
        `Codex analysis failed: ${message}\n`,
        "utf8",
      );
      if (telegramToken && telegramChatId) {
        void sendTelegramMessage(`🔍 Codex Analysis Failed: ${message}`);
      }
    }
  }
}

async function loadCodexSdk() {
  const result = await tryImportCodex();
  if (result) {
    return result;
  }

  const installResult = installDependencies();
  if (!installResult) {
    return null;
  }

  return await tryImportCodex();
}

async function tryImportCodex() {
  try {
    const mod = await import("@openai/codex-sdk");
    return mod.Codex;
  } catch (err) {
    return null;
  }
}

function installDependencies() {
  const cwd = __dirname;
  const pnpm = spawnSync("pnpm", ["--version"], { stdio: "ignore" });
  if (pnpm.status === 0) {
    const res = spawnSync("pnpm", ["install"], { cwd, stdio: "inherit" });
    return res.status === 0;
  }

  const corepack = spawnSync("corepack", ["--version"], { stdio: "ignore" });
  if (corepack.status === 0) {
    const res = spawnSync("corepack", ["pnpm", "install"], {
      cwd,
      stdio: "inherit",
    });
    return res.status === 0;
  }

  const npm = spawnSync("npm", ["install"], { cwd, stdio: "inherit" });
  return npm.status === 0;
}

async function ensureCodexSdkReady() {
  if (!codexEnabled) {
    return false;
  }
  const client = await loadCodexSdk();
  if (!client) {
    codexEnabled = false;
    codexDisabledReason =
      "Codex SDK not available (install failed or module missing)";
    console.warn(`[monitor] ${codexDisabledReason}`);
    return false;
  }
  CodexClient = client;
  return true;
}

function hasContextWindowError(text) {
  return contextPatterns.some((pattern) =>
    text.toLowerCase().includes(pattern.toLowerCase()),
  );
}

async function handleExit(code, signal, logPath) {
  if (shuttingDown) {
    return;
  }

  const logText = await readFile(logPath, "utf8").catch(() => "");
  const reason = signal ? `signal ${signal}` : `exit ${code}`;
  const isFileChangeRestart = pendingRestart && skipNextAnalyze;
  const isAbnormalExit = Boolean(signal) || code !== 0;
  const isCleanExit = !isAbnormalExit; // exit code 0, no signal

  if (pendingRestart) {
    pendingRestart = false;
    skipNextAnalyze = false;
    if (!skipNextRestartCount) {
      restartCount += 1;
    }
    skipNextRestartCount = false;

    // File-change restarts don't need analysis or auto-fix
    if (isFileChangeRestart) {
      startProcess();
      return;
    }
  }

  // ── Track quick exits for crash-loop detection ──────────────────────
  const runDurationMs = restartController.lastProcessStartAt
    ? Date.now() - restartController.lastProcessStartAt
    : Infinity;

  // ── Mutex-held: orchestrator found another instance holding the mutex ──
  const isMutexHeld =
    restartController.mutexHeldDetected ||
    logText.includes("Another orchestrator instance is already running") ||
    logText.includes("mutex held");
  const exitState = restartController.recordExit(runDurationMs, isMutexHeld);

  if (exitState.backoffReset) {
    console.log("[monitor] orchestrator ran >20s — resetting mutex backoff");
  }

  if (exitState.isMutexHeld) {
    console.warn(
      `[monitor] mutex held detected — backing off ${exitState.backoffMs / 1000}s ` +
        `(consecutive quick exits: ${exitState.consecutiveQuickExits})`,
    );
    if (telegramToken && telegramChatId) {
      void sendTelegramMessage(
        `⏳ Mutex held — backing off ${exitState.backoffMs / 1000}s before retry`,
      );
    }
    restartCount += 1;
    setTimeout(startProcess, exitState.backoffMs);
    return;
  }

  // ── Clean exit: skip autofix/analysis, handle backlog-empty gracefully ──
  if (isCleanExit) {
    const isEmptyBacklog =
      logText.includes("ALL TASKS COMPLETE") ||
      logText.includes("No more todo tasks in backlog") ||
      logText.includes("All tasks completed");

    if (isEmptyBacklog) {
      console.log(
        "[monitor] clean exit with empty backlog — triggering task planner",
      );
      // Trigger task planner to create more tasks
      await maybeTriggerTaskPlanner("backlog-empty-exit", {
        reason: "Orchestrator exited cleanly with empty backlog",
      });
      // Wait before restarting so the planner has time to create tasks
      const plannerWaitMs = 2 * 60 * 1000; // 2 minutes
      console.log(
        `[monitor] waiting ${plannerWaitMs / 1000}s for planner before restart`,
      );
      setTimeout(startProcess, plannerWaitMs);
      return;
    }

    // Other clean exits (e.g., Stop-Requested) — just restart normally
    console.log(
      `[monitor] clean exit (${reason}) — restarting without analysis`,
    );
    restartCount += 1;
    setTimeout(startProcess, restartDelayMs);
    return;
  }

  // ── Auto-fix: try to fix the crash automatically ──────────────────────
  if (autoFixEnabled && logText.length > 0) {
    const telegramFn =
      telegramToken && telegramChatId
        ? (msg) => void sendTelegramMessage(msg)
        : null;

    try {
      const result = await attemptAutoFix({
        logText: logText.slice(-15000),
        reason,
        repoRoot,
        logDir,
        onTelegram: telegramFn,
        recentMessages: getTelegramHistory(),
      });

      if (result.fixed) {
        // Fix was written to disk — the file watcher will restart us.
        // Don't call startProcess() manually — let the watcher handle it.
        console.log(
          "[monitor] auto-fix applied, waiting for file watcher to restart",
        );
        return;
      }

      // Not fixed — notify that autofix tried but couldn't help
      if (
        result.outcome &&
        result.outcome !== "clean-exit-skip" &&
        telegramFn
      ) {
        // Only notify if we haven't already (attemptAutoFix sends its own notifications)
        // but ensure the user knows the fallback path is happening
        console.log(
          `[monitor] auto-fix outcome: ${result.outcome.slice(0, 100)}`,
        );
      }
    } catch (err) {
      console.warn(`[monitor] auto-fix error: ${err.message || err}`);
      if (telegramToken && telegramChatId) {
        void sendTelegramMessage(
          `🔧 Auto-fix crashed: ${err.message || err}\nFalling back to Codex analysis.`,
        );
      }
    }
  }

  // ── Fallback: Codex SDK analysis (diagnosis only) ─────────────────────
  if (telegramToken && telegramChatId) {
    void sendTelegramMessage(
      `🔍 Codex analysis triggered (${reason}):\nAuto-fix was unable to resolve the crash — running diagnostic analysis.`,
    );
  }
  await analyzeWithCodex(logPath, logText.slice(-15000), reason);

  if (hasContextWindowError(logText)) {
    console.log(
      "[monitor] context window exhaustion detected — attempting fresh session",
    );
    const freshStarted = await attemptFreshSessionRetry(
      "context_window_exhausted",
      logText.slice(-3000),
    );
    if (freshStarted) {
      // Fresh session created — do NOT restart the current process.
      // The new session will handle the task in the same workspace.
      console.log("[monitor] fresh session handles task — skipping restart");
      return;
    }
    // Fallback: leave the hint file for manual recovery
    await writeFile(
      logPath.replace(/\.log$/, "-context.txt"),
      "Detected context window error. Fresh session retry failed — consider manual recovery.\n",
      "utf8",
    );
  }

  if (isAbnormalExit) {
    const restartCountNow = recordOrchestratorRestart();
    if (restartCountNow >= orchestratorRestartThreshold) {
      if (Date.now() >= orchestratorHaltedUntil) {
        orchestratorHaltedUntil = Date.now() + orchestratorPauseMs;
        const pauseMin = Math.round(orchestratorPauseMs / 60000);
        console.warn(
          `[monitor] crash loop detected (${restartCountNow} exits in 5m). Pausing orchestrator restarts for ${pauseMin}m.`,
        );
        if (!orchestratorResumeTimer) {
          orchestratorResumeTimer = setTimeout(() => {
            orchestratorResumeTimer = null;
            startProcess();
          }, orchestratorPauseMs);
        }
        if (telegramToken && telegramChatId) {
          void sendTelegramMessage(
            `🛑 Crash loop detected (${restartCountNow} exits in 5m). Pausing orchestrator restarts for ${pauseMin} minutes and requesting a fix.`,
          );
        }
        if (!orchestratorLoopFixInProgress) {
          orchestratorLoopFixInProgress = true;
          const fixResult = await attemptCrashLoopFix({
            reason,
            logText,
          }).catch((err) => ({
            fixed: false,
            outcome: err?.message || "crash-loop-fix-error",
          }));
          orchestratorLoopFixInProgress = false;
          if (fixResult.fixed) {
            if (telegramToken && telegramChatId) {
              void sendTelegramMessage(
                `🛠️ Crash-loop fix applied. Orchestrator will retry after cooldown.\n${fixResult.outcome}`,
              );
            }
          } else {
            // Crash loop fix failed — try a fresh session as last resort
            console.log(
              "[monitor] crash loop fix failed — attempting fresh session",
            );
            const freshStarted = await attemptFreshSessionRetry(
              "crash_loop_unresolvable",
              logText.slice(-3000),
            );
            if (freshStarted) {
              if (telegramToken && telegramChatId) {
                void sendTelegramMessage(
                  `🔄 Crash-loop fix failed but fresh session started. New agent will retry.`,
                );
              }
            } else if (telegramToken && telegramChatId) {
              void sendTelegramMessage(
                `⚠️ Crash-loop fix attempt failed: ${fixResult.outcome}. Fresh session also failed. Orchestrator remains paused.`,
              );
            }
          }
        }
      }
      return;
    }
  }

  if (maxRestarts > 0 && restartCount >= maxRestarts) {
    return;
  }

  const now = Date.now();
  if (now < orchestratorHaltedUntil || now < monitorSafeModeUntil) {
    const waitMs = Math.max(
      orchestratorHaltedUntil - now,
      monitorSafeModeUntil - now,
    );
    const waitSec = Math.max(5, Math.round(waitMs / 1000));
    console.warn(`[monitor] restart paused; retrying in ${waitSec}s`);
    setTimeout(startProcess, waitSec * 1000);
    return;
  }

  restartCount += 1;
  setTimeout(startProcess, restartDelayMs);
}

async function startProcess() {
  const now = Date.now();

  // ── Minimum restart interval — never restart faster than 15s ──────
  if (restartController.lastProcessStartAt > 0) {
    const sinceLast = now - restartController.lastProcessStartAt;
    const waitMs = restartController.getMinRestartDelay(now);
    if (waitMs > 0) {
      console.log(
        `[monitor] throttling restart — only ${Math.round(sinceLast / 1000)}s since last start, waiting ${Math.round(waitMs / 1000)}s`,
      );
      setTimeout(startProcess, waitMs);
      return;
    }
  }

  if (now < orchestratorHaltedUntil || now < monitorSafeModeUntil) {
    const waitMs = Math.max(
      orchestratorHaltedUntil - now,
      monitorSafeModeUntil - now,
    );
    const waitSec = Math.max(5, Math.round(waitMs / 1000));
    console.warn(
      `[monitor] orchestrator start blocked; retrying in ${waitSec}s`,
    );
    setTimeout(startProcess, waitSec * 1000);
    return;
  }
  if (!(await ensurePreflightReady("start"))) {
    return;
  }
  await ensureLogDir();
  const activeLogPath = resolve(logDir, "orchestrator-active.log");
  const archiveLogPath = resolve(logDir, `orchestrator-${nowStamp()}.log`);
  const logStream = await writeFile(activeLogPath, "", "utf8").then(() => null);

  // Guard: verify script exists before spawning to avoid cryptic exit 64
  if (!existsSync(scriptPath)) {
    console.error(
      `[monitor] orchestrator script not found: ${scriptPath}\n` +
        `  Set ORCHESTRATOR_SCRIPT to an absolute path or fix the relative path in .env`,
    );
    if (telegramToken && telegramChatId) {
      void sendTelegramMessage(
        `❌ Orchestrator script not found: ${scriptPath}\nSet ORCHESTRATOR_SCRIPT to a valid path.`,
      );
    }
    return;
  }

  // Reset mutex flag before spawn — will be re-set if this instance hits mutex
  restartController.noteProcessStarted(Date.now());

  const child = spawn("pwsh", ["-File", scriptPath, ...scriptArgs], {
    stdio: ["ignore", "pipe", "pipe"],
  });
  currentChild = child;

  const append = async (chunk) => {
    if (echoLogs) {
      try {
        process.stdout.write(chunk);
      } catch {
        /* EPIPE — ignore */
      }
    }
    const text = chunk.toString();
    try {
      await writeFile(activeLogPath, text, { flag: "a" });
    } catch {
      /* log file write failed — ignore */
    }
    logRemainder += text;
    const lines = logRemainder.split(/\r?\n/);
    logRemainder = lines.pop() || "";
    for (const line of lines) {
      if (isErrorLine(line, errorPatterns, errorNoisePatterns)) {
        notifyErrorLine(line);
      }
      if (line.includes("Merged PR") || line.includes("Marking task")) {
        notifyMerge(line);
      }
      if (line.includes("Merge notify: PR #")) {
        notifyMergeFailure(line);
      }
      // ── Mutex-held detection ─────────────────────────────────────
      restartController.noteLogLine(line);
      // ── Smart PR creation: detect completed/failed attempts ──────
      const prFlowMatch = line.match(
        /Attempt\s+([0-9a-f]{8})\s+finished\s+\((completed|failed)\)\s*[—–-]\s*marking review/i,
      );
      if (prFlowMatch) {
        const shortId = prFlowMatch[1];
        const finishStatus = prFlowMatch[2];
        void resolveAndTriggerSmartPR(shortId, finishStatus);
      }
      // ── "No remote branch" → trigger VK-based PR flow ──────────
      const noBranchMatch = line.match(
        /No remote branch for (ve\/([0-9a-f]{4})-\S+)/i,
      );
      if (noBranchMatch) {
        const shortId = noBranchMatch[2]; // 4-char prefix
        void resolveAndTriggerSmartPR(shortId, "no-remote-branch");
      }
      if (line.includes("ALL TASKS COMPLETE")) {
        if (!allCompleteNotified) {
          allCompleteNotified = true;
          void sendTelegramMessage(
            "All tasks completed. Orchestrator backlog is empty.",
          );
          void triggerTaskPlanner();
        }
      }
    }
  };

  child.stdout.on("data", (data) => append(data));
  child.stderr.on("data", (data) => append(data));
  // Prevent stream errors from bubbling up as uncaughtException
  child.stdout.on("error", () => {});
  child.stderr.on("error", () => {});

  child.on("exit", (code, signal) => {
    if (currentChild === child) {
      currentChild = null;
    }
    finalizeActiveLog(activeLogPath, archiveLogPath).finally(() => {
      handleExit(code, signal, archiveLogPath);
    });
  });
}

function requestRestart(reason) {
  if (shuttingDown) {
    return;
  }
  if (pendingRestart) {
    return;
  }
  // ── Suppress file-change restarts during mutex backoff ──────────
  if (restartController.shouldSuppressRestart(reason)) {
    console.log(
      `[monitor] suppressing file-change restart — mutex backoff active (${restartController.mutexBackoffMs / 1000}s)`,
    );
    return;
  }
  pendingRestart = true;
  skipNextAnalyze = true;
  skipNextRestartCount = true;

  console.log(`[monitor] restart requested (${reason})`);
  if (currentChild) {
    currentChild.kill("SIGTERM");
    setTimeout(() => {
      if (currentChild && !currentChild.killed) {
        currentChild.kill("SIGKILL");
      }
    }, 5000);
  } else {
    pendingRestart = false;
    startProcess();
  }
}

function stopWatcher() {
  if (watcher) {
    watcher.close();
    watcher = null;
  }
  watcherDebounce = null;
  watchFileName = null;
}

// ── Self-monitor watcher: restart when own .mjs files change ─────────────────
function stopSelfWatcher() {
  if (selfWatcher) {
    selfWatcher.close();
    selfWatcher = null;
  }
  if (selfWatcherDebounce) {
    clearTimeout(selfWatcherDebounce);
    selfWatcherDebounce = null;
  }
}

function selfRestartForSourceChange(filename) {
  // Defer restart if the primary agent is mid-turn — don't interrupt user tasks
  if (isPrimaryBusy()) {
    if (!pendingSelfRestart) {
      pendingSelfRestart = filename;
      console.log(
        `\n[monitor] source file changed: ${filename} — deferring restart (primary agent busy)`,
      );
    }
    return;
  }
  pendingSelfRestart = null;
  console.log(`\n[monitor] source file changed: ${filename}`);
  console.log("[monitor] exiting for self-restart (fresh ESM modules)...");
  shuttingDown = true;
  stopAutoUpdateLoop();
  stopSelfWatcher();
  stopWatcher();
  stopEnvWatchers();
  if (currentChild) {
    currentChild.kill("SIGTERM");
    setTimeout(() => {
      if (currentChild && !currentChild.killed) {
        currentChild.kill("SIGKILL");
      }
    }, 3000);
  }
  void releaseTelegramPollLock();
  stopTelegramBot({ preserveDigest: true });
  // Write self-restart marker so the new process suppresses startup notifications
  try {
    writeFileSync(
      resolve(repoRoot, ".cache", "ve-self-restart.marker"),
      String(Date.now()),
    );
  } catch {
    /* best effort */
  }
  // Exit with special code — cli.mjs re-forks with fresh module cache
  setTimeout(() => process.exit(SELF_RESTART_EXIT_CODE), 500);
}

function retryDeferredSelfRestart() {
  if (!pendingSelfRestart) return;
  if (isPrimaryBusy()) {
    // Still busy — check again in 5s
    setTimeout(retryDeferredSelfRestart, 5000);
    return;
  }
  const filename = pendingSelfRestart;
  console.log(
    `[monitor] primary agent finished — proceeding with deferred self-restart (${filename})`,
  );
  selfRestartForSourceChange(filename);
}

function startSelfWatcher() {
  stopSelfWatcher();
  try {
    selfWatcher = watch(__dirname, { persistent: true }, (_event, filename) => {
      // Only react to .mjs source files
      if (!filename || !filename.endsWith(".mjs")) return;
      // Ignore node_modules and log artifacts
      if (filename.includes("node_modules")) return;
      if (selfWatcherDebounce) {
        clearTimeout(selfWatcherDebounce);
      }
      selfWatcherDebounce = setTimeout(() => {
        selfRestartForSourceChange(filename);
        // If deferred, start polling for agent completion
        if (pendingSelfRestart) {
          setTimeout(retryDeferredSelfRestart, 5000);
        }
      }, 3000);
    });
    console.log("[monitor] watching own source files for self-restart");
  } catch (err) {
    console.warn(`[monitor] self-watcher failed: ${err.message}`);
  }
}

async function startWatcher(force = false) {
  if (!watchEnabled) {
    stopWatcher();
    return;
  }
  if (watcher && !force) {
    return;
  }
  if (watcher && force) {
    stopWatcher();
  }
  let targetPath = watchPath;
  try {
    const stats = await (await import("node:fs/promises")).stat(watchPath);
    if (stats.isFile()) {
      watchFileName = watchPath.split(/[\\/]/).pop();
      targetPath = watchPath.split(/[\\/]/).slice(0, -1).join("/") || ".";
    }
  } catch {
    // Default to watching the provided path.
  }

  watcher = watch(targetPath, { persistent: true }, (_event, filename) => {
    if (watchFileName && filename && filename !== watchFileName) {
      return;
    }
    if (watcherDebounce) {
      clearTimeout(watcherDebounce);
    }
    watcherDebounce = setTimeout(() => {
      requestRestart("file-change");
    }, 5000);
  });
}

function stopEnvWatchers() {
  for (const w of envWatchers) {
    try {
      w.close();
    } catch {
      /* best effort */
    }
  }
  envWatchers = [];
  envWatcherDebounce = null;
}

function scheduleEnvReload(reason) {
  if (envWatcherDebounce) {
    clearTimeout(envWatcherDebounce);
  }
  envWatcherDebounce = setTimeout(() => {
    void reloadConfig(reason || "env-change");
  }, 400);
}

function startEnvWatchers() {
  stopEnvWatchers();
  if (!envPaths || envPaths.length === 0) {
    return;
  }
  const dirMap = new Map();
  for (const envPath of envPaths) {
    const dir = resolve(envPath, "..");
    const file = envPath.split(/[\\/]/).pop();
    if (!file) continue;
    if (!dirMap.has(dir)) {
      dirMap.set(dir, new Set());
    }
    dirMap.get(dir).add(file);
  }
  for (const [dir, files] of dirMap.entries()) {
    try {
      const w = watch(dir, { persistent: true }, (_event, filename) => {
        if (!filename) return;
        if (!files.has(filename)) return;
        scheduleEnvReload(`env:${filename}`);
      });
      envWatchers.push(w);
    } catch {
      /* best effort */
    }
  }
}

function applyConfig(nextConfig, options = {}) {
  const { restartIfChanged = false, reason = "config-change" } = options;
  const prevScriptPath = scriptPath;
  const prevArgs = scriptArgs?.join(" ") || "";
  const prevWatchPath = watchPath;
  const prevTelegramInterval = telegramIntervalMin;
  const prevCodexEnabled = codexEnabled;
  const prevTelegramCommandEnabled = telegramCommandEnabled;
  const prevTelegramBotEnabled = telegramBotEnabled;
  const prevPreflightEnabled = preflightEnabled;

  config = nextConfig;
  projectName = nextConfig.projectName;
  scriptPath = nextConfig.scriptPath;
  scriptArgs = nextConfig.scriptArgs;
  restartDelayMs = nextConfig.restartDelayMs;
  maxRestarts = nextConfig.maxRestarts;
  logDir = nextConfig.logDir;
  watchEnabled = nextConfig.watchEnabled;
  watchPath = resolve(nextConfig.watchPath);
  echoLogs = nextConfig.echoLogs;
  autoFixEnabled = nextConfig.autoFixEnabled;
  preflightEnabled = nextConfig.preflightEnabled;
  preflightRetryMs = nextConfig.preflightRetryMs;
  repoRoot = nextConfig.repoRoot;
  statusPath = nextConfig.statusPath;
  telegramPollLockPath = nextConfig.telegramPollLockPath;
  telegramToken = nextConfig.telegramToken;
  telegramChatId = nextConfig.telegramChatId;
  telegramIntervalMin = nextConfig.telegramIntervalMin;
  telegramCommandPollTimeoutSec = nextConfig.telegramCommandPollTimeoutSec;
  telegramCommandConcurrency = nextConfig.telegramCommandConcurrency;
  telegramCommandMaxBatch = nextConfig.telegramCommandMaxBatch;
  telegramBotEnabled = nextConfig.telegramBotEnabled;
  telegramCommandEnabled = nextConfig.telegramCommandEnabled;
  repoSlug = nextConfig.repoSlug;
  repoUrlBase = nextConfig.repoUrlBase;
  vkRecoveryPort = nextConfig.vkRecoveryPort;
  vkRecoveryHost = nextConfig.vkRecoveryHost;
  vkEndpointUrl = nextConfig.vkEndpointUrl;
  vkPublicUrl = nextConfig.vkPublicUrl;
  vkTaskUrlTemplate = nextConfig.vkTaskUrlTemplate;
  // Invalidate VK caches when endpoint URL changes
  cachedRepoId = null;
  cachedProjectId = null;
  vkRecoveryCooldownMin = nextConfig.vkRecoveryCooldownMin;
  vkSpawnEnabled = nextConfig.vkSpawnEnabled;
  vkEnsureIntervalMs = nextConfig.vkEnsureIntervalMs;
  plannerPerCapitaThreshold = nextConfig.plannerPerCapitaThreshold;
  plannerIdleSlotThreshold = nextConfig.plannerIdleSlotThreshold;
  plannerDedupMs = nextConfig.plannerDedupMs;
  plannerMode = nextConfig.plannerMode || "kanban";
  agentPrompts = nextConfig.agentPrompts;
  executorScheduler = nextConfig.scheduler;
  agentSdk = nextConfig.agentSdk;
  envPaths = nextConfig.envPaths;
  codexEnabled = nextConfig.codexEnabled;
  codexDisabledReason = codexEnabled
    ? ""
    : process.env.CODEX_SDK_DISABLED === "1"
      ? "disabled via CODEX_SDK_DISABLED"
      : agentSdk?.primary && agentSdk.primary !== "codex"
        ? `disabled via agent_sdk.primary=${agentSdk.primary}`
        : "disabled via --no-codex";

  if (prevWatchPath !== watchPath || watchEnabled === false) {
    void startWatcher(true);
  }
  startEnvWatchers();

  if (prevTelegramInterval !== telegramIntervalMin) {
    void startTelegramNotifier();
  }
  if (!prevTelegramCommandEnabled && telegramCommandEnabled) {
    startTelegramCommandListener();
  }
  if (prevTelegramBotEnabled !== telegramBotEnabled) {
    if (telegramBotEnabled) {
      void startTelegramBot();
    } else {
      stopTelegramBot();
    }
  }
  if (prevCodexEnabled && !codexEnabled) {
    console.warn(
      `[monitor] Codex disabled: ${codexDisabledReason || "disabled"}`,
    );
  }
  if (!prevCodexEnabled && codexEnabled) {
    void ensureCodexSdkReady();
  }
  if (prevPreflightEnabled && !preflightEnabled && preflightRetryTimer) {
    clearTimeout(preflightRetryTimer);
    preflightRetryTimer = null;
  }

  const nextArgs = scriptArgs?.join(" ") || "";
  const scriptChanged = prevScriptPath !== scriptPath || prevArgs !== nextArgs;
  if (restartIfChanged && scriptChanged) {
    requestRestart(`config-change (${reason})`);
  }
}

async function reloadConfig(reason) {
  try {
    const nextConfig = loadConfig(process.argv, { reloadEnv: true });
    applyConfig(nextConfig, { restartIfChanged: true, reason });
    console.log(`[monitor] config reloaded (${reason})`);
    if (telegramToken && telegramChatId) {
      try {
        await sendTelegramMessage(
          `🔄 .env reloaded (${reason}). Runtime config updated.`,
          { dedupKey: "env-reload" },
        );
      } catch {
        /* best effort */
      }
    }
  } catch (err) {
    const message = err && err.message ? err.message : String(err);
    console.warn(`[monitor] failed to reload config: ${message}`);
  }
}

process.on("SIGINT", () => {
  shuttingDown = true;
  stopAutoUpdateLoop();
  stopSelfWatcher();
  stopEnvWatchers();
  if (watcher) {
    watcher.close();
  }
  if (currentChild) {
    currentChild.kill("SIGTERM");
  }
  void releaseTelegramPollLock();
  process.exit(0);
});

// Windows: closing the terminal window doesn't send SIGINT/SIGTERM reliably.
process.on("exit", () => {
  shuttingDown = true;
  void releaseTelegramPollLock();
});

process.on("SIGTERM", () => {
  shuttingDown = true;
  stopAutoUpdateLoop();
  stopSelfWatcher();
  stopEnvWatchers();
  if (watcher) {
    watcher.close();
  }
  if (currentChild) {
    currentChild.kill("SIGTERM");
  }
  void releaseTelegramPollLock();
  stopTelegramBot();
  process.exit(0);
});

// Stream noise patterns that should NEVER trigger recovery —
// they happen when child processes die or pipes break and are harmless.
function isStreamNoise(msg) {
  return (
    msg.includes("EPIPE") ||
    msg.includes("ERR_STREAM_PREMATURE_CLOSE") ||
    msg.includes("ERR_STREAM_DESTROYED") ||
    msg.includes("write after end") ||
    msg.includes("This socket has been ended") ||
    msg.includes("Cannot read properties of null") ||
    msg.includes("ECONNRESET") ||
    msg.includes("ECONNREFUSED") ||
    msg.includes("socket hang up")
  );
}

process.on("uncaughtException", (err) => {
  const msg = err?.message || "";
  // Always suppress stream noise — not just during shutdown
  if (isStreamNoise(msg)) {
    try {
      process.stderr.write(`[monitor] suppressed stream noise: ${msg}\n`);
    } catch {
      /* even stderr might be broken */
    }
    return;
  }
  if (shuttingDown) return;
  void handleMonitorFailure("uncaughtException", err);
});

process.on("unhandledRejection", (reason) => {
  const msg = reason?.message || String(reason || "");
  // Always suppress stream noise
  if (isStreamNoise(msg)) {
    try {
      process.stderr.write(`[monitor] suppressed stream noise: ${msg}\n`);
    } catch {
      /* even stderr might be broken */
    }
    return;
  }
  if (shuttingDown) return;
  const err =
    reason instanceof Error ? reason : new Error(String(reason || ""));
  void handleMonitorFailure("unhandledRejection", err);
});

// ── Singleton guard: prevent ghost monitors ─────────────────────────────────
if (!process.env.VITEST && !acquireMonitorLock(config.cacheDir)) {
  process.exit(1);
}

// ── Codex CLI config.toml: ensure VK MCP + stream timeouts ──────────────────
try {
  const vkPort = config.vkRecoveryPort || "54089";
  const vkBaseUrl = config.vkEndpointUrl || `http://127.0.0.1:${vkPort}`;
  const tomlResult = ensureCodexConfig({
    vkBaseUrl,
    vkPort,
    vkHost: "127.0.0.1",
  });
  if (!tomlResult.noChanges) {
    console.log("[monitor] updated ~/.codex/config.toml:");
    printConfigSummary(tomlResult);
  }
} catch (err) {
  console.warn(
    `[monitor] config.toml check failed (non-fatal): ${err.message}`,
  );
}

// ── Startup sweep: kill stale processes, prune worktrees ────────────────────
runMaintenanceSweep({ repoRoot });

setInterval(() => {
  void flushErrorQueue();
}, 60 * 1000);

// ── Periodic maintenance: every 5 min, reap stuck pushes & prune worktrees ──
const maintenanceIntervalMs = 5 * 60 * 1000;
setInterval(() => {
  const childPid = currentChild ? currentChild.pid : undefined;
  runMaintenanceSweep({ repoRoot, childPid });
}, maintenanceIntervalMs);

// ── Periodic merged PR check: every 10 min, move merged PRs to done ─────────
const mergedPRCheckIntervalMs = 10 * 60 * 1000;
setInterval(() => {
  void checkMergedPRsAndUpdateTasks();
}, mergedPRCheckIntervalMs);

// ── Log rotation: truncate oldest logs when folder exceeds size limit ───────
if (logMaxSizeMb > 0) {
  // Run once at startup (delayed 10s)
  setTimeout(() => void truncateOldLogs(), 10 * 1000);
  if (logCleanupIntervalMin > 0) {
    const logCleanupIntervalMs = logCleanupIntervalMin * 60 * 1000;
    setInterval(() => void truncateOldLogs(), logCleanupIntervalMs);
    console.log(
      `[monitor] log rotation enabled — max ${logMaxSizeMb} MB, checking every ${logCleanupIntervalMin} min`,
    );
  } else {
    console.log(
      `[monitor] log rotation enabled — max ${logMaxSizeMb} MB (startup check only)`,
    );
  }
}

// Run once immediately after startup (delayed by 30s to let things settle)
setTimeout(() => {
  void checkMergedPRsAndUpdateTasks();
  void checkAndMergeDependabotPRs();
}, 30 * 1000);

// ── Periodic Dependabot auto-merge check ─────────────────────────────────────
if (dependabotAutoMerge) {
  const depIntervalMs = (dependabotAutoMergeIntervalMin || 10) * 60 * 1000;
  setInterval(() => {
    void checkAndMergeDependabotPRs();
  }, depIntervalMs);
  console.log(
    `[dependabot] auto-merge enabled — checking every ${dependabotAutoMergeIntervalMin || 10} min for: ${dependabotAuthors.join(", ")}`,
  );
}

// ── Self-updating: poll npm every 10 min, auto-install + restart ────────────
startAutoUpdateLoop({
  onRestart: (reason) => restartSelf(reason),
  onNotify: (msg) => {
    try {
      void sendTelegramMessage(msg);
    } catch {
      /* best-effort */
    }
  },
});

startWatcher();
startEnvWatchers();
startSelfWatcher();
if (vkSpawnEnabled) {
  void ensureVibeKanbanRunning();
}
if (
  vkSpawnEnabled &&
  Number.isFinite(vkEnsureIntervalMs) &&
  vkEnsureIntervalMs > 0
) {
  setInterval(() => {
    void ensureVibeKanbanRunning();
  }, vkEnsureIntervalMs);
}
void ensureCodexSdkReady().then(() => {
  if (!codexEnabled) {
    const reason = codexDisabledReason || "disabled";
    console.warn(`[monitor] Codex disabled: ${reason}`);
  } else {
    console.log("[monitor] Codex enabled.");
  }
});
startProcess();
if (telegramCommandEnabled) {
  startTelegramCommandListener();
}
// Restore live digest state BEFORE any messages flow — so restarts continue the
// existing digest message instead of creating a new one.
// Chain notifier start after restore to prevent race conditions.
void restoreLiveDigest()
  .catch(() => {})
  .then(() => startTelegramNotifier());

// ── Two-way Telegram ↔ primary agent ────────────────────────────────────────
injectMonitorFunctions({
  sendTelegramMessage,
  readStatusData,
  readStatusSummary,
  getCurrentChild: () => currentChild,
  startProcess,
  getVibeKanbanUrl: () => vkPublicUrl || vkEndpointUrl,
  fetchVk,
  getRepoRoot: () => repoRoot,
  startFreshSession,
  attemptFreshSessionRetry,
  buildRetryPrompt,
  getActiveAttemptInfo,
  triggerTaskPlanner,
  reconcileTaskStatuses,
});
if (telegramBotEnabled) {
  void startTelegramBot();
}

// ── Named exports for testing ───────────────────────────────────────────────
export { fetchVk, updateTaskStatus, getTaskAgeMs };
