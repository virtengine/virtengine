package types

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// ============================================================================
// Verification Result Status
// ============================================================================

// VerificationResultStatus represents the outcome status of a verification
type VerificationResultStatus string

const (
	// VerificationResultStatusSuccess indicates successful verification
	VerificationResultStatusSuccess VerificationResultStatus = "success"

	// VerificationResultStatusPartial indicates partial verification (some scopes failed)
	VerificationResultStatusPartial VerificationResultStatus = "partial"

	// VerificationResultStatusFailed indicates verification failed
	VerificationResultStatusFailed VerificationResultStatus = "failed"

	// VerificationResultStatusError indicates an error occurred during verification
	VerificationResultStatusError VerificationResultStatus = "error"
)

// AllVerificationResultStatuses returns all valid result statuses
func AllVerificationResultStatuses() []VerificationResultStatus {
	return []VerificationResultStatus{
		VerificationResultStatusSuccess,
		VerificationResultStatusPartial,
		VerificationResultStatusFailed,
		VerificationResultStatusError,
	}
}

// IsValidVerificationResultStatus checks if a status is valid
func IsValidVerificationResultStatus(status VerificationResultStatus) bool {
	for _, s := range AllVerificationResultStatuses() {
		if s == status {
			return true
		}
	}
	return false
}

// ============================================================================
// Reason Codes
// ============================================================================

// ReasonCode represents a machine-readable code for verification outcomes
type ReasonCode string

const (
	// ReasonCodeSuccess indicates successful verification
	ReasonCodeSuccess ReasonCode = "SUCCESS"

	// ReasonCodeDecryptError indicates decryption failed
	ReasonCodeDecryptError ReasonCode = "DECRYPT_ERROR"

	// ReasonCodeInvalidScope indicates scope validation failed
	ReasonCodeInvalidScope ReasonCode = "INVALID_SCOPE"

	// ReasonCodeScopeNotFound indicates a requested scope was not found
	ReasonCodeScopeNotFound ReasonCode = "SCOPE_NOT_FOUND"

	// ReasonCodeScopeRevoked indicates a scope has been revoked
	ReasonCodeScopeRevoked ReasonCode = "SCOPE_REVOKED"

	// ReasonCodeScopeExpired indicates a scope has expired
	ReasonCodeScopeExpired ReasonCode = "SCOPE_EXPIRED"

	// ReasonCodeMLInferenceError indicates ML scoring failed
	ReasonCodeMLInferenceError ReasonCode = "ML_INFERENCE_ERROR"

	// ReasonCodeTimeout indicates processing timed out
	ReasonCodeTimeout ReasonCode = "TIMEOUT"

	// ReasonCodeMaxRetriesExceeded indicates max retries were exceeded
	ReasonCodeMaxRetriesExceeded ReasonCode = "MAX_RETRIES_EXCEEDED"

	// ReasonCodeInvalidPayload indicates payload was malformed
	ReasonCodeInvalidPayload ReasonCode = "INVALID_PAYLOAD"

	// ReasonCodeKeyNotFound indicates validator key was not found
	ReasonCodeKeyNotFound ReasonCode = "KEY_NOT_FOUND"

	// ReasonCodeInsufficientScopes indicates not enough valid scopes for scoring
	ReasonCodeInsufficientScopes ReasonCode = "INSUFFICIENT_SCOPES"

	// ReasonCodeFaceMismatch indicates facial verification failed
	ReasonCodeFaceMismatch ReasonCode = "FACE_MISMATCH"

	// ReasonCodeDocumentInvalid indicates document validation failed
	ReasonCodeDocumentInvalid ReasonCode = "DOCUMENT_INVALID"

	// ReasonCodeLivenessCheckFailed indicates liveness check failed
	ReasonCodeLivenessCheckFailed ReasonCode = "LIVENESS_CHECK_FAILED"

	// ReasonCodeLowConfidence indicates ML model has low confidence in prediction
	ReasonCodeLowConfidence ReasonCode = "LOW_CONFIDENCE"

	// ReasonCodeLowDocQuality indicates document quality is below threshold
	ReasonCodeLowDocQuality ReasonCode = "LOW_DOC_QUALITY"

	// ReasonCodeLowOCRConfidence indicates OCR extraction confidence is low
	ReasonCodeLowOCRConfidence ReasonCode = "LOW_OCR_CONFIDENCE"
)

// ============================================================================
// Scope Verification Result
// ============================================================================

