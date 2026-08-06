package types

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func evidenceEnvelopeTestValue() EvidenceEnvelopeV1 {
	return EvidenceEnvelopeV1{
		Domain:                       EvidenceEnvelopeDomain,
		Version:                      EvidenceEnvelopeVersion,
		ChainID:                      "ve-test-1",
		AccountAddress:               "virtengine1account",
		AccountBindingKeyFingerprint: strings.Repeat("1", 64),
		AccountBindingKeyAlgorithm:   ProofTypeSecp256k1,
		ScopeID:                      "scope-email-001",
		EvidenceType:                 AttestationTypeEmailVerification,
		EvidenceID:                   "evidence-001",
		Action:                       WebEvidenceActionSubmitEmail,
		IntendedVerifier:             "veid-module",
		PayloadDigest:                strings.Repeat("2", 64),
		SourceContextDigest:          strings.Repeat("7", 64),
		StorageCommitmentDigest:      strings.Repeat("3", 64),
		IssuerID:                     "did:virtengine:issuer:web",
		IssuerKeyID:                  "issuer-web:7",
		IssuerKeySequence:            7,
		IssuerKeyFingerprint:         strings.Repeat("4", 64),
		IssuerKeyAlgorithm:           ProofTypeEd25519,
		PolicyVersion:                "policy-2026-01",
		SchemaVersion:                "schema-1.0.0",
		ModelVersion:                 "model-3",
		Nonce:                        strings.Repeat("5", 64),
		Challenge:                    strings.Repeat("6", 64),
		IssuedAtUnix:                 1_710_000_000,
		ExpiresAtUnix:                1_710_003_600,
		IssuedHeight:                 123_456,
		ExpiresHeight:                123_556,
	}
}

func TestEvidenceEnvelopeGoldenCanonicalBytesAndDigests(t *testing.T) {
	envelope := evidenceEnvelopeTestValue()

	issuerBytes, err := envelope.IssuerSignBytes()
	require.NoError(t, err)
	accountBytes, err := envelope.AccountAuthorizationBytes()
	require.NoError(t, err)

	const issuerGolden = `{"sign_domain":"VEID_EVIDENCE_ENVELOPE_ISSUER_SIGN_V1","domain":"VEID_EVIDENCE_ENVELOPE_V1","version":"1","chain_id":"ve-test-1","account_address":"virtengine1account","account_binding_key_fingerprint":"1111111111111111111111111111111111111111111111111111111111111111","account_binding_key_algorithm":"EcdsaSecp256k1Signature2019","scope_id":"scope-email-001","evidence_type":"email_verification","evidence_id":"evidence-001","action":"submit_email_verification_proof","intended_verifier":"veid-module","payload_digest":"2222222222222222222222222222222222222222222222222222222222222222","source_context_digest":"7777777777777777777777777777777777777777777777777777777777777777","storage_commitment_digest":"3333333333333333333333333333333333333333333333333333333333333333","issuer_id":"did:virtengine:issuer:web","issuer_key_id":"issuer-web:7","issuer_key_sequence":7,"issuer_key_fingerprint":"4444444444444444444444444444444444444444444444444444444444444444","issuer_key_algorithm":"Ed25519Signature2020","policy_version":"policy-2026-01","schema_version":"schema-1.0.0","model_version":"model-3","nonce":"5555555555555555555555555555555555555555555555555555555555555555","challenge":"6666666666666666666666666666666666666666666666666666666666666666","issued_at_unix":1710000000,"expires_at_unix":1710003600,"issued_height":123456,"expires_height":123556}`
	const accountGolden = `{"sign_domain":"VEID_EVIDENCE_ENVELOPE_ACCOUNT_SIGN_V1","domain":"VEID_EVIDENCE_ENVELOPE_V1","version":"1","chain_id":"ve-test-1","account_address":"virtengine1account","account_binding_key_fingerprint":"1111111111111111111111111111111111111111111111111111111111111111","account_binding_key_algorithm":"EcdsaSecp256k1Signature2019","scope_id":"scope-email-001","evidence_type":"email_verification","evidence_id":"evidence-001","action":"submit_email_verification_proof","intended_verifier":"veid-module","payload_digest":"2222222222222222222222222222222222222222222222222222222222222222","source_context_digest":"7777777777777777777777777777777777777777777777777777777777777777","storage_commitment_digest":"3333333333333333333333333333333333333333333333333333333333333333","issuer_id":"did:virtengine:issuer:web","issuer_key_id":"issuer-web:7","issuer_key_sequence":7,"issuer_key_fingerprint":"4444444444444444444444444444444444444444444444444444444444444444","issuer_key_algorithm":"Ed25519Signature2020","policy_version":"policy-2026-01","schema_version":"schema-1.0.0","model_version":"model-3","nonce":"5555555555555555555555555555555555555555555555555555555555555555","challenge":"6666666666666666666666666666666666666666666666666666666666666666","issued_at_unix":1710000000,"expires_at_unix":1710003600,"issued_height":123456,"expires_height":123556}`
	require.Equal(t, issuerGolden, string(issuerBytes))
	require.Equal(t, accountGolden, string(accountBytes))

	digest, err := envelope.DigestHex()
	require.NoError(t, err)
	replay, err := envelope.ReplayContextDigestHex()
	require.NoError(t, err)
	globalNonce, err := envelope.GlobalNonceDigestHex()
	require.NoError(t, err)
	require.Equal(t, "ea1fa8ff1c2c97e503f04db317878d1e62adb9e401fe04a7365072fb630e59a7", digest)
	require.Equal(t, "147fd4d08b4368bd8544876d9143bf828c2cbd26fc9bd54b54f1d9223891f17f", replay)
	require.Equal(t, "9e74dd8c4900c56011c5e197cb428a8b3f9b389755ef289c2d18d33f436fda40", globalNonce)
}

