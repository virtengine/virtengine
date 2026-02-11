import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useProviderStore } from '@/stores/providerStore';
import { fetchPaginated } from '@/lib/api/chain';

vi.mock('@/lib/api/chain', async () => {
  const actual = await vi.importActual<typeof import('@/lib/api/chain')>('@/lib/api/chain');
  return {
    ...actual,
    fetchPaginated: vi.fn(),
  };
});

const { MockMultiProviderClient } = vi.hoisted(() => {
  class MockMultiProviderClient {
    initialize = vi.fn().mockResolvedValue(undefined);
    getProvider = vi.fn(() => ({ status: 'online', lastHealthCheck: new Date() }));
    getClient = vi.fn(() => ({
      listDeployments: vi.fn().mockResolvedValue({ deployments: [{ id: 'dep1' }] }),
      getDeploymentMetrics: vi.fn().mockResolvedValue({
        cpu: { usage: 1, limit: 2 },
        memory: { usage: 2, limit: 4 },
        storage: { usage: 3, limit: 6 },
      }),
    }));
  }
  return { MockMultiProviderClient };
});

vi.mock('@/lib/portal-adapter', () => ({
  MultiProviderClient: MockMultiProviderClient,
}));

type Mocked<T> = T & {
  mockReset: () => void;
  mockImplementation: (fn: (...args: unknown[]) => unknown) => void;
};

const fetchPaginatedMock = fetchPaginated as unknown as Mocked<typeof fetchPaginated>;
const initialState = useProviderStore.getState();

describe('providerStore', () => {
  beforeEach(() => {
    useProviderStore.setState(initialState, true);
    fetchPaginatedMock.mockReset();
  });

  it('loads provider dashboard data from chain and daemon metrics', async () => {
    fetchPaginatedMock.mockImplementation((paths, key) => {
      if (key === 'offerings') {
        return Promise.resolve({
          items: [
            {
              id: 'offering-1',
              name: 'Compute',
              category: 'compute',
              state: 'published',
              pricing: { base_price: '5', currency: 'uve' },
            },
          ],
          nextKey: null,
          total: 1,
        });
      }
      if (key === 'leases') {
        return Promise.resolve({
          items: [
            {
              id: 'lease-1',
              offering_name: 'Compute',
              owner: 've1customer',
              state: 'active',
              cpu: 1,
              memory: 2,
              storage: 3,
            },
          ],
          nextKey: null,
          total: 1,
        });
      }
      if (key === 'bids') {
        return Promise.resolve({
          items: [
            {
              id: 'bid-1',
              offering_name: 'Compute',
              owner: 've1customer',
              price: { amount: '5', denom: 'uve' },
            },
          ],
          nextKey: null,
          total: 1,
        });
      }
      return Promise.resolve({
        items: [
          {
            id: 'settlement-1',
            amount: { amount: '10', denom: 'uve' },
            status: 'pending',
          },
        ],
        nextKey: null,
        total: 1,
      });
    });

    await useProviderStore.getState().fetchDashboard('ve1provider');

    const state = useProviderStore.getState();
    expect(state.offerings).toHaveLength(1);
    expect(state.allocations).toHaveLength(1);
    expect(state.pendingBids).toHaveLength(1);
    expect(state.stats.totalOfferings).toBe(1);
    expect(state.capacity.resources.length).toBeGreaterThan(0);
  });
});
