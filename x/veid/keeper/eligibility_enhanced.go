package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Enhanced Eligibility Keeper (VE-3033)
// Provides detailed eligibility checking with remediation hints
// ============================================================================

// EnhancedCheckEligibility performs a comprehensive eligibility check with detailed results
// This method provides more information than the basic eligibility checks, including
// remediation hints and next steps for achieving eligibility.
func (k Keeper) EnhancedCheckEligibility(
	ctx sdk.Context,
	address sdk.AccAddress,
	requiredTier types.IdentityTier,
	requireMFA bool,
) (*types.EnhancedEligibilityResult, error) {
	now := ctx.BlockTime()
	result := types.NewEnhancedEligibilityResult(address.String(), now)
	result.RequiredTier = requiredTier

	// Get identity record
	record, found := k.GetIdentityRecord(ctx, address)
	if !found {
		// No identity record - not eligible
		result.CurrentTier = types.IdentityTierUnverified
		result.CurrentScore = 0
		result.AddCheck(types.NewEligibilityCheck(
			types.EligibilityCheckTypeAccountStatus,
			false,
			"No identity record found for this address",
		).WithValues("not_found", "verified"))

		result.AddRemediation(types.RemediationHint{
			Issue:         "Identity not registered",
			Action:        "Create an identity record and complete initial verification",
			EstimatedTime: 30 * time.Minute,
			Priority:      1,
			Category:      "registration",
		})
		result.AddNextStep("1. Register your identity with the VEID system")
		result.AddNextStep("2. Complete email verification")
		result.Finalize()
		return result, nil
	}

	// Set current values
	result.CurrentTier = record.Tier
	result.CurrentScore = record.CurrentScore
	result.RequiredScore = k.getMinScoreForTier(requiredTier)

	// Get verified scopes
	verifiedScopes := k.getVerifiedScopeTypes(ctx, address)
	result.VerifiedScopes = verifiedScopes

	// Get required scopes for target tier
	tierReqs := types.GetTierRequirements(requiredTier)
	if tierReqs != nil {
		missingScopes := k.findMissingScopes(verifiedScopes, tierReqs.RequiredScopes)
		result.MissingScopes = missingScopes
	}

	// Perform all eligibility checks
	k.checkAccountStatus(ctx, &record, result)
	k.checkTierRequirement(ctx, &record, requiredTier, result)
	k.checkScoreThreshold(ctx, &record, requiredTier, result)
	k.checkRequiredScopes(ctx, address, requiredTier, result)
	k.checkVerificationAge(ctx, &record, result)

	if requireMFA {
		k.checkMFARequirement(ctx, address, result)
	}

	// Generate remediation hints for failed checks
	k.generateRemediationHints(result)

	// Generate next steps
	k.generateNextSteps(result)

	// Finalize the result
	result.Finalize()

	k.Logger(ctx).Debug("Enhanced eligibility check completed",
		"address", address.String(),
		"eligible", result.IsEligible,
		"current_tier", result.CurrentTier,
		"required_tier", result.RequiredTier,
		"failed_checks", len(result.FailedChecks),
	)

	return result, nil
}

// EnhancedCheckEligibilityForMarket performs enhanced eligibility check for market participation
func (k Keeper) EnhancedCheckEligibilityForMarket(
	ctx sdk.Context,
	address sdk.AccAddress,
	marketType types.MarketType,
) (*types.EnhancedEligibilityResult, error) {
	// Get required level for market type
	requiredLevel, err := k.GetRequiredVEIDLevel(ctx, marketType)
	if err != nil {
		return nil, err
	}

	// Convert VEID level to identity tier
	requiredTier := k.veidLevelToTier(requiredLevel)

	// Check if market requires MFA (default based on market type sensitivity)
	_, found := k.GetMarketRequirements(ctx, marketType)
	requireMFA := false
	if found {
		// TEE/GPU markets typically require MFA
		requireMFA = marketType == types.MarketTypeTEE || marketType == types.MarketTypeGPU
	}

	// Perform enhanced eligibility check
	result, err := k.EnhancedCheckEligibility(ctx, address, requiredTier, requireMFA)
	if err != nil {
		return nil, err
	}

	// Add market context
	result.Context.MarketType = marketType
	result.Context.RequiresMFA = requireMFA

	return result, nil
}