func TestEvidenceEnvelopeDeterminismAndFreshSlices(t *testing.T) {
	envelope := evidenceEnvelopeTestValue()
	tests := []struct {
		name string
		get  func() ([]byte, error)
	}{
		{"issuer", envelope.IssuerSignBytes},
		{"account", envelope.AccountAuthorizationBytes},
		{"digest", envelope.Digest},
		{"replay", envelope.ReplayContextDigest},
		{"global nonce", envelope.GlobalNonceDigest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			first, err := test.get()
			require.NoError(t, err)
			original := append([]byte(nil), first...)
			second, err := test.get()
			require.NoError(t, err)
			require.Equal(t, original, second)
			first[0] ^= 0xff
			third, err := test.get()
			require.NoError(t, err)
			require.Equal(t, original, third)
		})
	}
}

func TestEvidenceEnvelopeValidationFailures(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*EvidenceEnvelopeV1)
	}{
		{"domain", func(e *EvidenceEnvelopeV1) { e.Domain = "wrong" }},
		{"version", func(e *EvidenceEnvelopeV1) { e.Version = "2" }},
		{"chain empty", func(e *EvidenceEnvelopeV1) { e.ChainID = "" }},
		{"chain untrimmed", func(e *EvidenceEnvelopeV1) { e.ChainID = " ve-test-1" }},
		{"chain too long", func(e *EvidenceEnvelopeV1) { e.ChainID = strings.Repeat("x", EvidenceEnvelopeMaxString+1) }},
		{"account", func(e *EvidenceEnvelopeV1) { e.AccountAddress = "" }},
		{"scope", func(e *EvidenceEnvelopeV1) { e.ScopeID = "" }},
		{"evidence id", func(e *EvidenceEnvelopeV1) { e.EvidenceID = "" }},
		{"action", func(e *EvidenceEnvelopeV1) { e.Action = "" }},
		{"verifier", func(e *EvidenceEnvelopeV1) { e.IntendedVerifier = "" }},
		{"issuer id", func(e *EvidenceEnvelopeV1) { e.IssuerID = "" }},
		{"issuer key id", func(e *EvidenceEnvelopeV1) { e.IssuerKeyID = "" }},
		{"policy", func(e *EvidenceEnvelopeV1) { e.PolicyVersion = "" }},
		{"schema", func(e *EvidenceEnvelopeV1) { e.SchemaVersion = "" }},
		{"model untrimmed", func(e *EvidenceEnvelopeV1) { e.ModelVersion = " model-3" }},
		{"model too long", func(e *EvidenceEnvelopeV1) { e.ModelVersion = strings.Repeat("m", EvidenceEnvelopeMaxString+1) }},
		{"evidence type zero", func(e *EvidenceEnvelopeV1) { e.EvidenceType = "" }},
		{"evidence type unknown", func(e *EvidenceEnvelopeV1) { e.EvidenceType = "unknown" }},
		{"account algorithm zero", func(e *EvidenceEnvelopeV1) { e.AccountBindingKeyAlgorithm = "" }},
		{"account algorithm unknown", func(e *EvidenceEnvelopeV1) { e.AccountBindingKeyAlgorithm = "unknown" }},
		{"issuer algorithm zero", func(e *EvidenceEnvelopeV1) { e.IssuerKeyAlgorithm = "" }},
		{"issuer algorithm unknown", func(e *EvidenceEnvelopeV1) { e.IssuerKeyAlgorithm = "unknown" }},
		{"account fingerprint empty", func(e *EvidenceEnvelopeV1) { e.AccountBindingKeyFingerprint = "" }},
		{"account fingerprint uppercase", func(e *EvidenceEnvelopeV1) { e.AccountBindingKeyFingerprint = strings.Repeat("A", 64) }},
		{"issuer fingerprint malformed", func(e *EvidenceEnvelopeV1) { e.IssuerKeyFingerprint = strings.Repeat("z", 64) }},
		{"payload digest short", func(e *EvidenceEnvelopeV1) { e.PayloadDigest = strings.Repeat("2", 63) }},
		{"source digest empty", func(e *EvidenceEnvelopeV1) { e.SourceContextDigest = "" }},
		{"source digest uppercase", func(e *EvidenceEnvelopeV1) { e.SourceContextDigest = strings.Repeat("A", 64) }},
		{"storage digest uppercase", func(e *EvidenceEnvelopeV1) { e.StorageCommitmentDigest = strings.Repeat("B", 64) }},
		{"nonce empty", func(e *EvidenceEnvelopeV1) { e.Nonce = "" }},
		{"nonce malformed", func(e *EvidenceEnvelopeV1) { e.Nonce = strings.Repeat("g", 64) }},
		{"challenge uppercase", func(e *EvidenceEnvelopeV1) { e.Challenge = strings.Repeat("C", 64) }},
		{"issuer sequence", func(e *EvidenceEnvelopeV1) { e.IssuerKeySequence = 0 }},
		{"issued time zero", func(e *EvidenceEnvelopeV1) { e.IssuedAtUnix = 0 }},
		{"expiry time equal", func(e *EvidenceEnvelopeV1) { e.ExpiresAtUnix = e.IssuedAtUnix }},
		{"expiry time before", func(e *EvidenceEnvelopeV1) { e.ExpiresAtUnix = e.IssuedAtUnix - 1 }},
		{"issued height zero", func(e *EvidenceEnvelopeV1) { e.IssuedHeight = 0 }},
		{"expiry height equal", func(e *EvidenceEnvelopeV1) { e.ExpiresHeight = e.IssuedHeight }},
		{"expiry height before", func(e *EvidenceEnvelopeV1) { e.ExpiresHeight = e.IssuedHeight - 1 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			envelope := evidenceEnvelopeTestValue()
			test.mutate(&envelope)
			require.Error(t, envelope.Validate())
		})
	}
}

