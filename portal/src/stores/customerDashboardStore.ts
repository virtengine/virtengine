/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Customer dashboard Zustand store with mock data.
 * In production, fetches from on-chain queries and provider usage reports.
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

// =============================================================================
// Store Interface
// =============================================================================

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
  fetchDashboard: () => Promise<void>;
  setAllocationFilter: (filter: CustomerAllocationStatus | 'all') => void;
  markNotificationRead: (id: string) => void;
  dismissNotification: (id: string) => void;
  terminateAllocation: (id: string) => Promise<void>;
  clearError: () => void;
}

export type CustomerDashboardStore = CustomerDashboardState & CustomerDashboardActions;

// =============================================================================
// Mock Data
// =============================================================================

const MOCK_STATS: CustomerDashboardStats = {
  activeAllocations: 5,
  totalOrders: 12,
  pendingOrders: 2,
  monthlySpend: 4250,
  spendChange: -8.2,
};

const MOCK_ALLOCATIONS: CustomerAllocation[] = [
  {
    id: 'calloc-001',
    orderId: 'order-1001',
    providerName: 'CloudCore',
    providerAddress: 'virtengine1prov1abc...7h3k',
    offeringName: 'NVIDIA A100 Cluster',
    status: 'running',
    resources: { cpu: 32, memory: 128, storage: 1000, gpu: 4 },
    costPerHour: 3.6,
    totalSpent: 2592,
    currency: 'USD',
    createdAt: '2025-01-05T10:00:00Z',
    updatedAt: '2025-02-06T08:00:00Z',
  },
  {
    id: 'calloc-002',
    orderId: 'order-1002',
    providerName: 'DataNexus',
    providerAddress: 'virtengine1prov2def...9j4m',
    offeringName: 'AMD EPYC Compute',
    status: 'running',
    resources: { cpu: 64, memory: 256, storage: 2000 },
    costPerHour: 1.2,
    totalSpent: 864,
    currency: 'USD',
    createdAt: '2025-01-20T14:00:00Z',
    updatedAt: '2025-02-05T12:00:00Z',
  },
  {
    id: 'calloc-003',
    orderId: 'order-1003',
    providerName: 'HPCGrid',
    providerAddress: 'virtengine1prov3ghi...2k7n',
    offeringName: 'HPC Batch Node',
    status: 'running',
    resources: { cpu: 16, memory: 64, storage: 500 },
    costPerHour: 0.45,
    totalSpent: 324,
    currency: 'USD',
    createdAt: '2025-01-25T09:00:00Z',
    updatedAt: '2025-02-06T06:00:00Z',
  },
  {
    id: 'calloc-004',
    orderId: 'order-1004',
    providerName: 'CloudCore',
    providerAddress: 'virtengine1prov1abc...7h3k',
    offeringName: 'NVMe Block Storage',
    status: 'running',
    resources: { cpu: 0, memory: 0, storage: 5000 },
    costPerHour: 0.1,
    totalSpent: 72,
    currency: 'USD',
    createdAt: '2025-02-01T11:00:00Z',
    updatedAt: '2025-02-06T10:00:00Z',
  },
  {
    id: 'calloc-005',
    orderId: 'order-1005',
    providerName: 'AICompute',
    providerAddress: 'virtengine1prov4jkl...5p8q',
    offeringName: 'RTX 4090 AI Instance',
    status: 'deploying',
    resources: { cpu: 16, memory: 64, storage: 500, gpu: 2 },
    costPerHour: 1.8,
    totalSpent: 0,
    currency: 'USD',
    createdAt: '2025-02-06T10:30:00Z',
    updatedAt: '2025-02-06T10:30:00Z',
  },
  {
    id: 'calloc-006',
    orderId: 'order-1006',
    providerName: 'DataNexus',
    providerAddress: 'virtengine1prov2def...9j4m',
    offeringName: 'ML Training Platform',
    status: 'failed',
    resources: { cpu: 48, memory: 192, storage: 2000, gpu: 8 },
    costPerHour: 6.0,
    totalSpent: 48,
    currency: 'USD',
    createdAt: '2025-02-04T13:00:00Z',
    updatedAt: '2025-02-04T21:00:00Z',
  },
  {
    id: 'calloc-007',
    orderId: 'order-1007',
    providerName: 'HPCGrid',
    providerAddress: 'virtengine1prov3ghi...2k7n',
    offeringName: 'HPC Batch Node',
    status: 'terminated',
    resources: { cpu: 32, memory: 128, storage: 1000 },
    costPerHour: 0.9,
    totalSpent: 350,
    currency: 'USD',
    createdAt: '2024-11-01T08:00:00Z',
    updatedAt: '2025-01-15T17:00:00Z',
  },
];

