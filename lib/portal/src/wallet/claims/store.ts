export interface ClaimKeyIdentity {
  epoch: string;
  fingerprint: string;
}

export interface DerivedClaimMetadata {
  credentialType?: string;
  status?: string;
  issuerReference?: string;
  statusReference?: string;
}

export interface PersistedDerivedClaim extends DerivedClaimMetadata {
  id: string;
  ciphertext: string;
  wrappedDek: string;
  keyEpoch: string;
  keyFingerprint: string;
  createdAt: string;
  updatedAt: string;
}

export interface PersistedClaimEnvelope {
  version: 1;
  keyEpoch: string;
  keyFingerprint: string;
  recoveryReference?: string;
  claims: PersistedDerivedClaim[];
}

export interface ClaimPersistence {
  load(): Promise<unknown | null>;
  replace(envelope: PersistedClaimEnvelope): Promise<void>;
}

export interface ClaimEncryptionAuthority {
  encrypt(plaintext: Uint8Array, dek: Uint8Array): Promise<string>;
  decrypt(ciphertext: string, dek: Uint8Array): Promise<Uint8Array>;
}

export interface ClaimKeySession {
  createDek(): Promise<Uint8Array>;
  unwrapDek(wrappedDek: string): Promise<Uint8Array>;
  wrapDek(dek: Uint8Array): Promise<string>;
  close?(): void | Promise<void>;
}

export interface ClaimKeyAuthority {
  unlock(identity: ClaimKeyIdentity): Promise<ClaimKeySession>;
  recover?(
    reference: string,
    identity: ClaimKeyIdentity,
  ): Promise<ClaimKeySession>;
}

export interface DerivedClaimStoreDependencies {
  persistence?: ClaimPersistence;
  encryption?: ClaimEncryptionAuthority;
  keys?: ClaimKeyAuthority;
  now?: () => Date;
}

export class ClaimStoreUnavailableError extends Error {
  constructor() {
    super("Encrypted claim store authorities are unavailable");
    this.name = "ClaimStoreUnavailableError";
  }
}

export class ClaimStoreLockedError extends Error {
  constructor() {
    super("Encrypted claim store is locked");
    this.name = "ClaimStoreLockedError";
  }
}

export class ClaimStoreStaleKeyError extends Error {
  constructor() {
    super("Claim key epoch or fingerprint is stale");
    this.name = "ClaimStoreStaleKeyError";
  }
}

export class InvalidPersistedClaimError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "InvalidPersistedClaimError";
  }
}

const ENVELOPE_KEYS = new Set([
  "version",
  "keyEpoch",
  "keyFingerprint",
  "recoveryReference",
  "claims",
]);
const CLAIM_KEYS = new Set([
  "id",
  "ciphertext",
  "wrappedDek",
  "credentialType",
  "status",
  "keyEpoch",
  "keyFingerprint",
  "issuerReference",
  "statusReference",
  "createdAt",
  "updatedAt",
]);
const FORBIDDEN_FIELD =
  /(?:kek|unwrapped|private|signature|access.?token|document|image|ocr|embedding|evidence|plaintext|claim.?value)/i;

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function requireOpaqueString(
  value: unknown,
  path: string,
): asserts value is string {
  if (typeof value !== "string" || value.length === 0) {
    throw new InvalidPersistedClaimError(
      `${path} must be a non-empty opaque string`,
    );
  }
}

function rejectForbiddenValues(value: unknown, path: string): void {
  if (value instanceof ArrayBuffer || ArrayBuffer.isView(value)) {
    throw new InvalidPersistedClaimError(
      `${path} must not contain binary data`,
    );
  }
  if (Array.isArray(value)) {
    value.forEach((item, index) =>
      rejectForbiddenValues(item, `${path}[${index}]`),
    );
    return;
  }
  if (!isRecord(value)) return;
  for (const [key, child] of Object.entries(value)) {
    if (FORBIDDEN_FIELD.test(key)) {
      throw new InvalidPersistedClaimError(`${path}.${key} is forbidden`);
    }
    rejectForbiddenValues(child, `${path}.${key}`);
  }
}

