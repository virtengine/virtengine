import { MsgAcceptBid, MsgAcceptBidResponse, MsgAckWaldurCommand, MsgAckWaldurCommandResponse, MsgCreateOffering, MsgCreateOfferingResponse, MsgCreateOrder, MsgCreateOrderResponse, MsgDeactivateOffering, MsgDeactivateOfferingResponse, MsgIngestWaldurOffering, MsgIngestWaldurOfferingResponse, MsgPauseAllocation, MsgPauseAllocationResponse, MsgPlaceBid, MsgPlaceBidResponse, MsgRegisterWaldurSource, MsgRegisterWaldurSourceResponse, MsgResizeAllocation, MsgResizeAllocationResponse, MsgSetOfferingVisibility, MsgSetOfferingVisibilityResponse, MsgTerminateAllocation, MsgTerminateAllocationResponse, MsgUpdateOffering, MsgUpdateOfferingResponse, MsgWaldurCallback, MsgWaldurCallbackResponse, MsgWithdrawBid, MsgWithdrawBidResponse } from "./tx.ts";

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
    createOrder: {
      name: "CreateOrder",
      input: MsgCreateOrder,
      output: MsgCreateOrderResponse,
      get parent() { return Msg; },
    },
    placeBid: {
      name: "PlaceBid",
      input: MsgPlaceBid,
      output: MsgPlaceBidResponse,
      get parent() { return Msg; },
    },
    withdrawBid: {
      name: "WithdrawBid",
      input: MsgWithdrawBid,
      output: MsgWithdrawBidResponse,
      get parent() { return Msg; },
    },
    registerWaldurSource: {
      name: "RegisterWaldurSource",
      input: MsgRegisterWaldurSource,
      output: MsgRegisterWaldurSourceResponse,
      get parent() { return Msg; },
    },
    ingestWaldurOffering: {
      name: "IngestWaldurOffering",
      input: MsgIngestWaldurOffering,
      output: MsgIngestWaldurOfferingResponse,
      get parent() { return Msg; },
    },
    setOfferingVisibility: {
      name: "SetOfferingVisibility",
      input: MsgSetOfferingVisibility,
      output: MsgSetOfferingVisibilityResponse,
      get parent() { return Msg; },
    },
    ackWaldurCommand: {
      name: "AckWaldurCommand",
      input: MsgAckWaldurCommand,
      output: MsgAckWaldurCommandResponse,
      get parent() { return Msg; },
    },
  },
} as const;
