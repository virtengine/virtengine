/**
 * Pure utility functions for codex-monitor
 * Extracted for unit testing (no I/O, no side effects)
 */

/**
 * Normalize text for deduplication by stripping timestamps and IDs
 * @param {string} text - Input text to normalize
 * @returns {string} Normalized text with numbers replaced by N
 */
export function normalizeDedupKey(text) {
  return (
    String(text || "")
      .trim()
      // Replace numbers (integers and decimals) with N, preserving surrounding text
      .replaceAll(/\d+(\.\d+)?/g, "N")
      // Collapse any resulting multi-N sequences (e.g., "N.N" → "N")
      .replaceAll(/N[.\-/:]N/g, "N")
      // Collapse whitespace
      .replaceAll(/\s+/g, " ")
  );
}

/**
 * Strip ANSI escape codes from text
 * PowerShell and colored CLI output includes \x1b[...m sequences that show
 * as garbage in Telegram messages.
 * @param {string} text - Input text with potential ANSI codes
 * @returns {string} Clean text without ANSI codes
 */
export function stripAnsi(text) {
  // eslint-disable-next-line no-control-regex
  return String(text || "")
    .replace(/\x1b\[[0-9;]*m/g, "")
    .replace(/\[\d+;?\d*m/g, "");
}

/**
 * Check if a line matches error patterns (excluding noise patterns)
 * @param {string} line - Log line to check
 * @param {RegExp[]} errorPatterns - Patterns that indicate errors
 * @param {RegExp[]} errorNoisePatterns - Patterns to exclude from error detection
 * @returns {boolean} True if line is an error
 */
export function isErrorLine(line, errorPatterns, errorNoisePatterns) {
  if (errorNoisePatterns.some((pattern) => pattern.test(line))) {
    return false;
  }
  return errorPatterns.some((pattern) => pattern.test(line));
}

/**
 * Escape HTML special characters
 * @param {any} value - Value to escape
 * @returns {string} HTML-escaped string
 */
export function escapeHtml(value) {
  return String(value)
    .replace(/&/g, "&amp;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
}

/**
 * Format an HTML link with proper escaping
 * @param {string} url - URL for the link
 * @param {string} label - Display text for the link
 * @returns {string} HTML anchor tag or escaped label if no URL
 */
export function formatHtmlLink(url, label) {
  if (!url) {
    return escapeHtml(label);
  }
  return `<a href="${escapeHtml(url)}">${escapeHtml(label)}</a>`;
}

/**
 * Generate a normalized fingerprint for an error line (for deduplication)
 * Strips timestamps, attempt IDs, and branch-specific parts
 * @param {string} line - Error line to fingerprint
 * @returns {string} Normalized fingerprint
 */
export function getErrorFingerprint(line) {
  // Normalize: strip timestamps, attempt IDs, branch-specific parts
  return line
    .replace(/\[\d{2}:\d{2}:\d{2}\]\s*/g, "")
    .replace(/\b[0-9a-f]{8}\b/gi, "<ID>") // attempt IDs
    .replace(/ve\/[\w.-]+/g, "ve/<BRANCH>") // branch names
    .trim();
}

/**
 * Parse -MaxParallel argument from command line arguments
 * Supports: -MaxParallel N, --maxparallel=N, --max-parallel N
 * Falls back to VK_MAX_PARALLEL or MAX_PARALLEL env vars
 * @param {string[]} argsList - Command line arguments array
 * @returns {number|null} Maximum parallel value or null if not found
 */
export function getMaxParallelFromArgs(argsList) {
  if (!Array.isArray(argsList)) {
    return null;
  }
  for (let i = 0; i < argsList.length; i += 1) {
    const arg = String(argsList[i] ?? "");
    const directMatch =
      arg.match(/^-{1,2}maxparallel(?:=|:)?(\d+)$/i) ||
      arg.match(/^--max-parallel(?:=|:)?(\d+)$/i);
    if (directMatch) {
      const value = Number(directMatch[1]);
      if (Number.isFinite(value) && value > 0) {
        return value;
      }
    }
    const normalized = arg.toLowerCase();
    if (
      normalized === "-maxparallel" ||
      normalized === "--maxparallel" ||
      normalized === "--max-parallel"
    ) {
      const next = Number(argsList[i + 1]);
      if (Number.isFinite(next) && next > 0) {
        return next;
      }
    }
  }
  const envValue = Number(
    process.env.VK_MAX_PARALLEL || process.env.MAX_PARALLEL,
  );
  if (Number.isFinite(envValue) && envValue > 0) {
    return envValue;
  }
  return null;
}

/**
 * Extract PR number from a GitHub pull request URL
 * @param {string} url - GitHub PR URL
 * @returns {number|null} PR number or null if not found
 */
export function parsePrNumberFromUrl(url) {
  if (!url) return null;
  const match = String(url).match(/\/pull\/(\d+)/i);
  if (!match) return null;
  const num = Number(match[1]);
  return Number.isFinite(num) ? num : null;
}
