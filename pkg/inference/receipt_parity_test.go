package inference_test

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"math"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/pkg/security"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

type receiptParityFixture struct {
	Schema    string   `json:"schema"`
	Version   int      `json:"version"`
	Nonclaims []string `json:"nonclaims"`
	Signing   struct {
		SeedHex         string `json:"seed_hex"`
		PublicKeyBase64 string `json:"public_key_base64"`
		FingerprintHex  string `json:"fingerprint_hex"`
	} `json:"signing"`
	Cases []receiptParityCase `json:"cases"`
}

type receiptParityCase struct {
	Name   string `json:"name"`
	Source struct {
		ChainID          string            `json:"chain_id"`
		AccountAddress   string            `json:"account_address"`
		RequestID        string            `json:"request_id"`
		ScopeIDs         []string          `json:"scope_ids"`
		Nonce            string            `json:"nonce"`
		PipelineVersion  string            `json:"pipeline_version"`
		DigestSources    map[string]string `json:"digest_sources"`
		Score            uint32            `json:"score"`
		RawOutputFloat32 json.RawMessage   `json:"raw_output_float32"`
		Status           string            `json:"status"`
		Confidence       uint32            `json:"confidence_millionths"`
		ReasonCodes      []string          `json:"reason_codes"`
		IssuedHeight     int64             `json:"issued_height"`
		IssuedAtUnix     int64             `json:"issued_at_unix"`
		ExpiresHeight    int64             `json:"expires_height"`
		ExpiresAtUnix    int64             `json:"expires_at_unix"`
		SignerKeyID      string            `json:"signer_key_id"`
		SignerSequence   uint64            `json:"signer_sequence"`
	} `json:"source"`
	Expected struct {
		CanonicalScopeIDs    []string `json:"canonical_scope_ids"`
		CanonicalReasonCodes []string `json:"canonical_reason_codes"`
		ConfigDigestHex      string   `json:"config_digest_hex"`
		SignBytesBase64      string   `json:"sign_bytes_base64"`
		DigestHex            string   `json:"digest_hex"`
		ContextDigestHex     string   `json:"context_digest_hex"`
		SignatureBase64      string   `json:"signature_base64"`
		QuantizedScore       *uint32  `json:"quantized_score"`
	} `json:"expected"`
}

func TestInferenceReceiptParityFixture(t *testing.T) {
	fixture := loadReceiptParityFixture(t)
	require.Equal(t, "virtengine.inference.receipt_parity", fixture.Schema)
	require.Equal(t, 1, fixture.Version)
	require.NotEmpty(t, fixture.Nonclaims)
	require.NotEmpty(t, fixture.Cases)

	seed := mustDecodeHex(t, fixture.Signing.SeedHex)
	require.Len(t, seed, ed25519.SeedSize)
	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	require.Equal(t, fixture.Signing.PublicKeyBase64, base64.StdEncoding.EncodeToString(publicKey))
	require.Equal(t, fixture.Signing.FingerprintHex, veidtypes.ComputeKeyFingerprint(publicKey))

	for _, testCase := range fixture.Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			receipt := receiptFromParitySource(t, testCase, fixture.Signing.FingerprintHex)
			require.Equal(t, testCase.Expected.CanonicalScopeIDs, receipt.ScopeIDs)
			require.Equal(t, testCase.Expected.CanonicalReasonCodes, reasonStrings(receipt.ReasonCodes))
			require.Equal(t, testCase.Expected.ConfigDigestHex, hex.EncodeToString(receipt.ConfigDigest))

			signBytes, err := receipt.SignBytes()
			require.NoError(t, err)
			require.Equal(t, mustDecodeBase64(t, testCase.Expected.SignBytesBase64), signBytes)

			digest, err := receipt.Digest()
			require.NoError(t, err)
			require.Equal(t, testCase.Expected.DigestHex, hex.EncodeToString(digest))
			contextDigest, err := receipt.ContextDigest()
			require.NoError(t, err)
			require.Equal(t, testCase.Expected.ContextDigestHex, hex.EncodeToString(contextDigest))

			require.NoError(t, receipt.Sign(privateKey))
			require.Equal(t, mustDecodeBase64(t, testCase.Expected.SignatureBase64), receipt.Signature)
			require.NoError(t, receipt.Validate())
			require.NoError(t, receipt.VerifySignature(publicKey))

			if len(testCase.Source.RawOutputFloat32) == 0 {
				require.Nil(t, testCase.Expected.QuantizedScore)
			} else {
				require.NotNil(t, testCase.Expected.QuantizedScore)
				rawOutput := decodeRawOutput(t, testCase.Source.RawOutputFloat32)
				require.Equal(t, *testCase.Expected.QuantizedScore,
					security.SafeFloat32ToUint32(rawOutput, 0, 100))
				require.Equal(t, receipt.Score, *testCase.Expected.QuantizedScore)
			}

			tampered := receipt
			tampered.Nonce += "-tampered"
			require.Error(t, tampered.VerifySignature(publicKey))
			receipt.Signature = bytes.Clone(receipt.Signature)
			receipt.Signature[0] ^= 0x01
			require.Error(t, receipt.VerifySignature(publicKey))
		})
	}
}

