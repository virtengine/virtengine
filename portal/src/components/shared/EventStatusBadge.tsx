/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Live status badge showing the real-time connection status and
 * recent event indicators for chain subscriptions.
 */

'use client';

import { useChainEventStore, selectIsConnected } from '@/stores/chainEventStore';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';
import type { ConnectionStatus } from '@/types/chain-events';

const STATUS_CONFIG: Record<
  ConnectionStatus,
  { label: string; variant: 'success' | 'warning' | 'destructive' | 'secondary' }
> = {
  connected: { label: 'Live', variant: 'success' },
  connecting: { label: 'Connecting', variant: 'warning' },
  reconnecting: { label: 'Reconnecting', variant: 'warning' },
  disconnected: { label: 'Offline', variant: 'destructive' },
};

interface EventStatusBadgeProps {
  /** Additional CSS class names. */
  className?: string;
  /** Show the polling indicator when falling back to polling. */
  showPolling?: boolean;
}

/**
 * EventStatusBadge renders a colour-coded badge reflecting the
 * chain event WebSocket connection status.
 */
export function EventStatusBadge({ className, showPolling = true }: EventStatusBadgeProps) {
  const connectionStatus = useChainEventStore((s) => s.connectionStatus);
  const isConnected = useChainEventStore(selectIsConnected);
  const isPolling = useChainEventStore((s) => s.isPolling);

  const config = STATUS_CONFIG[connectionStatus];

  return (
    <Badge
      variant={isPolling && !isConnected ? 'warning' : config.variant}
      size="sm"
      dot
      className={cn('gap-1', className)}
    >
      {isPolling && !isConnected && showPolling ? 'Polling' : config.label}
    </Badge>
  );
}
