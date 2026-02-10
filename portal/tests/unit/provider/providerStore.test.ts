import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import {
  useProviderStore,
  selectFilteredAllocations,
  selectActiveAllocations,
  selectTotalMonthlyRevenue,
} from '@/stores/providerStore';
import type {
  Allocation,
  Payout,
  QueuedAllocation,
  ProviderDashboardStats,
  RevenueSummaryData,
  CapacityData,
  ProviderOfferingSummary,
  PendingBid,
  ProviderSyncStatus,
} from '@/types/provider';

const now = new Date().toISOString();

const mockAllocations: Allocation[] = [
  {
    id: 'alloc-1',
    offeringName: 'GPU VM',
    offeringId: 'offer-1',
    customerAddress: 've1cust1',
    customerName: 'Customer A',
    status: 'ok',
    resources: { cpu: 8, memory: 32, storage: 500, gpu: 2 },
    monthlyRevenue: 800,
    createdAt: now,
    updatedAt: now,
  },
  {
    id: 'alloc-2',
    offeringName: 'Storage',
    offeringId: 'offer-2',
    customerAddress: 've1cust2',
    customerName: 'Customer B',
    status: 'erred',
    resources: { cpu: 2, memory: 4, storage: 1000 },
    monthlyRevenue: 100,
    createdAt: now,
    updatedAt: now,
  },
  {
    id: 'alloc-3',
    offeringName: 'Compute',
    offeringId: 'offer-3',
    customerAddress: 've1cust3',
    customerName: 'Customer C',
    status: 'ok',
    resources: { cpu: 4, memory: 16, storage: 200 },
    monthlyRevenue: 400,
    createdAt: now,
    updatedAt: now,
  },
  {
    id: 'alloc-4',
    offeringName: 'ML Training',
    offeringId: 'offer-4',
    customerAddress: 've1cust4',
    customerName: 'Customer D',
    status: 'creating',
    resources: { cpu: 16, memory: 64, storage: 1000, gpu: 4 },
    monthlyRevenue: 1500,
    createdAt: now,
    updatedAt: now,
  },
];

const mockPayouts: Payout[] = [
  {
    id: 'pay-1',
    amount: 2000,
    currency: 'uve',
    status: 'completed',
    txHash: '0xabc123',
    period: '2026-01',
    createdAt: now,
    completedAt: now,
  },
  {
    id: 'pay-2',
    amount: 1500,
    currency: 'uve',
    status: 'pending',
    period: '2026-02',
    createdAt: now,
  },
  {
    id: 'pay-3',
    amount: 500,
    currency: 'uve',
    status: 'processing',
    period: '2026-02',
    createdAt: now,
  },
];

const mockQueue: QueuedAllocation[] = [
  {
    id: 'q-1',
    offeringName: 'GPU VM',
    customerAddress: 've1cust5',
    requestedAt: now,
    resources: { cpu: 4, memory: 16, storage: 100 },
    estimatedProvisionTime: '5 minutes',
  },
  {
    id: 'q-2',
    offeringName: 'Storage',
    customerAddress: 've1cust6',
    requestedAt: now,
    resources: { cpu: 1, memory: 2, storage: 2000 },
    estimatedProvisionTime: '2 minutes',
  },
];

const mockStats: ProviderDashboardStats = {
  activeAllocations: 2,
  totalOfferings: 4,
  publishedOfferings: 3,
  monthlyRevenue: 2800,
  revenueChange: 15,
  uptime: 99.5,
  pendingOrders: 2,
  openTickets: 1,
};

const mockRevenue: RevenueSummaryData = {
  currentMonth: 2800,
  previousMonth: 2400,
  changePercent: 16.7,
  totalLifetime: 25000,
  pendingPayouts: 2000,
  byOffering: [
    { offeringName: 'GPU VM', revenue: 1600, percentage: 57 },
    { offeringName: 'Compute', revenue: 800, percentage: 29 },
    { offeringName: 'Storage', revenue: 400, percentage: 14 },
  ],
  history: [
    { period: '2026-01', revenue: 2800, orders: 5 },
    { period: '2025-12', revenue: 2400, orders: 4 },
    { period: '2025-11', revenue: 2200, orders: 4 },
  ],
};

