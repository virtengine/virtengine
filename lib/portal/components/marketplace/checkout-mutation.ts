export interface CheckoutMutationContext {
  chainId: string;
  customerAddress: string;
}

export interface CheckoutMutationRequest {
  chainId: string;
  customerAddress: string;
  offeringId: string;
  providerAddress: string;
  durationSeconds: number;
  priceAmount: string;
  priceDenom: string;
}

export interface CheckoutMutationSubmission {
  requestDigest: string;
  idempotencyKey: string;
  signal: AbortSignal;
}

export interface CheckoutMutationAdapter {
  submitOrder(
    request: CheckoutMutationRequest,
    submission: CheckoutMutationSubmission,
  ): Promise<unknown>;
}

export type CheckoutMutationProjector = (result: unknown) => unknown;

export interface CheckoutCommittedResult {
  status: "committed";
  code: 0;
  orderId: string;
  txHash: string;
  blockHeight: number;
  requestDigest: string;
  idempotencyKey: string;
  request: CheckoutMutationRequest;
}

export class CheckoutMutationError extends Error {
  constructor(
    readonly code:
      | "submission_cancelled"
      | "submission_timeout"
      | "request_changed"
      | "invalid_committed_result"
      | "submission_failed",
  ) {
    super(code);
    this.name = "CheckoutMutationError";
  }
}

function canonicalRequest(request: CheckoutMutationRequest): string {
  return JSON.stringify({
    chainId: request.chainId,
    customerAddress: request.customerAddress,
    offeringId: request.offeringId,
    providerAddress: request.providerAddress,
    durationSeconds: request.durationSeconds,
    priceAmount: request.priceAmount,
    priceDenom: request.priceDenom,
  });
}

export function checkoutMutationRequestsEqual(
  left: CheckoutMutationRequest | null,
  right: CheckoutMutationRequest | null,
): boolean {
  return Boolean(
    left && right && canonicalRequest(left) === canonicalRequest(right),
  );
}

export function buildCheckoutMutationRequest(
  context: CheckoutMutationContext | undefined,
  input: Omit<CheckoutMutationRequest, "chainId" | "customerAddress">,
): CheckoutMutationRequest | null {
  if (
    !context?.chainId.trim() ||
    !context.customerAddress.trim() ||
    !input.offeringId.trim() ||
    !input.providerAddress.trim() ||
    !Number.isInteger(input.durationSeconds) ||
    input.durationSeconds <= 0 ||
    !/^\d+(?:\.\d+)?$/.test(input.priceAmount) ||
    Number(input.priceAmount) <= 0 ||
    !input.priceDenom.trim()
  ) {
    return null;
  }

  return {
    ...input,
    chainId: context.chainId,
    customerAddress: context.customerAddress,
  };
}

export async function digestCheckoutMutationRequest(
  request: CheckoutMutationRequest,
): Promise<string> {
  const digest = await globalThis.crypto.subtle.digest(
    "SHA-256",
    new TextEncoder().encode(canonicalRequest(request)),
  );
  return Array.from(new Uint8Array(digest), (byte) =>
    byte.toString(16).padStart(2, "0"),
  ).join("");
}

