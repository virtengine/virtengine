import { TransportError } from "./TransportError.ts";
import type { Transport, TxCallOptions } from "./types.ts";

export function createNoopTransport(options: NoopTransportOptions): Transport<TxCallOptions> {
  return {
    async unary() {
      throw new TransportError(options.unaryErrorMessage, TransportError.Code.Unimplemented);
    },
    async stream() {
      throw new TransportError(options.streamErrorMessage || "Transaction transport doesn't support streaming", TransportError.Code.Unimplemented);
    },
  };
}

export interface NoopTransportOptions {
  unaryErrorMessage: string;
  streamErrorMessage?: string;
}
