package inference

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDeterministicInferenceOutputPinned(t *testing.T) {
	const expectedOutputHash = "80072992f02580f61fc1af5536ce5ff5246746f28b2294b07c4c98057bd4c8fd"

	tempDir, err := os.MkdirTemp("", "test-determinism-hash-*")
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

	inputs := pinnedDeterminismInputs()

	result, err := scorer.ComputeScore(inputs)
	if err != nil {
		t.Fatalf("inference failed: %v", err)
	}

	if result.OutputHash != expectedOutputHash {
		t.Fatalf("output hash mismatch: expected %s, got %s", expectedOutputHash, result.OutputHash)
	}
}

func pinnedDeterminismInputs() *ScoreInputs {
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
			BlockTime:        time.Unix(1700000000, 0),
			RequestID:        "req-001",
			ValidatorAddress: "virt1validator...",
		},
		ScopeTypes: []string{"id_document", "selfie"},
		ScopeCount: 2,
	}
}
