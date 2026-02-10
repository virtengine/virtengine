import { MsgAssignRole, MsgAssignRoleResponse, MsgNominateAdmin, MsgNominateAdminResponse, MsgRevokeRole, MsgRevokeRoleResponse, MsgSetAccountState, MsgSetAccountStateResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

export const Msg = {
  typeName: "virtengine.roles.v1.Msg",
  methods: {
    assignRole: {
      name: "AssignRole",
      input: MsgAssignRole,
      output: MsgAssignRoleResponse,
      get parent() { return Msg; },
    },
    revokeRole: {
      name: "RevokeRole",
      input: MsgRevokeRole,
      output: MsgRevokeRoleResponse,
      get parent() { return Msg; },
    },
    setAccountState: {
      name: "SetAccountState",
      input: MsgSetAccountState,
      output: MsgSetAccountStateResponse,
      get parent() { return Msg; },
    },
    nominateAdmin: {
      name: "NominateAdmin",
      input: MsgNominateAdmin,
      output: MsgNominateAdminResponse,
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
