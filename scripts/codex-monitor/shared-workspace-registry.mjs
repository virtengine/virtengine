import { randomUUID } from "node:crypto";
import { existsSync } from "node:fs";
import { mkdir, readFile, rename, writeFile } from "node:fs/promises";
import { homedir } from "node:os";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { resolveRepoIdentity } from "./repo-root.mjs";

const __dirname = resolve(fileURLToPath(new URL(".", import.meta.url)));

const DEFAULT_LEASE_TTL_MINUTES = 120;
const DEFAULT_REGISTRY = {
  version: 1,
  registry_name: "shared-cloud-workspaces",
  default_lease_ttl_minutes: DEFAULT_LEASE_TTL_MINUTES,
  workspaces: [],
};

const DEFAULT_STATE_DB = {
  version: 1,
  repositories: {},
};

const AVAILABILITY_STATES = new Set([
  "available",
  "leased",
  "maintenance",
  "offline",
  "disabled",
]);

const AVAILABILITY_ALIASES = {
  idle: "available",
  free: "available",
  busy: "leased",
  inuse: "leased",
};

function normalizeId(value) {
  return String(value || "").trim().toLowerCase();
}

function normalizeAvailability(value) {
  const raw = String(value || "").trim().toLowerCase();
  if (!raw) return "available";
  const aliased = AVAILABILITY_ALIASES[raw] || raw;
  return AVAILABILITY_STATES.has(aliased) ? aliased : "available";
}

function toIso(value) {
  if (!value) return null;
  const ts = Date.parse(value);
  if (!Number.isFinite(ts)) return null;
  return new Date(ts).toISOString();
}

function ensureIso(date) {
  return new Date(date).toISOString();
}

function normalizeLease(lease) {
  if (!lease) return null;
  const owner = String(lease.owner || "").trim();
  const claimedAt = toIso(lease.claimed_at);
  const expiresAt = toIso(lease.lease_expires_at);
  if (!owner || !claimedAt || !expiresAt) return null;
  const ttlMinutes = Number(lease.lease_ttl_minutes || 0);
  return {
    lease_id: lease.lease_id || randomUUID(),
    owner,
    claimed_at: claimedAt,
    lease_expires_at: expiresAt,
    lease_ttl_minutes: Number.isFinite(ttlMinutes) && ttlMinutes > 0 ? ttlMinutes : null,
    last_renewed_at: toIso(lease.last_renewed_at) || claimedAt,
    notes: lease.notes || "",
  };
}

function normalizeWorkspace(workspace) {
  if (!workspace) return null;
  const id = normalizeId(workspace.id);
  if (!id) return null;
  const availability = normalizeAvailability(workspace.availability);
  const lease = normalizeLease(workspace.lease);
  const resolvedAvailability = lease ? "leased" : availability;
  return {
    id,
    name: workspace.name || workspace.id || id,
    provider: workspace.provider || "vibe-kanban",
    region: workspace.region || "",
    owner: workspace.owner || "",
    availability_before_lease: workspace.availability_before_lease || null,
    availability: resolvedAvailability,
    lease,
    lease_ttl_minutes: workspace.lease_ttl_minutes || null,
    metadata: workspace.metadata || {},
  };
}

function normalizeRegistry(raw) {
  const registry = raw && typeof raw === "object" ? raw : {};
  const workspaces = Array.isArray(registry.workspaces)
    ? registry.workspaces.map(normalizeWorkspace).filter(Boolean)
    : [];
  const ttlMinutes = Number(registry.default_lease_ttl_minutes || 0);
  return {
    version: registry.version || DEFAULT_REGISTRY.version,
    registry_name: registry.registry_name || DEFAULT_REGISTRY.registry_name,
    default_lease_ttl_minutes:
      Number.isFinite(ttlMinutes) && ttlMinutes > 0
        ? ttlMinutes
        : DEFAULT_REGISTRY.default_lease_ttl_minutes,
    workspaces,
  };
}

