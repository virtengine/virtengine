/**
 * Portal Configuration
 * Integrates lib/portal configuration with Next.js environment
 */

import type { PortalConfig, ChainConfig } from '@/lib/portal-adapter';
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
 * Default portal config for use in providers
 */
export const portalConfig = createPortalConfig();

/**
 * Default chain config for use in providers
 */
export const chainConfig = createChainConfig();
