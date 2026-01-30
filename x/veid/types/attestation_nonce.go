// Package types provides VEID module types.
//
// This file defines replay protection and nonce tracking for VEID attestations.
//
// Task Reference: VE-1B - Verification Attestation Schema
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// ============================================================================
// Nonce Constants
// ============================================================================

const (
	// NonceMinLength is the minimum nonce length in bytes
	NonceMinLength = 16

	// NonceMaxLength is the maximum nonce length in bytes
	NonceMaxLength = 64

	// NonceDefaultLength is the default nonce length in bytes
	NonceDefaultLength = 32

	// NonceWindowDefaultSeconds is the default nonce validity window
	NonceWindowDefaultSeconds int64 = 3600 // 1 hour

	// NonceWindowMaxSeconds is the maximum nonce validity window
	NonceWindowMaxSeconds int64 = 86400 // 24 hours

	// NonceWindowMinSeconds is the minimum nonce validity window
	NonceWindowMinSeconds int64 = 300 // 5 minutes

	// DefaultMaxNoncesPerIssuer is the default max tracked nonces per issuer
	DefaultMaxNoncesPerIssuer = 10000

	// NonceCleanupBatchSize is the batch size for nonce cleanup operations
	NonceCleanupBatchSize = 1000
)

// ============================================================================
// Nonce Status
// ============================================================================

// NonceStatus represents the usage status of a nonce
type NonceStatus string

const (
	// NonceStatusUnused indicates the nonce has not been used
	NonceStatusUnused NonceStatus = "unused"

	// NonceStatusUsed indicates the nonce has been used
	NonceStatusUsed NonceStatus = "used"

	// NonceStatusExpired indicates the nonce has expired without use
	NonceStatusExpired NonceStatus = "expired"

	// NonceStatusInvalid indicates the nonce is invalid
	NonceStatusInvalid NonceStatus = "invalid"
)

// ============================================================================
// Replay Protection Policy
// ============================================================================

// ReplayProtectionPolicy defines the replay protection requirements
type ReplayProtectionPolicy struct {
	// NonceWindowSeconds is how long a nonce is valid
	NonceWindowSeconds int64 `json:"nonce_window_seconds"`

	// RequireTimestampBinding requires nonces to be bound to timestamps
	RequireTimestampBinding bool `json:"require_timestamp_binding"`

	// MaxClockSkewSeconds is the maximum allowed clock skew
	MaxClockSkewSeconds int64 `json:"max_clock_skew_seconds"`

	// MaxNoncesPerIssuer is the maximum tracked nonces per issuer
	MaxNoncesPerIssuer int `json:"max_nonces_per_issuer"`

	// RequireIssuerBinding requires nonces to be bound to issuers
	RequireIssuerBinding bool `json:"require_issuer_binding"`

	// RequireSubjectBinding requires nonces to be bound to subjects
	RequireSubjectBinding bool `json:"require_subject_binding"`

	// AllowNonceReuse allows nonce reuse after expiry (not recommended)
	AllowNonceReuse bool `json:"allow_nonce_reuse"`

	// TrackNonceHistory keeps history of used nonces (for audit)
	TrackNonceHistory bool `json:"track_nonce_history"`
}

// DefaultReplayProtectionPolicy returns the default replay protection policy
func DefaultReplayProtectionPolicy() ReplayProtectionPolicy {
	return ReplayProtectionPolicy{
		NonceWindowSeconds:      NonceWindowDefaultSeconds,
		RequireTimestampBinding: true,
		MaxClockSkewSeconds:     300, // 5 minutes
		MaxNoncesPerIssuer:      DefaultMaxNoncesPerIssuer,
		RequireIssuerBinding:    true,
		RequireSubjectBinding:   false, // Optional for batch attestations
		AllowNonceReuse:         false,
		TrackNonceHistory:       true,
	}
}

// Validate validates the replay protection policy
func (p ReplayProtectionPolicy) Validate() error {
	if p.NonceWindowSeconds < NonceWindowMinSeconds {
		return fmt.Errorf("nonce_window_seconds must be at least %d", NonceWindowMinSeconds)
	}

	if p.NonceWindowSeconds > NonceWindowMaxSeconds {
		return fmt.Errorf("nonce_window_seconds cannot exceed %d", NonceWindowMaxSeconds)
	}

	if p.MaxClockSkewSeconds < 0 {
		return fmt.Errorf("max_clock_skew_seconds cannot be negative")
	}

	if p.MaxClockSkewSeconds > p.NonceWindowSeconds/2 {
		return fmt.Errorf("max_clock_skew_seconds should not exceed half of nonce_window_seconds")
	}

	if p.MaxNoncesPerIssuer <= 0 {
		return fmt.Errorf("max_nonces_per_issuer must be positive")
	}

	return nil
}

// ============================================================================
// Nonce Record
// ============================================================================

