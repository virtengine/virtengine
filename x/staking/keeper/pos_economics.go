package keeper

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/staking/types"
)

const minVirtualStakeUnits int64 = types.FixedPointScale

// buildStakeDistribution returns per-validator stake and network total stake.
// It uses on-chain stake when available and deterministically falls back to
// virtual stake units derived from recorded performance.
func (k Keeper) buildStakeDistribution(ctx sdk.Context, performances []types.ValidatorPerformance) (map[string]int64, int64) {
	stakes := make(map[string]int64, len(performances))
	if len(performances) == 0 {
		return stakes, 0
	}

	if k.stakingKeeper == nil {
		var total int64
		for _, perf := range performances {
			stake := k.virtualStakeFromPerformance(perf)
			stakes[perf.ValidatorAddress] = stake
			total = saturatingAddPositiveInt64(total, stake)
		}
		if total == 0 {
			total = saturatingMulInt64(minVirtualStakeUnits, int64(len(performances)))
		}
		return stakes, total
	}

	var totalStake int64

	for _, perf := range performances {
		stake := int64(0)
		validatorAddr, err := sdk.AccAddressFromBech32(perf.ValidatorAddress)
		if err == nil {
			stake = k.stakingKeeper.GetValidatorStake(ctx, validatorAddr)
		}
		if stake <= 0 {
			stake = k.virtualStakeFromPerformance(perf)
		}
		stakes[perf.ValidatorAddress] = stake
		totalStake = saturatingAddPositiveInt64(totalStake, stake)
	}

	// Blend resolved per-validator stake with network-wide totals. Using the
	// larger value prevents over-distribution when some validators have stake
	// but no recorded performance entry in the current epoch.
	networkTotalStake := k.stakingKeeper.GetTotalStake(ctx)
	if networkTotalStake > totalStake {
		totalStake = networkTotalStake
	}
	if totalStake <= 0 {
		totalStake = saturatingMulInt64(minVirtualStakeUnits, int64(len(performances)))
	}

	return stakes, totalStake
}

func (k Keeper) virtualStakeFromPerformance(perf types.ValidatorPerformance) int64 {
	score := perf.OverallScore
	if score < 0 || score > types.MaxPerformanceScore {
		score = 0
	}
	if score == 0 {
		perfCopy := perf
		score = types.ComputeOverallScore(&perfCopy)
	}
	if score < 0 {
		score = 0
	}
	if score > types.MaxPerformanceScore {
		score = types.MaxPerformanceScore
	}

	activityUnits := int64(1)
	activityUnits = saturatingAddPositiveInt64(activityUnits, clampNonNegative(perf.BlocksExpected))
	activityUnits = saturatingAddPositiveInt64(activityUnits, clampNonNegative(perf.VEIDVerificationsExpected))

	// If expected workload is unavailable, use observed workload.
	if activityUnits == 1 {
		activityUnits = saturatingAddPositiveInt64(activityUnits, clampNonNegative(perf.BlocksProposed))
		activityUnits = saturatingAddPositiveInt64(activityUnits, clampNonNegative(perf.VEIDVerificationsCompleted))
	}

	qualityMultiplier := (types.MaxPerformanceScore / 2) + score // 0.5x to 1.5x
	virtualStake := saturatingMulInt64(activityUnits, qualityMultiplier)
	if virtualStake < minVirtualStakeUnits {
		return minVirtualStakeUnits
	}
	return virtualStake
}

// slashableStakeForValidator returns the best available stake estimate used
// for deterministic slash token calculation.
func (k Keeper) slashableStakeForValidator(ctx sdk.Context, validatorAddr string) int64 {
	if k.stakingKeeper != nil {
		if accAddr, err := sdk.AccAddressFromBech32(validatorAddr); err == nil {
			if stake := k.stakingKeeper.GetValidatorStake(ctx, accAddr); stake > 0 {
				return stake
			}
		}
	}

	var (
		latest      types.ValidatorPerformance
		latestEpoch uint64
		found       bool
	)
	k.WithValidatorPerformances(ctx, func(perf types.ValidatorPerformance) bool {
		if perf.ValidatorAddress != validatorAddr {
			return false
		}
		if !found || perf.EpochNumber >= latestEpoch {
			latest = perf
			latestEpoch = perf.EpochNumber
			found = true
		}
		return false
	})
	if found {
		return k.virtualStakeFromPerformance(latest)
	}

	params := k.GetParams(ctx)
	epochBlocks := safeInt64FromUint64Rewards(params.EpochLength)
	if epochBlocks <= 0 {
		epochBlocks = 1
	}
	baseReward := params.BaseRewardPerBlock
	if baseReward <= 0 {
		baseReward = 1
	}
	stake := saturatingMulInt64(baseReward, epochBlocks)
	if stake < minVirtualStakeUnits {
		return minVirtualStakeUnits
	}
	return stake
}

func clampNonNegative(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}

func saturatingAddPositiveInt64(a, b int64) int64 {
	if b <= 0 {
		return a
	}
	if a > math.MaxInt64-b {
		return math.MaxInt64
	}
	return a + b
}

func saturatingMulInt64(a, b int64) int64 {
	if a == 0 || b == 0 {
		return 0
	}
	if a > 0 && b > 0 && a > math.MaxInt64/b {
		return math.MaxInt64
	}
	if a < 0 && b < 0 && a < math.MaxInt64/b {
		return math.MaxInt64
	}
	if a > 0 && b < 0 && b < math.MinInt64/a {
		return math.MinInt64
	}
	if a < 0 && b > 0 && a < math.MinInt64/b {
		return math.MinInt64
	}
	return a * b
}
