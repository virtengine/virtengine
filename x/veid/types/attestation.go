// Package types provides VEID module types.
//
// This file defines verification attestation schema and signer requirements
// for VEID verification services.
//
// Task Reference: VE-1B - Verification Attestation Schema
package types

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// ============================================================================
// Attestation Schema Version
// ============================================================================

const (
	// AttestationSchemaVersion is the current schema version for attestations
	AttestationSchemaVersion = "1.0.0"

	// AttestationSchemaVersionMajor is the major version number
	AttestationSchemaVersionMajor = 1

	// AttestationSchemaVersionMinor is the minor version number
	AttestationSchemaVersionMinor = 0

	// AttestationSchemaVersionPatch is the patch version number
	AttestationSchemaVersionPatch = 0
)

// ============================================================================
// Attestation Types
// ============================================================================

// AttestationType identifies what kind of verification the attestation represents
type AttestationType string

const (
	// AttestationTypeFacialVerification for facial recognition attestations
	AttestationTypeFacialVerification AttestationType = "facial_verification"

	// AttestationTypeLivenessCheck for liveness detection attestations
	AttestationTypeLivenessCheck AttestationType = "liveness_check"

	// AttestationTypeDocumentVerification for ID document verification attestations
	AttestationTypeDocumentVerification AttestationType = "document_verification"

	// AttestationTypeEmailVerification for email verification attestations
	AttestationTypeEmailVerification AttestationType = "email_verification"

	// AttestationTypeSMSVerification for SMS/phone verification attestations
	AttestationTypeSMSVerification AttestationType = "sms_verification"

	// AttestationTypeDomainVerification for domain ownership verification
	AttestationTypeDomainVerification AttestationType = "domain_verification"

	// AttestationTypeSSOVerification for SSO provider verification
	AttestationTypeSSOVerification AttestationType = "sso_verification"

	// AttestationTypeBiometricVerification for biometric verification
	AttestationTypeBiometricVerification AttestationType = "biometric_verification"

	// AttestationTypeCompositeIdentity for combined identity attestations
	AttestationTypeCompositeIdentity AttestationType = "composite_identity"
)

// AllAttestationTypes returns all valid attestation types
func AllAttestationTypes() []AttestationType {
	return []AttestationType{
		AttestationTypeFacialVerification,
		AttestationTypeLivenessCheck,
		AttestationTypeDocumentVerification,
		AttestationTypeEmailVerification,
		AttestationTypeSMSVerification,
		AttestationTypeDomainVerification,
		AttestationTypeSSOVerification,
		AttestationTypeBiometricVerification,
		AttestationTypeCompositeIdentity,
	}
}

// IsValidAttestationType checks if the attestation type is valid
func IsValidAttestationType(t AttestationType) bool {
	for _, valid := range AllAttestationTypes() {
		if t == valid {
			return true
		}
	}
	return false
}

// ScopeTypeFromAttestationType maps attestation type to scope type for on-chain linkage
func ScopeTypeFromAttestationType(t AttestationType) ScopeType {
	switch t {
	case AttestationTypeFacialVerification:
		return ScopeTypeSelfie
	case AttestationTypeLivenessCheck:
		return ScopeTypeFaceVideo
	case AttestationTypeDocumentVerification:
		return ScopeTypeIDDocument
	case AttestationTypeEmailVerification:
		return ScopeTypeEmailProof
	case AttestationTypeSMSVerification:
		return ScopeTypeSMSProof
	case AttestationTypeDomainVerification:
		return ScopeTypeDomainVerify
	case AttestationTypeSSOVerification:
		return ScopeTypeSSOMetadata
	case AttestationTypeBiometricVerification:
		return ScopeTypeBiometric
	case AttestationTypeCompositeIdentity:
		return ScopeTypeIDDocument // Default to ID document for composite
	default:
		return ScopeTypeIDDocument
	}
}

// ============================================================================
// Attestation Proof Types
// ============================================================================

// AttestationProofType identifies the cryptographic signature algorithm used
type AttestationProofType string

const (
	// ProofTypeEd25519 uses Ed25519 signatures (recommended)
	ProofTypeEd25519 AttestationProofType = "Ed25519Signature2020"

	// ProofTypeSecp256k1 uses secp256k1 ECDSA signatures (Cosmos compatible)
	ProofTypeSecp256k1 AttestationProofType = "EcdsaSecp256k1Signature2019"

	// ProofTypeSr25519 uses sr25519 signatures (Substrate compatible)
	ProofTypeSr25519 AttestationProofType = "Sr25519Signature2020"
)

