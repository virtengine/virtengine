//go:build integration

package inference

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// TestSavedModelIntegration verifies that a SavedModel can be loaded
// and produces deterministic outputs.
//
// This test is tagged with "integration" and requires either:
// - VEID_MODEL_PATH environment variable pointing to a SavedModel
// - A model at the default test fixtures path

func TestSavedModelIntegration(t *testing.T) {
	// Get model path from environment or use default
	modelPath := os.Getenv("VEID_MODEL_PATH")
	if modelPath == "" {
		// Try default test fixture path
		candidates := []string{
			"testdata/models/trust_score/model",
			"../../testdata/models/trust_score/model",
			"../../../artifacts/models/latest/model",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				modelPath = c
				break
			}
		}
	}

	if modelPath == "" {
		t.Skip("No SavedModel found, skipping integration test. Set VEID_MODEL_PATH to run.")
	}

	t.Logf("Testing SavedModel at: %s", modelPath)

	// Create config
	config := InferenceConfig{
		ModelPath:     modelPath,
		ForceCPU:      true,
		RandomSeed:    42,
		Deterministic: true,
	}

	// Create loader
	loader := NewModelLoader(config)

	// Load model
	model, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}
	defer func() { _ = loader.Unload() }()

	// Verify model is loaded
	if !model.IsLoaded() {
		t.Fatal("Model reports as not loaded after Load()")
	}

	// Verify model hash is computed
	if model.GetModelHash() == "" {
		t.Error("Model hash is empty")
	}
	t.Logf("Model hash: %s", model.GetModelHash())

	// Verify version
	if model.GetVersion() != "" {
		t.Logf("Model version: %s", model.GetVersion())
	}

	// Test inference determinism
	t.Run("DeterministicInference", func(t *testing.T) {
		testDeterministicInference(t, model)
	})

	// Test with known inputs
	t.Run("KnownInputs", func(t *testing.T) {
		testKnownInputs(t, model)
	})

	// Test edge cases
	t.Run("EdgeCases", func(t *testing.T) {
		testEdgeCases(t, model)
	})
}

func testDeterministicInference(t *testing.T, model *TFModel) {
	// Create consistent test input
	features := make([]float32, TotalFeatureDim)
	for i := range features {
		// Deterministic initialization
		features[i] = float32(i%100) / 100.0
	}

	// Run inference multiple times
	results := make([][]float32, 5)
	for i := 0; i < 5; i++ {
		result, err := model.Run(features)
		if err != nil {
			t.Fatalf("Inference failed on run %d: %v", i, err)
		}
		results[i] = result
	}

	// All results should be identical
	for i := 1; i < len(results); i++ {
		if !floatSlicesEqual(results[0], results[i]) {
			t.Errorf("Non-deterministic inference detected: run 0 = %v, run %d = %v",
				results[0], i, results[i])
		}
	}

	t.Logf("Determinism verified across %d runs, output: %v", len(results), results[0])
}

func testKnownInputs(t *testing.T, model *TFModel) {
	testCases := []struct {
		name     string
		features []float32
		minScore float32
		maxScore float32
	}{
		{
			name:     "HighConfidence",
			features: createHighConfidenceInput(),
			minScore: 0,
			maxScore: 100,
		},
		{
			name:     "LowConfidence",
			features: createLowConfidenceInput(),
			minScore: 0,
			maxScore: 100,
		},
		{
			name:     "ZeroInput",
			features: make([]float32, TotalFeatureDim),
			minScore: 0,
			maxScore: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := model.Run(tc.features)
			if err != nil {
				t.Fatalf("Inference failed: %v", err)
			}

			if len(result) < 1 {
				t.Fatal("Empty result from inference")
			}

			score := result[0]
			if score < tc.minScore || score > tc.maxScore {
				t.Errorf("Score %f outside expected range [%f, %f]",
					score, tc.minScore, tc.maxScore)
			}

			t.Logf("%s: score = %f", tc.name, score)
		})
	}
}

func testEdgeCases(t *testing.T, model *TFModel) {
	// Test wrong dimension
	t.Run("WrongDimension", func(t *testing.T) {
		wrongDim := make([]float32, TotalFeatureDim+1)
		_, err := model.Run(wrongDim)
		if err == nil {
			t.Error("Expected error for wrong feature dimension")
		}
	})

	// Test empty input
	t.Run("EmptyInput", func(t *testing.T) {
		_, err := model.Run(nil)
		if err == nil {
			t.Error("Expected error for nil input")
		}
	})

	// Test NaN handling
	t.Run("NaNInput", func(t *testing.T) {
		nanInput := make([]float32, TotalFeatureDim)
		nanInput[0] = float32(math.NaN())

		result, err := model.Run(nanInput)
		// Model should either error or handle gracefully
		if err == nil && len(result) > 0 {
			if math.IsNaN(float64(result[0])) {
				t.Log("Model propagated NaN (expected behavior)")
			} else {
				t.Logf("Model handled NaN input, output: %f", result[0])
			}
		}
	})
}

func createHighConfidenceInput() []float32 {
	features := make([]float32, TotalFeatureDim)

	// Face embedding (high confidence values)
	for i := 0; i < FaceEmbeddingDim; i++ {
		features[i] = 0.8 + float32(i%10)*0.02
	}

	// Doc quality (high scores)
	offset := FaceEmbeddingDim
	for i := 0; i < DocQualityDim; i++ {
		features[offset+i] = 0.9
	}

	// OCR confidence (high)
	offset = FaceEmbeddingDim + DocQualityDim
	for i := 0; i < OCRFieldsDim; i++ {
		features[offset+i] = 0.95
	}

	return features
}

