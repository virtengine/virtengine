/* ─────────────────────────────────────────────────────────────
 *  Tab: Agents — thread/slot cards, capacity, detail expansion
 * ────────────────────────────────────────────────────────────── */
import { h } from "preact";
import { useState, useCallback, useEffect, useRef } from "preact/hooks";
import htm from "htm";

const html = htm.bind(h);

import { haptic, showConfirm } from "../modules/telegram.js";
import { apiFetch, sendCommandToChat } from "../modules/api.js";
import {
  executorData,
  agentsData,
  agentLogQuery,
  agentLogFile,
  showToast,
  refreshTab,
  scheduleRefresh,
} from "../modules/state.js";
import { navigateTo } from "../modules/router.js";
import { ICONS } from "../modules/icons.js";
import { formatRelative, truncate } from "../modules/utils.js";
import {
  Card,
  Badge,
  StatCard,
  SkeletonCard,
  EmptyState,
} from "../components/shared.js";
import { ProgressBar } from "../components/charts.js";
import { Collapsible } from "../components/forms.js";
import {
  SessionList,
  loadSessions,
  selectedSessionId,
  sessionsData,
} from "../components/session-list.js";
import { ChatView } from "../components/chat-view.js";
import { DiffViewer } from "../components/diff-viewer.js";

/* ─── Status indicator helpers ─── */
function statusColor(s) {
  const map = {
    idle: "var(--color-todo)",
    busy: "var(--color-inprogress)",
    running: "var(--color-inprogress)",
    error: "var(--color-error)",
    done: "var(--color-done)",
  };
  return map[(s || "").toLowerCase()] || "var(--text-secondary)";
}

function StatusDot({ status }) {
  return html`<span
    class="status-dot"
    style="background:${statusColor(status)}"
  ></span>`;
}

/* ─── Duration formatting ─── */
function formatDuration(startedAt) {
  if (!startedAt) return "";
  const sec = Math.round((Date.now() - new Date(startedAt).getTime()) / 1000);
  if (sec < 60) return `${sec}s`;
  if (sec < 3600) return `${Math.floor(sec / 60)}m ${sec % 60}s`;
  return `${Math.floor(sec / 3600)}h ${Math.floor((sec % 3600) / 60)}m`;
}

