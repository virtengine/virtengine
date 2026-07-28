package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	WebEvidenceVersion       = "1"
	WebEvidenceIssuerDomain  = "VEID_WEB_EVIDENCE_ISSUER_V1"
	WebEvidenceAccountDomain = "VEID_WEB_EVIDENCE_ACCOUNT_AUTH_V1"

	WebEvidenceActionSubmitSSO    = "submit_sso_verification_proof"
	WebEvidenceActionSubmitEmail  = "submit_email_verification_proof"
	WebEvidenceActionSubmitSMS    = "submit_sms_verification_proof"
	WebEvidenceActionSubmitSocial = "submit_social_media_scope"

	WebEvidenceMetadataPrefix              = "veid.web."
	WebEvidenceMetadataFieldPrefix         = WebEvidenceMetadataPrefix + "field."
	WebEvidenceMetadataVersion             = WebEvidenceMetadataPrefix + "version"
	WebEvidenceMetadataChainID             = WebEvidenceMetadataPrefix + "chain_id"
	WebEvidenceMetadataAccountAddress      = WebEvidenceMetadataPrefix + "account_address"
	WebEvidenceMetadataEvidenceType        = WebEvidenceMetadataPrefix + "evidence_type"
	WebEvidenceMetadataAction              = WebEvidenceMetadataPrefix + "action"
	WebEvidenceMetadataScopeID             = WebEvidenceMetadataPrefix + "scope_id"
	WebEvidenceMetadataAttestationDigest   = WebEvidenceMetadataPrefix + "attestation_digest"
	WebEvidenceMetadataIssuerID            = WebEvidenceMetadataPrefix + "issuer_id"
	WebEvidenceMetadataIssuerKeyID         = WebEvidenceMetadataPrefix + "issuer_key_id"
	WebEvidenceMetadataIssuerFingerprint   = WebEvidenceMetadataPrefix + "issuer_fingerprint"
	WebEvidenceMetadataIssuerAlgorithm     = WebEvidenceMetadataPrefix + "issuer_algorithm"
	WebEvidenceMetadataNonce               = WebEvidenceMetadataPrefix + "nonce"
	WebEvidenceMetadataChallenge           = WebEvidenceMetadataPrefix + "challenge"
	WebEvidenceMetadataIssuedAtUnix        = WebEvidenceMetadataPrefix + "issued_at_unix"
	WebEvidenceMetadataExpiresAtUnix       = WebEvidenceMetadataPrefix + "expires_at_unix"
	WebEvidenceMetadataServiceMetadataHash = WebEvidenceMetadataPrefix + "service_metadata_hash"

	SignerKeyMetadataActivationHeight    = "activation_height"
	SignerKeyMetadataExpiryHeight        = "expiry_height"
	SignerKeyMetadataRevokedHeight       = "revoked_height"
	SignerKeyMetadataEvidenceTypes       = "evidence_types"
	SignerKeyMetadataServiceMetadataHash = "service_metadata_hash"
)

// WebEvidenceContextConfig contains the inputs bound into issuer and account
// signatures for authenticated web-scope evidence.
type WebEvidenceContextConfig struct {
	ChainID             string
	AccountAddress      string
	EvidenceType        AttestationType
	Action              string
	ScopeID             string
	AttestationDigest   string
	Issuer              AttestationIssuer
	IssuerAlgorithm     AttestationProofType
	Nonce               string
	Challenge           string
	IssuedAt            time.Time
	ExpiresAt           time.Time
	ServiceMetadataHash string
	CallerFields        map[string]string
}

// WebEvidenceField is one caller-controlled field covered by the canonical
// evidence envelope.
type WebEvidenceField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// WebEvidenceContext is the normalized canonical envelope shared by issuer
// signatures, account authorizations, and strict attestation metadata.
type WebEvidenceContext struct {
	ChainID             string
	AccountAddress      string
	EvidenceType        AttestationType
	Action              string
	ScopeID             string
	AttestationDigest   string
	IssuerID            string
	IssuerKeyID         string
	IssuerFingerprint   string
	IssuerAlgorithm     AttestationProofType
	Nonce               string
	Challenge           string
	IssuedAt            time.Time
	ExpiresAt           time.Time
	ServiceMetadataHash string
	CallerFields        []WebEvidenceField
}

