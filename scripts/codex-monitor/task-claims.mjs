/**
 * task-claims.mjs — Distributed task claiming with idempotency and conflict resolution.
 *
 * Provides:
 *   - Idempotent task claiming across multiple workstations
 *   - Deterministic duplicate claim resolution
 *   - Persistent claim tokens
 *   - Integration with presence.mjs for fleet coordination
 *   - Telegram/VK channel announcement support
 *
 * Architecture:
 *   - Claims are stored in .cache/codex-monitor/task-claims.json
 *   - Each claim has a unique token (UUID) for idempotency
 *   - Claims include instance_id, timestamp, and TTL
 *   - Duplicate claims are resolved by instance priority (from presence.mjs)
 *   - Stale claims are auto-swept based on TTL
 *
 * Usage:
 *   import { claimTask, releaseTask, listClaims } from './task-claims.mjs';
 *
 *   const claim = await claimTask({
 *     taskId: 'abc123',
 *     instanceId: 'workstation-1',
 *     ttlMinutes: 60,
 *   });
 *
 *   if (claim.success) {
 *     // Work on task
 *     await releaseTask({ taskId: 'abc123', claimToken: claim.token });
 *   }
 */

import crypto from "node:crypto";
import { existsSync } from "node:fs";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { resolve } from "node:path";
import { getPresenceState, selectCoordinator } from "./presence.mjs";
import { resolveRepoSharedStatePaths } from "./shared-state-paths.mjs";

// ── Constants ────────────────────────────────────────────────────────────────

const CLAIMS_FILENAME = "task-claims.json";
const AUDIT_FILENAME = "task-claims-audit.jsonl";
const DEFAULT_TTL_MINUTES = 60;

// ── State ────────────────────────────────────────────────────────────────────

const state = {
  initialized: false,
  repoRoot: null,
  claimsPath: null,
  auditPath: null,
  legacyClaimsPath: null,
  legacyAuditPath: null,
};

// ── Initialization ───────────────────────────────────────────────────────────

/**
 * Initialize the task claims system.
 *
 * @param {object} opts
 * @param {string} [opts.repoRoot] - Repository root path
 * @returns {Promise<void>}
 */
export async function initTaskClaims(opts = {}) {
  const shared = resolveRepoSharedStatePaths({
    repoRoot: opts.repoRoot,
    cwd: opts.cwd,
    stateDir: opts.stateDir,
    stateRoot: opts.stateRoot,
    repoIdentity: opts.repoIdentity,
  });
  state.repoRoot = shared.repoRoot;
  const cacheDir = shared.repoStateDir;
  await mkdir(cacheDir, { recursive: true });
  state.claimsPath = opts.claimsPath || resolve(cacheDir, CLAIMS_FILENAME);
  state.auditPath = opts.auditPath || resolve(cacheDir, AUDIT_FILENAME);
  state.legacyClaimsPath =
    opts.legacyClaimsPath || resolve(shared.legacyCacheDir, CLAIMS_FILENAME);
  state.legacyAuditPath =
    opts.legacyAuditPath || resolve(shared.legacyCacheDir, AUDIT_FILENAME);
  state.initialized = true;
}

function ensureInitialized() {
  if (!state.initialized) {
    throw new Error("task-claims not initialized. Call initTaskClaims() first.");
  }
}

// ── Claim Registry I/O ───────────────────────────────────────────────────────

/**
 * Load the claims registry from disk.
 *
 * @returns {Promise<object>} Registry object with claims map
 */
async function loadClaimsRegistry() {
  ensureInitialized();
  const empty = { version: 1, claims: {}, updated_at: new Date().toISOString() };
  const candidatePaths = [state.claimsPath, state.legacyClaimsPath].filter(Boolean);
  for (const path of candidatePaths) {
    if (!existsSync(path)) continue;
    try {
      const raw = await readFile(path, "utf8");
      const data = JSON.parse(raw);
      const registry = {
        version: data.version || 1,
        claims: data.claims || {},
        updated_at: data.updated_at || new Date().toISOString(),
      };
      if (path !== state.claimsPath && !existsSync(state.claimsPath)) {
        await saveClaimsRegistry(registry);
      }
      return registry;
    } catch (err) {
      console.warn(`[task-claims] Failed to load registry: ${err.message}`);
    }
  }
  return empty;
}

/**
 * Save the claims registry to disk.
 *
 * @param {object} registry - Claims registry object
 * @returns {Promise<void>}
 */
async function saveClaimsRegistry(registry) {
  ensureInitialized();
  registry.updated_at = new Date().toISOString();
  await writeFile(state.claimsPath, JSON.stringify(registry, null, 2), "utf8");
}

// ── Audit Log ────────────────────────────────────────────────────────────────

/**
 * Append an audit entry to the claims audit log.
 *
 * @param {object} entry - Audit entry
 * @returns {Promise<void>}
 */
