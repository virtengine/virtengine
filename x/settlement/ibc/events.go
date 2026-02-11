// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

// Event types for settlement IBC.
const (
	EventTypePacketSent          = "settlement_ibc_packet_sent"
	EventTypePacketReceived      = "settlement_ibc_packet_received"
	EventTypePacketAcknowledged  = "settlement_ibc_packet_acknowledged"
	EventTypePacketTimeout       = "settlement_ibc_packet_timeout"
	EventTypeEscrowCreated       = "settlement_ibc_escrow_created"
	EventTypeEscrowReleased      = "settlement_ibc_escrow_released"
	EventTypeSettlementRecorded  = "settlement_ibc_settlement_recorded"
)

// Event attribute keys.
const (
	AttributeKeyPacketType        = "packet_type"
	AttributeKeySourceChannel     = "source_channel"
	AttributeKeyDestinationChannel = "destination_channel"
	AttributeKeySequence          = "sequence"
	AttributeKeyEscrowID          = "escrow_id"
	AttributeKeyOrderID           = "order_id"
	AttributeKeySettlementID      = "settlement_id"
	AttributeKeyReleaseType       = "release_type"
	AttributeKeyAckSuccess        = "ack_success"
	AttributeKeyRelayer           = "relayer"
)
