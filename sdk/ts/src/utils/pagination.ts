/**
 * Pagination helpers for VirtEngine TypeScript SDK
 */

/**
 * Cosmos SDK PageRequest structure
 */
export interface PageRequest {
  key?: Uint8Array;
  offset?: bigint;
  limit?: bigint;
  countTotal?: boolean;
  reverse?: boolean;
}

/**
 * Cosmos SDK PageResponse structure
 */
export interface PageResponse {
  nextKey?: Uint8Array;
  total?: bigint;
}

/**
 * Paginated result wrapper
 */
export interface PaginatedResult<T> {
  items: T[];
  pagination: PageResponse;
  hasMore: boolean;
}

/**
 * Options for creating a page request
 */
export interface PageRequestOptions {
  limit?: number;
  offset?: number;
  countTotal?: boolean;
  reverse?: boolean;
  key?: Uint8Array;
}

/**
 * Creates a PageRequest from options
 */
export function createPageRequest(options?: PageRequestOptions): PageRequest {
  return {
    limit: options?.limit ? BigInt(options.limit) : undefined,
    offset: options?.offset ? BigInt(options.offset) : undefined,
    countTotal: options?.countTotal,
    reverse: options?.reverse,
    key: options?.key,
  };
}

/**
 * Checks if there is a next page based on the response
 */
export function hasNextPage(response: PageResponse): boolean {
  return !!(response.nextKey && response.nextKey.length > 0);
}

/**
 * Query function type for paginated queries
 */
export type PaginatedQueryFn<T> = (
  pagination: PageRequest
) => Promise<{ items: T[]; pagination: PageResponse }>;

/**
 * Options for paginated iteration
 */
export interface PaginateOptions {
  limit?: number;
  maxPages?: number;
}

/**
 * Async iterator for paginated queries
 * Yields pages of items until all pages are exhausted or maxPages is reached
 */
export async function* paginateQuery<T>(
  queryFn: PaginatedQueryFn<T>,
  options?: PaginateOptions,
): AsyncGenerator<T[], void, undefined> {
  const limit = options?.limit ?? 100;
  const maxPages = options?.maxPages ?? Infinity;

  let nextKey: Uint8Array | undefined;
  let pageCount = 0;

  do {
    const result = await queryFn(createPageRequest({ limit, key: nextKey }));

    if (result.items.length > 0) {
      yield result.items;
    }

    nextKey = result.pagination.nextKey;
    pageCount++;
  } while (nextKey && nextKey.length > 0 && pageCount < maxPages);
}

/**
 * Options for collecting all paginated results
 */
export interface CollectAllOptions {
  limit?: number;
  maxItems?: number;
}

/**
 * Collects all pages into a single array
 */
export async function collectAll<T>(
  queryFn: PaginatedQueryFn<T>,
  options?: CollectAllOptions,
): Promise<T[]> {
  const limit = options?.limit ?? 100;
  const maxItems = options?.maxItems ?? Infinity;

  const allItems: T[] = [];

  for await (const items of paginateQuery(queryFn, { limit })) {
    allItems.push(...items);
    if (allItems.length >= maxItems) {
      return allItems.slice(0, maxItems);
    }
  }

  return allItems;
}

/**
 * Creates a PageRequest for cursor-based pagination
 */
export function createCursorRequest(cursor?: string, limit = 100): PageRequest {
  return {
    key: cursor ? base64ToUint8Array(cursor) : undefined,
    limit: BigInt(limit),
  };
}

/**
 * Encodes a pagination key as a base64 cursor string
 */
export function encodeCursor(key: Uint8Array): string {
  return uint8ArrayToBase64(key);
}

/**
 * Decodes a base64 cursor string to a pagination key
 */
export function decodeCursor(cursor: string): Uint8Array {
  return base64ToUint8Array(cursor);
}

/**
 * Helper to convert Uint8Array to base64 string
 */
function uint8ArrayToBase64(bytes: Uint8Array): string {
  if (typeof Buffer !== "undefined") {
    return Buffer.from(bytes).toString("base64");
  }
  // Browser-compatible fallback
  let binary = "";
  for (let i = 0; i < bytes.length; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

/**
 * Helper to convert base64 string to Uint8Array
 */
function base64ToUint8Array(base64: string): Uint8Array {
  if (typeof Buffer !== "undefined") {
    return Uint8Array.from(Buffer.from(base64, "base64"));
  }
  // Browser-compatible fallback
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

/**
 * Creates a paginated result wrapper
 */
export function createPaginatedResult<T>(
  items: T[],
  pagination: PageResponse,
): PaginatedResult<T> {
  return {
    items,
    pagination,
    hasMore: hasNextPage(pagination),
  };
}
