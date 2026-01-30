// Package keeper implements the staking module keeper.
//
// VE-921: Slashing logic for validator misbehavior
package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/staking/types"
)

// ============================================================================
// Slashing Records
// ============================================================================

// GetSlashRecord returns a slashing record
func (k Keeper) GetSlashRecord(ctx sdk.Context, slashID string) (types.SlashRecord, bool) {
	store := ctx.KVStore(k.skey)
	key := types.GetSlashingRecordKey(slashID)
	bz := store.Get(key)
	if bz == nil {
		return types.SlashRecord{}, false
	}

	var record types.SlashRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return types.SlashRecord{}, false
	}
	return record, true
}

// SetSlashRecord stores a slashing record
func (k Keeper) SetSlashRecord(ctx sdk.Context, record types.SlashRecord) error {
	if err := types.ValidateSlashRecord(&record); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := types.GetSlashingRecordKey(record.SlashId)
	bz, err := json.Marshal(record)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// GetSlashingRecordsByValidator returns all slashing records for a validator
func (k Keeper) GetSlashingRecordsByValidator(ctx sdk.Context, validatorAddr string) []types.SlashRecord {
	var records []types.SlashRecord

	k.WithSlashRecords(ctx, func(record types.SlashRecord) bool {
		if record.ValidatorAddress == validatorAddr {
			records = append(records, record)
		}
		return false
	})

	return records
}

// WithSlashRecords iterates over all slash records
func (k Keeper) WithSlashRecords(ctx sdk.Context, fn func(types.SlashRecord) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.SlashingRecordPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var record types.SlashRecord
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}
		if fn(record) {
			break
		}
	}
}

// ============================================================================
// Validator Signing Info
// ============================================================================

// GetValidatorSigningInfo returns a validator's signing info
func (k Keeper) GetValidatorSigningInfo(ctx sdk.Context, validatorAddr string) (types.ValidatorSigningInfo, bool) {
	store := ctx.KVStore(k.skey)
	key := types.GetValidatorSigningInfoKey(validatorAddr)
	bz := store.Get(key)
	if bz == nil {
		return types.ValidatorSigningInfo{}, false
	}

	var info types.ValidatorSigningInfo
	if err := json.Unmarshal(bz, &info); err != nil {
		return types.ValidatorSigningInfo{}, false
	}
	return info, true
}

