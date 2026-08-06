package inference

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIsRealInferenceEnabledRequiresSidecar(t *testing.T) {
	config := DefaultInferenceConfig()
	config.Enabled = true
	config.ModelPath = "models/trust_score"

	if config.IsRealInferenceEnabled() {
		t.Fatal("embedded placeholder mode must not report real inference as enabled")
	}

	config.UseSidecar = true
	config.UseFallbackOnError = false
	config.StrictDeterminism = true
	config.SidecarTLS = true
	config.SidecarTLSCAFile = "test-ca.pem"
	config.SidecarTLSCertFile = "test-cert.pem"
	config.SidecarTLSKeyFile = "test-key.pem"
	config.SidecarTLSServerName = "veid-inference.test"
	if !config.IsRealInferenceEnabled() {
		t.Fatal("strict sidecar-backed config should report real inference as enabled")
	}

	config.UseFallbackOnError = true
	if config.IsRealInferenceEnabled() {
		t.Fatal("fallback score config must not report real inference as enabled")
	}
}

func TestNewTensorFlowScorerRequiresExplicitStubOptIn(t *testing.T) {
	modelDir := createTestModelDir(t)

	config := DefaultInferenceConfig()
	config.ModelPath = modelDir
	setExpectedHashForModel(t, &config, modelDir)

	_, err := NewTensorFlowScorer(config)
	if !errors.Is(err, ErrSimulatedInferenceDisabled) {
		t.Fatalf("expected simulated inference disabled error, got %v", err)
	}
}

func TestModelRunFailsClosedWithoutStubOptIn(t *testing.T) {
	modelDir := createTestModelDir(t)

	config := DefaultInferenceConfig()
	config.ModelPath = modelDir
	setExpectedHashForModel(t, &config, modelDir)

	loader := NewModelLoader(config)
	model, err := loader.Load()
	if err != nil {
		t.Fatalf("load model: %v", err)
	}
	defer func() { _ = loader.Unload() }()

	_, err = model.Run(make([]float32, TotalFeatureDim))
	if !errors.Is(err, ErrSimulatedInferenceDisabled) {
		t.Fatalf("expected simulated inference disabled error, got %v", err)
	}
}

func TestSidecarClientRequiresExplicitStubOptIn(t *testing.T) {
	client, features, inputs := newNilGRPCSidecarClient(t, false)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.callSidecar(ctx, features, inputs)
	if !errors.Is(err, ErrSimulatedInferenceDisabled) {
		t.Fatalf("expected simulated inference disabled error from callSidecar, got %v", err)
	}

	if _, err := client.PerformHealthCheck(ctx); !errors.Is(err, ErrSimulatedInferenceDisabled) {
		t.Fatalf("expected simulated inference disabled error from health check, got %v", err)
	}

	if _, err := client.GetMetrics(ctx); !errors.Is(err, ErrSimulatedInferenceDisabled) {
		t.Fatalf("expected simulated inference disabled error from metrics, got %v", err)
	}

	if _, err := client.VerifyDeterminism(ctx, "test-vector"); !errors.Is(err, ErrSimulatedInferenceDisabled) {
		t.Fatalf("expected simulated inference disabled error from determinism check, got %v", err)
	}
}

func TestSidecarClientAllowsExplicitStubOptIn(t *testing.T) {
	client, features, inputs := newNilGRPCSidecarClient(t, true)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := client.callSidecar(ctx, features, inputs)
	if err != nil {
		t.Fatalf("expected explicit stub sidecar response, got error %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil score result")
	}

	health, err := client.PerformHealthCheck(ctx)
	if err != nil {
		t.Fatalf("expected stub health response, got error %v", err)
	}
	if !health.ModelLoaded {
		t.Fatal("expected stub health response to report model loaded")
	}

	metrics, err := client.GetMetrics(ctx)
	if err != nil {
		t.Fatalf("expected stub metrics response, got error %v", err)
	}
	if metrics.ModelVersion != client.config.ModelVersion {
		t.Fatalf("expected metrics model version %s, got %s", client.config.ModelVersion, metrics.ModelVersion)
	}

	determinism, err := client.VerifyDeterminism(ctx, "test-vector")
	if err != nil {
		t.Fatalf("expected stub determinism response, got error %v", err)
	}
	if !determinism.Passed {
		t.Fatal("expected stub determinism response to pass")
	}
}

func newNilGRPCSidecarClient(t *testing.T, allowStub bool) (*SidecarClient, []float32, *ScoreInputs) {
	t.Helper()

	config := DefaultInferenceConfig()
	config.UseSidecar = true
	config.SidecarAddress = testSidecarAddress
	config.SidecarTimeout = time.Second
	config.ModelVersion = "test-model"
	config.ExpectedHash = strings.Repeat("a", 64)
	config.AllowFallbackToStub = allowStub

	client := &SidecarClient{
		config:       config,
		extractor:    NewFeatureExtractor(DefaultFeatureExtractorConfig()),
		determinism:  NewDeterminismController(config.RandomSeed, config.ForceCPU),
		isConnected:  false,
		modelVersion: config.ModelVersion,
		modelHash:    config.ExpectedHash,
	}

	inputs := createTestInputs()
	features, err := client.extractor.ExtractFeatures(inputs)
	if err != nil {
		t.Fatalf("extract features: %v", err)
	}

	return client, features, inputs
}

func createTestModelDir(t *testing.T) string {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "inference-model-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tempDir) })

	modelDir := filepath.Join(tempDir, "model")
	if err := os.MkdirAll(modelDir, 0o750); err != nil {
		t.Fatalf("create model dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(modelDir, "saved_model.pb"), []byte("test"), 0o600); err != nil {
		t.Fatalf("create model file: %v", err)
	}

	return modelDir
}
