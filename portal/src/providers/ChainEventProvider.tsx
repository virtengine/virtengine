/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Provider component that initialises the chain event WebSocket
 * connection when mounted and cleans up on unmount.
 */

'use client';

import type { ReactNode } from 'react';
import { useChainEvents } from '@/hooks/useChainEvents';
import type { ChainEventType } from '@/types/chain-events';

interface ChainEventProviderProps {
  children: ReactNode;
  /** Disable the WebSocket connection (e.g. for testing). */
  disabled?: boolean;
  /** Override WebSocket URL. */
  wsUrl?: string;
  /** Event types to subscribe to. Default all. */
  subscriptions?: ChainEventType[];
}

/**
 * ChainEventProvider wraps the app to establish a single chain-event
 * WebSocket connection. Place inside AppProviders.
 */
export function ChainEventProvider({
  children,
  disabled = false,
  wsUrl,
  subscriptions,
}: ChainEventProviderProps) {
  useChainEvents({
    enabled: !disabled,
    wsUrl,
    subscriptions,
  });

  return <>{children}</>;
}
