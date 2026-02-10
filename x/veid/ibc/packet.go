// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"encoding/json"
	"fmt"
)

// PacketType identifies the type of IBC packet.
type PacketType string

const (
	// PacketTypeVEIDAttestation is the packet type for VEID attestation packets.
	PacketTypeVEIDAttestation PacketType = "veid_attestation"
)

// IBCPacketData wraps the packet type and payload for IBC transmission.
type IBCPacketData struct {
	Type PacketType      `json:"type"`
	Data json.RawMessage `json:"data"`
}

// EncodeVEIDAttestationPacket encodes a VEIDAttestationPacket into IBC packet bytes.
func EncodeVEIDAttestationPacket(packet VEIDAttestationPacket) ([]byte, error) {
	if err := packet.Validate(); err != nil {
		return nil, fmt.Errorf("invalid attestation packet: %w", err)
	}

	data, err := json.Marshal(packet)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal attestation packet: %w", err)
	}

	ibcPacket := IBCPacketData{
		Type: PacketTypeVEIDAttestation,
		Data: data,
	}

	bz, err := json.Marshal(ibcPacket)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal IBC packet: %w", err)
	}

	return bz, nil
}

// DecodeVEIDAttestationPacket decodes IBC packet bytes into a VEIDAttestationPacket.
func DecodeVEIDAttestationPacket(bz []byte) (VEIDAttestationPacket, error) {
	var ibcPacket IBCPacketData
	if err := json.Unmarshal(bz, &ibcPacket); err != nil {
		return VEIDAttestationPacket{}, fmt.Errorf("failed to unmarshal IBC packet: %w", err)
	}

	if ibcPacket.Type != PacketTypeVEIDAttestation {
		return VEIDAttestationPacket{}, fmt.Errorf("unexpected packet type: %s", ibcPacket.Type)
	}

	var packet VEIDAttestationPacket
	if err := json.Unmarshal(ibcPacket.Data, &packet); err != nil {
		return VEIDAttestationPacket{}, fmt.Errorf("failed to unmarshal attestation packet: %w", err)
	}

	if err := packet.Validate(); err != nil {
		return VEIDAttestationPacket{}, fmt.Errorf("invalid attestation packet: %w", err)
	}

	return packet, nil
}

// EncodeAck encodes a VEIDAttestationAck into bytes.
func EncodeAck(ack VEIDAttestationAck) ([]byte, error) {
	bz, err := json.Marshal(ack)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal acknowledgement: %w", err)
	}
	return bz, nil
}

// DecodeAck decodes bytes into a VEIDAttestationAck.
func DecodeAck(bz []byte) (VEIDAttestationAck, error) {
	var ack VEIDAttestationAck
	if err := json.Unmarshal(bz, &ack); err != nil {
		return VEIDAttestationAck{}, fmt.Errorf("failed to unmarshal acknowledgement: %w", err)
	}
	return ack, nil
}
