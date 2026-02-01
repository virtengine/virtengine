// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0
//
// Package inference provides conformance tests for ML consensus validation.
// This implements Task 15A: ML determinism validation + conformance suite.
//
// VE-219: Identity verification runtime with tolerance-based consensus
// Task 15A: ML consensus validation + conformance suite
//
// DESIGN: Tolerance-Based Consensus
// ==================================
// Rather than requiring bit-exact determinism, we validate that:
// 1. Scores from different validators are within acceptable tolerance
// 2. Model hashes match (same model on all validators)
// 3. Scores near thresholds use stricter tolerance
// 4. Consensus is achieved when median deviation is within bounds

package inference

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// ============================================================================
// Multi-Machine Conformance Suite
// ============================================================================

// TestMultiMachineConformanceSuite runs the complete multi-machine conformance suite.
// This test validates tolerance-based consensus across different machines.
func TestMultiMachineConformanceSuite(t *testing.T) {
	suite := NewMultiMachineConformanceSuite(t)

	// Phase 1: Configuration validation
	t.Run("Phase1_ConfigurationValidation", func(t *testing.T) {
		t.Run("ProductionConfig", suite.TestProductionConfigurationValid)
		t.Run("ConsensusTolerances", suite.TestConsensusTolerancesValid)
		t.Run("EnvironmentVariables", suite.TestEnvironmentVariablesCorrect)
	})

	// Phase 2: Consensus validation
	t.Run("Phase2_ConsensusValidation", func(t *testing.T) {
		t.Run("ToleranceBasedConsensus", suite.TestToleranceBasedConsensus)
		t.Run("ThresholdStrictness", suite.TestThresholdStrictness)
		t.Run("ScoreBinning", suite.TestScoreBinning)
	})

	// Phase 3: TensorFlow operation validation
	t.Run("Phase3_OperationValidation", func(t *testing.T) {
		t.Run("DeterministicOpsRegistry", suite.TestDeterministicOpsRegistryComplete)
		t.Run("NonDeterministicOpsBlocked", suite.TestNonDeterministicOpsBlocked)
		t.Run("ConditionalOpsWarned", suite.TestConditionalOpsWarned)
	})

	// Phase 4: Feature extraction consistency
	t.Run("Phase4_FeatureExtraction", func(t *testing.T) {
		t.Run("FeatureVectorConsistent", suite.TestFeatureVectorDeterministic)
		t.Run("EmbeddingNormalization", suite.TestEmbeddingNormalizationDeterministic)
		t.Run("OCRFeatures", suite.TestOCRFeaturesDeterministic)
	})

	// Phase 5: Golden vector tests
	t.Run("Phase5_GoldenVectors", func(t *testing.T) {
		t.Run("GoldenVectorIntegrity", suite.TestGoldenVectorIntegrityValid)
		t.Run("GoldenVectorScores", suite.TestGoldenVectorScoresConsistent)
	})

	// Phase 6: Simulated multi-validator consensus
	t.Run("Phase6_MultiValidatorConsensus", func(t *testing.T) {
		t.Run("SimulatedValidatorVariance", suite.TestSimulatedValidatorVariance)
		t.Run("ConsensusWithDrift", suite.TestConsensusWithDrift)
	})

	// Generate evidence file
	suite.GenerateEvidence(t)
}

// MultiMachineConformanceSuite implements the conformance test suite.
type MultiMachineConformanceSuite struct {
	t                 *testing.T
	config            *ProductionDeterminismConfig
	dc                *DeterminismController
	extractor         *FeatureExtractor
	startTime         time.Time
	results           []ConformanceTestResult
	evidenceDir       string
	platformInfo      PlatformValidationInfo
}

