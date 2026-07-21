import { MsgCreateExportJob, MsgCreateExportJobResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./msg_audit_log.ts";

export const MsgService = {
  typeName: "virtengine.audit.v1.MsgService",
  methods: {
    createExportJob: {
      name: "CreateExportJob",
      input: MsgCreateExportJob,
      output: MsgCreateExportJobResponse,
      get parent() { return MsgService; },
    },
    updateParams: {
      name: "UpdateParams",
      input: MsgUpdateParams,
      output: MsgUpdateParamsResponse,
      get parent() { return MsgService; },
    },
  },
} as const;
