import { MsgLoadTest, MsgLoadTestResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.benchmark.v1.Msg",
  methods: {
    loadTest: {
      name: "LoadTest",
      input: MsgLoadTest,
      output: MsgLoadTestResponse,
      get parent() { return Msg; },
    },
  },
} as const;
