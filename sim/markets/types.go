package markets

// MarketParams configures market dynamics.
type MarketParams struct {
	ComputeBasePrice float64
	StorageBasePrice float64
	GPUBasePrice     float64
	GasBasePrice     float64
	MinGasPrice      float64
	GasCapacity      float64
	FeeBurnBPS       int64

	PriceAdjustment float64
	MaxPriceMove    float64

	AdaptiveMinGasEnabled       bool
	GasTargetUtilizationBPS     int64
	GasAdjustmentRateBPS        int64
	GasMaxChangeBPS             int64
	GasCongestionThresholdBPS   int64
	GasCongestionMultiplierBPS  int64
	GasUtilizationSmoothingStep int64
}

// MarketState tracks market state across resource types.
type MarketState struct {
	ComputePrice float64
	StoragePrice float64
	GPUPrice     float64
	GasPrice     float64
	GasMinPrice  float64

	ComputeDemand float64
	StorageDemand float64
	GPUDemand     float64
	GasDemand     float64

	ComputeSupply float64
	StorageSupply float64
	GPUSupply     float64

	Utilization             float64
	GasUtilization          float64
	GasUtilizationEMA       float64
	GasCongestionMultiplier float64
	FeeRevenue              float64
}

// DefaultMarketParams provides sensible defaults.
//
// GasCapacity sizing (see the economics-sim gate, "avg_gas_utilization").
//
// GasCapacity used to be 300, which is EXACTLY the baseline scenario's mean
// gas demand: baseConfig() has 200 users at UserDemandMean 10, and
// agents/user.go splits each user's demand with GasDemand: demand * 0.15, so
//
//	200 * 10 * 0.15 = 300 = GasCapacity
//
// That arithmetic collision -- not a design decision -- pinned baseline
// utilization at ~1.0. UpdateGas computes utilization = demand/GasCapacity and
// sim/core/metrics.go averages it, so `ve-sim check` reported
//
//	avg_gas_utilization 0.9927890510656573 (band 0.05..0.95)
//
// and the Economics Simulation Suite has been red on every main run since the
// gate existed. The same collision also made the market's own controller dead
// code: GasCongestionThresholdBPS is 8500, so at ~0.99 utilization the
// congestion branch was taken on 366 of 366 steps and the adaptive min-gas
// controller could never return to its 0.65 target. A "baseline" that is
// permanently in the congestion regime is not a baseline.
//
// 500 puts baseline at 300/500 = 0.60, comfortably under the 0.65 target with
// headroom for demand variance (UserDemandStdDev 2.5), while still leaving
// congestion reachable: bull_market (350 users @ mean 18 -> 945 gas) and
// black_swan (600 @ 25 -> 2250) both exceed it, so the adaptive/congestion
// paths stay exercised by the scenario suite instead of being saturated away.
//
// Raising the DefaultThresholds band instead (0.05..0.95 -> up to 1.0) was the
// other option and was rejected: it would bless a permanently-saturated market
// as correct while leaving the model's 0.65 target unreachable. GasCapacity is
// the supply-side lever, and the band exists to catch supply-side faults.
func DefaultMarketParams() MarketParams {
	return MarketParams{
		ComputeBasePrice:            0.02,
		StorageBasePrice:            0.005,
		GPUBasePrice:                0.12,
		GasBasePrice:                0.001,
		MinGasPrice:                 0.0005,
		GasCapacity:                 500,
		FeeBurnBPS:                  2000,
		PriceAdjustment:             0.15,
		MaxPriceMove:                0.35,
		AdaptiveMinGasEnabled:       true,
		GasTargetUtilizationBPS:     6500,
		GasAdjustmentRateBPS:        2500,
		GasMaxChangeBPS:             2000,
		GasCongestionThresholdBPS:   8500,
		GasCongestionMultiplierBPS:  1500,
		GasUtilizationSmoothingStep: 8,
	}
}

// NewMarketState builds initial market state.
func NewMarketState(params MarketParams) MarketState {
	return MarketState{
		ComputePrice: params.ComputeBasePrice,
		StoragePrice: params.StorageBasePrice,
		GPUPrice:     params.GPUBasePrice,
		GasPrice:     params.GasBasePrice,
		GasMinPrice:  params.MinGasPrice,
	}
}
