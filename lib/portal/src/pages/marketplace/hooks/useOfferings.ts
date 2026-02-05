/**
 * useOfferings Hook
 * VE-703: Enhanced marketplace offering queries with caching
 *
 * Provides cached offering queries, pagination, and search functionality
 * optimized for the customer marketplace browse experience.
 */

import { useState, useCallback, useRef, useEffect } from 'react';
import type {
  Offering,
  OfferingFilter,
  OfferingSort,
  OfferingType,
} from '../../../../types/marketplace';
import type { QueryClient } from '../../../../types/chain';

/**
 * Cache entry for offerings
 */
interface CacheEntry {
  offerings: Offering[];
  totalCount: number;
  timestamp: number;
  key: string;
}

/**
 * Offerings query state
 */
export interface OfferingsState {
  offerings: Offering[];
  totalCount: number;
  page: number;
  pageSize: number;
  isLoading: boolean;
  error: string | null;
  hasMore: boolean;
}

/**
 * Offerings query actions
 */
export interface OfferingsActions {
  search: (filter: OfferingFilter, sort?: OfferingSort, page?: number) => Promise<void>;
  loadMore: () => Promise<void>;
  refresh: () => Promise<void>;
  clearCache: () => void;
  setPageSize: (size: number) => void;
}

/**
 * Offerings hook options
 */
export interface UseOfferingsOptions {
  queryClient: QueryClient;
  pageSize?: number;
  cacheTimeMs?: number;
  initialFilter?: OfferingFilter;
  initialSort?: OfferingSort;
}

/**
 * Default sort configuration
 */
const DEFAULT_SORT: OfferingSort = {
  field: 'reliability_score',
  direction: 'desc',
};

/**
 * Generate cache key from filter and sort
 */
function generateCacheKey(
  filter: OfferingFilter,
  sort: OfferingSort,
  page: number,
  pageSize: number
): string {
  return JSON.stringify({ filter, sort, page, pageSize });
}

/**
 * useOfferings hook
 *
 * Provides cached offering queries with pagination support.
 */
