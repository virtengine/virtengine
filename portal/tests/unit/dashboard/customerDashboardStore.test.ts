import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import {
  useCustomerDashboardStore,
  selectFilteredCustomerAllocations,
  selectActiveCustomerAllocations,
  selectTotalMonthlySpend,
  selectUnreadNotificationCount,
  selectAllocationById,
} from '@/stores/customerDashboardStore';

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

  async function runWithTimers<T>(fn: () => Promise<T>): Promise<T> {
    const promise = fn();
    await vi.advanceTimersByTimeAsync(1000);
    return promise;
  }

  describe('fetchDashboard', () => {
    it('loads all dashboard data from mock', async () => {
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());

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
      const promise = useCustomerDashboardStore.getState().fetchDashboard();
      expect(useCustomerDashboardStore.getState().isLoading).toBe(true);

      await vi.advanceTimersByTimeAsync(1000);
      await promise;

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
    it('marks notification as read', async () => {
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());

      const { notifications } = useCustomerDashboardStore.getState();
      const unread = notifications.find((n) => !n.read);
      expect(unread).toBeDefined();

      useCustomerDashboardStore.getState().markNotificationRead(unread!.id);

      const updated = useCustomerDashboardStore
        .getState()
        .notifications.find((n) => n.id === unread!.id);
      expect(updated!.read).toBe(true);
    });

    it('dismisses a notification', async () => {
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());

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
    it('returns all allocations when filter is all', async () => {
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());

      const state = useCustomerDashboardStore.getState();
      const filtered = selectFilteredCustomerAllocations(state);
      expect(filtered.length).toBe(state.allocations.length);
    });

    it('filters by status running', async () => {
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());
      useCustomerDashboardStore.getState().setAllocationFilter('running');

      const state = useCustomerDashboardStore.getState();
      const filtered = selectFilteredCustomerAllocations(state);
      expect(filtered.length).toBeGreaterThan(0);
      filtered.forEach((a) => expect(a.status).toBe('running'));
    });

    it('filters by status failed', async () => {
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());
      useCustomerDashboardStore.getState().setAllocationFilter('failed');

      const state = useCustomerDashboardStore.getState();
      const filtered = selectFilteredCustomerAllocations(state);
      expect(filtered.length).toBeGreaterThan(0);
      filtered.forEach((a) => expect(a.status).toBe('failed'));
    });

    it('returns empty for status with no matches', async () => {
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());
      useCustomerDashboardStore.getState().setAllocationFilter('paused');

      const state = useCustomerDashboardStore.getState();
      const filtered = selectFilteredCustomerAllocations(state);
      expect(filtered.length).toBe(0);
    });
  });

  describe('selectActiveCustomerAllocations', () => {
    it('returns only running/deploying/paused allocations', async () => {
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());

      const state = useCustomerDashboardStore.getState();
      const active = selectActiveCustomerAllocations(state);
      expect(active.length).toBeGreaterThan(0);
      active.forEach((a) => {
        expect(['running', 'deploying', 'paused']).toContain(a.status);
      });
    });

    it('excludes terminated and failed allocations', async () => {
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());

      const state = useCustomerDashboardStore.getState();
      const active = selectActiveCustomerAllocations(state);
      active.forEach((a) => {
        expect(a.status).not.toBe('terminated');
        expect(a.status).not.toBe('failed');
      });
    });
  });

  describe('selectTotalMonthlySpend', () => {
    it('sums estimated monthly spend from running allocations', async () => {
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());

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
    it('counts unread notifications', async () => {
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());

      const state = useCustomerDashboardStore.getState();
      const count = selectUnreadNotificationCount(state);
      const expected = state.notifications.filter((n) => !n.read).length;
      expect(count).toBe(expected);
      expect(count).toBeGreaterThan(0);
    });

    it('decreases after marking as read', async () => {
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());

      const before = selectUnreadNotificationCount(useCustomerDashboardStore.getState());
      const unread = useCustomerDashboardStore.getState().notifications.find((n) => !n.read);
      useCustomerDashboardStore.getState().markNotificationRead(unread!.id);
      const after = selectUnreadNotificationCount(useCustomerDashboardStore.getState());

      expect(after).toBe(before - 1);
    });
  });

  describe('billing data', () => {
    it('has valid billing structure', async () => {
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());

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
    it('has valid usage structure', async () => {
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());

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
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());

      const { allocations } = useCustomerDashboardStore.getState();
      const running = allocations.find((a) => a.status === 'running');
      expect(running).toBeDefined();

      await runWithTimers(() =>
        useCustomerDashboardStore.getState().terminateAllocation(running!.id)
      );

      const updated = useCustomerDashboardStore
        .getState()
        .allocations.find((a) => a.id === running!.id);
      expect(updated!.status).toBe('terminated');
    });

    it('does not affect other allocations', async () => {
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());

      const { allocations } = useCustomerDashboardStore.getState();
      const running = allocations.filter((a) => a.status === 'running');
      expect(running.length).toBeGreaterThan(1);

      await runWithTimers(() =>
        useCustomerDashboardStore.getState().terminateAllocation(running[0].id)
      );

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
    it('returns the allocation with matching id', async () => {
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());

      const state = useCustomerDashboardStore.getState();
      const first = state.allocations[0];
      const found = selectAllocationById(state, first.id);
      expect(found).toBeDefined();
      expect(found!.id).toBe(first.id);
    });

    it('returns undefined for non-existent id', async () => {
      await runWithTimers(() => useCustomerDashboardStore.getState().fetchDashboard());

      const state = useCustomerDashboardStore.getState();
      const found = selectAllocationById(state, 'non-existent-id');
      expect(found).toBeUndefined();
    });
  });
});
