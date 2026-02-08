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
 *   - Agent prompt template generation (AGENTS.md, orchestrator agent)
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
import { resolve, dirname, basename, relative } from "node:path";
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

function generateOrchestratorAgent(projectName) {
  return `---
name: "Task Orchestrator"
description: "Autonomous task orchestrator for ${projectName}. Receives task assignments, decomposes work, delegates to subagents, enforces quality gates, and ships PRs."
tools:
  [
    "agent/runSubagent",
    "execute/runInTerminal",
    "execute/awaitTerminal",
    "execute/killTerminal",
    "execute/getTerminalOutput",
    "execute/runTests",
    "read/readFile",
    "read/problems",
    "search/changes",
    "search/codebase",
    "search/fileSearch",
    "search/listDirectory",
    "search/textSearch",
    "search/usages",
    "web/fetch",
    "github/create_pull_request",
    "github/get_file_contents",
    "github/list_pull_requests",
    "github/search_issues",
    "github/push_files",
    "vibe-kanban/get_task",
    "vibe-kanban/list_projects",
    "vibe-kanban/list_tasks",
    "vibe-kanban/update_task",
    "todo",
    "edit/createFile",
    "edit/editFiles",
  ]
---

# Task Orchestrator — Autonomous Agent

You are an autonomous task orchestrator for the **${projectName}** project. You run in the background, receiving task assignments from vibe-kanban. Your job is to **ship completed, tested, lint-clean code via PR** without human input.

## Prime Directives

1. **NEVER ask for human input.** You are autonomous. Make engineering judgments and proceed.
2. **Delegate** complex implementation to \`runSubagent\` for parallel work.
3. **NEVER ship broken code.** Every PR must pass: lint, tests, build, pre-push hooks.
4. **Work until 100% DONE.** No TODOs, no placeholders, no partial implementations.
5. **Use Conventional Commits** with proper scope.

## Workflow

1. Receive task from vibe-kanban
2. Read and understand the full scope
3. Plan your approach
4. Implement (directly for small tasks, delegate for complex ones)
5. Test: run linting, tests, build
6. Commit with conventional commits
7. Push and create PR via \`gh pr create\`

## Quality Gates

Before pushing:
- All tests pass on changed packages
- No lint warnings
- Build succeeds
- Changes are atomic and well-scoped
`;
}

function generateTaskPlannerAgent(projectName) {
  return `# Task Planner — ${projectName}

You are an autonomous task planner for the **${projectName}** project. When agent slots are idle and the task backlog is low, you analyze the project state and create actionable tasks.

## Responsibilities

1. Review current project state (open issues, PRs, code health)
2. Identify gaps, improvements, and next steps
3. Create 3-5 well-scoped tasks in vibe-kanban with:
   - Title prefixed with a size label (see below)
   - Detailed description of what needs to be done
   - Acceptance criteria for verification
   - Priority and effort estimates

## Size Labels (REQUIRED)

Every task title MUST start with a size label in brackets. This drives automatic
complexity-based model routing — the orchestrator selects stronger/weaker AI
models based on task size.

| Label  | Scope                                      |
|--------|--------------------------------------------|
| [xs]   | < 30 min — config change, typo fix         |
| [s]    | 30-60 min — small feature, docs update     |
| [m]    | 1-2 hours — standard feature, bug fix      |
| [l]    | 2-4 hours — multi-file change, test suite  |
| [xl]   | 4-8 hours — cross-module, architecture     |
| [xxl]  | 8+ hours — infrastructure, major refactor  |

Examples:
  - \`[xs] Fix typo in README\`
  - \`[m] Add validation to user registration endpoint\`
  - \`[xl] Implement distributed task claiming protocol\`

## Guidelines

- Tasks should be completable in 1-4 hours by a single agent
- Prioritize: bug fixes > test coverage > tech debt > new features
- Check for existing similar tasks to avoid duplicates
- Consider dependencies between tasks
- Use appropriate size labels based on estimated complexity
`;
}

// ── VK Auto-Configuration ────────────────────────────────────────────────────

