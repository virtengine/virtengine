// Package types provides types for the VEID module.
//
// VEID-CORE-002: Identity Scope Scoring Algorithm
// This file implements the spec-compliant composite scoring algorithm per veid-flow-spec.md.
// Weights:
//   - Document Authenticity: 25%
//   - Face Match: 25%
//   - Liveness Detection: 20%
//   - Data Consistency: 15%
//   - Historical Signals: 10%
//   - Risk Indicators: 5%
package types

import (
	"crypto/sha256"
	"sort"
	"time"
)

// ============================================================================
// Spec-Compliant Composite Scoring Constants (veid-flow-spec.md)
// ============================================================================

// Composite scoring weight constants in basis points (10000 = 100%)
// Per veid-flow-spec.md ML Score Calculation section
const (
	// WeightDocumentAuthenticity is the weight for document authenticity (25%)
	WeightDocumentAuthenticity uint32 = 2500

	// WeightFaceMatch is the weight for face match confidence (25%)
	WeightFaceMatch uint32 = 2500

	// WeightLivenessDetection is the weight for liveness detection (20%)
	WeightLivenessDetection uint32 = 2000

	// WeightDataConsistency is the weight for data consistency (15%)
	WeightDataConsistency uint32 = 1500

	// WeightHistoricalSignals is the weight for historical signals (10%)
	WeightHistoricalSignals uint32 = 1000

	// WeightRiskIndicators is the weight for risk indicators (5%)
	WeightRiskIndicators uint32 = 500

	// TotalCompositeWeight is the sum of all weights (must equal 10000)
	TotalCompositeWeight uint32 = WeightDocumentAuthenticity + WeightFaceMatch +
		WeightLivenessDetection + WeightDataConsistency +
		WeightHistoricalSignals + WeightRiskIndicators
)

// Scoring algorithm version constants
const (
	// CompositeScoreVersion is the version of the composite scoring algorithm
	CompositeScoreVersion = "2.0.0-composite"

	// CompositeScoreVersionPrefix is the prefix for composite score versions
	CompositeScoreVersionPrefix = "composite-"
)

// ============================================================================
// Composite Scoring Component Names
// ============================================================================

// Component name constants for composite scoring contributions
const (
	ComponentDocumentAuthenticity = "document_authenticity"
	ComponentFaceMatch            = "face_match"
	ComponentLivenessDetection    = "liveness_detection"
	ComponentDataConsistency      = "data_consistency"
	ComponentHistoricalSignals    = "historical_signals"
	ComponentRiskIndicators       = "risk_indicators"
)

// AllCompositeComponentNames returns all component names in consistent order
func AllCompositeComponentNames() []string {
	return []string{
		ComponentDocumentAuthenticity,
		ComponentFaceMatch,
		ComponentLivenessDetection,
		ComponentDataConsistency,
		ComponentHistoricalSignals,
		ComponentRiskIndicators,
	}
}

// ============================================================================
// Composite Scoring Weights (Spec-Compliant)
// ============================================================================

// CompositeScoringWeights defines the spec-compliant weights per veid-flow-spec.md
type CompositeScoringWeights struct {
	// DocumentAuthenticity weight (25%) - Tamper detection, format validity
	DocumentAuthenticity uint32 `json:"document_authenticity"`

	// FaceMatch weight (25%) - ID photo ↔ selfie match score
	FaceMatch uint32 `json:"face_match"`

	// LivenessDetection weight (20%) - Anti-spoof measures (blink, motion)
	LivenessDetection uint32 `json:"liveness_detection"`

	// DataConsistency weight (15%) - Cross-field validation, age checks
	DataConsistency uint32 `json:"data_consistency"`

	// HistoricalSignals weight (10%) - Prior verifications, account age
	HistoricalSignals uint32 `json:"historical_signals"`

	// RiskIndicators weight (5%) - Known fraud patterns, device fingerprint
	RiskIndicators uint32 `json:"risk_indicators"`
}

// DefaultCompositeScoringWeights returns the spec-compliant weights from veid-flow-spec.md
func DefaultCompositeScoringWeights() CompositeScoringWeights {
	return CompositeScoringWeights{
		DocumentAuthenticity: WeightDocumentAuthenticity, // 25%
		FaceMatch:            WeightFaceMatch,            // 25%
		LivenessDetection:    WeightLivenessDetection,    // 20%
		DataConsistency:      WeightDataConsistency,      // 15%
		HistoricalSignals:    WeightHistoricalSignals,    // 10%
		RiskIndicators:       WeightRiskIndicators,       // 5%
	}
}

// Validate ensures the weights sum to exactly 10000 basis points
func (w CompositeScoringWeights) Validate() error {
	total := w.TotalWeight()
	if total != uint32(MaxBasisPoints) {
		return ErrInvalidScoringModel.Wrapf("composite weights must sum to %d, got %d", MaxBasisPoints, total)
	}
	return nil
}

