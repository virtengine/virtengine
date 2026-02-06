import { spawn, spawnSync } from "node:child_process";
import { watch } from "node:fs";
import { mkdir, readFile, writeFile } from "node:fs/promises";
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

async function ensureLogDir() {
  await mkdir(logDir, { recursive: true });
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
  const logPath = resolve(logDir, `orchestrator-${nowStamp()}.log`);
  const logStream = await writeFile(logPath, "", "utf8").then(() => null);

  const child = spawn("pwsh", ["-File", scriptPath, ...scriptArgs], {
    stdio: ["ignore", "pipe", "pipe"],
  });
  currentChild = child;

  const append = async (chunk) => {
    await writeFile(logPath, chunk, { flag: "a" });
  };

  child.stdout.on("data", (data) => append(data));
  child.stderr.on("data", (data) => append(data));

  child.on("exit", (code, signal) => {
    if (currentChild === child) {
      currentChild = null;
    }
    handleExit(code, signal, logPath);
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
