// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0
//
// ML Runtime Stub - Fallback implementation when ML runtime is not available
// VE-205: ML inference integration in Cosmos module

//go:build !mlruntime
// +build !mlruntime

package inference

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// ML Runtime (Stub Implementation)
// ============================================================================

// TensorFlowRuntime is a stub implementation when the ML runtime is not
// available. It provides the same interface but uses a deterministic
// fallback scoring algorithm.
//
// This allows the codebase to compile and run without ML framework dependencies,
// which is useful for:
//   - Development and testing without ML dependencies
//   - Environments where the inference sidecar is not available
//   - CI/CD pipelines that don't need real inference
type TensorFlowRuntime struct {
	// Configuration
	config      TFRuntimeConfig
	determinism *DeterminismController

	// State
	modelPath  string
	modelHash  string
	mu         sync.RWMutex
	isInit     atomic.Bool
	isHealthy  atomic.Bool
	lastHealth atomic.Int64

	// Metrics
	inferenceCount atomic.Uint64
	errorCount     atomic.Uint64
	totalLatencyNs atomic.Int64

	// Logging
	logger *log.Logger
}

// TFRuntimeConfig holds configuration for the ML runtime
type TFRuntimeConfig struct {
	// ModelPath is the path to the TensorFlow SavedModel directory
	ModelPath string

	// ModelVersion is the expected model version
	ModelVersion string

	// ExpectedHash is the expected SHA256 hash of the model (optional)
	ExpectedHash string

	// InputTensorName is the name of the input tensor
	InputTensorName string

	// OutputTensorName is the name of the output tensor
	OutputTensorName string

	// ServingTag is the SavedModel tag for serving
	ServingTag string

	// RandomSeed is the fixed random seed for determinism
	RandomSeed int64

	// ForceCPU forces CPU-only execution
	ForceCPU bool

	// InterOpParallelism is the number of inter-op threads (1 for determinism)
	InterOpParallelism int

	// IntraOpParallelism is the number of intra-op threads (1 for determinism)
	IntraOpParallelism int

	// EnableDeterministicOps enables deterministic TensorFlow ops
	EnableDeterministicOps bool

	// LogLevel controls TensorFlow logging verbosity
	LogLevel int

	// HealthCheckInterval is the interval between health checks
	HealthCheckInterval time.Duration
}

// DefaultTFRuntimeConfig returns the default runtime configuration
// with deterministic settings for blockchain consensus
func DefaultTFRuntimeConfig() TFRuntimeConfig {
	return TFRuntimeConfig{
		ModelPath:              "",
		ModelVersion:           "v1.0.0",
		ExpectedHash:           "",
		InputTensorName:        "serving_default_features",
		OutputTensorName:       "StatefulPartitionedCall",
		ServingTag:             "serve",
		RandomSeed:             42,
		ForceCPU:               true,
		InterOpParallelism:     1,
		IntraOpParallelism:     1,
		EnableDeterministicOps: true,
		LogLevel:               2,
		HealthCheckInterval:    30 * time.Second,
	}
}

// ============================================================================
// Constructor
// ============================================================================

// NewTensorFlowRuntime creates a new TensorFlow runtime stub.
// This is the fallback implementation when TensorFlow-Go is not available.
func NewTensorFlowRuntime(config TFRuntimeConfig) *TensorFlowRuntime {
	return &TensorFlowRuntime{
		config:      config,
		determinism: NewDeterminismController(config.RandomSeed, config.ForceCPU),
		logger:      log.New(os.Stderr, "[TFRuntime-Stub] ", log.LstdFlags|log.Lmsgprefix),
	}
}

// ============================================================================
// Initialization
// ============================================================================

// Initialize prepares the stub runtime.
// In stub mode, this verifies the model path exists and computes its hash.
func (r *TensorFlowRuntime) Initialize() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isInit.Load() {
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

	// Compute model hash for verification
	modelHash, err := r.computeModelHash(r.config.ModelPath)
	if err != nil {
		return fmt.Errorf("failed to compute model hash: %w", err)
	}
	r.modelHash = modelHash
	r.modelPath = r.config.ModelPath

	// Verify hash if expected
	if r.config.ExpectedHash != "" {
		expectedHash := normalizeExpectedHash(r.config.ExpectedHash)
		if r.modelHash != expectedHash {
			return fmt.Errorf("model hash mismatch: expected %s, got %s", expectedHash, r.modelHash)
		}
		r.logger.Printf("Model hash verified: %s", r.modelHash[:16])
	}

	r.isInit.Store(true)
	r.isHealthy.Store(true)
	r.lastHealth.Store(time.Now().UnixNano())

	r.logger.Printf("TensorFlow runtime (STUB) initialized: model=%s, hash=%s",
		filepath.Base(r.config.ModelPath), r.modelHash[:16])
	r.logger.Printf("WARNING: Using stub implementation - build with 'tensorflow' tag for real inference")

	return nil
}

