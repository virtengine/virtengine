import { MsgRegisterCounterparty, MsgRegisterCounterpartyResponse, MsgUpdateClientConfig, MsgUpdateClientConfigResponse } from "./tx.ts";

export const Msg = {
  typeName: "ibc.core.client.v2.Msg",
  methods: {
    registerCounterparty: {
      name: "RegisterCounterparty",
      input: MsgRegisterCounterparty,
      output: MsgRegisterCounterpartyResponse,
      get parent() { return Msg; },
    },
    updateClientConfig: {
      name: "UpdateClientConfig",
      input: MsgUpdateClientConfig,
      output: MsgUpdateClientConfigResponse,
      get parent() { return Msg; },
    },
  },
} as const;
