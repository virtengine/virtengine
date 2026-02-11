/**
 * @module task-executor
 * @description Internal Task Executor — runs agents locally using the SDK agent pool
 * instead of delegating to VK's cloud executor. Composes kanban-adapter, agent-pool,
 * and worktree-manager to provide full task lifecycle management with configurable
 * parallelism, SDK selection, and thread persistence/resume.
 */

import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import { readFileSync, existsSync } from "node:fs";
import { execSync, spawnSync } from "node:child_process";
import {
  getKanbanAdapter,
  listTasks,
  listProjects,
  getTask,
  updateTaskStatus,
} from "./kanban-adapter.mjs";
import {
  launchOrResumeThread,
  execWithRetry,
  invalidateThread,
  getActiveThreads,
  getPoolSdkName,
} from "./agent-pool.mjs";
import {
  acquireWorktree,
  releaseWorktree,
  getWorktreeStats,
} from "./worktree-manager.mjs";
import { loadConfig } from "./config.mjs";

// ── Constants ───────────────────────────────────────────────────────────────

const TAG = "[task-executor]";
const COOLDOWN_MS = 5 * 60 * 1000; // 5 minutes
const CONTEXT_CACHE_TTL = 10 * 60 * 1000; // 10 minutes
const GRACEFUL_SHUTDOWN_MS = 30_000; // 30 seconds

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// ── Helpers ─────────────────────────────────────────────────────────────────

/**
 * Convert text to a URL/branch-safe slug.
 * @param {string} text
 * @returns {string}
 */
function slugify(text) {
  return (text || "untitled")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 40);
}

// ── Typedefs ────────────────────────────────────────────────────────────────

/**
 * @typedef {Object} TaskExecutorOptions
 * @property {string}   mode            - "internal" | "vk" | "hybrid"
 * @property {number}   maxParallel     - Max concurrent agent slots (default: 3)
 * @property {number}   pollIntervalMs  - How often to check for tasks (default: 30000)
 * @property {string}   sdk             - SDK preference: "codex" | "copilot" | "claude" | "auto"
 * @property {number}   taskTimeoutMs   - Timeout per task execution (default: 90 * 60 * 1000)
 * @property {number}   maxRetries      - Retries per task via execWithRetry (default: 2)
 * @property {boolean}  autoCreatePr    - Create PR after agent completes (default: true)
 * @property {string}   projectId       - VK project ID to poll (null = auto-detect first project)
 * @property {string}   repoRoot        - Repository root path
 * @property {string}   repoSlug        - "owner/repo" for gh CLI
 * @property {Function} onTaskStarted   - callback(task, slotInfo)
 * @property {Function} onTaskCompleted - callback(task, result)
 * @property {Function} onTaskFailed    - callback(task, error)
 * @property {Function} sendTelegram    - optional telegram notifier function
 */

/**
 * @typedef {Object} SlotInfo
 * @property {string} taskId
 * @property {string} taskTitle
 * @property {string} branch
 * @property {string} worktreePath
 * @property {string} threadKey       - agent-pool thread key (taskId used as threadKey)
 * @property {number} startedAt       - timestamp
 * @property {string} sdk             - which SDK is running this
 * @property {number} attempt         - current attempt number
 * @property {"running"|"completing"|"failed"} status
 */

// ── TaskExecutor Class ──────────────────────────────────────────────────────

