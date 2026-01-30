package ibc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"

	"github.com/virtengine/virtengine/x/enclave/keeper"
	"github.com/virtengine/virtengine/x/enclave/types"
)

// IBCKeeper implements the IBC module interface for the enclave module
type IBCKeeper struct {
	cdc           codec.BinaryCodec
	skey          storetypes.StoreKey
	enclaveKeeper keeper.IKeeper
	channelKeeper ChannelKeeper
	portKeeper    PortKeeper
}

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, portID, channelID string) (channeltypes.Channel, bool)
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
	SendPacket(ctx sdk.Context, sourcePort string, sourceChannel string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64, data []byte) (uint64, error)
}

// PortKeeper defines the expected IBC port keeper
type PortKeeper interface {
	BindPort(ctx sdk.Context, portID string)
	IsBound(ctx sdk.Context, portID string) bool
}

// NewIBCKeeper creates a new IBC keeper for the enclave module
func NewIBCKeeper(
	cdc codec.BinaryCodec,
	skey storetypes.StoreKey,
	enclaveKeeper keeper.IKeeper,
	channelKeeper ChannelKeeper,
	portKeeper PortKeeper,
) IBCKeeper {
	return IBCKeeper{
		cdc:           cdc,
		skey:          skey,
		enclaveKeeper: enclaveKeeper,
		channelKeeper: channelKeeper,
		portKeeper:    portKeeper,
	}
}

// Logger returns a module-specific logger
func (k IBCKeeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+ModuleName)
}

// ============================================================================
// Port Binding
// ============================================================================

// BindPort binds to the enclave port
func (k IBCKeeper) BindPort(ctx sdk.Context) error {
	if k.portKeeper.IsBound(ctx, PortID) {
		return ErrPortAlreadyBound
	}
	k.portKeeper.BindPort(ctx, PortID)
	return nil
}

// IsBound checks if the port is bound
func (k IBCKeeper) IsBound(ctx sdk.Context) bool {
	return k.portKeeper.IsBound(ctx, PortID)
}

// ============================================================================
// Channel Metadata
// ============================================================================

// SetChannelMetadata stores channel metadata
func (k IBCKeeper) SetChannelMetadata(ctx sdk.Context, channelID string, metadata ChannelMetadata) {
	store := ctx.KVStore(k.skey)
	bz, _ := json.Marshal(metadata)
	store.Set(ChannelMetadataKey(channelID), bz)
}

// GetChannelMetadata retrieves channel metadata
func (k IBCKeeper) GetChannelMetadata(ctx sdk.Context, channelID string) (ChannelMetadata, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(ChannelMetadataKey(channelID))
	if bz == nil {
		return ChannelMetadata{}, false
	}
	var metadata ChannelMetadata
	if err := json.Unmarshal(bz, &metadata); err != nil {
		return ChannelMetadata{}, false
	}
	return metadata, true
}

// DeleteChannelMetadata deletes channel metadata
func (k IBCKeeper) DeleteChannelMetadata(ctx sdk.Context, channelID string) {
	store := ctx.KVStore(k.skey)
	store.Delete(ChannelMetadataKey(channelID))
}

// IsChannelTrusted checks if a channel is trusted for sync operations
func (k IBCKeeper) IsChannelTrusted(ctx sdk.Context, channelID string) bool {
	metadata, found := k.GetChannelMetadata(ctx, channelID)
	if !found {
		return false
	}
	return metadata.Trusted
}

// SetChannelTrusted sets the trusted status of a channel
func (k IBCKeeper) SetChannelTrusted(ctx sdk.Context, channelID string, trusted bool) error {
	metadata, found := k.GetChannelMetadata(ctx, channelID)
	if !found {
		return ErrChannelNotFound
	}
	metadata.Trusted = trusted
	k.SetChannelMetadata(ctx, channelID, metadata)
	return nil
}

// ============================================================================
// Trusted Chains
// ============================================================================

// SetTrustedChain marks a chain as trusted
func (k IBCKeeper) SetTrustedChain(ctx sdk.Context, chainID string, trusted bool) {
	store := ctx.KVStore(k.skey)
	if trusted {
		store.Set(TrustedChainKey(chainID), []byte{1})
	} else {
		store.Delete(TrustedChainKey(chainID))
	}
}

// IsChainTrusted checks if a chain is trusted
func (k IBCKeeper) IsChainTrusted(ctx sdk.Context, chainID string) bool {
	store := ctx.KVStore(k.skey)
	return store.Has(TrustedChainKey(chainID))
}

