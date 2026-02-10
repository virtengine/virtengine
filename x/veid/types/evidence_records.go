package types

import (
	"encoding/hex"
	"time"
)

// ============================================================================
// Evidence Records (Document + Biometric)
// ============================================================================

// EvidenceType represents the class of evidence being recorded.
type EvidenceType string

const (
	EvidenceTypeDocument  EvidenceType = "document"
	EvidenceTypeBiometric EvidenceType = "biometric"
)

// AllEvidenceTypes returns all supported evidence types.
func AllEvidenceTypes() []EvidenceType {
	return []EvidenceType{
		EvidenceTypeDocument,
		EvidenceTypeBiometric,
	}
}

// IsValidEvidenceType checks if an evidence type is valid.
func IsValidEvidenceType(t EvidenceType) bool {
	for _, valid := range AllEvidenceTypes() {
		if t == valid {
			return true
		}
	}
	return false
}

// EvidenceStatus represents the verification status of evidence.
type EvidenceStatus string

const (
	EvidenceStatusPending    EvidenceStatus = "pending"
	EvidenceStatusVerified   EvidenceStatus = "verified"
	EvidenceStatusRejected   EvidenceStatus = "rejected"
	EvidenceStatusOverridden EvidenceStatus = "overridden"
)

// EvidenceOverride captures reviewer overrides for evidence decisions.
type EvidenceOverride struct {
	ReviewerAddress string         `json:"reviewer_address"`
	Reason          string         `json:"reason"`
	PreviousStatus  EvidenceStatus `json:"previous_status"`
	OverriddenAt    time.Time      `json:"overridden_at"`
}

// EvidenceRecord represents a verification evidence record stored on-chain.
// SECURITY: Only hashes and envelope metadata are stored, never plaintext.
type EvidenceRecord struct {
	// EvidenceID is the unique identifier for this record
	EvidenceID string `json:"evidence_id"`

	// EvidenceType identifies document vs biometric evidence
	EvidenceType EvidenceType `json:"evidence_type"`

	// AccountAddress is the account this evidence belongs to
	AccountAddress string `json:"account_address"`

	// ScopeID references the scope that produced this evidence
	ScopeID string `json:"scope_id"`

	// ContentHash is the SHA-256 hash of the decrypted payload
	ContentHash []byte `json:"content_hash"`

	// EnvelopeHash is the hash of the encrypted envelope
	EnvelopeHash []byte `json:"envelope_hash"`

	// RecipientKeyIDs lists authorized recipient key fingerprints
	RecipientKeyIDs []string `json:"recipient_key_ids"`

	// AlgorithmID is the encryption algorithm used
	AlgorithmID string `json:"algorithm_id"`

	// Confidence is the evidence confidence in basis points (0-10000)
	Confidence uint32 `json:"confidence"`

	// ProvenanceHash is a deterministic hash of evidence provenance metadata
	ProvenanceHash []byte `json:"provenance_hash"`

	// Status is the current evidence decision status
	Status EvidenceStatus `json:"status"`

	// DecisionReason contains human-readable reason
	DecisionReason string `json:"decision_reason,omitempty"`

	// VerifiedAt is when evidence was verified/rejected
	VerifiedAt *time.Time `json:"verified_at,omitempty"`

	// VerifierKeyID is the fingerprint of the verifier key used
	VerifierKeyID string `json:"verifier_key_id,omitempty"`

	// Override captures reviewer override details (if any)
	Override *EvidenceOverride `json:"override,omitempty"`
}

// NewEvidenceRecord creates a new evidence record with default status.
func NewEvidenceRecord(
	evidenceID string,
	evidenceType EvidenceType,
	accountAddress string,
	scopeID string,
	contentHash []byte,
	envelopeHash []byte,
	recipientKeyIDs []string,
	algorithmID string,
	confidence uint32,
	provenanceHash []byte,
) *EvidenceRecord {
	return &EvidenceRecord{
		EvidenceID:      evidenceID,
		EvidenceType:    evidenceType,
		AccountAddress:  accountAddress,
		ScopeID:         scopeID,
		ContentHash:     contentHash,
		EnvelopeHash:    envelopeHash,
		RecipientKeyIDs: recipientKeyIDs,
		AlgorithmID:     algorithmID,
		Confidence:      confidence,
		ProvenanceHash:  provenanceHash,
		Status:          EvidenceStatusPending,
	}
}

// Validate validates the evidence record.
func (r *EvidenceRecord) Validate() error {
	if r.EvidenceID == "" {
		return ErrInvalidPayload.Wrap("evidence_id cannot be empty")
	}
	if !IsValidEvidenceType(r.EvidenceType) {
		return ErrInvalidPayload.Wrapf("invalid evidence_type: %s", r.EvidenceType)
	}
	if r.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account_address cannot be empty")
	}
	if r.ScopeID == "" {
		return ErrInvalidPayload.Wrap("scope_id cannot be empty")
	}
	if len(r.ContentHash) != 32 {
		return ErrInvalidPayloadHash.Wrap("content_hash must be 32 bytes")
	}
	if len(r.EnvelopeHash) != 32 {
		return ErrInvalidPayloadHash.Wrap("envelope_hash must be 32 bytes")
	}
	if len(r.RecipientKeyIDs) == 0 {
		return ErrInvalidPayload.Wrap("recipient_key_ids cannot be empty")
	}
	if r.AlgorithmID == "" {
		return ErrInvalidPayload.Wrap("algorithm_id cannot be empty")
	}
	if r.Confidence > uint32(MaxBasisPoints) {
		return ErrInvalidPayload.Wrap("confidence exceeds maximum basis points")
	}
	if len(r.ProvenanceHash) != 0 && len(r.ProvenanceHash) != 32 {
		return ErrInvalidPayloadHash.Wrap("provenance_hash must be 32 bytes")
	}
	if r.Status == "" {
		return ErrInvalidPayload.Wrap("status cannot be empty")
	}
	return nil
}

// SetDecision updates the evidence decision status.
func (r *EvidenceRecord) SetDecision(status EvidenceStatus, reason string, verifiedAt time.Time) {
	r.Status = status
	r.DecisionReason = reason
	r.VerifiedAt = &verifiedAt
}

// SetOverride applies a reviewer override decision.
func (r *EvidenceRecord) SetOverride(reviewer, reason string, overriddenAt time.Time) {
	previous := r.Status
	r.Status = EvidenceStatusOverridden
	r.Override = &EvidenceOverride{
		ReviewerAddress: reviewer,
		Reason:          reason,
		PreviousStatus:  previous,
		OverriddenAt:    overriddenAt,
	}
}

// EnvelopeHashHex returns the envelope hash as hex.
func (r *EvidenceRecord) EnvelopeHashHex() string {
	if len(r.EnvelopeHash) == 0 {
		return ""
	}
	return hex.EncodeToString(r.EnvelopeHash)
}

// ContentHashHex returns the content hash as hex.
func (r *EvidenceRecord) ContentHashHex() string {
	if len(r.ContentHash) == 0 {
		return ""
	}
	return hex.EncodeToString(r.ContentHash)
}

// ProvenanceHashHex returns the provenance hash as hex.
func (r *EvidenceRecord) ProvenanceHashHex() string {
	if len(r.ProvenanceHash) == 0 {
		return ""
	}
	return hex.EncodeToString(r.ProvenanceHash)
}