class TaskExecutor {
  /**
   * Create a new TaskExecutor.
   * @param {TaskExecutorOptions} options
   */
  constructor(options = {}) {
    const defaults = {
      mode: "vk",
      maxParallel: 3,
      pollIntervalMs: 30_000,
      sdk: "auto",
      taskTimeoutMs: 90 * 60 * 1000,
      maxRetries: 2,
      autoCreatePr: true,
      projectId: null,
      repoRoot: process.cwd(),
      repoSlug: "",
      onTaskStarted: null,
      onTaskCompleted: null,
      onTaskFailed: null,
      sendTelegram: null,
    };

    const merged = { ...defaults, ...options };

    this.mode = merged.mode;
    this.maxParallel = merged.maxParallel;
    this.pollIntervalMs = merged.pollIntervalMs;
    this.sdk = merged.sdk;
    this.taskTimeoutMs = merged.taskTimeoutMs;
    this.maxRetries = merged.maxRetries;
    this.autoCreatePr = merged.autoCreatePr;
    this.projectId = merged.projectId;
    this.repoRoot = merged.repoRoot;
    this.repoSlug = merged.repoSlug;
    this.onTaskStarted = merged.onTaskStarted;
    this.onTaskCompleted = merged.onTaskCompleted;
    this.onTaskFailed = merged.onTaskFailed;
    this.sendTelegram = merged.sendTelegram;

    /** @type {Map<string, SlotInfo>} */
    this._activeSlots = new Map();
    /** @type {Map<string, number>} taskId → timestamp */
    this._taskCooldowns = new Map();
    this._running = false;
    this._pollTimer = null;
    this._pollInProgress = false;
    this._resolvedProjectId = null;

    // Repo context cache (AGENTS.md, copilot-instructions.md)
    this._contextCache = null;
    this._contextCacheTime = 0;

    console.log(
      `${TAG} initialized (mode=${this.mode}, maxParallel=${this.maxParallel}, sdk=${this.sdk})`
    );
  }

  // ── Lifecycle ─────────────────────────────────────────────────────────────

  /**
   * Start the periodic poll loop for tasks.
   */
  start() {
    this._running = true;
    // Fire first poll immediately
    this._pollLoop();
    this._pollTimer = setInterval(() => this._pollLoop(), this.pollIntervalMs);
    console.log(
      `${TAG} started — polling every ${this.pollIntervalMs / 1000}s for up to ${this.maxParallel} parallel tasks`
    );
  }

  /**
   * Gracefully stop the executor, waiting for active tasks to finish.
   * @returns {Promise<void>}
   */
  async stop() {
    this._running = false;
    if (this._pollTimer) {
      clearInterval(this._pollTimer);
      this._pollTimer = null;
    }

    const activeCount = this._activeSlots.size;
    if (activeCount > 0) {
      console.log(`${TAG} stopping — waiting for ${activeCount} active task(s) to finish (${GRACEFUL_SHUTDOWN_MS / 1000}s grace)...`);
      const deadline = Date.now() + GRACEFUL_SHUTDOWN_MS;
      while (this._activeSlots.size > 0 && Date.now() < deadline) {
        await new Promise((r) => setTimeout(r, 1000));
      }
    }

    console.log(`${TAG} stopped (${this._activeSlots.size} tasks were active)`);
  }

  /**
   * Returns the current executor status for monitoring / Telegram.
   * @returns {Object}
   */
  getStatus() {
    return {
      running: this._running,
      mode: this.mode,
      maxParallel: this.maxParallel,
      sdk: this.sdk === "auto" ? getPoolSdkName() : this.sdk,
      activeSlots: this._activeSlots.size,
      slots: Array.from(this._activeSlots.values()).map((s) => ({
        taskId: s.taskId,
        taskTitle: s.taskTitle,
        branch: s.branch,
        sdk: s.sdk,
        attempt: s.attempt,
        runningFor: Math.round((Date.now() - s.startedAt) / 1000),
        status: s.status,
      })),
      cooldowns: this._taskCooldowns.size,
      pollIntervalMs: this.pollIntervalMs,
      taskTimeoutMs: this.taskTimeoutMs,
      maxRetries: this.maxRetries,
      projectId: this._resolvedProjectId || this.projectId || null,
    };
  }

  // ── Poll Loop ─────────────────────────────────────────────────────────────

