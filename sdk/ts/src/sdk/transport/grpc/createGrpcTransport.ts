import { createContextValues } from "@connectrpc/connect";
import {
  createAsyncIterable,
  pipe,
  pipeTo,
  transformCompressEnvelope,
  transformDecompressEnvelope,
  transformJoinEnvelopes,
  transformParseEnvelope,
  transformSerializeEnvelope,
  transformSplitEnvelope,
} from "@connectrpc/connect/protocol";
import { headerGrpcStatus, requestHeaderWithCompression, validateResponseWithCompression, validateTrailer } from "@connectrpc/connect/protocol-grpc";
import { createNodeHttpClient, type GrpcTransportOptions as ConnectGrpcTransportOptions, Http2SessionManager } from "@connectrpc/connect-node";

import type { MessageDesc, MessageInitShape, MessageShape, MethodDesc } from "../../client/types.ts";
import { runStreamingCall, runUnaryCall } from "../runCall.ts";
import { TransportError } from "../TransportError.ts";
import { coerceTimeoutMs, createMethodUrl, createSerialization } from "../transportUtils.ts";
import type { CallOptions, StreamRequest, StreamResponse, Transport, UnaryRequest, UnaryResponse } from "../types.ts";

export type GrpcCallOptions = Omit<CallOptions, "onHeader" | "onTrailer"> & {
  header?: HeadersInit;
};
export interface GrpcTransportOptions extends Omit<ConnectGrpcTransportOptions, "useBinaryFormat"> {
  httpClient?: ReturnType<typeof createNodeHttpClient>;
}

/**
 * At most, allow ~4GiB to be received or sent per message.
 * zlib used by Node.js caps maxOutputLength at this value. It also happens to
 * be the maximum theoretical message size supported by protobuf-es.
 */
const MAX_READ_MAX_BYTES = 0xffffffff;

/**
 * The default value for the compressMinBytes option. The CPU cost of compressing
 * very small messages usually isn't worth the small reduction in network I/O, so
 * the default value is 1 kibibyte.
 */
const DEFAULT_COMPRESS_MIN_BYTES = 1024;