// ============================================================================
// Federated Identities
// ============================================================================

// SetFederatedIdentity stores a federated identity
func (k IBCKeeper) SetFederatedIdentity(ctx sdk.Context, identity FederatedIdentity) error {
	validatorAddr, err := sdk.AccAddressFromBech32(identity.Identity.ValidatorAddress)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(identity)
	if err != nil {
		return err
	}
	store.Set(FederatedIdentityKey(identity.SourceChainID, validatorAddr), bz)
	return nil
}

// GetFederatedIdentity retrieves a federated identity
func (k IBCKeeper) GetFederatedIdentity(ctx sdk.Context, sourceChainID string, validatorAddr sdk.AccAddress) (FederatedIdentity, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(FederatedIdentityKey(sourceChainID, validatorAddr))
	if bz == nil {
		return FederatedIdentity{}, false
	}
	var identity FederatedIdentity
	if err := json.Unmarshal(bz, &identity); err != nil {
		return FederatedIdentity{}, false
	}
	return identity, true
}

// WithFederatedIdentities iterates over all federated identities from a chain
func (k IBCKeeper) WithFederatedIdentities(ctx sdk.Context, sourceChainID string, fn func(identity FederatedIdentity) bool) {
	store := ctx.KVStore(k.skey)
	prefix := FederatedIdentityPrefix(sourceChainID)
	iterator := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var identity FederatedIdentity
		if err := json.Unmarshal(iterator.Value(), &identity); err != nil {
			continue
		}
		if fn(identity) {
			break
		}
	}
}

// ============================================================================
// Federated Measurements
// ============================================================================

// SetFederatedMeasurement stores a federated measurement
func (k IBCKeeper) SetFederatedMeasurement(ctx sdk.Context, measurement FederatedMeasurement) {
	store := ctx.KVStore(k.skey)
	bz, _ := json.Marshal(measurement)
	store.Set(FederatedMeasurementKey(measurement.SourceChainID, measurement.Measurement.MeasurementHash), bz)
}

// GetFederatedMeasurement retrieves a federated measurement
func (k IBCKeeper) GetFederatedMeasurement(ctx sdk.Context, sourceChainID string, measurementHash []byte) (FederatedMeasurement, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(FederatedMeasurementKey(sourceChainID, measurementHash))
	if bz == nil {
		return FederatedMeasurement{}, false
	}
	var measurement FederatedMeasurement
	if err := json.Unmarshal(bz, &measurement); err != nil {
		return FederatedMeasurement{}, false
	}
	return measurement, true
}

// WithFederatedMeasurements iterates over all federated measurements from a chain
func (k IBCKeeper) WithFederatedMeasurements(ctx sdk.Context, sourceChainID string, fn func(measurement FederatedMeasurement) bool) {
	store := ctx.KVStore(k.skey)
	prefix := FederatedMeasurementPrefix(sourceChainID)
	iterator := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var measurement FederatedMeasurement
		if err := json.Unmarshal(iterator.Value(), &measurement); err != nil {
			continue
		}
		if fn(measurement) {
			break
		}
	}
}

// ============================================================================
// Packet Sending
// ============================================================================

// SendQueryEnclaveIdentityPacket sends a query for enclave identity
func (k IBCKeeper) SendQueryEnclaveIdentityPacket(
	ctx sdk.Context,
	sourceChannel string,
	timeoutHeight ibcexported.Height,
	timeoutTimestamp uint64,
	query QueryEnclaveIdentityPacket,
) (uint64, error) {
	packetData, err := NewEnclavePacketData(PacketTypeQueryEnclaveIdentity, query)
	if err != nil {
		return 0, err
	}

	return k.sendPacket(ctx, sourceChannel, timeoutHeight, timeoutTimestamp, packetData)
}

// SendQueryMeasurementAllowlistPacket sends a query for measurement allowlist
func (k IBCKeeper) SendQueryMeasurementAllowlistPacket(
	ctx sdk.Context,
	sourceChannel string,
	timeoutHeight ibcexported.Height,
	timeoutTimestamp uint64,
	query QueryMeasurementAllowlistPacket,
) (uint64, error) {
	packetData, err := NewEnclavePacketData(PacketTypeQueryMeasurementAllowlist, query)
	if err != nil {
		return 0, err
	}

	return k.sendPacket(ctx, sourceChannel, timeoutHeight, timeoutTimestamp, packetData)
}

