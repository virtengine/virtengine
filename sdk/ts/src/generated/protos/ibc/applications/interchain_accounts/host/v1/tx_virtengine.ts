import { MsgModuleQuerySafe, MsgModuleQuerySafeResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

export const Msg = {
  typeName: "ibc.applications.interchain_accounts.host.v1.Msg",
  methods: {
    updateParams: {
      name: "UpdateParams",
      input: MsgUpdateParams,
      output: MsgUpdateParamsResponse,
      get parent() { return Msg; },
    },
    moduleQuerySafe: {
      name: "ModuleQuerySafe",
      input: MsgModuleQuerySafe,
      output: MsgModuleQuerySafeResponse,
      get parent() { return Msg; },
    },
  },
} as const;
