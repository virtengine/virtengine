// Package types provides VEID module types.
//
// This file implements W3C Verifiable Credentials Data Model v1.1 compatible
// credential types for identity verification attestations.
//
// Task Reference: VE-3025 - Verifiable Credential Issuance
package types

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// ============================================================================
// W3C VC Contexts and Types
// ============================================================================

const (
	// W3C standard contexts
	ContextW3CCredentials = "https://www.w3.org/2018/credentials/v1"
	ContextVirtEngine     = "https://virtengine.io/credentials/v1"

	// W3C standard types
	TypeVerifiableCredential = "VerifiableCredential"

	// VirtEngine credential types
	TypeVEIDCredential        = "VEIDCredential"
	TypeIdentityVerification  = "IdentityVerificationCredential"
	TypeFacialVerification    = "FacialVerificationCredential"
	TypeDocumentVerification  = "DocumentVerificationCredential"
	TypeEmailVerification     = "EmailVerificationCredential"
	TypeSMSVerification       = "SMSVerificationCredential"
	TypeDomainVerification    = "DomainVerificationCredential"
	TypeSSOVerification       = "SSOVerificationCredential"
	TypeLivenessVerification  = "LivenessVerificationCredential"
	TypeBiometricVerification = "BiometricVerificationCredential"

	// Proof types
	ProofTypeEd25519Signature2020   = "Ed25519Signature2020"
	ProofTypeSecp256k1Signature2019 = "EcdsaSecp256k1Signature2019"

	// Proof purposes
	ProofPurposeAssertion = "assertionMethod"

	// Credential status types
	CredentialStatusActive  = "active"
	CredentialStatusRevoked = "revoked"
	CredentialStatusExpired = "expired"

	// DID method prefix for VirtEngine
	DIDMethodVirtEngine = "did:virtengine:"
)

// ============================================================================
// Verifiable Credential Types
// ============================================================================

// VerifiableCredential represents a W3C Verifiable Credential.
// This implements the W3C Verifiable Credentials Data Model v1.1 specification.
type VerifiableCredential struct {
	// Context is the JSON-LD context(s) for the credential
	Context []string `json:"@context"`

	// ID is the unique identifier for this credential
	ID string `json:"id"`

	// Type specifies the type(s) of the credential
	Type []string `json:"type"`

	// Issuer identifies who issued the credential
	Issuer CredentialIssuer `json:"issuer"`

	// IssuanceDate is when the credential was issued
	IssuanceDate time.Time `json:"issuanceDate"`

	// ExpirationDate is when the credential expires (optional)
	ExpirationDate *time.Time `json:"expirationDate,omitempty"`

	// CredentialSubject contains claims about the subject
	CredentialSubject CredentialSubject `json:"credentialSubject"`

	// CredentialStatus indicates the revocation status
	CredentialStatus *CredentialStatus `json:"credentialStatus,omitempty"`

	// Proof contains the cryptographic proof
	Proof CredentialProof `json:"proof"`
}

// CredentialIssuer identifies the issuer of a credential.
type CredentialIssuer struct {
	// ID is the DID of the issuer (validator)
	ID string `json:"id"`

	// Name is the human-readable name of the issuer
	Name string `json:"name,omitempty"`

	// ValidatorAddress is the on-chain validator address
	ValidatorAddress string `json:"validatorAddress,omitempty"`
}

// CredentialSubject contains claims about the credential subject.
type CredentialSubject struct {
	// ID is the DID of the subject (user)
	ID string `json:"id"`

	// VerificationType indicates what type of verification was performed
	VerificationType string `json:"verificationType"`

	// VerificationLevel indicates the assurance level (1-4)
	VerificationLevel int `json:"verificationLevel"`

	// TrustScore is the computed trust score (0.0-1.0)
	TrustScore float64 `json:"trustScore"`

	// Claims contains additional verification claims
	Claims map[string]interface{} `json:"claims,omitempty"`
}

// CredentialStatus represents the revocation status of a credential.
type CredentialStatus struct {
	// ID is the identifier for the status entry
	ID string `json:"id"`

	// Type is the status mechanism type
	Type string `json:"type"`

	// Status is the current status (active, revoked, expired)
	Status string `json:"status"`

	// StatusListIndex is the index in the status list
	StatusListIndex uint64 `json:"statusListIndex,omitempty"`

	// StatusListCredential is the URL of the status list credential
	StatusListCredential string `json:"statusListCredential,omitempty"`
}

