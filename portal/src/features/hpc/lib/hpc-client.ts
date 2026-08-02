/**
 * HPC client capability boundary.
 *
 * Production callers must inject authoritative query, signer, and provider
 * adapters. The default client is deliberately unavailable.
 */

import type { Job, JobOutput, JobStatus, SDKOffering, WorkloadTemplate } from '../types';
import {
  HPCClientUnavailableError,
  assertCommittedJobMutation,
  requireHPCSigner,
  type CommittedJobMutation,
  type HPCSignerAdapter,
  type SubmitJobParams,
} from '@/lib/portal-adapter';

export {
  HPCClientUnavailableError,
  HPCMutationNotCommittedError,
  assertCommittedJobMutation,
} from '@/lib/portal-adapter';
export type {
  CommittedJobMutation,
  HPCClientCapability,
  HPCSignerAdapter,
  SubmitJobParams,
} from '@/lib/portal-adapter';

/**
 * HPC Client Configuration
 */
export interface JobUsage {
  cpuPercent: number;
  memoryPercent: number;
  gpuPercent?: number;
  elapsedSeconds: number;
  estimatedRemainingSeconds?: number;
}

export interface JobCostEstimate {
  estimatedTotal: string;
  pricePerHour: string;
  breakdown: {
    compute: string;
    storage: string;
    network: string;
    gpu?: string;
  };
  denom: string;
}

export interface HPCQueryAdapter {
  listWorkloadTemplates(): Promise<WorkloadTemplate[]>;
  getWorkloadTemplate(templateId: string): Promise<WorkloadTemplate | null>;
  listOfferings(): Promise<SDKOffering[]>;
  getOffering(offeringId: string): Promise<SDKOffering | null>;
  listJobs(filters?: { status?: JobStatus[] }): Promise<Job[]>;
  getJob(jobId: string): Promise<Job | null>;
  estimateJobCost(
    offeringId: string,
    resources: SubmitJobParams['resources']
  ): Promise<JobCostEstimate>;
}

export interface HPCProviderAdapter {
  getJobLogs(
    jobId: string,
    options?: { tail?: number; since?: number }
  ): Promise<{ lines: string[]; hasMore: boolean }>;
  getJobOutputs(jobId: string): Promise<JobOutput[]>;
  getJobUsage(jobId: string): Promise<JobUsage>;
}

export interface HPCClientDependencies {
  query?: HPCQueryAdapter;
  signer?: HPCSignerAdapter;
  provider?: HPCProviderAdapter;
}

/**
 * HPC Client
 *
 * Provides methods for interacting with the HPC module on chain.
 */
export class HPCClient {
  constructor(private readonly dependencies: HPCClientDependencies = {}) {}

  /**
   * List available workload templates
   */
  async listWorkloadTemplates(): Promise<WorkloadTemplate[]> {
    return this.requireQuery().listWorkloadTemplates();
  }

  /**
   * Get workload template by ID
   */
  async getWorkloadTemplate(templateId: string): Promise<WorkloadTemplate | null> {
    return this.requireQuery().getWorkloadTemplate(templateId);
  }

  /**
   * List available offerings
   */
  async listOfferings(): Promise<SDKOffering[]> {
    return this.requireQuery().listOfferings();
  }

  /**
   * Get offering by ID
   */
  async getOffering(offeringId: string): Promise<SDKOffering | null> {
    return this.requireQuery().getOffering(offeringId);
  }

  /**
   * List user's jobs
   */
  async listJobs(filters?: { status?: JobStatus[] }): Promise<Job[]> {
    return this.requireQuery().listJobs(filters);
  }

  /**
   * Get job by ID
   */
  async getJob(jobId: string): Promise<Job | null> {
    return this.requireQuery().getJob(jobId);
  }

  /**
   * Submit a new job
   */
  async submitJob(params: SubmitJobParams): Promise<CommittedJobMutation> {
    const result = await this.requireSigner().submitJob(params);
    return this.requireCommittedMutation(result);
  }

  /**
   * Cancel a job
   */
  async cancelJob(jobId: string): Promise<CommittedJobMutation> {
    const result = await this.requireSigner().cancelJob(jobId);
    return this.requireCommittedMutation(result, jobId);
  }

  /**
   * Get job logs
   */
  async getJobLogs(
    jobId: string,
    options?: { tail?: number; since?: number }
  ): Promise<{ lines: string[]; hasMore: boolean }> {
    this.requireQuery();
    return this.requireProvider().getJobLogs(jobId, options);
  }

  /**
   * Get job outputs
   */
  async getJobOutputs(jobId: string): Promise<JobOutput[]> {
    this.requireQuery();
    return this.requireProvider().getJobOutputs(jobId);
  }

  /**
   * Get job resource usage
   */
  async getJobUsage(jobId: string): Promise<JobUsage> {
    this.requireQuery();
    return this.requireProvider().getJobUsage(jobId);
  }

  /**
   * Estimate job cost
   */
  async estimateJobCost(
    offeringId: string,
    resources: SubmitJobParams['resources']
  ): Promise<JobCostEstimate> {
    return this.requireQuery().estimateJobCost(offeringId, resources);
  }

