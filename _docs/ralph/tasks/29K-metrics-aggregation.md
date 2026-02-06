# Task 29K: Metrics/Dashboard Aggregation

**ID:** 29K  
**Title:** feat(portal): Metrics/dashboard aggregation  
**Priority:** P2 (Medium)  
**Wave:** 4 (After 29G)  
**Estimated LOC:** ~2500  
**Dependencies:** 29G (Multi-provider client)  
**Blocking:** None  

---

## Problem Statement

Users with multiple deployments across multiple providers need a unified metrics dashboard:

1. **Aggregated metrics** - Combined view across all deployments
2. **Historical data** - Time-series charts for trends
3. **Real-time updates** - Live metric streaming
4. **Alerting** - Threshold-based notifications
5. **Custom dashboards** - User-configurable views

---

## Acceptance Criteria

### AC-1: Unified Metrics Dashboard
- [ ] Total resource usage across all deployments
- [ ] Usage breakdown by provider
- [ ] Usage breakdown by deployment
- [ ] Current vs allocated comparison
- [ ] Health status overview

### AC-2: Time-Series Charts
- [ ] CPU usage over time
- [ ] Memory usage over time
- [ ] Storage usage over time
- [ ] Network I/O over time
- [ ] Customizable time ranges

### AC-3: Real-Time Streaming
- [ ] WebSocket for live metrics
- [ ] Auto-updating charts
- [ ] Connection status indicator
- [ ] Graceful reconnection

### AC-4: Alert Configuration
- [ ] Set threshold alerts (CPU > 80%)
- [ ] Alert notification channels
- [ ] Alert history view
- [ ] Alert acknowledgment

### AC-5: Custom Dashboards
- [ ] Create custom dashboard layouts
- [ ] Add/remove widgets
- [ ] Resize and reorder widgets
- [ ] Save dashboard configurations

### AC-6: Deployment Detail Metrics
- [ ] Per-deployment deep dive
- [ ] Container-level metrics
- [ ] Service-level breakdown
- [ ] Log correlation

---

## Technical Requirements

### Metrics Types

```typescript
// lib/portal/src/types/metrics.ts

export interface MetricPoint {
  timestamp: Date;
  value: number;
}

export interface MetricSeries {
  name: string;
  unit: string;
  data: MetricPoint[];
}

export interface AggregatedMetrics {
  timestamp: Date;
  cpu: ResourceMetric;
  memory: ResourceMetric;
  storage: ResourceMetric;
  network: NetworkMetric;
  gpu?: GPUMetric;
}

export interface ResourceMetric {
  used: number;
  limit: number;
  percent: number;
  unit: string;
}

export interface NetworkMetric {
  rxBytesPerSec: number;
  txBytesPerSec: number;
  rxPacketsPerSec: number;
  txPacketsPerSec: number;
}

export interface GPUMetric {
  utilizationPercent: number;
  memoryUsedMB: number;
  memoryTotalMB: number;
  temperatureC: number;
}

export interface DeploymentMetrics {
  deploymentId: string;
  provider: string;
  current: AggregatedMetrics;
  history: {
    cpu: MetricSeries;
    memory: MetricSeries;
    storage: MetricSeries;
    network: MetricSeries;
  };
  services: ServiceMetrics[];
}

export interface ServiceMetrics {
  name: string;
  replicas: number;
  current: AggregatedMetrics;
}

export interface Alert {
  id: string;
  name: string;
  deploymentId?: string;
  metric: 'cpu' | 'memory' | 'storage' | 'network';
  condition: 'gt' | 'lt' | 'eq';
  threshold: number;
  duration: number; // seconds
  status: 'active' | 'firing' | 'resolved';
  lastFired?: Date;
  notificationChannels: string[];
}

export interface AlertEvent {
  id: string;
  alertId: string;
  status: 'firing' | 'resolved';
  value: number;
  timestamp: Date;
  acknowledged?: boolean;
  acknowledgedBy?: string;
  acknowledgedAt?: Date;
}

export interface DashboardConfig {
  id: string;
  name: string;
  isDefault: boolean;
  layout: DashboardWidget[];
  createdAt: Date;
  updatedAt: Date;
}

export interface DashboardWidget {
  id: string;
  type: WidgetType;
  title: string;
  config: WidgetConfig;
  position: { x: number; y: number; w: number; h: number };
}

export type WidgetType = 
  | 'metric-card'
  | 'time-series-chart'
  | 'gauge'
  | 'table'
  | 'heatmap'
  | 'alert-list';

export interface WidgetConfig {
  metric?: string;
  deploymentId?: string;
  timeRange?: string;
  refreshInterval?: number;
  [key: string]: any;
}
```

### Metrics Hooks

