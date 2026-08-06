import { MsgActivateAllocation, MsgActivateAllocationResponse, MsgAllocateResources, MsgAllocateResourcesResponse, MsgProviderHeartbeat, MsgProviderHeartbeatResponse, MsgReleaseAllocation, MsgReleaseAllocationResponse, MsgUpdateParams, MsgUpdateParamsResponse } from "./tx.ts";

export const Msg = {
  typeName: "virtengine.resources.v1.Msg",
  methods: {
    providerHeartbeat: {
      name: "ProviderHeartbeat",
      input: MsgProviderHeartbeat,
      output: MsgProviderHeartbeatResponse,
      get parent() { return Msg; },
    },
    allocateResources: {
      name: "AllocateResources",
      input: MsgAllocateResources,
      output: MsgAllocateResourcesResponse,
      get parent() { return Msg; },
    },
    activateAllocation: {
      name: "ActivateAllocation",
      input: MsgActivateAllocation,
      output: MsgActivateAllocationResponse,
      get parent() { return Msg; },
    },
    releaseAllocation: {
      name: "ReleaseAllocation",
      input: MsgReleaseAllocation,
      output: MsgReleaseAllocationResponse,
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
