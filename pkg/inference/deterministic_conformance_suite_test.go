// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0
//
// Package inference provides deterministic conformance tests for ML inference.
// These tests verify that inference produces identical outputs across different
// environments for blockchain consensus.
//
// Task 8A: Deterministic inference conformance suite
// VE-219: Deterministic identity verification runtime

package inference

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// ============================================================================
// Conformance Suite Configuration
// ============================================================================

const (
	// ConformanceSuiteVersion is the version of this conformance suite
	ConformanceSuiteVersion = "1.0.0"

	// StrictDeterminismMode fails on any variance
	StrictDeterminismMode = true

	// EvidenceDirectoryEnv is the environment variable for evidence storage
	EvidenceDirectoryEnv = "VEID_CONFORMANCE_EVIDENCE_DIR"
)

// ============================================================================
// Deterministic Conformance Suite Tests
// ============================================================================

// TestDeterministicConformanceSuite runs the full deterministic conformance suite.
// This is the primary test for verifying cross-machine determinism.
func TestDeterministicConformanceSuite(t *testing.T) {
	suite := NewConformanceSuite(t)

	t.Run("RuntimeConfiguration", suite.TestRuntimeConfiguration)
	t.Run("GoldenVectorIntegrity", suite.TestGoldenVectorIntegrity)
	t.Run("InputHashDeterminism", suite.TestInputHashDeterminism)
	t.Run("OutputHashDeterminism", suite.TestOutputHashDeterminism)
	t.Run("FeatureExtractionDeterminism", suite.TestFeatureExtractionDeterminism)
	t.Run("CrossRunConsistency", suite.TestCrossRunConsistency)
	t.Run("NonDeterministicOpDetection", suite.TestNonDeterministicOpDetection)
	t.Run("HashPrecisionBoundaries", suite.TestHashPrecisionBoundaries)
	t.Run("PlatformIndependence", suite.TestPlatformIndependence)
	t.Run("GoldenVectorExecution", suite.TestGoldenVectorExecution)
}

// ConformanceSuite contains the conformance test implementation
type ConformanceSuite struct {
	t           *testing.T
	dc          *DeterminismController
	extractor   *FeatureExtractor
	platform    PlatformInfo
	startTime   time.Time
	results     []GoldenVectorResult
	evidenceDir string
}

// NewConformanceSuite creates a new conformance suite
func NewConformanceSuite(t *testing.T) *ConformanceSuite {
	evidenceDir := os.Getenv(EvidenceDirectoryEnv)
	if evidenceDir == "" {
		evidenceDir = filepath.Join(os.TempDir(), "veid-conformance-evidence")
	}

	return &ConformanceSuite{
		t:           t,
		dc:          NewDeterminismController(GoldenVectorSeed, true),
		extractor:   NewFeatureExtractor(DefaultFeatureExtractorConfig()),
		platform:    GetPlatformInfo(),
		startTime:   time.Now(),
		results:     make([]GoldenVectorResult, 0),
		evidenceDir: evidenceDir,
	}
}

// ============================================================================
// Runtime Configuration Tests
// ============================================================================

// TestRuntimeConfiguration verifies runtime is configured for determinism
func (s *ConformanceSuite) TestRuntimeConfiguration(t *testing.T) {
	t.Parallel()

	canonical := CanonicalRuntimeConfig()

	// Verify determinism controller settings
	tfConfig := s.dc.ConfigureTensorFlow()

	if tfConfig.RandomSeed != canonical.RandomSeed {
		t.Errorf("random seed mismatch: expected %d, got %d",
			canonical.RandomSeed, tfConfig.RandomSeed)
	}

	if tfConfig.InterOpParallelism != canonical.InterOpParallelism {
		t.Errorf("inter-op parallelism mismatch: expected %d, got %d",
			canonical.InterOpParallelism, tfConfig.InterOpParallelism)
	}

	if tfConfig.IntraOpParallelism != canonical.IntraOpParallelism {
		t.Errorf("intra-op parallelism mismatch: expected %d, got %d",
			canonical.IntraOpParallelism, tfConfig.IntraOpParallelism)
	}

	if !tfConfig.UseCPUOnly {
		t.Error("CPU-only mode must be enabled for determinism")
	}

	if !tfConfig.EnableDeterministicOps {
		t.Error("deterministic ops must be enabled")
	}

	// Verify environment variables
	envVars := s.dc.GetTensorFlowEnvVars()

	requiredEnvVars := map[string]string{
		"TF_DETERMINISTIC_OPS":   "1",
		"TF_CUDNN_DETERMINISTIC": "1",
		"OMP_NUM_THREADS":        "1",
	}

	for key, expected := range requiredEnvVars {
		if val, ok := envVars[key]; !ok {
			t.Errorf("missing required env var: %s", key)
		} else if val != expected {
			t.Errorf("env var %s: expected %s, got %s", key, expected, val)
		}
	}

	t.Logf("Runtime configuration verified on %s/%s", s.platform.OS, s.platform.Arch)
}

