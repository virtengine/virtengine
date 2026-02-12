import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

// ── Mocks ───────────────────────────────────────────────────────────────────

vi.mock("../kanban-adapter.mjs", () => ({
  getKanbanAdapter: vi.fn(),
  listTasks: vi.fn(() => []),
  listProjects: vi.fn(() => [{ id: "proj-1", name: "Test Project" }]),
  getTask: vi.fn(),
  updateTaskStatus: vi.fn(() => Promise.resolve()),
}));

vi.mock("../agent-pool.mjs", () => ({
  launchOrResumeThread: vi.fn(),
  execWithRetry: vi.fn(() =>
    Promise.resolve({ success: true, output: "done", attempts: 1 }),
  ),
  invalidateThread: vi.fn(),
  getActiveThreads: vi.fn(() => []),
  getPoolSdkName: vi.fn(() => "codex"),
  pruneAllExhaustedThreads: vi.fn(() => 0),
}));

vi.mock("../worktree-manager.mjs", () => ({
  acquireWorktree: vi.fn(() =>
    Promise.resolve({ path: "/fake/worktree", created: true }),
  ),
  releaseWorktree: vi.fn(() => Promise.resolve()),
  getWorktreeStats: vi.fn(() => ({ active: 0, total: 0 })),
}));

vi.mock("../config.mjs", () => ({
  loadConfig: vi.fn(() => ({})),
}));

vi.mock("node:child_process", () => ({
  execSync: vi.fn(() => ""),
  spawnSync: vi.fn(() => ({ status: 0, stdout: "", stderr: "" })),
}));

vi.mock("node:fs", () => ({
  readFileSync: vi.fn(() => ""),
  existsSync: vi.fn(() => false),
}));

// ── Imports (after mocks) ───────────────────────────────────────────────────

import {
  TaskExecutor,
  getTaskExecutor,
  loadExecutorOptionsFromConfig,
  isInternalExecutorEnabled,
  getExecutorMode,
} from "../task-executor.mjs";
import {
  listTasks,
  listProjects,
  updateTaskStatus,
} from "../kanban-adapter.mjs";
import { execWithRetry, getPoolSdkName } from "../agent-pool.mjs";
import { acquireWorktree, releaseWorktree } from "../worktree-manager.mjs";
import { loadConfig } from "../config.mjs";
import { existsSync } from "node:fs";

// ── Helpers ─────────────────────────────────────────────────────────────────

const mockTask = {
  id: "task-123-uuid",
  title: "Fix the bug",
  description: "There is a bug that needs fixing",
  status: "todo",
  branchName: "ve/task-123-fix-the-bug",
};

/** Saved env vars to restore after each test. */
const ENV_KEYS = [
  "EXECUTOR_MODE",
  "INTERNAL_EXECUTOR_PARALLEL",
  "INTERNAL_EXECUTOR_POLL_MS",
  "INTERNAL_EXECUTOR_SDK",
  "INTERNAL_EXECUTOR_TIMEOUT_MS",
  "INTERNAL_EXECUTOR_MAX_RETRIES",
  "INTERNAL_EXECUTOR_PROJECT_ID",
];

// ── Tests ───────────────────────────────────────────────────────────────────

