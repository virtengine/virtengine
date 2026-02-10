package types

import (
	"encoding/hex"
	"fmt"
)

// EnvelopeMetadata provides introspection into an envelope without decryption.
// This is safe to expose in public queries and CLI output.
type EnvelopeMetadata struct {
	// Version is the envelope format version
	Version uint32 `json:"version"`

	// AlgorithmID identifies the encryption algorithm
	AlgorithmID string `json:"algorithm_id"`

	// AlgorithmVersion is the algorithm version
	AlgorithmVersion uint32 `json:"algorithm_version"`

	// RecipientKeyIDs lists the recipients (fingerprints only)
	RecipientKeyIDs []string `json:"recipient_key_ids"`

	// EnvelopeHash is the hash of the entire envelope
	EnvelopeHash string `json:"envelope_hash"`

	// CiphertextSize is the size of encrypted payload in bytes
	CiphertextSize int `json:"ciphertext_size"`

	// NonceSize is the size of the nonce
	NonceSize int `json:"nonce_size"`

	// HasSignature indicates if a sender signature is present
	HasSignature bool `json:"has_signature"`

	// SenderPubKeyFingerprint is the fingerprint of sender's public key
	SenderPubKeyFingerprint string `json:"sender_pub_key_fingerprint,omitempty"`

	// Metadata contains public envelope metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ExtractEnvelopeMetadata extracts safe metadata from an envelope without decryption.
// This is used for introspection queries and CLI commands.
func ExtractEnvelopeMetadata(envelope *EncryptedPayloadEnvelope) (*EnvelopeMetadata, error) {
	if envelope == nil {
		return nil, ErrInvalidEnvelope.Wrap("envelope cannot be nil")
	}

	metadata := &EnvelopeMetadata{
		Version:          envelope.Version,
		AlgorithmID:      envelope.AlgorithmID,
		AlgorithmVersion: envelope.AlgorithmVersion,
		RecipientKeyIDs:  envelope.RecipientKeyIDs,
		EnvelopeHash:     hex.EncodeToString(envelope.Hash()),
		CiphertextSize:   len(envelope.Ciphertext),
		NonceSize:        len(envelope.Nonce),
		HasSignature:     len(envelope.SenderSignature) > 0,
		Metadata:         envelope.Metadata,
	}

	if len(envelope.SenderPubKey) > 0 {
		metadata.SenderPubKeyFingerprint = ComputeKeyFingerprint(envelope.SenderPubKey)
	}

	return metadata, nil
}

// String returns a human-readable representation of the metadata
func (m *EnvelopeMetadata) String() string {
	return fmt.Sprintf("Envelope{version=%d, algorithm=%s(v%d), recipients=%d, size=%d bytes, hash=%s}",
		m.Version, m.AlgorithmID, m.AlgorithmVersion, len(m.RecipientKeyIDs), m.CiphertextSize, m.EnvelopeHash[:16]+"...")
}

// EnvelopeValidationResult contains the result of envelope validation
type EnvelopeValidationResult struct {
	// Valid indicates if the envelope is structurally valid
	Valid bool `json:"valid"`

	// Errors contains validation error messages
	Errors []string `json:"errors,omitempty"`

	// Warnings contains validation warnings (non-fatal)
	Warnings []string `json:"warnings,omitempty"`

	// RecipientStatus maps recipient key IDs to their validation status
	RecipientStatus map[string]RecipientStatus `json:"recipient_status,omitempty"`
}

// RecipientStatus represents the validation status of a recipient key
type RecipientStatus struct {
	// KeyFound indicates if the key exists in the registry
	KeyFound bool `json:"key_found"`

	// KeyActive indicates if the key is active (not revoked)
	KeyActive bool `json:"key_active"`

	// KeyExpired indicates if the key is expired
	KeyExpired bool `json:"key_expired"`

	// KeyDeprecated indicates if the key is deprecated
	KeyDeprecated bool `json:"key_deprecated"`

	// Address is the owner address of the key (if found)
	Address string `json:"address,omitempty"`

	// Message contains a status message
	Message string `json:"message,omitempty"`
}

// AddError adds an error to the validation result
func (r *EnvelopeValidationResult) AddError(msg string) {
	r.Valid = false
	r.Errors = append(r.Errors, msg)
}

// AddWarning adds a warning to the validation result
func (r *EnvelopeValidationResult) AddWarning(msg string) {
	r.Warnings = append(r.Warnings, msg)
}

// SetRecipientStatus sets the status for a recipient
func (r *EnvelopeValidationResult) SetRecipientStatus(keyID string, status RecipientStatus) {
	if r.RecipientStatus == nil {
		r.RecipientStatus = make(map[string]RecipientStatus)
	}
	r.RecipientStatus[keyID] = status
}
