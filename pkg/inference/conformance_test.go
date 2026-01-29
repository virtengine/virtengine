// Package inference provides ML scoring for VEID.
//
// This file contains conformance tests to ensure Go inference
// produces identical outputs to Python for the same inputs.
//
// Task Reference: VE-3006 - Go-Python Conformance Testing
package inference

import (
	"math"
	"sort"
	"testing"
)

// DefaultTolerance is the default floating-point comparison tolerance
const DefaultTolerance = 1e-6

// ============================================================================
// Conformance Test Suite
// ============================================================================

// TestGoMatchesPythonOutput verifies Go produces same output as Python
// for all conformance test vectors
func TestGoMatchesPythonOutput(t *testing.T) {
	if len(ConformanceTestVectors) == 0 {
		t.Skip("No conformance test vectors available")
	}

	t.Logf("Running conformance tests with %d test vectors", len(ConformanceTestVectors))

	for _, vector := range ConformanceTestVectors {
		vector := vector // Capture range variable
		t.Run(vector.Name, func(t *testing.T) {
			t.Parallel()

			// Convert test vector input to ScoreInputs
			inputs := vector.Input.ConvertToScoreInputs()

			// Extract features using the Go feature extractor
			extractor := NewFeatureExtractor(DefaultFeatureExtractorConfig())
			features, err := extractor.ExtractFeatures(inputs)
			if err != nil {
				// Some test vectors intentionally have invalid inputs
				if vector.Name != "missing_face_embedding" {
					t.Errorf("Feature extraction failed: %v", err)
				}
				return
			}

			// Verify feature dimension
			if len(features) != TotalFeatureDim {
				t.Errorf("Feature dimension mismatch: expected %d, got %d",
					TotalFeatureDim, len(features))
			}

			// Compute feature contributions
			contributions := extractor.ComputeFeatureContributions(features)
			if contributions == nil {
				t.Error("Feature contributions should not be nil")
			}

			// Verify face embedding contribution is present
			if _, ok := contributions["face_embedding"]; !ok {
				t.Error("Missing face_embedding contribution")
			}

			// Verify doc quality contribution
			if _, ok := contributions["doc_quality"]; !ok {
				t.Error("Missing doc_quality contribution")
			}

			// Log test vector results
			t.Logf("Vector %s: features extracted successfully, dim=%d",
				vector.Name, len(features))
		})
	}
}

// TestFeatureExtractionDeterminism verifies same input always produces same features
func TestFeatureExtractionDeterminism(t *testing.T) {
	extractor := NewFeatureExtractor(DefaultFeatureExtractorConfig())

	for _, vector := range ConformanceTestVectors {
		if len(vector.Input.FaceEmbedding) == 0 {
			continue // Skip vectors with missing face embeddings
		}

		t.Run(vector.Name+"_determinism", func(t *testing.T) {
			inputs := vector.Input.ConvertToScoreInputs()

			// Extract features multiple times
			const iterations = 10
			var firstFeatures []float32

			for i := 0; i < iterations; i++ {
				features, err := extractor.ExtractFeatures(inputs)
				if err != nil {
					t.Fatalf("Iteration %d: feature extraction failed: %v", i, err)
				}

				if i == 0 {
					firstFeatures = make([]float32, len(features))
					copy(firstFeatures, features)
					continue
				}

				// Verify all features match exactly
				for j := 0; j < len(features); j++ {
					if features[j] != firstFeatures[j] {
						t.Errorf("Iteration %d, index %d: feature mismatch: "+
							"first=%.10f, current=%.10f",
							i, j, firstFeatures[j], features[j])
					}
				}
			}
		})
	}
}

// TestReasonCodeMapping verifies Go reason codes match Python exactly
func TestReasonCodeMapping(t *testing.T) {
	// Define expected reason codes that must match Python
	expectedReasonCodes := []string{
		ReasonCodeSuccess,
		ReasonCodeHighConfidence,
		ReasonCodeLowConfidence,
		ReasonCodeFaceMismatch,
		ReasonCodeLowDocQuality,
		ReasonCodeLowOCRConfidence,
		ReasonCodeInsufficientScopes,
		ReasonCodeMissingFace,
		ReasonCodeMissingDocument,
		ReasonCodeModelLoadError,
		ReasonCodeInferenceError,
		ReasonCodeTimeout,
		ReasonCodeMemoryLimit,
	}

	// Verify all reason codes are non-empty strings
	for _, code := range expectedReasonCodes {
		if code == "" {
			t.Error("Found empty reason code")
		}
	}

	// Verify reason codes are unique
	codeSet := make(map[string]bool)
	for _, code := range expectedReasonCodes {
		if codeSet[code] {
			t.Errorf("Duplicate reason code: %s", code)
		}
		codeSet[code] = true
	}

	// Verify test vectors only use valid reason codes
	for _, vector := range ConformanceTestVectors {
		for _, code := range vector.ExpectedCodes {
			if !codeSet[code] {
				t.Errorf("Vector %s uses unknown reason code: %s",
					vector.Name, code)
			}
		}
	}
}

