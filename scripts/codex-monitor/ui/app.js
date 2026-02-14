const state = {
  tab: "overview",
  bootstrap: null,
  tasks: null,
  agents: null,
  worktrees: null,
  logs: null,
  commandInput: "/status",
  maxParallel: "",
  connected: false,
};

const view = document.getElementById("view");
const connectionPill = document.getElementById("connection-pill");
const tabs = Array.from(document.querySelectorAll(".tab"));

function tg() {
  return window.Telegram?.WebApp || null;
}

function setConnection(connected, text = "") {
  state.connected = connected;
  connectionPill.textContent = connected
    ? `Connected ${text}`.trim()
    : `Offline ${text}`.trim();
}

async function apiFetch(path, options = {}) {
  const headers = {
    "Content-Type": "application/json",
    ...(options.headers || {}),
  };
  const initData = tg()?.initData;
  if (initData) headers["X-Telegram-InitData"] = initData;

  const response = await fetch(path, { ...options, headers });
  if (!response.ok) {
    const text = await response.text().catch(() => "");
    throw new Error(text || `HTTP ${response.status}`);
  }
  return response.json();
}

async function runCommand(command) {
  const res = await apiFetch("/api/command", {
    method: "POST",
    body: JSON.stringify({ command }),
  });
  if (!res.ok) throw new Error(res.error || "command failed");
  return res;
}

function renderOverview() {
  const status = state.bootstrap?.status || {};
  const counts = status.counts || {};
  const executor = state.bootstrap?.executor || {};

  return `
    <section class="card">
      <h2>System Overview</h2>
      <div class="grid columns-2">
        <div class="stat"><strong>${counts.running ?? 0}</strong>Running</div>
        <div class="stat"><strong>${counts.review ?? 0}</strong>In Review</div>
        <div class="stat"><strong>${counts.error ?? 0}</strong>Blocked</div>
        <div class="stat"><strong>${status.backlog_remaining ?? "?"}</strong>Backlog</div>
      </div>
    </section>
    <section class="card">
      <h3>Executor Snapshot</h3>
      <p class="meta">Mode: ${state.bootstrap?.executorMode || "vk"}</p>
      <p class="meta">Slots: ${executor.activeSlots ?? 0}/${executor.maxParallel ?? 0}</p>
      <div class="button-row">
        <button class="action warn" data-action="executor:pause">Pause</button>
        <button class="action secondary" data-action="executor:resume">Resume</button>
        <button class="action muted" data-action="command:/status">Send /status</button>
      </div>
    </section>
  `;
}

function renderTasks() {
  const tasks = state.tasks?.tasks || [];
  const items = tasks
    .map(
      (task) => `
      <div class="task-card">
        <header>
          <div>
            <div>${task.task_title || task.task_id || "Task"}</div>
            <div class="meta">${task.task_id || ""}</div>
          </div>
          <span class="badge">${task.status || "unknown"}</span>
        </header>
        <div class="button-row">
          <button class="action muted" data-action="command:/retry manual_ui">Retry</button>
          <button class="action muted" data-action="command:/tasks">Refresh Tasks</button>
        </div>
      </div>
    `,
    )
    .join("");

  return `
    <section class="card">
      <h2>Live Tasks</h2>
      <p class="meta">Derived from executor/status snapshot.</p>
      <div class="grid">${items || "<p class=\"meta\">No active tasks.</p>"}</div>
    </section>
  `;
}

function renderAgents() {
  const agents = state.agents?.agents || [];
  const rows = agents
    .map(
      (agent) => `
      <div class="task-card">
        <header>
          <div>
            <div>${agent.taskTitle || "(idle)"}</div>
            <div class="meta">Agent ${agent.agentInstanceId || "n/a"} · ${agent.sdk || "?"}</div>
          </div>
          <span class="badge">${agent.status || "unknown"}</span>
        </header>
        <div class="button-row">
          <button class="action muted" data-action="command:/agentlogs ${agent.branch || agent.taskId || ""}">Logs</button>
          <button class="action muted" data-action="command:/steer focus on ${agent.taskTitle || "current task"}">Steer</button>
        </div>
      </div>
    `,
    )
    .join("");

  return `
    <section class="card">
      <h2>Agent Fleet</h2>
      <div class="grid">${rows || "<p class=\"meta\">No active agents.</p>"}</div>
      <div class="button-row" style="margin-top:10px">
        <button class="action muted" data-action="command:/agents">Send /agents</button>
        <button class="action muted" data-action="command:/threads">Send /threads</button>
      </div>
    </section>
  `;
}

function renderWorktrees() {
  const worktrees = state.worktrees?.worktrees || [];
  const cards = worktrees
    .map(
      (wt) => `
      <div class="task-card">
        <header>
          <div>
            <div>${wt.branch || "(detached)"}</div>
            <div class="meta">${wt.path || ""}</div>
          </div>
          <span class="badge">${wt.status || "active"}</span>
        </header>
        <div class="button-row">
          <button class="action muted" data-action="command:/worktrees">Inspect</button>
        </div>
      </div>
    `,
    )
    .join("");

  return `
    <section class="card">
      <h2>Worktrees</h2>
      <div class="grid">${cards || "<p class=\"meta\">No worktrees reported.</p>"}</div>
    </section>
  `;
}

function renderLogs() {
  const log = state.logs?.tail || [];
  return `
    <section class="card">
      <h2>Monitor Logs</h2>
      <div class="input-row">
        <button class="action muted" data-action="logs:refresh">Reload</button>
        <button class="action muted" data-action="command:/logs 100">Send /logs 100</button>
      </div>
      <div class="log-box" style="margin-top:10px">${(log.length ? log.join("\n") : "No logs available.").replace(/</g, "&lt;")}</div>
    </section>
  `;
}

