package markets

import "math"

// FeeResult captures fee market outcomes.
type FeeResult struct {
	TotalFees float64
	Burned    float64
}

// ApplyFees calculates fee burn based on gas usage.
func ApplyFees(state MarketState, params MarketParams) (MarketState, FeeResult) {
	gasDemand := math.Max(state.GasDemand, 0.01)
	fees := gasDemand * state.GasPrice
	burned := fees * (float64(params.FeeBurnBPS) / 10000)

	state.FeeRevenue += fees - burned
	return state, FeeResult{TotalFees: fees, Burned: burned}
}
