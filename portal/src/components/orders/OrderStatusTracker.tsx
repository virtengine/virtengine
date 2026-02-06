/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Order status tracker with timeline, progress indicator, and status badges.
 */

'use client';

import type { OrderStatus } from '@/stores/orderStore';
import { Badge } from '@/components/ui/Badge';
import { Progress } from '@/components/ui/Progress';
import {
  type OrderStatusTimeline,
  ORDER_STATUS_CONFIG,
  getOrderProgress,
  estimateTimeRemaining,
  isOrderTerminal,
} from '@/features/orders/tracking-types';
import { formatRelativeTime } from '@/lib/utils';

// =============================================================================
// Status Header
// =============================================================================

interface OrderStatusHeaderProps {
  status: OrderStatus;
  estimatedCompletion?: string;
}

export function OrderStatusHeader({ status, estimatedCompletion }: OrderStatusHeaderProps) {
  const config = ORDER_STATUS_CONFIG[status];
  const progress = getOrderProgress(status);
  const timeRemaining = estimateTimeRemaining(estimatedCompletion);
  const terminal = isOrderTerminal(status);

  const progressVariant =
    status === 'failed'
      ? 'destructive'
      : status === 'running'
        ? 'success'
        : status === 'paused'
          ? 'warning'
          : 'default';

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <Badge variant={config.variant} dot size="lg">
          {config.label}
        </Badge>
        {timeRemaining && !terminal && (
          <span className="text-sm text-muted-foreground">
            {status === 'deploying' ? 'Est. ready in ' : 'Expires in '}
            {timeRemaining}
          </span>
        )}
      </div>

      {!terminal && status !== 'failed' && (
        <div className="space-y-1">
          <div className="flex justify-between text-xs text-muted-foreground">
            <span>Progress</span>
            <span>{progress}%</span>
          </div>
          <Progress value={progress} variant={progressVariant} size="sm" />
        </div>
      )}
    </div>
  );
}

// =============================================================================
// Status Timeline
// =============================================================================

interface OrderStatusTimelineProps {
  timeline: OrderStatusTimeline;
}

export function OrderStatusTimelineView({ timeline }: OrderStatusTimelineProps) {
  const { events } = timeline;
  const sortedEvents = [...events].sort(
    (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
  );

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold">Timeline</h3>
      <div className="space-y-0">
        {sortedEvents.map((event, index) => {
          const isFirst = index === 0;
          const isLast = index === sortedEvents.length - 1;
          const config = ORDER_STATUS_CONFIG[event.status];

          return (
            <div key={event.id} className="flex gap-4">
              <div className="relative flex flex-col items-center">
                <div
                  className={`z-10 h-3 w-3 rounded-full ${
                    isFirst
                      ? config.variant === 'success'
                        ? 'bg-success'
                        : config.variant === 'destructive'
                          ? 'bg-destructive'
                          : config.variant === 'warning'
                            ? 'bg-warning'
                            : 'bg-primary'
                      : 'bg-muted-foreground/40'
                  }`}
                />
                {!isLast && <div className="w-px flex-1 bg-border" />}
              </div>
              <div className={`pb-6 ${isFirst ? '' : 'opacity-70'}`}>
                <div className="flex items-center gap-2">
                  <span className="font-medium">{event.title}</span>
                  {isFirst && (
                    <Badge variant={config.variant} size="sm">
                      Current
                    </Badge>
                  )}
                </div>
                {event.description && (
                  <p className="mt-0.5 text-sm text-muted-foreground">{event.description}</p>
                )}
                <p className="mt-1 text-xs text-muted-foreground">
                  {formatRelativeTime(event.timestamp)}
                </p>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// =============================================================================
// Connection Status Indicator
// =============================================================================

interface ConnectionStatusProps {
  status: 'connecting' | 'connected' | 'disconnected' | 'error';
}

export function ConnectionStatusIndicator({ status }: ConnectionStatusProps) {
  const config = {
    connecting: { label: 'Connecting...', className: 'bg-warning animate-pulse' },
    connected: { label: 'Live', className: 'bg-success' },
    disconnected: { label: 'Offline', className: 'bg-muted-foreground' },
    error: { label: 'Error', className: 'bg-destructive' },
  }[status];

  return (
    <div className="flex items-center gap-2 text-xs text-muted-foreground">
      <span className={`h-2 w-2 rounded-full ${config.className}`} />
      {config.label}
    </div>
  );
}