// TotalWeight returns the sum of all weights
func (w CompositeScoringWeights) TotalWeight() uint32 {
	return w.DocumentAuthenticity + w.FaceMatch + w.LivenessDetection +
		w.DataConsistency + w.HistoricalSignals + w.RiskIndicators
}

// ============================================================================
// Composite Scoring Thresholds
// ============================================================================

// CompositeScoringThresholds defines thresholds for composite scoring
type CompositeScoringThresholds struct {
	// MinDocumentAuthenticity is the minimum document authenticity score (basis points)
	MinDocumentAuthenticity uint32 `json:"min_document_authenticity"`

	// MinFaceMatch is the minimum face match score (basis points)
	MinFaceMatch uint32 `json:"min_face_match"`

	// MinLivenessDetection is the minimum liveness score (basis points)
	MinLivenessDetection uint32 `json:"min_liveness_detection"`

	// MinDataConsistency is the minimum data consistency score (basis points)
	MinDataConsistency uint32 `json:"min_data_consistency"`

	// MinHistoricalSignals is the minimum historical signals score (basis points)
	MinHistoricalSignals uint32 `json:"min_historical_signals"`

	// MinRiskIndicators is the minimum acceptable risk score (basis points, lower is riskier)
	MinRiskIndicators uint32 `json:"min_risk_indicators"`

	// RequiredForPass is the minimum final score to pass verification (0-100)
	RequiredForPass uint32 `json:"required_for_pass"`
}

// DefaultCompositeScoringThresholds returns default thresholds for composite scoring
func DefaultCompositeScoringThresholds() CompositeScoringThresholds {
	return CompositeScoringThresholds{
		MinDocumentAuthenticity: 5000, // 50%
		MinFaceMatch:            7000, // 70%
		MinLivenessDetection:    7500, // 75%
		MinDataConsistency:      6000, // 60%
		MinHistoricalSignals:    0,    // No minimum - new accounts ok
		MinRiskIndicators:       5000, // 50% (lower risk score = more risky)
		RequiredForPass:         50,   // Minimum score to pass
	}
}

// Validate ensures thresholds are within valid ranges
func (t CompositeScoringThresholds) Validate() error {
	if t.MinDocumentAuthenticity > uint32(MaxBasisPoints) {
		return ErrInvalidScoringModel.Wrap("min_document_authenticity exceeds maximum")
	}
	if t.MinFaceMatch > uint32(MaxBasisPoints) {
		return ErrInvalidScoringModel.Wrap("min_face_match exceeds maximum")
	}
	if t.MinLivenessDetection > uint32(MaxBasisPoints) {
		return ErrInvalidScoringModel.Wrap("min_liveness_detection exceeds maximum")
	}
	if t.MinDataConsistency > uint32(MaxBasisPoints) {
		return ErrInvalidScoringModel.Wrap("min_data_consistency exceeds maximum")
	}
	if t.MinHistoricalSignals > uint32(MaxBasisPoints) {
		return ErrInvalidScoringModel.Wrap("min_historical_signals exceeds maximum")
	}
	if t.MinRiskIndicators > uint32(MaxBasisPoints) {
		return ErrInvalidScoringModel.Wrap("min_risk_indicators exceeds maximum")
	}
	if t.RequiredForPass > MaxScore {
		return ErrInvalidScoringModel.Wrap("required_for_pass exceeds maximum score")
	}
	return nil
}

// ============================================================================
// Composite Scoring Inputs
// ============================================================================

// DocumentAuthenticityInput represents document authenticity verification input
type DocumentAuthenticityInput struct {
	// TamperScore is the tamper detection score (basis points, higher = less tampered)
	TamperScore uint32 `json:"tamper_score"`

	// FormatValidityScore is the format validity score (basis points)
	FormatValidityScore uint32 `json:"format_validity_score"`

	// TemplateMatchScore is the template matching score (basis points)
	TemplateMatchScore uint32 `json:"template_match_score"`

	// SecurityFeaturesScore is the security features detection score (basis points)
	SecurityFeaturesScore uint32 `json:"security_features_score"`

	// Present indicates if document authenticity data was available
	Present bool `json:"present"`
}

// ComputeScore computes the overall document authenticity score
func (d DocumentAuthenticityInput) ComputeScore() uint32 {
	if !d.Present {
		return 0
	}

	// Weighted average: tamper (35%), format (25%), template (25%), security (15%)
	score := uint64(d.TamperScore)*35 +
		uint64(d.FormatValidityScore)*25 +
		uint64(d.TemplateMatchScore)*25 +
		uint64(d.SecurityFeaturesScore)*15

	score /= 100

	if score > uint64(MaxBasisPoints) {
		return uint32(MaxBasisPoints)
	}
	return uint32(score)
}

