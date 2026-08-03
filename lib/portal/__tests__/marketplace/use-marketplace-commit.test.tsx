import * as React from 'react';
import { act } from 'react';
import { createRoot, type Root } from 'react-dom/client';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { MarketplaceProvider, useMarketplace } from '../../hooks/useMarketplace';
import type { MarketplaceState, Offering, Order } from '../../types/marketplace';
import type { QueryClient } from '../../types/chain';
import type { CheckoutMutationAdapter } from '../../components/marketplace/checkout-mutation';

const offering: Offering = {
  id: 'offering-1',
  providerAddress: 'virtengine1provider',
  providerName: 'Provider',
  title: 'GPU Compute',
  description: 'Compute',
  type: 'gpu',
  status: 'active',
  region: 'us-east',
  resources: { cpuCores: 8, memoryGB: 32, storageGB: 100, attributes: {} },
  pricing: {
    basePrice: '25',
    denom: 'uve',
    depositRequired: '50',
    minDurationSeconds: 3600,
  },
  identityRequirements: { minScore: 0, requiredScopes: [], mfaRequired: false },
  reliabilityScore: 99,
  benchmarkSummary: {
    cpuScore: 99,
    memoryScore: 99,
    storageScore: 99,
    networkScore: 99,
    overallScore: 99,
    lastBenchmarkAt: 1,
    suiteVersion: '1',
  },
  createdAt: 1,
  updatedAt: 1,
  hasEncryptedDetails: false,
};

const order: Order = {
  id: 'order-1',
  offeringId: offering.id,
  customerAddress: 'virtengine1customer',
  providerAddress: offering.providerAddress,
  state: 'created',
  createdAt: 1,
  stateHistory: [],
  amount: '25',
  denom: 'uve',
  deposit: '50',
  durationSeconds: 3600,
  events: [],
  hasEncryptedDetails: false,
  txHash: 'ABC123',
};

const validationFor = (totalAmount: string) => ({
  isValid: true,
  errors: [],
  identityCheck: {
    passed: true,
    currentScore: 100,
    requiredScore: 0,
    missingScopes: [],
  },
  mfaCheck: { required: false, satisfied: true },
  priceQuote: {
    totalAmount,
    depositAmount: '50',
    unitPrice: totalAmount,
    duration: 3600,
    denom: 'uve',
  },
});

