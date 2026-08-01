export const UNSUPPORTED_DOCUMENT_TRIAGE_SCHEMA_VERSION =
  "unsupported-document-triage.v1" as const;

export const UNSUPPORTED_DOCUMENT_REASON_CODES = [
  "CATEGORY_NOT_SUPPORTED",
  "COUNTRY_NOT_SUPPORTED",
  "DOCUMENT_EXPIRED",
  "DOCUMENT_FORMAT_NOT_SUPPORTED",
  "DOCUMENT_QUALITY_INSUFFICIENT",
  "LANGUAGE_NOT_SUPPORTED",
  "REQUIRED_SIDE_MISSING",
] as const;

export const UNSUPPORTED_DOCUMENT_CATEGORIES = [
  "DRIVER_LICENSE",
  "NATIONAL_ID",
  "OTHER_GOVERNMENT_DOCUMENT",
  "PASSPORT",
  "RESIDENCE_PERMIT",
] as const;

export const SUPPORT_LANGUAGE_CODES = [
  "ar",
  "de",
  "en",
  "es",
  "fr",
  "hi",
  "id",
  "ja",
  "ko",
  "pt",
  "zh",
] as const;

export const REDACTED_REASON_TOKENS = [
  "category-mismatch",
  "country-restricted",
  "expired-document",
  "format-unsupported",
  "language-assistance",
  "quality-insufficient",
  "review-required",
  "side-missing",
] as const;

export const SAFE_NEXT_STEP_OPTIONS = [
  "CONTACT_SUPPORT",
  "REQUEST_ACCESSIBLE_ASSISTANCE",
  "RETRY_WITH_SUPPORTED_DOCUMENT",
  "REVIEW_SUPPORTED_DOCUMENTS",
] as const;

export const REDACTED_SUMMARY_TOKENS = [
  "assistance-requested",
  "document-unsupported",
  "manual-review-required",
  ...REDACTED_REASON_TOKENS,
] as const;

export type UnsupportedDocumentReasonCode =
  (typeof UNSUPPORTED_DOCUMENT_REASON_CODES)[number];
export type UnsupportedDocumentCategory =
  (typeof UNSUPPORTED_DOCUMENT_CATEGORIES)[number];
export type SupportLanguageCode = (typeof SUPPORT_LANGUAGE_CODES)[number];
export type RedactedReasonToken = (typeof REDACTED_REASON_TOKENS)[number];
export type SafeNextStepOption = (typeof SAFE_NEXT_STEP_OPTIONS)[number];

export interface UnsupportedDocumentTriageRequest {
  schemaVersion: typeof UNSUPPORTED_DOCUMENT_TRIAGE_SCHEMA_VERSION;
  reasonCodes: UnsupportedDocumentReasonCode[];
  supportLanguageCode: SupportLanguageCode;
  redactedNotes: RedactedReasonToken[];
  documentCategory?: UnsupportedDocumentCategory;
  countryCode?: string;
}

export interface UnsupportedDocumentTriageOutput {
  schemaVersion: typeof UNSUPPORTED_DOCUMENT_TRIAGE_SCHEMA_VERSION;
  unsupportedDocumentCategory: UnsupportedDocumentCategory;
  safeNextStepOptions: SafeNextStepOption[];
  redactedCaseSummary: string;
}

export interface UnsupportedDocumentTriageDraft extends UnsupportedDocumentTriageOutput {
  status: "draft";
  draftReference: string;
}

export interface HumanTriageConfirmation {
  confirmed: true;
  draftReference: string;
}

export interface AcceptedUnsupportedDocumentSupportNote extends UnsupportedDocumentTriageOutput {
  status: "accepted-support-note";
  draftReference: string;
  humanConfirmed: true;
}

export interface UnsupportedDocumentModelAuthority {
  proposeUnsupportedDocumentTriage(
    request: Readonly<UnsupportedDocumentTriageRequest>,
  ): Promise<unknown>;
}

export interface UnsupportedDocumentOutputValidator {
  validate(
    output: Readonly<UnsupportedDocumentTriageOutput>,
  ): void | Promise<void>;
}

export interface TriageDigestAuthority {
  sha256(canonicalValue: string): string | Promise<string>;
}

export interface UnsupportedDocumentTriagePolicy {
  allowDocumentCategory?: boolean;
  allowCountryCode?: boolean;
}

export interface UnsupportedDocumentTriageDependencies {
  modelAuthority?: UnsupportedDocumentModelAuthority;
  outputValidator?: UnsupportedDocumentOutputValidator;
  digestAuthority?: TriageDigestAuthority;
  policy?: UnsupportedDocumentTriagePolicy;
}