// FaceMatchInput represents face match verification input
type FaceMatchInput struct {
	// SimilarityScore is the face similarity score (basis points)
	SimilarityScore uint32 `json:"similarity_score"`

	// Confidence is the model confidence (basis points)
	Confidence uint32 `json:"confidence"`

	// QualityScore is the face image quality score (basis points)
	QualityScore uint32 `json:"quality_score"`

	// Present indicates if face match data was available
	Present bool `json:"present"`
}

// ComputeScore computes the overall face match score
func (f FaceMatchInput) ComputeScore() uint32 {
	if !f.Present {
		return 0
	}

	// Similarity weighted by confidence and quality
	// Effective = (similarity * confidence * quality_factor) / 10000^2
	qualityFactor := (f.QualityScore + uint32(MaxBasisPoints)) / 2 // Quality boosts score
	effectiveScore := (uint64(f.SimilarityScore) * uint64(f.Confidence)) / uint64(MaxBasisPoints)
	effectiveScore = (effectiveScore * uint64(qualityFactor)) / uint64(MaxBasisPoints)

	if effectiveScore > uint64(MaxBasisPoints) {
		return uint32(MaxBasisPoints)
	}
	return uint32(effectiveScore)
}

// LivenessDetectionInput represents liveness detection input
type LivenessDetectionInput struct {
	// LivenessScore is the primary liveness detection score (basis points)
	LivenessScore uint32 `json:"liveness_score"`

	// BlinkDetected indicates if blink was detected
	BlinkDetected bool `json:"blink_detected"`

	// HeadMovementDetected indicates if head movement was detected
	HeadMovementDetected bool `json:"head_movement_detected"`

	// DepthCheckPassed indicates if 3D depth check passed
	DepthCheckPassed bool `json:"depth_check_passed"`

	// AntiSpoofScore is the anti-spoofing detection score (basis points)
	AntiSpoofScore uint32 `json:"anti_spoof_score"`

	// Present indicates if liveness data was available
	Present bool `json:"present"`
}

// ComputeScore computes the overall liveness detection score
func (l LivenessDetectionInput) ComputeScore() uint32 {
	if !l.Present {
		return uint32(MaxBasisPoints) / 2 // Default 50% if not present
	}

	// Base score from liveness detection
	score := uint64(l.LivenessScore)

	// Bonuses for additional liveness signals
	if l.BlinkDetected {
		score = (score * 103) / 100 // 3% bonus
	}
	if l.HeadMovementDetected {
		score = (score * 103) / 100 // 3% bonus
	}
	if l.DepthCheckPassed {
		score = (score * 105) / 100 // 5% bonus
	}

	// Factor in anti-spoof score (30% weight)
	combinedScore := (score*70 + uint64(l.AntiSpoofScore)*30) / 100

	if combinedScore > uint64(MaxBasisPoints) {
		return uint32(MaxBasisPoints)
	}
	return uint32(combinedScore)
}

// DataConsistencyInput represents cross-field data consistency input
type DataConsistencyInput struct {
	// NameMatchScore is the name consistency score (basis points)
	NameMatchScore uint32 `json:"name_match_score"`

	// DOBConsistencyScore is the date of birth consistency score (basis points)
	DOBConsistencyScore uint32 `json:"dob_consistency_score"`

	// AgeVerificationPassed indicates if age verification passed
	AgeVerificationPassed bool `json:"age_verification_passed"`

	// AddressConsistencyScore is the address consistency score (basis points)
	AddressConsistencyScore uint32 `json:"address_consistency_score"`

	// DocumentExpiryValid indicates if document is not expired
	DocumentExpiryValid bool `json:"document_expiry_valid"`

	// CrossFieldValidation is the overall cross-field validation score (basis points)
	CrossFieldValidation uint32 `json:"cross_field_validation"`

	// Present indicates if data consistency data was available
	Present bool `json:"present"`
}

// ComputeScore computes the overall data consistency score
func (d DataConsistencyInput) ComputeScore() uint32 {
	if !d.Present {
		return 0
	}

	// Weighted average of consistency checks
	score := uint64(d.NameMatchScore)*25 +
		uint64(d.DOBConsistencyScore)*20 +
		uint64(d.AddressConsistencyScore)*15 +
		uint64(d.CrossFieldValidation)*40

	score /= 100

	// Penalties for failed checks
	if !d.AgeVerificationPassed {
		score = (score * 70) / 100 // 30% penalty
	}
	if !d.DocumentExpiryValid {
		score = (score * 80) / 100 // 20% penalty
	}

	if score > uint64(MaxBasisPoints) {
		return uint32(MaxBasisPoints)
	}
	return uint32(score)
}

