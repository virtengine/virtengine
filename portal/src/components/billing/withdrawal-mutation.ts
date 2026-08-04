export interface BillingWithdrawalContext {
  readonly chainId: string;
  readonly accountAddress: string;
}

export interface BillingWithdrawalRequest {
  readonly action: 'request_withdrawal';
  readonly chainId: string;
  readonly accountAddress: string;
}

export interface BillingWithdrawalSubmission {
  requestDigest: string;
  idempotencyKey: string;
  signal: AbortSignal;
}

export interface BillingWithdrawalMutationResult {
  status: 'committed';
  code: number;
  txHash: string;
  blockHeight: number;
  operationId: string;
  requestDigest: string;
  idempotencyKey: string;
  request: BillingWithdrawalRequest;
}

export interface BillingWithdrawalAdapter {
  requestWithdrawal(
    request: Readonly<BillingWithdrawalRequest>,
    submission: Readonly<BillingWithdrawalSubmission>
  ): Promise<unknown>;
}

export class BillingWithdrawalError extends Error {
  constructor(
    public readonly reason:
      | 'unavailable'
      | 'invalid_request'
      | 'invalid_committed_result'
      | 'request_changed'
  ) {
    super(reason);
    this.name = 'BillingWithdrawalError';
  }
}

const normalizedText = (value: unknown): string | null => {
  if (typeof value !== 'string') return null;
  const normalized = value.trim();
  return normalized.length > 0 ? normalized : null;
};

export const buildBillingWithdrawalRequest = (
  context: BillingWithdrawalContext | undefined
): Readonly<BillingWithdrawalRequest> | null => {
  const chainId = normalizedText(context?.chainId);
  const accountAddress = normalizedText(context?.accountAddress);
  if (!chainId || !accountAddress) return null;
  return Object.freeze({ action: 'request_withdrawal', chainId, accountAddress });
};

export const billingWithdrawalRequestsEqual = (
  left: BillingWithdrawalRequest | null,
  right: BillingWithdrawalRequest | null
): boolean =>
  Boolean(
    left &&
    right &&
    left.action === right.action &&
    left.chainId === right.chainId &&
    left.accountAddress === right.accountAddress
  );

const snapshotRequest = (
  value: unknown,
  requireCanonical = false
): Readonly<BillingWithdrawalRequest> => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    throw new BillingWithdrawalError('invalid_request');
  }
  const source = value as Partial<BillingWithdrawalRequest>;
  const action = source.action;
  const sourceChainId = source.chainId;
  const sourceAccountAddress = source.accountAddress;
  const chainId = normalizedText(sourceChainId);
  const accountAddress = normalizedText(sourceAccountAddress);
  if (
    action !== 'request_withdrawal' ||
    !chainId ||
    !accountAddress ||
    (requireCanonical && (sourceChainId !== chainId || sourceAccountAddress !== accountAddress))
  ) {
    throw new BillingWithdrawalError('invalid_request');
  }
  return Object.freeze({ action, chainId, accountAddress });
};

const digestRequest = async (request: BillingWithdrawalRequest): Promise<string> => {
  const bytes = new TextEncoder().encode(
    JSON.stringify({
      accountAddress: request.accountAddress,
      action: request.action,
      chainId: request.chainId,
    })
  );
  const digest = await crypto.subtle.digest('SHA-256', bytes);
  return Array.from(new Uint8Array(digest), (byte) => byte.toString(16).padStart(2, '0')).join('');
};

const materializeCommittedResult = (
  value: unknown,
  request: BillingWithdrawalRequest,
  requestDigest: string
): Readonly<BillingWithdrawalMutationResult> => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    throw new BillingWithdrawalError('invalid_committed_result');
  }
  const result = value as Partial<BillingWithdrawalMutationResult>;
  const status = result.status;
  const code = result.code;
  const sourceTxHash = result.txHash;
  const txHash = normalizedText(sourceTxHash);
  const blockHeight = result.blockHeight;
  const sourceOperationId = result.operationId;
  const operationId = normalizedText(sourceOperationId);
  const committedDigest = result.requestDigest;
  const idempotencyKey = result.idempotencyKey;
  const committedRequestValue = result.request;
  let committedRequest: Readonly<BillingWithdrawalRequest>;
  try {
    committedRequest = snapshotRequest(committedRequestValue, true);
  } catch {
    throw new BillingWithdrawalError('invalid_committed_result');
  }
  if (
    status !== 'committed' ||
    code !== 0 ||
    !txHash ||
    sourceTxHash !== txHash ||
    !Number.isSafeInteger(blockHeight) ||
    Number(blockHeight) <= 0 ||
    !operationId ||
    sourceOperationId !== operationId ||
    committedDigest !== requestDigest ||
    idempotencyKey !== requestDigest ||
    !billingWithdrawalRequestsEqual(committedRequest, request)
  ) {
    throw new BillingWithdrawalError('invalid_committed_result');
  }
  return Object.freeze({
    status: 'committed',
    code: 0,
    txHash,
    blockHeight: blockHeight!,
    operationId,
    requestDigest,
    idempotencyKey: requestDigest,
    request,
  });
};

export const submitBillingWithdrawal = async ({
  adapter,
  request,
  signal,
  getCurrentRequest,
}: {
  adapter: BillingWithdrawalAdapter;
  request: BillingWithdrawalRequest;
  signal: AbortSignal;
  getCurrentRequest: () => BillingWithdrawalRequest | null;
}): Promise<Readonly<BillingWithdrawalMutationResult>> => {
  const requestSnapshot = snapshotRequest(request);
  const requestDigest = await digestRequest(requestSnapshot);
  if (signal.aborted || !billingWithdrawalRequestsEqual(getCurrentRequest(), requestSnapshot)) {
    throw new BillingWithdrawalError('request_changed');
  }
  const requestWithdrawal = Reflect.get(adapter, 'requestWithdrawal') as
    | BillingWithdrawalAdapter['requestWithdrawal']
    | undefined;
  if (
    typeof requestWithdrawal !== 'function' ||
    signal.aborted ||
    !billingWithdrawalRequestsEqual(getCurrentRequest(), requestSnapshot)
  ) {
    throw new BillingWithdrawalError('request_changed');
  }
  const value = await requestWithdrawal.call(adapter, requestSnapshot, {
    requestDigest,
    idempotencyKey: requestDigest,
    signal,
  });
  if (signal.aborted || !billingWithdrawalRequestsEqual(getCurrentRequest(), requestSnapshot)) {
    throw new BillingWithdrawalError('request_changed');
  }
  const result = materializeCommittedResult(value, requestSnapshot, requestDigest);
  if (signal.aborted || !billingWithdrawalRequestsEqual(getCurrentRequest(), requestSnapshot)) {
    throw new BillingWithdrawalError('request_changed');
  }
  return result;
};
