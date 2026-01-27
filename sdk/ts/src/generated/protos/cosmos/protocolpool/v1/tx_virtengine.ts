import { MsgCancelContinuousFund, MsgCancelContinuousFundResponse, MsgCommunityPoolSpend, MsgCommunityPoolSpendResponse, MsgCreateContinuousFund, MsgCreateContinuousFundResponse, MsgFundCommunityPool, MsgFundCommunityPoolResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.protocolpool.v1.Msg",
  methods: {
    fundCommunityPool: {
      name: "FundCommunityPool",
      input: MsgFundCommunityPool,
      output: MsgFundCommunityPoolResponse,
      get parent() { return Msg; },
    },
    communityPoolSpend: {
      name: "CommunityPoolSpend",
      input: MsgCommunityPoolSpend,
      output: MsgCommunityPoolSpendResponse,
      get parent() { return Msg; },
    },
    createContinuousFund: {
      name: "CreateContinuousFund",
      input: MsgCreateContinuousFund,
      output: MsgCreateContinuousFundResponse,
      get parent() { return Msg; },
    },
    cancelContinuousFund: {
      name: "CancelContinuousFund",
      input: MsgCancelContinuousFund,
      output: MsgCancelContinuousFundResponse,
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
