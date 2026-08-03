import type { ChainEvent } from "../../types/chain";
import type {
  Job,
  JobEvent,
  JobOutputReference,
  JobParameter,
  JobPriceQuote,
  JobResources,
  JobStatus,
  JobStatusChange,
  WorkloadTemplate,
} from "../../types/hpc";
import { HPCClientUnavailableError } from "./hpc-mutation";

export interface HPCQueryAdapter {
  readonly chainId: string;
  readonly accountAddress: string;
  getWorkloadTemplates(): Promise<unknown>;
  getQuote(request: HPCQuoteRequest): Promise<unknown>;
  getJobs(): Promise<unknown>;
  getJob(jobId: string): Promise<unknown>;
  subscribeToJob?(
    jobId: string,
    callback: (event: unknown) => void,
  ): () => void;
}

export interface HPCQuoteRequest {
  offeringId: string;
  resources: JobResources;
}

export interface HPCQueryEnvelope {
  chainId: string;
  accountAddress: string;
}

export class HPCQueryValidationError extends Error {
  readonly code = "hpc_query_invalid";

  constructor() {
    super("HPC query did not return authoritative bound evidence");
    this.name = "HPCQueryValidationError";
  }
}

function fail(): never {
  throw new HPCQueryValidationError();
}
const isRecord = (value: unknown): value is Record<string, unknown> =>
  !!value && typeof value === "object" && !Array.isArray(value);
const text = (value: unknown): string =>
  typeof value === "string" && value.trim() ? value : fail();
const optionalText = (value: unknown): string | undefined =>
  value === undefined ? undefined : text(value);
const integer = (value: unknown, minimum = 0): number =>
  Number.isSafeInteger(value) && (value as number) >= minimum
    ? (value as number)
    : fail();
const finite = (value: unknown, minimum = 0): number =>
  typeof value === "number" && Number.isFinite(value) && value >= minimum
    ? value
    : fail();
const bounded = (value: unknown, minimum: number, maximum: number): number => {
  const result = finite(value, minimum);
  return result <= maximum ? result : fail();
};
const amount = (value: unknown): string => {
  const result = text(value);
  return /^(0|[1-9]\d*)(\.\d{1,18})?$/.test(result) ? result : fail();
};
const amountUnits = (value: string): bigint => {
  const [whole, fraction = ""] = value.split(".");
  return BigInt(`${whole}${fraction.padEnd(18, "0")}`);
};
const unique = <T>(values: readonly T[], key: (value: T) => string): void => {
  if (new Set(values.map(key)).size !== values.length) fail();
};

function envelope(
  value: unknown,
  expected: HPCQueryEnvelope,
): Record<string, unknown> {
  if (!isRecord(value)) fail();
  if (
    value.chainId !== expected.chainId ||
    value.accountAddress !== expected.accountAddress
  ) {
    fail();
  }
  return value;
}

function resources(value: unknown): JobResources {
  if (!isRecord(value)) fail();
  const result = {
    nodes: integer(value.nodes, 1),
    cpusPerNode: integer(value.cpusPerNode, 1),
    memoryGBPerNode: integer(value.memoryGBPerNode, 1),
    gpusPerNode:
      value.gpusPerNode === undefined
        ? undefined
        : integer(value.gpusPerNode, 0),
    gpuType: optionalText(value.gpuType),
    maxRuntimeSeconds: integer(value.maxRuntimeSeconds, 1),
    storageGB: integer(value.storageGB, 0),
  };
  if (result.gpuType && !result.gpusPerNode) fail();
  return Object.freeze(result);
}

function resourcesEqual(left: JobResources, right: JobResources): boolean {
  return (
    left.nodes === right.nodes &&
    left.cpusPerNode === right.cpusPerNode &&
    left.memoryGBPerNode === right.memoryGBPerNode &&
    left.gpusPerNode === right.gpusPerNode &&
    left.gpuType === right.gpuType &&
    left.maxRuntimeSeconds === right.maxRuntimeSeconds &&
    left.storageGB === right.storageGB
  );
}

