import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import {
  useCustomerDashboardStore,
  selectFilteredCustomerAllocations,
  selectActiveCustomerAllocations,
  selectTotalMonthlySpend,
  selectUnreadNotificationCount,
  selectAllocationById,
} from '@/stores/customerDashboardStore';
import type {
  CustomerAllocation,
  DashboardNotification,
  BillingSummaryData,
  UsageSummaryData,
  CustomerDashboardStats,
} from '@/types/customer';

const now = new Date().toISOString();

const mockAllocations: CustomerAllocation[] = [
  {
    id: 'alloc-1',
    orderId: 'order-1',
    providerName: 'Provider A',
    providerAddress: 've1prov1',
    offeringName: 'GPU Cluster',
    status: 'running',
    resources: { cpu: 8, memory: 32, storage: 500, gpu: 2 },
    costPerHour: 2.5,
    totalSpent: 1200,
    currency: 'uve',
    createdAt: now,
    updatedAt: now,
  },
  {
    id: 'alloc-2',
    orderId: 'order-2',
    providerName: 'Provider B',
    providerAddress: 've1prov2',
    offeringName: 'Storage Node',
    status: 'failed',
    resources: { cpu: 2, memory: 4, storage: 1000 },
    costPerHour: 0.5,
    totalSpent: 50,
    currency: 'uve',
    createdAt: now,
    updatedAt: now,
  },
  {
    id: 'alloc-3',
    orderId: 'order-3',
    providerName: 'Provider A',
    providerAddress: 've1prov1',
    offeringName: 'Compute Node',
    status: 'running',
    resources: { cpu: 4, memory: 16, storage: 200 },
    costPerHour: 1.0,
    totalSpent: 600,
    currency: 'uve',
    createdAt: now,
    updatedAt: now,
  },
  {
    id: 'alloc-4',
    orderId: 'order-4',
    providerName: 'Provider C',
    providerAddress: 've1prov3',
    offeringName: 'Web Server',
    status: 'deploying',
    resources: { cpu: 2, memory: 8, storage: 100 },
    costPerHour: 0.3,
    totalSpent: 0,
    currency: 'uve',
    createdAt: now,
    updatedAt: now,
  },
];

const mockNotifications: DashboardNotification[] = [
  {
    id: 'notif-1',
    title: 'Deployment ready',
    message: 'GPU Cluster is active on Provider A.',
    severity: 'success',
    read: false,
    createdAt: now,
  },
  {
    id: 'notif-2',
    title: 'Payment processed',
    message: 'Monthly payment processed.',
    severity: 'info',
    read: true,
    createdAt: now,
  },
  {
    id: 'notif-3',
    title: 'Allocation failed',
    message: 'Storage Node failed on Provider B.',
    severity: 'error',
    read: false,
    createdAt: now,
  },
];

const mockBilling: BillingSummaryData = {
  currentPeriodCost: 500,
  previousPeriodCost: 450,
  changePercent: 11,
  totalLifetimeSpend: 5000,
  outstandingBalance: 100,
  byProvider: [
    { providerName: 'Provider A', amount: 350, percentage: 70 },
    { providerName: 'Provider B', amount: 150, percentage: 30 },
  ],
  history: [
    { period: '2026-01', amount: 500, orders: 4 },
    { period: '2025-12', amount: 450, orders: 3 },
  ],
};

const mockUsage: UsageSummaryData = {
  resources: [
    { label: 'CPU', used: 10, allocated: 16, unit: 'cores' },
    { label: 'Memory', used: 40, allocated: 60, unit: 'GB' },
    { label: 'Storage', used: 800, allocated: 1800, unit: 'GB' },
    { label: 'GPU', used: 1, allocated: 2, unit: 'units' },
  ],
  overallUtilization: 65,
};

const mockStats: CustomerDashboardStats = {
  activeAllocations: 3,
  totalOrders: 10,
  pendingOrders: 1,
  monthlySpend: 500,
  spendChange: 11,
};

function seedCustomerData() {
  useCustomerDashboardStore.setState({
    allocations: mockAllocations,
    notifications: mockNotifications,
    billing: mockBilling,
    usage: mockUsage,
    stats: mockStats,
    isLoading: false,
    error: null,
    allocationFilter: 'all',
  });
}

