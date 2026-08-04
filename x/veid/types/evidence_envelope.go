package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

const (
	// EvidenceEnvelopeDomain identifies the frozen v1 evidence audit and transport contract.
	EvidenceEnvelopeDomain = "VEID_EVIDENCE_ENVELOPE_V1"
	// EvidenceEnvelopeVersion is the frozen evidence envelope version. Runtime activation and store migration remain later T1/T4 work.
	EvidenceEnvelopeVersion = "1"

	EvidenceEnvelopeIssuerSignDomain    = "VEID_EVIDENCE_ENVELOPE_ISSUER_SIGN_V1"
	EvidenceEnvelopeAccountSignDomain   = "VEID_EVIDENCE_ENVELOPE_ACCOUNT_SIGN_V1"
	EvidenceEnvelopeContextDigestDomain = "VEID_EVIDENCE_ENVELOPE_CONTEXT_DIGEST_V1"
	EvidenceEnvelopeGlobalNonceDomain   = "VEID_EVIDENCE_ENVELOPE_GLOBAL_NONCE_V1"
	EvidenceEnvelopeWebSourceDomain     = "VEID_EVIDENCE_ENVELOPE_WEB_SOURCE_V1"

	EvidenceEnvelopeMaxString = 128
)

// EvidenceEnvelopeV1 is the frozen v1 canonical audit and transport contract.
// Signatures are carried separately to avoid circular signing. Runtime
// activation and store migration remain later T1/T4 work.
type EvidenceEnvelopeV1 struct {
	Domain                       string               `json:"domain"`
	Version                      string               `json:"version"`
	ChainID                      string               `json:"chain_id"`
	AccountAddress               string               `json:"account_address"`
	AccountBindingKeyFingerprint string               `json:"account_binding_key_fingerprint"`
	AccountBindingKeyAlgorithm   AttestationProofType `json:"account_binding_key_algorithm"`
	ScopeID                      string               `json:"scope_id"`
	EvidenceType                 AttestationType      `json:"evidence_type"`
	EvidenceID                   string               `json:"evidence_id"`
	Action                       string               `json:"action"`
	IntendedVerifier             string               `json:"intended_verifier"`
	PayloadDigest                string               `json:"payload_digest"`
	SourceContextDigest          string               `json:"source_context_digest"`
	StorageCommitmentDigest      string               `json:"storage_commitment_digest"`
	IssuerID                     string               `json:"issuer_id"`
	IssuerKeyID                  string               `json:"issuer_key_id"`
	IssuerKeySequence            uint64               `json:"issuer_key_sequence"`
	IssuerKeyFingerprint         string               `json:"issuer_key_fingerprint"`
	IssuerKeyAlgorithm           AttestationProofType `json:"issuer_key_algorithm"`
	PolicyVersion                string               `json:"policy_version"`
	SchemaVersion                string               `json:"schema_version"`
	ModelVersion                 string               `json:"model_version"`
	Nonce                        string               `json:"nonce"`
	Challenge                    string               `json:"challenge"`
	IssuedAtUnix                 int64                `json:"issued_at_unix"`
	ExpiresAtUnix                int64                `json:"expires_at_unix"`
	IssuedHeight                 int64                `json:"issued_height"`
	ExpiresHeight                int64                `json:"expires_height"`
}

// EvidenceEnvelopeV1Config supplies fields not present in WebEvidenceContext.
type EvidenceEnvelopeV1Config struct {
	AccountBindingKeyFingerprint string
	AccountBindingKeyAlgorithm   AttestationProofType
	IssuerKeySequence            uint64
	EvidenceID                   string
	StorageCommitmentDigest      string
	PolicyVersion                string
	SchemaVersion                string
	ModelVersion                 string
	IntendedVerifier             string
	IssuedHeight                 int64
	ExpiresHeight                int64
}

