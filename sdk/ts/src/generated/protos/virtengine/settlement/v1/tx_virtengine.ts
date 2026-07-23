import { MsgAcknowledgeUsage, MsgAcknowledgeUsageResponse, MsgActivateEscrow, MsgActivateEscrowResponse, MsgAddFinancialClaim, MsgAddFinancialClaimResponse, MsgAppealFinancialCase, MsgAppealFinancialCaseResponse, MsgCancelFinancialCase, MsgCancelFinancialCaseResponse, MsgClaimRewards, MsgClaimRewardsResponse, MsgCreateEscrow, MsgCreateEscrowResponse, MsgDisputeEscrow, MsgDisputeEscrowResponse, MsgEscalateFinancialCase, MsgEscalateFinancialCaseResponse, MsgFinalizeFinancialCase, MsgFinalizeFinancialCaseResponse, MsgOpenFinancialCase, MsgOpenFinancialCaseResponse, MsgRecordFiatConversionObservation, MsgRecordFiatConversionObservationResponse, MsgRecordUsage, MsgRecordUsageResponse, MsgRefundEscrow, MsgRefundEscrowResponse, MsgReleaseEscrow, MsgReleaseEscrowResponse, MsgResolveFinancialCase, MsgResolveFinancialCaseResponse, MsgSettleOrder, MsgSettleOrderResponse, MsgSubmitFinancialCaseForReview, MsgSubmitFinancialCaseForReviewResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

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
    openFinancialCase: {
      name: "OpenFinancialCase",
      input: MsgOpenFinancialCase,
      output: MsgOpenFinancialCaseResponse,
      get parent() { return Msg; },
    },
    addFinancialClaim: {
      name: "AddFinancialClaim",
      input: MsgAddFinancialClaim,
      output: MsgAddFinancialClaimResponse,
      get parent() { return Msg; },
    },
    submitFinancialCaseForReview: {
      name: "SubmitFinancialCaseForReview",
      input: MsgSubmitFinancialCaseForReview,
      output: MsgSubmitFinancialCaseForReviewResponse,
      get parent() { return Msg; },
    },
    escalateFinancialCase: {
      name: "EscalateFinancialCase",
      input: MsgEscalateFinancialCase,
      output: MsgEscalateFinancialCaseResponse,
      get parent() { return Msg; },
    },
    resolveFinancialCase: {
      name: "ResolveFinancialCase",
      input: MsgResolveFinancialCase,
      output: MsgResolveFinancialCaseResponse,
      get parent() { return Msg; },
    },
    appealFinancialCase: {
      name: "AppealFinancialCase",
      input: MsgAppealFinancialCase,
      output: MsgAppealFinancialCaseResponse,
      get parent() { return Msg; },
    },
    cancelFinancialCase: {
      name: "CancelFinancialCase",
      input: MsgCancelFinancialCase,
      output: MsgCancelFinancialCaseResponse,
      get parent() { return Msg; },
    },
    finalizeFinancialCase: {
      name: "FinalizeFinancialCase",
      input: MsgFinalizeFinancialCase,
      output: MsgFinalizeFinancialCaseResponse,
      get parent() { return Msg; },
    },
    recordFiatConversionObservation: {
      name: "RecordFiatConversionObservation",
      input: MsgRecordFiatConversionObservation,
      output: MsgRecordFiatConversionObservationResponse,
      get parent() { return Msg; },
    },
    updateParams: {
      name: "UpdateParams",
      input: MsgUpdateParams,
      output: MsgUpdateParamsResponse,
      get parent() { return Msg; },
    },
  },
} as const;
