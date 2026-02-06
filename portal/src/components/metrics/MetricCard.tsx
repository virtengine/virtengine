/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Metric summary card with value, limit, progress bar, and trend indicator.
 */

'use client';

import { TrendingUp, TrendingDown, Minus } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { cn } from '@/lib/utils';
import type { MetricTrend } from '@virtengine/portal/types/metrics';

interface MetricCardProps {
  title: string;
  value: number;
  limit?: number;
  unit?: string;
  subtitle?: string;
  trend?: MetricTrend;
}

export function MetricCard({ title, value, limit, unit, subtitle, trend }: MetricCardProps) {
  const percent = limit ? (value / limit) * 100 : 0;
  const isWarning = percent > 80;
  const isCritical = percent > 95;

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">{title}</CardTitle>
        {trend && trend.direction !== 'stable' && (
          <div
            className={cn(
              'flex items-center gap-1 text-xs',
              trend.direction === 'up' && 'text-red-500',
              trend.direction === 'down' && 'text-green-500'
            )}
          >
            {trend.direction === 'up' ? (
              <TrendingUp className="h-3 w-3" />
            ) : (
              <TrendingDown className="h-3 w-3" />
            )}
            <span>{trend.percent}%</span>
          </div>
        )}
        {trend && trend.direction === 'stable' && (
          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <Minus className="h-3 w-3" />
          </div>
        )}
      </CardHeader>
      <CardContent>
        <div className="flex items-baseline gap-1">
          <span
            className={cn(
              'text-2xl font-bold',
              isCritical && 'text-red-500',
              isWarning && !isCritical && 'text-yellow-500'
            )}
          >
            {value.toFixed(1)}
          </span>
          {limit && (
            <span className="text-sm text-muted-foreground">
              / {limit} {unit}
            </span>
          )}
          {!limit && unit && <span className="text-sm text-muted-foreground">{unit}</span>}
        </div>

        {subtitle && <p className="mt-1 text-xs text-muted-foreground">{subtitle}</p>}

        {limit !== undefined && limit !== null && limit > 0 && (
          <div className="mt-3">
            <div className="h-1.5 overflow-hidden rounded-full bg-muted">
              <div
                className={cn(
                  'h-full transition-all',
                  isCritical && 'bg-red-500',
                  isWarning && !isCritical && 'bg-yellow-500',
                  !isWarning && 'bg-primary'
                )}
                style={{ width: `${Math.min(percent, 100)}%` }}
              />
            </div>
            <p className="mt-1 text-xs text-muted-foreground">{percent.toFixed(1)}% used</p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
