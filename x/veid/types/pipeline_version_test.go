package types

import (
	"testing"
	"time"
)

// TestPipelineVersionValidation tests PipelineVersion validation
func TestPipelineVersionValidation(t *testing.T) {
	now := time.Now().UTC()

	validManifest := createValidModelManifest(now)

	tests := []struct {
		name        string
		version     *PipelineVersion
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid pipeline version",
			version: &PipelineVersion{
				Version:           "1.0.0",
				ImageHash:         "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
				ImageRef:          "ghcr.io/virtengine/veid-pipeline:v1.0.0",
				ModelManifest:     *validManifest,
				Status:            PipelineVersionStatusPending,
				DeterminismConfig: DefaultPipelineDeterminismConfig(),
			},
			expectError: false,
		},
		{
			name: "empty version",
			version: &PipelineVersion{
				Version:           "",
				ImageHash:         "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
				ImageRef:          "ghcr.io/virtengine/veid-pipeline:v1.0.0",
				ModelManifest:     *validManifest,
				Status:            PipelineVersionStatusPending,
				DeterminismConfig: DefaultPipelineDeterminismConfig(),
			},
			expectError: true,
			errorMsg:    "version cannot be empty",
		},
		{
			name: "invalid version format",
			version: &PipelineVersion{
				Version:           "invalid",
				ImageHash:         "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
				ImageRef:          "ghcr.io/virtengine/veid-pipeline:v1.0.0",
				ModelManifest:     *validManifest,
				Status:            PipelineVersionStatusPending,
				DeterminismConfig: DefaultPipelineDeterminismConfig(),
			},
			expectError: true,
			errorMsg:    "invalid version format",
		},
		{
			name: "empty image hash",
			version: &PipelineVersion{
				Version:           "1.0.0",
				ImageHash:         "",
				ImageRef:          "ghcr.io/virtengine/veid-pipeline:v1.0.0",
				ModelManifest:     *validManifest,
				Status:            PipelineVersionStatusPending,
				DeterminismConfig: DefaultPipelineDeterminismConfig(),
			},
			expectError: true,
			errorMsg:    "image_hash cannot be empty",
		},
		{
			name: "invalid image hash format",
			version: &PipelineVersion{
				Version:           "1.0.0",
				ImageHash:         "invalid-hash",
				ImageRef:          "ghcr.io/virtengine/veid-pipeline:v1.0.0",
				ModelManifest:     *validManifest,
				Status:            PipelineVersionStatusPending,
				DeterminismConfig: DefaultPipelineDeterminismConfig(),
			},
			expectError: true,
			errorMsg:    "invalid image_hash format",
		},
		{
			name: "empty image ref",
			version: &PipelineVersion{
				Version:           "1.0.0",
				ImageHash:         "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
				ImageRef:          "",
				ModelManifest:     *validManifest,
				Status:            PipelineVersionStatusPending,
				DeterminismConfig: DefaultPipelineDeterminismConfig(),
			},
			expectError: true,
			errorMsg:    "image_ref cannot be empty",
		},
		{
			name: "invalid status",
			version: &PipelineVersion{
				Version:           "1.0.0",
				ImageHash:         "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
				ImageRef:          "ghcr.io/virtengine/veid-pipeline:v1.0.0",
				ModelManifest:     *validManifest,
				Status:            "invalid",
				DeterminismConfig: DefaultPipelineDeterminismConfig(),
			},
			expectError: true,
			errorMsg:    "invalid status",
		},
		{
			name: "version with v prefix",
			version: &PipelineVersion{
				Version:           "v1.0.0",
				ImageHash:         "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
				ImageRef:          "ghcr.io/virtengine/veid-pipeline:v1.0.0",
				ModelManifest:     *validManifest,
				Status:            PipelineVersionStatusActive,
				DeterminismConfig: DefaultPipelineDeterminismConfig(),
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.version.Validate()
			if tc.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errorMsg)
				} else if !containsString(err.Error(), tc.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tc.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestModelManifestValidation tests ModelManifest validation
func TestModelManifestValidation(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name        string
		manifest    *ModelManifest
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid manifest",
			manifest:    createValidModelManifest(now),
			expectError: false,
		},
		{
			name: "empty version",
			manifest: &ModelManifest{
				Version: "",
				Models: map[string]ModelInfo{
					"deepface": createValidModelInfo("deepface"),
				},
			},
			expectError: true,
			errorMsg:    "manifest version cannot be empty",
		},
		{
			name: "no models",
			manifest: &ModelManifest{
				Version: "1.0.0",
				Models:  map[string]ModelInfo{},
			},
			expectError: true,
			errorMsg:    "must contain at least one model",
		},
		{
			name: "model name mismatch",
			manifest: &ModelManifest{
				Version: "1.0.0",
				Models: map[string]ModelInfo{
					"deepface": {
						Name:        "wrong_name",
						Version:     "1.0.0",
						WeightsHash: "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
						Framework:   "tensorflow",
						Purpose:     ModelPurposeFaceRecognition,
					},
				},
			},
			expectError: true,
			errorMsg:    "model name mismatch",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.manifest.Validate()
			if tc.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errorMsg)
				} else if !containsString(err.Error(), tc.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tc.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestModelInfoValidation tests ModelInfo validation
func TestModelInfoValidation(t *testing.T) {
	tests := []struct {
		name        string
		model       *ModelInfo
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid model info",
			model: &ModelInfo{
				Name:        "deepface_facenet512",
				Version:     "1.0.0",
				WeightsHash: "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
				Framework:   "tensorflow",
				Purpose:     ModelPurposeFaceRecognition,
			},
			expectError: false,
		},
		{
			name: "empty name",
			model: &ModelInfo{
				Name:        "",
				Version:     "1.0.0",
				WeightsHash: "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
				Framework:   "tensorflow",
			},
			expectError: true,
			errorMsg:    "model name cannot be empty",
		},
		{
			name: "empty version",
			model: &ModelInfo{
				Name:        "deepface",
				Version:     "",
				WeightsHash: "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
				Framework:   "tensorflow",
			},
			expectError: true,
			errorMsg:    "model version cannot be empty",
		},
		{
			name: "empty weights hash",
			model: &ModelInfo{
				Name:        "deepface",
				Version:     "1.0.0",
				WeightsHash: "",
				Framework:   "tensorflow",
			},
			expectError: true,
			errorMsg:    "weights_hash cannot be empty",
		},
		{
			name: "invalid weights hash format",
			model: &ModelInfo{
				Name:        "deepface",
				Version:     "1.0.0",
				WeightsHash: "invalid",
				Framework:   "tensorflow",
			},
			expectError: true,
			errorMsg:    "invalid weights_hash format",
		},
		{
			name: "empty framework",
			model: &ModelInfo{
				Name:        "deepface",
				Version:     "1.0.0",
				WeightsHash: "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
				Framework:   "",
			},
			expectError: true,
			errorMsg:    "framework cannot be empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.model.Validate()
			if tc.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errorMsg)
				} else if !containsString(err.Error(), tc.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tc.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestManifestHashComputation tests deterministic hash computation
func TestManifestHashComputation(t *testing.T) {
	now := time.Now().UTC()

	// Create two identical manifests
	manifest1 := createValidModelManifest(now)
	manifest2 := createValidModelManifest(now)

	// Hashes should be identical
	hash1 := manifest1.ComputeHash()
	hash2 := manifest2.ComputeHash()

	if hash1 != hash2 {
		t.Errorf("identical manifests should produce same hash: %s vs %s", hash1, hash2)
	}

	// Modify one manifest
	manifest2.Models["deepface"] = ModelInfo{
		Name:        "deepface",
		Version:     "1.0.1", // Changed version
		WeightsHash: "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		Framework:   "tensorflow",
		Purpose:     ModelPurposeFaceRecognition,
	}

	hash3 := manifest2.ComputeHash()
	if hash1 == hash3 {
		t.Error("different manifests should produce different hashes")
	}
}

// TestPipelineHashComputation tests pipeline hash computation
func TestPipelineHashComputation(t *testing.T) {
	now := time.Now().UTC()
	manifest := createValidModelManifest(now)

	pv1 := NewPipelineVersion(
		"1.0.0",
		"sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		"ghcr.io/virtengine/veid-pipeline:v1.0.0",
		*manifest,
		now,
		100,
	)

	pv2 := NewPipelineVersion(
		"1.0.0",
		"sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		"ghcr.io/virtengine/veid-pipeline:v1.0.0",
		*manifest,
		now,
		100,
	)

	// Same pipeline versions should have same hash
	hash1 := pv1.ComputePipelineHash()
	hash2 := pv2.ComputePipelineHash()

	if hash1 != hash2 {
		t.Errorf("identical pipelines should produce same hash: %s vs %s", hash1, hash2)
	}

	// Different image hash should produce different pipeline hash
	pv3 := NewPipelineVersion(
		"1.0.0",
		"sha256:different00000000000000000000000000000000000000000000000000000000",
		"ghcr.io/virtengine/veid-pipeline:v1.0.0",
		*manifest,
		now,
		100,
	)

	hash3 := pv3.ComputePipelineHash()
	if hash1 == hash3 {
		t.Error("different image hashes should produce different pipeline hashes")
	}
}

// TestExecutionRecordComparison tests comparing execution records
func TestExecutionRecordComparison(t *testing.T) {
	now := time.Now().UTC()

	record1 := NewPipelineExecutionRecord(
		"1.0.0",
		"sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		"abcdef1234567890",
		now,
	)
	record1.InputHash = "inputhash123"
	record1.OutputHash = "outputhash456"

	record2 := NewPipelineExecutionRecord(
		"1.0.0",
		"sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		"abcdef1234567890",
		now,
	)
	record2.InputHash = "inputhash123"
	record2.OutputHash = "outputhash456"

	// Identical records should match
	result := CompareExecutionRecords(record1, record2)
	if !result.Match {
		t.Errorf("identical records should match, differences: %v", result.Differences)
	}

	// Different output hash should not match
	record2.OutputHash = "differentoutput"
	result = CompareExecutionRecords(record1, record2)
	if result.Match {
		t.Error("records with different output hashes should not match")
	}
	if len(result.Differences) == 0 {
		t.Error("expected differences to be recorded")
	}
}

// TestPipelineVersionStatusValidation tests status validation
func TestPipelineVersionStatusValidation(t *testing.T) {
	validStatuses := AllPipelineVersionStatuses()
	if len(validStatuses) != 4 {
		t.Errorf("expected 4 valid statuses, got %d", len(validStatuses))
	}

	// Test valid statuses
	for _, status := range validStatuses {
		if !IsValidPipelineVersionStatus(status) {
			t.Errorf("expected %s to be valid", status)
		}
	}

	// Test invalid status
	if IsValidPipelineVersionStatus("invalid_status") {
		t.Error("expected 'invalid_status' to be invalid")
	}
}

// TestDefaultDeterminismConfig tests default determinism configuration
func TestDefaultDeterminismConfig(t *testing.T) {
	config := DefaultPipelineDeterminismConfig()

	if config.RandomSeed != 42 {
		t.Errorf("expected RandomSeed 42, got %d", config.RandomSeed)
	}

	if !config.ForceCPU {
		t.Error("expected ForceCPU to be true")
	}

	if !config.SingleThread {
		t.Error("expected SingleThread to be true")
	}

	if config.FloatPrecision != 6 {
		t.Errorf("expected FloatPrecision 6, got %d", config.FloatPrecision)
	}

	if !config.TensorFlowDeterministic {
		t.Error("expected TensorFlowDeterministic to be true")
	}

	if !config.DisableCUDNN {
		t.Error("expected DisableCUDNN to be true")
	}
}

// TestPipelineVersionKeys tests pipeline version key generation
func TestPipelineVersionKeys(t *testing.T) {
	version := "1.0.0"
	key := PipelineVersionKey(version)

	if len(key) == 0 {
		t.Error("expected non-empty key")
	}

	// Key should start with prefix
	if key[0] != PrefixPipelineVersion[0] {
		t.Errorf("key should start with prefix 0x%02x, got 0x%02x", PrefixPipelineVersion[0], key[0])
	}

	// Different versions should have different keys
	key2 := PipelineVersionKey("2.0.0")
	if string(key) == string(key2) {
		t.Error("different versions should have different keys")
	}
}

// TestExecutionRecordKeys tests execution record key generation
func TestExecutionRecordKeys(t *testing.T) {
	requestID := "req-12345"
	key := PipelineExecutionRecordKey(requestID)

	if len(key) == 0 {
		t.Error("expected non-empty key")
	}

	validatorAddr := []byte("validator1")
	valKey := PipelineExecutionByValidatorKey(validatorAddr, requestID)

	if len(valKey) == 0 {
		t.Error("expected non-empty validator key")
	}

	// Different validators should have different keys
	valKey2 := PipelineExecutionByValidatorKey([]byte("validator2"), requestID)
	if string(valKey) == string(valKey2) {
		t.Error("different validators should have different keys")
	}
}

// Helper functions

func createValidModelManifest(createdAt time.Time) *ModelManifest {
	models := map[string]ModelInfo{
		"deepface": createValidModelInfo("deepface"),
		"craft":    createValidModelInfo("craft"),
		"unet":     createValidModelInfo("unet"),
	}
	return NewModelManifest("1.0.0", models, createdAt)
}

func createValidModelInfo(name string) ModelInfo {
	return ModelInfo{
		Name:        name,
		Version:     "1.0.0",
		WeightsHash: "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		Framework:   "tensorflow",
		Purpose:     ModelPurposeFaceRecognition,
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
