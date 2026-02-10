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
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-02T00:00:00Z',
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
    expect(state.allocations[0].providerName).toBe('Provider One');
  });
});
