import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { mkdir, writeFile, readdir, rm, readFile } from "node:fs/promises";
import { existsSync } from "node:fs";
import { resolve } from "node:path";
import { tmpdir } from "node:os";
import { randomBytes } from "node:crypto";
import {
  isOldEnoughToArchive,
  archiveTaskToFile,
  isAlreadyArchived,
  fetchCompletedTasks,
  deleteTaskFromVK,
  archiveCompletedTasks,
  loadArchivedTasks,
  pruneOldArchives,
  getArchiveStats,
  generateSprintReport,
  formatSprintReport,
  ARCHIVE_AGE_HOURS,
  ARCHIVE_RETENTION_DAYS,
  DEFAULT_MAX_ARCHIVE,
} from "../task-archiver.mjs";

// ── Helpers ──────────────────────────────────────────────────────────────────

function tmpArchiveDir() {
  return resolve(tmpdir(), `archiver-test-${randomBytes(6).toString("hex")}`);
}

function makeTask(overrides = {}) {
  return {
    id: overrides.id ?? randomBytes(8).toString("hex"),
    title: overrides.title ?? "Test task",
    status: overrides.status ?? "done",
    priority: overrides.priority ?? 1,
    completed_at:
      overrides.completed_at ??
      new Date(Date.now() - 48 * 60 * 60 * 1000).toISOString(),
    updated_at:
      overrides.updated_at ??
      new Date(Date.now() - 48 * 60 * 60 * 1000).toISOString(),
    ...overrides,
  };
}

// ── Constants ────────────────────────────────────────────────────────────────

describe("exported constants", () => {
  it("ARCHIVE_AGE_HOURS is 24", () => {
    expect(ARCHIVE_AGE_HOURS).toBe(24);
  });

  it("ARCHIVE_RETENTION_DAYS is 90", () => {
    expect(ARCHIVE_RETENTION_DAYS).toBe(90);
  });

  it("DEFAULT_MAX_ARCHIVE is 50", () => {
    expect(DEFAULT_MAX_ARCHIVE).toBe(50);
  });
});

// ── isOldEnoughToArchive ─────────────────────────────────────────────────────

describe("isOldEnoughToArchive", () => {
  const nowMs = Date.now();

  it("returns true for task completed 48h ago with 24h threshold", () => {
    const task = makeTask({
      completed_at: new Date(nowMs - 48 * 60 * 60 * 1000).toISOString(),
    });
    expect(isOldEnoughToArchive(task, { nowMs })).toBe(true);
  });

  it("returns false for task completed 1h ago with 24h threshold", () => {
    const task = makeTask({
      completed_at: new Date(nowMs - 1 * 60 * 60 * 1000).toISOString(),
    });
    expect(isOldEnoughToArchive(task, { nowMs })).toBe(false);
  });

  it("uses updated_at as fallback when completed_at is missing", () => {
    const task = {
      id: "t1",
      title: "fallback",
      status: "done",
      updated_at: new Date(nowMs - 48 * 60 * 60 * 1000).toISOString(),
    };
    expect(isOldEnoughToArchive(task, { nowMs })).toBe(true);
  });

  it("returns false for invalid date", () => {
    const task = { id: "t2", completed_at: "invalid-date" };
    expect(isOldEnoughToArchive(task)).toBe(false);
  });

  it("respects custom ageHours", () => {
    const task = makeTask({
      completed_at: new Date(nowMs - 2 * 60 * 60 * 1000).toISOString(),
    });
    expect(isOldEnoughToArchive(task, { nowMs, ageHours: 1 })).toBe(true);
    expect(isOldEnoughToArchive(task, { nowMs, ageHours: 4 })).toBe(false);
  });

  it("uses ARCHIVE_AGE_HOURS by default", () => {
    const task = makeTask({
      completed_at: new Date(nowMs - 25 * 60 * 60 * 1000).toISOString(),
    });
    expect(isOldEnoughToArchive(task, { nowMs })).toBe(true);
  });
});

// ── archiveTaskToFile ────────────────────────────────────────────────────────

