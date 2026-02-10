/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useEffect, useState } from 'react';
import { Skeleton } from '@/components/ui/Skeleton';
import type { UsageHistoryPoint } from '@virtengine/portal/types/billing';

interface CostTrendChartProps {
  data?: UsageHistoryPoint[];
  loading?: boolean;
}

/**
 * Simple bar chart for cost trend visualization.
 * Uses native HTML/CSS for zero-dependency rendering.
 */
export function CostTrendChart({ data, loading }: CostTrendChartProps) {
  const [maxCost, setMaxCost] = useState(0);

  useEffect(() => {
    if (data && data.length > 0) {
      const max = Math.max(...data.map((d) => parseFloat(d.cost)));
      setMaxCost(max > 0 ? max : 1);
    }
  }, [data]);

  if (loading) {
    return (
      <div className="flex h-48 items-end gap-1">
        {Array.from({ length: 14 }, (_, i) => (
          <Skeleton
            key={`cost-trend-skeleton-${i}`}
            className="flex-1"
            style={{ height: `${30 + Math.random() * 60}%` }}
          />
        ))}
      </div>
    );
  }

  if (!data || data.length === 0) {
    return (
      <div className="flex h-48 items-center justify-center text-sm text-muted-foreground">
        No cost data available
      </div>
    );
  }

  return (
    <div className="space-y-2">
      <div className="flex h-48 items-end gap-1" role="img" aria-label="Cost trend chart">
        {data.map((point, index) => {
          const height = maxCost > 0 ? (parseFloat(point.cost) / maxCost) * 100 : 0;
          const dateLabel = point.timestamp.toLocaleDateString(undefined, {
            month: 'short',
            day: 'numeric',
          });
          return (
            <div
              key={`cost-bar-${point.timestamp.toISOString()}`}
              className="group relative flex-1"
              title={`${dateLabel}: ${point.cost} VIRT`}
            >
              <div
                className="w-full rounded-t bg-primary transition-colors hover:bg-primary/80"
                style={{ height: `${Math.max(height, 2)}%` }}
              />
              {/* Tooltip on hover */}
              <div className="absolute bottom-full left-1/2 z-10 mb-1 hidden -translate-x-1/2 whitespace-nowrap rounded bg-popover px-2 py-1 text-xs shadow group-hover:block">
                <p className="font-medium">{point.cost} VIRT</p>
                <p className="text-muted-foreground">{dateLabel}</p>
              </div>
              {/* Label every 3rd bar on wide screens */}
              {index % 3 === 0 && (
                <p className="mt-1 hidden text-center text-2xs text-muted-foreground lg:block">
                  {dateLabel}
                </p>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