// EnhancedCheckEligibilityForOffering performs enhanced eligibility check for an offering type
func (k Keeper) EnhancedCheckEligibilityForOffering(
	ctx sdk.Context,
	address sdk.AccAddress,
	offeringType types.OfferingType,
) (*types.EnhancedEligibilityResult, error) {
	// Determine required tier based on offering type
	requiredTier := k.getRequiredTierForOffering(offeringType)
	requireMFA := k.doesOfferingRequireMFA(offeringType)

	// Perform enhanced eligibility check
	result, err := k.EnhancedCheckEligibility(ctx, address, requiredTier, requireMFA)
	if err != nil {
		return nil, err
	}

	// Add offering context
	result.Context.OfferingType = offeringType
	result.Context.RequiresMFA = requireMFA

	return result, nil
}

// GetRemediationHints returns specific remediation steps for an address
func (k Keeper) GetRemediationHints(
	ctx sdk.Context,
	address sdk.AccAddress,
	targetTier types.IdentityTier,
) ([]types.RemediationHint, error) {
	hints := make([]types.RemediationHint, 0)

	// Get identity record
	record, found := k.GetIdentityRecord(ctx, address)
	if !found {
		hints = append(hints, types.RemediationHint{
			Issue:         "Identity not registered",
			Action:        "Create an identity record with the VEID system",
			EstimatedTime: 30 * time.Minute,
			Priority:      1,
			Category:      "registration",
		})
		return hints, nil
	}

	// Get tier requirements
	tierReqs := types.GetTierRequirements(targetTier)
	if tierReqs == nil {
		return hints, nil
	}

	// Check score
	if record.CurrentScore < tierReqs.MinScore {
		hint := types.RemediationScoreImprovement
		hint.Issue = fmt.Sprintf("Current score (%d) below required minimum (%d)", record.CurrentScore, tierReqs.MinScore)
		hints = append(hints, hint)
	}

	// Check missing scopes
	verifiedScopes := k.getVerifiedScopeTypes(ctx, address)
	missingScopes := k.findMissingScopes(verifiedScopes, tierReqs.RequiredScopes)

	for _, scopeType := range missingScopes {
		hints = append(hints, types.GetRemediationForScope(scopeType))
	}

	// Check MFA if required - use HasActiveFactorOfType instead
	if tierReqs.RequiresMFA && k.mfaKeeper != nil {
		// Check if user has any enrolled MFA factors
		factors := k.mfaKeeper.GetFactorEnrollments(ctx, address)
		if len(factors) == 0 {
			hints = append(hints, types.RemediationMFA)
		}
	}

	// Sort by priority
	k.sortRemediationHints(hints)

	return hints, nil
}

// GetTierRequirements returns the requirements for a specific tier
func (k Keeper) GetTierRequirements(
	ctx sdk.Context,
	tier types.IdentityTier,
) (*types.TierRequirements, error) {
	reqs := types.GetTierRequirements(tier)
	if reqs == nil {
		return nil, types.ErrInvalidParams.Wrapf("unknown tier: %s", tier)
	}
	return reqs, nil
}

// GetAllTierRequirements returns requirements for all tiers
func (k Keeper) GetAllTierRequirements(ctx sdk.Context) []*types.TierRequirements {
	return types.GetAllTierRequirements()
}

// ExplainFailure returns a human-readable explanation of eligibility failure
func (k Keeper) ExplainFailure(
	ctx sdk.Context,
	address sdk.AccAddress,
	requiredTier types.IdentityTier,
) (*types.FailureExplanation, error) {
	// Perform enhanced eligibility check
	result, err := k.EnhancedCheckEligibility(ctx, address, requiredTier, false)
	if err != nil {
		return nil, err
	}

	// If eligible, no failure to explain
	if result.IsEligible {
		return nil, nil
	}

	// Create failure explanation
	explanation := types.NewFailureExplanation(result)
	return explanation, nil
}

