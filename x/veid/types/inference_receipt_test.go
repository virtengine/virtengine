package types

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
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
	require.Equal(t, "bdc6e5985ddb3774ed0b82bbb988dedaa7ecd48b643805c398324688505811ab", receiptDigestHexForTest(t, receipt))

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
		"domain":                    func(r *InferenceReceipt) { r.Domain = "wrong" },
		"version":                   func(r *InferenceReceipt) { r.Version++ },
		"chain":                     func(r *InferenceReceipt) { r.ChainID = "other-chain" },
		"account":                   func(r *InferenceReceipt) { r.AccountAddress = "other-account" },
		"request":                   func(r *InferenceReceipt) { r.RequestID = "other-request" },
		"scope_order":               func(r *InferenceReceipt) { r.ScopeIDs = []string{"scope-b", "scope-a"} },
		"nonce":                     func(r *InferenceReceipt) { r.Nonce = "other-nonce" },
		"input":                     func(r *InferenceReceipt) { r.InputDigest[0] ^= 0x01 },
		"feature":                   func(r *InferenceReceipt) { r.FeatureDigest[0] ^= 0x01 },
		"schema":                    func(r *InferenceReceipt) { r.SchemaDigest[0] ^= 0x01 },
		"lineage":                   func(r *InferenceReceipt) { r.EvidenceLineageDigest[0] ^= 0x01 },
		"pipeline_version":          func(r *InferenceReceipt) { r.PipelineVersion = "v2.0.0" },
		"manifest":                  func(r *InferenceReceipt) { r.ModelManifestDigest[0] ^= 0x01 },
		"model":                     func(r *InferenceReceipt) { r.ModelDigest[0] ^= 0x01 },
		"runtime_image":             func(r *InferenceReceipt) { r.RuntimeImageDigest[0] ^= 0x01 },
		"runtime":                   func(r *InferenceReceipt) { r.RuntimeDigest[0] ^= 0x01 },
		"config":                    func(r *InferenceReceipt) { r.ConfigDigest[0] ^= 0x01 },
		"profile_force_cpu":         func(r *InferenceReceipt) { r.DeterminismProfile.ForceCPU = false },
		"profile_random_seed":       func(r *InferenceReceipt) { r.DeterminismProfile.RandomSeed++ },
		"profile_deterministic_ops": func(r *InferenceReceipt) { r.DeterminismProfile.DeterministicOps = false },
		"profile_inter_op_threads":  func(r *InferenceReceipt) { r.DeterminismProfile.InterOpThreads++ },
		"profile_intra_op_threads":  func(r *InferenceReceipt) { r.DeterminismProfile.IntraOpThreads++ },
		"profile_disable_gpu":       func(r *InferenceReceipt) { r.DeterminismProfile.DisableGPU = false },
		"score":                     func(r *InferenceReceipt) { r.Score = MaxScore + 1 },
		"status":                    func(r *InferenceReceipt) { r.Status = VerificationResultStatus("unknown") },
		"confidence":                func(r *InferenceReceipt) { r.ConfidenceMillionths = InferenceReceiptMaxConfidencePPM + 1 },
		"reasons":                   func(r *InferenceReceipt) { r.ReasonCodes = []ReasonCode{ReasonCodeSuccess, ReasonCodeSuccess} },
		"unknown_reason":            func(r *InferenceReceipt) { r.ReasonCodes = []ReasonCode{"NOT_CANONICAL"} },
		"failed_success": func(r *InferenceReceipt) {
			r.Status = VerificationResultStatusFailed
			r.Score = 0
			r.ReasonCodes = []ReasonCode{ReasonCodeSuccess}
		},
		"issued_height": func(r *InferenceReceipt) { r.IssuedHeight = 0 },
		"expiry_height": func(r *InferenceReceipt) { r.ExpiresHeight = r.IssuedHeight },
		"issued_time":   func(r *InferenceReceipt) { r.IssuedAt = time.Time{} },
		"expires_at":    func(r *InferenceReceipt) { r.ExpiresAt = r.IssuedAt },
		"issued_precision": func(r *InferenceReceipt) {
			r.IssuedAt = r.IssuedAt.Add(time.Nanosecond)
		},
		"expires_precision": func(r *InferenceReceipt) {
			r.ExpiresAt = r.ExpiresAt.Add(time.Nanosecond)
		},
		"signer_key":      func(r *InferenceReceipt) { r.SignerKeyID = "other-key" },
		"fingerprint":     func(r *InferenceReceipt) { r.SignerFingerprint = "not-hex" },
		"signer_sequence": func(r *InferenceReceipt) { r.SignerSequence = 0 },
	}
	for name, tamper := range tamperCases {
		t.Run(name, func(t *testing.T) {
			receipt := cloneInferenceReceipt(base)
			tamper(&receipt)
			require.Error(t, receipt.VerifySignature(pub))
		})
	}
}