function resolveStateDbDir(options = {}) {
  if (options.stateDbDir) {
    return resolve(options.stateDbDir);
  }
  const envPath =
    process.env.VE_CODEX_MONITOR_STATE_DIR ||
    process.env.VE_STATE_DB_DIR ||
    process.env.VK_STATE_DB_DIR ||
    "";
  if (envPath) {
    return resolve(envPath);
  }
  const localAppData = process.env.LOCALAPPDATA || "";
  if (localAppData) {
    return resolve(localAppData, "codex-monitor", "state");
  }
  return resolve(homedir(), ".codex-monitor", "state");
}

function getStateDbPath(options = {}) {
  if (options.stateDbPath) {
    return resolve(options.stateDbPath);
  }
  return resolve(resolveStateDbDir(options), "repositories.json");
}

function normalizeRepositoryRecord(value) {
  const record = value && typeof value === "object" ? value : {};
  return {
    repo_id: String(record.repo_id || "").trim(),
    identity_seed: String(record.identity_seed || "").trim(),
    repo_root: String(record.repo_root || "").trim(),
    git_common_dir: String(record.git_common_dir || "").trim(),
    origin_url: String(record.origin_url || "").trim(),
    registry_path: String(record.registry_path || "").trim(),
    audit_log_path: String(record.audit_log_path || "").trim(),
    legacy_registry_path: String(record.legacy_registry_path || "").trim(),
    legacy_audit_log_path: String(record.legacy_audit_log_path || "").trim(),
    created_at: toIso(record.created_at),
    updated_at: toIso(record.updated_at),
  };
}

function normalizeStateDb(raw) {
  const payload = raw && typeof raw === "object" ? raw : {};
  const repositories = {};
  if (payload.repositories && typeof payload.repositories === "object") {
    for (const [repoId, record] of Object.entries(payload.repositories)) {
      const id = String(repoId || "").trim();
      if (!id) continue;
      repositories[id] = normalizeRepositoryRecord(record);
    }
  }
  return {
    version: payload.version || DEFAULT_STATE_DB.version,
    repositories,
  };
}

async function loadStateDb(path) {
  if (!existsSync(path)) {
    return normalizeStateDb(DEFAULT_STATE_DB);
  }
  try {
    const raw = await readFile(path, "utf8");
    return normalizeStateDb(JSON.parse(raw));
  } catch {
    return normalizeStateDb(DEFAULT_STATE_DB);
  }
}

async function writeStateDb(path, payload) {
  await mkdir(resolve(path, ".."), { recursive: true });
  const tempPath = `${path}.tmp-${Date.now()}`;
  await writeFile(tempPath, JSON.stringify(payload, null, 2), "utf8");
  await rename(tempPath, path);
}

export function resolveSharedWorkspaceStatePaths(options = {}) {
  const identity =
    options.repoIdentity ||
    resolveRepoIdentity({ cwd: options.cwd, repoRoot: options.repoRoot });
  const stateDbDir = resolveStateDbDir(options);
  const stateDbPath = getStateDbPath(options);
  const repoDir = resolve(stateDbDir, "repos", identity.repoId);
  const legacyBase = resolve(identity.repoRoot, ".cache", "codex-monitor");
  return {
    repo_id: identity.repoId,
    identity_seed: identity.identitySeed,
    repo_root: identity.repoRoot,
    git_common_dir: identity.gitCommonDir || "",
    origin_url: identity.originUrl || "",
    state_db_dir: stateDbDir,
    state_db_path: stateDbPath,
    registry_path: resolve(repoDir, "shared-workspaces.json"),
    audit_log_path: resolve(repoDir, "shared-workspace-audit.jsonl"),
    legacy_registry_path: resolve(legacyBase, "shared-workspaces.json"),
    legacy_audit_log_path: resolve(legacyBase, "shared-workspace-audit.jsonl"),
  };
}

