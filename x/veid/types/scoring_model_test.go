package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Scoring Model Test Suite
// ============================================================================

type ScoringModelTestSuite struct {
	suite.Suite
}

func TestScoringModelTestSuite(t *testing.T) {
	suite.Run(t, new(ScoringModelTestSuite))
}

// ============================================================================
// Scoring Weights Tests
// ============================================================================

func (s *ScoringModelTestSuite) TestDefaultScoringWeightsValidation() {
	weights := types.DefaultScoringWeights()
	err := weights.Validate()
	s.Require().NoError(err, "default weights should be valid")
	s.Require().Equal(uint32(10000), weights.TotalWeight(), "default weights should sum to 10000")
}

func (s *ScoringModelTestSuite) TestScoringWeightsValidation_Invalid() {
	// Weights that don't sum to 10000
	weights := types.ScoringWeights{
		FaceSimilarityWeight: 3000,
		OCRConfidenceWeight:  2500,
		DocIntegrityWeight:   2000,
		SaltBindingWeight:    1000,
		LivenessCheckWeight:  1000,
		CaptureQualityWeight: 400, // Wrong - should be 500
	}
	err := weights.Validate()
	s.Require().Error(err, "weights not summing to 10000 should fail")
	s.Require().Contains(err.Error(), "must sum to 10000")
}

func (s *ScoringModelTestSuite) TestScoringWeightsValidation_ExactlyValid() {
	weights := types.ScoringWeights{
		FaceSimilarityWeight: 2000,
		OCRConfidenceWeight:  2000,
		DocIntegrityWeight:   2000,
		SaltBindingWeight:    2000,
		LivenessCheckWeight:  1000,
		CaptureQualityWeight: 1000,
	}
	err := weights.Validate()
	s.Require().NoError(err)
	s.Require().Equal(uint32(10000), weights.TotalWeight())
}

// ============================================================================
// Scoring Thresholds Tests
// ============================================================================

func (s *ScoringModelTestSuite) TestDefaultScoringThresholdsValidation() {
	thresholds := types.DefaultScoringThresholds()
	err := thresholds.Validate()
	s.Require().NoError(err, "default thresholds should be valid")
}

func (s *ScoringModelTestSuite) TestScoringThresholds_ExceedsMaximum() {
	thresholds := types.DefaultScoringThresholds()
	thresholds.MinFaceSimilarity = 10001 // Exceeds max
	err := thresholds.Validate()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "min_face_similarity")
}

func (s *ScoringModelTestSuite) TestScoringThresholds_RequiredForPassExceedsMaxScore() {
	thresholds := types.DefaultScoringThresholds()
	thresholds.RequiredForPass = 101 // Exceeds max score of 100
	err := thresholds.Validate()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "required_for_pass")
}

// ============================================================================
// Scoring Input Tests
// ============================================================================

func (s *ScoringModelTestSuite) TestFaceSimilarityInput_Validate() {
	// Valid input
	input := types.FaceSimilarityInput{
		SimilarityScore: 8500,
		Confidence:      9000,
		Present:         true,
	}
	s.Require().NoError(input.Validate())

	// Invalid - exceeds max
	input.SimilarityScore = 10001
	s.Require().Error(input.Validate())
}

func (s *ScoringModelTestSuite) TestOCRConfidenceInput_ComputeFieldScore() {
	input := types.OCRConfidenceInput{
		ExtractedFieldCount: 8,
		ExpectedFieldCount:  10,
		Present:             true,
	}
	score := input.ComputeFieldScore()
	s.Require().Equal(uint32(8000), score) // 80% extraction rate

	// Zero expected fields
	input.ExpectedFieldCount = 0
	score = input.ComputeFieldScore()
	s.Require().Equal(uint32(0), score)
}

func (s *ScoringModelTestSuite) TestDocIntegrityInput_ComputeIntegrityScore() {
	input := types.DocIntegrityInput{
		QualityScore:          8000,
		FormatValid:           true,
		TemplateMatch:         true,
		TamperDetectionPassed: true,
		ExpiryValid:           true,
		Present:               true,
	}
	score := input.ComputeIntegrityScore()
	s.Require().Equal(uint32(8000), score) // Full score

	// With failed checks
	input.FormatValid = false // 30% penalty
	score = input.ComputeIntegrityScore()
	s.Require().Equal(uint32(5600), score) // 8000 * 0.7

	// Not present
	input.Present = false
	score = input.ComputeIntegrityScore()
	s.Require().Equal(uint32(0), score)
}