func TestInferenceReceiptValidSubstitutionsInvalidateSignature(t *testing.T) {
	pub, priv := deterministicReceiptKey(t)
	base := testInferenceReceipt(t, pub)
	require.NoError(t, base.Sign(priv))

	tests := map[string]func(*InferenceReceipt){
		"scope":      func(r *InferenceReceipt) { r.ScopeIDs = []string{"scope-a", "scope-c"} },
		"score":      func(r *InferenceReceipt) { r.Score = 90 },
		"confidence": func(r *InferenceReceipt) { r.ConfidenceMillionths = 900_000 },
		"status_and_reasons": func(r *InferenceReceipt) {
			r.Status = VerificationResultStatusPartial
			r.ReasonCodes = []ReasonCode{ReasonCodeLowConfidence}
		},
		"issued_height":      func(r *InferenceReceipt) { r.IssuedHeight++ },
		"expires_height":     func(r *InferenceReceipt) { r.ExpiresHeight-- },
		"issued_time":        func(r *InferenceReceipt) { r.IssuedAt = r.IssuedAt.Add(time.Second) },
		"expires_time":       func(r *InferenceReceipt) { r.ExpiresAt = r.ExpiresAt.Add(-time.Second) },
		"signer_fingerprint": func(r *InferenceReceipt) { r.SignerFingerprint = strings.Repeat("a", 64) },
		"signer_sequence":    func(r *InferenceReceipt) { r.SignerSequence++ },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			receipt := cloneInferenceReceipt(base)
			mutate(&receipt)
			require.NoError(t, receipt.Validate())
			require.ErrorContains(t, receipt.VerifySignature(pub), "invalid inference receipt signature")
		})
	}
}

func TestInferenceReceiptRejectsNonCanonicalFingerprintCase(t *testing.T) {
	pub, priv := deterministicReceiptKey(t)
	receipt := testInferenceReceipt(t, pub)
	require.NoError(t, receipt.Sign(priv))
	receipt.SignerFingerprint = strings.ToUpper(receipt.SignerFingerprint)

	require.ErrorContains(t, receipt.Validate(), "lowercase")
	require.ErrorContains(t, receipt.VerifySignature(pub), "lowercase")
}

func TestInferenceReceiptRejectsSubsecondTamperBeforeSignatureVerification(t *testing.T) {
	pub, priv := deterministicReceiptKey(t)
	receipt := testInferenceReceipt(t, pub)
	require.NoError(t, receipt.Sign(priv))
	receipt.ExpiresAt = receipt.ExpiresAt.Add(time.Nanosecond)
	receipt.Signature[0] ^= 0xff

	require.ErrorContains(t, receipt.VerifySignature(pub), "second-aligned")
}

func TestInferenceReceiptLifetimeBoundaries(t *testing.T) {
	pub, priv := deterministicReceiptKey(t)
	base := testInferenceReceipt(t, pub)
	base.ExpiresAt = base.IssuedAt.Add(InferenceReceiptMaxLifetime)
	base.ExpiresHeight = base.IssuedHeight + InferenceReceiptMaxHeightLifetime
	require.NoError(t, base.Sign(priv))
	require.NoError(t, base.Validate())

	overTime := cloneInferenceReceipt(base)
	overTime.ExpiresAt = overTime.ExpiresAt.Add(time.Second)
	require.ErrorContains(t, overTime.Validate(), "lifetime exceeds")

	overHeight := cloneInferenceReceipt(base)
	overHeight.ExpiresHeight++
	require.ErrorContains(t, overHeight.Validate(), "height lifetime exceeds")
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
