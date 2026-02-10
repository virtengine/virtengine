import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useIdentityStore } from '@/stores/identityStore';
import type { WalletSigner } from '@/lib/api/chain';
import { fetchChainJsonWithFallback, signAndBroadcastAmino } from '@/lib/api/chain';

vi.mock('@/lib/api/chain', async () => {
  const actual = await vi.importActual<typeof import('@/lib/api/chain')>('@/lib/api/chain');
  return {
    ...actual,
    fetchChainJsonWithFallback: vi.fn(),
    signAndBroadcastAmino: vi.fn(),
  };
});

const fetchChainMock = fetchChainJsonWithFallback as unknown as vi.MockedFunction<
  typeof fetchChainJsonWithFallback
>;
const signMock = signAndBroadcastAmino as unknown as vi.MockedFunction<
  typeof signAndBroadcastAmino
>;

const initialState = useIdentityStore.getState();

describe('identityStore', () => {
  beforeEach(() => {
    useIdentityStore.setState(initialState, true);
    fetchChainMock.mockReset();
    signMock.mockReset();
  });

  it('loads identity record from chain', async () => {
    fetchChainMock.mockResolvedValue({
      identity_record: {
        score: 72,
        status: 'verified',
        scopes: ['kyc'],
        updated_at: '2024-01-01T00:00:00Z',
      },
    });

    await useIdentityStore.getState().fetchIdentity('ve1owner');

    const state = useIdentityStore.getState();
    expect(state.veidScore).toBe(72);
    expect(state.isVerified).toBe(true);
    expect(state.scopes).toHaveLength(1);
  });

  it('requests verification via signing', async () => {
    const wallet: WalletSigner = {
      status: 'connected',
      chainId: 'virtengine-1',
      accounts: [{ address: 've1owner' }],
      activeAccountIndex: 0,
      signAmino: vi.fn(),
      estimateFee: vi
        .fn()
        .mockReturnValue({ amount: [{ denom: 'uve', amount: '1' }], gas: '200000' }),
    };
    signMock.mockResolvedValue({
      txHash: 'txhash',
      code: 0,
      rawLog: '',
      gasUsed: 100,
      gasWanted: 200,
    });

    await useIdentityStore.getState().requestVerification('ve1owner', ['kyc'], wallet);

    const state = useIdentityStore.getState();
    expect(state.status).toBe('pending');
  });
});
