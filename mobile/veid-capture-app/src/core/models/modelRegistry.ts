export type VerificationModelCapability = "ocr" | "mrz" | "pdf417" | "face_embedding" | "passive_pad" | "active_pad" | "deepfake";
export type ModelRuntime = "onnx" | "tflite" | "coreml" | "native";
export type DeploymentEnvironment = "test" | "development" | "production";

export interface ModelReleaseEvidence {
  evaluationId: string;
  signedReportSha256: string;
  releasedAt: string;
  metrics: Partial<Record<"far" | "frr" | "apcer" | "bpcer", number>>;
}

export interface VerificationModelManifest {
  id: string;
  version: string;
  capability: VerificationModelCapability;
  runtime: ModelRuntime;
  artifactSha256: string;
  signature: string;
  license: string;
  source: string;
  deterministic: boolean;
  randomWeights: boolean;
  releaseEvidence?: ModelReleaseEvidence;
}

const SHA256 = /^[a-f0-9]{64}$/i;

export class ModelReleaseGateError extends Error {
  constructor(readonly code: "invalid_manifest" | "unsigned_model" | "random_weights" | "missing_evidence", message: string) {
    super(message);
    this.name = "ModelReleaseGateError";
  }
}

/** Validates provenance; it never claims that a vendor model or metric is certified. */
export function validateModelManifest(manifest: VerificationModelManifest, environment: DeploymentEnvironment): void {
  if (!manifest.id || !manifest.version || !manifest.license || !manifest.source || !SHA256.test(manifest.artifactSha256)) {
    throw new ModelReleaseGateError("invalid_manifest", "Model manifest requires identity, license, source, and an artifact SHA-256.");
  }
  if (!manifest.signature.trim()) {
    throw new ModelReleaseGateError("unsigned_model", "Unsigned model artifacts cannot be used.");
  }
  if (environment === "production" && (manifest.randomWeights || !manifest.deterministic)) {
    throw new ModelReleaseGateError("random_weights", "Production verification cannot use random-weight or non-deterministic models.");
  }
  if (environment === "production" && !manifest.releaseEvidence) {
    throw new ModelReleaseGateError("missing_evidence", "Production models require signed evaluation and release evidence.");
  }
  if (manifest.releaseEvidence && !SHA256.test(manifest.releaseEvidence.signedReportSha256)) {
    throw new ModelReleaseGateError("invalid_manifest", "Model evaluation evidence must include a report SHA-256.");
  }
}

export function createModelRegistry(manifests: VerificationModelManifest[], environment: DeploymentEnvironment): ReadonlyMap<VerificationModelCapability, VerificationModelManifest> {
  const registry = new Map<VerificationModelCapability, VerificationModelManifest>();
  for (const manifest of manifests) {
    validateModelManifest(manifest, environment);
    if (registry.has(manifest.capability)) {
      throw new ModelReleaseGateError("invalid_manifest", `Multiple models registered for ${manifest.capability}.`);
    }
    registry.set(manifest.capability, Object.freeze({ ...manifest }));
  }
  return registry;
}
