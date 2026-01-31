// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

// Package chaos provides chaos engineering utilities for VirtEngine.
// This file implements Prometheus metrics for tracking chaos experiments,
// including experiment lifecycle, SLO violations, recovery times, and
// steady-state probe results.
package chaos

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// ============================================================================
// Metric Names
// ============================================================================

// Metric name constants for external use.
const (
	MetricNamespace = "virtengine"
	MetricSubsystem = "chaos"

	// Counter metrics
	MetricExperimentsTotal        = "experiments_total"
	MetricSLOViolationsTotal      = "slo_violations_total"
	MetricSteadyStateProbeSuccess = "steady_state_probe_success_total"
	MetricSteadyStateProbeFailed  = "steady_state_probe_failed_total"

	// Gauge metrics
	MetricExperimentsActive   = "experiments_active"
	MetricBlastRadiusAffected = "blast_radius_affected"

	// Histogram metrics
	MetricExperimentDuration     = "experiment_duration_seconds"
	MetricRecoveryDuration       = "recovery_duration_seconds"
	MetricFaultInjectionDuration = "fault_injection_duration_seconds"
	MetricMeanTimeToRecovery     = "mean_time_to_recovery_seconds"
	MetricMeanTimeToDetect       = "mean_time_to_detect_seconds"
)

// ============================================================================
// Histogram Buckets
// ============================================================================

var (
	// experimentDurationBuckets defines histogram buckets for experiment durations.
	// Experiments typically run from seconds to hours.
	experimentDurationBuckets = []float64{
		1, 5, 10, 30, 60, // 1s to 1min
		120, 300, 600, // 2min to 10min
		900, 1800, 3600, // 15min to 1hr
		7200, 14400, 28800, // 2hr to 8hr
	}

	// recoveryDurationBuckets defines histogram buckets for recovery times.
	// Recovery times are typically seconds to minutes.
	recoveryDurationBuckets = []float64{
		0.5, 1, 2, 5, 10, // sub-second to 10s
		30, 60, 120, 300, // 30s to 5min
		600, 900, 1800, // 10min to 30min
	}

	// faultInjectionBuckets defines histogram buckets for fault injection durations.
	// Fault injections are typically milliseconds to minutes.
	faultInjectionBuckets = []float64{
		0.001, 0.005, 0.01, 0.025, 0.05, // 1ms to 50ms
		0.1, 0.25, 0.5, 1, 2.5, // 100ms to 2.5s
		5, 10, 30, 60, 120, // 5s to 2min
	}

	// mttrBuckets defines histogram buckets for Mean Time To Recovery.
	// MTTR is typically seconds to hours.
	mttrBuckets = []float64{
		1, 5, 10, 30, 60, // 1s to 1min
		120, 300, 600, 900, // 2min to 15min
		1800, 3600, 7200, 14400, // 30min to 4hr
	}

	// mttdBuckets defines histogram buckets for Mean Time To Detect.
	// MTTD is typically sub-second to minutes.
	mttdBuckets = []float64{
		0.1, 0.5, 1, 2, 5, // 100ms to 5s
		10, 30, 60, 120, 300, // 10s to 5min
		600, 900, 1800, // 10min to 30min
	}
)

// ============================================================================
// PrometheusMetrics Struct
// ============================================================================

// PrometheusMetrics provides Prometheus metrics for chaos experiments.
// This is separate from the simpler Metrics type in controller.go which
// provides basic counters for in-memory tracking.
type PrometheusMetrics struct {
	mu sync.RWMutex

	// Counters
	experimentsTotal        *prometheus.CounterVec
	sloViolationsTotal      *prometheus.CounterVec
	steadyStateProbeSuccess *prometheus.CounterVec
	steadyStateProbeFailed  *prometheus.CounterVec

	// Gauges
	experimentsActive   prometheus.Gauge
	blastRadiusAffected *prometheus.GaugeVec

	// Histograms
	experimentDuration     *prometheus.HistogramVec
	recoveryDuration       *prometheus.HistogramVec
	faultInjectionDuration *prometheus.HistogramVec
	meanTimeToRecovery     prometheus.Histogram
	meanTimeToDetect       prometheus.Histogram

	// Registry for custom registrations
	registry prometheus.Registerer

	// collectors holds all collectors for Reset()
	collectors []prometheus.Collector
}

