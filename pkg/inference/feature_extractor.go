package inference

import (
	"fmt"
	"math"
)

// ============================================================================
// Feature Extractor
// ============================================================================

// FeatureExtractor transforms raw score inputs into the feature vector
// expected by the TensorFlow model.
//
// The output feature vector matches the training configuration:
// - 512 dimensions: face embedding
// - 1 dimension: face confidence
// - 6 dimensions: document quality features
// - 10 dimensions: OCR features (5 fields * 2)
// - 16 dimensions: metadata features
// - 223 dimensions: padding/reserved
// Total: 768 dimensions
type FeatureExtractor struct {
	// config holds extraction configuration
	config FeatureExtractorConfig
}

// FeatureExtractorConfig contains configuration for feature extraction
type FeatureExtractorConfig struct {
	// FaceEmbeddingDim is the expected face embedding dimension
	FaceEmbeddingDim int

	// NormalizeFeatures enables feature normalization
	NormalizeFeatures bool

	// FeatureMean is the mean for normalization (from training)
	FeatureMean []float32

	// FeatureStd is the std deviation for normalization (from training)
	FeatureStd []float32

	// OCRFieldNames is the list of expected OCR field names
	OCRFieldNames []string

	// RequireCompleteInputs rejects missing or invalid required production inputs.
	RequireCompleteInputs bool
}

// DefaultFeatureExtractorConfig returns the default configuration
func DefaultFeatureExtractorConfig() FeatureExtractorConfig {
	return FeatureExtractorConfig{
		FaceEmbeddingDim:      FaceEmbeddingDim,
		NormalizeFeatures:     false, // Normalization should be done during training
		FeatureMean:           nil,
		FeatureStd:            nil,
		OCRFieldNames:         OCRFieldNames,
		RequireCompleteInputs: false,
	}
}

// ProductionFeatureExtractorConfig returns fail-closed serving configuration.
func ProductionFeatureExtractorConfig() FeatureExtractorConfig {
	config := DefaultFeatureExtractorConfig()
	config.RequireCompleteInputs = true
	return config
}

// NewFeatureExtractor creates a new feature extractor
func NewFeatureExtractor(config FeatureExtractorConfig) *FeatureExtractor {
	return &FeatureExtractor{
		config: config,
	}
}

// ============================================================================
// Feature Extraction
// ============================================================================

// ExtractFeatures transforms ScoreInputs into a feature vector for model inference
func (fe *FeatureExtractor) ExtractFeatures(inputs *ScoreInputs) ([]float32, error) {
	if inputs == nil {
		return nil, fmt.Errorf("score inputs are required")
	}
	if issues := fe.validateInputs(inputs, fe.config.RequireCompleteInputs); len(issues) > 0 {
		return nil, fmt.Errorf("invalid feature inputs: %s", issues[0])
	}

	features := make([]float32, TotalFeatureDim)

	if err := fe.extractFaceFeatures(features, inputs); err != nil {
		return nil, fmt.Errorf("face feature extraction failed: %w", err)
	}
	fe.extractDocQualityFeatures(features, inputs)
	fe.extractOCRFeatures(features, inputs)
	fe.extractMetadataFeatures(features, inputs)

	if fe.config.NormalizeFeatures {
		if err := fe.normalizeFeatures(features); err != nil {
			if fe.config.RequireCompleteInputs {
				return nil, err
			}
		}
	}
	for index, value := range features {
		if !isFiniteFloat32(value) {
			return nil, fmt.Errorf("feature %d must be finite", index)
		}
	}

	return features, nil
}

