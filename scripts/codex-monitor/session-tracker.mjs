/**
 * session-tracker.mjs — Captures the last N agent messages for review handoff.
 *
 * When an agent completes (DONE/idle), the session tracker provides the last 10
 * messages as context for the reviewer agent, including both agent outputs and
 * tool calls/results.
 *
 * @module session-tracker
 */

const TAG = "[session-tracker]";

/** Default: keep last 10 messages per session. */
const DEFAULT_MAX_MESSAGES = 10;

/** Maximum characters per message entry to prevent memory bloat. */
const MAX_MESSAGE_CHARS = 2000;

/** Maximum total sessions to keep in memory. */
const MAX_SESSIONS = 50;

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

  /**
   * @param {Object} [options]
   * @param {number} [options.maxMessages=10]
   * @param {number} [options.idleThresholdMs=120000]
   */
  constructor(options = {}) {
    this.#maxMessages = options.maxMessages ?? DEFAULT_MAX_MESSAGES;
    this.#idleThresholdMs = options.idleThresholdMs ?? 120_000;
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
    if (!msg) return; // Skip uninteresting events

    // Push to ring buffer (keep only last N)
    session.messages.push(msg);
    if (session.messages.length > this.#maxMessages) {
      session.messages.shift();
    }
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
  if (!_instance) {
    _instance = new SessionTracker(options);
    console.log(`${TAG} initialized (maxMessages=${_instance.getStats ? DEFAULT_MAX_MESSAGES : "?"})`);
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
