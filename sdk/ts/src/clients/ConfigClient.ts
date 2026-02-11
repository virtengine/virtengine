import type {
  MsgReactivateApprovedClient,
  MsgRegisterApprovedClient,
  MsgRevokeApprovedClient,
  MsgSuspendApprovedClient,
  MsgUpdateApprovedClient,
} from "../generated/protos/virtengine/config/v1/tx.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";
import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ChainNodeSDK, ClientTxResult } from "./types.ts";
import { withTxResult } from "./types.ts";

export interface ConfigClientDeps {
  sdk: ChainNodeSDK;
}

/**
 * Client for Config module
 */
export class ConfigClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: ConfigClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  async registerApprovedClient(
    params: MsgRegisterApprovedClient,
    options?: TxCallOptions,
  ): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.config.v1.registerApprovedClient(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "registerApprovedClient");
    }
  }

  async updateApprovedClient(
    params: MsgUpdateApprovedClient,
    options?: TxCallOptions,
  ): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.config.v1.updateApprovedClient(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "updateApprovedClient");
    }
  }

  async suspendApprovedClient(
    params: MsgSuspendApprovedClient,
    options?: TxCallOptions,
  ): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.config.v1.suspendApprovedClient(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "suspendApprovedClient");
    }
  }

  async revokeApprovedClient(
    params: MsgRevokeApprovedClient,
    options?: TxCallOptions,
  ): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.config.v1.revokeApprovedClient(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "revokeApprovedClient");
    }
  }

  async reactivateApprovedClient(
    params: MsgReactivateApprovedClient,
    options?: TxCallOptions,
  ): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.config.v1.reactivateApprovedClient(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "reactivateApprovedClient");
    }
  }
}