// CredentialProof contains the cryptographic proof for a credential.
type CredentialProof struct {
	// Type is the signature algorithm type
	Type string `json:"type"`

	// Created is when the proof was created
	Created time.Time `json:"created"`

	// VerificationMethod is the key used to create the proof
	VerificationMethod string `json:"verificationMethod"`

	// ProofPurpose is the purpose of the proof
	ProofPurpose string `json:"proofPurpose"`

	// ProofValue is the Base64-encoded signature
	ProofValue string `json:"proofValue"`

	// Challenge is the optional challenge that was signed
	Challenge string `json:"challenge,omitempty"`

	// Domain is the optional domain binding
	Domain string `json:"domain,omitempty"`
}

// ============================================================================
// Credential Storage Types
// ============================================================================

// CredentialRecord is the on-chain storage record for a credential.
type CredentialRecord struct {
	// CredentialID is the unique identifier
	CredentialID string `json:"credential_id"`

	// SubjectAddress is the blockchain address of the subject
	SubjectAddress string `json:"subject_address"`

	// IssuerAddress is the blockchain address of the issuer (validator)
	IssuerAddress string `json:"issuer_address"`

	// CredentialHash is the SHA256 hash of the full credential JSON
	CredentialHash []byte `json:"credential_hash"`

	// CredentialTypes are the types of this credential
	CredentialTypes []string `json:"credential_types"`

	// IssuedAt is when the credential was issued
	IssuedAt time.Time `json:"issued_at"`

	// ExpiresAt is when the credential expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// RevokedAt is when the credential was revoked (if revoked)
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// RevocationReason is the reason for revocation
	RevocationReason string `json:"revocation_reason,omitempty"`

	// Status is the current credential status
	Status string `json:"status"`

	// VerificationRequestID links to the original verification request
	VerificationRequestID string `json:"verification_request_id,omitempty"`

	// BlockHeight is the block height when issued
	BlockHeight int64 `json:"block_height"`
}

// CredentialIssuanceRequest contains parameters for issuing a credential.
type CredentialIssuanceRequest struct {
	// SubjectAddress is the blockchain address of the subject
	SubjectAddress string `json:"subject_address"`

	// VerificationType is the type of verification performed
	VerificationType string `json:"verification_type"`

	// VerificationLevel is the assurance level achieved
	VerificationLevel int `json:"verification_level"`

	// TrustScore is the computed trust score
	TrustScore float64 `json:"trust_score"`

	// Claims are additional claims to include
	Claims map[string]interface{} `json:"claims,omitempty"`

	// ValidityDuration is how long the credential is valid
	ValidityDuration time.Duration `json:"validity_duration"`

	// VerificationRequestID links to the verification request
	VerificationRequestID string `json:"verification_request_id,omitempty"`
}

// ============================================================================
// Constructor Functions
// ============================================================================

// NewVerifiableCredential creates a new verifiable credential.
func NewVerifiableCredential(
	credentialID string,
	issuer CredentialIssuer,
	subject CredentialSubject,
	issuanceDate time.Time,
	expirationDate *time.Time,
	credentialTypes []string,
) *VerifiableCredential {
	// Build context
	context := []string{ContextW3CCredentials, ContextVirtEngine}

	// Build types - always include base type
	types := []string{TypeVerifiableCredential}
	types = append(types, credentialTypes...)

	return &VerifiableCredential{
		Context:           context,
		ID:                credentialID,
		Type:              types,
		Issuer:            issuer,
		IssuanceDate:      issuanceDate,
		ExpirationDate:    expirationDate,
		CredentialSubject: subject,
	}
}

// NewCredentialIssuer creates a new credential issuer.
func NewCredentialIssuer(validatorAddress string, name string) CredentialIssuer {
	return CredentialIssuer{
		ID:               DIDMethodVirtEngine + validatorAddress,
		Name:             name,
		ValidatorAddress: validatorAddress,
	}
}

// NewCredentialSubject creates a new credential subject.
func NewCredentialSubject(
	accountAddress string,
	verificationType string,
	verificationLevel int,
	trustScore float64,
	claims map[string]interface{},
) CredentialSubject {
	return CredentialSubject{
		ID:                DIDMethodVirtEngine + accountAddress,
		VerificationType:  verificationType,
		VerificationLevel: verificationLevel,
		TrustScore:        trustScore,
		Claims:            claims,
	}
}

// NewCredentialRecord creates a new credential storage record.
func NewCredentialRecord(
	credentialID string,
	subjectAddress string,
	issuerAddress string,
	credentialHash []byte,
	credentialTypes []string,
	issuedAt time.Time,
	expiresAt *time.Time,
	verificationRequestID string,
	blockHeight int64,
) *CredentialRecord {
	return &CredentialRecord{
		CredentialID:          credentialID,
		SubjectAddress:        subjectAddress,
		IssuerAddress:         issuerAddress,
		CredentialHash:        credentialHash,
		CredentialTypes:       credentialTypes,
		IssuedAt:              issuedAt,
		ExpiresAt:             expiresAt,
		Status:                CredentialStatusActive,
		VerificationRequestID: verificationRequestID,
		BlockHeight:           blockHeight,
	}
}

