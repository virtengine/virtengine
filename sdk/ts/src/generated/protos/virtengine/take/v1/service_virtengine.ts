import { MsgUpdateParams, MsgUpdateParamsResponse } from "./paramsmsg.ts";

export const Msg = {
  typeName: "virtengine.take.v1.Msg",
  methods: {
    updateParams: {
      name: "UpdateParams",
      input: MsgUpdateParams,
      output: MsgUpdateParamsResponse,
      get parent() { return Msg; },
    },
  },
} as const;
