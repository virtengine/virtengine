// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"cosmossdk.io/core/appmodule"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
)

var _ porttypes.IBCModule = IBCModule{}

// IBCModule implements the IBC module interface for VEID attestation packets.
type IBCModule struct {
	keeper IBCKeeper
}

// NewIBCModule creates a new IBCModule.
func NewIBCModule(k IBCKeeper) IBCModule {
	return IBCModule{keeper: k}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (IBCModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (IBCModule) IsAppModule() {}

// Ensure IBCModule satisfies appmodule interfaces for IBC compatibility.
var _ appmodule.AppModule = IBCModule{}

// OnChanOpenInit implements the IBCModule interface.
func (im IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	if order != channeltypes.ORDERED {
		return "", channeltypes.ErrInvalidChannelOrdering
	}
	if version != Version {
		return "", channeltypes.ErrInvalidChannelVersion.Wrapf("expected %s, got %s", Version, version)
	}
	return version, nil
}

// OnChanOpenTry implements the IBCModule interface.
func (im IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	if order != channeltypes.ORDERED {
		return "", channeltypes.ErrInvalidChannelOrdering
	}
	if counterpartyVersion != Version {
		return "", channeltypes.ErrInvalidChannelVersion.Wrapf("expected %s, got %s", Version, counterpartyVersion)
	}
	return Version, nil
}

// OnChanOpenAck implements the IBCModule interface.
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID string,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	if counterpartyVersion != Version {
		return channeltypes.ErrInvalidChannelVersion.Wrapf("expected %s, got %s", Version, counterpartyVersion)
	}
	return nil
}

// OnChanOpenConfirm implements the IBCModule interface.
func (im IBCModule) OnChanOpenConfirm(ctx sdk.Context, portID, channelID string) error {
	return nil
}

// OnChanCloseInit implements the IBCModule interface.
func (im IBCModule) OnChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	// Disallow user-initiated channel closing
	return channeltypes.ErrInvalidChannelState
}

// OnChanCloseConfirm implements the IBCModule interface.
func (im IBCModule) OnChanCloseConfirm(ctx sdk.Context, portID, channelID string) error {
	return nil
}

// OnRecvPacket implements the IBCModule interface.
func (im IBCModule) OnRecvPacket(ctx sdk.Context, _ string, packet channeltypes.Packet, relayer sdk.AccAddress) ibcexported.Acknowledgement {
	return im.keeper.OnRecvPacket(ctx, packet)
}

// OnAcknowledgementPacket implements the IBCModule interface.
func (im IBCModule) OnAcknowledgementPacket(ctx sdk.Context, _ string, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	return im.keeper.OnAcknowledgementPacket(ctx, packet, acknowledgement)
}

// OnTimeoutPacket implements the IBCModule interface.
func (im IBCModule) OnTimeoutPacket(ctx sdk.Context, _ string, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	return im.keeper.OnTimeoutPacket(ctx, packet)
}

// OnChanUpgradeInit implements the IBCModule interface.
func (im IBCModule) OnChanUpgradeInit(ctx sdk.Context, portID, channelID string, proposedOrder channeltypes.Order, connectionHops []string, proposedVersion string) (string, error) {
	return proposedVersion, nil
}

// OnChanUpgradeTry implements the IBCModule interface.
func (im IBCModule) OnChanUpgradeTry(ctx sdk.Context, portID, channelID string, proposedOrder channeltypes.Order, connectionHops []string, counterpartyVersion string) (string, error) {
	return counterpartyVersion, nil
}

// OnChanUpgradeAck implements the IBCModule interface.
func (im IBCModule) OnChanUpgradeAck(ctx sdk.Context, portID, channelID, counterpartyVersion string) error {
	return nil
}

// OnChanUpgradeOpen implements the IBCModule interface.
func (im IBCModule) OnChanUpgradeOpen(ctx sdk.Context, portID, channelID string, proposedOrder channeltypes.Order, connectionHops []string, proposedVersion string) {
}
