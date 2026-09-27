import * as React from 'react';
import { act } from 'react';
import { createRoot, type Root } from 'react-dom/client';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { CheckoutFlow } from '../../components/marketplace/CheckoutFlow';
import {
  buildCheckoutMutationRequest,
  CheckoutMutationError,
  digestCheckoutMutationRequest,
  submitCheckoutMutation,
  validateCheckoutCommittedResult,
  type CheckoutMutationAdapter,
} from '../../components/marketplace/checkout-mutation';
import type { Offering } from '../../types/marketplace';

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
  identityRequirements: {
    minScore: 0,
    requiredScopes: [],
    mfaRequired: false,
  },
  reliabilityScore: 99,
  benchmarkSummary: {
    overallScore: 99,
    cpuScore: 99,
    memoryScore: 99,
    storageScore: 99,
    networkScore: 99,
    lastBenchmarkAt: 1,
    suiteVersion: '1',
  },
  createdAt: 1,
  updatedAt: 1,
  hasEncryptedDetails: false,
};

function clickButton(container: HTMLElement, label: string): void {
  const button = Array.from(container.querySelectorAll('button')).find(
    (candidate) => candidate.textContent === label
  );
  if (!button) throw new Error(`missing button: ${label}`);
  button.click();
}

// CheckoutFlow hashes the request with crypto.subtle.digest before it reaches
// the adapter, so `await act()` (which only drains microtasks) can return while
// the submission is still queued on the digest threadpool. A test that re-renders
// straight after clicking therefore races the submission: the effect sees no
// active submission to abort and the adapter is never called. Wait for the
// adapter to actually be reached before asserting on its signal.
async function waitForSubmission(adapter: CheckoutMutationAdapter): Promise<void> {
  for (let attempt = 0; attempt < 200; attempt += 1) {
    if (vi.mocked(adapter.submitOrder).mock.calls.length > 0) return;
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 5));
    });
  }
  throw new Error('timed out waiting for the checkout submission to reach the adapter');
}

// Once the adapter has been reached, the commit result still has to travel back
// through submitCheckoutMutation's digest/race/validation awaits before the
// component projects it and calls onComplete. Poll the terminal signal rather
// than assuming one act() flush covers it.
async function waitForCondition(
  condition: () => boolean,
  description: string
): Promise<void> {
  for (let attempt = 0; attempt < 200; attempt += 1) {
    if (condition()) return;
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 5));
    });
  }
  throw new Error(`timed out waiting for ${description}`);
}

async function waitForCompletion(onComplete: { mock: { calls: unknown[][] } }): Promise<void> {
  await waitForCondition(
    () => onComplete.mock.calls.length > 0,
    'the checkout flow to call onComplete'
  );
}

// Deterministic regression guard for the digest/submitOrder race. CheckoutFlow
// hashes the request with a real crypto.subtle.digest before the adapter is
// reached, and `await act()` does not wait on that WebCrypto work, so every
// test above that clicks "Place Order" and then asserts on post-digest state is
// a latent race. Slowing the digest by 25ms reproduces the CI failure
// (submissionSignal stays undefined / onComplete is never called) on any machine,
// which is what the surrounding suite silently depends on not happening.
async function slowCheckoutDigest(delayMs = 25): Promise<() => void> {
  const subtle = globalThis.crypto.subtle;
  const realDigest = subtle.digest.bind(subtle);
  const spy = vi
    .spyOn(subtle, 'digest')
    .mockImplementation((async (...args: Parameters<typeof realDigest>) => {
      await new Promise((resolve) => setTimeout(resolve, delayMs));
      return realDigest(...args);
    }) as typeof realDigest);
  return () => spy.mockRestore();
}

