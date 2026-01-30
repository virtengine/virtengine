// Package types provides VEID module types.
//
// This file defines signer key management, rotation policy, and revocation
// requirements for VEID verification attestation signers.
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
// Key State Constants
// ============================================================================

// SignerKeyState represents the lifecycle state of a signer key
type SignerKeyState string

const (
	// SignerKeyStateActive indicates the key is currently valid for signing
	SignerKeyStateActive SignerKeyState = "active"

	// SignerKeyStatePending indicates the key is awaiting activation
	// Used during key rotation to pre-register successor keys
	SignerKeyStatePending SignerKeyState = "pending"

	// SignerKeyStateRotating indicates the key is in rotation transition
	// Both old and new keys are valid during this period
	SignerKeyStateRotating SignerKeyState = "rotating"

	// SignerKeyStateRevoked indicates the key has been revoked
	// Attestations signed with revoked keys should be rejected
	SignerKeyStateRevoked SignerKeyState = "revoked"

	// SignerKeyStateExpired indicates the key has expired naturally
	SignerKeyStateExpired SignerKeyState = "expired"
)

// AllSignerKeyStates returns all valid key states
func AllSignerKeyStates() []SignerKeyState {
	return []SignerKeyState{
		SignerKeyStateActive,
		SignerKeyStatePending,
		SignerKeyStateRotating,
		SignerKeyStateRevoked,
		SignerKeyStateExpired,
	}
}

// IsValidSignerKeyState checks if the key state is valid
func IsValidSignerKeyState(s SignerKeyState) bool {
	for _, valid := range AllSignerKeyStates() {
		if s == valid {
			return true
		}
	}
	return false
}

// CanSign returns true if keys in this state can create valid signatures
func (s SignerKeyState) CanSign() bool {
	return s == SignerKeyStateActive || s == SignerKeyStateRotating
}

// CanVerify returns true if signatures from keys in this state should be verified
func (s SignerKeyState) CanVerify() bool {
	return s == SignerKeyStateActive || s == SignerKeyStateRotating
}

// ============================================================================
// Revocation Reason Constants
// ============================================================================

// KeyRevocationReason indicates why a key was revoked
type KeyRevocationReason string

const (
	// RevocationReasonCompromised indicates the key was compromised
	RevocationReasonCompromised KeyRevocationReason = "compromised"

	// RevocationReasonRotation indicates normal key rotation
	RevocationReasonRotation KeyRevocationReason = "rotation"

	// RevocationReasonDecommissioned indicates the signer is decommissioned
	RevocationReasonDecommissioned KeyRevocationReason = "decommissioned"

	// RevocationReasonPolicyViolation indicates policy violation by the signer
	RevocationReasonPolicyViolation KeyRevocationReason = "policy_violation"

	// RevocationReasonAdministrative indicates administrative revocation
	RevocationReasonAdministrative KeyRevocationReason = "administrative"
)

// AllRevocationReasons returns all valid revocation reasons
func AllRevocationReasons() []KeyRevocationReason {
	return []KeyRevocationReason{
		RevocationReasonCompromised,
		RevocationReasonRotation,
		RevocationReasonDecommissioned,
		RevocationReasonPolicyViolation,
		RevocationReasonAdministrative,
	}
}

// IsValidRevocationReason checks if the revocation reason is valid
func IsValidRevocationReason(r KeyRevocationReason) bool {
	for _, valid := range AllRevocationReasons() {
		if r == valid {
			return true
		}
	}
	return false
}

// ============================================================================
// Key Rotation Policy
// ============================================================================

// SignerKeyPolicy defines the key rotation and management policy
type SignerKeyPolicy struct {
	// MaxKeyAgeSeconds is the maximum age of a key before rotation is required
	// Default: 7776000 (90 days)
	MaxKeyAgeSeconds int64 `json:"max_key_age_seconds"`

	// RotationOverlapSeconds is how long old and new keys are both valid
	// This allows attestations in-flight to complete verification
	// Default: 604800 (7 days)
	RotationOverlapSeconds int64 `json:"rotation_overlap_seconds"`

	// MinRotationNoticeSeconds is the minimum notice before key rotation
	// Default: 259200 (3 days)
	MinRotationNoticeSeconds int64 `json:"min_rotation_notice_seconds"`

	// MaxPendingKeys is the maximum number of pending keys allowed
	// Default: 2
	MaxPendingKeys int `json:"max_pending_keys"`

	// RequireSuccessorKey requires a successor key before rotation
	// Default: true
	RequireSuccessorKey bool `json:"require_successor_key"`

	// AllowEmergencyRevocation allows immediate revocation without overlap
	// Default: true (for compromised keys)
	AllowEmergencyRevocation bool `json:"allow_emergency_revocation"`

	// KeyAlgorithms is the list of allowed key algorithms
	KeyAlgorithms []AttestationProofType `json:"key_algorithms"`

	// MinKeyStrength is the minimum key strength in bits
	// Default: 256 for Ed25519
	MinKeyStrength int `json:"min_key_strength"`
}

