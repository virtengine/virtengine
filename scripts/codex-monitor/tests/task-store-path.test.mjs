import { afterEach, describe, expect, it, vi } from "vitest";
import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { resolve } from "node:path";

const tempDirs = [];

function makeTempDir(prefix) {
  const dir = mkdtempSync(resolve(tmpdir(), prefix));
  tempDirs.push(dir);
  return dir;
}

async function loadTaskStoreModule() {
  await vi.resetModules();
  return import("../task-store.mjs");
}

afterEach(() => {
  while (tempDirs.length) {
    const dir = tempDirs.pop();
    rmSync(dir, { recursive: true, force: true });
  }
});

describe("task-store path configuration", () => {
  it("configureTaskStore changes active store path", async () => {
    const taskStore = await loadTaskStoreModule();
    const tempDir = makeTempDir("ve-task-store-");
    const customStorePath = resolve(tempDir, "custom-state.json");

    taskStore.configureTaskStore({ storePath: customStorePath });

    expect(taskStore.getStorePath()).toBe(customStorePath);
  });

  it("getStorePath returns path configured via baseDir", async () => {
    const taskStore = await loadTaskStoreModule();
    const baseDir = makeTempDir("ve-task-store-base-");
    const expectedPath = resolve(
      baseDir,
      ".codex-monitor",
      ".cache",
      "kanban-state.json",
    );

    taskStore.configureTaskStore({ baseDir });

    expect(taskStore.getStorePath()).toBe(expectedPath);
  });

  it("defaults to shared repo-scoped state path when repoRoot is provided", async () => {
    const taskStore = await loadTaskStoreModule();
    const repoRoot = makeTempDir("ve-task-store-repo-");
    const stateDir = makeTempDir("ve-task-store-state-");

    taskStore.configureTaskStore({ repoRoot, stateDir, cwd: repoRoot });

    const resolvedPath = taskStore.getStorePath().replace(/\\/g, "/");
    expect(resolvedPath).toContain(stateDir.replace(/\\/g, "/"));
    expect(resolvedPath).toMatch(/\/repos\/repo-[a-f0-9]{16}\/kanban-state\.json$/);
  });

  it("migrates from legacy repo-local store when shared file is absent", async () => {
    const taskStore = await loadTaskStoreModule();
    const repoRoot = makeTempDir("ve-task-store-migrate-repo-");
    const stateDir = makeTempDir("ve-task-store-migrate-state-");
    const legacyDir = resolve(repoRoot, ".codex-monitor", ".cache");
    const legacyPath = resolve(legacyDir, "kanban-state.json");

    mkdirSync(legacyDir, { recursive: true });
    writeFileSync(
      legacyPath,
      JSON.stringify(
        {
          _meta: { version: 1 },
          tasks: {
            "legacy-task": {
              id: "legacy-task",
              title: "Legacy Task",
              status: "todo",
              statusHistory: [],
            },
          },
        },
        null,
        2,
      ),
      "utf8",
    );

    taskStore.configureTaskStore({ repoRoot, stateDir, cwd: repoRoot });
    taskStore.loadStore();

    const task = taskStore.getTask("legacy-task");
    expect(task).not.toBeNull();
    expect(task.title).toBe("Legacy Task");
  });

  it("reconfigure resets in-memory load state without throwing", async () => {
    const taskStore = await loadTaskStoreModule();
    const firstDir = makeTempDir("ve-task-store-first-");
    const secondDir = makeTempDir("ve-task-store-second-");
    const firstPath = resolve(firstDir, "first.json");
    const secondPath = resolve(secondDir, "second.json");

    taskStore.configureTaskStore({ storePath: firstPath });
    taskStore.loadStore();
    taskStore.addTask({ id: "task-1", title: "One" });
    expect(taskStore.getAllTasks().length).toBe(1);

    expect(() =>
      taskStore.configureTaskStore({ storePath: secondPath }),
    ).not.toThrow();

    taskStore.loadStore();
    expect(taskStore.getAllTasks()).toEqual([]);
  });
});
