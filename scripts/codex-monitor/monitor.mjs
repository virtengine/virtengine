import { execSync, spawn, spawnSync } from "node:child_process";
import { existsSync, watch } from "node:fs";
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
  getDefaultExecutorProfile,
  getExecutorProfileForModel,
  loadWorkspaceRegistry,
  normalizeModelToken,
  normalizeRole,
  workspaceSupportsModel,
} from "../shared/workspace-registry.mjs";

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
const scriptArgsRaw = getArg("--args", "-MaxParallel 6");
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
const telegramPollIntervalSec = Number(
  process.env.TELEGRAM_POLL_INTERVAL_SEC ||
    process.env.TELEGRAM_POLL_INTERVAL ||
    "8",
);
const telegramPollTimeoutSec = Number(
  process.env.TELEGRAM_POLL_TIMEOUT_SEC || "8",
);
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
let shuttingDown = false;
let currentChild = null;
let pendingRestart = false;
let skipNextAnalyze = false;
let skipNextRestartCount = false;
let watcher = null;
let watcherDebounce = null;
let watchFileName = null;
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
let workspaceRegistry = null;
let workspaceRegistryErrors = [];
let workspaceRegistrySource = "unknown";
let workspaceRegistryNotified = "";
let telegramUpdateOffset = 0;
let telegramPolling = false;

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
    const args = ["exec", "--full-auto", "-C", cwd, prompt];
    const child = spawn("codex", args, {
      cwd,
      stdio: ["ignore", "pipe", "pipe"],
      shell: true,
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
  const result = await runCodexExec(prompt, repoRoot);
  const filesAfter = detectChangedFiles(repoRoot);
  const newChanges = filesAfter.filter((f) => !filesBefore.includes(f));
  const changeSummary = getChangeSummary(repoRoot, newChanges);

  const stamp = nowStamp();
  const auditPath = resolve(logDir, `monitor-fix-${stamp}-attempt${attemptNum}.log`);
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
  const result = await runCodexExec(prompt, repoRoot, 180_000);
  const filesAfter = detectChangedFiles(repoRoot);
  const newChanges = filesAfter.filter((f) => !filesBefore.includes(f));
  const changeSummary = getChangeSummary(repoRoot, newChanges);

  const stamp = nowStamp();
  const auditPath = resolve(logDir, `crash-loop-fix-${stamp}-attempt${attemptNum}.log`);
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

const AGENT_COMMAND_USAGE =
  "/agent [--workspace <id>] [--role <role>] [--model <name>] [--queue] <message>";

async function loadWorkspaceRegistryConfig() {
  const { registry, errors, source } = await loadWorkspaceRegistry({
    env: process.env,
    baseDir: repoRoot,
  });
  workspaceRegistry = registry;
  workspaceRegistryErrors = errors;
  workspaceRegistrySource = source;
  reportWorkspaceRegistryErrors(errors);
  return registry;
}

function reportWorkspaceRegistryErrors(errors) {
  if (!errors || !errors.length) return;
  const signature = errors.join("|");
  if (signature === workspaceRegistryNotified) return;
  workspaceRegistryNotified = signature;
  const summary = errors.map((err) => `- ${err}`).join("\n");
  console.warn(
    `[monitor] workspace registry validation errors (source=${workspaceRegistrySource}):\n${summary}`,
  );
  if (telegramToken && telegramChatId) {
    void sendTelegramMessage(
      `⚠️ Workspace registry validation errors (${workspaceRegistrySource}):\n${summary}`,
    );
  }
}

function tokenizeCommand(text) {
  const tokens = [];
  const regex = /"([^"]*)"|'([^']*)'|(\S+)/g;
  let match;
  while ((match = regex.exec(text))) {
    tokens.push(match[1] ?? match[2] ?? match[3]);
  }
  return tokens;
}

function parseAgentCommand(text) {
  if (!text) return null;
  const match = String(text).trim().match(/^\/agent(?:@\w+)?\s*(.*)$/i);
  if (!match) return null;
  const remainder = match[1] || "";
  const tokens = tokenizeCommand(remainder);
  const result = { queue: false, errors: [] };
  const messageParts = [];

  for (let i = 0; i < tokens.length; i += 1) {
    const token = tokens[i];
    if (!token) continue;
    if (token === "--help" || token === "-h") {
      result.help = true;
      continue;
    }
    if (token === "--queue") {
      result.queue = true;
      continue;
    }
    if (token.startsWith("--workspace=")) {
      result.workspaceId = token.split("=").slice(1).join("=");
      continue;
    }
    if (token === "--workspace" || token === "-w") {
      const next = tokens[i + 1];
      if (!next) {
        result.errors.push("Missing value for --workspace");
      } else {
        result.workspaceId = next;
        i += 1;
      }
      continue;
    }
    if (token.startsWith("--role=")) {
      result.role = token.split("=").slice(1).join("=");
      continue;
    }
    if (token === "--role" || token === "-r") {
      const next = tokens[i + 1];
      if (!next) {
        result.errors.push("Missing value for --role");
      } else {
        result.role = next;
        i += 1;
      }
      continue;
    }
    if (token.startsWith("--model=")) {
      result.model = token.split("=").slice(1).join("=");
      continue;
    }
    if (token === "--model" || token === "-m") {
      const next = tokens[i + 1];
      if (!next) {
        result.errors.push("Missing value for --model");
      } else {
        result.model = next;
        i += 1;
      }
      continue;
    }
    messageParts.push(token);
  }

  result.message = messageParts.join(" ").trim();
  return result;
}

