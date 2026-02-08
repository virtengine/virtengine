/**
 * fleet-coordinator.mjs — Multi-workstation coordination for codex-monitor.
 *
 * Provides:
 *   - Repo fingerprinting: detect identical repos across workstations
 *   - Fleet discovery: enumerate active instances working on the same repo
 *   - Coordinator election integration: single leader dispatches tasks
 *   - Task slot aggregation: total parallel capacity across fleet
 *   - Conflict-aware task dispatch: order + assign tasks to minimize conflicts
 *   - Maintenance mode: when backlog is empty, fleet enters maintenance
 *
 * The coordinator (elected via presence.mjs) is the only instance that:
 *   1. Triggers the task planner to generate new backlog items
 *   2. Assigns execution order and workstation routing hints
 *   3. Broadcasts fleet status updates
 *
 * Non-coordinator instances:
 *   - Report their presence and capacity
 *   - Pull tasks from VK backlog in the assigned order
 *   - Contribute shared knowledge entries
 */

import crypto from "node:crypto";
import { execSync } from "node:child_process";
import { readFile, writeFile, mkdir } from "node:fs/promises";
import { existsSync, readFileSync } from "node:fs";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import {
  initPresence,
  buildLocalPresence,
  listActiveInstances,
  selectCoordinator,
  getPresenceState,
} from "./presence.mjs";

const __dirname = dirname(fileURLToPath(import.meta.url));

// ── Repo Fingerprinting ──────────────────────────────────────────────────────

/**
 * Generate a stable fingerprint for a git repository.
 * Two workstations with the same repo will produce the same fingerprint.
 *
 * Components (in order of reliability):
 *   1. Remote origin URL (normalized — strips .git suffix, protocol variance)
 *   2. Fallback: first commit hash (immutable root of the repo)
 */
export function computeRepoFingerprint(repoRoot) {
  if (!repoRoot) return null;

  // Try remote origin URL first (most reliable for same-repo detection)
  let remoteUrl = null;
  try {
    remoteUrl = execSync("git config --get remote.origin.url", {
      cwd: repoRoot,
      encoding: "utf8",
      stdio: ["ignore", "pipe", "ignore"],
    }).trim();
  } catch {
    // no remote configured
  }

  if (remoteUrl) {
    const normalized = normalizeGitUrl(remoteUrl);
    return {
      method: "remote-origin",
      raw: remoteUrl,
      normalized,
      hash: hashString(normalized),
    };
  }

  // Fallback: first commit hash (root of the DAG)
  try {
    const rootCommit = execSync("git rev-list --max-parents=0 HEAD", {
      cwd: repoRoot,
      encoding: "utf8",
      stdio: ["ignore", "pipe", "ignore"],
    }).trim().split("\n")[0];

    if (rootCommit) {
      return {
        method: "root-commit",
        raw: rootCommit,
        normalized: rootCommit,
        hash: hashString(rootCommit),
      };
    }
  } catch {
    // not a git repo or no commits
  }

  return null;
}

/**
 * Normalize a git URL to strip protocol/auth/suffix variance.
 * Examples:
 *   https://github.com/virtengine/virtengine.git → github.com/virtengine/virtengine
 *   git@github.com:virtengine/virtengine.git     → github.com/virtengine/virtengine
 *   ssh://git@github.com/virtengine/virtengine   → github.com/virtengine/virtengine
 */
