/* ─────────────────────────────────────────────────────────────
 *  Tab: Logs — system logs, agent log library, git snapshot
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

import { haptic, showAlert } from "../modules/telegram.js";
import { apiFetch, sendCommandToChat } from "../modules/api.js";
import {
  logsData,
  logsLines,
  gitDiff,
  gitBranches,
  agentLogFiles,
  agentLogFile,
  agentLogTail,
  agentLogLines,
  agentLogQuery,
  agentContext,
  loadLogs,
  loadAgentLogFileList,
  loadAgentLogTailData,
  loadAgentContextData,
  showToast,
  scheduleRefresh,
} from "../modules/state.js";
import { ICONS } from "../modules/icons.js";
import { formatBytes } from "../modules/utils.js";
import { Card, Badge, EmptyState } from "../components/shared.js";
import { SearchInput } from "../components/forms.js";

/* ─── Log level helpers ─── */
const LOG_LEVELS = [
  { value: "all", label: "All" },
  { value: "info", label: "Info" },
  { value: "warn", label: "Warn" },
  { value: "error", label: "Error" },
];

function filterByLevel(text, level) {
  if (!text || level === "all") return text;
  return text
    .split("\n")
    .filter((line) => {
      const lower = line.toLowerCase();
      if (level === "error")
        return (
          lower.includes("error") ||
          lower.includes("err") ||
          lower.includes("fatal")
        );
      if (level === "warn")
        return (
          lower.includes("warn") ||
          lower.includes("warning") ||
          lower.includes("error") ||
          lower.includes("fatal")
        );
      return true;
    })
    .join("\n");
}

