package inference

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Feature Pipeline
// ============================================================================

// FeaturePipeline orchestrates all feature extraction components to build
// the complete feature vector for ML inference.
//
// The pipeline:
// 1. Extracts face embeddings from selfie/video scopes
// 2. Extracts OCR and document quality features from ID document scopes
// 3. Integrates liveness detection results
// 4. Builds the final 768-dimensional feature vector
// 5. Validates and sanitizes all features
type FeaturePipeline struct {
	config PipelineConfig

	faceExtractor   *FaceExtractor
	ocrExtractor    *OCRExtractor
	livenessScorer  *LivenessScorer
	featureExtractor *FeatureExtractor

	mu               sync.RWMutex
	pipelineRunCount uint64
	errorCount       uint64
}

// PipelineConfig contains configuration for the feature pipeline
type PipelineConfig struct {
	// Face extraction configuration
	FaceConfig FaceExtractorConfig

	// OCR extraction configuration
	OCRConfig OCRExtractorConfig

	// Liveness scoring configuration
	LivenessConfig LivenessScorerConfig

	// Feature extraction configuration
	FeatureConfig FeatureExtractorConfig

	// Pipeline behavior
	// Timeout is the overall pipeline timeout
	Timeout time.Duration

	// ParallelExtraction enables parallel extraction of face/OCR/liveness
	ParallelExtraction bool

	// ContinueOnPartialFailure continues even if some extractions fail
	ContinueOnPartialFailure bool

	// ValidateOutputs validates all feature outputs
	ValidateOutputs bool

	// SanitizeOutputs sanitizes all feature outputs
	SanitizeOutputs bool
}

// DefaultPipelineConfig returns sensible defaults
func DefaultPipelineConfig() PipelineConfig {
	return PipelineConfig{
		FaceConfig:               DefaultFaceExtractorConfig(),
		OCRConfig:                DefaultOCRExtractorConfig(),
		LivenessConfig:           DefaultLivenessScorerConfig(),
		FeatureConfig:            DefaultFeatureExtractorConfig(),
		Timeout:                  30 * time.Second,
		ParallelExtraction:       true,
		ContinueOnPartialFailure: true,
		ValidateOutputs:          true,
		SanitizeOutputs:          true,
	}
}

// ScopeData represents input data from an identity scope
type ScopeData struct {
	// ScopeID is the unique identifier for this scope
	ScopeID string

	// ScopeType indicates what kind of data this scope contains
	ScopeType string

	// Data is the raw scope data (image bytes, video bytes, etc.)
	Data []byte

	// Metadata contains additional scope metadata
	Metadata map[string]string
}

// PipelineInput contains all inputs for feature extraction
type PipelineInput struct {
	// Scopes contains all identity scopes to process
	Scopes []ScopeData

	// AccountAddress is the blockchain address being verified
	AccountAddress string

	// BlockHeight is the current block height
	BlockHeight int64

	// BlockTime is the current block timestamp
	BlockTime time.Time

	// RequestID is a unique identifier for this request
	RequestID string

	// ValidatorAddress is the address of the validator performing inference
	ValidatorAddress string
}

// PipelineOutput contains the complete output from the feature pipeline
type PipelineOutput struct {
	// Features is the 768-dimensional feature vector
	Features []float32

	// ScoreInputs is the structured input for the scorer
	ScoreInputs *ScoreInputs

	// FaceResults contains per-scope face extraction results
	FaceResults map[string]*FaceExtractionResult

	// OCRResults contains per-scope OCR extraction results
	OCRResults map[string]*OCRExtractionResult

	// LivenessResult contains the liveness detection result
	LivenessResult *LivenessResult

	// InputHash is the SHA256 hash of all inputs for consensus verification
	InputHash string

	// FeatureHash is the SHA256 hash of the feature vector
	FeatureHash string

	// ProcessingTimeMs is the total pipeline processing time
	ProcessingTimeMs int64

	// ReasonCodes aggregates reason codes from all extractors
	ReasonCodes []string

	// Warnings contains non-fatal issues encountered
	Warnings []string

	// Success indicates if the pipeline completed successfully
	Success bool
}

// ============================================================================
// Pipeline Reason Codes
// ============================================================================