const REQUEST_REQUIRED_FIELDS = [
  "schemaVersion",
  "reasonCodes",
  "supportLanguageCode",
  "redactedNotes",
] as const;
const REQUEST_OPTIONAL_FIELDS = ["documentCategory", "countryCode"] as const;
const OUTPUT_FIELDS = [
  "schemaVersion",
  "unsupportedDocumentCategory",
  "safeNextStepOptions",
  "redactedCaseSummary",
] as const;
const DRAFT_FIELDS = [...OUTPUT_FIELDS, "status", "draftReference"] as const;
const CONFIRMATION_FIELDS = ["confirmed", "draftReference"] as const;

const FORBIDDEN_FIELD_NAMES = new Set([
  "account",
  "accountid",
  "address",
  "biometric",
  "biometrics",
  "ciphertext",
  "email",
  "embedding",
  "embeddings",
  "envelope",
  "evidence",
  "evidenceenvelope",
  "fullname",
  "id",
  "identifier",
  "image",
  "images",
  "name",
  "ocr",
  "ocrtext",
  "payment",
  "payout",
  "rawdocument",
  "rawdocuments",
  "rawimage",
  "rawimages",
  "secret",
  "secrets",
  "template",
  "templates",
  "token",
  "tokens",
  "transaction",
  "transactiondetails",
  "tx",
  "wallet",
  "walletaddress",
]);

const PROHIBITED_OUTPUT_FIELDS = new Set([
  "claim",
  "decision",
  "eligibility",
  "fund",
  "identity",
  "mint",
  "payment",
  "payout",
  "policy",
  "score",
  "sideeffect",
  "tool",
  "tools",
  "transaction",
  "uniqueness",
]);

const CONTROL_CHARACTERS = /[\u0000-\u001f\u007f-\u009f]/;
const PROMPT_INJECTION =
  /(?:ignore\s+(?:all\s+)?(?:previous|prior)|system\s*prompt|developer\s*message|assistant\s*:|jailbreak|follow\s+these\s+instructions)/i;
const SHA256_DIGEST = /^[a-f0-9]{64}$/;
const COUNTRY_CODE = /^[A-Z]{2}$/;
const MAX_REASON_CODES = 5;
const MAX_REDACTED_NOTES = 8;
const MAX_NEXT_STEPS = 4;
const MAX_SUMMARY_TOKENS = 12;

export class UnsupportedDocumentTriageValidationError extends Error {
  readonly path: string;

  constructor(path: string, message: string) {
    super(`Invalid unsupported-document triage value at ${path}: ${message}`);
    this.name = "UnsupportedDocumentTriageValidationError";
    this.path = path;
  }
}

export class UnsupportedDocumentTriageUnavailableError extends Error {
  constructor() {
    super("Unsupported-document triage is unavailable");
    this.name = "UnsupportedDocumentTriageUnavailableError";
  }
}

export class UnsupportedDocumentTriageConfirmationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "UnsupportedDocumentTriageConfirmationError";
  }
}

function normalizedFieldName(field: string): string {
  return field.replace(/[^a-z0-9]/gi, "").toLowerCase();
}

function isProhibitedFieldName(
  field: string,
  prohibitedFields: ReadonlySet<string>,
): boolean {
  const normalized = normalizedFieldName(field);
  for (const prohibited of prohibitedFields) {
    if (
      normalized === prohibited ||
      normalized.startsWith(prohibited) ||
      normalized.endsWith(prohibited)
    ) {
      return true;
    }
  }
  return false;
}

function assertPlainObject(
  value: unknown,
  path: string,
): asserts value is Record<string, unknown> {
  if (
    value === null ||
    typeof value !== "object" ||
    Array.isArray(value) ||
    Object.getPrototypeOf(value) !== Object.prototype
  ) {
    throw new UnsupportedDocumentTriageValidationError(
      path,
      "a plain object is required",
    );
  }
}

