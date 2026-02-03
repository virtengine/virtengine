// Package marketplace provides types for the marketplace on-chain module.
//
// ECON-002: Marketplace Economics Optimization
// This file defines the provider incentive system for rewarding quality providers.
package marketplace

import (
	"fmt"
	"time"
)

// IncentiveType represents the type of incentive
type IncentiveType uint8

const (
	// IncentiveTypeNone indicates no incentive
	IncentiveTypeNone IncentiveType = 0

	// IncentiveTypeUptime rewards providers for maintaining high uptime
	IncentiveTypeUptime IncentiveType = 1

	// IncentiveTypeQuality rewards providers for quality of service
	IncentiveTypeQuality IncentiveType = 2

	// IncentiveTypeEarlyAdopter rewards early platform adopters
	IncentiveTypeEarlyAdopter IncentiveType = 3

	// IncentiveTypeVolume rewards high-volume providers
	IncentiveTypeVolume IncentiveType = 4

	// IncentiveTypeLoyalty rewards long-term providers
	IncentiveTypeLoyalty IncentiveType = 5

	// IncentiveTypeReferral rewards providers for referrals
	IncentiveTypeReferral IncentiveType = 6

	// IncentiveTypeStaking rewards providers who stake tokens
	IncentiveTypeStaking IncentiveType = 7
)

// IncentiveTypeNames maps incentive types to human-readable names
var IncentiveTypeNames = map[IncentiveType]string{
	IncentiveTypeNone:         "none",
	IncentiveTypeUptime:       "uptime",
	IncentiveTypeQuality:      "quality",
	IncentiveTypeEarlyAdopter: "early_adopter",
	IncentiveTypeVolume:       "volume",
	IncentiveTypeLoyalty:      "loyalty",
	IncentiveTypeReferral:     "referral",
	IncentiveTypeStaking:      "staking",
}

