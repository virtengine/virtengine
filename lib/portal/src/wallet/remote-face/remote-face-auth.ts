export interface RemoteFaceProfileReference {
  profile: string;
  version: string;
  digest: string;
}

export interface RemoteFaceProfileBinding {
  t1: RemoteFaceProfileReference;
  t5: RemoteFaceProfileReference;
}

export interface RemoteFaceAuthorizationBinding {
  chainId: string;
  accountId: string;
  sessionId: string;
  deviceId: string;
  actionKind: "action" | "transaction";
  actionDigest: string;
}

export interface RemoteFaceProfileProjection {
  enabled: boolean;
  profile: RemoteFaceProfileBinding;
}

export interface RemoteFaceEvidenceProjection {
  valid: boolean;
  evidenceRef: string;
  profile: RemoteFaceProfileBinding;
  binding: RemoteFaceAuthorizationBinding;
  validatedAt: number;
  expiresAt: number;
}

export interface RemoteFaceLivenessProjection extends RemoteFaceEvidenceProjection {
  observedAt: number;
}

export interface RemoteFacePossessionProjection extends RemoteFaceEvidenceProjection {
  proofType: "passkey-assertion" | "other-possession-factor";
}

export interface RemoteFaceRecoveryFactorProjection {
  valid: boolean;
  factorClass: "recovery";
  factorRef: string;
  accountId: string;
  validatedAt: number;
  expiresAt: number;
}

export interface RemoteFaceRecoveryPolicyProjection {
  valid: boolean;
  accountId: string;
  threshold: number;
  factorRefs: readonly string[];
  delayMs: number;
  expiresAt: number;
}

export interface RemoteFaceFinalAuthorizationProjection {
  authorized: boolean;
  authorizationRef: string;
  nonce: string;
  profile: RemoteFaceProfileBinding;
  binding: RemoteFaceAuthorizationBinding;
  expiresAt: number;
}

export interface RemoteFaceAuthenticationRequest {
  profile: RemoteFaceProfileBinding;
  binding: RemoteFaceAuthorizationBinding;
  nonce: string;
  issuedAt: number;
  expiresAt: number;
  activeLiveness: unknown;
  passiveLiveness: unknown;
  deviceSessionAttestation: unknown;
  possessionProof: unknown;
  recoveryFactors: readonly [unknown, unknown];
  recoveryPolicy: unknown;
  remoteFaceResult: unknown;
}

export interface RemoteFaceValidatedEvidence {
  activeLiveness: RemoteFaceLivenessProjection;
  passiveLiveness: RemoteFaceLivenessProjection;
  deviceSessionAttestation: RemoteFaceEvidenceProjection;
  possession: RemoteFacePossessionProjection;
  recoveryFactors: readonly [
    RemoteFaceRecoveryFactorProjection,
    RemoteFaceRecoveryFactorProjection,
  ];
  recoveryPolicy: RemoteFaceRecoveryPolicyProjection;
}

export interface RemoteFaceAuthenticationAuthorities {
  profileAuthority?: {
    validateAndProject(profile: unknown): Promise<RemoteFaceProfileProjection>;
  };
  activeLivenessValidator?: {
    validateAndProject(
      evidence: unknown,
    ): Promise<RemoteFaceLivenessProjection>;
  };
  passiveLivenessValidator?: {
    validateAndProject(
      evidence: unknown,
    ): Promise<RemoteFaceLivenessProjection>;
  };
  deviceSessionAttestationValidator?: {
    validateAndProject(
      evidence: unknown,
    ): Promise<RemoteFaceEvidenceProjection>;
  };
  possessionValidator?: {
    validateAndProject(
      evidence: unknown,
    ): Promise<RemoteFacePossessionProjection>;
  };
  recoveryFactorValidator?: {
    validateAndProject(
      evidence: unknown,
    ): Promise<RemoteFaceRecoveryFactorProjection>;
  };
  recoveryPolicyValidator?: {
    validateAndProject(
      policy: unknown,
    ): Promise<RemoteFaceRecoveryPolicyProjection>;
  };
  nonceGuard?: {
    consume(nonce: string, expiresAt: number): Promise<boolean>;
  };
  finalAuthorizationAuthority?: {
    authorizeAndProject(input: {
      profile: RemoteFaceProfileBinding;
      binding: RemoteFaceAuthorizationBinding;
      nonce: string;
      expiresAt: number;
      evidence: RemoteFaceValidatedEvidence;
      remoteFaceResult: unknown;
    }): Promise<RemoteFaceFinalAuthorizationProjection>;
  };
}

