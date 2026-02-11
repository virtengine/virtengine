package gaspricing

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Params defines adaptive gas pricing parameters.
type Params struct {
	Enabled                    bool
	MinGasPrices               sdk.DecCoins
	MaxGasPrices               sdk.DecCoins
	TargetBlockUtilizationBPS  int64
	AdjustmentRateBPS          int64
	MaxChangeBPS               int64
	CongestionThresholdBPS     int64
	CongestionMultiplierBPS    int64
	UtilizationSmoothingWindow uint32
}

// DefaultParams returns default adaptive gas parameters.
func DefaultParams(minGas sdk.DecCoins) Params {
	minGas = minGas.Sort()
	return Params{
		Enabled:                    true,
		MinGasPrices:               minGas,
		MaxGasPrices:               scaleDecCoins(minGas, 100000), // 10x cap
		TargetBlockUtilizationBPS:  6500,
		AdjustmentRateBPS:          2500,
		MaxChangeBPS:               2000,
		CongestionThresholdBPS:     8500,
		CongestionMultiplierBPS:    1500,
		UtilizationSmoothingWindow: 8,
	}
}

// Validate ensures params are sane.
func (p Params) Validate() error {
	if p.TargetBlockUtilizationBPS < 0 || p.TargetBlockUtilizationBPS > 10000 {
		return fmt.Errorf("target_block_utilization_bps must be between 0 and 10000")
	}
	if p.CongestionThresholdBPS < 0 || p.CongestionThresholdBPS > 10000 {
		return fmt.Errorf("congestion_threshold_bps must be between 0 and 10000")
	}
	if p.AdjustmentRateBPS < 0 || p.AdjustmentRateBPS > 10000 {
		return fmt.Errorf("adjustment_rate_bps must be between 0 and 10000")
	}
	if p.MaxChangeBPS < 0 || p.MaxChangeBPS > 10000 {
		return fmt.Errorf("max_change_bps must be between 0 and 10000")
	}
	if p.CongestionMultiplierBPS < 0 || p.CongestionMultiplierBPS > 10000 {
		return fmt.Errorf("congestion_multiplier_bps must be between 0 and 10000")
	}
	if !p.MinGasPrices.IsValid() {
		return fmt.Errorf("min_gas_prices must be valid dec coins")
	}
	if !p.MaxGasPrices.IsValid() {
		return fmt.Errorf("max_gas_prices must be valid dec coins")
	}
	if !p.MaxGasPrices.IsZero() && !decCoinsAllGTE(p.MaxGasPrices, p.MinGasPrices) {
		return fmt.Errorf("max_gas_prices must be >= min_gas_prices")
	}
	return nil
}
