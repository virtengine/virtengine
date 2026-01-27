import { MsgCommunityPoolSpend, MsgCommunityPoolSpendResponse, MsgDepositValidatorRewardsPool, MsgDepositValidatorRewardsPoolResponse, MsgFundCommunityPool, MsgFundCommunityPoolResponse, MsgSetWithdrawAddress, MsgSetWithdrawAddressResponse, MsgUpdateParams, MsgUpdateParamsResponse, MsgWithdrawDelegatorReward, MsgWithdrawDelegatorRewardResponse, MsgWithdrawValidatorCommission, MsgWithdrawValidatorCommissionResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.distribution.v1beta1.Msg",
  methods: {
    setWithdrawAddress: {
      name: "SetWithdrawAddress",
      input: MsgSetWithdrawAddress,
      output: MsgSetWithdrawAddressResponse,
      get parent() { return Msg; },
    },
    withdrawDelegatorReward: {
      name: "WithdrawDelegatorReward",
      input: MsgWithdrawDelegatorReward,
      output: MsgWithdrawDelegatorRewardResponse,
      get parent() { return Msg; },
    },
    withdrawValidatorCommission: {
      name: "WithdrawValidatorCommission",
      input: MsgWithdrawValidatorCommission,
      output: MsgWithdrawValidatorCommissionResponse,
      get parent() { return Msg; },
    },
    fundCommunityPool: {
      name: "FundCommunityPool",
      input: MsgFundCommunityPool,
      output: MsgFundCommunityPoolResponse,
      get parent() { return Msg; },
    },
    updateParams: {
      name: "UpdateParams",
      input: MsgUpdateParams,
      output: MsgUpdateParamsResponse,
      get parent() { return Msg; },
    },
    communityPoolSpend: {
      name: "CommunityPoolSpend",
      input: MsgCommunityPoolSpend,
      output: MsgCommunityPoolSpendResponse,
      get parent() { return Msg; },
    },
    depositValidatorRewardsPool: {
      name: "DepositValidatorRewardsPool",
      input: MsgDepositValidatorRewardsPool,
      output: MsgDepositValidatorRewardsPoolResponse,
      get parent() { return Msg; },
    },
  },
} as const;
