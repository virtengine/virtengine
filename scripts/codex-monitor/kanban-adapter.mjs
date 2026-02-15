/**
 * kanban-adapter.mjs — Unified Kanban Board Abstraction
 *
 * Provides a common interface over multiple task-tracking backends:
 *   - Vibe-Kanban (VK)       — default, full-featured
 *   - GitHub Issues           — native GitHub integration with shared state persistence
 *   - Jira                    — enterprise project management (scaffolded, see JIRA_INTEGRATION.md)
 *
 * This module handles TASK LIFECYCLE (tracking, status, metadata) only.
 * Code execution is handled separately by agent-pool.mjs.
 *
 * Configuration:
 *   - `KANBAN_BACKEND` env var: "vk" | "github" | "jira" (default: "vk")
 *   - `codex-monitor.config.json` → `kanban.backend` field
 *
 * EXPORTS:
 *   getKanbanAdapter()                       → Returns the configured adapter instance
 *   setKanbanBackend(name)                   → Switch backend at runtime
 *   getAvailableBackends()                   → List available backends
 *   getKanbanBackendName()                   → Get active backend name
 *   listProjects()                           → Convenience: adapter.listProjects()
 *   listTasks(projectId, f?)                 → Convenience: adapter.listTasks()
 *   getTask(taskId)                          → Convenience: adapter.getTask()
 *   updateTaskStatus(id, s, opts?)           → Convenience: adapter.updateTaskStatus()
 *   createTask(projId, data)                 → Convenience: adapter.createTask()
 *   deleteTask(taskId)                       → Convenience: adapter.deleteTask()
 *   addComment(taskId, body)                 → Convenience: adapter.addComment()
 *   persistSharedStateToIssue(id, state)     → GitHub/Jira: persist agent state to issue
 *   readSharedStateFromIssue(id)             → GitHub/Jira: read agent state from issue
 *   markTaskIgnored(id, reason)              → GitHub/Jira: mark task as ignored
 *
 * Each adapter implements the KanbanAdapter interface:
 *   - listTasks(projectId, filters?)         → Task[]
 *   - getTask(taskId)                        → Task
 *   - updateTaskStatus(taskId, status, opts?)→ Task
 *   - createTask(projectId, task)            → Task
 *   - deleteTask(taskId)                     → boolean
 *   - listProjects()                         → Project[]
 *   - addComment(taskId, body)               → boolean
 *
 * GitHub adapter implements shared state methods:
 *   - persistSharedStateToIssue(num, state)  → boolean
 *   - readSharedStateFromIssue(num)          → SharedState|null
 *   - markTaskIgnored(num, reason)           → boolean
 *
 * Jira adapter has scaffolded shared state methods (throw "not implemented"):
 *   - persistSharedStateToIssue(key, state)  → boolean (throws)
 *   - readSharedStateFromIssue(key)          → SharedState|null (throws)
 *   - markTaskIgnored(key, reason)           → boolean (throws)
 *   See JIRA_INTEGRATION.md for implementation guide
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

function parseBooleanEnv(value, fallback = false) {
  if (value == null || value === "") return fallback;
  const key = String(value).trim().toLowerCase();
  if (["1", "true", "yes", "on"].includes(key)) return true;
  if (["0", "false", "no", "off"].includes(key)) return false;
  return fallback;
}

function normalizeLabels(raw) {
  const values = Array.isArray(raw)
    ? raw
    : String(raw || "")
        .split(",")
        .map((entry) => entry.trim())
        .filter(Boolean);
  const seen = new Set();
  const labels = [];
  for (const value of values) {
    const normalized = String(value || "")
      .trim()
      .toLowerCase();
    if (!normalized || seen.has(normalized)) continue;
    seen.add(normalized);
    labels.push(normalized);
  }
  return labels;
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
          typeof res.text === "function"
            ? await res.text().catch(() => "")
            : "";
        throw new Error(
          `VK API ${method} ${path} failed: ${res.status} ${text.slice(0, 200)}`,
        );
      }

      const contentTypeRaw =
        typeof res.headers?.get === "function"
          ? res.headers.get("content-type") || res.headers.get("Content-Type")
          : res.headers?.["content-type"] ||
            res.headers?.["Content-Type"] ||
            "";
      const contentType = String(contentTypeRaw || "").toLowerCase();

      if (contentType && !contentType.includes("application/json")) {
        const text =
          typeof res.text === "function"
            ? await res.text().catch(() => "")
            : "";
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

/**
 * @typedef {Object} SharedState
 * @property {string} ownerId - Workstation/agent identifier (e.g., "workstation-123/agent-456")
 * @property {string} attemptToken - Unique UUID for this claim attempt
 * @property {string} attemptStarted - ISO 8601 timestamp of claim start
 * @property {string} heartbeat - ISO 8601 timestamp of last heartbeat
 * @property {string} status - Current status: "claimed"|"working"|"stale"
 * @property {number} retryCount - Number of retry attempts
 */

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

    // Codex-monitor label scheme
    this._codexLabels = {
      claimed: "codex:claimed",
      working: "codex:working",
      stale: "codex:stale",
      ignore: "codex:ignore",
    };

    this._canonicalTaskLabel =
      process.env.CODEX_MONITOR_TASK_LABEL || "codex-monitor";
    this._taskScopeLabels = normalizeLabels(
      process.env.CODEX_MONITOR_TASK_LABELS ||
        `${this._canonicalTaskLabel},codex-mointor`,
    );
    this._enforceTaskLabel = parseBooleanEnv(
      process.env.CODEX_MONITOR_ENFORCE_TASK_LABEL,
      true,
    );

    this._autoAssignCreator = parseBooleanEnv(
      process.env.GITHUB_AUTO_ASSIGN_CREATOR,
      true,
    );
    this._defaultAssignee =
      process.env.GITHUB_DEFAULT_ASSIGNEE || this._owner || null;

    this._projectMode = String(process.env.GITHUB_PROJECT_MODE || "issues")
      .trim()
      .toLowerCase();
    this._projectOwner = process.env.GITHUB_PROJECT_OWNER || this._owner;
    this._projectTitle =
      process.env.GITHUB_PROJECT_TITLE ||
      process.env.PROJECT_NAME ||
      "Codex-Monitor";
    this._projectNumber =
      process.env.GITHUB_PROJECT_NUMBER ||
      process.env.GITHUB_PROJECT_ID ||
      null;
    this._cachedProjectNumber = this._projectNumber;
    this._projectFieldsCache = null;
    this._projectFieldsCacheTime = 0;
    this._projectFieldsCacheTTL = 300_000; // 5 minutes
  }

  /**
   * Get project fields with caching.
   * Returns status field ID and options for project board.
   * @private
   * @param {string} projectNumber - GitHub project number
   * @returns {Promise<{statusFieldId: string, statusOptions: Array<{id: string, name: string}>}|null>}
   */
  async _getProjectFields(projectNumber) {
    if (!projectNumber) return null;

    // Return cached value if still valid
    const now = Date.now();
    if (
      this._projectFieldsCache &&
      now - this._projectFieldsCacheTime < this._projectFieldsCacheTTL
    ) {
      return this._projectFieldsCache;
    }

    try {
      const owner = String(this._projectOwner || this._owner).trim();
      const fields = await this._gh([
        "project",
        "field-list",
        String(projectNumber),
        "--owner",
        owner,
        "--format",
        "json",
      ]);

      if (!Array.isArray(fields)) {
        console.warn(
          `${TAG} project field-list returned non-array for project ${projectNumber}`,
        );
        return null;
      }

      // Find the Status field
      const statusField = fields.find(
        (f) =>
          f.name === "Status" &&
          (f.type === "SINGLE_SELECT" || f.data_type === "SINGLE_SELECT"),
      );

      if (!statusField) {
        console.warn(
          `${TAG} no Status field found in project ${projectNumber}`,
        );
        return null;
      }

      // Extract options
      const statusOptions = (statusField.options || []).map((opt) => ({
        id: opt.id,
        name: opt.name,
      }));

      const result = {
        statusFieldId: statusField.id,
        statusOptions,
      };

      // Cache the result
      this._projectFieldsCache = result;
      this._projectFieldsCacheTime = now;

      return result;
    } catch (err) {
      console.warn(
        `${TAG} failed to fetch project fields for ${projectNumber}: ${err.message}`,
      );
      return null;
    }
  }

  /**
   * List tasks from a GitHub Project board.
   * Fetches project items, extracts issue URLs, and normalizes them.
   * @private
   * @param {string} projectNumber - GitHub project number
   * @returns {Promise<KanbanTask[]>}
   */
  async _listTasksFromProject(projectNumber) {
    if (!projectNumber) return [];

    try {
      const owner = String(this._projectOwner || this._owner).trim();
      const items = await this._gh([
        "project",
        "item-list",
        String(projectNumber),
        "--owner",
        owner,
        "--format",
        "json",
      ]);

      if (!Array.isArray(items)) {
        console.warn(
          `${TAG} project item-list returned non-array for project ${projectNumber}`,
        );
        return [];
      }

      // Extract issue URLs and fetch issue details
      const tasks = [];
      for (const item of items) {
        const issueUrl = item.content?.url;
        if (!issueUrl || !issueUrl.includes("/issues/")) {
          continue; // Skip non-issue items (e.g., draft issues, PRs)
        }

        // Extract issue number from URL
        const issueMatch = issueUrl.match(/\/issues\/(\d+)/);
        if (!issueMatch) continue;

        const issueNumber = issueMatch[1];

        try {
          // Fetch full issue details
          const issue = await this._gh([
            "issue",
            "view",
            issueNumber,
            "--repo",
            `${this._owner}/${this._repo}`,
            "--json",
            "number,title,body,state,url,assignees,labels,milestone,comments",
          ]);

          const task = this._normaliseIssue(issue);
          if (task) {
            // Add project metadata
            task.meta.projectNumber = projectNumber;
            task.meta.projectItemId = item.id;
            task.meta.projectStatus = item.status || null;

            tasks.push(task);
          }
        } catch (err) {
          console.warn(
            `${TAG} failed to fetch issue ${issueNumber} from project: ${err.message}`,
          );
        }
      }

      return tasks;
    } catch (err) {
      console.warn(
        `${TAG} failed to list tasks from project ${projectNumber}: ${err.message}`,
      );
      return [];
    }
  }

  /**
   * Sync task status to GitHub Project board.
   * Maps codex status to project Status field and updates via GraphQL.
   * @private
   * @param {string} issueUrl - Full GitHub issue URL
   * @param {string} projectNumber - GitHub project number
   * @param {string} status - Normalized status (todo/inprogress/inreview/done)
   * @returns {Promise<boolean>}
   */
  async _syncStatusToProject(issueUrl, projectNumber, status) {
    if (!issueUrl || !projectNumber || !status) return false;

    try {
      const owner = String(this._projectOwner || this._owner).trim();

      // Get project fields
      const fields = await this._getProjectFields(projectNumber);
      if (!fields || !fields.statusFieldId) {
        console.warn(`${TAG} cannot sync to project: no status field found`);
        return false;
      }

      // Map codex status to project status option
      const statusMapping = {
        todo: "Todo",
        inprogress: "In Progress",
        inreview: "In Review",
        done: "Done",
        cancelled: "Done",
      };

      const targetStatusName = statusMapping[status] || "Todo";
      const statusOption = fields.statusOptions.find(
        (opt) => opt.name.toLowerCase() === targetStatusName.toLowerCase(),
      );

      if (!statusOption) {
        console.warn(
          `${TAG} no matching project status for "${targetStatusName}"`,
        );
        return false;
      }

      // First, ensure issue is in the project
      try {
        await this._gh(
          [
            "project",
            "item-add",
            String(projectNumber),
            "--owner",
            owner,
            "--url",
            issueUrl,
          ],
          { parseJson: false },
        );
      } catch (err) {
        const text = String(err?.message || err).toLowerCase();
        if (!text.includes("already") && !text.includes("item")) {
          throw err;
        }
        // Item already in project, continue
      }

      // Get project and item IDs via GraphQL
      const projectIdQuery = `
        query {
          user(login: "${owner}") {
            projectV2(number: ${projectNumber}) {
              id
            }
          }
          organization(login: "${owner}") {
            projectV2(number: ${projectNumber}) {
              id
            }
          }
        }
      `;

      let projectId = null;
      try {
        const projectData = await this._gh([
          "api",
          "graphql",
          "-f",
          `query=${projectIdQuery}`,
        ]);
        projectId =
          projectData?.data?.user?.projectV2?.id ||
          projectData?.data?.organization?.projectV2?.id;
      } catch (err) {
        console.warn(`${TAG} failed to get project ID: ${err.message}`);
        return false;
      }

      if (!projectId) {
        console.warn(
          `${TAG} could not resolve project ID for ${owner}/${projectNumber}`,
        );
        return false;
      }

      // Get project item ID for this issue
      const itemQuery = `
        query {
          resource(url: "${issueUrl}") {
            ... on Issue {
              projectItems(first: 10) {
                nodes {
                  id
                  project {
                    id
                  }
                }
              }
            }
          }
        }
      `;

      let itemId = null;
      try {
        const itemData = await this._gh([
          "api",
          "graphql",
          "-f",
          `query=${itemQuery}`,
        ]);
        const items = itemData?.data?.resource?.projectItems?.nodes || [];
        const matchingItem = items.find(
          (item) => item.project?.id === projectId,
        );
        itemId = matchingItem?.id;
      } catch (err) {
        console.warn(`${TAG} failed to get project item ID: ${err.message}`);
        return false;
      }

      if (!itemId) {
        console.warn(
          `${TAG} issue not found in project ${owner}/${projectNumber}`,
        );
        return false;
      }

      // Update project item field value
      const mutation = `
        mutation {
          updateProjectV2ItemFieldValue(
            input: {
              projectId: "${projectId}",
              itemId: "${itemId}",
              fieldId: "${fields.statusFieldId}",
              value: {singleSelectOptionId: "${statusOption.id}"}
            }
          ) {
            projectV2Item {
              id
            }
          }
        }
      `;

      await this._gh(["api", "graphql", "-f", `query=${mutation}`]);

      console.log(
        `${TAG} synced issue ${issueUrl} to project status: ${targetStatusName}`,
      );
      return true;
    } catch (err) {
      console.warn(`${TAG} failed to sync status to project: ${err.message}`);
      return false;
    }
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
    // If project mode is enabled, read from project board
    if (this._projectMode === "kanban" && this._projectNumber) {
      const projectNumber = await this._resolveProjectNumber();
      if (projectNumber) {
        try {
          const tasks = await this._listTasksFromProject(projectNumber);

          // Apply filters
          let filtered = tasks;

          if (this._enforceTaskLabel) {
            filtered = filtered.filter((task) =>
              this._isTaskScopedForCodex(task),
            );
          }

          if (filters.status) {
            const normalizedFilter = normaliseStatus(filters.status);
            filtered = filtered.filter(
              (task) => task.status === normalizedFilter,
            );
          }

          // Enrich with shared state from comments
          for (const task of filtered) {
            try {
              const sharedState = await this.readSharedStateFromIssue(task.id);
              if (sharedState) {
                task.meta.sharedState = sharedState;
              }
            } catch (err) {
              console.warn(
                `[kanban] failed to read shared state for #${task.id}: ${err.message}`,
              );
            }
          }

          return filtered;
        } catch (err) {
          console.warn(
            `${TAG} failed to list tasks from project, falling back to issues: ${err.message}`,
          );
          // Fall through to regular issue listing
        }
      }
    }

    // Default: list from issues
    const limit =
      Number(filters.limit || process.env.GITHUB_ISSUES_LIST_LIMIT || 1000) ||
      1000;
    const args = [
      "issue",
      "list",
      "--repo",
      `${this._owner}/${this._repo}`,
      "--json",
      "number,title,body,state,url,assignees,labels,milestone,comments",
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
    let normalized = (Array.isArray(issues) ? issues : []).map((i) =>
      this._normaliseIssue(i),
    );

    if (this._enforceTaskLabel) {
      normalized = normalized.filter((task) =>
        this._isTaskScopedForCodex(task),
      );
    }

    // Enrich with shared state from comments
    for (const task of normalized) {
      try {
        const sharedState = await this.readSharedStateFromIssue(task.id);
        if (sharedState) {
          task.meta.sharedState = sharedState;
        }
      } catch (err) {
        // Non-critical - continue without shared state
        console.warn(
          `[kanban] failed to read shared state for #${task.id}: ${err.message}`,
        );
      }
    }

    return normalized;
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
      "number,title,body,state,url,assignees,labels,milestone,comments",
    ]);
    const task = this._normaliseIssue(issue);

    // Enrich with shared state from comments
    try {
      const sharedState = await this.readSharedStateFromIssue(num);
      if (sharedState) {
        task.meta.sharedState = sharedState;
      }
    } catch (err) {
      // Non-critical - continue without shared state
      console.warn(
        `[kanban] failed to read shared state for #${num}: ${err.message}`,
      );
    }

    return task;
  }

  async updateTaskStatus(issueNumber, status, options = {}) {
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
      await this._gh(
        ["issue", "reopen", num, "--repo", `${this._owner}/${this._repo}`],
        { parseJson: false },
      );

      // Keep status labels in sync for open issues.
      const labelByStatus = {
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

    // Optionally sync shared state if provided
    if (options.sharedState) {
      try {
        await this.persistSharedStateToIssue(num, options.sharedState);
      } catch (err) {
        console.warn(
          `[kanban] failed to persist shared state for #${num}: ${err.message}`,
        );
      }
    }

    // Sync to project if configured
    if (this._projectMode === "kanban" && this._projectNumber) {
      const projectNumber = await this._resolveProjectNumber();
      if (projectNumber) {
        const task = await this.getTask(num);
        if (task?.taskUrl) {
          try {
            await this._syncStatusToProject(
              task.taskUrl,
              projectNumber,
              normalised,
            );
          } catch (err) {
            // Log but don't fail - issue update should still succeed
            console.warn(
              `${TAG} failed to sync status to project: ${err.message}`,
            );
          }
        }
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
    const editArgs = [
      "issue",
      "edit",
      num,
      "--repo",
      `${this._owner}/${this._repo}`,
    ];
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
    const requestedLabels = normalizeLabels(taskData.labels || []);
    const labelsToApply = new Set(requestedLabels);
    labelsToApply.add(
      String(this._canonicalTaskLabel || "codex-monitor").toLowerCase(),
    );

    const requestedStatus = normaliseStatus(taskData.status || "todo");
    if (requestedStatus === "inprogress") labelsToApply.add("inprogress");
    if (requestedStatus === "inreview") labelsToApply.add("inreview");
    if (requestedStatus === "blocked") labelsToApply.add("blocked");

    for (const label of labelsToApply) {
      await this._ensureLabelExists(label);
    }

    const assignee =
      taskData.assignee ||
      (this._autoAssignCreator ? await this._resolveDefaultAssignee() : null);

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
    if (assignee) args.push("--assignee", assignee);
    if (labelsToApply.size > 0) {
      for (const label of labelsToApply) {
        args.push("--label", label);
      }
    }
    const result = await this._gh(args, { parseJson: false });
    const issueUrl = String(result || "").match(/https?:\/\/\S+/)?.[0] || "";
    const issueNum = issueUrl.match(/\/issues\/(\d+)/)?.[1] || null;
    if (issueUrl) {
      await this._ensureIssueLinkedToProject(issueUrl);
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

  /**
   * Persist shared state to a GitHub issue using structured comments and labels.
   *
   * Creates or updates a codex-monitor-state comment with JSON state and applies
   * appropriate labels (codex:claimed, codex:working, codex:stale).
   *
   * Error handling: Retries once on failure, logs and continues on second failure.
   *
   * @param {string|number} issueNumber - GitHub issue number
   * @param {SharedState} sharedState - State object to persist
   * @returns {Promise<boolean>} Success status
   *
   * @example
   * await adapter.persistSharedStateToIssue(123, {
   *   ownerId: "workstation-123/agent-456",
   *   attemptToken: "uuid-here",
   *   attemptStarted: "2026-02-14T17:00:00Z",
   *   heartbeat: "2026-02-14T17:30:00Z",
   *   status: "working",
   *   retryCount: 1
   * });
   */
  async persistSharedStateToIssue(issueNumber, sharedState) {
    const num = String(issueNumber).replace(/^#/, "");
    if (!/^\d+$/.test(num)) {
      throw new Error(`Invalid issue number: ${issueNumber}`);
    }

    const attemptWithRetry = async (fn, maxRetries = 1) => {
      for (let attempt = 0; attempt <= maxRetries; attempt++) {
        try {
          return await fn();
        } catch (err) {
          if (attempt === maxRetries) {
            console.error(
              `[kanban] persistSharedStateToIssue #${num} failed after ${maxRetries + 1} attempts: ${err.message}`,
            );
            return false;
          }
          console.warn(
            `[kanban] persistSharedStateToIssue #${num} attempt ${attempt + 1} failed, retrying: ${err.message}`,
          );
          await new Promise((resolve) => setTimeout(resolve, 1000));
        }
      }
    };

    // 1. Update labels based on status
    const labelsSuccess = await attemptWithRetry(async () => {
      const currentLabels = await this._getIssueLabels(num);
      const codexLabels = Object.values(this._codexLabels);
      const otherLabels = currentLabels.filter(
        (label) => !codexLabels.includes(label),
      );

      // Determine new codex label based on status
      let newCodexLabel = null;
      if (sharedState.status === "claimed") {
        newCodexLabel = this._codexLabels.claimed;
      } else if (sharedState.status === "working") {
        newCodexLabel = this._codexLabels.working;
      } else if (sharedState.status === "stale") {
        newCodexLabel = this._codexLabels.stale;
      }

      // Build new label set
      const newLabels = [...otherLabels];
      if (newCodexLabel) {
        newLabels.push(newCodexLabel);
      }

      // Apply labels
      const editArgs = [
        "issue",
        "edit",
        num,
        "--repo",
        `${this._owner}/${this._repo}`,
      ];

      // Remove old codex labels
      for (const label of codexLabels) {
        if (label !== newCodexLabel && currentLabels.includes(label)) {
          editArgs.push("--remove-label", label);
        }
      }

      // Add new codex label
      if (newCodexLabel && !currentLabels.includes(newCodexLabel)) {
        editArgs.push("--add-label", newCodexLabel);
      }

      if (editArgs.length > 6) {
        // Only run if we have label changes
        await this._gh(editArgs, { parseJson: false });
      }
      return true;
    });

    // Short-circuit: if labels failed, skip comment update to avoid hanging
    if (!labelsSuccess) return false;

    // 2. Create/update structured comment
    const commentSuccess = await attemptWithRetry(async () => {
      const comments = await this._getIssueComments(num);
      const stateCommentIndex = comments.findIndex((c) =>
        c.body?.includes("<!-- codex-monitor-state"),
      );

      const [agentId, workstationId] = sharedState.ownerId.split("/").reverse();
      const stateJson = JSON.stringify(sharedState, null, 2);
      const commentBody = `<!-- codex-monitor-state
${stateJson}
-->
**Codex Monitor Status**: Agent \`${agentId}\` on \`${workstationId}\` is ${sharedState.status === "working" ? "working on" : sharedState.status === "claimed" ? "claiming" : "stale for"} this task.
*Last heartbeat: ${sharedState.heartbeat}*`;

      if (stateCommentIndex >= 0) {
        // Update existing comment
        const commentId = comments[stateCommentIndex].id;
        await this._gh(
          [
            "api",
            `/repos/${this._owner}/${this._repo}/issues/comments/${commentId}`,
            "-X",
            "PATCH",
            "-f",
            `body=${commentBody}`,
          ],
          { parseJson: false },
        );
      } else {
        // Create new comment
        await this.addComment(num, commentBody);
      }
      return true;
    });

    return commentSuccess;
  }

  /**
   * Read shared state from a GitHub issue by parsing codex-monitor-state comments.
   *
   * Searches for the latest comment containing the structured state JSON and
   * returns the parsed SharedState object, or null if not found.
   *
   * @param {string|number} issueNumber - GitHub issue number
   * @returns {Promise<SharedState|null>} Parsed shared state or null
   *
   * @example
   * const state = await adapter.readSharedStateFromIssue(123);
   * if (state) {
   *   console.log(`Task claimed by ${state.ownerId}`);
   * }
   */
  async readSharedStateFromIssue(issueNumber) {
    const num = String(issueNumber).replace(/^#/, "");
    if (!/^\d+$/.test(num)) {
      throw new Error(`Invalid issue number: ${issueNumber}`);
    }

    try {
      const comments = await this._getIssueComments(num);
      const stateComment = comments
        .reverse()
        .find((c) => c.body?.includes("<!-- codex-monitor-state"));

      if (!stateComment) {
        return null;
      }

      // Extract JSON from comment
      const match = stateComment.body.match(
        /<!-- codex-monitor-state\s*\n([\s\S]*?)\n-->/,
      );
      if (!match) {
        return null;
      }

      const stateJson = match[1].trim();
      const state = JSON.parse(stateJson);

      // Validate required fields
      if (
        !state.ownerId ||
        !state.attemptToken ||
        !state.attemptStarted ||
        !state.heartbeat ||
        !state.status
      ) {
        console.warn(
          `[kanban] invalid shared state in #${num}: missing required fields`,
        );
        return null;
      }

      return state;
    } catch (err) {
      console.warn(
        `[kanban] failed to read shared state for #${num}: ${err.message}`,
      );
      return null;
    }
  }

  /**
   * Mark a task as ignored by codex-monitor.
   *
   * Adds the `codex:ignore` label and posts a comment explaining why the task
   * is being ignored. This prevents codex-monitor from repeatedly attempting
   * to claim or work on tasks that are not suitable for automation.
   *
   * @param {string|number} issueNumber - GitHub issue number
   * @param {string} reason - Human-readable reason for ignoring
   * @returns {Promise<boolean>} Success status
   *
   * @example
   * await adapter.markTaskIgnored(123, "Task requires manual security review");
   */
  async markTaskIgnored(issueNumber, reason) {
    const num = String(issueNumber).replace(/^#/, "");
    if (!/^\d+$/.test(num)) {
      throw new Error(`Invalid issue number: ${issueNumber}`);
    }

    try {
      // Add codex:ignore label
      await this._gh(
        [
          "issue",
          "edit",
          num,
          "--repo",
          `${this._owner}/${this._repo}`,
          "--add-label",
          this._codexLabels.ignore,
        ],
        { parseJson: false },
      );

      // Add comment explaining why
      const commentBody = `**Codex Monitor**: This task has been marked as ignored.

**Reason**: ${reason}

To re-enable codex-monitor for this task, remove the \`${this._codexLabels.ignore}\` label.`;

      await this.addComment(num, commentBody);

      return true;
    } catch (err) {
      console.error(
        `[kanban] failed to mark task #${num} as ignored: ${err.message}`,
      );
      return false;
    }
  }

  /**
   * Get all labels for an issue.
   * @private
   */
  async _getIssueLabels(issueNumber) {
    const issue = await this._gh([
      "issue",
      "view",
      issueNumber,
      "--repo",
      `${this._owner}/${this._repo}`,
      "--json",
      "labels",
    ]);
    return (issue.labels || []).map((l) =>
      typeof l === "string" ? l : l.name,
    );
  }

  /**
   * Get all comments for an issue.
   * @private
   */
  async _getIssueComments(issueNumber) {
    try {
      const result = await this._gh([
        "api",
        `/repos/${this._owner}/${this._repo}/issues/${issueNumber}/comments`,
        "--jq",
        ".",
      ]);
      return Array.isArray(result) ? result : [];
    } catch (err) {
      console.warn(
        `[kanban] failed to fetch comments for #${issueNumber}: ${err.message}`,
      );
      return [];
    }
  }

  _isTaskScopedForCodex(task) {
    const labels = normalizeLabels(
      (task?.meta?.labels || []).map((entry) =>
        typeof entry === "string" ? entry : entry?.name,
      ),
    );
    if (labels.length === 0) return false;
    return this._taskScopeLabels.some((label) => labels.includes(label));
  }

  async _resolveDefaultAssignee() {
    if (this._defaultAssignee) return this._defaultAssignee;
    try {
      const login = await this._gh(["api", "user", "--jq", ".login"], {
        parseJson: false,
      });
      const normalized = String(login || "").trim();
      if (normalized) {
        this._defaultAssignee = normalized;
      }
    } catch {
      this._defaultAssignee = null;
    }
    return this._defaultAssignee;
  }

  async _ensureLabelExists(label) {
    const normalized = String(label || "").trim();
    if (!normalized) return;
    try {
      await this._gh(
        [
          "api",
          `/repos/${this._owner}/${this._repo}/labels`,
          "-X",
          "POST",
          "-f",
          `name=${normalized}`,
          "-f",
          "color=1D76DB",
          "-f",
          "description=Managed by codex-monitor",
        ],
        { parseJson: false },
      );
    } catch (err) {
      const text = String(err?.message || err).toLowerCase();
      if (
        text.includes("already_exists") ||
        text.includes("already exists") ||
        text.includes("unprocessable") ||
        text.includes("422")
      ) {
        return;
      }
      console.warn(
        `[kanban] failed to ensure label "${normalized}": ${err.message || err}`,
      );
    }
  }

  _extractProjectNumber(value) {
    if (!value) return null;
    const text = String(value).trim();
    if (/^\d+$/.test(text)) return text;
    const match = text.match(/\/projects\/(\d+)(?:\b|$)/i);
    return match?.[1] || null;
  }

  async _resolveProjectNumber() {
    if (this._cachedProjectNumber) return this._cachedProjectNumber;
    const owner = String(this._projectOwner || "").trim();
    const title = String(this._projectTitle || "Codex-Monitor").trim();
    if (!owner || !title) return null;

    try {
      const projects = await this._gh(
        ["project", "list", "--owner", owner, "--format", "json"],
        { parseJson: true },
      );
      const list = Array.isArray(projects)
        ? projects
        : Array.isArray(projects?.projects)
          ? projects.projects
          : [];
      const existing = list.find(
        (project) =>
          String(project?.title || "")
            .trim()
            .toLowerCase() === title.toLowerCase(),
      );
      const existingNumber = this._extractProjectNumber(
        existing?.number || existing?.url,
      );
      if (existingNumber) {
        this._cachedProjectNumber = existingNumber;
        return existingNumber;
      }
    } catch {
      return null;
    }

    try {
      const output = await this._gh(
        ["project", "create", "--owner", owner, "--title", title],
        { parseJson: false },
      );
      const createdNumber = this._extractProjectNumber(output);
      if (createdNumber) {
        this._cachedProjectNumber = createdNumber;
        return createdNumber;
      }
    } catch {
      return null;
    }

    return null;
  }

  async _ensureIssueLinkedToProject(issueUrl) {
    if (this._projectMode !== "kanban") return;
    const owner = String(this._projectOwner || "").trim();
    if (!owner || !issueUrl) return;
    const projectNumber = await this._resolveProjectNumber();
    if (!projectNumber) return;

    try {
      await this._gh(
        [
          "project",
          "item-add",
          String(projectNumber),
          "--owner",
          owner,
          "--url",
          issueUrl,
        ],
        { parseJson: false },
      );
    } catch (err) {
      const text = String(err?.message || err).toLowerCase();
      if (text.includes("already") && text.includes("item")) {
        return;
      }
      console.warn(
        `[kanban] failed to add issue to project ${owner}/${projectNumber}: ${err.message || err}`,
      );
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
    } else if (labelSet.has("inprogress") || labelSet.has("in-progress")) {
      status = "inprogress";
    } else if (labelSet.has("inreview") || labelSet.has("in-review")) {
      status = "inreview";
    } else if (labelSet.has("blocked")) {
      status = "blocked";
    }

    // Check for codex-monitor labels
    const codexMeta = {
      isIgnored: labelSet.has("codex:ignore"),
      isClaimed: labelSet.has("codex:claimed"),
      isWorking: labelSet.has("codex:working"),
      isStale: labelSet.has("codex:stale"),
    };

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
      projectId: `${this._owner}/${this._repo}`,
      branchName: branchMatch?.[1] || null,
      prNumber: prMatch?.[1] || null,
      meta: {
        ...issue,
        task_url: issue.url || null,
        codex: codexMeta,
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

  /**
   * Persist shared state to a Jira issue.
   *
   * Implements the same shared state protocol as GitHubAdapter but using Jira-specific
   * mechanisms. The implementation should use a combination of:
   *
   * 1. **Jira Custom Fields** (preferred if available):
   *    - Create custom fields for codex-monitor state (e.g., "Codex Owner ID", "Codex Attempt Token")
   *    - Store structured data as JSON in a text custom field
   *    - Use Jira API v3: `PUT /rest/api/3/issue/{issueKey}`
   *    - Custom field IDs are like `customfield_10042`
   *
   * 2. **Jira Labels** (for status flags):
   *    - Use labels: `codex:claimed`, `codex:working`, `codex:stale`, `codex:ignore`
   *    - Labels API: `PUT /rest/api/3/issue/{issueKey}` with `update.labels` field
   *    - Remove conflicting codex labels before adding new ones
   *
   * 3. **Structured Comments** (fallback if custom fields unavailable):
   *    - Similar to GitHub: embed JSON in HTML comment markers
   *    - Format: `<!-- codex-monitor-state\n{json}\n-->`
   *    - Comments API: `POST /rest/api/3/issue/{issueKey}/comment`
   *    - Update via `PUT /rest/api/3/issue/{issueKey}/comment/{commentId}`
   *
   * **Jira API v3 Authentication**:
   * - Use Basic Auth with email + API token: `Authorization: Basic base64(email:token)`
   * - Token from: https://id.atlassian.com/manage-profile/security/api-tokens
   * - Base URL: `https://{domain}.atlassian.net`
   *
   * **Required Permissions**:
   * - Browse Projects
   * - Edit Issues
   * - Add Comments
   * - Manage Custom Fields (if using custom fields approach)
   *
   * @param {string} issueKey - Jira issue key (e.g., "PROJ-123")
   * @param {SharedState} sharedState - Agent state to persist
   * @param {string} sharedState.ownerId - Format: "workstation-id/agent-id"
   * @param {string} sharedState.attemptToken - Unique UUID for this attempt
   * @param {string} sharedState.attemptStarted - ISO 8601 timestamp
   * @param {string} sharedState.heartbeat - ISO 8601 timestamp
   * @param {string} sharedState.status - One of: "claimed", "working", "stale"
   * @param {number} sharedState.retryCount - Number of retry attempts
   * @returns {Promise<boolean>} Success status
   *
   * @example
   * await adapter.persistSharedStateToIssue("PROJ-123", {
   *   ownerId: "workstation-123/agent-456",
   *   attemptToken: "uuid-here",
   *   attemptStarted: "2026-02-14T17:00:00Z",
   *   heartbeat: "2026-02-14T17:30:00Z",
   *   status: "working",
   *   retryCount: 1
   * });
   *
   * @see {@link https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/}
   * @see GitHubIssuesAdapter.persistSharedStateToIssue for reference implementation
   */
  async persistSharedStateToIssue(_issueKey, _sharedState) {
    throw new Error(
      `${TAG} Jira adapter: persistSharedStateToIssue() not yet implemented. ` +
        `See JSDoc for Jira API v3 custom fields/labels/comments approach. ` +
        `Reference: GitHubAdapter uses structured comments + labels. ` +
        `Jira implementation should prefer custom fields if available, ` +
        `fall back to structured comments for compatibility. ` +
        `Set JIRA_BASE_URL, JIRA_API_TOKEN, and JIRA_EMAIL env vars when ready.`,
    );
  }

  /**
   * Read shared state from a Jira issue.
   *
   * Retrieves agent state previously written by persistSharedStateToIssue().
   * Implementation should check multiple sources in order of preference:
   *
   * 1. **Jira Custom Fields** (if configured):
   *    - Read custom field values via `GET /rest/api/3/issue/{issueKey}`
   *    - Parse JSON from custom field (e.g., `fields.customfield_10042`)
   *    - Validate required fields before returning
   *
   * 2. **Structured Comments** (fallback):
   *    - Fetch comments via `GET /rest/api/3/issue/{issueKey}/comment`
   *    - Search for latest comment containing `<!-- codex-monitor-state`
   *    - Extract and parse JSON from HTML comment markers
   *    - Return most recent valid state
   *
   * **Validation Requirements**:
   * - Must have: ownerId, attemptToken, attemptStarted, heartbeat, status
   * - Status must be one of: "claimed", "working", "stale"
   * - Timestamps must be valid ISO 8601 format
   * - Return null if state is missing, invalid, or corrupted
   *
   * **Jira API v3 Endpoints**:
   * - Issue details: `GET /rest/api/3/issue/{issueKey}?fields=customfield_*,comment`
   * - Comments only: `GET /rest/api/3/issue/{issueKey}/comment`
   *
   * @param {string} issueKey - Jira issue key (e.g., "PROJ-123")
   * @returns {Promise<SharedState|null>} Parsed shared state or null if not found
   *
   * @typedef {Object} SharedState
   * @property {string} ownerId - Workstation/agent identifier
   * @property {string} attemptToken - Unique UUID for this attempt
   * @property {string} attemptStarted - ISO 8601 timestamp
   * @property {string} heartbeat - ISO 8601 timestamp
   * @property {string} status - One of: "claimed", "working", "stale"
   * @property {number} retryCount - Number of retry attempts
   *
   * @example
   * const state = await adapter.readSharedStateFromIssue("PROJ-123");
   * if (state) {
   *   console.log(`Task claimed by ${state.ownerId}`);
   *   console.log(`Status: ${state.status}, Heartbeat: ${state.heartbeat}`);
   * } else {
   *   console.log("No shared state found - task is unclaimed");
   * }
   *
   * @see {@link https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-comments/}
   * @see GitHubIssuesAdapter.readSharedStateFromIssue for reference implementation
   */
  async readSharedStateFromIssue(_issueKey) {
    throw new Error(
      `${TAG} Jira adapter: readSharedStateFromIssue() not yet implemented. ` +
        `See JSDoc for Jira API v3 custom fields/comments parsing approach. ` +
        `Should return SharedState object with {ownerId, attemptToken, attemptStarted, ` +
        `heartbeat, status, retryCount} or null if not found. ` +
        `Reference: GitHubAdapter parses structured HTML comments. ` +
        `Set JIRA_BASE_URL, JIRA_API_TOKEN, and JIRA_EMAIL env vars when ready.`,
    );
  }

  /**
   * Mark a Jira issue as ignored by codex-monitor.
   *
   * Prevents codex-monitor from repeatedly attempting to claim or work on tasks
   * that are not suitable for automation. Uses Jira-specific mechanisms:
   *
   * 1. **Add Label**: `codex:ignore`
   *    - Labels API: `PUT /rest/api/3/issue/{issueKey}`
   *    - Request body: `{"update": {"labels": [{"add": "codex:ignore"}]}}`
   *    - Labels are case-sensitive in Jira
   *
   * 2. **Add Comment**: Human-readable explanation
   *    - Comments API: `POST /rest/api/3/issue/{issueKey}/comment`
   *    - Request body: `{"body": {"type": "doc", "version": 1, "content": [...]}}`
   *    - Jira uses Atlassian Document Format (ADF) for rich text
   *    - For simple text: `{"body": "text content"}` (legacy format)
   *
   * 3. **Optional: Transition Issue** (if workflow supports it):
   *    - Get transitions: `GET /rest/api/3/issue/{issueKey}/transitions`
   *    - Transition to "Won't Do" or similar: `POST /rest/api/3/issue/{issueKey}/transitions`
   *    - Not required if labels are sufficient
   *
   * **Jira ADF Comment Example**:
   * ```json
   * {
   *   "body": {
   *     "type": "doc",
   *     "version": 1,
   *     "content": [
   *       {
   *         "type": "paragraph",
   *         "content": [
   *           {"type": "text", "text": "Codex Monitor: Task marked as ignored."}
   *         ]
   *       }
   *     ]
   *   }
   * }
   * ```
   *
   * **Required Permissions**:
   * - Edit Issues (for labels)
   * - Add Comments
   * - Transition Issues (optional, if changing status)
   *
   * @param {string} issueKey - Jira issue key (e.g., "PROJ-123")
   * @param {string} reason - Human-readable reason for ignoring
   * @returns {Promise<boolean>} Success status
   *
   * @example
   * await adapter.markTaskIgnored("PROJ-123", "Task requires manual security review");
   * // Adds "codex:ignore" label and comment explaining why
   *
   * @example
   * await adapter.markTaskIgnored("PROJ-456", "Task dependencies not in automation scope");
   * // Prevents codex-monitor from claiming this task in future iterations
   *
   * @see {@link https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/}
   * @see {@link https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/}
   * @see GitHubIssuesAdapter.markTaskIgnored for reference implementation
   */
  async markTaskIgnored(_issueKey, _reason) {
    throw new Error(
      `${TAG} Jira adapter: markTaskIgnored() not yet implemented. ` +
        `See JSDoc for Jira API v3 labels/comments approach. ` +
        `Should add 'codex:ignore' label and post comment with reason. ` +
        `Consider using Atlassian Document Format (ADF) for rich comments. ` +
        `Reference: GitHubAdapter uses gh CLI for labels + comments. ` +
        `Set JIRA_BASE_URL, JIRA_API_TOKEN, and JIRA_EMAIL env vars when ready.`,
    );
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

/**
 * Persist shared state to an issue (GitHub adapter only).
 * @param {string} taskId - Task identifier (issue number for GitHub)
 * @param {SharedState} sharedState - State to persist
 * @returns {Promise<boolean>} Success status
 */
export async function persistSharedStateToIssue(taskId, sharedState) {
  const adapter = getKanbanAdapter();
  if (typeof adapter.persistSharedStateToIssue === "function") {
    return adapter.persistSharedStateToIssue(taskId, sharedState);
  }
  console.warn(
    `[kanban] persistSharedStateToIssue not supported by ${adapter.name} backend`,
  );
  return false;
}

/**
 * Read shared state from an issue (GitHub adapter only).
 * @param {string} taskId - Task identifier (issue number for GitHub)
 * @returns {Promise<SharedState|null>} Shared state or null
 */
export async function readSharedStateFromIssue(taskId) {
  const adapter = getKanbanAdapter();
  if (typeof adapter.readSharedStateFromIssue === "function") {
    return adapter.readSharedStateFromIssue(taskId);
  }
  return null;
}

/**
 * Mark a task as ignored by codex-monitor (GitHub adapter only).
 * @param {string} taskId - Task identifier (issue number for GitHub)
 * @param {string} reason - Human-readable reason for ignoring
 * @returns {Promise<boolean>} Success status
 */
export async function markTaskIgnored(taskId, reason) {
  const adapter = getKanbanAdapter();
  if (typeof adapter.markTaskIgnored === "function") {
    return adapter.markTaskIgnored(taskId, reason);
  }
  console.warn(
    `[kanban] markTaskIgnored not supported by ${adapter.name} backend`,
  );
  return false;
}