describe('CheckoutFlow', () => {
  let container: HTMLDivElement;
  let root: Root;

  beforeEach(() => {
    globalThis.IS_REACT_ACT_ENVIRONMENT = true;
    container = document.createElement('div');
    document.body.appendChild(container);
    root = createRoot(container);
  });

  afterEach(async () => {
    await act(async () => root.unmount());
    container.remove();
  });

  it('does not complete without an authoritative adapter', async () => {
    const onComplete = vi.fn();
    await act(async () =>
      root.render(<CheckoutFlow offering={offering} onComplete={onComplete} />)
    );
    const checkbox = container.querySelector('input[type=checkbox]') as HTMLInputElement;
    await act(async () => checkbox.click());
    await act(async () => clickButton(container, 'Continue'));

    const placeOrder = Array.from(container.querySelectorAll('button')).find(
      (button) => button.textContent === 'Place Order'
    ) as HTMLButtonElement;
    expect(placeOrder.disabled).toBe(true);
    expect(onComplete).not.toHaveBeenCalled();
    expect(container.textContent).toContain('Order placement is unavailable');
  });

  it('completes only from exact committed evidence', async () => {
    const onComplete = vi.fn();
    const adapter: CheckoutMutationAdapter = {
      submitOrder: vi.fn((request, submission) =>
        Promise.resolve({
          status: 'committed',
          code: 0,
          orderId: 'order-authoritative-1',
          txHash: 'ABC123',
          blockHeight: 42,
          requestDigest: submission.requestDigest,
          idempotencyKey: submission.idempotencyKey,
          request,
        })
      ),
    };
    await act(async () =>
      root.render(
        <CheckoutFlow
          offering={offering}
          onComplete={onComplete}
          mutationAdapter={adapter}
          mutationContext={{
            chainId: 'virtengine-1',
            customerAddress: 'virtengine1customer',
          }}
          resultProjector={(value) => value}
        />
      )
    );
    const checkbox = container.querySelector('input[type=checkbox]') as HTMLInputElement;
    await act(async () => checkbox.click());
    await act(async () => clickButton(container, 'Continue'));
    await act(async () => clickButton(container, 'Place Order'));

    await waitForCompletion(onComplete);
    expect(onComplete).toHaveBeenCalledWith('order-authoritative-1');
    expect(container.textContent).toContain('Order committed');
  });

  it('rejects malformed or mismatched committed evidence', async () => {
    const request = buildCheckoutMutationRequest(
      { chainId: 'virtengine-1', customerAddress: 'virtengine1customer' },
      {
        offeringId: offering.id,
        providerAddress: offering.providerAddress,
        durationSeconds: offering.pricing.minDurationSeconds,
        priceAmount: offering.pricing.basePrice!,
        depositAmount: offering.pricing.depositRequired,
        priceDenom: offering.pricing.denom,
      }
    )!;
    const requestDigest = await digestCheckoutMutationRequest(request);
    expect(() =>
      validateCheckoutCommittedResult(
        {
          status: 'committed',
          code: 0,
          orderId: 'order-1',
          txHash: 'ABC123',
          blockHeight: 42,
          requestDigest,
          idempotencyKey: requestDigest,
          request: { ...request, offeringId: 'other' },
        },
        request,
        requestDigest
      )
    ).toThrow(CheckoutMutationError);
  });

  it('returns immutable materialized committed evidence', async () => {
    const request = buildCheckoutMutationRequest(
      { chainId: 'virtengine-1', customerAddress: 'virtengine1customer' },
      {
        offeringId: offering.id,
        providerAddress: offering.providerAddress,
        durationSeconds: offering.pricing.minDurationSeconds,
        priceAmount: offering.pricing.basePrice!,
        depositAmount: offering.pricing.depositRequired,
        priceDenom: offering.pricing.denom,
      }
    )!;
    const requestDigest = await digestCheckoutMutationRequest(request);
    const source = {
      status: 'committed',
      code: 0,
      orderId: 'order-1',
      txHash: 'ABC123',
      blockHeight: 42,
      requestDigest,
      idempotencyKey: requestDigest,
      request,
    };
    const committed = validateCheckoutCommittedResult(source, request, requestDigest);
    source.orderId = 'mutated';
    request.offeringId = 'mutated';

    expect(committed.orderId).toBe('order-1');
    expect(committed.request.offeringId).toBe('offering-1');
    expect(Object.isFrozen(committed)).toBe(true);
    expect(Object.isFrozen(committed.request)).toBe(true);
  });

  it('rejects request drift before invoking the adapter', async () => {
    const request = buildCheckoutMutationRequest(
      { chainId: 'virtengine-1', customerAddress: 'virtengine1customer' },
      {
        offeringId: offering.id,
        providerAddress: offering.providerAddress,
        durationSeconds: offering.pricing.minDurationSeconds,
        priceAmount: offering.pricing.basePrice!,
        depositAmount: offering.pricing.depositRequired,
        priceDenom: offering.pricing.denom,
      }
    )!;
    const adapter: CheckoutMutationAdapter = { submitOrder: vi.fn() };

    await expect(
      submitCheckoutMutation({
        adapter,
        projector: (value) => value,
        request,
        getCurrentRequest: () => ({ ...request, durationSeconds: 7200 }),
      })
    ).rejects.toMatchObject({ code: 'request_changed' });
    expect(adapter.submitOrder).not.toHaveBeenCalled();
  });

  it('aborts an active submission and resets when authority-bound props change', async () => {
    let submissionSignal: AbortSignal | undefined;
    const adapter: CheckoutMutationAdapter = {
      submitOrder: vi.fn((_request, submission) => {
        submissionSignal = submission.signal;
        return new Promise(() => undefined);
      }),
    };
    const context = {
      chainId: 'virtengine-1',
      customerAddress: 'virtengine1customer',
    };
    const projector = (value: unknown) => value;
    await act(async () =>
      root.render(
        <CheckoutFlow
          offering={offering}
          onComplete={vi.fn()}
          mutationAdapter={adapter}
          mutationContext={context}
          resultProjector={projector}
        />
      )
    );
    const checkbox = container.querySelector('input[type=checkbox]') as HTMLInputElement;
    await act(async () => checkbox.click());
    await act(async () => clickButton(container, 'Continue'));
    await act(async () => clickButton(container, 'Place Order'));
    expect(container.textContent).toContain('Processing your order');
    await waitForSubmission(adapter);

    await act(async () =>
      root.render(
        <CheckoutFlow
          offering={{ ...offering, id: 'offering-2' }}
          onComplete={vi.fn()}
          mutationAdapter={adapter}
          mutationContext={context}
          resultProjector={projector}
        />
      )
    );

    expect(submissionSignal?.aborted).toBe(true);
    expect(container.textContent).toContain('Confirm Your Order');
    expect(container.textContent).not.toContain('Processing your order');
  });

  it('aborts an active submission when the bound deposit changes', async () => {
    let submissionSignal: AbortSignal | undefined;
    const adapter: CheckoutMutationAdapter = {
      submitOrder: vi.fn((_request, submission) => {
        submissionSignal = submission.signal;
        return new Promise(() => undefined);
      }),
    };
    const context = {
      chainId: 'virtengine-1',
      customerAddress: 'virtengine1customer',
    };
    const projector = (value: unknown) => value;
    const props = {
      onComplete: vi.fn(),
      mutationAdapter: adapter,
      mutationContext: context,
      resultProjector: projector,
    };
    await act(async () => root.render(<CheckoutFlow offering={offering} {...props} />));
    const checkbox = container.querySelector('input[type=checkbox]') as HTMLInputElement;
    await act(async () => checkbox.click());
    await act(async () => clickButton(container, 'Continue'));
    await act(async () => clickButton(container, 'Place Order'));
    await waitForSubmission(adapter);

    await act(async () =>
      root.render(
        <CheckoutFlow
          offering={{
            ...offering,
            pricing: { ...offering.pricing, depositRequired: '75' },
          }}
          {...props}
        />
      )
    );

    expect(submissionSignal?.aborted).toBe(true);
    expect(container.textContent).toContain('Confirm Your Order');
  });
});

