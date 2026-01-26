package types

import (
	"crypto/sha256"
	"fmt"
	"sort"
)

// MultiRecipientEnvelopeVersion is the current multi-recipient envelope format version
const MultiRecipientEnvelopeVersion uint32 = 2

// RecipientMode defines how recipients are selected for encryption
type RecipientMode string

const (
	// RecipientModeFullValidatorSet encrypts to all active validators
	RecipientModeFullValidatorSet RecipientMode = "full_validator_set"

	// RecipientModeCommittee encrypts to a designated identity committee subset
	RecipientModeCommittee RecipientMode = "committee"

	// RecipientModeSpecific encrypts to specific recipients
	RecipientModeSpecific RecipientMode = "specific"
)

// WrappedKeyEntry represents a per-recipient wrapped key
type WrappedKeyEntry struct {
	// RecipientID is the unique identifier for the recipient (key fingerprint or validator address)
	RecipientID string `json:"recipient_id"`

	// WrappedKey is the data encryption key wrapped for this recipient
	WrappedKey []byte `json:"wrapped_key"`

	// Algorithm is the key wrapping algorithm used
	Algorithm string `json:"algorithm,omitempty"`

	// EphemeralPubKey is the ephemeral public key used for this recipient (if applicable)
	EphemeralPubKey []byte `json:"ephemeral_pub_key,omitempty"`
}

// MultiRecipientEnvelope extends the encrypted envelope to support
// encrypting to multiple validator enclaves for consensus recomputation.
type MultiRecipientEnvelope struct {
	// Version is the envelope format version
	Version uint32 `json:"version"`

	// AlgorithmID identifies the payload encryption algorithm
	AlgorithmID string `json:"algorithm_id"`

	// RecipientMode specifies how recipients were selected
	RecipientMode RecipientMode `json:"recipient_mode"`

	// PayloadCiphertext is the encrypted payload (symmetric encryption)
	PayloadCiphertext []byte `json:"payload_ciphertext"`

	// PayloadNonce is the nonce used for payload encryption
	PayloadNonce []byte `json:"payload_nonce"`

	// WrappedKeys contains the data encryption key wrapped for each recipient
	// Each validator can find their entry by RecipientID
	WrappedKeys []WrappedKeyEntry `json:"wrapped_keys"`

	// ClientSignature is the approved client's signature over the payload
	ClientSignature []byte `json:"client_signature"`

	// ClientID is the approved client identifier
	ClientID string `json:"client_id"`

	// UserSignature is the user's signature over the payload
	UserSignature []byte `json:"user_signature"`

	// UserPubKey is the user's public key for signature verification
	UserPubKey []byte `json:"user_pub_key"`

	// Metadata contains additional public metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// CommitteeEpoch is the committee epoch if RecipientMode is committee
	CommitteeEpoch uint64 `json:"committee_epoch,omitempty"`
}

// NewMultiRecipientEnvelope creates a new multi-recipient envelope with defaults
func NewMultiRecipientEnvelope() *MultiRecipientEnvelope {
	return &MultiRecipientEnvelope{
		Version:       MultiRecipientEnvelopeVersion,
		AlgorithmID:   DefaultAlgorithm(),
		RecipientMode: RecipientModeFullValidatorSet,
		WrappedKeys:   make([]WrappedKeyEntry, 0),
		Metadata:      make(map[string]string),
	}
}

