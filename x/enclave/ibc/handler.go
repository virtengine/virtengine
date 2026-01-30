package ibc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"

	"github.com/virtengine/virtengine/x/enclave/keeper"
	"github.com/virtengine/virtengine/x/enclave/types"
)

// PacketHandler handles enclave IBC packets
type PacketHandler struct {
	ibcKeeper     IBCKeeper
	enclaveKeeper keeper.IKeeper
}

// NewPacketHandler creates a new packet handler
func NewPacketHandler(ibcKeeper IBCKeeper, enclaveKeeper keeper.IKeeper) PacketHandler {
	return PacketHandler{
		ibcKeeper:     ibcKeeper,
		enclaveKeeper: enclaveKeeper,
	}
}

// HandlePacket dispatches the packet to the appropriate handler based on type
func (h PacketHandler) HandlePacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data EnclavePacketData,
) ibcexported.Acknowledgement {
	switch data.Type {
	case PacketTypeQueryEnclaveIdentity:
		return h.handleQueryEnclaveIdentity(ctx, packet, data.Data)
	case PacketTypeQueryMeasurementAllowlist:
		return h.handleQueryMeasurementAllowlist(ctx, packet, data.Data)
	case PacketTypeSyncMeasurement:
		return h.handleSyncMeasurement(ctx, packet, data.Data)
	default:
		return NewErrorAcknowledgement(ErrUnknownPacketType.Wrapf("unknown packet type: %s", data.Type))
	}
}

// handleQueryEnclaveIdentity handles a query for enclave identity
func (h PacketHandler) handleQueryEnclaveIdentity(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data json.RawMessage,
) ibcexported.Acknowledgement {
	var query QueryEnclaveIdentityPacket
	if err := json.Unmarshal(data, &query); err != nil {
		return NewErrorAcknowledgement(ErrInvalidPacket.Wrap(err.Error()))
	}

	if err := query.Validate(); err != nil {
		return NewErrorAcknowledgement(err)
	}

	// Get matching identities
	identities := h.ibcKeeper.GetActiveIdentitiesForQuery(ctx, query)

	// Create response
	response := EnclaveIdentityResponse{
		Identities:  identities,
		ChainID:     ctx.ChainID(),
		BlockHeight: ctx.BlockHeight(),
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeIdentityQueried,
			sdk.NewAttribute(AttributeKeyValidatorAddress, query.ValidatorAddress),
			sdk.NewAttribute(AttributeKeyTEEType, query.TEEType),
			sdk.NewAttribute(AttributeKeyIdentityCount, fmt.Sprintf("%d", len(identities))),
			sdk.NewAttribute(AttributeKeySourceChannel, packet.GetSourceChannel()),
		),
	)

	return NewResultAcknowledgement(response)
}

// handleQueryMeasurementAllowlist handles a query for measurement allowlist
func (h PacketHandler) handleQueryMeasurementAllowlist(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data json.RawMessage,
) ibcexported.Acknowledgement {
	var query QueryMeasurementAllowlistPacket
	if err := json.Unmarshal(data, &query); err != nil {
		return NewErrorAcknowledgement(ErrInvalidPacket.Wrap(err.Error()))
	}

	if err := query.Validate(); err != nil {
		return NewErrorAcknowledgement(err)
	}

	// Get matching measurements
	measurements := h.ibcKeeper.GetMeasurementsForQuery(ctx, query)

	// Create response
	response := MeasurementAllowlistResponse{
		Measurements: measurements,
		ChainID:      ctx.ChainID(),
		BlockHeight:  ctx.BlockHeight(),
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeMeasurementQueried,
			sdk.NewAttribute(AttributeKeyTEEType, query.TEEType),
			sdk.NewAttribute(AttributeKeyMeasurementCount, fmt.Sprintf("%d", len(measurements))),
			sdk.NewAttribute(AttributeKeySourceChannel, packet.GetSourceChannel()),
		),
	)

	return NewResultAcknowledgement(response)
}

// handleSyncMeasurement handles a measurement sync request
func (h PacketHandler) handleSyncMeasurement(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data json.RawMessage,
) ibcexported.Acknowledgement {
	var sync SyncMeasurementPacket
	if err := json.Unmarshal(data, &sync); err != nil {
		return NewErrorAcknowledgement(ErrInvalidPacket.Wrap(err.Error()))
	}

	if err := sync.Validate(); err != nil {
		return NewErrorAcknowledgement(err)
	}

	// Process the sync
	err := h.ibcKeeper.ProcessSyncMeasurement(ctx, packet.GetDestChannel(), sync)
	if err != nil {
		return NewErrorAcknowledgement(err)
	}

	// Create acknowledgement
	ack := SyncMeasurementAck{
		Success:         true,
		MeasurementHash: hex.EncodeToString(sync.Measurement.MeasurementHash),
	}

	return NewResultAcknowledgement(ack)
}