type evidenceEnvelopeV1Core struct {
	Domain                       string               `json:"domain"`
	Version                      string               `json:"version"`
	ChainID                      string               `json:"chain_id"`
	AccountAddress               string               `json:"account_address"`
	AccountBindingKeyFingerprint string               `json:"account_binding_key_fingerprint"`
	AccountBindingKeyAlgorithm   AttestationProofType `json:"account_binding_key_algorithm"`
	ScopeID                      string               `json:"scope_id"`
	EvidenceType                 AttestationType      `json:"evidence_type"`
	EvidenceID                   string               `json:"evidence_id"`
	Action                       string               `json:"action"`
	IntendedVerifier             string               `json:"intended_verifier"`
	PayloadDigest                string               `json:"payload_digest"`
	SourceContextDigest          string               `json:"source_context_digest"`
	StorageCommitmentDigest      string               `json:"storage_commitment_digest"`
	IssuerID                     string               `json:"issuer_id"`
	IssuerKeyID                  string               `json:"issuer_key_id"`
	IssuerKeySequence            uint64               `json:"issuer_key_sequence"`
	IssuerKeyFingerprint         string               `json:"issuer_key_fingerprint"`
	IssuerKeyAlgorithm           AttestationProofType `json:"issuer_key_algorithm"`
	PolicyVersion                string               `json:"policy_version"`
	SchemaVersion                string               `json:"schema_version"`
	ModelVersion                 string               `json:"model_version"`
	Nonce                        string               `json:"nonce"`
	Challenge                    string               `json:"challenge"`
	IssuedAtUnix                 int64                `json:"issued_at_unix"`
	ExpiresAtUnix                int64                `json:"expires_at_unix"`
	IssuedHeight                 int64                `json:"issued_height"`
	ExpiresHeight                int64                `json:"expires_height"`
}

type evidenceEnvelopeV1Sign struct {
	SignDomain string `json:"sign_domain"`
	evidenceEnvelopeV1Core
}

type evidenceEnvelopeV1GlobalNonce struct {
	Domain               string               `json:"domain"`
	EnvelopeVersion      string               `json:"envelope_version"`
	ChainID              string               `json:"chain_id"`
	IssuerID             string               `json:"issuer_id"`
	IssuerKeyID          string               `json:"issuer_key_id"`
	IssuerKeySequence    uint64               `json:"issuer_key_sequence"`
	IssuerKeyFingerprint string               `json:"issuer_key_fingerprint"`
	IssuerKeyAlgorithm   AttestationProofType `json:"issuer_key_algorithm"`
	Nonce                string               `json:"nonce"`
	Challenge            string               `json:"challenge"`
}

// NewEvidenceEnvelopeV1FromWebEvidence extends the existing canonical web
// evidence context without changing WebEvidenceVersion or its sign bytes.
func NewEvidenceEnvelopeV1FromWebEvidence(web WebEvidenceContext, cfg EvidenceEnvelopeV1Config) (EvidenceEnvelopeV1, error) {
	sourceBytes, err := web.IssuerSignBytes()
	if err != nil {
		return EvidenceEnvelopeV1{}, err
	}
	sourceHasher := sha256.New()
	_, _ = sourceHasher.Write([]byte(EvidenceEnvelopeWebSourceDomain))
	_, _ = sourceHasher.Write(sourceBytes)
	envelope := EvidenceEnvelopeV1{
		Domain:                       EvidenceEnvelopeDomain,
		Version:                      EvidenceEnvelopeVersion,
		ChainID:                      web.ChainID,
		AccountAddress:               web.AccountAddress,
		AccountBindingKeyFingerprint: cfg.AccountBindingKeyFingerprint,
		AccountBindingKeyAlgorithm:   cfg.AccountBindingKeyAlgorithm,
		ScopeID:                      web.ScopeID,
		EvidenceType:                 web.EvidenceType,
		EvidenceID:                   cfg.EvidenceID,
		Action:                       web.Action,
		IntendedVerifier:             cfg.IntendedVerifier,
		PayloadDigest:                web.AttestationDigest,
		SourceContextDigest:          hex.EncodeToString(sourceHasher.Sum(nil)),
		StorageCommitmentDigest:      cfg.StorageCommitmentDigest,
		IssuerID:                     web.IssuerID,
		IssuerKeyID:                  web.IssuerKeyID,
		IssuerKeySequence:            cfg.IssuerKeySequence,
		IssuerKeyFingerprint:         web.IssuerFingerprint,
		IssuerKeyAlgorithm:           web.IssuerAlgorithm,
		PolicyVersion:                cfg.PolicyVersion,
		SchemaVersion:                cfg.SchemaVersion,
		ModelVersion:                 cfg.ModelVersion,
		Nonce:                        web.Nonce,
		Challenge:                    web.Challenge,
		IssuedAtUnix:                 web.IssuedAt.Unix(),
		ExpiresAtUnix:                web.ExpiresAt.Unix(),
		IssuedHeight:                 cfg.IssuedHeight,
		ExpiresHeight:                cfg.ExpiresHeight,
	}
	if err := envelope.Validate(); err != nil {
		return EvidenceEnvelopeV1{}, err
	}
	return envelope, nil
}

