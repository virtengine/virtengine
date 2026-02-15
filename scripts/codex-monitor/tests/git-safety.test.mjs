import { beforeEach, afterEach, describe, expect, it } from "vitest";
import { mkdtemp, mkdir, rm, writeFile } from "node:fs/promises";
import { execSync } from "node:child_process";
import { tmpdir } from "node:os";
import { resolve } from "node:path";

import {
  evaluateBranchSafetyForPush,
  normalizeBaseBranch,
} from "../git-safety.mjs";

describe("git-safety", () => {
  let repoDir = "";

  beforeEach(async () => {
    repoDir = await mkdtemp(resolve(tmpdir(), "git-safety-"));
    execSync("git init -b main", { cwd: repoDir, stdio: "pipe" });
    execSync('git config user.email "test@test.com"', {
      cwd: repoDir,
      stdio: "pipe",
    });
    execSync('git config user.name "Test"', { cwd: repoDir, stdio: "pipe" });

    await mkdir(resolve(repoDir, "src"), { recursive: true });
    const writes = [];
    for (let i = 0; i < 600; i += 1) {
      writes.push(
        writeFile(resolve(repoDir, "src", `file-${i}.txt`), `file ${i}\n`),
      );
    }
    await Promise.all(writes);
    execSync("git add -A && git commit -m base", { cwd: repoDir, stdio: "pipe" });

    const mainHead = execSync("git rev-parse HEAD", {
      cwd: repoDir,
      encoding: "utf8",
      stdio: "pipe",
    }).trim();
    // Simulate the ref normally present in worktrees with a fetched remote.
    execSync(`git update-ref refs/remotes/origin/main ${mainHead}`, {
      cwd: repoDir,
      stdio: "pipe",
    });
  });

  afterEach(async () => {
    if (repoDir) {
      await rm(repoDir, { recursive: true, force: true });
    }
  });

  it("normalizes base branch names", () => {
    expect(normalizeBaseBranch("main")).toEqual({
      branch: "main",
      remoteRef: "origin/main",
    });
    expect(normalizeBaseBranch("origin/main")).toEqual({
      branch: "main",
      remoteRef: "origin/main",
    });
    expect(normalizeBaseBranch("refs/remotes/origin/main")).toEqual({
      branch: "main",
      remoteRef: "origin/main",
    });
    expect(normalizeBaseBranch("origin/origin/main")).toEqual({
      branch: "main",
      remoteRef: "origin/main",
    });
  });

  it("flags README-only destructive branch states", async () => {
    execSync("git checkout -b ve/test", { cwd: repoDir, stdio: "pipe" });
    execSync("git rm -r --quiet src", { cwd: repoDir, stdio: "pipe" });
    await writeFile(resolve(repoDir, "README.md"), "# test\n");
    execSync("git add -A && git commit -m init", { cwd: repoDir, stdio: "pipe" });

    const safety = evaluateBranchSafetyForPush(repoDir, { baseBranch: "main" });
    expect(safety.safe).toBe(false);
    expect(safety.reason).toContain("HEAD tracks only");
  });

  it("supports explicit bypass for emergency pushes", async () => {
    execSync("git checkout -b ve/test", { cwd: repoDir, stdio: "pipe" });
    execSync("git rm -r --quiet src", { cwd: repoDir, stdio: "pipe" });
    await writeFile(resolve(repoDir, "README.md"), "# test\n");
    execSync("git add -A && git commit -m init", { cwd: repoDir, stdio: "pipe" });

    const prev = process.env.VE_ALLOW_DESTRUCTIVE_PUSH;
    process.env.VE_ALLOW_DESTRUCTIVE_PUSH = "1";
    try {
      const safety = evaluateBranchSafetyForPush(repoDir, {
        baseBranch: "main",
      });
      expect(safety.safe).toBe(true);
      expect(safety.bypassed).toBe(true);
    } finally {
      if (prev == null) delete process.env.VE_ALLOW_DESTRUCTIVE_PUSH;
      else process.env.VE_ALLOW_DESTRUCTIVE_PUSH = prev;
    }
  });
});
