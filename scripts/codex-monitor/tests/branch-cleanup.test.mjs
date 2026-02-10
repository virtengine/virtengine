import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { mkdir, rm, writeFile, mkdtemp } from "node:fs/promises";
import { resolve } from "node:path";
import { existsSync } from "node:fs";
import { tmpdir } from "node:os";
import { execSync } from "node:child_process";

import { cleanupStaleBranches } from "../maintenance.mjs";

let TEST_DIR = "";

/**
 * Helper: initialise a bare-bones git repo with an initial commit on `main`,
 * then optionally create local-only branches for testing.
 */
async function initTestRepo() {
  TEST_DIR = await mkdtemp(resolve(tmpdir(), "branch-cleanup-test-"));
  execSync("git init -b main", { cwd: TEST_DIR, windowsHide: true });
  execSync('git config user.email "test@test.com"', {
    cwd: TEST_DIR,
    windowsHide: true,
  });
  execSync('git config user.name "Test"', {
    cwd: TEST_DIR,
    windowsHide: true,
  });
  await writeFile(resolve(TEST_DIR, "README.md"), "# test");
  execSync("git add -A && git commit -m init", {
    cwd: TEST_DIR,
    windowsHide: true,
  });
  return TEST_DIR;
}

function createBranch(name, { backdate } = {}) {
  execSync(`git branch ${name}`, { cwd: TEST_DIR, windowsHide: true });
  if (backdate) {
    // Create a commit on the branch with an old date
    execSync(`git checkout ${name}`, { cwd: TEST_DIR, windowsHide: true });
    const dateStr = new Date(Date.now() - backdate).toISOString();
    execSync(`git commit --allow-empty -m "old commit" --date="${dateStr}"`, {
      cwd: TEST_DIR,
      windowsHide: true,
      env: {
        ...process.env,
        GIT_COMMITTER_DATE: dateStr,
        GIT_AUTHOR_DATE: dateStr,
      },
    });
    execSync("git checkout main", { cwd: TEST_DIR, windowsHide: true });
  }
}

function branchExists(name) {
  try {
    execSync(`git rev-parse --verify refs/heads/${name}`, {
      cwd: TEST_DIR,
      windowsHide: true,
      stdio: "pipe",
    });
    return true;
  } catch {
    return false;
  }
}