export function normalizeGitUrl(url) {
  if (!url) return "";
  let s = String(url).trim();

  // Strip protocol
  s = s.replace(/^(?:https?|ssh|git):\/\//, "");

  // Strip user@ prefix (git@github.com: or user@host/)
  s = s.replace(/^[^@]+@/, "");

  // Normalize SSH colon syntax (github.com:org/repo → github.com/org/repo)
  s = s.replace(/^([^/:]+):/, "$1/");

  // Strip .git suffix
  s = s.replace(/\.git$/, "");

  // Strip trailing slashes
  s = s.replace(/\/+$/, "");

  return s.toLowerCase();
}

function hashString(s) {
  return crypto.createHash("sha256").update(s).digest("hex").slice(0, 16);
}

// ── Fleet State ──────────────────────────────────────────────────────────────

const FLEET_STATE_FILENAME = "fleet-state.json";

const fleetState = {
  initialized: false,
  repoFingerprint: null,
  isCoordinator: false,
  fleetSize: 0,
  totalSlots: 0,
  localSlots: 0,
  mode: "solo", // solo | fleet | maintenance
  peers: [],        // instances with same repo fingerprint
  dispatchOrder: [], // task IDs in conflict-minimized order
  lastSyncAt: null,
};

/**
 * Initialize fleet coordination.
 * Must be called after presence.mjs is initialized.
 *
 * @param {object} opts
 * @param {string} opts.repoRoot - Git repository root
 * @param {number} opts.localSlots - Max parallel agents on this workstation
 * @param {number} [opts.ttlMs] - Presence TTL (default: 5 min)
 * @param {object} [opts.localWorkspace] - Workspace identity from registry
 */
export async function initFleet(opts = {}) {
  const { repoRoot, localSlots = 6, ttlMs = 5 * 60 * 1000, localWorkspace } = opts;

  if (!repoRoot) {
    console.warn("[fleet] No repo root provided — running in solo mode");
    fleetState.mode = "solo";
    fleetState.localSlots = localSlots;
    fleetState.totalSlots = localSlots;
    fleetState.initialized = true;
    return fleetState;
  }

  // Compute repo fingerprint
  fleetState.repoFingerprint = computeRepoFingerprint(repoRoot);
  fleetState.localSlots = localSlots;

  // Ensure presence is initialized
  await initPresence({ repoRoot, localWorkspace });

  // Discover fleet
  await refreshFleet({ ttlMs });

  fleetState.initialized = true;
  console.log(
    `[fleet] initialized: mode=${fleetState.mode}, peers=${fleetState.fleetSize}, ` +
    `totalSlots=${fleetState.totalSlots}, fingerprint=${fleetState.repoFingerprint?.hash || "none"}`,
  );

  return fleetState;
}

/**
 * Refresh fleet state from presence data.
 * Called periodically by the maintenance loop.
 */
export async function refreshFleet({ ttlMs = 5 * 60 * 1000 } = {}) {
  const nowMs = Date.now();
  const allInstances = listActiveInstances({ nowMs, ttlMs });
  const localFingerprint = fleetState.repoFingerprint?.hash;

  if (!localFingerprint || allInstances.length <= 1) {
    // Solo mode — only us
    fleetState.mode = "solo";
    fleetState.fleetSize = 1;
    fleetState.totalSlots = fleetState.localSlots;
    fleetState.peers = [];
    fleetState.isCoordinator = true; // solo = always coordinator
    fleetState.lastSyncAt = new Date().toISOString();
    return fleetState;
  }

  // Filter to peers with matching repo fingerprint
  const peers = allInstances.filter((inst) => {
    const peerFingerprint = inst.repo_fingerprint;
    return peerFingerprint && peerFingerprint === localFingerprint;
  });

  fleetState.peers = peers;
  fleetState.fleetSize = Math.max(1, peers.length);

  // Aggregate capacity
  let totalSlots = 0;
  for (const peer of peers) {
    totalSlots += typeof peer.max_parallel === "number" ? peer.max_parallel : 6;
  }
  // Ensure we count ourselves even if not yet in the presence list
  if (!peers.some((p) => p.instance_id === getPresenceState().instance_id)) {
    totalSlots += fleetState.localSlots;
    fleetState.fleetSize += 1;
  }
  fleetState.totalSlots = totalSlots || fleetState.localSlots;

  // Determine if we're the coordinator
  const coordinator = selectCoordinator({ nowMs, ttlMs });
  const myId = getPresenceState().instance_id;
  fleetState.isCoordinator = coordinator?.instance_id === myId;

  // Fleet vs solo
  fleetState.mode = fleetState.fleetSize > 1 ? "fleet" : "solo";
  fleetState.lastSyncAt = new Date().toISOString();

  return fleetState;
}

// ── Fleet-Aware Presence Payload ─────────────────────────────────────────────

/**
 * Build a presence payload enriched with fleet coordination data.
 * This is broadcast to other instances so they can match repos.
 */
export function buildFleetPresence(extra = {}) {
  const base = buildLocalPresence(extra);
  return {
    ...base,
    repo_fingerprint: fleetState.repoFingerprint?.hash || null,
    max_parallel: fleetState.localSlots,
    fleet_mode: fleetState.mode,
    is_coordinator: fleetState.isCoordinator,
  };
}

// ── Conflict-Aware Task Ordering ────────────────────────────────────────────

/**
 * File-path based conflict graph for tasks.
 * Tasks touching overlapping file paths should not run in parallel.
 *
 * @param {Array<{id: string, title: string, scope?: string, filePaths?: string[]}>} tasks
 * @returns {Array<Array<string>>} waves — groups of task IDs safe for parallel execution
 */
export function buildExecutionWaves(tasks) {
  if (!tasks || tasks.length === 0) return [];

  // Build scope-based conflict sets (tasks with same scope conflict)
  const scopeMap = new Map(); // scope → [taskId, ...]
  const fileMap = new Map();  // filePath → [taskId, ...]
  const taskById = new Map();

  for (const task of tasks) {
    const id = task.id || task.title;
    taskById.set(id, task);

    // Scope-based conflicts
    const scope = task.scope || extractScopeFromTask(task.title);
    if (scope) {
      if (!scopeMap.has(scope)) scopeMap.set(scope, []);
      scopeMap.get(scope).push(id);
    }

    // File-path based conflicts (when available)
    if (Array.isArray(task.filePaths)) {
      for (const fp of task.filePaths) {
        const normalizedPath = fp.replace(/\\/g, "/").toLowerCase();
        if (!fileMap.has(normalizedPath)) fileMap.set(normalizedPath, []);
        fileMap.get(normalizedPath).push(id);
      }
    }
  }

  // Build adjacency list (conflict graph)
  const conflicts = new Map(); // taskId → Set<conflicting taskIds>
  for (const [, taskIds] of [...scopeMap, ...fileMap]) {
    if (taskIds.length > 1) {
      for (let i = 0; i < taskIds.length; i++) {
        for (let j = i + 1; j < taskIds.length; j++) {
          if (!conflicts.has(taskIds[i])) conflicts.set(taskIds[i], new Set());
          if (!conflicts.has(taskIds[j])) conflicts.set(taskIds[j], new Set());
          conflicts.get(taskIds[i]).add(taskIds[j]);
          conflicts.get(taskIds[j]).add(taskIds[i]);
        }
      }
    }
  }

  // Greedy graph coloring (Welsh-Powell) for wave assignment
  const allIds = tasks.map((t) => t.id || t.title);
  const sortedIds = [...allIds].sort((a, b) => {
    const ca = conflicts.get(a)?.size || 0;
    const cb = conflicts.get(b)?.size || 0;
    return cb - ca; // highest degree first
  });

  const waves = [];
  const assigned = new Set();

  for (const taskId of sortedIds) {
    if (assigned.has(taskId)) continue;

    // Find first wave this task can join (no conflicts with existing members)
    let placed = false;
    for (const wave of waves) {
      const hasConflict = wave.some(
        (wId) => conflicts.get(taskId)?.has(wId) || conflicts.get(wId)?.has(taskId),
      );
      if (!hasConflict) {
        wave.push(taskId);
        assigned.add(taskId);
        placed = true;
        break;
      }
    }

    if (!placed) {
      waves.push([taskId]);
      assigned.add(taskId);
    }
  }

  return waves;
}

/**
 * Extract scope from a task title (conventional commit format).
 * E.g., "feat(veid): add flow" → "veid"
 */
function extractScopeFromTask(title) {
  if (!title) return null;
  const m = title.match(
    /^(?:\[P\d+\]\s*)?(?:feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)\(([^)]+)\)/i,
  );
  return m ? m[1].toLowerCase() : null;
}

