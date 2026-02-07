import { execSync, spawn, spawnSync } from "node:child_process";
import { existsSync, watch, statSync } from "node:fs";
import {
  copyFile,
  mkdir,
  readFile,
  rename,
  unlink,
  writeFile,
} from "node:fs/promises";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { acquireMonitorLock, runMaintenanceSweep } from "./maintenance.mjs";
import { attemptAutoFix, fixLoopingError } from "./autofix.mjs";
import {
  startTelegramBot,
  stopTelegramBot,
  injectMonitorFunctions,
  bumpAgentMessage,
  isAgentActive,
} from "./telegram-bot.mjs";
import { execCodexPrompt, isCodexBusy } from "./codex-shell.mjs";

const __dirname = resolve(fileURLToPath(new URL(".", import.meta.url)));

const defaultScript = resolve(__dirname, "..", "ve-orchestrator.ps1");
const args = process.argv.slice(2);

function getArg(name, fallback) {
  const idx = args.indexOf(name);
  if (idx === -1 || idx === args.length - 1) {
    return fallback;
  }
  return args[idx + 1];
}

function getFlag(name) {
  return args.includes(name);
}

const scriptPath = resolve(getArg("--script", defaultScript));
  const scriptArgsRaw = getArg("--args", "-MaxParallel 6 -SkipSecurityChecks");
const scriptArgs = scriptArgsRaw.split(" ").filter(Boolean);
const restartDelayMs = Number(getArg("--restart-delay", "10000"));
const maxRestarts = Number(getArg("--max-restarts", "0"));
const logDir = resolve(getArg("--log-dir", resolve(__dirname, "logs")));
const watchEnabled = !getFlag("--no-watch");
const watchPath = resolve(getArg("--watch-path", scriptPath));
const echoLogs = !getFlag("--no-echo-logs");
const autoFixEnabled = !getFlag("--no-autofix");
const vkEnsureIntervalMs = Number(getArg("--vk-ensure-interval", "60000"));
let codexEnabled =
  !getFlag("--no-codex") && process.env.CODEX_SDK_DISABLED !== "1";
const repoRoot = resolve(__dirname, "..", "..");
const statusPath = resolve(repoRoot, ".cache", "ve-orchestrator-status.json");
const telegramToken = process.env.TELEGRAM_BOT_TOKEN;
const telegramChatId = process.env.TELEGRAM_CHAT_ID;
const telegramIntervalMin = Number(process.env.TELEGRAM_INTERVAL_MIN || "10");
const repoSlug = process.env.GITHUB_REPO || "virtengine/virtengine";
const repoUrlBase =
  process.env.GITHUB_REPO_URL || `https://github.com/${repoSlug}`;
const vkRecoveryPort = process.env.VK_RECOVERY_PORT || "54089";
const vkRecoveryHost =
  process.env.VK_RECOVERY_HOST || process.env.VK_HOST || "0.0.0.0";
const vkEndpointUrl =
  process.env.VK_ENDPOINT_URL ||
  process.env.VK_BASE_URL ||
  `http://127.0.0.1:${vkRecoveryPort}`;
const vkPublicUrl = process.env.VK_PUBLIC_URL || process.env.VK_WEB_URL || "";
const vkRecoveryCooldownMin = Number(
  process.env.VK_RECOVERY_COOLDOWN_MIN || "10",
);

let CodexClient = null;
let codexDisabledReason = "";

if (!codexEnabled) {
  codexDisabledReason =
    process.env.CODEX_SDK_DISABLED === "1"
      ? "disabled via CODEX_SDK_DISABLED"
      : "disabled via --no-codex";
}

let restartCount = 0;

// ── Codex Shell adapter for autofix ─────────────────────────────────────────
// Wraps execCodexPrompt (persistent SDK thread with tools) into the
// {success, output, error} interface that autofix.mjs expects.
async function execViaCodexShell(prompt, _cwd) {
  if (isCodexBusy()) {
    return {
      success: false,
      output: "",
      error: "Codex Shell busy — another turn is in flight",
    };
  }
  try {
    const res = await execCodexPrompt(prompt, { timeoutMs: 5 * 60 * 1000 });
    return {
      success: true,
      output: res.finalResponse || "",
      error: null,
    };
  } catch (err) {
    return {
      success: false,
      output: "",
      error: err.message || String(err),
    };
  }
}

// Returns the adapter when Codex Shell is available, null otherwise
function getExecViaShell() {
  if (!codexEnabled) return null;
  return execViaCodexShell;
}
let shuttingDown = false;
let currentChild = null;
let pendingRestart = false;
let skipNextAnalyze = false;
let skipNextRestartCount = false;
let gracefulKill = false;
let watcher = null;
let watcherDebounce = null;
let watchFileName = null;
let watchFileMtimeMs = null;
let vkRecoveryLastAt = 0;
let vibeKanbanProcess = null;
let vibeKanbanStartedAt = 0;
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

function isOrchestratorProcessRunning() {
  try {
    if (process.platform === "win32") {
      const cmd =
        "Get-CimInstance Win32_Process -Filter \"Name='pwsh.exe'\" | Select-Object -ExpandProperty CommandLine";
      const out = execSync(
        `powershell -NoProfile -Command ${JSON.stringify(cmd)}`,
        {
          encoding: "utf8",
          timeout: 10000,
          stdio: ["ignore", "pipe", "ignore"],
        },
      );
      return out
        .split(/\r?\n/)
        .some((line) => line.toLowerCase().includes("ve-orchestrator.ps1"));
    }

    const out = execSync("ps -eo args 2>/dev/null", {
      encoding: "utf8",
      timeout: 10000,
    });
    return out
      .split(/\r?\n/)
      .some((line) => line.toLowerCase().includes("ve-orchestrator.ps1"));
  } catch {
    return false;
  }
}

let logRemainder = "";
const mergeNotified = new Set();
const pendingMerges = new Set();
const errorNotified = new Map();
const mergeFailureNotified = new Map();
const vkErrorNotified = new Map();
const telegramDedup = new Map();
let allCompleteNotified = false;
let backlogLowNotified = false;
let plannerTriggered = false;

// ── Per-task failure tracking for auto-reattempt ────────────────────────────
// Tracks individual task failures. If a task hits 3 singular failures
// (not caused by shared infrastructure issues), it gets auto-reattempted.
const taskFailureCounts = new Map(); // taskId → { count, lastFailedAt, lastBranch }
const REATTEMPT_THRESHOLD = 3;
const SHARED_FAILURE_WINDOW_MS = 60 * 1000; // If 3+ tasks fail within 60s, it's shared
const sharedFailureTimestamps = []; // timestamps of all task failures

