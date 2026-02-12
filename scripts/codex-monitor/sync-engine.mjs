/**
 * sync-engine.mjs — Two-way sync between internal task store and external kanban backends
 *
 * The internal task store (.cache/kanban-state.json via task-store.mjs) is the
 * **source of truth** for status and agent tracking.  External kanbans (VK,
 * GitHub Issues, Jira) are kept in sync:
 *
 *   - Pull: new tasks added externally flow INTO the internal store.
 *   - Push: status changes from the orchestrator flow OUT to the external kanban.
 *
 * EXPORTS:
 *   SyncEngine           — Main class
 *   createSyncEngine()   — Factory helper
 */

import {
  getTask,
  getAllTasks,
  addTask,
  updateTask,
  getDirtyTasks,
  markSynced,
  upsertFromExternal,
  setTaskStatus,
  removeTask,
} from "./task-store.mjs";

import {
  getKanbanAdapter,
  listTasks,
  updateTaskStatus as updateExternalStatus,
} from "./kanban-adapter.mjs";

const TAG = "[sync-engine]";

// ---------------------------------------------------------------------------
// Status ordering — higher = more advanced
// ---------------------------------------------------------------------------

const STATUS_ORDER = {
  todo: 0,
  blocked: 1,
  inprogress: 1,
  inreview: 2,
  done: 3,
  cancelled: 3,
};

function statusRank(s) {
  return STATUS_ORDER[s] ?? -1;
}

// ---------------------------------------------------------------------------
// SyncResult helper
// ---------------------------------------------------------------------------

function emptySyncResult() {
  return {
    pulled: 0,
    pushed: 0,
    conflicts: 0,
    errors: [],
    timestamp: new Date().toISOString(),
  };
}

function mergeSyncResults(a, b) {
  return {
    pulled: a.pulled + b.pulled,
    pushed: a.pushed + b.pushed,
    conflicts: a.conflicts + b.conflicts,
    errors: [...a.errors, ...b.errors],
    timestamp: new Date().toISOString(),
  };
}

// ---------------------------------------------------------------------------
// SyncEngine
// ---------------------------------------------------------------------------

export class SyncEngine {
  /** @type {string} */
  #projectId;
  /** @type {number} */
  #syncIntervalMs;
  /** @type {object|null} */
  #kanbanAdapter;
  /** @type {Function|null} */
  #sendTelegram;

  /** @type {ReturnType<typeof setInterval>|null} */
  #timer = null;
  /** @type {boolean} */
  #running = false;

  // Stats
  #lastSync = null;
  #nextSync = null;
  #syncsCompleted = 0;
  #consecutiveFailures = 0;
  #errors = [];

  // Back-off
  #baseIntervalMs;
  #backoffActive = false;
  static BACKOFF_INTERVAL_MS = 5 * 60 * 1000; // 5 minutes
  static BACKOFF_THRESHOLD = 5;
  static RATE_LIMIT_DELAY_MS = 60 * 1000; // 60 seconds

  /**
   * @param {object} options
   * @param {string}   options.projectId        VK / GitHub project ID
   * @param {number}   [options.syncIntervalMs]  Sync period (default 60 000)
   * @param {object}   [options.kanbanAdapter]   Override adapter from kanban-adapter.mjs
   * @param {Function} [options.sendTelegram]    Optional notification callback
   */
  constructor(options = {}) {
    if (!options.projectId) {
      throw new Error(`${TAG} projectId is required`);
    }
    this.#projectId = options.projectId;
    this.#syncIntervalMs = options.syncIntervalMs ?? 60_000;
    this.#baseIntervalMs = this.#syncIntervalMs;
    this.#kanbanAdapter = options.kanbanAdapter ?? null;
    this.#sendTelegram = options.sendTelegram ?? null;
  }

  // -----------------------------------------------------------------------
  // Lifecycle
  // -----------------------------------------------------------------------

  /** Start periodic sync. */
  start() {
    if (this.#running) {
      console.log(TAG, "Already running — skipping start()");
      return;
    }
    this.#running = true;
    console.log(
      TAG,
      `Starting periodic sync every ${this.#syncIntervalMs}ms for project ${this.#projectId}`,
    );
    this.#scheduleNext();
  }