func (s *ScoringModelTestSuite) TestSaltBindingInput_ComputeBindingScore() {
	// All valid
	input := types.SaltBindingInput{
		SaltPresent:          true,
		SaltValid:            true,
		ClientSignatureValid: true,
		UserSignatureValid:   true,
	}
	score := input.ComputeBindingScore()
	s.Require().Equal(uint32(10000), score) // 3333 + 3333 + 3334

	// Missing salt
	input.SaltPresent = false
	score = input.ComputeBindingScore()
	s.Require().Equal(uint32(0), score)

	// Partial validity
	input.SaltPresent = true
	input.SaltValid = false
	score = input.ComputeBindingScore()
	s.Require().Equal(uint32(6667), score) // 3333 + 3334
}

func (s *ScoringModelTestSuite) TestLivenessCheckInput_ComputeLivenessScore() {
	input := types.LivenessCheckInput{
		LivenessScore:        8000,
		BlinkDetected:        true,
		HeadMovementDetected: true,
		VideoFrameCount:      30,
		Present:              true,
	}
	score := input.ComputeLivenessScore()
	// 8000 * 1.05 * 1.05 * 1.03 = 9108.9 -> 9108
	s.Require().Greater(score, uint32(9000))
	s.Require().LessOrEqual(score, uint32(10000))

	// Not present
	input.Present = false
	score = input.ComputeLivenessScore()
	s.Require().Equal(uint32(0), score)
}

func (s *ScoringModelTestSuite) TestCaptureQualityInput_ComputeCaptureScore() {
	input := types.CaptureQualityInput{
		OverallQuality:     8000,
		LightingScore:      7500,
		FocusScore:         8500,
		AngleScore:         9000,
		ResolutionAdequate: true,
		Present:            true,
	}
	score := input.ComputeCaptureScore()
	// Weighted: (8000*40 + 7500*20 + 8500*25 + 9000*15) / 100 = 8175
	s.Require().Equal(uint32(8175), score)

	// With resolution penalty
	input.ResolutionAdequate = false
	score = input.ComputeCaptureScore()
	s.Require().Equal(uint32(5722), score) // 8175 * 0.7
}

// ============================================================================
// Scoring Input Hash Tests
// ============================================================================

func (s *ScoringModelTestSuite) TestScoringInputs_ComputeInputHash_Deterministic() {
	inputs := types.ScoringInputs{
		FaceSimilarity: types.FaceSimilarityInput{
			SimilarityScore: 8500,
			Confidence:      9000,
			Present:         true,
		},
		OCRConfidence: types.OCRConfidenceInput{
			OverallConfidence:   7500,
			ExtractedFieldCount: 8,
			ExpectedFieldCount:  10,
			Present:             true,
		},
		DocIntegrity: types.DocIntegrityInput{
			QualityScore:          8000,
			FormatValid:           true,
			TamperDetectionPassed: true,
			Present:               true,
		},
		SaltBinding: types.SaltBindingInput{
			SaltPresent:          true,
			SaltValid:            true,
			ClientSignatureValid: true,
			UserSignatureValid:   true,
		},
		LivenessCheck: types.LivenessCheckInput{
			LivenessScore: 8000,
			Present:       true,
		},
		CaptureQuality: types.CaptureQualityInput{
			OverallQuality: 7500,
			Present:        true,
		},
		AccountAddress: "virtengine1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5z5tpwx",
		BlockHeight:    1000,
		Timestamp:      time.Now(),
	}

	hash1 := inputs.ComputeInputHash()
	hash2 := inputs.ComputeInputHash()

	s.Require().Equal(hash1, hash2, "hash should be deterministic")
	s.Require().Len(hash1, 32, "hash should be SHA256")
}

func (s *ScoringModelTestSuite) TestScoringInputs_ComputeInputHash_DifferentInputs() {
	inputs1 := types.ScoringInputs{
		FaceSimilarity: types.FaceSimilarityInput{SimilarityScore: 8500, Present: true},
		AccountAddress: "virtengine1test1",
		BlockHeight:    1000,
	}

	inputs2 := types.ScoringInputs{
		FaceSimilarity: types.FaceSimilarityInput{SimilarityScore: 8501, Present: true}, // Different
		AccountAddress: "virtengine1test1",
		BlockHeight:    1000,
	}

	hash1 := inputs1.ComputeInputHash()
	hash2 := inputs2.ComputeInputHash()

	s.Require().NotEqual(hash1, hash2, "different inputs should produce different hashes")
}

