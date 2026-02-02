import type { QueryCache } from "../utils/cache.ts";
import { QueryError, VirtEngineSDKError } from "../utils/errors.ts";

/**
 * Options for configuring a client
 */
export interface ClientOptions {
  cache?: QueryCache;
  enableCaching?: boolean;
}

/**
 * Base client class providing caching and error handling
 */
export abstract class BaseClient {
  protected cache?: QueryCache;
  protected enableCaching: boolean;

  constructor(options?: ClientOptions) {
    this.cache = options?.cache;
    this.enableCaching = options?.enableCaching ?? true;
  }

  /**
   * Get a cached value if caching is enabled
   */
  protected getCached<T>(key: string): T | undefined {
    if (!this.enableCaching || !this.cache) return undefined;
    return this.cache.get<T>(key);
  }

  /**
   * Set a cached value if caching is enabled
   */
  protected setCached<T>(key: string, value: T, ttlMs?: number): void {
    if (!this.enableCaching || !this.cache) return;
    this.cache.set(key, value, ttlMs);
  }

  /**
   * Handle query errors and wrap them in QueryError
   */
  protected handleQueryError(error: unknown, method: string): never {
    if (error instanceof VirtEngineSDKError) {
      throw error;
    }

    const message = error instanceof Error ? error.message : String(error);
    throw new QueryError(message, method, undefined, undefined, { cause: error });
  }
}
