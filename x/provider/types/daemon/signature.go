// Package daemon provides on-chain types for provider daemon operations.
//
// VE-400: Provider Daemon key management and transaction signing
package daemon

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// DaemonSignature represents a cryptographic signature from the provider daemon
type DaemonSignature struct {
	// PublicKey is the public key used for signing (hex encoded)
	PublicKey string `json:"public_key"`

	// Signature is the cryptographic signature (hex encoded)
	Signature string `json:"signature"`

	// Algorithm is the signing algorithm used
	Algorithm string `json:"algorithm"`

	// KeyID is the identifier for the key used
	KeyID string `json:"key_id"`

	// SignedAt is when the signature was created
	SignedAt time.Time `json:"signed_at"`
}

// Validate validates the DaemonSignature
func (ds DaemonSignature) Validate() error {
	if ds.PublicKey == "" {
		return errors.New("public_key is required")
	}
	if ds.Signature == "" {
		return errors.New("signature is required")
	}
	if ds.Algorithm == "" {
		return errors.New("algorithm is required")
	}
	if ds.KeyID == "" {
		return errors.New("key_id is required")
	}
	if ds.SignedAt.IsZero() {
		return errors.New("signed_at is required")
	}

	// Validate algorithm
	if ds.Algorithm != "ed25519" && ds.Algorithm != "secp256k1" {
		return fmt.Errorf("unsupported algorithm: %s", ds.Algorithm)
	}

	// Validate hex encoding
	if _, err := hex.DecodeString(ds.PublicKey); err != nil {
		return fmt.Errorf("public_key is not valid hex: %w", err)
	}
	if _, err := hex.DecodeString(ds.Signature); err != nil {
		return fmt.Errorf("signature is not valid hex: %w", err)
	}

	return nil
}

