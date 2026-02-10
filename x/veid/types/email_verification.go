// Package types provides types for the VEID module.
//
// VE-224: Email verification scope v1
// This file defines types for email verification with proof of control and anti-replay.
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// EmailVerificationVersion is the current version of email verification format
const EmailVerificationVersion uint32 = 1

// EmailVerificationStatus represents the status of email verification
type EmailVerificationStatus string

const (
	// EmailStatusPending indicates verification is pending
	EmailStatusPending EmailVerificationStatus = "pending"

	// EmailStatusVerified indicates email ownership is verified
	EmailStatusVerified EmailVerificationStatus = "verified"

	// EmailStatusFailed indicates verification failed
	EmailStatusFailed EmailVerificationStatus = "failed"

	// EmailStatusRevoked indicates verification was revoked
	EmailStatusRevoked EmailVerificationStatus = "revoked"

	// EmailStatusExpired indicates verification has expired
	EmailStatusExpired EmailVerificationStatus = "expired"
)

// AllEmailVerificationStatuses returns all valid verification statuses
func AllEmailVerificationStatuses() []EmailVerificationStatus {
	return []EmailVerificationStatus{
		EmailStatusPending,
		EmailStatusVerified,
		EmailStatusFailed,
		EmailStatusRevoked,
		EmailStatusExpired,
	}
}

// IsValidEmailVerificationStatus checks if a status is valid
func IsValidEmailVerificationStatus(s EmailVerificationStatus) bool {
	for _, valid := range AllEmailVerificationStatuses() {
		if s == valid {
			return true
		}
	}
	return false
}

// EmailVerificationRecord represents on-chain email verification metadata
// Note: Actual email addresses are never stored on-chain, only hashes
type EmailVerificationRecord struct {
	// Version is the format version
	Version uint32 `json:"version"`

	// VerificationID is a unique identifier for this verification
	VerificationID string `json:"verification_id"`

	// AccountAddress is the account that owns this email verification
	AccountAddress string `json:"account_address"`

	// EmailHash is a SHA256 hash of the normalized email address
	EmailHash string `json:"email_hash"`

	// DomainHash is a SHA256 hash of the email domain
	DomainHash string `json:"domain_hash"`

	// Nonce is the verification nonce (to prevent replay)
	Nonce string `json:"nonce"`

	// NonceUsedAt marks when the nonce was consumed (prevents replay)
	NonceUsedAt *time.Time `json:"nonce_used_at,omitempty"`

	// Status is the current verification status
	Status EmailVerificationStatus `json:"status"`

	// VerifiedAt is when this email was verified
	VerifiedAt *time.Time `json:"verified_at,omitempty"`

	// ExpiresAt is when this verification expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// CreatedAt is when this record was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this record was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// AccountSignature is the signature binding this verification to the account
	AccountSignature []byte `json:"account_signature"`

	// IsOrganizational indicates if this is an organizational email
	IsOrganizational bool `json:"is_organizational"`

	// VerificationAttempts counts failed verification attempts
	VerificationAttempts uint32 `json:"verification_attempts"`

	// EvidenceHash is the SHA256 hash of the verification evidence payload
	EvidenceHash string `json:"evidence_hash,omitempty"`

	// EvidenceStorageBackend indicates where the encrypted evidence is stored
	EvidenceStorageBackend string `json:"evidence_storage_backend,omitempty"`

	// EvidenceStorageRef is a backend-specific reference to the encrypted evidence
	EvidenceStorageRef string `json:"evidence_storage_ref,omitempty"`

	// EvidenceMetadata contains optional evidence metadata (non-sensitive)
	EvidenceMetadata map[string]string `json:"evidence_metadata,omitempty"`
}

// NewEmailVerificationRecord creates a new email verification record
func NewEmailVerificationRecord(
	verificationID string,
	accountAddress string,
	email string, // Will be hashed, not stored
	nonce string,
	createdAt time.Time,
) *EmailVerificationRecord {
	emailHash, domainHash := HashEmail(email)
	return &EmailVerificationRecord{
		Version:          EmailVerificationVersion,
		VerificationID:   verificationID,
		AccountAddress:   accountAddress,
		EmailHash:        emailHash,
		DomainHash:       domainHash,
		Nonce:            nonce,
		Status:           EmailStatusPending,
		CreatedAt:        createdAt,
		UpdatedAt:        createdAt,
		EvidenceMetadata: make(map[string]string),
	}
}

// HashEmail creates SHA256 hashes of an email and its domain
func HashEmail(email string) (emailHash, domainHash string) {
	// Normalize email to lowercase
	normalized := make([]byte, len(email))
	for i := range email {
		if email[i] >= 'A' && email[i] <= 'Z' {
			normalized[i] = email[i] + 32
		} else {
			normalized[i] = email[i]
		}
	}

	// Find domain separator
	domainStart := -1
	for i := range normalized {
		if normalized[i] == '@' {
			domainStart = i + 1
			break
		}
	}

	// Hash full email
	fullHash := sha256.Sum256(normalized)
	emailHash = hex.EncodeToString(fullHash[:])

	// Hash domain if found
	if domainStart > 0 && domainStart < len(normalized) {
		domainBytes := normalized[domainStart:]
		domHash := sha256.Sum256(domainBytes)
		domainHash = hex.EncodeToString(domHash[:])
	} else {
		domainHash = ""
	}

	return emailHash, domainHash
}

