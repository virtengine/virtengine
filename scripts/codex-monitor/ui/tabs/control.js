/* ─────────────────────────────────────────────────────────────
 *  Tab: Control — executor, commands, routing, quick commands
 * ────────────────────────────────────────────────────────────── */
import { h } from "preact";
import { useState, useCallback, useEffect, useRef } from "preact/hooks";
import htm from "htm";

const html = htm.bind(h);

import { haptic, showConfirm } from "../modules/telegram.js";
import { apiFetch, sendCommandToChat } from "../modules/api.js";
import {
  executorData,
  configData,
  loadConfig,
  showToast,
  refreshTab,
  runOptimistic,
  scheduleRefresh,
} from "../modules/state.js";
import { ICONS } from "../modules/icons.js";
import { cloneValue } from "../modules/utils.js";
import { Card, Badge, SkeletonCard } from "../components/shared.js";
import { SegmentedControl, SliderControl } from "../components/forms.js";

/* ─── Command registry for autocomplete ─── */
const CMD_REGISTRY = [
  { cmd: '/status', desc: 'Show orchestrator status', cat: 'System' },
  { cmd: '/health', desc: 'Health check', cat: 'System' },
  { cmd: '/menu', desc: 'Show command menu', cat: 'System' },
  { cmd: '/helpfull', desc: 'Full help text', cat: 'System' },
  { cmd: '/plan', desc: 'Generate execution plan', cat: 'Tasks' },
  { cmd: '/logs', desc: 'View recent logs', cat: 'Logs' },
  { cmd: '/diff', desc: 'View git diff', cat: 'Git' },
  { cmd: '/steer', desc: 'Steer active agent', cat: 'Agent' },
  { cmd: '/ask', desc: 'Ask agent a question', cat: 'Agent' },
  { cmd: '/start', desc: 'Start a task', cat: 'Tasks' },
  { cmd: '/retry', desc: 'Retry failed task', cat: 'Tasks' },
  { cmd: '/cancel', desc: 'Cancel running task', cat: 'Tasks' },
  { cmd: '/shell', desc: 'Execute shell command', cat: 'Shell' },
  { cmd: '/git', desc: 'Execute git command', cat: 'Git' },
];

/* ─── Category badge colors ─── */
const CAT_COLORS = {
  System: '#6366f1', Tasks: '#f59e0b', Logs: '#10b981',
  Git: '#f97316', Agent: '#8b5cf6', Shell: '#64748b',
};

/* ─── Persistent history key & limits ─── */
const HISTORY_KEY = 've-cmd-history';
const MAX_HISTORY = 50;
const MAX_OUTPUTS = 3;
const POLL_INTERVAL = 2000;
const MAX_POLLS = 7;

