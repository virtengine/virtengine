export type VerificationFailureCode =
  | "ocr_module_unavailable"
  | "ocr_recognition_failed"
  | "ocr_empty_result"
  | "face_detector_unavailable"
  | "face_detection_failed"
  | "liveness_detector_unavailable"
  | "attestation_required"
  | "attestation_unavailable";

/** A terminal failure: callers must not replace it with an inferred success. */
export class VerificationTerminalError extends Error {
  readonly code: VerificationFailureCode;

  constructor(code: VerificationFailureCode, message: string, options?: { cause?: unknown }) {
    super(message);
    this.name = "VerificationTerminalError";
    this.code = code;
    if (options?.cause !== undefined) {
      (this as Error & { cause?: unknown }).cause = options.cause;
    }
  }
}
