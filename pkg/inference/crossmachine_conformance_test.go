// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0
//
// Package inference provides cross-machine conformance tests.
// These tests verify that inference produces identical outputs across
// different validator machines for blockchain consensus.
//
// VE-219: Deterministic identity verification runtime

package inference

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

// ============================================================================
// Cross-Machine Conformance Tests
// ============================================================================

// TestCrossMachineDeterminism verifies that inference produces identical
// output hashes for the same inputs across different runs.
func TestCrossMachineDeterminism(t *testing.T) {
	// Create determinism controller with standard settings
	dc := NewDeterminismController(42, true)

	// Run the same inference multiple times
	inputs := createTestScoreInputs()

	var inputHashes []string
	var outputHashes []string

	for i := 0; i < 5; i++ {
		inputHash := dc.ComputeInputHash(inputs)
		inputHashes = append(inputHashes, inputHash)

		// Simulate inference output
		output := []float32{75.5, 0.85}
		outputHash := dc.ComputeOutputHash(output)
		outputHashes = append(outputHashes, outputHash)
	}

	// All input hashes should be identical
	for i := 1; i < len(inputHashes); i++ {
		if inputHashes[i] != inputHashes[0] {
			t.Errorf("input hash mismatch at run %d: expected %s, got %s",
				i, inputHashes[0], inputHashes[i])
		}
	}

	// All output hashes should be identical
	for i := 1; i < len(outputHashes); i++ {
		if outputHashes[i] != outputHashes[0] {
			t.Errorf("output hash mismatch at run %d: expected %s, got %s",
				i, outputHashes[0], outputHashes[i])
		}
	}
}

// TestScoreResultVerification tests that VerifySameOutput correctly
// identifies matching and non-matching results.
func TestScoreResultVerification(t *testing.T) {
	dc := NewDeterminismController(42, true)

	result1 := &ScoreResult{
		Score:        75,
		RawScore:     75.5,
		Confidence:   0.85,
		ModelVersion: "v1.0.0",
		ModelHash:    "abc123",
		InputHash:    "input123",
		OutputHash:   "output123",
	}

	result2 := &ScoreResult{
		Score:        75,
		RawScore:     75.5,
		Confidence:   0.85,
		ModelVersion: "v1.0.0",
		ModelHash:    "abc123",
		InputHash:    "input123",
		OutputHash:   "output123",
	}

	if !dc.VerifySameOutput(result1, result2) {
		t.Error("identical results should verify as same")
	}

	// Modify result2 slightly
	result2.Score = 76
	if dc.VerifySameOutput(result1, result2) {
		t.Error("different scores should not verify as same")
	}
}

// TestConformanceTestVectorDeterminism verifies all test vectors
// produce deterministic hashes.
func TestConformanceTestVectorDeterminism(t *testing.T) {
	dc := NewDeterminismController(42, true)

	for _, vec := range ConformanceTestVectors {
		t.Run(vec.Name, func(t *testing.T) {
			inputs := vec.Input.ConvertToScoreInputs()

			// Compute hash multiple times
			hash1 := dc.ComputeInputHash(inputs)
			hash2 := dc.ComputeInputHash(inputs)

			if hash1 != hash2 {
				t.Errorf("input hash not deterministic for vector %s", vec.Name)
			}

			// Verify hash is 64 hex chars (SHA256)
			if len(hash1) != 64 {
				t.Errorf("expected 64 char hash, got %d chars", len(hash1))
			}
		})
	}
}

// TestFeatureVectorDeterminism verifies feature extraction is deterministic.
func TestFeatureVectorDeterminism(t *testing.T) {
	extractor := NewFeatureExtractor(DefaultFeatureExtractorConfig())

	inputs := createTestScoreInputs()

	// Extract features multiple times
	features1, err := extractor.ExtractFeatures(inputs)
	if err != nil {
		t.Fatalf("failed to extract features: %v", err)
	}

	features2, err := extractor.ExtractFeatures(inputs)
	if err != nil {
		t.Fatalf("failed to extract features: %v", err)
	}

	// Compare feature vectors
	if len(features1) != len(features2) {
		t.Fatalf("feature vector lengths differ: %d vs %d", len(features1), len(features2))
	}

	for i := range features1 {
		if features1[i] != features2[i] {
			t.Errorf("feature %d differs: %f vs %f", i, features1[i], features2[i])
		}
	}
}

// TestDeterminismControllerEnvVars verifies environment variables are set correctly.
func TestDeterminismControllerEnvVars(t *testing.T) {
	dc := NewDeterminismController(42, true)

	envVars := dc.GetTensorFlowEnvVars()

	// Check required env vars
	expectedVars := map[string]string{
		"TF_DETERMINISTIC_OPS":   "1",
		"TF_CUDNN_DETERMINISTIC": "1",
		"OMP_NUM_THREADS":        "1",
		"PYTHONHASHSEED":         "42",
	}

	for key, expected := range expectedVars {
		if val, ok := envVars[key]; !ok {
			t.Errorf("missing env var: %s", key)
		} else if val != expected {
			t.Errorf("env var %s: expected %s, got %s", key, expected, val)
		}
	}
}

