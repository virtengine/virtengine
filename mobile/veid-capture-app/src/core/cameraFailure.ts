import { VerificationTerminalError, type VerificationFailureCode } from "./verificationError";

const knownCameraFailureCodes = new Set<VerificationFailureCode>([
  "camera_permission_denied",
  "camera_unavailable",
  "camera_capture_failed"
]);

export function toTerminalCameraError(code: string, cause?: unknown): VerificationTerminalError {
  const terminalCode = knownCameraFailureCodes.has(code as VerificationFailureCode)
    ? (code as VerificationFailureCode)
    : "camera_unavailable";
  return new VerificationTerminalError(terminalCode, "Camera capture is required; verification cannot continue.", { cause });
}
