// Package inference provides TensorFlow-based ML inference for VEID identity scoring.
//
// This package implements the MLScorer interface from x/veid/keeper/scoring.go,
// providing TensorFlow-Go based model inference with deterministic controls
// for blockchain consensus.
//
// VE-205: TensorFlow-Go inference integration in Cosmos module
package inference

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"time"
)

// ============================================================================
// Inference Metadata
// ============================================================================

// InferenceMetadata contains contextual metadata about the inference request
type InferenceMetadata struct {
	// AccountAddress is the blockchain address being verified
	AccountAddress string

	// BlockHeight is the current block height
	BlockHeight int64

	// BlockTime is the current block timestamp
	BlockTime time.Time

	// RequestID is a unique identifier for this inference request
	RequestID string

	// ValidatorAddress is the address of the validator performing inference
	ValidatorAddress string
}

// ============================================================================
// Score Inputs
// ============================================================================

// ScoreInputs contains all inputs required for ML scoring
type ScoreInputs struct {
	// FaceEmbedding is the 512-dimensional face embedding vector
	// from facial verification pipeline
	FaceEmbedding []float32

	// FaceConfidence is the confidence score from face detection (0.0-1.0)
	FaceConfidence float32

	// DocQualityScore is the document quality score (0.0-1.0)
	DocQualityScore float32

	// DocQualityFeatures contains individual document quality metrics
	DocQualityFeatures DocQualityFeatures

	// OCRConfidences maps OCR field names to their confidence scores
	OCRConfidences map[string]float32

	// OCRFieldValidation maps OCR field names to validation status
	OCRFieldValidation map[string]bool

	// Metadata contains contextual information about the request
	Metadata InferenceMetadata

	// ScopeTypes lists the types of scopes being verified
	ScopeTypes []string

	// ScopeCount is the number of identity scopes provided
	ScopeCount int
}

// DocQualityFeatures contains individual document quality metrics
type DocQualityFeatures struct {
	// Sharpness is the image sharpness score (0.0-1.0)
	Sharpness float32

	// Brightness is the image brightness score (0.0-1.0)
	Brightness float32

	// Contrast is the image contrast score (0.0-1.0)
	Contrast float32

	// NoiseLevel is the noise level score (0.0-1.0, lower is better)
	NoiseLevel float32

	// BlurScore is the blur detection score (0.0-1.0, lower is better)
	BlurScore float32
}

// ============================================================================
// Score Result
// ============================================================================

// ScoreResult contains the output from ML inference
type ScoreResult struct {
	// Score is the computed identity score (0-100)
	Score uint32

	// Confidence is the model's confidence in the prediction (0.0-1.0)
	Confidence float32

	// ModelVersion is the version of the model used
	ModelVersion string

	// ModelHash is the SHA256 hash of the model weights
	ModelHash string

	// ReasonCodes provide explanations for the score
	ReasonCodes []string

	// ComputeTimeMs is the inference computation time in milliseconds
	ComputeTimeMs int64

	// InputHash is the SHA256 hash of the inputs for consensus verification
	InputHash string

	// OutputHash is the SHA256 hash of the raw output for determinism verification
	OutputHash string

	// RawScore is the raw model output before quantization
	RawScore float32

	// FeatureContributions shows how each feature type contributed
	FeatureContributions map[string]float32
}

// ============================================================================
// Reason Codes
// ============================================================================

// Reason code constants for score explanations
const (
	// ReasonCodeSuccess indicates successful verification
	ReasonCodeSuccess = "SUCCESS"

	// ReasonCodeHighConfidence indicates high model confidence
	ReasonCodeHighConfidence = "HIGH_CONFIDENCE"

	// ReasonCodeLowConfidence indicates low model confidence
	ReasonCodeLowConfidence = "LOW_CONFIDENCE"

	// ReasonCodeFaceMismatch indicates face verification failed
	ReasonCodeFaceMismatch = "FACE_MISMATCH"

	// ReasonCodeLowDocQuality indicates poor document quality
	ReasonCodeLowDocQuality = "LOW_DOC_QUALITY"

	// ReasonCodeLowOCRConfidence indicates poor OCR confidence
	ReasonCodeLowOCRConfidence = "LOW_OCR_CONFIDENCE"

	// ReasonCodeInsufficientScopes indicates not enough identity scopes
	ReasonCodeInsufficientScopes = "INSUFFICIENT_SCOPES"

	// ReasonCodeMissingFace indicates no face embedding provided
	ReasonCodeMissingFace = "MISSING_FACE"

	// ReasonCodeMissingDocument indicates no document provided
	ReasonCodeMissingDocument = "MISSING_DOCUMENT"

	// ReasonCodeModelLoadError indicates model loading failed
	ReasonCodeModelLoadError = "MODEL_LOAD_ERROR"

	// ReasonCodeInferenceError indicates inference execution failed
	ReasonCodeInferenceError = "INFERENCE_ERROR"

	// ReasonCodeTimeout indicates inference exceeded time limit
	ReasonCodeTimeout = "TIMEOUT"

	// ReasonCodeMemoryLimit indicates memory limit exceeded
	ReasonCodeMemoryLimit = "MEMORY_LIMIT"
)

// ============================================================================
// Feature Dimensions
// ============================================================================

// Feature dimension constants matching the ML training configuration
const (
	// FaceEmbeddingDim is the dimension of face embedding vectors
	FaceEmbeddingDim = 512

	// DocQualityDim is the dimension of document quality features
	DocQualityDim = 5

	// OCRFieldsDim is the dimension of OCR features
	OCRFieldsDim = 10 // 5 fields * 2 (confidence + validation)

	// MetadataDim is the dimension of metadata features
	MetadataDim = 16

	// TotalFeatureDim is the total combined feature vector dimension
	// Must match ml/training/config.py:FeatureConfig.combined_feature_dim
	TotalFeatureDim = 768

	// PaddingDim fills the gap between component features and total dimension
	// 768 - 512 - 5 - 10 - 16 = 225
	PaddingDim = TotalFeatureDim - FaceEmbeddingDim - DocQualityDim - OCRFieldsDim - MetadataDim
)

// OCR field names matching the training configuration
var OCRFieldNames = []string{
	"name",
	"date_of_birth",
	"document_number",
	"expiry_date",
	"nationality",
}

// ============================================================================
// Scorer Interface
// ============================================================================

// Scorer defines the interface for ML-based identity scoring
// This matches the MLScorer interface in x/veid/keeper/scoring.go
type Scorer interface {
	// ComputeScore runs ML inference on the provided inputs
	ComputeScore(inputs *ScoreInputs) (*ScoreResult, error)

	// GetModelVersion returns the current model version
	GetModelVersion() string

	// GetModelHash returns the SHA256 hash of the model weights
	GetModelHash() string

	// IsHealthy checks if the scorer is ready for inference
	IsHealthy() bool

	// Close releases resources held by the scorer
	Close() error
}

// ============================================================================
// Model Metadata
// ============================================================================

// ModelMetadata contains information about a loaded model
type ModelMetadata struct {
	// Version is the model version string
	Version string

	// Hash is the SHA256 hash of the model weights
	Hash string

	// InputShape is the expected input tensor shape
	InputShape []int64

	// OutputShape is the expected output tensor shape
	OutputShape []int64

	// InputName is the name of the input tensor
	InputName string

	// OutputName is the name of the output tensor
	OutputName string

	// ExportTimestamp is when the model was exported
	ExportTimestamp string

	// TensorFlowVersion is the TF version used for export
	TensorFlowVersion string

	// OpNames lists TensorFlow operation names used by the model (if provided)
	OpNames []string
}