function assertSafeTree(
  value: unknown,
  path: string,
  prohibitedFields: ReadonlySet<string>,
  ancestors = new Set<object>(),
): void {
  if (typeof value === "string") {
    if (CONTROL_CHARACTERS.test(value)) {
      throw new UnsupportedDocumentTriageValidationError(
        path,
        "control characters are not permitted",
      );
    }
    if (PROMPT_INJECTION.test(value)) {
      throw new UnsupportedDocumentTriageValidationError(
        path,
        "prompt-control text is not permitted",
      );
    }
    return;
  }
  if (value === null || typeof value !== "object") return;
  if (
    value instanceof ArrayBuffer ||
    ArrayBuffer.isView(value) ||
    (typeof Blob !== "undefined" && value instanceof Blob)
  ) {
    throw new UnsupportedDocumentTriageValidationError(
      path,
      "binary values are not permitted",
    );
  }
  if (ancestors.has(value)) {
    throw new UnsupportedDocumentTriageValidationError(
      path,
      "cyclic values are not permitted",
    );
  }
  if (
    !Array.isArray(value) &&
    Object.getPrototypeOf(value) !== Object.prototype
  ) {
    throw new UnsupportedDocumentTriageValidationError(
      path,
      "non-plain objects are not permitted",
    );
  }

  ancestors.add(value);
  try {
    if (Array.isArray(value)) {
      value.forEach((entry, index) =>
        assertSafeTree(entry, `${path}[${index}]`, prohibitedFields, ancestors),
      );
      return;
    }
    for (const [field, entry] of Object.entries(value)) {
      if (isProhibitedFieldName(field, prohibitedFields)) {
        throw new UnsupportedDocumentTriageValidationError(
          `${path}.${field}`,
          "prohibited field",
        );
      }
      assertSafeTree(entry, `${path}.${field}`, prohibitedFields, ancestors);
    }
  } finally {
    ancestors.delete(value);
  }
}

function assertExactFields(
  value: Record<string, unknown>,
  required: readonly string[],
  optional: readonly string[],
  path: string,
): void {
  const allowed = new Set([...required, ...optional]);
  for (const field of Object.keys(value)) {
    if (!allowed.has(field)) {
      throw new UnsupportedDocumentTriageValidationError(
        `${path}.${field}`,
        "unknown field",
      );
    }
  }
  for (const field of required) {
    if (!Object.prototype.hasOwnProperty.call(value, field)) {
      throw new UnsupportedDocumentTriageValidationError(
        `${path}.${field}`,
        "required field is missing",
      );
    }
  }
}

function assertAllowlistedArray<T extends string>(
  value: unknown,
  allowed: readonly T[],
  path: string,
  minimum: number,
  maximum: number,
): asserts value is T[] {
  if (
    !Array.isArray(value) ||
    value.length < minimum ||
    value.length > maximum
  ) {
    throw new UnsupportedDocumentTriageValidationError(
      path,
      `must contain between ${minimum} and ${maximum} entries`,
    );
  }
  const allowedValues = new Set<string>(allowed);
  const seen = new Set<string>();
  value.forEach((entry, index) => {
    if (typeof entry !== "string" || !allowedValues.has(entry)) {
      throw new UnsupportedDocumentTriageValidationError(
        `${path}[${index}]`,
        "value is not allowlisted",
      );
    }
    if (seen.has(entry)) {
      throw new UnsupportedDocumentTriageValidationError(
        `${path}[${index}]`,
        "duplicate entries are not permitted",
      );
    }
    seen.add(entry);
  });
}

function assertSchemaVersion(value: unknown, path: string): void {
  if (value !== UNSUPPORTED_DOCUMENT_TRIAGE_SCHEMA_VERSION) {
    throw new UnsupportedDocumentTriageValidationError(
      path,
      "unsupported schema version",
    );
  }
}

