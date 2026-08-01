import type { CapturePayload } from "../../core/captureModels";

export type CaptureUploadErrorCode =
  | "evidence_envelope_unavailable"
  | "upload_transport_unavailable"
  | "invalid_upload_request"
  | "upload_interrupted"
  | "invalid_upload_receipt";

export interface EvidenceEnvelopePayload {
  /** Canonical T1 envelope payload. T2 deliberately treats it as opaque. */
  envelope: unknown;
  payloadDigest: string;
  envelopeDigest: string;
}

export interface CaptureUploadRequest extends EvidenceEnvelopePayload {
  capture: CapturePayload;
  idempotencyKey: string;
}

export interface AuthenticatedUploadReceipt {
  authenticated: boolean;
  receiptId: string;
  payloadDigest: string;
  envelopeDigest: string;
  idempotencyKey: string;
  issuedAt: number;
  expiresAt: number;
}

export interface CaptureUploadTransport {
  upload(request: CaptureUploadRequest): Promise<AuthenticatedUploadReceipt>;
}

export type UploadResult =
  | { success: true; receipt: AuthenticatedUploadReceipt }
  | { success: false; error: CaptureUploadErrorCode };

export interface UploadCaptureOptions {
  capture: CapturePayload;
  evidence?: EvidenceEnvelopePayload;
  idempotencyKey: string;
  transport?: CaptureUploadTransport;
  now?: () => number;
}

const MAX_RECEIPT_CLOCK_SKEW_MS = 5 * 60 * 1000;
const MAX_RECEIPT_LIFETIME_MS = 15 * 60 * 1000;

function isAuthenticatedUploadReceipt(value: unknown): value is AuthenticatedUploadReceipt {
  if (typeof value !== "object" || value === null) return false;

  const receipt = value as Record<string, unknown>;
  return (
    receipt.authenticated === true &&
    typeof receipt.receiptId === "string" &&
    typeof receipt.payloadDigest === "string" &&
    typeof receipt.envelopeDigest === "string" &&
    typeof receipt.idempotencyKey === "string" &&
    typeof receipt.issuedAt === "number" &&
    typeof receipt.expiresAt === "number"
  );
}

export async function uploadCapture(options: UploadCaptureOptions): Promise<UploadResult> {
  const { capture, evidence, idempotencyKey, transport, now = Date.now } = options;
  if (!evidence) {
    return { success: false, error: "evidence_envelope_unavailable" };
  }
  if (!transport) {
    return { success: false, error: "upload_transport_unavailable" };
  }
  if (
    evidence.envelope == null ||
    !evidence.payloadDigest.trim() ||
    !evidence.envelopeDigest.trim() ||
    !idempotencyKey.trim()
  ) {
    return { success: false, error: "invalid_upload_request" };
  }

  let receipt: AuthenticatedUploadReceipt;
  try {
    receipt = await transport.upload({ capture, ...evidence, idempotencyKey });
  } catch {
    return { success: false, error: "upload_interrupted" };
  }

  const currentTime = now();
  const validReceipt =
    isAuthenticatedUploadReceipt(receipt) &&
    receipt.receiptId.trim().length > 0 &&
    receipt.payloadDigest === evidence.payloadDigest &&
    receipt.envelopeDigest === evidence.envelopeDigest &&
    receipt.idempotencyKey === idempotencyKey &&
    Number.isFinite(receipt.issuedAt) &&
    Number.isFinite(receipt.expiresAt) &&
    receipt.issuedAt <= currentTime + MAX_RECEIPT_CLOCK_SKEW_MS &&
    receipt.expiresAt > currentTime &&
    receipt.expiresAt > receipt.issuedAt &&
    receipt.expiresAt - receipt.issuedAt <= MAX_RECEIPT_LIFETIME_MS;

  return validReceipt
    ? { success: true, receipt }
    : { success: false, error: "invalid_upload_receipt" };
}
