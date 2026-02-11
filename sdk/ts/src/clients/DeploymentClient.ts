import Long from "long";

import type { DeploymentID } from "../generated/protos/virtengine/deployment/v1/deployment.ts";
import type { GroupID } from "../generated/protos/virtengine/deployment/v1/group.ts";
import type {
  MsgCloseDeployment,
  MsgCreateDeployment,
  MsgUpdateDeployment,
} from "../generated/protos/virtengine/deployment/v1beta4/deploymentmsg.ts";
import type { DeploymentFilters } from "../generated/protos/virtengine/deployment/v1beta4/filters.ts";
import type { Group } from "../generated/protos/virtengine/deployment/v1beta4/group.ts";
import type {
  QueryDeploymentResponse,
} from "../generated/protos/virtengine/deployment/v1beta4/query.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";
import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ChainNodeSDK, ClientTxResult, ListOptions } from "./types.ts";
import { toPageRequest, withTxResult } from "./types.ts";

export interface DeploymentClientDeps {
  sdk: ChainNodeSDK;
}

export interface DeploymentListFilters {
  owner?: string;
  state?: string;
}

/**
 * Client for Deployment module
 */
export class DeploymentClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: DeploymentClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  /**
   * Get deployment by ID
   */
  async getDeployment(id: DeploymentID): Promise<QueryDeploymentResponse | null> {
    try {
      const result = await this.sdk.virtengine.deployment.v1beta4.getDeployment({ id });
      return result ?? null;
    } catch (error) {
      this.handleQueryError(error, "getDeployment");
    }
  }

  /**
   * List deployments
   */
  async listDeployments(options?: ListOptions & DeploymentListFilters): Promise<QueryDeploymentResponse[]> {
    try {
      const filters: DeploymentFilters | undefined = options?.owner || options?.state
        ? {
            owner: options?.owner ?? "",
            state: options?.state ?? "",
            dseq: Long.fromValue(0),
          }
        : undefined;

      const result = await this.sdk.virtengine.deployment.v1beta4.getDeployments({
        filters,
        pagination: toPageRequest(options),
      });
      return result.deployments;
    } catch (error) {
      this.handleQueryError(error, "listDeployments");
    }
  }

  /**
   * Get deployment group by ID
   */
  async getDeploymentGroup(id: GroupID): Promise<Group | null> {
    try {
      const result = await this.sdk.virtengine.deployment.v1beta4.getGroup({ id });
      return result.group ?? null;
    } catch (error) {
      this.handleQueryError(error, "getDeploymentGroup");
    }
  }

  /**
   * Create a deployment
   */
  async createDeployment(params: MsgCreateDeployment, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.deployment.v1beta4.createDeployment(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "createDeployment");
    }
  }

  /**
   * Update a deployment
   */
  async updateDeployment(params: MsgUpdateDeployment, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.deployment.v1beta4.updateDeployment(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "updateDeployment");
    }
  }

  /**
   * Close a deployment
   */
  async closeDeployment(params: MsgCloseDeployment, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.deployment.v1beta4.closeDeployment(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "closeDeployment");
    }
  }
}
