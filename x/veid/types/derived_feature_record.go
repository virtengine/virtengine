package types

import (
	"crypto/sha256"
	"time"
)

// ============================================================================
// Derived Feature Verification Record
// ============================================================================

// DerivedFeatureVerificationRecordVersion is the current version
const DerivedFeatureVerificationRecordVersion uint32 = 1

// DerivedFeatureVerificationRecord represents the on-chain verification record
// that references derived features (hashes) for consensus verification
// SECURITY: This record contains ONLY hashes and references - NO raw biometric data
type DerivedFeatureVerificationRecord struct {
	// RecordID is the unique identifier for this record
	RecordID string `json:"record_id"`

	// AccountAddress is the account this record belongs to
	AccountAddress string `json:"account_address"`

	// Version is the record format version
	Version uint32 `json:"version"`

	// RequestID is the verification request this record responds to
	RequestID string `json:"request_id"`

	// FeatureReferences contains references to all derived features used
	FeatureReferences []DerivedFeatureReference `json:"feature_references"`

	// CompositeHash is the SHA-256 hash of all feature hashes combined
	// This allows validators to verify they used the same inputs
	CompositeHash []byte `json:"composite_hash"`

	// ModelVersion is the ML model version used for verification
	ModelVersion string `json:"model_version"`

	// ModelHash is the hash of the model weights
	ModelHash string `json:"model_hash"`

	// Score is the computed verification score (0-100)
	Score uint32 `json:"score"`

	// Confidence is the confidence level of the score (0-100)
	Confidence uint32 `json:"confidence"`

	// Status is the verification result status
	Status VerificationResultStatus `json:"status"`

	// ReasonCodes contains codes explaining the outcome
	ReasonCodes []ReasonCode `json:"reason_codes,omitempty"`

	// ComputedAt is when this record was computed
	ComputedAt time.Time `json:"computed_at"`

	// BlockHeight is the block at which this was computed
	BlockHeight int64 `json:"block_height"`

	// ComputedBy is the validator address that computed this
	ComputedBy string `json:"computed_by"`

	// ConsensusVotes tracks validator votes on this record
	ConsensusVotes []ConsensusVote `json:"consensus_votes,omitempty"`

	// Finalized indicates if the record has achieved consensus
	Finalized bool `json:"finalized"`

	// FinalizedAt is when the record was finalized
	FinalizedAt *time.Time `json:"finalized_at,omitempty"`

	// FinalizedAtBlock is the block at which this was finalized
	FinalizedAtBlock *int64 `json:"finalized_at_block,omitempty"`
}

// DerivedFeatureReference is a reference to a derived feature used in verification
type DerivedFeatureReference struct {
	// FeatureType is the type of feature (face_embedding, document_hash, etc.)
	FeatureType string `json:"feature_type"`

	// FeatureHash is the SHA-256 hash of the feature
	FeatureHash []byte `json:"feature_hash"`

	// SourceScopeID is the scope the feature was extracted from
	SourceScopeID string `json:"source_scope_id"`

	// EnvelopeID is the embedding envelope ID (for face embeddings)
	EnvelopeID string `json:"envelope_id,omitempty"`

	// FieldKey is the document field key (for document hashes)
	FieldKey string `json:"field_key,omitempty"`

	// Weight is the weight of this feature in scoring
	Weight uint32 `json:"weight"`

	// MatchResult indicates if this feature matched (for verification)
	MatchResult FeatureMatchResult `json:"match_result"`

	// MatchScore is the similarity score for this feature (0-100)
	MatchScore uint32 `json:"match_score"`
}

// FeatureMatchResult represents the result of a feature match
type FeatureMatchResult string

const (
	// FeatureMatchResultMatch indicates features matched
	FeatureMatchResultMatch FeatureMatchResult = "match"

	// FeatureMatchResultNoMatch indicates features did not match
	FeatureMatchResultNoMatch FeatureMatchResult = "no_match"

	// FeatureMatchResultPartial indicates a partial match
	FeatureMatchResultPartial FeatureMatchResult = "partial"

	// FeatureMatchResultError indicates an error during matching
	FeatureMatchResultError FeatureMatchResult = "error"

	// FeatureMatchResultSkipped indicates the feature was skipped
	FeatureMatchResultSkipped FeatureMatchResult = "skipped"
)

// AllFeatureMatchResults returns all valid match results
func AllFeatureMatchResults() []FeatureMatchResult {
	return []FeatureMatchResult{
		FeatureMatchResultMatch,
		FeatureMatchResultNoMatch,
		FeatureMatchResultPartial,
		FeatureMatchResultError,
		FeatureMatchResultSkipped,
	}
}

// IsValidFeatureMatchResult checks if a match result is valid
func IsValidFeatureMatchResult(r FeatureMatchResult) bool {
	for _, valid := range AllFeatureMatchResults() {
		if r == valid {
			return true
		}
	}
	return false
}

