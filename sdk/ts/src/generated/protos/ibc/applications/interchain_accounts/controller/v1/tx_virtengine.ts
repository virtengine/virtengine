import { MsgRegisterInterchainAccount, MsgRegisterInterchainAccountResponse, MsgSendTx, MsgSendTxResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

export const Msg = {
  typeName: "ibc.applications.interchain_accounts.controller.v1.Msg",
  methods: {
    registerInterchainAccount: {
      name: "RegisterInterchainAccount",
      input: MsgRegisterInterchainAccount,
      output: MsgRegisterInterchainAccountResponse,
      get parent() { return Msg; },
    },
    sendTx: {
      name: "SendTx",
      input: MsgSendTx,
      output: MsgSendTxResponse,
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