export function validateUnsupportedDocumentTriageRequest(
  value: unknown,
  policy: UnsupportedDocumentTriagePolicy = {},
): UnsupportedDocumentTriageRequest {
  assertSafeTree(value, "request", FORBIDDEN_FIELD_NAMES);
  assertPlainObject(value, "request");
  assertExactFields(
    value,
    REQUEST_REQUIRED_FIELDS,
    REQUEST_OPTIONAL_FIELDS,
    "request",
  );
  assertSchemaVersion(value.schemaVersion, "request.schemaVersion");
  assertAllowlistedArray(
    value.reasonCodes,
    UNSUPPORTED_DOCUMENT_REASON_CODES,
    "request.reasonCodes",
    1,
    MAX_REASON_CODES,
  );
  if (
    typeof value.supportLanguageCode !== "string" ||
    !SUPPORT_LANGUAGE_CODES.includes(
      value.supportLanguageCode as SupportLanguageCode,
    )
  ) {
    throw new UnsupportedDocumentTriageValidationError(
      "request.supportLanguageCode",
      "language code is not allowlisted",
    );
  }
  assertAllowlistedArray(
    value.redactedNotes,
    REDACTED_REASON_TOKENS,
    "request.redactedNotes",
    0,
    MAX_REDACTED_NOTES,
  );

  if (value.documentCategory !== undefined) {
    if (!policy.allowDocumentCategory) {
      throw new UnsupportedDocumentTriageValidationError(
        "request.documentCategory",
        "document category is not permitted by local policy",
      );
    }
    if (
      typeof value.documentCategory !== "string" ||
      !UNSUPPORTED_DOCUMENT_CATEGORIES.includes(
        value.documentCategory as UnsupportedDocumentCategory,
      )
    ) {
      throw new UnsupportedDocumentTriageValidationError(
        "request.documentCategory",
        "document category is not allowlisted",
      );
    }
  }
  if (value.countryCode !== undefined) {
    if (!policy.allowCountryCode) {
      throw new UnsupportedDocumentTriageValidationError(
        "request.countryCode",
        "country code is not permitted by local policy",
      );
    }
    if (
      typeof value.countryCode !== "string" ||
      !COUNTRY_CODE.test(value.countryCode)
    ) {
      throw new UnsupportedDocumentTriageValidationError(
        "request.countryCode",
        "country code must be two uppercase ASCII letters",
      );
    }
  }

  return Object.freeze({
    schemaVersion: UNSUPPORTED_DOCUMENT_TRIAGE_SCHEMA_VERSION,
    reasonCodes: Object.freeze([
      ...value.reasonCodes,
    ]) as UnsupportedDocumentReasonCode[],
    supportLanguageCode: value.supportLanguageCode as SupportLanguageCode,
    redactedNotes: Object.freeze([
      ...value.redactedNotes,
    ]) as RedactedReasonToken[],
    ...(value.documentCategory === undefined
      ? {}
      : {
          documentCategory:
            value.documentCategory as UnsupportedDocumentCategory,
        }),
    ...(value.countryCode === undefined
      ? {}
      : { countryCode: value.countryCode as string }),
  });
}

export function validateUnsupportedDocumentTriageOutput(
  value: unknown,
): UnsupportedDocumentTriageOutput {
  assertSafeTree(value, "output", PROHIBITED_OUTPUT_FIELDS);
  assertPlainObject(value, "output");
  assertExactFields(value, OUTPUT_FIELDS, [], "output");
  assertSchemaVersion(value.schemaVersion, "output.schemaVersion");
  if (
    typeof value.unsupportedDocumentCategory !== "string" ||
    !UNSUPPORTED_DOCUMENT_CATEGORIES.includes(
      value.unsupportedDocumentCategory as UnsupportedDocumentCategory,
    )
  ) {
    throw new UnsupportedDocumentTriageValidationError(
      "output.unsupportedDocumentCategory",
      "unsupported-document category is not allowlisted",
    );
  }
  assertAllowlistedArray(
    value.safeNextStepOptions,
    SAFE_NEXT_STEP_OPTIONS,
    "output.safeNextStepOptions",
    1,
    MAX_NEXT_STEPS,
  );
  if (typeof value.redactedCaseSummary !== "string") {
    throw new UnsupportedDocumentTriageValidationError(
      "output.redactedCaseSummary",
      "summary must be an allowlisted token string",
    );
  }
  const summaryTokens = value.redactedCaseSummary.split(" ");
  assertAllowlistedArray(
    summaryTokens,
    REDACTED_SUMMARY_TOKENS,
    "output.redactedCaseSummary",
    1,
    MAX_SUMMARY_TOKENS,
  );

  return Object.freeze({
    schemaVersion: UNSUPPORTED_DOCUMENT_TRIAGE_SCHEMA_VERSION,
    unsupportedDocumentCategory:
      value.unsupportedDocumentCategory as UnsupportedDocumentCategory,
    safeNextStepOptions: Object.freeze([
      ...value.safeNextStepOptions,
    ]) as SafeNextStepOption[],
    redactedCaseSummary: value.redactedCaseSummary,
  });
}

function canonicalDraftValue(
  request: UnsupportedDocumentTriageRequest,
  output: UnsupportedDocumentTriageOutput,
): string {
  return JSON.stringify({ request, output });
}

