package keeper

import (
	"crypto/sha256"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// Type alias for secp256k1 public key to avoid import collision
type secp256k1PubKey = secp256k1.PubKey

// ============================================================================
// Verification Status Management
// ============================================================================

// UpdateVerificationStatus updates the verification status of a scope
func (k Keeper) UpdateVerificationStatus(
	ctx sdk.Context,
	address sdk.AccAddress,
	scopeID string,
	status types.VerificationStatus,
	reason string,
	validatorAddr string,
) error {
	// Get the scope
	scope, found := k.GetScope(ctx, address, scopeID)
	if !found {
		return types.ErrScopeNotFound.Wrapf("scope %s not found", scopeID)
	}

	// Check if scope is revoked
	if scope.Revoked {
		return types.ErrScopeRevoked.Wrapf("scope %s is revoked", scopeID)
	}

	// Validate status transition
	if !scope.Status.CanTransitionTo(status) {
		return types.ErrInvalidStatusTransition.Wrapf(
			"cannot transition from %s to %s", scope.Status, status)
	}

	// Update scope status
	previousStatus := scope.Status
	scope.Status = status

	// Update verified time if transitioning to verified
	if status == types.VerificationStatusVerified {
		now := ctx.BlockTime()
		scope.VerifiedAt = &now
	}

	// Store updated scope
	if err := k.setScope(ctx, address, &scope); err != nil {
		return err
	}

	// Update identity record
	record, found := k.GetIdentityRecord(ctx, address)
	if found {
		for i, ref := range record.ScopeRefs {
			if ref.ScopeID == scopeID {
				record.ScopeRefs[i].Status = status
				break
			}
		}

		// Update last verified time if status is verified
		if status == types.VerificationStatusVerified {
			now := ctx.BlockTime()
			record.LastVerifiedAt = &now
		}

		record.UpdatedAt = ctx.BlockTime()
		if err := k.SetIdentityRecord(ctx, record); err != nil {
			return err
		}
	}

	// Store verification event
	event := types.NewVerificationEvent(
		generateEventID(ctx, address, scopeID),
		scopeID,
		previousStatus,
		status,
		ctx.BlockTime(),
		reason,
	)
	event.ValidatorAddress = validatorAddr
	k.storeVerificationEvent(ctx, address, event)

	return nil
}

// UpdateScore updates the identity score for an account
func (k Keeper) UpdateScore(
	ctx sdk.Context,
	address sdk.AccAddress,
	score uint32,
	scoreVersion string,
) error {
	if score > 100 {
		return types.ErrInvalidScore.Wrap("score cannot exceed 100")
	}

	record, found := k.GetIdentityRecord(ctx, address)
	if !found {
		return types.ErrIdentityRecordNotFound.Wrapf("identity record not found for %s", address.String())
	}

	if record.Locked {
		return types.ErrIdentityLocked.Wrap(record.LockedReason)
	}

	// Update score and tier
	previousScore := record.CurrentScore
	previousTier := record.Tier

	record.CurrentScore = score
	record.ScoreVersion = scoreVersion
	record.UpdateTier()
	record.UpdatedAt = ctx.BlockTime()

	if err := k.SetIdentityRecord(ctx, record); err != nil {
		return err
	}

	// Store verification event for score update
	event := types.NewVerificationEvent(
		generateEventID(ctx, address, "score"),
		"",
		types.VerificationStatusUnknown,
		types.VerificationStatusUnknown,
		ctx.BlockTime(),
		"score updated",
	)
	event.Score = &score
	event.Metadata = map[string]string{
		"previous_score": uintToString(previousScore),
		"new_score":      uintToString(score),
		"previous_tier":  string(previousTier),
		"new_tier":       string(record.Tier),
		"score_version":  scoreVersion,
	}
	k.storeVerificationEvent(ctx, address, event)

	return nil
}

// storeVerificationEvent stores a verification event
func (k Keeper) storeVerificationEvent(ctx sdk.Context, address sdk.AccAddress, event *types.VerificationEvent) {
	// Events are stored for audit purposes
	// For now, we emit as SDK events; persistent storage can be added if needed
	_ = ctx.EventManager().EmitTypedEvent(&types.EventStatusUpdated{
		AccountAddress:   address.String(),
		ScopeID:          event.ScopeID,
		PreviousStatus:   string(event.PreviousStatus),
		NewStatus:        string(event.NewStatus),
		Reason:           event.Reason,
		ValidatorAddress: event.ValidatorAddress,
		UpdatedAt:        event.Timestamp.Unix(),
	})
}

// GetVerificationHistory returns the verification history for an account
func (k Keeper) GetVerificationHistory(ctx sdk.Context, address sdk.AccAddress, limit uint32) []types.VerificationEvent {
	// For now, verification history is event-based
	// A persistent implementation would iterate over stored events
	return []types.VerificationEvent{}
}

// ============================================================================
// Signature Validation
// ============================================================================

// ValidateUserSignatureWithPubKey validates the user signature with the provided public key.
// This performs full cryptographic verification.
//
// Task Reference: VE-3022 - Cryptographic Signature Verification
func (k Keeper) ValidateUserSignatureWithPubKey(
	ctx sdk.Context,
	address sdk.AccAddress,
	pubKeyBytes []byte,
	signature []byte,
	payload []byte,
) error {
	params := k.GetParams(ctx)

	// Check if user signature is required
	if !params.RequireUserSignature {
		return nil
	}

	// Create secp256k1 public key from bytes
	if len(pubKeyBytes) != Secp256k1PublicKeySize {
		return types.ErrInvalidPublicKeyLength.Wrapf(
			"expected %d bytes, got %d bytes",
			Secp256k1PublicKeySize, len(pubKeyBytes),
		)
	}

	pubKey := &secp256k1PubKey{Key: pubKeyBytes}

	// Verify the public key derives to the expected address
	if err := VerifyAddressMatchesPubKey(pubKey, address); err != nil {
		return types.ErrInvalidUserSignature.Wrapf("address mismatch: %v", err)
	}

	// Verify the signature
	if err := VerifySecp256k1Signature(pubKey, payload, signature); err != nil {
		return types.ErrInvalidUserSignature.Wrapf("signature verification failed: %v", err)
	}

	return nil
}

// ValidateUploadSignatures validates both client and user signatures for an upload.
// This uses real cryptographic verification for client signatures.
//
// Task Reference: VE-3022 - Cryptographic Signature Verification
func (k Keeper) ValidateUploadSignatures(
	ctx sdk.Context,
	address sdk.AccAddress,
	metadata *types.UploadMetadata,
) error {
	// Validate client signature with real cryptographic verification
	clientPayload := metadata.SigningPayload()
	if err := k.ValidateClientSignature(ctx, metadata.ClientID, metadata.ClientSignature, clientPayload); err != nil {
		return err
	}

	// Validate user signature (basic validation, full verification requires pubkey)
	userPayload := metadata.UserSigningPayload()
	if err := k.ValidateUserSignature(ctx, address, metadata.UserSignature, userPayload); err != nil {
		return err
	}

	return nil
}

// ValidateUploadSignaturesWithPubKey validates both client and user signatures for an upload
// with full cryptographic verification using the provided user public key.
//
// Task Reference: VE-3022 - Cryptographic Signature Verification
func (k Keeper) ValidateUploadSignaturesWithPubKey(
	ctx sdk.Context,
	address sdk.AccAddress,
	userPubKey []byte,
	metadata *types.UploadMetadata,
) error {
	// Validate client signature with real cryptographic verification
	clientPayload := metadata.SigningPayload()
	if err := k.ValidateClientSignature(ctx, metadata.ClientID, metadata.ClientSignature, clientPayload); err != nil {
		return err
	}

	// Validate user signature with full cryptographic verification
	userPayload := metadata.UserSigningPayload()
	if err := k.ValidateUserSignatureWithPubKey(ctx, address, userPubKey, metadata.UserSignature, userPayload); err != nil {
		return err
	}

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// generateEventID generates a unique event ID
func generateEventID(ctx sdk.Context, address sdk.AccAddress, scopeID string) string {
	h := sha256.New()
	h.Write(address.Bytes())
	h.Write([]byte(scopeID))
	h.Write([]byte{byte(ctx.BlockHeight() >> 56), byte(ctx.BlockHeight() >> 48),
		byte(ctx.BlockHeight() >> 40), byte(ctx.BlockHeight() >> 32),
		byte(ctx.BlockHeight() >> 24), byte(ctx.BlockHeight() >> 16),
		byte(ctx.BlockHeight() >> 8), byte(ctx.BlockHeight())})
	sum := h.Sum(nil)
	return bytesToHex(sum[:16])
}

// bytesToHex converts bytes to hex string
func bytesToHex(b []byte) string {
	const hexChars = "0123456789abcdef"
	result := make([]byte, len(b)*2)
	for i, v := range b {
		result[i*2] = hexChars[v>>4]
		result[i*2+1] = hexChars[v&0x0f]
	}
	return string(result)
}

// uintToString converts uint32 to string
func uintToString(n uint32) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