/* ─── Workspace Viewer Modal ─── */
function WorkspaceViewer({ agent, onClose }) {
  const [logText, setLogText] = useState("Loading…");
  const [contextData, setContextData] = useState(null);
  const [steerInput, setSteerInput] = useState("");
  const logRef = useRef(null);

  const query = agent.branch || agent.taskId || "";

  useEffect(() => {
    if (!query) return;
    let active = true;

    const fetchLogs = () => {
      apiFetch(`/api/agent-logs/tail?query=${encodeURIComponent(query)}&lines=200`, { _silent: true })
        .then((res) => {
          if (!active) return;
          const data = res.data ?? res ?? "";
          setLogText(typeof data === "string" ? data : (data.lines || []).join("\n") || JSON.stringify(data, null, 2));
          if (logRef.current) logRef.current.scrollTop = logRef.current.scrollHeight;
        })
        .catch(() => { if (active) setLogText("(failed to load logs)"); });
    };

    const fetchContext = () => {
      apiFetch(`/api/agent-context?query=${encodeURIComponent(query)}`, { _silent: true })
        .then((res) => { if (active) setContextData(res.data ?? res ?? null); })
        .catch(() => {});
    };

    fetchLogs();
    fetchContext();
    const interval = setInterval(fetchLogs, 5000);
    return () => { active = false; clearInterval(interval); };
  }, [query]);

  const handleStop = async () => {
    const ok = await showConfirm(`Force-stop agent on "${truncate(agent.taskTitle || agent.taskId || "task", 40)}"?`);
    if (!ok) return;
    haptic("heavy");
    try {
      await apiFetch("/api/executor/stop-slot", {
        method: "POST",
        body: JSON.stringify({ slotIndex: agent.index, taskId: agent.taskId }),
      });
      showToast("Stop signal sent", "success");
      onClose();
      scheduleRefresh(200);
    } catch { /* toast via apiFetch */ }
  };

  const handleSteer = () => {
    if (!steerInput.trim()) return;
    sendCommandToChat(`/steer ${steerInput.trim()}`);
    showToast("Steer command sent", "success");
    setSteerInput("");
  };

  return html`
    <div class="modal-overlay" onClick=${(e) => e.target === e.currentTarget && onClose()}>
      <div class="modal-content">
        <div class="modal-handle" />
        <div class="workspace-viewer">
          <div class="workspace-header">
            <div>
              <div class="task-card-title">
                <${StatusDot} status=${agent.status || "busy"} />
                ${agent.taskTitle || "(no title)"}
              </div>
              <div class="task-card-meta">
                ${agent.branch || "?"} · Slot ${(agent.index ?? 0) + 1} · ${formatDuration(agent.startedAt)}
              </div>
            </div>
            <button class="btn btn-ghost btn-sm" onClick=${onClose}>✕</button>
          </div>

          <div class="workspace-log" ref=${logRef}>${logText}</div>

          ${contextData && html`
            <div style="padding:12px 16px;">
              <div class="card-subtitle">Workspace Context</div>
              ${contextData.changedFiles?.length > 0 && html`
                <div class="meta-text mb-sm">Changed: ${contextData.changedFiles.join(", ")}</div>
              `}
              ${contextData.diffSummary && html`
                <div class="meta-text">${contextData.diffSummary}</div>
              `}
              ${!contextData.changedFiles && !contextData.diffSummary && html`
                <div class="meta-text">No workspace context available.</div>
              `}
            </div>
          `}

          <div class="workspace-controls">
            <input
              class="input"
              placeholder="Steer agent…"
              value=${steerInput}
              onInput=${(e) => setSteerInput(e.target.value)}
              onKeyDown=${(e) => e.key === "Enter" && handleSteer()}
            />
            <button class="btn btn-primary btn-sm" onClick=${handleSteer}>🎯</button>
            <button class="btn btn-danger btn-sm" onClick=${handleStop}>⛔ Stop</button>
          </div>
        </div>
      </div>
    </div>
  `;
}

/* ─── Dispatch Section ─── */
function DispatchSection({ freeSlots }) {
  const [taskId, setTaskId] = useState("");
  const [prompt, setPrompt] = useState("");
  const [dispatching, setDispatching] = useState(false);

  const canDispatch = freeSlots > 0 && (taskId.trim() || prompt.trim());

  const handleDispatch = async () => {
    if (!canDispatch || dispatching) return;
    haptic();
    setDispatching(true);
    try {
      const body = taskId.trim()
        ? { taskId: taskId.trim() }
        : { prompt: prompt.trim() };
      const res = await apiFetch("/api/executor/dispatch", {
        method: "POST",
        body: JSON.stringify(body),
      });
      if (res.ok !== false) {
        showToast(`Dispatched to slot ${(res.slotIndex ?? 0) + 1}`, "success");
        setTaskId("");
        setPrompt("");
        scheduleRefresh(200);
      }
    } catch {
      /* toast via apiFetch */
    } finally {
      setDispatching(false);
    }
  };

  return html`
    <${Card} title="Dispatch Agent">
      <div class="dispatch-section">
        <div class="meta-text mb-sm">
          ${freeSlots > 0
            ? `${freeSlots} slot${freeSlots > 1 ? "s" : ""} available`
            : "No free slots"}
        </div>
        <div class="input-row">
          <input
            class="input"
            placeholder="Task ID"
            value=${taskId}
            onInput=${(e) => { setTaskId(e.target.value); if (e.target.value) setPrompt(""); }}
          />
        </div>
        <div class="divider-label">or</div>
        <textarea
          class="input"
          placeholder="Freeform prompt…"
          rows="2"
          value=${prompt}
          onInput=${(e) => { setPrompt(e.target.value); if (e.target.value) setTaskId(""); }}
        />
        <button
          class="btn btn-primary"
          disabled=${!canDispatch || dispatching}
          onClick=${handleDispatch}
        >
          ${dispatching ? "Dispatching…" : "🚀 Dispatch"}
        </button>
      </div>
    <//>
  `;
}