function validateDraft(value: unknown): UnsupportedDocumentTriageDraft {
  assertSafeTree(value, "draft", PROHIBITED_OUTPUT_FIELDS);
  assertPlainObject(value, "draft");
  assertExactFields(value, DRAFT_FIELDS, [], "draft");
  if (value.status !== "draft") {
    throw new UnsupportedDocumentTriageConfirmationError(
      "Only a draft can be confirmed",
    );
  }
  if (
    typeof value.draftReference !== "string" ||
    !/^sha256:[a-f0-9]{64}$/.test(value.draftReference)
  ) {
    throw new UnsupportedDocumentTriageConfirmationError(
      "Draft reference is invalid",
    );
  }
  const output = validateUnsupportedDocumentTriageOutput({
    schemaVersion: value.schemaVersion,
    unsupportedDocumentCategory: value.unsupportedDocumentCategory,
    safeNextStepOptions: value.safeNextStepOptions,
    redactedCaseSummary: value.redactedCaseSummary,
  });
  return { ...output, status: "draft", draftReference: value.draftReference };
}

export class UnsupportedDocumentTriageAssistant {
  private readonly issuedDrafts = new Map<string, string>();

  constructor(
    private readonly dependencies: UnsupportedDocumentTriageDependencies = {},
  ) {}

  isAvailable(): boolean {
    return Boolean(
      this.dependencies.modelAuthority &&
      this.dependencies.outputValidator &&
      this.dependencies.digestAuthority,
    );
  }

  async propose(value: unknown): Promise<UnsupportedDocumentTriageDraft> {
    const request = validateUnsupportedDocumentTriageRequest(
      value,
      this.dependencies.policy,
    );
    const { modelAuthority, outputValidator, digestAuthority } =
      this.dependencies;
    if (!modelAuthority || !outputValidator || !digestAuthority) {
      throw new UnsupportedDocumentTriageUnavailableError();
    }

    const output = validateUnsupportedDocumentTriageOutput(
      await modelAuthority.proposeUnsupportedDocumentTriage(request),
    );
    await outputValidator.validate(output);
    const canonicalValue = canonicalDraftValue(request, output);
    const digest = await digestAuthority.sha256(canonicalValue);
    if (!SHA256_DIGEST.test(digest)) {
      throw new UnsupportedDocumentTriageValidationError(
        "digest",
        "digest authority must return a lowercase SHA-256 hex digest",
      );
    }
    const draftReference = `sha256:${digest}`;
    const existingDraft = this.issuedDrafts.get(draftReference);
    if (existingDraft !== undefined && existingDraft !== canonicalValue) {
      throw new UnsupportedDocumentTriageValidationError(
        "digest",
        "digest collision is not permitted",
      );
    }
    this.issuedDrafts.set(draftReference, canonicalValue);
    return Object.freeze({
      ...output,
      status: "draft",
      draftReference,
    });
  }

  async confirm(
    originalRequest: unknown,
    draftValue: unknown,
    confirmationValue: unknown,
  ): Promise<AcceptedUnsupportedDocumentSupportNote> {
    const request = validateUnsupportedDocumentTriageRequest(
      originalRequest,
      this.dependencies.policy,
    );
    const draft = validateDraft(draftValue);
    assertPlainObject(confirmationValue, "confirmation");
    assertExactFields(
      confirmationValue,
      CONFIRMATION_FIELDS,
      [],
      "confirmation",
    );
    if (
      confirmationValue.confirmed !== true ||
      confirmationValue.draftReference !== draft.draftReference
    ) {
      throw new UnsupportedDocumentTriageConfirmationError(
        "Exact human confirmation of this draft is required",
      );
    }
    const { outputValidator, digestAuthority } = this.dependencies;
    if (!outputValidator || !digestAuthority) {
      throw new UnsupportedDocumentTriageUnavailableError();
    }
    const output = validateUnsupportedDocumentTriageOutput({
      schemaVersion: draft.schemaVersion,
      unsupportedDocumentCategory: draft.unsupportedDocumentCategory,
      safeNextStepOptions: draft.safeNextStepOptions,
      redactedCaseSummary: draft.redactedCaseSummary,
    });
    await outputValidator.validate(output);
    const canonicalValue = canonicalDraftValue(request, output);
    if (this.issuedDrafts.get(draft.draftReference) !== canonicalValue) {
      throw new UnsupportedDocumentTriageConfirmationError(
        "Draft was not issued for this request",
      );
    }
    const digest = await digestAuthority.sha256(canonicalValue);
    if (`sha256:${digest}` !== draft.draftReference) {
      throw new UnsupportedDocumentTriageConfirmationError(
        "Draft content does not match its audit reference",
      );
    }
    return Object.freeze({
      ...output,
      status: "accepted-support-note",
      draftReference: draft.draftReference,
      humanConfirmed: true,
    });
  }
}
