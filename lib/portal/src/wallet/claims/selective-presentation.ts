export interface RequestedClaimProjection {
  id: string;
  label: string;
}

export interface ConsentProjection {
  text: string;
  digest: string;
}

export interface CanonicalChallengeProjection {
  audience: string;
  orderOrCaseId: string;
  purpose: string;
  nonce: string;
  expiresAt: string;
  holder: string;
  issuerAllowlist: readonly string[];
  statusEpoch: string;
  policyDigest: string;
  policyVersion: string;
  requestedClaims: readonly RequestedClaimProjection[];
  consent: ConsentProjection;
}

export interface PresentationContext {
  audience: string;
  orderOrCaseId: string;
  purpose: string;
  nonce: string;
  holder: string;
  policyDigest: string;
  policyVersion: string;
}

export interface ChallengeValidator {
  validateAndProject(challenge: unknown): Promise<CanonicalChallengeProjection>;
}

export interface RequestedDecryptedClaim {
  id: string;
  value: unknown;
  issuer: string;
  statusEpoch: string;
}

export interface RequestedClaimReader {
  readRequested(
    claimIds: readonly string[],
  ): Promise<readonly RequestedDecryptedClaim[]>;
}

export interface ClaimStatusAuthority {
  checkStatus(claim: {
    id: string;
    issuer: string;
    statusEpoch: string;
  }): Promise<{ revoked: boolean; statusEpoch: string }>;
}

export interface PresentationBindingProjection
  extends Omit<CanonicalChallengeProjection, "requestedClaims"> {
  disclosedClaimIds: readonly string[];
}

export interface OpaquePresentationAuthority {
  createPresentation(input: {
    challenge: unknown;
    projection: CanonicalChallengeProjection;
    claims: readonly RequestedDecryptedClaim[];
  }): Promise<unknown>;
  validateAndProject(
    presentation: unknown,
    challenge: unknown,
  ): Promise<PresentationBindingProjection>;
}

export interface NonceReplayGuard {
  consume(nonce: string, expiresAt: string): Promise<boolean>;
}

export interface SelectivePresentationDependencies {
  challengeValidator?: ChallengeValidator;
  claimReader?: RequestedClaimReader;
  presenter?: OpaquePresentationAuthority;
  statusAuthority?: ClaimStatusAuthority;
  replayGuard?: NonceReplayGuard;
  now?: () => Date;
}

export interface SelectivePresentationReview {
  audience: string;
  orderOrCaseId: string;
  purpose: string;
  expiresAt: string;
  holder: string;
  requestedClaims: readonly RequestedClaimProjection[];
  consent: ConsentProjection;
}

export type SelectivePresentationErrorCode =
  | "unavailable"
  | "invalid_challenge"
  | "expired"
  | "binding_mismatch"
  | "unknown_claim"
  | "issuer_mismatch"
  | "stale_status"
  | "revoked"
  | "replayed_nonce"
  | "extra_disclosure"
  | "presentation_mismatch";

export class SelectivePresentationError extends Error {
  constructor(
    readonly code: SelectivePresentationErrorCode,
    message: string,
  ) {
    super(message);
    this.name = "SelectivePresentationError";
  }
}

function requireString(value: unknown, path: string): asserts value is string {
  if (typeof value !== "string" || value.length === 0) {
    throw new SelectivePresentationError(
      "invalid_challenge",
      `${path} must be a non-empty string`,
    );
  }
}

function assertUniqueStrings(
  values: readonly string[],
  path: string,
  allowEmpty = false,
): void {
  if (!allowEmpty && values.length === 0) {
    throw new SelectivePresentationError(
      "invalid_challenge",
      `${path} must not be empty`,
    );
  }
  const seen = new Set<string>();
  for (const [index, value] of values.entries()) {
    requireString(value, `${path}[${index}]`);
    if (seen.has(value)) {
      throw new SelectivePresentationError(
        "invalid_challenge",
        `${path} contains a duplicate`,
      );
    }
    seen.add(value);
  }
}

function sameSet(left: readonly string[], right: readonly string[]): boolean {
  return (
    left.length === right.length &&
    new Set(left).size === left.length &&
    new Set(right).size === right.length &&
    left.every((value) => right.includes(value))
  );
}

