/* ─────────────────────────────────────────────────────────────
 *  VirtEngine Control Center – Preact + HTM Entry Point
 *  Modular SPA for Telegram Mini App (no build step)
 * ────────────────────────────────────────────────────────────── */

import { h, render as preactRender } from "preact";
import { useState, useEffect } from "preact/hooks";
import htm from "htm";

const html = htm.bind(h);

/* ── Module imports ── */
import { ICONS } from "./modules/icons.js";
import {
  initTelegramApp,
  onThemeChange,
  getTg,
  isTelegramContext,
  showSettingsButton,
  getTelegramUser,
  colorScheme,
} from "./modules/telegram.js";
import {
  connectWebSocket,
  disconnectWebSocket,
  wsConnected,
} from "./modules/api.js";
import {
  connected,
  refreshTab,
  toasts,
  initWsInvalidationListener,
} from "./modules/state.js";
import { activeTab, navigateTo, TAB_CONFIG } from "./modules/router.js";

/* ── Component imports ── */
import { ToastContainer } from "./components/shared.js";
import { PullToRefresh } from "./components/forms.js";

/* ── Tab imports ── */
import { DashboardTab } from "./tabs/dashboard.js";
import { TasksTab } from "./tabs/tasks.js";
import { AgentsTab } from "./tabs/agents.js";
import { InfraTab } from "./tabs/infra.js";
import { ControlTab } from "./tabs/control.js";
import { LogsTab } from "./tabs/logs.js";
import { SettingsTab } from "./tabs/settings.js";

/* ── Tab component map ── */
const TAB_COMPONENTS = {
  dashboard: DashboardTab,
  tasks: TasksTab,
  agents: AgentsTab,
  infra: InfraTab,
  control: ControlTab,
  logs: LogsTab,
  settings: SettingsTab,
};

/* ═══════════════════════════════════════════════
 *  Header
 * ═══════════════════════════════════════════════ */
function Header() {
  const isConn = connected.value;
  const wsConn = wsConnected.value;
  const user = getTelegramUser();

  return html`
    <header class="app-header">
      <div class="app-header-left">
        <div class="app-header-logo">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
        </div>
        <div>
          <div class="app-header-title">VirtEngine</div>
          ${user
            ? html`<div class="app-header-user">${user.first_name}</div>`
            : null}
        </div>
      </div>
      <div class="header-actions">
        <div class="connection-pill ${isConn ? "connected" : "disconnected"}">
          <span class="connection-dot"></span>
          ${isConn ? "Live" : "Offline"}
        </div>
      </div>
    </header>
  `;
}

/* ═══════════════════════════════════════════════
 *  Bottom Navigation
 * ═══════════════════════════════════════════════ */
function BottomNav() {
  return html`
    <nav class="bottom-nav">
      ${TAB_CONFIG.filter((t) => t.id !== "settings").map(
        (tab) => html`
          <button
            key=${tab.id}
            class="nav-item ${activeTab.value === tab.id ? "active" : ""}"
            onClick=${() => navigateTo(tab.id)}
          >
            ${ICONS[tab.icon]}
            <span class="nav-label">${tab.label}</span>
          </button>
        `,
      )}
    </nav>
  `;
}

/* ═══════════════════════════════════════════════
 *  App Root
 * ═══════════════════════════════════════════════ */
function App() {
  useEffect(() => {
    // Initialize Telegram Mini App SDK
    initTelegramApp();

    // Theme change monitoring
    const unsub = onThemeChange(() => {
      colorScheme.value = getTg()?.colorScheme || "dark";
    });

    // Show settings button in Telegram header
    showSettingsButton(() => navigateTo("settings"));

    // Connect WebSocket + invalidation auto-refresh
    connectWebSocket();
    initWsInvalidationListener();

    // Load initial data for the default tab
    refreshTab("dashboard");

    return () => {
      unsub();
      disconnectWebSocket();
    };
  }, []);

  const CurrentTab = TAB_COMPONENTS[activeTab.value] || DashboardTab;

  return html`
    <${Header} />
    <${ToastContainer} />
    <${PullToRefresh} onRefresh=${() => refreshTab(activeTab.value)}>
      <main class="main-content">
        <${CurrentTab} />
      </main>
    <//>
    <${BottomNav} />
  `;
}

/* ─── Mount ─── */
preactRender(html`<${App} />`, document.getElementById("app"));