// ── Workstation Task Assignment ──────────────────────────────────────────────

/**
 * Given a set of execution waves and fleet peers, assign tasks to workstations.
 * Returns a dispatch plan that each workstation can consume.
 *
 * @param {Array<Array<string>>} waves - Output of buildExecutionWaves
 * @param {Array<{instance_id: string, max_parallel?: number, capabilities?: string[]}>} peers
 * @returns {object} dispatchPlan
 */
export function assignTasksToWorkstations(waves, peers, taskMap = new Map()) {
  if (!peers || peers.length === 0 || !waves || waves.length === 0) {
    return { assignments: [], totalTasks: 0, totalPeers: 0 };
  }

  const assignments = [];
  let waveIndex = 0;

  for (const wave of waves) {
    waveIndex++;
    const waveAssignments = [];

    // Round-robin distribute tasks in this wave across peers
    for (let i = 0; i < wave.length; i++) {
      const taskId = wave[i];
      const peer = peers[i % peers.length];
      const task = taskMap.get(taskId);

      // Try capability-based routing: if task has a scope/capability hint
      // and a peer has matching capabilities, prefer that peer
      let bestPeer = peer;
      if (task?.scope) {
        const capMatch = peers.find((p) =>
          Array.isArray(p.capabilities) &&
          p.capabilities.some((c) =>
            c.toLowerCase().includes(task.scope.toLowerCase()),
          ),
        );
        if (capMatch) bestPeer = capMatch;
      }

      waveAssignments.push({
        taskId,
        taskTitle: task?.title || taskId,
        wave: waveIndex,
        assignedTo: bestPeer.instance_id,
        assignedToLabel: bestPeer.instance_label || bestPeer.instance_id,
      });
    }

    assignments.push(...waveAssignments);
  }

  return {
    assignments,
    totalTasks: assignments.length,
    totalPeers: peers.length,
    waveCount: waves.length,
    createdAt: new Date().toISOString(),
  };
}