export function useOfferings(options: UseOfferingsOptions): {
  state: OfferingsState;
  actions: OfferingsActions;
  filter: OfferingFilter;
  sort: OfferingSort;
} {
  const {
    queryClient,
    pageSize: initialPageSize = 20,
    cacheTimeMs = 60000, // 1 minute cache
    initialFilter = {},
    initialSort = DEFAULT_SORT,
  } = options;

  // State
  const [state, setState] = useState<OfferingsState>({
    offerings: [],
    totalCount: 0,
    page: 1,
    pageSize: initialPageSize,
    isLoading: false,
    error: null,
    hasMore: false,
  });

  const [filter, setFilter] = useState<OfferingFilter>(initialFilter);
  const [sort, setSort] = useState<OfferingSort>(initialSort);

  // Cache ref
  const cacheRef = useRef<Map<string, CacheEntry>>(new Map());
  const abortControllerRef = useRef<AbortController | null>(null);

  /**
   * Check if cache entry is valid
   */
  const isCacheValid = useCallback(
    (entry: CacheEntry | undefined): entry is CacheEntry => {
      if (!entry) return false;
      return Date.now() - entry.timestamp < cacheTimeMs;
    },
    [cacheTimeMs]
  );

  /**
   * Build query parameters from filter
   */
  const buildQueryParams = useCallback(
    (
      queryFilter: OfferingFilter,
      querySort: OfferingSort,
      page: number,
      pageSizeParam: number
    ): Record<string, string> => {
      const params: Record<string, string> = {
        page: String(page),
        limit: String(pageSizeParam),
        sort_field: querySort.field,
        sort_direction: querySort.direction,
      };

      if (queryFilter.query) params.query = queryFilter.query;
      if (queryFilter.types?.length) params.types = queryFilter.types.join(',');
      if (queryFilter.regions?.length) params.regions = queryFilter.regions.join(',');
      if (queryFilter.minCpuCores) params.min_cpu = String(queryFilter.minCpuCores);
      if (queryFilter.minMemoryGB) params.min_memory = String(queryFilter.minMemoryGB);
      if (queryFilter.minStorageGB) params.min_storage = String(queryFilter.minStorageGB);
      if (queryFilter.requireGpu) params.require_gpu = 'true';
      if (queryFilter.minReliabilityScore)
        params.min_reliability = String(queryFilter.minReliabilityScore);
      if (queryFilter.maxPricePerHour) params.max_price = queryFilter.maxPricePerHour;
      if (queryFilter.providerAddresses?.length)
        params.providers = queryFilter.providerAddresses.join(',');
      if (queryFilter.onlyEligible) params.only_eligible = 'true';

      return params;
    },
    []
  );

  /**
   * Fetch offerings from chain
   */
  const fetchOfferings = useCallback(
    async (
      queryFilter: OfferingFilter,
      querySort: OfferingSort,
      page: number,
      pageSizeParam: number
    ): Promise<{ offerings: Offering[]; total: number }> => {
      const params = buildQueryParams(queryFilter, querySort, page, pageSizeParam);

      const result = await queryClient.query<{
        offerings: Offering[];
        total: number;
      }>('/marketplace/offerings', params);

      return {
        offerings: result.offerings || [],
        total: result.total || 0,
      };
    },
    [queryClient, buildQueryParams]
  );

  /**
   * Search offerings with filter and sort
   */
  const search = useCallback(
    async (
      newFilter: OfferingFilter,
      newSort: OfferingSort = sort,
      page: number = 1
    ): Promise<void> => {
      // Cancel any in-flight request
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
      abortControllerRef.current = new AbortController();

      const cacheKey = generateCacheKey(newFilter, newSort, page, state.pageSize);

      // Check cache
      const cachedEntry = cacheRef.current.get(cacheKey);
      if (isCacheValid(cachedEntry)) {
        setFilter(newFilter);
        setSort(newSort);
        setState((prev) => ({
          ...prev,
          offerings: cachedEntry.offerings,
          totalCount: cachedEntry.totalCount,
          page,
          isLoading: false,
          error: null,
          hasMore: page * prev.pageSize < cachedEntry.totalCount,
        }));
        return;
      }

      // Set loading state
      setFilter(newFilter);
      setSort(newSort);
      setState((prev) => ({
        ...prev,
        isLoading: true,
        error: null,
        page,
      }));

      try {
        const { offerings, total } = await fetchOfferings(
          newFilter,
          newSort,
          page,
          state.pageSize
        );

        // Cache result
        cacheRef.current.set(cacheKey, {
          offerings,
          totalCount: total,
          timestamp: Date.now(),
          key: cacheKey,
        });

        setState((prev) => ({
          ...prev,
          offerings,
          totalCount: total,
          page,
          isLoading: false,
          error: null,
          hasMore: page * prev.pageSize < total,
        }));
      } catch (error) {
        // Ignore abort errors
        if (error instanceof Error && error.name === 'AbortError') {
          return;
        }

        setState((prev) => ({
          ...prev,
          isLoading: false,
          error: error instanceof Error ? error.message : 'Failed to fetch offerings',
        }));
      }
    },
    [sort, state.pageSize, isCacheValid, fetchOfferings]
  );

  /**
   * Load more offerings (pagination)
   */
  const loadMore = useCallback(async (): Promise<void> => {
    if (state.isLoading || !state.hasMore) return;

    const nextPage = state.page + 1;
    const cacheKey = generateCacheKey(filter, sort, nextPage, state.pageSize);

    // Check cache
    const cachedEntry = cacheRef.current.get(cacheKey);
    if (isCacheValid(cachedEntry)) {
      setState((prev) => ({
        ...prev,
        offerings: [...prev.offerings, ...cachedEntry.offerings],
        page: nextPage,
        hasMore: nextPage * prev.pageSize < cachedEntry.totalCount,
      }));
      return;
    }

    setState((prev) => ({ ...prev, isLoading: true }));

    try {
      const { offerings, total } = await fetchOfferings(filter, sort, nextPage, state.pageSize);

      // Cache result
      cacheRef.current.set(cacheKey, {
        offerings,
        totalCount: total,
        timestamp: Date.now(),
        key: cacheKey,
      });

      setState((prev) => ({
        ...prev,
        offerings: [...prev.offerings, ...offerings],
        totalCount: total,
        page: nextPage,
        isLoading: false,
        hasMore: nextPage * prev.pageSize < total,
      }));
    } catch (error) {
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error: error instanceof Error ? error.message : 'Failed to load more offerings',
      }));
    }
  }, [state.isLoading, state.hasMore, state.page, state.pageSize, filter, sort, isCacheValid, fetchOfferings]);

  /**
   * Refresh current results (bypass cache)
   */
  const refresh = useCallback(async (): Promise<void> => {
    const cacheKey = generateCacheKey(filter, sort, state.page, state.pageSize);
    cacheRef.current.delete(cacheKey);
    await search(filter, sort, state.page);
  }, [filter, sort, state.page, state.pageSize, search]);

  /**
   * Clear all cache
   */
  const clearCache = useCallback((): void => {
    cacheRef.current.clear();
  }, []);

  /**
   * Set page size
   */
  const setPageSize = useCallback((size: number): void => {
    setState((prev) => ({
      ...prev,
      pageSize: size,
      page: 1,
    }));
  }, []);

  // Cleanup abort controller on unmount
  useEffect(() => {
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, []);

  return {
    state,
    actions: {
      search,
      loadMore,
      refresh,
      clearCache,
      setPageSize,
    },
    filter,
    sort,
  };
}

