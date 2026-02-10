/**
 * Chain configuration for VirtEngine
 */

export interface ChainInfo {
  chainId: string;
  chainName: string;
  rpcEndpoint: string;
  restEndpoint: string;
  wsEndpoint: string;
  explorerUrl: string;
  bip44: {
    coinType: number;
  };
  stakeCurrency: {
    coinDenom: string;
    coinMinimalDenom: string;
    coinDecimals: number;
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
  }>;
  feeCurrencies: Array<{
    coinDenom: string;
    coinMinimalDenom: string;
    coinDecimals: number;
    gasPriceStep: {
      low: number;
      average: number;
      high: number;
    };
  }>;
  features: string[];
}

export const MAINNET_CHAIN: ChainInfo = {
  chainId: 'virtengine-1',
  chainName: 'VirtEngine',
  rpcEndpoint: 'https://rpc.virtengine.com',
  restEndpoint: 'https://api.virtengine.com',
  wsEndpoint: 'wss://ws.virtengine.com',
  explorerUrl: 'https://explorer.virtengine.io',
  bip44: {
    coinType: 118,
  },
  stakeCurrency: {
    coinDenom: 'VE',
    coinMinimalDenom: 'uve',
    coinDecimals: 6,
  },
  bech32Config: {
    bech32PrefixAccAddr: 'virtengine',
    bech32PrefixAccPub: 'virtuenginepub',
    bech32PrefixValAddr: 'virtenginevaloper',
    bech32PrefixValPub: 'virtenginevaloperpub',
    bech32PrefixConsAddr: 'virtenginevalcons',
    bech32PrefixConsPub: 'virtenginevalconspub',
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
      gasPriceStep: {
        low: 0.01,
        average: 0.025,
        high: 0.04,
      },
    },
  ],
  features: ['cosmwasm', 'ibc-transfer', 'ibc-go'],
};

export const TESTNET_CHAIN: ChainInfo = {
  chainId: 'virtengine-testnet-1',
  chainName: 'VirtEngine Testnet',
  rpcEndpoint: 'https://rpc.testnet.virtengine.com',
  restEndpoint: 'https://api.testnet.virtengine.com',
  wsEndpoint: 'wss://ws.testnet.virtengine.com',
  explorerUrl: 'https://testnet.explorer.virtengine.io',
  bip44: {
    coinType: 118,
  },
  stakeCurrency: {
    coinDenom: 'VE',
    coinMinimalDenom: 'uve',
    coinDecimals: 6,
  },
  bech32Config: {
    bech32PrefixAccAddr: 'virtengine',
    bech32PrefixAccPub: 'virtuenginepub',
    bech32PrefixValAddr: 'virtenginevaloper',
    bech32PrefixValPub: 'virtenginevaloperpub',
    bech32PrefixConsAddr: 'virtenginevalcons',
    bech32PrefixConsPub: 'virtenginevalconspub',
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
      gasPriceStep: {
        low: 0.01,
        average: 0.025,
        high: 0.04,
      },
    },
  ],
  features: ['cosmwasm', 'ibc-transfer', 'ibc-go'],
};

export const DEVNET_CHAIN: ChainInfo = {
  ...TESTNET_CHAIN,
  chainId: 'virtengine-devnet-1',
  chainName: 'VirtEngine Devnet',
  rpcEndpoint: 'http://localhost:26657',
  restEndpoint: 'http://localhost:1317',
  wsEndpoint: 'ws://localhost:26657/websocket',
  explorerUrl: 'http://localhost:5173',
};

export const LOCALNET_CHAIN: ChainInfo = {
  ...DEVNET_CHAIN,
  chainId: 'virtengine-localnet-1',
  chainName: 'VirtEngine Localnet',
  explorerUrl: 'http://localhost:5173',
};

export function getChainInfo(): ChainInfo {
  const chainId =
    process.env.NEXT_PUBLIC_CHAIN_ID ??
    (process.env.NODE_ENV === 'development' ? 'virtengine-localnet-1' : 'virtengine-1');

  if (chainId.includes('testnet')) {
    return TESTNET_CHAIN;
  }

  if (chainId.includes('devnet')) {
    return DEVNET_CHAIN;
  }

  if (chainId.includes('localnet')) {
    return LOCALNET_CHAIN;
  }

  return MAINNET_CHAIN;
}

export function isTestnet(): boolean {
  const chainId = process.env.NEXT_PUBLIC_CHAIN_ID ?? '';
  return chainId.includes('testnet') || chainId.includes('devnet') || chainId.includes('localnet');
}