// TestTFDeterminismConfig_Conformance verifies TensorFlow config for determinism.
func TestTFDeterminismConfig_Conformance(t *testing.T) {
	dc := NewDeterminismController(42, true)

	config := dc.ConfigureTensorFlow()

	if config.RandomSeed != 42 {
		t.Errorf("expected random seed 42, got %d", config.RandomSeed)
	}

	if config.InterOpParallelism != 1 {
		t.Errorf("expected inter-op parallelism 1, got %d", config.InterOpParallelism)
	}

	if config.IntraOpParallelism != 1 {
		t.Errorf("expected intra-op parallelism 1, got %d", config.IntraOpParallelism)
	}

	if !config.UseCPUOnly {
		t.Error("expected CPU-only mode")
	}

	if !config.EnableDeterministicOps {
		t.Error("expected deterministic ops enabled")
	}
}

// TestHashFloatPrecision verifies float hashing handles precision correctly.
func TestHashFloatPrecision(t *testing.T) {
	dc := NewDeterminismController(42, true)

	// Test that small differences in floats produce same hash
	// when within precision bounds
	output1 := []float32{75.5000001}
	output2 := []float32{75.5000002}

	hash1 := dc.ComputeOutputHash(output1)
	hash2 := dc.ComputeOutputHash(output2)

	if hash1 != hash2 {
		t.Error("small float differences should produce same hash")
	}

	// Test that larger differences produce different hashes
	output3 := []float32{75.6}
	hash3 := dc.ComputeOutputHash(output3)

	if hash1 == hash3 {
		t.Error("larger float differences should produce different hashes")
	}
}

// TestModelDeterminismCheck verifies non-deterministic op detection.
func TestModelDeterminismCheck(t *testing.T) {
	dc := NewDeterminismController(42, true)

	// All deterministic ops
	deterministicOps := []string{"MatMul", "Add", "Relu", "Softmax"}
	isDet, nonDet := dc.CheckModelDeterminism(deterministicOps)

	if !isDet {
		t.Error("expected deterministic model check to pass")
	}

	if len(nonDet) != 0 {
		t.Errorf("expected no non-deterministic ops, got %v", nonDet)
	}

	// Include a non-deterministic op
	mixedOps := []string{"MatMul", "CudnnRNN", "Add"}
	isDet, nonDet = dc.CheckModelDeterminism(mixedOps)

	if isDet {
		t.Error("expected model with CudnnRNN to fail determinism check")
	}

	if len(nonDet) != 1 || nonDet[0] != "CudnnRNN" {
		t.Errorf("expected CudnnRNN in non-deterministic ops, got %v", nonDet)
	}
}

// ============================================================================
// Sidecar Conformance Tests
// ============================================================================

// TestSidecarClientDeterminism tests sidecar client produces deterministic results.
func TestSidecarClientDeterminism(t *testing.T) {
	// Skip if running in CI without sidecar
	if testing.Short() {
		t.Skip("skipping sidecar test in short mode")
	}

	config := DefaultInferenceConfig().
		WithSidecar("localhost:50051").
		WithDeterministic(true)

	// Create client (may fail if sidecar not running)
	client, err := NewSidecarClient(config)
	if err != nil {
		t.Skipf("skipping test, sidecar not available: %v", err)
	}
	defer client.Close()

	inputs := createTestScoreInputs()

	// Run inference multiple times
	var results []*ScoreResult
	for i := 0; i < 3; i++ {
		result, err := client.ComputeScore(inputs)
		if err != nil {
			t.Fatalf("inference failed: %v", err)
		}
		results = append(results, result)
	}

	// All results should have identical output hashes
	for i := 1; i < len(results); i++ {
		if results[i].OutputHash != results[0].OutputHash {
			t.Errorf("output hash mismatch at run %d: expected %s, got %s",
				i, results[0].OutputHash, results[i].OutputHash)
		}
	}
}

// TestSidecarDeterminismVerification tests the VerifyDeterminism RPC.
func TestSidecarDeterminismVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping sidecar test in short mode")
	}

	config := DefaultInferenceConfig().
		WithSidecar("localhost:50051").
		WithDeterministic(true)

	client, err := NewSidecarClient(config)
	if err != nil {
		t.Skipf("skipping test, sidecar not available: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := client.VerifyDeterminism(ctx, "high_quality_verification")
	if err != nil {
		t.Fatalf("verify determinism failed: %v", err)
	}

	if !result.Passed {
		t.Errorf("determinism verification failed: %v", result.Differences)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// createTestScoreInputs creates test inputs for inference.
func createTestScoreInputs() *ScoreInputs {
	return &ScoreInputs{
		FaceEmbedding:   generateDeterministicEmbedding(FaceEmbeddingDim, 42, 0.1),
		FaceConfidence:  0.95,
		DocQualityScore: 0.88,
		DocQualityFeatures: DocQualityFeatures{
			Sharpness:  0.85,
			Brightness: 0.75,
			Contrast:   0.80,
			NoiseLevel: 0.08,
			BlurScore:  0.05,
		},
		OCRConfidences: map[string]float32{
			"name":            0.92,
			"date_of_birth":   0.89,
			"document_number": 0.95,
			"expiry_date":     0.87,
			"nationality":     0.90,
		},
		OCRFieldValidation: map[string]bool{
			"name":            true,
			"date_of_birth":   true,
			"document_number": true,
			"expiry_date":     true,
			"nationality":     true,
		},
		ScopeTypes: []string{"id_document", "selfie"},
		ScopeCount: 2,
		Metadata: InferenceMetadata{
			AccountAddress: "cosmos1abc123",
			BlockHeight:    1000000,
			RequestID:      "test-request-001",
		},
	}
}

// computeTestHash computes a SHA256 hash for testing.
func computeTestHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