/* ─── AgentsTab ─── */
export function AgentsTab() {
  const executor = executorData.value;
  const agents = agentsData?.value || [];
  const execData = executor?.data;
  const slots = execData?.slots || [];
  const maxParallel = execData?.maxParallel || 0;
  const activeSlots = execData?.activeSlots || 0;

  const [expandedSlot, setExpandedSlot] = useState(null);
  const [selectedAgent, setSelectedAgent] = useState(null);

  /* Navigate to logs tab with agent query pre-filled */
  const viewAgentLogs = (query) => {
    haptic();
    if (agentLogQuery) agentLogQuery.value = query;
    if (agentLogFile) agentLogFile.value = "";
    navigateTo("logs");
  };

  /* Force stop a specific agent slot */
  const handleForceStop = async (slot) => {
    const ok = await showConfirm(
      `Force-stop agent working on "${truncate(slot.taskTitle || slot.taskId || "task", 40)}"?`,
    );
    if (!ok) return;
    haptic("heavy");
    try {
      await apiFetch("/api/executor/stop-slot", {
        method: "POST",
        body: JSON.stringify({ slotIndex: slot.index, taskId: slot.taskId }),
      });
      showToast("Stop signal sent", "success");
      scheduleRefresh(200);
    } catch {
      /* toast via apiFetch */
    }
  };

  /* Toggle expanded detail view for a slot */
  const toggleExpand = (i) => {
    haptic();
    setExpandedSlot(expandedSlot === i ? null : i);
  };

  /* Open workspace viewer for an agent */
  const openWorkspace = (slot, i) => {
    haptic();
    setSelectedAgent({ ...slot, index: i });
  };

  /* Capacity utilisation */
  const freeSlots = Math.max(0, maxParallel - activeSlots);
  const capacityPct =
    maxParallel > 0 ? Math.round((activeSlots / maxParallel) * 100) : 0;

  /* Aggregate stats */
  const totalCompleted = slots.reduce((n, s) => n + (s.completedCount || 0), 0);
  const avgTimeMs = slots.length
    ? slots.reduce((n, s) => n + (s.avgDurationMs || 0), 0) / slots.length
    : 0;
  const avgTimeStr = avgTimeMs > 0 ? `${Math.round(avgTimeMs / 1000)}s` : "—";

  /* Loading state */
  if (!executor && !agents.length)
    return html`<${Card} title="Loading…"><${SkeletonCard} count=${3} /><//>`;

  return html`
    <!-- Dispatch section -->
    <${DispatchSection} freeSlots=${freeSlots} />

    <!-- Capacity overview -->
    <${Card} title="Agent Capacity">
      <div class="stats-grid mb-sm">
        <${StatCard}
          value=${activeSlots}
          label="Active"
          color="var(--color-inprogress)"
        />
        <${StatCard} value=${maxParallel} label="Max" />
        <${StatCard}
          value=${totalCompleted}
          label="Completed"
          color="var(--color-done)"
        />
        <${StatCard} value=${avgTimeStr} label="Avg Time" />
      </div>
      <${ProgressBar} percent=${capacityPct} />
      <div class="meta-text text-center mt-xs">
        ${capacityPct}% capacity used
      </div>
    <//>

    <!-- Visual slot grid -->
    <${Card} title="Slot Grid">
      <div class="slot-grid">
        ${Array.from(
          { length: Math.max(maxParallel, slots.length, 1) },
          (_, i) => {
            const slot = slots[i];
            const st = slot ? slot.status || "busy" : "idle";
            return html`
              <div
                key=${i}
                class="slot-cell slot-${st}"
                title=${slot
                  ? `${slot.taskTitle || slot.taskId} (${st})`
                  : `Slot ${i + 1} idle`}
                onClick=${() => slot && openWorkspace(slot, i)}
              >
                <${StatusDot} status=${st} />
                <span class="slot-label">${i + 1}</span>
              </div>
            `;
          },
        )}
      </div>
    <//>

    <!-- Active agents / slots -->
    <${Card} title="Active Agents">
      ${slots.length
        ? slots.map(
            (slot, i) => html`
              <div
                key=${i}
                class="task-card ${expandedSlot === i
                  ? "task-card-expanded"
                  : ""}"
              >
                <div
                  class="task-card-header"
                  onClick=${() => toggleExpand(i)}
                  style="cursor:pointer"
                >
                  <div>
                    <div class="task-card-title">
                      <${StatusDot} status=${slot.status || "busy"} />
                      ${slot.taskTitle || "(no title)"}
                    </div>
                    <div class="task-card-meta">
                      ${slot.taskId || "?"} · Agent
                      ${slot.agentInstanceId || "n/a"} · ${slot.sdk || "?"}
                    </div>
                  </div>
                  <${Badge}
                    status=${slot.status || "busy"}
                    text=${slot.status || "busy"}
                  />
                </div>
                <div class="flex-between">
                  <div class="meta-text">Attempt ${slot.attempt || 1}</div>
                  ${slot.startedAt && html`
                    <div class="agent-duration">${formatDuration(slot.startedAt)}</div>
                  `}
                </div>

                <!-- Progress indicator for active tasks -->
                ${(slot.status === "running" || slot.status === "busy") &&
                html`
                  <div class="agent-progress-bar mt-sm">
                    <div
                      class="agent-progress-bar-fill agent-progress-pulse"
                    ></div>
                  </div>
                `}

                <!-- Expanded detail -->
                ${expandedSlot === i &&
                html`
                  <div class="agent-detail mt-sm">
                    ${slot.branch &&
                    html`<div class="meta-text">Branch: ${slot.branch}</div>`}
                    ${slot.startedAt &&
                    html`<div class="meta-text">
                      Started: ${formatRelative(slot.startedAt)}
                    </div>`}
                    ${slot.completedCount != null &&
                    html`<div class="meta-text">
                      Completed: ${slot.completedCount} tasks
                    </div>`}
                    ${slot.avgDurationMs &&
                    html`<div class="meta-text">
                      Avg: ${Math.round(slot.avgDurationMs / 1000)}s
                    </div>`}
                    ${slot.lastError &&
                    html`<div
                      class="meta-text"
                      style="color:var(--color-error)"
                    >
                      Last error: ${truncate(slot.lastError, 100)}
                    </div>`}
                  </div>
                `}

                <div class="btn-row mt-sm">
                  <button
                    class="btn btn-ghost btn-sm"
                    onClick=${() =>
                      viewAgentLogs(
                        (slot.taskId || slot.branch || "").slice(0, 12),
                      )}
                  >
                    📄 Logs
                  </button>
                  <button
                    class="btn btn-ghost btn-sm"
                    onClick=${() =>
                      sendCommandToChat(
                        `/steer focus on ${slot.taskTitle || slot.taskId}`,
                      )}
                  >
                    🎯 Steer
                  </button>
                  <button
                    class="btn btn-ghost btn-sm"
                    onClick=${() => openWorkspace(slot, i)}
                  >
                    🔍 View
                  </button>
                  <button
                    class="btn btn-danger btn-sm"
                    onClick=${() => handleForceStop({ ...slot, index: i })}
                  >
                    ⛔ Stop
                  </button>
                </div>
              </div>
            `,
          )
        : html`<${EmptyState} message="No active agents." />`}
    <//>

    <!-- Agent threads (if separate from slots) -->
    ${agents.length > 0 &&
    html`
      <${Collapsible} title="Agent Threads" defaultOpen=${false}>
        <${Card}>
          <div class="stats-grid">
            ${agents.map(
              (t, i) => html`
                <${StatCard}
                  key=${i}
                  value=${t.turnCount || 0}
                  label="${truncate(t.taskKey || `Thread ${i}`, 20)} (${t.sdk ||
                  "?"})"
                />
              `,
            )}
          </div>
        <//>
      <//>
    `}

    <!-- Workspace viewer modal -->
    ${selectedAgent && html`
      <${WorkspaceViewer}
        agent=${selectedAgent}
        onClose=${() => setSelectedAgent(null)}
      />
    `}

    <!-- Sessions panel -->
    <${SessionsPanel} />
  `;
}