async function fetchJson(url, options = {}) {
  const { method = "GET", body, signal } = options;
  const headers = { ...(options.headers || {}) };
  const payload = body ? JSON.stringify(body) : undefined;
  if (payload) {
    headers["content-type"] = "application/json";
  }
  const res = await fetch(url, { method, body: payload, headers, signal });
  const text = await res.text();
  let json = null;
  if (text) {
    try {
      json = JSON.parse(text);
    } catch {
      json = null;
    }
  }
  return { ok: res.ok, status: res.status, json, text };
}

function unwrapVkResponse(payload) {
  if (!payload) return null;
  if (payload.success === false) {
    return { error: payload.message || "VK API error", raw: payload };
  }
  return payload.data ?? payload;
}

async function probeWorkspaceHealth(workspace) {
  const host = workspace?.host;
  if (!host) {
    return { ok: false, score: 0, reason: "missing-host" };
  }
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 2500);
  const started = Date.now();
  try {
    const res = await fetch(`${host}/api/projects`, {
      signal: controller.signal,
    });
    const latencyMs = Date.now() - started;
    if (!res.ok) {
      return {
        ok: false,
        score: 0,
        reason: `http-${res.status}`,
        latencyMs,
      };
    }
    let score = 70;
    let running = 0;
    let failed = 0;
    let summaryOk = false;
    try {
      const summaryRes = await fetchJson(`${host}/api/task-attempts/summary`, {
        method: "POST",
        body: { archived: false },
        signal: controller.signal,
      });
      if (summaryRes.ok && summaryRes.json) {
        const payload = unwrapVkResponse(summaryRes.json) ?? summaryRes.json;
        const summaries = payload?.summaries || payload;
        if (Array.isArray(summaries)) {
          summaries.forEach((summary) => {
            const status = summary?.latest_process_status;
            if (status === "running") running += 1;
            if (status === "failed") failed += 1;
          });
          summaryOk = true;
        }
      }
    } catch {
      summaryOk = false;
    }
    if (summaryOk) {
      score = 100 - running * 5 - failed * 12;
      if (score < 10) score = 10;
    }
    return {
      ok: true,
      score,
      latencyMs,
      running,
      failed,
    };
  } catch (err) {
    return {
      ok: false,
      score: 0,
      reason: err?.message || String(err),
    };
  } finally {
    clearTimeout(timeout);
  }
}

async function probeWorkspaces(workspaces) {
  const entries = await Promise.all(
    workspaces.map(async (workspace) => ({
      workspace,
      health: await probeWorkspaceHealth(workspace),
    })),
  );
  const map = new Map();
  entries.forEach((entry) => {
    if (entry.workspace?.id) {
      map.set(entry.workspace.id, entry.health);
    }
  });
  return map;
}

function pickBestWorkspace(candidates, healthMap) {
  const scored = candidates
    .map((workspace) => ({
      workspace,
      health: healthMap.get(workspace.id),
    }))
    .filter((entry) => entry.health && entry.health.ok);
  if (!scored.length) return null;
  scored.sort((a, b) => (b.health.score || 0) - (a.health.score || 0));
  return scored[0];
}