  /**
   * Check kanban for todo tasks and dispatch execution.
   * Guarded against overlapping polls and slot saturation.
   * @private
   */
  async _pollLoop() {
    if (!this._running) return;
    if (this._pollInProgress) return;
    if (this._activeSlots.size >= this.maxParallel) return;

    this._pollInProgress = true;
    try {
      // Resolve project ID on first poll
      if (!this._resolvedProjectId) {
        if (this.projectId) {
          this._resolvedProjectId = this.projectId;
        } else {
          try {
            const projects = await listProjects();
            if (projects && projects.length > 0) {
              // Match by PROJECT_NAME if set, otherwise fall back to first project
              const wantName = (process.env.PROJECT_NAME || process.env.VK_PROJECT_NAME || "").toLowerCase();
              let matched;
              if (wantName) {
                matched = projects.find(
                  (p) => (p.name || p.title || "").toLowerCase() === wantName
                );
              }
              if (matched) {
                this._resolvedProjectId = matched.id || matched.project_id;
                console.log(`${TAG} matched project by name "${wantName}": ${this._resolvedProjectId}`);
              } else {
                this._resolvedProjectId = projects[0].id || projects[0].project_id;
                console.log(`${TAG} auto-detected project (first): ${this._resolvedProjectId}`);
              }
            } else {
              console.warn(`${TAG} no projects found — skipping poll`);
              return;
            }
          } catch (err) {
            console.warn(`${TAG} failed to list projects: ${err.message}`);
            return;
          }
        }
      }

      // Fetch todo tasks
      let tasks;
      try {
        tasks = await listTasks(this._resolvedProjectId, { status: "todo" });
      } catch (err) {
        console.warn(`${TAG} failed to list tasks: ${err.message}`);
        return;
      }

      if (!tasks || tasks.length === 0) return;

      const now = Date.now();

      // Filter out ineligible tasks
      const eligible = tasks.filter((t) => {
        const id = t.id || t.task_id;
        if (!id) return false;
        // Already running
        if (this._activeSlots.has(id)) return false;
        // In cooldown
        const cooldownUntil = this._taskCooldowns.get(id);
        if (cooldownUntil && now - cooldownUntil < COOLDOWN_MS) return false;
        // Must have a branch name derivable
        // (we can auto-generate one, so this is not strictly required)
        return true;
      });

      if (eligible.length === 0) return;

      // Fill remaining slots
      const remaining = this.maxParallel - this._activeSlots.size;
      const toDispatch = eligible.slice(0, remaining);

      for (const task of toDispatch) {
        // Normalize task id
        task.id = task.id || task.task_id;
        // Fire and forget — executeTask handles its own lifecycle
        this.executeTask(task).catch((err) => {
          console.error(`${TAG} unhandled error in executeTask for "${task.title}": ${err.message}`);
        });
      }
    } catch (err) {
      console.error(`${TAG} poll loop error: ${err.message}`);
    } finally {
      this._pollInProgress = false;
    }
  }

  // ── Task Execution ────────────────────────────────────────────────────────

