import { MsgSend, MsgSendResponse } from "./tx.ts";

export const Msg = {
  typeName: "cosmos.nft.v1beta1.Msg",
  methods: {
    send: {
      name: "Send",
      input: MsgSend,
      output: MsgSendResponse,
      get parent() { return Msg; },
    },
  },
} as const;
