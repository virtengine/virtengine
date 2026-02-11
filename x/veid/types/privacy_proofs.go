// Package types provides VEID module types.
//
// This file implements privacy-preserving proof types for selective disclosure
// of identity claims. Supports zero-knowledge proofs for proving properties
// about identity attributes without revealing the underlying data.
//
// Task Reference: VE-3029 - Add Privacy-Preserving Proof Types
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// ============================================================================
// Claim Type Definitions
// ============================================================================

// ClaimType represents the type of claim that can be proven
type ClaimType int

const (
	// ClaimTypeAgeOver18 proves the subject is at least 18 years old
	ClaimTypeAgeOver18 ClaimType = iota
	// ClaimTypeAgeOver21 proves the subject is at least 21 years old
	ClaimTypeAgeOver21
	// ClaimTypeAgeOver25 proves the subject is at least 25 years old
	ClaimTypeAgeOver25
	// ClaimTypeCountryResident proves residency in a specific country
	ClaimTypeCountryResident
	// ClaimTypeHumanVerified proves successful human verification
	ClaimTypeHumanVerified
	// ClaimTypeTrustScoreAbove proves trust score is above a threshold
	ClaimTypeTrustScoreAbove
	// ClaimTypeEmailVerified proves email has been verified
	ClaimTypeEmailVerified
	// ClaimTypeSMSVerified proves phone number has been verified
	ClaimTypeSMSVerified
	// ClaimTypeDomainVerified proves domain ownership verification
	ClaimTypeDomainVerified
	// ClaimTypeBiometricVerified proves biometric verification passed
	ClaimTypeBiometricVerified
)

// String returns the string representation of the claim type
func (ct ClaimType) String() string {
	switch ct {
	case ClaimTypeAgeOver18:
		return "age_over_18"
	case ClaimTypeAgeOver21:
		return "age_over_21"
	case ClaimTypeAgeOver25:
		return "age_over_25"
	case ClaimTypeCountryResident:
		return "country_resident"
	case ClaimTypeHumanVerified:
		return "human_verified"
	case ClaimTypeTrustScoreAbove:
		return "trust_score_above"
	case ClaimTypeEmailVerified:
		return "email_verified"
	case ClaimTypeSMSVerified:
		return "sms_verified"
	case ClaimTypeDomainVerified:
		return "domain_verified"
	case ClaimTypeBiometricVerified:
		return "biometric_verified"
	default:
		return string(AccountStatusUnknown)
	}
}

// IsValid checks if the claim type is valid
func (ct ClaimType) IsValid() bool {
	return ct >= ClaimTypeAgeOver18 && ct <= ClaimTypeBiometricVerified
}

// ParseClaimType parses a string into a ClaimType
func ParseClaimType(s string) (ClaimType, error) {
	switch s {
	case "age_over_18":
		return ClaimTypeAgeOver18, nil
	case "age_over_21":
		return ClaimTypeAgeOver21, nil
	case "age_over_25":
		return ClaimTypeAgeOver25, nil
	case "country_resident":
		return ClaimTypeCountryResident, nil
	case "human_verified":
		return ClaimTypeHumanVerified, nil
	case "trust_score_above":
		return ClaimTypeTrustScoreAbove, nil
	case "email_verified":
		return ClaimTypeEmailVerified, nil
	case "sms_verified":
		return ClaimTypeSMSVerified, nil
	case "domain_verified":
		return ClaimTypeDomainVerified, nil
	case "biometric_verified":
		return ClaimTypeBiometricVerified, nil
	default:
		return 0, fmt.Errorf("invalid claim type: %s", s)
	}
}

// ============================================================================
// Proof Scheme Definitions
// ============================================================================

// ProofScheme represents the cryptographic scheme used for the proof
type ProofScheme int

const (
	// ProofSchemeSNARK uses Succinct Non-interactive Arguments of Knowledge
	ProofSchemeSNARK ProofScheme = iota
	// ProofSchemeSTARK uses Scalable Transparent Arguments of Knowledge
	ProofSchemeSTARK
	// ProofSchemeBulletproofs uses Bulletproofs for range proofs
	ProofSchemeBulletproofs
	// ProofSchemeRangeProof uses simple range proofs for numeric comparisons
	ProofSchemeRangeProof
	// ProofSchemeCommitmentScheme uses Pedersen commitments
	ProofSchemeCommitmentScheme
)