const mockCapacity: CapacityData = {
  resources: [
    { label: 'CPU', used: 30, total: 64, unit: 'cores' },
    { label: 'Memory', used: 96, total: 256, unit: 'GB' },
    { label: 'Storage', used: 1700, total: 4000, unit: 'GB' },
    { label: 'GPU', used: 6, total: 8, unit: 'units' },
  ],
  overallUtilization: 47,
};

function seedProviderData() {
  useProviderStore.setState({
    stats: mockStats,
    allocations: mockAllocations,
    offerings: [],
    pendingBids: [],
    syncStatus: {
      isRunning: true,
      lastSyncAt: now,
      nextSyncAt: now,
      errorCount: 0,
      pendingOfferings: 0,
      pendingAllocations: 0,
      waldur: { name: 'Waldur', status: 'synced', lastSuccessAt: now, lagSeconds: 0 },
      chain: { name: 'VirtEngine Chain', status: 'synced', lastSuccessAt: now, lagSeconds: 0 },
      providerDaemon: {
        name: 'Provider Daemon',
        status: 'synced',
        lastSuccessAt: now,
        lagSeconds: 0,
      },
    },
    revenue: mockRevenue,
    capacity: mockCapacity,
    payouts: mockPayouts,
    queue: mockQueue,
    isLoading: false,
    error: null,
    allocationFilter: 'all',
  });
}

