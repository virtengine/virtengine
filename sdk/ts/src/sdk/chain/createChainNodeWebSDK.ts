import { createSDK as createCosmosSDK } from "../../generated/createCosmosSDK.ts";
import { createSDK as createNodeSDK } from "../../generated/createNodeSDK.ts";
import { patches as cosmosPatches } from "../../generated/patches/cosmosCustomTypePatches.ts";
import { patches as nodePatches } from "../../generated/patches/nodeCustomTypePatches.ts";
import { getMessageType } from "../getMessageType.ts";
import { createGrpcGatewayTransport } from "../transport/grpc-gateway/createGrpcGatewayTransport.ts";
import { getRetryInterceptors, type RetryOptions } from "../transport/interceptors/retry.ts";
import type { TxClient } from "../transport/tx/TxClient.ts";
import { ChainCapabilityController } from "./ChainCapability.ts";
import { createCapabilityQueryTransport, createCapabilityTxTransport } from "./createCapabilityTransports.ts";

export type { PayloadOf, ResponseOf } from "../types.ts";

export function createChainNodeWebSDK(options: ChainNodeWebSDKOptions) {
  const { retry: retryOptions, ...transportOptions } = options.query.transportOptions ?? {};
  const capability = new ChainCapabilityController(options.tx?.signer);
  const queryTransport = createCapabilityQueryTransport(capability, createGrpcGatewayTransport({
    ...transportOptions,
    baseUrl: options.query.baseUrl,
    interceptors: getRetryInterceptors(retryOptions),
  }));
  const txTransport = createCapabilityTxTransport(capability, getMessageType);
  const nodeSDK = createNodeSDK(queryTransport, txTransport, {
    clientOptions: { typePatches: { ...cosmosPatches, ...nodePatches } },
  });
  const cosmosSDK = createCosmosSDK(queryTransport, txTransport, {
    clientOptions: { typePatches: cosmosPatches },
  });
  return { ...nodeSDK, ...cosmosSDK, capability };
}

export interface ChainNodeWebSDKOptions {
  query: {
    /**
     * Blockchain gRPC gateway endpoint (also known as REST endpoint)
     */
    baseUrl: string;

    transportOptions?: {
      retry?: RetryOptions;
    };
  };
  tx?: {
    signer: TxClient;
  };
}