describe('MarketplaceProvider committed submission', () => {
  let container: HTMLDivElement;
  let root: Root;
  let marketplace: ReturnType<typeof useMarketplace>;
  let query: ReturnType<typeof vi.fn>;

  const Consumer = () => {
    marketplace = useMarketplace();
    return null;
  };

  const renderProvider = async (adapter?: CheckoutMutationAdapter) => {
    const queryClient = { query } as unknown as QueryClient;
    await act(async () =>
      root.render(
        <MarketplaceProvider
          queryClient={queryClient}
          accountAddress="virtengine1customer"
          mutationAdapter={adapter}
          mutationContext={{
            chainId: 'virtengine-1',
            customerAddress: 'virtengine1customer',
          }}
          resultProjector={(value) => value}
        >
          <Consumer />
        </MarketplaceProvider>
      )
    );
  };

  const prepareCheckout = async () => {
    await act(async () => marketplace.actions.selectOffering(offering));
    await act(async () =>
      marketplace.actions.startCheckout({
        offeringId: offering.id,
        durationSeconds: 3600,
      })
    );
    query.mockResolvedValueOnce(validationFor('25'));
    await act(async () => marketplace.actions.validateCheckout());
  };

  beforeEach(() => {
    globalThis.IS_REACT_ACT_ENVIRONMENT = true;
    container = document.createElement('div');
    document.body.appendChild(container);
    root = createRoot(container);
    query = vi.fn((path: string) =>
      Promise.resolve(path === '/marketplace/orders' ? { orders: [] } : undefined)
    );
  });

  afterEach(async () => {
    await act(async () => root.unmount());
    container.remove();
  });

  it('fails closed without a mutation adapter', async () => {
    await renderProvider();
    await prepareCheckout();
    query.mockClear();

    await expect(marketplace.actions.submitOrder()).rejects.toThrow('feature_unavailable');
    expect(marketplace.state.checkout?.step).toBe('confirm');
    expect(marketplace.state.orders).toEqual([]);
    expect(query).not.toHaveBeenCalled();
  });

  it('accepts only an exactly projected committed order', async () => {
    const adapter: CheckoutMutationAdapter = {
      submitOrder: vi.fn((request, submission) =>
        Promise.resolve({
          status: 'committed',
          code: 0,
          orderId: order.id,
          txHash: order.txHash,
          blockHeight: 42,
          requestDigest: submission.requestDigest,
          idempotencyKey: submission.idempotencyKey,
          request,
        })
      ),
    };
    await renderProvider(adapter);
    await prepareCheckout();
    query.mockClear();
    query.mockResolvedValueOnce(order);

    let result: Order | undefined;
    await act(async () => {
      result = await marketplace.actions.submitOrder();
    });

    expect(result).toEqual(order);
    expect(marketplace.state.checkout?.step).toBe('complete');
    expect(marketplace.state.orders).toEqual([order]);
    expect(query).toHaveBeenCalledWith(`/marketplace/orders/${order.id}`);
  });

  it('retries delayed order projection without rebroadcasting', async () => {
    const submitOrder = vi.fn((request, submission) =>
      Promise.resolve({
        status: 'committed',
        code: 0,
        orderId: order.id,
        txHash: order.txHash,
        blockHeight: 42,
        requestDigest: submission.requestDigest,
        idempotencyKey: submission.idempotencyKey,
        request,
      })
    );
    await renderProvider({ submitOrder });
    await prepareCheckout();
    query.mockClear();
    query.mockRejectedValueOnce(new Error('not indexed'));

    let projectionError: unknown;
    await act(async () => {
      try {
        await marketplace.actions.submitOrder();
      } catch (error) {
        projectionError = error;
      }
    });
    expect(projectionError).toMatchObject({
      message: 'order_projection_pending',
    });
    expect(marketplace.state.checkout?.step).toBe('committed');
    expect(marketplace.state.checkout?.commit).toMatchObject({
      orderId: order.id,
      txHash: order.txHash,
      blockHeight: 42,
      request: { priceAmount: '25', depositAmount: '50', priceDenom: 'uve' },
    });
    expect(submitOrder).toHaveBeenCalledTimes(1);

    query.mockResolvedValueOnce(order);
    await act(async () => {
      await marketplace.actions.submitOrder();
    });
    expect(submitOrder).toHaveBeenCalledTimes(1);
    expect(marketplace.state.checkout?.step).toBe('complete');
  });

  it('does not attach an aborted submission error to a replacement checkout', async () => {
    let rejectSubmission: ((error: Error) => void) | undefined;
    const adapter: CheckoutMutationAdapter = {
      submitOrder: vi.fn(() => new Promise((_resolve, reject) => (rejectSubmission = reject))),
    };
    await renderProvider(adapter);
    await prepareCheckout();

    let pending!: Promise<unknown>;
    await act(async () => {
      pending = marketplace.actions.submitOrder().catch((error) => error);
      await Promise.resolve();
    });
    await act(async () => marketplace.actions.cancelCheckout());
    await act(async () =>
      marketplace.actions.startCheckout({
        offeringId: offering.id,
        durationSeconds: 7200,
      })
    );
    rejectSubmission?.(new Error('late failure'));
    expect(await pending).toBeDefined();

    expect(marketplace.state.checkout?.request.durationSeconds).toBe(7200);
    expect(marketplace.state.checkout?.error).toBeNull();
  });

  it('blocks submission and ignores older overlapping validation', async () => {
    await renderProvider({ submitOrder: vi.fn() });
    await act(async () => marketplace.actions.selectOffering(offering));
    await act(async () =>
      marketplace.actions.startCheckout({
        offeringId: offering.id,
        durationSeconds: 3600,
      })
    );
    let resolveFirst!: (value: ReturnType<typeof validationFor>) => void;
    let resolveSecond!: (value: ReturnType<typeof validationFor>) => void;
    query.mockImplementationOnce(() => new Promise((resolve) => (resolveFirst = resolve)));
    query.mockImplementationOnce(() => new Promise((resolve) => (resolveSecond = resolve)));

    let first!: Promise<unknown>;
    let second!: Promise<unknown>;
    await act(async () => {
      first = marketplace.actions.validateCheckout().catch((error) => error);
      second = marketplace.actions.validateCheckout().catch((error) => error);
      await Promise.resolve();
    });
    await expect(marketplace.actions.submitOrder()).rejects.toThrow(
      'Checkout validation in progress'
    );

    await act(async () => {
      resolveSecond(validationFor('30'));
      await second;
    });
    await act(async () => {
      resolveFirst(validationFor('25'));
      await first;
    });

    expect(marketplace.state.checkout?.validation?.priceQuote.totalAmount).toBe('30');
    expect(marketplace.state.checkout?.error).toBeNull();
  });
});
