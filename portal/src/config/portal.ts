/**
 * Portal Configuration
 * Integrates lib/portal configuration with Next.js environment
 */

import type { PortalConfig, ChainConfig, WalletProviderConfig } from '@/lib/portal-adapter';
import { env } from './env';
import { getChainInfo } from './chains';

/**
 * Create portal configuration from environment
 */
export function createPortalConfig(): PortalConfig {
  const chainInfo = getChainInfo();

  return {
    chainEndpoint: chainInfo.wsEndpoint,
    chainRestEndpoint: chainInfo.restEndpoint,
    chainId: chainInfo.chainId,
    networkName: chainInfo.chainName,
    enableSSO: false,
    debug: env.isDev,
    sessionConfig: {
      tokenLifetimeSeconds: 3600,
      refreshThresholdSeconds: 300,
      autoRefresh: true,
      cookieName: 've_session',
      secureCookies: env.isProd,
    },
  };
}

/**
 * Create chain configuration from environment
 */
export function createChainConfig(): ChainConfig {
  const chainInfo = getChainInfo();

  return {
    wsEndpoint: chainInfo.wsEndpoint,
    restEndpoint: chainInfo.restEndpoint,
    chainId: chainInfo.chainId,
    autoReconnect: true,
    reconnectDelayMs: 1000,
    maxReconnectAttempts: 10,
    heartbeatIntervalMs: 30000,
    requestTimeoutMs: 30000,
  };
}

/**
 * Create wallet configuration from environment
 */
export function createWalletConfig(): WalletProviderConfig {
  const chainInfo = getChainInfo();

  return {
    chainInfo: {
      chainId: chainInfo.chainId,
      chainName: chainInfo.chainName,
      rpcEndpoint: chainInfo.rpcEndpoint,
      restEndpoint: chainInfo.restEndpoint,
      bech32Config: chainInfo.bech32Config,
      bip44: chainInfo.bip44,
      stakeCurrency: chainInfo.stakeCurrency,
      currencies: chainInfo.currencies,
      feeCurrencies: chainInfo.feeCurrencies,
      features: chainInfo.features,
    },
    autoConnect: true,
    walletConnectProjectId: env.walletConnectProjectId || undefined,
    metadata: env.walletConnectProjectId
      ? {
          name: 'VirtEngine Portal',
          description: 'VirtEngine Portal WalletConnect',
          url: env.appUrl,
          icons: [`${env.appUrl}/favicon.ico`],
        }
      : undefined,
  };
}

/**
 * Default portal config for use in providers
 */
export const portalConfig = createPortalConfig();

/**
 * Default chain config for use in providers
 */
export const chainConfig = createChainConfig();

/**
 * Default wallet config for use in providers
 */
export const walletConfig = createWalletConfig();