func TestEvidenceEnvelopeHumanReadableFieldsRequirePrintableASCII(t *testing.T) {
	fields := []struct {
		name string
		set  func(*EvidenceEnvelopeV1, string)
	}{
		{"chain", func(e *EvidenceEnvelopeV1, value string) { e.ChainID = value }},
		{"account", func(e *EvidenceEnvelopeV1, value string) { e.AccountAddress = value }},
		{"scope", func(e *EvidenceEnvelopeV1, value string) { e.ScopeID = value }},
		{"evidence id", func(e *EvidenceEnvelopeV1, value string) { e.EvidenceID = value }},
		{"action", func(e *EvidenceEnvelopeV1, value string) { e.Action = value }},
		{"verifier", func(e *EvidenceEnvelopeV1, value string) { e.IntendedVerifier = value }},
		{"issuer id", func(e *EvidenceEnvelopeV1, value string) { e.IssuerID = value }},
		{"issuer key id", func(e *EvidenceEnvelopeV1, value string) { e.IssuerKeyID = value }},
		{"policy", func(e *EvidenceEnvelopeV1, value string) { e.PolicyVersion = value }},
		{"schema", func(e *EvidenceEnvelopeV1, value string) { e.SchemaVersion = value }},
		{"model", func(e *EvidenceEnvelopeV1, value string) { e.ModelVersion = value }},
	}
	invalidValues := []struct {
		name  string
		value string
	}{
		{"control", "value\x1fvalue"},
		{"nul", "value\x00value"},
		{"non-ASCII byte", "value" + string([]byte{0xff})},
		{"Unicode composed", "caf\u00e9"},
		{"Unicode decomposed", "cafe\u0301"},
	}
	for _, field := range fields {
		for _, invalid := range invalidValues {
			t.Run(field.name+"/"+invalid.name, func(t *testing.T) {
				envelope := evidenceEnvelopeTestValue()
				field.set(&envelope, invalid.value)
				require.Error(t, envelope.Validate())
			})
		}
	}
}