export interface RemoteFaceAuthenticationOptions {
  now?: () => number;
  maxRequestLifetimeMs?: number;
  maxLivenessAgeMs?: number;
  minimumRecoveryDelayMs?: number;
}

export interface RemoteFaceAuthenticationResult {
  authorized: true;
  authorizationRef: string;
  expiresAt: number;
}

export type RemoteFaceAuthenticationErrorCode =
  | "unavailable"
  | "disabled_profile"
  | "profile_mismatch"
  | "invalid_request"
  | "expired"
  | "prohibited_biometric_data"
  | "invalid_evidence"
  | "binding_mismatch"
  | "stale_liveness"
  | "missing_possession"
  | "missing_recovery"
  | "replayed_nonce"
  | "authorization_denied"
  | "authorization_mismatch";

export class RemoteFaceAuthenticationError extends Error {
  constructor(
    readonly code: RemoteFaceAuthenticationErrorCode,
    message: string,
  ) {
    super(message);
    this.name = "RemoteFaceAuthenticationError";
  }
}

const DEFAULT_MAX_REQUEST_LIFETIME_MS = 2 * 60 * 1000;
const DEFAULT_MAX_LIVENESS_AGE_MS = 30 * 1000;
const DEFAULT_MINIMUM_RECOVERY_DELAY_MS = 24 * 60 * 60 * 1000;

const prohibitedKeys = [
  "image",
  "images",
  "photo",
  "selfie",
  "video",
  "frame",
  "template",
  "embedding",
  "vector",
  "landmark",
  "metric",
  "metrics",
  "score",
  "confidence",
  "similarity",
  "distance",
  "biometricid",
  "biometricidentifier",
  "faceid",
  "faceidentifier",
];

function fail(code: RemoteFaceAuthenticationErrorCode, message: string): never {
  throw new RemoteFaceAuthenticationError(code, message);
}

function requireString(value: unknown, path: string): asserts value is string {
  if (typeof value !== "string" || value.trim().length === 0) {
    fail("invalid_request", `${path} must be a non-empty string`);
  }
}

function assertNoBiometricData(
  value: unknown,
  path: string,
  seen = new Set<object>(),
): void {
  if (value === null || typeof value !== "object") return;
  if (seen.has(value)) fail("invalid_request", `${path} must not be cyclic`);

  seen.add(value);
  try {
    for (const [key, child] of Object.entries(value)) {
      const normalized = key.toLowerCase().replace(/[^a-z0-9]/g, "");
      if (
        prohibitedKeys.some(
          (prohibited) =>
            normalized === prohibited || normalized.endsWith(prohibited),
        )
      ) {
        fail(
          "prohibited_biometric_data",
          `${path}.${key} is prohibited biometric data`,
        );
      }
      assertNoBiometricData(child, `${path}.${key}`, seen);
    }
  } finally {
    seen.delete(value);
  }
}

function profileMatches(
  actual: RemoteFaceProfileBinding,
  expected: RemoteFaceProfileBinding,
): boolean {
  return (
    actual?.t1?.profile === expected.t1.profile &&
    actual?.t1?.version === expected.t1.version &&
    actual?.t1?.digest === expected.t1.digest &&
    actual?.t5?.profile === expected.t5.profile &&
    actual?.t5?.version === expected.t5.version &&
    actual?.t5?.digest === expected.t5.digest
  );
}

