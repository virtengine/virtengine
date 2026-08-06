export type ProviderShellSessionErrorCode =
  | "feature_unavailable"
  | "eligibility_unavailable"
  | "eligibility_expired"
  | "receipt_invalid"
  | "receipt_mismatch"
  | "receipt_expired"
  | "unsafe_transport";

export class ProviderShellSessionError extends Error {
  constructor(
    public readonly code: ProviderShellSessionErrorCode,
    message: string,
    public readonly cause?: unknown,
  ) {
    super(message);
    this.name = "ProviderShellSessionError";
  }
}

export interface ProviderShellSessionCapability {
  receiptVersion: "v1";
  transport: "one_time_reference" | "server_url";
  maxTtlSeconds: number;
}

export interface ShellEligibilityProjection {
  authorized: true;
  chainId: string;
  account: string;
  deploymentId: string;
  container: string;
  providerId: string;
  sessionId: string;
  policyEpoch: string;
  statusEpoch: string;
  policyExpiresAt: string;
  statusExpiresAt: string;
  capabilityDigest: string;
  policyDigest: string;
}

export interface ProviderShellSessionReceipt {
  version: "v1";
  deploymentId: string;
  container: string;
  account: string;
  providerId: string;
  chainId: string;
  eligibilitySessionId: string;
  oneTimeReference: string;
  issuedAt: string;
  expiresAt: string;
  capabilityDigest: string;
  policyDigest: string;
  websocketUrl?: string;
}

export interface ShellSessionValidationContext {
  eligibility?: ShellEligibilityProjection;
  capability?: ProviderShellSessionCapability;
  providerEndpoint: string;
  now?: Date;
}

const readString = (value: unknown, ...keys: string[]): string | undefined => {
  if (!value || typeof value !== "object") return undefined;
  const record = value as Record<string, unknown>;
  for (const key of keys) {
    const candidate = record[key];
    if (typeof candidate === "string" && candidate.trim()) {
      return candidate.trim();
    }
  }
  return undefined;
};

const requireDate = (
  value: string | undefined,
  code: ProviderShellSessionErrorCode,
  field: string,
): Date => {
  const timestamp = value ? Date.parse(value) : Number.NaN;
  if (Number.isNaN(timestamp)) {
    throw new ProviderShellSessionError(
      code,
      `Shell session ${field} is invalid`,
    );
  }
  return new Date(timestamp);
};

const validateEligibility = (
  eligibility: ShellEligibilityProjection | undefined,
  now: Date,
): ShellEligibilityProjection => {
  if (!eligibility?.authorized) {
    throw new ProviderShellSessionError(
      "eligibility_unavailable",
      "Authoritative shell eligibility is unavailable",
    );
  }

  const required = [
    eligibility.chainId,
    eligibility.account,
    eligibility.deploymentId,
    eligibility.container,
    eligibility.providerId,
    eligibility.sessionId,
    eligibility.policyEpoch,
    eligibility.statusEpoch,
    eligibility.capabilityDigest,
    eligibility.policyDigest,
  ];
  if (required.some((value) => !value.trim())) {
    throw new ProviderShellSessionError(
      "eligibility_unavailable",
      "Shell eligibility is missing an authoritative binding",
    );
  }

  const policyExpiry = requireDate(
    eligibility.policyExpiresAt,
    "eligibility_unavailable",
    "policy expiry",
  );
  const statusExpiry = requireDate(
    eligibility.statusExpiresAt,
    "eligibility_unavailable",
    "status expiry",
  );
  if (policyExpiry <= now || statusExpiry <= now) {
    throw new ProviderShellSessionError(
      "eligibility_expired",
      "Shell eligibility policy or status projection has expired",
    );
  }
  return eligibility;
};

const validateWebSocketUrl = (
  rawUrl: string,
  providerEndpoint: string,
): string => {
  let url: URL;
  let providerUrl: URL;
  try {
    url = new URL(rawUrl);
    providerUrl = new URL(providerEndpoint);
  } catch (cause) {
    throw new ProviderShellSessionError(
      "unsafe_transport",
      "Shell WebSocket URL is invalid",
      cause,
    );
  }

  const expectedProtocol = providerUrl.protocol === "https:" ? "wss:" : "ws:";
  if (
    url.protocol !== expectedProtocol ||
    url.host !== providerUrl.host ||
    url.username ||
    url.password
  ) {
    throw new ProviderShellSessionError(
      "unsafe_transport",
      "Shell WebSocket URL does not match the authoritative provider endpoint",
    );
  }
  for (const key of url.searchParams.keys()) {
    if (key !== "session_id") {
      throw new ProviderShellSessionError(
        "unsafe_transport",
        "Shell WebSocket URL contains an unsupported query credential",
      );
    }
  }
  return url.toString();
};