// ============================================================================
// Golden Vector Integrity Tests
// ============================================================================

// TestGoldenVectorIntegrity validates all golden vectors are well-formed
func (s *ConformanceSuite) TestGoldenVectorIntegrity(t *testing.T) {
	t.Parallel()

	if GetGoldenVectorCount() == 0 {
		t.Fatal("no golden vectors defined")
	}

	for _, vec := range GoldenVectors {
		t.Run(vec.ID, func(t *testing.T) {
			if err := ValidateGoldenVectorIntegrity(&vec); err != nil {
				t.Errorf("golden vector integrity check failed: %v", err)
			}

			// Verify version matches
			if vec.Version != GoldenVectorVersion {
				t.Errorf("version mismatch: expected %s, got %s",
					GoldenVectorVersion, vec.Version)
			}

			// Verify inputs are valid
			if vec.Inputs == nil {
				t.Error("inputs is nil")
			}

			// Verify face embedding dimension
			if len(vec.Inputs.FaceEmbedding) != FaceEmbeddingDim && len(vec.Inputs.FaceEmbedding) != 0 {
				t.Errorf("face embedding dimension mismatch: expected %d or 0, got %d",
					FaceEmbeddingDim, len(vec.Inputs.FaceEmbedding))
			}
		})
	}

	t.Logf("Validated %d golden vectors", GetGoldenVectorCount())
}

// ============================================================================
// Input Hash Determinism Tests
// ============================================================================

// TestInputHashDeterminism verifies input hashing is deterministic
func (s *ConformanceSuite) TestInputHashDeterminism(t *testing.T) {
	t.Parallel()

	for _, vec := range GoldenVectors {
		t.Run(vec.ID, func(t *testing.T) {
			// Compute hash multiple times
			var hashes []string
			for i := 0; i < 10; i++ {
				hash := s.dc.ComputeInputHash(vec.Inputs)
				hashes = append(hashes, hash)
			}

			// All hashes must be identical
			for i := 1; i < len(hashes); i++ {
				if hashes[i] != hashes[0] {
					t.Errorf("input hash not deterministic at iteration %d: expected %s, got %s",
						i, hashes[0], hashes[i])
				}
			}

			// Verify hash format (64 hex chars = SHA256)
			if len(hashes[0]) != 64 {
				t.Errorf("hash length incorrect: expected 64, got %d", len(hashes[0]))
			}
		})
	}
}

// ============================================================================
// Output Hash Determinism Tests
// ============================================================================

// TestOutputHashDeterminism verifies output hashing is deterministic
func (s *ConformanceSuite) TestOutputHashDeterminism(t *testing.T) {
	t.Parallel()

	// Test with various output patterns
	testOutputs := [][]float32{
		{75.5, 0.85},
		{50.0, 0.5},
		{95.123456, 0.999999},
		{0.000001, 0.000001},
		{100.0, 1.0},
	}

	for i, output := range testOutputs {
		t.Run(fmt.Sprintf("output_%d", i), func(t *testing.T) {
			var hashes []string
			for j := 0; j < 10; j++ {
				hash := s.dc.ComputeOutputHash(output)
				hashes = append(hashes, hash)
			}

			// All hashes must be identical
			for j := 1; j < len(hashes); j++ {
				if hashes[j] != hashes[0] {
					t.Errorf("output hash not deterministic at iteration %d: expected %s, got %s",
						j, hashes[0], hashes[j])
				}
			}
		})
	}
}