// SendSyncMeasurementPacket sends a measurement sync request
func (k IBCKeeper) SendSyncMeasurementPacket(
	ctx sdk.Context,
	sourceChannel string,
	timeoutHeight ibcexported.Height,
	timeoutTimestamp uint64,
	sync SyncMeasurementPacket,
) (uint64, error) {
	// Validate the sync packet
	if err := sync.Validate(); err != nil {
		return 0, err
	}

	packetData, err := NewEnclavePacketData(PacketTypeSyncMeasurement, sync)
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
	packetData EnclavePacketData,
) (uint64, error) {
	// Convert timeout height to client types
	height, ok := timeoutHeight.(clienttypes.Height)
	if !ok {
		height = clienttypes.NewHeight(0, uint64(ctx.BlockHeight())+1000)
	}

	// Send the packet
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

	// Store pending packet data for acknowledgement handling
	k.setPendingPacket(ctx, sourceChannel, sequence, packetData)

	// Emit event
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

// ============================================================================
// Pending Packets
// ============================================================================

func (k IBCKeeper) setPendingPacket(ctx sdk.Context, channelID string, sequence uint64, data EnclavePacketData) {
	store := ctx.KVStore(k.skey)
	bz, _ := json.Marshal(data)
	store.Set(PendingPacketKey(channelID, sequence), bz)
}

func (k IBCKeeper) getPendingPacket(ctx sdk.Context, channelID string, sequence uint64) (EnclavePacketData, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(PendingPacketKey(channelID, sequence))
	if bz == nil {
		return EnclavePacketData{}, false
	}
	var data EnclavePacketData
	if err := json.Unmarshal(bz, &data); err != nil {
		return EnclavePacketData{}, false
	}
	return data, true
}

func (k IBCKeeper) deletePendingPacket(ctx sdk.Context, channelID string, sequence uint64) {
	store := ctx.KVStore(k.skey)
	store.Delete(PendingPacketKey(channelID, sequence))
}

// ============================================================================
// IBC Callbacks
// ============================================================================

// OnRecvPacket handles incoming packets
func (k IBCKeeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	var packetData EnclavePacketData
	if err := json.Unmarshal(packet.GetData(), &packetData); err != nil {
		return NewErrorAcknowledgement(ErrInvalidPacket.Wrap(err.Error()))
	}

	if err := packetData.Validate(); err != nil {
		return NewErrorAcknowledgement(err)
	}

	// Emit receive event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypePacketReceived,
			sdk.NewAttribute(AttributeKeyPacketType, string(packetData.Type)),
			sdk.NewAttribute(AttributeKeySourceChannel, packet.GetSourceChannel()),
			sdk.NewAttribute(AttributeKeyDestinationChannel, packet.GetDestChannel()),
			sdk.NewAttribute(AttributeKeySequence, fmt.Sprintf("%d", packet.GetSequence())),
		),
	)

	// Handle the packet based on type
	handler := NewPacketHandler(k, k.enclaveKeeper)
	return handler.HandlePacket(ctx, packet, packetData)
}

// OnAcknowledgementPacket handles packet acknowledgements
func (k IBCKeeper) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	var ack Acknowledgement
	if err := json.Unmarshal(acknowledgement, &ack); err != nil {
		return ErrAcknowledgementFailed.Wrap(err.Error())
	}

	// Get the original packet data
	pendingData, found := k.getPendingPacket(ctx, packet.GetSourceChannel(), packet.GetSequence())
	if !found {
		// Packet not found - might have been cleaned up already
		return nil
	}

	// Clean up pending packet
	k.deletePendingPacket(ctx, packet.GetSourceChannel(), packet.GetSequence())

	// Handle acknowledgement based on packet type
	handler := NewPacketHandler(k, k.enclaveKeeper)
	if err := handler.HandleAcknowledgement(ctx, packet, pendingData, ack); err != nil {
		k.Logger(ctx).Error("failed to handle acknowledgement", "error", err)
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypePacketAcknowledged,
			sdk.NewAttribute(AttributeKeyPacketType, string(pendingData.Type)),
			sdk.NewAttribute(AttributeKeySourceChannel, packet.GetSourceChannel()),
			sdk.NewAttribute(AttributeKeySequence, fmt.Sprintf("%d", packet.GetSequence())),
			sdk.NewAttribute(AttributeKeyAckSuccess, fmt.Sprintf("%t", ack.Success())),
		),
	)

	return nil
}

