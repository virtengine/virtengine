package analysis

import (
	"context"
	"math"

	"github.com/virtengine/virtengine/sim/core"
)

// SensitivityConfig controls parameter sweep.
type SensitivityConfig struct {
	BaseConfig core.Config
	Steps      int
	Param      string
	Min        float64
	Max        float64
}

// SensitivityPoint captures a single sweep outcome.
type SensitivityPoint struct {
	ParamValue float64
	Metrics    core.Metrics
}

// SensitivityResult captures sweep results.
type SensitivityResult struct {
	Param   string
	Points  []SensitivityPoint
	Elastic float64
}

// RunSensitivity executes a parameter sweep.
func RunSensitivity(ctx context.Context, cfg SensitivityConfig) (SensitivityResult, error) {
	if cfg.Steps <= 1 {
		cfg.Steps = 5
	}
	points := make([]SensitivityPoint, 0, cfg.Steps)

	stepSize := (cfg.Max - cfg.Min) / float64(cfg.Steps-1)
	for i := 0; i < cfg.Steps; i++ {
		value := cfg.Min + float64(i)*stepSize
		config := cfg.BaseConfig
		applyParam(&config, cfg.Param, value)

		engine := core.NewEngine(config)
		if err := engine.Initialize(ctx); err != nil {
			return SensitivityResult{}, err
		}
		result, err := engine.Run(ctx)
		if err != nil {
			return SensitivityResult{}, err
		}

		points = append(points, SensitivityPoint{ParamValue: value, Metrics: result.Metrics})
	}

	elastic := estimateElasticity(points)

	return SensitivityResult{Param: cfg.Param, Points: points, Elastic: elastic}, nil
}

func estimateElasticity(points []SensitivityPoint) float64 {
	if len(points) < 2 {
		return 0
	}
	first := points[0]
	last := points[len(points)-1]
	if first.ParamValue == 0 || first.Metrics.AverageInflationBPS == 0 {
		return 0
	}

	changeParam := (last.ParamValue - first.ParamValue) / math.Abs(first.ParamValue)
	changeMetric := float64(last.Metrics.AverageInflationBPS-first.Metrics.AverageInflationBPS) / float64(first.Metrics.AverageInflationBPS)
	if changeParam == 0 {
		return 0
	}
	return changeMetric / changeParam
}
