import { execSync } from "node:child_process";
import { createHash } from "node:crypto";
import { resolve } from "node:path";

/**
 * Resolve the repo root for codex-monitor.
 *
 * Priority:
 *  1. Explicit REPO_ROOT env var.
 *  2. git rev-parse --show-toplevel (relative to cwd).
 *  3. process.cwd().
 */
export function resolveRepoRoot(options = {}) {
  if (options.repoRoot) {
    return resolve(options.repoRoot);
  }

  const envRoot = process.env.REPO_ROOT;
  if (envRoot) return resolve(envRoot);

  const cwd = options.cwd || process.cwd();
  try {
    const gitRoot = execSync("git rev-parse --show-toplevel", {
      encoding: "utf8",
      cwd,
      stdio: ["ignore", "pipe", "ignore"],
    }).trim();
    if (gitRoot) return gitRoot;
  } catch {
    // ignore - fall back to cwd
  }

  return resolve(cwd);
}

function runGitCommand(command, options = {}) {
  const cwd = options.cwd || process.cwd();
  try {
    const output = execSync(command, {
      encoding: "utf8",
      cwd,
      stdio: ["ignore", "pipe", "ignore"],
    }).trim();
    return output || null;
  } catch {
    return null;
  }
}

function normalizeIdentityValue(value) {
  return String(value || "")
    .trim()
    .replaceAll("\\", "/")
    .toLowerCase();
}

function normalizeOriginUrl(url) {
  const raw = String(url || "").trim();
  if (!raw) return "";
  const cleaned = raw.replace(/\.git$/i, "");
  return normalizeIdentityValue(cleaned);
}

function toRepoId(seed) {
  const hash = createHash("sha256").update(String(seed || "")).digest("hex");
  return `repo-${hash.slice(0, 16)}`;
}

/**
 * Resolve a canonical repository identity so codex-monitor state can be shared
 * across different working directories (worktrees, symlinked roots, etc.).
 */
export function resolveRepoIdentity(options = {}) {
  const cwd = options.cwd || process.cwd();
  const repoRoot = resolveRepoRoot(options);
  const gitCommonDirRaw =
    runGitCommand("git rev-parse --git-common-dir", { cwd: repoRoot }) ||
    runGitCommand("git rev-parse --absolute-git-dir", { cwd: repoRoot });
  const gitCommonDir = gitCommonDirRaw ? resolve(repoRoot, gitCommonDirRaw) : null;
  const originUrl = runGitCommand("git config --get remote.origin.url", {
    cwd: repoRoot,
  });

  const normalizedOrigin = normalizeOriginUrl(originUrl);
  const normalizedGitCommonDir = normalizeIdentityValue(gitCommonDir);
  const normalizedRepoRoot = normalizeIdentityValue(repoRoot);

  let seed = `repo-root:${normalizedRepoRoot}`;
  if (normalizedOrigin) {
    seed = `origin:${normalizedOrigin}`;
  } else if (normalizedGitCommonDir) {
    seed = `git-common-dir:${normalizedGitCommonDir}`;
  }

  return {
    repoRoot,
    cwd: resolve(cwd),
    gitCommonDir,
    originUrl: originUrl || null,
    identitySeed: seed,
    repoId: toRepoId(seed),
  };
}
