import type { HPCDispute } from "../generated/protos/virtengine/hpc/v1/types.ts";
import type {
  MsgActivateEscrow,
  MsgCreateEscrow,
  MsgDisputeEscrow,
  MsgRecordUsage,
  MsgRefundEscrow,
  MsgReleaseEscrow,
  MsgSettleOrder,
} from "../generated/protos/virtengine/settlement/v1/tx.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";
import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type {
  ClaimableRewards,
  EscrowAccount,
  PayoutRecord,
  SettlementRecord,
  UsageRecord,
} from "./settlementTypes.ts";
import type { ChainNodeSDK, ClientTxResult, ListOptions } from "./types.ts";
import { withTxResult } from "./types.ts";

export interface SettlementClientDeps {
  sdk: ChainNodeSDK;
}

export interface SettlementEscrowFilters {
  orderId?: string;
  state?: string;
}

export interface SettlementUsageFilters {
  orderId?: string;
}

export interface SettlementPayoutFilters {
  provider?: string;
}

interface SettlementQuerySDK {
  virtengine: {
    settlement: {
      v1: {
        getEscrow: (request: { escrowId: string }) => Promise<{ escrow?: EscrowAccount }>;
        getEscrowsByOrder: (request: { orderId: string }) => Promise<{ escrows: EscrowAccount[] }>;
        getEscrowsByState: (request: { state: string }) => Promise<{ escrows: EscrowAccount[] }>;
        getSettlement: (request: { settlementId: string }) => Promise<{ settlement?: SettlementRecord }>;
        getSettlementsByOrder: (request: { orderId: string }) => Promise<{ settlements: SettlementRecord[] }>;
        getUsageRecord: (request: { usageId: string }) => Promise<{ usageRecord?: UsageRecord }>;
        getUsageRecordsByOrder: (request: { orderId: string }) => Promise<{ usageRecords: UsageRecord[] }>;
        getPayout: (request: { payoutId: string }) => Promise<{ payout?: PayoutRecord }>;
        getPayoutsByProvider: (request: { provider: string }) => Promise<{ payouts: PayoutRecord[] }>;
        getClaimableRewards: (request: { address: string }) => Promise<{ rewards?: ClaimableRewards }>;
        createEscrow: (request: MsgCreateEscrow, options?: TxCallOptions) => Promise<unknown>;
        activateEscrow: (request: MsgActivateEscrow, options?: TxCallOptions) => Promise<unknown>;
        releaseEscrow: (request: MsgReleaseEscrow, options?: TxCallOptions) => Promise<unknown>;
        refundEscrow: (request: MsgRefundEscrow, options?: TxCallOptions) => Promise<unknown>;
        recordUsage: (request: MsgRecordUsage, options?: TxCallOptions) => Promise<unknown>;
        settleOrder: (request: MsgSettleOrder, options?: TxCallOptions) => Promise<unknown>;
        disputeEscrow: (request: MsgDisputeEscrow, options?: TxCallOptions) => Promise<unknown>;
      };
    };
  };
}

interface HpcDisputeSDK {
  virtengine: {
    hpc: {
      v1: {
        getDispute: (request: { disputeId: string }) => Promise<{ dispute?: HPCDispute }>;
      };
    };
  };
}

/**
 * Client for Settlement module (escrow settlement, usage, payouts)
 */