const (
	PipelineReasonCodeSuccess          = "PIPELINE_SUCCESS"
	PipelineReasonCodePartialSuccess   = "PIPELINE_PARTIAL_SUCCESS"
	PipelineReasonCodeNoFaceScopes     = "NO_FACE_SCOPES"
	PipelineReasonCodeNoIDScopes       = "NO_ID_DOCUMENT_SCOPES"
	PipelineReasonCodeNoVideoScopes    = "NO_VIDEO_SCOPES"
	PipelineReasonCodeExtractionError  = "PIPELINE_EXTRACTION_ERROR"
	PipelineReasonCodeValidationFailed = "PIPELINE_VALIDATION_FAILED"
	PipelineReasonCodeTimeout          = "PIPELINE_TIMEOUT"
)

// ============================================================================
// Constructor and Initialization
// ============================================================================

// NewFeaturePipeline creates a new feature pipeline
func NewFeaturePipeline(config PipelineConfig) *FeaturePipeline {
	return &FeaturePipeline{
		config:           config,
		faceExtractor:    NewFaceExtractor(config.FaceConfig),
		ocrExtractor:     NewOCRExtractor(config.OCRConfig),
		livenessScorer:   NewLivenessScorer(config.LivenessConfig),
		featureExtractor: NewFeatureExtractor(config.FeatureConfig),
	}
}

// NewFeaturePipelineWithClients creates a pipeline with custom clients
func NewFeaturePipelineWithClients(
	config PipelineConfig,
	faceClient FaceExtractionClient,
	ocrClient OCRExtractionClient,
	livenessClient LivenessClient,
) *FeaturePipeline {
	return &FeaturePipeline{
		config:           config,
		faceExtractor:    NewFaceExtractorWithClient(config.FaceConfig, faceClient),
		ocrExtractor:     NewOCRExtractorWithClient(config.OCRConfig, ocrClient),
		livenessScorer:   NewLivenessScorerWithClient(config.LivenessConfig, livenessClient),
		featureExtractor: NewFeatureExtractor(config.FeatureConfig),
	}
}

// ============================================================================
// Pipeline Execution
// ============================================================================

// Extract runs the complete feature extraction pipeline
func (p *FeaturePipeline) Extract(ctx context.Context, input *PipelineInput) (*PipelineOutput, error) {
	startTime := time.Now()

	p.mu.Lock()
	p.pipelineRunCount++
	p.mu.Unlock()

	// Initialize output
	output := &PipelineOutput{
		Features:     make([]float32, TotalFeatureDim),
		FaceResults:  make(map[string]*FaceExtractionResult),
		OCRResults:   make(map[string]*OCRExtractionResult),
		ReasonCodes:  []string{},
		Warnings:     []string{},
		Success:      true,
	}

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, p.config.Timeout)
	defer cancel()

	// Categorize scopes by type
	selfieScopes, idDocScopes, videoScopes := p.categorizeScopes(input.Scopes)

	// Add warnings for missing scope types
	if len(selfieScopes) == 0 {
		output.Warnings = append(output.Warnings, "no selfie scopes provided")
		output.ReasonCodes = append(output.ReasonCodes, PipelineReasonCodeNoFaceScopes)
	}
	if len(idDocScopes) == 0 {
		output.Warnings = append(output.Warnings, "no ID document scopes provided")
		output.ReasonCodes = append(output.ReasonCodes, PipelineReasonCodeNoIDScopes)
	}
	if len(videoScopes) == 0 {
		output.Warnings = append(output.Warnings, "no video scopes provided for liveness")
		output.ReasonCodes = append(output.ReasonCodes, PipelineReasonCodeNoVideoScopes)
	}

	// Run extractions
	var faceEmbedding []float32
	var faceConfidence float32
	var ocrResult *OCRExtractionResult
	var livenessResult *LivenessResult

	if p.config.ParallelExtraction {
		faceEmbedding, faceConfidence, ocrResult, livenessResult = p.extractParallel(
			ctx, selfieScopes, idDocScopes, videoScopes, output,
		)
	} else {
		faceEmbedding, faceConfidence, ocrResult, livenessResult = p.extractSequential(
			ctx, selfieScopes, idDocScopes, videoScopes, output,
		)
	}

	// Build ScoreInputs
	scoreInputs := p.buildScoreInputs(input, faceEmbedding, faceConfidence, ocrResult, livenessResult)
	output.ScoreInputs = scoreInputs
	output.LivenessResult = livenessResult

	// Extract feature vector using existing feature extractor
	features, err := p.featureExtractor.ExtractFeatures(scoreInputs)
	if err != nil {
		output.Warnings = append(output.Warnings, fmt.Sprintf("feature extraction error: %v", err))
		output.ReasonCodes = append(output.ReasonCodes, PipelineReasonCodeExtractionError)
		if !p.config.ContinueOnPartialFailure {
			output.Success = false
			return output, err
		}
		// Use default features on error
		features = make([]float32, TotalFeatureDim)
	}
	output.Features = features

	// Validate outputs
	if p.config.ValidateOutputs {
		issues := p.validateOutput(output)
		if len(issues) > 0 {
			output.Warnings = append(output.Warnings, issues...)
			output.ReasonCodes = append(output.ReasonCodes, PipelineReasonCodeValidationFailed)
		}
	}

	// Sanitize outputs
	if p.config.SanitizeOutputs {
		p.sanitizeOutput(output)
	}

	// Compute hashes
	output.InputHash = p.computeInputHash(input)
	output.FeatureHash = p.computeFeatureHash(features)

	// Finalize
	output.ProcessingTimeMs = time.Since(startTime).Milliseconds()

	// Set final success status
	if output.Success && len(output.Warnings) == 0 {
		output.ReasonCodes = append([]string{PipelineReasonCodeSuccess}, output.ReasonCodes...)
	} else if output.Success {
		output.ReasonCodes = append([]string{PipelineReasonCodePartialSuccess}, output.ReasonCodes...)
	}

	return output, nil
}

