// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0
//
// ML Runtime - Real ML inference implementation using sidecar pattern
// VE-205: ML inference integration in Cosmos module
//
// This implementation uses a sidecar process for inference, avoiding
// the need to include ML framework bindings directly in Go. The sidecar
// handles TensorFlow/ONNX execution while this code manages communication.
//
// For production deployment, operators should run the inference sidecar
// as a separate process or container.

//go:build mlruntime
// +build mlruntime

package inference

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// ML Runtime (Sidecar Pattern)
// ============================================================================

// TensorFlowRuntime manages ML model inference using a sidecar pattern.
// This implementation communicates with an external inference server
// (TensorFlow Serving, ONNX Runtime Server, or custom inference service)
// via HTTP, avoiding the need for direct ML framework bindings in Go.
//
// The name is kept as TensorFlowRuntime for API compatibility.
//
// Key features:
//   - Deterministic execution (CPU-only, single-threaded, fixed seed)
//   - Input/output hash computation for consensus verification
//   - Model hot-reload without restart
//   - Health check and metrics
type TensorFlowRuntime struct {
	// Sidecar connection
	sidecarURL string
	httpClient *http.Client

	// Model state
	modelPath string
	modelHash string

	// Configuration
	config      TFRuntimeConfig
	determinism *DeterminismController

	// State management
	mu             sync.RWMutex
	isInitialized  atomic.Bool
	isHealthy      atomic.Bool
	lastHealthTime atomic.Int64

	// Metrics
	inferenceCount atomic.Uint64
	errorCount     atomic.Uint64
	totalLatencyNs atomic.Int64

	// Logging
	logger *log.Logger
}

// SidecarRequest is the inference request format for the sidecar
type SidecarRequest struct {
	Features []float32 `json:"features"`
	ModelID  string    `json:"model_id,omitempty"`
}