// ── Telegram history ring buffer ────────────────────────────────────────────
// Stores the last N sent messages for context enrichment (fed to autofix prompts)
const TELEGRAM_HISTORY_MAX = 25;
const telegramHistory = [];

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

function runCodexExec(prompt, cwd, timeoutMs = 120_000) {
  return new Promise((resolve) => {
    // On Windows, 'codex' is a .cmd shim — shell: false can't find it.
    // Use cmd /c to preserve argument boundaries without shell arg-splitting.
    const isWin = process.platform === "win32";
    const codexArgs = ["exec", "--full-auto", "-C", cwd, prompt];
    const spawnCmd = isWin ? "cmd" : "codex";
    const spawnArgs = isWin ? ["/c", "codex", ...codexArgs] : codexArgs;

    const child = spawn(spawnCmd, spawnArgs, {
      cwd,
      stdio: ["ignore", "pipe", "pipe"],
      shell: false,
      timeout: timeoutMs,
      env: { ...process.env },
    });

    let stdout = "";
    let stderr = "";

    child.stdout.on("data", (chunk) => {
      stdout += chunk.toString();
    });
    child.stderr.on("data", (chunk) => {
      stderr += chunk.toString();
    });

    const timer = setTimeout(() => {
      try {
        child.kill("SIGTERM");
      } catch {
        /* best effort */
      }
      resolve({
        success: false,
        output: stdout,
        error: `timeout after ${timeoutMs}ms`,
      });
    }, timeoutMs);

    child.on("error", (err) => {
      clearTimeout(timer);
      resolve({
        success: false,
        output: stdout,
        error: err.message,
      });
    });

    child.on("exit", (code) => {
      clearTimeout(timer);
      resolve({
        success: code === 0,
        output: stdout + (stderr ? `\n${stderr}` : ""),
        error: code !== 0 ? `exit code ${code}` : null,
      });
    });
  });
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
  const prompt = `You are debugging the VirtEngine codex-monitor.

The monitor process hit an unexpected exception and needs a fix.
Please inspect and fix code in:
- scripts/codex-monitor/monitor.mjs
- scripts/codex-monitor/autofix.mjs
- scripts/codex-monitor/maintenance.mjs

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
  const runFix = getExecViaShell() || ((p, cwd) => runCodexExec(p, cwd));
  const result = await runFix(prompt, repoRoot);
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

async function handleMonitorFailure(reason, err) {
  if (monitorFailureHandling) return;
  monitorFailureHandling = true;
  const failureCount = recordMonitorFailure();
  const message = err && err.message ? err.message : String(err || reason);

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
      void sendTelegramMessage(
        `⚠️ codex-monitor exception (${reason}). Attempting recovery (count=${failureCount}).`,
      );
    }

    const fixResult = await attemptMonitorFix({
      error: err || new Error(reason),
      logText: logRemainder,
    });

    if (fixResult.fixed) {
      if (telegramToken && telegramChatId) {
        void sendTelegramMessage(
          `🛠️ codex-monitor auto-fix applied. Restarting monitor.\n${fixResult.outcome}`,
        );
      }
      restartSelf("monitor-fix-applied");
      return;
    }

    if (shouldRestartMonitor()) {
      monitorSafeModeUntil = Date.now() + orchestratorPauseMs;
      const pauseMin = Math.round(orchestratorPauseMs / 60000);
      if (telegramToken && telegramChatId) {
        void sendTelegramMessage(
          `🛑 codex-monitor entering safe mode after repeated failures (${failureCount} in 10m). Pausing restarts for ${pauseMin} minutes.`,
        );
      }
      return;
    }
  } catch (fatal) {
    console.warn(
      `[monitor] failure handler crashed: ${fatal.message || fatal}`,
    );
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
  const prompt = `You are a reliability engineer debugging a crash loop in VirtEngine automation.

The orchestrator is restarting repeatedly within minutes.
Please diagnose the likely root cause and apply a minimal fix.

Targets (edit only if needed):
- scripts/ve-orchestrator.ps1
- scripts/codex-monitor/monitor.mjs
- scripts/codex-monitor/autofix.mjs
- scripts/codex-monitor/maintenance.mjs

Recent log excerpt:
${logText.slice(-6000)}

Constraints:
1) Prevent rapid restart loops (introduce backoff or safe-mode).
2) Keep behavior stable and production-safe.
3) Do not refactor unrelated code.
4) Prefer small guardrails over big rewrites.`;

  const filesBefore = detectChangedFiles(repoRoot);
  const runFix =
    getExecViaShell() || ((p, cwd) => runCodexExec(p, cwd, 180_000));
  const result = await runFix(prompt, repoRoot);
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

function getErrorFingerprint(line) {
  // Normalize: strip timestamps, attempt IDs, branch-specific parts
  return line
    .replace(/\[\d{2}:\d{2}:\d{2}\]\s*/g, "")
    .replace(/\b[0-9a-f]{8}\b/gi, "<ID>") // attempt IDs
    .replace(/ve\/[\w.-]+/g, "ve/<BRANCH>") // branch names
    .trim();
}

/** Known Windows crash/exception exit codes that are NOT orchestrator bugs. */
const WINDOWS_CRASH_CODES = new Set([
  1073807364, // 0x40010004 — STATUS_LOG_HARD_ERROR (Windows dialog crash)
  3221226091, // 0xC000027B — STATUS_STOWED_EXCEPTION (COM/WinRT stowed exception)
  3221225477, // 0xC0000005 — STATUS_ACCESS_VIOLATION
  3221225725, // 0xC00000FD — STATUS_STACK_OVERFLOW
  3221225786, // 0xC000013A — STATUS_CONTROL_C_EXIT
]);

function trackErrorFrequency(line) {
  const fingerprint = getErrorFingerprint(line);
  if (!fingerprint) return;

  // Skip stats/metrics lines that contain error keywords but aren't actual errors.
  // These slip through isErrorLine() via generic patterns like /\bFailed\b/i.
  if (/First-shot:.*Failed:/i.test(line)) return;
  if (/Status:\s*running=\d/i.test(line)) return;

  // Skip orchestrator internal state tracking lines — these are informational
  // and should not trigger loop detection even if they contain error-like words.
  if (/Tracking new attempt:/i.test(line)) return;
  if (/Attempt\s+[0-9a-f]+\s+(finished|stale-running)/i.test(line)) return;
  if (/marking review|requires agent attention/i.test(line)) return;

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
      execViaShell: getExecViaShell(),
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
  "prompt token count",
  "token limit",
  "maximum context length",
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
  // Orchestrator status/stats lines — NOT real errors
  /First-shot:\s+\d+(\.\d+)?%.*\|\s*Fix:.*\|\s*Failed:/i, // success rate summary
  /Status:\s+running=\d+,\s*review=\d+,\s*error=\d+/i, // status counts
  /^\s*[│║╔╚└──┌┐┘┤├]+/, // box-drawing / banner lines
  /^\s*──\s*Cycle\s+\d+/, // cycle separator
  /ALL TASKS COMPLETE/i, // completion banner (normal)
  /^\s*Sleeping\s+\d+s\s+until\s+next\s+cycle/i, // sleep notice
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+Attempt\s+[0-9a-f]+\s+stale-running/i, // stale-running log (handled internally)
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+Deferring.*no available slots/i, // slot deferral (expected)
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+Health:\s+reported\s+/i, // executor health reports (handled by health system)
  /Reconnecting\.\.\.\s*\d+\/\d+/i, // reconnect loops (handled by degradation detector)
  /Tracking new attempt:\s+[0-9a-f]+/i, // attempt tracking after VK restart — NOT a real error
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+Tracking new attempt:/i, // timestamped variant
  /Attempt\s+[0-9a-f]+\s+finished\s*\(/i, // attempt finish status (internal orchestrator state)
  /marking review/i, // internal state transition, not an error
  /requires agent attention/i, // orchestrator info log, not monitor-actionable
  /^\s*\[\d{2}:\d{2}:\d{2}\]\s+Attempt\s+[0-9a-f]+\s+(failed|killed|timeout)\s+in workspace/i, // workspace exit status (handled by orchestrator retry logic)
  /follow-up failed.*An internal error occurred/i, // VK API internal errors on follow-up (transient, not executor degradation)
  /Follow-up failed for/i, // orchestrator follow-up retry log (handled internally)
];

const vkErrorPatterns = [
  /Failed to initialize vibe-kanban configuration/i,
  /HTTP GET http:\/\/127\.0\.0\.1:54089\/api\/projects failed/i,
];

// ─── Executor Degradation Patterns ───────────────────────────────────────────
// These detect signs that an executor (Codex/Copilot) is degraded, timing out,
// or being rate-limited — triggering failover notifications.

const degradationPatterns = [
  {
    pattern: /Reconnecting\.\.\.\s*(\d+)\/(\d+)/i,
    type: "reconnect_loop",
    severity: (match) => {
      const current = parseInt(match[1], 10);
      return current >= 5 ? "critical" : "warning";
    },
    label: "Reconnection loop",
  },
  {
    pattern: /oops.?you can.?t create more requests/i,
    type: "rate_limit",
    severity: () => "critical",
    label: "Copilot rate limit hit",
  },
  {
    pattern: /rate.?limit|too many requests|\b429\b|quota exceeded/i,
    type: "rate_limit",
    severity: () => "critical",
    label: "API rate limit",
  },
  {
    pattern: /hard timeout|idle timeout|operation timed out|deadline exceeded/i,
    type: "timeout",
    severity: () => "critical",
    label: "Executor timeout",
  },
  {
    pattern: /connection reset|ECONNRESET|ECONNREFUSED|ETIMEDOUT/i,
    type: "connection_error",
    severity: () => "warning",
    label: "Connection error",
  },
  {
    pattern: /capacity|server overloaded|503|502/i,
    type: "capacity",
    severity: () => "warning",
    label: "Executor capacity issue",
  },
];

/** Track recent degradation events for cooldown deduplication */
const degradationNotified = new Map();

function checkDegradationPatterns(line) {
  for (const deg of degradationPatterns) {
    const match = line.match(deg.pattern);
    if (match) {
      const severity = deg.severity(match);
      const key = `${deg.type}:${deg.label}`;
      const now = Date.now();
      const last = degradationNotified.get(key) || 0;
      // Deduplicate: critical every 5min, warning every 15min
      const cooldown = severity === "critical" ? 5 * 60 * 1000 : 15 * 60 * 1000;
      if (now - last < cooldown) continue;
      degradationNotified.set(key, now);

      const icon = severity === "critical" ? "\u{1F6A8}" : "\u26A0\uFE0F";
      const message =
        `${icon} <b>Executor Degradation</b>\n` +
        `Type: ${escapeHtml(deg.label)}\n` +
        `Pattern: <code>${escapeHtml(match[0].substring(0, 100))}</code>\n` +
        `Severity: ${severity}\n` +
        `Action: Health system will auto-failover to backup executors`;
      queueErrorMessage(message);
      return { type: deg.type, severity, label: deg.label, match: match[0] };
    }
  }
  return null;
}

// ─── Region Failover Tracking ────────────────────────────────────────────────
// Tracks sustained critical degradation events. If 3+ critical events happen
// within 10 minutes, triggers automatic region failover via PowerShell.

const criticalEvents = [];
const CRITICAL_THRESHOLD = 3;
const CRITICAL_WINDOW_MS = 10 * 60 * 1000; // 10 minutes
let lastRegionFailoverAt = 0;
const REGION_FAILOVER_COOLDOWN_MS = 30 * 60 * 1000; // 30 min between failovers

function trackRegionFailover(degradation) {
  if (!degradation || degradation.severity !== "critical") return;

  const now = Date.now();
  criticalEvents.push({ at: now, type: degradation.type });

  // Prune old events
  while (
    criticalEvents.length > 0 &&
    now - criticalEvents[0].at > CRITICAL_WINDOW_MS
  ) {
    criticalEvents.shift();
  }

  // Check if threshold reached
  if (criticalEvents.length >= CRITICAL_THRESHOLD) {
    if (now - lastRegionFailoverAt < REGION_FAILOVER_COOLDOWN_MS) {
      console.log(
        "[monitor] Region failover cooldown active, skipping auto-switch",
      );
      return;
    }

    lastRegionFailoverAt = now;
    criticalEvents.length = 0; // Reset after triggering

    // Trigger region switch via orchestrator's PowerShell functions
    const isWin = process.platform === "win32";
    const pwsh = isWin ? "powershell.exe" : "pwsh";
    const script = resolve(repoRoot, "scripts", "ve-kanban.ps1");
    try {
      const result = execSync(
        `${pwsh} -NoProfile -Command ". '${script}'; Initialize-CodexRegionTracking; $r = Switch-CodexRegion -Region 'sweden'; Write-Host ($r | ConvertTo-Json -Compress)"`,
        { cwd: repoRoot, encoding: "utf8", timeout: 15000 },
      );
      console.log(`[monitor] Region failover triggered: ${result.trim()}`);
      const message =
        `🌍 <b>Auto Region Failover</b>\n` +
        `${CRITICAL_THRESHOLD} critical degradation events in ${Math.round(CRITICAL_WINDOW_MS / 60000)}min\n` +
        `Switching Codex: US → Sweden\n` +
        `Auto-restore to US after cooldown (120min)\n\n` +
        `Use /region us to switch back manually`;
      queueErrorMessage(message);
    } catch (err) {
      console.error(`[monitor] Region failover failed: ${err.message}`);
    }
  }
}

function escapeHtml(value) {
  return String(value)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
}

function formatHtmlLink(url, label) {
  if (!url) {
    return escapeHtml(label);
  }
  return `<a href="${escapeHtml(url)}">${escapeHtml(label)}</a>`;
}

function isErrorLine(line) {
  if (errorNoisePatterns.some((pattern) => pattern.test(line))) {
    return false;
  }
  return errorPatterns.some((pattern) => pattern.test(line));
}

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
    "VirtEngine Orchestrator Warning",
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
    await writeFile(outPath, String(result), "utf8");
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
let vkExternallyManaged = false; // true when VK is running outside this monitor
let vkRestartScheduled = false; // prevent overlapping scheduled restarts

// Determine correct npx binary for the platform (avoids shell:true DEP0190)
const npxBin = process.platform === "win32" ? "npx.cmd" : "npx";

async function startVibeKanbanProcess() {
  if (vibeKanbanProcess && !vibeKanbanProcess.killed) {
    return;
  }

  // CRITICAL: Check if VK is already reachable before spawning a new instance.
  // This prevents the restart-death-loop when an external VK is running on the port.
  if (await isVibeKanbanOnline()) {
    console.log(
      `[monitor] vibe-kanban already reachable at ${vkEndpointUrl} — adopting external instance`,
    );
    vkExternallyManaged = true;
    vkRestartCount = 0;
    vkRestartScheduled = false;
    return;
  }

  vkExternallyManaged = false;

  const env = {
    ...process.env,
    PORT: vkRecoveryPort,
    HOST: vkRecoveryHost,
  };

  console.log(
    `[monitor] starting vibe-kanban via npx (HOST=${vkRecoveryHost} PORT=${vkRecoveryPort}, endpoint=${vkEndpointUrl})`,
  );

  vibeKanbanProcess = spawn(npxBin, ["--yes", "vibe-kanban"], {
    env,
    cwd: repoRoot,
    stdio: "ignore",
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
  if (vkRestartScheduled) return; // prevent overlapping restarts
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
  vkRestartScheduled = true;
  setTimeout(async () => {
    vkRestartScheduled = false;
    await startVibeKanbanProcess();
  }, delay);
}

async function isVibeKanbanOnline() {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 2000);
  try {
    const res = await fetch(`${vkEndpointUrl}/api/projects`, {
      signal: controller.signal,
    });
    return res.ok;
  } catch {
    return false;
  } finally {
    clearTimeout(timeout);
  }
}

async function ensureVibeKanbanRunning() {
  if (await isVibeKanbanOnline()) {
    // VK is responding — reset counters and track as externally managed if needed
    if (!vibeKanbanProcess || vibeKanbanProcess.killed) {
      // VK is online but we didn't spawn it — adopt it
      if (!vkExternallyManaged) {
        console.log(
          "[monitor] vibe-kanban is online (external instance) — skipping spawn",
        );
        vkExternallyManaged = true;
      }
    }
    vkRestartCount = 0;
    vkRestartScheduled = false;
    return;
  }
  // VK is offline — if a restart is already scheduled, let it handle things
  if (vkRestartScheduled) {
    return;
  }
  vkExternallyManaged = false;
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

const errorQueue = [];
const contextOverflowHandled = new Map();
const contextOverflowCooldownMs = 30 * 60 * 1000;
const contextOverflowPatterns = [
  /prompt token count/i,
  /ContextWindowExceeded/i,
  /context window/i,
  /ran out of room/i,
  /too many tokens/i,
  /token limit/i,
  /exceeds the limit/i,
  /maximum context length/i,
];

function queueErrorMessage(line) {
  errorQueue.push(line);
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
  const message = ["VirtEngine Orchestrator Error", ...lines].join("\n");
  await sendTelegramMessage(message);
}

function isContextOverflowLine(line) {
  return contextOverflowPatterns.some((pattern) => pattern.test(line));
}

function getContextOverflowKey(line) {
  return line
    .replace(/\d{4,}/g, "<N>")
    .replace(/\b[0-9a-f]{8,}\b/gi, "<ID>")
    .slice(0, 160);
}

function pickAttemptForContextOverflow(status) {
  if (!status || !status.attempts) return null;
  const attempts = Object.values(status.attempts)
    .filter(Boolean)
    .filter((a) => a.task_id);
  if (attempts.length === 0) return null;

  const candidates = attempts.filter((a) => {
    const lp = (a.last_process_status || "").toLowerCase();
    const st = (a.status || "").toLowerCase();
    return (
      ["failed", "timeout", "killed", "stopped"].includes(lp) ||
      ["review", "error"].includes(st)
    );
  });

  const target = (candidates.length ? candidates : attempts).slice();
  target.sort((a, b) => {
    const at = Date.parse(a.updated_at || "") || 0;
    const bt = Date.parse(b.updated_at || "") || 0;
    return bt - at;
  });
  return target[0] || null;
}

async function fetchVk(path, options = {}) {
  const url = `${vkEndpointUrl}${path}`;
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 10_000);
  try {
    const res = await fetch(url, {
      signal: controller.signal,
      headers: {
        "Content-Type": "application/json",
        ...(options.headers || {}),
      },
      ...options,
    });
    const text = await res.text();
    let data = null;
    try {
      data = text ? JSON.parse(text) : null;
    } catch {
      data = text || null;
    }
    return { ok: res.ok, status: res.status, data, raw: text };
  } catch (err) {
    return { ok: false, status: 0, data: null, error: err.message };
  } finally {
    clearTimeout(timeout);
  }
}

async function summarizeTaskWithCodex(task, errorLine) {
  if (!codexEnabled) {
    return {
      summary:
        "Codex SDK disabled. Manual summary required. Error: " + errorLine,
      usedCodex: false,
    };
  }

  const ready = await ensureCodexSdkReady();
  if (!ready) {
    return {
      summary:
        "Codex SDK unavailable. Manual summary required. Error: " + errorLine,
      usedCodex: false,
    };
  }

  const title = task?.title || task?.name || "Untitled task";
  const description = String(task?.description || task?.body || "")
    .replace(/\r/g, "")
    .slice(0, 6000);
  const prompt = `Summarize the following task for a fresh agent in <=10 bullets.
