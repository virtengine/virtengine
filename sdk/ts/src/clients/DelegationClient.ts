import type {
  MsgClaimRewards,
  MsgDelegate,
  MsgRedelegate,
  MsgUndelegate,
} from "../generated/protos/virtengine/delegation/v1/tx.ts";
import type {
  Delegation,
  DelegatorReward,
} from "../generated/protos/virtengine/delegation/v1/types.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";
import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ChainNodeSDK, ClientTxResult, ListOptions } from "./types.ts";
import { toPageRequest, withTxResult } from "./types.ts";

export interface DelegationClientDeps {
  sdk: ChainNodeSDK;
}

export interface DelegationFilters {
  delegatorAddress?: string;
  validatorAddress?: string;
}

/**
 * Client for Delegation module
 */
export class DelegationClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: DelegationClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  /**
   * Get a delegation by delegator and validator
   */
  async getDelegation(delegatorAddress: string, validatorAddress: string): Promise<Delegation | null> {
    try {
      const result = await this.sdk.virtengine.delegation.v1.getDelegation({
        delegatorAddress,
        validatorAddress,
      });
      return result.delegation ?? null;
    } catch (error) {
      this.handleQueryError(error, "getDelegation");
    }
  }

  /**
   * List delegations for a delegator or validator
   */
  async listDelegations(options?: ListOptions & DelegationFilters): Promise<Delegation[]> {
    try {
      if (options?.delegatorAddress) {
        const result = await this.sdk.virtengine.delegation.v1.getDelegatorDelegations({
          delegatorAddress: options.delegatorAddress,
          pagination: toPageRequest(options),
        });
        return result.delegations;
      }

      if (options?.validatorAddress) {
        const result = await this.sdk.virtengine.delegation.v1.getValidatorDelegations({
          validatorAddress: options.validatorAddress,
          pagination: toPageRequest(options),
        });
        return result.delegations;
      }

      return [];
    } catch (error) {
      this.handleQueryError(error, "listDelegations");
    }
  }

  /**
   * Get delegator rewards (optionally scoped to validator)
   */
  async getDelegatorRewards(
    delegatorAddress: string,
    validatorAddress?: string,
  ): Promise<DelegatorReward[]> {
    try {
      if (validatorAddress) {
        const result = await this.sdk.virtengine.delegation.v1.getDelegatorRewards({
          delegatorAddress,
          validatorAddress,
        });
        return result.rewards;
      }

      const result = await this.sdk.virtengine.delegation.v1.getDelegatorAllRewards({
        delegatorAddress,
        pagination: toPageRequest({ limit: 100 }),
      });
      return result.rewards;
    } catch (error) {
      this.handleQueryError(error, "getDelegatorRewards");
    }
  }

  /**
   * Get validator delegations
   */
  async getValidatorDelegators(
    validatorAddress: string,
    options?: ListOptions,
  ): Promise<Delegation[]> {
    try {
      const result = await this.sdk.virtengine.delegation.v1.getValidatorDelegations({
        validatorAddress,
        pagination: toPageRequest(options),
      });
      return result.delegations;
    } catch (error) {
      this.handleQueryError(error, "getValidatorDelegators");
    }
  }

  /**
   * Delegate tokens to a validator
   */
  async delegate(params: MsgDelegate, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.delegation.v1.delegate(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "delegate");
    }
  }

  /**
   * Undelegate tokens from a validator
   */
  async undelegate(params: MsgUndelegate, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.delegation.v1.undelegate(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "undelegate");
    }
  }

  /**
   * Redelegate tokens between validators
   */
  async redelegate(params: MsgRedelegate, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.delegation.v1.redelegate(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "redelegate");
    }
  }

  /**
   * Claim rewards from a validator
   */
  async claimRewards(params: MsgClaimRewards, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.delegation.v1.claimRewards(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "claimRewards");
    }
  }
}