function bindingMatches(
  actual: RemoteFaceAuthorizationBinding,
  expected: RemoteFaceAuthorizationBinding,
): boolean {
  return (
    actual?.chainId === expected.chainId &&
    actual?.accountId === expected.accountId &&
    actual?.sessionId === expected.sessionId &&
    actual?.deviceId === expected.deviceId &&
    actual?.actionKind === expected.actionKind &&
    actual?.actionDigest === expected.actionDigest
  );
}

function validateProfileShape(profile: RemoteFaceProfileBinding): void {
  for (const source of ["t1", "t5"] as const) {
    requireString(profile?.[source]?.profile, `profile.${source}.profile`);
    requireString(profile?.[source]?.version, `profile.${source}.version`);
    requireString(profile?.[source]?.digest, `profile.${source}.digest`);
  }
}

function validateBindingShape(binding: RemoteFaceAuthorizationBinding): void {
  requireString(binding?.chainId, "binding.chainId");
  requireString(binding?.accountId, "binding.accountId");
  requireString(binding?.sessionId, "binding.sessionId");
  requireString(binding?.deviceId, "binding.deviceId");
  requireString(binding?.actionDigest, "binding.actionDigest");
  if (
    binding?.actionKind !== "action" &&
    binding?.actionKind !== "transaction"
  ) {
    fail("invalid_request", "binding.actionKind is invalid");
  }
}

export class RemoteFaceAuthenticationGate {
  private readonly now: () => number;
  private readonly maxRequestLifetimeMs: number;
  private readonly maxLivenessAgeMs: number;
  private readonly minimumRecoveryDelayMs: number;

  constructor(
    private readonly authorities: RemoteFaceAuthenticationAuthorities = {},
    options: RemoteFaceAuthenticationOptions = {},
  ) {
    this.now = options.now ?? Date.now;
    this.maxRequestLifetimeMs = Math.min(
      options.maxRequestLifetimeMs ?? DEFAULT_MAX_REQUEST_LIFETIME_MS,
      DEFAULT_MAX_REQUEST_LIFETIME_MS,
    );
    this.maxLivenessAgeMs = Math.min(
      options.maxLivenessAgeMs ?? DEFAULT_MAX_LIVENESS_AGE_MS,
      DEFAULT_MAX_LIVENESS_AGE_MS,
    );
    this.minimumRecoveryDelayMs = Math.max(
      options.minimumRecoveryDelayMs ?? DEFAULT_MINIMUM_RECOVERY_DELAY_MS,
      DEFAULT_MINIMUM_RECOVERY_DELAY_MS,
    );
  }

