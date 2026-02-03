import { MsgReactivateApprovedClient, MsgReactivateApprovedClientResponse, MsgRegisterApprovedClient, MsgRegisterApprovedClientResponse, MsgRevokeApprovedClient, MsgRevokeApprovedClientResponse, MsgSuspendApprovedClient, MsgSuspendApprovedClientResponse, MsgUpdateApprovedClient, MsgUpdateApprovedClientResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

export const Msg = {
  typeName: "virtengine.config.v1.Msg",
  methods: {
    registerApprovedClient: {
      name: "RegisterApprovedClient",
      input: MsgRegisterApprovedClient,
      output: MsgRegisterApprovedClientResponse,
      get parent() { return Msg; },
    },
    updateApprovedClient: {
      name: "UpdateApprovedClient",
      input: MsgUpdateApprovedClient,
      output: MsgUpdateApprovedClientResponse,
      get parent() { return Msg; },
    },
    suspendApprovedClient: {
      name: "SuspendApprovedClient",
      input: MsgSuspendApprovedClient,
      output: MsgSuspendApprovedClientResponse,
      get parent() { return Msg; },
    },
    revokeApprovedClient: {
      name: "RevokeApprovedClient",
      input: MsgRevokeApprovedClient,
      output: MsgRevokeApprovedClientResponse,
      get parent() { return Msg; },
    },
    reactivateApprovedClient: {
      name: "ReactivateApprovedClient",
      input: MsgReactivateApprovedClient,
      output: MsgReactivateApprovedClientResponse,
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
