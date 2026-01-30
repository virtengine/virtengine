package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/virtengine/virtengine/pkg/inference"
	"github.com/virtengine/virtengine/x/veid/types"
)

func TestQualityGates_DefaultConfig(t *testing.T) {
	config := DefaultQualityGatesConfig()

	// Verify reasonable defaults
	assert.Equal(t, float32(0.7), config.MinFaceConfidence)
	assert.Equal(t, float32(0.4), config.MinFaceSharpness)
	assert.Equal(t, float32(0.2), config.MinFaceBrightness)
	assert.Equal(t, float32(0.9), config.MaxFaceBrightness)
	assert.Equal(t, float32(0.4), config.MinDocSharpness)
	assert.Equal(t, float32(0.6), config.MinOCRConfidence)
	assert.Equal(t, float32(0.5), config.MinLivenessScore)
	assert.True(t, config.EnableFaceGates)
	assert.True(t, config.EnableDocGates)
	assert.True(t, config.EnableOCRGates)
	assert.True(t, config.EnableLivenessGates)
}

func TestQualityGates_FaceGates(t *testing.T) {
	gates := NewQualityGates(DefaultQualityGatesConfig())

	tests := []struct {
		name           string
		features       *RealExtractedFeatures
		wantPassCount  int
		wantFailCount  int
		failedGateType QualityGateType
	}{
		{
			name: "all face gates pass",
			features: &RealExtractedFeatures{
				FaceEmbedding:  make([]float32, types.FaceEmbeddingDim),
				FaceConfidence: 0.85,
				FaceQuality: FaceQualityMetrics{
					Sharpness:  0.7,
					Brightness: 0.5,
					FaceSize:   0.2,
				},
			},
			wantPassCount: 4, // confidence, sharpness, brightness, size
			wantFailCount: 0,
		},
		{
			name: "low face confidence",
			features: &RealExtractedFeatures{
				FaceEmbedding:  make([]float32, types.FaceEmbeddingDim),
				FaceConfidence: 0.5, // Below 0.7 threshold
				FaceQuality: FaceQualityMetrics{
					Sharpness:  0.7,
					Brightness: 0.5,
					FaceSize:   0.2,
				},
			},
			wantPassCount:  3,
			wantFailCount:  1,
			failedGateType: QualityGateFaceConfidence,
		},
		{
			name: "face too blurry",
			features: &RealExtractedFeatures{
				FaceEmbedding:  make([]float32, types.FaceEmbeddingDim),
				FaceConfidence: 0.85,
				FaceQuality: FaceQualityMetrics{
					Sharpness:  0.2, // Below 0.4 threshold
					Brightness: 0.5,
					FaceSize:   0.2,
				},
			},
			wantPassCount:  3,
			wantFailCount:  1,
			failedGateType: QualityGateFaceSharpness,
		},
		{
			name: "face too dark",
			features: &RealExtractedFeatures{
				FaceEmbedding:  make([]float32, types.FaceEmbeddingDim),
				FaceConfidence: 0.85,
				FaceQuality: FaceQualityMetrics{
					Sharpness:  0.7,
					Brightness: 0.1, // Below 0.2 threshold
					FaceSize:   0.2,
				},
			},
			wantPassCount:  3,
			wantFailCount:  1,
			failedGateType: QualityGateFaceBrightness,
		},
		{
			name: "face too bright",
			features: &RealExtractedFeatures{
				FaceEmbedding:  make([]float32, types.FaceEmbeddingDim),
				FaceConfidence: 0.85,
				FaceQuality: FaceQualityMetrics{
					Sharpness:  0.7,
					Brightness: 0.95, // Above 0.9 threshold
					FaceSize:   0.2,
				},
			},
			wantPassCount:  3,
			wantFailCount:  1,
			failedGateType: QualityGateFaceBrightness,
		},
		{
			name: "face too small",
			features: &RealExtractedFeatures{
				FaceEmbedding:  make([]float32, types.FaceEmbeddingDim),
				FaceConfidence: 0.85,
				FaceQuality: FaceQualityMetrics{
					Sharpness:  0.7,
					Brightness: 0.5,
					FaceSize:   0.05, // Below 0.1 threshold
				},
			},
			wantPassCount:  3,
			wantFailCount:  1,
			failedGateType: QualityGateFaceSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Only check face gates
			config := DefaultQualityGatesConfig()
			config.EnableDocGates = false
			config.EnableOCRGates = false
			config.EnableLivenessGates = false
			gates := NewQualityGates(config)

			results := gates.Check(tt.features)

			passCount, totalCount := CountPassedGates(results)
			assert.Equal(t, tt.wantPassCount, passCount)
			assert.Equal(t, tt.wantFailCount, totalCount-passCount)

			if tt.wantFailCount > 0 {
				failed := GetFailedGates(results)
				assert.NotEmpty(t, failed)
				assert.Equal(t, tt.failedGateType, failed[0].Gate)
			}
		})
	}
}

