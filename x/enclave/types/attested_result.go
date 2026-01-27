package types

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// AttestedScoringResult represents an enclave-attested scoring output
// that is included in blocks for consensus verification.
type AttestedScoringResult struct {
	// ScopeID is the identity scope that was scored
	ScopeID string `json:"scope_id"`

	// AccountAddress is the account that owns the identity
	AccountAddress string `json:"account_address"`

	// Score is the computed identity score (0-100)
	Score uint32 `json:"score"`

	// Status is the verification status
	Status string `json:"status"`

	// ReasonCodes are structured reason codes for the score
	ReasonCodes []string `json:"reason_codes,omitempty"`

	// ModelVersionHash is the hash of the ML model used
	ModelVersionHash []byte `json:"model_version_hash"`

	// InputHash is the hash of the input data (for determinism verification)
	InputHash []byte `json:"input_hash"`

	// EvidenceHashes are hashes of evidence artifacts (face embeddings, OCR, etc.)
	EvidenceHashes [][]byte `json:"evidence_hashes,omitempty"`

	// EnclaveMeasurementHash is the measurement of the enclave that computed this
	EnclaveMeasurementHash []byte `json:"enclave_measurement_hash"`

	// EnclaveSignature is the signature from the enclave signing key
	EnclaveSignature []byte `json:"enclave_signature"`

	// AttestationReference is a reference to the attestation quote (hash or ID)
	AttestationReference []byte `json:"attestation_reference"`

	// ValidatorAddress is the validator that produced this result
	ValidatorAddress string `json:"validator_address"`

	// BlockHeight is the block height where this result was produced
	BlockHeight int64 `json:"block_height"`

	// Timestamp is when this result was computed
	Timestamp time.Time `json:"timestamp"`
}

// Validate validates the attested scoring result
func (a *AttestedScoringResult) Validate() error {
	if a.ScopeID == "" {
		return ErrInvalidAttestedResult.Wrap("scope ID cannot be empty")
	}

	if a.AccountAddress == "" {
		return ErrInvalidAttestedResult.Wrap("account address cannot be empty")
	}

	if a.Score > 100 {
		return ErrInvalidAttestedResult.Wrapf("score must be 0-100, got %d", a.Score)
	}

	if len(a.ModelVersionHash) == 0 {
		return ErrInvalidAttestedResult.Wrap("model version hash cannot be empty")
	}

	if len(a.InputHash) == 0 {
		return ErrInvalidAttestedResult.Wrap("input hash cannot be empty")
	}

	if len(a.EnclaveMeasurementHash) == 0 {
		return ErrInvalidAttestedResult.Wrap("enclave measurement hash cannot be empty")
	}

	if len(a.EnclaveSignature) == 0 {
		return ErrInvalidAttestedResult.Wrap("enclave signature cannot be empty")
	}

	if len(a.AttestationReference) == 0 {
		return ErrInvalidAttestedResult.Wrap("attestation reference cannot be empty")
	}

	if a.ValidatorAddress == "" {
		return ErrInvalidAttestedResult.Wrap("validator address cannot be empty")
	}

	if a.BlockHeight <= 0 {
		return ErrInvalidAttestedResult.Wrap("block height must be positive")
	}

	return nil
}

// SigningPayload returns the bytes that should be signed by the enclave
func (a *AttestedScoringResult) SigningPayload() []byte {
	h := sha256.New()

	// Include scope ID
	h.Write([]byte(a.ScopeID))

	// Include account address
	h.Write([]byte(a.AccountAddress))

	// Include score
	h.Write([]byte{byte(a.Score >> 24), byte(a.Score >> 16), byte(a.Score >> 8), byte(a.Score)})

	// Include status
	h.Write([]byte(a.Status))

	// Include model version hash
	h.Write(a.ModelVersionHash)

	// Include input hash
	h.Write(a.InputHash)

	// Include all evidence hashes
	for _, eh := range a.EvidenceHashes {
		h.Write(eh)
	}

	// Include enclave measurement hash
	h.Write(a.EnclaveMeasurementHash)

	// Include block height
	h.Write([]byte{
		byte(a.BlockHeight >> 56), byte(a.BlockHeight >> 48),
		byte(a.BlockHeight >> 40), byte(a.BlockHeight >> 32),
		byte(a.BlockHeight >> 24), byte(a.BlockHeight >> 16),
		byte(a.BlockHeight >> 8), byte(a.BlockHeight),
	})

	return h.Sum(nil)
}

// Hash returns a unique identifier for this result
func (a *AttestedScoringResult) Hash() []byte {
	h := sha256.New()
	h.Write(a.SigningPayload())
	h.Write(a.EnclaveSignature)
	return h.Sum(nil)
}

// MeasurementHashHex returns the enclave measurement hash as hex string
func (a *AttestedScoringResult) MeasurementHashHex() string {
	return hex.EncodeToString(a.EnclaveMeasurementHash)
}

// ConsensusVerificationRequest represents a request to verify a scoring result
// during consensus.
type ConsensusVerificationRequest struct {
	// ProposedResult is the result proposed by the block proposer
	ProposedResult *AttestedScoringResult `json:"proposed_result"`

	// EncryptedPayload is the encrypted identity data to recompute
	EncryptedPayload []byte `json:"encrypted_payload"`

	// WrappedKey is the wrapped decryption key for this validator
	WrappedKey []byte `json:"wrapped_key"`

	// BlockHeight is the block being proposed
	BlockHeight int64 `json:"block_height"`
}

// ConsensusVerificationResponse represents the result of consensus verification
type ConsensusVerificationResponse struct {
	// Valid indicates whether the proposed result is valid
	Valid bool `json:"valid"`

	// RecomputedResult is the result from local enclave recomputation
	RecomputedResult *AttestedScoringResult `json:"recomputed_result,omitempty"`

	// Reason is the reason for rejection (if not valid)
	Reason string `json:"reason,omitempty"`

	// ScoreDifference is the difference between proposed and recomputed scores
	ScoreDifference int32 `json:"score_difference,omitempty"`
}

// VoteEligibilityCheck represents the result of checking vote eligibility
type VoteEligibilityCheck struct {
	// Eligible indicates whether the validator is eligible to vote
	Eligible bool `json:"eligible"`

	// Reason is the reason for ineligibility (if not eligible)
	Reason string `json:"reason,omitempty"`

	// EnclaveAvailable indicates if the enclave runtime is available
	EnclaveAvailable bool `json:"enclave_available"`

	// MeasurementAllowlisted indicates if the enclave measurement is allowlisted
	MeasurementAllowlisted bool `json:"measurement_allowlisted"`

	// AttestationValid indicates if the attestation is valid
	AttestationValid bool `json:"attestation_valid"`

	// KeyAvailable indicates if the wrapped key is available for decryption
	KeyAvailable bool `json:"key_available"`
}