async function registerRepositoryState(options = {}) {
  const context = resolveSharedWorkspaceStatePaths(options);
  const db = await loadStateDb(context.state_db_path);
  const existing = db.repositories[context.repo_id] || {};
  const nowIso = ensureIso(new Date());
  db.repositories[context.repo_id] = {
    ...existing,
    repo_id: context.repo_id,
    identity_seed: context.identity_seed,
    repo_root: context.repo_root,
    git_common_dir: context.git_common_dir,
    origin_url: context.origin_url,
    registry_path: context.registry_path,
    audit_log_path: context.audit_log_path,
    legacy_registry_path: context.legacy_registry_path,
    legacy_audit_log_path: context.legacy_audit_log_path,
    created_at: toIso(existing.created_at) || nowIso,
    updated_at: nowIso,
  };
  await writeStateDb(context.state_db_path, db);
  return context;
}

function getRegistryPath(options = {}) {
  if (options.registryPath) {
    return resolve(options.registryPath);
  }
  const envPath =
    process.env.VE_SHARED_WORKSPACE_REGISTRY ||
    process.env.VE_SHARED_WORKSPACE_REGISTRY_PATH ||
    process.env.VK_SHARED_WORKSPACE_REGISTRY_PATH ||
    "";
  if (envPath) {
    return resolve(envPath);
  }
  return resolveSharedWorkspaceStatePaths(options).registry_path;
}

function getSeedPath(options = {}) {
  if (options.seedPath) {
    return resolve(options.seedPath);
  }
  return resolve(__dirname, "shared-workspaces.json");
}

function getAuditPath(options = {}) {
  if (options.auditPath) {
    return resolve(options.auditPath);
  }
  const envPath =
    process.env.VE_SHARED_WORKSPACE_AUDIT_LOG ||
    process.env.VE_SHARED_WORKSPACE_AUDIT_PATH ||
    process.env.VK_SHARED_WORKSPACE_AUDIT_PATH ||
    "";
  if (envPath) {
    return resolve(envPath);
  }
  return resolveSharedWorkspaceStatePaths(options).audit_log_path;
}

async function writeRegistryFile(path, registry) {
  await mkdir(resolve(path, ".."), { recursive: true });
  const payload = JSON.stringify(registry, null, 2);
  const tempPath = `${path}.tmp-${Date.now()}`;
  await writeFile(tempPath, payload, "utf8");
  await rename(tempPath, path);
}

async function appendAuditEntry(entry, options = {}) {
  const auditPath = getAuditPath(options);
  await mkdir(resolve(auditPath, ".."), { recursive: true });
  const payload = `${JSON.stringify(entry)}\n`;
  await writeFile(auditPath, payload, { encoding: "utf8", flag: "a" });
}

async function readRegistryFromFile(path, label = "registry") {
  if (!path || !existsSync(path)) {
    return null;
  }
  try {
    const raw = await readFile(path, "utf8");
    return normalizeRegistry(JSON.parse(raw));
  } catch (err) {
    console.warn(
      `[shared-workspace-registry] failed to read ${label} ${path}: ${err.message || err}`,
    );
    return null;
  }
}

export async function loadSharedWorkspaceRegistry(options = {}) {
  const stateContext = await registerRepositoryState(options);
  const registryPath = getRegistryPath({ ...options, ...stateContext });
  const legacyRegistryPath =
    options.legacyRegistryPath || stateContext.legacy_registry_path;
  let registry = await readRegistryFromFile(registryPath, "registry");
  let loadedFromLegacy = false;
  if (!registry && legacyRegistryPath) {
    registry = await readRegistryFromFile(legacyRegistryPath, "legacy registry");
    loadedFromLegacy = Boolean(registry);
  }
  if (!registry) {
    registry = await readRegistryFromFile(getSeedPath(options), "seed");
  }
  if (!registry) {
    registry = normalizeRegistry(DEFAULT_REGISTRY);
  }
  if (loadedFromLegacy && !existsSync(registryPath)) {
    await writeRegistryFile(registryPath, {
      version: registry.version || DEFAULT_REGISTRY.version,
      registry_name: registry.registry_name || DEFAULT_REGISTRY.registry_name,
      default_lease_ttl_minutes:
        registry.default_lease_ttl_minutes ||
        DEFAULT_REGISTRY.default_lease_ttl_minutes,
      workspaces: registry.workspaces || [],
    });
  }
  return {
    ...registry,
    registry_path: registryPath,
    registry_seed_path: getSeedPath(options),
    audit_log_path: getAuditPath({ ...options, ...stateContext }),
    legacy_registry_path: legacyRegistryPath,
    legacy_audit_log_path:
      options.legacyAuditPath || stateContext.legacy_audit_log_path,
    state_db_path: stateContext.state_db_path,
    repo_id: stateContext.repo_id,
    identity_seed: stateContext.identity_seed,
    loaded_from_legacy_cache: loadedFromLegacy,
  };
}