// ScopeVerificationResult represents the verification result for a single scope
type ScopeVerificationResult struct {
	// ScopeID is the scope that was verified
	ScopeID string `json:"scope_id"`

	// ScopeType is the type of the scope
	ScopeType ScopeType `json:"scope_type"`

	// Success indicates if the scope was successfully verified
	Success bool `json:"success"`

	// Score is the individual scope score contribution (0-100)
	Score uint32 `json:"score"`

	// Weight is the weight of this scope type in overall score
	Weight uint32 `json:"weight"`

	// ReasonCodes contains codes explaining the outcome
	ReasonCodes []ReasonCode `json:"reason_codes,omitempty"`

	// Details contains human-readable details (for debugging)
	Details string `json:"details,omitempty"`
}

// NewScopeVerificationResult creates a new scope verification result
func NewScopeVerificationResult(scopeID string, scopeType ScopeType) *ScopeVerificationResult {
	return &ScopeVerificationResult{
		ScopeID:     scopeID,
		ScopeType:   scopeType,
		Success:     false,
		Score:       0,
		Weight:      ScopeTypeWeight(scopeType),
		ReasonCodes: make([]ReasonCode, 0),
	}
}

// SetSuccess marks the scope as successfully verified with a score
func (r *ScopeVerificationResult) SetSuccess(score uint32) {
	r.Success = true
	r.Score = score
	r.ReasonCodes = []ReasonCode{ReasonCodeSuccess}
}

// SetFailure marks the scope as failed with reason codes
func (r *ScopeVerificationResult) SetFailure(codes ...ReasonCode) {
	r.Success = false
	r.Score = 0
	r.ReasonCodes = codes
}

// WeightedScore returns the weighted contribution of this scope
func (r *ScopeVerificationResult) WeightedScore() uint32 {
	if !r.Success {
		return 0
	}
	// Weighted score = (score * weight) / 100
	return (r.Score * r.Weight) / 100
}

// ============================================================================
// Verification Result
// ============================================================================

