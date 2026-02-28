import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useWalletTransaction } from '@/hooks/useWalletTransaction';

const walletState = {
  status: 'connected' as const,
  chainId: 'virtengine-1',
  accounts: [{ address: 'virtengine1testaddress' }],
  activeAccountIndex: 0,
  estimateFee: vi.fn(() => ({
    amount: [{ denom: 'uve', amount: '5000' }],
    gas: '200000',
  })),
  signAmino: vi.fn(),
};

const signAndBroadcastAmino = vi.fn();

vi.mock('@/lib/portal-adapter', () => ({
  useWallet: () => walletState,
}));

vi.mock('@/lib/api/chain', () => ({
  signAndBroadcastAmino: (...args: unknown[]) => signAndBroadcastAmino(...args),
}));

describe('useWalletTransaction', () => {
  beforeEach(() => {
    walletState.status = 'connected';
    walletState.chainId = 'virtengine-1';
    walletState.accounts = [{ address: 'virtengine1testaddress' }];
    walletState.activeAccountIndex = 0;
    walletState.estimateFee.mockClear();
    walletState.signAmino.mockClear();
    signAndBroadcastAmino.mockReset();
  });

  it('broadcasts typed messages instead of fabricating a local success result', async () => {
    signAndBroadcastAmino.mockResolvedValue({
      txHash: 'ABC123',
      code: 0,
      rawLog: 'accepted',
      gasUsed: 180000,
      gasWanted: 200000,
    });

    const { result } = renderHook(() => useWalletTransaction());

    let txResult: Awaited<ReturnType<typeof result.current.sendTransaction>> | undefined;
    await act(async () => {
      txResult = await result.current.sendTransaction([
        {
          typeUrl: '/virtengine.market.v1beta5.MsgCreateOrder',
          value: { owner: 'virtengine1testaddress' },
        },
      ]);
    });

    expect(signAndBroadcastAmino).toHaveBeenCalledWith(
      expect.objectContaining({ chainId: 'virtengine-1' }),
      [
        {
          typeUrl: '/virtengine.market.v1beta5.MsgCreateOrder',
          value: { owner: 'virtengine1testaddress' },
        },
      ],
      '',
      260000
    );
    expect(txResult).toEqual({
      txHash: 'ABC123',
      blockHeight: null,
      code: 0,
      rawLog: 'accepted',
      gasUsed: 180000,
      gasWanted: 200000,
    });
  });

  it('rejects messages without a type URL', async () => {
    const { result } = renderHook(() => useWalletTransaction());

    await expect(
      act(async () => {
        await result.current.sendTransaction([{ value: { owner: 'virtengine1testaddress' } }]);
      })
    ).rejects.toMatchObject({
      code: 'TRANSACTION_FAILED',
      message: 'Transaction message is missing a type URL',
    });

    expect(signAndBroadcastAmino).not.toHaveBeenCalled();
  });
});