// String returns the string representation of an IncentiveType
func (t IncentiveType) String() string {
	if name, ok := IncentiveTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// ProviderTier represents the provider's performance tier
type ProviderTier uint8

const (
	// ProviderTierUnranked indicates an unranked provider
	ProviderTierUnranked ProviderTier = 0

	// ProviderTierBronze is the entry-level tier
	ProviderTierBronze ProviderTier = 1

	// ProviderTierSilver is the intermediate tier
	ProviderTierSilver ProviderTier = 2

	// ProviderTierGold is the advanced tier
	ProviderTierGold ProviderTier = 3

	// ProviderTierPlatinum is the expert tier
	ProviderTierPlatinum ProviderTier = 4

	// ProviderTierDiamond is the elite tier
	ProviderTierDiamond ProviderTier = 5
)

// ProviderTierNames maps provider tiers to human-readable names
var ProviderTierNames = map[ProviderTier]string{
	ProviderTierUnranked: "unranked",
	ProviderTierBronze:   "bronze",
	ProviderTierSilver:   "silver",
	ProviderTierGold:     "gold",
	ProviderTierPlatinum: "platinum",
	ProviderTierDiamond:  "diamond",
}

// ProviderTierMultipliers maps tiers to reward multipliers (100 = 1x, 200 = 2x)
var ProviderTierMultipliers = map[ProviderTier]uint32{
	ProviderTierUnranked: 100, // 1x
	ProviderTierBronze:   110, // 1.1x
	ProviderTierSilver:   125, // 1.25x
	ProviderTierGold:     150, // 1.5x
	ProviderTierPlatinum: 200, // 2x
	ProviderTierDiamond:  300, // 3x
}

// String returns the string representation of a ProviderTier
func (t ProviderTier) String() string {
	if name, ok := ProviderTierNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// Multiplier returns the reward multiplier for this tier (100 = 1x)
func (t ProviderTier) Multiplier() uint32 {
	if mult, ok := ProviderTierMultipliers[t]; ok {
		return mult
	}
	return 100
}

// UptimeRewardConfig defines uptime-based reward configuration
type UptimeRewardConfig struct {
	// MinUptimeForRewardPct is the minimum uptime percentage to qualify for rewards
	MinUptimeForRewardPct uint32 `json:"min_uptime_for_reward_pct"`

	// BaseRewardPerBlock is the base reward per block for meeting minimum uptime
	BaseRewardPerBlock uint64 `json:"base_reward_per_block"`

	// BonusPerPercentAboveMin is the bonus per percentage point above minimum
	BonusPerPercentAboveMin uint64 `json:"bonus_per_percent_above_min"`

	// PerfectUptimeBonusBps is the bonus for 100% uptime in basis points
	PerfectUptimeBonusBps uint32 `json:"perfect_uptime_bonus_bps"`

	// MeasurementWindowBlocks is the number of blocks for uptime measurement
	MeasurementWindowBlocks int64 `json:"measurement_window_blocks"`
}

// DefaultUptimeRewardConfig returns the default uptime reward configuration
func DefaultUptimeRewardConfig() UptimeRewardConfig {
	return UptimeRewardConfig{
		MinUptimeForRewardPct:   95,     // 95% minimum
		BaseRewardPerBlock:      1000,   // 1000 tokens per block
		BonusPerPercentAboveMin: 100,    // 100 tokens per % above 95%
		PerfectUptimeBonusBps:   500,    // 5% bonus for perfect uptime
		MeasurementWindowBlocks: 100800, // ~1 week at 6s blocks
	}
}

// QualityRewardConfig defines quality-of-service reward configuration
type QualityRewardConfig struct {
	// MinQualityScoreForReward is the minimum QoS score to qualify for rewards
	MinQualityScoreForReward uint32 `json:"min_quality_score_for_reward"`

	// MaxQualityScore is the maximum possible QoS score
	MaxQualityScore uint32 `json:"max_quality_score"`

	// BaseRewardPerOrder is the base reward per completed order
	BaseRewardPerOrder uint64 `json:"base_reward_per_order"`

	// QualityMultiplierBps maps quality score to reward multiplier
	QualityMultiplierBps uint32 `json:"quality_multiplier_bps"`

	// DisputePenaltyBps is the penalty for disputed orders in basis points
	DisputePenaltyBps uint32 `json:"dispute_penalty_bps"`
}

// DefaultQualityRewardConfig returns the default quality reward configuration
func DefaultQualityRewardConfig() QualityRewardConfig {
	return QualityRewardConfig{
		MinQualityScoreForReward: 70, // 70/100 minimum
		MaxQualityScore:          100,
		BaseRewardPerOrder:       10000, // 10000 tokens per order
		QualityMultiplierBps:     100,   // 1% multiplier per quality point
		DisputePenaltyBps:        5000,  // 50% penalty for disputes
	}
}

// VolumeRewardConfig defines volume-based reward configuration
type VolumeRewardConfig struct {
	// VolumeThresholds are the volume thresholds for reward tiers
	VolumeThresholds []uint64 `json:"volume_thresholds"`

	// RewardRatesBps are the reward rates for each tier in basis points
	RewardRatesBps []uint32 `json:"reward_rates_bps"`

	// MeasurementPeriodDays is the period for volume measurement
	MeasurementPeriodDays uint32 `json:"measurement_period_days"`
}

// DefaultVolumeRewardConfig returns the default volume reward configuration
func DefaultVolumeRewardConfig() VolumeRewardConfig {
	return VolumeRewardConfig{
		VolumeThresholds: []uint64{
			100000000,   // 100 tokens
			1000000000,  // 1000 tokens
			10000000000, // 10000 tokens
			50000000000, // 50000 tokens
		},
		RewardRatesBps: []uint32{
			10,  // 0.1%
			25,  // 0.25%
			50,  // 0.5%
			100, // 1%
		},
		MeasurementPeriodDays: 30,
	}
}

// LoyaltyRewardConfig defines loyalty-based reward configuration
type LoyaltyRewardConfig struct {
	// TenureThresholdsDays are the tenure thresholds for loyalty tiers
	TenureThresholdsDays []uint32 `json:"tenure_thresholds_days"`

	// BonusMultipliers are the bonus multipliers for each tier (100 = 1x)
	BonusMultipliers []uint32 `json:"bonus_multipliers"`

	// ConsecutiveActiveMonthsBonus is the bonus per consecutive active month
	ConsecutiveActiveMonthsBonus uint64 `json:"consecutive_active_months_bonus"`
}

// DefaultLoyaltyRewardConfig returns the default loyalty reward configuration
func DefaultLoyaltyRewardConfig() LoyaltyRewardConfig {
	return LoyaltyRewardConfig{
		TenureThresholdsDays: []uint32{
			30,  // 1 month
			90,  // 3 months
			180, // 6 months
			365, // 1 year
		},
		BonusMultipliers: []uint32{
			105, // 1.05x
			110, // 1.1x
			120, // 1.2x
			150, // 1.5x
		},
		ConsecutiveActiveMonthsBonus: 1000, // 1000 tokens per consecutive month
	}
}

// ProviderIncentiveConfig holds all incentive configurations
type ProviderIncentiveConfig struct {
	// Enabled indicates if provider incentives are enabled
	Enabled bool `json:"enabled"`

	// UptimeConfig is the uptime reward configuration
	UptimeConfig UptimeRewardConfig `json:"uptime_config"`

	// QualityConfig is the quality reward configuration
	QualityConfig QualityRewardConfig `json:"quality_config"`

	// VolumeConfig is the volume reward configuration
	VolumeConfig VolumeRewardConfig `json:"volume_config"`

	// LoyaltyConfig is the loyalty reward configuration
	LoyaltyConfig LoyaltyRewardConfig `json:"loyalty_config"`

	// EarlyAdopterCutoffTime is the cutoff for early adopter status
	EarlyAdopterCutoffTime time.Time `json:"early_adopter_cutoff_time,omitempty"`

	// EarlyAdopterBonusBps is the early adopter bonus in basis points
	EarlyAdopterBonusBps uint32 `json:"early_adopter_bonus_bps"`

	// MaxRewardPerBlockPerProvider caps rewards to prevent exploitation
	MaxRewardPerBlockPerProvider uint64 `json:"max_reward_per_block_per_provider"`

	// RewardPoolAddress is the address holding reward funds
	RewardPoolAddress string `json:"reward_pool_address"`

	// TotalRewardBudgetPerEpoch is the total reward budget per epoch
	TotalRewardBudgetPerEpoch uint64 `json:"total_reward_budget_per_epoch"`

	// EpochLengthBlocks is the length of a reward epoch in blocks
	EpochLengthBlocks int64 `json:"epoch_length_blocks"`
}

// DefaultProviderIncentiveConfig returns the default incentive configuration
func DefaultProviderIncentiveConfig() ProviderIncentiveConfig {
	return ProviderIncentiveConfig{
		Enabled:                      true,
		UptimeConfig:                 DefaultUptimeRewardConfig(),
		QualityConfig:                DefaultQualityRewardConfig(),
		VolumeConfig:                 DefaultVolumeRewardConfig(),
		LoyaltyConfig:                DefaultLoyaltyRewardConfig(),
		EarlyAdopterBonusBps:         1000,       // 10% bonus
		MaxRewardPerBlockPerProvider: 100000000,  // 100 tokens max per block
		TotalRewardBudgetPerEpoch:    1000000000, // 1000 tokens per epoch
		EpochLengthBlocks:            14400,      // ~1 day at 6s blocks
	}
}

// Validate validates the incentive configuration
func (c *ProviderIncentiveConfig) Validate() error {
	if c.UptimeConfig.MinUptimeForRewardPct > 100 {
		return fmt.Errorf("min_uptime_for_reward_pct cannot exceed 100")
	}
	if c.QualityConfig.MinQualityScoreForReward > c.QualityConfig.MaxQualityScore {
		return fmt.Errorf("min_quality_score_for_reward cannot exceed max_quality_score")
	}
	if len(c.VolumeConfig.VolumeThresholds) != len(c.VolumeConfig.RewardRatesBps) {
		return fmt.Errorf("volume thresholds and reward rates must have same length")
	}
	if len(c.LoyaltyConfig.TenureThresholdsDays) != len(c.LoyaltyConfig.BonusMultipliers) {
		return fmt.Errorf("tenure thresholds and bonus multipliers must have same length")
	}
	if c.EpochLengthBlocks <= 0 {
		return fmt.Errorf("epoch_length_blocks must be positive")
	}
	return nil
}

// ProviderMetrics tracks a provider's performance metrics
type ProviderMetrics struct {
	// Address is the provider's address
	Address string `json:"address"`

	// Tier is the provider's current tier
	Tier ProviderTier `json:"tier"`

	// UptimePercentage is the provider's uptime in the measurement window
	UptimePercentage uint32 `json:"uptime_percentage"`

	// UptimeBlocks is the number of blocks the provider was online
	UptimeBlocks int64 `json:"uptime_blocks"`

	// TotalBlocks is the total blocks in the measurement window
	TotalBlocks int64 `json:"total_blocks"`

	// QualityScore is the provider's current quality score (0-100)
	QualityScore uint32 `json:"quality_score"`

	// TotalOrders is the total number of orders fulfilled
	TotalOrders uint64 `json:"total_orders"`

	// SuccessfulOrders is the number of successfully completed orders
	SuccessfulOrders uint64 `json:"successful_orders"`

	// DisputedOrders is the number of disputed orders
	DisputedOrders uint64 `json:"disputed_orders"`

	// TotalVolume is the total order volume
	TotalVolume uint64 `json:"total_volume"`

	// Volume30Day is the 30-day rolling volume
	Volume30Day uint64 `json:"volume_30day"`

	// FirstActiveAt is when the provider first became active
	FirstActiveAt *time.Time `json:"first_active_at,omitempty"`

	// LastActiveAt is when the provider was last active
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`

	// ConsecutiveActiveMonths is the number of consecutive active months
	ConsecutiveActiveMonths uint32 `json:"consecutive_active_months"`

	// TotalRewardsEarned is the total rewards earned by the provider
	TotalRewardsEarned uint64 `json:"total_rewards_earned"`

	// PendingRewards is the pending rewards to be claimed
	PendingRewards uint64 `json:"pending_rewards"`

	// LastRewardCalculationBlock is when rewards were last calculated
	LastRewardCalculationBlock int64 `json:"last_reward_calculation_block"`

	// IsEarlyAdopter indicates if the provider is an early adopter
	IsEarlyAdopter bool `json:"is_early_adopter"`

	// StakedAmount is the amount staked by the provider
	StakedAmount uint64 `json:"staked_amount"`
}

// NewProviderMetrics creates new provider metrics
func NewProviderMetrics(address string) *ProviderMetrics {
	return &ProviderMetrics{
		Address: address,
		Tier:    ProviderTierUnranked,
	}
}

// CalculateSuccessRate returns the order success rate as a percentage
func (m *ProviderMetrics) CalculateSuccessRate() uint32 {
	if m.TotalOrders == 0 {
		return 0
	}
	rate := (m.SuccessfulOrders * 100) / m.TotalOrders
	if rate > uint64(^uint32(0)) {
		return ^uint32(0)
	}
	return uint32(rate)
}

// CalculateDisputeRate returns the dispute rate as a percentage
func (m *ProviderMetrics) CalculateDisputeRate() uint32 {
	if m.TotalOrders == 0 {
		return 0
	}
	rate := (m.DisputedOrders * 100) / m.TotalOrders
	if rate > uint64(^uint32(0)) {
		return ^uint32(0)
	}
	return uint32(rate)
}

// TenureDays returns the provider's tenure in days
func (m *ProviderMetrics) TenureDays(now time.Time) uint32 {
	if m.FirstActiveAt == nil {
		return 0
	}
	return uint32(now.Sub(*m.FirstActiveAt).Hours() / 24)
}

// RewardCalculation represents a single reward calculation
type RewardCalculation struct {
	// ProviderAddress is the provider's address
	ProviderAddress string `json:"provider_address"`

	// Type is the type of incentive
	Type IncentiveType `json:"type"`

	// BaseAmount is the base reward amount
	BaseAmount uint64 `json:"base_amount"`

	// Multiplier is the applied multiplier (100 = 1x)
	Multiplier uint32 `json:"multiplier"`

	// FinalAmount is the final reward after multiplier
	FinalAmount uint64 `json:"final_amount"`

	// Reason describes why this reward was given
	Reason string `json:"reason"`

	// CalculatedAt is when the reward was calculated
	CalculatedAt time.Time `json:"calculated_at"`

	// BlockHeight is the block height when calculated
	BlockHeight int64 `json:"block_height"`
}

// IncentiveCalculator calculates provider incentives
type IncentiveCalculator struct {
	Config ProviderIncentiveConfig `json:"config"`
}

// NewIncentiveCalculator creates a new incentive calculator
func NewIncentiveCalculator(config ProviderIncentiveConfig) *IncentiveCalculator {
	return &IncentiveCalculator{Config: config}
}

// CalculateUptimeReward calculates the uptime reward for a provider
func (c *IncentiveCalculator) CalculateUptimeReward(metrics *ProviderMetrics, blockHeight int64, now time.Time) *RewardCalculation {
	if metrics.UptimePercentage < c.Config.UptimeConfig.MinUptimeForRewardPct {
		return nil
	}

	// Base reward
	baseAmount := c.Config.UptimeConfig.BaseRewardPerBlock

	// Bonus for uptime above minimum
	excessUptime := metrics.UptimePercentage - c.Config.UptimeConfig.MinUptimeForRewardPct
	bonusAmount := uint64(excessUptime) * c.Config.UptimeConfig.BonusPerPercentAboveMin
	baseAmount += bonusAmount

	// Perfect uptime bonus
	if metrics.UptimePercentage == 100 {
		perfectBonus := (baseAmount * uint64(c.Config.UptimeConfig.PerfectUptimeBonusBps)) / 10000
		baseAmount += perfectBonus
	}

	// Apply tier multiplier
	multiplier := metrics.Tier.Multiplier()
	finalAmount := (baseAmount * uint64(multiplier)) / 100

	return &RewardCalculation{
		ProviderAddress: metrics.Address,
		Type:            IncentiveTypeUptime,
		BaseAmount:      baseAmount,
		Multiplier:      multiplier,
		FinalAmount:     finalAmount,
		Reason:          fmt.Sprintf("uptime_%d_pct", metrics.UptimePercentage),
		CalculatedAt:    now,
		BlockHeight:     blockHeight,
	}
}

// CalculateQualityReward calculates the quality reward for an order
func (c *IncentiveCalculator) CalculateQualityReward(metrics *ProviderMetrics, orderValue uint64, blockHeight int64, now time.Time) *RewardCalculation {
	if metrics.QualityScore < c.Config.QualityConfig.MinQualityScoreForReward {
		return nil
	}

	baseAmount := c.Config.QualityConfig.BaseRewardPerOrder

	// Quality score multiplier
	qualityMultiplier := 100 + ((metrics.QualityScore-c.Config.QualityConfig.MinQualityScoreForReward)*
		c.Config.QualityConfig.QualityMultiplierBps)/100

	// Apply tier multiplier
	tierMultiplier := metrics.Tier.Multiplier()
	combinedMultiplier := (qualityMultiplier * tierMultiplier) / 100

	finalAmount := (baseAmount * uint64(combinedMultiplier)) / 100

	return &RewardCalculation{
		ProviderAddress: metrics.Address,
		Type:            IncentiveTypeQuality,
		BaseAmount:      baseAmount,
		Multiplier:      combinedMultiplier,
		FinalAmount:     finalAmount,
		Reason:          fmt.Sprintf("quality_score_%d", metrics.QualityScore),
		CalculatedAt:    now,
		BlockHeight:     blockHeight,
	}
}

// CalculateVolumeReward calculates the volume-based reward
func (c *IncentiveCalculator) CalculateVolumeReward(metrics *ProviderMetrics, blockHeight int64, now time.Time) *RewardCalculation {
	if metrics.Volume30Day == 0 {
		return nil
	}

	// Find applicable tier
	var rewardRateBps uint32
	for i := len(c.Config.VolumeConfig.VolumeThresholds) - 1; i >= 0; i-- {
		if metrics.Volume30Day >= c.Config.VolumeConfig.VolumeThresholds[i] {
			rewardRateBps = c.Config.VolumeConfig.RewardRatesBps[i]
			break
		}
	}

	if rewardRateBps == 0 {
		return nil
	}

	baseAmount := (metrics.Volume30Day * uint64(rewardRateBps)) / 10000

	// Apply tier multiplier
	multiplier := metrics.Tier.Multiplier()
	finalAmount := (baseAmount * uint64(multiplier)) / 100

	return &RewardCalculation{
		ProviderAddress: metrics.Address,
		Type:            IncentiveTypeVolume,
		BaseAmount:      baseAmount,
		Multiplier:      multiplier,
		FinalAmount:     finalAmount,
		Reason:          fmt.Sprintf("volume_30d_%d", metrics.Volume30Day),
		CalculatedAt:    now,
		BlockHeight:     blockHeight,
	}
}

// CalculateLoyaltyReward calculates the loyalty-based reward
func (c *IncentiveCalculator) CalculateLoyaltyReward(metrics *ProviderMetrics, blockHeight int64, now time.Time) *RewardCalculation {
	tenureDays := metrics.TenureDays(now)
	if tenureDays == 0 {
		return nil
	}

	// Find applicable tier
	var bonusMultiplier uint32 = 100
	for i := len(c.Config.LoyaltyConfig.TenureThresholdsDays) - 1; i >= 0; i-- {
		if tenureDays >= c.Config.LoyaltyConfig.TenureThresholdsDays[i] {
			bonusMultiplier = c.Config.LoyaltyConfig.BonusMultipliers[i]
			break
		}
	}

	// Consecutive months bonus
	consecutiveBonus := uint64(metrics.ConsecutiveActiveMonths) * c.Config.LoyaltyConfig.ConsecutiveActiveMonthsBonus

	// Base amount from consecutive bonus
	baseAmount := consecutiveBonus

	// Apply loyalty multiplier
	finalAmount := (baseAmount * uint64(bonusMultiplier)) / 100

	if finalAmount == 0 {
		return nil
	}

	return &RewardCalculation{
		ProviderAddress: metrics.Address,
		Type:            IncentiveTypeLoyalty,
		BaseAmount:      baseAmount,
		Multiplier:      bonusMultiplier,
		FinalAmount:     finalAmount,
		Reason:          fmt.Sprintf("tenure_%d_days_consecutive_%d_months", tenureDays, metrics.ConsecutiveActiveMonths),
		CalculatedAt:    now,
		BlockHeight:     blockHeight,
	}
}

// CalculateAllRewards calculates all applicable rewards for a provider
func (c *IncentiveCalculator) CalculateAllRewards(metrics *ProviderMetrics, blockHeight int64, now time.Time) []*RewardCalculation {
	if !c.Config.Enabled {
		return nil
	}

	rewards := make([]*RewardCalculation, 0)

	if uptimeReward := c.CalculateUptimeReward(metrics, blockHeight, now); uptimeReward != nil {
		rewards = append(rewards, uptimeReward)
	}

	if volumeReward := c.CalculateVolumeReward(metrics, blockHeight, now); volumeReward != nil {
		rewards = append(rewards, volumeReward)
	}

	if loyaltyReward := c.CalculateLoyaltyReward(metrics, blockHeight, now); loyaltyReward != nil {
		rewards = append(rewards, loyaltyReward)
	}

	// Apply early adopter bonus to all rewards
	if metrics.IsEarlyAdopter && c.Config.EarlyAdopterBonusBps > 0 {
		for _, reward := range rewards {
			bonus := (reward.FinalAmount * uint64(c.Config.EarlyAdopterBonusBps)) / 10000
			reward.FinalAmount += bonus
			reward.Reason += "_early_adopter"
		}
	}

	// Cap individual rewards
	for _, reward := range rewards {
		if reward.FinalAmount > c.Config.MaxRewardPerBlockPerProvider {
			reward.FinalAmount = c.Config.MaxRewardPerBlockPerProvider
			reward.Reason += "_capped"
		}
	}

	return rewards
}

// TotalRewards calculates the total rewards from a list of calculations
func TotalRewards(calculations []*RewardCalculation) uint64 {
	var total uint64
	for _, calc := range calculations {
		total += calc.FinalAmount
	}
	return total
}

// UpdateProviderTier updates the provider's tier based on metrics
func UpdateProviderTier(metrics *ProviderMetrics) {
	// Calculate composite score
	uptimeScore := uint32(0)
	if metrics.UptimePercentage >= 99 {
		uptimeScore = 30
	} else if metrics.UptimePercentage >= 95 {
		uptimeScore = 20
	} else if metrics.UptimePercentage >= 90 {
		uptimeScore = 10
	}

	qualityScore := (metrics.QualityScore * 30) / 100

	volumeScore := uint32(0)
	if metrics.Volume30Day >= 50000000000 {
		volumeScore = 20
	} else if metrics.Volume30Day >= 10000000000 {
		volumeScore = 15
	} else if metrics.Volume30Day >= 1000000000 {
		volumeScore = 10
	} else if metrics.Volume30Day >= 100000000 {
		volumeScore = 5
	}

	successRateScore := (metrics.CalculateSuccessRate() * 20) / 100

	compositeScore := uptimeScore + qualityScore + volumeScore + successRateScore

	// Map to tier
	switch {
	case compositeScore >= 90:
		metrics.Tier = ProviderTierDiamond
	case compositeScore >= 75:
		metrics.Tier = ProviderTierPlatinum
	case compositeScore >= 60:
		metrics.Tier = ProviderTierGold
	case compositeScore >= 40:
		metrics.Tier = ProviderTierSilver
	case compositeScore >= 20:
		metrics.Tier = ProviderTierBronze
	default:
		metrics.Tier = ProviderTierUnranked
	}
}
