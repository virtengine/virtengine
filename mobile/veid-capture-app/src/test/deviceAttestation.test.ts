import { describe, expect, it } from "vitest";
import { createDeviceAttestation } from "../core/deviceAttestation";

describe("createDeviceAttestation", () => {
  it("returns a supported mock attestation", () => {
    const attestation = createDeviceAttestation("1.0.0");
    expect(attestation.supported).toBe(true);
    expect(attestation.provider).toBe("mock");
    expect(attestation.integrityLevel).toBe("strong");
    expect(attestation.nonce.length).toBeGreaterThan(0);
  });
});
