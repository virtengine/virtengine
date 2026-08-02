export type EscrowMutationStatus = 'idle' | 'submitting' | 'committed' | 'error' | 'unavailable';

export type EscrowMutationRequest = {
  operation: 'deposit' | 'withdraw';
  chainId: string;
  accountAddress: string;
  escrowAccountId: string;
  amount: string;
  denom: string;
  destination: string;
};

export type EscrowMutationSubmission = {
  requestDigest: string;
  idempotencyKey: string;
  signal: AbortSignal;
};

export interface EscrowMutationAdapter {
  mutate(request: EscrowMutationRequest, submission: EscrowMutationSubmission): Promise<unknown>;
}

export type EscrowMutationResultProjector = (result: unknown) => unknown;

export type EscrowCommittedResult = {
  status: 'committed';
  txHash: string;
  code: 0;
  blockHeight: number;
  operationId: string;
  requestDigest: string;
  idempotencyKey: string;
  request: EscrowMutationRequest;
};

export type EscrowMutationContext = {
  chainId: string;
  accountAddress: string;
  escrowAccountId: string;
};

export class EscrowMutationError extends Error {
  constructor(
    readonly code:
      | 'feature_unavailable'
      | 'submission_cancelled'
      | 'submission_timeout'
      | 'request_changed'
      | 'invalid_committed_result'
      | 'submission_failed'
  ) {
    super(code);
    this.name = 'EscrowMutationError';
  }
}

function canonicalRequest(request: EscrowMutationRequest): string {
  return JSON.stringify({
    operation: request.operation,
    chainId: request.chainId,
    accountAddress: request.accountAddress,
    escrowAccountId: request.escrowAccountId,
    amount: request.amount,
    denom: request.denom,
    destination: request.destination,
  });
}

export function escrowMutationRequestsEqual(
  left: EscrowMutationRequest | null,
  right: EscrowMutationRequest | null
): boolean {
  return Boolean(left && right && canonicalRequest(left) === canonicalRequest(right));
}

export function isValidEscrowMutationAmount(amount: string): boolean {
  const normalizedAmount = amount.trim();
  const parsedAmount = Number(normalizedAmount);
  return (
    Boolean(normalizedAmount) &&
    /^\d+(?:\.\d+)?$/.test(normalizedAmount) &&
    Number.isFinite(parsedAmount) &&
    parsedAmount > 0
  );
}

export function isValidEscrowMutationContext(
  context: EscrowMutationContext | undefined
): context is EscrowMutationContext {
  return Boolean(
    context?.chainId.trim() && context.accountAddress.trim() && context.escrowAccountId.trim()
  );
}

export function buildEscrowMutationRequest(
  operation: EscrowMutationRequest['operation'],
  context: EscrowMutationContext,
  amount: string,
  denom: string
): EscrowMutationRequest | null {
  const normalizedAmount = amount.trim();
  if (
    !context.chainId.trim() ||
    !context.accountAddress.trim() ||
    !context.escrowAccountId.trim() ||
    !isValidEscrowMutationAmount(normalizedAmount) ||
    !denom.trim()
  ) {
    return null;
  }

  return {
    operation,
    chainId: context.chainId,
    accountAddress: context.accountAddress,
    escrowAccountId: context.escrowAccountId,
    amount: normalizedAmount,
    denom,
    destination: operation === 'deposit' ? context.escrowAccountId : context.accountAddress,
  };
}

export async function digestEscrowMutationRequest(request: EscrowMutationRequest): Promise<string> {
  const bytes = new TextEncoder().encode(canonicalRequest(request));
  const digest = await globalThis.crypto.subtle.digest('SHA-256', bytes);
  return Array.from(new Uint8Array(digest), (byte) => byte.toString(16).padStart(2, '0')).join('');
}

export function projectEscrowMutationResult(
  projector: EscrowMutationResultProjector,
  result: unknown
): unknown {
  try {
    return projector(result);
  } catch {
    throw new EscrowMutationError('invalid_committed_result');
  }
}

export function validateEscrowCommittedResult(
  value: unknown,
  request: EscrowMutationRequest,
  requestDigest: string
): EscrowCommittedResult {
  if (!value || typeof value !== 'object') {
    throw new EscrowMutationError('invalid_committed_result');
  }

  const result = value as Partial<EscrowCommittedResult>;
  let valid = false;
  try {
    valid =
      result.status === 'committed' &&
      result.code === 0 &&
      typeof result.txHash === 'string' &&
      result.txHash.trim().length > 0 &&
      typeof result.blockHeight === 'number' &&
      Number.isInteger(result.blockHeight) &&
      result.blockHeight > 0 &&
      typeof result.operationId === 'string' &&
      result.operationId.trim().length > 0 &&
      result.requestDigest === requestDigest &&
      result.idempotencyKey === requestDigest &&
      result.request !== undefined &&
      canonicalRequest(result.request) === canonicalRequest(request);
  } catch {
    valid = false;
  }

  if (!valid) throw new EscrowMutationError('invalid_committed_result');
  return result as EscrowCommittedResult;
}

export async function submitEscrowMutation({
  adapter,
  projector,
  request,
  getCurrentRequest,
  signal,
  timeoutMs = 30_000,
}: {
  adapter: EscrowMutationAdapter;
  projector: EscrowMutationResultProjector;
  request: EscrowMutationRequest;
  getCurrentRequest: () => EscrowMutationRequest | null;
  signal?: AbortSignal;
  timeoutMs?: number;
}): Promise<EscrowCommittedResult> {
  const submittedRequest: EscrowMutationRequest = Object.freeze({ ...request });
  const requestDigest = await digestEscrowMutationRequest(submittedRequest);
  const currentRequestBeforeSubmit = getCurrentRequest();
  if (
    signal?.aborted ||
    !currentRequestBeforeSubmit ||
    canonicalRequest(currentRequestBeforeSubmit) !== canonicalRequest(submittedRequest)
  ) {
    throw new EscrowMutationError('request_changed');
  }
  const abortController = new AbortController();
  let timeout: ReturnType<typeof setTimeout> | undefined;
  let externalAbortHandler: (() => void) | undefined;

  try {
    const cancellation = new Promise<never>((_, reject) => {
      if (!signal) return;
      externalAbortHandler = () => {
        abortController.abort();
        reject(new EscrowMutationError('submission_cancelled'));
      };
      signal.addEventListener('abort', externalAbortHandler, { once: true });
    });
    const rawResult = await Promise.race([
      adapter.mutate(submittedRequest, {
        requestDigest,
        idempotencyKey: requestDigest,
        signal: abortController.signal,
      }),
      new Promise<never>((_, reject) => {
        timeout = setTimeout(() => {
          abortController.abort();
          reject(new EscrowMutationError('submission_timeout'));
        }, timeoutMs);
      }),
      cancellation,
    ]);
    if (signal?.aborted) throw new EscrowMutationError('submission_cancelled');
    const currentRequest = getCurrentRequest();
    if (
      !currentRequest ||
      canonicalRequest(currentRequest) !== canonicalRequest(submittedRequest)
    ) {
      throw new EscrowMutationError('request_changed');
    }
    return validateEscrowCommittedResult(
      projectEscrowMutationResult(projector, rawResult),
      submittedRequest,
      requestDigest
    );
  } catch (error) {
    if (error instanceof EscrowMutationError) throw error;
    throw new EscrowMutationError('submission_failed');
  } finally {
    if (timeout) clearTimeout(timeout);
    if (externalAbortHandler) signal?.removeEventListener('abort', externalAbortHandler);
  }
}
