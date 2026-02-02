// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0
//
// Package inference provides comprehensive health checks for the ML inference system.
// VE-219: Deterministic identity verification runtime

package inference

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ============================================================================
// Prometheus Metrics for Health Checks
// ============================================================================

var (
	// healthCheckTotal counts total health check invocations
	healthCheckTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "veid",
			Subsystem: "inference",
			Name:      "health_check_total",
			Help:      "Total number of health check invocations",
		},
		[]string{"result"}, // "success", "failure"
	)

	// healthCheckFailuresTotal counts health check failures
	healthCheckFailuresTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "veid",
			Subsystem: "inference",
			Name:      "health_check_failures_total",
			Help:      "Total number of health check failures",
		},
	)

	// determinismVerifiedGauge tracks determinism verification status
	determinismVerifiedGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "veid",
			Subsystem: "inference",
			Name:      "determinism_verified",
			Help:      "Whether determinism has been verified (1=verified, 0=not verified or failed)",
		},
	)

	// healthCheckLatency tracks health check latency
	healthCheckLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "veid",
			Subsystem: "inference",
			Name:      "health_check_latency_seconds",
			Help:      "Health check latency in seconds",
			Buckets:   []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5},
		},
	)

	// lastHealthCheckTimestamp tracks the timestamp of the last health check
	lastHealthCheckTimestamp = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "veid",
			Subsystem: "inference",
			Name:      "last_health_check_timestamp_seconds",
			Help:      "Unix timestamp of the last health check",
		},
	)
)

// ============================================================================
// Health Configuration
// ============================================================================

// HealthConfig holds configuration for health checks
type HealthConfig struct {
	// CheckInterval is the interval between periodic health checks
	CheckInterval time.Duration

	// CheckTimeout is the timeout for individual health checks
	CheckTimeout time.Duration

	// DeterminismTestEnabled enables determinism verification during health checks
	DeterminismTestEnabled bool

	// ExpectedTestVectorHash is the expected hash of the test vector output
	// If empty, the first run will set this value
	ExpectedTestVectorHash string

	// TestVectorName is the name of the test vector to use for determinism checks
	TestVectorName string

	// FailOnDeterminismMismatch fails the health check if determinism verification fails
	FailOnDeterminismMismatch bool

	// MaxConsecutiveFailures is the maximum consecutive failures before marking degraded
	MaxConsecutiveFailures int
}

// DefaultHealthConfig returns the default health configuration
func DefaultHealthConfig() HealthConfig {
	return HealthConfig{
		CheckInterval:             30 * time.Second,
		CheckTimeout:              5 * time.Second,
		DeterminismTestEnabled:    true,
		ExpectedTestVectorHash:    "",
		TestVectorName:            "high_quality_verification",
		FailOnDeterminismMismatch: true,
		MaxConsecutiveFailures:    3,
	}
}

// ============================================================================
// Health Status
// ============================================================================

// HealthStatus represents the current health status of the inference system
type HealthStatus struct {
	// Healthy indicates if the inference system is fully operational
	Healthy bool

	// Degraded indicates if the system is operational but with reduced capability
	Degraded bool

	// ModelLoaded indicates if the model is successfully loaded
	ModelLoaded bool

	// ModelVersion is the currently loaded model version
	ModelVersion string

	// ModelHash is the SHA256 hash of the model weights
	ModelHash string

	// LastInferenceTime is the timestamp of the last inference
	LastInferenceTime time.Time

	// LastInferenceLatencyMs is the latency of the last inference in milliseconds
	LastInferenceLatencyMs int64

	// DeterminismVerified indicates if determinism has been verified
	DeterminismVerified bool

	// ErrorMessage contains any error message from the last check
	ErrorMessage string

	// UptimeSeconds is the uptime of the inference system in seconds
	UptimeSeconds int64

	// ConsecutiveFailures is the number of consecutive health check failures
	ConsecutiveFailures int

	// LastCheckTime is the timestamp of the last health check
	LastCheckTime time.Time

	// TestVectorHash is the hash from the last determinism test
	TestVectorHash string
}

// ============================================================================
// Health Checker
// ============================================================================

// HealthChecker performs comprehensive health checks on the inference system
type HealthChecker struct {
	// scorer is the scorer instance to check
	scorer Scorer

	// config holds health check configuration
	config HealthConfig

	// mu protects mutable state
	mu sync.RWMutex

	// lastStatus caches the last health check result
	lastStatus *HealthStatus

	// testVector is the pinned test vector for determinism verification
	testVector *SidecarTestVector

	// expectedHash is the expected output hash (set on first run if empty)
	expectedHash string

	// startTime tracks when the checker was created
	startTime time.Time

	// consecutiveFailures tracks consecutive failures
	consecutiveFailures int

	// lastInferenceTime tracks the last inference time
	lastInferenceTime time.Time

	// lastInferenceLatency tracks the last inference latency
	lastInferenceLatency time.Duration

	// stopChan signals periodic checks to stop
	stopChan chan struct{}

	// running indicates if periodic checks are running
	running bool

	// determinism controller for hash computation
	determinism *DeterminismController
}

