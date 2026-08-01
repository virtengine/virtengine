// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// ReconciliationMetrics exports bounded durable reconciliation state.
type ReconciliationMetrics struct {
	Results                *prometheus.GaugeVec
	LastCompletedTimestamp prometheus.Gauge
	Backlog                prometheus.Gauge
	ActionIntents          *prometheus.GaugeVec
}

func NewReconciliationMetrics(registerer prometheus.Registerer) (*ReconciliationMetrics, error) {
	if registerer == nil {
		return nil, fmt.Errorf("reconciliation metrics registerer is required")
	}
	results, err := registerGaugeVec(registerer, prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "virtengine", Subsystem: "provider_reconciliation", Name: "results",
		Help: "Current durable reconciliation results by bounded state.",
	}, []string{"state"}))
	if err != nil {
		return nil, err
	}
	lastCompleted, err := registerGauge(registerer, prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "virtengine", Subsystem: "provider_reconciliation", Name: "last_completed_timestamp_seconds",
		Help: "Unix timestamp of the newest durably completed reconciliation.",
	}))
	if err != nil {
		return nil, err
	}
	backlog, err := registerGauge(registerer, prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "virtengine", Subsystem: "provider_reconciliation", Name: "backlog",
		Help: "Number of durable reconciliation jobs without a committed result.",
	}))
	if err != nil {
		return nil, err
	}
	intents, err := registerGaugeVec(registerer, prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "virtengine", Subsystem: "provider_reconciliation", Name: "action_intents",
		Help: "Current durable action intents by bounded kind, severity, and status.",
	}, []string{"kind", "severity", "status"}))
	if err != nil {
		return nil, err
	}
	return &ReconciliationMetrics{Results: results, LastCompletedTimestamp: lastCompleted, Backlog: backlog, ActionIntents: intents}, nil
}

func (m *ReconciliationMetrics) ObserveProjection(projection *ReconciliationProjection) {
	if m == nil || projection == nil {
		return
	}
	m.Results.Reset()
	for _, state := range []ReconciliationState{
		ReconciliationStateMatched, ReconciliationStateMismatched, ReconciliationStateUnavailable,
		ReconciliationStateStale, ReconciliationStateUnresolved,
	} {
		m.Results.WithLabelValues(string(state)).Set(0)
	}
	latest := time.Time{}
	for _, result := range projection.Results {
		m.Results.WithLabelValues(string(result.Result.State)).Inc()
		if result.CompletedAt.After(latest) {
			latest = result.CompletedAt
		}
	}
	if latest.IsZero() {
		m.LastCompletedTimestamp.Set(0)
	} else {
		m.LastCompletedTimestamp.Set(float64(latest.Unix()))
	}
	m.Backlog.Set(float64(len(projection.Jobs) - len(projection.Results)))
	m.ActionIntents.Reset()
	for _, intent := range projection.Intents {
		m.ActionIntents.WithLabelValues(intent.Kind, intent.Severity, intent.Status).Inc()
	}
}

func registerGaugeVec(registerer prometheus.Registerer, collector *prometheus.GaugeVec) (*prometheus.GaugeVec, error) {
	if err := registerer.Register(collector); err != nil {
		alreadyRegistered, ok := err.(prometheus.AlreadyRegisteredError)
		if !ok {
			return nil, err
		}
		existing, ok := alreadyRegistered.ExistingCollector.(*prometheus.GaugeVec)
		if !ok {
			return nil, fmt.Errorf("registered reconciliation collector has unexpected type")
		}
		return existing, nil
	}
	return collector, nil
}

func registerGauge(registerer prometheus.Registerer, collector prometheus.Gauge) (prometheus.Gauge, error) {
	if err := registerer.Register(collector); err != nil {
		alreadyRegistered, ok := err.(prometheus.AlreadyRegisteredError)
		if !ok {
			return nil, err
		}
		existing, ok := alreadyRegistered.ExistingCollector.(prometheus.Gauge)
		if !ok {
			return nil, fmt.Errorf("registered reconciliation collector has unexpected type")
		}
		return existing, nil
	}
	return collector, nil
}