// NewWebEvidenceContext normalizes caller fields for deterministic signing.
func NewWebEvidenceContext(cfg WebEvidenceContextConfig) WebEvidenceContext {
	fields := make([]WebEvidenceField, 0, len(cfg.CallerFields))
	for k, v := range cfg.CallerFields {
		fields = append(fields, WebEvidenceField{Name: k, Value: v})
	}
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})

	return WebEvidenceContext{
		ChainID:             cfg.ChainID,
		AccountAddress:      cfg.AccountAddress,
		EvidenceType:        cfg.EvidenceType,
		Action:              cfg.Action,
		ScopeID:             cfg.ScopeID,
		AttestationDigest:   strings.ToLower(cfg.AttestationDigest),
		IssuerID:            cfg.Issuer.ID,
		IssuerKeyID:         cfg.Issuer.KeyID,
		IssuerFingerprint:   strings.ToLower(cfg.Issuer.KeyFingerprint),
		IssuerAlgorithm:     cfg.IssuerAlgorithm,
		Nonce:               cfg.Nonce,
		Challenge:           cfg.Challenge,
		IssuedAt:            cfg.IssuedAt.UTC(),
		ExpiresAt:           cfg.ExpiresAt.UTC(),
		ServiceMetadataHash: strings.ToLower(cfg.ServiceMetadataHash),
		CallerFields:        fields,
	}
}

func (c WebEvidenceContext) validate() error {
	required := map[string]string{
		"chain_id":           c.ChainID,
		"account_address":    c.AccountAddress,
		"evidence_type":      string(c.EvidenceType),
		"action":             c.Action,
		"scope_id":           c.ScopeID,
		"attestation_digest": c.AttestationDigest,
		"issuer_id":          c.IssuerID,
		"issuer_key_id":      c.IssuerKeyID,
		"issuer_fingerprint": c.IssuerFingerprint,
		"issuer_algorithm":   string(c.IssuerAlgorithm),
		"nonce":              c.Nonce,
		"challenge":          c.Challenge,
		"issued_at":          c.IssuedAt.Format(time.RFC3339Nano),
		"expires_at":         c.ExpiresAt.Format(time.RFC3339Nano),
	}
	for name, value := range required {
		if value == "" {
			return ErrInvalidAttestation.Wrapf("web evidence %s is required", name)
		}
	}
	if !IsValidAttestationType(c.EvidenceType) {
		return ErrInvalidAttestation.Wrapf("invalid web evidence type: %s", c.EvidenceType)
	}
	if !IsValidProofType(c.IssuerAlgorithm) {
		return ErrInvalidSignerKey.Wrapf("invalid web evidence algorithm: %s", c.IssuerAlgorithm)
	}
	if err := validateHexDigest(c.AttestationDigest, "attestation_digest"); err != nil {
		return err
	}
	if err := validateHexDigest(c.IssuerFingerprint, "issuer_fingerprint"); err != nil {
		return err
	}
	if c.ServiceMetadataHash != "" {
		if err := validateHexDigest(c.ServiceMetadataHash, "service_metadata_hash"); err != nil {
			return err
		}
	}
	if !c.ExpiresAt.After(c.IssuedAt) {
		return ErrInvalidTimestamp.Wrap("web evidence expires_at must be after issued_at")
	}
	for _, field := range c.CallerFields {
		if field.Name == "" {
			return ErrInvalidAttestation.Wrap("web evidence caller field name is required")
		}
		if strings.HasPrefix(field.Name, WebEvidenceMetadataPrefix) {
			return ErrInvalidAttestation.Wrap("web evidence caller field name uses reserved prefix")
		}
	}
	return nil
}

func validateHexDigest(value, name string) error {
	if len(value) != 64 {
		return ErrInvalidAttestation.Wrapf("%s must be a 64-character SHA256 hex digest", name)
	}
	if _, err := hex.DecodeString(value); err != nil {
		return ErrInvalidAttestation.Wrapf("%s must be valid hex", name)
	}
	return nil
}