// HistoricalSignalsInput represents historical signals input
type HistoricalSignalsInput struct {
	// PriorVerificationScore is the score from prior verifications (basis points)
	PriorVerificationScore uint32 `json:"prior_verification_score"`

	// AccountAgeScore is the account age contribution (basis points)
	// Older accounts get higher scores
	AccountAgeScore uint32 `json:"account_age_score"`

	// VerificationHistoryCount is the number of prior verifications
	VerificationHistoryCount uint32 `json:"verification_history_count"`

	// SuccessfulVerificationRate is the historical success rate (basis points)
	SuccessfulVerificationRate uint32 `json:"successful_verification_rate"`

	// Present indicates if historical data was available
	Present bool `json:"present"`
}

// ComputeScore computes the overall historical signals score
func (h HistoricalSignalsInput) ComputeScore() uint32 {
	if !h.Present {
		// New accounts get a neutral score
		return uint32(MaxBasisPoints) / 2 // 50%
	}

	// If no prior verifications, use account age only
	if h.VerificationHistoryCount == 0 {
		return h.AccountAgeScore
	}

	// Weighted combination of historical signals
	score := uint64(h.PriorVerificationScore)*40 +
		uint64(h.AccountAgeScore)*20 +
		uint64(h.SuccessfulVerificationRate)*40

	score /= 100

	if score > uint64(MaxBasisPoints) {
		return uint32(MaxBasisPoints)
	}
	return uint32(score)
}

// RiskIndicatorsInput represents risk indicators input
type RiskIndicatorsInput struct {
	// FraudPatternScore is the inverse fraud pattern match score (basis points)
	// Higher score = less fraud patterns detected
	FraudPatternScore uint32 `json:"fraud_pattern_score"`

	// DeviceFingerprintScore is the device trust score (basis points)
	DeviceFingerprintScore uint32 `json:"device_fingerprint_score"`

	// IPReputationScore is the IP address reputation score (basis points)
	IPReputationScore uint32 `json:"ip_reputation_score"`

	// VelocityCheckPassed indicates if velocity checks passed
	VelocityCheckPassed bool `json:"velocity_check_passed"`

	// GeoConsistencyScore is the geographic consistency score (basis points)
	GeoConsistencyScore uint32 `json:"geo_consistency_score"`

	// Present indicates if risk data was available
	Present bool `json:"present"`
}

// ComputeScore computes the overall risk indicators score
// Higher score = lower risk
func (r RiskIndicatorsInput) ComputeScore() uint32 {
	if !r.Present {
		// Default moderate risk for missing data
		return uint32(MaxBasisPoints) / 2 // 50%
	}

	// Weighted average of risk indicators
	score := uint64(r.FraudPatternScore)*30 +
		uint64(r.DeviceFingerprintScore)*25 +
		uint64(r.IPReputationScore)*20 +
		uint64(r.GeoConsistencyScore)*25

	score /= 100

	// Penalty if velocity check failed
	if !r.VelocityCheckPassed {
		score = (score * 60) / 100 // 40% penalty
	}

	if score > uint64(MaxBasisPoints) {
		return uint32(MaxBasisPoints)
	}
	return uint32(score)
}

// ============================================================================
// Composite Scoring Inputs
// ============================================================================

// CompositeScoringInputs aggregates all inputs for spec-compliant composite scoring
type CompositeScoringInputs struct {
	// DocumentAuthenticity contains document authenticity verification input
	DocumentAuthenticity DocumentAuthenticityInput `json:"document_authenticity"`

	// FaceMatch contains face match verification input
	FaceMatch FaceMatchInput `json:"face_match"`

	// LivenessDetection contains liveness detection input
	LivenessDetection LivenessDetectionInput `json:"liveness_detection"`

	// DataConsistency contains data consistency input
	DataConsistency DataConsistencyInput `json:"data_consistency"`

	// HistoricalSignals contains historical signals input
	HistoricalSignals HistoricalSignalsInput `json:"historical_signals"`

	// RiskIndicators contains risk indicators input
	RiskIndicators RiskIndicatorsInput `json:"risk_indicators"`

	// AccountAddress is the account being scored
	AccountAddress string `json:"account_address"`

	// BlockHeight is the block height when scoring
	BlockHeight int64 `json:"block_height"`

	// Timestamp is when scoring was performed
	Timestamp time.Time `json:"timestamp"`
}

// Validate validates all composite scoring inputs
func (c CompositeScoringInputs) Validate() error {
	if c.AccountAddress == "" {
		return ErrInvalidScoringInput.Wrap("account_address cannot be empty")
	}
	return nil
}