async function resolveWorkspaceTarget({ workspaceId, role, model }) {
  if (!workspaceRegistry || !workspaceRegistry.workspaces?.length) {
    return { error: "No workspace registry configured." };
  }
  const workspaces = workspaceRegistry.workspaces;
  const normalizedRole = normalizeRole(role || "primary");
  const normalizedModel = normalizeModelToken(model);

  let candidates = workspaces;
  let reason = "fallback";
  let explicitMissing = false;

  if (workspaceId) {
    const idToken = normalizeModelToken(workspaceId);
    const match = workspaces.find(
      (workspace) => normalizeModelToken(workspace.id) === idToken,
    );
    if (match) {
      candidates = [match];
      reason = "explicit";
    } else {
      explicitMissing = true;
    }
  } else if (normalizedRole) {
    const roleMatches = workspaces.filter(
      (workspace) => normalizeRole(workspace.role) === normalizedRole,
    );
    if (roleMatches.length) {
      candidates = roleMatches;
      reason = "role";
    }
  }

  if (normalizedModel) {
    const modelMatches = candidates.filter((workspace) =>
      workspaceSupportsModel(workspace, normalizedModel),
    );
    if (modelMatches.length) {
      candidates = modelMatches;
    }
  }

  const healthMap = await probeWorkspaces(workspaces);
  let selected = pickBestWorkspace(candidates, healthMap);
  let fallbackUsed = false;

  if (!selected) {
    const fallback = pickBestWorkspace(workspaces, healthMap);
    if (fallback) {
      selected = fallback;
      fallbackUsed = true;
    }
  }

  if (!selected) {
    return {
      error: "No available workspaces (all targets unreachable).",
      reason: explicitMissing ? "explicit-missing" : reason,
    };
  }

  return {
    workspace: selected.workspace,
    health: selected.health,
    reason,
    fallbackUsed,
    explicitMissing,
  };
}

function getVkBaseUrl(host) {
  return host.replace(/\/+$/, "");
}

async function getWorkspaceSessions(host, workspaceId) {
  const url = `${getVkBaseUrl(host)}/api/sessions?workspace_id=${encodeURIComponent(
    workspaceId,
  )}`;
  const res = await fetchJson(url);
  if (!res.ok || !res.json) return [];
  const payload = unwrapVkResponse(res.json) ?? res.json;
  return Array.isArray(payload) ? payload : [];
}

async function createWorkspaceSession(host, workspaceId) {
  const url = `${getVkBaseUrl(host)}/api/sessions`;
  const res = await fetchJson(url, {
    method: "POST",
    body: { workspace_id: workspaceId },
  });
  if (!res.ok || !res.json) return null;
  const payload = unwrapVkResponse(res.json) ?? res.json;
  return payload?.id ? payload : null;
}

function pickLatestSession(sessions) {
  if (!Array.isArray(sessions) || sessions.length === 0) return null;
  return sessions
    .slice()
    .sort((a, b) => {
      const aStamp = Date.parse(a.updated_at || a.created_at || 0) || 0;
      const bStamp = Date.parse(b.updated_at || b.created_at || 0) || 0;
      return bStamp - aStamp;
    })[0];
}

async function sendAgentMessage({
  workspace,
  message,
  model,
  queue = false,
}) {
  const host = workspace.host;
  const workspaceId = workspace.id;
  const sessions = await getWorkspaceSessions(host, workspaceId);
  let session = pickLatestSession(sessions);
  if (!session) {
    session = await createWorkspaceSession(host, workspaceId);
  }
  if (!session) {
    throw new Error("Failed to create or resolve a workspace session.");
  }

  const profileFromModel = getExecutorProfileForModel(workspace, model);
  const profileFromSession = session.executor
    ? getDefaultExecutorProfile(session.executor)
    : null;
  const executorProfile =
    profileFromModel || profileFromSession || getDefaultExecutorProfile("CODEX");

  const path = queue
    ? `/api/sessions/${session.id}/queue`
    : `/api/sessions/${session.id}/follow-up`;
  const body = queue
    ? { executor_profile_id: executorProfile, message }
    : { executor_profile_id: executorProfile, prompt: message };
  const res = await fetchJson(`${getVkBaseUrl(host)}${path}`, {
    method: "POST",
    body,
  });
  if (!res.ok) {
    throw new Error(`VK API error (${res.status || "unknown"})`);
  }
  const payload = unwrapVkResponse(res.json);
  if (payload && payload.error) {
    throw new Error(payload.error);
  }
  return {
    sessionId: session.id,
    executorProfile,
  };
}

async function handleAgentCommand(command, messageMeta) {
  await loadWorkspaceRegistryConfig();
  const chatId = String(messageMeta?.chat?.id || "");
  if (telegramChatId && chatId && chatId !== String(telegramChatId)) {
    return;
  }
  if (command.help || command.errors?.length) {
    const errors = command.errors?.length
      ? `Errors: ${command.errors.join("; ")}\n`
      : "";
    await sendTelegramMessage(`${errors}Usage: ${AGENT_COMMAND_USAGE}`);
    return;
  }
  if (!command.message) {
    await sendTelegramMessage(`Usage: ${AGENT_COMMAND_USAGE}`);
    return;
  }

  const resolution = await resolveWorkspaceTarget({
    workspaceId: command.workspaceId,
    role: command.role,
    model: command.model,
  });

  if (resolution.error) {
    await sendTelegramMessage(`⚠️ /agent failed: ${resolution.error}`);
    return;
  }

  const fallbackNote = resolution.fallbackUsed
    ? " (fallback)"
    : resolution.explicitMissing
      ? " (workspace not found, fallback)"
      : "";
  const workspace = resolution.workspace;
  try {
    const result = await sendAgentMessage({
      workspace,
      message: command.message,
      model: command.model,
      queue: command.queue,
    });
    const modelNote = command.model ? ` model=${command.model}` : "";
    await sendTelegramMessage(
      `✅ Routed /agent to ${workspace.name} (${workspace.id})${fallbackNote}. Session ${result.sessionId}.${modelNote}`,
    );
  } catch (err) {
    await sendTelegramMessage(
      `⚠️ /agent failed to send message: ${err?.message || err}`,
    );
  }
}

