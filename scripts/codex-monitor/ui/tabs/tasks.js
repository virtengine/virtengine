/* ─────────────────────────────────────────────────────────────
 *  Tab: Tasks — board, search, filters, task CRUD
 * ────────────────────────────────────────────────────────────── */
import { h } from "https://esm.sh/preact@10.25.4";
import {
  useState,
  useEffect,
  useRef,
  useCallback,
} from "https://esm.sh/preact@10.25.4/hooks";
import htm from "https://esm.sh/htm@3.1.1";

const html = htm.bind(h);

import { haptic, showConfirm } from "../modules/telegram.js";
import { apiFetch, sendCommandToChat } from "../modules/api.js";
import {
  tasksData,
  tasksPage,
  tasksPageSize,
  tasksFilter,
  tasksPriority,
  tasksSearch,
  tasksSort,
  tasksTotalPages,
  executorData,
  showToast,
  refreshTab,
  runOptimistic,
  scheduleRefresh,
  loadTasks,
} from "../modules/state.js";
import { ICONS } from "../modules/icons.js";
import {
  cloneValue,
  formatRelative,
  truncate,
  debounce,
} from "../modules/utils.js";
import {
  Card,
  Badge,
  StatCard,
  SkeletonCard,
  Modal,
  EmptyState,
  ListItem,
} from "../components/shared.js";
import { SegmentedControl, SearchInput, Toggle } from "../components/forms.js";

/* ─── Status chip definitions ─── */
const STATUS_CHIPS = [
  { value: "all", label: "All" },
  { value: "todo", label: "Todo" },
  { value: "inprogress", label: "Active" },
  { value: "inreview", label: "Review" },
  { value: "done", label: "Done" },
  { value: "error", label: "Error" },
];

const PRIORITY_CHIPS = [
  { value: "", label: "Any" },
  { value: "low", label: "Low" },
  { value: "medium", label: "Med" },
  { value: "high", label: "High" },
  { value: "critical", label: "Crit" },
];

const SORT_OPTIONS = [
  { value: "updated", label: "Updated" },
  { value: "created", label: "Created" },
  { value: "priority", label: "Priority" },
  { value: "title", label: "Title" },
];

