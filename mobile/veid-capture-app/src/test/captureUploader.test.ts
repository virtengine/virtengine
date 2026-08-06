import { describe, expect, it, vi } from "vitest";
import type { CapturePayload } from "../core/captureModels";
import { createCaptureUploadAttempt } from "../services/upload/captureUploadAttempt";
import {
  uploadCapture,
  type AuthenticatedUploadReceipt,
  type CaptureUploadRequest,
  type CaptureUploadTransport,
  type EvidenceEnvelopePayload
} from "../services/upload/captureUploader";

const now = 1_800_000_000_000;
const capture = {
  session: { sessionId: "session-1" },
  encryptedPayload: "ciphertext",
  payloadHash: "local-noncanonical-hash"
} as CapturePayload;
const evidence: EvidenceEnvelopePayload = {
  envelope: { opaque: true },
  payloadDigest: "sha256:canonical-payload",
  envelopeDigest: "sha256:canonical-envelope"
};

function receipt(overrides: Partial<AuthenticatedUploadReceipt> = {}): AuthenticatedUploadReceipt {
  return {
    authenticated: true,
    receiptId: "receipt-1",
    payloadDigest: evidence.payloadDigest,
    envelopeDigest: evidence.envelopeDigest,
    idempotencyKey: "upload-1",
    issuedAt: now - 1_000,
    expiresAt: now + 60_000,
    ...overrides
  };
}

function transport(
  implementation: (request: CaptureUploadRequest) => Promise<AuthenticatedUploadReceipt>
): CaptureUploadTransport {
  return { upload: vi.fn(implementation) };
}

