// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"

	"github.com/virtengine/virtengine/x/settlement/types"
)

// PendingPacket tracks packets awaiting acknowledgement.
type PendingPacket struct {
	PacketData SettlementPacketData `json:"packet_data"`
	Sender     string               `json:"sender,omitempty"`
	CreatedAt  time.Time            `json:"created_at"`
}

// HandshakeRecord tracks handshake start time for timeouts.
type HandshakeRecord struct {
	StartHeight int64     `json:"start_height"`
	StartTime   time.Time `json:"start_time"`
}

// SettlementKeeper defines the settlement keeper methods used by IBC.
type SettlementKeeper interface {
	CreateEscrow(ctx sdk.Context, orderID string, depositor sdk.AccAddress, amount sdk.Coins, expiresIn time.Duration, conditions []types.ReleaseCondition) (string, error)
	ReleaseEscrow(ctx sdk.Context, escrowID string, reason string) error
	RefundEscrow(ctx sdk.Context, escrowID string, reason string) error
	GetEscrow(ctx sdk.Context, escrowID string) (types.EscrowAccount, bool)
	GetEscrowByOrder(ctx sdk.Context, orderID string) (types.EscrowAccount, bool)
	SetSettlement(ctx sdk.Context, settlement types.SettlementRecord) error
	GetSettlement(ctx sdk.Context, settlementID string) (types.SettlementRecord, bool)
}

// ChannelKeeper defines the expected IBC channel keeper.
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, portID, channelID string) (channeltypes.Channel, bool)
	SendPacket(ctx sdk.Context, sourcePort, sourceChannel string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64, data []byte) (uint64, error)
}

// PortKeeper defines the expected IBC port keeper.
// Port binding is handled by routers in IBC v10; keeper is retained for compatibility.
type PortKeeper interface{}

// IBCKeeper implements settlement IBC logic.
type IBCKeeper struct {
	cdc              codec.BinaryCodec
	storeKey         storetypes.StoreKey
	settlementKeeper SettlementKeeper
	channelKeeper    ChannelKeeper
	portKeeper       PortKeeper
	relayerHooks     RelayerHooks
}

// NewIBCKeeper creates a new settlement IBC keeper.
func NewIBCKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	settlementKeeper SettlementKeeper,
	channelKeeper ChannelKeeper,
	portKeeper PortKeeper,
) IBCKeeper {
	return IBCKeeper{
		cdc:              cdc,
		storeKey:         storeKey,
		settlementKeeper: settlementKeeper,
		channelKeeper:    channelKeeper,
		portKeeper:       portKeeper,
		relayerHooks:     NoOpRelayerHooks{},
	}
}

// SetRelayerHooks installs relayer hooks.
func (k *IBCKeeper) SetRelayerHooks(hooks RelayerHooks) {
	if hooks == nil {
		k.relayerHooks = NoOpRelayerHooks{}
		return
	}
	k.relayerHooks = hooks
}

// Logger returns a module-specific logger.
func (k IBCKeeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+ModuleName)
}

// BindPort is a no-op for IBC v10 (router-based port binding).
func (k IBCKeeper) BindPort(_ sdk.Context) error {
	return nil
}

// IsBound always returns true for IBC v10 (router-based port binding).
func (k IBCKeeper) IsBound(_ sdk.Context) bool {
	return true
}

// StoreHandshakeRecord stores the handshake start info for timeout checks.
func (k IBCKeeper) StoreHandshakeRecord(ctx sdk.Context, channelID string) {
	store := ctx.KVStore(k.storeKey)
	record := HandshakeRecord{StartHeight: ctx.BlockHeight(), StartTime: ctx.BlockTime()}
	bz, err := json.Marshal(record)
	if err != nil {
		return
	}
	store.Set(HandshakeKey(channelID), bz)
}

// ClearHandshakeRecord removes handshake tracking.
func (k IBCKeeper) ClearHandshakeRecord(ctx sdk.Context, channelID string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(HandshakeKey(channelID))
}

// CheckHandshakeTimeout returns error if handshake exceeded limits.
func (k IBCKeeper) CheckHandshakeTimeout(ctx sdk.Context, channelID string) error {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(HandshakeKey(channelID))
	if bz == nil {
		return nil
	}

	var record HandshakeRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return nil
	}

	if record.StartHeight > 0 && ctx.BlockHeight() > record.StartHeight+100 {
		return ErrHandshakeTimedOut
	}
	if !record.StartTime.IsZero() && ctx.BlockTime().After(record.StartTime.Add(15*time.Minute)) {
		return ErrHandshakeTimedOut
	}
	return nil
}