/* ─── TaskDetailModal ─── */
export function TaskDetailModal({ task, onClose }) {
  const [title, setTitle] = useState(task?.title || "");
  const [description, setDescription] = useState(task?.description || "");
  const [status, setStatus] = useState(task?.status || "todo");
  const [priority, setPriority] = useState(task?.priority || "");
  const [saving, setSaving] = useState(false);

  const handleSave = async () => {
    setSaving(true);
    haptic("medium");
    const prev = cloneValue(tasksData.value);
    try {
      await runOptimistic(
        () => {
          tasksData.value = tasksData.value.map((t) =>
            t.id === task.id
              ? { ...t, title, description, status, priority: priority || null }
              : t,
          );
        },
        async () => {
          const res = await apiFetch("/api/tasks/edit", {
            method: "POST",
            body: JSON.stringify({
              taskId: task.id,
              title,
              description,
              status,
              priority,
            }),
          });
          if (res?.data)
            tasksData.value = tasksData.value.map((t) =>
              t.id === task.id ? { ...t, ...res.data } : t,
            );
          return res;
        },
        () => {
          tasksData.value = prev;
        },
      );
      showToast("Task saved", "success");
      onClose();
    } catch {
      /* toast via apiFetch */
    }
    setSaving(false);
  };

  const handleStatusUpdate = async (newStatus) => {
    haptic("medium");
    const prev = cloneValue(tasksData.value);
    try {
      await runOptimistic(
        () => {
          tasksData.value = tasksData.value.map((t) =>
            t.id === task.id ? { ...t, status: newStatus } : t,
          );
        },
        async () => {
          const res = await apiFetch("/api/tasks/update", {
            method: "POST",
            body: JSON.stringify({ taskId: task.id, status: newStatus }),
          });
          if (res?.data)
            tasksData.value = tasksData.value.map((t) =>
              t.id === task.id ? { ...t, ...res.data } : t,
            );
          return res;
        },
        () => {
          tasksData.value = prev;
        },
      );
      if (newStatus === "done" || newStatus === "cancelled") onClose();
      else setStatus(newStatus);
    } catch {
      /* toast */
    }
  };

  const handleStart = async () => {
    haptic("medium");
    const prev = cloneValue(tasksData.value);
    try {
      await runOptimistic(
        () => {
          tasksData.value = tasksData.value.map((t) =>
            t.id === task.id ? { ...t, status: "inprogress" } : t,
          );
        },
        () =>
          apiFetch("/api/tasks/start", {
            method: "POST",
            body: JSON.stringify({ taskId: task.id }),
          }),
        () => {
          tasksData.value = prev;
        },
      );
      onClose();
    } catch {
      /* toast */
    }
    scheduleRefresh(150);
  };

  const handleRetry = async () => {
    haptic("medium");
    try {
      await apiFetch("/api/tasks/retry", {
        method: "POST",
        body: JSON.stringify({ taskId: task.id }),
      });
      showToast("Task retried", "success");
      onClose();
      scheduleRefresh(150);
    } catch {
      /* toast */
    }
  };

  const handleCancel = async () => {
    const ok = await showConfirm("Cancel this task?");
    if (!ok) return;
    await handleStatusUpdate("cancelled");
  };

  return html`
    <${Modal} title=${task?.title || "Task Detail"} onClose=${onClose}>
      <div class="meta-text mb-sm" style="user-select:all">ID: ${task?.id}</div>
      <div class="flex-row gap-sm mb-md">
        <${Badge} status=${task?.status} text=${task?.status} />
        ${task?.priority &&
        html`<${Badge} status=${task.priority} text=${task.priority} />`}
      </div>

      <div class="flex-col gap-md">
        <input
          class="input"
          placeholder="Title"
          value=${title}
          onInput=${(e) => setTitle(e.target.value)}
        />
        <textarea
          class="input"
          rows="5"
          placeholder="Description"
          value=${description}
          onInput=${(e) => setDescription(e.target.value)}
        ></textarea>

        <div class="input-row">
          <select
            class="input"
            value=${status}
            onChange=${(e) => setStatus(e.target.value)}
          >
            ${["todo", "inprogress", "inreview", "done", "cancelled"].map(
              (s) => html`<option value=${s}>${s}</option>`,
            )}
          </select>
          <select
            class="input"
            value=${priority}
            onChange=${(e) => setPriority(e.target.value)}
          >
            <option value="">No priority</option>
            ${["low", "medium", "high", "critical"].map(
              (p) => html`<option value=${p}>${p}</option>`,
            )}
          </select>
        </div>

        <!-- Metadata -->
        ${task?.created_at &&
        html`
          <div class="meta-text">
            Created: ${new Date(task.created_at).toLocaleString()}
          </div>
        `}
        ${task?.updated_at &&
        html`
          <div class="meta-text">
            Updated: ${formatRelative(task.updated_at)}
          </div>
        `}
        ${task?.assignee &&
        html` <div class="meta-text">Assignee: ${task.assignee}</div> `}
        ${task?.branch &&
        html`
          <div class="meta-text" style="user-select:all">
            Branch: ${task.branch}
          </div>
        `}

        <!-- Action buttons -->
        <div class="btn-row">
          ${task?.status === "todo" &&
          html`
            <button class="btn btn-primary btn-sm" onClick=${handleStart}>
              ▶ Start
            </button>
          `}
          ${(task?.status === "error" || task?.status === "cancelled") &&
          html`
            <button class="btn btn-primary btn-sm" onClick=${handleRetry}>
              ↻ Retry
            </button>
          `}
          <button
            class="btn btn-secondary btn-sm"
            onClick=${handleSave}
            disabled=${saving}
          >
            ${saving ? "Saving…" : "💾 Save"}
          </button>
          <button
            class="btn btn-ghost btn-sm"
            onClick=${() => handleStatusUpdate("inreview")}
          >
            → Review
          </button>
          <button
            class="btn btn-ghost btn-sm"
            onClick=${() => handleStatusUpdate("done")}
          >
            ✓ Done
          </button>
          ${task?.status !== "cancelled" &&
          html`
            <button
              class="btn btn-ghost btn-sm"
              style="color:var(--color-error)"
              onClick=${handleCancel}
            >
              ✕ Cancel
            </button>
          `}
        </div>

        <!-- Agent log link -->
        ${task?.id &&
        html`
          <button
            class="btn btn-ghost btn-sm"
            onClick=${() => {
              haptic();
              sendCommandToChat("/logs " + task.id);
            }}
          >
            📄 View Agent Logs
          </button>
        `}
      </div>
    <//>
  `;
}

