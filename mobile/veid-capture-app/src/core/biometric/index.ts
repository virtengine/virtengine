import type { BiometricCapture, BiometricModality } from "../captureModels";
import type { BiometricProviderAdapter } from "./provider";

export function captureBiometric(
  modality: BiometricModality,
  provider?: BiometricProviderAdapter
): BiometricCapture {
  if (!provider) {
    return createUnsupportedCapture(modality, "biometric_provider_unavailable");
  }

  try {
    if (!provider.isSupported(modality)) {
      return createUnsupportedCapture(modality, "biometric_not_supported");
    }

    return provider.capture(modality);
  } catch {
    return createUnsupportedCapture(modality, "biometric_provider_error");
  }
}

function createUnsupportedCapture(
  modality: BiometricModality,
  failureReason: string
): BiometricCapture {
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
    failureReason
  };
}
