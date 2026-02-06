/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Real-time order tracking hook using WebSocket connection.
 * Provides live status updates, usage metrics, and event streaming.
 */

'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import type { OrderStatus } from '@/stores/orderStore';
import type {
  OrderDetail,
  OrderStatusEvent,
  OrderUsageData,
  ResourceAccessInfo,
} from './tracking-types';

// =============================================================================
// Types
// =============================================================================

export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

export interface OrderTrackingEvent {
  type: 'status_change' | 'usage_update' | 'alert' | 'access_ready' | 'error';
  orderId: string;
  timestamp: string;
  payload: unknown;
}

export interface UseOrderTrackingOptions {
  orderId: string;
  enabled?: boolean;
  pollingInterval?: number;
  onStatusChange?: (orderId: string, status: OrderStatus) => void;
  onError?: (error: Error) => void;
}

export interface UseOrderTrackingReturn {
  order: OrderDetail | null;
  connectionStatus: ConnectionStatus;
  lastEvent: OrderTrackingEvent | null;
  isLoading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
  disconnect: () => void;
}

// =============================================================================
// Mock Data Generators
// =============================================================================

function generateMockTimeline(orderId: string, status: OrderStatus) {
  const now = new Date();
  const createdAt = new Date(now.getTime() - 86400000 * 3);
  const events: OrderStatusEvent[] = [
    {
      id: `${orderId}-evt-1`,
      status: 'pending',
      title: 'Order Created',
      description: 'Order submitted to the marketplace',
      timestamp: createdAt.toISOString(),
    },
    {
      id: `${orderId}-evt-2`,
      status: 'matched',
      title: 'Provider Matched',
      description: 'Order matched with provider bid',
      timestamp: new Date(createdAt.getTime() + 300000).toISOString(),
    },
  ];

  if (status !== 'pending' && status !== 'matched') {
    events.push({
      id: `${orderId}-evt-3`,
      status: 'deploying',
      title: 'Deployment Started',
      description: 'Resources are being provisioned',
      timestamp: new Date(createdAt.getTime() + 600000).toISOString(),
    });
  }

  if (
    status === 'running' ||
    status === 'paused' ||
    status === 'stopped' ||
    status === 'completed'
  ) {
    events.push({
      id: `${orderId}-evt-4`,
      status: 'running',
      title: 'Deployment Running',
      description: 'Resources provisioned and accessible',
      timestamp: new Date(createdAt.getTime() + 1800000).toISOString(),
    });
  }

  if (status === 'failed') {
    events.push({
      id: `${orderId}-evt-fail`,
      status: 'failed',
      title: 'Deployment Failed',
      description: 'Provider encountered a provisioning error',
      timestamp: new Date(createdAt.getTime() + 900000).toISOString(),
    });
  }

  if (status === 'stopped' || status === 'completed') {
    events.push({
      id: `${orderId}-evt-5`,
      status,
      title: status === 'completed' ? 'Order Completed' : 'Order Stopped',
      description: status === 'completed' ? 'Lease duration expired' : 'Order stopped by user',
      timestamp: new Date(now.getTime() - 3600000).toISOString(),
    });
  }

  const progressMap: Record<OrderStatus, number> = {
    pending: 10,
    matched: 30,
    deploying: 60,
    running: 100,
    paused: 100,
    stopped: 100,
    completed: 100,
    failed: 0,
  };

  return {
    events,
    currentStatus: status,
    progressPercent: progressMap[status],
    estimatedCompletion:
      status === 'deploying' ? new Date(now.getTime() + 600000).toISOString() : undefined,
  };
}

function generateMockAccess(orderId: string, status: OrderStatus): ResourceAccessInfo {
  const isProvisioned = status === 'running' || status === 'paused';
  return {
    orderId,
    isProvisioned,
    credentials: isProvisioned
      ? [
          {
            type: 'ssh',
            label: 'SSH Access',
            host: '10.42.3.17',
            port: 22,
            username: 'virtuser',
            privateKey:
              '-----BEGIN OPENSSH PRIVATE KEY-----\n[key redacted]\n-----END OPENSSH PRIVATE KEY-----',
          },
          {
            type: 'api',
            label: 'API Endpoint',
            host: 'api.provider.virtengine.io',
            port: 443,
            apiKey: 've_live_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx',
            url: `https://api.provider.virtengine.io/v1/workloads/${orderId}`,
          },
          {
            type: 'console',
            label: 'Web Console',
            host: 'console.provider.virtengine.io',
            url: `https://console.provider.virtengine.io/session/${orderId}`,
          },
        ]
      : [],
    endpoints: isProvisioned
      ? [
          {
            name: 'Workload API',
            url: `https://api.provider.virtengine.io/v1/workloads/${orderId}`,
            method: 'GET',
            description: 'Query workload status and metrics',
          },
          {
            name: 'Logs API',
            url: `https://api.provider.virtengine.io/v1/workloads/${orderId}/logs`,
            method: 'GET',
            description: 'Stream workload logs',
          },
        ]
      : [],
    consoleUrl: isProvisioned
      ? `https://console.provider.virtengine.io/session/${orderId}`
      : undefined,
  };
}

