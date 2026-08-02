// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type ModelStage string
type StageResult string
type DecisionKind string
type DecisionOutcome string
type OperationalReasonCode string

const (
	ModelStageDocumentRouting      ModelStage = "document_routing"
	ModelStageOCR                  ModelStage = "ocr"
	ModelStageDocumentAuthenticity ModelStage = "document_authenticity"
	ModelStageFaceDetection        ModelStage = "face_detection"
	ModelStageFaceComparison       ModelStage = "face_comparison"
	ModelStageLiveness             ModelStage = "liveness"
	ModelStageFraudFusion          ModelStage = "fraud_fusion"
	ModelStagePolicyEvaluation     ModelStage = "policy_evaluation"

	StageResultSuccess     StageResult = "success"
	StageResultUnavailable StageResult = "unavailable"
	StageResultError       StageResult = "error"

	DecisionDocument    DecisionKind = "document"
	DecisionFaceMatch   DecisionKind = "face_match"
	DecisionLiveness    DecisionKind = "liveness"
	DecisionRisk        DecisionKind = "risk"
	DecisionIdentity    DecisionKind = "identity"
	DecisionUniqueness  DecisionKind = "uniqueness"
	DecisionEligibility DecisionKind = "eligibility"

	DecisionOutcomePass        DecisionOutcome = "pass"
	DecisionOutcomeReview      DecisionOutcome = "review"
	DecisionOutcomeReject      DecisionOutcome = "reject"
	DecisionOutcomeUnavailable DecisionOutcome = "unavailable"
	DecisionOutcomeError       DecisionOutcome = "error"

	ReasonInvalidInput     OperationalReasonCode = "invalid_input"
	ReasonInputUnavailable OperationalReasonCode = "input_unavailable"
	ReasonModelUnavailable OperationalReasonCode = "model_unavailable"
	ReasonModelError       OperationalReasonCode = "model_error"
	ReasonTimeout          OperationalReasonCode = "timeout"
	ReasonLowConfidence    OperationalReasonCode = "low_confidence"
	ReasonThresholdNotMet  OperationalReasonCode = "threshold_not_met"
	ReasonHardGateFailed   OperationalReasonCode = "hard_gate_failed"
	ReasonReviewRequired   OperationalReasonCode = "review_required"
	ReasonPolicyDenied     OperationalReasonCode = "policy_denied"
	ReasonUnknown          OperationalReasonCode = "unknown"
)

var (
	operationalFreshnessMu sync.Mutex

	allowedModelStages = map[ModelStage]struct{}{
		ModelStageDocumentRouting: {}, ModelStageOCR: {}, ModelStageDocumentAuthenticity: {},
		ModelStageFaceDetection: {}, ModelStageFaceComparison: {}, ModelStageLiveness: {},
		ModelStageFraudFusion: {}, ModelStagePolicyEvaluation: {},
	}
	allowedStageResults  = map[StageResult]struct{}{StageResultSuccess: {}, StageResultUnavailable: {}, StageResultError: {}}
	allowedDecisionKinds = map[DecisionKind]struct{}{
		DecisionDocument: {}, DecisionFaceMatch: {}, DecisionLiveness: {}, DecisionRisk: {},
		DecisionIdentity: {}, DecisionUniqueness: {}, DecisionEligibility: {},
	}
	allowedDecisionOutcomes = map[DecisionOutcome]struct{}{
		DecisionOutcomePass: {}, DecisionOutcomeReview: {}, DecisionOutcomeReject: {},
		DecisionOutcomeUnavailable: {}, DecisionOutcomeError: {},
	}
	allowedOperationalReasons = map[OperationalReasonCode]struct{}{
		ReasonInvalidInput: {}, ReasonInputUnavailable: {}, ReasonModelUnavailable: {}, ReasonModelError: {},
		ReasonTimeout: {}, ReasonLowConfidence: {}, ReasonThresholdNotMet: {}, ReasonHardGateFailed: {},
		ReasonReviewRequired: {}, ReasonPolicyDenied: {}, ReasonUnknown: {},
	}
)

type OperationalMetrics struct {
	stageOperations     *prometheus.CounterVec
	stageLatency        *prometheus.HistogramVec
	stageLastSuccess    *prometheus.GaugeVec
	stageReasons        *prometheus.CounterVec
	decisionOperations  *prometheus.CounterVec
	decisionLatency     *prometheus.HistogramVec
	decisionLastSuccess *prometheus.GaugeVec
	decisionReasons     *prometheus.CounterVec
	profile             *operationalProfileCollector
	now                 func() time.Time
}

