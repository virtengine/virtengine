package types

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInferenceReceiptCanonicalSignDigestAndVerify(t *testing.T) {
	pub, priv := deterministicReceiptKey(t)
	receipt := testInferenceReceipt(t, pub)

	signBytes, err := receipt.SignBytes()
	require.NoError(t, err)
	require.NoError(t, receipt.Sign(priv))
	require.NoError(t, receipt.VerifySignature(pub))

	signBytesAgain, err := receipt.SignBytes()
	require.NoError(t, err)
	require.True(t, bytes.Equal(signBytes, signBytesAgain))

	digest, err := receipt.Digest()
	require.NoError(t, err)
	require.Len(t, digest, sha256.Size)
	require.Equal(t, "328a6979c5953f439be0f75248e67f170564bd2fb83f4b21223435ed00e84a5e", receiptDigestHexForTest(t, receipt))

	contextDigest, err := receipt.ContextDigest()
	require.NoError(t, err)
	require.Len(t, contextDigest, sha256.Size)
	require.NotEqual(t, digest, contextDigest)

	changedOutput := cloneInferenceReceipt(receipt)
	changedOutput.Score = 88
	require.NoError(t, changedOutput.Sign(priv))
	changedOutputDigest, err := changedOutput.Digest()
	require.NoError(t, err)
	changedOutputContextDigest, err := changedOutput.ContextDigest()
	require.NoError(t, err)
	require.NotEqual(t, digest, changedOutputDigest)
	require.Equal(t, contextDigest, changedOutputContextDigest)

	receipt.Signature[0] ^= 0xff
	require.Error(t, receipt.VerifySignature(pub))
}

func TestInferenceReceiptRejectsTamperedFields(t *testing.T) {
	pub, priv := deterministicReceiptKey(t)
	base := testInferenceReceipt(t, pub)
	require.NoError(t, base.Sign(priv))

	tamperCases := map[string]func(*InferenceReceipt){
		"domain":         func(r *InferenceReceipt) { r.Domain = "wrong" },
		"version":        func(r *InferenceReceipt) { r.Version++ },
		"chain":          func(r *InferenceReceipt) { r.ChainID = "other-chain" },
		"account":        func(r *InferenceReceipt) { r.AccountAddress = "other-account" },
		"request":        func(r *InferenceReceipt) { r.RequestID = "other-request" },
		"nonce":          func(r *InferenceReceipt) { r.Nonce = "other-nonce" },
		"scope_order":    func(r *InferenceReceipt) { r.ScopeIDs = []string{"scope-b", "scope-a"} },
		"scope_missing":  func(r *InferenceReceipt) { r.ScopeIDs = []string{"scope-a"} },
		"scope_extra":    func(r *InferenceReceipt) { r.ScopeIDs = []string{"scope-a", "scope-b", "scope-c"} },
		"input":          func(r *InferenceReceipt) { r.InputDigest[0] ^= 0x01 },
		"feature":        func(r *InferenceReceipt) { r.FeatureDigest[0] ^= 0x01 },
		"schema":         func(r *InferenceReceipt) { r.SchemaDigest[0] ^= 0x01 },
		"lineage":        func(r *InferenceReceipt) { r.EvidenceLineageDigest[0] ^= 0x01 },
		"pipeline":       func(r *InferenceReceipt) { r.PipelineVersion = "v2.0.0" },
		"manifest":       func(r *InferenceReceipt) { r.ModelManifestDigest[0] ^= 0x01 },
		"model":          func(r *InferenceReceipt) { r.ModelDigest[0] ^= 0x01 },
		"runtime_image":  func(r *InferenceReceipt) { r.RuntimeImageDigest[0] ^= 0x01 },
		"runtime":        func(r *InferenceReceipt) { r.RuntimeDigest[0] ^= 0x01 },
		"config":         func(r *InferenceReceipt) { r.ConfigDigest[0] ^= 0x01 },
		"profile":        func(r *InferenceReceipt) { r.DeterminismProfile.ForceCPU = false },
		"score":          func(r *InferenceReceipt) { r.Score = MaxScore + 1 },
		"status":         func(r *InferenceReceipt) { r.Status = VerificationResultStatus("unknown") },
		"confidence":     func(r *InferenceReceipt) { r.ConfidenceMillionths = InferenceReceiptMaxConfidencePPM + 1 },
		"reasons":        func(r *InferenceReceipt) { r.ReasonCodes = []ReasonCode{ReasonCodeSuccess, ReasonCodeSuccess} },
		"unknown_reason": func(r *InferenceReceipt) { r.ReasonCodes = []ReasonCode{"NOT_CANONICAL"} },
		"failed_success": func(r *InferenceReceipt) {
			r.Status = VerificationResultStatusFailed
			r.Score = 0
			r.ReasonCodes = []ReasonCode{ReasonCodeSuccess}
		},
		"issued_height": func(r *InferenceReceipt) { r.IssuedHeight = 0 },
		"expiry_height": func(r *InferenceReceipt) { r.ExpiresHeight = r.IssuedHeight },
		"issued_time":   func(r *InferenceReceipt) { r.IssuedAt = time.Time{} },
		"expires_time":  func(r *InferenceReceipt) { r.ExpiresAt = r.IssuedAt },
		"signer_key":    func(r *InferenceReceipt) { r.SignerKeyID = "other-key" },
		"fingerprint":   func(r *InferenceReceipt) { r.SignerFingerprint = "not-hex" },
		"fingerprint_case": func(r *InferenceReceipt) {
			r.SignerFingerprint = strings.ToUpper(r.SignerFingerprint)
		},
		"signer_sequence": func(r *InferenceReceipt) { r.SignerSequence = 0 },
		"signature":       func(r *InferenceReceipt) { r.Signature[0] ^= 0xff },
	}
	for name, tamper := range tamperCases {
		t.Run(name, func(t *testing.T) {
			receipt := cloneInferenceReceipt(base)
			tamper(&receipt)
			require.Error(t, receipt.VerifySignature(pub))
		})
	}
}