```typescript
// lib/portal/src/hooks/useMetrics.ts

import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useState, useCallback } from 'react';
import { useMultiProvider } from '../multi-provider/context';

export function useAggregatedMetrics() {
  const { client } = useMultiProvider();

  return useQuery({
    queryKey: ['aggregated-metrics'],
    queryFn: () => client!.getAggregatedMetrics(),
    enabled: !!client,
    refetchInterval: 30_000,
  });
}

export function useMetricsHistory(options: {
  deploymentId?: string;
  metric: 'cpu' | 'memory' | 'storage' | 'network';
  timeRange: '1h' | '6h' | '24h' | '7d' | '30d';
  granularity: 'minute' | 'hour' | 'day';
}) {
  const { client } = useMultiProvider();

  return useQuery({
    queryKey: ['metrics-history', options],
    queryFn: async () => {
      const now = new Date();
      const ranges = {
        '1h': 60 * 60 * 1000,
        '6h': 6 * 60 * 60 * 1000,
        '24h': 24 * 60 * 60 * 1000,
        '7d': 7 * 24 * 60 * 60 * 1000,
        '30d': 30 * 24 * 60 * 60 * 1000,
      };
      
      const start = new Date(now.getTime() - ranges[options.timeRange]);
      
      if (options.deploymentId) {
        const deployment = await client!.getDeployment(options.deploymentId);
        const providerClient = client!.getClient(deployment!.providerId);
        
        return providerClient!.request<MetricSeries>(
          'GET',
          `/api/v1/deployments/${options.deploymentId}/metrics/history`,
          {
            metric: options.metric,
            start: start.toISOString(),
            end: now.toISOString(),
            granularity: options.granularity,
          }
        );
      }
      
      // Aggregate across all deployments
      const deployments = await client!.listAllDeployments();
      const allSeries: MetricSeries[] = [];
      
      await Promise.allSettled(
        deployments.map(async (d) => {
          const providerClient = client!.getClient(d.providerId);
          if (!providerClient) return;
          
          const series = await providerClient.request<MetricSeries>(
            'GET',
            `/api/v1/deployments/${d.id}/metrics/history`,
            {
              metric: options.metric,
              start: start.toISOString(),
              end: now.toISOString(),
              granularity: options.granularity,
            }
          );
          
          allSeries.push(series);
        })
      );
      
      return aggregateMetricSeries(allSeries);
    },
    enabled: !!client,
  });
}

export function useRealtimeMetrics(deploymentId: string) {
  const { client } = useMultiProvider();
  const queryClient = useQueryClient();
  const [isConnected, setIsConnected] = useState(false);
  const [latestMetrics, setLatestMetrics] = useState<AggregatedMetrics | null>(null);

  useEffect(() => {
    if (!client || !deploymentId) return;

    const connectWebSocket = async () => {
      const deployment = await client.getDeployment(deploymentId);
      if (!deployment) return;

      const providerClient = client.getClient(deployment.providerId);
      if (!providerClient) return;

      const ws = new WebSocket(
        providerClient.buildWebSocketUrl(`/api/v1/deployments/${deploymentId}/metrics/stream`)
      );

      ws.onopen = () => setIsConnected(true);
      ws.onclose = () => setIsConnected(false);
      
      ws.onmessage = (event) => {
        const metrics = JSON.parse(event.data) as AggregatedMetrics;
        metrics.timestamp = new Date(metrics.timestamp);
        setLatestMetrics(metrics);
        
        // Update query cache
        queryClient.setQueryData(
          ['deployment-metrics', deploymentId],
          metrics
        );
      };

      return () => {
        ws.close();
      };
    };

    connectWebSocket();
  }, [client, deploymentId, queryClient]);

  return { isConnected, metrics: latestMetrics };
}
```

### Dashboard Components