func validateFiniteFeatureInputs(inputs *ScoreInputs) error {
	check := func(name string, value float32) error {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return fmt.Errorf("%s must be finite", name)
		}
		return nil
	}

	for i, value := range inputs.FaceEmbedding {
		if err := check(fmt.Sprintf("face embedding[%d]", i), value); err != nil {
			return err
		}
	}
	values := []struct {
		name  string
		value float32
	}{
		{"face confidence", inputs.FaceConfidence},
		{"document quality score", inputs.DocQualityScore},
		{"document sharpness", inputs.DocQualityFeatures.Sharpness},
		{"document brightness", inputs.DocQualityFeatures.Brightness},
		{"document contrast", inputs.DocQualityFeatures.Contrast},
		{"document noise level", inputs.DocQualityFeatures.NoiseLevel},
	}
	for _, item := range values {
		if err := check(item.name, item.value); err != nil {
			return err
		}
	}
	for field, value := range inputs.OCRConfidences {
		if err := check(fmt.Sprintf("OCR confidence[%s]", field), value); err != nil {
			return err
		}
	}
	return nil
}

// extractFaceFeatures extracts face embedding and confidence
func (fe *FeatureExtractor) extractFaceFeatures(features []float32, inputs *ScoreInputs) error {
	// Validate face embedding dimension
	if len(inputs.FaceEmbedding) == 0 {
		features[FaceConfidenceOffset] = inputs.FaceConfidence
		return nil
	}

	if len(inputs.FaceEmbedding) != fe.config.FaceEmbeddingDim {
		return fmt.Errorf(
			"face embedding dimension mismatch: expected %d, got %d",
			fe.config.FaceEmbeddingDim,
			len(inputs.FaceEmbedding),
		)
	}

	copy(features[SelfieEmbeddingOffset:FaceConfidenceOffset], inputs.FaceEmbedding)
	fe.normalizeEmbedding(features[SelfieEmbeddingOffset:FaceConfidenceOffset])
	features[FaceConfidenceOffset] = inputs.FaceConfidence
	return nil
}

// normalizeEmbedding normalizes an embedding to unit length
func (fe *FeatureExtractor) normalizeEmbedding(embedding []float32) {
	var sumSquares float64
	for _, v := range embedding {
		sumSquares += float64(v) * float64(v)
	}

	norm := math.Sqrt(sumSquares)
	if norm > 1e-10 {
		for i := range embedding {
			embedding[i] = float32(float64(embedding[i]) / norm)
		}
	}
}

// extractDocQualityFeatures extracts document quality features
func (fe *FeatureExtractor) extractDocQualityFeatures(features []float32, inputs *ScoreInputs) {
	// Overall document quality score
	features[DocQualityOffset] = inputs.DocQualityScore

	// Individual quality features
	features[DocQualityOffset+1] = inputs.DocQualityFeatures.Sharpness
	features[DocQualityOffset+2] = inputs.DocQualityFeatures.Brightness
	features[DocQualityOffset+3] = inputs.DocQualityFeatures.Contrast
	// Invert noise and blur so higher is better
	features[DocQualityOffset+4] = 1.0 - inputs.DocQualityFeatures.NoiseLevel
	features[DocQualityOffset+5] = 1.0 - inputs.DocQualityFeatures.BlurScore
}

// extractOCRFeatures extracts OCR-related features
func (fe *FeatureExtractor) extractOCRFeatures(features []float32, inputs *ScoreInputs) {
	// Extract confidence and validation for each expected OCR field
	for i, fieldName := range fe.config.OCRFieldNames {
		baseIdx := OCROffset + (i * 2)

		// Confidence score (default 0 if not present)
		if confidence, ok := inputs.OCRConfidences[fieldName]; ok {
			features[baseIdx] = confidence
		} else {
			features[baseIdx] = 0.0
		}

		// Validation status (1.0 if valid, 0.0 otherwise)
		if validated, ok := inputs.OCRFieldValidation[fieldName]; ok && validated {
			features[baseIdx+1] = 1.0
		} else {
			features[baseIdx+1] = 0.0
		}
	}

}