  private requireQuery(): HPCQueryAdapter {
    if (!this.dependencies.query) throw new HPCClientUnavailableError('query');
    return this.dependencies.query;
  }

  private requireSigner(): HPCSignerAdapter {
    return requireHPCSigner(this.dependencies.signer);
  }

  private requireProvider(): HPCProviderAdapter {
    if (!this.dependencies.provider) throw new HPCClientUnavailableError('provider');
    return this.dependencies.provider;
  }

  private requireCommittedMutation(result: unknown, expectedJobId?: string): CommittedJobMutation {
    assertCommittedJobMutation(result, expectedJobId);
    return result;
  }
}

/**
 * Mock Templates
 */
const MOCK_TEMPLATES: WorkloadTemplate[] = [
  {
    id: 'pytorch-training',
    name: 'PyTorch Training',
    description:
      'Train deep learning models with PyTorch. Supports distributed training across multiple GPUs.',
    category: 'ml_training',
    defaultResources: {
      nodes: 1,
      cpusPerNode: 8,
      memoryGBPerNode: 64,
      gpusPerNode: 2,
      gpuType: 'nvidia-a100',
      maxRuntimeSeconds: 86400,
      storageGB: 100,
    },
    defaultParameters: {},
    requiredIdentityScore: 0,
    mfaRequired: false,
    estimatedCostPerHour: '5.50',
    version: '1.0.0',
  },
  {
    id: 'tensorflow',
    name: 'TensorFlow',
    description: 'TensorFlow training pipeline with Keras integration and TensorBoard support.',
    category: 'ml_training',
    defaultResources: {
      nodes: 1,
      cpusPerNode: 8,
      memoryGBPerNode: 64,
      gpusPerNode: 2,
      gpuType: 'nvidia-a100',
      maxRuntimeSeconds: 86400,
      storageGB: 100,
    },
    defaultParameters: {},
    requiredIdentityScore: 0,
    mfaRequired: false,
    estimatedCostPerHour: '5.50',
    version: '1.0.0',
  },
  {
    id: 'openfoam',
    name: 'OpenFOAM',
    description: 'Computational fluid dynamics simulations with OpenFOAM.',
    category: 'scientific',
    defaultResources: {
      nodes: 4,
      cpusPerNode: 32,
      memoryGBPerNode: 128,
      maxRuntimeSeconds: 172800,
      storageGB: 500,
    },
    defaultParameters: {},
    requiredIdentityScore: 0,
    mfaRequired: false,
    estimatedCostPerHour: '12.00',
    version: '1.0.0',
  },
  {
    id: 'blender-render',
    name: 'Blender Render',
    description: '3D rendering and animation with Blender.',
    category: 'rendering',
    defaultResources: {
      nodes: 1,
      cpusPerNode: 16,
      memoryGBPerNode: 32,
      gpusPerNode: 1,
      gpuType: 'nvidia-a100',
      maxRuntimeSeconds: 43200,
      storageGB: 200,
    },
    defaultParameters: {},
    requiredIdentityScore: 0,
    mfaRequired: false,
    estimatedCostPerHour: '3.50',
    version: '1.0.0',
  },
];

/**
 * Mock Offerings
 */
const MOCK_OFFERINGS: SDKOffering[] = [
  {
    offeringId: 'offering-1',
    clusterId: 'cluster-1',
    providerAddress: 'virtengine1provider1...',
    name: 'Standard GPU Cluster',
    description: 'General purpose GPU compute with A100s',
    pricing: {
      baseNodeHourPrice: '1.00',
      cpuCoreHourPrice: '0.10',
      memoryGbHourPrice: '0.05',
      storageGbPrice: '0.01',
      networkGbPrice: '0.02',
      currency: 'uakt',
    },
    maxRuntimeSeconds: 604800, // 1 week
    supportsCustomWorkloads: true,
    preconfiguredWorkloads: [],
  },
];

/**
 * Mock Jobs
 */