// SendEscrowDepositPacket sends an escrow deposit packet.
func (k IBCKeeper) SendEscrowDepositPacket(
	ctx sdk.Context,
	sourceChannel string,
	timeoutHeight ibcexported.Height,
	timeoutTimestamp uint64,
	deposit EscrowDepositPacket,
) (uint64, error) {
	if err := deposit.Validate(); err != nil {
		return 0, err
	}

	packetData, err := NewPacketData(PacketTypeEscrowDeposit, deposit)
	if err != nil {
		return 0, err
	}

	return k.sendPacket(ctx, sourceChannel, timeoutHeight, timeoutTimestamp, packetData)
}

// SendEscrowReleasePacket sends an escrow release/refund packet.
func (k IBCKeeper) SendEscrowReleasePacket(
	ctx sdk.Context,
	sourceChannel string,
	timeoutHeight ibcexported.Height,
	timeoutTimestamp uint64,
	release EscrowReleasePacket,
) (uint64, error) {
	if err := release.Validate(); err != nil {
		return 0, err
	}

	packetData, err := NewPacketData(PacketTypeEscrowRelease, release)
	if err != nil {
		return 0, err
	}

	return k.sendPacket(ctx, sourceChannel, timeoutHeight, timeoutTimestamp, packetData)
}

// SendSettlementRecordPacket sends a settlement record packet.
func (k IBCKeeper) SendSettlementRecordPacket(
	ctx sdk.Context,
	sourceChannel string,
	timeoutHeight ibcexported.Height,
	timeoutTimestamp uint64,
	record SettlementRecordPacket,
) (uint64, error) {
	if err := record.Validate(); err != nil {
		return 0, err
	}

	packetData, err := NewPacketData(PacketTypeSettlementRecord, record)
	if err != nil {
		return 0, err
	}

	return k.sendPacket(ctx, sourceChannel, timeoutHeight, timeoutTimestamp, packetData)
}

func (k IBCKeeper) sendPacket(
	ctx sdk.Context,
	sourceChannel string,
	timeoutHeight ibcexported.Height,
	timeoutTimestamp uint64,
	packetData SettlementPacketData,
) (uint64, error) {
	if err := packetData.Validate(); err != nil {
		return 0, err
	}

	height, ok := timeoutHeight.(clienttypes.Height)
	if !ok || height.IsZero() {
		height = clienttypes.NewHeight(0, uint64(ctx.BlockHeight())+DefaultTimeoutHeightDelta)
	}

	if timeoutTimestamp == 0 {
		timeoutTimestamp = uint64(ctx.BlockTime().UnixNano()) + DefaultTimeoutTimestampDelta
	}

	sequence, err := k.channelKeeper.SendPacket(
		ctx,
		PortID,
		sourceChannel,
		height,
		timeoutTimestamp,
		packetData.GetBytes(),
	)
	if err != nil {
		return 0, err
	}

	k.setPendingPacket(ctx, sourceChannel, sequence, PendingPacket{
		PacketData: packetData,
		CreatedAt:  ctx.BlockTime(),
	})

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypePacketSent,
			sdk.NewAttribute(AttributeKeyPacketType, string(packetData.Type)),
			sdk.NewAttribute(AttributeKeySourceChannel, sourceChannel),
			sdk.NewAttribute(AttributeKeySequence, fmt.Sprintf("%d", sequence)),
		),
	)

	return sequence, nil
}