describe("archiveTaskToFile", () => {
  let dir;

  beforeEach(() => {
    dir = tmpArchiveDir();
  });

  afterEach(async () => {
    await rm(dir, { recursive: true, force: true }).catch(() => {});
  });

  it("creates archive file with correct structure", async () => {
    const task = makeTask({ id: "abc123" });
    const path = await archiveTaskToFile(task, null, dir);

    expect(path).not.toBeNull();
    expect(existsSync(path)).toBe(true);

    const content = JSON.parse(await readFile(path, "utf8"));
    expect(content.task.id).toBe("abc123");
    expect(content.archived_at).toBeDefined();
    expect(content.archiver_version).toBe(2);
  });

  it("is idempotent — returns same path on second call", async () => {
    const task = makeTask({ id: "idem123" });
    const path1 = await archiveTaskToFile(task, null, dir);
    const path2 = await archiveTaskToFile(task, null, dir);

    expect(path1).toBe(path2);
    // Verify only one file exists
    const files = await readdir(dir);
    const matching = files.filter((f) => f.includes("idem123"));
    expect(matching).toHaveLength(1);
  });

  it("includes attempt data when provided", async () => {
    const task = makeTask();
    const attempt = { id: "att-1", branch: "ve/test" };
    const path = await archiveTaskToFile(task, attempt, dir);

    const content = JSON.parse(await readFile(path, "utf8"));
    expect(content.attempt).toEqual(attempt);
  });

  it("creates directory if it doesn't exist", async () => {
    const nested = resolve(dir, "nested", "deep");
    const task = makeTask();
    const path = await archiveTaskToFile(task, null, nested);
    expect(path).not.toBeNull();
    expect(existsSync(nested)).toBe(true);
  });

  it("returns null on undefined task", async () => {
    const path = await archiveTaskToFile(undefined, null, dir);
    expect(path).toBeNull();
  });
});

// ── isAlreadyArchived ────────────────────────────────────────────────────────

describe("isAlreadyArchived", () => {
  let dir;

  beforeEach(() => {
    dir = tmpArchiveDir();
  });

  afterEach(async () => {
    await rm(dir, { recursive: true, force: true }).catch(() => {});
  });

  it("returns false for non-existent directory", async () => {
    expect(await isAlreadyArchived("test-id", dir)).toBe(false);
  });

  it("returns false when task not archived", async () => {
    await mkdir(dir, { recursive: true });
    await writeFile(resolve(dir, "2026-01-01-other.json"), "{}");
    expect(await isAlreadyArchived("test-id", dir)).toBe(false);
  });

  it("returns true when task file exists", async () => {
    await mkdir(dir, { recursive: true });
    await writeFile(resolve(dir, "2026-01-01-match123.json"), "{}");
    expect(await isAlreadyArchived("match123", dir)).toBe(true);
  });

  it("returns false for empty taskId", async () => {
    expect(await isAlreadyArchived("", dir)).toBe(false);
    expect(await isAlreadyArchived(null, dir)).toBe(false);
  });
});

// ── fetchCompletedTasks ──────────────────────────────────────────────────────

describe("fetchCompletedTasks", () => {
  it("returns empty array when fetchVk is null", async () => {
    expect(await fetchCompletedTasks(null, "proj-1")).toEqual([]);
  });

  it("returns empty array when projectId is null", async () => {
    expect(await fetchCompletedTasks(() => {}, null)).toEqual([]);
  });

  it("fetches done and cancelled tasks", async () => {
    const doneTasks = [makeTask({ status: "done" })];
    const cancelledTasks = [makeTask({ status: "cancelled" })];

    const fetchVk = async (path) => {
      if (path.includes("status=done"))
        return { success: true, data: doneTasks };
      if (path.includes("status=cancelled"))
        return { success: true, data: cancelledTasks };
      return { success: false };
    };

    const result = await fetchCompletedTasks(fetchVk, "proj-1");
    expect(result).toHaveLength(2);
  });

  it("handles API errors gracefully", async () => {
    const fetchVk = async () => {
      throw new Error("API down");
    };
    const result = await fetchCompletedTasks(fetchVk, "proj-1");
    expect(result).toEqual([]);
  });

  it("handles non-success responses", async () => {
    const fetchVk = async () => ({ success: false });
    const result = await fetchCompletedTasks(fetchVk, "proj-1");
    expect(result).toEqual([]);
  });
});

