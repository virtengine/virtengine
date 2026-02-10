import { MsgAcknowledgement, MsgAcknowledgementResponse, MsgChannelCloseConfirm, MsgChannelCloseConfirmResponse, MsgChannelCloseInit, MsgChannelCloseInitResponse, MsgChannelOpenAck, MsgChannelOpenAckResponse, MsgChannelOpenConfirm, MsgChannelOpenConfirmResponse, MsgChannelOpenInit, MsgChannelOpenInitResponse, MsgChannelOpenTry, MsgChannelOpenTryResponse, MsgRecvPacket, MsgRecvPacketResponse, MsgTimeout, MsgTimeoutOnClose, MsgTimeoutOnCloseResponse, MsgTimeoutResponse } from "./tx.ts";

export const Msg = {
  typeName: "ibc.core.channel.v1.Msg",
  methods: {
    channelOpenInit: {
      name: "ChannelOpenInit",
      input: MsgChannelOpenInit,
      output: MsgChannelOpenInitResponse,
      get parent() { return Msg; },
    },
    channelOpenTry: {
      name: "ChannelOpenTry",
      input: MsgChannelOpenTry,
      output: MsgChannelOpenTryResponse,
      get parent() { return Msg; },
    },
    channelOpenAck: {
      name: "ChannelOpenAck",
      input: MsgChannelOpenAck,
      output: MsgChannelOpenAckResponse,
      get parent() { return Msg; },
    },
    channelOpenConfirm: {
      name: "ChannelOpenConfirm",
      input: MsgChannelOpenConfirm,
      output: MsgChannelOpenConfirmResponse,
      get parent() { return Msg; },
    },
    channelCloseInit: {
      name: "ChannelCloseInit",
      input: MsgChannelCloseInit,
      output: MsgChannelCloseInitResponse,
      get parent() { return Msg; },
    },
    channelCloseConfirm: {
      name: "ChannelCloseConfirm",
      input: MsgChannelCloseConfirm,
      output: MsgChannelCloseConfirmResponse,
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
    timeoutOnClose: {
      name: "TimeoutOnClose",
      input: MsgTimeoutOnClose,
      output: MsgTimeoutOnCloseResponse,
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
