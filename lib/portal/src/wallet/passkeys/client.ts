export type PasskeyCeremony = "registration" | "authentication";

export interface PasskeyBinding {
  chainId: string;
  account: string;
  sessionId: string;
}

export type PasskeyAction =
  | { kind: "login"; purpose: string }
  | {
      kind: "authorization";
      payloadKind: "transaction" | "message";
      digest: string;
    };

export interface PasskeyChallengeProjection {
  challenge: Uint8Array;
  clientDataChallenge: string;
}

export interface PasskeyChallengeProjector {
  project(input: {
    ceremony: PasskeyCeremony;
    rpId: string;
    binding: PasskeyBinding;
    action: PasskeyAction;
    expiresAt: number;
  }): Promise<PasskeyChallengeProjection>;
}

export interface PasskeyRegistrationRequest extends PasskeyBinding {
  rp: { id: string; name: string };
  user: { id: Uint8Array; name: string; displayName: string };
  sessionPurpose: string;
  expiresAt: number;
}

export interface PasskeyAuthenticationRequest extends PasskeyBinding {
  action: PasskeyAction;
  expiresAt: number;
  credentialId?: string;
}

export interface PlatformRegistrationOptions {
  challenge: Uint8Array;
  rp: { id: string; name: string };
  user: { id: Uint8Array; name: string; displayName: string };
  pubKeyCredParams: readonly { type: "public-key"; alg: number }[];
  authenticatorSelection: {
    authenticatorAttachment: "platform";
    residentKey: "required";
    requireResidentKey: true;
    userVerification: "required";
  };
  attestation: AttestationConveyancePreference;
}

export interface PlatformAuthenticationOptions {
  challenge: Uint8Array;
  rpId: string;
  userVerification: "required";
  allowCredentials?: readonly {
    id: Uint8Array;
    type: "public-key";
    transports?: readonly AuthenticatorTransport[];
  }[];
}

export interface PlatformRegistrationCredential {
  id: string;
  rawId: Uint8Array;
  type: "public-key";
  authenticatorAttachment: "platform" | null;
  response: {
    clientDataJSON: Uint8Array;
    attestationObject: Uint8Array;
    transports?: readonly AuthenticatorTransport[];
  };
}

export interface PlatformAssertionCredential {
  id: string;
  rawId: Uint8Array;
  type: "public-key";
  authenticatorAttachment: "platform" | null;
  response: {
    clientDataJSON: Uint8Array;
    authenticatorData: Uint8Array;
    signature: Uint8Array;
    userHandle: Uint8Array | null;
  };
}

export interface PlatformPasskeyAuthority {
  isPlatformAvailable(): boolean;
  create(options: PlatformRegistrationOptions): Promise<PlatformRegistrationCredential>;
  get(options: PlatformAuthenticationOptions): Promise<PlatformAssertionCredential>;
}

export interface PasskeyVerificationFacts {
  verified: boolean;
  rpIdHashValid: boolean;
  userPresent: boolean;
  userVerified: boolean;
  credentialId: string;
  counter: number;
}

export interface PasskeyServerVerifier {
  verifyRegistration(input: {
    credential: PlatformRegistrationCredential;
    rpId: string;
    origin: string;
    binding: PasskeyBinding;
    projection: PasskeyChallengeProjection;
  }): Promise<PasskeyVerificationFacts>;
  verifyAssertion(input: {
    credential: PlatformAssertionCredential;
    rpId: string;
    origin: string;
    binding: PasskeyBinding;
    action: PasskeyAction;
    projection: PasskeyChallengeProjection;
  }): Promise<PasskeyVerificationFacts>;
}

export interface PasskeyCounterStore {
  initialize(credentialId: string, counter: number): Promise<boolean>;
  advance(credentialId: string, counter: number): Promise<boolean>;
}

export interface PasskeyAttestationPolicy {
  registrationPreference(): AttestationConveyancePreference;
}

export interface PasskeyClientDependencies {
  authority?: PlatformPasskeyAuthority;
  verifier?: PasskeyServerVerifier;
  challengeProjector?: PasskeyChallengeProjector;
  counterStore?: PasskeyCounterStore;
  attestationPolicy?: PasskeyAttestationPolicy;
  now?: () => number;
}

export interface PasskeyRegistrationResult {
  credentialId: string;
  rawCredentialId: Uint8Array;
  clientDataJSON: Uint8Array;
  attestationObject: Uint8Array;
  transports: readonly AuthenticatorTransport[];
  counter: number;
}

export interface PasskeyAssertionResult {
  credentialId: string;
  rawCredentialId: Uint8Array;
  clientDataJSON: Uint8Array;
  authenticatorData: Uint8Array;
  signature: Uint8Array;
  userHandle: Uint8Array | null;
  counter: number;
}

