// Package types provides types for the VEID module.
//
// VE-222: SSO verification scope v1
// This file defines types for SSO (Single Sign-On) verification linking
// authorized online accounts to VEID wallets.
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// SSOProviderType identifies supported SSO providers
type SSOProviderType string

const (
	// SSOProviderGoogle represents Google OAuth provider
	SSOProviderGoogle SSOProviderType = "google"

	// SSOProviderMicrosoft represents Microsoft OAuth provider
	SSOProviderMicrosoft SSOProviderType = "microsoft"

	// SSOProviderGitHub represents GitHub OAuth provider
	SSOProviderGitHub SSOProviderType = "github"

	// SSOProviderOIDC represents generic OIDC providers
	SSOProviderOIDC SSOProviderType = "oidc"

	// SSOProviderEduGAIN represents EduGAIN SAML federation providers
	SSOProviderEduGAIN SSOProviderType = "edugain"
)

// AllSSOProviderTypes returns all valid SSO provider types
func AllSSOProviderTypes() []SSOProviderType {
	return []SSOProviderType{
		SSOProviderGoogle,
		SSOProviderMicrosoft,
		SSOProviderGitHub,
		SSOProviderOIDC,
		SSOProviderEduGAIN,
	}
}

// IsValidSSOProviderType checks if a provider type is valid
func IsValidSSOProviderType(p SSOProviderType) bool {
	for _, valid := range AllSSOProviderTypes() {
		if p == valid {
			return true
		}
	}
	return false
}

// SSOVerificationVersion is the current version of the SSO verification format
const SSOVerificationVersion uint32 = 1

// SSOVerificationStatus represents the status of an SSO verification
type SSOVerificationStatus string

const (
	// SSOStatusPending indicates verification is pending
	SSOStatusPending SSOVerificationStatus = "pending"

	// SSOStatusVerified indicates verification is complete
	SSOStatusVerified SSOVerificationStatus = "verified"

	// SSOStatusFailed indicates verification failed
	SSOStatusFailed SSOVerificationStatus = "failed"

	// SSOStatusRevoked indicates verification was revoked
	SSOStatusRevoked SSOVerificationStatus = "revoked"

	// SSOStatusExpired indicates verification has expired
	SSOStatusExpired SSOVerificationStatus = "expired"
)

// AllSSOVerificationStatuses returns all valid SSO verification statuses
func AllSSOVerificationStatuses() []SSOVerificationStatus {
	return []SSOVerificationStatus{
		SSOStatusPending,
		SSOStatusVerified,
		SSOStatusFailed,
		SSOStatusRevoked,
		SSOStatusExpired,
	}
}

// IsValidSSOVerificationStatus checks if a status is valid
func IsValidSSOVerificationStatus(s SSOVerificationStatus) bool {
	for _, valid := range AllSSOVerificationStatuses() {
		if s == valid {
			return true
		}
	}
	return false
}

// SSOLinkageMetadata contains the minimal on-chain metadata for SSO linkage
// Note: No sensitive tokens or claims are stored here
type SSOLinkageMetadata struct {
	// Version is the format version
	Version uint32 `json:"version"`

	// LinkageID is a unique identifier for this linkage
	LinkageID string `json:"linkage_id"`

	// Provider identifies the SSO provider
	Provider SSOProviderType `json:"provider"`

	// Issuer is the OAuth issuer URL (e.g., "https://accounts.google.com")
	Issuer string `json:"issuer"`

	// SubjectHash is a SHA256 hash of the SSO subject identifier
	// This allows verification without exposing the raw subject ID
	SubjectHash string `json:"subject_hash"`

	// Nonce is the challenge nonce used during verification
	Nonce string `json:"nonce"`

	// VerifiedAt is when this linkage was verified
	VerifiedAt time.Time `json:"verified_at"`

	// ExpiresAt is when this linkage expires (optional)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// AccountSignature is the signature of the account binding this SSO linkage
	// This proves the user authorized this linkage to their wallet
	AccountSignature []byte `json:"account_signature"`

	// Status is the current status of this linkage
	Status SSOVerificationStatus `json:"status"`

	// EmailDomainHash is a hash of the email domain (optional, for org matching)
	EmailDomainHash string `json:"email_domain_hash,omitempty"`

	// OrgIDHash is a hash of the organization ID if applicable (Microsoft tenant, etc.)
	OrgIDHash string `json:"org_id_hash,omitempty"`

	// EvidenceHash is the SHA256 hash of the verification evidence payload
	EvidenceHash string `json:"evidence_hash,omitempty"`

	// EvidenceStorageBackend indicates where the encrypted evidence is stored
	EvidenceStorageBackend string `json:"evidence_storage_backend,omitempty"`

	// EvidenceStorageRef is a backend-specific reference to the encrypted evidence
	EvidenceStorageRef string `json:"evidence_storage_ref,omitempty"`

	// EvidenceMetadata contains optional evidence metadata (non-sensitive)
	EvidenceMetadata map[string]string `json:"evidence_metadata,omitempty"`
}