describe('customerDashboardStore', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    useCustomerDashboardStore.setState({
      allocations: [],
      notifications: [],
      isLoading: false,
      error: null,
      allocationFilter: 'all',
    });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe('fetchDashboard', () => {
    it('loads all dashboard data from seed', () => {
      seedCustomerData();

      const state = useCustomerDashboardStore.getState();
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();
      expect(state.allocations.length).toBeGreaterThan(0);
      expect(state.notifications.length).toBeGreaterThan(0);
      expect(state.stats.activeAllocations).toBeGreaterThan(0);
      expect(state.billing.currentPeriodCost).toBeGreaterThan(0);
      expect(state.usage.resources.length).toBeGreaterThan(0);
    });

    it('sets loading state while fetching', async () => {
      useCustomerDashboardStore.setState({ isLoading: true });
      expect(useCustomerDashboardStore.getState().isLoading).toBe(true);

      useCustomerDashboardStore.setState({ isLoading: false });
      expect(useCustomerDashboardStore.getState().isLoading).toBe(false);
    });
  });

  describe('allocationFilter', () => {
    it('defaults to all', () => {
      expect(useCustomerDashboardStore.getState().allocationFilter).toBe('all');
    });

    it('sets allocation filter', () => {
      useCustomerDashboardStore.getState().setAllocationFilter('running');
      expect(useCustomerDashboardStore.getState().allocationFilter).toBe('running');
    });
  });

  describe('notifications', () => {
    it('marks notification as read', () => {
      seedCustomerData();

      const { notifications } = useCustomerDashboardStore.getState();
      const unread = notifications.find((n) => !n.read);
      expect(unread).toBeDefined();

      useCustomerDashboardStore.getState().markNotificationRead(unread!.id);

      const updated = useCustomerDashboardStore
        .getState()
        .notifications.find((n) => n.id === unread!.id);
      expect(updated!.read).toBe(true);
    });

    it('dismisses a notification', () => {
      seedCustomerData();

      const { notifications } = useCustomerDashboardStore.getState();
      const before = notifications.length;

      useCustomerDashboardStore.getState().dismissNotification(notifications[0].id);

      expect(useCustomerDashboardStore.getState().notifications.length).toBe(before - 1);
    });
  });

  describe('clearError', () => {
    it('clears the error state', () => {
      useCustomerDashboardStore.setState({ error: 'test error' });
      useCustomerDashboardStore.getState().clearError();
      expect(useCustomerDashboardStore.getState().error).toBeNull();
    });
  });

  describe('selectFilteredCustomerAllocations', () => {
    it('returns all allocations when filter is all', () => {
      seedCustomerData();

      const state = useCustomerDashboardStore.getState();
      const filtered = selectFilteredCustomerAllocations(state);
      expect(filtered.length).toBe(state.allocations.length);
    });

    it('filters by status running', () => {
      seedCustomerData();
      useCustomerDashboardStore.getState().setAllocationFilter('running');

      const state = useCustomerDashboardStore.getState();
      const filtered = selectFilteredCustomerAllocations(state);
      expect(filtered.length).toBeGreaterThan(0);
      filtered.forEach((a) => expect(a.status).toBe('running'));
    });

    it('filters by status failed', () => {
      seedCustomerData();
      useCustomerDashboardStore.getState().setAllocationFilter('failed');

      const state = useCustomerDashboardStore.getState();
      const filtered = selectFilteredCustomerAllocations(state);
      expect(filtered.length).toBeGreaterThan(0);
      filtered.forEach((a) => expect(a.status).toBe('failed'));
    });

    it('returns empty for status with no matches', () => {
      seedCustomerData();
      useCustomerDashboardStore.getState().setAllocationFilter('paused');

      const state = useCustomerDashboardStore.getState();
      const filtered = selectFilteredCustomerAllocations(state);
      expect(filtered.length).toBe(0);
    });
  });

  describe('selectActiveCustomerAllocations', () => {
    it('returns only running/deploying/paused allocations', () => {
      seedCustomerData();

      const state = useCustomerDashboardStore.getState();
      const active = selectActiveCustomerAllocations(state);
      expect(active.length).toBeGreaterThan(0);
      active.forEach((a) => {
        expect(['running', 'deploying', 'paused']).toContain(a.status);
      });
    });

    it('excludes terminated and failed allocations', () => {
      seedCustomerData();

      const state = useCustomerDashboardStore.getState();
      const active = selectActiveCustomerAllocations(state);
      active.forEach((a) => {
        expect(a.status).not.toBe('terminated');
        expect(a.status).not.toBe('failed');
      });
    });
  });

  describe('selectTotalMonthlySpend', () => {
    it('sums estimated monthly spend from running allocations', () => {
      seedCustomerData();

      const state = useCustomerDashboardStore.getState();
      const total = selectTotalMonthlySpend(state);
      expect(total).toBeGreaterThan(0);

      const running = state.allocations.filter((a) => a.status === 'running');
      const expected = running.reduce((sum, a) => sum + a.costPerHour * 730, 0);
      expect(total).toBeCloseTo(expected, 2);
    });

    it('returns 0 when no running allocations exist', () => {
      useCustomerDashboardStore.setState({ allocations: [] });
      const total = selectTotalMonthlySpend(useCustomerDashboardStore.getState());
      expect(total).toBe(0);
    });
  });

  describe('selectUnreadNotificationCount', () => {
    it('counts unread notifications', () => {
      seedCustomerData();

      const state = useCustomerDashboardStore.getState();
      const count = selectUnreadNotificationCount(state);
      const expected = state.notifications.filter((n) => !n.read).length;
      expect(count).toBe(expected);
      expect(count).toBeGreaterThan(0);
    });

    it('decreases after marking as read', () => {
      seedCustomerData();

      const before = selectUnreadNotificationCount(useCustomerDashboardStore.getState());
      const unread = useCustomerDashboardStore.getState().notifications.find((n) => !n.read);
      useCustomerDashboardStore.getState().markNotificationRead(unread!.id);
      const after = selectUnreadNotificationCount(useCustomerDashboardStore.getState());

      expect(after).toBe(before - 1);
    });
  });

  describe('billing data', () => {
    it('has valid billing structure', () => {
      seedCustomerData();

      const { billing } = useCustomerDashboardStore.getState();
      expect(billing.currentPeriodCost).toBeGreaterThan(0);
      expect(billing.totalLifetimeSpend).toBeGreaterThan(billing.currentPeriodCost);
      expect(billing.history.length).toBeGreaterThan(0);
      expect(billing.byProvider.length).toBeGreaterThan(0);

      const totalPercentage = billing.byProvider.reduce((sum, b) => sum + b.percentage, 0);
      expect(totalPercentage).toBeCloseTo(100, 0);
    });
  });

  describe('usage data', () => {
    it('has valid usage structure', () => {
      seedCustomerData();

      const { usage } = useCustomerDashboardStore.getState();
      expect(usage.resources.length).toBeGreaterThan(0);
      expect(usage.overallUtilization).toBeGreaterThan(0);
      expect(usage.overallUtilization).toBeLessThanOrEqual(100);

      usage.resources.forEach((r) => {
        expect(r.used).toBeLessThanOrEqual(r.allocated);
        expect(r.label).toBeTruthy();
        expect(r.unit).toBeTruthy();
      });
    });
  });

  describe('terminateAllocation', () => {
    it('sets allocation status to terminated', async () => {
      seedCustomerData();

      const { allocations } = useCustomerDashboardStore.getState();
      const running = allocations.find((a) => a.status === 'running');
      expect(running).toBeDefined();

      await useCustomerDashboardStore.getState().terminateAllocation(running!.id);

      const updated = useCustomerDashboardStore
        .getState()
        .allocations.find((a) => a.id === running!.id);
      expect(updated!.status).toBe('terminated');
    });

    it('does not affect other allocations', async () => {
      seedCustomerData();

      const { allocations } = useCustomerDashboardStore.getState();
      const running = allocations.filter((a) => a.status === 'running');
      expect(running.length).toBeGreaterThan(1);

      await useCustomerDashboardStore.getState().terminateAllocation(running[0].id);

      const others = useCustomerDashboardStore
        .getState()
        .allocations.filter((a) => a.id !== running[0].id);
      const originalOthers = allocations.filter((a) => a.id !== running[0].id);
      expect(others.length).toBe(originalOthers.length);
      others.forEach((a, i) => {
        expect(a.status).toBe(originalOthers[i].status);
      });
    });
  });

  describe('selectAllocationById', () => {
    it('returns the allocation with matching id', () => {
      seedCustomerData();

      const state = useCustomerDashboardStore.getState();
      const first = state.allocations[0];
      const found = selectAllocationById(state, first.id);
      expect(found).toBeDefined();
      expect(found!.id).toBe(first.id);
    });

    it('returns undefined for non-existent id', () => {
      seedCustomerData();

      const state = useCustomerDashboardStore.getState();
      const found = selectAllocationById(state, 'non-existent-id');
      expect(found).toBeUndefined();
    });
  });
});
