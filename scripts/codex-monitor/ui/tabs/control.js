/* ─────────────────────────────────────────────────────────────
 *  Tab: Control — executor, commands, routing, shell/git
 * ────────────────────────────────────────────────────────────── */
import { h } from "https://esm.sh/preact@10.25.4";
import { useState, useCallback } from "https://esm.sh/preact@10.25.4/hooks";
import htm from "https://esm.sh/htm@3.1.1";

const html = htm.bind(h);

import { haptic, showConfirm } from "../modules/telegram.js";
import { apiFetch, sendCommandToChat } from "../modules/api.js";
import {
  executorData,
  showToast,
  refreshTab,
  runOptimistic,
  scheduleRefresh,
} from "../modules/state.js";
import { ICONS } from "../modules/icons.js";
import { cloneValue } from "../modules/utils.js";
import { Card, Badge } from "../components/shared.js";
import { SegmentedControl, SliderControl } from "../components/forms.js";

/* ─── Command history (up to 10 recent) ─── */
const MAX_HISTORY = 10;

/* ─── ControlTab ─── */
export function ControlTab() {
  const executor = executorData.value;
  const execData = executor?.data;
  const mode = executor?.mode || "vk";

  /* Form inputs */
  const [commandInput, setCommandInput] = useState("");
  const [startTaskInput, setStartTaskInput] = useState("");
  const [retryInput, setRetryInput] = useState("");
  const [askInput, setAskInput] = useState("");
  const [steerInput, setSteerInput] = useState("");
  const [shellInput, setShellInput] = useState("");
  const [gitInput, setGitInput] = useState("");
  const [maxParallel, setMaxParallel] = useState(execData?.maxParallel ?? 0);
  const [cmdHistory, setCmdHistory] = useState([]);
  const [showHistory, setShowHistory] = useState(false);

  /* ── Command history helper ── */
  const pushHistory = useCallback((cmd) => {
    setCmdHistory((prev) => {
      const next = [cmd, ...prev.filter((c) => c !== cmd)].slice(
        0,
        MAX_HISTORY,
      );
      return next;
    });
  }, []);

  const sendCmd = useCallback(
    (cmd) => {
      if (!cmd.trim()) return;
      sendCommandToChat(cmd.trim());
      pushHistory(cmd.trim());
    },
    [pushHistory],
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

  return html`
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
            onInput=${(e) => setCommandInput(e.target.value)}
            onFocus=${() => setShowHistory(true)}
            onBlur=${() => setTimeout(() => setShowHistory(false), 200)}
            onKeyDown=${(e) => {
              if (e.key === "Enter" && commandInput.trim()) {
                sendCmd(commandInput.trim());
                setCommandInput("");
              }
            }}
          />
          <!-- Command history dropdown -->
          ${showHistory &&
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
        value=""
        onChange=${(v) => sendCmd(`/sdk ${v}`)}
      />
      <div class="card-subtitle mt-sm">Kanban</div>
      <${SegmentedControl}
        options=${[
          { value: "vk", label: "VK" },
          { value: "github", label: "GitHub" },
          { value: "jira", label: "Jira" },
        ]}
        value=""
        onChange=${(v) => sendCmd(`/kanban ${v}`)}
      />
      <div class="card-subtitle mt-sm">Region</div>
      <${SegmentedControl}
        options=${[
          { value: "us", label: "US" },
          { value: "sweden", label: "Sweden" },
          { value: "auto", label: "Auto" },
        ]}
        value=""
        onChange=${(v) => sendCmd(`/region ${v}`)}
      />
    <//>

    <!-- ── Shell / Git ── -->
    <${Card} title="Shell / Git">
      <div class="input-row mb-sm">
        <input
          class="input"
          placeholder="ls -la"
          value=${shellInput}
          onInput=${(e) => setShellInput(e.target.value)}
          onKeyDown=${(e) => {
            if (e.key === "Enter" && shellInput.trim())
              sendCmd(`/shell ${shellInput.trim()}`);
          }}
        />
        <button
          class="btn btn-secondary btn-sm"
          onClick=${() => sendCmd(`/shell ${shellInput.trim()}`.trim())}
        >
          🖥 Shell
        </button>
      </div>
      <div class="input-row">
        <input
          class="input"
          placeholder="status --short"
          value=${gitInput}
          onInput=${(e) => setGitInput(e.target.value)}
          onKeyDown=${(e) => {
            if (e.key === "Enter" && gitInput.trim())
              sendCmd(`/git ${gitInput.trim()}`);
          }}
        />
        <button
          class="btn btn-secondary btn-sm"
          onClick=${() => sendCmd(`/git ${gitInput.trim()}`.trim())}
        >
          🔀 Git
        </button>
      </div>
    <//>
  `;
}