// OnTimeoutPacket handles packet timeouts
func (k IBCKeeper) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	// Get the original packet data
	pendingData, found := k.getPendingPacket(ctx, packet.GetSourceChannel(), packet.GetSequence())
	if !found {
		return nil
	}

	// Clean up pending packet
	k.deletePendingPacket(ctx, packet.GetSourceChannel(), packet.GetSequence())

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypePacketTimeout,
			sdk.NewAttribute(AttributeKeyPacketType, string(pendingData.Type)),
			sdk.NewAttribute(AttributeKeySourceChannel, packet.GetSourceChannel()),
			sdk.NewAttribute(AttributeKeySequence, fmt.Sprintf("%d", packet.GetSequence())),
		),
	)

	k.Logger(ctx).Warn("packet timed out",
		"type", pendingData.Type,
		"channel", packet.GetSourceChannel(),
		"sequence", packet.GetSequence(),
	)

	return nil
}

// ============================================================================
// Query Helpers
// ============================================================================

// GetActiveIdentitiesForQuery returns active enclave identities matching the query
func (k IBCKeeper) GetActiveIdentitiesForQuery(ctx sdk.Context, query QueryEnclaveIdentityPacket) []types.EnclaveIdentity {
	var identities []types.EnclaveIdentity
	currentHeight := ctx.BlockHeight()

	// If specific validator is requested
	if query.ValidatorAddress != "" {
		validatorAddr, err := sdk.AccAddressFromBech32(query.ValidatorAddress)
		if err != nil {
			return identities
		}
		identity, found := k.enclaveKeeper.GetEnclaveIdentity(ctx, validatorAddr)
		if found {
			if query.IncludeExpired || !types.IsIdentityExpired(identity, currentHeight) {
				if query.TEEType == "" || string(identity.TeeType) == query.TEEType {
					identities = append(identities, *identity)
				}
			}
		}
		return identities
	}

	// Return all matching identities
	k.enclaveKeeper.WithEnclaveIdentities(ctx, func(identity types.EnclaveIdentity) bool {
		if !query.IncludeExpired && types.IsIdentityExpired(&identity, currentHeight) {
			return false
		}
		if query.TEEType != "" && string(identity.TeeType) != query.TEEType {
			return false
		}
		if identity.Status == types.EnclaveIdentityStatusActive ||
			identity.Status == types.EnclaveIdentityStatusRotating {
			identities = append(identities, identity)
		}
		return false
	})

	return identities
}

// GetMeasurementsForQuery returns measurements matching the query
func (k IBCKeeper) GetMeasurementsForQuery(ctx sdk.Context, query QueryMeasurementAllowlistPacket) []types.MeasurementRecord {
	return k.enclaveKeeper.GetMeasurementAllowlist(ctx, query.TEEType, query.IncludeRevoked)
}

// ProcessSyncMeasurement processes a measurement sync from a remote chain
func (k IBCKeeper) ProcessSyncMeasurement(
	ctx sdk.Context,
	channelID string,
	sync SyncMeasurementPacket,
) error {
	// Check if channel is trusted
	if !k.IsChannelTrusted(ctx, channelID) {
		return ErrUntrustedChannel
	}

	// Check if chain is trusted
	if !k.IsChainTrusted(ctx, sync.SourceChainID) {
		return ErrFederatedIdentityNotTrusted
	}

	// Check if measurement already exists locally
	if _, found := k.enclaveKeeper.GetMeasurement(ctx, sync.Measurement.MeasurementHash); found {
		return ErrMeasurementAlreadyExists
	}

	// Store as federated measurement
	fedMeasurement := FederatedMeasurement{
		Measurement:     sync.Measurement,
		SourceChainID:   sync.SourceChainID,
		SourceChannelID: channelID,
		ReceivedHeight:  ctx.BlockHeight(),
		Trusted:         true,
	}
	k.SetFederatedMeasurement(ctx, fedMeasurement)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeMeasurementSynced,
			sdk.NewAttribute(AttributeKeyMeasurementHash, hex.EncodeToString(sync.Measurement.MeasurementHash)),
			sdk.NewAttribute(AttributeKeySourceChainID, sync.SourceChainID),
			sdk.NewAttribute(AttributeKeySourceChannel, channelID),
			sdk.NewAttribute(AttributeKeyTEEType, string(sync.Measurement.TeeType)),
		),
	)

	return nil
}

// GetAppVersion returns the application version
func (k IBCKeeper) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	channel, found := k.channelKeeper.GetChannel(ctx, portID, channelID)
	if !found {
		return "", false
	}
	return channel.Version, true
}