// OnRecvPacket handles incoming packets.
func (k IBCKeeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	var packetData SettlementPacketData
	if err := json.Unmarshal(packet.GetData(), &packetData); err != nil {
		return NewErrorAcknowledgement(ErrInvalidPacket.Wrap(err.Error()))
	}

	if err := packetData.Validate(); err != nil {
		return NewErrorAcknowledgement(err)
	}

	if err := k.CheckRateLimit(ctx, relayer, packetData.Type); err != nil {
		return NewErrorAcknowledgement(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypePacketReceived,
			sdk.NewAttribute(AttributeKeyPacketType, string(packetData.Type)),
			sdk.NewAttribute(AttributeKeySourceChannel, packet.GetSourceChannel()),
			sdk.NewAttribute(AttributeKeyDestinationChannel, packet.GetDestChannel()),
			sdk.NewAttribute(AttributeKeySequence, fmt.Sprintf("%d", packet.GetSequence())),
			sdk.NewAttribute(AttributeKeyRelayer, relayer.String()),
		),
	)

	k.relayerHooks.OnPacketReceived(ctx, relayer, packet, packetData.Type)

	switch packetData.Type {
	case PacketTypeEscrowDeposit:
		var deposit EscrowDepositPacket
		if err := json.Unmarshal(packetData.Data, &deposit); err != nil {
			return NewErrorAcknowledgement(ErrInvalidPacket.Wrap(err.Error()))
		}
		ack, err := k.handleEscrowDeposit(ctx, packet, deposit)
		if err != nil {
			return NewErrorAcknowledgement(err)
		}
		return NewResultAcknowledgement(ack)

	case PacketTypeEscrowRelease:
		var release EscrowReleasePacket
		if err := json.Unmarshal(packetData.Data, &release); err != nil {
			return NewErrorAcknowledgement(ErrInvalidPacket.Wrap(err.Error()))
		}
		ack, err := k.handleEscrowRelease(ctx, packet, release)
		if err != nil {
			return NewErrorAcknowledgement(err)
		}
		return NewResultAcknowledgement(ack)

	case PacketTypeSettlementRecord:
		var record SettlementRecordPacket
		if err := json.Unmarshal(packetData.Data, &record); err != nil {
			return NewErrorAcknowledgement(ErrInvalidPacket.Wrap(err.Error()))
		}
		ack, err := k.handleSettlementRecord(ctx, packet, record)
		if err != nil {
			return NewErrorAcknowledgement(err)
		}
		return NewResultAcknowledgement(ack)
	}

	return NewErrorAcknowledgement(ErrInvalidPacket.Wrap("unsupported packet type"))
}

// OnAcknowledgementPacket handles packet acknowledgements.
func (k IBCKeeper) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	var ack Acknowledgement
	if err := json.Unmarshal(acknowledgement, &ack); err != nil {
		return ErrInvalidPacket.Wrap(err.Error())
	}

	pending, found := k.getPendingPacket(ctx, packet.GetSourceChannel(), packet.GetSequence())
	if !found {
		return nil
	}

	k.deletePendingPacket(ctx, packet.GetSourceChannel(), packet.GetSequence())
	k.storeAck(ctx, packet.GetSourceChannel(), packet.GetSequence(), ack)

	k.relayerHooks.OnPacketAcknowledged(ctx, relayer, packet, pending.PacketData.Type, ack.Success())

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypePacketAcknowledged,
			sdk.NewAttribute(AttributeKeyPacketType, string(pending.PacketData.Type)),
			sdk.NewAttribute(AttributeKeySourceChannel, packet.GetSourceChannel()),
			sdk.NewAttribute(AttributeKeySequence, fmt.Sprintf("%d", packet.GetSequence())),
			sdk.NewAttribute(AttributeKeyAckSuccess, fmt.Sprintf("%t", ack.Success())),
			sdk.NewAttribute(AttributeKeyRelayer, relayer.String()),
		),
	)

	return nil
}

// OnTimeoutPacket handles packet timeouts.
func (k IBCKeeper) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	pending, found := k.getPendingPacket(ctx, packet.GetSourceChannel(), packet.GetSequence())
	if !found {
		return nil
	}

	k.deletePendingPacket(ctx, packet.GetSourceChannel(), packet.GetSequence())
	k.storeTimeout(ctx, packet.GetSourceChannel(), packet.GetSequence())

	k.relayerHooks.OnPacketTimeout(ctx, relayer, packet, pending.PacketData.Type)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypePacketTimeout,
			sdk.NewAttribute(AttributeKeyPacketType, string(pending.PacketData.Type)),
			sdk.NewAttribute(AttributeKeySourceChannel, packet.GetSourceChannel()),
			sdk.NewAttribute(AttributeKeySequence, fmt.Sprintf("%d", packet.GetSequence())),
			sdk.NewAttribute(AttributeKeyRelayer, relayer.String()),
		),
	)

	k.Logger(ctx).Warn("packet timed out",
		"type", pending.PacketData.Type,
		"channel", packet.GetSourceChannel(),
		"sequence", packet.GetSequence(),
	)

	return nil
}