// extractMetadataFeatures extracts metadata-related features
func (fe *FeatureExtractor) extractMetadataFeatures(features []float32, inputs *ScoreInputs) {
	// Scope count (normalized)
	features[MetadataOffset] = float32(inputs.ScopeCount) / 10.0 // Max 10 scopes
	if features[MetadataOffset] > 1.0 {
		features[MetadataOffset] = 1.0
	}

	// Scope type indicators (one-hot encoding for common types)
	scopeTypeSet := make(map[string]bool)
	for _, st := range inputs.ScopeTypes {
		scopeTypeSet[st] = true
	}

	// Common scope types (8 indicators)
	scopeTypes := []string{
		"id_document", "selfie", "face_video", "biometric",
		"sso_metadata", "email_proof", "sms_proof", "domain_verify",
	}

	for i, st := range scopeTypes {
		if scopeTypeSet[st] {
			features[MetadataOffset+1+i] = 1.0
		} else {
			features[MetadataOffset+1+i] = 0.0
		}
	}

	// Block height normalized (for temporal features)
	// Normalize to a reasonable range
	normalizedHeight := float32(inputs.Metadata.BlockHeight%1000000) / 1000000.0
	features[MetadataOffset+9] = normalizedHeight

	// Remaining metadata features (padding)
	// Indices 10-15 are reserved for future use
}

// normalizeFeatures applies z-score normalization using training statistics
func (fe *FeatureExtractor) normalizeFeatures(features []float32) error {
	if len(fe.config.FeatureMean) != len(features) ||
		len(fe.config.FeatureStd) != len(features) {
		return fmt.Errorf("normalization dimensions must match feature dimension %d", len(features))
	}

	for i := range features {
		if !isFiniteFloat32(fe.config.FeatureMean[i]) || !isFiniteFloat32(fe.config.FeatureStd[i]) {
			return fmt.Errorf("normalization parameters at %d must be finite", i)
		}
		if fe.config.FeatureStd[i] > 1e-8 {
			features[i] = (features[i] - fe.config.FeatureMean[i]) / fe.config.FeatureStd[i]
		}
	}
	return nil
}

// ============================================================================
// Feature Validation
// ============================================================================

// ValidateInputs validates the score inputs before feature extraction
func (fe *FeatureExtractor) ValidateInputs(inputs *ScoreInputs) []string {
	issues := fe.validateInputs(inputs, fe.config.RequireCompleteInputs)
	if inputs != nil && len(inputs.FaceEmbedding) == 0 && !fe.config.RequireCompleteInputs {
		issues = append([]string{"missing face embedding"}, issues...)
	}
	return issues
}

