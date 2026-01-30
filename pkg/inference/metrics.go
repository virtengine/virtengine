// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0
//
// Package inference provides metrics for ML inference observability.
// VE-219: Deterministic identity verification runtime

package inference

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ============================================================================
// Prometheus Metrics
// ============================================================================

var (
	// inferenceCounter counts total inference requests
	inferenceCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "veid",
			Subsystem: "inference",
			Name:      "total",
			Help:      "Total number of inference requests",
		},
		[]string{"status", "mode"},
	)

	// inferenceLatency tracks inference latency
	inferenceLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "veid",
			Subsystem: "inference",
			Name:      "latency_seconds",
			Help:      "Inference latency in seconds",
			Buckets:   []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
		[]string{"mode"},
	)

	// modelInfoGauge exposes model information
	modelInfoGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "veid",
			Subsystem: "inference",
			Name:      "model_info",
			Help:      "Model information (constant 1, labels contain version/hash)",
		},
		[]string{"version", "hash"},
	)

	// sidecarHealthGauge tracks sidecar health status
	sidecarHealthGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "veid",
			Subsystem: "inference",
			Name:      "sidecar_healthy",
			Help:      "Whether the inference sidecar is healthy (1=healthy, 0=unhealthy)",
		},
	)

	// sidecarUptimeGauge tracks sidecar uptime
	sidecarUptimeGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "veid",
			Subsystem: "inference",
			Name:      "sidecar_uptime_seconds",
			Help:      "Sidecar uptime in seconds",
		},
	)

	// scoreDistribution tracks the distribution of computed scores
	scoreDistribution = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "veid",
			Subsystem: "inference",
			Name:      "score_distribution",
			Help:      "Distribution of computed identity scores",
			Buckets:   []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
		},
	)

	// confidenceDistribution tracks the distribution of confidence values
	confidenceDistribution = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "veid",
			Subsystem: "inference",
			Name:      "confidence_distribution",
			Help:      "Distribution of inference confidence values",
			Buckets:   []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
		},
	)

	// hashVerificationCounter counts hash verification results
	hashVerificationCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "veid",
			Subsystem: "inference",
			Name:      "hash_verification_total",
			Help:      "Total number of hash verifications",
		},
		[]string{"result"}, // "passed", "failed"
	)

	// featureDimensionGauge tracks feature vector dimensions
	featureDimensionGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "veid",
			Subsystem: "inference",
			Name:      "feature_dimension",
			Help:      "Expected feature vector dimension",
		},
	)
)

func init() {
	// Set feature dimension gauge
	featureDimensionGauge.Set(float64(TotalFeatureDim))
}

// ============================================================================
// Metrics Collector
// ============================================================================

// MetricsCollector collects and reports inference metrics.
type MetricsCollector struct {
	mu sync.RWMutex

	// Counters
	totalInferences      uint64
	successfulInferences uint64
	failedInferences     uint64
	timeoutInferences    uint64

	// Latency tracking
	latencySum   float64
	latencyCount int64
	latencyMax   float64

	// Score tracking
	scoreSum   float64
	scoreCount int64

	// Model info
	modelVersion string
	modelHash    string

	// Start time
	startTime time.Time
}

// NewMetricsCollector creates a new metrics collector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime: time.Now(),
	}
}

// RecordInference records an inference request.
func (mc *MetricsCollector) RecordInference(mode string, duration time.Duration, result *ScoreResult, err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.totalInferences++

	// Record latency
	latencySeconds := duration.Seconds()
	mc.latencySum += latencySeconds
	mc.latencyCount++
	if latencySeconds > mc.latencyMax {
		mc.latencyMax = latencySeconds
	}

	// Update Prometheus metrics
	inferenceLatency.WithLabelValues(mode).Observe(latencySeconds)

	status := "success"
	if err != nil {
		status = "error"
		mc.failedInferences++
		inferenceCounter.WithLabelValues(status, mode).Inc()
		return
	}

	mc.successfulInferences++
	inferenceCounter.WithLabelValues(status, mode).Inc()

	// Record score distribution
	if result != nil {
		scoreDistribution.Observe(float64(result.Score))
		confidenceDistribution.Observe(float64(result.Confidence))
		mc.scoreSum += float64(result.Score)
		mc.scoreCount++
	}
}

// RecordTimeout records a timeout.
func (mc *MetricsCollector) RecordTimeout(mode string) {
	mc.mu.Lock()
	mc.timeoutInferences++
	mc.mu.Unlock()

	inferenceCounter.WithLabelValues("timeout", mode).Inc()
}

// RecordHashVerification records a hash verification result.
func (mc *MetricsCollector) RecordHashVerification(passed bool) {
	if passed {
		hashVerificationCounter.WithLabelValues("passed").Inc()
	} else {
		hashVerificationCounter.WithLabelValues("failed").Inc()
	}
}

// SetModelInfo sets the current model information.
func (mc *MetricsCollector) SetModelInfo(version, hash string) {
	mc.mu.Lock()
	mc.modelVersion = version
	mc.modelHash = hash
	mc.mu.Unlock()

	// Update Prometheus gauge
	modelInfoGauge.WithLabelValues(version, hash).Set(1)
}

// SetSidecarHealth sets the sidecar health status.
func (mc *MetricsCollector) SetSidecarHealth(healthy bool) {
	if healthy {
		sidecarHealthGauge.Set(1)
	} else {
		sidecarHealthGauge.Set(0)
	}
}

// SetSidecarUptime sets the sidecar uptime.
func (mc *MetricsCollector) SetSidecarUptime(seconds int64) {
	sidecarUptimeGauge.Set(float64(seconds))
}

// GetStats returns current statistics.
func (mc *MetricsCollector) GetStats() MetricsStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	avgLatency := float64(0)
	if mc.latencyCount > 0 {
		avgLatency = mc.latencySum / float64(mc.latencyCount)
	}

	avgScore := float64(0)
	if mc.scoreCount > 0 {
		avgScore = mc.scoreSum / float64(mc.scoreCount)
	}

	return MetricsStats{
		TotalInferences:      mc.totalInferences,
		SuccessfulInferences: mc.successfulInferences,
		FailedInferences:     mc.failedInferences,
		TimeoutInferences:    mc.timeoutInferences,
		AverageLatencyMs:     avgLatency * 1000,
		MaxLatencyMs:         mc.latencyMax * 1000,
		AverageScore:         avgScore,
		ModelVersion:         mc.modelVersion,
		ModelHash:            mc.modelHash,
		UptimeSeconds:        int64(time.Since(mc.startTime).Seconds()),
	}
}

// MetricsStats contains current statistics.
type MetricsStats struct {
	TotalInferences      uint64
	SuccessfulInferences uint64
	FailedInferences     uint64
	TimeoutInferences    uint64
	AverageLatencyMs     float64
	MaxLatencyMs         float64
	AverageScore         float64
	ModelVersion         string
	ModelHash            string
	UptimeSeconds        int64
}

// ============================================================================
// Global Metrics Collector
// ============================================================================

var globalMetricsCollector = NewMetricsCollector()

// GetGlobalMetricsCollector returns the global metrics collector.
func GetGlobalMetricsCollector() *MetricsCollector {
	return globalMetricsCollector
}
