/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Metrics Zustand store with mock data.
 * Provides aggregated metrics, deployment metrics, and time-series history.
 */

import { create } from 'zustand';
import { generateId } from '@/lib/utils';
import type {
  Alert,
  AlertEvent,
  DeploymentMetrics,
  MetricPoint,
  MetricSeries,
  MetricsSummary,
  MetricTrend,
  ProviderMetrics,
  ResourceMetric,
  TimeRange,
} from '@virtengine/portal/types/metrics';

// =============================================================================
// Mock Data Generators
// =============================================================================

function mockResourceMetric(used: number, limit: number, unit: string): ResourceMetric {
  return { used, limit, percent: limit > 0 ? (used / limit) * 100 : 0, unit };
}

function generateTimeSeries(
  name: string,
  unit: string,
  points: number,
  baseValue: number,
  variance: number,
  intervalMs: number
): MetricSeries {
  const now = Date.now();
  const data: MetricPoint[] = [];
  for (let i = points - 1; i >= 0; i--) {
    data.push({
      timestamp: now - i * intervalMs,
      value: Math.max(0, baseValue + (Math.random() - 0.5) * variance * 2),
    });
  }
  return { name, unit, data };
}

const MOCK_PROVIDERS: ProviderMetrics[] = [
  {
    providerName: 'CloudCore',
    providerAddress: 'virtengine1prov1abc...7h3k',
    deploymentCount: 3,
    cpu: mockResourceMetric(48, 64, 'cores'),
    memory: mockResourceMetric(192, 256, 'GB'),
    storage: mockResourceMetric(3.2, 5, 'TB'),
  },
  {
    providerName: 'DataNexus',
    providerAddress: 'virtengine1prov2def...9j4m',
    deploymentCount: 2,
    cpu: mockResourceMetric(24, 32, 'cores'),
    memory: mockResourceMetric(96, 128, 'GB'),
    storage: mockResourceMetric(1.8, 3, 'TB'),
  },
  {
    providerName: 'HPCGrid',
    providerAddress: 'virtengine1prov3ghi...2k7n',
    deploymentCount: 1,
    cpu: mockResourceMetric(14, 16, 'cores'),
    memory: mockResourceMetric(52, 64, 'GB'),
    storage: mockResourceMetric(0.8, 1, 'TB'),
  },
];

