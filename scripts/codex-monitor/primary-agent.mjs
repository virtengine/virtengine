/**
 * primary-agent.mjs - Unified adapter for Codex or Claude primary agents.
 */

import {
  execCodexPrompt,
  steerCodexPrompt,
  isCodexBusy,
  getThreadInfo,
  resetThread,
  initCodexShell,
} from "./codex-shell.mjs";
import {
  execClaudePrompt,
  steerClaudePrompt,
  isClaudeBusy,
  getSessionInfo,
  resetClaudeSession,
  initClaudeShell,
} from "./claude-shell.mjs";

const ADAPTERS = {
  "codex-sdk": {
    name: "codex-sdk",
    exec: execCodexPrompt,
    steer: steerCodexPrompt,
    isBusy: isCodexBusy,
    getInfo: () => {
      const info = getThreadInfo();
      return { ...info, sessionId: info.threadId };
    },
    reset: resetThread,
    init: initCodexShell,
  },
  "claude-sdk": {
    name: "claude-sdk",
    exec: execClaudePrompt,
    steer: steerClaudePrompt,
    isBusy: isClaudeBusy,
    getInfo: () => getSessionInfo(),
    reset: resetClaudeSession,
    init: initClaudeShell,
  },
};

let activeAdapter = ADAPTERS["codex-sdk"];

function normalizePrimaryAgent(value) {
  const raw = String(value || "").trim().toLowerCase();
  if (!raw) return "codex-sdk";
  if (["codex", "codex-sdk"].includes(raw)) return "codex-sdk";
  if (["claude", "claude-sdk", "claude_code", "claude-code"].includes(raw))
    return "claude-sdk";
  return raw;
}

export function setPrimaryAgent(name) {
  const normalized = normalizePrimaryAgent(name);
  activeAdapter = ADAPTERS[normalized] || ADAPTERS["codex-sdk"];
  return activeAdapter.name;
}

export function getPrimaryAgentName() {
  return activeAdapter?.name || "codex-sdk";
}

export async function initPrimaryAgent(name) {
  if (name) {
    setPrimaryAgent(name);
  } else if (process.env.PRIMARY_AGENT) {
    setPrimaryAgent(process.env.PRIMARY_AGENT);
  }
  if (activeAdapter?.init) {
    await activeAdapter.init();
  }
}

export async function execPrimaryPrompt(...args) {
  return activeAdapter.exec(...args);
}

export async function steerPrimaryPrompt(...args) {
  return activeAdapter.steer(...args);
}

export function isPrimaryBusy() {
  return activeAdapter.isBusy();
}

export function getPrimaryAgentInfo() {
  const info = activeAdapter.getInfo ? activeAdapter.getInfo() : {};
  return {
    adapter: activeAdapter.name,
    sessionId: info.sessionId || info.threadId || null,
    threadId: info.threadId || null,
    turnCount: info.turnCount || 0,
    isActive: !!info.isActive,
    isBusy: !!info.isBusy,
  };
}

export async function resetPrimaryAgent() {
  if (activeAdapter.reset) {
    await activeAdapter.reset();
  }
}
