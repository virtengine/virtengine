package analysis

import (
	"fmt"

	"github.com/virtengine/virtengine/sim/core"
)

// Threshold defines acceptable bounds for a metric.
type Threshold struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// Violation describes a metric outside the threshold.
type Violation struct {
	Metric string  `json:"metric"`
	Value  float64 `json:"value"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
}

// DefaultThresholds provides baseline bounds for economics metrics.
func DefaultThresholds() map[string]Threshold {
	return map[string]Threshold{
		"avg_gas_price":       {Min: 0.0004, Max: 0.01},
		"avg_min_gas_price":   {Min: 0.0004, Max: 0.008},
		"avg_gas_utilization": {Min: 0.05, Max: 0.95},
		"avg_inflation_bps":   {Min: 200, Max: 1500},
		"avg_staking_bps":     {Min: 3500, Max: 9000},
		"avg_apr_bps":         {Min: 200, Max: 20000},
	}
}

// CheckThresholds returns violations for metrics outside bounds.
func CheckThresholds(metrics core.Metrics, thresholds map[string]Threshold) []Violation {
	values := map[string]float64{
		"avg_gas_price":       metrics.AvgGasPrice,
		"avg_min_gas_price":   metrics.AvgMinGasPrice,
		"avg_gas_utilization": metrics.AvgGasUtilization,
		"avg_inflation_bps":   float64(metrics.AverageInflationBPS),
		"avg_staking_bps":     float64(metrics.AverageStakingBPS),
		"avg_apr_bps":         float64(metrics.AverageAPR),
	}

	violations := make([]Violation, 0)
	for name, threshold := range thresholds {
		value, ok := values[name]
		if !ok {
			continue
		}
		if (threshold.Min > 0 && value < threshold.Min) || (threshold.Max > 0 && value > threshold.Max) {
			violations = append(violations, Violation{
				Metric: name,
				Value:  value,
				Min:    threshold.Min,
				Max:    threshold.Max,
			})
		}
	}
	return violations
}

// ValidateThresholds ensures thresholds are sane.
func ValidateThresholds(thresholds map[string]Threshold) error {
	for name, t := range thresholds {
		if t.Min < 0 || t.Max < 0 {
			return fmt.Errorf("thresholds must be non-negative: %s", name)
		}
		if t.Max > 0 && t.Min > t.Max {
			return fmt.Errorf("threshold min greater than max: %s", name)
		}
	}
	return nil
}
