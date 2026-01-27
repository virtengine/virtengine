package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"os"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/pkg/inference"
	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// ML Scoring Configuration
// ============================================================================

// MLScoringConfig holds configuration for ML-based identity scoring
type MLScoringConfig struct {
	// ModelVersion is the current ML model version in use
	ModelVersion string

	// MinScopesForScoring is the minimum number of valid scopes required
	MinScopesForScoring int

	// RequiredScopeTypes lists scope types that must be present
	RequiredScopeTypes []types.ScopeType

	// MaxInferenceTime is the maximum time allowed for ML inference in milliseconds
	MaxInferenceTime int64

	// FallbackScore is returned when ML inference fails but basic validation passes
	FallbackScore uint32

	// UseTensorFlow enables TensorFlow-based scoring (VE-205)
	UseTensorFlow bool

	// TensorFlowConfig holds TensorFlow-specific configuration
	TensorFlowConfig *TensorFlowScoringConfig
}

// TensorFlowScoringConfig holds TensorFlow-specific configuration
type TensorFlowScoringConfig struct {
	// ModelPath is the path to the TensorFlow SavedModel
	ModelPath string

	// ExpectedHash is the expected SHA256 hash of the model
	ExpectedHash string

	// UseSidecar enables the gRPC sidecar mode
	UseSidecar bool

	// SidecarAddress is the gRPC address of the inference sidecar
	SidecarAddress string

	// Deterministic forces deterministic inference mode
	Deterministic bool

	// ForceCPU forces CPU-only execution
	ForceCPU bool
}

// DefaultMLScoringConfig returns default ML scoring configuration
func DefaultMLScoringConfig() MLScoringConfig {
	return MLScoringConfig{
		ModelVersion:        "v1.0.0",
		MinScopesForScoring: 1,
		RequiredScopeTypes: []types.ScopeType{
			types.ScopeTypeIDDocument,
			types.ScopeTypeSelfie,
		},
		MaxInferenceTime: 2000, // 2 seconds
		FallbackScore:    0,
		UseTensorFlow:    isTensorFlowEnabled(),
		TensorFlowConfig: DefaultTensorFlowScoringConfig(),
	}
}

// DefaultTensorFlowScoringConfig returns default TensorFlow configuration
func DefaultTensorFlowScoringConfig() *TensorFlowScoringConfig {
	return &TensorFlowScoringConfig{
		ModelPath:      getEnvOrDefault("VEID_INFERENCE_MODEL_PATH", "models/trust_score"),
		ExpectedHash:   os.Getenv("VEID_INFERENCE_MODEL_HASH"),
		UseSidecar:     os.Getenv("VEID_INFERENCE_USE_SIDECAR") == "true",
		SidecarAddress: getEnvOrDefault("VEID_INFERENCE_SIDECAR_ADDR", "localhost:50051"),
		Deterministic:  true,
		ForceCPU:       true,
	}
}

// isTensorFlowEnabled checks if TensorFlow scoring is enabled
func isTensorFlowEnabled() bool {
	return os.Getenv("VEID_USE_TENSORFLOW") == "true"
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ============================================================================
// Scoring Input
// ============================================================================

// ScoringInput represents the input data for ML scoring
type ScoringInput struct {
	// AccountAddress is the account being scored
	AccountAddress string

	// DecryptedScopes are the decrypted scope contents
	DecryptedScopes []DecryptedScope

	// ScopeResults are the current results for each scope
	ScopeResults []types.ScopeVerificationResult

	// PreviousScore is the account's previous score (if any)
	PreviousScore uint32

	// RequestTime is when the scoring was requested
	RequestTime time.Time

	// BlockHeight is the current block height
	BlockHeight int64
}

// NewScoringInput creates a new scoring input
func NewScoringInput(
	accountAddress string,
	decryptedScopes []DecryptedScope,
	scopeResults []types.ScopeVerificationResult,
	requestTime time.Time,
	blockHeight int64,
) *ScoringInput {
	return &ScoringInput{
		AccountAddress:  accountAddress,
		DecryptedScopes: decryptedScopes,
		ScopeResults:    scopeResults,
		PreviousScore:   0,
		RequestTime:     requestTime,
		BlockHeight:     blockHeight,
	}
}

// ComputeInputHash computes a deterministic hash of the scoring inputs
// This is used for consensus verification
func (si *ScoringInput) ComputeInputHash() []byte {
	h := sha256.New()

	// Include account address
	h.Write([]byte(si.AccountAddress))

	// Include block height for determinism
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, uint64(si.BlockHeight))
	h.Write(heightBytes)

	// Include content hashes of all decrypted scopes (in order)
	for _, scope := range si.DecryptedScopes {
		h.Write([]byte(scope.ScopeID))
		h.Write(scope.ContentHash)
	}

	return h.Sum(nil)
}

