import { createSDK } from "../../generated/createProviderSDK.ts";
import type { GrpcTransportOptions } from "../transport/grpc/createGrpcTransport.ts";
import { createGrpcTransport } from "../transport/grpc/createGrpcTransport.ts";
import { getRetryInterceptors, type RetryOptions } from "../transport/interceptors/retry.ts";

export type { PayloadOf, ResponseOf } from "../types.ts";

export type ProviderSDK = ReturnType<typeof createSDK>;

export function createProviderSDK(options: ProviderSDKOptions): ProviderSDK {
  const { retry: retryOptions, ...transportOptions } = options.transportOptions ?? {};
  const certificateOptions = options.authentication?.type === "mtls"
    ? {
        cert: options.authentication?.cert,
        key: options.authentication?.key,
      }
    : null;

  return createSDK(
    createGrpcTransport({
      ...transportOptions,
      interceptors: getRetryInterceptors(retryOptions),
      baseUrl: options.baseUrl,
      nodeOptions: {
        ...certificateOptions,
        rejectUnauthorized: false,
      },
    }),
  );
}

export interface ProviderSDKOptions {
  /**
   * Provider gRPC endpoint
   */
  baseUrl: string;

  /**
   * Authentication options
   */
  authentication?: {
    type: "mtls";
    cert: string;
    key: string;
  };

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
}