// ============================================================================
// Inference (Stub)
// ============================================================================

// Run executes stub inference on the given features.
// This provides deterministic output based on feature values,
// simulating what a real model would produce.
func (r *TensorFlowRuntime) Run(features []float32) ([]float32, error) {
	if !r.isInit.Load() {
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

	// Run stub inference
	output := r.stubInference(features)

	// Update metrics
	r.inferenceCount.Add(1)
	r.totalLatencyNs.Add(time.Since(startTime).Nanoseconds())
	r.isHealthy.Store(true)
	r.lastHealth.Store(time.Now().UnixNano())

	return output, nil
}

// stubInference computes a deterministic score based on feature values.
// This simulates the behavior of the real model for testing.
func (r *TensorFlowRuntime) stubInference(features []float32) []float32 {
	var sum float32
	var count float32

	// Weight face embedding (indices 0-511)
	for i := 0; i < FaceEmbeddingDim && i < len(features); i++ {
		sum += absFloat32(features[i]) * 0.5
		count++
	}

	// Weight document quality (indices 512-516)
	docOffset := FaceEmbeddingDim
	if docOffset+DocQualityDim <= len(features) {
		for i := 0; i < DocQualityDim; i++ {
			sum += features[docOffset+i] * 1.5
			count++
		}
	}

	// Weight OCR features (indices 517-526)
	ocrOffset := FaceEmbeddingDim + DocQualityDim
	if ocrOffset+OCRFieldsDim <= len(features) {
		for i := 0; i < OCRFieldsDim; i++ {
			sum += features[ocrOffset+i]
			count++
		}
	}

	// Normalize to 0-100 range
	var rawScore float32
	if count > 0 {
		rawScore = (sum / count) * 100
		if rawScore > 100 {
			rawScore = 100
		}
		if rawScore < 0 {
			rawScore = 0
		}
	}

	return []float32{rawScore}
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
		bits := make([]byte, 4)
		binary.BigEndian.PutUint32(bits, uint32(val))
		h.Write(bits)
	}

	return hex.EncodeToString(h.Sum(nil))
}

// ============================================================================
// Hot Reload
// ============================================================================

// Reload loads a new model from the specified path.
// In stub mode, this just updates the model path and hash.
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

	// Update state
	r.modelPath = modelPath
	r.modelHash = newHash
	r.isHealthy.Store(true)
	r.lastHealth.Store(time.Now().UnixNano())

	r.logger.Printf("Model reloaded successfully (stub): hash=%s", newHash[:16])

	return nil
}

// ============================================================================
// Health Check
// ============================================================================

// IsHealthy returns whether the runtime is healthy
func (r *TensorFlowRuntime) IsHealthy() bool {
	return r.isInit.Load() && r.isHealthy.Load()
}

// HealthCheck performs an active health check
func (r *TensorFlowRuntime) HealthCheck() error {
	if !r.isInit.Load() {
		return fmt.Errorf("runtime not initialized")
	}

	// Create test input
	testFeatures := make([]float32, TotalFeatureDim)
	for i := range testFeatures {
		testFeatures[i] = 0.5
	}

	// Run test inference
	_, err := r.Run(testFeatures)
	if err != nil {
		r.isHealthy.Store(false)
		return fmt.Errorf("health check inference failed: %w", err)
	}

	r.isHealthy.Store(true)
	r.lastHealth.Store(time.Now().UnixNano())

	return nil
}

// GetHealthStatus returns detailed health status
func (r *TensorFlowRuntime) GetHealthStatus() RuntimeHealthStatus {
	return RuntimeHealthStatus{
		IsHealthy:      r.isHealthy.Load(),
		IsInitialized:  r.isInit.Load(),
		ModelPath:      r.modelPath,
		ModelHash:      r.modelHash,
		InferenceCount: r.inferenceCount.Load(),
		ErrorCount:     r.errorCount.Load(),
		LastHealthTime: time.Unix(0, r.lastHealth.Load()),
		AvgLatencyNs:   r.getAverageLatency(),
	}
}

// RuntimeHealthStatus contains TensorFlow runtime health information.
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

// Close releases resources (no-op for stub)
func (r *TensorFlowRuntime) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.isHealthy.Store(false)
	r.isInit.Store(false)

	r.logger.Printf("TensorFlow runtime (stub) closed")

	return nil
}

// ============================================================================
// Model Hash Computation
// ============================================================================

// computeModelHash computes SHA256 hash of all model files
func (r *TensorFlowRuntime) computeModelHash(modelPath string) (string, error) {
	h := sha256.New()

	// Collect all files
	var files []string
	err := filepath.Walk(modelPath, func(path string, info os.FileInfo, err error) error {
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
