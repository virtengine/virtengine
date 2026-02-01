package inference

import (
	"context"
	"fmt"
	"sync"
	"time"

	verrors "github.com/virtengine/virtengine/pkg/errors"
	"github.com/virtengine/virtengine/pkg/security"
)

// ============================================================================
// TensorFlow Scorer
// ============================================================================

// TensorFlowScorer implements the Scorer interface using TensorFlow-Go
// for ML inference in the VEID module.
//
// This scorer:
// - Loads TensorFlow SavedModel exported from training pipeline
// - Extracts features from score inputs
// - Runs inference with timeout and memory bounds
// - Ensures deterministic execution for blockchain consensus
// - Computes hashes for verification
type TensorFlowScorer struct {
	// model is the loaded TensorFlow model
	model *TFModel

	// loader handles model loading
	loader *ModelLoader

	// config holds inference configuration
	config InferenceConfig

	// extractor transforms inputs to features
	extractor *FeatureExtractor

	// determinism ensures deterministic execution
	determinism *DeterminismController

	// mu protects scorer state
	mu sync.RWMutex

	// isHealthy indicates if the scorer is ready
	isHealthy bool

	// lastHealthCheck is the time of the last health check
	lastHealthCheck time.Time

	// inferenceCount tracks total inferences
	inferenceCount uint64

	// errorCount tracks inference errors
	errorCount uint64
}

// NewTensorFlowScorer creates a new TensorFlow-based scorer
func NewTensorFlowScorer(config InferenceConfig) (*TensorFlowScorer, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid inference config: %w", err)
	}

	scorer := &TensorFlowScorer{
		config:          config,
		loader:          NewModelLoader(config),
		extractor:       NewFeatureExtractor(DefaultFeatureExtractorConfig()),
		determinism:     NewDeterminismController(config.RandomSeed, config.ForceCPU),
		isHealthy:       false,
		lastHealthCheck: time.Time{},
	}

	// Load the model
	model, err := scorer.loader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load model: %w", err)
	}

	scorer.model = model
	scorer.isHealthy = true
	scorer.lastHealthCheck = time.Now()

	return scorer, nil
}

// ============================================================================
// Scorer Interface Implementation
// ============================================================================

// ComputeScore runs ML inference on the provided inputs
func (s *TensorFlowScorer) ComputeScore(inputs *ScoreInputs) (*ScoreResult, error) {
	return s.ComputeScoreWithContext(context.Background(), inputs)
}

