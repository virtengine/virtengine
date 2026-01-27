import type { CallOptions, Transport } from "../transport/types.ts";
import type { TypePatches } from "./applyPatches.ts";
import { applyPatches } from "./applyPatches.ts";
import { createAsyncIterable, handleStreamResponse, mapStream } from "./stream.ts";
import type { MessageDesc, MessageInitShape, MessageShape, MethodDesc, ServiceDesc } from "./types.ts";

export type Client<Desc extends ServiceDesc, TCallOptions> = {
  [P in keyof Desc["methods"]]:
  Desc["methods"][P] extends MethodDesc<"server_streaming"> ? (input: MessageInitShape<Desc["methods"][P]["input"]>, options?: TCallOptions) => AsyncIterable<MessageShape<Desc["methods"][P]["output"]>>
    : Desc["methods"][P] extends MethodDesc<"client_streaming"> ? (input: AsyncIterable<MessageInitShape<Desc["methods"][P]["input"]>>, options?: TCallOptions) => Promise<MessageShape<Desc["methods"][P]["output"]>>
      : Desc["methods"][P] extends MethodDesc<"bidi_streaming"> ? (input: AsyncIterable<MessageInitShape<Desc["methods"][P]["input"]>>, options?: TCallOptions) => AsyncIterable<MessageShape<Desc["methods"][P]["output"]>>
        : Desc["methods"][P] extends MethodDesc<"unary"> | Omit<MethodDesc<"unary">, "kind"> ? (input: MessageInitShape<Desc["methods"][P]["input"]>, options?: TCallOptions) => Promise<MessageShape<Desc["methods"][P]["output"]>>
          : never;
};

const defaultEncoder: MethodOptions = {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  encode: (type: MessageDesc<unknown>, value: unknown) => type.fromPartial(value as any),
  decode: (_: MessageDesc<unknown>, value: unknown) => value,
};

export function createServiceClient<TSchema extends ServiceDesc, TCallOptions>(
  service: TSchema,
  transport: Transport<TCallOptions>,
  options?: ServiceClientOptions,
): Client<TSchema, TCallOptions> {
  const methodOptions: MethodOptions = transport.requiresTypePatching && options?.typePatches
    ? { encode: createEncodeWithPatches(options.typePatches), decode: createDecodeWithPatches(options.typePatches) }
    : defaultEncoder;
  const client: Record<string, ReturnType<typeof createMethod>> = {};
  const methodNames = Object.keys(service.methods);
  for (let i = 0; i < methodNames.length; i++) {
    const methodDesc = service.methods[methodNames[i]];
    client[methodNames[i]] = createMethod(methodDesc as MethodDesc, transport, methodOptions);
  }

  return client as Client<TSchema, TCallOptions>;
}

export interface ServiceClientOptions {
  typePatches?: TypePatches;
}

function createMethod(methodDesc: MethodDesc, transport: Transport, options: MethodOptions) {
  switch (methodDesc.kind) {
    case "server_streaming":
      return createServerStreamingFn(transport, methodDesc as MethodDesc<"server_streaming", MessageDesc, MessageDesc>, options);
    case "client_streaming":
      return createClientStreamingFn(transport, methodDesc as MethodDesc<"client_streaming", MessageDesc, MessageDesc>, options);
    case "bidi_streaming":
      return createBiDiStreamingFn(transport, methodDesc as MethodDesc<"bidi_streaming", MessageDesc, MessageDesc>, options);
    default:
      return createUnaryFn(transport, methodDesc as MethodDesc<"unary", MessageDesc, MessageDesc>, options);
  }
}

interface MethodOptions {
  encode(schema: MessageDesc<unknown>, value: unknown): unknown;
  decode(schema: MessageDesc<unknown>, value: unknown): unknown;
}

type UnaryFn<I extends MessageDesc<unknown>, O extends MessageDesc<unknown>> = (
  input: MessageInitShape<I>,
  options?: CallOptions,
) => Promise<MessageShape<O>>;

