import { MsgCloseBid, MsgCloseBidResponse, MsgCreateBid, MsgCreateBidResponse } from "./bidmsg.ts";
import { MsgCloseLease, MsgCloseLeaseResponse, MsgCreateLease, MsgCreateLeaseResponse, MsgWithdrawLease, MsgWithdrawLeaseResponse } from "./leasemsg.ts";
import { MsgUpdateParams, MsgUpdateParamsResponse } from "./paramsmsg.ts";

export const Msg = {
  typeName: "virtengine.market.v1beta5.Msg",
  methods: {
    createBid: {
      name: "CreateBid",
      input: MsgCreateBid,
      output: MsgCreateBidResponse,
      get parent() { return Msg; },
    },
    closeBid: {
      name: "CloseBid",
      input: MsgCloseBid,
      output: MsgCloseBidResponse,
      get parent() { return Msg; },
    },
    withdrawLease: {
      name: "WithdrawLease",
      input: MsgWithdrawLease,
      output: MsgWithdrawLeaseResponse,
      get parent() { return Msg; },
    },
    createLease: {
      name: "CreateLease",
      input: MsgCreateLease,
      output: MsgCreateLeaseResponse,
      get parent() { return Msg; },
    },
    closeLease: {
      name: "CloseLease",
      input: MsgCloseLease,
      output: MsgCloseLeaseResponse,
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
