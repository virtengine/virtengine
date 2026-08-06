package inference

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

func TestReceiptProducerDeterministicInProcessVector(t *testing.T) {
	scorer := validReceiptScorer()
	signer := newTestReceiptSigner(t)
	producer, err := NewReceiptProducer(scorer, signer)
	require.NoError(t, err)
	request := validReceiptProductionRequest()
	originalInputs := cloneScoreInputs(request.ScoreInputs)
	originalReasons := append([]string(nil), scorer.result.ReasonCodes...)

	first, err := producer.Produce(request)
	require.NoError(t, err)
	require.Equal(t, 1, scorer.calls)
	require.Equal(t, originalInputs, request.ScoreInputs)
	require.Equal(t, originalReasons, scorer.result.ReasonCodes)
	require.Equal(t, []string{"scope-a", "scope-b"}, first.ScopeIDs)
	require.Equal(t, uint32(876_543), first.ConfidenceMillionths)
	require.Equal(t, veidtypes.VerificationResultStatusSuccess, first.Status)
	require.Equal(t, []veidtypes.ReasonCode{veidtypes.ReasonCodeSuccess}, first.ReasonCodes)
	require.NoError(t, first.VerifySignature(signer.publicKey))
	signBytes, err := first.SignBytes()
	require.NoError(t, err)
	require.Equal(t, signBytes, signer.signBytes)

	secondScorer := validReceiptScorer()
	secondSigner := newTestReceiptSigner(t)
	secondProducer, err := NewReceiptProducer(secondScorer, secondSigner)
	require.NoError(t, err)
	second, err := secondProducer.Produce(request)
	require.NoError(t, err)
	firstDigest, err := first.DigestHex()
	require.NoError(t, err)
	secondDigest, err := second.DigestHex()
	require.NoError(t, err)
	require.Equal(t, "1612825d75caee35c4b030d158bd2769dca0973fc399d9fde2e528b82b06847d", firstDigest)
	require.Equal(t, firstDigest, secondDigest)
	require.True(t, reflect.DeepEqual(first, second))

	request.ScopeIDs[0] = "changed"
	request.InputDigest[0] ^= 0xff
	request.ScoreInputs.FaceEmbedding[0] = 99
	request.ScoreInputs.OCRConfidences["name"] = 0
	scorer.result.ReasonCodes[0] = ReasonCodeLowConfidence
	require.Equal(t, []string{"scope-a", "scope-b"}, first.ScopeIDs)
	require.Equal(t, byte(1), first.InputDigest[0])
	require.Equal(t, []veidtypes.ReasonCode{veidtypes.ReasonCodeSuccess}, first.ReasonCodes)
}

