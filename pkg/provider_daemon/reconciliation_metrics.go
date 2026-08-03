// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"errors"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var reconciliationMetricStates = [...]ReconciliationState{
	ReconciliationStateMatched,
	ReconciliationStateMismatched,
	ReconciliationStateUnavailable,
	ReconciliationStateStale,
	ReconciliationStateUnresolved,
}

type reconciliationIntentMetricKey struct {
	kind     string
	severity string
	status   string
}

type reconciliationMetricSnapshot struct {
	available     bool
	results       map[ReconciliationState]float64
	lastCompleted float64
	backlog       float64
	intents       map[reconciliationIntentMetricKey]float64
}

// ReconciliationMetrics exports one atomic bounded durable reconciliation projection.
type ReconciliationMetrics struct {
	collector *reconciliationMetricsCollector
}

func NewReconciliationMetrics(registerer prometheus.Registerer) (*ReconciliationMetrics, error) {
	if registerer == nil {
		return nil, errors.New("reconciliation metrics registerer is required")
	}
	candidate := newReconciliationMetricsCollector()
	if err := registerer.Register(candidate); err != nil {
		already, ok := err.(prometheus.AlreadyRegisteredError)
		if !ok {
			return nil, err
		}
		existing, ok := already.ExistingCollector.(*reconciliationMetricsCollector)
		if !ok {
			return nil, errors.New("registered reconciliation collector has unexpected type")
		}
		candidate = existing
	}
	return &ReconciliationMetrics{collector: candidate}, nil
}

func (metrics *ReconciliationMetrics) ObserveProjection(projection *ReconciliationProjection) error {
	if metrics == nil || metrics.collector == nil {
		return errors.New("reconciliation metrics are required")
	}
	snapshot, err := buildReconciliationMetricSnapshot(projection)
	if err != nil {
		metrics.ObserveProjectionFailure()
		return err
	}
	metrics.collector.mu.Lock()
	metrics.collector.snapshot = snapshot
	metrics.collector.mu.Unlock()
	return nil
}

func (metrics *ReconciliationMetrics) ObserveProjectionFailure() {
	if metrics == nil || metrics.collector == nil {
		return
	}
	metrics.collector.mu.Lock()
	metrics.collector.snapshot = reconciliationMetricSnapshot{}
	metrics.collector.mu.Unlock()
}

func buildReconciliationMetricSnapshot(projection *ReconciliationProjection) (reconciliationMetricSnapshot, error) {
	if projection == nil || len(projection.Results) > len(projection.Jobs) {
		return reconciliationMetricSnapshot{}, errors.New("invalid reconciliation metrics projection")
	}
	snapshot := reconciliationMetricSnapshot{
		available: true,
		results:   make(map[ReconciliationState]float64, len(reconciliationMetricStates)),
		backlog:   float64(len(projection.Jobs) - len(projection.Results)),
		intents:   make(map[reconciliationIntentMetricKey]float64),
	}
	for jobID, result := range projection.Results {
		job, ok := projection.Jobs[jobID]
		if !ok || result.JobID != jobID || result.Result.AllocationID != job.AllocationID ||
			validateReconciliationResultSemantics(result.Result) != nil || result.CompletedAt.IsZero() {
			return reconciliationMetricSnapshot{}, errors.New("invalid reconciliation metrics result")
		}
		snapshot.results[result.Result.State]++
		if completed := float64(result.CompletedAt.Unix()); completed > snapshot.lastCompleted {
			snapshot.lastCompleted = completed
		}
	}
	for _, intent := range projection.Intents {
		result, ok := projection.Results[intent.JobID]
		if !ok || validateReconciliationIntent(intent, result) != nil {
			return reconciliationMetricSnapshot{}, errors.New("invalid reconciliation metrics intent")
		}
		key := reconciliationIntentMetricKey{kind: intent.Kind, severity: intent.Severity, status: intent.Status}
		snapshot.intents[key]++
	}
	return snapshot, nil
}

func validReconciliationMetricState(state ReconciliationState) bool {
	for _, allowed := range reconciliationMetricStates {
		if state == allowed {
			return true
		}
	}
	return false
}