Focus on: goals, required work areas/files, acceptance criteria, failure modes, tests, and operator docs.
Mention that the previous attempt failed due to prompt token limit.

Title: ${title}
Description:
${description || "(no description)"}

Output format:
Summary:
- ...
Next steps:
- ...`;

  try {
    const codex = new CodexClient();
    const thread = codex.startThread();
    const result = await thread.run(prompt);
    const summary = String(result || "").trim();
    return {
      summary:
        summary || "Summary unavailable (Codex returned empty response).",
      usedCodex: true,
    };
  } catch (err) {
    return {
      summary: `Codex summary failed: ${err.message || err}`,
      usedCodex: false,
    };
  }
}

function buildTaskSummaryDescription(original, summary) {
  const stamp = new Date().toISOString();
  const header = `AUTO-SUMMARY (${stamp})\n${summary.trim()}\n`;
  const separator = "\n---\n";
  const base = original ? String(original) : "";
  let combined = `${header}${separator}${base}`;
  const maxLen = 9000;
  if (combined.length > maxLen) {
    const trimmed = base.slice(
      0,
      Math.max(0, maxLen - header.length - separator.length - 200),
    );
    combined = `${header}${separator}${trimmed}\n\n[truncated by codex-monitor]\n`;
  }
  return combined;
}

async function handleContextOverflow(line) {
  const key = getContextOverflowKey(line);
  const lastAt = contextOverflowHandled.get(key) || 0;
  if (Date.now() - lastAt < contextOverflowCooldownMs) {
    return;
  }
  contextOverflowHandled.set(key, Date.now());

  const status = await readStatusData();
  const attempt = pickAttemptForContextOverflow(status);
  if (!attempt || !attempt.task_id) {
    if (telegramToken && telegramChatId) {
      void sendTelegramMessage(
        `⚠️ Context window error detected but no task attempt found. Line: ${line.slice(0, 200)}`,
      );
    }
    return;
  }

  const taskId = attempt.task_id;
  const taskRes = await fetchVk(`/api/tasks/${taskId}`);
  const task = taskRes.ok ? taskRes.data : null;

  const { summary, usedCodex } = await summarizeTaskWithCodex(task, line);
  const originalDesc = task?.description || task?.body || "";
  const newDesc = buildTaskSummaryDescription(originalDesc, summary);

  let updateOk = false;
  if (taskId) {
    const updateRes = await fetchVk(`/api/tasks/${taskId}`, {
      method: "PUT",
      body: JSON.stringify({
        status: "todo",
        description: newDesc,
      }),
    });
    if (updateRes.ok) {
      updateOk = true;
    } else {
      const fallbackRes = await fetchVk(`/api/tasks/${taskId}`, {
        method: "PUT",
        body: JSON.stringify({ status: "todo" }),
      });
      updateOk = fallbackRes.ok;
    }
  }

  const logPath = resolve(
    logDir,
    `context-overflow-${nowStamp()}-${taskId.slice(0, 8)}.log`,
  );
  await writeFile(
    logPath,
    [
      `# Context overflow detected`,
      `# Task: ${taskId}`,
      `# Branch: ${attempt.branch || "unknown"}`,
      `# Updated: ${new Date().toISOString()}`,
      `# Update OK: ${updateOk}`,
      `# Codex used: ${usedCodex}`,
      "",
      "## Error line",
      line,
      "",
      "## Summary",
      summary,
    ].join("\n"),
    "utf8",
  );

  if (telegramToken && telegramChatId) {
    const taskTitle = task?.title || task?.name || taskId;
    const statusMsg = updateOk ? "re-queued to TODO" : "update failed";
    void sendTelegramMessage(
      `🧠 Context limit hit. ${taskTitle} ${statusMsg}.\nSummary saved in logs.\n${summary.slice(0, 500)}`,
    );
  }
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
        text: "VirtEngine Orchestrator Update\nStatus: unavailable (missing status file)",
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
      "VirtEngine Orchestrator 10-min Update",
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
      `Counts: running=${running}, review=${review}, error=${error}, manual_review=${manualReview}`,
      `Success: ${successLine}`,
    ].join("\n");

    return { text: message, parseMode: "HTML" };
  } catch (err) {
    return {
      text: "VirtEngine Orchestrator Update\nStatus: unavailable (missing status file)",
      parseMode: null,
    };
  }
}