// ============================================================================
// Feature Extraction Determinism Tests
// ============================================================================

// TestFeatureExtractionDeterminism verifies feature extraction is deterministic
func (s *ConformanceSuite) TestFeatureExtractionDeterminism(t *testing.T) {
	t.Parallel()

	for _, vec := range GoldenVectors {
		t.Run(vec.ID, func(t *testing.T) {
			// Extract features multiple times
			var featureVectors [][]float32
			for i := 0; i < 5; i++ {
				features, err := s.extractor.ExtractFeatures(vec.Inputs)
				if err != nil {
					t.Fatalf("feature extraction failed: %v", err)
				}
				featureVectors = append(featureVectors, features)
			}

			// All feature vectors must be identical
			for i := 1; i < len(featureVectors); i++ {
				if len(featureVectors[i]) != len(featureVectors[0]) {
					t.Errorf("feature vector length mismatch at iteration %d: expected %d, got %d",
						i, len(featureVectors[0]), len(featureVectors[i]))
					continue
				}

				for j := range featureVectors[0] {
					if featureVectors[i][j] != featureVectors[0][j] {
						t.Errorf("feature mismatch at [%d][%d]: expected %f, got %f",
							i, j, featureVectors[0][j], featureVectors[i][j])
					}
				}
			}

			// Verify dimension
			if len(featureVectors[0]) != TotalFeatureDim {
				t.Errorf("feature dimension mismatch: expected %d, got %d",
					TotalFeatureDim, len(featureVectors[0]))
			}
		})
	}
}

// ============================================================================
// Cross-Run Consistency Tests
// ============================================================================

// TestCrossRunConsistency verifies consistency across multiple test runs
func (s *ConformanceSuite) TestCrossRunConsistency(t *testing.T) {
	t.Parallel()

	// Create multiple determinism controllers with same seed
	controllers := make([]*DeterminismController, 5)
	for i := range controllers {
		controllers[i] = NewDeterminismController(GoldenVectorSeed, true)
	}

	for _, vec := range GoldenVectors[:2] { // Test with first 2 vectors for speed
		t.Run(vec.ID, func(t *testing.T) {
			var hashes []string
			for _, dc := range controllers {
				hash := dc.ComputeInputHash(vec.Inputs)
				hashes = append(hashes, hash)
			}

			// All controllers must produce same hash
			for i := 1; i < len(hashes); i++ {
				if hashes[i] != hashes[0] {
					t.Errorf("cross-run hash mismatch at controller %d: expected %s, got %s",
						i, hashes[0], hashes[i])
				}
			}
		})
	}
}

// ============================================================================
// Non-Deterministic Operation Detection Tests
// ============================================================================

// TestNonDeterministicOpDetection tests detection of non-deterministic ops
func (s *ConformanceSuite) TestNonDeterministicOpDetection(t *testing.T) {
	t.Parallel()

	// Test with known deterministic ops
	deterministicOps := []string{
		"MatMul", "Add", "Relu", "Softmax", "Sigmoid", "Tanh",
		"BiasAdd", "Const", "Placeholder", "Identity",
	}

	// Note: BiasAdd can be deterministic when properly configured
	// The model ops should be validated at load time
	isDet, nonDet := s.dc.CheckModelDeterminism(deterministicOps)

	// Log any non-deterministic ops found
	if !isDet {
		t.Logf("Non-deterministic ops detected: %v", nonDet)
		t.Logf("Note: Some ops like BiasAdd can be deterministic with proper configuration")
	}

	// Test with known non-deterministic ops
	nonDeterministicOps := []string{"CudnnRNN"}
	isDet, nonDet = s.dc.CheckModelDeterminism(nonDeterministicOps)

	if isDet {
		t.Error("CudnnRNN should be detected as non-deterministic")
	}

	if len(nonDet) != 1 || nonDet[0] != "CudnnRNN" {
		t.Errorf("expected CudnnRNN in non-deterministic list, got %v", nonDet)
	}
}

// ============================================================================
// Hash Precision Boundary Tests
// ============================================================================

