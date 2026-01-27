// Package types provides types for the VEID module.
//
// VE-220: VEID scoring model v1 - feature fusion from doc OCR + face match + metadata
// This file defines the deterministic scoring model for identity verification.
package types

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
	"time"
)

// ============================================================================
// Fixed-Point Arithmetic Constants
// ============================================================================
// Use fixed-point arithmetic for deterministic consensus computation.
// All internal calculations use basis points (1/10000) for precision.

const (
	// BasisPointsScale is the multiplier for fixed-point arithmetic (10000 = 100.00%)
	BasisPointsScale uint64 = 10000

	// MaxBasisPoints is the maximum value (100% = 10000 basis points)
	MaxBasisPoints uint64 = 10000

	// ScorePrecision is the precision for score calculations
	ScorePrecision uint64 = 100

	// DefaultScoringModelVersion is the current default scoring model version
	DefaultScoringModelVersion = "1.0.0"
)

// ============================================================================
// Scoring Model Version
// ============================================================================

// ScoringModelVersion represents a versioned scoring model configuration
type ScoringModelVersion struct {
	// Version is the semantic version of the scoring model (e.g., "1.0.0")
	Version string `json:"version"`

	// CreatedAt is when this version was created
	CreatedAt time.Time `json:"created_at"`

	// ActivatedAt is when this version became active (nil if pending)
	ActivatedAt *time.Time `json:"activated_at,omitempty"`

	// DeprecatedAt is when this version was deprecated (nil if active)
	DeprecatedAt *time.Time `json:"deprecated_at,omitempty"`

	// Weights contains the feature weights for this version
	Weights ScoringWeights `json:"weights"`

	// Thresholds contains the score thresholds for this version
	Thresholds ScoringThresholds `json:"thresholds"`

	// Config contains additional scoring configuration
	Config ScoringConfig `json:"config"`

	// Description is a human-readable description of this version
	Description string `json:"description,omitempty"`
}

// ScoringModelStatus represents the status of a scoring model version
type ScoringModelStatus string

const (
	// ScoringModelStatusPending indicates the version is pending activation
	ScoringModelStatusPending ScoringModelStatus = "pending"

	// ScoringModelStatusActive indicates the version is active
	ScoringModelStatusActive ScoringModelStatus = "active"

	// ScoringModelStatusDeprecated indicates the version is deprecated
	ScoringModelStatusDeprecated ScoringModelStatus = "deprecated"
)

// ============================================================================
// Scoring Weights (Fixed-Point, Basis Points)
// ============================================================================

// ScoringWeights defines the weights for each scoring component in basis points.
// All weights must sum to 10000 (100%).
type ScoringWeights struct {
	// FaceSimilarityWeight is the weight for face similarity score (basis points)
	FaceSimilarityWeight uint32 `json:"face_similarity_weight"`

	// OCRConfidenceWeight is the weight for OCR confidence score (basis points)
	OCRConfidenceWeight uint32 `json:"ocr_confidence_weight"`

	// DocIntegrityWeight is the weight for document integrity checks (basis points)
	DocIntegrityWeight uint32 `json:"doc_integrity_weight"`

	// SaltBindingWeight is the weight for salt-binding verification (basis points)
	SaltBindingWeight uint32 `json:"salt_binding_weight"`

	// LivenessCheckWeight is the weight for liveness detection (basis points)
	LivenessCheckWeight uint32 `json:"liveness_check_weight"`

	// CaptureQualityWeight is the weight for capture quality metrics (basis points)
	CaptureQualityWeight uint32 `json:"capture_quality_weight"`
}

// DefaultScoringWeights returns the default v1 scoring weights
func DefaultScoringWeights() ScoringWeights {
	return ScoringWeights{
		FaceSimilarityWeight: 3000, // 30%
		OCRConfidenceWeight:  2500, // 25%
		DocIntegrityWeight:   2000, // 20%
		SaltBindingWeight:    1000, // 10%
		LivenessCheckWeight:  1000, // 10%
		CaptureQualityWeight: 500,  // 5%
	}
}

// Validate ensures the weights sum to exactly 10000 basis points
func (w ScoringWeights) Validate() error {
	total := uint32(0)
	total += w.FaceSimilarityWeight
	total += w.OCRConfidenceWeight
	total += w.DocIntegrityWeight
	total += w.SaltBindingWeight
	total += w.LivenessCheckWeight
	total += w.CaptureQualityWeight

	if total != uint32(MaxBasisPoints) {
		return ErrInvalidScoringModel.Wrapf("weights must sum to %d, got %d", MaxBasisPoints, total)
	}
	return nil
}

// TotalWeight returns the sum of all weights
func (w ScoringWeights) TotalWeight() uint32 {
	return w.FaceSimilarityWeight + w.OCRConfidenceWeight + w.DocIntegrityWeight +
		w.SaltBindingWeight + w.LivenessCheckWeight + w.CaptureQualityWeight
}

// ============================================================================
// Scoring Thresholds
// ============================================================================

// ScoringThresholds defines thresholds for scoring decisions
type ScoringThresholds struct {
	// MinFaceSimilarity is the minimum face similarity score to pass (basis points)
	MinFaceSimilarity uint32 `json:"min_face_similarity"`

	// MinOCRConfidence is the minimum OCR confidence to pass (basis points)
	MinOCRConfidence uint32 `json:"min_ocr_confidence"`

	// MinDocIntegrity is the minimum document integrity score (basis points)
	MinDocIntegrity uint32 `json:"min_doc_integrity"`

	// MinLivenessScore is the minimum liveness score to pass (basis points)
	MinLivenessScore uint32 `json:"min_liveness_score"`

	// MinCaptureQuality is the minimum capture quality score (basis points)
	MinCaptureQuality uint32 `json:"min_capture_quality"`

	// RequiredForPass is the minimum final score to pass verification (0-100)
	RequiredForPass uint32 `json:"required_for_pass"`

	// MissingSelfieMaxScore is the maximum score when selfie is missing (0-100)
	MissingSelfieMaxScore uint32 `json:"missing_selfie_max_score"`

	// MissingDocMaxScore is the maximum score when document is missing (0-100)
	MissingDocMaxScore uint32 `json:"missing_doc_max_score"`
}

