/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Provider dashboard store backed by chain + provider daemon data.
 */

import { create } from 'zustand';
import type {
  Allocation,
  AllocationStatus,
  CapacityData,
  PendingBid,
  Payout,
  ProviderDashboardStats,
  ProviderOfferingSummary,
  ProviderSyncStatus,
  QueuedAllocation,
  RevenueSummaryData,
} from '@/types/provider';
import type { OfferingPublicationStatus, OfferingCategory } from '@/types/offering';
import { fetchPaginated, coerceNumber, coerceString, toDate } from '@/lib/api/chain';
import { getPortalEndpoints } from '@/lib/config';
import { MultiProviderClient } from '@/lib/portal-adapter';

export interface ProviderStoreState {
  stats: ProviderDashboardStats;
  allocations: Allocation[];
  offerings: ProviderOfferingSummary[];
  pendingBids: PendingBid[];
  syncStatus: ProviderSyncStatus;
  revenue: RevenueSummaryData;
  capacity: CapacityData;
  payouts: Payout[];
  queue: QueuedAllocation[];
  isLoading: boolean;
  error: string | null;
  allocationFilter: AllocationStatus | 'all';
}

export interface ProviderStoreActions {
  fetchDashboard: (providerAddress: string) => Promise<void>;
  setAllocationFilter: (filter: AllocationStatus | 'all') => void;
  clearError: () => void;
}

export type ProviderStore = ProviderStoreState & ProviderStoreActions;

const OFFERING_ENDPOINTS = [
  '/virtengine/market/v1/offerings',
  '/virtengine/market/v1beta5/offerings',
];
const LEASE_ENDPOINTS = ['/virtengine/market/v1/leases', '/virtengine/market/v1beta5/leases'];
const BID_ENDPOINTS = ['/virtengine/market/v1beta5/bids', '/virtengine/market/v1/bids'];
const SETTLEMENT_ENDPOINTS = [
  '/virtengine/settlement/v1/settlements',
  '/virtengine/settlement/v1beta1/settlements',
];

