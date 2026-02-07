import { describe, expect, it } from "vitest";
import { captureBiometric } from "../core/biometric";

describe("captureBiometric", () => {
  it("captures a fingerprint with liveness and anti-spoofing", () => {
    const result = captureBiometric("fingerprint");
    expect(result.supported).toBe(true);
    expect(result.liveness.passed).toBe(true);
    expect(result.antiSpoofing.passed).toBe(true);
    expect(result.template.length).toBeGreaterThan(0);
  });
});