func TestReceiptProducerSnapshotsEntireRequestBeforeScoring(t *testing.T) {
	request := validReceiptProductionRequest()
	expected := cloneReceiptProductionRequest(request)
	scorer := validReceiptScorer()
	scorer.mutate = func() {
		request.ChainID = "mutated-chain"
		request.AccountAddress = "mutated-account"
		request.RequestID = "mutated-request"
		request.ScopeIDs[0] = "mutated-scope"
		request.Nonce = "mutated-nonce"
		request.InputDigest[0] ^= 0xff
		request.FeatureDigest[0] ^= 0xff
		request.SchemaDigest[0] ^= 0xff
		request.EvidenceLineageDigest[0] ^= 0xff
		request.PipelineVersion = "mutated-pipeline"
		request.ModelManifestDigest[0] ^= 0xff
		request.ModelDigest[0] ^= 0xff
		request.RuntimeImageDigest[0] ^= 0xff
		request.RuntimeDigest[0] ^= 0xff
		request.ConfigDigest[0] ^= 0xff
		request.IssuedHeight++
		request.IssuedAt = request.IssuedAt.Add(time.Second)
		request.ExpiresHeight++
		request.ExpiresAt = request.ExpiresAt.Add(time.Second)
		request.ExpectedScorerModelVersion = "mutated-model"
		request.ExpectedScorerModelHash = strings.Repeat("f", sha256.Size*2)
		request.ScoreInputs.FaceEmbedding[0] = 99
		request.ScoreInputs.OCRConfidences["name"] = 0
		request.ScoreInputs.OCRFieldValidation["name"] = false
		request.ScoreInputs.ScopeTypes[0] = "mutated"
	}
	producer, err := NewReceiptProducer(scorer, newTestReceiptSigner(t))
	require.NoError(t, err)

	receipt, err := producer.Produce(request)
	require.NoError(t, err)
	require.Equal(t, expected.ChainID, receipt.ChainID)
	require.Equal(t, expected.AccountAddress, receipt.AccountAddress)
	require.Equal(t, expected.RequestID, receipt.RequestID)
	require.Equal(t, []string{"scope-a", "scope-b"}, receipt.ScopeIDs)
	require.Equal(t, expected.Nonce, receipt.Nonce)
	require.Equal(t, expected.InputDigest, receipt.InputDigest)
	require.Equal(t, expected.FeatureDigest, receipt.FeatureDigest)
	require.Equal(t, expected.SchemaDigest, receipt.SchemaDigest)
	require.Equal(t, expected.EvidenceLineageDigest, receipt.EvidenceLineageDigest)
	require.Equal(t, expected.PipelineVersion, receipt.PipelineVersion)
	require.Equal(t, expected.ModelManifestDigest, receipt.ModelManifestDigest)
	require.Equal(t, expected.ModelDigest, receipt.ModelDigest)
	require.Equal(t, expected.RuntimeImageDigest, receipt.RuntimeImageDigest)
	require.Equal(t, expected.RuntimeDigest, receipt.RuntimeDigest)
	require.Equal(t, expected.ConfigDigest, receipt.ConfigDigest)
	require.Equal(t, expected.IssuedHeight, receipt.IssuedHeight)
	require.Equal(t, expected.IssuedAt, receipt.IssuedAt)
	require.Equal(t, expected.ExpiresHeight, receipt.ExpiresHeight)
	require.Equal(t, expected.ExpiresAt, receipt.ExpiresAt)
}