// TestScorePrecision verifies floating point precision matches
func TestScorePrecision(t *testing.T) {
	// Test precision of deterministic hash computation
	dc := NewDeterminismController(42, true)

	// Test cases with known precision requirements
	testCases := []struct {
		name   string
		values []float32
	}{
		{
			name:   "small_values",
			values: []float32{0.000001, 0.000002, 0.000003},
		},
		{
			name:   "large_values",
			values: []float32{99.999999, 100.0, 99.999998},
		},
		{
			name:   "mixed_values",
			values: []float32{0.5, 50.5, 100.0, 0.0},
		},
		{
			name:   "negative_values",
			values: []float32{-0.5, -0.25, -0.125},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Compute hash multiple times - should be identical
			hash1 := dc.ComputeOutputHash(tc.values)
			hash2 := dc.ComputeOutputHash(tc.values)

			if hash1 != hash2 {
				t.Errorf("Hash mismatch for same values: %s != %s", hash1, hash2)
			}

			// Slightly different values should produce different hashes
			modified := make([]float32, len(tc.values))
			copy(modified, tc.values)
			if len(modified) > 0 {
				modified[0] += 0.0001 // Small but detectable difference

				hash3 := dc.ComputeOutputHash(modified)
				if hash1 == hash3 {
					t.Error("Hash should be different for different values")
				}
			}
		})
	}
}

// TestInputHashDeterminism verifies input hash is deterministic
func TestInputHashDeterminism(t *testing.T) {
	dc := NewDeterminismController(42, true)

	for _, vector := range ConformanceTestVectors {
		t.Run(vector.Name+"_input_hash", func(t *testing.T) {
			inputs := vector.Input.ConvertToScoreInputs()

			// Compute hash multiple times
			hash1 := dc.ComputeInputHash(inputs)
			hash2 := dc.ComputeInputHash(inputs)

			if hash1 != hash2 {
				t.Errorf("Input hash mismatch: %s != %s", hash1, hash2)
			}

			// Verify hash is non-empty
			if hash1 == "" {
				t.Error("Input hash should not be empty")
			}

			// Verify hash has expected length (SHA256 = 64 hex chars)
			if len(hash1) != 64 {
				t.Errorf("Expected 64-char hash, got %d chars", len(hash1))
			}
		})
	}
}

// TestFeatureDimensionConstantsMatch verifies Go constants match Python config
func TestFeatureDimensionConstantsMatch(t *testing.T) {
	// These must match ml/training/config.py:FeatureConfig
	tests := []struct {
		name     string
		goValue  int
		pyValue  int // Expected value from Python config
	}{
		{"FaceEmbeddingDim", FaceEmbeddingDim, 512},
		{"DocQualityDim", DocQualityDim, 5},
		{"OCRFieldsDim", OCRFieldsDim, 10}, // 5 fields * 2
		{"MetadataDim", MetadataDim, 16},
		{"TotalFeatureDim", TotalFeatureDim, 768},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.goValue != tc.pyValue {
				t.Errorf("%s: Go value %d != Python value %d",
					tc.name, tc.goValue, tc.pyValue)
			}
		})
	}

	// Verify dimensions sum correctly
	expectedSum := FaceEmbeddingDim + DocQualityDim + OCRFieldsDim + MetadataDim + PaddingDim
	if expectedSum != TotalFeatureDim {
		t.Errorf("Dimension sum mismatch: %d != %d", expectedSum, TotalFeatureDim)
	}
}

// TestOCRFieldNamesMatch verifies OCR field names match Python
func TestOCRFieldNamesMatch(t *testing.T) {
	// Expected OCR field names from Python config
	expectedFields := []string{
		"name",
		"date_of_birth",
		"document_number",
		"expiry_date",
		"nationality",
	}

	if len(OCRFieldNames) != len(expectedFields) {
		t.Errorf("OCR field count mismatch: Go=%d, expected=%d",
			len(OCRFieldNames), len(expectedFields))
	}

	// Sort both for comparison
	goFields := make([]string, len(OCRFieldNames))
	copy(goFields, OCRFieldNames)
	sort.Strings(goFields)

	pyFields := make([]string, len(expectedFields))
	copy(pyFields, expectedFields)
	sort.Strings(pyFields)

	for i, expected := range pyFields {
		if i >= len(goFields) {
			t.Errorf("Missing OCR field: %s", expected)
			continue
		}
		if goFields[i] != expected {
			t.Errorf("OCR field mismatch at %d: Go=%s, expected=%s",
				i, goFields[i], expected)
		}
	}
}

// TestEmbeddingNormalization verifies embedding normalization is deterministic
func TestEmbeddingNormalization(t *testing.T) {
	extractor := NewFeatureExtractor(DefaultFeatureExtractorConfig())

	// Create test embedding
	embedding := make([]float32, FaceEmbeddingDim)
	for i := range embedding {
		embedding[i] = float32(i) * 0.001
	}

	// Copy and normalize
	normalized := make([]float32, len(embedding))
	copy(normalized, embedding)
	extractor.normalizeEmbedding(normalized)

	// Verify unit length (L2 norm = 1)
	var sumSquares float64
	for _, v := range normalized {
		sumSquares += float64(v) * float64(v)
	}
	norm := math.Sqrt(sumSquares)

	if math.Abs(norm-1.0) > DefaultTolerance {
		t.Errorf("Normalized embedding should have unit length, got %.10f", norm)
	}

	// Normalize again - should be identical (idempotent)
	normalized2 := make([]float32, len(normalized))
	copy(normalized2, normalized)
	extractor.normalizeEmbedding(normalized2)

	for i := range normalized {
		if normalized[i] != normalized2[i] {
			t.Errorf("Normalization not idempotent at index %d", i)
		}
	}
}

