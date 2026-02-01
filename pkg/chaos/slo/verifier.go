// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

// Package slo provides SLO (Service Level Objective) verification for chaos experiments.
// It implements steady-state hypothesis checking using various probe types.
package slo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/virtengine/virtengine/pkg/chaos"
)

// Verifier implements the chaos.SLOVerifier interface for verifying
// SLO compliance during chaos experiments.
type Verifier struct {
	// httpClient is used for HTTP probes
	httpClient *http.Client

	// prometheusURL is the base URL for Prometheus queries
	prometheusURL string

	// tolerance is the default tolerance for SLO violations (0.0-1.0)
	tolerance float64
}

// VerifierOption configures the Verifier.
type VerifierOption func(*Verifier)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) VerifierOption {
	return func(v *Verifier) {
		v.httpClient = client
	}
}

// WithPrometheusURL sets the Prometheus URL for metric queries.
func WithPrometheusURL(url string) VerifierOption {
	return func(v *Verifier) {
		v.prometheusURL = url
	}
}

// WithTolerance sets the default tolerance for SLO checks.
func WithTolerance(tolerance float64) VerifierOption {
	return func(v *Verifier) {
		v.tolerance = tolerance
	}
}

// NewVerifier creates a new SLO verifier with the given options.
func NewVerifier(opts ...VerifierOption) *Verifier {
	v := &Verifier{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		prometheusURL: "http://localhost:9090",
		tolerance:     0.05, // 5% default tolerance
	}

	for _, opt := range opts {
		opt(v)
	}

	return v
}

// Verify checks if the system meets SLO requirements during an experiment.
func (v *Verifier) Verify(ctx context.Context, experiment *chaos.ExperimentSpec) (bool, []chaos.Violation, error) {
	if experiment == nil {
		return false, nil, fmt.Errorf("experiment is nil")
	}

	if experiment.SteadyStateHypothesis == nil {
		return true, nil, nil
	}

	violations, err := v.VerifySteadyState(ctx, experiment.SteadyStateHypothesis)
	if err != nil {
		return false, nil, err
	}

	// Convert SteadyStateViolations to Violations
	chaosViolations := make([]chaos.Violation, 0, len(violations))
	for _, ssv := range violations {
		chaosViolations = append(chaosViolations, chaos.Violation{
			SLOName:       ssv.ProbeName,
			Timestamp:     ssv.Timestamp,
			ExpectedValue: ssv.Deviation,
			ActualValue:   ssv.Deviation,
			Severity:      "warning",
			Message:       ssv.Message,
			ExperimentID:  experiment.ID,
		})
	}

	return len(chaosViolations) == 0, chaosViolations, nil
}

// VerifySteadyState verifies all probes in a steady state hypothesis.
func (v *Verifier) VerifySteadyState(ctx context.Context, hypothesis *chaos.SteadyStateHypothesis) ([]chaos.SteadyStateViolation, error) {
	if hypothesis == nil {
		return nil, nil
	}

	var violations []chaos.SteadyStateViolation
	tolerance := hypothesis.Tolerance
	if tolerance <= 0 {
		tolerance = v.tolerance
	}

	for _, probe := range hypothesis.Probes {
		value, err := v.CheckProbe(ctx, probe)
		if err != nil {
			violations = append(violations, chaos.SteadyStateViolation{
				Timestamp:     time.Now(),
				ProbeName:     probe.Name,
				ExpectedValue: probe.ExpectedValue,
				ActualValue:   err.Error(),
				Message:       fmt.Sprintf("probe check failed: %v", err),
			})
			continue
		}

		// Check if value meets success criteria
		if !v.meetsSuccessCriteria(probe, value, tolerance) {
			violations = append(violations, chaos.SteadyStateViolation{
				Timestamp:     time.Now(),
				ProbeName:     probe.Name,
				ExpectedValue: probe.ExpectedValue,
				ActualValue:   value,
				Deviation:     v.calculateDeviation(probe.ExpectedValue, value),
				Tolerance:     tolerance,
				Message:       fmt.Sprintf("probe %s value %.2f does not meet success criteria: %s", probe.Name, value, probe.SuccessCriteria),
			})
		}
	}

	return violations, nil
}

// CheckProbe executes a single probe and returns its value.
func (v *Verifier) CheckProbe(ctx context.Context, probe chaos.Probe) (float64, error) {
	// Apply initial delay if configured
	if probe.InitialDelay > 0 {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(probe.InitialDelay):
		}
	}

	// Create a context with timeout
	probeCtx := ctx
	if probe.Timeout > 0 {
		var cancel context.CancelFunc
		probeCtx, cancel = context.WithTimeout(ctx, probe.Timeout)
		defer cancel()
	}

	switch probe.Type {
	case chaos.ProbeTypeHTTP:
		return v.checkHTTPProbe(probeCtx, probe)
	case chaos.ProbeTypePrometheus:
		return v.checkPrometheusProbe(probeCtx, probe)
	case chaos.ProbeTypeCommand:
		return v.checkCommandProbe(probeCtx, probe)
	case chaos.ProbeTypeGRPC:
		return v.checkGRPCProbe(probeCtx, probe)
	case chaos.ProbeTypeKubernetes:
		return v.checkKubernetesProbe(probeCtx, probe)
	default:
		return 0, fmt.Errorf("unsupported probe type: %s", probe.Type)
	}
}