// Validate validates the email verification record
func (r *EmailVerificationRecord) Validate() error {
	if r.Version == 0 || r.Version > EmailVerificationVersion {
		return ErrInvalidEmail.Wrapf("unsupported version: %d", r.Version)
	}

	if r.VerificationID == "" {
		return ErrInvalidEmail.Wrap("verification_id cannot be empty")
	}

	if r.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account address cannot be empty")
	}

	if r.EmailHash == "" {
		return ErrInvalidEmail.Wrap("email_hash cannot be empty")
	}

	if len(r.EmailHash) != 64 { // SHA256 hex = 64 chars
		return ErrInvalidEmail.Wrap("email_hash must be a valid SHA256 hex string")
	}

	if r.Nonce == "" {
		return ErrInvalidEmail.Wrap("nonce cannot be empty")
	}

	if !IsValidEmailVerificationStatus(r.Status) {
		return ErrInvalidEmail.Wrapf("invalid status: %s", r.Status)
	}

	if r.CreatedAt.IsZero() {
		return ErrInvalidEmail.Wrap("created_at cannot be zero")
	}

	if err := validateEvidencePointer(r.EvidenceHash, r.EvidenceStorageBackend, r.EvidenceStorageRef, r.Status == EmailStatusVerified); err != nil {
		return ErrInvalidEmail.Wrap(err.Error())
	}

	return nil
}

// IsActive returns true if the verification is currently valid
func (r *EmailVerificationRecord) IsActive() bool {
	return r.IsActiveAt(time.Now())
}

// IsActiveAt returns true if the verification is currently valid at the provided time.
func (r *EmailVerificationRecord) IsActiveAt(now time.Time) bool {
	if r.Status != EmailStatusVerified {
		return false
	}
	if r.ExpiresAt != nil && now.After(*r.ExpiresAt) {
		return false
	}
	return true
}

// MarkVerified marks the record as verified
func (r *EmailVerificationRecord) MarkVerified(verifiedAt time.Time, expiresAt *time.Time) {
	r.Status = EmailStatusVerified
	r.VerifiedAt = &verifiedAt
	r.NonceUsedAt = &verifiedAt
	r.ExpiresAt = expiresAt
	r.UpdatedAt = verifiedAt
}

// String returns a string representation (non-sensitive)
func (r *EmailVerificationRecord) String() string {
	return fmt.Sprintf("EmailVerification{ID: %s, Status: %s, IsOrg: %t}",
		r.VerificationID, r.Status, r.IsOrganizational)
}

