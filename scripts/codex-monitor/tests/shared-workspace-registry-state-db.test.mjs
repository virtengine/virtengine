import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { mkdir, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { existsSync } from "node:fs";
import { resolve } from "node:path";
import { tmpdir } from "node:os";
import {
  loadSharedWorkspaceRegistry,
  resolveSharedWorkspaceStatePaths,
} from "../shared-workspace-registry.mjs";

let TEST_DIR = "";

describe("shared workspace registry state db", () => {
  beforeEach(async () => {
    TEST_DIR = await mkdtemp(resolve(tmpdir(), "codex-shared-state-"));
  });

  afterEach(async () => {
    if (TEST_DIR && existsSync(TEST_DIR)) {
      await rm(TEST_DIR, { recursive: true, force: true });
    }
  });

  it("resolves global state-db-backed registry paths", () => {
    const repoRoot = resolve(TEST_DIR, "repo-a");
    const stateDbDir = resolve(TEST_DIR, "global-state");
    const paths = resolveSharedWorkspaceStatePaths({ repoRoot, stateDbDir });

    expect(paths.repo_id).toMatch(/^repo-[a-f0-9]{16}$/);
    expect(paths.registry_path.startsWith(stateDbDir)).toBe(true);
    expect(paths.audit_log_path.startsWith(stateDbDir)).toBe(true);
    expect(paths.state_db_path).toBe(resolve(stateDbDir, "repositories.json"));
  });

  it("migrates from legacy repo cache to global state db path", async () => {
    const repoRoot = resolve(TEST_DIR, "repo-b");
    const stateDbDir = resolve(TEST_DIR, "global-state");
    const legacyDir = resolve(repoRoot, ".cache", "codex-monitor");
    const legacyRegistryPath = resolve(legacyDir, "shared-workspaces.json");

    await mkdir(legacyDir, { recursive: true });
    await writeFile(
      legacyRegistryPath,
      JSON.stringify(
        {
          version: 1,
          registry_name: "legacy",
          default_lease_ttl_minutes: 60,
          workspaces: [{ id: "legacy-ws", name: "Legacy Workspace" }],
        },
        null,
        2,
      ),
      "utf8",
    );

    const loaded = await loadSharedWorkspaceRegistry({
      repoRoot,
      cwd: repoRoot,
      stateDbDir,
    });

    expect(loaded.loaded_from_legacy_cache).toBe(true);
    expect(loaded.workspaces).toHaveLength(1);
    expect(loaded.workspaces[0].id).toBe("legacy-ws");
    expect(loaded.registry_path.startsWith(stateDbDir)).toBe(true);
    expect(existsSync(loaded.registry_path)).toBe(true);
    expect(existsSync(resolve(stateDbDir, "repositories.json"))).toBe(true);

    await rm(legacyRegistryPath, { force: true });
    const reloaded = await loadSharedWorkspaceRegistry({
      repoRoot,
      cwd: repoRoot,
      stateDbDir,
    });

    expect(reloaded.loaded_from_legacy_cache).toBe(false);
    expect(reloaded.workspaces).toHaveLength(1);
    expect(reloaded.workspaces[0].id).toBe("legacy-ws");

    const dbRaw = await readFile(resolve(stateDbDir, "repositories.json"), "utf8");
    const db = JSON.parse(dbRaw);
    expect(Object.keys(db.repositories || {}).length).toBeGreaterThan(0);
  });
});