// ============================================================================
// Constructors
// ============================================================================

// NewPrometheusMetrics creates a new PrometheusMetrics instance and registers
// all metrics with the provided registry.
func NewPrometheusMetrics(registry prometheus.Registerer) *PrometheusMetrics {
	m := &PrometheusMetrics{
		registry: registry,
	}

	// Create counters
	m.experimentsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Subsystem: MetricSubsystem,
			Name:      MetricExperimentsTotal,
			Help:      "Total number of chaos experiments by type and status",
		},
		[]string{"type", "status"},
	)

	m.sloViolationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Subsystem: MetricSubsystem,
			Name:      MetricSLOViolationsTotal,
			Help:      "Total number of SLO violations during chaos experiments",
		},
		[]string{"slo_name"},
	)

	m.steadyStateProbeSuccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Subsystem: MetricSubsystem,
			Name:      MetricSteadyStateProbeSuccess,
			Help:      "Total number of successful steady-state probe checks",
		},
		[]string{"probe_name"},
	)

	m.steadyStateProbeFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Subsystem: MetricSubsystem,
			Name:      MetricSteadyStateProbeFailed,
			Help:      "Total number of failed steady-state probe checks",
		},
		[]string{"probe_name"},
	)

	// Create gauges
	m.experimentsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: MetricNamespace,
			Subsystem: MetricSubsystem,
			Name:      MetricExperimentsActive,
			Help:      "Number of currently active chaos experiments",
		},
	)

	m.blastRadiusAffected = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: MetricNamespace,
			Subsystem: MetricSubsystem,
			Name:      MetricBlastRadiusAffected,
			Help:      "Number of resources affected by a chaos experiment",
		},
		[]string{"experiment_id"},
	)

	// Create histograms
	m.experimentDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: MetricNamespace,
			Subsystem: MetricSubsystem,
			Name:      MetricExperimentDuration,
			Help:      "Duration of chaos experiments in seconds",
			Buckets:   experimentDurationBuckets,
		},
		[]string{"type"},
	)

	m.recoveryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: MetricNamespace,
			Subsystem: MetricSubsystem,
			Name:      MetricRecoveryDuration,
			Help:      "Time taken for system to recover after chaos injection in seconds",
			Buckets:   recoveryDurationBuckets,
		},
		[]string{"type"},
	)

	m.faultInjectionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: MetricNamespace,
			Subsystem: MetricSubsystem,
			Name:      MetricFaultInjectionDuration,
			Help:      "Duration of fault injection in seconds",
			Buckets:   faultInjectionBuckets,
		},
		[]string{"fault_type"},
	)

	m.meanTimeToRecovery = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: MetricNamespace,
			Subsystem: MetricSubsystem,
			Name:      MetricMeanTimeToRecovery,
			Help:      "Mean Time To Recovery (MTTR) across all experiments in seconds",
			Buckets:   mttrBuckets,
		},
	)

	m.meanTimeToDetect = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: MetricNamespace,
			Subsystem: MetricSubsystem,
			Name:      MetricMeanTimeToDetect,
			Help:      "Mean Time To Detect (MTTD) failures in seconds",
			Buckets:   mttdBuckets,
		},
	)

	// Collect all collectors for registration and reset
	m.collectors = []prometheus.Collector{
		m.experimentsTotal,
		m.sloViolationsTotal,
		m.steadyStateProbeSuccess,
		m.steadyStateProbeFailed,
		m.experimentsActive,
		m.blastRadiusAffected,
		m.experimentDuration,
		m.recoveryDuration,
		m.faultInjectionDuration,
		m.meanTimeToRecovery,
		m.meanTimeToDetect,
	}

	// Register all metrics
	for _, c := range m.collectors {
		registry.MustRegister(c)
	}

	return m
}

// DefaultPrometheusMetrics creates a new PrometheusMetrics instance using the
// default Prometheus registerer.
func DefaultPrometheusMetrics() *PrometheusMetrics {
	return NewPrometheusMetrics(prometheus.DefaultRegisterer)
}