// NewSSOLinkageMetadata creates a new SSO linkage metadata record
func NewSSOLinkageMetadata(
	linkageID string,
	provider SSOProviderType,
	issuer string,
	subjectID string, // Will be hashed
	nonce string,
	verifiedAt time.Time,
) *SSOLinkageMetadata {
	return &SSOLinkageMetadata{
		Version:          SSOVerificationVersion,
		LinkageID:        linkageID,
		Provider:         provider,
		Issuer:           issuer,
		SubjectHash:      HashSubjectID(subjectID),
		Nonce:            nonce,
		VerifiedAt:       verifiedAt,
		Status:           SSOStatusVerified,
		EvidenceMetadata: make(map[string]string),
	}
}

// HashSubjectID creates a SHA256 hash of an SSO subject ID
func HashSubjectID(subjectID string) string {
	hash := sha256.Sum256([]byte(subjectID))
	return hex.EncodeToString(hash[:])
}

// HashEmailDomain creates a SHA256 hash of an email domain
func HashEmailDomain(domain string) string {
	hash := sha256.Sum256([]byte(domain))
	return hex.EncodeToString(hash[:])
}

// Validate validates the SSO linkage metadata
func (m *SSOLinkageMetadata) Validate() error {
	if m.Version == 0 || m.Version > SSOVerificationVersion {
		return ErrInvalidSSO.Wrapf("unsupported version: %d", m.Version)
	}

	if m.LinkageID == "" {
		return ErrInvalidSSO.Wrap("linkage_id cannot be empty")
	}

	if !IsValidSSOProviderType(m.Provider) {
		return ErrInvalidSSO.Wrapf("invalid provider: %s", m.Provider)
	}

	if m.Issuer == "" {
		return ErrInvalidSSO.Wrap("issuer cannot be empty")
	}

	if m.SubjectHash == "" {
		return ErrInvalidSSO.Wrap("subject_hash cannot be empty")
	}

	if len(m.SubjectHash) != 64 { // SHA256 hex = 64 chars
		return ErrInvalidSSO.Wrap("subject_hash must be a valid SHA256 hex string")
	}

	if m.Nonce == "" {
		return ErrInvalidSSO.Wrap("nonce cannot be empty")
	}

	if m.VerifiedAt.IsZero() {
		return ErrInvalidSSO.Wrap("verified_at cannot be zero")
	}

	if !IsValidSSOVerificationStatus(m.Status) {
		return ErrInvalidSSO.Wrapf("invalid status: %s", m.Status)
	}

	if err := validateEvidencePointer(m.EvidenceHash, m.EvidenceStorageBackend, m.EvidenceStorageRef, m.Status == SSOStatusVerified); err != nil {
		return ErrInvalidSSO.Wrap(err.Error())
	}

	return nil
}

// IsActive returns true if the linkage is currently valid
func (m *SSOLinkageMetadata) IsActive() bool {
	if m.Status != SSOStatusVerified {
		return false
	}
	if m.ExpiresAt != nil && time.Now().After(*m.ExpiresAt) {
		return false
	}
	return true
}

// String returns a string representation (non-sensitive)
func (m *SSOLinkageMetadata) String() string {
	return fmt.Sprintf("SSOLinkage{ID: %s, Provider: %s, Status: %s}",
		m.LinkageID, m.Provider, m.Status)
}

