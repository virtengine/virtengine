/**
 * codex-config.mjs — Manages the Codex CLI config (~/.codex/config.toml)
 *
 * Ensures the user's Codex CLI configuration has:
 *   1. A vibe_kanban MCP server section with the correct env vars
 *   2. Sufficient stream_idle_timeout_ms on all model providers
 *   3. Recommended defaults for long-running agentic workloads
 *
 * Uses string-based TOML manipulation (no parser dependency) — we only
 * append or patch well-known sections rather than rewriting the whole file.
 */

import { existsSync, readFileSync, writeFileSync, mkdirSync } from "node:fs";
import { resolve } from "node:path";
import { homedir } from "node:os";

// ── Constants ────────────────────────────────────────────────────────────────

const CODEX_DIR = resolve(homedir(), ".codex");
const CONFIG_PATH = resolve(CODEX_DIR, "config.toml");

/** Minimum recommended stream idle timeout (ms) for complex agentic tasks. */
const MIN_STREAM_IDLE_TIMEOUT_MS = 300_000; // 5 minutes

/** The recommended (generous) timeout for heavy reasoning models. */
const RECOMMENDED_STREAM_IDLE_TIMEOUT_MS = 3_600_000; // 60 minutes

// ── Agent SDK Selection (config.toml) ───────────────────────────────────────

const AGENT_SDK_HEADER = "[agent_sdk]";
const AGENT_SDK_CAPS_HEADER = "[agent_sdk.capabilities]";

const DEFAULT_AGENT_SDK_BLOCK = [
  "",
  "# ── Agent SDK selection (added by codex-monitor) ──",
  AGENT_SDK_HEADER,
  "# Primary agent SDK used for in-process automation.",
  '# Supported: "codex", "copilot", "claude"',
  'primary = "codex"',
  "",
  AGENT_SDK_CAPS_HEADER,
  "# Live steering updates during an active run.",
  "steering = true",
  "# Ability to spawn subagents/child tasks.",
  "subagents = true",
  "# Access to VS Code tools (Copilot extension).",
  "vscode_tools = false",
  "",
].join("\n");

// ── Public API ───────────────────────────────────────────────────────────────

/**
 * Read the current config.toml (or return empty string if it doesn't exist).
 */
export function readCodexConfig() {
  if (!existsSync(CONFIG_PATH)) return "";
  return readFileSync(CONFIG_PATH, "utf8");
}

/**
 * Write the config.toml, creating ~/.codex/ if needed.
 */
export function writeCodexConfig(content) {
  mkdirSync(CODEX_DIR, { recursive: true });
  writeFileSync(CONFIG_PATH, content, "utf8");
}

/**
 * Get the path to the Codex config file.
 */
export function getConfigPath() {
  return CONFIG_PATH;
}

/**
 * Check whether the config already has a [mcp_servers.vibe_kanban] section.
 */
export function hasVibeKanbanMcp(toml) {
  return /^\[mcp_servers\.vibe_kanban\]/m.test(toml);
}

/**
 * Check whether the config already has a [mcp_servers.vibe_kanban.env] section.
 */
export function hasVibeKanbanEnv(toml) {
  return /^\[mcp_servers\.vibe_kanban\.env\]/m.test(toml);
}

/**
 * Check whether the config already has an [agent_sdk] section.
 */
export function hasAgentSdkConfig(toml) {
  return /^\[agent_sdk\]/m.test(toml);
}

/**
 * Build the default agent SDK block.
 */
export function buildAgentSdkBlock() {
  return DEFAULT_AGENT_SDK_BLOCK;
}

/**
 * Build the vibe_kanban MCP server block (including env vars).
 *
 * @param {object} opts
 * @param {string} opts.vkBaseUrl   e.g. "http://127.0.0.1:54089"
 * @param {string} opts.vkPort      e.g. "54089"
 * @param {string} opts.vkHost      e.g. "127.0.0.1"
 */
