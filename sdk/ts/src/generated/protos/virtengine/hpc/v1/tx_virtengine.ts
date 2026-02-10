import { MsgCancelJob, MsgCancelJobResponse, MsgCreateOffering, MsgCreateOfferingResponse, MsgDeregisterCluster, MsgDeregisterClusterResponse, MsgFlagDispute, MsgFlagDisputeResponse, MsgRegisterCluster, MsgRegisterClusterResponse, MsgReportJobStatus, MsgReportJobStatusResponse, MsgResolveDispute, MsgResolveDisputeResponse, MsgSubmitJob, MsgSubmitJobResponse, MsgUpdateCluster, MsgUpdateClusterResponse, MsgUpdateNodeMetadata, MsgUpdateNodeMetadataResponse, MsgUpdateOffering, MsgUpdateOfferingResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

export const Msg = {
  typeName: "virtengine.hpc.v1.Msg",
  methods: {
    registerCluster: {
      name: "RegisterCluster",
      input: MsgRegisterCluster,
      output: MsgRegisterClusterResponse,
      get parent() { return Msg; },
    },
    updateCluster: {
      name: "UpdateCluster",
      input: MsgUpdateCluster,
      output: MsgUpdateClusterResponse,
      get parent() { return Msg; },
    },
    deregisterCluster: {
      name: "DeregisterCluster",
      input: MsgDeregisterCluster,
      output: MsgDeregisterClusterResponse,
      get parent() { return Msg; },
    },
    createOffering: {
      name: "CreateOffering",
      input: MsgCreateOffering,
      output: MsgCreateOfferingResponse,
      get parent() { return Msg; },
    },
    updateOffering: {
      name: "UpdateOffering",
      input: MsgUpdateOffering,
      output: MsgUpdateOfferingResponse,
      get parent() { return Msg; },
    },
    submitJob: {
      name: "SubmitJob",
      input: MsgSubmitJob,
      output: MsgSubmitJobResponse,
      get parent() { return Msg; },
    },
    cancelJob: {
      name: "CancelJob",
      input: MsgCancelJob,
      output: MsgCancelJobResponse,
      get parent() { return Msg; },
    },
    reportJobStatus: {
      name: "ReportJobStatus",
      input: MsgReportJobStatus,
      output: MsgReportJobStatusResponse,
      get parent() { return Msg; },
    },
    updateNodeMetadata: {
      name: "UpdateNodeMetadata",
      input: MsgUpdateNodeMetadata,
      output: MsgUpdateNodeMetadataResponse,
      get parent() { return Msg; },
    },
    flagDispute: {
      name: "FlagDispute",
      input: MsgFlagDispute,
      output: MsgFlagDisputeResponse,
      get parent() { return Msg; },
    },
    resolveDispute: {
      name: "ResolveDispute",
      input: MsgResolveDispute,
      output: MsgResolveDisputeResponse,
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