func (k IBCKeeper) handleEscrowDeposit(ctx sdk.Context, packet channeltypes.Packet, deposit EscrowDepositPacket) (EscrowDepositAck, error) {
	if err := deposit.Validate(); err != nil {
		return EscrowDepositAck{}, err
	}

	if existing, found := k.settlementKeeper.GetEscrowByOrder(ctx, deposit.OrderID); found {
		return EscrowDepositAck{EscrowID: existing.EscrowID, OrderID: deposit.OrderID, Status: "already_exists"}, nil
	}

	depositor, err := sdk.AccAddressFromBech32(deposit.Depositor)
	if err != nil {
		return EscrowDepositAck{}, err
	}

	expiresIn := time.Duration(deposit.ExpiresInSeconds) * time.Second
	escrowID, err := k.settlementKeeper.CreateEscrow(ctx, deposit.OrderID, depositor, deposit.Amount, expiresIn, deposit.Conditions)
	if err != nil {
		return EscrowDepositAck{}, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeEscrowCreated,
			sdk.NewAttribute(AttributeKeyEscrowID, escrowID),
			sdk.NewAttribute(AttributeKeyOrderID, deposit.OrderID),
		),
	)

	return EscrowDepositAck{EscrowID: escrowID, OrderID: deposit.OrderID, Status: "created"}, nil
}

func (k IBCKeeper) handleEscrowRelease(ctx sdk.Context, packet channeltypes.Packet, release EscrowReleasePacket) (EscrowReleaseAck, error) {
	if err := release.Validate(); err != nil {
		return EscrowReleaseAck{}, err
	}

	switch release.ReleaseType {
	case ReleaseTypeRefund:
		if err := k.settlementKeeper.RefundEscrow(ctx, release.EscrowID, release.Reason); err != nil {
			return EscrowReleaseAck{}, err
		}
	case ReleaseTypeRelease:
		if err := k.settlementKeeper.ReleaseEscrow(ctx, release.EscrowID, release.Reason); err != nil {
			return EscrowReleaseAck{}, err
		}
	default:
		return EscrowReleaseAck{}, ErrInvalidPacket.Wrap("unknown release type")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeEscrowReleased,
			sdk.NewAttribute(AttributeKeyEscrowID, release.EscrowID),
			sdk.NewAttribute(AttributeKeyOrderID, release.OrderID),
			sdk.NewAttribute(AttributeKeyReleaseType, string(release.ReleaseType)),
		),
	)

	return EscrowReleaseAck{EscrowID: release.EscrowID, Status: "released"}, nil
}

func (k IBCKeeper) handleSettlementRecord(ctx sdk.Context, packet channeltypes.Packet, record SettlementRecordPacket) (SettlementRecordAck, error) {
	if err := record.Validate(); err != nil {
		return SettlementRecordAck{}, err
	}

	if _, found := k.settlementKeeper.GetSettlement(ctx, record.Record.SettlementID); found {
		return SettlementRecordAck{SettlementID: record.Record.SettlementID, Status: "already_exists"}, nil
	}

	if err := k.settlementKeeper.SetSettlement(ctx, record.Record); err != nil {
		return SettlementRecordAck{}, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeSettlementRecorded,
			sdk.NewAttribute(AttributeKeySettlementID, record.Record.SettlementID),
			sdk.NewAttribute(AttributeKeyOrderID, record.Record.OrderID),
		),
	)

	return SettlementRecordAck{SettlementID: record.Record.SettlementID, Status: "recorded"}, nil
}

func (k IBCKeeper) setPendingPacket(ctx sdk.Context, channelID string, sequence uint64, packet PendingPacket) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(packet)
	if err != nil {
		return
	}
	store.Set(PendingPacketKey(channelID, sequence), bz)
}

func (k IBCKeeper) getPendingPacket(ctx sdk.Context, channelID string, sequence uint64) (PendingPacket, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(PendingPacketKey(channelID, sequence))
	if bz == nil {
		return PendingPacket{}, false
	}
	var packet PendingPacket
	if err := json.Unmarshal(bz, &packet); err != nil {
		return PendingPacket{}, false
	}
	return packet, true
}

func (k IBCKeeper) deletePendingPacket(ctx sdk.Context, channelID string, sequence uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(PendingPacketKey(channelID, sequence))
}

func (k IBCKeeper) storeAck(ctx sdk.Context, channelID string, sequence uint64, ack Acknowledgement) {
	store := ctx.KVStore(k.storeKey)
	store.Set(AckPacketKey(channelID, sequence), ack.GetBytes())
}

func (k IBCKeeper) storeTimeout(ctx sdk.Context, channelID string, sequence uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(TimeoutPacketKey(channelID, sequence), []byte{1})
}

// GetAppVersion returns the channel version.
func (k IBCKeeper) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	channel, found := k.channelKeeper.GetChannel(ctx, portID, channelID)
	if !found {
		return "", false
	}
	return channel.Version, true
}
