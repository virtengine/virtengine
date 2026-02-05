/**
 * VE-25H: Offering Sync Hook
 * React hook for managing offering sync state with the provider daemon API.
 */

'use client';

import { useCallback, useEffect, useState } from 'react';
import type {
  OfferingFilters,
  OfferingListResponse,
  OfferingPublication,
  OfferingStats,
  SyncStatus,
  UpdatePricingRequest,
} from '@/types/offering';

// =============================================================================
// API Configuration
// =============================================================================

const API_BASE_URL = process.env.NEXT_PUBLIC_PROVIDER_API_URL || '/api/provider';

interface ApiOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
  body?: unknown;
  headers?: Record<string, string>;
}

interface ApiErrorResponse {
  message?: string;
}

async function apiCall<T>(endpoint: string, options: ApiOptions = {}): Promise<T> {
  const { method = 'GET', body, headers = {} } = options;

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    method,
    headers: {
      'Content-Type': 'application/json',
      ...headers,
    },
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!response.ok) {
    let error: ApiErrorResponse = { message: response.statusText };
    try {
      error = (await response.json()) as ApiErrorResponse;
    } catch {
      // Use default error message
    }
    throw new Error(error.message ?? `API error: ${response.status}`);
  }

  return response.json() as Promise<T>;
}

// =============================================================================
// Hook State Types
// =============================================================================

interface UseOfferingSyncState {
  offerings: OfferingPublication[];
  stats: OfferingStats | null;
  syncStatus: SyncStatus | null;
  isLoading: boolean;
  error: string | null;
  total: number;
  page: number;
  pageSize: number;
}

interface UseOfferingSyncActions {
  refresh: () => Promise<void>;
  setFilters: (filters: OfferingFilters) => void;
  publishOffering: (waldurUuid: string) => Promise<void>;
  pauseOffering: (offeringId: string) => Promise<void>;
  activateOffering: (offeringId: string) => Promise<void>;
  deprecateOffering: (offeringId: string) => Promise<void>;
  updatePricing: (offeringId: string, pricing: UpdatePricingRequest) => Promise<void>;
  getOffering: (offeringId: string) => Promise<OfferingPublication>;
}

export type UseOfferingSyncReturn = UseOfferingSyncState & UseOfferingSyncActions;

// =============================================================================
// Hook Implementation
// =============================================================================