// ── deleteTaskFromVK ─────────────────────────────────────────────────────────

describe("deleteTaskFromVK", () => {
  it("returns false for null fetchVk", async () => {
    expect(await deleteTaskFromVK(null, "task-1")).toBe(false);
  });

  it("returns false for null taskId", async () => {
    expect(await deleteTaskFromVK(() => {}, null)).toBe(false);
  });

  it("returns true on successful DELETE", async () => {
    const fetchVk = async () => ({ success: true });
    expect(await deleteTaskFromVK(fetchVk, "task-1")).toBe(true);
  });

  it("falls back to PUT when DELETE fails", async () => {
    let callCount = 0;
    const fetchVk = async (path, opts) => {
      callCount++;
      if (opts?.method === "DELETE") return { success: false };
      if (opts?.method === "PUT") return { success: true };
      return { success: false };
    };
    expect(await deleteTaskFromVK(fetchVk, "task-1")).toBe(true);
    expect(callCount).toBe(2);
  });

  it("returns false when both DELETE and PUT fail", async () => {
    const fetchVk = async () => ({ success: false });
    expect(await deleteTaskFromVK(fetchVk, "task-1")).toBe(false);
  });

  it("returns false on exception", async () => {
    const fetchVk = async () => {
      throw new Error("network");
    };
    expect(await deleteTaskFromVK(fetchVk, "task-1")).toBe(false);
  });
});

// ── pruneOldArchives ─────────────────────────────────────────────────────────

describe("pruneOldArchives", () => {
  let dir;

  beforeEach(() => {
    dir = tmpArchiveDir();
  });

  afterEach(async () => {
    await rm(dir, { recursive: true, force: true }).catch(() => {});
  });

  it("returns 0 for non-existent directory", async () => {
    expect(await pruneOldArchives({ archiveDir: dir })).toBe(0);
  });

  it("does not prune recent files", async () => {
    await mkdir(dir, { recursive: true });
    await writeFile(resolve(dir, "recent.json"), "{}");
    const pruned = await pruneOldArchives({
      archiveDir: dir,
      retentionDays: 90,
    });
    expect(pruned).toBe(0);
    expect(existsSync(resolve(dir, "recent.json"))).toBe(true);
  });

  it("prunes expired files based on mtime", async () => {
    await mkdir(dir, { recursive: true });
    const filePath = resolve(dir, "old-task.json");
    await writeFile(filePath, "{}");
    // Use nowMs far in the future to make file appear old
    const farFuture = Date.now() + 200 * 24 * 60 * 60 * 1000;
    const pruned = await pruneOldArchives({
      archiveDir: dir,
      retentionDays: 90,
      nowMs: farFuture,
    });
    expect(pruned).toBe(1);
    expect(existsSync(filePath)).toBe(false);
  });

  it("skips non-json files", async () => {
    await mkdir(dir, { recursive: true });
    await writeFile(resolve(dir, "readme.txt"), "hello");
    const farFuture = Date.now() + 200 * 24 * 60 * 60 * 1000;
    const pruned = await pruneOldArchives({
      archiveDir: dir,
      retentionDays: 90,
      nowMs: farFuture,
    });
    expect(pruned).toBe(0);
    expect(existsSync(resolve(dir, "readme.txt"))).toBe(true);
  });
});

// ── getArchiveStats ──────────────────────────────────────────────────────────