// DefaultScoringThresholds returns the default v1 scoring thresholds
func DefaultScoringThresholds() ScoringThresholds {
	return ScoringThresholds{
		MinFaceSimilarity:     7000, // 70%
		MinOCRConfidence:      6000, // 60%
		MinDocIntegrity:       5000, // 50%
		MinLivenessScore:      7500, // 75%
		MinCaptureQuality:     5000, // 50%
		RequiredForPass:       50,   // Minimum score to pass
		MissingSelfieMaxScore: 0,    // Cannot score without selfie
		MissingDocMaxScore:    30,   // Limited score without document
	}
}

// Validate ensures thresholds are within valid ranges
func (t ScoringThresholds) Validate() error {
	if t.MinFaceSimilarity > uint32(MaxBasisPoints) {
		return ErrInvalidScoringModel.Wrap("min_face_similarity exceeds maximum")
	}
	if t.MinOCRConfidence > uint32(MaxBasisPoints) {
		return ErrInvalidScoringModel.Wrap("min_ocr_confidence exceeds maximum")
	}
	if t.MinDocIntegrity > uint32(MaxBasisPoints) {
		return ErrInvalidScoringModel.Wrap("min_doc_integrity exceeds maximum")
	}
	if t.MinLivenessScore > uint32(MaxBasisPoints) {
		return ErrInvalidScoringModel.Wrap("min_liveness_score exceeds maximum")
	}
	if t.MinCaptureQuality > uint32(MaxBasisPoints) {
		return ErrInvalidScoringModel.Wrap("min_capture_quality exceeds maximum")
	}
	if t.RequiredForPass > MaxScore {
		return ErrInvalidScoringModel.Wrap("required_for_pass exceeds maximum score")
	}
	if t.MissingSelfieMaxScore > MaxScore {
		return ErrInvalidScoringModel.Wrap("missing_selfie_max_score exceeds maximum")
	}
	if t.MissingDocMaxScore > MaxScore {
		return ErrInvalidScoringModel.Wrap("missing_doc_max_score exceeds maximum")
	}
	return nil
}

// ============================================================================
// Scoring Configuration
// ============================================================================

// ScoringConfig contains additional scoring configuration
type ScoringConfig struct {
	// RequireSelfie indicates if selfie is mandatory for scoring
	RequireSelfie bool `json:"require_selfie"`

	// RequireDocument indicates if document is mandatory for scoring
	RequireDocument bool `json:"require_document"`

	// RequireLiveness indicates if liveness check is mandatory
	RequireLiveness bool `json:"require_liveness"`

	// AllowFallbackOnMissingSelfie allows fallback scoring when selfie is missing
	AllowFallbackOnMissingSelfie bool `json:"allow_fallback_on_missing_selfie"`

	// AllowFallbackOnMissingDoc allows fallback scoring when document is missing
	AllowFallbackOnMissingDoc bool `json:"allow_fallback_on_missing_doc"`

	// MaxScoreWithoutLiveness caps score when liveness is not available
	MaxScoreWithoutLiveness uint32 `json:"max_score_without_liveness"`
}

// DefaultScoringConfig returns the default v1 scoring configuration
func DefaultScoringConfig() ScoringConfig {
	return ScoringConfig{
		RequireSelfie:                true,
		RequireDocument:              true,
		RequireLiveness:              false, // Optional in v1
		AllowFallbackOnMissingSelfie: false,
		AllowFallbackOnMissingDoc:    true,
		MaxScoreWithoutLiveness:      70,
	}
}

// ============================================================================
// Scoring Inputs (Feature Fusion Sources)
// ============================================================================

// FaceSimilarityInput represents face similarity verification input
type FaceSimilarityInput struct {
	// SimilarityScore is the face similarity score in basis points (0-10000)
	SimilarityScore uint32 `json:"similarity_score"`

	// Confidence is the model confidence in the similarity (0-10000)
	Confidence uint32 `json:"confidence"`

	// EmbeddingHash is the SHA256 hash of the face embedding used
	EmbeddingHash []byte `json:"embedding_hash,omitempty"`

	// ModelVersion is the face detection model version used
	ModelVersion string `json:"model_version,omitempty"`

	// Present indicates if face data was available for scoring
	Present bool `json:"present"`
}

// ToBasisPoints converts a 0-1 float to basis points (fixed-point)
// This should be used at the boundary when receiving external input
func ToBasisPoints(value float64) uint32 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return uint32(MaxBasisPoints)
	}
	return uint32(value * float64(MaxBasisPoints))
}

// FromBasisPoints converts basis points to a 0-100 score
func FromBasisPoints(bp uint32) uint32 {
	return (bp * 100) / uint32(MaxBasisPoints)
}

// Validate validates the face similarity input
func (f FaceSimilarityInput) Validate() error {
	if f.SimilarityScore > uint32(MaxBasisPoints) {
		return ErrInvalidScoringInput.Wrap("similarity_score exceeds maximum")
	}
	if f.Confidence > uint32(MaxBasisPoints) {
		return ErrInvalidScoringInput.Wrap("confidence exceeds maximum")
	}
	return nil
}

// OCRConfidenceInput represents OCR extraction confidence input
type OCRConfidenceInput struct {
	// FieldConfidences maps field names to confidence scores in basis points
	FieldConfidences map[string]uint32 `json:"field_confidences"`

	// FieldValidation maps field names to validation pass/fail
	FieldValidation map[string]bool `json:"field_validation"`

	// OverallConfidence is the weighted average confidence (basis points)
	OverallConfidence uint32 `json:"overall_confidence"`

	// ExtractedFieldCount is the number of successfully extracted fields
	ExtractedFieldCount uint32 `json:"extracted_field_count"`

	// ExpectedFieldCount is the number of expected fields
	ExpectedFieldCount uint32 `json:"expected_field_count"`

	// Present indicates if OCR data was available for scoring
	Present bool `json:"present"`
}