// ComputeInputHash computes a deterministic hash of all inputs for consensus
func (c CompositeScoringInputs) ComputeInputHash() []byte {
	h := sha256.New()

	// Account and block context
	h.Write([]byte(c.AccountAddress))
	h.Write(encodeInt64(c.BlockHeight))
	h.Write(encodeInt64(c.Timestamp.Unix()))

	// Document Authenticity inputs
	h.Write(encodeUint32(c.DocumentAuthenticity.TamperScore))
	h.Write(encodeUint32(c.DocumentAuthenticity.FormatValidityScore))
	h.Write(encodeUint32(c.DocumentAuthenticity.TemplateMatchScore))
	h.Write(encodeUint32(c.DocumentAuthenticity.SecurityFeaturesScore))
	h.Write(encodeBool(c.DocumentAuthenticity.Present))

	// Face Match inputs
	h.Write(encodeUint32(c.FaceMatch.SimilarityScore))
	h.Write(encodeUint32(c.FaceMatch.Confidence))
	h.Write(encodeUint32(c.FaceMatch.QualityScore))
	h.Write(encodeBool(c.FaceMatch.Present))

	// Liveness Detection inputs
	h.Write(encodeUint32(c.LivenessDetection.LivenessScore))
	h.Write(encodeBool(c.LivenessDetection.BlinkDetected))
	h.Write(encodeBool(c.LivenessDetection.HeadMovementDetected))
	h.Write(encodeBool(c.LivenessDetection.DepthCheckPassed))
	h.Write(encodeUint32(c.LivenessDetection.AntiSpoofScore))
	h.Write(encodeBool(c.LivenessDetection.Present))

	// Data Consistency inputs
	h.Write(encodeUint32(c.DataConsistency.NameMatchScore))
	h.Write(encodeUint32(c.DataConsistency.DOBConsistencyScore))
	h.Write(encodeBool(c.DataConsistency.AgeVerificationPassed))
	h.Write(encodeUint32(c.DataConsistency.AddressConsistencyScore))
	h.Write(encodeBool(c.DataConsistency.DocumentExpiryValid))
	h.Write(encodeUint32(c.DataConsistency.CrossFieldValidation))
	h.Write(encodeBool(c.DataConsistency.Present))

	// Historical Signals inputs
	h.Write(encodeUint32(c.HistoricalSignals.PriorVerificationScore))
	h.Write(encodeUint32(c.HistoricalSignals.AccountAgeScore))
	h.Write(encodeUint32(c.HistoricalSignals.VerificationHistoryCount))
	h.Write(encodeUint32(c.HistoricalSignals.SuccessfulVerificationRate))
	h.Write(encodeBool(c.HistoricalSignals.Present))

	// Risk Indicators inputs
	h.Write(encodeUint32(c.RiskIndicators.FraudPatternScore))
	h.Write(encodeUint32(c.RiskIndicators.DeviceFingerprintScore))
	h.Write(encodeUint32(c.RiskIndicators.IPReputationScore))
	h.Write(encodeBool(c.RiskIndicators.VelocityCheckPassed))
	h.Write(encodeUint32(c.RiskIndicators.GeoConsistencyScore))
	h.Write(encodeBool(c.RiskIndicators.Present))

	return h.Sum(nil)
}

// ============================================================================
// Composite Scoring Reason Codes
// ============================================================================

// CompositeReasonCode represents a detailed reason for composite scoring outcome
type CompositeReasonCode string

const (
	// CompositeReasonSuccess indicates successful composite scoring
	CompositeReasonSuccess CompositeReasonCode = "COMPOSITE_SUCCESS"

	// CompositeReasonMissingDocument indicates document was not provided
	CompositeReasonMissingDocument CompositeReasonCode = "MISSING_DOCUMENT"

	// CompositeReasonMissingSelfie indicates selfie/face was not provided
	CompositeReasonMissingSelfie CompositeReasonCode = "MISSING_SELFIE"

	// CompositeReasonLowDocumentAuthenticity indicates document authenticity below threshold
	CompositeReasonLowDocumentAuthenticity CompositeReasonCode = "LOW_DOCUMENT_AUTHENTICITY"

	// CompositeReasonLowFaceMatch indicates face match below threshold
	CompositeReasonLowFaceMatch CompositeReasonCode = "LOW_FACE_MATCH"

	// CompositeReasonLowLiveness indicates liveness detection below threshold
	CompositeReasonLowLiveness CompositeReasonCode = "LOW_LIVENESS"

	// CompositeReasonLowDataConsistency indicates data consistency below threshold
	CompositeReasonLowDataConsistency CompositeReasonCode = "LOW_DATA_CONSISTENCY"

	// CompositeReasonLowHistoricalSignals indicates historical signals below threshold
	CompositeReasonLowHistoricalSignals CompositeReasonCode = "LOW_HISTORICAL_SIGNALS"

	// CompositeReasonHighRisk indicates risk indicators above acceptable threshold
	CompositeReasonHighRisk CompositeReasonCode = "HIGH_RISK"

	// CompositeReasonBelowPassThreshold indicates final score below pass threshold
	CompositeReasonBelowPassThreshold CompositeReasonCode = "BELOW_PASS_THRESHOLD"
)

