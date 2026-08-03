import type {
  DeviceAttestation,
  DeviceAttestationProvider,
  DeviceIntegrityLevel,
  DevicePlatform
} from "./captureModels";
import { createId } from "../utils/id";

export interface DeviceAttestationProviderAdapter {
  getPlatform(): DevicePlatform;
  getProvider(): DeviceAttestationProvider;
  supportsAttestation(): boolean;
  attest(request: DeviceAttestationRequest): DeviceAttestationResponse;
}

export interface DeviceAttestationRequest {
  appId: string;
  appVersion: string;
  nonce: string;
}

export interface DeviceAttestationResponse {
  supported: boolean;
  integrityLevel: DeviceIntegrityLevel;
  integrityScore: number;
  deviceModel: string;
  osVersion: string;
  attestationPayload: string;
  attestationSignature: string;
  verdicts: Record<string, boolean>;
  failureReason?: string;
}

export class MockDeviceAttestationProvider implements DeviceAttestationProviderAdapter {
  getPlatform(): DevicePlatform {
    return "android";
  }

  getProvider(): DeviceAttestationProvider {
    return "mock";
  }

  supportsAttestation(): boolean {
    return true;
  }

  attest(request: DeviceAttestationRequest): DeviceAttestationResponse {
    return {
      supported: true,
      integrityLevel: "strong",
      integrityScore: 92,
      deviceModel: "Mock Device",
      osVersion: "Android 16",
      attestationPayload: `mock:${request.appId}:${request.nonce}`,
      attestationSignature: "mock_signature",
      verdicts: {
        basic_integrity: true,
        strong_integrity: true,
        hardware_backed: true
      }
    };
  }
}

export function createDeviceAttestation(
  appVersion: string,
  appId = "com.virtengine.veid",
  provider?: DeviceAttestationProviderAdapter
): DeviceAttestation {
  const nonce = createId("nonce");

  if (!provider) {
    return createUnavailableAttestation(appVersion, appId, nonce, "attestation_provider_unavailable");
  }

  let platform: DevicePlatform = "unknown";
  let attestationProvider: DeviceAttestationProvider = "unavailable";
  try {
    platform = provider.getPlatform();
    attestationProvider = provider.getProvider();
    if (!provider.supportsAttestation()) {
      return createUnavailableAttestation(
        appVersion,
        appId,
        nonce,
        "attestation_not_supported",
        platform,
        attestationProvider
      );
    }

    const response = provider.attest({ appId, appVersion, nonce });
    if (!response.supported) {
      return createUnavailableAttestation(
        appVersion,
        appId,
        nonce,
        response.failureReason ?? "unsupported_device",
        platform,
        attestationProvider
      );
    }

    return {
      deviceId: createId("device"),
      deviceModel: response.deviceModel,
      osVersion: response.osVersion,
      appVersion,
      appId,
      platform,
      provider: attestationProvider,
      integrityLevel: response.integrityLevel,
      integrityScore: response.integrityScore,
      supported: true,
      nonce,
      verdicts: response.verdicts,
      attestationPayload: response.attestationPayload,
      attestedAt: Date.now(),
      attestationSignature: response.attestationSignature
    };
  } catch {
    return createUnavailableAttestation(
      appVersion,
      appId,
      nonce,
      "attestation_provider_error",
      platform,
      attestationProvider
    );
  }
}

function createUnavailableAttestation(
  appVersion: string,
  appId: string,
  nonce: string,
  failureReason: string,
  platform: DevicePlatform = "unknown",
  provider: DeviceAttestationProvider = "unavailable"
): DeviceAttestation {
  return {
    deviceId: createId("device"),
    deviceModel: "unknown",
    osVersion: "unknown",
    appVersion,
    appId,
    platform,
    provider,
    integrityLevel: "unsupported",
    integrityScore: 0,
    supported: false,
    failureReason,
    nonce,
    verdicts: {},
    attestationPayload: "",
    attestedAt: Date.now(),
    attestationSignature: ""
  };
}
