/**
 * HPC Client
 *
 * Wrapper around VirtEngine SDK for HPC operations.
 */

import type {
  HPCJob,
  HPCOffering,
  JobState,
  PreconfiguredWorkload,
  VirtEngineClient,
} from '@virtengine/chain-sdk';
import { getChainClient } from '@/lib/chain-sdk';
import type {
  Job,
  JobOutput,
  JobStatus,
  SDKOffering,
  WorkloadTemplate,
  JobResources,
} from '../types';

/**
 * HPC Client Configuration
 */
export interface HPCClientConfig {
  /**
   * Chain RPC endpoint
   */
  rpcEndpoint?: string;

  /**
   * User address (for filtering user's jobs)
   */
  userAddress?: string;

  /**
   * Optional injected SDK client
   */
  chainClient?: VirtEngineClient;
}

/**
 * Job submission parameters
 */
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
}

const WORKLOAD_CATEGORIES = new Set([
  'ml_training',
  'ml_inference',
  'scientific',
  'rendering',
  'simulation',
  'data_processing',
  'custom',
]);

type SDKJobResources = NonNullable<HPCJob['resources']>;

function normalizeCategory(value?: string): WorkloadTemplate['category'] {
  if (!value) return 'custom';
  const normalized = value.toLowerCase().replace(/\s+/g, '_');
  return WORKLOAD_CATEGORIES.has(normalized)
    ? (normalized as WorkloadTemplate['category'])
    : 'custom';
}

function longToNumber(value?: { toNumber?: () => number } | number | null): number {
  if (value === null || value === undefined) return 0;
  if (typeof value === 'number') return value;
  if (typeof value === 'object' && typeof value.toNumber === 'function') {
    return value.toNumber();
  }
  return Number(value);
}

export function mapJobResources(
  resources: SDKJobResources | undefined,
  maxRuntimeSeconds?: { toNumber?: () => number } | number | null
): JobResources {
  const runtime = longToNumber(maxRuntimeSeconds);
  return {
    nodes: resources?.nodes ?? 1,
    cpusPerNode: resources?.cpuCoresPerNode ?? 1,
    memoryGBPerNode: resources?.memoryGbPerNode ?? 1,
    gpusPerNode: resources?.gpusPerNode || undefined,
    gpuType: resources?.gpuType || undefined,
    maxRuntimeSeconds: runtime > 0 ? runtime : 86400,
    storageGB: resources?.storageGb ?? 0,
  };
}

export function mapJobState(state: JobState): JobStatus {
  switch (state) {
    case JobState.JOB_STATE_PENDING:
      return 'pending';
    case JobState.JOB_STATE_QUEUED:
      return 'queued';
    case JobState.JOB_STATE_RUNNING:
      return 'running';
    case JobState.JOB_STATE_COMPLETED:
      return 'completed';
    case JobState.JOB_STATE_FAILED:
      return 'failed';
    case JobState.JOB_STATE_CANCELLED:
      return 'cancelled';
    case JobState.JOB_STATE_TIMEOUT:
      return 'timeout';
    default:
      return 'pending';
  }
}

export function mapSdkJob(job: HPCJob): Job {
  const status = mapJobState(job.state);
  const resources = mapJobResources(job.resources, job.maxRuntimeSeconds);

  return {
    id: job.jobId,
    name: job.workloadSpec?.preconfiguredWorkloadId || job.jobId,
    customerAddress: job.customerAddress,
    providerAddress: job.providerAddress,
    offeringId: job.offeringId,
    templateId: job.workloadSpec?.isPreconfigured
      ? job.workloadSpec.preconfiguredWorkloadId
      : undefined,
    status,
    createdAt: job.createdAt?.getTime() ?? 0,
    startedAt: job.startedAt?.getTime(),
    completedAt: job.completedAt?.getTime(),
    resources,
    statusHistory: [],
    events: [],
    outputRefs: [],
    totalCost: job.agreedPrice?.[0]?.amount ?? '0',
    depositAmount: '0',
    depositStatus:
      status === 'completed'
        ? 'released'
        : status === 'failed' || status === 'cancelled' || status === 'timeout'
          ? 'forfeited'
          : 'held',
    txHash: '',
  };
}

export function mapSdkWorkload(workload: PreconfiguredWorkload): WorkloadTemplate {
  const resources = mapJobResources(workload.requiredResources, null);
  const estimatedCostPerHour =
    (workload as { estimatedCostPerHour?: string }).estimatedCostPerHour ?? '0';
  return {
    id: workload.workloadId,
    name: workload.name || workload.workloadId,
    description: workload.description ?? '',
    category: normalizeCategory(workload.category),
    defaultResources: resources,
    defaultParameters: {},
    requiredIdentityScore: 0,
    mfaRequired: false,
    estimatedCostPerHour,
    version: workload.version || '1.0.0',
  };
}

