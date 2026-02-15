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
  backlog: "backlog",
  "to do": "backlog",
  todo: "todo",
  ready: "ready",
  queued: "ready",
  triaged: "ready",
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
  "in progress": "inprogress",
  review: "inreview",
  resolved: "done",
};

function normaliseStatus(raw) {
  if (!raw) return "ready";
  const key = String(raw).toLowerCase().trim();
  return STATUS_MAP[key] || "ready";
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

      let res;
      try {
        const runtimeFetch = globalThis.fetch;
        if (typeof runtimeFetch !== "function") {
          throw new Error("global fetch is unavailable");
        }
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
        res = await runtimeFetch(url, fetchOpts);
      } catch (err) {
        // Network error, timeout, abort - res is undefined
        throw new Error(
          `VK API ${method} ${path} network error: ${err.message || err}`,
        );
      } finally {
        clearTimeout(timeout);
      }

      if (!res || typeof res.ok === "undefined") {
        throw new Error(
          `VK API ${method} ${path} invalid response object (res=${!!res}, res.ok=${res?.ok})`,
        );
      }

      if (!res.ok) {
        const text =
          typeof res.text === "function" ? await res.text().catch(() => "") : "";
        throw new Error(
          `VK API ${method} ${path} failed: ${res.status} ${text.slice(0, 200)}`,
        );
      }

      const contentTypeRaw =
        typeof res.headers?.get === "function"
          ? res.headers.get("content-type") || res.headers.get("Content-Type")
          : res.headers?.["content-type"] || res.headers?.["Content-Type"] || "";
      const contentType = String(contentTypeRaw || "").toLowerCase();

      if (contentType && !contentType.includes("application/json")) {
        const text =
          typeof res.text === "function" ? await res.text().catch(() => "") : "";
        // VK sometimes mislabels JSON as text/plain in proxy setups.
        if (text) {
          try {
            return JSON.parse(text);
          } catch {
            // Fall through to explicit non-JSON error below.
          }
        }
        throw new Error(
          `VK API ${method} ${path} non-JSON response (${contentType})`,
        );
      }

      try {
        return await res.json();
      } catch (err) {
        throw new Error(
          `VK API ${method} ${path} invalid JSON: ${err.message}`,
        );
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
    // Use /api/tasks?project_id=... (query param style) instead of
    // /api/projects/:id/tasks which gets caught by the SPA catch-all.
    const params = [`project_id=${encodeURIComponent(projectId)}`];
    if (filters.status)
      params.push(`status=${encodeURIComponent(filters.status)}`);
    if (filters.limit) params.push(`limit=${filters.limit}`);
    const url = `/api/tasks?${params.join("&")}`;
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
    return this.updateTask(taskId, { status });
  }

  async updateTask(taskId, patch = {}) {
    const fetchVk = await this._getFetchVk();
    const body = {};
    if (typeof patch.status === "string" && patch.status.trim()) {
      body.status = patch.status.trim();
    }
    if (typeof patch.title === "string") {
      body.title = patch.title;
    }
    if (typeof patch.description === "string") {
      body.description = patch.description;
    }
    if (typeof patch.priority === "string" && patch.priority.trim()) {
      body.priority = patch.priority.trim();
    }
    if (Object.keys(body).length === 0) {
      return this.getTask(taskId);
    }
    const result = await fetchVk(`/api/tasks/${taskId}`, {
      method: "PUT",
      body,
    });
    const task = result?.data || result;
    return this._normaliseTask(task);
  }

  async createTask(projectId, taskData) {
    const fetchVk = await this._getFetchVk();
    // Use /api/tasks with project_id in body instead of
    // /api/projects/:id/tasks which gets caught by the SPA catch-all.
    const result = await fetchVk(`/api/tasks`, {
      method: "POST",
      body: { ...taskData, project_id: projectId },
    });
    const task = result?.data || result;
    return this._normaliseTask(task, projectId);
  }

  async deleteTask(taskId) {
    const fetchVk = await this._getFetchVk();
    await fetchVk(`/api/tasks/${taskId}`, { method: "DELETE" });
    return true;
  }

  async addComment(_taskId, _body) {
    return false; // VK backend doesn't support issue comments
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
    const config = loadConfig();
    const slug =
      process.env.GITHUB_REPOSITORY ||
      config?.repoSlug ||
      "virtengine/virtengine";
    const [slugOwner, slugRepo] = String(slug).split("/", 2);
    this._owner = process.env.GITHUB_REPO_OWNER || slugOwner || "virtengine";
    this._repo = process.env.GITHUB_REPO_NAME || slugRepo || "virtengine";
    const projectConfig = config?.kanban?.githubProject || {};
    const genericProjectId =
      process.env.KANBAN_PROJECT_ID || config?.kanban?.projectId || null;

    this._projectOwner =
      process.env.GITHUB_PROJECT_OWNER ||
      projectConfig.owner ||
      this._owner ||
      null;

    const configuredProjectNumber =
      process.env.GITHUB_PROJECT_NUMBER || projectConfig.number || null;
    this._projectNumber =
      configuredProjectNumber != null && configuredProjectNumber !== ""
        ? Number(configuredProjectNumber)
        : genericProjectId && /^\d+$/.test(String(genericProjectId))
          ? Number(genericProjectId)
          : null;

    this._projectId =
      process.env.GITHUB_PROJECT_ID ||
      projectConfig.id ||
      (genericProjectId && /^PVT_/i.test(String(genericProjectId))
        ? String(genericProjectId)
        : null);

    this._projectStatusFieldName =
      process.env.GITHUB_PROJECT_STATUS_FIELD ||
      projectConfig.statusFieldName ||
      "Status";
    this._todoAssigneeMode = (
      process.env.GITHUB_TODO_ASSIGNEE_MODE || "open-or-self"
    )
      .trim()
      .toLowerCase();
    this._autoAssignOnStart =
      String(process.env.GITHUB_AUTO_ASSIGN_ON_START || "true")
        .trim()
        .toLowerCase() !== "false";
    this._ghLogin = null;
    this._ghLoginLoaded = false;
    this._projectContext = null;
    this._projectContextLoaded = false;
  }

  /** Execute a gh CLI command and return parsed JSON */
  async _gh(args, options = {}) {
    const { parseJson = true } = options;
    const { execFile } = await import("node:child_process");
    const { promisify } = await import("node:util");
    const execFileAsync = promisify(execFile);
    try {
      const { stdout } = await execFileAsync("gh", args, {
        maxBuffer: 10 * 1024 * 1024,
        timeout: 30_000,
      });
      const text = String(stdout || "").trim();
      if (!parseJson) return text;
      if (!text) return null;
      return JSON.parse(text);
    } catch (err) {
      throw new Error(`gh CLI failed: ${err.message}`);
    }
  }

  async _ghGraphql(query, variables = {}) {
    const args = ["api", "graphql", "-f", `query=${query}`];
    for (const [key, value] of Object.entries(variables || {})) {
      if (value === undefined || value === null || value === "") continue;
      if (typeof value === "number" && Number.isFinite(value)) {
        args.push("-F", `${key}=${value}`);
      } else {
        args.push("-f", `${key}=${String(value)}`);
      }
    }
    const result = await this._gh(args);
    if (result?.errors?.length) {
      throw new Error(
        `gh GraphQL failed: ${result.errors.map((e) => e?.message || String(e)).join("; ")}`,
      );
    }
    return result?.data || null;
  }

  async _getGhLogin() {
    if (this._ghLoginLoaded) return this._ghLogin;
    this._ghLoginLoaded = true;
    try {
      const login = await this._gh(["api", "user", "--jq", ".login"], {
        parseJson: false,
      });
      this._ghLogin = String(login || "").trim().toLowerCase() || null;
    } catch {
      this._ghLogin = null;
    }
    return this._ghLogin;
  }

  async _filterDispatchableTasksByAssignee(tasks, filters = {}) {
    if (!Array.isArray(tasks) || tasks.length === 0) return tasks || [];
    const wantedStatus = filters?.status ? normaliseStatus(filters.status) : null;
    if (!["todo", "ready"].includes(wantedStatus || "")) return tasks;
    if (!["open-or-self", "self-only"].includes(this._todoAssigneeMode)) {
      return tasks;
    }
    const login = await this._getGhLogin();
    if (!login) return tasks;

    return tasks.filter((task) => {
      const assignee = String(task?.assignee || "")
        .trim()
        .toLowerCase();
      if (!assignee) return this._todoAssigneeMode === "open-or-self";
      return assignee === login;
    });
  }

  async _maybeAutoAssignIssue(issueNumber, status) {
    const normalized = normaliseStatus(status);
    if (!this._autoAssignOnStart || normalized !== "inprogress") return;
    try {
      await this._gh(
        [
          "issue",
          "edit",
          String(issueNumber),
          "--repo",
          `${this._owner}/${this._repo}`,
          "--add-assignee",
          "@me",
        ],
        { parseJson: false },
      );
    } catch {
      // Non-fatal: some repos restrict assignment permissions
    }
  }

  _hasProjectConfig() {
    return (
      Number.isFinite(this._projectNumber) ||
      Boolean(this._projectId && this._projectOwner)
    );
  }

  _extractProjectFieldMeta(project) {
    const fields = project?.fields?.nodes || [];
    const statusField = fields.find(
      (field) =>
        String(field?.name || "").toLowerCase() ===
        String(this._projectStatusFieldName || "status").toLowerCase(),
    );
    const optionByStatus = new Map();
    const options = Array.isArray(statusField?.options) ? statusField.options : [];
    for (const option of options) {
      if (!option?.id || !option?.name) continue;
      const direct = normaliseStatus(option.name);
      optionByStatus.set(direct, option.id);

      const key = String(option.name || "")
        .trim()
        .toLowerCase()
        .replace(/[\s_-]+/g, " ");
      if (key.includes("backlog")) {
        optionByStatus.set("backlog", option.id);
      }
      if (key.includes("ready") || key.includes("queue")) {
        optionByStatus.set("ready", option.id);
      }
      if (key.includes("to do") || key.includes("backlog")) {
        optionByStatus.set("todo", option.id);
      }
      if (key.includes("review")) {
        optionByStatus.set("inreview", option.id);
      }
      if (key.includes("progress") || key.includes("doing")) {
        optionByStatus.set("inprogress", option.id);
      }
      if (key.includes("done") || key.includes("complete") || key.includes("merged")) {
        optionByStatus.set("done", option.id);
      }
      if (
        key.includes("cancel") ||
        key.includes("not planned") ||
        key.includes("wontfix")
      ) {
        optionByStatus.set("cancelled", option.id);
      }
    }

    return {
      statusFieldId: statusField?.id || null,
      optionByStatus,
    };
  }

  async _loadProjectContext() {
    if (!this._hasProjectConfig()) return null;
    if (this._projectContextLoaded) return this._projectContext;

    let project = null;
    if (Number.isFinite(this._projectNumber) && this._projectOwner) {
      const queryByNumber = `
        query($owner: String!, $number: Int!) {
          organization(login: $owner) {
            projectV2(number: $number) {
              id
              number
              title
              fields(first: 50) {
                nodes {
                  ... on ProjectV2FieldCommon { id name }
                  ... on ProjectV2SingleSelectField {
                    id
                    name
                    options { id name }
                  }
                }
              }
            }
          }
          user(login: $owner) {
            projectV2(number: $number) {
              id
              number
              title
              fields(first: 50) {
                nodes {
                  ... on ProjectV2FieldCommon { id name }
                  ... on ProjectV2SingleSelectField {
                    id
                    name
                    options { id name }
                  }
                }
              }
            }
          }
        }
      `;
      const data = await this._ghGraphql(queryByNumber, {
        owner: this._projectOwner,
        number: this._projectNumber,
      });
      project = data?.organization?.projectV2 || data?.user?.projectV2 || null;
    } else if (this._projectId) {
      const queryById = `
        query($projectId: ID!) {
          node(id: $projectId) {
            ... on ProjectV2 {
              id
              number
              title
              owner {
                ... on Organization { login }
                ... on User { login }
              }
              fields(first: 50) {
                nodes {
                  ... on ProjectV2FieldCommon { id name }
                  ... on ProjectV2SingleSelectField {
                    id
                    name
                    options { id name }
                  }
                }
              }
            }
          }
        }
      `;
      const data = await this._ghGraphql(queryById, {
        projectId: this._projectId,
      });
      project = data?.node || null;
      if (project?.owner?.login && !this._projectOwner) {
        this._projectOwner = project.owner.login;
      }
      if (Number.isFinite(project?.number) && !Number.isFinite(this._projectNumber)) {
        this._projectNumber = Number(project.number);
      }
    }

    if (!project?.id || !Number.isFinite(Number(project?.number))) {
      this._projectContextLoaded = true;
      this._projectContext = null;
      return null;
    }

    const fieldMeta = this._extractProjectFieldMeta(project);
    this._projectContext = {
      id: project.id,
      number: Number(project.number),
      owner: this._projectOwner,
      title: project.title || `Project ${project.number}`,
      statusFieldId: fieldMeta.statusFieldId,
      optionByStatus: fieldMeta.optionByStatus,
    };
    this._projectContextLoaded = true;
    return this._projectContext;
  }

  _extractProjectItemStatus(item) {
    const fieldValues =
      item?.fieldValues ||
      item?.field_values ||
      item?.fields ||
      item?.fieldValueByName ||
      [];
    const values = Array.isArray(fieldValues)
      ? fieldValues
      : Object.values(fieldValues || {});
    const wantedField = String(this._projectStatusFieldName || "status").toLowerCase();
    for (const value of values) {
      const fieldName = String(value?.field?.name || value?.name || "").toLowerCase();
      if (fieldName !== wantedField && fieldName !== "status") continue;
      const candidate =
        value?.optionName ||
        value?.name ||
        value?.value ||
        value?.text ||
        value?.fieldValueName ||
        value?.label;
      if (candidate) return candidate;
    }

    const mapFields = [
      item?.status,
      item?.statusName,
      item?.status_name,
      item?.projectStatus,
      item?.project_status,
    ];
    for (const candidate of mapFields) {
      if (candidate) return candidate;
    }
    return null;
  }

  _normaliseProjectItem(item, projectContext) {
    const content = item?.content || item?.issue || item?.task || item || {};
    const number = content?.number || item?.number || null;
    if (!number) return null;
    const statusRaw = this._extractProjectItemStatus(item);
    const labels = (content?.labels || item?.labels || []).map((label) =>
      typeof label === "string" ? { name: label } : label,
    );

    return this._normaliseIssue({
      number,
      title: content?.title || item?.title || "",
      body: content?.body || content?.description || item?.body || "",
      state: content?.state || item?.state || "open",
      url: content?.url || item?.url || null,
      assignees: content?.assignees || item?.assignees || [],
      labels,
      status: statusRaw,
      project_status: statusRaw,
      projectId: projectContext?.id,
      projectItemId: item?.id || item?.itemId || null,
    });
  }

  async _listProjectItems(filters = {}) {
    const project = await this._loadProjectContext();
    if (!project?.owner || !Number.isFinite(project?.number)) return [];
    const limit =
      Number(filters.limit || process.env.GITHUB_PROJECT_ITEMS_LIMIT || 1000) ||
      1000;
    const args = [
      "project",
      "item-list",
      String(project.number),
      "--owner",
      String(project.owner),
      "--format",
      "json",
      "--limit",
      String(limit),
    ];
    const items = await this._gh(args);
    const normalized = (Array.isArray(items) ? items : [])
      .map((item) => this._normaliseProjectItem(item, project))
      .filter(Boolean);

    if (filters.status) {
      const wanted = normaliseStatus(filters.status);
      return normalized.filter((task) => task.status === wanted);
    }
    return normalized;
  }

  async _resolveProjectItemIdForIssue(issueNumber) {
    const projectItems = await this._listProjectItems({
      limit: Number(process.env.GITHUB_PROJECT_ITEMS_LIMIT || 1000) || 1000,
    });
    const wanted = String(issueNumber).replace(/^#/, "");
    const match = projectItems.find((task) => String(task.id) === wanted);
    return match?.meta?.projectItemId || null;
  }

  async _syncProjectStatus(issueNumber, status) {
    const project = await this._loadProjectContext();
    if (!project?.id || !project?.statusFieldId) return;

    const normalized = normaliseStatus(status);
    const optionId = project.optionByStatus?.get(normalized) || null;
    if (!optionId) {
      console.warn(
        `${TAG} github project status option missing for "${normalized}" in project ${project.owner}/${project.number}`,
      );
      return;
    }

    const itemId = await this._resolveProjectItemIdForIssue(issueNumber);
    if (!itemId) {
      console.warn(
        `${TAG} github project item not found for issue #${issueNumber} in project ${project.owner}/${project.number}`,
      );
      return;
    }

    const mutation = `
      mutation($projectId: ID!, $itemId: ID!, $fieldId: ID!, $optionId: String!) {
        updateProjectV2ItemFieldValue(
          input: {
            projectId: $projectId
            itemId: $itemId
            fieldId: $fieldId
            value: { singleSelectOptionId: $optionId }
          }
        ) {
          projectV2Item { id }
        }
      }
    `;
    await this._ghGraphql(mutation, {
      projectId: project.id,
      itemId,
      fieldId: project.statusFieldId,
      optionId,
    });
  }

  async _addIssueToProject(issueUrl) {
    const project = await this._loadProjectContext();
    if (!project?.owner || !Number.isFinite(project?.number) || !issueUrl) {
      return;
    }
    try {
      await this._gh(
        [
          "project",
          "item-add",
          String(project.number),
          "--owner",
          String(project.owner),
          "--url",
          String(issueUrl),
        ],
        { parseJson: false },
      );
    } catch (err) {
      console.warn(
        `${TAG} failed to add issue to github project ${project.owner}/${project.number}: ${err.message}`,
      );
    }
  }

  async listProjects() {
    const project = await this._loadProjectContext().catch((err) => {
      console.warn(`${TAG} github project load failed: ${err.message}`);
      return null;
    });
    if (project) {
      return [
        {
          id: project.id,
          name: project.title,
          meta: {
            owner: project.owner,
            number: project.number,
            statusFieldId: project.statusFieldId,
          },
          backend: "github",
        },
      ];
    }

    // Fallback: repo as a pseudo-project
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
    if (this._hasProjectConfig()) {
      try {
        const projectTasks = await this._listProjectItems(filters);
        return await this._filterDispatchableTasksByAssignee(projectTasks, filters);
      } catch (err) {
        console.warn(
          `${TAG} github project list failed, falling back to issues: ${err.message}`,
        );
      }
    }

    const limit =
      Number(filters.limit || process.env.GITHUB_ISSUES_LIST_LIMIT || 1000) ||
      1000;
    const args = [
      "issue",
      "list",
      "--repo",
      `${this._owner}/${this._repo}`,
      "--json",
      "number,title,body,state,url,assignees,labels,milestone",
      "--limit",
      String(limit),
    ];
    if (filters.status === "done") {
      args.push("--state", "closed");
    } else if (filters.status && filters.status !== "todo") {
      args.push("--state", "open");
      args.push("--label", filters.status);
    } else {
      args.push("--state", "open");
    }
    const issues = await this._gh(args);
    const normalized = (Array.isArray(issues) ? issues : []).map((i) =>
      this._normaliseIssue(i),
    );
    return await this._filterDispatchableTasksByAssignee(normalized, filters);
  }

  async getTask(issueNumber) {
    const num = String(issueNumber).replace(/^#/, "");
    if (!/^\d+$/.test(num)) {
      throw new Error(
        `GitHub Issues: invalid issue number "${issueNumber}" — expected a numeric ID, got a UUID or non-numeric string`,
      );
    }
    const issue = await this._gh([
      "issue",
      "view",
      num,
      "--repo",
      `${this._owner}/${this._repo}`,
      "--json",
      "number,title,body,state,url,assignees,labels,milestone",
    ]);
    return this._normaliseIssue(issue);
  }

  async updateTaskStatus(issueNumber, status) {
    const num = String(issueNumber).replace(/^#/, "");
    if (!/^\d+$/.test(num)) {
      throw new Error(
        `GitHub Issues: invalid issue number "${issueNumber}" — expected a numeric ID, got a UUID or non-numeric string`,
      );
    }
    const normalised = normaliseStatus(status);
    if (normalised === "done" || normalised === "cancelled") {
      const closeArgs = [
        "issue",
        "close",
        num,
        "--repo",
        `${this._owner}/${this._repo}`,
      ];
      if (normalised === "cancelled") {
        closeArgs.push("--reason", "not planned");
      }
      await this._gh(closeArgs, { parseJson: false });
    } else {
      await this._maybeAutoAssignIssue(num, normalised);
      await this._gh(
        ["issue", "reopen", num, "--repo", `${this._owner}/${this._repo}`],
        { parseJson: false },
      );

      // Keep status labels in sync for open issues.
      const labelByStatus = {
        ready: "ready",
        inprogress: "inprogress",
        inreview: "inreview",
        blocked: "blocked",
      };
      const nextLabel = labelByStatus[normalised] || null;
      const statusLabels = [
        "inprogress",
        "in-progress",
        "inreview",
        "in-review",
        "blocked",
      ];
      const removeLabels = statusLabels.filter((label) => label !== nextLabel);
      const editArgs = [
        "issue",
        "edit",
        num,
        "--repo",
        `${this._owner}/${this._repo}`,
      ];
      if (nextLabel) {
        editArgs.push("--add-label", nextLabel);
      }
      for (const label of removeLabels) {
        editArgs.push("--remove-label", label);
      }
      try {
        await this._gh(editArgs, { parseJson: false });
      } catch {
        // Label might not exist — non-critical
      }
    }

    if (this._hasProjectConfig()) {
      try {
        await this._syncProjectStatus(num, status);
      } catch (err) {
        console.warn(
          `${TAG} github project status sync failed for issue #${num}: ${err.message}`,
        );
      }
    }

    return this.getTask(issueNumber);
  }

  async updateTask(issueNumber, patch = {}) {
    const num = String(issueNumber).replace(/^#/, "");
    if (!/^\d+$/.test(num)) {
      throw new Error(
        `GitHub Issues: invalid issue number "${issueNumber}" — expected a numeric ID, got a UUID or non-numeric string`,
      );
    }
    const editArgs = ["issue", "edit", num, "--repo", `${this._owner}/${this._repo}`];
    let hasEditArgs = false;
    if (typeof patch.title === "string") {
      editArgs.push("--title", patch.title);
      hasEditArgs = true;
    }
    if (typeof patch.description === "string") {
      editArgs.push("--body", patch.description);
      hasEditArgs = true;
    }
    if (hasEditArgs) {
      await this._gh(editArgs, { parseJson: false });
    }
    if (typeof patch.status === "string" && patch.status.trim()) {
      return this.updateTaskStatus(num, patch.status.trim());
    }
    return this.getTask(num);
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
    const result = await this._gh(args, { parseJson: false });
    const issueUrl = String(result || "").match(/https?:\/\/\S+/)?.[0] || "";
    const issueNum = issueUrl.match(/\/issues\/(\d+)/)?.[1] || null;

    if (issueUrl && this._hasProjectConfig()) {
      await this._addIssueToProject(issueUrl);
      if (taskData?.status && issueNum) {
        try {
          await this._syncProjectStatus(issueNum, taskData.status);
        } catch {
          // Non-critical: issue creation still succeeded
        }
      }
    }

    if (issueNum) {
      return this.getTask(issueNum);
    }
    const numericFallback = String(result || "")
      .trim()
      .match(/^#?(\d+)$/)?.[1];
    if (numericFallback) {
      return this.getTask(numericFallback);
    }
    return { id: issueUrl || String(result || "").trim(), backend: "github" };
  }

  async deleteTask(issueNumber) {
    // GitHub issues can't be deleted — close with "not planned"
    const num = String(issueNumber).replace(/^#/, "");
    if (!/^\d+$/.test(num)) {
      throw new Error(
        `GitHub Issues: invalid issue number "${issueNumber}" — expected a numeric ID`,
      );
    }
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

  async addComment(issueNumber, body) {
    const num = String(issueNumber).replace(/^#/, "");
    if (!/^\d+$/.test(num) || !body) return false;
    try {
      await this._gh(
        [
          "issue",
          "comment",
          num,
          "--repo",
          `${this._owner}/${this._repo}`,
          "--body",
          String(body).slice(0, 65536),
        ],
        { parseJson: false },
      );
      return true;
    } catch (err) {
      console.warn(
        `[kanban] failed to comment on issue #${num}: ${err.message}`,
      );
      return false;
    }
  }

  _normaliseIssue(issue) {
    if (!issue) return null;
    const labels = (issue.labels || []).map((l) =>
      typeof l === "string" ? l : l.name,
    );
    const labelSet = new Set(
      labels.map((l) =>
        String(l || "")
          .trim()
          .toLowerCase(),
      ),
    );
    let status = "todo";
    if (issue.state === "closed" || issue.state === "CLOSED") {
      status = "done";
    } else if (issue.project_status) {
      status = normaliseStatus(issue.project_status);
    } else if (labelSet.has("inprogress") || labelSet.has("in-progress")) {
      status = "inprogress";
    } else if (labelSet.has("inreview") || labelSet.has("in-review")) {
      status = "inreview";
    } else if (labelSet.has("blocked")) {
      status = "blocked";
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
      priority: labelSet.has("critical")
        ? "critical"
        : labelSet.has("high")
          ? "high"
          : null,
      projectId: issue.projectId || `${this._owner}/${this._repo}`,
      branchName: branchMatch?.[1] || null,
      prNumber: prMatch?.[1] || null,
      meta: {
        ...issue,
        task_url: issue.url || null,
        projectItemId: issue.projectItemId || null,
      },
      taskUrl: issue.url || null,
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
  async updateTask(_taskId, _patch) {
    this._notImplemented("updateTask");
  }
  async createTask(_projectId, _taskData) {
    this._notImplemented("createTask");
  }
  async deleteTask(_taskId) {
    this._notImplemented("deleteTask");
  }

  async addComment(_taskId, _body) {
    return false; // Jira comments not yet implemented
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

export async function updateTask(taskId, patch) {
  const adapter = getKanbanAdapter();
  if (typeof adapter.updateTask === "function") {
    return adapter.updateTask(taskId, patch);
  }
  if (patch?.status) {
    return adapter.updateTaskStatus(taskId, patch.status);
  }
  return adapter.getTask(taskId);
}

export async function createTask(projectId, taskData) {
  return getKanbanAdapter().createTask(projectId, taskData);
}

export async function deleteTask(taskId) {
  return getKanbanAdapter().deleteTask(taskId);
}

export async function addComment(taskId, body) {
  return getKanbanAdapter().addComment(taskId, body);
}
