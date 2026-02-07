// Copyright 2026 VirtEngine Authors
import { readFile } from "node:fs/promises";
import { resolve } from "node:path";

const DEFAULT_ROLE = "primary";
const DEFAULT_WORKSPACE_ID = "local";
const DEFAULT_WORKSPACE_NAME = "Local";

function parseJson(text, source, errors) {
  try {
    return JSON.parse(text);
  } catch (err) {
    errors.push(
      `Failed to parse workspace registry JSON from ${source}: ${
        err?.message || err
      }`,
    );
    return null;
  }
}

function normalizeString(value) {
  if (typeof value !== "string") return "";
  return value.trim();
}

function parseList(value) {
  if (!value) return [];
  if (Array.isArray(value)) {
    return value
      .map((item) => normalizeString(String(item)))
      .filter(Boolean);
  }
  if (typeof value === "string") {
    return value
      .split(",")
      .map((item) => normalizeString(item))
      .filter(Boolean);
  }
  return [];
}

function normalizeHost(host) {
  const trimmed = normalizeString(host);
  if (!trimmed) return "";
  const withScheme = /^(https?:)?\/\//i.test(trimmed)
    ? trimmed
    : `http://${trimmed}`;
  return withScheme.replace(/\/+$/, "");
}

function normalizeModelPriorityEntry(entry) {
  if (!entry) return null;
  if (typeof entry === "string") {
    const trimmed = normalizeString(entry);
    if (!trimmed) return null;
    if (trimmed.includes("/")) {
      const [executorRaw, variantRaw] = trimmed.split("/", 2);
      const executor = normalizeString(executorRaw).toUpperCase();
      const variant = normalizeString(variantRaw);
      if (executor && variant) {
        return { model: trimmed, executor, variant };
      }
    }
    return { model: trimmed };
  }
  if (typeof entry === "object") {
    const model = normalizeString(entry.model || entry.name || entry.id);
    const executor = normalizeString(entry.executor || entry.exec).toUpperCase();
    const variant = normalizeString(entry.variant || entry.profile || "");
    const normalized = {};
    if (model) normalized.model = model;
    if (executor) normalized.executor = executor;
    if (variant) normalized.variant = variant;
    return Object.keys(normalized).length ? normalized : null;
  }
  return null;
}

function parseModelPriorityInput(value, errors, source) {
  if (!value) return [];
  if (Array.isArray(value)) {
    return value
      .map((entry) => normalizeModelPriorityEntry(entry))
      .filter(Boolean);
  }
  if (typeof value === "string") {
    const trimmed = value.trim();
    if (!trimmed) return [];
    if (trimmed.startsWith("[") || trimmed.startsWith("{")) {
      const parsed = parseJson(trimmed, source || "env", errors);
      if (parsed) {
        return parseModelPriorityInput(parsed, errors, source);
      }
      return [];
    }
    return trimmed
      .split(",")
      .map((entry) => normalizeModelPriorityEntry(entry))
      .filter(Boolean);
  }
  if (typeof value === "object") {
    const normalized = normalizeModelPriorityEntry(value);
    return normalized ? [normalized] : [];
  }
  errors.push(
    `Invalid model priority list (${source || "unknown"}): must be array or string`,
  );
  return [];
}

function buildDefaultRegistry(env) {
  const host =
    env.VK_WORKSPACE_HOST ||
    env.VK_ENDPOINT_URL ||
    env.VK_BASE_URL ||
    `http://127.0.0.1:${env.VK_RECOVERY_PORT || "54089"}`;
  return {
    workspaces: [
      {
        id: env.VK_WORKSPACE_ID || DEFAULT_WORKSPACE_ID,
        name: env.VK_WORKSPACE_NAME || DEFAULT_WORKSPACE_NAME,
        role: env.VK_WORKSPACE_ROLE || DEFAULT_ROLE,
        host,
        capabilities: parseList(env.VK_WORKSPACE_CAPABILITIES),
        model_priority: parseModelPriorityInput(
          env.VK_WORKSPACE_MODEL_PRIORITY || env.VK_WORKSPACE_MODELS,
          [],
          "env",
        ),
      },
    ],
  };
}

