import type { CapturePayload, CaptureSession, DocumentType } from "./captureModels";
import { createDeviceAttestation, type DeviceAttestationProviderAdapter } from "./deviceAttestation";
import { encryptPayload } from "./encryption";
import { createId } from "../utils/id";
import { hashString } from "../utils/hash";

export function initializeCaptureSession(documentType: DocumentType): CaptureSession {
  return {
    sessionId: createId("session"),
    createdAt: Date.now(),
    documentType,
    socialMedia: []
  };
}

export function finalizeCaptureSession(
  session: CaptureSession,
  appVersion: string,
  attestationProvider?: DeviceAttestationProviderAdapter
): CaptureSession {
  return {
    ...session,
    deviceAttestation:
      session.deviceAttestation ??
      createDeviceAttestation(appVersion, "com.virtengine.veid", attestationProvider)
  };
}

export function buildCapturePayload(
  session: CaptureSession,
  uploadUrl: string,
  allowInsecureEncryption = false
): CapturePayload {
  const serialized = JSON.stringify(session);
  const encrypted = encryptPayload(serialized, { allowInsecure: allowInsecureEncryption });

  return {
    session,
    encryptedPayload: encrypted.ciphertext,
    payloadHash: hashString(encrypted.ciphertext),
    transport: {
      uploadUrl,
      retryCount: 0
    }
  };
}