export async function saveSharedWorkspaceRegistry(registry, options = {}) {
  if (!registry) return;
  await registerRepositoryState(options);
  const path = registry.registry_path || getRegistryPath(options);
  const payload = {
    version: registry.version || DEFAULT_REGISTRY.version,
    registry_name: registry.registry_name || DEFAULT_REGISTRY.registry_name,
    default_lease_ttl_minutes:
      registry.default_lease_ttl_minutes || DEFAULT_REGISTRY.default_lease_ttl_minutes,
    workspaces: registry.workspaces || [],
  };
  await writeRegistryFile(path, payload);
}

export function resolveSharedWorkspace(registry, candidateId) {
  if (!registry || !Array.isArray(registry.workspaces)) return null;
  const target = normalizeId(candidateId);
  if (!target) return null;
  return registry.workspaces.find((ws) => ws.id === target) || null;
}

export function isLeaseExpired(lease, now = new Date()) {
  if (!lease || !lease.lease_expires_at) return false;
  const expiry = Date.parse(lease.lease_expires_at);
  if (!Number.isFinite(expiry)) return false;
  return expiry <= now.getTime();
}

function buildLease(owner, ttlMinutes, now, note) {
  const claimedAt = ensureIso(now);
  const expiresAt = ensureIso(now.getTime() + ttlMinutes * 60 * 1000);
  return {
    lease_id: randomUUID(),
    owner,
    claimed_at: claimedAt,
    lease_expires_at: expiresAt,
    lease_ttl_minutes: ttlMinutes,
    last_renewed_at: claimedAt,
    notes: note || "",
  };
}

function restoreAvailability(workspace) {
  const fallback = normalizeAvailability(
    workspace.availability_before_lease || "available",
  );
  const resolved = fallback === "leased" ? "available" : fallback;
  workspace.availability = resolved;
  workspace.availability_before_lease = null;
}

export async function sweepExpiredLeases(options = {}) {
  const now = options.now ? new Date(options.now) : new Date();
  const registry = options.registry
    ? options.registry
    : await loadSharedWorkspaceRegistry(options);
  if (!registry || !Array.isArray(registry.workspaces)) {
    return { registry, expired: [] };
  }
  const expired = [];
  for (const workspace of registry.workspaces) {
    if (!workspace?.lease) continue;
    if (!isLeaseExpired(workspace.lease, now)) continue;
    const lease = workspace.lease;
    workspace.lease = null;
    restoreAvailability(workspace);
    workspace.last_released_at = ensureIso(now);
    expired.push({ workspace, lease });
    await appendAuditEntry(
      {
        ts: ensureIso(now),
        action: "lease_expired",
        workspace_id: workspace.id,
        owner: lease.owner,
        lease_id: lease.lease_id,
        actor: options.actor || "system",
        lease_expires_at: lease.lease_expires_at,
      },
      options,
    );
  }
  if (expired.length > 0) {
    await saveSharedWorkspaceRegistry(registry, options);
  }
  return { registry, expired };
}

