package types

import (
	"crypto/sha256"
	"fmt"
)

// EnvelopeVersion is the current envelope format version
const EnvelopeVersion uint32 = 1

// EncryptedPayloadEnvelope is the canonical encrypted payload structure
// for all sensitive fields stored on-chain.
//
// Design principles:
// - Supports multiple recipients (each with their own encrypted copy of the data key)
// - Includes sender signature for authenticity verification
// - Uses hybrid encryption: asymmetric for key exchange, symmetric for data
// - Nonce is unique per encryption to prevent replay attacks
type EncryptedPayloadEnvelope struct {
	// Version is the envelope format version for future compatibility
	Version uint32 `json:"version"`

	// AlgorithmID identifies the encryption algorithm used
	// e.g., "X25519-XSALSA20-POLY1305" or "AGE-X25519"
	AlgorithmID string `json:"algorithm_id"`

	// RecipientKeyIDs are the fingerprints of intended recipients' public keys
	// Used to identify which key can decrypt this envelope
	RecipientKeyIDs []string `json:"recipient_key_ids"`

	// EncryptedKeys contains the data encryption key encrypted for each recipient
	// Index corresponds to RecipientKeyIDs index
	// For single-recipient NaCl box, this may be empty (key derived from DH)
	EncryptedKeys [][]byte `json:"encrypted_keys,omitempty"`

	// Nonce is the initialization vector / nonce for encryption
	// Must be unique for each encryption with the same key
	Nonce []byte `json:"nonce"`

	// Ciphertext is the encrypted payload data
	Ciphertext []byte `json:"ciphertext"`

	// SenderSignature is the signature over hash(version || algorithm || ciphertext || nonce || recipients)
	// Used to verify authenticity without decryption
	SenderSignature []byte `json:"sender_signature"`

	// SenderPubKey is the sender's public key for signature verification
	// Also used as ephemeral public key in NaCl box scheme
	SenderPubKey []byte `json:"sender_pub_key"`

	// Metadata contains optional public or encrypted metadata
	// Keys starting with "_" are reserved for system use
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewEncryptedPayloadEnvelope creates a new envelope with defaults
func NewEncryptedPayloadEnvelope() *EncryptedPayloadEnvelope {
	return &EncryptedPayloadEnvelope{
		Version:         EnvelopeVersion,
		AlgorithmID:     DefaultAlgorithm(),
		RecipientKeyIDs: make([]string, 0),
		Metadata:        make(map[string]string),
	}
}

// Validate performs basic validation of the envelope structure
func (e *EncryptedPayloadEnvelope) Validate() error {
	if e.Version == 0 {
		return ErrInvalidEnvelope.Wrap("version cannot be zero")
	}

	if e.Version > EnvelopeVersion {
		return ErrUnsupportedVersion.Wrapf("envelope version %d not supported (max: %d)",
			e.Version, EnvelopeVersion)
	}

	if !IsAlgorithmSupported(e.AlgorithmID) {
		return ErrUnsupportedAlgorithm.Wrapf("algorithm %s is not supported", e.AlgorithmID)
	}

	if len(e.RecipientKeyIDs) == 0 {
		return ErrInvalidEnvelope.Wrap("at least one recipient required")
	}

	if len(e.Nonce) == 0 {
		return ErrInvalidEnvelope.Wrap("nonce cannot be empty")
	}

	if len(e.Ciphertext) == 0 {
		return ErrInvalidEnvelope.Wrap("ciphertext cannot be empty")
	}

	if len(e.SenderPubKey) == 0 {
		return ErrInvalidEnvelope.Wrap("sender public key required")
	}

	// Validate algorithm-specific parameters
	algInfo, err := GetAlgorithmInfo(e.AlgorithmID)
	if err != nil {
		return err
	}

	if len(e.Nonce) != algInfo.NonceSize {
		return ErrInvalidEnvelope.Wrapf("nonce size mismatch: expected %d, got %d",
			algInfo.NonceSize, len(e.Nonce))
	}

	if len(e.SenderPubKey) != algInfo.KeySize {
		return ErrInvalidEnvelope.Wrapf("sender public key size mismatch: expected %d, got %d",
			algInfo.KeySize, len(e.SenderPubKey))
	}

	return nil
}

// SigningPayload returns the bytes that should be signed/verified
// This ensures the signature covers the important envelope components
func (e *EncryptedPayloadEnvelope) SigningPayload() []byte {
	h := sha256.New()

	// Include version
	h.Write([]byte{byte(e.Version >> 24), byte(e.Version >> 16), byte(e.Version >> 8), byte(e.Version)})

	// Include algorithm
	h.Write([]byte(e.AlgorithmID))

	// Include ciphertext
	h.Write(e.Ciphertext)

	// Include nonce
	h.Write(e.Nonce)

	// Include all recipient key IDs
	for _, kid := range e.RecipientKeyIDs {
		h.Write([]byte(kid))
	}

	return h.Sum(nil)
}

// Hash returns a unique identifier for this envelope (SHA256 of signing payload + ciphertext)
func (e *EncryptedPayloadEnvelope) Hash() []byte {
	h := sha256.New()
	h.Write(e.SigningPayload())
	return h.Sum(nil)
}

// GetRecipientIndex returns the index of a recipient by key ID, or -1 if not found
func (e *EncryptedPayloadEnvelope) GetRecipientIndex(keyID string) int {
	for i, kid := range e.RecipientKeyIDs {
		if kid == keyID {
			return i
		}
	}
	return -1
}

// IsRecipient checks if a key ID is in the recipients list
func (e *EncryptedPayloadEnvelope) IsRecipient(keyID string) bool {
	return e.GetRecipientIndex(keyID) >= 0
}

// AddMetadata adds a metadata key-value pair
func (e *EncryptedPayloadEnvelope) AddMetadata(key, value string) error {
	if len(key) == 0 {
		return fmt.Errorf("metadata key cannot be empty")
	}
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return nil
}

// GetMetadata retrieves a metadata value by key
func (e *EncryptedPayloadEnvelope) GetMetadata(key string) (string, bool) {
	if e.Metadata == nil {
		return "", false
	}
	val, ok := e.Metadata[key]
	return val, ok
}

// RecipientKeyRecord represents a registered public key for receiving encrypted data
type RecipientKeyRecord struct {
	// Address is the account address that owns this key
	Address string `json:"address"`

	// PublicKey is the X25519 public key bytes
	PublicKey []byte `json:"public_key"`

	// KeyFingerprint is a unique identifier derived from the public key
	KeyFingerprint string `json:"key_fingerprint"`

	// AlgorithmID specifies which algorithm this key is for
	AlgorithmID string `json:"algorithm_id"`

	// RegisteredAt is the block time when the key was registered
	RegisteredAt int64 `json:"registered_at"`

	// RevokedAt is the block time when the key was revoked (0 if active)
	RevokedAt int64 `json:"revoked_at,omitempty"`

	// Label is an optional human-readable label for the key
	Label string `json:"label,omitempty"`
}

// IsActive returns true if the key has not been revoked
func (r *RecipientKeyRecord) IsActive() bool {
	return r.RevokedAt == 0
}

// Validate validates the recipient key record
func (r *RecipientKeyRecord) Validate() error {
	if len(r.Address) == 0 {
		return ErrInvalidAddress.Wrap("address cannot be empty")
	}

	if len(r.PublicKey) == 0 {
		return ErrInvalidPublicKey.Wrap("public key cannot be empty")
	}

	if len(r.KeyFingerprint) == 0 {
		return ErrInvalidPublicKey.Wrap("key fingerprint cannot be empty")
	}

	if !IsAlgorithmSupported(r.AlgorithmID) {
		return ErrUnsupportedAlgorithm.Wrapf("algorithm %s is not supported", r.AlgorithmID)
	}

	algInfo, err := GetAlgorithmInfo(r.AlgorithmID)
	if err != nil {
		return err
	}

	if len(r.PublicKey) != algInfo.KeySize {
		return ErrInvalidPublicKey.Wrapf("public key size mismatch: expected %d, got %d",
			algInfo.KeySize, len(r.PublicKey))
	}

	return nil
}

// ComputeKeyFingerprint computes the fingerprint for a public key
// Uses first 20 bytes of SHA256(publicKey)
func ComputeKeyFingerprint(publicKey []byte) string {
	hash := sha256.Sum256(publicKey)
	return fmt.Sprintf("%x", hash[:KeyFingerprintSize])
}