function assertExactKeys(
  value: Record<string, unknown>,
  allowed: Set<string>,
  path: string,
): void {
  for (const key of Object.keys(value)) {
    if (!allowed.has(key))
      throw new InvalidPersistedClaimError(`${path}.${key} is not allowed`);
  }
}

export function validatePersistedClaimEnvelope(
  value: unknown,
): PersistedClaimEnvelope {
  rejectForbiddenValues(value, "claimStore");
  if (!isRecord(value))
    throw new InvalidPersistedClaimError("claimStore must be an object");
  assertExactKeys(value, ENVELOPE_KEYS, "claimStore");
  if (value.version !== 1)
    throw new InvalidPersistedClaimError("claimStore.version is unsupported");
  requireOpaqueString(value.keyEpoch, "claimStore.keyEpoch");
  requireOpaqueString(value.keyFingerprint, "claimStore.keyFingerprint");
  if (value.recoveryReference !== undefined) {
    requireOpaqueString(
      value.recoveryReference,
      "claimStore.recoveryReference",
    );
  }
  if (!Array.isArray(value.claims)) {
    throw new InvalidPersistedClaimError("claimStore.claims must be an array");
  }
  const ids = new Set<string>();
  for (const [index, claim] of value.claims.entries()) {
    const path = `claimStore.claims[${index}]`;
    if (!isRecord(claim))
      throw new InvalidPersistedClaimError(`${path} must be an object`);
    assertExactKeys(claim, CLAIM_KEYS, path);
    for (const field of [
      "id",
      "ciphertext",
      "wrappedDek",
      "keyEpoch",
      "keyFingerprint",
      "createdAt",
      "updatedAt",
    ]) {
      requireOpaqueString(claim[field], `${path}.${field}`);
    }
    for (const field of [
      "credentialType",
      "status",
      "issuerReference",
      "statusReference",
    ]) {
      if (claim[field] !== undefined)
        requireOpaqueString(claim[field], `${path}.${field}`);
    }
    if (
      claim.keyEpoch !== value.keyEpoch ||
      claim.keyFingerprint !== value.keyFingerprint
    ) {
      throw new InvalidPersistedClaimError(
        `${path} key identity does not match its envelope`,
      );
    }
    if (ids.has(claim.id as string))
      throw new InvalidPersistedClaimError(`${path}.id is duplicated`);
    ids.add(claim.id as string);
  }
  return structuredClone(value) as unknown as PersistedClaimEnvelope;
}

function zero(bytes: Uint8Array | undefined): void {
  bytes?.fill(0);
}

export class EncryptedDerivedClaimStore {
  private readonly currentKey: ClaimKeyIdentity;
  private envelope: PersistedClaimEnvelope | null = null;
  private session: ClaimKeySession | null = null;
  private readonly deks = new Map<string, Uint8Array>();
  private operation: Promise<void> = Promise.resolve();

  constructor(
    currentKey: ClaimKeyIdentity,
    private readonly dependencies: DerivedClaimStoreDependencies = {},
  ) {
    this.currentKey = { ...currentKey };
  }

  isAvailable(): boolean {
    return Boolean(
      this.dependencies.persistence &&
      this.dependencies.encryption &&
      this.dependencies.keys,
    );
  }

  isLocked(): boolean {
    return this.session === null;
  }

  unlock(identity: ClaimKeyIdentity): Promise<void> {
    return this.exclusive(async () => {
      this.requireDependencies();
      this.assertCurrent(identity);
      const envelope = await this.loadEnvelope();
      this.assertEnvelopeCurrent(envelope);
      const session = await this.dependencies.keys!.unlock(identity);
      await this.activate(session, envelope);
    });
  }

