// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestOperationalMetricsAllowedVocabulary(t *testing.T) {
	metrics, err := NewOperationalMetrics(prometheus.NewRegistry())
	require.NoError(t, err)
	stages := []ModelStage{
		ModelStageDocumentRouting, ModelStageOCR, ModelStageDocumentAuthenticity, ModelStageFaceDetection,
		ModelStageFaceComparison, ModelStageLiveness, ModelStageFraudFusion, ModelStagePolicyEvaluation,
	}
	for _, stage := range stages {
		for _, result := range []StageResult{StageResultSuccess, StageResultUnavailable, StageResultError} {
			reasons := []OperationalReasonCode(nil)
			if result != StageResultSuccess {
				reasons = []OperationalReasonCode{ReasonUnknown}
			}
			require.NoError(t, metrics.ObserveStage(stage, result, time.Millisecond, reasons...))
		}
	}
	decisions := []DecisionKind{
		DecisionDocument, DecisionFaceMatch, DecisionLiveness, DecisionRisk, DecisionIdentity, DecisionUniqueness, DecisionEligibility,
	}
	for _, decision := range decisions {
		for _, outcome := range []DecisionOutcome{
			DecisionOutcomePass, DecisionOutcomeReview, DecisionOutcomeReject, DecisionOutcomeUnavailable, DecisionOutcomeError,
		} {
			reasons := []OperationalReasonCode(nil)
			if outcome != DecisionOutcomePass {
				reasons = []OperationalReasonCode{ReasonUnknown}
			}
			require.NoError(t, metrics.ObserveDecision(decision, outcome, time.Millisecond, reasons...))
		}
	}
	for _, test := range []struct {
		outcome DecisionOutcome
		reason  OperationalReasonCode
	}{
		{DecisionOutcomeError, ReasonInvalidInput},
		{DecisionOutcomeUnavailable, ReasonInputUnavailable},
		{DecisionOutcomeUnavailable, ReasonModelUnavailable},
		{DecisionOutcomeError, ReasonModelError},
		{DecisionOutcomeError, ReasonTimeout},
		{DecisionOutcomeReview, ReasonLowConfidence},
		{DecisionOutcomeReject, ReasonThresholdNotMet},
		{DecisionOutcomeReject, ReasonHardGateFailed},
		{DecisionOutcomeReview, ReasonReviewRequired},
		{DecisionOutcomeReject, ReasonPolicyDenied},
		{DecisionOutcomeError, ReasonUnknown},
	} {
		require.NoError(t, metrics.ObserveDecision(DecisionIdentity, test.outcome, time.Millisecond, test.reason))
	}
}

func TestOperationalMetricsRejectBeforeMutationAndDoNotLeak(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics, err := NewOperationalMetrics(registry)
	require.NoError(t, err)
	secret := "account-ve1-secret-biometric-nullifier"
	err = metrics.ObserveStage(ModelStage(secret), StageResultSuccess, time.Second, ReasonUnknown)
	require.EqualError(t, err, "unsupported model stage")
	require.NotContains(t, err.Error(), secret)
	err = metrics.ObserveDecision(DecisionIdentity, DecisionOutcome(secret), time.Second)
	require.EqualError(t, err, "unsupported decision outcome")
	require.NotContains(t, err.Error(), secret)
	err = metrics.ObserveStage(ModelStageOCR, StageResultSuccess, -time.Second)
	require.EqualError(t, err, "stage duration cannot be negative")
	err = metrics.ObserveStage(ModelStageOCR, StageResultUnavailable, time.Second)
	require.EqualError(t, err, "non-success stage result requires an operational reason")
	err = metrics.ObserveDecision(DecisionIdentity, DecisionOutcomeReject, time.Second)
	require.EqualError(t, err, "non-pass decision outcome requires an operational reason")
	err = metrics.ObserveStage(ModelStageOCR, StageResultSuccess, time.Second, ReasonModelError)
	require.EqualError(t, err, "successful stage result cannot include an operational reason")
	err = metrics.ObserveDecision(DecisionIdentity, DecisionOutcomeReject, time.Second, ReasonModelError)
	require.EqualError(t, err, "operational reason is incompatible with decision outcome")

	families, err := registry.Gather()
	require.NoError(t, err)
	var exposition bytes.Buffer
	for _, family := range families {
		exposition.WriteString(family.String())
	}
	require.NotContains(t, exposition.String(), secret)
	require.Equal(t, float64(0), testutil.ToFloat64(metrics.stageOperations.WithLabelValues(string(ModelStageOCR), string(StageResultSuccess))))
	err = metrics.ObserveStage(ModelStageOCR, StageResultSuccess, time.Second, OperationalReasonCode(secret))
	require.EqualError(t, err, "unsupported operational reason")
	require.Equal(t, float64(0), testutil.ToFloat64(metrics.stageOperations.WithLabelValues(string(ModelStageOCR), string(StageResultSuccess))))
}

