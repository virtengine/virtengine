package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/enclave/types"
)

// BeginBlocker performs automatic maintenance tasks at the beginning of each block
func (k Keeper) BeginBlocker(ctx sdk.Context) {
	// Record metrics
	k.RecordMetrics(ctx)

	// Auto-complete expired key rotations
	k.processExpiredKeyRotations(ctx)

	// Update expired enclave identities
	k.updateExpiredIdentities(ctx)

	// Clean up old expired measurements (optional, configurable)
	k.cleanupExpiredMeasurements(ctx)

	// Clean up old rate limit tracking data (every 1000 blocks)
	if ctx.BlockHeight()%1000 == 0 {
		k.CleanupOldRegistrationCounts(ctx)
	}
}

// processExpiredKeyRotations automatically completes key rotations that have
// passed their overlap period end height
func (k Keeper) processExpiredKeyRotations(ctx sdk.Context) {
	currentHeight := ctx.BlockHeight()

	k.WithEnclaveIdentities(ctx, func(identity types.EnclaveIdentity) bool {
		if identity.Status != types.EnclaveIdentityStatusRotating {
			return false
		}

		validatorAddr, err := sdk.AccAddressFromBech32(identity.ValidatorAddress)
		if err != nil {
			k.Logger(ctx).Error(
				"invalid validator address during rotation check",
				"address", identity.ValidatorAddress,
				"error", err,
			)
			return false
		}

		rotation, exists := k.GetActiveKeyRotation(ctx, validatorAddr)
		if !exists {
			// Identity is in rotating status but no active rotation found
			// Reset status to active
			identity.Status = types.EnclaveIdentityStatusActive
			if err := k.UpdateEnclaveIdentity(ctx, &identity); err != nil {
				k.Logger(ctx).Error(
					"failed to reset identity status",
					"validator", identity.ValidatorAddress,
					"error", err,
				)
			}
			return false
		}

		// Check if overlap period has ended
		if currentHeight >= rotation.OverlapEndHeight {
			if err := k.CompleteKeyRotation(ctx, validatorAddr); err != nil {
				k.Logger(ctx).Error(
					"failed to auto-complete key rotation",
					"validator", identity.ValidatorAddress,
					"height", currentHeight,
					"overlap_end", rotation.OverlapEndHeight,
					"error", err,
				)
			} else {
				k.Logger(ctx).Info(
					"auto-completed key rotation",
					"validator", identity.ValidatorAddress,
					"height", currentHeight,
					"new_key", rotation.NewKeyFingerprint,
				)
			}
		}

		return false
	})
}

// updateExpiredIdentities automatically marks identities as expired when they
// reach their expiry height
func (k Keeper) updateExpiredIdentities(ctx sdk.Context) {
	currentHeight := ctx.BlockHeight()

	k.WithEnclaveIdentities(ctx, func(identity types.EnclaveIdentity) bool {
		// Skip if already expired or revoked
		if identity.Status == types.EnclaveIdentityStatusExpired ||
			identity.Status == types.EnclaveIdentityStatusRevoked {
			return false
		}

		// Check if identity has expired
		if types.IsIdentityExpired(&identity, currentHeight) {
			identity.Status = types.EnclaveIdentityStatusExpired
			if err := k.UpdateEnclaveIdentity(ctx, &identity); err != nil {
				k.Logger(ctx).Error(
					"failed to update expired identity",
					"validator", identity.ValidatorAddress,
					"expiry_height", identity.ExpiryHeight,
					"error", err,
				)
			} else {
				k.Logger(ctx).Info(
					"marked identity as expired",
					"validator", identity.ValidatorAddress,
					"expiry_height", identity.ExpiryHeight,
				)

				// Emit expiry event
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						types.EventTypeEnclaveIdentityExpired,
						sdk.NewAttribute(types.AttributeKeyValidator, identity.ValidatorAddress),
						sdk.NewAttribute(types.AttributeKeyExpiryHeight, math.NewInt(identity.ExpiryHeight).String()),
					),
				)
			}
		}

		return false
	})
}

// cleanupExpiredMeasurements removes expired measurements from the store
// This is optional and controlled by parameters
func (k Keeper) cleanupExpiredMeasurements(ctx sdk.Context) {
	params := k.GetParams(ctx)

	// Only cleanup if enabled in parameters
	if !params.EnableMeasurementCleanup {
		return
	}

	currentHeight := ctx.BlockHeight()
	cleanupCount := 0

	k.WithMeasurements(ctx, func(measurement types.MeasurementRecord) bool {
		// Skip if not expired
		if measurement.ExpiryHeight == 0 || currentHeight < measurement.ExpiryHeight {
			return false
		}

		// Skip if already revoked (keep revoked measurements for audit)
		if measurement.Revoked {
			return false
		}

		// Delete expired measurement
		store := ctx.KVStore(k.skey)
		store.Delete(types.MeasurementAllowlistKey(measurement.MeasurementHash))
		cleanupCount++

		k.Logger(ctx).Debug(
			"cleaned up expired measurement",
			"measurement", types.MeasurementHashHex(measurement.MeasurementHash),
			"expiry_height", measurement.ExpiryHeight,
		)

		return false
	})

	if cleanupCount > 0 {
		k.Logger(ctx).Info(
			"cleaned up expired measurements",
			"count", cleanupCount,
			"height", currentHeight,
		)
	}
}

// EndBlocker performs any necessary tasks at the end of each block
func (k Keeper) EndBlocker(ctx sdk.Context) {
	// Currently no end-block logic needed
	// This is a placeholder for future functionality
}
