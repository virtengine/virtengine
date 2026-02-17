/**
 * codex-config.mjs — Manages the Codex CLI config (~/.codex/config.toml)
 *
 * Ensures the user's Codex CLI configuration has:
 *   1. A vibe_kanban MCP server section with the correct env vars
 *   2. Sufficient stream_idle_timeout_ms on all model providers
 *   3. Recommended defaults for long-running agentic workloads
 *   4. Feature flags for sub-agents, memory, undo, collaboration
 *   5. Sandbox permissions and shell environment policy
 *   6. Common MCP servers (context7, microsoft-docs)
 *
 * Uses string-based TOML manipulation (no parser dependency) — we only
 * append or patch well-known sections rather than rewriting the whole file.
 */

import { existsSync, readFileSync, writeFileSync, mkdirSync } from "node:fs";
import { resolve, dirname } from "node:path";
import { homedir } from "node:os";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

/**
 * Read the vibe-kanban version from the local package.json dependency.
 * Falls back to "latest" if not found (shouldn't happen in normal usage).
 */
function getVibeKanbanVersion() {
  try {
    const pkgPath = resolve(__dirname, "package.json");
    const pkg = JSON.parse(readFileSync(pkgPath, "utf8"));
    return pkg.dependencies?.["vibe-kanban"] || "latest";
  } catch {
    return "latest";
  }
}

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

// ── Feature Flags ────────────────────────────────────────────────────────────

/**
 * Feature flags that should be enabled for sub-agents, collaboration,
 * memory, and continuous operation.  Keys are the [features] TOML keys;
 * values are { default, envVar, comment }.
 */
const RECOMMENDED_FEATURES = {
  // Sub-agents & collaboration
  child_agents_md:        { default: true, envVar: "CODEX_FEATURES_CHILD_AGENTS_MD",   comment: "Enable sub-agent discovery via CODEX.md" },
  collab:                 { default: true, envVar: "CODEX_FEATURES_COLLAB",             comment: "Enable collaboration mode" },
  collaboration_modes:    { default: true, envVar: "CODEX_FEATURES_COLLABORATION_MODES", comment: "Enable collaboration mode selection" },

  // Continuity & recovery
  memory_tool:            { default: true, envVar: "CODEX_FEATURES_MEMORY_TOOL",        comment: "Persistent memory across sessions" },
  undo:                   { default: true, envVar: "CODEX_FEATURES_UNDO",               comment: "Safe rollback of agent changes" },
  steer:                  { default: true, envVar: "CODEX_FEATURES_STEER",              comment: "Live steering during runs" },
  personality:            { default: true, envVar: "CODEX_FEATURES_PERSONALITY",         comment: "Agent personality persistence" },

  // Sandbox & execution
  use_linux_sandbox_bwrap:{ default: true, envVar: "CODEX_FEATURES_BWRAP",              comment: "Linux bubblewrap sandbox" },
  shell_tool:             { default: true, envVar: null,                                 comment: "Shell tool access" },
  unified_exec:           { default: true, envVar: null,                                 comment: "Unified execution" },
  shell_snapshot:         { default: true, envVar: null,                                 comment: "Shell state snapshots" },
  request_rule:           { default: true, envVar: null,                                 comment: "Request-level approval rules" },

  // Performance & networking
  enable_request_compression: { default: true, envVar: null,                             comment: "Compress requests" },
  remote_models:          { default: true, envVar: null,                                 comment: "Remote model support" },
  skill_mcp_dependency_install: { default: true, envVar: null,                           comment: "Auto-install MCP skill deps" },

  // Experimental (disabled by default unless explicitly enabled)
  apps:                   { default: true, envVar: "CODEX_FEATURES_APPS",               comment: "ChatGPT Apps integration" },
};

const CRITICAL_ALWAYS_ON_FEATURES = new Set([
  "child_agents_md",
  "memory_tool",
  "collab",
  "collaboration_modes",
  "shell_tool",
  "unified_exec",
]);

/**
 * Check whether config has a [features] section.
 */
export function hasFeaturesSection(toml) {
  return /^\[features\]/m.test(toml);
}

/**
 * Check whether config has a [shell_environment_policy] section.
 */
export function hasShellEnvPolicy(toml) {
  return /^\[shell_environment_policy\]/m.test(toml);
}

/**
 * Check whether config has sandbox_permissions set (top-level key).
 */
export function hasSandboxPermissions(toml) {
  return /^sandbox_permissions\s*=/m.test(toml);
}