describe("task-executor", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    for (const key of ENV_KEYS) delete process.env[key];
  });

  afterEach(() => {
    vi.restoreAllMocks();
    for (const key of ENV_KEYS) delete process.env[key];
  });

  // ────────────────────────────────────────────────────────────────────────
  // Constructor
  // ────────────────────────────────────────────────────────────────────────

  describe("TaskExecutor constructor", () => {
    it("sets default options when none provided", () => {
      const ex = new TaskExecutor();
      expect(ex.mode).toBe("vk");
      expect(ex.maxParallel).toBe(3);
      expect(ex.pollIntervalMs).toBe(30_000);
      expect(ex.sdk).toBe("auto");
      expect(ex.taskTimeoutMs).toBe(90 * 60 * 1000);
      expect(ex.maxRetries).toBe(2);
      expect(ex.autoCreatePr).toBe(true);
      expect(ex.projectId).toBeNull();
    });

    it("overrides defaults with provided options", () => {
      const ex = new TaskExecutor({
        mode: "internal",
        maxParallel: 5,
        sdk: "copilot",
        projectId: "proj-42",
      });
      expect(ex.mode).toBe("internal");
      expect(ex.maxParallel).toBe(5);
      expect(ex.sdk).toBe("copilot");
      expect(ex.projectId).toBe("proj-42");
      // untouched defaults
      expect(ex.pollIntervalMs).toBe(30_000);
    });

    it("initializes empty _activeSlots Map", () => {
      const ex = new TaskExecutor();
      expect(ex._activeSlots).toBeInstanceOf(Map);
      expect(ex._activeSlots.size).toBe(0);
    });

    it("sets _running to false initially", () => {
      const ex = new TaskExecutor();
      expect(ex._running).toBe(false);
    });
  });

  // ────────────────────────────────────────────────────────────────────────
  // getStatus
  // ────────────────────────────────────────────────────────────────────────

  describe("getStatus", () => {
    it("returns running state", () => {
      const ex = new TaskExecutor();
      expect(ex.getStatus().running).toBe(false);
      ex._running = true;
      expect(ex.getStatus().running).toBe(true);
    });

    it("returns correct mode, maxParallel, sdk", () => {
      const ex = new TaskExecutor({ mode: "hybrid", maxParallel: 7 });
      const status = ex.getStatus();
      expect(status.mode).toBe("hybrid");
      expect(status.maxParallel).toBe(7);
      // sdk "auto" delegates to getPoolSdkName()
      expect(status.sdk).toBe("codex");
      expect(getPoolSdkName).toHaveBeenCalled();
    });

    it("returns empty slots when none active", () => {
      const ex = new TaskExecutor();
      const status = ex.getStatus();
      expect(status.activeSlots).toBe(0);
      expect(status.slots).toEqual([]);
    });

    it("returns correct slot info when tasks are running", () => {
      const ex = new TaskExecutor();
      ex._activeSlots.set("task-abc", {
        taskId: "task-abc",
        taskTitle: "Some task",
        branch: "ve/task-abc-some-task",
        sdk: "codex",
        attempt: 1,
        startedAt: Date.now() - 5000,
        status: "running",
      });

      const status = ex.getStatus();
      expect(status.activeSlots).toBe(1);
      expect(status.slots).toHaveLength(1);
      expect(status.slots[0].taskId).toBe("task-abc");
      expect(status.slots[0].taskTitle).toBe("Some task");
      expect(status.slots[0].status).toBe("running");
      expect(status.slots[0].runningFor).toBeGreaterThanOrEqual(4);
    });
  });

  // ────────────────────────────────────────────────────────────────────────
  // start / stop
  // ────────────────────────────────────────────────────────────────────────

  describe("start / stop", () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("start() sets _running to true and creates poll timer", () => {
      const ex = new TaskExecutor({ pollIntervalMs: 10_000 });
      ex.start();
      expect(ex._running).toBe(true);
      expect(ex._pollTimer).not.toBeNull();
      // cleanup
      ex._running = false;
      clearInterval(ex._pollTimer);
    });

    it("stop() sets _running to false and clears poll timer", async () => {
      const ex = new TaskExecutor({ pollIntervalMs: 60_000 });
      ex.start();
      expect(ex._running).toBe(true);

      const stopPromise = ex.stop();
      // No active slots, should resolve quickly
      await stopPromise;

      expect(ex._running).toBe(false);
      expect(ex._pollTimer).toBeNull();
    });

    it("stop() waits for active slots gracefully", async () => {
      const ex = new TaskExecutor({ pollIntervalMs: 60_000 });
      ex.start();

      // Simulate an active slot
      ex._activeSlots.set("slot-1", {
        taskId: "slot-1",
        taskTitle: "test",
        startedAt: Date.now(),
        status: "running",
      });

      const stopPromise = ex.stop();

      // Advance timers to trigger the 1-second check intervals
      // Then remove the active slot to let stop() finish
      await vi.advanceTimersByTimeAsync(2000);
      ex._activeSlots.delete("slot-1");
      await vi.advanceTimersByTimeAsync(2000);

      await stopPromise;
      expect(ex._running).toBe(false);
    });
  });

  // ────────────────────────────────────────────────────────────────────────
  // _pollLoop guards
  // ────────────────────────────────────────────────────────────────────────

  describe("_pollLoop guards", () => {
    it("skips when _running is false", async () => {
      const ex = new TaskExecutor();
      ex._running = false;
      await ex._pollLoop();
      expect(listProjects).not.toHaveBeenCalled();
      expect(listTasks).not.toHaveBeenCalled();
    });

    it("skips when _pollInProgress is true", async () => {
      const ex = new TaskExecutor();
      ex._running = true;
      ex._pollInProgress = true;
      await ex._pollLoop();
      expect(listProjects).not.toHaveBeenCalled();
    });

    it("skips when slots are full", async () => {
      const ex = new TaskExecutor({ maxParallel: 1 });
      ex._running = true;
      ex._activeSlots.set("existing", { taskId: "existing" });
      await ex._pollLoop();
      expect(listProjects).not.toHaveBeenCalled();
    });

    it("resolves project ID on first poll from listProjects", async () => {
      const ex = new TaskExecutor();
      ex._running = true;
      listProjects.mockResolvedValueOnce([
        { id: "auto-proj", name: "Auto Detected" },
      ]);
      listTasks.mockResolvedValueOnce([]);

      await ex._pollLoop();

      expect(listProjects).toHaveBeenCalled();
      expect(ex._resolvedProjectId).toBe("auto-proj");
    });

    it("uses provided projectId without calling listProjects", async () => {
      const ex = new TaskExecutor({ projectId: "my-proj" });
      ex._running = true;
      listTasks.mockResolvedValueOnce([]);

      await ex._pollLoop();

      expect(listProjects).not.toHaveBeenCalled();
      expect(ex._resolvedProjectId).toBe("my-proj");
    });

    it("resets _pollInProgress even on error", async () => {
      const ex = new TaskExecutor({ projectId: "p1" });
      ex._running = true;
      listTasks.mockRejectedValueOnce(new Error("network"));

      await ex._pollLoop();
      expect(ex._pollInProgress).toBe(false);
    });
  });

  // ────────────────────────────────────────────────────────────────────────
  // executeTask
  // ────────────────────────────────────────────────────────────────────────

  describe("executeTask", () => {
    beforeEach(() => {
      existsSync.mockReturnValue(true);
    });

    it("allocates a slot for the task", async () => {
      const ex = new TaskExecutor();
      const promise = ex.executeTask({ ...mockTask });
      // Slot should be set immediately (synchronous part)
      expect(ex._activeSlots.has("task-123-uuid")).toBe(true);
      await promise;
    });

    it("calls updateTaskStatus with inprogress", async () => {
      const ex = new TaskExecutor();
      await ex.executeTask({ ...mockTask });
      expect(updateTaskStatus).toHaveBeenCalledWith(
        "task-123-uuid",
        "inprogress",
      );
    });

    it("calls acquireWorktree with correct branch", async () => {
      const ex = new TaskExecutor();
      await ex.executeTask({ ...mockTask });
      expect(acquireWorktree).toHaveBeenCalledWith(
        "ve/task-123-fix-the-bug",
        "task-123-uuid",
        expect.objectContaining({ owner: "task-executor" }),
      );
    });

    it("calls execWithRetry with built prompt", async () => {
      const ex = new TaskExecutor();
      await ex.executeTask({ ...mockTask });
      expect(execWithRetry).toHaveBeenCalledWith(
        expect.stringContaining("Fix the bug"),
        expect.objectContaining({
          taskKey: "task-123-uuid",
          cwd: "/fake/worktree",
        }),
      );
    });

    it("releases slot and worktree after completion", async () => {
      const ex = new TaskExecutor();
      await ex.executeTask({ ...mockTask });
      expect(releaseWorktree).toHaveBeenCalledWith("task-123-uuid");
      expect(ex._activeSlots.has("task-123-uuid")).toBe(false);
    });

    it("handles failure — calls onTaskFailed callback", async () => {
      execWithRetry.mockResolvedValueOnce({
        success: false,
        attempts: 2,
        error: "tests failed",
      });
      const onTaskFailed = vi.fn();
      const ex = new TaskExecutor({ onTaskFailed });

      await ex.executeTask({ ...mockTask });

      expect(onTaskFailed).toHaveBeenCalledWith(
        expect.objectContaining({ id: "task-123-uuid" }),
        expect.objectContaining({ success: false }),
      );
    });

    it("calls onTaskFailed when worktree acquisition fails", async () => {
      acquireWorktree.mockRejectedValueOnce(new Error("no space"));
      const onTaskFailed = vi.fn();
      const ex = new TaskExecutor({ onTaskFailed });

      await ex.executeTask({ ...mockTask });

      expect(onTaskFailed).toHaveBeenCalledWith(
        expect.objectContaining({ id: "task-123-uuid" }),
        expect.any(Error),
      );
      // Slot should be cleaned up
      expect(ex._activeSlots.has("task-123-uuid")).toBe(false);
    });

    it("generates branch name from task id and title when branchName missing", async () => {
      const task = {
        id: "abcd1234-uuid",
        title: "Add New Feature!",
        description: "desc",
        status: "todo",
      };
      const ex = new TaskExecutor();
      await ex.executeTask(task);

      // Branch should be auto-generated: ve/<first-8-of-id>-<slug>
      expect(acquireWorktree).toHaveBeenCalledWith(
        expect.stringMatching(/^ve\/abcd1234-add-new-feature/),
        "abcd1234-uuid",
        expect.any(Object),
      );
    });
  });

  // ────────────────────────────────────────────────────────────────────────
  // loadExecutorOptionsFromConfig
  // ────────────────────────────────────────────────────────────────────────

  describe("loadExecutorOptionsFromConfig", () => {
    it("returns defaults when nothing configured", () => {
      loadConfig.mockReturnValue({});
      const opts = loadExecutorOptionsFromConfig();
      expect(opts.mode).toBe("vk");
      expect(opts.maxParallel).toBe(3);
      expect(opts.sdk).toBe("auto");
      expect(opts.maxRetries).toBe(2);
    });

    it("reads from env vars", () => {
      process.env.EXECUTOR_MODE = "internal";
      process.env.INTERNAL_EXECUTOR_PARALLEL = "8";
      process.env.INTERNAL_EXECUTOR_SDK = "copilot";
      loadConfig.mockReturnValue({});

      const opts = loadExecutorOptionsFromConfig();
      expect(opts.mode).toBe("internal");
      expect(opts.maxParallel).toBe(8);
      expect(opts.sdk).toBe("copilot");
    });

    it("reads from config.internalExecutor", () => {
      loadConfig.mockReturnValue({
        internalExecutor: {
          mode: "hybrid",
          maxParallel: 4,
          sdk: "claude",
          maxRetries: 5,
        },
      });

      const opts = loadExecutorOptionsFromConfig();
      expect(opts.mode).toBe("hybrid");
      expect(opts.maxParallel).toBe(4);
      expect(opts.sdk).toBe("claude");
      expect(opts.maxRetries).toBe(5);
    });

    it("env vars take priority over config", () => {
      process.env.EXECUTOR_MODE = "internal";
      process.env.INTERNAL_EXECUTOR_PARALLEL = "10";
      loadConfig.mockReturnValue({
        internalExecutor: {
          mode: "hybrid",
          maxParallel: 2,
        },
      });

      const opts = loadExecutorOptionsFromConfig();
      expect(opts.mode).toBe("internal");
      expect(opts.maxParallel).toBe(10);
    });

    it("validates mode values — uses env when set", () => {
      process.env.EXECUTOR_MODE = "vk";
      loadConfig.mockReturnValue({
        internalExecutor: { mode: "internal" },
      });

      const opts = loadExecutorOptionsFromConfig();
      expect(opts.mode).toBe("vk");
    });

    it("reads from config.taskExecutor as fallback key", () => {
      loadConfig.mockReturnValue({
        taskExecutor: { mode: "internal", maxParallel: 6 },
      });

      const opts = loadExecutorOptionsFromConfig();
      expect(opts.mode).toBe("internal");
      expect(opts.maxParallel).toBe(6);
    });

    it("handles loadConfig throwing", () => {
      loadConfig.mockImplementation(() => {
        throw new Error("config missing");
      });

      const opts = loadExecutorOptionsFromConfig();
      expect(opts.mode).toBe("vk");
      expect(opts.maxParallel).toBe(3);
    });
  });

  // ────────────────────────────────────────────────────────────────────────
  // isInternalExecutorEnabled
  // ────────────────────────────────────────────────────────────────────────

  describe("isInternalExecutorEnabled", () => {
    it("returns true for EXECUTOR_MODE=internal", () => {
      process.env.EXECUTOR_MODE = "internal";
      expect(isInternalExecutorEnabled()).toBe(true);
    });

    it("returns true for EXECUTOR_MODE=hybrid", () => {
      process.env.EXECUTOR_MODE = "hybrid";
      expect(isInternalExecutorEnabled()).toBe(true);
    });

    it("returns false for EXECUTOR_MODE=vk", () => {
      process.env.EXECUTOR_MODE = "vk";
      expect(isInternalExecutorEnabled()).toBe(false);
    });

    it("falls back to config when env var not set", () => {
      loadConfig.mockReturnValue({
        internalExecutor: { mode: "internal" },
      });
      expect(isInternalExecutorEnabled()).toBe(true);
    });

    it("falls back to config.taskExecutor", () => {
      loadConfig.mockReturnValue({
        taskExecutor: { mode: "hybrid" },
      });
      expect(isInternalExecutorEnabled()).toBe(true);
    });

    it("returns false when nothing configured", () => {
      loadConfig.mockReturnValue({});
      expect(isInternalExecutorEnabled()).toBe(false);
    });

    it("returns false when loadConfig throws", () => {
      loadConfig.mockImplementation(() => {
        throw new Error("oops");
      });
      expect(isInternalExecutorEnabled()).toBe(false);
    });
  });

  // ────────────────────────────────────────────────────────────────────────
  // getExecutorMode
  // ────────────────────────────────────────────────────────────────────────

  describe("getExecutorMode", () => {
    it("returns env EXECUTOR_MODE when valid", () => {
      process.env.EXECUTOR_MODE = "internal";
      expect(getExecutorMode()).toBe("internal");
    });

    it("returns hybrid from env", () => {
      process.env.EXECUTOR_MODE = "hybrid";
      expect(getExecutorMode()).toBe("hybrid");
    });

    it("falls through to config when env invalid", () => {
      process.env.EXECUTOR_MODE = "bogus";
      loadConfig.mockReturnValue({
        internalExecutor: { mode: "internal" },
      });
      expect(getExecutorMode()).toBe("internal");
    });

    it("returns 'vk' as default", () => {
      loadConfig.mockReturnValue({});
      expect(getExecutorMode()).toBe("vk");
    });

    it("returns 'vk' when loadConfig throws", () => {
      loadConfig.mockImplementation(() => {
        throw new Error("fail");
      });
      expect(getExecutorMode()).toBe("vk");
    });
  });

  // ────────────────────────────────────────────────────────────────────────
  // getTaskExecutor singleton
  // ────────────────────────────────────────────────────────────────────────

  describe("getTaskExecutor singleton", () => {
    // The module-level _instance can't be reset without vi.resetModules().
    // We test basic behavior and then re-import for isolation.

    it("returns a TaskExecutor instance", async () => {
      // Use dynamic import with resetModules to get a fresh module
      vi.resetModules();

      // Re-apply mocks for the fresh module
      vi.doMock("../kanban-adapter.mjs", () => ({
        getKanbanAdapter: vi.fn(),
        listTasks: vi.fn(() => []),
        listProjects: vi.fn(() => []),
        getTask: vi.fn(),
        updateTaskStatus: vi.fn(),
      }));
      vi.doMock("../agent-pool.mjs", () => ({
        launchOrResumeThread: vi.fn(),
        execWithRetry: vi.fn(() => Promise.resolve({ success: true })),
        invalidateThread: vi.fn(),
        getActiveThreads: vi.fn(() => []),
        getPoolSdkName: vi.fn(() => "codex"),
      }));
      vi.doMock("../worktree-manager.mjs", () => ({
        acquireWorktree: vi.fn(),
        releaseWorktree: vi.fn(),
        getWorktreeStats: vi.fn(() => ({ active: 0, total: 0 })),
      }));
      vi.doMock("../config.mjs", () => ({
        loadConfig: vi.fn(() => ({})),
      }));
      vi.doMock("node:child_process", () => ({
        execSync: vi.fn(() => ""),
        spawnSync: vi.fn(() => ({ status: 0, stdout: "", stderr: "" })),
      }));
      vi.doMock("node:fs", () => ({
        readFileSync: vi.fn(() => ""),
        existsSync: vi.fn(() => false),
      }));

      const mod = await import("../task-executor.mjs");
      const inst = mod.getTaskExecutor({ mode: "vk" });
      expect(inst).toBeInstanceOf(mod.TaskExecutor);
    });

    it("returns same instance on second call", async () => {
      vi.resetModules();

      vi.doMock("../kanban-adapter.mjs", () => ({
        getKanbanAdapter: vi.fn(),
        listTasks: vi.fn(() => []),
        listProjects: vi.fn(() => []),
        getTask: vi.fn(),
        updateTaskStatus: vi.fn(),
      }));
      vi.doMock("../agent-pool.mjs", () => ({
        launchOrResumeThread: vi.fn(),
        execWithRetry: vi.fn(() => Promise.resolve({ success: true })),
        invalidateThread: vi.fn(),
        getActiveThreads: vi.fn(() => []),
        getPoolSdkName: vi.fn(() => "codex"),
      }));
      vi.doMock("../worktree-manager.mjs", () => ({
        acquireWorktree: vi.fn(),
        releaseWorktree: vi.fn(),
        getWorktreeStats: vi.fn(() => ({ active: 0, total: 0 })),
      }));
      vi.doMock("../config.mjs", () => ({
        loadConfig: vi.fn(() => ({})),
      }));
      vi.doMock("node:child_process", () => ({
        execSync: vi.fn(() => ""),
        spawnSync: vi.fn(() => ({ status: 0, stdout: "", stderr: "" })),
      }));
      vi.doMock("node:fs", () => ({
        readFileSync: vi.fn(() => ""),
        existsSync: vi.fn(() => false),
      }));

      const mod = await import("../task-executor.mjs");
      const first = mod.getTaskExecutor({ mode: "internal" });
      const second = mod.getTaskExecutor({ mode: "hybrid" });
      expect(first).toBe(second);
      // Mode should be from the first call
      expect(first.mode).toBe("internal");
    });
  });

  // ────────────────────────────────────────────────────────────────────────
  // _pollLoop task dispatching
  // ────────────────────────────────────────────────────────────────────────

  describe("_pollLoop task dispatching", () => {
    it("dispatches eligible tasks up to maxParallel", async () => {
      const ex = new TaskExecutor({ projectId: "proj-1", maxParallel: 2 });
      ex._running = true;
      existsSync.mockReturnValue(true);

      listTasks.mockResolvedValueOnce([
        { id: "t1", title: "Task 1", status: "todo", branchName: "ve/t1" },
        { id: "t2", title: "Task 2", status: "todo", branchName: "ve/t2" },
        { id: "t3", title: "Task 3", status: "todo", branchName: "ve/t3" },
      ]);

      await ex._pollLoop();

      // Should dispatch at most maxParallel (2) tasks
      // Wait a tick for the fire-and-forget executeTask promises
      await new Promise((r) => setTimeout(r, 50));
      expect(updateTaskStatus).toHaveBeenCalledWith("t1", "inprogress");
      expect(updateTaskStatus).toHaveBeenCalledWith("t2", "inprogress");
    });

    it("skips tasks already in active slots", async () => {
      const ex = new TaskExecutor({ projectId: "proj-1", maxParallel: 5 });
      ex._running = true;
      ex._activeSlots.set("t1", { taskId: "t1" });

      listTasks.mockResolvedValueOnce([
        { id: "t1", title: "Task 1", status: "todo", branchName: "ve/t1" },
        { id: "t2", title: "Task 2", status: "todo", branchName: "ve/t2" },
      ]);

      existsSync.mockReturnValue(true);
      await ex._pollLoop();
      await new Promise((r) => setTimeout(r, 50));

      // t1 was already in slots, should not be dispatched again
      expect(updateTaskStatus).not.toHaveBeenCalledWith("t1", "inprogress");
      expect(updateTaskStatus).toHaveBeenCalledWith("t2", "inprogress");
    });

    it("skips tasks in cooldown", async () => {
      const ex = new TaskExecutor({ projectId: "proj-1", maxParallel: 5 });
      ex._running = true;
      // Set cooldown just now → within COOLDOWN_MS window
      ex._taskCooldowns.set("t1", Date.now());

      listTasks.mockResolvedValueOnce([
        { id: "t1", title: "Task 1", status: "todo", branchName: "ve/t1" },
      ]);

      await ex._pollLoop();
      await new Promise((r) => setTimeout(r, 50));

      expect(updateTaskStatus).not.toHaveBeenCalled();
    });

    it("handles no projects found gracefully", async () => {
      const ex = new TaskExecutor();
      ex._running = true;
      listProjects.mockResolvedValueOnce([]);

      await ex._pollLoop();

      expect(listTasks).not.toHaveBeenCalled();
    });
  });
});
