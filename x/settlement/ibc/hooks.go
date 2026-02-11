// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
)

// RelayerHooks allows custom handling for relayer events.
type RelayerHooks interface {
	OnPacketReceived(ctx sdk.Context, relayer sdk.AccAddress, packet channeltypes.Packet, packetType PacketType)
	OnPacketAcknowledged(ctx sdk.Context, relayer sdk.AccAddress, packet channeltypes.Packet, packetType PacketType, success bool)
	OnPacketTimeout(ctx sdk.Context, relayer sdk.AccAddress, packet channeltypes.Packet, packetType PacketType)
}

// NoOpRelayerHooks is a no-op implementation.
type NoOpRelayerHooks struct{}

func (NoOpRelayerHooks) OnPacketReceived(ctx sdk.Context, relayer sdk.AccAddress, packet channeltypes.Packet, packetType PacketType) {
}

func (NoOpRelayerHooks) OnPacketAcknowledged(ctx sdk.Context, relayer sdk.AccAddress, packet channeltypes.Packet, packetType PacketType, success bool) {
}

func (NoOpRelayerHooks) OnPacketTimeout(ctx sdk.Context, relayer sdk.AccAddress, packet channeltypes.Packet, packetType PacketType) {
}