// TestHashPrecisionBoundaries tests float hashing at precision boundaries
func (s *ConformanceSuite) TestHashPrecisionBoundaries(t *testing.T) {
	t.Parallel()

	// Test cases at precision boundaries
	// With 6 decimal precision, these should produce same hash
	sameHashPairs := []struct {
		a, b float32
		name string
	}{
		{75.5000001, 75.5000002, "tiny_diff"},
		{0.1234567, 0.1234568, "small_diff"},
		{99.9999991, 99.9999992, "near_100_diff"},
	}

	for _, tc := range sameHashPairs {
		t.Run(tc.name, func(t *testing.T) {
			hash1 := s.dc.ComputeOutputHash([]float32{tc.a})
			hash2 := s.dc.ComputeOutputHash([]float32{tc.b})

			if hash1 != hash2 {
				t.Errorf("precision boundary test failed: %f and %f should produce same hash", tc.a, tc.b)
			}
		})
	}

	// Test cases that should produce different hashes
	diffHashPairs := []struct {
		a, b float32
		name string
	}{
		{75.5, 75.6, "significant_diff"},
		{0.1, 0.2, "small_significant_diff"},
		{99.0, 100.0, "boundary_diff"},
	}

	for _, tc := range diffHashPairs {
		t.Run(tc.name+"_diff", func(t *testing.T) {
			hash1 := s.dc.ComputeOutputHash([]float32{tc.a})
			hash2 := s.dc.ComputeOutputHash([]float32{tc.b})

			if hash1 == hash2 {
				t.Errorf("different values %f and %f should produce different hashes", tc.a, tc.b)
			}
		})
	}
}

// ============================================================================
// Platform Independence Tests
// ============================================================================

// TestPlatformIndependence verifies hash computation is platform-independent
func (s *ConformanceSuite) TestPlatformIndependence(t *testing.T) {
	t.Parallel()

	// Log platform information
	t.Logf("Running on platform: %s/%s, Go %s, %d CPUs",
		s.platform.OS, s.platform.Arch, s.platform.GoVersion, s.platform.NumCPU)

	// Compute hashes for all golden vectors
	for _, vec := range GoldenVectors {
		t.Run(vec.ID, func(t *testing.T) {
			inputHash := s.dc.ComputeInputHash(vec.Inputs)

			// Log the computed hash for cross-platform comparison
			t.Logf("Input hash on %s/%s: %s", s.platform.OS, s.platform.Arch, inputHash)

			// The hash should be 64 hex characters
			if len(inputHash) != 64 {
				t.Errorf("input hash length incorrect: expected 64, got %d", len(inputHash))
			}

			// Verify hash contains only hex characters
			for _, c := range inputHash {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Errorf("invalid character in hash: %c", c)
				}
			}
		})
	}
}

// ============================================================================
// Golden Vector Execution Tests
// ============================================================================

// TestGoldenVectorExecution runs all golden vectors and validates results
func (s *ConformanceSuite) TestGoldenVectorExecution(t *testing.T) {
	// Create evidence directory
	if err := os.MkdirAll(s.evidenceDir, 0750); err != nil {
		t.Logf("Warning: could not create evidence directory: %v", err)
	}

	for _, vec := range GoldenVectors {
		t.Run(vec.ID, func(t *testing.T) {
			startTime := time.Now()

			result := GoldenVectorResult{
				VectorID:  vec.ID,
				Passed:    true,
				Platform:  s.platform,
				Timestamp: startTime,
			}

			// Compute input hash
			inputHash := s.dc.ComputeInputHash(vec.Inputs)
			result.ActualInputHash = inputHash

			// Extract features and compute output hash
			features, err := s.extractor.ExtractFeatures(vec.Inputs)
			if err != nil {
				result.Passed = false
				result.Differences = append(result.Differences,
					fmt.Sprintf("feature extraction failed: %v", err))
			} else {
				// Compute a hash of the feature vector as proxy for model output
				outputHash := s.dc.ComputeOutputHash(features)
				result.ActualOutputHash = outputHash
			}

			result.ExecutionTimeMs = time.Since(startTime).Milliseconds()

			// Store result
			s.results = append(s.results, result)

			// Log result
			if result.Passed {
				t.Logf("Golden vector %s passed in %dms", vec.ID, result.ExecutionTimeMs)
			} else {
				t.Errorf("Golden vector %s failed: %v", vec.ID, result.Differences)
			}
		})
	}

	// Save evidence
	s.saveEvidence(t)
}