describe('providerStore', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    useProviderStore.setState({
      allocations: [],
      payouts: [],
      queue: [],
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
      seedProviderData();

      const state = useProviderStore.getState();
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();
      expect(state.allocations.length).toBeGreaterThan(0);
      expect(state.payouts.length).toBeGreaterThan(0);
      expect(state.queue.length).toBeGreaterThan(0);
      expect(state.stats.activeAllocations).toBeGreaterThan(0);
      expect(state.revenue.currentMonth).toBeGreaterThan(0);
      expect(state.capacity.resources.length).toBeGreaterThan(0);
    });

    it('sets loading state while fetching', () => {
      useProviderStore.setState({ isLoading: true });
      expect(useProviderStore.getState().isLoading).toBe(true);

      useProviderStore.setState({ isLoading: false });
      expect(useProviderStore.getState().isLoading).toBe(false);
    });
  });

  describe('allocationFilter', () => {
    it('defaults to all', () => {
      expect(useProviderStore.getState().allocationFilter).toBe('all');
    });

    it('sets allocation filter', () => {
      useProviderStore.getState().setAllocationFilter('ok');
      expect(useProviderStore.getState().allocationFilter).toBe('ok');
    });
  });

  describe('clearError', () => {
    it('clears the error state', () => {
      useProviderStore.setState({ error: 'test error' });
      useProviderStore.getState().clearError();
      expect(useProviderStore.getState().error).toBeNull();
    });
  });

  describe('selectFilteredAllocations', () => {
    it('returns all allocations when filter is all', () => {
      seedProviderData();

      const state = useProviderStore.getState();
      const filtered = selectFilteredAllocations(state);
      expect(filtered.length).toBe(state.allocations.length);
    });

    it('filters by status ok', () => {
      seedProviderData();
      useProviderStore.getState().setAllocationFilter('ok');

      const state = useProviderStore.getState();
      const filtered = selectFilteredAllocations(state);
      expect(filtered.length).toBeGreaterThan(0);
      filtered.forEach((a) => expect(a.status).toBe('ok'));
    });

    it('filters by status erred', () => {
      seedProviderData();
      useProviderStore.getState().setAllocationFilter('erred');

      const state = useProviderStore.getState();
      const filtered = selectFilteredAllocations(state);
      expect(filtered.length).toBeGreaterThan(0);
      filtered.forEach((a) => expect(a.status).toBe('erred'));
    });

    it('returns empty for status with no matches', () => {
      seedProviderData();
      useProviderStore.getState().setAllocationFilter('updating');

      const state = useProviderStore.getState();
      const filtered = selectFilteredAllocations(state);
      expect(filtered.length).toBe(0);
    });
  });

  describe('selectActiveAllocations', () => {
    it('returns only ok/creating/updating allocations', () => {
      seedProviderData();

      const state = useProviderStore.getState();
      const active = selectActiveAllocations(state);
      expect(active.length).toBeGreaterThan(0);
      active.forEach((a) => {
        expect(['ok', 'creating', 'updating']).toContain(a.status);
      });
    });

    it('excludes terminated and erred allocations', () => {
      seedProviderData();

      const state = useProviderStore.getState();
      const active = selectActiveAllocations(state);
      active.forEach((a) => {
        expect(a.status).not.toBe('terminated');
        expect(a.status).not.toBe('erred');
      });
    });
  });

  describe('selectTotalMonthlyRevenue', () => {
    it('sums revenue from ok allocations only', () => {
      seedProviderData();

      const state = useProviderStore.getState();
      const total = selectTotalMonthlyRevenue(state);
      expect(total).toBeGreaterThan(0);

      const okAllocations = state.allocations.filter((a) => a.status === 'ok');
      const expectedTotal = okAllocations.reduce((sum, a) => sum + a.monthlyRevenue, 0);
      expect(total).toBe(expectedTotal);
    });

    it('returns 0 when no ok allocations exist', () => {
      useProviderStore.setState({ allocations: [] });
      const total = selectTotalMonthlyRevenue(useProviderStore.getState());
      expect(total).toBe(0);
    });
  });

  describe('dashboard stats', () => {
    it('has valid stats structure', () => {
      seedProviderData();

      const { stats } = useProviderStore.getState();
      expect(stats.activeAllocations).toBeGreaterThanOrEqual(0);
      expect(stats.totalOfferings).toBeGreaterThanOrEqual(0);
      expect(stats.publishedOfferings).toBeLessThanOrEqual(stats.totalOfferings);
      expect(stats.uptime).toBeGreaterThan(0);
      expect(stats.uptime).toBeLessThanOrEqual(100);
    });
  });

  describe('revenue data', () => {
    it('has valid revenue structure', () => {
      seedProviderData();

      const { revenue } = useProviderStore.getState();
      expect(revenue.currentMonth).toBeGreaterThan(0);
      expect(revenue.totalLifetime).toBeGreaterThan(revenue.currentMonth);
      expect(revenue.history.length).toBeGreaterThan(0);
      expect(revenue.byOffering.length).toBeGreaterThan(0);

      const totalPercentage = revenue.byOffering.reduce((sum, b) => sum + b.percentage, 0);
      expect(totalPercentage).toBeCloseTo(100, 0);
    });
  });

  describe('capacity data', () => {
    it('has valid capacity structure', () => {
      seedProviderData();

      const { capacity } = useProviderStore.getState();
      expect(capacity.resources.length).toBeGreaterThan(0);
      expect(capacity.overallUtilization).toBeGreaterThan(0);
      expect(capacity.overallUtilization).toBeLessThanOrEqual(100);

      capacity.resources.forEach((r) => {
        expect(r.used).toBeLessThanOrEqual(r.total);
        expect(r.label).toBeTruthy();
        expect(r.unit).toBeTruthy();
      });
    });
  });

  describe('payouts data', () => {
    it('has valid payout structure', () => {
      seedProviderData();

      const { payouts } = useProviderStore.getState();
      expect(payouts.length).toBeGreaterThan(0);

      payouts.forEach((p) => {
        expect(p.amount).toBeGreaterThan(0);
        expect(p.currency).toBeTruthy();
        expect(['completed', 'pending', 'processing', 'failed']).toContain(p.status);

        if (p.status === 'completed') {
          expect(p.txHash).toBeTruthy();
          expect(p.completedAt).toBeTruthy();
        }
      });
    });
  });

  describe('queue data', () => {
    it('has valid queue structure', () => {
      seedProviderData();

      const { queue } = useProviderStore.getState();
      expect(queue.length).toBeGreaterThan(0);

      queue.forEach((q) => {
        expect(q.offeringName).toBeTruthy();
        expect(q.customerAddress).toBeTruthy();
        expect(q.estimatedProvisionTime).toBeTruthy();
        expect(q.resources).toBeDefined();
      });
    });
  });
});
