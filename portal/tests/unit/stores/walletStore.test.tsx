import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import type { ReactNode } from 'react';
import { WalletProvider, useWallet, type WalletProviderConfig } from '@/lib/portal-adapter';

const walletConfig: WalletProviderConfig = {
  chain: {
    chainId: 'virtengine-1',
    chainName: 'VirtEngine',
    rpcEndpoint: 'https://rpc.virtengine.com',
    restEndpoint: 'https://api.virtengine.com',
    bech32Prefix: 'virtengine',
    stakeCurrency: {
      coinDenom: 'VE',
      coinMinimalDenom: 'uve',
      coinDecimals: 6,
    },
    currencies: [
      {
        coinDenom: 'VE',
        coinMinimalDenom: 'uve',
        coinDecimals: 6,
      },
    ],
    feeCurrencies: [
      {
        coinDenom: 'VE',
        coinMinimalDenom: 'uve',
        coinDecimals: 6,
        gasPriceStep: { low: 0.01, average: 0.025, high: 0.04 },
      },
    ],
    features: ['cosmwasm'],
  },
  autoConnect: false,
};

function wrapper({ children }: { children: ReactNode }) {
  return <WalletProvider config={walletConfig}>{children}</WalletProvider>;
}

// Mock Keplr wallet in window
beforeEach(() => {
  // @ts-expect-error - mocking window.keplr
  window.keplr = {
    enable: vi.fn(),
    getKey: vi.fn().mockResolvedValue({
      bech32Address: 'virtengine1testaddress',
      pubKey: new Uint8Array([1, 2, 3]),
    }),
    signAmino: vi.fn(),
    signDirect: vi.fn(),
    experimentalSuggestChain: vi.fn(),
    getOfflineSignerAuto: vi.fn().mockResolvedValue({
      getAccounts: vi.fn().mockResolvedValue([
        { address: 'virtengine1testaddress' },
      ]),
    }),
  };
});

describe('wallet context', () => {
  it('should have initial state', () => {
    const { result } = renderHook(() => useWallet(), { wrapper });

    expect(result.current.state.isConnected).toBe(false);
    expect(result.current.state.isConnecting).toBe(false);
    expect(result.current.state.address).toBeNull();
    expect(result.current.state.walletType).toBeNull();
  });

  it('should connect wallet', async () => {
    const { result } = renderHook(() => useWallet(), { wrapper });

    await act(async () => {
      await result.current.actions.connect('keplr');
    });

    expect(result.current.state.isConnected).toBe(true);
    expect(result.current.state.walletType).toBe('keplr');
    expect(result.current.state.address).toBe('virtengine1testaddress');
  });

  it('should disconnect wallet', async () => {
    const { result } = renderHook(() => useWallet(), { wrapper });

    await act(async () => {
      await result.current.actions.connect('keplr');
    });

    expect(result.current.state.isConnected).toBe(true);

    await act(async () => {
      await result.current.actions.disconnect();
    });

    expect(result.current.state.isConnected).toBe(false);
    expect(result.current.state.address).toBeNull();
    expect(result.current.state.walletType).toBeNull();
  });
});