describe("getArchiveStats", () => {
  let dir;

  beforeEach(() => {
    dir = tmpArchiveDir();
  });

  afterEach(async () => {
    await rm(dir, { recursive: true, force: true }).catch(() => {});
  });

  it("returns zeros for non-existent directory", async () => {
    const stats = await getArchiveStats(dir);
    expect(stats).toEqual({ count: 0, totalBytes: 0 });
  });

  it("counts json files and sums sizes", async () => {
    await mkdir(dir, { recursive: true });
    await writeFile(resolve(dir, "a.json"), '{"x":1}');
    await writeFile(resolve(dir, "b.json"), '{"y":2}');
    await writeFile(resolve(dir, "c.txt"), "not counted");

    const stats = await getArchiveStats(dir);
    expect(stats.count).toBe(2);
    expect(stats.totalBytes).toBeGreaterThan(0);
  });
});

// ── archiveCompletedTasks (integration) ──────────────────────────────────────

describe("archiveCompletedTasks", () => {
  let dir;

  beforeEach(() => {
    dir = tmpArchiveDir();
  });

  afterEach(async () => {
    await rm(dir, { recursive: true, force: true }).catch(() => {});
  });

  it("archives old tasks and deletes from VK", async () => {
    const oldTask = makeTask({
      id: "old-task-1",
      title: "Old task",
      completed_at: new Date(Date.now() - 48 * 60 * 60 * 1000).toISOString(),
    });

    const fetchVk = async (path, opts) => {
      if (path.includes("status=done"))
        return { success: true, data: [oldTask] };
      if (path.includes("status=cancelled")) return { success: true, data: [] };
      if (opts?.method === "DELETE") return { success: true };
      return { success: false };
    };

    const result = await archiveCompletedTasks(fetchVk, "proj-1", {
      archiveDir: dir,
      prune: false,
    });

    expect(result.archived).toBe(1);
    expect(result.deleted).toBe(1);
    expect(result.errors).toBe(0);
  });

  it("skips tasks that are too recent", async () => {
    const recentTask = makeTask({
      id: "recent-1",
      completed_at: new Date(Date.now() - 1 * 60 * 60 * 1000).toISOString(),
    });

    const fetchVk = async (path) => {
      if (path.includes("status=done"))
        return { success: true, data: [recentTask] };
      return { success: true, data: [] };
    };

    const result = await archiveCompletedTasks(fetchVk, "proj-1", {
      archiveDir: dir,
      prune: false,
    });
    expect(result.archived).toBe(0);
  });

  it("dry-run archives but doesn't delete", async () => {
    const task = makeTask({ id: "dryrun-1" });
    let deleteCalled = false;

    const fetchVk = async (path, opts) => {
      if (path.includes("status=done")) return { success: true, data: [task] };
      if (path.includes("status=cancelled")) return { success: true, data: [] };
      if (opts?.method === "DELETE") {
        deleteCalled = true;
        return { success: true };
      }
      return { success: false };
    };

    const result = await archiveCompletedTasks(fetchVk, "proj-1", {
      archiveDir: dir,
      dryRun: true,
      prune: false,
    });

    expect(result.archived).toBe(1);
    expect(result.deleted).toBe(0);
    expect(deleteCalled).toBe(false);
  });

  it("skips already-archived tasks (idempotent)", async () => {
    const task = makeTask({ id: "idem-task" });

    // Pre-archive the task
    await archiveTaskToFile(task, null, dir);

    const fetchVk = async (path) => {
      if (path.includes("status=done")) return { success: true, data: [task] };
      return { success: true, data: [] };
    };

    const result = await archiveCompletedTasks(fetchVk, "proj-1", {
      archiveDir: dir,
      prune: false,
    });

    expect(result.archived).toBe(0);
    expect(result.skipped).toBe(1);
  });

  it("respects maxArchive limit", async () => {
    const tasks = Array.from({ length: 10 }, (_, i) =>
      makeTask({ id: `limit-${i}` }),
    );

    const fetchVk = async (path, opts) => {
      if (path.includes("status=done")) return { success: true, data: tasks };
      if (path.includes("status=cancelled")) return { success: true, data: [] };
      if (opts?.method === "DELETE") return { success: true };
      return { success: false };
    };

    const result = await archiveCompletedTasks(fetchVk, "proj-1", {
      archiveDir: dir,
      maxArchive: 3,
      prune: false,
    });

    expect(result.archived).toBe(3);
  });

  it("handles fetchVk error gracefully (caught by fetchCompletedTasks)", async () => {
    const fetchVk = async () => {
      throw new Error("API down");
    };
    const result = await archiveCompletedTasks(fetchVk, "proj-1", {
      archiveDir: dir,
      prune: false,
    });
    // fetchCompletedTasks catches internally and returns [], so no errors bubble
    expect(result.archived).toBe(0);
    expect(result.deleted).toBe(0);
    expect(result.errors).toBe(0);
  });

  it("returns zero result for empty completed tasks", async () => {
    const fetchVk = async () => ({ success: true, data: [] });
    const result = await archiveCompletedTasks(fetchVk, "proj-1", {
      archiveDir: dir,
      prune: false,
    });
    expect(result.archived).toBe(0);
    expect(result.deleted).toBe(0);
    expect(result.errors).toBe(0);
  });
});

