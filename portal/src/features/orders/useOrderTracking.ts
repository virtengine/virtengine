/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Real-time order tracking hook backed by chain and provider-daemon data.
 */

'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import type { OrderStatus } from '@/stores/orderStore';
import type {
  OrderDetail,
  OrderStatusEvent,
  OrderUsageData,
  ResourceAccessInfo,
  ResourceUsageMetric,
  UsageDataPoint,
} from './tracking-types';
import {
  fetchPaginated,
  fetchChainJsonWithFallback,
  coerceNumber,
  coerceString,
  toDate,
} from '@/lib/api/chain';
import { getPortalEndpoints } from '@/lib/config';
import { MultiProviderClient } from '@/lib/portal-adapter';

export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

export interface OrderTrackingEvent {
  type: 'status_change' | 'usage_update' | 'alert' | 'access_ready' | 'error';
  orderId: string;
  timestamp: string;
  payload: unknown;
}

export interface UseOrderTrackingOptions {
  orderId: string;
  ownerAddress?: string;
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

const ORDER_ENDPOINTS = ['/virtengine/market/v1beta5/orders', '/virtengine/market/v1/orders'];
const LEASE_ENDPOINTS = ['/virtengine/market/v1/leases', '/virtengine/market/v1beta5/leases'];
const ESCROW_ENDPOINTS = ['/virtengine/escrow/v1/accounts', '/virtengine/escrow/v1beta1/accounts'];
const PROVIDER_ENDPOINTS = (address: string) => [
  `/virtengine/provider/v1/providers/${address}`,
  `/virtengine/provider/v1beta4/providers/${address}`,
];

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

const parseOrderStatus = (value: unknown): OrderStatus => {
  const normalized = coerceString(value, '').toLowerCase();
  if (normalized.includes('match')) return 'matched';
  if (normalized.includes('deploy')) return 'deploying';
  if (normalized.includes('run') || normalized.includes('active')) return 'running';
  if (normalized.includes('pause')) return 'paused';
  if (normalized.includes('close') || normalized.includes('complete')) return 'completed';
  if (normalized.includes('fail') || normalized.includes('error')) return 'failed';
  if (normalized.includes('stop') || normalized.includes('cancel')) return 'stopped';
  return 'pending';
};

const parseProviderName = (raw: Record<string, unknown>, fallback: string) => {
  const attributes = Array.isArray(raw.attributes) ? raw.attributes : [];
  for (const attr of attributes) {
    if (!attr || typeof attr !== 'object') continue;
    const record = attr as Record<string, unknown>;
    const key = coerceString(record.key, '').toLowerCase();
    if (['name', 'provider_name', 'moniker', 'organization'].includes(key)) {
      const value = coerceString(record.value, '');
      if (value) return value;
    }
  }
  const info =
    raw.info && typeof raw.info === 'object' ? (raw.info as Record<string, unknown>) : undefined;
  return coerceString(info?.name, '') || fallback;
};

const buildLeaseId = (raw: Record<string, unknown>): string => {
  const id = raw.id ?? raw.lease_id ?? raw.leaseId;
  if (typeof id === 'string') return id;
  if (id && typeof id === 'object') {
    const record = id as Record<string, unknown>;
    const owner = coerceString(record.owner, '');
    const dseq = coerceString(record.dseq, '');
    const gseq = coerceString(record.gseq, '');
    const oseq = coerceString(record.oseq, '');
    const provider = coerceString(record.provider, '');
    if (owner && dseq) {
      return [owner, dseq, gseq, oseq, provider].filter(Boolean).join('/');
    }
  }
  return coerceString(raw.lease_id ?? raw.leaseId ?? raw.id, '');
};

const buildTimeline = (
  orderId: string,
  status: OrderStatus,
  createdAt: Date,
  updatedAt: Date,
  leaseId?: string
) => {
  const events: OrderStatusEvent[] = [
    {
      id: `${orderId}-created`,
      status: 'pending',
      title: 'Order created',
      description: 'Order accepted by the marketplace.',
      timestamp: createdAt.toISOString(),
    },
  ];

  if (leaseId) {
    events.push({
      id: `${orderId}-lease`,
      status: status === 'pending' ? 'matched' : status,
      title: 'Lease linked',
      description: `Lease ${leaseId} is associated with this order.`,
      timestamp: updatedAt.toISOString(),
      metadata: { leaseId },
    });
  }

  if (status !== 'pending') {
    events.push({
      id: `${orderId}-status`,
      status,
      title: `Order ${status.replace('_', ' ')}`,
      description: `Latest on-chain state is ${status.replace('_', ' ')}.`,
      timestamp: updatedAt.toISOString(),
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
    events: events.sort(
      (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
    ),
    currentStatus: status,
    estimatedCompletion: status === 'deploying' ? updatedAt.toISOString() : undefined,
    progressPercent: progressMap[status],
  };
};

const metricHistory = (timestamp: string, value: number): UsageDataPoint[] => [
  { timestamp, value },
];

const buildUsageMetric = (
  resourceType: ResourceUsageMetric['resourceType'],
  label: string,
  current: number,
  allocated: number,
  unit: string,
  timestamp: string
): ResourceUsageMetric => ({
  resourceType,
  label,
  current,
  allocated,
  unit,
  history: metricHistory(timestamp, current),
});

const buildUsage = (
  orderId: string,
  status: OrderStatus,
  timestamp: string,
  metrics: {
    cpu?: { usage?: number; limit?: number };
    memory?: { usage?: number; limit?: number };
    storage?: { usage?: number; limit?: number };
    gpu?: { usage?: number; limit?: number };
    network?: { rxBytes?: number; txBytes?: number };
    cost?: { amount?: string; currency?: string };
  } | null,
  fallbackResources: { cpu: number; memory: number; storage: number; gpu?: number },
  totalCost: number,
  hourlyRate: number,
  escrowTotal: number
): OrderUsageData => {
  const usageMetrics: ResourceUsageMetric[] = [
    buildUsageMetric(
      'cpu',
      'CPU',
      coerceNumber(metrics?.cpu?.usage, 0),
      coerceNumber(metrics?.cpu?.limit, fallbackResources.cpu),
      'cores',
      timestamp
    ),
    buildUsageMetric(
      'memory',
      'Memory',
      coerceNumber(metrics?.memory?.usage, 0),
      coerceNumber(metrics?.memory?.limit, fallbackResources.memory),
      'GB',
      timestamp
    ),
    buildUsageMetric(
      'storage',
      'Storage',
      coerceNumber(metrics?.storage?.usage, 0),
      coerceNumber(metrics?.storage?.limit, fallbackResources.storage),
      'GB',
      timestamp
    ),
  ];

  if ((metrics?.gpu?.limit ?? fallbackResources.gpu ?? 0) > 0) {
    usageMetrics.push(
      buildUsageMetric(
        'gpu',
        'GPU',
        coerceNumber(metrics?.gpu?.usage, 0),
        coerceNumber(metrics?.gpu?.limit, fallbackResources.gpu ?? 0),
        'units',
        timestamp
      )
    );
  }

  if (metrics?.network) {
    usageMetrics.push(
      buildUsageMetric(
        'network',
        'Network',
        coerceNumber(metrics.network.rxBytes, 0) + coerceNumber(metrics.network.txBytes, 0),
        0,
        'bytes/s',
        timestamp
      )
    );
  }

  const alerts =
    status === 'running' && metrics?.gpu?.limit && metrics.gpu.usage
      ? metrics.gpu.usage / metrics.gpu.limit > 0.85
        ? [
            {
              id: `${orderId}-gpu-alert`,
              type: 'warning' as const,
              resourceType: 'gpu',
              message: 'GPU utilization exceeded 85% on the latest provider sample.',
              threshold: 85,
              currentValue: Math.round((metrics.gpu.usage / metrics.gpu.limit) * 100),
              createdAt: timestamp,
              dismissed: false,
            },
          ]
        : []
      : [];

  const escrowBalance = Math.max(escrowTotal - totalCost, 0);

  return {
    orderId,
    lastUpdated: timestamp,
    metrics: usageMetrics,
    cost: {
      currentPeriodCost: totalCost,
      projectedMonthlyCost: hourlyRate > 0 ? hourlyRate * 730 : totalCost,
      escrowBalance,
      escrowTotal,
      currency: coerceString(metrics?.cost?.currency, 'uve'),
      costHistory: metricHistory(timestamp, totalCost),
    },
    alerts,
  };
};

const buildAccess = (
  orderId: string,
  status: OrderStatus,
  deploymentStatus: {
    services?: Array<{ name?: string; ports?: Array<{ port?: number; protocol?: string }> }>;
  } | null,
  providerEndpoint?: string
): ResourceAccessInfo => {
  const isProvisioned = status === 'running' || status === 'paused';
  const services = deploymentStatus?.services ?? [];

  const endpoints = services.flatMap((service) =>
    (service.ports ?? []).map((port) => ({
      name: coerceString(service.name, 'Service'),
      url: providerEndpoint
        ? `${providerEndpoint}:${coerceNumber(port.port, 0)}`
        : `${coerceNumber(port.port, 0)}`,
      method: 'TCP',
      description: `${coerceString(port.protocol, 'tcp').toUpperCase()} port ${coerceNumber(port.port, 0)}`,
    }))
  );

  return {
    orderId,
    isProvisioned,
    credentials: [],
    endpoints,
    consoleUrl:
      isProvisioned && providerEndpoint ? `${providerEndpoint}/deployments/${orderId}` : undefined,
  };
};

const fetchOrderDetail = async (orderId: string, ownerAddress: string): Promise<OrderDetail> => {
  const [ordersResult, leaseResult, escrowResult] = await Promise.all([
    fetchPaginated<Record<string, unknown>>(ORDER_ENDPOINTS, 'orders', {
      params: { owner: ownerAddress },
    }),
    fetchPaginated<Record<string, unknown>>(LEASE_ENDPOINTS, 'leases', {
      params: { owner: ownerAddress },
    }),
    fetchPaginated<Record<string, unknown>>(ESCROW_ENDPOINTS, 'accounts', {
      params: { owner: ownerAddress },
    }),
  ]);

  const order = ordersResult.items.find((record) => {
    const id = coerceString(record.id ?? record.order_id ?? record.orderId, '');
    return id === orderId;
  });

  if (!order) {
    throw new Error('Order not found for the connected wallet.');
  }

  const matchingLease = leaseResult.items.find((record) => {
    const linkedOrderId = coerceString(record.order_id ?? record.orderId, '');
    return linkedOrderId === orderId || buildLeaseId(record) === orderId;
  });

  const providerAddress = coerceString(
    matchingLease?.provider ??
      (matchingLease?.id as Record<string, unknown> | undefined)?.provider ??
      order.provider ??
      order.provider_address ??
      order.providerAddress,
    ''
  );

  let providerName = providerAddress;
  try {
    const providerPayload = await fetchChainJsonWithFallback<Record<string, unknown>>(
      PROVIDER_ENDPOINTS(providerAddress)
    );
    const rawProvider =
      (providerPayload.provider as Record<string, unknown> | undefined) ?? providerPayload;
    providerName = parseProviderName(rawProvider, providerAddress);
  } catch {
    providerName = providerAddress;
  }

  const leaseId = matchingLease ? buildLeaseId(matchingLease) : undefined;
  const client = await getProviderClient();
  const daemonClient = providerAddress ? client.getClient(providerAddress) : null;
  const providerRecord = providerAddress ? client.getProvider(providerAddress) : undefined;

  let metrics: Awaited<
    ReturnType<NonNullable<typeof daemonClient>['getDeploymentMetrics']>
  > | null = null;
  let deploymentStatus: Awaited<
    ReturnType<NonNullable<typeof daemonClient>['getDeploymentStatus']>
  > | null = null;
  if (daemonClient && leaseId) {
    try {
      metrics = await daemonClient.getDeploymentMetrics(leaseId);
    } catch {
      metrics = null;
    }
    try {
      deploymentStatus = await daemonClient.getDeploymentStatus(leaseId);
    } catch {
      deploymentStatus = null;
    }
  }

  const createdAt = toDate(order.created_at ?? order.createdAt);
  const updatedAt = toDate(
    matchingLease?.updated_at ??
      matchingLease?.updatedAt ??
      order.updated_at ??
      order.updatedAt ??
      createdAt
  );
  const resources =
    matchingLease?.resources && typeof matchingLease.resources === 'object'
      ? (matchingLease.resources as Record<string, unknown>)
      : order.resources && typeof order.resources === 'object'
        ? (order.resources as Record<string, unknown>)
        : {};

  const hourlyRate = coerceNumber(order.hourly_rate ?? order.price_per_hour, 0);
  const totalCost = coerceNumber(order.total_cost ?? order.cost, 0);
  const escrowTotal = escrowResult.items.reduce((sum, record) => {
    const balance =
      record.balance && typeof record.balance === 'object'
        ? (record.balance as Record<string, unknown>)
        : undefined;
    return sum + coerceNumber(balance?.amount ?? record.amount, 0);
  }, 0);
  const status = parseOrderStatus(
    matchingLease?.state ?? matchingLease?.status ?? order.state ?? order.status
  );

  return {
    id: orderId,
    providerId: providerAddress,
    providerName,
    providerAddress,
    offeringName: coerceString(
      matchingLease?.offering_name ??
        matchingLease?.offeringName ??
        order.offering_name ??
        order.offeringName,
      leaseId ? `Lease ${leaseId}` : 'Deployment'
    ),
    resourceType: coerceString(order.resource_type ?? order.resourceType, 'Compute'),
    status,
    region: coerceString(providerRecord?.attributes?.region, '') || 'unavailable',
    createdAt: createdAt.toISOString(),
    updatedAt: updatedAt.toISOString(),
    expiresAt:
      matchingLease?.expires_at || matchingLease?.expiresAt
        ? toDate(matchingLease.expires_at ?? matchingLease.expiresAt).toISOString()
        : undefined,
    cost: {
      hourlyRate,
      totalCost,
      currency: coerceString(order.currency ?? metrics?.cost?.currency, 'uve'),
      denom: coerceString(order.currency ?? metrics?.cost?.currency, 'uve'),
    },
    resources: {
      cpu: coerceNumber(resources.cpu ?? order.cpu, 0),
      memory: coerceNumber(resources.memory ?? order.memory, 0),
      storage: coerceNumber(resources.storage ?? order.storage, 0),
      gpu: coerceNumber(resources.gpu ?? order.gpu, 0) || undefined,
    },
    timeline: buildTimeline(orderId, status, createdAt, updatedAt, leaseId),
    access: buildAccess(orderId, status, deploymentStatus, providerRecord?.endpoint),
    usage: buildUsage(
      orderId,
      status,
      updatedAt.toISOString(),
      metrics,
      {
        cpu: coerceNumber(resources.cpu ?? order.cpu, 0),
        memory: coerceNumber(resources.memory ?? order.memory, 0),
        storage: coerceNumber(resources.storage ?? order.storage, 0),
        gpu: coerceNumber(resources.gpu ?? order.gpu, 0) || undefined,
      },
      totalCost,
      hourlyRate,
      escrowTotal
    ),
    txHash: coerceString(order.tx_hash ?? order.txHash, '') || undefined,
  };
};

export function useOrderTracking({
  orderId,
  ownerAddress,
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
    if (!orderId || !ownerAddress) return;

    try {
      setIsLoading(true);
      setError(null);

      const detail = await fetchOrderDetail(orderId, ownerAddress);

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
  }, [orderId, ownerAddress, onStatusChange, onError]);

  useEffect(() => {
    if (!enabled || !orderId) {
      setConnectionStatus('disconnected');
      return;
    }

    if (!ownerAddress) {
      setOrder(null);
      setError('Connect the wallet that owns this order to load live order details.');
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
  }, [enabled, orderId, ownerAddress, pollingInterval, fetchOrder]);

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