  /** Stop periodic sync. */
  stop() {
    this.#running = false;
    if (this.#timer) {
      clearTimeout(this.#timer);
      this.#timer = null;
    }
    this.#nextSync = null;
    console.log(TAG, "Stopped periodic sync");
  }

  // -----------------------------------------------------------------------
  // Pull: External → Internal
  // -----------------------------------------------------------------------

  /**
   * Fetch tasks from the external kanban and reconcile into the internal store.
   * @returns {Promise<SyncResult>}
   */
  async pullFromExternal() {
    const result = emptySyncResult();

    /** @type {Array} */
    let externalTasks;
    try {
      externalTasks = await this.#listExternal();
    } catch (err) {
      const msg = `Pull failed — could not list external tasks: ${err.message}`;
      console.warn(TAG, msg);
      result.errors.push(msg);
      return result;
    }

    const internalTasks = getAllTasks();
    const internalById = new Map(internalTasks.map((t) => [t.id, t]));
    const externalIds = new Set();

    for (const ext of externalTasks) {
      if (!ext || !ext.id) continue;
      externalIds.add(ext.id);

      try {
        const internal = internalById.get(ext.id);

        if (!internal) {
          // ── New task from external ──
          upsertFromExternal({
            ...ext,
            projectId: ext.projectId ?? this.#projectId,
            externalBackend: ext.backend ?? null,
          });
          result.pulled++;
          console.log(TAG, `Pulled new task ${ext.id}: ${ext.title}`);
          continue;
        }

        // ── Existing task — check for external status change ──
        const oldExternal = internal.externalStatus;
        const newExternal = ext.status;

        if (newExternal && newExternal !== oldExternal) {
          // External status changed (human edited it)
          const internalRank = statusRank(internal.status);
          const newExternalRank = statusRank(newExternal);
          const oldExternalRank = statusRank(oldExternal);

          if (newExternalRank < oldExternalRank) {
            // External moved BACKWARD — but only accept as human override
            // if internal is ALSO behind (i.e., orchestrator didn't advance it).
            // When internal is at or ahead of the old external, the backward
            // move is most likely a VK restart / state loss — re-push instead.
            if (internalRank >= oldExternalRank) {
              // Internal was actively progressed by orchestrator — ignore
              // the stale external value and mark dirty for re-push.
              updateTask(ext.id, { externalStatus: newExternal, syncDirty: true });
              console.log(
                TAG,
                `External reverted ${ext.id}: ${oldExternal} → ${newExternal} but internal=${internal.status} is ahead — will re-push`,
              );
            } else {
              // Internal is truly behind old-external, so external moving
              // backward is a genuine human override — accept it.
              setTaskStatus(ext.id, newExternal, "external");
              updateTask(ext.id, {
                externalStatus: newExternal,
                syncDirty: false,
              });
              result.pulled++;
              console.log(
                TAG,
                `External moved backward ${ext.id}: ${oldExternal} → ${newExternal} (human override)`,
              );
            }
          } else if (newExternalRank > internalRank) {
            // External moved FORWARD past internal → respect external
            setTaskStatus(ext.id, newExternal, "external");
            updateTask(ext.id, {
              externalStatus: newExternal,
              syncDirty: false,
            });
            result.pulled++;
            console.log(
              TAG,
              `External advanced past internal ${ext.id}: ${internal.status} → ${newExternal}`,
            );
          } else if (internalRank > newExternalRank) {
            // Internal is more advanced → skip, push will handle it
            updateTask(ext.id, { externalStatus: newExternal });
            console.log(
              TAG,
              `Internal ahead of external for ${ext.id}: internal=${internal.status} external=${newExternal} — skipping pull`,
            );
          } else {
            // Same rank, different status (e.g., blocked vs inprogress) or equal
            upsertFromExternal({
              ...ext,
              projectId: ext.projectId ?? this.#projectId,
              externalBackend: ext.backend ?? null,
            });
            result.pulled++;
          }
        } else {
          // No status change — still update metadata (title, description, etc.)
          upsertFromExternal({
            ...ext,
            projectId: ext.projectId ?? this.#projectId,
            externalBackend: ext.backend ?? null,
          });
        }
      } catch (err) {
        const msg = `Pull error for task ${ext.id}: ${err.message}`;
        console.warn(TAG, msg);
        result.errors.push(msg);
      }
    }

    // ── Tasks deleted externally ──
    for (const internal of internalTasks) {
      if (
        internal.projectId === this.#projectId &&
        !externalIds.has(internal.id) &&
        internal.status !== "cancelled" &&
        internal.status !== "done"
      ) {
        try {
          setTaskStatus(internal.id, "cancelled", "external");
          updateTask(internal.id, {
            externalStatus: "cancelled",
            syncDirty: false,
            blockedReason: "Deleted from external kanban",
          });
          console.log(
            TAG,
            `Task ${internal.id} deleted externally — marked cancelled`,
          );
        } catch (err) {
          const msg = `Failed to cancel externally-deleted task ${internal.id}: ${err.message}`;
          console.warn(TAG, msg);
          result.errors.push(msg);
        }
      }
    }

    return result;
  }

  // -----------------------------------------------------------------------
  // Push: Internal → External
  // -----------------------------------------------------------------------

  /**
   * Push dirty internal tasks to the external kanban.
   * @returns {Promise<SyncResult>}
   */
  async pushToExternal() {
    const result = emptySyncResult();
    const dirtyTasks = getDirtyTasks();

    if (dirtyTasks.length === 0) {
      return result;
    }

    console.log(TAG, `Pushing ${dirtyTasks.length} dirty task(s) to external`);

    for (const task of dirtyTasks) {
      try {
        await this.#updateExternal(task.id, task.status);
        markSynced(task.id);
        result.pushed++;
        console.log(TAG, `Pushed ${task.id} → ${task.status}`);
      } catch (err) {
        if (this.#isRateLimited(err)) {
          const msg = `Rate limited — backing off for ${SyncEngine.RATE_LIMIT_DELAY_MS / 1000}s`;
          console.warn(TAG, msg);
          result.errors.push(msg);
          await this.#sleep(SyncEngine.RATE_LIMIT_DELAY_MS);
          // Retry once after back-off
          try {
            await this.#updateExternal(task.id, task.status);
            markSynced(task.id);
            result.pushed++;
            console.log(
              TAG,
              `Pushed ${task.id} → ${task.status} (after rate-limit retry)`,
            );
          } catch (retryErr) {
            const retryMsg = `Push retry failed for ${task.id}: ${retryErr.message}`;
            console.warn(TAG, retryMsg);
            result.errors.push(retryMsg);
          }
        } else if (this.#isNotFound(err)) {
          // Task was deleted from the external kanban — stop retrying
          console.warn(
            TAG,
            `Task ${task.id} returned 404 — removing orphaned task from internal store`,
          );
          removeTask(task.id);
        } else {
          const msg = `Push failed for ${task.id}: ${err.message}`;
          console.warn(TAG, msg);
          result.errors.push(msg);
        }
      }
    }

    return result;
  }

  // -----------------------------------------------------------------------
  // Full sync
  // -----------------------------------------------------------------------

  /**
   * Run a complete pull + push cycle.
   * @returns {Promise<SyncResult>}
   */
  async fullSync() {
    console.log(TAG, "Starting full sync…");
    const t0 = Date.now();

    let pullResult = emptySyncResult();
    let pushResult = emptySyncResult();

    try {
      pullResult = await this.pullFromExternal();
    } catch (err) {
      pullResult.errors.push(`Pull phase crashed: ${err.message}`);
      console.warn(TAG, `Pull phase error: ${err.message}`);
    }

    try {
      pushResult = await this.pushToExternal();
    } catch (err) {
      pushResult.errors.push(`Push phase crashed: ${err.message}`);
      console.warn(TAG, `Push phase error: ${err.message}`);
    }

    const combined = mergeSyncResults(pullResult, pushResult);
    const elapsed = Date.now() - t0;

    // Track consecutive failures for back-off
    if (combined.errors.length > 0) {
      this.#consecutiveFailures++;
      this.#errors = combined.errors.slice(-20); // keep last 20

      if (
        this.#consecutiveFailures >= SyncEngine.BACKOFF_THRESHOLD &&
        !this.#backoffActive
      ) {
        this.#backoffActive = true;
        this.#syncIntervalMs = SyncEngine.BACKOFF_INTERVAL_MS;
        console.warn(
          TAG,
          `${this.#consecutiveFailures} consecutive failures — slowing sync to ${this.#syncIntervalMs}ms`,
        );
        if (this.#sendTelegram) {
          this.#sendTelegram(
            `⚠️ Sync engine: ${this.#consecutiveFailures} consecutive failures, backing off to 5 min interval`,
          ).catch(() => {});
        }
      }
    } else {
      // Successful sync — reset failures and restore normal interval
      if (this.#consecutiveFailures > 0) {
        console.log(
          TAG,
          `Sync succeeded after ${this.#consecutiveFailures} failure(s) — resetting`,
        );
      }
      this.#consecutiveFailures = 0;
      if (this.#backoffActive) {
        this.#backoffActive = false;
        this.#syncIntervalMs = this.#baseIntervalMs;
        console.log(
          TAG,
          `Back-off cleared — restoring interval to ${this.#syncIntervalMs}ms`,
        );
      }
    }

    this.#syncsCompleted++;
    this.#lastSync = combined.timestamp;

    console.log(
      TAG,
      `Full sync complete in ${elapsed}ms — pulled=${combined.pulled} pushed=${combined.pushed} conflicts=${combined.conflicts} errors=${combined.errors.length}`,
    );

    return combined;
  }

