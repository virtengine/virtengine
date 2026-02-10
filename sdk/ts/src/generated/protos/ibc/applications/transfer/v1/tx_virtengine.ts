import { MsgTransfer, MsgTransferResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

export const Msg = {
  typeName: "ibc.applications.transfer.v1.Msg",
  methods: {
    transfer: {
      name: "Transfer",
      input: MsgTransfer,
      output: MsgTransferResponse,
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
