package inference

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	inferencepb "github.com/virtengine/virtengine/pkg/inference/proto"
)

// ============================================================================
// Sidecar Client
// ============================================================================

// SidecarClient implements the Scorer interface by calling an external
// gRPC inference sidecar service. This provides an alternative to embedded
// TensorFlow when:
// - TensorFlow C library is not available
// - Memory isolation is desired
// - Different inference hardware (GPU) is needed
//
// The sidecar must implement deterministic inference and return the same
// hashes for the same inputs across all validators.
type SidecarClient struct {
	// config holds the inference configuration
	config InferenceConfig

	// extractor transforms inputs to features
	extractor *FeatureExtractor

	// determinism ensures deterministic hashing
	determinism *DeterminismController

	// conn is the gRPC connection
	conn *grpc.ClientConn

	// client is the gRPC service client
	grpcClient inferencepb.InferenceServiceClient

	// mu protects client state
	mu sync.RWMutex

	// isConnected indicates if connected to sidecar
	isConnected bool

	// modelVersion cached from sidecar
	modelVersion string

	// modelHash cached from sidecar
	modelHash string

	// inferenceCount tracks total inferences
	inferenceCount uint64

	// errorCount tracks inference errors
	errorCount uint64

	// lastHealthCheck tracks last health check time
	lastHealthCheck time.Time

	// useTLS indicates if TLS should be used
	useTLS bool
}

// NewSidecarClient creates a new sidecar client
func NewSidecarClient(config InferenceConfig) (*SidecarClient, error) {
	if !config.UseSidecar {
		return nil, fmt.Errorf("sidecar mode not enabled in config")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	client := &SidecarClient{
		config:      config,
		extractor:   NewFeatureExtractor(DefaultFeatureExtractorConfig()),
		determinism: NewDeterminismController(config.RandomSeed, config.ForceCPU),
		isConnected: false,
	}

	// Connect to sidecar
	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to sidecar: %w", err)
	}

	return client, nil
}

// ============================================================================
// Connection Management
// ============================================================================

// connect establishes connection to the inference sidecar
func (sc *SidecarClient) connect() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), sc.config.SidecarTimeout)
	defer cancel()

	// Build gRPC dial options
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock())

	if sc.useTLS {
		// Use TLS with system root CAs
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		// Use insecure connection for local development
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.DialContext(ctx, sc.config.SidecarAddress, opts...)
	if err != nil {
		return fmt.Errorf("failed to dial sidecar at %s: %w", sc.config.SidecarAddress, err)
	}

	sc.conn = conn
	sc.grpcClient = inferencepb.NewInferenceServiceClient(conn)
	sc.isConnected = true

	// Get model info from sidecar
	if err := sc.refreshModelInfo(); err != nil {
		return fmt.Errorf("failed to get model info: %w", err)
	}

	return nil
}

// refreshModelInfo fetches model version and hash from sidecar
func (sc *SidecarClient) refreshModelInfo() error {
	if sc.grpcClient == nil {
		// Fallback for testing when no real connection
		sc.modelVersion = sc.config.ModelVersion
		sc.modelHash = sc.config.ExpectedHash
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), sc.config.SidecarTimeout)
	defer cancel()

	resp, err := sc.grpcClient.GetModelInfo(ctx, &inferencepb.GetModelInfoRequest{})
	if err != nil {
		return fmt.Errorf("GetModelInfo RPC failed: %w", err)
	}

	sc.modelVersion = resp.Version
	sc.modelHash = resp.Hash
	sc.lastHealthCheck = time.Now()

	// Verify model hash matches expected if configured
	if sc.config.ExpectedHash != "" && sc.modelHash != sc.config.ExpectedHash {
		return fmt.Errorf("model hash mismatch: expected %s, got %s",
			sc.config.ExpectedHash, sc.modelHash)
	}

	return nil
}

// reconnect attempts to reconnect to the sidecar
func (sc *SidecarClient) reconnect() error {
	sc.mu.Lock()
	if sc.conn != nil {
		_ = sc.conn.Close()
		sc.conn = nil
	}
	sc.grpcClient = nil
	sc.isConnected = false
	sc.mu.Unlock()

	return sc.connect()
}

