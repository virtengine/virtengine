import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ClientTxResult, ListOptions } from "./types.ts";

export type JobState =
  | "JOB_STATE_UNSPECIFIED"
  | "JOB_STATE_PENDING"
  | "JOB_STATE_QUEUED"
  | "JOB_STATE_RUNNING"
  | "JOB_STATE_COMPLETED"
  | "JOB_STATE_FAILED"
  | "JOB_STATE_CANCELLED";

export interface HPCCluster {
  clusterId: string;
  providerAddress: string;
  name: string;
  description?: string;
  region: string;
  totalNodes: number;
  availableNodes: number;
  state: string;
}

export interface HPCOffering {
  offeringId: string;
  clusterId: string;
  name: string;
  description?: string;
  pricing: HPCPricing;
  requiredIdentityThreshold: number;
  maxRuntimeSeconds: number;
  active: boolean;
}

export interface HPCPricing {
  cpuCoreHourUakt: string;
  memoryGbHourUakt: string;
  gpuHourUakt: string;
  storageGbHourUakt: string;
}

export interface HPCJob {
  jobId: string;
  offeringId: string;
  customerAddress: string;
  state: JobState;
  createdAt: number;
  startedAt?: number;
  completedAt?: number;
}

export interface SubmitJobParams {
  offeringId: string;
  queueName: string;
  workloadSpec: {
    containerImage?: string;
    script?: string;
    command?: string[];
    environment?: Record<string, string>;
  };
  resources: {
    cpuCores: number;
    memoryMb: number;
    gpus?: number;
    storageMb?: number;
  };
  maxRuntimeSeconds: number;
  maxPrice: Array<{ denom: string; amount: string }>;
}

export interface HPCClientDeps {
  sdk: unknown;
}

/**
 * Client for HPC high-performance computing module
 */
export class HPCClient extends BaseClient {
  private sdk: unknown;

  constructor(deps: HPCClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  // Cluster queries

  /**
   * Get cluster by ID
   */
  async getCluster(_clusterId: string): Promise<HPCCluster | null> {
    try {
      throw new Error("HPC module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "getCluster");
    }
  }

  /**
   * List available clusters
   */
  async listClusters(_options?: ListOptions & { region?: string }): Promise<HPCCluster[]> {
    try {
      throw new Error("HPC module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "listClusters");
    }
  }

  // Offering queries

  /**
   * Get offering by ID
   */
  async getOffering(_offeringId: string): Promise<HPCOffering | null> {
    try {
      throw new Error("HPC module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "getOffering");
    }
  }

  /**
   * List available offerings
   */
  async listOfferings(_options?: ListOptions & { clusterId?: string; activeOnly?: boolean }): Promise<HPCOffering[]> {
    try {
      throw new Error("HPC module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "listOfferings");
    }
  }

  // Job queries

  /**
   * Get job by ID
   */
  async getJob(_jobId: string): Promise<HPCJob | null> {
    try {
      throw new Error("HPC module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "getJob");
    }
  }

  /**
   * List jobs
   */
  async listJobs(_options?: ListOptions & { state?: JobState; customerAddress?: string }): Promise<HPCJob[]> {
    try {
      throw new Error("HPC module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "listJobs");
    }
  }

  // Transaction methods

  /**
   * Submit a new HPC job
   */
  async submitJob(_params: SubmitJobParams): Promise<ClientTxResult & { jobId: string; escrowId: string }> {
    try {
      throw new Error("HPC module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "submitJob");
    }
  }

  /**
   * Cancel a running job
   */
  async cancelJob(_jobId: string, _reason?: string): Promise<ClientTxResult> {
    try {
      throw new Error("HPC module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "cancelJob");
    }
  }

  /**
   * Register a new HPC cluster
   */
  async registerCluster(_params: {
    name: string;
    description?: string;
    region: string;
    totalNodes: number;
    slurmVersion: string;
  }): Promise<ClientTxResult & { clusterId: string }> {
    try {
      throw new Error("HPC module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "registerCluster");
    }
  }
}
