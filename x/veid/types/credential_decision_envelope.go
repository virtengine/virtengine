package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const (
	CredentialDecisionEnvelopeVersion      = uint32(1)
	CredentialDecisionEnvelopeDomain       = "VEID_CREDENTIAL_DECISION_ENVELOPE_V1"
	CredentialDecisionEnvelopeSignDomain   = "VEID_CREDENTIAL_DECISION_SIGN_V1"
	CredentialDecisionEnvelopeDigestDomain = "VEID_CREDENTIAL_DECISION_DIGEST_V1"
	CredentialDecisionMaxDigests           = 32
	CredentialDecisionMaxString            = 128
)

// CredentialDecisionStatus is a read-only credential eligibility outcome.
type CredentialDecisionStatus string

const (
	CredentialDecisionStatusEligible      CredentialDecisionStatus = "eligible"
	CredentialDecisionStatusIneligible    CredentialDecisionStatus = "ineligible"
	CredentialDecisionStatusIndeterminate CredentialDecisionStatus = "indeterminate"
)

// CredentialDecisionEnvelopeV1 binds a credential decision only to
// authenticated evidence, verified receipts, governed epochs, and consent.
// Signatures and claims are deliberately carried elsewhere. Task 90D owns
// issuance, custody, status, presentation, and genuine claim witnesses.
type CredentialDecisionEnvelopeV1 struct {
	ChainID                       string                   `json:"chain_id"`
	Subject                       string                   `json:"subject"`
	SubjectKeyEpoch               uint64                   `json:"subject_key_epoch"`
	AcceptedEvidenceDigests       []string                 `json:"accepted_evidence_digests"`
	AcceptedReceiptDigests        []string                 `json:"accepted_receipt_digests"`
	PolicyEpoch                   uint64                   `json:"policy_epoch"`
	ScoreEpoch                    uint64                   `json:"score_epoch"`
	ConsentPurposeReferenceDigest string                   `json:"consent_purpose_reference_digest"`
	ExpiresAtUnix                 int64                    `json:"expires_at_unix"`
	Status                        CredentialDecisionStatus `json:"status"`
}

type credentialDecisionEnvelopeV1Core struct {
	Domain                        string                   `json:"domain"`
	Version                       uint32                   `json:"version"`
	ChainID                       string                   `json:"chain_id"`
	Subject                       string                   `json:"subject"`
	SubjectKeyEpoch               uint64                   `json:"subject_key_epoch"`
	AcceptedEvidenceDigests       []string                 `json:"accepted_evidence_digests"`
	AcceptedReceiptDigests        []string                 `json:"accepted_receipt_digests"`
	PolicyEpoch                   uint64                   `json:"policy_epoch"`
	ScoreEpoch                    uint64                   `json:"score_epoch"`
	ConsentPurposeReferenceDigest string                   `json:"consent_purpose_reference_digest"`
	ExpiresAtUnix                 int64                    `json:"expires_at_unix"`
	Status                        CredentialDecisionStatus `json:"status"`
}

type credentialDecisionEnvelopeV1Sign struct {
	SignDomain string `json:"sign_domain"`
	credentialDecisionEnvelopeV1Core
}

// NewCredentialDecisionEnvelopeV1 constructs a defensive, canonical envelope.
func NewCredentialDecisionEnvelopeV1(envelope CredentialDecisionEnvelopeV1) (CredentialDecisionEnvelopeV1, error) {
	envelope.AcceptedEvidenceDigests = append([]string{}, envelope.AcceptedEvidenceDigests...)
	envelope.AcceptedReceiptDigests = append([]string{}, envelope.AcceptedReceiptDigests...)
	if err := envelope.Validate(); err != nil {
		return CredentialDecisionEnvelopeV1{}, err
	}
	return envelope, nil
}

