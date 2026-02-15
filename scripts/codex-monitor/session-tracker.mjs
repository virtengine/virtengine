/**
 * session-tracker.mjs — Captures the last N agent messages for review handoff.
 *
 * When an agent completes (DONE/idle), the session tracker provides the last 10
 * messages as context for the reviewer agent, including both agent outputs and
 * tool calls/results.
 *
 * @module session-tracker
 */

import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { resolveRepoSharedStatePaths } from "./shared-state-paths.mjs";

const TAG = "[session-tracker]";
const SESSION_TRACKER_FILENAME = "session-tracker.json";

/** Default: keep last 10 messages per session. */
const DEFAULT_MAX_MESSAGES = 10;

/** Maximum characters per message entry to prevent memory bloat. */
const MAX_MESSAGE_CHARS = 2000;

/** Maximum total sessions to keep in memory. */
const MAX_SESSIONS = 50;

function sanitizeSessionRecord(raw, maxMessages) {
  if (!raw || typeof raw !== "object") return null;
  const taskId = String(raw.taskId || "").trim();
  if (!taskId) return null;
  const messages = Array.isArray(raw.messages) ? raw.messages : [];
  const normalizedMessages = messages
    .filter((entry) => entry && typeof entry === "object")
    .map((entry) => ({
      type: String(entry.type || "system"),
      content: String(entry.content || "").slice(0, MAX_MESSAGE_CHARS),
      timestamp: String(entry.timestamp || new Date().toISOString()),
      meta: entry.meta && typeof entry.meta === "object" ? entry.meta : undefined,
    }))
    .slice(-Math.max(1, maxMessages));

  return {
    taskId,
    taskTitle: String(raw.taskTitle || taskId),
    startedAt: Number(raw.startedAt || Date.now()),
    endedAt: raw.endedAt == null ? null : Number(raw.endedAt),
    messages: normalizedMessages,
    totalEvents: Number(raw.totalEvents || normalizedMessages.length || 0),
    status: String(raw.status || "active"),
    lastActivityAt: Number(raw.lastActivityAt || Date.now()),
  };
}

function resolvePersistenceConfig(options = {}) {
  const shared = resolveRepoSharedStatePaths({
    repoRoot: options.repoRoot,
    cwd: options.cwd,
    stateDir: options.stateDir,
    stateRoot: options.stateRoot,
    repoIdentity: options.repoIdentity,
  });
  const storagePath = options.sessionPath
    ? resolve(options.sessionPath)
    : shared.file(SESSION_TRACKER_FILENAME);
  const legacyPaths = Array.isArray(options.legacyPaths)
    ? options.legacyPaths.map((path) => resolve(path))
    : [
        resolve(shared.legacyCacheDir, SESSION_TRACKER_FILENAME),
        resolve(shared.legacyCodexCacheDir, SESSION_TRACKER_FILENAME),
      ];
  return { storagePath, legacyPaths };
}

// ── Message Types ───────────────────────────────────────────────────────────

/**
 * @typedef {Object} SessionMessage
 * @property {string} type        - "agent_message"|"tool_call"|"tool_result"|"error"|"system"
 * @property {string} content     - Truncated content
 * @property {string} timestamp   - ISO timestamp
 * @property {Object} [meta]      - Optional metadata (tool name, etc.)
 */

/**
 * @typedef {Object} SessionRecord
 * @property {string} taskId
 * @property {string} taskTitle
 * @property {number} startedAt
 * @property {number|null} endedAt
 * @property {SessionMessage[]} messages
 * @property {number} totalEvents     - Total events received (before truncation)
 * @property {string} status          - "active"|"completed"|"idle"|"failed"
 * @property {number} lastActivityAt  - Timestamp of last event
 */

// ── SessionTracker Class ────────────────────────────────────────────────────

export class SessionTracker {
  /** @type {Map<string, SessionRecord>} taskId → session record */
  #sessions = new Map();

  /** @type {number} */
  #maxMessages;

  /** @type {number} idle threshold (ms) — 2 minutes without events = idle */
  #idleThresholdMs;

  /** @type {boolean} */
  #persist;

  /** @type {string|null} */
  #storagePath;

  /** @type {string[]} */
  #legacyPaths;

  /** @type {boolean} */
  #loaded = false;

