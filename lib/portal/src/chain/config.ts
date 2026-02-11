/**
 * Chain endpoint configuration and defaults.
 */

import type {
  ChainClientConfig,
  ChainEndpointConfig,
  ChainEnvironment,
  ChainRetryConfig,
  ChainTimeoutConfig,
} from './types';

export const MAINNET_ENDPOINTS: ChainEndpointConfig = {
  name: 'VirtEngine Mainnet',
  chainId: 'virtengine-1',
  rpcEndpoints: ['https://rpc.virtengine.com'],
  restEndpoints: ['https://api.virtengine.com'],
  wsEndpoints: ['wss://ws.virtengine.com'],
  bech32Prefix: 'virtengine',
  feeDenom: 'uve',
  coinDecimals: 6,
  gasPriceStep: {
    low: 0.01,
    average: 0.025,
    high: 0.04,
  },
};

export const TESTNET_ENDPOINTS: ChainEndpointConfig = {
  name: 'VirtEngine Testnet',
  chainId: 'virtengine-testnet-1',
  rpcEndpoints: ['https://rpc.testnet.virtengine.com'],
  restEndpoints: ['https://api.testnet.virtengine.com'],
  wsEndpoints: ['wss://ws.testnet.virtengine.com'],
  bech32Prefix: 'virtengine',
  feeDenom: 'uve',
  coinDecimals: 6,
  gasPriceStep: {
    low: 0.01,
    average: 0.025,
    high: 0.04,
  },
};

export const LOCALNET_ENDPOINTS: ChainEndpointConfig = {
  name: 'VirtEngine Localnet',
  chainId: 'virtengine-localnet-1',
  rpcEndpoints: ['http://localhost:26657'],
  restEndpoints: ['http://localhost:1317'],
  wsEndpoints: ['ws://localhost:26657/websocket'],
  bech32Prefix: 'virtengine',
  feeDenom: 'uve',
  coinDecimals: 6,
  gasPriceStep: {
    low: 0.01,
    average: 0.025,
    high: 0.04,
  },
};

export const DEFAULT_TIMEOUTS: ChainTimeoutConfig = {
  requestMs: 15000,
  connectMs: 5000,
};

export const DEFAULT_RETRY: ChainRetryConfig = {
  maxRetries: 2,
  baseDelayMs: 400,
  maxDelayMs: 2000,
  jitterMs: 200,
  retryableStatusCodes: [408, 425, 429, 500, 502, 503, 504],
};

export interface ChainConfigOverrides {
  environment?: ChainEnvironment;
  chainId?: string;
  rpcEndpoints?: string[];
  restEndpoints?: string[];
  wsEndpoints?: string[];
  bech32Prefix?: string;
  feeDenom?: string;
  coinDecimals?: number;
  gasPriceStep?: ChainEndpointConfig['gasPriceStep'];
  timeouts?: Partial<ChainTimeoutConfig>;
  retry?: Partial<ChainRetryConfig>;
  fallbackEndpoints?: ChainEndpointConfig[];
  headers?: Record<string, string>;
  userAgent?: string;
}

export function resolveChainEnvironment(chainId?: string): ChainEnvironment {
  if (!chainId) return 'mainnet';
  const normalized = chainId.toLowerCase();
  if (normalized.includes('localnet')) return 'localnet';
  if (normalized.includes('devnet')) return 'localnet';
  if (normalized.includes('testnet')) return 'testnet';
  return 'mainnet';
}

export function defaultEndpointsForEnvironment(environment: ChainEnvironment): ChainEndpointConfig {
  switch (environment) {
    case 'localnet':
      return { ...LOCALNET_ENDPOINTS };
    case 'testnet':
      return { ...TESTNET_ENDPOINTS };
    case 'mainnet':
    default:
      return { ...MAINNET_ENDPOINTS };
  }
}

export function defaultFallbackEndpoints(environment: ChainEnvironment): ChainEndpointConfig[] {
  if (environment === 'localnet') {
    return [{ ...TESTNET_ENDPOINTS }, { ...MAINNET_ENDPOINTS }];
  }
  if (environment === 'testnet') {
    return [{ ...MAINNET_ENDPOINTS }];
  }
  return [];
}

/**
 * Builds a chain client configuration with sensible defaults.
 */
export function createChainConfig(overrides: ChainConfigOverrides = {}): ChainClientConfig {
  const environment = overrides.environment ?? resolveChainEnvironment(overrides.chainId);
  const base = defaultEndpointsForEnvironment(environment);

  const endpoints: ChainEndpointConfig = {
    ...base,
    chainId: overrides.chainId ?? base.chainId,
    rpcEndpoints: overrides.rpcEndpoints ?? base.rpcEndpoints,
    restEndpoints: overrides.restEndpoints ?? base.restEndpoints,
    wsEndpoints: overrides.wsEndpoints ?? base.wsEndpoints,
    bech32Prefix: overrides.bech32Prefix ?? base.bech32Prefix,
    feeDenom: overrides.feeDenom ?? base.feeDenom,
    coinDecimals: overrides.coinDecimals ?? base.coinDecimals,
    gasPriceStep: overrides.gasPriceStep ?? base.gasPriceStep,
  };

  return {
    environment,
    endpoints,
    fallbackEndpoints: overrides.fallbackEndpoints ?? defaultFallbackEndpoints(environment),
    timeouts: { ...DEFAULT_TIMEOUTS, ...overrides.timeouts },
    retry: { ...DEFAULT_RETRY, ...overrides.retry },
    headers: overrides.headers,
    userAgent: overrides.userAgent,
  };
}
