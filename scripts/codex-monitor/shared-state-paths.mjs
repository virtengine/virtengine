import { homedir } from "node:os";
import { resolve } from "node:path";
import { resolveRepoIdentity } from "./repo-root.mjs";

function readStateRootOverride(options = {}) {
  if (options.stateDir) return String(options.stateDir).trim();
  if (options.stateRoot) return String(options.stateRoot).trim();
  return (
    process.env.VE_CODEX_MONITOR_STATE_DIR ||
    process.env.VE_STATE_DB_DIR ||
    process.env.VK_STATE_DB_DIR ||
    ""
  )
    .toString()
    .trim();
}

export function resolveSharedStateRoot(options = {}) {
  const override = readStateRootOverride(options);
  if (override) {
    return resolve(override);
  }

  const localAppData = process.env.LOCALAPPDATA || "";
  if (localAppData) {
    return resolve(localAppData, "codex-monitor", "state");
  }
  return resolve(homedir(), ".codex-monitor", "state");
}

export function resolveRepoSharedStatePaths(options = {}) {
  const identity =
    options.repoIdentity ||
    resolveRepoIdentity({ cwd: options.cwd, repoRoot: options.repoRoot });
  const stateRoot = resolveSharedStateRoot(options);
  const repoDir = resolve(stateRoot, "repos", identity.repoId);
  const legacyCacheDir = resolve(identity.repoRoot, ".cache", "codex-monitor");
  const legacyCodexCacheDir = resolve(identity.repoRoot, ".codex-monitor", ".cache");

  return {
    repoId: identity.repoId,
    repoRoot: identity.repoRoot,
    stateRoot,
    repoStateDir: repoDir,
    identitySeed: identity.identitySeed,
    originUrl: identity.originUrl || null,
    gitCommonDir: identity.gitCommonDir || null,
    legacyCacheDir,
    legacyCodexCacheDir,
    file: (...segments) => resolve(repoDir, ...segments),
  };
}

export function resolveRepoSharedStateFile(relativePath, options = {}) {
  return resolveRepoSharedStatePaths(options).file(relativePath);
}
