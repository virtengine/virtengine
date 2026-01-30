package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// Composite Scoring Weight Tests
// ============================================================================

func TestCompositeScoringWeightConstants(t *testing.T) {
	// Verify individual weights match spec (veid-flow-spec.md)
	require.Equal(t, uint32(2500), WeightDocumentAuthenticity, "Document Authenticity should be 25%")
	require.Equal(t, uint32(2500), WeightFaceMatch, "Face Match should be 25%")
	require.Equal(t, uint32(2000), WeightLivenessDetection, "Liveness Detection should be 20%")
	require.Equal(t, uint32(1500), WeightDataConsistency, "Data Consistency should be 15%")
	require.Equal(t, uint32(1000), WeightHistoricalSignals, "Historical Signals should be 10%")
	require.Equal(t, uint32(500), WeightRiskIndicators, "Risk Indicators should be 5%")

	// Verify total constant is correct
	require.Equal(t, uint32(10000), TotalCompositeWeight, "Total weight constant should be 10000")
}

func TestDefaultCompositeScoringWeights(t *testing.T) {
	weights := DefaultCompositeScoringWeights()

	// Verify weights are set correctly
	require.Equal(t, WeightDocumentAuthenticity, weights.DocumentAuthenticity)
	require.Equal(t, WeightFaceMatch, weights.FaceMatch)
	require.Equal(t, WeightLivenessDetection, weights.LivenessDetection)
	require.Equal(t, WeightDataConsistency, weights.DataConsistency)
	require.Equal(t, WeightHistoricalSignals, weights.HistoricalSignals)
	require.Equal(t, WeightRiskIndicators, weights.RiskIndicators)

	// Verify total
	require.Equal(t, uint32(10000), weights.TotalWeight())

	// Verify validation passes
	require.NoError(t, weights.Validate())
}

func TestCompositeScoringWeightsValidation(t *testing.T) {
	t.Run("valid weights", func(t *testing.T) {
		weights := DefaultCompositeScoringWeights()
		require.NoError(t, weights.Validate())
	})

	t.Run("invalid weights - sum too high", func(t *testing.T) {
		weights := CompositeScoringWeights{
			DocumentAuthenticity: 3000,
			FaceMatch:            3000,
			LivenessDetection:    2000,
			DataConsistency:      1500,
			HistoricalSignals:    1000,
			RiskIndicators:       1000, // Total = 11500
		}
		err := weights.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "must sum to 10000")
	})

	t.Run("invalid weights - sum too low", func(t *testing.T) {
		weights := CompositeScoringWeights{
			DocumentAuthenticity: 2000,
			FaceMatch:            2000,
			LivenessDetection:    2000,
			DataConsistency:      1000,
			HistoricalSignals:    500,
			RiskIndicators:       500, // Total = 8000
		}
		err := weights.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "must sum to 10000")
	})
}

// ============================================================================
// Composite Scoring Threshold Tests
// ============================================================================

func TestDefaultCompositeScoringThresholds(t *testing.T) {
	thresholds := DefaultCompositeScoringThresholds()

	// Verify defaults are within valid range
	require.LessOrEqual(t, thresholds.MinDocumentAuthenticity, uint32(10000))
	require.LessOrEqual(t, thresholds.MinFaceMatch, uint32(10000))
	require.LessOrEqual(t, thresholds.MinLivenessDetection, uint32(10000))
	require.LessOrEqual(t, thresholds.MinDataConsistency, uint32(10000))
	require.LessOrEqual(t, thresholds.MinHistoricalSignals, uint32(10000))
	require.LessOrEqual(t, thresholds.MinRiskIndicators, uint32(10000))
	require.LessOrEqual(t, thresholds.RequiredForPass, MaxScore)

	// Verify validation passes
	require.NoError(t, thresholds.Validate())
}

func TestCompositeScoringThresholdsValidation(t *testing.T) {
	t.Run("valid thresholds", func(t *testing.T) {
		thresholds := DefaultCompositeScoringThresholds()
		require.NoError(t, thresholds.Validate())
	})

	t.Run("invalid - exceeds max basis points", func(t *testing.T) {
		thresholds := CompositeScoringThresholds{
			MinDocumentAuthenticity: 15000, // > 10000
		}
		err := thresholds.Validate()
		require.Error(t, err)
	})

	t.Run("invalid - required for pass exceeds 100", func(t *testing.T) {
		thresholds := CompositeScoringThresholds{
			RequiredForPass: 150, // > 100
		}
		err := thresholds.Validate()
		require.Error(t, err)
	})
}

// ============================================================================
// Component Input Score Computation Tests
// ============================================================================