function parameter(value: unknown): JobParameter {
  if (!isRecord(value)) fail();
  const type = value.type;
  if (
    !["string", "number", "boolean", "select", "file"].includes(type as string)
  )
    fail();
  if (typeof value.required !== "boolean") fail();
  const defaultValue = value.defaultValue;
  if (
    defaultValue !== undefined &&
    typeof defaultValue !== "string" &&
    typeof defaultValue !== "number" &&
    typeof defaultValue !== "boolean"
  )
    fail();
  if (typeof defaultValue === "number" && !Number.isFinite(defaultValue))
    fail();
  const options =
    value.options === undefined
      ? undefined
      : Array.isArray(value.options)
        ? value.options.map((option) => {
            if (!isRecord(option)) fail();
            return Object.freeze({
              value: text(option.value),
              label: text(option.label),
            });
          })
        : fail();
  if (type === "select" && (!options || options.length === 0)) fail();
  if (options) unique(options, (option) => option.value);
  const min = value.min === undefined ? undefined : finite(value.min);
  const max = value.max === undefined ? undefined : finite(value.max);
  if (min !== undefined && max !== undefined && min > max) fail();
  const validationPattern = optionalText(value.validationPattern);
  if (validationPattern) {
    try {
      new RegExp(validationPattern);
    } catch {
      fail();
    }
  }
  if (defaultValue !== undefined) {
    if (type === "number" && typeof defaultValue !== "number") fail();
    if (type === "boolean" && typeof defaultValue !== "boolean") fail();
    if (
      ["string", "file", "select"].includes(type as string) &&
      typeof defaultValue !== "string"
    ) {
      fail();
    }
    if (
      type === "select" &&
      !options?.some((option) => option.value === defaultValue)
    )
      fail();
    if (typeof defaultValue === "number") {
      if (min !== undefined && defaultValue < min) fail();
      if (max !== undefined && defaultValue > max) fail();
    }
    if (typeof defaultValue === "string" && validationPattern) {
      if (!new RegExp(validationPattern).test(defaultValue)) fail();
    }
  }
  const validatedDefaultValue = defaultValue as
    | string
    | number
    | boolean
    | undefined;
  return Object.freeze({
    name: text(value.name),
    type: type as JobParameter["type"],
    description: text(value.description),
    required: value.required,
    defaultValue: validatedDefaultValue,
    options: options
      ? (Object.freeze(options) as unknown as NonNullable<
          JobParameter["options"]
        >)
      : undefined,
    validationPattern,
    min,
    max,
  });
}

function template(value: unknown): WorkloadTemplate {
  if (!isRecord(value) || !isRecord(value.defaultParameters)) fail();
  const category = value.category;
  if (
    ![
      "ml_training",
      "ml_inference",
      "scientific",
      "rendering",
      "simulation",
      "data_processing",
      "custom",
    ].includes(category as string)
  )
    fail();
  const defaultParameters = Object.fromEntries(
    Object.entries(value.defaultParameters).map(([key, item]) => [
      key,
      parameter(item),
    ]),
  );
  return Object.freeze({
    id: text(value.id),
    name: text(value.name),
    description: text(value.description),
    category: category as WorkloadTemplate["category"],
    defaultResources: resources(value.defaultResources),
    defaultParameters: Object.freeze(defaultParameters),
    requiredIdentityScore: bounded(value.requiredIdentityScore, 0, 100),
    mfaRequired:
      typeof value.mfaRequired === "boolean" ? value.mfaRequired : fail(),
    estimatedCostPerHour: amount(value.estimatedCostPerHour),
    version: text(value.version),
    iconUrl: optionalText(value.iconUrl),
    docsUrl: optionalText(value.docsUrl),
  });
}

