package types

import (
	"crypto/sha256"
	"encoding/hex"

	v1 "github.com/virtengine/virtengine/sdk/go/node/enclave/v1"
)

// ValidateAttestedScoringResult validates the attested scoring result
func ValidateAttestedScoringResult(a *v1.AttestedScoringResult) error {
	if a.ScopeId == "" {
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
func SigningPayload(a *v1.AttestedScoringResult) []byte {
	h := sha256.New()

	// Include scope ID
	h.Write([]byte(a.ScopeId))

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

// AttestedResultHash returns a unique identifier for this result
func AttestedResultHash(a *v1.AttestedScoringResult) []byte {
	h := sha256.New()
	h.Write(SigningPayload(a))
	h.Write(a.EnclaveSignature)
	return h.Sum(nil)
}

// MeasurementHashHex returns the enclave measurement hash as hex string
func AttestedResultMeasurementHashHex(a *v1.AttestedScoringResult) string {
	return hex.EncodeToString(a.EnclaveMeasurementHash)
}

// ConsensusVerificationRequest represents a request to verify a scoring result
// during consensus. This type is kept locally as it's not in the proto.
type ConsensusVerificationRequest struct {
	// ProposedResult is the result proposed by the block proposer
	ProposedResult *v1.AttestedScoringResult `json:"proposed_result"`

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
	RecomputedResult *v1.AttestedScoringResult `json:"recomputed_result,omitempty"`

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
