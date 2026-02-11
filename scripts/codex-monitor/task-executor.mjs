/**
 * @module task-executor
 * @description Internal Task Executor — runs agents locally using the SDK agent pool
 * instead of delegating to VK's cloud executor. Composes kanban-adapter, agent-pool,
 * and worktree-manager to provide full task lifecycle management with configurable
 * parallelism, SDK selection, and thread persistence/resume.
 */

import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import { readFileSync, existsSync, appendFileSync, mkdirSync, writeFileSync } from "node:fs";
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
  forceNewThread,
  pruneAllExhaustedThreads,
  getActiveThreads,
  getPoolSdkName,
} from "./agent-pool.mjs";
import {
  acquireWorktree,
  releaseWorktree,
  getWorktreeStats,
} from "./worktree-manager.mjs";
import { loadConfig } from "./config.mjs";
import {
  loadStore as loadTaskStore,
  setTaskStatus as setInternalStatus,
  recordAgentAttempt,
  recordErrorPattern,
  setTaskCooldown,
  clearTaskCooldown,
  isTaskCoolingDown,
  updateTask as updateInternalTask,
  getTask as getInternalTask,
} from "./task-store.mjs";
import { createErrorDetector } from "./error-detector.mjs";

// ── Constants ───────────────────────────────────────────────────────────────

const TAG = "[task-executor]";
const COOLDOWN_MS = 5 * 60 * 1000; // 5 minutes
const CONTEXT_CACHE_TTL = 10 * 60 * 1000; // 10 minutes
const GRACEFUL_SHUTDOWN_MS = 30_000; // 30 seconds
const MAX_NO_COMMIT_ATTEMPTS = 3; // Stop picking up a task after N consecutive no-commit completions
const NO_COMMIT_COOLDOWN_BASE_MS = 15 * 60 * 1000; // 15 minutes base cooldown for no-commit
const NO_COMMIT_MAX_COOLDOWN_MS = 2 * 60 * 60 * 1000; // 2 hours max cooldown
const NO_COMMIT_STATE_FILE = resolve(dirname(fileURLToPath(import.meta.url)), ".cache", "no-commit-state.json");

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// ── Agent Log Streaming ─────────────────────────────────────────────────────

const AGENT_LOGS_DIR = resolve(__dirname, "logs", "agents");

/**
 * Create an onEvent callback that streams agent SDK events to a per-task log file.
 * @param {string} taskId
 * @param {string} taskTitle
 * @returns {Function}
 */