// ============================================================================
// Composite Score Result
// ============================================================================

// CompositeScoreContribution represents a single component's contribution
type CompositeScoreContribution struct {
	// ComponentName is the name of the component
	ComponentName string `json:"component_name"`

	// RawScore is the component's raw score in basis points (0-10000)
	RawScore uint32 `json:"raw_score"`

	// Weight is the component's weight in basis points
	Weight uint32 `json:"weight"`

	// WeightedScore is (RawScore * Weight) / 10000
	WeightedScore uint32 `json:"weighted_score"`

	// PassedThreshold indicates if the component passed its threshold
	PassedThreshold bool `json:"passed_threshold"`

	// ReasonCode is the reason code if component failed
	ReasonCode CompositeReasonCode `json:"reason_code,omitempty"`
}

// CompositeScoreResult contains the complete composite scoring result
type CompositeScoreResult struct {
	// FinalScore is the computed identity score (0-100)
	FinalScore uint32 `json:"final_score"`

	// Passed indicates if the verification passed
	Passed bool `json:"passed"`

	// ScoreVersion is the scoring algorithm version
	ScoreVersion string `json:"score_version"`

	// Contributions contains each component's contribution to the score
	Contributions []CompositeScoreContribution `json:"contributions"`

	// ReasonCodes contains codes explaining the scoring outcome
	ReasonCodes []CompositeReasonCode `json:"reason_codes"`

	// InputHash is the hash of all inputs for verification
	InputHash []byte `json:"input_hash"`

	// ComputedAt is when the score was computed
	ComputedAt time.Time `json:"computed_at"`

	// BlockHeight is the block height when computed
	BlockHeight int64 `json:"block_height"`

	// ComponentPresence tracks which components were present
	ComponentPresence map[string]bool `json:"component_presence"`

	// ThresholdsApplied contains threshold values that were applied
	ThresholdsApplied map[string]uint32 `json:"thresholds_applied,omitempty"`
}

// NewCompositeScoreResult creates a new composite score result
func NewCompositeScoreResult(blockHeight int64, computedAt time.Time) *CompositeScoreResult {
	return &CompositeScoreResult{
		ScoreVersion:      CompositeScoreVersion,
		BlockHeight:       blockHeight,
		ComputedAt:        computedAt,
		Contributions:     make([]CompositeScoreContribution, 0, 6),
		ReasonCodes:       make([]CompositeReasonCode, 0),
		ComponentPresence: make(map[string]bool),
		ThresholdsApplied: make(map[string]uint32),
	}
}

// AddContribution adds a component contribution to the result
func (r *CompositeScoreResult) AddContribution(contrib CompositeScoreContribution) {
	r.Contributions = append(r.Contributions, contrib)
	r.ComponentPresence[contrib.ComponentName] = contrib.RawScore > 0
}

// AddReasonCode adds a reason code to the result, avoiding duplicates
func (r *CompositeScoreResult) AddReasonCode(code CompositeReasonCode) {
	for _, existing := range r.ReasonCodes {
		if existing == code {
			return
		}
	}
	r.ReasonCodes = append(r.ReasonCodes, code)
}

// SetResult sets the final result
func (r *CompositeScoreResult) SetResult(score uint32, passed bool, inputHash []byte) {
	r.FinalScore = score
	r.Passed = passed
	r.InputHash = inputHash

	if passed {
		r.AddReasonCode(CompositeReasonSuccess)
	}
}

// ============================================================================
// Deterministic Composite Score Computation
// ============================================================================

