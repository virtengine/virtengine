#!/usr/bin/env node

/**
 * codex-monitor — Setup Wizard
 *
 * Interactive CLI that configures codex-monitor for a new or existing repository.
 * Handles:
 *   - Prerequisites validation
 *   - Environment file generation (.env + codex-monitor.config.json)
 *   - Executor/model configuration (N executors with weights & failover)
 *   - Multi-repo setup (separate backend/frontend repos)
 *   - Vibe-Kanban auto-wiring (project, repos, executor profiles, agent appends)
 *   - Prompt template scaffolding (.codex-monitor/agents/*.md)
 *   - First-run auto-detection (launches automatically on virgin installs)
 *
 * Usage:
 *   codex-monitor --setup              # interactive wizard
 *   codex-monitor-setup                # same (bin alias)
 *   npx @virtengine/codex-monitor setup
 *   node setup.mjs --non-interactive   # use env vars, skip prompts
 */

import { createInterface } from "node:readline";
import { existsSync, readFileSync, writeFileSync, mkdirSync } from "node:fs";
import { resolve, dirname, basename, relative, isAbsolute } from "node:path";
import { execSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import {
  readCodexConfig,
  getConfigPath,
  hasVibeKanbanMcp,
  auditStreamTimeouts,
  ensureCodexConfig,
  printConfigSummary,
} from "./codex-config.mjs";
import {
  ensureAgentPromptWorkspace,
  getAgentPromptDefinitions,
  PROMPT_WORKSPACE_DIR,
} from "./agent-prompts.mjs";
import {
  buildHookScaffoldOptionsFromEnv,
  normalizeHookTargets,
  scaffoldAgentHookFiles,
} from "./hook-profiles.mjs";

const __dirname = dirname(fileURLToPath(import.meta.url));

const isNonInteractive =
  process.argv.includes("--non-interactive") || process.argv.includes("-y");

// ── Zero-dependency terminal styling (replaces chalk) ────────────────────────
const isTTY = process.stdout.isTTY;
const chalk = {
  bold: (s) => (isTTY ? `\x1b[1m${s}\x1b[22m` : s),
  dim: (s) => (isTTY ? `\x1b[2m${s}\x1b[22m` : s),
  cyan: (s) => (isTTY ? `\x1b[36m${s}\x1b[39m` : s),
  green: (s) => (isTTY ? `\x1b[32m${s}\x1b[39m` : s),
  yellow: (s) => (isTTY ? `\x1b[33m${s}\x1b[39m` : s),
  red: (s) => (isTTY ? `\x1b[31m${s}\x1b[39m` : s),
};

// ── Helpers ──────────────────────────────────────────────────────────────────

function getVersion() {
  try {
    return JSON.parse(readFileSync(resolve(__dirname, "package.json"), "utf8"))
      .version;
  } catch {
    return "0.0.0";
  }
}

function hasSetupMarkers(dir) {
  const markers = [
    ".env",
    "codex-monitor.config.json",
    ".codex-monitor.json",
    "codex-monitor.json",
  ];
  return markers.some((name) => existsSync(resolve(dir, name)));
}

function isPathInside(parent, child) {
  const rel = relative(parent, child);
  return rel === "" || (!rel.startsWith("..") && !isAbsolute(rel));
}

function resolveConfigDir(repoRoot) {
  const explicit = process.env.CODEX_MONITOR_DIR;
  if (explicit) return resolve(explicit);
  const repoPath = resolve(repoRoot || process.cwd());
  const packageDir = resolve(__dirname);
  if (isPathInside(repoPath, packageDir) || hasSetupMarkers(packageDir)) {
    return packageDir;
  }
  const baseDir =
    process.env.APPDATA ||
    process.env.LOCALAPPDATA ||
    process.env.HOME ||
    process.env.USERPROFILE ||
    process.cwd();
  return resolve(baseDir, "codex-monitor");
}

function printBanner() {
  const ver = getVersion();
  const title = `Codex Monitor — Setup Wizard  v${ver}`;
  const pad = Math.max(0, 57 - title.length);
  const left = Math.floor(pad / 2);
  const right = pad - left;
  console.log("");
  console.log(
    "  ╔═══════════════════════════════════════════════════════════════╗",
  );
  console.log(`  ║${" ".repeat(left + 3)}${title}${" ".repeat(right + 3)}║`);
  console.log(
    "  ╚═══════════════════════════════════════════════════════════════╝",
  );
  console.log("");
  console.log(
    chalk.dim("  This wizard will configure codex-monitor for your project."),
  );
  console.log(
    chalk.dim("  Press Enter to accept defaults shown in [brackets]."),
  );
  console.log("");
}

function heading(text) {
  const line = "\u2500".repeat(Math.max(0, 59 - text.length));
  console.log(`\n  ${chalk.bold(text)} ${chalk.dim(line)}\n`);
}

function check(label, ok, hint) {
  const icon = ok ? "✅" : "❌";
  console.log(`  ${icon} ${label}`);
  if (!ok && hint) console.log(`     → ${hint}`);
  return ok;
}

function info(msg) {
  console.log(`  ℹ️  ${msg}`);
}

function success(msg) {
  console.log(`  ✅ ${msg}`);
}

function warn(msg) {
  console.log(`  ⚠️  ${msg}`);
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

/**
 * Check if a binary exists in the package's own node_modules/.bin/.
 * When installed globally, npm only symlinks the top-level package's bin
 * entries to the global path — transitive dependency binaries (like
 * vibe-kanban) live here instead.
 */
function bundledBinExists(cmd) {
  const base = resolve(__dirname, "node_modules", ".bin", cmd);
  return existsSync(base) || existsSync(base + ".cmd");
}

function detectRepoSlug(cwd) {
  try {
    const remote = execSync("git remote get-url origin", {
      encoding: "utf8",
      cwd: cwd || process.cwd(),
      stdio: ["pipe", "pipe", "ignore"],
    }).trim();
    const match = remote.match(/github\.com[/:]([^/]+\/[^/.]+)/);
    return match ? match[1] : null;
  } catch {
    return null;
  }
}

function detectRepoRoot(cwd) {
  try {
    return execSync("git rev-parse --show-toplevel", {
      encoding: "utf8",
      cwd: cwd || process.cwd(),
      stdio: ["pipe", "pipe", "ignore"],
    }).trim();
  } catch {
    return cwd || process.cwd();
  }
}

function detectProjectName(repoRoot) {
  const pkgPath = resolve(repoRoot, "package.json");
  if (existsSync(pkgPath)) {
    try {
      const pkg = JSON.parse(readFileSync(pkgPath, "utf8"));
      if (pkg.name) return pkg.name.replace(/^@[^/]+\//, "");
    } catch {
      /* skip */
    }
  }
  return basename(repoRoot);
}

function getDefaultPromptOverrides() {
  const entries = getAgentPromptDefinitions().map((def) => [
    def.key,
    `${PROMPT_WORKSPACE_DIR}/${def.filename}`,
  ]);
  return Object.fromEntries(entries);
}

function ensureRepoGitIgnoreEntry(repoRoot, entry) {
  const gitignorePath = resolve(repoRoot, ".gitignore");
  const normalizedEntry = String(entry || "").trim();
  if (!normalizedEntry) return false;

  let existing = "";
  if (existsSync(gitignorePath)) {
    existing = readFileSync(gitignorePath, "utf8");
  }

  const hasEntry = existing
    .split(/\r?\n/)
    .map((line) => line.trim())
    .includes(normalizedEntry);
  if (hasEntry) return false;

  const next =
    existing.endsWith("\n") || !existing ? existing : `${existing}\n`;
  writeFileSync(gitignorePath, `${next}${normalizedEntry}\n`, "utf8");
  return true;
}

function parseHookCommandInput(rawValue) {
  const raw = String(rawValue || "").trim();
  if (!raw) return null;
  const lowered = raw.toLowerCase();
  if (["none", "off", "disable", "disabled"].includes(lowered)) {
    return [];
  }
  return raw
    .split(/\s*;;\s*|\r?\n/)
    .map((part) => part.trim())
    .filter(Boolean);
}

function printHookScaffoldSummary(result) {
  if (!result || !result.enabled) {
    info("Agent hook scaffolding disabled.");
    return;
  }

  const totalChanged = result.written.length + result.updated.length;
  if (totalChanged > 0) {
    success(`Configured ${totalChanged} agent hook file(s).`);
  } else {
    info("Agent hook files already existed — no file changes needed.");
  }

  if (result.written.length > 0) {
    for (const path of result.written) {
      console.log(`    + ${path}`);
    }
  }
  if (result.updated.length > 0) {
    for (const path of result.updated) {
      console.log(`    ~ ${path}`);
    }
  }
  if (result.skipped.length > 0) {
    for (const path of result.skipped) {
      console.log(`    = ${path} (kept existing)`);
    }
  }
  if (result.warnings.length > 0) {
    for (const warning of result.warnings) {
      warn(warning);
    }
  }
}

// ── Prompt System ────────────────────────────────────────────────────────────

function createPrompt() {
  // Fix for Windows PowerShell readline issues
  // Only use terminal mode if stdin is actually a TTY
  // This prevents both double-echo and output duplication
  const rl = createInterface({
    input: process.stdin,
    output: process.stdout,
    terminal: process.stdin.isTTY && process.stdout.isTTY,
  });

  return {
    ask(question, defaultValue) {
      return new Promise((res) => {
        const suffix = defaultValue ? ` [${defaultValue}]` : "";
        rl.question(`  ${question}${suffix}: `, (answer) => {
          res(answer.trim() || defaultValue || "");
        });
      });
    },
    confirm(question, defaultYes = true) {
      return new Promise((res) => {
        const hint = defaultYes ? "[Y/n]" : "[y/N]";
        rl.question(`  ${question} ${hint}: `, (answer) => {
          const a = answer.trim().toLowerCase();
          if (!a) res(defaultYes);
          else res(a === "y" || a === "yes");
        });
      });
    },
    choose(question, options, defaultIdx = 0) {
      return new Promise((res) => {
        console.log(`  ${question}`);
        options.forEach((opt, i) => {
          const marker = i === defaultIdx ? "→" : " ";
          console.log(`  ${marker} ${i + 1}) ${opt}`);
        });
        rl.question(`  Choice [${defaultIdx + 1}]: `, (answer) => {
          const idx = answer.trim() ? Number(answer.trim()) - 1 : defaultIdx;
          res(Math.max(0, Math.min(idx, options.length - 1)));
        });
      });
    },
    close() {
      rl.close();
    },
  };
}

// ── Executor Templates ───────────────────────────────────────────────────────

const EXECUTOR_PRESETS = {
  "copilot-codex": [
    {
      name: "copilot-claude",
      executor: "COPILOT",
      variant: "CLAUDE_OPUS_4_6",
      weight: 50,
      role: "primary",
    },
    {
      name: "codex-default",
      executor: "CODEX",
      variant: "DEFAULT",
      weight: 50,
      role: "backup",
    },
  ],
  "copilot-only": [
    {
      name: "copilot-claude",
      executor: "COPILOT",
      variant: "CLAUDE_OPUS_4_6",
      weight: 100,
      role: "primary",
    },
  ],
  "codex-only": [
    {
      name: "codex-default",
      executor: "CODEX",
      variant: "DEFAULT",
      weight: 100,
      role: "primary",
    },
  ],
  triple: [
    {
      name: "copilot-claude",
      executor: "COPILOT",
      variant: "CLAUDE_OPUS_4_6",
      weight: 40,
      role: "primary",
    },
    {
      name: "codex-default",
      executor: "CODEX",
      variant: "DEFAULT",
      weight: 35,
      role: "backup",
    },
    {
      name: "copilot-gpt",
      executor: "COPILOT",
      variant: "GPT_4_1",
      weight: 25,
      role: "tertiary",
    },
  ],
};

const FAILOVER_STRATEGIES = [
  {
    name: "next-in-line",
    desc: "Use the next executor by role priority (primary → backup → tertiary)",
  },
  {
    name: "weighted-random",
    desc: "Randomly select from remaining executors by weight",
  },
  { name: "round-robin", desc: "Cycle through remaining executors evenly" },
];

const DISTRIBUTION_MODES = [
  {
    name: "weighted",
    desc: "Distribute tasks by configured weight percentages",
  },
  { name: "round-robin", desc: "Alternate between executors equally" },
  {
    name: "primary-only",
    desc: "Always use primary; others only for failover",
  },
];

const SETUP_PROFILES = [
  {
    key: "recommended",
    label: "Recommended — configure important choices, keep safe defaults",
  },
  {
    key: "advanced",
    label: "Advanced — full control over all setup options",
  },
];

function toPositiveInt(value, fallback) {
  const n = Number(value);
  if (!Number.isFinite(n) || n <= 0) return fallback;
  return Math.round(n);
}

function normalizeEnum(value, allowed, fallback) {
  const normalized = String(value || "")
    .trim()
    .toLowerCase();
  return allowed.includes(normalized) ? normalized : fallback;
}

function parseBooleanEnvValue(value, fallback = false) {
  if (value === undefined || value === null || value === "") {
    return fallback;
  }
  const normalized = String(value).trim().toLowerCase();
  if (["1", "true", "yes", "on", "y"].includes(normalized)) {
    return true;
  }
  if (["0", "false", "no", "off", "n"].includes(normalized)) {
    return false;
  }
  return fallback;
}

function toBooleanEnvString(value, fallback = false) {
  return parseBooleanEnvValue(value, fallback) ? "true" : "false";
}

function normalizeSetupConfiguration({ env, configJson, repoRoot, slug }) {
  env.PROJECT_NAME =
    env.PROJECT_NAME || configJson.projectName || basename(repoRoot);
  env.REPO_ROOT = env.REPO_ROOT || repoRoot;
  env.GITHUB_REPO = env.GITHUB_REPO || slug || "";

  env.MAX_PARALLEL = String(toPositiveInt(env.MAX_PARALLEL || "6", 6));
  env.TELEGRAM_INTERVAL_MIN = String(
    toPositiveInt(env.TELEGRAM_INTERVAL_MIN || "10", 10),
  );

  env.KANBAN_BACKEND = normalizeEnum(
    env.KANBAN_BACKEND,
    ["vk", "github", "jira"],
    "vk",
  );
  env.EXECUTOR_MODE = normalizeEnum(
    env.EXECUTOR_MODE,
    ["internal", "vk", "hybrid"],
    "internal",
  );

  env.VK_BASE_URL = env.VK_BASE_URL || "http://127.0.0.1:54089";
  env.VK_RECOVERY_PORT = String(
    toPositiveInt(env.VK_RECOVERY_PORT || "54089", 54089),
  );

  env.CODEX_TRANSPORT = normalizeEnum(
    env.CODEX_TRANSPORT || process.env.CODEX_TRANSPORT,
    ["sdk", "auto", "cli"],
    "sdk",
  );
  env.COPILOT_TRANSPORT = normalizeEnum(
    env.COPILOT_TRANSPORT || process.env.COPILOT_TRANSPORT,
    ["sdk", "auto", "cli", "url"],
    "sdk",
  );
  env.CLAUDE_TRANSPORT = normalizeEnum(
    env.CLAUDE_TRANSPORT || process.env.CLAUDE_TRANSPORT,
    ["sdk", "auto", "cli"],
    "sdk",
  );

  env.WHATSAPP_ENABLED = toBooleanEnvString(env.WHATSAPP_ENABLED, false);

  env.CONTAINER_ENABLED = toBooleanEnvString(env.CONTAINER_ENABLED, false);

  env.CONTAINER_RUNTIME = normalizeEnum(
    env.CONTAINER_RUNTIME,
    ["auto", "docker", "podman", "container"],
    "auto",
  );

  if (
    !Array.isArray(configJson.executors) ||
    configJson.executors.length === 0
  ) {
    configJson.executors = EXECUTOR_PRESETS["codex-only"];
  }
  configJson.executors = configJson.executors.map((executor, index) => ({
    ...executor,
    name: executor.name || `executor-${index + 1}`,
    executor: String(executor.executor || "CODEX").toUpperCase(),
    variant: executor.variant || "DEFAULT",
    weight: toPositiveInt(executor.weight || 1, 1),
    role:
      executor.role ||
      (index === 0
        ? "primary"
        : index === 1
          ? "backup"
          : `executor-${index + 1}`),
    enabled: executor.enabled !== false,
  }));

  configJson.failover = {
    strategy: normalizeEnum(
      configJson.failover?.strategy || env.FAILOVER_STRATEGY || "next-in-line",
      ["next-in-line", "weighted-random", "round-robin"],
      "next-in-line",
    ),
    maxRetries: toPositiveInt(
      configJson.failover?.maxRetries || env.FAILOVER_MAX_RETRIES || 3,
      3,
    ),
    cooldownMinutes: toPositiveInt(
      configJson.failover?.cooldownMinutes || env.FAILOVER_COOLDOWN_MIN || 5,
      5,
    ),
    disableOnConsecutiveFailures: toPositiveInt(
      configJson.failover?.disableOnConsecutiveFailures ||
        env.FAILOVER_DISABLE_AFTER ||
        3,
      3,
    ),
  };

  configJson.distribution = normalizeEnum(
    configJson.distribution || env.EXECUTOR_DISTRIBUTION || "weighted",
    ["weighted", "round-robin", "primary-only"],
    "weighted",
  );

  if (
    !Array.isArray(configJson.repositories) ||
    configJson.repositories.length === 0
  ) {
    configJson.repositories = [
      {
        name: basename(repoRoot),
        slug: env.GITHUB_REPO,
        primary: true,
      },
    ];
  }

  configJson.projectName = env.PROJECT_NAME;
  configJson.kanban = { backend: env.KANBAN_BACKEND };
  configJson.internalExecutor = {
    ...(configJson.internalExecutor || {}),
    mode: env.EXECUTOR_MODE,
  };
}

function formatEnvValue(value) {
  const raw = String(value ?? "");
  const needsQuotes = /\s|#|=/.test(raw);
  if (!needsQuotes) return raw;
  return `"${raw.replace(/"/g, '\\"')}"`;
}

export function buildStandardizedEnvFile(templateText, envEntries) {
  const lines = templateText.split(/\r?\n/);
  const entryMap = new Map(
    Object.entries(envEntries)
      .filter(([key]) => !key.startsWith("_"))
      .map(([key, value]) => [key, String(value ?? "")]),
  );

  const consumed = new Set();
  const seenKeys = new Set();
  const updated = lines.flatMap((line) => {
    const match = line.match(/^\s*#?\s*([A-Z0-9_]+)=.*$/);
    if (!match) return [line];
    const key = match[1];
    if (seenKeys.has(key)) return [];
    seenKeys.add(key);
    if (!entryMap.has(key)) return [line];
    consumed.add(key);
    return [`${key}=${formatEnvValue(entryMap.get(key))}`];
  });

  const extras = [...entryMap.keys()].filter((key) => !consumed.has(key));
  if (extras.length > 0) {
    updated.push("");
    updated.push("# Added by setup wizard");
    for (const key of extras.sort()) {
      updated.push(`${key}=${formatEnvValue(entryMap.get(key))}`);
    }
  }

  const header = [
    "# Generated by codex-monitor setup wizard",
    `# ${new Date().toISOString()}`,
    "",
  ];
  return [...header, ...updated].join("\n") + "\n";
}

// ── Agent Template ───────────────────────────────────────────────────────────

function generateAgentsMd(projectName, repoSlug) {
  return `# ${projectName} — Agent Guide

## CRITICAL

Always work on tasks longer than you think are needed to accommodate edge cases, testing, and quality.
Ensure tests pass and build is clean with 0 warnings before deciding a task is complete.
When working on a task, do not stop until it is COMPLETELY done end-to-end.

Before finishing a task — create a commit using conventional commits and push.

### PR Creation

After committing:
- Run \`gh pr create\` to open the PR
- Ensure pre-push hooks pass
- Fix any lint or test errors encountered

## Overview

- Repository: \`${repoSlug}\`
- Task management: Vibe-Kanban (auto-configured by codex-monitor)

## Build & Test

\`\`\`bash
# Add your build commands here
npm run build
npm test
\`\`\`

## Commit Conventions

Use [Conventional Commits](https://www.conventionalcommits.org/):

\`\`\`
type(scope): description
\`\`\`

Valid types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert

## Pre-commit / Pre-push

Linting and formatting are enforced before commit.
Tests and builds are verified before push.
`;
}

// ── VK Auto-Configuration ────────────────────────────────────────────────────

function generateVkSetupScript(config) {
  const repoRoot = config.repoRoot.replace(/\\/g, "/");
  const monitorDir = config.monitorDir.replace(/\\/g, "/");

  return `#!/usr/bin/env bash
# Auto-generated by codex-monitor setup
# VK workspace setup script for: ${config.projectName}

set -euo pipefail

echo "Setting up workspace for ${config.projectName}..."

# ── PATH propagation ──────────────────────────────────────────────────────────
# Ensure common tool directories are on PATH so agents can find gh, pwsh, node,
# go, etc. without using full absolute paths. The host user's PATH may not be
# inherited by the workspace shell.
_add_to_path() { case ":\$PATH:" in *":\$1:"*) ;; *) export PATH="\$1:\$PATH" ;; esac; }

for _dir in \\
  /usr/local/bin \\
  /usr/local/sbin \\
  /usr/bin \\
  "\$HOME/.local/bin" \\
  "\$HOME/bin" \\
  "\$HOME/go/bin" \\
  "\$HOME/.cargo/bin" \\
  /snap/bin \\
  /opt/homebrew/bin; do
  [ -d "\$_dir" ] && _add_to_path "\$_dir"
done

# Windows-specific paths (Git Bash / MSYS2 environment)
case "\$(uname -s 2>/dev/null)" in
  MINGW*|MSYS*|CYGWIN*)
    for _wdir in \\
      "/c/Program Files/GitHub CLI" \\
      "/c/Program Files/PowerShell/7" \\
      "/c/Program Files/nodejs"; do
      [ -d "\$_wdir" ] && _add_to_path "\$_wdir"
    done
    ;;
esac

# ── Git credential guard ─────────────────────────────────────────────────────
# NEVER run 'gh auth setup-git' inside a workspace — it writes the container's
# gh path into .git/config, corrupting pushes from other environments.
# Rely on GH_TOKEN/GITHUB_TOKEN env vars or the global credential helper.
if git config --local credential.helper &>/dev/null; then
  _local_helper=\$(git config --local credential.helper)
  if echo "\$_local_helper" | grep -qE '/home/.*/gh(\\.exe)?|/tmp/.*/gh'; then
    echo "  [setup] Removing stale local credential.helper: \$_local_helper"
    git config --local --unset credential.helper || true
  fi
fi

# ── Git worktree cleanup ─────────────────────────────────────────────────────
# Prune stale worktree references to prevent path corruption errors.
# This happens when worktree directories are deleted but git metadata remains.
if [ -f ".git" ]; then
  _gitdir=\$(cat .git | sed 's/^gitdir: //')
  _repo_root=\$(dirname "\$_gitdir" | xargs dirname | xargs dirname)
  if [ -d "\$_repo_root/.git/worktrees" ]; then
    echo "  [setup] Pruning stale worktrees..."
    ( cd "\$_repo_root" && git worktree prune -v 2>&1 | sed 's/^/  [prune] /' ) || true
  fi
fi

# ── GitHub auth verification ─────────────────────────────────────────────────
if command -v gh &>/dev/null; then
  echo "  [setup] gh CLI found at: \$(command -v gh)"
  gh auth status 2>/dev/null || echo "  [setup] gh not authenticated — ensure GH_TOKEN is set"
else
  echo "  [setup] WARNING: gh CLI not found on PATH"
  echo "  [setup] Current PATH: \$PATH"
fi

# Install dependencies
if [ -f "package.json" ]; then
  if command -v pnpm &>/dev/null; then
    pnpm install
  elif command -v npm &>/dev/null; then
    npm install
  fi
fi

# Install codex-monitor dependencies
if [ -d "${relative(config.repoRoot, monitorDir)}" ]; then
  cd "${relative(config.repoRoot, monitorDir)}"
  if command -v pnpm &>/dev/null; then
    pnpm install
  elif command -v npm &>/dev/null; then
    npm install
  fi
  cd -
fi

echo "Workspace setup complete."
`;
}

function generateVkCleanupScript(config) {
  return `#!/usr/bin/env bash
# Auto-generated by codex-monitor setup
# VK workspace cleanup script for: ${config.projectName}

set -euo pipefail

echo "Cleaning up workspace for ${config.projectName}..."

# Create PR if branch has commits
BRANCH=$(git branch --show-current 2>/dev/null || true)
if [ -n "$BRANCH" ] && [ "$BRANCH" != "main" ] && [ "$BRANCH" != "master" ]; then
  COMMITS=$(git log main.."$BRANCH" --oneline 2>/dev/null | wc -l || echo 0)
  if [ "$COMMITS" -gt 0 ]; then
    echo "Branch $BRANCH has $COMMITS commit(s) — creating PR..."
    gh pr create --fill 2>/dev/null || echo "PR creation skipped"
  fi
fi

echo "Cleanup complete."
`;
}

// ── Main Setup Flow ──────────────────────────────────────────────────────────

async function main() {
  printBanner();

  // ── Step 1: Prerequisites ───────────────────────────────
  heading("Step 1 of 9 — Prerequisites");
  const hasNode = check(
    "Node.js ≥ 18",
    Number(process.versions.node.split(".")[0]) >= 18,
  );
  const hasGit = check("git", commandExists("git"));
  check(
    "PowerShell (pwsh)",
    commandExists("pwsh"),
    "Install: https://github.com/PowerShell/PowerShell",
  );
  check(
    "GitHub CLI (gh)",
    commandExists("gh"),
    "Recommended: https://cli.github.com/",
  );
  const hasVk = check(
    "Vibe-Kanban CLI",
    commandExists("vibe-kanban") || bundledBinExists("vibe-kanban"),
    "Bundled with @virtengine/codex-monitor as a dependency",
  );

  if (!hasVk) {
    warn(
      "vibe-kanban not found. This is bundled with codex-monitor, so this is unexpected.",
    );
    info("Try reinstalling:");
    console.log("     npm uninstall -g @virtengine/codex-monitor");
    console.log("     npm install -g @virtengine/codex-monitor\n");
  }

  if (!hasNode) {
    console.error("\n  Node.js 18+ is required. Aborting.\n");
    process.exit(1);
  }

  const repoRoot = detectRepoRoot();
  const configDir = resolveConfigDir(repoRoot);
  const slug = detectRepoSlug();
  const projectName = detectProjectName(repoRoot);

  const env = {};
  const configJson = {
    projectName,
    executors: [],
    failover: {},
    distribution: "weighted",
    repositories: [],
    agentPrompts: {},
  };

  env.REPO_ROOT = process.env.REPO_ROOT || repoRoot;

  if (isNonInteractive) {
    return runNonInteractive({
      env,
      configJson,
      repoRoot,
      slug,
      projectName,
      configDir,
    });
  }

  const prompt = createPrompt();

  try {
    // ── Step 2: Setup Mode + Project Identity ─────────────
    heading("Step 2 of 9 — Setup Mode & Project Identity");
    const setupProfileIdx = await prompt.choose(
      "How much setup detail do you want?",
      SETUP_PROFILES.map((profile) => profile.label),
      0,
    );
    const setupProfile = SETUP_PROFILES[setupProfileIdx]?.key || "recommended";
    const isAdvancedSetup = setupProfile === "advanced";
    info(
      isAdvancedSetup
        ? "Advanced mode enabled — all sections will prompt for detailed overrides."
        : "Recommended mode enabled — only key decisions are prompted; safe defaults fill the rest.",
    );

    env.PROJECT_NAME = await prompt.ask("Project name", projectName);
    env.GITHUB_REPO = await prompt.ask(
      "GitHub repo slug (org/repo)",
      process.env.GITHUB_REPO || slug || "",
    );
    configJson.projectName = env.PROJECT_NAME;

    // ── Step 3: Repository ─────────────────────────────────
    heading("Step 3 of 9 — Repository Configuration");
    const multiRepo = isAdvancedSetup
      ? await prompt.confirm(
          "Do you have multiple repositories (e.g. separate backend/frontend)?",
          false,
        )
      : false;

    if (multiRepo) {
      info("Configure each repository. The first is the primary.\n");
      let addMore = true;
      let repoIdx = 0;
      while (addMore) {
        const repoName = await prompt.ask(
          `  Repo ${repoIdx + 1} — name`,
          repoIdx === 0 ? basename(repoRoot) : "",
        );
        const repoPath = await prompt.ask(
          `  Repo ${repoIdx + 1} — local path`,
          repoIdx === 0 ? repoRoot : "",
        );
        const repoSlug = await prompt.ask(
          `  Repo ${repoIdx + 1} — GitHub slug`,
          repoIdx === 0 ? env.GITHUB_REPO : "",
        );
        configJson.repositories.push({
          name: repoName,
          path: repoPath,
          slug: repoSlug,
          primary: repoIdx === 0,
        });
        repoIdx++;
        addMore = await prompt.confirm("Add another repository?", false);
      }
    } else {
      // Single-repo: omit path — config.mjs auto-detects via git
      configJson.repositories.push({
        name: basename(repoRoot),
        slug: env.GITHUB_REPO,
        primary: true,
      });
      if (!isAdvancedSetup) {
        info(
          "Using single-repo defaults (recommended mode). Re-run setup in Advanced mode for multi-repo config.",
        );
      }
    }

    // ── Step 4: Executor Configuration ─────────────────────
    heading("Step 4 of 9 — Executor / Agent Configuration");
    console.log("  Executors are the AI agents that work on tasks.\n");

    const presetOptions = isAdvancedSetup
      ? [
          "Codex only",
          "Copilot + Codex (50/50 split)",
          "Copilot only (Claude Opus 4.6)",
          "Triple (Copilot Claude 40%, Codex 35%, Copilot GPT 25%)",
          "Custom — I'll define my own executors",
        ]
      : [
          "Codex only",
          "Copilot + Codex (50/50 split)",
          "Copilot only (Claude Opus 4.6)",
          "Triple (Copilot Claude 40%, Codex 35%, Copilot GPT 25%)",
        ];

    const presetIdx = await prompt.choose(
      "Select executor preset:",
      presetOptions,
      0,
    );

    const presetNames = isAdvancedSetup
      ? ["codex-only", "copilot-codex", "copilot-only", "triple", "custom"]
      : ["codex-only", "copilot-codex", "copilot-only", "triple"];
    const presetKey = presetNames[presetIdx] || "codex-only";

    if (presetKey === "custom") {
      info("Define your executors. Enter empty name to finish.\n");
      let execIdx = 0;
      const roles = ["primary", "backup", "tertiary"];
      while (true) {
        const eName = await prompt.ask(
          `  Executor ${execIdx + 1} — name (empty to finish)`,
          "",
        );
        if (!eName) break;
        const eType = await prompt.ask("  Executor type", "COPILOT");
        const eVariant = await prompt.ask("  Model variant", "CLAUDE_OPUS_4_6");
        const eWeight = Number(await prompt.ask("  Weight (1-100)", "50"));
        configJson.executors.push({
          name: eName,
          executor: eType.toUpperCase(),
          variant: eVariant,
          weight: eWeight,
          role: roles[execIdx] || `executor-${execIdx + 1}`,
          enabled: true,
        });
        execIdx++;
      }
    } else {
      configJson.executors = EXECUTOR_PRESETS[presetKey];
    }

    // Show executor summary
    console.log("\n  Configured executors:");
    const totalWeight = configJson.executors.reduce((s, e) => s + e.weight, 0);
    for (const e of configJson.executors) {
      const pct = Math.round((e.weight / totalWeight) * 100);
      console.log(
        `    ${e.role.padEnd(10)} ${e.executor}:${e.variant} — ${pct}%`,
      );
    }

    if (isAdvancedSetup) {
      console.log();
      console.log(
        chalk.dim("  What happens when an executor fails repeatedly?"),
      );
      console.log();

      const failoverIdx = await prompt.choose(
        "Select failover strategy:",
        FAILOVER_STRATEGIES.map((f) => `${f.name} — ${f.desc}`),
        0,
      );
      configJson.failover = {
        strategy: FAILOVER_STRATEGIES[failoverIdx].name,
        maxRetries: Number(
          await prompt.ask("Max retries before failover", "3"),
        ),
        cooldownMinutes: Number(
          await prompt.ask("Cooldown after disabling executor (minutes)", "5"),
        ),
        disableOnConsecutiveFailures: Number(
          await prompt.ask(
            "Disable executor after N consecutive failures",
            "3",
          ),
        ),
      };

      const distIdx = await prompt.choose(
        "\n  Task distribution mode:",
        DISTRIBUTION_MODES.map((d) => `${d.name} — ${d.desc}`),
        0,
      );
      configJson.distribution = DISTRIBUTION_MODES[distIdx].name;
    } else {
      configJson.failover = {
        strategy: "next-in-line",
        maxRetries: 3,
        cooldownMinutes: 5,
        disableOnConsecutiveFailures: 3,
      };
      configJson.distribution = "weighted";
      info(
        "Using recommended routing defaults: weighted distribution, next-in-line failover.",
      );
    }

    // ── Step 5: AI Provider ────────────────────────────────
    heading("Step 5 of 9 — AI / Codex Provider");
    console.log(
      "  Codex Monitor uses the Codex SDK for crash analysis & autofix.\n",
    );

    const providerIdx = await prompt.choose(
      "Select AI provider:",
      [
        "OpenAI (default)",
        "Azure OpenAI",
        "Local model (Ollama, vLLM, etc.)",
        "Other OpenAI-compatible endpoint",
        "None — disable AI features",
      ],
      0,
    );

    if (providerIdx < 4) {
      env.OPENAI_API_KEY = await prompt.ask(
        "API Key",
        process.env.OPENAI_API_KEY || "",
      );
    }
    if (providerIdx === 1) {
      env.OPENAI_BASE_URL = await prompt.ask(
        "Azure endpoint URL",
        process.env.OPENAI_BASE_URL || "",
      );
      env.CODEX_MODEL = await prompt.ask(
        "Deployment/model name",
        process.env.CODEX_MODEL || "",
      );
    } else if (providerIdx === 2) {
      env.OPENAI_API_KEY = env.OPENAI_API_KEY || "ollama";
      env.OPENAI_BASE_URL = await prompt.ask(
        "Local API URL",
        "http://localhost:11434/v1",
      );
      env.CODEX_MODEL = await prompt.ask("Model name", "codex");
    } else if (providerIdx === 3) {
      env.OPENAI_BASE_URL = await prompt.ask("API Base URL", "");
      env.CODEX_MODEL = await prompt.ask("Model name", "");
    } else if (providerIdx === 4) {
      env.CODEX_SDK_DISABLED = "true";
    }

    // ── Step 6: Telegram ──────────────────────────────────
    heading("Step 6 of 9 — Telegram Notifications");
    console.log(
      "  The Telegram bot sends real-time notifications and lets you\n" +
        "  control the orchestrator via /status, /tasks, /restart, etc.\n",
    );

    const wantTelegram = await prompt.confirm(
      "Set up Telegram notifications?",
      true,
    );
    if (wantTelegram) {
      // Step 1: Create bot
      console.log(
        "\n" +
          chalk.bold("Step 1: Create Your Bot") +
          chalk.dim(" (if you haven't already)"),
      );
      console.log(
        "  1. Open Telegram and search for " + chalk.cyan("@BotFather"),
      );
      console.log("  2. Send: " + chalk.cyan("/newbot"));
      console.log("  3. Choose a display name (e.g., 'MyProject Monitor')");
      console.log(
        "  4. Choose a username ending in 'bot' (e.g., 'myproject_monitor_bot')",
      );
      console.log("  5. Copy the bot token BotFather gives you");
      console.log();

      const hasBotReady = await prompt.confirm(
        "Have you created your bot and have the token ready?",
        false,
      );

      if (!hasBotReady) {
        warn("No problem! You can set up Telegram later by:");
        console.log("  1. Adding TELEGRAM_BOT_TOKEN to .env");
        console.log("  2. Adding TELEGRAM_CHAT_ID to .env");
        console.log("  3. Or re-running: codex-monitor --setup");
        console.log();
      } else {
        // Step 2: Get bot token
        console.log("\n" + chalk.bold("Step 2: Enter Your Bot Token"));
        console.log(
          chalk.dim(
            "  Looks like: 1234567890:ABCdefGHIjklMNOpqrsTUVwxyz-1234567890",
          ),
        );
        console.log();

        env.TELEGRAM_BOT_TOKEN = await prompt.ask(
          "Bot Token",
          process.env.TELEGRAM_BOT_TOKEN || "",
        );

        if (env.TELEGRAM_BOT_TOKEN && env.TELEGRAM_BOT_TOKEN.length > 20) {
          // Validate token format
          const tokenValid = /^\d+:[A-Za-z0-9_-]+$/.test(
            env.TELEGRAM_BOT_TOKEN,
          );
          if (!tokenValid) {
            warn(
              "Token format looks incorrect. Make sure you copied the full token from BotFather.",
            );
          } else {
            info("✓ Token format looks good");
          }

          // Step 3: Get chat ID
          console.log("\n" + chalk.bold("Step 3: Get Your Chat ID"));
          console.log("  Your chat ID tells the bot where to send messages.");
          console.log();

          const knowsChatId = await prompt.confirm(
            "Do you already know your chat ID?",
            false,
          );

          if (knowsChatId) {
            env.TELEGRAM_CHAT_ID = await prompt.ask(
              "Chat ID (numeric, e.g., 123456789)",
              process.env.TELEGRAM_CHAT_ID || "",
            );
          } else {
            // Guide user to get chat ID
            console.log("\n" + chalk.cyan("To get your chat ID:") + "\n");
            console.log(
              "  1. Open Telegram and search for your bot's username",
            );
            console.log(
              "  2. Click " +
                chalk.cyan("START") +
                " or send any message (e.g., 'Hello')",
            );
            console.log("  3. Come back here and we'll detect your chat ID");
            console.log();

            const ready = await prompt.confirm(
              "Ready? (I've messaged my bot)",
              false,
            );

            if (ready) {
              // Try to fetch chat ID from Telegram API
              info("Fetching your chat ID from Telegram...");
              try {
                const response = await fetch(
                  `https://api.telegram.org/bot${env.TELEGRAM_BOT_TOKEN}/getUpdates`,
                );
                const data = await response.json();

                if (data.ok && data.result && data.result.length > 0) {
                  // Find the most recent message
                  const latestMessage = data.result[data.result.length - 1];
                  const chatId = latestMessage?.message?.chat?.id;

                  if (chatId) {
                    env.TELEGRAM_CHAT_ID = String(chatId);
                    info(`✓ Found your chat ID: ${chatId}`);
                    console.log();
                  } else {
                    warn(
                      "Couldn't find a chat ID. Make sure you sent a message to your bot.",
                    );
                    env.TELEGRAM_CHAT_ID = await prompt.ask(
                      "Enter chat ID manually",
                      "",
                    );
                  }
                } else {
                  warn(
                    "No messages found. Make sure you sent a message to your bot first.",
                  );
                  console.log(
                    chalk.dim(
                      "  Or run: codex-monitor-chat-id (after starting the bot)",
                    ),
                  );
                  env.TELEGRAM_CHAT_ID = await prompt.ask(
                    "Enter chat ID manually (or leave empty to set up later)",
                    "",
                  );
                }
              } catch (err) {
                warn(`Failed to fetch chat ID: ${err.message}`);
                console.log(
                  chalk.dim(
                    "  You can run: codex-monitor-chat-id (after starting the bot)",
                  ),
                );
                env.TELEGRAM_CHAT_ID = await prompt.ask(
                  "Enter chat ID manually (or leave empty to set up later)",
                  "",
                );
              }
            } else {
              console.log();
              info("No problem! You can get your chat ID later by:");
              console.log(
                "  • Running: " +
                  chalk.cyan("codex-monitor-chat-id") +
                  " (after starting codex-monitor)",
              );
              console.log(
                "  • Or manually: " +
                  chalk.cyan(
                    "curl 'https://api.telegram.org/bot<TOKEN>/getUpdates'",
                  ),
              );
              console.log("  Then add TELEGRAM_CHAT_ID to .env");
              console.log();
            }
          }

          // Step 4: Verify setup
          if (env.TELEGRAM_CHAT_ID) {
            console.log("\n" + chalk.bold("Step 4: Test Your Setup"));
            const testNow = await prompt.confirm(
              "Send a test message to verify setup?",
              true,
            );

            if (testNow) {
              info("Sending test message...");
              try {
                const testMsg =
                  "🤖 *Telegram Bot Test*\n\n" +
                  "Your codex-monitor Telegram bot is configured correctly!\n\n" +
                  `Project: ${env.PROJECT_NAME || configJson.projectName || "Unknown"}\n` +
                  "Try: /status, /tasks, /help";

                const response = await fetch(
                  `https://api.telegram.org/bot${env.TELEGRAM_BOT_TOKEN}/sendMessage`,
                  {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                      chat_id: env.TELEGRAM_CHAT_ID,
                      text: testMsg,
                      parse_mode: "Markdown",
                    }),
                  },
                );

                const result = await response.json();
                if (result.ok) {
                  info("✓ Test message sent! Check your Telegram.");
                } else {
                  warn(
                    `Test message failed: ${result.description || "Unknown error"}`,
                  );
                }
              } catch (err) {
                warn(`Failed to send test message: ${err.message}`);
              }
            }
          }
        } else {
          warn(
            "Bot token is required for Telegram setup. You can add it to .env later.",
          );
        }
      }
    }

    // ── Step 7: Kanban + Execution ─────────────────────────
    heading("Step 7 of 9 — Kanban & Execution");
    const backendDefault = String(
      process.env.KANBAN_BACKEND || configJson.kanban?.backend || "vk",
    )
      .trim()
      .toLowerCase();
    const backendIdx = await prompt.choose(
      "Select task board backend:",
      ["Vibe-Kanban (vk)", "GitHub Issues (github)"],
      backendDefault === "github" ? 1 : 0,
    );
    const selectedKanbanBackend = backendIdx === 1 ? "github" : "vk";
    env.KANBAN_BACKEND = selectedKanbanBackend;
    configJson.kanban = { backend: selectedKanbanBackend };

    const modeDefault = String(
      process.env.EXECUTOR_MODE || configJson.internalExecutor?.mode || "vk",
    )
      .trim()
      .toLowerCase();
    const execModeIdx = await prompt.choose(
      "Select execution mode:",
      [
        "Internal executor (recommended)",
        "VK executor/orchestrator",
        "Hybrid (internal + VK)",
      ],
      selectedKanbanBackend === "github"
        ? 0
        : modeDefault === "hybrid"
          ? 2
          : modeDefault === "internal"
            ? 0
            : 1,
    );
    const selectedExecutorMode =
      execModeIdx === 0 ? "internal" : execModeIdx === 1 ? "vk" : "hybrid";
    env.EXECUTOR_MODE = selectedExecutorMode;
    configJson.internalExecutor = {
      ...(configJson.internalExecutor || {}),
      mode: selectedExecutorMode,
    };

    const vkNeeded =
      selectedKanbanBackend === "vk" ||
      selectedExecutorMode === "vk" ||
      selectedExecutorMode === "hybrid";

    if (selectedKanbanBackend === "github") {
      const [slugOwner, slugRepo] = String(slug || "").split("/", 2);
      env.GITHUB_REPO_OWNER = await prompt.ask(
        "GitHub owner/org",
        process.env.GITHUB_REPO_OWNER || slugOwner || "",
      );
      env.GITHUB_REPO_NAME = await prompt.ask(
        "GitHub repository name",
        process.env.GITHUB_REPO_NAME || slugRepo || basename(repoRoot),
      );
      if (env.GITHUB_REPO_OWNER && env.GITHUB_REPO_NAME) {
        env.GITHUB_REPOSITORY = `${env.GITHUB_REPO_OWNER}/${env.GITHUB_REPO_NAME}`;
        env.KANBAN_PROJECT_ID = env.GITHUB_REPOSITORY;
      }
      info(
        "GitHub Issues selected as board. New issues and manual closes will sync into internal state.",
      );
    }

    if (vkNeeded) {
      if (isAdvancedSetup) {
        env.VK_BASE_URL = await prompt.ask(
          "VK API URL",
          process.env.VK_BASE_URL || "http://127.0.0.1:54089",
        );
        env.VK_RECOVERY_PORT = await prompt.ask(
          "VK port",
          process.env.VK_RECOVERY_PORT || "54089",
        );
      } else {
        env.VK_BASE_URL = process.env.VK_BASE_URL || "http://127.0.0.1:54089";
        env.VK_RECOVERY_PORT = process.env.VK_RECOVERY_PORT || "54089";
      }
      const spawnVk = await prompt.confirm(
        "Auto-spawn vibe-kanban if not running?",
        true,
      );
      if (!spawnVk) env.VK_NO_SPAWN = "true";
    } else {
      env.VK_NO_SPAWN = "true";
      info("VK runtime disabled (not selected as board or executor).");
    }

    // ── Codex CLI Config (config.toml) ─────────────────────
    heading("Codex CLI Config");
    console.log(chalk.dim("  ~/.codex/config.toml — agent-level config\n"));

    const existingToml = readCodexConfig();
    const configTomlPath = getConfigPath();

    if (!existingToml) {
      info(
        "No Codex CLI config found. Will create one with recommended settings.",
      );
    } else {
      info(`Found existing config: ${configTomlPath}`);
    }

    // Check vibe-kanban MCP
    if (vkNeeded) {
      if (existingToml && hasVibeKanbanMcp(existingToml)) {
        info("Vibe-Kanban MCP server already configured in config.toml.");
        const updateVk = await prompt.confirm(
          "Update VK env vars to match your setup values?",
          true,
        );
        if (!updateVk) {
          env._SKIP_VK_TOML = "1";
        }
      } else {
        info("Will add Vibe-Kanban MCP server to Codex config for agent use.");
      }
    } else {
      env._SKIP_VK_TOML = "1";
      info(
        "Skipping Vibe-Kanban MCP setup (VK not selected as board or executor).",
      );
    }

    // Check stream timeouts
    const timeouts = auditStreamTimeouts(existingToml);
    const lowTimeouts = timeouts.filter((t) => t.needsUpdate);
    if (lowTimeouts.length > 0) {
      for (const t of lowTimeouts) {
        const label =
          t.currentValue === null
            ? "not set"
            : `${(t.currentValue / 1000).toFixed(0)}s`;
        warn(
          `[${t.provider}] stream_idle_timeout_ms is ${label} — too low for complex reasoning.`,
        );
      }
      const fixTimeouts = await prompt.confirm(
        "Set stream timeouts to 60 minutes (recommended for agentic workloads)?",
        true,
      );
      if (!fixTimeouts) {
        env._SKIP_TIMEOUT_FIX = "1";
      }
    } else if (timeouts.length > 0) {
      success("Stream timeouts look good across all providers.");
    }

    // ── Orchestrator ──────────────────────────────────────
    heading("Orchestrator Script");
    console.log(
      chalk.dim(
        "  The orchestrator manages task execution and agent spawning.\n",
      ),
    );

    // Check for default scripts in codex-monitor directory
    const defaultOrchestrator = resolve(__dirname, "ve-orchestrator.ps1");
    const defaultKanban = resolve(__dirname, "ve-kanban.ps1");
    const hasDefaultScripts =
      existsSync(defaultOrchestrator) && existsSync(defaultKanban);

    if (hasDefaultScripts) {
      info(`Found default orchestrator scripts in codex-monitor:`);
      info(`  - ve-orchestrator.ps1`);
      info(`  - ve-kanban.ps1`);

      const useDefault = isAdvancedSetup
        ? await prompt.confirm(
            "Use the default ve-orchestrator.ps1 script?",
            true,
          )
        : true;

      if (useDefault) {
        env.ORCHESTRATOR_SCRIPT = defaultOrchestrator;
        success("Using default ve-orchestrator.ps1");
      } else {
        const customPath = await prompt.ask(
          "Path to your custom orchestrator script (or leave blank for Vibe-Kanban direct mode)",
          "",
        );
        if (customPath) {
          env.ORCHESTRATOR_SCRIPT = customPath;
        } else {
          info(
            "No orchestrator script configured. Codex-monitor will manage tasks directly via Vibe-Kanban.",
          );
        }
      }
    } else {
      const hasOrcScript = isAdvancedSetup
        ? await prompt.confirm(
            "Do you have an existing orchestrator script?",
            false,
          )
        : false;
      if (hasOrcScript) {
        env.ORCHESTRATOR_SCRIPT = await prompt.ask(
          "Path to orchestrator script",
          "",
        );
      } else {
        info(
          "No orchestrator script configured. Codex-monitor will manage tasks directly via Vibe-Kanban.",
        );
      }
    }

    env.MAX_PARALLEL = await prompt.ask(
      "Max parallel agent slots",
      process.env.MAX_PARALLEL || "6",
    );

    // ── Agent Templates ───────────────────────────────────
    heading("Agent Templates");
    console.log(
      chalk.dim(
        "  codex-monitor prompt templates are scaffolded to .codex-monitor/agents and loaded automatically.\n",
      ),
    );
    const generateAgents = isAdvancedSetup
      ? await prompt.confirm(
          "Scaffold .codex-monitor/agents prompt files?",
          true,
        )
      : true;

    if (generateAgents) {
      const promptsResult = ensureAgentPromptWorkspace(repoRoot);
      const addedGitIgnore = ensureRepoGitIgnoreEntry(
        repoRoot,
        "/.codex-monitor/",
      );
      configJson.agentPrompts = getDefaultPromptOverrides();

      if (addedGitIgnore) {
        success("Updated .gitignore with '/.codex-monitor/'");
      }
      if (promptsResult.written.length > 0) {
        success(
          `Created ${promptsResult.written.length} prompt template file(s) in ${relative(repoRoot, promptsResult.workspaceDir)}`,
        );
      } else {
        info("Prompt templates already exist — keeping existing files");
      }

      // Optional AGENTS.md seed
      const agentsMdPath = resolve(repoRoot, "AGENTS.md");
      if (!existsSync(agentsMdPath)) {
        const createAgentsGuide = await prompt.confirm(
          "Create AGENTS.md guide file as well?",
          true,
        );
        if (createAgentsGuide) {
          writeFileSync(
            agentsMdPath,
            generateAgentsMd(env.PROJECT_NAME, env.GITHUB_REPO),
            "utf8",
          );
          success(`Created ${relative(repoRoot, agentsMdPath)}`);
        }
      } else {
        info("AGENTS.md already exists — leaving unchanged");
      }
    } else {
      configJson.agentPrompts = getDefaultPromptOverrides();
    }

    // ── Agent Hooks ───────────────────────────────────────
    heading("Agent Hooks");
    console.log(
      chalk.dim(
        "  Configure shared hook policies for Codex, Claude Code, and Copilot CLI.\n",
      ),
    );

    const scaffoldHooks = isAdvancedSetup
      ? await prompt.confirm(
          "Scaffold hook configs for Codex/Claude/Copilot?",
          true,
        )
      : true;

    if (scaffoldHooks) {
      const profileMap = ["strict", "balanced", "lightweight", "none"];
      let profile = "balanced";
      let targets = ["codex", "claude", "copilot"];
      let prePushRaw = process.env.CODEX_MONITOR_HOOK_PREPUSH || "";
      let preCommitRaw = process.env.CODEX_MONITOR_HOOK_PRECOMMIT || "";
      let taskCompleteRaw = process.env.CODEX_MONITOR_HOOK_TASK_COMPLETE || "";
      let overwriteHooks = false;

      if (isAdvancedSetup) {
        const profileIdx = await prompt.choose(
          "Select hook policy:",
          [
            "Strict — pre-commit + pre-push + task validation",
            "Balanced — pre-push + task validation",
            "Lightweight — session/audit hooks only (no validation gates)",
            "None — disable codex-monitor built-in validation hooks",
          ],
          0,
        );
        profile = profileMap[profileIdx] || "strict";

        const targetIdx = await prompt.choose(
          "Hook files to scaffold:",
          [
            "All agents (Codex + Claude + Copilot)",
            "Codex + Claude",
            "Codex + Copilot",
            "Codex only",
            "Custom target list",
          ],
          0,
        );

        if (targetIdx === 0) targets = ["codex", "claude", "copilot"];
        else if (targetIdx === 1) targets = ["codex", "claude"];
        else if (targetIdx === 2) targets = ["codex", "copilot"];
        else if (targetIdx === 3) targets = ["codex"];
        else {
          const customTargets = await prompt.ask(
            "Custom targets (comma-separated: codex,claude,copilot)",
            "codex,claude,copilot",
          );
          targets = normalizeHookTargets(customTargets);
        }

        console.log(
          chalk.dim(
            "  Optional command overrides: use ';;' between commands, or 'none' to disable a hook event.\n",
          ),
        );

        prePushRaw = await prompt.ask(
          "Pre-push command override",
          process.env.CODEX_MONITOR_HOOK_PREPUSH || "",
        );
        preCommitRaw = await prompt.ask(
          "Pre-commit command override",
          process.env.CODEX_MONITOR_HOOK_PRECOMMIT || "",
        );
        taskCompleteRaw = await prompt.ask(
          "Task-complete command override",
          process.env.CODEX_MONITOR_HOOK_TASK_COMPLETE || "",
        );

        overwriteHooks = await prompt.confirm(
          "Overwrite existing generated hook files when present?",
          false,
        );
      } else {
        info(
          "Using recommended hook defaults: balanced policy for codex, claude, and copilot.",
        );
      }

      const hookResult = scaffoldAgentHookFiles(repoRoot, {
        enabled: true,
        profile,
        targets,
        overwriteExisting: overwriteHooks,
        commands: {
          PrePush: parseHookCommandInput(prePushRaw),
          PreCommit: parseHookCommandInput(preCommitRaw),
          TaskComplete: parseHookCommandInput(taskCompleteRaw),
        },
      });

      printHookScaffoldSummary(hookResult);
      Object.assign(env, hookResult.env);
      configJson.hookProfiles = {
        enabled: true,
        profile,
        targets,
        overwriteExisting: overwriteHooks,
      };
    } else {
      const hookResult = scaffoldAgentHookFiles(repoRoot, { enabled: false });
      Object.assign(env, hookResult.env);
      configJson.hookProfiles = {
        enabled: false,
      };
      info("Hook scaffolding skipped by user selection.");
    }

    // ── VK Auto-Wiring ────────────────────────────────────
    if (vkNeeded) {
      heading("Vibe-Kanban Auto-Configuration");
      const autoWireVk = isAdvancedSetup
        ? await prompt.confirm(
            "Auto-configure Vibe-Kanban project, repos, and executor profiles?",
            true,
          )
        : true;

      if (autoWireVk) {
        const vkConfig = {
          projectName: env.PROJECT_NAME,
          repoRoot,
          monitorDir: __dirname,
        };

        // Generate VK scripts
        const setupScript = generateVkSetupScript(vkConfig);
        const cleanupScript = generateVkCleanupScript(vkConfig);

        // Get current PATH for VK executor profiles
        const currentPath = process.env.PATH || "";

        // Write to config for VK API auto-wiring
        configJson.vkAutoConfig = {
          setupScript,
          cleanupScript,
          executorProfiles: configJson.executors.map((e) => ({
            executor: e.executor,
            variant: e.variant,
            environmentVariables: {
              PATH: currentPath,
              // Ensure GitHub token is available in workspace
              GH_TOKEN: "${GH_TOKEN}",
              GITHUB_TOKEN: "${GITHUB_TOKEN}",
            },
          })),
        };

        info("VK configuration will be applied on first launch.");
        info("Setup and cleanup scripts generated for your workspace.");
        info(
          `PATH environment variable configured for ${configJson.executors.length} executor profile(s)`,
        );
      }
    } else {
      info("Skipping VK auto-configuration (VK not selected).");
      delete configJson.vkAutoConfig;
    }

    // ── Step 8: Optional Channels ─────────────────────────
    heading("Step 8 of 9 — Optional Channels (WhatsApp & Container)");

    console.log(
      chalk.dim(
        "  These are optional features. Skip them if you only want Telegram.",
      ),
    );

    // WhatsApp
    const enableWhatsApp = await prompt.confirm(
      "Enable WhatsApp channel?",
      false,
    );
    if (enableWhatsApp) {
      env.WHATSAPP_ENABLED = "true";
      env.WHATSAPP_CHAT_ID = await prompt.ask(
        "WhatsApp Chat/Group ID (JID)",
        process.env.WHATSAPP_CHAT_ID || "",
      );
      env.WHATSAPP_ASSISTANT_NAME = isAdvancedSetup
        ? await prompt.ask(
            "WhatsApp assistant display name",
            env.PROJECT_NAME || "Codex Monitor",
          )
        : env.PROJECT_NAME || "Codex Monitor";
      info(
        "Run `codex-monitor --whatsapp-auth` after setup to authenticate with WhatsApp.",
      );
    } else {
      env.WHATSAPP_ENABLED = "false";
    }

    // Container isolation
    const enableContainer = await prompt.confirm(
      "Enable container isolation for agent execution?",
      false,
    );
    if (enableContainer) {
      env.CONTAINER_ENABLED = "true";
      if (isAdvancedSetup) {
        const runtimeIdx = await prompt.choose(
          "Container runtime",
          ["docker", "podman", "auto-detect"],
          2,
        );
        env.CONTAINER_RUNTIME = ["docker", "podman", "auto"][runtimeIdx];
        env.CONTAINER_IMAGE = await prompt.ask(
          "Container image",
          process.env.CONTAINER_IMAGE || "node:22-slim",
        );
        env.CONTAINER_MEMORY_LIMIT = await prompt.ask(
          "Memory limit (e.g. 2g)",
          process.env.CONTAINER_MEMORY_LIMIT || "4g",
        );
      } else {
        env.CONTAINER_RUNTIME = process.env.CONTAINER_RUNTIME || "auto";
        env.CONTAINER_IMAGE = process.env.CONTAINER_IMAGE || "node:22-slim";
      }
    } else {
      env.CONTAINER_ENABLED = "false";
    }

    // ── Step 9: Startup Service ────────────────────────────
    heading("Step 9 of 9 — Startup Service");

    const { getStartupStatus, getStartupMethodName } =
      await import("./startup-service.mjs");
    const currentStartup = getStartupStatus();
    const methodName = getStartupMethodName();

    if (currentStartup.installed) {
      info(`Startup service already installed via ${currentStartup.method}.`);
      const reinstall = await prompt.confirm(
        "Re-install startup service?",
        false,
      );
      env._STARTUP_SERVICE = reinstall ? "1" : "skip";
    } else {
      console.log(
        chalk.dim(
          `  Auto-start codex-monitor when you log in using ${methodName}.`,
        ),
      );
      console.log(
        chalk.dim(
          "  It will run in daemon mode (background) with auto-restart on failure.",
        ),
      );
      const enableStartup = await prompt.confirm(
        "Enable auto-start on login?",
        true,
      );
      env._STARTUP_SERVICE = enableStartup ? "1" : "0";
    }
  } finally {
    prompt.close();
  }

  // ── Write Files ─────────────────────────────────────────
  normalizeSetupConfiguration({ env, configJson, repoRoot, slug });
  await writeConfigFiles({ env, configJson, repoRoot, configDir });
}

// ── Non-Interactive Mode ─────────────────────────────────────────────────────

async function runNonInteractive({
  env,
  configJson,
  repoRoot,
  slug,
  projectName,
  configDir,
}) {
  env.PROJECT_NAME = process.env.PROJECT_NAME || projectName;
  env.REPO_ROOT = process.env.REPO_ROOT || repoRoot;
  env.GITHUB_REPO = process.env.GITHUB_REPO || slug || "";
  env.TELEGRAM_BOT_TOKEN = process.env.TELEGRAM_BOT_TOKEN || "";
  env.TELEGRAM_CHAT_ID = process.env.TELEGRAM_CHAT_ID || "";
  env.KANBAN_BACKEND = process.env.KANBAN_BACKEND || "vk";
  env.EXECUTOR_MODE = process.env.EXECUTOR_MODE || "vk";
  env.VK_BASE_URL = process.env.VK_BASE_URL || "http://127.0.0.1:54089";
  env.VK_RECOVERY_PORT = process.env.VK_RECOVERY_PORT || "54089";
  env.GITHUB_REPO_OWNER =
    process.env.GITHUB_REPO_OWNER || (slug ? String(slug).split("/")[0] : "");
  env.GITHUB_REPO_NAME =
    process.env.GITHUB_REPO_NAME || (slug ? String(slug).split("/")[1] : "");
  env.GITHUB_REPOSITORY =
    process.env.GITHUB_REPOSITORY ||
    (env.GITHUB_REPO_OWNER && env.GITHUB_REPO_NAME
      ? `${env.GITHUB_REPO_OWNER}/${env.GITHUB_REPO_NAME}`
      : "");
  if (!env.GITHUB_REPO && env.GITHUB_REPOSITORY) {
    env.GITHUB_REPO = env.GITHUB_REPOSITORY;
  }
  env.OPENAI_API_KEY = process.env.OPENAI_API_KEY || "";
  env.MAX_PARALLEL = process.env.MAX_PARALLEL || "6";

  // Optional channels
  env.WHATSAPP_ENABLED = process.env.WHATSAPP_ENABLED || "false";
  env.WHATSAPP_CHAT_ID = process.env.WHATSAPP_CHAT_ID || "";
  env.CONTAINER_ENABLED = process.env.CONTAINER_ENABLED || "false";
  env.CONTAINER_RUNTIME = process.env.CONTAINER_RUNTIME || "auto";

  // Copilot cloud: disabled by default — set to 0 to allow @copilot PR comments
  env.COPILOT_CLOUD_DISABLED = process.env.COPILOT_CLOUD_DISABLED || "true";

  // Parse EXECUTORS env if set, else use default preset
  if (process.env.EXECUTORS) {
    const entries = process.env.EXECUTORS.split(",").map((e) => e.trim());
    const roles = ["primary", "backup", "tertiary"];
    for (let i = 0; i < entries.length; i++) {
      const parts = entries[i].split(":");
      if (parts.length >= 2) {
        configJson.executors.push({
          name: `${parts[0].toLowerCase()}-${parts[1].toLowerCase()}`,
          executor: parts[0].toUpperCase(),
          variant: parts[1],
          weight: parts[2]
            ? Number(parts[2])
            : Math.floor(100 / entries.length),
          role: roles[i] || `executor-${i + 1}`,
          enabled: true,
        });
      }
    }
  }
  if (!configJson.executors.length) {
    configJson.executors = EXECUTOR_PRESETS["codex-only"];
  }

  configJson.projectName = env.PROJECT_NAME;
  configJson.kanban = { backend: env.KANBAN_BACKEND || "vk" };
  configJson.internalExecutor = {
    ...(configJson.internalExecutor || {}),
    mode: env.EXECUTOR_MODE || "vk",
  };
  configJson.failover = {
    strategy: process.env.FAILOVER_STRATEGY || "next-in-line",
    maxRetries: Number(process.env.FAILOVER_MAX_RETRIES || "3"),
    cooldownMinutes: Number(process.env.FAILOVER_COOLDOWN_MIN || "5"),
    disableOnConsecutiveFailures: Number(
      process.env.FAILOVER_DISABLE_AFTER || "3",
    ),
  };
  configJson.distribution = process.env.EXECUTOR_DISTRIBUTION || "weighted";
  configJson.repositories = [
    {
      name: basename(repoRoot),
      slug: env.GITHUB_REPO,
      primary: true,
    },
  ];
  configJson.agentPrompts = getDefaultPromptOverrides();
  ensureAgentPromptWorkspace(repoRoot);
  ensureRepoGitIgnoreEntry(repoRoot, "/.codex-monitor/");

  const hookOptions = buildHookScaffoldOptionsFromEnv(process.env);
  const hookResult = scaffoldAgentHookFiles(repoRoot, hookOptions);
  Object.assign(env, hookResult.env);
  configJson.hookProfiles = {
    enabled: hookResult.enabled,
    profile: hookResult.profile,
    targets: hookResult.targets,
    overwriteExisting: Boolean(hookOptions.overwriteExisting),
  };
  printHookScaffoldSummary(hookResult);

  // Startup service: respect STARTUP_SERVICE env in non-interactive mode
  if (parseBooleanEnvValue(process.env.STARTUP_SERVICE, false)) {
    env._STARTUP_SERVICE = "1";
  } else if (
    process.env.STARTUP_SERVICE !== undefined &&
    !parseBooleanEnvValue(process.env.STARTUP_SERVICE, true)
  ) {
    env._STARTUP_SERVICE = "0";
  }
  // else: don't set — writeConfigFiles will skip silently

  if (
    (env.KANBAN_BACKEND || "").toLowerCase() !== "vk" &&
    !["vk", "hybrid"].includes((env.EXECUTOR_MODE || "").toLowerCase())
  ) {
    env.VK_NO_SPAWN = "true";
    delete configJson.vkAutoConfig;
  }

  normalizeSetupConfiguration({ env, configJson, repoRoot, slug });
  await writeConfigFiles({ env, configJson, repoRoot, configDir });
}

// ── File Writing ─────────────────────────────────────────────────────────────

async function writeConfigFiles({ env, configJson, repoRoot, configDir }) {
  heading("Writing Configuration");
  const targetDir = resolve(configDir || __dirname);
  mkdirSync(targetDir, { recursive: true });
  ensureAgentPromptWorkspace(repoRoot);
  ensureRepoGitIgnoreEntry(repoRoot, "/.codex-monitor/");
  if (
    !configJson.agentPrompts ||
    Object.keys(configJson.agentPrompts).length === 0
  ) {
    configJson.agentPrompts = getDefaultPromptOverrides();
  }

  // ── .env file ──────────────────────────────────────────
  const envPath = resolve(targetDir, ".env");
  const targetEnvPath = existsSync(envPath)
    ? resolve(targetDir, ".env.generated")
    : envPath;

  if (existsSync(envPath)) {
    warn(`.env already exists. Writing to .env.generated`);
  }

  const envTemplatePath = resolve(__dirname, ".env.example");
  const templateText = existsSync(envTemplatePath)
    ? readFileSync(envTemplatePath, "utf8")
    : "";

  const envOut = templateText
    ? buildStandardizedEnvFile(templateText, env)
    : buildStandardizedEnvFile("", env);

  writeFileSync(targetEnvPath, envOut, "utf8");
  success(`Environment written to ${relative(repoRoot, targetEnvPath)}`);

  // ── codex-monitor.config.json ──────────────────────────
  // Write config with schema reference for editor autocomplete
  const configOut = { $schema: "./codex-monitor.schema.json", ...configJson };
  // Keep vkAutoConfig in config file for monitor to apply on first launch
  // (includes executorProfiles with environment variables like PATH)
  const configPath = resolve(targetDir, "codex-monitor.config.json");
  writeFileSync(configPath, JSON.stringify(configOut, null, 2) + "\n", "utf8");
  success(`Config written to ${relative(repoRoot, configPath)}`);

  // ── Codex CLI config.toml ─────────────────────────────
  heading("Codex CLI Config");

  if (env._SKIP_VK_TOML === "1") {
    info("Skipped Vibe-Kanban MCP config update.");
  } else {
    const vkPort = env.VK_RECOVERY_PORT || "54089";
    const vkBaseUrl = env.VK_BASE_URL || `http://127.0.0.1:${vkPort}`;
    const tomlResult = ensureCodexConfig({
      vkBaseUrl,
      dryRun: false,
    });
    printConfigSummary(tomlResult, (msg) => console.log(msg));
  }

  // ── Install dependencies ───────────────────────────────
  heading("Installing Dependencies");
  try {
    if (commandExists("pnpm")) {
      execSync("pnpm install", { cwd: __dirname, stdio: "inherit" });
    } else {
      execSync("npm install", { cwd: __dirname, stdio: "inherit" });
    }
    success("Dependencies installed");
  } catch {
    warn(
      "Dependency install failed — run manually: pnpm install (or) npm install",
    );
  }

  // ── Startup Service ────────────────────────────────────
  if (env._STARTUP_SERVICE === "1") {
    heading("Startup Service");
    try {
      const { installStartupService } = await import("./startup-service.mjs");
      const result = await installStartupService({ daemon: true });
      if (result.success) {
        success(`Registered via ${result.method}`);
        if (result.name) info(`Service name: ${result.name}`);
        if (result.path) info(`Config path: ${result.path}`);
      } else {
        warn(`Could not register startup service: ${result.error}`);
        info("You can try manually later: codex-monitor --enable-startup");
      }
    } catch (err) {
      warn(`Startup service registration failed: ${err.message}`);
      info("You can try manually later: codex-monitor --enable-startup");
    }
  } else if (env._STARTUP_SERVICE === "0") {
    info(
      "Startup service skipped — enable anytime: codex-monitor --enable-startup",
    );
  }

  // ── Summary ────────────────────────────────────────────
  console.log("");
  console.log(
    "  ╔═══════════════════════════════════════════════════════════════╗",
  );
  console.log(
    "  ║                    ✅ Setup Complete!                        ║",
  );
  console.log(
    "  ╚═══════════════════════════════════════════════════════════════╝",
  );
  console.log("");

  // Executor summary
  const totalWeight = configJson.executors.reduce((s, e) => s + e.weight, 0);
  console.log(chalk.bold("  Executors:"));
  for (const e of configJson.executors) {
    const pct =
      totalWeight > 0 ? Math.round((e.weight / totalWeight) * 100) : 0;
    console.log(
      `    ${e.role.padEnd(10)} ${e.executor}:${e.variant} — ${pct}%`,
    );
  }
  console.log(
    chalk.dim(
      `  Strategy: ${configJson.distribution} distribution, ${configJson.failover.strategy} failover`,
    ),
  );

  // Missing items
  console.log("");
  if (!env.TELEGRAM_BOT_TOKEN) {
    info("Telegram not configured — add TELEGRAM_BOT_TOKEN to .env later.");
  }
  if (
    !env.OPENAI_API_KEY &&
    !parseBooleanEnvValue(env.CODEX_SDK_DISABLED, false)
  ) {
    info("No API key set — AI analysis & autofix will be disabled.");
  }

  console.log("");
  console.log(chalk.bold("  Next steps:"));
  console.log("");
  console.log(chalk.green("    codex-monitor"));
  console.log(chalk.dim("    Start the orchestrator supervisor\n"));
  console.log(chalk.green("    codex-monitor --setup"));
  console.log(chalk.dim("    Re-run this wizard anytime\n"));
  console.log(chalk.green("    codex-monitor --enable-startup"));
  console.log(chalk.dim("    Register auto-start on login\n"));
  console.log(chalk.green("    codex-monitor --help"));
  console.log(chalk.dim("    See all options & env vars\n"));
}

// ── Auto-Launch Detection ────────────────────────────────────────────────────

/**
 * Check whether setup should run automatically (first launch detection).
 * Called from monitor.mjs before starting the main loop.
 */
export function shouldRunSetup() {
  const repoRoot = detectRepoRoot();
  const configDir = resolveConfigDir(repoRoot);
  return !hasSetupMarkers(configDir);
}

/**
 * Run setup wizard. Can be imported and called from monitor.mjs.
 */
export async function runSetup() {
  await main();
}

// ── Entry Point ──────────────────────────────────────────────────────────────

// Only run the wizard when executed directly (not when imported by cli.mjs)
const __filename_setup = fileURLToPath(import.meta.url);
if (process.argv[1] && resolve(process.argv[1]) === resolve(__filename_setup)) {
  main().catch((err) => {
    console.error(`\n  Setup failed: ${err.message}\n`);
    process.exit(1);
  });
}