// Validate rejects incomplete or non-canonical decision inputs.
func (e CredentialDecisionEnvelopeV1) Validate() error {
	for name, value := range map[string]string{
		"chain_id": e.ChainID,
		"subject":  e.Subject,
	} {
		if value == "" || value != strings.TrimSpace(value) || len(value) > CredentialDecisionMaxString || !isPrintableEvidenceEnvelopeASCII(value) {
			return ErrInvalidCredential.Wrapf("credential decision %s must be nonempty canonical ASCII", name)
		}
	}
	if e.SubjectKeyEpoch == 0 || e.PolicyEpoch == 0 || e.ScoreEpoch == 0 {
		return ErrInvalidCredential.Wrap("credential decision epochs must be positive")
	}
	if err := validateCredentialDecisionDigests(e.AcceptedEvidenceDigests, "accepted_evidence_digests", true); err != nil {
		return err
	}
	if e.AcceptedReceiptDigests == nil {
		return ErrInvalidCredential.Wrap("accepted_receipt_digests must use a canonical array")
	}
	if err := validateCredentialDecisionDigests(e.AcceptedReceiptDigests, "accepted_receipt_digests", e.Status == CredentialDecisionStatusEligible); err != nil {
		return err
	}
	if err := validateEvidenceEnvelopeHex(e.ConsentPurposeReferenceDigest, "consent_purpose_reference_digest"); err != nil {
		return ErrInvalidCredential.Wrap(err.Error())
	}
	if e.ExpiresAtUnix <= 0 {
		return ErrInvalidCredential.Wrap("credential decision expiry must be positive")
	}
	switch e.Status {
	case CredentialDecisionStatusEligible, CredentialDecisionStatusIneligible, CredentialDecisionStatusIndeterminate:
	default:
		return ErrInvalidCredential.Wrapf("invalid credential decision status %q", e.Status)
	}
	return nil
}

// ValidateAt applies deterministic consensus-time expiry validation.
func (e CredentialDecisionEnvelopeV1) ValidateAt(blockTimeUnix int64) error {
	if err := e.Validate(); err != nil {
		return err
	}
	if blockTimeUnix <= 0 {
		return ErrInvalidCredential.Wrap("credential decision block time must be positive")
	}
	if blockTimeUnix >= e.ExpiresAtUnix {
		return ErrCredentialExpired.Wrap("credential decision has expired")
	}
	return nil
}

// ValidateForChainAt applies exact chain binding and deterministic expiry.
func (e CredentialDecisionEnvelopeV1) ValidateForChainAt(expectedChainID string, blockTimeUnix int64) error {
	if err := e.ValidateAt(blockTimeUnix); err != nil {
		return err
	}
	if expectedChainID == "" || e.ChainID != expectedChainID {
		return ErrInvalidCredential.Wrap("credential decision chain_id mismatch")
	}
	return nil
}

func validateCredentialDecisionDigests(values []string, name string, required bool) error {
	if required && len(values) == 0 {
		return ErrInvalidCredential.Wrapf("%s is required", name)
	}
	if len(values) > CredentialDecisionMaxDigests {
		return ErrInvalidCredential.Wrapf("%s exceeds %d entries", name, CredentialDecisionMaxDigests)
	}
	if !sort.StringsAreSorted(values) {
		return ErrInvalidCredential.Wrapf("%s must be sorted", name)
	}
	for index, value := range values {
		if err := validateEvidenceEnvelopeHex(value, name); err != nil {
			return ErrInvalidCredential.Wrap(err.Error())
		}
		if index > 0 && values[index-1] == value {
			return ErrInvalidCredential.Wrapf("%s must be unique", name)
		}
	}
	return nil
}

func (e CredentialDecisionEnvelopeV1) core() credentialDecisionEnvelopeV1Core {
	return credentialDecisionEnvelopeV1Core{
		Domain: CredentialDecisionEnvelopeDomain, Version: CredentialDecisionEnvelopeVersion,
		ChainID: e.ChainID, Subject: e.Subject, SubjectKeyEpoch: e.SubjectKeyEpoch,
		AcceptedEvidenceDigests: append([]string{}, e.AcceptedEvidenceDigests...),
		AcceptedReceiptDigests:  append([]string{}, e.AcceptedReceiptDigests...),
		PolicyEpoch:             e.PolicyEpoch, ScoreEpoch: e.ScoreEpoch,
		ConsentPurposeReferenceDigest: e.ConsentPurposeReferenceDigest,
		ExpiresAtUnix:                 e.ExpiresAtUnix, Status: e.Status,
	}
}

// SignBytes returns deterministic bytes for an external governed signer.
func (e CredentialDecisionEnvelopeV1) SignBytes() ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(credentialDecisionEnvelopeV1Sign{
		SignDomain:                       CredentialDecisionEnvelopeSignDomain,
		credentialDecisionEnvelopeV1Core: e.core(),
	})
}

// Digest returns a fresh domain-separated SHA-256 decision digest.
func (e CredentialDecisionEnvelopeV1) Digest() ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	bz, err := json.Marshal(e.core())
	if err != nil {
		return nil, fmt.Errorf("marshal credential decision envelope: %w", err)
	}
	hasher := sha256.New()
	_, _ = hasher.Write([]byte(CredentialDecisionEnvelopeDigestDomain))
	_, _ = hasher.Write(bz)
	return hasher.Sum(nil), nil
}

// DigestHex returns the lowercase hexadecimal decision digest.
func (e CredentialDecisionEnvelopeV1) DigestHex() (string, error) {
	digest, err := e.Digest()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(digest), nil
}
