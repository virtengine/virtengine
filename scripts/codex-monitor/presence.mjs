import { execSync } from "node:child_process";
import crypto from "node:crypto";
import { existsSync } from "node:fs";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import os from "node:os";
import { dirname, resolve } from "node:path";
import { resolveRepoSharedStatePaths } from "./shared-state-paths.mjs";

const PRESENCE_PREFIX = "[ve-presence]";
const PRESENCE_VERSION = 1;
const INSTANCE_ID_FILENAME = "instance-id.json";
const PRESENCE_FILENAME = "presence.json";

const state = {
  initialized: false,
  repoRoot: null,
  presencePath: null,
  instanceId: null,
  startedAt: new Date().toISOString(),
  localWorkspace: null,
  localMeta: null,
  instances: new Map(),
};

function safeParseNumber(value, fallback) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

async function ensurePresenceDir(repoRoot) {
  const dir = resolveRepoSharedStatePaths({ repoRoot }).repoStateDir;
  await mkdir(dir, { recursive: true });
  return dir;
}

async function loadOrCreateInstanceId(paths, explicitId) {
  if (explicitId) {
    return String(explicitId).trim();
  }
  await mkdir(dirname(paths.instanceIdPath), { recursive: true });
  const candidatePaths = [
    paths.instanceIdPath,
    ...(paths.legacyInstanceIdPaths || []),
  ];

  for (const idPath of candidatePaths) {
    if (!existsSync(idPath)) continue;
    try {
      const raw = await readFile(idPath, "utf8");
      const parsed = JSON.parse(raw);
      if (parsed?.instance_id) {
        const value = String(parsed.instance_id).trim();
        if (idPath !== paths.instanceIdPath) {
          await writeFile(
            paths.instanceIdPath,
            JSON.stringify(
              { instance_id: value, created_at: parsed.created_at || new Date().toISOString() },
              null,
              2,
            ),
            "utf8",
          );
        }
        return value;
      }
    } catch {
      // best effort
    }
  }
  const host = os.hostname() || "host";
  const suffix = crypto.randomUUID().slice(0, 8);
  const newId = `${host}-${suffix}`;
  await writeFile(
    paths.instanceIdPath,
    JSON.stringify({ instance_id: newId, created_at: new Date().toISOString() }, null, 2),
    "utf8",
  );
  return newId;
}

function resolvePresencePaths(options = {}) {
  const shared = resolveRepoSharedStatePaths({
    repoRoot: options.repoRoot,
    cwd: options.cwd,
    stateDir: options.stateDir,
    stateRoot: options.stateRoot,
    repoIdentity: options.repoIdentity,
  });
  return {
    repoRoot: shared.repoRoot,
    presencePath: options.presencePath || shared.file(PRESENCE_FILENAME),
    instanceIdPath: options.instanceIdPath || shared.file(INSTANCE_ID_FILENAME),
    legacyPresencePaths: [
      resolve(shared.legacyCacheDir, PRESENCE_FILENAME),
      resolve(shared.legacyCodexCacheDir, PRESENCE_FILENAME),
    ],
    legacyInstanceIdPaths: [
      resolve(shared.legacyCacheDir, INSTANCE_ID_FILENAME),
      resolve(shared.legacyCodexCacheDir, INSTANCE_ID_FILENAME),
    ],
  };
}

function readGitInfo(repoRoot) {
  const info = { git_branch: null, git_sha: null };
  try {
    const branch = execSync("git rev-parse --abbrev-ref HEAD", {
      cwd: repoRoot,
      encoding: "utf8",
      stdio: ["ignore", "pipe", "ignore"],
    });
    info.git_branch = branch.trim();
  } catch {
    // ignore
  }
  try {
    const sha = execSync("git rev-parse --short HEAD", {
      cwd: repoRoot,
      encoding: "utf8",
      stdio: ["ignore", "pipe", "ignore"],
    });
    info.git_sha = sha.trim();
  } catch {
    // ignore
  }
  return info;
}