func TestReceiptProducerRequestValidation(t *testing.T) {
	long := strings.Repeat("x", veidtypes.InferenceReceiptMaxString+1)
	cases := map[string]func(*ReceiptProductionRequest){
		"nil_inputs":       func(r *ReceiptProductionRequest) { r.ScoreInputs = nil },
		"chain":            func(r *ReceiptProductionRequest) { r.ChainID = "" },
		"account":          func(r *ReceiptProductionRequest) { r.AccountAddress = "" },
		"request":          func(r *ReceiptProductionRequest) { r.RequestID = "" },
		"nonce":            func(r *ReceiptProductionRequest) { r.Nonce = "" },
		"pipeline":         func(r *ReceiptProductionRequest) { r.PipelineVersion = "" },
		"identifier_bound": func(r *ReceiptProductionRequest) { r.ChainID = long },
		"no_scopes":        func(r *ReceiptProductionRequest) { r.ScopeIDs = nil },
		"scope_bound":      func(r *ReceiptProductionRequest) { r.ScopeIDs = []string{long} },
		"too_many_scopes": func(r *ReceiptProductionRequest) {
			r.ScopeIDs = numberedScopes(veidtypes.InferenceReceiptMaxScopes + 1)
		},
		"input_digest":           func(r *ReceiptProductionRequest) { r.InputDigest = r.InputDigest[:31] },
		"feature_digest":         func(r *ReceiptProductionRequest) { r.FeatureDigest = r.FeatureDigest[:31] },
		"schema_digest":          func(r *ReceiptProductionRequest) { r.SchemaDigest = r.SchemaDigest[:31] },
		"lineage_digest":         func(r *ReceiptProductionRequest) { r.EvidenceLineageDigest = r.EvidenceLineageDigest[:31] },
		"manifest_digest":        func(r *ReceiptProductionRequest) { r.ModelManifestDigest = r.ModelManifestDigest[:31] },
		"model_digest":           func(r *ReceiptProductionRequest) { r.ModelDigest = r.ModelDigest[:31] },
		"runtime_image_digest":   func(r *ReceiptProductionRequest) { r.RuntimeImageDigest = r.RuntimeImageDigest[:31] },
		"runtime_digest":         func(r *ReceiptProductionRequest) { r.RuntimeDigest = r.RuntimeDigest[:31] },
		"config_digest_length":   func(r *ReceiptProductionRequest) { r.ConfigDigest = r.ConfigDigest[:31] },
		"config_digest_value":    func(r *ReceiptProductionRequest) { r.ConfigDigest[0] ^= 1 },
		"issued_height":          func(r *ReceiptProductionRequest) { r.IssuedHeight = 0 },
		"expires_height":         func(r *ReceiptProductionRequest) { r.ExpiresHeight = r.IssuedHeight },
		"issued_time":            func(r *ReceiptProductionRequest) { r.IssuedAt = time.Time{} },
		"expires_time":           func(r *ReceiptProductionRequest) { r.ExpiresAt = r.IssuedAt },
		"issued_not_utc":         func(r *ReceiptProductionRequest) { r.IssuedAt = r.IssuedAt.In(time.FixedZone("offset", 3600)) },
		"expires_not_utc":        func(r *ReceiptProductionRequest) { r.ExpiresAt = r.ExpiresAt.In(time.FixedZone("offset", 3600)) },
		"issued_precision":       func(r *ReceiptProductionRequest) { r.IssuedAt = r.IssuedAt.Add(time.Nanosecond) },
		"expires_precision":      func(r *ReceiptProductionRequest) { r.ExpiresAt = r.ExpiresAt.Add(time.Nanosecond) },
		"height_lifetime": func(r *ReceiptProductionRequest) {
			r.ExpiresHeight = r.IssuedHeight + veidtypes.InferenceReceiptMaxHeightLifetime + 1
		},
		"time_lifetime": func(r *ReceiptProductionRequest) {
			r.ExpiresAt = r.IssuedAt.Add(veidtypes.InferenceReceiptMaxLifetime + time.Second)
		},
		"expected_version_empty": func(r *ReceiptProductionRequest) { r.ExpectedScorerModelVersion = "" },
		"expected_hash_empty":    func(r *ReceiptProductionRequest) { r.ExpectedScorerModelHash = "" },
		"expected_hash_upper": func(r *ReceiptProductionRequest) {
			r.ExpectedScorerModelHash = strings.ToUpper(r.ExpectedScorerModelHash)
		},
		"expected_hash_prefix": func(r *ReceiptProductionRequest) { r.ExpectedScorerModelHash = "sha256:" + r.ExpectedScorerModelHash },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			scorer := validReceiptScorer()
			producer, err := NewReceiptProducer(scorer, newTestReceiptSigner(t))
			require.NoError(t, err)
			request := validReceiptProductionRequest()
			mutate(request)
			receipt, err := producer.Produce(request)
			require.Error(t, err)
			require.Nil(t, receipt)
			require.Zero(t, scorer.calls)
		})
	}

	producer, err := NewReceiptProducer(validReceiptScorer(), newTestReceiptSigner(t))
	require.NoError(t, err)
	receipt, err := producer.Produce(nil)
	require.Error(t, err)
	require.Nil(t, receipt)

	boundary := validReceiptProductionRequest()
	boundary.ExpiresHeight = boundary.IssuedHeight + veidtypes.InferenceReceiptMaxHeightLifetime
	boundary.ExpiresAt = boundary.IssuedAt.Add(veidtypes.InferenceReceiptMaxLifetime)
	receipt, err = producer.Produce(boundary)
	require.NoError(t, err)
	require.NotNil(t, receipt)
}

func TestReceiptProducerDependencyAndSignerValidation(t *testing.T) {
	var nilScorer *testReceiptScorer
	var nilSigner *testReceiptSigner
	producer, err := NewReceiptProducer(nilScorer, newTestReceiptSigner(t))
	require.Error(t, err)
	require.Nil(t, producer)
	producer, err = NewReceiptProducer(validReceiptScorer(), nilSigner)
	require.Error(t, err)
	require.Nil(t, producer)

	cases := map[string]func(*testReceiptSigner){
		"key_id":           func(s *testReceiptSigner) { s.keyID = "" },
		"fingerprint":      func(s *testReceiptSigner) { s.fingerprint = "bad" },
		"fingerprint_case": func(s *testReceiptSigner) { s.fingerprint = strings.ToUpper(s.fingerprint) },
		"sequence":         func(s *testReceiptSigner) { s.sequence = 0 },
		"sign_error":       func(s *testReceiptSigner) { s.err = errors.New("sign failed") },
		"signature_length": func(s *testReceiptSigner) { s.signature = make([]byte, ed25519.SignatureSize-1) },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			signer := newTestReceiptSigner(t)
			mutate(signer)
			producer, err := NewReceiptProducer(validReceiptScorer(), signer)
			require.NoError(t, err)
			receipt, err := producer.Produce(validReceiptProductionRequest())
			require.Error(t, err)
			require.Nil(t, receipt)
		})
	}
}

