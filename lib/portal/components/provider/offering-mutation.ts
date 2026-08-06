import type { OfferingDraft, ProviderOffering } from "../../types/provider";

export type ProviderOfferingMutationAction =
  | "create"
  | "update"
  | "publish"
  | "pause";

export interface ProviderOfferingMutationRequest {
  chainId: string;
  accountAddress: string;
  action: ProviderOfferingMutationAction;
  offeringId?: string;
  draft?: OfferingDraft | Partial<OfferingDraft>;
}

export interface ProviderOfferingMutationContext {
  requestDigest: string;
  idempotencyKey: string;
  signal: AbortSignal;
}

export interface ProviderOfferingMutationAdapter {
  readonly chainId: string;
  readonly accountAddress: string;
  mutateOffering(
    request: ProviderOfferingMutationRequest,
    context: ProviderOfferingMutationContext,
  ): Promise<unknown>;
}

export interface CommittedProviderOfferingMutation {
  status: "committed";
  code: 0;
  txHash: string;
  blockHeight: number;
  operationId: string;
  requestDigest: string;
  idempotencyKey: string;
  request: ProviderOfferingMutationRequest;
  offering: ProviderOffering;
}

export class ProviderOfferingMutationError extends Error {
  constructor(
    readonly code:
      | "feature_unavailable"
      | "invalid_request"
      | "invalid_committed_result"
      | "request_changed"
      | "submission_cancelled"
      | "submission_in_progress",
  ) {
    super(code);
    this.name = "ProviderOfferingMutationError";
  }
}

const canonical = (request: ProviderOfferingMutationRequest) =>
  JSON.stringify(request);
const text = (value: unknown): string | null =>
  typeof value === "string" && value.trim() ? value.trim() : null;

const cloneFrozen = (value: unknown): unknown => {
  if (Array.isArray(value)) return Object.freeze(value.map(cloneFrozen));
  if (value && typeof value === "object") {
    const source = value as Record<string, unknown>;
    return Object.freeze(
      Object.fromEntries(
        Object.entries(source).map(([key, item]) => [key, cloneFrozen(item)]),
      ),
    );
  }
  if (
    value === undefined ||
    value === null ||
    typeof value === "string" ||
    typeof value === "boolean" ||
    (typeof value === "number" && Number.isFinite(value))
  ) {
    return value;
  }
  throw new ProviderOfferingMutationError("invalid_request");
};

export function buildProviderOfferingMutationRequest(
  action: ProviderOfferingMutationAction,
  binding: { chainId: string; accountAddress: string },
  offeringId?: string,
  draft?: OfferingDraft | Partial<OfferingDraft>,
): ProviderOfferingMutationRequest {
  if (
    !["create", "update", "publish", "pause"].includes(action) ||
    !text(binding.chainId) ||
    !text(binding.accountAddress)
  ) {
    throw new ProviderOfferingMutationError("invalid_request");
  }
  if (action === "create" && (!draft || offeringId !== undefined)) {
    throw new ProviderOfferingMutationError("invalid_request");
  }
  if (
    (action === "update" && (!text(offeringId) || !draft)) ||
    ((action === "publish" || action === "pause") &&
      (!text(offeringId) || draft !== undefined))
  ) {
    throw new ProviderOfferingMutationError("invalid_request");
  }
  return Object.freeze({
    chainId: binding.chainId,
    accountAddress: binding.accountAddress,
    action,
    offeringId: offeringId?.trim(),
    draft: cloneFrozen(draft) as
      | OfferingDraft
      | Partial<OfferingDraft>
      | undefined,
  });
}

export async function digestProviderOfferingRequest(
  request: ProviderOfferingMutationRequest,
): Promise<string> {
  const bytes = new TextEncoder().encode(canonical(request));
  const digest = await globalThis.crypto.subtle.digest("SHA-256", bytes);
  return Array.from(new Uint8Array(digest), (byte) =>
    byte.toString(16).padStart(2, "0"),
  ).join("");
}

const materializeOffering = (value: unknown): ProviderOffering => {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new ProviderOfferingMutationError("invalid_committed_result");
  }
  const source = value as Partial<ProviderOffering>;
  if (
    !text(source.id) ||
    !text(source.title) ||
    !text(source.type) ||
    !text(source.status) ||
    !Number.isSafeInteger(source.activeOrders) ||
    (source.activeOrders ?? -1) < 0 ||
    !Number.isSafeInteger(source.totalOrders) ||
    (source.totalOrders ?? -1) < 0 ||
    typeof source.capacityUtilization !== "number" ||
    !Number.isFinite(source.capacityUtilization) ||
    source.capacityUtilization < 0 ||
    source.capacityUtilization > 100 ||
    !text(source.totalRevenue) ||
    !Number.isSafeInteger(source.createdAt) ||
    !Number.isSafeInteger(source.updatedAt) ||
    (source.createdAt ?? 0) <= 0 ||
    (source.updatedAt ?? 0) < (source.createdAt ?? 0)
  ) {
    throw new ProviderOfferingMutationError("invalid_committed_result");
  }
  return Object.freeze({
    id: source.id!.trim(),
    title: source.title!.trim(),
    type: source.type!.trim(),
    status: source.status!.trim(),
    activeOrders: source.activeOrders!,
    totalOrders: source.totalOrders!,
    capacityUtilization: source.capacityUtilization,
    totalRevenue: source.totalRevenue!.trim(),
    createdAt: source.createdAt!,
    updatedAt: source.updatedAt!,
  });
};

export function validateCommittedProviderOfferingMutation(
  value: unknown,
  request: ProviderOfferingMutationRequest,
  requestDigest: string,
): CommittedProviderOfferingMutation {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new ProviderOfferingMutationError("invalid_committed_result");
  }
  const source = value as Partial<CommittedProviderOfferingMutation>;
  const offering = materializeOffering(source.offering);
  if (
    source.status !== "committed" ||
    source.code !== 0 ||
    !text(source.txHash) ||
    !Number.isSafeInteger(source.blockHeight) ||
    (source.blockHeight ?? 0) <= 0 ||
    !text(source.operationId) ||
    source.requestDigest !== requestDigest ||
    source.idempotencyKey !== requestDigest ||
    !source.request ||
    canonical(source.request) !== canonical(request) ||
    (request.offeringId && offering.id !== request.offeringId) ||
    (request.action === "create" &&
      offering.status !==
        ((request.draft as OfferingDraft).autoPublish ? "active" : "draft")) ||
    (request.action === "publish" && offering.status !== "active") ||
    (request.action === "pause" && offering.status !== "paused")
  ) {
    throw new ProviderOfferingMutationError("invalid_committed_result");
  }
  return Object.freeze({
    status: "committed",
    code: 0,
    txHash: source.txHash!.trim(),
    blockHeight: source.blockHeight!,
    operationId: source.operationId!.trim(),
    requestDigest,
    idempotencyKey: requestDigest,
    request,
    offering,
  });
}

export function requireProviderOfferingMutationAdapter(
  adapter: ProviderOfferingMutationAdapter | undefined,
  binding: { chainId: string; accountAddress: string },
): ProviderOfferingMutationAdapter {
  if (
    !adapter ||
    adapter.chainId !== binding.chainId ||
    adapter.accountAddress !== binding.accountAddress
  ) {
    throw new ProviderOfferingMutationError("feature_unavailable");
  }
  return adapter;
}