function assertEqual(
  actual: string,
  expected: string,
  path: string,
  code: SelectivePresentationErrorCode = "binding_mismatch",
): void {
  if (actual !== expected) {
    throw new SelectivePresentationError(code, `${path} does not match`);
  }
}

export class SelectivePresentationAdapter {
  constructor(
    private readonly dependencies: SelectivePresentationDependencies = {},
  ) {}

  isAvailable(): boolean {
    return Boolean(
      this.dependencies.challengeValidator &&
      this.dependencies.claimReader &&
      this.dependencies.presenter &&
      this.dependencies.statusAuthority &&
      this.dependencies.replayGuard,
    );
  }

  async review(
    challenge: unknown,
    context: PresentationContext,
  ): Promise<SelectivePresentationReview> {
    const projection = await this.projectChallenge(challenge, context);
    return {
      audience: projection.audience,
      orderOrCaseId: projection.orderOrCaseId,
      purpose: projection.purpose,
      expiresAt: projection.expiresAt,
      holder: projection.holder,
      requestedClaims: projection.requestedClaims.map(({ id, label }) => ({
        id,
        label,
      })),
      consent: { ...projection.consent },
    };
  }

  async present(challenge: unknown, context: PresentationContext): Promise<unknown> {
    const projection = await this.projectChallenge(challenge, context);
    if (
      !(await this.dependencies.replayGuard!.consume(
        projection.nonce,
        projection.expiresAt,
      ))
    ) {
      throw new SelectivePresentationError(
        "replayed_nonce",
        "Challenge nonce was already consumed",
      );
    }

    const requestedIds = projection.requestedClaims.map(({ id }) => id);
    const claims = await this.dependencies.claimReader!.readRequested(
      requestedIds,
    );
    this.assertRequestedClaims(claims, requestedIds, projection);
    await this.assertStatuses(claims, projection);

    const presentation = await this.dependencies.presenter!.createPresentation({
      challenge,
      projection,
      claims,
    });
    const binding = await this.dependencies.presenter!.validateAndProject(
      presentation,
      challenge,
    );
    this.assertPresentationBinding(binding, projection);
    return presentation;
  }

  private requireDependencies(): void {
    if (!this.isAvailable()) {
      throw new SelectivePresentationError(
        "unavailable",
        "Selective presentation authorities are unavailable",
      );
    }
  }

  private async projectChallenge(
    challenge: unknown,
    context: PresentationContext,
  ): Promise<CanonicalChallengeProjection> {
    this.requireDependencies();
    const projection = await this.dependencies.challengeValidator!.validateAndProject(
      challenge,
    );
    for (const [path, value] of Object.entries({
      audience: projection.audience,
      orderOrCaseId: projection.orderOrCaseId,
      purpose: projection.purpose,
      nonce: projection.nonce,
      expiresAt: projection.expiresAt,
      holder: projection.holder,
      statusEpoch: projection.statusEpoch,
      policyDigest: projection.policyDigest,
      policyVersion: projection.policyVersion,
      consentText: projection.consent?.text,
      consentDigest: projection.consent?.digest,
    })) {
      requireString(value, path);
    }
    if (!Array.isArray(projection.issuerAllowlist)) {
      throw new SelectivePresentationError(
        "invalid_challenge",
        "issuerAllowlist must be an array",
      );
    }
    assertUniqueStrings(projection.issuerAllowlist, "issuerAllowlist");
    if (!Array.isArray(projection.requestedClaims)) {
      throw new SelectivePresentationError(
        "invalid_challenge",
        "requestedClaims must be an array",
      );
    }
    for (const [index, claim] of projection.requestedClaims.entries()) {
      requireString(claim?.id, `requestedClaims[${index}].id`);
      requireString(claim?.label, `requestedClaims[${index}].label`);
    }
    assertUniqueStrings(
      projection.requestedClaims.map(({ id }) => id),
      "requestedClaims",
    );
    const expiry = Date.parse(projection.expiresAt);
    if (!Number.isFinite(expiry)) {
      throw new SelectivePresentationError(
        "invalid_challenge",
        "expiresAt must be a valid timestamp",
      );
    }
    if (expiry <= (this.dependencies.now ?? (() => new Date()))().getTime()) {
      throw new SelectivePresentationError("expired", "Challenge has expired");
    }

    for (const key of [
      "audience",
      "orderOrCaseId",
      "purpose",
      "nonce",
      "holder",
      "policyDigest",
      "policyVersion",
    ] as const) {
      requireString(context[key], `context.${key}`);
      assertEqual(projection[key], context[key], key);
    }
    return projection;
  }