// ============================================================================
// Scoring Output
// ============================================================================

// ScoringOutput represents the output from ML scoring
type ScoringOutput struct {
	// Score is the computed identity score (0-100)
	Score uint32

	// ModelVersion is the ML model version used
	ModelVersion string

	// ReasonCodes provide explanations for the score
	ReasonCodes []types.ReasonCode

	// ScopeScores are individual scores for each scope
	ScopeScores map[string]uint32

	// Confidence is the model's confidence in the score (0.0-1.0)
	Confidence float64

	// ProcessingTime is how long inference took in milliseconds
	ProcessingTime int64

	// InputHash is the hash of inputs for consensus
	InputHash []byte
}

// ============================================================================
// ML Scoring Interface
// ============================================================================

// MLScorer defines the interface for ML-based identity scoring
// VE-205 will provide a TensorFlow implementation
type MLScorer interface {
	// Score computes an identity score from the given inputs
	Score(input *ScoringInput) (*ScoringOutput, error)

	// GetModelVersion returns the current model version
	GetModelVersion() string

	// IsHealthy checks if the ML service is healthy
	IsHealthy() bool

	// Close releases any resources
	Close() error
}

// ============================================================================
// Stub ML Scorer Implementation
// ============================================================================

// StubMLScorer is a stub implementation for development and testing
// This will be replaced by TensorFlow integration in VE-205
type StubMLScorer struct {
	config MLScoringConfig
}

// NewStubMLScorer creates a new stub ML scorer
func NewStubMLScorer(config MLScoringConfig) *StubMLScorer {
	return &StubMLScorer{
		config: config,
	}
}

// Score implements MLScorer.Score
// Stub implementation returns deterministic scores based on scope types and counts
func (s *StubMLScorer) Score(input *ScoringInput) (*ScoringOutput, error) {
	startTime := time.Now()

	output := &ScoringOutput{
		Score:        0,
		ModelVersion: s.config.ModelVersion,
		ReasonCodes:  make([]types.ReasonCode, 0),
		ScopeScores:  make(map[string]uint32),
		Confidence:   0.0,
		InputHash:    input.ComputeInputHash(),
	}

	// Check minimum scopes requirement
	if len(input.DecryptedScopes) < s.config.MinScopesForScoring {
		output.ReasonCodes = append(output.ReasonCodes, types.ReasonCodeInsufficientScopes)
		output.ProcessingTime = time.Since(startTime).Milliseconds()
		return output, nil
	}

	// Compute individual scope scores
	var totalWeight uint32
	var weightedSum uint32
	hasIDDocument := false
	hasSelfie := false

	for _, scope := range input.DecryptedScopes {
		// Compute deterministic score based on scope type and content hash
		scopeScore := s.computeScopeScore(scope)
		output.ScopeScores[scope.ScopeID] = scopeScore

		weight := types.ScopeTypeWeight(scope.ScopeType)
		weightedSum += scopeScore * weight
		totalWeight += weight

		// Track required scope types
		switch scope.ScopeType {
		case types.ScopeTypeIDDocument:
			hasIDDocument = true
		case types.ScopeTypeSelfie:
			hasSelfie = true
		}
	}

	// Check for required scope types
	if !hasIDDocument || !hasSelfie {
		// Apply penalty for missing required scopes
		output.ReasonCodes = append(output.ReasonCodes, types.ReasonCodeInsufficientScopes)
		// Reduce score by 30% for missing required scopes
		if totalWeight > 0 {
			output.Score = (weightedSum / totalWeight) * 70 / 100
		}
	} else if totalWeight > 0 {
		output.Score = weightedSum / totalWeight
	}

	// Cap at max score
	if output.Score > types.MaxScore {
		output.Score = types.MaxScore
	}

	// Set confidence based on number and types of scopes
	output.Confidence = s.computeConfidence(input.DecryptedScopes)

	// Set success reason if score is reasonable
	if output.Score >= types.ThresholdBasic {
		output.ReasonCodes = append([]types.ReasonCode{types.ReasonCodeSuccess}, output.ReasonCodes...)
	}

	output.ProcessingTime = time.Since(startTime).Milliseconds()
	return output, nil
}

