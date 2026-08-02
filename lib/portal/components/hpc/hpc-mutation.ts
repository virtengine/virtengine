export type HPCClientCapability = "query" | "signer" | "provider";

export class HPCClientUnavailableError extends Error {
  readonly code = "hpc_client_unavailable";

  constructor(readonly capability: HPCClientCapability) {
    super(`HPC ${capability} capability is unavailable`);
    this.name = "HPCClientUnavailableError";
  }
}

export class HPCMutationNotCommittedError extends Error {
  readonly code = "hpc_mutation_not_committed";

  constructor() {
    super(
      "HPC mutation did not return authoritative committed transaction and job state",
    );
    this.name = "HPCMutationNotCommittedError";
  }
}

export interface SubmitJobParams {
  offeringId: string;
  name: string;
  description?: string;
  templateId?: string;
  resources: {
    nodes: number;
    cpusPerNode: number;
    memoryGBPerNode: number;
    gpusPerNode?: number;
    gpuType?: string;
    maxRuntimeSeconds: number;
    storageGB: number;
  };
  command?: string;
  containerImage?: string;
  environment?: Record<string, string>;
  parameters?: Record<string, string | number | boolean>;
  encryptedInputs?: Record<string, unknown>;
  inputRefs?: string[];
}

export interface CommittedJobMutation {
  committed: true;
  jobId: string;
  txHash: string;
  code: 0;
  blockHeight: number;
}

export function assertCommittedJobMutation(
  result: unknown,
  expectedJobId?: string,
): asserts result is CommittedJobMutation {
  if (
    typeof result !== "object" ||
    result === null ||
    !("committed" in result) ||
    result.committed !== true ||
    !("jobId" in result) ||
    typeof result.jobId !== "string" ||
    result.jobId.trim().length === 0 ||
    (expectedJobId !== undefined && result.jobId !== expectedJobId) ||
    !("txHash" in result) ||
    typeof result.txHash !== "string" ||
    result.txHash.trim().length === 0 ||
    !("code" in result) ||
    result.code !== 0 ||
    !("blockHeight" in result) ||
    typeof result.blockHeight !== "number" ||
    !Number.isInteger(result.blockHeight) ||
    result.blockHeight <= 0
  ) {
    throw new HPCMutationNotCommittedError();
  }
}

export interface HPCSignerAdapter {
  readonly state: "query-only" | "signing-ready";
  readonly chainId: string;
  readonly accountAddress: string;
  submitJob(params: SubmitJobParams): Promise<unknown>;
  cancelJob(jobId: string): Promise<unknown>;
}

export function requireHPCSigner(
  adapter: HPCSignerAdapter | undefined,
  expected?: { chainId: string; accountAddress: string },
): HPCSignerAdapter {
  if (!adapter || adapter.state !== "signing-ready") {
    throw new HPCClientUnavailableError("signer");
  }
  if (
    expected &&
    (adapter.chainId !== expected.chainId ||
      adapter.accountAddress !== expected.accountAddress)
  ) {
    throw new HPCClientUnavailableError("signer");
  }
  return adapter;
}