/**
 * Build a [features] block with the recommended flags.
 * Reads environment overrides: set CODEX_FEATURES_<NAME>=false to disable.
 *
 * @param {object} [envOverrides]  Optional env map (defaults to process.env)
 * @returns {string}  TOML block
 */
export function buildFeaturesBlock(envOverrides = process.env) {
  const lines = [
    "",
    "# ── Feature flags (added by codex-monitor) ──",
    "[features]",
  ];

  for (const [key, meta] of Object.entries(RECOMMENDED_FEATURES)) {
    let enabled = meta.default;
    // Check env override
    if (meta.envVar && envOverrides[meta.envVar] !== undefined) {
      enabled = parseBoolEnv(envOverrides[meta.envVar]);
    }
    if (meta.comment) lines.push(`# ${meta.comment}`);
    lines.push(`${key} = ${enabled}`);
  }

  lines.push("");
  return lines.join("\n");
}

/**
 * Ensure all recommended feature flags are present in an existing [features]
 * section.  Only adds missing keys — never overwrites user choices.
 *
 * @param {string} toml  Current config.toml content
 * @param {object} [envOverrides]  Optional env map (defaults to process.env)
 * @returns {{ toml: string, added: string[] }}
 */
export function ensureFeatureFlags(toml, envOverrides = process.env) {
  const added = [];

  if (!hasFeaturesSection(toml)) {
    toml += buildFeaturesBlock(envOverrides);
    added.push(...Object.keys(RECOMMENDED_FEATURES));
    return { toml, added };
  }

  // Find the [features] section boundaries
  const header = "[features]";
  const headerIdx = toml.indexOf(header);
  const afterHeader = headerIdx + header.length;
  const nextSection = toml.indexOf("\n[", afterHeader);
  const sectionEnd = nextSection === -1 ? toml.length : nextSection;
  let section = toml.substring(afterHeader, sectionEnd);

  for (const [key, meta] of Object.entries(RECOMMENDED_FEATURES)) {
    const keyRegex = new RegExp(`^${escapeRegex(key)}\\s*=`, "m");
    if (!keyRegex.test(section)) {
      let enabled = meta.default;
      if (meta.envVar && envOverrides[meta.envVar] !== undefined) {
        enabled = parseBoolEnv(envOverrides[meta.envVar]);
      }
      section = section.trimEnd() + `\n${key} = ${enabled}\n`;
      added.push(key);
      continue;
    }

    if (CRITICAL_ALWAYS_ON_FEATURES.has(key)) {
      const disabledRegex = new RegExp(
        `^(${escapeRegex(key)}\\s*=\\s*)false\\b.*$`,
        "m",
      );
      if (disabledRegex.test(section)) {
        section = section.replace(disabledRegex, `$1true`);
      }
    }
  }

  toml = toml.substring(0, afterHeader) + section + toml.substring(sectionEnd);
  return { toml, added };
}

/**
 * Build the sandbox_permissions top-level key.
 * Default: ["disk-full-read-access"] for agentic workloads.
 *
 * @param {string} [envValue]  CODEX_SANDBOX_PERMISSIONS env var value
 * @returns {string}  TOML line(s)
 */
export function buildSandboxPermissions(envValue) {
  const perms = envValue
    ? envValue.split(",").map((s) => `"${s.trim()}"`)
    : ['"disk-full-read-access"'];
  return `\n# Sandbox permissions (added by codex-monitor)\nsandbox_permissions = [${perms.join(", ")}]\n`;
}

/**
 * Build the [shell_environment_policy] section.
 * Default: inherit = "all" so .NET, Go, Node etc. env vars are visible.
 *
 * @param {string} [policy]  "all" | "none" | "allowlist"
 * @returns {string}  TOML block
 */
export function buildShellEnvPolicy(policy = "all") {
  return [
    "",
    "# ── Shell environment policy (added by codex-monitor) ──",
    "[shell_environment_policy]",
    `inherit = "${policy}"`,
    "",
  ].join("\n");
}

/**
 * Check whether config has a [mcp_servers.context7] section.
 */
export function hasContext7Mcp(toml) {
  return /^\[mcp_servers\.context7\]/m.test(toml);
}

/**
 * Check whether config has a [mcp_servers.microsoft-docs] or microsoft_docs section.
 */
export function hasMicrosoftDocsMcp(toml) {
  return /^\[mcp_servers\.microsoft[_-]docs\]/m.test(toml);
}

/**
 * Build MCP server blocks for context7 and microsoft-docs.
 * These are universally useful for documentation lookups.
 */
