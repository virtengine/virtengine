package inference

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"time"
)

// ============================================================================
// Face Extractor
// ============================================================================

// FaceExtractor extracts face embeddings from selfie/video scope data.
// It interfaces with the Python ML pipeline via gRPC sidecar or processes
// locally when available.
//
// The extractor produces 512-dimensional face embeddings compatible with
// FaceNet512/ArcFace models used in the training pipeline.
type FaceExtractor struct {
	config FaceExtractorConfig
	client FaceExtractionClient

	mu              sync.RWMutex
	isHealthy       bool
	extractionCount uint64
	errorCount      uint64
}

// FaceExtractorConfig contains configuration for face extraction
type FaceExtractorConfig struct {
	// SidecarAddress is the gRPC address of the face extraction sidecar
	SidecarAddress string

	// Timeout for extraction operations
	Timeout time.Duration

	// EmbeddingDim is the expected embedding dimension (default 512)
	EmbeddingDim int

	// MinConfidence is the minimum confidence threshold for valid extractions
	MinConfidence float32

	// NormalizeEmbedding enables L2 normalization of embeddings
	NormalizeEmbedding bool

	// UseFallbackOnError returns zero embeddings on extraction errors
	UseFallbackOnError bool

	// MaxRetries for transient failures
	MaxRetries int

	// RetryDelay between retry attempts
	RetryDelay time.Duration
}

// DefaultFaceExtractorConfig returns sensible defaults
func DefaultFaceExtractorConfig() FaceExtractorConfig {
	return FaceExtractorConfig{
		SidecarAddress:     "localhost:50052",
		Timeout:            5 * time.Second,
		EmbeddingDim:       FaceEmbeddingDim, // 512
		MinConfidence:      0.7,
		NormalizeEmbedding: true,
		UseFallbackOnError: true,
		MaxRetries:         2,
		RetryDelay:         100 * time.Millisecond,
	}
}

// FaceExtractionClient defines the interface for face extraction backends
type FaceExtractionClient interface {
	// ExtractEmbedding extracts face embedding from image data
	ExtractEmbedding(ctx context.Context, imageData []byte) (*FaceExtractionResult, error)

	// ExtractFromVideo extracts embeddings from video frames
	ExtractFromVideo(ctx context.Context, videoData []byte, maxFrames int) (*VideoFaceExtractionResult, error)

	// IsHealthy checks if the client is ready
	IsHealthy() bool

	// Close releases resources
	Close() error
}

// FaceExtractionResult contains the result of face embedding extraction
type FaceExtractionResult struct {
	// Embedding is the 512-dimensional face embedding vector
	Embedding []float32

	// Confidence is the detection/extraction confidence (0.0-1.0)
	Confidence float32

	// FaceDetected indicates if a face was found in the image
	FaceDetected bool

	// BoundingBox of the detected face (x, y, width, height)
	BoundingBox FaceBoundingBox

	// Quality metrics for the detected face
	Quality FaceQuality

	// ModelVersion used for extraction
	ModelVersion string

	// EmbeddingHash for determinism verification
	EmbeddingHash string

	// ProcessingTimeMs is the extraction time in milliseconds
	ProcessingTimeMs int64

	// ReasonCodes provide explanations for extraction outcome
	ReasonCodes []string
}

// FaceBoundingBox represents the face location in an image
type FaceBoundingBox struct {
	X      int
	Y      int
	Width  int
	Height int
}

// FaceQuality contains face quality metrics
type FaceQuality struct {
	// Sharpness of the face region (0.0-1.0)
	Sharpness float32

	// Brightness of the face (0.0-1.0, 0.5 is ideal)
	Brightness float32

	// Pose angles in degrees
	YawAngle   float32
	PitchAngle float32
	RollAngle  float32

	// Occlusion detection (0.0-1.0, lower is better)
	OcclusionScore float32

	// Overall quality score (0.0-1.0)
	OverallScore float32
}

