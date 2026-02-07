import type { BiometricCapture, BiometricModality } from "../captureModels";
import { MockBiometricProvider, type BiometricProviderAdapter } from "./provider";

export function captureBiometric(
  modality: BiometricModality,
  provider: BiometricProviderAdapter = new MockBiometricProvider()
): BiometricCapture {
  if (!provider.isSupported(modality)) {
    return {
      modality,
      templateFormat: "unknown",
      template: "",
      capturedAt: Date.now(),
      liveness: {
        passed: false,
        score: 0,
        method: "software",
        detectedSignals: []
      },
      antiSpoofing: {
        passed: false,
        score: 0,
        signals: []
      },
      deviceInfo: {
        manufacturer: "unknown",
        model: "unknown",
        sensorType: "unknown",
        securityLevel: "unknown",
        firmwareVersion: "unknown"
      },
      supported: false,
      failureReason: "biometric_not_supported"
    };
  }

  return provider.capture(modality);
}
