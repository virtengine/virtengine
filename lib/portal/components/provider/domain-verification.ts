import type {
  DomainChallenge,
  DomainVerification,
  DomainVerificationMethod,
} from "../../types/provider";

export interface ProviderDomainBinding {
  chainId: string;
  accountAddress: string;
}

export interface ProviderDomainChallenge
  extends DomainChallenge, ProviderDomainBinding {
  challengeId: string;
}

export interface ProviderDomainVerificationEvidence
  extends DomainVerification, ProviderDomainBinding {
  status: "verified";
  challengeId: string;
  evidenceId: string;
}

export interface ProviderDomainVerifier {
  readonly chainId: string;
  readonly accountAddress: string;
  issueChallenge(
    domain: string,
    method: DomainVerificationMethod,
  ): Promise<unknown>;
  verifyChallenge(challenge: ProviderDomainChallenge): Promise<unknown>;
}

export class ProviderDomainVerificationError extends Error {
  constructor(
    readonly code:
      | "feature_unavailable"
      | "invalid_domain"
      | "invalid_challenge"
      | "invalid_verification"
      | "authority_changed"
      | "challenge_in_progress"
      | "verification_in_progress",
  ) {
    super(code);
    this.name = "ProviderDomainVerificationError";
  }
}

export function normalizeProviderDomain(value: string): string {
  const domain = value.trim().toLowerCase().replace(/\.$/, "");
  if (
    domain.length > 253 ||
    !domain.includes(".") ||
    !domain
      .split(".")
      .every(
        (label) =>
          label.length > 0 &&
          label.length <= 63 &&
          /^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$/.test(label),
      )
  ) {
    throw new ProviderDomainVerificationError("invalid_domain");
  }
  return domain;
}

export function requireProviderDomainVerifier(
  verifier: ProviderDomainVerifier | undefined,
  binding: ProviderDomainBinding,
): ProviderDomainVerifier {
  if (
    !verifier ||
    typeof verifier.issueChallenge !== "function" ||
    typeof verifier.verifyChallenge !== "function" ||
    verifier.chainId !== binding.chainId ||
    verifier.accountAddress !== binding.accountAddress
  ) {
    throw new ProviderDomainVerificationError("feature_unavailable");
  }
  return verifier;
}

const record = (value: unknown): Record<string, unknown> => {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new ProviderDomainVerificationError("invalid_challenge");
  }
  return value as Record<string, unknown>;
};

const text = (
  value: unknown,
  code: "invalid_challenge" | "invalid_verification",
): string => {
  if (typeof value !== "string" || !value.trim()) {
    throw new ProviderDomainVerificationError(code);
  }
  return value;
};

const timestamp = (
  value: unknown,
  code: "invalid_challenge" | "invalid_verification",
): number => {
  if (typeof value !== "number" || !Number.isSafeInteger(value) || value <= 0) {
    throw new ProviderDomainVerificationError(code);
  }
  return value;
};

export function validateProviderDomainChallenge(
  value: unknown,
  binding: ProviderDomainBinding,
  domain: string,
  method: DomainVerificationMethod,
  now = Date.now(),
): ProviderDomainChallenge {
  if (method !== "dns_txt" && method !== "http_file") {
    throw new ProviderDomainVerificationError("invalid_challenge");
  }
  const source = record(value);
  const normalizedDomain = normalizeProviderDomain(
    text(source.domain, "invalid_challenge"),
  );
  const expiresAt = timestamp(source.expiresAt, "invalid_challenge");
  if (
    source.chainId !== binding.chainId ||
    source.accountAddress !== binding.accountAddress ||
    normalizedDomain !== domain ||
    source.method !== method ||
    expiresAt <= now
  ) {
    throw new ProviderDomainVerificationError("invalid_challenge");
  }
  const result: ProviderDomainChallenge = {
    chainId: binding.chainId,
    accountAddress: binding.accountAddress,
    challengeId: text(source.challengeId, "invalid_challenge"),
    domain,
    method,
    challengeValue: text(source.challengeValue, "invalid_challenge"),
    expiresAt,
    instructions: text(source.instructions, "invalid_challenge"),
    dnsRecordName:
      source.dnsRecordName === undefined
        ? undefined
        : text(source.dnsRecordName, "invalid_challenge"),
    httpFilePath:
      source.httpFilePath === undefined
        ? undefined
        : text(source.httpFilePath, "invalid_challenge"),
  };
  if (
    (method === "dns_txt" && !result.dnsRecordName) ||
    (method === "http_file" && !result.httpFilePath)
  ) {
    throw new ProviderDomainVerificationError("invalid_challenge");
  }
  return Object.freeze(result);
}

export function validateProviderDomainVerification(
  value: unknown,
  binding: ProviderDomainBinding,
  challenge: ProviderDomainChallenge,
  now = Date.now(),
): ProviderDomainVerificationEvidence {
  const source = record(value);
  const verifiedAt = timestamp(source.verifiedAt, "invalid_verification");
  const expiresAt = timestamp(source.expiresAt, "invalid_verification");
  if (
    source.chainId !== binding.chainId ||
    source.accountAddress !== binding.accountAddress ||
    normalizeProviderDomain(text(source.domain, "invalid_verification")) !==
      challenge.domain ||
    source.method !== challenge.method ||
    source.challengeId !== challenge.challengeId ||
    source.status !== "verified" ||
    challenge.expiresAt <= now ||
    verifiedAt > now ||
    expiresAt <= now
  ) {
    throw new ProviderDomainVerificationError("invalid_verification");
  }
  return Object.freeze({
    chainId: binding.chainId,
    accountAddress: binding.accountAddress,
    challengeId: challenge.challengeId,
    evidenceId: text(source.evidenceId, "invalid_verification"),
    domain: challenge.domain,
    method: challenge.method,
    status: "verified" as const,
    verifiedAt,
    expiresAt,
  });
}
