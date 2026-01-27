import type { Interceptor } from "@connectrpc/connect";
import {
  ExponentialBackoff,
  handleWhen,
  retry,
} from "cockatiel";

import { TransportError } from "../TransportError.ts";

const RETRIABLE_ERROR_CODES = new Set([
  TransportError.Code.Unavailable,
  TransportError.Code.DeadlineExceeded,
  TransportError.Code.Internal,
  TransportError.Code.Unknown,
]);
export function createRetryInterceptor(options: RetryOptions): Interceptor {
  const retryPolicy = retry(handleWhen((error) => error instanceof TransportError && RETRIABLE_ERROR_CODES.has(error.code)), {
    maxAttempts: Math.min(3, options.maxAttempts),
    backoff: new ExponentialBackoff({
      initialDelay: 250,
      maxDelay: options.maxDelayMs ?? 5_000,
    }),
  });

  return (next) => async (req) => retryPolicy.execute(() => next(req));
}

export function isRetryEnabled(options: RetryOptions | undefined): options is RetryOptions {
  return !!options?.maxAttempts && !Number.isNaN(options.maxAttempts) && options.maxAttempts > 0;
}

export interface RetryOptions {
  /**
   * Maximum number of attempts to make after first failure.
   * Maximum allowed value is 3.
   */
  maxAttempts: number;
  /**
   * Maximum delay between attempts in milliseconds. Used to restrict exponential backoff calculated value.
   * Default value is 5000ms.
   */
  maxDelayMs?: number;
}
