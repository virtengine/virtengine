/**
 * logger.mjs — Centralized logging for codex-monitor.
 *
 * Provides leveled logging that separates human-facing CLI output from
 * debug/trace noise. All messages are always written to the log file;
 * only messages at or above the configured console level appear in the terminal.
 *
 * Levels (lowest to highest):
 *   TRACE → DEBUG → INFO → WARN → ERROR → SILENT
 *
 * Usage:
 *   import { createLogger, setLogLevel, setLogFile } from "./lib/logger.mjs";
 *
 *   const log = createLogger("monitor");
 *   log.info("Task completed");     // Shown in terminal (default)
 *   log.debug("Cache hit ratio: 0.95"); // Hidden in terminal, written to log file
 *   log.trace("Processing line 42");    // Only in log file at TRACE level
 *   log.error("Fatal: no config");      // Always shown
 *
 * The module prefix (e.g. [monitor]) is automatically prepended.
 */

import { appendFileSync, mkdirSync } from "node:fs";
import { dirname } from "node:path";
import { stripAnsi } from "../utils.mjs";

// ── Log levels ──────────────────────────────────────────────────────────────

export const LogLevel = /** @type {const} */ ({
  TRACE: 0,
  DEBUG: 1,
  INFO: 2,
  WARN: 3,
  ERROR: 4,
  SILENT: 5,
});

/** @type {Record<string, number>} */
const LEVEL_MAP = {
  trace: LogLevel.TRACE,
  debug: LogLevel.DEBUG,
  info: LogLevel.INFO,
  warn: LogLevel.WARN,
  error: LogLevel.ERROR,
  silent: LogLevel.SILENT,
};

// ── Global state ────────────────────────────────────────────────────────────

/** @type {number} Console output threshold — messages below this level are suppressed in terminal */
let consoleLevel = LogLevel.INFO;

/** @type {number} File output threshold — messages below this level are not written to log file */
let fileLevel = LogLevel.DEBUG;

/** @type {string|null} Path to the log file */
let logFilePath = null;

/** @type {boolean} Whether the log file directory has been ensured */
let logDirEnsured = false;

/** @type {Set<string>} Modules to always show at DEBUG level even when console is at INFO */
const verboseModules = new Set();

// ── Configuration ───────────────────────────────────────────────────────────

/**
 * Set the minimum console log level.
 * @param {string|number} level - Level name ("trace", "debug", "info", "warn", "error", "silent") or LogLevel value
 */
export function setConsoleLevel(level) {
  if (typeof level === "string") {
    consoleLevel = LEVEL_MAP[level.toLowerCase()] ?? LogLevel.INFO;
  } else {
    consoleLevel = level;
  }
}

/**
 * Set the minimum file log level.
 * @param {string|number} level
 */
export function setFileLevel(level) {
  if (typeof level === "string") {
    fileLevel = LEVEL_MAP[level.toLowerCase()] ?? LogLevel.DEBUG;
  } else {
    fileLevel = level;
  }
}

/**
 * Set the log file path. All messages at or above the file level are appended here.
 * @param {string|null} path
 */
export function setLogFile(path) {
  logFilePath = path;
  logDirEnsured = false;
}

/**
 * Parse CLI args for logging flags and configure accordingly.
 * Call this once at startup before any logging.
 *
 * Flags:
 *   --quiet        Only show errors and warnings
 *   --verbose      Show debug messages too
 *   --trace        Show everything
 *   --silent       No console output
 *   --log-level X  Set explicit level
 *
 * @param {string[]} argv
 */
export function configureFromArgs(argv) {
  if (argv.includes("--silent")) {
    setConsoleLevel(LogLevel.SILENT);
  } else if (argv.includes("--quiet") || argv.includes("-q")) {
    setConsoleLevel(LogLevel.WARN);
  } else if (argv.includes("--trace")) {
    setConsoleLevel(LogLevel.TRACE);
  } else if (argv.includes("--verbose") || argv.includes("-V")) {
    setConsoleLevel(LogLevel.DEBUG);
  }

  const levelIdx = argv.indexOf("--log-level");
  if (levelIdx !== -1 && argv[levelIdx + 1]) {
    setConsoleLevel(argv[levelIdx + 1]);
  }
}

/**
 * Enable verbose (DEBUG-level) output for specific modules even at INFO level.
 * @param {string[]} modules
 */
export function setVerboseModules(modules) {
  verboseModules.clear();
  for (const m of modules) verboseModules.add(m.toLowerCase());
}

/**
 * Get the current console log level.
 * @returns {number}
 */