// categorizeScopes separates scopes by type
func (p *FeaturePipeline) categorizeScopes(scopes []ScopeData) (selfie, idDoc, video []ScopeData) {
	for _, scope := range scopes {
		switch scope.ScopeType {
		case "selfie":
			selfie = append(selfie, scope)
		case "id_document":
			idDoc = append(idDoc, scope)
		case "face_video":
			video = append(video, scope)
		case "biometric":
			// Biometric could contain face data
			if _, ok := scope.Metadata["biometric_type"]; ok {
				if scope.Metadata["biometric_type"] == "face" {
					selfie = append(selfie, scope)
				}
			}
		}
	}
	return
}

// extractParallel runs all extractions in parallel
func (p *FeaturePipeline) extractParallel(
	ctx context.Context,
	selfieScopes, idDocScopes, videoScopes []ScopeData,
	output *PipelineOutput,
) ([]float32, float32, *OCRExtractionResult, *LivenessResult) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	var faceEmbedding []float32
	var faceConfidence float32
	var ocrResult *OCRExtractionResult
	var livenessResult *LivenessResult

	// Extract face embeddings
	wg.Add(1)
	go func() {
		defer wg.Done()
		emb, conf, results := p.extractFaceFromScopes(ctx, selfieScopes)
		mu.Lock()
		faceEmbedding = emb
		faceConfidence = conf
		for id, result := range results {
			output.FaceResults[id] = result
		}
		mu.Unlock()
	}()

	// Extract OCR features
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, results := p.extractOCRFromScopes(ctx, idDocScopes)
		mu.Lock()
		ocrResult = result
		for id, r := range results {
			output.OCRResults[id] = r
		}
		mu.Unlock()
	}()

	// Extract liveness
	wg.Add(1)
	go func() {
		defer wg.Done()
		result := p.extractLivenessFromScopes(ctx, videoScopes)
		mu.Lock()
		livenessResult = result
		mu.Unlock()
	}()

	wg.Wait()

	return faceEmbedding, faceConfidence, ocrResult, livenessResult
}

