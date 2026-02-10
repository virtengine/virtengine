import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { MockedFunction } from 'vitest';
import { useOrderStore } from '@/stores/orderStore';
import type { WalletSigner } from '@/lib/api/chain';
import { fetchPaginated, fetchChainJsonWithFallback, signAndBroadcastAmino } from '@/lib/api/chain';

vi.mock('@/lib/api/chain', async () => {
  const actual = await vi.importActual<typeof import('@/lib/api/chain')>('@/lib/api/chain');
  return {
    ...actual,
    fetchPaginated: vi.fn(),
    fetchChainJsonWithFallback: vi.fn(),
    signAndBroadcastAmino: vi.fn(),
  };
});

const fetchPaginatedMock = fetchPaginated as unknown as MockedFunction<typeof fetchPaginated>;
const fetchChainMock = fetchChainJsonWithFallback as unknown as MockedFunction<
  typeof fetchChainJsonWithFallback
>;
const signAndBroadcastMock = signAndBroadcastAmino as unknown as MockedFunction<
  typeof signAndBroadcastAmino
>;

const initialState = useOrderStore.getState();

describe('orderStore', () => {
  beforeEach(() => {
    useOrderStore.setState(initialState, true);
    fetchPaginatedMock.mockReset();
    fetchChainMock.mockReset();
    signAndBroadcastMock.mockReset();
  });

  it('fetches orders from chain and maps provider names', async () => {
    fetchPaginatedMock.mockResolvedValue({
      items: [
        {
          id: 'order-1',
          provider: 've1provider',
          state: 'open',
          resource_type: 'compute',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-02T00:00:00Z',
          resources: { cpu: 2, memory: 4, storage: 10 },
          price_per_hour: '5',
          currency: 'uve',
        },
      ],
      nextKey: null,
      total: 1,
    });
    fetchChainMock.mockResolvedValue({ provider: { info: { name: 'Provider One' } } });

    await useOrderStore.getState().fetchOrders('ve1owner');

    const { orders } = useOrderStore.getState();
    expect(orders).toHaveLength(1);
    expect(orders[0].providerName).toBe('Provider One');
    expect(orders[0].status).toBe('pending');
  });

  it('creates and closes orders with signing', async () => {
    const wallet: WalletSigner = {
      status: 'connected',
      chainId: 'virtengine-1',
      accounts: [{ address: 've1owner', pubKey: new Uint8Array(), algo: 'secp256k1' }],
      activeAccountIndex: 0,
      signAmino: vi.fn(),
      estimateFee: vi
        .fn()
        .mockReturnValue({ amount: [{ denom: 'uve', amount: '1' }], gas: '200000' }),
    };

    signAndBroadcastMock.mockResolvedValue({
      txHash: 'txhash',
      code: 0,
      rawLog: '',
      gasUsed: 100,
      gasWanted: 200,
    });

    const txHash = await useOrderStore.getState().createOrder(
      {
        owner: 've1owner',
        offeringId: 've1provider/1',
        resources: [{ resourceType: 'cpu', unit: 'core', quantity: 1 }],
        deposit: { denom: 'uve', amount: '5' },
      },
      wallet
    );

    expect(txHash).toBe('txhash');

    useOrderStore.setState({
      orders: [
        {
          id: 'order-1',
          providerId: 've1provider',
          providerName: 'Provider One',
          resourceType: 'compute',
          status: 'running',
          createdAt: new Date(),
          updatedAt: new Date(),
          cost: { hourlyRate: 1, totalCost: 2, currency: 'uve' },
          resources: { cpu: 1, memory: 1, storage: 1 },
        },
      ],
    });

    await useOrderStore.getState().closeOrder('order-1', 've1owner', wallet);

    const { orders } = useOrderStore.getState();
    expect(orders[0].status).toBe('stopped');
  });
});
