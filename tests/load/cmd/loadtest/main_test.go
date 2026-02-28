// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/tests/load/framework"
)

func TestNewScenarioByNameSupportsAllDeclaredScenarios(t *testing.T) {
	accounts := defaultScenarioAccounts()

	for _, name := range []string{"veid_submit", "order_create", "bid_submit", "settlement"} {
		scenario, err := newScenarioByName(name, "mock://local", accounts)
		require.NoError(t, err, name)
		require.Equal(t, name, scenario.Name())
	}

	_, err := newScenarioByName("unknown", "mock://local", accounts)
	require.ErrorContains(t, err, "unknown scenario")
}

func TestAnalyzeRegressionFlagsThroughputAndLatency(t *testing.T) {
	baseline := &framework.TestReport{
		RequestsPerSec: 100,
		P95Latency:     100 * time.Millisecond,
		ErrorRate:      1,
	}
	current := &framework.TestReport{
		RequestsPerSec: 70,
		P95Latency:     140 * time.Millisecond,
		ErrorRate:      15,
	}

	lines := analyzeRegression(baseline, current, 10)
	require.True(t, hasRegression(lines))
	require.Contains(t, lines[len(lines)-3:], "REGRESSION: requests/sec dropped below 90.00")
	require.Contains(t, lines[len(lines)-2:], "REGRESSION: p95 latency exceeded 110ms")
	require.Contains(t, lines[len(lines)-1:], "REGRESSION: error rate exceeded 11.00%")
}

func TestReadReportRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.json")

	expected := &framework.TestReport{
		Name:           "contract",
		Duration:       2 * time.Second,
		TotalRequests:  42,
		SuccessCount:   42,
		FailureCount:   0,
		P95Latency:     75 * time.Millisecond,
		RequestsPerSec: 21,
		ErrorRate:      0,
	}

	require.NoError(t, expected.WriteJSON(path))

	report, err := readReport(path)
	require.NoError(t, err)
	require.Equal(t, expected.Name, report.Name)
	require.Equal(t, expected.TotalRequests, report.TotalRequests)
	require.Equal(t, expected.P95Latency, report.P95Latency)
}
