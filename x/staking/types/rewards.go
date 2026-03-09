// Package types contains types for the staking module.
//
// VE-921: Reward types for validator staking rewards
// This file provides utility methods for reward types (generated proto types).
package types

import (
	"fmt"
	"math/big"
	"sort"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	stakingv1 "github.com/virtengine/virtengine/sdk/go/node/staking/v1"
)

// NewRewardEpoch creates a new reward epoch
func NewRewardEpoch(epochNumber uint64, startHeight int64, startTime time.Time) *stakingv1.RewardEpoch {
	return &stakingv1.RewardEpoch{
		EpochNumber:             epochNumber,
		StartHeight:             startHeight,
		StartTime:               &startTime,
		TotalRewardsDistributed: sdk.NewCoins(),
		BlockProposalRewards:    sdk.NewCoins(),
		VEIDRewards:             sdk.NewCoins(),
		UptimeRewards:           sdk.NewCoins(),
	}
}

// ValidateRewardEpoch validates the reward epoch
func ValidateRewardEpoch(e *stakingv1.RewardEpoch) error {
	if e.StartHeight < 0 {
		return fmt.Errorf("start_height cannot be negative")
	}

	if e.EndHeight != 0 && e.EndHeight < e.StartHeight {
		return fmt.Errorf("end_height cannot be before start_height")
	}

	if e.ValidatorCount < 0 {
		return fmt.Errorf("validator_count cannot be negative")
	}

	return nil
}

// EpochDuration returns the epoch duration in blocks
func EpochDuration(e *stakingv1.RewardEpoch) int64 {
	if e.EndHeight == 0 {
		return 0
	}
	return e.EndHeight - e.StartHeight
}

// NewValidatorReward creates a new validator reward
func NewValidatorReward(validatorAddr string, epochNumber uint64) *stakingv1.ValidatorReward {
	return &stakingv1.ValidatorReward{
		ValidatorAddress:      validatorAddr,
		EpochNumber:           epochNumber,
		TotalReward:           sdk.NewCoins(),
		BlockProposalReward:   sdk.NewCoins(),
		VEIDReward:            sdk.NewCoins(),
		UptimeReward:          sdk.NewCoins(),
		IdentityNetworkReward: sdk.NewCoins(),
	}
}

// ValidateValidatorReward validates the validator reward
func ValidateValidatorReward(r *stakingv1.ValidatorReward) error {
	if r.ValidatorAddress == "" {
		return fmt.Errorf("validator_address cannot be empty")
	}

	if r.PerformanceScore < 0 || r.PerformanceScore > MaxPerformanceScore {
		return fmt.Errorf("performance_score must be between 0 and %d", MaxPerformanceScore)
	}

	return nil
}

// ComputeTotalReward computes the total reward from components
func ComputeTotalReward(r *stakingv1.ValidatorReward) sdk.Coins {
	total := sdk.NewCoins()
	total = total.Add(r.BlockProposalReward...)
	total = total.Add(r.VEIDReward...)
	total = total.Add(r.UptimeReward...)
	total = total.Add(r.IdentityNetworkReward...)
	return total
}

// RewardCalculationInput represents input for reward calculation
type RewardCalculationInput struct {
	// ValidatorAddress is the validator's address
	ValidatorAddress string

	// Performance is the validator's performance metrics
	Performance *stakingv1.ValidatorPerformance

	// StakeAmount is the validator's stake amount
	StakeAmount int64

	// TotalStake is the total network stake
	TotalStake int64

	// EpochRewardPool is the total reward pool for the epoch
	EpochRewardPool int64

	// BlocksInEpoch is the number of blocks in the epoch
	BlocksInEpoch int64

	// VEIDVerificationsInEpoch is total VEID verifications in the epoch
	VEIDVerificationsInEpoch int64
}