// ── Backlog Depth Calculator ─────────────────────────────────────────────────

/**
 * Calculate how many tasks should be in the backlog based on fleet size.
 * More workstations = deeper backlog to keep everyone busy.
 *
 * @param {object} opts
 * @param {number} opts.totalSlots - Total parallel agent slots across fleet
 * @param {number} opts.currentBacklog - Current todo tasks in VK
 * @param {number} [opts.bufferMultiplier=3] - How many rounds of work to buffer
 * @param {number} [opts.minTasks=6] - Minimum backlog depth
 * @param {number} [opts.maxTasks=100] - Cap to prevent over-generation
 * @returns {object} { targetDepth, deficit, shouldGenerate }
 */
export function calculateBacklogDepth(opts = {}) {
  const {
    totalSlots = 6,
    currentBacklog = 0,
    bufferMultiplier = 3,
    minTasks = 6,
    maxTasks = 100,
  } = opts;

  // Target: enough tasks for N full rounds of parallel execution
  const rawTarget = totalSlots * bufferMultiplier;
  const targetDepth = Math.max(minTasks, Math.min(rawTarget, maxTasks));
  const deficit = Math.max(0, targetDepth - currentBacklog);

  return {
    totalSlots,
    currentBacklog,
    targetDepth,
    deficit,
    shouldGenerate: deficit > 0,
    formula: `${totalSlots} slots × ${bufferMultiplier} buffer = ${rawTarget} (clamped to ${targetDepth})`,
  };
}

// ── Maintenance Mode Detection ───────────────────────────────────────────────

