package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/veid/types"
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
func (ms msgServer) UploadScope(goCtx context.Context, msg *types.MsgUploadScope) (*types.MsgUploadScopeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	params := ms.keeper.GetParams(ctx)

	// Validate client is approved
	if params.RequireClientSignature && !ms.keeper.IsClientApproved(ctx, msg.ClientId) {
		return nil, types.ErrClientNotApproved.Wrapf("client %s is not approved", msg.ClientId)
	}

	// Create upload metadata
	metadata := types.NewUploadMetadata(
		msg.Salt,
		msg.DeviceFingerprint,
		msg.ClientId,
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

	// Convert proto envelope to local type
	localPayload := encryptedPayloadFromProto(&msg.EncryptedPayload)

	// Create the scope
	scope := types.NewIdentityScope(
		msg.ScopeId,
		types.ScopeTypeFromProto(msg.ScopeType),
		localPayload,
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
		ScopeID:           msg.ScopeId,
		ScopeType:         string(types.ScopeTypeFromProto(msg.ScopeType)),
		Version:           types.ScopeSchemaVersion,
		ClientID:          msg.ClientId,
		PayloadHash:       metadata.PayloadHashHex(),
		DeviceFingerprint: msg.DeviceFingerprint,
		UploadedAt:        ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgUploadScopeResponse{
		ScopeId:    msg.ScopeId,
		Status:     types.VerificationStatusPBPending,
		UploadedAt: ctx.BlockTime().Unix(),
	}, nil
}

// RevokeScope handles revoking an identity scope
func (ms msgServer) RevokeScope(goCtx context.Context, msg *types.MsgRevokeScope) (*types.MsgRevokeScopeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Revoke the scope
	if err := ms.keeper.RevokeScope(ctx, sender, msg.ScopeId, msg.Reason); err != nil {
		return nil, err
	}

	// Get scope for event data
	scope, _ := ms.keeper.GetScope(ctx, sender, msg.ScopeId)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventScopeRevoked{
		AccountAddress: sender.String(),
		ScopeID:        msg.ScopeId,
		ScopeType:      string(scope.ScopeType),
		Reason:         msg.Reason,
		RevokedAt:      ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgRevokeScopeResponse{
		ScopeId:   msg.ScopeId,
		RevokedAt: ctx.BlockTime().Unix(),
	}, nil
}

// RequestVerification handles requesting verification for a scope
func (ms msgServer) RequestVerification(goCtx context.Context, msg *types.MsgRequestVerification) (*types.MsgRequestVerificationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Get the scope
	scope, found := ms.keeper.GetScope(ctx, sender, msg.ScopeId)
	if !found {
		return nil, types.ErrScopeNotFound.Wrapf("scope %s not found", msg.ScopeId)
	}

	// Check if scope can be verified
	if !scope.CanBeVerified() {
		if scope.Revoked {
			return nil, types.ErrScopeRevoked.Wrapf("scope %s is revoked", msg.ScopeId)
		}
		if scope.Status == types.VerificationStatusInProgress {
			return nil, types.ErrVerificationInProgress.Wrapf("verification already in progress for scope %s", msg.ScopeId)
		}
		return nil, types.ErrInvalidStatusTransition.Wrapf("scope %s cannot be verified in status %s", msg.ScopeId, scope.Status)
	}

	// Update status to in progress
	err = ms.keeper.UpdateVerificationStatus(
		ctx,
		sender,
		msg.ScopeId,
		types.VerificationStatusInProgress,
		"verification requested",
		"",
	)
	if err != nil {
		return nil, err
	}

	// Emit legacy event for backwards compatibility
	err = ctx.EventManager().EmitTypedEvent(&types.EventVerificationRequested{
		AccountAddress: sender.String(),
		ScopeID:        msg.ScopeId,
		ScopeType:      string(scope.ScopeType),
		RequestedAt:    ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	// Emit spec-defined verification submitted event
	err = ctx.EventManager().EmitTypedEvent(&types.EventVerificationSubmitted{
		Account:     sender.String(),
		ScopeID:     msg.ScopeId,
		ScopeType:   string(scope.ScopeType),
		RequestID:   msg.ScopeId, // Using scope ID as request ID for now
		BlockHeight: ctx.BlockHeight(),
		Timestamp:   ctx.BlockTime().Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgRequestVerificationResponse{
		ScopeId:     msg.ScopeId,
		Status:      types.VerificationStatusPBInProgress,
		RequestedAt: ctx.BlockTime().Unix(),
	}, nil
}

// UpdateVerificationStatus handles validators updating verification status
func (ms msgServer) UpdateVerificationStatus(goCtx context.Context, msg *types.MsgUpdateVerificationStatus) (*types.MsgUpdateVerificationStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	accountAddr, err := sdk.AccAddressFromBech32(msg.AccountAddress)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidAccountAddr)
	}

	// Validate that sender is a bonded validator
	// Only validators can submit verification status updates
	if !ms.keeper.IsValidator(ctx, sender) {
		return nil, types.ErrUnauthorized.Wrap("only validators can submit verification updates")
	}

	// Get current scope for previous status
	scope, found := ms.keeper.GetScope(ctx, accountAddr, msg.ScopeId)
	if !found {
		return nil, types.ErrScopeNotFound.Wrapf("scope %s not found", msg.ScopeId)
	}
	previousStatus := scope.Status

	// Convert proto status to local type
	localNewStatus := types.VerificationStatusFromProto(msg.NewStatus)

	// Update verification status
	err = ms.keeper.UpdateVerificationStatus(
		ctx,
		accountAddr,
		msg.ScopeId,
		localNewStatus,
		msg.Reason,
		sender.String(),
	)
	if err != nil {
		return nil, err
	}

	// Emit appropriate event
	if msg.NewStatus == types.VerificationStatusPBVerified {
		err = ctx.EventManager().EmitTypedEvent(&types.EventScopeVerified{
			AccountAddress:   msg.AccountAddress,
			ScopeID:          msg.ScopeId,
			ScopeType:        string(scope.ScopeType),
			ValidatorAddress: sender.String(),
			VerifiedAt:       ctx.BlockTime().Unix(),
		})
	} else if msg.NewStatus == types.VerificationStatusPBRejected {
		err = ctx.EventManager().EmitTypedEvent(&types.EventScopeRejected{
			AccountAddress:   msg.AccountAddress,
			ScopeID:          msg.ScopeId,
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
		ScopeId:        msg.ScopeId,
		PreviousStatus: types.VerificationStatusToProto(previousStatus),
		NewStatus:      msg.NewStatus,
		UpdatedAt:      ctx.BlockTime().Unix(),
	}, nil
}

// UpdateScore handles validators updating identity score
func (ms msgServer) UpdateScore(goCtx context.Context, msg *types.MsgUpdateScore) (*types.MsgUpdateScoreResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	accountAddr, err := sdk.AccAddressFromBech32(msg.AccountAddress)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidAccountAddr)
	}

	// Validate that sender is a bonded validator
	// Only validators can submit ML score updates
	if !ms.keeper.IsValidator(ctx, sender) {
		return nil, types.ErrUnauthorized.Wrap("only validators can submit ML score updates")
	}

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

	// Emit legacy score updated event for backwards compatibility
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

	// Emit spec-defined tier changed event if tier actually changed
	if previousTier != updatedRecord.Tier {
		err = ctx.EventManager().EmitTypedEvent(&types.EventTierChanged{
			Account:     msg.AccountAddress,
			OldTier:     string(previousTier),
			NewTier:     string(updatedRecord.Tier),
			Score:       msg.NewScore,
			BlockHeight: ctx.BlockHeight(),
			Timestamp:   ctx.BlockTime().Unix(),
		})
		if err != nil {
			return nil, err
		}
	}

	return &types.MsgUpdateScoreResponse{
		AccountAddress: msg.AccountAddress,
		PreviousScore:  previousScore,
		NewScore:       msg.NewScore,
		PreviousTier:   types.IdentityTierToProto(previousTier),
		NewTier:        types.IdentityTierToProto(updatedRecord.Tier),
		UpdatedAt:      ctx.BlockTime().Unix(),
	}, nil
}

// CreateIdentityWallet handles creating a new identity wallet
func (ms msgServer) CreateIdentityWallet(goCtx context.Context, msg *types.MsgCreateIdentityWallet) (*types.MsgCreateIdentityWalletResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	wallet, err := ms.keeper.CreateWallet(ctx, sender, msg.BindingSignature, msg.BindingPubKey)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateIdentityWalletResponse{
		WalletId:  wallet.WalletID,
		CreatedAt: ctx.BlockTime().Unix(),
	}, nil
}

// AddScopeToWallet handles adding a scope reference to a wallet
func (ms msgServer) AddScopeToWallet(goCtx context.Context, msg *types.MsgAddScopeToWallet) (*types.MsgAddScopeToWalletResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Create scope reference from message
	scopeRef := types.ScopeReference{
		ScopeID:      msg.ScopeId,
		ScopeType:    types.ScopeTypeFromProto(msg.ScopeType),
		EnvelopeHash: msg.EnvelopeHash,
		AddedAt:      ctx.BlockTime(),
	}

	if err := ms.keeper.AddScopeToWallet(ctx, sender, scopeRef, msg.UserSignature); err != nil {
		return nil, err
	}

	return &types.MsgAddScopeToWalletResponse{
		ScopeId: msg.ScopeId,
		AddedAt: ctx.BlockTime().Unix(),
	}, nil
}

// RevokeScopeFromWallet handles revoking a scope from a wallet
func (ms msgServer) RevokeScopeFromWallet(goCtx context.Context, msg *types.MsgRevokeScopeFromWallet) (*types.MsgRevokeScopeFromWalletResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	if err := ms.keeper.RevokeScopeFromWallet(ctx, sender, msg.ScopeId, msg.Reason, msg.UserSignature); err != nil {
		return nil, err
	}

	return &types.MsgRevokeScopeFromWalletResponse{
		ScopeId:   msg.ScopeId,
		RevokedAt: ctx.BlockTime().Unix(),
	}, nil
}

// UpdateConsentSettings handles updating consent settings
func (ms msgServer) UpdateConsentSettings(goCtx context.Context, msg *types.MsgUpdateConsentSettings) (*types.MsgUpdateConsentSettingsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Convert Unix timestamp to *time.Time if provided
	var expiresAt *time.Time
	if msg.ExpiresAt > 0 {
		t := time.Unix(msg.ExpiresAt, 0)
		expiresAt = &t
	}

	// Convert proto GlobalSettings to local type
	var globalSettings *types.GlobalConsentUpdate
	if msg.GlobalSettings != nil {
		globalSettings = &types.GlobalConsentUpdate{
			ShareWithProviders:         boolPtrFromBool(msg.GlobalSettings.ShareWithProviders),
			ShareForVerification:       boolPtrFromBool(msg.GlobalSettings.ShareForVerification),
			AllowReVerification:        boolPtrFromBool(msg.GlobalSettings.AllowReVerification),
			AllowDerivedFeatureSharing: boolPtrFromBool(msg.GlobalSettings.AllowDerivedFeatureSharing),
		}
	}

	update := types.ConsentUpdateRequest{
		ScopeID:        msg.ScopeId,
		GrantConsent:   msg.GrantConsent,
		Purpose:        msg.Purpose,
		ExpiresAt:      expiresAt,
		GlobalSettings: globalSettings,
	}

	if err := ms.keeper.UpdateConsent(ctx, sender, update, msg.UserSignature); err != nil {
		return nil, err
	}

	wallet, found := ms.keeper.GetWallet(ctx, sender)
	if !found {
		return nil, types.ErrWalletNotFound.Wrap("wallet not found after consent update")
	}

	return &types.MsgUpdateConsentSettingsResponse{
		UpdatedAt:      ctx.BlockTime().Unix(),
		ConsentVersion: wallet.ConsentSettings.ConsentVersion,
	}, nil
}

// boolPtrFromBool converts a bool to *bool for optional settings
func boolPtrFromBool(b bool) *bool {
	return &b
}

// RebindWallet handles rebinding a wallet to a new address
func (ms msgServer) RebindWallet(goCtx context.Context, msg *types.MsgRebindWallet) (*types.MsgRebindWalletResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	if err := ms.keeper.RebindWallet(ctx, sender, msg.NewBindingPubKey, msg.NewBindingSignature, msg.OldSignature); err != nil {
		return nil, err
	}

	return &types.MsgRebindWalletResponse{
		ReboundAt: ctx.BlockTime().Unix(),
	}, nil
}

// UpdateDerivedFeatures handles updating derived features
func (ms msgServer) UpdateDerivedFeatures(goCtx context.Context, msg *types.MsgUpdateDerivedFeatures) (*types.MsgUpdateDerivedFeaturesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	accountAddr, err := sdk.AccAddressFromBech32(msg.AccountAddress)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid account address")
	}

	update := &types.DerivedFeaturesUpdate{
		AccountAddress:    msg.AccountAddress,
		FaceEmbeddingHash: msg.FaceEmbeddingHash,
		DocFieldHashes:    msg.DocFieldHashes,
		BiometricHash:     msg.BiometricHash,
		LivenessProofHash: msg.LivenessProofHash,
		ModelVersion:      msg.ModelVersion,
		ValidatorAddress:  sender.String(),
	}
	if err := ms.keeper.UpdateDerivedFeatures(ctx, accountAddr, update); err != nil {
		return nil, err
	}

	return &types.MsgUpdateDerivedFeaturesResponse{
		UpdatedAt: ctx.BlockTime().Unix(),
	}, nil
}

// CompleteBorderlineFallback handles completing a borderline fallback verification
func (ms msgServer) CompleteBorderlineFallback(goCtx context.Context, msg *types.MsgCompleteBorderlineFallback) (*types.MsgCompleteBorderlineFallbackResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	// Find the fallback record by challenge ID
	fallbackRecord, found := ms.keeper.GetBorderlineFallbackByChallenge(ctx, msg.ChallengeId)
	if !found {
		return nil, types.ErrBorderlineFallbackNotFound.Wrapf("no fallback found for challenge %s", msg.ChallengeId)
	}

	// Verify the sender matches the account in the fallback record
	if fallbackRecord.AccountAddress != msg.Sender {
		return nil, types.ErrUnauthorized.Wrap("sender does not match fallback account")
	}

	// Handle the borderline fallback completion
	err = ms.keeper.HandleBorderlineFallbackCompleted(
		ctx,
		msg.Sender,
		msg.ChallengeId,
		msg.FactorsSatisfied,
	)
	if err != nil {
		return nil, err
	}

	// Retrieve the updated fallback record for response
	updatedFallback, _ := ms.keeper.GetBorderlineFallbackRecord(ctx, fallbackRecord.FallbackID)

	// Determine factor class from satisfied factors
	factorClass := ms.keeper.DetermineFactorClass(msg.FactorsSatisfied)

	// Emit completion event (using SDK event since typed events not available)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBorderlineFallbackCompleted,
			sdk.NewAttribute(types.AttributeKeyAccountAddress, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyFallbackID, fallbackRecord.FallbackID),
			sdk.NewAttribute(types.AttributeKeyChallengeID, msg.ChallengeId),
			sdk.NewAttribute(types.AttributeKeyFactorClass, factorClass),
			sdk.NewAttribute(types.AttributeKeyFinalStatus, string(types.VerificationStatusVerified)),
		),
	)

	return &types.MsgCompleteBorderlineFallbackResponse{
		FallbackId:  updatedFallback.FallbackID,
		FinalStatus: types.VerificationStatusToProto(updatedFallback.FinalVerificationStatus),
		FactorClass: factorClass,
	}, nil
}

// UpdateBorderlineParams handles updating borderline parameters (governance)
func (ms msgServer) UpdateBorderlineParams(goCtx context.Context, msg *types.MsgUpdateBorderlineParams) (*types.MsgUpdateBorderlineParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Authority check - only governance can update borderline params
	if msg.Authority != ms.keeper.GetAuthority() {
		return nil, types.ErrUnauthorized.Wrapf("invalid authority; expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}

	if err := ms.keeper.SetBorderlineParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateBorderlineParamsResponse{}, nil
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

// UpdateParams updates the module parameters.
// Only the authority (typically governance) can execute this message.
func (ms msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check authority
	if msg.Authority != ms.keeper.authority {
		return nil, types.ErrUnauthorized.Wrapf("invalid authority: expected %s, got %s", ms.keeper.authority, msg.Authority)
	}

	// Convert proto params to local params
	localParams := types.Params{
		MaxScopesPerAccount:    msg.Params.MaxScopesPerAccount,
		MaxScopesPerType:       msg.Params.MaxScopesPerType,
		SaltMinBytes:           msg.Params.SaltMinBytes,
		SaltMaxBytes:           msg.Params.SaltMaxBytes,
		RequireClientSignature: msg.Params.RequireClientSignature,
		RequireUserSignature:   msg.Params.RequireUserSignature,
		VerificationExpiryDays: msg.Params.VerificationExpiryDays,
		// Keep existing MinScoreForTier from current params
		MinScoreForTier: ms.keeper.GetParams(ctx).MinScoreForTier,
	}

	// Set the new params
	if err := ms.keeper.SetParams(ctx, localParams); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// encryptedPayloadFromProto converts a proto EncryptedPayloadEnvelope to local type
func encryptedPayloadFromProto(protoPayload *types.EncryptedPayloadEnvelopePB) encryptiontypes.EncryptedPayloadEnvelope {
	if protoPayload == nil {
		return encryptiontypes.EncryptedPayloadEnvelope{}
	}

	return encryptiontypes.EncryptedPayloadEnvelope{
		Version:             protoPayload.Version,
		AlgorithmID:         protoPayload.AlgorithmId,
		AlgorithmVersion:    protoPayload.AlgorithmVersion,
		RecipientKeyIDs:     protoPayload.RecipientKeyIds,
		RecipientPublicKeys: protoPayload.RecipientPublicKeys,
		EncryptedKeys:       protoPayload.EncryptedKeys,
		Nonce:               protoPayload.Nonce,
		Ciphertext:          protoPayload.Ciphertext,
		SenderSignature:     protoPayload.SenderSignature,
		SenderPubKey:        protoPayload.SenderPubKey,
		Metadata:            protoPayload.Metadata,
	}
}
