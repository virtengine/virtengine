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
  quote?: {
    estimatedTotal: string;
    depositRequired: string;
    pricePerHour: string;
    maxHours: number;
    denom: string;
  };
}

export function assertValidSubmitJobParams(
  params: SubmitJobParams,
  requireQuote = false,
): void {
  const resources = params.resources;
  const positiveInteger = (value: number) =>
    Number.isSafeInteger(value) && value > 0;
  if (
    !params.offeringId.trim() ||
    !params.name.trim() ||
    !positiveInteger(resources.nodes) ||
    !positiveInteger(resources.cpusPerNode) ||
    !positiveInteger(resources.memoryGBPerNode) ||
    !positiveInteger(resources.maxRuntimeSeconds) ||
    !Number.isSafeInteger(resources.storageGB) ||
    resources.storageGB < 0 ||
    (resources.gpusPerNode !== undefined &&
      (!Number.isSafeInteger(resources.gpusPerNode) ||
        resources.gpusPerNode < 0)) ||
    Boolean(resources.gpuType) !== Boolean(resources.gpusPerNode)
  ) {
    throw new HPCMutationNotCommittedError();
  }
  if (requireQuote) {
    const quote = params.quote;
    const amount = (value: string | undefined) =>
      typeof value === "string" && /^(0|[1-9]\d*)(\.\d{1,18})?$/.test(value);
    if (
      !quote ||
      !amount(quote.estimatedTotal) ||
      !amount(quote.depositRequired) ||
      !amount(quote.pricePerHour) ||
      !Number.isFinite(quote.maxHours) ||
      quote.maxHours !== resources.maxRuntimeSeconds / 3600 ||
      !/^[a-z][a-z0-9/._-]{1,127}$/.test(quote.denom)
    ) {
      throw new HPCMutationNotCommittedError();
    }
  }
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