func TestCanonicalInferenceDeterminismProfile(t *testing.T) {
	profile := CanonicalInferenceDeterminismProfile()
	require.True(t, profile.IsCanonical())
	require.True(t, profile.ForceCPU)
	require.Equal(t, int64(42), profile.RandomSeed)
	require.True(t, profile.DeterministicOps)
	require.True(t, profile.DisableGPU)
	require.NotEqual(t, profile.Digest(), CanonicalInferenceDeterminismConfigDigest())
	require.Equal(t,
		InferencePipelineDeterminismConfigDigest(StrictInferencePipelineDeterminismConfig()),
		CanonicalInferenceDeterminismConfigDigest(),
	)

	profile.RandomSeed = 7
	require.False(t, profile.IsCanonical())
}

func TestInferenceReceiptCanonicalSignedBytesRoundTripAndDefensiveCopy(t *testing.T) {
	pub, priv := deterministicReceiptKey(t)
	receipt := testInferenceReceipt(t, pub)
	require.NoError(t, receipt.Sign(priv))

	encoded, err := receipt.CanonicalSignedBytes()
	require.NoError(t, err)
	require.NotEmpty(t, encoded)
	require.LessOrEqual(t, len(encoded), InferenceReceiptMaxSignedBytes)
	require.NotEqual(t, byte('{'), encoded[0])
	encodedHash := sha256.Sum256(encoded)
	require.Equal(t, "214ec4c1be6bf4f70270521bac15a03948f86fe71f2d85aaf60d883ac31072b0", hex.EncodeToString(encodedHash[:]))

	decoded, err := DecodeCanonicalSignedInferenceReceipt(encoded)
	require.NoError(t, err)
	require.NoError(t, decoded.VerifySignature(pub))
	require.Equal(t, receiptDigestHexForTest(t, receipt), receiptDigestHexForTest(t, decoded))

	encodedAgain, err := decoded.CanonicalSignedBytes()
	require.NoError(t, err)
	require.Equal(t, encoded, encodedAgain)

	decoded.Signature[0] ^= 0xff
	decoded.InputDigest[0] ^= 0xff
	redecoded, err := DecodeCanonicalSignedInferenceReceipt(encoded)
	require.NoError(t, err)
	require.NoError(t, redecoded.VerifySignature(pub))
	require.NotEqual(t, decoded.Signature, redecoded.Signature)
	require.NotEqual(t, decoded.InputDigest, redecoded.InputDigest)
}