// SetValidatorSigningInfo stores validator signing info
func (k Keeper) SetValidatorSigningInfo(ctx sdk.Context, info types.ValidatorSigningInfo) error {
	store := ctx.KVStore(k.skey)
	key := types.GetValidatorSigningInfoKey(info.ValidatorAddress)
	bz, err := json.Marshal(info)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// ============================================================================
// Slashing Operations
// ============================================================================

// SlashValidator slashes a validator for misbehavior
func (k Keeper) SlashValidator(ctx sdk.Context, validatorAddr string, reason types.SlashReason, infractionHeight int64, evidence string) (*types.SlashRecord, error) {
	// Check if validator is tombstoned
	signingInfo, found := k.GetValidatorSigningInfo(ctx, validatorAddr)
	if found && signingInfo.Tombstoned {
		return nil, types.ErrValidatorJailed.Wrap("validator is tombstoned")
	}

	// Get slash configuration
	slashConfig := types.GetSlashConfig(reason)

	// Calculate escalation based on previous infractions
	escalatedSlashPercent := slashConfig.SlashPercent
	escalatedJailDuration := slashConfig.JailDuration

	if found {
		// Apply escalation for repeat offenders
		for i := int64(0); i < signingInfo.InfractionCount; i++ {
			escalatedSlashPercent = (escalatedSlashPercent * slashConfig.EscalationMultiplier)
			if escalatedSlashPercent > types.FixedPointScale {
				escalatedSlashPercent = types.FixedPointScale // Cap at 100%
			}
			escalatedJailDuration = escalatedJailDuration * slashConfig.EscalationMultiplier
		}
	}

	// Calculate slash amount (placeholder - would use actual stake in production)
	var slashAmount sdk.Coins
	if k.stakingKeeper != nil {
		validatorAccAddr, _ := sdk.AccAddressFromBech32(validatorAddr)
		stake := k.stakingKeeper.GetValidatorStake(ctx, validatorAccAddr)
		slashTokens := (stake * escalatedSlashPercent) / types.FixedPointScale
		slashAmount = sdk.NewCoins(sdk.NewInt64Coin(k.GetParams(ctx).RewardDenom, slashTokens))
	} else {
		// Placeholder amount
		slashAmount = sdk.NewCoins(sdk.NewInt64Coin(k.GetParams(ctx).RewardDenom, escalatedSlashPercent))
	}

	// Create slash record
	slashID := k.generateSlashID(ctx, validatorAddr, reason)
	slashRecord := types.NewSlashRecord(
		slashID,
		validatorAddr,
		reason,
		slashAmount,
		escalatedSlashPercent,
		infractionHeight,
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)
	slashRecord.Evidence = evidence

	// Apply jail
	jailDuration := time.Duration(escalatedJailDuration) * time.Second
	if slashConfig.IsTombstone {
		slashRecord.Tombstoned = true
		if err := k.TombstoneValidator(ctx, validatorAddr); err != nil {
			k.Logger(ctx).Error("failed to tombstone validator", "error", err)
		}
	} else if escalatedJailDuration > 0 {
		slashRecord.Jailed = true
		slashRecord.JailDuration = escalatedJailDuration
		jailedUntil := ctx.BlockTime().Add(jailDuration)
		slashRecord.JailedUntil = &jailedUntil
		if err := k.JailValidator(ctx, validatorAddr, jailDuration); err != nil {
			k.Logger(ctx).Error("failed to jail validator", "error", err)
		}
	}

	// Store slash record
	if err := k.SetSlashRecord(ctx, *slashRecord); err != nil {
		return nil, err
	}

	// Update signing info
	if !found {
		signingInfo = *types.NewValidatorSigningInfo(validatorAddr, ctx.BlockHeight())
	}
	signingInfo.InfractionCount++
	if slashRecord.Tombstoned {
		signingInfo.Tombstoned = true
	}
	if slashRecord.Jailed {
		signingInfo.JailedUntil = slashRecord.JailedUntil
	}
	if err := k.SetValidatorSigningInfo(ctx, signingInfo); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeValidatorSlashed,
			sdk.NewAttribute(types.AttributeKeyValidatorAddress, validatorAddr),
			sdk.NewAttribute(types.AttributeKeySlashReason, string(reason)),
			sdk.NewAttribute(types.AttributeKeySlashAmount, slashAmount.String()),
			sdk.NewAttribute(types.AttributeKeySlashPercent, sdkmath.LegacyNewDecWithPrec(escalatedSlashPercent, 6).String()),
		),
	)

	k.Logger(ctx).Info("validator slashed",
		"validator", validatorAddr,
		"reason", reason,
		"amount", slashAmount,
		"percent", escalatedSlashPercent,
		"jailed", slashRecord.Jailed,
		"tombstoned", slashRecord.Tombstoned,
	)

	return slashRecord, nil
}

// SlashForDoubleSigning slashes a validator for double signing
func (k Keeper) SlashForDoubleSigning(ctx sdk.Context, validatorAddr string, height int64, evidence types.DoubleSignEvidence) (*types.SlashRecord, error) {
	k.Logger(ctx).Warn("double signing detected",
		"validator", validatorAddr,
		"height1", evidence.Height1,
		"height2", evidence.Height2,
	)

	// Store evidence
	store := ctx.KVStore(k.skey)
	evidenceKey := types.GetDoubleSignEvidenceKey(evidence.EvidenceID)
	evidenceBz, _ := json.Marshal(evidence)
	store.Set(evidenceKey, evidenceBz)

	// Emit double sign event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDoubleSignDetected,
			sdk.NewAttribute(types.AttributeKeyValidatorAddress, validatorAddr),
		),
	)

	// Execute slash
	return k.SlashValidator(ctx, validatorAddr, types.SlashReasonDoubleSigning, height, evidence.EvidenceID)
}

