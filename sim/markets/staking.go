package markets

// StakingMarket models staking equilibrium.
type StakingMarket struct {
	TargetRatioBPS int64
	AdjustmentBPS  int64
}

// UpdateRatio adjusts staking ratio based on APR incentives.
func (s StakingMarket) UpdateRatio(currentRatioBPS, aprBPS int64) int64 {
	if currentRatioBPS == 0 {
		return 0
	}
	adjustment := s.AdjustmentBPS
	if aprBPS > 1200 {
		currentRatioBPS += adjustment
	} else if aprBPS < 600 {
		currentRatioBPS -= adjustment
	}
	if currentRatioBPS < 0 {
		return 0
	}
	return currentRatioBPS
}
