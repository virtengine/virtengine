import { MsgMultiSend, MsgMultiSendResponse, MsgSend, MsgSendResponse, MsgSetSendEnabled, MsgSetSendEnabledResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.bank.v1beta1.Msg",
  methods: {
    send: {
      name: "Send",
      input: MsgSend,
      output: MsgSendResponse,
      get parent() { return Msg; },
    },
    multiSend: {
      name: "MultiSend",
      input: MsgMultiSend,
      output: MsgMultiSendResponse,
      get parent() { return Msg; },
    },
    updateParams: {
      name: "UpdateParams",
      input: MsgUpdateParams,
      output: MsgUpdateParamsResponse,
      get parent() { return Msg; },
    },
    setSendEnabled: {
      name: "SetSendEnabled",
      input: MsgSetSendEnabled,
      output: MsgSetSendEnabledResponse,
      get parent() { return Msg; },
    },
  },
} as const;