function buildMockDeploymentMetrics(): DeploymentMetrics[] {
  return [
    {
      deploymentId: 'dep-001',
      provider: 'CloudCore',
      current: {
        timestamp: Date.now(),
        cpu: mockResourceMetric(14, 32, 'cores'),
        memory: mockResourceMetric(58, 128, 'GB'),
        storage: mockResourceMetric(680, 1000, 'GB'),
        network: {
          rxBytesPerSec: 45000,
          txBytesPerSec: 32000,
          rxPacketsPerSec: 1200,
          txPacketsPerSec: 980,
        },
        gpu: {
          utilizationPercent: 72,
          memoryUsedMB: 28000,
          memoryTotalMB: 40960,
          temperatureC: 68,
        },
      },
      history: {
        cpu: generateTimeSeries('CPU', 'cores', 24, 14, 4, 3600000),
        memory: generateTimeSeries('Memory', 'GB', 24, 58, 10, 3600000),
        storage: generateTimeSeries('Storage', 'GB', 24, 680, 20, 3600000),
        network: generateTimeSeries('Network', 'KB/s', 24, 45, 15, 3600000),
      },
      services: [
        {
          name: 'api-gateway',
          replicas: 2,
          current: {
            timestamp: Date.now(),
            cpu: mockResourceMetric(4, 8, 'cores'),
            memory: mockResourceMetric(12, 32, 'GB'),
            storage: mockResourceMetric(20, 50, 'GB'),
            network: {
              rxBytesPerSec: 20000,
              txBytesPerSec: 15000,
              rxPacketsPerSec: 600,
              txPacketsPerSec: 500,
            },
          },
        },
        {
          name: 'inference-worker',
          replicas: 6,
          current: {
            timestamp: Date.now(),
            cpu: mockResourceMetric(10, 24, 'cores'),
            memory: mockResourceMetric(46, 96, 'GB'),
            storage: mockResourceMetric(660, 950, 'GB'),
            network: {
              rxBytesPerSec: 25000,
              txBytesPerSec: 17000,
              rxPacketsPerSec: 600,
              txPacketsPerSec: 480,
            },
            gpu: {
              utilizationPercent: 72,
              memoryUsedMB: 28000,
              memoryTotalMB: 40960,
              temperatureC: 68,
            },
          },
        },
      ],
    },
    {
      deploymentId: 'dep-002',
      provider: 'DataNexus',
      current: {
        timestamp: Date.now(),
        cpu: mockResourceMetric(38, 64, 'cores'),
        memory: mockResourceMetric(148, 256, 'GB'),
        storage: mockResourceMetric(1200, 2000, 'GB'),
        network: {
          rxBytesPerSec: 85000,
          txBytesPerSec: 62000,
          rxPacketsPerSec: 2400,
          txPacketsPerSec: 1800,
        },
      },
      history: {
        cpu: generateTimeSeries('CPU', 'cores', 24, 38, 8, 3600000),
        memory: generateTimeSeries('Memory', 'GB', 24, 148, 20, 3600000),
        storage: generateTimeSeries('Storage', 'GB', 24, 1200, 50, 3600000),
        network: generateTimeSeries('Network', 'KB/s', 24, 85, 25, 3600000),
      },
      services: [
        {
          name: 'data-pipeline',
          replicas: 4,
          current: {
            timestamp: Date.now(),
            cpu: mockResourceMetric(38, 64, 'cores'),
            memory: mockResourceMetric(148, 256, 'GB'),
            storage: mockResourceMetric(1200, 2000, 'GB'),
            network: {
              rxBytesPerSec: 85000,
              txBytesPerSec: 62000,
              rxPacketsPerSec: 2400,
              txPacketsPerSec: 1800,
            },
          },
        },
      ],
    },
    {
      deploymentId: 'dep-003',
      provider: 'HPCGrid',
      current: {
        timestamp: Date.now(),
        cpu: mockResourceMetric(14, 16, 'cores'),
        memory: mockResourceMetric(52, 64, 'GB'),
        storage: mockResourceMetric(380, 500, 'GB'),
        network: {
          rxBytesPerSec: 12000,
          txBytesPerSec: 8000,
          rxPacketsPerSec: 400,
          txPacketsPerSec: 300,
        },
      },
      history: {
        cpu: generateTimeSeries('CPU', 'cores', 24, 14, 3, 3600000),
        memory: generateTimeSeries('Memory', 'GB', 24, 52, 8, 3600000),
        storage: generateTimeSeries('Storage', 'GB', 24, 380, 15, 3600000),
        network: generateTimeSeries('Network', 'KB/s', 24, 12, 5, 3600000),
      },
      services: [
        {
          name: 'batch-node',
          replicas: 1,
          current: {
            timestamp: Date.now(),
            cpu: mockResourceMetric(14, 16, 'cores'),
            memory: mockResourceMetric(52, 64, 'GB'),
            storage: mockResourceMetric(380, 500, 'GB'),
            network: {
              rxBytesPerSec: 12000,
              txBytesPerSec: 8000,
              rxPacketsPerSec: 400,
              txPacketsPerSec: 300,
            },
          },
        },
      ],
    },
  ];
}

