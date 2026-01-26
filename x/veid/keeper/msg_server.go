package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"pkg.akt.dev/node/x/veid/types"
)

// Error message constants
const (
	errMsgInvalidSenderAddr  = "invalid sender address"
	errMsgInvalidAccountAddr = "invalid account address"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the veid MsgServer interface
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

var _ types.MsgServer = msgServer{}

// UploadScope handles uploading a new identity scope
func (ms msgServer) UploadScope(ctx sdk.Context, msg *types.MsgUploadScope) (*types.MsgUploadScopeResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	params := ms.keeper.GetParams(ctx)

	// Validate client is approved
	if params.RequireClientSignature && !ms.keeper.IsClientApproved(ctx, msg.ClientID) {
		return nil, types.ErrClientNotApproved.Wrapf("client %s is not approved", msg.ClientID)
	}

	// Create upload metadata
	metadata := types.NewUploadMetadata(
		msg.Salt,
		msg.DeviceFingerprint,
		msg.ClientID,
		msg.ClientSignature,
		msg.UserSignature,
		msg.PayloadHash,
	)
	metadata.CaptureTimestamp = msg.CaptureTimestamp
	metadata.GeoHint = msg.GeoHint

	// Validate signatures
	if err := ms.keeper.ValidateUploadSignatures(ctx, sender, metadata); err != nil {
		return nil, err
	}

	// Create the scope
	scope := types.NewIdentityScope(
		msg.ScopeID,
		msg.ScopeType,
		msg.EncryptedPayload,
		*metadata,
		ctx.BlockTime(),
	)

	// Upload the scope
	if err := ms.keeper.UploadScope(ctx, sender, scope); err != nil {
		return nil, err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventScopeUploaded{
		AccountAddress:    sender.String(),
		ScopeID:           msg.ScopeID,
		ScopeType:         string(msg.ScopeType),
		Version:           types.ScopeSchemaVersion,
		ClientID:          msg.ClientID,
		PayloadHash:       metadata.PayloadHashHex(),
		DeviceFingerprint: msg.DeviceFingerprint,
		UploadedAt:        ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgUploadScopeResponse{
		ScopeID:    msg.ScopeID,
		Status:     types.VerificationStatusPending,
		UploadedAt: ctx.BlockTime().Unix(),
	}, nil
}

// RevokeScope handles revoking an identity scope
func (ms msgServer) RevokeScope(ctx sdk.Context, msg *types.MsgRevokeScope) (*types.MsgRevokeScopeResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Revoke the scope
	if err := ms.keeper.RevokeScope(ctx, sender, msg.ScopeID, msg.Reason); err != nil {
		return nil, err
	}

	// Get scope for event data
	scope, _ := ms.keeper.GetScope(ctx, sender, msg.ScopeID)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventScopeRevoked{
		AccountAddress: sender.String(),
		ScopeID:        msg.ScopeID,
		ScopeType:      string(scope.ScopeType),
		Reason:         msg.Reason,
		RevokedAt:      ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgRevokeScopeResponse{
		ScopeID:   msg.ScopeID,
		RevokedAt: ctx.BlockTime().Unix(),
	}, nil
}

// RequestVerification handles requesting verification for a scope
func (ms msgServer) RequestVerification(ctx sdk.Context, msg *types.MsgRequestVerification) (*types.MsgRequestVerificationResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Get the scope
	scope, found := ms.keeper.GetScope(ctx, sender, msg.ScopeID)
	if !found {
		return nil, types.ErrScopeNotFound.Wrapf("scope %s not found", msg.ScopeID)
	}

	// Check if scope can be verified
	if !scope.CanBeVerified() {
		if scope.Revoked {
			return nil, types.ErrScopeRevoked.Wrapf("scope %s is revoked", msg.ScopeID)
		}
		if scope.Status == types.VerificationStatusInProgress {
			return nil, types.ErrVerificationInProgress.Wrapf("verification already in progress for scope %s", msg.ScopeID)
		}
		return nil, types.ErrInvalidStatusTransition.Wrapf("scope %s cannot be verified in status %s", msg.ScopeID, scope.Status)
	}

	// Update status to in progress
	err = ms.keeper.UpdateVerificationStatus(
		ctx,
		sender,
		msg.ScopeID,
		types.VerificationStatusInProgress,
		"verification requested",
		"",
	)
	if err != nil {
		return nil, err
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventVerificationRequested{
		AccountAddress: sender.String(),
		ScopeID:        msg.ScopeID,
		ScopeType:      string(scope.ScopeType),
		RequestedAt:    ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgRequestVerificationResponse{
		ScopeID:     msg.ScopeID,
		Status:      types.VerificationStatusInProgress,
		RequestedAt: ctx.BlockTime().Unix(),
	}, nil
}

// UpdateVerificationStatus handles validators updating verification status
func (ms msgServer) UpdateVerificationStatus(ctx sdk.Context, msg *types.MsgUpdateVerificationStatus) (*types.MsgUpdateVerificationStatusResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	accountAddr, err := sdk.AccAddressFromBech32(msg.AccountAddress)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidAccountAddr)
	}

	// TODO: Validate that sender is a validator
	// This would check against the staking module's validator set
	// For now, we allow any sender for development purposes

	// Get current scope for previous status
	scope, found := ms.keeper.GetScope(ctx, accountAddr, msg.ScopeID)
	if !found {
		return nil, types.ErrScopeNotFound.Wrapf("scope %s not found", msg.ScopeID)
	}
	previousStatus := scope.Status

	// Update verification status
	err = ms.keeper.UpdateVerificationStatus(
		ctx,
		accountAddr,
		msg.ScopeID,
		msg.NewStatus,
		msg.Reason,
		sender.String(),
	)
	if err != nil {
		return nil, err
	}

	// Emit appropriate event
	if msg.NewStatus == types.VerificationStatusVerified {
		err = ctx.EventManager().EmitTypedEvent(&types.EventScopeVerified{
			AccountAddress:   msg.AccountAddress,
			ScopeID:          msg.ScopeID,
			ScopeType:        string(scope.ScopeType),
			ValidatorAddress: sender.String(),
			VerifiedAt:       ctx.BlockTime().Unix(),
		})
	} else if msg.NewStatus == types.VerificationStatusRejected {
		err = ctx.EventManager().EmitTypedEvent(&types.EventScopeRejected{
			AccountAddress:   msg.AccountAddress,
			ScopeID:          msg.ScopeID,
			ScopeType:        string(scope.ScopeType),
			Reason:           msg.Reason,
			ValidatorAddress: sender.String(),
			RejectedAt:       ctx.BlockTime().Unix(),
		})
	}
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateVerificationStatusResponse{
		ScopeID:        msg.ScopeID,
		PreviousStatus: previousStatus,
		NewStatus:      msg.NewStatus,
		UpdatedAt:      ctx.BlockTime().Unix(),
	}, nil
}

// UpdateScore handles validators updating identity score
func (ms msgServer) UpdateScore(ctx sdk.Context, msg *types.MsgUpdateScore) (*types.MsgUpdateScoreResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	accountAddr, err := sdk.AccAddressFromBech32(msg.AccountAddress)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidAccountAddr)
	}

	// TODO: Validate that sender is a validator
	// This would check against the staking module's validator set
	_ = sender

	// Get current record for previous values
	record, found := ms.keeper.GetIdentityRecord(ctx, accountAddr)
	if !found {
		return nil, types.ErrIdentityRecordNotFound.Wrapf("identity record not found for %s", msg.AccountAddress)
	}
	previousScore := record.CurrentScore
	previousTier := record.Tier

	// Update score
	err = ms.keeper.UpdateScore(ctx, accountAddr, msg.NewScore, msg.ScoreVersion)
	if err != nil {
		return nil, err
	}

	// Get updated record for new tier
	updatedRecord, _ := ms.keeper.GetIdentityRecord(ctx, accountAddr)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventScoreUpdated{
		AccountAddress: msg.AccountAddress,
		PreviousScore:  previousScore,
		NewScore:       msg.NewScore,
		ScoreVersion:   msg.ScoreVersion,
		PreviousTier:   string(previousTier),
		NewTier:        string(updatedRecord.Tier),
		UpdatedAt:      ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateScoreResponse{
		AccountAddress: msg.AccountAddress,
		PreviousScore:  previousScore,
		NewScore:       msg.NewScore,
		PreviousTier:   previousTier,
		NewTier:        updatedRecord.Tier,
		UpdatedAt:      ctx.BlockTime().Unix(),
	}, nil
}

// MsgServerWithContext wraps msgServer for gRPC context handling
type MsgServerWithContext struct {
	msgServer
}

// NewMsgServerWithContext returns a wrapped message server with context support
func NewMsgServerWithContext(k Keeper) MsgServerWithContext {
	return MsgServerWithContext{
		msgServer: msgServer{keeper: k},
	}
}