function buildLocalMeta() {
  const envLabel = process.env.VE_INSTANCE_LABEL || "";
  const priority = safeParseNumber(
    process.env.VE_COORDINATOR_PRIORITY,
    null,
  );
  const eligibleRaw = process.env.VE_COORDINATOR_ELIGIBLE;
  const eligible =
    eligibleRaw === undefined
      ? true
      : !["0", "false", "no"].includes(String(eligibleRaw).toLowerCase());

  const workspace = state.localWorkspace || {};
  const role = workspace.role || "workspace";
  const basePriority = role === "coordinator" ? 10 : 100;

  return {
    v: PRESENCE_VERSION,
    instance_id: state.instanceId,
    instance_label: envLabel || workspace.name || null,
    workspace_id: workspace.id || null,
    workspace_role: role,
    coordinator_priority: Number.isFinite(priority) ? priority : basePriority,
    coordinator_eligible: eligible,
    capabilities: Array.isArray(workspace.capabilities)
      ? workspace.capabilities
      : [],
    host: os.hostname(),
    platform: os.platform(),
    arch: os.arch(),
    node: process.version,
    pid: process.pid,
    started_at: state.startedAt,
  };
}

async function loadPresenceRegistry() {
  if (!state.presencePath) return;
  const candidatePaths = [state.presencePath, ...(state.legacyPresencePaths || [])];
  for (const path of candidatePaths) {
    if (!existsSync(path)) continue;
    try {
      const raw = await readFile(path, "utf8");
      const parsed = JSON.parse(raw);
      if (Array.isArray(parsed?.instances)) {
        for (const entry of parsed.instances) {
          if (entry?.instance_id) {
            state.instances.set(String(entry.instance_id), entry);
          }
        }
        if (path !== state.presencePath && !existsSync(state.presencePath)) {
          await persistPresenceRegistry();
        }
        return;
      }
    } catch {
      // ignore
    }
  }
}

async function persistPresenceRegistry() {
  if (!state.presencePath) return;
  const instances = [...state.instances.values()];
  const payload = {
    updated_at: new Date().toISOString(),
    instances,
  };
  await writeFile(state.presencePath, JSON.stringify(payload, null, 2), "utf8");
}

function normalizePresencePayload(payload) {
  if (!payload || !payload.instance_id) return null;
  return {
    ...payload,
    instance_id: String(payload.instance_id),
    workspace_id: payload.workspace_id ? String(payload.workspace_id) : null,
    workspace_role: payload.workspace_role || payload.role || null,
    coordinator_priority: safeParseNumber(payload.coordinator_priority, 100),
    coordinator_eligible:
      payload.coordinator_eligible === undefined ? true : !!payload.coordinator_eligible,
  };
}

export async function initPresence(options = {}) {
  const forceReset = options.force || process.env.VITEST;
  if (state.initialized && !forceReset) return state;
  if (forceReset) {
    state.initialized = false;
    state.repoRoot = null;
    state.presencePath = null;
    state.instanceId = null;
    state.startedAt = new Date().toISOString();
    state.localWorkspace = null;
    state.localMeta = null;
    state.instances = new Map();
  }
  const paths = resolvePresencePaths(options);
  state.repoRoot = paths.repoRoot;
  state.presencePath = paths.presencePath;
  state.legacyPresencePaths = paths.legacyPresencePaths;
  state.localWorkspace = options.localWorkspace || null;
  state.instanceId = await loadOrCreateInstanceId(
    {
      instanceIdPath: paths.instanceIdPath,
      legacyInstanceIdPaths: paths.legacyInstanceIdPaths,
    },
    options.instanceId || process.env.VE_INSTANCE_ID,
  );
  state.localMeta = buildLocalMeta();
  await mkdir(dirname(state.presencePath), { recursive: true });
  const shouldLoadRegistry =
    options.loadRegistry ?? (!options.skipLoad && !process.env.VITEST);
  if (shouldLoadRegistry) {
    await loadPresenceRegistry();
  }
  state.initialized = true;
  return state;
}

export function getPresencePrefix() {
  return PRESENCE_PREFIX;
}

export function formatPresenceMessage(payload) {
  return `${PRESENCE_PREFIX} ${JSON.stringify(payload)}`;
}

export function parsePresenceMessage(text) {
  const raw = String(text || "");
  const idx = raw.indexOf(PRESENCE_PREFIX);
  if (idx === -1) return null;
  const jsonPart = raw.slice(idx + PRESENCE_PREFIX.length).trim();
  if (!jsonPart.startsWith("{")) return null;
  try {
    const parsed = JSON.parse(jsonPart);
    return normalizePresencePayload(parsed);
  } catch {
    return null;
  }
}

