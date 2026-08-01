// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestReconciliationMetricsReuseRegisteredCollectors(t *testing.T) {
	registry := prometheus.NewRegistry()
	first, err := NewReconciliationMetrics(registry)
	require.NoError(t, err)
	second, err := NewReconciliationMetrics(registry)
	require.NoError(t, err)

	second.Backlog.Set(3)
	require.Equal(t, float64(3), testutil.ToFloat64(first.Backlog))
	second.Results.WithLabelValues(string(ReconciliationStateMatched)).Set(2)
	require.Equal(t, float64(2), testutil.ToFloat64(first.Results.WithLabelValues(string(ReconciliationStateMatched))))
}

func TestReconciliationMetricsObserveBoundedProjection(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics, err := NewReconciliationMetrics(registry)
	require.NoError(t, err)
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	projection := &ReconciliationProjection{
		Jobs: map[string]ReconciliationJob{"job-1": {ID: "job-1"}, "job-2": {ID: "job-2"}},
		Results: map[string]DurableReconciliationResult{
			"job-1": {CompletedAt: now, Result: ReconciliationResult{State: ReconciliationStateUnavailable}},
		},
		Intents: map[string]ReconciliationActionIntent{
			"intent-1": {Kind: "alert_discrepancy", Severity: "high", Status: "pending"},
		},
	}
	metrics.ObserveProjection(projection)

	require.Equal(t, float64(1), testutil.ToFloat64(metrics.Backlog))
	require.Equal(t, float64(now.Unix()), testutil.ToFloat64(metrics.LastCompletedTimestamp))
	require.Equal(t, float64(1), testutil.ToFloat64(metrics.Results.WithLabelValues(string(ReconciliationStateUnavailable))))
	require.Equal(t, float64(1), testutil.ToFloat64(metrics.ActionIntents.WithLabelValues("alert_discrepancy", "high", "pending")))
	count, err := testutil.GatherAndCount(registry,
		"virtengine_provider_reconciliation_results",
		"virtengine_provider_reconciliation_last_completed_timestamp_seconds",
		"virtengine_provider_reconciliation_backlog",
		"virtengine_provider_reconciliation_action_intents",
	)
	require.NoError(t, err)
	require.Equal(t, 8, count)
}

func TestReconciliationIntentRejectsUnboundedSeverity(t *testing.T) {
	result := testDurableReconciliationResult(testReconciliationJob(), 1)
	intent := testReconciliationIntent(result)
	intent.Severity = "allocation-12345"
	require.ErrorContains(t, validateReconciliationIntent(intent, result), "severity")
}
