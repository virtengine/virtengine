// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
)

// TransferLifecycleHooks performs deterministic custody, accounting, and audit
// writes in the callback's cached SDK context. Implementations must not do I/O.
type TransferLifecycleHooks interface {
	PrepareTransfer(ctx sdk.Context, packet SettlementPacketData) (string, error)
	FinalizeTransfer(ctx sdk.Context, pending PendingPacket, transition TransferTransition) error
	CompensateTransfer(ctx sdk.Context, pending PendingPacket, transition TransferTransition) error
	RecordAccounting(ctx sdk.Context, pending PendingPacket, transition TransferTransition) error
	RecordAudit(ctx sdk.Context, pending PendingPacket, transition TransferTransition) error
}

type unavailableTransferLifecycleHooks struct{}

func (unavailableTransferLifecycleHooks) PrepareTransfer(sdk.Context, SettlementPacketData) (string, error) {
	return "", ErrLifecycleHooksUnavailable
}

func (unavailableTransferLifecycleHooks) FinalizeTransfer(sdk.Context, PendingPacket, TransferTransition) error {
	return ErrLifecycleHooksUnavailable
}

func (unavailableTransferLifecycleHooks) CompensateTransfer(sdk.Context, PendingPacket, TransferTransition) error {
	return ErrLifecycleHooksUnavailable
}

func (unavailableTransferLifecycleHooks) RecordAccounting(sdk.Context, PendingPacket, TransferTransition) error {
	return ErrLifecycleHooksUnavailable
}

func (unavailableTransferLifecycleHooks) RecordAudit(sdk.Context, PendingPacket, TransferTransition) error {
	return ErrLifecycleHooksUnavailable
}

func validateCustodyReference(reference string) error {
	if reference == "" {
		return fmt.Errorf("custody reference cannot be empty")
	}
	return nil
}

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
