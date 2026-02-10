import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ChainNodeSDK, ClientTxResult, ListOptions } from "./types.ts";
import { toPageRequest, withTxResult } from "./types.ts";
import type {
  ClusterState,
  HPCCluster,
  HPCJob,
  HPCOffering,
  JobState,
} from "../generated/protos/virtengine/hpc/v1/types.ts";
import type {
  MsgCancelJob,
  MsgRegisterCluster,
  MsgSubmitJob,
} from "../generated/protos/virtengine/hpc/v1/tx.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";

export interface HPCClientDeps {
  sdk: ChainNodeSDK;
}

export interface HPCClusterFilters {
  region?: string;
  state?: ClusterState;
}

export interface HPCOfferingFilters {
  clusterId?: string;
  activeOnly?: boolean;
}

export interface HPCJobFilters {
  state?: JobState;
  clusterId?: string;
  customerAddress?: string;
  providerAddress?: string;
}

/**
 * Client for HPC high-performance computing module
 */
export class HPCClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: HPCClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  // Cluster queries

  /**
   * Get cluster by ID
   */
  async getCluster(clusterId: string): Promise<HPCCluster | null> {
    const cacheKey = `hpc:cluster:${clusterId}`;
    const cached = this.getCached<HPCCluster>(cacheKey);
    if (cached) return cached;

    try {
      const result = await this.sdk.virtengine.hpc.v1.getCluster({ clusterId });
      this.setCached(cacheKey, result.cluster);
      return result.cluster ?? null;
    } catch (error) {
      this.handleQueryError(error, "getCluster");
    }
  }

  /**
   * List available clusters
   */
  async listClusters(options?: ListOptions & HPCClusterFilters): Promise<HPCCluster[]> {
    const cacheKey = `hpc:clusters:${options?.region ?? ""}:${options?.state ?? ""}:${options?.limit ?? ""}:${options?.offset ?? ""}:${options?.cursor ?? ""}`;
    const cached = this.getCached<HPCCluster[]>(cacheKey);
    if (cached) return cached;

    try {
      const result = await this.sdk.virtengine.hpc.v1.getClusters({
        region: options?.region ?? "",
        state: options?.state ?? 0,
        pagination: toPageRequest(options),
      });
      this.setCached(cacheKey, result.clusters);
      return result.clusters;
    } catch (error) {
      this.handleQueryError(error, "listClusters");
    }
  }

  // Offering queries

  /**
   * Get offering by ID
   */
  async getOffering(offeringId: string): Promise<HPCOffering | null> {
    const cacheKey = `hpc:offering:${offeringId}`;
    const cached = this.getCached<HPCOffering>(cacheKey);
    if (cached) return cached;

    try {
      const result = await this.sdk.virtengine.hpc.v1.getOffering({ offeringId });
      this.setCached(cacheKey, result.offering);
      return result.offering ?? null;
    } catch (error) {
      this.handleQueryError(error, "getOffering");
    }
  }

  /**
   * List available offerings
   */
  async listOfferings(options?: ListOptions & HPCOfferingFilters): Promise<HPCOffering[]> {
    const cacheKey = `hpc:offerings:${options?.clusterId ?? ""}:${options?.activeOnly ?? ""}:${options?.limit ?? ""}:${options?.offset ?? ""}:${options?.cursor ?? ""}`;
    const cached = this.getCached<HPCOffering[]>(cacheKey);
    if (cached) return cached;

    try {
      if (options?.clusterId) {
        const result = await this.sdk.virtengine.hpc.v1.getOfferingsByCluster({
          clusterId: options.clusterId,
          pagination: toPageRequest(options),
        });
        this.setCached(cacheKey, result.offerings);
        return result.offerings;
      }

      const result = await this.sdk.virtengine.hpc.v1.getOfferings({
        activeOnly: options?.activeOnly ?? false,
        pagination: toPageRequest(options),
      });
      this.setCached(cacheKey, result.offerings);
      return result.offerings;
    } catch (error) {
      this.handleQueryError(error, "listOfferings");
    }
  }

  // Job queries

  /**
   * Get job by ID
   */
  async getJob(jobId: string): Promise<HPCJob | null> {
    const cacheKey = `hpc:job:${jobId}`;
    const cached = this.getCached<HPCJob>(cacheKey);
    if (cached) return cached;

    try {
      const result = await this.sdk.virtengine.hpc.v1.getJob({ jobId });
      this.setCached(cacheKey, result.job);
      return result.job ?? null;
    } catch (error) {
      this.handleQueryError(error, "getJob");
    }
  }

  /**
   * List jobs
   */
  async listJobs(options?: ListOptions & HPCJobFilters): Promise<HPCJob[]> {
    const cacheKey = `hpc:jobs:${options?.state ?? ""}:${options?.clusterId ?? ""}:${options?.customerAddress ?? ""}:${options?.providerAddress ?? ""}:${options?.limit ?? ""}:${options?.offset ?? ""}:${options?.cursor ?? ""}`;
    const cached = this.getCached<HPCJob[]>(cacheKey);
    if (cached) return cached;

    try {
      if (options?.customerAddress) {
        const result = await this.sdk.virtengine.hpc.v1.getJobsByCustomer({
          customerAddress: options.customerAddress,
          pagination: toPageRequest(options),
        });
        this.setCached(cacheKey, result.jobs);
        return result.jobs;
      }

      if (options?.providerAddress) {
        const result = await this.sdk.virtengine.hpc.v1.getJobsByProvider({
          providerAddress: options.providerAddress,
          pagination: toPageRequest(options),
        });
        this.setCached(cacheKey, result.jobs);
        return result.jobs;
      }

      const result = await this.sdk.virtengine.hpc.v1.getJobs({
        state: options?.state ?? 0,
        clusterId: options?.clusterId ?? "",
        pagination: toPageRequest(options),
      });
      this.setCached(cacheKey, result.jobs);
      return result.jobs;
    } catch (error) {
      this.handleQueryError(error, "listJobs");
    }
  }

  // Transaction methods

  /**
   * Submit a new HPC job
   */
  async submitJob(
    params: MsgSubmitJob,
    options?: TxCallOptions,
  ): Promise<ClientTxResult & { jobId: string; escrowId: string }> {
    try {
      const { response, txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.hpc.v1.submitJob(params, txOptions), options);

      return {
        ...txResult,
        jobId: response.jobId,
        escrowId: response.escrowId,
      };
    } catch (error) {
      this.handleQueryError(error, "submitJob");
    }
  }

  /**
   * Cancel a running job
   */
  async cancelJob(params: MsgCancelJob, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.hpc.v1.cancelJob(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "cancelJob");
    }
  }

  /**
   * Register a new HPC cluster
   */
  async registerCluster(
    params: MsgRegisterCluster,
    options?: TxCallOptions,
  ): Promise<ClientTxResult & { clusterId: string }> {
    try {
      const { response, txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.hpc.v1.registerCluster(params, txOptions), options);

      return {
        ...txResult,
        clusterId: response.clusterId,
      };
    } catch (error) {
      this.handleQueryError(error, "registerCluster");
    }
  }
}