// HandleAcknowledgement processes acknowledgements for sent packets
func (h PacketHandler) HandleAcknowledgement(
	ctx sdk.Context,
	packet channeltypes.Packet,
	originalData EnclavePacketData,
	ack Acknowledgement,
) error {
	if !ack.Success() {
		ctx.Logger().Warn("packet acknowledgement failed",
			"type", originalData.Type,
			"error", ack.Error,
			"channel", packet.GetSourceChannel(),
		)
		return nil
	}

	switch originalData.Type {
	case PacketTypeQueryEnclaveIdentity:
		return h.handleEnclaveIdentityAck(ctx, packet, ack)
	case PacketTypeQueryMeasurementAllowlist:
		return h.handleMeasurementAllowlistAck(ctx, packet, ack)
	case PacketTypeSyncMeasurement:
		return h.handleSyncMeasurementAck(ctx, packet, ack)
	default:
		return nil
	}
}

// handleEnclaveIdentityAck processes enclave identity query acknowledgement
func (h PacketHandler) handleEnclaveIdentityAck(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack Acknowledgement,
) error {
	var response EnclaveIdentityResponse
	if err := json.Unmarshal(ack.Result, &response); err != nil {
		return ErrAcknowledgementFailed.Wrap(err.Error())
	}

	// Store received identities as federated identities
	for _, identity := range response.Identities {
		fedIdentity := FederatedIdentity{
			Identity:        identity,
			SourceChainID:   response.ChainID,
			SourceChannelID: packet.GetSourceChannel(),
			ReceivedHeight:  ctx.BlockHeight(),
			Verified:        false, // Requires local verification
		}

		if err := h.ibcKeeper.SetFederatedIdentity(ctx, fedIdentity); err != nil {
			ctx.Logger().Error("failed to store federated identity",
				"validator", identity.ValidatorAddress,
				"chain", response.ChainID,
				"error", err,
			)
			continue
		}

		// Emit event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				EventTypeFederatedIdentityReceived,
				sdk.NewAttribute(AttributeKeyValidatorAddress, identity.ValidatorAddress),
				sdk.NewAttribute(AttributeKeySourceChainID, response.ChainID),
				sdk.NewAttribute(AttributeKeyTEEType, string(identity.TEEType)),
				sdk.NewAttribute(AttributeKeyMeasurementHash, hex.EncodeToString(identity.MeasurementHash)),
			),
		)
	}

	return nil
}

// handleMeasurementAllowlistAck processes measurement allowlist query acknowledgement
func (h PacketHandler) handleMeasurementAllowlistAck(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack Acknowledgement,
) error {
	var response MeasurementAllowlistResponse
	if err := json.Unmarshal(ack.Result, &response); err != nil {
		return ErrAcknowledgementFailed.Wrap(err.Error())
	}

	// Store received measurements as federated measurements
	for _, measurement := range response.Measurements {
		fedMeasurement := FederatedMeasurement{
			Measurement:     measurement,
			SourceChainID:   response.ChainID,
			SourceChannelID: packet.GetSourceChannel(),
			ReceivedHeight:  ctx.BlockHeight(),
			Trusted:         h.ibcKeeper.IsChainTrusted(ctx, response.ChainID),
		}

		h.ibcKeeper.SetFederatedMeasurement(ctx, fedMeasurement)
	}

	return nil
}

// handleSyncMeasurementAck processes measurement sync acknowledgement
func (h PacketHandler) handleSyncMeasurementAck(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack Acknowledgement,
) error {
	var syncAck SyncMeasurementAck
	if err := json.Unmarshal(ack.Result, &syncAck); err != nil {
		return ErrAcknowledgementFailed.Wrap(err.Error())
	}

	if !syncAck.Success {
		ctx.Logger().Warn("measurement sync failed on remote chain",
			"measurement", syncAck.MeasurementHash,
			"error", syncAck.Error,
			"channel", packet.GetSourceChannel(),
		)
	}

	return nil
}

// VerifyFederatedIdentity verifies a federated identity against local measurement allowlist
func (h PacketHandler) VerifyFederatedIdentity(ctx sdk.Context, identity FederatedIdentity) error {
	// Check if the measurement is in local allowlist
	if !h.enclaveKeeper.IsMeasurementAllowed(ctx, identity.Identity.MeasurementHash, ctx.BlockHeight()) {
		// Check federated measurements
		fedMeasurement, found := h.ibcKeeper.GetFederatedMeasurement(
			ctx,
			identity.SourceChainID,
			identity.Identity.MeasurementHash,
		)
		if !found {
			return types.ErrMeasurementNotAllowlisted
		}
		if !fedMeasurement.Trusted {
			return ErrFederatedIdentityNotTrusted
		}
	}

	// Validate the identity
	if err := identity.Identity.Validate(); err != nil {
		return err
	}

	return nil
}
