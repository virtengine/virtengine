import { describe, it, expect, beforeEach, vi } from "vitest";

const mockTaskStore = vi.hoisted(() => ({
  getTask: vi.fn(),
  getAllTasks: vi.fn(() => []),
  addTask: vi.fn(),
  updateTask: vi.fn(),
  getDirtyTasks: vi.fn(() => []),
  markSynced: vi.fn(),
  upsertFromExternal: vi.fn(),
  setTaskStatus: vi.fn(),
  removeTask: vi.fn(),
}));

const mockKanban = vi.hoisted(() => ({
  getKanbanAdapter: vi.fn(),
  listTasks: vi.fn(() => Promise.resolve([])),
  updateTaskStatus: vi.fn(() => Promise.resolve({})),
}));

vi.mock("../task-store.mjs", () => mockTaskStore);
vi.mock("../kanban-adapter.mjs", () => mockKanban);

const { SyncEngine } = await import("../sync-engine.mjs");

describe("sync-engine backward external status handling", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockTaskStore.getAllTasks.mockReturnValue([]);
    mockKanban.listTasks.mockResolvedValue([]);
  });

  it("re-pushes backward move when internal equals old external status", async () => {
    // When internal is at the same rank as oldExternal, the code treats the
    // backward move as VK state loss and marks dirty for re-push.
    mockTaskStore.getAllTasks.mockReturnValue([
      {
        id: "task-1",
        projectId: "proj-1",
        status: "inprogress",
        externalStatus: "inprogress",
        syncDirty: false,
      },
    ]);
    mockKanban.listTasks.mockResolvedValue([
      { id: "task-1", status: "todo", projectId: "proj-1" },
    ]);

    const engine = new SyncEngine({ projectId: "proj-1" });
    const result = await engine.pullFromExternal();

    expect(mockTaskStore.setTaskStatus).not.toHaveBeenCalled();
    expect(mockTaskStore.updateTask).toHaveBeenCalledWith("task-1", {
      externalStatus: "todo",
      syncDirty: true,
    });
    expect(result.pulled).toBe(0);
  });

  it("re-pushes external backward move when internal task is dirty", async () => {
    mockTaskStore.getAllTasks.mockReturnValue([
      {
        id: "task-2",
        projectId: "proj-1",
        status: "inprogress",
        externalStatus: "inprogress",
        syncDirty: true,
      },
    ]);
    mockKanban.listTasks.mockResolvedValue([
      { id: "task-2", status: "todo", projectId: "proj-1" },
    ]);

    const engine = new SyncEngine({ projectId: "proj-1" });
    const result = await engine.pullFromExternal();

    expect(mockTaskStore.setTaskStatus).not.toHaveBeenCalled();
    expect(mockTaskStore.updateTask).toHaveBeenCalledWith("task-2", {
      externalStatus: "todo",
      syncDirty: true,
    });
    expect(result.pulled).toBe(0);
  });
});

describe("sync-engine push 404 orphan handling", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockTaskStore.getAllTasks.mockReturnValue([]);
    mockKanban.listTasks.mockResolvedValue([]);
  });

  it("removes orphaned task from internal store on 404 push error", async () => {
    mockTaskStore.getDirtyTasks.mockReturnValue([
      { id: "orphan-1", status: "inprogress", syncDirty: true },
    ]);
    mockKanban.updateTaskStatus.mockRejectedValue(
      new Error("VK API PUT /api/tasks/orphan-1 failed: 404"),
    );

    const engine = new SyncEngine({ projectId: "proj-1" });
    const result = await engine.pushToExternal();

    expect(mockTaskStore.removeTask).toHaveBeenCalledWith("orphan-1");
    expect(mockTaskStore.markSynced).not.toHaveBeenCalled();
    expect(result.pushed).toBe(0);
    // 404 is handled gracefully, should not appear in errors
    expect(result.errors.length).toBe(0);
  });

  it("does not remove task on non-404 push errors", async () => {
    mockTaskStore.getDirtyTasks.mockReturnValue([
      { id: "task-x", status: "inprogress", syncDirty: true },
    ]);
    mockKanban.updateTaskStatus.mockRejectedValue(
      new Error("VK API PUT /api/tasks/task-x failed: 500"),
    );

    const engine = new SyncEngine({ projectId: "proj-1" });
    const result = await engine.pushToExternal();

    expect(mockTaskStore.removeTask).not.toHaveBeenCalled();
    expect(result.errors.length).toBe(1);
    expect(result.errors[0]).toContain("500");
  });
});
