package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
)

var (
	_ porttypes.IBCModule = (*IBCModule)(nil)
)

// IBCModule implements the IBC Module interface for the enclave module
type IBCModule struct {
	keeper IBCKeeper
}

// NewIBCModule creates a new IBCModule
func NewIBCModule(keeper IBCKeeper) IBCModule {
	return IBCModule{
		keeper: keeper,
	}
}

// OnChanOpenInit implements the IBCModule interface
func (m IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	// Validate port ID
	if portID != PortID {
		return "", ErrInvalidPort.Wrapf("expected %s, got %s", PortID, portID)
	}

	// Require unordered channels for enclave IBC
	if order != channeltypes.UNORDERED {
		return "", channeltypes.ErrInvalidChannelOrdering.Wrapf(
			"expected %s channel, got %s",
			channeltypes.UNORDERED,
			order,
		)
	}

	// Validate version
	if version != "" && version != Version {
		return "", ErrInvalidVersion.Wrapf("expected %s, got %s", Version, version)
	}

	m.keeper.Logger(ctx).Info("channel open init",
		"port", portID,
		"channel", channelID,
		"version", Version,
	)

	return Version, nil
}

// OnChanOpenTry implements the IBCModule interface
func (m IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	// Validate port ID
	if portID != PortID {
		return "", ErrInvalidPort.Wrapf("expected %s, got %s", PortID, portID)
	}

	// Require unordered channels
	if order != channeltypes.UNORDERED {
		return "", channeltypes.ErrInvalidChannelOrdering.Wrapf(
			"expected %s channel, got %s",
			channeltypes.UNORDERED,
			order,
		)
	}

	// Validate counterparty version
	if counterpartyVersion != Version {
		return "", ErrInvalidVersion.Wrapf("expected %s, got %s", Version, counterpartyVersion)
	}

	m.keeper.Logger(ctx).Info("channel open try",
		"port", portID,
		"channel", channelID,
		"counterparty_version", counterpartyVersion,
	)

	return Version, nil
}

// OnChanOpenAck implements the IBCModule interface
func (m IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID string,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	// Validate counterparty version
	if counterpartyVersion != Version {
		return ErrInvalidVersion.Wrapf("expected %s, got %s", Version, counterpartyVersion)
	}

	// Store channel metadata
	m.keeper.SetChannelMetadata(ctx, channelID, ChannelMetadata{
		Version: Version,
		Trusted: false, // Channels are untrusted by default, must be set via governance
	})

	m.keeper.Logger(ctx).Info("channel open ack",
		"port", portID,
		"channel", channelID,
		"counterparty_channel", counterpartyChannelID,
	)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeChannelOpened,
			sdk.NewAttribute(AttributeKeyChannelID, channelID),
			sdk.NewAttribute(AttributeKeyPortID, portID),
			sdk.NewAttribute(AttributeKeyVersion, Version),
		),
	)

	return nil
}

// OnChanOpenConfirm implements the IBCModule interface
func (m IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID string,
	channelID string,
) error {
	// Store channel metadata
	m.keeper.SetChannelMetadata(ctx, channelID, ChannelMetadata{
		Version: Version,
		Trusted: false,
	})

	m.keeper.Logger(ctx).Info("channel open confirm",
		"port", portID,
		"channel", channelID,
	)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeChannelOpened,
			sdk.NewAttribute(AttributeKeyChannelID, channelID),
			sdk.NewAttribute(AttributeKeyPortID, portID),
			sdk.NewAttribute(AttributeKeyVersion, Version),
		),
	)

	return nil
}

// OnChanCloseInit implements the IBCModule interface
func (m IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID string,
	channelID string,
) error {
	// Disallow user-initiated channel closing
	return ErrInvalidChannelState.Wrap("user cannot close channel")
}

// OnChanCloseConfirm implements the IBCModule interface
func (m IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID string,
	channelID string,
) error {
	// Clean up channel metadata
	m.keeper.DeleteChannelMetadata(ctx, channelID)

	m.keeper.Logger(ctx).Info("channel closed",
		"port", portID,
		"channel", channelID,
	)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeChannelClosed,
			sdk.NewAttribute(AttributeKeyChannelID, channelID),
			sdk.NewAttribute(AttributeKeyPortID, portID),
		),
	)

	return nil
}

// OnRecvPacket implements the IBCModule interface
func (m IBCModule) OnRecvPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	return m.keeper.OnRecvPacket(ctx, packet, relayer)
}

// OnAcknowledgementPacket implements the IBCModule interface
func (m IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return m.keeper.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface
func (m IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return m.keeper.OnTimeoutPacket(ctx, packet, relayer)
}

// OnChanUpgradeInit implements the IBCModule interface
func (m IBCModule) OnChanUpgradeInit(
	ctx sdk.Context,
	portID string,
	channelID string,
	proposedOrder channeltypes.Order,
	proposedConnectionHops []string,
	proposedVersion string,
) (string, error) {
	if proposedVersion != Version {
		return "", ErrInvalidVersion.Wrapf("expected %s, got %s", Version, proposedVersion)
	}

	if proposedOrder != channeltypes.UNORDERED {
		return "", channeltypes.ErrInvalidChannelOrdering.Wrapf(
			"expected %s channel, got %s",
			channeltypes.UNORDERED,
			proposedOrder,
		)
	}

	return Version, nil
}

// OnChanUpgradeTry implements the IBCModule interface
func (m IBCModule) OnChanUpgradeTry(
	ctx sdk.Context,
	portID string,
	channelID string,
	proposedOrder channeltypes.Order,
	proposedConnectionHops []string,
	counterpartyVersion string,
) (string, error) {
	if counterpartyVersion != Version {
		return "", ErrInvalidVersion.Wrapf("expected %s, got %s", Version, counterpartyVersion)
	}

	if proposedOrder != channeltypes.UNORDERED {
		return "", channeltypes.ErrInvalidChannelOrdering.Wrapf(
			"expected %s channel, got %s",
			channeltypes.UNORDERED,
			proposedOrder,
		)
	}

	return Version, nil
}

// OnChanUpgradeAck implements the IBCModule interface
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

// OnChanUpgradeOpen implements the IBCModule interface
func (m IBCModule) OnChanUpgradeOpen(
	ctx sdk.Context,
	portID string,
	channelID string,
	proposedOrder channeltypes.Order,
	proposedConnectionHops []string,
	proposedVersion string,
) {
	m.keeper.Logger(ctx).Info("channel upgraded",
		"port", portID,
		"channel", channelID,
		"version", proposedVersion,
	)
}

// UnmarshalPacketData unmarshals packet data
func (m IBCModule) UnmarshalPacketData(ctx sdk.Context, portID, channelID string, bz []byte) (interface{}, string, error) {
	var packetData EnclavePacketData
	if err := packetData.Validate(); err != nil {
		return nil, "", err
	}
	return packetData, Version, nil
}

// GetAppVersion returns the application version string
func (m IBCModule) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return m.keeper.GetAppVersion(ctx, portID, channelID)
}

// ValidateEnclavePacket validates an enclave packet
func ValidateEnclavePacket(packetData EnclavePacketData) error {
	if err := packetData.Validate(); err != nil {
		return fmt.Errorf("invalid packet data: %w", err)
	}
	return nil
}
