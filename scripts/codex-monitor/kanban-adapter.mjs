/**
 * kanban-adapter.mjs — Unified Kanban Board Abstraction
 *
 * Provides a common interface over multiple task-tracking backends:
 *   - Vibe-Kanban (VK)       — default, full-featured
 *   - GitHub Issues           — native GitHub integration
 *   - Jira (stub)             — enterprise project management
 *
 * This module handles TASK LIFECYCLE (tracking, status, metadata) only.
 * Code execution is handled separately by agent-pool.mjs.
 *
 * Configuration:
 *   - `KANBAN_BACKEND` env var: "vk" | "github" | "jira" (default: "vk")
 *   - `codex-monitor.config.json` → `kanban.backend` field
 *
 * EXPORTS:
 *   getKanbanAdapter()        → Returns the configured adapter instance
 *   setKanbanBackend(name)    → Switch backend at runtime
 *   getAvailableBackends()    → List available backends
 *   getKanbanBackendName()    → Get active backend name
 *   listProjects()            → Convenience: adapter.listProjects()
 *   listTasks(projectId, f?)  → Convenience: adapter.listTasks()
 *   getTask(taskId)           → Convenience: adapter.getTask()
 *   updateTaskStatus(id, s)   → Convenience: adapter.updateTaskStatus()
 *   createTask(projId, data)  → Convenience: adapter.createTask()
 *   deleteTask(taskId)        → Convenience: adapter.deleteTask()
 *
 * Each adapter implements the KanbanAdapter interface:
 *   - listTasks(projectId, filters?)     → Task[]
 *   - getTask(taskId)                    → Task
 *   - updateTaskStatus(taskId, status)   → Task
 *   - createTask(projectId, task)        → Task
 *   - deleteTask(taskId)                 → boolean
 *   - listProjects()                     → Project[]
 */

import { loadConfig } from "./config.mjs";

const TAG = "[kanban]";

// ---------------------------------------------------------------------------
// Normalised Task & Project Types
// ---------------------------------------------------------------------------

/**
 * @typedef {Object} KanbanTask
 * @property {string}      id          Unique task identifier.
 * @property {string}      title       Task title/summary.
 * @property {string}      description Full task description/body.
 * @property {string}      status      Normalised status: "todo"|"inprogress"|"inreview"|"done"|"cancelled".
 * @property {string|null} assignee    Assigned user/agent.
 * @property {string|null} priority    "low"|"medium"|"high"|"critical".
 * @property {string|null} projectId   Parent project identifier.
 * @property {string|null} branchName  Associated git branch.
 * @property {string|null} prNumber    Associated PR number.
 * @property {object}      meta        Backend-specific metadata.
 * @property {string}      backend     Which backend this came from.
 */

/**
 * @typedef {Object} KanbanProject
 * @property {string} id     Unique project identifier.
 * @property {string} name   Project name.
 * @property {object} meta   Backend-specific metadata.
 * @property {string} backend Which backend.
 */

// ---------------------------------------------------------------------------
// Status Normalisation
// ---------------------------------------------------------------------------

/** Map from various backend status strings to our canonical set */
const STATUS_MAP = {
  // VK statuses
  todo: "todo",
  inprogress: "inprogress",
  "in-progress": "inprogress",
  in_progress: "inprogress",
  inreview: "inreview",
  "in-review": "inreview",
  in_review: "inreview",
  done: "done",
  cancelled: "cancelled",
  canceled: "cancelled",
  // GitHub Issues
  open: "todo",
  closed: "done",
  // Jira-style
  "to do": "todo",
  "in progress": "inprogress",
  review: "inreview",
  resolved: "done",
};

function normaliseStatus(raw) {
  if (!raw) return "todo";
  const key = String(raw).toLowerCase().trim();
  return STATUS_MAP[key] || "todo";
}

// ---------------------------------------------------------------------------
// VK Adapter (Vibe-Kanban)
// ---------------------------------------------------------------------------

class VKAdapter {
  constructor() {
    this.name = "vk";
    this._fetchVk = null;
  }