const MOCK_JOBS: Job[] = [
  {
    id: 'job-401',
    name: 'ML Training - ResNet50',
    customerAddress: 'virtengine1customer...',
    providerAddress: 'virtengine1provider...',
    offeringId: 'offering-1',
    templateId: 'pytorch-training',
    status: 'running',
    createdAt: Date.now() - 7200000, // 2 hours ago
    startedAt: Date.now() - 6000000,
    resources: {
      nodes: 1,
      cpusPerNode: 8,
      memoryGBPerNode: 64,
      gpusPerNode: 2,
      gpuType: 'nvidia-a100',
      maxRuntimeSeconds: 86400,
      storageGB: 100,
    },
    statusHistory: [],
    events: [],
    outputRefs: [],
    totalCost: '11.00',
    depositAmount: '132.00',
    depositStatus: 'held',
    txHash: '0xabc123...',
  },
  {
    id: 'job-402',
    name: 'CFD Simulation',
    customerAddress: 'virtengine1customer...',
    providerAddress: 'virtengine1provider...',
    offeringId: 'offering-1',
    templateId: 'openfoam',
    status: 'queued',
    createdAt: Date.now() - 3600000, // 1 hour ago
    resources: {
      nodes: 4,
      cpusPerNode: 32,
      memoryGBPerNode: 128,
      maxRuntimeSeconds: 172800,
      storageGB: 500,
    },
    statusHistory: [],
    events: [],
    outputRefs: [],
    totalCost: '0.00',
    depositAmount: '576.00',
    depositStatus: 'held',
    txHash: '0xdef456...',
  },
  {
    id: 'job-403',
    name: 'Render Job #42',
    customerAddress: 'virtengine1customer...',
    providerAddress: 'virtengine1provider...',
    offeringId: 'offering-1',
    templateId: 'blender-render',
    status: 'completed',
    createdAt: Date.now() - 86400000, // 1 day ago
    startedAt: Date.now() - 82800000,
    completedAt: Date.now() - 72000000,
    resources: {
      nodes: 1,
      cpusPerNode: 16,
      memoryGBPerNode: 32,
      gpusPerNode: 1,
      gpuType: 'nvidia-a100',
      maxRuntimeSeconds: 43200,
      storageGB: 200,
    },
    statusHistory: [],
    events: [],
    outputRefs: [],
    totalCost: '10.50',
    depositAmount: '42.00',
    depositStatus: 'released',
    txHash: '0xghi789...',
  },
];

/**
 * Mock Log Lines
 */
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

/**
 * Mock Outputs
 */
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

/**
 * Create HPC client instance
 */
export function createHPCClient(dependencies: HPCClientDependencies = {}): HPCClient {
  return new HPCClient(dependencies);
}

/** Explicit fixture client for tests and development stories only. */
export function createMockHPCClient(): HPCClient {
  const query: HPCQueryAdapter = {
    listWorkloadTemplates: () => Promise.resolve(MOCK_TEMPLATES),
    getWorkloadTemplate: (templateId) =>
      Promise.resolve(MOCK_TEMPLATES.find((template) => template.id === templateId) ?? null),
    listOfferings: () => Promise.resolve(MOCK_OFFERINGS),
    getOffering: (offeringId) =>
      Promise.resolve(
        MOCK_OFFERINGS.find((offering) => offering.offeringId === offeringId) ?? null
      ),
    listJobs: (filters) =>
      Promise.resolve(
        filters?.status?.length
          ? MOCK_JOBS.filter((job) => filters.status?.includes(job.status))
          : MOCK_JOBS
      ),
    getJob: (jobId) => Promise.resolve(MOCK_JOBS.find((job) => job.id === jobId) ?? null),
    estimateJobCost: (_offeringId, resources) => {
      const compute = resources.nodes * resources.cpusPerNode * 0.5;
      const gpu = (resources.gpusPerNode ?? 0) * resources.nodes * 2.5;
      const storage = resources.storageGB * 0.01;
      const pricePerHour = compute + gpu + storage + 0.5;
      return Promise.resolve({
        estimatedTotal: (pricePerHour * (resources.maxRuntimeSeconds / 3600)).toFixed(2),
        pricePerHour: pricePerHour.toFixed(2),
        breakdown: {
          compute: compute.toFixed(2),
          storage: storage.toFixed(2),
          network: '0.50',
          gpu: resources.gpusPerNode ? gpu.toFixed(2) : undefined,
        },
        denom: 'uakt',
      });
    },
  };
  const signer: HPCSignerAdapter = {
    state: 'signing-ready',
    chainId: 'virtengine-1',
    accountAddress: 'virtengine1fixture',
    submitJob: () =>
      Promise.resolve({
        committed: true,
        jobId: 'fixture-job',
        txHash: 'fixture-submit',
        code: 0,
        blockHeight: 1,
      }),
    cancelJob: (jobId) =>
      Promise.resolve({
        committed: true,
        jobId,
        txHash: 'fixture-cancel',
        code: 0,
        blockHeight: 1,
      }),
  };
  const provider: HPCProviderAdapter = {
    getJobLogs: (_jobId, options) => {
      const tail = options?.tail ?? 100;
      return Promise.resolve({
        lines: MOCK_LOG_LINES.slice(-tail),
        hasMore: MOCK_LOG_LINES.length > tail,
      });
    },
    getJobOutputs: (jobId) =>
      Promise.resolve(
        MOCK_JOBS.find((job) => job.id === jobId)?.status === 'completed' ? MOCK_OUTPUTS : []
      ),
    getJobUsage: (jobId) => {
      const isRunning = MOCK_JOBS.find((job) => job.id === jobId)?.status === 'running';
      return Promise.resolve(
        isRunning
          ? {
              cpuPercent: 72,
              memoryPercent: 58,
              gpuPercent: 85,
              elapsedSeconds: 6000,
              estimatedRemainingSeconds: 80400,
            }
          : {
              cpuPercent: 0,
              memoryPercent: 0,
              elapsedSeconds: 0,
            }
      );
    },
  };

  return createHPCClient({ query, signer, provider });
}