func NewOperationalMetrics(registerer prometheus.Registerer) (*OperationalMetrics, error) {
	if registerer == nil {
		return nil, errors.New("operational metrics registerer is required")
	}
	registered := make([]prometheus.Collector, 0, 9)
	rollback := func() {
		for index := len(registered) - 1; index >= 0; index-- {
			registerer.Unregister(registered[index])
		}
	}
	stageOperations, added, err := registerOperationalCollector(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "virtengine", Subsystem: "inference_model", Name: "stage_operations_total",
		Help: "Model stage operations by bounded stage and result.",
	}, []string{"stage", "result"}))
	if err != nil {
		return nil, err
	}
	if added {
		registered = append(registered, stageOperations)
	}
	stageLatency, added, err := registerOperationalCollector(registerer, prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "virtengine", Subsystem: "inference_model", Name: "stage_latency_seconds",
		Help: "Model stage latency by bounded stage and result.", Buckets: operationalLatencyBuckets(),
	}, []string{"stage", "result"}))
	if err != nil {
		rollback()
		return nil, err
	}
	if added {
		registered = append(registered, stageLatency)
	}
	stageLastSuccess, added, err := registerOperationalCollector(registerer, prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "virtengine", Subsystem: "inference_model", Name: "stage_last_success_timestamp_seconds",
		Help: "Unix timestamp of the latest successful model stage operation.",
	}, []string{"stage"}))
	if err != nil {
		rollback()
		return nil, err
	}
	initializeStageFreshness := added
	if added {
		registered = append(registered, stageLastSuccess)
	}
	stageReasons, added, err := registerOperationalCollector(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "virtengine", Subsystem: "inference_model", Name: "stage_reasons_total",
		Help: "Model stage operational reasons by bounded stage and reason code.",
	}, []string{"stage", "reason_code"}))
	if err != nil {
		rollback()
		return nil, err
	}
	if added {
		registered = append(registered, stageReasons)
	}
	decisionOperations, added, err := registerOperationalCollector(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "virtengine", Subsystem: "inference_model", Name: "decisions_total",
		Help: "Model decisions by bounded decision and outcome.",
	}, []string{"decision", "outcome"}))
	if err != nil {
		rollback()
		return nil, err
	}
	if added {
		registered = append(registered, decisionOperations)
	}
	decisionLatency, added, err := registerOperationalCollector(registerer, prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "virtengine", Subsystem: "inference_model", Name: "decision_latency_seconds",
		Help: "Model decision latency by bounded decision and outcome.", Buckets: operationalLatencyBuckets(),
	}, []string{"decision", "outcome"}))
	if err != nil {
		rollback()
		return nil, err
	}
	if added {
		registered = append(registered, decisionLatency)
	}
	decisionLastSuccess, added, err := registerOperationalCollector(registerer, prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "virtengine", Subsystem: "inference_model", Name: "decision_last_success_timestamp_seconds",
		Help: "Unix timestamp of the latest available model decision.",
	}, []string{"decision"}))
	if err != nil {
		rollback()
		return nil, err
	}
	initializeDecisionFreshness := added
	if added {
		registered = append(registered, decisionLastSuccess)
	}
	decisionReasons, added, err := registerOperationalCollector(registerer, prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "virtengine", Subsystem: "inference_model", Name: "decision_reasons_total",
		Help: "Model decision operational reasons by bounded decision and reason code.",
	}, []string{"decision", "reason_code"}))
	if err != nil {
		rollback()
		return nil, err
	}
	if added {
		registered = append(registered, decisionReasons)
	}
	profile, added, err := registerOperationalCollector(registerer, newOperationalProfileCollector())
	if err != nil {
		rollback()
		return nil, err
	}
	if added {
		registered = append(registered, profile)
	}
	metrics := &OperationalMetrics{
		stageOperations: stageOperations, stageLatency: stageLatency, stageLastSuccess: stageLastSuccess, stageReasons: stageReasons,
		decisionOperations: decisionOperations, decisionLatency: decisionLatency, decisionLastSuccess: decisionLastSuccess,
		decisionReasons: decisionReasons, profile: profile, now: time.Now,
	}
	metrics.initializeBoundedSeries(initializeStageFreshness, initializeDecisionFreshness)
	return metrics, nil
}