// extractSequential runs all extractions sequentially
func (p *FeaturePipeline) extractSequential(
	ctx context.Context,
	selfieScopes, idDocScopes, videoScopes []ScopeData,
	output *PipelineOutput,
) ([]float32, float32, *OCRExtractionResult, *LivenessResult) {
	// Extract face embeddings
	faceEmbedding, faceConfidence, faceResults := p.extractFaceFromScopes(ctx, selfieScopes)
	for id, result := range faceResults {
		output.FaceResults[id] = result
	}

	// Extract OCR features
	ocrResult, ocrResults := p.extractOCRFromScopes(ctx, idDocScopes)
	for id, result := range ocrResults {
		output.OCRResults[id] = result
	}

	// Extract liveness
	livenessResult := p.extractLivenessFromScopes(ctx, videoScopes)

	return faceEmbedding, faceConfidence, ocrResult, livenessResult
}

// extractFaceFromScopes extracts face embeddings from selfie scopes
func (p *FeaturePipeline) extractFaceFromScopes(
	ctx context.Context,
	scopes []ScopeData,
) ([]float32, float32, map[string]*FaceExtractionResult) {
	results := make(map[string]*FaceExtractionResult)

	if len(scopes) == 0 {
		// Return zero embedding if no scopes
		return make([]float32, FaceEmbeddingDim), 0.0, results
	}

	var bestEmbedding []float32
	var bestConfidence float32

	for _, scope := range scopes {
		result, err := p.faceExtractor.ExtractFromSelfie(ctx, scope.Data)
		if err != nil {
			results[scope.ScopeID] = &FaceExtractionResult{
				FaceDetected: false,
				ReasonCodes:  []string{FaceReasonCodeExtractionError},
			}
			continue
		}

		results[scope.ScopeID] = result

		// Keep the best result
		if result.FaceDetected && result.Confidence > bestConfidence {
			bestEmbedding = result.Embedding
			bestConfidence = result.Confidence
		}
	}

	if bestEmbedding == nil {
		bestEmbedding = make([]float32, FaceEmbeddingDim)
	}

	return bestEmbedding, bestConfidence, results
}

// extractOCRFromScopes extracts OCR features from ID document scopes
func (p *FeaturePipeline) extractOCRFromScopes(
	ctx context.Context,
	scopes []ScopeData,
) (*OCRExtractionResult, map[string]*OCRExtractionResult) {
	results := make(map[string]*OCRExtractionResult)

	if len(scopes) == 0 {
		// Return empty result if no scopes
		return &OCRExtractionResult{
			Fields:            make(map[string]*ExtractedField),
			DocumentQuality:   DocumentQualityResult{},
			OverallConfidence: 0.0,
			Success:           false,
		}, results
	}

	var bestResult *OCRExtractionResult
	var bestConfidence float32

	for _, scope := range scopes {
		docType := scope.Metadata["document_type"]
		if docType == "" {
			docType = "unknown"
		}

		result, err := p.ocrExtractor.Extract(ctx, scope.Data, docType)
		if err != nil {
			results[scope.ScopeID] = &OCRExtractionResult{
				Success:     false,
				ReasonCodes: []string{OCRReasonCodeExtractionError},
			}
			continue
		}

		results[scope.ScopeID] = result

		// Keep the best result
		if result.Success && result.OverallConfidence > bestConfidence {
			bestResult = result
			bestConfidence = result.OverallConfidence
		}
	}

	if bestResult == nil {
		bestResult = &OCRExtractionResult{
			Fields:            make(map[string]*ExtractedField),
			DocumentQuality:   DocumentQualityResult{},
			OverallConfidence: 0.0,
			Success:           false,
		}
	}

	return bestResult, results
}

// extractLivenessFromScopes extracts liveness from video scopes
func (p *FeaturePipeline) extractLivenessFromScopes(
	ctx context.Context,
	scopes []ScopeData,
) *LivenessResult {
	if len(scopes) == 0 {
		return &LivenessResult{
			IsLive:        false,
			Decision:      "uncertain",
			LivenessScore: 0.0,
			Confidence:    0.0,
			ReasonCodes:   []string{LivenessReasonCodeInsufficientFrames},
		}
	}

	// Use the first video scope for liveness
	// In practice, could combine results from multiple videos
	videoData := scopes[0].Data

	result, err := p.livenessScorer.CheckLiveness(ctx, videoData)
	if err != nil {
		return &LivenessResult{
			IsLive:        false,
			Decision:      "uncertain",
			LivenessScore: 0.0,
			Confidence:    0.0,
			ReasonCodes:   []string{LivenessReasonCodeExtractionError},
		}
	}

	return result
}