func TestQualityGates_DocumentGates(t *testing.T) {
	config := DefaultQualityGatesConfig()
	config.EnableFaceGates = false
	config.EnableOCRGates = false
	config.EnableLivenessGates = false
	gates := NewQualityGates(config)

	tests := []struct {
		name           string
		docQuality     inference.DocQualityFeatures
		expectAllPass  bool
		failedGateType QualityGateType
	}{
		{
			name: "all doc gates pass",
			docQuality: inference.DocQualityFeatures{
				Sharpness:  0.8,
				Brightness: 0.6,
				Contrast:   0.7,
				NoiseLevel: 0.1,
				BlurScore:  0.2,
			},
			expectAllPass: true,
		},
		{
			name: "document too blurry (sharpness)",
			docQuality: inference.DocQualityFeatures{
				Sharpness:  0.2, // Below 0.4
				Brightness: 0.6,
				Contrast:   0.7,
				NoiseLevel: 0.1,
				BlurScore:  0.2,
			},
			expectAllPass:  false,
			failedGateType: QualityGateDocSharpness,
		},
		{
			name: "document too noisy",
			docQuality: inference.DocQualityFeatures{
				Sharpness:  0.8,
				Brightness: 0.6,
				Contrast:   0.7,
				NoiseLevel: 0.6, // Above 0.4
				BlurScore:  0.2,
			},
			expectAllPass:  false,
			failedGateType: QualityGateDocNoise,
		},
		{
			name: "document blur too high",
			docQuality: inference.DocQualityFeatures{
				Sharpness:  0.8,
				Brightness: 0.6,
				Contrast:   0.7,
				NoiseLevel: 0.1,
				BlurScore:  0.7, // Above 0.5
			},
			expectAllPass:  false,
			failedGateType: QualityGateDocBlur,
		},
		{
			name: "document low contrast",
			docQuality: inference.DocQualityFeatures{
				Sharpness:  0.8,
				Brightness: 0.6,
				Contrast:   0.2, // Below 0.3
				NoiseLevel: 0.1,
				BlurScore:  0.2,
			},
			expectAllPass:  false,
			failedGateType: QualityGateDocContrast,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features := &RealExtractedFeatures{
				DocQualityScore:    0.5, // Non-zero to enable doc gates
				DocQualityFeatures: tt.docQuality,
			}

			results := gates.Check(features)
			allPassed := AllQualityGatesPassed(results)
			assert.Equal(t, tt.expectAllPass, allPassed)

			if !tt.expectAllPass {
				failed := GetFailedGates(results)
				found := false
				for _, f := range failed {
					if f.Gate == tt.failedGateType {
						found = true
						break
					}
				}
				assert.True(t, found, "expected gate %s to fail", tt.failedGateType)
			}
		})
	}
}

func TestQualityGates_OCRGates(t *testing.T) {
	config := DefaultQualityGatesConfig()
	config.EnableFaceGates = false
	config.EnableDocGates = false
	config.EnableLivenessGates = false
	gates := NewQualityGates(config)

	tests := []struct {
		name          string
		ocrConf       map[string]float32
		ocrValid      map[string]bool
		expectAllPass bool
	}{
		{
			name: "high OCR confidence all valid",
			ocrConf: map[string]float32{
				"name": 0.9, "date_of_birth": 0.85, "document_number": 0.88,
			},
			ocrValid: map[string]bool{
				"name": true, "date_of_birth": true, "document_number": true,
			},
			expectAllPass: true,
		},
		{
			name: "low OCR confidence",
			ocrConf: map[string]float32{
				"name": 0.4, "date_of_birth": 0.35, "document_number": 0.38,
			},
			ocrValid: map[string]bool{
				"name": true, "date_of_birth": true, "document_number": true,
			},
			expectAllPass: false,
		},
		{
			name: "low validation ratio",
			ocrConf: map[string]float32{
				"name": 0.9, "date_of_birth": 0.85, "document_number": 0.88, "expiry": 0.9, "nationality": 0.85,
			},
			ocrValid: map[string]bool{
				"name": false, "date_of_birth": false, "document_number": false, "expiry": false, "nationality": true,
			},
			expectAllPass: false, // Only 20% validated, below 40% threshold
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features := &RealExtractedFeatures{
				OCRConfidences:     tt.ocrConf,
				OCRFieldValidation: tt.ocrValid,
			}

			results := gates.Check(features)
			allPassed := AllQualityGatesPassed(results)
			assert.Equal(t, tt.expectAllPass, allPassed)
		})
	}
}

