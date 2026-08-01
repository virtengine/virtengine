import type { CapturePayload, CaptureSession } from "../../core/captureModels";

function wipeImageUri(capture: { image: { uri: string } } | undefined): void {
  if (capture) capture.image.uri = "";
}

export function wipeCaptureSession(session: CaptureSession): void {
  wipeImageUri(session.documentFront);
  wipeImageUri(session.documentBack);
  wipeImageUri(session.selfie);

  if (session.biometric) {
    session.biometric.template = "";
    session.biometric.liveness.detectedSignals.length = 0;
    session.biometric.antiSpoofing.signals.length = 0;
  }
  if (session.liveness) {
    for (const challenge of session.liveness.challenges) challenge.notes = undefined;
    session.liveness.challenges.length = 0;
    session.liveness.failureReason = undefined;
  }
  if (session.ocr) {
    session.ocr.rawText = "";
    for (const field of session.ocr.fields) field.value = "";
    session.ocr.fields.length = 0;
  }
  if (session.socialMedia) {
    for (const profile of session.socialMedia) {
      profile.profileNameHash = "";
      profile.emailHash = undefined;
      profile.usernameHash = undefined;
      profile.orgHash = undefined;
    }
    session.socialMedia.length = 0;
  }
  if (session.deviceAttestation) {
    session.deviceAttestation.attestationPayload = "";
    session.deviceAttestation.attestationSignature = "";
    session.deviceAttestation.nonce = "";
    session.deviceAttestation.verdicts = {};
  }

  session.documentFront = undefined;
  session.documentBack = undefined;
  session.selfie = undefined;
  session.liveness = undefined;
  session.biometric = undefined;
  session.ocr = undefined;
  session.deviceAttestation = undefined;
  session.socialMedia = [];
}

export function wipeCapturePayload(payload: CapturePayload): void {
  wipeCaptureSession(payload.session);
  payload.encryptedPayload = "";
  payload.payloadHash = "";
  payload.transport.uploadUrl = "";
  payload.transport.retryCount = 0;
}