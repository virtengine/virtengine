import { createSDK as createCosmosSDK } from "../../generated/createCosmosSDK.ts";
import { createSDK as createNodeSDK } from "../../generated/createNodeSDK.ts";
import { patches as cosmosPatches } from "../../generated/patches/cosmosCustomTypePatches.ts";
import { patches as nodePatches } from "../../generated/patches/nodeCustomTypePatches.ts";
import { getMessageType } from "../getMessageType.ts";
import { createNoopTransport } from "../transport/createNoopTransport.ts";
import type { GrpcTransportOptions } from "../transport/grpc/createGrpcTransport.ts";
import { createGrpcTransport } from "../transport/grpc/createGrpcTransport.ts";
import type { RetryOptions } from "../transport/interceptors/retry.ts";
import { createRetryInterceptor, isRetryEnabled } from "../transport/interceptors/retry.ts";
import { createTxTransport } from "../transport/tx/createTxTransport.ts";
import type { TxClient } from "../transport/tx/TxClient.ts";
import type { Transport, TxCallOptions } from "../transport/types.ts";

export type { PayloadOf, ResponseOf } from "../types.ts";

export function createChainNodeSDK(options: ChainNodeSDKOptions) {
  const { retry: retryOptions, ...transportOptions } = options.query.transportOptions ?? {};
  const queryTransport = createGrpcTransport({
    ...transportOptions,
    baseUrl: options.query.baseUrl,
    interceptors: isRetryEnabled(retryOptions) ? [createRetryInterceptor(retryOptions)] : [],
  });
  let txTransport: Transport<TxCallOptions>;

  if (options.tx) {
    txTransport = createTxTransport({
      getMessageType,
      client: options.tx.signer,
    });
  } else {
    txTransport = createNoopTransport({
      unaryErrorMessage: `Unable to sign transaction. "tx" option is not provided during chain SDK creation`,
    });
  }
  const nodeSDK = createNodeSDK(queryTransport, txTransport, {
    clientOptions: { typePatches: { ...cosmosPatches, ...nodePatches } },
  });
  const cosmosSDK = createCosmosSDK(queryTransport, txTransport, {
    clientOptions: { typePatches: cosmosPatches },
  });
  return { ...nodeSDK, ...cosmosSDK };
}

export interface ChainNodeSDKOptions {
  query: {
    /**
     * Blockchain gRPC endpoint
     */
    baseUrl: string;

    /**
     * Options for the gRPC transport
     */
    transportOptions?: {
      pingIdleConnection?: GrpcTransportOptions["pingIdleConnection"];
      pingIntervalMs?: GrpcTransportOptions["pingIntervalMs"];
      pingTimeoutMs?: GrpcTransportOptions["pingTimeoutMs"];
      idleConnectionTimeoutMs?: GrpcTransportOptions["idleConnectionTimeoutMs"];
      defaultTimeoutMs?: GrpcTransportOptions["defaultTimeoutMs"];
      retry?: RetryOptions;
    };
  };
  tx?: {
    signer: TxClient;
  };
}