func TestDocumentAuthenticityInputComputeScore(t *testing.T) {
	t.Run("not present returns 0", func(t *testing.T) {
		input := DocumentAuthenticityInput{Present: false}
		require.Equal(t, uint32(0), input.ComputeScore())
	})

	t.Run("all max returns max", func(t *testing.T) {
		input := DocumentAuthenticityInput{
			Present:               true,
			TamperScore:           10000,
			FormatValidityScore:   10000,
			TemplateMatchScore:    10000,
			SecurityFeaturesScore: 10000,
		}
		score := input.ComputeScore()
		require.Equal(t, uint32(10000), score)
	})

	t.Run("partial scores", func(t *testing.T) {
		input := DocumentAuthenticityInput{
			Present:               true,
			TamperScore:           8000,
			FormatValidityScore:   7000,
			TemplateMatchScore:    7000,
			SecurityFeaturesScore: 6000,
		}
		score := input.ComputeScore()
		require.Greater(t, score, uint32(0))
		require.LessOrEqual(t, score, uint32(10000))
	})
}

func TestFaceMatchInputComputeScore(t *testing.T) {
	t.Run("not present returns 0", func(t *testing.T) {
		input := FaceMatchInput{Present: false}
		require.Equal(t, uint32(0), input.ComputeScore())
	})

	t.Run("all max returns high score", func(t *testing.T) {
		input := FaceMatchInput{
			Present:         true,
			SimilarityScore: 10000,
			Confidence:      10000,
			QualityScore:    10000,
		}
		score := input.ComputeScore()
		require.Greater(t, score, uint32(9000))
	})

	t.Run("low confidence reduces score", func(t *testing.T) {
		highConfidence := FaceMatchInput{
			Present:         true,
			SimilarityScore: 8000,
			Confidence:      9000,
			QualityScore:    8000,
		}
		lowConfidence := FaceMatchInput{
			Present:         true,
			SimilarityScore: 8000,
			Confidence:      5000,
			QualityScore:    8000,
		}
		require.Greater(t, highConfidence.ComputeScore(), lowConfidence.ComputeScore())
	})
}

func TestLivenessDetectionInputComputeScore(t *testing.T) {
	t.Run("not present returns default 50%", func(t *testing.T) {
		input := LivenessDetectionInput{Present: false}
		require.Equal(t, uint32(5000), input.ComputeScore())
	})

	t.Run("bonus for blink detection", func(t *testing.T) {
		withBlink := LivenessDetectionInput{
			Present:       true,
			LivenessScore: 8000,
			BlinkDetected: true,
		}
		withoutBlink := LivenessDetectionInput{
			Present:       true,
			LivenessScore: 8000,
			BlinkDetected: false,
		}
		require.Greater(t, withBlink.ComputeScore(), withoutBlink.ComputeScore())
	})
}

func TestDataConsistencyInputComputeScore(t *testing.T) {
	t.Run("not present returns 0", func(t *testing.T) {
		input := DataConsistencyInput{Present: false}
		require.Equal(t, uint32(0), input.ComputeScore())
	})

	t.Run("age verification failure penalizes score", func(t *testing.T) {
		passing := DataConsistencyInput{
			Present:               true,
			NameMatchScore:        8000,
			DOBConsistencyScore:   8000,
			AgeVerificationPassed: true,
			CrossFieldValidation:  8000,
			DocumentExpiryValid:   true,
		}
		failing := DataConsistencyInput{
			Present:               true,
			NameMatchScore:        8000,
			DOBConsistencyScore:   8000,
			AgeVerificationPassed: false,
			CrossFieldValidation:  8000,
			DocumentExpiryValid:   true,
		}
		require.Greater(t, passing.ComputeScore(), failing.ComputeScore())
	})
}

func TestHistoricalSignalsInputComputeScore(t *testing.T) {
	t.Run("not present returns default 50%", func(t *testing.T) {
		input := HistoricalSignalsInput{Present: false}
		require.Equal(t, uint32(5000), input.ComputeScore())
	})

	t.Run("no prior verifications uses account age", func(t *testing.T) {
		input := HistoricalSignalsInput{
			Present:                  true,
			VerificationHistoryCount: 0,
			AccountAgeScore:          7000,
		}
		require.Equal(t, uint32(7000), input.ComputeScore())
	})
}

