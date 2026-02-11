// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
)

var _ porttypes.IBCModule = (*IBCModule)(nil)

// IBCModule implements the IBC Module interface for settlement.
type IBCModule struct {
	keeper IBCKeeper
}

// NewIBCModule creates a new settlement IBC module.
func NewIBCModule(keeper IBCKeeper) IBCModule {
	return IBCModule{keeper: keeper}
}

// OnChanOpenInit implements the IBCModule interface.
func (m IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	if portID != PortID {
		return "", ErrInvalidPort.Wrapf("expected %s, got %s", PortID, portID)
	}

	if order != channeltypes.ORDERED {
		return "", channeltypes.ErrInvalidChannelOrdering.Wrapf("expected %s channel", channeltypes.ORDERED)
	}

	if version != "" && version != Version {
		return "", ErrInvalidVersion.Wrapf("expected %s, got %s", Version, version)
	}

	if len(connectionHops) == 0 {
		return "", channeltypes.ErrInvalidConnectionHops
	}

	m.keeper.StoreHandshakeRecord(ctx, channelID)

	return Version, nil
}

// OnChanOpenTry implements the IBCModule interface.
func (m IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	if portID != PortID {
		return "", ErrInvalidPort.Wrapf("expected %s, got %s", PortID, portID)
	}

	if order != channeltypes.ORDERED {
		return "", channeltypes.ErrInvalidChannelOrdering.Wrapf("expected %s channel", channeltypes.ORDERED)
	}

	if counterpartyVersion != Version {
		return "", ErrInvalidVersion.Wrapf("expected %s, got %s", Version, counterpartyVersion)
	}

	if len(connectionHops) == 0 {
		return "", channeltypes.ErrInvalidConnectionHops
	}

	m.keeper.StoreHandshakeRecord(ctx, channelID)

	return Version, nil
}

// OnChanOpenAck implements the IBCModule interface.
func (m IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID string,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	if counterpartyVersion != Version {
		return ErrInvalidVersion.Wrapf("expected %s, got %s", Version, counterpartyVersion)
	}

	if err := m.keeper.CheckHandshakeTimeout(ctx, channelID); err != nil {
		return err
	}

	m.keeper.ClearHandshakeRecord(ctx, channelID)
	return nil
}

// OnChanOpenConfirm implements the IBCModule interface.
func (m IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID string,
	channelID string,
) error {
	if err := m.keeper.CheckHandshakeTimeout(ctx, channelID); err != nil {
		return err
	}

	m.keeper.ClearHandshakeRecord(ctx, channelID)
	return nil
}

// OnChanCloseInit implements the IBCModule interface.
func (m IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID string,
	channelID string,
) error {
	return channeltypes.ErrInvalidChannelState.Wrap("user cannot close channel")
}

// OnChanCloseConfirm implements the IBCModule interface.
func (m IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID string,
	channelID string,
) error {
	m.keeper.ClearHandshakeRecord(ctx, channelID)
	return nil
}

// OnRecvPacket implements the IBCModule interface.
func (m IBCModule) OnRecvPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	return m.keeper.OnRecvPacket(ctx, packet, relayer)
}

// OnAcknowledgementPacket implements the IBCModule interface.
func (m IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return m.keeper.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface.
func (m IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return m.keeper.OnTimeoutPacket(ctx, packet, relayer)
}

// OnChanUpgradeInit implements the IBCModule interface.
func (m IBCModule) OnChanUpgradeInit(
	ctx sdk.Context,
	portID string,
	channelID string,
	proposedOrder channeltypes.Order,
	connectionHops []string,
	proposedVersion string,
) (string, error) {
	if proposedVersion != Version {
		return "", ErrInvalidVersion.Wrapf("expected %s, got %s", Version, proposedVersion)
	}
	if proposedOrder != channeltypes.ORDERED {
		return "", channeltypes.ErrInvalidChannelOrdering.Wrapf("expected %s channel", channeltypes.ORDERED)
	}
	return Version, nil
}

// OnChanUpgradeTry implements the IBCModule interface.
func (m IBCModule) OnChanUpgradeTry(
	ctx sdk.Context,
	portID string,
	channelID string,
	proposedOrder channeltypes.Order,
	connectionHops []string,
	counterpartyVersion string,
) (string, error) {
	if counterpartyVersion != Version {
		return "", ErrInvalidVersion.Wrapf("expected %s, got %s", Version, counterpartyVersion)
	}
	if proposedOrder != channeltypes.ORDERED {
		return "", channeltypes.ErrInvalidChannelOrdering.Wrapf("expected %s channel", channeltypes.ORDERED)
	}
	return Version, nil
}

// OnChanUpgradeAck implements the IBCModule interface.
func (m IBCModule) OnChanUpgradeAck(
	ctx sdk.Context,
	portID string,
	channelID string,
	counterpartyVersion string,
) error {
	if counterpartyVersion != Version {
		return ErrInvalidVersion.Wrapf("expected %s, got %s", Version, counterpartyVersion)
	}
	return nil
}

// OnChanUpgradeOpen implements the IBCModule interface.
func (m IBCModule) OnChanUpgradeOpen(
	ctx sdk.Context,
	portID string,
	channelID string,
	proposedOrder channeltypes.Order,
	connectionHops []string,
	proposedVersion string,
) {
	if proposedVersion != Version {
		m.keeper.Logger(ctx).Error("unexpected channel version", "version", proposedVersion)
		return
	}
	m.keeper.Logger(ctx).Info("channel upgraded", "port", portID, "channel", channelID, "version", proposedVersion)
}

// UnmarshalPacketData is unsupported for settlement IBC.
func (m IBCModule) UnmarshalPacketData(ctx sdk.Context, portID, channelID string, bz []byte) (interface{}, string, error) {
	return nil, "", fmt.Errorf("unmarshal not supported")
}

// GetAppVersion returns the application version string.
func (m IBCModule) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return m.keeper.GetAppVersion(ctx, portID, channelID)
}
