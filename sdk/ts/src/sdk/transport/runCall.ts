import type { Interceptor } from "@connectrpc/connect";
import { getAbortSignalReason } from "@connectrpc/connect/protocol";

import type { DeepPartial } from "../../encoding/typeEncodingHelpers.ts";
import { mapStream } from "../client/stream.ts";
import type { MessageDesc, MessageInitShape, MessageShape } from "../client/types.ts";
import { TransportError } from "./TransportError.ts";
import type { StreamRequest, StreamResponse, UnaryRequest, UnaryResponse } from "./types.ts";

/**
 * UnaryFn represents the client-side invocation of a unary RPC - a method
 * that takes a single input message, and responds with a single output
 * message.
 * A Transport implements such a function, and makes it available to
 * interceptors.
 */
type UnaryFn<
  I extends MessageDesc,
  O extends MessageDesc,
> = (req: UnaryRequest<I, O>) => Promise<UnaryResponse<I, O>>;

/**
 * Runs a unary method with the given interceptors. Note that this function
 * is only used when implementing a Transport.
 */
export function runUnaryCall<
  I extends MessageDesc,
  O extends MessageDesc,
>(options: {
  req: Omit<UnaryRequest<I, O>, "signal" | "message"> & {
    message: MessageInitShape<I>;
  };
  next: UnaryFn<I, O>;
  timeoutMs?: number;
  signal?: AbortSignal;
  interceptors?: Interceptor[];
}): Promise<UnaryResponse<I, O>> {
  const next = composeInterceptors(options.next, options.interceptors);
  const { signal, abort } = createAbortSignal(options);

  const req = {
    ...options.req,
    message: options.req.method.input.fromPartial(options.req.message as DeepPartial<I>) as MessageShape<I>,
    signal,
  };
  return next(req).catch(abort);
}

/**
 * StreamingFn represents the client-side invocation of a streaming RPC - a
 * method that takes zero or more input messages, and responds with zero or
 * more output messages.
 * A Transport implements such a function, and makes it available to
 * interceptors.
 */
type StreamingFn<
  I extends MessageDesc = MessageDesc,
  O extends MessageDesc = MessageDesc,
> = (req: StreamRequest<I, O>) => Promise<StreamResponse<I, O>>;

/**
 * Runs a server-streaming method with the given interceptors. Note that this
 * function is only used when implementing a Transport.
 */
export function runStreamingCall<
  I extends MessageDesc,
  O extends MessageDesc,
>(options: {
  req: Omit<StreamRequest<I, O>, "signal" | "message"> & {
    message: AsyncIterable<MessageInitShape<I>>;
  };
  next: StreamingFn<I, O>;
  timeoutMs?: number;
  signal?: AbortSignal;
  interceptors?: Interceptor[];
}): Promise<StreamResponse<I, O>> {
  const next = composeInterceptors(options.next, options.interceptors);
  const { signal, abort } = createAbortSignal(options);
  const req = {
    ...options.req,
    message: mapStream(options.req.message, (message) => options.req.method.input.fromPartial(message as DeepPartial<I>) as MessageShape<I>),
    signal,
  };
  let doneCalled = false;
  // Call return on the request iterable to indicate
  // that we will no longer consume it and it should
  // cleanup any allocated resources.
  signal.addEventListener("abort", function () {
    const it = options.req.message[Symbol.asyncIterator]();
    // If the signal is aborted due to an error, we want to throw
    // the error to the request iterator.
    if (!doneCalled) {
      it.throw?.(this.reason).catch(() => {
        // throw returns a promise, which we don't care about.
        //
        // Uncaught promises are thrown at sometime/somewhere by the event loop,
        // this is to ensure error is caught and ignored.
      });
    }
    it.return?.().catch(() => {
      // return returns a promise, which we don't care about.
      //
      // Uncaught promises are thrown at sometime/somewhere by the event loop,
      // this is to ensure error is caught and ignored.
    });
  });
  return next(req).then((res) => {
    return {
      ...res,
      message: {
        [Symbol.asyncIterator]() {
          const it = res.message[Symbol.asyncIterator]();
          return {
            next() {
              return it.next().then((r) => {
                if (r.done == true) {
                  doneCalled = true;
                }
                return r;
              }, abort);
            },
            // We deliberately omit throw/return.
          };
        },
      },
    };
  }, abort);
}

function createAbortSignal(options: {
  timeoutMs?: number;
  signal?: AbortSignal;
}) {
  const controller = new AbortController();
  const signals: AbortSignal[] = [controller.signal];
  let timeoutSignal: AbortSignal | undefined;
  if (options.timeoutMs !== undefined) {
    timeoutSignal = AbortSignal.timeout(options.timeoutMs);
    signals.push(timeoutSignal);
  }
  if (options.signal !== undefined) {
    signals.push(options.signal);
  }
  const signal = AbortSignal.any(signals);
  return {
    signal,
    abort(reason: unknown): Promise<never> {
      // We peek at the deadline signal because fetch() will throw an error on
      // abort that discards the signal reason.
      const error = timeoutSignal?.aborted
        ? TransportError.from(getAbortSignalReason(timeoutSignal), TransportError.Code.DeadlineExceeded)
        : TransportError.from(reason);
      controller.abort(error);
      return Promise.reject(error);
    },
  };
}

function composeInterceptors<T>(
  next: T,
  interceptors: Interceptor[] | undefined,
): T {
  if (!interceptors) return next;

  let i = interceptors.length;
  while (i--) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    next = interceptors[i](next as (() => any)) as T;
  }
  return next;
}
