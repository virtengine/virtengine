package v1beta4

import (
	"crypto/ed25519"
	"fmt"
)

// ProviderPublicKeyRecord stores a provider's public key with metadata.
// This is used for encrypted communication and benchmark signature verification.
type ProviderPublicKeyRecord struct {
	// PublicKey is the raw bytes of the provider's public key
	PublicKey []byte `json:"public_key" yaml:"public_key"`

	// KeyType indicates the cryptographic algorithm: "ed25519", "x25519", or "secp256k1"
	KeyType string `json:"key_type" yaml:"key_type"`

	// UpdatedAt is the block height when this key was last set or rotated
	UpdatedAt int64 `json:"updated_at" yaml:"updated_at"`

	// RotationCount tracks how many times this key has been rotated
	RotationCount uint32 `json:"rotation_count" yaml:"rotation_count"`
}

// NewProviderPublicKeyRecord creates a new ProviderPublicKeyRecord
func NewProviderPublicKeyRecord(pubKey []byte, keyType string, blockHeight int64) ProviderPublicKeyRecord {
	return ProviderPublicKeyRecord{
		PublicKey:     pubKey,
		KeyType:       keyType,
		UpdatedAt:     blockHeight,
		RotationCount: 0,
	}
}

// Validate checks if the public key record is valid
func (r ProviderPublicKeyRecord) Validate() error {
	if len(r.PublicKey) == 0 {
		return ErrInvalidPublicKey.Wrap("public key cannot be empty")
	}

	if err := ValidatePublicKeyType(r.KeyType); err != nil {
		return err
	}

	expectedLen := GetExpectedKeyLength(r.KeyType)
	if len(r.PublicKey) != expectedLen {
		return ErrInvalidPublicKey.Wrapf("expected %d bytes for %s, got %d", expectedLen, r.KeyType, len(r.PublicKey))
	}

	return nil
}

// ValidatePublicKeyType checks if the key type is supported
func ValidatePublicKeyType(keyType string) error {
	switch keyType {
	case PublicKeyTypeEd25519, PublicKeyTypeX25519, PublicKeyTypeSecp256k1:
		return nil
	default:
		return ErrInvalidPublicKeyType.Wrapf("unsupported key type: %s", keyType)
	}
}

// GetExpectedKeyLength returns the expected byte length for a given key type
func GetExpectedKeyLength(keyType string) int {
	switch keyType {
	case PublicKeyTypeEd25519:
		return ed25519.PublicKeySize // 32 bytes
	case PublicKeyTypeX25519:
		return 32 // X25519 public keys are 32 bytes
	case PublicKeyTypeSecp256k1:
		return 33 // Compressed secp256k1 public key
	default:
		return 0
	}
}

// String implements fmt.Stringer for ProviderPublicKeyRecord
func (r ProviderPublicKeyRecord) String() string {
	return fmt.Sprintf("PublicKeyRecord{Type: %s, UpdatedAt: %d, RotationCount: %d, KeyLen: %d}",
		r.KeyType, r.UpdatedAt, r.RotationCount, len(r.PublicKey))
}

// IsEmpty returns true if the record has no public key
func (r ProviderPublicKeyRecord) IsEmpty() bool {
	return len(r.PublicKey) == 0
}