export function collectWorkloadTemplates(offerings: HPCOffering[]): WorkloadTemplate[] {
  const templates = new Map<string, WorkloadTemplate>();

  offerings.forEach((offering) => {
    offering.preconfiguredWorkloads?.forEach((workload) => {
      const mapped = mapSdkWorkload(workload);
      if (!templates.has(mapped.id)) {
        templates.set(mapped.id, mapped);
      }
    });
  });

  return Array.from(templates.values());
}

export function estimateJobCostFromOffering(
  offering: HPCOffering | null,
  resources: JobResources
): {
  estimatedTotal: string;
  pricePerHour: string;
  breakdown: {
    compute: string;
    storage: string;
    network: string;
    gpu?: string;
  };
  denom: string;
} {
  if (!offering?.pricing) {
    return {
      estimatedTotal: '0.00',
      pricePerHour: '0.00',
      breakdown: {
        compute: '0.00',
        storage: '0.00',
        network: '0.00',
      },
      denom: 'uvirt',
    };
  }

  const pricing = offering.pricing;
  const toNumber = (value?: string) => (value ? Number.parseFloat(value) : 0);

  const basePrice = resources.nodes * toNumber(pricing.baseNodeHourPrice);
  const cpuPrice = resources.nodes * resources.cpusPerNode * toNumber(pricing.cpuCoreHourPrice);
  const memoryPrice =
    resources.nodes * resources.memoryGBPerNode * toNumber(pricing.memoryGbHourPrice);
  const gpuPrice = (resources.gpusPerNode ?? 0) * resources.nodes * toNumber(pricing.gpuHourPrice);
  const storagePrice = resources.storageGB * toNumber(pricing.storageGbPrice);
  const networkPrice = toNumber(pricing.networkGbPrice);

  const hourlyRate = basePrice + cpuPrice + memoryPrice + gpuPrice + storagePrice + networkPrice;
  const maxHours = Math.max(1, resources.maxRuntimeSeconds / 3600);
  const total = hourlyRate * maxHours;

  return {
    estimatedTotal: total.toFixed(2),
    pricePerHour: hourlyRate.toFixed(2),
    breakdown: {
      compute: (basePrice + cpuPrice + memoryPrice).toFixed(2),
      storage: storagePrice.toFixed(2),
      network: networkPrice.toFixed(2),
      gpu: resources.gpusPerNode ? gpuPrice.toFixed(2) : undefined,
    },
    denom: pricing.currency || 'uvirt',
  };
}

/**
 * HPC Client
 *
 * Provides methods for interacting with the HPC module on chain.
 */
export class HPCClient {
  private config: HPCClientConfig;
  private chainClient: Promise<VirtEngineClient>;

  constructor(config: HPCClientConfig = {}) {
    this.config = config;
    this.chainClient = config.chainClient ? Promise.resolve(config.chainClient) : getChainClient();
  }

  private async sdk(): Promise<VirtEngineClient> {
    return this.chainClient;
  }

  /**
   * List available workload templates
   */
  async listWorkloadTemplates(): Promise<WorkloadTemplate[]> {
    const client = await this.sdk();
    const offerings = await client.hpc.listOfferings({ activeOnly: true });
    return collectWorkloadTemplates(offerings);
  }

  /**
   * Get workload template by ID
   */
  async getWorkloadTemplate(templateId: string): Promise<WorkloadTemplate | null> {
    const templates = await this.listWorkloadTemplates();
    return templates.find((t) => t.id === templateId) ?? null;
  }

  /**
   * List available offerings
   */
  async listOfferings(): Promise<SDKOffering[]> {
    const client = await this.sdk();
    return client.hpc.listOfferings({ activeOnly: true });
  }

  /**
   * Get offering by ID
   */
  async getOffering(offeringId: string): Promise<SDKOffering | null> {
    const client = await this.sdk();
    return client.hpc.getOffering(offeringId);
  }

  /**
   * List user's jobs
   */
  async listJobs(filters?: { status?: JobStatus[] }): Promise<Job[]> {
    const client = await this.sdk();
    const jobs = await client.hpc.listJobs({
      customerAddress: this.config.userAddress,
    });

    const mapped = jobs.map(mapSdkJob);
    if (!filters?.status?.length) return mapped;

    return mapped.filter((job) => filters.status?.includes(job.status));
  }

  /**
   * Get job by ID
   */
  async getJob(jobId: string): Promise<Job | null> {
    const client = await this.sdk();
    const job = await client.hpc.getJob(jobId);
    return job ? mapSdkJob(job) : null;
  }

  /**
   * Submit a new job
   */
  submitJob(_params: SubmitJobParams): Promise<{ jobId: string; txHash: string }> {
    const jobId = `job-${Date.now()}`;
    const txHash = `0x${Math.random().toString(16).substring(2, 66)}`;

    return Promise.resolve({ jobId, txHash });
  }

  /**
   * Cancel a job
   */
  cancelJob(_jobId: string): Promise<{ txHash: string }> {
    const txHash = `0x${Math.random().toString(16).substring(2, 66)}`;

    return Promise.resolve({ txHash });
  }