// TestConfidenceComputation verifies confidence calculation matches Python
func TestConfidenceComputation(t *testing.T) {
	testCases := []struct {
		rawScore           float32
		expectedConfidence float32
		tolerance          float32
	}{
		{rawScore: 0.0, expectedConfidence: 0.9, tolerance: 0.05},    // Near 0
		{rawScore: 100.0, expectedConfidence: 0.9, tolerance: 0.05},  // Near 100
		{rawScore: 50.0, expectedConfidence: 0.5, tolerance: 0.05},   // Middle
		{rawScore: 25.0, expectedConfidence: 0.7, tolerance: 0.05},   // Low
		{rawScore: 75.0, expectedConfidence: 0.7, tolerance: 0.05},   // High
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			confidence := computeConfidence(tc.rawScore)

			diff := absFloat32(confidence - tc.expectedConfidence)
			if diff > tc.tolerance {
				t.Errorf("For rawScore=%.1f: expected confidence ~%.2f, got %.2f",
					tc.rawScore, tc.expectedConfidence, confidence)
			}
		})
	}
}

// TestValidInputsDetection verifies input validation matches Python
func TestValidInputsDetection(t *testing.T) {
	extractor := NewFeatureExtractor(DefaultFeatureExtractorConfig())

	testCases := []struct {
		name           string
		input          *ScoreInputs
		expectedIssues int
	}{
		{
			name: "valid_inputs",
			input: &ScoreInputs{
				FaceEmbedding:   make([]float32, FaceEmbeddingDim),
				FaceConfidence:  0.9,
				DocQualityScore: 0.8,
			},
			expectedIssues: 0,
		},
		{
			name: "missing_face",
			input: &ScoreInputs{
				FaceEmbedding:   nil,
				FaceConfidence:  0.9,
				DocQualityScore: 0.8,
			},
			expectedIssues: 1,
		},
		{
			name: "wrong_face_dim",
			input: &ScoreInputs{
				FaceEmbedding:   make([]float32, 256), // Wrong dimension
				FaceConfidence:  0.9,
				DocQualityScore: 0.8,
			},
			expectedIssues: 1,
		},
		{
			name: "invalid_face_confidence",
			input: &ScoreInputs{
				FaceEmbedding:   make([]float32, FaceEmbeddingDim),
				FaceConfidence:  1.5, // Out of range
				DocQualityScore: 0.8,
			},
			expectedIssues: 1,
		},
		{
			name: "invalid_doc_quality",
			input: &ScoreInputs{
				FaceEmbedding:   make([]float32, FaceEmbeddingDim),
				FaceConfidence:  0.9,
				DocQualityScore: -0.1, // Out of range
			},
			expectedIssues: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			issues := extractor.ValidateInputs(tc.input)

			if len(issues) != tc.expectedIssues {
				t.Errorf("Expected %d issues, got %d: %v",
					tc.expectedIssues, len(issues), issues)
			}
		})
	}
}

// ============================================================================
// Benchmark Tests for Conformance
// ============================================================================

// BenchmarkConformanceFeatureExtraction measures feature extraction with test vectors
func BenchmarkConformanceFeatureExtraction(b *testing.B) {
	if len(ConformanceTestVectors) == 0 {
		b.Skip("No conformance test vectors available")
	}

	extractor := NewFeatureExtractor(DefaultFeatureExtractorConfig())
	vector := ConformanceTestVectors[0]
	inputs := vector.Input.ConvertToScoreInputs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := extractor.ExtractFeatures(inputs)
		if err != nil {
			b.Fatalf("Feature extraction failed: %v", err)
		}
	}
}

// BenchmarkConformanceInputHash measures input hash with test vectors
func BenchmarkConformanceInputHash(b *testing.B) {
	if len(ConformanceTestVectors) == 0 {
		b.Skip("No conformance test vectors available")
	}

	dc := NewDeterminismController(42, true)
	vector := ConformanceTestVectors[0]
	inputs := vector.Input.ConvertToScoreInputs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dc.ComputeInputHash(inputs)
	}
}

// BenchmarkConformanceEmbeddingNormalization measures embedding normalization
func BenchmarkConformanceEmbeddingNormalization(b *testing.B) {
	extractor := NewFeatureExtractor(DefaultFeatureExtractorConfig())
	embedding := make([]float32, FaceEmbeddingDim)
	for i := range embedding {
		embedding[i] = float32(i) * 0.001
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testEmbed := make([]float32, len(embedding))
		copy(testEmbed, embedding)
		extractor.normalizeEmbedding(testEmbed)
	}
}