export function buildCommonMcpBlocks() {
  return [
    "",
    "# ── Common MCP servers (added by codex-monitor) ──",
    "[mcp_servers.context7]",
    'command = "npx"',
    'args = ["-y", "@upstash/context7-mcp"]',
    "",
    "[mcp_servers.sequential-thinking]",
    'command = "npx"',
    'args = ["-y", "@modelcontextprotocol/server-sequential-thinking"]',
    "",
    "[mcp_servers.playwright]",
    'command = "npx"',
    'args = ["-y", "@playwright/mcp@latest"]',
    "",
    "[mcp_servers.microsoft-docs]",
    'url = "https://learn.microsoft.com/api/mcp"',
    "",
  ].join("\n");
}

function hasNamedMcpServer(toml, name) {
  return new RegExp(`^\\[mcp_servers\\.${escapeRegex(name)}\\]`, "m").test(
    toml,
  );
}

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
 * Remove the [mcp_servers.vibe_kanban] and [mcp_servers.vibe_kanban.env]
 * sections (and their contents) from config.toml.
 * Returns the cleaned TOML string.
 */
export function removeVibeKanbanMcp(toml) {
  // Line-based approach: walk lines and skip VK-related sections.
  const lines = toml.split("\n");
  const out = [];
  let skipping = false;
  // Track comment lines immediately preceding a VK section header
  let pendingComments = [];

  for (const line of lines) {
    // Detect section headers: lines starting with [ that aren't array values
    const isSectionHeader = /^\[[\w]/.test(line);
    const isVkSection =
      /^\[mcp_servers\.vibe_kanban\b/.test(line);

    if (isVkSection) {
      // Drop any pending comment lines (they belong to this VK section)
      pendingComments = [];
      skipping = true;
      continue;
    }

    if (skipping && isSectionHeader) {
      // We've reached the next non-VK section — stop skipping
      skipping = false;
    }

    if (skipping) continue;

    // Buffer comment/blank lines that might precede a VK section header
    if (/^#.*[Vv]ibe.[Kk]anban/.test(line) || /^# ── .*[Vv]ibe.[Kk]anban/.test(line)) {
      pendingComments.push(line);
      continue;
    }

    // Flush pending comments (they weren't followed by a VK header)
    if (pendingComments.length) {
      out.push(...pendingComments);
      pendingComments = [];
    }

    out.push(line);
  }

  // Flush any remaining pending comments
  if (pendingComments.length) {
    out.push(...pendingComments);
  }

  // Clean up excessive blank lines
  return out.join("\n").replace(/\n{3,}/g, "\n\n");
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
 * The version is pinned from the local package.json dependency to avoid
 * slow npx re-downloads when @latest resolves to a new version.
 *
 * Only VK_BASE_URL and VK_ENDPOINT_URL are set — the MCP server reads
 * the backend port from the VK port file, so PORT/HOST env vars are not
 * needed and were removed to avoid confusion.
 *
 * @param {object} opts
 * @param {string} opts.vkBaseUrl   e.g. "http://127.0.0.1:54089"
 */
export function buildVibeKanbanBlock({
  vkBaseUrl = "http://127.0.0.1:54089",
} = {}) {
  const vkVersion = getVibeKanbanVersion();
  return [
    "",
    "# ── Vibe-Kanban MCP (added by codex-monitor) ──",
    "[mcp_servers.vibe_kanban]",
    "startup_timeout_sec = 120",
    "args = [",
    '    "-y",',
    `    "vibe-kanban@${vkVersion}",`,
    '    "--mcp",',
    "]",
    'command = "npx"',
    'tools = ["*"]',
    "",
    "[mcp_servers.vibe_kanban.env]",
    "# Ensure MCP always targets the correct VK API endpoint.",
    `VK_BASE_URL = "${vkBaseUrl}"`,
    `VK_ENDPOINT_URL = "${vkBaseUrl}"`,
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
 *   { created, vkAdded, vkEnvUpdated, timeoutsFixed[], retriesAdded[],
 *     featuresAdded[], sandboxAdded, shellEnvAdded, commonMcpAdded, path }
 *
 * @param {object} opts
 * @param {string}  [opts.vkBaseUrl]
 * @param {boolean} [opts.skipVk]
 * @param {boolean} [opts.dryRun]  If true, returns result without writing
 * @param {object}  [opts.env]     Environment overrides (defaults to process.env)
 */
export function ensureCodexConfig({
  vkBaseUrl = "http://127.0.0.1:54089",
  skipVk = false,
  dryRun = false,
  env = process.env,
} = {}) {
  const result = {
    path: CONFIG_PATH,
    created: false,
    vkAdded: false,
    vkRemoved: false,
    vkEnvUpdated: false,
    agentSdkAdded: false,
    featuresAdded: [],
    sandboxAdded: false,
    shellEnvAdded: false,
    commonMcpAdded: false,
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

  // ── 1. Vibe-Kanban MCP server ────────────────────────────
  // When VK is not the active kanban backend, remove the MCP section
  // so the Codex CLI doesn't try to spawn it.

  if (skipVk) {
    if (hasVibeKanbanMcp(toml)) {
      toml = removeVibeKanbanMcp(toml);
      result.vkRemoved = true;
    }
  } else if (!hasVibeKanbanMcp(toml)) {
    toml += buildVibeKanbanBlock({ vkBaseUrl });
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

  // ── 1c. Ensure feature flags (sub-agents, memory, etc.) ──

  {
    const { toml: updated, added } = ensureFeatureFlags(toml, env);
    if (added.length > 0) {
      toml = updated;
      result.featuresAdded = added;
    }
  }

  // ── 1d. Ensure sandbox permissions ────────────────────────

  if (!hasSandboxPermissions(toml)) {
    const envPerms = env.CODEX_SANDBOX_PERMISSIONS || "";
    toml = toml.trimEnd() + "\n" + buildSandboxPermissions(envPerms || undefined);
    result.sandboxAdded = true;
  }

  // ── 1e. Ensure shell environment policy ───────────────────

  if (!hasShellEnvPolicy(toml)) {
    const policy = env.CODEX_SHELL_ENV_POLICY || "all";
    toml += buildShellEnvPolicy(policy);
    result.shellEnvAdded = true;
  }

  // ── 1f. Ensure common MCP servers ───────────────────────────

  {
    const missing = [];
    if (!hasContext7Mcp(toml)) missing.push("context7");
    if (!hasNamedMcpServer(toml, "sequential-thinking")) {
      missing.push("sequential-thinking");
    }
    if (!hasNamedMcpServer(toml, "playwright")) missing.push("playwright");
    if (!hasMicrosoftDocsMcp(toml)) missing.push("microsoft-docs");

    if (missing.length > 0) {
      if (missing.length >= 4) {
        toml += buildCommonMcpBlocks();
      } else {
        if (missing.includes("context7")) {
          toml += [
            "",
            "[mcp_servers.context7]",
            'command = "npx"',
            'args = ["-y", "@upstash/context7-mcp"]',
            "",
          ].join("\n");
        }
        if (missing.includes("sequential-thinking")) {
          toml += [
            "",
            "[mcp_servers.sequential-thinking]",
            'command = "npx"',
            'args = ["-y", "@modelcontextprotocol/server-sequential-thinking"]',
            "",
          ].join("\n");
        }
        if (missing.includes("playwright")) {
          toml += [
            "",
            "[mcp_servers.playwright]",
            'command = "npx"',
            'args = ["-y", "@playwright/mcp@latest"]',
            "",
          ].join("\n");
        }
        if (missing.includes("microsoft-docs")) {
          toml += [
            "",
            "[mcp_servers.microsoft-docs]",
            'url = "https://learn.microsoft.com/api/mcp"',
            "",
          ].join("\n");
        }
      }
      result.commonMcpAdded = true;
    }
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

  if (result.vkRemoved) {
    log("  🗑️  Removed Vibe-Kanban MCP server (VK backend not active)");
  }

  if (result.vkEnvUpdated) {
    log("  ✅ Updated Vibe-Kanban MCP environment variables");
  }

  if (result.agentSdkAdded) {
    log("  ✅ Added agent SDK selection block");
  }

  if (result.featuresAdded && result.featuresAdded.length > 0) {
    const key = result.featuresAdded.length <= 5
      ? result.featuresAdded.join(", ")
      : `${result.featuresAdded.length} feature flags`;
    log(`  ✅ Added feature flags: ${key}`);
  }

  if (result.sandboxAdded) {
    log("  ✅ Added sandbox permissions (disk-full-read-access)");
  }

  if (result.shellEnvAdded) {
    log("  ✅ Added shell environment policy (inherit=all)");
  }

  if (result.commonMcpAdded) {
    log(
      "  ✅ Added common MCP servers (context7, sequential-thinking, playwright, microsoft-docs)",
    );
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

function parseBoolEnv(value) {
  const raw = String(value ?? "").trim().toLowerCase();
  if (["0", "false", "no", "off", "n"].includes(raw)) return false;
  return true;
}