  unlockWithRecovery(identity: ClaimKeyIdentity): Promise<void> {
    return this.exclusive(async () => {
      this.requireDependencies();
      this.assertCurrent(identity);
      const envelope = await this.loadEnvelope();
      this.assertEnvelopeCurrent(envelope);
      if (!envelope.recoveryReference || !this.dependencies.keys!.recover) {
        throw new ClaimStoreUnavailableError();
      }
      const session = await this.dependencies.keys!.recover(
        envelope.recoveryReference,
        identity,
      );
      await this.activate(session, envelope);
    });
  }

  lock(): Promise<void> {
    return this.exclusive(async () => this.clearUnlockedState());
  }

  importRecoveryReference(reference: string): Promise<void> {
    return this.exclusive(async () => {
      this.requireDependencies();
      requireOpaqueString(reference, "recoveryReference");
      const envelope = await this.loadEnvelope();
      this.assertEnvelopeCurrent(envelope);
      const replacement = { ...envelope, recoveryReference: reference };
      await this.persist(replacement);
      this.envelope = replacement;
    });
  }

  write(
    id: string,
    plaintext: Uint8Array,
    metadata: DerivedClaimMetadata = {},
  ): Promise<void> {
    const plaintextCopy = new Uint8Array(plaintext);
    return this.exclusive(async () => {
      let generatedDek: Uint8Array | undefined;
      try {
        this.requireUnlocked();
        requireOpaqueString(id, "claim.id");
        const existing = this.envelope!.claims.find((claim) => claim.id === id);
        let dek = this.deks.get(id);
        if (!dek) {
          generatedDek = await this.session!.createDek();
          this.requireKeyBytes(generatedDek);
          dek = generatedDek;
        }
        const ciphertext = await this.dependencies.encryption!.encrypt(
          plaintextCopy,
          dek,
        );
        requireOpaqueString(ciphertext, "claim.ciphertext");
        const wrappedDek =
          existing?.wrappedDek ?? (await this.session!.wrapDek(dek));
        const timestamp = (
          this.dependencies.now ?? (() => new Date())
        )().toISOString();
        const claim: PersistedDerivedClaim = {
          id,
          ciphertext,
          wrappedDek,
          ...(metadata.credentialType === undefined
            ? {}
            : { credentialType: metadata.credentialType }),
          ...(metadata.status === undefined ? {} : { status: metadata.status }),
          ...(metadata.issuerReference === undefined
            ? {}
            : { issuerReference: metadata.issuerReference }),
          ...(metadata.statusReference === undefined
            ? {}
            : { statusReference: metadata.statusReference }),
          keyEpoch: this.currentKey.epoch,
          keyFingerprint: this.currentKey.fingerprint,
          createdAt: existing?.createdAt ?? timestamp,
          updatedAt: timestamp,
        };
        const claims = existing
          ? this.envelope!.claims.map((item) => (item.id === id ? claim : item))
          : [...this.envelope!.claims, claim];
        await this.persist({ ...this.envelope!, claims });
        this.envelope = { ...this.envelope!, claims };
        if (!existing) {
          this.deks.set(id, dek);
          generatedDek = undefined;
        }
      } finally {
        zero(plaintextCopy);
        zero(generatedDek);
      }
    });
  }

  read(id: string): Promise<Uint8Array | null> {
    return this.exclusive(async () => {
      this.requireUnlocked();
      const claim = this.envelope!.claims.find((item) => item.id === id);
      if (!claim) return null;
      const dek = this.deks.get(id);
      if (!dek) throw new ClaimStoreLockedError();
      return this.dependencies.encryption!.decrypt(claim.ciphertext, dek);
    });
  }

