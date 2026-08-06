import { MsgAddSupportResponse, MsgAddSupportResponseResponse, MsgArchiveSupportRequest, MsgArchiveSupportRequestResponse, MsgCreateSupportRequest, MsgCreateSupportRequestResponse, MsgRegisterExternalTicket, MsgRegisterExternalTicketResponse, MsgRemoveExternalTicket, MsgRemoveExternalTicketResponse, MsgUpdateExternalTicket, MsgUpdateExternalTicketResponse, MsgUpdateParams, MsgUpdateParamsResponse, MsgUpdateSupportRequest, MsgUpdateSupportRequestResponse } from "./tx.ts";

export const Msg = {
  typeName: "virtengine.support.v1.Msg",
  methods: {
    createSupportRequest: {
      name: "CreateSupportRequest",
      input: MsgCreateSupportRequest,
      output: MsgCreateSupportRequestResponse,
      get parent() { return Msg; },
    },
    updateSupportRequest: {
      name: "UpdateSupportRequest",
      input: MsgUpdateSupportRequest,
      output: MsgUpdateSupportRequestResponse,
      get parent() { return Msg; },
    },
    addSupportResponse: {
      name: "AddSupportResponse",
      input: MsgAddSupportResponse,
      output: MsgAddSupportResponseResponse,
      get parent() { return Msg; },
    },
    archiveSupportRequest: {
      name: "ArchiveSupportRequest",
      input: MsgArchiveSupportRequest,
      output: MsgArchiveSupportRequestResponse,
      get parent() { return Msg; },
    },
    registerExternalTicket: {
      name: "RegisterExternalTicket",
      input: MsgRegisterExternalTicket,
      output: MsgRegisterExternalTicketResponse,
      get parent() { return Msg; },
    },
    updateExternalTicket: {
      name: "UpdateExternalTicket",
      input: MsgUpdateExternalTicket,
      output: MsgUpdateExternalTicketResponse,
      get parent() { return Msg; },
    },
    removeExternalTicket: {
      name: "RemoveExternalTicket",
      input: MsgRemoveExternalTicket,
      output: MsgRemoveExternalTicketResponse,
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