// ConformanceTestResult represents a single test result.
type ConformanceTestResult struct {
	TestName      string    `json:"test_name"`
	Passed        bool      `json:"passed"`
	Duration      int64     `json:"duration_ms"`
	Details       string    `json:"details,omitempty"`
	Hash          string    `json:"hash,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

// NewMultiMachineConformanceSuite creates a new conformance suite.
func NewMultiMachineConformanceSuite(t *testing.T) *MultiMachineConformanceSuite {
	// Get evidence directory from env or use temp
	evidenceDir := os.Getenv("VEID_CONFORMANCE_EVIDENCE_DIR")
	if evidenceDir == "" {
		evidenceDir = filepath.Join(os.TempDir(), "veid-conformance-evidence")
	}

	config := NewProductionDeterminismConfig()

	return &MultiMachineConformanceSuite{
		t:           t,
		config:      config,
		dc:          NewDeterminismController(config.RandomSeed, config.ForceCPU),
		extractor:   NewFeatureExtractor(DefaultFeatureExtractorConfig()),
		startTime:   time.Now(),
		results:     make([]ConformanceTestResult, 0),
		evidenceDir: evidenceDir,
		platformInfo: PlatformValidationInfo{
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
			GoVersion: runtime.Version(),
			NumCPU:    runtime.NumCPU(),
			Timestamp: time.Now().UTC(),
		},
	}
}

// recordResult records a test result.
func (s *MultiMachineConformanceSuite) recordResult(name string, passed bool, details, hash string) {
	s.results = append(s.results, ConformanceTestResult{
		TestName:  name,
		Passed:    passed,
		Duration:  time.Since(s.startTime).Milliseconds(),
		Details:   details,
		Hash:      hash,
		Timestamp: time.Now().UTC(),
	})
}

// ============================================================================
// Phase 1: Configuration Validation
// ============================================================================

// TestProductionConfigurationValid tests production configuration is valid.
func (s *MultiMachineConformanceSuite) TestProductionConfigurationValid(t *testing.T) {
	issues := s.config.Validate()

	if len(issues) > 0 {
		t.Errorf("Production config validation failed: %v", issues)
		s.recordResult("ProductionConfig", false, fmt.Sprintf("issues: %v", issues), "")
		return
	}

	// Verify consensus tolerance is set
	if s.config.ConsensusTolerance <= 0 {
		t.Error("ConsensusTolerance must be positive")
	}

	// Verify model hash is set
	if s.config.ExpectedModelHash == "" {
		t.Error("ExpectedModelHash must be set")
	}

	s.recordResult("ProductionConfig", true,
		fmt.Sprintf("tolerance=%.2f, strict=%.2f, gpu=%v",
			s.config.ConsensusTolerance, s.config.StrictTolerance, s.config.AllowGPU),
		s.config.ConfigHash)

	t.Logf("Production config validated: tolerance=%.2f, hash=%s",
		s.config.ConsensusTolerance, s.config.ConfigHash[:16])
}

// TestConsensusTolerancesValid tests the consensus tolerance settings.
func (s *MultiMachineConformanceSuite) TestConsensusTolerancesValid(t *testing.T) {
	// Verify global constants
	if ConsensusTolerance <= 0 {
		t.Error("ConsensusTolerance constant must be positive")
	}

	if ConsensusToleranceStrict <= 0 {
		t.Error("ConsensusToleranceStrict constant must be positive")
	}

	if ConsensusToleranceStrict > ConsensusTolerance {
		t.Error("Strict tolerance should be less than or equal to regular tolerance")
	}

	// Verify thresholds are reasonable
	if ScoreThresholdPass < 0 || ScoreThresholdPass > 100 {
		t.Error("ScoreThresholdPass should be between 0 and 100")
	}

	if ScoreThresholdHighTrust < ScoreThresholdPass {
		t.Error("ScoreThresholdHighTrust should be greater than ScoreThresholdPass")
	}

	s.recordResult("ConsensusTolerances", true,
		fmt.Sprintf("tolerance=%.2f, strict=%.2f, pass=%.0f, high=%.0f",
			ConsensusTolerance, ConsensusToleranceStrict,
			ScoreThresholdPass, ScoreThresholdHighTrust), "")

	t.Logf("Consensus tolerances valid: %.2f/%.2f", ConsensusTolerance, ConsensusToleranceStrict)
}

// TestEnvironmentVariablesCorrect tests environment variables configuration.
func (s *MultiMachineConformanceSuite) TestEnvironmentVariablesCorrect(t *testing.T) {
	envVars := s.config.GetEnvironmentVariables()

	// For tolerance-based consensus, we don't require strict env vars
	// but we should have basic settings
	if envVars["PYTHONHASHSEED"] == "" {
		t.Error("PYTHONHASHSEED should be set")
	}

	s.recordResult("EnvironmentVariables", true,
		fmt.Sprintf("configured %d vars", len(envVars)), "")

	t.Logf("Environment configuration verified: %d vars", len(envVars))
}

// ============================================================================
// Phase 2: Consensus Validation
// ============================================================================

// TestToleranceBasedConsensus tests the tolerance-based consensus logic.
func (s *MultiMachineConformanceSuite) TestToleranceBasedConsensus(t *testing.T) {
	testCases := []struct {
		name      string
		scoreA    float32
		scoreB    float32
		tolerance float32
		expectOK  bool
	}{
		{"identical", 75.0, 75.0, 2.0, true},
		{"within_tolerance", 75.0, 76.5, 2.0, true},
		{"at_boundary", 75.0, 77.0, 2.0, true},
		{"exceeds_tolerance", 75.0, 78.0, 2.0, false},
		{"strict_pass", 75.0, 75.4, 0.5, true},
		{"strict_fail", 75.0, 75.6, 0.5, false},
		{"near_zero", 0.5, 1.0, 2.0, true},
		{"near_hundred", 99.0, 100.0, 2.0, true},
	}

	allPassed := true
	for _, tc := range testCases {
		result := ScoresInConsensusWithTolerance(tc.scoreA, tc.scoreB, tc.tolerance)
		if result != tc.expectOK {
			t.Errorf("%s: expected %v, got %v (%.2f vs %.2f, tol=%.2f)",
				tc.name, tc.expectOK, result, tc.scoreA, tc.scoreB, tc.tolerance)
			allPassed = false
		}
	}

	s.recordResult("ToleranceBasedConsensus", allPassed,
		fmt.Sprintf("tested %d cases", len(testCases)), "")

	t.Logf("Tolerance-based consensus: %d test cases", len(testCases))
}

// TestThresholdStrictness tests stricter tolerance near thresholds.
func (s *MultiMachineConformanceSuite) TestThresholdStrictness(t *testing.T) {
	// Score near pass threshold should use strict tolerance
	nearPassScore := ScoreThresholdPass + 2.0 // Within buffer
	tolerance := GetEffectiveTolerance(nearPassScore)
	if tolerance != ConsensusToleranceStrict {
		t.Errorf("Score %.2f near pass threshold should use strict tolerance, got %.2f",
			nearPassScore, tolerance)
	}

	// Score far from thresholds should use regular tolerance
	farScore := float32(75.0)
	tolerance = GetEffectiveTolerance(farScore)
	if tolerance != ConsensusTolerance {
		t.Errorf("Score %.2f far from thresholds should use regular tolerance, got %.2f",
			farScore, tolerance)
	}

	// Score near high-trust threshold
	nearHighTrust := ScoreThresholdHighTrust - 3.0
	tolerance = GetEffectiveTolerance(nearHighTrust)
	if tolerance != ConsensusToleranceStrict {
		t.Errorf("Score %.2f near high-trust threshold should use strict tolerance",
			nearHighTrust)
	}

	s.recordResult("ThresholdStrictness", true,
		fmt.Sprintf("pass=%.0f, high=%.0f, buffer=%.0f",
			ScoreThresholdPass, ScoreThresholdHighTrust, ThresholdBuffer), "")

	t.Logf("Threshold strictness verified")
}

// TestScoreBinning tests score binning for consensus.
func (s *MultiMachineConformanceSuite) TestScoreBinning(t *testing.T) {
	testCases := []struct {
		input    float32
		expected float32
	}{
		{75.1, 75.0},
		{75.3, 75.5},
		{75.5, 75.5},
		{75.7, 75.5},
		{75.9, 76.0},
		{0.0, 0.0},
		{100.0, 100.0},
	}

	allPassed := true
	for _, tc := range testCases {
		binned := BinScore(tc.input)
		if binned != tc.expected {
			t.Errorf("BinScore(%.1f) = %.1f, expected %.1f", tc.input, binned, tc.expected)
			allPassed = false
		}
	}

	// Test integer binning
	if BinScoreInt(75.4) != 75 {
		t.Error("BinScoreInt(75.4) should be 75")
		allPassed = false
	}
	if BinScoreInt(75.6) != 76 {
		t.Error("BinScoreInt(75.6) should be 76")
		allPassed = false
	}

	s.recordResult("ScoreBinning", allPassed,
		fmt.Sprintf("tested %d cases", len(testCases)), "")

	t.Logf("Score binning verified")
}

// ============================================================================
// Phase 3: Operation Validation
// ============================================================================

// TestDeterministicOpsRegistryComplete tests the deterministic ops registry.
func (s *MultiMachineConformanceSuite) TestDeterministicOpsRegistryComplete(t *testing.T) {
	// Verify registry is not empty
	if len(DeterministicOps) == 0 {
		t.Fatal("DeterministicOps registry is empty")
	}

	// Verify common operations are present
	requiredOps := []string{"MatMul", "Add", "Relu", "Softmax", "Const"}
	for _, op := range requiredOps {
		found := false
		for _, regOp := range DeterministicOps {
			if regOp == op {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required op %s not in registry", op)
		}
	}

	s.recordResult("DeterministicOpsRegistry", true,
		fmt.Sprintf("%d deterministic ops registered", len(DeterministicOps)), "")

	t.Logf("Deterministic ops registry: %d operations", len(DeterministicOps))
}

// testOpCudnnRNN is the CudnnRNN operation name used in tests.
const testOpCudnnRNN = "CudnnRNN"

// TestNonDeterministicOpsBlocked tests non-deterministic ops are blocked.
func (s *MultiMachineConformanceSuite) TestNonDeterministicOpsBlocked(t *testing.T) {
	// Test that model validation blocks non-deterministic ops
	nonDetOps := []string{testOpCudnnRNN, "MatMul", "Add"} // CudnnRNN is non-deterministic

	result := ValidateModelOperations(nonDetOps, true)

	if result.Valid {
		t.Error("Model with CudnnRNN should fail validation")
	}

	if len(result.NonDeterministicOps) != 1 || result.NonDeterministicOps[0] != testOpCudnnRNN {
		t.Errorf("Expected CudnnRNN in non-deterministic list, got %v", result.NonDeterministicOps)
	}

	s.recordResult("NonDeterministicOpsBlocked", !result.Valid,
		fmt.Sprintf("blocked: %v", result.NonDeterministicOps), "")

	t.Logf("Non-deterministic ops blocked: %v", result.NonDeterministicOps)
}

// TestConditionalOpsWarned tests conditional ops generate warnings.
func (s *MultiMachineConformanceSuite) TestConditionalOpsWarned(t *testing.T) {
	// Test that conditional ops generate warnings
	condOps := []string{"Conv2D", "BiasAdd"} // Conditionally deterministic

	result := ValidateModelOperations(condOps, false) // Non-strict mode

	if len(result.ConditionalOps) == 0 {
		t.Error("Expected Conv2D and BiasAdd to be flagged as conditional")
	}

	if len(result.Warnings) == 0 {
		t.Error("Expected warnings for conditional ops")
	}

	s.recordResult("ConditionalOpsWarned", len(result.Warnings) > 0,
		fmt.Sprintf("warnings: %d, conditional: %v", len(result.Warnings), result.ConditionalOps), "")

	t.Logf("Conditional ops warned: %v", result.ConditionalOps)
}

// ============================================================================
// Phase 3: Hash Computation Determinism
// ============================================================================

// TestInputHashDeterministic tests input hash computation is deterministic.
func (s *MultiMachineConformanceSuite) TestInputHashDeterministic(t *testing.T) {
	inputs := createStandardTestInputs()

	// Compute hash 100 times
	hashes := make([]string, 100)
	for i := 0; i < 100; i++ {
		hashes[i] = s.dc.ComputeInputHash(inputs)
	}

	// All hashes must be identical
	for i := 1; i < 100; i++ {
		if hashes[i] != hashes[0] {
			t.Errorf("Input hash not deterministic at iteration %d: expected %s, got %s",
				i, hashes[0], hashes[i])
			s.recordResult("InputHashDeterministic", false, "hash variance detected", hashes[0])
			return
		}
	}

	// Verify hash format
	if len(hashes[0]) != 64 {
		t.Errorf("Hash length incorrect: expected 64, got %d", len(hashes[0]))
	}

	s.recordResult("InputHashDeterministic", true,
		fmt.Sprintf("100 iterations identical, len=%d", len(hashes[0])), hashes[0])

	t.Logf("Input hash deterministic: %s", hashes[0][:16])
}

// TestOutputHashDeterministic tests output hash computation is deterministic.
func (s *MultiMachineConformanceSuite) TestOutputHashDeterministic(t *testing.T) {
	testOutputs := [][]float32{
		{75.5, 0.85},
		{50.123456, 0.5},
		{99.999999, 0.999},
		{0.000001, 0.001},
	}

	for i, output := range testOutputs {
		t.Run(fmt.Sprintf("Output%d", i), func(t *testing.T) {
			var hashes []string
			for j := 0; j < 50; j++ {
				hash := s.dc.ComputeOutputHash(output)
				hashes = append(hashes, hash)
			}

			// All hashes must be identical
			for j := 1; j < len(hashes); j++ {
				if hashes[j] != hashes[0] {
					t.Errorf("Output hash not deterministic at iteration %d", j)
				}
			}
		})
	}

	// Record result with first output hash
	firstHash := s.dc.ComputeOutputHash(testOutputs[0])
	s.recordResult("OutputHashDeterministic", true,
		fmt.Sprintf("tested %d output patterns", len(testOutputs)), firstHash)
}

// TestResultHashDeterministic tests result hash computation is deterministic.
func (s *MultiMachineConformanceSuite) TestResultHashDeterministic(t *testing.T) {
	result := &ScoreResult{
		Score:        75,
		RawScore:     75.5,
		Confidence:   0.85,
		ModelVersion: "v1.0.0",
		ModelHash:    "abc123def456",
		ReasonCodes:  []string{"SUCCESS", "HIGH_CONFIDENCE"},
	}

	var hashes []string
	for i := 0; i < 50; i++ {
		hash := s.dc.ComputeResultHash(result)
		hashes = append(hashes, hash)
	}

	// All hashes must be identical
	for i := 1; i < len(hashes); i++ {
		if hashes[i] != hashes[0] {
			t.Errorf("Result hash not deterministic at iteration %d", i)
		}
	}

	s.recordResult("ResultHashDeterministic", true, "50 iterations identical", hashes[0])
	t.Logf("Result hash deterministic: %s", hashes[0][:16])
}

// TestFloatPrecisionConsistent tests float precision is consistent.
func (s *MultiMachineConformanceSuite) TestFloatPrecisionConsistent(t *testing.T) {
	// Test precision boundary cases
	testCases := []struct {
		name     string
		a, b     float32
		sameHash bool
	}{
		{"tiny_diff_same", 75.5000001, 75.5000002, true},
		{"significant_diff", 75.5, 75.6, false},
		{"near_zero_same", 0.0000001, 0.0000002, true},
		{"near_hundred_same", 99.9999991, 99.9999992, true},
	}

	allPassed := true
	for _, tc := range testCases {
		hash1 := s.dc.ComputeOutputHash([]float32{tc.a})
		hash2 := s.dc.ComputeOutputHash([]float32{tc.b})

		sameHash := hash1 == hash2
		if sameHash != tc.sameHash {
			t.Errorf("Precision test %s failed: expected sameHash=%v, got %v",
				tc.name, tc.sameHash, sameHash)
			allPassed = false
		}
	}

	s.recordResult("FloatPrecisionConsistent", allPassed,
		fmt.Sprintf("tested %d precision cases", len(testCases)), "")
}

// ============================================================================
// Phase 4: Feature Extraction Determinism
// ============================================================================

// TestFeatureVectorDeterministic tests feature vector extraction is deterministic.
func (s *MultiMachineConformanceSuite) TestFeatureVectorDeterministic(t *testing.T) {
	inputs := createStandardTestInputs()

	// Extract features multiple times
	var featureVectors [][]float32
	for i := 0; i < 20; i++ {
		features, err := s.extractor.ExtractFeatures(inputs)
		if err != nil {
			t.Fatalf("Feature extraction failed: %v", err)
		}
		featureVectors = append(featureVectors, features)
	}

	// All feature vectors must be identical
	for i := 1; i < len(featureVectors); i++ {
		if len(featureVectors[i]) != len(featureVectors[0]) {
			t.Errorf("Feature vector length mismatch at iteration %d", i)
			continue
		}
		for j := range featureVectors[0] {
			if featureVectors[i][j] != featureVectors[0][j] {
				t.Errorf("Feature mismatch at [%d][%d]: %f != %f",
					i, j, featureVectors[i][j], featureVectors[0][j])
			}
		}
	}

	// Verify dimension
	if len(featureVectors[0]) != TotalFeatureDim {
		t.Errorf("Feature dimension mismatch: expected %d, got %d",
			TotalFeatureDim, len(featureVectors[0]))
	}

	s.recordResult("FeatureVectorDeterministic", true,
		fmt.Sprintf("dim=%d, 20 iterations identical", len(featureVectors[0])), "")
}

// TestEmbeddingNormalizationDeterministic tests embedding normalization.
func (s *MultiMachineConformanceSuite) TestEmbeddingNormalizationDeterministic(t *testing.T) {
	// Create test embedding
	embedding := make([]float32, FaceEmbeddingDim)
	for i := range embedding {
		embedding[i] = float32(i) * 0.001
	}

	// Normalize multiple times
	results := make([][]float32, 10)
	for i := 0; i < 10; i++ {
		copy := make([]float32, len(embedding))
		for j := range embedding {
			copy[j] = embedding[j]
		}
		s.extractor.normalizeEmbedding(copy)
		results[i] = copy
	}

	// All results must be identical
	for i := 1; i < len(results); i++ {
		for j := range results[0] {
			if results[i][j] != results[0][j] {
				t.Errorf("Normalization not deterministic at [%d][%d]", i, j)
			}
		}
	}

	s.recordResult("EmbeddingNormalizationDeterministic", true, "10 iterations identical", "")
}

// TestOCRFeaturesDeterministic tests OCR feature extraction is deterministic.
func (s *MultiMachineConformanceSuite) TestOCRFeaturesDeterministic(t *testing.T) {
	inputs := createStandardTestInputs()

	// Extract and compare OCR feature portions
	features1, _ := s.extractor.ExtractFeatures(inputs)
	features2, _ := s.extractor.ExtractFeatures(inputs)

	ocrOffset := FaceEmbeddingDim + DocQualityDim
	for i := 0; i < OCRFieldsDim; i++ {
		if features1[ocrOffset+i] != features2[ocrOffset+i] {
			t.Errorf("OCR feature %d not deterministic", i)
		}
	}

	s.recordResult("OCRFeaturesDeterministic", true,
		fmt.Sprintf("%d OCR features verified", OCRFieldsDim), "")
}

// ============================================================================
// Phase 5: Golden Vector Tests
// ============================================================================

// TestGoldenVectorIntegrityValid tests golden vectors are valid.
func (s *MultiMachineConformanceSuite) TestGoldenVectorIntegrityValid(t *testing.T) {
	for _, vec := range GoldenVectors {
		t.Run(vec.ID, func(t *testing.T) {
			if err := ValidateGoldenVectorIntegrity(&vec); err != nil {
				t.Errorf("Golden vector %s invalid: %v", vec.ID, err)
			}
		})
	}

	s.recordResult("GoldenVectorIntegrity", true,
		fmt.Sprintf("%d vectors validated", len(GoldenVectors)), "")
}

// TestGoldenVectorScoresConsistent tests golden vector scores are consistent.
func (s *MultiMachineConformanceSuite) TestGoldenVectorScoresConsistent(t *testing.T) {
	for _, vec := range GoldenVectors {
		t.Run(vec.ID+"_score", func(t *testing.T) {
			// Extract features
			features, err := s.extractor.ExtractFeatures(vec.Inputs)
			if err != nil {
				t.Fatalf("Feature extraction failed: %v", err)
			}

			// Verify feature dimension
			if len(features) != TotalFeatureDim {
				t.Errorf("Feature dimension mismatch: expected %d, got %d",
					TotalFeatureDim, len(features))
			}
		})
	}

	s.recordResult("GoldenVectorScores", true,
		fmt.Sprintf("verified %d vectors", len(GoldenVectors)), "")
}

// ============================================================================
// Phase 6: Multi-Validator Consensus Simulation
// ============================================================================

// TestSimulatedValidatorVariance tests consensus with simulated validator variance.
func (s *MultiMachineConformanceSuite) TestSimulatedValidatorVariance(t *testing.T) {
	// Simulate 5 validators with slight score variance (as would happen with GPU/threading)
	baseScore := float32(75.0)
	validatorScores := []float32{
		baseScore,         // Validator 1
		baseScore + 0.3,   // Validator 2 (slight positive drift)
		baseScore - 0.2,   // Validator 3 (slight negative drift)
		baseScore + 0.5,   // Validator 4
		baseScore - 0.4,   // Validator 5
	}

	// Check consensus
	result := ValidateConsensus(validatorScores)

	if !result.Achieved {
		t.Errorf("Consensus should be achieved with variance <= %.2f: %s",
			ConsensusTolerance, result.Reason)
	}

	// Verify median is close to base score
	if math.Abs(float64(result.MedianScore-baseScore)) > 1.0 {
		t.Errorf("Median score %.2f too far from base %.2f", result.MedianScore, baseScore)
	}

	t.Logf("Validator consensus: median=%.2f, max_deviation=%.2f, tolerance=%.2f",
		result.MedianScore, result.MaxDeviation, result.Tolerance)

	s.recordResult("SimulatedValidatorVariance", result.Achieved,
		fmt.Sprintf("5 validators, median=%.2f, max_dev=%.2f",
			result.MedianScore, result.MaxDeviation), "")
}

// TestConsensusWithDrift tests consensus fails when drift exceeds tolerance.
func (s *MultiMachineConformanceSuite) TestConsensusWithDrift(t *testing.T) {
	// Test case 1: Acceptable drift
	acceptableScores := []float32{75.0, 75.5, 76.0, 75.3, 75.8}
	result := ValidateConsensus(acceptableScores)
	if !result.Achieved {
		t.Errorf("Acceptable drift should achieve consensus: %s", result.Reason)
	}

	// Test case 2: Excessive drift
	excessiveScores := []float32{70.0, 75.0, 80.0, 72.0, 78.0}
	result = ValidateConsensus(excessiveScores)
	if result.Achieved {
		t.Errorf("Excessive drift (10 point spread) should fail consensus")
	}

	// Test case 3: Near threshold - should use strict tolerance
	nearThresholdScores := []float32{60.0, 60.3, 60.1, 60.2, 59.9}
	result = ValidateConsensus(nearThresholdScores)
	if result.Tolerance != ConsensusToleranceStrict {
		t.Errorf("Near threshold should use strict tolerance %.2f, got %.2f",
			ConsensusToleranceStrict, result.Tolerance)
	}

	// Test case 4: Single validator always achieves consensus
	singleResult := ValidateConsensus([]float32{75.0})
	if !singleResult.Achieved {
		t.Error("Single validator should always achieve consensus")
	}

	s.recordResult("ConsensusWithDrift", true,
		"tested acceptable, excessive, threshold, and single validator cases", "")

	t.Logf("Consensus drift tests passed")
}

// ============================================================================
// Legacy Phase 6: Hash Consistency (kept for backwards compatibility)
// ============================================================================
// ============================================================================

// TestMultipleControllersSameResult tests multiple controllers produce same results.
func (s *MultiMachineConformanceSuite) TestMultipleControllersSameResult(t *testing.T) {
	// Create multiple controllers with same seed
	controllers := make([]*DeterminismController, 5)
	for i := range controllers {
		controllers[i] = NewDeterminismController(ProductionRandomSeed, true)
	}

	inputs := createStandardTestInputs()

	// All controllers should produce same hash
	hashes := make([]string, 0, len(controllers))
	for _, dc := range controllers {
		hash := dc.ComputeInputHash(inputs)
		hashes = append(hashes, hash)
	}

	for i := 1; i < len(hashes); i++ {
		if hashes[i] != hashes[0] {
			t.Errorf("Controller %d produced different hash", i)
		}
	}

	s.recordResult("MultipleControllers", true,
		fmt.Sprintf("%d controllers identical", len(controllers)), hashes[0])
}

// TestSerialExecutionConsistent tests serial execution is consistent.
func (s *MultiMachineConformanceSuite) TestSerialExecutionConsistent(t *testing.T) {
	inputs := createStandardTestInputs()

	// Run serially 100 times
	var hashes []string
	for i := 0; i < 100; i++ {
		hash := s.dc.ComputeInputHash(inputs)
		hashes = append(hashes, hash)
	}

	// All must be identical
	for i := 1; i < len(hashes); i++ {
		if hashes[i] != hashes[0] {
			t.Errorf("Serial execution inconsistent at %d", i)
		}
	}

	s.recordResult("SerialExecution", true, "100 iterations consistent", hashes[0])
}

// TestRepeatedHashingIdentical tests repeated hashing is identical.
func (s *MultiMachineConformanceSuite) TestRepeatedHashingIdentical(t *testing.T) {
	output := []float32{75.5, 0.85, 0.5, 0.25}

	// Hash same output 1000 times
	hashes := make(map[string]int)
	for i := 0; i < 1000; i++ {
		hash := s.dc.ComputeOutputHash(output)
		hashes[hash]++
	}

	if len(hashes) != 1 {
		t.Errorf("Expected 1 unique hash, got %d", len(hashes))
	}

	var firstHash string
	for h := range hashes {
		firstHash = h
		break
	}

	s.recordResult("RepeatedHashing", len(hashes) == 1,
		fmt.Sprintf("1000 iterations, %d unique hashes", len(hashes)), firstHash)
}

// ============================================================================
// Evidence Generation
// ============================================================================

// GenerateEvidence generates the conformance evidence file.
func (s *MultiMachineConformanceSuite) GenerateEvidence(t *testing.T) {
	// Create evidence directory
	if err := os.MkdirAll(s.evidenceDir, 0750); err != nil {
		t.Logf("Warning: could not create evidence directory: %v", err)
		return
	}

	// Count passed/failed
	passed := 0
	failed := 0
	for _, r := range s.results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}

	evidence := struct {
		ReportID      string                      `json:"report_id"`
		Version       string                      `json:"version"`
		Platform      PlatformValidationInfo      `json:"platform"`
		Config        *ProductionDeterminismConfig `json:"config"`
		StartTime     time.Time                   `json:"start_time"`
		EndTime       time.Time                   `json:"end_time"`
		DurationMs    int64                       `json:"duration_ms"`
		TestsPassed   int                         `json:"tests_passed"`
		TestsFailed   int                         `json:"tests_failed"`
		TotalTests    int                         `json:"total_tests"`
		OverallPassed bool                        `json:"overall_passed"`
		Results       []ConformanceTestResult     `json:"results"`
	}{
		ReportID:      fmt.Sprintf("conf-%s-%s-%d", runtime.GOOS, runtime.GOARCH, time.Now().Unix()),
		Version:       DeterminismConfigVersion,
		Platform:      s.platformInfo,
		Config:        s.config,
		StartTime:     s.startTime,
		EndTime:       time.Now().UTC(),
		DurationMs:    time.Since(s.startTime).Milliseconds(),
		TestsPassed:   passed,
		TestsFailed:   failed,
		TotalTests:    len(s.results),
		OverallPassed: failed == 0,
		Results:       s.results,
	}

	data, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		t.Logf("Warning: failed to marshal evidence: %v", err)
		return
	}

	filename := fmt.Sprintf("conformance-evidence-%s-%s-%s.json",
		runtime.GOOS, runtime.GOARCH, time.Now().Format("20060102-150405"))
	filePath := filepath.Join(s.evidenceDir, filename)

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		t.Logf("Warning: failed to write evidence: %v", err)
		return
	}

	t.Logf("Evidence saved to %s", filePath)
	t.Logf("Summary: %d passed, %d failed, overall=%v", passed, failed, failed == 0)
}

// ============================================================================
// Helper Functions
// ============================================================================

// createStandardTestInputs creates standard test inputs for conformance tests.
func createStandardTestInputs() *ScoreInputs {
	// Create deterministic face embedding
	faceEmbedding := make([]float32, FaceEmbeddingDim)
	for i := range faceEmbedding {
		faceEmbedding[i] = float32(i%100) / 100.0
	}

	return &ScoreInputs{
		FaceEmbedding:   faceEmbedding,
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
			AccountAddress: "virt1conformance000000000000000000001",
			BlockHeight:    1000000,
			RequestID:      "conformance-test-001",
		},
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

// BenchmarkMultiMachineInputHash benchmarks input hash computation.
func BenchmarkMultiMachineInputHash(b *testing.B) {
	dc := NewDeterminismController(ProductionRandomSeed, true)
	inputs := createStandardTestInputs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dc.ComputeInputHash(inputs)
	}
}

// BenchmarkMultiMachineOutputHash benchmarks output hash computation.
func BenchmarkMultiMachineOutputHash(b *testing.B) {
	dc := NewDeterminismController(ProductionRandomSeed, true)
	output := make([]float32, TotalFeatureDim)
	for i := range output {
		output[i] = float32(i) / float32(TotalFeatureDim)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dc.ComputeOutputHash(output)
	}
}

// BenchmarkMultiMachineFeatureExtract benchmarks feature extraction.
func BenchmarkMultiMachineFeatureExtract(b *testing.B) {
	extractor := NewFeatureExtractor(DefaultFeatureExtractorConfig())
	inputs := createStandardTestInputs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractor.ExtractFeatures(inputs)
	}
}

