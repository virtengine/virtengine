package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Borderline Fallback Completion Handler
// ============================================================================

// HandleBorderlineFallbackCompleted processes the completion of a borderline fallback
// after the MFA challenge has been satisfied.
func (k Keeper) HandleBorderlineFallbackCompleted(
	ctx sdk.Context,
	accountAddr string,
	challengeID string,
	factorsSatisfied []string,
) error {
	// Find the fallback record by challenge ID
	fallbackRecord, found := k.GetBorderlineFallbackByChallenge(ctx, challengeID)
	if !found {
		return types.ErrBorderlineFallbackNotFound.Wrapf("no fallback found for challenge %s", challengeID)
	}

	// Verify the account address matches
	if fallbackRecord.AccountAddress != accountAddr {
		return types.ErrUnauthorized.Wrap("account address mismatch")
	}

	// Check if fallback is still pending
	if !fallbackRecord.IsPending() {
		return types.ErrBorderlineFallbackAlreadyCompleted.Wrapf("fallback status is %s", fallbackRecord.Status)
	}

	// Check if fallback has expired
	now := ctx.BlockTime().Unix()
	if fallbackRecord.IsExpired(now) {
		// Mark as expired and emit event
		fallbackRecord.MarkExpired(now)
		if err := k.setBorderlineFallbackRecord(ctx, fallbackRecord); err != nil {
			return err
		}
		k.removeFromPendingFallbackQueue(ctx, fallbackRecord)
		k.emitBorderlineFallbackExpiredEvent(ctx, fallbackRecord)

		// Emit spec-defined authorization expired event (per veid-flow-spec.md)
		_ = k.EmitAuthorizationExpiredEvent(
			ctx,
			fallbackRecord.AccountAddress,
			fallbackRecord.FallbackID,
			"verification_fallback",
			fallbackRecord.CreatedAt,
		)
		return types.ErrBorderlineFallbackExpired
	}

	// Verify MFA challenge was actually satisfied
	if err := k.verifyMFAChallengeCompleted(ctx, challengeID); err != nil {
		k.Logger(ctx).Warn("MFA challenge not satisfied",
			"account", accountAddr,
			"challenge_id", challengeID,
			"error", err,
		)

		// Mark fallback as failed
		fallbackRecord.MarkFailed(now)
		if err := k.setBorderlineFallbackRecord(ctx, fallbackRecord); err != nil {
			return err
		}
		k.removeFromPendingFallbackQueue(ctx, fallbackRecord)
		k.emitBorderlineFallbackFailedEvent(ctx, fallbackRecord, "MFA challenge not satisfied")

		return types.ErrMFAChallengeNotSatisfied
	}

	// Check minimum factors satisfied requirement
	params := k.GetBorderlineParams(ctx)
	if uint32(len(factorsSatisfied)) < params.MinFactorsSatisfied {
		return types.ErrMFAChallengeNotSatisfied.Wrapf(
			"need at least %d factors, got %d",
			params.MinFactorsSatisfied,
			len(factorsSatisfied),
		)
	}

	// Update fallback record to completed
	fallbackRecord.MarkCompleted(factorsSatisfied, types.VerificationStatusVerified, now)
	if err := k.setBorderlineFallbackRecord(ctx, fallbackRecord); err != nil {
		return err
	}

	// Remove from pending queue
	k.removeFromPendingFallbackQueue(ctx, fallbackRecord)

	// Update the account's verification status to Verified
	address, err := sdk.AccAddressFromBech32(accountAddr)
	if err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}

	// Update score with verified status
	err = k.SetScoreWithDetails(ctx, accountAddr, fallbackRecord.BorderlineScore, ScoreDetails{
		Status:       types.AccountStatusVerified,
		ModelVersion: "borderline-fallback",
		Reason:       fmt.Sprintf("borderline fallback completed via %s", strings.Join(factorsSatisfied, ",")),
	})
	if err != nil {
		return err
	}

	// Determine factor class for audit
	factorClass := k.determineFactorClass(factorsSatisfied)

	// Emit completion event
	k.emitBorderlineFallbackCompletedEvent(ctx, fallbackRecord, factorsSatisfied, factorClass)

	// Emit spec-defined authorization granted event (per veid-flow-spec.md)
	// This is emitted when MFA successfully grants authorization for verification
	_ = k.EmitAuthorizationGrantedEvent(
		ctx,
		accountAddr,
		fallbackRecord.FallbackID,
		"verification_fallback",
		factorsSatisfied,
		0, // No expiry for completed fallback
	)

	k.Logger(ctx).Info("borderline fallback completed successfully",
		"account", accountAddr,
		"fallback_id", fallbackRecord.FallbackID,
		"challenge_id", challengeID,
		"factors_satisfied", strings.Join(factorsSatisfied, ","),
		"factor_class", factorClass,
		"borderline_score", fallbackRecord.BorderlineScore,
	)

	// Update identity record tier if it exists (after successful MFA fallback)
	if record, found := k.GetIdentityRecord(ctx, address); found {
		record.Tier = types.IdentityTierVerified
		if err := k.SetIdentityRecord(ctx, record); err != nil {
			k.Logger(ctx).Error("failed to update identity record tier", "error", err)
		}
	}

	return nil
}

