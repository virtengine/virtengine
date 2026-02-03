import { MsgAcknowledgeUsage, MsgAcknowledgeUsageResponse, MsgActivateEscrow, MsgActivateEscrowResponse, MsgClaimRewards, MsgClaimRewardsResponse, MsgCreateEscrow, MsgCreateEscrowResponse, MsgDisputeEscrow, MsgDisputeEscrowResponse, MsgRecordUsage, MsgRecordUsageResponse, MsgRefundEscrow, MsgRefundEscrowResponse, MsgReleaseEscrow, MsgReleaseEscrowResponse, MsgSettleOrder, MsgSettleOrderResponse } from "./tx.ts";

export const Msg = {
  typeName: "virtengine.settlement.v1.Msg",
  methods: {
    createEscrow: {
      name: "CreateEscrow",
      input: MsgCreateEscrow,
      output: MsgCreateEscrowResponse,
      get parent() { return Msg; },
    },
    activateEscrow: {
      name: "ActivateEscrow",
      input: MsgActivateEscrow,
      output: MsgActivateEscrowResponse,
      get parent() { return Msg; },
    },
    releaseEscrow: {
      name: "ReleaseEscrow",
      input: MsgReleaseEscrow,
      output: MsgReleaseEscrowResponse,
      get parent() { return Msg; },
    },
    refundEscrow: {
      name: "RefundEscrow",
      input: MsgRefundEscrow,
      output: MsgRefundEscrowResponse,
      get parent() { return Msg; },
    },
    disputeEscrow: {
      name: "DisputeEscrow",
      input: MsgDisputeEscrow,
      output: MsgDisputeEscrowResponse,
      get parent() { return Msg; },
    },
    settleOrder: {
      name: "SettleOrder",
      input: MsgSettleOrder,
      output: MsgSettleOrderResponse,
      get parent() { return Msg; },
    },
    recordUsage: {
      name: "RecordUsage",
      input: MsgRecordUsage,
      output: MsgRecordUsageResponse,
      get parent() { return Msg; },
    },
    acknowledgeUsage: {
      name: "AcknowledgeUsage",
      input: MsgAcknowledgeUsage,
      output: MsgAcknowledgeUsageResponse,
      get parent() { return Msg; },
    },
    claimRewards: {
      name: "ClaimRewards",
      input: MsgClaimRewards,
      output: MsgClaimRewardsResponse,
      get parent() { return Msg; },
    },
  },
} as const;