/* ─── TasksTab ─── */
export function TasksTab() {
  const [showCreate, setShowCreate] = useState(false);
  const [detailTask, setDetailTask] = useState(null);
  const [manualMode, setManualMode] = useState(false);
  const [batchMode, setBatchMode] = useState(false);
  const [selectedIds, setSelectedIds] = useState(new Set());
  const searchRef = useRef(null);

  const tasks = tasksData.value || [];
  const filterVal = tasksFilter?.value ?? "todo";
  const priorityVal = tasksPriority?.value ?? "";
  const searchVal = tasksSearch?.value ?? "";
  const sortVal = tasksSort?.value ?? "updated";
  const page = tasksPage?.value ?? 0;
  const pageSize = tasksPageSize?.value ?? 8;
  const totalPages = tasksTotalPages?.value ?? 1;

  /* Search (local fuzzy filter on already-loaded data) */
  const searchLower = searchVal.trim().toLowerCase();
  const visible = searchLower
    ? tasks.filter((t) =>
        `${t.title || ""} ${t.description || ""} ${t.id || ""}`
          .toLowerCase()
          .includes(searchLower),
      )
    : tasks;

  const canManual = Boolean(executorData.value?.data);

  /* ── Handlers ── */
  const handleFilter = async (s) => {
    haptic();
    if (tasksFilter) tasksFilter.value = s;
    if (tasksPage) tasksPage.value = 0;
    await refreshTab();
  };

  const handlePriorityFilter = async (p) => {
    haptic();
    if (tasksPriority) tasksPriority.value = p;
    if (tasksPage) tasksPage.value = 0;
    await refreshTab();
  };

  const handleSort = async (e) => {
    haptic();
    if (tasksSort) tasksSort.value = e.target.value;
    if (tasksPage) tasksPage.value = 0;
    await refreshTab();
  };

  const handleSearch = useCallback(
    debounce((val) => {
      if (tasksSearch) tasksSearch.value = val;
    }, 250),
    [],
  );

  const handlePrev = async () => {
    if (tasksPage) tasksPage.value = Math.max(0, page - 1);
    await refreshTab();
  };

  const handleNext = async () => {
    if (tasksPage) tasksPage.value = page + 1;
    await refreshTab();
  };

  const handleStatusUpdate = async (taskId, newStatus) => {
    haptic("medium");
    const prev = cloneValue(tasks);
    await runOptimistic(
      () => {
        tasksData.value = tasksData.value.map((t) =>
          t.id === taskId ? { ...t, status: newStatus } : t,
        );
      },
      async () => {
        const res = await apiFetch("/api/tasks/update", {
          method: "POST",
          body: JSON.stringify({ taskId, status: newStatus }),
        });
        if (res?.data)
          tasksData.value = tasksData.value.map((t) =>
            t.id === taskId ? { ...t, ...res.data } : t,
          );
      },
      () => {
        tasksData.value = prev;
      },
    ).catch(() => {});
  };

  const handleStart = async (taskId) => {
    haptic("medium");
    const prev = cloneValue(tasks);
    await runOptimistic(
      () => {
        tasksData.value = tasksData.value.map((t) =>
          t.id === taskId ? { ...t, status: "inprogress" } : t,
        );
      },
      () =>
        apiFetch("/api/tasks/start", {
          method: "POST",
          body: JSON.stringify({ taskId }),
        }),
      () => {
        tasksData.value = prev;
      },
    ).catch(() => {});
    scheduleRefresh(150);
  };

  const openDetail = async (taskId) => {
    haptic();
    const local = tasks.find((t) => t.id === taskId);
    const result = await apiFetch(
      `/api/tasks/detail?taskId=${encodeURIComponent(taskId)}`,
      { _silent: true },
    ).catch(() => ({ data: local }));
    setDetailTask(result.data || local);
  };

  /* ── Batch operations ── */
  const toggleSelect = (id) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  };

  const handleBatchDone = async () => {
    if (!selectedIds.size) return;
    const ok = await showConfirm(`Mark ${selectedIds.size} tasks as done?`);
    if (!ok) return;
    haptic("medium");
    for (const id of selectedIds) {
      await handleStatusUpdate(id, "done");
    }
    setSelectedIds(new Set());
    setBatchMode(false);
    scheduleRefresh(150);
  };

  const handleBatchCancel = async () => {
    if (!selectedIds.size) return;
    const ok = await showConfirm(`Cancel ${selectedIds.size} tasks?`);
    if (!ok) return;
    haptic("medium");
    for (const id of selectedIds) {
      await handleStatusUpdate(id, "cancelled");
    }
    setSelectedIds(new Set());
    setBatchMode(false);
    scheduleRefresh(150);
  };

  /* ── Render ── */
  if (!tasks.length && !searchVal)
    return html`<${Card} title="Loading Tasks…"><${SkeletonCard} /><//>`;

  return html`
    <!-- Sticky search bar -->
    <div class="sticky-search">
      <${SearchInput}
        ref=${searchRef}
        placeholder="Search tasks…"
        value=${searchVal}
        onInput=${(e) => handleSearch(e.target.value)}
      />
    </div>

    <!-- Filters -->
    <${Card} title="Task Board">
      <div class="chip-group mb-sm">
        ${STATUS_CHIPS.map(
          (s) => html`
            <button
              key=${s.value}
              class="chip ${filterVal === s.value ? "active" : ""}"
              onClick=${() => handleFilter(s.value)}
            >
              ${s.label}
            </button>
          `,
        )}
      </div>
      <div class="chip-group mb-sm">
        ${PRIORITY_CHIPS.map(
          (p) => html`
            <button
              key=${p.value}
              class="chip chip-outline ${priorityVal === p.value
                ? "active"
                : ""}"
              onClick=${() => handlePriorityFilter(p.value)}
            >
              ${p.label}
            </button>
          `,
        )}
      </div>
      <div class="flex-between mb-sm">
        <select
          class="input input-sm"
          value=${sortVal}
          onChange=${handleSort}
          style="max-width:140px"
        >
          ${SORT_OPTIONS.map(
            (o) =>
              html`<option key=${o.value} value=${o.value}>${o.label}</option>`,
          )}
        </select>
        <span class="pill">${visible.length} shown</span>
      </div>

      <!-- Manual mode + batch mode toggles -->
      <div class="flex-between mb-sm">
        <label
          class="meta-text toggle-label"
          onClick=${() => {
            if (canManual) {
              setManualMode(!manualMode);
              haptic();
            }
          }}
        >
          <input
            type="checkbox"
            checked=${manualMode}
            disabled=${!canManual}
            style="accent-color:var(--accent)"
          />
          Manual Mode
        </label>
        <label
          class="meta-text toggle-label"
          onClick=${() => {
            setBatchMode(!batchMode);
            haptic();
            setSelectedIds(new Set());
          }}
        >
          <input
            type="checkbox"
            checked=${batchMode}
            style="accent-color:var(--accent)"
          />
          Batch Select
        </label>
      </div>

      <!-- Batch action bar -->
      ${batchMode &&
      selectedIds.size > 0 &&
      html`
        <div class="btn-row mb-md batch-action-bar">
          <span class="pill">${selectedIds.size} selected</span>
          <button class="btn btn-primary btn-sm" onClick=${handleBatchDone}>
            ✓ Done All
          </button>
          <button class="btn btn-danger btn-sm" onClick=${handleBatchCancel}>
            ✕ Cancel All
          </button>
          <button
            class="btn btn-ghost btn-sm"
            onClick=${() => {
              setSelectedIds(new Set());
              haptic();
            }}
          >
            Clear
          </button>
        </div>
      `}
    <//>

    <!-- Task list -->
    ${visible.map(
      (task) => html`
        <div
          key=${task.id}
          class="task-card ${batchMode && selectedIds.has(task.id)
            ? "task-card-selected"
            : ""} task-card-enter"
          onClick=${() =>
            batchMode ? toggleSelect(task.id) : openDetail(task.id)}
        >
          ${batchMode &&
          html`
            <input
              type="checkbox"
              checked=${selectedIds.has(task.id)}
              class="task-checkbox"
              onClick=${(e) => {
                e.stopPropagation();
                toggleSelect(task.id);
              }}
              style="accent-color:var(--accent)"
            />
          `}
          <div class="task-card-header">
            <div>
              <div class="task-card-title">${task.title || "(untitled)"}</div>
              <div class="task-card-meta">
                ${task.id}${task.priority
                  ? html` ·
                      <${Badge}
                        status=${task.priority}
                        text=${task.priority}
                      />`
                  : ""}
                ${task.updated_at
                  ? html` · ${formatRelative(task.updated_at)}`
                  : ""}
              </div>
            </div>
            <${Badge} status=${task.status} text=${task.status} />
          </div>
          <div class="meta-text">
            ${task.description
              ? truncate(task.description, 120)
              : "No description."}
          </div>
          ${!batchMode &&
          html`
            <div class="btn-row mt-sm" onClick=${(e) => e.stopPropagation()}>
              ${manualMode &&
              task.status === "todo" &&
              canManual &&
              html`
                <button
                  class="btn btn-primary btn-sm"
                  onClick=${() => handleStart(task.id)}
                >
                  ▶ Start
                </button>
              `}
              <button
                class="btn btn-secondary btn-sm"
                onClick=${() => handleStatusUpdate(task.id, "inreview")}
              >
                → Review
              </button>
              <button
                class="btn btn-ghost btn-sm"
                onClick=${() => handleStatusUpdate(task.id, "done")}
              >
                ✓ Done
              </button>
            </div>
          `}
        </div>
      `,
    )}
    ${!visible.length && html`<${EmptyState} message="No tasks found." />`}

    <!-- Pagination -->
    <div class="pager">
      <button
        class="btn btn-secondary btn-sm"
        onClick=${handlePrev}
        disabled=${page <= 0}
      >
        ← Prev
      </button>
      <span class="pager-info">Page ${page + 1} / ${totalPages}</span>
      <button
        class="btn btn-secondary btn-sm"
        onClick=${handleNext}
        disabled=${page + 1 >= totalPages}
      >
        Next →
      </button>
    </div>

    <!-- FAB -->
    <button
      class="fab"
      onClick=${() => {
        haptic();
        setShowCreate(true);
      }}
    >
      ${ICONS.plus}
    </button>

    <!-- Modals -->
    ${showCreate &&
    html`
      <!-- re-use CreateTaskModal from dashboard.js -->
      <${CreateTaskModalInline} onClose=${() => setShowCreate(false)} />
    `}
    ${detailTask &&
    html`
      <${TaskDetailModal}
        task=${detailTask}
        onClose=${() => setDetailTask(null)}
      />
    `}
  `;
}

