import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { MockedFunction } from 'vitest';
import { useCustomerDashboardStore } from '@/stores/customerDashboardStore';
import { fetchPaginated, fetchChainJsonWithFallback } from '@/lib/api/chain';

vi.mock('@/lib/api/chain', async () => {
  const actual = await vi.importActual<typeof import('@/lib/api/chain')>('@/lib/api/chain');
  return {
    ...actual,
    fetchPaginated: vi.fn(),
    fetchChainJsonWithFallback: vi.fn(),
  };
});

const { MockMultiProviderClient } = vi.hoisted(() => {
  class MockMultiProviderClient {
    initialize = vi.fn().mockResolvedValue(undefined);
    getClient = vi.fn(() => null);
  }
  return { MockMultiProviderClient };
});

vi.mock('@/lib/portal-adapter', () => ({
  MultiProviderClient: MockMultiProviderClient,
}));

const fetchPaginatedMock = fetchPaginated as unknown as MockedFunction<typeof fetchPaginated>;
const fetchChainMock = fetchChainJsonWithFallback as unknown as MockedFunction<
  typeof fetchChainJsonWithFallback
>;

const initialState = useCustomerDashboardStore.getState();

describe('customerDashboardStore', () => {
  beforeEach(() => {
    useCustomerDashboardStore.setState(initialState, true);
    fetchPaginatedMock.mockReset();
    fetchChainMock.mockReset();
  });

  it('loads allocations, stats, and billing from chain data', async () => {
    fetchPaginatedMock.mockImplementation((paths, key) => {
      if (key === 'leases') {
        const now = new Date();
        return Promise.resolve({
          items: [
            {
              id: {
                owner: 've1owner',
                dseq: '1',
                gseq: '1',
                oseq: '1',
                provider: 've1provider',
              },
              provider: 've1provider',
              state: 'running',
              resources: { cpu: 2, memory: 4, storage: 10 },
              price: { amount: '5', denom: 'uve' },
              total_spent: '20',
              created_at: now.toISOString(),
              updated_at: now.toISOString(),
              offering_name: 'Compute',
            },
          ],
          nextKey: null,
          total: 1,
        });
      }
      if (key === 'orders') {
        return Promise.resolve({
          items: [{ id: 'order-1', state: 'open' }],
          nextKey: null,
          total: 1,
        });
      }
      return Promise.resolve({
        items: [
          {
            provider: 've1provider',
            balance: { amount: '20', denom: 'uve' },
          },
        ],
        nextKey: null,
        total: 1,
      });
    });

    fetchChainMock.mockResolvedValue({
      provider: { info: { name: 'Provider One' } },
    });

    await useCustomerDashboardStore.getState().fetchDashboard('ve1owner');

    const state = useCustomerDashboardStore.getState();
    expect(state.allocations).toHaveLength(1);
    expect(state.stats.totalOrders).toBe(1);
    expect(state.billing.currentPeriodCost).toBe(20);
    expect(state.billing.outstandingBalance).toBe(20);
    expect(state.escrowAccounts).toHaveLength(1);
    expect(state.allocations[0].providerName).toBe('Provider One');
  });

  it('does not block the first paint on a slow escrow payments endpoint', async () => {
    // Regression test: escrow payments used to be awaited in the same
    // Promise.all as the allocation-bearing reads, so an unreachable payments
    // endpoint held the whole dashboard in its loading skeleton until the
    // request failed. Only the authoritative reads may gate isLoading=false.
    let releasePayments: (() => void) | undefined;
    const paymentsGate = new Promise<void>((resolve) => {
      releasePayments = resolve;
    });

    fetchPaginatedMock.mockImplementation((paths, key) => {
      if (key === 'payments') {
        // Never resolves until we release it, then resolves empty.
        return paymentsGate.then(() => ({ items: [], nextKey: null, total: 0 }));
      }
      if (key === 'leases') {
        return Promise.resolve({
          items: [
            {
              id: {
                owner: 've1owner',
                dseq: '1',
                gseq: '1',
                oseq: '1',
                provider: 've1provider',
              },
              provider: 've1provider',
              state: 'running',
              resources: { cpu: 2, memory: 4, storage: 10 },
              price: { amount: '5', denom: 'uve' },
              total_spent: '20',
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
              offering_name: 'Compute',
            },
          ],
          nextKey: null,
          total: 1,
        });
      }
      return Promise.resolve({ items: [], nextKey: null, total: 0 });
    });
    fetchChainMock.mockResolvedValue({ provider: { info: { name: 'Provider One' } } });

    await useCustomerDashboardStore.getState().fetchDashboard('ve1owner');

    // The dashboard has rendered even though the payments request is still in
    // flight.
    const state = useCustomerDashboardStore.getState();
    expect(state.isLoading).toBe(false);
    expect(state.allocations).toHaveLength(1);
    expect(state.escrowPayments).toEqual([]);

    releasePayments?.();
    await paymentsGate;
  });

  it('commits escrow payments once the detached request resolves', async () => {
    fetchPaginatedMock.mockImplementation((paths, key) => {
      if (key === 'payments') {
        return Promise.resolve({
          items: [
            {
              payment_id: 'payment-1',
              account_id: { scope: 's1', xid: 'x1' },
              state: 'open',
              rate: { amount: '7' },
              balance: { amount: '11' },
              withdrawn: { amount: '3' },
            },
          ],
          nextKey: null,
          total: 1,
        });
      }
      return Promise.resolve({ items: [], nextKey: null, total: 0 });
    });
    fetchChainMock.mockResolvedValue({});

    await useCustomerDashboardStore.getState().fetchDashboard('ve1owner');
    // Let the detached payments promise settle and run its `.then`.
    await Promise.resolve();
    await Promise.resolve();
    await Promise.resolve();

    const payments = useCustomerDashboardStore.getState().escrowPayments;
    expect(payments).toHaveLength(1);
    expect(payments[0]).toMatchObject({
      paymentId: 'payment-1',
      scope: 's1',
      xid: 'x1',
      rateAmount: 7,
      balanceAmount: 11,
      withdrawnAmount: 3,
    });
  });
});
