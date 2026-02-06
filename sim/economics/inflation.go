package economics

import (
	"math/big"
)

// InflationModel computes inflation adjustments.
type InflationModel struct {
	TargetInflationBPS int64
	MinInflationBPS    int64
	MaxInflationBPS    int64
	TargetStakingBPS   int64
}

// AdjustInflation returns inflation based on staking ratio.
func (m InflationModel) AdjustInflation(stakingRatioBPS int64) int64 {
	if stakingRatioBPS == m.TargetStakingBPS {
		return m.TargetInflationBPS
	}
	deviation := m.TargetStakingBPS - stakingRatioBPS
	adjustment := deviation / 10
	inflation := m.TargetInflationBPS + adjustment
	if inflation < m.MinInflationBPS {
		return m.MinInflationBPS
	}
	if inflation > m.MaxInflationBPS {
		return m.MaxInflationBPS
	}
	return inflation
}

// MintForPeriod calculates minted tokens for a period.
func (m InflationModel) MintForPeriod(supply *big.Int, inflationBPS int64, periods int64) *big.Int {
	if supply == nil || supply.Sign() == 0 {
		return big.NewInt(0)
	}
	mint := new(big.Int).Mul(supply, big.NewInt(inflationBPS))
	mint.Div(mint, big.NewInt(10000))
	mint.Div(mint, big.NewInt(periods))
	return mint
}