  /**
   * Get job logs
   */
  getJobLogs(
    _jobId: string,
    options?: { tail?: number; since?: number }
  ): Promise<{ lines: string[]; hasMore: boolean }> {
    const tail = options?.tail ?? 100;
    const lines = MOCK_LOG_LINES.slice(-tail);

    return Promise.resolve({ lines, hasMore: MOCK_LOG_LINES.length > tail });
  }

  /**
   * Get job outputs
   */
  async getJobOutputs(_jobId: string): Promise<JobOutput[]> {
    const job = await this.getJob(_jobId);
    if (!job || job.status !== 'completed') {
      return [];
    }
    return MOCK_OUTPUTS;
  }

  /**
   * Get job resource usage
   */
  async getJobUsage(_jobId: string): Promise<{
    cpuPercent: number;
    memoryPercent: number;
    gpuPercent?: number;
    elapsedSeconds: number;
    estimatedRemainingSeconds?: number;
  }> {
    const job = await this.getJob(_jobId);
    if (!job || job.status !== 'running') {
      return {
        cpuPercent: 0,
        memoryPercent: 0,
        elapsedSeconds: 0,
      };
    }
    return {
      cpuPercent: 42,
      memoryPercent: 67,
      elapsedSeconds: Math.floor((Date.now() - (job.startedAt ?? Date.now())) / 1000),
    };
  }

  /**
   * Estimate job cost
   */
  async estimateJobCost(
    offeringId: string,
    resources: SubmitJobParams['resources']
  ): Promise<{
    estimatedTotal: string;
    pricePerHour: string;
    breakdown: {
      compute: string;
      storage: string;
      network: string;
      gpu?: string;
    };
    denom: string;
  }> {
    const offering = await this.getOffering(offeringId);
    const mappedResources: JobResources = {
      nodes: resources.nodes,
      cpusPerNode: resources.cpusPerNode,
      memoryGBPerNode: resources.memoryGBPerNode,
      gpusPerNode: resources.gpusPerNode,
      gpuType: resources.gpuType,
      maxRuntimeSeconds: resources.maxRuntimeSeconds,
      storageGB: resources.storageGB,
    };

    return estimateJobCostFromOffering(offering, mappedResources);
  }
}

/**
 * Create HPC client instance
 */
export function createHPCClient(config?: HPCClientConfig): HPCClient {
  return new HPCClient(config);
}

const MOCK_LOG_LINES: string[] = [
  '[2026-02-06T22:00:01Z] INFO  Starting job initialization...',
  '[2026-02-06T22:00:02Z] INFO  Loading container image: pytorch/pytorch:2.1-cuda12',
  '[2026-02-06T22:00:05Z] INFO  Image loaded successfully',
  '[2026-02-06T22:00:06Z] INFO  Mounting storage volumes...',
  '[2026-02-06T22:00:07Z] INFO  Volume /data mounted (100GB)',
  '[2026-02-06T22:00:08Z] INFO  Setting up environment variables',
  '[2026-02-06T22:00:09Z] INFO  GPU devices detected: 2x NVIDIA A100',
  '[2026-02-06T22:00:10Z] INFO  CUDA version: 12.1',
  '[2026-02-06T22:00:11Z] INFO  Starting training script...',
  '[2026-02-06T22:01:00Z] INFO  Epoch 1/100 - loss: 2.3456 - acc: 0.1234',
  '[2026-02-06T22:02:00Z] INFO  Epoch 2/100 - loss: 1.8901 - acc: 0.2567',
  '[2026-02-06T22:03:00Z] INFO  Epoch 3/100 - loss: 1.5432 - acc: 0.3891',
  '[2026-02-06T22:04:00Z] INFO  Epoch 4/100 - loss: 1.2100 - acc: 0.4890',
  '[2026-02-06T22:05:00Z] INFO  Epoch 5/100 - loss: 0.9876 - acc: 0.5678',
  '[2026-02-06T22:05:30Z] INFO  Checkpoint saved: epoch_5.pt',
  '[2026-02-06T22:06:00Z] INFO  Epoch 6/100 - loss: 0.8123 - acc: 0.6234',
  '[2026-02-06T22:07:00Z] INFO  Epoch 7/100 - loss: 0.6890 - acc: 0.6891',
  '[2026-02-06T22:08:00Z] INFO  Epoch 8/100 - loss: 0.5432 - acc: 0.7456',
];

const MOCK_OUTPUTS: JobOutput[] = [
  {
    refId: 'out-1',
    name: 'model_final.pt',
    type: 'model',
    accessUrl: '#',
    urlExpiresAt: Date.now() + 86400000,
    sizeBytes: 1048576000,
    mimeType: 'application/octet-stream',
  },
  {
    refId: 'out-2',
    name: 'training.log',
    type: 'logs',
    accessUrl: '#',
    urlExpiresAt: Date.now() + 86400000,
    sizeBytes: 524288,
    mimeType: 'text/plain',
  },
  {
    refId: 'out-3',
    name: 'metrics.json',
    type: 'metrics',
    accessUrl: '#',
    urlExpiresAt: Date.now() + 86400000,
    sizeBytes: 8192,
    mimeType: 'application/json',
  },
];
