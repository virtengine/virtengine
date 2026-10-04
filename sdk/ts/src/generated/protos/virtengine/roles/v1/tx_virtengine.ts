import { MsgAssignRole, MsgAssignRoleResponse, MsgConfirmSanction, MsgConfirmSanctionResponse, MsgImposeSanction, MsgImposeSanctionResponse, MsgNominateAdmin, MsgNominateAdminResponse, MsgOpenSanctionAppeal, MsgOpenSanctionAppealResponse, MsgResolveSanctionAppeal, MsgResolveSanctionAppealResponse, MsgRevokeRole, MsgRevokeRoleResponse, MsgRevokeSanction, MsgRevokeSanctionResponse, MsgSetAccountState, MsgSetAccountStateResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

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
    imposeSanction: {
      name: "ImposeSanction",
      input: MsgImposeSanction,
      output: MsgImposeSanctionResponse,
      get parent() { return Msg; },
    },
    confirmSanction: {
      name: "ConfirmSanction",
      input: MsgConfirmSanction,
      output: MsgConfirmSanctionResponse,
      get parent() { return Msg; },
    },
    revokeSanction: {
      name: "RevokeSanction",
      input: MsgRevokeSanction,
      output: MsgRevokeSanctionResponse,
      get parent() { return Msg; },
    },
    openSanctionAppeal: {
      name: "OpenSanctionAppeal",
      input: MsgOpenSanctionAppeal,
      output: MsgOpenSanctionAppealResponse,
      get parent() { return Msg; },
    },
    resolveSanctionAppeal: {
      name: "ResolveSanctionAppeal",
      input: MsgResolveSanctionAppeal,
      output: MsgResolveSanctionAppealResponse,
      get parent() { return Msg; },
    },
  },
} as const;
