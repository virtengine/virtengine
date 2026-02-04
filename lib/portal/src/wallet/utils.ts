import type { WalletChainConfig } from './types';

export interface KeplrChainInfo {
  chainId: string;
  chainName: string;
  rpc: string;
  rest: string;
  bip44: {
    coinType: number;
  };
  bech32Config: {
    bech32PrefixAccAddr: string;
    bech32PrefixAccPub: string;
    bech32PrefixValAddr: string;
    bech32PrefixValPub: string;
    bech32PrefixConsAddr: string;
    bech32PrefixConsPub: string;
  };
  currencies: Array<{
    coinDenom: string;
    coinMinimalDenom: string;
    coinDecimals: number;
    coinGeckoId?: string;
    coinImageUrl?: string;
  }>;
  feeCurrencies: Array<{
    coinDenom: string;
    coinMinimalDenom: string;
    coinDecimals: number;
    gasPriceStep?: {
      low: number;
      average: number;
      high: number;
    };
    coinGeckoId?: string;
    coinImageUrl?: string;
  }>;
  stakeCurrency: {
    coinDenom: string;
    coinMinimalDenom: string;
    coinDecimals: number;
    coinGeckoId?: string;
    coinImageUrl?: string;
  };
  features?: string[];
}

export function toKeplrChainInfo(chain: WalletChainConfig): KeplrChainInfo {
  const prefix = chain.bech32Prefix;
  return {
    chainId: chain.chainId,
    chainName: chain.chainName,
    rpc: chain.rpcEndpoint,
    rest: chain.restEndpoint,
    bip44: { coinType: chain.slip44 ?? 118 },
    bech32Config: {
      bech32PrefixAccAddr: prefix,
      bech32PrefixAccPub: `${prefix}pub`,
      bech32PrefixValAddr: `${prefix}valoper`,
      bech32PrefixValPub: `${prefix}valoperpub`,
      bech32PrefixConsAddr: `${prefix}valcons`,
      bech32PrefixConsPub: `${prefix}valconspub`,
    },
    currencies: chain.currencies,
    feeCurrencies: chain.feeCurrencies,
    stakeCurrency: chain.stakeCurrency,
    features: chain.features,
  };
}

export function isBrowser(): boolean {
  return typeof window !== 'undefined';
}

export function toBase64(bytes: Uint8Array): string {
  if (typeof Buffer !== 'undefined') {
    return Buffer.from(bytes).toString('base64');
  }
  let binary = '';
  for (const byte of bytes) {
    binary += String.fromCharCode(byte);
  }
  return btoa(binary);
}

export function fromBase64(value: string): Uint8Array {
  if (typeof Buffer !== 'undefined') {
    return Uint8Array.from(Buffer.from(value, 'base64'));
  }
  const binary = atob(value);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}
