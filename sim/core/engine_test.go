package core_test

import (
	"context"
	"testing"
	"time"

	"github.com/virtengine/virtengine/sim/core"
	"github.com/virtengine/virtengine/sim/scenarios"
)

func TestEngineRunBaseline(t *testing.T) {
	cfg := scenarios.BaselineConfig()
	cfg.EndTime = cfg.StartTime.Add(30 * 24 * time.Hour)
	cfg.TimeStep = 24 * time.Hour
	cfg.NumUsers = 20
	cfg.NumProviders = 5
	cfg.NumValidators = 8
	cfg.NumAttackers = 1

	engine := core.NewEngine(cfg)
	if err := engine.Initialize(context.Background()); err != nil {
		t.Fatalf("init: %v", err)
	}

	result, err := engine.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if len(result.Snapshots) == 0 {
		t.Fatalf("expected snapshots")
	}
	if result.Final.TokenSupply.Sign() == 0 {
		t.Fatalf("expected supply")
	}
}
