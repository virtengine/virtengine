import { describe, expect, it } from "vitest";
import { requireProductionAttestation } from "../core/deviceAttestation";
import {
  createModelRegistry,
  ModelReleaseGateError,
  validateModelManifest,
  type VerificationModelManifest
} from "../core/models/modelRegistry";
import { VerificationTerminalError } from "../core/verificationError";
import { extractOcr } from "../services/ocr/ocrService";

const SHA = "a".repeat(64);
const manifest: VerificationModelManifest = {
  id: "ocr-vetted", version: "1.0.0", capability: "ocr", runtime: "onnx",
  artifactSha256: SHA, signature: "sigstore:bundle", license: "Apache-2.0", source: "approved-source",
  deterministic: true, randomWeights: false,
  releaseEvidence: { evaluationId: "eval-1", signedReportSha256: SHA, releasedAt: "2026-08-05T00:00:00Z", metrics: { far: 0.001 } }
};

describe("verification safety gates", () => {
  it("does not silently turn unavailable OCR into an empty success", async () => {
    await expect(extractOcr("file://capture.jpg")).rejects.toBeInstanceOf(VerificationTerminalError);
  });

  it("rejects random-weight models in production", () => {
    expect(() => validateModelManifest({ ...manifest, randomWeights: true }, "production"))
      .toThrow(ModelReleaseGateError);
  });

  it("requires signed evaluation evidence for production registry entries", () => {
    expect(() => createModelRegistry([{ ...manifest, releaseEvidence: undefined }], "production"))
      .toThrow(/release evidence/);
  });

  it("requires a real attestation, not a mock or unsupported placeholder", () => {
    expect(() => requireProductionAttestation(undefined)).toThrow(VerificationTerminalError);
    expect(() => requireProductionAttestation({ supported: true, provider: "mock", attestationPayload: "x", attestationSignature: "x" } as any))
      .toThrow(VerificationTerminalError);
  });
});