// VerificationResult represents the complete result of a verification request
type VerificationResult struct {
	// RequestID is the request this result is for
	RequestID string `json:"request_id"`

	// AccountAddress is the account that was verified
	AccountAddress string `json:"account_address"`

	// Score is the computed identity score (0-100)
	Score uint32 `json:"score"`

	// Status is the overall verification result status
	Status VerificationResultStatus `json:"status"`

	// ModelVersion is the version of the ML model used
	ModelVersion string `json:"model_version"`

	// ComputedAt is when the verification was computed
	ComputedAt time.Time `json:"computed_at"`

	// BlockHeight is the block at which this result was computed
	BlockHeight int64 `json:"block_height"`

	// ReasonCodes contains overall verification reason codes
	ReasonCodes []ReasonCode `json:"reason_codes,omitempty"`

	// InputHash is the SHA256 hash of all decrypted inputs for consensus verification
	InputHash []byte `json:"input_hash"`

	// ScopeResults contains individual scope verification results
	ScopeResults []ScopeVerificationResult `json:"scope_results,omitempty"`

	// ValidatorAddress is the validator that computed this result
	ValidatorAddress string `json:"validator_address,omitempty"`

	// ProcessingDuration is how long verification took in milliseconds
	ProcessingDuration int64 `json:"processing_duration_ms"`

	// Metadata contains additional result-specific data
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewVerificationResult creates a new verification result
func NewVerificationResult(
	requestID string,
	accountAddress string,
	computedAt time.Time,
	blockHeight int64,
) *VerificationResult {
	return &VerificationResult{
		RequestID:      requestID,
		AccountAddress: accountAddress,
		Score:          0,
		Status:         VerificationResultStatusFailed,
		ModelVersion:   "",
		ComputedAt:     computedAt,
		BlockHeight:    blockHeight,
		ReasonCodes:    make([]ReasonCode, 0),
		InputHash:      make([]byte, 0),
		ScopeResults:   make([]ScopeVerificationResult, 0),
		Metadata:       make(map[string]string),
	}
}

// Validate validates the verification result
func (r *VerificationResult) Validate() error {
	if r.RequestID == "" {
		return ErrInvalidVerificationResult.Wrap("request_id cannot be empty")
	}

	if r.AccountAddress == "" {
		return ErrInvalidVerificationResult.Wrap("account_address cannot be empty")
	}

	if r.Score > MaxScore {
		return ErrInvalidVerificationResult.Wrapf("score %d exceeds maximum %d", r.Score, MaxScore)
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

	return nil
}

// SetSuccess sets the result as successful with the computed score
func (r *VerificationResult) SetSuccess(score uint32, modelVersion string, inputHash []byte) {
	r.Score = score
	r.Status = VerificationResultStatusSuccess
	r.ModelVersion = modelVersion
	r.InputHash = inputHash
	r.ReasonCodes = []ReasonCode{ReasonCodeSuccess}
}

// SetPartial sets the result as partially successful
func (r *VerificationResult) SetPartial(score uint32, modelVersion string, inputHash []byte, codes []ReasonCode) {
	r.Score = score
	r.Status = VerificationResultStatusPartial
	r.ModelVersion = modelVersion
	r.InputHash = inputHash
	r.ReasonCodes = codes
}

// SetFailed sets the result as failed
func (r *VerificationResult) SetFailed(codes ...ReasonCode) {
	r.Score = 0
	r.Status = VerificationResultStatusFailed
	r.ReasonCodes = codes
}

// SetError sets the result as errored
func (r *VerificationResult) SetError(code ReasonCode, details string) {
	r.Score = 0
	r.Status = VerificationResultStatusError
	r.ReasonCodes = []ReasonCode{code}
	r.Metadata["error_details"] = details
}

// AddScopeResult adds a scope verification result
func (r *VerificationResult) AddScopeResult(scopeResult ScopeVerificationResult) {
	r.ScopeResults = append(r.ScopeResults, scopeResult)
}

// ComputeOverallScore computes the overall score from scope results
// Uses weighted average based on scope type weights
func (r *VerificationResult) ComputeOverallScore() uint32 {
	if len(r.ScopeResults) == 0 {
		return 0
	}

	var totalWeight uint32
	var weightedSum uint32

	for _, sr := range r.ScopeResults {
		if sr.Success {
			weightedSum += sr.Score * sr.Weight
			totalWeight += sr.Weight
		}
	}

	if totalWeight == 0 {
		return 0
	}

	// Compute weighted average, capped at MaxScore
	score := weightedSum / totalWeight
	if score > MaxScore {
		score = MaxScore
	}

	return score
}

// CountSuccessfulScopes returns the number of successfully verified scopes
func (r *VerificationResult) CountSuccessfulScopes() int {
	count := 0
	for _, sr := range r.ScopeResults {
		if sr.Success {
			count++
		}
	}
	return count
}

// DetermineStatus determines the result status based on scope results
func (r *VerificationResult) DetermineStatus() {
	total := len(r.ScopeResults)
	if total == 0 {
		r.Status = VerificationResultStatusFailed
		r.ReasonCodes = []ReasonCode{ReasonCodeInsufficientScopes}
		return
	}

	successful := r.CountSuccessfulScopes()

	switch successful {
	case 0:
		r.Status = VerificationResultStatusFailed
		// Collect reason codes from failed scopes
		var codes []ReasonCode
		for _, sr := range r.ScopeResults {
			codes = append(codes, sr.ReasonCodes...)
		}
		r.ReasonCodes = codes
	case total:
		r.Status = VerificationResultStatusSuccess
		r.ReasonCodes = []ReasonCode{ReasonCodeSuccess}
	default:
		r.Status = VerificationResultStatusPartial
		// Include failure reason codes
		var codes []ReasonCode
		for _, sr := range r.ScopeResults {
			if !sr.Success {
				codes = append(codes, sr.ReasonCodes...)
			}
		}
		r.ReasonCodes = codes
	}
}

// ComputeInputHash computes SHA256 hash of scope contents for consensus verification
func ComputeInputHash(scopeContents [][]byte) []byte {
	h := sha256.New()
	for _, content := range scopeContents {
		h.Write(content)
	}
	return h.Sum(nil)
}

// String returns a string representation of the result
func (r *VerificationResult) String() string {
	return fmt.Sprintf(
		"VerificationResult{RequestID: %s, Account: %s, Score: %d, Status: %s, Scopes: %d/%d}",
		r.RequestID, r.AccountAddress, r.Score, r.Status,
		r.CountSuccessfulScopes(), len(r.ScopeResults))
}

// ============================================================================
// Store Keys
// ============================================================================

var (
	// PrefixVerificationResult is the prefix for verification result storage
	// Key: PrefixVerificationResult | request_id -> VerificationResult
	PrefixVerificationResult = []byte{0x13}

	// PrefixVerificationResultByAccount is the prefix for lookup by account
	// Key: PrefixVerificationResultByAccount | account_address | block_height -> request_id
	PrefixVerificationResultByAccount = []byte{0x14}
)

// VerificationResultKey returns the store key for a verification result
func VerificationResultKey(requestID string) []byte {
	key := make([]byte, 0, len(PrefixVerificationResult)+len(requestID))
	key = append(key, PrefixVerificationResult...)
	key = append(key, []byte(requestID)...)
	return key
}

// VerificationResultByAccountKey returns the key for results by account
func VerificationResultByAccountKey(accountAddress string, blockHeight int64) []byte {
	heightBytes := encodeInt64(blockHeight)
	key := make([]byte, 0, len(PrefixVerificationResultByAccount)+len(accountAddress)+1+8)
	key = append(key, PrefixVerificationResultByAccount...)
	key = append(key, []byte(accountAddress)...)
	key = append(key, byte('/'))
	key = append(key, heightBytes...)
	return key
}
