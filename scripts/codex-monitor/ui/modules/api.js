/* ─────────────────────────────────────────────────────────────
 *  VirtEngine Control Center – API Client & WebSocket
 *  Handles REST calls, WS connection, and command sending
 * ────────────────────────────────────────────────────────────── */

import { signal } from "https://esm.sh/@preact/signals@1.3.1";
import { getInitData } from "./telegram.js";

/** Reactive signal: whether the WebSocket is currently connected */
export const wsConnected = signal(false);

/* ─── REST API Client ─── */

/**
 * Fetch from the API (same-origin). Automatically injects the
 * X-Telegram-InitData header and handles JSON parsing / errors.
 *
 * @param {string} path  - API path, e.g. "/api/status"
 * @param {RequestInit & {_silent?: boolean}} options
 * @returns {Promise<any>} parsed JSON body
 */
export async function apiFetch(path, options = {}) {
  const headers = { ...options.headers };
  headers["Content-Type"] = headers["Content-Type"] || "application/json";

  const initData = getInitData();
  if (initData) {
    headers["X-Telegram-InitData"] = initData;
  }

  const silent = options._silent;
  delete options._silent;

  try {
    const res = await fetch(path, { ...options, headers });
    if (!res.ok) {
      const text = await res.text().catch(() => "");
      throw new Error(text || `Request failed (${res.status})`);
    }
    return await res.json();
  } catch (err) {
    // Re-throw so callers can catch, but don't toast on silent requests
    if (!silent) {
      // Dispatch a custom event so the state layer can show a toast
      try {
        globalThis.dispatchEvent(
          new CustomEvent("ve:api-error", { detail: { message: err.message } }),
        );
      } catch {
        /* noop */
      }
    }
    throw err;
  }
}

/* ─── Command Sending ─── */

/**
 * Send a slash-command to the backend via POST /api/command.
 * @param {string} cmd  - e.g. "/status" or "/starttask abc123"
 * @returns {Promise<any>}
 */
export async function sendCommandToChat(cmd) {
  return apiFetch("/api/command", {
    method: "POST",
    body: JSON.stringify({ command: cmd }),
  });
}

/* ─── WebSocket ─── */

/** @type {WebSocket|null} */
let ws = null;
/** @type {ReturnType<typeof setTimeout>|null} */
let reconnectTimer = null;
let retryMs = 1000;

/** Registered message handlers */
const wsHandlers = new Set();

/**
 * Register a handler for incoming WS messages.
 * Returns an unsubscribe function.
 * @param {(data: any) => void} handler
 * @returns {() => void}
 */
export function onWsMessage(handler) {
  wsHandlers.add(handler);
  return () => wsHandlers.delete(handler);
}

/**
 * Open (or re-open) a WebSocket connection to /ws.
 * Automatically reconnects on close with exponential backoff.
 */
export function connectWebSocket() {
  // Prevent double connections
  if (
    ws &&
    (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)
  ) {
    return;
  }

  const proto = globalThis.location.protocol === "https:" ? "wss" : "ws";
  const wsUrl = new URL(`${proto}://${globalThis.location.host}/ws`);

  const initData = getInitData();
  if (initData) wsUrl.searchParams.set("initData", initData);

  const socket = new WebSocket(wsUrl.toString());
  ws = socket;

  socket.addEventListener("open", () => {
    wsConnected.value = true;
    retryMs = 1000; // reset backoff on successful connect
  });

  socket.addEventListener("message", (event) => {
    let msg;
    try {
      msg = JSON.parse(event.data || "{}");
    } catch {
      return;
    }
    // Dispatch to all registered handlers
    for (const handler of wsHandlers) {
      try {
        handler(msg);
      } catch {
        /* handler errors shouldn't crash the WS loop */
      }
    }
  });

  socket.addEventListener("close", () => {
    wsConnected.value = false;
    ws = null;
    // Auto-reconnect with exponential backoff (max 15 s)
    if (reconnectTimer) clearTimeout(reconnectTimer);
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null;
      connectWebSocket();
    }, retryMs);
    retryMs = Math.min(15000, retryMs * 2);
  });

  socket.addEventListener("error", () => {
    wsConnected.value = false;
  });
}

/**
 * Disconnect the WebSocket and cancel any pending reconnect.
 */
export function disconnectWebSocket() {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  if (ws) {
    try {
      ws.close();
    } catch {
      /* noop */
    }
    ws = null;
  }
  wsConnected.value = false;
}

/**
 * Send a raw JSON message over the open WebSocket.
 * @param {any} data
 */
export function wsSend(data) {
  if (ws?.readyState === WebSocket.OPEN) {
    ws.send(typeof data === "string" ? data : JSON.stringify(data));
  }
}