// NewHealthChecker creates a new health checker for the given scorer
func NewHealthChecker(scorer Scorer, config HealthConfig) *HealthChecker {
	// Get the test vector
	var testVector *SidecarTestVector
	if config.DeterminismTestEnabled {
		if config.TestVectorName != "" {
			testVector = GetTestVector(config.TestVectorName)
		}
		if testVector == nil {
			testVector = GetDefaultTestVector()
		}
	}

	return &HealthChecker{
		scorer:       scorer,
		config:       config,
		testVector:   testVector,
		expectedHash: config.ExpectedTestVectorHash,
		startTime:    time.Now(),
		stopChan:     make(chan struct{}),
		determinism:  NewDeterminismController(42, true), // Default determinism settings
	}
}

// Check performs a health check and returns the current status
func (hc *HealthChecker) Check() (*HealthStatus, error) {
	startTime := time.Now()
	defer func() {
		healthCheckLatency.Observe(time.Since(startTime).Seconds())
		lastHealthCheckTimestamp.Set(float64(time.Now().Unix()))
	}()

	status := &HealthStatus{
		LastCheckTime: time.Now(),
		UptimeSeconds: int64(time.Since(hc.startTime).Seconds()),
	}

	hc.mu.RLock()
	status.LastInferenceTime = hc.lastInferenceTime
	status.LastInferenceLatencyMs = hc.lastInferenceLatency.Milliseconds()
	hc.mu.RUnlock()

	// Check if scorer is healthy
	if hc.scorer == nil {
		status.Healthy = false
		status.ErrorMessage = "scorer is nil"
		hc.recordFailure(status)
		return status, fmt.Errorf("scorer is nil")
	}

	if !hc.scorer.IsHealthy() {
		status.Healthy = false
		status.Degraded = true
		status.ErrorMessage = "scorer reports unhealthy"
		hc.recordFailure(status)
		return status, fmt.Errorf("scorer is unhealthy")
	}

	// Get model information
	status.ModelLoaded = true
	status.ModelVersion = hc.scorer.GetModelVersion()
	status.ModelHash = hc.scorer.GetModelHash()

	// Run determinism test if enabled
	if hc.config.DeterminismTestEnabled && hc.testVector != nil {
		passed, actualHash, expectedHash := hc.RunDeterminismTest()
		status.DeterminismVerified = passed
		status.TestVectorHash = actualHash

		if passed {
			determinismVerifiedGauge.Set(1)
		} else {
			determinismVerifiedGauge.Set(0)
			if hc.config.FailOnDeterminismMismatch {
				status.Healthy = false
				status.Degraded = true
				status.ErrorMessage = fmt.Sprintf("determinism test failed: expected %s, got %s", expectedHash, actualHash)
				hc.recordFailure(status)
				return status, fmt.Errorf("determinism verification failed")
			}
		}
	} else {
		status.DeterminismVerified = true // Skip test, assume verified
		determinismVerifiedGauge.Set(1)
	}

	// All checks passed
	status.Healthy = true
	status.Degraded = false
	hc.recordSuccess(status)

	return status, nil
}

// RunDeterminismTest runs the determinism verification test
// Returns: passed, actualHash, expectedHash
func (hc *HealthChecker) RunDeterminismTest() (passed bool, actualHash string, expectedHash string) {
	if hc.testVector == nil {
		return false, "", ""
	}

	// Get the test vector entry
	vec, ok := GetTestVectorByName(hc.testVector.ID)
	if !ok {
		return false, "", ""
	}

	// Convert to score inputs and run inference
	inputs := vec.Input.ConvertToScoreInputs()

	startTime := time.Now()
	result, err := hc.scorer.ComputeScore(inputs)
	inferenceLatency := time.Since(startTime)

	// Update inference tracking
	hc.mu.Lock()
	hc.lastInferenceTime = time.Now()
	hc.lastInferenceLatency = inferenceLatency
	hc.mu.Unlock()

	if err != nil {
		return false, "", hc.expectedHash
	}

	// Compute deterministic hash of the output
	actualHash = hc.computeOutputHash(result)

	// If we don't have an expected hash yet, set it from the first run
	hc.mu.Lock()
	if hc.expectedHash == "" {
		hc.expectedHash = actualHash
	}
	expectedHash = hc.expectedHash
	hc.mu.Unlock()

	// Compare hashes
	passed = actualHash == expectedHash

	// Record hash verification metric
	GetGlobalMetricsCollector().RecordHashVerification(passed)

	return passed, actualHash, expectedHash
}

