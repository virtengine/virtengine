package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

// SocialMediaScopeVersion is the current version of social media scope format.
const SocialMediaScopeVersion uint32 = 1

// SocialMediaProviderType identifies supported social media providers.
type SocialMediaProviderType string

const (
	// SocialMediaProviderGoogle represents Google social profile data.
	SocialMediaProviderGoogle SocialMediaProviderType = "google"

	// SocialMediaProviderFacebook represents Facebook social profile data.
	SocialMediaProviderFacebook SocialMediaProviderType = "facebook"

	// SocialMediaProviderMicrosoft represents Microsoft social profile data.
	SocialMediaProviderMicrosoft SocialMediaProviderType = "microsoft"
)

// AllSocialMediaProviders returns all valid social media providers.
func AllSocialMediaProviders() []SocialMediaProviderType {
	return []SocialMediaProviderType{
		SocialMediaProviderGoogle,
		SocialMediaProviderFacebook,
		SocialMediaProviderMicrosoft,
	}
}

// IsValidSocialMediaProvider checks if a provider is valid.
func IsValidSocialMediaProvider(p SocialMediaProviderType) bool {
	for _, valid := range AllSocialMediaProviders() {
		if p == valid {
			return true
		}
	}
	return false
}

// SocialMediaVerificationStatus represents the status of a social media verification.
type SocialMediaVerificationStatus string

const (
	// SocialMediaStatusPending indicates verification is pending.
	SocialMediaStatusPending SocialMediaVerificationStatus = "pending"

	// SocialMediaStatusVerified indicates verification is complete.
	SocialMediaStatusVerified SocialMediaVerificationStatus = "verified"

	// SocialMediaStatusFailed indicates verification failed.
	SocialMediaStatusFailed SocialMediaVerificationStatus = "failed"

	// SocialMediaStatusRevoked indicates verification was revoked.
	SocialMediaStatusRevoked SocialMediaVerificationStatus = "revoked"

	// SocialMediaStatusExpired indicates verification has expired.
	SocialMediaStatusExpired SocialMediaVerificationStatus = "expired"
)

// AllSocialMediaVerificationStatuses returns all valid verification statuses.
func AllSocialMediaVerificationStatuses() []SocialMediaVerificationStatus {
	return []SocialMediaVerificationStatus{
		SocialMediaStatusPending,
		SocialMediaStatusVerified,
		SocialMediaStatusFailed,
		SocialMediaStatusRevoked,
		SocialMediaStatusExpired,
	}
}

// IsValidSocialMediaVerificationStatus checks if a status is valid.
func IsValidSocialMediaVerificationStatus(s SocialMediaVerificationStatus) bool {
	for _, valid := range AllSocialMediaVerificationStatuses() {
		if s == valid {
			return true
		}
	}
	return false
}

// SocialMediaScope represents on-chain social media metadata.
// Only hashed PII is stored; full payloads are encrypted.
type SocialMediaScope struct {
	// Version is the schema version.
	Version uint32 `json:"version"`

	// ScopeID is the unique scope identifier.
	ScopeID string `json:"scope_id"`

	// AccountAddress is the owner of this scope.
	AccountAddress string `json:"account_address"`

	// Provider is the social media provider.
	Provider SocialMediaProviderType `json:"provider"`

	// ProfileNameHash is a SHA256 hash of the display name.
	ProfileNameHash string `json:"profile_name_hash"`

	// EmailHash is a SHA256 hash of the email address (optional).
	EmailHash string `json:"email_hash,omitempty"`

	// UsernameHash is a SHA256 hash of the username/handle (optional).
	UsernameHash string `json:"username_hash,omitempty"`

	// OrgHash is a SHA256 hash of the org/tenant membership (optional).
	OrgHash string `json:"org_hash,omitempty"`

	// AccountCreatedAt is when the social media account was created (optional).
	AccountCreatedAt *time.Time `json:"account_created_at,omitempty"`

	// AccountAgeDays is the account age in days at capture time.
	AccountAgeDays uint32 `json:"account_age_days"`

	// IsVerified indicates a verified/badged account.
	IsVerified bool `json:"is_verified"`

	// FriendCountRange indicates friend count bracket (Facebook) if available.
	FriendCountRange string `json:"friend_count_range,omitempty"`

	// Status is the verification status of the scope.
	Status SocialMediaVerificationStatus `json:"status"`

	// CreatedAt is when the record was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the record was last updated.
	UpdatedAt time.Time `json:"updated_at"`

	// EncryptedPayload contains encrypted social media profile data.
	EncryptedPayload encryptiontypes.EncryptedPayloadEnvelope `json:"encrypted_payload"`

	// EvidenceHash is the SHA256 hash of the attestation evidence.
	EvidenceHash string `json:"evidence_hash,omitempty"`

	// EvidenceStorageBackend indicates where encrypted evidence is stored.
	EvidenceStorageBackend string `json:"evidence_storage_backend,omitempty"`

	// EvidenceStorageRef is a backend-specific reference to the evidence payload.
	EvidenceStorageRef string `json:"evidence_storage_ref,omitempty"`

	// EvidenceMetadata contains optional non-sensitive metadata.
	EvidenceMetadata map[string]string `json:"evidence_metadata,omitempty"`
}