// EmailVerificationChallenge represents a pending email verification challenge
type EmailVerificationChallenge struct {
	// ChallengeID is a unique identifier for this challenge
	ChallengeID string `json:"challenge_id"`

	// AccountAddress is the account requesting email verification
	AccountAddress string `json:"account_address"`

	// EmailHash is a hash of the email being verified
	EmailHash string `json:"email_hash"`

	// Nonce is the one-time verification nonce
	Nonce string `json:"nonce"`

	// NonceHash is a hash of the nonce for on-chain storage
	NonceHash string `json:"nonce_hash"`

	// CreatedAt is when this challenge was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when this challenge expires
	ExpiresAt time.Time `json:"expires_at"`

	// Status is the challenge status
	Status EmailVerificationStatus `json:"status"`

	// Attempts is the number of verification attempts
	Attempts uint32 `json:"attempts"`

	// MaxAttempts is the maximum allowed attempts
	MaxAttempts uint32 `json:"max_attempts"`

	// LastAttemptAt is when the last attempt was made
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`

	// IsConsumed indicates if the nonce has been consumed
	IsConsumed bool `json:"is_consumed"`
}

// NewEmailVerificationChallenge creates a new email verification challenge
func NewEmailVerificationChallenge(
	challengeID string,
	accountAddress string,
	emailHash string,
	nonce string,
	createdAt time.Time,
	ttlSeconds int64,
	maxAttempts uint32,
) *EmailVerificationChallenge {
	expiresAt := createdAt.Add(time.Duration(ttlSeconds) * time.Second)
	nonceHashBytes := sha256.Sum256([]byte(nonce))
	nonceHash := hex.EncodeToString(nonceHashBytes[:])

	return &EmailVerificationChallenge{
		ChallengeID:    challengeID,
		AccountAddress: accountAddress,
		EmailHash:      emailHash,
		Nonce:          nonce,
		NonceHash:      nonceHash,
		CreatedAt:      createdAt,
		ExpiresAt:      expiresAt,
		Status:         EmailStatusPending,
		MaxAttempts:    maxAttempts,
	}
}

// Validate validates the email verification challenge
func (c *EmailVerificationChallenge) Validate() error {
	if c.ChallengeID == "" {
		return ErrInvalidEmail.Wrap("challenge_id cannot be empty")
	}
	if c.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account address cannot be empty")
	}
	if c.EmailHash == "" {
		return ErrInvalidEmail.Wrap("email_hash cannot be empty")
	}
	if c.Nonce == "" {
		return ErrInvalidEmail.Wrap("nonce cannot be empty")
	}
	if c.CreatedAt.IsZero() {
		return ErrInvalidEmail.Wrap("created_at cannot be zero")
	}
	if c.ExpiresAt.IsZero() {
		return ErrInvalidEmail.Wrap("expires_at cannot be zero")
	}
	if c.MaxAttempts == 0 {
		return ErrInvalidEmail.Wrap("max_attempts must be positive")
	}
	return nil
}

// IsExpired returns true if the challenge has expired
func (c *EmailVerificationChallenge) IsExpired(now time.Time) bool {
	return now.After(c.ExpiresAt)
}

// CanAttempt returns true if another attempt is allowed
func (c *EmailVerificationChallenge) CanAttempt() bool {
	return !c.IsConsumed && c.Attempts < c.MaxAttempts
}

// RecordAttempt records a verification attempt
func (c *EmailVerificationChallenge) RecordAttempt(attemptTime time.Time, success bool) {
	c.Attempts++
	c.LastAttemptAt = &attemptTime
	if success {
		c.IsConsumed = true
		c.Status = EmailStatusVerified
	} else if c.Attempts >= c.MaxAttempts {
		c.Status = EmailStatusFailed
	}
}

// VerifyNonce checks if the provided nonce matches
func (c *EmailVerificationChallenge) VerifyNonce(providedNonce string) bool {
	if c.IsConsumed {
		return false
	}
	nonceHashBytes := sha256.Sum256([]byte(providedNonce))
	providedHash := hex.EncodeToString(nonceHashBytes[:])
	return providedHash == c.NonceHash
}

// EmailScoringWeight defines the weight of email verification in VEID scoring
type EmailScoringWeight struct {
	// BaseWeight is the base score weight in basis points
	BaseWeight uint32 `json:"base_weight"`

	// OrganizationalBonus is additional weight for organizational emails
	OrganizationalBonus uint32 `json:"organizational_bonus"`

	// VerificationAgeBonus is bonus per year of verification age
	VerificationAgeBonusPerYear uint32 `json:"verification_age_bonus_per_year"`

	// MaxAgeBonus is the maximum age bonus
	MaxAgeBonus uint32 `json:"max_age_bonus"`
}

// DefaultEmailScoringWeight returns the default email scoring weight
func DefaultEmailScoringWeight() EmailScoringWeight {
	return EmailScoringWeight{
		BaseWeight:                  300, // 3% weight
		OrganizationalBonus:         200, // +2% for organizational email
		VerificationAgeBonusPerYear: 25,  // +0.25% per year
		MaxAgeBonus:                 100, // max +1% from age
	}
}

// CalculateEmailScore calculates the score contribution for an email verification
func CalculateEmailScore(record *EmailVerificationRecord, weight EmailScoringWeight, now time.Time) uint32 {
	if !record.IsActiveAt(now) {
		return 0
	}

	score := weight.BaseWeight

	if record.IsOrganizational {
		score += weight.OrganizationalBonus
	}

	if record.VerifiedAt != nil {
		ageYears := uint32(now.Sub(*record.VerifiedAt).Hours() / (24 * 365))
		ageBonus := ageYears * weight.VerificationAgeBonusPerYear
		if ageBonus > weight.MaxAgeBonus {
			ageBonus = weight.MaxAgeBonus
		}
		score += ageBonus
	}

	return score
}

// UsedNonceRecord tracks consumed nonces for anti-replay
type UsedNonceRecord struct {
	// NonceHash is the hash of the consumed nonce
	NonceHash string `json:"nonce_hash"`

	// UsedAt is when the nonce was consumed
	UsedAt time.Time `json:"used_at"`

	// AccountAddress is the account that used this nonce
	AccountAddress string `json:"account_address"`

	// VerificationID is the verification this nonce was used for
	VerificationID string `json:"verification_id"`

	// ExpiresAt is when this record can be pruned
	ExpiresAt time.Time `json:"expires_at"`
}

// NewUsedNonceRecord creates a new used nonce record
func NewUsedNonceRecord(
	nonce string,
	usedAt time.Time,
	accountAddress string,
	verificationID string,
	retentionDays int,
) *UsedNonceRecord {
	nonceHashBytes := sha256.Sum256([]byte(nonce))
	nonceHash := hex.EncodeToString(nonceHashBytes[:])
	expiresAt := usedAt.Add(time.Duration(retentionDays) * 24 * time.Hour)

	return &UsedNonceRecord{
		NonceHash:      nonceHash,
		UsedAt:         usedAt,
		AccountAddress: accountAddress,
		VerificationID: verificationID,
		ExpiresAt:      expiresAt,
	}
}

// IsNonceUsed checks if a nonce has been used (anti-replay check)
func IsNonceUsed(nonceHash string, usedNonces []UsedNonceRecord) bool {
	for _, record := range usedNonces {
		if record.NonceHash == nonceHash {
			return true
		}
	}
	return false
}