export function buildLocalPresence(extra = {}) {
  const meta = state.localMeta || {};
  const gitInfo = state.repoRoot ? readGitInfo(state.repoRoot) : {};
  return normalizePresencePayload({
    ...meta,
    ...gitInfo,
    ...extra,
    updated_at: new Date().toISOString(),
  });
}

export async function notePresence(payload, options = {}) {
  const normalized = normalizePresencePayload(payload);
  if (!normalized) return null;
  const now = options.receivedAt || new Date().toISOString();
  const entry = {
    ...normalized,
    last_seen_at: now,
    source: options.source || normalized.source || "telegram",
  };
  state.instances.set(normalized.instance_id, entry);
  await persistPresenceRegistry();
  return entry;
}

export function listActiveInstances({ nowMs, ttlMs } = {}) {
  const now = Number.isFinite(nowMs) ? nowMs : Date.now();
  const ttl = Number.isFinite(ttlMs) ? ttlMs : 0;
  const instances = [];
  for (const entry of state.instances.values()) {
    const last = Date.parse(entry.last_seen_at || entry.updated_at || "");
    if (ttl > 0 && (!Number.isFinite(last) || now - last >= ttl)) {
      continue;
    }
    instances.push(entry);
  }
  return instances;
}

export function selectCoordinator({ nowMs, ttlMs } = {}) {
  const active = listActiveInstances({ nowMs, ttlMs });
  if (!active.length) return null;
  const eligible = active.filter((entry) => entry.coordinator_eligible !== false);
  const pool = eligible.length ? eligible : active;
  const preferred = pool.filter(
    (entry) => String(entry.workspace_role || "").toLowerCase() === "coordinator",
  );
  const candidates = preferred.length ? preferred : pool;
  const sorted = [...candidates].sort((a, b) => {
    const pa = safeParseNumber(a.coordinator_priority, 100);
    const pb = safeParseNumber(b.coordinator_priority, 100);
    if (pa !== pb) return pa - pb;
    const sa = Date.parse(a.started_at || "");
    const sb = Date.parse(b.started_at || "");
    if (Number.isFinite(sa) && Number.isFinite(sb) && sa !== sb) {
      return sa - sb;
    }
    return String(a.instance_id).localeCompare(String(b.instance_id));
  });
  return sorted[0] || null;
}

export function formatPresenceSummary({ nowMs, ttlMs } = {}) {
  const active = listActiveInstances({ nowMs, ttlMs });
  if (!active.length) {
    return "No active instances reported.";
  }
  const coordinator = selectCoordinator({ nowMs, ttlMs });
  const lines = ["🛰️ Codex Monitor Presence"];
  for (const entry of active) {
    const name = entry.instance_label || entry.instance_id;
    const role = entry.workspace_role || "workspace";
    const host = entry.host || "unknown";
    const lastSeen = entry.last_seen_at
      ? entry.last_seen_at.slice(11, 19)
      : "--:--:--";
    const coordTag =
      coordinator && coordinator.instance_id === entry.instance_id ? " (coordinator)" : "";
    lines.push(`- ${name}${coordTag} — ${role} @ ${host} (last ${lastSeen})`);
  }
  return lines.join("\n");
}

export function formatCoordinatorSummary({ nowMs, ttlMs } = {}) {
  const coordinator = selectCoordinator({ nowMs, ttlMs });
  if (!coordinator) {
    return "No coordinator selected (no active instances).";
  }
  const name = coordinator.instance_label || coordinator.instance_id;
  const role = coordinator.workspace_role || "workspace";
  const host = coordinator.host || "unknown";
  const lastSeen = coordinator.last_seen_at || coordinator.updated_at || "unknown";
  return [
    "⭐ Coordinator",
    `Instance: ${name}`,
    `Role: ${role}`,
    `Host: ${host}`,
    `Last seen: ${lastSeen}`,
  ].join("\n");
}

export function getPresenceState() {
  return {
    instance_id: state.instanceId,
    started_at: state.startedAt,
    instances: [...state.instances.values()],
  };
}