// ============================================================================
// Deterministic Score Computation Tests
// ============================================================================

func (s *ScoringModelTestSuite) TestComputeDeterministicScore_FullInputs() {
	model := types.DefaultScoringModelVersion()
	inputs := createFullValidInputs()

	summary, err := types.ComputeDeterministicScore(inputs, model)
	s.Require().NoError(err)
	s.Require().NotNil(summary)
	s.Require().True(summary.Passed, "full valid inputs should pass")
	s.Require().Greater(summary.FinalScore, uint32(50), "score should be above pass threshold")
	s.Require().Equal(model.Version, summary.ModelVersion)
	s.Require().Len(summary.Contributions, 6, "should have 6 feature contributions")
}

func (s *ScoringModelTestSuite) TestComputeDeterministicScore_MissingSelfie_RequiredFails() {
	model := types.DefaultScoringModelVersion()
	model.Config.RequireSelfie = true
	model.Config.AllowFallbackOnMissingSelfie = false

	inputs := createFullValidInputs()
	inputs.FaceSimilarity.Present = false

	summary, err := types.ComputeDeterministicScore(inputs, model)
	s.Require().NoError(err)
	s.Require().NotNil(summary)
	s.Require().False(summary.Passed, "missing required selfie should fail")
	s.Require().Equal(uint32(0), summary.FinalScore)
	s.Require().Contains(summary.ReasonCodes, types.ScoringReasonMissingSelfie)
}

func (s *ScoringModelTestSuite) TestComputeDeterministicScore_MissingSelfie_FallbackAllowed() {
	model := types.DefaultScoringModelVersion()
	model.Config.RequireSelfie = true
	model.Config.AllowFallbackOnMissingSelfie = true
	model.Thresholds.MissingSelfieMaxScore = 30

	inputs := createFullValidInputs()
	inputs.FaceSimilarity.Present = false

	summary, err := types.ComputeDeterministicScore(inputs, model)
	s.Require().NoError(err)
	s.Require().NotNil(summary)
	// Score should be capped at MissingSelfieMaxScore
	s.Require().LessOrEqual(summary.FinalScore, uint32(30))
	s.Require().Contains(summary.ReasonCodes, types.ScoringReasonFallbackApplied)
}

func (s *ScoringModelTestSuite) TestComputeDeterministicScore_MissingDocument_Capped() {
	model := types.DefaultScoringModelVersion()
	model.Config.RequireDocument = true
	model.Config.AllowFallbackOnMissingDoc = true
	model.Thresholds.MissingDocMaxScore = 30

	inputs := createFullValidInputs()
	inputs.DocIntegrity.Present = false

	summary, err := types.ComputeDeterministicScore(inputs, model)
	s.Require().NoError(err)
	s.Require().NotNil(summary)
	s.Require().LessOrEqual(summary.FinalScore, uint32(30))
	s.Require().Contains(summary.ReasonCodes, types.ScoringReasonFallbackApplied)
}

func (s *ScoringModelTestSuite) TestComputeDeterministicScore_LowFaceSimilarity() {
	model := types.DefaultScoringModelVersion()
	model.Thresholds.MinFaceSimilarity = 8000 // 80%

	inputs := createFullValidInputs()
	inputs.FaceSimilarity.SimilarityScore = 5000 // 50% - below threshold

	summary, err := types.ComputeDeterministicScore(inputs, model)
	s.Require().NoError(err)
	s.Require().Contains(summary.ReasonCodes, types.ScoringReasonLowFaceSimilarity)
}

