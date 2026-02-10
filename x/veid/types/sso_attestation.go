// Package types provides VEID module types.
//
// VE-4B: SSO/OIDC Attestation Schema
// This file defines the attestation payload schema for SSO/OIDC verification.
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// ============================================================================
// SSO Attestation Schema Version
// ============================================================================

const (
	// SSOAttestationSchemaVersion is the current schema version
	SSOAttestationSchemaVersion = "1.0.0"

	// DefaultSSOAttestationValidityDuration is the default validity for SSO attestations
	DefaultSSOAttestationValidityDuration = 365 * 24 * time.Hour // 1 year
)

// ============================================================================
// SSO Attestation Types
// ============================================================================

// SSOAttestation represents a signed attestation from SSO/OIDC verification.
// It extends the base VerificationAttestation with SSO-specific fields.
type SSOAttestation struct {
	// Embed base attestation fields
	VerificationAttestation

	// OIDCIssuer is the OIDC issuer URL (e.g., "https://accounts.google.com")
	OIDCIssuer string `json:"oidc_issuer"`

	// SubjectHash is the SHA256 hash of the OIDC subject identifier
	SubjectHash string `json:"subject_hash"`

	// EmailHash is the SHA256 hash of the verified email (optional)
	EmailHash string `json:"email_hash,omitempty"`

	// EmailDomainHash is the SHA256 hash of the email domain (for org matching)
	EmailDomainHash string `json:"email_domain_hash,omitempty"`

	// TenantIDHash is the hash of the organization tenant ID (if applicable)
	TenantIDHash string `json:"tenant_id_hash,omitempty"`

	// ProviderType identifies the SSO provider category
	ProviderType SSOProviderType `json:"provider_type"`

	// OIDCNonce is the nonce used in the OIDC flow (prevents replay)
	OIDCNonce string `json:"oidc_nonce"`

	// OIDCAuthTime is when the user authenticated (if available)
	OIDCAuthTime *time.Time `json:"oidc_auth_time,omitempty"`

	// OIDCACRValues are the authentication context class reference values
	OIDCACRValues []string `json:"oidc_acr_values,omitempty"`

	// OIDCAMRValues are the authentication methods references
	OIDCAMRValues []string `json:"oidc_amr_values,omitempty"`

	// EmailVerified indicates if the email was verified by the IdP
	EmailVerified bool `json:"email_verified"`

	// LinkedAccountAddress is the blockchain account this SSO is linked to
	LinkedAccountAddress string `json:"linked_account_address"`

	// LinkageSignature is the user's signature authorizing the linkage
	LinkageSignature []byte `json:"linkage_signature"`
}

// NewSSOAttestation creates a new SSO attestation.
func NewSSOAttestation(
	issuer AttestationIssuer,
	subject AttestationSubject,
	oidcIssuer string,
	subjectID string,
	providerType SSOProviderType,
	oidcNonce string,
	nonce []byte,
	issuedAt time.Time,
	validityDuration time.Duration,
) *SSOAttestation {
	base := NewVerificationAttestation(
		issuer,
		subject,
		AttestationTypeSSOVerification,
		nonce,
		issuedAt,
		validityDuration,
		100, // SSO verification always has 100% score when valid
		100, // 100% confidence
	)

	return &SSOAttestation{
		VerificationAttestation: *base,
		OIDCIssuer:              oidcIssuer,
		SubjectHash:             hashIdentifier(subjectID),
		ProviderType:            providerType,
		OIDCNonce:               oidcNonce,
		LinkedAccountAddress:    subject.AccountAddress,
	}
}

// hashIdentifier creates a SHA256 hash of an identifier.
func hashIdentifier(id string) string {
	if id == "" {
		return ""
	}
	hash := sha256.Sum256([]byte(id))
	return hex.EncodeToString(hash[:])
}

// SetEmail sets the email information on the attestation.
func (a *SSOAttestation) SetEmail(email string, domain string, verified bool) {
	a.EmailHash = hashIdentifier(email)
	a.EmailDomainHash = hashIdentifier(domain)
	a.EmailVerified = verified
}

// SetTenantID sets the tenant/organization ID.
func (a *SSOAttestation) SetTenantID(tenantID string) {
	a.TenantIDHash = hashIdentifier(tenantID)
}

// SetAuthContext sets authentication context information.
func (a *SSOAttestation) SetAuthContext(authTime *time.Time, acrValues, amrValues []string) {
	a.OIDCAuthTime = authTime
	a.OIDCACRValues = acrValues
	a.OIDCAMRValues = amrValues
}