func TestEvidenceEnvelopeInferenceReceiptRequiresModelVersion(t *testing.T) {
	for _, test := range []struct {
		evidenceType AttestationType
		action       string
	}{
		{AttestationTypeEmailVerification, WebEvidenceActionSubmitEmail},
		{AttestationTypeSMSVerification, WebEvidenceActionSubmitSMS},
		{AttestationTypeSSOVerification, WebEvidenceActionSubmitSSO},
		{AttestationTypeSocialMediaVerification, WebEvidenceActionSubmitSocial},
	} {
		envelope := evidenceEnvelopeTestValue()
		envelope.EvidenceType = test.evidenceType
		envelope.Action = test.action
		envelope.ModelVersion = ""
		require.NoError(t, envelope.Validate())
	}

	inference := evidenceEnvelopeTestValue()
	inference.EvidenceType = AttestationTypeInferenceReceipt
	inference.Action = "submit_inference_receipt"
	inference.ModelVersion = ""
	require.Error(t, inference.Validate())
	inference.ModelVersion = "model-3"
	require.NoError(t, inference.Validate())
}

func TestEvidenceEnvelopeWebActionTypePairing(t *testing.T) {
	valid := []struct {
		evidenceType AttestationType
		action       string
	}{
		{AttestationTypeSSOVerification, WebEvidenceActionSubmitSSO},
		{AttestationTypeEmailVerification, WebEvidenceActionSubmitEmail},
		{AttestationTypeSMSVerification, WebEvidenceActionSubmitSMS},
		{AttestationTypeSocialMediaVerification, WebEvidenceActionSubmitSocial},
	}
	for _, test := range valid {
		require.NoError(t, validateWebEvidenceActionPair(test.evidenceType, test.action))
	}

	require.Error(t, validateWebEvidenceActionPair(AttestationTypeEmailVerification, WebEvidenceActionSubmitSMS))
	require.Error(t, validateWebEvidenceActionPair(AttestationTypeDocumentVerification, WebEvidenceActionSubmitEmail))
	require.NoError(t, validateWebEvidenceActionPair(AttestationTypeDocumentVerification, "submit_document_verification"))
}

