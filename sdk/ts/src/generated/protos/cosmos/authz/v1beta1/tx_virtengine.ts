import { MsgExec, MsgExecResponse, MsgGrant, MsgGrantResponse, MsgRevoke, MsgRevokeResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.authz.v1beta1.Msg",
  methods: {
    grant: {
      name: "Grant",
      input: MsgGrant,
      output: MsgGrantResponse,
      get parent() { return Msg; },
    },
    exec: {
      name: "Exec",
      input: MsgExec,
      output: MsgExecResponse,
      get parent() { return Msg; },
    },
    revoke: {
      name: "Revoke",
      input: MsgRevoke,
      output: MsgRevokeResponse,
      get parent() { return Msg; },
    },
  },
} as const;