function createAgentLogStreamer(taskId, taskTitle) {
  const shortId = taskId.substring(0, 8);
  const logFile = resolve(AGENT_LOGS_DIR, `agent-${shortId}.log`);

  // Ensure log dir exists
  try { mkdirSync(AGENT_LOGS_DIR, { recursive: true }); } catch { /* ok */ }

  // Write header
  try {
    appendFileSync(logFile, `\n${"=".repeat(80)}\n[${new Date().toISOString()}] Task: ${taskTitle}\nTask ID: ${taskId}\n${"=".repeat(80)}\n`, "utf8");
  } catch { /* ok */ }

  return (event) => {
    try {
      const ts = new Date().toISOString();
      if (event.type === "item.completed" && event.item) {
        const item = event.item;
        if (item.type === "agent_message" && item.text) {
          appendFileSync(logFile, `[${ts}] AGENT: ${item.text.slice(0, 2000)}\n`, "utf8");
        } else if (item.type === "function_call") {
          appendFileSync(logFile, `[${ts}] TOOL: ${item.name}(${(item.arguments || "").slice(0, 200)})\n`, "utf8");
        } else if (item.type === "function_call_output") {
          const out = (item.output || "").slice(0, 500);
          appendFileSync(logFile, `[${ts}] RESULT: ${out}\n`, "utf8");
        } else {
          appendFileSync(logFile, `[${ts}] ITEM[${item.type}]: ${JSON.stringify(item).slice(0, 300)}\n`, "utf8");
        }
      } else if (event.type === "item.created") {
        const item = event.item || {};
        appendFileSync(logFile, `[${ts}] +${item.type || event.type}\n`, "utf8");
      } else if (event.type) {
        // Log any other event type for debugging
        appendFileSync(logFile, `[${ts}] EVT[${event.type}]\n`, "utf8");
      }
    } catch { /* never let logging crash the agent */ }
  };
}

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

    // Anti-thrash: track consecutive no-commit completions per task
    /** @type {Map<string, number>} taskId → consecutive no-commit count */
    this._noCommitCounts = new Map();
    /** @type {Map<string, number>} taskId → skip-until timestamp */
    this._skipUntil = new Map();

    // Repo context cache (AGENTS.md, copilot-instructions.md)
    this._contextCache = null;
    this._contextCacheTime = 0;

    // Error detector for classifying agent failures
    this._errorDetector = createErrorDetector({
      sendTelegram: this.sendTelegram,
      onErrorDetected: (taskId, classification) => {
        console.log(`${TAG} error detected for ${taskId}: ${classification.pattern} (${classification.confidence.toFixed(2)})`);
      },
    });

    console.log(
      `${TAG} initialized (mode=${this.mode}, maxParallel=${this.maxParallel}, sdk=${this.sdk})`
    );
  }

  /** Load anti-thrash state from disk (survives restarts). */
  _loadNoCommitState() {
    try {
      if (existsSync(NO_COMMIT_STATE_FILE)) {
        const raw = readFileSync(NO_COMMIT_STATE_FILE, "utf8");
        const data = JSON.parse(raw);
        if (data && typeof data === "object") {
          for (const [id, count] of Object.entries(data.noCommitCounts || {})) {
            this._noCommitCounts.set(id, count);
          }
          for (const [id, until] of Object.entries(data.skipUntil || {})) {
            if (until > Date.now()) {
              this._skipUntil.set(id, until);
            }
          }
          console.log(`${TAG} restored anti-thrash state: ${this._noCommitCounts.size} tasks tracked`);
        }
      }
    } catch (err) {
      console.warn(`${TAG} failed to load anti-thrash state: ${err.message}`);
    }
  }

  /** Persist anti-thrash state to disk. */
  _saveNoCommitState() {
    try {
      const dir = resolve(__dirname, ".cache");
      mkdirSync(dir, { recursive: true });
      const data = {
        noCommitCounts: Object.fromEntries(this._noCommitCounts),
        skipUntil: Object.fromEntries(this._skipUntil),
        savedAt: new Date().toISOString(),
      };
      writeFileSync(NO_COMMIT_STATE_FILE, JSON.stringify(data, null, 2), "utf8");
    } catch (err) {
      console.warn(`${TAG} failed to save anti-thrash state: ${err.message}`);
    }
  }

  // ── Lifecycle ─────────────────────────────────────────────────────────────

  /**
   * Start the periodic poll loop for tasks.
   */
  start() {
    // Load internal task store
    try { loadTaskStore(); } catch (err) { console.warn(`${TAG} task store load warning: ${err.message}`); }
    // Restore anti-thrash state from disk
    this._loadNoCommitState();

    // Clean up zombie threads from prior runs
    const pruned = pruneAllExhaustedThreads();
    if (pruned > 0) {
      console.log(`${TAG} cleaned up ${pruned} stale agent threads on startup`);
    }

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
      blockedTasks: this._getBlockedTaskIds(),
      noCommitCounts: Object.fromEntries(this._noCommitCounts),
      pollIntervalMs: this.pollIntervalMs,
      taskTimeoutMs: this.taskTimeoutMs,
      maxRetries: this.maxRetries,
      projectId: this._resolvedProjectId || this.projectId || null,
    };
  }

  /**
   * Check if a task is currently managed by the internal executor
   * (active, in cooldown, or blocked). Used by monitor to avoid
   * double-recovering tasks.
   * @param {string} taskId
   * @returns {boolean}
   */
  isTaskManaged(taskId) {
    // Currently executing
    if (this._activeSlots.has(taskId)) return true;
    // In anti-thrash cooldown
    const skipUntil = this._skipUntil.get(taskId);
    if (skipUntil && Date.now() < skipUntil) return true;
    // In failure cooldown
    const cooldownAt = this._taskCooldowns.get(taskId);
    if (cooldownAt && Date.now() - cooldownAt < COOLDOWN_MS) return true;
    // Permanently blocked for this session
    const noCommitCount = this._noCommitCounts.get(taskId) || 0;
    if (noCommitCount >= MAX_NO_COMMIT_ATTEMPTS) return true;
    return false;
  }

  /**
   * Get list of task IDs that are permanently blocked (exceeded no-commit limit).
   * @returns {string[]}
   * @private
   */
  _getBlockedTaskIds() {
    const blocked = [];
    for (const [id, count] of this._noCommitCounts) {
      if (count >= MAX_NO_COMMIT_ATTEMPTS) blocked.push(id);
    }
    return blocked;
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
        // In cooldown (failure cooldown)
        const cooldownUntil = this._taskCooldowns.get(id);
        if (cooldownUntil && now - cooldownUntil < COOLDOWN_MS) return false;
        // Anti-thrash: skip tasks that repeatedly complete with no commits
        const skipUntil = this._skipUntil.get(id);
        if (skipUntil && now < skipUntil) {
          return false; // still in anti-thrash cooldown
        } else if (skipUntil && now >= skipUntil) {
          this._skipUntil.delete(id); // cooldown expired, allow retry
        }
        // Hard block: exceeded max no-commit attempts
        const noCommitCount = this._noCommitCounts.get(id) || 0;
        if (noCommitCount >= MAX_NO_COMMIT_ATTEMPTS) {
          return false; // permanently blocked for this executor session
        }
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
      // Mirror to internal store
      try { setInternalStatus(taskId, "inprogress", "task-executor"); } catch { /* best-effort */ }

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
        onEvent: createAgentLogStreamer(taskId, taskTitle),
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

    // Agent endpoint info for self-reporting
    const endpointPort = process.env.AGENT_ENDPOINT_PORT || "18432";
    lines.push(
      `## Agent Status Endpoint`,
      `You can report your status to the orchestrator at: http://127.0.0.1:${endpointPort}/api/tasks/${task.id || task.task_id}`,
      `- POST /status with {"status": "inreview"} when you've pushed and created a PR`,
      `- POST /heartbeat with {} to indicate you're still alive`,
      `- POST /error with {"error": "description"} if you encounter a fatal error`,
      `- POST /complete with {"hasCommits": true} when fully done`,
      ``,
    );

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
    // Check for plan-stuck pattern
    const classification = this._errorDetector.classify(
      lastResult?.output || "",
      lastResult?.error || ""
    );

    if (classification.pattern === "plan_stuck") {
      return this._errorDetector.getPlanStuckRecoveryPrompt(
        task.title,
        lastResult?.output || ""
      );
    }

    if (classification.pattern === "token_overflow") {
      return this._errorDetector.getTokenOverflowRecoveryPrompt(task.title);
    }

    // Default retry prompt
    return [
      `# ERROR RECOVERY — Attempt ${attemptNumber}`,
      ``,
      `Your previous attempt on task "${task.title}" encountered an issue:`,
      "```",
      (lastResult?.error || lastResult?.output || "(unknown error)").slice(0, 3000),
      "```",
      ``,
      `Error classification: ${classification.pattern} (confidence: ${classification.confidence.toFixed(2)})`,
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
        // Real work done — reset the no-commit counter
        this._noCommitCounts.delete(task.id);
        this._skipUntil.delete(task.id);
        this._saveNoCommitState();

        // Record success in internal store
        try {
          recordAgentAttempt(task.id, { output: result.output, hasCommits: true });
          setInternalStatus(task.id, "inreview", "task-executor");
          this._errorDetector.resetTask(task.id);
        } catch { /* best-effort */ }

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
        // Real work done — reset the no-commit counter
        this._noCommitCounts.delete(task.id);
        this._skipUntil.delete(task.id);
        this._saveNoCommitState();

        // Record success in internal store
        try {
          recordAgentAttempt(task.id, { output: result.output, hasCommits: true });
          setInternalStatus(task.id, "inreview", "task-executor");
          this._errorDetector.resetTask(task.id);
        } catch { /* best-effort */ }

        try {
          await updateTaskStatus(task.id, "inreview");
        } catch { /* best-effort */ }
        this.sendTelegram?.(
          `✅ Task completed: "${task.title}" (auto-PR disabled)`
        );
      } else {
        // No commits — agent completed without making changes.
        // This is NOT a real completion. Apply anti-thrash protection.
        const prevCount = this._noCommitCounts.get(task.id) || 0;
        const noCommitCount = prevCount + 1;
        this._noCommitCounts.set(task.id, noCommitCount);
        this._saveNoCommitState();

        // Force fresh thread on next attempt — the current thread is clearly not productive
        try { forceNewThread(task.id, `no-commit completion #${noCommitCount}`); } catch { /* ok */ }

        // Record no-commit attempt in internal store
        try {
          recordAgentAttempt(task.id, { output: result.output, hasCommits: false });
          const noCommitClassification = this._errorDetector.classify(result.output || "");
          if (noCommitClassification.pattern === "plan_stuck") {
            recordErrorPattern(task.id, "plan_stuck");
          }
        } catch { /* best-effort */ }

        // Escalating cooldown: 15min → 30min → 1h → 2h (capped)
        const cooldownMs = Math.min(
          NO_COMMIT_COOLDOWN_BASE_MS * Math.pow(2, noCommitCount - 1),
          NO_COMMIT_MAX_COOLDOWN_MS,
        );
        const cooldownMin = Math.round(cooldownMs / 60_000);
        this._skipUntil.set(task.id, Date.now() + cooldownMs);
        this._taskCooldowns.set(task.id, Date.now());

        console.warn(
          `${tag} completed but no commits found (attempt ${noCommitCount}/${MAX_NO_COMMIT_ATTEMPTS}, cooldown ${cooldownMin}m)`,
        );

        // Set back to todo — NOT inreview (nothing to review)
        try {
          await updateTaskStatus(task.id, "todo");
        } catch { /* best-effort */ }

        if (noCommitCount >= MAX_NO_COMMIT_ATTEMPTS) {
          console.warn(
            `${tag} task "${task.title}" blocked — ${MAX_NO_COMMIT_ATTEMPTS} consecutive no-commit completions. Skipping until executor restart.`,
          );
          this.sendTelegram?.(
            `🚫 Task blocked (${MAX_NO_COMMIT_ATTEMPTS}x no-commit): "${task.title}" — will not retry until executor restart`,
          );
        } else {
          this.sendTelegram?.(
            `⚠️ Task completed but no commits (${noCommitCount}/${MAX_NO_COMMIT_ATTEMPTS}): "${task.title}" — cooldown ${cooldownMin}m`,
          );
        }
      }

      this.onTaskCompleted?.(task, result);
    } else {
      console.warn(
        `${tag} failed after ${result.attempts} attempt(s): ${result.error}`
      );
      // Invalidate thread so next attempt starts fresh
      try { forceNewThread(task.id, `task failed: ${(result.error || "").slice(0, 100)}`); } catch { /* ok */ }
      this._taskCooldowns.set(task.id, Date.now());

      // Classify the error
      const classification = this._errorDetector.classify(
        result.output || "",
        result.error || ""
      );
      const recovery = this._errorDetector.recordError(task.id, classification);

      // Record in internal store
      try {
        recordAgentAttempt(task.id, { output: result.output, error: result.error, hasCommits: false });
        recordErrorPattern(task.id, classification.pattern);
      } catch { /* best-effort */ }

      // If plan-stuck, use recovery prompt instead of generic retry
      if (classification.pattern === "plan_stuck" && recovery.action === "retry_with_prompt") {
        console.log(`${TAG} plan-stuck detected — will use recovery prompt on next attempt`);
      }

      // If rate limiting, check executor pause
      if (this._errorDetector.shouldPauseExecutor()) {
        console.warn(`${TAG} too many rate limits — pausing executor for 5 minutes`);
        this._running = false;
        setTimeout(() => {
          this._running = true;
          console.log(`${TAG} executor resumed after rate limit pause`);
        }, 5 * 60 * 1000);
      }

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
      // Method 1: Check vs upstream tracking branch
      const result = spawnSync("git", ["log", "@{u}..HEAD", "--oneline"], {
        cwd: worktreePath,
        encoding: "utf8",
        timeout: 10_000,
      });
      if (result.status === 0 && (result.stdout || "").trim().length > 0) {
        return true;
      }

      // Method 2: Check vs origin/main (fetch first to be current)
      try {
        spawnSync("git", ["fetch", "origin", "main", "--quiet"], {
          cwd: worktreePath,
          encoding: "utf8",
          timeout: 15_000,
        });
      } catch { /* best-effort */ }

      const diff = spawnSync("git", ["log", "origin/main..HEAD", "--oneline"], {
        cwd: worktreePath,
        encoding: "utf8",
        timeout: 10_000,
      });
      if (diff.status === 0 && (diff.stdout || "").trim().length > 0) {
        return true;
      }

      // Method 3: Fallback — check if there are ANY commits not in main
      const diff2 = spawnSync("git", ["log", "main..HEAD", "--oneline"], {
        cwd: worktreePath,
        encoding: "utf8",
        timeout: 10_000,
      });
      return diff2.status === 0 && (diff2.stdout || "").trim().length > 0;
    } catch {
      return false;
    }
  }

  /**
   * Push the current branch to origin. Must be called before creating a PR.
   * Handles both fresh push (--set-upstream) and subsequent pushes.
   * Skips pre-push hooks to avoid blocking on lint/test (agent already validated).
   * @param {string} worktreePath
   * @param {string} branch
   * @returns {{ success: boolean, error?: string }}
   * @private
   */
  _pushBranch(worktreePath, branch) {
    try {
      // First merge upstream main to avoid conflicts
      try {
        spawnSync("git", ["fetch", "origin", "main", "--quiet"], {
          cwd: worktreePath, encoding: "utf8", timeout: 30_000,
        });
        const mergeResult = spawnSync(
          "git", ["merge", "origin/main", "--no-edit", "--strategy-option=theirs"],
          { cwd: worktreePath, encoding: "utf8", timeout: 30_000 }
        );
        if (mergeResult.status !== 0) {
          const mergeErr = (mergeResult.stderr || "").trim();
          // If merge fails with conflicts, try aborting and continuing without merge
          if (mergeErr.includes("CONFLICT") || mergeErr.includes("conflict")) {
            console.warn(`${TAG} merge conflict during upstream merge — aborting merge, will push as-is`);
            spawnSync("git", ["merge", "--abort"], {
              cwd: worktreePath, encoding: "utf8", timeout: 10_000,
            });
          }
        }
      } catch { /* best-effort upstream merge */ }

      // Push with --set-upstream, skip pre-push hooks
      const result = spawnSync(
        "git",
        ["push", "--set-upstream", "origin", branch, "--no-verify"],
        {
          cwd: worktreePath,
          encoding: "utf8",
          timeout: 120_000, // 2 min — push can be slow
          env: { ...process.env },
        }
      );

      if (result.status === 0) {
        console.log(`${TAG} pushed branch ${branch} to origin`);
        return { success: true };
      } else {
        const stderr = (result.stderr || "").trim();
        console.warn(`${TAG} push failed for ${branch}: ${stderr}`);
        return { success: false, error: stderr };
      }
    } catch (err) {
      console.warn(`${TAG} push error for ${branch}: ${err.message}`);
      return { success: false, error: err.message };
    }
  }

  /**
   * Enable GitHub auto-merge on a PR so it merges automatically when CI passes.
   * @param {string|number} prNumber
   * @param {string} worktreePath
   * @private
   */
  _enableAutoMerge(prNumber, worktreePath) {
    try {
      const result = spawnSync(
        "gh",
        ["pr", "merge", String(prNumber), "--auto", "--squash"],
        {
          cwd: worktreePath,
          encoding: "utf8",
          timeout: 15_000,
          env: { ...process.env },
        }
      );
      if (result.status === 0) {
        console.log(`${TAG} auto-merge enabled for PR #${prNumber}`);
      } else {
        const stderr = (result.stderr || "").trim();
        console.warn(`${TAG} auto-merge failed for PR #${prNumber}: ${stderr}`);
      }
    } catch (err) {
      console.warn(`${TAG} auto-merge error: ${err.message}`);
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

      if (!branch) {
        console.warn(`${TAG} cannot create PR — no branch name detected`);
        return null;
      }

      // ── Step 1: Push branch to origin ──────────────────────────────────
      const pushResult = this._pushBranch(worktreePath, branch);
      if (!pushResult.success) {
        console.warn(`${TAG} cannot create PR — push failed: ${pushResult.error}`);
        // Still try to create PR in case agent already pushed
      }

      // ── Step 2: Create the PR ──────────────────────────────────────────
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

      let prUrl = null;
      let prNumber = null;

      if (result.status === 0) {
        prUrl = (result.stdout || "").trim();
        console.log(`${TAG} PR created: ${prUrl}`);
        // Extract PR number from URL (e.g., https://github.com/owner/repo/pull/123)
        const prMatch = prUrl.match(/\/pull\/(\d+)/);
        prNumber = prMatch ? prMatch[1] : null;
      } else {
        const stderr = (result.stderr || "").trim();
        if (stderr.includes("already exists")) {
          console.log(`${TAG} PR already exists for ${branch}`);
          // Try to get the existing PR number
          try {
            const prList = spawnSync(
              "gh", ["pr", "list", "--head", branch, "--json", "number,url", "--limit", "1"],
              { cwd: worktreePath, encoding: "utf8", timeout: 10_000, env: { ...process.env } }
            );
            if (prList.status === 0) {
              const prs = JSON.parse(prList.stdout || "[]");
              if (prs.length > 0) {
                prUrl = prs[0].url;
                prNumber = String(prs[0].number);
              }
            }
          } catch { /* best-effort */ }
          prUrl = prUrl || "(existing)";
        } else {
          console.warn(`${TAG} PR creation failed: ${stderr}`);
          return null;
        }
      }

      // ── Step 3: Enable auto-merge ──────────────────────────────────────
      if (prNumber) {
        this._enableAutoMerge(prNumber, worktreePath);
      }

      return { url: prUrl, branch, prNumber };
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
