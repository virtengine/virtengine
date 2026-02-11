/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Portal configuration helpers for chain + provider daemon endpoints.
 */

import { env } from '@/config/env';
import { getChainInfo } from '@/config/chains';

export interface PortalEndpoints {
  chainRest: string;
  chainWs: string;
  chainRpc: string;
  providerDaemon: string | null;
}

export function getPortalEndpoints(): PortalEndpoints {
  const chain = getChainInfo();
  return {
    chainRest: env.chainRest || chain.restEndpoint,
    chainWs: env.chainWs || chain.wsEndpoint,
    chainRpc: env.chainRpc || chain.rpcEndpoint,
    providerDaemon: env.providerDaemonUrl || null,
  };
}
