import { MsgBeginRedelegate, MsgBeginRedelegateResponse, MsgCancelUnbondingDelegation, MsgCancelUnbondingDelegationResponse, MsgCreateValidator, MsgCreateValidatorResponse, MsgDelegate, MsgDelegateResponse, MsgEditValidator, MsgEditValidatorResponse, MsgUndelegate, MsgUndelegateResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.staking.v1beta1.Msg",
  methods: {
    createValidator: {
      name: "CreateValidator",
      input: MsgCreateValidator,
      output: MsgCreateValidatorResponse,
      get parent() { return Msg; },
    },
    editValidator: {
      name: "EditValidator",
      input: MsgEditValidator,
      output: MsgEditValidatorResponse,
      get parent() { return Msg; },
    },
    delegate: {
      name: "Delegate",
      input: MsgDelegate,
      output: MsgDelegateResponse,
      get parent() { return Msg; },
    },
    beginRedelegate: {
      name: "BeginRedelegate",
      input: MsgBeginRedelegate,
      output: MsgBeginRedelegateResponse,
      get parent() { return Msg; },
    },
    undelegate: {
      name: "Undelegate",
      input: MsgUndelegate,
      output: MsgUndelegateResponse,
      get parent() { return Msg; },
    },
    cancelUnbondingDelegation: {
      name: "CancelUnbondingDelegation",
      input: MsgCancelUnbondingDelegation,
      output: MsgCancelUnbondingDelegationResponse,
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
