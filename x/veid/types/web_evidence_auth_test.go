package types

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const webEvidenceTestMetadataValue = "value"

func TestWebEvidenceContextCanonicalBytesAreDeterministicAndDomainSeparated(t *testing.T) {
	issuedAt := time.Unix(1710000000, 0).UTC()
	expiresAt := issuedAt.Add(time.Hour)
	issuer := AttestationIssuer{
		ID:             "did:virtengine:issuer:web",
		KeyID:          "issuer-web:1",
		KeyFingerprint: hex.EncodeToString(make([]byte, 32)),
	}

	ctxA := NewWebEvidenceContext(WebEvidenceContextConfig{
		ChainID:             "ve-test-1",
		AccountAddress:      "virtengine1account",
		EvidenceType:        AttestationTypeEmailVerification,
		Action:              WebEvidenceActionSubmitEmail,
		ScopeID:             "email-1",
		AttestationDigest:   hex.EncodeToString([]byte("01234567890123456789012345678901")),
		Issuer:              issuer,
		IssuerAlgorithm:     ProofTypeEd25519,
		Nonce:               "nonce-hex",
		Challenge:           "challenge-1",
		IssuedAt:            issuedAt,
		ExpiresAt:           expiresAt,
		ServiceMetadataHash: hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz012345")),
		CallerFields: map[string]string{
			"email_hash":        "email-hash",
			"is_organizational": "true",
		},
	})
	ctxB := NewWebEvidenceContext(WebEvidenceContextConfig{
		ChainID:             "ve-test-1",
		AccountAddress:      "virtengine1account",
		EvidenceType:        AttestationTypeEmailVerification,
		Action:              WebEvidenceActionSubmitEmail,
		ScopeID:             "email-1",
		AttestationDigest:   hex.EncodeToString([]byte("01234567890123456789012345678901")),
		Issuer:              issuer,
		IssuerAlgorithm:     ProofTypeEd25519,
		Nonce:               "nonce-hex",
		Challenge:           "challenge-1",
		IssuedAt:            issuedAt,
		ExpiresAt:           expiresAt,
		ServiceMetadataHash: hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz012345")),
		CallerFields: map[string]string{
			"is_organizational": "true",
			"email_hash":        "email-hash",
		},
	})

	issuerBytesA, err := ctxA.IssuerSignBytes()
	require.NoError(t, err)
	issuerBytesB, err := ctxB.IssuerSignBytes()
	require.NoError(t, err)
	accountBytes, err := ctxA.AccountAuthorizationBytes()
	require.NoError(t, err)

	require.Equal(t, issuerBytesA, issuerBytesB)
	require.NotEqual(t, issuerBytesA, accountBytes)
	require.Contains(t, string(issuerBytesA), WebEvidenceIssuerDomain)
	require.Contains(t, string(accountBytes), WebEvidenceAccountDomain)
}

func TestWebEvidenceContextStrictMetadataValidation(t *testing.T) {
	issuedAt := time.Unix(1710000000, 0).UTC()
	expiresAt := issuedAt.Add(time.Hour)
	issuer := AttestationIssuer{
		ID:             "did:virtengine:issuer:web",
		KeyID:          "issuer-web:1",
		KeyFingerprint: hex.EncodeToString(make([]byte, 32)),
	}
	subject := NewAttestationSubject("virtengine1account")
	att := NewVerificationAttestation(
		issuer,
		subject,
		AttestationTypeEmailVerification,
		[]byte("nonce-for-web-evidence-123456789"),
		issuedAt,
		time.Hour,
		100,
		100,
	)
	digest, err := WebEvidenceAttestationDigestHex(att)
	require.NoError(t, err)

	evidence := NewWebEvidenceContext(WebEvidenceContextConfig{
		ChainID:             "ve-test-1",
		AccountAddress:      subject.AccountAddress,
		EvidenceType:        AttestationTypeEmailVerification,
		Action:              WebEvidenceActionSubmitEmail,
		ScopeID:             "email-1",
		AttestationDigest:   digest,
		Issuer:              issuer,
		IssuerAlgorithm:     ProofTypeEd25519,
		Nonce:               att.Nonce,
		Challenge:           "email-challenge",
		IssuedAt:            issuedAt,
		ExpiresAt:           expiresAt,
		ServiceMetadataHash: mustWebEvidenceServiceMetadataHash(t, map[string]string{"source": "unit-test"}),
		CallerFields: map[string]string{
			"email_hash": "hash-1",
		},
	})

	require.NoError(t, evidence.ApplyToAttestation(att))
	require.NoError(t, evidence.ValidateAttestationMetadata(att))

	att.Metadata[WebEvidenceMetadataFieldPrefix+"email_hash"] = "tampered"
	require.Error(t, evidence.ValidateAttestationMetadata(att))

	require.NoError(t, evidence.ApplyToAttestation(att))
	delete(att.Metadata, WebEvidenceMetadataIssuerKeyID)
	require.Error(t, evidence.ValidateAttestationMetadata(att))

	require.NoError(t, evidence.ApplyToAttestation(att))
	att.Metadata[WebEvidenceMetadataPrefix+"unexpected"] = webEvidenceTestMetadataValue
	require.Error(t, evidence.ValidateAttestationMetadata(att))

	require.NoError(t, evidence.ApplyToAttestation(att))
	att.Metadata["unsigned.extra"] = webEvidenceTestMetadataValue
	require.Error(t, evidence.ValidateAttestationMetadata(att))
}

