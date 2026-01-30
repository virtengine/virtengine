package inference

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
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
// - 5 dimensions: document quality features
// - 10 dimensions: OCR features (5 fields * 2)
// - 16 dimensions: metadata features
// - 225 dimensions: padding/reserved
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
}

// DefaultFeatureExtractorConfig returns the default configuration
func DefaultFeatureExtractorConfig() FeatureExtractorConfig {
	return FeatureExtractorConfig{
		FaceEmbeddingDim:  FaceEmbeddingDim,
		NormalizeFeatures: false, // Normalization should be done during training
		FeatureMean:       nil,
		FeatureStd:        nil,
		OCRFieldNames:     OCRFieldNames,
	}
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
	features := make([]float32, TotalFeatureDim)
	offset := 0

	// 1. Face embedding (512 dimensions)
	offset, err := fe.extractFaceFeatures(features, offset, inputs)
	if err != nil {
		return nil, fmt.Errorf("face feature extraction failed: %w", err)
	}

	// 2. Document quality features (5 dimensions)
	offset = fe.extractDocQualityFeatures(features, offset, inputs)

	// 3. OCR features (10 dimensions)
	offset = fe.extractOCRFeatures(features, offset, inputs)

	// 4. Metadata features (16 dimensions)
	_ = fe.extractMetadataFeatures(features, offset, inputs)

	// 5. Padding (remaining dimensions)
	// Already initialized to 0.0

	// 6. Apply normalization if configured
	if fe.config.NormalizeFeatures && fe.config.FeatureMean != nil && fe.config.FeatureStd != nil {
		fe.normalizeFeatures(features)
	}

	return features, nil
}

// extractFaceFeatures extracts face embedding and confidence
func (fe *FeatureExtractor) extractFaceFeatures(features []float32, offset int, inputs *ScoreInputs) (int, error) {
	// Validate face embedding dimension
	if len(inputs.FaceEmbedding) == 0 {
		// No face embedding - fill with zeros
		for i := 0; i < fe.config.FaceEmbeddingDim; i++ {
			features[offset+i] = 0.0
		}
		return offset + fe.config.FaceEmbeddingDim, nil
	}

	if len(inputs.FaceEmbedding) != fe.config.FaceEmbeddingDim {
		return offset, fmt.Errorf(
			"face embedding dimension mismatch: expected %d, got %d",
			fe.config.FaceEmbeddingDim,
			len(inputs.FaceEmbedding),
		)
	}

	// Copy face embedding
	copy(features[offset:offset+fe.config.FaceEmbeddingDim], inputs.FaceEmbedding)

	// Normalize face embedding to unit length for consistency
	fe.normalizeEmbedding(features[offset : offset+fe.config.FaceEmbeddingDim])

	return offset + fe.config.FaceEmbeddingDim, nil
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
func (fe *FeatureExtractor) extractDocQualityFeatures(features []float32, offset int, inputs *ScoreInputs) int {
	// Overall document quality score
	features[offset] = inputs.DocQualityScore

	// Individual quality features
	features[offset+1] = inputs.DocQualityFeatures.Sharpness
	features[offset+2] = inputs.DocQualityFeatures.Brightness
	features[offset+3] = inputs.DocQualityFeatures.Contrast
	// Invert noise and blur so higher is better
	features[offset+4] = 1.0 - inputs.DocQualityFeatures.NoiseLevel

	return offset + DocQualityDim
}

// extractOCRFeatures extracts OCR-related features
func (fe *FeatureExtractor) extractOCRFeatures(features []float32, offset int, inputs *ScoreInputs) int {
	// Extract confidence and validation for each expected OCR field
	for i, fieldName := range fe.config.OCRFieldNames {
		baseIdx := offset + (i * 2)

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

	return offset + OCRFieldsDim
}

// extractMetadataFeatures extracts metadata-related features
func (fe *FeatureExtractor) extractMetadataFeatures(features []float32, offset int, inputs *ScoreInputs) int {
	// Scope count (normalized)
	features[offset] = float32(inputs.ScopeCount) / 10.0 // Max 10 scopes
	if features[offset] > 1.0 {
		features[offset] = 1.0
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
			features[offset+1+i] = 1.0
		} else {
			features[offset+1+i] = 0.0
		}
	}

	// Face confidence as a metadata feature
	features[offset+9] = inputs.FaceConfidence

	// Block height normalized (for temporal features)
	// Normalize to a reasonable range
	normalizedHeight := float32(inputs.Metadata.BlockHeight%1000000) / 1000000.0
	features[offset+10] = normalizedHeight

	// Remaining metadata features (padding)
	// Indices 11-15 are reserved for future use

	return offset + MetadataDim
}

// normalizeFeatures applies z-score normalization using training statistics
func (fe *FeatureExtractor) normalizeFeatures(features []float32) {
	if len(fe.config.FeatureMean) != len(features) ||
		len(fe.config.FeatureStd) != len(features) {
		return // Cannot normalize with mismatched dimensions
	}

	for i := range features {
		if fe.config.FeatureStd[i] > 1e-8 {
			features[i] = (features[i] - fe.config.FeatureMean[i]) / fe.config.FeatureStd[i]
		}
	}
}

// ============================================================================
// Feature Validation
// ============================================================================

// ValidateInputs validates the score inputs before feature extraction
func (fe *FeatureExtractor) ValidateInputs(inputs *ScoreInputs) []string {
	var issues []string

	// Check face embedding
	if len(inputs.FaceEmbedding) == 0 {
		issues = append(issues, "missing face embedding")
	} else if len(inputs.FaceEmbedding) != fe.config.FaceEmbeddingDim {
		issues = append(issues, fmt.Sprintf(
			"invalid face embedding dimension: expected %d, got %d",
			fe.config.FaceEmbeddingDim,
			len(inputs.FaceEmbedding),
		))
	}

	// Check face confidence range
	if inputs.FaceConfidence < 0 || inputs.FaceConfidence > 1 {
		issues = append(issues, fmt.Sprintf(
			"face confidence out of range [0,1]: %.4f",
			inputs.FaceConfidence,
		))
	}

	// Check document quality score range
	if inputs.DocQualityScore < 0 || inputs.DocQualityScore > 1 {
		issues = append(issues, fmt.Sprintf(
			"document quality score out of range [0,1]: %.4f",
			inputs.DocQualityScore,
		))
	}

	// Check OCR confidences
	for field, conf := range inputs.OCRConfidences {
		if conf < 0 || conf > 1 {
			issues = append(issues, fmt.Sprintf(
				"OCR confidence for '%s' out of range [0,1]: %.4f",
				field, conf,
			))
		}
	}

	// Check scope count
	if inputs.ScopeCount < 0 {
		issues = append(issues, "negative scope count")
	}

	return issues
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
	docOffset := FaceEmbeddingDim
	var docSum float32
	for i := 0; i < DocQualityDim; i++ {
		docSum += features[docOffset+i]
	}
	contributions["doc_quality"] = docSum / float32(DocQualityDim)

	// OCR contribution
	ocrOffset := FaceEmbeddingDim + DocQualityDim
	var ocrSum float32
	for i := 0; i < OCRFieldsDim; i++ {
		ocrSum += features[ocrOffset+i]
	}
	contributions["ocr"] = ocrSum / float32(OCRFieldsDim)

	// Metadata contribution
	metaOffset := FaceEmbeddingDim + DocQualityDim + OCRFieldsDim
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
