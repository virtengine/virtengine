import { execSync, spawn, spawnSync } from "node:child_process";
import {
  existsSync,
  mkdirSync,
  readFileSync,
  watch,
  writeFileSync,
  appendFileSync,
} from "node:fs";
import {
  copyFile,
  mkdir,
  readFile,
  rename,
  unlink,
  writeFile,
} from "node:fs/promises";
import { clearLine, createInterface, cursorTo } from "node:readline";
import net from "node:net";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { acquireMonitorLock, runMaintenanceSweep } from "./maintenance.mjs";
import { archiveCompletedTasks } from "./task-archiver.mjs";
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
  getDigestSnapshot,
  startStatusFileWriter,
  stopStatusFileWriter,
} from "./telegram-bot.mjs";
import { PRCleanupDaemon } from "./pr-cleanup-daemon.mjs";
import {
  execPrimaryPrompt,
  initPrimaryAgent,
  setPrimaryAgent,
  getPrimaryAgentName,
  switchPrimaryAgent,
} from "./primary-agent.mjs";
import { execPooledPrompt } from "./agent-pool.mjs";
import { loadConfig } from "./config.mjs";
import { formatPreflightReport, runPreflightChecks } from "./preflight.mjs";
import { startAutoUpdateLoop, stopAutoUpdateLoop } from "./update-check.mjs";
import { ensureCodexConfig, printConfigSummary } from "./codex-config.mjs";
import { RestartController } from "./restart-controller.mjs";
import {
  analyzeMergeStrategy,
  executeDecision,
  resetMergeStrategyDedup,
} from "./merge-strategy.mjs";
import { assessTask, quickAssess } from "./task-assessment.mjs";
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
import {
  initFleet,
  refreshFleet,
  buildFleetPresence,
  getFleetState,
  isFleetCoordinator,
  getFleetMode,
  getTotalFleetSlots,
  buildExecutionWaves,
  assignTasksToWorkstations,
  calculateBacklogDepth,
  detectMaintenanceMode,
  formatFleetSummary,
  persistFleetState,
} from "./fleet-coordinator.mjs";
import {
  getComplexityMatrix,
  assessCompletionConfidence,
  classifyComplexity,
  COMPLEXITY_TIERS,
  DEFAULT_MODEL_PROFILES,
} from "./task-complexity.mjs";
import {
  getDirtyTasks,
  prioritizeDirtyTasks,
  shouldReserveDirtySlot,
  getDirtySlotReservation,
  buildConflictResolutionPrompt,
  isFileOverlapWithDirtyPR,
  registerDirtyTask,
  clearDirtyTask,
  isDirtyTask,
  getHighTierForDirty,
  isOnResolutionCooldown,
  recordResolutionAttempt,
  formatDirtyTaskSummary,
  DIRTY_TASK_DEFAULTS,
} from "./conflict-resolver.mjs";
import {
  resolveConflictsWithSDK,
  isSDKResolutionOnCooldown,
  isSDKResolutionExhausted,
  clearSDKResolutionState,
} from "./sdk-conflict-resolver.mjs";
import {
  initSharedKnowledge,
  buildKnowledgeEntry,
  appendKnowledgeEntry,
  formatKnowledgeSummary,
} from "./shared-knowledge.mjs";
import { WorkspaceMonitor } from "./workspace-monitor.mjs";
import { VkLogStream } from "./vk-log-stream.mjs";
import { VKErrorResolver } from "./vk-error-resolver.mjs";
import { createAnomalyDetector } from "./anomaly-detector.mjs";
import {
  getWorktreeManager,
  acquireWorktree,
  releaseWorktree,
  releaseWorktreeByBranch,
  findWorktreeForBranch as findManagedWorktree,
  pruneStaleWorktrees,
  getWorktreeStats,
} from "./worktree-manager.mjs";
import {
  getTaskExecutor,
  isInternalExecutorEnabled,
  getExecutorMode,
  loadExecutorOptionsFromConfig,
} from "./task-executor.mjs";
import { configureFromArgs, installConsoleInterceptor } from "./lib/logger.mjs";
// ── Task management subsystem imports ──────────────────────────────────────
import {
  loadStore as loadTaskStore,
  getStats as getTaskStoreStats,
  setTaskStatus as setInternalTaskStatus,
  setReviewResult,
  getTasksPendingReview,
  getStaleInProgressTasks,
  getStaleInReviewTasks,
  getAllTasks as getAllInternalTasks,
} from "./task-store.mjs";
import { createAgentEndpoint } from "./agent-endpoint.mjs";
import { createReviewAgent } from "./review-agent.mjs";
import { createSyncEngine } from "./sync-engine.mjs";
import { createErrorDetector } from "./error-detector.mjs";
const __dirname = resolve(fileURLToPath(new URL(".", import.meta.url)));

// ── Anomaly signal file path (shared with ve-orchestrator.ps1) ──────────────
const ANOMALY_SIGNAL_PATH = resolve(
  __dirname,
  "..",
  ".cache",
  "anomaly-signals.json",
);

/**
 * Write an anomaly signal to the shared signal file for the orchestrator to pick up.
 * The orchestrator reads this file in Process-AnomalySignals and acts accordingly.
 */
function writeAnomalySignal(anomaly) {
  try {
    const dir = resolve(__dirname, "..", ".cache");
    mkdirSync(dir, { recursive: true });
    let signals = [];
    try {
      const raw = readFileSync(ANOMALY_SIGNAL_PATH, "utf8");
      signals = JSON.parse(raw);
      if (!Array.isArray(signals)) signals = [];
    } catch {
      /* file doesn't exist yet */
    }
    signals.push({
      type: anomaly.type,
      severity: anomaly.severity,
      action: anomaly.action,
      shortId: anomaly.shortId,
      processId: anomaly.processId,
      message: anomaly.message,
      timestamp: new Date().toISOString(),
    });
    // Cap at 50 signals to prevent unbounded growth
    if (signals.length > 50) signals = signals.slice(-50);
    writeFileSync(ANOMALY_SIGNAL_PATH, JSON.stringify(signals, null, 2));
  } catch (err) {
    console.warn(
      `[anomaly-detector] writeAnomalySignal failed: ${err.message}`,
    );
  }
}

// ── Configure logging before anything else ──────────────────────────────────
configureFromArgs(process.argv.slice(2));

// ── Load unified configuration ──────────────────────────────────────────────
let config = loadConfig();

// Install console interceptor with log file (after config provides logDir)
{
  const _logDir = config.logDir || resolve(__dirname, "logs");
  const _logFile = resolve(_logDir, "monitor.log");
  installConsoleInterceptor({ logFile: _logFile });
}

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
  interactiveShellEnabled,
  autoFixEnabled,
  preflightEnabled: configPreflightEnabled,
  preflightRetryMs: configPreflightRetryMs,
  primaryAgent,
  primaryAgentEnabled,
  agentPoolEnabled,
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
  branchRouting,
  telegramVerbosity,
  fleet: fleetConfig,
  internalExecutor: internalExecutorConfig,
  executorMode: configExecutorMode,
} = config;

let watchPath = resolve(configWatchPath);
let codexEnabled = config.codexEnabled;
let plannerMode = configPlannerMode; // "codex-sdk" | "kanban" | "disabled"
console.log(`[monitor] task planner mode: ${plannerMode}`);
let primaryAgentName = primaryAgent;
let primaryAgentReady = primaryAgentEnabled;

// ── Workspace monitor: track agent workspaces with git state + stuck detection ──
const workspaceMonitor = new WorkspaceMonitor({
  cacheDir: resolve(repoRoot, ".cache", "workspace-logs"),
  repoRoot,
  onStuckDetected: ({ attemptId, reason, recommendation }) => {
    const msg = `⚠️ Agent ${attemptId.substring(0, 8)} stuck: ${reason}\nRecommendation: ${recommendation}`;
    console.warn(`[workspace-monitor] ${msg}`);
    void notify?.(msg, { dedupKey: `stuck-${attemptId.substring(0, 8)}` });
  },
});

// ── Devmode Auto Code Fix: background agent that fixes issues from digest ──
const devmodeAutoCodeFix = {
  enabled: ["1", "true", "yes"].includes(
    String(process.env.DEVMODE_AUTO_CODE_FIX || "").toLowerCase(),
  ),
  cycleCount: 0,
  cycleInterval: Math.max(
    1,
    Number(process.env.DEVMODE_AUTO_CODE_FIX_CYCLE_INTERVAL || "2"),
  ),
  timeoutMs: Number(process.env.DEVMODE_AUTO_CODE_FIX_TIMEOUT_MS || "300000"), // 5 min
  running: false,
  lastRunAt: 0,
  lastDigestText: "",
  branch: process.env.DEVMODE_AUTO_CODE_FIX_BRANCH || "",
};
if (devmodeAutoCodeFix.enabled) {
  console.log(
    `[monitor] devmode auto code fix ENABLED (every ${devmodeAutoCodeFix.cycleInterval} digest cycle(s), timeout ${devmodeAutoCodeFix.timeoutMs}ms)`,
  );
}

// ── Interactive shell state ────────────────────────────────────────────────
const shellState = {
  enabled: !!interactiveShellEnabled,
  active: false,
  rl: null,
  prompt: "",
  agentStreaming: false,
  agentStreamed: false,
  agentPrefixPrinted: false,
  abortController: null,
  queue: Promise.resolve(),
};
const shellIsTTY = process.stdin.isTTY && process.stdout.isTTY;
const shellAnsi = {
  cyan: (s) => (process.stdout.isTTY ? `\x1b[36m${s}\x1b[0m` : s),
  green: (s) => (process.stdout.isTTY ? `\x1b[32m${s}\x1b[0m` : s),
  yellow: (s) => (process.stdout.isTTY ? `\x1b[33m${s}\x1b[0m` : s),
  dim: (s) => (process.stdout.isTTY ? `\x1b[2m${s}\x1b[22m` : s),
  red: (s) => (process.stdout.isTTY ? `\x1b[31m${s}\x1b[0m` : s),
};
const shellPromptText = shellAnsi.cyan("[agent]") + " > ";
const shellInfoPrefix = shellAnsi.dim("[shell]") + " ";
console.log(`[monitor] task planner mode: ${plannerMode}`);

function shellWriteRaw(chunk) {
  try {
    process.stdout.write(chunk);
  } catch {
    /* ignore write failures */
  }
}

function shellWriteLine(text) {
  shellWriteRaw(`${shellInfoPrefix}${text}\n`);
}

function startInteractiveShell() {
  if (!shellIsTTY || shellState.active) {
    return;
  }
  const rl = createInterface({
    input: process.stdin,
    output: process.stdout,
    prompt: shellPromptText,
    terminal: true,
  });
  shellState.rl = rl;
  shellState.active = true;
  rl.on("line", (line) => {
    const trimmed = line.trim();
    if (!trimmed) {
      rl.prompt();
      return;
    }
    if (["exit", "quit"].includes(trimmed.toLowerCase())) {
      rl.close();
      return;
    }
    shellState.queue = shellState.queue
      .then(async () => {
        if (!primaryAgentReady) {
          shellWriteLine("Primary agent not ready.");
          return;
        }
        await execPrimaryPrompt(trimmed, { timeoutMs: 15 * 60 * 1000 });
      })
      .catch((err) => {
        shellWriteLine(`Error: ${err.message || err}`);
      })
      .finally(() => {
        rl.prompt();
      });
  });
  rl.on("close", () => {
    shellState.active = false;
    shellState.rl = null;
  });
  rl.prompt();
}
let codexDisabledReason = codexEnabled
  ? ""
  : process.env.CODEX_SDK_DISABLED === "1"
    ? "disabled via CODEX_SDK_DISABLED"
    : agentSdk?.primary && agentSdk.primary !== "codex"
      ? `disabled via agent_sdk.primary=${agentSdk.primary}`
      : "disabled via --no-codex";
setPrimaryAgent(primaryAgentName);
let preflightEnabled = configPreflightEnabled;
let preflightRetryMs = configPreflightRetryMs;
if (primaryAgentReady) {
  void initPrimaryAgent(primaryAgentName);
}

// Merge strategy: Codex-powered merge decision analysis
// Enabled by default unless CODEX_ANALYZE_MERGE_STRATEGY=false
const codexAnalyzeMergeStrategy =
  agentPoolEnabled &&
  (process.env.CODEX_ANALYZE_MERGE_STRATEGY || "").toLowerCase() !== "false";
const mergeStrategyMode = String(
  process.env.MERGE_STRATEGY_MODE || "smart",
).toLowerCase();
const codexResolveConflictsEnabled =
  agentPoolEnabled &&
  (process.env.CODEX_RESOLVE_CONFLICTS || "true").toLowerCase() !== "false";
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
const SELF_RESTART_QUIET_MS = Math.max(
  90_000,
  Number(process.env.SELF_RESTART_QUIET_MS || "90000"),
);
let selfWatcher = null;
let selfWatcherDebounce = null;
let selfRestartTimer = null;
let selfRestartLastChangeAt = 0;
let selfRestartLastFile = null;
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

// ── VK WebSocket log stream — captures real-time agent logs from execution processes ──
let vkLogStream = null;

// ── VK Error Resolver — auto-resolves errors from VK logs ──
let vkErrorResolver = null;
let vkSessionDiscoveryTimer = null;
let vkSessionDiscoveryInFlight = false;
const vkSessionCache = new Map();

const VK_SESSION_KEEP_STATUSES = new Set([
  "running",
  "review",
  "manual_review",
  "in_review",
  "inreview",
]);

function normalizeAttemptStatus(status) {
  return String(status || "")
    .toLowerCase()
    .trim()
    .replace(/[\s-]+/g, "_");
}

function shouldKeepSessionForStatus(status) {
  return VK_SESSION_KEEP_STATUSES.has(normalizeAttemptStatus(status));
}

// ── Anomaly detector — plaintext pattern matching for death loops, stalls, etc. ──
let anomalyDetector = null;
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
let lastErrorLine = "";
let lastErrorAt = 0;
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
  const result = await runCodexExec(prompt, repoRoot, 1_800_000);
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

// Infrastructure error patterns that should NEVER trigger loop-fix autofix.
// These are transient git/rebase failures handled by persistent cooldowns.
const infraErrorPatterns = [
  /Direct rebase failed/i,
  /checkout failed/i,
  /rebase cooldown/i,
  /worktree.*has (rebase in progress|uncommitted changes)/i,
  /No worktree found/i,
  /VK rebase (failed|unavailable)/i,
  /git fetch failed in worktree/i,
  /Cannot rebase/i,
  /merge conflict.*not auto-resolvable/i,
];

function isInfraError(line) {
  return infraErrorPatterns.some((p) => p.test(line));
}