  /**
   * Execute a single task through its full lifecycle:
   * slot allocation → status update → worktree → agent → result → cleanup.
   * @param {Object} task - Task object from kanban adapter
   * @returns {Promise<void>}
   */
  async executeTask(task) {
    const taskId = task.id || task.task_id;
    const taskTitle = task.title || "(untitled)";
    const branch =
      task.branchName ||
      task.meta?.branch_name ||
      `ve/${taskId.substring(0, 8)}-${slugify(taskTitle)}`;

    // 1. Allocate slot
    /** @type {SlotInfo} */
    const slot = {
      taskId,
      taskTitle,
      branch,
      worktreePath: null,
      threadKey: taskId,
      startedAt: Date.now(),
      sdk: this.sdk === "auto" ? getPoolSdkName() : this.sdk,
      attempt: 0,
      status: "running",
    };
    this._activeSlots.set(taskId, slot);

    try {
      this.onTaskStarted?.(task, slot);

      // 2. Update task status → "inprogress"
      try {
        await updateTaskStatus(taskId, "inprogress");
      } catch (err) {
        console.warn(`${TAG} failed to set task to inprogress: ${err.message}`);
      }

      // 3. Acquire worktree
      let wt;
      try {
        wt = await acquireWorktree(branch, taskId, {
          owner: "task-executor",
          baseBranch: "main",
        });
      } catch (err) {
        console.error(`${TAG} worktree acquisition failed for "${taskTitle}": ${err.message}`);
        this._taskCooldowns.set(taskId, Date.now());
        try {
          await updateTaskStatus(taskId, "todo");
        } catch { /* best-effort */ }
        this._activeSlots.delete(taskId);
        this.onTaskFailed?.(task, new Error(`Worktree acquisition failed: ${err.message}`));
        return;
      }

      if (!wt || !wt.path || !existsSync(wt.path)) {
        console.error(`${TAG} worktree path invalid for "${taskTitle}": ${wt?.path}`);
        this._taskCooldowns.set(taskId, Date.now());
        try {
          await releaseWorktree(taskId);
        } catch { /* best-effort */ }
        try {
          await updateTaskStatus(taskId, "todo");
        } catch { /* best-effort */ }
        this._activeSlots.delete(taskId);
        this.onTaskFailed?.(task, new Error("Worktree path invalid or missing"));
        return;
      }

      slot.worktreePath = wt.path;

      // 4. Build prompt
      const prompt = this._buildTaskPrompt(task, wt.path);

      // 5. Execute agent
      console.log(`${TAG} executing task "${taskTitle}" in ${wt.path} on branch ${branch}`);
      const result = await execWithRetry(prompt, {
        taskKey: taskId,
        cwd: wt.path,
        timeoutMs: this.taskTimeoutMs,
        maxRetries: this.maxRetries,
        sdk: this.sdk !== "auto" ? this.sdk : undefined,
        buildRetryPrompt: (lastResult, attempt) =>
          this._buildRetryPrompt(task, lastResult, attempt),
      });

      // Track attempts on task for PR body
      task._executionResult = result;

      // 6. Handle result
      slot.status = result.success ? "completing" : "failed";
      await this._handleTaskResult(task, result, wt.path);

      // 7. Cleanup
      try {
        await releaseWorktree(taskId);
      } catch (err) {
        console.warn(`${TAG} worktree release warning: ${err.message}`);
      }
      this._activeSlots.delete(taskId);
    } catch (err) {
      // Catch-all: ensure slot is always cleaned up
      console.error(`${TAG} fatal error executing task "${taskTitle}": ${err.message}`);
      slot.status = "failed";
      this._taskCooldowns.set(taskId, Date.now());

      try {
        await updateTaskStatus(taskId, "todo");
      } catch { /* best-effort */ }
      try {
        await releaseWorktree(taskId);
      } catch { /* best-effort */ }

      this._activeSlots.delete(taskId);
      this.onTaskFailed?.(task, err);
      this.sendTelegram?.(
        `❌ Task executor error: "${taskTitle}" — ${(err.message || "").slice(0, 200)}`
      );
    }
  }

  // ── Prompt Building ───────────────────────────────────────────────────────