// SlashForDowntime slashes a validator for excessive downtime
func (k Keeper) SlashForDowntime(ctx sdk.Context, validatorAddr string, missedBlocks int64) (*types.SlashRecord, error) {
	params := k.GetParams(ctx)

	// Check if threshold is exceeded
	if missedBlocks < params.DowntimeThreshold {
		return nil, nil // No slash needed
	}

	k.Logger(ctx).Warn("excessive downtime detected",
		"validator", validatorAddr,
		"missed_blocks", missedBlocks,
		"threshold", params.DowntimeThreshold,
	)

	// Emit downtime event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDowntimeDetected,
			sdk.NewAttribute(types.AttributeKeyValidatorAddress, validatorAddr),
			sdk.NewAttribute(types.AttributeKeyBlocksMissed, fmt.Sprintf("%d", missedBlocks)),
		),
	)

	evidence := fmt.Sprintf("missed_blocks:%d,threshold:%d", missedBlocks, params.DowntimeThreshold)
	return k.SlashValidator(ctx, validatorAddr, types.SlashReasonDowntime, ctx.BlockHeight()-missedBlocks, evidence)
}

// SlashForInvalidAttestation slashes a validator for invalid VEID attestation
func (k Keeper) SlashForInvalidAttestation(ctx sdk.Context, validatorAddr string, attestation types.InvalidVEIDAttestation) (*types.SlashRecord, error) {
	k.Logger(ctx).Warn("invalid VEID attestation detected",
		"validator", validatorAddr,
		"attestation_id", attestation.AttestationID,
		"reason", attestation.Reason,
		"score_diff", attestation.ScoreDifference,
	)

	// Store invalid attestation record
	store := ctx.KVStore(k.skey)
	attestationKey := types.GetInvalidAttestationKey(attestation.RecordID)
	attestationBz, _ := json.Marshal(attestation)
	store.Set(attestationKey, attestationBz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeInvalidAttestationDetected,
			sdk.NewAttribute(types.AttributeKeyValidatorAddress, validatorAddr),
		),
	)

	evidence := fmt.Sprintf("attestation:%s,reason:%s,expected:%d,actual:%d",
		attestation.AttestationID, attestation.Reason, attestation.ExpectedScore, attestation.ActualScore)
	return k.SlashValidator(ctx, validatorAddr, types.SlashReasonInvalidVEIDAttestation, attestation.DetectedHeight, evidence)
}

// ============================================================================
// Jailing Operations
// ============================================================================

// JailValidator jails a validator for a duration
func (k Keeper) JailValidator(ctx sdk.Context, validatorAddr string, duration time.Duration) error {
	signingInfo, found := k.GetValidatorSigningInfo(ctx, validatorAddr)
	if !found {
		signingInfo = *types.NewValidatorSigningInfo(validatorAddr, ctx.BlockHeight())
	}

	if signingInfo.Tombstoned {
		return types.ErrValidatorJailed.Wrap("validator is tombstoned")
	}

	jailedUntil := ctx.BlockTime().Add(duration)
	signingInfo.JailedUntil = &jailedUntil

	if err := k.SetValidatorSigningInfo(ctx, signingInfo); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeValidatorJailed,
			sdk.NewAttribute(types.AttributeKeyValidatorAddress, validatorAddr),
			sdk.NewAttribute(types.AttributeKeyJailDuration, fmt.Sprintf("%d", int64(duration.Seconds()))),
		),
	)

	k.Logger(ctx).Info("validator jailed",
		"validator", validatorAddr,
		"duration", duration,
		"until", signingInfo.JailedUntil,
	)

	return nil
}

