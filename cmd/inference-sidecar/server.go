// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0
//
// Server implementation for inference sidecar.

package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/virtengine/virtengine/pkg/inference"
	inferencepb "github.com/virtengine/virtengine/pkg/inference/proto"
)

// InferenceSidecarServer implements the InferenceService gRPC server.
type InferenceSidecarServer struct {
	inferencepb.UnimplementedInferenceServiceServer

	// config holds inference configuration
	config inference.InferenceConfig

	// scorer is the TensorFlow scorer
	scorer inference.Scorer

	// model holds loaded model metadata
	model *inference.TFModel

	// loader handles model loading
	loader *inference.ModelLoader

	// determinism ensures deterministic hashing
	determinism *inference.DeterminismController

	// extractor transforms inputs to features
	extractor *inference.FeatureExtractor

	// log is the logger
	log Logger

	// Metrics
	startTime            time.Time
	totalInferences      atomic.Uint64
	successfulInferences atomic.Uint64
	failedInferences     atomic.Uint64
	lastInferenceTime    atomic.Int64

	// Latency tracking
	latencyMu         sync.RWMutex
	latencySum        float64
	latencyCount      int64
	latencyHistogram  map[string]uint64
	latencyPercentile []float64
}

// NewInferenceSidecarServer creates a new inference sidecar server.
func NewInferenceSidecarServer(config inference.InferenceConfig, log Logger) (*InferenceSidecarServer, error) {
	// Set determinism environment variables
	determinism := inference.NewDeterminismController(config.RandomSeed, config.ForceCPU)
	for k, v := range determinism.GetTensorFlowEnvVars() {
		if err := setEnvIfNotSet(k, v); err != nil {
			log.Warn("Failed to set env var", "key", k, "error", err)
		}
	}

	log.Info("Creating inference server",
		"model_path", config.ModelPath,
		"model_version", config.ModelVersion,
		"force_cpu", config.ForceCPU,
		"random_seed", config.RandomSeed,
	)

	// Create model loader
	loader := inference.NewModelLoader(config)

	// Load the model
	model, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load model: %w", err)
	}

	log.Info("Model loaded successfully",
		"version", model.GetVersion(),
		"hash", model.GetModelHash(),
	)

	// Verify model hash if expected
	if config.ExpectedHash != "" && model.GetModelHash() != config.ExpectedHash {
		return nil, fmt.Errorf("model hash mismatch: expected %s, got %s",
			config.ExpectedHash, model.GetModelHash())
	}

	server := &InferenceSidecarServer{
		config:           config,
		model:            model,
		loader:           loader,
		determinism:      determinism,
		extractor:        inference.NewFeatureExtractor(inference.DefaultFeatureExtractorConfig()),
		log:              log,
		startTime:        time.Now(),
		latencyHistogram: make(map[string]uint64),
	}

	// Create a minimal scorer wrapper
	server.scorer = &modelScorer{
		model:       model,
		determinism: determinism,
		config:      config,
	}

	return server, nil
}

// Close releases resources held by the server.
func (s *InferenceSidecarServer) Close() error {
	if s.loader != nil {
		return s.loader.Unload()
	}
	return nil
}

// ============================================================================
// gRPC Service Methods
// ============================================================================

// GetModelInfo implements InferenceServiceServer.GetModelInfo.
func (s *InferenceSidecarServer) GetModelInfo(ctx context.Context, req *inferencepb.GetModelInfoRequest) (*inferencepb.GetModelInfoResponse, error) {
	s.log.Debug("GetModelInfo called")

	metadata := s.model.GetMetadata()
	tfVersion := ""
	exportTimestamp := ""
	if metadata != nil {
		tfVersion = metadata.TensorFlowVersion
		exportTimestamp = metadata.ExportTimestamp
	}

	return &inferencepb.GetModelInfoResponse{
		Version:           s.model.GetVersion(),
		Hash:              s.model.GetModelHash(),
		InputDim:          int32(inference.TotalFeatureDim),
		OutputDim:         1,
		TensorFlowVersion: tfVersion,
		ExportTimestamp:   exportTimestamp,
		PipelineVersion:   s.config.ModelVersion,
		DeterminismConfig: &inferencepb.DeterminismConfig{
			ForceCPU:           s.config.ForceCPU,
			RandomSeed:         s.config.RandomSeed,
			InterOpParallelism: 1,
			IntraOpParallelism: 1,
			DeterministicOps:   true,
		},
	}, nil
}

