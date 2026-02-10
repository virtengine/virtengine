import { createSDK as createCosmosSDK } from "../../generated/createCosmosSDK.ts";
import { createSDK as createNodeSDK } from "../../generated/createNodeSDK.ts";
import { patches as cosmosPatches } from "../../generated/patches/cosmosCustomTypePatches.ts";
import { patches as nodePatches } from "../../generated/patches/nodeCustomTypePatches.ts";
import { getMessageType } from "../getMessageType.ts";
import { createNoopTransport } from "../transport/createNoopTransport.ts";
import { createGrpcGatewayTransport } from "../transport/grpc-gateway/createGrpcGatewayTransport.ts";
import type { RetryOptions } from "../transport/interceptors/retry.ts";
import { createRetryInterceptor, isRetryEnabled } from "../transport/interceptors/retry.ts";
import { createTxTransport } from "../transport/tx/createTxTransport.ts";
import type { TxClient } from "../transport/tx/TxClient.ts";

export type { PayloadOf, ResponseOf } from "../types.ts";

export function createChainNodeWebSDK(options: ChainNodeWebSDKOptions) {
  const { retry: retryOptions, ...transportOptions } = options.query.transportOptions ?? {};
  const queryTransport = createGrpcGatewayTransport({
    ...transportOptions,
    baseUrl: options.query.baseUrl,
    interceptors: isRetryEnabled(retryOptions) ? [createRetryInterceptor(retryOptions)] : [],
  });
  const txTransport = options.tx
    ? createTxTransport({
        getMessageType,
        client: options.tx.signer,
      })
    : createNoopTransport({
        unaryErrorMessage: `Unable to sign transaction. "tx" option is not provided during chain SDK creation`,
      });
  const nodeSDK = createNodeSDK(queryTransport, txTransport, {
    clientOptions: { typePatches: { ...cosmosPatches, ...nodePatches } },
  });
  const cosmosSDK = createCosmosSDK(queryTransport, txTransport, {
    clientOptions: { typePatches: cosmosPatches },
  });
  return { ...nodeSDK, ...cosmosSDK };
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
