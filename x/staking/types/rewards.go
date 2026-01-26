// Package types contains types for the staking module.
//
// VE-921: Reward types for validator staking rewards
package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RewardType indicates the type of reward
type RewardType string

const (
	// RewardTypeBlockProposal is for block proposal rewards
	RewardTypeBlockProposal RewardType = "block_proposal"

	// RewardTypeVEIDVerification is for VEID verification rewards
	RewardTypeVEIDVerification RewardType = "veid_verification"

	// RewardTypeUptime is for uptime-based rewards
	RewardTypeUptime RewardType = "uptime"

	// RewardTypeIdentityNetwork is for identity network participation rewards
	RewardTypeIdentityNetwork RewardType = "identity_network"

	// RewardTypeStaking is for base staking rewards
	RewardTypeStaking RewardType = "staking"
)

// RewardEpoch represents a reward epoch
type RewardEpoch struct {
	// EpochNumber is the epoch identifier
	EpochNumber uint64 `json:"epoch_number"`

	// StartHeight is the starting block height
	StartHeight int64 `json:"start_height"`

	// EndHeight is the ending block height
	EndHeight int64 `json:"end_height"`

	// StartTime is when the epoch started
	StartTime time.Time `json:"start_time"`

	// EndTime is when the epoch ended (zero if current)
	EndTime time.Time `json:"end_time,omitempty"`

	// TotalRewardsDistributed is the total rewards distributed
	TotalRewardsDistributed sdk.Coins `json:"total_rewards_distributed"`

	// BlockProposalRewards is rewards from block proposals
	BlockProposalRewards sdk.Coins `json:"block_proposal_rewards"`

	// VEIDRewards is rewards from VEID verification work
	VEIDRewards sdk.Coins `json:"veid_rewards"`

	// UptimeRewards is rewards from uptime
	UptimeRewards sdk.Coins `json:"uptime_rewards"`

	// ValidatorCount is the number of validators in this epoch
	ValidatorCount int64 `json:"validator_count"`

	// TotalStake is the total stake in this epoch
	TotalStake string `json:"total_stake"`

	// Finalized indicates if this epoch is finalized
	Finalized bool `json:"finalized"`
}

// NewRewardEpoch creates a new reward epoch
func NewRewardEpoch(epochNumber uint64, startHeight int64, startTime time.Time) *RewardEpoch {
	return &RewardEpoch{
		EpochNumber:             epochNumber,
		StartHeight:             startHeight,
		StartTime:               startTime,
		TotalRewardsDistributed: sdk.NewCoins(),
		BlockProposalRewards:    sdk.NewCoins(),
		VEIDRewards:             sdk.NewCoins(),
		UptimeRewards:           sdk.NewCoins(),
	}
}

// Validate validates the reward epoch
func (e *RewardEpoch) Validate() error {
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

// Duration returns the epoch duration in blocks
func (e *RewardEpoch) Duration() int64 {
	if e.EndHeight == 0 {
		return 0
	}
	return e.EndHeight - e.StartHeight
}

// IsFinalized returns true if the epoch is finalized
func (e *RewardEpoch) IsFinalized() bool {
	return e.Finalized
}

// ValidatorReward represents a validator's rewards for an epoch
type ValidatorReward struct {
	// ValidatorAddress is the validator's address
	ValidatorAddress string `json:"validator_address"`

	// EpochNumber is the epoch this reward belongs to
	EpochNumber uint64 `json:"epoch_number"`

	// TotalReward is the total reward amount
	TotalReward sdk.Coins `json:"total_reward"`

	// BlockProposalReward is reward from block proposals
	BlockProposalReward sdk.Coins `json:"block_proposal_reward"`

	// VEIDReward is reward from VEID verification
	VEIDReward sdk.Coins `json:"veid_reward"`

	// UptimeReward is reward from uptime
	UptimeReward sdk.Coins `json:"uptime_reward"`

	// IdentityNetworkReward is reward from identity network participation
	IdentityNetworkReward sdk.Coins `json:"identity_network_reward"`

	// PerformanceScore is the performance score used for calculation
	PerformanceScore int64 `json:"performance_score"`

	// StakeWeight is the stake weight used for calculation
	StakeWeight string `json:"stake_weight"`

	// CalculatedAt is when the reward was calculated
	CalculatedAt time.Time `json:"calculated_at"`

	// BlockHeight is when the reward was recorded
	BlockHeight int64 `json:"block_height"`

	// Claimed indicates if the reward has been claimed
	Claimed bool `json:"claimed"`

	// ClaimedAt is when the reward was claimed
	ClaimedAt *time.Time `json:"claimed_at,omitempty"`
}

// NewValidatorReward creates a new validator reward
func NewValidatorReward(validatorAddr string, epochNumber uint64) *ValidatorReward {
	return &ValidatorReward{
		ValidatorAddress:      validatorAddr,
		EpochNumber:           epochNumber,
		TotalReward:           sdk.NewCoins(),
		BlockProposalReward:   sdk.NewCoins(),
		VEIDReward:            sdk.NewCoins(),
		UptimeReward:          sdk.NewCoins(),
		IdentityNetworkReward: sdk.NewCoins(),
	}
}

// Validate validates the validator reward
func (r *ValidatorReward) Validate() error {
	if r.ValidatorAddress == "" {
		return fmt.Errorf("validator_address cannot be empty")
	}

	if r.PerformanceScore < 0 || r.PerformanceScore > MaxPerformanceScore {
		return fmt.Errorf("performance_score must be between 0 and %d", MaxPerformanceScore)
	}

	return nil
}

// ComputeTotal computes the total reward from components
func (r *ValidatorReward) ComputeTotal() sdk.Coins {
	total := sdk.NewCoins()
	total = total.Add(r.BlockProposalReward...)
	total = total.Add(r.VEIDReward...)
	total = total.Add(r.UptimeReward...)
	total = total.Add(r.IdentityNetworkReward...)
	r.TotalReward = total
	return total
}

// RewardCalculationInput represents input for reward calculation
type RewardCalculationInput struct {
	// ValidatorAddress is the validator's address
	ValidatorAddress string

	// Performance is the validator's performance metrics
	Performance *ValidatorPerformance

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
func CalculateRewards(input RewardCalculationInput, denom string) *ValidatorReward {
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

	reward.ComputeTotal()

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
