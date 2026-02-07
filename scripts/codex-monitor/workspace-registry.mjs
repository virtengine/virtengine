/**
 * workspace-registry.mjs — Workspace registry loader + validation for codex-monitor.
 */

import { readFile } from "node:fs/promises";
import { resolve as resolvePath, isAbsolute } from "node:path";

const DEFAULT_MODEL_PRIORITY = [
  "gpt-5.2-codex",
  "gpt-5.1-codex-max",
  "gpt-5.1-codex-mini",
  "claude-opus-4.6",
  "claude-code",
];

const DEFAULT_CAPABILITIES = ["vibe-kanban"];

function getDefaultHost() {
  return (
    process.env.VE_WORKSPACE_DEFAULT_HOST ||
    process.env.VK_ENDPOINT_URL ||
    process.env.VK_BASE_URL ||
    "http://127.0.0.1:54089"
  );
}

function parseJson(raw, label, errors) {
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch (err) {
    errors.push(`Invalid JSON in ${label}: ${err.message}`);
    return null;
  }
}

function normalizeRegistry(input, errors, warnings) {
  const defaultHost = getDefaultHost();
  const defaults = {
    role: "primary",
    host: defaultHost,
    capabilities: [...DEFAULT_CAPABILITIES],
    model_priority: [...DEFAULT_MODEL_PRIORITY],
  };

  const registry = {
    defaults: { ...defaults },
    workspaces: [],
  };

  if (input) {
    if (Array.isArray(input)) {
      registry.workspaces = input;
    } else if (typeof input === "object") {
      if (input.defaults && typeof input.defaults === "object") {
        registry.defaults = {
          ...registry.defaults,
          ...input.defaults,
        };
      }
      if (Array.isArray(input.workspaces)) {
        registry.workspaces = input.workspaces;
      }
    }
  }

  if (!Array.isArray(registry.workspaces) || registry.workspaces.length === 0) {
    warnings.push("No workspaces configured. Using default local workspace.");
    registry.workspaces = [
      {
        id: "local",
        name: "Local",
        role: "primary",
        host: defaultHost,
        capabilities: [...DEFAULT_CAPABILITIES],
        model_priority: [...DEFAULT_MODEL_PRIORITY],
      },
    ];
  }

  const seenIds = new Set();
  const normalized = [];

  for (const [index, ws] of registry.workspaces.entries()) {
    if (!ws || typeof ws !== "object") {
      errors.push(`Workspace #${index + 1} is not an object.`);
      continue;
    }

    const id = typeof ws.id === "string" ? ws.id.trim() : "";
    if (!id) {
      errors.push(`Workspace #${index + 1} is missing required field: id`);
      continue;
    }
    if (seenIds.has(id)) {
      errors.push(`Duplicate workspace id: ${id}`);
      continue;
    }
    seenIds.add(id);

    const name =
      typeof ws.name === "string" && ws.name.trim()
        ? ws.name.trim()
        : id;
    const role =
      typeof ws.role === "string" && ws.role.trim()
        ? ws.role.trim()
        : registry.defaults.role;
    const host =
      typeof ws.host === "string" && ws.host.trim()
        ? ws.host.trim()
        : registry.defaults.host;

    let capabilities = ws.capabilities;
    if (!Array.isArray(capabilities)) {
      if (capabilities !== undefined) {
        errors.push(`Workspace ${id}: capabilities must be an array.`);
      }
      capabilities = registry.defaults.capabilities;
    }
    capabilities = capabilities.map((c) => String(c)).filter((c) => c.trim());

    let modelPriority = ws.model_priority ?? ws.modelPriority ?? ws.models;
    if (!Array.isArray(modelPriority)) {
      if (modelPriority !== undefined) {
        errors.push(`Workspace ${id}: model_priority must be an array.`);
      }
      modelPriority = registry.defaults.model_priority;
    }

    const normalizedPriority = modelPriority
      .map((entry) => (typeof entry === "string" ? entry.trim() : entry))
      .filter((entry) => (typeof entry === "string" ? entry : true));

    if (normalizedPriority.length === 0) {
      warnings.push(`Workspace ${id}: model_priority empty. Using defaults.`);
    }

    normalized.push({
      id,
      name,
      role,
      host,
      capabilities,
      model_priority:
        normalizedPriority.length > 0
          ? normalizedPriority
          : [...DEFAULT_MODEL_PRIORITY],
    });
  }

  return {
    registry: {
      defaults: registry.defaults,
      workspaces: normalized,
    },
    errors,
    warnings,
  };
}

function resolveRegistryFilePath(rawPath) {
  if (!rawPath || typeof rawPath !== "string") return null;
  if (isAbsolute(rawPath)) return rawPath;
  return resolvePath(process.cwd(), rawPath);
}

export async function loadWorkspaceRegistry() {
  const errors = [];
  const warnings = [];
  let merged = null;

  const filePath = resolveRegistryFilePath(process.env.VE_WORKSPACE_REGISTRY_FILE);
  if (filePath) {
    try {
      const raw = await readFile(filePath, "utf8");
      const parsed = parseJson(raw, `file ${filePath}`, errors);
      if (parsed) merged = parsed;
    } catch (err) {
      errors.push(`Failed to read workspace registry file ${filePath}: ${err.message}`);
    }
  }

  if (process.env.VE_WORKSPACE_REGISTRY) {
    const parsed = parseJson(
      process.env.VE_WORKSPACE_REGISTRY,
      "VE_WORKSPACE_REGISTRY",
      errors,
    );
    if (parsed) {
      merged = merged
        ? { ...(typeof merged === "object" ? merged : {}), ...parsed }
        : parsed;
    }
  }

  return normalizeRegistry(merged, errors, warnings);
}

export function formatRegistryDiagnostics(errors, warnings) {
  const lines = [];
  if (errors.length > 0) {
    lines.push("❌ Workspace registry errors:");
    for (const err of errors) {
      lines.push(`  - ${err}`);
    }
  }
  if (warnings.length > 0) {
    if (lines.length > 0) lines.push("");
    lines.push("⚠️ Workspace registry warnings:");
    for (const warn of warnings) {
      lines.push(`  - ${warn}`);
    }
  }
  return lines.join("\n");
}

export function getDefaultModelPriority() {
  return [...DEFAULT_MODEL_PRIORITY];
}