// checkHTTPProbe executes an HTTP probe.
func (v *Verifier) checkHTTPProbe(ctx context.Context, probe chaos.Probe) (float64, error) {
	method := probe.Method
	if method == "" {
		method = http.MethodGet
	}

	req, err := http.NewRequestWithContext(ctx, method, probe.URL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range probe.Headers {
		req.Header.Set(key, value)
	}

	start := time.Now()
	resp, err := v.httpClient.Do(req)
	latency := time.Since(start).Seconds()

	if err != nil {
		return 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// If success criteria is based on status code
	if strings.Contains(probe.SuccessCriteria, "status") {
		return float64(resp.StatusCode), nil
	}

	// If success criteria is based on latency
	if strings.Contains(probe.SuccessCriteria, "latency") {
		return latency, nil
	}

	// Default: return status code
	return float64(resp.StatusCode), nil
}

// checkPrometheusProbe executes a Prometheus query probe.
func (v *Verifier) checkPrometheusProbe(ctx context.Context, probe chaos.Probe) (float64, error) {
	url := fmt.Sprintf("%s/api/v1/query?query=%s", v.prometheusURL, probe.Query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create Prometheus request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("prometheus query failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read Prometheus response: %w", err)
	}

	// Parse Prometheus response
	var promResp prometheusResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		return 0, fmt.Errorf("failed to parse Prometheus response: %w", err)
	}

	if promResp.Status != "success" {
		return 0, fmt.Errorf("prometheus query error: %s", promResp.Status)
	}

	// Extract value from result
	if len(promResp.Data.Result) == 0 {
		return 0, fmt.Errorf("no data returned from Prometheus query")
	}

	if len(promResp.Data.Result[0].Value) < 2 {
		return 0, fmt.Errorf("invalid Prometheus result format")
	}

	valueStr, ok := promResp.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, fmt.Errorf("unexpected value type in Prometheus result")
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse Prometheus value: %w", err)
	}

	return value, nil
}

// prometheusResponse represents the Prometheus API response structure.
type prometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// checkCommandProbe executes a command probe.
func (v *Verifier) checkCommandProbe(ctx context.Context, probe chaos.Probe) (float64, error) {
	//nolint:gosec // G204: probe.Command is validated SLO configuration from trusted source
	cmd := exec.CommandContext(ctx, "sh", "-c", probe.Command)
	output, err := cmd.Output()
	if err != nil {
		// If the command failed, return 0 (failure)
		return 0, fmt.Errorf("command failed: %w", err)
	}

	// Try to parse output as a number
	value, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		// If output is not a number, return 1 for success (command completed)
		return 1, nil
	}

	return value, nil
}

// checkGRPCProbe executes a gRPC health check probe.
func (v *Verifier) checkGRPCProbe(_ context.Context, probe chaos.Probe) (float64, error) {
	// Simplified gRPC check - in production, use grpc-health-probe or similar
	// For now, return error indicating not implemented
	return 0, fmt.Errorf("gRPC probe not implemented: %s", probe.URL)
}

// checkKubernetesProbe checks Kubernetes resource status.
func (v *Verifier) checkKubernetesProbe(ctx context.Context, probe chaos.Probe) (float64, error) {
	// Use kubectl to check resource status
	//nolint:gosec // G204: probe.Query is validated resource name from trusted SLO config
	cmd := exec.CommandContext(ctx, "kubectl", "get", probe.Query, "-o", "jsonpath={.status.phase}")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("kubectl failed: %w", err)
	}

	status := strings.TrimSpace(string(output))
	if status == "Running" || status == "Succeeded" || status == "Active" {
		return 1, nil
	}

	return 0, nil
}

// meetsSuccessCriteria checks if a probe value meets the success criteria.
func (v *Verifier) meetsSuccessCriteria(probe chaos.Probe, value float64, tolerance float64) bool {
	criteria := probe.SuccessCriteria
	if criteria == "" {
		return true
	}

	// Parse simple criteria like "status == 200", "value > 0.9", "latency < 1.0"
	parts := strings.Fields(criteria)
	if len(parts) < 3 {
		return true // Invalid criteria, assume success
	}

	operator := parts[1]
	threshold, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return true // Invalid threshold, assume success
	}

	switch operator {
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "~=", "~": // Approximate equality with tolerance
		return (value >= threshold*(1-tolerance)) && (value <= threshold*(1+tolerance))
	default:
		return true
	}
}

// calculateDeviation calculates the percentage deviation from expected value.
func (v *Verifier) calculateDeviation(expected interface{}, actual float64) float64 {
	expectedFloat, ok := expected.(float64)
	if !ok {
		// Try to parse as string
		if s, ok := expected.(string); ok {
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				expectedFloat = f
			} else {
				return 0
			}
		} else {
			return 0
		}
	}

	if expectedFloat == 0 {
		return actual * 100
	}

	return ((actual - expectedFloat) / expectedFloat) * 100
}