// ComputeCompositeScore computes the identity score using the spec-compliant
// composite scoring algorithm with fixed-point arithmetic for determinism.
func ComputeCompositeScore(
	inputs CompositeScoringInputs,
	weights CompositeScoringWeights,
	thresholds CompositeScoringThresholds,
) (*CompositeScoreResult, error) {
	// Validate inputs
	if err := inputs.Validate(); err != nil {
		return nil, err
	}

	// Validate weights
	if err := weights.Validate(); err != nil {
		return nil, err
	}

	// Validate thresholds
	if err := thresholds.Validate(); err != nil {
		return nil, err
	}

	// Create result
	result := NewCompositeScoreResult(inputs.BlockHeight, inputs.Timestamp)

	// Track reason codes
	var reasonCodes []CompositeReasonCode

	// Check for required components
	if !inputs.DocumentAuthenticity.Present {
		reasonCodes = append(reasonCodes, CompositeReasonMissingDocument)
	}
	if !inputs.FaceMatch.Present {
		reasonCodes = append(reasonCodes, CompositeReasonMissingSelfie)
	}

	// Compute individual component contributions

	// 1. Document Authenticity (25%)
	docContrib := computeDocAuthenticityContribution(
		inputs.DocumentAuthenticity,
		weights.DocumentAuthenticity,
		thresholds.MinDocumentAuthenticity,
	)
	result.AddContribution(docContrib)
	if !docContrib.PassedThreshold && inputs.DocumentAuthenticity.Present {
		reasonCodes = append(reasonCodes, CompositeReasonLowDocumentAuthenticity)
	}

	// 2. Face Match (25%)
	faceContrib := computeFaceMatchContribution(
		inputs.FaceMatch,
		weights.FaceMatch,
		thresholds.MinFaceMatch,
	)
	result.AddContribution(faceContrib)
	if !faceContrib.PassedThreshold && inputs.FaceMatch.Present {
		reasonCodes = append(reasonCodes, CompositeReasonLowFaceMatch)
	}

	// 3. Liveness Detection (20%)
	livenessContrib := computeLivenessContribution2(
		inputs.LivenessDetection,
		weights.LivenessDetection,
		thresholds.MinLivenessDetection,
	)
	result.AddContribution(livenessContrib)
	if !livenessContrib.PassedThreshold && inputs.LivenessDetection.Present {
		reasonCodes = append(reasonCodes, CompositeReasonLowLiveness)
	}

	// 4. Data Consistency (15%)
	consistencyContrib := computeDataConsistencyContribution(
		inputs.DataConsistency,
		weights.DataConsistency,
		thresholds.MinDataConsistency,
	)
	result.AddContribution(consistencyContrib)
	if !consistencyContrib.PassedThreshold && inputs.DataConsistency.Present {
		reasonCodes = append(reasonCodes, CompositeReasonLowDataConsistency)
	}

	// 5. Historical Signals (10%)
	historyContrib := computeHistoricalContribution(
		inputs.HistoricalSignals,
		weights.HistoricalSignals,
		thresholds.MinHistoricalSignals,
	)
	result.AddContribution(historyContrib)
	if !historyContrib.PassedThreshold && inputs.HistoricalSignals.Present {
		reasonCodes = append(reasonCodes, CompositeReasonLowHistoricalSignals)
	}

	// 6. Risk Indicators (5%)
	riskContrib := computeRiskContribution(
		inputs.RiskIndicators,
		weights.RiskIndicators,
		thresholds.MinRiskIndicators,
	)
	result.AddContribution(riskContrib)
	if !riskContrib.PassedThreshold && inputs.RiskIndicators.Present {
		reasonCodes = append(reasonCodes, CompositeReasonHighRisk)
	}

	// Compute total weighted score using fixed-point arithmetic
	var totalWeightedScore uint64
	for _, c := range result.Contributions {
		totalWeightedScore += uint64(c.WeightedScore)
	}

	// Convert from basis points to 0-100 scale
	// totalWeightedScore is the sum of (component_score * weight / 10000)
	// We need to divide by total possible weight contribution to normalize
	finalScore := uint32((totalWeightedScore * 100) / uint64(MaxBasisPoints))

	// Ensure score doesn't exceed maximum
	if finalScore > MaxScore {
		finalScore = MaxScore
	}

	// Check if passed
	passed := finalScore >= thresholds.RequiredForPass
	if !passed {
		reasonCodes = append(reasonCodes, CompositeReasonBelowPassThreshold)
	}

	// Set thresholds applied for transparency
	result.ThresholdsApplied = map[string]uint32{
		"min_document_authenticity": thresholds.MinDocumentAuthenticity,
		"min_face_match":            thresholds.MinFaceMatch,
		"min_liveness_detection":    thresholds.MinLivenessDetection,
		"min_data_consistency":      thresholds.MinDataConsistency,
		"min_historical_signals":    thresholds.MinHistoricalSignals,
		"min_risk_indicators":       thresholds.MinRiskIndicators,
		"required_for_pass":         thresholds.RequiredForPass,
	}

	// Deduplicate and sort reason codes for determinism
	uniqueReasons := deduplicateCompositeReasonCodes(reasonCodes)
	sort.Slice(uniqueReasons, func(i, j int) bool {
		return string(uniqueReasons[i]) < string(uniqueReasons[j])
	})
	result.ReasonCodes = uniqueReasons

	// Compute input hash and set result
	result.SetResult(finalScore, passed, inputs.ComputeInputHash())

	return result, nil
}

// ============================================================================
// Helper Functions for Component Contributions
// ============================================================================

func computeDocAuthenticityContribution(
	input DocumentAuthenticityInput,
	weight uint32,
	threshold uint32,
) CompositeScoreContribution {
	contrib := CompositeScoreContribution{
		ComponentName: ComponentDocumentAuthenticity,
		Weight:        weight,
	}

	if !input.Present {
		contrib.RawScore = 0
		contrib.WeightedScore = 0
		contrib.PassedThreshold = false
		contrib.ReasonCode = CompositeReasonMissingDocument
		return contrib
	}

	contrib.RawScore = input.ComputeScore()
	contrib.WeightedScore = uint32((uint64(contrib.RawScore) * uint64(weight)) / uint64(MaxBasisPoints))
	contrib.PassedThreshold = contrib.RawScore >= threshold

	if !contrib.PassedThreshold {
		contrib.ReasonCode = CompositeReasonLowDocumentAuthenticity
	}

	return contrib
}