/* ── Inline CreateTask (duplicated here to keep tasks.js self-contained) ── */
function CreateTaskModalInline({ onClose }) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [priority, setPriority] = useState("medium");
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async () => {
    if (!title.trim()) {
      showToast("Title is required", "error");
      return;
    }
    setSubmitting(true);
    haptic("medium");
    try {
      await apiFetch("/api/tasks/create", {
        method: "POST",
        body: JSON.stringify({
          title: title.trim(),
          description: description.trim(),
          priority,
        }),
      });
      showToast("Task created", "success");
      onClose();
      await refreshTab();
    } catch {
      /* toast */
    }
    setSubmitting(false);
  };

  useEffect(() => {
    const tg = globalThis.Telegram?.WebApp;
    if (tg?.MainButton) {
      tg.MainButton.setText("Create Task");
      tg.MainButton.show();
      tg.MainButton.onClick(handleSubmit);
      return () => {
        tg.MainButton.hide();
        tg.MainButton.offClick(handleSubmit);
      };
    }
  }, [title, description, priority]);

  return html`
    <${Modal} title="New Task" onClose=${onClose}>
      <div class="flex-col gap-md">
        <input
          class="input"
          placeholder="Task title"
          value=${title}
          onInput=${(e) => setTitle(e.target.value)}
        />
        <textarea
          class="input"
          rows="4"
          placeholder="Description"
          value=${description}
          onInput=${(e) => setDescription(e.target.value)}
        ></textarea>
        <${SegmentedControl}
          options=${[
            { value: "low", label: "Low" },
            { value: "medium", label: "Medium" },
            { value: "high", label: "High" },
            { value: "critical", label: "Critical" },
          ]}
          value=${priority}
          onChange=${(v) => {
            haptic();
            setPriority(v);
          }}
        />
        <button
          class="btn btn-primary"
          onClick=${handleSubmit}
          disabled=${submitting}
        >
          ${submitting ? "Creating…" : "Create Task"}
        </button>
      </div>
    <//>
  `;
}