describe("cleanupStaleBranches", () => {
  beforeEach(async () => {
    await initTestRepo();
  });

  afterEach(async () => {
    if (existsSync(TEST_DIR)) {
      await rm(TEST_DIR, { recursive: true, force: true });
    }
  });

  it("should return empty results when no VE branches exist", () => {
    const result = cleanupStaleBranches(TEST_DIR);
    expect(result.deleted).toEqual([]);
    expect(result.skipped).toEqual([]);
    expect(result.errors).toEqual([]);
  });

  it("should skip protected branches", () => {
    // main already exists as the init branch and is protected by default
    const result = cleanupStaleBranches(TEST_DIR, { minAgeMs: 0 });
    // main should never appear in deleted
    expect(result.deleted).not.toContain("main");
  });

  it("should skip branches that are too recent", () => {
    // Create a ve/ branch without backdating — it's just seconds old
    createBranch("ve/test-recent");
    const result = cleanupStaleBranches(TEST_DIR);
    expect(result.deleted).toEqual([]);
    const skipped = result.skipped.find((s) => s.branch === "ve/test-recent");
    expect(skipped).toBeDefined();
    expect(skipped.reason).toBe("too-recent");
  });

  it("should skip the currently checked-out branch", () => {
    execSync("git checkout -b ve/current-branch", {
      cwd: TEST_DIR,
      windowsHide: true,
    });
    const result = cleanupStaleBranches(TEST_DIR, { minAgeMs: 0 });
    expect(result.deleted).toEqual([]);
    const skipped = result.skipped.find(
      (s) => s.branch === "ve/current-branch",
    );
    expect(skipped).toBeDefined();
    expect(skipped.reason).toBe("checked-out");
    // Restore
    execSync("git checkout main", { cwd: TEST_DIR, windowsHide: true });
  });

  it("should delete old merged branches", () => {
    // Create a ve/ branch, merge it into main, backdate it
    const twodays = 2 * 24 * 60 * 60 * 1000;
    createBranch("ve/old-merged", { backdate: twodays });

    // Merge into main
    execSync("git merge --no-ff ve/old-merged -m merge", {
      cwd: TEST_DIR,
      windowsHide: true,
    });

    expect(branchExists("ve/old-merged")).toBe(true);
    const result = cleanupStaleBranches(TEST_DIR);
    expect(result.deleted).toContain("ve/old-merged");
    expect(branchExists("ve/old-merged")).toBe(false);
  });

  it("should skip old branches that are not pushed and not merged", () => {
    const twodays = 2 * 24 * 60 * 60 * 1000;
    createBranch("ve/old-unmerged", { backdate: twodays });

    const result = cleanupStaleBranches(TEST_DIR);
    const skipped = result.skipped.find((s) => s.branch === "ve/old-unmerged");
    expect(skipped).toBeDefined();
    expect(skipped.reason).toBe("not-pushed-not-merged");
    expect(branchExists("ve/old-unmerged")).toBe(true);
  });

  it("should support dry-run mode", () => {
    const twodays = 2 * 24 * 60 * 60 * 1000;
    createBranch("ve/dry-run-test", { backdate: twodays });
    execSync("git merge --no-ff ve/dry-run-test -m merge", {
      cwd: TEST_DIR,
      windowsHide: true,
    });

    const result = cleanupStaleBranches(TEST_DIR, { dryRun: true });
    expect(result.deleted).toContain("ve/dry-run-test");
    // Branch should still exist because it's dry-run
    expect(branchExists("ve/dry-run-test")).toBe(true);
  });

  it("should only target specified patterns", () => {
    const twodays = 2 * 24 * 60 * 60 * 1000;
    createBranch("ve/should-target", { backdate: twodays });
    createBranch("feature/should-ignore", { backdate: twodays });

    // Merge both
    execSync("git merge --no-ff ve/should-target -m merge1", {
      cwd: TEST_DIR,
      windowsHide: true,
    });
    execSync("git merge --no-ff feature/should-ignore -m merge2", {
      cwd: TEST_DIR,
      windowsHide: true,
    });

    const result = cleanupStaleBranches(TEST_DIR);
    expect(result.deleted).toContain("ve/should-target");
    // feature/ branch should not be touched
    expect(branchExists("feature/should-ignore")).toBe(true);
  });

  it("should handle custom patterns", () => {
    const twodays = 2 * 24 * 60 * 60 * 1000;
    createBranch("custom/old-branch", { backdate: twodays });
    execSync("git merge --no-ff custom/old-branch -m merge", {
      cwd: TEST_DIR,
      windowsHide: true,
    });

    const result = cleanupStaleBranches(TEST_DIR, {
      patterns: ["custom/"],
    });
    expect(result.deleted).toContain("custom/old-branch");
  });

  it("should handle null repoRoot gracefully", () => {
    const result = cleanupStaleBranches(null);
    expect(result.deleted).toEqual([]);
    expect(result.skipped).toEqual([]);
    expect(result.errors).toEqual([]);
  });

  it("should respect custom minAgeMs", () => {
    // Branch is ~1 hour old
    const oneHour = 60 * 60 * 1000;
    createBranch("ve/hour-old", { backdate: oneHour + 5000 });
    execSync("git merge --no-ff ve/hour-old -m merge", {
      cwd: TEST_DIR,
      windowsHide: true,
    });

    // With default 24h threshold, should skip
    const result1 = cleanupStaleBranches(TEST_DIR, {
      minAgeMs: 24 * 60 * 60 * 1000,
    });
    expect(result1.deleted).not.toContain("ve/hour-old");

    // With 30min threshold, should delete
    const result2 = cleanupStaleBranches(TEST_DIR, {
      minAgeMs: 30 * 60 * 1000,
    });
    expect(result2.deleted).toContain("ve/hour-old");
  });
});
