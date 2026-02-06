/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Provider dashboard Zustand store with mock data.
 * In production, fetches from provider daemon API and on-chain queries.
 */

import { create } from 'zustand';
import type {
  Allocation,
  AllocationStatus,
  CapacityData,
  Payout,
  ProviderDashboardStats,
  QueuedAllocation,
  RevenueSummaryData,
} from '@/types/provider';

// =============================================================================
// Store Interface
// =============================================================================

export interface ProviderStoreState {
  stats: ProviderDashboardStats;
  allocations: Allocation[];
  revenue: RevenueSummaryData;
  capacity: CapacityData;
  payouts: Payout[];
  queue: QueuedAllocation[];
  isLoading: boolean;
  error: string | null;
  allocationFilter: AllocationStatus | 'all';
}

export interface ProviderStoreActions {
  fetchDashboard: () => Promise<void>;
  setAllocationFilter: (filter: AllocationStatus | 'all') => void;
  clearError: () => void;
}

export type ProviderStore = ProviderStoreState & ProviderStoreActions;

// =============================================================================
// Mock Data
// =============================================================================

const MOCK_STATS: ProviderDashboardStats = {
  activeAllocations: 23,
  totalOfferings: 6,
  publishedOfferings: 5,
  monthlyRevenue: 12450,
  revenueChange: 15.3,
  uptime: 99.82,
  pendingOrders: 4,
  openTickets: 2,
};

const MOCK_ALLOCATIONS: Allocation[] = [
  {
    id: 'alloc-001',
    offeringName: 'NVIDIA A100 Cluster',
    offeringId: 'virtengine1provider1abc-1',
    customerAddress: 'virtengine1cust1abc...7h3k',
    customerName: 'Acme Corp',
    status: 'ok',
    resources: { cpu: 32, memory: 128, storage: 1000, gpu: 4 },
    monthlyRevenue: 1800,
    createdAt: '2024-12-01T10:00:00Z',
    updatedAt: '2025-02-05T08:00:00Z',
  },
  {
    id: 'alloc-002',
    offeringName: 'AMD EPYC 7763 Instance',
    offeringId: 'virtengine1provider2xyz-1',
    customerAddress: 'virtengine1cust2def...9j4m',
    customerName: 'DataLab Inc',
    status: 'ok',
    resources: { cpu: 64, memory: 256, storage: 2000 },
    monthlyRevenue: 950,
    createdAt: '2025-01-10T14:00:00Z',
    updatedAt: '2025-02-04T12:00:00Z',
  },
  {
    id: 'alloc-003',
    offeringName: 'HPC Compute Node',
    offeringId: 'virtengine1provider3def-1',
    customerAddress: 'virtengine1cust3ghi...2k7n',
    customerName: 'ResearchLab',
    status: 'ok',
    resources: { cpu: 80, memory: 512, storage: 4000 },
    monthlyRevenue: 5760,
    createdAt: '2024-11-15T09:00:00Z',
    updatedAt: '2025-02-06T06:00:00Z',
  },
  {
    id: 'alloc-004',
    offeringName: 'RTX 4090 Gaming/AI',
    offeringId: 'virtengine1provider4ghi-1',
    customerAddress: 'virtengine1cust4jkl...5p8q',
    customerName: 'AI Startup Co',
    status: 'creating',
    resources: { cpu: 16, memory: 64, storage: 500, gpu: 2 },
    monthlyRevenue: 360,
    createdAt: '2025-02-05T16:00:00Z',
    updatedAt: '2025-02-05T16:00:00Z',
  },
  {
    id: 'alloc-005',
    offeringName: 'NVMe Block Storage',
    offeringId: 'virtengine1provider1abc-2',
    customerAddress: 'virtengine1cust1abc...7h3k',
    customerName: 'Acme Corp',
    status: 'ok',
    resources: { cpu: 0, memory: 0, storage: 5000 },
    monthlyRevenue: 500,
    createdAt: '2024-12-20T11:00:00Z',
    updatedAt: '2025-02-03T10:00:00Z',
  },
  {
    id: 'alloc-006',
    offeringName: 'ML Training Platform',
    offeringId: 'virtengine1provider5jkl-1',
    customerAddress: 'virtengine1cust5mno...3r9s',
    customerName: 'DeepLearn Ltd',
    status: 'erred',
    resources: { cpu: 48, memory: 192, storage: 2000, gpu: 8 },
    monthlyRevenue: 0,
    createdAt: '2025-02-04T13:00:00Z',
    updatedAt: '2025-02-06T02:00:00Z',
  },
  {
    id: 'alloc-007',
    offeringName: 'NVIDIA A100 Cluster',
    offeringId: 'virtengine1provider1abc-1',
    customerAddress: 'virtengine1cust6pqr...8t2u',
    customerName: 'BioSim Research',
    status: 'terminated',
    resources: { cpu: 32, memory: 128, storage: 1000, gpu: 4 },
    monthlyRevenue: 0,
    createdAt: '2024-09-01T08:00:00Z',
    updatedAt: '2025-01-15T17:00:00Z',
  },
];