  /**
   * @param {Object} [options]
   * @param {number} [options.maxMessages=10]
   * @param {number} [options.idleThresholdMs=120000]
   */
  constructor(options = {}) {
    this.#maxMessages = options.maxMessages ?? DEFAULT_MAX_MESSAGES;
    this.#idleThresholdMs = options.idleThresholdMs ?? 180_000; // 3 minutes — gives agents breathing room
    const persistDefault = Boolean(options.sessionPath || options.repoRoot);
    this.#persist = options.persist ?? persistDefault;

    if (this.#persist) {
      const persistence = resolvePersistenceConfig(options);
      this.#storagePath = persistence.storagePath;
      this.#legacyPaths = persistence.legacyPaths;
      this.#loadFromDisk();
    } else {
      this.#storagePath = null;
      this.#legacyPaths = [];
    }
  }

  #loadFromDisk() {
    if (!this.#persist || this.#loaded) return;
    this.#loaded = true;

    const candidatePaths = [this.#storagePath, ...this.#legacyPaths].filter(Boolean);
    for (const path of candidatePaths) {
      if (!existsSync(path)) continue;
      try {
        const raw = JSON.parse(readFileSync(path, "utf8"));
        const items = Array.isArray(raw?.sessions) ? raw.sessions : [];
        for (const item of items) {
          const record = sanitizeSessionRecord(item, this.#maxMessages);
          if (record) this.#sessions.set(record.taskId, record);
        }
        if (path !== this.#storagePath && this.#storagePath && !existsSync(this.#storagePath)) {
          this.#persistNow();
        }
        return;
      } catch {
        // continue
      }
    }
  }

  #persistNow() {
    if (!this.#persist || !this.#storagePath) return;
    try {
      mkdirSync(dirname(this.#storagePath), { recursive: true });
      const payload = {
        version: 1,
        updated_at: new Date().toISOString(),
        sessions: [...this.#sessions.values()],
      };
      writeFileSync(this.#storagePath, JSON.stringify(payload, null, 2), "utf8");
    } catch {
      // best-effort persistence only
    }
  }

  /**
   * Start tracking a new session for a task.
   * If a session already exists, it's replaced.
   *
   * @param {string} taskId
   * @param {string} taskTitle
   */
  startSession(taskId, taskTitle) {
    // Evict oldest sessions if at capacity
    if (this.#sessions.size >= MAX_SESSIONS && !this.#sessions.has(taskId)) {
      const oldest = [...this.#sessions.entries()]
        .sort((a, b) => a[1].startedAt - b[1].startedAt)
        .slice(0, Math.ceil(MAX_SESSIONS / 4));
      for (const [id] of oldest) {
        this.#sessions.delete(id);
      }
    }

    this.#sessions.set(taskId, {
      taskId,
      taskTitle,
      startedAt: Date.now(),
      endedAt: null,
      messages: [],
      totalEvents: 0,
      status: "active",
      lastActivityAt: Date.now(),
    });
    this.#persistNow();
  }

  /**
   * Record an agent SDK event for a task session.
   * Call this from the `onEvent` callback inside `execWithRetry`.
   *
   * Normalizes events from all 3 SDKs:
   * - Codex: { type: "item.completed"|"item.created", item: {...} }
   * - Copilot: { type: "message"|"tool_call"|"tool_result", ... }
   * - Claude: { type: "content_block_delta"|"message_stop", ... }
   *
   * @param {string} taskId
   * @param {Object} event - Raw SDK event
   */
  recordEvent(taskId, event) {
    const session = this.#sessions.get(taskId);
    if (!session) return;

    session.totalEvents++;
    session.lastActivityAt = Date.now();

    const msg = this.#normalizeEvent(event);
    if (!msg) {
      this.#persistNow();
      return;
    }

    // Push to ring buffer (keep only last N)
    session.messages.push(msg);
    if (session.messages.length > this.#maxMessages) {
      session.messages.shift();
    }
    this.#persistNow();
  }

  /**
   * Mark a session as completed.
   * @param {string} taskId
   * @param {"completed"|"failed"|"idle"} [status="completed"]
   */
  endSession(taskId, status = "completed") {
    const session = this.#sessions.get(taskId);
    if (!session) return;

    session.endedAt = Date.now();
    session.status = status;
    this.#persistNow();
  }

  /**
   * Get the last N messages for a task session.
   * @param {string} taskId
   * @param {number} [n] - defaults to maxMessages
   * @returns {SessionMessage[]}
   */
  getLastMessages(taskId, n) {
    const session = this.#sessions.get(taskId);
    if (!session) return [];
    const count = n ?? this.#maxMessages;
    return session.messages.slice(-count);
  }

  /**
   * Get a formatted summary of the last N messages.
   * This is the string that gets passed to the review agent.
   *
   * @param {string} taskId
   * @param {number} [n]
   * @returns {string}
   */
  getMessageSummary(taskId, n) {
    const messages = this.getLastMessages(taskId, n);
    if (messages.length === 0) return "(no session messages recorded)";

    const session = this.#sessions.get(taskId);
    const header = [
      `Session: ${session?.taskTitle || taskId}`,
      `Total events: ${session?.totalEvents ?? 0}`,
      `Duration: ${session ? Math.round((Date.now() - session.startedAt) / 1000) : 0}s`,
      `Status: ${session?.status ?? "unknown"}`,
      `--- Last ${messages.length} messages ---`,
    ].join("\n");

    const lines = messages.map((msg) => {
      const ts = new Date(msg.timestamp).toISOString().slice(11, 19);
      const prefix = this.#typePrefix(msg.type);
      const meta = msg.meta?.toolName ? ` [${msg.meta.toolName}]` : "";
      return `[${ts}] ${prefix}${meta}: ${msg.content}`;
    });

    return `${header}\n${lines.join("\n")}`;
  }

  /**
   * Check if a session appears to be idle (no events for > idleThreshold).
   * @param {string} taskId
   * @returns {boolean}
   */
  isSessionIdle(taskId) {
    const session = this.#sessions.get(taskId);
    if (!session || session.status !== "active") return false;
    return Date.now() - session.lastActivityAt > this.#idleThresholdMs;
  }

  /**
   * Get detailed progress status for a running session.
   * Returns a structured assessment of agent progress suitable for mid-execution monitoring.
   *
   * @param {string} taskId
   * @returns {{ status: "active"|"idle"|"stalled"|"not_found"|"ended", idleMs: number, totalEvents: number, lastEventType: string|null, hasEdits: boolean, hasCommits: boolean, elapsedMs: number, recommendation: "none"|"continue"|"nudge"|"abort" }}
   */
  getProgressStatus(taskId) {
    const session = this.#sessions.get(taskId);
    if (!session) {
      return {
        status: "not_found", idleMs: 0, totalEvents: 0,
        lastEventType: null, hasEdits: false, hasCommits: false,
        elapsedMs: 0, recommendation: "none",
      };
    }

    if (session.status !== "active") {
      return {
        status: "ended", idleMs: 0, totalEvents: session.totalEvents,
        lastEventType: session.messages.at(-1)?.type ?? null,
        hasEdits: false, hasCommits: false,
        elapsedMs: (session.endedAt || Date.now()) - session.startedAt,
        recommendation: "none",
      };
    }

    const now = Date.now();
    const idleMs = now - session.lastActivityAt;
    const elapsedMs = now - session.startedAt;

    // Check if agent has done any meaningful edits or commits
    const hasEdits = session.messages.some((m) => {
      if (m.type !== "tool_call") return false;
      const c = (m.content || "").toLowerCase();
      return c.includes("write") || c.includes("edit") || c.includes("create") ||
        c.includes("replace") || c.includes("patch") || c.includes("append");
    });

    const hasCommits = session.messages.some((m) => {
      if (m.type !== "tool_call") return false;
      const c = (m.content || "").toLowerCase();
      return c.includes("git commit") || c.includes("git push");
    });

    // Determine status — check stalled FIRST (it's the stricter condition)
    let status = "active";
    if (idleMs > this.#idleThresholdMs * 2) {
      status = "stalled";
    } else if (idleMs > this.#idleThresholdMs) {
      status = "idle";
    }

    // Determine recommendation
    let recommendation = "none";
    if (status === "stalled") {
      recommendation = "abort";
    } else if (status === "idle") {
      // If agent was idle but had some activity, try CONTINUE
      recommendation = session.totalEvents > 0 ? "continue" : "nudge";
    } else if (elapsedMs > 30 * 60_000 && session.totalEvents < 5) {
      // 30 min with < 5 events — agent is stalled even if not technically idle
      recommendation = "continue";
    }

    return {
      status, idleMs, totalEvents: session.totalEvents,
      lastEventType: session.messages.at(-1)?.type ?? null,
      hasEdits, hasCommits, elapsedMs, recommendation,
    };
  }

  /**
   * Get all active sessions (for watchdog scanning).
   * @returns {Array<{ taskId: string, taskTitle: string, idleMs: number, totalEvents: number, elapsedMs: number }>}
   */
  getActiveSessions() {
    const result = [];
    const now = Date.now();
    for (const [taskId, session] of this.#sessions) {
      if (session.status !== "active") continue;
      result.push({
        taskId,
        taskTitle: session.taskTitle,
        idleMs: now - session.lastActivityAt,
        totalEvents: session.totalEvents,
        elapsedMs: now - session.startedAt,
      });
    }
    return result;
  }

  /**
   * Get the full session record.
   * @param {string} taskId
   * @returns {SessionRecord|null}
   */
  getSession(taskId) {
    return this.#sessions.get(taskId) ?? null;
  }

  /**
   * Remove a session from tracking (after review handoff).
   * @param {string} taskId
   */
  removeSession(taskId) {
    this.#sessions.delete(taskId);
    this.#persistNow();
  }

  /**
   * Get stats about tracked sessions.
   * @returns {{ active: number, completed: number, total: number }}
   */
  getStats() {
    let active = 0;
    let completed = 0;
    for (const session of this.#sessions.values()) {
      if (session.status === "active") active++;
      else completed++;
    }
    return { active, completed, total: this.#sessions.size };
  }

  getStoragePath() {
    return this.#storagePath;
  }

  // ── Private helpers ───────────────────────────────────────────────────────

  /**
   * Normalize a raw SDK event into a SessionMessage.
   * Returns null for events that shouldn't be tracked (noise).
   *
   * @param {Object} event
   * @returns {SessionMessage|null}
   * @private
   */
  #normalizeEvent(event) {
    if (!event || !event.type) return null;

    const ts = new Date().toISOString();

    // ── Codex SDK events ──
    if (event.type === "item.completed" && event.item) {
      const item = event.item;

      if (item.type === "agent_message" && item.text) {
        return {
          type: "agent_message",
          content: item.text.slice(0, MAX_MESSAGE_CHARS),
          timestamp: ts,
        };
      }

      if (item.type === "function_call") {
        return {
          type: "tool_call",
          content: `${item.name}(${(item.arguments || "").slice(0, 500)})`,
          timestamp: ts,
          meta: { toolName: item.name },
        };
      }

      if (item.type === "function_call_output") {
        return {
          type: "tool_result",
          content: (item.output || "").slice(0, MAX_MESSAGE_CHARS),
          timestamp: ts,
        };
      }

      return null; // Skip other item types
    }

    // ── Copilot SDK events ──
    if (event.type === "message" && event.content) {
      return {
        type: "agent_message",
        content: (typeof event.content === "string" ? event.content : JSON.stringify(event.content))
          .slice(0, MAX_MESSAGE_CHARS),
        timestamp: ts,
      };
    }

    if (event.type === "tool_call") {
      return {
        type: "tool_call",
        content: `${event.name || event.tool || "tool"}(${(event.arguments || event.input || "").slice(0, 500)})`,
        timestamp: ts,
        meta: { toolName: event.name || event.tool },
      };
    }

    if (event.type === "tool_result" || event.type === "tool_output") {
      return {
        type: "tool_result",
        content: (event.output || event.result || "").slice(0, MAX_MESSAGE_CHARS),
        timestamp: ts,
      };
    }

    // ── Claude SDK events ──
    if (event.type === "content_block_delta" && event.delta?.text) {
      return {
        type: "agent_message",
        content: event.delta.text.slice(0, MAX_MESSAGE_CHARS),
        timestamp: ts,
      };
    }

    if (event.type === "message_stop" || event.type === "message_delta") {
      return {
        type: "system",
        content: `${event.type}${event.delta?.stop_reason ? ` (${event.delta.stop_reason})` : ""}`,
        timestamp: ts,
      };
    }

    // ── Error events (any SDK) ──
    if (event.type === "error" || event.type === "stream_error") {
      return {
        type: "error",
        content: (event.error?.message || event.message || JSON.stringify(event)).slice(0, MAX_MESSAGE_CHARS),
        timestamp: ts,
      };
    }

    return null;
  }

  /**
   * Get a display prefix for a message type.
   * @param {string} type
   * @returns {string}
   * @private
   */
  #typePrefix(type) {
    switch (type) {
      case "agent_message": return "AGENT";
      case "tool_call":     return "TOOL";
      case "tool_result":   return "RESULT";
      case "error":         return "ERROR";
      case "system":        return "SYS";
      default:              return type.toUpperCase();
    }
  }
}

// ── Singleton ───────────────────────────────────────────────────────────────

/** @type {SessionTracker|null} */
let _instance = null;

/**
 * Get or create the singleton SessionTracker.
 * @param {Object} [options]
 * @returns {SessionTracker}
 */
export function getSessionTracker(options) {
  const opts = options || {};
  const forcePersist = opts.persist ?? true;
  if (!_instance || (opts.repoRoot && !_instance.getStoragePath())) {
    _instance = new SessionTracker({ ...opts, persist: forcePersist });
    console.log(`${TAG} initialized (maxMessages=${DEFAULT_MAX_MESSAGES})`);
  }
  return _instance;
}

/**
 * Create a standalone SessionTracker (for testing).
 * @param {Object} [options]
 * @returns {SessionTracker}
 */
export function createSessionTracker(options) {
  return new SessionTracker(options);
}