function outputReference(value: unknown): JobOutputReference {
  if (!isRecord(value)) fail();
  const type = value.type;
  if (
    !["model", "checkpoint", "logs", "metrics", "artifact", "data"].includes(
      type as string,
    )
  )
    fail();
  const createdAt = integer(value.createdAt, 1);
  const expiresAt =
    value.expiresAt === undefined ? undefined : integer(value.expiresAt, 1);
  if (expiresAt !== undefined && expiresAt <= createdAt) fail();
  return Object.freeze({
    id: text(value.id),
    name: text(value.name),
    type: type as JobOutputReference["type"],
    sizeBytes: integer(value.sizeBytes),
    createdAt,
    encryptedRef: text(value.encryptedRef),
    contentHash: text(value.contentHash),
    expiresAt,
  });
}

const statuses: readonly JobStatus[] = [
  "pending",
  "queued",
  "running",
  "completing",
  "completed",
  "failed",
  "cancelled",
  "timeout",
];

const allowedTransitions: Readonly<Record<JobStatus, readonly JobStatus[]>> = {
  pending: ["queued", "failed", "cancelled"],
  queued: ["running", "failed", "cancelled", "timeout"],
  running: ["completing", "failed", "cancelled", "timeout"],
  completing: ["completed", "failed", "timeout"],
  completed: [],
  failed: [],
  cancelled: [],
  timeout: [],
};

function statusChange(value: unknown): JobStatusChange {
  if (
    !isRecord(value) ||
    !statuses.includes(value.fromStatus as JobStatus) ||
    !statuses.includes(value.toStatus as JobStatus)
  )
    fail();
  return Object.freeze({
    fromStatus: value.fromStatus as JobStatus,
    toStatus: value.toStatus as JobStatus,
    timestamp: integer(value.timestamp, 1),
    blockHeight: integer(value.blockHeight, 1),
    txHash: text(value.txHash),
    reason: optionalText(value.reason),
  });
}

function frozenValue(value: unknown): unknown {
  if (Array.isArray(value)) return Object.freeze(value.map(frozenValue));
  if (isRecord(value)) return frozenData(value);
  if (value === null || typeof value === "string" || typeof value === "boolean")
    return value;
  if (typeof value === "number" && Number.isFinite(value)) return value;
  return fail();
}

function frozenData(
  value: Record<string, unknown>,
): Readonly<Record<string, unknown>> {
  const result: Record<string, unknown> = {};
  for (const [key, item] of Object.entries(value)) {
    if (!key.trim()) fail();
    result[key] = frozenValue(item);
  }
  return Object.freeze(result);
}

function event(value: unknown): JobEvent {
  if (!isRecord(value)) fail();
  const type = value.type;
  if (
    ![
      "job_submitted",
      "job_scheduled",
      "job_started",
      "checkpoint_saved",
      "progress_updated",
      "output_available",
      "job_completed",
      "job_failed",
      "job_cancelled",
      "usage_recorded",
      "settlement_processed",
    ].includes(type as string)
  )
    fail();
  return Object.freeze({
    id: text(value.id),
    type: type as JobEvent["type"],
    timestamp: integer(value.timestamp, 1),
    blockHeight: integer(value.blockHeight, 1),
    data: isRecord(value.data) ? frozenData(value.data) : fail(),
  });
}

