import { MsgIncreaseCounter, MsgIncreaseCountResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.counter.v1.Msg",
  methods: {
    increaseCount: {
      name: "IncreaseCount",
      input: MsgIncreaseCounter,
      output: MsgIncreaseCountResponse,
      get parent() { return Msg; },
    },
  },
} as const;