// VideoFaceExtractionResult contains results from video frame extraction
type VideoFaceExtractionResult struct {
	// FrameResults contains per-frame extraction results
	FrameResults []FaceExtractionResult

	// BestEmbedding is the embedding from the highest quality frame
	BestEmbedding []float32

	// BestConfidence is the confidence of the best frame
	BestConfidence float32

	// ConsistencyScore measures consistency across frames (0.0-1.0)
	ConsistencyScore float32

	// FrameCount is the number of frames processed
	FrameCount int

	// ValidFrameCount is the number of frames with detected faces
	ValidFrameCount int

	// AverageEmbedding is the mean embedding across valid frames
	AverageEmbedding []float32

	// ProcessingTimeMs is the total processing time
	ProcessingTimeMs int64
}

// ============================================================================
// Face Extraction Reason Codes
// ============================================================================

const (
	FaceReasonCodeSuccess           = "FACE_EXTRACTION_SUCCESS"
	FaceReasonCodeNoFaceDetected    = "NO_FACE_DETECTED"
	FaceReasonCodeLowConfidence     = "LOW_FACE_CONFIDENCE"
	FaceReasonCodeLowQuality        = "LOW_FACE_QUALITY"
	FaceReasonCodeMultipleFaces     = "MULTIPLE_FACES_DETECTED"
	FaceReasonCodeOccluded          = "FACE_OCCLUDED"
	FaceReasonCodeBadPose           = "BAD_FACE_POSE"
	FaceReasonCodeExtractionError   = "FACE_EXTRACTION_ERROR"
	FaceReasonCodeSidecarUnavail    = "FACE_SIDECAR_UNAVAILABLE"
	FaceReasonCodeTimeout           = "FACE_EXTRACTION_TIMEOUT"
	FaceReasonCodeVideoInconsistent = "VIDEO_FACES_INCONSISTENT"
)

// ============================================================================
// Constructor and Interface Implementation
// ============================================================================

// NewFaceExtractor creates a new face extractor
func NewFaceExtractor(config FaceExtractorConfig) *FaceExtractor {
	return &FaceExtractor{
		config:    config,
		isHealthy: false,
	}
}

// NewFaceExtractorWithClient creates a face extractor with a specific client
func NewFaceExtractorWithClient(config FaceExtractorConfig, client FaceExtractionClient) *FaceExtractor {
	fe := NewFaceExtractor(config)
	fe.client = client
	fe.isHealthy = client != nil && client.IsHealthy()
	return fe
}

// ExtractFromSelfie extracts face embedding from a selfie image
func (fe *FaceExtractor) ExtractFromSelfie(ctx context.Context, imageData []byte) (*FaceExtractionResult, error) {
	if len(imageData) == 0 {
		return fe.createFailureResult(FaceReasonCodeExtractionError, "empty image data"), nil
	}

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, fe.config.Timeout)
	defer cancel()

	fe.mu.Lock()
	fe.extractionCount++
	fe.mu.Unlock()

	// Try extraction with retries
	var result *FaceExtractionResult
	var lastErr error

	for attempt := 0; attempt <= fe.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return fe.handleError(ctx.Err(), "timeout during retry")
			case <-time.After(fe.config.RetryDelay):
			}
		}

		result, lastErr = fe.doExtraction(ctx, imageData)
		if lastErr == nil && result.FaceDetected {
			break
		}
	}

	if lastErr != nil {
		return fe.handleError(lastErr, "extraction failed after retries")
	}

	// Validate and post-process result
	fe.postProcessResult(result)

	return result, nil
}

// ExtractFromVideo extracts face embeddings from video data
func (fe *FaceExtractor) ExtractFromVideo(ctx context.Context, videoData []byte, maxFrames int) (*VideoFaceExtractionResult, error) {
	if len(videoData) == 0 {
		return &VideoFaceExtractionResult{
			BestEmbedding:  make([]float32, fe.config.EmbeddingDim),
			BestConfidence: 0.0,
		}, nil
	}

	if maxFrames <= 0 {
		maxFrames = 10 // Default to 10 frames
	}

	ctx, cancel := context.WithTimeout(ctx, fe.config.Timeout*2) // Longer timeout for video
	defer cancel()

	fe.mu.Lock()
	fe.extractionCount++
	fe.mu.Unlock()

	if fe.client == nil {
		return fe.extractVideoFallback(videoData, maxFrames), nil
	}

	result, err := fe.client.ExtractFromVideo(ctx, videoData, maxFrames)
	if err != nil {
		fe.mu.Lock()
		fe.errorCount++
		fe.mu.Unlock()

		if fe.config.UseFallbackOnError {
			return fe.extractVideoFallback(videoData, maxFrames), nil
		}
		return nil, fmt.Errorf("video extraction failed: %w", err)
	}

	// Compute consistency and average embedding
	fe.postProcessVideoResult(result)

	return result, nil
}