export class SettlementClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: SettlementClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  private get settlementSdk() {
    return (this.sdk as unknown as SettlementQuerySDK).virtengine.settlement.v1;
  }

  private get hpcSdk() {
    return (this.sdk as unknown as HpcDisputeSDK).virtengine.hpc.v1;
  }

  /**
   * Get escrow by ID
   */
  async getEscrow(escrowId: string): Promise<EscrowAccount | null> {
    try {
      const result = await this.settlementSdk.getEscrow({ escrowId });
      return result.escrow ?? null;
    } catch (error) {
      this.handleQueryError(error, "getEscrow");
    }
  }

  /**
   * List escrows by order or state
   */
  async listEscrows(options?: ListOptions & SettlementEscrowFilters): Promise<EscrowAccount[]> {
    try {
      if (options?.orderId) {
        const result = await this.settlementSdk.getEscrowsByOrder({ orderId: options.orderId });
        return result.escrows;
      }

      if (options?.state) {
        const result = await this.settlementSdk.getEscrowsByState({ state: options.state });
        return result.escrows;
      }

      return [];
    } catch (error) {
      this.handleQueryError(error, "listEscrows");
    }
  }

  /**
   * Get settlement by ID
   */
  async getSettlement(settlementId: string): Promise<SettlementRecord | null> {
    try {
      const result = await this.settlementSdk.getSettlement({ settlementId });
      return result.settlement ?? null;
    } catch (error) {
      this.handleQueryError(error, "getSettlement");
    }
  }

  /**
   * List payouts for a provider
   */
  async listPayouts(options?: SettlementPayoutFilters): Promise<PayoutRecord[]> {
    try {
      if (!options?.provider) return [];
      const result = await this.settlementSdk.getPayoutsByProvider({ provider: options.provider });
      return result.payouts;
    } catch (error) {
      this.handleQueryError(error, "listPayouts");
    }
  }

  /**
   * Get usage record by ID
   */
  async getUsageRecord(usageId: string): Promise<UsageRecord | null> {
    try {
      const result = await this.settlementSdk.getUsageRecord({ usageId });
      return result.usageRecord ?? null;
    } catch (error) {
      this.handleQueryError(error, "getUsageRecord");
    }
  }

  /**
   * List usage records by order
   */
  async listUsageRecords(options?: SettlementUsageFilters): Promise<UsageRecord[]> {
    try {
      if (!options?.orderId) return [];
      const result = await this.settlementSdk.getUsageRecordsByOrder({ orderId: options.orderId });
      return result.usageRecords;
    } catch (error) {
      this.handleQueryError(error, "listUsageRecords");
    }
  }

  /**
   * Get dispute record by ID (from HPC disputes)
   */
  async getDispute(disputeId: string): Promise<HPCDispute | null> {
    try {
      const result = await this.hpcSdk.getDispute({ disputeId });
      return result.dispute ?? null;
    } catch (error) {
      this.handleQueryError(error, "getDispute");
    }
  }

  /**
   * Estimate claimable rewards for an address
   */
  async estimateRewards(address: string): Promise<ClaimableRewards | null> {
    try {
      const result = await this.settlementSdk.getClaimableRewards({ address });
      return result.rewards ?? null;
    } catch (error) {
      this.handleQueryError(error, "estimateRewards");
    }
  }

  /**
   * Create a new escrow account
   */
  async createEscrow(params: MsgCreateEscrow, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.settlementSdk.createEscrow(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "createEscrow");
    }
  }

  /**
   * Activate escrow for a lease
   */
  async activateEscrow(params: MsgActivateEscrow, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.settlementSdk.activateEscrow(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "activateEscrow");
    }
  }

  /**
   * Release escrow funds
   */
  async releaseEscrow(params: MsgReleaseEscrow, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.settlementSdk.releaseEscrow(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "releaseEscrow");
    }
  }

  /**
   * Refund escrow funds
   */
  async refundEscrow(params: MsgRefundEscrow, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.settlementSdk.refundEscrow(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "refundEscrow");
    }
  }

  /**
   * Record usage for a lease
   */
  async recordUsage(params: MsgRecordUsage, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.settlementSdk.recordUsage(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "recordUsage");
    }
  }

  /**
   * Settle an order
   */
  async settleOrder(params: MsgSettleOrder, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.settlementSdk.settleOrder(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "settleOrder");
    }
  }

  /**
   * Open a dispute on escrow
   */
  async openDispute(params: MsgDisputeEscrow, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.settlementSdk.disputeEscrow(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "openDispute");
    }
  }

  /**
   * Resolve a dispute by releasing escrow
   */
  async resolveDispute(params: MsgReleaseEscrow, options?: TxCallOptions): Promise<ClientTxResult> {
    return this.releaseEscrow(params, options);
  }

  /**
   * Issue a refund from escrow
   */
  async issueRefund(params: MsgRefundEscrow, options?: TxCallOptions): Promise<ClientTxResult> {
    return this.refundEscrow(params, options);
  }
}