export function buildVibeKanbanBlock({
  vkBaseUrl = "http://127.0.0.1:54089",
  vkPort = "54089",
  vkHost = "127.0.0.1",
} = {}) {
  return [
    "",
    "# ── Vibe-Kanban MCP (added by codex-monitor) ──",
    "[mcp_servers.vibe_kanban]",
    "args = [",
    '    "-y",',
    '    "vibe-kanban@latest",',
    '    "--mcp",',
    "]",
    'command = "npx"',
    'tools = ["*"]',
    "",
    "[mcp_servers.vibe_kanban.env]",
    "# Ensure MCP always targets the correct VK API endpoint.",
    `VK_BASE_URL = "${vkBaseUrl}"`,
    `VK_ENDPOINT_URL = "${vkBaseUrl}"`,
    `VK_RECOVERY_PORT = "${vkPort}"`,
    "# Also bind the MCP-hosted VK instance to the expected port if it starts one.",
    `PORT = "${vkPort}"`,
    `HOST = "${vkHost}"`,
    "",
  ].join("\n");
}

/**
 * Update the env vars inside an existing [mcp_servers.vibe_kanban.env] section.
 * If a key already exists with a different value, it is replaced.
 * If a key is missing, it is appended to the section.
 *
 * @param {string} toml  Current config.toml content
 * @param {object} envVars  Key-value pairs to ensure
 * @returns {string}  Updated TOML
 */
export function updateVibeKanbanEnv(toml, envVars) {
  const envHeader = "[mcp_servers.vibe_kanban.env]";
  const headerIdx = toml.indexOf(envHeader);
  if (headerIdx === -1) return toml; // section doesn't exist

  // Find the end of this section (next [header] or EOF)
  const afterHeader = headerIdx + envHeader.length;
  const nextSection = toml.indexOf("\n[", afterHeader);
  const sectionEnd = nextSection === -1 ? toml.length : nextSection;

  let section = toml.substring(afterHeader, sectionEnd);

  for (const [key, value] of Object.entries(envVars)) {
    // Check if key already exists in section
    const keyRegex = new RegExp(`^${escapeRegex(key)}\\s*=\\s*.*$`, "m");
    const match = section.match(keyRegex);
    if (match) {
      // Replace existing value
      section = section.replace(keyRegex, `${key} = "${value}"`);
    } else {
      // Append before end of section
      section = section.trimEnd() + `\n${key} = "${value}"\n`;
    }
  }

  return toml.substring(0, afterHeader) + section + toml.substring(sectionEnd);
}

/**
 * Scan all [model_providers.*] sections for stream_idle_timeout_ms.
 * Returns an array of { provider, currentValue, needsUpdate }.
 */
export function auditStreamTimeouts(toml) {
  const results = [];
  // Find all model_providers sections
  const providerRegex = /^\[model_providers\.(\w+)\]/gm;
  let match;
  while ((match = providerRegex.exec(toml)) !== null) {
    const providerName = match[1];
    const sectionStart = match.index + match[0].length;
    const nextSection = toml.indexOf("\n[", sectionStart);
    const sectionEnd = nextSection === -1 ? toml.length : nextSection;
    const section = toml.substring(sectionStart, sectionEnd);

    const timeoutMatch = section.match(/stream_idle_timeout_ms\s*=\s*(\d+)/);
    const currentValue = timeoutMatch ? Number(timeoutMatch[1]) : null;

    results.push({
      provider: providerName,
      currentValue,
      needsUpdate:
        currentValue === null || currentValue < MIN_STREAM_IDLE_TIMEOUT_MS,
      recommended: RECOMMENDED_STREAM_IDLE_TIMEOUT_MS,
    });
  }
  return results;
}

/**
 * Set stream_idle_timeout_ms on a specific model provider section.
 * If the key already exists, update it.  If not, append it at the end of the section.
 *
 * @param {string} toml  Current TOML content
 * @param {string} providerName  e.g. "azure", "openai"
 * @param {number} value  Timeout in ms
 * @returns {string}  Updated TOML
 */
