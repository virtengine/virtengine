package types

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func credentialDecisionTestValue() CredentialDecisionEnvelopeV1 {
	return CredentialDecisionEnvelopeV1{
		ChainID:                       "ve-test-1",
		Subject:                       "virtengine1subject",
		SubjectKeyEpoch:               7,
		AcceptedEvidenceDigests:       []string{strings.Repeat("1", 64), strings.Repeat("2", 64)},
		AcceptedReceiptDigests:        []string{strings.Repeat("3", 64)},
		PolicyEpoch:                   11,
		ScoreEpoch:                    13,
		ConsentPurposeReferenceDigest: strings.Repeat("4", 64),
		ExpiresAtUnix:                 1_800_000_000,
		Status:                        CredentialDecisionStatusEligible,
	}
}

func TestCredentialDecisionEnvelopeGolden(t *testing.T) {
	envelope := credentialDecisionTestValue()
	signBytes, err := envelope.SignBytes()
	require.NoError(t, err)

	const golden = `{"sign_domain":"VEID_CREDENTIAL_DECISION_SIGN_V1","domain":"VEID_CREDENTIAL_DECISION_ENVELOPE_V1","version":1,"chain_id":"ve-test-1","subject":"virtengine1subject","subject_key_epoch":7,"accepted_evidence_digests":["1111111111111111111111111111111111111111111111111111111111111111","2222222222222222222222222222222222222222222222222222222222222222"],"accepted_receipt_digests":["3333333333333333333333333333333333333333333333333333333333333333"],"policy_epoch":11,"score_epoch":13,"consent_purpose_reference_digest":"4444444444444444444444444444444444444444444444444444444444444444","expires_at_unix":1800000000,"status":"eligible"}`
	require.Equal(t, golden, string(signBytes))
	digest, err := envelope.DigestHex()
	require.NoError(t, err)
	require.Equal(t, "7382399e4a2adfd5548521cccc1efe69f44dd313c033e793fb6708c48fe2f3ee", digest)
}

func TestCredentialDecisionEnvelopeDeterministicDefensiveCopies(t *testing.T) {
	input := credentialDecisionTestValue()
	envelope, err := NewCredentialDecisionEnvelopeV1(input)
	require.NoError(t, err)
	input.AcceptedEvidenceDigests[0] = strings.Repeat("a", 64)
	require.Equal(t, strings.Repeat("1", 64), envelope.AcceptedEvidenceDigests[0])

	first, err := envelope.SignBytes()
	require.NoError(t, err)
	second, err := envelope.SignBytes()
	require.NoError(t, err)
	require.Equal(t, first, second)
	first[0] ^= 0xff
	third, err := envelope.SignBytes()
	require.NoError(t, err)
	require.Equal(t, second, third)

	digest, err := envelope.Digest()
	require.NoError(t, err)
	original := append([]byte(nil), digest...)
	digest[0] ^= 0xff
	again, err := envelope.Digest()
	require.NoError(t, err)
	require.Equal(t, original, again)
}

func TestCredentialDecisionEnvelopeValidation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CredentialDecisionEnvelopeV1)
	}{
		{"chain empty", func(e *CredentialDecisionEnvelopeV1) { e.ChainID = "" }},
		{"chain unicode", func(e *CredentialDecisionEnvelopeV1) { e.ChainID = "ve-\u00e9" }},
		{"subject untrimmed", func(e *CredentialDecisionEnvelopeV1) { e.Subject = " subject" }},
		{"subject epoch", func(e *CredentialDecisionEnvelopeV1) { e.SubjectKeyEpoch = 0 }},
		{"policy epoch", func(e *CredentialDecisionEnvelopeV1) { e.PolicyEpoch = 0 }},
		{"score epoch", func(e *CredentialDecisionEnvelopeV1) { e.ScoreEpoch = 0 }},
		{"evidence missing", func(e *CredentialDecisionEnvelopeV1) { e.AcceptedEvidenceDigests = nil }},
		{"evidence unsorted", func(e *CredentialDecisionEnvelopeV1) {
			e.AcceptedEvidenceDigests[0], e.AcceptedEvidenceDigests[1] = e.AcceptedEvidenceDigests[1], e.AcceptedEvidenceDigests[0]
		}},
		{"evidence duplicate", func(e *CredentialDecisionEnvelopeV1) { e.AcceptedEvidenceDigests[1] = e.AcceptedEvidenceDigests[0] }},
		{"evidence uppercase", func(e *CredentialDecisionEnvelopeV1) { e.AcceptedEvidenceDigests[0] = strings.Repeat("A", 64) }},
		{"evidence malformed", func(e *CredentialDecisionEnvelopeV1) { e.AcceptedEvidenceDigests[0] = strings.Repeat("z", 64) }},
		{"evidence oversized", func(e *CredentialDecisionEnvelopeV1) {
			e.AcceptedEvidenceDigests = makeDecisionDigests(CredentialDecisionMaxDigests + 1)
		}},
		{"eligible receipt missing", func(e *CredentialDecisionEnvelopeV1) { e.AcceptedReceiptDigests = nil }},
		{"receipt duplicate", func(e *CredentialDecisionEnvelopeV1) {
			e.AcceptedReceiptDigests = []string{strings.Repeat("3", 64), strings.Repeat("3", 64)}
		}},
		{"consent missing", func(e *CredentialDecisionEnvelopeV1) { e.ConsentPurposeReferenceDigest = "" }},
		{"expiry", func(e *CredentialDecisionEnvelopeV1) { e.ExpiresAtUnix = 0 }},
		{"status", func(e *CredentialDecisionEnvelopeV1) { e.Status = "unknown" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			envelope := credentialDecisionTestValue()
			test.mutate(&envelope)
			require.Error(t, envelope.Validate())
		})
	}

	ineligible := credentialDecisionTestValue()
	ineligible.Status = CredentialDecisionStatusIneligible
	ineligible.AcceptedReceiptDigests = []string{}
	require.NoError(t, ineligible.Validate())
	ineligible.AcceptedReceiptDigests = nil
	require.Error(t, ineligible.Validate())
}