```typescript
// lib/portal/src/components/metrics/MetricsDashboard.tsx

import { useAggregatedMetrics } from '../../hooks/useMetrics';
import { MetricCard } from './MetricCard';
import { TimeSeriesChart } from './TimeSeriesChart';
import { ProviderBreakdown } from './ProviderBreakdown';
import { AlertsPanel } from './AlertsPanel';

export function MetricsDashboard() {
  const { data: metrics, isLoading } = useAggregatedMetrics();

  if (isLoading) {
    return <MetricsDashboardSkeleton />;
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">Metrics Dashboard</h1>
        <TimeRangeSelector />
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <MetricCard
          title="CPU Usage"
          value={metrics?.totalCPU.used || 0}
          limit={metrics?.totalCPU.limit || 0}
          unit="cores"
          trend={calculateTrend('cpu')}
        />
        <MetricCard
          title="Memory Usage"
          value={metrics?.totalMemory.used || 0}
          limit={metrics?.totalMemory.limit || 0}
          unit="GB"
          trend={calculateTrend('memory')}
        />
        <MetricCard
          title="Storage Usage"
          value={metrics?.totalStorage.used || 0}
          limit={metrics?.totalStorage.limit || 0}
          unit="GB"
          trend={calculateTrend('storage')}
        />
        <MetricCard
          title="Active Deployments"
          value={metrics?.byProvider.size || 0}
          subtitle="across providers"
        />
      </div>

      {/* Charts */}
      <div className="grid gap-6 lg:grid-cols-2">
        <div className="border rounded-lg p-4">
          <h2 className="font-semibold mb-4">CPU Usage</h2>
          <TimeSeriesChart metric="cpu" />
        </div>
        <div className="border rounded-lg p-4">
          <h2 className="font-semibold mb-4">Memory Usage</h2>
          <TimeSeriesChart metric="memory" />
        </div>
      </div>

      {/* Breakdown */}
      <div className="grid gap-6 lg:grid-cols-2">
        <ProviderBreakdown metrics={metrics} />
        <AlertsPanel />
      </div>
    </div>
  );
}

// lib/portal/src/components/metrics/TimeSeriesChart.tsx

import { useMemo } from 'react';
import { useMetricsHistory } from '../../hooks/useMetrics';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Area,
  ComposedChart,
} from 'recharts';
import { format } from 'date-fns';

interface TimeSeriesChartProps {
  metric: 'cpu' | 'memory' | 'storage' | 'network';
  deploymentId?: string;
  timeRange?: '1h' | '6h' | '24h' | '7d' | '30d';
  height?: number;
}

export function TimeSeriesChart({
  metric,
  deploymentId,
  timeRange = '24h',
  height = 300,
}: TimeSeriesChartProps) {
  const { data: series, isLoading } = useMetricsHistory({
    metric,
    deploymentId,
    timeRange,
    granularity: timeRange === '1h' ? 'minute' : timeRange === '24h' ? 'hour' : 'day',
  });

  const chartData = useMemo(() => {
    if (!series?.data) return [];
    
    return series.data.map((point) => ({
      timestamp: point.timestamp,
      value: point.value,
      formatted: format(point.timestamp, 'HH:mm'),
    }));
  }, [series]);

  if (isLoading) {
    return <ChartSkeleton height={height} />;
  }

  return (
    <ResponsiveContainer width="100%" height={height}>
      <ComposedChart data={chartData}>
        <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
        <XAxis
          dataKey="formatted"
          tick={{ fontSize: 12 }}
          className="text-muted-foreground"
        />
        <YAxis
          tick={{ fontSize: 12 }}
          className="text-muted-foreground"
          domain={[0, 'auto']}
        />
        <Tooltip
          content={({ active, payload }) => {
            if (!active || !payload?.length) return null;
            const data = payload[0].payload;
            return (
              <div className="bg-background border rounded-lg p-2 shadow-lg">
                <p className="text-sm font-medium">
                  {format(data.timestamp, 'MMM d, HH:mm')}
                </p>
                <p className="text-sm text-muted-foreground">
                  {data.value.toFixed(2)} {series?.unit}
                </p>
              </div>
            );
          }}
        />
        <Area
          type="monotone"
          dataKey="value"
          fill="hsl(var(--primary) / 0.2)"
          stroke="hsl(var(--primary))"
          strokeWidth={2}
        />
      </ComposedChart>
    </ResponsiveContainer>
  );
}

// lib/portal/src/components/metrics/MetricCard.tsx

import { TrendingUp, TrendingDown, Minus } from 'lucide-react';
import { cn } from '../../utils/cn';

interface MetricCardProps {
  title: string;
  value: number;
  limit?: number;
  unit?: string;
  subtitle?: string;
  trend?: {
    direction: 'up' | 'down' | 'stable';
    percent: number;
  };
}

export function MetricCard({
  title,
  value,
  limit,
  unit,
  subtitle,
  trend,
}: MetricCardProps) {
  const percent = limit ? (value / limit) * 100 : 0;
  const isWarning = percent > 80;
  const isCritical = percent > 95;

  return (
    <div className="border rounded-lg p-4">
      <div className="flex justify-between items-start">
        <p className="text-sm text-muted-foreground">{title}</p>
        {trend && (
          <div
            className={cn(
              'flex items-center gap-1 text-xs',
              trend.direction === 'up' && 'text-red-500',
              trend.direction === 'down' && 'text-green-500',
              trend.direction === 'stable' && 'text-muted-foreground'
            )}
          >
            {trend.direction === 'up' && <TrendingUp className="h-3 w-3" />}
            {trend.direction === 'down' && <TrendingDown className="h-3 w-3" />}
            {trend.direction === 'stable' && <Minus className="h-3 w-3" />}
            <span>{trend.percent}%</span>
          </div>
        )}
      </div>

      <div className="mt-2">
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
          <span className="text-muted-foreground">
            {' '}
            / {limit} {unit}
          </span>
        )}
        {!limit && unit && (
          <span className="text-muted-foreground"> {unit}</span>
        )}
      </div>

      {subtitle && (
        <p className="text-xs text-muted-foreground mt-1">{subtitle}</p>
      )}

      {limit && (
        <div className="mt-3">
          <div className="h-1.5 bg-muted rounded-full overflow-hidden">
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
          <p className="text-xs text-muted-foreground mt-1">
            {percent.toFixed(1)}% used
          </p>
        </div>
      )}
    </div>
  );
}

// lib/portal/src/components/metrics/AlertsPanel.tsx

import { useAlerts, useAcknowledgeAlert } from '../../hooks/useAlerts';
import { Alert, AlertEvent } from '../../types/metrics';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { Bell, Check } from 'lucide-react';

export function AlertsPanel() {
  const { data: alerts } = useAlerts();
  const acknowledgeAlert = useAcknowledgeAlert();

  const firingAlerts = alerts?.filter((a) => a.status === 'firing') || [];
  const recentEvents = alerts?.flatMap((a) => a.events || [])
    .sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime())
    .slice(0, 5) || [];

  return (
    <div className="border rounded-lg p-4">
      <div className="flex justify-between items-center mb-4">
        <h2 className="font-semibold">Alerts</h2>
        <Badge variant={firingAlerts.length > 0 ? 'destructive' : 'secondary'}>
          {firingAlerts.length} active
        </Badge>
      </div>

      {firingAlerts.length === 0 ? (
        <div className="text-center py-8 text-muted-foreground">
          <Bell className="h-8 w-8 mx-auto mb-2 opacity-50" />
          <p>No active alerts</p>
        </div>
      ) : (
        <div className="space-y-3">
          {firingAlerts.map((alert) => (
            <div
              key={alert.id}
              className="flex items-center justify-between p-3 bg-red-50 dark:bg-red-900/20 rounded-lg"
            >
              <div>
                <p className="font-medium text-red-700 dark:text-red-400">
                  {alert.name}
                </p>
                <p className="text-sm text-red-600 dark:text-red-300">
                  {alert.metric} {alert.condition} {alert.threshold}
                </p>
              </div>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => acknowledgeAlert.mutate(alert.id)}
              >
                <Check className="h-4 w-4" />
              </Button>
            </div>
          ))}
        </div>
      )}

      {recentEvents.length > 0 && (
        <div className="mt-4 pt-4 border-t">
          <h3 className="text-sm font-medium mb-2">Recent Events</h3>
          <div className="space-y-2">
            {recentEvents.map((event) => (
              <div
                key={event.id}
                className="flex items-center justify-between text-sm"
              >
                <span
                  className={cn(
                    event.status === 'firing' ? 'text-red-500' : 'text-green-500'
                  )}
                >
                  {event.status === 'firing' ? '▲' : '▼'} {event.alertName}
                </span>
                <span className="text-muted-foreground">
                  {formatDistanceToNow(event.timestamp)} ago
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
```