// Validate performs validation of the multi-recipient envelope
func (e *MultiRecipientEnvelope) Validate() error {
	if e.Version == 0 {
		return ErrInvalidEnvelope.Wrap("version cannot be zero")
	}

	if e.Version > MultiRecipientEnvelopeVersion {
		return ErrUnsupportedVersion.Wrapf(
			"envelope version %d not supported (max: %d)",
			e.Version, MultiRecipientEnvelopeVersion,
		)
	}

	if !IsAlgorithmSupported(e.AlgorithmID) {
		return ErrUnsupportedAlgorithm.Wrapf("algorithm %s is not supported", e.AlgorithmID)
	}

	if len(e.PayloadCiphertext) == 0 {
		return ErrInvalidEnvelope.Wrap("payload ciphertext cannot be empty")
	}

	if len(e.PayloadNonce) == 0 {
		return ErrInvalidEnvelope.Wrap("payload nonce cannot be empty")
	}

	if len(e.WrappedKeys) == 0 {
		return ErrInvalidEnvelope.Wrap("at least one wrapped key entry required")
	}

	// Validate each wrapped key entry
	seenRecipients := make(map[string]bool)
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
	}

	if len(e.ClientSignature) == 0 {
		return ErrInvalidEnvelope.Wrap("client signature required")
	}

	if e.ClientID == "" {
		return ErrInvalidEnvelope.Wrap("client ID required")
	}

	if len(e.UserSignature) == 0 {
		return ErrInvalidEnvelope.Wrap("user signature required")
	}

	if len(e.UserPubKey) == 0 {
		return ErrInvalidEnvelope.Wrap("user public key required")
	}

	return nil
}

// GetWrappedKey returns the wrapped key for a specific recipient
func (e *MultiRecipientEnvelope) GetWrappedKey(recipientID string) ([]byte, bool) {
	for _, entry := range e.WrappedKeys {
		if entry.RecipientID == recipientID {
			return entry.WrappedKey, true
		}
	}
	return nil, false
}

// HasRecipient checks if a recipient is included in the envelope
func (e *MultiRecipientEnvelope) HasRecipient(recipientID string) bool {
	for _, entry := range e.WrappedKeys {
		if entry.RecipientID == recipientID {
			return true
		}
	}
	return false
}

// RecipientCount returns the number of recipients
func (e *MultiRecipientEnvelope) RecipientCount() int {
	return len(e.WrappedKeys)
}

// RecipientIDs returns all recipient IDs in deterministic order
func (e *MultiRecipientEnvelope) RecipientIDs() []string {
	ids := make([]string, len(e.WrappedKeys))
	for i, entry := range e.WrappedKeys {
		ids[i] = entry.RecipientID
	}
	sort.Strings(ids)
	return ids
}

// SigningPayload returns the bytes that should be signed
// Uses deterministic serialization for consensus safety
func (e *MultiRecipientEnvelope) SigningPayload() []byte {
	h := sha256.New()

	// Include version
	h.Write([]byte{byte(e.Version >> 24), byte(e.Version >> 16), byte(e.Version >> 8), byte(e.Version)})

	// Include algorithm
	h.Write([]byte(e.AlgorithmID))

	// Include recipient mode
	h.Write([]byte(e.RecipientMode))

	// Include payload ciphertext
	h.Write(e.PayloadCiphertext)

	// Include payload nonce
	h.Write(e.PayloadNonce)

	// Include recipient IDs in sorted order for determinism
	ids := e.RecipientIDs()
	for _, id := range ids {
		h.Write([]byte(id))
	}

	// Include client ID
	h.Write([]byte(e.ClientID))

	return h.Sum(nil)
}

// Hash returns a unique identifier for this envelope
func (e *MultiRecipientEnvelope) Hash() []byte {
	h := sha256.New()
	h.Write(e.SigningPayload())
	return h.Sum(nil)
}

