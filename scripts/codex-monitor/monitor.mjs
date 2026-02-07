import { spawn, spawnSync } from "node:child_process";
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
const vkRecoveryHost = process.env.VK_RECOVERY_HOST || "0.0.0.0";
const vkBaseUrl =
  process.env.VK_BASE_URL || `http://127.0.0.1:${vkRecoveryPort}`;
const vkPublicUrl = process.env.VK_PUBLIC_URL || process.env.VK_WEB_URL || "";
const vkRecoveryCooldownMin = Number(
  process.env.VK_RECOVERY_COOLDOWN_MIN || "10",
);
const vkAutoInstall = process.env.VK_AUTO_INSTALL === "1";

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
  const vkLink = formatHtmlLink(vkBaseUrl, "VK_BASE_URL");
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

function startVibeKanbanProcess() {
  if (vibeKanbanProcess && !vibeKanbanProcess.killed) {
    return;
  }
  const runner = resolveVibeKanbanRunner();
  if (!runner) {
    console.warn("[monitor] vibe-kanban runner unavailable; skipping start");
    return;
  }
  const env = {
    ...process.env,
    PORT: vkRecoveryPort,
    HOST: vkRecoveryHost,
  };
  vibeKanbanProcess = spawn(runner.command, runner.args, {
    env,
    cwd: repoRoot,
    stdio: "ignore",
  });
  vibeKanbanProcess.on("error", (err) => {
    vibeKanbanProcess = null;
    const message = err && err.message ? err.message : String(err);
    console.warn(`[monitor] vibe-kanban spawn failed: ${message}`);
    if (telegramToken && telegramChatId) {
      void sendTelegramMessage(
        `Vibe-kanban spawn failed: ${message}. Check VK_BASE_URL and local npm tooling.`,
      );
    }
  });
  vibeKanbanProcess.on("exit", () => {
    vibeKanbanProcess = null;
  });
}

function resolveVibeKanbanRunner() {
  const localBin = resolve(repoRoot, "node_modules", ".bin");
  const candidates = [
    resolve(localBin, "vibe-kanban"),
    resolve(localBin, "vibe-kanban.cmd"),
  ];
  const localRunner = candidates.find((candidate) => existsSync(candidate));
  if (localRunner) {
    return { command: localRunner, args: [] };
  }

  const npx = spawnSync("npx", ["--version"], { stdio: "ignore" });
  if (npx.status === 0) {
    return { command: "npx", args: ["vibe-kanban"] };
  }

  if (vkAutoInstall) {
    const installed = installWorkspaceDependencies();
    if (installed) {
      return resolveVibeKanbanRunner();
    }
  }

  return null;
}

function installWorkspaceDependencies() {
  const pnpm = spawnSync("pnpm", ["--version"], { stdio: "ignore" });
  if (pnpm.status === 0) {
    const res = spawnSync("pnpm", ["install"], {
      cwd: repoRoot,
      stdio: "inherit",
    });
    return res.status === 0;
  }

  const corepack = spawnSync("corepack", ["--version"], { stdio: "ignore" });
  if (corepack.status === 0) {
    const res = spawnSync("corepack", ["pnpm", "install"], {
      cwd: repoRoot,
      stdio: "inherit",
    });
    return res.status === 0;
  }

  const npm = spawnSync("npm", ["install"], {
    cwd: repoRoot,
    stdio: "inherit",
  });
  return npm.status === 0;
}

async function isVibeKanbanOnline() {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 2000);
  try {
    const res = await fetch(`${vkBaseUrl}/api/projects`, {
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
    return;
  }
  startVibeKanbanProcess();
}

function restartVibeKanbanProcess() {
  if (vibeKanbanProcess && !vibeKanbanProcess.killed) {
    const existing = vibeKanbanProcess;
    try {
      existing.kill("SIGTERM");
    } catch {
      // Best effort.
    }
    const killTimer = setTimeout(() => {
      if (existing && !existing.killed) {
        try {
          existing.kill("SIGKILL");
        } catch {
          // Best effort.
        }
      }
    }, 5000);
    existing.once("exit", () => {
      clearTimeout(killTimer);
      startVibeKanbanProcess();
    });
    return;
  }
  startVibeKanbanProcess();
}

async function triggerVibeKanbanRecovery(reason) {
  const now = Date.now();
  const cooldownMs = vkRecoveryCooldownMin * 60 * 1000;
  if (now - vkRecoveryLastAt < cooldownMs) {
    return;
  }
  vkRecoveryLastAt = now;

  if (telegramToken && telegramChatId) {
    const link = formatHtmlLink(vkBaseUrl, "VK_BASE_URL");
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
  sendUpdate();
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
    await writeFile(analysisPath, String(result), "utf8");
  } catch (err) {
    const analysisPath = logPath.replace(/\.log$/, "-analysis.txt");
    const message = err && err.message ? err.message : String(err);
    await writeFile(analysisPath, `Codex SDK failed: ${message}\n`, "utf8");
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

  if (pendingRestart) {
    pendingRestart = false;
    if (!skipNextAnalyze) {
      const logText = await readFile(logPath, "utf8").catch(() => "");
      const reason = signal ? `signal ${signal}` : `exit ${code}`;
      await analyzeWithCodex(logPath, logText.slice(-15000), reason);
    }
    skipNextAnalyze = false;
    if (!skipNextRestartCount) {
      restartCount += 1;
    }
    skipNextRestartCount = false;
    startProcess();
    return;
  }

  const logText = await readFile(logPath, "utf8").catch(() => "");
  const reason = signal ? `signal ${signal}` : `exit ${code}`;
  await analyzeWithCodex(logPath, logText.slice(-15000), reason);

  if (hasContextWindowError(logText)) {
    await writeFile(
      logPath.replace(/\.log$/, "-context.txt"),
      "Detected context window error. Consider creating a new workspace session and re-sending follow-up.\n",
      "utf8",
    );
  }

  if (maxRestarts > 0 && restartCount >= maxRestarts) {
    return;
  }

  restartCount += 1;
  setTimeout(startProcess, restartDelayMs);
}

async function startProcess() {
  await ensureLogDir();
  const activeLogPath = resolve(logDir, "orchestrator-active.log");
  const archiveLogPath = resolve(logDir, `orchestrator-${nowStamp()}.log`);
  const logStream = await writeFile(activeLogPath, "", "utf8").then(() => null);

  const child = spawn("pwsh", ["-File", scriptPath, ...scriptArgs], {
    stdio: ["ignore", "pipe", "pipe"],
  });
  currentChild = child;

  const append = async (chunk) => {
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
    }, 250);
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

setInterval(() => {
  void flushErrorQueue();
}, 60 * 1000);

startWatcher();
void ensureVibeKanbanRunning();
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