// ── loadArchivedTasks ────────────────────────────────────────────────────────

describe("loadArchivedTasks", () => {
  let dir;

  beforeEach(() => {
    dir = tmpArchiveDir();
  });

  afterEach(async () => {
    await rm(dir, { recursive: true, force: true }).catch(() => {});
  });

  it("returns empty array for non-existent directory", async () => {
    expect(await loadArchivedTasks({ archiveDir: dir })).toEqual([]);
  });

  it("loads and sorts archives newest first", async () => {
    await mkdir(dir, { recursive: true });

    const older = {
      task: makeTask({ id: "old" }),
      archived_at: "2026-01-01T00:00:00.000Z",
    };
    const newer = {
      task: makeTask({ id: "new" }),
      archived_at: "2026-02-01T00:00:00.000Z",
    };

    await writeFile(resolve(dir, "2026-01-01-old.json"), JSON.stringify(older));
    await writeFile(resolve(dir, "2026-02-01-new.json"), JSON.stringify(newer));

    const result = await loadArchivedTasks({ archiveDir: dir });
    expect(result).toHaveLength(2);
    expect(result[0].task.id).toBe("new");
    expect(result[1].task.id).toBe("old");
  });

  it("filters by since date", async () => {
    await mkdir(dir, { recursive: true });

    const old = {
      task: makeTask({ id: "old" }),
      archived_at: "2025-12-01T00:00:00.000Z",
    };
    const recent = {
      task: makeTask({ id: "recent" }),
      archived_at: "2026-02-01T00:00:00.000Z",
    };

    await writeFile(resolve(dir, "old.json"), JSON.stringify(old));
    await writeFile(resolve(dir, "recent.json"), JSON.stringify(recent));

    const result = await loadArchivedTasks({
      archiveDir: dir,
      since: "2026-01-01",
    });
    expect(result).toHaveLength(1);
    expect(result[0].task.id).toBe("recent");
  });

  it("filters by status", async () => {
    await mkdir(dir, { recursive: true });

    const doneTask = {
      task: makeTask({ id: "done-1", status: "done" }),
      archived_at: "2026-02-01T00:00:00.000Z",
    };
    const cancelledTask = {
      task: makeTask({ id: "cancelled-1", status: "cancelled" }),
      archived_at: "2026-02-01T00:00:00.000Z",
    };

    await writeFile(resolve(dir, "done.json"), JSON.stringify(doneTask));
    await writeFile(
      resolve(dir, "cancelled.json"),
      JSON.stringify(cancelledTask),
    );

    const result = await loadArchivedTasks({
      archiveDir: dir,
      status: "done",
    });
    expect(result).toHaveLength(1);
    expect(result[0].task.id).toBe("done-1");
  });

  it("skips corrupted JSON files", async () => {
    await mkdir(dir, { recursive: true });
    await writeFile(
      resolve(dir, "good.json"),
      JSON.stringify({
        task: makeTask({ id: "good" }),
        archived_at: "2026-02-01T00:00:00.000Z",
      }),
    );
    await writeFile(resolve(dir, "bad.json"), "not valid json{{{");

    const result = await loadArchivedTasks({ archiveDir: dir });
    expect(result).toHaveLength(1);
    expect(result[0].task.id).toBe("good");
  });
});