// ComputeScoreWithContext runs ML inference with context support
func (s *TensorFlowScorer) ComputeScoreWithContext(ctx context.Context, inputs *ScoreInputs) (*ScoreResult, error) {
	startTime := time.Now()

	s.mu.Lock()
	s.inferenceCount++
	s.mu.Unlock()

	// Create result with defaults
	result := &ScoreResult{
		Score:         0,
		Confidence:    0.0,
		ModelVersion:  s.GetModelVersion(),
		ModelHash:     s.GetModelHash(),
		ReasonCodes:   make([]string, 0),
		ComputeTimeMs: 0,
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	// Run inference with timeout
	errChan := make(chan error, 1)
	resultChan := make(chan *ScoreResult, 1)

	verrors.SafeGo("", func() {
		defer func() {}() // WG Done if needed
		res, err := s.runInference(inputs)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- res
	})

	// Wait for result or timeout
	select {
	case err := <-errChan:
		s.mu.Lock()
		s.errorCount++
		s.mu.Unlock()

		result.ComputeTimeMs = time.Since(startTime).Milliseconds()
		result.ReasonCodes = append(result.ReasonCodes, ReasonCodeInferenceError)

		if s.config.UseFallbackOnError {
			result.Score = s.config.FallbackScore
			return result, nil
		}
		return result, err

	case res := <-resultChan:
		res.ComputeTimeMs = time.Since(startTime).Milliseconds()
		return res, nil

	case <-ctx.Done():
		s.mu.Lock()
		s.errorCount++
		s.mu.Unlock()

		result.ComputeTimeMs = time.Since(startTime).Milliseconds()
		result.ReasonCodes = append(result.ReasonCodes, ReasonCodeTimeout)

		if s.config.UseFallbackOnError {
			result.Score = s.config.FallbackScore
			return result, nil
		}
		return result, fmt.Errorf("inference timeout after %v", s.config.Timeout)
	}
}

// runInference performs the actual inference
func (s *TensorFlowScorer) runInference(inputs *ScoreInputs) (*ScoreResult, error) {
	result := &ScoreResult{
		ModelVersion:         s.GetModelVersion(),
		ModelHash:            s.GetModelHash(),
		ReasonCodes:          make([]string, 0),
		FeatureContributions: make(map[string]float32),
	}

	// Compute input hash for consensus verification
	result.InputHash = s.determinism.ComputeInputHash(inputs)

	// Validate inputs
	issues := s.extractor.ValidateInputs(inputs)
	if len(issues) > 0 {
		// Add reason codes for validation issues
		for _, issue := range issues {
			switch {
			case contains(issue, "face embedding"):
				result.ReasonCodes = append(result.ReasonCodes, ReasonCodeMissingFace)
			case contains(issue, "document"):
				result.ReasonCodes = append(result.ReasonCodes, ReasonCodeMissingDocument)
			}
		}
	}

	// Extract features
	features, err := s.extractor.ExtractFeatures(inputs)
	if err != nil {
		result.ReasonCodes = append(result.ReasonCodes, ReasonCodeInferenceError)
		return result, fmt.Errorf("feature extraction failed: %w", err)
	}

	// Run model inference
	output, err := s.model.Run(features)
	if err != nil {
		result.ReasonCodes = append(result.ReasonCodes, ReasonCodeInferenceError)
		return result, fmt.Errorf("model inference failed: %w", err)
	}

	// Process output
	if len(output) == 0 {
		result.ReasonCodes = append(result.ReasonCodes, ReasonCodeInferenceError)
		return result, fmt.Errorf("model returned empty output")
	}

	rawScore := output[0]
	result.RawScore = rawScore

	// Compute output hash for determinism verification
	result.OutputHash = s.determinism.ComputeOutputHash(output)

	// Quantize to 0-100 score using safe conversion
	// This prevents integer overflow issues on 32-bit systems (CWE-190)
	score := security.SafeFloat32ToUint32(rawScore, 0, 100)
	result.Score = score

	// Compute confidence based on raw score distribution
	// Scores closer to 0 or 100 are more confident
	result.Confidence = computeConfidence(rawScore)

	// Add appropriate reason codes
	s.addReasonCodes(result, inputs)

	// Compute feature contributions
	result.FeatureContributions = s.extractor.ComputeFeatureContributions(features)

	return result, nil
}

// addReasonCodes adds appropriate reason codes based on score and inputs
func (s *TensorFlowScorer) addReasonCodes(result *ScoreResult, inputs *ScoreInputs) {
	// Success if score meets basic threshold
	if result.Score >= 50 {
		result.ReasonCodes = append([]string{ReasonCodeSuccess}, result.ReasonCodes...)
	}

	// Confidence-based reason codes
	if result.Confidence >= 0.8 {
		result.ReasonCodes = append(result.ReasonCodes, ReasonCodeHighConfidence)
	} else if result.Confidence < 0.5 {
		result.ReasonCodes = append(result.ReasonCodes, ReasonCodeLowConfidence)
	}

	// Check for low document quality
	if inputs.DocQualityScore < 0.6 {
		result.ReasonCodes = append(result.ReasonCodes, ReasonCodeLowDocQuality)
	}

	// Check for low OCR confidence
	var avgOCRConfidence float32
	for _, conf := range inputs.OCRConfidences {
		avgOCRConfidence += conf
	}
	if len(inputs.OCRConfidences) > 0 {
		avgOCRConfidence /= float32(len(inputs.OCRConfidences))
		if avgOCRConfidence < 0.5 {
			result.ReasonCodes = append(result.ReasonCodes, ReasonCodeLowOCRConfidence)
		}
	}

	// Check for insufficient scopes
	if inputs.ScopeCount < 2 {
		result.ReasonCodes = append(result.ReasonCodes, ReasonCodeInsufficientScopes)
	}
}

// computeConfidence computes confidence based on score distance from boundaries
func computeConfidence(rawScore float32) float32 {
	// Scores near 0 or 100 are more confident
	// Scores near 50 are less confident
	distanceFromMiddle := absFloat32(rawScore-50) / 50.0
	confidence := 0.5 + (distanceFromMiddle * 0.4)
	if confidence > 0.95 {
		confidence = 0.95
	}
	if confidence < 0.3 {
		confidence = 0.3
	}
	return confidence
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetModelVersion returns the current model version
func (s *TensorFlowScorer) GetModelVersion() string {
	if s.model != nil {
		return s.model.GetVersion()
	}
	return s.config.ModelVersion
}

// GetModelHash returns the SHA256 hash of the model weights
func (s *TensorFlowScorer) GetModelHash() string {
	if s.model != nil {
		return s.model.GetModelHash()
	}
	return ""
}

// IsHealthy checks if the scorer is ready for inference
func (s *TensorFlowScorer) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isHealthy {
		return false
	}

	if s.model == nil {
		return false
	}

	return s.model.IsLoaded()
}

// Close releases resources held by the scorer
func (s *TensorFlowScorer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.isHealthy = false

	if s.loader != nil {
		return s.loader.Unload()
	}

	return nil
}

// ============================================================================
// Statistics
// ============================================================================

// GetStats returns scorer statistics
func (s *TensorFlowScorer) GetStats() ScorerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return ScorerStats{
		InferenceCount:  s.inferenceCount,
		ErrorCount:      s.errorCount,
		IsHealthy:       s.isHealthy,
		LastHealthCheck: s.lastHealthCheck,
		ModelVersion:    s.GetModelVersion(),
		ModelHash:       s.GetModelHash(),
	}
}

// ScorerStats contains scorer statistics
type ScorerStats struct {
	InferenceCount  uint64
	ErrorCount      uint64
	IsHealthy       bool
	LastHealthCheck time.Time
	ModelVersion    string
	ModelHash       string
}

// ============================================================================
// Factory Functions
// ============================================================================

// NewScorer creates a new scorer based on configuration
// Returns either a TensorFlow scorer or sidecar client
func NewScorer(config InferenceConfig) (Scorer, error) {
	if config.UseSidecar {
		return NewSidecarClient(config)
	}
	return NewTensorFlowScorer(config)
}

// MustNewScorer creates a new scorer or panics
func MustNewScorer(config InferenceConfig) Scorer {
	scorer, err := NewScorer(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create scorer: %v", err))
	}
	return scorer
}