func decodeRawOutput(t *testing.T, raw json.RawMessage) float32 {
	t.Helper()
	var numeric float32
	if err := json.Unmarshal(raw, &numeric); err == nil {
		return numeric
	}
	var special string
	require.NoError(t, json.Unmarshal(raw, &special))
	switch special {
	case "nan":
		return float32(math.NaN())
	case "positive_infinity":
		return float32(math.Inf(1))
	case "negative_infinity":
		return float32(math.Inf(-1))
	default:
		t.Fatalf("unknown raw output token %q", special)
		return 0
	}
}

func loadReceiptParityFixture(t *testing.T) receiptParityFixture {
	t.Helper()
	data, err := os.ReadFile("conformance/testdata/receipt_parity_v1.json")
	require.NoError(t, err)
	var fixture receiptParityFixture
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	require.NoError(t, decoder.Decode(&fixture))
	var trailing any
	require.ErrorIs(t, decoder.Decode(&trailing), io.EOF)
	return fixture
}

func receiptFromParitySource(t *testing.T, testCase receiptParityCase, fingerprint string) veidtypes.InferenceReceipt {
	t.Helper()
	source := testCase.Source
	reasons := make([]veidtypes.ReasonCode, len(source.ReasonCodes))
	for index, reason := range source.ReasonCodes {
		reasons[index] = veidtypes.ReasonCode(reason)
	}
	return veidtypes.InferenceReceipt{
		Domain: veidtypes.InferenceReceiptDomain, Version: veidtypes.InferenceReceiptVersion,
		ChainID: source.ChainID, AccountAddress: source.AccountAddress, RequestID: source.RequestID,
		ScopeIDs: veidtypes.CanonicalInferenceReceiptScopeIDs(source.ScopeIDs), Nonce: source.Nonce,
		InputDigest:           sourceDigest(t, source.DigestSources, "input_digest"),
		FeatureDigest:         sourceDigest(t, source.DigestSources, "feature_digest"),
		SchemaDigest:          sourceDigest(t, source.DigestSources, "schema_digest"),
		EvidenceLineageDigest: sourceDigest(t, source.DigestSources, "evidence_lineage_digest"),
		PipelineVersion:       source.PipelineVersion,
		ModelManifestDigest:   sourceDigest(t, source.DigestSources, "model_manifest_digest"),
		ModelDigest:           sourceDigest(t, source.DigestSources, "model_digest"),
		RuntimeImageDigest:    sourceDigest(t, source.DigestSources, "runtime_image_digest"),
		RuntimeDigest:         sourceDigest(t, source.DigestSources, "runtime_digest"),
		ConfigDigest:          veidtypes.CanonicalInferenceDeterminismConfigDigest(),
		DeterminismProfile:    veidtypes.CanonicalInferenceDeterminismProfile(),
		Score:                 source.Score, Status: veidtypes.VerificationResultStatus(source.Status),
		ConfidenceMillionths: source.Confidence,
		ReasonCodes:          veidtypes.CanonicalInferenceReceiptReasonCodes(reasons),
		IssuedHeight:         source.IssuedHeight, IssuedAt: time.Unix(source.IssuedAtUnix, 0).UTC(),
		ExpiresHeight: source.ExpiresHeight, ExpiresAt: time.Unix(source.ExpiresAtUnix, 0).UTC(),
		SignerKeyID: source.SignerKeyID, SignerFingerprint: fingerprint, SignerSequence: source.SignerSequence,
	}
}

func sourceDigest(t *testing.T, sources map[string]string, name string) []byte {
	t.Helper()
	value, ok := sources[name]
	require.True(t, ok, "missing digest source %s", name)
	require.NotEmpty(t, value)
	digest := sha256.Sum256([]byte(value))
	return digest[:]
}

func reasonStrings(reasons []veidtypes.ReasonCode) []string {
	result := make([]string, len(reasons))
	for index, reason := range reasons {
		result[index] = string(reason)
	}
	return result
}

func mustDecodeHex(t *testing.T, value string) []byte {
	t.Helper()
	decoded, err := hex.DecodeString(value)
	require.NoError(t, err)
	return decoded
}

func mustDecodeBase64(t *testing.T, value string) []byte {
	t.Helper()
	decoded, err := base64.StdEncoding.Strict().DecodeString(value)
	require.NoError(t, err)
	return decoded
}