func TestOperationalMetricsFreshnessAndReasonDeduplication(t *testing.T) {
	metrics, err := NewOperationalMetrics(prometheus.NewRegistry())
	require.NoError(t, err)
	fixed := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	metrics.now = func() time.Time { return fixed }

	require.NoError(t, metrics.ObserveStage(ModelStageOCR, StageResultUnavailable, time.Second, ReasonModelUnavailable, ReasonModelUnavailable))
	require.Equal(t, float64(1), testutil.ToFloat64(metrics.stageReasons.WithLabelValues(string(ModelStageOCR), string(ReasonModelUnavailable))))
	require.Equal(t, float64(0), testutil.ToFloat64(metrics.stageLastSuccess.WithLabelValues(string(ModelStageOCR))))
	require.NoError(t, metrics.ObserveStage(ModelStageOCR, StageResultSuccess, time.Second))
	require.Equal(t, float64(fixed.Unix()), testutil.ToFloat64(metrics.stageLastSuccess.WithLabelValues(string(ModelStageOCR))))

	require.NoError(t, metrics.ObserveDecision(DecisionIdentity, DecisionOutcomeError, time.Second, ReasonModelError))
	require.Equal(t, float64(0), testutil.ToFloat64(metrics.decisionLastSuccess.WithLabelValues(string(DecisionIdentity))))
	require.NoError(t, metrics.ObserveDecision(DecisionIdentity, DecisionOutcomeReject, time.Second, ReasonPolicyDenied))
	require.Equal(t, float64(0), testutil.ToFloat64(metrics.decisionLastSuccess.WithLabelValues(string(DecisionIdentity))))
	require.NoError(t, metrics.ObserveDecision(DecisionIdentity, DecisionOutcomePass, time.Second))
	require.Equal(t, float64(fixed.Unix()), testutil.ToFloat64(metrics.decisionLastSuccess.WithLabelValues(string(DecisionIdentity))))
}

func TestOperationalMetricsDuplicateRegistrationSharesCollectors(t *testing.T) {
	registry := prometheus.NewRegistry()
	first, err := NewOperationalMetrics(registry)
	require.NoError(t, err)
	fixed := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	first.now = func() time.Time { return fixed }
	require.NoError(t, first.ObserveStage(ModelStageOCR, StageResultSuccess, time.Millisecond))
	second, err := NewOperationalMetrics(registry)
	require.NoError(t, err)
	require.Equal(t, float64(fixed.Unix()), testutil.ToFloat64(second.stageLastSuccess.WithLabelValues(string(ModelStageOCR))))
	require.NoError(t, second.ObserveStage(ModelStageLiveness, StageResultError, time.Second, ReasonTimeout))
	require.Equal(t, float64(1), testutil.ToFloat64(first.stageOperations.WithLabelValues(string(ModelStageLiveness), string(StageResultError))))

	firstDigest := profileDigest("first")
	secondDigest := profileDigest("second")
	require.NoError(t, first.SetProfileDigest(firstDigest))
	require.NoError(t, second.SetProfileDigest(firstDigest))
	require.EqualError(t, second.SetProfileDigest(secondDigest), "operational profile digest is immutable for collector lifetime")
	families, err := registry.Gather()
	require.NoError(t, err)
	profileSeries := 0
	for _, family := range families {
		if family.GetName() != "virtengine_inference_model_profile_info" {
			continue
		}
		profileSeries += len(family.Metric)
		require.Equal(t, firstDigest, family.Metric[0].Label[0].GetValue())
	}
	require.Equal(t, 1, profileSeries)
}