// ============================================================================
// Scorer Interface Implementation
// ============================================================================

// ComputeScore runs inference via the sidecar
func (sc *SidecarClient) ComputeScore(inputs *ScoreInputs) (*ScoreResult, error) {
	return sc.ComputeScoreWithContext(context.Background(), inputs)
}

// ComputeScoreWithContext runs inference with context support
func (sc *SidecarClient) ComputeScoreWithContext(ctx context.Context, inputs *ScoreInputs) (*ScoreResult, error) {
	startTime := time.Now()

	sc.mu.Lock()
	sc.inferenceCount++
	sc.mu.Unlock()

	result := &ScoreResult{
		Score:         0,
		Confidence:    0.0,
		ModelVersion:  sc.GetModelVersion(),
		ModelHash:     sc.GetModelHash(),
		ReasonCodes:   make([]string, 0),
		ComputeTimeMs: 0,
	}

	// Check connection
	if !sc.IsHealthy() {
		// Try to reconnect
		if err := sc.reconnect(); err != nil {
			result.ComputeTimeMs = time.Since(startTime).Milliseconds()
			result.ReasonCodes = append(result.ReasonCodes, ReasonCodeInferenceError)
			return result, fmt.Errorf("sidecar not connected: %w", err)
		}
	}

	// Compute input hash locally for verification
	result.InputHash = sc.determinism.ComputeInputHash(inputs)

	// Extract features
	features, err := sc.extractor.ExtractFeatures(inputs)
	if err != nil {
		result.ComputeTimeMs = time.Since(startTime).Milliseconds()
		result.ReasonCodes = append(result.ReasonCodes, ReasonCodeInferenceError)
		return result, fmt.Errorf("feature extraction failed: %w", err)
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, sc.config.SidecarTimeout)
	defer cancel()

	// Call sidecar
	sidecarResult, err := sc.callSidecar(ctx, features, inputs)
	if err != nil {
		sc.mu.Lock()
		sc.errorCount++
		sc.mu.Unlock()

		result.ComputeTimeMs = time.Since(startTime).Milliseconds()

		// Check if timeout
		if ctx.Err() == context.DeadlineExceeded {
			result.ReasonCodes = append(result.ReasonCodes, ReasonCodeTimeout)
		} else {
			result.ReasonCodes = append(result.ReasonCodes, ReasonCodeInferenceError)
		}

		if sc.config.UseFallbackOnError {
			result.Score = sc.config.FallbackScore
			return result, nil
		}
		return result, err
	}

	// Copy sidecar result
	result.Score = sidecarResult.Score
	result.RawScore = sidecarResult.RawScore
	result.Confidence = sidecarResult.Confidence
	result.OutputHash = sidecarResult.OutputHash
	result.ReasonCodes = sidecarResult.ReasonCodes
	result.FeatureContributions = sidecarResult.FeatureContributions
	result.ComputeTimeMs = time.Since(startTime).Milliseconds()

	return result, nil
}

// callSidecar makes the actual gRPC call to the inference sidecar
func (sc *SidecarClient) callSidecar(ctx context.Context, features []float32, inputs *ScoreInputs) (*ScoreResult, error) {
	// If no real gRPC client, fall back to simulated response
	if sc.grpcClient == nil {
		return sc.simulateSidecarResponse(features, inputs)
	}

	// Build gRPC request
	req := &inferencepb.ComputeScoreRequest{
		Features: features,
		Metadata: &inferencepb.InferenceMetadata{
			AccountAddress:   inputs.Metadata.AccountAddress,
			BlockHeight:      inputs.Metadata.BlockHeight,
			RequestID:        inputs.Metadata.RequestID,
			ValidatorAddress: inputs.Metadata.ValidatorAddress,
		},
		ReturnContributions: true,
	}

	// Call sidecar
	resp, err := sc.grpcClient.ComputeScore(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("sidecar ComputeScore RPC failed: %w", err)
	}

	// Verify output hash if we computed one locally
	localOutputHash := sc.determinism.ComputeOutputHash([]float32{resp.RawScore})
	if resp.OutputHash != "" && localOutputHash != resp.OutputHash {
		return nil, fmt.Errorf("output hash mismatch: local=%s, remote=%s",
			localOutputHash, resp.OutputHash)
	}

	// Convert response to ScoreResult
	result := &ScoreResult{
		Score:                resp.Score,
		RawScore:             resp.RawScore,
		Confidence:           resp.Confidence,
		InputHash:            resp.InputHash,
		OutputHash:           resp.OutputHash,
		ModelVersion:         resp.ModelVersion,
		ModelHash:            resp.ModelHash,
		ReasonCodes:          resp.ReasonCodes,
		ComputeTimeMs:        resp.ComputeTimeMs,
		FeatureContributions: resp.FeatureContributions,
	}

	return result, nil
}