func TestDecodeCanonicalSignedInferenceReceiptRejectsMalformedNonCanonicalAndInvalid(t *testing.T) {
	pub, priv := deterministicReceiptKey(t)
	receipt := testInferenceReceipt(t, pub)
	require.NoError(t, receipt.Sign(priv))
	encoded, err := receipt.CanonicalSignedBytes()
	require.NoError(t, err)

	invalidDomain := append([]byte(nil), encoded...)
	invalidDomain[8] ^= 0x20
	oversizedLength := append([]byte(nil), encoded...)
	binary.BigEndian.PutUint64(oversizedLength[:8], uint64(InferenceReceiptMaxString+1))
	truncatedDomain := make([]byte, 8)
	binary.BigEndian.PutUint64(truncatedDomain, 10)
	missingSignature := append([]byte(nil), encoded[:len(encoded)-ed25519.SignatureSize]...)
	invalidBool := append([]byte(nil), encoded...)
	profileOffset := bytes.Index(invalidBool, encodedInferenceReceiptProfileForTest())
	require.NotEqual(t, -1, profileOffset)
	invalidBool[profileOffset] = 2
	invalidUTF8 := append([]byte(nil), encoded...)
	accountOffset := bytes.Index(invalidUTF8, []byte("virt1account"))
	require.NotEqual(t, -1, accountOffset)
	invalidUTF8[accountOffset] = 0xff
	oversizedScopeCount := append([]byte(nil), encoded...)
	firstScopeOffset := bytes.Index(oversizedScopeCount, []byte("scope-a"))
	require.GreaterOrEqual(t, firstScopeOffset, 16)
	binary.BigEndian.PutUint64(oversizedScopeCount[firstScopeOffset-16:firstScopeOffset-8], uint64(InferenceReceiptMaxScopes+1))
	oversizedReasonCount := append([]byte(nil), encoded...)
	firstReasonOffset := bytes.Index(oversizedReasonCount, []byte("SUCCESS"))
	require.GreaterOrEqual(t, firstReasonOffset, 16)
	binary.BigEndian.PutUint64(oversizedReasonCount[firstReasonOffset-16:firstReasonOffset-8], uint64(InferenceReceiptMaxReasonCodes+1))
	cases := map[string][]byte{
		"empty":                  nil,
		"json_payload":           []byte(`{"domain":"VEID_INFERENCE_RECEIPT"}`),
		"oversized":              bytes.Repeat([]byte{0x42}, InferenceReceiptMaxSignedBytes+1),
		"truncated_domain":       truncatedDomain,
		"truncated":              encoded[:len(encoded)-1],
		"truncated_length":       encoded[:7],
		"truncated_version":      encoded[:len(encodedInferenceReceiptDomainForTest())+3],
		"truncated_scope":        encoded[:firstScopeOffset+3],
		"truncated_reason":       encoded[:firstReasonOffset+3],
		"trailing_data":          append(append([]byte(nil), encoded...), 0),
		"invalid_domain":         invalidDomain,
		"oversized_length":       oversizedLength,
		"oversized_scope_count":  oversizedScopeCount,
		"oversized_reason_count": oversizedReasonCount,
		"missing_signature":      missingSignature,
		"invalid_bool":           invalidBool,
		"invalid_utf8":           invalidUTF8,
	}

	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := DecodeCanonicalSignedInferenceReceipt(input)
			require.Error(t, err)
		})
	}
}

func TestInferenceReceiptRejectsSubsecondTimestamps(t *testing.T) {
	pub, priv := deterministicReceiptKey(t)
	receipt := testInferenceReceipt(t, pub)
	receipt.IssuedAt = receipt.IssuedAt.Add(time.Nanosecond)
	require.ErrorContains(t, receipt.Sign(priv), "whole UTC seconds")

	receipt = testInferenceReceipt(t, pub)
	receipt.ExpiresAt = receipt.ExpiresAt.Add(time.Nanosecond)
	require.ErrorContains(t, receipt.Sign(priv), "whole UTC seconds")
}

func TestInferenceReceiptCanonicalEncoderRejectsInvalidUTF8(t *testing.T) {
	pub, priv := deterministicReceiptKey(t)
	testCases := map[string]func(*InferenceReceipt){
		"top_level": func(receipt *InferenceReceipt) {
			receipt.Nonce = string([]byte{0xff})
		},
		"scope": func(receipt *InferenceReceipt) {
			receipt.ScopeIDs = []string{string([]byte{0xff})}
		},
		"reason": func(receipt *InferenceReceipt) {
			receipt.ReasonCodes = []ReasonCode{ReasonCode(string([]byte{0xff}))}
		},
	}

	for name, mutate := range testCases {
		t.Run(name, func(t *testing.T) {
			receipt := testInferenceReceipt(t, pub)
			mutate(&receipt)
			require.ErrorContains(t, receipt.Sign(priv), "UTF-8")
		})
	}
}

