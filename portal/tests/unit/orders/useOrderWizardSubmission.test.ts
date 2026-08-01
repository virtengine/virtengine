import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import {
  buildOrderCreateRequest,
  digestOrderCreateRequest,
  useOrderWizard,
  type OrderCreateRequest,
  type OrderSubmissionAdapter,
  type OrderSubmissionContext,
} from '@/features/orders';
import type { Offering } from '@/types/offerings';

const offering: Offering = {
  id: { providerAddress: 'provider-1', sequence: 7 },
  state: 'active',
  category: 'compute',
  name: 'Compute',
  description: '',
  version: '1',
  pricing: { currency: 'VE', model: 'hourly', basePrice: '1' },
  prices: [{ resourceType: 'cpu', unit: 'vcpu-hour', price: { denom: 'uve', amount: '1' } }],
  allowBidding: false,
  identityRequirement: {
    minScore: 0,
    requiredStatus: '',
    requireVerifiedEmail: false,
    requireVerifiedDomain: false,
    requireMFA: false,
  },
  requireMFAForOrders: false,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
  totalOrderCount: 0,
  activeOrderCount: 0,
};

function renderSubmission(adapter?: OrderSubmissionAdapter, onComplete = vi.fn()) {
  return {
    onComplete,
    ...renderHook(() =>
      useOrderWizard({
        offering,
        walletBalance: '10000000000',
        submissionAdapter: adapter,
        resultProjector: adapter ? (value) => value : undefined,
        onComplete,
      })
    ),
  };
}

async function advanceToEscrow(result: ReturnType<typeof renderSubmission>['result']) {
  act(() => result.current.nextStep());
  act(() => result.current.nextStep());
}

async function committedFor(request: OrderCreateRequest) {
  return {
    status: 'committed' as const,
    code: 0 as const,
    orderId: 'authority-order-9',
    txHash: 'ABC123',
    blockHeight: 99,
    requestDigest: await digestOrderCreateRequest(request),
    request,
  };
}

describe('useOrderWizard authoritative submission', () => {
  it('fails closed with typed feature_unavailable', async () => {
    const { result, onComplete } = renderSubmission();
    await advanceToEscrow(result);

    await act(() => result.current.submitOrder());

    expect(result.current.state.error).toBe('feature_unavailable');
    expect(result.current.state.currentStep).toBe('escrow');
    expect(onComplete).not.toHaveBeenCalled();
  });

  it('shows loading and prevents duplicate concurrent submission', async () => {
    let resolveSubmission!: (value: unknown) => void;
    const submitOrder = vi.fn(() => new Promise<unknown>((resolve) => (resolveSubmission = resolve)));
    const { result } = renderSubmission({ submitOrder });
    await advanceToEscrow(result);

    act(() => void result.current.submitOrder());
    await waitFor(() => expect(result.current.state.isSubmitting).toBe(true));
    await act(() => result.current.submitOrder());
    expect(submitOrder).toHaveBeenCalledOnce();

    const request = buildOrderCreateRequest(result.current.state)!;
    await act(async () => resolveSubmission(await committedFor(request)));
  });

  it('rejects adapter failures without completion', async () => {
    const onComplete = vi.fn();
    const { result } = renderSubmission(
      { submitOrder: vi.fn().mockRejectedValue(new Error('wallet rejected')) },
      onComplete
    );
    await advanceToEscrow(result);

    await act(() => result.current.submitOrder());

    expect(result.current.state.error).toBe('submission_rejected');
    expect(result.current.state.currentStep).toBe('escrow');
    expect(onComplete).not.toHaveBeenCalled();
  });

  it('aborts a timed-out submission without completion', async () => {
    vi.useFakeTimers();
    try {
      const submitOrder = vi.fn(
        (_request: OrderCreateRequest, _context: OrderSubmissionContext) =>
          new Promise<unknown>(() => undefined)
      );
      const { result, onComplete } = renderSubmission({ submitOrder });
      await advanceToEscrow(result);

      let submission!: Promise<void>;
      act(() => {
        submission = result.current.submitOrder();
      });
      await act(async () => {
        await Promise.resolve();
        await Promise.resolve();
      });
      const context = submitOrder.mock.calls[0][1];
      await act(async () => {
        await vi.advanceTimersByTimeAsync(30_000);
        await submission;
      });

      expect(context.signal.aborted).toBe(true);
      expect(result.current.state.error).toBe('submission_timeout');
      expect(result.current.state.currentStep).toBe('escrow');
      expect(onComplete).not.toHaveBeenCalled();
    } finally {
      vi.useRealTimers();
    }
  });

  it('rejects a committed result after request state changes', async () => {
    let resolveSubmission!: (value: unknown) => void;
    const submitOrder = vi.fn(() => new Promise<unknown>((resolve) => (resolveSubmission = resolve)));
    const { result, onComplete } = renderSubmission({ submitOrder });
    await advanceToEscrow(result);
    const submittedRequest = buildOrderCreateRequest(result.current.state)!;

    act(() => void result.current.submitOrder());
    await waitFor(() => expect(submitOrder).toHaveBeenCalledOnce());
    act(() => result.current.setResources({ cpu: submittedRequest.resources.cpu + 1 }));
    await act(async () => resolveSubmission(await committedFor(submittedRequest)));

    expect(result.current.state.error).toBe('order_state_changed');
    expect(result.current.state.currentStep).toBe('escrow');
    expect(onComplete).not.toHaveBeenCalled();
  });

  it('accepts only an exactly bound committed result', async () => {
    const submitOrder = vi.fn(
      async (request: OrderCreateRequest, _context: OrderSubmissionContext) =>
        committedFor(request)
    );
    const { result, onComplete } = renderSubmission({ submitOrder });
    await advanceToEscrow(result);

    await act(() => result.current.submitOrder());

    expect(result.current.state.currentStep).toBe('confirmation');
    expect(result.current.state.orderResult).toMatchObject({
      status: 'committed',
      code: 0,
      orderId: 'authority-order-9',
      txHash: 'ABC123',
      blockHeight: 99,
    });
    const [request, context] = submitOrder.mock.calls[0];
    expect(request).toEqual(buildOrderCreateRequest(result.current.state));
    expect(request.offeringId).toEqual({ providerAddress: 'provider-1', sequence: 7 });
    expect(context.idempotencyKey).toBe(context.requestDigest);
    expect(context.requestDigest).toBe(await digestOrderCreateRequest(request));
    expect(onComplete).toHaveBeenCalledOnce();
  });

  it.each([
    { code: 2 },
    { orderId: '' },
    { txHash: '' },
    { blockHeight: 0 },
    { status: 'pending' },
    { requestDigest: 'wrong' },
    { request: { offeringId: { providerAddress: 'other', sequence: 7 } } },
  ])('rejects malformed or mismatched authority: %o', async (override) => {
    const submitOrder = vi.fn(async (request: OrderCreateRequest) => ({
      ...(await committedFor(request)),
      ...override,
    }));
    const { result, onComplete } = renderSubmission({ submitOrder });
    await advanceToEscrow(result);

    await act(() => result.current.submitOrder());

    expect(result.current.state.error).toBe('invalid_committed_result');
    expect(result.current.state.currentStep).toBe('escrow');
    expect(onComplete).not.toHaveBeenCalled();
  });
});