export function setStreamTimeout(toml, providerName, value) {
  const header = `[model_providers.${providerName}]`;
  const headerIdx = toml.indexOf(header);
  if (headerIdx === -1) return toml;

  const afterHeader = headerIdx + header.length;
  const nextSection = toml.indexOf("\n[", afterHeader);
  const sectionEnd = nextSection === -1 ? toml.length : nextSection;

  let section = toml.substring(afterHeader, sectionEnd);

  const timeoutRegex = /^stream_idle_timeout_ms\s*=\s*\d+.*$/m;
  if (timeoutRegex.test(section)) {
    section = section.replace(
      timeoutRegex,
      `stream_idle_timeout_ms = ${value}  # Updated by codex-monitor`,
    );
  } else {
    // Append to end of section
    section =
      section.trimEnd() +
      `\nstream_idle_timeout_ms = ${value}  # Added by codex-monitor\n`;
  }

  return toml.substring(0, afterHeader) + section + toml.substring(sectionEnd);
}

/**
 * Ensure retry settings exist on a model provider section.
 * Adds sensible defaults for long-running agentic workloads.
 */
export function ensureRetrySettings(toml, providerName) {
  const header = `[model_providers.${providerName}]`;
  const headerIdx = toml.indexOf(header);
  if (headerIdx === -1) return toml;

  const afterHeader = headerIdx + header.length;
  const nextSection = toml.indexOf("\n[", afterHeader);
  const sectionEnd = nextSection === -1 ? toml.length : nextSection;

  let section = toml.substring(afterHeader, sectionEnd);

  const defaults = {
    request_max_retries: 6,
    stream_max_retries: 15,
  };

  for (const [key, defaultVal] of Object.entries(defaults)) {
    const keyRegex = new RegExp(`^${key}\\s*=`, "m");
    if (!keyRegex.test(section)) {
      section =
        section.trimEnd() +
        `\n${key} = ${defaultVal}  # Added by codex-monitor\n`;
    }
  }

  return toml.substring(0, afterHeader) + section + toml.substring(sectionEnd);
}

/**
 * High-level: ensure the config.toml is properly configured for codex-monitor.
 *
 * Returns an object describing what was done:
 *   { created, vkAdded, vkEnvUpdated, timeoutsFixed[], retriesAdded[], path }
 *
 * @param {object} opts
 * @param {string} [opts.vkBaseUrl]
 * @param {string} [opts.vkPort]
 * @param {string} [opts.vkHost]
 * @param {boolean} [opts.dryRun]  If true, returns result without writing
 */