// String returns the string representation of the proof scheme
func (ps ProofScheme) String() string {
	switch ps {
	case ProofSchemeSNARK:
		return "snark"
	case ProofSchemeSTARK:
		return "stark"
	case ProofSchemeBulletproofs:
		return "bulletproofs"
	case ProofSchemeRangeProof:
		return "range_proof"
	case ProofSchemeCommitmentScheme:
		return "commitment_scheme"
	default:
		return string(AccountStatusUnknown)
	}
}

// IsValid checks if the proof scheme is valid
func (ps ProofScheme) IsValid() bool {
	return ps >= ProofSchemeSNARK && ps <= ProofSchemeCommitmentScheme
}

// ParseProofScheme parses a string into a ProofScheme
func ParseProofScheme(s string) (ProofScheme, error) {
	switch s {
	case "snark":
		return ProofSchemeSNARK, nil
	case "stark":
		return ProofSchemeSTARK, nil
	case "bulletproofs":
		return ProofSchemeBulletproofs, nil
	case "range_proof":
		return ProofSchemeRangeProof, nil
	case "commitment_scheme":
		return ProofSchemeCommitmentScheme, nil
	default:
		return 0, fmt.Errorf("invalid proof scheme: %s", s)
	}
}

// ============================================================================
// Selective Disclosure Proof Types
// ============================================================================