export async function claimSharedWorkspace(options = {}) {
  const now = options.now ? new Date(options.now) : new Date();
  let registry = options.registry
    ? options.registry
    : await loadSharedWorkspaceRegistry(options);
  const sweepResult = await sweepExpiredLeases({
    registry,
    now,
    actor: options.actor,
    auditPath: options.auditPath,
    registryPath: options.registryPath,
  });
  registry = sweepResult.registry;
  const workspace = resolveSharedWorkspace(registry, options.workspaceId);
  if (!workspace) {
    return { error: `Unknown shared workspace '${options.workspaceId}'.` };
  }
  if (workspace.lease && !options.force) {
    return {
      error: `Workspace '${workspace.id}' is already leased to ${workspace.lease.owner}.`,
    };
  }
  if (workspace.availability !== "available" && !options.force) {
    return {
      error: `Workspace '${workspace.id}' is not available (state: ${workspace.availability}).`,
    };
  }
  const ttlMinutes = Number(
    options.ttlMinutes || workspace.lease_ttl_minutes || registry.default_lease_ttl_minutes,
  );
  if (!Number.isFinite(ttlMinutes) || ttlMinutes <= 0) {
    return { error: "Invalid lease TTL minutes." };
  }
  const owner = String(options.owner || options.actor || "unknown").trim();
  if (!owner) {
    return { error: "Owner is required to claim a workspace." };
  }
  const previousLease = workspace.lease && options.force ? workspace.lease : null;
  workspace.availability_before_lease =
    workspace.availability === "leased"
      ? workspace.availability_before_lease || "available"
      : workspace.availability;
  workspace.availability = "leased";
  workspace.lease = buildLease(owner, ttlMinutes, now, options.note);
  workspace.last_claimed_at = ensureIso(now);

  await saveSharedWorkspaceRegistry(registry, options);
  if (previousLease) {
    await appendAuditEntry(
      {
        ts: ensureIso(now),
        action: "force_release",
        workspace_id: workspace.id,
        owner: previousLease.owner,
        lease_id: previousLease.lease_id,
        actor: options.actor || owner,
        reason: "overridden_by_claim",
      },
      options,
    );
  }
  await appendAuditEntry(
    {
      ts: ensureIso(now),
      action: "claim",
      workspace_id: workspace.id,
      owner,
      lease_id: workspace.lease.lease_id,
      lease_expires_at: workspace.lease.lease_expires_at,
      lease_ttl_minutes: ttlMinutes,
      actor: options.actor || owner,
      note: options.note || "",
    },
    options,
  );

  return { registry, workspace, lease: workspace.lease };
}

export async function releaseSharedWorkspace(options = {}) {
  const now = options.now ? new Date(options.now) : new Date();
  const registry = options.registry
    ? options.registry
    : await loadSharedWorkspaceRegistry(options);
  const workspace = resolveSharedWorkspace(registry, options.workspaceId);
  if (!workspace) {
    return { error: `Unknown shared workspace '${options.workspaceId}'.` };
  }
  if (!workspace.lease) {
    return { error: `Workspace '${workspace.id}' is not leased.` };
  }
  const owner = String(options.owner || "").trim();
  if (owner && normalizeId(owner) !== normalizeId(workspace.lease.owner) && !options.force) {
    return {
      error: `Workspace '${workspace.id}' is leased to ${workspace.lease.owner}. Use --force to release anyway.`,
    };
  }
  const previousLease = workspace.lease;
  workspace.lease = null;
  restoreAvailability(workspace);
  workspace.last_released_at = ensureIso(now);

  await saveSharedWorkspaceRegistry(registry, options);
  await appendAuditEntry(
    {
      ts: ensureIso(now),
      action: "release",
      workspace_id: workspace.id,
      owner: previousLease.owner,
      lease_id: previousLease.lease_id,
      actor: options.actor || owner || previousLease.owner,
      reason: options.reason || "",
    },
    options,
  );

  return { registry, workspace, previousLease };
}