func canonicalWebEvidenceFields(metadata map[string]string, allowReserved bool) ([]WebEvidenceField, error) {
	if len(metadata) == 0 {
		return nil, nil
	}
	fields := make([]WebEvidenceField, 0, len(metadata))
	for key, value := range metadata {
		if key == "" {
			return nil, ErrInvalidAttestation.Wrap("web evidence metadata key is required")
		}
		if !allowReserved && strings.HasPrefix(key, WebEvidenceMetadataPrefix) {
			return nil, ErrInvalidAttestation.Wrap("web evidence service metadata uses reserved prefix")
		}
		fields = append(fields, WebEvidenceField{Name: key, Value: value})
	}
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})
	return fields, nil
}

// WebEvidenceServiceMetadataHash hashes caller-supplied service metadata as an
// explicit sorted key/value list. It intentionally avoids relying on JSON map
// key ordering for a security-sensitive signature input.
func WebEvidenceServiceMetadataHash(metadata map[string]string) (string, error) {
	fields, err := canonicalWebEvidenceFields(metadata, false)
	if err != nil {
		return "", err
	}
	if len(fields) == 0 {
		return "", nil
	}
	env := struct {
		Domain  string             `json:"domain"`
		Version string             `json:"version"`
		Fields  []WebEvidenceField `json:"fields"`
	}{
		Domain:  "VEID_WEB_EVIDENCE_SERVICE_METADATA_V1",
		Version: WebEvidenceVersion,
		Fields:  fields,
	}
	bz, err := json.Marshal(env)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(bz)
	return hex.EncodeToString(sum[:]), nil
}

type webEvidenceSignEnvelope struct {
	Domain              string               `json:"domain"`
	Version             string               `json:"version"`
	ChainID             string               `json:"chain_id"`
	AccountAddress      string               `json:"account_address"`
	EvidenceType        AttestationType      `json:"evidence_type"`
	Action              string               `json:"action"`
	ScopeID             string               `json:"scope_id"`
	AttestationDigest   string               `json:"attestation_digest"`
	IssuerID            string               `json:"issuer_id"`
	IssuerKeyID         string               `json:"issuer_key_id"`
	IssuerFingerprint   string               `json:"issuer_fingerprint"`
	IssuerAlgorithm     AttestationProofType `json:"issuer_algorithm"`
	Nonce               string               `json:"nonce"`
	Challenge           string               `json:"challenge"`
	IssuedAtUnix        int64                `json:"issued_at_unix"`
	ExpiresAtUnix       int64                `json:"expires_at_unix"`
	ServiceMetadataHash string               `json:"service_metadata_hash,omitempty"`
	CallerFields        []WebEvidenceField   `json:"caller_fields"`
}

func (c WebEvidenceContext) signBytes(domain string) ([]byte, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	env := webEvidenceSignEnvelope{
		Domain:              domain,
		Version:             WebEvidenceVersion,
		ChainID:             c.ChainID,
		AccountAddress:      c.AccountAddress,
		EvidenceType:        c.EvidenceType,
		Action:              c.Action,
		ScopeID:             c.ScopeID,
		AttestationDigest:   c.AttestationDigest,
		IssuerID:            c.IssuerID,
		IssuerKeyID:         c.IssuerKeyID,
		IssuerFingerprint:   c.IssuerFingerprint,
		IssuerAlgorithm:     c.IssuerAlgorithm,
		Nonce:               c.Nonce,
		Challenge:           c.Challenge,
		IssuedAtUnix:        c.IssuedAt.Unix(),
		ExpiresAtUnix:       c.ExpiresAt.Unix(),
		ServiceMetadataHash: c.ServiceMetadataHash,
		CallerFields:        append([]WebEvidenceField(nil), c.CallerFields...),
	}
	return json.Marshal(env)
}

// IssuerSignBytes returns the domain-separated deterministic bytes signed by
// the governed issuer key.
func (c WebEvidenceContext) IssuerSignBytes() ([]byte, error) {
	return c.signBytes(WebEvidenceIssuerDomain)
}

// AccountAuthorizationBytes returns the distinct domain-separated bytes signed
// by the account wallet binding key.
func (c WebEvidenceContext) AccountAuthorizationBytes() ([]byte, error) {
	return c.signBytes(WebEvidenceAccountDomain)
}