// Validate rejects non-canonical or incomplete frozen v1 envelopes. It does
// not consult a local wall clock.
func (e EvidenceEnvelopeV1) Validate() error {
	if e.Domain != EvidenceEnvelopeDomain {
		return ErrInvalidAttestation.Wrap("invalid evidence envelope domain")
	}
	if e.Version != EvidenceEnvelopeVersion {
		return ErrInvalidAttestation.Wrap("unsupported evidence envelope version")
	}
	required := []struct {
		name  string
		value string
	}{
		{"chain_id", e.ChainID},
		{"account_address", e.AccountAddress},
		{"scope_id", e.ScopeID},
		{"evidence_id", e.EvidenceID},
		{"action", e.Action},
		{"intended_verifier", e.IntendedVerifier},
		{"issuer_id", e.IssuerID},
		{"issuer_key_id", e.IssuerKeyID},
		{"policy_version", e.PolicyVersion},
		{"schema_version", e.SchemaVersion},
	}
	for _, field := range required {
		if field.value == "" || field.value != strings.TrimSpace(field.value) {
			return ErrInvalidAttestation.Wrapf("evidence envelope %s must be trimmed and nonempty", field.name)
		}
		if len(field.value) > EvidenceEnvelopeMaxString {
			return ErrInvalidAttestation.Wrapf("evidence envelope %s exceeds %d bytes", field.name, EvidenceEnvelopeMaxString)
		}
		if !isPrintableEvidenceEnvelopeASCII(field.value) {
			return ErrInvalidAttestation.Wrapf("evidence envelope %s must contain printable ASCII only", field.name)
		}
	}
	if e.ModelVersion != "" {
		if e.ModelVersion != strings.TrimSpace(e.ModelVersion) || len(e.ModelVersion) > EvidenceEnvelopeMaxString || !isPrintableEvidenceEnvelopeASCII(e.ModelVersion) {
			return ErrInvalidAttestation.Wrap("invalid evidence envelope model_version")
		}
	}
	if !IsValidAttestationType(e.EvidenceType) {
		return ErrInvalidAttestation.Wrapf("invalid evidence envelope type: %s", e.EvidenceType)
	}
	if err := validateWebEvidenceActionPair(e.EvidenceType, e.Action); err != nil {
		return err
	}
	if e.EvidenceType == AttestationTypeInferenceReceipt && e.ModelVersion == "" {
		return ErrInvalidAttestation.Wrap("evidence envelope model_version is required for inference receipts")
	}
	if !IsValidProofType(e.AccountBindingKeyAlgorithm) {
		return ErrInvalidSignerKey.Wrapf("invalid account binding key algorithm: %s", e.AccountBindingKeyAlgorithm)
	}
	if !IsValidProofType(e.IssuerKeyAlgorithm) {
		return ErrInvalidSignerKey.Wrapf("invalid issuer key algorithm: %s", e.IssuerKeyAlgorithm)
	}
	for _, field := range []struct {
		name  string
		value string
	}{
		{"account_binding_key_fingerprint", e.AccountBindingKeyFingerprint},
		{"payload_digest", e.PayloadDigest},
		{"source_context_digest", e.SourceContextDigest},
		{"storage_commitment_digest", e.StorageCommitmentDigest},
		{"issuer_key_fingerprint", e.IssuerKeyFingerprint},
		{"nonce", e.Nonce},
		{"challenge", e.Challenge},
	} {
		if err := validateEvidenceEnvelopeHex(field.value, field.name); err != nil {
			return err
		}
	}
	if e.IssuerKeySequence == 0 {
		return ErrInvalidSignerKey.Wrap("issuer key sequence is required")
	}
	if e.IssuedAtUnix <= 0 || e.ExpiresAtUnix <= e.IssuedAtUnix {
		return ErrInvalidTimestamp.Wrap("invalid evidence envelope time bounds")
	}
	if e.IssuedHeight <= 0 || e.ExpiresHeight <= e.IssuedHeight {
		return ErrInvalidTimestamp.Wrap("invalid evidence envelope height bounds")
	}
	return nil
}

