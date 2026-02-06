/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useEffect, useState } from 'react';
import { Skeleton } from '@/components/ui/Skeleton';
import type { UsageHistoryPoint, UsageGranularity } from '@virtengine/portal/types/billing';

interface UsageChartProps {
  data?: UsageHistoryPoint[];
  granularity: UsageGranularity;
  loading?: boolean;
}

const RESOURCE_COLORS: Record<string, string> = {
  cpu: 'bg-blue-500',
  memory: 'bg-green-500',
  storage: 'bg-amber-500',
  bandwidth: 'bg-purple-500',
  gpu: 'bg-red-500',
};

const RESOURCE_LABELS: Record<string, string> = {
  cpu: 'CPU',
  memory: 'Memory',
  storage: 'Storage',
  bandwidth: 'Bandwidth',
  gpu: 'GPU',
};

/**
 * Stacked area chart for resource usage over time.
 * Uses native HTML/CSS for zero-dependency rendering.
 */
export function UsageChart({ data, granularity, loading }: UsageChartProps) {
  const [maxValue, setMaxValue] = useState(1);

  useEffect(() => {
    if (data && data.length > 0) {
      const max = Math.max(...data.map((d) => d.cpu + d.memory + d.storage + d.bandwidth + d.gpu));
      setMaxValue(max > 0 ? max : 1);
    }
  }, [data]);

  if (loading) {
    return <Skeleton className="h-64 w-full" />;
  }

  if (!data || data.length === 0) {
    return (
      <div className="flex h-64 items-center justify-center text-sm text-muted-foreground">
        No usage data available for this period
      </div>
    );
  }

  const formatLabel = (timestamp: Date): string => {
    switch (granularity) {
      case 'hour':
        return timestamp.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
      case 'day':
        return timestamp.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
      case 'week':
        return `W${Math.ceil(timestamp.getDate() / 7)}`;
      case 'month':
        return timestamp.toLocaleDateString(undefined, { month: 'short', year: '2-digit' });
      default:
        return timestamp.toLocaleDateString();
    }
  };

  return (
    <div className="space-y-4">
      {/* Chart */}
      <div className="flex h-64 items-end gap-px" role="img" aria-label="Resource usage chart">
        {data.map((point) => {
          const total = point.cpu + point.memory + point.storage + point.bandwidth + point.gpu;
          const heightPercent = (total / maxValue) * 100;

          return (
            <div
              key={`usage-bar-${point.timestamp.toISOString()}`}
              className="group relative flex flex-1 flex-col justify-end"
              style={{ height: `${Math.max(heightPercent, 1)}%` }}
              title={`${formatLabel(point.timestamp)}: ${total.toFixed(1)} total`}
            >
              {/* Stacked segments */}
              {point.gpu > 0 && (
                <div
                  className={`w-full ${RESOURCE_COLORS.gpu}`}
                  style={{ height: `${(point.gpu / total) * 100}%` }}
                />
              )}
              <div
                className={`w-full ${RESOURCE_COLORS.bandwidth}`}
                style={{ height: `${(point.bandwidth / total) * 100}%` }}
              />
              <div
                className={`w-full ${RESOURCE_COLORS.storage}`}
                style={{ height: `${(point.storage / total) * 100}%` }}
              />
              <div
                className={`w-full ${RESOURCE_COLORS.memory}`}
                style={{ height: `${(point.memory / total) * 100}%` }}
              />
              <div
                className={`w-full rounded-t ${RESOURCE_COLORS.cpu}`}
                style={{ height: `${(point.cpu / total) * 100}%` }}
              />
            </div>
          );
        })}
      </div>

      {/* Legend */}
      <div className="flex flex-wrap gap-4">
        {Object.entries(RESOURCE_LABELS).map(([key, label]) => (
          <div key={`legend-${key}`} className="flex items-center gap-1.5">
            <div className={`h-3 w-3 rounded-sm ${RESOURCE_COLORS[key]}`} />
            <span className="text-xs text-muted-foreground">{label}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