// ============================================================================
// Validation Methods
// ============================================================================

// Validate validates the verifiable credential.
func (vc *VerifiableCredential) Validate() error {
	if len(vc.Context) == 0 {
		return ErrInvalidCredential.Wrap("context is required")
	}

	// Must include W3C context
	hasW3CContext := false
	for _, ctx := range vc.Context {
		if ctx == ContextW3CCredentials {
			hasW3CContext = true
			break
		}
	}
	if !hasW3CContext {
		return ErrInvalidCredential.Wrap("W3C credentials context is required")
	}

	if vc.ID == "" {
		return ErrInvalidCredential.Wrap("credential ID is required")
	}

	if len(vc.Type) == 0 {
		return ErrInvalidCredential.Wrap("type is required")
	}

	// Must include VerifiableCredential type
	hasVCType := false
	for _, t := range vc.Type {
		if t == TypeVerifiableCredential {
			hasVCType = true
			break
		}
	}
	if !hasVCType {
		return ErrInvalidCredential.Wrap("VerifiableCredential type is required")
	}

	if err := vc.Issuer.Validate(); err != nil {
		return ErrInvalidCredential.Wrapf("invalid issuer: %v", err)
	}

	if vc.IssuanceDate.IsZero() {
		return ErrInvalidCredential.Wrap("issuance date is required")
	}

	if vc.ExpirationDate != nil && vc.ExpirationDate.Before(vc.IssuanceDate) {
		return ErrInvalidCredential.Wrap("expiration date must be after issuance date")
	}

	if err := vc.CredentialSubject.Validate(); err != nil {
		return ErrInvalidCredential.Wrapf("invalid credential subject: %v", err)
	}

	return nil
}

// Validate validates the credential issuer.
func (ci CredentialIssuer) Validate() error {
	if ci.ID == "" {
		return fmt.Errorf("issuer ID is required")
	}
	return nil
}

// Validate validates the credential subject.
func (cs CredentialSubject) Validate() error {
	if cs.ID == "" {
		return fmt.Errorf("subject ID is required")
	}

	if cs.VerificationType == "" {
		return fmt.Errorf("verification type is required")
	}

	if cs.VerificationLevel < 0 || cs.VerificationLevel > 4 {
		return fmt.Errorf("verification level must be 0-4")
	}

	if cs.TrustScore < 0 || cs.TrustScore > 1 {
		return fmt.Errorf("trust score must be 0.0-1.0")
	}

	return nil
}

// Validate validates the credential proof.
func (cp CredentialProof) Validate() error {
	if cp.Type == "" {
		return ErrInvalidProof.Wrap("proof type is required")
	}

	if cp.Created.IsZero() {
		return ErrInvalidProof.Wrap("proof created timestamp is required")
	}

	if cp.VerificationMethod == "" {
		return ErrInvalidProof.Wrap("verification method is required")
	}

	if cp.ProofPurpose == "" {
		return ErrInvalidProof.Wrap("proof purpose is required")
	}

	if cp.ProofValue == "" {
		return ErrInvalidProof.Wrap("proof value is required")
	}

	return nil
}

// Validate validates the credential record.
func (cr *CredentialRecord) Validate() error {
	if cr.CredentialID == "" {
		return ErrInvalidCredential.Wrap("credential ID is required")
	}

	if cr.SubjectAddress == "" {
		return ErrInvalidCredential.Wrap("subject address is required")
	}

	if cr.IssuerAddress == "" {
		return ErrInvalidCredential.Wrap("issuer address is required")
	}

	if len(cr.CredentialHash) != sha256.Size {
		return ErrInvalidCredential.Wrapf("credential hash must be %d bytes", sha256.Size)
	}

	if cr.IssuedAt.IsZero() {
		return ErrInvalidCredential.Wrap("issued at is required")
	}

	if cr.Status == "" {
		return ErrInvalidCredential.Wrap("status is required")
	}

	return nil
}

// ============================================================================
// Utility Methods
// ============================================================================