// NonceRecord tracks the usage of an attestation nonce
type NonceRecord struct {
	// NonceHash is the SHA256 hash of the nonce (for efficient lookup)
	NonceHash string `json:"nonce_hash"`

	// IssuerFingerprint is the issuer key fingerprint (for issuer binding)
	IssuerFingerprint string `json:"issuer_fingerprint"`

	// SubjectAddress is the subject address (for subject binding, optional)
	SubjectAddress string `json:"subject_address,omitempty"`

	// AttestationType is the type of attestation this nonce was used for
	AttestationType AttestationType `json:"attestation_type"`

	// Status is the current status of the nonce
	Status NonceStatus `json:"status"`

	// CreatedAt is when the nonce record was created
	CreatedAt time.Time `json:"created_at"`

	// UsedAt is when the nonce was used (nil if unused)
	UsedAt *time.Time `json:"used_at,omitempty"`

	// ExpiresAt is when the nonce expires
	ExpiresAt time.Time `json:"expires_at"`

	// AttestationID is the ID of the attestation that used this nonce
	AttestationID string `json:"attestation_id,omitempty"`

	// BlockHeight is the block height when the nonce was used
	BlockHeight int64 `json:"block_height,omitempty"`
}

// NewNonceRecord creates a new nonce record
func NewNonceRecord(
	nonce []byte,
	issuerFingerprint string,
	attestationType AttestationType,
	createdAt time.Time,
	windowSeconds int64,
) *NonceRecord {
	return &NonceRecord{
		NonceHash:         ComputeNonceHash(nonce),
		IssuerFingerprint: issuerFingerprint,
		AttestationType:   attestationType,
		Status:            NonceStatusUnused,
		CreatedAt:         createdAt,
		ExpiresAt:         createdAt.Add(time.Duration(windowSeconds) * time.Second),
	}
}

// ComputeNonceHash computes the SHA256 hash of a nonce
func ComputeNonceHash(nonce []byte) string {
	hash := sha256.Sum256(nonce)
	return hex.EncodeToString(hash[:])
}

// Validate validates the nonce record
func (r *NonceRecord) Validate() error {
	if r.NonceHash == "" {
		return ErrInvalidNonce.Wrap("nonce_hash is required")
	}

	if len(r.NonceHash) != 64 {
		return ErrInvalidNonce.Wrap("nonce_hash must be 64 hex characters (SHA256)")
	}

	if _, err := hex.DecodeString(r.NonceHash); err != nil {
		return ErrInvalidNonce.Wrap("nonce_hash must be valid hex encoding")
	}

	if r.IssuerFingerprint == "" {
		return ErrInvalidNonce.Wrap("issuer_fingerprint is required")
	}

	if !IsValidAttestationType(r.AttestationType) {
		return ErrInvalidNonce.Wrapf("invalid attestation_type: %s", r.AttestationType)
	}

	if r.CreatedAt.IsZero() {
		return ErrInvalidNonce.Wrap("created_at is required")
	}

	if r.ExpiresAt.IsZero() {
		return ErrInvalidNonce.Wrap("expires_at is required")
	}

	if !r.ExpiresAt.After(r.CreatedAt) {
		return ErrInvalidNonce.Wrap("expires_at must be after created_at")
	}

	return nil
}

// MarkUsed marks the nonce as used
func (r *NonceRecord) MarkUsed(usedAt time.Time, attestationID string, blockHeight int64) error {
	if r.Status == NonceStatusUsed {
		return ErrNonceAlreadyUsed.Wrap("nonce has already been used")
	}

	if r.Status == NonceStatusExpired {
		return ErrNonceExpired.Wrap("nonce has expired")
	}

	if usedAt.After(r.ExpiresAt) {
		return ErrNonceExpired.Wrap("nonce has expired at time of use")
	}

	r.Status = NonceStatusUsed
	r.UsedAt = &usedAt
	r.AttestationID = attestationID
	r.BlockHeight = blockHeight
	return nil
}

// MarkExpired marks the nonce as expired
func (r *NonceRecord) MarkExpired() {
	if r.Status == NonceStatusUnused {
		r.Status = NonceStatusExpired
	}
}

// IsExpired checks if the nonce has expired
func (r *NonceRecord) IsExpired(now time.Time) bool {
	return now.After(r.ExpiresAt)
}

// IsUsed checks if the nonce has been used
func (r *NonceRecord) IsUsed() bool {
	return r.Status == NonceStatusUsed
}

// CanBeUsed checks if the nonce can be used
func (r *NonceRecord) CanBeUsed(now time.Time) bool {
	return r.Status == NonceStatusUnused && !r.IsExpired(now)
}

// ============================================================================
// Nonce Validation
// ============================================================================

// NonceValidationRequest contains parameters for nonce validation
type NonceValidationRequest struct {
	// Nonce is the raw nonce bytes
	Nonce []byte `json:"nonce"`

	// IssuerFingerprint is the expected issuer fingerprint
	IssuerFingerprint string `json:"issuer_fingerprint"`

	// SubjectAddress is the expected subject address (optional)
	SubjectAddress string `json:"subject_address,omitempty"`

	// AttestationType is the attestation type
	AttestationType AttestationType `json:"attestation_type"`

	// Timestamp is the attestation timestamp
	Timestamp time.Time `json:"timestamp"`
}

