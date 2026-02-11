/**
 * Chain client shared types and errors.
 */

export type ChainEnvironment = 'mainnet' | 'testnet' | 'localnet' | 'custom';

export interface ChainEndpointConfig {
  /** Human-friendly network label. */
  name: string;
  /** Chain ID (e.g., virtengine-1). */
  chainId: string;
  /** RPC endpoints for Stargate/Tendermint. */
  rpcEndpoints: string[];
  /** REST API endpoints for gRPC gateway. */
  restEndpoints: string[];
  /** Optional websocket endpoints. */
  wsEndpoints?: string[];
  /** Bech32 prefix for addresses. */
  bech32Prefix: string;
  /** Fee denomination used for gas pricing. */
  feeDenom: string;
  /** Number of decimals in the fee denomination. */
  coinDecimals: number;
  /** Suggested gas price steps (low/avg/high). */
  gasPriceStep?: {
    low: number;
    average: number;
    high: number;
  };
}

export interface ChainTimeoutConfig {
  /** Timeout for HTTP requests in milliseconds. */
  requestMs: number;
  /** Timeout for RPC connection attempts in milliseconds. */
  connectMs: number;
}

export interface ChainRetryConfig {
  /** Maximum number of retries for transient errors. */
  maxRetries: number;
  /** Base delay for retry backoff in milliseconds. */
  baseDelayMs: number;
  /** Maximum delay for retry backoff in milliseconds. */
  maxDelayMs: number;
  /** Jitter added to retry delays in milliseconds. */
  jitterMs: number;
  /** HTTP status codes that are treated as retryable. */
  retryableStatusCodes: number[];
}

export interface ChainClientConfig {
  /** Selected chain environment. */
  environment: ChainEnvironment;
  /** Primary endpoint configuration. */
  endpoints: ChainEndpointConfig;
  /** Optional fallback endpoint configurations. */
  fallbackEndpoints?: ChainEndpointConfig[];
  /** Timeout settings. */
  timeouts: ChainTimeoutConfig;
  /** Retry settings. */
  retry: ChainRetryConfig;
  /** Optional extra headers appended to REST requests. */
  headers?: Record<string, string>;
  /** Optional user agent identifier. */
  userAgent?: string;
}

export interface ChainRequestOptions {
  /** Override the request timeout. */
  timeoutMs?: number;
  /** Override retry settings for a single request. */
  retry?: Partial<ChainRetryConfig>;
  /** Override request headers. */
  headers?: Record<string, string>;
  /** AbortSignal for cancellation. */
  signal?: AbortSignal;
  /** Status codes that should trigger a path-level fallback. */
  fallbackOnStatusCodes?: number[];
}

export interface ChainRequestResult<T> {
  data: T;
  endpoint: string;
  status: number;
}

export interface ChainStatus {
  chainId: string;
  latestHeight: number | null;
  validatorCount: number | null;
  rpcEndpoint?: string;
  restEndpoint?: string;
}

export type ChainErrorCode =
  | 'timeout'
  | 'network'
  | 'http_error'
  | 'invalid_response'
  | 'not_found'
  | 'retry_exhausted'
  | 'connection_failed'
  | 'wallet_error'
  | 'signing_error'
  | 'unknown';

/**
 * Base error for chain client failures.
 */
export class ChainClientError extends Error {
  readonly code: ChainErrorCode;
  readonly status?: number;
  readonly endpoint?: string;
  readonly cause?: unknown;

  constructor(
    code: ChainErrorCode,
    message: string,
    options?: { status?: number; endpoint?: string; cause?: unknown }
  ) {
    super(message);
    this.name = 'ChainClientError';
    this.code = code;
    this.status = options?.status;
    this.endpoint = options?.endpoint;
    this.cause = options?.cause;
  }
}

/**
 * Error for HTTP responses that indicate a failure.
 */
export class ChainHttpError extends ChainClientError {
  constructor(message: string, status: number, endpoint?: string, cause?: unknown) {
    super(status === 404 ? 'not_found' : 'http_error', message, { status, endpoint, cause });
    this.name = 'ChainHttpError';
  }
}

/**
 * Error for request timeouts.
 */
export class ChainTimeoutError extends ChainClientError {
  constructor(message: string, endpoint?: string, cause?: unknown) {
    super('timeout', message, { endpoint, cause });
    this.name = 'ChainTimeoutError';
  }
}

/**
 * Error for retry exhaustion.
 */
export class ChainRetryError extends ChainClientError {
  constructor(message: string, endpoint?: string, cause?: unknown) {
    super('retry_exhausted', message, { endpoint, cause });
    this.name = 'ChainRetryError';
  }
}

/**
 * Determines whether a thrown error is a ChainClientError.
 */
export function isChainClientError(error: unknown): error is ChainClientError {
  return error instanceof ChainClientError;
}

/**
 * Determines whether an error is likely transient and safe to retry.
 */
export function isRetryableError(
  error: unknown,
  retryableStatusCodes: number[]
): boolean {
  if (!error) return false;
  if (error instanceof ChainTimeoutError) return true;
  if (error instanceof ChainHttpError) {
    return retryableStatusCodes.includes(error.status ?? 0);
  }
  if (error instanceof ChainClientError) {
    return error.code === 'network' || error.code === 'connection_failed';
  }
  return false;
}

export interface PaginationRequest {
  key?: string;
  offset?: number;
  limit?: number;
  countTotal?: boolean;
}

export interface PaginationResponse {
  next_key?: string | null;
  total?: string | number;
}