// verifyMFAChallengeCompleted checks if the MFA challenge was successfully completed
func (k Keeper) verifyMFAChallengeCompleted(ctx sdk.Context, challengeID string) error {
	if k.mfaKeeper == nil {
		return types.ErrMFAChallengeNotSatisfied.Wrap("MFA keeper not configured")
	}

	challenge, found := k.mfaKeeper.GetChallenge(ctx, challengeID)
	if !found {
		return types.ErrMFAChallengeNotSatisfied.Wrap("challenge not found")
	}

	if challenge.Status != mfatypes.ChallengeStatusVerified {
		return types.ErrMFAChallengeNotSatisfied.Wrapf("challenge status is %s, expected verified", challenge.Status.String())
	}

	return nil
}

// DetermineFactorClass determines the security class of the satisfied factors
// Returns "high", "medium", or "low" based on the security level of the factors
func (k Keeper) DetermineFactorClass(factorsSatisfied []string) string {
	hasHigh := false
	hasMedium := false

	for _, factorName := range factorsSatisfied {
		factorType, err := mfatypes.FactorTypeFromString(factorName)
		if err != nil {
			continue
		}

		level := factorType.GetSecurityLevel()
		switch level {
		case mfatypes.FactorSecurityLevelHigh:
			hasHigh = true
		case mfatypes.FactorSecurityLevelMedium:
			hasMedium = true
		}
	}

	if hasHigh {
		return "high"
	}
	if hasMedium {
		return "medium"
	}
	return "low"
}

// determineFactorClass is the private version for internal use
func (k Keeper) determineFactorClass(factorsSatisfied []string) string {
	return k.DetermineFactorClass(factorsSatisfied)
}

// ============================================================================
// Fallback Cancellation
// ============================================================================

// CancelBorderlineFallback cancels a pending borderline fallback
func (k Keeper) CancelBorderlineFallback(
	ctx sdk.Context,
	accountAddr string,
	fallbackID string,
) error {
	fallbackRecord, found := k.GetBorderlineFallbackRecord(ctx, fallbackID)
	if !found {
		return types.ErrBorderlineFallbackNotFound
	}

	if fallbackRecord.AccountAddress != accountAddr {
		return types.ErrUnauthorized.Wrap("account address mismatch")
	}

	if !fallbackRecord.IsPending() {
		return types.ErrBorderlineFallbackAlreadyCompleted.Wrapf("fallback status is %s", fallbackRecord.Status)
	}

	now := ctx.BlockTime().Unix()
	fallbackRecord.MarkCancelled(now)

	if err := k.setBorderlineFallbackRecord(ctx, fallbackRecord); err != nil {
		return err
	}

	k.removeFromPendingFallbackQueue(ctx, fallbackRecord)

	k.Logger(ctx).Info("borderline fallback cancelled",
		"account", accountAddr,
		"fallback_id", fallbackID,
	)

	return nil
}

// ============================================================================
// Expired Fallback Processing
// ============================================================================

