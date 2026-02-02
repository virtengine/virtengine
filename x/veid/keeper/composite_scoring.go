// Package keeper provides keeper functions for the VEID module.
//
// VEID-CORE-002: Identity Scope Scoring Algorithm
// This file implements keeper integration for the spec-compliant composite scoring algorithm.
package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Composite Scoring Keeper Methods
// ============================================================================

// ComputeCompositeIdentityScore computes the identity score using the spec-compliant
// composite scoring algorithm per veid-flow-spec.md.
//
// This method:
//  1. Uses the spec-defined weights (Doc Auth 25%, Face Match 25%, Liveness 20%,
//     Data Consistency 15%, Historical 10%, Risk 5%)
//  2. Stores the score version with each computed score
//  3. Ensures deterministic computation for consensus
//  4. Returns a score 0-100 with reason codes
func (k Keeper) ComputeCompositeIdentityScore(
	ctx sdk.Context,
	inputs types.CompositeScoringInputs,
) (*types.CompositeScoreResult, error) {
	// Set context values in inputs
	inputs.BlockHeight = ctx.BlockHeight()
	inputs.Timestamp = ctx.BlockTime()

	// Get weights and thresholds (use defaults for now, can be made configurable via params)
	weights := types.DefaultCompositeScoringWeights()
	thresholds := types.DefaultCompositeScoringThresholds()

	// Compute the composite score
	result, err := types.ComputeCompositeScore(inputs, weights, thresholds)
	if err != nil {
		k.Logger(ctx).Error("composite scoring failed",
			"account", inputs.AccountAddress,
			"error", err,
		)
		return nil, types.ErrScoringFailed.Wrap(err.Error())
	}

	k.Logger(ctx).Info("composite identity score computed",
		"account", inputs.AccountAddress,
		"score", result.FinalScore,
		"version", result.ScoreVersion,
		"passed", result.Passed,
		"block_height", result.BlockHeight,
	)

	return result, nil
}

// ComputeAndStoreCompositeScore computes a composite score and stores it for the account
func (k Keeper) ComputeAndStoreCompositeScore(
	ctx sdk.Context,
	accountAddr string,
	inputs types.CompositeScoringInputs,
) (*types.CompositeScoreResult, error) {
	// Set account address in inputs
	inputs.AccountAddress = accountAddr

	// Compute the score
	result, err := k.ComputeCompositeIdentityScore(ctx, inputs)
	if err != nil {
		return nil, err
	}

	// Determine account status based on result
	status := types.AccountStatusRejected
	if result.Passed {
		status = types.AccountStatusVerified
	}

	// Store the score with version information
	err = k.SetScoreWithDetails(ctx, accountAddr, result.FinalScore, ScoreDetails{
		Status:           status,
		ModelVersion:     result.ScoreVersion,
		VerificationHash: result.InputHash,
		Reason:           k.getCompositeReasonSummary(result),
	})
	if err != nil {
		return nil, err
	}

	// Record in scoring history via evidence summary conversion
	evidenceSummary := k.convertToEvidenceSummary(result)
	if err := k.RecordScoringResult(ctx, accountAddr, evidenceSummary); err != nil {
		k.Logger(ctx).Error("failed to record composite scoring result", "error", err)
		// Non-fatal - continue
	}

	return result, nil
}

// getCompositeReasonSummary creates a brief reason summary from the composite result
func (k Keeper) getCompositeReasonSummary(result *types.CompositeScoreResult) string {
	if result.Passed {
		return "composite verification passed"
	}

	if len(result.ReasonCodes) > 0 {
		return string(result.ReasonCodes[0])
	}

	return "composite verification failed"
}

// convertToEvidenceSummary converts a CompositeScoreResult to an EvidenceSummary
// for storage in the scoring history
func (k Keeper) convertToEvidenceSummary(result *types.CompositeScoreResult) *types.EvidenceSummary {
	summary := types.NewEvidenceSummary(result.ScoreVersion, result.BlockHeight, result.ComputedAt)

	// Convert contributions
	for _, contrib := range result.Contributions {
		featureContrib := types.FeatureContribution{
			FeatureName:     contrib.ComponentName,
			RawScore:        contrib.RawScore,
			Weight:          contrib.Weight,
			WeightedScore:   contrib.WeightedScore,
			PassedThreshold: contrib.PassedThreshold,
		}
		if contrib.ReasonCode != "" {
			featureContrib.ReasonCode = types.ReasonCode(contrib.ReasonCode)
		}
		summary.AddContribution(featureContrib)
	}

	// Convert reason codes
	for _, code := range result.ReasonCodes {
		summary.AddReasonCode(types.ScoringReasonCode(code))
	}

	// Set result
	summary.SetResult(result.FinalScore, result.Passed, result.InputHash)

	// Copy component presence
	for k, v := range result.ComponentPresence {
		summary.FeaturePresence[k] = v
	}

	// Copy thresholds applied
	for k, v := range result.ThresholdsApplied {
		summary.ThresholdsApplied[k] = v
	}

	return summary
}