// AllProofTypes returns all valid proof types
func AllProofTypes() []AttestationProofType {
	return []AttestationProofType{
		ProofTypeEd25519,
		ProofTypeSecp256k1,
		ProofTypeSr25519,
	}
}

// IsValidProofType checks if the proof type is valid
func IsValidProofType(t AttestationProofType) bool {
	for _, valid := range AllProofTypes() {
		if t == valid {
			return true
		}
	}
	return false
}

// ============================================================================
// Verification Attestation
// ============================================================================

// VerificationAttestation represents a cryptographically signed attestation
// of a verification result from a VEID verification service.
//
// The attestation follows a simplified W3C Verifiable Credential structure
// optimized for on-chain storage and verification.
type VerificationAttestation struct {
	// ID is the unique identifier for this attestation
	// Format: "veid:attestation:<issuer_fingerprint>:<nonce_hex>"
	ID string `json:"id"`

	// SchemaVersion identifies the attestation schema version
	SchemaVersion string `json:"schema_version"`

	// Type identifies what kind of verification this attestation represents
	Type AttestationType `json:"type"`

	// Issuer contains information about the signer that created this attestation
	Issuer AttestationIssuer `json:"issuer"`

	// Subject identifies the identity being attested
	Subject AttestationSubject `json:"subject"`

	// Nonce is a unique value for replay protection (32 bytes, hex-encoded)
	Nonce string `json:"nonce"`

	// IssuedAt is when the attestation was created (RFC3339 timestamp)
	IssuedAt time.Time `json:"issued_at"`

	// ExpiresAt is when the attestation expires (RFC3339 timestamp)
	ExpiresAt time.Time `json:"expires_at"`

	// VerificationProofs contains evidence from the verification process
	VerificationProofs []VerificationProofDetail `json:"verification_proofs"`

	// Score is the verification score (0-100)
	Score uint32 `json:"score"`

	// Confidence is the confidence level of the verification (0-100)
	Confidence uint32 `json:"confidence"`

	// ModelVersion identifies the ML model version used for verification
	ModelVersion string `json:"model_version,omitempty"`

	// Metadata contains additional attestation-specific data
	Metadata map[string]string `json:"metadata,omitempty"`

	// Proof contains the cryptographic signature
	Proof AttestationProof `json:"proof"`
}

// AttestationIssuer identifies the signer that created the attestation
type AttestationIssuer struct {
	// ID is the DID or identifier of the issuer
	// Format: "did:virtengine:validator:<address>" or "did:virtengine:signer:<fingerprint>"
	ID string `json:"id"`

	// KeyFingerprint is the SHA256 fingerprint of the signing key (hex-encoded)
	KeyFingerprint string `json:"key_fingerprint"`

	// KeyID is an optional key identifier for key rotation support
	KeyID string `json:"key_id,omitempty"`

	// ValidatorAddress is the on-chain validator address (if applicable)
	ValidatorAddress string `json:"validator_address,omitempty"`

	// ServiceEndpoint is an optional endpoint for the signer service
	ServiceEndpoint string `json:"service_endpoint,omitempty"`
}

// AttestationSubject identifies the identity being attested
type AttestationSubject struct {
	// ID is the DID of the subject
	// Format: "did:virtengine:<account_address>"
	ID string `json:"id"`

	// AccountAddress is the blockchain address of the subject
	AccountAddress string `json:"account_address"`

	// ScopeID is the identity scope this attestation relates to (if applicable)
	ScopeID string `json:"scope_id,omitempty"`

	// RequestID links to the original verification request
	RequestID string `json:"request_id,omitempty"`
}

// VerificationProofDetail contains evidence from the verification process
type VerificationProofDetail struct {
	// ProofType identifies the type of proof (e.g., "facial_match", "liveness_score")
	ProofType string `json:"proof_type"`

	// ContentHash is the hash of the verified content (without revealing content)
	ContentHash string `json:"content_hash"`

	// Score is the individual proof score (0-100)
	Score uint32 `json:"score"`

	// Passed indicates if this proof passed the threshold
	Passed bool `json:"passed"`

	// Threshold is the required threshold for this proof
	Threshold uint32 `json:"threshold,omitempty"`

	// Timestamp is when this proof was computed
	Timestamp time.Time `json:"timestamp"`
}