  /**
   * Build a comprehensive prompt for the agent from task details and repo context.
   * @param {Object} task
   * @param {string} worktreePath
   * @returns {string}
   * @private
   */
  _buildTaskPrompt(task, worktreePath) {
    const branch =
      task.branchName ||
      task.meta?.branch_name ||
      spawnSync("git", ["branch", "--show-current"], {
        cwd: worktreePath,
        encoding: "utf8",
        timeout: 5000,
      }).stdout?.trim() ||
      "unknown";

    const lines = [
      `# Task: ${task.title}`,
      ``,
      `## Description`,
      task.description || "No description provided. Check the task URL for details.",
      ``,
      `## Environment`,
      `- Working Directory: ${worktreePath}`,
      `- Branch: ${branch}`,
      `- Repository: ${this.repoSlug}`,
      ``,
      `## Instructions`,
      `You are working autonomously on a VirtEngine blockchain task.`,
      `1. Read the task description carefully`,
      `2. Analyze the codebase to understand what needs to change`,
      `3. Implement the required changes`,
      `4. Run tests to verify: \`go test ./x/... ./pkg/...\` or relevant package tests`,
      `5. Run linting: \`golangci-lint run\` on changed packages`,
      `6. Commit your changes using conventional commit format: type(scope): description`,
      `7. Push your branch: \`git push --set-upstream origin ${branch}\``,
      ``,
      `## Critical Rules`,
      `- NEVER ask for user input — this is an autonomous task`,
      `- NEVER create placeholders or stubs — implement real, complete code`,
      `- NEVER skip tests — verify your changes work`,
      `- Use conventional commits: feat|fix|docs|refactor|test(scope): description`,
      `- Follow existing code patterns in the repository`,
      ``,
    ];

    // Append task URL if available
    const taskUrl = task.meta?.task_url || task.taskUrl || task.url;
    if (taskUrl) {
      lines.push(`## Task Reference`, `- URL: ${taskUrl}`, ``);
    }

    // Append cached repo context (AGENTS.md + copilot-instructions.md)
    const context = this._getRepoContext();
    if (context) {
      lines.push(`## Repository Context`, ``, context, ``);
    }

    return lines.join("\n");
  }

  /**
   * Build a retry prompt for the agent after a failed attempt.
   * @param {Object} task
   * @param {Object} lastResult
   * @param {number} attemptNumber
   * @returns {string}
   * @private
   */
  _buildRetryPrompt(task, lastResult, attemptNumber) {
    return [
      `# ERROR RECOVERY — Attempt ${attemptNumber}`,
      ``,
      `Your previous attempt on task "${task.title}" encountered an issue:`,
      "```",
      (lastResult?.error || lastResult?.output || "(unknown error)").slice(0, 3000),
      "```",
      ``,
      `Please:`,
      `1. Diagnose what went wrong`,
      `2. Fix the issue`,
      `3. Re-run tests to verify`,
      `4. Commit and push your fixes`,
      ``,
      `Original task description:`,
      task.description || "See task URL for details.",
    ].join("\n");
  }

  /**
   * Load and cache repo context files (AGENTS.md, copilot-instructions.md).
   * Cached for CONTEXT_CACHE_TTL to avoid re-reading on every task.
   * @returns {string|null}
   * @private
   */
  _getRepoContext() {
    const now = Date.now();
    if (this._contextCache && now - this._contextCacheTime < CONTEXT_CACHE_TTL) {
      return this._contextCache;
    }

    const parts = [];
    const contextFiles = [
      { rel: "AGENTS.md", label: "AGENTS.md" },
      { rel: ".github/copilot-instructions.md", label: "Copilot Instructions" },
    ];

    for (const cf of contextFiles) {
      try {
        const fullPath = resolve(this.repoRoot, cf.rel);
        if (existsSync(fullPath)) {
          const content = readFileSync(fullPath, "utf8");
          // Truncate to 4000 chars to keep prompt reasonable
          const truncated = content.length > 4000
            ? content.slice(0, 4000) + "\n...(truncated)"
            : content;
          parts.push(`### ${cf.label}\n\n${truncated}`);
        }
      } catch {
        // Ignore read errors
      }
    }

    this._contextCache = parts.length > 0 ? parts.join("\n\n---\n\n") : null;
    this._contextCacheTime = now;
    return this._contextCache;
  }

  // ── Result Handling ───────────────────────────────────────────────────────