  /**
   * Lazy-load the fetchVk helper from monitor.mjs or fall back to a minimal
   * implementation using the VK endpoint URL from config.
   */
  async _getFetchVk() {
    if (this._fetchVk) return this._fetchVk;

    // Try importing a standalone vk-api module first
    try {
      const mod = await import("./vk-api.mjs");
      const fn = mod.fetchVk || mod.default?.fetchVk || mod.default;
      if (typeof fn === "function") {
        this._fetchVk = fn;
        return this._fetchVk;
      }
    } catch {
      // Not available — build a minimal fetch wrapper
    }

    // Minimal fetch wrapper using config
    const cfg = loadConfig();
    const baseUrl = cfg.vkEndpointUrl || "http://127.0.0.1:54089";
    this._fetchVk = async (path, opts = {}) => {
      const url = `${baseUrl}${path.startsWith("/") ? path : "/" + path}`;
      const method = (opts.method || "GET").toUpperCase();
      const controller = new AbortController();
      const timeout = setTimeout(
        () => controller.abort(),
        opts.timeoutMs || 15_000,
      );
      try {
        const fetchOpts = {
          method,
          signal: controller.signal,
          headers: { "Content-Type": "application/json" },
        };
        if (opts.body && method !== "GET") {
          fetchOpts.body =
            typeof opts.body === "string"
              ? opts.body
              : JSON.stringify(opts.body);
        }
        const res = await fetch(url, fetchOpts);
        if (!res.ok) {
          const text = await res.text().catch(() => "");
          throw new Error(
            `VK API ${method} ${path} failed: ${res.status} ${text.slice(0, 200)}`,
          );
        }
        return await res.json();
      } finally {
        clearTimeout(timeout);
      }
    };
    return this._fetchVk;
  }

  async listProjects() {
    const fetchVk = await this._getFetchVk();
    const result = await fetchVk("/api/projects");
    const projects = Array.isArray(result) ? result : result?.data || [];
    return projects.map((p) => ({
      id: p.id,
      name: p.name || p.title || p.id,
      meta: p,
      backend: "vk",
    }));
  }

  async listTasks(projectId, filters = {}) {
    const fetchVk = await this._getFetchVk();
    let url = `/api/projects/${projectId}/tasks`;
    const params = [];
    if (filters.status)
      params.push(`status=${encodeURIComponent(filters.status)}`);
    if (filters.limit) params.push(`limit=${filters.limit}`);
    if (params.length) url += `?${params.join("&")}`;
    const result = await fetchVk(url);
    const tasks = Array.isArray(result)
      ? result
      : result?.data || result?.tasks || [];
    return tasks.map((t) => this._normaliseTask(t, projectId));
  }

  async getTask(taskId) {
    const fetchVk = await this._getFetchVk();
    const result = await fetchVk(`/api/tasks/${taskId}`);
    const task = result?.data || result;
    return this._normaliseTask(task);
  }

  async updateTaskStatus(taskId, status) {
    const fetchVk = await this._getFetchVk();
    const result = await fetchVk(`/api/tasks/${taskId}`, {
      method: "PATCH",
      body: { status },
    });
    const task = result?.data || result;
    return this._normaliseTask(task);
  }

  async createTask(projectId, taskData) {
    const fetchVk = await this._getFetchVk();
    const result = await fetchVk(`/api/projects/${projectId}/tasks`, {
      method: "POST",
      body: taskData,
    });
    const task = result?.data || result;
    return this._normaliseTask(task, projectId);
  }

  async deleteTask(taskId) {
    const fetchVk = await this._getFetchVk();
    await fetchVk(`/api/tasks/${taskId}`, { method: "DELETE" });
    return true;
  }

  _normaliseTask(raw, projectId = null) {
    if (!raw) return null;
    return {
      id: raw.id || raw.task_id || "",
      title: raw.title || raw.name || "",
      description: raw.description || raw.body || "",
      status: normaliseStatus(raw.status),
      assignee: raw.assignee || raw.assigned_to || null,
      priority: raw.priority || null,
      projectId: raw.project_id || projectId,
      branchName: raw.branch_name || raw.branchName || null,
      prNumber: raw.pr_number || raw.prNumber || null,
      meta: raw,
      backend: "vk",
    };
  }
}

// ---------------------------------------------------------------------------
// GitHub Issues Adapter
// ---------------------------------------------------------------------------

class GitHubIssuesAdapter {
  constructor() {
    this.name = "github";
    this._owner = process.env.GITHUB_REPO_OWNER || "virtengine";
    this._repo = process.env.GITHUB_REPO_NAME || "virtengine";
  }

  /** Execute a gh CLI command and return parsed JSON */
  async _gh(args) {
    const { execFile } = await import("node:child_process");
    const { promisify } = await import("node:util");
    const execFileAsync = promisify(execFile);
    try {
      const { stdout } = await execFileAsync("gh", args, {
        maxBuffer: 10 * 1024 * 1024,
        timeout: 30_000,
      });
      return JSON.parse(stdout);
    } catch (err) {
      throw new Error(`gh CLI failed: ${err.message}`);
    }
  }