export function getConsoleLevel() {
  return consoleLevel;
}

// ── Timestamp ───────────────────────────────────────────────────────────────

function timestamp() {
  const d = new Date();
  const hh = String(d.getHours()).padStart(2, "0");
  const mm = String(d.getMinutes()).padStart(2, "0");
  const ss = String(d.getSeconds()).padStart(2, "0");
  return `${hh}:${mm}:${ss}`;
}

function datestamp() {
  return new Date().toISOString();
}

// ── File writing ────────────────────────────────────────────────────────────

function writeToFile(levelName, module, msg) {
  if (!logFilePath) return;
  if (!logDirEnsured) {
    try {
      mkdirSync(dirname(logFilePath), { recursive: true });
    } catch {
      /* best effort */
    }
    logDirEnsured = true;
  }
  const clean = typeof msg === "string" ? stripAnsi(msg) : String(msg);
  const line = `${datestamp()} [${levelName}] [${module}] ${clean}\n`;
  try {
    appendFileSync(logFilePath, line);
  } catch {
    /* best effort */
  }
}

// ── Logger factory ──────────────────────────────────────────────────────────

/**
 * @typedef {Object} Logger
 * @property {(...args: any[]) => void} error - Always shown in terminal
 * @property {(...args: any[]) => void} warn  - Shown at WARN+ level
 * @property {(...args: any[]) => void} info  - Shown at INFO+ level (default)
 * @property {(...args: any[]) => void} debug - Shown at DEBUG+ level or via --verbose
 * @property {(...args: any[]) => void} trace - Shown at TRACE+ level or via --trace
 */

/**
 * Create a logger for a specific module.
 *
 * @param {string} module - Module name (e.g. "monitor", "fleet", "telegram-bot")
 * @returns {Logger}
 */
export function createLogger(module) {
  const prefix = `[${module}]`;
  const moduleLower = module.toLowerCase();

  function emit(level, levelName, consoleFn, args) {
    const msg = args
      .map((a) => (typeof a === "string" ? a : String(a)))
      .join(" ");

    // Always write to file if above file threshold
    if (level >= fileLevel) {
      writeToFile(levelName, module, msg);
    }

    // Console output: check level threshold
    // Module-specific verbose override: show DEBUG for this module even at INFO level
    const effectiveLevel =
      level === LogLevel.DEBUG && verboseModules.has(moduleLower)
        ? LogLevel.INFO
        : level;

    if (effectiveLevel >= consoleLevel) {
      const ts = timestamp();
      consoleFn(`  ${ts} ${prefix} ${msg}`);
    }
  }

  return {
    error: (...args) => emit(LogLevel.ERROR, "ERROR", console.error, args),
    warn: (...args) => emit(LogLevel.WARN, "WARN", console.warn, args),
    info: (...args) => emit(LogLevel.INFO, "INFO", console.log, args),
    debug: (...args) => emit(LogLevel.DEBUG, "DEBUG", console.log, args),
    trace: (...args) => emit(LogLevel.TRACE, "TRACE", console.log, args),
  };
}

// ── Console interceptor ─────────────────────────────────────────────────────
//
// Intercepts console.log / console.warn / console.error globally and routes
// messages through the leveled logger.  This lets us filter 700+ existing
// console.* calls without touching every call-site.
//
// Classification rules:
//   1. Messages starting with `[tag] ` are auto-classified by tag + keywords.
//   2. Messages without a tag prefix pass through as INFO (human-facing output).
//   3. console.warn → WARN, console.error → ERROR (always pass through).
//

/** @type {boolean} */
let interceptorInstalled = false;

// Keywords that promote a tagged message to INFO — MUST be narrow.
// Only truly human-critical events that require operator attention.
const INFO_KEYWORDS = [
  "fatal",
  "crash",
  "circuit breaker",
  "self-restart",
  "shutting down",
  "all tasks complete",
  "stuck",
  "preflight failed",
  "manual resolution",
  "backlog empty",
  "permanently failed",
  "retries exhausted",
];

// Keywords that push a tagged message down to TRACE (very noisy internals)
const TRACE_KEYWORDS = [
  "skipping",
  "dedup",
  "rate limit",
  "throttl",
  "already in progress",
  "no change",
  "unchanged",
  "cache miss",
  "cache hit",
  "nothing to",
  "same as last",
  "too soon",
  "polling",
  "heartbeat",
  "ping",
  "byte",
  "chunk",
  "offset",
  "cursor",
  "checking port",
  "line count",
  "no old completed",
  "cooldown",
  "score:",
  "attempt ",
];