export async function renewSharedWorkspaceLease(options = {}) {
  const now = options.now ? new Date(options.now) : new Date();
  const registry = options.registry
    ? options.registry
    : await loadSharedWorkspaceRegistry(options);
  const workspace = resolveSharedWorkspace(registry, options.workspaceId);
  if (!workspace) {
    return { error: `Unknown shared workspace '${options.workspaceId}'.` };
  }
  if (!workspace.lease) {
    return { error: `Workspace '${workspace.id}' is not currently leased.` };
  }
  const owner = String(options.owner || "").trim();
  if (owner && normalizeId(owner) !== normalizeId(workspace.lease.owner)) {
    return {
      error: `Workspace '${workspace.id}' is leased to ${workspace.lease.owner}, cannot renew.`,
    };
  }
  const ttlMinutes = Number(
    options.ttlMinutes || workspace.lease.lease_ttl_minutes || registry.default_lease_ttl_minutes,
  );
  if (!Number.isFinite(ttlMinutes) || ttlMinutes <= 0) {
    return { error: "Invalid lease TTL minutes for renewal." };
  }
  const previousExpiry = workspace.lease.lease_expires_at;
  const newExpiresAt = ensureIso(now.getTime() + ttlMinutes * 60 * 1000);
  workspace.lease.lease_expires_at = newExpiresAt;
  workspace.lease.last_renewed_at = ensureIso(now);
  workspace.lease.lease_ttl_minutes = ttlMinutes;

  await saveSharedWorkspaceRegistry(registry, options);
  await appendAuditEntry(
    {
      ts: ensureIso(now),
      action: "renew_lease",
      workspace_id: workspace.id,
      owner: workspace.lease.owner,
      lease_id: workspace.lease.lease_id,
      previous_expires_at: previousExpiry,
      new_expires_at: newExpiresAt,
      lease_ttl_minutes: ttlMinutes,
      actor: options.actor || workspace.lease.owner,
    },
    options,
  );

  return { registry, workspace, lease: workspace.lease };
}

function formatExpiresIn(expiresAt, now) {
  if (!expiresAt) return "unknown";
  const expiry = Date.parse(expiresAt);
  if (!Number.isFinite(expiry)) return "unknown";
  const diffMs = expiry - now.getTime();
  if (diffMs <= 0) return "expired";
  const minutes = Math.round(diffMs / 60000);
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  const remain = minutes % 60;
  if (remain === 0) return `${hours}h`;
  return `${hours}h${remain}m`;
}

export function formatSharedWorkspaceSummary(registry, options = {}) {
  const now = options.now ? new Date(options.now) : new Date();
  const lines = ["Shared Cloud Workspaces"];
  const workspaces = Array.isArray(registry?.workspaces) ? registry.workspaces : [];
  if (workspaces.length === 0) {
    lines.push("No shared workspaces configured.");
    return lines.join("\n");
  }
  for (const workspace of workspaces) {
    const base = `${workspace.id}: ${workspace.name || workspace.id}`;
    const availability = workspace.availability || "available";
    if (workspace.lease) {
      const expiresIn = formatExpiresIn(workspace.lease.lease_expires_at, now);
      lines.push(
        `- ${base} — leased by ${workspace.lease.owner} (expires in ${expiresIn})`,
      );
      continue;
    }
    lines.push(`- ${base} — ${availability}`);
  }
  return lines.join("\n");
}

export function formatSharedWorkspaceDetail(workspace, options = {}) {
  if (!workspace) return "Workspace not found.";
  const now = options.now ? new Date(options.now) : new Date();
  const lines = [`${workspace.id}: ${workspace.name || workspace.id}`];
  lines.push(`provider: ${workspace.provider || "vibe-kanban"}`);
  if (workspace.region) {
    lines.push(`region: ${workspace.region}`);
  }
  lines.push(`availability: ${workspace.availability || "available"}`);
  if (workspace.lease) {
    const lease = workspace.lease;
    lines.push(
      `lease owner: ${lease.owner}`,
      `lease expires: ${lease.lease_expires_at}`,
      `lease remaining: ${formatExpiresIn(lease.lease_expires_at, now)}`,
    );
    if (lease.notes) {
      lines.push(`lease notes: ${lease.notes}`);
    }
  }
  return lines.join("\n");
}

export function getSharedAvailabilityMap(registry) {
  const map = new Map();
  const workspaces = Array.isArray(registry?.workspaces)
    ? registry.workspaces
    : [];
  for (const workspace of workspaces) {
    if (!workspace?.id) continue;
    const state = workspace.lease
      ? "leased"
      : workspace.availability || "available";
    map.set(workspace.id, {
      state,
      owner: workspace.lease?.owner || null,
      lease_expires_at: workspace.lease?.lease_expires_at || null,
    });
  }
  return map;
}
export function getSharedRegistryTemplate() {
  return JSON.stringify(DEFAULT_REGISTRY, null, 2);
}
