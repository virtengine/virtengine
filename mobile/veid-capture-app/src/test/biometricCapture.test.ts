import { describe, expect, it } from "vitest";
import { captureBiometric } from "../core/biometric";
import { MockBiometricProvider } from "../core/biometric/provider";

describe("captureBiometric", () => {
  it("fails closed when no production provider is configured", () => {
    const result = captureBiometric("fingerprint");

    expect(result.supported).toBe(false);
    expect(result.failureReason).toBe("biometric_provider_unavailable");
    expect(result.template).toBe("");
    expect(result.liveness.passed).toBe(false);
    expect(result.antiSpoofing.passed).toBe(false);
  });

  it("captures a fingerprint with an explicitly injected fixture provider", () => {
    const result = captureBiometric("fingerprint", new MockBiometricProvider());
    expect(result.supported).toBe(true);
    expect(result.liveness.passed).toBe(true);
    expect(result.antiSpoofing.passed).toBe(true);
    expect(result.template.length).toBeGreaterThan(0);
  });

  it("fails closed when an injected provider throws", () => {
    const result = captureBiometric("iris", {
      isSupported: () => {
        throw new Error("native bridge unavailable");
      },
      capture: () => {
        throw new Error("capture should not run");
      }
    });

    expect(result.supported).toBe(false);
    expect(result.failureReason).toBe("biometric_provider_error");
    expect(result.template).toBe("");
  });
});
