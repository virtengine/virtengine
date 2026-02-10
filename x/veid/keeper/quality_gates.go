package keeper

import (
	"fmt"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Quality Gates Configuration
// ============================================================================

// QualityGatesConfig configures quality threshold enforcement
type QualityGatesConfig struct {
	// Face quality thresholds
	MinFaceConfidence float32
	MinFaceSharpness  float32
	MinFaceBrightness float32
	MaxFaceBrightness float32
	MinFaceSize       float32

	// Document quality thresholds
	MinDocSharpness  float32
	MinDocBrightness float32
	MaxDocBrightness float32
	MinDocContrast   float32
	MaxDocNoiseLevel float32
	MaxDocBlurScore  float32

	// OCR thresholds
	MinOCRConfidence       float32
	MinValidatedFieldRatio float32

	// Liveness thresholds
	MinLivenessScore float32

	// Enable/disable individual gates
	EnableFaceGates     bool
	EnableDocGates      bool
	EnableOCRGates      bool
	EnableLivenessGates bool
}

// DefaultQualityGatesConfig returns default quality gate configuration
func DefaultQualityGatesConfig() QualityGatesConfig {
	return QualityGatesConfig{
		// Face thresholds
		MinFaceConfidence: 0.7,
		MinFaceSharpness:  0.4,
		MinFaceBrightness: 0.2,
		MaxFaceBrightness: 0.9,
		MinFaceSize:       0.1, // Face should be at least 10% of image

		// Document thresholds
		MinDocSharpness:  0.4,
		MinDocBrightness: 0.2,
		MaxDocBrightness: 0.9,
		MinDocContrast:   0.3,
		MaxDocNoiseLevel: 0.4,
		MaxDocBlurScore:  0.5,

		// OCR thresholds
		MinOCRConfidence:       0.6,
		MinValidatedFieldRatio: 0.4, // At least 40% of fields must validate

		// Liveness thresholds
		MinLivenessScore: 0.5,

		// Enable all gates by default
		EnableFaceGates:     true,
		EnableDocGates:      true,
		EnableOCRGates:      true,
		EnableLivenessGates: true,
	}
}

// ============================================================================
// Quality Gate Types
// ============================================================================

// QualityGateType identifies the type of quality gate
type QualityGateType string

const (
	QualityGateFaceConfidence   QualityGateType = "face_confidence"
	QualityGateFaceSharpness    QualityGateType = "face_sharpness"
	QualityGateFaceBrightness   QualityGateType = "face_brightness"
	QualityGateFaceSize         QualityGateType = "face_size"
	QualityGateDocSharpness     QualityGateType = "doc_sharpness"
	QualityGateDocBrightness    QualityGateType = "doc_brightness"
	QualityGateDocContrast      QualityGateType = "doc_contrast"
	QualityGateDocNoise         QualityGateType = "doc_noise"
	QualityGateDocBlur          QualityGateType = "doc_blur"
	QualityGateOCRConfidence    QualityGateType = "ocr_confidence"
	QualityGateOCRValidation    QualityGateType = "ocr_validation"
	QualityGateLivenessScore    QualityGateType = "liveness_score"
	QualityGateLivenessDecision QualityGateType = "liveness_decision"
)

// QualityGateResult represents the result of a quality gate check
type QualityGateResult struct {
	// Gate is the type of quality gate
	Gate QualityGateType

	// Passed indicates whether the gate passed
	Passed bool

	// Value is the actual measured value
	Value float32

	// Threshold is the threshold that was applied
	Threshold float32

	// Operator is "min", "max", or "range"
	Operator string

	// ReasonCode is the failure reason code if not passed
	ReasonCode types.ReasonCode

	// Details provides human-readable explanation
	Details string
}

// ============================================================================
// Quality Gates Implementation
// ============================================================================

// QualityGates enforces quality thresholds on extracted features
type QualityGates struct {
	config QualityGatesConfig
}

// NewQualityGates creates a new quality gates checker
func NewQualityGates(config QualityGatesConfig) *QualityGates {
	return &QualityGates{
		config: config,
	}
}

// Check runs all quality gates on extracted features
func (qg *QualityGates) Check(features *RealExtractedFeatures) []QualityGateResult {
	results := make([]QualityGateResult, 0)

	// Face quality gates
	if qg.config.EnableFaceGates && len(features.FaceEmbedding) > 0 {
		results = append(results, qg.checkFaceGates(features)...)
	}

	// Document quality gates
	if qg.config.EnableDocGates && features.DocQualityScore > 0 {
		results = append(results, qg.checkDocumentGates(features)...)
	}

	// OCR quality gates
	if qg.config.EnableOCRGates && len(features.OCRConfidences) > 0 {
		results = append(results, qg.checkOCRGates(features)...)
	}

	// Liveness quality gates
	if qg.config.EnableLivenessGates && features.LivenessScore > 0 {
		results = append(results, qg.checkLivenessGates(features)...)
	}

	return results
}

// checkFaceGates checks face-related quality gates
func (qg *QualityGates) checkFaceGates(features *RealExtractedFeatures) []QualityGateResult {
	results := make([]QualityGateResult, 0, 4)

	// Face confidence gate
	results = append(results, qg.checkMinThreshold(
		QualityGateFaceConfidence,
		features.FaceConfidence,
		qg.config.MinFaceConfidence,
		types.ReasonCodeFaceMismatch,
		"Face detection confidence below threshold",
	))

	// Face sharpness gate
	results = append(results, qg.checkMinThreshold(
		QualityGateFaceSharpness,
		features.FaceQuality.Sharpness,
		qg.config.MinFaceSharpness,
		types.ReasonCodeLowDocQuality,
		"Face image too blurry",
	))

	// Face brightness gate (range check)
	brightnessResult := qg.checkRangeThreshold(
		QualityGateFaceBrightness,
		features.FaceQuality.Brightness,
		qg.config.MinFaceBrightness,
		qg.config.MaxFaceBrightness,
		types.ReasonCodeLowDocQuality,
	)
	results = append(results, brightnessResult)

	// Face size gate
	results = append(results, qg.checkMinThreshold(
		QualityGateFaceSize,
		features.FaceQuality.FaceSize,
		qg.config.MinFaceSize,
		types.ReasonCodeFaceMismatch,
		"Face too small in image",
	))

	return results
}

// checkDocumentGates checks document quality gates
func (qg *QualityGates) checkDocumentGates(features *RealExtractedFeatures) []QualityGateResult {
	results := make([]QualityGateResult, 0, 5)

	// Document sharpness gate
	results = append(results, qg.checkMinThreshold(
		QualityGateDocSharpness,
		features.DocQualityFeatures.Sharpness,
		qg.config.MinDocSharpness,
		types.ReasonCodeLowDocQuality,
		"Document image too blurry",
	))

	// Document brightness gate (range check)
	results = append(results, qg.checkRangeThreshold(
		QualityGateDocBrightness,
		features.DocQualityFeatures.Brightness,
		qg.config.MinDocBrightness,
		qg.config.MaxDocBrightness,
		types.ReasonCodeLowDocQuality,
	))

	// Document contrast gate
	results = append(results, qg.checkMinThreshold(
		QualityGateDocContrast,
		features.DocQualityFeatures.Contrast,
		qg.config.MinDocContrast,
		types.ReasonCodeLowDocQuality,
		"Document image has low contrast",
	))

	// Document noise gate (max threshold)
	results = append(results, qg.checkMaxThreshold(
		QualityGateDocNoise,
		features.DocQualityFeatures.NoiseLevel,
		qg.config.MaxDocNoiseLevel,
		types.ReasonCodeLowDocQuality,
		"Document image is too noisy",
	))

	// Document blur gate (max threshold)
	results = append(results, qg.checkMaxThreshold(
		QualityGateDocBlur,
		features.DocQualityFeatures.BlurScore,
		qg.config.MaxDocBlurScore,
		types.ReasonCodeLowDocQuality,
		"Document image is too blurry",
	))

	return results
}

// checkOCRGates checks OCR quality gates
func (qg *QualityGates) checkOCRGates(features *RealExtractedFeatures) []QualityGateResult {
	results := make([]QualityGateResult, 0, 2)

	// Check average OCR confidence
	if len(features.OCRConfidences) > 0 {
		var sum float32
		for _, conf := range features.OCRConfidences {
			sum += conf
		}
		avgConf := sum / float32(len(features.OCRConfidences))

		results = append(results, qg.checkMinThreshold(
			QualityGateOCRConfidence,
			avgConf,
			qg.config.MinOCRConfidence,
			types.ReasonCodeLowOCRConfidence,
			"OCR confidence below threshold",
		))
	}

	// Check validated field ratio
	if len(features.OCRFieldValidation) > 0 {
		var validCount int
		for _, valid := range features.OCRFieldValidation {
			if valid {
				validCount++
			}
		}
		validRatio := float32(validCount) / float32(len(features.OCRFieldValidation))

		results = append(results, qg.checkMinThreshold(
			QualityGateOCRValidation,
			validRatio,
			qg.config.MinValidatedFieldRatio,
			types.ReasonCodeLowOCRConfidence,
			"Too many OCR fields failed validation",
		))
	}

	return results
}

// checkLivenessGates checks liveness quality gates
func (qg *QualityGates) checkLivenessGates(features *RealExtractedFeatures) []QualityGateResult {
	results := make([]QualityGateResult, 0, 2)

	// Liveness score gate
	results = append(results, qg.checkMinThreshold(
		QualityGateLivenessScore,
		features.LivenessScore,
		qg.config.MinLivenessScore,
		types.ReasonCodeLivenessCheckFailed,
		"Liveness score below threshold",
	))

	// Liveness decision gate
	livenessDecisionPassed := features.LivenessDecision == "live"
	details := fmt.Sprintf("Liveness decision: %s", features.LivenessDecision)
	reasonCode := types.ReasonCodeSuccess
	if !livenessDecisionPassed {
		reasonCode = types.ReasonCodeLivenessCheckFailed
	}

	results = append(results, QualityGateResult{
		Gate:       QualityGateLivenessDecision,
		Passed:     livenessDecisionPassed,
		Value:      features.LivenessScore,
		Threshold:  0.0, // N/A for decision gate
		Operator:   "decision",
		ReasonCode: reasonCode,
		Details:    details,
	})

	return results
}

// ============================================================================
// Threshold Helpers
// ============================================================================

func (qg *QualityGates) checkMinThreshold(
	gate QualityGateType,
	value float32,
	threshold float32,
	failCode types.ReasonCode,
	failMessage string,
) QualityGateResult {
	passed := value >= threshold
	reasonCode := types.ReasonCodeSuccess
	details := fmt.Sprintf("%.3f >= %.3f", value, threshold)

	if !passed {
		reasonCode = failCode
		details = fmt.Sprintf("%s (%.3f < %.3f)", failMessage, value, threshold)
	}

	return QualityGateResult{
		Gate:       gate,
		Passed:     passed,
		Value:      value,
		Threshold:  threshold,
		Operator:   "min",
		ReasonCode: reasonCode,
		Details:    details,
	}
}

func (qg *QualityGates) checkMaxThreshold(
	gate QualityGateType,
	value float32,
	threshold float32,
	failCode types.ReasonCode,
	failMessage string,
) QualityGateResult {
	passed := value <= threshold
	reasonCode := types.ReasonCodeSuccess
	details := fmt.Sprintf("%.3f <= %.3f", value, threshold)

	if !passed {
		reasonCode = failCode
		details = fmt.Sprintf("%s (%.3f > %.3f)", failMessage, value, threshold)
	}

	return QualityGateResult{
		Gate:       gate,
		Passed:     passed,
		Value:      value,
		Threshold:  threshold,
		Operator:   "max",
		ReasonCode: reasonCode,
		Details:    details,
	}
}

func (qg *QualityGates) checkRangeThreshold(
	gate QualityGateType,
	value float32,
	minThreshold float32,
	maxThreshold float32,
	failCode types.ReasonCode,
) QualityGateResult {
	passed := value >= minThreshold && value <= maxThreshold
	reasonCode := types.ReasonCodeSuccess
	details := fmt.Sprintf("%.3f in [%.3f, %.3f]", value, minThreshold, maxThreshold)

	if !passed {
		reasonCode = failCode
		if value < minThreshold {
			details = fmt.Sprintf("Value %.3f below minimum %.3f", value, minThreshold)
		} else {
			details = fmt.Sprintf("Value %.3f above maximum %.3f", value, maxThreshold)
		}
	}

	return QualityGateResult{
		Gate:       gate,
		Passed:     passed,
		Value:      value,
		Threshold:  minThreshold, // Store min as primary threshold
		Operator:   "range",
		ReasonCode: reasonCode,
		Details:    details,
	}
}

// ============================================================================
// Aggregate Results
// ============================================================================

// AllPassed returns true if all quality gates passed
func AllQualityGatesPassed(results []QualityGateResult) bool {
	for _, r := range results {
		if !r.Passed {
			return false
		}
	}
	return true
}

// GetFailedGates returns all failed quality gates
func GetFailedGates(results []QualityGateResult) []QualityGateResult {
	failed := make([]QualityGateResult, 0)
	for _, r := range results {
		if !r.Passed {
			failed = append(failed, r)
		}
	}
	return failed
}

// GetFailureReasonCodes returns reason codes for all failed gates
func GetFailureReasonCodes(results []QualityGateResult) []types.ReasonCode {
	codes := make([]types.ReasonCode, 0)
	seen := make(map[types.ReasonCode]bool)

	for _, r := range results {
		if !r.Passed && !seen[r.ReasonCode] {
			codes = append(codes, r.ReasonCode)
			seen[r.ReasonCode] = true
		}
	}
	return codes
}

// CountPassedGates returns the count of passed gates
func CountPassedGates(results []QualityGateResult) (passed, total int) {
	total = len(results)
	for _, r := range results {
		if r.Passed {
			passed++
		}
	}
	return passed, total
}
