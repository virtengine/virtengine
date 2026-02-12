/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Portal configuration helpers for chain + provider daemon endpoints.
 */

import { env } from '@/config/env';
import { getChainInfo } from '@/config/chains';
import {
  createChainConfig,
  createChainQueryClient,
  type ChainClientConfig,
  type ChainQueryClient,
} from 'virtengine-portal-lib';

export interface PortalEndpoints {
  chainRest: string;
  chainWs: string;
  chainRpc: string;
  providerDaemon: string | null;
}

let cachedChainClient: ChainQueryClient | null = null;

export function getPortalEndpoints(): PortalEndpoints {
  const chain = getChainInfo();
  return {
    chainRest: env.chainRest || chain.restEndpoint,
    chainWs: env.chainWs || chain.wsEndpoint,
    chainRpc: env.chainRpc || chain.rpcEndpoint,
    providerDaemon: env.providerDaemonUrl || null,
  };
}

export function getPortalChainConfig(): ChainClientConfig {
  const chain = getChainInfo();
  const endpoints = getPortalEndpoints();
  return createChainConfig({
    chainId: env.chainId || chain.chainId,
    rpcEndpoints: [endpoints.chainRpc],
    restEndpoints: [endpoints.chainRest],
    wsEndpoints: [endpoints.chainWs],
    bech32Prefix: chain.bech32Config.bech32PrefixAccAddr,
    feeDenom: chain.stakeCurrency.coinMinimalDenom,
    coinDecimals: chain.stakeCurrency.coinDecimals,
    gasPriceStep: chain.feeCurrencies?.[0]?.gasPriceStep,
  });
}

export function getPortalChainClient(): ChainQueryClient {
  if (!cachedChainClient) {
    cachedChainClient = createChainQueryClient(getPortalChainConfig());
  }
  return cachedChainClient;
}