func TestEvidenceEnvelopeEveryFieldTamper(t *testing.T) {
	base := evidenceEnvelopeTestValue()
	baseIssuer, err := base.IssuerSignBytes()
	require.NoError(t, err)
	baseAccount, err := base.AccountAuthorizationBytes()
	require.NoError(t, err)
	baseDigest, err := base.Digest()
	require.NoError(t, err)
	baseReplay, err := base.ReplayContextDigest()
	require.NoError(t, err)
	baseGlobal, err := base.GlobalNonceDigest()
	require.NoError(t, err)

	tests := []struct {
		name          string
		globalChanges bool
		mutate        func(*EvidenceEnvelopeV1)
	}{
		{"domain", false, func(e *EvidenceEnvelopeV1) { e.Domain = "other" }},
		{"version", false, func(e *EvidenceEnvelopeV1) { e.Version = "2" }},
		{"chain id", true, func(e *EvidenceEnvelopeV1) { e.ChainID = "ve-test-2" }},
		{"account address", false, func(e *EvidenceEnvelopeV1) { e.AccountAddress += "2" }},
		{"account fingerprint", false, func(e *EvidenceEnvelopeV1) { e.AccountBindingKeyFingerprint = strings.Repeat("7", 64) }},
		{"account algorithm", false, func(e *EvidenceEnvelopeV1) { e.AccountBindingKeyAlgorithm = ProofTypeEd25519 }},
		{"scope id", false, func(e *EvidenceEnvelopeV1) { e.ScopeID += "2" }},
		{"evidence type", false, func(e *EvidenceEnvelopeV1) { e.EvidenceType = AttestationTypeSMSVerification }},
		{"evidence id", false, func(e *EvidenceEnvelopeV1) { e.EvidenceID += "2" }},
		{"action", false, func(e *EvidenceEnvelopeV1) { e.Action = WebEvidenceActionSubmitSMS }},
		{"intended verifier", false, func(e *EvidenceEnvelopeV1) { e.IntendedVerifier += "2" }},
		{"payload digest", false, func(e *EvidenceEnvelopeV1) { e.PayloadDigest = strings.Repeat("7", 64) }},
		{"source context digest", false, func(e *EvidenceEnvelopeV1) { e.SourceContextDigest = strings.Repeat("8", 64) }},
		{"storage commitment", false, func(e *EvidenceEnvelopeV1) { e.StorageCommitmentDigest = strings.Repeat("9", 64) }},
		{"issuer id", true, func(e *EvidenceEnvelopeV1) { e.IssuerID += "2" }},
		{"issuer key id", true, func(e *EvidenceEnvelopeV1) { e.IssuerKeyID += "2" }},
		{"issuer sequence", true, func(e *EvidenceEnvelopeV1) { e.IssuerKeySequence++ }},
		{"issuer fingerprint", true, func(e *EvidenceEnvelopeV1) { e.IssuerKeyFingerprint = strings.Repeat("9", 64) }},
		{"issuer algorithm", true, func(e *EvidenceEnvelopeV1) { e.IssuerKeyAlgorithm = ProofTypeSr25519 }},
		{"policy version", false, func(e *EvidenceEnvelopeV1) { e.PolicyVersion += "2" }},
		{"schema version", false, func(e *EvidenceEnvelopeV1) { e.SchemaVersion += "2" }},
		{"model version", false, func(e *EvidenceEnvelopeV1) { e.ModelVersion += "2" }},
		{"nonce", true, func(e *EvidenceEnvelopeV1) { e.Nonce = strings.Repeat("a", 64) }},
		{"challenge", true, func(e *EvidenceEnvelopeV1) { e.Challenge = strings.Repeat("b", 64) }},
		{"issued time", false, func(e *EvidenceEnvelopeV1) { e.IssuedAtUnix++ }},
		{"expires time", false, func(e *EvidenceEnvelopeV1) { e.ExpiresAtUnix++ }},
		{"issued height", false, func(e *EvidenceEnvelopeV1) { e.IssuedHeight++ }},
		{"expires height", false, func(e *EvidenceEnvelopeV1) { e.ExpiresHeight++ }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := base
			test.mutate(&mutated)
			if err := mutated.Validate(); err != nil {
				require.Contains(t, []string{"domain", "version", "evidence type", "action"}, test.name)
				return
			}
			issuer, err := mutated.IssuerSignBytes()
			require.NoError(t, err)
			account, err := mutated.AccountAuthorizationBytes()
			require.NoError(t, err)
			digest, err := mutated.Digest()
			require.NoError(t, err)
			replay, err := mutated.ReplayContextDigest()
			require.NoError(t, err)
			global, err := mutated.GlobalNonceDigest()
			require.NoError(t, err)
			require.NotEqual(t, baseIssuer, issuer)
			require.NotEqual(t, baseAccount, account)
			require.NotEqual(t, baseDigest, digest)
			require.NotEqual(t, baseReplay, replay)
			if test.globalChanges {
				require.NotEqual(t, baseGlobal, global)
			} else {
				require.Equal(t, baseGlobal, global)
			}
		})
	}
}

