#!/usr/bin/env node

/**
 * codex-monitor — Configuration System
 *
 * Loads configuration from (in priority order):
 *   1. CLI flags (--key value)
 *   2. Environment variables
 *   3. .env file
 *   4. codex-monitor.config.json (project config)
 *   5. Built-in defaults
 *
 * Executor configuration supports N executors with weights and failover.
 */

import { existsSync, readFileSync } from "node:fs";
import { resolve, dirname, basename } from "node:path";
import { execSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));

// ── .env loader ──────────────────────────────────────────────────────────────

function loadDotEnv(dir, options = {}) {
  const { override = false } = options;
  const envPath = resolve(dir, ".env");
  if (!existsSync(envPath)) return;
  const lines = readFileSync(envPath, "utf8").split("\n");
  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;
    const eqIdx = trimmed.indexOf("=");
    if (eqIdx === -1) continue;
    const key = trimmed.slice(0, eqIdx).trim();
    let val = trimmed.slice(eqIdx + 1).trim();
    // Strip surrounding quotes
    if (
      (val.startsWith('"') && val.endsWith('"')) ||
      (val.startsWith("'") && val.endsWith("'"))
    ) {
      val = val.slice(1, -1);
    }
    if (override || !(key in process.env)) {
      process.env[key] = val;
    }
  }
}

// ── CLI arg parser ───────────────────────────────────────────────────────────

function parseArgs(argv) {
  const args = argv.slice(2);
  const result = { _positional: [], _flags: new Set() };
  for (let i = 0; i < args.length; i++) {
    if (args[i].startsWith("--")) {
      const key = args[i].slice(2);
      if (i + 1 < args.length && !args[i + 1].startsWith("--")) {
        result[key] = args[i + 1];
        i++;
      } else {
        result._flags.add(key);
      }
    } else {
      result._positional.push(args[i]);
    }
  }
  return result;
}

// ── Git helpers ──────────────────────────────────────────────────────────────

function detectRepoSlug() {
  try {
    const remote = execSync("git remote get-url origin", {
      encoding: "utf8",
      stdio: ["pipe", "pipe", "ignore"],
    }).trim();
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
    return process.cwd();
  }
}

// ── Executor Configuration ───────────────────────────────────────────────────

/**
 * Executor config schema:
 *
 * {
 *   "executors": [
 *     {
 *       "name": "copilot-claude",
 *       "executor": "COPILOT",
 *       "variant": "CLAUDE_OPUS_4_6",
 *       "weight": 50,
 *       "role": "primary",
 *       "enabled": true
 *     },
 *     {
 *       "name": "codex-default",
 *       "executor": "CODEX",
 *       "variant": "DEFAULT",
 *       "weight": 50,
 *       "role": "backup",
 *       "enabled": true
 *     }
 *   ],
 *   "failover": {
 *     "strategy": "next-in-line",   // "next-in-line" | "weighted-random" | "round-robin"
 *     "maxRetries": 3,
 *     "cooldownMinutes": 5,
 *     "disableOnConsecutiveFailures": 3
 *   },
 *   "distribution": "weighted"      // "weighted" | "round-robin" | "primary-only"
 * }
 */

const DEFAULT_EXECUTORS = {
  executors: [
    {
      name: "copilot-claude",
      executor: "COPILOT",
      variant: "CLAUDE_OPUS_4_6",
      weight: 50,
      role: "primary",
      enabled: true,
    },
    {
      name: "codex-default",
      executor: "CODEX",
      variant: "DEFAULT",
      weight: 50,
      role: "backup",
      enabled: true,
    },
  ],
  failover: {
    strategy: "next-in-line",
    maxRetries: 3,
    cooldownMinutes: 5,
    disableOnConsecutiveFailures: 3,
  },
  distribution: "weighted",
};

function parseExecutorsFromEnv() {
  // EXECUTORS=COPILOT:CLAUDE_OPUS_4_6:50,CODEX:DEFAULT:50
  const raw = process.env.EXECUTORS;
  if (!raw) return null;
  const entries = raw.split(",").map((e) => e.trim());
  const executors = [];
  const roles = ["primary", "backup", "tertiary"];
  for (let i = 0; i < entries.length; i++) {
    const parts = entries[i].split(":");
    if (parts.length < 2) continue;
    executors.push({
      name: `${parts[0].toLowerCase()}-${parts[1].toLowerCase()}`,
      executor: parts[0].toUpperCase(),
      variant: parts[1],
      weight: parts[2] ? Number(parts[2]) : Math.floor(100 / entries.length),
      role: roles[i] || `executor-${i + 1}`,
      enabled: true,
    });
  }
  return executors.length ? executors : null;
}