// GetProgressToTier returns progress information toward achieving a tier
func (k Keeper) GetProgressToTier(
	ctx sdk.Context,
	address sdk.AccAddress,
	targetTier types.IdentityTier,
) (*TierProgress, error) {
	record, found := k.GetIdentityRecord(ctx, address)
	if !found {
		return &TierProgress{
			TargetTier:      targetTier,
			CurrentTier:     types.IdentityTierUnverified,
			PercentComplete: 0,
			ScopesComplete:  0,
			ScopesRequired:  len(types.GetRequiredScopesForTier(targetTier)),
			ScoreProgress:   0,
			RemainingSteps:  k.getInitialStepsForTier(targetTier),
			EstimatedTime:   60 * time.Minute,
		}, nil
	}

	tierReqs := types.GetTierRequirements(targetTier)
	if tierReqs == nil {
		return nil, types.ErrInvalidParams.Wrapf("unknown tier: %s", targetTier)
	}

	// Calculate scope progress
	verifiedScopes := k.getVerifiedScopeTypes(ctx, address)
	missingScopes := k.findMissingScopes(verifiedScopes, tierReqs.RequiredScopes)
	scopesComplete := len(tierReqs.RequiredScopes) - len(missingScopes)
	scopesRequired := len(tierReqs.RequiredScopes)

	// Calculate score progress
	scoreProgress := float64(record.CurrentScore) / float64(tierReqs.MinScore) * 100
	if scoreProgress > 100 {
		scoreProgress = 100
	}

	// Calculate overall progress (weighted: 60% score, 40% scopes)
	scopeProgress := float64(scopesComplete) / float64(scopesRequired) * 100
	if scopesRequired == 0 {
		scopeProgress = 100
	}
	overallProgress := (scoreProgress * 0.6) + (scopeProgress * 0.4)

	// Get remaining steps
	remainingSteps := make([]string, 0)
	for _, scope := range missingScopes {
		remainingSteps = append(remainingSteps, fmt.Sprintf("Complete %s verification", types.ScopeTypeDescription(scope)))
	}
	if record.CurrentScore < tierReqs.MinScore {
		remainingSteps = append(remainingSteps, fmt.Sprintf("Increase score from %d to %d", record.CurrentScore, tierReqs.MinScore))
	}

	// Estimate time
	estimatedTime := time.Duration(len(missingScopes)) * 15 * time.Minute

	return &TierProgress{
		TargetTier:      targetTier,
		CurrentTier:     record.Tier,
		PercentComplete: overallProgress,
		ScopesComplete:  scopesComplete,
		ScopesRequired:  scopesRequired,
		ScoreProgress:   scoreProgress,
		CurrentScore:    record.CurrentScore,
		RequiredScore:   tierReqs.MinScore,
		MissingScopes:   missingScopes,
		RemainingSteps:  remainingSteps,
		EstimatedTime:   estimatedTime,
	}, nil
}

// TierProgress represents progress toward achieving a tier
type TierProgress struct {
	TargetTier      types.IdentityTier `json:"target_tier"`
	CurrentTier     types.IdentityTier `json:"current_tier"`
	PercentComplete float64            `json:"percent_complete"`
	ScopesComplete  int                `json:"scopes_complete"`
	ScopesRequired  int                `json:"scopes_required"`
	ScoreProgress   float64            `json:"score_progress"`
	CurrentScore    uint32             `json:"current_score"`
	RequiredScore   uint32             `json:"required_score"`
	MissingScopes   []types.ScopeType  `json:"missing_scopes,omitempty"`
	RemainingSteps  []string           `json:"remaining_steps,omitempty"`
	EstimatedTime   time.Duration      `json:"estimated_time,omitempty"`
}

// ============================================================================
// Internal Helper Methods
// ============================================================================

// checkAccountStatus verifies the account status
func (k Keeper) checkAccountStatus(_ sdk.Context, record *types.IdentityRecord, result *types.EnhancedEligibilityResult) {
	if record.Locked {
		result.AddCheck(types.NewEligibilityCheck(
			types.EligibilityCheckTypeAccountStatus,
			false,
			"Account is locked",
		).WithValues("locked", "active"))
		return
	}

	result.AddCheck(types.NewEligibilityCheck(
		types.EligibilityCheckTypeAccountStatus,
		true,
		"Account is active",
	).WithValues("active", "active"))
}

