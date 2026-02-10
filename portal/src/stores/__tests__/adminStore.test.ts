import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useAdminStore } from '@/stores/adminStore';
import { fetchPaginated, fetchChainJsonWithFallback } from '@/lib/api/chain';
import { apiClient } from '@/lib/api-client';

vi.mock('@/lib/api/chain', async () => {
  const actual = await vi.importActual<typeof import('@/lib/api/chain')>('@/lib/api/chain');
  return {
    ...actual,
    fetchPaginated: vi.fn(),
    fetchChainJsonWithFallback: vi.fn(),
  };
});

vi.mock('@/lib/api-client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
  },
}));

const { MockMultiProviderClient } = vi.hoisted(() => {
  class MockMultiProviderClient {
    initialize = vi.fn().mockResolvedValue(undefined);
    getAggregatedMetrics = vi.fn().mockResolvedValue({
      totalCPU: { used: 4, limit: 8 },
      totalMemory: { used: 16, limit: 32 },
      totalStorage: { used: 64, limit: 128 },
    });
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
const fetchChainJsonMock = fetchChainJsonWithFallback as unknown as Mocked<
  typeof fetchChainJsonWithFallback
>;
const apiClientMock = apiClient as unknown as { get: Mocked<typeof apiClient.get> };
const initialState = useAdminStore.getState();

describe('adminStore', () => {
  beforeEach(() => {
    useAdminStore.setState(initialState, true);
    fetchPaginatedMock.mockReset();
    fetchChainJsonMock.mockReset();
    apiClientMock.get.mockReset();
  });

  it('loads admin data from chain endpoints and provider metrics', async () => {
    fetchPaginatedMock.mockImplementation((_paths, key) => {
      if (key === 'proposals') {
        return Promise.resolve({
          items: [
            {
              id: '1',
              status: 'PROPOSAL_STATUS_VOTING_PERIOD',
              metadata: JSON.stringify({ title: 'Upgrade', summary: 'Upgrade chain' }),
              final_tally_result: { yes_count: '1', no_count: '0' },
            },
          ],
          nextKey: null,
          total: 1,
        });
      }
      if (key === 'validators') {
        return Promise.resolve({
          items: [
            {
              operator_address: 'veval1',
              status: 'BOND_STATUS_BONDED',
              tokens: '100',
              delegator_shares: '100',
              description: { moniker: 'Validator One' },
            },
          ],
          nextKey: null,
          total: 1,
        });
      }
      if (key === 'accounts') {
        return Promise.resolve({
          items: [{ balance: { amount: '250' } }],
          nextKey: null,
          total: 1,
        });
      }
      if (key === 'providers') {
        return Promise.resolve({
          items: [
            {
              owner: 've1provider',
              info: { name: 'Provider One' },
              attributes: [{ key: 'region', value: 'us-east' }],
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
              provider: 've1provider',
              owner: 've1customer',
              offering_name: 'Compute',
            },
          ],
          nextKey: null,
          total: 1,
        });
      }
      if (key === 'reports') {
        return Promise.resolve({
          items: [{ id: 'report-1', parties: 've1a,ve1b', amount: '5', status: 'open' }],
          nextKey: null,
          total: 1,
        });
      }
      return Promise.resolve({ items: [], nextKey: null, total: 0 });
    });

    fetchChainJsonMock.mockImplementation((paths) => {
      const path = paths[0] ?? '';
      if (path.includes('/blocks/latest')) {
        return Promise.resolve({
          block: { header: { height: '100', time: new Date().toISOString() } },
        });
      }
      if (path.includes('/blocks/')) {
        return Promise.resolve({
          block: {
            header: { proposer_address: 'veval1', time: new Date().toISOString() },
            data: { txs: ['tx1'] },
          },
        });
      }
      if (path.includes('/params')) {
        return Promise.resolve({ params: { example: 'value' } });
      }
      if (path.includes('/roles/')) {
        return Promise.resolve({ roles: ['operator'] });
      }
      return Promise.resolve({});
    });

    apiClientMock.get.mockResolvedValue({
      tickets: [{ id: 'ticket-1', title: 'Help', status: 'open', priority: 'low' }],
    });

    await useAdminStore.getState().fetchAdminData('ve1admin');

    const state = useAdminStore.getState();
    expect(state.proposals).toHaveLength(1);
    expect(state.validators).toHaveLength(1);
    expect(state.providers).toHaveLength(1);
    expect(state.providerLeases).toHaveLength(1);
    expect(state.escrowOverview.totalEscrow).toBe(250);
    expect(state.supportTickets).toHaveLength(1);
    expect(state.resourceUtilization.length).toBeGreaterThan(0);
    expect(state.currentUserRoles).toContain('operator');
    expect(state.systemHealth.blockHeight).toBeGreaterThan(0);
  });
});
