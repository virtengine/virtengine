// Package marketplace provides types for the marketplace on-chain module.
//
// MARKET-HW-SAFEGUARD-1: VEID is recorded as one advisory fraud signal, and
// listings carry hashed capacity/ownership attestations. Neither is a guarantee.
//
// Constitution 6.1.3/6.2.3 (privacy-preserving anti-fraud, equal access) and
// 39.1-39.2 (data minimisation): attestation evidence is stored as a hash of the
// supplied document. Raw documents never enter consensus state.
package marketplace

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// VEIDFraudSignal is the VEID outcome recorded on an order or dispute as an
// *advisory* fraud signal. It is never a pass/fail guarantee and settlement
// logic must not branch on it alone.
type VEIDFraudSignal struct {
	// Score is the counterparty's VEID identity score at evaluation time (0-100).
	Score uint32 `json:"score"`

	// Tier is the identity tier at evaluation time.
	Tier int `json:"tier"`

	// Verified reports whether identity verification was complete.
	Verified bool `json:"verified"`

	// RequirementSatisfied reports whether the order's stated identity
	// requirement was met.
	RequirementSatisfied bool `json:"requirement_satisfied"`

	// Advisory is always true: this record informs a risk decision, it does not
	// decide one.
	Advisory bool `json:"advisory"`

	// NotDeterminative is always true: this signal alone must never authorise
	// or block a settlement.
	NotDeterminative bool `json:"not_determinative"`

	// Source names the evaluator, so the record is auditable.
	Source string `json:"source"`

	// EvaluatedAt is the block time of evaluation.
	EvaluatedAt time.Time `json:"evaluated_at"`
}

// NewVEIDFraudSignal builds an advisory VEID signal. Advisory and
// NotDeterminative are set unconditionally by the constructor so no caller can
// mint a signal that claims to be determinative.
func NewVEIDFraudSignal(score uint32, tier int, verified, requirementSatisfied bool, now time.Time) *VEIDFraudSignal {
	return &VEIDFraudSignal{
		Score:                score,
		Tier:                 tier,
		Verified:             verified,
		RequirementSatisfied: requirementSatisfied,
		Advisory:             true,
		NotDeterminative:     true,
		Source:               "veid",
		EvaluatedAt:          now.UTC(),
	}
}

// Validate validates the signal.
func (s *VEIDFraudSignal) Validate() error {
	if s.Score > 100 {
		return fmt.Errorf("veid signal score %d exceeds 100", s.Score)
	}
	if !s.Advisory {
		return fmt.Errorf("veid signal must be advisory")
	}
	if !s.NotDeterminative {
		return fmt.Errorf("veid signal must be non-determinative")
	}
	if strings.TrimSpace(s.Source) == "" {
		return fmt.Errorf("veid signal source is required")
	}
	return nil
}

// NotDeterminativeText is the "what this does not prove" framing shown alongside
// the signal in any UI or dispute record.
func (s *VEIDFraudSignal) NotDeterminativeText() string {
	return "VEID confirms an identity signal only. It does not prove that hardware or " +
		"capacity exists, is owned by the seller, matches the advertised specification, " +
		"or will be delivered. Treat it as one advisory input, never as a guarantee."
}

// AttestationKind identifies the class of evidence a listing carries.
type AttestationKind string

const (
	// AttestationKindCapacity is a provider-supplied capacity report.
	AttestationKindCapacity AttestationKind = "capacity"

	// AttestationKindOwnership is a hashed ownership/entitlement proof.
	AttestationKindOwnership AttestationKind = "ownership"
)

// IsValidAttestationKind reports whether the kind is known.
func IsValidAttestationKind(kind AttestationKind) bool {
	switch kind {
	case AttestationKindCapacity, AttestationKindOwnership:
		return true
	default:
		return false
	}
}