  // -----------------------------------------------------------------------
  // Single-task sync
  // -----------------------------------------------------------------------

  /**
   * Force-sync a specific task: push internal state to external.
   * @param {string} taskId
   */
  async syncTask(taskId) {
    const task = getTask(taskId);
    if (!task) {
      console.warn(TAG, `syncTask: task ${taskId} not found in internal store`);
      return;
    }

    try {
      await this.#updateExternal(taskId, task.status);
      markSynced(taskId);
      console.log(TAG, `Force-synced task ${taskId} → ${task.status}`);
    } catch (err) {
      console.warn(TAG, `syncTask failed for ${taskId}: ${err.message}`);
    }
  }

  // -----------------------------------------------------------------------
  // Status
  // -----------------------------------------------------------------------

  /**
   * Return current sync engine status.
   */
  getStatus() {
    return {
      lastSync: this.#lastSync,
      nextSync: this.#nextSync,
      syncsCompleted: this.#syncsCompleted,
      errors: [...this.#errors],
      running: this.#running,
      consecutiveFailures: this.#consecutiveFailures,
      backoffActive: this.#backoffActive,
      currentIntervalMs: this.#syncIntervalMs,
    };
  }

  // -----------------------------------------------------------------------
  // Private helpers
  // -----------------------------------------------------------------------

  /** Schedule the next sync tick. */
  #scheduleNext() {
    if (!this.#running) return;

    this.#nextSync = new Date(Date.now() + this.#syncIntervalMs).toISOString();
    this.#timer = setTimeout(async () => {
      try {
        await this.fullSync();
      } catch (err) {
        console.warn(TAG, `Periodic sync error: ${err.message}`);
      }
      this.#scheduleNext();
    }, this.#syncIntervalMs);

    // Prevent timer from blocking process exit
    if (this.#timer && typeof this.#timer.unref === "function") {
      this.#timer.unref();
    }
  }