/**
 * Determine if the fleet should enter maintenance mode.
 * Maintenance mode means: all functional work is done, switch to
 * housekeeping (dependency updates, test coverage, refactoring, docs).
 *
 * @param {object} status - VK project status
 * @returns {object} { isMaintenanceMode, reason }
 */
export function detectMaintenanceMode(status) {
  if (!status) return { isMaintenanceMode: false, reason: "no status data" };

  const counts = status.counts || {};
  const backlog = status.backlog_remaining ?? 0;
  const running = counts.running ?? 0;
  const review = counts.review ?? 0;
  const todo = counts.todo ?? 0;

  // Maintenance mode: nothing to do AND nothing in progress
  if (backlog === 0 && todo === 0 && running === 0 && review === 0) {
    return {
      isMaintenanceMode: true,
      reason: "all tasks completed — no backlog, no active work",
    };
  }

  return {
    isMaintenanceMode: false,
    reason: `active: backlog=${backlog} todo=${todo} running=${running} review=${review}`,
  };
}

// ── Fleet State Persistence ──────────────────────────────────────────────────

const FLEET_STATE_DIR = ".cache/codex-monitor";

async function getFleetStatePath(repoRoot) {
  const dir = resolve(repoRoot || process.cwd(), FLEET_STATE_DIR);
  await mkdir(dir, { recursive: true });
  return resolve(dir, FLEET_STATE_FILENAME);
}

export async function persistFleetState(repoRoot) {
  try {
    const path = await getFleetStatePath(repoRoot);
    const payload = {
      ...fleetState,
      peers: fleetState.peers.map((p) => ({
        instance_id: p.instance_id,
        instance_label: p.instance_label,
        max_parallel: p.max_parallel,
        capabilities: p.capabilities,
        host: p.host,
      })),
      updatedAt: new Date().toISOString(),
    };
    await writeFile(path, JSON.stringify(payload, null, 2), "utf8");
  } catch (err) {
    console.warn(`[fleet] persist error: ${err.message}`);
  }
}

export async function loadFleetState(repoRoot) {
  try {
    const path = await getFleetStatePath(repoRoot);
    if (!existsSync(path)) return null;
    const raw = await readFile(path, "utf8");
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

// ── Public Getters ───────────────────────────────────────────────────────────

export function getFleetState() {
  return { ...fleetState };
}

export function isFleetCoordinator() {
  return fleetState.isCoordinator;
}

export function getFleetMode() {
  return fleetState.mode;
}

export function getFleetSize() {
  return fleetState.fleetSize;
}

export function getTotalFleetSlots() {
  return fleetState.totalSlots;
}

/**
 * Format a human-readable fleet status summary.
 */
export function formatFleetSummary() {
  const fp = fleetState.repoFingerprint;
  const lines = [
    `🛰️ Fleet Status: ${fleetState.mode.toUpperCase()}`,
    `Repo: ${fp?.normalized || "unknown"} (${fp?.hash?.slice(0, 8) || "no fingerprint"})`,
    `Coordinator: ${fleetState.isCoordinator ? "THIS INSTANCE" : "remote"}`,
    `Fleet size: ${fleetState.fleetSize} workstation(s)`,
    `Total slots: ${fleetState.totalSlots}`,
    `Local slots: ${fleetState.localSlots}`,
  ];

  if (fleetState.peers.length > 0) {
    lines.push("", "Peers:");
    for (const peer of fleetState.peers) {
      const label = peer.instance_label || peer.instance_id;
      const slots = peer.max_parallel ?? "?";
      const host = peer.host || "unknown";
      const coordTag = peer.is_coordinator ? " ⭐" : "";
      lines.push(`  • ${label}${coordTag} — ${host} (${slots} slots)`);
    }
  }

  if (fleetState.lastSyncAt) {
    lines.push("", `Last sync: ${fleetState.lastSyncAt}`);
  }

  return lines.join("\n");
}