export function ensureCodexConfig({
  vkBaseUrl = "http://127.0.0.1:54089",
  vkPort = "54089",
  vkHost = "127.0.0.1",
  dryRun = false,
} = {}) {
  const result = {
    path: CONFIG_PATH,
    created: false,
    vkAdded: false,
    vkEnvUpdated: false,
    agentSdkAdded: false,
    timeoutsFixed: [],
    retriesAdded: [],
    noChanges: false,
  };

  let toml = readCodexConfig();

  // If config.toml doesn't exist at all, create a minimal one
  if (!toml) {
    result.created = true;
    toml = [
      "# Codex CLI configuration",
      "# Generated by codex-monitor setup wizard",
      "#",
      "# See: codex --help or https://github.com/openai/codex for details.",
      "",
      "",
    ].join("\n");
  }

  // ── 1. Ensure vibe_kanban MCP server ──────────────────────

  if (!hasVibeKanbanMcp(toml)) {
    toml += buildVibeKanbanBlock({ vkBaseUrl, vkPort, vkHost });
    result.vkAdded = true;
  } else {
    // MCP section exists — ensure env vars are up to date
    if (!hasVibeKanbanEnv(toml)) {
      // Has the server but no env section — append env block
      const envBlock = [
        "",
        "[mcp_servers.vibe_kanban.env]",
        "# Ensure MCP always targets the correct VK API endpoint.",
        `VK_BASE_URL = "${vkBaseUrl}"`,
        `VK_ENDPOINT_URL = "${vkBaseUrl}"`,
        `VK_RECOVERY_PORT = "${vkPort}"`,
        "# Also bind the MCP-hosted VK instance to the expected port if it starts one.",
        `PORT = "${vkPort}"`,
        `HOST = "${vkHost}"`,
        "",
      ].join("\n");

      // Insert after [mcp_servers.vibe_kanban] section content, before next section
      const vkHeader = "[mcp_servers.vibe_kanban]";
      const vkIdx = toml.indexOf(vkHeader);
      const afterVk = vkIdx + vkHeader.length;
      const nextSectionAfterVk = toml.indexOf("\n[", afterVk);

      if (nextSectionAfterVk === -1) {
        toml += envBlock;
      } else {
        toml =
          toml.substring(0, nextSectionAfterVk) +
          "\n" +
          envBlock +
          toml.substring(nextSectionAfterVk);
      }
      result.vkEnvUpdated = true;
    } else {
      // Both server and env exist — ensure values match
      const envVars = {
        VK_BASE_URL: vkBaseUrl,
        VK_ENDPOINT_URL: vkBaseUrl,
        VK_RECOVERY_PORT: vkPort,
        PORT: vkPort,
        HOST: vkHost,
      };
      const before = toml;
      toml = updateVibeKanbanEnv(toml, envVars);
      if (toml !== before) {
        result.vkEnvUpdated = true;
      }
    }
  }

  // ── 1b. Ensure agent SDK selection block ──────────────────

  if (!hasAgentSdkConfig(toml)) {
    toml += buildAgentSdkBlock();
    result.agentSdkAdded = true;
  }

  // ── 2. Audit and fix stream timeouts ──────────────────────

  const timeouts = auditStreamTimeouts(toml);
  for (const t of timeouts) {
    if (t.needsUpdate) {
      toml = setStreamTimeout(toml, t.provider, t.recommended);
      result.timeoutsFixed.push({
        provider: t.provider,
        from: t.currentValue,
        to: t.recommended,
      });
    }
  }

  // ── 3. Ensure retry settings ──────────────────────────────

  for (const t of timeouts) {
    const before = toml;
    toml = ensureRetrySettings(toml, t.provider);
    if (toml !== before) {
      result.retriesAdded.push(t.provider);
    }
  }

  // ── Check if anything changed ─────────────────────────────

  const original = readCodexConfig();
  if (toml === original && !result.created) {
    result.noChanges = true;
    return result;
  }

  // ── Write ─────────────────────────────────────────────────

  if (!dryRun) {
    writeCodexConfig(toml);
  }

  return result;
}

/**
 * Print a human-friendly summary of what ensureCodexConfig() did.
 * @param {object} result  Return value from ensureCodexConfig()
 * @param {(msg: string) => void} [log]  Logger (default: console.log)
 */
export function printConfigSummary(result, log = console.log) {
  if (result.noChanges) {
    log("  ✅ Codex CLI config is already up to date");
    log(`     ${result.path}`);
    return;
  }

  if (result.created) {
    log("  📝 Created new Codex CLI config");
  }

  if (result.vkAdded) {
    log("  ✅ Added Vibe-Kanban MCP server to Codex config");
  }

  if (result.vkEnvUpdated) {
    log("  ✅ Updated Vibe-Kanban MCP environment variables");
  }

  if (result.agentSdkAdded) {
    log("  ✅ Added agent SDK selection block");
  }

  for (const t of result.timeoutsFixed) {
    const fromLabel =
      t.from === null ? "not set" : `${(t.from / 1000).toFixed(0)}s`;
    const toLabel = `${(t.to / 1000 / 60).toFixed(0)} min`;
    log(
      `  ✅ Set stream_idle_timeout_ms on [${t.provider}]: ${fromLabel} → ${toLabel}`,
    );
  }

  for (const p of result.retriesAdded) {
    log(`  ✅ Added retry settings to [${p}]`);
  }

  log(`     Config: ${result.path}`);
}

// ── Internal Helpers ─────────────────────────────────────────────────────────

function escapeRegex(str) {
  return str.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