// ============================================================================
// Input Construction Helpers
// ============================================================================

// BuildCompositeInputsFromScopes builds CompositeScoringInputs from decrypted scopes
// and extracted features. This bridges the existing scope-based system with
// the new composite scoring algorithm.
func (k Keeper) BuildCompositeInputsFromScopes(
	ctx sdk.Context,
	accountAddress string,
	scopes []DecryptedScope,
	extractedFeatures ExtractedFeatures,
) types.CompositeScoringInputs {
	inputs := types.CompositeScoringInputs{
		AccountAddress: accountAddress,
		BlockHeight:    ctx.BlockHeight(),
		Timestamp:      ctx.BlockTime(),
	}

	// Build Document Authenticity input from document scopes
	inputs.DocumentAuthenticity = k.buildDocAuthenticityInput(scopes, extractedFeatures)

	// Build Face Match input from selfie and document scopes
	inputs.FaceMatch = k.buildFaceMatchInput(scopes, extractedFeatures)

	// Build Liveness Detection input from face video scopes
	inputs.LivenessDetection = k.buildLivenessInput(scopes, extractedFeatures)

	// Build Data Consistency input from OCR and cross-validation
	inputs.DataConsistency = k.buildDataConsistencyInput(scopes, extractedFeatures)

	// Build Historical Signals input from account history
	inputs.HistoricalSignals = k.buildHistoricalInput(ctx, accountAddress)

	// Build Risk Indicators input from security signals
	inputs.RiskIndicators = k.buildRiskInput(scopes, extractedFeatures)

	return inputs
}

// ExtractedFeatures contains features extracted from ML inference
type ExtractedFeatures struct {
	// Document features
	TamperScore           uint32
	FormatValid           bool
	TemplateMatchScore    uint32
	SecurityFeaturesScore uint32
	DocumentQuality       uint32

	// Face features
	FaceSimilarity uint32
	FaceConfidence uint32
	FaceQuality    uint32

	// Liveness features
	LivenessScore  uint32
	BlinkDetected  bool
	HeadMovement   bool
	DepthCheck     bool
	AntiSpoofScore uint32

	// OCR features
	NameMatchScore        uint32
	DOBConsistency        uint32
	AddressConsistency    uint32
	CrossFieldValidation  uint32
	AgeVerificationPassed bool
	DocumentExpiryValid   bool

	// Risk features
	FraudPatternScore      uint32
	DeviceFingerprintScore uint32
	IPReputationScore      uint32
	VelocityCheckPassed    bool
	GeoConsistencyScore    uint32
}

// buildDocAuthenticityInput builds document authenticity input from scopes
func (k Keeper) buildDocAuthenticityInput(scopes []DecryptedScope, features ExtractedFeatures) types.DocumentAuthenticityInput {
	input := types.DocumentAuthenticityInput{}

	// Check if we have an ID document scope
	for _, scope := range scopes {
		if scope.ScopeType == types.ScopeTypeIDDocument {
			input.Present = true
			break
		}
	}

	if input.Present {
		input.TamperScore = features.TamperScore
		input.FormatValidityScore = boolToScore(features.FormatValid)
		input.TemplateMatchScore = features.TemplateMatchScore
		input.SecurityFeaturesScore = features.SecurityFeaturesScore
	}

	return input
}

// buildFaceMatchInput builds face match input from scopes
func (k Keeper) buildFaceMatchInput(scopes []DecryptedScope, features ExtractedFeatures) types.FaceMatchInput {
	input := types.FaceMatchInput{}

	// Check if we have a selfie scope
	for _, scope := range scopes {
		if scope.ScopeType == types.ScopeTypeSelfie {
			input.Present = true
			break
		}
	}

	if input.Present {
		input.SimilarityScore = features.FaceSimilarity
		input.Confidence = features.FaceConfidence
		input.QualityScore = features.FaceQuality
	}

	return input
}

// buildLivenessInput builds liveness detection input from scopes
func (k Keeper) buildLivenessInput(scopes []DecryptedScope, features ExtractedFeatures) types.LivenessDetectionInput {
	input := types.LivenessDetectionInput{}

	// Check if we have a face video scope
	for _, scope := range scopes {
		if scope.ScopeType == types.ScopeTypeFaceVideo {
			input.Present = true
			break
		}
	}

	if input.Present {
		input.LivenessScore = features.LivenessScore
		input.BlinkDetected = features.BlinkDetected
		input.HeadMovementDetected = features.HeadMovement
		input.DepthCheckPassed = features.DepthCheck
		input.AntiSpoofScore = features.AntiSpoofScore
	}

	return input
}