func TestOperationalMetricsProfileDigestValidation(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics, err := NewOperationalMetrics(registry)
	require.NoError(t, err)
	for _, invalid := range []string{"", "ABCDEF", profileDigest("upper")[:63], "account-secret-not-a-digest"} {
		err := metrics.SetProfileDigest(invalid)
		require.EqualError(t, err, "profile digest must be canonical lowercase SHA-256 hex")
		if invalid != "" {
			require.NotContains(t, err.Error(), invalid)
		}
	}
	require.NoError(t, metrics.SetProfileDigest(profileDigest("valid")))
}

func TestOperationalMetricsConcurrentObservation(t *testing.T) {
	metrics, err := NewOperationalMetrics(prometheus.NewRegistry())
	require.NoError(t, err)
	var wg sync.WaitGroup
	errors := make(chan error, 96)
	for index := 0; index < 32; index++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errors <- metrics.ObserveStage(ModelStageFaceComparison, StageResultSuccess, time.Millisecond)
			errors <- metrics.ObserveDecision(DecisionFaceMatch, DecisionOutcomePass, time.Millisecond)
			errors <- metrics.SetProfileDigest(profileDigest("concurrent-profile"))
		}()
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
	require.Equal(t, float64(32), testutil.ToFloat64(metrics.stageOperations.WithLabelValues(string(ModelStageFaceComparison), string(StageResultSuccess))))
}

func TestOperationalMetricsFreshnessIsMonotonic(t *testing.T) {
	metrics, err := NewOperationalMetrics(prometheus.NewRegistry())
	require.NoError(t, err)
	newer := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	older := newer.Add(-time.Hour)
	metrics.now = func() time.Time { return newer }
	require.NoError(t, metrics.ObserveStage(ModelStageOCR, StageResultSuccess, time.Millisecond))
	require.NoError(t, metrics.ObserveDecision(DecisionIdentity, DecisionOutcomePass, time.Millisecond))
	metrics.now = func() time.Time { return older }
	require.NoError(t, metrics.ObserveStage(ModelStageOCR, StageResultSuccess, time.Millisecond))
	require.NoError(t, metrics.ObserveDecision(DecisionIdentity, DecisionOutcomePass, time.Millisecond))
	require.Equal(t, float64(newer.Unix()), testutil.ToFloat64(metrics.stageLastSuccess.WithLabelValues(string(ModelStageOCR))))
	require.Equal(t, float64(newer.Unix()), testutil.ToFloat64(metrics.decisionLastSuccess.WithLabelValues(string(DecisionIdentity))))
}