// Verify verifies the signature against the provided message hash
func (ds DaemonSignature) Verify(messageHash []byte) error {
	pubKeyBytes, err := hex.DecodeString(ds.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to decode public key: %w", err)
	}

	sigBytes, err := hex.DecodeString(ds.Signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	switch ds.Algorithm {
	case "ed25519":
		if len(pubKeyBytes) != ed25519.PublicKeySize {
			return fmt.Errorf("invalid ed25519 public key size: %d", len(pubKeyBytes))
		}
		if len(sigBytes) != ed25519.SignatureSize {
			return fmt.Errorf("invalid ed25519 signature size: %d", len(sigBytes))
		}
		if !ed25519.Verify(pubKeyBytes, messageHash, sigBytes) {
			return errors.New("signature verification failed")
		}
		return nil

	case "secp256k1":
		// secp256k1 verification would use a different library
		// For now, return an error indicating it's not yet implemented
		return errors.New("secp256k1 verification not yet implemented")

	default:
		return fmt.Errorf("unsupported algorithm: %s", ds.Algorithm)
	}
}

// KeyRotationRecord tracks key rotation for a provider daemon
type KeyRotationRecord struct {
	// ProviderAddress is the provider's blockchain address
	ProviderAddress string `json:"provider_address"`

	// OldKeyID is the ID of the old key being rotated out
	OldKeyID string `json:"old_key_id"`

	// NewKeyID is the ID of the new key
	NewKeyID string `json:"new_key_id"`

	// OldPublicKey is the old public key (hex encoded)
	OldPublicKey string `json:"old_public_key"`

	// NewPublicKey is the new public key (hex encoded)
	NewPublicKey string `json:"new_public_key"`

	// RotatedAt is when the rotation occurred
	RotatedAt time.Time `json:"rotated_at"`

	// GracePeriodEnd is when the old key stops being valid
	GracePeriodEnd time.Time `json:"grace_period_end"`

	// RotationSignature is signed by the old key proving authorization
	RotationSignature DaemonSignature `json:"rotation_signature"`

	// BlockHeight is the block at which this was recorded
	BlockHeight int64 `json:"block_height"`
}

// Validate validates the KeyRotationRecord
func (kr KeyRotationRecord) Validate() error {
	if kr.ProviderAddress == "" {
		return errors.New("provider_address is required")
	}
	if kr.OldKeyID == "" {
		return errors.New("old_key_id is required")
	}
	if kr.NewKeyID == "" {
		return errors.New("new_key_id is required")
	}
	if kr.OldKeyID == kr.NewKeyID {
		return errors.New("old_key_id and new_key_id cannot be the same")
	}
	if kr.OldPublicKey == "" {
		return errors.New("old_public_key is required")
	}
	if kr.NewPublicKey == "" {
		return errors.New("new_public_key is required")
	}
	if kr.OldPublicKey == kr.NewPublicKey {
		return errors.New("old_public_key and new_public_key cannot be the same")
	}
	if kr.RotatedAt.IsZero() {
		return errors.New("rotated_at is required")
	}
	if kr.GracePeriodEnd.IsZero() {
		return errors.New("grace_period_end is required")
	}
	if kr.GracePeriodEnd.Before(kr.RotatedAt) {
		return errors.New("grace_period_end cannot be before rotated_at")
	}
	if err := kr.RotationSignature.Validate(); err != nil {
		return fmt.Errorf("rotation_signature: %w", err)
	}
	return nil
}

// Hash computes the hash of the rotation record content (excluding signature)
func (kr KeyRotationRecord) Hash() ([]byte, error) {
	content := fmt.Sprintf("%s:%s:%s:%s:%s:%d:%d",
		kr.ProviderAddress,
		kr.OldKeyID,
		kr.NewKeyID,
		kr.OldPublicKey,
		kr.NewPublicKey,
		kr.RotatedAt.Unix(),
		kr.GracePeriodEnd.Unix(),
	)
	hash := sha256.Sum256([]byte(content))
	return hash[:], nil
}

// IsOldKeyValid checks if the old key is still valid (within grace period)
func (kr KeyRotationRecord) IsOldKeyValid(now time.Time) bool {
	return now.Before(kr.GracePeriodEnd)
}

// ProviderDaemonKey represents a registered provider daemon key
type ProviderDaemonKey struct {
	// ProviderAddress is the provider's blockchain address
	ProviderAddress string `json:"provider_address"`

	// KeyID is the unique identifier for this key
	KeyID string `json:"key_id"`

	// PublicKey is the public key (hex encoded)
	PublicKey string `json:"public_key"`

	// Algorithm is the key algorithm
	Algorithm string `json:"algorithm"`

	// RegisteredAt is when the key was registered
	RegisteredAt time.Time `json:"registered_at"`

	// ExpiresAt is when the key expires (optional, zero means no expiry)
	ExpiresAt time.Time `json:"expires_at,omitempty"`

	// Status is the key status (active, rotated, revoked)
	Status string `json:"status"`

	// BlockHeight is the block at which this was registered
	BlockHeight int64 `json:"block_height"`
}

// Validate validates the ProviderDaemonKey
func (pk ProviderDaemonKey) Validate() error {
	if pk.ProviderAddress == "" {
		return errors.New("provider_address is required")
	}
	if pk.KeyID == "" {
		return errors.New("key_id is required")
	}
	if pk.PublicKey == "" {
		return errors.New("public_key is required")
	}
	if pk.Algorithm == "" {
		return errors.New("algorithm is required")
	}
	if pk.RegisteredAt.IsZero() {
		return errors.New("registered_at is required")
	}
	if pk.Status == "" {
		return errors.New("status is required")
	}
	if pk.Status != "active" && pk.Status != "rotated" && pk.Status != "revoked" {
		return fmt.Errorf("invalid status: %s", pk.Status)
	}
	return nil
}

// IsActive checks if the key is currently active
func (pk ProviderDaemonKey) IsActive(now time.Time) bool {
	if pk.Status != "active" {
		return false
	}
	if !pk.ExpiresAt.IsZero() && now.After(pk.ExpiresAt) {
		return false
	}
	return true
}
