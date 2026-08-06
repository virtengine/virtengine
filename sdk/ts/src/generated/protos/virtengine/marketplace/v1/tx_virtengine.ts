import { MsgAcceptBid, MsgAcceptBidResponse, MsgCreateOffering, MsgCreateOfferingResponse, MsgDeactivateOffering, MsgDeactivateOfferingResponse, MsgPauseAllocation, MsgPauseAllocationResponse, MsgResizeAllocation, MsgResizeAllocationResponse, MsgTerminateAllocation, MsgTerminateAllocationResponse, MsgUpdateOffering, MsgUpdateOfferingResponse, MsgWaldurCallback, MsgWaldurCallbackResponse } from "./tx.ts";

export const Msg = {
  typeName: "virtengine.marketplace.v1.Msg",
  methods: {
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
    deactivateOffering: {
      name: "DeactivateOffering",
      input: MsgDeactivateOffering,
      output: MsgDeactivateOfferingResponse,
      get parent() { return Msg; },
    },
    acceptBid: {
      name: "AcceptBid",
      input: MsgAcceptBid,
      output: MsgAcceptBidResponse,
      get parent() { return Msg; },
    },
    terminateAllocation: {
      name: "TerminateAllocation",
      input: MsgTerminateAllocation,
      output: MsgTerminateAllocationResponse,
      get parent() { return Msg; },
    },
    resizeAllocation: {
      name: "ResizeAllocation",
      input: MsgResizeAllocation,
      output: MsgResizeAllocationResponse,
      get parent() { return Msg; },
    },
    pauseAllocation: {
      name: "PauseAllocation",
      input: MsgPauseAllocation,
      output: MsgPauseAllocationResponse,
      get parent() { return Msg; },
    },
    waldurCallback: {
      name: "WaldurCallback",
      input: MsgWaldurCallback,
      output: MsgWaldurCallbackResponse,
      get parent() { return Msg; },
    },
  },
} as const;
