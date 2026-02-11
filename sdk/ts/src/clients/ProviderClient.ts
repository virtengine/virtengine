import type { HPCCluster, HPCJob } from "../generated/protos/virtengine/hpc/v1/types.ts";
import type {
  MsgCreateProvider,
  MsgDeleteProvider,
  MsgUpdateProvider,
} from "../generated/protos/virtengine/provider/v1beta4/msg.ts";
import type { Provider } from "../generated/protos/virtengine/provider/v1beta4/provider.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";
import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ChainNodeSDK, ClientTxResult, ListOptions } from "./types.ts";
import { toPageRequest, withTxResult } from "./types.ts";

export interface ProviderClientDeps {
  sdk: ChainNodeSDK;
}

export interface ProviderCapacityFilters {
  providerAddress: string;
}

/**
 * Client for Provider module
 */
export class ProviderClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: ProviderClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  /**
   * Get provider by address
   */
  async getProvider(address: string): Promise<Provider | null> {
    try {
      const result = await this.sdk.virtengine.provider.v1beta4.getProvider({ owner: address });
      return result.provider ?? null;
    } catch (error) {
      this.handleQueryError(error, "getProvider");
    }
  }

  /**
   * List providers
   */
  async listProviders(options?: ListOptions): Promise<Provider[]> {
    try {
      const result = await this.sdk.virtengine.provider.v1beta4.getProviders({
        pagination: toPageRequest(options),
      });
      return result.providers;
    } catch (error) {
      this.handleQueryError(error, "listProviders");
    }
  }

  /**
   * Get provider capacity (clusters) from HPC module
   */
  async getProviderCapacity(filters: ProviderCapacityFilters): Promise<HPCCluster[]> {
    try {
      const result = await this.sdk.virtengine.hpc.v1.getClustersByProvider({
        providerAddress: filters.providerAddress,
        pagination: toPageRequest({ limit: 100 }),
      });
      return result.clusters;
    } catch (error) {
      this.handleQueryError(error, "getProviderCapacity");
    }
  }

  /**
   * Get provider orders (jobs) from HPC module
   */
  async getProviderOrders(providerAddress: string, options?: ListOptions): Promise<HPCJob[]> {
    try {
      const result = await this.sdk.virtengine.hpc.v1.getJobsByProvider({
        providerAddress,
        pagination: toPageRequest(options),
      });
      return result.jobs;
    } catch (error) {
      this.handleQueryError(error, "getProviderOrders");
    }
  }

  /**
   * Register a provider
   */
  async registerProvider(params: MsgCreateProvider, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.provider.v1beta4.createProvider(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "registerProvider");
    }
  }

  /**
   * Update provider information
   */
  async updateProvider(params: MsgUpdateProvider, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.provider.v1beta4.updateProvider(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "updateProvider");
    }
  }

  /**
   * Deactivate provider
   */
  async deactivateProvider(params: MsgDeleteProvider, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.provider.v1beta4.deleteProvider(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "deactivateProvider");
    }
  }
}
