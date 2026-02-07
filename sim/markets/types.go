package markets

// MarketParams configures market dynamics.
type MarketParams struct {
	ComputeBasePrice float64
	StorageBasePrice float64
	GPUBasePrice     float64
	GasBasePrice     float64
	MinGasPrice      float64
	FeeBurnBPS       int64

	PriceAdjustment float64
	MaxPriceMove    float64
}

// MarketState tracks market state across resource types.
type MarketState struct {
	ComputePrice float64
	StoragePrice float64
	GPUPrice     float64
	GasPrice     float64

	ComputeDemand float64
	StorageDemand float64
	GPUDemand     float64
	GasDemand     float64

	ComputeSupply float64
	StorageSupply float64
	GPUSupply     float64

	Utilization float64
	FeeRevenue  float64
}

// DefaultMarketParams provides sensible defaults.
func DefaultMarketParams() MarketParams {
	return MarketParams{
		ComputeBasePrice: 0.02,
		StorageBasePrice: 0.005,
		GPUBasePrice:     0.12,
		GasBasePrice:     0.001,
		MinGasPrice:      0.0005,
		FeeBurnBPS:       2000,
		PriceAdjustment:  0.15,
		MaxPriceMove:     0.35,
	}
}

// NewMarketState builds initial market state.
func NewMarketState(params MarketParams) MarketState {
	return MarketState{
		ComputePrice: params.ComputeBasePrice,
		StoragePrice: params.StorageBasePrice,
		GPUPrice:     params.GPUBasePrice,
		GasPrice:     params.GasBasePrice,
	}
}
