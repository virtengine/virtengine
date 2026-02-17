/* ─────────────────────────────────────────────────────────────
 *  VirtEngine Control Center – Preact + HTM Entry Point
 *  Modular SPA for Telegram Mini App (no build step)
 * ────────────────────────────────────────────────────────────── */

import { h, render as preactRender } from "preact";
import { useState, useEffect, useCallback } from "preact/hooks";
import { signal } from "@preact/signals";
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
  loadNotificationPrefs,
  applyStoredDefaults,
} from "./modules/state.js";
import { activeTab, navigateTo, TAB_CONFIG } from "./modules/router.js";

/* ── Component imports ── */
import { ToastContainer } from "./components/shared.js";
import { PullToRefresh } from "./components/forms.js";
import {
  CommandPalette,
  useCommandPalette,
} from "./components/command-palette.js";

/* ── Tab imports ── */
import { DashboardTab } from "./tabs/dashboard.js";
import { TasksTab } from "./tabs/tasks.js";
import { AgentsTab } from "./tabs/agents.js";
import { InfraTab } from "./tabs/infra.js";
import { ControlTab } from "./tabs/control.js";
import { LogsTab } from "./tabs/logs.js";
import { SettingsTab } from "./tabs/settings.js";

/* ── Placeholder signals for connection quality (may be provided by api.js) ── */
let wsLatency = signal(null);
let wsReconnectIn = signal(null);
let dataFreshness = signal(null);
try {
  const apiMod = await import("./modules/api.js");
  if (apiMod.wsLatency) wsLatency = apiMod.wsLatency;
  if (apiMod.wsReconnectIn) wsReconnectIn = apiMod.wsReconnectIn;
} catch { /* use placeholder signals */ }
try {
  const stateMod = await import("./modules/state.js");
  if (stateMod.dataFreshness) dataFreshness = stateMod.dataFreshness;
} catch { /* use placeholder signals */ }

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
  const latency = wsLatency.value;
  const reconnect = wsReconnectIn.value;
  const freshness = dataFreshness.value;

  // Connection quality label
  let connLabel = "Offline";
  let connClass = "disconnected";
  if (isConn && latency != null) {
    connLabel = `${latency}ms`;
    connClass = "connected";
  } else if (isConn) {
    connLabel = "Live";
    connClass = "connected";
  } else if (reconnect != null && reconnect > 0) {
    connLabel = `Reconnecting in ${reconnect}s…`;
    connClass = "reconnecting";
  }

  // Freshness label
  let freshnessLabel = "";
  if (freshness != null) {
    const ago = Math.round((Date.now() - freshness) / 1000);
    if (ago < 5) freshnessLabel = "Updated just now";
    else if (ago < 60) freshnessLabel = `Updated ${ago}s ago`;
    else freshnessLabel = `Updated ${Math.round(ago / 60)}m ago`;
  }

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
        <div class="connection-pill ${connClass}">
          <span class="connection-dot"></span>
          ${connLabel}
        </div>
        ${freshnessLabel
          ? html`<div class="header-freshness" style="font-size:11px;opacity:0.55;margin-top:2px">${freshnessLabel}</div>`
          : null}
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
  const { open: paletteOpen, onClose: paletteClose } = useCommandPalette();

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

    // Load notification preferences early (non-blocking)
    loadNotificationPrefs();

    // Load initial data for the default tab, then apply stored executor defaults
    refreshTab("dashboard").then(() => applyStoredDefaults());

    // Global keyboard shortcuts (1-7 for tabs, Escape for modals)
    function handleGlobalKeys(e) {
      const tag = (document.activeElement?.tagName || "").toLowerCase();
      if (tag === "input" || tag === "textarea" || tag === "select") return;
      if (document.activeElement?.isContentEditable) return;

      // Number keys 1-7 to switch tabs
      const num = parseInt(e.key, 10);
      if (num >= 1 && num <= 7 && !e.metaKey && !e.ctrlKey && !e.altKey) {
        const tabCfg = TAB_CONFIG[num - 1];
        if (tabCfg) {
          e.preventDefault();
          navigateTo(tabCfg.id);
        }
        return;
      }

      // Escape to close modals/palette
      if (e.key === "Escape") {
        globalThis.dispatchEvent(new CustomEvent("ve:close-modals"));
      }
    }
    document.addEventListener("keydown", handleGlobalKeys);

    return () => {
      unsub();
      disconnectWebSocket();
      document.removeEventListener("keydown", handleGlobalKeys);
    };
  }, []);

  const CurrentTab = TAB_COMPONENTS[activeTab.value] || DashboardTab;

  return html`
    <${Header} />
    <${ToastContainer} />
    <${CommandPalette} open=${paletteOpen} onClose=${paletteClose} />
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
