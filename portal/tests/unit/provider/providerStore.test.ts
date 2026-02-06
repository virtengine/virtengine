import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import {
  useProviderStore,
  selectFilteredAllocations,
  selectActiveAllocations,
  selectTotalMonthlyRevenue,
} from '@/stores/providerStore';

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

  async function runWithTimers<T>(fn: () => Promise<T>): Promise<T> {
    const promise = fn();
    await vi.advanceTimersByTimeAsync(1000);
    return promise;
  }

  describe('fetchDashboard', () => {
    it('loads all dashboard data from mock', async () => {
      await runWithTimers(() => useProviderStore.getState().fetchDashboard());

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

    it('sets loading state while fetching', async () => {
      const promise = useProviderStore.getState().fetchDashboard();
      expect(useProviderStore.getState().isLoading).toBe(true);

      await vi.advanceTimersByTimeAsync(1000);
      await promise;

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
    it('returns all allocations when filter is all', async () => {
      await runWithTimers(() => useProviderStore.getState().fetchDashboard());

      const state = useProviderStore.getState();
      const filtered = selectFilteredAllocations(state);
      expect(filtered.length).toBe(state.allocations.length);
    });

    it('filters by status ok', async () => {
      await runWithTimers(() => useProviderStore.getState().fetchDashboard());
      useProviderStore.getState().setAllocationFilter('ok');

      const state = useProviderStore.getState();
      const filtered = selectFilteredAllocations(state);
      expect(filtered.length).toBeGreaterThan(0);
      filtered.forEach((a) => expect(a.status).toBe('ok'));
    });

    it('filters by status erred', async () => {
      await runWithTimers(() => useProviderStore.getState().fetchDashboard());
      useProviderStore.getState().setAllocationFilter('erred');

      const state = useProviderStore.getState();
      const filtered = selectFilteredAllocations(state);
      expect(filtered.length).toBeGreaterThan(0);
      filtered.forEach((a) => expect(a.status).toBe('erred'));
    });

    it('returns empty for status with no matches', async () => {
      await runWithTimers(() => useProviderStore.getState().fetchDashboard());
      useProviderStore.getState().setAllocationFilter('updating');

      const state = useProviderStore.getState();
      const filtered = selectFilteredAllocations(state);
      expect(filtered.length).toBe(0);
    });
  });

  describe('selectActiveAllocations', () => {
    it('returns only ok/creating/updating allocations', async () => {
      await runWithTimers(() => useProviderStore.getState().fetchDashboard());

      const state = useProviderStore.getState();
      const active = selectActiveAllocations(state);
      expect(active.length).toBeGreaterThan(0);
      active.forEach((a) => {
        expect(['ok', 'creating', 'updating']).toContain(a.status);
      });
    });

    it('excludes terminated and erred allocations', async () => {
      await runWithTimers(() => useProviderStore.getState().fetchDashboard());

      const state = useProviderStore.getState();
      const active = selectActiveAllocations(state);
      active.forEach((a) => {
        expect(a.status).not.toBe('terminated');
        expect(a.status).not.toBe('erred');
      });
    });
  });

  describe('selectTotalMonthlyRevenue', () => {
    it('sums revenue from ok allocations only', async () => {
      await runWithTimers(() => useProviderStore.getState().fetchDashboard());

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
    it('has valid stats structure', async () => {
      await runWithTimers(() => useProviderStore.getState().fetchDashboard());

      const { stats } = useProviderStore.getState();
      expect(stats.activeAllocations).toBeGreaterThanOrEqual(0);
      expect(stats.totalOfferings).toBeGreaterThanOrEqual(0);
      expect(stats.publishedOfferings).toBeLessThanOrEqual(stats.totalOfferings);
      expect(stats.uptime).toBeGreaterThan(0);
      expect(stats.uptime).toBeLessThanOrEqual(100);
    });
  });

  describe('revenue data', () => {
    it('has valid revenue structure', async () => {
      await runWithTimers(() => useProviderStore.getState().fetchDashboard());

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
    it('has valid capacity structure', async () => {
      await runWithTimers(() => useProviderStore.getState().fetchDashboard());

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
    it('has valid payout structure', async () => {
      await runWithTimers(() => useProviderStore.getState().fetchDashboard());

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
    it('has valid queue structure', async () => {
      await runWithTimers(() => useProviderStore.getState().fetchDashboard());

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
