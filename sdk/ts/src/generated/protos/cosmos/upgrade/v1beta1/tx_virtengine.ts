import { MsgCancelUpgrade, MsgCancelUpgradeResponse, MsgSoftwareUpgrade, MsgSoftwareUpgradeResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.upgrade.v1beta1.Msg",
  methods: {
    softwareUpgrade: {
      name: "SoftwareUpgrade",
      input: MsgSoftwareUpgrade,
      output: MsgSoftwareUpgradeResponse,
      get parent() { return Msg; },
    },
    cancelUpgrade: {
      name: "CancelUpgrade",
      input: MsgCancelUpgrade,
      output: MsgCancelUpgradeResponse,
      get parent() { return Msg; },
    },
  },
} as const;