// simulateSidecarResponse simulates sidecar response for testing
func (sc *SidecarClient) simulateSidecarResponse(features []float32, _ *ScoreInputs) (*ScoreResult, error) {
	// Compute a deterministic score based on features
	var sum float32
	var count float32

	for i := 0; i < len(features) && i < TotalFeatureDim; i++ {
		sum += absFloat32(features[i])
		count++
	}

	rawScore := float32(0.0)
	if count > 0 {
		rawScore = (sum / count) * 100
		if rawScore > 100 {
			rawScore = 100
		}
	}

	score := uint32(rawScore)

	result := &ScoreResult{
		Score:                score,
		RawScore:             rawScore,
		Confidence:           computeConfidence(rawScore),
		OutputHash:           sc.determinism.ComputeOutputHash([]float32{rawScore}),
		ReasonCodes:          make([]string, 0),
		FeatureContributions: sc.extractor.ComputeFeatureContributions(features),
	}

	// Add reason codes
	if score >= 50 {
		result.ReasonCodes = append(result.ReasonCodes, ReasonCodeSuccess)
	}
	if result.Confidence >= 0.8 {
		result.ReasonCodes = append(result.ReasonCodes, ReasonCodeHighConfidence)
	}

	return result, nil
}

// GetModelVersion returns the model version from sidecar
func (sc *SidecarClient) GetModelVersion() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.modelVersion
}

// GetModelHash returns the model hash from sidecar
func (sc *SidecarClient) GetModelHash() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.modelHash
}

// IsHealthy checks if connected to sidecar
func (sc *SidecarClient) IsHealthy() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.isConnected
}

// Close closes the sidecar connection
func (sc *SidecarClient) Close() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.isConnected = false

	if sc.conn != nil {
		err := sc.conn.Close()
		sc.conn = nil
		sc.grpcClient = nil
		return err
	}

	return nil
}

// ============================================================================
// Health Check
// ============================================================================

// PerformHealthCheck performs a health check against the sidecar
func (sc *SidecarClient) PerformHealthCheck(ctx context.Context) (*SidecarHealthStatus, error) {
	if sc.grpcClient == nil {
		return &SidecarHealthStatus{
			Healthy:      sc.isConnected,
			ModelLoaded:  true,
			ModelVersion: sc.modelVersion,
			ModelHash:    sc.modelHash,
		}, nil
	}

	resp, err := sc.grpcClient.HealthCheck(ctx, &inferencepb.HealthCheckRequest{})
	if err != nil {
		return nil, fmt.Errorf("health check RPC failed: %w", err)
	}

	return &SidecarHealthStatus{
		Healthy:                resp.Status == inferencepb.HealthStatus_HEALTH_STATUS_HEALTHY,
		Degraded:               resp.Status == inferencepb.HealthStatus_HEALTH_STATUS_DEGRADED,
		ModelLoaded:            resp.ModelLoaded,
		ModelVersion:           resp.ModelVersion,
		ModelHash:              resp.ModelHash,
		UptimeSeconds:          resp.UptimeSeconds,
		LastInferenceTimestamp: resp.LastInferenceTimestamp,
		ErrorMessage:           resp.ErrorMessage,
	}, nil
}

// SidecarHealthStatus represents the health status of the sidecar.
type SidecarHealthStatus struct {
	Healthy                bool
	Degraded               bool
	ModelLoaded            bool
	ModelVersion           string
	ModelHash              string
	UptimeSeconds          int64
	LastInferenceTimestamp string
	ErrorMessage           string
}

