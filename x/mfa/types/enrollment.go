package types

import (
	"time"
)

// FactorEnrollmentStatus represents the status of a factor enrollment
type FactorEnrollmentStatus uint8

const (
	// EnrollmentStatusUnspecified represents an unspecified status
	EnrollmentStatusUnspecified FactorEnrollmentStatus = 0
	// EnrollmentStatusPending represents a pending enrollment awaiting verification
	EnrollmentStatusPending FactorEnrollmentStatus = 1
	// EnrollmentStatusActive represents an active, verified enrollment
	EnrollmentStatusActive FactorEnrollmentStatus = 2
	// EnrollmentStatusRevoked represents a revoked enrollment
	EnrollmentStatusRevoked FactorEnrollmentStatus = 3
	// EnrollmentStatusExpired represents an expired enrollment
	EnrollmentStatusExpired FactorEnrollmentStatus = 4
)

// String returns the string representation of an enrollment status
func (s FactorEnrollmentStatus) String() string {
	switch s {
	case EnrollmentStatusPending:
		return "pending"
	case EnrollmentStatusActive:
		return "active"
	case EnrollmentStatusRevoked:
		return "revoked"
	case EnrollmentStatusExpired:
		return "expired"
	default:
		return "unspecified"
	}
}

// IsValid returns true if the status is valid
func (s FactorEnrollmentStatus) IsValid() bool {
	return s >= EnrollmentStatusPending && s <= EnrollmentStatusExpired
}

// FactorEnrollment represents a factor enrollment record
// NOTE: This NEVER stores raw secrets (TOTP seeds, FIDO2 private keys) on-chain
// Only public identifiers and metadata are stored
type FactorEnrollment struct {
	// AccountAddress is the account that owns this enrollment
	AccountAddress string `json:"account_address"`

	// FactorType is the type of factor
	FactorType FactorType `json:"factor_type"`

	// FactorID is a unique identifier for this factor enrollment
	// For FIDO2: credential ID (public)
	// For TOTP: hash of the seed (for verification, not the seed itself)
	// For SMS: last 4 digits of phone (masked)
	// For Email: domain and hash of local part (masked)
	// For TrustedDevice: device fingerprint hash
	FactorID string `json:"factor_id"`

	// PublicIdentifier is the public component of the factor
	// For FIDO2: public key bytes
	// For TOTP: not stored (verification happens off-chain)
	// For SMS: not stored
	// For Email: not stored
	// For VEID: threshold score
	// For TrustedDevice: device binding hash
	PublicIdentifier []byte `json:"public_identifier,omitempty"`

	// Label is a user-friendly label for the factor
	Label string `json:"label"`

	// Status is the current enrollment status
	Status FactorEnrollmentStatus `json:"status"`

	// EnrolledAt is the timestamp when the factor was enrolled
	EnrolledAt int64 `json:"enrolled_at"`

	// VerifiedAt is the timestamp when the enrollment was verified (if applicable)
	VerifiedAt int64 `json:"verified_at,omitempty"`

	// RevokedAt is the timestamp when the factor was revoked (if applicable)
	RevokedAt int64 `json:"revoked_at,omitempty"`

	// LastUsedAt is the timestamp of the last successful use
	LastUsedAt int64 `json:"last_used_at,omitempty"`

	// UseCount tracks how many times this factor has been used
	UseCount uint64 `json:"use_count"`

	// Metadata contains additional factor-specific metadata
	Metadata *FactorMetadata `json:"metadata,omitempty"`
}

// FactorMetadata contains type-specific metadata for enrollments
type FactorMetadata struct {
	// VEIDThreshold is the minimum VEID score required (for VEID factor)
	VEIDThreshold uint32 `json:"veid_threshold,omitempty"`

	// DeviceInfo contains device information (for TrustedDevice factor)
	DeviceInfo *DeviceInfo `json:"device_info,omitempty"`

	// FIDO2Info contains FIDO2-specific metadata
	FIDO2Info *FIDO2CredentialInfo `json:"fido2_info,omitempty"`

	// ContactHash contains a hash of the contact info (for SMS/Email verification tracking)
	ContactHash string `json:"contact_hash,omitempty"`

	// HardwareKeyInfo contains hardware key/X.509/smart card metadata (VE-925)
	HardwareKeyInfo *HardwareKeyEnrollment `json:"hardware_key_info,omitempty"`
}

