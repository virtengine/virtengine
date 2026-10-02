package markets

import "testing"

// TestUpdateGasMinGasPriceIsBounded pins the regression where the adaptive min
// gas price compounded without limit.
//
// UpdateGas used to compute the adaptive floor as
//
//	minGas = state.GasMinPrice * (1 + adjustment)
//
// which reads the PREVIOUS step's value. At a sustained utilisation above the
// target the adjustment stays positive, so the level compounds by
// (1 + adjustment) every step and nothing ever pulls it back. At the baseline
// simulation's 0.993 utilisation the adjustment pins at +0.0857, giving
// 1.0857^step: 114143 by step 100 and 9.5e33 by step 400. That is what made
// `ve-sim check` report
//
//	avg_gas_price 5.521088774418586e+28 (bound 0.0004..0.01)
//	avg_min_gas_price 5.521088774418586e+28 (bound 0.0004..0.008)
//
// The floor is now anchored on params.MinGasPrice, so it is bounded by
// MinGasPrice * (1 +/- GasMaxChangeBPS/10000) * (1 + congestion).
func TestUpdateGasMinGasPriceIsBounded(t *testing.T) {
	params := DefaultMarketParams()
	state := NewMarketState(params)

	// The utilisation the failing baseline run reported: 0.9927890510656573.
	const util = 0.9927890510656573
	demand := params.GasCapacity * util

	// The economics suite runs the baseline scenario for 366 steps; run well past
	// that so a slow re-introduction of compounding cannot pass.
	const steps = 2000

	// params.MinGasPrice * (1 + GasMaxChangeBPS/10000) * (1 + congestion multiplier)
	upper := params.MinGasPrice * (1 + float64(params.GasMaxChangeBPS)/10000) *
		(1 + float64(params.GasCongestionMultiplierBPS)/10000)

	// GasPrice itself is NOT recursive: it is recomputed from scratch every step
	// as GasBasePrice * (1 + min(MaxPriceMove, PriceAdjustment*(demand-1))), so it
	// is capped at GasBasePrice*(1+MaxPriceMove) regardless of history. It was
	// only ever out of bounds because the next line floors it UP to the runaway
	// adaptive minimum, so bound it by whichever of the two sources is larger.
	priceUpper := params.GasBasePrice * (1 + params.MaxPriceMove)
	if upper > priceUpper {
		priceUpper = upper
	}

	for step := 0; step < steps; step++ {
		state.GasDemand = demand
		state = UpdateGas(state, params)

		if state.GasMinPrice > upper {
			t.Fatalf("step %d: GasMinPrice = %g exceeds its bound %g; the adaptive floor is compounding again",
				step, state.GasMinPrice, upper)
		}
		if state.GasMinPrice < params.MinGasPrice {
			t.Fatalf("step %d: GasMinPrice = %g fell below params.MinGasPrice %g",
				step, state.GasMinPrice, params.MinGasPrice)
		}
		// GasPrice is floored at the adaptive minimum, so it inherits the same
		// defect and must stay bounded too.
		if state.GasPrice > priceUpper {
			t.Fatalf("step %d: GasPrice = %g exceeds the bound %g",
				step, state.GasPrice, priceUpper)
		}
		// The economics gate also asserts avg_gas_price stays within 0.0004..0.01.
		if state.GasPrice > 0.01 {
			t.Fatalf("step %d: GasPrice = %g exceeds the economics bound 0.01", step, state.GasPrice)
		}
	}

	t.Logf("after %d steps at utilization %.6f: GasMinPrice=%g GasPrice=%g (bound %g)",
		steps, util, state.GasMinPrice, state.GasPrice, priceUpper)
}

// TestUpdateGasMinGasPriceSettles checks the floor still RESPONDS to
// utilisation, so the bound above cannot be met by simply pinning the value.
func TestUpdateGasMinGasPriceSettles(t *testing.T) {
	params := DefaultMarketParams()

	under := NewMarketState(params)
	over := NewMarketState(params)

	const steps = 500
	for step := 0; step < steps; step++ {
		under.GasDemand = params.GasCapacity * 0.10 // well under the 0.65 target
		under = UpdateGas(under, params)
		over.GasDemand = params.GasCapacity * 0.99 // above target, into congestion
		over = UpdateGas(over, params)
	}

	if over.GasMinPrice <= under.GasMinPrice {
		t.Errorf("adaptive floor did not rise with utilisation: under=%g over=%g",
			under.GasMinPrice, over.GasMinPrice)
	}
	if under.GasMinPrice < params.MinGasPrice {
		t.Errorf("under-utilised floor %g fell below params.MinGasPrice %g",
			under.GasMinPrice, params.MinGasPrice)
	}
	t.Logf("under=%g over=%g min=%g", under.GasMinPrice, over.GasMinPrice, params.MinGasPrice)
}