/**
 * Category definition for marketplace
 */
export interface OfferingCategory {
  id: string;
  name: string;
  description: string;
  icon: string;
  types: OfferingType[];
  featured?: boolean;
}

/**
 * Predefined marketplace categories
 */
export const OFFERING_CATEGORIES: OfferingCategory[] = [
  {
    id: 'all',
    name: 'All Offerings',
    description: 'Browse all available offerings',
    icon: 'grid',
    types: [],
  },
  {
    id: 'compute',
    name: 'Compute',
    description: 'General purpose virtual machines',
    icon: 'cpu',
    types: ['compute'],
    featured: true,
  },
  {
    id: 'gpu',
    name: 'GPU Compute',
    description: 'GPU-accelerated instances for AI/ML',
    icon: 'gpu',
    types: ['gpu'],
    featured: true,
  },
  {
    id: 'storage',
    name: 'Storage',
    description: 'Block and object storage solutions',
    icon: 'database',
    types: ['storage'],
  },
  {
    id: 'kubernetes',
    name: 'Kubernetes',
    description: 'Managed Kubernetes clusters',
    icon: 'kubernetes',
    types: ['kubernetes'],
    featured: true,
  },
  {
    id: 'hpc',
    name: 'HPC / SLURM',
    description: 'High-performance computing clusters',
    icon: 'server',
    types: ['slurm'],
  },
  {
    id: 'custom',
    name: 'Custom',
    description: 'Specialized and custom offerings',
    icon: 'settings',
    types: ['custom'],
  },
];

/**
 * Region definition
 */
export interface Region {
  id: string;
  name: string;
  flag: string;
}

/**
 * Available regions
 */
export const REGIONS: Region[] = [
  { id: 'us-east', name: 'US East', flag: '🇺🇸' },
  { id: 'us-west', name: 'US West', flag: '🇺🇸' },
  { id: 'eu-west', name: 'EU West', flag: '🇪🇺' },
  { id: 'eu-central', name: 'EU Central', flag: '🇪🇺' },
  { id: 'asia-east', name: 'Asia East', flag: '🇯🇵' },
  { id: 'asia-south', name: 'Asia South', flag: '🇸🇬' },
];