function buildEnvWorkspaceOverride(env) {
  const override = {};
  const fields = [
    "VK_WORKSPACE_ID",
    "VK_WORKSPACE_NAME",
    "VK_WORKSPACE_ROLE",
    "VK_WORKSPACE_HOST",
    "VK_WORKSPACE_CAPABILITIES",
    "VK_WORKSPACE_MODEL_PRIORITY",
    "VK_WORKSPACE_MODELS",
  ];
  const hasOverride = fields.some((key) => env[key]);
  if (!hasOverride) return null;
  if (env.VK_WORKSPACE_ID) override.id = env.VK_WORKSPACE_ID;
  if (env.VK_WORKSPACE_NAME) override.name = env.VK_WORKSPACE_NAME;
  if (env.VK_WORKSPACE_ROLE) override.role = env.VK_WORKSPACE_ROLE;
  if (env.VK_WORKSPACE_HOST) override.host = env.VK_WORKSPACE_HOST;
  if (env.VK_WORKSPACE_CAPABILITIES) {
    override.capabilities = parseList(env.VK_WORKSPACE_CAPABILITIES);
  }
  if (env.VK_WORKSPACE_MODEL_PRIORITY || env.VK_WORKSPACE_MODELS) {
    override.model_priority = parseModelPriorityInput(
      env.VK_WORKSPACE_MODEL_PRIORITY || env.VK_WORKSPACE_MODELS,
      [],
      "env",
    );
  }
  return override;
}

function normalizeWorkspace(raw, errors, idx) {
  const workspace = {
    id: normalizeString(raw?.id),
    name: normalizeString(raw?.name),
    role: normalizeString(raw?.role) || DEFAULT_ROLE,
    host: normalizeHost(raw?.host),
    capabilities: parseList(raw?.capabilities),
    model_priority: parseModelPriorityInput(
      raw?.model_priority ?? raw?.modelPriority ?? raw?.models,
      errors,
      `workspace[${idx}].model_priority`,
    ),
  };
  if (!workspace.id) {
    errors.push(`workspace[${idx}].id is required`);
  }
  if (!workspace.name) {
    errors.push(`workspace[${idx}].name is required`);
  }
  if (!workspace.role) {
    errors.push(`workspace[${idx}].role is required`);
  }
  if (!workspace.host) {
    errors.push(`workspace[${idx}].host is required`);
  } else {
    try {
      new URL(workspace.host);
    } catch (err) {
      errors.push(
        `workspace[${idx}].host is invalid: ${err?.message || err}`,
      );
    }
  }
  return workspace;
}

function normalizeRegistryInput(input, errors) {
  if (!input || typeof input !== "object") {
    errors.push("workspace registry must be an object");
    return { workspaces: [] };
  }
  const rawWorkspaces = Array.isArray(input)
    ? input
    : input.workspaces || input.registry || [];
  if (!Array.isArray(rawWorkspaces)) {
    errors.push("workspace registry must include a workspaces array");
    return { workspaces: [] };
  }
  const workspaces = rawWorkspaces.map((raw, idx) =>
    normalizeWorkspace(raw, errors, idx),
  );
  return { workspaces };
}

