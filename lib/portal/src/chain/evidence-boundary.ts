/**
 * Dependency-neutral boundary between evidence ingestion and chain submission.
 */

const FORBIDDEN_EVIDENCE_FIELDS = new Set([
  "rawimage",
  "rawimages",
  "image",
  "images",
  "ocr",
  "ocrtext",
  "embedding",
  "embeddings",
  "encryptedpayload",
  "ciphertext",
  "recipient",
  "recipients",
  "recipientlist",
  "recipientkeyids",
  "wrappedkeys",
  "envelope",
  "envelopes",
  "recipientenvelope",
  "fullrecipientenvelope",
  "completeenvelope",
]);

const MSG_UPLOAD_SCOPE = "/virtengine.veid.v1.MsgUploadScope";
const TYPE_DISCRIMINATOR_FIELDS = new Set(["type", "typename", "typeurl"]);

export type OpaqueEvidenceReference = unknown;
export type OpaqueCommitmentStatusRequest = unknown;

export interface OffChainEvidenceIngestionTransport<
  TPayload,
  TReference = OpaqueEvidenceReference,
> {
  ingest(payload: TPayload): Promise<TReference>;
}

export interface T5EvidenceReferenceAdapter<
  TReference = OpaqueEvidenceReference,
  TRequest = OpaqueCommitmentStatusRequest,
> {
  validateAndBuildCommitmentRequest(reference: TReference): Promise<TRequest>;
}

export interface EvidenceChainSubmitter<
  TRequest = OpaqueCommitmentStatusRequest,
  TResult = unknown,
> {
  submit(request: TRequest): Promise<TResult>;
}

export interface EvidenceReconnectStore<TMetadata extends object> {
  persist(metadata: TMetadata): Promise<void>;
}

export interface EvidenceBoundaryDependencies<
  TPayload,
  TReference,
  TRequest,
  TResult,
  TMetadata extends object,
> {
  ingestionTransport?: OffChainEvidenceIngestionTransport<TPayload, TReference>;
  referenceAdapter?: T5EvidenceReferenceAdapter<TReference, TRequest>;
  chainSubmitter?: EvidenceChainSubmitter<TRequest, TResult>;
  reconnectStore?: EvidenceReconnectStore<TMetadata>;
}

export interface EvidenceSubmissionResult<TReference, TResult> {
  reference: TReference;
  chainResult: TResult;
}

export class EvidenceBoundaryUnavailableError extends Error {
  constructor(message = "Evidence submission boundary is unavailable") {
    super(message);
    this.name = "EvidenceBoundaryUnavailableError";
  }
}

export class ForbiddenEvidenceError extends Error {
  readonly path: string;

  constructor(path: string, reason: string) {
    super(`Forbidden evidence at ${path}: ${reason}`);
    this.name = "ForbiddenEvidenceError";
    this.path = path;
  }
}

function normalizedFieldName(field: string): string {
  return field.replace(/[^a-z0-9]/gi, "").toLowerCase();
}

function isForbiddenEvidenceField(field: string): boolean {
  const normalized = normalizedFieldName(field);
  return (
    FORBIDDEN_EVIDENCE_FIELDS.has(normalized) ||
    normalized.startsWith("ocr") ||
    normalized.startsWith("embedding") ||
    normalized.startsWith("recipient") ||
    normalized.includes("encryptedpayload") ||
    normalized.includes("ciphertext") ||
    normalized.includes("envelope")
  );
}

function isForbiddenEvidenceType(field: string, value: unknown): boolean {
  if (
    typeof value !== "string" ||
    !TYPE_DISCRIMINATOR_FIELDS.has(normalizedFieldName(field))
  ) {
    return false;
  }
  const normalized = normalizedFieldName(value);
  return (
    normalized.includes("msguploadscope") ||
    normalized.includes("evidenceobjectref") ||
    normalized.includes("encryptedpayload") ||
    normalized.includes("ciphertext") ||
    normalized.includes("recipient") ||
    normalized.includes("envelope") ||
    normalized.includes("embedding") ||
    normalized.includes("ocr")
  );
}