function trackErrorFrequency(line) {
  // Skip infrastructure errors — they have their own cooldown/retry logic
  if (isInfraError(line)) return;

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

function triggerLoopFix(errorLine, repeatCount) {
  if (!autoFixEnabled) return;
  loopFixInProgress = true;

  const telegramFn =
    telegramToken && telegramChatId
      ? (msg) => void sendTelegramMessage(msg)
      : null;

  // Fire-and-forget: never block the stdout pipeline
  void (async () => {
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
      } else {
        console.log(
          `[monitor] loop fix returned no changes: ${result.outcome || "no-fix"}`,
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
  })();
}

const contextPatterns = [
  "ContextWindowExceeded",
  "context window",
  "ran out of room",
  "prompt token count",
  "token count of",
  "context length exceeded",
  "maximum context length",
  "exceeds the limit",
  "token limit",
  "too many tokens",
  "prompt too large",
  "failed to get response from the ai model",
  "capierror",
];

const errorPatterns = [
  /\bERROR\b/i,
  /Exception/i,
  /Traceback/i,
  /SetValueInvocationException/i,
  /Cannot bind argument/i,
  /Unhandled/i,
  /\bFailed to compile\b/i,
  /\bFailed to start\b/i,
  /\bFATAL\b/i,
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
  // Attempt lifecycle lines that include "failed" but are expected status updates
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+Attempt [0-9a-f]{8} finished \(failed\)\s+—\s+marking review/i,
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+Attempt [0-9a-f]{8} failed in workspace — requires agent attention/i,
  // Agent work logger noise (handled separately, not a monitor crash)
  /^\s*\[agent-logger\]\s+Session ended:/i,
  /^\s*\[agent-logger\]\s+Error logged:/i,
  // Attempt lifecycle lines that include "failed" but are normal status updates
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+Attempt [0-9a-f]{8} finished \(failed\)\s+—\s+marking review/i,
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+Attempt [0-9a-f]{8} failed in workspace — requires agent attention/i,
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
  // ── Infrastructure failures: rebase/checkout/worktree issues ──
  // These are transient git infra failures, NOT code bugs.
  // The orchestrator handles them with cooldowns; do NOT trigger autofix.
  /Direct rebase failed:.*checkout failed/i,
  /Direct rebase failed:.*merge conflict/i,
  /Direct rebase failed:.*push failed/i,
  /Direct rebase failed:.*setting cooldown/i,
  /Direct merge-rebase (succeeded|failed)/i,
  /Branch .* is on rebase cooldown/i,
  /Worktree .* has (rebase in progress|uncommitted changes)/i,
  /No worktree found for .* — using VK API/i,
  /Cannot rebase: (working tree is dirty|git rebase already in progress)/i,
  /VK rebase (failed|requested|unavailable)/i,
  /git fetch failed in worktree/i,
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

  // Use shell: true only when running through npx (string command).
  // When using the local binary directly, avoid shell to prevent DEP0190
  // deprecation warning ("Passing args to child process with shell true").
  vibeKanbanProcess = spawn(spawnCmd, spawnArgs, {
    env,
    cwd: repoRoot,
    stdio: "ignore",
    shell: process.platform === "win32" || !useLocal,
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
    // Start VK log stream if not already running
    ensureVkLogStream();
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
  // Stop log stream — will restart when VK comes back online
  if (vkLogStream) {
    vkLogStream.stop();
    vkLogStream = null;
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

/**
 * Ensure the VK log stream is running. Creates a new VkLogStream instance
 * if one doesn't exist, connecting to VK's execution-process WebSocket
 * endpoints to capture real-time agent stdout/stderr.
 *
 * Two log outputs:
 *   1. Raw per-process logs → .cache/agent-logs/vk-exec-{shortId}.log
 *   2. Structured session logs → logs/vk-sessions/vk-session-{stamp}-{shortId}.log
 *      (mirrors codex-exec format with task metadata headers for autofix analysis)
 *
 * Discovery model: No REST list endpoint exists for execution processes.
 * Instead, connectToSession(sessionId) is called when sessions are created
 * (see startFreshSession). On startup, we also scan active_attempts for any
 * existing session IDs to connect to.
 */
function ensureVkLogStream() {
  if (vkLogStream) return;
  console.log("[monitor] ensureVkLogStream: creating VkLogStream instance");

  // Initialize anomaly detector if not already running
  if (!anomalyDetector) {
    anomalyDetector = createAnomalyDetector({
      onAnomaly: (anomaly) => {
        const icon =
          anomaly.severity === "CRITICAL"
            ? "🔴"
            : anomaly.severity === "HIGH"
              ? "🟠"
              : "🟡";
        console.warn(
          `[anomaly-detector] ${icon} ${anomaly.severity} ${anomaly.type} [${anomaly.shortId}]: ${anomaly.message}`,
        );

        // Act on kill/restart actions — write signal file for the orchestrator
        // AND directly kill the VK process WebSocket to stop further resource
        // wastage immediately. The signal file ensures the orchestrator also
        // archives and retries the attempt on its next loop.
        if (anomaly.action === "kill" || anomaly.action === "restart") {
          console.warn(
            `[anomaly-detector] writing signal for action="${anomaly.action}" ${anomaly.type} on process ${anomaly.shortId}`,
          );
          writeAnomalySignal(anomaly);

          // Directly kill the VK log stream for this process so the agent
          // stops consuming compute immediately. Don't wait for the
          // orchestrator's next poll cycle.
          if (vkLogStream && anomaly.processId) {
            const killed = vkLogStream.killProcess(
              anomaly.processId,
              `anomaly: ${anomaly.type} (${anomaly.action})`,
            );
            if (killed) {
              console.warn(
                `[anomaly-detector] killed VK process stream ${anomaly.shortId} directly`,
              );
            }
          }
        }
      },
      notify: (text, options) => {
        sendTelegramMessage(text, options).catch(() => {});
      },
    });
    console.log("[monitor] anomaly detector started");
  }

  const agentLogDir = resolve(repoRoot, ".cache", "agent-logs");
  const sessionLogDir = resolve(__dirname, "logs", "vk-sessions");
  vkLogStream = new VkLogStream(vkEndpointUrl, {
    logDir: agentLogDir,
    sessionLogDir,
    // Always keep VK log streaming silent in the CLI.
    echo: false,
    filterLine: (line) => {
      // Drop verbose VK/Codex event chatter and token streams.
      if (!line) return false;
      if (line.length > 6000) return false;
      if (line.startsWith('{"method":"codex/event/')) return false;
      if (line.startsWith('{"method":"item/')) return false;
      if (line.startsWith('{"method":"thread/')) return false;
      if (line.startsWith('{"method":"account/')) return false;
      if (line.includes('"type":"reasoning_content_delta"')) return false;
      if (line.includes('"type":"agent_reasoning_delta"')) return false;
      if (line.includes('"type":"token_count"')) return false;
      if (line.includes('"type":"item_started"')) return false;
      if (line.includes('"type":"item_completed"')) return false;
      if (line.includes('"type":"exec_command_begin"')) return false;
      if (line.includes('"type":"exec_command_output_delta"')) return false;
      if (line.includes('"type":"exec_command_end"')) return false;
      if (line.includes('"method":"codex/event/reasoning_content_delta"'))
        return false;
      if (line.includes('"method":"codex/event/agent_reasoning_delta"'))
        return false;
      if (line.includes('"method":"codex/event/token_count"')) return false;
      if (line.includes('"method":"codex/event/item_started"')) return false;
      if (line.includes('"method":"codex/event/item_completed"')) return false;
      if (line.includes('"method":"codex/event/exec_command_')) return false;
      if (line.includes('"method":"item/reasoning/summaryTextDelta"'))
        return false;
      if (line.includes('"method":"item/commandExecution/outputDelta"'))
        return false;
      if (line.includes('"method":"codex/event/agent_reasoning"')) return false;
      return true;
    },
    onLine: (line, meta) => {
      // Feed every agent log line to the anomaly detector for real-time
      // pattern matching (death loops, token overflow, stalls, etc.).
      if (anomalyDetector) {
        try {
          anomalyDetector.processLine(line, meta);
        } catch {
          /* detector error — non-fatal */
        }
      }

      // Feed log lines to VK error resolver for auto-resolution
      if (vkErrorResolver) {
        try {
          void vkErrorResolver.handleLogLine(line);
        } catch (err) {
          console.error(`[monitor] vkErrorResolver error: ${err.message}`);
        }
      }
    },
    onProcessConnected: (processId, meta) => {
      // When a new execution process is discovered via the session stream,
      // look up task metadata from status data and enrich the process
      void (async () => {
        try {
          const statusData = await readStatusData();
          const attempts = statusData?.attempts || {};
          // Find the attempt that matches this session
          // VK processes belong to sessions which belong to workspaces (= attempts)
          for (const [attemptId, info] of Object.entries(attempts)) {
            if (!info) continue;
            // Match by session_id if available, or if the process was connected
            // for a session belonging to this attempt
            if (
              meta.sessionId &&
              (info.session_id === meta.sessionId ||
                attemptId === meta.sessionId)
            ) {
              vkLogStream.setProcessMeta(processId, {
                attemptId,
                taskId: info.task_id,
                taskTitle: info.task_title || info.name,
                branch: info.branch,
                sessionId: meta.sessionId,
                executor: info.executor,
                executorVariant: info.executor_variant,
              });
              break;
            }
          }
        } catch {
          /* best effort */
        }
      })();
    },
  });
  vkLogStream.start();

  // Initialize VK error resolver
  const vkAutoResolveEnabled = config.vkAutoResolveErrors ?? true;
  if (vkAutoResolveEnabled) {
    console.log("[monitor] initializing VK error resolver...");
    vkErrorResolver = new VKErrorResolver(repoRoot, vkEndpointUrl, {
      enabled: true,
      onResolve: (resolution) => {
        console.log(
          `[monitor] VK auto-resolution: ${resolution.errorType} - ${resolution.result.success ? "✓ success" : "✗ failed"}`,
        );

        // Notify via Telegram
        const emoji = resolution.result.success ? "🤖" : "⚠️";
        const status = resolution.result.success ? "resolved" : "failed";
        const branch =
          resolution.context.branch || `PR #${resolution.context.prNumber}`;
        notify(`${emoji} Auto-${status} ${resolution.errorType} on ${branch}`);
      },
    });
    console.log("[monitor] VK error resolver initialized");
  }

  // Discover any active sessions immediately and keep polling for new sessions
  void refreshVkSessionStreams("startup");
  ensureVkSessionDiscoveryLoop();
}

function ensureVkSessionDiscoveryLoop() {
  if (vkSessionDiscoveryTimer) return;
  if (!Number.isFinite(vkEnsureIntervalMs) || vkEnsureIntervalMs <= 0) return;
  vkSessionDiscoveryTimer = setInterval(() => {
    void refreshVkSessionStreams("periodic");
  }, vkEnsureIntervalMs);
}

async function refreshVkSessionStreams(reason = "manual") {
  if (!vkLogStream) {
    console.log(`[monitor] refreshVkSessionStreams(${reason}): no vkLogStream`);
    return;
  }
  if (vkSessionDiscoveryInFlight) return;
  vkSessionDiscoveryInFlight = true;

  try {
    // ── 1. Collect attempts from orchestrator status file ──────────────
    const statusData = await readStatusData();
    const statusAttempts = statusData?.attempts || {};
    const statusDataAvailable = !!statusData;

    // ── 2. Also query VK directly for all non-archived attempts ────────
    //    The status file can be stale/incomplete (e.g. after restarts or
    //    when attempts were submitted in previous orchestrator cycles).
    let vkAttempts = [];
    let vkAttemptsAvailable = false;
    try {
      const vkRes = await fetchVk("/api/task-attempts?archived=false");
      if (vkRes?.success && Array.isArray(vkRes.data)) {
        vkAttempts = vkRes.data;
        vkAttemptsAvailable = true;
      }
    } catch (err) {
      console.warn(
        `[monitor] refreshVkSessionStreams: VK attempt fetch failed: ${err.message}`,
      );
    }

    const allowedAttemptIds = new Set();
    const allowedSessions = new Set();

    // ── 3. Merge: build unified map of attemptId → metadata ───────────
    /** @type {Map<string, {task_id?:string, task_title?:string, branch?:string, session_id?:string, executor?:string, executor_variant?:string}>} */
    const mergedAttempts = new Map();

    // Status file attempts (mark running + review states)
    for (const [attemptId, info] of Object.entries(statusAttempts)) {
      if (!attemptId || !info) continue;
      if (!shouldKeepSessionForStatus(info.status)) continue;
      allowedAttemptIds.add(attemptId);
      if (info.session_id) {
        allowedSessions.add(info.session_id);
      }
      mergedAttempts.set(attemptId, {
        task_id: info.task_id,
        task_title: info.task_title || info.name,
        branch: info.branch,
        session_id: info.session_id,
        executor: info.executor,
        executor_variant: info.executor_variant,
        source: "status",
      });
    }

    // VK API attempts (add any not already present from status file)
    for (const vkAttempt of vkAttempts) {
      if (!vkAttempt?.id) continue;
      if (mergedAttempts.has(vkAttempt.id)) continue; // status file takes precedence
      const vkStatus = vkAttempt.status ?? vkAttempt.state ?? "";
      if (vkStatus && !shouldKeepSessionForStatus(vkStatus)) {
        continue;
      }
      allowedAttemptIds.add(vkAttempt.id);
      mergedAttempts.set(vkAttempt.id, {
        task_id: vkAttempt.task_id,
        task_title: vkAttempt.name,
        branch: vkAttempt.branch,
        session_id: null,
        executor: null,
        executor_variant: null,
        source: "vk-api",
      });
    }

    console.log(
      `[monitor] refreshVkSessionStreams(${reason}): ${mergedAttempts.size} attempts ` +
        `(${Object.values(statusAttempts).filter((i) => i?.status === "running").length} status + ` +
        `${vkAttempts.length} vk-api, merged)`,
    );

    // Keep cached sessions for allowed attempts
    for (const attemptId of allowedAttemptIds) {
      const cachedSession = vkSessionCache.get(attemptId);
      if (cachedSession) {
        allowedSessions.add(cachedSession);
      }
    }

    // ── 4. Discover sessions and connect ──────────────────────────────
    for (const [attemptId, info] of mergedAttempts) {
      let sessionId = info.session_id || vkSessionCache.get(attemptId) || null;

      if (!sessionId) {
        sessionId = await fetchLatestVkSessionId(attemptId);
        if (sessionId) {
          vkSessionCache.set(attemptId, sessionId);
          console.log(
            `[monitor] refreshVkSessionStreams: discovered session ${sessionId.slice(0, 8)} for attempt ${attemptId.slice(0, 8)} (${info.source})`,
          );
        }
      }

      if (!sessionId) continue; // no session yet — will retry next cycle

      allowedSessions.add(sessionId);
      vkLogStream.setProcessMeta(attemptId, {
        attemptId,
        taskId: info.task_id,
        taskTitle: info.task_title,
        branch: info.branch,
        sessionId,
        executor: info.executor,
        executorVariant: info.executor_variant,
      });
      vkLogStream.connectToSession(sessionId);
    }

    if (statusDataAvailable || vkAttemptsAvailable) {
      for (const attemptId of Array.from(vkSessionCache.keys())) {
        if (!allowedAttemptIds.has(attemptId)) {
          vkSessionCache.delete(attemptId);
        }
      }
      if (vkLogStream?.pruneSessions) {
        const pruned = vkLogStream.pruneSessions(
          allowedSessions,
          "session no longer active",
        );
        if (pruned > 0) {
          console.log(
            `[monitor] refreshVkSessionStreams(${reason}): pruned ${pruned} stale session streams`,
          );
        }
      }
    }
  } catch (err) {
    console.warn(
      `[monitor] VK session discovery (${reason}) failed: ${err.message || err}`,
    );
  } finally {
    vkSessionDiscoveryInFlight = false;
  }
}

async function fetchLatestVkSessionId(workspaceId) {
  const res = await fetchVk(
    `/api/sessions?workspace_id=${encodeURIComponent(workspaceId)}`,
  );
  if (!res?.success || !Array.isArray(res.data)) return null;
  const sessions = res.data;
  if (!sessions.length) return null;
  const ordered = sessions.slice().sort((a, b) => {
    const aTs = Date.parse(a?.updated_at || a?.created_at || 0) || 0;
    const bTs = Date.parse(b?.updated_at || b?.created_at || 0) || 0;
    return bTs - aTs;
  });
  return ordered[0]?.id || null;
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

/**
 * Find the worktree path for a given branch.
 * Delegates to the centralized WorktreeManager; falls back to direct git parsing
 * for branches not tracked in the registry.
 */
function findWorktreeForBranch(branch) {
  if (!branch) return null;
  // Try centralized manager first (has registry + git porcelain search)
  try {
    const managed = findManagedWorktree(branch);
    if (managed) return managed;
  } catch {
    // Manager may not be initialized — fall through
  }
  // Fallback: direct git worktree list parsing
  try {
    const result = spawnSync("git", ["worktree", "list", "--porcelain"], {
      cwd: repoRoot,
      stdio: ["ignore", "pipe", "pipe"],
      timeout: 10000,
      encoding: "utf8",
      shell: process.platform === "win32",
    });
    if (result.status !== 0 || !result.stdout) return null;

    const lines = result.stdout.split("\n");
    let currentPath = null;
    for (const line of lines) {
      if (line.startsWith("worktree ")) {
        currentPath = line.slice(9).trim();
      } else if (line.startsWith("branch ") && currentPath) {
        const branchRef = line.slice(7).trim();
        const branchName = branchRef.replace(/^refs\/heads\//, "");
        if (branchName === branch) {
          return currentPath;
        }
      } else if (line.trim() === "") {
        currentPath = null;
      }
    }
    return null;
  } catch {
    return null;
  }
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
  // Guard: internal executor mode runs tasks via agent-pool, not VK sessions
  const execMode = configExecutorMode || getExecutorMode();
  if (execMode === "internal") {
    console.log(
      `[monitor] startFreshSession skipped — executor mode is "internal"`,
    );
    return {
      success: false,
      reason: "internal executor mode — VK sessions disabled",
    };
  }

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

    // Connect the VK log stream to this session for real-time log capture
    if (vkLogStream) {
      // Set metadata so structured session logs get proper headers
      const attemptInfo = await getAttemptInfo(workspaceId);
      if (attemptInfo) {
        vkLogStream.setProcessMeta(workspaceId, {
          attemptId: workspaceId,
          taskId: attemptInfo.task_id,
          taskTitle: attemptInfo.task_title || attemptInfo.name,
          branch: attemptInfo.branch,
          sessionId: session.id,
          executor: attemptInfo.executor,
          executorVariant: attemptInfo.executor_variant,
        });
      }
      vkLogStream.connectToSession(session.id);
    }

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
  // Guard: internal executor mode runs tasks via agent-pool, not VK sessions
  const execMode = configExecutorMode || getExecutorMode();
  if (execMode === "internal") {
    console.log(
      `[monitor] attemptFreshSessionRetry skipped — executor mode is "internal"`,
    );
    return false;
  }

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
 * Return the task's "version" timestamp used for cache invalidation.
 * Prefers updated_at, falls back to created_at.
 * @param {object} task
 * @returns {string} ISO-ish timestamp string or empty string
 */
function getTaskUpdatedAt(task) {
  return task?.updated_at || task?.created_at || "";
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
  const ok = res?.success === true;
  // Clear recovery caches — task status changed, so it needs re-evaluation
  if (ok) clearRecoveryCaches(taskId);
  return ok;
}

function parseTaskTimestamp(value) {
  if (!value) return null;
  const raw =
    value.created_at ||
    value.createdAt ||
    value.created ||
    value.updated_at ||
    value.updatedAt ||
    value.updated ||
    value.started_at ||
    value.startedAt ||
    value;
  if (!raw) return null;
  const ts = Date.parse(raw);
  return Number.isFinite(ts) ? ts : null;
}

function isPlannerTaskData(task) {
  if (!task) return false;
  const title = String(task.title || "").toLowerCase();
  const desc = String(task.description || task.body || "").toLowerCase();
  if (title.includes("plan next tasks") || title.includes("plan next phase")) {
    return true;
  }
  if (title.includes("task planner")) {
    return true;
  }
  return (
    desc.includes("task planner — auto-created by codex-monitor") ||
    desc.includes("task planner - auto-created by codex-monitor")
  );
}

async function verifyPlannerTaskCompletion(taskData, attemptInfo) {
  const projectId =
    taskData?.project_id ||
    taskData?.projectId ||
    attemptInfo?.project_id ||
    attemptInfo?.projectId ||
    (await findVkProjectId());
  if (!projectId) {
    return { completed: false, reason: "project_not_found" };
  }
  const tasksRes = await fetchVk(`/api/tasks?project_id=${projectId}`);
  const tasks = Array.isArray(tasksRes?.data)
    ? tasksRes.data
    : Array.isArray(tasksRes?.tasks)
      ? tasksRes.tasks
      : Array.isArray(tasksRes)
        ? tasksRes
        : [];
  const sinceMs =
    parseTaskTimestamp(taskData) ||
    parseTaskTimestamp(attemptInfo) ||
    Date.now();
  const candidates = tasks.filter((t) => {
    if (!t || t.id === taskData?.id) return false;
    if (isPlannerTaskData(t)) return false;
    const createdMs = parseTaskTimestamp(t);
    return createdMs && createdMs > sinceMs;
  });
  const backlogCandidates = candidates.filter((t) => {
    if (!t?.status) return true;
    const status = String(t.status).toLowerCase();
    return (
      status === "todo" || status === "inprogress" || status === "inreview"
    );
  });
  const finalCandidates =
    backlogCandidates.length > 0 ? backlogCandidates : candidates;
  return {
    completed: finalCandidates.length > 0,
    createdCount: finalCandidates.length,
    projectId,
    sinceMs,
    sampleTitles: finalCandidates
      .slice(0, 3)
      .map((t) => t.title || t.id)
      .filter(Boolean),
  };
}

/**
 * Safe recovery: re-fetches a task's live status from VK before moving it
 * to "todo".  If the user has since cancelled/done the task, the recovery
 * is aborted.  This prevents the loop where:
 *   user cancels → monitor moves to todo → orchestrator re-dispatches.
 *
 * @param {string} taskId - Task UUID
 * @param {string} taskTitle - Human-readable title (for logging)
 * @param {string} reason - Why the recovery is happening (for logging)
 * @returns {Promise<boolean>} true if moved to todo, false if skipped/failed
 */
async function safeRecoverTask(taskId, taskTitle, reason) {
  // In internal executor mode, only update task status — never start VK sessions
  const execMode = configExecutorMode || getExecutorMode();
  const isInternal = execMode === "internal";

  try {
    const res = await fetchVk(`/api/tasks/${taskId}`);
    const liveStatus = res?.data?.status || res?.status;
    const liveUpdatedAt = res?.data?.updated_at || res?.data?.created_at || "";
    if (!liveStatus) {
      console.warn(
        `[monitor] safeRecover: could not re-fetch status for "${taskTitle}" (${taskId.substring(0, 8)}...) — skipping`,
      );
      return false;
    }
    // If the user has moved the task out of inprogress (cancelled, done,
    // or even already todo), do NOT touch it.
    if (liveStatus === "cancelled" || liveStatus === "done") {
      console.log(
        `[monitor] safeRecover: task "${taskTitle}" is now ${liveStatus} — aborting recovery`,
      );
      // Cache so we skip this task for RECOVERY_SKIP_CACHE_MS
      recoverySkipCache.set(taskId, {
        resolvedStatus: liveStatus,
        timestamp: Date.now(),
        updatedAt: liveUpdatedAt,
        status: liveStatus,
      });
      scheduleRecoveryCacheSave();
      return false;
    }
    if (liveStatus === "todo") {
      console.log(
        `[monitor] safeRecover: task "${taskTitle}" is already todo — no action needed`,
      );
      // Cache so we skip this task for RECOVERY_SKIP_CACHE_MS
      recoverySkipCache.set(taskId, {
        resolvedStatus: liveStatus,
        timestamp: Date.now(),
        updatedAt: liveUpdatedAt,
        status: liveStatus,
      });
      scheduleRecoveryCacheSave();
      return false;
    }
    const success = await updateTaskStatus(taskId, "todo");
    if (success) {
      if (isInternal) {
        console.log(
          `[monitor] ♻️ Recovered "${taskTitle}" from ${liveStatus} → todo (${reason}) [internal mode — VK session skipped]`,
        );
      } else {
        console.log(
          `[monitor] ♻️ Recovered "${taskTitle}" from ${liveStatus} → todo (${reason})`,
        );
      }
    }
    return success;
  } catch (err) {
    console.warn(
      `[monitor] safeRecover failed for "${taskTitle}": ${err.message || err}`,
    );
    return false;
  }
}

/**
 * Checks if a git branch has been merged into the target base branch.
 * Uses GitHub CLI + git commands to determine merge status.
 *
 * IMPORTANT: "branch not on remote" does NOT mean merged. The agent may
 * never have pushed, the PR may have been closed without merging, or the
 * branch was manually deleted. We must verify via GitHub PR state.
 *
 * @param {string} branch - Branch name (e.g., "ve/1234-feat-auth")
 * @param {string} [baseBranch] - Upstream/base branch to compare against
 * @returns {Promise<boolean>} true if definitively merged, false otherwise
 */
async function isBranchMerged(branch, baseBranch) {
  if (!branch) return false;

  try {
    const target = normalizeBranchName(baseBranch) || DEFAULT_TARGET_BRANCH;

    const splitRemoteRef = (ref, defaultRemote = "origin") => {
      const match = String(ref || "").match(/^([^/]+)\/(.+)$/);
      if (match) return { remote: match[1], name: match[2] };
      return { remote: defaultRemote, name: ref };
    };

    const branchInfo = splitRemoteRef(normalizeBranchName(branch), "origin");
    const baseInfo = splitRemoteRef(target, "origin");
    const branchRef = `${branchInfo.remote}/${branchInfo.name}`;
    const baseRef = `${baseInfo.remote}/${baseInfo.name}`;
    const ghHead = branchInfo.name || branch;

    // ── Strategy 1: Check GitHub for a merged PR with this head branch ──
    // This is the most reliable signal — if GitHub says merged, it's merged.
    if (ghAvailable()) {
      try {
        const ghResult = execSync(
          `gh pr list --head "${ghHead}" --state merged --json number,mergedAt --limit 1`,
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
    const branchExistsCmd = `git ls-remote --heads ${branchInfo.remote} ${branchInfo.name}`;
    const branchExists = execSync(branchExistsCmd, {
      cwd: repoRoot,
      encoding: "utf8",
      stdio: ["pipe", "pipe", "ignore"],
    }).trim();

    // Branch NOT on remote — this does NOT prove it was merged.
    // Without a confirmed merged PR (strategy 1), we must assume NOT merged.
    if (!branchExists) {
      console.log(
        `[monitor] Branch ${branchRef} not found on ${branchInfo.remote} — no merged PR found against ${baseRef}, treating as NOT merged`,
      );
      return false;
    }

    // ── Strategy 3: Branch exists on remote — check if ancestor of main ─
    execSync(`git fetch ${baseInfo.remote} ${baseInfo.name} --quiet`, {
      cwd: repoRoot,
      stdio: "ignore",
      timeout: 15000,
    });
    execSync(`git fetch ${branchInfo.remote} ${branchInfo.name} --quiet`, {
      cwd: repoRoot,
      stdio: "ignore",
      timeout: 15000,
    });

    // Check if the branch is fully merged into origin/main
    // Returns non-zero exit code if not merged
    const mergeCheckCmd = `git merge-base --is-ancestor ${branchRef} ${baseRef}`;
    execSync(mergeCheckCmd, {
      cwd: repoRoot,
      stdio: "ignore",
      timeout: 10000,
    });

    // If we get here, the branch is merged
    console.log(
      `[monitor] Branch ${branchRef} is ancestor of ${baseRef} (merged)`,
    );
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

// ── Recovery/Idle caches (persistent) ───────────────────────────────────────

const recoveryCacheEnabled =
  String(process.env.RECOVERY_CACHE_ENABLED || "true").toLowerCase() !==
  "false";
const recoveryLogDedupMs =
  Number(process.env.RECOVERY_LOG_DEDUP_MINUTES || "30") * 60 * 1000;
const recoveryCacheMaxEntries = Number(
  process.env.RECOVERY_CACHE_MAX || "2000",
);

const recoveryCachePath = resolve(
  config.cacheDir || resolve(config.repoRoot, ".cache"),
  "ve-task-recovery-cache.json",
);

/**
 * Cooldown cache for tasks whose branches are all unresolvable (deleted,
 * no PR, abandoned).  We re-check them every 30 min instead of every cycle.
 * After STALE_MAX_STRIKES consecutive stale checks the task is moved back
 * to "todo" so another agent can pick it up.
 * Key = task ID, Value = { lastCheck: timestamp, strikes: number, updatedAt?: string, status?: string }.
 * @type {Map<string, {lastCheck: number, strikes: number, updatedAt?: string, status?: string}>}
 */
const staleBranchCooldown = new Map();

/**
 * Cache for tasks whose recovery was a no-op (already todo/cancelled/done).
 * Prevents redundant VK API calls, branch/PR checks, and log spam every cycle.
 * Key = task ID, Value = { resolvedStatus: string, timestamp: number, updatedAt?: string, status?: string }.
 * Expires after RECOVERY_SKIP_CACHE_MS so we re-check periodically.
 * @type {Map<string, {resolvedStatus: string, timestamp: number, updatedAt?: string, status?: string}>}
 */
const recoverySkipCache = new Map();

/**
 * Log dedup for repeated "no attempt found" messages.
 * Key = task ID, Value = { lastLogAt: number, updatedAt?: string, status?: string, reason?: string }.
 * @type {Map<string, {lastLogAt: number, updatedAt?: string, status?: string, reason?: string}>}
 */
const noAttemptLogCache = new Map();

let recoveryCacheDirty = false;
let recoveryCacheSaveTimer = null;

function taskVersionMatches(task, entry, status) {
  if (!entry) return false;
  const updatedAt = getTaskUpdatedAt(task);
  if (!updatedAt) return false;
  if (!entry.updatedAt) return false;
  if (entry.updatedAt !== updatedAt) return false;
  if (entry.status && status && entry.status !== status) return false;
  return true;
}

function scheduleRecoveryCacheSave() {
  if (!recoveryCacheEnabled) return;
  recoveryCacheDirty = true;
  if (recoveryCacheSaveTimer) return;
  recoveryCacheSaveTimer = setTimeout(() => {
    recoveryCacheSaveTimer = null;
    if (!recoveryCacheDirty) return;
    recoveryCacheDirty = false;
    saveRecoveryCache();
  }, 1000);
  if (typeof recoveryCacheSaveTimer.unref === "function") {
    recoveryCacheSaveTimer.unref();
  }
}

function buildCacheObject(map, tsField) {
  const entries = [...map.entries()];
  entries.sort((a, b) => (b[1]?.[tsField] || 0) - (a[1]?.[tsField] || 0));
  const limited =
    recoveryCacheMaxEntries > 0
      ? entries.slice(0, recoveryCacheMaxEntries)
      : entries;
  const obj = {};
  for (const [id, value] of limited) {
    obj[id] = value;
  }
  return obj;
}

function saveRecoveryCache() {
  if (!recoveryCacheEnabled) return;
  try {
    const payload = {
      version: 1,
      savedAt: new Date().toISOString(),
      staleCooldown: buildCacheObject(staleBranchCooldown, "lastCheck"),
      recoverySkip: buildCacheObject(recoverySkipCache, "timestamp"),
      noAttemptLog: buildCacheObject(noAttemptLogCache, "lastLogAt"),
    };
    writeFileSync(recoveryCachePath, JSON.stringify(payload, null, 2), "utf8");
  } catch {
    /* best-effort */
  }
}

function loadRecoveryCache() {
  if (!recoveryCacheEnabled) return;
  try {
    if (!existsSync(recoveryCachePath)) return;
    const raw = readFileSync(recoveryCachePath, "utf8");
    const data = JSON.parse(raw);
    const now = Date.now();
    const staleEntries = data?.staleCooldown || {};
    for (const [id, entry] of Object.entries(staleEntries)) {
      if (!entry?.lastCheck) continue;
      if (now - entry.lastCheck > STALE_COOLDOWN_MS) continue;
      staleBranchCooldown.set(id, entry);
    }
    const skipEntries = data?.recoverySkip || {};
    for (const [id, entry] of Object.entries(skipEntries)) {
      if (!entry?.timestamp) continue;
      if (now - entry.timestamp > RECOVERY_SKIP_CACHE_MS) continue;
      recoverySkipCache.set(id, entry);
    }
    const logEntries = data?.noAttemptLog || {};
    for (const [id, entry] of Object.entries(logEntries)) {
      if (!entry?.lastLogAt) continue;
      if (
        recoveryLogDedupMs > 0 &&
        now - entry.lastLogAt > recoveryLogDedupMs
      ) {
        continue;
      }
      noAttemptLogCache.set(id, entry);
    }
    const total =
      staleBranchCooldown.size +
      recoverySkipCache.size +
      noAttemptLogCache.size;
    if (total > 0) {
      console.log(
        `[monitor] Restored ${total} recovery cache entries (stale=${staleBranchCooldown.size}, skip=${recoverySkipCache.size}, logs=${noAttemptLogCache.size})`,
      );
    }
  } catch {
    /* best-effort */
  }
}

function clearRecoveryCaches(taskId) {
  let changed = false;
  if (staleBranchCooldown.delete(taskId)) changed = true;
  if (recoverySkipCache.delete(taskId)) changed = true;
  if (noAttemptLogCache.delete(taskId)) changed = true;
  if (changed) scheduleRecoveryCacheSave();
}

function shouldLogNoAttempt(task, taskStatus, reason) {
  if (!recoveryCacheEnabled || recoveryLogDedupMs <= 0) return true;
  const entry = noAttemptLogCache.get(task.id);
  if (!entry) return true;
  if (entry.reason && entry.reason !== reason) return true;
  if (!taskVersionMatches(task, entry, taskStatus)) {
    noAttemptLogCache.delete(task.id);
    scheduleRecoveryCacheSave();
    return true;
  }
  return Date.now() - entry.lastLogAt >= recoveryLogDedupMs;
}

function recordNoAttemptLog(task, taskStatus, reason) {
  if (!recoveryCacheEnabled) return;
  const updatedAt = getTaskUpdatedAt(task);
  if (!updatedAt) return;
  noAttemptLogCache.set(task.id, {
    lastLogAt: Date.now(),
    updatedAt,
    status: taskStatus,
    reason,
  });
  scheduleRecoveryCacheSave();
}

/** Maximum number of tasks to process per sweep (0 = unlimited) */
const MERGE_CHECK_BATCH_SIZE = 0;

/** Small delay between GitHub API calls to avoid rate-limiting (ms) */
const MERGE_CHECK_THROTTLE_MS = 1500;

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
const CONFLICT_MAX_ATTEMPTS = 3; // Max resolution attempts per task before giving up
const conflictResolutionAttempts = new Map(); // task ID → attempt count

const RECOVERY_SKIP_CACHE_MS = 30 * 60 * 1000; // 30 minutes

// Load recovery cache on startup (after constants initialize)
loadRecoveryCache();

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
      // Find the attempt associated with this task — first in local status,
      // then fall back to the VK API (which includes archived attempts)
      let attempt = attempts.find((a) => a?.task_id === task.id);
      if (!attempt) {
        // VK API fallback: find the most recent attempt for this task
        const vkMatch = vkAttempts
          .filter((a) => a?.task_id === task.id)
          .sort(
            (a, b) =>
              new Date(b.created_at).getTime() -
              new Date(a.created_at).getTime(),
          );
        if (vkMatch.length > 0) {
          attempt = vkMatch[0];
          console.log(
            `[monitor] Found VK attempt for task "${task.title}" via API fallback (branch: ${attempt.branch})`,
          );
        } else {
          if (shouldLogNoAttempt(task, taskStatus, "no_attempt")) {
            console.log(
              `[monitor] No attempt found for task "${task.title}" (${task.id.substring(0, 8)}...) — cannot resolve branch/PR`,
            );
            recordNoAttemptLog(task, taskStatus, "no_attempt");
          }
        }
      }
      const branch =
        attempt?.branch ||
        task?.branch ||
        task?.workspace_branch ||
        task?.git_branch;
      const prNumber =
        attempt?.pr_number ||
        task?.pr_number ||
        parsePrNumberFromUrl(attempt?.pr_url) ||
        parsePrNumberFromUrl(task?.pr_url);
      let prInfo = null;
      if (prNumber) {
        prInfo = await getPullRequestByNumber(prNumber);
      }
      const isMerged =
        !!prInfo?.mergedAt ||
        (!!prInfo?.merged_at && prInfo.merged_at !== null);
      const prState = prInfo?.state ? String(prInfo.state).toUpperCase() : "";

      // ── Skip cancelled/done tasks — they should never be recovered ──
      if (taskStatus === "cancelled" || taskStatus === "done") {
        continue;
      }

      // ── Recovery skip cache: skip tasks we already resolved recently ──
      // safeRecoverTask caches tasks that are already todo/cancelled/done,
      // so we skip the entire branch/PR lookup and recovery attempt.
      const skipEntry = recoverySkipCache.get(task.id);
      if (skipEntry) {
        if (!taskVersionMatches(task, skipEntry, taskStatus)) {
          recoverySkipCache.delete(task.id);
          scheduleRecoveryCacheSave();
        } else if (Date.now() - skipEntry.timestamp < RECOVERY_SKIP_CACHE_MS) {
          continue;
        }
      }

      // ── Stale cooldown: skip tasks we already checked recently ──
      const staleEntry = staleBranchCooldown.get(task.id);
      if (staleEntry) {
        if (!taskVersionMatches(task, staleEntry, taskStatus)) {
          staleBranchCooldown.delete(task.id);
          scheduleRecoveryCacheSave();
        } else if (Date.now() - staleEntry.lastCheck < STALE_COOLDOWN_MS) {
          continue;
        }
      }

      // ── Gather ALL attempts for this task (local + VK API) ──
      // VK can have multiple attempts with different branches. An older
      // attempt may have the merged PR while the newest was abandoned.
      const localAttempt = attempts.find((a) => a?.task_id === task.id);
      const allVkAttempts = vkAttempts
        .filter((a) => a?.task_id === task.id)
        .sort(
          (a, b) =>
            new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
        );

      // Build a deduplicated list of all branches + PR numbers to check
      /** @type {Array<{branch?: string, prNumber?: number, attemptId?: string, baseBranch?: string}>} */
      const candidates = [];
      const seenBranches = new Set();

      const addCandidate = (src) => {
        const b = src?.branch;
        const pr = src?.pr_number || parsePrNumberFromUrl(src?.pr_url);
        const aid = src?.id; // attempt UUID
        const baseBranch = resolveAttemptTargetBranch(src, task);
        if (b && !seenBranches.has(b)) {
          seenBranches.add(b);
          candidates.push({
            branch: b,
            prNumber: pr || undefined,
            attemptId: aid,
            baseBranch,
          });
        } else if (b && baseBranch) {
          const existing = candidates.find((c) => c.branch === b);
          if (existing && !existing.baseBranch) {
            existing.baseBranch = baseBranch;
          }
        } else if (pr && !candidates.some((c) => c.prNumber === pr)) {
          candidates.push({
            branch: b,
            prNumber: pr,
            attemptId: aid,
            baseBranch,
          });
        }
      };

      if (localAttempt) addCandidate(localAttempt);
      for (const a of allVkAttempts) addCandidate(a);
      // Also check task-level fields
      addCandidate({
        branch: task?.branch || task?.workspace_branch || task?.git_branch,
        pr_number: task?.pr_number,
        pr_url: task?.pr_url,
      });

      if (candidates.length > 0) {
        if (noAttemptLogCache.delete(task.id)) {
          scheduleRecoveryCacheSave();
        }
      }

      if (candidates.length === 0) {
        // ── Internal executor guard ──
        // If the internal executor is managing this task (active, cooldown,
        // or blocked), do NOT recover it — the executor handles its own lifecycle.
        if (
          internalTaskExecutor &&
          internalTaskExecutor.isTaskManaged?.(task.id)
        ) {
          if (
            shouldLogNoAttempt(task, taskStatus, "internal_executor_managed")
          ) {
            console.log(
              `[monitor] Task "${task.title}" (${task.id.substring(0, 8)}...) is managed by internal executor — skipping recovery`,
            );
            recordNoAttemptLog(task, taskStatus, "internal_executor_managed");
          }
          continue;
        }

        // ── Only recover idle inprogress tasks — never inreview ──
        // inreview tasks are monitored by merge/conflict checks.
        // inprogress tasks with an active agent should not be touched.
        if (taskStatus !== "inprogress") {
          if (shouldLogNoAttempt(task, taskStatus, "no_attempt_skip_status")) {
            console.log(
              `[monitor] No attempt found for task "${task.title}" (${task.id.substring(0, 8)}...) in ${taskStatus} — skipping (only idle inprogress tasks are recovered)`,
            );
            recordNoAttemptLog(task, taskStatus, "no_attempt_skip_status");
          }
          continue;
        }

        // Check if an agent is actively working on this task
        const hasActiveAgent =
          task.has_in_progress_attempt === true || !!localAttempt;
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
            `[monitor] No attempt found for idle task "${task.title}" (${task.id.substring(0, 8)}...) — stale for ${ageHours}h, attempting recovery`,
          );
          const success = await safeRecoverTask(
            task.id,
            task.title,
            `age-based: ${ageHours}h, no agent, no branch/PR`,
          );
          if (success) {
            movedTodoCount++;
            recoveredTaskNames.push(task.title);
            staleBranchCooldown.delete(task.id);
            scheduleRecoveryCacheSave();
          }
          continue;
        }

        const prev = staleBranchCooldown.get(task.id);
        const strikes = (prev?.strikes || 0) + 1;
        staleBranchCooldown.set(task.id, {
          lastCheck: Date.now(),
          strikes,
          updatedAt: getTaskUpdatedAt(task),
          status: taskStatus,
        });
        scheduleRecoveryCacheSave();
        console.log(
          `[monitor] No attempt found for idle task "${task.title}" (${task.id.substring(0, 8)}...) — strike ${strikes}/${STALE_MAX_STRIKES}`,
        );
        if (strikes >= STALE_MAX_STRIKES) {
          const success = await safeRecoverTask(
            task.id,
            task.title,
            `no branch/PR after ${strikes} checks`,
          );
          if (success) {
            movedTodoCount++;
            recoveredTaskNames.push(task.title);
            staleBranchCooldown.delete(task.id);
            scheduleRecoveryCacheSave();
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
            // Assess completion confidence for merged PR
            const sizeLabel =
              task.title?.match(/\[(xs|s|m|l|xl|xxl)\]/i)?.[1] || "m";
            const taskComplexity = classifyComplexity({
              sizeLabel,
              title: task.title,
              description: task.description,
            });
            const confidence = assessCompletionConfidence({
              testsPass: true, // PR was merged → CI must have passed
              buildClean: true,
              lintClean: true,
              filesChanged: prInfo?.changed_files || prInfo?.changedFiles || 0,
              attemptCount: allVkAttempts.length || 1,
              complexityTier: taskComplexity.tier,
            });
            console.log(
              `[monitor] Task "${task.title}" (${task.id.substring(0, 8)}...) has merged PR #${cand.prNumber}, updating to done [confidence=${confidence.confidence}, ${confidence.reason}]`,
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
            // ── Trigger downstream rebase for tasks on same upstream ──
            const mergedBase =
              cand.baseBranch ||
              resolveUpstreamFromTask(task) ||
              DEFAULT_TARGET_BRANCH;
            void rebaseDownstreamTasks(mergedBase, cand.attemptId);
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
        const merged = await isBranchMerged(cand.branch, cand.baseBranch);
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
          // ── Trigger downstream rebase for tasks on same upstream ──
          const mergedBase2 =
            cand.baseBranch ||
            resolveUpstreamFromTask(task) ||
            DEFAULT_TARGET_BRANCH;
          void rebaseDownstreamTasks(mergedBase2, cand.attemptId);
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
                // Register as dirty for slot reservation + file-overlap
                registerDirtyTask({
                  taskId: task.id,
                  prNumber: branchPr.number,
                  branch: cand.branch,
                  title: task.title,
                  files: fullPrInfo?.files?.map((f) => f.filename || f) || [],
                });
              }
            }
          }
        }
      }

      if (resolved) continue;

      // ── Conflict resolution for open PRs with merge conflicts ──
      // DEDUPLICATION: The PS1 orchestrator owns direct rebase with persistent
      // disk-based cooldowns (survives restarts). monitor.mjs only defers to
      // the orchestrator by logging and registering the dirty task for slot
      // reservation. We do NOT trigger smartPRFlow("conflict") here to avoid
      // the thundering herd where both systems race to fix the same PR.
      if (conflictCandidates.length > 0) {
        const lastConflictCheck = conflictResolutionCooldown.get(task.id);
        const onCooldown =
          lastConflictCheck &&
          Date.now() - lastConflictCheck < CONFLICT_COOLDOWN_MS;
        const onDirtyCooldown = isOnResolutionCooldown(task.id);
        if (!onCooldown && !onDirtyCooldown) {
          // Check if we've exhausted max resolution attempts for this task
          const attempts = conflictResolutionAttempts.get(task.id) || 0;
          if (attempts >= CONFLICT_MAX_ATTEMPTS) {
            console.warn(
              `[monitor] ⚠️ Task "${task.title}" PR #${conflictCandidates[0].prNumber} conflict resolution exhausted (${attempts}/${CONFLICT_MAX_ATTEMPTS} attempts) — skipping`,
            );
          } else {
            conflictResolutionAttempts.set(task.id, attempts + 1);
            const cc = conflictCandidates[0];
            let resolveAttemptId = cc.attemptId;
            if (!resolveAttemptId) {
              const matchAttempt = allVkAttempts.find(
                (a) => a.branch === cc.branch || a.pr_number === cc.prNumber,
              );
              resolveAttemptId = matchAttempt?.id || localAttempt?.id;
            }
            if (resolveAttemptId) {
              const shortId = resolveAttemptId.substring(0, 8);
              conflictResolutionCooldown.set(task.id, Date.now());
              recordResolutionAttempt(task.id);

              const sdkOnCooldown = isSDKResolutionOnCooldown(cc.branch);
              const sdkExhausted = isSDKResolutionExhausted(cc.branch);

              if (!sdkOnCooldown && !sdkExhausted) {
                console.log(
                  `[monitor] ⚠️ Task "${task.title}" PR #${cc.prNumber} has merge conflicts — launching SDK resolver (attempt ${shortId})`,
                );
                if (telegramToken && telegramChatId) {
                  void sendTelegramMessage(
                    `🔀 PR #${cc.prNumber} for "${task.title}" has merge conflicts — launching SDK resolver (attempt ${shortId})`,
                  );
                }

                let worktreePath = null;
                const attemptInfo = await getAttemptInfo(resolveAttemptId);
                worktreePath =
                  attemptInfo?.worktree_dir || attemptInfo?.worktree || null;
                if (!worktreePath) {
                  worktreePath = findWorktreeForBranch(cc.branch);
                }

                // Create worktree via centralized manager if none found
                if (!worktreePath && cc.branch) {
                  try {
                    const taskKey = task.id || cc.branch;
                    const wt = await acquireWorktree(cc.branch, taskKey, {
                      owner: "monitor-conflict",
                    });
                    if (wt?.path) {
                      worktreePath = wt.path;
                      console.log(
                        `[monitor] Acquired worktree for ${cc.branch} at ${wt.path} (${wt.created ? "created" : "existing"})`,
                      );
                    }
                  } catch (wErr) {
                    console.warn(
                      `[monitor] Worktree acquisition error: ${wErr.message}`,
                    );
                  }
                }

                if (worktreePath) {
                  void (async () => {
                    try {
                      const result = await resolveConflictsWithSDK({
                        worktreePath,
                        branch: cc.branch,
                        baseBranch: resolveAttemptTargetBranch(
                          attemptInfo,
                          task,
                        ),
                        prNumber: cc.prNumber,
                        taskTitle: task.title,
                        taskDescription: task.description || "",
                        logDir: logDir,
                      });
                      if (result.success) {
                        console.log(
                          `[monitor] ✅ SDK resolved conflicts for PR #${cc.prNumber} (${result.resolvedFiles.length} files)`,
                        );
                        clearDirtyTask(task.id);
                        clearSDKResolutionState(cc.branch);
                        conflictResolutionAttempts.delete(task.id); // Reset on success
                        if (telegramToken && telegramChatId) {
                          void sendTelegramMessage(
                            `✅ SDK resolved merge conflicts for PR #${cc.prNumber} "${task.title}" (${result.resolvedFiles.length} files)`,
                          );
                        }
                      } else {
                        console.warn(
                          `[monitor] ❌ SDK conflict resolution failed for PR #${cc.prNumber}: ${result.error}`,
                        );
                        if (telegramToken && telegramChatId) {
                          void sendTelegramMessage(
                            `❌ SDK conflict resolution failed for PR #${cc.prNumber} "${task.title}": ${result.error}\nFalling back to orchestrator.`,
                          );
                        }
                        conflictsTriggered++;
                        void smartPRFlow(resolveAttemptId, shortId, "conflict");
                      }
                    } catch (err) {
                      console.warn(
                        `[monitor] SDK conflict resolution threw: ${err.message}`,
                      );
                    }
                  })();
                } else {
                  console.warn(
                    `[monitor] No worktree found for ${cc.branch} — deferring to orchestrator`,
                  );
                  if (telegramToken && telegramChatId) {
                    void sendTelegramMessage(
                      `🔀 PR #${cc.prNumber} for "${task.title}" has merge conflicts — no worktree, orchestrator will handle (attempt ${shortId})`,
                    );
                  }
                  conflictsTriggered++;
                  void smartPRFlow(resolveAttemptId, shortId, "conflict");
                }
              } else {
                const reason = sdkExhausted
                  ? "SDK attempts exhausted"
                  : "SDK on cooldown";
                console.log(
                  `[monitor] ⚠️ Task "${task.title}" PR #${cc.prNumber} has merge conflicts — ${reason}, deferring to orchestrator (attempt ${shortId})`,
                );
                if (telegramToken && telegramChatId) {
                  void sendTelegramMessage(
                    `🔀 PR #${cc.prNumber} for "${task.title}" has merge conflicts — ${reason}, orchestrator will handle (attempt ${shortId})`,
                  );
                }
                conflictsTriggered++;
                void smartPRFlow(resolveAttemptId, shortId, "conflict");
              }
            } else {
              console.warn(
                `[monitor] Task "${task.title}" PR #${cc.prNumber} has conflicts but no attempt ID — cannot trigger resolution`,
              );
            }
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
          task.has_in_progress_attempt === true || !!localAttempt;
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
            `[monitor] Idle task "${task.title}" (${task.id.substring(0, 8)}...): no branch/PR, stale for ${ageHours}h — attempting recovery`,
          );
          const success = await safeRecoverTask(
            task.id,
            task.title,
            `age-based: ${ageHours}h, no agent, no branch/PR`,
          );
          if (success) {
            movedTodoCount++;
            recoveredTaskNames.push(task.title);
            staleBranchCooldown.delete(task.id);
            scheduleRecoveryCacheSave();
          }
        } else {
          // Not old enough — use the strike-based system
          const prev = staleBranchCooldown.get(task.id);
          const strikes = (prev?.strikes || 0) + 1;
          staleBranchCooldown.set(task.id, {
            lastCheck: Date.now(),
            strikes,
            updatedAt: getTaskUpdatedAt(task),
            status: taskStatus,
          });
          scheduleRecoveryCacheSave();
          console.log(
            `[monitor] Idle task "${task.title}" (${task.id.substring(0, 8)}...): no branch, no PR (strike ${strikes}/${STALE_MAX_STRIKES})`,
          );
          if (strikes >= STALE_MAX_STRIKES) {
            const success = await safeRecoverTask(
              task.id,
              task.title,
              `abandoned — ${strikes} stale checks`,
            );
            if (success) {
              movedTodoCount++;
              recoveredTaskNames.push(task.title);
              staleBranchCooldown.delete(task.id);
              scheduleRecoveryCacheSave();
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
    console.log(`[monitor] ${formatDirtyTaskSummary()}`);
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
    return {
      checked: 0,
      movedDone: 0,
      movedReview: 0,
      movedTodo: 0,
      error: err,
    };
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
    const telegramFn =
      telegramToken && telegramChatId
        ? (msg) => void sendTelegramMessage(msg)
        : null;

    const decision = await analyzeMergeStrategy(ctx, {
      execCodex: execPooledPrompt,
      timeoutMs:
        parseInt(process.env.MERGE_STRATEGY_TIMEOUT_MS, 10) || 10 * 60 * 1000,
      logDir,
      onTelegram: telegramFn,
    });

    if (!decision || !decision.success) {
      console.warn(`[${tag}] analysis failed — falling back to manual review`);
      return;
    }

    // ── Execute the decision via centralized executor ────────────
    console.log(
      `[${tag}] → ${decision.action}${decision.reason ? ": " + decision.reason.slice(0, 100) : ""}`,
    );

    const execResult = await executeDecision(decision, ctx, {
      logDir,
      onTelegram: telegramFn,
      timeoutMs:
        parseInt(process.env.MERGE_STRATEGY_TIMEOUT_MS, 10) || 15 * 60 * 1000,
    });

    // ── Post-execution handling ──────────────────────────────────
    if (execResult.action === "wait" && execResult.waitSeconds) {
      // Re-run analysis after the wait period
      setTimeout(
        () => {
          void runMergeStrategyAnalysis({
            ...ctx,
            ciStatus: "re-check",
          });
        },
        (execResult.waitSeconds || 300) * 1000,
      );
    }

    if (!execResult.success && execResult.error) {
      console.warn(`[${tag}] execution issue: ${execResult.error}`);
    }
  } catch (err) {
    console.warn(
      `[${tag}] merge strategy analysis error: ${err.message || err}`,
    );
  }
}

// ── Auto-Rebase Downstream Tasks on PR Merge ────────────────────────────────

/**
 * When a PR is merged into an upstream branch, find all active tasks that
 * share the same upstream and trigger a rebase on each of them.
 *
 * This prevents tasks from drifting behind their upstream and accumulating
 * merge conflicts.
 *
 * @param {string} mergedUpstreamBranch - The branch the PR was merged into
 * @param {string} [excludeAttemptId]   - Attempt to exclude (the one that just merged)
 */
async function rebaseDownstreamTasks(mergedUpstreamBranch, excludeAttemptId) {
  if (!branchRouting?.autoRebaseOnMerge) {
    console.log("[rebase-downstream] auto-rebase disabled in config");
    return;
  }

  const tag = "rebase-downstream";
  console.log(
    `[${tag}] PR merged into ${mergedUpstreamBranch} — checking for downstream tasks to rebase`,
  );

  try {
    // Get all active tasks
    const statuses = ["inprogress", "inreview"];
    const tasksByStatus = await Promise.all(
      statuses.map((status) => fetchTasksByStatus(status)),
    );
    const allTasks = [];
    for (const tasks of tasksByStatus) {
      for (const task of tasks) {
        if (task?.id) allTasks.push(task);
      }
    }

    // Get active attempts from status file
    const statusData = await readStatusData();
    const attempts = Array.isArray(statusData?.active_attempts)
      ? statusData.active_attempts
      : Object.values(statusData?.attempts || {});

    // Also fetch VK task-attempts as fallback
    let vkAttempts = [];
    try {
      const vkRes = await fetchVk("/api/task-attempts");
      const vkData = vkRes?.data ?? vkRes;
      if (Array.isArray(vkData)) vkAttempts = vkData;
    } catch {
      /* best-effort */
    }

    let rebasedCount = 0;
    let failedCount = 0;
    const rebaseResults = [];

    for (const task of allTasks) {
      // Resolve this task's upstream branch
      const taskUpstream =
        resolveUpstreamFromTask(task) || DEFAULT_TARGET_BRANCH;

      // Normalize both branches for comparison (strip "origin/" prefix)
      const normalize = (b) => b?.replace(/^origin\//, "") || "";
      if (normalize(taskUpstream) !== normalize(mergedUpstreamBranch)) {
        continue; // Different upstream — not affected
      }

      // Find the attempt for this task
      let attempt = attempts.find((a) => a?.task_id === task.id);
      if (!attempt) {
        const vkMatch = vkAttempts
          .filter((a) => a?.task_id === task.id)
          .sort(
            (a, b) =>
              new Date(b.created_at).getTime() -
              new Date(a.created_at).getTime(),
          );
        if (vkMatch.length > 0) attempt = vkMatch[0];
      }

      if (!attempt || attempt.id === excludeAttemptId) continue;
      if (!attempt.branch) continue;

      console.log(
        `[${tag}] rebasing task "${task.title}" (${attempt.id.substring(0, 8)}) onto ${mergedUpstreamBranch}`,
      );

      try {
        const rebaseResult = await rebaseAttempt(
          attempt.id,
          mergedUpstreamBranch,
        );

        if (rebaseResult?.success || rebaseResult?.data?.success) {
          rebasedCount++;
          rebaseResults.push({
            taskTitle: task.title,
            attemptId: attempt.id,
            status: "success",
          });
          console.log(
            `[${tag}] ✓ rebased "${task.title}" (${attempt.id.substring(0, 8)}) onto ${mergedUpstreamBranch}`,
          );
        } else {
          failedCount++;
          const error =
            rebaseResult?.error || rebaseResult?.message || "unknown";
          rebaseResults.push({
            taskTitle: task.title,
            attemptId: attempt.id,
            status: "failed",
            error,
          });
          console.warn(
            `[${tag}] ✗ rebase failed for "${task.title}" (${attempt.id.substring(0, 8)}): ${error}`,
          );

          // ── Run task assessment on rebase failure ──────────────
          if (branchRouting?.assessWithSdk && agentPoolEnabled) {
            void runTaskAssessment({
              taskId: task.id,
              taskTitle: task.title,
              taskDescription: task.description,
              attemptId: attempt.id,
              shortId: attempt.id.substring(0, 8),
              trigger: "rebase_failed",
              branch: attempt.branch,
              upstreamBranch: mergedUpstreamBranch,
              rebaseError: error,
              conflictFiles:
                rebaseResult?.conflicted_files ||
                rebaseResult?.data?.conflicted_files ||
                [],
            });
          }
        }
      } catch (err) {
        failedCount++;
        rebaseResults.push({
          taskTitle: task.title,
          attemptId: attempt.id,
          status: "error",
          error: err.message || String(err),
        });
        console.warn(
          `[${tag}] error rebasing "${task.title}": ${err.message || err}`,
        );
      }
    }

    if (rebasedCount > 0 || failedCount > 0) {
      const summary = `Downstream rebase after merge to ${mergedUpstreamBranch}: ${rebasedCount} rebased, ${failedCount} failed`;
      console.log(`[${tag}] ${summary}`);
      void sendTelegramMessage(
        `🔄 ${summary}\n${rebaseResults.map((r) => `  ${r.status === "success" ? "✓" : "✗"} ${r.taskTitle}`).join("\n")}`,
      );
    } else {
      console.log(
        `[${tag}] no downstream tasks found on upstream ${mergedUpstreamBranch}`,
      );
    }
  } catch (err) {
    console.warn(`[${tag}] error: ${err.message || err}`);
  }
}

// ── Task Assessment Integration ─────────────────────────────────────────────

/**
 * Run a full task lifecycle assessment using Codex/Copilot SDK.
 * First tries quickAssess (heuristic, no SDK call), then falls back to
 * full SDK assessment if needed.
 *
 * After getting a decision, ACTS on it — sends prompts, triggers retries, etc.
 *
 * @param {import("./task-assessment.mjs").TaskAssessmentContext} ctx
 */
async function runTaskAssessment(ctx) {
  const tag = `assessment(${ctx.shortId})`;
  try {
    // ── Quick heuristic assessment first ───────────────────
    const quick = quickAssess(ctx);
    if (quick) {
      console.log(
        `[${tag}] quick decision: ${quick.action} — ${(quick.reason || "").slice(0, 100)}`,
      );
      await actOnAssessment(ctx, quick);
      return;
    }

    // ── Full SDK assessment ───────────────────────────────
    if (!agentPoolEnabled) {
      console.log(`[${tag}] skipping SDK assessment — agent disabled`);
      return;
    }

    const telegramFn =
      telegramToken && telegramChatId
        ? (msg) => void sendTelegramMessage(msg)
        : null;

    const decision = await assessTask(ctx, {
      execCodex: execPooledPrompt,
      timeoutMs: 5 * 60 * 1000,
      logDir,
      onTelegram: telegramFn,
    });

    if (!decision?.success) {
      console.warn(`[${tag}] assessment failed — no action taken`);
      return;
    }

    await actOnAssessment(ctx, decision);
  } catch (err) {
    console.warn(`[${tag}] error: ${err.message || err}`);
  }
}

/**
 * Act on an assessment decision — execute the recommended action.
 *
 * @param {import("./task-assessment.mjs").TaskAssessmentContext} ctx
 * @param {import("./task-assessment.mjs").TaskAssessmentDecision} decision
 */
async function actOnAssessment(ctx, decision) {
  const tag = `assessment-act(${ctx.shortId})`;

  switch (decision.action) {
    case "merge":
      console.log(`[${tag}] → merge`);
      // Handled by VK cleanup script / auto-merge
      break;

    case "reprompt_same":
      console.log(`[${tag}] → reprompt same session`);
      if (decision.prompt && agentPoolEnabled) {
        void execPooledPrompt(decision.prompt, { timeoutMs: 15 * 60 * 1000 });
      }
      break;

    case "reprompt_new_session":
      console.log(`[${tag}] → reprompt new session`);
      if (typeof startFreshSession === "function") {
        startFreshSession(
          decision.prompt || `Resume task: ${ctx.taskTitle}`,
          decision.reason,
        );
      } else if (typeof attemptFreshSessionRetry === "function") {
        await attemptFreshSessionRetry(
          "assessment_new_session",
          decision.reason || "Assessment recommended new session",
        );
      }
      break;

    case "new_attempt":
      console.log(
        `[${tag}] → new attempt (agent: ${decision.agentType || "auto"})`,
      );
      // Move task back to todo for re-scheduling
      if (ctx.taskId) {
        await updateTaskStatus(ctx.taskId, "todo");
      }
      void sendTelegramMessage(
        `🆕 Assessment: starting new attempt for "${ctx.taskTitle}" — ${decision.reason || ""}`,
      );
      break;

    case "wait": {
      const waitSec = decision.waitSeconds || 300;
      console.log(`[${tag}] → wait ${waitSec}s`);
      setTimeout(() => {
        void runTaskAssessment({
          ...ctx,
          trigger: "reassessment",
        });
      }, waitSec * 1000);
      break;
    }

    case "manual_review":
      console.log(`[${tag}] → manual review`);
      void sendTelegramMessage(
        `👀 Assessment: manual review needed for "${ctx.taskTitle}" — ${decision.reason || ""}`,
      );
      break;

    case "close_and_replan":
      console.log(`[${tag}] → close and replan`);
      if (ctx.taskId) {
        await updateTaskStatus(ctx.taskId, "todo");
      }
      void sendTelegramMessage(
        `🚫 Assessment: closing and replanning "${ctx.taskTitle}" — ${decision.reason || ""}`,
      );
      break;

    case "noop":
      console.log(`[${tag}] → noop`);
      break;

    default:
      console.warn(`[${tag}] unknown action: ${decision.action}`);
  }
}

// ── Smart PR creation flow ──────────────────────────────────────────────────

// Use config-driven branch routing instead of hardcoded defaults
const DEFAULT_TARGET_BRANCH =
  branchRouting?.defaultBranch || process.env.VK_TARGET_BRANCH || "origin/main";
const DEFAULT_CODEX_MONITOR_UPSTREAM =
  branchRouting?.scopeMap?.["codex-monitor"] ||
  process.env.CODEX_MONITOR_TASK_UPSTREAM ||
  "origin/ve/codex-monitor-generic";

/**
 * Extract the conventional commit scope from a task title.
 * E.g. "feat(codex-monitor): add caching" → "codex-monitor"
 *      "[P1] fix(veid): broken flow"      → "veid"
 *      "chore(provider): cleanup"         → "provider"
 * @param {string} title
 * @returns {string|null}
 */
function extractScopeFromTitle(title) {
  if (!title) return null;
  // Match conventional commit patterns: type(scope): ... or [P*] type(scope): ...
  const match = String(title).match(
    /(?:^\[P\d+\]\s*)?(?:feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)\(([^)]+)\)/i,
  );
  return match ? match[1].toLowerCase().trim() : null;
}

/**
 * Resolve the upstream branch for a task using config-based scope routing.
 * Priority:
 *   1. Task-level explicit fields (target_branch, base_branch, etc.)
 *   2. Task metadata fields
 *   3. Task labels with upstream/base/target patterns
 *   4. Text body extraction
 *   5. Config scopeMap matching (title scope → branch)
 *   6. Config scopeMap matching (keyword-based)
 *   7. Legacy codex-monitor keyword detection
 *   8. Config defaultBranch
 * @param {object} task
 * @returns {string|null}
 */
function resolveUpstreamFromConfig(task) {
  if (!task) return null;

  // ── Priority 5+: Config-based scope routing ──────────────
  const scope = extractScopeFromTitle(task.title || task.name);
  if (scope && branchRouting?.scopeMap) {
    // Exact scope match
    const exactMatch = branchRouting.scopeMap[scope];
    if (exactMatch) return exactMatch;

    // Partial scope match — check if any config key is contained in the scope
    for (const [key, branch] of Object.entries(branchRouting.scopeMap)) {
      if (scope.includes(key) || key.includes(scope)) return branch;
    }
  }

  // ── Priority 6: Keyword-based scope matching from task text ─
  if (branchRouting?.scopeMap) {
    const text = getTaskTextBlob(task).toLowerCase();
    for (const [key, branch] of Object.entries(branchRouting.scopeMap)) {
      // Check if the routing key appears as a keyword in the task text
      if (text.includes(key.toLowerCase())) return branch;
    }
  }

  return null;
}

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
  for (const field of [
    "labels",
    "label",
    "tags",
    "tag",
    "categories",
    "category",
  ]) {
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
  for (const field of [
    "title",
    "name",
    "description",
    "body",
    "details",
    "content",
  ]) {
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
      if (task.metadata[field])
        return normalizeBranchName(task.metadata[field]);
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

  // ── Config-based scope routing ────────────────────────────
  const fromConfig = resolveUpstreamFromConfig(task);
  if (fromConfig) return fromConfig;

  // ── Legacy codex-monitor keyword detection ────────────────
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

// ── Conflict Classification ─────────────────────────────────────────────────
// Auto-resolvable file patterns for rebase conflicts:
//   "theirs" = accept upstream version (lock files, generated files)
//   "ours"   = keep our version (changelogs, coverage reports)
const AUTO_RESOLVE_THEIRS = [
  "pnpm-lock.yaml",
  "package-lock.json",
  "yarn.lock",
  "go.sum",
];
const AUTO_RESOLVE_OURS = ["CHANGELOG.md", "coverage.txt", "results.txt"];
const AUTO_RESOLVE_LOCK_EXTENSIONS = [".lock"];

/**
 * Classify conflicted files into auto-resolvable and manual categories.
 * @param {string[]} files - List of conflicted file paths
 * @returns {{ allResolvable: boolean, manualFiles: string[], summary: string }}
 */
function classifyConflictedFiles(files) {
  const manualFiles = [];
  const strategies = [];

  for (const file of files) {
    const fileName = file.split("/").pop();
    let strategy = null;

    if (AUTO_RESOLVE_THEIRS.includes(fileName)) {
      strategy = "theirs";
    } else if (AUTO_RESOLVE_OURS.includes(fileName)) {
      strategy = "ours";
    } else if (
      AUTO_RESOLVE_LOCK_EXTENSIONS.some((ext) => fileName.endsWith(ext))
    ) {
      strategy = "theirs";
    }

    if (strategy) {
      strategies.push(`${fileName}→${strategy}`);
    } else {
      manualFiles.push(file);
    }
  }

  return {
    allResolvable: manualFiles.length === 0,
    manualFiles,
    summary: strategies.join(", ") || "none",
  };
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
    // ── Step 0: Check if task/branch is already merged ───────────
    // Prevents infinite retry loops for tasks that were completed in previous sessions
    const attemptInfo = await getAttemptInfo(attemptId);
    let taskData = null;
    if (attemptInfo?.branch) {
      if (mergedBranchCache.has(attemptInfo.branch)) {
        console.log(
          `[monitor] ${tag}: branch already in merged cache — archiving`,
        );
        await archiveAttempt(attemptId);
        return;
      }
      const merged = await isBranchMerged(attemptInfo.branch);
      if (merged) {
        console.log(
          `[monitor] ${tag}: branch ${attemptInfo.branch} confirmed merged — completing task`,
        );
        mergedBranchCache.add(attemptInfo.branch);
        if (attemptInfo.task_id) {
          mergedTaskCache.add(attemptInfo.task_id);
          void updateTaskStatus(attemptInfo.task_id, "done");
        }
        await archiveAttempt(attemptId);
        saveMergedTaskCache();
        return;
      }
    }

    // ── Step 0b: Check task description for "already completed" signals ──
    if (attemptInfo?.task_id) {
      try {
        const taskRes = await fetchVk(`/api/tasks/${attemptInfo.task_id}`);
        taskData = taskRes?.data || taskRes || null;
        const desc = String(
          taskData?.description || taskData?.body || "",
        ).toLowerCase();
        const completionSignals = [
          "superseded by",
          "already completed",
          "this task has been completed",
          "merged in",
          "completed via",
          "no longer needed",
          "already merged",
        ];
        const isDescComplete = completionSignals.some((s) => desc.includes(s));
        if (isDescComplete) {
          console.log(
            `[monitor] ${tag}: task description indicates already completed — archiving`,
          );
          void updateTaskStatus(attemptInfo.task_id, "done");
          await archiveAttempt(attemptId);
          return;
        }
        if (isPlannerTaskData(taskData)) {
          const verify = await verifyPlannerTaskCompletion(
            taskData,
            attemptInfo,
          );
          if (verify.completed) {
            console.log(
              `[monitor] ${tag}: planner task verified (${verify.createdCount} new task(s)) — marking done`,
            );
            void updateTaskStatus(attemptInfo.task_id, "done");
            await archiveAttempt(attemptId);
            if (telegramToken && telegramChatId) {
              const suffix = verify.sampleTitles?.length
                ? ` Examples: ${verify.sampleTitles.join(", ")}`
                : "";
              void sendTelegramMessage(
                `✅ Task planner verified: ${verify.createdCount} new task(s) detected.${suffix}`,
              );
            }
            return;
          }
          console.warn(
            `[monitor] ${tag}: planner task incomplete — no new backlog tasks detected`,
          );
          void updateTaskStatus(attemptInfo.task_id, "todo");
          await archiveAttempt(attemptId);
          if (telegramToken && telegramChatId) {
            void sendTelegramMessage(
              "⚠️ Task planner incomplete: no new backlog tasks detected. Returned to todo.",
            );
          }
          return;
        }
      } catch {
        /* best effort */
      }
    }

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

    // No commits and no changes → archive stale attempt (unless called for conflict resolution)
    if (
      commits_ahead === 0 &&
      !has_uncommitted_changes &&
      status !== "conflict"
    ) {
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
      if (primaryAgentReady) {
        void execPooledPrompt(
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
    if (!taskData && attempt?.task_id) {
      try {
        const taskRes = await fetchVk(`/api/tasks/${attempt.task_id}`);
        if (taskRes?.success && taskRes.data) {
          taskData = taskRes.data;
        } else if (taskRes?.data || taskRes) {
          taskData = taskRes.data || taskRes;
        }
        if (taskData) {
          attempt.task_title = attempt.task_title || taskData.title;
          attempt.task_description =
            taskData.description || taskData.body || "";
        }
      } catch {
        /* best effort */
      }
    }
    const targetBranch = resolveAttemptTargetBranch(attempt, taskData);

    // ── Step 3: Rebase onto target branch ────────────────────────
    console.log(`[monitor] ${tag}: rebasing onto ${targetBranch}...`);
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
      // Rebase has conflicts → try smart auto-resolve based on file type
      if (errorData?.type === "merge_conflicts") {
        const files = errorData.conflicted_files || [];
        console.warn(
          `[monitor] ${tag}: rebase conflicts in ${files.join(", ")} — attempting smart auto-resolve`,
        );

        // Classify conflicted files
        const autoResolvable = classifyConflictedFiles(files);
        if (autoResolvable.allResolvable) {
          console.log(
            `[monitor] ${tag}: all ${files.length} conflicted files are auto-resolvable (${autoResolvable.summary})`,
          );
        } else {
          console.warn(
            `[monitor] ${tag}: ${autoResolvable.manualFiles.length} files need manual resolution: ${autoResolvable.manualFiles.join(", ")}`,
          );
        }

        // Try VK resolve-conflicts API first (it does "accept ours")
        const resolveResult = await resolveConflicts(attemptId);
        if (resolveResult?.success) {
          console.log(`[monitor] ${tag}: conflicts resolved via VK API`);
        } else {
          const attemptInfo = await getAttemptInfo(attemptId);
          let worktreeDir =
            attemptInfo?.worktree_dir || attemptInfo?.worktree || null;
          // Fallback: look up worktree by branch name from git
          if (!worktreeDir && (attemptInfo?.branch || attempt?.branch)) {
            worktreeDir = findWorktreeForBranch(
              attemptInfo?.branch || attempt?.branch,
            );
          }
          if (codexResolveConflictsEnabled) {
            console.warn(
              `[monitor] ${tag}: auto-resolve failed — running Codex SDK conflict resolution (worktree: ${worktreeDir || "UNKNOWN"})`,
            );
            const classification = classifyConflictedFiles(files);
            const fileGuidance = files
              .map((f) => {
                const fn = f.split("/").pop();
                if (
                  AUTO_RESOLVE_THEIRS.includes(fn) ||
                  AUTO_RESOLVE_LOCK_EXTENSIONS.some((ext) => fn.endsWith(ext))
                ) {
                  return `  - ${f}: Accept THEIRS (upstream version — lock/generated file)`;
                }
                if (AUTO_RESOLVE_OURS.includes(fn)) {
                  return `  - ${f}: Accept OURS (keep our version)`;
                }
                return `  - ${f}: Resolve MANUALLY (inspect both sides, merge intelligently)`;
              })
              .join("\n");
            const prompt = `You are fixing a git rebase conflict in a Vibe-Kanban worktree.
Worktree: ${worktreeDir || "(unknown)"}
Attempt: ${shortId}
Conflicted files: ${files.join(", ") || "(unknown)"}

Per-file resolution strategy:
${fileGuidance}

Instructions:
1) cd into the worktree directory.
2) For each conflicted file, apply the strategy above:
   - THEIRS: git checkout --theirs -- <file> && git add <file>
   - OURS: git checkout --ours -- <file> && git add <file>
   - MANUAL: Open the file, remove conflict markers (<<<< ==== >>>>), merge both sides intelligently, then git add <file>
3) After resolving all files, run: git rebase --continue
4) If more conflicts appear, repeat steps 2-3.
5) Once rebase completes, push the branch: git push --force-with-lease
6) Verify the build still passes if possible.
Return a short summary of what you did and any files that needed manual resolution.`;
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
          if (primaryAgentReady) {
            void execPooledPrompt(
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
      if (codexAnalyzeMergeStrategy) {
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
      if (primaryAgentReady) {
        void execPooledPrompt(
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
      if (primaryAgentReady) {
        void execPooledPrompt(
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

    // ── Early merged-branch check: skip if branch is already merged ──
    const resolvedAttempt = match;
    if (resolvedAttempt?.branch) {
      if (mergedBranchCache.has(resolvedAttempt.branch)) {
        console.log(
          `[monitor] smartPR(${shortId}): branch ${resolvedAttempt.branch} already in mergedBranchCache — skipping`,
        );
        return;
      }
      // Check GitHub for a merged PR with this head branch
      const merged = await isBranchMerged(resolvedAttempt.branch);
      if (merged) {
        console.log(
          `[monitor] smartPR(${shortId}): branch ${resolvedAttempt.branch} confirmed merged — completing task and skipping PR flow`,
        );
        mergedBranchCache.add(resolvedAttempt.branch);
        if (resolvedAttempt.task_id) {
          mergedTaskCache.add(resolvedAttempt.task_id);
          void updateTaskStatus(resolvedAttempt.task_id, "done");
        }
        await archiveAttempt(resolvedAttempt.id || shortId);
        saveMergedTaskCache();
        return;
      }
    }

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
    console.log(
      `[monitor] task planner skipped: codex-sdk mode but Codex disabled`,
    );
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
    console.log(
      `[monitor] task planner skipped: deduped (last triggered ${lastAt})`,
    );
    return;
  }
  try {
    const result = await triggerTaskPlanner(reason, details);
    console.log(
      `[monitor] task planner result: ${result?.status || "unknown"} (${reason})`,
    );
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
  // Orchestrator periodic updates contain counter labels like "Failed: 0" and
  // "error=0" which should NOT trigger error classification.
  const isPositive =
    textLower.includes("✅") ||
    textLower.includes("task completed") ||
    textLower.includes("branch merged") ||
    textLower.includes("pr merged") ||
    (textLower.includes("orchestrator") && textLower.includes("-min update"));

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
  const localMaxParallel = getMaxParallelFromArgs(scriptArgs) || running || 1;

  // Fleet-aware capacity: use total fleet slots when fleet is active
  const fleet = fleetConfig?.enabled ? getFleetState() : null;
  const maxParallel =
    fleet && fleet.mode === "fleet" && fleet.totalSlots > 0
      ? fleet.totalSlots
      : localMaxParallel;
  const backlogPerCapita =
    maxParallel > 0 ? backlogRemaining / maxParallel : backlogRemaining;
  const idleSlots = Math.max(0, maxParallel - running);

  // Fleet-aware backlog depth check: if fleet is active, check if we need
  // more tasks to keep all workstations busy
  if (fleet && fleet.mode === "fleet") {
    const depth = calculateBacklogDepth({
      totalSlots: fleet.totalSlots,
      currentBacklog: backlogRemaining,
      bufferMultiplier: fleetConfig?.bufferMultiplier || 3,
    });
    if (depth.shouldGenerate && depth.deficit > 0) {
      // Only coordinator triggers planner to avoid duplicates
      if (isFleetCoordinator()) {
        await maybeTriggerTaskPlanner("fleet-deficit", {
          backlogRemaining,
          targetDepth: depth.targetDepth,
          deficit: depth.deficit,
          fleetSize: fleet.fleetSize,
          totalSlots: fleet.totalSlots,
          formula: depth.formula,
        });
      }
    }

    // Maintenance mode detection
    const maintenance = detectMaintenanceMode({
      backlog_remaining: backlogRemaining,
      counts,
    });
    if (maintenance.isMaintenanceMode && isFleetCoordinator()) {
      if (!allCompleteNotified) {
        allCompleteNotified = true;
        await sendTelegramMessage(
          `🛰️ Fleet entering maintenance mode: ${maintenance.reason}`,
        );
      }
      return;
    }
  }

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
  const defaultPlannerTaskCount = Number(
    process.env.TASK_PLANNER_DEFAULT_COUNT || "30",
  );
  const numTasks =
    taskCount && Number.isFinite(taskCount) && taskCount > 0
      ? taskCount
      : defaultPlannerTaskCount;
  const plannerTaskSizeLabel = String(
    process.env.TASK_PLANNER_TASK_SIZE_LABEL || "m",
  ).toLowerCase();
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
      title.startsWith("plan next tasks") || title.startsWith("plan next phase")
    );
  });
  if (existingPlanner) {
    console.log(
      `[monitor] task planner VK task already exists in backlog — skipping: "${existingPlanner.title}" (${existingPlanner.id})`,
    );
    const desiredTitle = `[${plannerTaskSizeLabel}] Plan next tasks (${reason || "backlog-empty"})`;
    const desiredDescription = [
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
      `- Create ${numTasks} well-scoped tasks in vibe-kanban (minimum 30)`,
      "- Tasks must be **production-ready** (no placeholders) and thorough",
      "- Every created task should default to **[xl]** unless clearly smaller",
      "- If a placeholder is unavoidable, create a paired follow-up task immediately",
      "- **IMPORTANT:** Every task title MUST start with a size label: [xs], [s], [m], [l], [xl], or [xxl]",
      "  This drives automatic complexity-based model routing for task execution.",
      "- **NOTE:** The planner task itself is [m] so it fits in a single capacity slot",
    ].join("\n");
    // Best-effort: keep backlog task aligned with current requirements
    if (
      existingPlanner.title !== desiredTitle ||
      (existingPlanner.description || "") !== desiredDescription
    ) {
      try {
        await fetchVk(`/api/tasks/${existingPlanner.id}`, {
          method: "PUT",
          body: {
            title: desiredTitle,
            description: desiredDescription,
          },
          timeoutMs: 15000,
        });
        console.log(
          `[monitor] task planner VK task updated with latest requirements (${existingPlanner.id})`,
        );
      } catch {
        /* best-effort */
      }
    }
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
    title: `[${plannerTaskSizeLabel}] Plan next tasks (${reason || "backlog-empty"})`,
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
      `- Create ${numTasks} well-scoped tasks in vibe-kanban (minimum 30)`,
      "- Tasks must be **production-ready** (no placeholders) and thorough",
      "- Every created task should default to **[xl]** unless clearly smaller",
      "- If a placeholder is unavoidable, create a paired follow-up task immediately",
      "- **IMPORTANT:** Every task title MUST start with a size label: [xs], [s], [m], [l], [xl], or [xxl]",
      "  This drives automatic complexity-based model routing for task execution.",
      "- **NOTE:** The planner task itself is [m] so it fits in a single capacity slot",
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
    const result = await runCodexExec(prompt, repoRoot, 1_800_000);

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
  const isSigKill = signal === "SIGKILL";

  // ── Check if this is an intentional restart BEFORE clearing flags ──
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
      console.log(
        `[monitor] intentional restart (${reason}) — skipping autofix`,
      );
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

  // ── External kill (SIGKILL): treat as non-actionable, restart quietly ──
  if (isSigKill) {
    console.warn(
      `[monitor] orchestrator killed by ${reason} — skipping autofix/analysis`,
    );
    restartCount += 1;
    setTimeout(startProcess, restartDelayMs);
    return;
  }

  // ── Benign exit 1: orchestrator ran normally but PowerShell propagated a
  // non-zero $LASTEXITCODE from the last native command (git/gh).  Detect by
  // checking that the log has no actual errors — just normal cycle messages.
  if (
    code === 1 &&
    !signal &&
    logText.length > 200 &&
    !logText.includes("ERROR") &&
    !logText.includes("FATAL") &&
    !logText.includes("Unhandled exception") &&
    (logText.includes("Sleeping") || logText.includes("next cycle"))
  ) {
    console.log(
      `[monitor] benign exit 1 detected (no errors in log, normal cycles) — restarting without autofix`,
    );
    restartCount += 1;
    setTimeout(startProcess, restartDelayMs);
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

  // ── Auto-fix: runs in BACKGROUND only for genuine monitor/orchestrator crashes ──
  // STRICT trigger: only fire when the orchestrator ITSELF crashed (unhandled
  // exception, stack trace from our code, import error, etc.) — NOT when the
  // log merely contains "ERROR" from normal task lifecycle messages.
  //
  // If autofix writes changes, the devmode file watcher triggers a clean restart.
  // If no changes are needed, autofix just logs the outcome — no restart.
  const hasMonitorCrash =
    logText.includes("Unhandled exception") ||
    logText.includes("Unhandled rejection") ||
    logText.includes("SyntaxError:") ||
    logText.includes("ReferenceError:") ||
    logText.includes("TypeError:") ||
    logText.includes("Cannot find module") ||
    logText.includes("FATAL ERROR") ||
    logText.includes("Traceback (most recent call last)") ||
    // PowerShell internal crash
    logText.includes("TerminatingError") ||
    logText.includes("script block termination") ||
    // Very short runtime with high exit code = likely startup crash
    (code > 1 && runDurationMs < 30_000);

  if (autoFixEnabled && logText.length > 0 && hasMonitorCrash) {
    const telegramFn =
      telegramToken && telegramChatId
        ? (msg) => void sendTelegramMessage(msg)
        : null;

    // Fire-and-forget: autofix runs in background, orchestrator restarts now
    void (async () => {
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
          console.log(
            "[monitor] background auto-fix applied — file watcher will restart orchestrator if needed",
          );
          return;
        }

        if (result.outcome && result.outcome !== "clean-exit-skip") {
          console.log(
            `[monitor] background auto-fix outcome: ${result.outcome.slice(0, 100)}`,
          );
        }

        // Auto-fix couldn't help — run diagnostic analysis in background too
        console.log(
          "[monitor] auto-fix unsuccessful — running background Codex analysis",
        );
        await analyzeWithCodex(logPath, logText.slice(-15000), reason);
      } catch (err) {
        console.warn(
          `[monitor] background auto-fix error: ${err.message || err}`,
        );
      }
    })();
  } else if (autoFixEnabled && logText.length > 0 && !hasMonitorCrash) {
    // Not a monitor crash — normal exit with task errors. Skip autofix entirely.
    console.log(
      `[monitor] exit ${reason} — no monitor crash detected — skipping autofix`,
    );
  }

  // ── Context window exhaustion: attempt fresh session (non-blocking) ───
  if (hasContextWindowError(logText)) {
    console.log(
      "[monitor] context window exhaustion detected — attempting fresh session in background",
    );
    void (async () => {
      const freshStarted = await attemptFreshSessionRetry(
        "context_window_exhausted",
        logText.slice(-3000),
      );
      if (freshStarted) {
        console.log(
          "[monitor] fresh session started for context-exhausted task",
        );
      } else {
        await writeFile(
          logPath.replace(/\.log$/, "-context.txt"),
          "Detected context window error. Fresh session retry failed — consider manual recovery.\n",
          "utf8",
        );
      }
    })();
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
            `🛑 Crash loop detected (${restartCountNow} exits in 5m). Pausing orchestrator restarts for ${pauseMin} minutes. Background fix running.`,
          );
        }
        // ── Background crash-loop fix: runs while orchestrator is paused ──
        // Does NOT block handleExit. If it writes changes, file watcher restarts.
        // If it fails, the pause timer will restart the orchestrator anyway.
        if (!orchestratorLoopFixInProgress) {
          orchestratorLoopFixInProgress = true;
          void (async () => {
            try {
              const fixResult = await attemptCrashLoopFix({
                reason,
                logText,
              });
              if (fixResult.fixed) {
                console.log(
                  "[monitor] background crash-loop fix applied — file watcher will handle restart",
                );
                if (telegramToken && telegramChatId) {
                  void sendTelegramMessage(
                    `🛠️ Crash-loop fix applied. File watcher will restart orchestrator.\n${fixResult.outcome}`,
                  );
                }
              } else {
                console.log(
                  `[monitor] background crash-loop fix unsuccessful: ${fixResult.outcome}`,
                );
                // Try fresh session as background last resort
                const freshStarted = await attemptFreshSessionRetry(
                  "crash_loop_unresolvable",
                  logText.slice(-3000),
                );
                if (freshStarted && telegramToken && telegramChatId) {
                  void sendTelegramMessage(
                    `🔄 Crash-loop fix failed but fresh session started. New agent will retry.`,
                  );
                } else if (!freshStarted && telegramToken && telegramChatId) {
                  void sendTelegramMessage(
                    `⚠️ Crash-loop fix failed: ${fixResult.outcome}. Orchestrator will resume after ${pauseMin}m pause.`,
                  );
                }
              }
            } catch (err) {
              console.warn(
                `[monitor] background crash-loop fix error: ${err.message || err}`,
              );
            } finally {
              orchestratorLoopFixInProgress = false;
            }
          })();
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

// ── Devmode Auto Code Fix: background agent triggered after digest seal ──────

/**
 * Called when a Live Digest window is sealed. Collects errors/warnings from the
 * digest, then dispatches a background Codex/Copilot agent to fix the issues.
 *
 * Only active when DEVMODE_AUTO_CODE_FIX=true env var is set.
 * Runs every N digest cycles (default 2) to allow the agent time to fix before
 * the next cycle's changes are picked up by the self-watcher.
 */
async function handleDigestSealed({ entries, text }) {
  if (!devmodeAutoCodeFix.enabled) return;

  devmodeAutoCodeFix.cycleCount++;
  const cycle = devmodeAutoCodeFix.cycleCount;
  const interval = devmodeAutoCodeFix.cycleInterval;

  // Only run every N cycles
  if (cycle % interval !== 0) {
    console.log(
      `[devmode-fix] digest cycle ${cycle} — skipping (runs every ${interval} cycles)`,
    );
    return;
  }

  // Don't run if already running
  if (devmodeAutoCodeFix.running) {
    console.log(`[devmode-fix] skipping — previous run still in progress`);
    return;
  }

  // Filter to error and warning entries only
  const actionableEntries = entries.filter((e) => e.priority <= 3);
  if (actionableEntries.length === 0) {
    console.log(
      `[devmode-fix] digest cycle ${cycle} — no errors/warnings, skipping`,
    );
    return;
  }

  // Build the digest summary for the agent
  const errorLines = actionableEntries
    .map((e) => `${e.time} ${e.emoji} ${e.text}`)
    .join("\n");

  // Read recent orchestrator log tail for additional context
  let recentLogTail = "";
  try {
    const activeLogPath = resolve(logDir, "orchestrator-active.log");
    if (existsSync(activeLogPath)) {
      const logContent = readFileSync(activeLogPath, "utf8");
      const logLines = logContent.split("\n");
      recentLogTail = logLines.slice(-100).join("\n");
    }
  } catch {
    /* best effort */
  }

  // Determine target branch
  const targetBranch = devmodeAutoCodeFix.branch || "";
  const branchInstruction = targetBranch
    ? `Work on the existing branch: ${targetBranch}. Do NOT create a new branch.`
    : "Work on the current branch.";

  const prompt = [
    "You are a background debugging agent for the VirtEngine codex-monitor orchestration system.",
    "The following errors and warnings were collected from the most recent Live Digest cycle.",
    "Your job is to identify the root causes and fix them by modifying the source code.",
    "",
    "## IMPORTANT RULES",
    "- Do NOT commit your changes. Only modify files.",
    "- Do NOT run git push or create PRs.",
    `- ${branchInstruction}`,
    "- Focus on fixing the ROOT CAUSE, not symptoms.",
    "- The codex-monitor source code is in: scripts/codex-monitor/",
    "- The orchestrator is: scripts/codex-monitor/ve-orchestrator.ps1",
    "- The monitor is: scripts/codex-monitor/monitor.mjs",
    "- After you make changes, the monitor will auto-restart (file watcher).",
    "",
    "## Live Digest Errors/Warnings",
    "",
    errorLines,
    "",
    "## Recent Orchestrator Log (last 100 lines)",
    "",
    recentLogTail,
    "",
    "## Instructions",
    "",
    "1. Analyze the errors above to identify root causes",
    "2. Search the codebase for relevant code",
    "3. Make targeted fixes to resolve the issues",
    "4. Do NOT commit — the file watcher will auto-restart the monitor",
  ].join("\n");

  console.log(
    `[devmode-fix] digest cycle ${cycle} — dispatching background agent (${actionableEntries.length} actionable entries)`,
  );
  devmodeAutoCodeFix.running = true;
  devmodeAutoCodeFix.lastRunAt = Date.now();
  devmodeAutoCodeFix.lastDigestText = text;

  // Log the dispatch
  try {
    const fixLogDir = resolve(repoRoot, ".cache", "devmode-fix-logs");
    await mkdir(fixLogDir, { recursive: true });
    const stamp = new Date().toISOString().replace(/[:.]/g, "-");
    await writeFile(
      resolve(fixLogDir, `devmode-fix-${stamp}.prompt.md`),
      prompt,
      "utf8",
    );
  } catch {
    /* best effort */
  }

  // Dispatch to Codex SDK in background
  try {
    const result = await runCodexExec(
      prompt,
      repoRoot,
      devmodeAutoCodeFix.timeoutMs,
      resolve(repoRoot, ".cache", "devmode-fix-logs"),
    );

    if (result.success) {
      console.log(`[devmode-fix] background agent completed successfully`);
      void notify?.(
        `🔧 Devmode auto-fix completed (cycle ${cycle}). ${actionableEntries.length} issues addressed. Monitor will auto-restart.`,
        4,
        { dedupKey: "devmode-fix-complete" },
      );
    } else {
      console.warn(
        `[devmode-fix] background agent failed: ${result.error || "unknown"}`,
      );
      void notify?.(
        `⚠️ Devmode auto-fix failed (cycle ${cycle}): ${result.error || "no output"}`,
        3,
        { dedupKey: "devmode-fix-failed" },
      );
    }
  } catch (err) {
    console.error(`[devmode-fix] dispatch error: ${err.message}`);
  } finally {
    devmodeAutoCodeFix.running = false;
  }
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

  // ── Workspace monitor: initialize for this process session ──
  try {
    await workspaceMonitor.init();
  } catch (err) {
    console.warn(`[monitor] workspace monitor init failed: ${err.message}`);
  }

  // ── Agent log streaming: fan out per-attempt log lines to .cache/agent-logs/ ──
  const agentLogDir = resolve(repoRoot, ".cache", "agent-logs");
  try {
    await mkdir(agentLogDir, { recursive: true });
  } catch {
    /* best effort */
  }
  /** @type {Map<string, import('fs').WriteStream>} */
  const agentLogStreams = new Map();
  const AGENT_LOG_PATTERN =
    /\b([0-9a-f]{8})(?:-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})?\b/i;

  /**
   * Stream a log line to the per-attempt log file if it contains an attempt short ID.
   * @param {string} line - The log line
   */
  function streamToAgentLog(line) {
    const match = line.match(AGENT_LOG_PATTERN);
    if (!match) return;
    const shortId = match[1].toLowerCase();
    // Filter out common false positives (git SHAs in non-attempt context)
    if (
      line.includes("HEAD") ||
      line.includes("commit ") ||
      line.includes("Deleted branch")
    ) {
      return;
    }
    const logPath = resolve(agentLogDir, `${shortId}.log`);
    try {
      appendFileSync(logPath, `${line}\n`);
    } catch {
      /* best effort — non-critical */
    }
  }

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
        shellWriteRaw(chunk);
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
      // ── Agent log streaming: fan out to per-attempt log files ──
      streamToAgentLog(line);

      // ── Workspace monitoring: detect attempt lifecycle from orchestrator logs ──
      const trackMatch = line.match(
        /Tracking new attempt:\s+([0-9a-f]{8})\s*→\s*(\S+)/i,
      );
      if (trackMatch) {
        const shortId = trackMatch[1];
        const branch = trackMatch[2];
        const worktreePath = findWorktreeForBranch(branch);
        if (worktreePath) {
          void workspaceMonitor
            .startMonitoring(shortId, worktreePath, {
              taskId: shortId,
              executor: "unknown",
              branch,
            })
            .catch((err) =>
              console.warn(
                `[workspace-monitor] failed to start for ${shortId}: ${err.message}`,
              ),
            );
        }
      }

      if (isErrorLine(line, errorPatterns, errorNoisePatterns)) {
        lastErrorLine = line;
        lastErrorAt = Date.now();
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
        // Stop workspace monitoring for this attempt
        void workspaceMonitor
          .stopMonitoring(shortId, finishStatus)
          .catch(() => {});
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
  if (selfRestartTimer) {
    clearTimeout(selfRestartTimer);
    selfRestartTimer = null;
  }
}

function selfRestartForSourceChange(filename) {
  pendingSelfRestart = null;
  console.log(
    `\n[monitor] source files stable for ${Math.round(SELF_RESTART_QUIET_MS / 1000)}s — restarting (${filename})`,
  );
  console.log("[monitor] exiting for self-restart (fresh ESM modules)...");
  shuttingDown = true;
  if (vkLogStream) {
    vkLogStream.stop();
    vkLogStream = null;
  }
  // ── Stop internal executor + agent endpoint BEFORE exit ──
  // This ensures running agents get a chance to finish and ports are released.
  const shutdownPromises = [];
  if (internalTaskExecutor) {
    console.log("[monitor] stopping internal task executor for restart...");
    shutdownPromises.push(
      Promise.resolve(internalTaskExecutor.stop()).catch((e) =>
        console.warn(`[monitor] executor stop error: ${e.message}`),
      ),
    );
  }
  if (agentEndpoint) {
    console.log("[monitor] stopping agent endpoint for restart...");
    shutdownPromises.push(
      Promise.resolve(agentEndpoint.stop()).catch((e) =>
        console.warn(`[monitor] endpoint stop error: ${e.message}`),
      ),
    );
  }
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
  // Wait for executor/endpoint shutdown, then exit
  Promise.allSettled(shutdownPromises).then(() => {
    // Exit with special code — cli.mjs re-forks with fresh module cache
    setTimeout(() => process.exit(SELF_RESTART_EXIT_CODE), 500);
  });
  // Safety net: exit after 10s even if shutdown hangs
  setTimeout(() => process.exit(SELF_RESTART_EXIT_CODE), 10000);
}

function attemptSelfRestartAfterQuiet() {
  if (selfRestartTimer) {
    clearTimeout(selfRestartTimer);
    selfRestartTimer = null;
  }
  if (!selfRestartLastChangeAt) return;
  const now = Date.now();
  const sinceLastChange = now - selfRestartLastChangeAt;
  if (sinceLastChange < SELF_RESTART_QUIET_MS) {
    const waitMs = SELF_RESTART_QUIET_MS - sinceLastChange;
    selfRestartTimer = setTimeout(attemptSelfRestartAfterQuiet, waitMs);
    return;
  }
  const filename = selfRestartLastFile || "unknown";
  // Self-restart no longer blocked by primary agent — ephemeral pool
  // threads run independently and don't need to be drained.
  selfRestartForSourceChange(filename);
}

function queueSelfRestart(filename) {
  selfRestartLastChangeAt = Date.now();
  selfRestartLastFile = filename;
  if (selfRestartTimer) {
    clearTimeout(selfRestartTimer);
  }
  console.log(
    `\n[monitor] source file changed: ${filename} — waiting ${Math.round(SELF_RESTART_QUIET_MS / 1000)}s for quiet before restart`,
  );
  selfRestartTimer = setTimeout(
    attemptSelfRestartAfterQuiet,
    SELF_RESTART_QUIET_MS,
  );
}

function retryDeferredSelfRestart() {
  if (!pendingSelfRestart) return;
  attemptSelfRestartAfterQuiet();
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
        queueSelfRestart(filename);
      }, 1000);
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
  const prevPrimaryAgentName = primaryAgentName;
  const prevPrimaryAgentReady = primaryAgentReady;
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
  shellState.enabled = !!nextConfig.interactiveShellEnabled;
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

  // ── Internal executor hot-reload ──────────────────────────────────────
  if (nextConfig.internalExecutor) {
    internalExecutorConfig = nextConfig.internalExecutor;
  }

  codexEnabled = nextConfig.codexEnabled;
  primaryAgentName = nextConfig.primaryAgent;
  primaryAgentReady = nextConfig.primaryAgentEnabled;
  codexDisabledReason = codexEnabled
    ? ""
    : process.env.CODEX_SDK_DISABLED === "1"
      ? "disabled via CODEX_SDK_DISABLED"
      : agentSdk?.primary && agentSdk.primary !== "codex"
        ? `disabled via agent_sdk.primary=${agentSdk.primary}`
        : "disabled via --no-codex";

  const primaryAgentChanged = prevPrimaryAgentName !== primaryAgentName;
  if (primaryAgentChanged) {
    setPrimaryAgent(primaryAgentName);
  }
  if (
    (primaryAgentChanged && primaryAgentReady) ||
    (!prevPrimaryAgentReady && primaryAgentReady)
  ) {
    void initPrimaryAgent(primaryAgentName);
  }

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

  if (shellState.enabled && !shellState.active) {
    startInteractiveShell();
  } else if (!shellState.enabled && shellState.active && shellState.rl) {
    shellState.rl.close();
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
  if (vkLogStream) {
    vkLogStream.stop();
    vkLogStream = null;
  }
  if (internalTaskExecutor) {
    void internalTaskExecutor.stop();
    stopStatusFileWriter();
  }
  stopAutoUpdateLoop();
  stopSelfWatcher();
  stopEnvWatchers();
  if (watcher) {
    watcher.close();
  }
  if (currentChild) {
    currentChild.kill("SIGTERM");
  }
  void workspaceMonitor.shutdown();
  void releaseTelegramPollLock();
  process.exit(0);
});

// Windows: closing the terminal window doesn't send SIGINT/SIGTERM reliably.
process.on("exit", () => {
  shuttingDown = true;
  if (vkLogStream) {
    vkLogStream.stop();
    vkLogStream = null;
  }
  void workspaceMonitor.shutdown();
  void releaseTelegramPollLock();
});

process.on("SIGTERM", () => {
  shuttingDown = true;
  if (vkLogStream) {
    vkLogStream.stop();
    vkLogStream = null;
  }
  if (internalTaskExecutor) {
    void internalTaskExecutor.stop();
  }
  stopAutoUpdateLoop();
  stopSelfWatcher();
  stopEnvWatchers();
  if (watcher) {
    watcher.close();
  }
  if (currentChild) {
    currentChild.kill("SIGTERM");
  }
  void workspaceMonitor.shutdown();
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

// ── Startup sweep: kill stale processes, prune worktrees, archive old tasks ──
runMaintenanceSweep({
  repoRoot,
  archiveCompletedTasks: async () => {
    const projectId = await findVkProjectId();
    if (!projectId) return { archived: 0 };
    return await archiveCompletedTasks(fetchVk, projectId, { maxArchive: 50 });
  },
});

setInterval(() => {
  void flushErrorQueue();
}, 60 * 1000);

// ── Periodic maintenance: every 5 min, reap stuck pushes & prune worktrees ──
const maintenanceIntervalMs = 5 * 60 * 1000;
setInterval(() => {
  const childPid = currentChild ? currentChild.pid : undefined;
  runMaintenanceSweep({
    repoRoot,
    childPid,
    archiveCompletedTasks: async () => {
      const projectId = await findVkProjectId();
      if (!projectId) return { archived: 0 };
      return await archiveCompletedTasks(fetchVk, projectId, {
        maxArchive: 25,
        dryRun: false,
      });
    },
  });
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

// ── Fleet Coordination ───────────────────────────────────────────────────────
if (fleetConfig?.enabled) {
  const maxParallel = getMaxParallelFromArgs(scriptArgs) || 6;
  void initFleet({
    repoRoot,
    localSlots: maxParallel,
    ttlMs: fleetConfig.presenceTtlMs,
  })
    .then((state) => {
      console.log(
        `[fleet] ready: mode=${state.mode}, peers=${state.fleetSize}, totalSlots=${state.totalSlots}`,
      );
      void persistFleetState(repoRoot);
    })
    .catch((err) => {
      console.warn(`[fleet] init failed (continuing solo): ${err.message}`);
    });

  // Periodic fleet sync
  const syncMs = fleetConfig.syncIntervalMs || 2 * 60 * 1000;
  setInterval(() => {
    void refreshFleet({ ttlMs: fleetConfig.presenceTtlMs })
      .then(() => persistFleetState(repoRoot))
      .catch((err) => {
        console.warn(`[fleet] sync error: ${err.message}`);
      });
  }, syncMs);
  console.log(
    `[fleet] sync every ${Math.round(syncMs / 1000)}s, TTL=${Math.round((fleetConfig.presenceTtlMs || 300000) / 1000)}s`,
  );

  // Shared knowledge system
  if (fleetConfig.knowledgeEnabled) {
    initSharedKnowledge({
      repoRoot,
      targetFile: fleetConfig.knowledgeFile || "AGENTS.md",
    });
    console.log(
      `[fleet] shared knowledge enabled → ${fleetConfig.knowledgeFile || "AGENTS.md"}`,
    );
  }
} else {
  console.log("[fleet] disabled (set FLEET_ENABLED=true to enable)");
}

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
startInteractiveShell();
if (vkSpawnEnabled) {
  void ensureVibeKanbanRunning();
}
// When VK is externally managed (not spawned by monitor), still connect the
// log stream so agent logs are captured to .cache/agent-logs/.
if (!vkSpawnEnabled && vkEndpointUrl) {
  void isVibeKanbanOnline().then((online) => {
    if (online) ensureVkLogStream();
  });
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
// Periodically reconnect log stream for externally-managed VK (e.g. after VK restart).
// Session discovery is handled by ensureVkSessionDiscoveryLoop() inside ensureVkLogStream().
if (
  !vkSpawnEnabled &&
  vkEndpointUrl &&
  Number.isFinite(vkEnsureIntervalMs) &&
  vkEnsureIntervalMs > 0
) {
  setInterval(() => {
    if (!vkLogStream) {
      void isVibeKanbanOnline().then((online) => {
        if (online) ensureVkLogStream();
      });
    }
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

// ── Log complexity routing matrix at startup ──────────────────────────────────
try {
  const complexityMatrix = getComplexityMatrix(config.complexityRouting);
  const matrixLines = [];
  for (const [exec, tiers] of Object.entries(complexityMatrix)) {
    for (const [tier, profile] of Object.entries(tiers)) {
      matrixLines.push(
        `  ${exec}/${tier}: ${profile.model || "default"} (${profile.reasoningEffort || "default"})`,
      );
    }
  }
  console.log(
    `[monitor] complexity routing matrix:\n${matrixLines.join("\n")}`,
  );
} catch (err) {
  console.warn(`[monitor] complexity matrix log failed: ${err.message}`);
}

// ── Clean stale status data on startup ─────────────────────────────────────
try {
  const statusRaw = existsSync(statusPath)
    ? readFileSync(statusPath, "utf8")
    : null;
  if (statusRaw) {
    const statusData = JSON.parse(statusRaw);
    const attempts = statusData.attempts || {};
    const STALE_AGE_MS = 12 * 60 * 60 * 1000; // 12 hours
    const now = Date.now();
    let cleaned = 0;
    for (const [key, attempt] of Object.entries(attempts)) {
      const ts = attempt?.updated_at || attempt?.created_at;
      if (!ts) continue;
      const age = now - Date.parse(ts);
      if (age > STALE_AGE_MS && attempt.status === "running") {
        // Mark stale running attempts as "stale" so they don't show as active
        attempt.status = "stale";
        attempt._stale_reason = `No update for ${Math.round(age / 3600000)}h — marked stale on startup`;
        cleaned++;
      }
    }
    if (cleaned > 0) {
      statusData.updated_at = new Date().toISOString();
      writeFileSync(statusPath, JSON.stringify(statusData, null, 2), "utf8");
      console.log(
        `[monitor] cleaned ${cleaned} stale attempts from status file`,
      );
    }
  }
} catch (err) {
  console.warn(`[monitor] stale cleanup failed: ${err.message}`);
}

// ── Internal Executor / VK Orchestrator startup ──────────────────────────────
/** @type {import("./task-executor.mjs").TaskExecutor|null} */
let internalTaskExecutor = null;
/** @type {import("./agent-endpoint.mjs").AgentEndpoint|null} */
let agentEndpoint = null;
/** @type {import("./review-agent.mjs").ReviewAgent|null} */
let reviewAgent = null;
/** @type {import("./sync-engine.mjs").SyncEngine|null} */
let syncEngine = null;
/** @type {import("./error-detector.mjs").ErrorDetector|null} */
let errorDetector = null;

// ── Task Management Subsystem Initialization ────────────────────────────────
try {
  loadTaskStore();
  console.log("[monitor] internal task store loaded");
} catch (err) {
  console.warn(`[monitor] task store init warning: ${err.message}`);
}

// Error detector
try {
  errorDetector = createErrorDetector({
    sendTelegram:
      telegramToken && telegramChatId
        ? (msg) => void sendTelegramMessage(msg)
        : null,
  });
  console.log("[monitor] error detector initialized");
} catch (err) {
  console.warn(`[monitor] error detector init failed: ${err.message}`);
}

const executorMode = configExecutorMode || getExecutorMode();
console.log(`[monitor] executor mode: ${executorMode}`);

if (executorMode === "internal" || executorMode === "hybrid") {
  // Start internal executor
  try {
    const execOpts = {
      ...internalExecutorConfig,
      repoRoot,
      repoSlug,
      sendTelegram:
        telegramToken && telegramChatId
          ? (msg) => void sendTelegramMessage(msg)
          : null,
      onTaskStarted: (task, slot) => {
        console.log(
          `[task-executor] 🚀 started: "${task.title}" (${slot.sdk}) in ${slot.worktreePath}`,
        );
      },
      onTaskCompleted: (task, result) => {
        console.log(
          `[task-executor] ✅ completed: "${task.title}" (${result.attempts} attempt(s))`,
        );
        // Queue review if task moved to inreview and has a PR
        if (reviewAgent && result.success) {
          try {
            reviewAgent.queueReview({
              id: task.id || task.task_id,
              title: task.title,
              branchName: task.branchName || task.meta?.branch_name,
              description: task.description || "",
            });
          } catch {
            /* best-effort */
          }
        }
      },
      onTaskFailed: (task, err) => {
        console.warn(
          `[task-executor] ❌ failed: "${task.title}" — ${err?.message || err}`,
        );
      },
    };
    internalTaskExecutor = getTaskExecutor(execOpts);
    internalTaskExecutor.start();

    // Write executor slots to status file every 30s for Telegram /tasks
    startStatusFileWriter(30000);
    console.log(
      `[monitor] internal executor started (maxParallel=${execOpts.maxParallel || 3}, sdk=${execOpts.sdk || "auto"})`,
    );

    // ── Agent Endpoint ──
    try {
      agentEndpoint = createAgentEndpoint({
        port: Number(process.env.AGENT_ENDPOINT_PORT || 18432),
        onTaskComplete: (taskId, body) => {
          console.log(`[monitor] agent self-reported complete for ${taskId}`);
          try {
            setInternalTaskStatus(taskId, "inreview", "agent-endpoint");
          } catch {
            /* best-effort */
          }
          if (reviewAgent) {
            const task = internalTaskExecutor?._activeSlots?.get(taskId);
            if (task)
              reviewAgent.queueReview({
                id: taskId,
                title: task.taskTitle,
                prNumber: body?.prNumber,
                branchName: task.branch,
                description: body?.description || "",
              });
          }
        },
        onTaskError: (taskId, body) => {
          console.warn(
            `[monitor] agent self-reported error for ${taskId}: ${body?.error}`,
          );
          if (errorDetector) {
            const classification = errorDetector.classify(
              body?.output || "",
              body?.error || "",
            );
            errorDetector.recordError(taskId, classification);
          }
        },
        onStatusChange: (taskId, newStatus) => {
          console.log(
            `[monitor] agent status change for ${taskId}: ${newStatus}`,
          );
          try {
            setInternalTaskStatus(taskId, newStatus, "agent-endpoint");
          } catch {
            /* best-effort */
          }
        },
      });
      agentEndpoint
        .start()
        .then(() => {
          console.log("[monitor] agent endpoint started");
        })
        .catch((err) => {
          console.warn(
            `[monitor] agent endpoint failed to start: ${err.message}`,
          );
          agentEndpoint = null;
        });
    } catch (err) {
      console.warn(`[monitor] agent endpoint creation failed: ${err.message}`);
      agentEndpoint = null;
    }

    // ── Review Agent ──
    try {
      reviewAgent = createReviewAgent({
        maxConcurrent: 2,
        sendTelegram:
          telegramToken && telegramChatId
            ? (msg) => void sendTelegramMessage(msg)
            : null,
        onReviewComplete: (taskId, result) => {
          console.log(
            `[monitor] review complete for ${taskId}: ${result?.verdict || "unknown"}`,
          );
          try {
            setReviewResult(taskId, {
              approved: result?.verdict === "approved",
              issues: result?.issues || [],
            });
          } catch {
            /* best-effort */
          }

          if (result?.verdict === "approved") {
            // Auto-mark as done since reviewer is happy
            console.log(
              `[monitor] review approved — marking ${taskId} as done`,
            );
            try {
              setInternalTaskStatus(taskId, "done", "review-agent");
            } catch {
              /* best-effort */
            }
            try {
              updateTaskStatus(taskId, "done");
            } catch {
              /* best-effort */
            }
          } else {
            console.log(
              `[monitor] review found issues for ${taskId} — task stays inreview`,
            );
          }
        },
      });
      reviewAgent.start();
      console.log("[monitor] review agent started");
    } catch (err) {
      console.warn(`[monitor] review agent failed to start: ${err.message}`);
    }

    // ── Sync Engine ──
    try {
      const projectId =
        process.env.INTERNAL_EXECUTOR_PROJECT_ID ||
        process.env.VK_PROJECT_ID ||
        internalExecutorConfig?.projectId;
      if (projectId) {
        syncEngine = createSyncEngine({
          projectId,
          syncIntervalMs: 60_000, // 1 minute
          sendTelegram:
            telegramToken && telegramChatId
              ? (msg) => void sendTelegramMessage(msg)
              : null,
        });
        syncEngine.start();
        console.log("[monitor] sync engine started (interval: 60s)");
      } else {
        console.log("[monitor] sync engine skipped — no project ID configured");
      }
    } catch (err) {
      console.warn(`[monitor] sync engine failed to start: ${err.message}`);
    }
  } catch (err) {
    console.error(
      `[monitor] internal executor failed to start: ${err.message}`,
    );
  }
}

if (executorMode === "vk" || executorMode === "hybrid") {
  // Start VK orchestrator (ve-orchestrator.ps1)
  startProcess();
} else {
  console.log("[monitor] VK orchestrator skipped (executor mode = internal)");
}
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
  onDigestSealed: devmodeAutoCodeFix.enabled ? handleDigestSealed : null,
  getAnomalyReport: () =>
    anomalyDetector
      ? anomalyDetector.getStatusReport()
      : "Anomaly detector not running.",
  getInternalExecutor: () => internalTaskExecutor,
  getExecutorMode: () => executorMode,
  getAgentEndpoint: () => agentEndpoint,
  getReviewAgent: () => reviewAgent,
  getSyncEngine: () => syncEngine,
  getErrorDetector: () => errorDetector,
  getTaskStoreStats: () => {
    try {
      return getTaskStoreStats();
    } catch {
      return null;
    }
  },
});
if (telegramBotEnabled) {
  void startTelegramBot();
}

// ── Start PR Cleanup Daemon ──────────────────────────────────────────────────
// Automatically resolves PR conflicts and CI failures every 30 minutes
let prCleanupDaemon = null;
if (config.prCleanupEnabled !== false) {
  console.log("[monitor] Starting PR cleanup daemon...");
  prCleanupDaemon = new PRCleanupDaemon({
    intervalMs: 30 * 60 * 1000, // 30 minutes
    maxConcurrentCleanups: 3,
    dryRun: false,
    autoMerge: true,
  });
  prCleanupDaemon.start();
}

// ── Named exports for testing ───────────────────────────────────────────────
export {
  fetchVk,
  updateTaskStatus,
  safeRecoverTask,
  recoverySkipCache,
  getTaskAgeMs,
  classifyConflictedFiles,
  AUTO_RESOLVE_THEIRS,
  AUTO_RESOLVE_OURS,
  AUTO_RESOLVE_LOCK_EXTENSIONS,
  extractScopeFromTitle,
  resolveUpstreamFromConfig,
  rebaseDownstreamTasks,
  runTaskAssessment,
  // Internal executor
  internalTaskExecutor,
  // Task management subsystems
  agentEndpoint,
  reviewAgent,
  syncEngine,
  errorDetector,
  // Fleet coordination re-exports for external consumers
  getFleetState,
  isFleetCoordinator,
  getFleetMode,
  formatFleetSummary,
  buildExecutionWaves,
  calculateBacklogDepth,
  detectMaintenanceMode,
  appendKnowledgeEntry,
  buildKnowledgeEntry,
  formatKnowledgeSummary,
};