func TestReceiptProducerScorerFailuresAndMetadata(t *testing.T) {
	cases := map[string]func(*testReceiptScorer){
		"unhealthy":           func(s *testReceiptScorer) { s.healthy = false },
		"error":               func(s *testReceiptScorer) { s.err = errors.New("score failed") },
		"nil_result":          func(s *testReceiptScorer) { s.result = nil },
		"score_bound":         func(s *testReceiptScorer) { s.result.Score = 101 },
		"confidence_negative": func(s *testReceiptScorer) { s.result.Confidence = -0.1 },
		"confidence_high":     func(s *testReceiptScorer) { s.result.Confidence = 1.1 },
		"confidence_nan":      func(s *testReceiptScorer) { s.result.Confidence = float32(math.NaN()) },
		"confidence_inf":      func(s *testReceiptScorer) { s.result.Confidence = float32(math.Inf(1)) },
		"accessor_version":    func(s *testReceiptScorer) { s.modelVersion = "other" },
		"accessor_hash":       func(s *testReceiptScorer) { s.modelHash = strings.Repeat("b", 64) },
		"result_version":      func(s *testReceiptScorer) { s.result.ModelVersion = "other" },
		"result_hash":         func(s *testReceiptScorer) { s.result.ModelHash = strings.Repeat("b", 64) },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			scorer := validReceiptScorer()
			mutate(scorer)
			producer, err := NewReceiptProducer(scorer, newTestReceiptSigner(t))
			require.NoError(t, err)
			receipt, err := producer.Produce(validReceiptProductionRequest())
			require.Error(t, err)
			require.Nil(t, receipt)
			if name == "unhealthy" || strings.HasPrefix(name, "accessor_") {
				require.Zero(t, scorer.calls)
			} else {
				require.Equal(t, 1, scorer.calls)
			}
		})
	}
}

func TestReceiptProducerReasonNormalization(t *testing.T) {
	cases := []struct {
		scorerReason string
		want         veidtypes.ReasonCode
	}{
		{ReasonCodeSuccess, ""},
		{ReasonCodeHighConfidence, ""},
		{ReasonCodeLowConfidence, veidtypes.ReasonCodeLowConfidence},
		{ReasonCodeFaceMismatch, veidtypes.ReasonCodeFaceMismatch},
		{ReasonCodeLowDocQuality, veidtypes.ReasonCodeLowDocQuality},
		{ReasonCodeLowOCRConfidence, veidtypes.ReasonCodeLowOCRConfidence},
		{ReasonCodeInsufficientScopes, veidtypes.ReasonCodeInsufficientScopes},
		{ReasonCodeMissingFace, veidtypes.ReasonCodeFaceMismatch},
		{ReasonCodeMissingDocument, veidtypes.ReasonCodeDocumentInvalid},
		{ReasonCodeModelLoadError, veidtypes.ReasonCodeMLInferenceError},
		{ReasonCodeInferenceError, veidtypes.ReasonCodeMLInferenceError},
		{ReasonCodeTimeout, veidtypes.ReasonCodeTimeout},
		{ReasonCodeMemoryLimit, veidtypes.ReasonCodeMLInferenceError},
	}
	for _, tc := range cases {
		t.Run(tc.scorerReason, func(t *testing.T) {
			scorer := validReceiptScorer()
			scorer.result.ReasonCodes = []string{ReasonCodeSuccess, ReasonCodeHighConfidence, tc.scorerReason, tc.scorerReason}
			producer, err := NewReceiptProducer(scorer, newTestReceiptSigner(t))
			require.NoError(t, err)
			receipt, err := producer.Produce(validReceiptProductionRequest())
			require.NoError(t, err)
			if tc.want == "" {
				require.Equal(t, veidtypes.VerificationResultStatusSuccess, receipt.Status)
				require.Equal(t, []veidtypes.ReasonCode{veidtypes.ReasonCodeSuccess}, receipt.ReasonCodes)
			} else {
				require.Equal(t, veidtypes.VerificationResultStatusPartial, receipt.Status)
				require.Equal(t, []veidtypes.ReasonCode{tc.want}, receipt.ReasonCodes)
			}
			require.Equal(t, scorer.result.Score, receipt.Score)
		})
	}

	scorer := validReceiptScorer()
	scorer.result.ReasonCodes = []string{ReasonCodeLowOCRConfidence, ReasonCodeLowConfidence, ReasonCodeLowOCRConfidence}
	producer, err := NewReceiptProducer(scorer, newTestReceiptSigner(t))
	require.NoError(t, err)
	receipt, err := producer.Produce(validReceiptProductionRequest())
	require.NoError(t, err)
	require.Equal(t, []veidtypes.ReasonCode{veidtypes.ReasonCodeLowConfidence, veidtypes.ReasonCodeLowOCRConfidence}, receipt.ReasonCodes)

	scorer = validReceiptScorer()
	scorer.result.ReasonCodes = []string{"UNKNOWN"}
	producer, err = NewReceiptProducer(scorer, newTestReceiptSigner(t))
	require.NoError(t, err)
	receipt, err = producer.Produce(validReceiptProductionRequest())
	require.Error(t, err)
	require.Nil(t, receipt)
}