function buildMockSummary(deployments: DeploymentMetrics[]): MetricsSummary {
  const totalCPU = deployments.reduce((acc, d) => acc + d.current.cpu.used, 0);
  const totalCPULimit = deployments.reduce((acc, d) => acc + d.current.cpu.limit, 0);
  const totalMem = deployments.reduce((acc, d) => acc + d.current.memory.used, 0);
  const totalMemLimit = deployments.reduce((acc, d) => acc + d.current.memory.limit, 0);
  const totalStor = deployments.reduce((acc, d) => acc + d.current.storage.used, 0);
  const totalStorLimit = deployments.reduce((acc, d) => acc + d.current.storage.limit, 0);

  return {
    totalCPU: mockResourceMetric(totalCPU, totalCPULimit, 'cores'),
    totalMemory: mockResourceMetric(totalMem, totalMemLimit, 'GB'),
    totalStorage: mockResourceMetric(totalStor, totalStorLimit, 'GB'),
    totalNetwork: {
      rxBytesPerSec: deployments.reduce((s, d) => s + d.current.network.rxBytesPerSec, 0),
      txBytesPerSec: deployments.reduce((s, d) => s + d.current.network.txBytesPerSec, 0),
      rxPacketsPerSec: deployments.reduce((s, d) => s + d.current.network.rxPacketsPerSec, 0),
      txPacketsPerSec: deployments.reduce((s, d) => s + d.current.network.txPacketsPerSec, 0),
    },
    activeDeployments: deployments.length,
    totalProviders: MOCK_PROVIDERS.length,
    byProvider: MOCK_PROVIDERS,
    byDeployment: deployments,
    healthOverview: { healthy: 2, degraded: 1, warning: 0, critical: 0 },
  };
}

const MOCK_ALERTS: Alert[] = [
  {
    id: 'alert-001',
    name: 'High CPU Usage',
    deploymentId: 'dep-002',
    metric: 'cpu',
    condition: 'gt',
    threshold: 80,
    duration: 300,
    status: 'firing',
    lastFired: Date.now() - 120000,
    notificationChannels: ['email', 'webhook'],
  },
  {
    id: 'alert-002',
    name: 'Storage Nearly Full',
    deploymentId: 'dep-001',
    metric: 'storage',
    condition: 'gt',
    threshold: 90,
    duration: 600,
    status: 'active',
    notificationChannels: ['email'],
  },
  {
    id: 'alert-003',
    name: 'Memory Pressure',
    metric: 'memory',
    condition: 'gt',
    threshold: 85,
    duration: 300,
    status: 'resolved',
    lastFired: Date.now() - 3600000,
    notificationChannels: ['email', 'webhook'],
  },
];

const MOCK_ALERT_EVENTS: AlertEvent[] = [
  {
    id: 'evt-001',
    alertId: 'alert-001',
    alertName: 'High CPU Usage',
    status: 'firing',
    value: 82.5,
    timestamp: Date.now() - 120000,
    acknowledged: false,
  },
  {
    id: 'evt-002',
    alertId: 'alert-003',
    alertName: 'Memory Pressure',
    status: 'resolved',
    value: 72.1,
    timestamp: Date.now() - 3600000,
    acknowledged: true,
    acknowledgedBy: 'admin',
    acknowledgedAt: Date.now() - 3500000,
  },
  {
    id: 'evt-003',
    alertId: 'alert-003',
    alertName: 'Memory Pressure',
    status: 'firing',
    value: 88.3,
    timestamp: Date.now() - 7200000,
    acknowledged: true,
    acknowledgedBy: 'admin',
    acknowledgedAt: Date.now() - 7100000,
  },
];

// =============================================================================
// Store Interface
// =============================================================================

export interface MetricsState {
  summary: MetricsSummary | null;
  deploymentMetrics: DeploymentMetrics[];
  alerts: Alert[];
  alertEvents: AlertEvent[];
  selectedTimeRange: TimeRange;
  selectedDeploymentId: string | null;
  isLoading: boolean;
  isStreaming: boolean;
  error: string | null;
}

export interface MetricsActions {
  fetchMetrics: () => Promise<void>;
  setTimeRange: (range: TimeRange) => void;
  selectDeployment: (id: string | null) => void;
  toggleStreaming: () => void;
  createAlert: (alert: Omit<Alert, 'id' | 'status' | 'lastFired'>) => void;
  deleteAlert: (id: string) => void;
  acknowledgeAlertEvent: (eventId: string) => void;
  clearError: () => void;
}

export type MetricsStore = MetricsState & MetricsActions;

// =============================================================================
// Store Implementation
// =============================================================================

const initialState: MetricsState = {
  summary: null,
  deploymentMetrics: [],
  alerts: [],
  alertEvents: [],
  selectedTimeRange: '24h',
  selectedDeploymentId: null,
  isLoading: false,
  isStreaming: false,
  error: null,
};

