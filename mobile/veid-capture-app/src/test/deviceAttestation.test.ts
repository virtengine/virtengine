import { describe, expect, it } from "vitest";
import {
  createDeviceAttestation,
  MockDeviceAttestationProvider
} from "../core/deviceAttestation";
import { finalizeCaptureSession, initializeCaptureSession } from "../core/captureSession";

describe("createDeviceAttestation", () => {
  it("fails closed when no production provider is configured", () => {
    const attestation = createDeviceAttestation("1.0.0");

    expect(attestation.supported).toBe(false);
    expect(attestation.provider).toBe("unavailable");
    expect(attestation.integrityLevel).toBe("unsupported");
    expect(attestation.integrityScore).toBe(0);
    expect(attestation.failureReason).toBe("attestation_provider_unavailable");
    expect(attestation.attestationPayload).toBe("");
    expect(attestation.attestationSignature).toBe("");
  });

  it("returns a supported mock attestation only when explicitly injected", () => {
    const attestation = createDeviceAttestation(
      "1.0.0",
      "com.virtengine.veid",
      new MockDeviceAttestationProvider()
    );
    expect(attestation.supported).toBe(true);
    expect(attestation.provider).toBe("mock");
    expect(attestation.integrityLevel).toBe("strong");
    expect(attestation.nonce.length).toBeGreaterThan(0);
  });

  it("finalizes a production session as unsupported without an injected provider", () => {
    const session = finalizeCaptureSession(initializeCaptureSession("passport"), "1.0.0");

    expect(session.deviceAttestation?.supported).toBe(false);
    expect(session.deviceAttestation?.failureReason).toBe("attestation_provider_unavailable");
  });

  it("preserves explicit fixture attestation during session finalization", () => {
    const session = finalizeCaptureSession(
      initializeCaptureSession("passport"),
      "1.0.0",
      new MockDeviceAttestationProvider()
    );

    expect(session.deviceAttestation?.supported).toBe(true);
    expect(session.deviceAttestation?.provider).toBe("mock");
  });
});
