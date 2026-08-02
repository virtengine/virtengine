// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestReconciliationMetricsReuseRegisteredCollector(t *testing.T) {
	registry := prometheus.NewRegistry()
	first, err := NewReconciliationMetrics(registry)
	require.NoError(t, err)
	second, err := NewReconciliationMetrics(registry)
	require.NoError(t, err)
	require.Same(t, first.collector, second.collector)
	require.NoError(t, second.ObserveProjection(validMetricProjection()))
	require.Equal(t, float64(1), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_projection_available", nil))
}

func TestReconciliationMetricsObserveBoundedProjection(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics, err := NewReconciliationMetrics(registry)
	require.NoError(t, err)
	require.Equal(t, float64(0), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_projection_available", nil))
	require.NoError(t, metrics.ObserveProjection(validMetricProjection()))

	require.Equal(t, float64(1), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_projection_available", nil))
	require.Equal(t, float64(1), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_backlog", nil))
	require.Equal(t, float64(metricCompletedAt.Unix()), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_last_completed_timestamp_seconds", nil))
	require.Equal(t, float64(1), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_results", map[string]string{"state": string(ReconciliationStateUnavailable)}))
	require.Equal(t, float64(1), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_action_intents", map[string]string{"kind": "alert_discrepancy", "severity": "high", "status": "pending"}))
	require.Equal(t, 5, reconciliationMetricSeriesCount(t, registry, "virtengine_provider_reconciliation_results"))
}

func TestReconciliationMetricsFailureOmitsOperationalValues(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics, err := NewReconciliationMetrics(registry)
	require.NoError(t, err)
	require.NoError(t, metrics.ObserveProjection(validMetricProjection()))
	metrics.ObserveProjectionFailure()

	require.Equal(t, float64(0), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_projection_available", nil))
	for _, name := range []string{
		"virtengine_provider_reconciliation_results",
		"virtengine_provider_reconciliation_backlog",
		"virtengine_provider_reconciliation_last_completed_timestamp_seconds",
		"virtengine_provider_reconciliation_action_intents",
	} {
		require.Equal(t, 0, reconciliationMetricSeriesCount(t, registry, name), name)
	}
}

func TestReconciliationMetricsRejectsInvalidProjectionBeforePublication(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics, err := NewReconciliationMetrics(registry)
	require.NoError(t, err)
	projection := validMetricProjection()
	jobID := testReconciliationJob().ID
	result := projection.Results[jobID]
	result.Result.State = ReconciliationState("allocation-private-canary")
	projection.Results[jobID] = result
	require.EqualError(t, metrics.ObserveProjection(projection), "invalid reconciliation metrics result")
	require.Equal(t, float64(0), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_projection_available", nil))
	require.Equal(t, 0, reconciliationMetricSeriesCount(t, registry, "virtengine_provider_reconciliation_results"))

	projection = validMetricProjection()
	projection.Jobs = map[string]ReconciliationJob{}
	require.EqualError(t, metrics.ObserveProjection(projection), "invalid reconciliation metrics projection")

	projection = validMetricProjection()
	result = projection.Results[testReconciliationJob().ID]
	result.Result.ReasonCode = ReconciliationReasonExactMatch
	projection.Results[result.JobID] = result
	require.EqualError(t, metrics.ObserveProjection(projection), "invalid reconciliation metrics result")

	projection = validMetricProjection()
	result = projection.Results[testReconciliationJob().ID]
	result.Result.AllocationID = "other-allocation"
	projection.Results[result.JobID] = result
	require.EqualError(t, metrics.ObserveProjection(projection), "invalid reconciliation metrics result")

	projection = validMetricProjection()
	result = projection.Results[testReconciliationJob().ID]
	result.Result.State = ReconciliationStateMatched
	result.Result.ReasonCode = ReconciliationReasonExactMatch
	result.Result.Score = 99
	projection.Results[result.JobID] = result
	require.EqualError(t, metrics.ObserveProjection(projection), "invalid reconciliation metrics result")
}

func TestReconciliationMetricsConcurrentSnapshotsAreAtomic(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics, err := NewReconciliationMetrics(registry)
	require.NoError(t, err)
	projection := validMetricProjection()
	var wait sync.WaitGroup
	for iteration := 0; iteration < 100; iteration++ {
		wait.Add(1)
		go func(available bool) {
			defer wait.Done()
			if available {
				_ = metrics.ObserveProjection(projection)
			} else {
				metrics.ObserveProjectionFailure()
			}
		}(iteration%2 == 0)
		families, err := registry.Gather()
		require.NoError(t, err)
		available := reconciliationFamilyGaugeValue(t, families, "virtengine_provider_reconciliation_projection_available")
		if available == 0 {
			require.Zero(t, reconciliationFamilySeriesCount(families, "virtengine_provider_reconciliation_results"))
		} else {
			require.Equal(t, len(reconciliationMetricStates), reconciliationFamilySeriesCount(families, "virtengine_provider_reconciliation_results"))
			require.Equal(t, 1, reconciliationFamilySeriesCount(families, "virtengine_provider_reconciliation_backlog"))
		}
	}
	wait.Wait()
}

func TestReconciliationIntentRejectsUnboundedSeverity(t *testing.T) {
	result := testDurableReconciliationResult(testReconciliationJob(), 1)
	intent := testReconciliationIntent(result)
	intent.Severity = "allocation-12345"
	require.ErrorContains(t, validateReconciliationIntent(intent, result), "severity")
	intent = testReconciliationIntent(result)
	intent.AllocationID = "other-allocation"
	require.ErrorContains(t, validateReconciliationIntent(intent, result), "invalid reconciliation action intent")
}

var metricCompletedAt = time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

func validMetricProjection() *ReconciliationProjection {
	job := testReconciliationJob()
	result := testDurableReconciliationResult(job, 1)
	result.CompletedAt = metricCompletedAt
	result.Result.State = ReconciliationStateUnavailable
	result.Result.ReasonCode = ReconciliationReasonIndependentEvidenceUnavailable
	result.ResultDigest, _ = canonicalReconciliationDigest("result", result.Result)
	intent := testReconciliationIntent(result)
	return &ReconciliationProjection{
		Jobs: map[string]ReconciliationJob{
			job.ID:  job,
			"job-2": {ID: "job-2"},
		},
		Results: map[string]DurableReconciliationResult{job.ID: result},
		Intents: map[string]ReconciliationActionIntent{intent.ID: intent},
	}
}

func reconciliationMetricValue(t *testing.T, registry *prometheus.Registry, name string, labels map[string]string) float64 {
	t.Helper()
	families, err := registry.Gather()
	require.NoError(t, err)
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, metric := range family.Metric {
			if reconciliationMetricLabelsMatch(metric, labels) {
				return metric.GetGauge().GetValue()
			}
		}
	}
	t.Fatalf("metric %s with labels %v not found", name, labels)
	return 0
}

func reconciliationMetricSeriesCount(t *testing.T, registry *prometheus.Registry, name string) int {
	t.Helper()
	families, err := registry.Gather()
	require.NoError(t, err)
	for _, family := range families {
		if family.GetName() == name {
			return len(family.Metric)
		}
	}
	return 0
}

func reconciliationMetricLabelsMatch(metric *dto.Metric, expected map[string]string) bool {
	for name, value := range expected {
		found := false
		for _, label := range metric.Label {
			if label.GetName() == name && label.GetValue() == value {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func reconciliationFamilyGaugeValue(t *testing.T, families []*dto.MetricFamily, name string) float64 {
	t.Helper()
	for _, family := range families {
		if family.GetName() == name && len(family.Metric) == 1 {
			return family.Metric[0].GetGauge().GetValue()
		}
	}
	t.Fatalf("metric family %s not found", name)
	return 0
}

func reconciliationFamilySeriesCount(families []*dto.MetricFamily, name string) int {
	for _, family := range families {
		if family.GetName() == name {
			return len(family.Metric)
		}
	}
	return 0
}