async function sendTelegramMessage(text, options = {}) {
  if (!telegramToken || !telegramChatId) {
    return;
  }
  const key = String(text || "").trim();
  if (key) {
    const now = Date.now();
    const last = telegramDedup.get(key) || 0;
    if (now - last < 5 * 60 * 1000) {
      return;
    }
    telegramDedup.set(key, now);
  }

  // Always record to history ring buffer (even deduped messages are useful context)
  pushTelegramHistory(String(text || ""));

  const url = `https://api.telegram.org/bot${telegramToken}/sendMessage`;
  const payload = {
    chat_id: telegramChatId,
    text,
  };
  if (options.parseMode) {
    payload.parse_mode = options.parseMode;
  }
  if (options.disablePreview) {
    payload.disable_web_page_preview = true;
  }
  try {
    const res = await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const body = await res.text();
      console.warn(`[monitor] telegram send failed: ${res.status} ${body}`);
    } else {
      // ── Bottom-pinning: bump the active agent message down ──────
      // If the Telegram bot has an active agent session, re-send its message
      // so it stays at the bottom of the chat (below this notification).
      if (isAgentActive()) {
        // Small delay to ensure Telegram processes the notification first
        setTimeout(() => void bumpAgentMessage(), 500);
      }
    }
  } catch (err) {
    console.warn(`[monitor] telegram send failed: ${err.message || err}`);
  }
}