async function pollTelegramUpdates() {
  if (!telegramToken || !telegramChatId) {
    return;
  }
  if (telegramPolling) return;
  telegramPolling = true;
  try {
    const offset = telegramUpdateOffset
      ? `&offset=${telegramUpdateOffset}`
      : "";
    const timeout = Number.isFinite(telegramPollTimeoutSec)
      ? telegramPollTimeoutSec
      : 8;
    const url = `https://api.telegram.org/bot${telegramToken}/getUpdates?timeout=${timeout}${offset}`;
    const res = await fetchJson(url);
    if (!res.ok || !res.json || res.json.ok === false) {
      return;
    }
    const updates = res.json.result || [];
    for (const update of updates) {
      if (typeof update.update_id === "number") {
        telegramUpdateOffset = update.update_id + 1;
      }
      const message = update.message || update.edited_message;
      if (!message?.text) continue;
      if (String(message.chat?.id || "") !== String(telegramChatId)) {
        continue;
      }
      const command = parseAgentCommand(message.text);
      if (command) {
        await handleAgentCommand(command, message);
      }
    }
  } catch (err) {
    console.warn(
      `[monitor] telegram polling error: ${err?.message || err}`,
    );
  } finally {
    telegramPolling = false;
  }
}

function startTelegramCommandListener() {
  if (!telegramToken || !telegramChatId) {
    return;
  }
  if (!Number.isFinite(telegramPollIntervalSec) || telegramPollIntervalSec <= 0) {
    return;
  }
  setInterval(() => {
    void pollTelegramUpdates();
  }, telegramPollIntervalSec * 1000);
  void pollTelegramUpdates();
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
];

const vkErrorPatterns = [
  /Failed to initialize vibe-kanban configuration/i,
  /HTTP GET http:\/\/127\.0\.0\.1:54089\/api\/projects failed/i,
];

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

function startVibeKanbanProcess() {
  if (vibeKanbanProcess && !vibeKanbanProcess.killed) {
    return;
  }

  const env = {
    ...process.env,
    PORT: vkRecoveryPort,
    HOST: vkRecoveryHost,
  };

  console.log(
    `[monitor] starting vibe-kanban via npx (HOST=${vkRecoveryHost} PORT=${vkRecoveryPort}, endpoint=${vkEndpointUrl})`,
  );

  vibeKanbanProcess = spawn("npx", ["--yes", "vibe-kanban"], {
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
  setTimeout(() => startVibeKanbanProcess(), delay);
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
  startVibeKanbanProcess();
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
    startVibeKanbanProcess();
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
    const prompt = `You are monitoring a long-running PowerShell orchestration script.
The script just exited with an error. Analyze the log excerpt and implement a fix on the file (scripts/ve-orchestrator.ps1).

Reason: ${reason}

Log excerpt:\n${logText}\n
Return a short diagnosis and a concrete fix.`;

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
  const isAbnormalExit = Boolean(signal) || code !== 0;

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
    const waitMs = Math.max(orchestratorHaltedUntil - now, monitorSafeModeUntil - now);
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
  if (now < orchestratorHaltedUntil || now < monitorSafeModeUntil) {
    const waitMs = Math.max(orchestratorHaltedUntil - now, monitorSafeModeUntil - now);
    const waitSec = Math.max(5, Math.round(waitMs / 1000));
    console.warn(`[monitor] orchestrator start blocked; retrying in ${waitSec}s`);
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
      if (isErrorLine(line)) {
        notifyErrorLine(line);
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

async function startWatcher() {
  if (!watchEnabled) {
    return;
  }
  if (watcher) {
    return;
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
void ensureVibeKanbanRunning();
if (Number.isFinite(vkEnsureIntervalMs) && vkEnsureIntervalMs > 0) {
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
void loadWorkspaceRegistryConfig();
startProcess();
startTelegramNotifier();
startTelegramCommandListener();
