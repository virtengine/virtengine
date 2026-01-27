package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/enclave/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

// RegisterEnclaveIdentity handles MsgRegisterEnclaveIdentity
func (m msgServer) RegisterEnclaveIdentity(goCtx context.Context, msg *types.MsgRegisterEnclaveIdentity) (*types.MsgRegisterEnclaveIdentityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := m.keeper.GetParams(ctx)

	// Create the enclave identity
	identity := &types.EnclaveIdentity{
		ValidatorAddress: msg.ValidatorAddress,
		TEEType:          msg.TEEType,
		MeasurementHash:  msg.MeasurementHash,
		SignerHash:       msg.SignerHash,
		EncryptionPubKey: msg.EncryptionPubKey,
		SigningPubKey:    msg.SigningPubKey,
		AttestationQuote: msg.AttestationQuote,
		AttestationChain: msg.AttestationChain,
		ISVProdID:        msg.ISVProdID,
		ISVSVN:           msg.ISVSVN,
		QuoteVersion:     msg.QuoteVersion,
		DebugMode:        false, // Always false for registration
		ExpiryHeight:     ctx.BlockHeight() + params.DefaultExpiryBlocks,
	}

	if err := m.keeper.RegisterEnclaveIdentity(ctx, identity); err != nil {
		return nil, err
	}

	return &types.MsgRegisterEnclaveIdentityResponse{
		KeyFingerprint: identity.KeyFingerprint(),
		ExpiryHeight:   identity.ExpiryHeight,
	}, nil
}

// RotateEnclaveIdentity handles MsgRotateEnclaveIdentity
func (m msgServer) RotateEnclaveIdentity(goCtx context.Context, msg *types.MsgRotateEnclaveIdentity) (*types.MsgRotateEnclaveIdentityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validatorAddr, err := sdk.AccAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, types.ErrInvalidEnclaveIdentity.Wrapf("invalid validator address: %v", err)
	}

	// Get existing identity
	existingIdentity, exists := m.keeper.GetEnclaveIdentity(ctx, validatorAddr)
	if !exists {
		return nil, types.ErrEnclaveIdentityNotFound
	}

	// Create rotation record
	rotation := &types.KeyRotationRecord{
		ValidatorAddress:   msg.ValidatorAddress,
		Epoch:              existingIdentity.Epoch + 1,
		OldKeyFingerprint:  existingIdentity.KeyFingerprint(),
		NewKeyFingerprint:  "", // Will be set after creating new identity
		OverlapStartHeight: ctx.BlockHeight(),
		OverlapEndHeight:   ctx.BlockHeight() + msg.OverlapBlocks,
	}

	// Update identity with new keys
	newMeasurement := existingIdentity.MeasurementHash
	if len(msg.NewMeasurementHash) > 0 {
		newMeasurement = msg.NewMeasurementHash
	}

	updatedIdentity := &types.EnclaveIdentity{
		ValidatorAddress: msg.ValidatorAddress,
		TEEType:          existingIdentity.TEEType,
		MeasurementHash:  newMeasurement,
		SignerHash:       existingIdentity.SignerHash,
		EncryptionPubKey: msg.NewEncryptionPubKey,
		SigningPubKey:    msg.NewSigningPubKey,
		AttestationQuote: msg.NewAttestationQuote,
		AttestationChain: msg.NewAttestationChain,
		ISVProdID:        existingIdentity.ISVProdID,
		ISVSVN:           msg.NewISVSVN,
		QuoteVersion:     existingIdentity.QuoteVersion,
		DebugMode:        false,
		Epoch:            existingIdentity.Epoch + 1,
		ExpiryHeight:     existingIdentity.ExpiryHeight,
		RegisteredAt:     existingIdentity.RegisteredAt,
		Status:           types.EnclaveIdentityStatusRotating,
	}

	rotation.NewKeyFingerprint = updatedIdentity.KeyFingerprint()

	// Initiate rotation
	if err := m.keeper.InitiateKeyRotation(ctx, rotation); err != nil {
		return nil, err
	}

	// Update identity
	if err := m.keeper.UpdateEnclaveIdentity(ctx, updatedIdentity); err != nil {
		return nil, err
	}

	return &types.MsgRotateEnclaveIdentityResponse{
		NewKeyFingerprint:  rotation.NewKeyFingerprint,
		OverlapStartHeight: rotation.OverlapStartHeight,
		OverlapEndHeight:   rotation.OverlapEndHeight,
	}, nil
}

// ProposeMeasurement handles MsgProposeMeasurement
func (m msgServer) ProposeMeasurement(goCtx context.Context, msg *types.MsgProposeMeasurement) (*types.MsgProposeMeasurementResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check authority
	if msg.Authority != m.keeper.GetAuthority() {
		return nil, types.ErrUnauthorized.Wrapf("invalid authority: expected %s, got %s", m.keeper.GetAuthority(), msg.Authority)
	}

	var expiryHeight int64
	if msg.ExpiryBlocks > 0 {
		expiryHeight = ctx.BlockHeight() + msg.ExpiryBlocks
	}

	measurement := &types.MeasurementRecord{
		MeasurementHash: msg.MeasurementHash,
		TEEType:         msg.TEEType,
		Description:     msg.Description,
		MinISVSVN:       msg.MinISVSVN,
		ExpiryHeight:    expiryHeight,
	}

	if err := m.keeper.AddMeasurement(ctx, measurement); err != nil {
		return nil, err
	}

	return &types.MsgProposeMeasurementResponse{
		MeasurementHash: measurement.MeasurementHashHex(),
	}, nil
}

// RevokeMeasurement handles MsgRevokeMeasurement
func (m msgServer) RevokeMeasurement(goCtx context.Context, msg *types.MsgRevokeMeasurement) (*types.MsgRevokeMeasurementResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check authority
	if msg.Authority != m.keeper.GetAuthority() {
		return nil, types.ErrUnauthorized.Wrapf("invalid authority: expected %s, got %s", m.keeper.GetAuthority(), msg.Authority)
	}

	// In production, this would be tied to a governance proposal
	if err := m.keeper.RevokeMeasurement(ctx, msg.MeasurementHash, msg.Reason, 0); err != nil {
		return nil, err
	}

	return &types.MsgRevokeMeasurementResponse{}, nil
}