function startTelegramNotifier() {
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
  void sendTelegramMessage("VirtEngine Orchestrator Notifier started.");
  setTimeout(sendUpdate, intervalMs);
  setInterval(sendUpdate, intervalMs);
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
    await triggerTaskPlanner();
  }

  if (!backlogLowNotified && backlogRemaining > 0 && backlogRemaining < 5) {
    backlogLowNotified = true;
    await sendTelegramMessage(
      `Backlog low: ${backlogRemaining} tasks remaining. Triggering task planner.`,
    );
    await triggerTaskPlanner();
  }

  // ── Auto-reattempt: check for tasks with repeated singular failures ──
  await checkAutoReattempt(status);
}

/**
 * Check for tasks that have hit the reattempt threshold (3 singular failures).
 * Singular = the failure affects only this specific task, not a shared infrastructure issue.
 */
async function checkAutoReattempt(status) {
  if (!status) return;
  const attempts = status.attempts || {};
  const errorTasks = Object.entries(attempts).filter(
    ([, a]) => a?.status === "error",
  );

  if (errorTasks.length === 0) return;

  const now = Date.now();

  // ── Shared failure detection ──────────────────────────────────────
  // If 3+ tasks fail within SHARED_FAILURE_WINDOW_MS, it's a shared issue
  // (e.g., VK crash, monitor crash, network outage) — don't count these.
  for (const [, attempt] of errorTasks) {
    const failTime = Date.parse(attempt.updated_at || "") || now;
    sharedFailureTimestamps.push(failTime);
  }
  // Prune old timestamps
  while (
    sharedFailureTimestamps.length > 0 &&
    now - sharedFailureTimestamps[0] > SHARED_FAILURE_WINDOW_MS
  ) {
    sharedFailureTimestamps.shift();
  }
  const recentFailures = sharedFailureTimestamps.filter(
    (ts) => now - ts < SHARED_FAILURE_WINDOW_MS,
  );
  const isSharedFailure = recentFailures.length >= 3;

  if (isSharedFailure) {
    console.log(
      `[monitor] ${recentFailures.length} tasks failed within ${SHARED_FAILURE_WINDOW_MS / 1000}s — shared failure, skipping auto-reattempt`,
    );
    return;
  }

  // ── Per-task singular failure counting ────────────────────────────
  for (const [attemptId, attempt] of errorTasks) {
    const taskId = attempt.task_id || attemptId;
    const existing = taskFailureCounts.get(taskId) || {
      count: 0,
      lastFailedAt: 0,
      lastBranch: "",
    };

    // Only increment if this is a new failure (different timestamp)
    const failTime = Date.parse(attempt.updated_at || "") || now;
    if (failTime > existing.lastFailedAt) {
      existing.count += 1;
      existing.lastFailedAt = failTime;
      existing.lastBranch = attempt.branch || "";
      taskFailureCounts.set(taskId, existing);

      console.log(
        `[monitor] Task ${(attempt.task_title || taskId).slice(0, 40)} failure count: ${existing.count}/${REATTEMPT_THRESHOLD}`,
      );
    }

    // ── Trigger auto-reattempt at threshold ──────────────────────
    if (existing.count >= REATTEMPT_THRESHOLD) {
      await performAutoReattempt(attemptId, attempt, taskId);
      // Reset counter after reattempt
      taskFailureCounts.delete(taskId);
    }
  }
}

