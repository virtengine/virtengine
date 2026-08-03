export const providerDeploymentActions = [
  "start",
  "stop",
  "restart",
  "update",
  "terminate",
] as const;

export type ProviderDeploymentAction =
  (typeof providerDeploymentActions)[number];

export type ProviderDeploymentActionStatus = "accepted" | "committed";

export interface ProviderDeploymentActionTxEvidence {
  hash: string;
  chainId: string;
  height: number;
}

export interface ProviderDeploymentActionReceipt {
  operationId: string;
  action: ProviderDeploymentAction;
  deploymentId: string;
  providerId: string;
  status: ProviderDeploymentActionStatus;
  issuedAt: Date;
  completedAt: Date;
  state: string;
  version: string;
  revision: string;
  txEvidence?: ProviderDeploymentActionTxEvidence;
}

export interface ProviderDeploymentActionCapability {
  receiptVersion: "v1";
  requiresChainSigning: boolean;
}

export type ProviderDeploymentActionErrorCode =
  | "feature_unavailable"
  | "action_rejected"
  | "malformed_receipt"
  | "receipt_mismatch"
  | "refresh_failed"
  | "deployment_drift"
  | "duplicate_action"
  | "chain_signing_required";

export class ProviderDeploymentActionError extends Error {
  readonly code: ProviderDeploymentActionErrorCode;
  readonly cause?: unknown;

  constructor(
    code: ProviderDeploymentActionErrorCode,
    message: string,
    cause?: unknown,
  ) {
    super(message);
    this.name = "ProviderDeploymentActionError";
    this.code = code;
    this.cause = cause;
  }
}

export type ProviderDeploymentTxEvidenceValidator = (
  evidence: ProviderDeploymentActionTxEvidence,
  receipt: Omit<ProviderDeploymentActionReceipt, "txEvidence">,
) => boolean | Promise<boolean>;

export interface ProviderDeploymentActionValidationContext {
  action: ProviderDeploymentAction;
  deploymentId: string;
  providerId: string;
  validateTxEvidence?: ProviderDeploymentTxEvidenceValidator;
}

export type ProviderDeploymentActionReceiptValidator = (
  value: unknown,
  context: ProviderDeploymentActionValidationContext,
) => Promise<ProviderDeploymentActionReceipt>;

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const requiredString = (
  value: Record<string, unknown>,
  camelKey: string,
  snakeKey: string,
): string | null => {
  const candidate = value[camelKey] ?? value[snakeKey];
  return typeof candidate === "string" && candidate.trim()
    ? candidate.trim()
    : null;
};

const requiredDate = (
  value: Record<string, unknown>,
  camelKey: string,
  snakeKey: string,
): Date | null => {
  const candidate = value[camelKey] ?? value[snakeKey];
  const date =
    candidate instanceof Date ? candidate : new Date(String(candidate));
  return Number.isNaN(date.getTime()) ? null : date;
};

const requiredBinding = (
  value: Record<string, unknown>,
  key: string,
): string | null => {
  const candidate = value[key];
  if (typeof candidate === "number" && Number.isFinite(candidate)) {
    return String(candidate);
  }
  return typeof candidate === "string" && candidate.trim()
    ? candidate.trim()
    : null;
};

const malformed = (message: string) =>
  new ProviderDeploymentActionError("malformed_receipt", message);

export const validateProviderDeploymentActionReceipt: ProviderDeploymentActionReceiptValidator =
  async (value, context) => {
    if (!isRecord(value)) {
      throw malformed("Provider action response is not a receipt");
    }
    if (
      "success" in value &&
      !("operationId" in value || "operation_id" in value)
    ) {
      throw new ProviderDeploymentActionError(
        "feature_unavailable",
        "Provider action endpoint returned a legacy response without receipt capability",
      );
    }

    const operationId = requiredString(value, "operationId", "operation_id");
    const action = requiredString(value, "action", "action");
    const deploymentId = requiredString(value, "deploymentId", "deployment_id");
    const providerId = requiredString(value, "providerId", "provider_id");
    const status = requiredString(value, "status", "status");
    const issuedAt = requiredDate(value, "issuedAt", "issued_at");
    const completedAt = requiredDate(value, "completedAt", "completed_at");
    const state = requiredString(value, "state", "state");
    const version = requiredBinding(value, "version");
    const revision = requiredBinding(value, "revision");

    if (
      !operationId ||
      !providerDeploymentActions.includes(action as ProviderDeploymentAction) ||
      !deploymentId ||
      !providerId ||
      (status !== "accepted" && status !== "committed") ||
      !issuedAt ||
      !completedAt ||
      !state ||
      !version ||
      !revision
    ) {
      throw malformed(
        "Provider action receipt is missing authoritative fields",
      );
    }
    if (
      action !== context.action ||
      deploymentId !== context.deploymentId ||
      providerId !== context.providerId
    ) {
      throw new ProviderDeploymentActionError(
        "receipt_mismatch",
        "Provider action receipt does not match the requested action",
      );
    }
    if (completedAt.getTime() < issuedAt.getTime()) {
      throw malformed("Provider action receipt completed before it was issued");
    }

    const receipt: ProviderDeploymentActionReceipt = {
      operationId,
      action: action as ProviderDeploymentAction,
      deploymentId,
      providerId,
      status,
      issuedAt,
      completedAt,
      state,
      version,
      revision,
    };

    const rawEvidence = value.txEvidence ?? value.tx_evidence;
    if (rawEvidence !== undefined) {
      if (!isRecord(rawEvidence)) {
        throw malformed("Provider action transaction evidence is malformed");
      }
      const hash = requiredString(rawEvidence, "hash", "hash");
      const chainId = requiredString(rawEvidence, "chainId", "chain_id");
      const height = rawEvidence.height;
      if (
        !hash ||
        !chainId ||
        typeof height !== "number" ||
        !Number.isSafeInteger(height)
      ) {
        throw malformed("Provider action transaction evidence is incomplete");
      }
      const evidence = { hash, chainId, height };
      if (
        !context.validateTxEvidence ||
        !(await context.validateTxEvidence(evidence, receipt))
      ) {
        throw malformed(
          "Provider action transaction evidence was not validated",
        );
      }
      receipt.txEvidence = evidence;
    }

    return receipt;
  };