export function validateCheckoutCommittedResult(
  value: unknown,
  request: CheckoutMutationRequest,
  requestDigest: string,
): CheckoutCommittedResult {
  if (!value || typeof value !== "object") {
    throw new CheckoutMutationError("invalid_committed_result");
  }
  const result = value as Partial<CheckoutCommittedResult>;
  let materialized: CheckoutCommittedResult | undefined;
  let valid = false;
  try {
    const resultRequest = result.request;
    const requestSnapshot: CheckoutMutationRequest | undefined = resultRequest
      ? Object.freeze({
          chainId: resultRequest.chainId,
          customerAddress: resultRequest.customerAddress,
          offeringId: resultRequest.offeringId,
          providerAddress: resultRequest.providerAddress,
          durationSeconds: resultRequest.durationSeconds,
          priceAmount: resultRequest.priceAmount,
          priceDenom: resultRequest.priceDenom,
        })
      : undefined;
    materialized = requestSnapshot
      ? Object.freeze({
          status: result.status as "committed",
          code: result.code as 0,
          orderId: result.orderId as string,
          txHash: result.txHash as string,
          blockHeight: result.blockHeight as number,
          requestDigest: result.requestDigest as string,
          idempotencyKey: result.idempotencyKey as string,
          request: requestSnapshot,
        })
      : undefined;
    valid =
      materialized !== undefined &&
      materialized.status === "committed" &&
      materialized.code === 0 &&
      typeof materialized.orderId === "string" &&
      materialized.orderId.trim().length > 0 &&
      typeof materialized.txHash === "string" &&
      materialized.txHash.trim().length > 0 &&
      typeof materialized.blockHeight === "number" &&
      Number.isInteger(materialized.blockHeight) &&
      materialized.blockHeight > 0 &&
      materialized.requestDigest === requestDigest &&
      materialized.idempotencyKey === requestDigest &&
      canonicalRequest(materialized.request) === canonicalRequest(request);
  } catch {
    valid = false;
  }
  if (!valid) throw new CheckoutMutationError("invalid_committed_result");
  return materialized!;
}

export async function submitCheckoutMutation({
  adapter,
  projector,
  request,
  getCurrentRequest,
  signal,
  timeoutMs = 30_000,
}: {
  adapter: CheckoutMutationAdapter;
  projector: CheckoutMutationProjector;
  request: CheckoutMutationRequest;
  getCurrentRequest: () => CheckoutMutationRequest | null;
  signal?: AbortSignal;
  timeoutMs?: number;
}): Promise<CheckoutCommittedResult> {
  const submittedRequest: CheckoutMutationRequest = Object.freeze({
    ...request,
  });
  const requestDigest = await digestCheckoutMutationRequest(submittedRequest);
  if (signal?.aborted) throw new CheckoutMutationError("submission_cancelled");
  if (!checkoutMutationRequestsEqual(getCurrentRequest(), submittedRequest)) {
    throw new CheckoutMutationError("request_changed");
  }
  const adapterController = new AbortController();
  let timeout: ReturnType<typeof setTimeout> | undefined;
  let abortHandler: (() => void) | undefined;

  try {
    const cancellation = new Promise<never>((_, reject) => {
      if (!signal) return;
      abortHandler = () => {
        adapterController.abort();
        reject(new CheckoutMutationError("submission_cancelled"));
      };
      signal.addEventListener("abort", abortHandler, { once: true });
    });
    const rawResult = await Promise.race([
      adapter.submitOrder(submittedRequest, {
        requestDigest,
        idempotencyKey: requestDigest,
        signal: adapterController.signal,
      }),
      new Promise<never>((_, reject) => {
        timeout = setTimeout(() => {
          adapterController.abort();
          reject(new CheckoutMutationError("submission_timeout"));
        }, timeoutMs);
      }),
      cancellation,
    ]);
    if (signal?.aborted)
      throw new CheckoutMutationError("submission_cancelled");
    if (!checkoutMutationRequestsEqual(getCurrentRequest(), submittedRequest)) {
      throw new CheckoutMutationError("request_changed");
    }
    let projected: unknown;
    try {
      projected = projector(rawResult);
    } catch {
      throw new CheckoutMutationError("invalid_committed_result");
    }
    return validateCheckoutCommittedResult(
      projected,
      submittedRequest,
      requestDigest,
    );
  } catch (error) {
    if (error instanceof CheckoutMutationError) throw error;
    throw new CheckoutMutationError("submission_failed");
  } finally {
    if (timeout) clearTimeout(timeout);
    if (abortHandler) signal?.removeEventListener("abort", abortHandler);
  }
}