  async authenticate(
    request: RemoteFaceAuthenticationRequest,
  ): Promise<RemoteFaceAuthenticationResult> {
    const {
      profileAuthority,
      activeLivenessValidator,
      passiveLivenessValidator,
      deviceSessionAttestationValidator,
      possessionValidator,
      recoveryFactorValidator,
      recoveryPolicyValidator,
      nonceGuard,
      finalAuthorizationAuthority,
    } = this.authorities;
    if (
      !profileAuthority ||
      !activeLivenessValidator ||
      !passiveLivenessValidator ||
      !deviceSessionAttestationValidator ||
      !possessionValidator ||
      !recoveryFactorValidator ||
      !recoveryPolicyValidator ||
      !nonceGuard ||
      !finalAuthorizationAuthority
    ) {
      fail("unavailable", "Remote-face authentication is unavailable");
    }

    assertNoBiometricData(request, "request");
    validateProfileShape(request.profile);
    validateBindingShape(request.binding);
    requireString(request.nonce, "nonce");
    if (request.nonce.length > 256) {
      fail("invalid_request", "nonce is too long");
    }
    const now = this.now();
    if (
      !Number.isFinite(request.issuedAt) ||
      !Number.isFinite(request.expiresAt) ||
      request.issuedAt > now ||
      request.expiresAt <= request.issuedAt ||
      request.expiresAt - request.issuedAt > this.maxRequestLifetimeMs
    ) {
      fail("invalid_request", "Request timestamps are invalid");
    }
    if (request.expiresAt <= now) fail("expired", "Request has expired");
    if (
      !Array.isArray(request.recoveryFactors) ||
      request.recoveryFactors.length !== 2
    ) {
      fail("missing_recovery", "Exactly two recovery factors are required");
    }
    if (
      request.activeLiveness == null ||
      request.passiveLiveness == null ||
      request.deviceSessionAttestation == null
    ) {
      fail(
        "invalid_evidence",
        "Active, passive, and attestation evidence are required",
      );
    }
    if (request.possessionProof == null) {
      fail("missing_possession", "Possession proof is required");
    }
    if (
      request.recoveryFactors.some((factor) => factor == null) ||
      request.recoveryPolicy == null
    ) {
      fail("missing_recovery", "Recovery factors and policy are required");
    }
    if (request.remoteFaceResult == null) {
      fail("invalid_evidence", "Opaque supplemental face result is required");
    }

    let projectedProfile: RemoteFaceProfileProjection;
    try {
      projectedProfile = await profileAuthority.validateAndProject(
        request.profile,
      );
    } catch (error) {
      if (error instanceof RemoteFaceAuthenticationError) throw error;
      fail("disabled_profile", "Profile authority rejected the profile");
    }
    assertNoBiometricData(projectedProfile, "projectedProfile");
    validateProfileShape(projectedProfile.profile);
    if (!projectedProfile.enabled)
      fail("disabled_profile", "Profile is disabled");
    if (!profileMatches(projectedProfile.profile, request.profile)) {
      fail("profile_mismatch", "Profile projection does not match the request");
    }

    let evidence: RemoteFaceValidatedEvidence;
    try {
      const [
        activeLiveness,
        passiveLiveness,
        deviceSessionAttestation,
        possession,
        firstRecovery,
        secondRecovery,
        recoveryPolicy,
      ] = await Promise.all([
        activeLivenessValidator.validateAndProject(request.activeLiveness),
        passiveLivenessValidator.validateAndProject(request.passiveLiveness),
        deviceSessionAttestationValidator.validateAndProject(
          request.deviceSessionAttestation,
        ),
        possessionValidator.validateAndProject(request.possessionProof),
        recoveryFactorValidator.validateAndProject(request.recoveryFactors[0]),
        recoveryFactorValidator.validateAndProject(request.recoveryFactors[1]),
        recoveryPolicyValidator.validateAndProject(request.recoveryPolicy),
      ]);
      evidence = {
        activeLiveness,
        passiveLiveness,
        deviceSessionAttestation,
        possession,
        recoveryFactors: [firstRecovery, secondRecovery],
        recoveryPolicy,
      };
    } catch (error) {
      if (error instanceof RemoteFaceAuthenticationError) throw error;
      fail("invalid_evidence", "An evidence authority rejected the request");
    }
    assertNoBiometricData(evidence, "projectedEvidence");

    for (const projection of [
      evidence.activeLiveness,
      evidence.passiveLiveness,
      evidence.deviceSessionAttestation,
    ]) {
      requireString(projection.evidenceRef, "evidence.evidenceRef");
      if (!projection.valid) fail("invalid_evidence", "Evidence is not valid");
      if (!profileMatches(projection.profile, request.profile)) {
        fail("profile_mismatch", "Evidence profile does not match");
      }
      if (!bindingMatches(projection.binding, request.binding)) {
        fail("binding_mismatch", "Evidence binding does not match");
      }
      if (
        !Number.isFinite(projection.validatedAt) ||
        !Number.isFinite(projection.expiresAt) ||
        projection.validatedAt > now ||
        projection.expiresAt <= now ||
        projection.expiresAt > request.expiresAt
      ) {
        fail("invalid_evidence", "Evidence validity window does not match");
      }
    }

    for (const liveness of [
      evidence.activeLiveness,
      evidence.passiveLiveness,
    ]) {
      if (
        !Number.isFinite(liveness.observedAt) ||
        liveness.observedAt > now ||
        now - liveness.observedAt > this.maxLivenessAgeMs
      ) {
        fail("stale_liveness", "Liveness evidence is stale");
      }
    }
    if (
      evidence.activeLiveness.evidenceRef ===
      evidence.passiveLiveness.evidenceRef
    ) {
      fail("invalid_evidence", "Liveness evidence must be independent");
    }

    const possession = evidence.possession;
    requireString(possession.evidenceRef, "possession.evidenceRef");
    if (!possession.valid) {
      fail("missing_possession", "Possession proof is required");
    }
    if (!profileMatches(possession.profile, request.profile)) {
      fail("profile_mismatch", "Possession profile does not match");
    }
    if (!bindingMatches(possession.binding, request.binding)) {
      fail("binding_mismatch", "Possession binding does not match");
    }
    if (
      !Number.isFinite(possession.validatedAt) ||
      !Number.isFinite(possession.expiresAt) ||
      possession.validatedAt > now ||
      possession.expiresAt <= now ||
      possession.expiresAt > request.expiresAt
    ) {
      fail("invalid_evidence", "Possession validity window does not match");
    }

    const recoveryRefs = evidence.recoveryFactors.map(
      (factor) => factor.factorRef,
    );
    for (const factor of evidence.recoveryFactors) {
      requireString(factor.factorRef, "recovery.factorRef");
      if (
        !factor.valid ||
        factor.factorClass !== "recovery" ||
        factor.accountId !== request.binding.accountId ||
        !Number.isFinite(factor.validatedAt) ||
        !Number.isFinite(factor.expiresAt) ||
        factor.validatedAt > now ||
        factor.expiresAt <= now ||
        factor.expiresAt > request.expiresAt
      ) {
        fail("missing_recovery", "Recovery factor is invalid");
      }
    }
    if (new Set(recoveryRefs).size !== 2) {
      fail("missing_recovery", "Recovery factors must be independent");
    }
    const policy = evidence.recoveryPolicy;
    if (
      !policy.valid ||
      policy.accountId !== request.binding.accountId ||
      !Number.isFinite(policy.threshold) ||
      !Number.isFinite(policy.delayMs) ||
      !Number.isFinite(policy.expiresAt) ||
      policy.threshold < 2 ||
      policy.delayMs < this.minimumRecoveryDelayMs ||
      policy.expiresAt <= now ||
      policy.expiresAt > request.expiresAt ||
      policy.factorRefs.length !== 2 ||
      !recoveryRefs.every((factorRef) => policy.factorRefs.includes(factorRef))
    ) {
      fail("missing_recovery", "Delayed threshold recovery policy is invalid");
    }

    if (!(await nonceGuard.consume(request.nonce, request.expiresAt))) {
      fail("replayed_nonce", "Nonce was already consumed");
    }

    let authorization: RemoteFaceFinalAuthorizationProjection;
    try {
      authorization = await finalAuthorizationAuthority.authorizeAndProject({
        profile: request.profile,
        binding: request.binding,
        nonce: request.nonce,
        expiresAt: request.expiresAt,
        evidence,
        remoteFaceResult: request.remoteFaceResult,
      });
    } catch (error) {
      if (error instanceof RemoteFaceAuthenticationError) throw error;
      fail("authorization_denied", "Final authority denied authorization");
    }
    assertNoBiometricData(authorization, "authorization");
    requireString(
      authorization.authorizationRef,
      "authorization.authorizationRef",
    );
    if (!authorization.authorized) {
      fail("authorization_denied", "Final authority denied authorization");
    }
    if (
      authorization.nonce !== request.nonce ||
      !profileMatches(authorization.profile, request.profile) ||
      !bindingMatches(authorization.binding, request.binding) ||
      !Number.isFinite(authorization.expiresAt) ||
      authorization.expiresAt <= now ||
      authorization.expiresAt > request.expiresAt
    ) {
      fail("authorization_mismatch", "Final authorization does not match");
    }

    return {
      authorized: true,
      authorizationRef: authorization.authorizationRef,
      expiresAt: authorization.expiresAt,
    };
  }
}