// AttestationProof contains the cryptographic signature for the attestation
type AttestationProof struct {
	// Type identifies the signature algorithm
	Type AttestationProofType `json:"type"`

	// Created is when the signature was created
	Created time.Time `json:"created"`

	// VerificationMethod identifies the key used for signing
	// Format: "<issuer_id>#<key_id>" or "<issuer_id>#keys-1"
	VerificationMethod string `json:"verification_method"`

	// ProofPurpose is always "assertionMethod" for attestations
	ProofPurpose string `json:"proof_purpose"`

	// ProofValue is the base64-encoded signature
	ProofValue string `json:"proof_value"`

	// Nonce binds the proof to the attestation nonce (prevents proof reuse)
	Nonce string `json:"nonce"`

	// Domain is an optional domain binding for the signature
	Domain string `json:"domain,omitempty"`

	// Challenge is an optional challenge that was signed
	Challenge string `json:"challenge,omitempty"`
}

// ============================================================================
// Constructor Functions
// ============================================================================

// NewVerificationAttestation creates a new verification attestation
func NewVerificationAttestation(
	issuer AttestationIssuer,
	subject AttestationSubject,
	attestationType AttestationType,
	nonce []byte,
	issuedAt time.Time,
	validityDuration time.Duration,
	score uint32,
	confidence uint32,
) *VerificationAttestation {
	nonceHex := hex.EncodeToString(nonce)
	attestationID := fmt.Sprintf("veid:attestation:%s:%s", issuer.KeyFingerprint[:16], nonceHex[:16])

	return &VerificationAttestation{
		ID:                 attestationID,
		SchemaVersion:      AttestationSchemaVersion,
		Type:               attestationType,
		Issuer:             issuer,
		Subject:            subject,
		Nonce:              nonceHex,
		IssuedAt:           issuedAt,
		ExpiresAt:          issuedAt.Add(validityDuration),
		VerificationProofs: make([]VerificationProofDetail, 0),
		Score:              score,
		Confidence:         confidence,
		Metadata:           make(map[string]string),
	}
}

// NewAttestationIssuer creates a new attestation issuer
func NewAttestationIssuer(keyFingerprint string, validatorAddress string) AttestationIssuer {
	var issuerID string
	if validatorAddress != "" {
		issuerID = fmt.Sprintf("did:virtengine:validator:%s", validatorAddress)
	} else {
		issuerID = fmt.Sprintf("did:virtengine:signer:%s", keyFingerprint[:16])
	}

	return AttestationIssuer{
		ID:               issuerID,
		KeyFingerprint:   keyFingerprint,
		ValidatorAddress: validatorAddress,
	}
}

// NewAttestationSubject creates a new attestation subject
func NewAttestationSubject(accountAddress string) AttestationSubject {
	return AttestationSubject{
		ID:             fmt.Sprintf("did:virtengine:%s", accountAddress),
		AccountAddress: accountAddress,
	}
}

// NewVerificationProofDetail creates a new verification proof detail
func NewVerificationProofDetail(
	proofType string,
	contentHash string,
	score uint32,
	threshold uint32,
	timestamp time.Time,
) VerificationProofDetail {
	return VerificationProofDetail{
		ProofType:   proofType,
		ContentHash: contentHash,
		Score:       score,
		Passed:      score >= threshold,
		Threshold:   threshold,
		Timestamp:   timestamp,
	}
}

// NewAttestationProof creates a new attestation proof
func NewAttestationProof(
	proofType AttestationProofType,
	created time.Time,
	verificationMethod string,
	signature []byte,
	nonce string,
) AttestationProof {
	return AttestationProof{
		Type:               proofType,
		Created:            created,
		VerificationMethod: verificationMethod,
		ProofPurpose:       "assertionMethod",
		ProofValue:         base64.StdEncoding.EncodeToString(signature),
		Nonce:              nonce,
	}
}

// ============================================================================
// Validation Methods
// ============================================================================

