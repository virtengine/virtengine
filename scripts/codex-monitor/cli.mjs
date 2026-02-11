#!/usr/bin/env node

/**
 * codex-monitor — CLI Entry Point
 *
 * Usage:
 *   codex-monitor                        # start with default config
 *   codex-monitor --setup                # run setup wizard
 *   codex-monitor --args "-MaxParallel 6" # pass orchestrator args
 *   codex-monitor --help                 # show help
 *
 * The CLI handles:
 *   1. First-run detection → auto-launches setup wizard
 *   2. Command routing (setup, help, version, main start)
 *   3. Configuration loading from config.mjs
 */

import { resolve, dirname } from "node:path";
import { existsSync, readFileSync, writeFileSync, unlinkSync, mkdirSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { fork, spawn } from "node:child_process";
import os from "node:os";

const __dirname = dirname(fileURLToPath(import.meta.url));
const args = process.argv.slice(2);

// ── Version (read from package.json — single source of truth) ────────────────

const VERSION = JSON.parse(
  readFileSync(resolve(__dirname, "package.json"), "utf8"),
).version;

// ── Help ─────────────────────────────────────────────────────────────────────

function showHelp() {
  console.log(`
  codex-monitor v${VERSION}
  AI-powered orchestrator supervisor with executor failover, smart PR flow, and Telegram notifications.

  USAGE
    codex-monitor [options]

  COMMANDS
    --setup                     Run the interactive setup wizard
    --help                      Show this help
    --version                   Show version
    --update                    Check for and install latest version
    --no-update-check           Skip automatic update check on startup
    --no-auto-update            Disable background auto-update polling
    --daemon, -d                Run as a background daemon (detached, with PID file)
    --stop-daemon               Stop a running daemon process
    --daemon-status             Check if daemon is running

  ORCHESTRATOR
    --script <path>             Path to the orchestrator script
    --args "<args>"             Arguments passed to the script (default: "-MaxParallel 6")
    --restart-delay <ms>        Delay before restart (default: 10000)
    --max-restarts <n>          Max restarts, 0 = unlimited (default: 0)

  LOGGING
    --log-dir <path>            Log directory (default: ./logs)
    --echo-logs                 Echo raw orchestrator output to console (off by default)
    --quiet, -q                 Only show warnings and errors in terminal
    --verbose, -V               Show debug-level messages in terminal
    --trace                     Show all messages including trace-level
    --log-level <level>         Set explicit log level (trace|debug|info|warn|error|silent)

  AI / CODEX
    --no-codex                  Disable Codex SDK analysis
    --no-autofix                Disable automatic error fixing
    --primary-agent <name>      Override primary agent (codex|copilot|claude)
    --shell, --interactive      Enable interactive shell mode in monitor

  TELEGRAM
    --no-telegram-bot           Disable the interactive Telegram bot
    --telegram-commands         Enable monitor-side Telegram polling (advanced)

  VIBE-KANBAN
    --no-vk-spawn               Don't auto-spawn Vibe-Kanban
    --vk-ensure-interval <ms>   VK health check interval (default: 60000)

  FILE WATCHING
    --no-watch                  Disable file watching for auto-restart
    --watch-path <path>         File to watch (default: script path)

  CONFIGURATION
    --config-dir <path>         Directory containing config files
    --repo-root <path>          Repository root (auto-detected)
    --project-name <name>       Project name for display
    --repo <org/repo>           GitHub repo slug
    --repo-name <name>          Select repository from multi-repo config
    --profile <name>            Environment profile selection
    --mode <name>               Override mode (virtengine/generic)

  ENVIRONMENT
    Configuration is loaded from (in priority order):
    1. CLI flags
    2. Environment variables
    3. .env file
    4. codex-monitor.config.json
    5. Built-in defaults

    Auto-update environment variables:
      CODEX_MONITOR_SKIP_UPDATE_CHECK=1     Disable startup version check
      CODEX_MONITOR_SKIP_AUTO_UPDATE=1      Disable background polling
      CODEX_MONITOR_UPDATE_INTERVAL_MS=N    Override poll interval (default: 600000)

    See .env.example for all environment variables.

  EXECUTOR CONFIG (codex-monitor.config.json)
    {
      "projectName": "my-project",
      "executors": [
        { "name": "copilot-claude", "executor": "COPILOT", "variant": "CLAUDE_OPUS_4_6", "weight": 50, "role": "primary" },
        { "name": "codex-default", "executor": "CODEX", "variant": "DEFAULT", "weight": 50, "role": "backup" }
      ],
      "failover": {
        "strategy": "next-in-line",
        "maxRetries": 3,
        "cooldownMinutes": 5,
        "disableOnConsecutiveFailures": 3
      },
      "distribution": "weighted"
    }

  EXECUTOR ENV SHORTHAND
    EXECUTORS=COPILOT:CLAUDE_OPUS_4_6:50,CODEX:DEFAULT:50

  EXAMPLES
    codex-monitor                                          # start with defaults
    codex-monitor --setup                                  # interactive setup
    codex-monitor --script ./my-orchestrator.ps1            # custom script
    codex-monitor --args "-MaxParallel 4" --no-telegram-bot # custom args
    codex-monitor --no-codex --no-autofix                  # minimal mode

  DOCS
    https://github.com/virtengine/virtengine/tree/main/scripts/codex-monitor
`);
}

// ── Main ─────────────────────────────────────────────────────────────────────

// ── Daemon Mode ──────────────────────────────────────────────────────────────

const PID_FILE = resolve(__dirname, ".cache", "codex-monitor.pid");
const DAEMON_LOG = resolve(__dirname, "logs", "daemon.log");

function getDaemonPid() {
  try {
    if (!existsSync(PID_FILE)) return null;
    const pid = parseInt(readFileSync(PID_FILE, "utf8").trim(), 10);
    if (isNaN(pid)) return null;
    // Check if process is alive
    try { process.kill(pid, 0); return pid; } catch { return null; }
  } catch { return null; }
}

function writePidFile(pid) {
  try {
    mkdirSync(dirname(PID_FILE), { recursive: true });
    writeFileSync(PID_FILE, String(pid), "utf8");
  } catch { /* best effort */ }
}

function removePidFile() {
  try { if (existsSync(PID_FILE)) unlinkSync(PID_FILE); } catch { /* ok */ }
}

function startDaemon() {
  const existing = getDaemonPid();
  if (existing) {
    console.log(`  codex-monitor daemon is already running (PID ${existing})`);
    console.log(`  Use --stop-daemon to stop it first.`);
    process.exit(1);
  }

  // Ensure log directory exists
  try { mkdirSync(dirname(DAEMON_LOG), { recursive: true }); } catch { /* ok */ }

  const child = spawn(process.execPath, [
    "--max-old-space-size=4096",
    fileURLToPath(new URL("./cli.mjs", import.meta.url)),
    ...process.argv.slice(2).filter(a => a !== "--daemon" && a !== "-d"),
    "--daemon-child",
  ], {
    detached: true,
    stdio: "ignore",
    env: { ...process.env, CODEX_MONITOR_DAEMON: "1" },
    cwd: process.cwd(),
  });

  child.unref();
  writePidFile(child.pid);

  console.log(`
  ╭──────────────────────────────────────────────────────────╮
  │ codex-monitor daemon started (PID ${String(child.pid).padEnd(24)}│
  ╰──────────────────────────────────────────────────────────╯

  Logs: ${DAEMON_LOG}
  PID:  ${PID_FILE}

  Commands:
    codex-monitor --daemon-status   Check if running
    codex-monitor --stop-daemon     Stop the daemon
  `);
  process.exit(0);
}

function stopDaemon() {
  const pid = getDaemonPid();
  if (!pid) {
    console.log("  No daemon running (PID file not found or process dead).");
    removePidFile();
    process.exit(0);
  }
  console.log(`  Stopping codex-monitor daemon (PID ${pid})...`);
  try {
    process.kill(pid, "SIGTERM");
    // Wait briefly for graceful shutdown
    let tries = 0;
    const check = () => {
      try { process.kill(pid, 0); } catch { 
        removePidFile();
        console.log("  ✓ Daemon stopped.");
        process.exit(0);
      }
      if (++tries > 10) {
        console.log("  Sending SIGKILL...");
        try { process.kill(pid, "SIGKILL"); } catch { /* ok */ }
        removePidFile();
        console.log("  ✓ Daemon killed.");
        process.exit(0);
      }
      setTimeout(check, 500);
    };
    setTimeout(check, 500);
  } catch (err) {
    console.error(`  Failed to stop daemon: ${err.message}`);
    removePidFile();
    process.exit(1);
  }
}

function daemonStatus() {
  const pid = getDaemonPid();
  if (pid) {
    console.log(`  codex-monitor daemon is running (PID ${pid})`);
  } else {
    console.log("  codex-monitor daemon is not running.");
    removePidFile();
  }
  process.exit(0);
}

async function main() {
  // Handle --help
  if (args.includes("--help") || args.includes("-h")) {
    showHelp();
    process.exit(0);
  }

  // Handle --version
  if (args.includes("--version") || args.includes("-v")) {
    console.log(`codex-monitor v${VERSION}`);
    process.exit(0);
  }

  // Handle --daemon
  if (args.includes("--daemon") || args.includes("-d")) {
    startDaemon();
    return;
  }
  if (args.includes("--stop-daemon")) {
    stopDaemon();
    return;
  }
  if (args.includes("--daemon-status")) {
    daemonStatus();
    return;
  }

  // Write PID file if running as daemon child
  if (args.includes("--daemon-child") || process.env.CODEX_MONITOR_DAEMON === "1") {
    writePidFile(process.pid);
    // Redirect console to log file on daemon child
    const { createWriteStream } = await import("node:fs");
    const logStream = createWriteStream(DAEMON_LOG, { flags: "a" });
    const origStdout = process.stdout.write.bind(process.stdout);
    const origStderr = process.stderr.write.bind(process.stderr);
    process.stdout.write = (chunk, ...a) => { logStream.write(chunk); return origStdout(chunk, ...a); };
    process.stderr.write = (chunk, ...a) => { logStream.write(chunk); return origStderr(chunk, ...a); };
    console.log(`\n[daemon] codex-monitor started at ${new Date().toISOString()} (PID ${process.pid})`);
  }

  // Handle --update (force update)
  if (args.includes("--update")) {
    const { forceUpdate } = await import("./update-check.mjs");
    await forceUpdate(VERSION);
    process.exit(0);
  }

  // ── Startup banner with update check ──────────────────────────────────────
  console.log("");
  console.log("  ╭──────────────────────────────────────────────────────────╮");
  console.log(
    `  │ >_ codex-monitor (v${VERSION})${" ".repeat(Math.max(0, 39 - VERSION.length))}│`,
  );
  console.log("  ╰──────────────────────────────────────────────────────────╯");

  // Non-blocking update check (don't delay startup)
  if (!args.includes("--no-update-check")) {
    import("./update-check.mjs")
      .then(({ checkForUpdate }) => checkForUpdate(VERSION))
      .catch(() => {}); // silent — never block startup
  }

  // Propagate --no-auto-update to env for monitor.mjs to pick up
  if (args.includes("--no-auto-update")) {
    process.env.CODEX_MONITOR_SKIP_AUTO_UPDATE = "1";
  }

  // Handle --setup
  if (args.includes("--setup") || args.includes("setup")) {
    const { runSetup } = await import("./setup.mjs");
    await runSetup();
    process.exit(0);
  }

  // First-run detection
  const { shouldRunSetup } = await import("./setup.mjs");
  if (shouldRunSetup()) {
    console.log("\n  🚀 First run detected — launching setup wizard...\n");
    const { runSetup } = await import("./setup.mjs");
    await runSetup();
    console.log("\n  Setup complete! Starting codex-monitor...\n");
  }

  // Fork monitor as a child process — enables self-restart on source changes.
  // When monitor exits with code 75, cli re-forks with a fresh ESM module cache.
  await runMonitor();
}

// ── Crash notification (last resort — raw fetch when monitor can't start) ─────

function readEnvCredentials() {
  const envPath = resolve(__dirname, ".env");
  if (!existsSync(envPath)) return {};
  const vars = {};
  try {
    const lines = readFileSync(envPath, "utf8").split("\n");
    for (const line of lines) {
      const trimmed = line.trim();
      if (!trimmed || trimmed.startsWith("#")) continue;
      const eqIdx = trimmed.indexOf("=");
      if (eqIdx === -1) continue;
      const key = trimmed.slice(0, eqIdx).trim();
      let val = trimmed.slice(eqIdx + 1).trim();
      if (
        (val.startsWith('"') && val.endsWith('"')) ||
        (val.startsWith("'") && val.endsWith("'"))
      ) {
        val = val.slice(1, -1);
      }
      if (
        key === "TELEGRAM_BOT_TOKEN" ||
        key === "TELEGRAM_CHAT_ID" ||
        key === "PROJECT_NAME"
      ) {
        vars[key] = val;
      }
    }
  } catch {
    // best effort
  }
  return vars;
}

async function sendCrashNotification(exitCode, signal) {
  const env = readEnvCredentials();
  const token = env.TELEGRAM_BOT_TOKEN || process.env.TELEGRAM_BOT_TOKEN;
  const chatId = env.TELEGRAM_CHAT_ID || process.env.TELEGRAM_CHAT_ID;
  if (!token || !chatId) return;

  const project = env.PROJECT_NAME || process.env.PROJECT_NAME || "";
  const host = os.hostname();
  const tag = project ? `[${project}]` : "";
  const reason = signal ? `signal ${signal}` : `exit code ${exitCode}`;
  const text =
    `🔥 *CRASH* ${tag} codex-monitor v${VERSION} died unexpectedly\n` +
    `Host: \`${host}\`\n` +
    `Reason: \`${reason}\`\n` +
    `Time: ${new Date().toISOString()}\n\n` +
    `Monitor is no longer running. Manual restart required.`;

  const url = `https://api.telegram.org/bot${token}/sendMessage`;
  try {
    await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        chat_id: chatId,
        text,
        parse_mode: "Markdown",
      }),
      signal: AbortSignal.timeout(10_000),
    });
  } catch {
    // best effort — if Telegram is unreachable, nothing we can do
  }
}

