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
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const args = process.argv.slice(2);

// ── Version ──────────────────────────────────────────────────────────────────

const VERSION = "0.3.0";

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

  ORCHESTRATOR
    --script <path>             Path to the orchestrator script
    --args "<args>"             Arguments passed to the script (default: "-MaxParallel 6")
    --restart-delay <ms>        Delay before restart (default: 10000)
    --max-restarts <n>          Max restarts, 0 = unlimited (default: 0)

  LOGGING
    --log-dir <path>            Log directory (default: ./logs)
    --no-echo-logs              Don't echo orchestrator output to console

  AI / CODEX
    --no-codex                  Disable Codex SDK analysis
    --no-autofix                Disable automatic error fixing

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

  ENVIRONMENT
    Configuration is loaded from (in priority order):
    1. CLI flags
    2. Environment variables
    3. .env file
    4. codex-monitor.config.json
    5. Built-in defaults

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
    https://github.com/virtengine/codex-monitor
`);
}

// ── Main ─────────────────────────────────────────────────────────────────────

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

  // Load and start monitor
  await import("./monitor.mjs");
}

main().catch((err) => {
  console.error(`codex-monitor failed: ${err.message}`);
  process.exit(1);
});