// CalculateRewards calculates rewards deterministically using integer arithmetic
func CalculateRewards(input RewardCalculationInput, denom string) *stakingv1.ValidatorReward {
	reward := NewValidatorReward(input.ValidatorAddress, 0)

	if input.StakeAmount <= 0 || input.TotalStake <= 0 || input.EpochRewardPool <= 0 {
		return reward
	}

	stakeWeight := formatFixedPointRatio(input.StakeAmount, input.TotalStake)
	blockScore, veidScore, uptimeScore, overallScore := rewardComponentScores(input.Performance)

	type rewardComponent int
	const (
		rewardComponentBlock rewardComponent = iota
		rewardComponentVEID
		rewardComponentUptime
	)

	type rewardAllocation struct {
		component rewardComponent
		weight    int64
		score     int64
	}

	type rewardRemainder struct {
		component rewardComponent
		remainder *big.Int
	}

	allocations := []rewardAllocation{
		{component: rewardComponentBlock, weight: WeightBlockProposal, score: blockScore},
		{component: rewardComponentVEID, weight: WeightVEIDVerification, score: veidScore},
		{component: rewardComponentUptime, weight: WeightUptime, score: uptimeScore},
	}

	denominator := big.NewInt(input.TotalStake)
	denominator.Mul(denominator, big.NewInt(TotalWeight))
	denominator.Mul(denominator, big.NewInt(MaxPerformanceScore))

	remainders := make([]rewardRemainder, 0, len(allocations))
	totalAllocated := big.NewInt(0)
	totalNumerator := big.NewInt(0)

	for _, allocation := range allocations {
		numerator := big.NewInt(input.EpochRewardPool)
		numerator.Mul(numerator, big.NewInt(input.StakeAmount))
		numerator.Mul(numerator, big.NewInt(allocation.weight))
		numerator.Mul(numerator, big.NewInt(allocation.score))
		totalNumerator.Add(totalNumerator, numerator)

		quotient := new(big.Int)
		remainder := new(big.Int)
		quotient.QuoRem(numerator, denominator, remainder)
		totalAllocated.Add(totalAllocated, quotient)
		remainders = append(remainders, rewardRemainder{component: allocation.component, remainder: remainder})

		switch allocation.component {
		case rewardComponentBlock:
			reward.BlockProposalReward = coinSetFromBigInt(denom, quotient)
		case rewardComponentVEID:
			reward.VEIDReward = coinSetFromBigInt(denom, quotient)
		case rewardComponentUptime:
			reward.UptimeReward = coinSetFromBigInt(denom, quotient)
		}
	}

	totalEarned := new(big.Int).Quo(totalNumerator, denominator)
	leftover := new(big.Int).Sub(totalEarned, totalAllocated)
	if leftover.Sign() > 0 {
		sort.SliceStable(remainders, func(i, j int) bool {
			cmp := remainders[i].remainder.Cmp(remainders[j].remainder)
			if cmp != 0 {
				return cmp > 0
			}
			return remainders[i].component < remainders[j].component
		})

		one := big.NewInt(1)
		for idx := 0; idx < len(remainders) && leftover.Sign() > 0; idx++ {
			switch remainders[idx].component {
			case rewardComponentBlock:
				reward.BlockProposalReward = incrementCoinSet(reward.BlockProposalReward, denom)
			case rewardComponentVEID:
				reward.VEIDReward = incrementCoinSet(reward.VEIDReward, denom)
			case rewardComponentUptime:
				reward.UptimeReward = incrementCoinSet(reward.UptimeReward, denom)
			}
			leftover.Sub(leftover, one)
		}
	}

	reward.PerformanceScore = overallScore
	reward.StakeWeight = stakeWeight

	reward.TotalReward = ComputeTotalReward(reward)

	return reward
}

// IdentityNetworkRewardInput represents input for identity network reward calculation
type IdentityNetworkRewardInput struct {
	// ValidatorAddress is the validator's address
	ValidatorAddress string

	// VerificationsCompleted is the number of verifications completed
	VerificationsCompleted int64

	// TotalVerifications is the total verifications in the epoch
	TotalVerifications int64

	// AverageVerificationScore is the average quality score
	AverageVerificationScore int64

	// RewardPool is the identity network reward pool
	RewardPool int64
}