/* ─── LogsTab ─── */
export function LogsTab() {
  const logRef = useRef(null);
  const tailRef = useRef(null);

  const [localLogLines, setLocalLogLines] = useState(logsLines?.value ?? 200);
  const [localAgentLines, setLocalAgentLines] = useState(
    agentLogLines?.value ?? 200,
  );
  const [contextQuery, setContextQuery] = useState("");
  const [logLevel, setLogLevel] = useState("all");
  const [logSearch, setLogSearch] = useState("");
  const [autoScroll, setAutoScroll] = useState(true);

  /* Raw log text */
  const rawLogText = logsData?.value?.lines
    ? logsData.value.lines.join("\n")
    : "No logs yet.";

  const rawTailText = agentLogTail?.value?.lines
    ? agentLogTail.value.lines.join("\n")
    : "Select a log file.";

  /* Filtered log text */
  const filteredLogText = (() => {
    let text = filterByLevel(rawLogText, logLevel);
    if (logSearch.trim()) {
      const lines = text.split("\n");
      const matches = lines.filter((l) =>
        l.toLowerCase().includes(logSearch.toLowerCase()),
      );
      text = matches.join("\n") || "No matching lines.";
    }
    return text;
  })();

  /* Auto-scroll */
  useEffect(() => {
    if (autoScroll && logRef.current) {
      logRef.current.scrollTop = logRef.current.scrollHeight;
    }
  }, [filteredLogText, autoScroll]);

  useEffect(() => {
    if (autoScroll && tailRef.current) {
      tailRef.current.scrollTop = tailRef.current.scrollHeight;
    }
  }, [rawTailText, autoScroll]);

  /* ── System log handlers ── */
  const handleLogLinesChange = async (value) => {
    setLocalLogLines(value);
    if (logsLines) logsLines.value = value;
    await loadLogs();
  };

  /* ── Agent log handlers ── */
  const handleAgentSearch = async () => {
    if (agentLogFile) agentLogFile.value = "";
    await loadAgentLogFileList();
    await loadAgentLogTailData();
  };

  const handleAgentOpen = async (name) => {
    haptic();
    if (agentLogFile) agentLogFile.value = name;
    await loadAgentLogTailData();
  };

  const handleAgentLinesChange = async (value) => {
    setLocalAgentLines(value);
    if (agentLogLines) agentLogLines.value = value;
    await loadAgentLogTailData();
  };

  /* ── Context handler ── */
  const handleContextLoad = async () => {
    haptic();
    await loadAgentContextData(contextQuery.trim());
  };

  /* ── Git handler ── */
  const handleGitRefresh = async () => {
    haptic();
    const [branches, diff] = await Promise.all([
      apiFetch("/api/git/branches", { _silent: true }).catch(() => ({
        data: [],
      })),
      apiFetch("/api/git/diff", { _silent: true }).catch(() => ({ data: "" })),
    ]);
    if (gitBranches) gitBranches.value = branches.data || [];
    if (gitDiff) gitDiff.value = diff.data || "";
  };

  /* ── Copy to clipboard ── */
  const copyToClipboard = async (text, label) => {
    haptic();
    try {
      if (navigator.clipboard) {
        await navigator.clipboard.writeText(text);
      } else {
        const ta = document.createElement("textarea");
        ta.value = text;
        ta.style.position = "fixed";
        ta.style.left = "-9999px";
        document.body.appendChild(ta);
        ta.select();
        document.execCommand("copy");
        document.body.removeChild(ta);
      }
      showToast(`${label} copied`, "success");
    } catch {
      showToast("Copy failed", "error");
    }
  };

  return html`
    <!-- ── System Logs ── -->
    <${Card} title="System Logs">
      <div class="range-row mb-sm">
        <input
          type="range"
          min="20"
          max="800"
          step="20"
          value=${localLogLines}
          onInput=${(e) => setLocalLogLines(Number(e.target.value))}
          onChange=${(e) => handleLogLinesChange(Number(e.target.value))}
        />
        <span class="pill">${localLogLines} lines</span>
      </div>
      <div class="chip-group mb-sm">
        ${[50, 200, 500].map(
          (n) => html`
            <button
              key=${n}
              class="chip ${(logsLines?.value ?? localLogLines) === n
                ? "active"
                : ""}"
              onClick=${() => handleLogLinesChange(n)}
            >
              ${n}
            </button>
          `,
        )}
      </div>
      <div class="chip-group mb-sm">
        ${LOG_LEVELS.map(
          (l) => html`
            <button
              key=${l.value}
              class="chip chip-outline ${logLevel === l.value ? "active" : ""}"
              onClick=${() => {
                haptic();
                setLogLevel(l.value);
              }}
            >
              ${l.label}
            </button>
          `,
        )}
      </div>
      <div class="input-row mb-sm">
        <input
          class="input"
          placeholder="Search/grep logs…"
          value=${logSearch}
          onInput=${(e) => setLogSearch(e.target.value)}
        />
        <label
          class="meta-text toggle-label"
          style="white-space:nowrap"
          onClick=${() => {
            setAutoScroll(!autoScroll);
            haptic();
          }}
        >
          <input
            type="checkbox"
            checked=${autoScroll}
            style="accent-color:var(--accent)"
          />
          Auto-scroll
        </label>
      </div>
      <div ref=${logRef} class="log-box">${filteredLogText}</div>
      <div class="btn-row mt-sm">
        <button
          class="btn btn-ghost btn-sm"
          onClick=${() =>
            sendCommandToChat(`/logs ${logsLines?.value ?? localLogLines}`)}
        >
          /logs to chat
        </button>
        <button
          class="btn btn-ghost btn-sm"
          onClick=${() => copyToClipboard(filteredLogText, "Logs")}
        >
          📋 Copy
        </button>
      </div>
    <//>

    <!-- ── Agent Log Library ── -->
    <${Card} title="Agent Log Library">
      <div class="input-row mb-sm">
        <input
          class="input"
          placeholder="Search log files"
          value=${agentLogQuery?.value ?? ""}
          onInput=${(e) => {
            if (agentLogQuery) agentLogQuery.value = e.target.value;
          }}
        />
        <button class="btn btn-secondary btn-sm" onClick=${handleAgentSearch}>
          🔍 Search
        </button>
      </div>
      <div class="range-row mb-md">
        <input
          type="range"
          min="50"
          max="800"
          step="50"
          value=${localAgentLines}
          onInput=${(e) => setLocalAgentLines(Number(e.target.value))}
          onChange=${(e) => handleAgentLinesChange(Number(e.target.value))}
        />
        <span class="pill">${localAgentLines} lines</span>
      </div>
    <//>

    <!-- ── Log Files list ── -->
    <${Card} title="Log Files">
      ${(agentLogFiles?.value || []).length
        ? (agentLogFiles.value || []).map(
            (file) => html`
              <div
                key=${file.name}
                class="task-card"
                style="cursor:pointer"
                onClick=${() => handleAgentOpen(file.name)}
              >
                <div class="task-card-header">
                  <div>
                    <div class="task-card-title">${file.name}</div>
                    <div class="task-card-meta">
                      ${formatBytes
                        ? formatBytes(file.size)
                        : Math.round(file.size / 1024) + "kb"}
                      · ${new Date(file.mtime).toLocaleString()}
                    </div>
                  </div>
                  <${Badge} status="log" text="log" />
                </div>
              </div>
            `,
          )
        : html`<${EmptyState} message="No log files found." />`}
    <//>

    <!-- ── Log Tail viewer ── -->
    <${Card} title=${agentLogFile?.value || "Log Tail"}>
      ${agentLogTail?.value?.truncated &&
      html`<span class="pill mb-sm">Tail clipped</span>`}
      <div ref=${tailRef} class="log-box">${rawTailText}</div>
      <div class="btn-row mt-sm">
        <button
          class="btn btn-ghost btn-sm"
          onClick=${() => copyToClipboard(rawTailText, "Log tail")}
        >
          📋 Copy
        </button>
      </div>
    <//>

    <!-- ── Worktree Context ── -->
    <${Card} title="Worktree Context">
      <div class="input-row mb-sm">
        <input
          class="input"
          placeholder="Branch fragment"
          value=${contextQuery}
          onInput=${(e) => setContextQuery(e.target.value)}
          onKeyDown=${(e) => {
            if (e.key === "Enter") handleContextLoad();
          }}
        />
        <button class="btn btn-secondary btn-sm" onClick=${handleContextLoad}>
          📂 Load
        </button>
      </div>
      <div class="log-box">
        ${agentContext?.value
          ? [
              "Worktree: " + (agentContext.value.name || "?"),
              "",
              agentContext.value.gitLog || "No git log.",
              "",
              agentContext.value.gitStatus || "Clean worktree.",
              "",
              agentContext.value.diffStat || "No diff stat.",
            ].join("\n")
          : "Load a worktree context to view git log/status."}
      </div>
      ${agentContext?.value &&
      html`
        <div class="btn-row mt-sm">
          <button
            class="btn btn-ghost btn-sm"
            onClick=${() =>
              copyToClipboard(
                [
                  agentContext.value.gitLog,
                  agentContext.value.gitStatus,
                  agentContext.value.diffStat,
                ]
                  .filter(Boolean)
                  .join("\n\n"),
                "Context",
              )}
          >
            📋 Copy
          </button>
        </div>
      `}
    <//>

    <!-- ── Git Snapshot ── -->
    <${Card} title="Git Snapshot">
      <div class="btn-row mb-sm">
        <button class="btn btn-secondary btn-sm" onClick=${handleGitRefresh}>
          ${ICONS.refresh} Refresh
        </button>
        <button
          class="btn btn-ghost btn-sm"
          onClick=${() => sendCommandToChat("/diff")}
        >
          /diff
        </button>
        <button
          class="btn btn-ghost btn-sm"
          onClick=${() => copyToClipboard(gitDiff?.value || "", "Diff")}
        >
          📋 Copy
        </button>
      </div>
      <div class="log-box mb-md">
        ${gitDiff?.value || "Clean working tree."}
      </div>
      <div class="card-subtitle">Recent Branches</div>
      ${(gitBranches?.value || []).length
        ? (gitBranches.value || []).map(
            (line, i) => html`<div key=${i} class="meta-text">${line}</div>`,
          )
        : html`<div class="meta-text">No branches found.</div>`}
    <//>
  `;
}
