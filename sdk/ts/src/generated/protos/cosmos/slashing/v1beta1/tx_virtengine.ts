import { MsgUnjail, MsgUnjailResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.slashing.v1beta1.Msg",
  methods: {
    unjail: {
      name: "Unjail",
      input: MsgUnjail,
      output: MsgUnjailResponse,
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