  rotate(identity: ClaimKeyIdentity, keys: ClaimKeyAuthority): Promise<void> {
    return this.exclusive(async () => {
      this.requireUnlocked();
      requireOpaqueString(identity.epoch, "key.epoch");
      requireOpaqueString(identity.fingerprint, "key.fingerprint");
      const nextSession = await keys.unlock(identity);
      try {
        const claims: PersistedDerivedClaim[] = [];
        for (const claim of this.envelope!.claims) {
          const dek = this.deks.get(claim.id);
          if (!dek) throw new ClaimStoreLockedError();
          claims.push({
            ...claim,
            wrappedDek: await nextSession.wrapDek(dek),
            keyEpoch: identity.epoch,
            keyFingerprint: identity.fingerprint,
          });
        }
        const replacement = {
          ...this.envelope!,
          keyEpoch: identity.epoch,
          keyFingerprint: identity.fingerprint,
          claims,
        };
        await this.persist(replacement);
        const previousSession = this.session;
        this.session = nextSession;
        this.envelope = replacement;
        this.currentKey.epoch = identity.epoch;
        this.currentKey.fingerprint = identity.fingerprint;
        try {
          await previousSession?.close?.();
        } catch {
          // Rotation is already committed and the new session is authoritative.
        }
      } catch (error) {
        await nextSession.close?.();
        throw error;
      }
    });
  }

  private exclusive<T>(operation: () => Promise<T>): Promise<T> {
    const result = this.operation.then(operation, operation);
    this.operation = result.then(
      () => undefined,
      () => undefined,
    );
    return result;
  }

  private requireDependencies(): void {
    if (!this.isAvailable()) throw new ClaimStoreUnavailableError();
  }

  private requireUnlocked(): void {
    this.requireDependencies();
    if (!this.session || !this.envelope) throw new ClaimStoreLockedError();
  }

  private assertCurrent(identity: ClaimKeyIdentity): void {
    if (
      identity.epoch !== this.currentKey.epoch ||
      identity.fingerprint !== this.currentKey.fingerprint
    ) {
      throw new ClaimStoreStaleKeyError();
    }
  }

  private assertEnvelopeCurrent(envelope: PersistedClaimEnvelope): void {
    if (
      envelope.keyEpoch !== this.currentKey.epoch ||
      envelope.keyFingerprint !== this.currentKey.fingerprint
    ) {
      throw new ClaimStoreStaleKeyError();
    }
  }

  private requireKeyBytes(value: unknown): asserts value is Uint8Array {
    if (!(value instanceof Uint8Array) || value.length === 0) {
      throw new InvalidPersistedClaimError(
        "Key authority returned invalid DEK bytes",
      );
    }
  }

  private async loadEnvelope(): Promise<PersistedClaimEnvelope> {
    const stored = await this.dependencies.persistence!.load();
    if (stored === null) {
      return {
        version: 1,
        keyEpoch: this.currentKey.epoch,
        keyFingerprint: this.currentKey.fingerprint,
        claims: [],
      };
    }
    return validatePersistedClaimEnvelope(stored);
  }

  private async activate(
    session: ClaimKeySession,
    envelope: PersistedClaimEnvelope,
  ): Promise<void> {
    const opened = new Map<string, Uint8Array>();
    try {
      for (const claim of envelope.claims) {
        const dek = await session.unwrapDek(claim.wrappedDek);
        this.requireKeyBytes(dek);
        opened.set(claim.id, dek);
      }
      await this.clearUnlockedState();
      this.session = session;
      this.envelope = envelope;
      opened.forEach((dek, id) => this.deks.set(id, dek));
    } catch (error) {
      opened.forEach(zero);
      await session.close?.();
      throw error;
    }
  }

  private async clearUnlockedState(): Promise<void> {
    this.deks.forEach(zero);
    this.deks.clear();
    await this.session?.close?.();
    this.session = null;
    this.envelope = null;
  }

  private async persist(envelope: PersistedClaimEnvelope): Promise<void> {
    const safe = validatePersistedClaimEnvelope(envelope);
    await this.dependencies.persistence!.replace(safe);
  }
}