export type PasskeyErrorCode =
  | "unavailable"
  | "unsupported_platform"
  | "invalid_request"
  | "expired"
  | "invalid_client_data"
  | "verification_failed"
  | "counter_replay";

export class PasskeyError extends Error {
  constructor(
    readonly code: PasskeyErrorCode,
    message: string,
  ) {
    super(message);
    this.name = "PasskeyError";
  }
}

const DEFAULT_ALGORITHMS = [
  { type: "public-key" as const, alg: -7 },
  { type: "public-key" as const, alg: -257 },
];

function requireText(value: unknown, field: string): asserts value is string {
  if (typeof value !== "string" || value.length === 0) {
    throw new PasskeyError("invalid_request", `${field} is required`);
  }
}

function validateBinding(binding: PasskeyBinding): void {
  requireText(binding.chainId, "chainId");
  requireText(binding.account, "account");
  requireText(binding.sessionId, "sessionId");
}

function validateAction(action: PasskeyAction): void {
  if (!action || typeof action !== "object") {
    throw new PasskeyError("invalid_request", "an explicit action is required");
  }
  const candidate = action as PasskeyAction & Record<string, unknown>;
  if (candidate.kind === "login") {
    requireText(candidate.purpose, "login purpose");
    if ("digest" in candidate || "payloadKind" in candidate) {
      throw new PasskeyError("invalid_request", "login action is ambiguous");
    }
    return;
  }
  if (candidate.kind === "authorization") {
    requireText(candidate.digest, "authorization digest");
    if (candidate.payloadKind !== "transaction" && candidate.payloadKind !== "message") {
      throw new PasskeyError("invalid_request", "payloadKind is invalid");
    }
    if ("purpose" in candidate) {
      throw new PasskeyError("invalid_request", "authorization action is ambiguous");
    }
    return;
  }
  throw new PasskeyError("invalid_request", "action kind is invalid");
}

function parseClientData(bytes: Uint8Array): {
  type: string;
  challenge: string;
  origin: string;
} {
  try {
    const parsed: unknown = JSON.parse(new TextDecoder("utf-8", { fatal: true }).decode(bytes));
    if (!parsed || typeof parsed !== "object") throw new Error("not an object");
    const data = parsed as Record<string, unknown>;
    if (
      typeof data.type !== "string" ||
      typeof data.challenge !== "string" ||
      typeof data.origin !== "string"
    ) {
      throw new Error("required fields missing");
    }
    return { type: data.type, challenge: data.challenge, origin: data.origin };
  } catch {
    throw new PasskeyError("invalid_client_data", "clientDataJSON is invalid");
  }
}

function validateClientData(
  bytes: Uint8Array,
  expectedType: "webauthn.create" | "webauthn.get",
  projection: PasskeyChallengeProjection,
  origin: string,
): void {
  const data = parseClientData(bytes);
  if (
    data.type !== expectedType ||
    data.challenge !== projection.clientDataChallenge ||
    data.origin !== origin
  ) {
    throw new PasskeyError("invalid_client_data", "client data binding mismatch");
  }
}

function validateFacts(facts: PasskeyVerificationFacts, credentialId: string): void {
  if (
    !facts.verified ||
    !facts.rpIdHashValid ||
    !facts.userPresent ||
    !facts.userVerified ||
    facts.credentialId !== credentialId ||
    !Number.isSafeInteger(facts.counter) ||
    facts.counter < 0
  ) {
    throw new PasskeyError("verification_failed", "server verification rejected the passkey");
  }
}

function clone(bytes: Uint8Array): Uint8Array {
  return new Uint8Array(bytes);
}

export class PlatformPasskeyClient {
  private readonly now: () => number;

  constructor(
    private readonly rpId: string,
    private readonly origin: string,
    private readonly dependencies: PasskeyClientDependencies = {},
  ) {
    this.now = dependencies.now ?? Date.now;
  }

  isAvailable(): boolean {
    const { authority, verifier, challengeProjector, counterStore } = this.dependencies;
    return Boolean(
      authority &&
        verifier &&
        challengeProjector &&
        counterStore &&
        authority.isPlatformAvailable(),
    );
  }