// DefaultSignerKeyPolicy returns the default key rotation policy
func DefaultSignerKeyPolicy() SignerKeyPolicy {
	return SignerKeyPolicy{
		MaxKeyAgeSeconds:         7776000,  // 90 days
		RotationOverlapSeconds:   604800,   // 7 days
		MinRotationNoticeSeconds: 259200,   // 3 days
		MaxPendingKeys:           2,
		RequireSuccessorKey:      true,
		AllowEmergencyRevocation: true,
		KeyAlgorithms: []AttestationProofType{
			ProofTypeEd25519,
			ProofTypeSecp256k1,
		},
		MinKeyStrength: 256,
	}
}

// Validate validates the signer key policy
func (p SignerKeyPolicy) Validate() error {
	if p.MaxKeyAgeSeconds <= 0 {
		return fmt.Errorf("max_key_age_seconds must be positive")
	}

	if p.RotationOverlapSeconds < 0 {
		return fmt.Errorf("rotation_overlap_seconds cannot be negative")
	}

	if p.RotationOverlapSeconds >= p.MaxKeyAgeSeconds {
		return fmt.Errorf("rotation_overlap_seconds must be less than max_key_age_seconds")
	}

	if p.MinRotationNoticeSeconds < 0 {
		return fmt.Errorf("min_rotation_notice_seconds cannot be negative")
	}

	if p.MaxPendingKeys < 0 {
		return fmt.Errorf("max_pending_keys cannot be negative")
	}

	if len(p.KeyAlgorithms) == 0 {
		return fmt.Errorf("at least one key algorithm must be specified")
	}

	for _, alg := range p.KeyAlgorithms {
		if !IsValidProofType(alg) {
			return fmt.Errorf("invalid key algorithm: %s", alg)
		}
	}

	if p.MinKeyStrength < 128 {
		return fmt.Errorf("min_key_strength must be at least 128 bits")
	}

	return nil
}

// IsAlgorithmAllowed checks if a key algorithm is allowed by the policy
func (p SignerKeyPolicy) IsAlgorithmAllowed(alg AttestationProofType) bool {
	for _, allowed := range p.KeyAlgorithms {
		if allowed == alg {
			return true
		}
	}
	return false
}

// ============================================================================
// Signer Key Info
// ============================================================================

