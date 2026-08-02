import { describe, expect, it, vi } from 'vitest';
import {
  buildEscrowMutationRequest,
  digestEscrowMutationRequest,
  EscrowMutationError,
  projectEscrowMutationResult,
  submitEscrowMutation,
  type EscrowCommittedResult,
  type EscrowMutationRequest,
  validateEscrowCommittedResult,
} from '@/components/escrow/mutations';

const context = {
  chainId: 'virtengine-1',
  accountAddress: 'virtengine1account',
  escrowAccountId: 'escrow-1',
};

type InvalidResultOverride = Partial<Omit<EscrowCommittedResult, 'code' | 'request'>> & {
  code?: number;
  request?: Partial<EscrowMutationRequest>;
};

const invalidResultOverrides: [string, InvalidResultOverride][] = [
  ['missing status', { status: undefined }],
  ['empty hash', { txHash: '' }],
  ['nonzero code', { code: 7 }],
  ['nonpositive height', { blockHeight: 0 }],
  ['missing operation id', { operationId: '' }],
  ['digest mismatch', { requestDigest: 'other' }],
  ['idempotency mismatch', { idempotencyKey: 'other' }],
  ['chain mismatch', { request: { chainId: 'other' } }],
  ['account mismatch', { request: { accountAddress: 'other' } }],
  ['escrow mismatch', { request: { escrowAccountId: 'other' } }],
  ['amount mismatch', { request: { amount: '51' } }],
  ['denom mismatch', { request: { denom: 'other' } }],
  ['destination mismatch', { request: { destination: 'other' } }],
];

describe('escrow mutations', () => {
  it('builds an exactly bound wallet request and validates committed evidence', async () => {
    const request = buildEscrowMutationRequest('deposit', context, '50.00', 'uve');
    expect(request).not.toBeNull();
    const requestDigest = await digestEscrowMutationRequest(request!);
    const result = {
      status: 'committed',
      txHash: 'ABC123',
      code: 0,
      blockHeight: 42,
      operationId: 'escrow-operation-1',
      requestDigest,
      idempotencyKey: requestDigest,
      request,
    };

    expect(validateEscrowCommittedResult(result, request!, requestDigest)).toEqual(result);
  });

  it('fails closed when request identity is incomplete', () => {
    expect(
      buildEscrowMutationRequest('withdraw', { ...context, chainId: '' }, '10', 'uve')
    ).toBeNull();
    expect(buildEscrowMutationRequest('withdraw', context, 'NaN', 'uve')).toBeNull();
  });

  it.each(invalidResultOverrides)('rejects %s', async (_name, override) => {
    const request = buildEscrowMutationRequest('withdraw', context, '50', 'uve')!;
    const requestDigest = await digestEscrowMutationRequest(request);
    const result = {
      status: 'committed',
      txHash: 'ABC123',
      code: 0,
      blockHeight: 42,
      operationId: 'escrow-operation-1',
      requestDigest,
      idempotencyKey: requestDigest,
      request,
      ...override,
      ...(override.request ? { request: { ...request, ...override.request } } : {}),
    };

    expect(() => validateEscrowCommittedResult(result, request, requestDigest)).toThrow(
      EscrowMutationError
    );
  });

  it('fails closed when the result projector throws', () => {
    expect(() =>
      projectEscrowMutationResult(() => {
        throw new Error('malformed');
      }, {})
    ).toThrow(EscrowMutationError);
  });

  it('rejects changed request state after submission', async () => {
    const request = buildEscrowMutationRequest('deposit', context, '50', 'uve')!;
    const adapter = {
      mutate: vi.fn(),
    };

    await expect(
      submitEscrowMutation({
        adapter,
        projector: (value) => value,
        request,
        getCurrentRequest: () => ({ ...request, amount: '51' }),
      })
    ).rejects.toMatchObject({ code: 'request_changed' });
    expect(adapter.mutate).not.toHaveBeenCalled();
  });

  it('maps adapter rejection to a closed failure', async () => {
    const request = buildEscrowMutationRequest('withdraw', context, '50', 'uve')!;
    await expect(
      submitEscrowMutation({
        adapter: { mutate: vi.fn().mockRejectedValue(new Error('broadcast failed')) },
        projector: (value) => value,
        request,
        getCurrentRequest: () => request,
      })
    ).rejects.toMatchObject({ code: 'submission_failed' });
  });

  it('freezes the request snapshot before invoking the adapter', async () => {
    const request = buildEscrowMutationRequest('withdraw', context, '50', 'uve')!;
    let receivedRequest: EscrowMutationRequest | undefined;

    await expect(
      submitEscrowMutation({
        adapter: {
          mutate: vi.fn((adapterRequest) => {
            receivedRequest = adapterRequest;
            adapterRequest.amount = '999';
            return Promise.resolve({});
          }),
        },
        projector: (value) => value,
        request,
        getCurrentRequest: () => request,
      })
    ).rejects.toMatchObject({ code: 'submission_failed' });
    expect(Object.isFrozen(receivedRequest)).toBe(true);
    expect(request.amount).toBe('50');
  });

  it('rejects caller cancellation that races with adapter completion', async () => {
    const request = buildEscrowMutationRequest('withdraw', context, '50', 'uve')!;
    const controller = new AbortController();

    await expect(
      submitEscrowMutation({
        adapter: {
          mutate: vi.fn((_request, submission) => {
            controller.abort();
            return Promise.resolve({
              status: 'committed',
              txHash: 'ABC123',
              code: 0,
              blockHeight: 42,
              operationId: 'operation-1',
              requestDigest: submission.requestDigest,
              idempotencyKey: submission.idempotencyKey,
              request,
            });
          }),
        },
        projector: (value) => value,
        request,
        getCurrentRequest: () => request,
        signal: controller.signal,
      })
    ).rejects.toMatchObject({ code: 'submission_cancelled' });
  });

  it('aborts and rejects submissions that time out', async () => {
    const request = buildEscrowMutationRequest('withdraw', context, '50', 'uve')!;
    let signal: AbortSignal | undefined;
    const result = submitEscrowMutation({
      adapter: {
        mutate: vi.fn((_request, submission) => {
          signal = submission.signal;
          return new Promise(() => undefined);
        }),
      },
      projector: (value) => value,
      request,
      getCurrentRequest: () => request,
      timeoutMs: 5,
    });

    await expect(result).rejects.toMatchObject({ code: 'submission_timeout' });
    expect(signal?.aborted).toBe(true);
  });
});