export function createGrpcTransport(options: GrpcTransportOptions): Transport<GrpcCallOptions> {
  const useBinaryFormat = true;
  const httpClient = options.httpClient ?? createDefaultHttpClient(options);

  return {
    requiresTypePatching: true,
    async unary<I extends MessageDesc, O extends MessageDesc>(
      method: MethodDesc<"unary", I, O>,
      message: MessageInitShape<I>,
      callOptions?: GrpcCallOptions,
    ): Promise<UnaryResponse<I, O>> {
      const timeoutMs = coerceTimeoutMs(callOptions?.timeoutMs, options.defaultTimeoutMs);

      return await runUnaryCall<I, O>({
        interceptors: options.interceptors,
        signal: callOptions?.signal,
        timeoutMs,
        req: {
          service: method.parent,
          stream: false,
          method,
          requestMethod: "POST",
          url: createMethodUrl(options.baseUrl, method),
          header: requestHeaderWithCompression(
            useBinaryFormat,
            timeoutMs,
            callOptions?.header,
            options.acceptCompression ?? [],
            options.sendCompression || null,
          ),
          contextValues: callOptions?.contextValues ?? createContextValues(),
          message,
        },
        next: async (req: UnaryRequest<I, O>): Promise<UnaryResponse<I, O>> => {
          const uRes = await httpClient({
            url: req.url,
            method: "POST",
            header: req.header ?? new Headers(),
            signal: req.signal,
            body: pipe(
              createAsyncIterable([req.message]),
              transformSerializeEnvelope(createSerialization(req.method.input)),
              transformCompressEnvelope(
                options.sendCompression || null,
                options.compressMinBytes || DEFAULT_COMPRESS_MIN_BYTES,
              ),
              transformJoinEnvelopes(),
              {
                propagateDownStreamError: true,
              },
            ),
          });
          const { compression, headerError } = validateResponseWithCompression(
            options.acceptCompression || [],
            uRes.status,
            uRes.header,
          );
          const message = await pipeTo(
            uRes.body,
            transformSplitEnvelope(options.readMaxBytes || MAX_READ_MAX_BYTES),
            transformDecompressEnvelope(compression ?? null, options.readMaxBytes || MAX_READ_MAX_BYTES),
            transformParseEnvelope<MessageShape<O>>(createSerialization(req.method.output)),
            async (iterable) => {
              let message: MessageShape<O> | undefined;
              for await (const chunk of iterable) {
                if (message !== undefined) {
                  throw new TransportError(
                    "protocol error: received extra output message for unary method",
                    TransportError.Code.Unimplemented,
                  );
                }
                message = chunk;
              }
              return message;
            },
            { propagateDownStreamError: false },
          );
          validateTrailer(uRes.trailer, uRes.header);
          if (message === undefined) {
            // Trailers only response
            if (headerError) {
              throw TransportError.from(headerError);
            }
            throw new TransportError(
              "protocol error: missing output message for unary method",
              uRes.trailer.has(headerGrpcStatus)
                ? TransportError.Code.Unimplemented
                : TransportError.Code.Unknown,
            );
          }
          if (headerError) {
            throw new TransportError(
              "protocol error: received output message for unary method with error status",
              TransportError.Code.Unknown,
            );
          }
          return {
            stream: false,
            method,
            header: uRes.header,
            trailer: uRes.trailer,
            message,
          } satisfies UnaryResponse<I, O>;
        },
      });
    },
    async stream<I extends MessageDesc, O extends MessageDesc>(
      method: MethodDesc<"server_streaming" | "client_streaming" | "bidi_streaming", I, O>,
      input: AsyncIterable<MessageInitShape<I>>,
      callOptions?: GrpcCallOptions,
    ): Promise<StreamResponse<I, O>> {
      const timeoutMs = coerceTimeoutMs(callOptions?.timeoutMs, options.defaultTimeoutMs);
      return runStreamingCall<I, O>({
        interceptors: options.interceptors,
        signal: callOptions?.signal,
        timeoutMs,
        req: {
          stream: true,
          service: method.parent,
          method,
          requestMethod: "POST",
          url: createMethodUrl(options.baseUrl, method),
          header: requestHeaderWithCompression(
            useBinaryFormat,
            timeoutMs,
            callOptions?.header,
            options.acceptCompression || [],
            options.sendCompression || null,
          ),
          contextValues: callOptions?.contextValues ?? createContextValues(),
          message: input,
        },
        next: async (req: StreamRequest<I, O>) => {
          const uRes = await httpClient({
            url: req.url,
            method: "POST",
            header: req.header ?? new Headers(),
            signal: req.signal,
            body: pipe(
              req.message,
              transformSerializeEnvelope(createSerialization(req.method.input)),
              transformCompressEnvelope(
                options.sendCompression || null,
                options.compressMinBytes || DEFAULT_COMPRESS_MIN_BYTES,
              ),
              transformJoinEnvelopes(),
              { propagateDownStreamError: true },
            ),
          });
          const { compression, foundStatus, headerError }
            = validateResponseWithCompression(
              options.acceptCompression || [],
              uRes.status,
              uRes.header,
            );
          if (headerError) {
            throw TransportError.from(headerError);
          }
          return {
            ...req,
            header: uRes.header,
            trailer: uRes.trailer,
            message: pipe(
              uRes.body,
              transformSplitEnvelope(options.readMaxBytes || MAX_READ_MAX_BYTES),
              transformDecompressEnvelope(
                compression ?? null,
                options.readMaxBytes || MAX_READ_MAX_BYTES,
              ),
              transformParseEnvelope(createSerialization(req.method.output)),
              async function* (iterable) {
                yield * iterable;
                if (!foundStatus) {
                  validateTrailer(uRes.trailer, uRes.header);
                }
              },
              { propagateDownStreamError: true },
            ),
          } satisfies StreamResponse<I, O>;
        },
      });
    },
  };
}

function createDefaultHttpClient(options: GrpcTransportOptions) {
  const sessionManager = options.sessionManager ?? new Http2SessionManager(
    options.baseUrl,
    {
      pingIntervalMs: options.pingIntervalMs,
      pingIdleConnection: options.pingIdleConnection,
      pingTimeoutMs: options.pingTimeoutMs,
      idleConnectionTimeoutMs: options.idleConnectionTimeoutMs,
    },
    options.nodeOptions,
  );
  return createNodeHttpClient({
    httpVersion: "2",
    sessionProvider: () => sessionManager,
  });
}