// buildScoreInputs builds ScoreInputs from extraction results
//
//nolint:unparam // livenessResult kept for future liveness-based score weighting
func (p *FeaturePipeline) buildScoreInputs(
	input *PipelineInput,
	faceEmbedding []float32,
	faceConfidence float32,
	ocrResult *OCRExtractionResult,
	_ *LivenessResult,
) *ScoreInputs {
	// Get OCR features
	var docQualityScore float32
	var docQualityFeatures DocQualityFeatures
	var ocrConfidences map[string]float32
	var ocrValidation map[string]bool

	if ocrResult != nil {
		docQualityScore, docQualityFeatures, ocrConfidences, ocrValidation = p.ocrExtractor.ToFeatureInputs(ocrResult)
	} else {
		ocrConfidences = make(map[string]float32)
		ocrValidation = make(map[string]bool)
	}

	// Build scope types list
	scopeTypes := make([]string, 0, len(input.Scopes))
	for _, scope := range input.Scopes {
		scopeTypes = append(scopeTypes, scope.ScopeType)
	}

	return &ScoreInputs{
		FaceEmbedding:      faceEmbedding,
		FaceConfidence:     faceConfidence,
		DocQualityScore:    docQualityScore,
		DocQualityFeatures: docQualityFeatures,
		OCRConfidences:     ocrConfidences,
		OCRFieldValidation: ocrValidation,
		Metadata: InferenceMetadata{
			AccountAddress:   input.AccountAddress,
			BlockHeight:      input.BlockHeight,
			BlockTime:        input.BlockTime,
			RequestID:        input.RequestID,
			ValidatorAddress: input.ValidatorAddress,
		},
		ScopeTypes: scopeTypes,
		ScopeCount: len(input.Scopes),
	}
}

// validateOutput validates the pipeline output
func (p *FeaturePipeline) validateOutput(output *PipelineOutput) []string {
	var issues []string

	// Validate feature vector dimension
	if len(output.Features) != TotalFeatureDim {
		issues = append(issues, fmt.Sprintf(
			"feature vector dimension mismatch: expected %d, got %d",
			TotalFeatureDim, len(output.Features),
		))
	}

	// Validate face embedding in ScoreInputs
	if output.ScoreInputs != nil {
		faceIssues := p.faceExtractor.ValidateEmbedding(output.ScoreInputs.FaceEmbedding)
		issues = append(issues, faceIssues...)

		// Validate OCR confidences
		for field, conf := range output.ScoreInputs.OCRConfidences {
			if conf < 0 || conf > 1 {
				issues = append(issues, fmt.Sprintf(
					"OCR confidence for '%s' out of range [0,1]: %.4f",
					field, conf,
				))
			}
		}
	}

	// Validate liveness result
	if output.LivenessResult != nil {
		livenessIssues := p.livenessScorer.ValidateResult(output.LivenessResult)
		issues = append(issues, livenessIssues...)
	}

	return issues
}

// sanitizeOutput sanitizes all values in the output
func (p *FeaturePipeline) sanitizeOutput(output *PipelineOutput) {
	// Sanitize feature vector
	for i := range output.Features {
		output.Features[i] = sanitizeFloat32(output.Features[i], -10.0, 10.0)
	}

	// Sanitize ScoreInputs
	if output.ScoreInputs != nil {
		output.ScoreInputs.FaceEmbedding = p.faceExtractor.SanitizeEmbedding(
			output.ScoreInputs.FaceEmbedding,
		)
		output.ScoreInputs.FaceConfidence = sanitizeFloat32(
			output.ScoreInputs.FaceConfidence, 0.0, 1.0,
		)
		output.ScoreInputs.DocQualityScore = sanitizeFloat32(
			output.ScoreInputs.DocQualityScore, 0.0, 1.0,
		)

		// Sanitize OCR confidences
		for field := range output.ScoreInputs.OCRConfidences {
			output.ScoreInputs.OCRConfidences[field] = sanitizeFloat32(
				output.ScoreInputs.OCRConfidences[field], 0.0, 1.0,
			)
		}
	}

	// Sanitize liveness result
	if output.LivenessResult != nil {
		p.livenessScorer.SanitizeResult(output.LivenessResult)
	}
}