// SidecarResponse is the inference response format from the sidecar
type SidecarResponse struct {
	Scores    []float32 `json:"scores"`
	InputHash string    `json:"input_hash"`
	ModelHash string    `json:"model_hash,omitempty"`
	Latency   float64   `json:"latency_ms,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// TFRuntimeConfig holds configuration for the ML runtime
type TFRuntimeConfig struct {
	// ModelPath is the path to the ML model file/directory
	ModelPath string

	// ModelVersion is the expected model version
	ModelVersion string

	// ExpectedHash is the expected SHA256 hash of the model (optional)
	ExpectedHash string

	// InputTensorName is the name of the input tensor
	InputTensorName string

	// OutputTensorName is the name of the output tensor
	OutputTensorName string

	// ServingTag is the model tag for serving
	ServingTag string

	// RandomSeed is the fixed random seed for determinism
	RandomSeed int64

	// ForceCPU forces CPU-only execution
	ForceCPU bool

	// InterOpParallelism is the number of inter-op threads (1 for determinism)
	InterOpParallelism int

	// IntraOpParallelism is the number of intra-op threads (1 for determinism)
	IntraOpParallelism int

	// EnableDeterministicOps enables deterministic ops
	EnableDeterministicOps bool

	// LogLevel controls logging verbosity
	LogLevel int

	// HealthCheckInterval is the interval between health checks
	HealthCheckInterval time.Duration

	// SidecarURL is the URL of the inference sidecar service
	SidecarURL string

	// SidecarTimeout is the timeout for sidecar requests
	SidecarTimeout time.Duration
}

// DefaultTFRuntimeConfig returns the default runtime configuration
// with deterministic settings for blockchain consensus
func DefaultTFRuntimeConfig() TFRuntimeConfig {
	return TFRuntimeConfig{
		ModelPath:              "",
		ModelVersion:           "v1.0.0",
		ExpectedHash:           "",
		InputTensorName:        "features",
		OutputTensorName:       "trust_score",
		ServingTag:             "serve",
		RandomSeed:             42,
		ForceCPU:               true,
		InterOpParallelism:     1,
		IntraOpParallelism:     1,
		EnableDeterministicOps: true,
		LogLevel:               2,
		HealthCheckInterval:    30 * time.Second,
		SidecarURL:             "http://localhost:8501",
		SidecarTimeout:         10 * time.Second,
	}
}

// ============================================================================
// Constructor
// ============================================================================

// NewTensorFlowRuntime creates a new ML runtime with the given config.
// The runtime is not initialized until Initialize() is called.
func NewTensorFlowRuntime(config TFRuntimeConfig) *TensorFlowRuntime {
	return &TensorFlowRuntime{
		config:      config,
		determinism: NewDeterminismController(config.RandomSeed, config.ForceCPU),
		logger:      log.New(os.Stderr, "[MLRuntime] ", log.LstdFlags|log.Lmsgprefix),
	}
}

// ============================================================================
// Initialization
// ============================================================================

// Initialize sets up the ML runtime and connects to the sidecar.
func (r *TensorFlowRuntime) Initialize() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isInitialized.Load() {
		return nil
	}

	// Validate configuration
	if r.config.ModelPath == "" {
		return fmt.Errorf("model path is required")
	}

	// Verify model path exists
	if _, err := os.Stat(r.config.ModelPath); os.IsNotExist(err) {
		return fmt.Errorf("model path does not exist: %s", r.config.ModelPath)
	}

	// Configure environment for determinism
	if err := r.configureEnvironment(); err != nil {
		return fmt.Errorf("failed to configure environment: %w", err)
	}

	// Compute model hash for verification
	modelHash, err := r.computeModelHash(r.config.ModelPath)
	if err != nil {
		return fmt.Errorf("failed to compute model hash: %w", err)
	}
	r.modelHash = modelHash

	// Verify hash if expected
	if r.config.ExpectedHash != "" {
		expectedHash := normalizeExpectedHash(r.config.ExpectedHash)
		if r.modelHash != expectedHash {
			return fmt.Errorf("model hash mismatch: expected %s, got %s", expectedHash, r.modelHash)
		}
		r.logger.Printf("Model hash verified: %s", r.modelHash[:16])
	}

	// Set up sidecar URL
	r.sidecarURL = r.config.SidecarURL
	if r.sidecarURL == "" {
		r.sidecarURL = os.Getenv("VEID_INFERENCE_SIDECAR_URL")
	}
	if r.sidecarURL == "" {
		r.sidecarURL = "http://localhost:8501"
	}

	// Create HTTP client
	timeout := r.config.SidecarTimeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	r.httpClient = &http.Client{
		Timeout: timeout,
	}

	// Verify sidecar is healthy
	if err := r.checkSidecarHealth(); err != nil {
		r.logger.Printf("Warning: sidecar health check failed: %v", err)
		// Don't fail initialization - sidecar may start later
	}

	r.modelPath = r.config.ModelPath
	r.isInitialized.Store(true)
	r.isHealthy.Store(true)
	r.lastHealthTime.Store(time.Now().UnixNano())

	r.logger.Printf("ML runtime initialized: model=%s, hash=%s, sidecar=%s",
		filepath.Base(r.config.ModelPath), r.modelHash[:16], r.sidecarURL)

	return nil
}

// configureEnvironment sets environment variables for deterministic inference
func (r *TensorFlowRuntime) configureEnvironment() error {
	envVars := r.determinism.GetTensorFlowEnvVars()

	// Set environment variables for determinism
	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}

	return nil
}

// checkSidecarHealth checks if the sidecar is healthy
func (r *TensorFlowRuntime) checkSidecarHealth() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", r.sidecarURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sidecar returned status %d", resp.StatusCode)
	}

	return nil
}

// ============================================================================
// Inference
// ============================================================================

// Run executes inference on the given features and returns the output.
// Input features must match the model's expected dimension (typically 768).
func (r *TensorFlowRuntime) Run(features []float32) ([]float32, error) {
	if !r.isInitialized.Load() {
		return nil, fmt.Errorf("runtime not initialized")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	startTime := time.Now()

	// Validate input dimension
	if len(features) != TotalFeatureDim {
		r.errorCount.Add(1)
		return nil, fmt.Errorf("feature dimension mismatch: expected %d, got %d",
			TotalFeatureDim, len(features))
	}

	// Create request
	reqBody := SidecarRequest{
		Features: features,
		ModelID:  r.config.ModelVersion,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		r.errorCount.Add(1)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request to sidecar
	ctx, cancel := context.WithTimeout(context.Background(), r.config.SidecarTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", r.sidecarURL+"/v1/models/trust_score:predict", bytes.NewReader(jsonData))
	if err != nil {
		r.errorCount.Add(1)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		r.errorCount.Add(1)
		r.isHealthy.Store(false)
		return nil, fmt.Errorf("sidecar request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.errorCount.Add(1)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		r.errorCount.Add(1)
		r.isHealthy.Store(false)
		return nil, fmt.Errorf("sidecar returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var sidecarResp SidecarResponse
	if err := json.Unmarshal(body, &sidecarResp); err != nil {
		r.errorCount.Add(1)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if sidecarResp.Error != "" {
		r.errorCount.Add(1)
		return nil, fmt.Errorf("sidecar error: %s", sidecarResp.Error)
	}

	if len(sidecarResp.Scores) == 0 {
		r.errorCount.Add(1)
		return nil, fmt.Errorf("empty inference result")
	}

	// Update metrics
	r.inferenceCount.Add(1)
	r.totalLatencyNs.Add(time.Since(startTime).Nanoseconds())
	r.isHealthy.Store(true)
	r.lastHealthTime.Store(time.Now().UnixNano())

	return sidecarResp.Scores, nil
}

// RunWithHashes executes inference and returns output with consensus hashes
func (r *TensorFlowRuntime) RunWithHashes(features []float32) (output []float32, inputHash, outputHash string, err error) {
	// Compute input hash
	inputHash = r.computeFeatureHash(features)

	// Run inference
	output, err = r.Run(features)
	if err != nil {
		return nil, inputHash, "", err
	}

	// Compute output hash
	outputHash = r.determinism.ComputeOutputHash(output)

	return output, inputHash, outputHash, nil
}

// computeFeatureHash computes SHA256 hash of feature vector
func (r *TensorFlowRuntime) computeFeatureHash(features []float32) string {
	h := sha256.New()

	for _, val := range features {
		// Round to 6 decimal places for determinism
		rounded := float32(math.Round(float64(val)*1e6) / 1e6)
		bits := make([]byte, 4)
		binary.BigEndian.PutUint32(bits, math.Float32bits(rounded))
		h.Write(bits)
	}

	return hex.EncodeToString(h.Sum(nil))
}

// ============================================================================
// Hot Reload
// ============================================================================

// Reload loads a new model from the specified path, replacing the current model.
// This triggers a reload on the sidecar.
func (r *TensorFlowRuntime) Reload(modelPath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if modelPath == "" {
		return fmt.Errorf("model path is required")
	}

	// Verify new model path exists
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return fmt.Errorf("new model path does not exist: %s", modelPath)
	}

	r.logger.Printf("Reloading model from: %s", modelPath)

	// Compute hash of new model
	newHash, err := r.computeModelHash(modelPath)
	if err != nil {
		return fmt.Errorf("failed to compute new model hash: %w", err)
	}

	// Skip reload if same model
	if newHash == r.modelHash {
		r.logger.Printf("Model hash unchanged, skipping reload")
		return nil
	}

	// Verify hash if expected
	if r.config.ExpectedHash != "" {
		expectedHash := normalizeExpectedHash(r.config.ExpectedHash)
		if newHash != expectedHash {
			return fmt.Errorf("new model hash mismatch: expected %s, got %s", expectedHash, newHash)
		}
	}

	// Notify sidecar to reload model
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reloadReq := map[string]string{
		"model_path": modelPath,
		"model_hash": newHash,
	}
	jsonData, err := json.Marshal(reloadReq)
	if err != nil {
		return fmt.Errorf("failed to marshal reload request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.sidecarURL+"/v1/models/reload", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create reload request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sidecar reload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sidecar reload failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Update runtime state
	r.modelPath = modelPath
	r.modelHash = newHash

	r.isHealthy.Store(true)
	r.lastHealthTime.Store(time.Now().UnixNano())

	r.logger.Printf("Model reloaded successfully: hash=%s", newHash[:16])

	return nil
}

// ============================================================================
// Health Check
// ============================================================================

// IsHealthy returns whether the runtime is healthy and ready for inference
func (r *TensorFlowRuntime) IsHealthy() bool {
	return r.isInitialized.Load() && r.isHealthy.Load()
}

// HealthCheck performs an active health check by running a test inference
func (r *TensorFlowRuntime) HealthCheck() error {
	if !r.isInitialized.Load() {
		return fmt.Errorf("runtime not initialized")
	}

	// Check sidecar health
	if err := r.checkSidecarHealth(); err != nil {
		r.isHealthy.Store(false)
		return fmt.Errorf("sidecar health check failed: %w", err)
	}

	// Create test input
	testFeatures := make([]float32, TotalFeatureDim)
	for i := range testFeatures {
		testFeatures[i] = 0.5 // Neutral test values
	}

	// Run test inference
	_, err := r.Run(testFeatures)
	if err != nil {
		r.isHealthy.Store(false)
		return fmt.Errorf("health check inference failed: %w", err)
	}

	r.isHealthy.Store(true)
	r.lastHealthTime.Store(time.Now().UnixNano())

	return nil
}

// GetHealthStatus returns detailed health status
func (r *TensorFlowRuntime) GetHealthStatus() RuntimeHealthStatus {
	return RuntimeHealthStatus{
		IsHealthy:      r.isHealthy.Load(),
		IsInitialized:  r.isInitialized.Load(),
		ModelPath:      r.modelPath,
		ModelHash:      r.modelHash,
		InferenceCount: r.inferenceCount.Load(),
		ErrorCount:     r.errorCount.Load(),
		LastHealthTime: time.Unix(0, r.lastHealthTime.Load()),
		AvgLatencyNs:   r.getAverageLatency(),
	}
}

// RuntimeHealthStatus contains ML runtime health information.
// This is distinct from the package-level HealthStatus for the inference system.
type RuntimeHealthStatus struct {
	IsHealthy      bool
	IsInitialized  bool
	ModelPath      string
	ModelHash      string
	InferenceCount uint64
	ErrorCount     uint64
	LastHealthTime time.Time
	AvgLatencyNs   int64
}

// getAverageLatency calculates average inference latency
func (r *TensorFlowRuntime) getAverageLatency() int64 {
	count := r.inferenceCount.Load()
	if count == 0 {
		return 0
	}
	return r.totalLatencyNs.Load() / int64(count)
}

// ============================================================================
// Cleanup
// ============================================================================

// Close releases all runtime resources
func (r *TensorFlowRuntime) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.isHealthy.Store(false)
	r.isInitialized.Store(false)

	if r.httpClient != nil {
		r.httpClient.CloseIdleConnections()
		r.httpClient = nil
	}

	r.logger.Printf("ML runtime closed")

	return nil
}

// ============================================================================
// Model Hash Computation
// ============================================================================

// computeModelHash computes SHA256 hash of all model files
func (r *TensorFlowRuntime) computeModelHash(modelPath string) (string, error) {
	h := sha256.New()

	// Check if it's a file or directory
	info, err := os.Stat(modelPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat model path: %w", err)
	}

	if !info.IsDir() {
		// Single file - hash it directly
		//nolint:gosec // G304: path is from trusted model directory
		data, err := os.ReadFile(modelPath)
		if err != nil {
			return "", fmt.Errorf("failed to read model file: %w", err)
		}
		h.Write(data)
		return hex.EncodeToString(h.Sum(nil)), nil
	}

	// Directory - collect and hash all files
	var files []string
	err = filepath.Walk(modelPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// Skip metadata files
		if filepath.Base(path) == "export_metadata.json" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to walk model directory: %w", err)
	}

	// Sort for determinism
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i] > files[j] {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Hash each file
	for _, path := range files {
		//nolint:gosec // G304: path is from trusted model directory
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read %s: %w", path, err)
		}
		h.Write(data)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// ============================================================================
// Getters
// ============================================================================

// GetModelHash returns the SHA256 hash of the loaded model
func (r *TensorFlowRuntime) GetModelHash() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.modelHash
}

// GetModelPath returns the path to the loaded model
func (r *TensorFlowRuntime) GetModelPath() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.modelPath
}

// GetConfig returns the runtime configuration
func (r *TensorFlowRuntime) GetConfig() TFRuntimeConfig {
	return r.config
}

// GetInferenceCount returns the total number of inferences
func (r *TensorFlowRuntime) GetInferenceCount() uint64 {
	return r.inferenceCount.Load()
}

// GetErrorCount returns the total number of errors
func (r *TensorFlowRuntime) GetErrorCount() uint64 {
	return r.errorCount.Load()
}