/* ─── Context Viewer for session detail tab ─── */
function ContextViewer({ sessionId }) {
  const [ctx, setCtx] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const intervalRef = useRef(null);

  const fetchContext = useCallback(() => {
    if (!sessionId) return;
    apiFetch(`/api/agent-context?query=${encodeURIComponent(sessionId)}`, { _silent: true })
      .then((res) => {
        const d = res.data ?? res ?? null;
        setCtx(d);
        setLoading(false);
        setError(null);
      })
      .catch((err) => {
        setLoading(false);
        setError(err.message || "Failed to load context");
      });
  }, [sessionId]);

  useEffect(() => {
    setLoading(true);
    setError(null);
    setCtx(null);
    fetchContext();
    intervalRef.current = setInterval(fetchContext, 10000);
    return () => { if (intervalRef.current) clearInterval(intervalRef.current); };
  }, [fetchContext]);

  const parseCommits = (detailed) => {
    if (!detailed) return [];
    return detailed.split("\n").filter(Boolean).map((line) => {
      const parts = line.split("||");
      return { hash: parts[0] || "", message: parts[1] || "", time: parts[2] || "" };
    });
  };

  const parseStatus = (raw) => {
    if (!raw) return [];
    return raw.split("\n").filter(Boolean).map((line) => {
      const code = line.substring(0, 2).trim() || "?";
      const file = line.substring(3);
      return { code, file };
    });
  };

  const parseAheadBehind = (raw) => {
    if (!raw) return { ahead: 0, behind: 0 };
    const parts = raw.split(/\s+/);
    return { ahead: parseInt(parts[0], 10) || 0, behind: parseInt(parts[1], 10) || 0 };
  };

  const statusColor = (code) => {
    if (code === "M" || code === "MM") return "var(--color-inprogress)";
    if (code === "A") return "var(--color-done)";
    if (code === "D") return "var(--color-error)";
    if (code === "?" || code === "??") return "var(--text-secondary)";
    return "var(--text-primary)";
  };

  const statusLabel = (code) => {
    const map = { M: "Modified", MM: "Modified", A: "Added", D: "Deleted", "?": "Untracked", "??": "Untracked", R: "Renamed", C: "Copied" };
    return map[code] || code;
  };

  const copyContext = () => {
    if (!ctx?.context) return;
    const c = ctx.context;
    const ab = parseAheadBehind(c.gitAheadBehind);
    const commits = parseCommits(c.gitLogDetailed);
    const files = parseStatus(c.gitStatus);
    let text = `## Workspace Context\n`;
    text += `Branch: ${c.gitBranch || "unknown"}\n`;
    text += `Path: ${c.path || "unknown"}\n`;
    text += `Status: ${files.length === 0 ? "Clean" : `${files.length} changed file(s)`}\n`;
    if (ab.ahead || ab.behind) text += `Ahead: ${ab.ahead}, Behind: ${ab.behind}\n`;
    if (commits.length) {
      text += `\n### Recent Commits\n`;
      commits.forEach((cm) => { text += `${cm.hash} ${cm.message} (${cm.time})\n`; });
    }
    if (files.length) {
      text += `\n### Modified Files\n`;
      files.forEach((f) => { text += `[${f.code}] ${f.file}\n`; });
    }
    navigator.clipboard.writeText(text).then(() => showToast("Context copied", "success")).catch(() => showToast("Copy failed", "error"));
  };

  if (loading) {
    return html`<div class="chat-view" style="padding:16px;">
      <${SkeletonCard} height="40px" />
      <${SkeletonCard} height="120px" className="mt-sm" />
      <${SkeletonCard} height="80px" className="mt-sm" />
    </div>`;
  }

  if (error) {
    return html`<div class="chat-view chat-empty-state">
      <div class="session-empty-icon" style="color:var(--color-error)">⚠️</div>
      <div class="session-empty-text">${error}</div>
      <button class="btn btn-primary btn-sm mt-sm" onClick=${() => { setLoading(true); setError(null); fetchContext(); }}>🔄 Retry</button>
    </div>`;
  }

  if (!ctx?.context) {
    return html`<div class="chat-view chat-empty-state">
      <div class="session-empty-icon">📋</div>
      <div class="session-empty-text">No context available for this session</div>
    </div>`;
  }

  const c = ctx.context;
  const ab = parseAheadBehind(c.gitAheadBehind);
  const commits = parseCommits(c.gitLogDetailed);
  const files = parseStatus(c.gitStatus);
  const isDirty = files.length > 0;

  return html`
    <div class="chat-view" style="padding:12px; overflow-y:auto;">
      <!-- Toolbar -->
      <div style="display:flex; gap:8px; justify-content:flex-end; margin-bottom:12px;">
        <button class="btn btn-ghost btn-sm" onClick=${() => { setLoading(true); fetchContext(); }}>
          <span class="icon-inline">${ICONS.refresh}</span> Refresh
        </button>
        <button class="btn btn-ghost btn-sm" onClick=${copyContext}>
          <span class="icon-inline">${ICONS.copy}</span> Copy Context
        </button>
      </div>

      <!-- Branch & Status -->
      <div class="card mb-sm">
        <div class="card-title" style="display:flex; align-items:center; gap:8px;">
          <span class="icon-inline">${ICONS.git}</span> Branch & Status
        </div>
        <div style="display:flex; flex-wrap:wrap; gap:12px; margin-top:8px;">
          <div style="flex:1; min-width:120px;">
            <div class="meta-text">Branch</div>
            <div style="font-weight:600; font-family:monospace; font-size:13px;">${c.gitBranch || "unknown"}</div>
          </div>
          <div>
            <div class="meta-text">Status</div>
            <${Badge}
              status=${isDirty ? "inprogress" : "done"}
              text=${isDirty ? `${files.length} changed` : "Clean"}
            />
          </div>
          ${(ab.ahead > 0 || ab.behind > 0) && html`
            <div>
              <div class="meta-text">Sync</div>
              <div style="font-size:13px;">
                ${ab.ahead > 0 ? html`<span style="color:var(--color-done)">↑${ab.ahead}</span>` : null}
                ${ab.ahead > 0 && ab.behind > 0 ? " " : null}
                ${ab.behind > 0 ? html`<span style="color:var(--color-error)">↓${ab.behind}</span>` : null}
              </div>
            </div>
          `}
        </div>
      </div>

      <!-- Working Directory -->
      <div class="card mb-sm">
        <div class="card-title" style="display:flex; align-items:center; gap:8px;">
          <span class="icon-inline">${ICONS.folder}</span> Working Directory
        </div>
        <div style="font-family:monospace; font-size:12px; color:var(--text-secondary); margin-top:6px; word-break:break-all;">
          ${c.path || "unknown"}
        </div>
      </div>

      <!-- Recent Commits -->
      ${commits.length > 0 && html`
        <div class="card mb-sm">
          <div class="card-title" style="display:flex; align-items:center; gap:8px;">
            <span class="icon-inline">${ICONS.clock}</span> Recent Commits
          </div>
          <div style="margin-top:8px;">
            ${commits.map((cm) => html`
              <div key=${cm.hash} style="display:flex; gap:8px; align-items:baseline; padding:4px 0; border-bottom:1px solid var(--border-color, rgba(255,255,255,0.06));">
                <code style="color:var(--color-inprogress); font-size:12px; flex-shrink:0;">${cm.hash}</code>
                <span style="flex:1; font-size:13px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;">${cm.message}</span>
                <span class="meta-text" style="flex-shrink:0; font-size:11px;">${cm.time}</span>
              </div>
            `)}
          </div>
        </div>
      `}

      <!-- Modified Files -->
      ${files.length > 0 && html`
        <div class="card mb-sm">
          <div class="card-title" style="display:flex; align-items:center; gap:8px;">
            <span class="icon-inline">${ICONS.edit}</span> Modified Files
            <${Badge} text="${files.length}" className="ml-auto" />
          </div>
          <div style="margin-top:8px;">
            ${files.map((f) => html`
              <div key=${f.file} style="display:flex; gap:8px; align-items:center; padding:4px 0; border-bottom:1px solid var(--border-color, rgba(255,255,255,0.06));">
                <code style="color:${statusColor(f.code)}; font-size:11px; font-weight:700; min-width:20px; text-align:center;" title=${statusLabel(f.code)}>${f.code}</code>
                <span style="font-family:monospace; font-size:12px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;">${f.file}</span>
              </div>
            `)}
          </div>
        </div>
      `}

      <!-- Diff Stats -->
      ${c.gitDiffStat && html`
        <div class="card mb-sm">
          <div class="card-title" style="display:flex; align-items:center; gap:8px;">
            <span class="icon-inline">${ICONS.terminal}</span> Diff Summary
          </div>
          <pre style="font-size:11px; margin:8px 0 0; white-space:pre-wrap; color:var(--text-secondary); overflow-x:auto;">${c.gitDiffStat}</pre>
        </div>
      `}
    </div>
  `;
}