// ValidateAt applies deterministic consensus time and height policy after
// validating the canonical envelope. It never consults a local wall clock.
func (e EvidenceEnvelopeV1) ValidateAt(blockTimeUnix, blockHeight, maxAgeSeconds, maxLifetimeSeconds, maxHeightSpan int64) error {
	if err := e.Validate(); err != nil {
		return err
	}
	if blockTimeUnix <= 0 || blockHeight <= 0 || maxAgeSeconds <= 0 || maxLifetimeSeconds <= 0 || maxHeightSpan <= 0 {
		return ErrInvalidTimestamp.Wrap("evidence envelope consensus values and policy limits must be positive")
	}
	if e.IssuedAtUnix > blockTimeUnix || e.IssuedHeight > blockHeight {
		return ErrInvalidTimestamp.Wrap("evidence envelope issuance is in the future")
	}
	if e.ExpiresAtUnix <= blockTimeUnix || e.ExpiresHeight <= blockHeight {
		return ErrInvalidTimestamp.Wrap("evidence envelope is expired")
	}
	if blockTimeUnix-e.IssuedAtUnix > maxAgeSeconds {
		return ErrInvalidTimestamp.Wrap("evidence envelope exceeds maximum age")
	}
	if e.ExpiresAtUnix-e.IssuedAtUnix > maxLifetimeSeconds {
		return ErrInvalidTimestamp.Wrap("evidence envelope exceeds maximum lifetime")
	}
	if e.ExpiresHeight-e.IssuedHeight > maxHeightSpan {
		return ErrInvalidTimestamp.Wrap("evidence envelope exceeds maximum height span")
	}
	return nil
}

func isPrintableEvidenceEnvelopeASCII(value string) bool {
	for i := 0; i < len(value); i++ {
		if value[i] < 0x20 || value[i] > 0x7e {
			return false
		}
	}
	return true
}

func validateEvidenceEnvelopeHex(value, name string) error {
	if len(value) != sha256.Size*2 || value != strings.ToLower(value) {
		return ErrInvalidAttestation.Wrapf("%s must be a lowercase 64-character SHA256 hex value", name)
	}
	if _, err := hex.DecodeString(value); err != nil {
		return ErrInvalidAttestation.Wrapf("%s must be valid lowercase hex", name)
	}
	return nil
}