func TestReceiptProducerModelDigestIsIndependentScorerCommitment(t *testing.T) {
	producer, err := NewReceiptProducer(validReceiptScorer(), newTestReceiptSigner(t))
	require.NoError(t, err)
	firstRequest := validReceiptProductionRequest()
	first, err := producer.Produce(firstRequest)
	require.NoError(t, err)

	secondRequest := validReceiptProductionRequest()
	secondRequest.ModelDigest[0] ^= 0xff
	second, err := producer.Produce(secondRequest)
	require.NoError(t, err)
	require.Equal(t, secondRequest.ModelDigest, second.ModelDigest)
	firstDigest, err := first.DigestHex()
	require.NoError(t, err)
	secondDigest, err := second.DigestHex()
	require.NoError(t, err)
	require.NotEqual(t, firstDigest, secondDigest)
}

func TestReceiptProducerCommitmentTamperingChangesDigestAfterResigning(t *testing.T) {
	signer := newTestReceiptSigner(t)
	producer, err := NewReceiptProducer(validReceiptScorer(), signer)
	require.NoError(t, err)
	base, err := producer.Produce(validReceiptProductionRequest())
	require.NoError(t, err)
	baseDigest, err := base.DigestHex()
	require.NoError(t, err)

	cases := map[string]func(*veidtypes.InferenceReceipt){
		"input":         func(r *veidtypes.InferenceReceipt) { r.InputDigest[0] ^= 1 },
		"feature":       func(r *veidtypes.InferenceReceipt) { r.FeatureDigest[0] ^= 1 },
		"schema":        func(r *veidtypes.InferenceReceipt) { r.SchemaDigest[0] ^= 1 },
		"lineage":       func(r *veidtypes.InferenceReceipt) { r.EvidenceLineageDigest[0] ^= 1 },
		"manifest":      func(r *veidtypes.InferenceReceipt) { r.ModelManifestDigest[0] ^= 1 },
		"model":         func(r *veidtypes.InferenceReceipt) { r.ModelDigest[0] ^= 1 },
		"runtime_image": func(r *veidtypes.InferenceReceipt) { r.RuntimeImageDigest[0] ^= 1 },
		"runtime":       func(r *veidtypes.InferenceReceipt) { r.RuntimeDigest[0] ^= 1 },
		"pipeline":      func(r *veidtypes.InferenceReceipt) { r.PipelineVersion = "pipeline-v2" },
	}
	for name, tamper := range cases {
		t.Run(name, func(t *testing.T) {
			changed := cloneInferenceReceipt(base)
			tamper(changed)
			require.NoError(t, changed.Sign(signer.privateKey))
			require.NoError(t, changed.Validate())
			changedDigest, err := changed.DigestHex()
			require.NoError(t, err)
			require.NotEqual(t, baseDigest, changedDigest)
		})
	}
}

type testReceiptScorer struct {
	result       *ScoreResult
	err          error
	healthy      bool
	modelVersion string
	modelHash    string
	calls        int
	mutate       func()
}