describe("uploadCapture", () => {
  it("fails typed unavailable without the T1 envelope or injected transport", async () => {
    await expect(
      uploadCapture({ capture, idempotencyKey: "upload-1", now: () => now })
    ).resolves.toEqual({ success: false, error: "evidence_envelope_unavailable" });
    await expect(
      uploadCapture({ capture, evidence, idempotencyKey: "upload-1", now: () => now })
    ).resolves.toEqual({ success: false, error: "upload_transport_unavailable" });
  });

  it("reports interruption without claiming success", async () => {
    const interrupted = transport(async () => {
      throw new Error("connection reset");
    });

    await expect(
      uploadCapture({ capture, evidence, idempotencyKey: "upload-1", transport: interrupted })
    ).resolves.toEqual({ success: false, error: "upload_interrupted" });
  });

  it("reuses the caller idempotency key across a retry", async () => {
    const requests: CaptureUploadRequest[] = [];
    const retrying = transport(async (request) => {
      requests.push(request);
      if (requests.length === 1) throw new Error("interrupted");
      return receipt();
    });
    const options = {
      capture,
      evidence,
      idempotencyKey: "upload-1",
      transport: retrying,
      now: () => now
    };

    expect(await uploadCapture(options)).toEqual({ success: false, error: "upload_interrupted" });
    expect(await uploadCapture(options)).toEqual({ success: true, receipt: receipt() });
    expect(requests.map((request) => request.idempotencyKey)).toEqual(["upload-1", "upload-1"]);
  });

  it("keeps the caller payload and idempotency key stable after interruption", async () => {
    const requests: CaptureUploadRequest[] = [];
    const retrying = transport(async (request) => {
      requests.push(request);
      if (requests.length === 1) throw new Error("interrupted");
      return receipt();
    });
    const createIdempotencyKey = vi.fn(() => "upload-1");
    const attempt = createCaptureUploadAttempt(capture, {
      evidence,
      transport: retrying,
      createIdempotencyKey,
      now: () => now
    });

    await expect(attempt.upload()).resolves.toEqual({
      success: false,
      error: "upload_interrupted"
    });
    await expect(attempt.upload()).resolves.toEqual({ success: true, receipt: receipt() });
    expect(createIdempotencyKey).toHaveBeenCalledTimes(1);
    expect(requests.map((request) => request.capture)).toEqual([capture, capture]);
    expect(requests.map((request) => request.idempotencyKey)).toEqual(["upload-1", "upload-1"]);
  });

  it("rejects replacing the payload within an upload attempt", async () => {
    const uploaded = transport(async () => receipt());
    const attempt = createCaptureUploadAttempt(capture, {
      evidence,
      transport: uploaded,
      createIdempotencyKey: () => "upload-1",
      now: () => now
    });

    await expect(
      attempt.upload({ ...capture, encryptedPayload: "changed-ciphertext" })
    ).resolves.toEqual({ success: false, error: "invalid_upload_request" });
    expect(uploaded.upload).not.toHaveBeenCalled();
  });

  it("rejects mutating the payload within an upload attempt", async () => {
    const mutableCapture = { ...capture };
    const uploaded = transport(async () => receipt());
    const attempt = createCaptureUploadAttempt(mutableCapture, {
      evidence,
      transport: uploaded,
      createIdempotencyKey: () => "upload-1",
      now: () => now
    });

    mutableCapture.encryptedPayload = "changed-ciphertext";

    await expect(attempt.upload()).resolves.toEqual({
      success: false,
      error: "invalid_upload_request"
    });
    expect(uploaded.upload).not.toHaveBeenCalled();
  });

  it("destructively wipes payload and old session references", () => {
    const sensitiveCapture = {
      ...capture,
      transport: { uploadUrl: "https://upload.example", retryCount: 0 },
      session: {
        sessionId: "session-1",
        documentFront: { image: { uri: "file:///document.jpg" } },
        ocr: { rawText: "secret", fields: [{ value: "private" }] },
        biometric: {
          template: "embedding",
          liveness: { detectedSignals: [] },
          antiSpoofing: { signals: [] }
        }
      }
    } as unknown as CapturePayload;
    const oldSession = sensitiveCapture.session;
    const attempt = createCaptureUploadAttempt(sensitiveCapture, {
      createIdempotencyKey: () => "upload-1"
    });

    attempt.wipe();

    expect(sensitiveCapture.encryptedPayload).toBe("");
    expect(oldSession.documentFront).toBeUndefined();
    expect(oldSession.ocr).toBeUndefined();
    expect(oldSession.biometric).toBeUndefined();
  });

  it.each([
    { authenticated: false },
    { receiptId: "" },
    { payloadDigest: "sha256:changed" },
    { envelopeDigest: "sha256:wrong-envelope" },
    { idempotencyKey: "different-upload" },
    { issuedAt: now + 5 * 60 * 1000 + 1 },
    { expiresAt: now },
    { expiresAt: now + 16 * 60 * 1000 }
  ])("rejects a bad authenticated receipt %#", async (overrides) => {
    const invalid = transport(async () => receipt(overrides));

    await expect(
      uploadCapture({
        capture,
        evidence,
        idempotencyKey: "upload-1",
        transport: invalid,
        now: () => now
      })
    ).resolves.toEqual({ success: false, error: "invalid_upload_receipt" });
  });

  it("rejects a malformed receipt instead of throwing", async () => {
    const malformed = transport(async () => null as unknown as AuthenticatedUploadReceipt);

    await expect(
      uploadCapture({
        capture,
        evidence,
        idempotencyKey: "upload-1",
        transport: malformed,
        now: () => now
      })
    ).resolves.toEqual({ success: false, error: "invalid_upload_receipt" });
  });

  it("rejects a receipt for a changed canonical payload on retry", async () => {
    const staleReceipt = transport(async () => receipt());

    await expect(
      uploadCapture({
        capture: { ...capture, encryptedPayload: "changed-ciphertext" },
        evidence: { ...evidence, payloadDigest: "sha256:new-payload" },
        idempotencyKey: "upload-1",
        transport: staleReceipt,
        now: () => now
      })
    ).resolves.toEqual({ success: false, error: "invalid_upload_receipt" });
  });
});