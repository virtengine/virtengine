import { MsgAccountDeposit, MsgAccountDepositResponse } from "./msg.ts";

export const Msg = {
  typeName: "virtengine.escrow.v1.Msg",
  methods: {
    accountDeposit: {
      name: "AccountDeposit",
      input: MsgAccountDeposit,
      output: MsgAccountDepositResponse,
      get parent() { return Msg; },
    },
  },
} as const;