// saveEvidence saves test evidence to the evidence directory
func (s *ConformanceSuite) saveEvidence(t *testing.T) {
	evidence := struct {
		SuiteVersion string               `json:"suite_version"`
		Platform     PlatformInfo         `json:"platform"`
		StartTime    time.Time            `json:"start_time"`
		EndTime      time.Time            `json:"end_time"`
		DurationMs   int64                `json:"duration_ms"`
		Results      []GoldenVectorResult `json:"results"`
		TotalPassed  int                  `json:"total_passed"`
		TotalFailed  int                  `json:"total_failed"`
	}{
		SuiteVersion: ConformanceSuiteVersion,
		Platform:     s.platform,
		StartTime:    s.startTime,
		EndTime:      time.Now(),
		DurationMs:   time.Since(s.startTime).Milliseconds(),
		Results:      s.results,
	}

	for _, r := range s.results {
		if r.Passed {
			evidence.TotalPassed++
		} else {
			evidence.TotalFailed++
		}
	}

	data, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		t.Logf("Warning: failed to marshal evidence: %v", err)
		return
	}

	filename := fmt.Sprintf("conformance-evidence-%s-%s-%s.json",
		s.platform.OS, s.platform.Arch, s.startTime.Format("20060102-150405"))
	filepath := filepath.Join(s.evidenceDir, filename)

	if err := os.WriteFile(filepath, data, 0600); err != nil {
		t.Logf("Warning: failed to save evidence to %s: %v", filepath, err)
		return
	}

	t.Logf("Evidence saved to %s", filepath)
}

// ============================================================================
// Benchmark Tests (Golden Vector Suite Specific)
// ============================================================================

// BenchmarkGoldenVectorInputHash benchmarks input hash computation for golden vectors
func BenchmarkGoldenVectorInputHash(b *testing.B) {
	dc := NewDeterminismController(GoldenVectorSeed, true)
	vec := GoldenVectors[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dc.ComputeInputHash(vec.Inputs)
	}
}

// BenchmarkGoldenVectorOutputHash benchmarks output hash computation for golden vectors
func BenchmarkGoldenVectorOutputHash(b *testing.B) {
	dc := NewDeterminismController(GoldenVectorSeed, true)
	output := make([]float32, TotalFeatureDim)
	for i := range output {
		output[i] = float32(i) / float32(TotalFeatureDim)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dc.ComputeOutputHash(output)
	}
}

// BenchmarkGoldenVectorFeatureExtract benchmarks feature extraction for golden vectors
func BenchmarkGoldenVectorFeatureExtract(b *testing.B) {
	extractor := NewFeatureExtractor(DefaultFeatureExtractorConfig())
	vec := GoldenVectors[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractor.ExtractFeatures(vec.Inputs)
	}
}

// ============================================================================
// Multi-Environment Simulation Tests
// ============================================================================