/* ─── Sessions Panel — split view with list + detail ─── */
function SessionsPanel() {
  const [detailTab, setDetailTab] = useState("chat");
  const sessionId = selectedSessionId.value;

  const handleBack = useCallback(() => {
    selectedSessionId.value = null;
  }, []);

  return html`
    <${Card} title="Sessions">
      <div class="session-split">
        <${SessionList} onSelect=${() => setDetailTab("chat")} />
        <div class="session-detail">
          ${sessionId && html`
            <button class="session-back-btn" onClick=${handleBack}>
              ← Back to sessions
            </button>
            <div class="session-detail-tabs">
              <button
                class="session-detail-tab ${detailTab === "chat" ? "active" : ""}"
                onClick=${() => setDetailTab("chat")}
              >💬 Chat</button>
              <button
                class="session-detail-tab ${detailTab === "diff" ? "active" : ""}"
                onClick=${() => setDetailTab("diff")}
              >📝 Diff</button>
              <button
                class="session-detail-tab ${detailTab === "context" ? "active" : ""}"
                onClick=${() => setDetailTab("context")}
              >📋 Context</button>
            </div>
          `}
          ${detailTab === "chat" && html`<${ChatView} sessionId=${sessionId} />`}
          ${detailTab === "diff" && sessionId && html`<${DiffViewer} sessionId=${sessionId} />`}
          ${detailTab === "context" && sessionId && html`<${ContextViewer} sessionId=${sessionId} />`}
          ${!sessionId && detailTab !== "chat" && html`
            <div class="chat-view chat-empty-state">
              <div class="session-empty-icon">💬</div>
              <div class="session-empty-text">Select a session</div>
            </div>
          `}
        </div>
      </div>
    <//>
  `;
}
