/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Customer dashboard store backed by chain + provider daemon data.
 */

import { create } from 'zustand';
import type {
  BillingSummaryData,
  CustomerAllocation,
  CustomerAllocationStatus,
  CustomerDashboardStats,
  DashboardNotification,
  UsageSummaryData,
} from '@/types/customer';
import {
  fetchPaginated,
  fetchChainJsonWithFallback,
  coerceNumber,
  coerceString,
  toDate,
} from '@/lib/api/chain';
import { getPortalEndpoints } from '@/lib/config';
import { MultiProviderClient } from '@/lib/portal-adapter';

export interface CustomerDashboardState {
  stats: CustomerDashboardStats;
  allocations: CustomerAllocation[];
  usage: UsageSummaryData;
  billing: BillingSummaryData;
  notifications: DashboardNotification[];
  isLoading: boolean;
  error: string | null;
  allocationFilter: CustomerAllocationStatus | 'all';
}

export interface CustomerDashboardActions {
  fetchDashboard: (ownerAddress: string) => Promise<void>;
  setAllocationFilter: (filter: CustomerAllocationStatus | 'all') => void;
  markNotificationRead: (id: string) => void;
  dismissNotification: (id: string) => void;
  terminateAllocation: (id: string) => Promise<void>;
  clearError: () => void;
}

export type CustomerDashboardStore = CustomerDashboardState & CustomerDashboardActions;

const LEASE_ENDPOINTS = ['/virtengine/market/v1/leases', '/virtengine/market/v1beta5/leases'];
const ORDER_ENDPOINTS = ['/virtengine/market/v1beta5/orders', '/virtengine/market/v1/orders'];
const ESCROW_ENDPOINTS = ['/virtengine/escrow/v1/accounts', '/virtengine/escrow/v1beta1/accounts'];
const PROVIDER_ENDPOINTS = (address: string) => [
  `/virtengine/provider/v1/providers/${address}`,
  `/virtengine/provider/v1beta4/providers/${address}`,
];

