import type { CallOptions, StreamResponse } from "../transport/types.ts";
import type { MessageDesc, MessageShape, MethodDesc } from "./types.ts";

export { createAsyncIterable } from "@connectrpc/connect/protocol";

export function handleStreamResponse<I extends MessageDesc, O extends MessageDesc>(
  method: MethodDesc<"server_streaming" | "client_streaming" | "bidi_streaming", I, O>,
  stream: Promise<StreamResponse<I, O>>,
  options?: CallOptions,
  transform: (schema: MessageDesc, value: MessageShape<O>) => MessageShape<O> = (_, value) => value as MessageShape<O>,
): AsyncIterable<MessageShape<O>> {
  const it = (async function* () {
    const response = await stream;
    options?.onHeader?.(response.header);
    yield * response.message;
    options?.onTrailer?.(response.trailer);
  })();
  return mapStream(it, (value) => transform(method.output, value) as MessageShape<O>);
}

export function mapStream<T, U>(stream: AsyncIterable<T>, transform: (value: T) => U): AsyncIterable<U> {
  function mapIteratorResult(result: IteratorResult<T>) {
    if (result.done === true) {
      return result;
    }
    return {
      done: result.done,
      value: transform(result.value),
    };
  }

  return {
    [Symbol.asyncIterator]: () => {
      const it = stream[Symbol.asyncIterator]();
      const mappedIterator: AsyncIterator<U> = {
        next: () => it.next().then(mapIteratorResult),
      };

      if (it.throw !== undefined) {
        mappedIterator.throw = (e: unknown) =>
          (it as Required<typeof it>).throw(e).then(mapIteratorResult);
      }
      if (it.return !== undefined) {
        mappedIterator.return = (v: unknown) =>
          (it as Required<typeof it>).return(v).then(mapIteratorResult);
      }

      return mappedIterator;
    },
  };
}