function loadExecutorConfig(configDir) {
  // 1. Try env var
  const fromEnv = parseExecutorsFromEnv();

  // 2. Try config file
  let fromFile = null;
  for (const name of [
    "codex-monitor.config.json",
    ".codex-monitor.json",
    "codex-monitor.json",
  ]) {
    const p = resolve(configDir, name);
    if (existsSync(p)) {
      try {
        const raw = JSON.parse(readFileSync(p, "utf8"));
        fromFile = raw.executors ? raw : null;
        break;
      } catch {
        /* invalid JSON — skip */
      }
    }
  }

  const executors =
    fromEnv || fromFile?.executors || DEFAULT_EXECUTORS.executors;
  const failover = fromFile?.failover || {
    strategy:
      process.env.FAILOVER_STRATEGY || DEFAULT_EXECUTORS.failover.strategy,
    maxRetries: Number(
      process.env.FAILOVER_MAX_RETRIES || DEFAULT_EXECUTORS.failover.maxRetries,
    ),
    cooldownMinutes: Number(
      process.env.FAILOVER_COOLDOWN_MIN ||
        DEFAULT_EXECUTORS.failover.cooldownMinutes,
    ),
    disableOnConsecutiveFailures: Number(
      process.env.FAILOVER_DISABLE_AFTER ||
        DEFAULT_EXECUTORS.failover.disableOnConsecutiveFailures,
    ),
  };
  const distribution =
    fromFile?.distribution ||
    process.env.EXECUTOR_DISTRIBUTION ||
    DEFAULT_EXECUTORS.distribution;

  return { executors, failover, distribution };
}

// ── Executor Scheduler ───────────────────────────────────────────────────────

class ExecutorScheduler {
  constructor(config) {
    this.executors = config.executors.filter((e) => e.enabled !== false);
    this.failover = config.failover;
    this.distribution = config.distribution;
    this._roundRobinIndex = 0;
    this._failureCounts = new Map(); // name → consecutive failures
    this._disabledUntil = new Map(); // name → timestamp
  }

  /** Get the next executor based on distribution strategy */
  next() {
    const available = this._getAvailable();
    if (!available.length) {
      // All disabled — reset and use primary
      this._disabledUntil.clear();
      this._failureCounts.clear();
      return this.executors[0];
    }

    switch (this.distribution) {
      case "round-robin":
        return this._roundRobin(available);
      case "primary-only":
        return available[0];
      case "weighted":
      default:
        return this._weightedSelect(available);
    }
  }

  /** Report a failure for an executor */
  recordFailure(executorName) {
    const count = (this._failureCounts.get(executorName) || 0) + 1;
    this._failureCounts.set(executorName, count);
    if (count >= this.failover.disableOnConsecutiveFailures) {
      const until = Date.now() + this.failover.cooldownMinutes * 60 * 1000;
      this._disabledUntil.set(executorName, until);
      this._failureCounts.set(executorName, 0);
    }
  }

  /** Report a success for an executor */
  recordSuccess(executorName) {
    this._failureCounts.set(executorName, 0);
    this._disabledUntil.delete(executorName);
  }

  /** Get failover executor when current one fails */
  getFailover(currentName) {
    const available = this._getAvailable().filter(
      (e) => e.name !== currentName,
    );
    if (!available.length) return null;

    switch (this.failover.strategy) {
      case "weighted-random":
        return this._weightedSelect(available);
      case "round-robin":
        return available[0];
      case "next-in-line":
      default: {
        // Find the next one by role priority
        const roleOrder = [
          "primary",
          "backup",
          "tertiary",
          ...Array.from({ length: 20 }, (_, i) => `executor-${i + 1}`),
        ];
        available.sort(
          (a, b) => roleOrder.indexOf(a.role) - roleOrder.indexOf(b.role),
        );
        return available[0];
      }
    }
  }

