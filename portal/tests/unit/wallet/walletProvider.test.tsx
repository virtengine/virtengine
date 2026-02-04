import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { WalletProvider, useWallet, type WalletProviderConfig } from '@/lib/portal-adapter';
import type { WalletChainInfo } from '@/lib/portal-adapter';

const chainInfo: WalletChainInfo = {
  chainId: 'virtengine-1',
  chainName: 'VirtEngine',
  rpcEndpoint: 'https://rpc.virtengine.com',
  restEndpoint: 'https://api.virtengine.com',
  bech32Config: {
    bech32PrefixAccAddr: 'virtengine',
    bech32PrefixAccPub: 'virtenginepub',
    bech32PrefixValAddr: 'virtenginevaloper',
    bech32PrefixValPub: 'virtenginevaloperpub',
    bech32PrefixConsAddr: 'virtenginevalcons',
    bech32PrefixConsPub: 'virtenginevalconspub',
  },
  bip44: { coinType: 118 },
  stakeCurrency: {
    coinDenom: 'VE',
    coinMinimalDenom: 'uve',
    coinDecimals: 6,
  },
  currencies: [
    { coinDenom: 'VE', coinMinimalDenom: 'uve', coinDecimals: 6 },
  ],
  feeCurrencies: [
    { coinDenom: 'VE', coinMinimalDenom: 'uve', coinDecimals: 6, gasPriceStep: { low: 0.01, average: 0.025, high: 0.04 } },
  ],
  features: ['cosmwasm'],
};

const wrapper = ({ children }: { children: React.ReactNode }) => {
  const config: WalletProviderConfig = {
    chainInfo,
    autoConnect: false,
  };
  return <WalletProvider config={config}>{children}</WalletProvider>;
};

beforeEach(() => {
  // @ts-expect-error mock keplr
  window.keplr = {
    enable: vi.fn(),
    getKey: vi.fn().mockResolvedValue({
      bech32Address: 'virtengine1testaddress',
      pubKey: new Uint8Array([1, 2, 3]),
      algo: 'secp256k1',
    }),
    signAmino: vi.fn(),
    signDirect: vi.fn(),
  };

  // @ts-expect-error mock offline signer
  window.getOfflineSignerAuto = vi.fn().mockResolvedValue({
    getAccounts: vi.fn().mockResolvedValue([
      { address: 'virtengine1testaddress', algo: 'secp256k1', pubkey: new Uint8Array([1, 2, 3]) },
    ]),
  });
});

describe('WalletProvider', () => {
  it('initializes with idle status', () => {
    const { result } = renderHook(() => useWallet(), { wrapper });
    expect(result.current.status).toBe('idle');
    expect(result.current.accounts).toHaveLength(0);
  });

  it('connects and disconnects keplr wallet', async () => {
    const { result } = renderHook(() => useWallet(), { wrapper });

    await act(async () => {
      await result.current.connect('keplr');
    });

    expect(result.current.status).toBe('connected');
    expect(result.current.accounts).toHaveLength(1);

    await act(async () => {
      await result.current.disconnect();
    });

    expect(result.current.status).toBe('idle');
    expect(result.current.accounts).toHaveLength(0);
  });
});