  /**
   * Handle the result of a task execution — PR creation, status update, notifications.
   * @param {Object} task
   * @param {Object} result - { success, attempts, error, output }
   * @param {string} worktreePath
   * @returns {Promise<void>}
   * @private
   */
  async _handleTaskResult(task, result, worktreePath) {
    const taskTitle = (task.title || "").slice(0, 50);
    const tag = `${TAG} task "${taskTitle}"`;

    if (result.success) {
      console.log(`${tag} completed successfully (${result.attempts} attempt(s))`);

      // Check if there are commits to push
      const hasCommits = this._hasUnpushedCommits(worktreePath);

      if (hasCommits && this.autoCreatePr) {
        const pr = await this._createPR(task, worktreePath);
        if (pr) {
          try {
            await updateTaskStatus(task.id, "inreview");
          } catch { /* best-effort */ }
          this.sendTelegram?.(
            `✅ Task completed: "${task.title}"\nPR: ${pr.url || pr}`
          );
        } else {
          try {
            await updateTaskStatus(task.id, "inreview");
          } catch { /* best-effort */ }
          this.sendTelegram?.(
            `✅ Task completed: "${task.title}" (PR creation failed — manual review needed)`
          );
        }
      } else if (hasCommits) {
        try {
          await updateTaskStatus(task.id, "inreview");
        } catch { /* best-effort */ }
        this.sendTelegram?.(
          `✅ Task completed: "${task.title}" (auto-PR disabled)`
        );
      } else {
        // No commits — agent may have determined no changes needed
        console.warn(`${tag} completed but no commits found`);
        try {
          await updateTaskStatus(task.id, "inreview");
        } catch { /* best-effort */ }
        this.sendTelegram?.(
          `⚠️ Task completed but no commits: "${task.title}"`
        );
      }

      this.onTaskCompleted?.(task, result);
    } else {
      console.warn(
        `${tag} failed after ${result.attempts} attempt(s): ${result.error}`
      );
      this._taskCooldowns.set(task.id, Date.now());
      try {
        await updateTaskStatus(task.id, "todo");
      } catch { /* best-effort */ }
      this.sendTelegram?.(
        `❌ Task failed: "${task.title}" — ${(result.error || "").slice(0, 200)}`
      );
      this.onTaskFailed?.(task, result);
    }
  }

  // ── Git Helpers ───────────────────────────────────────────────────────────

  /**
   * Check whether a worktree has unpushed commits.
   * @param {string} worktreePath
   * @returns {boolean}
   * @private
   */
  _hasUnpushedCommits(worktreePath) {
    try {
      const result = spawnSync("git", ["log", "@{u}..HEAD", "--oneline"], {
        cwd: worktreePath,
        encoding: "utf8",
        timeout: 10_000,
      });
      // If upstream not set, check if HEAD differs from main
      if (result.status !== 0) {
        const diff = spawnSync("git", ["log", "main..HEAD", "--oneline"], {
          cwd: worktreePath,
          encoding: "utf8",
          timeout: 10_000,
        });
        return diff.status === 0 && (diff.stdout || "").trim().length > 0;
      }
      return (result.stdout || "").trim().length > 0;
    } catch {
      return false;
    }
  }

  /**
   * Create a pull request for the completed task using the gh CLI.
   * @param {Object} task
   * @param {string} worktreePath
   * @returns {Promise<{url: string, branch: string}|null>}
   * @private
   */
  async _createPR(task, worktreePath) {
    try {
      const branch =
        task.branchName ||
        spawnSync("git", ["branch", "--show-current"], {
          cwd: worktreePath,
          encoding: "utf8",
          timeout: 5000,
        }).stdout?.trim();

      const title = task.title;
      const body = [
        `## Summary`,
        task.description ? task.description.slice(0, 2000) : "Automated task completion.",
        ``,
        `## Task Reference`,
        `- Task ID: \`${task.id}\``,
        task.meta?.task_url ? `- Task URL: ${task.meta.task_url}` : "",
        ``,
        `## Executor`,
        `- Mode: internal (task-executor)`,
        `- SDK: ${getPoolSdkName()}`,
        `- Attempts: ${task._executionResult?.attempts || "N/A"}`,
      ]
        .filter(Boolean)
        .join("\n");

      const result = spawnSync(
        "gh",
        [
          "pr",
          "create",
          "--title",
          title,
          "--body",
          body,
          "--head",
          branch,
          "--base",
          "main",
        ],
        {
          cwd: worktreePath,
          encoding: "utf8",
          timeout: 30_000,
          env: { ...process.env },
        }
      );

      if (result.status === 0) {
        const prUrl = (result.stdout || "").trim();
        console.log(`${TAG} PR created: ${prUrl}`);
        return { url: prUrl, branch };
      } else {
        const stderr = (result.stderr || "").trim();
        // PR might already exist
        if (stderr.includes("already exists")) {
          console.log(`${TAG} PR already exists for ${branch}`);
          return { url: "(existing)", branch };
        }
        console.warn(`${TAG} PR creation failed: ${stderr}`);
        return null;
      }
    } catch (err) {
      console.warn(`${TAG} PR creation error: ${err.message}`);
      return null;
    }
  }
}