  /**
   * List tasks from the external kanban, using adapter override or the
   * module-level listTasks convenience function.
   */
  async #listExternal() {
    if (
      this.#kanbanAdapter &&
      typeof this.#kanbanAdapter.listTasks === "function"
    ) {
      return await this.#kanbanAdapter.listTasks(this.#projectId);
    }
    return await listTasks(this.#projectId);
  }

  /**
   * Update a single task's status in the external kanban.
   */
  async #updateExternal(taskId, status) {
    if (
      this.#kanbanAdapter &&
      typeof this.#kanbanAdapter.updateTaskStatus === "function"
    ) {
      return await this.#kanbanAdapter.updateTaskStatus(taskId, status);
    }
    return await updateExternalStatus(taskId, status);
  }

  /**
   * Determine whether an error is a 429 rate-limit response.
   */
  #isRateLimited(err) {
    if (!err) return false;
    const msg = String(err.message || err).toLowerCase();
    return (
      msg.includes("429") ||
      msg.includes("rate limit") ||
      msg.includes("too many requests")
    );
  }

  /**
   * Determine whether an error is a 404 Not Found response.
   */
  #isNotFound(err) {
    if (!err) return false;
    const msg = String(err.message || err).toLowerCase();
    return msg.includes("404") || msg.includes("not found");
  }

  /** Simple async sleep. */
  #sleep(ms) {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

/**
 * Create and return a SyncEngine instance.
 * @param {object} options  Same as SyncEngine constructor
 * @returns {SyncEngine}
 */
export function createSyncEngine(options) {
  return new SyncEngine(options);
}