func TestEvidenceEnvelopeGlobalNonceTuple(t *testing.T) {
	base := evidenceEnvelopeTestValue()
	baseDigest, err := base.GlobalNonceDigest()
	require.NoError(t, err)

	for _, test := range []struct {
		name    string
		changes bool
		mutate  func(*EvidenceEnvelopeV1)
	}{
		{"account", false, func(e *EvidenceEnvelopeV1) { e.AccountAddress += "2" }},
		{"scope", false, func(e *EvidenceEnvelopeV1) { e.ScopeID += "2" }},
		{"evidence id", false, func(e *EvidenceEnvelopeV1) { e.EvidenceID += "2" }},
		{"chain", true, func(e *EvidenceEnvelopeV1) { e.ChainID += "2" }},
		{"issuer epoch", true, func(e *EvidenceEnvelopeV1) { e.IssuerKeySequence++ }},
		{"nonce", true, func(e *EvidenceEnvelopeV1) { e.Nonce = strings.Repeat("a", 64) }},
		{"challenge", true, func(e *EvidenceEnvelopeV1) { e.Challenge = strings.Repeat("b", 64) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			mutated := base
			test.mutate(&mutated)
			digest, err := mutated.GlobalNonceDigest()
			require.NoError(t, err)
			if test.changes {
				require.NotEqual(t, baseDigest, digest)
			} else {
				require.Equal(t, baseDigest, digest)
			}
		})
	}
}