// SetLinkageSignature sets the user's linkage authorization signature.
func (a *SSOAttestation) SetLinkageSignature(signature []byte) {
	a.LinkageSignature = make([]byte, len(signature))
	copy(a.LinkageSignature, signature)
}

// Validate validates the SSO attestation.
func (a *SSOAttestation) Validate() error {
	// Validate base attestation
	if err := a.VerificationAttestation.Validate(); err != nil {
		return err
	}

	if a.OIDCIssuer == "" {
		return ErrInvalidSSO.Wrap("oidc_issuer is required")
	}

	if a.SubjectHash == "" {
		return ErrInvalidSSO.Wrap("subject_hash is required")
	}

	if len(a.SubjectHash) != 64 {
		return ErrInvalidSSO.Wrap("subject_hash must be a valid SHA256 hex string")
	}

	if !IsValidSSOProviderType(a.ProviderType) {
		return ErrInvalidSSO.Wrapf("invalid provider_type: %s", a.ProviderType)
	}

	if a.OIDCNonce == "" {
		return ErrInvalidSSO.Wrap("oidc_nonce is required")
	}

	if a.LinkedAccountAddress == "" {
		return ErrInvalidAddress.Wrap("linked_account_address is required")
	}

	return nil
}

// CanonicalBytes returns the canonical byte representation for signing.
func (a *SSOAttestation) CanonicalBytes() ([]byte, error) {
	canonical := struct {
		// Base attestation fields
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
		// SSO-specific fields
		OIDCIssuer           string          `json:"oidc_issuer"`
		SubjectHash          string          `json:"subject_hash"`
		EmailHash            string          `json:"email_hash,omitempty"`
		EmailDomainHash      string          `json:"email_domain_hash,omitempty"`
		TenantIDHash         string          `json:"tenant_id_hash,omitempty"`
		ProviderType         SSOProviderType `json:"provider_type"`
		OIDCNonce            string          `json:"oidc_nonce"`
		OIDCAuthTime         *time.Time      `json:"oidc_auth_time,omitempty"`
		OIDCACRValues        []string        `json:"oidc_acr_values,omitempty"`
		OIDCAMRValues        []string        `json:"oidc_amr_values,omitempty"`
		EmailVerified        bool            `json:"email_verified"`
		LinkedAccountAddress string          `json:"linked_account_address"`
	}{
		ID:                   a.ID,
		SchemaVersion:        a.SchemaVersion,
		Type:                 a.Type,
		Issuer:               a.Issuer,
		Subject:              a.Subject,
		Nonce:                a.Nonce,
		IssuedAt:             a.IssuedAt.UTC().Format(time.RFC3339Nano),
		ExpiresAt:            a.ExpiresAt.UTC().Format(time.RFC3339Nano),
		VerificationProofs:   a.VerificationProofs,
		Score:                a.Score,
		Confidence:           a.Confidence,
		OIDCIssuer:           a.OIDCIssuer,
		SubjectHash:          a.SubjectHash,
		EmailHash:            a.EmailHash,
		EmailDomainHash:      a.EmailDomainHash,
		TenantIDHash:         a.TenantIDHash,
		ProviderType:         a.ProviderType,
		OIDCNonce:            a.OIDCNonce,
		OIDCAuthTime:         a.OIDCAuthTime,
		OIDCACRValues:        a.OIDCACRValues,
		OIDCAMRValues:        a.OIDCAMRValues,
		EmailVerified:        a.EmailVerified,
		LinkedAccountAddress: a.LinkedAccountAddress,
	}

	return json.Marshal(canonical)
}

// ToLinkageMetadata converts the attestation to on-chain linkage metadata.
func (a *SSOAttestation) ToLinkageMetadata(linkageID string) *SSOLinkageMetadata {
	meta := &SSOLinkageMetadata{
		Version:     SSOVerificationVersion,
		LinkageID:   linkageID,
		Provider:    a.ProviderType,
		Issuer:      a.OIDCIssuer,
		SubjectHash: a.SubjectHash,
		Nonce:       a.OIDCNonce,
		VerifiedAt:  a.IssuedAt,
		Status:      SSOStatusVerified,
	}

	if !a.ExpiresAt.IsZero() {
		meta.ExpiresAt = &a.ExpiresAt
	}

	meta.EmailDomainHash = a.EmailDomainHash
	meta.OrgIDHash = a.TenantIDHash
	meta.AccountSignature = a.LinkageSignature

	return meta
}

