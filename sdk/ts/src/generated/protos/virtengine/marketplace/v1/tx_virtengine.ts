import { MsgWaldurCallback, MsgWaldurCallbackResponse } from "./tx.ts";

export const Msg = {
  typeName: "virtengine.marketplace.v1.Msg",
  methods: {
    waldurCallback: {
      name: "WaldurCallback",
      input: MsgWaldurCallback,
      output: MsgWaldurCallbackResponse,
      get parent() { return Msg; },
    },
  },
} as const;
