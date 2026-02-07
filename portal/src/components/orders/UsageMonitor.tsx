/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Usage monitoring component showing resource metrics, cost tracking, and alerts.
 */

'use client';

import { useCallback } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Progress } from '@/components/ui/Progress';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/Alert';
import type {
  OrderUsageData,
  ResourceUsageMetric,
  CostAccumulation,
  UsageAlert,
} from '@/features/orders/tracking-types';
import { formatCurrency } from '@/lib/utils';

// =============================================================================
// Main Component
// =============================================================================

interface UsageMonitorProps {
  usage: OrderUsageData;
  onDismissAlert?: (alertId: string) => void;
}

export function UsageMonitor({ usage, onDismissAlert }: UsageMonitorProps) {
  return (
    <div className="space-y-6">
      {/* Alerts */}
      {usage.alerts.filter((a) => !a.dismissed).length > 0 && (
        <UsageAlerts alerts={usage.alerts.filter((a) => !a.dismissed)} onDismiss={onDismissAlert} />
      )}

      {/* Resource Metrics */}
      <ResourceMetrics metrics={usage.metrics} />

      {/* Cost Tracking */}
      <CostTracking cost={usage.cost} />
    </div>
  );
}

// =============================================================================
// Resource Metrics
// =============================================================================

function ResourceMetrics({ metrics }: { metrics: ResourceUsageMetric[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Resource Usage</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid gap-4 sm:grid-cols-2">
          {metrics.map((metric) => (
            <ResourceMetricCard key={metric.resourceType} metric={metric} />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

function ResourceMetricCard({ metric }: { metric: ResourceUsageMetric }) {
  const percentage =
    metric.allocated > 0 ? Math.round((metric.current / metric.allocated) * 100) : 0;

  const variant: 'default' | 'success' | 'warning' | 'destructive' =
    percentage >= 90 ? 'destructive' : percentage >= 75 ? 'warning' : 'success';

  return (
    <div className="rounded-lg border border-border p-4">
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium">{metric.label}</span>
        <span className="text-sm text-muted-foreground">{percentage}%</span>
      </div>
      <div className="mt-2">
        <Progress value={percentage} variant={variant} size="sm" />
      </div>
      <div className="mt-2 flex justify-between text-xs text-muted-foreground">
        <span>
          {formatMetricValue(metric.current, metric.unit)} /{' '}
          {formatMetricValue(metric.allocated, metric.unit)}
        </span>
        <span>{metric.unit}</span>
      </div>

      {/* Mini sparkline using bar chart */}
      {metric.history.length > 0 && (
        <div className="mt-3 flex h-8 items-end gap-px">
          {metric.history.slice(-12).map((point) => {
            const maxVal = Math.max(...metric.history.slice(-12).map((p) => p.value), 1);
            const height = Math.max(2, (point.value / maxVal) * 100);
            return (
              <div
                key={`${metric.resourceType}-${point.timestamp}`}
                className="flex-1 rounded-t-sm bg-primary/30 transition-all hover:bg-primary/60"
                style={{ height: `${height}%` }}
                title={`${formatMetricValue(point.value, metric.unit)} ${metric.unit}`}
              />
            );
          })}
        </div>
      )}
    </div>
  );
}

function formatMetricValue(value: number, unit: string): string {
  if (unit === '%') return `${Math.round(value)}`;
  if (unit === 'GB' || unit === 'TB') return value.toFixed(1);
  return value.toLocaleString();
}

// =============================================================================
// Cost Tracking
// =============================================================================

function CostTracking({ cost }: { cost: CostAccumulation }) {
  const escrowPercent =
    cost.escrowTotal > 0 ? Math.round((cost.escrowBalance / cost.escrowTotal) * 100) : 0;

  const escrowVariant: 'success' | 'warning' | 'destructive' =
    escrowPercent < 20 ? 'destructive' : escrowPercent < 50 ? 'warning' : 'success';

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Cost Summary</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-3 sm:grid-cols-2">
          <div className="rounded-lg bg-muted/50 p-3">
            <div className="text-xs text-muted-foreground">Current Period</div>
            <div className="mt-1 text-xl font-bold">
              {formatCurrency(cost.currentPeriodCost, cost.currency)}
            </div>
          </div>
          <div className="rounded-lg bg-muted/50 p-3">
            <div className="text-xs text-muted-foreground">Projected Monthly</div>
            <div className="mt-1 text-xl font-bold">
              {formatCurrency(cost.projectedMonthlyCost, cost.currency)}
            </div>
          </div>
        </div>

        {/* Escrow Balance */}
        <div className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Escrow Balance</span>
            <span className="font-medium">
              {formatCurrency(cost.escrowBalance, cost.currency)} /{' '}
              {formatCurrency(cost.escrowTotal, cost.currency)}
            </span>
          </div>
          <Progress value={escrowPercent} variant={escrowVariant} size="sm" />
          {escrowPercent < 20 && (
            <p className="text-xs text-destructive">
              Low escrow balance. Consider adding funds to avoid service interruption.
            </p>
          )}
        </div>

        {/* Cost Sparkline */}
        {cost.costHistory.length > 0 && (
          <div>
            <div className="mb-1 text-xs text-muted-foreground">Cost Rate (24h)</div>
            <div className="flex h-12 items-end gap-px">
              {cost.costHistory.slice(-24).map((point) => {
                const maxVal = Math.max(...cost.costHistory.slice(-24).map((p) => p.value), 1);
                const height = Math.max(2, (point.value / maxVal) * 100);
                return (
                  <div
                    key={`cost-${point.timestamp}`}
                    className="flex-1 rounded-t-sm bg-success/40 transition-all hover:bg-success/70"
                    style={{ height: `${height}%` }}
                  />
                );
              })}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// =============================================================================
// Usage Alerts
// =============================================================================

function UsageAlerts({
  alerts,
  onDismiss,
}: {
  alerts: UsageAlert[];
  onDismiss?: (alertId: string) => void;
}) {
  const handleDismiss = useCallback(
    (alertId: string) => {
      onDismiss?.(alertId);
    },
    [onDismiss]
  );

  const alertVariants: Record<string, 'default' | 'destructive'> = {
    critical: 'destructive',
    warning: 'default',
    info: 'default',
  };

  return (
    <div className="space-y-2">
      {alerts.map((alert) => (
        <Alert key={alert.id} variant={alertVariants[alert.type] ?? 'default'}>
          <div className="flex items-start justify-between">
            <div>
              <AlertTitle className="flex items-center gap-2">
                <Badge
                  variant={
                    alert.type === 'critical'
                      ? 'destructive'
                      : alert.type === 'warning'
                        ? 'warning'
                        : 'info'
                  }
                  size="sm"
                >
                  {alert.type.toUpperCase()}
                </Badge>
                {alert.resourceType}
              </AlertTitle>
              <AlertDescription className="mt-1">{alert.message}</AlertDescription>
            </div>
            {onDismiss && (
              <Button
                variant="ghost"
                size="icon-sm"
                onClick={() => handleDismiss(alert.id)}
                aria-label="Dismiss alert"
              >
                ✕
              </Button>
            )}
          </div>
        </Alert>
      ))}
    </div>
  );
}