// String returns a string representation.
func (a *SSOAttestation) String() string {
	return fmt.Sprintf("SSOAttestation{ID: %s, Provider: %s, Issuer: %s, Account: %s}",
		a.ID, a.ProviderType, a.OIDCIssuer, a.LinkedAccountAddress)
}

// ============================================================================
// SSO Attestation Request
// ============================================================================

// SSOAttestationRequest contains parameters for requesting an SSO attestation.
type SSOAttestationRequest struct {
	// AccountAddress is the blockchain account to link
	AccountAddress string `json:"account_address"`

	// ProviderType is the SSO provider to use
	ProviderType SSOProviderType `json:"provider_type"`

	// OIDCIssuer is the specific OIDC issuer (for generic OIDC)
	OIDCIssuer string `json:"oidc_issuer,omitempty"`

	// RedirectURI is where to redirect after OIDC flow
	RedirectURI string `json:"redirect_uri"`

	// State is the OAuth2 state parameter (for CSRF protection)
	State string `json:"state"`

	// Nonce is the OIDC nonce for replay protection
	Nonce string `json:"nonce"`

	// RequiredScopes specifies additional OAuth scopes to request
	RequiredScopes []string `json:"required_scopes,omitempty"`

	// LinkageSignatureMessage is the message to sign for account linkage
	LinkageSignatureMessage string `json:"linkage_signature_message,omitempty"`

	// CreatedAt is when the request was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the request expires
	ExpiresAt time.Time `json:"expires_at"`
}

// NewSSOAttestationRequest creates a new SSO attestation request.
func NewSSOAttestationRequest(
	accountAddress string,
	providerType SSOProviderType,
	redirectURI string,
	state string,
	nonce string,
	ttlSeconds int64,
) *SSOAttestationRequest {
	now := time.Now()
	return &SSOAttestationRequest{
		AccountAddress: accountAddress,
		ProviderType:   providerType,
		RedirectURI:    redirectURI,
		State:          state,
		Nonce:          nonce,
		CreatedAt:      now,
		ExpiresAt:      now.Add(time.Duration(ttlSeconds) * time.Second),
	}
}

// Validate validates the SSO attestation request.
func (r *SSOAttestationRequest) Validate() error {
	if r.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account_address is required")
	}
	if !IsValidSSOProviderType(r.ProviderType) {
		return ErrInvalidSSO.Wrapf("invalid provider_type: %s", r.ProviderType)
	}
	if r.RedirectURI == "" {
		return ErrInvalidSSO.Wrap("redirect_uri is required")
	}
	if r.State == "" {
		return ErrInvalidSSO.Wrap("state is required")
	}
	if r.Nonce == "" {
		return ErrInvalidSSO.Wrap("nonce is required")
	}
	return nil
}

// IsExpired checks if the request has expired.
func (r *SSOAttestationRequest) IsExpired(now time.Time) bool {
	return now.After(r.ExpiresAt)
}

// ============================================================================
// SSO Attestation Response
// ============================================================================

// SSOAttestationResponse contains the result of SSO attestation.
type SSOAttestationResponse struct {
	// Success indicates if attestation was successful
	Success bool `json:"success"`

	// Attestation is the signed attestation (if successful)
	Attestation *SSOAttestation `json:"attestation,omitempty"`

	// LinkageID is the on-chain linkage ID (if created)
	LinkageID string `json:"linkage_id,omitempty"`

	// Error contains error details (if failed)
	Error *SSOAttestationError `json:"error,omitempty"`

	// ProcessedAt is when the response was generated
	ProcessedAt time.Time `json:"processed_at"`
}

// SSOAttestationError contains error details for failed attestation.
type SSOAttestationError struct {
	// Code is the error code
	Code string `json:"code"`

	// Message is the error message
	Message string `json:"message"`

	// Details contains additional error context
	Details map[string]string `json:"details,omitempty"`
}

// NewSSOAttestationSuccess creates a successful response.
func NewSSOAttestationSuccess(attestation *SSOAttestation, linkageID string) *SSOAttestationResponse {
	return &SSOAttestationResponse{
		Success:     true,
		Attestation: attestation,
		LinkageID:   linkageID,
		ProcessedAt: time.Now(),
	}
}

// NewSSOAttestationError creates an error response.
func NewSSOAttestationError(code, message string) *SSOAttestationResponse {
	return &SSOAttestationResponse{
		Success: false,
		Error: &SSOAttestationError{
			Code:    code,
			Message: message,
		},
		ProcessedAt: time.Now(),
	}
}
