import type { CallOptions as ConnectCallOptions, ContextValues } from "@connectrpc/connect";
import type { DeliverTxResponse, StdFee } from "@cosmjs/stargate";

import type { MessageDesc, MessageInitShape, MessageShape, MethodDesc, ServiceDesc } from "../client/types.ts";
import type { TxRaw } from "./tx/TxClient.ts";

// eslint-disable-next-line @typescript-eslint/no-empty-object-type
export interface CallOptions extends ConnectCallOptions {}

export interface TxCallOptions {
  afterSign?: (tx: TxRaw) => void;
  afterBroadcast?: (tx: DeliverTxResponse) => void;
  memo?: string;
  fee?: Partial<StdFee>;
}

export interface Transport<TCallOptions = unknown> {
  requiresTypePatching?: boolean;
  /**
   * Call a unary RPC - a method that takes a single input message, and
   * responds with a single output message.
   */
  unary<I extends MessageDesc, O extends MessageDesc>(method: MethodDesc<"unary", I, O>, input: MessageInitShape<I>, options?: TCallOptions): Promise<UnaryResponse<I, O>>;
  /**
   * Call a streaming RPC - a method that takes zero or more input messages,
   * and responds with zero or more output messages.
   */
  stream<I extends MessageDesc, O extends MessageDesc>(method: MethodDesc<"server_streaming" | "client_streaming" | "bidi_streaming", I, O>, input: AsyncIterable<MessageInitShape<I>>, options?: TCallOptions): Promise<StreamResponse<I, O>>;
}

export interface UnaryResponse<I extends MessageDesc, O extends MessageDesc> extends ResponseCommon {
  stream: false;
  message: MessageShape<O>;
  method: MethodDesc<"unary", I, O>;
}

export interface StreamResponse<I extends MessageDesc, O extends MessageDesc> extends ResponseCommon {
  stream: true;
  message: AsyncIterable<MessageShape<O>>;
  method: MethodDesc<"server_streaming" | "client_streaming" | "bidi_streaming", I, O>;
}

interface ResponseCommon {
  header: Headers;
  trailer: Headers;
}

export interface UnaryRequest<I extends MessageDesc, O extends MessageDesc> extends RequestCommon {
  stream: false;
  message: MessageShape<I>;
  method: MethodDesc<"unary", I, O>;
}

export interface StreamRequest<I extends MessageDesc, O extends MessageDesc> extends RequestCommon {
  stream: true;
  message: AsyncIterable<MessageShape<I>>;
  method: MethodDesc<"server_streaming" | "client_streaming" | "bidi_streaming", I, O>;
}

interface RequestCommon {
  /**
   * Metadata related to the service that is being called.
   */
  service: ServiceDesc;
  /**
   * HTTP method of the request. Server-side interceptors may use this value
   * to identify Connect GET requests.
   */
  requestMethod: string;
  /**
   * The URL the request is going to hit for the clients or the
   * URL received by the server.
   */
  url: string;
  /**
   * The AbortSignal for the current call.
   */
  signal: AbortSignal;
  /**
   * Headers that will be sent along with the request.
   */
  header?: Headers;
  /**
   * The context values for the current call.
   */
  contextValues?: ContextValues;
}
