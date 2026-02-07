/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * React hook for subscribing to chain events with automatic toast
 * notifications and polling fallback when WebSocket is unavailable.
 */

'use client';

import { useEffect, useRef, useCallback } from 'react';
import { env } from '@/config/env';
import { useChainEventStore, selectIsConnected } from '@/stores/chainEventStore';
import { toast } from '@/hooks/use-toast';
import { CHAIN_EVENT_LABELS } from '@/types/chain-events';
import type { ChainEvent, ChainEventConfig, ChainEventType } from '@/types/chain-events';

export interface UseChainEventsOptions {
  /** Whether to connect automatically. Default true. */
  enabled?: boolean;
  /** Override the default WS URL from env. */
  wsUrl?: string;
  /** Event types to subscribe to. Default all. */
  subscriptions?: ChainEventType[];
  /** Show toast notifications for incoming events. Default true. */
  showToasts?: boolean;
  /** Callback invoked for every incoming event. */
  onEvent?: (event: ChainEvent) => void;
  /** Polling interval in ms when falling back to polling. Default 10000. */
  pollingIntervalMs?: number;
}

const TOAST_VARIANT_MAP: Record<ChainEventType, 'default' | 'success' | 'info' | 'warning'> = {
  'order.created': 'info',
  'bid.created': 'info',
  'allocation.status_changed': 'default',
  'settlement.executed': 'success',
  'hpc_job.status_changed': 'default',
};

/**
 * useChainEvents connects to the CometBFT WebSocket for real-time chain
 * events. Falls back to polling the REST API when WebSocket fails.
 */
export function useChainEvents(options: UseChainEventsOptions = {}) {
  const {
    enabled = true,
    wsUrl,
    subscriptions,
    showToasts = true,
    onEvent,
    pollingIntervalMs = 10_000,
  } = options;

  const connect = useChainEventStore((s) => s.connect);
  const disconnect = useChainEventStore((s) => s.disconnect);
  const connectionStatus = useChainEventStore((s) => s.connectionStatus);
  const events = useChainEventStore((s) => s.events);
  const isConnected = useChainEventStore(selectIsConnected);
  const isPolling = useChainEventStore((s) => s.isPolling);
  const enablePolling = useChainEventStore((s) => s.enablePolling);
  const disablePolling = useChainEventStore((s) => s.disablePolling);
  const addEvent = useChainEventStore((s) => s.addEvent);
  const error = useChainEventStore((s) => s.error);

  // Track latest onEvent ref to avoid re-subscriptions.
  const onEventRef = useRef(onEvent);
  onEventRef.current = onEvent;

  // Track seen event IDs for dedup in polling.
  const seenIdsRef = useRef(new Set<string>());

  // Toast + callback handler for incoming events.
  const handleEvent = useCallback(
    (event: ChainEvent) => {
      if (showToasts) {
        const label = CHAIN_EVENT_LABELS[event.type] ?? event.type;
        const variant = TOAST_VARIANT_MAP[event.type] ?? 'default';
        toast({
          title: label,
          description: `Block ${event.blockHeight}`,
          variant,
        });
      }
      onEventRef.current?.(event);
    },
    [showToasts]
  );

  // Subscribe to store events and invoke handler.
  const prevLenRef = useRef(0);
  useEffect(() => {
    if (events.length > prevLenRef.current) {
      const newEvents = events.slice(0, events.length - prevLenRef.current);
      for (const evt of newEvents) {
        handleEvent(evt);
      }
    }
    prevLenRef.current = events.length;
  }, [events, handleEvent]);

  // Connect/disconnect lifecycle.
  useEffect(() => {
    if (!enabled) return;

    const config: Partial<ChainEventConfig> = {
      wsUrl: wsUrl ?? env.chainWs + '/websocket',
    };
    if (subscriptions) {
      config.subscriptions = subscriptions;
    }

    connect(config);

    return () => {
      disconnect();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enabled, wsUrl]);

  // Polling fallback: activate after WebSocket stays disconnected.
  const failTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  useEffect(() => {
    if (!enabled) return;

    if (connectionStatus === 'disconnected' && !isPolling) {
      // Wait 5s before enabling polling to allow reconnect attempts.
      failTimerRef.current = setTimeout(() => {
        enablePolling();
      }, 5000);
    } else if (connectionStatus === 'connected' && isPolling) {
      disablePolling();
    }

    return () => {
      if (failTimerRef.current) {
        clearTimeout(failTimerRef.current);
        failTimerRef.current = null;
      }
    };
  }, [connectionStatus, enabled, isPolling, enablePolling, disablePolling]);

  // Polling interval.
  useEffect(() => {
    if (!isPolling || !enabled) return;

    const poll = () => {
      void (async () => {
        try {
          const res = await fetch(
            `${env.chainRest}/cosmos/tx/v1beta1/txs?events=tx.height>0&order_by=ORDER_BY_DESC&pagination.limit=5`
          );
          if (!res.ok) return;
          const data = (await res.json()) as PollingResponse;

          const txResponses = data.tx_responses ?? [];
          for (const tx of txResponses) {
            const id = `poll-${tx.txhash}`;
            if (seenIdsRef.current.has(id)) continue;
            seenIdsRef.current.add(id);

            // Cap the dedup set.
            if (seenIdsRef.current.size > 500) {
              const arr = Array.from(seenIdsRef.current);
              seenIdsRef.current = new Set(arr.slice(arr.length - 250));
            }

            for (const evt of tx.events ?? []) {
              const type = mapRawEventType(evt.type);
              if (!type) continue;
              const attrs: Record<string, string> = {};
              for (const attr of evt.attributes ?? []) {
                attrs[attr.key] = attr.value;
              }
              addEvent({
                id: `${id}-${evt.type}`,
                type,
                blockHeight: parseInt(tx.height ?? '0', 10),
                timestamp: new Date(tx.timestamp),
                txHash: tx.txhash ?? '',
                attributes: attrs,
              });
            }
          }
        } catch {
          // Polling errors are non-fatal.
        }
      })();
    };

    const interval = setInterval(poll, pollingIntervalMs);

    return () => clearInterval(interval);
  }, [isPolling, enabled, pollingIntervalMs, addEvent]);

  return {
    connectionStatus,
    isConnected,
    isPolling,
    events,
    error,
  };
}

/** Shape of the Cosmos SDK tx search REST response (minimal). */
interface PollingTxEvent {
  type: string;
  attributes: Array<{ key: string; value: string }>;
}

interface PollingTxResponse {
  txhash: string;
  height: string;
  timestamp: string;
  events: PollingTxEvent[];
}

interface PollingResponse {
  tx_responses?: PollingTxResponse[];
}

function mapRawEventType(raw: string): ChainEventType | null {
  const map: Record<string, ChainEventType> = {
    create_order: 'order.created',
    create_bid: 'bid.created',
    update_allocation_status: 'allocation.status_changed',
    execute_settlement: 'settlement.executed',
    update_hpc_job_status: 'hpc_job.status_changed',
  };
  return map[raw] ?? null;
}
