import { MsgConnectionOpenAck, MsgConnectionOpenAckResponse, MsgConnectionOpenConfirm, MsgConnectionOpenConfirmResponse, MsgConnectionOpenInit, MsgConnectionOpenInitResponse, MsgConnectionOpenTry, MsgConnectionOpenTryResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

export const Msg = {
  typeName: "ibc.core.connection.v1.Msg",
  methods: {
    connectionOpenInit: {
      name: "ConnectionOpenInit",
      input: MsgConnectionOpenInit,
      output: MsgConnectionOpenInitResponse,
      get parent() { return Msg; },
    },
    connectionOpenTry: {
      name: "ConnectionOpenTry",
      input: MsgConnectionOpenTry,
      output: MsgConnectionOpenTryResponse,
      get parent() { return Msg; },
    },
    connectionOpenAck: {
      name: "ConnectionOpenAck",
      input: MsgConnectionOpenAck,
      output: MsgConnectionOpenAckResponse,
      get parent() { return Msg; },
    },
    connectionOpenConfirm: {
      name: "ConnectionOpenConfirm",
      input: MsgConnectionOpenConfirm,
      output: MsgConnectionOpenConfirmResponse,
      get parent() { return Msg; },
    },
    updateConnectionParams: {
      name: "UpdateConnectionParams",
      input: MsgUpdateParams,
      output: MsgUpdateParamsResponse,
      get parent() { return Msg; },
    },
  },
} as const;