// buildDataConsistencyInput builds data consistency input from scopes
func (k Keeper) buildDataConsistencyInput(scopes []DecryptedScope, features ExtractedFeatures) types.DataConsistencyInput {
	input := types.DataConsistencyInput{}

	// Data consistency requires document presence
	for _, scope := range scopes {
		if scope.ScopeType == types.ScopeTypeIDDocument {
			input.Present = true
			break
		}
	}

	if input.Present {
		input.NameMatchScore = features.NameMatchScore
		input.DOBConsistencyScore = features.DOBConsistency
		input.AgeVerificationPassed = features.AgeVerificationPassed
		input.AddressConsistencyScore = features.AddressConsistency
		input.DocumentExpiryValid = features.DocumentExpiryValid
		input.CrossFieldValidation = features.CrossFieldValidation
	}

	return input
}

// buildHistoricalInput builds historical signals input from account history
func (k Keeper) buildHistoricalInput(ctx sdk.Context, accountAddr string) types.HistoricalSignalsInput {
	input := types.HistoricalSignalsInput{}

	// Get account's identity record
	address, err := sdk.AccAddressFromBech32(accountAddr)
	if err != nil {
		return input
	}

	record, found := k.GetIdentityRecord(ctx, address)
	if !found {
		// New account - no historical data
		return input
	}

	input.Present = true

	// Calculate account age score (basis points)
	// Older accounts get higher scores, capped at 10000 for 1 year+
	accountAge := ctx.BlockTime().Sub(record.CreatedAt)
	accountAgeDays := accountAge.Hours() / 24
	ageScore := uint32(accountAgeDays * 27) // ~10000 for 365 days
	if ageScore > uint32(types.MaxBasisPoints) {
		ageScore = uint32(types.MaxBasisPoints)
	}
	input.AccountAgeScore = ageScore

	// Get prior verification history
	history := k.GetScoreHistory(ctx, accountAddr)
	input.VerificationHistoryCount = uint32(len(history))

	if len(history) > 0 {
		// Calculate average prior score
		var totalScore uint32
		var successCount uint32
		for _, entry := range history {
			totalScore += entry.Score
			if entry.Status == types.AccountStatusVerified {
				successCount++
			}
		}

		// Prior verification score is the average (in basis points)
		avgScore := totalScore / uint32(len(history))
		input.PriorVerificationScore = avgScore * 100 // Convert 0-100 to basis points

		// Success rate in basis points
		input.SuccessfulVerificationRate = (successCount * uint32(types.MaxBasisPoints)) / uint32(len(history))
	}

	return input
}

// buildRiskInput builds risk indicators input from security signals
func (k Keeper) buildRiskInput(scopes []DecryptedScope, features ExtractedFeatures) types.RiskIndicatorsInput {
	input := types.RiskIndicatorsInput{
		Present: true, // Risk indicators are always evaluated
	}

	input.FraudPatternScore = features.FraudPatternScore
	input.DeviceFingerprintScore = features.DeviceFingerprintScore
	input.IPReputationScore = features.IPReputationScore
	input.VelocityCheckPassed = features.VelocityCheckPassed
	input.GeoConsistencyScore = features.GeoConsistencyScore

	// Default to reasonable scores if not provided
	if input.FraudPatternScore == 0 {
		input.FraudPatternScore = uint32(types.MaxBasisPoints) // No fraud detected = max score
	}
	if input.DeviceFingerprintScore == 0 {
		input.DeviceFingerprintScore = uint32(types.MaxBasisPoints) / 2 // Unknown device = 50%
	}
	if input.IPReputationScore == 0 {
		input.IPReputationScore = uint32(types.MaxBasisPoints) / 2 // Unknown IP = 50%
	}
	if input.GeoConsistencyScore == 0 {
		input.GeoConsistencyScore = uint32(types.MaxBasisPoints) / 2 // Unknown geo = 50%
	}

	return input
}

// boolToScore converts a boolean to a basis points score
func boolToScore(b bool) uint32 {
	if b {
		return uint32(types.MaxBasisPoints)
	}
	return 0
}

// ============================================================================
// Scoring Version Management
// ============================================================================

// GetCompositeScoreVersion returns the current composite scoring algorithm version
func (k Keeper) GetCompositeScoreVersion() string {
	return types.CompositeScoreVersion
}

// IsCompositeScoreVersion checks if a version string is a composite score version
func (k Keeper) IsCompositeScoreVersion(version string) bool {
	return version == types.CompositeScoreVersion ||
		len(version) > len(types.CompositeScoreVersionPrefix) &&
			version[:len(types.CompositeScoreVersionPrefix)] == types.CompositeScoreVersionPrefix
}
