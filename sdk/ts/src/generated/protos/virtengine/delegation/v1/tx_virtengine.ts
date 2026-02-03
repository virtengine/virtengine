import { MsgClaimAllRewards, MsgClaimAllRewardsResponse, MsgClaimRewards, MsgClaimRewardsResponse, MsgDelegate, MsgDelegateResponse, MsgRedelegate, MsgRedelegateResponse, MsgUndelegate, MsgUndelegateResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

export const Msg = {
  typeName: "virtengine.delegation.v1.Msg",
  methods: {
    delegate: {
      name: "Delegate",
      input: MsgDelegate,
      output: MsgDelegateResponse,
      get parent() { return Msg; },
    },
    undelegate: {
      name: "Undelegate",
      input: MsgUndelegate,
      output: MsgUndelegateResponse,
      get parent() { return Msg; },
    },
    redelegate: {
      name: "Redelegate",
      input: MsgRedelegate,
      output: MsgRedelegateResponse,
      get parent() { return Msg; },
    },
    claimRewards: {
      name: "ClaimRewards",
      input: MsgClaimRewards,
      output: MsgClaimRewardsResponse,
      get parent() { return Msg; },
    },
    claimAllRewards: {
      name: "ClaimAllRewards",
      input: MsgClaimAllRewards,
      output: MsgClaimAllRewardsResponse,
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