export const useMetricsStore = create<MetricsStore>()((set, _get) => ({
  ...initialState,

  fetchMetrics: async () => {
    set({ isLoading: true, error: null });

    try {
      // In production, fetches from provider APIs and aggregates:
      // const metrics = await multiProviderClient.getAggregatedMetrics();

      await new Promise((resolve) => setTimeout(resolve, 500));

      const deployments = buildMockDeploymentMetrics();
      const summary = buildMockSummary(deployments);

      set({
        summary,
        deploymentMetrics: deployments,
        alerts: MOCK_ALERTS,
        alertEvents: MOCK_ALERT_EVENTS,
        isLoading: false,
      });
    } catch (error) {
      set({
        isLoading: false,
        error: error instanceof Error ? error.message : 'Failed to load metrics',
      });
    }
  },

  setTimeRange: (range) => {
    set({ selectedTimeRange: range });
  },

  selectDeployment: (id) => {
    set({ selectedDeploymentId: id });
  },

  toggleStreaming: () => {
    set((state) => ({ isStreaming: !state.isStreaming }));
  },

  createAlert: (alertData) => {
    const newAlert: Alert = {
      ...alertData,
      id: generateId('alert'),
      status: 'active',
    };
    set((state) => ({
      alerts: [...state.alerts, newAlert],
    }));
  },

  deleteAlert: (id) => {
    set((state) => ({
      alerts: state.alerts.filter((a) => a.id !== id),
      alertEvents: state.alertEvents.filter((e) => e.alertId !== id),
    }));
  },

  acknowledgeAlertEvent: (eventId) => {
    set((state) => ({
      alertEvents: state.alertEvents.map((e) =>
        e.id === eventId
          ? { ...e, acknowledged: true, acknowledgedBy: 'user', acknowledgedAt: Date.now() }
          : e
      ),
    }));
  },

  clearError: () => {
    set({ error: null });
  },
}));

// =============================================================================
// Selectors
// =============================================================================

export const selectFiringAlerts = (state: MetricsStore): Alert[] =>
  state.alerts.filter((a) => a.status === 'firing');

export const selectActiveAlerts = (state: MetricsStore): Alert[] =>
  state.alerts.filter((a) => a.status !== 'resolved');

export const selectRecentAlertEvents = (state: MetricsStore): AlertEvent[] =>
  [...state.alertEvents].sort((a, b) => b.timestamp - a.timestamp).slice(0, 10);

export const selectSelectedDeploymentMetrics = (
  state: MetricsStore
): DeploymentMetrics | undefined =>
  state.selectedDeploymentId
    ? state.deploymentMetrics.find((d) => d.deploymentId === state.selectedDeploymentId)
    : undefined;

export const selectCPUTrend = (state: MetricsStore): MetricTrend => {
  const deployment = state.deploymentMetrics[0];
  if (!deployment) return { direction: 'stable', percent: 0 };
  const data = deployment.history.cpu.data;
  if (data.length < 2) return { direction: 'stable', percent: 0 };
  const recentPt = data[data.length - 1];
  const prevPt = data[Math.floor(data.length / 2)];
  if (!recentPt || !prevPt) return { direction: 'stable', percent: 0 };
  return computeTrend(recentPt.value, prevPt.value);
};

export const selectMemoryTrend = (state: MetricsStore): MetricTrend => {
  const deployment = state.deploymentMetrics[0];
  if (!deployment) return { direction: 'stable', percent: 0 };
  const data = deployment.history.memory.data;
  if (data.length < 2) return { direction: 'stable', percent: 0 };
  const recentPt = data[data.length - 1];
  const prevPt = data[Math.floor(data.length / 2)];
  if (!recentPt || !prevPt) return { direction: 'stable', percent: 0 };
  return computeTrend(recentPt.value, prevPt.value);
};

function computeTrend(current: number, previous: number): MetricTrend {
  if (previous === 0) return { direction: 'stable', percent: 0 };
  const diff = ((current - previous) / previous) * 100;
  const absDiff = Math.abs(diff);
  if (absDiff < 1) return { direction: 'stable', percent: 0 };
  return {
    direction: diff > 0 ? 'up' : 'down',
    percent: Math.round(absDiff * 10) / 10,
  };
}