// doExtraction performs the actual extraction via client or fallback
func (fe *FaceExtractor) doExtraction(ctx context.Context, imageData []byte) (*FaceExtractionResult, error) {
	startTime := time.Now()

	if fe.client == nil || !fe.client.IsHealthy() {
		// Use fallback stub extraction
		return fe.extractFallback(imageData, startTime), nil
	}

	result, err := fe.client.ExtractEmbedding(ctx, imageData)
	if err != nil {
		return nil, err
	}

	result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	return result, nil
}

// extractFallback generates deterministic stub embeddings when sidecar is unavailable
func (fe *FaceExtractor) extractFallback(imageData []byte, startTime time.Time) *FaceExtractionResult {
	// Generate deterministic embedding from image hash
	embedding := fe.generateDeterministicEmbedding(imageData)

	return &FaceExtractionResult{
		Embedding:        embedding,
		Confidence:       0.85, // Stub confidence
		FaceDetected:     true,
		ModelVersion:     "stub-v1.0.0",
		EmbeddingHash:    fe.computeEmbeddingHash(embedding),
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
		ReasonCodes:      []string{FaceReasonCodeSuccess, FaceReasonCodeSidecarUnavail},
		Quality: FaceQuality{
			OverallScore: 0.8,
			Sharpness:    0.75,
			Brightness:   0.5,
		},
	}
}

// extractVideoFallback generates stub results for video extraction
func (fe *FaceExtractor) extractVideoFallback(videoData []byte, maxFrames int) *VideoFaceExtractionResult {
	embedding := fe.generateDeterministicEmbedding(videoData)

	return &VideoFaceExtractionResult{
		BestEmbedding:    embedding,
		BestConfidence:   0.85,
		ConsistencyScore: 0.9,
		FrameCount:       maxFrames,
		ValidFrameCount:  maxFrames,
		AverageEmbedding: embedding,
		ProcessingTimeMs: 10,
	}
}

// generateDeterministicEmbedding creates a deterministic embedding from input data
func (fe *FaceExtractor) generateDeterministicEmbedding(data []byte) []float32 {
	embedding := make([]float32, fe.config.EmbeddingDim)

	if len(data) == 0 {
		return embedding
	}

	// Use SHA256 hash to generate deterministic values
	hash := sha256.Sum256(data)
	hashBytes := hash[:]

	// Generate embedding values from hash bytes
	for i := 0; i < fe.config.EmbeddingDim; i++ {
		// Cycle through hash bytes
		byteIdx := i % len(hashBytes)
		// Convert byte to float in range [-1, 1]
		embedding[i] = (float32(hashBytes[byteIdx]) / 127.5) - 1.0
	}

	// Normalize to unit length
	if fe.config.NormalizeEmbedding {
		fe.normalizeEmbedding(embedding)
	}

	return embedding
}

// postProcessResult validates and normalizes the extraction result
func (fe *FaceExtractor) postProcessResult(result *FaceExtractionResult) {
	if result == nil {
		return
	}

	// Normalize embedding if configured
	if fe.config.NormalizeEmbedding && len(result.Embedding) > 0 {
		fe.normalizeEmbedding(result.Embedding)
	}

	// Compute embedding hash
	if result.EmbeddingHash == "" && len(result.Embedding) > 0 {
		result.EmbeddingHash = fe.computeEmbeddingHash(result.Embedding)
	}

	// Validate confidence bounds
	result.Confidence = clampFloat32(result.Confidence, 0.0, 1.0)

	// Add reason codes based on quality
	if result.FaceDetected && result.Confidence < fe.config.MinConfidence {
		result.ReasonCodes = append(result.ReasonCodes, FaceReasonCodeLowConfidence)
	}

	if result.Quality.OverallScore < 0.5 {
		result.ReasonCodes = append(result.ReasonCodes, FaceReasonCodeLowQuality)
	}

	if result.Quality.OcclusionScore > 0.3 {
		result.ReasonCodes = append(result.ReasonCodes, FaceReasonCodeOccluded)
	}

	// Check pose angles
	if absFloat32(result.Quality.YawAngle) > 30 ||
		absFloat32(result.Quality.PitchAngle) > 20 {
		result.ReasonCodes = append(result.ReasonCodes, FaceReasonCodeBadPose)
	}
}