function generateVkSetupScript(config) {
  const repoRoot = config.repoRoot.replace(/\\/g, "/");
  const monitorDir = config.monitorDir.replace(/\\/g, "/");
  const agentAppend = config.agentFile ? ` --agent "${config.agentFile}"` : "";

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
  heading("Step 1 of 7 — Prerequisites");
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

  if (isNonInteractive) {
    return runNonInteractive({ env, configJson, repoRoot, slug, projectName });
  }

  const prompt = createPrompt();

  try {
    // ── Step 2: Project Identity ──────────────────────────
    heading("Step 2 of 7 — Project Identity");
    env.PROJECT_NAME = await prompt.ask("Project name", projectName);
    env.GITHUB_REPO = await prompt.ask(
      "GitHub repo slug (org/repo)",
      process.env.GITHUB_REPO || slug || "",
    );
    configJson.projectName = env.PROJECT_NAME;

    // ── Step 3: Repository ─────────────────────────────────
    heading("Step 3 of 7 — Repository Configuration");
    const multiRepo = await prompt.confirm(
      "Do you have multiple repositories (e.g. separate backend/frontend)?",
      false,
    );

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
    }

    // ── Step 4: Executor Configuration ─────────────────────
    heading("Step 4 of 7 — Executor / Agent Configuration");
    console.log("  Executors are the AI agents that work on tasks.\n");

    const presetIdx = await prompt.choose(
      "Select executor preset:",
      [
        "Codex only",
        "Copilot + Codex (50/50 split)",
        "Copilot only (Claude Opus 4.6)",
        "Triple (Copilot Claude 40%, Codex 35%, Copilot GPT 25%)",
        "Custom — I'll define my own executors",
      ],
      0,
    );

    const presetNames = [
      "codex-only",
      "copilot-codex",
      "copilot-only",
      "triple",
      "custom",
    ];
    const presetKey = presetNames[presetIdx];

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

    // Failover strategy
    console.log();
    console.log(chalk.dim("  What happens when an executor fails repeatedly?"));
    console.log();

    const failoverIdx = await prompt.choose(
      "Select failover strategy:",
      FAILOVER_STRATEGIES.map((f) => `${f.name} — ${f.desc}`),
      0,
    );
    configJson.failover = {
      strategy: FAILOVER_STRATEGIES[failoverIdx].name,
      maxRetries: Number(await prompt.ask("Max retries before failover", "3")),
      cooldownMinutes: Number(
        await prompt.ask("Cooldown after disabling executor (minutes)", "5"),
      ),
      disableOnConsecutiveFailures: Number(
        await prompt.ask("Disable executor after N consecutive failures", "3"),
      ),
    };

    // Distribution mode
    const distIdx = await prompt.choose(
      "\n  Task distribution mode:",
      DISTRIBUTION_MODES.map((d) => `${d.name} — ${d.desc}`),
      0,
    );
    configJson.distribution = DISTRIBUTION_MODES[distIdx].name;

    // ── Step 5: AI Provider ────────────────────────────────
    heading("Step 5 of 7 — AI / Codex Provider");
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
      env.CODEX_SDK_DISABLED = "1";
    }

    // ── Step 6: Telegram ──────────────────────────────────
    heading("Step 6 of 7 — Telegram Notifications");
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

    // ── Step 7: Vibe-Kanban ───────────────────────────────
    heading("Step 7 of 7 — Vibe-Kanban & Orchestration");
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

      const useDefault = await prompt.confirm(
        "Use the default ve-orchestrator.ps1 script?",
        true,
      );

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
      const hasOrcScript = await prompt.confirm(
        "Do you have an existing orchestrator script?",
        false,
      );
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

    env.MAX_PARALLEL = await prompt.ask("Max parallel agent slots", "6");

    // ── Agent Templates ───────────────────────────────────
    heading("Agent Templates");
    console.log(
      chalk.dim(
        "  Agent templates (AGENTS.md) guide AI agents working on your codebase.\n",
      ),
    );
    const generateAgents = await prompt.confirm(
      "Generate agent template files for this project?",
      true,
    );

    if (generateAgents) {
      const agentDir = resolve(repoRoot, ".github", "agents");
      mkdirSync(agentDir, { recursive: true });

      // AGENTS.md at repo root
      const agentsMdPath = resolve(repoRoot, "AGENTS.md");
      if (!existsSync(agentsMdPath)) {
        writeFileSync(
          agentsMdPath,
          generateAgentsMd(env.PROJECT_NAME, env.GITHUB_REPO),
          "utf8",
        );
        success(`Created ${relative(repoRoot, agentsMdPath)}`);
      } else {
        info(`AGENTS.md already exists — skipping`);
      }

      // Orchestrator agent
      const orchPath = resolve(agentDir, "orchestrator.agent.md");
      if (!existsSync(orchPath)) {
        writeFileSync(
          orchPath,
          generateOrchestratorAgent(env.PROJECT_NAME),
          "utf8",
        );
        success(`Created ${relative(repoRoot, orchPath)}`);
        configJson.agentPrompts.orchestrator = relative(__dirname, orchPath);
      } else {
        info(`orchestrator.agent.md already exists — using existing`);
        configJson.agentPrompts.orchestrator = relative(__dirname, orchPath);
      }

      // Task planner agent
      const plannerPath = resolve(agentDir, "task-planner.agent.md");
      if (!existsSync(plannerPath)) {
        writeFileSync(
          plannerPath,
          generateTaskPlannerAgent(env.PROJECT_NAME),
          "utf8",
        );
        success(`Created ${relative(repoRoot, plannerPath)}`);
        configJson.agentPrompts.planner = relative(__dirname, plannerPath);
      } else {
        info(`task-planner.agent.md already exists — using existing`);
        configJson.agentPrompts.planner = relative(__dirname, plannerPath);
      }
    }

    // ── VK Auto-Wiring ────────────────────────────────────
    heading("Vibe-Kanban Auto-Configuration");
    const autoWireVk = await prompt.confirm(
      "Auto-configure Vibe-Kanban project, repos, and executor profiles?",
      true,
    );

    if (autoWireVk) {
      const vkConfig = {
        projectName: env.PROJECT_NAME,
        repoRoot,
        monitorDir: __dirname,
        agentFile: configJson.agentPrompts.orchestrator
          ? resolve(__dirname, configJson.agentPrompts.orchestrator)
          : null,
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
      info(`PATH environment variable configured for ${configJson.executors.length} executor profile(s)`);
    }
  } finally {
    prompt.close();
  }

  // ── Write Files ─────────────────────────────────────────
  await writeConfigFiles({ env, configJson, repoRoot });
}