func validateReconciliationResultSemantics(result ReconciliationResult) error {
	if result.AllocationID == "" || result.ReconciliationTime.IsZero() || result.Score < 0 || result.Score > 100 || !validReconciliationMetricState(result.State) {
		return errors.New("invalid reconciliation result semantics")
	}
	validReason := false
	switch result.State {
	case ReconciliationStateMatched:
		validReason = (result.ReasonCode == ReconciliationReasonExactMatch || result.ReasonCode == ReconciliationReasonWithinTolerance) &&
			result.Score == 100 && len(result.Discrepancies) == 0
	case ReconciliationStateMismatched:
		validReason = result.ReasonCode == ReconciliationReasonMetricThresholdExceeded
	case ReconciliationStateUnavailable:
		validReason = result.ReasonCode == ReconciliationReasonProviderEvidenceUnavailable || result.ReasonCode == ReconciliationReasonIndependentEvidenceUnavailable
	case ReconciliationStateStale:
		validReason = result.ReasonCode == ReconciliationReasonProviderEvidenceStale || result.ReasonCode == ReconciliationReasonIndependentEvidenceStale
	case ReconciliationStateUnresolved:
		validReason = result.ReasonCode == ReconciliationReasonPartialEvidence || result.ReasonCode == ReconciliationReasonMalformedEvidence
	}
	if !validReason {
		return errors.New("reconciliation state and reason are inconsistent")
	}
	return nil
}

type reconciliationMetricsCollector struct {
	mu       sync.RWMutex
	snapshot reconciliationMetricSnapshot

	available     *prometheus.Desc
	results       *prometheus.Desc
	lastCompleted *prometheus.Desc
	backlog       *prometheus.Desc
	intents       *prometheus.Desc
}

func newReconciliationMetricsCollector() *reconciliationMetricsCollector {
	return &reconciliationMetricsCollector{
		available: prometheus.NewDesc(
			"virtengine_provider_reconciliation_projection_available",
			"Whether the complete durable reconciliation projection is valid and available.", nil, nil,
		),
		results: prometheus.NewDesc(
			"virtengine_provider_reconciliation_results",
			"Current durable reconciliation results by bounded state.", []string{"state"}, nil,
		),
		lastCompleted: prometheus.NewDesc(
			"virtengine_provider_reconciliation_last_completed_timestamp_seconds",
			"Unix timestamp of the newest durably completed reconciliation.", nil, nil,
		),
		backlog: prometheus.NewDesc(
			"virtengine_provider_reconciliation_backlog",
			"Number of durable reconciliation jobs without a committed result.", nil, nil,
		),
		intents: prometheus.NewDesc(
			"virtengine_provider_reconciliation_action_intents",
			"Current durable action intents by bounded kind, severity, and status.", []string{"kind", "severity", "status"}, nil,
		),
	}
}

func (collector *reconciliationMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.available
	ch <- collector.results
	ch <- collector.lastCompleted
	ch <- collector.backlog
	ch <- collector.intents
}

func (collector *reconciliationMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	collector.mu.RLock()
	defer collector.mu.RUnlock()
	snapshot := collector.snapshot
	ch <- prometheus.MustNewConstMetric(collector.available, prometheus.GaugeValue, reconciliationBoolValue(snapshot.available))
	if !snapshot.available {
		return
	}
	for _, state := range reconciliationMetricStates {
		ch <- prometheus.MustNewConstMetric(collector.results, prometheus.GaugeValue, snapshot.results[state], string(state))
	}
	ch <- prometheus.MustNewConstMetric(collector.lastCompleted, prometheus.GaugeValue, snapshot.lastCompleted)
	ch <- prometheus.MustNewConstMetric(collector.backlog, prometheus.GaugeValue, snapshot.backlog)
	for key, count := range snapshot.intents {
		ch <- prometheus.MustNewConstMetric(collector.intents, prometheus.GaugeValue, count, key.kind, key.severity, key.status)
	}
}

func reconciliationBoolValue(value bool) float64 {
	if value {
		return 1
	}
	return 0
}