// ============================================================================
// Recording Methods
// ============================================================================

// RecordExperimentStart records the start of a chaos experiment.
// Uses the Experiment type from controller.go.
func (m *PrometheusMetrics) RecordExperimentStart(exp *Experiment) {
	if exp == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Increment active experiments
	m.experimentsActive.Inc()

	// Record experiment started (status: started)
	// Use experiment name as type since Experiment doesn't have a Type field
	expType := "unknown"
	if len(exp.Tags) > 0 {
		expType = exp.Tags[0]
	}
	m.experimentsTotal.WithLabelValues(expType, "started").Inc()
}

// RecordExperimentComplete records the completion of a chaos experiment.
func (m *PrometheusMetrics) RecordExperimentComplete(exp *Experiment, success bool, duration time.Duration) {
	if exp == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Decrement active experiments
	m.experimentsActive.Dec()

	// Record status
	status := "success"
	if !success {
		status = "failed"
	}

	// Use experiment tag or name as type
	expType := "unknown"
	if len(exp.Tags) > 0 {
		expType = exp.Tags[0]
	}
	m.experimentsTotal.WithLabelValues(expType, status).Inc()

	// Record duration
	m.experimentDuration.WithLabelValues(expType).Observe(duration.Seconds())
}

// RecordSLOViolation records an SLO violation during a chaos experiment.
func (m *PrometheusMetrics) RecordSLOViolation(sloName string, _ *Experiment) {
	if sloName == "" {
		return
	}

	m.sloViolationsTotal.WithLabelValues(sloName).Inc()
}

// RecordRecoveryTime records the time taken for system recovery.
func (m *PrometheusMetrics) RecordRecoveryTime(expType string, duration time.Duration) {
	m.recoveryDuration.WithLabelValues(expType).Observe(duration.Seconds())
}

// RecordProbeResult records a steady-state probe result.
func (m *PrometheusMetrics) RecordProbeResult(probeName string, success bool) {
	if probeName == "" {
		return
	}

	if success {
		m.steadyStateProbeSuccess.WithLabelValues(probeName).Inc()
	} else {
		m.steadyStateProbeFailed.WithLabelValues(probeName).Inc()
	}
}

// RecordFaultInjection records a fault injection event.
func (m *PrometheusMetrics) RecordFaultInjection(faultType string, duration time.Duration) {
	if faultType == "" {
		return
	}

	m.faultInjectionDuration.WithLabelValues(faultType).Observe(duration.Seconds())
}

// RecordMTTR records Mean Time To Recovery.
func (m *PrometheusMetrics) RecordMTTR(duration time.Duration) {
	m.meanTimeToRecovery.Observe(duration.Seconds())
}

// RecordMTTD records Mean Time To Detect.
func (m *PrometheusMetrics) RecordMTTD(duration time.Duration) {
	m.meanTimeToDetect.Observe(duration.Seconds())
}

// SetBlastRadius sets the number of affected resources for an experiment.
func (m *PrometheusMetrics) SetBlastRadius(expID string, affected int) {
	if expID == "" {
		return
	}

	m.blastRadiusAffected.WithLabelValues(expID).Set(float64(affected))
}

// ============================================================================
// Utility Methods
// ============================================================================

// Reset resets all metrics to their initial state.
// This is primarily intended for testing purposes.
func (m *PrometheusMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Reset counters
	m.experimentsTotal.Reset()
	m.sloViolationsTotal.Reset()
	m.steadyStateProbeSuccess.Reset()
	m.steadyStateProbeFailed.Reset()

	// Reset gauges
	m.experimentsActive.Set(0)
	m.blastRadiusAffected.Reset()

	// Note: Histograms cannot be reset in Prometheus.
	// For testing, create a new PrometheusMetrics instance with a new registry.
}

// Unregister unregisters all metrics from the registry.
// This is useful for testing cleanup.
func (m *PrometheusMetrics) Unregister() {
	for _, c := range m.collectors {
		m.registry.Unregister(c)
	}
}