---

## Files to Create

| Path | Description | Est. Lines |
|------|-------------|------------|
| `lib/portal/src/types/metrics.ts` | Metrics types | 150 |
| `lib/portal/src/hooks/useMetrics.ts` | Metrics hooks | 200 |
| `lib/portal/src/hooks/useAlerts.ts` | Alert hooks | 150 |
| `lib/portal/src/components/metrics/MetricsDashboard.tsx` | Dashboard | 150 |
| `lib/portal/src/components/metrics/MetricCard.tsx` | Metric card | 100 |
| `lib/portal/src/components/metrics/TimeSeriesChart.tsx` | Chart | 150 |
| `lib/portal/src/components/metrics/ProviderBreakdown.tsx` | Breakdown | 100 |
| `lib/portal/src/components/metrics/AlertsPanel.tsx` | Alerts | 120 |
| `lib/portal/src/components/metrics/AlertConfigDialog.tsx` | Config | 150 |
| `lib/portal/src/components/metrics/CustomDashboard.tsx` | Custom | 200 |
| `lib/portal/src/components/metrics/DashboardWidget.tsx` | Widget | 150 |
| `lib/portal/src/stores/dashboard.ts` | Dashboard store | 80 |
| `portal/src/app/metrics/page.tsx` | Metrics page | 50 |
| `portal/src/app/metrics/alerts/page.tsx` | Alerts page | 50 |

**Total: ~1800 lines**

---

## Vibe-Kanban Task ID

`7f4df190-89dd-4b1f-b51d-c8da7d141bc0`
