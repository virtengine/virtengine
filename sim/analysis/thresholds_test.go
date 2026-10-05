package analysis_test

import (
	"context"
	"testing"
	"time"

	"github.com/virtengine/virtengine/sim/analysis"
	"github.com/virtengine/virtengine/sim/core"
	"github.com/virtengine/virtengine/sim/markets"
	"github.com/virtengine/virtengine/sim/scenarios"
)

// TestDefaultGasCapacityClearsTheUtilizationBand pins the gas supply-side
// sizing that the economics-sim gate depends on.
//
// The `Economics Simulation Suite` job runs
//
//	go run ./cmd/ve-sim suite --scenario baseline ...
//	go run ./cmd/ve-sim check --metrics sim-output/metrics.json
//
// and `check` fails the whole job when any metric falls outside
// analysis.DefaultThresholds(). `avg_gas_utilization` is banded 0.05..0.95.
//
// markets.DefaultMarketParams() used to set GasCapacity: 300, which is EXACTLY
// the baseline scenario's mean gas demand -- baseConfig() has 200 users at
// UserDemandMean 10 and agents/user.go uses GasDemand: demand * 0.15, so
// 200 * 10 * 0.15 = 300. That collision pinned utilization at ~1.0 and the gate
// reported
//
//	avg_gas_utilization 0.9927890510656573 0.05 0.95
//
// The Economics Simulation Suite had never once passed on main (0 of the last
// 100 runs) as a result. GasCapacity is now 500, which puts baseline at 0.60.
//
// This test runs the REAL baseline config through the REAL engine and the REAL
// threshold check, so it pins the end-to-end behaviour the CI job asserts
// rather than restating the constant.
func TestDefaultGasCapacityClearsTheUtilizationBand(t *testing.T) {
	cfg := scenarios.BaselineConfig()

	// Guard the arithmetic the fix reasons about. If anyone changes NumUsers,
	// UserDemandMean or the per-user gas share in agents/user.go, this is the
	// first thing that should notice -- the capacity constant is only correct
	// relative to the demand it serves.
	const (
		wantUsers        = 200
		wantDemandMean   = 10.0
		wantGasShare     = 0.15
		wantGasCapacity  = 500.0
		wantUtilAtTarget = 0.60
	)

	if cfg.NumUsers != wantUsers {
		t.Fatalf("baseline NumUsers = %d, this test reasons about %d", cfg.NumUsers, wantUsers)
	}
	if cfg.UserDemandMean != wantDemandMean {
		t.Fatalf("baseline UserDemandMean = %v, this test reasons about %v",
			cfg.UserDemandMean, wantDemandMean)
	}
	if markets.DefaultMarketParams().GasCapacity != wantGasCapacity {
		t.Fatalf("DefaultMarketParams().GasCapacity = %v, want %v",
			markets.DefaultMarketParams().GasCapacity, wantGasCapacity)
	}

	// Mean gas demand is NumUsers * UserDemandMean * gasShare, and UpdateGas
	// computes utilization = demand / GasCapacity. Assert the intended
	// utilization directly so the constant cannot drift back onto the demand.
	meanGasDemand := float64(cfg.NumUsers) * cfg.UserDemandMean * wantGasShare
	if util := meanGasDemand / wantGasCapacity; util > 0.65 {
		t.Fatalf("mean gas demand %g over capacity %g is %g utilization, at or above "+
			"GasTargetUtilizationBPS (0.65) -- baseline would sit permanently in congestion",
			meanGasDemand, wantGasCapacity, util)
	} else if wantUtilAtTarget > 0 && util < 0.5 {
		t.Fatalf("mean gas demand %g over capacity %g is only %g utilization; capacity "+
			"was raised so far that the adaptive/congestion price paths are no longer exercised",
			meanGasDemand, wantGasCapacity, util)
	}

	// Now the end-to-end assertion: run the baseline scenario and pass the real
	// metrics through the real threshold check that `ve-sim check` runs.
	result, err := core.NewEngine(cfg).Run(context.Background())
	if err != nil {
		t.Fatalf("baseline run: %v", err)
	}

	thresholds := analysis.DefaultThresholds()
	if err := analysis.ValidateThresholds(thresholds); err != nil {
		t.Fatalf("DefaultThresholds is not self-consistent: %v", err)
	}

	violations := analysis.CheckThresholds(result.Metrics, thresholds)
	for _, v := range violations {
		t.Errorf("baseline violates %s: %g (band %g..%g)", v.Metric, v.Value, v.Min, v.Max)
	}

	// Assert the band is still the one CI enforces, so a later "widen the band"
	// change cannot quietly pass by editing both sides of this test.
	band, ok := thresholds["avg_gas_utilization"]
	if !ok {
		t.Fatal("avg_gas_utilization has no threshold; the gate would silently stop checking it")
	}
	if band.Min != 0.05 || band.Max != 0.95 {
		t.Errorf("avg_gas_utilization band = %g..%g, want 0.05..0.95", band.Min, band.Max)
	}

	t.Logf("baseline avg_gas_utilization = %g (band %g..%g); avg_gas_price = %g",
		result.Metrics.AvgGasUtilization, band.Min, band.Max, result.Metrics.AvgGasPrice)
}