// OfferingAttestation is a capacity or ownership attestation attached to a
// listing. Only the hash of the source document is stored (Constitution 39.1-39.2).
type OfferingAttestation struct {
	// Kind is the attestation class.
	Kind AttestationKind `json:"kind"`

	// EvidenceHash is the lowercase hex sha256 of the off-chain document. The
	// document itself is never stored on chain.
	EvidenceHash string `json:"evidence_hash"`

	// Issuer is the party that issued the underlying document.
	Issuer string `json:"issuer"`

	// ProviderSupplied reports that the evidence came from the seller. This is
	// surfaced verbatim so a counterparty can weigh it correctly.
	ProviderSupplied bool `json:"provider_supplied"`

	// IssuedAt is when the underlying document was issued.
	IssuedAt time.Time `json:"issued_at"`

	// ExpiresAt optionally bounds the validity of the evidence.
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// NewOfferingAttestation hashes the supplied document and discards its bytes.
func NewOfferingAttestation(kind AttestationKind, issuer string, document []byte, providerSupplied bool, issuedAt time.Time) *OfferingAttestation {
	digest := sha256.Sum256(document)
	return &OfferingAttestation{
		Kind:             kind,
		EvidenceHash:     hex.EncodeToString(digest[:]),
		Issuer:           issuer,
		ProviderSupplied: providerSupplied,
		IssuedAt:         issuedAt.UTC(),
	}
}

// Validate validates the attestation.
func (a *OfferingAttestation) Validate() error {
	if !IsValidAttestationKind(a.Kind) {
		return fmt.Errorf("invalid attestation kind %q", a.Kind)
	}
	if len(a.EvidenceHash) != sha256.Size*2 {
		return fmt.Errorf("attestation evidence_hash must be a %d-character hex sha256 digest", sha256.Size*2)
	}
	if _, err := hex.DecodeString(a.EvidenceHash); err != nil {
		return fmt.Errorf("attestation evidence_hash is not valid hex: %w", err)
	}
	if strings.TrimSpace(a.Issuer) == "" {
		return fmt.Errorf("attestation issuer is required")
	}
	if a.ExpiresAt != nil && a.ExpiresAt.Before(a.IssuedAt) {
		return fmt.Errorf("attestation expires_at precedes issued_at")
	}
	return nil
}

// AttestationRequirementText returns the evidence the seller owes for a kind.
func AttestationRequirementText(kind AttestationKind) string {
	switch kind {
	case AttestationKindCapacity:
		return "Seller must attach a capacity report for the advertised hardware or compute."
	case AttestationKindOwnership:
		return "Seller must attach evidence that they are entitled to offer this hardware or capacity."
	default:
		return "Unknown attestation kind."
	}
}

// AttestationDoesNotProveText returns the "what this does not prove" framing for
// a kind, matched to the identity requirement display.
func AttestationDoesNotProveText(kind AttestationKind) string {
	shared := "Only a hash of the document is recorded on chain, and neither the hash nor its " +
		"presence proves the document is accurate, current, or independently audited."
	switch kind {
	case AttestationKindCapacity:
		return "Attestation evidence does not prove the hardware exists, meets the advertised " +
			"specification, or is available for the term of the order; it shows only that a " +
			"capacity document was supplied. " + shared
	case AttestationKindOwnership:
		return "Attestation evidence does not prove the seller owns or may transfer the hardware; " +
			"it shows only that an ownership claim was made. " + shared
	default:
		return "Attestation evidence does not prove the underlying claim. " + shared
	}
}

// OfferingListingView is the listing projection the UI renders: whether evidence
// is present, the requirement text, and the staged-payment terms. It is a query
// result shape and deliberately carries no raw evidence.
type OfferingListingView struct {
	// OfferingID is the offering the view describes.
	OfferingID string `json:"offering_id"`

	// AttestationPresent reports whether an attestation is attached.
	AttestationPresent bool `json:"attestation_present"`

	// AttestationKind is the attached attestation class, when present.
	AttestationKind AttestationKind `json:"attestation_kind,omitempty"`

	// AttestationEvidenceHash is the hashed proof, when present.
	AttestationEvidenceHash string `json:"attestation_evidence_hash,omitempty"`

	// AttestationProviderSupplied reports the evidence origin, when present.
	AttestationProviderSupplied bool `json:"attestation_provider_supplied,omitempty"`

	// AttestationRequirement is the requirement text for the expected evidence.
	AttestationRequirement string `json:"attestation_requirement"`

	// AttestationDoesNotProve is the "what this does not prove" framing.
	AttestationDoesNotProve string `json:"attestation_does_not_prove"`

	// Milestones is the staged-payment schedule that would apply to an order.
	Milestones MilestoneSet `json:"milestones"`

	// IdentityRequirementText describes the identity requirement in advisory terms.
	IdentityRequirementText string `json:"identity_requirement_text"`
}