async function appendAuditEntry(entry) {
  ensureInitialized();
  const line = JSON.stringify({
    ...entry,
    timestamp: entry.timestamp || new Date().toISOString(),
  });
  try {
    await writeFile(state.auditPath, line + "\n", { flag: "a" });
  } catch (err) {
    console.warn(`[task-claims] Failed to write audit entry: ${err.message}`);
  }
}

// ── Claim Token Generation ───────────────────────────────────────────────────

/**
 * Generate a unique claim token.
 *
 * @returns {string} UUID-based claim token
 */
function generateClaimToken() {
  return crypto.randomUUID();
}

// ── Claim Expiry ─────────────────────────────────────────────────────────────

/**
 * Check if a claim is expired.
 *
 * @param {object} claim - Claim object
 * @param {Date} [now] - Current time (for testing)
 * @returns {boolean} True if expired
 */
function isClaimExpired(claim, now = new Date()) {
  if (!claim || !claim.expires_at) return true;
  const expiresAt = new Date(claim.expires_at);
  return now >= expiresAt;
}

/**
 * Sweep expired claims from the registry.
 *
 * @param {object} registry - Claims registry
 * @param {Date} [now] - Current time (for testing)
 * @returns {object} { registry, expiredCount }
 */
function sweepExpiredClaims(registry, now = new Date()) {
  let expiredCount = 0;
  for (const [taskId, claim] of Object.entries(registry.claims)) {
    if (isClaimExpired(claim, now)) {
      delete registry.claims[taskId];
      expiredCount++;
    }
  }
  return { registry, expiredCount };
}

// ── Duplicate Claim Resolution ───────────────────────────────────────────────

/**
 * Resolve a duplicate claim conflict deterministically.
 *
 * When two instances claim the same task, we resolve by:
 *   1. Coordinator priority (coordinator always wins)
 *   2. Coordinator priority number (lower wins)
 *   3. Timestamp (earlier claim wins)
 *   4. Instance ID (lexicographic comparison for determinism)
 *
 * @param {object} existingClaim - The existing claim
 * @param {object} newClaim - The new claim attempting to claim
 * @param {object} opts - Resolution options
 * @param {number} [opts.ttlMs] - Presence TTL for coordinator selection
 * @returns {object} { winner, loser, reason }
 */
function resolveDuplicateClaim(existingClaim, newClaim, opts = {}) {
  const { ttlMs = 5 * 60 * 1000 } = opts;
  const nowMs = Date.now();

  // Get coordinator from presence system
  const coordinator = selectCoordinator({ nowMs, ttlMs });
  const coordinatorId = coordinator?.instance_id;

  // Rule 1: Coordinator always wins
  if (coordinatorId) {
    if (existingClaim.instance_id === coordinatorId && newClaim.instance_id !== coordinatorId) {
      return {
        winner: existingClaim,
        loser: newClaim,
        reason: "existing_is_coordinator",
      };
    }
    if (newClaim.instance_id === coordinatorId && existingClaim.instance_id !== coordinatorId) {
      return {
        winner: newClaim,
        loser: existingClaim,
        reason: "new_is_coordinator",
      };
    }
  }

  // Rule 2: Lower coordinator priority wins (if both have priorities)
  const existingPriority = existingClaim.coordinator_priority ?? 100;
  const newPriority = newClaim.coordinator_priority ?? 100;
  if (existingPriority !== newPriority) {
    return existingPriority < newPriority
      ? {
          winner: existingClaim,
          loser: newClaim,
          reason: "existing_lower_priority",
        }
      : {
          winner: newClaim,
          loser: existingClaim,
          reason: "new_lower_priority",
        };
  }

  // Rule 3: Earlier timestamp wins
  const existingTime = new Date(existingClaim.claimed_at).getTime();
  const newTime = new Date(newClaim.claimed_at).getTime();
  if (existingTime !== newTime) {
    return existingTime < newTime
      ? {
          winner: existingClaim,
          loser: newClaim,
          reason: "existing_earlier",
        }
      : {
          winner: newClaim,
          loser: existingClaim,
          reason: "new_earlier",
        };
  }

  // Rule 4: Lexicographic instance ID comparison (deterministic tie-breaker)
  const comparison = existingClaim.instance_id.localeCompare(newClaim.instance_id);
  if (comparison < 0) {
    return {
      winner: existingClaim,
      loser: newClaim,
      reason: "existing_instance_id_lower",
    };
  } else if (comparison > 0) {
    return {
      winner: newClaim,
      loser: existingClaim,
      reason: "new_instance_id_lower",
    };
  }

  // Should never reach here (same instance claiming twice)
  return {
    winner: existingClaim,
    loser: newClaim,
    reason: "same_instance",
  };
}

