/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package sla

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// CalculateCredit computes the credit percentage owed for a breach.
func CalculateCredit(definition *SLADefinition, breach *BreachRecord) (decimal.Decimal, error) {
	if definition == nil {
		return decimal.Zero, fmt.Errorf("sla definition is nil")
	}
	if breach == nil {
		return decimal.Zero, fmt.Errorf("breach is nil")
	}

	metric, ok := findMetric(definition.Metrics, breach.MetricType)
	if !ok {
		return decimal.Zero, fmt.Errorf("sla metric %s not found", breach.MetricType)
	}

	deviation, err := calculateDeviation(metric, breach.ActualValue)
	if err != nil {
		return decimal.Zero, err
	}
	if deviation.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, nil
	}

	credit := decimal.Zero
	for _, bracket := range definition.CreditPolicy.Brackets {
		if deviation.GreaterThanOrEqual(bracket.MinDeviation) {
			if bracket.MaxDeviation.IsZero() || deviation.LessThanOrEqual(bracket.MaxDeviation) {
				credit = bracket.CreditPercent
				break
			}
		}
	}

	if !definition.CreditPolicy.MaxCredit.IsZero() && credit.GreaterThan(definition.CreditPolicy.MaxCredit) {
		credit = definition.CreditPolicy.MaxCredit
	}

	return credit, nil
}

func calculateDeviation(metric SLAMetric, actual decimal.Decimal) (decimal.Decimal, error) {
	switch metric.Comparison {
	case ComparisonGTE:
		if actual.GreaterThanOrEqual(metric.Target) {
			return decimal.Zero, nil
		}
		return metric.Target.Sub(actual), nil
	case ComparisonLTE:
		if actual.LessThanOrEqual(metric.Target) {
			return decimal.Zero, nil
		}
		return actual.Sub(metric.Target), nil
	default:
		return decimal.Zero, fmt.Errorf("unsupported comparison type: %s", metric.Comparison)
	}
}

func findMetric(metrics []SLAMetric, metricType SLAMetricType) (SLAMetric, bool) {
	for _, metric := range metrics {
		if metric.Type == metricType {
			return metric, true
		}
	}

	return SLAMetric{}, false
}