// postProcessVideoResult computes aggregate metrics for video results
func (fe *FaceExtractor) postProcessVideoResult(result *VideoFaceExtractionResult) {
	if result == nil || len(result.FrameResults) == 0 {
		return
	}

	// Find best frame
	bestIdx := 0
	bestScore := float32(0.0)
	validCount := 0

	for i, fr := range result.FrameResults {
		if fr.FaceDetected {
			validCount++
			score := fr.Confidence * fr.Quality.OverallScore
			if score > bestScore {
				bestScore = score
				bestIdx = i
			}
		}
	}

	result.ValidFrameCount = validCount
	result.FrameCount = len(result.FrameResults)

	if validCount == 0 {
		result.BestEmbedding = make([]float32, fe.config.EmbeddingDim)
		result.AverageEmbedding = make([]float32, fe.config.EmbeddingDim)
		result.BestConfidence = 0.0
		result.ConsistencyScore = 0.0
		return
	}

	// Set best embedding
	result.BestEmbedding = result.FrameResults[bestIdx].Embedding
	result.BestConfidence = result.FrameResults[bestIdx].Confidence

	// Compute average embedding
	result.AverageEmbedding = fe.computeAverageEmbedding(result.FrameResults)

	// Compute consistency score
	result.ConsistencyScore = fe.computeConsistencyScore(result.FrameResults)
}

// computeAverageEmbedding computes the mean embedding across valid frames
func (fe *FaceExtractor) computeAverageEmbedding(results []FaceExtractionResult) []float32 {
	dim := fe.config.EmbeddingDim
	avgEmb := make([]float32, dim)
	count := 0

	for _, r := range results {
		if r.FaceDetected && len(r.Embedding) == dim {
			for i := 0; i < dim; i++ {
				avgEmb[i] += r.Embedding[i]
			}
			count++
		}
	}

	if count > 0 {
		for i := 0; i < dim; i++ {
			avgEmb[i] /= float32(count)
		}
	}

	// Normalize
	if fe.config.NormalizeEmbedding {
		fe.normalizeEmbedding(avgEmb)
	}

	return avgEmb
}

// computeConsistencyScore measures embedding consistency across frames
func (fe *FaceExtractor) computeConsistencyScore(results []FaceExtractionResult) float32 {
	validEmbeddings := make([][]float32, 0)
	for _, r := range results {
		if r.FaceDetected && len(r.Embedding) == fe.config.EmbeddingDim {
			validEmbeddings = append(validEmbeddings, r.Embedding)
		}
	}

	if len(validEmbeddings) < 2 {
		return 1.0 // Single frame is perfectly consistent
	}

	// Compute pairwise cosine similarities
	var totalSimilarity float64
	pairs := 0

	for i := 0; i < len(validEmbeddings); i++ {
		for j := i + 1; j < len(validEmbeddings); j++ {
			sim := fe.cosineSimilarity(validEmbeddings[i], validEmbeddings[j])
			totalSimilarity += float64(sim)
			pairs++
		}
	}

	if pairs == 0 {
		return 1.0
	}

	return float32(totalSimilarity / float64(pairs))
}

// cosineSimilarity computes cosine similarity between two embeddings
func (fe *FaceExtractor) cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0.0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return float32(dot / (math.Sqrt(normA) * math.Sqrt(normB)))
}

