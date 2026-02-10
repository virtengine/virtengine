// Package types contains types for the staking module.
//
// VE-921: Reward types for validator staking rewards
// This file provides utility methods for reward types (generated proto types).
package types

import (
	"fmt"
	"time"

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

	if input.TotalStake == 0 || input.EpochRewardPool == 0 {
		return reward
	}

	// Calculate stake weight (fixed-point, 1e6 scale)
	stakeWeight := (input.StakeAmount * FixedPointScale) / input.TotalStake

	// Base stake reward (proportional to stake)
	baseReward := (input.EpochRewardPool * stakeWeight) / FixedPointScale

	// Performance multiplier (0.5 to 1.5x based on score)
	// Score 0 = 50% multiplier, Score 10000 = 150% multiplier
	performanceScore := int64(0)
	if input.Performance != nil {
		performanceScore = input.Performance.OverallScore
	}
	// Convert to multiplier: (5000 + score) / 10000 gives 0.5 to 1.5
	performanceMultiplier := 5000 + performanceScore

	// Apply performance multiplier
	adjustedReward := (baseReward * performanceMultiplier) / MaxPerformanceScore

	// Split reward into components based on weights
	blockReward := (adjustedReward * WeightBlockProposal) / TotalWeight
	veidReward := (adjustedReward * WeightVEIDVerification) / TotalWeight
	uptimeReward := (adjustedReward * WeightUptime) / TotalWeight

	// Additional VEID bonus for high-quality verifications
	var veidBonus int64
	if input.Performance != nil && input.Performance.VEIDVerificationScore >= 9000 {
		// 10% bonus for excellent VEID verification
		veidBonus = (veidReward * 1000) / MaxPerformanceScore
	}

	reward.BlockProposalReward = sdk.NewCoins(sdk.NewInt64Coin(denom, blockReward))
	reward.VEIDReward = sdk.NewCoins(sdk.NewInt64Coin(denom, veidReward+veidBonus))
	reward.UptimeReward = sdk.NewCoins(sdk.NewInt64Coin(denom, uptimeReward))
	reward.PerformanceScore = performanceScore
	reward.StakeWeight = fmt.Sprintf("%d", stakeWeight)

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
	if input.TotalVerifications == 0 || input.RewardPool == 0 {
		return sdk.NewCoins()
	}

	// Base share proportional to verifications completed
	baseShare := (input.RewardPool * input.VerificationsCompleted) / input.TotalVerifications

	// Quality multiplier (1.0 to 1.2x based on average score)
	// Score 0 = 1.0x, Score 10000 = 1.2x
	qualityMultiplier := MaxPerformanceScore + (input.AverageVerificationScore * 2000 / MaxPerformanceScore)
	adjustedReward := (baseShare * qualityMultiplier) / MaxPerformanceScore

	return sdk.NewCoins(sdk.NewInt64Coin(denom, adjustedReward))
}
