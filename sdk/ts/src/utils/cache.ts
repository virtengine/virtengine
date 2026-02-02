/**
 * Simple caching layer for VirtEngine TypeScript SDK
 */

/**
 * Query cache interface for caching RPC responses
 */
export interface QueryCache {
  get<T>(key: string): T | undefined;
  set<T>(key: string, value: T, ttlMs?: number): void;
  delete(key: string): boolean;
  clear(): void;
}

interface CacheEntry<T> {
  value: T;
  expiresAt: number;
}

/**
 * Cache configuration options
 */
export interface CacheOptions {
  /** Time-to-live in milliseconds (default: 30000) */
  ttlMs?: number;
  /** Maximum number of entries (default: 1000) */
  maxSize?: number;
}

/**
 * Simple in-memory cache with TTL and size limits
 */
export class MemoryCache implements QueryCache {
  private cache = new Map<string, CacheEntry<unknown>>();
  private defaultTtlMs: number;
  private maxSize: number;

  constructor(options?: CacheOptions | number) {
    if (typeof options === "number") {
      // Legacy constructor support
      this.defaultTtlMs = options;
      this.maxSize = 1000;
    } else {
      this.defaultTtlMs = options?.ttlMs ?? 30_000;
      this.maxSize = options?.maxSize ?? 1000;
    }
  }

  get<T>(key: string): T | undefined {
    const entry = this.cache.get(key);
    if (!entry) return undefined;

    if (Date.now() > entry.expiresAt) {
      this.cache.delete(key);
      return undefined;
    }

    return entry.value as T;
  }

  set<T>(key: string, value: T, ttlMs?: number): void {
    // Evict oldest entries if at capacity
    if (this.cache.size >= this.maxSize) {
      this.evictExpiredAndOldest();
    }

    const expiresAt = Date.now() + (ttlMs ?? this.defaultTtlMs);
    this.cache.set(key, { value, expiresAt });
  }

  delete(key: string): boolean {
    return this.cache.delete(key);
  }

  clear(): void {
    this.cache.clear();
  }

  /**
   * Check if a key exists and is not expired
   */
  has(key: string): boolean {
    const entry = this.cache.get(key);
    if (!entry) return false;
    if (Date.now() > entry.expiresAt) {
      this.cache.delete(key);
      return false;
    }
    return true;
  }

  /**
   * Get the current number of entries in the cache
   */
  size(): number {
    return this.cache.size;
  }

  /**
   * Remove all expired entries
   */
  prune(): number {
    const now = Date.now();
    let removed = 0;

    for (const [key, entry] of this.cache) {
      if (now > entry.expiresAt) {
        this.cache.delete(key);
        removed++;
      }
    }

    return removed;
  }

  /**
   * Evict expired entries and oldest entries if still at capacity
   */
  private evictExpiredAndOldest(): void {
    const now = Date.now();
    const keysToDelete: string[] = [];

    // First pass: collect expired entries
    for (const [key, entry] of this.cache) {
      if (entry.expiresAt < now) {
        keysToDelete.push(key);
      }
    }

    // Delete expired entries
    keysToDelete.forEach((key) => this.cache.delete(key));

    // If still at capacity, delete oldest entries (first in map)
    if (this.cache.size >= this.maxSize) {
      const entriesToRemove = this.cache.size - this.maxSize + 1;
      let removed = 0;
      for (const key of this.cache.keys()) {
        if (removed >= entriesToRemove) break;
        this.cache.delete(key);
        removed++;
      }
    }
  }

  /**
   * Generate a cache key from query method and parameters
   */
  static createKey(method: string, params: Record<string, unknown>): string {
    const sortedParams = Object.keys(params)
      .sort()
      .reduce((acc, key) => {
        const value = params[key];
        if (value !== undefined && value !== null) {
          acc[key] = value;
        }
        return acc;
      }, {} as Record<string, unknown>);

    return `${method}:${JSON.stringify(sortedParams)}`;
  }
}

/**
 * Higher-order function that wraps a function with caching
 */
export function withCache<TArgs extends unknown[], TResult>(
  cache: QueryCache,
  keyFn: (...args: TArgs) => string,
  ttlMs?: number,
): <TFn extends (...args: TArgs) => Promise<TResult>>(fn: TFn) => TFn {
  return <TFn extends (...args: TArgs) => Promise<TResult>>(fn: TFn): TFn => {
    return (async (...args: TArgs): Promise<TResult> => {
      const key = keyFn(...args);

      const cached = cache.get<TResult>(key);
      if (cached !== undefined) {
        return cached;
      }

      const result = await fn(...args);
      cache.set(key, result, ttlMs);
      return result;
    }) as TFn;
  };
}

/**
 * Options for creating a cached query
 */
export interface CachedQueryOptions<TParams> {
  cache?: QueryCache;
  keyFn?: (params: TParams) => string;
  ttlMs?: number;
}

/**
 * Creates a cached version of a query function
 */
export function createCachedQuery<TParams extends Record<string, unknown>, TResult>(
  queryFn: (params: TParams) => Promise<TResult>,
  options?: CachedQueryOptions<TParams>,
): (params: TParams) => Promise<TResult> {
  const cache = options?.cache ?? new MemoryCache();
  const keyFn = options?.keyFn ?? ((params: TParams) =>
    MemoryCache.createKey("query", params)
  );
  const ttlMs = options?.ttlMs;

  return async (params: TParams): Promise<TResult> => {
    const key = keyFn(params);

    const cached = cache.get<TResult>(key);
    if (cached !== undefined) {
      return cached;
    }

    const result = await queryFn(params);
    cache.set(key, result, ttlMs);
    return result;
  };
}

/**
 * Cache decorator for class methods
 */
export function cached(ttlMs?: number) {
  return function <T extends object>(
    _target: T,
    propertyKey: string | symbol,
    descriptor: PropertyDescriptor,
  ): PropertyDescriptor {
    const originalMethod = descriptor.value as (...args: unknown[]) => Promise<unknown>;
    const cacheMap = new WeakMap<T, MemoryCache>();

    descriptor.value = async function (this: T, ...args: unknown[]): Promise<unknown> {
      let cache = cacheMap.get(this);
      if (!cache) {
        cache = new MemoryCache({ ttlMs });
        cacheMap.set(this, cache);
      }

      const key = MemoryCache.createKey(String(propertyKey), { args });

      const cachedValue = cache.get(key);
      if (cachedValue !== undefined) {
        return cachedValue;
      }

      const result = await originalMethod.apply(this, args);
      cache.set(key, result, ttlMs);
      return result;
    };

    return descriptor;
  };
}
