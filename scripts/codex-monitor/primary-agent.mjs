/**
 * primary-agent.mjs — Adapter that selects the primary agent implementation.
 *
 * Chooses between Codex SDK and Copilot SDK based on executor configuration.
 */

import { loadConfig } from "./config.mjs";
import {
  execCodexPrompt,
  isCodexBusy,
  getThreadInfo,
  resetThread,
  initCodexShell,
  steerCodexPrompt,
} from "./codex-shell.mjs";
import {
  execCopilotPrompt,
  isCopilotBusy,
  getSessionInfo,
  resetSession,
  initCopilotShell,
  steerCopilotPrompt,
} from "./copilot-shell.mjs";

let primaryKind = null;
let primaryProfile = null;
let primaryFallbackReason = null;
let initialized = false;

function selectPrimaryExecutor(config) {
  const executors = config?.executorConfig?.executors || [];
  if (!executors.length) return null;
  const primary = executors.find(
    (e) => (e.role || "").toLowerCase() === "primary",
  );
  return primary || executors[0];
}

export async function initPrimaryAgent(config = null) {
  if (initialized) return primaryKind;
  const cfg = config || loadConfig();
  primaryProfile = selectPrimaryExecutor(cfg);
  const executor = primaryProfile?.executor
    ? String(primaryProfile.executor).toUpperCase()
    : "CODEX";
  primaryKind = executor === "COPILOT" ? "COPILOT" : "CODEX";

  if (primaryKind === "CODEX" && process.env.CODEX_SDK_DISABLED === "1") {
    primaryFallbackReason = "Codex SDK disabled — attempting Copilot";
    const ok = await initCopilotShell();
    if (ok) {
      primaryKind = "COPILOT";
      initialized = true;
      return primaryKind;
    }
    primaryFallbackReason = "Codex SDK disabled — Copilot unavailable";
  }

  if (primaryKind === "COPILOT") {
    const ok = await initCopilotShell();
    if (!ok) {
      primaryFallbackReason = "Copilot SDK unavailable — falling back to Codex";
      primaryKind = "CODEX";
      await initCodexShell();
    }
  } else {
    await initCodexShell();
  }

  initialized = true;
  return primaryKind;
}

export function getPrimaryAgentInfo() {
  const info =
    primaryKind === "COPILOT" ? getSessionInfo() : getThreadInfo();
  return {
    provider: primaryKind || "CODEX",
    profile: primaryProfile,
    fallbackReason: primaryFallbackReason,
    ...info,
  };
}

export function isPrimaryBusy() {
  if (primaryKind === "COPILOT") return isCopilotBusy();
  return isCodexBusy();
}

export async function execPrimaryPrompt(userMessage, options = {}) {
  if (!initialized) {
    await initPrimaryAgent();
  }
  if (primaryKind === "COPILOT") {
    return execCopilotPrompt(userMessage, options);
  }
  return execCodexPrompt(userMessage, options);
}

export async function resetPrimaryAgent() {
  if (!initialized) {
    await initPrimaryAgent();
  }
  if (primaryKind === "COPILOT") {
    return resetSession();
  }
  return resetThread();
}

export async function steerPrimaryPrompt(message) {
  if (!initialized) {
    await initPrimaryAgent();
  }
  if (primaryKind === "COPILOT") {
    return steerCopilotPrompt(message);
  }
  return steerCodexPrompt(message);
}
