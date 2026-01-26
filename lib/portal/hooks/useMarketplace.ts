/**
 * useMarketplace Hook
 * VE-703: Marketplace discovery, offering details, and checkout
 *
 * Provides marketplace browsing, filtering, and order management.
 */

import { useState, useCallback, useEffect, useContext, createContext } from 'react';
import type { ReactNode } from 'react';
import type {
  MarketplaceState,
  Offering,
  OfferingFilter,
  OfferingSort,
  Order,
  CheckoutRequest,
  CheckoutValidation,
  CheckoutState,
} from '../types/marketplace';
import { initialMarketplaceState } from '../types/marketplace';
import type { QueryClient, ChainEvent } from '../types/chain';

/**
 * Marketplace context value
 */
interface MarketplaceContextValue {
  state: MarketplaceState;
  actions: MarketplaceActions;
}

/**
 * Marketplace actions
 */
interface MarketplaceActions {
  /**
   * Search offerings with filter and sort
   */
  searchOfferings: (filter: OfferingFilter, sort?: OfferingSort, page?: number) => Promise<void>;

  /**
   * Get offering details
   */
  getOffering: (offeringId: string) => Promise<Offering>;

  /**
   * Select an offering for detail view
   */
  selectOffering: (offering: Offering | null) => void;

  /**
   * Start checkout flow
   */
  startCheckout: (request: CheckoutRequest) => Promise<void>;

  /**
   * Validate checkout
   */
  validateCheckout: () => Promise<CheckoutValidation>;

  /**
   * Submit order
   */
  submitOrder: () => Promise<Order>;

  /**
   * Cancel checkout
   */
  cancelCheckout: () => void;

  /**
   * Get user's orders
   */
  getOrders: () => Promise<void>;

  /**
   * Get order details
   */
  getOrder: (orderId: string) => Promise<Order>;

  /**
   * Subscribe to order events
   */
  subscribeToOrder: (orderId: string, callback: (event: ChainEvent) => void) => () => void;

  /**
   * Clear error
   */
  clearError: () => void;
}

/**
 * Marketplace context
 */
const MarketplaceContext = createContext<MarketplaceContextValue | null>(null);

/**
 * Marketplace provider props
 */
export interface MarketplaceProviderProps {
  children: ReactNode;
  queryClient: QueryClient;
  accountAddress: string | null;
  onEvent?: (event: ChainEvent) => void;
}

/**
 * Marketplace provider component
 */