// DeterministicBytes returns a canonical byte representation for consensus
// The serialization must be deterministic (stable ordering, no floating point)
func (e *MultiRecipientEnvelope) DeterministicBytes() ([]byte, error) {
	// Ensure wrapped keys are in sorted order by recipient ID
	sortedKeys := make([]WrappedKeyEntry, len(e.WrappedKeys))
	copy(sortedKeys, e.WrappedKeys)
	sort.Slice(sortedKeys, func(i, j int) bool {
		return sortedKeys[i].RecipientID < sortedKeys[j].RecipientID
	})

	// Create a copy with sorted keys for serialization
	sorted := &MultiRecipientEnvelope{
		Version:           e.Version,
		AlgorithmID:       e.AlgorithmID,
		RecipientMode:     e.RecipientMode,
		PayloadCiphertext: e.PayloadCiphertext,
		PayloadNonce:      e.PayloadNonce,
		WrappedKeys:       sortedKeys,
		ClientSignature:   e.ClientSignature,
		ClientID:          e.ClientID,
		UserSignature:     e.UserSignature,
		UserPubKey:        e.UserPubKey,
		CommitteeEpoch:    e.CommitteeEpoch,
	}

	// Sort metadata keys for deterministic serialization
	if len(e.Metadata) > 0 {
		sortedMeta := make(map[string]string)
		keys := make([]string, 0, len(e.Metadata))
		for k := range e.Metadata {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sortedMeta[k] = e.Metadata[k]
		}
		sorted.Metadata = sortedMeta
	}

	// For production, use a deterministic serialization format (e.g., canonical JSON or CBOR)
	// This is a simplified representation
	return sorted.Hash(), nil
}

// MultiRecipientEnvelopeBuilder helps construct multi-recipient envelopes
type MultiRecipientEnvelopeBuilder struct {
	envelope *MultiRecipientEnvelope
}

// NewMultiRecipientEnvelopeBuilder creates a new envelope builder
func NewMultiRecipientEnvelopeBuilder() *MultiRecipientEnvelopeBuilder {
	return &MultiRecipientEnvelopeBuilder{
		envelope: NewMultiRecipientEnvelope(),
	}
}

// WithAlgorithm sets the algorithm
func (b *MultiRecipientEnvelopeBuilder) WithAlgorithm(algorithmID string) *MultiRecipientEnvelopeBuilder {
	b.envelope.AlgorithmID = algorithmID
	return b
}

// WithRecipientMode sets the recipient mode
func (b *MultiRecipientEnvelopeBuilder) WithRecipientMode(mode RecipientMode) *MultiRecipientEnvelopeBuilder {
	b.envelope.RecipientMode = mode
	return b
}

// WithPayload sets the encrypted payload
func (b *MultiRecipientEnvelopeBuilder) WithPayload(ciphertext, nonce []byte) *MultiRecipientEnvelopeBuilder {
	b.envelope.PayloadCiphertext = ciphertext
	b.envelope.PayloadNonce = nonce
	return b
}

// AddRecipient adds a recipient with their wrapped key
func (b *MultiRecipientEnvelopeBuilder) AddRecipient(recipientID string, wrappedKey []byte) *MultiRecipientEnvelopeBuilder {
	b.envelope.WrappedKeys = append(b.envelope.WrappedKeys, WrappedKeyEntry{
		RecipientID: recipientID,
		WrappedKey:  wrappedKey,
	})
	return b
}

// WithClientSignature sets the client signature
func (b *MultiRecipientEnvelopeBuilder) WithClientSignature(clientID string, signature []byte) *MultiRecipientEnvelopeBuilder {
	b.envelope.ClientID = clientID
	b.envelope.ClientSignature = signature
	return b
}

// WithUserSignature sets the user signature
func (b *MultiRecipientEnvelopeBuilder) WithUserSignature(pubKey, signature []byte) *MultiRecipientEnvelopeBuilder {
	b.envelope.UserPubKey = pubKey
	b.envelope.UserSignature = signature
	return b
}

// WithMetadata adds metadata
func (b *MultiRecipientEnvelopeBuilder) WithMetadata(key, value string) *MultiRecipientEnvelopeBuilder {
	if b.envelope.Metadata == nil {
		b.envelope.Metadata = make(map[string]string)
	}
	b.envelope.Metadata[key] = value
	return b
}

// Build validates and returns the envelope
func (b *MultiRecipientEnvelopeBuilder) Build() (*MultiRecipientEnvelope, error) {
	if err := b.envelope.Validate(); err != nil {
		return nil, fmt.Errorf("envelope validation failed: %w", err)
	}
	return b.envelope, nil
}