// ProcessExpiredFallbacks processes and marks expired fallbacks
// This should be called in EndBlock or via a governance mechanism
func (k Keeper) ProcessExpiredFallbacks(ctx sdk.Context) int {
	now := ctx.BlockTime().Unix()
	store := ctx.KVStore(k.skey)
	iterator := store.Iterator(types.PendingBorderlineFallbackPrefixKey(), nil)
	defer iterator.Close()

	expiredCount := 0

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		// Extract expiry time from key
		// Key format: prefix | expires_at (8 bytes) | "/" | fallback_id
		if len(key) < len(types.PrefixPendingBorderlineFallback)+9 {
			continue
		}

		expiryBytes := key[len(types.PrefixPendingBorderlineFallback) : len(types.PrefixPendingBorderlineFallback)+8]
		expiresAt := decodeInt64(expiryBytes)

		// If not yet expired, we can stop (ordered by expiry time)
		if expiresAt > now {
			break
		}

		// Extract fallback ID
		remainder := key[len(types.PrefixPendingBorderlineFallback)+9:]
		fallbackID := string(remainder)

		// Get and expire the fallback
		fallbackRecord, found := k.GetBorderlineFallbackRecord(ctx, fallbackID)
		if !found {
			// Clean up orphaned queue entry
			store.Delete(key)
			continue
		}

		if fallbackRecord.IsPending() {
			fallbackRecord.MarkExpired(now)
			_ = k.setBorderlineFallbackRecord(ctx, fallbackRecord)
			k.emitBorderlineFallbackExpiredEvent(ctx, fallbackRecord)

			// Emit spec-defined authorization expired event (per veid-flow-spec.md)
			_ = k.EmitAuthorizationExpiredEvent(
				ctx,
				fallbackRecord.AccountAddress,
				fallbackRecord.FallbackID,
				"verification_fallback",
				fallbackRecord.CreatedAt,
			)
			expiredCount++
		}

		// Remove from queue
		store.Delete(key)
	}

	if expiredCount > 0 {
		k.Logger(ctx).Info("processed expired borderline fallbacks", "count", expiredCount)
	}

	return expiredCount
}

// decodeInt64 decodes an int64 from big-endian bytes
func decodeInt64(b []byte) int64 {
	if len(b) != 8 {
		return 0
	}
	return int64(b[0])<<56 | int64(b[1])<<48 | int64(b[2])<<40 | int64(b[3])<<32 |
		int64(b[4])<<24 | int64(b[5])<<16 | int64(b[6])<<8 | int64(b[7])
}

// ============================================================================
// Event Emission
// ============================================================================

// emitBorderlineFallbackCompletedEvent emits an event when borderline fallback completes
func (k Keeper) emitBorderlineFallbackCompletedEvent(
	ctx sdk.Context,
	record *types.BorderlineFallbackRecord,
	factorsSatisfied []string,
	factorClass string,
) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBorderlineFallbackCompleted,
			sdk.NewAttribute(types.AttributeKeyAccountAddress, record.AccountAddress),
			sdk.NewAttribute(types.AttributeKeyFallbackID, record.FallbackID),
			sdk.NewAttribute(types.AttributeKeyChallengeID, record.ChallengeID),
			sdk.NewAttribute(types.AttributeKeySatisfiedFactors, strings.Join(factorsSatisfied, ",")),
			sdk.NewAttribute(types.AttributeKeyFactorClass, factorClass),
			sdk.NewAttribute(types.AttributeKeyFinalStatus, string(record.FinalVerificationStatus)),
			sdk.NewAttribute(types.AttributeKeyBorderlineScore, fmt.Sprintf("%d", record.BorderlineScore)),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", record.CompletedAt)),
		),
	)
}

// emitBorderlineFallbackFailedEvent emits an event when borderline fallback fails
func (k Keeper) emitBorderlineFallbackFailedEvent(
	ctx sdk.Context,
	record *types.BorderlineFallbackRecord,
	reason string,
) {
	// Get attempt count from MFA challenge if available
	attemptCount := uint32(0)
	if k.mfaKeeper != nil {
		if challenge, found := k.mfaKeeper.GetChallenge(ctx, record.ChallengeID); found {
			attemptCount = challenge.AttemptCount
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBorderlineFallbackFailed,
			sdk.NewAttribute(types.AttributeKeyAccountAddress, record.AccountAddress),
			sdk.NewAttribute(types.AttributeKeyFallbackID, record.FallbackID),
			sdk.NewAttribute(types.AttributeKeyChallengeID, record.ChallengeID),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
			sdk.NewAttribute("attempt_count", fmt.Sprintf("%d", attemptCount)),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", record.CompletedAt)),
		),
	)
}

// emitBorderlineFallbackExpiredEvent emits an event when borderline fallback expires
func (k Keeper) emitBorderlineFallbackExpiredEvent(
	ctx sdk.Context,
	record *types.BorderlineFallbackRecord,
) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBorderlineFallbackExpired,
			sdk.NewAttribute(types.AttributeKeyAccountAddress, record.AccountAddress),
			sdk.NewAttribute(types.AttributeKeyFallbackID, record.FallbackID),
			sdk.NewAttribute(types.AttributeKeyChallengeID, record.ChallengeID),
			sdk.NewAttribute(types.AttributeKeyBorderlineScore, fmt.Sprintf("%d", record.BorderlineScore)),
			sdk.NewAttribute("created_at", fmt.Sprintf("%d", record.CreatedAt)),
			sdk.NewAttribute("expired_at", fmt.Sprintf("%d", record.ExpiresAt)),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)
}