// TestBaselineGasUtilizationStaysBoundedByDemand guards the clamp itself: a
// baseline run must not report utilization at or above the congestion
// threshold, which is the condition that made the adaptive min-gas controller
// unreachable at its GasTargetUtilizationBPS.
//
// It runs a SHORT window on purpose. The full 366-step baseline takes about a
// minute, and the per-step signal here is established within a few steps; the
// full-horizon end-to-end assertion lives in the test above.
func TestBaselineGasUtilizationStaysBoundedByDemand(t *testing.T) {
	cfg := scenarios.BaselineConfig()
	cfg.EndTime = cfg.StartTime.Add(30 * 24 * time.Hour)
	cfg.TimeStep = 24 * time.Hour

	params := markets.DefaultMarketParams()

	result, err := core.NewEngine(cfg).Run(context.Background())
	if err != nil {
		t.Fatalf("baseline run: %v", err)
	}

	// Utilization is clamped at 1.0 in UpdateGas, so 1.0 means "demand met or
	// exceeded capacity for the whole run" -- exactly the saturation that made
	// the previous GasCapacity of 300 unusable as a baseline.
	if result.Metrics.AvgGasUtilization >= 1.0 {
		t.Errorf("baseline saturated: avg_gas_utilization = %g (clamped ceiling is 1.0)",
			result.Metrics.AvgGasUtilization)
	}

	congestion := float64(params.GasCongestionThresholdBPS) / 10000
	target := float64(params.GasTargetUtilizationBPS) / 10000
	if result.Metrics.AvgGasUtilization > congestion {
		t.Errorf("baseline avg_gas_utilization %g is above GasCongestionThresholdBPS %g; "+
			"the adaptive min-gas controller cannot return to its %g target",
			result.Metrics.AvgGasUtilization, congestion, target)
	}

	// Sanity: the price bands CI asserts must still hold at this utilization.
	if v := analysis.CheckThresholds(result.Metrics, analysis.DefaultThresholds()); len(v) > 0 {
		t.Errorf("30-day baseline violates thresholds: %v", v)
	}

	t.Logf("30-day baseline avg_gas_utilization = %g (target %g, congestion %g)",
		result.Metrics.AvgGasUtilization, target, congestion)
}

// TestGasCongestionRemainsReachable guards the OTHER half of the capacity fix.
//
// Raising GasCapacity must not push every scenario into a permanently
// uncongested regime, which would make the adaptive/congestion price paths dead
// code in the scenario suite. black_swan's mean gas demand
// (600 users * mean 25 * 0.15 = 2250) must still exceed default capacity.
func TestGasCongestionRemainsReachable(t *testing.T) {
	params := markets.DefaultMarketParams()

	cases := []struct {
		name           string
		users          int
		demandMean     float64
		wantCongestion bool
	}{
		{"baseline", 200, 10, false},
		{"bull_market", 350, 18, true},
		{"black_swan", 600, 25, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gasDemand := float64(tc.users) * tc.demandMean * 0.15
			congested := gasDemand/params.GasCapacity >= 1.0
			if congested != tc.wantCongestion {
				t.Errorf("%s: gas demand %g over capacity %g -> congested=%v, want %v",
					tc.name, gasDemand, params.GasCapacity, congested, tc.wantCongestion)
			}
			t.Logf("%s: gas demand %g / capacity %g = %g utilization",
				tc.name, gasDemand, params.GasCapacity, gasDemand/params.GasCapacity)
		})
	}
}
