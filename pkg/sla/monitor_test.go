/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package sla

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestEvaluateMetric(t *testing.T) {
	monitor := NewMonitor(nil, nil, nil, nil, WithWarningBuffer(decimal.NewFromFloat(0.05)), WithNowFn(func() time.Time {
		return time.Unix(0, 0)
	}))

	t.Run("gte", func(t *testing.T) {
		metric := SLAMetric{
			Type:       MetricUptime,
			Target:     decimal.NewFromInt(100),
			Comparison: ComparisonGTE,
		}

		cases := []struct {
			name   string
			value  decimal.Decimal
			status HealthStatus
		}{
			{name: "healthy", value: decimal.NewFromInt(110), status: StatusHealthy},
			{name: "warning", value: decimal.NewFromInt(102), status: StatusWarning},
			{name: "breach", value: decimal.NewFromInt(99), status: StatusBreached},
		}

		for _, tc := range cases {
			if got := monitor.evaluateMetric(metric, tc.value); got.Status != tc.status {
				t.Fatalf("%s: expected %s, got %s", tc.name, tc.status, got.Status)
			}
		}
	})

	t.Run("lte", func(t *testing.T) {
		metric := SLAMetric{
			Type:       MetricLatencyP99,
			Target:     decimal.NewFromInt(500),
			Comparison: ComparisonLTE,
		}

		cases := []struct {
			name   string
			value  decimal.Decimal
			status HealthStatus
		}{
			{name: "healthy", value: decimal.NewFromInt(470), status: StatusHealthy},
			{name: "warning", value: decimal.NewFromInt(480), status: StatusWarning},
			{name: "breach", value: decimal.NewFromInt(510), status: StatusBreached},
		}

		for _, tc := range cases {
			if got := monitor.evaluateMetric(metric, tc.value); got.Status != tc.status {
				t.Fatalf("%s: expected %s, got %s", tc.name, tc.status, got.Status)
			}
		}
	})
}