  /** Get summary for display */
  getSummary() {
    const total = this.executors.reduce((s, e) => s + e.weight, 0);
    return this.executors.map((e) => {
      const pct = total > 0 ? Math.round((e.weight / total) * 100) : 0;
      const disabled = this._isDisabled(e.name);
      return {
        ...e,
        percentage: pct,
        status: disabled ? "cooldown" : e.enabled ? "active" : "disabled",
        consecutiveFailures: this._failureCounts.get(e.name) || 0,
      };
    });
  }

  /** Format a display string like "COPILOT ⇄ CODEX (50/50)" */
  toDisplayString() {
    const summary = this.getSummary().filter((e) => e.status === "active");
    if (!summary.length) return "No executors available";
    return summary
      .map((e) => `${e.executor}:${e.variant}(${e.percentage}%)`)
      .join(" ⇄ ");
  }

  _getAvailable() {
    return this.executors.filter(
      (e) => e.enabled !== false && !this._isDisabled(e.name),
    );
  }

  _isDisabled(name) {
    const until = this._disabledUntil.get(name);
    if (!until) return false;
    if (Date.now() >= until) {
      this._disabledUntil.delete(name);
      return false;
    }
    return true;
  }

  _roundRobin(available) {
    const idx = this._roundRobinIndex % available.length;
    this._roundRobinIndex++;
    return available[idx];
  }

  _weightedSelect(available) {
    const totalWeight = available.reduce((s, e) => s + (e.weight || 1), 0);
    let r = Math.random() * totalWeight;
    for (const e of available) {
      r -= e.weight || 1;
      if (r <= 0) return e;
    }
    return available[available.length - 1];
  }
}

// ── Multi-Repo Support ───────────────────────────────────────────────────────

/**
 * Multi-repo config schema:
 *
 * {
 *   "repositories": [
 *     {
 *       "name": "backend",
 *       "path": "/path/to/backend",
 *       "slug": "org/backend",
 *       "orchestratorScript": "./orchestrator.ps1",
 *       "primary": true
 *     },
 *     {
 *       "name": "frontend",
 *       "path": "/path/to/frontend",
 *       "slug": "org/frontend"
 *     }
 *   ]
 * }
 */

function loadRepoConfig(configDir) {
  // Try config file for multi-repo
  for (const name of [
    "codex-monitor.config.json",
    ".codex-monitor.json",
    "codex-monitor.json",
  ]) {
    const p = resolve(configDir, name);
    if (existsSync(p)) {
      try {
        const raw = JSON.parse(readFileSync(p, "utf8"));
        if (raw.repositories && Array.isArray(raw.repositories)) {
          return raw.repositories;
        }
      } catch {
        /* skip */
      }
    }
  }

  // Single-repo from env
  const repoRoot = detectRepoRoot();
  const slug = detectRepoSlug();
  return [
    {
      name: basename(repoRoot),
      path: repoRoot,
      slug: process.env.GITHUB_REPO || slug || "unknown/unknown",
      primary: true,
    },
  ];
}

// ── Agent Prompt Templates ───────────────────────────────────────────────────

const DEFAULT_AGENT_PROMPT = `# Task Orchestrator Agent

You are an autonomous task orchestrator agent. You receive task assignments from vibe-kanban and execute them to completion.

## Prime Directives

1. **NEVER ask for human input.** You are autonomous. Make engineering judgments and proceed.
2. **Delegate implementation** to subagents when tasks are complex.
3. **NEVER ship broken code.** Every PR must have zero lint errors, zero test failures, zero build errors.
4. **Work until 100% DONE.** No TODOs, no placeholders, no partial implementations.
5. **Use Conventional Commits** with proper scope.

## Pre-Push Checklist

Before committing and pushing:
- Run linting and formatting
- Run unit tests on changed packages
- Ensure build passes
- Verify no regressions

## How You Receive Work

You receive a task description (from vibe-kanban or inline). Your job:
1. Understand the full scope
2. Plan your approach
3. Implement (or delegate to subagents)
4. Test and verify
5. Commit with conventional commits
6. Push and create a PR

## Quality Gates

- All tests pass
- No lint warnings
- Build succeeds
- Changes are atomic and well-scoped
`;