// DeviceInfo contains information about a trusted device
type DeviceInfo struct {
	// Fingerprint is a unique hash identifying the device
	Fingerprint string `json:"fingerprint"`

	// UserAgent is the user agent string (sanitized)
	UserAgent string `json:"user_agent,omitempty"`

	// FirstSeenAt is when this device was first seen
	FirstSeenAt int64 `json:"first_seen_at"`

	// LastSeenAt is when this device was last seen
	LastSeenAt int64 `json:"last_seen_at"`

	// IPHash is a hash of the last known IP (for change detection, not tracking)
	IPHash string `json:"ip_hash,omitempty"`

	// TrustExpiresAt is when the device trust expires
	TrustExpiresAt int64 `json:"trust_expires_at"`

	// TrustTokenHash is the bcrypt hash of the trust token for this device
	TrustTokenHash string `json:"trust_token_hash,omitempty"`
}

// FIDO2CredentialInfo contains FIDO2-specific credential information
type FIDO2CredentialInfo struct {
	// CredentialID is the FIDO2 credential identifier
	CredentialID []byte `json:"credential_id"`

	// PublicKey is the COSE-encoded public key
	PublicKey []byte `json:"public_key"`

	// AAGUID is the Authenticator Attestation GUID
	AAGUID []byte `json:"aaguid,omitempty"`

	// SignCount is the signature counter for clone detection
	SignCount uint32 `json:"sign_count"`

	// AttestationType indicates the attestation type used
	AttestationType string `json:"attestation_type,omitempty"`
}

// Validate validates the factor enrollment
func (e *FactorEnrollment) Validate() error {
	if e.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account address cannot be empty")
	}

	if !e.FactorType.IsValid() {
		return ErrInvalidFactorType.Wrapf("invalid factor type: %d", e.FactorType)
	}

	if e.FactorID == "" {
		return ErrInvalidEnrollment.Wrap("factor ID cannot be empty")
	}

	if !e.Status.IsValid() {
		return ErrInvalidEnrollment.Wrapf("invalid status: %d", e.Status)
	}

	if e.EnrolledAt == 0 {
		return ErrInvalidEnrollment.Wrap("enrolled_at timestamp cannot be zero")
	}

	// Factor-specific validation
	switch e.FactorType {
	case FactorTypeFIDO2:
		if len(e.PublicIdentifier) == 0 {
			return ErrInvalidEnrollment.Wrap("FIDO2 enrollment requires public identifier")
		}
		if e.Metadata == nil || e.Metadata.FIDO2Info == nil {
			return ErrInvalidEnrollment.Wrap("FIDO2 enrollment requires FIDO2Info metadata")
		}
	case FactorTypeVEID:
		if e.Metadata == nil || e.Metadata.VEIDThreshold == 0 {
			return ErrInvalidEnrollment.Wrap("VEID enrollment requires threshold metadata")
		}
	case FactorTypeTrustedDevice:
		if e.Metadata == nil || e.Metadata.DeviceInfo == nil {
			return ErrInvalidEnrollment.Wrap("TrustedDevice enrollment requires DeviceInfo metadata")
		}
	case FactorTypeHardwareKey:
		if len(e.PublicIdentifier) == 0 {
			return ErrInvalidEnrollment.Wrap("HardwareKey enrollment requires public identifier (public key fingerprint)")
		}
		if e.Metadata == nil || e.Metadata.HardwareKeyInfo == nil {
			return ErrInvalidEnrollment.Wrap("HardwareKey enrollment requires HardwareKeyInfo metadata")
		}
		if err := e.Metadata.HardwareKeyInfo.Validate(); err != nil {
			return ErrInvalidEnrollment.Wrapf("invalid hardware key metadata: %v", err)
		}
	}

	return nil
}

// IsActive returns true if the enrollment is active
func (e *FactorEnrollment) IsActive() bool {
	return e.Status == EnrollmentStatusActive
}

// IsExpired returns true if the enrollment has expired
func (e *FactorEnrollment) IsExpired(now time.Time) bool {
	if e.Status == EnrollmentStatusExpired {
		return true
	}

	// Check device trust expiration for trusted devices
	if e.FactorType == FactorTypeTrustedDevice && e.Metadata != nil && e.Metadata.DeviceInfo != nil {
		if e.Metadata.DeviceInfo.TrustExpiresAt > 0 && now.Unix() > e.Metadata.DeviceInfo.TrustExpiresAt {
			return true
		}
	}

	return false
}

// CanVerify returns true if this enrollment can be used for verification
func (e *FactorEnrollment) CanVerify(now time.Time) bool {
	return e.IsActive() && !e.IsExpired(now)
}

// UpdateLastUsed updates the last used timestamp and increments use count
func (e *FactorEnrollment) UpdateLastUsed(timestamp int64) {
	e.LastUsedAt = timestamp
	e.UseCount++
}