// SelectiveDisclosureProof represents a zero-knowledge proof for selective disclosure
// of identity claims. The proof allows verifiers to confirm specific claims
// without learning any additional information about the subject.
type SelectiveDisclosureProof struct {
	// ProofID is the unique identifier for this proof
	ProofID string `json:"proof_id"`

	// SubjectAddress is the address of the identity subject
	SubjectAddress string `json:"subject_address"`

	// ClaimTypes lists the types of claims this proof covers
	ClaimTypes []ClaimType `json:"claim_types"`

	// DisclosedClaims contains only the revealed claim values
	// Keys are claim identifiers, values are the disclosed data
	DisclosedClaims map[string]interface{} `json:"disclosed_claims"`

	// CommitmentHash is a cryptographic commitment to the full claim set
	// This allows future binding proofs without revealing full claims
	CommitmentHash []byte `json:"commitment_hash"`

	// ProofValue contains the zero-knowledge proof bytes
	ProofValue []byte `json:"proof_value"`

	// ProofScheme identifies the cryptographic scheme used
	ProofScheme ProofScheme `json:"proof_scheme"`

	// ValidUntil is the expiration time for this proof
	ValidUntil time.Time `json:"valid_until"`

	// Nonce prevents replay attacks
	Nonce []byte `json:"nonce"`

	// CreatedAt is when the proof was generated
	CreatedAt time.Time `json:"created_at"`

	// IssuerAddress is the validator/issuer who generated the proof
	IssuerAddress string `json:"issuer_address,omitempty"`

	// Metadata contains additional proof metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewSelectiveDisclosureProof creates a new selective disclosure proof
func NewSelectiveDisclosureProof(
	proofID string,
	subjectAddress string,
	claimTypes []ClaimType,
	scheme ProofScheme,
	validDuration time.Duration,
) *SelectiveDisclosureProof {
	now := time.Now().UTC()
	return &SelectiveDisclosureProof{
		ProofID:         proofID,
		SubjectAddress:  subjectAddress,
		ClaimTypes:      claimTypes,
		DisclosedClaims: make(map[string]interface{}),
		ProofScheme:     scheme,
		ValidUntil:      now.Add(validDuration),
		CreatedAt:       now,
		Metadata:        make(map[string]string),
	}
}

// Validate validates the selective disclosure proof
func (sdp *SelectiveDisclosureProof) Validate() error {
	if sdp.ProofID == "" {
		return ErrInvalidProof.Wrap("proof_id cannot be empty")
	}

	if sdp.SubjectAddress == "" {
		return ErrInvalidProof.Wrap("subject_address cannot be empty")
	}

	if len(sdp.ClaimTypes) == 0 {
		return ErrInvalidProof.Wrap("claim_types cannot be empty")
	}

	for _, ct := range sdp.ClaimTypes {
		if !ct.IsValid() {
			return ErrInvalidProof.Wrapf("invalid claim type: %d", ct)
		}
	}

	if !sdp.ProofScheme.IsValid() {
		return ErrInvalidProof.Wrapf("invalid proof scheme: %d", sdp.ProofScheme)
	}

	if len(sdp.ProofValue) == 0 {
		return ErrInvalidProof.Wrap("proof_value cannot be empty")
	}

	if len(sdp.CommitmentHash) == 0 {
		return ErrInvalidProof.Wrap("commitment_hash cannot be empty")
	}

	if len(sdp.Nonce) == 0 {
		return ErrInvalidProof.Wrap("nonce cannot be empty")
	}

	if sdp.CreatedAt.IsZero() {
		return ErrInvalidProof.Wrap("created_at cannot be zero")
	}

	if sdp.ValidUntil.IsZero() {
		return ErrInvalidProof.Wrap("valid_until cannot be zero")
	}

	if sdp.ValidUntil.Before(sdp.CreatedAt) {
		return ErrInvalidProof.Wrap("valid_until cannot be before created_at")
	}

	return nil
}

// IsExpired checks if the proof has expired
func (sdp *SelectiveDisclosureProof) IsExpired(now time.Time) bool {
	return now.After(sdp.ValidUntil)
}

// HasClaimType checks if the proof covers a specific claim type
func (sdp *SelectiveDisclosureProof) HasClaimType(ct ClaimType) bool {
	for _, claimType := range sdp.ClaimTypes {
		if claimType == ct {
			return true
		}
	}
	return false
}

// GetProofHash returns a hash of the proof for verification purposes
func (sdp *SelectiveDisclosureProof) GetProofHash() []byte {
	h := sha256.New()
	h.Write([]byte(sdp.ProofID))
	h.Write([]byte(sdp.SubjectAddress))
	h.Write(sdp.CommitmentHash)
	h.Write(sdp.ProofValue)
	h.Write(sdp.Nonce)
	return h.Sum(nil)
}

// ============================================================================
// Selective Disclosure Request
// ============================================================================

// SelectiveDisclosureRequest represents a request for selective disclosure
type SelectiveDisclosureRequest struct {
	// RequestID is the unique identifier for this request
	RequestID string `json:"request_id"`

	// RequesterAddress is the address requesting the disclosure
	RequesterAddress string `json:"requester_address"`

	// SubjectAddress is the address of the identity subject
	SubjectAddress string `json:"subject_address"`

	// RequestedClaims lists the claim types being requested
	RequestedClaims []ClaimType `json:"requested_claims"`

	// ClaimParameters contains additional parameters for claim verification
	// For example, age threshold for age proofs, country code for residency, etc.
	ClaimParameters map[string]interface{} `json:"claim_parameters,omitempty"`

	// PreferredScheme is the preferred proof scheme (optional)
	PreferredScheme *ProofScheme `json:"preferred_scheme,omitempty"`

	// Purpose describes why the disclosure is being requested
	Purpose string `json:"purpose"`

	// ValidityDuration is how long the proof should be valid
	ValidityDuration time.Duration `json:"validity_duration"`

	// CreatedAt is when the request was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the request expires
	ExpiresAt time.Time `json:"expires_at"`

	// Nonce for request uniqueness
	Nonce []byte `json:"nonce"`
}

// NewSelectiveDisclosureRequest creates a new selective disclosure request
func NewSelectiveDisclosureRequest(
	requestID string,
	requesterAddress string,
	subjectAddress string,
	requestedClaims []ClaimType,
	purpose string,
	validityDuration time.Duration,
	requestExpiry time.Duration,
) *SelectiveDisclosureRequest {
	now := time.Now().UTC()
	return &SelectiveDisclosureRequest{
		RequestID:        requestID,
		RequesterAddress: requesterAddress,
		SubjectAddress:   subjectAddress,
		RequestedClaims:  requestedClaims,
		ClaimParameters:  make(map[string]interface{}),
		Purpose:          purpose,
		ValidityDuration: validityDuration,
		CreatedAt:        now,
		ExpiresAt:        now.Add(requestExpiry),
	}
}

// Validate validates the selective disclosure request
func (sdr *SelectiveDisclosureRequest) Validate() error {
	if sdr.RequestID == "" {
		return ErrInvalidProofRequest.Wrap("request_id cannot be empty")
	}

	if sdr.RequesterAddress == "" {
		return ErrInvalidProofRequest.Wrap("requester_address cannot be empty")
	}

	if sdr.SubjectAddress == "" {
		return ErrInvalidProofRequest.Wrap("subject_address cannot be empty")
	}

	if len(sdr.RequestedClaims) == 0 {
		return ErrInvalidProofRequest.Wrap("requested_claims cannot be empty")
	}

	for _, ct := range sdr.RequestedClaims {
		if !ct.IsValid() {
			return ErrInvalidProofRequest.Wrapf("invalid claim type: %d", ct)
		}
	}

	if sdr.Purpose == "" {
		return ErrInvalidProofRequest.Wrap("purpose cannot be empty")
	}

	if sdr.ValidityDuration <= 0 {
		return ErrInvalidProofRequest.Wrap("validity_duration must be positive")
	}

	if sdr.CreatedAt.IsZero() {
		return ErrInvalidProofRequest.Wrap("created_at cannot be zero")
	}

	if sdr.ExpiresAt.IsZero() || sdr.ExpiresAt.Before(sdr.CreatedAt) {
		return ErrInvalidProofRequest.Wrap("expires_at must be after created_at")
	}

	return nil
}

// IsExpired checks if the request has expired
func (sdr *SelectiveDisclosureRequest) IsExpired(now time.Time) bool {
	return now.After(sdr.ExpiresAt)
}

// ============================================================================
// Specialized Proof Types
// ============================================================================

// AgeProof represents a zero-knowledge proof of age without revealing DOB
type AgeProof struct {
	// ProofID is the unique identifier for this proof
	ProofID string `json:"proof_id"`

	// SubjectAddress is the address of the identity subject
	SubjectAddress string `json:"subject_address"`

	// AgeThreshold is the minimum age being proven (e.g., 18, 21, 25)
	AgeThreshold uint32 `json:"age_threshold"`

	// SatisfiesThreshold indicates whether the subject meets the threshold
	SatisfiesThreshold bool `json:"satisfies_threshold"`

	// CommitmentHash is a commitment to the actual date of birth
	CommitmentHash []byte `json:"commitment_hash"`

	// ProofValue contains the zero-knowledge proof bytes
	ProofValue []byte `json:"proof_value"`

	// ProofScheme identifies the cryptographic scheme used
	ProofScheme ProofScheme `json:"proof_scheme"`

	// ValidUntil is the expiration time for this proof
	ValidUntil time.Time `json:"valid_until"`

	// Nonce prevents replay attacks
	Nonce []byte `json:"nonce"`

	// CreatedAt is when the proof was generated
	CreatedAt time.Time `json:"created_at"`
}

// NewAgeProof creates a new age proof
func NewAgeProof(
	proofID string,
	subjectAddress string,
	ageThreshold uint32,
	validDuration time.Duration,
) *AgeProof {
	now := time.Now().UTC()
	return &AgeProof{
		ProofID:        proofID,
		SubjectAddress: subjectAddress,
		AgeThreshold:   ageThreshold,
		ProofScheme:    ProofSchemeRangeProof,
		ValidUntil:     now.Add(validDuration),
		CreatedAt:      now,
	}
}

// Validate validates the age proof
func (ap *AgeProof) Validate() error {
	if ap.ProofID == "" {
		return ErrInvalidProof.Wrap("proof_id cannot be empty")
	}

	if ap.SubjectAddress == "" {
		return ErrInvalidProof.Wrap("subject_address cannot be empty")
	}

	if ap.AgeThreshold == 0 {
		return ErrInvalidProof.Wrap("age_threshold must be positive")
	}

	if !ap.ProofScheme.IsValid() {
		return ErrInvalidProof.Wrapf("invalid proof scheme: %d", ap.ProofScheme)
	}

	if len(ap.ProofValue) == 0 {
		return ErrInvalidProof.Wrap("proof_value cannot be empty")
	}

	if len(ap.CommitmentHash) == 0 {
		return ErrInvalidProof.Wrap("commitment_hash cannot be empty")
	}

	if len(ap.Nonce) == 0 {
		return ErrInvalidProof.Wrap("nonce cannot be empty")
	}

	return nil
}

// IsExpired checks if the proof has expired
func (ap *AgeProof) IsExpired(now time.Time) bool {
	return now.After(ap.ValidUntil)
}

// ResidencyProof represents a zero-knowledge proof of residency without revealing address
type ResidencyProof struct {
	// ProofID is the unique identifier for this proof
	ProofID string `json:"proof_id"`

	// SubjectAddress is the address of the identity subject
	SubjectAddress string `json:"subject_address"`

	// CountryCode is the ISO 3166-1 alpha-2 country code being proven
	CountryCode string `json:"country_code"`

	// IsResident indicates whether the subject is a resident of the country
	IsResident bool `json:"is_resident"`

	// CommitmentHash is a commitment to the actual address
	CommitmentHash []byte `json:"commitment_hash"`

	// ProofValue contains the zero-knowledge proof bytes
	ProofValue []byte `json:"proof_value"`

	// ProofScheme identifies the cryptographic scheme used
	ProofScheme ProofScheme `json:"proof_scheme"`

	// ValidUntil is the expiration time for this proof
	ValidUntil time.Time `json:"valid_until"`

	// Nonce prevents replay attacks
	Nonce []byte `json:"nonce"`

	// CreatedAt is when the proof was generated
	CreatedAt time.Time `json:"created_at"`
}

// NewResidencyProof creates a new residency proof
func NewResidencyProof(
	proofID string,
	subjectAddress string,
	countryCode string,
	validDuration time.Duration,
) *ResidencyProof {
	now := time.Now().UTC()
	return &ResidencyProof{
		ProofID:        proofID,
		SubjectAddress: subjectAddress,
		CountryCode:    countryCode,
		ProofScheme:    ProofSchemeCommitmentScheme,
		ValidUntil:     now.Add(validDuration),
		CreatedAt:      now,
	}
}

// Validate validates the residency proof
func (rp *ResidencyProof) Validate() error {
	if rp.ProofID == "" {
		return ErrInvalidProof.Wrap("proof_id cannot be empty")
	}

	if rp.SubjectAddress == "" {
		return ErrInvalidProof.Wrap("subject_address cannot be empty")
	}

	if rp.CountryCode == "" {
		return ErrInvalidProof.Wrap("country_code cannot be empty")
	}

	if len(rp.CountryCode) != 2 {
		return ErrInvalidProof.Wrap("country_code must be ISO 3166-1 alpha-2 format")
	}

	if !rp.ProofScheme.IsValid() {
		return ErrInvalidProof.Wrapf("invalid proof scheme: %d", rp.ProofScheme)
	}

	if len(rp.ProofValue) == 0 {
		return ErrInvalidProof.Wrap("proof_value cannot be empty")
	}

	if len(rp.CommitmentHash) == 0 {
		return ErrInvalidProof.Wrap("commitment_hash cannot be empty")
	}

	if len(rp.Nonce) == 0 {
		return ErrInvalidProof.Wrap("nonce cannot be empty")
	}

	return nil
}

// IsExpired checks if the proof has expired
func (rp *ResidencyProof) IsExpired(now time.Time) bool {
	return now.After(rp.ValidUntil)
}

// ScoreThresholdProof represents a zero-knowledge proof that trust score exceeds a threshold
type ScoreThresholdProof struct {
	// ProofID is the unique identifier for this proof
	ProofID string `json:"proof_id"`

	// SubjectAddress is the address of the identity subject
	SubjectAddress string `json:"subject_address"`

	// ScoreThreshold is the minimum score being proven
	ScoreThreshold uint32 `json:"score_threshold"`

	// ExceedsThreshold indicates whether the subject's score exceeds the threshold
	ExceedsThreshold bool `json:"exceeds_threshold"`

	// CommitmentHash is a commitment to the actual score
	CommitmentHash []byte `json:"commitment_hash"`

	// ProofValue contains the zero-knowledge proof bytes
	ProofValue []byte `json:"proof_value"`

	// ProofScheme identifies the cryptographic scheme used
	ProofScheme ProofScheme `json:"proof_scheme"`

	// ValidUntil is the expiration time for this proof
	ValidUntil time.Time `json:"valid_until"`

	// Nonce prevents replay attacks
	Nonce []byte `json:"nonce"`

	// CreatedAt is when the proof was generated
	CreatedAt time.Time `json:"created_at"`

	// ScoreVersion is the version of the scoring model used
	ScoreVersion string `json:"score_version,omitempty"`
}

// NewScoreThresholdProof creates a new score threshold proof
func NewScoreThresholdProof(
	proofID string,
	subjectAddress string,
	scoreThreshold uint32,
	validDuration time.Duration,
) *ScoreThresholdProof {
	now := time.Now().UTC()
	return &ScoreThresholdProof{
		ProofID:        proofID,
		SubjectAddress: subjectAddress,
		ScoreThreshold: scoreThreshold,
		ProofScheme:    ProofSchemeBulletproofs,
		ValidUntil:     now.Add(validDuration),
		CreatedAt:      now,
	}
}

// Validate validates the score threshold proof
func (stp *ScoreThresholdProof) Validate() error {
	if stp.ProofID == "" {
		return ErrInvalidProof.Wrap("proof_id cannot be empty")
	}

	if stp.SubjectAddress == "" {
		return ErrInvalidProof.Wrap("subject_address cannot be empty")
	}

	if stp.ScoreThreshold == 0 {
		return ErrInvalidProof.Wrap("score_threshold must be positive")
	}

	if stp.ScoreThreshold > 100 {
		return ErrInvalidProof.Wrap("score_threshold cannot exceed 100")
	}

	if !stp.ProofScheme.IsValid() {
		return ErrInvalidProof.Wrapf("invalid proof scheme: %d", stp.ProofScheme)
	}

	if len(stp.ProofValue) == 0 {
		return ErrInvalidProof.Wrap("proof_value cannot be empty")
	}

	if len(stp.CommitmentHash) == 0 {
		return ErrInvalidProof.Wrap("commitment_hash cannot be empty")
	}

	if len(stp.Nonce) == 0 {
		return ErrInvalidProof.Wrap("nonce cannot be empty")
	}

	return nil
}

// IsExpired checks if the proof has expired
func (stp *ScoreThresholdProof) IsExpired(now time.Time) bool {
	return now.After(stp.ValidUntil)
}

// ============================================================================
// Proof Verification Result
// ============================================================================

// ProofVerificationResult represents the result of verifying a privacy-preserving proof
type ProofVerificationResult struct {
	// IsValid indicates whether the proof is valid
	IsValid bool `json:"is_valid"`

	// ClaimsVerified lists which claim types were successfully verified
	ClaimsVerified []ClaimType `json:"claims_verified"`

	// VerifiedAt is when verification occurred
	VerifiedAt time.Time `json:"verified_at"`

	// VerifierAddress is the address that performed verification
	VerifierAddress string `json:"verifier_address"`

	// Error contains any error message if verification failed
	Error string `json:"error,omitempty"`

	// ProofHash is a hash of the verified proof for audit purposes
	ProofHash string `json:"proof_hash"`
}

// NewProofVerificationResult creates a new proof verification result
func NewProofVerificationResult(isValid bool, claimsVerified []ClaimType, verifierAddress string) *ProofVerificationResult {
	return &ProofVerificationResult{
		IsValid:         isValid,
		ClaimsVerified:  claimsVerified,
		VerifiedAt:      time.Now().UTC(),
		VerifierAddress: verifierAddress,
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// GenerateProofID generates a unique proof ID from components
func GenerateProofID(subjectAddress string, claimTypes []ClaimType, nonce []byte) string {
	h := sha256.New()
	h.Write([]byte(subjectAddress))
	for _, ct := range claimTypes {
		h.Write([]byte(ct.String()))
	}
	h.Write(nonce)
	return "proof_" + hex.EncodeToString(h.Sum(nil)[:16])
}

// GenerateRequestID generates a unique request ID from components
func GenerateRequestID(requesterAddress, subjectAddress string, nonce []byte) string {
	h := sha256.New()
	h.Write([]byte(requesterAddress))
	h.Write([]byte(subjectAddress))
	h.Write(nonce)
	return "req_" + hex.EncodeToString(h.Sum(nil)[:16])
}

// ComputeCommitmentHash computes a cryptographic commitment for hidden values.
// Implemented in proof_types.go using Pedersen commitments.
