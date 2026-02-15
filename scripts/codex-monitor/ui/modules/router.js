/* ─────────────────────────────────────────────────────────────
 *  VirtEngine Control Center – Router / Tab Navigation
 *  Manages active tab, history stack, and Telegram BackButton
 * ────────────────────────────────────────────────────────────── */

import { signal } from "https://esm.sh/@preact/signals@1.3.1";
import { haptic, showBackButton, hideBackButton } from "./telegram.js";
import { refreshTab } from "./state.js";

/** Currently active tab ID */
export const activeTab = signal("dashboard");

/** Navigation history stack (for back button) */
const tabHistory = [];

/**
 * Navigate to a new tab. Pushes current tab onto the history stack
 * and refreshes data for the target tab.
 * @param {string} tab
 */
export function navigateTo(tab) {
  if (tab === activeTab.value) return;
  haptic("light");
  tabHistory.push(activeTab.value);
  activeTab.value = tab;
  refreshTab(tab);

  // Show Telegram BackButton when there is history
  if (tabHistory.length > 0) {
    showBackButton(goBack);
  }
}

/**
 * Go back to the previous tab (from history stack).
 */
export function goBack() {
  const prev = tabHistory.pop();
  if (prev) {
    haptic("light");
    activeTab.value = prev;
    refreshTab(prev);
  }
  if (tabHistory.length === 0) {
    hideBackButton();
  }
}

/**
 * Ordered list of tabs with metadata for rendering the navigation UI.
 * The `icon` key maps to a property on the ICONS object in modules/icons.js.
 */
export const TAB_CONFIG = [
  { id: "dashboard", label: "Home", icon: "grid" },
  { id: "tasks", label: "Tasks", icon: "check" },
  { id: "agents", label: "Agents", icon: "cpu" },
  { id: "infra", label: "Infra", icon: "server" },
  { id: "control", label: "Control", icon: "sliders" },
  { id: "logs", label: "Logs", icon: "terminal" },
  { id: "settings", label: "Settings", icon: "settings" },
];