// ── Core API ─────────────────────────────────────────────────────────────────

/**
 * Claim a task for this instance.
 *
 * @param {object} opts
 * @param {string} opts.taskId - Task ID to claim
 * @param {string} [opts.instanceId] - Instance ID (defaults to presence state)
 * @param {number} [opts.ttlMinutes] - Claim TTL in minutes
 * @param {string} [opts.claimToken] - Idempotency token (auto-generated if not provided)
 * @param {object} [opts.metadata] - Additional metadata
 * @returns {Promise<object>} { success, token, claim?, error?, resolution? }
 */
export async function claimTask(opts = {}) {
  ensureInitialized();

  const {
    taskId,
    instanceId = getPresenceState().instance_id,
    ttlMinutes = DEFAULT_TTL_MINUTES,
    claimToken = generateClaimToken(),
    metadata = {},
  } = opts;

  if (!taskId) {
    return { success: false, error: "taskId is required" };
  }

  if (!instanceId) {
    return { success: false, error: "instanceId is required" };
  }

  const now = new Date();
  const expiresAt = new Date(now.getTime() + ttlMinutes * 60 * 1000);

  // Load registry and sweep expired claims
  let registry = await loadClaimsRegistry();
  const sweepResult = sweepExpiredClaims(registry, now);
  registry = sweepResult.registry;

  // Check for existing claim
  const existingClaim = registry.claims[taskId];

  // Build new claim
  const presenceState = getPresenceState();
  const newClaim = {
    task_id: taskId,
    instance_id: instanceId,
    claim_token: claimToken,
    claimed_at: now.toISOString(),
    expires_at: expiresAt.toISOString(),
    ttl_minutes: ttlMinutes,
    coordinator_priority: presenceState.coordinator_priority ?? 100,
    metadata,
  };

  // If no existing claim, grant immediately
  if (!existingClaim) {
    registry.claims[taskId] = newClaim;
    await saveClaimsRegistry(registry);
    await appendAuditEntry({
      action: "claim",
      task_id: taskId,
      instance_id: instanceId,
      claim_token: claimToken,
      expires_at: expiresAt.toISOString(),
    });
    return { success: true, token: claimToken, claim: newClaim };
  }

  // Idempotency: If existing claim has same token, return it
  if (existingClaim.claim_token === claimToken) {
    return { success: true, token: claimToken, claim: existingClaim, idempotent: true };
  }

  // Duplicate claim detected — resolve conflict
  const resolution = resolveDuplicateClaim(existingClaim, newClaim);

  if (resolution.winner === newClaim) {
    // New claim wins — replace existing
    registry.claims[taskId] = newClaim;
    await saveClaimsRegistry(registry);
    await appendAuditEntry({
      action: "claim_override",
      task_id: taskId,
      instance_id: instanceId,
      claim_token: claimToken,
      expires_at: expiresAt.toISOString(),
      previous_instance: existingClaim.instance_id,
      previous_token: existingClaim.claim_token,
      resolution_reason: resolution.reason,
    });
    return {
      success: true,
      token: claimToken,
      claim: newClaim,
      resolution: {
        override: true,
        reason: resolution.reason,
        previous_instance: existingClaim.instance_id,
      },
    };
  } else {
    // Existing claim wins — reject new claim
    await appendAuditEntry({
      action: "claim_rejected",
      task_id: taskId,
      instance_id: instanceId,
      claim_token: claimToken,
      existing_instance: existingClaim.instance_id,
      existing_token: existingClaim.claim_token,
      resolution_reason: resolution.reason,
    });
    return {
      success: false,
      error: "task_already_claimed",
      existing_instance: existingClaim.instance_id,
      existing_claim: existingClaim,
      resolution: {
        override: false,
        reason: resolution.reason,
      },
    };
  }
}

/**
 * Release a claimed task.
 *
 * @param {object} opts
 * @param {string} opts.taskId - Task ID to release
 * @param {string} [opts.claimToken] - Claim token (for verification)
 * @param {string} [opts.instanceId] - Instance ID (defaults to presence state)
 * @param {boolean} [opts.force] - Force release even if not owned
 * @returns {Promise<object>} { success, error? }
 */
export async function releaseTask(opts = {}) {
  ensureInitialized();

  const {
    taskId,
    claimToken,
    instanceId = getPresenceState().instance_id,
    force = false,
  } = opts;

  if (!taskId) {
    return { success: false, error: "taskId is required" };
  }

  const registry = await loadClaimsRegistry();
  const claim = registry.claims[taskId];

  if (!claim) {
    return { success: false, error: "task_not_claimed" };
  }

  // Verify ownership unless force=true
  if (!force) {
    if (claim.instance_id !== instanceId) {
      return {
        success: false,
        error: "task_claimed_by_different_instance",
        owner: claim.instance_id,
      };
    }
    if (claimToken && claim.claim_token !== claimToken) {
      return {
        success: false,
        error: "claim_token_mismatch",
      };
    }
  }

  // Release the claim
  delete registry.claims[taskId];
  await saveClaimsRegistry(registry);
  await appendAuditEntry({
    action: force ? "release_forced" : "release",
    task_id: taskId,
    instance_id: instanceId,
    claim_token: claimToken,
    previous_owner: claim.instance_id,
  });

  return { success: true };
}