// Validate validates the verification attestation
func (a *VerificationAttestation) Validate() error {
	if a.ID == "" {
		return ErrInvalidAttestation.Wrap("attestation ID is required")
	}

	if a.SchemaVersion == "" {
		return ErrInvalidAttestation.Wrap("schema version is required")
	}

	if !IsValidAttestationType(a.Type) {
		return ErrInvalidAttestation.Wrapf("invalid attestation type: %s", a.Type)
	}

	if err := a.Issuer.Validate(); err != nil {
		return ErrInvalidAttestation.Wrapf("invalid issuer: %v", err)
	}

	if err := a.Subject.Validate(); err != nil {
		return ErrInvalidAttestation.Wrapf("invalid subject: %v", err)
	}

	if len(a.Nonce) < 32 {
		return ErrInvalidAttestation.Wrap("nonce must be at least 32 hex characters (16 bytes)")
	}

	if _, err := hex.DecodeString(a.Nonce); err != nil {
		return ErrInvalidAttestation.Wrap("nonce must be valid hex encoding")
	}

	if a.IssuedAt.IsZero() {
		return ErrInvalidAttestation.Wrap("issued_at is required")
	}

	if a.ExpiresAt.IsZero() {
		return ErrInvalidAttestation.Wrap("expires_at is required")
	}

	if !a.ExpiresAt.After(a.IssuedAt) {
		return ErrInvalidAttestation.Wrap("expires_at must be after issued_at")
	}

	if a.Score > 100 {
		return ErrInvalidAttestation.Wrap("score cannot exceed 100")
	}

	if a.Confidence > 100 {
		return ErrInvalidAttestation.Wrap("confidence cannot exceed 100")
	}

	for i, proof := range a.VerificationProofs {
		if err := proof.Validate(); err != nil {
			return ErrInvalidAttestation.Wrapf("invalid verification proof at index %d: %v", i, err)
		}
	}

	if err := a.Proof.Validate(); err != nil {
		return ErrInvalidAttestation.Wrapf("invalid proof: %v", err)
	}

	return nil
}

// Validate validates the attestation issuer
func (i AttestationIssuer) Validate() error {
	if i.ID == "" {
		return fmt.Errorf("issuer ID is required")
	}

	if i.KeyFingerprint == "" {
		return fmt.Errorf("key fingerprint is required")
	}

	if len(i.KeyFingerprint) < 32 {
		return fmt.Errorf("key fingerprint must be at least 32 hex characters")
	}

	if _, err := hex.DecodeString(i.KeyFingerprint); err != nil {
		return fmt.Errorf("key fingerprint must be valid hex encoding")
	}

	return nil
}

// Validate validates the attestation subject
func (s AttestationSubject) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("subject ID is required")
	}

	if s.AccountAddress == "" {
		return fmt.Errorf("account address is required")
	}

	return nil
}

// Validate validates the verification proof detail
func (p VerificationProofDetail) Validate() error {
	if p.ProofType == "" {
		return fmt.Errorf("proof type is required")
	}

	if p.ContentHash == "" {
		return fmt.Errorf("content hash is required")
	}

	if p.Score > 100 {
		return fmt.Errorf("score cannot exceed 100")
	}

	if p.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}

	return nil
}

// Validate validates the attestation proof
func (p AttestationProof) Validate() error {
	if !IsValidProofType(p.Type) {
		return ErrInvalidProof.Wrapf("invalid proof type: %s", p.Type)
	}

	if p.Created.IsZero() {
		return ErrInvalidProof.Wrap("proof created timestamp is required")
	}

	if p.VerificationMethod == "" {
		return ErrInvalidProof.Wrap("verification method is required")
	}

	if p.ProofPurpose == "" {
		return ErrInvalidProof.Wrap("proof purpose is required")
	}

	if p.ProofValue == "" {
		return ErrInvalidProof.Wrap("proof value is required")
	}

	// Verify base64 encoding
	if _, err := base64.StdEncoding.DecodeString(p.ProofValue); err != nil {
		return ErrInvalidProof.Wrap("proof value must be valid base64 encoding")
	}

	if p.Nonce == "" {
		return ErrInvalidProof.Wrap("proof nonce is required")
	}

	return nil
}

// ============================================================================
// Canonical Serialization
// ============================================================================

