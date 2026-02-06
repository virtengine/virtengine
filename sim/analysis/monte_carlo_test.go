package analysis_test

import (
	"context"
	"testing"
	"time"

	"github.com/virtengine/virtengine/sim/analysis"
	"github.com/virtengine/virtengine/sim/scenarios"
)

func TestMonteCarloRun(t *testing.T) {
	cfg := scenarios.BaselineConfig()
	cfg.EndTime = cfg.StartTime.Add(10 * 24 * time.Hour)
	cfg.TimeStep = 24 * time.Hour
	cfg.NumUsers = 10
	cfg.NumProviders = 3
	cfg.NumValidators = 5
	cfg.NumAttackers = 1

	mc := analysis.NewMonteCarloAnalyzer(analysis.MonteCarloConfig{
		Runs:         3,
		Parallelism:  2,
		Confidence:   0.9,
		ParamRanges:  map[string]analysis.ParameterRange{"inflation_target_bps": {Min: 500, Max: 900, Dist: "normal"}},
		BaseConfig:   cfg,
		ScenarioName: "baseline",
	})

	results, err := mc.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected results")
	}
}
