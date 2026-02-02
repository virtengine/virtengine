package types

import (
	"fmt"
	"time"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

// ScopeType represents the type of identity scope
type ScopeType string

// Identity scope type constants
const (
	// ScopeTypeIDDocument represents government-issued ID documents (passport, driver's license, etc.)
	ScopeTypeIDDocument ScopeType = "id_document"

	// ScopeTypeSelfie represents a selfie photo for face verification
	ScopeTypeSelfie ScopeType = "selfie"

	// ScopeTypeFaceVideo represents a video for liveness detection
	ScopeTypeFaceVideo ScopeType = "face_video"

	// ScopeTypeBiometric represents biometric data (fingerprint, voice, etc.)
	ScopeTypeBiometric ScopeType = "biometric"

	// ScopeTypeSSOMetadata represents SSO provider metadata pointers
	ScopeTypeSSOMetadata ScopeType = "sso_metadata"

	// ScopeTypeEmailProof represents email verification proof
	ScopeTypeEmailProof ScopeType = "email_proof"

	// ScopeTypeSMSProof represents SMS/phone verification proof
	ScopeTypeSMSProof ScopeType = "sms_proof"

	// ScopeTypeDomainVerify represents domain ownership verification
	ScopeTypeDomainVerify ScopeType = "domain_verify"

	// ScopeTypeADSSO represents Active Directory SSO verification (VE-907)
	ScopeTypeADSSO ScopeType = "ad_sso"
)

// ScopeSchemaVersion is the current schema version for identity scopes
const ScopeSchemaVersion uint32 = 1

// Error message constants
const errMsgScopeIDEmpty = "scope_id cannot be empty"

// AllScopeTypes returns all valid scope types
func AllScopeTypes() []ScopeType {
	return []ScopeType{
		ScopeTypeIDDocument,
		ScopeTypeSelfie,
		ScopeTypeFaceVideo,
		ScopeTypeBiometric,
		ScopeTypeSSOMetadata,
		ScopeTypeEmailProof,
		ScopeTypeSMSProof,
		ScopeTypeDomainVerify,
		ScopeTypeADSSO,
	}
}

// IsValidScopeType checks if a scope type is valid
func IsValidScopeType(scopeType ScopeType) bool {
	for _, t := range AllScopeTypes() {
		if t == scopeType {
			return true
		}
	}
	return false
}

// ScopeTypeWeight returns the weight/importance of a scope type for scoring
// Higher weight = more contribution to identity score
func ScopeTypeWeight(scopeType ScopeType) uint32 {
	switch scopeType {
	case ScopeTypeIDDocument:
		return 30 // Highest weight - government ID
	case ScopeTypeFaceVideo:
		return 25 // High weight - liveness proof
	case ScopeTypeSelfie:
		return 20 // Medium-high weight - face verification
	case ScopeTypeBiometric:
		return 20 // Medium-high weight - biometric data
	case ScopeTypeDomainVerify:
		return 15 // Medium weight - domain ownership
	case ScopeTypeEmailProof:
		return 10 // Lower weight - email verification
	case ScopeTypeSMSProof:
		return 10 // Lower weight - SMS verification
	case ScopeTypeSSOMetadata:
		return 5 // Lowest weight - SSO metadata
	case ScopeTypeADSSO:
		return 12 // Medium weight - enterprise AD SSO
	default:
		return 0
	}
}

// String returns the string representation of a ScopeType
func (s ScopeType) String() string {
	return string(s)
}

// ScopeTypeDescription returns a human-readable description
func ScopeTypeDescription(scopeType ScopeType) string {
	switch scopeType {
	case ScopeTypeIDDocument:
		return "Government-issued identification document"
	case ScopeTypeSelfie:
		return "Selfie photo for face verification"
	case ScopeTypeFaceVideo:
		return "Video recording for liveness detection"
	case ScopeTypeBiometric:
		return "Biometric data (fingerprint, voice, etc.)"
	case ScopeTypeSSOMetadata:
		return "SSO provider metadata and verification pointers"
	case ScopeTypeEmailProof:
		return "Email address verification proof"
	case ScopeTypeSMSProof:
		return "Phone number verification via SMS"
	case ScopeTypeDomainVerify:
		return "Domain ownership verification"
	case ScopeTypeADSSO:
		return "Active Directory SSO verification (Azure AD, SAML, LDAP)"
	default:
		return "Unknown scope type"
	}
}

// IdentityScope represents a single piece of identity information
type IdentityScope struct {
	// ScopeID is the unique identifier for this scope
	ScopeID string `json:"scope_id"`

	// ScopeType indicates what kind of identity data this scope contains
	ScopeType ScopeType `json:"scope_type"`

	// Version is the schema version for this scope
	Version uint32 `json:"version"`

	// EncryptedPayload contains the encrypted identity data
	// Uses the EncryptedPayloadEnvelope from x/encryption module
	EncryptedPayload encryptiontypes.EncryptedPayloadEnvelope `json:"encrypted_payload"`

	// UploadMetadata contains metadata about the upload (salt, device info, signatures)
	UploadMetadata UploadMetadata `json:"upload_metadata"`

	// Status is the current verification status of this scope
	Status VerificationStatus `json:"status"`

	// UploadedAt is when this scope was uploaded
	UploadedAt time.Time `json:"uploaded_at"`

	// VerifiedAt is when this scope was verified (nil if not yet verified)
	VerifiedAt *time.Time `json:"verified_at,omitempty"`

	// ExpiresAt is when this scope expires (optional, scope-type dependent)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// Revoked indicates if this scope has been revoked
	Revoked bool `json:"revoked"`

	// RevokedAt is when this scope was revoked (nil if not revoked)
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// RevokedReason is the reason for revocation (if revoked)
	RevokedReason string `json:"revoked_reason,omitempty"`
}

// NewIdentityScope creates a new identity scope
func NewIdentityScope(
	scopeID string,
	scopeType ScopeType,
	payload encryptiontypes.EncryptedPayloadEnvelope,
	metadata UploadMetadata,
	uploadedAt time.Time,
) *IdentityScope {
	return &IdentityScope{
		ScopeID:          scopeID,
		ScopeType:        scopeType,
		Version:          ScopeSchemaVersion,
		EncryptedPayload: payload,
		UploadMetadata:   metadata,
		Status:           VerificationStatusPending,
		UploadedAt:       uploadedAt,
		Revoked:          false,
	}
}

// Validate validates the identity scope
func (s *IdentityScope) Validate() error {
	if s.ScopeID == "" {
		return ErrInvalidScope.Wrap(errMsgScopeIDEmpty)
	}

	if !IsValidScopeType(s.ScopeType) {
		return ErrInvalidScopeType.Wrapf("invalid scope type: %s", s.ScopeType)
	}

	if s.Version == 0 || s.Version > ScopeSchemaVersion {
		return ErrInvalidScopeVersion.Wrapf("unsupported version: %d", s.Version)
	}

	if err := s.EncryptedPayload.Validate(); err != nil {
		return ErrInvalidPayload.Wrap(err.Error())
	}

	if err := s.UploadMetadata.Validate(); err != nil {
		return err
	}

	if !IsValidVerificationStatus(s.Status) {
		return ErrInvalidVerificationStatus.Wrapf("invalid status: %s", s.Status)
	}

	if s.UploadedAt.IsZero() {
		return ErrInvalidScope.Wrap("uploaded_at cannot be zero")
	}

	return nil
}

// IsActive checks if the scope is active (not revoked and not expired)
func (s *IdentityScope) IsActive() bool {
	if s.Revoked {
		return false
	}

	if s.ExpiresAt != nil && time.Now().After(*s.ExpiresAt) {
		return false
	}

	return true
}

// IsVerified checks if the scope has been verified
func (s *IdentityScope) IsVerified() bool {
	return s.Status == VerificationStatusVerified && s.VerifiedAt != nil
}

// CanBeVerified checks if the scope can be verified
func (s *IdentityScope) CanBeVerified() bool {
	if s.Revoked {
		return false
	}

	switch s.Status {
	case VerificationStatusPending, VerificationStatusInProgress:
		return true
	default:
		return false
	}
}

// String returns a string representation
func (s *IdentityScope) String() string {
	return fmt.Sprintf("IdentityScope{ID: %s, Type: %s, Version: %d, Status: %s}",
		s.ScopeID, s.ScopeType, s.Version, s.Status)
}

// ScopeRef is a lightweight reference to a scope (used in IdentityRecord)
type ScopeRef struct {
	ScopeID    string             `json:"scope_id"`
	ScopeType  ScopeType          `json:"scope_type"`
	Status     VerificationStatus `json:"status"`
	UploadedAt int64              `json:"uploaded_at"` // Unix timestamp
}

// NewScopeRef creates a scope reference from a full scope
func NewScopeRef(scope *IdentityScope) ScopeRef {
	return ScopeRef{
		ScopeID:    scope.ScopeID,
		ScopeType:  scope.ScopeType,
		Status:     scope.Status,
		UploadedAt: scope.UploadedAt.Unix(),
	}
}