func TestOperationalMetricsDescriptorContract(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics, err := NewOperationalMetrics(registry)
	require.NoError(t, err)
	require.NoError(t, metrics.ObserveStage(ModelStageOCR, StageResultSuccess, time.Millisecond))
	require.NoError(t, metrics.ObserveDecision(DecisionIdentity, DecisionOutcomeReject, time.Millisecond, ReasonPolicyDenied))
	families, err := registry.Gather()
	require.NoError(t, err)
	names := make(map[string]bool, len(families))
	for _, family := range families {
		names[family.GetName()] = true
	}
	for _, name := range []string{
		"virtengine_inference_model_stage_operations_total",
		"virtengine_inference_model_stage_latency_seconds",
		"virtengine_inference_model_stage_last_success_timestamp_seconds",
		"virtengine_inference_model_stage_reasons_total",
		"virtengine_inference_model_decisions_total",
		"virtengine_inference_model_decision_latency_seconds",
		"virtengine_inference_model_decision_last_success_timestamp_seconds",
		"virtengine_inference_model_decision_reasons_total",
	} {
		require.True(t, names[name], name)
	}
	requireMetricLabelNames(t, families, "virtengine_inference_model_stage_operations_total", []string{"result", "stage"})
	requireMetricLabelNames(t, families, "virtengine_inference_model_stage_latency_seconds", []string{"result", "stage"})
	requireMetricLabelNames(t, families, "virtengine_inference_model_stage_last_success_timestamp_seconds", []string{"stage"})
	requireMetricLabelNames(t, families, "virtengine_inference_model_stage_reasons_total", []string{"reason_code", "stage"})
	requireMetricLabelNames(t, families, "virtengine_inference_model_decisions_total", []string{"decision", "outcome"})
	requireMetricLabelNames(t, families, "virtengine_inference_model_decision_latency_seconds", []string{"decision", "outcome"})
	requireMetricLabelNames(t, families, "virtengine_inference_model_decision_last_success_timestamp_seconds", []string{"decision"})
	requireMetricLabelNames(t, families, "virtengine_inference_model_decision_reasons_total", []string{"decision", "reason_code"})
	requireMetricSeriesCount(t, families, "virtengine_inference_model_stage_operations_total", 24)
	requireMetricSeriesCount(t, families, "virtengine_inference_model_stage_latency_seconds", 24)
	requireMetricSeriesCount(t, families, "virtengine_inference_model_stage_last_success_timestamp_seconds", 8)
	requireMetricSeriesCount(t, families, "virtengine_inference_model_stage_reasons_total", 88)
	requireMetricSeriesCount(t, families, "virtengine_inference_model_decisions_total", 35)
	requireMetricSeriesCount(t, families, "virtengine_inference_model_decision_latency_seconds", 35)
	requireMetricSeriesCount(t, families, "virtengine_inference_model_decision_last_success_timestamp_seconds", 7)
	requireMetricSeriesCount(t, families, "virtengine_inference_model_decision_reasons_total", 77)
}

func TestOperationalMetricsRejectsConflictingRegistration(t *testing.T) {
	registry := prometheus.NewRegistry()
	require.NoError(t, registry.Register(prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "virtengine", Subsystem: "inference_model", Name: "stage_operations_total",
		Help: "Conflicting help text.",
	}, []string{"stage", "result"})))
	_, err := NewOperationalMetrics(registry)
	require.Error(t, err)
}

func TestOperationalMetricsRollsBackLateRegistrationConflict(t *testing.T) {
	registry := prometheus.NewRegistry()
	require.NoError(t, registry.Register(prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "virtengine_inference_model_profile_info",
		Help: "Current governed operational profile digest.",
	}, []string{"profile_digest"})))
	_, err := NewOperationalMetrics(registry)
	require.EqualError(t, err, "registered operational collector has unexpected type")

	// Registration succeeds only if the failed constructor removed its earlier collectors.
	require.NoError(t, registry.Register(prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "virtengine", Subsystem: "inference_model", Name: "stage_operations_total",
		Help: "Model stage operations by bounded stage and result.",
	}, []string{"stage", "result"})))
}

func profileDigest(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func requireMetricLabelNames(t *testing.T, families []*dto.MetricFamily, name string, expected []string) {
	t.Helper()
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		require.NotEmpty(t, family.Metric)
		actual := make([]string, 0, len(family.Metric[0].Label))
		for _, label := range family.Metric[0].Label {
			actual = append(actual, label.GetName())
		}
		require.Equal(t, expected, actual)
		return
	}
	t.Fatalf("metric family %s not found", name)
}

func requireMetricSeriesCount(t *testing.T, families []*dto.MetricFamily, name string, expected int) {
	t.Helper()
	for _, family := range families {
		if family.GetName() == name {
			require.Len(t, family.Metric, expected)
			return
		}
	}
	t.Fatalf("metric family %s not found", name)
}
