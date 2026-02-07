/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * WebSocket client for CometBFT real-time chain event subscriptions.
 * Handles connection lifecycle, reconnection with exponential backoff,
 * and event parsing from the CometBFT JSON-RPC WebSocket API.
 */

import type {
  ChainEvent,
  ChainEventConfig,
  ChainEventType,
  ConnectionStatus,
} from '@/types/chain-events';
import { DEFAULT_CHAIN_EVENT_CONFIG } from '@/types/chain-events';

type EventHandler = (event: ChainEvent) => void;
type StatusHandler = (status: ConnectionStatus) => void;

/** Maps chain event types to CometBFT event query strings. */
const EVENT_QUERIES: Record<ChainEventType, string> = {
  'order.created': "message.action='CreateOrder'",
  'bid.created': "message.action='CreateBid'",
  'allocation.status_changed': "message.action='UpdateAllocationStatus'",
  'settlement.executed': "message.action='ExecuteSettlement'",
  'hpc_job.status_changed': "message.action='UpdateHPCJobStatus'",
};

/** Counter for JSON-RPC request IDs. */
let rpcIdCounter = 0;

/**
 * ChainEventClient manages a WebSocket connection to a CometBFT node,
 * subscribes to relevant transaction events, and emits parsed ChainEvent
 * objects to registered handlers.
 */
export class ChainEventClient {
  private ws: WebSocket | null = null;
  private config: ChainEventConfig;
  private eventHandlers: EventHandler[] = [];
  private statusHandlers: StatusHandler[] = [];
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private status: ConnectionStatus = 'disconnected';
  private disposed = false;

  constructor(config: Partial<ChainEventConfig> = {}) {
    this.config = { ...DEFAULT_CHAIN_EVENT_CONFIG, ...config };
  }

  /** Register a handler for parsed chain events. Returns unsubscribe fn. */
  onEvent(handler: EventHandler): () => void {
    this.eventHandlers.push(handler);
    return () => {
      this.eventHandlers = this.eventHandlers.filter((h) => h !== handler);
    };
  }

  /** Register a handler for connection status changes. Returns unsubscribe fn. */
  onStatusChange(handler: StatusHandler): () => void {
    this.statusHandlers.push(handler);
    return () => {
      this.statusHandlers = this.statusHandlers.filter((h) => h !== handler);
    };
  }

  /** Current connection status. */
  getStatus(): ConnectionStatus {
    return this.status;
  }

  /** Open the WebSocket connection and subscribe to events. */
  connect(): void {
    if (this.disposed) return;
    if (this.ws?.readyState === WebSocket.OPEN || this.ws?.readyState === WebSocket.CONNECTING) {
      return;
    }

    this.setStatus('connecting');

    try {
      this.ws = new WebSocket(this.config.wsUrl);
    } catch {
      this.setStatus('disconnected');
      this.scheduleReconnect();
      return;
    }

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
      this.setStatus('connected');
      this.subscribeAll();
    };

    this.ws.onmessage = (msg) => {
      this.handleMessage(msg.data);
    };

    this.ws.onclose = () => {
      this.setStatus('disconnected');
      if (!this.disposed) {
        this.scheduleReconnect();
      }
    };

    this.ws.onerror = () => {
      // onclose will fire after onerror — reconnect handled there.
    };
  }

  /** Close the connection and clean up. */
  disconnect(): void {
    this.disposed = true;
    this.clearReconnectTimer();
    if (this.ws) {
      this.ws.onclose = null;
      this.ws.onerror = null;
      this.ws.onmessage = null;
      this.ws.onopen = null;
      this.ws.close();
      this.ws = null;
    }
    this.setStatus('disconnected');
  }

  // ---------------------------------------------------------------------------
  // Internal helpers
  // ---------------------------------------------------------------------------

  private setStatus(status: ConnectionStatus): void {
    if (this.status === status) return;
    this.status = status;
    for (const handler of this.statusHandlers) {
      handler(status);
    }
  }

  private subscribeAll(): void {
    const topics =
      this.config.subscriptions.length > 0
        ? this.config.subscriptions
        : (Object.keys(EVENT_QUERIES) as ChainEventType[]);

    for (const topic of topics) {
      const query = EVENT_QUERIES[topic];
      if (query && this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(
          JSON.stringify({
            jsonrpc: '2.0',
            method: 'subscribe',
            params: { query: `tm.event='Tx' AND ${query}` },
            id: ++rpcIdCounter,
          })
        );
      }
    }
  }

  private handleMessage(raw: string): void {
    try {
      const data = JSON.parse(raw);
      // Subscription confirmations have result: {} — skip them.
      if (!data.result?.data?.value?.TxResult) return;

      const txResult = data.result.data.value.TxResult;
      const txHash: string = data.result.events?.['tx.hash']?.[0] ?? '';
      const height: number = parseInt(txResult.height ?? '0', 10);

      const events = txResult.result?.events ?? [];
      for (const evt of events) {
        const parsed = this.parseEvent(evt, txHash, height);
        if (parsed) {
          for (const handler of this.eventHandlers) {
            handler(parsed);
          }
        }
      }
    } catch {
      // Malformed messages are silently ignored.
    }
  }

  private parseEvent(
    evt: { type?: string; attributes?: Array<{ key: string; value: string }> },
    txHash: string,
    blockHeight: number
  ): ChainEvent | null {
    const rawType = evt.type ?? '';
    const eventType = this.matchEventType(rawType);
    if (!eventType) return null;

    const attrs: Record<string, string> = {};
    for (const attr of evt.attributes ?? []) {
      attrs[attr.key] = attr.value;
    }

    const idx = Object.keys(attrs).length;
    return {
      id: `${txHash}-${rawType}-${idx}`,
      type: eventType,
      blockHeight,
      timestamp: new Date(),
      txHash,
      attributes: attrs,
    };
  }

  private matchEventType(rawType: string): ChainEventType | null {
    const typeMap: Record<string, ChainEventType> = {
      create_order: 'order.created',
      create_bid: 'bid.created',
      update_allocation_status: 'allocation.status_changed',
      execute_settlement: 'settlement.executed',
      update_hpc_job_status: 'hpc_job.status_changed',
    };
    return typeMap[rawType] ?? null;
  }

  private scheduleReconnect(): void {
    if (!this.config.autoReconnect) return;
    if (this.reconnectAttempts >= this.config.maxReconnectAttempts) return;

    this.clearReconnectTimer();
    this.setStatus('reconnecting');

    const delay = Math.min(
      this.config.reconnectDelayMs * Math.pow(2, this.reconnectAttempts),
      this.config.maxReconnectDelayMs
    );
    this.reconnectAttempts++;

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, delay);
  }

  private clearReconnectTimer(): void {
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }
}