// ConsensusVote represents a validator's vote on a verification record
type ConsensusVote struct {
	// ValidatorAddress is the voting validator
	ValidatorAddress string `json:"validator_address"`

	// Agreed indicates if the validator agrees with the record
	Agreed bool `json:"agreed"`

	// ComputedHash is the hash the validator computed
	ComputedHash []byte `json:"computed_hash"`

	// ComputedScore is the score the validator computed
	ComputedScore uint32 `json:"computed_score"`

	// VotedAt is when the vote was cast
	VotedAt time.Time `json:"voted_at"`

	// BlockHeight is the block at which the vote was cast
	BlockHeight int64 `json:"block_height"`

	// Signature is the validator's signature on the vote
	Signature []byte `json:"signature"`
}

// NewDerivedFeatureVerificationRecord creates a new verification record
func NewDerivedFeatureVerificationRecord(
	recordID string,
	accountAddress string,
	requestID string,
	modelVersion string,
	modelHash string,
	computedAt time.Time,
	blockHeight int64,
	computedBy string,
) *DerivedFeatureVerificationRecord {
	return &DerivedFeatureVerificationRecord{
		RecordID:          recordID,
		AccountAddress:    accountAddress,
		Version:           DerivedFeatureVerificationRecordVersion,
		RequestID:         requestID,
		FeatureReferences: make([]DerivedFeatureReference, 0),
		ModelVersion:      modelVersion,
		ModelHash:         modelHash,
		Score:             0,
		Confidence:        0,
		Status:            VerificationResultStatusFailed,
		ReasonCodes:       make([]ReasonCode, 0),
		ComputedAt:        computedAt,
		BlockHeight:       blockHeight,
		ComputedBy:        computedBy,
		ConsensusVotes:    make([]ConsensusVote, 0),
		Finalized:         false,
	}
}

// Validate validates the verification record
func (r *DerivedFeatureVerificationRecord) Validate() error {
	if r.RecordID == "" {
		return ErrInvalidVerificationResult.Wrap("record_id cannot be empty")
	}

	if r.AccountAddress == "" {
		return ErrInvalidVerificationResult.Wrap("account_address cannot be empty")
	}

	if r.Version == 0 || r.Version > DerivedFeatureVerificationRecordVersion {
		return ErrInvalidVerificationResult.Wrapf("unsupported version: %d", r.Version)
	}

	if r.RequestID == "" {
		return ErrInvalidVerificationResult.Wrap("request_id cannot be empty")
	}

	// Composite hash must be 32 bytes if present
	if len(r.CompositeHash) > 0 && len(r.CompositeHash) != 32 {
		return ErrInvalidVerificationResult.Wrap("composite_hash must be 32 bytes")
	}

	if r.ModelVersion == "" {
		return ErrInvalidVerificationResult.Wrap("model_version cannot be empty")
	}

	if r.Score > MaxScore {
		return ErrInvalidVerificationResult.Wrapf("score %d exceeds maximum %d", r.Score, MaxScore)
	}

	if r.Confidence > 100 {
		return ErrInvalidVerificationResult.Wrapf("confidence %d exceeds 100", r.Confidence)
	}

	if !IsValidVerificationResultStatus(r.Status) {
		return ErrInvalidVerificationResult.Wrapf("invalid status: %s", r.Status)
	}

	if r.ComputedAt.IsZero() {
		return ErrInvalidVerificationResult.Wrap("computed_at cannot be zero")
	}

	if r.BlockHeight < 0 {
		return ErrInvalidVerificationResult.Wrap("block_height cannot be negative")
	}

	// Validate feature references
	for i, ref := range r.FeatureReferences {
		if ref.FeatureType == "" {
			return ErrInvalidVerificationResult.Wrapf("feature_reference[%d] has empty feature_type", i)
		}
		if len(ref.FeatureHash) != 32 {
			return ErrInvalidVerificationResult.Wrapf("feature_reference[%d] hash must be 32 bytes", i)
		}
		if ref.SourceScopeID == "" {
			return ErrInvalidVerificationResult.Wrapf("feature_reference[%d] has empty source_scope_id", i)
		}
		if !IsValidFeatureMatchResult(ref.MatchResult) {
			return ErrInvalidVerificationResult.Wrapf("feature_reference[%d] has invalid match_result", i)
		}
	}

	return nil
}

// AddFeatureReference adds a feature reference to the record
func (r *DerivedFeatureVerificationRecord) AddFeatureReference(ref DerivedFeatureReference) {
	r.FeatureReferences = append(r.FeatureReferences, ref)
}

// ComputeCompositeHash computes the composite hash from all feature references
func (r *DerivedFeatureVerificationRecord) ComputeCompositeHash() {
	h := sha256.New()
	
	// Include all feature hashes in deterministic order
	for _, ref := range r.FeatureReferences {
		h.Write([]byte(ref.FeatureType))
		h.Write(ref.FeatureHash)
		h.Write([]byte(ref.SourceScopeID))
	}
	
	// Include model version for reproducibility
	h.Write([]byte(r.ModelVersion))
	h.Write([]byte(r.ModelHash))
	
	r.CompositeHash = h.Sum(nil)
}