// SSOVerificationChallenge represents a pending SSO verification challenge
type SSOVerificationChallenge struct {
	// ChallengeID is a unique identifier for this challenge
	ChallengeID string `json:"challenge_id"`

	// AccountAddress is the account requesting SSO linkage
	AccountAddress string `json:"account_address"`

	// Provider is the SSO provider to use
	Provider SSOProviderType `json:"provider"`

	// Nonce is the challenge nonce (state parameter)
	Nonce string `json:"nonce"`

	// CreatedAt is when this challenge was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when this challenge expires
	ExpiresAt time.Time `json:"expires_at"`

	// Status indicates the challenge status
	Status SSOVerificationStatus `json:"status"`

	// RedirectURI is where the OAuth flow should redirect
	RedirectURI string `json:"redirect_uri,omitempty"`

	// CompletedAt is when this challenge was completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// NewSSOVerificationChallenge creates a new SSO verification challenge
func NewSSOVerificationChallenge(
	challengeID string,
	accountAddress string,
	provider SSOProviderType,
	nonce string,
	createdAt time.Time,
	ttlSeconds int64,
) *SSOVerificationChallenge {
	expiresAt := createdAt.Add(time.Duration(ttlSeconds) * time.Second)
	return &SSOVerificationChallenge{
		ChallengeID:    challengeID,
		AccountAddress: accountAddress,
		Provider:       provider,
		Nonce:          nonce,
		CreatedAt:      createdAt,
		ExpiresAt:      expiresAt,
		Status:         SSOStatusPending,
	}
}

// Validate validates the SSO verification challenge
func (c *SSOVerificationChallenge) Validate() error {
	if c.ChallengeID == "" {
		return ErrInvalidSSO.Wrap("challenge_id cannot be empty")
	}
	if c.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account address cannot be empty")
	}
	if !IsValidSSOProviderType(c.Provider) {
		return ErrInvalidSSO.Wrapf("invalid provider: %s", c.Provider)
	}
	if c.Nonce == "" {
		return ErrInvalidSSO.Wrap("nonce cannot be empty")
	}
	if c.CreatedAt.IsZero() {
		return ErrInvalidSSO.Wrap("created_at cannot be zero")
	}
	if c.ExpiresAt.IsZero() {
		return ErrInvalidSSO.Wrap("expires_at cannot be zero")
	}
	return nil
}

// IsExpired returns true if the challenge has expired
func (c *SSOVerificationChallenge) IsExpired(now time.Time) bool {
	return now.After(c.ExpiresAt)
}

// SSOScoringWeight defines the weight of SSO verification in VEID scoring
type SSOScoringWeight struct {
	// Provider is the SSO provider type
	Provider SSOProviderType `json:"provider"`

	// Weight is the score weight in basis points (out of 10000)
	Weight uint32 `json:"weight"`

	// MinVerificationAge is the minimum age of verification to count
	MinVerificationAgeSeconds int64 `json:"min_verification_age_seconds,omitempty"`

	// RequireOrgMatch if true, requires organizational email/domain match
	RequireOrgMatch bool `json:"require_org_match,omitempty"`
}

// DefaultSSOScoringWeights returns default scoring weights for SSO providers
func DefaultSSOScoringWeights() []SSOScoringWeight {
	return []SSOScoringWeight{
		{Provider: SSOProviderGoogle, Weight: 250},    // 2.5% weight
		{Provider: SSOProviderMicrosoft, Weight: 300}, // 3.0% weight (enterprise)
		{Provider: SSOProviderGitHub, Weight: 200},    // 2.0% weight
		{Provider: SSOProviderOIDC, Weight: 150},      // 1.5% weight (generic)
		{Provider: SSOProviderEduGAIN, Weight: 350},   // 3.5% weight (SAML federation)
	}
}

// GetSSOScoringWeight returns the scoring weight for a provider
func GetSSOScoringWeight(provider SSOProviderType) uint32 {
	weights := DefaultSSOScoringWeights()
	for _, w := range weights {
		if w.Provider == provider {
			return w.Weight
		}
	}
	return 0
}