// ValidateNonce validates a nonce for replay protection
// This function validates the nonce format; actual uniqueness checking
// requires querying the nonce store.
func ValidateNonce(nonce []byte) error {
	if len(nonce) < NonceMinLength {
		return ErrInvalidNonce.Wrapf("nonce must be at least %d bytes", NonceMinLength)
	}

	if len(nonce) > NonceMaxLength {
		return ErrInvalidNonce.Wrapf("nonce cannot exceed %d bytes", NonceMaxLength)
	}

	// Check for weak nonces (all zeros or all ones)
	allZero := true
	allOne := true
	for _, b := range nonce {
		if b != 0 {
			allZero = false
		}
		if b != 0xFF {
			allOne = false
		}
	}

	if allZero {
		return ErrInvalidNonce.Wrap("nonce cannot be all zeros")
	}

	if allOne {
		return ErrInvalidNonce.Wrap("nonce cannot be all ones")
	}

	return nil
}

// ValidateNonceHex validates a hex-encoded nonce
func ValidateNonceHex(nonceHex string) error {
	nonce, err := hex.DecodeString(nonceHex)
	if err != nil {
		return ErrInvalidNonce.Wrap("nonce must be valid hex encoding")
	}

	return ValidateNonce(nonce)
}

// ============================================================================
// Nonce Validation Result
// ============================================================================

// NonceValidationResult contains the result of nonce validation
type NonceValidationResult struct {
	// Valid indicates if the nonce is valid for use
	Valid bool `json:"valid"`

	// NonceHash is the computed nonce hash
	NonceHash string `json:"nonce_hash"`

	// Error contains the error message if validation failed
	Error string `json:"error,omitempty"`

	// ExistingRecord contains the existing record if nonce was already used
	ExistingRecord *NonceRecord `json:"existing_record,omitempty"`

	// TimestampValid indicates if the timestamp is within bounds
	TimestampValid bool `json:"timestamp_valid"`

	// IssuerBindingValid indicates if issuer binding is valid
	IssuerBindingValid bool `json:"issuer_binding_valid"`
}

// NewNonceValidationResultSuccess creates a successful validation result
func NewNonceValidationResultSuccess(nonceHash string) *NonceValidationResult {
	return &NonceValidationResult{
		Valid:              true,
		NonceHash:          nonceHash,
		TimestampValid:     true,
		IssuerBindingValid: true,
	}
}

// NewNonceValidationResultFailure creates a failed validation result
func NewNonceValidationResultFailure(nonceHash string, err string) *NonceValidationResult {
	return &NonceValidationResult{
		Valid:     false,
		NonceHash: nonceHash,
		Error:     err,
	}
}

// ============================================================================
// Timestamp Validation
// ============================================================================

// ValidateTimestamp validates an attestation timestamp
func ValidateTimestamp(
	timestamp time.Time,
	now time.Time,
	policy ReplayProtectionPolicy,
) error {
	if !policy.RequireTimestampBinding {
		return nil
	}

	// Check if timestamp is in the future (with clock skew allowance)
	maxFuture := now.Add(time.Duration(policy.MaxClockSkewSeconds) * time.Second)
	if timestamp.After(maxFuture) {
		return ErrInvalidTimestamp.Wrap("timestamp is too far in the future")
	}

	// Check if timestamp is too old (within nonce window + clock skew)
	minPast := now.Add(-time.Duration(policy.NonceWindowSeconds+policy.MaxClockSkewSeconds) * time.Second)
	if timestamp.Before(minPast) {
		return ErrInvalidTimestamp.Wrap("timestamp is too far in the past")
	}

	return nil
}

// ============================================================================
// Nonce History Entry
// ============================================================================

// NonceHistoryEntry is a compact entry for nonce usage history (for audit)
type NonceHistoryEntry struct {
	// NonceHash is the SHA256 hash of the nonce
	NonceHash string `json:"nonce_hash"`

	// IssuerFingerprint is the issuer key fingerprint (first 16 chars)
	IssuerFingerprint string `json:"issuer_fingerprint"`

	// AttestationID is the attestation that used this nonce
	AttestationID string `json:"attestation_id"`

	// UsedAt is when the nonce was used
	UsedAt time.Time `json:"used_at"`

	// BlockHeight is the block height
	BlockHeight int64 `json:"block_height"`
}

// NewNonceHistoryEntry creates a new nonce history entry from a nonce record
func NewNonceHistoryEntry(record *NonceRecord) *NonceHistoryEntry {
	if record.UsedAt == nil {
		return nil
	}

	issuerFP := record.IssuerFingerprint
	if len(issuerFP) > 16 {
		issuerFP = issuerFP[:16]
	}

	return &NonceHistoryEntry{
		NonceHash:         record.NonceHash,
		IssuerFingerprint: issuerFP,
		AttestationID:     record.AttestationID,
		UsedAt:            *record.UsedAt,
		BlockHeight:       record.BlockHeight,
	}
}