// SignerKeyInfo represents a signer's key with metadata for rotation tracking
type SignerKeyInfo struct {
	// KeyID is the unique identifier for this key
	// Format: "<signer_id>:<sequence_number>" or "<fingerprint_prefix>"
	KeyID string `json:"key_id"`

	// Fingerprint is the SHA256 fingerprint of the public key (hex-encoded)
	Fingerprint string `json:"fingerprint"`

	// PublicKey is the raw public key bytes (base64 or hex encoded based on algorithm)
	PublicKey []byte `json:"public_key"`

	// Algorithm is the key/signature algorithm
	Algorithm AttestationProofType `json:"algorithm"`

	// State is the current lifecycle state of the key
	State SignerKeyState `json:"state"`

	// SignerID identifies the signer/validator that owns this key
	SignerID string `json:"signer_id"`

	// SequenceNumber is the key sequence for this signer (1, 2, 3...)
	SequenceNumber uint64 `json:"sequence_number"`

	// CreatedAt is when this key was created/registered
	CreatedAt time.Time `json:"created_at"`

	// ActivatedAt is when this key became active (nil if pending)
	ActivatedAt *time.Time `json:"activated_at,omitempty"`

	// ExpiresAt is when this key is scheduled to expire
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// RevokedAt is when this key was revoked (nil if not revoked)
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// RevocationReason is why the key was revoked
	RevocationReason KeyRevocationReason `json:"revocation_reason,omitempty"`

	// SuccessorKeyID is the ID of the key that replaced this one
	SuccessorKeyID string `json:"successor_key_id,omitempty"`

	// PredecessorKeyID is the ID of the key this one replaced
	PredecessorKeyID string `json:"predecessor_key_id,omitempty"`

	// Metadata contains additional key metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewSignerKeyInfo creates a new signer key info
func NewSignerKeyInfo(
	signerID string,
	publicKey []byte,
	algorithm AttestationProofType,
	sequenceNumber uint64,
	createdAt time.Time,
) *SignerKeyInfo {
	fingerprint := ComputeKeyFingerprint(publicKey)
	keyID := fmt.Sprintf("%s:%d", signerID, sequenceNumber)

	return &SignerKeyInfo{
		KeyID:          keyID,
		Fingerprint:    fingerprint,
		PublicKey:      publicKey,
		Algorithm:      algorithm,
		State:          SignerKeyStatePending,
		SignerID:       signerID,
		SequenceNumber: sequenceNumber,
		CreatedAt:      createdAt,
		Metadata:       make(map[string]string),
	}
}

// ComputeKeyFingerprint computes the SHA256 fingerprint of a public key
func ComputeKeyFingerprint(publicKey []byte) string {
	hash := sha256.Sum256(publicKey)
	return hex.EncodeToString(hash[:])
}

// Validate validates the signer key info
func (k *SignerKeyInfo) Validate() error {
	if k.KeyID == "" {
		return ErrInvalidSignerKey.Wrap("key_id is required")
	}

	if k.Fingerprint == "" {
		return ErrInvalidSignerKey.Wrap("fingerprint is required")
	}

	if len(k.Fingerprint) != 64 {
		return ErrInvalidSignerKey.Wrap("fingerprint must be 64 hex characters (SHA256)")
	}

	if _, err := hex.DecodeString(k.Fingerprint); err != nil {
		return ErrInvalidSignerKey.Wrap("fingerprint must be valid hex encoding")
	}

	if len(k.PublicKey) == 0 {
		return ErrInvalidSignerKey.Wrap("public_key is required")
	}

	if !IsValidProofType(k.Algorithm) {
		return ErrInvalidSignerKey.Wrapf("invalid algorithm: %s", k.Algorithm)
	}

	if !IsValidSignerKeyState(k.State) {
		return ErrInvalidSignerKey.Wrapf("invalid state: %s", k.State)
	}

	if k.SignerID == "" {
		return ErrInvalidSignerKey.Wrap("signer_id is required")
	}

	if k.CreatedAt.IsZero() {
		return ErrInvalidSignerKey.Wrap("created_at is required")
	}

	// Verify fingerprint matches public key
	expectedFingerprint := ComputeKeyFingerprint(k.PublicKey)
	if k.Fingerprint != expectedFingerprint {
		return ErrInvalidSignerKey.Wrap("fingerprint does not match public key")
	}

	return nil
}

// Activate activates the key
func (k *SignerKeyInfo) Activate(activatedAt time.Time, expiresAt time.Time) error {
	if k.State != SignerKeyStatePending {
		return fmt.Errorf("can only activate pending keys, current state: %s", k.State)
	}

	k.State = SignerKeyStateActive
	k.ActivatedAt = &activatedAt
	k.ExpiresAt = &expiresAt
	return nil
}

// StartRotation marks the key as rotating
func (k *SignerKeyInfo) StartRotation(successorKeyID string) error {
	if k.State != SignerKeyStateActive {
		return fmt.Errorf("can only rotate active keys, current state: %s", k.State)
	}

	k.State = SignerKeyStateRotating
	k.SuccessorKeyID = successorKeyID
	return nil
}

// Revoke revokes the key
func (k *SignerKeyInfo) Revoke(revokedAt time.Time, reason KeyRevocationReason) error {
	if k.State == SignerKeyStateRevoked {
		return fmt.Errorf("key is already revoked")
	}

	k.State = SignerKeyStateRevoked
	k.RevokedAt = &revokedAt
	k.RevocationReason = reason
	return nil
}

// Expire marks the key as expired
func (k *SignerKeyInfo) Expire() error {
	if k.State == SignerKeyStateRevoked {
		return fmt.Errorf("cannot expire revoked key")
	}

	k.State = SignerKeyStateExpired
	return nil
}

// IsActive returns true if the key is currently active for signing
func (k *SignerKeyInfo) IsActive() bool {
	return k.State.CanSign()
}

// IsExpired returns true if the key has expired
func (k *SignerKeyInfo) IsExpired(now time.Time) bool {
	if k.ExpiresAt == nil {
		return false
	}
	return now.After(*k.ExpiresAt)
}

// ShouldRotate returns true if the key should be rotated based on the policy
func (k *SignerKeyInfo) ShouldRotate(now time.Time, policy SignerKeyPolicy) bool {
	if k.State != SignerKeyStateActive {
		return false
	}

	if k.ActivatedAt == nil {
		return false
	}

	keyAge := now.Sub(*k.ActivatedAt).Seconds()
	return keyAge >= float64(policy.MaxKeyAgeSeconds-policy.MinRotationNoticeSeconds)
}

// ============================================================================
// Key Rotation Record
// ============================================================================

// KeyRotationRecord tracks a key rotation event for audit purposes
type KeyRotationRecord struct {
	// RotationID is the unique identifier for this rotation event
	RotationID string `json:"rotation_id"`

	// SignerID identifies the signer
	SignerID string `json:"signer_id"`

	// OldKeyID is the key being rotated from
	OldKeyID string `json:"old_key_id"`

	// OldKeyFingerprint is the fingerprint of the old key
	OldKeyFingerprint string `json:"old_key_fingerprint"`

	// NewKeyID is the key being rotated to
	NewKeyID string `json:"new_key_id"`

	// NewKeyFingerprint is the fingerprint of the new key
	NewKeyFingerprint string `json:"new_key_fingerprint"`

	// InitiatedAt is when the rotation was initiated
	InitiatedAt time.Time `json:"initiated_at"`

	// CompletedAt is when the rotation was completed (nil if in progress)
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// OverlapEndsAt is when the overlap period ends
	OverlapEndsAt time.Time `json:"overlap_ends_at"`

	// Status indicates the rotation status
	Status KeyRotationStatus `json:"status"`

	// Reason indicates why the rotation was performed
	Reason KeyRevocationReason `json:"reason"`

	// InitiatedBy indicates who initiated the rotation
	InitiatedBy string `json:"initiated_by"`

	// BlockHeight is the block height when rotation was initiated
	BlockHeight int64 `json:"block_height"`

	// Notes contains optional notes about the rotation
	Notes string `json:"notes,omitempty"`
}

// KeyRotationStatus represents the status of a key rotation
type KeyRotationStatus string

const (
	// RotationStatusPending indicates rotation is pending
	RotationStatusPending KeyRotationStatus = "pending"

	// RotationStatusInProgress indicates rotation is in progress
	RotationStatusInProgress KeyRotationStatus = "in_progress"

	// RotationStatusCompleted indicates rotation completed successfully
	RotationStatusCompleted KeyRotationStatus = "completed"

	// RotationStatusFailed indicates rotation failed
	RotationStatusFailed KeyRotationStatus = "failed"

	// RotationStatusCancelled indicates rotation was cancelled
	RotationStatusCancelled KeyRotationStatus = "cancelled"
)

// NewKeyRotationRecord creates a new key rotation record
func NewKeyRotationRecord(
	rotationID string,
	signerID string,
	oldKey *SignerKeyInfo,
	newKey *SignerKeyInfo,
	initiatedAt time.Time,
	overlapSeconds int64,
	reason KeyRevocationReason,
	initiatedBy string,
	blockHeight int64,
) *KeyRotationRecord {
	return &KeyRotationRecord{
		RotationID:        rotationID,
		SignerID:          signerID,
		OldKeyID:          oldKey.KeyID,
		OldKeyFingerprint: oldKey.Fingerprint,
		NewKeyID:          newKey.KeyID,
		NewKeyFingerprint: newKey.Fingerprint,
		InitiatedAt:       initiatedAt,
		OverlapEndsAt:     initiatedAt.Add(time.Duration(overlapSeconds) * time.Second),
		Status:            RotationStatusInProgress,
		Reason:            reason,
		InitiatedBy:       initiatedBy,
		BlockHeight:       blockHeight,
	}
}

// Validate validates the key rotation record
func (r *KeyRotationRecord) Validate() error {
	if r.RotationID == "" {
		return fmt.Errorf("rotation_id is required")
	}

	if r.SignerID == "" {
		return fmt.Errorf("signer_id is required")
	}

	if r.OldKeyID == "" {
		return fmt.Errorf("old_key_id is required")
	}

	if r.NewKeyID == "" {
		return fmt.Errorf("new_key_id is required")
	}

	if r.InitiatedAt.IsZero() {
		return fmt.Errorf("initiated_at is required")
	}

	if r.OverlapEndsAt.IsZero() {
		return fmt.Errorf("overlap_ends_at is required")
	}

	return nil
}

// Complete marks the rotation as completed
func (r *KeyRotationRecord) Complete(completedAt time.Time) {
	r.CompletedAt = &completedAt
	r.Status = RotationStatusCompleted
}

// Fail marks the rotation as failed
func (r *KeyRotationRecord) Fail(reason string) {
	r.Status = RotationStatusFailed
	r.Notes = reason
}

// IsOverlapPeriodActive returns true if the overlap period is still active
func (r *KeyRotationRecord) IsOverlapPeriodActive(now time.Time) bool {
	return now.Before(r.OverlapEndsAt) && r.Status == RotationStatusInProgress
}

// ============================================================================
// Signer Registry Entry
// ============================================================================

// SignerRegistryEntry represents a registered attestation signer
type SignerRegistryEntry struct {
	// SignerID is the unique identifier for the signer
	SignerID string `json:"signer_id"`

	// Name is the human-readable name of the signer
	Name string `json:"name"`

	// ValidatorAddress is the associated validator address (if applicable)
	ValidatorAddress string `json:"validator_address,omitempty"`

	// ActiveKeyID is the currently active key ID
	ActiveKeyID string `json:"active_key_id"`

	// KeyHistory contains IDs of all keys ever used by this signer
	KeyHistory []string `json:"key_history"`

	// Policy is the key management policy for this signer
	Policy SignerKeyPolicy `json:"policy"`

	// RegisteredAt is when the signer was registered
	RegisteredAt time.Time `json:"registered_at"`

	// LastRotationAt is when the last key rotation occurred
	LastRotationAt *time.Time `json:"last_rotation_at,omitempty"`

	// Active indicates if the signer is currently active
	Active bool `json:"active"`

	// DeactivatedAt is when the signer was deactivated
	DeactivatedAt *time.Time `json:"deactivated_at,omitempty"`

	// Metadata contains additional signer metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewSignerRegistryEntry creates a new signer registry entry
func NewSignerRegistryEntry(
	signerID string,
	name string,
	validatorAddress string,
	initialKeyID string,
	registeredAt time.Time,
) *SignerRegistryEntry {
	return &SignerRegistryEntry{
		SignerID:         signerID,
		Name:             name,
		ValidatorAddress: validatorAddress,
		ActiveKeyID:      initialKeyID,
		KeyHistory:       []string{initialKeyID},
		Policy:           DefaultSignerKeyPolicy(),
		RegisteredAt:     registeredAt,
		Active:           true,
		Metadata:         make(map[string]string),
	}
}

// Validate validates the signer registry entry
func (e *SignerRegistryEntry) Validate() error {
	if e.SignerID == "" {
		return ErrInvalidSignerKey.Wrap("signer_id is required")
	}

	if e.Name == "" {
		return ErrInvalidSignerKey.Wrap("name is required")
	}

	if e.ActiveKeyID == "" && e.Active {
		return ErrInvalidSignerKey.Wrap("active signer must have an active_key_id")
	}

	if e.RegisteredAt.IsZero() {
		return ErrInvalidSignerKey.Wrap("registered_at is required")
	}

	if err := e.Policy.Validate(); err != nil {
		return ErrInvalidSignerKey.Wrapf("invalid policy: %v", err)
	}

	return nil
}

// RotateKey updates the signer to use a new key
func (e *SignerRegistryEntry) RotateKey(newKeyID string, rotatedAt time.Time) {
	e.KeyHistory = append(e.KeyHistory, newKeyID)
	e.ActiveKeyID = newKeyID
	e.LastRotationAt = &rotatedAt
}

// Deactivate deactivates the signer
func (e *SignerRegistryEntry) Deactivate(deactivatedAt time.Time) {
	e.Active = false
	e.DeactivatedAt = &deactivatedAt
}