func TestRiskIndicatorsInputComputeScore(t *testing.T) {
	t.Run("not present returns default 50%", func(t *testing.T) {
		input := RiskIndicatorsInput{Present: false}
		require.Equal(t, uint32(5000), input.ComputeScore())
	})

	t.Run("velocity check failure penalizes", func(t *testing.T) {
		passing := RiskIndicatorsInput{
			Present:             true,
			FraudPatternScore:   8000,
			VelocityCheckPassed: true,
		}
		failing := RiskIndicatorsInput{
			Present:             true,
			FraudPatternScore:   8000,
			VelocityCheckPassed: false,
		}
		require.Greater(t, passing.ComputeScore(), failing.ComputeScore())
	})
}

// ============================================================================
// Composite Score Computation Tests
// ============================================================================

func TestComputeCompositeScore(t *testing.T) {
	weights := DefaultCompositeScoringWeights()
	thresholds := DefaultCompositeScoringThresholds()

	t.Run("all max inputs produces high score", func(t *testing.T) {
		inputs := createMaxInputs()

		result, err := ComputeCompositeScore(inputs, weights, thresholds)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should produce score >= 90
		require.GreaterOrEqual(t, result.FinalScore, uint32(90))
		require.True(t, result.Passed)

		// Should have success reason code
		hasSuccess := false
		for _, code := range result.ReasonCodes {
			if code == CompositeReasonSuccess {
				hasSuccess = true
				break
			}
		}
		require.True(t, hasSuccess)
	})

	t.Run("all min inputs produces low score", func(t *testing.T) {
		inputs := createMinInputs()

		result, err := ComputeCompositeScore(inputs, weights, thresholds)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should produce score < 50
		require.Less(t, result.FinalScore, uint32(50))
		require.False(t, result.Passed)
	})

	t.Run("score is 0-100 range", func(t *testing.T) {
		inputs := createMaxInputs()

		result, err := ComputeCompositeScore(inputs, weights, thresholds)
		require.NoError(t, err)

		require.LessOrEqual(t, result.FinalScore, uint32(100))
		require.GreaterOrEqual(t, result.FinalScore, uint32(0))
	})

	t.Run("missing document adds reason code", func(t *testing.T) {
		inputs := createMaxInputs()
		inputs.DocumentAuthenticity.Present = false

		result, err := ComputeCompositeScore(inputs, weights, thresholds)
		require.NoError(t, err)

		hasMissingDoc := false
		for _, code := range result.ReasonCodes {
			if code == CompositeReasonMissingDocument {
				hasMissingDoc = true
				break
			}
		}
		require.True(t, hasMissingDoc)
	})

	t.Run("missing face adds reason code", func(t *testing.T) {
		inputs := createMaxInputs()
		inputs.FaceMatch.Present = false

		result, err := ComputeCompositeScore(inputs, weights, thresholds)
		require.NoError(t, err)

		hasMissingSelfie := false
		for _, code := range result.ReasonCodes {
			if code == CompositeReasonMissingSelfie {
				hasMissingSelfie = true
				break
			}
		}
		require.True(t, hasMissingSelfie)
	})

	t.Run("all 6 contributions present", func(t *testing.T) {
		inputs := createMaxInputs()

		result, err := ComputeCompositeScore(inputs, weights, thresholds)
		require.NoError(t, err)

		require.Len(t, result.Contributions, 6)

		// Verify each component is present
		componentNames := map[string]bool{
			ComponentDocumentAuthenticity: false,
			ComponentFaceMatch:            false,
			ComponentLivenessDetection:    false,
			ComponentDataConsistency:      false,
			ComponentHistoricalSignals:    false,
			ComponentRiskIndicators:       false,
		}

		for _, contrib := range result.Contributions {
			componentNames[contrib.ComponentName] = true
		}

		for name, found := range componentNames {
			require.True(t, found, "Component %s should be present", name)
		}
	})

	t.Run("score version is set", func(t *testing.T) {
		inputs := createMaxInputs()

		result, err := ComputeCompositeScore(inputs, weights, thresholds)
		require.NoError(t, err)

		require.Equal(t, CompositeScoreVersion, result.ScoreVersion)
	})

	t.Run("input hash is computed", func(t *testing.T) {
		inputs := createMaxInputs()

		result, err := ComputeCompositeScore(inputs, weights, thresholds)
		require.NoError(t, err)

		require.NotEmpty(t, result.InputHash)
		require.Len(t, result.InputHash, 32) // SHA256
	})
}