  private assertRequestedClaims(
    claims: readonly RequestedDecryptedClaim[],
    requestedIds: readonly string[],
    projection: CanonicalChallengeProjection,
  ): void {
    const returnedIds = claims.map(({ id }) => id);
    if (!sameSet(returnedIds, requestedIds)) {
      const hasExtra = returnedIds.some((id) => !requestedIds.includes(id));
      throw new SelectivePresentationError(
        hasExtra ? "extra_disclosure" : "unknown_claim",
        hasExtra
          ? "Claim reader returned an unrequested claim"
          : "A requested claim is unavailable",
      );
    }
    for (const claim of claims) {
      requireString(claim.id, "claim.id");
      requireString(claim.issuer, `claim.${claim.id}.issuer`);
      requireString(claim.statusEpoch, `claim.${claim.id}.statusEpoch`);
      if (claim.value === undefined) {
        throw new SelectivePresentationError(
          "unknown_claim",
          `Claim ${claim.id} is unavailable`,
        );
      }
      if (!projection.issuerAllowlist.includes(claim.issuer)) {
        throw new SelectivePresentationError(
          "issuer_mismatch",
          `Claim ${claim.id} has an unapproved issuer`,
        );
      }
      assertEqual(
        claim.statusEpoch,
        projection.statusEpoch,
        `claim.${claim.id}.statusEpoch`,
        "stale_status",
      );
    }
  }

  private async assertStatuses(
    claims: readonly RequestedDecryptedClaim[],
    projection: CanonicalChallengeProjection,
  ): Promise<void> {
    for (const claim of claims) {
      const status = await this.dependencies.statusAuthority!.checkStatus({
        id: claim.id,
        issuer: claim.issuer,
        statusEpoch: claim.statusEpoch,
      });
      if (status.revoked) {
        throw new SelectivePresentationError(
          "revoked",
          `Claim ${claim.id} is revoked`,
        );
      }
      assertEqual(
        status.statusEpoch,
        projection.statusEpoch,
        `status.${claim.id}.statusEpoch`,
        "stale_status",
      );
    }
  }

  private assertPresentationBinding(
    binding: PresentationBindingProjection,
    projection: CanonicalChallengeProjection,
  ): void {
    for (const key of [
      "audience",
      "orderOrCaseId",
      "purpose",
      "nonce",
      "expiresAt",
      "holder",
      "statusEpoch",
      "policyDigest",
      "policyVersion",
    ] as const) {
      assertEqual(binding[key], projection[key], `presentation.${key}`, "presentation_mismatch");
    }
    if (
      !Array.isArray(binding.issuerAllowlist) ||
      !Array.isArray(binding.disclosedClaimIds) ||
      !binding.consent ||
      typeof binding.consent.text !== "string" ||
      typeof binding.consent.digest !== "string"
    ) {
      throw new SelectivePresentationError(
        "presentation_mismatch",
        "Presentation binding projection is malformed",
      );
    }
    if (!sameSet(binding.issuerAllowlist, projection.issuerAllowlist)) {
      throw new SelectivePresentationError(
        "presentation_mismatch",
        "Presentation issuer allowlist does not match",
      );
    }
    assertEqual(
      binding.consent.text,
      projection.consent.text,
      "presentation.consent.text",
      "presentation_mismatch",
    );
    assertEqual(
      binding.consent.digest,
      projection.consent.digest,
      "presentation.consent.digest",
      "presentation_mismatch",
    );
    const requestedIds = projection.requestedClaims.map(({ id }) => id);
    if (!sameSet(binding.disclosedClaimIds, requestedIds)) {
      const hasExtra = binding.disclosedClaimIds.some(
        (id) => !requestedIds.includes(id),
      );
      throw new SelectivePresentationError(
        hasExtra ? "extra_disclosure" : "presentation_mismatch",
        hasExtra
          ? "Presentation disclosed an unrequested claim"
          : "Presentation omitted a requested claim",
      );
    }
  }
}