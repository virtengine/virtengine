import { createVirtEngineClient, type VirtEngineClient } from '@virtengine/chain-sdk';
import { getChainInfo } from '@/config/chains';

let chainClientPromise: Promise<VirtEngineClient> | null = null;

export function getChainClient(): Promise<VirtEngineClient> {
  if (!chainClientPromise) {
    const chain = getChainInfo();
    chainClientPromise = createVirtEngineClient({
      rpcEndpoint: chain.rpcEndpoint,
      restEndpoint: chain.restEndpoint,
      useWeb: true,
      enableCaching: true,
    });
  }

  return chainClientPromise;
}

export async function getChainSdk(): Promise<VirtEngineClient> {
  return getChainClient();
}