// computeScopeScore computes a deterministic score for a single scope
// This is a stub - real implementation will use TensorFlow
func (s *StubMLScorer) computeScopeScore(scope DecryptedScope) uint32 {
	// Use content hash to generate deterministic "random" score
	// This simulates ML inference returning consistent results
	if len(scope.ContentHash) < 4 {
		return 50 // Default mid-range score
	}

	// Use first 4 bytes of content hash to generate score
	hashValue := binary.BigEndian.Uint32(scope.ContentHash[:4])

	// Map to score range based on scope type
	var baseScore, variance uint32

	switch scope.ScopeType {
	case types.ScopeTypeIDDocument:
		baseScore = 75
		variance = 20
	case types.ScopeTypeSelfie:
		baseScore = 80
		variance = 15
	case types.ScopeTypeFaceVideo:
		baseScore = 85
		variance = 10
	case types.ScopeTypeBiometric:
		baseScore = 80
		variance = 15
	case types.ScopeTypeSSOMetadata:
		baseScore = 60
		variance = 20
	case types.ScopeTypeEmailProof:
		baseScore = 70
		variance = 20
	case types.ScopeTypeSMSProof:
		baseScore = 70
		variance = 20
	case types.ScopeTypeDomainVerify:
		baseScore = 75
		variance = 20
	default:
		baseScore = 50
		variance = 30
	}

	// Generate score in range [baseScore - variance/2, baseScore + variance/2]
	adjustment := hashValue % variance
	score := baseScore - variance/2 + adjustment

	// Ensure score is in valid range
	if score > types.MaxScore {
		score = types.MaxScore
	}

	return score
}

// computeConfidence computes a confidence value based on available scopes
func (s *StubMLScorer) computeConfidence(scopes []DecryptedScope) float64 {
	if len(scopes) == 0 {
		return 0.0
	}

	// Base confidence from scope count
	confidence := 0.3 + (float64(len(scopes)) * 0.1)
	if confidence > 0.9 {
		confidence = 0.9
	}

	// Boost confidence for high-value scope types
	for _, scope := range scopes {
		switch scope.ScopeType {
		case types.ScopeTypeIDDocument:
			confidence += 0.05
		case types.ScopeTypeSelfie:
			confidence += 0.03
		case types.ScopeTypeFaceVideo:
			confidence += 0.05
		}
	}

	// Cap at 0.95 (never 100% confident)
	if confidence > 0.95 {
		confidence = 0.95
	}

	return confidence
}

// GetModelVersion implements MLScorer.GetModelVersion
func (s *StubMLScorer) GetModelVersion() string {
	return s.config.ModelVersion
}

// IsHealthy implements MLScorer.IsHealthy
func (s *StubMLScorer) IsHealthy() bool {
	return true // Stub is always healthy
}

// Close implements MLScorer.Close
func (s *StubMLScorer) Close() error {
	return nil // Nothing to close
}

// ============================================================================
// Keeper Scoring Methods
// ============================================================================