const initialState: ProviderStoreState = {
  stats: {
    activeAllocations: 0,
    totalOfferings: 0,
    publishedOfferings: 0,
    monthlyRevenue: 0,
    revenueChange: 0,
    uptime: 0,
    pendingOrders: 0,
    openTickets: 0,
  },
  allocations: [],
  offerings: [],
  pendingBids: [],
  syncStatus: {
    isRunning: false,
    lastSyncAt: '',
    nextSyncAt: '',
    errorCount: 0,
    pendingOfferings: 0,
    pendingAllocations: 0,
    waldur: {
      name: 'Waldur',
      status: 'offline',
      lastSuccessAt: '',
      lagSeconds: 0,
    },
    chain: {
      name: 'VirtEngine Chain',
      status: 'offline',
      lastSuccessAt: '',
      lagSeconds: 0,
    },
    providerDaemon: {
      name: 'Provider Daemon',
      status: 'offline',
      lastSuccessAt: '',
      lagSeconds: 0,
    },
  },
  revenue: {
    currentMonth: 0,
    previousMonth: 0,
    changePercent: 0,
    totalLifetime: 0,
    pendingPayouts: 0,
    byOffering: [],
    history: [],
  },
  capacity: {
    resources: [],
    overallUtilization: 0,
  },
  payouts: [],
  queue: [],
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

const parseAllocationStatus = (value: unknown): AllocationStatus => {
  const normalized = coerceString(value, '').toLowerCase();
  if (normalized.includes('create') || normalized.includes('allocating')) return 'creating';
  if (normalized.includes('update')) return 'updating';
  if (normalized.includes('term')) return 'terminating';
  if (normalized.includes('closed')) return 'terminated';
  if (normalized.includes('error') || normalized.includes('failed')) return 'erred';
  if (normalized.includes('pending')) return 'pending';
  return 'ok';
};

const parseOfferingStatus = (value: unknown): OfferingPublicationStatus => {
  const normalized = coerceString(value, '').toLowerCase();
  if (normalized.includes('pause')) return 'paused';
  if (normalized.includes('draft')) return 'draft';
  if (normalized.includes('deprecated')) return 'deprecated';
  if (normalized.includes('fail') || normalized.includes('suspend')) return 'failed';
  if (normalized.includes('pending')) return 'pending';
  return 'published';
};

const parseOfferingCategory = (value: unknown): OfferingCategory => {
  const normalized = coerceString(value, '').toLowerCase();
  if (normalized.includes('gpu')) return 'gpu';
  if (normalized.includes('storage')) return 'storage';
  if (normalized.includes('network')) return 'network';
  if (normalized.includes('hpc')) return 'hpc';
  if (normalized.includes('ml')) return 'ml';
  if (normalized.includes('compute')) return 'compute';
  return 'other';
};

export const useProviderStore = create<ProviderStore>()((set) => ({
  ...initialState,

  fetchDashboard: async (providerAddress: string) => {
    set({ isLoading: true, error: null });

    try {
      if (!providerAddress) {
        throw new Error('Provider address is required.');
      }

      const [offeringsResult, leaseResult, bidResult, settlementsResult] = await Promise.all([
        fetchPaginated<Record<string, unknown>>(OFFERING_ENDPOINTS, 'offerings', {
          params: { provider: providerAddress },
        }),
        fetchPaginated<Record<string, unknown>>(LEASE_ENDPOINTS, 'leases', {
          params: { provider: providerAddress },
        }),
        fetchPaginated<Record<string, unknown>>(BID_ENDPOINTS, 'bids', {
          params: { provider: providerAddress },
        }),
        fetchPaginated<Record<string, unknown>>(SETTLEMENT_ENDPOINTS, 'settlements', {
          params: { provider: providerAddress },
        }),
      ]);

      const offerings = offeringsResult.items.map((record) => {
        const pricing =
          record.pricing && typeof record.pricing === 'object'
            ? (record.pricing as Record<string, unknown>)
            : {};
        return {
          id: coerceString(record.id ?? record.offering_id ?? record.offeringId, ''),
          name: coerceString(record.name, 'Offering'),
          category: parseOfferingCategory(record.category),
          status: parseOfferingStatus(record.state ?? record.status),
          syncStatus: 'synced' as ProviderOfferingSummary['syncStatus'],
          activeOrders: coerceNumber(record.active_order_count ?? record.activeOrderCount, 0),
          totalOrders: coerceNumber(record.total_order_count ?? record.totalOrderCount, 0),
          basePrice: coerceNumber(pricing.base_price ?? pricing.basePrice, 0),
          currency: coerceString(pricing.currency, 'uve'),
          updatedAt: toDate(record.updated_at ?? record.updatedAt).toISOString(),
          lastSyncedAt: toDate(record.updated_at ?? record.updatedAt).toISOString(),
        };
      });

      const allocations = leaseResult.items.map((record) => {
        const leaseId = coerceString(
          record.id ?? record.lease_id ?? record.leaseId,
          coerceString(record.order_id ?? record.orderId, '')
        );
        const resources =
          record.resources && typeof record.resources === 'object'
            ? (record.resources as Record<string, unknown>)
            : {};
        return {
          id: leaseId,
          offeringName: coerceString(record.offering_name ?? record.offeringName, 'Offering'),
          offeringId: coerceString(record.offering_id ?? record.offeringId, ''),
          customerAddress: coerceString(
            record.owner ?? record.customer ?? record.customer_address,
            ''
          ),
          customerName: coerceString(record.customer_name ?? record.customerName, 'Customer'),
          status: parseAllocationStatus(record.state ?? record.status),
          resources: {
            cpu: coerceNumber(record.cpu ?? resources.cpu, 0),
            memory: coerceNumber(record.memory ?? resources.memory, 0),
            storage: coerceNumber(record.storage ?? resources.storage, 0),
            gpu: coerceNumber(record.gpu ?? resources.gpu, 0) || undefined,
          },
          monthlyRevenue: coerceNumber(record.monthly_revenue ?? record.revenue, 0),
          createdAt: toDate(record.created_at ?? record.createdAt).toISOString(),
          updatedAt: toDate(record.updated_at ?? record.updatedAt).toISOString(),
        };
      });

      const pendingBids = bidResult.items.map((record) => {
        const price =
          record.price && typeof record.price === 'object'
            ? (record.price as Record<string, unknown>)
            : undefined;
        const resources =
          record.resources && typeof record.resources === 'object'
            ? (record.resources as Record<string, unknown>)
            : {};
        return {
          id: coerceString(record.id ?? record.bid_id ?? record.bidId, ''),
          offeringName: coerceString(record.offering_name ?? record.offeringName, 'Offering'),
          customerName: coerceString(record.customer_name ?? record.customerName, 'Customer'),
          customerAddress: coerceString(
            record.owner ?? record.customer ?? record.customer_address,
            ''
          ),
          status: 'awaiting_customer' as PendingBid['status'],
          bidAmount: coerceNumber(price?.amount ?? record.amount, 0),
          currency: coerceString(price?.denom ?? record.currency, 'uve'),
          duration: coerceString(record.duration ?? record.duration_str, ''),
          createdAt: toDate(record.created_at ?? record.createdAt).toISOString(),
          expiresAt: toDate(record.expires_at ?? record.expiresAt ?? record.expiry).toISOString(),
          resources: {
            cpu: coerceNumber(record.cpu ?? resources.cpu, 0),
            memory: coerceNumber(record.memory ?? resources.memory, 0),
            storage: coerceNumber(record.storage ?? resources.storage, 0),
            gpu: coerceNumber(record.gpu ?? resources.gpu, 0) || undefined,
          },
        };
      });

      const payouts = settlementsResult.items.map((record) => {
        const amount =
          record.amount && typeof record.amount === 'object'
            ? (record.amount as Record<string, unknown>)
            : undefined;
        return {
          id: coerceString(record.id ?? record.settlement_id ?? record.settlementId, ''),
          amount: coerceNumber(amount?.amount ?? record.amount, 0),
          currency: coerceString(amount?.denom ?? record.currency, 'uve'),
          status: coerceString(record.status, 'pending') as Payout['status'],
          txHash: coerceString(record.tx_hash ?? record.txHash, '') || undefined,
          period: coerceString(record.period, ''),
          createdAt: toDate(record.created_at ?? record.createdAt).toISOString(),
          completedAt: record.completed_at ? toDate(record.completed_at).toISOString() : undefined,
        } as Payout;
      });

      const revenueTotal = payouts.reduce((sum, payout) => sum + payout.amount, 0);
      const revenue: RevenueSummaryData = {
        currentMonth: revenueTotal,
        previousMonth: 0,
        changePercent: 0,
        totalLifetime: revenueTotal,
        pendingPayouts: payouts
          .filter((payout) => payout.status === 'pending')
          .reduce((sum, payout) => sum + payout.amount, 0),
        byOffering: offerings.map((offering) => ({
          offeringName: offering.name,
          revenue: offering.totalOrders,
          percentage: 0,
        })),
        history: [],
      };

      const stats: ProviderDashboardStats = {
        activeAllocations: allocations.filter((a) => a.status === 'ok').length,
        totalOfferings: offerings.length,
        publishedOfferings: offerings.filter((o) => o.status === 'published').length,
        monthlyRevenue: revenue.currentMonth,
        revenueChange: revenue.changePercent,
        uptime: 0,
        pendingOrders: pendingBids.length,
        openTickets: 0,
      };

      let capacity: CapacityData = { resources: [], overallUtilization: 0 };
      let syncStatus = initialState.syncStatus;

      try {
        const client = await getProviderClient();
        const provider = client.getProvider(providerAddress);
        const daemonClient = client.getClient(providerAddress);

        if (daemonClient) {
          const deployments = await daemonClient.listDeployments();
          const metrics = await Promise.allSettled(
            deployments.deployments.map((deployment) =>
              daemonClient.getDeploymentMetrics(deployment.id)
            )
          );
          const totals = {
            cpu: { used: 0, total: 0 },
            memory: { used: 0, total: 0 },
            storage: { used: 0, total: 0 },
            gpu: { used: 0, total: 0 },
          };
          metrics.forEach((result) => {
            if (result.status !== 'fulfilled') return;
            totals.cpu.used += result.value.cpu.usage ?? 0;
            totals.cpu.total += result.value.cpu.limit ?? 0;
            totals.memory.used += result.value.memory.usage ?? 0;
            totals.memory.total += result.value.memory.limit ?? 0;
            totals.storage.used += result.value.storage.usage ?? 0;
            totals.storage.total += result.value.storage.limit ?? 0;
            if (result.value.gpu) {
              totals.gpu.used += result.value.gpu.usage ?? 0;
              totals.gpu.total += result.value.gpu.limit ?? 0;
            }
          });
          capacity = {
            resources: [
              { label: 'CPU', used: totals.cpu.used, total: totals.cpu.total, unit: 'cores' },
              { label: 'Memory', used: totals.memory.used, total: totals.memory.total, unit: 'GB' },
              {
                label: 'Storage',
                used: totals.storage.used,
                total: totals.storage.total,
                unit: 'GB',
              },
              { label: 'GPU', used: totals.gpu.used, total: totals.gpu.total, unit: 'units' },
            ],
            overallUtilization:
              totals.cpu.total > 0 ? Math.round((totals.cpu.used / totals.cpu.total) * 100) : 0,
          };
        }

        syncStatus = {
          isRunning: true,
          lastSyncAt: new Date().toISOString(),
          nextSyncAt: new Date(Date.now() + 5 * 60 * 1000).toISOString(),
          errorCount: 0,
          pendingOfferings: 0,
          pendingAllocations: 0,
          waldur: {
            name: 'Waldur',
            status: 'synced',
            lastSuccessAt: new Date().toISOString(),
            lagSeconds: 0,
          },
          chain: {
            name: 'VirtEngine Chain',
            status: 'synced',
            lastSuccessAt: new Date().toISOString(),
            lagSeconds: 0,
          },
          providerDaemon: {
            name: 'Provider Daemon',
            status: provider?.status === 'online' ? 'synced' : 'offline',
            lastSuccessAt: provider?.lastHealthCheck?.toISOString() ?? '',
            lagSeconds: 0,
            message: provider?.error ?? undefined,
          },
        };
      } catch {
        // provider daemon not available
      }

      set({
        stats,
        allocations,
        offerings,
        pendingBids,
        syncStatus,
        revenue,
        capacity,
        payouts,
        queue: [],
        isLoading: false,
      });
    } catch (error) {
      set({
        isLoading: false,
        error: error instanceof Error ? error.message : 'Failed to load provider dashboard',
      });
    }
  },

  setAllocationFilter: (filter) => {
    set({ allocationFilter: filter });
  },

  clearError: () => {
    set({ error: null });
  },
}));

export const selectFilteredAllocations = (state: ProviderStore): Allocation[] => {
  if (state.allocationFilter === 'all') return state.allocations;
  return state.allocations.filter((a) => a.status === state.allocationFilter);
};

export const selectActiveAllocations = (state: ProviderStore): Allocation[] => {
  return state.allocations.filter(
    (a) => a.status === 'ok' || a.status === 'creating' || a.status === 'updating'
  );
};

export const selectTotalMonthlyRevenue = (state: ProviderStore): number => {
  return state.allocations
    .filter((a) => a.status === 'ok')
    .reduce((sum, a) => sum + a.monthlyRevenue, 0);
};