// CalculateIdentityNetworkReward calculates identity network rewards
func CalculateIdentityNetworkReward(input IdentityNetworkRewardInput, denom string) sdk.Coins {
	if input.VerificationsCompleted <= 0 || input.TotalVerifications <= 0 || input.RewardPool <= 0 {
		return sdk.NewCoins()
	}

	score := clampRewardScore(input.AverageVerificationScore)
	numerator := big.NewInt(input.RewardPool)
	numerator.Mul(numerator, big.NewInt(input.VerificationsCompleted))
	numerator.Mul(numerator, big.NewInt(score))

	denominator := big.NewInt(input.TotalVerifications)
	denominator.Mul(denominator, big.NewInt(MaxPerformanceScore))

	amount := new(big.Int).Quo(numerator, denominator)
	return coinSetFromBigInt(denom, amount)
}

func rewardComponentScores(perf *stakingv1.ValidatorPerformance) (int64, int64, int64, int64) {
	if perf == nil {
		return 0, 0, MaxPerformanceScore, 0
	}

	blockScore := int64(0)
	if perf.BlocksExpected > 0 {
		blockScore = (perf.BlocksProposed * MaxPerformanceScore) / perf.BlocksExpected
		if blockScore > MaxPerformanceScore {
			blockScore = MaxPerformanceScore
		}
	} else if perf.BlocksProposed > 0 {
		blockScore = MaxPerformanceScore
	}

	veidScore := clampRewardScore(perf.VEIDVerificationScore)
	if perf.VEIDVerificationsExpected > 0 {
		completionRate := (perf.VEIDVerificationsCompleted * MaxPerformanceScore) / perf.VEIDVerificationsExpected
		if completionRate > MaxPerformanceScore {
			completionRate = MaxPerformanceScore
		}
		// Clamp quality score before combining so out-of-range inputs do not
		// inflate rewards and to keep keeper/type math consistent.
		veidScore = (completionRate + clampRewardScore(veidScore)) / 2
	}

	uptimeScore := MaxPerformanceScore
	totalTime := perf.UptimeSeconds + perf.DowntimeSeconds
	if totalTime > 0 {
		uptimeScore = (perf.UptimeSeconds * MaxPerformanceScore) / totalTime
	}

	overallScore := perf.OverallScore
	if overallScore <= 0 {
		overallScore = (blockScore*WeightBlockProposal + veidScore*WeightVEIDVerification + uptimeScore*WeightUptime) / TotalWeight
	}

	return clampRewardScore(blockScore), clampRewardScore(veidScore), clampRewardScore(uptimeScore), clampRewardScore(overallScore)
}

func clampRewardScore(score int64) int64 {
	if score < 0 {
		return 0
	}
	if score > MaxPerformanceScore {
		return MaxPerformanceScore
	}
	return score
}

func coinSetFromBigInt(denom string, amount *big.Int) sdk.Coins {
	if amount == nil || amount.Sign() <= 0 {
		return sdk.NewCoins()
	}
	return sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewIntFromBigInt(amount)))
}
func incrementCoinSet(coins sdk.Coins, denom string) sdk.Coins {
	amount := coins.AmountOf(denom)
	if amount.IsNil() {
		amount = sdkmath.ZeroInt()
	}
	return sdk.NewCoins(sdk.NewCoin(denom, amount.AddRaw(1)))
}

func formatFixedPointRatio(numerator, denominator int64) string {
	if numerator <= 0 || denominator <= 0 {
		return "0"
	}

	ratio := big.NewInt(numerator)
	ratio.Mul(ratio, big.NewInt(FixedPointScale))
	ratio.Div(ratio, big.NewInt(denominator))
	return ratio.String()
}