func TestWebEvidenceServiceMetadataHashUsesSortedCanonicalFields(t *testing.T) {
	first, err := WebEvidenceServiceMetadataHash(map[string]string{
		"zeta":  "last",
		"alpha": "first",
		"mid":   "middle",
	})
	require.NoError(t, err)
	second, err := WebEvidenceServiceMetadataHash(map[string]string{
		"mid":   "middle",
		"zeta":  "last",
		"alpha": "first",
	})
	require.NoError(t, err)
	changed, err := WebEvidenceServiceMetadataHash(map[string]string{
		"alpha": "first",
		"mid":   "middle",
		"zeta":  "changed",
	})
	require.NoError(t, err)

	require.Equal(t, first, second)
	require.Equal(t, "1ea696b57d21010b39b829700c3573cb6ca97cb1f7f963926ce1ffa5598ea39b", first)
	require.NotEqual(t, first, changed)
}

func TestWebEvidenceServiceMetadataHashAllowsEmptyMetadata(t *testing.T) {
	hash, err := WebEvidenceServiceMetadataHash(nil)
	require.NoError(t, err)
	require.Empty(t, hash)

	hash, err = WebEvidenceServiceMetadataHash(map[string]string{})
	require.NoError(t, err)
	require.Empty(t, hash)
}

func TestWebEvidenceServiceMetadataHashRejectsInvalidKeys(t *testing.T) {
	_, err := WebEvidenceServiceMetadataHash(map[string]string{"": "value"})
	require.Error(t, err)

	_, err = WebEvidenceServiceMetadataHash(map[string]string{
		WebEvidenceMetadataVersion: "1",
	})
	require.Error(t, err)
}

func TestWebEvidenceContextAllowsEmptyServiceMetadataHash(t *testing.T) {
	issuedAt := time.Unix(1710000000, 0).UTC()
	expiresAt := issuedAt.Add(time.Hour)
	issuer := AttestationIssuer{
		ID:             "did:virtengine:issuer:web",
		KeyID:          "issuer-web:1",
		KeyFingerprint: hex.EncodeToString(make([]byte, 32)),
	}
	subject := NewAttestationSubject("virtengine1account")
	att := NewVerificationAttestation(
		issuer,
		subject,
		AttestationTypeEmailVerification,
		[]byte("nonce-for-web-evidence-123456789"),
		issuedAt,
		time.Hour,
		100,
		100,
	)
	digest, err := WebEvidenceAttestationDigestHex(att)
	require.NoError(t, err)

	evidence := NewWebEvidenceContext(WebEvidenceContextConfig{
		ChainID:           "ve-test-1",
		AccountAddress:    subject.AccountAddress,
		EvidenceType:      AttestationTypeEmailVerification,
		Action:            WebEvidenceActionSubmitEmail,
		ScopeID:           "email-1",
		AttestationDigest: digest,
		Issuer:            issuer,
		IssuerAlgorithm:   ProofTypeEd25519,
		Nonce:             att.Nonce,
		Challenge:         "email-challenge",
		IssuedAt:          issuedAt,
		ExpiresAt:         expiresAt,
		CallerFields: map[string]string{
			"email_hash": "hash-1",
		},
	})
	_, err = evidence.IssuerSignBytes()
	require.NoError(t, err)
	_, err = evidence.AccountAuthorizationBytes()
	require.NoError(t, err)

	metadata, err := evidence.Metadata()
	require.NoError(t, err)
	require.NotContains(t, metadata, WebEvidenceMetadataServiceMetadataHash)

	require.NoError(t, evidence.ApplyToAttestation(att))
	require.NoError(t, evidence.ValidateAttestationMetadata(att))
}