  async register(request: PasskeyRegistrationRequest): Promise<PasskeyRegistrationResult> {
    const dependencies = this.availableDependencies();
    validateBinding(request);
    requireText(request.rp.id, "rp.id");
    requireText(request.rp.name, "rp.name");
    requireText(request.user.name, "user.name");
    requireText(request.user.displayName, "user.displayName");
    requireText(request.sessionPurpose, "sessionPurpose");
    if (request.rp.id !== this.rpId || request.user.id.length === 0) {
      throw new PasskeyError("invalid_request", "registration RP or user is invalid");
    }
    this.requireFresh(request.expiresAt);
    const action: PasskeyAction = { kind: "login", purpose: request.sessionPurpose };
    const projection = await dependencies.challengeProjector.project({
      ceremony: "registration",
      rpId: this.rpId,
      binding: request,
      action,
      expiresAt: request.expiresAt,
    });
    const credential = await dependencies.authority.create({
      challenge: clone(projection.challenge),
      rp: { ...request.rp },
      user: { ...request.user, id: clone(request.user.id) },
      pubKeyCredParams: DEFAULT_ALGORITHMS,
      authenticatorSelection: {
        authenticatorAttachment: "platform",
        residentKey: "required",
        requireResidentKey: true,
        userVerification: "required",
      },
      attestation:
        this.dependencies.attestationPolicy?.registrationPreference() ?? "none",
    });
    this.requireFresh(request.expiresAt);
    this.validateCredential(credential);
    validateClientData(
      credential.response.clientDataJSON,
      "webauthn.create",
      projection,
      this.origin,
    );
    const facts = await dependencies.verifier.verifyRegistration({
      credential,
      rpId: this.rpId,
      origin: this.origin,
      binding: request,
      projection,
    });
    validateFacts(facts, credential.id);
    if (!(await dependencies.counterStore.initialize(credential.id, facts.counter))) {
      throw new PasskeyError("counter_replay", "credential counter already exists");
    }
    return {
      credentialId: credential.id,
      rawCredentialId: clone(credential.rawId),
      clientDataJSON: clone(credential.response.clientDataJSON),
      attestationObject: clone(credential.response.attestationObject),
      transports: [...(credential.response.transports ?? [])],
      counter: facts.counter,
    };
  }

  async authenticate(request: PasskeyAuthenticationRequest): Promise<PasskeyAssertionResult> {
    const dependencies = this.availableDependencies();
    validateBinding(request);
    validateAction(request.action);
    this.requireFresh(request.expiresAt);
    const projection = await dependencies.challengeProjector.project({
      ceremony: "authentication",
      rpId: this.rpId,
      binding: request,
      action: request.action,
      expiresAt: request.expiresAt,
    });
    const credential = await dependencies.authority.get({
      challenge: clone(projection.challenge),
      rpId: this.rpId,
      userVerification: "required",
    });
    this.requireFresh(request.expiresAt);
    this.validateCredential(credential);
    if (request.credentialId && credential.id !== request.credentialId) {
      throw new PasskeyError("verification_failed", "credential ID mismatch");
    }
    validateClientData(
      credential.response.clientDataJSON,
      "webauthn.get",
      projection,
      this.origin,
    );
    const facts = await dependencies.verifier.verifyAssertion({
      credential,
      rpId: this.rpId,
      origin: this.origin,
      binding: request,
      action: request.action,
      projection,
    });
    validateFacts(facts, credential.id);
    if (!(await dependencies.counterStore.advance(credential.id, facts.counter))) {
      throw new PasskeyError("counter_replay", "credential counter did not advance");
    }
    return {
      credentialId: credential.id,
      rawCredentialId: clone(credential.rawId),
      clientDataJSON: clone(credential.response.clientDataJSON),
      authenticatorData: clone(credential.response.authenticatorData),
      signature: clone(credential.response.signature),
      userHandle: credential.response.userHandle
        ? clone(credential.response.userHandle)
        : null,
      counter: facts.counter,
    };
  }

  private availableDependencies(): Required<
    Pick<
      PasskeyClientDependencies,
      "authority" | "verifier" | "challengeProjector" | "counterStore"
    >
  > {
    const { authority, verifier, challengeProjector, counterStore } = this.dependencies;
    if (!authority || !verifier || !challengeProjector || !counterStore) {
      throw new PasskeyError("unavailable", "passkey dependencies are unavailable");
    }
    if (!authority.isPlatformAvailable()) {
      throw new PasskeyError("unsupported_platform", "platform passkeys are unavailable");
    }
    return { authority, verifier, challengeProjector, counterStore };
  }

  private requireFresh(expiresAt: number): void {
    if (!Number.isFinite(expiresAt) || expiresAt <= this.now()) {
      throw new PasskeyError("expired", "passkey challenge has expired");
    }
  }

  private validateCredential(
    credential: PlatformRegistrationCredential | PlatformAssertionCredential,
  ): void {
    if (
      credential.type !== "public-key" ||
      credential.authenticatorAttachment !== "platform" ||
      credential.id.length === 0 ||
      credential.rawId.length === 0
    ) {
      throw new PasskeyError("verification_failed", "non-platform credential rejected");
    }
  }
}