func (s *ScoringModelTestSuite) TestComputeDeterministicScore_ScoreBoundaries() {
	model := types.DefaultScoringModelVersion()

	// Test minimum boundary
	inputs := types.ScoringInputs{
		FaceSimilarity: types.FaceSimilarityInput{SimilarityScore: 0, Confidence: 0, Present: true},
		OCRConfidence:  types.OCRConfidenceInput{OverallConfidence: 0, Present: true},
		DocIntegrity:   types.DocIntegrityInput{QualityScore: 0, Present: true},
		SaltBinding:    types.SaltBindingInput{SaltPresent: false},
		LivenessCheck:  types.LivenessCheckInput{LivenessScore: 0, Present: false},
		CaptureQuality: types.CaptureQualityInput{OverallQuality: 0, Present: false},
		BlockHeight:    1000,
		Timestamp:      time.Now(),
	}

	summary, err := types.ComputeDeterministicScore(inputs, model)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(summary.FinalScore, uint32(0))
	s.Require().LessOrEqual(summary.FinalScore, types.MaxScore)

	// Test maximum boundary
	inputs = types.ScoringInputs{
		FaceSimilarity: types.FaceSimilarityInput{SimilarityScore: 10000, Confidence: 10000, Present: true},
		OCRConfidence:  types.OCRConfidenceInput{OverallConfidence: 10000, ExtractedFieldCount: 10, ExpectedFieldCount: 10, Present: true},
		DocIntegrity:   types.DocIntegrityInput{QualityScore: 10000, FormatValid: true, TemplateMatch: true, TamperDetectionPassed: true, ExpiryValid: true, Present: true},
		SaltBinding:    types.SaltBindingInput{SaltPresent: true, SaltValid: true, ClientSignatureValid: true, UserSignatureValid: true},
		LivenessCheck:  types.LivenessCheckInput{LivenessScore: 10000, BlinkDetected: true, HeadMovementDetected: true, VideoFrameCount: 30, Present: true},
		CaptureQuality: types.CaptureQualityInput{OverallQuality: 10000, LightingScore: 10000, FocusScore: 10000, AngleScore: 10000, ResolutionAdequate: true, Present: true},
		BlockHeight:    1000,
		Timestamp:      time.Now(),
	}

	summary, err = types.ComputeDeterministicScore(inputs, model)
	s.Require().NoError(err)
	s.Require().LessOrEqual(summary.FinalScore, types.MaxScore)
}

func (s *ScoringModelTestSuite) TestComputeDeterministicScore_Determinism() {
	model := types.DefaultScoringModelVersion()
	inputs := createFullValidInputs()

	// Compute multiple times
	summary1, err1 := types.ComputeDeterministicScore(inputs, model)
	summary2, err2 := types.ComputeDeterministicScore(inputs, model)
	summary3, err3 := types.ComputeDeterministicScore(inputs, model)

	s.Require().NoError(err1)
	s.Require().NoError(err2)
	s.Require().NoError(err3)

	// All results should be identical
	s.Require().Equal(summary1.FinalScore, summary2.FinalScore)
	s.Require().Equal(summary2.FinalScore, summary3.FinalScore)
	s.Require().Equal(summary1.InputHash, summary2.InputHash)
	s.Require().Equal(summary2.InputHash, summary3.InputHash)
}

func (s *ScoringModelTestSuite) TestComputeDeterministicScore_InvalidInputs() {
	model := types.DefaultScoringModelVersion()
	inputs := types.ScoringInputs{
		FaceSimilarity: types.FaceSimilarityInput{
			SimilarityScore: 20000, // Invalid - exceeds max
			Present:         true,
		},
	}

	_, err := types.ComputeDeterministicScore(inputs, model)
	s.Require().Error(err)
}

func (s *ScoringModelTestSuite) TestComputeDeterministicScore_InvalidModel() {
	model := types.ScoringModelVersion{
		Version: "invalid",
		Weights: types.ScoringWeights{
			FaceSimilarityWeight: 5000, // Doesn't sum to 10000
		},
	}
	inputs := createFullValidInputs()

	_, err := types.ComputeDeterministicScore(inputs, model)
	s.Require().Error(err)
}

// ============================================================================
// Evidence Summary Tests
// ============================================================================

func (s *ScoringModelTestSuite) TestEvidenceSummary_ReasonCodes() {
	model := types.DefaultScoringModelVersion()
	model.Thresholds.RequiredForPass = 80 // High threshold

	inputs := createFullValidInputs()
	// Lower scores to fail threshold
	inputs.FaceSimilarity.SimilarityScore = 5000
	inputs.OCRConfidence.OverallConfidence = 5000

	summary, err := types.ComputeDeterministicScore(inputs, model)
	s.Require().NoError(err)
	s.Require().False(summary.Passed)
	s.Require().Contains(summary.ReasonCodes, types.ScoringReasonBelowPassThreshold)
}