/**
 * Automatically reattempt a task: archive branch, close PR, reset to TODO.
 */
async function performAutoReattempt(attemptId, attempt, taskId) {
  const title = attempt.task_title || taskId;
  console.log(
    `[monitor] Auto-reattempting task: ${title} (${REATTEMPT_THRESHOLD} failures)`,
  );

  const results = [];

  // Step 1: Archive the branch
  if (attempt.branch) {
    try {
      const archiveName = `archive/${attempt.branch}-${Date.now()}`;
      execSync(`git branch -m ${attempt.branch} ${archiveName}`, {
        cwd: repoRoot,
        encoding: "utf8",
        timeout: 10000,
        stdio: "pipe",
      });
      results.push(`Branch archived: ${attempt.branch}`);
    } catch {
      results.push(`Branch ${attempt.branch} not found locally`);
    }
  }

  // Step 2: Close PR if exists
  if (attempt.pr_number) {
    try {
      execSync(
        `gh pr close ${attempt.pr_number} --comment "Auto-reattempt after ${REATTEMPT_THRESHOLD} failures"`,
        { cwd: repoRoot, encoding: "utf8", timeout: 15000, stdio: "pipe" },
      );
      results.push(`PR #${attempt.pr_number} closed`);
    } catch {
      results.push(`PR #${attempt.pr_number} could not be closed`);
    }
  }

  // Step 3: Reset task to TODO in Vibe Kanban
  if (taskId) {
    try {
      const res = await fetchVk(`/api/tasks/${taskId}`, {
        method: "PUT",
        body: JSON.stringify({
          status: "todo",
          description:
            (attempt.original_description || "") +
            `\n\n---\n⚠️ Auto-reattempt triggered at ${new Date().toISOString()} after ${REATTEMPT_THRESHOLD} consecutive failures.`,
        }),
      });
      if (res.ok) {
        results.push("Task reset to TODO in Vibe Kanban");
      } else {
        results.push(`VK API error: ${res.status}`);
      }
    } catch (err) {
      results.push(`VK API error: ${err.message}`);
    }
  }

  // Step 4: Update local status
  try {
    const raw = await readFile(statusPath, "utf8");
    const data = JSON.parse(raw);
    if (data.attempts && data.attempts[attemptId]) {
      data.attempts[attemptId].status = "reattempted";
      data.attempts[attemptId].reattempted_at = new Date().toISOString();
      data.attempts[attemptId].failure_count = REATTEMPT_THRESHOLD;
    }
    // Store failure counts in status for /tasks visibility
    if (!data.task_failure_counts) data.task_failure_counts = {};
    data.task_failure_counts[taskId] = REATTEMPT_THRESHOLD;
    await writeFile(statusPath, JSON.stringify(data, null, 2), "utf8");
  } catch {
    /* best effort */
  }

  // Notify via Telegram
  if (telegramToken && telegramChatId) {
    void sendTelegramMessage(
      `🔄 Auto-reattempt: ${title}\n\n` +
        `This task failed ${REATTEMPT_THRESHOLD} times consecutively (singular failures, not shared).\n\n` +
        results.map((r) => `• ${r}`).join("\n") +
        "\n\nTask will be picked up fresh in the next orchestrator cycle.",
    );
  }
}

