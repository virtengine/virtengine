/* ─────────────────────────────────────────────────────────────
 *  VirtEngine Control Center – Reactive State Layer
 *  All signals, data loaders, toast system, and tab refresh logic
 * ────────────────────────────────────────────────────────────── */

import { signal } from "https://esm.sh/@preact/signals@1.3.1";
import { apiFetch, onWsMessage } from "./api.js";
import { cloneValue } from "./utils.js";
import { generateId } from "./utils.js";

/* ═══════════════════════════════════════════════════════════════
 *  SIGNALS — Single source of truth for UI state
 * ═══════════════════════════════════════════════════════════════ */

// ── Overall connectivity
export const connected = signal(false);

// ── Dashboard
export const statusData = signal(null);
export const executorData = signal(null);
export const projectSummary = signal(null);

// ── Tasks
export const tasksData = signal([]);
export const tasksPage = signal(1);
export const tasksPageSize = signal(20);
export const tasksFilter = signal("all");
export const tasksPriority = signal("all");
export const tasksSearch = signal("");
export const tasksSort = signal("updated");
export const tasksTotalPages = signal(1);

// ── Agents
export const agentsData = signal([]);

// ── Infra
export const worktreeData = signal([]);
export const sharedWorkspaces = signal([]);
export const presenceInstances = signal([]);
export const coordinatorInfo = signal(null);
export const infraData = signal(null);

// ── Logs
export const logsData = signal(null);
export const logsLines = signal(100);
export const gitDiff = signal(null);
export const gitBranches = signal([]);
export const agentLogFiles = signal([]);
export const agentLogFile = signal("");
export const agentLogTail = signal(null);
export const agentLogLines = signal(200);
export const agentLogQuery = signal("");
export const agentContext = signal(null);

// ── Toasts
export const toasts = signal([]);

/* ═══════════════════════════════════════════════════════════════
 *  TOAST SYSTEM
 * ═══════════════════════════════════════════════════════════════ */

/**
 * Show a toast notification that auto-dismisses after 3 s.
 * @param {string} message
 * @param {'info'|'success'|'error'|'warning'} type
 */
export function showToast(message, type = "info") {
  const id = generateId();
  toasts.value = [...toasts.value, { id, message, type }];
  setTimeout(() => {
    toasts.value = toasts.value.filter((t) => t.id !== id);
  }, 3000);
}

// Listen for api-error events dispatched by the api module
if (typeof globalThis !== "undefined") {
  try {
    globalThis.addEventListener("ve:api-error", (e) => {
      showToast(e.detail?.message || "Request failed", "error");
    });
  } catch {
    /* SSR guard */
  }
}

/* ═══════════════════════════════════════════════════════════════
 *  DATA LOADERS — each calls apiFetch and updates its signal(s)
 * ═══════════════════════════════════════════════════════════════ */

/** Load system status → statusData */
export async function loadStatus() {
  const res = await apiFetch("/api/status", { _silent: true }).catch(() => ({
    data: null,
  }));
  statusData.value = res.data ?? res ?? null;
  connected.value = true;
}

/** Load executor state → executorData */
export async function loadExecutor() {
  const res = await apiFetch("/api/executor", { _silent: true }).catch(() => ({
    data: null,
  }));
  executorData.value = res ?? null;
}

/** Load tasks with current filter/page/sort → tasksData + tasksTotalPages */
export async function loadTasks() {
  const params = new URLSearchParams({
    page: String(tasksPage.value),
    pageSize: String(tasksPageSize.value),
  });
  if (tasksFilter.value && tasksFilter.value !== "all")
    params.set("filter", tasksFilter.value);
  if (tasksPriority.value && tasksPriority.value !== "all")
    params.set("priority", tasksPriority.value);
  if (tasksSearch.value) params.set("search", tasksSearch.value);
  if (tasksSort.value) params.set("sort", tasksSort.value);

  const res = await apiFetch(`/api/tasks?${params}`, { _silent: true }).catch(
    () => ({
      data: [],
      total: 0,
      totalPages: 1,
    }),
  );
  tasksData.value = res.data || [];
  tasksTotalPages.value =
    res.totalPages ||
    Math.max(1, Math.ceil((res.total || 0) / tasksPageSize.value));
}

/** Load active agents → agentsData */
export async function loadAgents() {
  const res = await apiFetch("/api/agents", { _silent: true }).catch(() => ({
    data: [],
  }));
  agentsData.value = res.data || [];
}

/** Load worktrees → worktreeData */
export async function loadWorktrees() {
  const res = await apiFetch("/api/worktrees", { _silent: true }).catch(() => ({
    data: [],
    stats: null,
  }));
  worktreeData.value = res.data || [];
}

/** Load infrastructure overview → infraData */
export async function loadInfra() {
  const res = await apiFetch("/api/infra", { _silent: true }).catch(() => ({
    data: null,
  }));
  infraData.value = res.data ?? res ?? null;
}

/** Load system logs → logsData */
export async function loadLogs() {
  const res = await apiFetch(`/api/logs?lines=${logsLines.value}`, {
    _silent: true,
  }).catch(() => ({ data: null }));
  logsData.value = res.data ?? res ?? null;
}

/** Load git branches + diff → gitBranches, gitDiff */
export async function loadGit() {
  const [branches, diff] = await Promise.all([
    apiFetch("/api/git/branches", { _silent: true }).catch(() => ({
      data: [],
    })),
    apiFetch("/api/git/diff", { _silent: true }).catch(() => ({ data: "" })),
  ]);
  gitBranches.value = branches.data || [];
  gitDiff.value = diff.data || "";
}