func TestQualityGates_LivenessGates(t *testing.T) {
	config := DefaultQualityGatesConfig()
	config.EnableFaceGates = false
	config.EnableDocGates = false
	config.EnableOCRGates = false
	gates := NewQualityGates(config)

	tests := []struct {
		name             string
		livenessScore    float32
		livenessDecision string
		expectAllPass    bool
		expectFailCount  int
	}{
		{
			name:             "liveness passed",
			livenessScore:    0.85,
			livenessDecision: "live",
			expectAllPass:    true,
			expectFailCount:  0,
		},
		{
			name:             "liveness score too low",
			livenessScore:    0.3, // Below 0.5 threshold
			livenessDecision: "uncertain",
			expectAllPass:    false,
			expectFailCount:  2, // score and decision
		},
		{
			name:             "spoof detected",
			livenessScore:    0.2,
			livenessDecision: "spoof",
			expectAllPass:    false,
			expectFailCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features := &RealExtractedFeatures{
				LivenessScore:    tt.livenessScore,
				LivenessDecision: tt.livenessDecision,
			}

			results := gates.Check(features)
			allPassed := AllQualityGatesPassed(results)
			assert.Equal(t, tt.expectAllPass, allPassed)

			if !tt.expectAllPass {
				failed := GetFailedGates(results)
				assert.Equal(t, tt.expectFailCount, len(failed))
			}
		})
	}
}

func TestQualityGates_DisabledGates(t *testing.T) {
	config := DefaultQualityGatesConfig()
	config.EnableFaceGates = false
	config.EnableDocGates = false
	config.EnableOCRGates = false
	config.EnableLivenessGates = false
	gates := NewQualityGates(config)

	features := &RealExtractedFeatures{
		FaceConfidence:   0.3, // Would fail if enabled
		LivenessScore:    0.2, // Would fail if enabled
		LivenessDecision: "spoof",
	}

	results := gates.Check(features)
	assert.Empty(t, results, "disabled gates should produce no results")
}

func TestQualityGateResult_ReasonCodeMapping(t *testing.T) {
	gates := NewQualityGates(DefaultQualityGatesConfig())

	features := &RealExtractedFeatures{
		FaceEmbedding:  make([]float32, types.FaceEmbeddingDim),
		FaceConfidence: 0.3, // Low
		FaceQuality:    FaceQualityMetrics{Sharpness: 0.7, Brightness: 0.5, FaceSize: 0.2},
		DocQualityScore: 0.5,
		DocQualityFeatures: inference.DocQualityFeatures{
			Sharpness: 0.2, Brightness: 0.6, Contrast: 0.2, NoiseLevel: 0.6, BlurScore: 0.8,
		},
		OCRConfidences:     map[string]float32{"name": 0.3},
		OCRFieldValidation: map[string]bool{"name": false},
		LivenessScore:      0.2,
		LivenessDecision:   "spoof",
	}

	results := gates.Check(features)
	codes := GetFailureReasonCodes(results)

	// Should have multiple different reason codes
	assert.True(t, len(codes) >= 3)

	// Verify specific reason codes
	codeSet := make(map[types.ReasonCode]bool)
	for _, c := range codes {
		codeSet[c] = true
	}

	assert.True(t, codeSet[types.ReasonCodeFaceMismatch], "should have face mismatch code")
	assert.True(t, codeSet[types.ReasonCodeLowDocQuality], "should have low doc quality code")
	assert.True(t, codeSet[types.ReasonCodeLivenessCheckFailed], "should have liveness failed code")
}
