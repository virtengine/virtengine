import { MsgCloseDeployment, MsgCloseDeploymentResponse, MsgCreateDeployment, MsgCreateDeploymentResponse, MsgUpdateDeployment, MsgUpdateDeploymentResponse } from "./deploymentmsg.ts";
import { MsgCloseGroup, MsgCloseGroupResponse, MsgPauseGroup, MsgPauseGroupResponse, MsgStartGroup, MsgStartGroupResponse } from "./groupmsg.ts";
import { MsgUpdateParams, MsgUpdateParamsResponse } from "./paramsmsg.ts";

export const Msg = {
  typeName: "virtengine.deployment.v1beta5.Msg",
  methods: {
    createDeployment: {
      name: "CreateDeployment",
      input: MsgCreateDeployment,
      output: MsgCreateDeploymentResponse,
      get parent() { return Msg; },
    },
    updateDeployment: {
      name: "UpdateDeployment",
      input: MsgUpdateDeployment,
      output: MsgUpdateDeploymentResponse,
      get parent() { return Msg; },
    },
    closeDeployment: {
      name: "CloseDeployment",
      input: MsgCloseDeployment,
      output: MsgCloseDeploymentResponse,
      get parent() { return Msg; },
    },
    closeGroup: {
      name: "CloseGroup",
      input: MsgCloseGroup,
      output: MsgCloseGroupResponse,
      get parent() { return Msg; },
    },
    pauseGroup: {
      name: "PauseGroup",
      input: MsgPauseGroup,
      output: MsgPauseGroupResponse,
      get parent() { return Msg; },
    },
    startGroup: {
      name: "StartGroup",
      input: MsgStartGroup,
      output: MsgStartGroupResponse,
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
