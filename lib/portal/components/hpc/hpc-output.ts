import type { JobOutput, JobOutputReference } from "../../types/hpc";
import { HPCClientUnavailableError } from "./hpc-mutation";

export interface HPCOutputAdapter {
  readonly chainId: string;
  readonly accountAddress: string;
  getOutputs(jobId: string): Promise<unknown>;
  resolveOutput(outputRef: JobOutputReference): Promise<unknown>;
}

export interface HPCOutputReferenceEnvelope {
  chainId: string;
  accountAddress: string;
  jobId: string;
  outputs: unknown;
}

export interface HPCResolvedOutputEnvelope {
  chainId: string;
  accountAddress: string;
  jobId: string;
  output: unknown;
}

export class HPCOutputValidationError extends Error {
  readonly code = "hpc_output_invalid";

  constructor() {
    super("HPC output did not return authoritative bound access evidence");
    this.name = "HPCOutputValidationError";
  }
}

export function requireHPCOutputAdapter(
  adapter: HPCOutputAdapter | undefined,
  expected: { chainId: string; accountAddress: string },
): HPCOutputAdapter {
  if (
    !adapter ||
    adapter.chainId !== expected.chainId ||
    adapter.accountAddress !== expected.accountAddress
  ) {
    throw new HPCClientUnavailableError("provider");
  }
  return adapter;
}

function materializeReference(value: unknown): JobOutputReference {
  if (!value || typeof value !== "object") throw new HPCOutputValidationError();
  const source = value as Partial<JobOutputReference>;
  const reference = Object.freeze({
    id: source.id as string,
    name: source.name as string,
    type: source.type as JobOutputReference["type"],
    sizeBytes: source.sizeBytes as number,
    createdAt: source.createdAt as number,
    encryptedRef: source.encryptedRef as string,
    contentHash: source.contentHash as string,
    expiresAt: source.expiresAt,
  });
  if (
    typeof reference.id !== "string" ||
    !reference.id.trim() ||
    typeof reference.name !== "string" ||
    !reference.name.trim() ||
    !["model", "checkpoint", "logs", "metrics", "artifact", "data"].includes(
      reference.type,
    ) ||
    !Number.isInteger(reference.sizeBytes) ||
    reference.sizeBytes < 0 ||
    !Number.isInteger(reference.createdAt) ||
    reference.createdAt <= 0 ||
    typeof reference.encryptedRef !== "string" ||
    !reference.encryptedRef.trim() ||
    typeof reference.contentHash !== "string" ||
    !reference.contentHash.trim() ||
    (reference.expiresAt !== undefined &&
      (!Number.isInteger(reference.expiresAt) ||
        reference.expiresAt <= reference.createdAt))
  ) {
    throw new HPCOutputValidationError();
  }
  return reference;
}

export function validateHPCOutputReferences(
  value: unknown,
  expected: { chainId: string; accountAddress: string; jobId: string },
): readonly JobOutputReference[] {
  if (!value || typeof value !== "object") throw new HPCOutputValidationError();
  const envelope = value as Partial<HPCOutputReferenceEnvelope>;
  if (
    envelope.chainId !== expected.chainId ||
    envelope.accountAddress !== expected.accountAddress ||
    envelope.jobId !== expected.jobId ||
    !Array.isArray(envelope.outputs)
  ) {
    throw new HPCOutputValidationError();
  }
  const references = envelope.outputs.map(materializeReference);
  if (
    new Set(references.map((reference) => reference.id)).size !==
    references.length
  ) {
    throw new HPCOutputValidationError();
  }
  return Object.freeze(references);
}

export function validateResolvedHPCOutput(
  value: unknown,
  expected: JobOutputReference,
  binding: { chainId: string; accountAddress: string; jobId: string },
  now = Date.now(),
): JobOutput {
  if (!value || typeof value !== "object") throw new HPCOutputValidationError();
  const envelope = value as Partial<HPCResolvedOutputEnvelope>;
  if (
    envelope.chainId !== binding.chainId ||
    envelope.accountAddress !== binding.accountAddress ||
    envelope.jobId !== binding.jobId ||
    !envelope.output ||
    typeof envelope.output !== "object"
  ) {
    throw new HPCOutputValidationError();
  }
  const source = envelope.output as Partial<JobOutput>;
  const output = Object.freeze({
    refId: source.refId as string,
    name: source.name as string,
    type: source.type as JobOutput["type"],
    accessUrl: source.accessUrl as string,
    urlExpiresAt: source.urlExpiresAt as number,
    sizeBytes: source.sizeBytes as number,
    mimeType: source.mimeType as string,
  });
  let url: URL;
  try {
    url = new URL(output.accessUrl);
  } catch {
    throw new HPCOutputValidationError();
  }
  if (
    output.refId !== expected.id ||
    output.name !== expected.name ||
    output.type !== expected.type ||
    output.sizeBytes !== expected.sizeBytes ||
    typeof output.accessUrl !== "string" ||
    url.protocol !== "https:" ||
    !url.hostname ||
    !Number.isInteger(output.urlExpiresAt) ||
    output.urlExpiresAt <= now ||
    (expected.expiresAt !== undefined &&
      (expected.expiresAt <= now ||
        output.urlExpiresAt > expected.expiresAt)) ||
    typeof output.mimeType !== "string" ||
    !output.mimeType.trim()
  ) {
    throw new HPCOutputValidationError();
  }
  return output;
}