const MOCK_USAGE: UsageSummaryData = {
  resources: [
    { label: 'CPU', used: 86, allocated: 128, unit: 'cores' },
    { label: 'Memory', used: 320, allocated: 512, unit: 'GB' },
    { label: 'Storage', used: 5.8, allocated: 9, unit: 'TB' },
    { label: 'GPU', used: 4, allocated: 6, unit: 'units' },
  ],
  overallUtilization: 65,
};

const MOCK_BILLING: BillingSummaryData = {
  currentPeriodCost: 4250,
  previousPeriodCost: 4630,
  changePercent: -8.2,
  totalLifetimeSpend: 28400,
  outstandingBalance: 1250,
  byProvider: [
    { providerName: 'CloudCore', amount: 2664, percentage: 62.7 },
    { providerName: 'DataNexus', amount: 912, percentage: 21.5 },
    { providerName: 'HPCGrid', amount: 324, percentage: 7.6 },
    { providerName: 'Other', amount: 350, percentage: 8.2 },
  ],
  history: [
    { period: 'Sep 2024', amount: 2100, orders: 3 },
    { period: 'Oct 2024', amount: 3200, orders: 5 },
    { period: 'Nov 2024', amount: 3800, orders: 6 },
    { period: 'Dec 2024', amount: 4100, orders: 8 },
    { period: 'Jan 2025', amount: 4630, orders: 10 },
    { period: 'Feb 2025', amount: 4250, orders: 12 },
  ],
};

const MOCK_NOTIFICATIONS: DashboardNotification[] = [
  {
    id: 'notif-001',
    title: 'Allocation deployed',
    message: 'RTX 4090 AI Instance is now deploying on AICompute.',
    severity: 'info',
    read: false,
    createdAt: '2025-02-06T10:30:00Z',
  },
  {
    id: 'notif-002',
    title: 'Allocation failed',
    message: 'ML Training Platform on DataNexus encountered a provisioning error.',
    severity: 'error',
    read: false,
    createdAt: '2025-02-04T21:00:00Z',
  },
  {
    id: 'notif-003',
    title: 'High GPU utilization',
    message: 'A100 Cluster GPU utilization exceeded 90% for the last hour.',
    severity: 'warning',
    read: false,
    createdAt: '2025-02-06T09:15:00Z',
  },
  {
    id: 'notif-004',
    title: 'Invoice ready',
    message: 'January 2025 invoice ($4,630.00) is ready for review.',
    severity: 'info',
    read: true,
    createdAt: '2025-02-01T00:00:00Z',
  },
  {
    id: 'notif-005',
    title: 'Payment processed',
    message: 'Payment of $4,100.00 for December 2024 has been settled.',
    severity: 'success',
    read: true,
    createdAt: '2025-01-05T06:00:00Z',
  },
];

// =============================================================================
// Store Implementation
// =============================================================================

const initialState: CustomerDashboardState = {
  stats: MOCK_STATS,
  allocations: [],
  usage: MOCK_USAGE,
  billing: MOCK_BILLING,
  notifications: [],
  isLoading: false,
  error: null,
  allocationFilter: 'all',
};

export const useCustomerDashboardStore = create<CustomerDashboardStore>()((set, get) => ({
  ...initialState,

  fetchDashboard: async () => {
    set({ isLoading: true, error: null });

    try {
      // In production, fetches from on-chain queries and provider APIs:
      // const [stats, allocations, usage, billing, notifications] = await Promise.all([
      //   fetch(`${API_BASE}/customer/stats`),
      //   fetch(`${API_BASE}/customer/allocations`),
      //   fetch(`${API_BASE}/customer/usage`),
      //   fetch(`${API_BASE}/customer/billing`),
      //   fetch(`${API_BASE}/customer/notifications`),
      // ]);

      await new Promise((resolve) => setTimeout(resolve, 600));

      set({
        stats: MOCK_STATS,
        allocations: MOCK_ALLOCATIONS,
        usage: MOCK_USAGE,
        billing: MOCK_BILLING,
        notifications: MOCK_NOTIFICATIONS,
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
      // In production: MsgTerminateAllocation broadcast via wallet
      await new Promise((resolve) => setTimeout(resolve, 500));

      const { allocations } = get();
      set({
        allocations: allocations.map((a) =>
          a.id === id
            ? { ...a, status: 'terminated' as const, updatedAt: new Date().toISOString() }
            : a
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

// =============================================================================
// Selectors
// =============================================================================

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