function job(value: unknown, expectedAccount: string): Job {
  if (
    !isRecord(value) ||
    value.customerAddress !== expectedAccount ||
    !statuses.includes(value.status as JobStatus)
  )
    fail();
  const statusHistory = Array.isArray(value.statusHistory)
    ? value.statusHistory.map(statusChange)
    : fail();
  const events = Array.isArray(value.events) ? value.events.map(event) : fail();
  const outputRefs = Array.isArray(value.outputRefs)
    ? value.outputRefs.map(outputReference)
    : fail();
  unique(events, (item) => item.id);
  unique(outputRefs, (item) => item.id);
  const createdAt = integer(value.createdAt, 1);
  const startedAt =
    value.startedAt === undefined
      ? undefined
      : integer(value.startedAt, createdAt);
  const completedAt =
    value.completedAt === undefined
      ? undefined
      : integer(value.completedAt, startedAt ?? createdAt);
  for (let index = 0; index < statusHistory.length; index += 1) {
    const change = statusHistory[index];
    if (!allowedTransitions[change.fromStatus].includes(change.toStatus))
      fail();
    if (
      change.timestamp < createdAt ||
      (completedAt !== undefined && change.timestamp > completedAt)
    ) {
      fail();
    }
    if (index > 0) {
      const previous = statusHistory[index - 1];
      if (
        change.fromStatus !== previous.toStatus ||
        change.timestamp < previous.timestamp ||
        change.blockHeight < previous.blockHeight
      )
        fail();
    }
  }
  if (statusHistory.length > 0 && statusHistory[0].fromStatus !== "pending")
    fail();
  if (
    statusHistory.length > 0 &&
    statusHistory[statusHistory.length - 1].toStatus !== value.status
  ) {
    fail();
  }
  if (
    !["pending", "queued"].includes(value.status as string) &&
    statusHistory.length === 0
  )
    fail();
  const requiresStart = ["running", "completing", "completed"].includes(
    value.status as string,
  );
  const terminal = ["completed", "failed", "cancelled", "timeout"].includes(
    value.status as string,
  );
  if (requiresStart && startedAt === undefined) fail();
  if (
    ["pending", "queued"].includes(value.status as string) &&
    startedAt !== undefined
  )
    fail();
  if (
    (terminal && completedAt === undefined) ||
    (!terminal && completedAt !== undefined)
  )
    fail();
  const runningChange = statusHistory.find(
    (change) => change.toStatus === "running",
  );
  const terminalChange = statusHistory.find(
    (change) => change.toStatus === value.status,
  );
  if (startedAt !== undefined && runningChange?.timestamp !== startedAt) fail();
  if (completedAt !== undefined && terminalChange?.timestamp !== completedAt)
    fail();
  for (const item of events) {
    if (
      item.timestamp < createdAt ||
      (completedAt !== undefined && item.timestamp > completedAt)
    )
      fail();
  }
  const depositStatus = value.depositStatus;
  if (!["held", "released", "forfeited"].includes(depositStatus as string))
    fail();
  const totalCost = amount(value.totalCost);
  const depositAmount = amount(value.depositAmount);
  if (amountUnits(depositAmount) < amountUnits(totalCost)) fail();
  if (!terminal && depositStatus !== "held") fail();
  if (
    ["completed", "cancelled"].includes(value.status as string) &&
    depositStatus !== "released"
  ) {
    fail();
  }
  if (
    ["failed", "timeout"].includes(value.status as string) &&
    depositStatus === "held"
  )
    fail();
  return Object.freeze({
    id: text(value.id),
    name: text(value.name),
    customerAddress: expectedAccount,
    providerAddress: text(value.providerAddress),
    offeringId: text(value.offeringId),
    templateId: optionalText(value.templateId),
    status: value.status as JobStatus,
    createdAt,
    startedAt,
    completedAt,
    resources: resources(value.resources),
    statusHistory: Object.freeze(statusHistory) as unknown as JobStatusChange[],
    events: Object.freeze(events) as unknown as JobEvent[],
    outputRefs: Object.freeze(outputRefs) as unknown as JobOutputReference[],
    totalCost,
    depositAmount,
    depositStatus: depositStatus as Job["depositStatus"],
    txHash: text(value.txHash),
  });
}

export function requireHPCQueryAdapter(
  adapter: HPCQueryAdapter | undefined,
  expected: HPCQueryEnvelope,
): HPCQueryAdapter {
  if (
    !adapter ||
    adapter.chainId !== expected.chainId ||
    adapter.accountAddress !== expected.accountAddress
  ) {
    throw new HPCClientUnavailableError("query");
  }
  return adapter;
}