function isBinaryEvidence(value: unknown): boolean {
  if (value instanceof ArrayBuffer || ArrayBuffer.isView(value)) return true;
  return typeof Blob !== "undefined" && value instanceof Blob;
}

function snapshotPrivacySafeValue(
  value: unknown,
  path: string,
  ancestors: Set<object>,
): unknown {
  if (isBinaryEvidence(value)) {
    throw new ForbiddenEvidenceError(path, "binary evidence is not permitted");
  }
  if (typeof value === "string" && value.includes(MSG_UPLOAD_SCOPE)) {
    throw new ForbiddenEvidenceError(path, "MsgUploadScope is not permitted");
  }
  if (value === null || typeof value !== "object") return value;
  if (ancestors.has(value)) {
    throw new ForbiddenEvidenceError(path, "cyclic values are not permitted");
  }
  if (
    !Array.isArray(value) &&
    Object.getPrototypeOf(value) !== Object.prototype
  ) {
    throw new ForbiddenEvidenceError(
      path,
      "non-plain objects are not permitted",
    );
  }

  ancestors.add(value);
  try {
    if (Array.isArray(value)) {
      return value.map((entry, index) =>
        snapshotPrivacySafeValue(entry, `${path}[${index}]`, ancestors),
      );
    }

    const snapshot: Record<string, unknown> = {};
    for (const [field, entry] of Object.entries(value)) {
      const fieldPath = `${path}.${field}`;
      if (isForbiddenEvidenceField(field)) {
        throw new ForbiddenEvidenceError(fieldPath, "forbidden evidence field");
      }
      if (isForbiddenEvidenceType(field, entry)) {
        throw new ForbiddenEvidenceError(fieldPath, "forbidden evidence type");
      }
      snapshot[field] = snapshotPrivacySafeValue(entry, fieldPath, ancestors);
    }
    return snapshot;
  } finally {
    ancestors.delete(value);
  }
}

export function snapshotPrivacySafeClientValue<T>(value: T, path = "value"): T {
  return snapshotPrivacySafeValue(value, path, new Set()) as T;
}

export class SupplementalEvidenceBoundary<
  TPayload,
  TReference = OpaqueEvidenceReference,
  TRequest = OpaqueCommitmentStatusRequest,
  TResult = unknown,
  TMetadata extends object = Record<string, unknown>,
> {
  constructor(
    private readonly dependencies: EvidenceBoundaryDependencies<
      TPayload,
      TReference,
      TRequest,
      TResult,
      TMetadata
    > = {},
  ) {}

  isAvailable(): boolean {
    return Boolean(
      this.dependencies.ingestionTransport &&
      this.dependencies.referenceAdapter &&
      this.dependencies.chainSubmitter,
    );
  }

  async submit(
    payload: TPayload,
    reconnectMetadata?: TMetadata,
  ): Promise<EvidenceSubmissionResult<TReference, TResult>> {
    const {
      ingestionTransport,
      referenceAdapter,
      chainSubmitter,
      reconnectStore,
    } = this.dependencies;
    if (!ingestionTransport || !referenceAdapter || !chainSubmitter) {
      throw new EvidenceBoundaryUnavailableError();
    }
    if (reconnectMetadata && !reconnectStore) {
      throw new EvidenceBoundaryUnavailableError(
        "Evidence reconnect persistence is unavailable",
      );
    }

    const safeMetadata = reconnectMetadata
      ? snapshotPrivacySafeClientValue(reconnectMetadata, "reconnectMetadata")
      : undefined;
    const reference = await ingestionTransport.ingest(payload);
    const request =
      await referenceAdapter.validateAndBuildCommitmentRequest(reference);
    const safeRequest = snapshotPrivacySafeClientValue(request, "chainRequest");
    const chainResult = await chainSubmitter.submit(safeRequest);

    if (safeMetadata && reconnectStore)
      await reconnectStore.persist(safeMetadata);
    return { reference, chainResult };
  }
}