func TestInferenceReceiptCanonicalBytesTamperInvalidatesSignature(t *testing.T) {
	pub, priv := deterministicReceiptKey(t)
	receipt := testInferenceReceipt(t, pub)
	require.NoError(t, receipt.Sign(priv))
	encoded, err := receipt.CanonicalSignedBytes()
	require.NoError(t, err)

	tampered := append([]byte(nil), encoded...)
	tampered[len(tampered)-1] ^= 0x01
	decoded, err := DecodeCanonicalSignedInferenceReceipt(tampered)
	require.NoError(t, err)
	require.Error(t, decoded.VerifySignature(pub))
}

func encodedInferenceReceiptProfileForTest() []byte {
	enc := &inferenceReceiptBinaryEncoder{}
	writeInferenceDeterminismProfile(enc, CanonicalInferenceDeterminismProfile())
	return enc.bytes()
}

func encodedInferenceReceiptDomainForTest() []byte {
	enc := &inferenceReceiptBinaryEncoder{}
	enc.writeString(InferenceReceiptDomain)
	return enc.bytes()
}

func deterministicReceiptKey(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	seed := sha256.Sum256([]byte("virtengine/inference/receipt/test-key/v1"))
	pub, priv, err := ed25519.GenerateKey(bytes.NewReader(seed[:]))
	require.NoError(t, err)
	return pub, priv
}

func testInferenceReceipt(t *testing.T, pub ed25519.PublicKey) InferenceReceipt {
	t.Helper()
	now := time.Unix(1_700_000_000, 0).UTC()
	return InferenceReceipt{
		Domain:                InferenceReceiptDomain,
		Version:               InferenceReceiptVersion,
		ChainID:               "chain-A",
		AccountAddress:        "virt1account",
		RequestID:             "request-1",
		ScopeIDs:              []string{"scope-a", "scope-b"},
		Nonce:                 "nonce-1",
		InputDigest:           testReceiptDigest(0x01),
		FeatureDigest:         testReceiptDigest(0x02),
		SchemaDigest:          testReceiptDigest(0x03),
		EvidenceLineageDigest: testReceiptDigest(0x04),
		PipelineVersion:       "v1.0.0",
		ModelManifestDigest:   testReceiptDigest(0x05),
		ModelDigest:           testReceiptDigest(0x06),
		RuntimeImageDigest:    testReceiptDigest(0x07),
		RuntimeDigest:         testReceiptDigest(0x07),
		ConfigDigest:          CanonicalInferenceDeterminismConfigDigest(),
		DeterminismProfile:    CanonicalInferenceDeterminismProfile(),
		Score:                 91,
		Status:                VerificationResultStatusSuccess,
		ConfidenceMillionths:  910_000,
		ReasonCodes:           []ReasonCode{ReasonCodeSuccess},
		IssuedHeight:          10,
		IssuedAt:              now,
		ExpiresHeight:         12,
		ExpiresAt:             now.Add(2 * time.Minute),
		SignerKeyID:           "did:virtengine:inference:1",
		SignerFingerprint:     ComputeKeyFingerprint(pub),
		SignerSequence:        1,
	}
}

func cloneInferenceReceipt(receipt InferenceReceipt) InferenceReceipt {
	clone := receipt
	clone.ScopeIDs = append([]string(nil), receipt.ScopeIDs...)
	clone.InputDigest = append([]byte(nil), receipt.InputDigest...)
	clone.FeatureDigest = append([]byte(nil), receipt.FeatureDigest...)
	clone.SchemaDigest = append([]byte(nil), receipt.SchemaDigest...)
	clone.EvidenceLineageDigest = append([]byte(nil), receipt.EvidenceLineageDigest...)
	clone.ModelManifestDigest = append([]byte(nil), receipt.ModelManifestDigest...)
	clone.ModelDigest = append([]byte(nil), receipt.ModelDigest...)
	clone.RuntimeImageDigest = append([]byte(nil), receipt.RuntimeImageDigest...)
	clone.RuntimeDigest = append([]byte(nil), receipt.RuntimeDigest...)
	clone.ConfigDigest = append([]byte(nil), receipt.ConfigDigest...)
	clone.ReasonCodes = append([]ReasonCode(nil), receipt.ReasonCodes...)
	clone.Signature = append([]byte(nil), receipt.Signature...)
	return clone
}

func testReceiptDigest(value byte) []byte {
	out := make([]byte, sha256.Size)
	for i := range out {
		out[i] = value
	}
	return out
}

func receiptDigestHexForTest(t *testing.T, receipt InferenceReceipt) string {
	t.Helper()
	digest, err := receipt.DigestHex()
	require.NoError(t, err)
	return digest
}
