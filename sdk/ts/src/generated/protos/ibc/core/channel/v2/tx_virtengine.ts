import { MsgAcknowledgement, MsgAcknowledgementResponse, MsgRecvPacket, MsgRecvPacketResponse, MsgSendPacket, MsgSendPacketResponse, MsgTimeout, MsgTimeoutResponse } from "./tx.ts";

export const Msg = {
  typeName: "ibc.core.channel.v2.Msg",
  methods: {
    sendPacket: {
      name: "SendPacket",
      input: MsgSendPacket,
      output: MsgSendPacketResponse,
      get parent() { return Msg; },
    },
    recvPacket: {
      name: "RecvPacket",
      input: MsgRecvPacket,
      output: MsgRecvPacketResponse,
      get parent() { return Msg; },
    },
    timeout: {
      name: "Timeout",
      input: MsgTimeout,
      output: MsgTimeoutResponse,
      get parent() { return Msg; },
    },
    acknowledgement: {
      name: "Acknowledgement",
      input: MsgAcknowledgement,
      output: MsgAcknowledgementResponse,
      get parent() { return Msg; },
    },
  },
} as const;
