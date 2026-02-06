/**
 * HPC Client
 *
 * Wrapper around VirtEngine SDK for HPC operations.
 * For now, provides mock data until SDK integration is complete.
 *
 * TODO: Replace mock implementation with real SDK calls when:
 * - SDK is added to portal dependencies
 * - Wallet connection provides signing capability
 * - Provider network is live
 */

import type {
  Job,
  JobStatus,
  SDKJob,
  SDKOffering,
  SDKWorkloadTemplate,
  WorkloadTemplate,
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

/**
 * HPC Client
 *
 * Provides methods for interacting with the HPC module on chain.
 */
export class HPCClient {
  private config: HPCClientConfig;

  constructor(config: HPCClientConfig = {}) {
    this.config = config;
  }

  /**
   * List available workload templates
   */
  async listWorkloadTemplates(): Promise<WorkloadTemplate[]> {
    // Mock data for now
    await this.delay(300);

    return MOCK_TEMPLATES;
  }

  /**
   * Get workload template by ID
   */
  async getWorkloadTemplate(templateId: string): Promise<WorkloadTemplate | null> {
    await this.delay(200);

    return MOCK_TEMPLATES.find((t) => t.id === templateId) ?? null;
  }

  /**
   * List available offerings
   */
  async listOfferings(): Promise<SDKOffering[]> {
    await this.delay(300);

    return MOCK_OFFERINGS;
  }

  /**
   * Get offering by ID
   */
  async getOffering(offeringId: string): Promise<SDKOffering | null> {
    await this.delay(200);

    return MOCK_OFFERINGS.find((o) => o.offeringId === offeringId) ?? null;
  }

  /**
   * List user's jobs
   */
  async listJobs(filters?: { status?: JobStatus[] }): Promise<Job[]> {
    await this.delay(400);

    let jobs = MOCK_JOBS;

    if (filters?.status && filters.status.length > 0) {
      jobs = jobs.filter((job) => filters.status!.includes(job.status));
    }

    return jobs;
  }

  /**
   * Get job by ID
   */
  async getJob(jobId: string): Promise<Job | null> {
    await this.delay(200);

    return MOCK_JOBS.find((j) => j.id === jobId) ?? null;
  }

  /**
   * Submit a new job
   */
  async submitJob(params: SubmitJobParams): Promise<{ jobId: string; txHash: string }> {
    await this.delay(1000); // Simulate blockchain transaction time

    // Mock response
    const jobId = `job-${Date.now()}`;
    const txHash = `0x${Math.random().toString(16).substring(2, 66)}`;

    return { jobId, txHash };
  }

  /**
   * Cancel a job
   */
  async cancelJob(jobId: string): Promise<{ txHash: string }> {
    await this.delay(800);

    const txHash = `0x${Math.random().toString(16).substring(2, 66)}`;

    return { txHash };
  }

  /**
   * Estimate job cost
   */
  async estimateJobCost(
    offeringId: string,
    resources: SubmitJobParams['resources'],
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
    await this.delay(200);

    // Simple mock calculation
    const basePrice = resources.nodes * resources.cpusPerNode * 0.5;
    const gpuPrice = (resources.gpusPerNode ?? 0) * resources.nodes * 2.5;
    const storagePrice = resources.storageGB * 0.01;

    const hourlyRate = basePrice + gpuPrice + storagePrice;
    const maxHours = resources.maxRuntimeSeconds / 3600;
    const total = hourlyRate * maxHours;

    return {
      estimatedTotal: `${total.toFixed(2)}`,
      pricePerHour: `${hourlyRate.toFixed(2)}`,
      breakdown: {
        compute: `${basePrice.toFixed(2)}`,
        storage: `${storagePrice.toFixed(2)}`,
        network: '0.50',
        gpu: resources.gpusPerNode ? `${gpuPrice.toFixed(2)}` : undefined,
      },
      denom: 'uakt',
    };
  }

  private delay(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

/**
 * Mock Templates
 */
const MOCK_TEMPLATES: WorkloadTemplate[] = [
  {
    id: 'pytorch-training',
    name: 'PyTorch Training',
    description: 'Train deep learning models with PyTorch. Supports distributed training across multiple GPUs.',
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
 * Create HPC client instance
 */
export function createHPCClient(config?: HPCClientConfig): HPCClient {
  return new HPCClient(config);
}