func TestComputeCompositeScoreDeterminism(t *testing.T) {
	weights := DefaultCompositeScoringWeights()
	thresholds := DefaultCompositeScoringThresholds()
	inputs := createMaxInputs()

	// Compute multiple times
	var results []*CompositeScoreResult
	for i := 0; i < 10; i++ {
		result, err := ComputeCompositeScore(inputs, weights, thresholds)
		require.NoError(t, err)
		results = append(results, result)
	}

	// All results must be identical
	for i := 1; i < len(results); i++ {
		require.Equal(t, results[0].FinalScore, results[i].FinalScore, "Score should be deterministic")
		require.Equal(t, results[0].Passed, results[i].Passed, "Pass/fail should be deterministic")
		require.Equal(t, results[0].InputHash, results[i].InputHash, "Input hash should be deterministic")
		require.Equal(t, len(results[0].Contributions), len(results[i].Contributions))

		for j := range results[0].Contributions {
			require.Equal(t, results[0].Contributions[j].WeightedScore, results[i].Contributions[j].WeightedScore)
		}
	}
}

func TestInputHashUniqueness(t *testing.T) {
	inputs1 := createMaxInputs()
	inputs2 := createMaxInputs()
	inputs2.FaceMatch.SimilarityScore = 9000 // Different value

	hash1 := inputs1.ComputeInputHash()
	hash2 := inputs2.ComputeInputHash()

	require.NotEqual(t, hash1, hash2, "Different inputs should produce different hashes")
}

// ============================================================================
// Helper Functions
// ============================================================================

func createMaxInputs() CompositeScoringInputs {
	return CompositeScoringInputs{
		AccountAddress: "test-address",
		BlockHeight:    100,
		Timestamp:      time.Unix(1000000, 0),
		DocumentAuthenticity: DocumentAuthenticityInput{
			Present:               true,
			TamperScore:           10000,
			FormatValidityScore:   10000,
			TemplateMatchScore:    10000,
			SecurityFeaturesScore: 10000,
		},
		FaceMatch: FaceMatchInput{
			Present:         true,
			SimilarityScore: 10000,
			Confidence:      10000,
			QualityScore:    10000,
		},
		LivenessDetection: LivenessDetectionInput{
			Present:              true,
			LivenessScore:        10000,
			BlinkDetected:        true,
			HeadMovementDetected: true,
			DepthCheckPassed:     true,
			AntiSpoofScore:       10000,
		},
		DataConsistency: DataConsistencyInput{
			Present:                 true,
			NameMatchScore:          10000,
			DOBConsistencyScore:     10000,
			AgeVerificationPassed:   true,
			AddressConsistencyScore: 10000,
			DocumentExpiryValid:     true,
			CrossFieldValidation:    10000,
		},
		HistoricalSignals: HistoricalSignalsInput{
			Present:                    true,
			PriorVerificationScore:     10000,
			AccountAgeScore:            10000,
			VerificationHistoryCount:   5,
			SuccessfulVerificationRate: 10000,
		},
		RiskIndicators: RiskIndicatorsInput{
			Present:                true,
			FraudPatternScore:      10000,
			DeviceFingerprintScore: 10000,
			IPReputationScore:      10000,
			VelocityCheckPassed:    true,
			GeoConsistencyScore:    10000,
		},
	}
}

func createMinInputs() CompositeScoringInputs {
	return CompositeScoringInputs{
		AccountAddress: "test-address",
		BlockHeight:    100,
		Timestamp:      time.Unix(1000000, 0),
		DocumentAuthenticity: DocumentAuthenticityInput{
			Present:               true,
			TamperScore:           0,
			FormatValidityScore:   0,
			TemplateMatchScore:    0,
			SecurityFeaturesScore: 0,
		},
		FaceMatch: FaceMatchInput{
			Present:         true,
			SimilarityScore: 0,
			Confidence:      0,
			QualityScore:    0,
		},
		LivenessDetection: LivenessDetectionInput{
			Present:              true,
			LivenessScore:        0,
			BlinkDetected:        false,
			HeadMovementDetected: false,
			DepthCheckPassed:     false,
			AntiSpoofScore:       0,
		},
		DataConsistency: DataConsistencyInput{
			Present:                 true,
			NameMatchScore:          0,
			DOBConsistencyScore:     0,
			AgeVerificationPassed:   false,
			AddressConsistencyScore: 0,
			DocumentExpiryValid:     false,
			CrossFieldValidation:    0,
		},
		HistoricalSignals: HistoricalSignalsInput{
			Present:                    true,
			PriorVerificationScore:     0,
			AccountAgeScore:            0,
			VerificationHistoryCount:   0,
			SuccessfulVerificationRate: 0,
		},
		RiskIndicators: RiskIndicatorsInput{
			Present:                true,
			FraudPatternScore:      0,
			DeviceFingerprintScore: 0,
			IPReputationScore:      0,
			VelocityCheckPassed:    false,
			GeoConsistencyScore:    0,
		},
	}
}
