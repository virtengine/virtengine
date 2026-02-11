package gaspricing

import sdk "github.com/cosmos/cosmos-sdk/types"

// State tracks adaptive gas pricing state.
type State struct {
	CurrentMinGasPrices        sdk.DecCoins `json:"current_min_gas_prices"`
	SmoothedUtilizationBPS     int64        `json:"smoothed_utilization_bps"`
	LastBlockHeight            int64        `json:"last_block_height"`
	LastComputedUtilizationBPS int64        `json:"last_computed_utilization_bps"`
}

// DefaultState returns the initial state for adaptive gas pricing.
func DefaultState(params Params) State {
	return State{
		CurrentMinGasPrices:        params.MinGasPrices,
		SmoothedUtilizationBPS:     params.TargetBlockUtilizationBPS,
		LastBlockHeight:            0,
		LastComputedUtilizationBPS: params.TargetBlockUtilizationBPS,
	}
}