func (e EvidenceEnvelopeV1) core() evidenceEnvelopeV1Core {
	return evidenceEnvelopeV1Core{
		Domain: e.Domain, Version: e.Version, ChainID: e.ChainID, AccountAddress: e.AccountAddress,
		AccountBindingKeyFingerprint: e.AccountBindingKeyFingerprint, AccountBindingKeyAlgorithm: e.AccountBindingKeyAlgorithm,
		ScopeID: e.ScopeID, EvidenceType: e.EvidenceType, EvidenceID: e.EvidenceID, Action: e.Action,
		IntendedVerifier: e.IntendedVerifier, PayloadDigest: e.PayloadDigest, SourceContextDigest: e.SourceContextDigest,
		StorageCommitmentDigest: e.StorageCommitmentDigest,
		IssuerID:                e.IssuerID, IssuerKeyID: e.IssuerKeyID, IssuerKeySequence: e.IssuerKeySequence,
		IssuerKeyFingerprint: e.IssuerKeyFingerprint, IssuerKeyAlgorithm: e.IssuerKeyAlgorithm,
		PolicyVersion: e.PolicyVersion, SchemaVersion: e.SchemaVersion, ModelVersion: e.ModelVersion,
		Nonce: e.Nonce, Challenge: e.Challenge, IssuedAtUnix: e.IssuedAtUnix, ExpiresAtUnix: e.ExpiresAtUnix,
		IssuedHeight: e.IssuedHeight, ExpiresHeight: e.ExpiresHeight,
	}
}

func (e EvidenceEnvelopeV1) signBytes(domain string) ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(evidenceEnvelopeV1Sign{SignDomain: domain, evidenceEnvelopeV1Core: e.core()})
}

// IssuerSignBytes returns fresh deterministic bytes for the governed issuer.
func (e EvidenceEnvelopeV1) IssuerSignBytes() ([]byte, error) {
	return e.signBytes(EvidenceEnvelopeIssuerSignDomain)
}

// AccountAuthorizationBytes returns fresh deterministic bytes for the account binding key.
func (e EvidenceEnvelopeV1) AccountAuthorizationBytes() ([]byte, error) {
	return e.signBytes(EvidenceEnvelopeAccountSignDomain)
}

// Digest returns a fresh SHA-256 digest of the canonical envelope core.
func (e EvidenceEnvelopeV1) Digest() ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	bz, err := json.Marshal(e.core())
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(bz)
	return append([]byte(nil), sum[:]...), nil
}

// DigestHex returns the lowercase hexadecimal envelope digest.
func (e EvidenceEnvelopeV1) DigestHex() (string, error) {
	digest, err := e.Digest()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(digest), nil
}

// ReplayContextDigest binds the complete envelope under the replay-context domain.
func (e EvidenceEnvelopeV1) ReplayContextDigest() ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	bz, err := json.Marshal(e.core())
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	_, _ = h.Write([]byte(EvidenceEnvelopeContextDigestDomain))
	_, _ = h.Write(bz)
	return h.Sum(nil), nil
}

// ReplayContextDigestHex returns the lowercase hexadecimal replay-context digest.
func (e EvidenceEnvelopeV1) ReplayContextDigestHex() (string, error) {
	digest, err := e.ReplayContextDigest()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(digest), nil
}

// GlobalNonceDigest binds exactly the envelope version, chain, issuer identity
// and key epoch, nonce, and challenge under a global anti-replay domain.
// Account, scope, and evidence ID are intentionally excluded so reusing the
// same issuer nonce across those contexts is rejected globally.
func (e EvidenceEnvelopeV1) GlobalNonceDigest() ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	bz, err := json.Marshal(evidenceEnvelopeV1GlobalNonce{
		Domain: EvidenceEnvelopeGlobalNonceDomain, EnvelopeVersion: e.Version, ChainID: e.ChainID,
		IssuerID: e.IssuerID, IssuerKeyID: e.IssuerKeyID, IssuerKeySequence: e.IssuerKeySequence,
		IssuerKeyFingerprint: e.IssuerKeyFingerprint, IssuerKeyAlgorithm: e.IssuerKeyAlgorithm,
		Nonce: e.Nonce, Challenge: e.Challenge,
	})
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(bz)
	return append([]byte(nil), sum[:]...), nil
}

// GlobalNonceDigestHex returns the lowercase hexadecimal global nonce digest.
func (e EvidenceEnvelopeV1) GlobalNonceDigestHex() (string, error) {
	digest, err := e.GlobalNonceDigest()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(digest), nil
}