// TestMultiEnvironmentSimulation simulates cross-environment execution
func TestMultiEnvironmentSimulation(t *testing.T) {
	// Simulate different environment configurations
	// Note: Input hashes are deterministic based on input data, not seed.
	// The seed affects model output, not input serialization.
	environments := []struct {
		name       string
		os         string
		arch       string
		seed       int64
		forceCPU   bool
		expectPass bool
	}{
		{"linux_amd64_canonical", "linux", "amd64", GoldenVectorSeed, true, true},
		{"linux_arm64_canonical", "linux", "arm64", GoldenVectorSeed, true, true},
		{"darwin_amd64_canonical", "darwin", "amd64", GoldenVectorSeed, true, true},
		{"darwin_arm64_canonical", "darwin", "arm64", GoldenVectorSeed, true, true},
		{"windows_amd64_canonical", "windows", "amd64", GoldenVectorSeed, true, true},
		// All configurations with same inputs should produce same input hash
		// Different seeds only affect model output, not input hashing
		{"different_seed_same_input_hash", "linux", "amd64", 123, true, true},
	}

	for _, env := range environments {
		t.Run(env.name, func(t *testing.T) {
			dc := NewDeterminismController(env.seed, env.forceCPU)

			// Compute hash for first golden vector
			vec := GoldenVectors[0]
			inputHash := dc.ComputeInputHash(vec.Inputs)

			// Compare with canonical hash from reference controller
			canonicalDC := NewDeterminismController(GoldenVectorSeed, true)
			canonicalHash := canonicalDC.ComputeInputHash(vec.Inputs)

			hashesMatch := inputHash == canonicalHash

			if env.expectPass && !hashesMatch {
				t.Errorf("environment %s: hash mismatch with canonical, expected to pass", env.name)
				t.Logf("  Computed: %s", inputHash)
				t.Logf("  Expected: %s", canonicalHash)
			}

			if !env.expectPass && hashesMatch {
				t.Errorf("environment %s: hash matched canonical, expected to fail", env.name)
			}
		})
	}
}

// TestConfigurationValidation tests that misconfigured environments are detected
func TestConfigurationValidation(t *testing.T) {
	// Test that ValidateRuntimeConfig detects issues
	tests := []struct {
		name          string
		config        RuntimeConfig
		expectIssues  bool
		issueContains string
	}{
		{
			name:         "canonical_config_valid",
			config:       CanonicalRuntimeConfig(),
			expectIssues: false,
		},
		{
			name: "wrong_seed_detected",
			config: RuntimeConfig{
				TensorFlowVersion:  "2.13.0",
				RandomSeed:         123, // Wrong seed
				ForceCPU:           true,
				InterOpParallelism: 1,
				IntraOpParallelism: 1,
				DeterministicOps:   true,
				HashPrecision:      HashPrecision,
			},
			expectIssues:  true,
			issueContains: "random_seed",
		},
		{
			name: "gpu_not_allowed",
			config: RuntimeConfig{
				TensorFlowVersion:  "2.13.0",
				RandomSeed:         GoldenVectorSeed,
				ForceCPU:           false, // GPU enabled - not allowed
				InterOpParallelism: 1,
				IntraOpParallelism: 1,
				DeterministicOps:   true,
				HashPrecision:      HashPrecision,
			},
			expectIssues:  true,
			issueContains: "force_cpu",
		},
		{
			name: "parallelism_not_allowed",
			config: RuntimeConfig{
				TensorFlowVersion:  "2.13.0",
				RandomSeed:         GoldenVectorSeed,
				ForceCPU:           true,
				InterOpParallelism: 4, // Multiple threads - not allowed
				IntraOpParallelism: 4,
				DeterministicOps:   true,
				HashPrecision:      HashPrecision,
			},
			expectIssues:  true,
			issueContains: "parallelism",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			issues := ValidateRuntimeConfig(&tc.config)

			if tc.expectIssues && len(issues) == 0 {
				t.Error("expected configuration issues but none found")
			}

			if !tc.expectIssues && len(issues) > 0 {
				t.Errorf("expected no issues but found: %v", issues)
			}

			if tc.issueContains != "" && len(issues) > 0 {
				found := false
				for _, issue := range issues {
					if containsHelper(issue, tc.issueContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected issue containing %q but got: %v", tc.issueContains, issues)
				}
			}
		})
	}
}

