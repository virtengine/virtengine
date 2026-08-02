/**
 * useMarketplace Hook
 * VE-703: Marketplace discovery, offering details, and checkout
 *
 * Provides marketplace browsing, filtering, and order management.
 */

import { useState, useCallback, useEffect, useContext, createContext, useRef } from 'react';
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
  MarketplaceError,
} from '../types/marketplace';
import { initialMarketplaceState } from '../types/marketplace';
import type { QueryClient, ChainEvent } from '../types/chain';
import {
  buildCheckoutMutationRequest,
  checkoutMutationRequestsEqual,
  submitCheckoutMutation,
  type CheckoutMutationAdapter,
  type CheckoutMutationContext,
  type CheckoutMutationProjector,
  type CheckoutMutationRequest,
} from '../components/marketplace/checkout-mutation';

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
  mutationAdapter?: CheckoutMutationAdapter;
  mutationContext?: CheckoutMutationContext;
  resultProjector?: CheckoutMutationProjector;
  mutationTimeoutMs?: number;
}

/**
 * Marketplace provider component
 */
export function MarketplaceProvider({
  children,
  queryClient,
  accountAddress,
  onEvent,
  mutationAdapter,
  mutationContext,
  resultProjector,
  mutationTimeoutMs,
}: MarketplaceProviderProps) {
  const [state, setState] = useState<MarketplaceState>(initialMarketplaceState);
  const stateRef = useRef(state);
  const activeSubmission = useRef<AbortController | null>(null);
  const submissionInFlight = useRef(false);
  const submissionToken = useRef(0);
  const checkoutNonce = useRef(0);
  const validationToken = useRef(0);
  const validationInFlight = useRef(false);
  const ordersToken = useRef(0);
  stateRef.current = state;

  const currentMutationRequest = useCallback((): CheckoutMutationRequest | null => {
    const checkout = stateRef.current.checkout;
    const offering = stateRef.current.selectedOffering;
    const validation = checkout?.validation;
    const boundContext =
      accountAddress && mutationContext?.customerAddress === accountAddress
        ? mutationContext
        : undefined;
    if (
      !checkout ||
      !offering ||
      offering.id !== checkout.request.offeringId ||
      !validation?.isValid ||
      validation.priceQuote.duration !== checkout.request.durationSeconds
    ) {
      return null;
    }
    return buildCheckoutMutationRequest(boundContext, {
      offeringId: checkout.request.offeringId,
      providerAddress: offering.providerAddress,
      durationSeconds: checkout.request.durationSeconds,
      priceAmount: validation.priceQuote.totalAmount,
      depositAmount: validation.priceQuote.depositAmount,
      priceDenom: validation.priceQuote.denom,
    });
  }, [accountAddress, mutationContext]);

  useEffect(() => () => activeSubmission.current?.abort(), []);

  useEffect(() => {
    activeSubmission.current?.abort();
    activeSubmission.current = null;
    submissionInFlight.current = false;
    submissionToken.current += 1;
    validationToken.current += 1;
    validationInFlight.current = false;
    setState((prev) => ({
      ...prev,
      checkout: prev.checkout
        ? prev.checkout.commit
          ? prev.checkout
          : {
              ...prev.checkout,
              validation: null,
              step: 'configure',
              error: null,
              nonce: ++checkoutNonce.current,
            }
        : null,
    }));
  }, [accountAddress, mutationContext?.chainId, mutationContext?.customerAddress]);

  useEffect(() => {
    activeSubmission.current?.abort();
    activeSubmission.current = null;
    submissionInFlight.current = false;
    submissionToken.current += 1;
  }, [mutationAdapter, resultProjector]);

  /**
   * Search offerings
   */
  const searchOfferings = useCallback(
    async (
      filter: OfferingFilter,
      sort: OfferingSort = { field: 'reliability_score', direction: 'desc' },
      page: number = 1
    ) => {
      setState((prev) => ({ ...prev, isLoading: true, filter, sort, page }));

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

        setState((prev) => ({
          ...prev,
          isLoading: false,
          offerings: result.offerings,
          totalCount: result.total,
          error: null,
        }));
      } catch (error) {
        setState((prev) => ({
          ...prev,
          isLoading: false,
          error: {
            code: 'network_error',
            message: error instanceof Error ? error.message : 'Failed to search offerings',
          },
        }));
      }
    },
    [queryClient, accountAddress]
  );

  /**
   * Get offering details
   */
  const getOffering = useCallback(
    async (offeringId: string): Promise<Offering> => {
      const result = await queryClient.query<Offering>(`/marketplace/offerings/${offeringId}`);
      return result;
    },
    [queryClient]
  );

  /**
   * Select offering for detail view
   */
  const selectOffering = useCallback((offering: Offering | null) => {
    setState((prev) => ({ ...prev, selectedOffering: offering }));
  }, []);

  /**
   * Start checkout flow
   */
  const startCheckout = useCallback(async (request: CheckoutRequest) => {
    submissionToken.current += 1;
    validationToken.current += 1;
    validationInFlight.current = false;
    activeSubmission.current?.abort();
    activeSubmission.current = null;
    submissionInFlight.current = false;
    setState((prev) => ({
      ...prev,
      checkout: {
        step: 'configure',
        request,
        validation: null,
        mfaChallenge: null,
        error: null,
        commit: null,
        nonce: ++checkoutNonce.current,
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

    const request = state.checkout.request;
    const checkoutNonce = state.checkout.nonce;
    const token = ++validationToken.current;
    validationInFlight.current = true;
    setState((prev) => ({
      ...prev,
      checkout: prev.checkout
        ? {
            ...prev.checkout,
            step: 'validate',
            validation: null,
            error: null,
          }
        : null,
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

      const currentCheckout = stateRef.current.checkout;
      if (
        validationToken.current !== token ||
        !currentCheckout ||
        currentCheckout.request.offeringId !== request.offeringId ||
        currentCheckout.request.durationSeconds !== request.durationSeconds ||
        currentCheckout.nonce !== checkoutNonce
      ) {
        throw new Error('request_changed');
      }
      setState((prev) => ({
        ...prev,
        checkout: prev.checkout
          ? {
              ...prev.checkout,
              validation,
              step:
                validation.mfaCheck.required && !validation.mfaCheck.satisfied ? 'mfa' : 'confirm',
            }
          : null,
      }));

      return validation;
    } catch (error) {
      if (validationToken.current !== token || stateRef.current.checkout?.nonce !== checkoutNonce) {
        throw error;
      }
      const marketplaceError = {
        code: 'network_error' as const,
        message: error instanceof Error ? error.message : 'Validation failed',
      };
      setState((prev) => ({
        ...prev,
        checkout: prev.checkout ? { ...prev.checkout, error: marketplaceError } : null,
      }));
      throw error;
    } finally {
      if (validationToken.current === token) validationInFlight.current = false;
    }
  }, [state.checkout, queryClient, accountAddress]);

  /**
   * Submit order
   */
  const submitOrder = useCallback(async (): Promise<Order> => {
    if (!state.checkout) {
      throw new Error('Checkout not validated');
    }
    if (validationInFlight.current) {
      throw new Error('Checkout validation in progress');
    }
    const existingCommit = state.checkout.commit;
    if (!existingCommit && !state.checkout.validation?.isValid) {
      throw new Error('Checkout not validated');
    }
    const request = existingCommit?.request ?? currentMutationRequest();
    if (!request) {
      throw new Error('feature_unavailable');
    }
    if (
      existingCommit &&
      (!accountAddress ||
        !mutationContext ||
        mutationContext.customerAddress !== accountAddress ||
        request.customerAddress !== accountAddress ||
        request.chainId !== mutationContext.chainId)
    ) {
      throw new Error('feature_unavailable');
    }
    if (submissionInFlight.current) {
      throw new Error('submission_in_progress');
    }

    if (!existingCommit && (!mutationAdapter || !resultProjector)) {
      throw new Error('feature_unavailable');
    }
    submissionInFlight.current = true;
    const token = ++submissionToken.current;
    const controller = new AbortController();
    activeSubmission.current = controller;
    setState((prev) => ({
      ...prev,
      checkout: prev.checkout ? { ...prev.checkout, step: 'submit' } : null,
    }));

    let acceptedCommit = existingCommit;
    try {
      const committed =
        existingCommit ??
        (await submitCheckoutMutation({
          adapter: mutationAdapter!,
          projector: resultProjector!,
          request,
          getCurrentRequest: currentMutationRequest,
          signal: controller.signal,
          timeoutMs: mutationTimeoutMs,
        }));
      acceptedCommit = committed;
      if (submissionToken.current !== token) throw new Error('submission_cancelled');
      if (!existingCommit) {
        setState((prev) => ({
          ...prev,
          checkout: prev.checkout
            ? {
                ...prev.checkout,
                step: 'committed',
                commit: committed,
                error: null,
              }
            : null,
        }));
      }
      let order: Order;
      try {
        order = await queryClient.query<Order>(`/marketplace/orders/${committed.orderId}`);
      } catch {
        throw new Error('order_projection_pending');
      }
      if (
        order.id !== committed.orderId ||
        order.txHash !== committed.txHash ||
        order.offeringId !== request.offeringId ||
        order.providerAddress !== request.providerAddress ||
        order.customerAddress !== request.customerAddress ||
        order.durationSeconds !== request.durationSeconds ||
        order.amount !== request.priceAmount ||
        order.deposit !== request.depositAmount ||
        order.denom !== request.priceDenom ||
        (!existingCommit && !checkoutMutationRequestsEqual(currentMutationRequest(), request))
      ) {
        throw new Error('invalid_committed_order');
      }

      if (submissionToken.current !== token) throw new Error('submission_cancelled');
      setState((prev) => ({
        ...prev,
        checkout: { ...prev.checkout!, step: 'complete' },
        orders: [...prev.orders, order],
      }));

      return order;
    } catch (error) {
      if (submissionToken.current !== token) throw error;
      const errorCode: MarketplaceError['code'] =
        error instanceof Error && error.message === 'order_projection_pending'
          ? 'order_not_found'
          : 'order_failed';
      const marketplaceError = {
        code: errorCode,
        message: error instanceof Error ? error.message : 'Order submission failed',
      };
      setState((prev) => ({
        ...prev,
        checkout: prev.checkout
          ? {
              ...prev.checkout,
              step: acceptedCommit ? 'committed' : prev.checkout.step,
              commit: acceptedCommit ?? prev.checkout.commit,
              error: marketplaceError,
            }
          : null,
      }));
      throw error;
    } finally {
      if (submissionToken.current === token) {
        if (activeSubmission.current === controller) activeSubmission.current = null;
        submissionInFlight.current = false;
      }
    }
  }, [
    state.checkout,
    currentMutationRequest,
    mutationAdapter,
    mutationTimeoutMs,
    queryClient,
    resultProjector,
  ]);

  /**
   * Cancel checkout
   */
  const cancelCheckout = useCallback(() => {
    submissionToken.current += 1;
    validationToken.current += 1;
    validationInFlight.current = false;
    activeSubmission.current?.abort();
    activeSubmission.current = null;
    submissionInFlight.current = false;
    setState((prev) => ({ ...prev, checkout: null }));
  }, []);

  /**
   * Get user's orders
   */
  const getOrders = useCallback(async () => {
    const token = ++ordersToken.current;
    if (!accountAddress) {
      setState((prev) => ({ ...prev, orders: [] }));
      return;
    }

    setState((prev) => ({ ...prev, orders: [] }));

    try {
      const result = await queryClient.query<{ orders: Order[] }>('/marketplace/orders', {
        customer: accountAddress,
      });

      if (ordersToken.current !== token) return;
      setState((prev) => ({ ...prev, orders: result.orders }));
    } catch (error) {
      // Handle error silently, orders will be empty
    }
  }, [queryClient, accountAddress]);

  /**
   * Get order details
   */
  const getOrder = useCallback(
    async (orderId: string): Promise<Order> => {
      const order = await queryClient.query<Order>(`/marketplace/orders/${orderId}`);
      return order;
    },
    [queryClient]
  );

  /**
   * Subscribe to order events
   */
  const subscribeToOrder = useCallback(
    (orderId: string, callback: (event: ChainEvent) => void): (() => void) => {
      // This would set up a WebSocket subscription
      // Return cleanup function
      return () => {
        // Cleanup subscription
      };
    },
    []
  );

  /**
   * Clear error
   */
  const clearError = useCallback(() => {
    setState((prev) => ({ ...prev, error: null }));
  }, []);

  // Fetch orders when account changes
  useEffect(() => {
    void getOrders();
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
    <MarketplaceContext.Provider value={{ state, actions }}>{children}</MarketplaceContext.Provider>
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