func TestEvidenceEnvelopeValidateAtBoundaries(t *testing.T) {
	base := evidenceEnvelopeTestValue()
	tests := []struct {
		name                                                 string
		blockTime, blockHeight, maxAge, maxLifetime, maxSpan int64
		wantErr                                              bool
	}{
		{"exact limits", base.IssuedAtUnix + 60, base.IssuedHeight + 1, 60, 3600, 100, false},
		{"future issuance time", base.IssuedAtUnix - 1, base.IssuedHeight, 60, 3600, 100, true},
		{"future issuance height", base.IssuedAtUnix, base.IssuedHeight - 1, 60, 3600, 100, true},
		{"expiration time boundary", base.ExpiresAtUnix, base.IssuedHeight + 1, 3600, 3600, 100, true},
		{"expiration height boundary", base.IssuedAtUnix, base.ExpiresHeight, 3600, 3600, 100, true},
		{"age exceeded", base.IssuedAtUnix + 61, base.IssuedHeight + 1, 60, 3600, 100, true},
		{"lifetime exceeded", base.IssuedAtUnix, base.IssuedHeight, 60, 3599, 100, true},
		{"height span exceeded", base.IssuedAtUnix, base.IssuedHeight, 60, 3600, 99, true},
		{"zero block time", 0, base.IssuedHeight, 60, 3600, 100, true},
		{"zero block height", base.IssuedAtUnix, 0, 60, 3600, 100, true},
		{"zero max age", base.IssuedAtUnix, base.IssuedHeight, 0, 3600, 100, true},
		{"zero max lifetime", base.IssuedAtUnix, base.IssuedHeight, 60, 0, 100, true},
		{"negative max height span", base.IssuedAtUnix, base.IssuedHeight, 60, 3600, -1, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := base.ValidateAt(test.blockTime, test.blockHeight, test.maxAge, test.maxLifetime, test.maxSpan)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEvidenceEnvelopeWebEvidenceAdapter(t *testing.T) {
	issuedAt := time.Unix(1_710_000_000, 0).UTC()
	web := NewWebEvidenceContext(WebEvidenceContextConfig{
		ChainID: "ve-test-1", AccountAddress: "virtengine1account",
		EvidenceType: AttestationTypeEmailVerification, Action: WebEvidenceActionSubmitEmail,
		ScopeID: "scope-email-001", AttestationDigest: strings.Repeat("2", 64),
		Issuer:          AttestationIssuer{ID: "did:virtengine:issuer:web", KeyID: "issuer-web:7", KeyFingerprint: strings.Repeat("4", 64)},
		IssuerAlgorithm: ProofTypeEd25519, Nonce: strings.Repeat("5", 64), Challenge: strings.Repeat("6", 64),
		IssuedAt: issuedAt, ExpiresAt: issuedAt.Add(time.Hour),
		ServiceMetadataHash: strings.Repeat("8", 64),
		CallerFields: map[string]string{
			"email":       "alice@example.com",
			"provider_id": "provider-7",
		},
	})
	cfg := EvidenceEnvelopeV1Config{
		AccountBindingKeyFingerprint: strings.Repeat("1", 64), AccountBindingKeyAlgorithm: ProofTypeSecp256k1,
		IssuerKeySequence: 7, EvidenceID: "evidence-001", StorageCommitmentDigest: strings.Repeat("3", 64),
		PolicyVersion: "policy-2026-01", SchemaVersion: "schema-1.0.0", ModelVersion: "model-3",
		IntendedVerifier: "veid-module", IssuedHeight: 123_456, ExpiresHeight: 123_556,
	}

	envelope, err := NewEvidenceEnvelopeV1FromWebEvidence(web, cfg)
	require.NoError(t, err)
	require.Equal(t, web.ChainID, envelope.ChainID)
	require.Equal(t, web.AccountAddress, envelope.AccountAddress)
	require.Equal(t, web.ScopeID, envelope.ScopeID)
	require.Equal(t, web.EvidenceType, envelope.EvidenceType)
	require.Equal(t, web.Action, envelope.Action)
	require.Equal(t, web.AttestationDigest, envelope.PayloadDigest)
	sourceBytes, err := web.IssuerSignBytes()
	require.NoError(t, err)
	sourceHash := sha256.New()
	_, _ = sourceHash.Write([]byte(EvidenceEnvelopeWebSourceDomain))
	_, _ = sourceHash.Write(sourceBytes)
	require.Equal(t, hex.EncodeToString(sourceHash.Sum(nil)), envelope.SourceContextDigest)
	require.Equal(t, web.IssuerID, envelope.IssuerID)
	require.Equal(t, web.IssuerKeyID, envelope.IssuerKeyID)
	require.Equal(t, web.IssuerFingerprint, envelope.IssuerKeyFingerprint)
	require.Equal(t, web.IssuerAlgorithm, envelope.IssuerKeyAlgorithm)
	require.Equal(t, web.Nonce, envelope.Nonce)
	require.Equal(t, web.Challenge, envelope.Challenge)
	require.Equal(t, web.IssuedAt.Unix(), envelope.IssuedAtUnix)
	require.Equal(t, web.ExpiresAt.Unix(), envelope.ExpiresAtUnix)
	require.Equal(t, cfg.StorageCommitmentDigest, envelope.StorageCommitmentDigest)

	for _, mutate := range []func(*EvidenceEnvelopeV1Config){
		func(c *EvidenceEnvelopeV1Config) { c.AccountBindingKeyFingerprint = "" },
		func(c *EvidenceEnvelopeV1Config) { c.StorageCommitmentDigest = "" },
	} {
		invalid := cfg
		mutate(&invalid)
		_, err := NewEvidenceEnvelopeV1FromWebEvidence(web, invalid)
		require.Error(t, err)
	}
}

func TestEvidenceEnvelopeWebEvidenceAdapterCommitsSignedSourceMetadata(t *testing.T) {
	issuedAt := time.Unix(1_710_000_000, 0).UTC()
	web := NewWebEvidenceContext(WebEvidenceContextConfig{
		ChainID: "ve-test-1", AccountAddress: "virtengine1account",
		EvidenceType: AttestationTypeEmailVerification, Action: WebEvidenceActionSubmitEmail,
		ScopeID: "scope-email-001", AttestationDigest: strings.Repeat("2", 64),
		Issuer:          AttestationIssuer{ID: "did:virtengine:issuer:web", KeyID: "issuer-web:7", KeyFingerprint: strings.Repeat("4", 64)},
		IssuerAlgorithm: ProofTypeEd25519, Nonce: strings.Repeat("5", 64), Challenge: strings.Repeat("6", 64),
		IssuedAt: issuedAt, ExpiresAt: issuedAt.Add(time.Hour), ServiceMetadataHash: strings.Repeat("8", 64),
		CallerFields: map[string]string{"email": "alice@example.com", "provider_id": "provider-7"},
	})
	cfg := EvidenceEnvelopeV1Config{
		AccountBindingKeyFingerprint: strings.Repeat("1", 64), AccountBindingKeyAlgorithm: ProofTypeSecp256k1,
		IssuerKeySequence: 7, EvidenceID: "evidence-001", StorageCommitmentDigest: strings.Repeat("3", 64),
		PolicyVersion: "policy-2026-01", SchemaVersion: "schema-1.0.0", ModelVersion: "model-3",
		IntendedVerifier: "veid-module", IssuedHeight: 123_456, ExpiresHeight: 123_556,
	}
	base, err := NewEvidenceEnvelopeV1FromWebEvidence(web, cfg)
	require.NoError(t, err)
	baseSign, err := base.IssuerSignBytes()
	require.NoError(t, err)
	baseDigest, err := base.Digest()
	require.NoError(t, err)

	mutations := []struct {
		name   string
		mutate func(*WebEvidenceContext)
	}{
		{"service metadata hash", func(c *WebEvidenceContext) { c.ServiceMetadataHash = strings.Repeat("9", 64) }},
	}
	for index := range web.CallerFields {
		index := index
		mutations = append(mutations,
			struct {
				name   string
				mutate func(*WebEvidenceContext)
			}{"caller field " + web.CallerFields[index].Name + " name", func(c *WebEvidenceContext) { c.CallerFields[index].Name += "2" }},
			struct {
				name   string
				mutate func(*WebEvidenceContext)
			}{"caller field " + web.CallerFields[index].Name + " value", func(c *WebEvidenceContext) { c.CallerFields[index].Value += "2" }},
		)
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			mutatedWeb := web
			mutatedWeb.CallerFields = append([]WebEvidenceField(nil), web.CallerFields...)
			mutation.mutate(&mutatedWeb)
			mutated, err := NewEvidenceEnvelopeV1FromWebEvidence(mutatedWeb, cfg)
			require.NoError(t, err)
			require.Equal(t, base.PayloadDigest, mutated.PayloadDigest)
			require.NotEqual(t, base.SourceContextDigest, mutated.SourceContextDigest)
			mutatedSign, err := mutated.IssuerSignBytes()
			require.NoError(t, err)
			mutatedDigest, err := mutated.Digest()
			require.NoError(t, err)
			require.NotEqual(t, baseSign, mutatedSign)
			require.NotEqual(t, baseDigest, mutatedDigest)
		})
	}
}