// computeOutputHash computes a deterministic hash of the score result
func (hc *HealthChecker) computeOutputHash(result *ScoreResult) string {
	if result == nil {
		return ""
	}

	h := sha256.New()

	// Include key fields that should be deterministic
	h.Write([]byte(fmt.Sprintf("%d", result.Score)))
	h.Write([]byte(fmt.Sprintf("%.6f", result.Confidence)))
	h.Write([]byte(fmt.Sprintf("%.6f", result.RawScore)))
	h.Write([]byte(result.OutputHash))

	return hex.EncodeToString(h.Sum(nil))
}

// StartPeriodicChecks starts background periodic health monitoring
func (hc *HealthChecker) StartPeriodicChecks(interval time.Duration) {
	hc.mu.Lock()
	if hc.running {
		hc.mu.Unlock()
		return
	}
	hc.running = true
	hc.stopChan = make(chan struct{})
	hc.mu.Unlock()

	if interval <= 0 {
		interval = hc.config.CheckInterval
	}

	go hc.runPeriodicChecks(interval)
}

// runPeriodicChecks runs the periodic health check loop
func (hc *HealthChecker) runPeriodicChecks(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run initial check immediately
	_, _ = hc.Check()

	for {
		select {
		case <-ticker.C:
			_, _ = hc.Check()
		case <-hc.stopChan:
			return
		}
	}
}

// Stop stops periodic health checks
func (hc *HealthChecker) Stop() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if !hc.running {
		return
	}

	close(hc.stopChan)
	hc.running = false
}

// GetLastStatus returns the last cached health status
func (hc *HealthChecker) GetLastStatus() *HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	if hc.lastStatus == nil {
		return &HealthStatus{
			Healthy:      false,
			ErrorMessage: "no health check performed yet",
		}
	}

	// Return a copy to prevent mutation
	status := *hc.lastStatus
	return &status
}

// recordSuccess records a successful health check
func (hc *HealthChecker) recordSuccess(status *HealthStatus) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.consecutiveFailures = 0
	status.ConsecutiveFailures = 0
	hc.lastStatus = status

	healthCheckTotal.WithLabelValues("success").Inc()
}

// recordFailure records a failed health check
func (hc *HealthChecker) recordFailure(status *HealthStatus) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.consecutiveFailures++
	status.ConsecutiveFailures = hc.consecutiveFailures

	// Mark as degraded if too many failures
	if hc.consecutiveFailures >= hc.config.MaxConsecutiveFailures {
		status.Degraded = true
	}

	hc.lastStatus = status

	healthCheckTotal.WithLabelValues("failure").Inc()
	healthCheckFailuresTotal.Inc()
}

// IsHealthy returns a quick health check result
func (hc *HealthChecker) IsHealthy() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	if hc.lastStatus == nil {
		return false
	}
	return hc.lastStatus.Healthy
}

// IsDegraded returns whether the system is in a degraded state
func (hc *HealthChecker) IsDegraded() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	if hc.lastStatus == nil {
		return true // Unknown state is degraded
	}
	return hc.lastStatus.Degraded
}

// GetExpectedHash returns the expected output hash for determinism verification
func (hc *HealthChecker) GetExpectedHash() string {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.expectedHash
}

// SetExpectedHash sets the expected output hash for determinism verification
func (hc *HealthChecker) SetExpectedHash(hash string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.expectedHash = hash
}

// ============================================================================
// Health Check with Context
// ============================================================================

// CheckWithContext performs a health check with context support for cancellation
func (hc *HealthChecker) CheckWithContext(ctx context.Context) (*HealthStatus, error) {
	// Create a channel for the result
	type result struct {
		status *HealthStatus
		err    error
	}
	resultChan := make(chan result, 1)

	go func() {
		status, err := hc.Check()
		resultChan <- result{status: status, err: err}
	}()

	select {
	case <-ctx.Done():
		return &HealthStatus{
			Healthy:       false,
			ErrorMessage:  "health check cancelled",
			LastCheckTime: time.Now(),
		}, ctx.Err()
	case r := <-resultChan:
		return r.status, r.err
	}
}

// ============================================================================
// Factory Functions
// ============================================================================

// NewHealthCheckerWithDefaults creates a health checker with default configuration
func NewHealthCheckerWithDefaults(scorer Scorer) *HealthChecker {
	return NewHealthChecker(scorer, DefaultHealthConfig())
}

// MustNewHealthChecker creates a health checker or panics
func MustNewHealthChecker(scorer Scorer, config HealthConfig) *HealthChecker {
	if scorer == nil {
		panic("scorer cannot be nil")
	}
	return NewHealthChecker(scorer, config)
}
