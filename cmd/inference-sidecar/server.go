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
	"github.com/virtengine/virtengine/pkg/security"
)

// InferenceSidecarServer implements the InferenceService gRPC server.
type InferenceSidecarServer struct {
	inferencepb.UnimplementedInferenceServiceServer

	// config holds inference configuration
	config inference.InferenceConfig

	// model holds loaded model metadata
	model *inference.TFModel

	// loader handles model loading
	loader *inference.ModelLoader

	// determinism ensures deterministic hashing
	determinism *inference.DeterminismController

	// extractor transforms inputs to features
	extractor *inference.FeatureExtractor

	// servingClient executes inference via TensorFlow Serving
	servingClient *inference.TFServingClient

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
func NewInferenceSidecarServer(config inference.InferenceConfig, servingConfig inference.TFServingConfig, log Logger) (*InferenceSidecarServer, error) {
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

	inputName := ""
	outputName := ""
	if metadata := model.GetMetadata(); metadata != nil {
		inputName = metadata.InputName
		outputName = metadata.OutputName
	}
	if servingConfig.InputName == "" {
		servingConfig.InputName = inputName
	}
	if servingConfig.OutputName == "" {
		servingConfig.OutputName = outputName
	}

	servingClient, err := inference.NewTFServingClient(servingConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to configure tf serving client: %w", err)
	}
	server.servingClient = servingClient

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

	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	result, err := s.scoreFeatures(ctx, req.Features, req.ReturnContributions)
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

	s.log.Debug("ComputeScore completed",
		"score", result.Score,
		"confidence", result.Confidence,
		"latency_ms", latencyMs,
	)

	result.ComputeTimeMs = latencyMs
	return &inferencepb.ComputeScoreResponse{
		Score:                result.Score,
		RawScore:             result.RawScore,
		Confidence:           result.Confidence,
		InputHash:            result.InputHash,
		OutputHash:           result.OutputHash,
		ModelVersion:         result.ModelVersion,
		ModelHash:            result.ModelHash,
		ReasonCodes:          result.ReasonCodes,
		ComputeTimeMs:        latencyMs,
		FeatureContributions: result.FeatureContributions,
	}, nil
}

// HealthCheck implements InferenceServiceServer.HealthCheck.
func (s *InferenceSidecarServer) HealthCheck(ctx context.Context, req *inferencepb.HealthCheckRequest) (*inferencepb.HealthCheckResponse, error) {
	s.log.Debug("HealthCheck called")

	status := inferencepb.HealthStatus_HEALTH_STATUS_HEALTHY
	errorMsg := ""

	if s.model == nil || !s.model.IsLoaded() {
		status = inferencepb.HealthStatus_HEALTH_STATUS_UNHEALTHY
		errorMsg = "model not loaded"
	} else if s.servingClient != nil {
		if _, err := s.servingClient.CheckHealth(ctx); err != nil {
			status = inferencepb.HealthStatus_HEALTH_STATUS_DEGRADED
			errorMsg = err.Error()
		}
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
	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	result, err := s.scoreFeatures(ctx, features, false)
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

func (s *InferenceSidecarServer) scoreFeatures(ctx context.Context, features []float32, includeContributions bool) (*inference.ScoreResult, error) {
	output, endpoint, err := s.runInference(ctx, features)
	if err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, fmt.Errorf("model returned empty output")
	}

	rawScore := output[0]
	score := security.SafeFloat32ToUint32(rawScore, 0, 100)
	confidence := computeConfidence(rawScore)
	reasonCodes := s.reasonCodesFromFeatures(features, score, confidence)

	result := &inference.ScoreResult{
		Score:        score,
		RawScore:     rawScore,
		Confidence:   confidence,
		InputHash:    s.determinism.ComputeFeatureHash(features),
		OutputHash:   s.determinism.ComputeOutputHash(output),
		ModelVersion: s.model.GetVersion(),
		ModelHash:    s.model.GetModelHash(),
		ReasonCodes:  reasonCodes,
	}

	if includeContributions {
		result.FeatureContributions = s.extractor.ComputeFeatureContributions(features)
	}

	if endpoint != "" {
		s.log.Debug("Inference backend", "endpoint", endpoint)
	}

	return result, nil
}

func (s *InferenceSidecarServer) runInference(ctx context.Context, features []float32) ([]float32, string, error) {
	if s.servingClient == nil {
		if s.model == nil {
			return nil, "", fmt.Errorf("no inference backend configured")
		}
		output, err := s.model.Run(features)
		return output, "local_stub", err
	}

	output, endpoint, _, err := s.servingClient.Predict(ctx, features)
	if err == nil {
		return output, endpoint, nil
	}

	if s.config.AllowFallbackToStub && s.model != nil {
		fallbackOutput, fallbackErr := s.model.Run(features)
		if fallbackErr == nil {
			return fallbackOutput, "local_stub", nil
		}
	}

	return nil, endpoint, err
}

func (s *InferenceSidecarServer) reasonCodesFromFeatures(features []float32, score uint32, confidence float32) []string {
	reasons := make([]string, 0, 4)

	if score >= 50 {
		reasons = append(reasons, inference.ReasonCodeSuccess)
	}
	if confidence >= 0.8 {
		reasons = append(reasons, inference.ReasonCodeHighConfidence)
	} else if confidence < 0.5 {
		reasons = append(reasons, inference.ReasonCodeLowConfidence)
	}

	docQualityScore := features[inference.FaceEmbeddingDim]
	if docQualityScore < 0.6 {
		reasons = append(reasons, inference.ReasonCodeLowDocQuality)
	}

	ocrOffset := inference.FaceEmbeddingDim + inference.DocQualityDim
	var ocrSum float32
	var ocrCount int
	for i := 0; i < inference.OCRFieldsDim; i += 2 {
		if ocrOffset+i >= len(features) {
			break
		}
		ocrSum += features[ocrOffset+i]
		ocrCount++
	}
	if ocrCount > 0 && (ocrSum/float32(ocrCount)) < 0.5 {
		reasons = append(reasons, inference.ReasonCodeLowOCRConfidence)
	}

	metaOffset := inference.FaceEmbeddingDim + inference.DocQualityDim + inference.OCRFieldsDim
	if metaOffset < len(features) {
		scopeCount := int(features[metaOffset] * 10.0)
		if scopeCount < 2 {
			reasons = append(reasons, inference.ReasonCodeInsufficientScopes)
		}
	}

	return reasons
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