const DEFAULT_PLANNER_PROMPT = `# Task Planner Agent

You are an autonomous task planner. When the task backlog is low, you analyze the project state and create well-scoped, actionable tasks.

## Responsibilities

1. Review the current project state (open issues, PRs, code quality)
2. Identify gaps, improvements, and next steps
3. Create tasks in vibe-kanban with clear:
   - Title (concise, action-oriented)
   - Description (what needs to be done)
   - Acceptance criteria (how to verify completion)
   - Priority and effort estimates

## Guidelines

- Create 3-5 tasks per planning session
- Tasks should be completable in 1-4 hours by a single agent
- Prioritize bug fixes and test coverage over new features
- Consider technical debt and code quality improvements
- Check for existing similar tasks to avoid duplicates
`;

function loadAgentPrompts(configDir, repoRoot) {
  const prompts = {
    orchestrator: DEFAULT_AGENT_PROMPT,
    planner: DEFAULT_PLANNER_PROMPT,
  };

  // Try loading custom prompts from repo
  const agentDirs = [
    resolve(repoRoot, ".github", "agents"),
    resolve(repoRoot, ".agents"),
    resolve(configDir, "agents"),
  ];

  for (const dir of agentDirs) {
    // Look for orchestrator agent
    for (const name of [
      "orchestrator.agent.md",
      "ralph-orchestrator.agent.md",
      "agent.md",
    ]) {
      const p = resolve(dir, name);
      if (existsSync(p)) {
        prompts.orchestrator = readFileSync(p, "utf8");
        break;
      }
    }
    // Look for planner agent
    for (const name of [
      "planner.agent.md",
      "task-planner.agent.md",
      "Task Planner.agent.md",
    ]) {
      const p = resolve(dir, name);
      if (existsSync(p)) {
        prompts.planner = readFileSync(p, "utf8");
        break;
      }
    }
  }

  // Try config file overrides
  for (const name of [
    "codex-monitor.config.json",
    ".codex-monitor.json",
    "codex-monitor.json",
  ]) {
    const p = resolve(configDir, name);
    if (existsSync(p)) {
      try {
        const raw = JSON.parse(readFileSync(p, "utf8"));
        if (raw.agentPrompts?.orchestrator) {
          const opPath = resolve(configDir, raw.agentPrompts.orchestrator);
          if (existsSync(opPath)) {
            prompts.orchestrator = readFileSync(opPath, "utf8");
          }
        }
        if (raw.agentPrompts?.planner) {
          const ppPath = resolve(configDir, raw.agentPrompts.planner);
          if (existsSync(ppPath)) {
            prompts.planner = readFileSync(ppPath, "utf8");
          }
        }
      } catch {
        /* skip */
      }
    }
  }

  return prompts;
}

// ── Main Configuration Loader ────────────────────────────────────────────────

/**
 * Load the full codex-monitor configuration.
 * Returns a frozen config object used by all modules.
 */