async function triggerTaskPlanner() {
  if (plannerTriggered || !codexEnabled) {
    return;
  }
  plannerTriggered = true;
  try {
    notifyCodexTrigger("task planner run");
    if (!CodexClient) {
      CodexClient = await loadCodexSdk();
    }
    if (!CodexClient) {
      throw new Error("Codex SDK not available");
    }
    const agentPath = resolve(
      repoRoot,
      ".github",
      "agents",
      "Task Planner.agent.md",
    );
    const agentPrompt = await readFile(agentPath, "utf8");
    const codex = new CodexClient();
    const thread = codex.startThread();
    const prompt = `${agentPrompt}\n\nPlease execute the task planning instructions above.`;
    const result = await thread.run(prompt);
    const outPath = resolve(logDir, `task-planner-${nowStamp()}.md`);
    await writeFile(outPath, String(result), "utf8");
    await sendTelegramMessage(
      "Task planner run completed. Output saved to logs.",
    );
  } catch (err) {
    const message = err && err.message ? err.message : String(err);
    await sendTelegramMessage(`Task planner run failed: ${message}`);
  }
}

async function ensureLogDir() {
  await mkdir(logDir, { recursive: true });
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

async function analyzeWithCodex(logPath, logText, reason) {
  if (!codexEnabled) {
    return;
  }
  try {
    notifyCodexTrigger(`orchestrator analysis (${reason})`);
    if (!CodexClient) {
      const ready = await ensureCodexSdkReady();
      if (!ready) {
        throw new Error(codexDisabledReason || "Codex SDK not available");
      }
    }
    const codex = new CodexClient();
    const thread = codex.startThread();
    const prompt = `You are monitoring the VirtEngine task orchestrator (a PowerShell script).
The script just crashed. Diagnose the root cause from the log and apply a minimal fix.

Key files (read these first):
- scripts/ve-orchestrator.ps1 — main orchestrator (PowerShell)
- scripts/ve-kanban.ps1 — vibe-kanban API helpers
- scripts/codex-monitor/monitor.mjs — process supervisor
- scripts/codex-monitor/autofix.mjs — auto-fix engine
- AGENTS.md — project conventions

Exit reason: ${reason}

Log excerpt (last lines):
${logText}

Rules:
1. Only edit the files listed above.
2. Fix the specific error shown in the log — do not refactor.
3. Validate your fix compiles: run the PowerShell parser or node --check.
4. If the error is a PowerShell ParserError, find the exact line and fix the syntax.
5. If the error is a null parameter, add a null guard.
6. Do NOT touch unrelated code.`;

    const result = await thread.run(prompt);
    const analysisPath = logPath.replace(/\.log$/, "-analysis.txt");
    const analysisText = String(result);
    await writeFile(analysisPath, analysisText, "utf8");

    // Notify user with Codex analysis outcome
    if (telegramToken && telegramChatId) {
      // Extract first 500 chars of diagnosis for telegram
      const summary = analysisText.slice(0, 500).replace(/\n{3,}/g, "\n\n");
      void sendTelegramMessage(
        `🔍 Codex Analysis Result (${reason}):\n${summary}${analysisText.length > 500 ? "\n...(truncated)" : ""}`,
      );
    }
  } catch (err) {
    const analysisPath = logPath.replace(/\.log$/, "-analysis.txt");
    const message = err && err.message ? err.message : String(err);
    await writeFile(analysisPath, `Codex SDK failed: ${message}\n`, "utf8");
    if (telegramToken && telegramChatId) {
      void sendTelegramMessage(`🔍 Codex Analysis Failed: ${message}`);
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
  const wasGracefulKill = gracefulKill;
  const isAbnormalExit = Boolean(signal) || code !== 0;

  // ── Graceful restart (file-change, monitor-initiated kill) ────────────
  // When the monitor itself killed the child (requestRestart), skip ALL
  // autofix and codex analysis.  Just restart cleanly.
  if (pendingRestart || wasGracefulKill) {
    pendingRestart = false;
    skipNextAnalyze = false;
    gracefulKill = false;
    if (!skipNextRestartCount) {
      restartCount += 1;
    }
    skipNextRestartCount = false;

    console.log(
      `[monitor] graceful restart (${reason}) — skipping autofix/analysis`,
    );
    startProcess();
    return;
  }

  // ── Clean exit (code 0, no signal) — don't treat as crash ─────────────
  if (code === 0 && !signal) {
    console.log(
      `[monitor] clean exit (${reason}) — restarting without analysis`,
    );
    restartCount += 1;
    setTimeout(startProcess, restartDelayMs);
    return;
  }

  // ── External kill (hotfix/runner) — skip autofix/analysis ─────────────
  // External supervisors often terminate the orchestrator with SIGTERM/SIGKILL
  // during hotfix restarts. Treat these as non-actionable to avoid Codex noise.
  if (signal === "SIGTERM" || signal === "SIGKILL") {
    console.warn(
      `[monitor] external kill (${reason}) — skipping autofix/analysis`,
    );
    restartCount += 1;
    setTimeout(startProcess, restartDelayMs);
    return;
  }

  // ── Windows crash codes — not orchestrator bugs ───────────────────────
  // Exit codes like STATUS_LOG_HARD_ERROR (0x40010004) or STATUS_STOWED_EXCEPTION
  // (0xC000027B) are Windows runtime / COM crashes in the child process (pwsh/Codex),
  // not bugs in our scripts. Auto-fix cannot help — just restart with backoff.
  if (code !== null && WINDOWS_CRASH_CODES.has(code)) {
    console.warn(
      `[monitor] Windows crash code ${code} (0x${code.toString(16).toUpperCase()}) — skipping autofix (not an orchestrator bug)`,
    );
    if (telegramToken && telegramChatId) {
      void sendTelegramMessage(
        `⚠️ Orchestrator exited with Windows crash code ${code} (0x${code.toString(16).toUpperCase()}). ` +
          `This is a host/runtime issue, not a script bug. Restarting with backoff.`,
      );
    }
    restartCount += 1;
    // Use longer delay for Windows crashes — they often recur if the system is stressed
    const backoffMs = Math.min(
      restartDelayMs * Math.pow(2, Math.min(restartCount, 5)),
      5 * 60 * 1000,
    );
    console.log(
      `[monitor] Windows crash backoff: restarting in ${Math.round(backoffMs / 1000)}s`,
    );
    setTimeout(startProcess, backoffMs);
    return;
  }

  // ── Abnormal exit — attempt autofix ───────────────────────────────────
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
        execViaShell: getExecViaShell(),
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

  // ── Fallback: Codex SDK analysis (diagnosis only, abnormal exits only) ──
  if (isAbnormalExit && codexEnabled) {
    if (telegramToken && telegramChatId) {
      void sendTelegramMessage(
        `🔍 Codex analysis triggered (${reason}):\nAuto-fix was unable to resolve the crash — running diagnostic analysis.`,
      );
    }
    await analyzeWithCodex(logPath, logText.slice(-15000), reason);
  }

  if (hasContextWindowError(logText)) {
    await writeFile(
      logPath.replace(/\.log$/, "-context.txt"),
      "Detected context window error. Consider creating a new workspace session and re-sending follow-up.\n",
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
          } else if (telegramToken && telegramChatId) {
            void sendTelegramMessage(
              `⚠️ Crash-loop fix attempt failed: ${fixResult.outcome}. Orchestrator remains paused.`,
            );
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
  if (currentChild && !currentChild.killed) {
    return;
  }
  if (!currentChild && isOrchestratorProcessRunning()) {
    console.warn(
      "[monitor] detected existing orchestrator process; deferring start",
    );
    setTimeout(startProcess, restartDelayMs);
    return;
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
  await ensureLogDir();
  const activeLogPath = resolve(logDir, "orchestrator-active.log");
  const archiveLogPath = resolve(logDir, `orchestrator-${nowStamp()}.log`);
  const logStream = await writeFile(activeLogPath, "", "utf8").then(() => null);

  const child = spawn("pwsh", ["-File", scriptPath, ...scriptArgs], {
    stdio: ["ignore", "pipe", "pipe"],
  });
  currentChild = child;

  const append = async (chunk) => {
    if (echoLogs) {
      process.stdout.write(chunk);
    }
    const text = chunk.toString();
    await writeFile(activeLogPath, text, { flag: "a" });
    logRemainder += text;
    const lines = logRemainder.split(/\r?\n/);
    logRemainder = lines.pop() || "";
    for (const line of lines) {
      if (isContextOverflowLine(line)) {
        void handleContextOverflow(line);
      }
      if (isErrorLine(line)) {
        notifyErrorLine(line);
      }
      // Check for executor degradation (reconnect loops, rate limits, timeouts)
      const degradation = checkDegradationPatterns(line);
      if (degradation) {
        trackRegionFailover(degradation);
      }
      if (line.includes("Merged PR") || line.includes("Marking task")) {
        notifyMerge(line);
      }
      if (line.includes("Merge notify: PR #")) {
        notifyMergeFailure(line);
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
  pendingRestart = true;
  skipNextAnalyze = true;
  skipNextRestartCount = true;
  gracefulKill = true;

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
    gracefulKill = false;
    startProcess();
  }
}

async function startWatcher() {
  if (!watchEnabled) {
    return;
  }
  if (watcher) {
    return;
  }
  let targetPath = watchPath;
  try {
    const stats = statSync(watchPath);
    if (stats.isFile()) {
      watchFileName = watchPath.split(/[\\/]/).pop();
      targetPath = watchPath.split(/[\\/]/).slice(0, -1).join("/") || ".";
      watchFileMtimeMs = stats.mtimeMs;
    }
  } catch {
    // Default to watching the provided path.
  }

  watcher = watch(targetPath, { persistent: true }, (_event, filename) => {
    if (watchFileName) {
      if (filename) {
        const fileMatches =
          process.platform === "win32"
            ? filename.toLowerCase() === watchFileName.toLowerCase()
            : filename === watchFileName;
        if (!fileMatches) {
          return;
        }
        try {
          const stats = statSync(watchPath);
          watchFileMtimeMs = stats.mtimeMs;
        } catch {
          // Ignore stat failures; we'll allow restart below.
        }
      } else {
        // Windows often omits filename for fs.watch; only restart if target file changed.
        try {
          const stats = statSync(watchPath);
          if (watchFileMtimeMs !== null && stats.mtimeMs === watchFileMtimeMs) {
            return;
          }
          watchFileMtimeMs = stats.mtimeMs;
        } catch {
          // If file is missing or unreadable, fall through and restart.
        }
      }
    }
    if (watcherDebounce) {
      clearTimeout(watcherDebounce);
    }
    watcherDebounce = setTimeout(() => {
      requestRestart("file-change");
    }, 5000);
  });
}

process.on("SIGINT", () => {
  shuttingDown = true;
  if (watcher) {
    watcher.close();
  }
  if (currentChild) {
    currentChild.kill("SIGTERM");
  }
  process.exit(0);
});

process.on("SIGTERM", () => {
  shuttingDown = true;
  if (watcher) {
    watcher.close();
  }
  if (currentChild) {
    currentChild.kill("SIGTERM");
  }
  process.exit(0);
});

process.on("uncaughtException", (err) => {
  void handleMonitorFailure("uncaughtException", err);
});

process.on("unhandledRejection", (reason) => {
  const err =
    reason instanceof Error ? reason : new Error(String(reason || ""));
  void handleMonitorFailure("unhandledRejection", err);
});

// ── Singleton guard: prevent ghost monitors ─────────────────────────────────
const cacheDir = resolve(repoRoot, ".cache");
if (!acquireMonitorLock(cacheDir)) {
  process.exit(1);
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

startWatcher();
void ensureVibeKanbanRunning().catch(() => {});
if (Number.isFinite(vkEnsureIntervalMs) && vkEnsureIntervalMs > 0) {
  setInterval(() => {
    void ensureVibeKanbanRunning().catch(() => {});
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
startTelegramNotifier();

// ── Two-way Telegram ↔ Codex shell ──────────────────────────────────────────
injectMonitorFunctions({
  sendTelegramMessage,
  readStatusData,
  readStatusSummary,
  getCurrentChild: () => currentChild,
  startProcess,
  getVibeKanbanUrl: () => vkPublicUrl || vkEndpointUrl,
  fetchVk,
  getRepoRoot: () => repoRoot,
});
void startTelegramBot();
