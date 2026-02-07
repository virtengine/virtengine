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
  provider: DeviceAttestationProviderAdapter = new MockDeviceAttestationProvider()
): DeviceAttestation {
  const nonce = createId("nonce");
  const response = provider.attest({ appId, appVersion, nonce });
  const supported = provider.supportsAttestation() && response.supported;

  return {
    deviceId: createId("device"),
    deviceModel: response.deviceModel,
    osVersion: response.osVersion,
    appVersion,
    appId,
    platform: provider.getPlatform(),
    provider: provider.getProvider(),
    integrityLevel: supported ? response.integrityLevel : "unsupported",
    integrityScore: supported ? response.integrityScore : 50,
    supported,
    failureReason: supported ? undefined : response.failureReason ?? "unsupported_device",
    nonce,
    verdicts: response.verdicts,
    attestationPayload: response.attestationPayload,
    attestedAt: Date.now(),
    attestationSignature: response.attestationSignature
  };
}