/**
 * Renew an existing claim (extend TTL).
 *
 * @param {object} opts
 * @param {string} opts.taskId - Task ID
 * @param {string} [opts.claimToken] - Claim token (for verification)
 * @param {string} [opts.instanceId] - Instance ID (defaults to presence state)
 * @param {number} [opts.ttlMinutes] - New TTL in minutes
 * @returns {Promise<object>} { success, claim?, error? }
 */
export async function renewClaim(opts = {}) {
  ensureInitialized();

  const {
    taskId,
    claimToken,
    instanceId = getPresenceState().instance_id,
    ttlMinutes = DEFAULT_TTL_MINUTES,
  } = opts;

  if (!taskId) {
    return { success: false, error: "taskId is required" };
  }

  const registry = await loadClaimsRegistry();
  const claim = registry.claims[taskId];

  if (!claim) {
    return { success: false, error: "task_not_claimed" };
  }

  // Verify ownership
  if (claim.instance_id !== instanceId) {
    return {
      success: false,
      error: "task_claimed_by_different_instance",
      owner: claim.instance_id,
    };
  }
  if (claimToken && claim.claim_token !== claimToken) {
    return {
      success: false,
      error: "claim_token_mismatch",
    };
  }

  // Renew the claim
  const now = new Date();
  const expiresAt = new Date(now.getTime() + ttlMinutes * 60 * 1000);
  claim.expires_at = expiresAt.toISOString();
  claim.ttl_minutes = ttlMinutes;
  claim.renewed_at = now.toISOString();

  await saveClaimsRegistry(registry);
  await appendAuditEntry({
    action: "renew",
    task_id: taskId,
    instance_id: instanceId,
    claim_token: claimToken,
    expires_at: expiresAt.toISOString(),
  });

  return { success: true, claim };
}

/**
 * Get a claim by task ID.
 *
 * @param {string} taskId - Task ID
 * @returns {Promise<object|null>} Claim object or null
 */
export async function getClaim(taskId) {
  ensureInitialized();
  const registry = await loadClaimsRegistry();
  return registry.claims[taskId] || null;
}

/**
 * List all active claims.
 *
 * @param {object} opts
 * @param {string} [opts.instanceId] - Filter by instance ID
 * @param {boolean} [opts.includeExpired] - Include expired claims
 * @returns {Promise<Array<object>>} Array of claim objects
 */
export async function listClaims(opts = {}) {
  ensureInitialized();
  const { instanceId, includeExpired = false } = opts;

  let registry = await loadClaimsRegistry();

  if (!includeExpired) {
    const sweepResult = sweepExpiredClaims(registry);
    registry = sweepResult.registry;
  }

  let claims = Object.values(registry.claims);

  if (instanceId) {
    claims = claims.filter((c) => c.instance_id === instanceId);
  }

  return claims;
}

/**
 * Check if a task is claimed.
 *
 * @param {string} taskId - Task ID
 * @returns {Promise<boolean>} True if claimed (and not expired)
 */
export async function isTaskClaimed(taskId) {
  ensureInitialized();
  const claim = await getClaim(taskId);
  if (!claim) return false;
  return !isClaimExpired(claim);
}

/**
 * Get claim statistics.
 *
 * @returns {Promise<object>} Statistics object
 */
export async function getClaimStats() {
  ensureInitialized();
  const registry = await loadClaimsRegistry();
  const now = new Date();

  let active = 0;
  let expired = 0;
  const byInstance = new Map();

  for (const claim of Object.values(registry.claims)) {
    if (isClaimExpired(claim, now)) {
      expired++;
    } else {
      active++;
      const count = byInstance.get(claim.instance_id) || 0;
      byInstance.set(claim.instance_id, count + 1);
    }
  }

  return {
    total: active + expired,
    active,
    expired,
    by_instance: Object.fromEntries(byInstance),
  };
}

// ── Public API ───────────────────────────────────────────────────────────────

// For testing
export const _test = {
  sweepExpiredClaims,
  resolveDuplicateClaim,
  isClaimExpired,
  loadClaimsRegistry,
  saveClaimsRegistry,
  generateClaimToken,
  getPaths: () => ({
    claimsPath: state.claimsPath,
    auditPath: state.auditPath,
    legacyClaimsPath: state.legacyClaimsPath,
    legacyAuditPath: state.legacyAuditPath,
  }),
};
