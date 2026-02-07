/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Zustand store for chain event state: connection status, recent events,
 * and actions to connect/disconnect the WebSocket client.
 */

import { create } from 'zustand';
import type { ChainEvent, ChainEventConfig, ConnectionStatus } from '@/types/chain-events';
import { ChainEventClient } from '@/lib/chain-events';

const MAX_EVENTS = 100;

// =============================================================================
// Store Interface
// =============================================================================

export interface ChainEventState {
  /** Current WebSocket connection status. */
  connectionStatus: ConnectionStatus;
  /** Recent chain events (newest first), capped at MAX_EVENTS. */
  events: ChainEvent[];
  /** Whether polling fallback is active. */
  isPolling: boolean;
  /** Error message from last connection attempt. */
  error: string | null;
}

export interface ChainEventActions {
  /** Connect the WebSocket client. */
  connect: (config?: Partial<ChainEventConfig>) => void;
  /** Disconnect the WebSocket client. */
  disconnect: () => void;
  /** Add a chain event (used by client or polling). */
  addEvent: (event: ChainEvent) => void;
  /** Clear all stored events. */
  clearEvents: () => void;
  /** Enable polling fallback mode. */
  enablePolling: () => void;
  /** Disable polling fallback mode. */
  disablePolling: () => void;
  /** Clear the error state. */
  clearError: () => void;
}

export type ChainEventStore = ChainEventState & ChainEventActions;

// =============================================================================
// Client singleton (lives outside the store to avoid serialization issues)
// =============================================================================

let clientInstance: ChainEventClient | null = null;

/** Retrieve the active ChainEventClient (for testing/external use). */
export function getChainEventClient(): ChainEventClient | null {
  return clientInstance;
}

// =============================================================================
// Store
// =============================================================================

const initialState: ChainEventState = {
  connectionStatus: 'disconnected',
  events: [],
  isPolling: false,
  error: null,
};

export const useChainEventStore = create<ChainEventStore>()((set, get) => ({
  ...initialState,

  connect: (config) => {
    // Disconnect existing client first.
    const { disconnect } = get();
    if (clientInstance) {
      disconnect();
    }

    const client = new ChainEventClient(config);
    clientInstance = client;

    client.onStatusChange((status) => {
      set({ connectionStatus: status });
      if (status === 'disconnected' && get().connectionStatus !== 'disconnected') {
        set({ error: 'WebSocket disconnected' });
      }
    });

    client.onEvent((event) => {
      get().addEvent(event);
    });

    try {
      client.connect();
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Connection failed' });
    }
  },

  disconnect: () => {
    if (clientInstance) {
      clientInstance.disconnect();
      clientInstance = null;
    }
    set({ connectionStatus: 'disconnected' });
  },

  addEvent: (event) => {
    set((state) => ({
      events: [event, ...state.events].slice(0, MAX_EVENTS),
    }));
  },

  clearEvents: () => {
    set({ events: [] });
  },

  enablePolling: () => {
    set({ isPolling: true });
  },

  disablePolling: () => {
    set({ isPolling: false });
  },

  clearError: () => {
    set({ error: null });
  },
}));

// =============================================================================
// Selectors
// =============================================================================

export const selectIsConnected = (state: ChainEventStore) => state.connectionStatus === 'connected';

export const selectRecentEvents = (limit: number) => (state: ChainEventStore) =>
  state.events.slice(0, limit);

export const selectEventsByType = (type: ChainEvent['type']) => (state: ChainEventStore) =>
  state.events.filter((e) => e.type === type);