// checkTierRequirement verifies the tier requirement
func (k Keeper) checkTierRequirement(_ sdk.Context, record *types.IdentityRecord, requiredTier types.IdentityTier, result *types.EnhancedEligibilityResult) {
	currentTierLevel := k.tierToLevel(record.Tier)
	requiredTierLevel := k.tierToLevel(requiredTier)

	if currentTierLevel >= requiredTierLevel {
		result.AddCheck(types.NewEligibilityCheck(
			types.EligibilityCheckTypeTierRequirement,
			true,
			fmt.Sprintf("Current tier (%s) meets required tier (%s)", record.Tier, requiredTier),
		).WithValues(string(record.Tier), string(requiredTier)))
		return
	}

	result.AddCheck(types.NewEligibilityCheck(
		types.EligibilityCheckTypeTierRequirement,
		false,
		fmt.Sprintf("Current tier (%s) does not meet required tier (%s)", record.Tier, requiredTier),
	).WithValues(string(record.Tier), string(requiredTier)))
}

// checkScoreThreshold verifies the score threshold
func (k Keeper) checkScoreThreshold(_ sdk.Context, record *types.IdentityRecord, requiredTier types.IdentityTier, result *types.EnhancedEligibilityResult) {
	minScore := k.getMinScoreForTier(requiredTier)

	if record.CurrentScore >= minScore {
		result.AddCheck(types.NewEligibilityCheck(
			types.EligibilityCheckTypeScoreThreshold,
			true,
			fmt.Sprintf("Current score (%d) meets minimum threshold (%d)", record.CurrentScore, minScore),
		).WithValues(record.CurrentScore, minScore))
		return
	}

	result.AddCheck(types.NewEligibilityCheck(
		types.EligibilityCheckTypeScoreThreshold,
		false,
		fmt.Sprintf("Current score (%d) below minimum threshold (%d)", record.CurrentScore, minScore),
	).WithValues(record.CurrentScore, minScore))
}

// checkRequiredScopes verifies required scopes are present
func (k Keeper) checkRequiredScopes(ctx sdk.Context, address sdk.AccAddress, requiredTier types.IdentityTier, result *types.EnhancedEligibilityResult) {
	tierReqs := types.GetTierRequirements(requiredTier)
	if tierReqs == nil || len(tierReqs.RequiredScopes) == 0 {
		result.AddCheck(types.NewEligibilityCheck(
			types.EligibilityCheckTypeScopeRequired,
			true,
			"No specific scopes required for this tier",
		))
		return
	}

	verifiedScopes := k.getVerifiedScopeTypes(ctx, address)
	missingScopes := k.findMissingScopes(verifiedScopes, tierReqs.RequiredScopes)

	if len(missingScopes) == 0 {
		result.AddCheck(types.NewEligibilityCheck(
			types.EligibilityCheckTypeScopeRequired,
			true,
			fmt.Sprintf("All required scopes verified (%d/%d)", len(tierReqs.RequiredScopes), len(tierReqs.RequiredScopes)),
		).WithValues(len(verifiedScopes), len(tierReqs.RequiredScopes)))
		return
	}

	missingScopeStrings := make([]string, len(missingScopes))
	for i, s := range missingScopes {
		missingScopeStrings[i] = string(s)
	}

	result.AddCheck(types.NewEligibilityCheck(
		types.EligibilityCheckTypeScopeRequired,
		false,
		fmt.Sprintf("Missing required scopes: %v", missingScopeStrings),
	).WithValues(len(verifiedScopes), len(tierReqs.RequiredScopes)))
}

