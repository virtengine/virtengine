/**
 * Chain query client with endpoint discovery, retry, and fallback.
 */

import { StargateClient } from '@cosmjs/stargate';
import {
  ChainClientConfig,
  ChainClientError,
  ChainHttpError,
  ChainRequestOptions,
  ChainRequestResult,
  ChainRetryError,
  ChainStatus,
  ChainTimeoutError,
  isRetryableError,
} from './types';

const DEFAULT_FALLBACK_STATUS = [404, 501];

function mergeHeaders(
  base?: Record<string, string>,
  override?: Record<string, string>
): Record<string, string> | undefined {
  if (!base && !override) return undefined;
  return { ...(base ?? {}), ...(override ?? {}) };
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function withTimeout<T>(
  promise: Promise<T>,
  timeoutMs: number,
  endpoint?: string
): Promise<T> {
  if (!timeoutMs || timeoutMs <= 0) return promise;
  let timeoutHandle: ReturnType<typeof setTimeout> | null = null;

  const timeoutPromise = new Promise<T>((_, reject) => {
    timeoutHandle = setTimeout(() => {
      reject(new ChainTimeoutError(`Request timed out after ${timeoutMs}ms`, endpoint));
    }, timeoutMs);
  });

  return Promise.race([promise, timeoutPromise]).finally(() => {
    if (timeoutHandle) {
      clearTimeout(timeoutHandle);
    }
  }) as Promise<T>;
}

function normalizeEndpoint(endpoint: string): string {
  return endpoint.replace(/\/$/, '');
}

function buildUrl(endpoint: string, path: string, params?: Record<string, string>): string {
  const base = normalizeEndpoint(endpoint);
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  const url = new URL(`${base}${normalizedPath}`);
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== '') {
        url.searchParams.set(key, value);
      }
    });
  }
  return url.toString();
}

function calcBackoff(baseDelayMs: number, maxDelayMs: number, jitterMs: number, attempt: number) {
  const exp = Math.min(maxDelayMs, baseDelayMs * Math.pow(2, attempt));
  const jitter = Math.random() * jitterMs;
  return exp + jitter;
}

export class ChainQueryClient {
  private readonly config: ChainClientConfig;
  private stargateClient: StargateClient | null = null;
  private stargateEndpoint: string | null = null;

  constructor(config: ChainClientConfig) {
    this.config = config;
  }

  /** Returns the current chain configuration. */
  getConfig(): ChainClientConfig {
    return this.config;
  }

  /** Returns all REST endpoints, including fallbacks. */
  getRestEndpoints(): string[] {
    const endpoints = [this.config.endpoints, ...(this.config.fallbackEndpoints ?? [])];
    return endpoints.flatMap((entry) => entry.restEndpoints ?? []).filter(Boolean);
  }

  /** Returns all RPC endpoints, including fallbacks. */
  getRpcEndpoints(): string[] {
    const endpoints = [this.config.endpoints, ...(this.config.fallbackEndpoints ?? [])];
    return endpoints.flatMap((entry) => entry.rpcEndpoints ?? []).filter(Boolean);
  }

  /** Returns all WS endpoints, including fallbacks. */
  getWsEndpoints(): string[] {
    const endpoints = [this.config.endpoints, ...(this.config.fallbackEndpoints ?? [])];
    return endpoints.flatMap((entry) => entry.wsEndpoints ?? []).filter(Boolean);
  }

  /** Fetch JSON from a REST endpoint with retry + fallback. */
  async getJson<T>(
    path: string,
    params?: Record<string, string>,
    options: ChainRequestOptions = {}
  ): Promise<ChainRequestResult<T>> {
    return this.requestJson<T>('GET', path, undefined, params, options);
  }

  /** POST JSON to a REST endpoint with retry + fallback. */
  async postJson<T>(
    path: string,
    body?: unknown,
    params?: Record<string, string>,
    options: ChainRequestOptions = {}
  ): Promise<ChainRequestResult<T>> {
    return this.requestJson<T>('POST', path, body, params, options);
  }

  /**
   * Fetch JSON while trying multiple path variants (useful when endpoints vary).
   */
  async getJsonWithPathFallback<T>(
    paths: string[],
    params?: Record<string, string>,
    options: ChainRequestOptions = {}
  ): Promise<ChainRequestResult<T>> {
    const fallbackStatuses = options.fallbackOnStatusCodes ?? DEFAULT_FALLBACK_STATUS;
    let lastError: unknown = null;

    for (const path of paths) {
      try {
        return await this.getJson<T>(path, params, options);
      } catch (error) {
        lastError = error;
        if (error instanceof ChainHttpError && fallbackStatuses.includes(error.status ?? 0)) {
          continue;
        }
        if (error instanceof ChainClientError && error.code === 'not_found') {
          continue;
        }
        break;
      }
    }

    throw lastError ?? new ChainClientError('unknown', 'Chain request failed');
  }

  /**
   * Returns a connected Stargate client using the first reachable RPC endpoint.
   */
  async getStargateClient(): Promise<StargateClient> {
    if (this.stargateClient) return this.stargateClient;

    const rpcEndpoints = this.getRpcEndpoints();
    if (rpcEndpoints.length === 0) {
      throw new ChainClientError('connection_failed', 'No RPC endpoints configured');
    }

    let lastError: unknown = null;

    for (const endpoint of rpcEndpoints) {
      try {
        const client = await withTimeout(
          StargateClient.connect(endpoint),
          this.config.timeouts.connectMs,
          endpoint
        );
        this.stargateClient = client;
        this.stargateEndpoint = endpoint;
        return client;
      } catch (error) {
        lastError = error;
      }
    }

    throw new ChainClientError(
      'connection_failed',
      'Unable to connect to any RPC endpoint',
      { cause: lastError }
    );
  }