// ComputeScore implements InferenceServiceServer.ComputeScore.
func (s *InferenceSidecarServer) ComputeScore(ctx context.Context, req *inferencepb.ComputeScoreRequest) (*inferencepb.ComputeScoreResponse, error) {
	startTime := time.Now()
	s.totalInferences.Add(1)

	s.log.Debug("ComputeScore called",
		"features_len", len(req.Features),
		"account", req.Metadata.GetAccountAddress(),
		"block_height", req.Metadata.GetBlockHeight(),
	)

	// Validate input dimension
	if len(req.Features) != inference.TotalFeatureDim {
		s.failedInferences.Add(1)
		return nil, fmt.Errorf("invalid feature dimension: expected %d, got %d",
			inference.TotalFeatureDim, len(req.Features))
	}

	// Build ScoreInputs from features
	// Note: Timeout configured in s.config.Timeout but not applied here as
	// the underlying scorer doesn't support context-based cancellation yet
	inputs := s.buildScoreInputs(req)

	// Run inference
	result, err := s.scorer.ComputeScore(inputs)
	if err != nil {
		s.failedInferences.Add(1)
		s.log.Error("Inference failed", "error", err)
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	// Track latency
	latencyMs := time.Since(startTime).Milliseconds()
	s.recordLatency(float64(latencyMs))
	s.lastInferenceTime.Store(time.Now().UnixNano())
	s.successfulInferences.Add(1)

	// Build response
	resp := &inferencepb.ComputeScoreResponse{
		Score:         result.Score,
		RawScore:      result.RawScore,
		Confidence:    result.Confidence,
		InputHash:     result.InputHash,
		OutputHash:    result.OutputHash,
		ModelVersion:  result.ModelVersion,
		ModelHash:     result.ModelHash,
		ReasonCodes:   result.ReasonCodes,
		ComputeTimeMs: latencyMs,
	}

	if req.ReturnContributions && result.FeatureContributions != nil {
		resp.FeatureContributions = result.FeatureContributions
	}

	s.log.Debug("ComputeScore completed",
		"score", result.Score,
		"confidence", result.Confidence,
		"latency_ms", latencyMs,
	)

	return resp, nil
}

// HealthCheck implements InferenceServiceServer.HealthCheck.
func (s *InferenceSidecarServer) HealthCheck(ctx context.Context, req *inferencepb.HealthCheckRequest) (*inferencepb.HealthCheckResponse, error) {
	s.log.Debug("HealthCheck called")

	status := inferencepb.HealthStatus_HEALTH_STATUS_HEALTHY
	errorMsg := ""

	if s.model == nil || !s.model.IsLoaded() {
		status = inferencepb.HealthStatus_HEALTH_STATUS_UNHEALTHY
		errorMsg = "model not loaded"
	} else if !s.scorer.IsHealthy() {
		status = inferencepb.HealthStatus_HEALTH_STATUS_DEGRADED
		errorMsg = "scorer not healthy"
	}

	lastInference := ""
	if ts := s.lastInferenceTime.Load(); ts > 0 {
		lastInference = time.Unix(0, ts).Format(time.RFC3339)
	}

	return &inferencepb.HealthCheckResponse{
		Status:                 status,
		ModelLoaded:            s.model != nil && s.model.IsLoaded(),
		ModelVersion:           s.model.GetVersion(),
		ModelHash:              s.model.GetModelHash(),
		UptimeSeconds:          int64(time.Since(s.startTime).Seconds()),
		LastInferenceTimestamp: lastInference,
		ErrorMessage:           errorMsg,
	}, nil
}

// GetMetrics implements InferenceServiceServer.GetMetrics.
func (s *InferenceSidecarServer) GetMetrics(ctx context.Context, req *inferencepb.GetMetricsRequest) (*inferencepb.GetMetricsResponse, error) {
	s.log.Debug("GetMetrics called")

	s.latencyMu.RLock()
	avgLatency := float32(0)
	if s.latencyCount > 0 {
		avgLatency = float32(s.latencySum / float64(s.latencyCount))
	}
	p99Latency := float32(0)
	if len(s.latencyPercentile) > 0 {
		idx := int(float64(len(s.latencyPercentile)) * 0.99)
		if idx >= len(s.latencyPercentile) {
			idx = len(s.latencyPercentile) - 1
		}
		p99Latency = float32(s.latencyPercentile[idx])
	}
	histogramCopy := make(map[string]uint64)
	for k, v := range s.latencyHistogram {
		histogramCopy[k] = v
	}
	s.latencyMu.RUnlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &inferencepb.GetMetricsResponse{
		TotalInferences:      s.totalInferences.Load(),
		SuccessfulInferences: s.successfulInferences.Load(),
		FailedInferences:     s.failedInferences.Load(),
		AverageLatencyMs:     avgLatency,
		P99LatencyMs:         p99Latency,
		ModelVersion:         s.model.GetVersion(),
		ModelHash:            s.model.GetModelHash(),
		UptimeSeconds:        int64(time.Since(s.startTime).Seconds()),
		MemoryUsageMB:        float32(memStats.Alloc) / (1024 * 1024),
		LatencyHistogram:     histogramCopy,
	}, nil
}

// VerifyDeterminism implements InferenceServiceServer.VerifyDeterminism.
func (s *InferenceSidecarServer) VerifyDeterminism(ctx context.Context, req *inferencepb.VerifyDeterminismRequest) (*inferencepb.VerifyDeterminismResponse, error) {
	s.log.Debug("VerifyDeterminism called", "test_vector_id", req.TestVectorID)

	var features []float32
	var expectedHash string
	testVectorID := req.TestVectorID

	if req.TestVectorID != "" {
		// Use predefined test vector
		tv := inference.GetTestVector(req.TestVectorID)
		if tv == nil {
			return nil, fmt.Errorf("test vector not found: %s", req.TestVectorID)
		}
		features = tv.Features
		expectedHash = tv.ExpectedOutputHash
	} else if len(req.CustomInput) > 0 {
		// Use custom input
		features = req.CustomInput
		expectedHash = req.ExpectedOutputHash
		testVectorID = "custom"
	} else {
		// Use default test vector
		tv := inference.GetDefaultTestVector()
		features = tv.Features
		expectedHash = tv.ExpectedOutputHash
		testVectorID = tv.ID
	}

	// Validate feature dimension
	if len(features) != inference.TotalFeatureDim {
		return nil, fmt.Errorf("invalid feature dimension: expected %d, got %d",
			inference.TotalFeatureDim, len(features))
	}

	// Run inference
	inputs := &inference.ScoreInputs{
		FaceEmbedding: features[:inference.FaceEmbeddingDim],
		Metadata: inference.InferenceMetadata{
			RequestID: "determinism-check",
		},
		ScopeCount: 1,
	}

	result, err := s.scorer.ComputeScore(inputs)
	if err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	passed := result.OutputHash == expectedHash
	var differences []string
	if !passed && expectedHash != "" {
		differences = append(differences,
			fmt.Sprintf("output hash mismatch: expected %s, got %s", expectedHash, result.OutputHash))
	}

	return &inferencepb.VerifyDeterminismResponse{
		Passed:             passed || expectedHash == "",
		ActualOutputHash:   result.OutputHash,
		ExpectedOutputHash: expectedHash,
		Differences:        differences,
		TestVectorID:       testVectorID,
	}, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// buildScoreInputs creates ScoreInputs from a gRPC request.
func (s *InferenceSidecarServer) buildScoreInputs(req *inferencepb.ComputeScoreRequest) *inference.ScoreInputs {
	// Extract feature components from the flat vector
	faceEnd := inference.FaceEmbeddingDim
	docEnd := faceEnd + inference.DocQualityDim

	// Handle nil metadata
	var accountAddr string
	var blockHeight int64
	var requestID string
	if req.Metadata != nil {
		accountAddr = req.Metadata.AccountAddress
		blockHeight = req.Metadata.BlockHeight
		requestID = req.Metadata.RequestID
	}

	inputs := &inference.ScoreInputs{
		FaceEmbedding:   req.Features[:faceEnd],
		FaceConfidence:  0.9, // Default
		DocQualityScore: 0.8, // Default
		OCRConfidences:  make(map[string]float32),
		Metadata: inference.InferenceMetadata{
			AccountAddress: accountAddr,
			BlockHeight:    blockHeight,
			RequestID:      requestID,
		},
		ScopeCount: 1,
	}

	// Extract document quality features
	if len(req.Features) > docEnd {
		inputs.DocQualityFeatures = inference.DocQualityFeatures{
			Sharpness:  req.Features[faceEnd],
			Brightness: req.Features[faceEnd+1],
			Contrast:   req.Features[faceEnd+2],
			NoiseLevel: req.Features[faceEnd+3],
			BlurScore:  req.Features[faceEnd+4],
		}
	}

	return inputs
}

// recordLatency records a latency sample.
func (s *InferenceSidecarServer) recordLatency(latencyMs float64) {
	s.latencyMu.Lock()
	defer s.latencyMu.Unlock()

	s.latencySum += latencyMs
	s.latencyCount++
	s.latencyPercentile = append(s.latencyPercentile, latencyMs)

	// Keep percentile array sorted and bounded
	if len(s.latencyPercentile) > 10000 {
		sort.Float64s(s.latencyPercentile)
		// Keep last 5000 samples
		s.latencyPercentile = s.latencyPercentile[5000:]
	}

	// Update histogram buckets
	bucket := latencyBucket(latencyMs)
	s.latencyHistogram[bucket]++
}

// latencyBucket returns the histogram bucket for a latency value.
func latencyBucket(latencyMs float64) string {
	switch {
	case latencyMs < 10:
		return "<10ms"
	case latencyMs < 50:
		return "10-50ms"
	case latencyMs < 100:
		return "50-100ms"
	case latencyMs < 500:
		return "100-500ms"
	case latencyMs < 1000:
		return "500ms-1s"
	default:
		return ">1s"
	}
}

// setEnvIfNotSet sets an environment variable if it's not already set.
func setEnvIfNotSet(key, value string) error {
	if _, exists := os.LookupEnv(key); !exists {
		return os.Setenv(key, value)
	}
	return nil
}

// ============================================================================
// Model Scorer Wrapper
// ============================================================================

// modelScorer wraps TFModel to implement the Scorer interface.
type modelScorer struct {
	model       *inference.TFModel
	determinism *inference.DeterminismController
	config      inference.InferenceConfig
}

func (s *modelScorer) ComputeScore(inputs *inference.ScoreInputs) (*inference.ScoreResult, error) {
	startTime := time.Now()

	result := &inference.ScoreResult{
		ModelVersion:         s.model.GetVersion(),
		ModelHash:            s.model.GetModelHash(),
		ReasonCodes:          make([]string, 0),
		FeatureContributions: make(map[string]float32),
	}

	// Compute input hash
	result.InputHash = s.determinism.ComputeInputHash(inputs)

	// Build feature vector
	features := buildFeatureVector(inputs)

	// Run model inference
	output, err := s.model.Run(features)
	if err != nil {
		result.ReasonCodes = append(result.ReasonCodes, inference.ReasonCodeInferenceError)
		return result, fmt.Errorf("model inference failed: %w", err)
	}

	if len(output) == 0 {
		result.ReasonCodes = append(result.ReasonCodes, inference.ReasonCodeInferenceError)
		return result, fmt.Errorf("model returned empty output")
	}

	// Process output
	rawScore := output[0]
	result.RawScore = rawScore
	result.OutputHash = s.determinism.ComputeOutputHash(output)

	// Quantize to 0-100
	score := uint32(rawScore)
	if rawScore < 0 {
		score = 0
	} else if rawScore > 100 {
		score = 100
	}
	result.Score = score

	// Compute confidence
	result.Confidence = computeConfidence(rawScore)

	// Add reason codes
	if score >= 50 {
		result.ReasonCodes = append(result.ReasonCodes, inference.ReasonCodeSuccess)
	}
	if result.Confidence >= 0.8 {
		result.ReasonCodes = append(result.ReasonCodes, inference.ReasonCodeHighConfidence)
	}

	result.ComputeTimeMs = time.Since(startTime).Milliseconds()
	return result, nil
}

func (s *modelScorer) GetModelVersion() string {
	return s.model.GetVersion()
}

func (s *modelScorer) GetModelHash() string {
	return s.model.GetModelHash()
}

func (s *modelScorer) IsHealthy() bool {
	return s.model != nil && s.model.IsLoaded()
}

func (s *modelScorer) Close() error {
	return nil
}

// buildFeatureVector constructs the feature vector from ScoreInputs.
func buildFeatureVector(inputs *inference.ScoreInputs) []float32 {
	features := make([]float32, inference.TotalFeatureDim)

	// Copy face embedding
	copy(features[:inference.FaceEmbeddingDim], inputs.FaceEmbedding)

	// Add document quality features
	offset := inference.FaceEmbeddingDim
	features[offset] = inputs.DocQualityFeatures.Sharpness
	features[offset+1] = inputs.DocQualityFeatures.Brightness
	features[offset+2] = inputs.DocQualityFeatures.Contrast
	features[offset+3] = inputs.DocQualityFeatures.NoiseLevel
	features[offset+4] = inputs.DocQualityFeatures.BlurScore

	return features
}

// computeConfidence calculates confidence based on score distance from boundaries.
func computeConfidence(rawScore float32) float32 {
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

func absFloat32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
