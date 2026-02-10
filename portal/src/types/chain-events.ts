/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Types for real-time chain event subscriptions via CometBFT WebSocket.
 */

/** Chain event topic identifiers matching on-chain event types. */
export type ChainEventType =
  | 'order.created'
  | 'bid.created'
  | 'allocation.status_changed'
  | 'settlement.executed'
  | 'hpc_job.status_changed';

/** WebSocket connection states. */
export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'reconnecting';

/** A parsed chain event received from the CometBFT WebSocket. */
export interface ChainEvent {
  /** Unique event identifier (tx hash + event index). */
  id: string;
  /** The event topic. */
  type: ChainEventType;
  /** Block height at which the event occurred. */
  blockHeight: number;
  /** Timestamp from the block header. */
  timestamp: Date;
  /** Transaction hash that emitted the event. */
  txHash: string;
  /** Key-value attributes from the event. */
  attributes: Record<string, string>;
}

/** Configuration for the chain event WebSocket client. */
export interface ChainEventConfig {
  /** WebSocket endpoint (e.g. wss://ws.virtengine.com/websocket). */
  wsUrl: string;
  /** Event types to subscribe to. Empty means all known types. */
  subscriptions: ChainEventType[];
  /** Whether to auto-reconnect on disconnect. Default true. */
  autoReconnect: boolean;
  /** Base reconnect delay in ms. Exponential backoff applied. Default 1000. */
  reconnectDelayMs: number;
  /** Maximum reconnect delay in ms. Default 30000. */
  maxReconnectDelayMs: number;
  /** Maximum number of reconnect attempts before giving up. Default 10. */
  maxReconnectAttempts: number;
}

/** Human-readable labels for event types, used in toast notifications. */
export const CHAIN_EVENT_LABELS: Record<ChainEventType, string> = {
  'order.created': 'New Order',
  'bid.created': 'New Bid',
  'allocation.status_changed': 'Allocation Updated',
  'settlement.executed': 'Settlement Executed',
  'hpc_job.status_changed': 'HPC Job Updated',
};

/** Default config values for the chain event client. */
export const DEFAULT_CHAIN_EVENT_CONFIG: ChainEventConfig = {
  wsUrl: 'wss://ws.virtengine.com/websocket',
  subscriptions: [
    'order.created',
    'bid.created',
    'allocation.status_changed',
    'settlement.executed',
    'hpc_job.status_changed',
  ],
  autoReconnect: true,
  reconnectDelayMs: 1000,
  maxReconnectDelayMs: 30000,
  maxReconnectAttempts: 10,
};