export async function loadWorkspaceRegistry({
  env = process.env,
  baseDir = process.cwd(),
} = {}) {
  const errors = [];
  const registryPath =
    env.VK_WORKSPACE_REGISTRY_FILE || env.VK_WORKSPACE_REGISTRY_PATH;
  const registryJson = env.VK_WORKSPACE_REGISTRY_JSON;

  let rawRegistry = null;
  let source = "default";

  if (registryPath) {
    const resolvedPath = resolve(baseDir, registryPath);
    try {
      const contents = await readFile(resolvedPath, "utf8");
      rawRegistry = parseJson(contents, resolvedPath, errors);
      source = `file:${resolvedPath}`;
    } catch (err) {
      errors.push(
        `Failed to read workspace registry file ${resolvedPath}: ${
          err?.message || err
        }`,
      );
    }
  }

  if (!rawRegistry && registryJson) {
    rawRegistry = parseJson(registryJson, "VK_WORKSPACE_REGISTRY_JSON", errors);
    source = "env:VK_WORKSPACE_REGISTRY_JSON";
  }

  if (!rawRegistry) {
    rawRegistry = buildDefaultRegistry(env);
  }

  const envOverride = buildEnvWorkspaceOverride(env);
  if (envOverride) {
    const targetId = envOverride.id;
    if (targetId && Array.isArray(rawRegistry.workspaces)) {
      const idx = rawRegistry.workspaces.findIndex((w) => w?.id === targetId);
      if (idx >= 0) {
        rawRegistry.workspaces[idx] = {
          ...rawRegistry.workspaces[idx],
          ...envOverride,
        };
      } else {
        rawRegistry.workspaces = [envOverride, ...rawRegistry.workspaces];
      }
    } else if (Array.isArray(rawRegistry.workspaces)) {
      rawRegistry.workspaces = rawRegistry.workspaces.length
        ? [
            {
              ...rawRegistry.workspaces[0],
              ...envOverride,
            },
            ...rawRegistry.workspaces.slice(1),
          ]
        : [envOverride];
    } else {
      rawRegistry = { workspaces: [envOverride] };
    }
  }

  const registry = normalizeRegistryInput(rawRegistry, errors);
  return { registry, errors, source };
}

export function normalizeRole(role) {
  return normalizeString(role).toLowerCase();
}

export function normalizeModelToken(model) {
  return normalizeString(model).toLowerCase();
}

export function guessExecutorProfile(modelToken) {
  const token = normalizeModelToken(modelToken);
  if (!token) return null;
  if (token.includes("claude") || token.includes("copilot")) {
    return { executor: "COPILOT", variant: "CLAUDE_OPUS_4_6" };
  }
  return { executor: "CODEX", variant: "DEFAULT" };
}

export function selectModelPriority(workspace, requestedModel) {
  const list = Array.isArray(workspace?.model_priority)
    ? workspace.model_priority
    : [];
  if (!list.length) return null;
  if (requestedModel) {
    const token = normalizeModelToken(requestedModel);
    const match = list.find((entry) => {
      const modelToken = normalizeModelToken(entry.model);
      return modelToken && modelToken === token;
    });
    if (match) return match;
  }
  return list[0];
}

export function workspaceSupportsModel(workspace, requestedModel) {
  if (!requestedModel) return true;
  const token = normalizeModelToken(requestedModel);
  if (!token) return true;
  const list = Array.isArray(workspace?.model_priority)
    ? workspace.model_priority
    : [];
  return list.some((entry) => {
    const modelToken = normalizeModelToken(entry.model);
    return modelToken && modelToken === token;
  });
}

export function getExecutorProfileForModel(workspace, requestedModel) {
  const preferred = selectModelPriority(workspace, requestedModel);
  if (!preferred) return null;
  if (preferred.executor && preferred.variant) {
    return { executor: preferred.executor, variant: preferred.variant };
  }
  if (preferred.model) {
    return guessExecutorProfile(preferred.model);
  }
  return null;
}

export function getDefaultExecutorProfile(executorName) {
  switch (normalizeString(executorName).toUpperCase()) {
    case "COPILOT":
      return { executor: "COPILOT", variant: "CLAUDE_OPUS_4_6" };
    case "CODEX":
    default:
      return { executor: "CODEX", variant: "DEFAULT" };
  }
}

export function normalizeWorkspaceRegistry(registry) {
  const errors = [];
  const normalized = normalizeRegistryInput(registry, errors);
  return { registry: normalized, errors };
}