export const validateProviderShellSessionReceipt = (
  response: unknown,
  context: ShellSessionValidationContext,
): ProviderShellSessionReceipt => {
  const now = context.now ?? new Date();
  const eligibility = validateEligibility(context.eligibility, now);
  const capability = context.capability;
  if (
    !capability ||
    capability.receiptVersion !== "v1" ||
    capability.maxTtlSeconds <= 0
  ) {
    throw new ProviderShellSessionError(
      "feature_unavailable",
      "Provider does not declare authoritative shell sessions",
    );
  }

  if (readString(response, "token", "access_token", "bearer")) {
    throw new ProviderShellSessionError(
      "unsafe_transport",
      "Provider returned a reusable shell credential",
    );
  }

  const receipt: ProviderShellSessionReceipt = {
    version: readString(response, "version") as "v1",
    deploymentId:
      readString(response, "deploymentId", "deployment_id", "deployment") ?? "",
    container: readString(response, "container") ?? "",
    account: readString(response, "account") ?? "",
    providerId: readString(response, "providerId", "provider_id") ?? "",
    chainId: readString(response, "chainId", "chain_id") ?? "",
    eligibilitySessionId:
      readString(response, "eligibilitySessionId", "eligibility_session_id") ??
      "",
    oneTimeReference:
      readString(
        response,
        "oneTimeReference",
        "one_time_reference",
        "session_id",
      ) ?? "",
    issuedAt: readString(response, "issuedAt", "issued_at") ?? "",
    expiresAt: readString(response, "expiresAt", "expires_at") ?? "",
    capabilityDigest:
      readString(response, "capabilityDigest", "capability_digest") ?? "",
    policyDigest: readString(response, "policyDigest", "policy_digest") ?? "",
    websocketUrl: readString(response, "websocketUrl", "websocket_url"),
  };

  if (receipt.version !== "v1" || !receipt.oneTimeReference) {
    throw new ProviderShellSessionError(
      "feature_unavailable",
      "Provider returned a legacy shell session response",
    );
  }
  if (/[\s/?#=&]/.test(receipt.oneTimeReference)) {
    throw new ProviderShellSessionError(
      "receipt_invalid",
      "Shell session reference is not opaque",
    );
  }

  const bindings: Array<[string, string]> = [
    [receipt.deploymentId, eligibility.deploymentId],
    [receipt.container, eligibility.container],
    [receipt.account, eligibility.account],
    [receipt.providerId, eligibility.providerId],
    [receipt.chainId, eligibility.chainId],
    [receipt.eligibilitySessionId, eligibility.sessionId],
    [receipt.capabilityDigest, eligibility.capabilityDigest],
    [receipt.policyDigest, eligibility.policyDigest],
  ];
  if (bindings.some(([actual, expected]) => !actual || actual !== expected)) {
    throw new ProviderShellSessionError(
      "receipt_mismatch",
      "Provider shell receipt does not match authoritative eligibility",
    );
  }

  const issuedAt = requireDate(
    receipt.issuedAt,
    "receipt_invalid",
    "issue time",
  );
  const expiresAt = requireDate(receipt.expiresAt, "receipt_invalid", "expiry");
  if (issuedAt > now || expiresAt <= now) {
    throw new ProviderShellSessionError(
      "receipt_expired",
      "Provider shell receipt is not current",
    );
  }
  if (
    expiresAt.getTime() - issuedAt.getTime() >
    capability.maxTtlSeconds * 1000
  ) {
    throw new ProviderShellSessionError(
      "receipt_invalid",
      "Provider shell receipt exceeds the declared short lifetime",
    );
  }

  if (receipt.websocketUrl) {
    receipt.websocketUrl = validateWebSocketUrl(
      receipt.websocketUrl,
      context.providerEndpoint,
    );
    const url = new URL(receipt.websocketUrl);
    if (url.searchParams.get("session_id") !== receipt.oneTimeReference) {
      throw new ProviderShellSessionError(
        "receipt_mismatch",
        "Shell WebSocket URL does not match the one-time session reference",
      );
    }
  } else if (capability.transport === "server_url") {
    throw new ProviderShellSessionError(
      "feature_unavailable",
      "Provider requires a server-issued shell URL",
    );
  }
  return receipt;
};

export const buildProviderShellWebSocketUrl = (
  providerEndpoint: string,
  deploymentId: string,
  receipt: ProviderShellSessionReceipt,
): string => {
  if (receipt.websocketUrl) return receipt.websocketUrl;
  const endpoint = new URL(providerEndpoint);
  endpoint.protocol = endpoint.protocol === "https:" ? "wss:" : "ws:";
  endpoint.pathname = `/api/v1/deployments/${encodeURIComponent(deploymentId)}/shell`;
  endpoint.search = "";
  endpoint.searchParams.set("session_id", receipt.oneTimeReference);
  return endpoint.toString();
};