export function loadConfig(argv = process.argv, options = {}) {
  const { reloadEnv = false } = options;
  const cli = parseArgs(argv);

  // Determine config directory (where codex-monitor lives)
  const configDir =
    cli["config-dir"] || process.env.CODEX_MONITOR_DIR || __dirname;

  // Load .env from config dir
  loadDotEnv(configDir, { override: reloadEnv });

  // Also load .env from repo root if different
  const repoRoot =
    cli["repo-root"] || process.env.REPO_ROOT || detectRepoRoot();
  if (resolve(repoRoot) !== resolve(configDir)) {
    loadDotEnv(repoRoot, { override: reloadEnv });
  }
  const envPaths = [
    resolve(configDir, ".env"),
    resolve(repoRoot, ".env"),
  ].filter((p, i, arr) => arr.indexOf(p) === i);

  // ── Project identity ─────────────────────────────────────
  const projectName =
    cli["project-name"] ||
    process.env.PROJECT_NAME ||
    process.env.VK_PROJECT_NAME ||
    detectProjectName(configDir, repoRoot);

  const repoSlug =
    cli["repo"] ||
    process.env.GITHUB_REPO ||
    detectRepoSlug() ||
    "unknown/unknown";

  const repoUrlBase =
    process.env.GITHUB_REPO_URL || `https://github.com/${repoSlug}`;

  // ── Orchestrator ─────────────────────────────────────────
  const defaultScript = findOrchestratorScript(configDir, repoRoot);
  const scriptPath = resolve(
    cli.script || process.env.ORCHESTRATOR_SCRIPT || defaultScript,
  );
  const scriptArgsRaw =
    cli.args || process.env.ORCHESTRATOR_ARGS || "-MaxParallel 6 -WaitForMutex";
  const scriptArgs = scriptArgsRaw.split(" ").filter(Boolean);

  // ── Timing ───────────────────────────────────────────────
  const restartDelayMs = Number(
    cli["restart-delay"] || process.env.RESTART_DELAY_MS || "10000",
  );
  const maxRestarts = Number(
    cli["max-restarts"] || process.env.MAX_RESTARTS || "0",
  );

  // ── Logging ──────────────────────────────────────────────
  const logDir = resolve(
    cli["log-dir"] || process.env.LOG_DIR || resolve(configDir, "logs"),
  );

  // ── Feature flags ────────────────────────────────────────
  const flags = cli._flags;
  const watchEnabled = !flags.has("no-watch");
  const watchPath = resolve(
    cli["watch-path"] || process.env.WATCH_PATH || scriptPath,
  );
  const echoLogs = !flags.has("no-echo-logs");
  const autoFixEnabled = !flags.has("no-autofix");
  const codexEnabled =
    !flags.has("no-codex") && process.env.CODEX_SDK_DISABLED !== "1";

  // ── Vibe-Kanban ──────────────────────────────────────────
  const vkRecoveryPort = process.env.VK_RECOVERY_PORT || "54089";
  const vkRecoveryHost =
    process.env.VK_RECOVERY_HOST || process.env.VK_HOST || "0.0.0.0";
  const vkEndpointUrl =
    process.env.VK_ENDPOINT_URL ||
    process.env.VK_BASE_URL ||
    `http://127.0.0.1:${vkRecoveryPort}`;
  const vkPublicUrl = process.env.VK_PUBLIC_URL || process.env.VK_WEB_URL || "";
  const vkTaskUrlTemplate = process.env.VK_TASK_URL_TEMPLATE || "";
  const vkRecoveryCooldownMin = Number(
    process.env.VK_RECOVERY_COOLDOWN_MIN || "10",
  );
  const vkSpawnEnabled =
    !flags.has("no-vk-spawn") && process.env.VK_NO_SPAWN !== "1";
  const vkEnsureIntervalMs = Number(
    cli["vk-ensure-interval"] || process.env.VK_ENSURE_INTERVAL || "60000",
  );

  // ── Telegram ─────────────────────────────────────────────
  const telegramToken = process.env.TELEGRAM_BOT_TOKEN || "";
  const telegramChatId = process.env.TELEGRAM_CHAT_ID || "";
  const telegramIntervalMin = Number(process.env.TELEGRAM_INTERVAL_MIN || "10");
  const telegramCommandPollTimeoutSec = Math.max(
    5,
    Number(process.env.TELEGRAM_COMMAND_POLL_TIMEOUT_SEC || "20"),
  );
  const telegramCommandConcurrency = Math.max(
    1,
    Number(process.env.TELEGRAM_COMMAND_CONCURRENCY || "2"),
  );
  const telegramCommandMaxBatch = Math.max(
    1,
    Number(process.env.TELEGRAM_COMMAND_MAX_BATCH || "25"),
  );
  const telegramBotEnabled = !flags.has("no-telegram-bot") && !!telegramToken;
  const telegramCommandEnabled = flags.has("telegram-commands")
    ? !telegramBotEnabled
    : false;

  // ── Task Planner ─────────────────────────────────────────
  // Mode: "codex-sdk" (default) runs Codex directly, "kanban" creates a VK
  // task for a real agent to plan, "disabled" turns off the planner entirely.
  const plannerMode = (process.env.TASK_PLANNER_MODE || "kanban").toLowerCase();
  const plannerPerCapitaThreshold = Number(
    process.env.TASK_PLANNER_PER_CAPITA_THRESHOLD || "1",
  );
  const plannerIdleSlotThreshold = Number(
    process.env.TASK_PLANNER_IDLE_SLOT_THRESHOLD || "1",
  );
  const plannerDedupHours = Number(process.env.TASK_PLANNER_DEDUP_HOURS || "6");
  const plannerDedupMs = Number.isFinite(plannerDedupHours)
    ? plannerDedupHours * 60 * 60 * 1000
    : 24 * 60 * 60 * 1000;

  // ── Status file ──────────────────────────────────────────
  const cacheDir = resolve(repoRoot, ".cache");
  // Default matches ve-orchestrator.ps1's $script:StatusStatePath
  const statusPath =
    process.env.STATUS_FILE || resolve(cacheDir, "ve-orchestrator-status.json");
  const telegramPollLockPath = resolve(cacheDir, "telegram-getupdates.lock");

  // ── Executors ────────────────────────────────────────────
  const executorConfig = loadExecutorConfig(configDir);
  const scheduler = new ExecutorScheduler(executorConfig);

  // ── Repos ────────────────────────────────────────────────
  const repositories = loadRepoConfig(configDir);

  // ── Agent prompts ────────────────────────────────────────
  const agentPrompts = loadAgentPrompts(configDir, repoRoot);

  // ── First-run detection ──────────────────────────────────
  const isFirstRun =
    !existsSync(resolve(configDir, ".env")) &&
    !existsSync(resolve(configDir, "codex-monitor.config.json")) &&
    !existsSync(resolve(configDir, ".codex-monitor.json"));

  const config = {
    // Identity
    projectName,
    repoSlug,
    repoUrlBase,
    repoRoot,
    configDir,
    envPaths,

    // Orchestrator
    scriptPath,
    scriptArgs,
    restartDelayMs,
    maxRestarts,

    // Logging
    logDir,

    // Feature flags
    watchEnabled,
    watchPath,
    echoLogs,
    autoFixEnabled,
    codexEnabled,

    // Vibe-Kanban
    vkRecoveryPort,
    vkRecoveryHost,
    vkEndpointUrl,
    vkPublicUrl,
    vkTaskUrlTemplate,
    vkRecoveryCooldownMin,
    vkSpawnEnabled,
    vkEnsureIntervalMs,

    // Telegram
    telegramToken,
    telegramChatId,
    telegramIntervalMin,
    telegramCommandPollTimeoutSec,
    telegramCommandConcurrency,
    telegramCommandMaxBatch,
    telegramBotEnabled,
    telegramCommandEnabled,

    // Task Planner
    plannerMode,
    plannerPerCapitaThreshold,
    plannerIdleSlotThreshold,
    plannerDedupHours,
    plannerDedupMs,

    // Paths
    statusPath,
    telegramPollLockPath,
    cacheDir,

    // Executors
    executorConfig,
    scheduler,

    // Multi-repo
    repositories,

    // Agent prompts
    agentPrompts,

    // First run
    isFirstRun,
  };

  return Object.freeze(config);
}