function createUnaryFn<I extends MessageDesc<unknown>, O extends MessageDesc<unknown>>(
  transport: Transport,
  method: MethodDesc<"unary", I, O>,
  methodOptions: MethodOptions,
): UnaryFn<I, O> {
  return async (input, options) => {
    const response = await transport.unary(
      method,
      methodOptions.encode(method.input, input) as MessageInitShape<I>,
      options,
    );
    options?.onHeader?.(response.header);
    options?.onTrailer?.(response.trailer);

    return methodOptions.decode(method.output, response.message) as MessageShape<O>;
  };
}

type ServerStreamingFn<I extends MessageDesc, O extends MessageDesc> = (
  input: MessageInitShape<I>,
  options?: CallOptions,
) => AsyncIterable<MessageShape<O>>;

function createServerStreamingFn<
  I extends MessageDesc,
  O extends MessageDesc,
>(
  transport: Transport,
  method: MethodDesc<"server_streaming", I, O>,
  methodOptions: MethodOptions,
): ServerStreamingFn<I, O> {
  return (input, options) => {
    return handleStreamResponse(
      method,
      transport.stream(
        method,
        createAsyncIterable([methodOptions.encode(method.input, input) as MessageInitShape<I>]),
        options,
      ),
      options,
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      methodOptions.decode as any,
    );
  };
}

type ClientStreamingFn<I extends MessageDesc, O extends MessageDesc> = (
  input: AsyncIterable<MessageInitShape<I>>,
  options?: CallOptions,
) => Promise<MessageShape<O>>;

function createClientStreamingFn<
  I extends MessageDesc,
  O extends MessageDesc,
>(
  transport: Transport,
  method: MethodDesc<"client_streaming", I, O>,
  methodOptions: MethodOptions,
): ClientStreamingFn<I, O> {
  return async (input, options) => {
    const response = await transport.stream(
      method,
      mapStream(input, (json) => methodOptions.encode(method.input, json) as MessageInitShape<I>),
      options,
    );
    options?.onHeader?.(response.header);
    let singleMessage: MessageShape<O> | undefined;
    let count = 0;
    for await (const message of response.message) {
      singleMessage = message;
      count++;
    }
    if (!singleMessage) {
      throw new Error("VirtEngine SDK protocol error: missing response message");
    }
    if (count > 1) {
      throw new Error("VirtEngine SDK protocol error: received extra messages for client streaming method");
    }
    options?.onTrailer?.(response.trailer);
    return methodOptions.decode(method.output, singleMessage) as MessageShape<O>;
  };
}

type BiDiStreamingFn<I extends MessageDesc, O extends MessageDesc> = (
  input: AsyncIterable<MessageInitShape<I>>,
  options?: CallOptions,
) => AsyncIterable<MessageShape<O>>;

function createBiDiStreamingFn<
  I extends MessageDesc,
  O extends MessageDesc,
>(
  transport: Transport,
  method: MethodDesc<"bidi_streaming", I, O>,
  methodOptions: MethodOptions,
): BiDiStreamingFn<I, O> {
  return (input, options) => {
    return handleStreamResponse(
      method,
      transport.stream(
        method,
        mapStream(input, (json) => methodOptions.encode(method.input, json) as MessageInitShape<I>),
        options,
      ),
      options,
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      methodOptions.decode as any,
    );
  };
}

const PATCHED_FROM_JSON_CACHE = new Map<TypePatches, MethodOptions["encode"]>();
function createEncodeWithPatches(patches: TypePatches): MethodOptions["encode"] {
  if (PATCHED_FROM_JSON_CACHE.has(patches)) return PATCHED_FROM_JSON_CACHE.get(patches)!;

  const encode: MethodOptions["encode"] = (messageDesc, value) => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return applyPatches("encode", messageDesc, messageDesc.fromPartial(value as any), patches);
  };
  PATCHED_FROM_JSON_CACHE.set(patches, encode);
  return encode;
}

const PATCHED_TO_JSON_CACHE = new Map<TypePatches, MethodOptions["decode"]>();
function createDecodeWithPatches(patches: TypePatches): MethodOptions["decode"] {
  if (PATCHED_TO_JSON_CACHE.has(patches)) return PATCHED_TO_JSON_CACHE.get(patches)!;

  const decode: MethodOptions["decode"] = (schema, message) => {
    return applyPatches("decode", schema, message, patches);
  };
  PATCHED_TO_JSON_CACHE.set(patches, decode);
  return decode;
}
