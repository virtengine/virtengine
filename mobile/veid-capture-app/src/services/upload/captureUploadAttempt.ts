import type { CapturePayload } from "../../core/captureModels";
import {
  uploadCapture,
  type CaptureUploadTransport,
  type EvidenceEnvelopePayload,
  type UploadResult
} from "./captureUploader";

export interface CaptureUploadDependencies {
  evidence?: EvidenceEnvelopePayload;
  transport?: CaptureUploadTransport;
  createIdempotencyKey(): string;
  now?: () => number;
}

export interface CaptureUploadAttempt {
  readonly capture: CapturePayload;
  readonly idempotencyKey: string;
  upload(capture?: CapturePayload): Promise<UploadResult>;
}

export function createCaptureUploadAttempt(
  capture: CapturePayload,
  dependencies: CaptureUploadDependencies
): CaptureUploadAttempt {
  const idempotencyKey = dependencies.createIdempotencyKey();
  const captureSnapshot = JSON.stringify(capture);
  const options = {
    capture,
    evidence: dependencies.evidence ? { ...dependencies.evidence } : undefined,
    idempotencyKey,
    transport: dependencies.transport,
    now: dependencies.now
  };

  return {
    capture,
    idempotencyKey,
    upload(candidate = capture) {
      if (JSON.stringify(candidate) !== captureSnapshot) {
        return Promise.resolve({ success: false, error: "invalid_upload_request" });
      }
      return uploadCapture(options);
    }
  };
}