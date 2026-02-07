/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Dashboard widget wrapper that renders the appropriate widget type.
 */

'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { X } from 'lucide-react';
import { MetricCard } from './MetricCard';
import { TimeSeriesChart } from './TimeSeriesChart';
import { AlertsPanel } from './AlertsPanel';
import { ProviderBreakdown } from './ProviderBreakdown';
import { useMetricsStore, selectCPUTrend, selectMemoryTrend } from '@/stores/metricsStore';
import type { DashboardWidget as DashboardWidgetType } from '@virtengine/portal/types/metrics';
import { granularityForRange } from '@virtengine/portal/types/metrics';
import type { TimeRange } from '@virtengine/portal/types/metrics';

interface DashboardWidgetProps {
  widget: DashboardWidgetType;
  isEditing: boolean;
  onRemove: () => void;
}

export function DashboardWidget({ widget, isEditing, onRemove }: DashboardWidgetProps) {
  const summary = useMetricsStore((s) => s.summary);
  const deploymentMetrics = useMetricsStore((s) => s.deploymentMetrics);
  const timeRange = useMetricsStore((s) => s.selectedTimeRange);
  const cpuTrend = useMetricsStore(selectCPUTrend);
  const memoryTrend = useMetricsStore(selectMemoryTrend);

  function renderWidget() {
    switch (widget.type) {
      case 'metric-card': {
        if (!summary) return <WidgetPlaceholder title={widget.title} />;

        switch (widget.config.metric) {
          case 'cpu':
            return (
              <MetricCard
                title="CPU Usage"
                value={summary.totalCPU.used}
                limit={summary.totalCPU.limit}
                unit={summary.totalCPU.unit}
                trend={cpuTrend}
              />
            );
          case 'memory':
            return (
              <MetricCard
                title="Memory Usage"
                value={summary.totalMemory.used}
                limit={summary.totalMemory.limit}
                unit={summary.totalMemory.unit}
                trend={memoryTrend}
              />
            );
          case 'storage':
            return (
              <MetricCard
                title="Storage Usage"
                value={summary.totalStorage.used}
                limit={summary.totalStorage.limit}
                unit={summary.totalStorage.unit}
              />
            );
          case 'deployments':
            return (
              <MetricCard
                title="Active Deployments"
                value={summary.activeDeployments}
                subtitle={`across ${summary.totalProviders} providers`}
              />
            );
          default:
            return <WidgetPlaceholder title={widget.title} />;
        }
      }

      case 'time-series-chart': {
        const deplMetrics = deploymentMetrics[0];
        if (!deplMetrics) return <WidgetPlaceholder title={widget.title} />;

        const metricKey = widget.config.metric as keyof typeof deplMetrics.history;
        const series = deplMetrics.history[metricKey];
        const range = (widget.config.timeRange ?? timeRange) as unknown as TimeRange;
        return (
          <TimeSeriesChart
            title={widget.title}
            series={series}
            granularity={granularityForRange(range)}
          />
        );
      }

      case 'alert-list':
        return <AlertsPanel />;

      case 'table':
        return <ProviderBreakdown providers={summary?.byProvider ?? []} />;

      case 'gauge':
      case 'heatmap':
      default:
        return <WidgetPlaceholder title={widget.title} />;
    }
  }

  if (isEditing) {
    return (
      <div className="relative">
        <Button
          size="icon-sm"
          variant="destructive"
          className="absolute -right-2 -top-2 z-10"
          onClick={onRemove}
        >
          <X className="h-3 w-3" />
        </Button>
        {renderWidget()}
      </div>
    );
  }

  return renderWidget();
}

function WidgetPlaceholder({ title }: { title: string }) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex h-24 items-center justify-center text-sm text-muted-foreground">
          Widget: {title}
        </div>
      </CardContent>
    </Card>
  );
}