// ComputeIdentityScore computes the identity score for decrypted scopes
// This is the main entry point for ML scoring
func (k Keeper) ComputeIdentityScore(
	ctx sdk.Context,
	accountAddress string,
	decryptedScopes []DecryptedScope,
	scopeResults []types.ScopeVerificationResult,
) (score uint32, modelVersion string, reasonCodes []types.ReasonCode, inputHash []byte, err error) {
	// Create scoring input
	input := NewScoringInput(
		accountAddress,
		decryptedScopes,
		scopeResults,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)

	// Get or create ML scorer
	scorer := k.getMLScorer()
	defer func() {
		if scorer != nil {
			_ = scorer.Close()
		}
	}()

	// Check if scorer is healthy
	if !scorer.IsHealthy() {
		k.Logger(ctx).Error("ML scorer is not healthy")
		return 0, "", []types.ReasonCode{types.ReasonCodeMLInferenceError}, nil, types.ErrMLInferenceFailed.Wrap("scorer not healthy")
	}

	// Perform scoring
	output, err := scorer.Score(input)
	if err != nil {
		k.Logger(ctx).Error("ML scoring failed", "error", err)
		return 0, "", []types.ReasonCode{types.ReasonCodeMLInferenceError}, nil, types.ErrMLInferenceFailed.Wrap(err.Error())
	}

	k.Logger(ctx).Info("identity score computed",
		"account", accountAddress,
		"score", output.Score,
		"model_version", output.ModelVersion,
		"confidence", output.Confidence,
		"processing_time_ms", output.ProcessingTime,
	)

	return output.Score, output.ModelVersion, output.ReasonCodes, output.InputHash, nil
}

// getMLScorer returns the ML scorer instance
// Returns TensorFlow scorer when VEID_USE_TENSORFLOW=true, otherwise stub
func (k Keeper) getMLScorer() MLScorer {
	config := DefaultMLScoringConfig()

	// Use TensorFlow scorer if enabled (VE-205)
	if config.UseTensorFlow && config.TensorFlowConfig != nil {
		scorer, err := k.createTensorFlowScorer(config)
		if err != nil {
			// Fall back to stub on TensorFlow initialization error
			return NewStubMLScorer(config)
		}
		return scorer
	}

	return NewStubMLScorer(config)
}

// createTensorFlowScorer creates a TensorFlow-based scorer
func (k Keeper) createTensorFlowScorer(config MLScoringConfig) (MLScorer, error) {
	tfConfig := config.TensorFlowConfig

	// Build inference configuration
	inferConfig := inference.InferenceConfig{
		ModelPath:          tfConfig.ModelPath,
		ModelVersion:       config.ModelVersion,
		ExpectedHash:       tfConfig.ExpectedHash,
		Timeout:            time.Duration(config.MaxInferenceTime) * time.Millisecond,
		MaxMemoryMB:        512,
		UseSidecar:         tfConfig.UseSidecar,
		SidecarAddress:     tfConfig.SidecarAddress,
		SidecarTimeout:     5 * time.Second,
		Deterministic:      tfConfig.Deterministic,
		ForceCPU:           tfConfig.ForceCPU,
		RandomSeed:         42,
		ExpectedInputDim:   inference.TotalFeatureDim,
		UseFallbackOnError: true,
		FallbackScore:      config.FallbackScore,
	}

	// Validate configuration
	if err := inferConfig.Validate(); err != nil {
		return nil, err
	}

	// Create the appropriate scorer
	scorer, err := inference.NewScorer(inferConfig)
	if err != nil {
		return nil, err
	}

	// Wrap in adapter that implements MLScorer interface
	return &TensorFlowScorerAdapter{
		scorer: scorer,
		config: config,
	}, nil
}

// TensorFlowScorerAdapter adapts the inference.Scorer to MLScorer interface
type TensorFlowScorerAdapter struct {
	scorer inference.Scorer
	config MLScoringConfig
}

// Score implements MLScorer.Score using TensorFlow inference
func (a *TensorFlowScorerAdapter) Score(input *ScoringInput) (*ScoringOutput, error) {
	startTime := time.Now()

	// Convert ScoringInput to inference.ScoreInputs
	inferInputs := a.convertToInferenceInputs(input)

	// Run TensorFlow inference
	result, err := a.scorer.ComputeScore(inferInputs)
	if err != nil {
		return nil, err
	}

	// Convert result to ScoringOutput
	output := &ScoringOutput{
		Score:          result.Score,
		ModelVersion:   result.ModelVersion,
		ReasonCodes:    a.convertReasonCodes(result.ReasonCodes),
		ScopeScores:    make(map[string]uint32),
		Confidence:     float64(result.Confidence),
		ProcessingTime: time.Since(startTime).Milliseconds(),
		InputHash:      []byte(result.InputHash),
	}

	return output, nil
}

