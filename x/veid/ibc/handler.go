// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
)

// OnRecvPacket handles incoming VEID attestation packets from other chains.
func (k IBCKeeper) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet) ibcexported.Acknowledgement {
	// Decode the packet
	attestation, err := DecodeVEIDAttestationPacket(packet.GetData())
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("failed to decode packet: %w", err))
	}

	// Process the attestation
	record, err := k.ProcessAttestation(ctx, attestation, packet.GetSourceChannel())
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("failed to process attestation: %w", err))
	}

	// Emit event
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"veid_cross_chain_attestation_received",
		sdk.NewAttribute("source_chain_id", attestation.SourceChainID),
		sdk.NewAttribute("account_address", attestation.AccountAddress),
		sdk.NewAttribute("original_score", fmt.Sprintf("%d", attestation.TrustScore)),
		sdk.NewAttribute("recognized_score", fmt.Sprintf("%d", record.RecognizedScore)),
		sdk.NewAttribute("tier_level", fmt.Sprintf("%d", attestation.TierLevel)),
		sdk.NewAttribute("source_channel", packet.GetSourceChannel()),
	))

	// Return success acknowledgement
	ack := VEIDAttestationAck{
		Success:         true,
		RecognizedScore: record.RecognizedScore,
	}

	ackBz, err := EncodeAck(ack)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("failed to encode ack: %w", err))
	}

	return channeltypes.NewResultAcknowledgement(ackBz)
}

// OnAcknowledgementPacket handles acknowledgements for sent VEID attestation packets.
func (k IBCKeeper) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte) error {
	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return fmt.Errorf("failed to unmarshal acknowledgement: %w", err)
	}

	switch resp := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		// Decode our custom ack from the result
		veidAck, err := DecodeAck(resp.Result)
		if err != nil {
			return fmt.Errorf("failed to decode VEID ack: %w", err)
		}

		ctx.EventManager().EmitEvent(sdk.NewEvent(
			"veid_cross_chain_attestation_ack",
			sdk.NewAttribute("success", "true"),
			sdk.NewAttribute("recognized_score", fmt.Sprintf("%d", veidAck.RecognizedScore)),
			sdk.NewAttribute("dest_channel", packet.GetDestChannel()),
		))

	case *channeltypes.Acknowledgement_Error:
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			"veid_cross_chain_attestation_ack",
			sdk.NewAttribute("success", "false"),
			sdk.NewAttribute("error", resp.Error),
			sdk.NewAttribute("dest_channel", packet.GetDestChannel()),
		))
	}

	return nil
}

// OnTimeoutPacket handles timeout for sent VEID attestation packets.
func (k IBCKeeper) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet) error {
	attestation, err := DecodeVEIDAttestationPacket(packet.GetData())
	if err != nil {
		return fmt.Errorf("failed to decode timed out packet: %w", err)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"veid_cross_chain_attestation_timeout",
		sdk.NewAttribute("account_address", attestation.AccountAddress),
		sdk.NewAttribute("dest_channel", packet.GetDestChannel()),
		sdk.NewAttribute("timeout_height", packet.GetTimeoutHeight().String()),
	))

	return nil
}
