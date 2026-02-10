/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package sla

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculateCredit(t *testing.T) {
	uptimeDefinition := &SLADefinition{
		ProviderID: "provider-1",
		Metrics: []SLAMetric{
			{
				Type:       MetricUptime,
				Target:     decimal.NewFromFloat(99.9),
				Comparison: ComparisonGTE,
			},
		},
		CreditPolicy: CreditPolicy{
			Brackets: []CreditBracket{
				{
					MinDeviation:  decimal.NewFromFloat(0.1),
					MaxDeviation:  decimal.NewFromFloat(0.5),
					CreditPercent: decimal.NewFromInt(10),
				},
				{
					MinDeviation:  decimal.NewFromFloat(0.5),
					MaxDeviation:  decimal.NewFromFloat(2.0),
					CreditPercent: decimal.NewFromInt(20),
				},
			},
			MaxCredit: decimal.NewFromInt(25),
		},
	}

	latencyDefinition := &SLADefinition{
		ProviderID: "provider-1",
		Metrics: []SLAMetric{
			{
				Type:       MetricLatencyP99,
				Target:     decimal.NewFromInt(500),
				Comparison: ComparisonLTE,
			},
		},
		CreditPolicy: CreditPolicy{
			Brackets: []CreditBracket{
				{
					MinDeviation:  decimal.NewFromInt(10),
					MaxDeviation:  decimal.NewFromInt(50),
					CreditPercent: decimal.NewFromInt(10),
				},
			},
		},
	}

	type testCase struct {
		name       string
		definition *SLADefinition
		breach     *BreachRecord
		want       decimal.Decimal
	}

	cases := []testCase{
		{
			name:       "uptime mild deviation",
			definition: uptimeDefinition,
			breach: &BreachRecord{
				MetricType:  MetricUptime,
				ActualValue: decimal.NewFromFloat(99.7),
			},
			want: decimal.NewFromInt(10),
		},
		{
			name:       "uptime larger deviation",
			definition: uptimeDefinition,
			breach: &BreachRecord{
				MetricType:  MetricUptime,
				ActualValue: decimal.NewFromFloat(98.2),
			},
			want: decimal.NewFromInt(20),
		},
		{
			name:       "latency deviation",
			definition: latencyDefinition,
			breach: &BreachRecord{
				MetricType:  MetricLatencyP99,
				ActualValue: decimal.NewFromInt(510),
			},
			want: decimal.NewFromInt(10),
		},
		{
			name:       "no matching bracket",
			definition: uptimeDefinition,
			breach: &BreachRecord{
				MetricType:  MetricUptime,
				ActualValue: decimal.NewFromFloat(96.0),
			},
			want: decimal.Zero,
		},
	}

	for _, tc := range cases {
		credit, err := CalculateCredit(tc.definition, tc.breach)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}

		if !credit.Equal(tc.want) {
			t.Fatalf("%s: expected %s, got %s", tc.name, tc.want, credit)
		}
	}
}
