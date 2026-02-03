package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
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

	// AlgorithmVersion is the version of the algorithm used
	AlgorithmVersion uint32 `json:"algorithm_version"`

	// RecipientKeyIDs are the fingerprints of intended recipients' public keys
	// Used to identify which key can decrypt this envelope
	RecipientKeyIDs []string `json:"recipient_key_ids"`

	// RecipientPublicKeys are the public keys for intended recipients
	// Optional for on-chain storage; if provided, must align with RecipientKeyIDs
	RecipientPublicKeys [][]byte `json:"recipient_public_keys,omitempty"`

	// EncryptedKeys contains the data encryption key encrypted for each recipient
	// Index corresponds to RecipientKeyIDs index
	// For single-recipient NaCl box, this may be empty (key derived from DH)
	EncryptedKeys [][]byte `json:"encrypted_keys,omitempty"`

	// WrappedKeys contains per-recipient wrapped DEKs keyed by recipient ID
	// This is the preferred multi-recipient representation for validator enclaves
	WrappedKeys []WrappedKeyEntry `json:"wrapped_keys,omitempty"`

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
	algInfo, err := GetAlgorithmInfo(DefaultAlgorithm())
	if err != nil {
		algInfo = AlgorithmInfo{Version: AlgorithmVersionV1}
	}
	return &EncryptedPayloadEnvelope{
		Version:          EnvelopeVersion,
		AlgorithmID:      DefaultAlgorithm(),
		AlgorithmVersion: algInfo.Version,
		RecipientKeyIDs:  make([]string, 0),
		Metadata:         make(map[string]string),
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

	if e.AlgorithmVersion == 0 {
		return ErrInvalidEnvelope.Wrap("algorithm version cannot be zero")
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

	if len(e.SenderSignature) == 0 {
		return ErrInvalidEnvelope.Wrap("sender signature required")
	}

	// Validate algorithm-specific parameters
	algInfo, err := GetAlgorithmInfo(e.AlgorithmID)
	if err != nil {
		return err
	}

	if e.AlgorithmVersion != algInfo.Version {
		return ErrUnsupportedVersion.Wrapf("algorithm version %d not supported for %s (expected %d)",
			e.AlgorithmVersion, e.AlgorithmID, algInfo.Version)
	}

	if len(e.Nonce) != algInfo.NonceSize {
		return ErrInvalidEnvelope.Wrapf("nonce size mismatch: expected %d, got %d",
			algInfo.NonceSize, len(e.Nonce))
	}

	if len(e.SenderPubKey) != algInfo.KeySize {
		return ErrInvalidEnvelope.Wrapf("sender public key size mismatch: expected %d, got %d",
			algInfo.KeySize, len(e.SenderPubKey))
	}

	if len(e.RecipientPublicKeys) > 0 {
		if len(e.RecipientPublicKeys) != len(e.RecipientKeyIDs) {
			return ErrInvalidEnvelope.Wrap("recipient public keys must align with recipient key IDs")
		}
		for i, pubKey := range e.RecipientPublicKeys {
			if len(pubKey) != algInfo.KeySize {
				return ErrInvalidEnvelope.Wrapf("recipient public key size mismatch at index %d: expected %d, got %d",
					i, algInfo.KeySize, len(pubKey))
			}
			fingerprint := ComputeKeyFingerprint(pubKey)
			if fingerprint != e.RecipientKeyIDs[i] {
				return ErrInvalidEnvelope.Wrapf("recipient key id mismatch at index %d", i)
			}
		}
	}

	if len(e.EncryptedKeys) > 0 && len(e.EncryptedKeys) != len(e.RecipientKeyIDs) {
		return ErrInvalidEnvelope.Wrap("encrypted keys must align with recipient key IDs")
	}

	if len(e.WrappedKeys) > 0 {
		seenRecipients := make(map[string]bool)
		recipientSet := make(map[string]bool)
		for _, id := range e.RecipientKeyIDs {
			recipientSet[id] = true
		}
		for i, entry := range e.WrappedKeys {
			if entry.RecipientID == "" {
				return ErrInvalidEnvelope.Wrapf("wrapped key entry %d has empty recipient ID", i)
			}
			if len(entry.WrappedKey) == 0 {
				return ErrInvalidEnvelope.Wrapf("wrapped key entry %d has empty wrapped key", i)
			}
			if seenRecipients[entry.RecipientID] {
				return ErrInvalidEnvelope.Wrapf("duplicate recipient ID: %s", entry.RecipientID)
			}
			seenRecipients[entry.RecipientID] = true
			if !recipientSet[entry.RecipientID] {
				return ErrInvalidEnvelope.Wrapf("wrapped key recipient not in recipient key IDs: %s", entry.RecipientID)
			}
		}
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

	// Include algorithm version
	h.Write([]byte{byte(e.AlgorithmVersion >> 24), byte(e.AlgorithmVersion >> 16), byte(e.AlgorithmVersion >> 8), byte(e.AlgorithmVersion)})

	// Include ciphertext
	h.Write(e.Ciphertext)

	// Include nonce
	h.Write(e.Nonce)

	// Include all recipient key IDs in the provided order
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

// DeterministicBytes returns a canonical byte representation for consensus safety.
// The serialization is deterministic and stable regardless of recipient ordering.
func (e *EncryptedPayloadEnvelope) DeterministicBytes() ([]byte, error) {
	if e == nil {
		return nil, ErrInvalidEnvelope.Wrap("envelope cannot be nil")
	}

	recipientPubKeys := make(map[string][]byte)
	if len(e.RecipientPublicKeys) == len(e.RecipientKeyIDs) {
		for i, id := range e.RecipientKeyIDs {
			recipientPubKeys[id] = e.RecipientPublicKeys[i]
		}
	}

	encryptedKeys := make(map[string][]byte)
	if len(e.EncryptedKeys) == len(e.RecipientKeyIDs) {
		for i, id := range e.RecipientKeyIDs {
			encryptedKeys[id] = e.EncryptedKeys[i]
		}
	}

	wrappedKeys := make(map[string]WrappedKeyEntry)
	for _, entry := range e.WrappedKeys {
		wrappedKeys[entry.RecipientID] = entry
	}

	recipientIDs := make([]string, len(e.RecipientKeyIDs))
	copy(recipientIDs, e.RecipientKeyIDs)
	sort.Strings(recipientIDs)

	var buf bytes.Buffer
	writeUint32 := func(v uint32) {
		_ = binary.Write(&buf, binary.BigEndian, v)
	}
	writeBytes := func(b []byte) {
		writeUint32(safeUint32FromInt(len(b)))
		buf.Write(b)
	}
	writeString := func(s string) {
		writeBytes([]byte(s))
	}

	writeUint32(e.Version)
	writeString(e.AlgorithmID)
	writeUint32(e.AlgorithmVersion)
	writeBytes(e.Nonce)
	writeBytes(e.Ciphertext)
	writeBytes(e.SenderPubKey)
	writeBytes(e.SenderSignature)

	writeUint32(safeUint32FromInt(len(recipientIDs)))
	for _, id := range recipientIDs {
		writeString(id)
		writeBytes(recipientPubKeys[id])
		writeBytes(encryptedKeys[id])

		entry, ok := wrappedKeys[id]
		if ok {
			writeBytes(entry.WrappedKey)
			writeString(entry.Algorithm)
			writeBytes(entry.EphemeralPubKey)
		} else {
			writeBytes(nil)
			writeString("")
			writeBytes(nil)
		}
	}

	if len(e.Metadata) == 0 {
		writeUint32(0)
		return buf.Bytes(), nil
	}

	keys := make([]string, 0, len(e.Metadata))
	for k := range e.Metadata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	writeUint32(safeUint32FromInt(len(keys)))
	for _, k := range keys {
		writeString(k)
		writeString(e.Metadata[k])
	}

	return buf.Bytes(), nil
}

func safeUint32FromInt(value int) uint32 {
	if value < 0 {
		return 0
	}
	if value > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(value)
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
