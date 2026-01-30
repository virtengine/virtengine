package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// VEID Tier Transition Logic
// ============================================================================
//
// Implements the Account Tier System as defined in veid-flow-spec.md:
// - Tier 0 (Unverified): Score 0-49 or non-verified status
// - Tier 1 (Basic): Score 50-69
// - Tier 2 (Standard): Score 70-84
// - Tier 3 (Premium): Score 85-100
//
// Task Reference: VEID-CORE-001
// ============================================================================

// TierTransitionResult contains the result of a tier transition operation
type TierTransitionResult struct {
	// Address is the account address
	Address string `json:"address"`

	// OldTier is the tier before the transition
	OldTier int `json:"old_tier"`

	// NewTier is the tier after the transition
	NewTier int `json:"new_tier"`

	// CompositeScore is the calculated composite score
	CompositeScore uint32 `json:"composite_score"`

	// Changed indicates if the tier actually changed
	Changed bool `json:"changed"`

	// VerifiedScopeCount is the number of verified scopes
	VerifiedScopeCount int `json:"verified_scope_count"`
}

// UpdateAccountTier fetches all verified scopes for an account, calculates
// the composite score, maps it to a tier, stores the new tier, and emits
// a VEIDTierChanged event if the tier changed.
//
// Tier mapping (per veid-flow-spec.md):
// - Score 0-49: Tier 0 (Unverified)
// - Score 50-69: Tier 1 (Basic)
// - Score 70-84: Tier 2 (Standard)
// - Score 85-100: Tier 3 (Premium)
func (k Keeper) UpdateAccountTier(ctx sdk.Context, addr sdk.AccAddress) (*TierTransitionResult, error) {
	// Get identity record
	record, found := k.GetIdentityRecord(ctx, addr)
	if !found {
		return nil, types.ErrIdentityRecordNotFound.Wrapf("identity record not found for %s", addr.String())
	}

	// Store old tier for comparison
	oldTier := k.computeTierInt(record.Tier)

	// Calculate composite score from verified scopes
	compositeScore, verifiedCount := k.calculateCompositeScore(ctx, addr)

	// Determine new tier from composite score
	newTier := k.determineTier(compositeScore, record)

	// Create result
	result := &TierTransitionResult{
		Address:            addr.String(),
		OldTier:            oldTier,
		NewTier:            newTier,
		CompositeScore:     compositeScore,
		Changed:            oldTier != newTier,
		VerifiedScopeCount: verifiedCount,
	}

	// Update record if score or tier changed
	record.CurrentScore = compositeScore
	record.Tier = k.tierIntToIdentityTier(newTier)
	record.UpdatedAt = ctx.BlockTime()

	if err := k.SetIdentityRecord(ctx, record); err != nil {
		return nil, err
	}

	// Emit tier changed event if tier changed
	if result.Changed {
		if err := k.EmitTierChangedEvent(
			ctx,
			addr.String(),
			types.TierToString(oldTier),
			types.TierToString(newTier),
			compositeScore,
		); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// calculateCompositeScore calculates the composite identity score from all
// verified scopes for an account. Each scope type has a weight defined in
// types.ScopeTypeWeight().
//
// Returns the composite score (capped at 100) and count of verified scopes.
func (k Keeper) calculateCompositeScore(ctx sdk.Context, addr sdk.AccAddress) (uint32, int) {
	var totalWeight uint32
	var verifiedCount int

	// Iterate through all scopes for the account
	k.WithScopes(ctx, addr, func(scope types.IdentityScope) bool {
		// Only count verified, non-revoked scopes
		if scope.Status == types.VerificationStatusVerified && scope.IsActive() {
			weight := types.ScopeTypeWeight(scope.ScopeType)
			totalWeight += weight
			verifiedCount++
		}
		return false // continue iteration
	})

	// Cap the score at MaxScore (100)
	if totalWeight > types.MaxScore {
		totalWeight = types.MaxScore
	}

	return totalWeight, verifiedCount
}

// determineTier maps a score to a tier based on the thresholds defined
// in veid-flow-spec.md. Non-verified accounts are always Tier 0.
func (k Keeper) determineTier(score uint32, record types.IdentityRecord) int {
	// Locked accounts are Tier 0
	if record.Locked {
		return types.TierUnverified
	}

	// Apply tier thresholds per spec
	switch {
	case score >= types.ThresholdPremium: // 85-100
		return types.TierPremium
	case score >= types.ThresholdStandard: // 70-84
		return types.TierStandard
	case score >= types.ThresholdBasic: // 50-69
		return types.TierBasic
	default: // 0-49
		return types.TierUnverified
	}
}

// computeTierInt converts an IdentityTier string to its numeric tier value
func (k Keeper) computeTierInt(tier types.IdentityTier) int {
	switch tier {
	case types.IdentityTierPremium, types.IdentityTierTrusted:
		return types.TierPremium
	case types.IdentityTierStandard, types.IdentityTierVerified:
		return types.TierStandard
	case types.IdentityTierBasic:
		return types.TierBasic
	default:
		return types.TierUnverified
	}
}

// tierIntToIdentityTier converts a numeric tier to IdentityTier
func (k Keeper) tierIntToIdentityTier(tier int) types.IdentityTier {
	switch tier {
	case types.TierPremium:
		return types.IdentityTierPremium
	case types.TierStandard:
		return types.IdentityTierStandard
	case types.TierBasic:
		return types.IdentityTierBasic
	default:
		return types.IdentityTierUnverified
	}
}

// ============================================================================
// Query Helpers
// ============================================================================

// GetAccountTierDetails returns detailed tier information for an account.
// This extends the GetAccountTier method with additional context.
type AccountTierDetails struct {
	// Address is the account address
	Address string `json:"address"`

	// Tier is the current tier (0-3)
	Tier int `json:"tier"`

	// TierName is the human-readable tier name
	TierName string `json:"tier_name"`

	// Score is the current composite score
	Score uint32 `json:"score"`

	// Status is the account verification status
	Status types.AccountStatus `json:"status"`

	// NextTierThreshold is the score needed for the next tier (0 if at max)
	NextTierThreshold uint32 `json:"next_tier_threshold"`

	// PointsToNextTier is how many more points needed for next tier
	PointsToNextTier uint32 `json:"points_to_next_tier"`

	// VerifiedScopeCount is the number of verified scopes
	VerifiedScopeCount int `json:"verified_scope_count"`

	// Locked indicates if the account is locked
	Locked bool `json:"locked"`
}

// GetAccountTierDetails returns detailed tier information for an account
func (k Keeper) GetAccountTierDetails(ctx sdk.Context, addr sdk.AccAddress) (*AccountTierDetails, error) {
	record, found := k.GetIdentityRecord(ctx, addr)
	if !found {
		return nil, types.ErrIdentityRecordNotFound.Wrapf("identity record not found for %s", addr.String())
	}

	// Get current score and status
	score, status, _ := k.GetScore(ctx, addr.String())
	if score == 0 {
		score = record.CurrentScore
		if !record.Locked && record.CurrentScore > 0 {
			status = types.AccountStatusVerified
		}
	}

	tier := types.ComputeTierFromScoreValue(score, status)

	// Calculate next tier threshold
	var nextThreshold uint32
	var pointsNeeded uint32

	switch tier {
	case types.TierUnverified:
		nextThreshold = types.ThresholdBasic
		if score < nextThreshold {
			pointsNeeded = nextThreshold - score
		}
	case types.TierBasic:
		nextThreshold = types.ThresholdStandard
		pointsNeeded = nextThreshold - score
	case types.TierStandard:
		nextThreshold = types.ThresholdPremium
		pointsNeeded = nextThreshold - score
	case types.TierPremium:
		nextThreshold = 0 // Already at max
		pointsNeeded = 0
	}

	// Count verified scopes
	verifiedCount := record.CountVerifiedScopes()

	return &AccountTierDetails{
		Address:            addr.String(),
		Tier:               tier,
		TierName:           types.TierToString(tier),
		Score:              score,
		Status:             status,
		NextTierThreshold:  nextThreshold,
		PointsToNextTier:   pointsNeeded,
		VerifiedScopeCount: verifiedCount,
		Locked:             record.Locked,
	}, nil
}

// ============================================================================
// Threshold Helpers
// ============================================================================

// MeetsScoreThreshold checks if an account's score meets or exceeds the
// specified threshold. Returns false if the account doesn't exist,
// isn't verified, or is locked.
func (k Keeper) MeetsScoreThreshold(ctx sdk.Context, addr sdk.AccAddress, threshold uint32) bool {
	// Get identity record
	record, found := k.GetIdentityRecord(ctx, addr)
	if !found {
		return false
	}

	// Locked accounts never meet thresholds
	if record.Locked {
		return false
	}

	// Get the current score and status
	score, status, found := k.GetScore(ctx, addr.String())
	if !found {
		// Fall back to record's score
		score = record.CurrentScore
		if record.CurrentScore == 0 {
			return false
		}
		status = types.AccountStatusVerified
	}

	// Must be verified to pass threshold checks
	if status != types.AccountStatusVerified {
		return false
	}

	return score >= threshold
}

// MeetsTierRequirement checks if an account meets the minimum tier requirement
func (k Keeper) MeetsTierRequirement(ctx sdk.Context, addr sdk.AccAddress, requiredTier int) bool {
	tier, err := k.GetAccountTier(ctx, addr.String())
	if err != nil {
		return false
	}
	return tier >= requiredTier
}

// GetTierThreshold returns the minimum score threshold for a tier
func (k Keeper) GetTierThreshold(tier int) uint32 {
	return types.GetMinimumScoreForTier(tier)
}

// ============================================================================
// Batch Operations
// ============================================================================

// RecalculateAllAccountTiers recalculates tiers for all accounts.
// This is useful after ML model updates or system-wide recalibration.
// Returns the number of accounts processed and number of tier changes.
func (k Keeper) RecalculateAllAccountTiers(ctx sdk.Context) (processed int, changed int) {
	k.WithIdentityRecords(ctx, func(record types.IdentityRecord) bool {
		addr, err := sdk.AccAddressFromBech32(record.AccountAddress)
		if err != nil {
			return false // continue
		}

		result, err := k.UpdateAccountTier(ctx, addr)
		if err != nil {
			k.Logger(ctx).Error("failed to recalculate tier",
				"address", record.AccountAddress,
				"error", err,
			)
			return false // continue
		}

		processed++
		if result.Changed {
			changed++
		}

		return false // continue
	})

	return processed, changed
}