export function validateHPCWorkloadTemplates(
  value: unknown,
  expected: HPCQueryEnvelope,
): readonly WorkloadTemplate[] {
  const source = envelope(value, expected).templates;
  if (!Array.isArray(source)) fail();
  const result = source.map(template);
  unique(result, (item) => item.id);
  return Object.freeze(result);
}

export function validateHPCJobPriceQuote(
  value: unknown,
  expected: HPCQueryEnvelope,
  expectedRequest: HPCQuoteRequest,
): JobPriceQuote {
  const sourceEnvelope = envelope(value, expected);
  if (sourceEnvelope.offeringId !== expectedRequest.offeringId) fail();
  const quotedResources = resources(sourceEnvelope.resources);
  if (!resourcesEqual(quotedResources, expectedRequest.resources)) fail();
  const source = sourceEnvelope.quote;
  if (!isRecord(source) || !isRecord(source.breakdown)) fail();
  const result = Object.freeze({
    estimatedTotal: amount(source.estimatedTotal),
    depositRequired: amount(source.depositRequired),
    breakdown: Object.freeze({
      compute: amount(source.breakdown.compute),
      storage: amount(source.breakdown.storage),
      network: amount(source.breakdown.network),
      gpu:
        source.breakdown.gpu === undefined
          ? undefined
          : amount(source.breakdown.gpu),
    }),
    pricePerHour: amount(source.pricePerHour),
    maxHours: finite(source.maxHours, Number.MIN_VALUE),
    denom: text(source.denom),
  });
  if (!/^[a-z][a-z0-9/._-]{1,127}$/.test(result.denom)) fail();
  const components = [
    result.breakdown.compute,
    result.breakdown.storage,
    result.breakdown.network,
    result.breakdown.gpu ?? "0",
  ].map(amountUnits);
  const componentTotal = components.reduce(
    (sum, component) => sum + component,
    0n,
  );
  const total = amountUnits(result.estimatedTotal);
  const deposit = amountUnits(result.depositRequired);
  const hourly = amountUnits(result.pricePerHour);
  if (componentTotal !== total) fail();
  if (deposit < total) fail();
  if (result.maxHours !== expectedRequest.resources.maxRuntimeSeconds / 3600)
    fail();
  if (
    total * 3600n !==
    hourly * BigInt(expectedRequest.resources.maxRuntimeSeconds)
  )
    fail();
  return result;
}

export function validateHPCQuoteRequest(
  value: HPCQuoteRequest,
): HPCQuoteRequest {
  return Object.freeze({
    offeringId: text(value.offeringId),
    resources: resources(value.resources),
  });
}

export function validateHPCJobSubscriptionEvent(
  value: unknown,
  jobId: string,
): ChainEvent {
  if (
    !isRecord(value) ||
    !isRecord(value.attributes) ||
    value.attributes.jobId !== jobId
  )
    fail();
  const attributes: Record<string, string> = {};
  for (const [key, item] of Object.entries(value.attributes)) {
    if (!key.trim() || typeof item !== "string") fail();
    attributes[key] = item;
  }
  return Object.freeze({
    query: text(value.query),
    type: text(value.type),
    attributes: Object.freeze(attributes),
    blockHeight: integer(value.blockHeight, 1),
    txHash: optionalText(value.txHash),
    timestamp: integer(value.timestamp, 1),
  });
}

export function validateHPCJobs(
  value: unknown,
  expected: HPCQueryEnvelope,
): readonly Job[] {
  const source = envelope(value, expected).jobs;
  if (!Array.isArray(source)) fail();
  const result = source.map((item) => job(item, expected.accountAddress));
  unique(result, (item) => item.id);
  return Object.freeze(result);
}

export function validateHPCJob(
  value: unknown,
  expected: HPCQueryEnvelope & { jobId: string },
): Job {
  const source = envelope(value, expected);
  if (source.jobId !== expected.jobId) fail();
  const result = job(source.job, expected.accountAddress);
  if (result.id !== expected.jobId) fail();
  return result;
}