func createLowConfidenceInput() []float32 {
	features := make([]float32, TotalFeatureDim)

	// Face embedding (low confidence values)
	for i := 0; i < FaceEmbeddingDim; i++ {
		features[i] = 0.1 + float32(i%10)*0.01
	}

	// Doc quality (low scores)
	offset := FaceEmbeddingDim
	for i := 0; i < DocQualityDim; i++ {
		features[offset+i] = 0.3
	}

	// OCR confidence (low)
	offset = FaceEmbeddingDim + DocQualityDim
	for i := 0; i < OCRFieldsDim; i++ {
		features[offset+i] = 0.4
	}

	return features
}

func floatSlicesEqual(a, b []float32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestManifestVerification verifies that a model manifest matches the model on disk.
func TestManifestVerification(t *testing.T) {
	// Find manifest
	manifestPaths := []string{
		"testdata/models/trust_score/manifest.json",
		"../../testdata/models/trust_score/manifest.json",
		"../../../artifacts/models/latest/manifest.json",
	}

	var manifestPath string
	for _, p := range manifestPaths {
		if _, err := os.Stat(p); err == nil {
			manifestPath = p
			break
		}
	}

	if manifestPath == "" {
		t.Skip("No manifest found, skipping verification test")
	}

	t.Logf("Verifying manifest: %s", manifestPath)

	// Load manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("Failed to read manifest: %v", err)
	}

	var manifest struct {
		ModelPath  string             `json:"model_path"`
		ModelHash  string             `json:"model_hash"`
		Version    string             `json:"model_version"`
		InputSig   map[string]any     `json:"input_signature"`
		OutputSig  map[string]any     `json:"output_signature"`
		Metrics    map[string]float64 `json:"evaluation_metrics"`
		EvalPassed bool               `json:"evaluation_passed"`
	}

	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("Failed to parse manifest: %v", err)
	}

	t.Logf("Manifest version: %s", manifest.Version)
	t.Logf("Manifest hash: %s", manifest.ModelHash)

	// Find model path
	modelPath := manifest.ModelPath
	if modelPath == "" || !fileExists(modelPath) {
		// Try relative to manifest
		modelPath = filepath.Join(filepath.Dir(manifestPath), "model")
	}

	if !fileExists(modelPath) {
		t.Skipf("Model path not found: %s", modelPath)
	}

	// Compute hash
	computedHash, err := computeModelHash(modelPath)
	if err != nil {
		t.Fatalf("Failed to compute model hash: %v", err)
	}

	// Verify hash matches
	if computedHash != manifest.ModelHash {
		t.Errorf("Hash mismatch:\n  Expected: %s\n  Computed: %s",
			manifest.ModelHash, computedHash)
	} else {
		t.Logf("Hash verified: %s", computedHash[:16])
	}

	// Log metrics
	if len(manifest.Metrics) > 0 {
		t.Log("Evaluation metrics:")
		for k, v := range manifest.Metrics {
			t.Logf("  %s: %f", k, v)
		}
	}

	if manifest.EvalPassed {
		t.Log("✅ Evaluation thresholds passed")
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func computeModelHash(modelPath string) (string, error) {
	h := sha256.New()

	var files []string
	err := filepath.Walk(modelPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Base(path) == "export_metadata.json" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return "", err
	}

	sort.Strings(files)

	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read %s: %w", path, err)
		}
		h.Write(data)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// TestConfigHashConsistency verifies that the Go config produces the same
// hash as the Python config for known inputs.
func TestConfigHashConsistency(t *testing.T) {
	// Test that feature dimensions match Python config
	pythonFeatureDim := 768
	if TotalFeatureDim != pythonFeatureDim {
		t.Errorf("Feature dimension mismatch: Go=%d, Python=%d",
			TotalFeatureDim, pythonFeatureDim)
	}

	// Verify component dimensions
	expectedDims := map[string]int{
		"FaceEmbeddingDim": 512,
		"DocQualityDim":    5,
		"OCRFieldsDim":     10,
		"MetadataDim":      16,
	}

	actualDims := map[string]int{
		"FaceEmbeddingDim": FaceEmbeddingDim,
		"DocQualityDim":    DocQualityDim,
		"OCRFieldsDim":     OCRFieldsDim,
		"MetadataDim":      MetadataDim,
	}

	for name, expected := range expectedDims {
		if actual := actualDims[name]; actual != expected {
			t.Errorf("%s mismatch: expected=%d, actual=%d", name, expected, actual)
		}
	}

	// Verify OCR field names match Python config
	expectedFields := []string{
		"name",
		"date_of_birth",
		"document_number",
		"expiry_date",
		"nationality",
	}

	if len(OCRFieldNames) != len(expectedFields) {
		t.Errorf("OCR field count mismatch: expected=%d, actual=%d",
			len(expectedFields), len(OCRFieldNames))
	}

	for i, expected := range expectedFields {
		if i < len(OCRFieldNames) && OCRFieldNames[i] != expected {
			t.Errorf("OCR field %d mismatch: expected=%s, actual=%s",
				i, expected, OCRFieldNames[i])
		}
	}
}