// ── Non-Interactive Mode ─────────────────────────────────────────────────────

async function runNonInteractive({
  env,
  configJson,
  repoRoot,
  slug,
  projectName,
}) {
  env.PROJECT_NAME = process.env.PROJECT_NAME || projectName;
  env.GITHUB_REPO = process.env.GITHUB_REPO || slug || "";
  env.TELEGRAM_BOT_TOKEN = process.env.TELEGRAM_BOT_TOKEN || "";
  env.TELEGRAM_CHAT_ID = process.env.TELEGRAM_CHAT_ID || "";
  env.VK_BASE_URL = process.env.VK_BASE_URL || "http://127.0.0.1:54089";
  env.VK_RECOVERY_PORT = process.env.VK_RECOVERY_PORT || "54089";
  env.OPENAI_API_KEY = process.env.OPENAI_API_KEY || "";
  env.MAX_PARALLEL = process.env.MAX_PARALLEL || "6";

  // Copilot cloud: disabled by default — set to 0 to allow @copilot PR comments
  env.COPILOT_CLOUD_DISABLED = process.env.COPILOT_CLOUD_DISABLED || "1";

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

  await writeConfigFiles({ env, configJson, repoRoot });
}

// ── File Writing ─────────────────────────────────────────────────────────────

async function writeConfigFiles({ env, configJson, repoRoot }) {
  heading("Writing Configuration");

  // ── .env file ──────────────────────────────────────────
  const envPath = resolve(__dirname, ".env");
  const targetEnvPath = existsSync(envPath)
    ? resolve(__dirname, ".env.generated")
    : envPath;

  if (existsSync(envPath)) {
    warn(`.env already exists. Writing to .env.generated`);
  }

  const lines = [
    "# Generated by codex-monitor setup wizard",
    `# ${new Date().toISOString()}`,
    "",
  ];

  for (const [key, value] of Object.entries(env)) {
    // Skip internal flags — not user-facing config
    if (key.startsWith("_")) continue;
    if (value) {
      // Quote values that contain spaces or special chars
      const needsQuotes = value.includes(" ") || value.includes("=");
      lines.push(`${key}=${needsQuotes ? `"${value}"` : value}`);
    } else {
      lines.push(`# ${key}=`);
    }
  }

  writeFileSync(targetEnvPath, lines.join("\n") + "\n", "utf8");
  success(`Environment written to ${relative(repoRoot, targetEnvPath)}`);

  // ── codex-monitor.config.json ──────────────────────────
  // Write config with schema reference for editor autocomplete
  const configOut = { $schema: "./codex-monitor.schema.json", ...configJson };
  // Keep vkAutoConfig in config file for monitor to apply on first launch
  // (includes executorProfiles with environment variables like PATH)
  const configPath = resolve(__dirname, "codex-monitor.config.json");
  writeFileSync(configPath, JSON.stringify(configOut, null, 2) + "\n", "utf8");
  success(`Config written to ${relative(repoRoot, configPath)}`);

  // ── Codex CLI config.toml ─────────────────────────────
  heading("Codex CLI Config");

  const vkPort = env.VK_RECOVERY_PORT || "54089";
  const vkBaseUrl = env.VK_BASE_URL || `http://127.0.0.1:${vkPort}`;
  const vkHost = "127.0.0.1";

  const tomlResult = ensureCodexConfig({
    vkBaseUrl,
    vkPort,
    vkHost,
    dryRun: false,
  });
  printConfigSummary(tomlResult, (msg) => console.log(msg));

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
  if (!env.OPENAI_API_KEY && env.CODEX_SDK_DISABLED !== "1") {
    info("No API key set — AI analysis & autofix will be disabled.");
  }

  console.log("");
  console.log(chalk.bold("  Next steps:"));
  console.log("");
  console.log(chalk.green("    codex-monitor"));
  console.log(chalk.dim("    Start the orchestrator supervisor\n"));
  console.log(chalk.green("    codex-monitor --setup"));
  console.log(chalk.dim("    Re-run this wizard anytime\n"));
  console.log(chalk.green("    codex-monitor --help"));
  console.log(chalk.dim("    See all options & env vars\n"));
}

// ── Auto-Launch Detection ────────────────────────────────────────────────────

/**
 * Check whether setup should run automatically (first launch detection).
 * Called from monitor.mjs before starting the main loop.
 */
export function shouldRunSetup() {
  const hasEnv = existsSync(resolve(__dirname, ".env"));
  const hasConfig =
    existsSync(resolve(__dirname, "codex-monitor.config.json")) ||
    existsSync(resolve(__dirname, ".codex-monitor.json"));
  return !hasEnv && !hasConfig;
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