func (s *ScoringModelTestSuite) TestEvidenceSummary_Contributions() {
	model := types.DefaultScoringModelVersion()
	inputs := createFullValidInputs()

	summary, err := types.ComputeDeterministicScore(inputs, model)
	s.Require().NoError(err)
	s.Require().Len(summary.Contributions, 6)

	// Check all features are represented
	featureNames := make(map[string]bool)
	for _, contrib := range summary.Contributions {
		featureNames[contrib.FeatureName] = true
	}

	s.Require().True(featureNames[types.FeatureNameFaceSimilarity])
	s.Require().True(featureNames[types.FeatureNameOCRConfidence])
	s.Require().True(featureNames[types.FeatureNameDocIntegrity])
	s.Require().True(featureNames[types.FeatureNameSaltBinding])
	s.Require().True(featureNames[types.FeatureNameLivenessCheck])
	s.Require().True(featureNames[types.FeatureNameCaptureQuality])
}

// ============================================================================
// Fixed-Point Arithmetic Tests
// ============================================================================

func (s *ScoringModelTestSuite) TestToBasisPoints() {
	s.Require().Equal(uint32(0), types.ToBasisPoints(0.0))
	s.Require().Equal(uint32(5000), types.ToBasisPoints(0.5))
	s.Require().Equal(uint32(10000), types.ToBasisPoints(1.0))
	s.Require().Equal(uint32(0), types.ToBasisPoints(-0.5)) // Negative clamped
	s.Require().Equal(uint32(10000), types.ToBasisPoints(1.5)) // Over 1 clamped
}

func (s *ScoringModelTestSuite) TestFromBasisPoints() {
	s.Require().Equal(uint32(0), types.FromBasisPoints(0))
	s.Require().Equal(uint32(50), types.FromBasisPoints(5000))
	s.Require().Equal(uint32(100), types.FromBasisPoints(10000))
	s.Require().Equal(uint32(75), types.FromBasisPoints(7500))
}

// ============================================================================
// Scoring Model Version Tests
// ============================================================================

func (s *ScoringModelTestSuite) TestScoringModelVersion_Validate() {
	model := types.DefaultScoringModelVersion()
	s.Require().NoError(model.Validate())

	// Empty version
	model.Version = ""
	s.Require().Error(model.Validate())
}

func (s *ScoringModelTestSuite) TestScoringModelVersion_ComputeModelHash() {
	model := types.DefaultScoringModelVersion()
	hash1 := model.ComputeModelHash()
	hash2 := model.ComputeModelHash()

	s.Require().Equal(hash1, hash2, "hash should be deterministic")
	s.Require().Len(hash1, 32)

	// Different weights should produce different hash
	model.Weights.FaceSimilarityWeight = 3001
	model.Weights.CaptureQualityWeight = 499 // Adjust to keep sum at 10000
	hash3 := model.ComputeModelHash()

	s.Require().NotEqual(hash1, hash3)
}

// ============================================================================
// Helper Functions
// ============================================================================

func createFullValidInputs() types.ScoringInputs {
	return types.ScoringInputs{
		FaceSimilarity: types.FaceSimilarityInput{
			SimilarityScore: 8500,
			Confidence:      9000,
			Present:         true,
		},
		OCRConfidence: types.OCRConfidenceInput{
			OverallConfidence:   8000,
			ExtractedFieldCount: 8,
			ExpectedFieldCount:  10,
			FieldConfidences: map[string]uint32{
				"name": 9000,
				"dob":  8500,
			},
			Present: true,
		},
		DocIntegrity: types.DocIntegrityInput{
			QualityScore:          8500,
			FormatValid:           true,
			TemplateMatch:         true,
			TamperDetectionPassed: true,
			ExpiryValid:           true,
			SharpnessScore:        8000,
			BrightnessScore:       7500,
			ContrastScore:         8000,
			Present:               true,
		},
		SaltBinding: types.SaltBindingInput{
			SaltPresent:          true,
			SaltValid:            true,
			ClientSignatureValid: true,
			UserSignatureValid:   true,
		},
		LivenessCheck: types.LivenessCheckInput{
			LivenessScore:        8000,
			BlinkDetected:        true,
			HeadMovementDetected: true,
			VideoFrameCount:      30,
			Present:              true,
		},
		CaptureQuality: types.CaptureQualityInput{
			OverallQuality:     8000,
			LightingScore:      7500,
			FocusScore:         8500,
			AngleScore:         8000,
			ResolutionAdequate: true,
			Present:            true,
		},
		AccountAddress: "virtengine1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5z5tpwx",
		BlockHeight:    1000,
		Timestamp:      time.Now(),
	}
}