func (m *OperationalMetrics) ObserveStage(stage ModelStage, result StageResult, duration time.Duration, reasons ...OperationalReasonCode) error {
	if _, ok := allowedModelStages[stage]; !ok {
		return errors.New("unsupported model stage")
	}
	if _, ok := allowedStageResults[result]; !ok {
		return errors.New("unsupported stage result")
	}
	uniqueReasons, err := validateOperationalReasons(reasons)
	if err != nil {
		return err
	}
	if err := validateStageReasons(result, uniqueReasons); err != nil {
		return err
	}
	if duration < 0 {
		return errors.New("stage duration cannot be negative")
	}
	m.stageOperations.WithLabelValues(string(stage), string(result)).Inc()
	m.stageLatency.WithLabelValues(string(stage), string(result)).Observe(duration.Seconds())
	if result == StageResultSuccess {
		setOperationalFreshness(m.stageLastSuccess.WithLabelValues(string(stage)), m.now().Unix())
	}
	for _, reason := range uniqueReasons {
		m.stageReasons.WithLabelValues(string(stage), string(reason)).Inc()
	}
	return nil
}

func (m *OperationalMetrics) ObserveDecision(decision DecisionKind, outcome DecisionOutcome, duration time.Duration, reasons ...OperationalReasonCode) error {
	if _, ok := allowedDecisionKinds[decision]; !ok {
		return errors.New("unsupported decision kind")
	}
	if _, ok := allowedDecisionOutcomes[outcome]; !ok {
		return errors.New("unsupported decision outcome")
	}
	uniqueReasons, err := validateOperationalReasons(reasons)
	if err != nil {
		return err
	}
	if err := validateDecisionReasons(outcome, uniqueReasons); err != nil {
		return err
	}
	if duration < 0 {
		return errors.New("decision duration cannot be negative")
	}
	m.decisionOperations.WithLabelValues(string(decision), string(outcome)).Inc()
	m.decisionLatency.WithLabelValues(string(decision), string(outcome)).Observe(duration.Seconds())
	if outcome == DecisionOutcomePass {
		setOperationalFreshness(m.decisionLastSuccess.WithLabelValues(string(decision)), m.now().Unix())
	}
	for _, reason := range uniqueReasons {
		m.decisionReasons.WithLabelValues(string(decision), string(reason)).Inc()
	}
	return nil
}

func (m *OperationalMetrics) SetProfileDigest(digest string) error {
	decoded, err := hex.DecodeString(digest)
	if err != nil || len(decoded) != 32 || hex.EncodeToString(decoded) != digest {
		return errors.New("profile digest must be canonical lowercase SHA-256 hex")
	}
	return m.profile.Set(digest)
}

func (m *OperationalMetrics) initializeBoundedSeries(initializeStageFreshness, initializeDecisionFreshness bool) {
	for stage := range allowedModelStages {
		if initializeStageFreshness {
			m.stageLastSuccess.WithLabelValues(string(stage)).Set(0)
		}
		for result := range allowedStageResults {
			m.stageOperations.WithLabelValues(string(stage), string(result))
			m.stageLatency.WithLabelValues(string(stage), string(result))
		}
		for reason := range allowedOperationalReasons {
			m.stageReasons.WithLabelValues(string(stage), string(reason))
		}
	}
	for decision := range allowedDecisionKinds {
		if initializeDecisionFreshness {
			m.decisionLastSuccess.WithLabelValues(string(decision)).Set(0)
		}
		for outcome := range allowedDecisionOutcomes {
			m.decisionOperations.WithLabelValues(string(decision), string(outcome))
			m.decisionLatency.WithLabelValues(string(decision), string(outcome))
		}
		for reason := range allowedOperationalReasons {
			m.decisionReasons.WithLabelValues(string(decision), string(reason))
		}
	}
}

func setOperationalFreshness(gauge prometheus.Gauge, timestamp int64) {
	operationalFreshnessMu.Lock()
	defer operationalFreshnessMu.Unlock()
	metric := &dto.Metric{}
	if err := gauge.Write(metric); err == nil && float64(timestamp) > metric.GetGauge().GetValue() {
		gauge.Set(float64(timestamp))
	}
}