// ── generateSprintReport ─────────────────────────────────────────────────────

describe("generateSprintReport", () => {
  it("returns empty report for non-array input", () => {
    const report = generateSprintReport(null);
    expect(report.total).toBe(0);
    expect(report.tasks).toEqual([]);
  });

  it("returns empty report for empty array", () => {
    const report = generateSprintReport([]);
    expect(report.total).toBe(0);
  });

  it("groups by status, priority, and date", () => {
    const items = [
      {
        task: makeTask({
          id: "t1",
          status: "done",
          priority: 1,
          completed_at: "2026-02-01T10:00:00Z",
        }),
        archived_at: "2026-02-02T00:00:00Z",
      },
      {
        task: makeTask({
          id: "t2",
          status: "cancelled",
          priority: 2,
          completed_at: "2026-02-01T12:00:00Z",
        }),
        archived_at: "2026-02-02T01:00:00Z",
      },
      {
        task: makeTask({
          id: "t3",
          status: "done",
          priority: 1,
          completed_at: "2026-02-02T10:00:00Z",
        }),
        archived_at: "2026-02-03T00:00:00Z",
      },
    ];

    const report = generateSprintReport(items);
    expect(report.total).toBe(3);
    expect(report.by_status.done).toBe(2);
    expect(report.by_status.cancelled).toBe(1);
    expect(report.by_priority[1]).toBe(2);
    expect(report.by_priority[2]).toBe(1);
    expect(report.by_date["2026-02-01"]).toBe(2);
    expect(report.by_date["2026-02-02"]).toBe(1);
    expect(report.tasks).toHaveLength(3);
  });

  it("skips items without task", () => {
    const items = [
      { archived_at: "2026-01-01" },
      { task: makeTask({ id: "ok" }), archived_at: "2026-01-01" },
    ];
    const report = generateSprintReport(items);
    expect(report.tasks).toHaveLength(1);
  });

  it("handles task with invalid date", () => {
    const items = [
      {
        task: makeTask({
          id: "invalid",
          completed_at: "not-a-date",
          updated_at: undefined,
        }),
        archived_at: "2026-01-01",
      },
    ];
    const report = generateSprintReport(items);
    expect(report.by_date["unknown"]).toBe(1);
  });
});

// ── formatSprintReport ───────────────────────────────────────────────────────

describe("formatSprintReport", () => {
  it("returns 'No report data.' for null", () => {
    expect(formatSprintReport(null)).toBe("No report data.");
    expect(formatSprintReport(undefined)).toBe("No report data.");
  });

  it("formats a complete report", () => {
    const report = {
      total: 2,
      by_status: { done: 2 },
      by_priority: { 1: 2 },
      by_date: { "2026-02-01": 2 },
      tasks: [
        { id: "abcdefgh", title: "Task Alpha", status: "done" },
        { id: "12345678", title: "Task Beta", status: "done" },
      ],
    };

    const text = formatSprintReport(report);
    expect(text).toContain("Sprint Review Report");
    expect(text).toContain("Total Tasks Completed: 2");
    expect(text).toContain("done: 2");
    expect(text).toContain("2026-02-01: 2 tasks");
    expect(text).toContain("Task Alpha");
    expect(text).toContain("abcdefgh");
  });

  it("handles empty sections gracefully", () => {
    const report = {
      total: 0,
      by_status: {},
      by_priority: {},
      by_date: {},
      tasks: [],
    };
    const text = formatSprintReport(report);
    expect(text).toContain("Total Tasks Completed: 0");
    expect(text).not.toContain("By Status:");
  });

  it("handles missing task fields", () => {
    const report = {
      total: 1,
      by_status: {},
      by_priority: {},
      by_date: {},
      tasks: [{ id: undefined, title: undefined, status: undefined }],
    };
    const text = formatSprintReport(report);
    expect(text).toContain("untitled");
    expect(text).toContain("[?]");
  });
});
