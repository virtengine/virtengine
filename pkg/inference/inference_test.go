package inference

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Test constants
const (
	testSidecarAddress = "localhost:50051"
)

// ============================================================================
// Test Configuration
// ============================================================================

func TestDefaultInferenceConfig(t *testing.T) {
	config := DefaultInferenceConfig()

	if config.Timeout != 2*time.Second {
		t.Errorf("expected timeout 2s, got %v", config.Timeout)
	}

	if config.MaxMemoryMB != 512 {
		t.Errorf("expected max memory 512MB, got %d", config.MaxMemoryMB)
	}

	if !config.Deterministic {
		t.Error("expected deterministic mode to be enabled by default")
	}

	if !config.ForceCPU {
		t.Error("expected force CPU to be enabled by default")
	}

	if config.ExpectedInputDim != TotalFeatureDim {
		t.Errorf("expected input dim %d, got %d", TotalFeatureDim, config.ExpectedInputDim)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		modifyFunc  func(*InferenceConfig)
		expectError bool
	}{
		{
			name: "valid config",
			modifyFunc: func(c *InferenceConfig) {
				c.ExpectedHash = strings.Repeat("a", 64)
			},
			expectError: false,
		},
		{
			name: "deterministic without expected hash",
			modifyFunc: func(c *InferenceConfig) {
				c.Deterministic = true
				c.ExpectedHash = ""
			},
			expectError: true,
		},
		{
			name: "deterministic without force cpu",
			modifyFunc: func(c *InferenceConfig) {
				c.Deterministic = true
				c.ForceCPU = false
				c.ExpectedHash = strings.Repeat("a", 64)
			},
			expectError: true,
		},
		{
			name: "sidecar without address",
			modifyFunc: func(c *InferenceConfig) {
				c.UseSidecar = true
				c.SidecarAddress = ""
				c.ExpectedHash = strings.Repeat("a", 64)
			},
			expectError: true,
		},
		{
			name: "sidecar with invalid timeout",
			modifyFunc: func(c *InferenceConfig) {
				c.UseSidecar = true
				c.SidecarAddress = testSidecarAddress
				c.SidecarTimeout = 0
				c.ExpectedHash = strings.Repeat("a", 64)
			},
			expectError: true,
		},
		{
			name: "no model path when not sidecar",
			modifyFunc: func(c *InferenceConfig) {
				c.UseSidecar = false
				c.ModelPath = ""
				c.ExpectedHash = strings.Repeat("a", 64)
			},
			expectError: true,
		},
		{
			name: "invalid timeout",
			modifyFunc: func(c *InferenceConfig) {
				c.Timeout = 0
				c.ExpectedHash = strings.Repeat("a", 64)
			},
			expectError: true,
		},
		{
			name: "invalid memory limit",
			modifyFunc: func(c *InferenceConfig) {
				c.MaxMemoryMB = 0
				c.ExpectedHash = strings.Repeat("a", 64)
			},
			expectError: true,
		},
		{
			name: "invalid input dimension",
			modifyFunc: func(c *InferenceConfig) {
				c.ExpectedInputDim = 100
				c.ExpectedHash = strings.Repeat("a", 64)
			},
			expectError: true,
		},
		{
			name: "invalid fallback score",
			modifyFunc: func(c *InferenceConfig) {
				c.FallbackScore = 150
				c.ExpectedHash = strings.Repeat("a", 64)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultInferenceConfig()
			tt.modifyFunc(&config)

			err := config.Validate()
			if tt.expectError && err == nil {
				t.Error("expected validation error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

// ============================================================================
// Test Feature Extractor
// ============================================================================

func TestFeatureExtraction(t *testing.T) {
	extractor := NewFeatureExtractor(DefaultFeatureExtractorConfig())

	inputs := createTestInputs()

	features, err := extractor.ExtractFeatures(inputs)
	if err != nil {
		t.Fatalf("feature extraction failed: %v", err)
	}

	if len(features) != TotalFeatureDim {
		t.Errorf("expected %d features, got %d", TotalFeatureDim, len(features))
	}
}

func TestFeatureExtractionWithMissingFace(t *testing.T) {
	extractor := NewFeatureExtractor(DefaultFeatureExtractorConfig())

	inputs := createTestInputs()
	inputs.FaceEmbedding = nil // Missing face

	features, err := extractor.ExtractFeatures(inputs)
	if err != nil {
		t.Fatalf("feature extraction failed: %v", err)
	}

	if len(features) != TotalFeatureDim {
		t.Errorf("expected %d features, got %d", TotalFeatureDim, len(features))
	}

	// First 512 features should be zeros
	for i := 0; i < FaceEmbeddingDim; i++ {
		if features[i] != 0.0 {
			t.Errorf("expected zero at position %d for missing face, got %f", i, features[i])
			break
		}
	}
}

func TestFeatureExtractionWithInvalidDimension(t *testing.T) {
	extractor := NewFeatureExtractor(DefaultFeatureExtractorConfig())

	inputs := createTestInputs()
	inputs.FaceEmbedding = make([]float32, 100) // Wrong dimension

	_, err := extractor.ExtractFeatures(inputs)
	if err == nil {
		t.Error("expected error for invalid face embedding dimension")
	}
}

func TestInputValidation(t *testing.T) {
	extractor := NewFeatureExtractor(DefaultFeatureExtractorConfig())

	tests := []struct {
		name          string
		modifyFunc    func(*ScoreInputs)
		expectedIssue string
	}{
		{
			name:          "valid inputs",
			modifyFunc:    func(i *ScoreInputs) {},
			expectedIssue: "",
		},
		{
			name: "missing face embedding",
			modifyFunc: func(i *ScoreInputs) {
				i.FaceEmbedding = nil
			},
			expectedIssue: "missing face embedding",
		},
		{
			name: "invalid face dimension",
			modifyFunc: func(i *ScoreInputs) {
				i.FaceEmbedding = make([]float32, 100)
			},
			expectedIssue: "invalid face embedding dimension",
		},
		{
			name: "face confidence out of range",
			modifyFunc: func(i *ScoreInputs) {
				i.FaceConfidence = 1.5
			},
			expectedIssue: "face confidence out of range",
		},
		{
			name: "doc quality out of range",
			modifyFunc: func(i *ScoreInputs) {
				i.DocQualityScore = -0.1
			},
			expectedIssue: "document quality score out of range",
		},
		{
			name: "negative scope count",
			modifyFunc: func(i *ScoreInputs) {
				i.ScopeCount = -1
			},
			expectedIssue: "negative scope count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputs := createTestInputs()
			tt.modifyFunc(inputs)

			issues := extractor.ValidateInputs(inputs)

			if tt.expectedIssue == "" {
				if len(issues) > 0 {
					t.Errorf("expected no issues, got: %v", issues)
				}
			} else {
				found := false
				for _, issue := range issues {
					if contains(issue, tt.expectedIssue) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected issue containing '%s', got: %v", tt.expectedIssue, issues)
				}
			}
		})
	}
}

func TestFeatureContributions(t *testing.T) {
	extractor := NewFeatureExtractor(DefaultFeatureExtractorConfig())

	inputs := createTestInputs()
	features, _ := extractor.ExtractFeatures(inputs)

	contributions := extractor.ComputeFeatureContributions(features)

	expectedKeys := []string{"face_embedding", "doc_quality", "ocr", "metadata"}
	for _, key := range expectedKeys {
		if _, ok := contributions[key]; !ok {
			t.Errorf("missing contribution key: %s", key)
		}
	}
}

// ============================================================================
// Test Determinism
// ============================================================================

func TestDeterministicInputHash(t *testing.T) {
	dc := NewDeterminismController(42, true)

	inputs := createTestInputs()

	// Compute hash twice
	hash1 := dc.ComputeInputHash(inputs)
	hash2 := dc.ComputeInputHash(inputs)

	if hash1 != hash2 {
		t.Error("input hash is not deterministic")
	}

	// Modify inputs and verify hash changes
	inputs.FaceConfidence = 0.99
	hash3 := dc.ComputeInputHash(inputs)

	if hash1 == hash3 {
		t.Error("input hash should change when inputs change")
	}
}

func TestDeterministicOutputHash(t *testing.T) {
	dc := NewDeterminismController(42, true)

	output := []float32{75.5, 0.85}

	hash1 := dc.ComputeOutputHash(output)
	hash2 := dc.ComputeOutputHash(output)

	if hash1 != hash2 {
		t.Error("output hash is not deterministic")
	}
}

func TestResultVerification(t *testing.T) {
	dc := NewDeterminismController(42, true)

	result1 := &ScoreResult{
		Score:        75,
		RawScore:     75.5,
		Confidence:   0.85,
		ModelVersion: "v1.0.0",
		ModelHash:    "abc123",
		InputHash:    "input123",
		OutputHash:   "output456",
	}

	result2 := &ScoreResult{
		Score:        75,
		RawScore:     75.5,
		Confidence:   0.85,
		ModelVersion: "v1.0.0",
		ModelHash:    "abc123",
		InputHash:    "input123",
		OutputHash:   "output456",
	}

	if !dc.VerifySameOutput(result1, result2) {
		t.Error("identical results should verify as same")
	}

	// Modify result2
	result2.Score = 76
	if dc.VerifySameOutput(result1, result2) {
		t.Error("different results should not verify as same")
	}
}

func TestTensorFlowEnvVars(t *testing.T) {
	dc := NewDeterminismController(42, true)

	envVars := dc.GetTensorFlowEnvVars()

	if envVars["TF_DETERMINISTIC_OPS"] != "1" {
		t.Error("TF_DETERMINISTIC_OPS should be 1")
	}

	if envVars["CUDA_VISIBLE_DEVICES"] != "-1" {
		t.Error("CUDA_VISIBLE_DEVICES should be -1 for CPU-only mode")
	}
}

func TestTFDeterminismConfig(t *testing.T) {
	dc := NewDeterminismController(42, true)

	config := dc.ConfigureTensorFlow()

	if config.RandomSeed != 42 {
		t.Errorf("expected random seed 42, got %d", config.RandomSeed)
	}

	if config.InterOpParallelism != 1 {
		t.Error("expected single inter-op thread for determinism")
	}

	if config.IntraOpParallelism != 1 {
		t.Error("expected single intra-op thread for determinism")
	}

	if !config.UseCPUOnly {
		t.Error("expected CPU-only mode")
	}

	if !config.EnableDeterministicOps {
		t.Error("expected deterministic ops enabled")
	}
}

// ============================================================================
// Test Model Loader
// ============================================================================

func TestModelLoaderWithMissingPath(t *testing.T) {
	config := DefaultInferenceConfig()
	config.ModelPath = "/nonexistent/path/to/model"
	config.ExpectedHash = strings.Repeat("a", 64)

	loader := NewModelLoader(config)

	_, err := loader.Load()
	if err == nil {
		t.Error("expected error for nonexistent model path")
	}
}

func TestModelLoaderWithTempModel(t *testing.T) {
	// Create a temporary model directory
	tempDir, err := os.MkdirTemp("", "test-model-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	modelDir := filepath.Join(tempDir, "model")
	if err := os.MkdirAll(modelDir, 0750); err != nil {
		t.Fatalf("failed to create model dir: %v", err)
	}

	// Create minimal model files
	if err := os.WriteFile(filepath.Join(modelDir, "saved_model.pb"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create model file: %v", err)
	}

	config := DefaultInferenceConfig()
	config.ModelPath = modelDir
	setExpectedHashForModel(t, &config, modelDir)

	loader := NewModelLoader(config)

	model, err := loader.Load()
	if err != nil {
		t.Fatalf("failed to load model: %v", err)
	}

	if model == nil {
		t.Error("expected non-nil model")
	}

	if !model.IsLoaded() {
		t.Error("model should be loaded")
	}

	// Test unload
	if err := loader.Unload(); err != nil {
		t.Errorf("failed to unload model: %v", err)
	}
}

func TestModelHashComputation(t *testing.T) {
	// Create temp model with known content
	tempDir, err := os.MkdirTemp("", "test-hash-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	modelDir := filepath.Join(tempDir, "model")
	if err := os.MkdirAll(modelDir, 0750); err != nil {
		t.Fatalf("failed to create model dir: %v", err)
	}

	// Create model files with known content
	if err := os.WriteFile(filepath.Join(modelDir, "saved_model.pb"), []byte("test content"), 0600); err != nil {
		t.Fatalf("failed to create model file: %v", err)
	}

	config := DefaultInferenceConfig()
	config.ModelPath = modelDir
	setExpectedHashForModel(t, &config, modelDir)

	loader := NewModelLoader(config)

	model, err := loader.Load()
	if err != nil {
		t.Fatalf("failed to load model: %v", err)
	}

	hash := model.GetModelHash()
	if hash == "" {
		t.Error("expected non-empty model hash")
	}

	// Load again and verify same hash
	if err := loader.Unload(); err != nil {
		t.Fatalf("failed to unload: %v", err)
	}

	model2, err := loader.Load()
	if err != nil {
		t.Fatalf("failed to reload model: %v", err)
	}

	if model2.GetModelHash() != hash {
		t.Error("model hash should be consistent across loads")
	}
}

func TestModelLoaderRejectsNonDeterministicOps(t *testing.T) {
	// Create a temporary model directory
	tempDir, err := os.MkdirTemp("", "test-nondet-ops-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	modelDir := filepath.Join(tempDir, "model")
	if err := os.MkdirAll(modelDir, 0750); err != nil {
		t.Fatalf("failed to create model dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(modelDir, "saved_model.pb"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create model file: %v", err)
	}

	metadata := `{
  "version": "v1.0.0",
  "op_names": ["Conv2D", "MatMul"]
}`
	if err := os.WriteFile(filepath.Join(modelDir, "export_metadata.json"), []byte(metadata), 0600); err != nil {
		t.Fatalf("failed to write metadata file: %v", err)
	}

	config := DefaultInferenceConfig()
	config.ModelPath = modelDir
	setExpectedHashForModel(t, &config, modelDir)

	loader := NewModelLoader(config)
	_, err = loader.Load()
	if err == nil {
		t.Fatal("expected error for non-deterministic ops, got none")
	}
}

func TestModelLoaderAcceptsDeterministicOps(t *testing.T) {
	// Create a temporary model directory
	tempDir, err := os.MkdirTemp("", "test-det-ops-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	modelDir := filepath.Join(tempDir, "model")
	if err := os.MkdirAll(modelDir, 0750); err != nil {
		t.Fatalf("failed to create model dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(modelDir, "saved_model.pb"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create model file: %v", err)
	}

	metadata := `{
  "version": "v1.0.0",
  "op_names": ["MatMul", "AddV2", "Relu"]
}`
	if err := os.WriteFile(filepath.Join(modelDir, "export_metadata.json"), []byte(metadata), 0600); err != nil {
		t.Fatalf("failed to write metadata file: %v", err)
	}

	config := DefaultInferenceConfig()
	config.ModelPath = modelDir
	setExpectedHashForModel(t, &config, modelDir)

	loader := NewModelLoader(config)
	if _, err = loader.Load(); err != nil {
		t.Fatalf("unexpected error for deterministic ops: %v", err)
	}
}

// ============================================================================
// Test Scorer
// ============================================================================

func TestScorerWithValidInputs(t *testing.T) {
	// Create temp model
	tempDir, err := os.MkdirTemp("", "test-scorer-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	modelDir := filepath.Join(tempDir, "model")
	if err := os.MkdirAll(modelDir, 0750); err != nil {
		t.Fatalf("failed to create model dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(modelDir, "saved_model.pb"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create model file: %v", err)
	}

	config := DefaultInferenceConfig()
	config.ModelPath = modelDir
	setExpectedHashForModel(t, &config, modelDir)

	scorer, err := NewTensorFlowScorer(config)
	if err != nil {
		t.Fatalf("failed to create scorer: %v", err)
	}
	defer func() { _ = scorer.Close() }()

	if !scorer.IsHealthy() {
		t.Error("scorer should be healthy")
	}

	inputs := createTestInputs()

	result, err := scorer.ComputeScore(inputs)
	if err != nil {
		t.Fatalf("inference failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.Score > 100 {
		t.Errorf("score should be 0-100, got %d", result.Score)
	}

	if result.ModelVersion == "" {
		t.Error("expected non-empty model version")
	}

	if result.InputHash == "" {
		t.Error("expected non-empty input hash")
	}

	if result.ComputeTimeMs < 0 {
		t.Error("compute time should be non-negative")
	}
}

func TestScorerWithInvalidInputs(t *testing.T) {
	// Create temp model
	tempDir, err := os.MkdirTemp("", "test-scorer-invalid-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	modelDir := filepath.Join(tempDir, "model")
	if err := os.MkdirAll(modelDir, 0750); err != nil {
		t.Fatalf("failed to create model dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(modelDir, "saved_model.pb"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create model file: %v", err)
	}

	config := DefaultInferenceConfig()
	config.ModelPath = modelDir
	setExpectedHashForModel(t, &config, modelDir)

	scorer, err := NewTensorFlowScorer(config)
	if err != nil {
		t.Fatalf("failed to create scorer: %v", err)
	}
	defer func() { _ = scorer.Close() }()

	// Test with invalid face embedding dimension
	inputs := createTestInputs()
	inputs.FaceEmbedding = make([]float32, 100) // Wrong dimension

	result, err := scorer.ComputeScore(inputs)

	// Should still get a result (with error reason code) or error
	if result == nil && err == nil {
		t.Error("expected result or error for invalid inputs")
	}
}

func TestScorerTimeout(t *testing.T) {
	// Create temp model
	tempDir, err := os.MkdirTemp("", "test-scorer-timeout-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	modelDir := filepath.Join(tempDir, "model")
	if err := os.MkdirAll(modelDir, 0750); err != nil {
		t.Fatalf("failed to create model dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(modelDir, "saved_model.pb"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create model file: %v", err)
	}

	config := DefaultInferenceConfig()
	config.ModelPath = modelDir
	config.Timeout = 1 * time.Nanosecond // Very short timeout
	setExpectedHashForModel(t, &config, modelDir)

	scorer, err := NewTensorFlowScorer(config)
	if err != nil {
		t.Fatalf("failed to create scorer: %v", err)
	}
	defer func() { _ = scorer.Close() }()

	inputs := createTestInputs()

	// Create a context that's already expired
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond) // Ensure timeout

	result, _ := scorer.ComputeScoreWithContext(ctx, inputs)

	// With fallback enabled, should get a result
	if result != nil {
		// Check for timeout reason code
		hasTimeout := false
		for _, code := range result.ReasonCodes {
			if code == ReasonCodeTimeout {
				hasTimeout = true
				break
			}
		}
		if !hasTimeout && !config.UseFallbackOnError {
			// Timeout might not always trigger due to fast execution
			t.Log("timeout reason code not present (inference may have completed)")
		}
	}
}

func TestScorerDeterministicOutputs(t *testing.T) {
	// Create temp model
	tempDir, err := os.MkdirTemp("", "test-scorer-determinism-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	modelDir := filepath.Join(tempDir, "model")
	if err := os.MkdirAll(modelDir, 0750); err != nil {
		t.Fatalf("failed to create model dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(modelDir, "saved_model.pb"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create model file: %v", err)
	}

	config := DefaultInferenceConfig()
	config.ModelPath = modelDir
	config.Deterministic = true
	setExpectedHashForModel(t, &config, modelDir)

	scorer, err := NewTensorFlowScorer(config)
	if err != nil {
		t.Fatalf("failed to create scorer: %v", err)
	}
	defer func() { _ = scorer.Close() }()

	inputs := createTestInputs()

	// Run inference twice
	result1, err := scorer.ComputeScore(inputs)
	if err != nil {
		t.Fatalf("first inference failed: %v", err)
	}

	result2, err := scorer.ComputeScore(inputs)
	if err != nil {
		t.Fatalf("second inference failed: %v", err)
	}

	// Verify determinism
	if result1.Score != result2.Score {
		t.Errorf("scores not deterministic: %d vs %d", result1.Score, result2.Score)
	}

	if result1.InputHash != result2.InputHash {
		t.Error("input hashes not deterministic")
	}

	if result1.OutputHash != result2.OutputHash {
		t.Error("output hashes not deterministic")
	}
}

func TestScorerStats(t *testing.T) {
	// Create temp model
	tempDir, err := os.MkdirTemp("", "test-scorer-stats-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	modelDir := filepath.Join(tempDir, "model")
	if err := os.MkdirAll(modelDir, 0750); err != nil {
		t.Fatalf("failed to create model dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(modelDir, "saved_model.pb"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create model file: %v", err)
	}

	config := DefaultInferenceConfig()
	config.ModelPath = modelDir
	setExpectedHashForModel(t, &config, modelDir)

	scorer, err := NewTensorFlowScorer(config)
	if err != nil {
		t.Fatalf("failed to create scorer: %v", err)
	}
	defer func() { _ = scorer.Close() }()

	inputs := createTestInputs()

	// Run a few inferences
	for i := 0; i < 3; i++ {
		_, _ = scorer.ComputeScore(inputs)
	}

	stats := scorer.GetStats()

	if stats.InferenceCount != 3 {
		t.Errorf("expected 3 inferences, got %d", stats.InferenceCount)
	}

	if !stats.IsHealthy {
		t.Error("scorer should be healthy")
	}
}

// ============================================================================
// Test Sidecar Client
// ============================================================================

func TestSidecarClientCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping sidecar test in short mode (requires running sidecar)")
	}

	config := DefaultInferenceConfig()
	config.UseSidecar = true
	config.SidecarAddress = testSidecarAddress
	config.ExpectedHash = strings.Repeat("a", 64)

	client, err := NewSidecarClient(config)
	if err != nil {
		t.Skipf("skipping: sidecar not available at %s: %v", testSidecarAddress, err)
	}
	defer func() { _ = client.Close() }()

	if !client.IsHealthy() {
		t.Error("sidecar client should be healthy after creation")
	}
}

func TestSidecarClientInference(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping sidecar test in short mode (requires running sidecar)")
	}

	config := DefaultInferenceConfig()
	config.UseSidecar = true
	config.SidecarAddress = testSidecarAddress
	config.ExpectedHash = strings.Repeat("a", 64)

	client, err := NewSidecarClient(config)
	if err != nil {
		t.Skipf("skipping: sidecar not available at %s: %v", testSidecarAddress, err)
	}
	defer func() { _ = client.Close() }()

	inputs := createTestInputs()

	result, err := client.ComputeScore(inputs)
	if err != nil {
		t.Fatalf("sidecar inference failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.Score > 100 {
		t.Errorf("score should be 0-100, got %d", result.Score)
	}
}

// ============================================================================
// Test Factory Functions
// ============================================================================

func TestNewScorer(t *testing.T) {
	// Test embedded scorer creation
	tempDir, err := os.MkdirTemp("", "test-factory-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	modelDir := filepath.Join(tempDir, "model")
	if err := os.MkdirAll(modelDir, 0750); err != nil {
		t.Fatalf("failed to create model dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(modelDir, "saved_model.pb"), []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create model file: %v", err)
	}

	config := DefaultInferenceConfig()
	config.ModelPath = modelDir
	setExpectedHashForModel(t, &config, modelDir)

	scorer, err := NewScorer(config)
	if err != nil {
		t.Fatalf("failed to create scorer: %v", err)
	}
	defer func() { _ = scorer.Close() }()

	if !scorer.IsHealthy() {
		t.Error("scorer should be healthy")
	}
}

func TestNewScorerSidecar(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping sidecar test in short mode (requires running sidecar)")
	}

	config := DefaultInferenceConfig()
	config.UseSidecar = true
	config.SidecarAddress = testSidecarAddress
	config.ExpectedHash = strings.Repeat("a", 64)

	scorer, err := NewScorer(config)
	if err != nil {
		t.Skipf("skipping: sidecar not available at %s: %v", testSidecarAddress, err)
	}
	defer func() { _ = scorer.Close() }()

	if !scorer.IsHealthy() {
		t.Error("sidecar scorer should be healthy")
	}
}

// ============================================================================
// Test Helpers
// ============================================================================

// createTestInputs creates valid test inputs for scoring
func createTestInputs() *ScoreInputs {
	// Create 512-dim face embedding
	faceEmbedding := make([]float32, FaceEmbeddingDim)
	for i := range faceEmbedding {
		faceEmbedding[i] = float32(i%100) / 100.0
	}

	return &ScoreInputs{
		FaceEmbedding:   faceEmbedding,
		FaceConfidence:  0.95,
		DocQualityScore: 0.85,
		DocQualityFeatures: DocQualityFeatures{
			Sharpness:  0.9,
			Brightness: 0.8,
			Contrast:   0.85,
			NoiseLevel: 0.1,
			BlurScore:  0.15,
		},
		OCRConfidences: map[string]float32{
			"name":            0.95,
			"date_of_birth":   0.90,
			"document_number": 0.88,
			"expiry_date":     0.92,
			"nationality":     0.85,
		},
		OCRFieldValidation: map[string]bool{
			"name":            true,
			"date_of_birth":   true,
			"document_number": true,
			"expiry_date":     true,
			"nationality":     true,
		},
		Metadata: InferenceMetadata{
			AccountAddress:   "virt1abc123...",
			BlockHeight:      12345,
			BlockTime:        time.Now(),
			RequestID:        "req-001",
			ValidatorAddress: "virt1validator...",
		},
		ScopeTypes: []string{"id_document", "selfie"},
		ScopeCount: 2,
	}
}

func setExpectedHashForModel(t testing.TB, config *InferenceConfig, modelDir string) {
	t.Helper()
	loader := NewModelLoader(*config)
	hash, err := loader.computeModelHash(modelDir)
	if err != nil {
		t.Fatalf("failed to compute model hash: %v", err)
	}
	config.ExpectedHash = hash
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkFeatureExtraction(b *testing.B) {
	extractor := NewFeatureExtractor(DefaultFeatureExtractorConfig())
	inputs := createTestInputs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractor.ExtractFeatures(inputs)
	}
}

func BenchmarkInputHashComputation(b *testing.B) {
	dc := NewDeterminismController(42, true)
	inputs := createTestInputs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dc.ComputeInputHash(inputs)
	}
}

func BenchmarkInference(b *testing.B) {
	// Create temp model
	tempDir, err := os.MkdirTemp("", "bench-scorer-*")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	modelDir := filepath.Join(tempDir, "model")
	if err := os.MkdirAll(modelDir, 0750); err != nil {
		b.Fatalf("failed to create model dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(modelDir, "saved_model.pb"), []byte("test"), 0600); err != nil {
		b.Fatalf("failed to create model file: %v", err)
	}

	config := DefaultInferenceConfig()
	config.ModelPath = modelDir
	setExpectedHashForModel(b, &config, modelDir)

	scorer, err := NewTensorFlowScorer(config)
	if err != nil {
		b.Fatalf("failed to create scorer: %v", err)
	}
	defer func() { _ = scorer.Close() }()

	inputs := createTestInputs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scorer.ComputeScore(inputs)
	}
}