// ── Self-restart exit code (must match monitor.mjs SELF_RESTART_EXIT_CODE) ───
const SELF_RESTART_EXIT_CODE = 75;
let monitorChild = null;

function runMonitor() {
  return new Promise((resolve, reject) => {
    const monitorPath = fileURLToPath(
      new URL("./monitor.mjs", import.meta.url),
    );
    monitorChild = fork(monitorPath, process.argv.slice(2), {
      stdio: "inherit",
      execArgv: ["--max-old-space-size=4096"],
    });

    monitorChild.on("exit", (code, signal) => {
      monitorChild = null;
      if (code === SELF_RESTART_EXIT_CODE) {
        console.log(
          "\n  \u21BB Monitor source changed \u2014 restarting with fresh modules...\n",
        );
        // Small delay to let file writes settle
        setTimeout(() => resolve(runMonitor()), 1000);
      } else {
        const exitCode = code ?? (signal ? 1 : 0);
        // 4294967295 (0xFFFFFFFF / -1 signed) = OS killed the process (OOM, external termination)
        // Auto-restart after a cooldown instead of treating as a fatal crash
        const isOSKill = exitCode === 4294967295 || exitCode === -1;
        if (isOSKill && !gracefulShutdown) {
          console.error(
            `\n  ⚠ Monitor killed by OS (exit ${exitCode}) — likely OOM. Restarting in 5s...`,
          );
          sendCrashNotification(exitCode, signal).catch(() => {});
          setTimeout(() => resolve(runMonitor()), 5000);
        } else if (exitCode !== 0 && !gracefulShutdown) {
          console.error(
            `\n  ✖ Monitor crashed (${signal ? `signal ${signal}` : `exit code ${exitCode}`}) — sending crash notification...`,
          );
          sendCrashNotification(exitCode, signal).finally(() =>
            process.exit(exitCode),
          );
        } else {
          process.exit(exitCode);
        }
      }
    });

    monitorChild.on("error", (err) => {
      monitorChild = null;
      console.error(`\n  ✖ Monitor failed to start: ${err.message}`);
      sendCrashNotification(1, null).finally(() => reject(err));
    });
  });
}

// Let forked monitor handle signal cleanup — prevent parent from dying first
let gracefulShutdown = false;
process.on("SIGINT", () => {
  gracefulShutdown = true;
  if (!monitorChild) process.exit(0);
  // Child gets SIGINT too via shared terminal — just wait for it to exit
});
process.on("SIGTERM", () => {
  gracefulShutdown = true;
  if (!monitorChild) process.exit(0);
  try {
    monitorChild.kill("SIGTERM");
  } catch {
    /* best effort */
  }
});

main().catch(async (err) => {
  console.error(`codex-monitor failed: ${err.message}`);
  await sendCrashNotification(1, null).catch(() => {});
  process.exit(1);
});