func TestCredentialDecisionEnvelopeValidateAt(t *testing.T) {
	envelope := credentialDecisionTestValue()
	require.Error(t, envelope.ValidateAt(0))
	require.NoError(t, envelope.ValidateAt(envelope.ExpiresAtUnix-1))
	require.ErrorIs(t, envelope.ValidateAt(envelope.ExpiresAtUnix), ErrCredentialExpired)
	require.ErrorIs(t, envelope.ValidateAt(envelope.ExpiresAtUnix+1), ErrCredentialExpired)
	require.NoError(t, envelope.ValidateForChainAt(envelope.ChainID, envelope.ExpiresAtUnix-1))
	require.Error(t, envelope.ValidateForChainAt("ve-test-2", envelope.ExpiresAtUnix-1))
}

func TestCredentialDecisionEnvelopeCanonicalEmptyReceiptArray(t *testing.T) {
	for _, status := range []CredentialDecisionStatus{CredentialDecisionStatusIneligible, CredentialDecisionStatusIndeterminate} {
		envelope := credentialDecisionTestValue()
		envelope.Status = status
		envelope.AcceptedReceiptDigests = []string{}
		signBytes, err := envelope.SignBytes()
		require.NoError(t, err)
		require.Contains(t, string(signBytes), `"accepted_receipt_digests":[]`)
		envelope.AcceptedReceiptDigests = nil
		require.Error(t, envelope.Validate())
	}
}

func TestCredentialDecisionEnvelopeEveryFieldTamper(t *testing.T) {
	base := credentialDecisionTestValue()
	baseSign, err := base.SignBytes()
	require.NoError(t, err)
	baseDigest, err := base.Digest()
	require.NoError(t, err)

	tests := []struct {
		name   string
		mutate func(*CredentialDecisionEnvelopeV1)
	}{
		{"chain", func(e *CredentialDecisionEnvelopeV1) { e.ChainID = "ve-test-2" }},
		{"subject", func(e *CredentialDecisionEnvelopeV1) { e.Subject += "2" }},
		{"subject epoch", func(e *CredentialDecisionEnvelopeV1) { e.SubjectKeyEpoch++ }},
		{"evidence", func(e *CredentialDecisionEnvelopeV1) { e.AcceptedEvidenceDigests[0] = strings.Repeat("0", 64) }},
		{"receipt", func(e *CredentialDecisionEnvelopeV1) { e.AcceptedReceiptDigests[0] = strings.Repeat("5", 64) }},
		{"policy epoch", func(e *CredentialDecisionEnvelopeV1) { e.PolicyEpoch++ }},
		{"score epoch", func(e *CredentialDecisionEnvelopeV1) { e.ScoreEpoch++ }},
		{"consent", func(e *CredentialDecisionEnvelopeV1) { e.ConsentPurposeReferenceDigest = strings.Repeat("6", 64) }},
		{"expiry", func(e *CredentialDecisionEnvelopeV1) { e.ExpiresAtUnix++ }},
		{"status", func(e *CredentialDecisionEnvelopeV1) { e.Status = CredentialDecisionStatusIneligible }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := credentialDecisionTestValue()
			test.mutate(&mutated)
			require.NoError(t, mutated.Validate())
			signBytes, err := mutated.SignBytes()
			require.NoError(t, err)
			digest, err := mutated.Digest()
			require.NoError(t, err)
			require.NotEqual(t, baseSign, signBytes)
			require.NotEqual(t, baseDigest, digest)
		})
	}
}

func TestCredentialDecisionEnvelopeContainsNoClaimsOrKeyMaterial(t *testing.T) {
	envelope := credentialDecisionTestValue()
	bz, err := json.Marshal(envelope)
	require.NoError(t, err)
	signBytes, err := envelope.SignBytes()
	require.NoError(t, err)
	serialized := strings.ToLower(string(append(bz, signBytes...)))
	for _, prohibited := range []string{"age", "date_of_birth", "residency", "country", "claims", "plaintext", "private_key", "public_key", "signature", "proof"} {
		require.NotContains(t, serialized, prohibited)
	}

	typeOfEnvelope := reflect.TypeOf(CredentialDecisionEnvelopeV1{})
	actualTags := make([]string, 0, typeOfEnvelope.NumField())
	for index := 0; index < typeOfEnvelope.NumField(); index++ {
		actualTags = append(actualTags, strings.Split(typeOfEnvelope.Field(index).Tag.Get("json"), ",")[0])
	}
	sort.Strings(actualTags)
	expectedTags := []string{
		"accepted_evidence_digests", "accepted_receipt_digests", "chain_id",
		"consent_purpose_reference_digest", "expires_at_unix", "policy_epoch",
		"score_epoch", "status", "subject", "subject_key_epoch",
	}
	sort.Strings(expectedTags)
	require.Equal(t, expectedTags, actualTags)
}

func makeDecisionDigests(count int) []string {
	values := make([]string, count)
	for index := range values {
		values[index] = strings.Repeat(string(rune('a'+index%6)), 63) + string(rune('0'+index%10))
	}
	return values
}