// CanonicalBytes returns the canonical byte representation for signing.
// The canonical form excludes the proof field and uses deterministic JSON encoding.
func (a *VerificationAttestation) CanonicalBytes() ([]byte, error) {
	// Create a copy without the proof for signing
	canonical := struct {
		ID                 string                    `json:"id"`
		SchemaVersion      string                    `json:"schema_version"`
		Type               AttestationType           `json:"type"`
		Issuer             AttestationIssuer         `json:"issuer"`
		Subject            AttestationSubject        `json:"subject"`
		Nonce              string                    `json:"nonce"`
		IssuedAt           string                    `json:"issued_at"`
		ExpiresAt          string                    `json:"expires_at"`
		VerificationProofs []VerificationProofDetail `json:"verification_proofs"`
		Score              uint32                    `json:"score"`
		Confidence         uint32                    `json:"confidence"`
		ModelVersion       string                    `json:"model_version,omitempty"`
		Metadata           map[string]string         `json:"metadata,omitempty"`
	}{
		ID:                 a.ID,
		SchemaVersion:      a.SchemaVersion,
		Type:               a.Type,
		Issuer:             a.Issuer,
		Subject:            a.Subject,
		Nonce:              a.Nonce,
		IssuedAt:           a.IssuedAt.UTC().Format(time.RFC3339Nano),
		ExpiresAt:          a.ExpiresAt.UTC().Format(time.RFC3339Nano),
		VerificationProofs: a.VerificationProofs,
		Score:              a.Score,
		Confidence:         a.Confidence,
		ModelVersion:       a.ModelVersion,
		Metadata:           a.sortedMetadata(),
	}

	return json.Marshal(canonical)
}

// Hash computes the SHA256 hash of the canonical representation
func (a *VerificationAttestation) Hash() ([]byte, error) {
	canonicalBytes, err := a.CanonicalBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to compute canonical bytes: %w", err)
	}

	hash := sha256.Sum256(canonicalBytes)
	return hash[:], nil
}

// HashHex returns the hex-encoded SHA256 hash
func (a *VerificationAttestation) HashHex() (string, error) {
	hash, err := a.Hash()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash), nil
}

// sortedMetadata returns metadata with keys sorted for deterministic serialization
func (a *VerificationAttestation) sortedMetadata() map[string]string {
	if len(a.Metadata) == 0 {
		return nil
	}

	keys := make([]string, 0, len(a.Metadata))
	for k := range a.Metadata {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sorted := make(map[string]string, len(a.Metadata))
	for _, k := range keys {
		sorted[k] = a.Metadata[k]
	}
	return sorted
}

// ToJSON serializes the attestation to JSON
func (a *VerificationAttestation) ToJSON() ([]byte, error) {
	return json.MarshalIndent(a, "", "  ")
}

// AttestationFromJSON deserializes an attestation from JSON
func AttestationFromJSON(data []byte) (*VerificationAttestation, error) {
	var a VerificationAttestation
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attestation: %w", err)
	}
	return &a, nil
}

// ============================================================================
// Utility Methods
// ============================================================================

// IsExpired checks if the attestation has expired
func (a *VerificationAttestation) IsExpired(now time.Time) bool {
	return now.After(a.ExpiresAt)
}

// IsValid checks if the attestation is currently valid (not expired and after issuance)
func (a *VerificationAttestation) IsValid(now time.Time) bool {
	return now.After(a.IssuedAt) && now.Before(a.ExpiresAt)
}

// AddVerificationProof adds a verification proof detail
func (a *VerificationAttestation) AddVerificationProof(proof VerificationProofDetail) {
	a.VerificationProofs = append(a.VerificationProofs, proof)
}

// SetProof sets the cryptographic proof on the attestation
func (a *VerificationAttestation) SetProof(proof AttestationProof) {
	a.Proof = proof
}

// SetMetadata sets a metadata key-value pair
func (a *VerificationAttestation) SetMetadata(key, value string) {
	if a.Metadata == nil {
		a.Metadata = make(map[string]string)
	}
	a.Metadata[key] = value
}

// GetProofBytes returns the decoded signature bytes from the proof
func (a *VerificationAttestation) GetProofBytes() ([]byte, error) {
	return base64.StdEncoding.DecodeString(a.Proof.ProofValue)
}

// ToScopeType returns the corresponding scope type for on-chain linkage
func (a *VerificationAttestation) ToScopeType() ScopeType {
	return ScopeTypeFromAttestationType(a.Type)
}

// String returns a string representation of the attestation
func (a *VerificationAttestation) String() string {
	return fmt.Sprintf("VerificationAttestation{ID: %s, Type: %s, Subject: %s, Score: %d}",
		a.ID, a.Type, a.Subject.AccountAddress, a.Score)
}
