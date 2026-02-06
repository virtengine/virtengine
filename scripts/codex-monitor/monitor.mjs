import { spawn, spawnSync } from "node:child_process";
import { watch } from "node:fs";
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
const codexEnabled =
  !getFlag("--no-codex") && process.env.CODEX_SDK_DISABLED !== "1";
const repoRoot = resolve(__dirname, "..", "..");
const statusPath = resolve(repoRoot, ".cache", "ve-orchestrator-status.json");
const telegramToken = process.env.TELEGRAM_BOT_TOKEN;
const telegramChatId = process.env.TELEGRAM_CHAT_ID;
const telegramIntervalMin = Number(process.env.TELEGRAM_INTERVAL_MIN || "30");

let CodexClient = null;

let restartCount = 0;
let shuttingDown = false;
let currentChild = null;
let pendingRestart = false;
let skipNextAnalyze = false;
let skipNextRestartCount = false;
let watcher = null;
let watcherDebounce = null;
let watchFileName = null;

const contextPatterns = [
  "ContextWindowExceeded",
  "context window",
  "ran out of room",
];

async function readStatusSummary() {
  try {
    const raw = await readFile(statusPath, "utf8");
    const status = JSON.parse(raw);
    const counts = status.counts || {};
    const completed = Array.isArray(status.completed_tasks)
      ? status.completed_tasks
      : [];
    const recent = completed
      .slice(-5)
      .map((item) => (item.pr_number ? `#${item.pr_number}` : null))
      .filter(Boolean)
      .join(", ");

    const updatedAt = status.updated_at || "unknown";
    const running = counts.running ?? 0;
    const review = counts.review ?? 0;
    const error = counts.error ?? 0;
    const tasksCompleted = status.tasks_completed ?? 0;
    const tasksSubmitted = status.tasks_submitted ?? 0;
    const backlogRemaining = status.backlog_remaining ?? 0;

    const now = Date.now();
    const cutoffMs = 4 * 60 * 60 * 1000;
    const recentWindow = completed.filter((item) => {
      if (!item.completed_at) {
        return false;
      }
      const ts = Date.parse(item.completed_at);
      return Number.isFinite(ts) && now - ts <= cutoffMs;
    });
    const ratePerHour = recentWindow.length / 4;
    const etaHours = ratePerHour > 0 ? backlogRemaining / ratePerHour : null;
    const etaText =
      etaHours && Number.isFinite(etaHours) ? `${etaHours.toFixed(1)}h` : "n/a";

    return [
      "VirtEngine Orchestrator Update",
      `Updated: ${updatedAt}`,
      `Counts: running=${running}, review=${review}, error=${error}`,
      `Tasks: completed=${tasksCompleted}, submitted=${tasksSubmitted}`,
      `Backlog remaining: ${backlogRemaining} (ETA ${etaText})`,
      recent ? `Recent merged PRs: ${recent}` : "Recent merged PRs: none",
    ].join("\n");
  } catch (err) {
    return "VirtEngine Orchestrator Update\nStatus: unavailable (missing status file)";
  }
}

async function sendTelegramMessage(text) {
  if (!telegramToken || !telegramChatId) {
    return;
  }
  const url = `https://api.telegram.org/bot${telegramToken}/sendMessage`;
  const payload = {
    chat_id: telegramChatId,
    text,
  };
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
    const message = await readStatusSummary();
    await sendTelegramMessage(message);
  };
  sendUpdate();
  setInterval(sendUpdate, intervalMs);
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
    if (!CodexClient) {
      CodexClient = await loadCodexSdk();
    }
    if (!CodexClient) {
      throw new Error("Codex SDK not available");
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
  const pnpm = spawnSync("pnpm", ["install"], { cwd, stdio: "inherit" });
  if (pnpm.status === 0) {
    return true;
  }
  const npm = spawnSync("npm", ["install"], { cwd, stdio: "inherit" });
  return npm.status === 0;
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
    await writeFile(activeLogPath, chunk, { flag: "a" });
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

startWatcher();
startProcess();
startTelegramNotifier();