  /** Disconnects the cached Stargate client. */
  disconnect(): void {
    if (this.stargateClient) {
      this.stargateClient.disconnect();
    }
    this.stargateClient = null;
    this.stargateEndpoint = null;
  }

  /** Retrieves chain status (chain ID, height, validator count). */
  async getChainStatus(): Promise<ChainStatus> {
    const restEndpoints = this.getRestEndpoints();
    if (restEndpoints.length === 0) {
      throw new ChainClientError('connection_failed', 'No REST endpoints configured');
    }

    const [blockRes, validatorRes] = await Promise.all([
      this.getJsonWithPathFallback<{ block?: { header?: { chain_id?: string; height?: string } } }>(
        ['/cosmos/base/tendermint/v1beta1/blocks/latest', '/blocks/latest']
      ),
      this.getJsonWithPathFallback<{ validators?: unknown[] }>(
        ['/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED']
      ),
    ]);

    const rawHeight = blockRes.data?.block?.header?.height ?? '0';
    const parsedHeight = Number(rawHeight);

    return {
      chainId: blockRes.data?.block?.header?.chain_id ?? this.config.endpoints.chainId,
      latestHeight: Number.isNaN(parsedHeight) ? null : parsedHeight,
      validatorCount: blockRes ? (validatorRes.data?.validators?.length ?? 0) : null,
      rpcEndpoint: this.stargateEndpoint ?? undefined,
      restEndpoint: blockRes.endpoint,
    };
  }

  private async requestJson<T>(
    method: 'GET' | 'POST',
    path: string,
    body?: unknown,
    params?: Record<string, string>,
    options: ChainRequestOptions = {}
  ): Promise<ChainRequestResult<T>> {
    const endpoints = this.getRestEndpoints();
    if (endpoints.length === 0) {
      throw new ChainClientError('connection_failed', 'No REST endpoints configured');
    }

    const requestHeaders = mergeHeaders(this.config.headers, options.headers);
    const fallbackStatuses = options.fallbackOnStatusCodes ?? DEFAULT_FALLBACK_STATUS;
    let lastError: unknown = null;

    for (const endpoint of endpoints) {
      try {
        return await this.requestWithRetry<T>(
          endpoint,
          method,
          path,
          body,
          params,
          requestHeaders,
          options
        );
      } catch (error) {
        lastError = error;
        if (error instanceof ChainHttpError && fallbackStatuses.includes(error.status ?? 0)) {
          continue;
        }
        if (error instanceof ChainTimeoutError) {
          continue;
        }
        if (error instanceof ChainClientError && error.code === 'network') {
          continue;
        }
        break;
      }
    }

    throw lastError ?? new ChainClientError('unknown', 'Chain request failed');
  }

  private async requestWithRetry<T>(
    endpoint: string,
    method: 'GET' | 'POST',
    path: string,
    body: unknown,
    params: Record<string, string> | undefined,
    headers: Record<string, string> | undefined,
    options: ChainRequestOptions
  ): Promise<ChainRequestResult<T>> {
    const retryConfig = { ...this.config.retry, ...options.retry };

    let attempt = 0;
    let lastError: unknown = null;

    while (attempt <= retryConfig.maxRetries) {
      try {
        const url = buildUrl(endpoint, path, params);
        const response = await withTimeout(
          fetch(url, {
            method,
            headers: {
              Accept: 'application/json',
              ...(method === 'POST' ? { 'Content-Type': 'application/json' } : {}),
              ...(headers ?? {}),
              ...(this.config.userAgent ? { 'User-Agent': this.config.userAgent } : {}),
            },
            body: body !== undefined ? JSON.stringify(body) : undefined,
            signal: options.signal,
          }),
          options.timeoutMs ?? this.config.timeouts.requestMs,
          endpoint
        );

        if (!response.ok) {
          const text = await response.text();
          throw new ChainHttpError(
            text || `Request failed with status ${response.status}`,
            response.status,
            endpoint
          );
        }

        const data = (await response.json()) as T;
        return { data, endpoint, status: response.status };
      } catch (error) {
        lastError = error;
        if (attempt >= retryConfig.maxRetries || !isRetryableError(error, retryConfig.retryableStatusCodes)) {
          break;
        }

        const delay = calcBackoff(
          retryConfig.baseDelayMs,
          retryConfig.maxDelayMs,
          retryConfig.jitterMs,
          attempt
        );
        await sleep(delay);
        attempt += 1;
      }
    }

    if (lastError instanceof ChainHttpError || lastError instanceof ChainTimeoutError) {
      throw lastError;
    }

    if (lastError) {
      throw new ChainRetryError('Chain request failed after retries', endpoint, lastError);
    }

    throw new ChainClientError('unknown', 'Chain request failed');
  }
}

/**
 * Convenience factory for creating a ChainQueryClient.
 */
export function createChainQueryClient(config: ChainClientConfig): ChainQueryClient {
  return new ChainQueryClient(config);
}