// UnjailValidator unjails a validator
func (k Keeper) UnjailValidator(ctx sdk.Context, validatorAddr string) error {
	signingInfo, found := k.GetValidatorSigningInfo(ctx, validatorAddr)
	if !found {
		return types.ErrValidatorNotFound.Wrapf("validator %s not found", validatorAddr)
	}

	if signingInfo.Tombstoned {
		return types.ErrValidatorJailed.Wrap("validator is tombstoned and cannot be unjailed")
	}

	// Check if jail period has passed
	if signingInfo.JailedUntil != nil && ctx.BlockTime().Before(*signingInfo.JailedUntil) {
		return types.ErrValidatorJailed.Wrapf("validator still jailed until %s", signingInfo.JailedUntil)
	}

	// Clear jail
	signingInfo.JailedUntil = nil
	signingInfo.MissedBlocksCounter = 0

	if err := k.SetValidatorSigningInfo(ctx, signingInfo); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeValidatorUnjailed,
			sdk.NewAttribute(types.AttributeKeyValidatorAddress, validatorAddr),
		),
	)

	k.Logger(ctx).Info("validator unjailed", "validator", validatorAddr)

	return nil
}

// IsValidatorJailed checks if a validator is currently jailed
func (k Keeper) IsValidatorJailed(ctx sdk.Context, validatorAddr string) bool {
	signingInfo, found := k.GetValidatorSigningInfo(ctx, validatorAddr)
	if !found {
		return false
	}

	return signingInfo.Tombstoned || (signingInfo.JailedUntil != nil && ctx.BlockTime().Before(*signingInfo.JailedUntil))
}

// TombstoneValidator permanently bans a validator
func (k Keeper) TombstoneValidator(ctx sdk.Context, validatorAddr string) error {
	signingInfo, found := k.GetValidatorSigningInfo(ctx, validatorAddr)
	if !found {
		signingInfo = *types.NewValidatorSigningInfo(validatorAddr, ctx.BlockHeight())
	}

	if signingInfo.Tombstoned {
		return nil // Already tombstoned
	}

	signingInfo.Tombstoned = true

	if err := k.SetValidatorSigningInfo(ctx, signingInfo); err != nil {
		return err
	}

	k.Logger(ctx).Warn("validator tombstoned", "validator", validatorAddr)

	return nil
}

// ============================================================================
// Signature Handling
// ============================================================================

// HandleValidatorSignature handles a validator's signature (or lack thereof) for a block
func (k Keeper) HandleValidatorSignature(ctx sdk.Context, validatorAddr string, signed bool) error {
	params := k.GetParams(ctx)

	// Get or create signing info
	signingInfo, found := k.GetValidatorSigningInfo(ctx, validatorAddr)
	if !found {
		signingInfo = *types.NewValidatorSigningInfo(validatorAddr, ctx.BlockHeight())
	}

	// Check if tombstoned
	if signingInfo.Tombstoned {
		return nil // Ignore tombstoned validators
	}

	// Update signing info
	if signed {
		// Reset missed blocks counter on signing
		if signingInfo.MissedBlocksCounter > 0 {
			signingInfo.MissedBlocksCounter = 0
		}
	} else {
		// Increment missed blocks
		signingInfo.MissedBlocksCounter++

		// Check if we should slash for downtime
		if signingInfo.MissedBlocksCounter >= params.DowntimeThreshold {
			if _, err := k.SlashForDowntime(ctx, validatorAddr, signingInfo.MissedBlocksCounter); err != nil {
				k.Logger(ctx).Error("failed to slash for downtime", "error", err)
			}
			signingInfo.MissedBlocksCounter = 0
		}
	}

	// Update performance
	update := PerformanceUpdate{
		BlockSigned: signed,
		BlockMissed: !signed,
	}
	if err := k.UpdateValidatorPerformance(ctx, validatorAddr, update); err != nil {
		k.Logger(ctx).Error("failed to update performance", "error", err)
	}

	return k.SetValidatorSigningInfo(ctx, signingInfo)
}