// ── Singleton & Module Exports ──────────────────────────────────────────────

/** @type {TaskExecutor|null} */
let _instance = null;

/**
 * Get or create the singleton TaskExecutor.
 * @param {TaskExecutorOptions} [opts]
 * @returns {TaskExecutor}
 */
export function getTaskExecutor(opts) {
  if (!_instance) {
    _instance = new TaskExecutor(opts || loadExecutorOptionsFromConfig());
  }
  return _instance;
}

/**
 * Load executor options from config/env.
 * @returns {TaskExecutorOptions}
 */
export function loadExecutorOptionsFromConfig() {
  let config;
  try {
    config = loadConfig();
  } catch {
    config = {};
  }

  const envMode = (process.env.EXECUTOR_MODE || "").toLowerCase();
  const configExec = config.internalExecutor || config.taskExecutor || {};

  return {
    mode: envMode || configExec.mode || "vk",
    maxParallel: Number(
      process.env.INTERNAL_EXECUTOR_PARALLEL || configExec.maxParallel || 3
    ),
    pollIntervalMs: Number(
      process.env.INTERNAL_EXECUTOR_POLL_MS || configExec.pollIntervalMs || 30_000
    ),
    sdk: process.env.INTERNAL_EXECUTOR_SDK || configExec.sdk || "auto",
    taskTimeoutMs: Number(
      process.env.INTERNAL_EXECUTOR_TIMEOUT_MS ||
        configExec.taskTimeoutMs ||
        90 * 60 * 1000
    ),
    maxRetries: Number(
      process.env.INTERNAL_EXECUTOR_MAX_RETRIES || configExec.maxRetries || 2
    ),
    autoCreatePr: configExec.autoCreatePr !== false,
    projectId:
      process.env.INTERNAL_EXECUTOR_PROJECT_ID || configExec.projectId || null,
    repoRoot: config.repoRoot || process.cwd(),
    repoSlug: config.repoSlug || "",
  };
}

/**
 * Check whether internal executor mode is enabled.
 * @returns {boolean}
 */
export function isInternalExecutorEnabled() {
  const mode = (process.env.EXECUTOR_MODE || "").toLowerCase();
  if (mode === "internal" || mode === "hybrid") return true;
  try {
    const config = loadConfig();
    const configMode = (
      config.internalExecutor?.mode ||
      config.taskExecutor?.mode ||
      ""
    ).toLowerCase();
    return configMode === "internal" || configMode === "hybrid";
  } catch {
    return false;
  }
}

/**
 * Convenience: get executor mode.
 * @returns {"vk"|"internal"|"hybrid"}
 */
export function getExecutorMode() {
  const mode = (process.env.EXECUTOR_MODE || "").toLowerCase();
  if (["vk", "internal", "hybrid"].includes(mode)) return mode;
  try {
    const config = loadConfig();
    const configMode = (
      config.internalExecutor?.mode ||
      config.taskExecutor?.mode ||
      "vk"
    ).toLowerCase();
    if (["vk", "internal", "hybrid"].includes(configMode)) return configMode;
  } catch {
    /* ignore */
  }
  return "vk";
}

export { TaskExecutor };
export default TaskExecutor;
