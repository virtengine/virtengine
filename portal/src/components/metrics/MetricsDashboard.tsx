/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Main metrics dashboard page component.
 * Aggregates resource metrics across all deployments with time-series charts,
 * provider breakdown, and alert integration.
 */

'use client';

import { useEffect } from 'react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/Select';
import { Skeleton } from '@/components/ui/Skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Wifi, WifiOff } from 'lucide-react';
import {
  useMetricsStore,
  selectCPUTrend,
  selectMemoryTrend,
  selectFiringAlerts,
} from '@/stores/metricsStore';
import { MetricCard } from './MetricCard';
import { TimeSeriesChart } from './TimeSeriesChart';
import { ProviderBreakdown } from './ProviderBreakdown';
import { AlertsPanel } from './AlertsPanel';
import { CustomDashboard } from './CustomDashboard';
import type { TimeRange } from '@virtengine/portal/types/metrics';
import { TIME_RANGE_LABELS, granularityForRange } from '@virtengine/portal/types/metrics';

export function MetricsDashboard() {
  const {
    summary,
    deploymentMetrics,
    selectedTimeRange,
    isLoading,
    isStreaming,
    fetchMetrics,
    setTimeRange,
    toggleStreaming,
  } = useMetricsStore();

  const cpuTrend = useMetricsStore(selectCPUTrend);
  const memoryTrend = useMetricsStore(selectMemoryTrend);
  const firingAlerts = useMetricsStore(selectFiringAlerts);

  useEffect(() => {
    void fetchMetrics();
  }, [fetchMetrics]);

  // Simulate streaming with polling
  useEffect(() => {
    if (!isStreaming) return;
    const interval = setInterval(() => {
      void fetchMetrics();
    }, 30000);
    return () => clearInterval(interval);
  }, [isStreaming, fetchMetrics]);

  if (isLoading && !summary) {
    return <MetricsDashboardSkeleton />;
  }

  const firstDeployment = deploymentMetrics[0];
  const granularity = granularityForRange(selectedTimeRange);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold">Metrics Dashboard</h1>
          <p className="text-sm text-muted-foreground">Resource usage across all deployments</p>
        </div>

        <div className="flex items-center gap-3">
          {/* Streaming indicator */}
          <Button size="sm" variant={isStreaming ? 'default' : 'outline'} onClick={toggleStreaming}>
            {isStreaming ? <Wifi className="mr-1 h-3 w-3" /> : <WifiOff className="mr-1 h-3 w-3" />}
            {isStreaming ? 'Live' : 'Paused'}
          </Button>

          {/* Alert badge */}
          {firingAlerts.length > 0 && (
            <Badge variant="destructive" dot>
              {firingAlerts.length} alert{firingAlerts.length !== 1 ? 's' : ''}
            </Badge>
          )}

          {/* Time range selector */}
          <Select value={selectedTimeRange} onValueChange={(v) => setTimeRange(v as TimeRange)}>
            <SelectTrigger className="w-[140px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {(Object.keys(TIME_RANGE_LABELS) as TimeRange[]).map((range) => (
                <SelectItem key={range} value={range}>
                  {TIME_RANGE_LABELS[range]}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      <Tabs defaultValue="overview">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="custom">Custom Dashboard</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-6">
          {/* Summary Cards */}
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <MetricCard
              title="CPU Usage"
              value={summary?.totalCPU.used ?? 0}
              limit={summary?.totalCPU.limit ?? 0}
              unit={summary?.totalCPU.unit ?? 'cores'}
              trend={cpuTrend}
            />
            <MetricCard
              title="Memory Usage"
              value={summary?.totalMemory.used ?? 0}
              limit={summary?.totalMemory.limit ?? 0}
              unit={summary?.totalMemory.unit ?? 'GB'}
              trend={memoryTrend}
            />
            <MetricCard
              title="Storage Usage"
              value={summary?.totalStorage.used ?? 0}
              limit={summary?.totalStorage.limit ?? 0}
              unit={summary?.totalStorage.unit ?? 'GB'}
            />
            <MetricCard
              title="Active Deployments"
              value={summary?.activeDeployments ?? 0}
              subtitle={`across ${summary?.totalProviders ?? 0} providers`}
            />
          </div>

          {/* Health Overview */}
          {summary?.healthOverview && (
            <div className="flex gap-3">
              <Badge variant="success" dot>
                {summary.healthOverview.healthy} Healthy
              </Badge>
              {summary.healthOverview.degraded > 0 && (
                <Badge variant="warning" dot>
                  {summary.healthOverview.degraded} Degraded
                </Badge>
              )}
              {summary.healthOverview.warning > 0 && (
                <Badge variant="warning" dot>
                  {summary.healthOverview.warning} Warning
                </Badge>
              )}
              {summary.healthOverview.critical > 0 && (
                <Badge variant="destructive" dot>
                  {summary.healthOverview.critical} Critical
                </Badge>
              )}
            </div>
          )}

          {/* Charts */}
          <div className="grid gap-6 lg:grid-cols-2">
            <TimeSeriesChart
              title="CPU Usage"
              series={firstDeployment?.history.cpu}
              granularity={granularity}
              isLoading={isLoading}
            />
            <TimeSeriesChart
              title="Memory Usage"
              series={firstDeployment?.history.memory}
              granularity={granularity}
              isLoading={isLoading}
            />
          </div>

          <div className="grid gap-6 lg:grid-cols-2">
            <TimeSeriesChart
              title="Storage Usage"
              series={firstDeployment?.history.storage}
              granularity={granularity}
              isLoading={isLoading}
            />
            <TimeSeriesChart
              title="Network I/O"
              series={firstDeployment?.history.network}
              granularity={granularity}
              isLoading={isLoading}
            />
          </div>

          {/* Breakdown + Alerts */}
          <div className="grid gap-6 lg:grid-cols-2">
            <ProviderBreakdown providers={summary?.byProvider ?? []} />
            <AlertsPanel />
          </div>

          {/* Deployment Detail Table */}
          {summary && summary.byDeployment.length > 0 && (
            <div className="rounded-lg border">
              <div className="p-4">
                <h2 className="text-sm font-medium">Deployment Metrics</h2>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead className="border-t bg-muted/50">
                    <tr>
                      <th className="px-4 py-2 text-left font-medium">Deployment</th>
                      <th className="px-4 py-2 text-left font-medium">Provider</th>
                      <th className="px-4 py-2 text-right font-medium">CPU</th>
                      <th className="px-4 py-2 text-right font-medium">Memory</th>
                      <th className="px-4 py-2 text-right font-medium">Storage</th>
                      <th className="px-4 py-2 text-right font-medium">Services</th>
                    </tr>
                  </thead>
                  <tbody>
                    {summary.byDeployment.map((d) => (
                      <tr key={d.deploymentId} className="border-t">
                        <td className="px-4 py-2 font-mono text-xs">{d.deploymentId}</td>
                        <td className="px-4 py-2">{d.provider}</td>
                        <td className="px-4 py-2 text-right">
                          {d.current.cpu.used.toFixed(1)}/{d.current.cpu.limit} {d.current.cpu.unit}
                        </td>
                        <td className="px-4 py-2 text-right">
                          {d.current.memory.used.toFixed(0)}/{d.current.memory.limit}{' '}
                          {d.current.memory.unit}
                        </td>
                        <td className="px-4 py-2 text-right">
                          {d.current.storage.used.toFixed(0)}/{d.current.storage.limit}{' '}
                          {d.current.storage.unit}
                        </td>
                        <td className="px-4 py-2 text-right">{d.services.length}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </TabsContent>

        <TabsContent value="custom">
          <CustomDashboard />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function MetricsDashboardSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-9 w-36" />
      </div>
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          // eslint-disable-next-line react/no-array-index-key
          <Skeleton key={i} className="h-32" />
        ))}
      </div>
      <div className="grid gap-6 lg:grid-cols-2">
        <Skeleton className="h-64" />
        <Skeleton className="h-64" />
      </div>
    </div>
  );
}