// NewSocialMediaScope creates a new social media scope.
func NewSocialMediaScope(
	scopeID string,
	accountAddress string,
	provider SocialMediaProviderType,
	profileNameHash string,
	encryptedPayload encryptiontypes.EncryptedPayloadEnvelope,
	now time.Time,
) *SocialMediaScope {
	return &SocialMediaScope{
		Version:          SocialMediaScopeVersion,
		ScopeID:          scopeID,
		AccountAddress:   accountAddress,
		Provider:         provider,
		ProfileNameHash:  profileNameHash,
		Status:           SocialMediaStatusVerified,
		CreatedAt:        now,
		UpdatedAt:        now,
		EncryptedPayload: encryptedPayload,
		EvidenceMetadata: make(map[string]string),
	}
}

// Validate validates the social media scope record.
func (s *SocialMediaScope) Validate() error {
	if s.Version == 0 || s.Version > SocialMediaScopeVersion {
		return ErrInvalidScope.Wrapf("unsupported version: %d", s.Version)
	}

	if s.ScopeID == "" {
		return ErrInvalidScope.Wrap("scope_id cannot be empty")
	}

	if s.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account address cannot be empty")
	}

	if !IsValidSocialMediaProvider(s.Provider) {
		return ErrInvalidScope.Wrapf("invalid provider: %s", s.Provider)
	}

	if s.ProfileNameHash == "" {
		return ErrInvalidScope.Wrap("profile_name_hash cannot be empty")
	}

	if err := validateSHA256Hex(s.ProfileNameHash, "profile_name_hash"); err != nil {
		return ErrInvalidScope.Wrap(err.Error())
	}

	if s.EmailHash != "" {
		if err := validateSHA256Hex(s.EmailHash, "email_hash"); err != nil {
			return ErrInvalidScope.Wrap(err.Error())
		}
	}

	if s.UsernameHash != "" {
		if err := validateSHA256Hex(s.UsernameHash, "username_hash"); err != nil {
			return ErrInvalidScope.Wrap(err.Error())
		}
	}

	if s.OrgHash != "" {
		if err := validateSHA256Hex(s.OrgHash, "org_hash"); err != nil {
			return ErrInvalidScope.Wrap(err.Error())
		}
	}

	if s.AccountCreatedAt == nil && s.AccountAgeDays == 0 {
		return ErrInvalidScope.Wrap("account age or creation time must be provided")
	}

	if !IsValidSocialMediaVerificationStatus(s.Status) {
		return ErrInvalidScope.Wrapf("invalid status: %s", s.Status)
	}

	if s.CreatedAt.IsZero() {
		return ErrInvalidScope.Wrap("created_at cannot be zero")
	}

	if s.UpdatedAt.IsZero() {
		return ErrInvalidScope.Wrap("updated_at cannot be zero")
	}

	if err := s.EncryptedPayload.Validate(); err != nil {
		return ErrInvalidPayload.Wrap(err.Error())
	}

	if err := validateEvidencePointer(s.EvidenceHash, s.EvidenceStorageBackend, s.EvidenceStorageRef, s.Status == SocialMediaStatusVerified); err != nil {
		return ErrInvalidScope.Wrap(err.Error())
	}

	return nil
}

// HashSocialMediaField hashes a social media field with normalization.
func HashSocialMediaField(value string) string {
	normalized := normalizeSocialMediaValue(value)
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

func normalizeSocialMediaValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validateSHA256Hex(value string, fieldName string) error {
	if value == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	if len(value) != 64 {
		return fmt.Errorf("%s must be a valid SHA256 hex string", fieldName)
	}
	if _, err := hex.DecodeString(value); err != nil {
		return fmt.Errorf("%s must be valid hex", fieldName)
	}
	return nil
}