/** Load agent log file list → agentLogFiles */
export async function loadAgentLogFileList() {
  const params = new URLSearchParams();
  if (agentLogQuery.value) params.set("query", agentLogQuery.value);
  const path = params.toString()
    ? `/api/agent-logs?${params}`
    : "/api/agent-logs";
  const res = await apiFetch(path, { _silent: true }).catch(() => ({
    data: [],
  }));
  agentLogFiles.value = res.data || [];
}

/** Load tail of the currently selected agent log → agentLogTail */
export async function loadAgentLogTailData() {
  if (!agentLogFile.value) {
    agentLogTail.value = null;
    return;
  }
  const params = new URLSearchParams({
    file: agentLogFile.value,
    lines: String(agentLogLines.value),
  });
  const res = await apiFetch(`/api/agent-logs/tail?${params}`, {
    _silent: true,
  }).catch(() => ({ data: null }));
  agentLogTail.value = res.data ?? res ?? null;
}

/**
 * Load worktree context for a branch/query → agentContext
 * @param {string} query
 */
export async function loadAgentContextData(query) {
  if (!query) {
    agentContext.value = null;
    return;
  }
  const res = await apiFetch(
    `/api/agent-context?query=${encodeURIComponent(query)}`,
    { _silent: true },
  ).catch(() => ({ data: null }));
  agentContext.value = res.data ?? res ?? null;
}

/** Load shared workspaces → sharedWorkspaces */
export async function loadSharedWorkspaces() {
  const res = await apiFetch("/api/shared-workspaces", { _silent: true }).catch(
    () => ({
      data: [],
    }),
  );
  sharedWorkspaces.value = res.data || res.workspaces || [];
}

/** Load presence / coordinator → presenceInstances, coordinatorInfo */
export async function loadPresence() {
  const res = await apiFetch("/api/presence", { _silent: true }).catch(() => ({
    data: null,
  }));
  const data = res.data || res || {};
  presenceInstances.value = data.instances || [];
  coordinatorInfo.value = data.coordinator || null;
}

/** Load project summary → projectSummary */
export async function loadProjectSummary() {
  const res = await apiFetch("/api/project-summary", { _silent: true }).catch(
    () => ({
      data: null,
    }),
  );
  projectSummary.value = res.data ?? res ?? null;
}

/* ═══════════════════════════════════════════════════════════════
 *  TAB REFRESH — map tab names to their required loaders
 * ═══════════════════════════════════════════════════════════════ */

const TAB_LOADERS = {
  dashboard: () =>
    Promise.all([loadStatus(), loadExecutor(), loadProjectSummary()]),
  tasks: () => loadTasks(),
  agents: () => loadAgents(),
  infra: () =>
    Promise.all([
      loadWorktrees(),
      loadInfra(),
      loadSharedWorkspaces(),
      loadPresence(),
    ]),
  control: () => loadExecutor(),
  logs: () =>
    Promise.all([loadLogs(), loadAgentLogFileList(), loadAgentLogTailData()]),
  settings: () => loadStatus(),
};

/**
 * Refresh all data for a given tab.
 * @param {string} tabName
 */
export async function refreshTab(tabName) {
  const loader = TAB_LOADERS[tabName];
  if (loader) {
    try {
      await loader();
    } catch {
      /* errors handled by individual loaders */
    }
  }
}

/* ═══════════════════════════════════════════════════════════════
 *  HELPERS
 * ═══════════════════════════════════════════════════════════════ */

/**
 * Optimistic update pattern:
 * 1. Apply the optimistic change immediately
 * 2. Run the async fetch
 * 3. On error, revert via rollback
 *
 * @param {() => void} applyFn   – mutate signals optimistically
 * @param {() => Promise<any>} fetchFn  – the actual API call
 * @param {() => void} revertFn  – undo the optimistic change on error
 * @returns {Promise<any>}
 */
export async function runOptimistic(applyFn, fetchFn, revertFn) {
  try {
    applyFn();
    return await fetchFn();
  } catch (err) {
    if (typeof revertFn === "function") revertFn();
    throw err;
  }
}

/** @type {ReturnType<typeof setTimeout>|null} */
let _scheduleTimer = null;

/**
 * Schedule a tab refresh after a short delay (debounced).
 * Uses the the current activeTab from the router layer via import.
 * Falls back to refreshing 'dashboard'.
 *
 * @param {number} ms
 */
export function scheduleRefresh(ms = 5000) {
  if (_scheduleTimer) clearTimeout(_scheduleTimer);
  _scheduleTimer = setTimeout(async () => {
    _scheduleTimer = null;
    // Dynamic import to avoid circular dependency at module load time
    try {
      const { activeTab } = await import("./router.js");
      await refreshTab(activeTab.value);
    } catch {
      await refreshTab("dashboard");
    }
  }, ms);
}

/* ─── WebSocket invalidation listener ─── */

const WS_CHANNEL_MAP = {
  dashboard: ["overview", "executor", "tasks", "agents"],
  tasks: ["tasks"],
  agents: ["agents", "executor"],
  infra: ["worktrees", "workspaces", "presence"],
  control: ["executor", "overview"],
  logs: ["*"],
  settings: ["overview"],
};

/** Start listening for WS invalidation messages and auto-refreshing. */
export function initWsInvalidationListener() {
  onWsMessage((msg) => {
    if (msg?.type !== "invalidate") return;
    const channels = Array.isArray(msg.channels) ? msg.channels : [];

    // Determine interested channels based on active tab
    import("./router.js")
      .then(({ activeTab }) => {
        const interested = WS_CHANNEL_MAP[activeTab.value] || ["*"];
        if (
          channels.includes("*") ||
          channels.some((c) => interested.includes(c))
        ) {
          scheduleRefresh(150);
        }
      })
      .catch(() => {
        /* noop */
      });
  });
}