// Hash computes the SHA256 hash of the credential (excluding proof).
func (vc *VerifiableCredential) Hash() ([]byte, error) {
	// Create a copy without the proof for hashing
	hashable := struct {
		Context           []string          `json:"@context"`
		ID                string            `json:"id"`
		Type              []string          `json:"type"`
		Issuer            CredentialIssuer  `json:"issuer"`
		IssuanceDate      time.Time         `json:"issuanceDate"`
		ExpirationDate    *time.Time        `json:"expirationDate,omitempty"`
		CredentialSubject CredentialSubject `json:"credentialSubject"`
		CredentialStatus  *CredentialStatus `json:"credentialStatus,omitempty"`
	}{
		Context:           vc.Context,
		ID:                vc.ID,
		Type:              vc.Type,
		Issuer:            vc.Issuer,
		IssuanceDate:      vc.IssuanceDate,
		ExpirationDate:    vc.ExpirationDate,
		CredentialSubject: vc.CredentialSubject,
		CredentialStatus:  vc.CredentialStatus,
	}

	jsonBytes, err := json.Marshal(hashable)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal credential for hashing: %w", err)
	}

	hash := sha256.Sum256(jsonBytes)
	return hash[:], nil
}

// ToJSON serializes the credential to JSON.
func (vc *VerifiableCredential) ToJSON() ([]byte, error) {
	return json.MarshalIndent(vc, "", "  ")
}

// FromJSON deserializes a credential from JSON.
func FromJSON(data []byte) (*VerifiableCredential, error) {
	var vc VerifiableCredential
	if err := json.Unmarshal(data, &vc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credential: %w", err)
	}
	return &vc, nil
}

// SetProof sets the proof on the credential.
func (vc *VerifiableCredential) SetProof(proof CredentialProof) {
	vc.Proof = proof
}

// IsExpired checks if the credential has expired.
func (vc *VerifiableCredential) IsExpired(now time.Time) bool {
	if vc.ExpirationDate == nil {
		return false
	}
	return now.After(*vc.ExpirationDate)
}

// IsExpired checks if the credential record has expired.
func (cr *CredentialRecord) IsExpired(now time.Time) bool {
	if cr.ExpiresAt == nil {
		return false
	}
	return now.After(*cr.ExpiresAt)
}

// IsRevoked checks if the credential record is revoked.
func (cr *CredentialRecord) IsRevoked() bool {
	return cr.RevokedAt != nil || cr.Status == CredentialStatusRevoked
}

// IsActive checks if the credential record is active.
func (cr *CredentialRecord) IsActive(now time.Time) bool {
	return !cr.IsRevoked() && !cr.IsExpired(now)
}

// Revoke marks the credential record as revoked.
func (cr *CredentialRecord) Revoke(revokedAt time.Time, reason string) {
	cr.RevokedAt = &revokedAt
	cr.RevocationReason = reason
	cr.Status = CredentialStatusRevoked
}

// ============================================================================
// Proof Helper Functions
// ============================================================================

// NewCredentialProof creates a new credential proof.
func NewCredentialProof(
	proofType string,
	created time.Time,
	verificationMethod string,
	proofPurpose string,
	signature []byte,
) CredentialProof {
	return CredentialProof{
		Type:               proofType,
		Created:            created,
		VerificationMethod: verificationMethod,
		ProofPurpose:       proofPurpose,
		ProofValue:         base64.StdEncoding.EncodeToString(signature),
	}
}

// GetProofBytes returns the decoded proof bytes.
func (cp CredentialProof) GetProofBytes() ([]byte, error) {
	return base64.StdEncoding.DecodeString(cp.ProofValue)
}

// ============================================================================
// Credential Type Helpers
// ============================================================================

// CredentialTypeFromScopeType maps a scope type to a credential type.
func CredentialTypeFromScopeType(scopeType ScopeType) string {
	switch scopeType {
	case ScopeTypeIDDocument:
		return TypeDocumentVerification
	case ScopeTypeSelfie:
		return TypeFacialVerification
	case ScopeTypeFaceVideo:
		return TypeLivenessVerification
	case ScopeTypeBiometric:
		return TypeBiometricVerification
	case ScopeTypeEmailProof:
		return TypeEmailVerification
	case ScopeTypeSMSProof:
		return TypeSMSVerification
	case ScopeTypeDomainVerify:
		return TypeDomainVerification
	case ScopeTypeSSOMetadata, ScopeTypeADSSO:
		return TypeSSOVerification
	default:
		return TypeVEIDCredential
	}
}

// VerificationLevelFromScore maps a score to a verification level.
func VerificationLevelFromScore(score uint32) int {
	switch {
	case score >= 90:
		return 4 // Highest assurance
	case score >= 70:
		return 3 // Substantial assurance
	case score >= 50:
		return 2 // Basic assurance
	case score >= 30:
		return 1 // Low assurance
	default:
		return 0 // Unverified
	}
}

// TrustScoreFromScore converts a uint32 score (0-100) to float64 (0.0-1.0).
func TrustScoreFromScore(score uint32) float64 {
	if score > 100 {
		score = 100
	}
	return float64(score) / 100.0
}