func computeFaceMatchContribution(
	input FaceMatchInput,
	weight uint32,
	threshold uint32,
) CompositeScoreContribution {
	contrib := CompositeScoreContribution{
		ComponentName: ComponentFaceMatch,
		Weight:        weight,
	}

	if !input.Present {
		contrib.RawScore = 0
		contrib.WeightedScore = 0
		contrib.PassedThreshold = false
		contrib.ReasonCode = CompositeReasonMissingSelfie
		return contrib
	}

	contrib.RawScore = input.ComputeScore()
	contrib.WeightedScore = uint32((uint64(contrib.RawScore) * uint64(weight)) / uint64(MaxBasisPoints))
	contrib.PassedThreshold = contrib.RawScore >= threshold

	if !contrib.PassedThreshold {
		contrib.ReasonCode = CompositeReasonLowFaceMatch
	}

	return contrib
}

func computeLivenessContribution2(
	input LivenessDetectionInput,
	weight uint32,
	threshold uint32,
) CompositeScoreContribution {
	contrib := CompositeScoreContribution{
		ComponentName: ComponentLivenessDetection,
		Weight:        weight,
	}

	// Liveness is optional - get default score if not present
	contrib.RawScore = input.ComputeScore()
	contrib.WeightedScore = uint32((uint64(contrib.RawScore) * uint64(weight)) / uint64(MaxBasisPoints))

	// If present, check threshold; if not present, automatically pass
	if input.Present {
		contrib.PassedThreshold = contrib.RawScore >= threshold
		if !contrib.PassedThreshold {
			contrib.ReasonCode = CompositeReasonLowLiveness
		}
	} else {
		contrib.PassedThreshold = true // Optional component
	}

	return contrib
}

func computeDataConsistencyContribution(
	input DataConsistencyInput,
	weight uint32,
	threshold uint32,
) CompositeScoreContribution {
	contrib := CompositeScoreContribution{
		ComponentName: ComponentDataConsistency,
		Weight:        weight,
	}

	if !input.Present {
		contrib.RawScore = 0
		contrib.WeightedScore = 0
		contrib.PassedThreshold = false
		contrib.ReasonCode = CompositeReasonLowDataConsistency
		return contrib
	}

	contrib.RawScore = input.ComputeScore()
	contrib.WeightedScore = uint32((uint64(contrib.RawScore) * uint64(weight)) / uint64(MaxBasisPoints))
	contrib.PassedThreshold = contrib.RawScore >= threshold

	if !contrib.PassedThreshold {
		contrib.ReasonCode = CompositeReasonLowDataConsistency
	}

	return contrib
}

func computeHistoricalContribution(
	input HistoricalSignalsInput,
	weight uint32,
	threshold uint32,
) CompositeScoreContribution {
	contrib := CompositeScoreContribution{
		ComponentName: ComponentHistoricalSignals,
		Weight:        weight,
	}

	// Historical signals are optional - get default score
	contrib.RawScore = input.ComputeScore()
	contrib.WeightedScore = uint32((uint64(contrib.RawScore) * uint64(weight)) / uint64(MaxBasisPoints))

	// Historical threshold is typically 0 for new accounts
	contrib.PassedThreshold = contrib.RawScore >= threshold

	if !contrib.PassedThreshold {
		contrib.ReasonCode = CompositeReasonLowHistoricalSignals
	}

	return contrib
}

func computeRiskContribution(
	input RiskIndicatorsInput,
	weight uint32,
	threshold uint32,
) CompositeScoreContribution {
	contrib := CompositeScoreContribution{
		ComponentName: ComponentRiskIndicators,
		Weight:        weight,
	}

	// Risk indicators are optional - get default score
	contrib.RawScore = input.ComputeScore()
	contrib.WeightedScore = uint32((uint64(contrib.RawScore) * uint64(weight)) / uint64(MaxBasisPoints))

	// For risk, higher score = lower risk, so check if above threshold
	contrib.PassedThreshold = contrib.RawScore >= threshold

	if !contrib.PassedThreshold {
		contrib.ReasonCode = CompositeReasonHighRisk
	}

	return contrib
}

func deduplicateCompositeReasonCodes(codes []CompositeReasonCode) []CompositeReasonCode {
	seen := make(map[CompositeReasonCode]bool)
	result := make([]CompositeReasonCode, 0, len(codes))
	for _, code := range codes {
		if !seen[code] {
			seen[code] = true
			result = append(result, code)
		}
	}
	return result
}