func TestWebEvidenceApplyToAttestationReplacesUnsignedMetadata(t *testing.T) {
	issuedAt := time.Unix(1710000000, 0).UTC()
	expiresAt := issuedAt.Add(time.Hour)
	issuer := AttestationIssuer{
		ID:             "did:virtengine:issuer:web",
		KeyID:          "issuer-web:1",
		KeyFingerprint: hex.EncodeToString(make([]byte, 32)),
	}
	subject := NewAttestationSubject("virtengine1account")
	att := NewVerificationAttestation(
		issuer,
		subject,
		AttestationTypeEmailVerification,
		[]byte("nonce-for-web-evidence-123456789"),
		issuedAt,
		time.Hour,
		100,
		100,
	)
	att.Metadata["unsigned.extra"] = webEvidenceTestMetadataValue
	digest, err := WebEvidenceAttestationDigestHex(att)
	require.NoError(t, err)

	evidence := NewWebEvidenceContext(WebEvidenceContextConfig{
		ChainID:             "ve-test-1",
		AccountAddress:      subject.AccountAddress,
		EvidenceType:        AttestationTypeEmailVerification,
		Action:              WebEvidenceActionSubmitEmail,
		ScopeID:             "email-1",
		AttestationDigest:   digest,
		Issuer:              issuer,
		IssuerAlgorithm:     ProofTypeEd25519,
		Nonce:               att.Nonce,
		Challenge:           "email-challenge",
		IssuedAt:            issuedAt,
		ExpiresAt:           expiresAt,
		ServiceMetadataHash: mustWebEvidenceServiceMetadataHash(t, map[string]string{"source": "unit-test"}),
		CallerFields: map[string]string{
			"email_hash": "hash-1",
		},
	})
	expected, err := evidence.Metadata()
	require.NoError(t, err)

	require.NoError(t, evidence.ApplyToAttestation(att))
	require.Equal(t, expected, att.Metadata)
	require.NotContains(t, att.Metadata, "unsigned.extra")
	require.NoError(t, evidence.ValidateAttestationMetadata(att))
}

func TestWebEvidenceAttestationDigestExcludesMetadataAndProof(t *testing.T) {
	issuedAt := time.Unix(1710000000, 0).UTC()
	issuer := AttestationIssuer{
		ID:             "did:virtengine:issuer:web",
		KeyID:          "issuer-web:1",
		KeyFingerprint: hex.EncodeToString(make([]byte, 32)),
	}
	att := NewVerificationAttestation(
		issuer,
		NewAttestationSubject("virtengine1account"),
		AttestationTypeSMSVerification,
		[]byte("nonce-for-web-evidence-123456789"),
		issuedAt,
		time.Hour,
		100,
		100,
	)

	before, err := WebEvidenceAttestationDigestHex(att)
	require.NoError(t, err)

	att.Metadata["any"] = "metadata"
	att.SetProof(NewAttestationProof(ProofTypeEd25519, issuedAt, "issuer-web:1", []byte("signature"), att.Nonce))

	after, err := WebEvidenceAttestationDigestHex(att)
	require.NoError(t, err)
	require.Equal(t, before, after)
}

func mustWebEvidenceServiceMetadataHash(t *testing.T, metadata map[string]string) string {
	t.Helper()
	hash, err := WebEvidenceServiceMetadataHash(metadata)
	require.NoError(t, err)
	return hash
}