export function useOfferingSync(initialFilters: OfferingFilters = {}): UseOfferingSyncReturn {
  const [state, setState] = useState<UseOfferingSyncState>({
    offerings: [],
    stats: null,
    syncStatus: null,
    isLoading: true,
    error: null,
    total: 0,
    page: 1,
    pageSize: 20,
  });

  const [filters, setFilters] = useState<OfferingFilters>(initialFilters);

  // Fetch offerings list
  const fetchOfferings = useCallback(async () => {
    setState((prev) => ({ ...prev, isLoading: true, error: null }));

    try {
      const queryParams = new URLSearchParams();
      if (filters.status) queryParams.set('status', filters.status);
      if (filters.category) queryParams.set('category', filters.category);
      if (filters.search) queryParams.set('search', filters.search);
      if (filters.sortBy) queryParams.set('sortBy', filters.sortBy);
      if (filters.sortOrder) queryParams.set('sortOrder', filters.sortOrder);
      if (filters.page) queryParams.set('page', filters.page.toString());
      if (filters.pageSize) queryParams.set('pageSize', filters.pageSize.toString());

      const response = await apiCall<OfferingListResponse>(
        `/offerings?${queryParams.toString()}`
      );

      setState((prev) => ({
        ...prev,
        offerings: response.offerings,
        total: response.total,
        page: response.page,
        pageSize: response.pageSize,
        isLoading: false,
      }));
    } catch (err) {
      setState((prev) => ({
        ...prev,
        error: err instanceof Error ? err.message : 'Failed to fetch offerings',
        isLoading: false,
      }));
    }
  }, [filters]);

  // Fetch stats
  const fetchStats = useCallback(async () => {
    try {
      const stats = await apiCall<OfferingStats>('/offerings/stats');
      setState((prev) => ({ ...prev, stats }));
    } catch (err) {
      console.error('Failed to fetch offering stats:', err);
    }
  }, []);

  // Fetch sync status
  const fetchSyncStatus = useCallback(async () => {
    try {
      const syncStatus = await apiCall<SyncStatus>('/offerings/sync-status');
      setState((prev) => ({ ...prev, syncStatus }));
    } catch (err) {
      console.error('Failed to fetch sync status:', err);
    }
  }, []);

  // Refresh all data
  const refresh = useCallback(async () => {
    await Promise.all([fetchOfferings(), fetchStats(), fetchSyncStatus()]);
  }, [fetchOfferings, fetchStats, fetchSyncStatus]);

  // Initial load
  useEffect(() => {
    void refresh();
  }, [refresh]);

  // Auto-refresh every 30 seconds when sync is running
  useEffect(() => {
    if (state.syncStatus?.isRunning) {
      const interval = setInterval(() => {
        void refresh();
      }, 30000);
      return () => clearInterval(interval);
    }
  }, [state.syncStatus?.isRunning, refresh]);

  // Actions
  const publishOffering = useCallback(async (waldurUuid: string) => {
    setState((prev) => ({ ...prev, isLoading: true }));
    try {
      await apiCall(`/offerings/${waldurUuid}/publish`, { method: 'POST' });
      await refresh();
    } catch (err) {
      setState((prev) => ({
        ...prev,
        error: err instanceof Error ? err.message : 'Failed to publish offering',
        isLoading: false,
      }));
      throw err;
    }
  }, [refresh]);

  const pauseOffering = useCallback(async (offeringId: string) => {
    setState((prev) => ({ ...prev, isLoading: true }));
    try {
      await apiCall(`/offerings/${offeringId}/pause`, { method: 'POST' });
      await refresh();
    } catch (err) {
      setState((prev) => ({
        ...prev,
        error: err instanceof Error ? err.message : 'Failed to pause offering',
        isLoading: false,
      }));
      throw err;
    }
  }, [refresh]);

  const activateOffering = useCallback(async (offeringId: string) => {
    setState((prev) => ({ ...prev, isLoading: true }));
    try {
      await apiCall(`/offerings/${offeringId}/activate`, { method: 'POST' });
      await refresh();
    } catch (err) {
      setState((prev) => ({
        ...prev,
        error: err instanceof Error ? err.message : 'Failed to activate offering',
        isLoading: false,
      }));
      throw err;
    }
  }, [refresh]);

  const deprecateOffering = useCallback(async (offeringId: string) => {
    setState((prev) => ({ ...prev, isLoading: true }));
    try {
      await apiCall(`/offerings/${offeringId}/deprecate`, { method: 'POST' });
      await refresh();
    } catch (err) {
      setState((prev) => ({
        ...prev,
        error: err instanceof Error ? err.message : 'Failed to deprecate offering',
        isLoading: false,
      }));
      throw err;
    }
  }, [refresh]);

  const updatePricing = useCallback(async (offeringId: string, pricing: UpdatePricingRequest) => {
    setState((prev) => ({ ...prev, isLoading: true }));
    try {
      await apiCall(`/offerings/${offeringId}/pricing`, {
        method: 'PUT',
        body: pricing,
      });
      await refresh();
    } catch (err) {
      setState((prev) => ({
        ...prev,
        error: err instanceof Error ? err.message : 'Failed to update pricing',
        isLoading: false,
      }));
      throw err;
    }
  }, [refresh]);

  const getOffering = useCallback(async (offeringId: string): Promise<OfferingPublication> => {
    return apiCall<OfferingPublication>(`/offerings/${offeringId}`);
  }, []);

  return {
    ...state,
    setFilters,
    refresh,
    publishOffering,
    pauseOffering,
    activateOffering,
    deprecateOffering,
    updatePricing,
    getOffering,
  };
}

// =============================================================================
// Utility Hooks
// =============================================================================

/**
 * Hook for polling sync status during publication
 */
export function useSyncPolling(enabled: boolean, intervalMs = 5000) {
  const [syncStatus, setSyncStatus] = useState<SyncStatus | null>(null);

  useEffect(() => {
    if (!enabled) return;

    const fetchStatus = () => {
      void apiCall<SyncStatus>('/offerings/sync-status')
        .then((status) => setSyncStatus(status))
        .catch(() => {
          // Ignore errors during polling
        });
    };

    fetchStatus();
    const interval = setInterval(fetchStatus, intervalMs);
    return () => clearInterval(interval);
  }, [enabled, intervalMs]);

  return syncStatus;
}

/**
 * Get human-readable status label
 */
export function getStatusLabel(status: string): string {
  const labels: Record<string, string> = {
    pending: 'Pending',
    published: 'Published',
    failed: 'Failed',
    paused: 'Paused',
    deprecated: 'Deprecated',
    draft: 'Draft',
  };
  return labels[status] || status;
}

/**
 * Get status color class for styling
 */
export function getStatusColor(status: string): string {
  const colors: Record<string, string> = {
    pending: 'text-yellow-500',
    published: 'text-green-500',
    failed: 'text-red-500',
    paused: 'text-gray-500',
    deprecated: 'text-orange-500',
    draft: 'text-blue-500',
  };
  return colors[status] || 'text-gray-500';
}

/**
 * Get category icon name
 */
export function getCategoryIcon(category: string): string {
  const icons: Record<string, string> = {
    compute: '💻',
    storage: '💾',
    network: '🌐',
    hpc: '🚀',
    gpu: '🎮',
    ml: '🤖',
    other: '📦',
  };
  return icons[category] || '📦';
}