// checkVerificationAge checks if verification is still valid
func (k Keeper) checkVerificationAge(ctx sdk.Context, record *types.IdentityRecord, result *types.EnhancedEligibilityResult) {
	params := k.GetParams(ctx)
	if params.VerificationExpiryDays == 0 {
		result.AddCheck(types.NewEligibilityCheck(
			types.EligibilityCheckTypeVerificationAge,
			true,
			"Verification expiry not configured",
		))
		return
	}

	if record.LastVerifiedAt == nil {
		result.AddCheck(types.NewEligibilityCheck(
			types.EligibilityCheckTypeVerificationAge,
			false,
			"Account has never been verified",
		).WithValues("never", "within expiry"))
		return
	}

	expiryDuration := time.Duration(params.VerificationExpiryDays) * 24 * time.Hour
	expiresAt := record.LastVerifiedAt.Add(expiryDuration)
	now := ctx.BlockTime()

	result.Context.LastVerifiedAt = record.LastVerifiedAt
	result.Context.VerificationExpiresAt = &expiresAt

	if now.After(expiresAt) {
		result.AddCheck(types.NewEligibilityCheck(
			types.EligibilityCheckTypeVerificationAge,
			false,
			fmt.Sprintf("Verification expired on %s", expiresAt.Format(time.RFC3339)),
		).WithValues(now.Format(time.RFC3339), expiresAt.Format(time.RFC3339)))
		return
	}

	result.AddCheck(types.NewEligibilityCheck(
		types.EligibilityCheckTypeVerificationAge,
		true,
		fmt.Sprintf("Verification valid until %s", expiresAt.Format(time.RFC3339)),
	).WithValues(now.Format(time.RFC3339), expiresAt.Format(time.RFC3339)))
}

// checkMFARequirement checks if MFA is enabled when required
func (k Keeper) checkMFARequirement(ctx sdk.Context, address sdk.AccAddress, result *types.EnhancedEligibilityResult) {
	if k.mfaKeeper == nil {
		result.AddCheck(types.NewEligibilityCheck(
			types.EligibilityCheckTypeMFARequired,
			true,
			"MFA module not available - check skipped",
		))
		return
	}

	// Check if user has any enrolled factors
	factors := k.mfaKeeper.GetFactorEnrollments(ctx, address)
	enabled := len(factors) > 0
	result.Context.MFAEnabled = enabled
	result.Context.RequiresMFA = true

	if enabled {
		result.AddCheck(types.NewEligibilityCheck(
			types.EligibilityCheckTypeMFARequired,
			true,
			"MFA is enabled",
		).WithValues(true, true))
		return
	}

	result.AddCheck(types.NewEligibilityCheck(
		types.EligibilityCheckTypeMFARequired,
		false,
		"MFA is required but not enabled",
	).WithValues(false, true))
}

// generateRemediationHints generates remediation hints for failed checks
func (k Keeper) generateRemediationHints(result *types.EnhancedEligibilityResult) {
	for _, check := range result.FailedChecks {
		switch check.CheckType {
		case types.EligibilityCheckTypeAccountStatus:
			result.AddRemediation(types.RemediationHint{
				Issue:         "Account status issue",
				Action:        "Contact support to resolve account status",
				EstimatedTime: 24 * time.Hour,
				Priority:      1,
				Category:      "account",
			})

		case types.EligibilityCheckTypeTierRequirement, types.EligibilityCheckTypeScoreThreshold:
			result.AddRemediation(types.RemediationScoreImprovement)

		case types.EligibilityCheckTypeScopeRequired:
			for _, scope := range result.MissingScopes {
				result.AddRemediation(types.GetRemediationForScope(scope))
			}

		case types.EligibilityCheckTypeMFARequired:
			result.AddRemediation(types.RemediationMFA)

		case types.EligibilityCheckTypeVerificationAge:
			result.AddRemediation(types.RemediationHint{
				Issue:         "Verification has expired",
				Action:        "Re-verify your identity to renew your verification status",
				EstimatedTime: 30 * time.Minute,
				Priority:      1,
				Category:      "reverification",
			})

		case types.EligibilityCheckTypeLivenessRequired:
			result.AddRemediation(types.RemediationLiveness)

		case types.EligibilityCheckTypeBiometricRequired:
			result.AddRemediation(types.GetRemediationForScope(types.ScopeTypeBiometric))
		}
	}
}

// generateNextSteps generates ordered next steps for achieving eligibility
func (k Keeper) generateNextSteps(result *types.EnhancedEligibilityResult) {
	stepNum := 1

	// Sort remediation hints by priority
	k.sortRemediationHints(result.Remediation)

	// Convert to numbered steps
	for _, hint := range result.Remediation {
		step := fmt.Sprintf("%d. %s", stepNum, hint.Action)
		result.AddNextStep(step)
		stepNum++
	}
}

