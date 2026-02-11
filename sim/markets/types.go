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
func DefaultMarketParams() MarketParams {
	return MarketParams{
		ComputeBasePrice:            0.02,
		StorageBasePrice:            0.005,
		GPUBasePrice:                0.12,
		GasBasePrice:                0.001,
		MinGasPrice:                 0.0005,
		GasCapacity:                 300,
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
