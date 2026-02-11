import type {
  MsgRecordPerformance,
  MsgSlashValidator,
  MsgUnjailValidator,
} from "../generated/protos/virtengine/staking/v1/tx.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";
import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ChainNodeSDK, ClientTxResult } from "./types.ts";
import { withTxResult } from "./types.ts";

export interface VirtStakingClientDeps {
  sdk: ChainNodeSDK;
}

/**
 * Client for VirtEngine staking module
 */
export class VirtStakingClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: VirtStakingClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  async recordPerformance(params: MsgRecordPerformance, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.staking.v1.recordPerformance(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "recordPerformance");
    }
  }

  async slashValidator(params: MsgSlashValidator, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.staking.v1.slashValidator(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "slashValidator");
    }
  }

  async unjailValidator(params: MsgUnjailValidator, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.staking.v1.unjailValidator(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "unjailValidator");
    }
  }
}