// Validate validates the OCR confidence input
func (o OCRConfidenceInput) Validate() error {
	if o.OverallConfidence > uint32(MaxBasisPoints) {
		return ErrInvalidScoringInput.Wrap("overall_confidence exceeds maximum")
	}
	for field, conf := range o.FieldConfidences {
		if conf > uint32(MaxBasisPoints) {
			return ErrInvalidScoringInput.Wrapf("field %s confidence exceeds maximum", field)
		}
	}
	return nil
}

// ComputeFieldScore computes a score based on field extraction success
func (o OCRConfidenceInput) ComputeFieldScore() uint32 {
	if o.ExpectedFieldCount == 0 {
		return 0
	}
	// Return percentage of fields extracted as basis points
	return (o.ExtractedFieldCount * uint32(MaxBasisPoints)) / o.ExpectedFieldCount
}

// DocIntegrityInput represents document integrity verification input
type DocIntegrityInput struct {
	// QualityScore is the document quality score in basis points (0-10000)
	QualityScore uint32 `json:"quality_score"`

	// FormatValid indicates if the document format is valid
	FormatValid bool `json:"format_valid"`

	// TemplateMatch indicates if document matches known templates
	TemplateMatch bool `json:"template_match"`

	// TamperDetectionPassed indicates if tampering detection passed
	TamperDetectionPassed bool `json:"tamper_detection_passed"`

	// ExpiryValid indicates if the document is not expired
	ExpiryValid bool `json:"expiry_valid"`

	// SharpnessScore is the image sharpness score (basis points)
	SharpnessScore uint32 `json:"sharpness_score"`

	// BrightnessScore is the image brightness score (basis points)
	BrightnessScore uint32 `json:"brightness_score"`

	// ContrastScore is the image contrast score (basis points)
	ContrastScore uint32 `json:"contrast_score"`

	// Present indicates if document data was available for scoring
	Present bool `json:"present"`
}

// Validate validates the document integrity input
func (d DocIntegrityInput) Validate() error {
	if d.QualityScore > uint32(MaxBasisPoints) {
		return ErrInvalidScoringInput.Wrap("quality_score exceeds maximum")
	}
	if d.SharpnessScore > uint32(MaxBasisPoints) {
		return ErrInvalidScoringInput.Wrap("sharpness_score exceeds maximum")
	}
	if d.BrightnessScore > uint32(MaxBasisPoints) {
		return ErrInvalidScoringInput.Wrap("brightness_score exceeds maximum")
	}
	if d.ContrastScore > uint32(MaxBasisPoints) {
		return ErrInvalidScoringInput.Wrap("contrast_score exceeds maximum")
	}
	return nil
}

// ComputeIntegrityScore computes an overall integrity score
func (d DocIntegrityInput) ComputeIntegrityScore() uint32 {
	if !d.Present {
		return 0
	}

	// Base score from quality
	score := uint64(d.QualityScore)

	// Penalties for failed checks
	if !d.FormatValid {
		score = (score * 70) / 100 // 30% penalty
	}
	if !d.TemplateMatch {
		score = (score * 90) / 100 // 10% penalty
	}
	if !d.TamperDetectionPassed {
		score = (score * 50) / 100 // 50% penalty
	}
	if !d.ExpiryValid {
		score = (score * 80) / 100 // 20% penalty
	}

	if score > uint64(MaxBasisPoints) {
		return uint32(MaxBasisPoints)
	}
	return uint32(score)
}

// SaltBindingInput represents salt-binding verification input
type SaltBindingInput struct {
	// SaltPresent indicates if salt was present in upload
	SaltPresent bool `json:"salt_present"`

	// SaltValid indicates if salt binding passed verification
	SaltValid bool `json:"salt_valid"`

	// SaltHash is the hash of the verified salt
	SaltHash []byte `json:"salt_hash,omitempty"`

	// ClientSignatureValid indicates if client signature is valid
	ClientSignatureValid bool `json:"client_signature_valid"`

	// UserSignatureValid indicates if user signature is valid
	UserSignatureValid bool `json:"user_signature_valid"`
}

// ComputeBindingScore computes a score for salt binding verification
func (s SaltBindingInput) ComputeBindingScore() uint32 {
	if !s.SaltPresent {
		return 0
	}

	var score uint32 = 0

	// Each valid component contributes equally
	if s.SaltValid {
		score += 3333 // ~33%
	}
	if s.ClientSignatureValid {
		score += 3333 // ~33%
	}
	if s.UserSignatureValid {
		score += 3334 // ~34%
	}

	return score
}

// LivenessCheckInput represents liveness detection input
type LivenessCheckInput struct {
	// LivenessScore is the liveness detection score in basis points
	LivenessScore uint32 `json:"liveness_score"`

	// LivenessMethod is the method used for liveness detection
	LivenessMethod string `json:"liveness_method,omitempty"`

	// VideoFrameCount is the number of video frames analyzed
	VideoFrameCount uint32 `json:"video_frame_count"`

	// BlinkDetected indicates if eye blink was detected
	BlinkDetected bool `json:"blink_detected"`

	// HeadMovementDetected indicates if head movement was detected
	HeadMovementDetected bool `json:"head_movement_detected"`

	// Present indicates if liveness data was available
	Present bool `json:"present"`
}

// Validate validates the liveness check input
func (l LivenessCheckInput) Validate() error {
	if l.LivenessScore > uint32(MaxBasisPoints) {
		return ErrInvalidScoringInput.Wrap("liveness_score exceeds maximum")
	}
	return nil
}

// ComputeLivenessScore computes an overall liveness score
func (l LivenessCheckInput) ComputeLivenessScore() uint32 {
	if !l.Present {
		return 0
	}

	score := uint64(l.LivenessScore)

	// Bonus for additional liveness signals
	if l.BlinkDetected {
		score = (score * 105) / 100 // 5% bonus
	}
	if l.HeadMovementDetected {
		score = (score * 105) / 100 // 5% bonus
	}
	if l.VideoFrameCount >= 30 {
		score = (score * 103) / 100 // 3% bonus for adequate frames
	}

	if score > uint64(MaxBasisPoints) {
		return uint32(MaxBasisPoints)
	}
	return uint32(score)
}