export function MarketplaceProvider({
  children,
  queryClient,
  accountAddress,
  onEvent,
}: MarketplaceProviderProps) {
  const [state, setState] = useState<MarketplaceState>(initialMarketplaceState);

  /**
   * Search offerings
   */
  const searchOfferings = useCallback(async (
    filter: OfferingFilter,
    sort: OfferingSort = { field: 'reliability_score', direction: 'desc' },
    page: number = 1
  ) => {
    setState(prev => ({ ...prev, isLoading: true, filter, sort, page }));

    try {
      // Build query parameters
      const params: Record<string, string> = {
        page: String(page),
        limit: '20',
        sort_field: sort.field,
        sort_direction: sort.direction,
      };

      if (filter.query) params.query = filter.query;
      if (filter.types?.length) params.types = filter.types.join(',');
      if (filter.regions?.length) params.regions = filter.regions.join(',');
      if (filter.minCpuCores) params.min_cpu = String(filter.minCpuCores);
      if (filter.minMemoryGB) params.min_memory = String(filter.minMemoryGB);
      if (filter.minStorageGB) params.min_storage = String(filter.minStorageGB);
      if (filter.requireGpu) params.require_gpu = 'true';
      if (filter.minReliabilityScore) params.min_reliability = String(filter.minReliabilityScore);
      if (filter.maxPricePerHour) params.max_price = filter.maxPricePerHour;
      if (filter.onlyEligible && accountAddress) params.eligible_for = accountAddress;

      // Query offerings
      const result = await queryClient.query<{
        offerings: Offering[];
        total: number;
      }>('/marketplace/offerings', params);

      setState(prev => ({
        ...prev,
        isLoading: false,
        offerings: result.offerings,
        totalCount: result.total,
        error: null,
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        isLoading: false,
        error: {
          code: 'network_error',
          message: error instanceof Error ? error.message : 'Failed to search offerings',
        },
      }));
    }
  }, [queryClient, accountAddress]);

  /**
   * Get offering details
   */
  const getOffering = useCallback(async (offeringId: string): Promise<Offering> => {
    const result = await queryClient.query<Offering>(`/marketplace/offerings/${offeringId}`);
    return result;
  }, [queryClient]);

  /**
   * Select offering for detail view
   */
  const selectOffering = useCallback((offering: Offering | null) => {
    setState(prev => ({ ...prev, selectedOffering: offering }));
  }, []);

  /**
   * Start checkout flow
   */
  const startCheckout = useCallback(async (request: CheckoutRequest) => {
    setState(prev => ({
      ...prev,
      checkout: {
        step: 'configure',
        request,
        validation: null,
        mfaChallenge: null,
        error: null,
      },
    }));
  }, []);

  /**
   * Validate checkout
   */
  const validateCheckout = useCallback(async (): Promise<CheckoutValidation> => {
    if (!state.checkout) {
      throw new Error('No checkout in progress');
    }

    setState(prev => ({
      ...prev,
      checkout: prev.checkout ? { ...prev.checkout, step: 'validate' } : null,
    }));

    try {
      const validation = await queryClient.query<CheckoutValidation>(
        '/marketplace/checkout/validate',
        {
          offering_id: state.checkout.request.offeringId,
          duration: String(state.checkout.request.durationSeconds),
          customer: accountAddress || '',
        }
      );

      setState(prev => ({
        ...prev,
        checkout: prev.checkout ? {
          ...prev.checkout,
          validation,
          step: validation.mfaCheck.required && !validation.mfaCheck.satisfied ? 'mfa' : 'confirm',
        } : null,
      }));

      return validation;
    } catch (error) {
      const marketplaceError = {
        code: 'network_error' as const,
        message: error instanceof Error ? error.message : 'Validation failed',
      };
      setState(prev => ({
        ...prev,
        checkout: prev.checkout ? { ...prev.checkout, error: marketplaceError } : null,
      }));
      throw error;
    }
  }, [state.checkout, queryClient, accountAddress]);

  /**
   * Submit order
   */
  const submitOrder = useCallback(async (): Promise<Order> => {
    if (!state.checkout || !state.checkout.validation?.isValid) {
      throw new Error('Checkout not validated');
    }

    setState(prev => ({
      ...prev,
      checkout: prev.checkout ? { ...prev.checkout, step: 'submit' } : null,
    }));

    try {
      // This would actually sign and broadcast a transaction
      const order = await queryClient.query<Order>('/marketplace/orders', {
        // Would be a POST with tx bytes
      });

      setState(prev => ({
        ...prev,
        checkout: { ...prev.checkout!, step: 'complete' },
        orders: [...prev.orders, order],
      }));

      return order;
    } catch (error) {
      const marketplaceError = {
        code: 'order_failed' as const,
        message: error instanceof Error ? error.message : 'Order submission failed',
      };
      setState(prev => ({
        ...prev,
        checkout: prev.checkout ? { ...prev.checkout, error: marketplaceError } : null,
      }));
      throw error;
    }
  }, [state.checkout, queryClient]);

  /**
   * Cancel checkout
   */
  const cancelCheckout = useCallback(() => {
    setState(prev => ({ ...prev, checkout: null }));
  }, []);

  /**
   * Get user's orders
   */
  const getOrders = useCallback(async () => {
    if (!accountAddress) {
      setState(prev => ({ ...prev, orders: [] }));
      return;
    }

    try {
      const result = await queryClient.query<{ orders: Order[] }>(
        '/marketplace/orders',
        { customer: accountAddress }
      );

      setState(prev => ({ ...prev, orders: result.orders }));
    } catch (error) {
      // Handle error silently, orders will be empty
    }
  }, [queryClient, accountAddress]);

  /**
   * Get order details
   */
  const getOrder = useCallback(async (orderId: string): Promise<Order> => {
    const order = await queryClient.query<Order>(`/marketplace/orders/${orderId}`);
    return order;
  }, [queryClient]);

  /**
   * Subscribe to order events
   */
  const subscribeToOrder = useCallback((
    orderId: string,
    callback: (event: ChainEvent) => void
  ): (() => void) => {
    // This would set up a WebSocket subscription
    // Return cleanup function
    return () => {
      // Cleanup subscription
    };
  }, []);

  /**
   * Clear error
   */
  const clearError = useCallback(() => {
    setState(prev => ({ ...prev, error: null }));
  }, []);

  // Fetch orders when account changes
  useEffect(() => {
    if (accountAddress) {
      getOrders();
    }
  }, [accountAddress, getOrders]);

  const actions: MarketplaceActions = {
    searchOfferings,
    getOffering,
    selectOffering,
    startCheckout,
    validateCheckout,
    submitOrder,
    cancelCheckout,
    getOrders,
    getOrder,
    subscribeToOrder,
    clearError,
  };

  return (
    <MarketplaceContext.Provider value={{ state, actions }}>
      {children}
    </MarketplaceContext.Provider>
  );
}

/**
 * Use marketplace hook
 */
export function useMarketplace(): MarketplaceContextValue {
  const context = useContext(MarketplaceContext);
  if (!context) {
    throw new Error('useMarketplace must be used within a MarketplaceProvider');
  }
  return context;
}
