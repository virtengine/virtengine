import type { BiometricCapture, BiometricModality } from "../captureModels";
import { hashString } from "../../utils/hash";

export interface BiometricProviderAdapter {
  isSupported(modality: BiometricModality): boolean;
  capture(modality: BiometricModality): BiometricCapture;
}

export class MockBiometricProvider implements BiometricProviderAdapter {
  isSupported(_modality: BiometricModality): boolean {
    return true;
  }

  capture(modality: BiometricModality): BiometricCapture {
    const timestamp = Date.now();
    const templateSeed = `${modality}:${timestamp}`;
    const template = hashString(templateSeed);

    return {
      modality,
      templateFormat: modality === "fingerprint" ? "iso_19794_2" : "iso_19794_6",
      template,
      capturedAt: timestamp,
      liveness: {
        passed: true,
        score: 94,
        method: "combined",
        detectedSignals: ["pulse_detection", "texture_analysis"]
      },
      antiSpoofing: {
        passed: true,
        score: 91,
        signals: ["anti_replay", "surface_reflection"]
      },
      deviceInfo: {
        manufacturer: "Mock Biometrics",
        model: modality === "fingerprint" ? "Fingerprint X1" : "Iris Vision Pro",
        sensorType: modality === "fingerprint" ? "ultrasonic" : "iris",
        securityLevel: "hardware_backed",
        firmwareVersion: "2.4.1"
      },
      supported: true
    };
  }
}
