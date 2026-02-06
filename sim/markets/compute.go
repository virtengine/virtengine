package markets

import "math"

// UpdateCompute updates compute market prices based on demand/supply.
func UpdateCompute(state MarketState, params MarketParams) MarketState {
	state.ComputePrice = adjustPrice(state.ComputePrice, state.ComputeDemand, state.ComputeSupply, params)
	return state
}

// UpdateStorage updates storage market prices based on demand/supply.
func UpdateStorage(state MarketState, params MarketParams) MarketState {
	state.StoragePrice = adjustPrice(state.StoragePrice, state.StorageDemand, state.StorageSupply, params)
	return state
}

// UpdateGPU updates GPU market prices based on demand/supply.
func UpdateGPU(state MarketState, params MarketParams) MarketState {
	state.GPUPrice = adjustPrice(state.GPUPrice, state.GPUDemand, state.GPUSupply, params)
	return state
}

// UpdateGas updates gas price based on demand and min gas.
func UpdateGas(state MarketState, params MarketParams) MarketState {
	demand := state.GasDemand
	if demand < 0.01 {
		demand = 0.01
	}
	price := params.GasBasePrice * (1 + math.Min(params.MaxPriceMove, params.PriceAdjustment*(demand-1)))
	if price < params.MinGasPrice {
		price = params.MinGasPrice
	}
	state.GasPrice = price
	return state
}

func adjustPrice(current, demand, supply float64, params MarketParams) float64 {
	if supply <= 0 {
		supply = 0.1
	}
	if demand <= 0 {
		demand = 0.1
	}
	ratio := demand / supply
	move := params.PriceAdjustment * (ratio - 1)
	if move > params.MaxPriceMove {
		move = params.MaxPriceMove
	}
	if move < -params.MaxPriceMove {
		move = -params.MaxPriceMove
	}
	newPrice := current * (1 + move)
	if newPrice < current*0.25 {
		newPrice = current * 0.25
	}
	return newPrice
}