const MOCK_REVENUE: RevenueSummaryData = {
  currentMonth: 12450,
  previousMonth: 10800,
  changePercent: 15.3,
  totalLifetime: 87250,
  pendingPayouts: 3200,
  byOffering: [
    { offeringName: 'HPC Compute Node', revenue: 5760, percentage: 46.3 },
    { offeringName: 'NVIDIA A100 Cluster', revenue: 3600, percentage: 28.9 },
    { offeringName: 'AMD EPYC 7763 Instance', revenue: 950, percentage: 7.6 },
    { offeringName: 'NVMe Block Storage', revenue: 500, percentage: 4.0 },
    { offeringName: 'Other', revenue: 1640, percentage: 13.2 },
  ],
  history: [
    { period: 'Sep 2024', revenue: 6200, orders: 8 },
    { period: 'Oct 2024', revenue: 7800, orders: 12 },
    { period: 'Nov 2024', revenue: 9100, orders: 15 },
    { period: 'Dec 2024', revenue: 10200, orders: 18 },
    { period: 'Jan 2025', revenue: 10800, orders: 20 },
    { period: 'Feb 2025', revenue: 12450, orders: 23 },
  ],
};

const MOCK_CAPACITY: CapacityData = {
  resources: [
    { label: 'CPU', used: 272, total: 512, unit: 'cores' },
    { label: 'Memory', used: 1280, total: 2048, unit: 'GB' },
    { label: 'GPU', used: 14, total: 24, unit: 'units' },
    { label: 'Storage', used: 15.5, total: 50, unit: 'TB' },
  ],
  overallUtilization: 58,
};

const MOCK_PAYOUTS: Payout[] = [
  {
    id: 'pay-001',
    amount: 10800,
    currency: 'USD',
    status: 'completed',
    txHash: '0xabc123...def456',
    period: 'January 2025',
    createdAt: '2025-02-01T00:00:00Z',
    completedAt: '2025-02-01T06:00:00Z',
  },
  {
    id: 'pay-002',
    amount: 10200,
    currency: 'USD',
    status: 'completed',
    txHash: '0x789ghi...012jkl',
    period: 'December 2024',
    createdAt: '2025-01-01T00:00:00Z',
    completedAt: '2025-01-01T05:30:00Z',
  },
  {
    id: 'pay-003',
    amount: 3200,
    currency: 'USD',
    status: 'pending',
    period: 'February 2025 (partial)',
    createdAt: '2025-02-06T00:00:00Z',
  },
  {
    id: 'pay-004',
    amount: 9100,
    currency: 'USD',
    status: 'completed',
    txHash: '0x345mno...678pqr',
    period: 'November 2024',
    createdAt: '2024-12-01T00:00:00Z',
    completedAt: '2024-12-01T04:45:00Z',
  },
];

const MOCK_QUEUE: QueuedAllocation[] = [
  {
    id: 'q-001',
    offeringName: 'RTX 4090 Gaming/AI',
    customerAddress: 'virtengine1cust4jkl...5p8q',
    requestedAt: '2025-02-06T10:30:00Z',
    resources: { cpu: 16, memory: 64, storage: 500, gpu: 2 },
    estimatedProvisionTime: '~5 min',
  },
  {
    id: 'q-002',
    offeringName: 'NVIDIA A100 Cluster',
    customerAddress: 'virtengine1cust7stu...4v6w',
    requestedAt: '2025-02-06T11:00:00Z',
    resources: { cpu: 32, memory: 128, storage: 1000, gpu: 4 },
    estimatedProvisionTime: '~10 min',
  },
  {
    id: 'q-003',
    offeringName: 'AMD EPYC 7763 Instance',
    customerAddress: 'virtengine1cust8xyz...1a3b',
    requestedAt: '2025-02-06T11:15:00Z',
    resources: { cpu: 64, memory: 256, storage: 2000 },
    estimatedProvisionTime: '~8 min',
  },
];

// =============================================================================
// Store Implementation
// =============================================================================

const initialState: ProviderStoreState = {
  stats: MOCK_STATS,
  allocations: [],
  revenue: MOCK_REVENUE,
  capacity: MOCK_CAPACITY,
  payouts: [],
  queue: [],
  isLoading: false,
  error: null,
  allocationFilter: 'all',
};

export const useProviderStore = create<ProviderStore>()((set) => ({
  ...initialState,

  fetchDashboard: async () => {
    set({ isLoading: true, error: null });

    try {
      // In production, fetches from provider daemon API:
      // const [stats, allocations, revenue, capacity, payouts, queue] = await Promise.all([
      //   fetch(`${API_BASE}/provider/stats`),
      //   fetch(`${API_BASE}/provider/allocations`),
      //   fetch(`${API_BASE}/provider/revenue`),
      //   fetch(`${API_BASE}/provider/capacity`),
      //   fetch(`${API_BASE}/provider/payouts`),
      //   fetch(`${API_BASE}/provider/queue`),
      // ]);

      await new Promise((resolve) => setTimeout(resolve, 600));

      set({
        stats: MOCK_STATS,
        allocations: MOCK_ALLOCATIONS,
        revenue: MOCK_REVENUE,
        capacity: MOCK_CAPACITY,
        payouts: MOCK_PAYOUTS,
        queue: MOCK_QUEUE,
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

// =============================================================================
// Selectors
// =============================================================================

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
