import { createSDK } from "../../generated/createProviderSDK.ts";
import type { MessageDesc, MessageInitShape, MethodDesc } from "../client/types.ts";
import { createGrpcGatewayTransport, type GrpcGatewayCallOptions, type GrpcGatewayTransportOptions } from "../transport/grpc-gateway/createGrpcGatewayTransport.ts";
import { getRetryInterceptors, type RetryOptions } from "../transport/interceptors/retry.ts";
import { TransportError } from "../transport/TransportError.ts";
import type { StreamResponse, Transport, UnaryResponse } from "../transport/types.ts";

export type { PayloadOf, ResponseOf } from "../types.ts";

export type ProviderWebSDK = ReturnType<typeof createSDK>;

export function createProviderWebSDK(options: ProviderWebSDKOptions): ProviderWebSDK {
  const { retry: retryOptions, ...transportOptions } = options.transportOptions ?? {};

  return createSDK(
    createProviderWebTransport({
      ...transportOptions,
      baseUrl: options.baseUrl,
      interceptors: getRetryInterceptors(retryOptions),
    }),
  );
}

function createProviderWebTransport(options: GrpcGatewayTransportOptions): Transport<GrpcGatewayCallOptions> {
  const transport = createGrpcGatewayTransport(options);

  return {
    async unary<I extends MessageDesc, O extends MessageDesc>(
      method: MethodDesc<"unary", I, O>,
      message: MessageInitShape<I>,
      callOptions?: GrpcGatewayCallOptions,
    ): Promise<UnaryResponse<I, O>> {
      if (!method.httpPath) {
        throw createUnsupportedProviderWebMethodError(
          method,
          "this RPC does not expose an HTTP gateway route. Use createProviderSDK() for direct gRPC access.",
        );
      }

      return transport.unary(method, message, callOptions);
    },
    async stream<I extends MessageDesc, O extends MessageDesc>(
      method: MethodDesc<"server_streaming" | "client_streaming" | "bidi_streaming", I, O>,
      _input: AsyncIterable<MessageInitShape<I>>,
      _callOptions?: GrpcGatewayCallOptions,
    ): Promise<StreamResponse<I, O>> {
      throw createUnsupportedProviderWebMethodError(
        method,
        "streaming provider methods are not available over the browser transport. Use createProviderSDK() for gRPC streaming access.",
      );
    },
  };
}

function createUnsupportedProviderWebMethodError(
  method: { parent: { typeName: string }; name: string },
  reason: string,
) {
  return new TransportError(
    `Provider web SDK cannot call ${method.parent.typeName}.${method.name}: ${reason}`,
    TransportError.Code.Unimplemented,
  );
}

export interface ProviderWebSDKOptions {
  /**
   * Provider HTTP gateway endpoint (also known as REST endpoint)
   */
  baseUrl: string;

  /**
   * Options for the gRPC gateway transport
   */
  transportOptions?: Omit<GrpcGatewayTransportOptions, "baseUrl" | "interceptors"> & {
    retry?: RetryOptions;
  };
}
