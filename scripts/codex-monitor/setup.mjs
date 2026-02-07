#!/usr/bin/env node

/**
 * codex-monitor setup wizard
 *
 * Interactive CLI that walks the user through configuring codex-monitor
 * for a new repository. Generates a .env file and validates prerequisites.
 *
 * Usage:
 *   npx codex-monitor setup          # interactive
 *   node setup.mjs                   # same thing
 *   node setup.mjs --non-interactive # use defaults + env vars already set
 */

import { createInterface } from "node:readline";
import { existsSync, readFileSync, writeFileSync } from "node:fs";
import { resolve, dirname } from "node:path";
import { execSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const envPath = resolve(__dirname, ".env");
const envExamplePath = resolve(__dirname, ".env.example");

const isNonInteractive =
  process.argv.includes("--non-interactive") || process.argv.includes("-y");

// ── Helpers ──────────────────────────────────────────────────────────────────

function printBanner() {
  console.log("");
  console.log(
    "  ╔═══════════════════════════════════════════════════════════╗",
  );
  console.log(
    "  ║               Codex Monitor — Setup Wizard               ║",
  );
  console.log(
    "  ╚═══════════════════════════════════════════════════════════╝",
  );
  console.log("");
}

function check(label, ok, hint) {
  const icon = ok ? "✅" : "❌";
  console.log(`  ${icon} ${label}`);
  if (!ok && hint) console.log(`     → ${hint}`);
  return ok;
}

function commandExists(cmd) {
  try {
    execSync(`${process.platform === "win32" ? "where" : "which"} ${cmd}`, {
      stdio: "ignore",
    });
    return true;
  } catch {
    return false;
  }
}

function detectRepoSlug() {
  try {
    const remote = execSync("git remote get-url origin", {
      encoding: "utf8",
      stdio: ["pipe", "pipe", "ignore"],
    }).trim();
    // https://github.com/org/repo.git → org/repo
    const match = remote.match(/github\.com[/:]([^/]+\/[^/.]+)/);
    return match ? match[1] : null;
  } catch {
    return null;
  }
}

function detectRepoRoot() {
  try {
    return execSync("git rev-parse --show-toplevel", {
      encoding: "utf8",
      stdio: ["pipe", "pipe", "ignore"],
    }).trim();
  } catch {
    return null;
  }
}

// ── Interactive prompt ───────────────────────────────────────────────────────

function createPrompt() {
  const rl = createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  return {
    ask(question, defaultValue) {
      return new Promise((resolve) => {
        const suffix = defaultValue ? ` [${defaultValue}]` : "";
        rl.question(`  ${question}${suffix}: `, (answer) => {
          resolve(answer.trim() || defaultValue || "");
        });
      });
    },
    confirm(question, defaultYes = true) {
      return new Promise((resolve) => {
        const hint = defaultYes ? "[Y/n]" : "[y/N]";
        rl.question(`  ${question} ${hint}: `, (answer) => {
          const a = answer.trim().toLowerCase();
          if (!a) resolve(defaultYes);
          else resolve(a === "y" || a === "yes");
        });
      });
    },
    close() {
      rl.close();
    },
  };
}

// ── Main ─────────────────────────────────────────────────────────────────────

async function main() {
  printBanner();

  // ── Prerequisites check ─────────────────────────────────
  console.log("  Checking prerequisites...\n");
  const hasNode = check("Node.js ≥ 18", Number(process.versions.node.split(".")[0]) >= 18);
  const hasGit = check("git", commandExists("git"));
  const hasPwsh = check(
    "PowerShell (pwsh)",
    commandExists("pwsh"),
    "Install: https://github.com/PowerShell/PowerShell",
  );
  const hasGh = check(
    "GitHub CLI (gh)",
    commandExists("gh"),
    "Optional but recommended: https://cli.github.com/",
  );
  console.log("");

  if (!hasNode) {
    console.error("  Node.js 18+ is required. Aborting.");
    process.exit(1);
  }

  const env = {};
  const repoSlug = detectRepoSlug();
  const repoRoot = detectRepoRoot();

  if (isNonInteractive) {
    // Use environment variables or defaults
    env.TELEGRAM_BOT_TOKEN = process.env.TELEGRAM_BOT_TOKEN || "";
    env.TELEGRAM_CHAT_ID = process.env.TELEGRAM_CHAT_ID || "";
    env.VK_BASE_URL = process.env.VK_BASE_URL || "http://127.0.0.1:54089";
    env.VK_RECOVERY_PORT = process.env.VK_RECOVERY_PORT || "54089";
    env.OPENAI_API_KEY = process.env.OPENAI_API_KEY || "";
    env.GITHUB_REPO = process.env.GITHUB_REPO || repoSlug || "";
    env.MAX_PARALLEL = process.env.MAX_PARALLEL || "6";
  } else {
    const prompt = createPrompt();

    try {
      // ── Telegram ──────────────────────────────────────────
      console.log("  ── Telegram Bot Configuration ──\n");
      console.log("  Create a bot via @BotFather on Telegram.\n");
      env.TELEGRAM_BOT_TOKEN = await prompt.ask(
        "Telegram Bot Token",
        process.env.TELEGRAM_BOT_TOKEN,
      );
      if (env.TELEGRAM_BOT_TOKEN) {
        env.TELEGRAM_CHAT_ID = await prompt.ask(
          "Telegram Chat ID (run 'node get-telegram-chat-id.mjs' to find it)",
          process.env.TELEGRAM_CHAT_ID,
        );
      }
      console.log("");

      // ── Vibe-Kanban ───────────────────────────────────────
      console.log("  ── Vibe-Kanban Configuration ──\n");
      env.VK_BASE_URL = await prompt.ask(
        "VK API URL",
        process.env.VK_BASE_URL || "http://127.0.0.1:54089",
      );
      env.VK_RECOVERY_PORT = await prompt.ask(
        "VK port",
        process.env.VK_RECOVERY_PORT || "54089",
      );
      const spawnVk = await prompt.confirm(
        "Auto-spawn vibe-kanban if not running?",
        true,
      );
      if (!spawnVk) env.VK_NO_SPAWN = "1";
      console.log("");

      // ── AI Provider ───────────────────────────────────────
      console.log("  ── AI / Codex Configuration ──\n");
      console.log("  Codex Monitor uses the Codex SDK (OpenAI-compatible).\n");
      env.OPENAI_API_KEY = await prompt.ask(
        "OpenAI API Key (or compatible key)",
        process.env.OPENAI_API_KEY,
      );
      const customBase = await prompt.confirm(
        "Use a custom API base URL? (for Azure, local models, etc.)",
        false,
      );
      if (customBase) {
        env.OPENAI_BASE_URL = await prompt.ask(
          "API Base URL",
          process.env.OPENAI_BASE_URL || "https://api.openai.com/v1",
        );
      }
      const customModel = await prompt.confirm("Use a custom model?", false);
      if (customModel) {
        env.CODEX_MODEL = await prompt.ask(
          "Model name",
          process.env.CODEX_MODEL || "gpt-4o",
        );
      }
      console.log("");

      // ── GitHub ────────────────────────────────────────────
      console.log("  ── GitHub Configuration ──\n");
      env.GITHUB_REPO = await prompt.ask(
        "GitHub repo slug (org/repo)",
        process.env.GITHUB_REPO || repoSlug || "",
      );
      console.log("");

      // ── Orchestrator ──────────────────────────────────────
      console.log("  ── Orchestrator Configuration ──\n");
      env.MAX_PARALLEL = await prompt.ask(
        "Max parallel agent slots",
        process.env.MAX_PARALLEL || "6",
      );
      console.log("");
    } finally {
      prompt.close();
    }
  }

  // ── Write .env file ────────────────────────────────────────
  if (existsSync(envPath)) {
    console.log(`  ⚠️  .env already exists. Writing to .env.generated`);
  }
  const targetPath = existsSync(envPath)
    ? resolve(__dirname, ".env.generated")
    : envPath;

  const lines = [
    "# Generated by codex-monitor setup wizard",
    `# ${new Date().toISOString()}`,
    "",
  ];

  for (const [key, value] of Object.entries(env)) {
    if (value) {
      lines.push(`${key}=${value}`);
    } else {
      lines.push(`# ${key}=`);
    }
  }

  writeFileSync(targetPath, lines.join("\n") + "\n", "utf8");
  console.log(`  ✅ Configuration written to ${targetPath}`);

  // ── Install dependencies ───────────────────────────────────
  console.log("\n  Installing dependencies...\n");
  try {
    if (commandExists("pnpm")) {
      execSync("pnpm install", { cwd: __dirname, stdio: "inherit" });
    } else {
      execSync("npm install", { cwd: __dirname, stdio: "inherit" });
    }
    console.log("\n  ✅ Dependencies installed");
  } catch {
    console.warn("\n  ⚠️  Dependency install failed — run manually:");
    console.warn("     pnpm install  (or)  npm install");
  }

  // ── Summary ────────────────────────────────────────────────
  console.log("\n  ── Setup Complete ──\n");
  console.log("  To start codex-monitor:\n");
  console.log("    node monitor.mjs\n");
  console.log("  Or with arguments:\n");
  console.log(
    '    node monitor.mjs --args "-MaxParallel 6" --restart-delay 10000\n',
  );

  if (!env.TELEGRAM_BOT_TOKEN) {
    console.log(
      "  ℹ️  Telegram is not configured. Add TELEGRAM_BOT_TOKEN to .env for notifications.",
    );
  }
  if (!env.OPENAI_API_KEY) {
    console.log(
      "  ℹ️  No API key set. Codex analysis/autofix will be disabled.",
    );
  }
  console.log("");
}

main().catch((err) => {
  console.error(`Setup failed: ${err.message}`);
  process.exit(1);
});
