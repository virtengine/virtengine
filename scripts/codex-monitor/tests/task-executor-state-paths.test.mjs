import { afterEach, describe, expect, it, vi } from "vitest";
import { mkdtempSync, mkdirSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { resolve } from "node:path";

const tempDirs = [];

function makeTempDir(prefix) {
  const dir = mkdtempSync(resolve(tmpdir(), prefix));
  tempDirs.push(dir);
  return dir;
}

async function loadModule() {
  await vi.resetModules();
  return import("../task-executor.mjs");
}

afterEach(() => {
  while (tempDirs.length) {
    rmSync(tempDirs.pop(), { recursive: true, force: true });
  }
});

describe("task-executor shared state paths", () => {
  it("resolves runtime state paths under shared repo identity dir", async () => {
    const mod = await loadModule();
    const repoRoot = makeTempDir("ve-task-executor-repo-");
    const stateDir = makeTempDir("ve-task-executor-state-");
    mkdirSync(resolve(repoRoot, ".git"), { recursive: true });

    const paths = mod.resolveTaskExecutorStatePaths({
      repoRoot,
      cwd: repoRoot,
      stateDir,
    });

    const normalizedNoCommit = paths.noCommitPath.replace(/\\/g, "/");
    const normalizedRuntime = paths.runtimePath.replace(/\\/g, "/");
    const normalizedStateDir = stateDir.replace(/\\/g, "/");

    expect(normalizedNoCommit.startsWith(normalizedStateDir)).toBe(true);
    expect(normalizedRuntime.startsWith(normalizedStateDir)).toBe(true);
    expect(normalizedNoCommit).toMatch(/\/repos\/repo-[a-f0-9]{16}\/no-commit-state\.json$/);
    expect(normalizedRuntime).toMatch(/\/repos\/repo-[a-f0-9]{16}\/task-executor-runtime\.json$/);
  });

  it("keeps legacy candidate paths for repo-local migration", async () => {
    const mod = await loadModule();
    const repoRoot = makeTempDir("ve-task-executor-legacy-");
    mkdirSync(resolve(repoRoot, ".git"), { recursive: true });

    const paths = mod.resolveTaskExecutorStatePaths({ repoRoot, cwd: repoRoot });

    expect(paths.legacyNoCommitPaths).toContain(
      resolve(repoRoot, ".cache", "codex-monitor", "no-commit-state.json"),
    );
    expect(paths.legacyRuntimePaths).toContain(
      resolve(repoRoot, ".cache", "codex-monitor", "task-executor-runtime.json"),
    );
  });
});