// convertToInferenceInputs converts keeper inputs to inference inputs
func (a *TensorFlowScorerAdapter) convertToInferenceInputs(input *ScoringInput) *inference.ScoreInputs {
	// Extract face embedding and other features from decrypted scopes
	var faceEmbedding []float32
	var faceConfidence float32
	var docQualityScore float32
	var docQualityFeatures inference.DocQualityFeatures
	ocrConfidences := make(map[string]float32)
	ocrValidation := make(map[string]bool)
	var scopeTypes []string

	// Process each decrypted scope to extract features
	for _, scope := range input.DecryptedScopes {
		scopeTypes = append(scopeTypes, string(scope.ScopeType))

		// In a real implementation, we would parse the plaintext to extract
		// embeddings, quality scores, etc. For now, we use deterministic
		// values based on content hash.
		switch scope.ScopeType {
		case types.ScopeTypeSelfie:
			// Extract face embedding from selfie scope
			faceEmbedding = a.extractFaceEmbedding(scope)
			faceConfidence = a.computeConfidenceFromHash(scope.ContentHash)

		case types.ScopeTypeIDDocument:
			// Extract document quality and OCR from ID document
			docQualityScore = a.computeQualityFromHash(scope.ContentHash)
			docQualityFeatures = a.extractDocQualityFeatures(scope)
			ocrConfidences, ocrValidation = a.extractOCRFeatures(scope)
		}
	}

	// Ensure we have a valid face embedding
	if faceEmbedding == nil {
		faceEmbedding = make([]float32, inference.FaceEmbeddingDim)
	}

	return &inference.ScoreInputs{
		FaceEmbedding:      faceEmbedding,
		FaceConfidence:     faceConfidence,
		DocQualityScore:    docQualityScore,
		DocQualityFeatures: docQualityFeatures,
		OCRConfidences:     ocrConfidences,
		OCRFieldValidation: ocrValidation,
		Metadata: inference.InferenceMetadata{
			AccountAddress: input.AccountAddress,
			BlockHeight:    input.BlockHeight,
			BlockTime:      input.RequestTime,
		},
		ScopeTypes: scopeTypes,
		ScopeCount: len(input.DecryptedScopes),
	}
}

// extractFaceEmbedding extracts face embedding from scope content
// In production, this would parse the actual embedding from the scope
func (a *TensorFlowScorerAdapter) extractFaceEmbedding(scope DecryptedScope) []float32 {
	embedding := make([]float32, inference.FaceEmbeddingDim)

	// Generate deterministic embedding from content hash
	// In production, this would be the actual face embedding
	if len(scope.ContentHash) >= 4 {
		seed := uint32(scope.ContentHash[0])<<24 | uint32(scope.ContentHash[1])<<16 |
			uint32(scope.ContentHash[2])<<8 | uint32(scope.ContentHash[3])

		for i := range embedding {
			// Deterministic pseudo-random values based on hash
			seed = seed*1103515245 + 12345
			embedding[i] = float32(seed%1000) / 1000.0
		}
	}

	return embedding
}

// computeConfidenceFromHash computes a confidence score from content hash
func (a *TensorFlowScorerAdapter) computeConfidenceFromHash(hash []byte) float32 {
	if len(hash) < 4 {
		return 0.5
	}
	// Use hash bytes to generate deterministic confidence
	val := uint32(hash[0])<<24 | uint32(hash[1])<<16 | uint32(hash[2])<<8 | uint32(hash[3])
	return 0.5 + float32(val%50)/100.0 // Range: 0.5 - 1.0
}

// computeQualityFromHash computes quality score from content hash
func (a *TensorFlowScorerAdapter) computeQualityFromHash(hash []byte) float32 {
	if len(hash) < 4 {
		return 0.5
	}
	val := uint32(hash[0])<<16 | uint32(hash[1])<<8 | uint32(hash[2])
	return 0.6 + float32(val%40)/100.0 // Range: 0.6 - 1.0
}