function renderExecutor() {
  const executor = state.bootstrap?.executor || {};
  return `
    <section class="card">
      <h2>Executor Controls</h2>
      <p class="meta">Current max parallel: ${executor.maxParallel ?? "?"}</p>
      <div class="input-row">
        <input id="max-parallel-input" type="number" min="0" max="20" value="${state.maxParallel || executor.maxParallel || 0}" />
        <button class="action" data-action="executor:setmax">Apply</button>
      </div>
      <div class="button-row" style="margin-top:10px">
        <button class="action warn" data-action="executor:pause">Pause</button>
        <button class="action secondary" data-action="executor:resume">Resume</button>
      </div>
    </section>
  `;
}

function renderCommands() {
  return `
    <section class="card">
      <h2>Command Console</h2>
      <p class="meta">Run any Telegram slash command through the bot process.</p>
      <div class="input-row">
        <input id="command-input" type="text" value="${state.commandInput}" placeholder="/status" />
        <button class="action" data-action="command:run">Run</button>
      </div>
      <div class="button-row" style="margin-top:10px">
        <button class="action muted" data-action="command:/status">/status</button>
        <button class="action muted" data-action="command:/tasks">/tasks</button>
        <button class="action muted" data-action="command:/agents">/agents</button>
        <button class="action muted" data-action="command:/helpfull">/helpfull</button>
      </div>
    </section>
  `;
}

function render() {
  const map = {
    overview: renderOverview,
    tasks: renderTasks,
    agents: renderAgents,
    worktrees: renderWorktrees,
    logs: renderLogs,
    executor: renderExecutor,
    commands: renderCommands,
  };
  const renderFn = map[state.tab] || renderOverview;
  view.innerHTML = renderFn();
}

async function refreshCurrentTab() {
  if (state.tab === "overview" || state.tab === "executor") {
    state.bootstrap = await apiFetch("/api/bootstrap");
  }
  if (state.tab === "tasks") {
    state.tasks = await apiFetch("/api/tasks");
  }
  if (state.tab === "agents") {
    state.agents = await apiFetch("/api/agents");
  }
  if (state.tab === "worktrees") {
    state.worktrees = await apiFetch("/api/worktrees");
  }
  if (state.tab === "logs") {
    state.logs = await apiFetch("/api/logs?lines=200");
  }
  render();
}

async function refreshAll() {
  const [bootstrap, tasks, agents, worktrees, logs] = await Promise.all([
    apiFetch("/api/bootstrap").catch(() => null),
    apiFetch("/api/tasks").catch(() => null),
    apiFetch("/api/agents").catch(() => null),
    apiFetch("/api/worktrees").catch(() => null),
    apiFetch("/api/logs?lines=200").catch(() => null),
  ]);
  state.bootstrap = bootstrap;
  state.tasks = tasks;
  state.agents = agents;
  state.worktrees = worktrees;
  state.logs = logs;
  render();
}

async function handleAction(action, el) {
  if (!action) return;
  if (action.startsWith("tab:")) {
    state.tab = action.slice(4);
    tabs.forEach((tab) =>
      tab.classList.toggle("active", tab.dataset.action === action),
    );
    await refreshCurrentTab();
    return;
  }

  if (action === "refresh" || action === "logs:refresh") {
    await refreshCurrentTab();
    return;
  }

  if (action === "command:run") {
    const input = document.getElementById("command-input");
    const command = String(input?.value || "").trim();
    if (!command.startsWith("/")) {
      alert("Command must start with /");
      return;
    }
    state.commandInput = command;
    await runCommand(command);
    return;
  }

  if (action.startsWith("command:")) {
    const command = action.slice(8);
    await runCommand(command);
    return;
  }

  if (action === "executor:pause") {
    await apiFetch("/api/executor/pause", { method: "POST" });
    await refreshCurrentTab();
    return;
  }

  if (action === "executor:resume") {
    await apiFetch("/api/executor/resume", { method: "POST" });
    await refreshCurrentTab();
    return;
  }

  if (action === "executor:setmax") {
    const input = document.getElementById("max-parallel-input");
    const value = Number(input?.value || "0");
    if (!Number.isFinite(value) || value < 0 || value > 20) {
      alert("Value must be between 0 and 20");
      return;
    }
    state.maxParallel = String(value);
    await apiFetch("/api/executor/maxparallel", {
      method: "POST",
      body: JSON.stringify({ value }),
    });
    await refreshCurrentTab();
  }
}

document.addEventListener("click", async (event) => {
  const target = event.target.closest("[data-action]");
  if (!target) return;
  const action = target.dataset.action;
  try {
    await handleAction(action, target);
    setConnection(true);
  } catch (err) {
    console.error(err);
    setConnection(false, "(action failed)");
    alert(err.message || String(err));
  }
});

async function boot() {
  const webApp = tg();
  if (webApp) {
    webApp.ready();
    webApp.expand();
  }

  try {
    await refreshAll();
    setConnection(true);
  } catch (err) {
    console.error(err);
    setConnection(false, "(init failed)");
    view.innerHTML = `<section class="card"><h2>Unable to load data</h2><p>${err.message}</p></section>`;
  }

  setInterval(async () => {
    try {
      await refreshCurrentTab();
      setConnection(true);
    } catch {
      setConnection(false, "(polling)");
    }
  }, 5000);
}

boot();
