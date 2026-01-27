import { MsgGrantAllowance, MsgGrantAllowanceResponse, MsgPruneAllowances, MsgPruneAllowancesResponse, MsgRevokeAllowance, MsgRevokeAllowanceResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.feegrant.v1beta1.Msg",
  methods: {
    grantAllowance: {
      name: "GrantAllowance",
      input: MsgGrantAllowance,
      output: MsgGrantAllowanceResponse,
      get parent() { return Msg; },
    },
    revokeAllowance: {
      name: "RevokeAllowance",
      input: MsgRevokeAllowance,
      output: MsgRevokeAllowanceResponse,
      get parent() { return Msg; },
    },
    pruneAllowances: {
      name: "PruneAllowances",
      input: MsgPruneAllowances,
      output: MsgPruneAllowancesResponse,
      get parent() { return Msg; },
    },
  },
} as const;
