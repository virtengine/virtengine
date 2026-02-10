/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Metrics and dashboard aggregation types.
 * Used by metrics hooks, stores, and dashboard components.
 */

// =============================================================================
// Metric Primitives
// =============================================================================

export interface MetricPoint {
  timestamp: number;
  value: number;
}

export interface MetricSeries {
  name: string;
  unit: string;
  data: MetricPoint[];
}

// =============================================================================
// Resource Metrics
// =============================================================================

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

// =============================================================================
// Aggregated Metrics
// =============================================================================

export interface AggregatedMetrics {
  timestamp: number;
  cpu: ResourceMetric;
  memory: ResourceMetric;
  storage: ResourceMetric;
  network: NetworkMetric;
  gpu?: GPUMetric;
}

// =============================================================================
// Deployment Metrics
// =============================================================================

export interface ServiceMetrics {
  name: string;
  replicas: number;
  current: AggregatedMetrics;
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

// =============================================================================
// Dashboard Summary
// =============================================================================

export interface MetricsSummary {
  totalCPU: ResourceMetric;
  totalMemory: ResourceMetric;
  totalStorage: ResourceMetric;
  totalNetwork: NetworkMetric;
  activeDeployments: number;
  totalProviders: number;
  byProvider: ProviderMetrics[];
  byDeployment: DeploymentMetrics[];
  healthOverview: HealthOverview;
}

export interface ProviderMetrics {
  providerName: string;
  providerAddress: string;
  deploymentCount: number;
  cpu: ResourceMetric;
  memory: ResourceMetric;
  storage: ResourceMetric;
}

export interface HealthOverview {
  healthy: number;
  degraded: number;
  warning: number;
  critical: number;
}

// =============================================================================
// Alerts
// =============================================================================

export type AlertMetric = "cpu" | "memory" | "storage" | "network";
export type AlertCondition = "gt" | "lt" | "eq";
export type AlertStatus = "active" | "firing" | "resolved";

export interface Alert {
  id: string;
  name: string;
  deploymentId?: string;
  metric: AlertMetric;
  condition: AlertCondition;
  threshold: number;
  duration: number;
  status: AlertStatus;
  lastFired?: number;
  notificationChannels: string[];
}

export interface AlertEvent {
  id: string;
  alertId: string;
  alertName: string;
  status: "firing" | "resolved";
  value: number;
  timestamp: number;
  acknowledged: boolean;
  acknowledgedBy?: string;
  acknowledgedAt?: number;
}

// =============================================================================
// Dashboard Configuration
// =============================================================================

export type WidgetType =
  | "metric-card"
  | "time-series-chart"
  | "gauge"
  | "table"
  | "heatmap"
  | "alert-list";

export interface WidgetConfig {
  metric?: string;
  deploymentId?: string;
  timeRange?: TimeRange;
  refreshInterval?: number;
}

export interface WidgetPosition {
  x: number;
  y: number;
  w: number;
  h: number;
}

export interface DashboardWidget {
  id: string;
  type: WidgetType;
  title: string;
  config: WidgetConfig;
  position: WidgetPosition;
}

export interface DashboardConfig {
  id: string;
  name: string;
  isDefault: boolean;
  layout: DashboardWidget[];
  createdAt: number;
  updatedAt: number;
}

// =============================================================================
// Time Ranges & Granularity
// =============================================================================

export type TimeRange = "1h" | "6h" | "24h" | "7d" | "30d";
export type Granularity = "minute" | "hour" | "day";

export const TIME_RANGE_MS: Record<TimeRange, number> = {
  "1h": 60 * 60 * 1000,
  "6h": 6 * 60 * 60 * 1000,
  "24h": 24 * 60 * 60 * 1000,
  "7d": 7 * 24 * 60 * 60 * 1000,
  "30d": 30 * 24 * 60 * 60 * 1000,
};

export const TIME_RANGE_LABELS: Record<TimeRange, string> = {
  "1h": "Last hour",
  "6h": "Last 6 hours",
  "24h": "Last 24 hours",
  "7d": "Last 7 days",
  "30d": "Last 30 days",
};

export function granularityForRange(range: TimeRange): Granularity {
  switch (range) {
    case "1h":
      return "minute";
    case "6h":
    case "24h":
      return "hour";
    default:
      return "day";
  }
}

// =============================================================================
// Trend Helpers
// =============================================================================

export interface MetricTrend {
  direction: "up" | "down" | "stable";
  percent: number;
}

export function computeTrend(current: number, previous: number): MetricTrend {
  if (previous === 0) return { direction: "stable", percent: 0 };
  const diff = ((current - previous) / previous) * 100;
  const absDiff = Math.abs(diff);
  if (absDiff < 1) return { direction: "stable", percent: 0 };
  return {
    direction: diff > 0 ? "up" : "down",
    percent: Math.round(absDiff * 10) / 10,
  };
}

// =============================================================================
// Formatting Helpers
// =============================================================================

export function formatMetricValue(value: number, unit: string): string {
  if (unit === "bytes" || unit === "B") {
    if (value >= 1e9) return `${(value / 1e9).toFixed(1)} GB`;
    if (value >= 1e6) return `${(value / 1e6).toFixed(1)} MB`;
    if (value >= 1e3) return `${(value / 1e3).toFixed(1)} KB`;
    return `${value} B`;
  }
  if (unit === "percent" || unit === "%") {
    return `${value.toFixed(1)}%`;
  }
  return `${value.toFixed(1)} ${unit}`;
}

export function formatTimestamp(ts: number, granularity: Granularity): string {
  const d = new Date(ts);
  switch (granularity) {
    case "minute":
      return d.toLocaleTimeString(undefined, {
        hour: "2-digit",
        minute: "2-digit",
      });
    case "hour":
      return d.toLocaleTimeString(undefined, {
        hour: "2-digit",
        minute: "2-digit",
      });
    case "day":
      return d.toLocaleDateString(undefined, {
        month: "short",
        day: "numeric",
      });
  }
}

// =============================================================================
// Alert Status Variants
// =============================================================================

export const ALERT_STATUS_VARIANT: Record<
  AlertStatus,
  "default" | "destructive" | "success" | "secondary"
> = {
  active: "secondary",
  firing: "destructive",
  resolved: "success",
};
