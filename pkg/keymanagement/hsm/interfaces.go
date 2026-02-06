package hsm

import (
	"context"
	"crypto"
	"time"
)

// KeyType represents the cryptographic algorithm for an HSM key.
type KeyType string

const (
	// KeyTypeEd25519 is an Ed25519 signing key.
	KeyTypeEd25519 KeyType = "ed25519"

	// KeyTypeSecp256k1 is a secp256k1 signing key (Cosmos default).
	KeyTypeSecp256k1 KeyType = "secp256k1"

	// KeyTypeX25519 is an X25519 key-agreement key.
	KeyTypeX25519 KeyType = "x25519"

	// KeyTypeP256 is a NIST P-256 signing key.
	KeyTypeP256 KeyType = "p256"
)

// KeyInfo contains metadata about a key stored in the HSM.
type KeyInfo struct {
	// Label is the human-readable key label.
	Label string

	// ID is the HSM-internal key identifier.
	ID []byte

	// Type is the cryptographic algorithm.
	Type KeyType

	// Size is the key size in bits.
	Size int

	// Extractable indicates whether the key can be exported.
	Extractable bool

	// CreatedAt is the key creation timestamp.
	CreatedAt time.Time

	// Fingerprint is the SHA-256 hex fingerprint of the public key.
	Fingerprint string
}

// HSMProvider is the interface that all HSM backends must implement.
type HSMProvider interface {
	// Connect establishes a connection to the HSM.
	Connect(ctx context.Context) error

	// Close releases HSM resources.
	Close() error

	// GenerateKey creates a new key pair in the HSM.
	GenerateKey(ctx context.Context, keyType KeyType, label string) (*KeyInfo, error)

	// ImportKey imports an existing private key into the HSM.
	ImportKey(ctx context.Context, keyType KeyType, label string, key []byte) (*KeyInfo, error)

	// GetKey retrieves key info by label.
	GetKey(ctx context.Context, label string) (*KeyInfo, error)

	// ListKeys returns all keys stored in the HSM.
	ListKeys(ctx context.Context) ([]*KeyInfo, error)

	// DeleteKey removes a key from the HSM.
	DeleteKey(ctx context.Context, label string) error

	// Sign signs data using the specified key.
	Sign(ctx context.Context, label string, data []byte) ([]byte, error)

	// GetPublicKey returns the public key for the given label.
	GetPublicKey(ctx context.Context, label string) (crypto.PublicKey, error)
}

// Signer provides signing operations for a specific key. It bridges
// the HSM interface with the standard [crypto.Signer] interface.
type Signer interface {
	crypto.Signer

	// Label returns the HSM key label.
	Label() string

	// KeyInfo returns metadata about the signing key.
	KeyInfo() *KeyInfo
}