// Metadata returns the strict metadata schema required on web attestations.
func (c WebEvidenceContext) Metadata() (map[string]string, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	metadata := map[string]string{
		WebEvidenceMetadataVersion:           WebEvidenceVersion,
		WebEvidenceMetadataChainID:           c.ChainID,
		WebEvidenceMetadataAccountAddress:    c.AccountAddress,
		WebEvidenceMetadataEvidenceType:      string(c.EvidenceType),
		WebEvidenceMetadataAction:            c.Action,
		WebEvidenceMetadataScopeID:           c.ScopeID,
		WebEvidenceMetadataAttestationDigest: c.AttestationDigest,
		WebEvidenceMetadataIssuerID:          c.IssuerID,
		WebEvidenceMetadataIssuerKeyID:       c.IssuerKeyID,
		WebEvidenceMetadataIssuerFingerprint: c.IssuerFingerprint,
		WebEvidenceMetadataIssuerAlgorithm:   string(c.IssuerAlgorithm),
		WebEvidenceMetadataNonce:             c.Nonce,
		WebEvidenceMetadataChallenge:         c.Challenge,
		WebEvidenceMetadataIssuedAtUnix:      fmt.Sprintf("%d", c.IssuedAt.Unix()),
		WebEvidenceMetadataExpiresAtUnix:     fmt.Sprintf("%d", c.ExpiresAt.Unix()),
	}
	if c.ServiceMetadataHash != "" {
		metadata[WebEvidenceMetadataServiceMetadataHash] = c.ServiceMetadataHash
	}
	for _, field := range c.CallerFields {
		metadata[WebEvidenceMetadataFieldPrefix+field.Name] = field.Value
	}
	return metadata, nil
}

// ApplyToAttestation replaces attestation metadata with the strict signed
// schema. It is an off-chain/test construction helper only; handlers must
// validate caller-supplied metadata with ValidateAttestationMetadata and must
// never normalize or replace metadata on behalf of a submitter.
func (c WebEvidenceContext) ApplyToAttestation(att *VerificationAttestation) error {
	if att == nil {
		return ErrInvalidAttestation.Wrap("attestation cannot be nil")
	}
	metadata, err := c.Metadata()
	if err != nil {
		return err
	}
	att.Metadata = metadata
	return nil
}

// ValidateAttestationMetadata verifies attestation metadata exactly matches the
// signed web-evidence schema, rejecting unsigned caller-supplied extras.
func (c WebEvidenceContext) ValidateAttestationMetadata(att *VerificationAttestation) error {
	if att == nil {
		return ErrInvalidAttestation.Wrap("attestation cannot be nil")
	}
	expected, err := c.Metadata()
	if err != nil {
		return err
	}
	if att.Metadata == nil {
		return ErrInvalidAttestation.Wrap("web evidence metadata is required")
	}
	if len(att.Metadata) != len(expected) {
		return ErrInvalidAttestation.Wrap("web evidence metadata contains unexpected keys")
	}
	for key, value := range expected {
		if got, ok := att.Metadata[key]; !ok || got != value {
			return ErrInvalidAttestation.Wrapf("web evidence metadata mismatch for %s", key)
		}
	}
	for key := range att.Metadata {
		if _, ok := expected[key]; !ok {
			return ErrInvalidAttestation.Wrapf("unknown web evidence metadata key: %s", key)
		}
	}
	return nil
}

// WebEvidenceAttestationDigestHex returns a digest of the attestation core. It
// intentionally excludes Metadata and Proof so metadata can carry the digest
// without creating a circular signature dependency.
func WebEvidenceAttestationDigestHex(att *VerificationAttestation) (string, error) {
	if att == nil {
		return "", ErrInvalidAttestation.Wrap("attestation cannot be nil")
	}
	clone := *att
	clone.Metadata = nil
	clone.Proof = AttestationProof{}
	bz, err := clone.CanonicalBytes()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(bz)
	return hex.EncodeToString(sum[:]), nil
}

// WebEvidenceDigestHex hashes already-canonical attestation bytes, such as the
// SSO-specific canonical form.
func WebEvidenceDigestHex(canonical []byte) string {
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:])
}