  async listProjects() {
    // GitHub doesn't have "projects" in the same sense — return repo as project
    return [
      {
        id: `${this._owner}/${this._repo}`,
        name: this._repo,
        meta: { owner: this._owner, repo: this._repo },
        backend: "github",
      },
    ];
  }

  async listTasks(_projectId, filters = {}) {
    const args = [
      "issue",
      "list",
      "--repo",
      `${this._owner}/${this._repo}`,
      "--json",
      "number,title,body,state,assignees,labels,milestone",
      "--limit",
      String(filters.limit || 50),
    ];
    if (filters.status === "done") {
      args.push("--state", "closed");
    } else if (filters.status && filters.status !== "todo") {
      args.push("--label", filters.status);
    } else {
      args.push("--state", "open");
    }
    const issues = await this._gh(args);
    return (Array.isArray(issues) ? issues : []).map((i) =>
      this._normaliseIssue(i),
    );
  }

  async getTask(issueNumber) {
    const num = String(issueNumber).replace(/^#/, "");
    const issue = await this._gh([
      "issue",
      "view",
      num,
      "--repo",
      `${this._owner}/${this._repo}`,
      "--json",
      "number,title,body,state,assignees,labels,milestone",
    ]);
    return this._normaliseIssue(issue);
  }

  async updateTaskStatus(issueNumber, status) {
    const num = String(issueNumber).replace(/^#/, "");
    const normalised = normaliseStatus(status);
    if (normalised === "done" || normalised === "cancelled") {
      await this._gh([
        "issue",
        "close",
        num,
        "--repo",
        `${this._owner}/${this._repo}`,
      ]);
    } else {
      await this._gh([
        "issue",
        "reopen",
        num,
        "--repo",
        `${this._owner}/${this._repo}`,
      ]);
      // Add label for tracking intermediate states
      if (normalised !== "todo") {
        try {
          await this._gh([
            "issue",
            "edit",
            num,
            "--repo",
            `${this._owner}/${this._repo}`,
            "--add-label",
            normalised,
          ]);
        } catch {
          // Label might not exist — non-critical
        }
      }
    }
    return this.getTask(issueNumber);
  }

  async createTask(_projectId, taskData) {
    const args = [
      "issue",
      "create",
      "--repo",
      `${this._owner}/${this._repo}`,
      "--title",
      taskData.title || "New task",
      "--body",
      taskData.description || "",
    ];
    if (taskData.assignee) args.push("--assignee", taskData.assignee);
    if (taskData.labels) {
      for (const label of [].concat(taskData.labels)) {
        args.push("--label", label);
      }
    }
    const result = await this._gh(args);
    return typeof result === "object"
      ? this._normaliseIssue(result)
      : { id: result, backend: "github" };
  }

  async deleteTask(issueNumber) {
    // GitHub issues can't be deleted — close with "not planned"
    const num = String(issueNumber).replace(/^#/, "");
    await this._gh([
      "issue",
      "close",
      num,
      "--repo",
      `${this._owner}/${this._repo}`,
      "--reason",
      "not planned",
    ]);
    return true;
  }

  _normaliseIssue(issue) {
    if (!issue) return null;
    const labels = (issue.labels || []).map((l) =>
      typeof l === "string" ? l : l.name,
    );
    let status = "todo";
    if (issue.state === "closed" || issue.state === "CLOSED") {
      status = "done";
    } else if (
      labels.includes("inprogress") ||
      labels.includes("in-progress")
    ) {
      status = "inprogress";
    } else if (labels.includes("inreview") || labels.includes("in-review")) {
      status = "inreview";
    }

    // Extract branch name from issue body if present
    const branchMatch = (issue.body || "").match(/branch:\s*`?([^\s`]+)`?/i);
    const prMatch = (issue.body || "").match(/pr:\s*#?(\d+)/i);

    return {
      id: String(issue.number || ""),
      title: issue.title || "",
      description: issue.body || "",
      status,
      assignee: issue.assignees?.[0]?.login || null,
      priority: labels.includes("critical")
        ? "critical"
        : labels.includes("high")
          ? "high"
          : null,
      projectId: `${this._owner}/${this._repo}`,
      branchName: branchMatch?.[1] || null,
      prNumber: prMatch?.[1] || null,
      meta: issue,
      backend: "github",
    };
  }
}

// ---------------------------------------------------------------------------
// Jira Adapter (Stub — ready for future implementation)
// ---------------------------------------------------------------------------

class JiraAdapter {
  constructor() {
    this.name = "jira";
    this._baseUrl = process.env.JIRA_BASE_URL || null;
    this._token = process.env.JIRA_API_TOKEN || null;
    this._email = process.env.JIRA_EMAIL || null;
  }

  _notImplemented(method) {
    throw new Error(
      `${TAG} Jira adapter: ${method}() not yet implemented. ` +
        `Set JIRA_BASE_URL, JIRA_API_TOKEN, and JIRA_EMAIL env vars when ready.`,
    );
  }

  async listProjects() {
    this._notImplemented("listProjects");
  }
  async listTasks(_projectId, _filters) {
    this._notImplemented("listTasks");
  }
  async getTask(_taskId) {
    this._notImplemented("getTask");
  }
  async updateTaskStatus(_taskId, _status) {
    this._notImplemented("updateTaskStatus");
  }
  async createTask(_projectId, _taskData) {
    this._notImplemented("createTask");
  }
  async deleteTask(_taskId) {
    this._notImplemented("deleteTask");
  }
}

// ---------------------------------------------------------------------------
// Adapter Registry & Resolution
// ---------------------------------------------------------------------------

const ADAPTERS = {
  vk: () => new VKAdapter(),
  github: () => new GitHubIssuesAdapter(),
  jira: () => new JiraAdapter(),
};

/** @type {Object|null} Cached adapter instance */
let activeAdapter = null;
/** @type {string|null} Cached backend name */
let activeBackendName = null;

/**
 * Resolve which kanban backend to use (synchronous).
 *
 * Resolution order:
 *   1. Runtime override via setKanbanBackend()
 *   2. KANBAN_BACKEND env var
 *   3. codex-monitor.config.json → kanban.backend field
 *   4. Default: "vk"
 *
 * @returns {string}
 */
function resolveBackendName() {
  if (activeBackendName) return activeBackendName;

  // 1. Env var
  const envBackend = (process.env.KANBAN_BACKEND || "").trim().toLowerCase();
  if (envBackend && ADAPTERS[envBackend]) return envBackend;

  // 2. Config file (loadConfig is imported statically — always sync-safe)
  try {
    const config = loadConfig();
    const configBackend = (config?.kanban?.backend || "").toLowerCase();
    if (configBackend && ADAPTERS[configBackend]) return configBackend;
  } catch {
    // Config not available — fall through to default
  }

  // 3. Default
  return "vk";
}

/**
 * Get the active kanban adapter.
 * @returns {VKAdapter|GitHubIssuesAdapter|JiraAdapter} Adapter instance.
 */
export function getKanbanAdapter() {
  const name = resolveBackendName();
  if (activeAdapter && activeBackendName === name) return activeAdapter;
  const factory = ADAPTERS[name];
  if (!factory) throw new Error(`${TAG} unknown kanban backend: ${name}`);
  activeAdapter = factory();
  activeBackendName = name;
  console.log(`${TAG} using ${name} backend`);
  return activeAdapter;
}

/**
 * Switch the kanban backend at runtime.
 * @param {string} name Backend name ("vk", "github", "jira").
 */
export function setKanbanBackend(name) {
  const normalised = (name || "").trim().toLowerCase();
  if (!ADAPTERS[normalised]) {
    throw new Error(
      `${TAG} unknown kanban backend: "${name}". Valid: ${Object.keys(ADAPTERS).join(", ")}`,
    );
  }
  activeBackendName = normalised;
  activeAdapter = null; // Force re-create on next getKanbanAdapter()
  console.log(`${TAG} switched to ${normalised} backend`);
}

/**
 * Get list of available kanban backends.
 * @returns {string[]}
 */
export function getAvailableBackends() {
  return Object.keys(ADAPTERS);
}

/**
 * Get the name of the active backend.
 * @returns {string}
 */
export function getKanbanBackendName() {
  return resolveBackendName();
}

// ---------------------------------------------------------------------------
// Convenience exports: direct task operations via active adapter
// ---------------------------------------------------------------------------

export async function listProjects() {
  return getKanbanAdapter().listProjects();
}

export async function listTasks(projectId, filters) {
  return getKanbanAdapter().listTasks(projectId, filters);
}

export async function getTask(taskId) {
  return getKanbanAdapter().getTask(taskId);
}

export async function updateTaskStatus(taskId, status) {
  return getKanbanAdapter().updateTaskStatus(taskId, status);
}

export async function createTask(projectId, taskData) {
  return getKanbanAdapter().createTask(projectId, taskData);
}

export async function deleteTask(taskId) {
  return getKanbanAdapter().deleteTask(taskId);
}