// Regression guard for the digest/submitOrder race that made this file flaky in
// CI (red on develop run 36293470613). Each test below replays one of the three
// scenarios that raced the pending crypto.subtle.digest, with the digest
// deliberately slowed by 25ms so the window is guaranteed to be open on every
// machine. If a future edit drops a wait between "Place Order" and the
// post-digest assertion, these fail deterministically instead of intermittently
// on a loaded CI runner.
describe('CheckoutFlow submission race (slowed request digest)', () => {
  let container: HTMLDivElement;
  let root: Root;
  let restoreDigest: (() => void) | undefined;

  beforeEach(() => {
    globalThis.IS_REACT_ACT_ENVIRONMENT = true;
    container = document.createElement('div');
    document.body.appendChild(container);
    root = createRoot(container);
  });

  afterEach(async () => {
    restoreDigest?.();
    restoreDigest = undefined;
    await act(async () => root.unmount());
    container.remove();
    vi.restoreAllMocks();
  });

  const context = {
    chainId: 'virtengine-1',
    customerAddress: 'virtengine1customer',
  };
  const projector = (value: unknown) => value;

  async function placeOrder(): Promise<void> {
    const checkbox = container.querySelector('input[type=checkbox]') as HTMLInputElement;
    await act(async () => checkbox.click());
    await act(async () => clickButton(container, 'Continue'));
    await act(async () => clickButton(container, 'Place Order'));
  }

  it('commits only after the slowed digest resolves', async () => {
    restoreDigest = await slowCheckoutDigest();
    const onComplete = vi.fn();
    const adapter: CheckoutMutationAdapter = {
      submitOrder: vi.fn((request, submission) =>
        Promise.resolve({
          status: 'committed',
          code: 0,
          orderId: 'order-slow-digest-1',
          txHash: 'ABC123',
          blockHeight: 42,
          requestDigest: submission.requestDigest,
          idempotencyKey: submission.idempotencyKey,
          request,
        })
      ),
    };
    await act(async () =>
      root.render(
        <CheckoutFlow
          offering={offering}
          onComplete={onComplete}
          mutationAdapter={adapter}
          mutationContext={context}
          resultProjector={projector}
        />
      )
    );
    await placeOrder();

    await waitForCompletion(onComplete);
    expect(adapter.submitOrder).toHaveBeenCalledTimes(1);
    expect(onComplete).toHaveBeenCalledWith('order-slow-digest-1');
    expect(container.textContent).toContain('Order committed');
  });

  it('aborts the active submission when authority-bound props change', async () => {
    restoreDigest = await slowCheckoutDigest();
    let submissionSignal: AbortSignal | undefined;
    const adapter: CheckoutMutationAdapter = {
      submitOrder: vi.fn((_request, submission) => {
        submissionSignal = submission.signal;
        return new Promise(() => undefined);
      }),
    };
    await act(async () =>
      root.render(
        <CheckoutFlow
          offering={offering}
          onComplete={vi.fn()}
          mutationAdapter={adapter}
          mutationContext={context}
          resultProjector={projector}
        />
      )
    );
    await placeOrder();
    await waitForSubmission(adapter);

    await act(async () =>
      root.render(
        <CheckoutFlow
          offering={{ ...offering, id: 'offering-2' }}
          onComplete={vi.fn()}
          mutationAdapter={adapter}
          mutationContext={context}
          resultProjector={projector}
        />
      )
    );

    expect(submissionSignal?.aborted).toBe(true);
    expect(container.textContent).toContain('Confirm Your Order');
    expect(container.textContent).not.toContain('Processing your order');
  });

  it('aborts the active submission when the bound deposit changes', async () => {
    restoreDigest = await slowCheckoutDigest();
    let submissionSignal: AbortSignal | undefined;
    const adapter: CheckoutMutationAdapter = {
      submitOrder: vi.fn((_request, submission) => {
        submissionSignal = submission.signal;
        return new Promise(() => undefined);
      }),
    };
    const props = {
      onComplete: vi.fn(),
      mutationAdapter: adapter,
      mutationContext: context,
      resultProjector: projector,
    };
    await act(async () => root.render(<CheckoutFlow offering={offering} {...props} />));
    await placeOrder();
    await waitForSubmission(adapter);

    await act(async () =>
      root.render(
        <CheckoutFlow
          offering={{ ...offering, pricing: { ...offering.pricing, depositRequired: '75' } }}
          {...props}
        />
      )
    );

    expect(submissionSignal?.aborted).toBe(true);
    expect(container.textContent).toContain('Confirm Your Order');
  });
});