func validateOperationalReasons(reasons []OperationalReasonCode) ([]OperationalReasonCode, error) {
	seen := make(map[OperationalReasonCode]struct{}, len(reasons))
	unique := make([]OperationalReasonCode, 0, len(reasons))
	for _, reason := range reasons {
		if _, ok := allowedOperationalReasons[reason]; !ok {
			return nil, errors.New("unsupported operational reason")
		}
		if _, ok := seen[reason]; ok {
			continue
		}
		seen[reason] = struct{}{}
		unique = append(unique, reason)
	}
	return unique, nil
}

func validateStageReasons(result StageResult, reasons []OperationalReasonCode) error {
	if result == StageResultSuccess {
		if len(reasons) != 0 {
			return errors.New("successful stage result cannot include an operational reason")
		}
		return nil
	}
	if len(reasons) == 0 {
		return errors.New("non-success stage result requires an operational reason")
	}
	for _, reason := range reasons {
		allowed := false
		switch result {
		case StageResultUnavailable:
			allowed = reason == ReasonInputUnavailable || reason == ReasonModelUnavailable || reason == ReasonTimeout || reason == ReasonUnknown
		case StageResultError:
			allowed = reason == ReasonInvalidInput || reason == ReasonModelError || reason == ReasonTimeout || reason == ReasonUnknown
		}
		if !allowed {
			return errors.New("operational reason is incompatible with stage result")
		}
	}
	return nil
}

func validateDecisionReasons(outcome DecisionOutcome, reasons []OperationalReasonCode) error {
	if outcome == DecisionOutcomePass {
		if len(reasons) != 0 {
			return errors.New("passing decision cannot include an operational reason")
		}
		return nil
	}
	if len(reasons) == 0 {
		return errors.New("non-pass decision outcome requires an operational reason")
	}
	for _, reason := range reasons {
		allowed := false
		switch outcome {
		case DecisionOutcomeReview:
			allowed = reason == ReasonLowConfidence || reason == ReasonReviewRequired || reason == ReasonUnknown
		case DecisionOutcomeReject:
			allowed = reason == ReasonThresholdNotMet || reason == ReasonHardGateFailed || reason == ReasonPolicyDenied || reason == ReasonUnknown
		case DecisionOutcomeUnavailable:
			allowed = reason == ReasonInputUnavailable || reason == ReasonModelUnavailable || reason == ReasonTimeout || reason == ReasonUnknown
		case DecisionOutcomeError:
			allowed = reason == ReasonInvalidInput || reason == ReasonModelError || reason == ReasonTimeout || reason == ReasonUnknown
		}
		if !allowed {
			return errors.New("operational reason is incompatible with decision outcome")
		}
	}
	return nil
}

func operationalLatencyBuckets() []float64 {
	return []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5}
}

type operationalProfileCollector struct {
	mu     sync.RWMutex
	desc   *prometheus.Desc
	digest string
}

func newOperationalProfileCollector() *operationalProfileCollector {
	return &operationalProfileCollector{desc: prometheus.NewDesc(
		"virtengine_inference_model_profile_info", "Current governed operational profile digest.", []string{"profile_digest"}, nil,
	)}
}
func (c *operationalProfileCollector) Set(digest string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.digest != "" && c.digest != digest {
		return errors.New("operational profile digest is immutable for collector lifetime")
	}
	c.digest = digest
	return nil
}
func (c *operationalProfileCollector) Describe(ch chan<- *prometheus.Desc) { ch <- c.desc }
func (c *operationalProfileCollector) Collect(ch chan<- prometheus.Metric) {
	c.mu.RLock()
	digest := c.digest
	c.mu.RUnlock()
	if digest != "" {
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, 1, digest)
	}
}

func registerOperationalCollector[T prometheus.Collector](registerer prometheus.Registerer, collector T) (T, bool, error) {
	if err := registerer.Register(collector); err != nil {
		already, ok := err.(prometheus.AlreadyRegisteredError)
		if !ok {
			var zero T
			return zero, false, err
		}
		existing, ok := already.ExistingCollector.(T)
		if !ok {
			var zero T
			return zero, false, fmt.Errorf("registered operational collector has unexpected type")
		}
		return existing, false, nil
	}
	return collector, true, nil
}