// SetSuccess sets the record as successful
func (r *DerivedFeatureVerificationRecord) SetSuccess(score uint32, confidence uint32) {
	r.Score = score
	r.Confidence = confidence
	r.Status = VerificationResultStatusSuccess
	r.ReasonCodes = []ReasonCode{ReasonCodeSuccess}
}

// SetPartial sets the record as partially successful
func (r *DerivedFeatureVerificationRecord) SetPartial(score uint32, confidence uint32, codes []ReasonCode) {
	r.Score = score
	r.Confidence = confidence
	r.Status = VerificationResultStatusPartial
	r.ReasonCodes = codes
}

// SetFailed sets the record as failed
func (r *DerivedFeatureVerificationRecord) SetFailed(codes ...ReasonCode) {
	r.Score = 0
	r.Confidence = 0
	r.Status = VerificationResultStatusFailed
	r.ReasonCodes = codes
}

// AddConsensusVote adds a consensus vote to the record
func (r *DerivedFeatureVerificationRecord) AddConsensusVote(vote ConsensusVote) {
	r.ConsensusVotes = append(r.ConsensusVotes, vote)
}

// Finalize finalizes the record
func (r *DerivedFeatureVerificationRecord) Finalize(finalizedAt time.Time, blockHeight int64) {
	r.Finalized = true
	r.FinalizedAt = &finalizedAt
	r.FinalizedAtBlock = &blockHeight
}

// CountAgreements counts the number of agreeing votes
func (r *DerivedFeatureVerificationRecord) CountAgreements() int {
	count := 0
	for _, vote := range r.ConsensusVotes {
		if vote.Agreed {
			count++
		}
	}
	return count
}

// CountDisagreements counts the number of disagreeing votes
func (r *DerivedFeatureVerificationRecord) CountDisagreements() int {
	count := 0
	for _, vote := range r.ConsensusVotes {
		if !vote.Agreed {
			count++
		}
	}
	return count
}

// HasConsensus checks if the record has achieved consensus
// Requires 2/3+ agreement from validators
func (r *DerivedFeatureVerificationRecord) HasConsensus(totalValidators int) bool {
	if totalValidators == 0 {
		return false
	}
	
	agreements := r.CountAgreements()
	threshold := (totalValidators * 2) / 3
	
	return agreements > threshold
}

// GetMatchingFeatures returns all features with a match result
func (r *DerivedFeatureVerificationRecord) GetMatchingFeatures() []DerivedFeatureReference {
	var matching []DerivedFeatureReference
	for _, ref := range r.FeatureReferences {
		if ref.MatchResult == FeatureMatchResultMatch {
			matching = append(matching, ref)
		}
	}
	return matching
}

// GetFailedFeatures returns all features that failed to match
func (r *DerivedFeatureVerificationRecord) GetFailedFeatures() []DerivedFeatureReference {
	var failed []DerivedFeatureReference
	for _, ref := range r.FeatureReferences {
		if ref.MatchResult == FeatureMatchResultNoMatch || ref.MatchResult == FeatureMatchResultError {
			failed = append(failed, ref)
		}
	}
	return failed
}

// ComputeWeightedScore computes the weighted score from feature references
func (r *DerivedFeatureVerificationRecord) ComputeWeightedScore() uint32 {
	var totalWeight uint32
	var weightedSum uint32

	for _, ref := range r.FeatureReferences {
		if ref.MatchResult == FeatureMatchResultMatch || ref.MatchResult == FeatureMatchResultPartial {
			weightedSum += ref.MatchScore * ref.Weight
			totalWeight += ref.Weight
		}
	}

	if totalWeight == 0 {
		return 0
	}

	score := weightedSum / totalWeight
	if score > MaxScore {
		score = MaxScore
	}

	return score
}

// NewDerivedFeatureReference creates a new feature reference
func NewDerivedFeatureReference(
	featureType string,
	featureHash []byte,
	sourceScopeID string,
	weight uint32,
) DerivedFeatureReference {
	return DerivedFeatureReference{
		FeatureType:   featureType,
		FeatureHash:   featureHash,
		SourceScopeID: sourceScopeID,
		Weight:        weight,
		MatchResult:   FeatureMatchResultSkipped,
		MatchScore:    0,
	}
}

// SetMatch sets the match result for a feature reference
func (r *DerivedFeatureReference) SetMatch(score uint32) {
	r.MatchResult = FeatureMatchResultMatch
	r.MatchScore = score
}

// SetNoMatch sets a no-match result for a feature reference
func (r *DerivedFeatureReference) SetNoMatch() {
	r.MatchResult = FeatureMatchResultNoMatch
	r.MatchScore = 0
}

// SetPartial sets a partial match result
func (r *DerivedFeatureReference) SetPartial(score uint32) {
	r.MatchResult = FeatureMatchResultPartial
	r.MatchScore = score
}

// SetError sets an error result for a feature reference
func (r *DerivedFeatureReference) SetError() {
	r.MatchResult = FeatureMatchResultError
	r.MatchScore = 0
}
