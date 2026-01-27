import { MsgUpdateParams, MsgUpdateParamsResponse, MsgVerifyInvariant, MsgVerifyInvariantResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.crisis.v1beta1.Msg",
  methods: {
    verifyInvariant: {
      name: "VerifyInvariant",
      input: MsgVerifyInvariant,
      output: MsgVerifyInvariantResponse,
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