// computeInputHash computes a hash of the pipeline inputs
func (p *FeaturePipeline) computeInputHash(input *PipelineInput) string {
	h := sha256.New()

	// Hash account and metadata
	h.Write([]byte(input.AccountAddress))
	h.Write([]byte(fmt.Sprintf("|%d|%s|", input.BlockHeight, input.RequestID)))

	// Hash scope data
	for _, scope := range input.Scopes {
		h.Write([]byte(scope.ScopeID))
		h.Write([]byte(scope.ScopeType))
		h.Write(scope.Data)
	}

	return hex.EncodeToString(h.Sum(nil))
}

// computeFeatureHash computes a hash of the feature vector
func (p *FeaturePipeline) computeFeatureHash(features []float32) string {
	h := sha256.New()
	for _, f := range features {
		// Round to 6 decimal places for determinism
		h.Write([]byte(fmt.Sprintf("%.6f|", f)))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// ============================================================================
// Convenience Methods
// ============================================================================

// ExtractFromScoreInputs builds a feature vector from pre-built ScoreInputs
func (p *FeaturePipeline) ExtractFromScoreInputs(inputs *ScoreInputs) ([]float32, error) {
	return p.featureExtractor.ExtractFeatures(inputs)
}

// ValidateScoreInputs validates ScoreInputs
func (p *FeaturePipeline) ValidateScoreInputs(inputs *ScoreInputs) []string {
	return p.featureExtractor.ValidateInputs(inputs)
}

// ============================================================================
// Statistics and Health
// ============================================================================

// IsHealthy returns whether all pipeline components are ready
func (p *FeaturePipeline) IsHealthy() bool {
	return p.faceExtractor.IsHealthy() &&
		p.ocrExtractor.IsHealthy() &&
		p.livenessScorer.IsHealthy()
}

// GetStats returns pipeline statistics
func (p *FeaturePipeline) GetStats() PipelineStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return PipelineStats{
		PipelineRunCount: p.pipelineRunCount,
		ErrorCount:       p.errorCount,
		IsHealthy:        p.IsHealthy(),
		FaceStats:        p.faceExtractor.GetStats(),
		OCRStats:         p.ocrExtractor.GetStats(),
		LivenessStats:    p.livenessScorer.GetStats(),
	}
}

// PipelineStats contains pipeline statistics
type PipelineStats struct {
	PipelineRunCount uint64
	ErrorCount       uint64
	IsHealthy        bool
	FaceStats        FaceExtractorStats
	OCRStats         OCRExtractorStats
	LivenessStats    LivenessScorerStats
}

// Close releases all resources
func (p *FeaturePipeline) Close() error {
	var errs []error

	if err := p.faceExtractor.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := p.ocrExtractor.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := p.livenessScorer.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("pipeline close errors: %v", errs)
	}
	return nil
}

// ============================================================================
// Factory Functions
// ============================================================================

// NewDefaultFeaturePipeline creates a pipeline with default configuration
func NewDefaultFeaturePipeline() *FeaturePipeline {
	return NewFeaturePipeline(DefaultPipelineConfig())
}

// NewProductionFeaturePipeline creates a pipeline configured for production
func NewProductionFeaturePipeline(
	faceSidecar, ocrSidecar, livenessSidecar string,
) *FeaturePipeline {
	config := DefaultPipelineConfig()

	config.FaceConfig.SidecarAddress = faceSidecar
	config.FaceConfig.UseFallbackOnError = false

	config.OCRConfig.SidecarAddress = ocrSidecar
	config.OCRConfig.UseFallbackOnError = false

	config.LivenessConfig.SidecarAddress = livenessSidecar
	config.LivenessConfig.UseFallbackOnError = false
	config.LivenessConfig.StrictMode = true

	config.ContinueOnPartialFailure = false
	config.ValidateOutputs = true
	config.SanitizeOutputs = true

	return NewFeaturePipeline(config)
}
