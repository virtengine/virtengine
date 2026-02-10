import { QueryNextSequenceSendRequest, QueryNextSequenceSendResponse, QueryPacketAcknowledgementRequest, QueryPacketAcknowledgementResponse, QueryPacketAcknowledgementsRequest, QueryPacketAcknowledgementsResponse, QueryPacketCommitmentRequest, QueryPacketCommitmentResponse, QueryPacketCommitmentsRequest, QueryPacketCommitmentsResponse, QueryPacketReceiptRequest, QueryPacketReceiptResponse, QueryUnreceivedAcksRequest, QueryUnreceivedAcksResponse, QueryUnreceivedPacketsRequest, QueryUnreceivedPacketsResponse } from "./query.ts";

export const Query = {
  typeName: "ibc.core.channel.v2.Query",
  methods: {
    nextSequenceSend: {
      name: "NextSequenceSend",
      httpPath: "/ibc/core/channel/v2/clients/{client_id}/next_sequence_send",
      input: QueryNextSequenceSendRequest,
      output: QueryNextSequenceSendResponse,
      get parent() { return Query; },
    },
    packetCommitment: {
      name: "PacketCommitment",
      httpPath: "/ibc/core/channel/v2/clients/{client_id}/packet_commitments/{sequence}",
      input: QueryPacketCommitmentRequest,
      output: QueryPacketCommitmentResponse,
      get parent() { return Query; },
    },
    packetCommitments: {
      name: "PacketCommitments",
      httpPath: "/ibc/core/channel/v2/clients/{client_id}/packet_commitments",
      input: QueryPacketCommitmentsRequest,
      output: QueryPacketCommitmentsResponse,
      get parent() { return Query; },
    },
    packetAcknowledgement: {
      name: "PacketAcknowledgement",
      httpPath: "/ibc/core/channel/v2/clients/{client_id}/packet_acks/{sequence}",
      input: QueryPacketAcknowledgementRequest,
      output: QueryPacketAcknowledgementResponse,
      get parent() { return Query; },
    },
    packetAcknowledgements: {
      name: "PacketAcknowledgements",
      httpPath: "/ibc/core/channel/v2/clients/{client_id}/packet_acknowledgements",
      input: QueryPacketAcknowledgementsRequest,
      output: QueryPacketAcknowledgementsResponse,
      get parent() { return Query; },
    },
    packetReceipt: {
      name: "PacketReceipt",
      httpPath: "/ibc/core/channel/v2/clients/{client_id}/packet_receipts/{sequence}",
      input: QueryPacketReceiptRequest,
      output: QueryPacketReceiptResponse,
      get parent() { return Query; },
    },
    unreceivedPackets: {
      name: "UnreceivedPackets",
      httpPath: "/ibc/core/channel/v2/clients/{client_id}/packet_commitments/{sequences}/unreceived_packets",
      input: QueryUnreceivedPacketsRequest,
      output: QueryUnreceivedPacketsResponse,
      get parent() { return Query; },
    },
    unreceivedAcks: {
      name: "UnreceivedAcks",
      httpPath: "/ibc/core/channel/v2/clients/{client_id}/packet_commitments/{packet_ack_sequences}/unreceived_acks",
      input: QueryUnreceivedAcksRequest,
      output: QueryUnreceivedAcksResponse,
      get parent() { return Query; },
    },
  },
} as const;