// Modules whose tagged messages default to DEBUG unless INFO_KEYWORDS match
const DEBUG_MODULES = new Set([
  "monitor",
  "fleet",
  "telegram-bot",
  "telegram",
  "vk",
  "vk-log",
  "vk-dispatch",
  "workspace-monitor",
  "maintenance",
  "autofix",
  "config",
  "conflict",
  "merge-strategy",
  "task-complexity",
  "shared-knowledge",
  "preflight",
  "codex-config",
  "task-archiver",
  "update-check",
  "restart-controller",
  "sdk-conflict",
  "fleet-coordinator",
  "conflict-resolver",
  "primary-agent",
  "task-assessment",
  "setup",
  "analyze",
]);

// Pattern: [tag] message  or  [tag]  message (extra space)
const TAG_RE = /^\[([^\]]+)\]\s*/;

/**
 * Classify a console.log message into a LogLevel.
 * @param {string} msg - The first argument to console.log
 * @returns {number} LogLevel
 */
function classifyMessage(msg) {
  if (typeof msg !== "string") return LogLevel.INFO;

  const tagMatch = msg.match(TAG_RE);
  if (!tagMatch) {
    // No [tag] prefix — human-facing output, always INFO
    return LogLevel.INFO;
  }

  const tag = tagMatch[1].toLowerCase();
  const body = msg.slice(tagMatch[0].length).toLowerCase();

  // Check for TRACE keywords first (most restrictive)
  for (const kw of TRACE_KEYWORDS) {
    if (body.includes(kw)) return LogLevel.TRACE;
  }

  // Check for INFO keywords (important events worth showing)
  for (const kw of INFO_KEYWORDS) {
    if (body.includes(kw)) return LogLevel.INFO;
  }

  // If the module is in our DEBUG set, default to DEBUG
  if (DEBUG_MODULES.has(tag)) return LogLevel.DEBUG;

  // Unknown tags → INFO (assume user-facing)
  return LogLevel.INFO;
}

/**
 * Install the global console interceptor.
 * Call once at startup before any significant logging.
 *
 * This replaces console.log / console.warn / console.error with filtered
 * versions that respect the configured console level.
 *
 * - console.error → always ERROR level (passes through)
 * - console.warn  → always WARN level
 * - console.log   → auto-classified by tag/content
 *
 * @param {Object} [opts]
 * @param {string} [opts.logFile] - Path to the log file
 */
export function installConsoleInterceptor(opts = {}) {
  if (interceptorInstalled) return;
  interceptorInstalled = true;

  if (opts.logFile) setLogFile(opts.logFile);

  const _origLog = console.log.bind(console);
  const _origWarn = console.warn.bind(console);
  const _origError = console.error.bind(console);

  // Replace console.log with filtered version
  console.log = (...args) => {
    const first = args[0];
    const level = classifyMessage(first);

    // Always write to file
    if (logFilePath && level >= fileLevel) {
      const msg = args
        .map((a) => (typeof a === "string" ? a : String(a)))
        .join(" ");
      const tagMatch = typeof first === "string" ? first.match(TAG_RE) : null;
      const mod = tagMatch ? tagMatch[1] : "stdout";
      const levelName =
        Object.keys(LogLevel).find((k) => LogLevel[k] === level) || "INFO";
      writeToFile(levelName, mod, msg);
    }

    // Console output
    if (level >= consoleLevel) {
      _origLog(...args);
    }
  };

  // console.warn → always WARN level
  console.warn = (...args) => {
    if (logFilePath && LogLevel.WARN >= fileLevel) {
      const msg = args
        .map((a) => (typeof a === "string" ? a : String(a)))
        .join(" ");
      const tagMatch = typeof msg === "string" ? msg.match(TAG_RE) : null;
      writeToFile("WARN", tagMatch?.[1] || "stderr", msg);
    }
    if (LogLevel.WARN >= consoleLevel) {
      _origWarn(...args);
    }
  };

  // console.error → always ERROR level (always passes through)
  console.error = (...args) => {
    if (logFilePath && LogLevel.ERROR >= fileLevel) {
      const msg = args
        .map((a) => (typeof a === "string" ? a : String(a)))
        .join(" ");
      const tagMatch = typeof msg === "string" ? msg.match(TAG_RE) : null;
      writeToFile("ERROR", tagMatch?.[1] || "stderr", msg);
    }
    // Errors always pass through
    _origError(...args);
  };
}

/**
 * Restore original console methods (for testing).
 */
export function uninstallConsoleInterceptor() {
  // Can't easily restore — this is a best-effort for tests
  interceptorInstalled = false;
}

export default createLogger;
