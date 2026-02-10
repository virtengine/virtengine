import { MsgCreatePeriodicVestingAccount, MsgCreatePeriodicVestingAccountResponse, MsgCreatePermanentLockedAccount, MsgCreatePermanentLockedAccountResponse, MsgCreateVestingAccount, MsgCreateVestingAccountResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.vesting.v1beta1.Msg",
  methods: {
    createVestingAccount: {
      name: "CreateVestingAccount",
      input: MsgCreateVestingAccount,
      output: MsgCreateVestingAccountResponse,
      get parent() { return Msg; },
    },
    createPermanentLockedAccount: {
      name: "CreatePermanentLockedAccount",
      input: MsgCreatePermanentLockedAccount,
      output: MsgCreatePermanentLockedAccountResponse,
      get parent() { return Msg; },
    },
    createPeriodicVestingAccount: {
      name: "CreatePeriodicVestingAccount",
      input: MsgCreatePeriodicVestingAccount,
      output: MsgCreatePeriodicVestingAccountResponse,
      get parent() { return Msg; },
    },
  },
} as const;