// extractDocQualityFeatures extracts document quality features from scope
func (a *TensorFlowScorerAdapter) extractDocQualityFeatures(scope DecryptedScope) inference.DocQualityFeatures {
	hash := scope.ContentHash
	if len(hash) < 8 {
		return inference.DocQualityFeatures{
			Sharpness:  0.8,
			Brightness: 0.7,
			Contrast:   0.75,
			NoiseLevel: 0.1,
			BlurScore:  0.15,
		}
	}

	// Generate deterministic features from hash
	return inference.DocQualityFeatures{
		Sharpness:  0.7 + float32(hash[0]%30)/100.0,
		Brightness: 0.6 + float32(hash[1]%40)/100.0,
		Contrast:   0.65 + float32(hash[2]%35)/100.0,
		NoiseLevel: float32(hash[3]%20) / 100.0,
		BlurScore:  float32(hash[4]%25) / 100.0,
	}
}

// extractOCRFeatures extracts OCR features from scope
func (a *TensorFlowScorerAdapter) extractOCRFeatures(scope DecryptedScope) (map[string]float32, map[string]bool) {
	confidences := make(map[string]float32)
	validation := make(map[string]bool)

	fields := inference.OCRFieldNames
	hash := scope.ContentHash

	for i, field := range fields {
		// Generate deterministic confidence from hash
		if len(hash) > i {
			confidences[field] = 0.7 + float32(hash[i]%30)/100.0
			validation[field] = hash[i]%10 > 2 // ~70% validation rate
		} else {
			confidences[field] = 0.8
			validation[field] = true
		}
	}

	return confidences, validation
}

// convertReasonCodes converts inference reason codes to types.ReasonCode
func (a *TensorFlowScorerAdapter) convertReasonCodes(codes []string) []types.ReasonCode {
	result := make([]types.ReasonCode, 0, len(codes))

	for _, code := range codes {
		switch code {
		case inference.ReasonCodeSuccess:
			result = append(result, types.ReasonCodeSuccess)
		case inference.ReasonCodeHighConfidence:
			// No direct mapping, skip
		case inference.ReasonCodeLowConfidence:
			result = append(result, types.ReasonCodeLowConfidence)
		case inference.ReasonCodeFaceMismatch:
			result = append(result, types.ReasonCodeFaceMismatch)
		case inference.ReasonCodeLowDocQuality:
			result = append(result, types.ReasonCodeLowDocQuality)
		case inference.ReasonCodeLowOCRConfidence:
			result = append(result, types.ReasonCodeLowOCRConfidence)
		case inference.ReasonCodeInsufficientScopes:
			result = append(result, types.ReasonCodeInsufficientScopes)
		case inference.ReasonCodeMissingFace:
			result = append(result, types.ReasonCodeFaceMismatch)
		case inference.ReasonCodeMissingDocument:
			result = append(result, types.ReasonCodeInsufficientScopes)
		case inference.ReasonCodeTimeout:
			result = append(result, types.ReasonCodeMLInferenceError)
		case inference.ReasonCodeInferenceError:
			result = append(result, types.ReasonCodeMLInferenceError)
		}
	}

	return result
}

// GetModelVersion implements MLScorer.GetModelVersion
func (a *TensorFlowScorerAdapter) GetModelVersion() string {
	return a.scorer.GetModelVersion()
}

// IsHealthy implements MLScorer.IsHealthy
func (a *TensorFlowScorerAdapter) IsHealthy() bool {
	return a.scorer.IsHealthy()
}

// Close implements MLScorer.Close
func (a *TensorFlowScorerAdapter) Close() error {
	return a.scorer.Close()
}

// UpdateScopeVerificationResults updates the scope results based on ML scoring output
func (k Keeper) UpdateScopeVerificationResults(
	results []types.ScopeVerificationResult,
	scopeScores map[string]uint32,
) []types.ScopeVerificationResult {
	updated := make([]types.ScopeVerificationResult, len(results))
	copy(updated, results)

	for i := range updated {
		if score, ok := scopeScores[updated[i].ScopeID]; ok {
			if score >= types.ThresholdBasic/2 { // Scope passes if above half of basic threshold
				updated[i].SetSuccess(score)
			} else {
				updated[i].SetFailure(types.ReasonCodeFaceMismatch)
			}
		}
	}

	return updated
}