function generateMockUsage(orderId: string, status: OrderStatus): OrderUsageData {
  const now = new Date();
  const isActive = status === 'running' || status === 'paused';

  const generateHistory = (baseValue: number, variance: number) =>
    Array.from({ length: 24 }, (_, i) => ({
      timestamp: new Date(now.getTime() - (23 - i) * 3600000).toISOString(),
      value: isActive ? Math.max(0, baseValue + (Math.random() - 0.5) * variance) : 0,
    }));

  return {
    orderId,
    lastUpdated: now.toISOString(),
    metrics: [
      {
        resourceType: 'cpu',
        label: 'CPU',
        current: isActive ? 42 : 0,
        allocated: 100,
        unit: '%',
        history: generateHistory(42, 20),
      },
      {
        resourceType: 'memory',
        label: 'Memory',
        current: isActive ? 16.4 : 0,
        allocated: 64,
        unit: 'GB',
        history: generateHistory(16.4, 4),
      },
      {
        resourceType: 'storage',
        label: 'Storage',
        current: isActive ? 187 : 0,
        allocated: 500,
        unit: 'GB',
        history: generateHistory(187, 10),
      },
      {
        resourceType: 'gpu',
        label: 'GPU',
        current: isActive ? 87 : 0,
        allocated: 100,
        unit: '%',
        history: generateHistory(87, 10),
      },
    ],
    cost: {
      currentPeriodCost: isActive ? 190.0 : 0,
      projectedMonthlyCost: isActive ? 1825.0 : 0,
      escrowBalance: isActive ? 310.0 : 0,
      escrowTotal: 500.0,
      currency: 'USD',
      costHistory: generateHistory(2.5, 0.3),
    },
    alerts: isActive
      ? [
          {
            id: `${orderId}-alert-1`,
            type: 'warning',
            resourceType: 'gpu',
            message: 'GPU utilization above 85% for over 1 hour',
            threshold: 85,
            currentValue: 87,
            createdAt: new Date(now.getTime() - 3600000).toISOString(),
            dismissed: false,
          },
        ]
      : [],
  };
}

function generateMockOrderDetail(orderId: string): OrderDetail {
  const statusOptions: OrderStatus[] = ['running', 'pending', 'deploying', 'running', 'running'];
  const idx = Math.abs(hashCode(orderId)) % statusOptions.length;
  const status = statusOptions[idx];
  const now = new Date();
  const createdAt = new Date(now.getTime() - 86400000 * 3);

  return {
    id: orderId,
    providerId: 'provider-1',
    providerName: 'CloudCore',
    providerAddress: 'virtengine1prov1abc...7h3k',
    offeringName: 'NVIDIA A100 Cluster',
    resourceType: 'GPU Compute',
    status,
    region: 'US-East',
    createdAt: createdAt.toISOString(),
    updatedAt: now.toISOString(),
    expiresAt: new Date(now.getTime() + 86400000 * 4).toISOString(),
    cost: {
      hourlyRate: 2.5,
      totalCost: 190.0,
      currency: 'USD',
      denom: 'uve',
    },
    resources: {
      cpu: 16,
      memory: 64,
      storage: 500,
      gpu: 2,
    },
    timeline: generateMockTimeline(orderId, status),
    access: generateMockAccess(orderId, status),
    usage: generateMockUsage(orderId, status),
    txHash: 'ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890',
  };
}

function hashCode(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash |= 0;
  }
  return hash;
}

// =============================================================================
// Hook Implementation
// =============================================================================

/**
 * Real-time order tracking hook.
 * In production, connects via WebSocket for live updates.
 * Currently uses polling with mock data.
 */
export function useOrderTracking({
  orderId,
  enabled = true,
  pollingInterval = 30000,
  onStatusChange,
  onError,
}: UseOrderTrackingOptions): UseOrderTrackingReturn {
  const [order, setOrder] = useState<OrderDetail | null>(null);
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('disconnected');
  const [lastEvent, setLastEvent] = useState<OrderTrackingEvent | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const prevStatusRef = useRef<OrderStatus | null>(null);

  const fetchOrder = useCallback(async () => {
    if (!orderId) return;

    try {
      setIsLoading(true);
      setError(null);

      // In production: await apiClient.get<OrderDetail>(`/orders/${orderId}`)
      await new Promise((resolve) => setTimeout(resolve, 300));
      const detail = generateMockOrderDetail(orderId);

      setOrder(detail);
      setConnectionStatus('connected');

      if (prevStatusRef.current && prevStatusRef.current !== detail.status) {
        const event: OrderTrackingEvent = {
          type: 'status_change',
          orderId,
          timestamp: new Date().toISOString(),
          payload: { from: prevStatusRef.current, to: detail.status },
        };
        setLastEvent(event);
        onStatusChange?.(orderId, detail.status);
      }
      prevStatusRef.current = detail.status;
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to fetch order';
      setError(msg);
      setConnectionStatus('error');
      onError?.(err instanceof Error ? err : new Error(msg));
    } finally {
      setIsLoading(false);
    }
  }, [orderId, onStatusChange, onError]);

  useEffect(() => {
    if (!enabled || !orderId) {
      setConnectionStatus('disconnected');
      return;
    }

    setConnectionStatus('connecting');
    void fetchOrder();

    intervalRef.current = setInterval(() => void fetchOrder(), pollingInterval);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      setConnectionStatus('disconnected');
    };
  }, [enabled, orderId, pollingInterval, fetchOrder]);

  const disconnect = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
    setConnectionStatus('disconnected');
  }, []);

  return {
    order,
    connectionStatus,
    lastEvent,
    isLoading,
    error,
    refresh: fetchOrder,
    disconnect,
  };
}