/* ─── ControlTab ─── */
export function ControlTab() {
  const executor = executorData.value;
  const execData = executor?.data;
  const mode = executor?.mode || "vk";
  const config = configData.value;

  /* Form inputs */
  const [commandInput, setCommandInput] = useState("");
  const [startTaskInput, setStartTaskInput] = useState("");
  const [retryInput, setRetryInput] = useState("");
  const [askInput, setAskInput] = useState("");
  const [steerInput, setSteerInput] = useState("");
  const [quickCmdInput, setQuickCmdInput] = useState("");
  const [quickCmdPrefix, setQuickCmdPrefix] = useState("shell");
  const [quickCmdFeedback, setQuickCmdFeedback] = useState("");
  const [maxParallel, setMaxParallel] = useState(execData?.maxParallel ?? 0);
  const [cmdHistory, setCmdHistory] = useState([]);
  const [showHistory, setShowHistory] = useState(false);

  /* ── Autocomplete state ── */
  const [acItems, setAcItems] = useState([]);
  const [acIndex, setAcIndex] = useState(-1);
  const [showAc, setShowAc] = useState(false);

  /* ── Persistent history state ── */
  const [historyIndex, setHistoryIndex] = useState(-1);
  const savedInputRef = useRef("");

  /* ── Inline output state ── */
  const [cmdOutputs, setCmdOutputs] = useState([]);
  const [runningCmd, setRunningCmd] = useState(null);
  const [expandedOutputs, setExpandedOutputs] = useState({});
  const pollRef = useRef(null);

  /* ── Load persistent history on mount ── */
  useEffect(() => {
    try {
      const saved = localStorage.getItem(HISTORY_KEY);
      if (saved) {
        const parsed = JSON.parse(saved);
        if (Array.isArray(parsed)) setCmdHistory(parsed.slice(0, MAX_HISTORY));
      }
    } catch (_) { /* ignore corrupt data */ }
    return () => { if (pollRef.current) clearInterval(pollRef.current); };
  }, []);

  /* ── Autocomplete filter ── */
  useEffect(() => {
    if (commandInput.startsWith('/') && commandInput.length > 0) {
      const q = commandInput.toLowerCase();
      const matches = CMD_REGISTRY.filter((r) => r.cmd.toLowerCase().includes(q));
      setAcItems(matches);
      setAcIndex(-1);
      setShowAc(matches.length > 0);
    } else {
      setShowAc(false);
      setAcItems([]);
      setAcIndex(-1);
    }
  }, [commandInput]);

  /* ── Command history helper (persistent) ── */
  const pushHistory = useCallback((cmd) => {
    setCmdHistory((prev) => {
      const next = [cmd, ...prev.filter((c) => c !== cmd)].slice(0, MAX_HISTORY);
      try { localStorage.setItem(HISTORY_KEY, JSON.stringify(next)); } catch (_) {}
      return next;
    });
  }, []);

  /* ── Inline output polling ── */
  const startOutputPolling = useCallback((cmd) => {
    if (pollRef.current) clearInterval(pollRef.current);
    const ts = new Date().toISOString();
    setRunningCmd(cmd);
    let pollCount = 0;
    let lastContent = '';

    pollRef.current = setInterval(async () => {
      pollCount++;
      try {
        const res = await apiFetch('/api/logs?lines=15', { _silent: true });
        const text = typeof res === 'string' ? res : (res?.logs || res?.data || JSON.stringify(res, null, 2));
        if (text === lastContent || pollCount >= MAX_POLLS) {
          clearInterval(pollRef.current);
          pollRef.current = null;
          setRunningCmd(null);
          setCmdOutputs((prev) => {
            const entry = { cmd, ts, output: text || '(no output)' };
            const next = [entry, ...prev].slice(0, MAX_OUTPUTS);
            return next;
          });
          setExpandedOutputs((prev) => ({ ...prev, [0]: true }));
        }
        lastContent = text;
      } catch (_) {
        clearInterval(pollRef.current);
        pollRef.current = null;
        setRunningCmd(null);
        setCmdOutputs((prev) => {
          const entry = { cmd, ts, output: '(failed to fetch output)' };
          return [entry, ...prev].slice(0, MAX_OUTPUTS);
        });
      }
    }, POLL_INTERVAL);
  }, []);

  const sendCmd = useCallback(
    (cmd) => {
      if (!cmd.trim()) return;
      sendCommandToChat(cmd.trim());
      pushHistory(cmd.trim());
      setHistoryIndex(-1);
      startOutputPolling(cmd.trim());
    },
    [pushHistory, startOutputPolling],
  );

  /* ── Config update helper ── */
  const updateConfig = useCallback(
    async (key, value) => {
      haptic();
      try {
        await apiFetch("/api/config/update", {
          method: "POST",
          body: JSON.stringify({ key, value }),
        });
        await loadConfig();
        showToast(`${key} → ${value}`, "success");
      } catch {
        showToast(`Failed to update ${key}`, "error");
      }
    },
    [],
  );

  /* ── Executor controls ── */
  const handlePause = async () => {
    const ok = await showConfirm(
      "Pause the executor? Running tasks will finish but no new tasks will start.",
    );
    if (!ok) return;
    haptic("medium");
    const prev = cloneValue(executor);
    await runOptimistic(
      () => {
        if (executorData.value)
          executorData.value = { ...executorData.value, paused: true };
      },
      () => apiFetch("/api/executor/pause", { method: "POST" }),
      () => {
        executorData.value = prev;
      },
    ).catch(() => {});
    scheduleRefresh(120);
  };

  const handleResume = async () => {
    haptic("medium");
    const prev = cloneValue(executor);
    await runOptimistic(
      () => {
        if (executorData.value)
          executorData.value = { ...executorData.value, paused: false };
      },
      () => apiFetch("/api/executor/resume", { method: "POST" }),
      () => {
        executorData.value = prev;
      },
    ).catch(() => {});
    scheduleRefresh(120);
  };

  const handleMaxParallel = async (value) => {
    setMaxParallel(value);
    haptic();
    const prev = cloneValue(executor);
    await runOptimistic(
      () => {
        if (executorData.value?.data)
          executorData.value.data.maxParallel = value;
      },
      () =>
        apiFetch("/api/executor/maxparallel", {
          method: "POST",
          body: JSON.stringify({ value }),
        }),
      () => {
        executorData.value = prev;
      },
    ).catch(() => {});
    scheduleRefresh(120);
  };

  /* ── Region options from config ── */
  const regions = config?.regions || ["auto"];
  const regionOptions = regions.map((r) => ({
    value: r,
    label: r.charAt(0).toUpperCase() + r.slice(1),
  }));

  /* ── Quick command submit ── */
  const handleQuickCmd = useCallback(() => {
    const input = quickCmdInput.trim();
    if (!input) return;
    const cmd = `/${quickCmdPrefix} ${input}`;
    sendCmd(cmd);
    setQuickCmdInput("");
    setQuickCmdFeedback("✓ Command sent to monitor");
    setTimeout(() => setQuickCmdFeedback(""), 4000);
  }, [quickCmdInput, quickCmdPrefix, sendCmd]);

  /* ── Autocomplete select helper ── */
  const selectAcItem = useCallback((item) => {
    setCommandInput(item.cmd + ' ');
    setShowAc(false);
    setAcIndex(-1);
  }, []);

  /* ── Console input keydown handler ── */
  const handleConsoleKeyDown = useCallback((e) => {
    // Autocomplete navigation
    if (showAc && acItems.length > 0) {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setAcIndex((prev) => (prev + 1) % acItems.length);
        return;
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault();
        setAcIndex((prev) => (prev <= 0 ? acItems.length - 1 : prev - 1));
        return;
      }
      if (e.key === 'Enter' && acIndex >= 0) {
        e.preventDefault();
        selectAcItem(acItems[acIndex]);
        return;
      }
      if (e.key === 'Escape') {
        e.preventDefault();
        setShowAc(false);
        return;
      }
    }

    // History navigation (when input is empty or already in history mode)
    if (!showAc && (commandInput === '' || historyIndex >= 0)) {
      if (e.key === 'ArrowUp' && cmdHistory.length > 0) {
        e.preventDefault();
        const nextIdx = historyIndex + 1;
        if (nextIdx < cmdHistory.length) {
          if (historyIndex === -1) savedInputRef.current = commandInput;
          setHistoryIndex(nextIdx);
          setCommandInput(cmdHistory[nextIdx]);
        }
        return;
      }
      if (e.key === 'ArrowDown' && historyIndex >= 0) {
        e.preventDefault();
        const nextIdx = historyIndex - 1;
        if (nextIdx < 0) {
          setHistoryIndex(-1);
          setCommandInput(savedInputRef.current);
        } else {
          setHistoryIndex(nextIdx);
          setCommandInput(cmdHistory[nextIdx]);
        }
        return;
      }
    }

    // Submit
    if (e.key === 'Enter' && commandInput.trim()) {
      sendCmd(commandInput.trim());
      setCommandInput('');
      setShowAc(false);
    }
  }, [showAc, acItems, acIndex, commandInput, historyIndex, cmdHistory, sendCmd, selectAcItem]);

  /* ── Toggle output accordion ── */
  const toggleOutput = useCallback((idx) => {
    setExpandedOutputs((prev) => ({ ...prev, [idx]: !prev[idx] }));
  }, []);

  return html`
    <!-- Loading skeleton -->
    ${!executor && !config && html`<${Card} title="Loading…"><${SkeletonCard} /><//>`}

    <!-- ── Executor Controls ── -->
    <${Card} title="Executor Controls">
      <div class="meta-text mb-sm">
        Mode: <strong>${mode}</strong> · Slots:
        ${execData?.activeSlots ?? 0}/${execData?.maxParallel ?? "—"} ·
        ${executor?.paused
          ? html`<${Badge} status="error" text="Paused" />`
          : html`<${Badge} status="done" text="Running" />`}
      </div>
      <div class="meta-text mb-sm">
        Poll:
        ${execData?.pollIntervalMs ? execData.pollIntervalMs / 1000 : "—"}s ·
        Timeout:
        ${execData?.taskTimeoutMs
          ? Math.round(execData.taskTimeoutMs / 60000)
          : "—"}m
      </div>

      <div class="range-row mb-md">
        <input
          type="range"
          min="0"
          max="20"
          step="1"
          value=${maxParallel}
          onInput=${(e) => setMaxParallel(Number(e.target.value))}
          onChange=${(e) => handleMaxParallel(Number(e.target.value))}
        />
        <span class="pill">Max ${maxParallel}</span>
      </div>

      <div class="btn-row">
        <button class="btn btn-primary btn-sm" onClick=${handlePause}>
          ⏸ Pause
        </button>
        <button class="btn btn-secondary btn-sm" onClick=${handleResume}>
          ▶ Resume
        </button>
        <button
          class="btn btn-ghost btn-sm"
          onClick=${() => sendCmd("/executor")}
        >
          /executor
        </button>
      </div>
    <//>

    <!-- ── Command Console ── -->
    <${Card} title="Command Console">
      <div class="input-row mb-sm">
        <div style="position:relative;flex:1">
          <input
            class="input"
            placeholder="/status"
            value=${commandInput}
            onInput=${(e) => {
              setCommandInput(e.target.value);
              setHistoryIndex(-1);
            }}
            onFocus=${() => setShowHistory(true)}
            onBlur=${() => setTimeout(() => { setShowHistory(false); setShowAc(false); }, 200)}
            onKeyDown=${handleConsoleKeyDown}
          />
          <!-- Autocomplete dropdown (above input) -->
          ${showAc && acItems.length > 0 && html`
            <div class="cmd-dropdown">
              ${acItems.map((item, i) => html`
                <div
                  key=${item.cmd}
                  class="cmd-dropdown-item${i === acIndex ? ' selected' : ''}"
                  onMouseDown=${(e) => { e.preventDefault(); selectAcItem(item); }}
                  onMouseEnter=${() => setAcIndex(i)}
                >
                  <div>
                    <span style="font-weight:600;color:#e2e8f0">${item.cmd}</span>
                    <span style="margin-left:8px;color:#94a3b8;font-size:0.85em">${item.desc}</span>
                  </div>
                  <span style=${{
                    fontSize: '0.7rem', padding: '2px 8px', borderRadius: '9999px',
                    background: (CAT_COLORS[item.cat] || '#6366f1') + '33',
                    color: CAT_COLORS[item.cat] || '#6366f1', fontWeight: 600,
                  }}>${item.cat}</span>
                </div>
              `)}
            </div>
          `}
          <!-- Command history dropdown (legacy, when no autocomplete) -->
          ${!showAc && showHistory &&
          cmdHistory.length > 0 &&
          html`
            <div class="cmd-history-dropdown">
              ${cmdHistory.map(
                (c, i) => html`
                  <button
                    key=${i}
                    class="cmd-history-item"
                    onMouseDown=${(e) => {
                      e.preventDefault();
                      setCommandInput(c);
                      setShowHistory(false);
                    }}
                  >
                    ${c}
                  </button>
                `,
              )}
            </div>
          `}
        </div>
        <button
          class="btn btn-primary btn-sm"
          onClick=${() => {
            if (commandInput.trim()) {
              sendCmd(commandInput.trim());
              setCommandInput("");
            }
          }}
        >
          ${ICONS.send}
        </button>
      </div>

      <!-- Quick command chips -->
      <div class="btn-row">
        ${["/status", "/health", "/menu", "/helpfull"].map(
          (cmd) => html`
            <button
              key=${cmd}
              class="btn btn-ghost btn-sm"
              onClick=${() => sendCmd(cmd)}
            >
              ${cmd}
            </button>
          `,
        )}
      </div>

      <!-- Running indicator -->
      ${runningCmd && html`
        <div style="margin-top:8px;display:flex;align-items:center;gap:8px;color:#94a3b8;font-size:0.85rem">
          <span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:#facc15;animation:pulse 1s infinite"></span>
          Running: <code style="color:#e2e8f0">${runningCmd}</code>
        </div>
      `}

      <!-- Inline command outputs accordion -->
      ${cmdOutputs.length > 0 && html`
        <div style="margin-top:12px">
          ${cmdOutputs.map((entry, idx) => html`
            <div key=${idx} style="margin-bottom:6px;border:1px solid rgba(255,255,255,0.06);border-radius:8px;overflow:hidden">
              <button
                style="width:100%;text-align:left;padding:6px 12px;background:rgba(255,255,255,0.03);border:none;color:#cbd5e1;cursor:pointer;display:flex;justify-content:space-between;align-items:center;font-size:0.8rem"
                onClick=${() => toggleOutput(idx)}
              >
                <span><code style="color:#818cf8">${entry.cmd}</code></span>
                <span style="color:#64748b;font-size:0.75rem">${new Date(entry.ts).toLocaleTimeString()} ${expandedOutputs[idx] ? '▲' : '▼'}</span>
              </button>
              ${expandedOutputs[idx] && html`
                <div class="cmd-output-panel">${entry.output}</div>
              `}
            </div>
          `)}
        </div>
      `}
    <//>

    <!-- ── Task Ops ── -->
    <${Card} title="Task Ops">
      <div class="input-row mb-sm">
        <input
          class="input"
          placeholder="Task ID"
          value=${startTaskInput}
          onInput=${(e) => setStartTaskInput(e.target.value)}
        />
        <button
          class="btn btn-secondary btn-sm"
          onClick=${() => {
            if (startTaskInput.trim())
              sendCmd(`/starttask ${startTaskInput.trim()}`);
          }}
        >
          ▶ Start
        </button>
      </div>
      <div class="input-row">
        <input
          class="input"
          placeholder="Retry reason"
          value=${retryInput}
          onInput=${(e) => setRetryInput(e.target.value)}
        />
        <button
          class="btn btn-secondary btn-sm"
          onClick=${() =>
            sendCmd(
              retryInput.trim() ? `/retry ${retryInput.trim()}` : "/retry",
            )}
        >
          ↻ Retry
        </button>
        <button class="btn btn-ghost btn-sm" onClick=${() => sendCmd("/plan")}>
          📋 Plan
        </button>
      </div>
    <//>

    <!-- ── Agent Control ── -->
    <${Card} title="Agent Control">
      <textarea
        class="input mb-sm"
        rows="2"
        placeholder="Ask the agent…"
        value=${askInput}
        onInput=${(e) => setAskInput(e.target.value)}
      ></textarea>
      <div class="btn-row mb-md">
        <button
          class="btn btn-primary btn-sm"
          onClick=${() => {
            if (askInput.trim()) {
              sendCmd(`/ask ${askInput.trim()}`);
              setAskInput("");
            }
          }}
        >
          💬 Ask
        </button>
      </div>
      <div class="input-row">
        <input
          class="input"
          placeholder="Steer prompt (focus on…)"
          value=${steerInput}
          onInput=${(e) => setSteerInput(e.target.value)}
        />
        <button
          class="btn btn-secondary btn-sm"
          onClick=${() => {
            if (steerInput.trim()) {
              sendCmd(`/steer ${steerInput.trim()}`);
              setSteerInput("");
            }
          }}
        >
          🎯 Steer
        </button>
      </div>
    <//>

    <!-- ── Routing ── -->
    <${Card} title="Routing">
      <div class="card-subtitle">SDK</div>
      <${SegmentedControl}
        options=${[
          { value: "codex", label: "Codex" },
          { value: "copilot", label: "Copilot" },
          { value: "claude", label: "Claude" },
          { value: "auto", label: "Auto" },
        ]}
        value=${config?.sdk || "auto"}
        onChange=${(v) => updateConfig("sdk", v)}
      />
      <div class="card-subtitle mt-sm">Kanban</div>
      <${SegmentedControl}
        options=${[
          { value: "vk", label: "VK" },
          { value: "github", label: "GitHub" },
          { value: "jira", label: "Jira" },
        ]}
        value=${config?.kanbanBackend || "github"}
        onChange=${(v) => updateConfig("kanban", v)}
      />
      ${regions.length > 1 && html`
        <div class="card-subtitle mt-sm">Region</div>
        <${SegmentedControl}
          options=${regionOptions}
          value=${regions[0]}
          onChange=${(v) => updateConfig("region", v)}
        />
      `}
    <//>

    <!-- ── Quick Commands ── -->
    <${Card} title="Quick Commands">
      <div class="input-row mb-sm">
        <select
          class="input"
          style="flex:0 0 auto;width:80px"
          value=${quickCmdPrefix}
          onChange=${(e) => setQuickCmdPrefix(e.target.value)}
        >
          <option value="shell">Shell</option>
          <option value="git">Git</option>
        </select>
        <input
          class="input"
          placeholder=${quickCmdPrefix === "shell" ? "ls -la" : "status --short"}
          value=${quickCmdInput}
          onInput=${(e) => setQuickCmdInput(e.target.value)}
          onKeyDown=${(e) => {
            if (e.key === "Enter") handleQuickCmd();
          }}
          style="flex:1"
        />
        <button class="btn btn-secondary btn-sm" onClick=${handleQuickCmd}>
          ▶ Run
        </button>
      </div>
      ${quickCmdFeedback && html`
        <div class="meta-text mb-sm" style="color:var(--tg-theme-link-color,#4ea8d6)">
          ${quickCmdFeedback}
        </div>
      `}
      <div class="meta-text">
        Output appears in agent logs. ${""}
        <a
          href="#"
          style="color:var(--tg-theme-link-color,#4ea8d6);text-decoration:underline;cursor:pointer"
          onClick=${(e) => {
            e.preventDefault();
            import("../modules/router.js").then(({ navigateTo }) => navigateTo("logs"));
          }}
        >Open Logs tab →</a>
      </div>
    <//>

    <!-- Inline styles for new elements -->
    <style>
      .cmd-dropdown { position: absolute; bottom: 100%; left: 0; right: 0; background: var(--glass-bg, rgba(15,23,42,0.9)); border: 1px solid var(--glass-border, rgba(255,255,255,0.08)); border-radius: 12px; max-height: 240px; overflow-y: auto; z-index: 50; backdrop-filter: blur(12px); }
      .cmd-dropdown-item { padding: 8px 12px; cursor: pointer; display: flex; justify-content: space-between; align-items: center; }
      .cmd-dropdown-item.selected { background: rgba(99,102,241,0.2); }
      .cmd-dropdown-item:hover { background: rgba(99,102,241,0.15); }
      .cmd-output-panel { margin-top: 0; background: rgba(0,0,0,0.4); border-radius: 0 0 8px 8px; padding: 8px 12px; font-family: monospace; font-size: 0.8rem; color: #4ade80; max-height: 200px; overflow-y: auto; white-space: pre-wrap; }
      @keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.4; } }
    </style>
  `;
}
