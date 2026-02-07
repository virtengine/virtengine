// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"sort"
	"sync"
	"time"
)

// MetricsCollector collects and aggregates execution metrics
type MetricsCollector struct {
	mu        sync.Mutex
	latencies []time.Duration
	successes int64
	failures  int64
	errors    map[string]int
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		latencies: make([]time.Duration, 0, 10000),
		errors:    make(map[string]int),
	}
}

// Record records a single execution result
func (mc *MetricsCollector) Record(result *ExecutionResult) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.latencies = append(mc.latencies, result.Duration)

	if result.Success {
		mc.successes++
	} else {
		mc.failures++
		if result.Error != nil {
			errMsg := result.Error.Error()
			mc.errors[errMsg]++
		}
	}
}

// GenerateReport creates a test report from collected metrics
func (mc *MetricsCollector) GenerateReport(name string, duration time.Duration) *TestReport {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	totalRequests := mc.successes + mc.failures
	if totalRequests == 0 {
		return &TestReport{
			Name:     name,
			Duration: duration,
			Errors:   mc.errors,
		}
	}

	sort.Slice(mc.latencies, func(i, j int) bool {
		return mc.latencies[i] < mc.latencies[j]
	})

	var avgLatency time.Duration
	for _, lat := range mc.latencies {
		avgLatency += lat
	}
	avgLatency /= time.Duration(len(mc.latencies))

	p50 := mc.latencies[len(mc.latencies)*50/100]
	p95 := mc.latencies[len(mc.latencies)*95/100]
	p99 := mc.latencies[len(mc.latencies)*99/100]
	maxLatency := mc.latencies[len(mc.latencies)-1]

	rps := float64(totalRequests) / duration.Seconds()
	errorRate := float64(mc.failures) / float64(totalRequests) * 100

	return &TestReport{
		Name:           name,
		Duration:       duration,
		TotalRequests:  totalRequests,
		SuccessCount:   mc.successes,
		FailureCount:   mc.failures,
		AvgLatency:     avgLatency,
		P50Latency:     p50,
		P95Latency:     p95,
		P99Latency:     p99,
		MaxLatency:     maxLatency,
		RequestsPerSec: rps,
		ErrorRate:      errorRate,
		Errors:         mc.errors,
	}
}