// ── Helpers ──────────────────────────────────────────────────────────────────

function detectProjectName(configDir, repoRoot) {
  // Try package.json in repo root
  const pkgPath = resolve(repoRoot, "package.json");
  if (existsSync(pkgPath)) {
    try {
      const pkg = JSON.parse(readFileSync(pkgPath, "utf8"));
      if (pkg.name) return pkg.name.replace(/^@[^/]+\//, "");
    } catch {
      /* skip */
    }
  }
  // Fallback to directory name
  return basename(repoRoot);
}

function findOrchestratorScript(configDir, repoRoot) {
  // Search for orchestrator scripts in common locations
  const candidates = [
    // Sibling to codex-monitor dir (scripts/ve-orchestrator.ps1)
    resolve(configDir, "..", "ve-orchestrator.ps1"),
    resolve(configDir, "..", "orchestrator.ps1"),
    resolve(configDir, "..", "orchestrator.sh"),
    // Inside codex-monitor dir
    resolve(configDir, "orchestrator.ps1"),
    resolve(configDir, "orchestrator.sh"),
    // Repo root scripts dir
    resolve(repoRoot, "scripts", "ve-orchestrator.ps1"),
    resolve(repoRoot, "scripts", "orchestrator.ps1"),
    resolve(repoRoot, "scripts", "orchestrator.sh"),
    // Repo root
    resolve(repoRoot, "orchestrator.ps1"),
    resolve(repoRoot, "orchestrator.sh"),
    // CWD
    resolve(process.cwd(), "ve-orchestrator.ps1"),
    resolve(process.cwd(), "orchestrator.ps1"),
    resolve(process.cwd(), "scripts", "ve-orchestrator.ps1"),
  ];
  for (const p of candidates) {
    if (existsSync(p)) return p;
  }
  // Default — user will need to configure via --script or ORCHESTRATOR_SCRIPT
  return resolve(configDir, "orchestrator.ps1");
}

// ── Exports ──────────────────────────────────────────────────────────────────

export {
  ExecutorScheduler,
  loadExecutorConfig,
  loadRepoConfig,
  loadAgentPrompts,
};
export default loadConfig;