// CaptureQualityInput represents capture quality metrics input
type CaptureQualityInput struct {
	// OverallQuality is the overall capture quality score (basis points)
	OverallQuality uint32 `json:"overall_quality"`

	// ResolutionAdequate indicates if image resolution meets minimum
	ResolutionAdequate bool `json:"resolution_adequate"`

	// LightingScore is the lighting quality score (basis points)
	LightingScore uint32 `json:"lighting_score"`

	// FocusScore is the focus/sharpness score (basis points)
	FocusScore uint32 `json:"focus_score"`

	// AngleScore is the capture angle score (basis points)
	AngleScore uint32 `json:"angle_score"`

	// Present indicates if capture quality data was available
	Present bool `json:"present"`
}

// Validate validates the capture quality input
func (c CaptureQualityInput) Validate() error {
	if c.OverallQuality > uint32(MaxBasisPoints) {
		return ErrInvalidScoringInput.Wrap("overall_quality exceeds maximum")
	}
	if c.LightingScore > uint32(MaxBasisPoints) {
		return ErrInvalidScoringInput.Wrap("lighting_score exceeds maximum")
	}
	if c.FocusScore > uint32(MaxBasisPoints) {
		return ErrInvalidScoringInput.Wrap("focus_score exceeds maximum")
	}
	if c.AngleScore > uint32(MaxBasisPoints) {
		return ErrInvalidScoringInput.Wrap("angle_score exceeds maximum")
	}
	return nil
}

// ComputeCaptureScore computes an overall capture quality score
func (c CaptureQualityInput) ComputeCaptureScore() uint32 {
	if !c.Present {
		return uint32(MaxBasisPoints) / 2 // Default 50% if not present
	}

	// Weighted average of quality components
	score := uint64(c.OverallQuality)*40 +
		uint64(c.LightingScore)*20 +
		uint64(c.FocusScore)*25 +
		uint64(c.AngleScore)*15

	score = score / 100

	// Penalty for inadequate resolution
	if !c.ResolutionAdequate {
		score = (score * 70) / 100 // 30% penalty
	}

	if score > uint64(MaxBasisPoints) {
		return uint32(MaxBasisPoints)
	}
	return uint32(score)
}

// ============================================================================
// Composite Scoring Input
// ============================================================================

// ScoringInputs aggregates all scoring inputs for feature fusion
type ScoringInputs struct {
	// FaceSimilarity contains face verification input
	FaceSimilarity FaceSimilarityInput `json:"face_similarity"`

	// OCRConfidence contains OCR extraction input
	OCRConfidence OCRConfidenceInput `json:"ocr_confidence"`

	// DocIntegrity contains document integrity input
	DocIntegrity DocIntegrityInput `json:"doc_integrity"`

	// SaltBinding contains salt-binding verification input
	SaltBinding SaltBindingInput `json:"salt_binding"`

	// LivenessCheck contains liveness detection input
	LivenessCheck LivenessCheckInput `json:"liveness_check"`

	// CaptureQuality contains capture quality metrics input
	CaptureQuality CaptureQualityInput `json:"capture_quality"`

	// AccountAddress is the account being scored
	AccountAddress string `json:"account_address"`

	// BlockHeight is the block height when scoring
	BlockHeight int64 `json:"block_height"`

	// Timestamp is when scoring was performed
	Timestamp time.Time `json:"timestamp"`
}

// Validate validates all scoring inputs
func (s ScoringInputs) Validate() error {
	if err := s.FaceSimilarity.Validate(); err != nil {
		return err
	}
	if err := s.OCRConfidence.Validate(); err != nil {
		return err
	}
	if err := s.DocIntegrity.Validate(); err != nil {
		return err
	}
	if err := s.LivenessCheck.Validate(); err != nil {
		return err
	}
	if err := s.CaptureQuality.Validate(); err != nil {
		return err
	}
	return nil
}

// ComputeInputHash computes a deterministic hash of all inputs
func (s ScoringInputs) ComputeInputHash() []byte {
	h := sha256.New()

	// Account and block context
	h.Write([]byte(s.AccountAddress))
	blockBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(blockBytes, uint64(s.BlockHeight))
	h.Write(blockBytes)

	// Face similarity inputs
	h.Write(encodeUint32(s.FaceSimilarity.SimilarityScore))
	h.Write(encodeUint32(s.FaceSimilarity.Confidence))
	h.Write(encodeBool(s.FaceSimilarity.Present))

	// OCR inputs
	h.Write(encodeUint32(s.OCRConfidence.OverallConfidence))
	h.Write(encodeUint32(s.OCRConfidence.ExtractedFieldCount))
	h.Write(encodeBool(s.OCRConfidence.Present))

	// Doc integrity inputs
	h.Write(encodeUint32(s.DocIntegrity.QualityScore))
	h.Write(encodeBool(s.DocIntegrity.FormatValid))
	h.Write(encodeBool(s.DocIntegrity.TamperDetectionPassed))
	h.Write(encodeBool(s.DocIntegrity.Present))

	// Salt binding inputs
	h.Write(encodeBool(s.SaltBinding.SaltPresent))
	h.Write(encodeBool(s.SaltBinding.SaltValid))
	h.Write(encodeBool(s.SaltBinding.ClientSignatureValid))
	h.Write(encodeBool(s.SaltBinding.UserSignatureValid))

	// Liveness inputs
	h.Write(encodeUint32(s.LivenessCheck.LivenessScore))
	h.Write(encodeBool(s.LivenessCheck.Present))

	// Capture quality inputs
	h.Write(encodeUint32(s.CaptureQuality.OverallQuality))
	h.Write(encodeBool(s.CaptureQuality.Present))

	return h.Sum(nil)
}

