import { describe, expect, it } from "vitest";
import type { LivenessCheckResult } from "../types/capture";
import {
  requestLivenessEvidence,
  validateLivenessEvidence,
} from "../utils/liveness";

const context = {
  sessionId: "session-1",
  challengeId: "challenge-1",
  timeoutMs: 10_000,
};
const validEvidence: LivenessCheckResult = {
  passed: true,
  providerId: "approved-liveness",
  providerVersion: "1.0.0",
  challengeId: context.challengeId,
  sessionId: context.sessionId,
  score: 0.98,
  challengeType: "passive",
  challengeDurationMs: 450,
  evidenceDigest: "sha256:abc123",
};

describe("validateLivenessEvidence", () => {
  it("accepts valid evidence bound to the capture", () => {
    expect(validateLivenessEvidence(validEvidence, context)).toBe(
      validEvidence,
    );
  });

  it.each([
    undefined,
    { ...validEvidence, passed: false },
    { ...validEvidence, sessionId: "other-session" },
    { ...validEvidence, challengeId: "other-challenge" },
    { ...validEvidence, challengeDurationMs: context.timeoutMs + 1 },
    { ...validEvidence, evidenceDigest: "" },
  ])(
    "rejects missing, failed, mismatched, expired, or incomplete evidence",
    (evidence) => {
      expect(() =>
        validateLivenessEvidence(
          evidence as LivenessCheckResult | undefined,
          context,
        ),
      ).toThrow();
    },
  );
});

describe("requestLivenessEvidence", () => {
  it("rejects an unavailable provider", async () => {
    await expect(requestLivenessEvidence(undefined, context)).rejects.toThrow(
      "unavailable",
    );
  });

  it("rejects a provider timeout instead of succeeding", async () => {
    const provider = {
      verify: () => new Promise<LivenessCheckResult>(() => undefined),
    };
    await expect(
      requestLivenessEvidence(provider, { ...context, timeoutMs: 5 }),
    ).rejects.toThrow("timed out");
  });

  it("returns valid injected evidence once", async () => {
    let calls = 0;
    const provider = {
      verify: async () => {
        calls += 1;
        return validEvidence;
      },
    };
    await expect(requestLivenessEvidence(provider, context)).resolves.toBe(
      validEvidence,
    );
    expect(calls).toBe(1);
  });
});
