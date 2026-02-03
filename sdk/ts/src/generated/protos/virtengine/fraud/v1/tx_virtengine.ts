import { MsgAssignModerator, MsgAssignModeratorResponse, MsgEscalateFraudReport, MsgEscalateFraudReportResponse, MsgRejectFraudReport, MsgRejectFraudReportResponse, MsgResolveFraudReport, MsgResolveFraudReportResponse, MsgSubmitFraudReport, MsgSubmitFraudReportResponse, MsgUpdateParams, MsgUpdateParamsResponse, MsgUpdateReportStatus, MsgUpdateReportStatusResponse } from "./tx.ts";

export const Msg = {
  typeName: "virtengine.fraud.v1.Msg",
  methods: {
    submitFraudReport: {
      name: "SubmitFraudReport",
      input: MsgSubmitFraudReport,
      output: MsgSubmitFraudReportResponse,
      get parent() { return Msg; },
    },
    assignModerator: {
      name: "AssignModerator",
      input: MsgAssignModerator,
      output: MsgAssignModeratorResponse,
      get parent() { return Msg; },
    },
    updateReportStatus: {
      name: "UpdateReportStatus",
      input: MsgUpdateReportStatus,
      output: MsgUpdateReportStatusResponse,
      get parent() { return Msg; },
    },
    resolveFraudReport: {
      name: "ResolveFraudReport",
      input: MsgResolveFraudReport,
      output: MsgResolveFraudReportResponse,
      get parent() { return Msg; },
    },
    rejectFraudReport: {
      name: "RejectFraudReport",
      input: MsgRejectFraudReport,
      output: MsgRejectFraudReportResponse,
      get parent() { return Msg; },
    },
    escalateFraudReport: {
      name: "EscalateFraudReport",
      input: MsgEscalateFraudReport,
      output: MsgEscalateFraudReportResponse,
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
