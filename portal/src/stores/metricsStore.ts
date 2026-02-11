/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Metrics store backed by provider daemon aggregation.
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
import { MultiProviderClient } from '@/lib/portal-adapter';
import { getPortalEndpoints } from '@/lib/config';

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

let providerClient: MultiProviderClient | null = null;
let providerClientInit: Promise<void> | null = null;

const getProviderClient = async () => {
  if (!providerClient) {
    providerClient = new MultiProviderClient({
      chainEndpoint: getPortalEndpoints().chainRest,
    });
  }
  if (!providerClientInit) {
    providerClientInit = providerClient.initialize().catch(() => undefined);
  }
  await providerClientInit;
  return providerClient;
};

const buildResourceMetric = (used: number, limit: number, unit: string): ResourceMetric => ({
  used,
  limit,
  percent: limit > 0 ? (used / limit) * 100 : 0,
  unit,
});

const seriesFromPoint = (name: string, unit: string, point: MetricPoint): MetricSeries => ({
  name,
  unit,
  data: [point],
});

const buildDeploymentMetrics = (
  deploymentId: string,
  providerName: string,
  metrics: {
    cpu: { usage: number; limit: number };
    memory: { usage: number; limit: number };
    storage: { usage: number; limit: number };
    network?: { rxBytes?: number; txBytes?: number };
    gpu?: { usage?: number; limit?: number };
  }
): DeploymentMetrics => {
  const timestamp = Date.now();
  return {
    deploymentId,
    provider: providerName,
    current: {
      timestamp,
      cpu: buildResourceMetric(metrics.cpu.usage, metrics.cpu.limit, 'cores'),
      memory: buildResourceMetric(metrics.memory.usage, metrics.memory.limit, 'GB'),
      storage: buildResourceMetric(metrics.storage.usage, metrics.storage.limit, 'GB'),
      network: {
        rxBytesPerSec: metrics.network?.rxBytes ?? 0,
        txBytesPerSec: metrics.network?.txBytes ?? 0,
        rxPacketsPerSec: 0,
        txPacketsPerSec: 0,
      },
      gpu: metrics.gpu
        ? {
            utilizationPercent: metrics.gpu.limit
              ? ((metrics.gpu.usage ?? 0) / metrics.gpu.limit) * 100
              : 0,
            memoryUsedMB: metrics.gpu.usage ?? 0,
            memoryTotalMB: metrics.gpu.limit ?? 0,
            temperatureC: 0,
          }
        : undefined,
    },
    history: {
      cpu: seriesFromPoint('CPU', 'cores', { timestamp, value: metrics.cpu.usage }),
      memory: seriesFromPoint('Memory', 'GB', { timestamp, value: metrics.memory.usage }),
      storage: seriesFromPoint('Storage', 'GB', { timestamp, value: metrics.storage.usage }),
      network: seriesFromPoint('Network', 'bytes/s', {
        timestamp,
        value: (metrics.network?.rxBytes ?? 0) + (metrics.network?.txBytes ?? 0),
      }),
    },
    services: [],
  };
};

export const useMetricsStore = create<MetricsStore>()((set) => ({
  ...initialState,

  fetchMetrics: async () => {
    set({ isLoading: true, error: null });

    try {
      const client = await getProviderClient();
      const providers = client.getProviders();
      const deployments = await client.listAllDeployments({ refresh: true });

      const metricsResults = await Promise.allSettled(
        deployments.map(async (deployment) => {
          const provider = providers.find((p) => p.address === deployment.providerId);
          const daemonClient = client.getClient(deployment.providerId);
          if (!daemonClient) return null;
          const metrics = await daemonClient.getDeploymentMetrics(deployment.id);
          return {
            deploymentId: deployment.id,
            providerName: provider?.name ?? deployment.providerId,
            providerId: deployment.providerId,
            metrics,
          };
        })
      );

      const deploymentMetrics: DeploymentMetrics[] = [];
      const providerMetricsMap = new Map<string, ProviderMetrics>();
      let totalCpuUsed = 0;
      let totalCpuLimit = 0;
      let totalMemUsed = 0;
      let totalMemLimit = 0;
      let totalStorageUsed = 0;
      let totalStorageLimit = 0;
      let totalNetRx = 0;
      let totalNetTx = 0;

      metricsResults.forEach((result) => {
        if (result.status !== 'fulfilled' || !result.value) return;
        const { deploymentId, providerName, providerId, metrics } = result.value;
        deploymentMetrics.push(
          buildDeploymentMetrics(deploymentId, providerName, {
            cpu: metrics.cpu,
            memory: metrics.memory,
            storage: metrics.storage,
            network: metrics.network,
            gpu: metrics.gpu,
          })
        );

        totalCpuUsed += metrics.cpu.usage ?? 0;
        totalCpuLimit += metrics.cpu.limit ?? 0;
        totalMemUsed += metrics.memory.usage ?? 0;
        totalMemLimit += metrics.memory.limit ?? 0;
        totalStorageUsed += metrics.storage.usage ?? 0;
        totalStorageLimit += metrics.storage.limit ?? 0;
        totalNetRx += metrics.network?.rxBytes ?? 0;
        totalNetTx += metrics.network?.txBytes ?? 0;

        const existing = providerMetricsMap.get(providerId);
        if (existing) {
          existing.cpu.used += metrics.cpu.usage ?? 0;
          existing.cpu.limit += metrics.cpu.limit ?? 0;
          existing.memory.used += metrics.memory.usage ?? 0;
          existing.memory.limit += metrics.memory.limit ?? 0;
          existing.storage.used += metrics.storage.usage ?? 0;
          existing.storage.limit += metrics.storage.limit ?? 0;
          existing.deploymentCount += 1;
        } else {
          providerMetricsMap.set(providerId, {
            providerName,
            providerAddress: providerId,
            deploymentCount: 1,
            cpu: buildResourceMetric(metrics.cpu.usage ?? 0, metrics.cpu.limit ?? 0, 'cores'),
            memory: buildResourceMetric(metrics.memory.usage ?? 0, metrics.memory.limit ?? 0, 'GB'),
            storage: buildResourceMetric(
              metrics.storage.usage ?? 0,
              metrics.storage.limit ?? 0,
              'GB'
            ),
          });
        }
      });

      const summary: MetricsSummary = {
        totalCPU: buildResourceMetric(totalCpuUsed, totalCpuLimit, 'cores'),
        totalMemory: buildResourceMetric(totalMemUsed, totalMemLimit, 'GB'),
        totalStorage: buildResourceMetric(totalStorageUsed, totalStorageLimit, 'GB'),
        totalNetwork: {
          rxBytesPerSec: totalNetRx,
          txBytesPerSec: totalNetTx,
          rxPacketsPerSec: 0,
          txPacketsPerSec: 0,
        },
        activeDeployments: deploymentMetrics.length,
        totalProviders: providerMetricsMap.size,
        byProvider: Array.from(providerMetricsMap.values()),
        byDeployment: deploymentMetrics,
        healthOverview: {
          healthy: deploymentMetrics.length,
          degraded: 0,
          warning: 0,
          critical: 0,
        },
      };

      set({
        summary,
        deploymentMetrics,
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