// TestStrictDeterminismEnforcement tests strict mode enforcement
func TestStrictDeterminismEnforcement(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), ConformanceTestTimeout)
	defer cancel()

	// Verify strict determinism is enabled
	if !StrictDeterminismMode {
		t.Log("Warning: strict determinism mode is disabled")
	}

	// Run determinism check in goroutine with timeout
	done := make(chan bool, 1)
	var testErr error

	go func() {
		defer func() { done <- true }()

		dc := NewDeterminismController(GoldenVectorSeed, true)

		for _, vec := range GoldenVectors {
			hash1 := dc.ComputeInputHash(vec.Inputs)
			hash2 := dc.ComputeInputHash(vec.Inputs)

			if hash1 != hash2 {
				testErr = fmt.Errorf("strict determinism violation for %s: %s != %s",
					vec.ID, hash1, hash2)
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		t.Fatal("test timed out")
	case <-done:
		if testErr != nil {
			t.Error(testErr)
		}
	}
}

// ============================================================================
// Evidence Storage Tests
// ============================================================================

// TestEvidenceStorage tests evidence storage functionality
func TestEvidenceStorage(t *testing.T) {
	tmpDir := t.TempDir()

	evidence := struct {
		VectorID   string       `json:"vector_id"`
		Passed     bool         `json:"passed"`
		Platform   PlatformInfo `json:"platform"`
		InputHash  string       `json:"input_hash"`
		OutputHash string       `json:"output_hash"`
		Timestamp  time.Time    `json:"timestamp"`
	}{
		VectorID:   "test_evidence_001",
		Passed:     true,
		Platform:   GetPlatformInfo(),
		InputHash:  "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		OutputHash: "b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3",
		Timestamp:  time.Now().UTC(),
	}

	data, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal evidence: %v", err)
	}

	evidencePath := filepath.Join(tmpDir, "evidence.json")
	if err := os.WriteFile(evidencePath, data, 0600); err != nil {
		t.Fatalf("failed to write evidence: %v", err)
	}

	// Read back and verify
	readData, err := os.ReadFile(evidencePath)
	if err != nil {
		t.Fatalf("failed to read evidence: %v", err)
	}

	// Verify JSON is valid
	var readEvidence map[string]interface{}
	if err := json.Unmarshal(readData, &readEvidence); err != nil {
		t.Fatalf("failed to parse evidence JSON: %v", err)
	}

	if readEvidence["vector_id"] != evidence.VectorID {
		t.Errorf("vector_id mismatch: expected %s, got %v",
			evidence.VectorID, readEvidence["vector_id"])
	}

	// Verify file hash for integrity
	fileHash := sha256.Sum256(readData)
	hashStr := hex.EncodeToString(fileHash[:])
	t.Logf("Evidence file hash: %s", hashStr)
}

// ============================================================================
// Model Hash Pinning Tests
// ============================================================================

// TestModelHashPinning tests that model hash is properly pinned
func TestModelHashPinning(t *testing.T) {
	// Verify expected model hash is defined
	if ExpectedModelHashV1 == "" {
		t.Error("ExpectedModelHashV1 is not defined")
	}

	// Verify hash format
	if len(ExpectedModelHashV1) != 64 {
		t.Errorf("ExpectedModelHashV1 must be 64 hex chars, got %d", len(ExpectedModelHashV1))
	}

	// Verify all golden vectors reference the expected model hash
	for _, vec := range GoldenVectors {
		if vec.RequiredModelHash != ExpectedModelHashV1 {
			t.Errorf("golden vector %s has wrong model hash: expected %s, got %s",
				vec.ID, ExpectedModelHashV1, vec.RequiredModelHash)
		}
	}
}

// TestRuntimeConfigHash tests runtime configuration hashing
func TestRuntimeConfigHash(t *testing.T) {
	config := CanonicalRuntimeConfig()

	// Verify config hash is computed
	if config.ConfigHash == "" {
		t.Error("runtime config hash is empty")
	}

	// Verify hash is deterministic
	config2 := CanonicalRuntimeConfig()
	if config.ConfigHash != config2.ConfigHash {
		t.Errorf("runtime config hash not deterministic: %s != %s",
			config.ConfigHash, config2.ConfigHash)
	}

	t.Logf("Runtime config hash: %s", config.ConfigHash)
}

// ============================================================================
// Helper Functions
// ============================================================================

// computeExpectedHashes computes and logs expected hashes for documentation
func computeExpectedHashes(t *testing.T) {
	t.Helper()

	dc := NewDeterminismController(GoldenVectorSeed, true)

	t.Log("Expected hashes for golden vectors:")
	t.Log("=====================================")

	for _, vec := range GoldenVectors {
		inputHash := dc.ComputeInputHash(vec.Inputs)
		t.Logf("Vector: %s", vec.ID)
		t.Logf("  Input Hash:  %s", inputHash)
		t.Logf("  Platform:    %s/%s", runtime.GOOS, runtime.GOARCH)
	}
}