func (fe *FeatureExtractor) validateInputs(inputs *ScoreInputs, requireComplete bool) []string {
	var issues []string
	if inputs == nil {
		return []string{"score inputs are required"}
	}

	// Check face embedding
	if len(inputs.FaceEmbedding) == 0 {
		if requireComplete {
			issues = append(issues, "missing face embedding")
		}
	} else if len(inputs.FaceEmbedding) != fe.config.FaceEmbeddingDim {
		issues = append(issues, fmt.Sprintf(
			"invalid face embedding dimension: expected %d, got %d",
			fe.config.FaceEmbeddingDim,
			len(inputs.FaceEmbedding),
		))
	}
	for index, value := range inputs.FaceEmbedding {
		if !isFiniteFloat32(value) {
			issues = append(issues, fmt.Sprintf("face embedding[%d] must be finite", index))
			break
		}
	}
	if requireComplete && len(inputs.FaceEmbedding) == fe.config.FaceEmbeddingDim {
		var sumSquares float64
		for _, value := range inputs.FaceEmbedding {
			sumSquares += float64(value) * float64(value)
		}
		if sumSquares <= 1e-20 {
			issues = append(issues, "face embedding must have nonzero norm")
		}
	}

	// Check required scalar ranges.
	if !validUnitScalar(inputs.FaceConfidence) {
		issues = append(issues, fmt.Sprintf(
			"face confidence out of range [0,1]: %.4f",
			inputs.FaceConfidence,
		))
	}

	if !validUnitScalar(inputs.DocQualityScore) {
		issues = append(issues, fmt.Sprintf(
			"document quality score out of range [0,1]: %.4f",
			inputs.DocQualityScore,
		))
	}
	documentScalars := []struct {
		name  string
		value float32
	}{
		{"document sharpness", inputs.DocQualityFeatures.Sharpness},
		{"document brightness", inputs.DocQualityFeatures.Brightness},
		{"document contrast", inputs.DocQualityFeatures.Contrast},
		{"document noise level", inputs.DocQualityFeatures.NoiseLevel},
		{"document blur score", inputs.DocQualityFeatures.BlurScore},
	}
	for _, scalar := range documentScalars {
		if !validUnitScalar(scalar.value) {
			issues = append(issues, fmt.Sprintf("%s out of range [0,1]: %.4f", scalar.name, scalar.value))
		}
	}

	// Check OCR confidences
	for field, conf := range inputs.OCRConfidences {
		if !validUnitScalar(conf) {
			issues = append(issues, fmt.Sprintf(
				"OCR confidence for '%s' out of range [0,1]: %.4f",
				field, conf,
			))
		}
	}
	if requireComplete {
		if inputs.OCRConfidences == nil {
			issues = append(issues, "missing OCR confidences")
		}
		if inputs.OCRFieldValidation == nil {
			issues = append(issues, "missing OCR field validation")
		}
		for _, field := range fe.config.OCRFieldNames {
			if _, ok := inputs.OCRConfidences[field]; !ok {
				issues = append(issues, fmt.Sprintf("missing OCR confidence for %q", field))
			}
			if _, ok := inputs.OCRFieldValidation[field]; !ok {
				issues = append(issues, fmt.Sprintf("missing OCR validation for %q", field))
			}
		}
	}

	// Check scope count
	if inputs.ScopeCount < 0 {
		issues = append(issues, "negative scope count")
	} else if requireComplete && inputs.ScopeCount > 10 {
		issues = append(issues, "scope count out of range [0,10]")
	}
	if requireComplete && inputs.ScopeTypes == nil {
		issues = append(issues, "missing scope types")
	}
	if requireComplete && inputs.Metadata.BlockHeight < 0 {
		issues = append(issues, "negative block height")
	}

	return issues
}

func validUnitScalar(value float32) bool {
	return isFiniteFloat32(value) && value >= 0 && value <= 1
}

func isFiniteFloat32(value float32) bool {
	return !float32IsNaN(value) && !float32IsInf(value)
}

func float32IsNaN(value float32) bool {
	return value != value
}

func float32IsInf(value float32) bool {
	return value > math.MaxFloat32 || value < -math.MaxFloat32
}

// ============================================================================
// Feature Contribution Analysis
// ============================================================================

// ComputeFeatureContributions estimates feature importance for a prediction
// This is a simplified approximation - not the actual model attention
func (fe *FeatureExtractor) ComputeFeatureContributions(features []float32) map[string]float32 {
	contributions := make(map[string]float32)

	// Face embedding contribution (mean of absolute values)
	var faceSum float32
	for i := 0; i < FaceEmbeddingDim; i++ {
		faceSum += absFloat32(features[i])
	}
	contributions["face_embedding"] = faceSum / float32(FaceEmbeddingDim)

	// Document quality contribution
	docOffset := DocQualityOffset
	var docSum float32
	for i := 0; i < DocQualityDim; i++ {
		docSum += features[docOffset+i]
	}
	contributions["doc_quality"] = docSum / float32(DocQualityDim)

	// OCR contribution
	ocrOffset := OCROffset
	var ocrSum float32
	for i := 0; i < OCRFieldsDim; i++ {
		ocrSum += features[ocrOffset+i]
	}
	contributions["ocr"] = ocrSum / float32(OCRFieldsDim)

	// Metadata contribution
	metaOffset := MetadataOffset
	var metaSum float32
	for i := 0; i < MetadataDim; i++ {
		metaSum += features[metaOffset+i]
	}
	contributions["metadata"] = metaSum / float32(MetadataDim)

	return contributions
}

// absFloat32 returns the absolute value of a float32
func absFloat32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