// GetMetrics fetches metrics from the sidecar
func (sc *SidecarClient) GetMetrics(ctx context.Context) (*SidecarMetrics, error) {
	if sc.grpcClient == nil {
		return &SidecarMetrics{
			TotalInferences:      sc.inferenceCount,
			SuccessfulInferences: sc.inferenceCount - sc.errorCount,
			FailedInferences:     sc.errorCount,
			ModelVersion:         sc.modelVersion,
			ModelHash:            sc.modelHash,
		}, nil
	}

	resp, err := sc.grpcClient.GetMetrics(ctx, &inferencepb.GetMetricsRequest{})
	if err != nil {
		return nil, fmt.Errorf("get metrics RPC failed: %w", err)
	}

	return &SidecarMetrics{
		TotalInferences:      resp.TotalInferences,
		SuccessfulInferences: resp.SuccessfulInferences,
		FailedInferences:     resp.FailedInferences,
		AverageLatencyMs:     resp.AverageLatencyMs,
		P99LatencyMs:         resp.P99LatencyMs,
		ModelVersion:         resp.ModelVersion,
		ModelHash:            resp.ModelHash,
		UptimeSeconds:        resp.UptimeSeconds,
		MemoryUsageMB:        resp.MemoryUsageMB,
		LatencyHistogram:     resp.LatencyHistogram,
	}, nil
}

// SidecarMetrics represents metrics from the sidecar.
type SidecarMetrics struct {
	TotalInferences      uint64
	SuccessfulInferences uint64
	FailedInferences     uint64
	AverageLatencyMs     float32
	P99LatencyMs         float32
	ModelVersion         string
	ModelHash            string
	UptimeSeconds        int64
	MemoryUsageMB        float32
	LatencyHistogram     map[string]uint64
}

// VerifyDeterminism runs a determinism verification check
func (sc *SidecarClient) VerifyDeterminism(ctx context.Context, testVectorID string) (*DeterminismResult, error) {
	if sc.grpcClient == nil {
		return &DeterminismResult{
			Passed:       true,
			TestVectorID: testVectorID,
		}, nil
	}

	resp, err := sc.grpcClient.VerifyDeterminism(ctx, &inferencepb.VerifyDeterminismRequest{
		TestVectorID: testVectorID,
	})
	if err != nil {
		return nil, fmt.Errorf("verify determinism RPC failed: %w", err)
	}

	return &DeterminismResult{
		Passed:             resp.Passed,
		ActualOutputHash:   resp.ActualOutputHash,
		ExpectedOutputHash: resp.ExpectedOutputHash,
		Differences:        resp.Differences,
		TestVectorID:       resp.TestVectorID,
	}, nil
}

// DeterminismResult represents the result of a determinism check.
type DeterminismResult struct {
	Passed             bool
	ActualOutputHash   string
	ExpectedOutputHash string
	Differences        []string
	TestVectorID       string
}

// ============================================================================
// Sidecar Protocol Buffer Definitions
// ============================================================================

// Note: These would typically be generated from a .proto file
// For now, we define the expected message types

// InferenceRequest is the request message for the inference sidecar
type InferenceRequest struct {
	// Features is the feature vector for inference
	Features []float32

	// Metadata contains contextual information
	Metadata *RequestMetadata
}

// RequestMetadata contains request metadata
type RequestMetadata struct {
	AccountAddress string
	BlockHeight    int64
	RequestID      string
}

// InferenceResponse is the response message from the inference sidecar
type InferenceResponse struct {
	// Score is the quantized score (0-100)
	Score uint32

	// RawScore is the raw model output
	RawScore float32

	// Confidence is the prediction confidence
	Confidence float32

	// OutputHash is the hash of raw outputs
	OutputHash string

	// ReasonCodes explain the score
	ReasonCodes []string

	// ComputeTimeMs is inference time in milliseconds
	ComputeTimeMs int64
}

// ModelInfoResponse is the response from GetModelInfo
type ModelInfoResponse struct {
	// Version is the model version
	Version string

	// Hash is the model hash
	Hash string

	// InputDim is the expected input dimension
	InputDim int

	// TensorFlowVersion is the TF version
	TensorFlowVersion string
}