// sortRemediationHints sorts hints by priority
func (k Keeper) sortRemediationHints(hints []types.RemediationHint) {
	// Simple bubble sort by priority (lower = higher priority)
	for i := 0; i < len(hints)-1; i++ {
		for j := 0; j < len(hints)-i-1; j++ {
			if hints[j].Priority > hints[j+1].Priority {
				hints[j], hints[j+1] = hints[j+1], hints[j]
			}
		}
	}
}

// getVerifiedScopeTypes returns the scope types that are verified for an address
func (k Keeper) getVerifiedScopeTypes(ctx sdk.Context, address sdk.AccAddress) []types.ScopeType {
	scopeTypes := make([]types.ScopeType, 0)
	seen := make(map[types.ScopeType]bool)

	k.WithScopes(ctx, address, func(scope types.IdentityScope) bool {
		if scope.Status == types.VerificationStatusVerified && !seen[scope.ScopeType] {
			scopeTypes = append(scopeTypes, scope.ScopeType)
			seen[scope.ScopeType] = true
		}
		return false
	})

	return scopeTypes
}

// findMissingScopes finds scopes that are required but not verified
func (k Keeper) findMissingScopes(verified []types.ScopeType, required []types.ScopeType) []types.ScopeType {
	verifiedMap := make(map[types.ScopeType]bool)
	for _, s := range verified {
		verifiedMap[s] = true
	}

	missing := make([]types.ScopeType, 0)
	for _, s := range required {
		if !verifiedMap[s] {
			missing = append(missing, s)
		}
	}
	return missing
}

// getMinScoreForTier returns the minimum score for a tier
func (k Keeper) getMinScoreForTier(tier types.IdentityTier) uint32 {
	reqs := types.GetTierRequirements(tier)
	if reqs == nil {
		return 0
	}
	return reqs.MinScore
}

// tierToLevel converts tier to numeric level for comparison
func (k Keeper) tierToLevel(tier types.IdentityTier) int {
	switch tier {
	case types.IdentityTierUnverified:
		return 0
	case types.IdentityTierBasic:
		return 1
	case types.IdentityTierStandard:
		return 2
	case types.IdentityTierPremium:
		return 3
	default:
		return 0
	}
}

// veidLevelToTier converts VEID level to identity tier
func (k Keeper) veidLevelToTier(level types.VEIDLevel) types.IdentityTier {
	switch level {
	case types.VEIDLevelNone:
		return types.IdentityTierUnverified
	case types.VEIDLevelBasic:
		return types.IdentityTierBasic
	case types.VEIDLevelStandard:
		return types.IdentityTierStandard
	case types.VEIDLevelPremium, types.VEIDLevelEnterprise:
		return types.IdentityTierPremium
	default:
		return types.IdentityTierUnverified
	}
}

// getRequiredTierForOffering returns the required tier for an offering type
func (k Keeper) getRequiredTierForOffering(offeringType types.OfferingType) types.IdentityTier {
	switch offeringType {
	case types.OfferingTypeValidator:
		return types.IdentityTierPremium
	case types.OfferingTypeProvider:
		return types.IdentityTierStandard
	case types.OfferingTypePremium:
		return types.IdentityTierStandard
	case types.OfferingTypeStandard:
		return types.IdentityTierBasic
	case types.OfferingTypeBasic:
		return types.IdentityTierBasic
	default:
		return types.IdentityTierBasic
	}
}

// doesOfferingRequireMFA returns whether the offering requires MFA
func (k Keeper) doesOfferingRequireMFA(offeringType types.OfferingType) bool {
	switch offeringType {
	case types.OfferingTypeValidator, types.OfferingTypeProvider:
		return true
	default:
		return false
	}
}

// getInitialStepsForTier returns initial steps for unregistered users
func (k Keeper) getInitialStepsForTier(tier types.IdentityTier) []string {
	steps := []string{
		"Register your identity with the VEID system",
		"Complete email verification",
	}

	tierReqs := types.GetTierRequirements(tier)
	if tierReqs == nil {
		return steps
	}

	for _, scope := range tierReqs.RequiredScopes {
		if scope != types.ScopeTypeEmailProof {
			steps = append(steps, fmt.Sprintf("Complete %s", types.ScopeTypeDescription(scope)))
		}
	}

	if tierReqs.RequiresMFA {
		steps = append(steps, "Enable multi-factor authentication")
	}

	return steps
}