// Helper functions for deterministic encoding
func encodeUint32(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

func encodeBool(v bool) []byte {
	if v {
		return []byte{1}
	}
	return []byte{0}
}

// ============================================================================
// Scoring Output (Evidence Summary)
// ============================================================================

// FeatureContribution represents a single feature's contribution to the score
type FeatureContribution struct {
	// FeatureName is the name of the feature
	FeatureName string `json:"feature_name"`

	// RawScore is the feature's raw score in basis points
	RawScore uint32 `json:"raw_score"`

	// Weight is the feature's weight in basis points
	Weight uint32 `json:"weight"`

	// WeightedScore is (RawScore * Weight) / 10000
	WeightedScore uint32 `json:"weighted_score"`

	// PassedThreshold indicates if the feature passed its threshold
	PassedThreshold bool `json:"passed_threshold"`

	// ReasonCode is the reason code if feature failed
	ReasonCode ReasonCode `json:"reason_code,omitempty"`
}

// ScoringReasonCode represents a detailed reason for the scoring outcome
type ScoringReasonCode string

const (
	// ScoringReasonSuccess indicates successful scoring
	ScoringReasonSuccess ScoringReasonCode = "SCORING_SUCCESS"

	// ScoringReasonMissingSelfie indicates selfie was not provided
	ScoringReasonMissingSelfie ScoringReasonCode = "MISSING_SELFIE"

	// ScoringReasonMissingDocument indicates document was not provided
	ScoringReasonMissingDocument ScoringReasonCode = "MISSING_DOCUMENT"

	// ScoringReasonLowFaceSimilarity indicates face similarity below threshold
	ScoringReasonLowFaceSimilarity ScoringReasonCode = "LOW_FACE_SIMILARITY"

	// ScoringReasonLowOCRConfidence indicates OCR confidence below threshold
	ScoringReasonLowOCRConfidence ScoringReasonCode = "LOW_OCR_CONFIDENCE"

	// ScoringReasonLowDocIntegrity indicates document integrity issues
	ScoringReasonLowDocIntegrity ScoringReasonCode = "LOW_DOC_INTEGRITY"

	// ScoringReasonInvalidSaltBinding indicates salt binding failed
	ScoringReasonInvalidSaltBinding ScoringReasonCode = "INVALID_SALT_BINDING"

	// ScoringReasonLivenessCheckFailed indicates liveness check failed
	ScoringReasonLivenessCheckFailed ScoringReasonCode = "LIVENESS_CHECK_FAILED"

	// ScoringReasonLowCaptureQuality indicates capture quality issues
	ScoringReasonLowCaptureQuality ScoringReasonCode = "LOW_CAPTURE_QUALITY"

	// ScoringReasonBelowPassThreshold indicates final score below pass threshold
	ScoringReasonBelowPassThreshold ScoringReasonCode = "BELOW_PASS_THRESHOLD"

	// ScoringReasonFallbackApplied indicates fallback scoring was applied
	ScoringReasonFallbackApplied ScoringReasonCode = "FALLBACK_APPLIED"

	// ScoringReasonScoreCapped indicates score was capped due to missing components
	ScoringReasonScoreCapped ScoringReasonCode = "SCORE_CAPPED"
)

// EvidenceSummary contains a non-sensitive summary of scoring evidence
type EvidenceSummary struct {
	// FinalScore is the computed identity score (0-100)
	FinalScore uint32 `json:"final_score"`

	// Passed indicates if the verification passed
	Passed bool `json:"passed"`

	// ModelVersion is the scoring model version used
	ModelVersion string `json:"model_version"`

	// Contributions contains each feature's contribution to the score
	Contributions []FeatureContribution `json:"contributions"`

	// ReasonCodes contains codes explaining the scoring outcome
	ReasonCodes []ScoringReasonCode `json:"reason_codes"`

	// InputHash is the hash of all inputs for verification
	InputHash []byte `json:"input_hash"`

	// ComputedAt is when the score was computed
	ComputedAt time.Time `json:"computed_at"`

	// BlockHeight is the block height when computed
	BlockHeight int64 `json:"block_height"`

	// FeaturePresence tracks which features were present
	FeaturePresence map[string]bool `json:"feature_presence"`

	// ThresholdsApplied contains threshold values that were applied
	ThresholdsApplied map[string]uint32 `json:"thresholds_applied,omitempty"`
}

// NewEvidenceSummary creates a new evidence summary
func NewEvidenceSummary(modelVersion string, blockHeight int64, computedAt time.Time) *EvidenceSummary {
	return &EvidenceSummary{
		ModelVersion:      modelVersion,
		BlockHeight:       blockHeight,
		ComputedAt:        computedAt,
		Contributions:     make([]FeatureContribution, 0),
		ReasonCodes:       make([]ScoringReasonCode, 0),
		FeaturePresence:   make(map[string]bool),
		ThresholdsApplied: make(map[string]uint32),
	}
}

// AddContribution adds a feature contribution to the summary
func (e *EvidenceSummary) AddContribution(contrib FeatureContribution) {
	e.Contributions = append(e.Contributions, contrib)
	e.FeaturePresence[contrib.FeatureName] = contrib.RawScore > 0
}

// AddReasonCode adds a reason code to the summary
func (e *EvidenceSummary) AddReasonCode(code ScoringReasonCode) {
	// Avoid duplicates
	for _, existing := range e.ReasonCodes {
		if existing == code {
			return
		}
	}
	e.ReasonCodes = append(e.ReasonCodes, code)
}

// SetResult sets the final result
func (e *EvidenceSummary) SetResult(score uint32, passed bool, inputHash []byte) {
	e.FinalScore = score
	e.Passed = passed
	e.InputHash = inputHash

	if passed {
		e.AddReasonCode(ScoringReasonSuccess)
	}
}

// ============================================================================
// Score Version History
// ============================================================================

// ScoreVersionTransition records a transition between scoring model versions
type ScoreVersionTransition struct {
	// FromVersion is the previous scoring model version
	FromVersion string `json:"from_version"`

	// ToVersion is the new scoring model version
	ToVersion string `json:"to_version"`

	// AccountAddress is the account affected
	AccountAddress string `json:"account_address"`

	// PreviousScore is the score before transition
	PreviousScore uint32 `json:"previous_score"`

	// NewScore is the score after transition
	NewScore uint32 `json:"new_score"`

	// TransitionReason explains why the transition occurred
	TransitionReason string `json:"transition_reason"`

	// TransitionTime is when the transition occurred
	TransitionTime time.Time `json:"transition_time"`

	// BlockHeight is the block height of the transition
	BlockHeight int64 `json:"block_height"`
}

// ScoringHistoryEntry represents a historical scoring record
type ScoringHistoryEntry struct {
	// Score is the identity score (0-100)
	Score uint32 `json:"score"`

	// ModelVersion is the scoring model version used
	ModelVersion string `json:"model_version"`

	// EvidenceSummaryHash is the hash of the evidence summary
	EvidenceSummaryHash []byte `json:"evidence_summary_hash"`

	// ReasonCodes contains the primary reason codes
	ReasonCodes []ScoringReasonCode `json:"reason_codes"`

	// ComputedAt is when the score was computed
	ComputedAt time.Time `json:"computed_at"`

	// BlockHeight is the block height when computed
	BlockHeight int64 `json:"block_height"`
}

// ============================================================================
// Default Scoring Model
// ============================================================================

// DefaultScoringModel returns the default v1 scoring model
func DefaultScoringModel() ScoringModelVersion {
	now := time.Now()
	return ScoringModelVersion{
		Version:     DefaultScoringModelVersion,
		CreatedAt:   now,
		ActivatedAt: &now,
		Weights:     DefaultScoringWeights(),
		Thresholds:  DefaultScoringThresholds(),
		Config:      DefaultScoringConfig(),
		Description: "VEID Scoring Model v1.0.0 - Feature fusion from doc OCR + face match + metadata",
	}
}

// ============================================================================
// Scoring Model Validation
// ============================================================================

// Validate validates the complete scoring model version
func (m ScoringModelVersion) Validate() error {
	if m.Version == "" {
		return ErrInvalidScoringModel.Wrap("version cannot be empty")
	}

	if err := m.Weights.Validate(); err != nil {
		return fmt.Errorf("invalid weights: %w", err)
	}

	if err := m.Thresholds.Validate(); err != nil {
		return fmt.Errorf("invalid thresholds: %w", err)
	}

	return nil
}

// ComputeModelHash computes a deterministic hash of the model configuration
func (m ScoringModelVersion) ComputeModelHash() []byte {
	h := sha256.New()

	// Version
	h.Write([]byte(m.Version))

	// Weights (in fixed order)
	h.Write(encodeUint32(m.Weights.FaceSimilarityWeight))
	h.Write(encodeUint32(m.Weights.OCRConfidenceWeight))
	h.Write(encodeUint32(m.Weights.DocIntegrityWeight))
	h.Write(encodeUint32(m.Weights.SaltBindingWeight))
	h.Write(encodeUint32(m.Weights.LivenessCheckWeight))
	h.Write(encodeUint32(m.Weights.CaptureQualityWeight))

	// Thresholds (in fixed order)
	h.Write(encodeUint32(m.Thresholds.MinFaceSimilarity))
	h.Write(encodeUint32(m.Thresholds.MinOCRConfidence))
	h.Write(encodeUint32(m.Thresholds.MinDocIntegrity))
	h.Write(encodeUint32(m.Thresholds.MinLivenessScore))
	h.Write(encodeUint32(m.Thresholds.MinCaptureQuality))
	h.Write(encodeUint32(m.Thresholds.RequiredForPass))

	// Config (in fixed order)
	h.Write(encodeBool(m.Config.RequireSelfie))
	h.Write(encodeBool(m.Config.RequireDocument))
	h.Write(encodeBool(m.Config.RequireLiveness))

	return h.Sum(nil)
}

// ============================================================================
// Feature Names (Constants)
// ============================================================================

// Feature name constants for scoring contributions
const (
	FeatureNameFaceSimilarity = "face_similarity"
	FeatureNameOCRConfidence  = "ocr_confidence"
	FeatureNameDocIntegrity   = "doc_integrity"
	FeatureNameSaltBinding    = "salt_binding"
	FeatureNameLivenessCheck  = "liveness_check"
	FeatureNameCaptureQuality = "capture_quality"
)

// AllFeatureNames returns all feature names in consistent order
func AllFeatureNames() []string {
	return []string{
		FeatureNameFaceSimilarity,
		FeatureNameOCRConfidence,
		FeatureNameDocIntegrity,
		FeatureNameSaltBinding,
		FeatureNameLivenessCheck,
		FeatureNameCaptureQuality,
	}
}

// ============================================================================
// Deterministic Score Computation
// ============================================================================

// ComputeDeterministicScore computes the identity score using fixed-point arithmetic
// This function is deterministic and consensus-safe.
func ComputeDeterministicScore(
	inputs ScoringInputs,
	model ScoringModelVersion,
) (*EvidenceSummary, error) {
	// Validate inputs
	if err := inputs.Validate(); err != nil {
		return nil, err
	}

	// Validate model
	if err := model.Validate(); err != nil {
		return nil, err
	}

	// Create evidence summary
	summary := NewEvidenceSummary(model.Version, inputs.BlockHeight, inputs.Timestamp)

	// Track reason codes and whether we can proceed
	var reasonCodes []ScoringReasonCode

	// Check for required components
	hasSelfie := inputs.FaceSimilarity.Present
	hasDocument := inputs.DocIntegrity.Present

	// Handle missing required components
	if !hasSelfie && model.Config.RequireSelfie && !model.Config.AllowFallbackOnMissingSelfie {
		reasonCodes = append(reasonCodes, ScoringReasonMissingSelfie)
		summary.ReasonCodes = reasonCodes
		summary.SetResult(0, false, inputs.ComputeInputHash())
		return summary, nil
	}

	if !hasDocument && model.Config.RequireDocument && !model.Config.AllowFallbackOnMissingDoc {
		reasonCodes = append(reasonCodes, ScoringReasonMissingDocument)
		summary.ReasonCodes = reasonCodes
		summary.SetResult(0, false, inputs.ComputeInputHash())
		return summary, nil
	}

	// Compute individual feature scores (all in basis points)
	var contributions []FeatureContribution

	// 1. Face Similarity
	faceContrib := computeFaceContribution(inputs.FaceSimilarity, model.Weights, model.Thresholds)
	contributions = append(contributions, faceContrib)
	if !faceContrib.PassedThreshold && inputs.FaceSimilarity.Present {
		reasonCodes = append(reasonCodes, ScoringReasonLowFaceSimilarity)
	}

	// 2. OCR Confidence
	ocrContrib := computeOCRContribution(inputs.OCRConfidence, model.Weights, model.Thresholds)
	contributions = append(contributions, ocrContrib)
	if !ocrContrib.PassedThreshold && inputs.OCRConfidence.Present {
		reasonCodes = append(reasonCodes, ScoringReasonLowOCRConfidence)
	}

	// 3. Document Integrity
	docContrib := computeDocContribution(inputs.DocIntegrity, model.Weights, model.Thresholds)
	contributions = append(contributions, docContrib)
	if !docContrib.PassedThreshold && inputs.DocIntegrity.Present {
		reasonCodes = append(reasonCodes, ScoringReasonLowDocIntegrity)
	}

	// 4. Salt Binding
	saltContrib := computeSaltContribution(inputs.SaltBinding, model.Weights)
	contributions = append(contributions, saltContrib)
	if !saltContrib.PassedThreshold {
		reasonCodes = append(reasonCodes, ScoringReasonInvalidSaltBinding)
	}

	// 5. Liveness Check
	livenessContrib := computeLivenessContribution(inputs.LivenessCheck, model.Weights, model.Thresholds)
	contributions = append(contributions, livenessContrib)
	if !livenessContrib.PassedThreshold && inputs.LivenessCheck.Present {
		reasonCodes = append(reasonCodes, ScoringReasonLivenessCheckFailed)
	}

	// 6. Capture Quality
	captureContrib := computeCaptureContribution(inputs.CaptureQuality, model.Weights, model.Thresholds)
	contributions = append(contributions, captureContrib)
	if !captureContrib.PassedThreshold && inputs.CaptureQuality.Present {
		reasonCodes = append(reasonCodes, ScoringReasonLowCaptureQuality)
	}

	// Add all contributions to summary
	for _, c := range contributions {
		summary.AddContribution(c)
	}

	// Compute total weighted score (fixed-point arithmetic)
	// Total = sum(weighted_scores) / 100 (to convert from basis points to 0-100)
	var totalWeightedScore uint64 = 0
	var activeWeight uint64 = 0

	for _, c := range contributions {
		totalWeightedScore += uint64(c.WeightedScore)
		if c.RawScore > 0 {
			activeWeight += uint64(c.Weight)
		}
	}

	// Normalize score to 0-100 range
	var finalScore uint32
	if activeWeight > 0 {
		// Proportional scoring based on active features
		finalScore = uint32((totalWeightedScore * 100) / uint64(MaxBasisPoints))
	}

	// Apply caps for missing components
	scoreCapped := false
	if !hasSelfie && model.Config.AllowFallbackOnMissingSelfie {
		if finalScore > model.Thresholds.MissingSelfieMaxScore {
			finalScore = model.Thresholds.MissingSelfieMaxScore
			scoreCapped = true
		}
		reasonCodes = append(reasonCodes, ScoringReasonFallbackApplied)
	}

	if !hasDocument && model.Config.AllowFallbackOnMissingDoc {
		if finalScore > model.Thresholds.MissingDocMaxScore {
			finalScore = model.Thresholds.MissingDocMaxScore
			scoreCapped = true
		}
		reasonCodes = append(reasonCodes, ScoringReasonFallbackApplied)
	}

	// Cap score when liveness is not present
	if !inputs.LivenessCheck.Present && model.Config.MaxScoreWithoutLiveness < 100 {
		if finalScore > model.Config.MaxScoreWithoutLiveness {
			finalScore = model.Config.MaxScoreWithoutLiveness
			scoreCapped = true
		}
	}

	if scoreCapped {
		reasonCodes = append(reasonCodes, ScoringReasonScoreCapped)
	}

	// Ensure score doesn't exceed maximum
	if finalScore > MaxScore {
		finalScore = MaxScore
	}

	// Check if passed
	passed := finalScore >= model.Thresholds.RequiredForPass
	if !passed {
		reasonCodes = append(reasonCodes, ScoringReasonBelowPassThreshold)
	}

	// Set thresholds applied for transparency
	summary.ThresholdsApplied = map[string]uint32{
		"min_face_similarity": model.Thresholds.MinFaceSimilarity,
		"min_ocr_confidence":  model.Thresholds.MinOCRConfidence,
		"min_doc_integrity":   model.Thresholds.MinDocIntegrity,
		"required_for_pass":   model.Thresholds.RequiredForPass,
	}

	// Deduplicate and sort reason codes for determinism
	uniqueReasons := deduplicateReasonCodes(reasonCodes)
	sort.Slice(uniqueReasons, func(i, j int) bool {
		return string(uniqueReasons[i]) < string(uniqueReasons[j])
	})
	summary.ReasonCodes = uniqueReasons

	// Compute input hash and set result
	summary.SetResult(finalScore, passed, inputs.ComputeInputHash())

	return summary, nil
}

// Helper functions for computing individual feature contributions

func computeFaceContribution(input FaceSimilarityInput, weights ScoringWeights, thresholds ScoringThresholds) FeatureContribution {
	contrib := FeatureContribution{
		FeatureName: FeatureNameFaceSimilarity,
		Weight:      weights.FaceSimilarityWeight,
	}

	if !input.Present {
		contrib.RawScore = 0
		contrib.WeightedScore = 0
		contrib.PassedThreshold = false
		contrib.ReasonCode = ReasonCodeFaceMismatch
		return contrib
	}

	// Use similarity score weighted by confidence
	// Effective score = (similarity * confidence) / 10000
	effectiveScore := (uint64(input.SimilarityScore) * uint64(input.Confidence)) / uint64(MaxBasisPoints)
	contrib.RawScore = uint32(effectiveScore)

	// Compute weighted contribution
	contrib.WeightedScore = uint32((effectiveScore * uint64(weights.FaceSimilarityWeight)) / uint64(MaxBasisPoints))

	// Check threshold
	contrib.PassedThreshold = contrib.RawScore >= thresholds.MinFaceSimilarity

	if !contrib.PassedThreshold {
		contrib.ReasonCode = ReasonCodeFaceMismatch
	}

	return contrib
}

func computeOCRContribution(input OCRConfidenceInput, weights ScoringWeights, thresholds ScoringThresholds) FeatureContribution {
	contrib := FeatureContribution{
		FeatureName: FeatureNameOCRConfidence,
		Weight:      weights.OCRConfidenceWeight,
	}

	if !input.Present {
		contrib.RawScore = 0
		contrib.WeightedScore = 0
		contrib.PassedThreshold = false
		contrib.ReasonCode = ReasonCodeLowOCRConfidence
		return contrib
	}

	// Combine overall confidence with field extraction success
	fieldScore := input.ComputeFieldScore()
	combinedScore := (uint64(input.OverallConfidence)*70 + uint64(fieldScore)*30) / 100
	contrib.RawScore = uint32(combinedScore)

	// Compute weighted contribution
	contrib.WeightedScore = uint32((combinedScore * uint64(weights.OCRConfidenceWeight)) / uint64(MaxBasisPoints))

	// Check threshold
	contrib.PassedThreshold = contrib.RawScore >= thresholds.MinOCRConfidence

	if !contrib.PassedThreshold {
		contrib.ReasonCode = ReasonCodeLowOCRConfidence
	}

	return contrib
}

func computeDocContribution(input DocIntegrityInput, weights ScoringWeights, thresholds ScoringThresholds) FeatureContribution {
	contrib := FeatureContribution{
		FeatureName: FeatureNameDocIntegrity,
		Weight:      weights.DocIntegrityWeight,
	}

	if !input.Present {
		contrib.RawScore = 0
		contrib.WeightedScore = 0
		contrib.PassedThreshold = false
		contrib.ReasonCode = ReasonCodeDocumentInvalid
		return contrib
	}

	// Use computed integrity score
	contrib.RawScore = input.ComputeIntegrityScore()

	// Compute weighted contribution
	contrib.WeightedScore = uint32((uint64(contrib.RawScore) * uint64(weights.DocIntegrityWeight)) / uint64(MaxBasisPoints))

	// Check threshold
	contrib.PassedThreshold = contrib.RawScore >= thresholds.MinDocIntegrity

	if !contrib.PassedThreshold {
		contrib.ReasonCode = ReasonCodeDocumentInvalid
	}

	return contrib
}

func computeSaltContribution(input SaltBindingInput, weights ScoringWeights) FeatureContribution {
	contrib := FeatureContribution{
		FeatureName: FeatureNameSaltBinding,
		Weight:      weights.SaltBindingWeight,
	}

	// Use computed binding score
	contrib.RawScore = input.ComputeBindingScore()

	// Compute weighted contribution
	contrib.WeightedScore = uint32((uint64(contrib.RawScore) * uint64(weights.SaltBindingWeight)) / uint64(MaxBasisPoints))

	// Salt binding passes if all components are valid
	contrib.PassedThreshold = input.SaltPresent && input.SaltValid &&
		input.ClientSignatureValid && input.UserSignatureValid

	if !contrib.PassedThreshold {
		contrib.ReasonCode = ReasonCodeInvalidPayload
	}

	return contrib
}

func computeLivenessContribution(input LivenessCheckInput, weights ScoringWeights, thresholds ScoringThresholds) FeatureContribution {
	contrib := FeatureContribution{
		FeatureName: FeatureNameLivenessCheck,
		Weight:      weights.LivenessCheckWeight,
	}

	if !input.Present {
		// Liveness not present - may be optional
		contrib.RawScore = uint32(MaxBasisPoints) / 2 // Default 50% when not present
		contrib.WeightedScore = uint32((uint64(contrib.RawScore) * uint64(weights.LivenessCheckWeight)) / uint64(MaxBasisPoints))
		contrib.PassedThreshold = true // Not required to pass
		return contrib
	}

	// Use computed liveness score
	contrib.RawScore = input.ComputeLivenessScore()

	// Compute weighted contribution
	contrib.WeightedScore = uint32((uint64(contrib.RawScore) * uint64(weights.LivenessCheckWeight)) / uint64(MaxBasisPoints))

	// Check threshold
	contrib.PassedThreshold = contrib.RawScore >= thresholds.MinLivenessScore

	if !contrib.PassedThreshold {
		contrib.ReasonCode = ReasonCodeLivenessCheckFailed
	}

	return contrib
}

func computeCaptureContribution(input CaptureQualityInput, weights ScoringWeights, thresholds ScoringThresholds) FeatureContribution {
	contrib := FeatureContribution{
		FeatureName: FeatureNameCaptureQuality,
		Weight:      weights.CaptureQualityWeight,
	}

	// Use computed capture score
	contrib.RawScore = input.ComputeCaptureScore()

	// Compute weighted contribution
	contrib.WeightedScore = uint32((uint64(contrib.RawScore) * uint64(weights.CaptureQualityWeight)) / uint64(MaxBasisPoints))

	// Check threshold
	contrib.PassedThreshold = contrib.RawScore >= thresholds.MinCaptureQuality

	if !contrib.PassedThreshold {
		contrib.ReasonCode = ReasonCodeLowDocQuality
	}

	return contrib
}

func deduplicateReasonCodes(codes []ScoringReasonCode) []ScoringReasonCode {
	seen := make(map[ScoringReasonCode]bool)
	result := make([]ScoringReasonCode, 0, len(codes))
	for _, code := range codes {
		if !seen[code] {
			seen[code] = true
			result = append(result, code)
		}
	}
	return result
}