func (s *testReceiptScorer) ComputeScore(inputs *ScoreInputs) (*ScoreResult, error) {
	s.calls++
	inputs.FaceEmbedding[0] = -1
	inputs.OCRConfidences["name"] = -1
	inputs.ScopeTypes[0] = "mutated"
	if s.mutate != nil {
		s.mutate()
	}
	return s.result, s.err
}
func (s *testReceiptScorer) GetModelVersion() string { return s.modelVersion }
func (s *testReceiptScorer) GetModelHash() string    { return s.modelHash }
func (s *testReceiptScorer) IsHealthy() bool         { return s.healthy }
func (s *testReceiptScorer) Close() error            { return nil }

type testReceiptSigner struct {
	publicKey   ed25519.PublicKey
	privateKey  ed25519.PrivateKey
	keyID       string
	fingerprint string
	sequence    uint64
	signBytes   []byte
	signature   []byte
	err         error
}

func newTestReceiptSigner(t *testing.T) *testReceiptSigner {
	t.Helper()
	seed := sha256.Sum256([]byte("virtengine/receipt-producer/test-key/v1"))
	publicKey, privateKey, err := ed25519.GenerateKey(bytes.NewReader(seed[:]))
	require.NoError(t, err)
	return &testReceiptSigner{
		publicKey: publicKey, privateKey: privateKey,
		keyID:       "did:virtengine:receipt-producer:1",
		fingerprint: veidtypes.ComputeKeyFingerprint(publicKey), sequence: 7,
	}
}

func (s *testReceiptSigner) KeyID() string       { return s.keyID }
func (s *testReceiptSigner) Fingerprint() string { return s.fingerprint }
func (s *testReceiptSigner) Sequence() uint64    { return s.sequence }
func (s *testReceiptSigner) Sign(value []byte) ([]byte, error) {
	s.signBytes = append([]byte(nil), value...)
	if s.err != nil {
		return nil, s.err
	}
	if s.signature != nil {
		return append([]byte(nil), s.signature...), nil
	}
	return ed25519.Sign(s.privateKey, value), nil
}

func validReceiptScorer() *testReceiptScorer {
	hash := strings.Repeat("a", sha256.Size*2)
	return &testReceiptScorer{
		healthy: true, modelVersion: "model-v1", modelHash: hash,
		result: &ScoreResult{Score: 91, Confidence: 0.876543, ModelVersion: "model-v1", ModelHash: hash, ReasonCodes: []string{ReasonCodeSuccess}},
	}
}

func validReceiptProductionRequest() *ReceiptProductionRequest {
	now := time.Unix(1_700_000_000, 0).UTC()
	return &ReceiptProductionRequest{
		ScoreInputs: &ScoreInputs{
			FaceEmbedding: []float32{1, 2}, OCRConfidences: map[string]float32{"name": 0.9},
			OCRFieldValidation: map[string]bool{"name": true}, ScopeTypes: []string{"face"}, ScopeCount: 1,
		},
		ChainID: "chain-A", AccountAddress: "virt1account", RequestID: "request-1",
		ScopeIDs: []string{"scope-b", "scope-a", "scope-a"}, Nonce: "nonce-1",
		InputDigest: testProducerDigest(1), FeatureDigest: testProducerDigest(2), SchemaDigest: testProducerDigest(3),
		EvidenceLineageDigest: testProducerDigest(4), PipelineVersion: "pipeline-v1",
		ModelManifestDigest: testProducerDigest(5), ModelDigest: testProducerDigest(6),
		RuntimeImageDigest: testProducerDigest(7), RuntimeDigest: testProducerDigest(8),
		ConfigDigest: veidtypes.CanonicalInferenceDeterminismConfigDigest(),
		IssuedHeight: 10, IssuedAt: now, ExpiresHeight: 12, ExpiresAt: now.Add(time.Minute),
		ExpectedScorerModelVersion: "model-v1", ExpectedScorerModelHash: strings.Repeat("a", sha256.Size*2),
	}
}

func testProducerDigest(value byte) []byte {
	return bytes.Repeat([]byte{value}, sha256.Size)
}

func numberedScopes(count int) []string {
	scopes := make([]string, count)
	for index := range scopes {
		scopes[index] = hex.EncodeToString([]byte{byte(index)})
	}
	return scopes
}
