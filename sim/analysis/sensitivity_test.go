package analysis_test

import (
	"context"
	"testing"
	"time"

	"github.com/virtengine/virtengine/sim/analysis"
	"github.com/virtengine/virtengine/sim/scenarios"
)

func TestSensitivityRun(t *testing.T) {
	cfg := scenarios.BaselineConfig()
	cfg.EndTime = cfg.StartTime.Add(7 * 24 * time.Hour)
	cfg.TimeStep = 24 * time.Hour
	cfg.NumUsers = 8
	cfg.NumProviders = 2
	cfg.NumValidators = 4
	cfg.NumAttackers = 0

	result, err := analysis.RunSensitivity(context.Background(), analysis.SensitivityConfig{
		BaseConfig: cfg,
		Steps:      3,
		Param:      "inflation_target_bps",
		Min:        500,
		Max:        900,
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(result.Points) != 3 {
		t.Fatalf("expected 3 points")
	}
}