const initialState: CustomerDashboardState = {
  stats: {
    activeAllocations: 0,
    totalOrders: 0,
    pendingOrders: 0,
    monthlySpend: 0,
    spendChange: 0,
  },
  allocations: [],
  usage: {
    resources: [],
    overallUtilization: 0,
  },
  billing: {
    currentPeriodCost: 0,
    previousPeriodCost: 0,
    changePercent: 0,
    totalLifetimeSpend: 0,
    outstandingBalance: 0,
    byProvider: [],
    history: [],
  },
  notifications: [],
  isLoading: false,
  error: null,
  allocationFilter: 'all',
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

const parseAllocationStatus = (state: unknown): CustomerAllocationStatus => {
  const normalized = coerceString(state, '').toLowerCase();
  if (normalized.includes('pending')) return 'pending';
  if (normalized.includes('deploy')) return 'deploying';
  if (normalized.includes('pause')) return 'paused';
  if (normalized.includes('fail') || normalized.includes('error')) return 'failed';
  if (normalized.includes('term') || normalized.includes('close')) return 'terminated';
  return 'running';
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

const parseResourceSpec = (raw: Record<string, unknown>) => {
  const resources = (raw.resources ?? raw.resource ?? { cpu: 0, memory: 0, storage: 0 }) as Record<
    string,
    unknown
  >;
  return {
    cpu: coerceNumber(resources.cpu ?? resources.cpu_cores ?? resources.cores, 0),
    memory: coerceNumber(resources.memory ?? resources.memory_gb ?? resources.ram, 0),
    storage: coerceNumber(resources.storage ?? resources.storage_gb ?? resources.disk, 0),
    gpu: coerceNumber(resources.gpu ?? resources.gpu_count, 0) || undefined,
  };
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
  const info = raw.info as Record<string, unknown> | undefined;
  const name = info ? coerceString(info.name, '') : '';
  return name || fallback;
};

const buildNotification = (
  id: string,
  title: string,
  message: string,
  severity: DashboardNotification['severity'],
  createdAt: Date
): DashboardNotification => ({
  id,
  title,
  message,
  severity,
  read: false,
  createdAt: createdAt.toISOString(),
});

export const useCustomerDashboardStore = create<CustomerDashboardStore>()((set, get) => ({
  ...initialState,

  fetchDashboard: async (ownerAddress: string) => {
    set({ isLoading: true, error: null });

    try {
      if (!ownerAddress) {
        throw new Error('Wallet address is required to load dashboard data.');
      }

      const [leaseResult, orderResult, escrowResult] = await Promise.all([
        fetchPaginated<Record<string, unknown>>(LEASE_ENDPOINTS, 'leases', {
          params: { owner: ownerAddress },
        }),
        fetchPaginated<Record<string, unknown>>(ORDER_ENDPOINTS, 'orders', {
          params: { owner: ownerAddress },
        }),
        fetchPaginated<Record<string, unknown>>(ESCROW_ENDPOINTS, 'accounts', {
          params: { owner: ownerAddress },
        }),
      ]);

      const providerMap = new Map<string, string>();
      await Promise.all(
        leaseResult.items.map(async (lease) => {
          const provider = coerceString(
            lease.provider ?? (lease.id as Record<string, unknown>)?.provider,
            ''
          );
          if (!provider || providerMap.has(provider)) return;
          try {
            const payload = await fetchChainJsonWithFallback<Record<string, unknown>>(
              PROVIDER_ENDPOINTS(provider)
            );
            const rawProvider =
              (payload.provider as Record<string, unknown> | undefined) ?? payload;
            providerMap.set(provider, parseProviderName(rawProvider, provider));
          } catch {
            providerMap.set(provider, provider);
          }
        })
      );

      const allocations = leaseResult.items.map((record) => {
        const leaseIdRecord =
          record.id && typeof record.id === 'object'
            ? (record.id as Record<string, unknown>)
            : undefined;
        const provider = coerceString(record.provider ?? leaseIdRecord?.provider, '');
        const status = parseAllocationStatus(record.state ?? record.status);
        const leaseId = buildLeaseId(record);
        const createdAt = toDate(record.created_at ?? record.createdAt ?? record.created);
        const updatedAt = toDate(record.updated_at ?? record.updatedAt ?? record.updated);
        const resources = parseResourceSpec(record);
        const priceRecord =
          record.price && typeof record.price === 'object'
            ? (record.price as Record<string, unknown>)
            : undefined;

        return {
          id: leaseId,
          orderId: coerceString(record.order_id ?? record.orderId, leaseId),
          providerName: providerMap.get(provider) ?? provider,
          providerAddress: provider,
          offeringName: coerceString(record.offering_name ?? record.offeringName, 'Compute'),
          status,
          resources,
          costPerHour: coerceNumber(priceRecord?.amount ?? record.price_per_hour ?? 0, 0),
          totalSpent: coerceNumber(record.total_spent ?? record.totalSpent, 0),
          currency: coerceString(priceRecord?.denom ?? record.currency, 'uve'),
          createdAt: createdAt.toISOString(),
          updatedAt: updatedAt.toISOString(),
        };
      });

      const activeAllocations = allocations.filter((a) =>
        ['running', 'deploying', 'paused'].includes(a.status)
      );

      const pendingOrders = orderResult.items.filter((order) => {
        const status = coerceString(order.state ?? order.status, '').toLowerCase();
        return status.includes('open') || status.includes('pending');
      }).length;

      const totalOrders = orderResult.items.length;

      const escrowAccounts = escrowResult.items;
      const escrowTotals = escrowAccounts.reduce<{
        total: number;
        byProvider: Map<string, number>;
      }>(
        (acc, record) => {
          const balance =
            record.balance && typeof record.balance === 'object'
              ? (record.balance as Record<string, unknown>)
              : undefined;
          const amount = coerceNumber(balance?.amount ?? record.amount, 0);
          acc.total += amount;
          const provider = coerceString(record.provider ?? record.provider_address, '');
          if (provider) {
            acc.byProvider.set(provider, (acc.byProvider.get(provider) ?? 0) + amount);
          }
          return acc;
        },
        { total: 0, byProvider: new Map<string, number>() }
      );

      const byProvider = Array.from(escrowTotals.byProvider.entries()).map(([provider, amount]) => {
        const name = providerMap.get(provider) ?? provider;
        return { providerName: name, amount, percentage: 0 };
      });

      const billing: BillingSummaryData = {
        currentPeriodCost: escrowTotals.total,
        previousPeriodCost: 0,
        changePercent: 0,
        totalLifetimeSpend: escrowTotals.total,
        outstandingBalance: escrowTotals.total,
        byProvider,
        history: [],
      };

      const stats: CustomerDashboardStats = {
        activeAllocations: activeAllocations.length,
        totalOrders,
        pendingOrders,
        monthlySpend: billing.currentPeriodCost,
        spendChange: billing.changePercent,
      };

      const usageTotals = {
        cpu: { used: 0, limit: 0 },
        memory: { used: 0, limit: 0 },
        storage: { used: 0, limit: 0 },
        gpu: { used: 0, limit: 0 },
      };

      try {
        const client = await getProviderClient();
        await Promise.all(
          allocations.map(async (allocation) => {
            if (!allocation.id) return;
            try {
              const metrics = await client
                .getClient(allocation.providerAddress)
                ?.getDeploymentMetrics(allocation.id);
              if (!metrics) return;
              usageTotals.cpu.used += metrics.cpu.usage ?? 0;
              usageTotals.cpu.limit += metrics.cpu.limit ?? 0;
              usageTotals.memory.used += metrics.memory.usage ?? 0;
              usageTotals.memory.limit += metrics.memory.limit ?? 0;
              usageTotals.storage.used += metrics.storage.usage ?? 0;
              usageTotals.storage.limit += metrics.storage.limit ?? 0;
              if (metrics.gpu) {
                usageTotals.gpu.used += metrics.gpu.usage ?? 0;
                usageTotals.gpu.limit += metrics.gpu.limit ?? 0;
              }
            } catch {
              // ignore individual metric failures
            }
          })
        );
      } catch {
        // provider metrics not available
      }

      const usage: UsageSummaryData = {
        resources: [
          {
            label: 'CPU',
            used: usageTotals.cpu.used,
            allocated:
              usageTotals.cpu.limit || allocations.reduce((sum, a) => sum + a.resources.cpu, 0),
            unit: 'cores',
          },
          {
            label: 'Memory',
            used: usageTotals.memory.used,
            allocated:
              usageTotals.memory.limit ||
              allocations.reduce((sum, a) => sum + a.resources.memory, 0),
            unit: 'GB',
          },
          {
            label: 'Storage',
            used: usageTotals.storage.used,
            allocated:
              usageTotals.storage.limit ||
              allocations.reduce((sum, a) => sum + a.resources.storage, 0),
            unit: 'GB',
          },
          {
            label: 'GPU',
            used: usageTotals.gpu.used,
            allocated:
              usageTotals.gpu.limit ||
              allocations.reduce((sum, a) => sum + (a.resources.gpu ?? 0), 0),
            unit: 'units',
          },
        ],
        overallUtilization:
          usageTotals.cpu.limit > 0
            ? Math.round((usageTotals.cpu.used / usageTotals.cpu.limit) * 100)
            : 0,
      };

      const notifications: DashboardNotification[] = [];
      allocations.slice(0, 6).forEach((allocation) => {
        const createdAt = toDate(allocation.updatedAt);
        if (allocation.status === 'failed') {
          notifications.push(
            buildNotification(
              `alloc-${allocation.id}`,
              'Allocation failed',
              `${allocation.offeringName} failed on ${allocation.providerName}.`,
              'error',
              createdAt
            )
          );
        } else if (allocation.status === 'deploying') {
          notifications.push(
            buildNotification(
              `alloc-${allocation.id}`,
              'Allocation deploying',
              `${allocation.offeringName} is deploying on ${allocation.providerName}.`,
              'info',
              createdAt
            )
          );
        } else if (allocation.status === 'running') {
          notifications.push(
            buildNotification(
              `alloc-${allocation.id}`,
              'Allocation active',
              `${allocation.offeringName} is active on ${allocation.providerName}.`,
              'success',
              createdAt
            )
          );
        }
      });

      set({
        stats,
        allocations,
        usage,
        billing,
        notifications,
        isLoading: false,
      });
    } catch (error) {
      set({
        isLoading: false,
        error: error instanceof Error ? error.message : 'Failed to load customer dashboard',
      });
    }
  },

  setAllocationFilter: (filter) => {
    set({ allocationFilter: filter });
  },

  markNotificationRead: (id) => {
    const { notifications } = get();
    set({
      notifications: notifications.map((n) => (n.id === id ? { ...n, read: true } : n)),
    });
  },

  dismissNotification: (id) => {
    const { notifications } = get();
    set({
      notifications: notifications.filter((n) => n.id !== id),
    });
  },

  terminateAllocation: async (id) => {
    try {
      await Promise.resolve();
      const { allocations } = get();
      set({
        allocations: allocations.map((a) =>
          a.id === id ? { ...a, status: 'terminated', updatedAt: new Date().toISOString() } : a
        ),
      });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Failed to terminate allocation',
      });
    }
  },

  clearError: () => {
    set({ error: null });
  },
}));

export const selectFilteredCustomerAllocations = (
  state: CustomerDashboardStore
): CustomerAllocation[] => {
  if (state.allocationFilter === 'all') return state.allocations;
  return state.allocations.filter((a) => a.status === state.allocationFilter);
};

export const selectActiveCustomerAllocations = (
  state: CustomerDashboardStore
): CustomerAllocation[] => {
  return state.allocations.filter(
    (a) => a.status === 'running' || a.status === 'deploying' || a.status === 'paused'
  );
};

export const selectTotalMonthlySpend = (state: CustomerDashboardStore): number => {
  return state.allocations
    .filter((a) => a.status === 'running')
    .reduce((sum, a) => sum + a.costPerHour * 730, 0);
};

export const selectUnreadNotificationCount = (state: CustomerDashboardStore): number => {
  return state.notifications.filter((n) => !n.read).length;
};

export const selectAllocationById = (
  state: CustomerDashboardStore,
  id: string
): CustomerAllocation | undefined => {
  return state.allocations.find((a) => a.id === id);
};