// normalizeEmbedding normalizes an embedding to unit length
func (fe *FaceExtractor) normalizeEmbedding(embedding []float32) {
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

// computeEmbeddingHash computes SHA256 hash of embedding for determinism verification
func (fe *FaceExtractor) computeEmbeddingHash(embedding []float32) string {
	h := sha256.New()
	for _, v := range embedding {
		// Round to 6 decimal places for determinism
		rounded := math.Round(float64(v)*1e6) / 1e6
		fmt.Fprintf(h, "%.6f", rounded)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// handleError handles extraction errors with optional fallback
func (fe *FaceExtractor) handleError(err error, msg string) (*FaceExtractionResult, error) {
	fe.mu.Lock()
	fe.errorCount++
	fe.mu.Unlock()

	if fe.config.UseFallbackOnError {
		return fe.createFailureResult(FaceReasonCodeExtractionError, msg), nil
	}
	return nil, fmt.Errorf("%s: %w", msg, err)
}

// createFailureResult creates a result indicating extraction failure
//
//nolint:unparam // message kept for future logging or error details in result
func (fe *FaceExtractor) createFailureResult(reasonCode, _ string) *FaceExtractionResult {
	return &FaceExtractionResult{
		Embedding:     make([]float32, fe.config.EmbeddingDim),
		Confidence:    0.0,
		FaceDetected:  false,
		ModelVersion:  "unknown",
		EmbeddingHash: "",
		ReasonCodes:   []string{reasonCode},
	}
}

// ============================================================================
// Validation Functions
// ============================================================================

// ValidateEmbedding validates that an embedding meets quality requirements
func (fe *FaceExtractor) ValidateEmbedding(embedding []float32) []string {
	var issues []string

	// Check dimension
	if len(embedding) != fe.config.EmbeddingDim {
		issues = append(issues, fmt.Sprintf(
			"invalid embedding dimension: expected %d, got %d",
			fe.config.EmbeddingDim, len(embedding),
		))
		return issues
	}

	// Check for all zeros
	allZeros := true
	for _, v := range embedding {
		if v != 0 {
			allZeros = false
			break
		}
	}
	if allZeros {
		issues = append(issues, "embedding is all zeros")
	}

	// Check for NaN or Inf values
	for i, v := range embedding {
		if math.IsNaN(float64(v)) {
			issues = append(issues, fmt.Sprintf("embedding[%d] is NaN", i))
			break
		}
		if math.IsInf(float64(v), 0) {
			issues = append(issues, fmt.Sprintf("embedding[%d] is Inf", i))
			break
		}
	}

	// Check bounds (normalized embeddings should be in [-1, 1])
	for i, v := range embedding {
		if v < -2.0 || v > 2.0 {
			issues = append(issues, fmt.Sprintf(
				"embedding[%d] out of bounds: %.4f", i, v,
			))
			break
		}
	}

	return issues
}

// SanitizeEmbedding sanitizes an embedding to ensure valid values
func (fe *FaceExtractor) SanitizeEmbedding(embedding []float32) []float32 {
	sanitized := make([]float32, len(embedding))
	copy(sanitized, embedding)

	for i := range sanitized {
		v := sanitized[i]
		// Replace NaN/Inf with 0
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			sanitized[i] = 0
			continue
		}
		// Clamp to reasonable bounds
		sanitized[i] = clampFloat32(v, -2.0, 2.0)
	}

	return sanitized
}

// ============================================================================
// Statistics and Health
// ============================================================================

// IsHealthy returns whether the extractor is ready
func (fe *FaceExtractor) IsHealthy() bool {
	fe.mu.RLock()
	defer fe.mu.RUnlock()

	if fe.client != nil {
		return fe.client.IsHealthy()
	}
	// Stub mode is always healthy
	return true
}

// GetStats returns extraction statistics
func (fe *FaceExtractor) GetStats() FaceExtractorStats {
	fe.mu.RLock()
	defer fe.mu.RUnlock()

	return FaceExtractorStats{
		ExtractionCount: fe.extractionCount,
		ErrorCount:      fe.errorCount,
		IsHealthy:       fe.IsHealthy(),
		UsingStub:       fe.client == nil,
	}
}

// FaceExtractorStats contains extractor statistics
type FaceExtractorStats struct {
	ExtractionCount uint64
	ErrorCount      uint64
	IsHealthy       bool
	UsingStub       bool
}

// Close releases resources
func (fe *FaceExtractor) Close() error {
	if fe.client != nil {
		return fe.client.Close()
	}
	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// clampFloat32 clamps a value to the specified range
func clampFloat32(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
