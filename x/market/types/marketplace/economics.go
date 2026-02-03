// Package marketplace provides types for the marketplace on-chain module.
//
// ECON-002: Marketplace Economics Optimization
// This file defines the dynamic fee adjustment mechanism for the marketplace.
package marketplace

import (
	"fmt"
	"math"
	"time"
)

// FeeAdjustmentType defines how fees are adjusted
type FeeAdjustmentType uint8

const (
	// FeeAdjustmentTypeNone indicates no adjustment
	FeeAdjustmentTypeNone FeeAdjustmentType = 0

	// FeeAdjustmentTypeUtilization adjusts based on market utilization
	FeeAdjustmentTypeUtilization FeeAdjustmentType = 1

	// FeeAdjustmentTypeVolume adjusts based on volume
	FeeAdjustmentTypeVolume FeeAdjustmentType = 2

	// FeeAdjustmentTypeTime adjusts based on time of day/week
	FeeAdjustmentTypeTime FeeAdjustmentType = 3

	// FeeAdjustmentTypeTier adjusts based on user tier
	FeeAdjustmentTypeTier FeeAdjustmentType = 4
)

// FeeTier represents a tier-based fee discount level
type FeeTier uint8

const (
	// FeeTierStandard is the default tier with no discount
	FeeTierStandard FeeTier = 0

	// FeeTierBronze provides 10% fee discount
	FeeTierBronze FeeTier = 1

	// FeeTierSilver provides 20% fee discount
	FeeTierSilver FeeTier = 2

	// FeeTierGold provides 30% fee discount
	FeeTierGold FeeTier = 3

	// FeeTierPlatinum provides 40% fee discount
	FeeTierPlatinum FeeTier = 4

	// FeeTierDiamond provides 50% fee discount
	FeeTierDiamond FeeTier = 5
)

// FeeTierNames maps fee tiers to human-readable names
var FeeTierNames = map[FeeTier]string{
	FeeTierStandard: "standard",
	FeeTierBronze:   "bronze",
	FeeTierSilver:   "silver",
	FeeTierGold:     "gold",
	FeeTierPlatinum: "platinum",
	FeeTierDiamond:  "diamond",
}

// FeeTierDiscounts maps fee tiers to discount percentages (in basis points, 10000 = 100%)
var FeeTierDiscounts = map[FeeTier]uint32{
	FeeTierStandard: 0,    // 0%
	FeeTierBronze:   1000, // 10%
	FeeTierSilver:   2000, // 20%
	FeeTierGold:     3000, // 30%
	FeeTierPlatinum: 4000, // 40%
	FeeTierDiamond:  5000, // 50%
}

// FeeTierVolumeThresholds maps fee tiers to minimum 30-day volume thresholds (in smallest token unit)
var FeeTierVolumeThresholds = map[FeeTier]uint64{
	FeeTierStandard: 0,
	FeeTierBronze:   100000000,   // 100 tokens
	FeeTierSilver:   500000000,   // 500 tokens
	FeeTierGold:     2000000000,  // 2000 tokens
	FeeTierPlatinum: 10000000000, // 10000 tokens
	FeeTierDiamond:  50000000000, // 50000 tokens
}

// String returns the string representation of a FeeTier
func (t FeeTier) String() string {
	if name, ok := FeeTierNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// Discount returns the discount percentage in basis points
func (t FeeTier) Discount() uint32 {
	if discount, ok := FeeTierDiscounts[t]; ok {
		return discount
	}
	return 0
}

// VolumeThreshold returns the minimum 30-day volume for this tier
func (t FeeTier) VolumeThreshold() uint64 {
	if threshold, ok := FeeTierVolumeThresholds[t]; ok {
		return threshold
	}
	return 0
}

// FeeSchedule defines the fee structure for marketplace transactions
type FeeSchedule struct {
	// BaseTakeRateBps is the base take rate in basis points (10000 = 100%)
	BaseTakeRateBps uint32 `json:"base_take_rate_bps"`

	// MinTakeRateBps is the minimum take rate after all adjustments
	MinTakeRateBps uint32 `json:"min_take_rate_bps"`

	// MaxTakeRateBps is the maximum take rate after all adjustments
	MaxTakeRateBps uint32 `json:"max_take_rate_bps"`

	// MakerRebateBps is the rebate given to makers (liquidity providers) in basis points
	MakerRebateBps uint32 `json:"maker_rebate_bps"`

	// UtilizationMultiplierBps is the fee multiplier per 10% utilization above threshold
	UtilizationMultiplierBps uint32 `json:"utilization_multiplier_bps"`

	// UtilizationThresholdPct is the utilization percentage above which fees increase
	UtilizationThresholdPct uint32 `json:"utilization_threshold_pct"`

	// VolumeDiscountEnabled enables volume-based fee discounts
	VolumeDiscountEnabled bool `json:"volume_discount_enabled"`

	// TierDiscountEnabled enables tier-based fee discounts
	TierDiscountEnabled bool `json:"tier_discount_enabled"`

	// EarlyAdopterBonusBps is extra discount for early adopters in basis points
	EarlyAdopterBonusBps uint32 `json:"early_adopter_bonus_bps"`

	// EarlyAdopterCutoffTime is the cutoff timestamp for early adopter status
	EarlyAdopterCutoffTime time.Time `json:"early_adopter_cutoff_time,omitempty"`
}

// DefaultFeeSchedule returns the default fee schedule
func DefaultFeeSchedule() FeeSchedule {
	return FeeSchedule{
		BaseTakeRateBps:          200, // 2%
		MinTakeRateBps:           50,  // 0.5%
		MaxTakeRateBps:           500, // 5%
		MakerRebateBps:           25,  // 0.25%
		UtilizationMultiplierBps: 10,  // 0.1% per 10% utilization
		UtilizationThresholdPct:  70,  // Start increasing fees at 70% utilization
		VolumeDiscountEnabled:    true,
		TierDiscountEnabled:      true,
		EarlyAdopterBonusBps:     50, // 0.5% extra discount
	}
}

// Validate validates the fee schedule
func (fs *FeeSchedule) Validate() error {
	if fs.BaseTakeRateBps == 0 {
		return fmt.Errorf("base_take_rate_bps must be positive")
	}
	if fs.MinTakeRateBps > fs.BaseTakeRateBps {
		return fmt.Errorf("min_take_rate_bps cannot exceed base_take_rate_bps")
	}
	if fs.MaxTakeRateBps < fs.BaseTakeRateBps {
		return fmt.Errorf("max_take_rate_bps must be at least base_take_rate_bps")
	}
	if fs.MaxTakeRateBps > 10000 {
		return fmt.Errorf("max_take_rate_bps cannot exceed 10000 (100%%)")
	}
	if fs.UtilizationThresholdPct > 100 {
		return fmt.Errorf("utilization_threshold_pct cannot exceed 100")
	}
	return nil
}

// UtilizationMetrics captures market utilization data for fee calculation
type UtilizationMetrics struct {
	// TotalCapacity is the total available capacity in the market
	TotalCapacity uint64 `json:"total_capacity"`

	// UsedCapacity is the currently used capacity
	UsedCapacity uint64 `json:"used_capacity"`

	// ActiveOrders is the number of active orders
	ActiveOrders uint64 `json:"active_orders"`

	// ActiveProviders is the number of active providers
	ActiveProviders uint64 `json:"active_providers"`

	// TotalOrderVolume24h is the 24-hour order volume
	TotalOrderVolume24h uint64 `json:"total_order_volume_24h"`

	// AverageBidCount is the average number of bids per order
	AverageBidCount uint32 `json:"average_bid_count"`

	// LastUpdated is when metrics were last updated
	LastUpdated time.Time `json:"last_updated"`
}

// UtilizationPercent returns the utilization percentage
func (um *UtilizationMetrics) UtilizationPercent() uint32 {
	if um.TotalCapacity == 0 {
		return 0
	}
	percent := (um.UsedCapacity * 100) / um.TotalCapacity
	if percent > uint64(^uint32(0)) {
		return ^uint32(0)
	}
	return uint32(percent)
}

// FeeCalculationInput contains all inputs needed to calculate fees
type FeeCalculationInput struct {
	// OrderValue is the total order value
	OrderValue uint64 `json:"order_value"`

	// UserAddress is the user's address
	UserAddress string `json:"user_address"`

	// UserTier is the user's fee tier
	UserTier FeeTier `json:"user_tier"`

	// User30DayVolume is the user's 30-day trading volume
	User30DayVolume uint64 `json:"user_30day_volume"`

	// IsMaker indicates if the user is adding liquidity (maker) vs taking (taker)
	IsMaker bool `json:"is_maker"`

	// IsEarlyAdopter indicates if the user qualifies for early adopter bonus
	IsEarlyAdopter bool `json:"is_early_adopter"`

	// Utilization is the current market utilization metrics
	Utilization UtilizationMetrics `json:"utilization"`

	// Timestamp is when the fee is being calculated
	Timestamp time.Time `json:"timestamp"`
}

// FeeCalculationResult contains the calculated fee details
type FeeCalculationResult struct {
	// GrossFee is the fee before any adjustments
	GrossFee uint64 `json:"gross_fee"`

	// NetFee is the final fee after all adjustments
	NetFee uint64 `json:"net_fee"`

	// EffectiveRateBps is the effective rate in basis points
	EffectiveRateBps uint32 `json:"effective_rate_bps"`

	// TierDiscount is the discount from user tier
	TierDiscount uint64 `json:"tier_discount"`

	// VolumeDiscount is the discount from volume
	VolumeDiscount uint64 `json:"volume_discount"`

	// EarlyAdopterDiscount is the early adopter discount
	EarlyAdopterDiscount uint64 `json:"early_adopter_discount"`

	// MakerRebate is the rebate for makers (negative fee)
	MakerRebate uint64 `json:"maker_rebate"`

	// UtilizationAdjustment is the adjustment from utilization
	UtilizationAdjustment int64 `json:"utilization_adjustment"`

	// Breakdown provides a detailed breakdown of adjustments
	Breakdown map[string]int64 `json:"breakdown"`
}

// DynamicFeeCalculator calculates fees based on market conditions and user attributes
type DynamicFeeCalculator struct {
	Schedule FeeSchedule `json:"schedule"`
}

// NewDynamicFeeCalculator creates a new fee calculator with the given schedule
func NewDynamicFeeCalculator(schedule FeeSchedule) *DynamicFeeCalculator {
	return &DynamicFeeCalculator{Schedule: schedule}
}

// CalculateFee computes the fee for a given input
func (c *DynamicFeeCalculator) CalculateFee(input FeeCalculationInput) (*FeeCalculationResult, error) {
	if input.OrderValue == 0 {
		return &FeeCalculationResult{
			Breakdown: make(map[string]int64),
		}, nil
	}

	result := &FeeCalculationResult{
		Breakdown: make(map[string]int64),
	}

	// Calculate base fee
	baseFee := (input.OrderValue * uint64(c.Schedule.BaseTakeRateBps)) / 10000
	result.GrossFee = baseFee
	result.Breakdown["base_fee"] = safeInt64FromUint64Value(baseFee)

	// Apply utilization adjustment
	utilizationPct := input.Utilization.UtilizationPercent()
	if utilizationPct > c.Schedule.UtilizationThresholdPct {
		excessUtilization := utilizationPct - c.Schedule.UtilizationThresholdPct
		multiplierSteps := excessUtilization / 10
		adjustment := (baseFee * uint64(multiplierSteps) * uint64(c.Schedule.UtilizationMultiplierBps)) / 10000
		result.UtilizationAdjustment = safeInt64FromUint64Value(adjustment)
		result.Breakdown["utilization_adjustment"] = safeInt64FromUint64Value(adjustment)
	}

	// Apply tier discount
	if c.Schedule.TierDiscountEnabled && input.UserTier > FeeTierStandard {
		tierDiscount := (baseFee * uint64(input.UserTier.Discount())) / 10000
		result.TierDiscount = tierDiscount
		result.Breakdown["tier_discount"] = -safeInt64FromUint64Value(tierDiscount)
	}

	// Apply volume discount (additional to tier)
	if c.Schedule.VolumeDiscountEnabled {
		volumeDiscount := c.calculateVolumeDiscount(baseFee, input.User30DayVolume)
		result.VolumeDiscount = volumeDiscount
		result.Breakdown["volume_discount"] = -safeInt64FromUint64Value(volumeDiscount)
	}

	// Apply early adopter bonus
	if input.IsEarlyAdopter && c.Schedule.EarlyAdopterBonusBps > 0 {
		earlyAdopterDiscount := (baseFee * uint64(c.Schedule.EarlyAdopterBonusBps)) / 10000
		result.EarlyAdopterDiscount = earlyAdopterDiscount
		result.Breakdown["early_adopter_discount"] = -safeInt64FromUint64Value(earlyAdopterDiscount)
	}

	// Apply maker rebate
	if input.IsMaker && c.Schedule.MakerRebateBps > 0 {
		makerRebate := (input.OrderValue * uint64(c.Schedule.MakerRebateBps)) / 10000
		result.MakerRebate = makerRebate
		result.Breakdown["maker_rebate"] = -safeInt64FromUint64Value(makerRebate)
	}

	// Calculate net fee
	netFee := safeInt64FromUint64Value(baseFee) + result.UtilizationAdjustment -
		safeInt64FromUint64Value(result.TierDiscount) -
		safeInt64FromUint64Value(result.VolumeDiscount) -
		safeInt64FromUint64Value(result.EarlyAdopterDiscount) -
		safeInt64FromUint64Value(result.MakerRebate)

	// Apply min/max bounds
	minFee := safeInt64FromUint64Value((input.OrderValue * uint64(c.Schedule.MinTakeRateBps)) / 10000)
	maxFee := safeInt64FromUint64Value((input.OrderValue * uint64(c.Schedule.MaxTakeRateBps)) / 10000)

	if netFee < minFee {
		netFee = minFee
		result.Breakdown["min_bound_applied"] = 1
	}
	if netFee > maxFee {
		netFee = maxFee
		result.Breakdown["max_bound_applied"] = 1
	}

	// Handle negative fees (net rebate to maker)
	if netFee < 0 {
		netFee = 0
	}

	result.NetFee = safeUint64FromInt64Value(netFee)

	// Calculate effective rate
	if input.OrderValue > 0 {
		netFeeValue := safeUint64FromInt64Value(netFee)
		result.EffectiveRateBps = safeUint32FromUint64Value((netFeeValue * 10000) / input.OrderValue)
	}

	return result, nil
}

func safeUint64FromInt64Value(value int64) uint64 {
	if value < 0 {
		return 0
	}
	return uint64(value)
}

func safeInt64FromUint64Value(value uint64) int64 {
	if value > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(value)
}

func safeUint32FromUint64Value(value uint64) uint32 {
	if value > uint64(^uint32(0)) {
		return ^uint32(0)
	}
	//nolint:gosec // range checked above
	return uint32(value)
}

// calculateVolumeDiscount computes additional volume-based discount
func (c *DynamicFeeCalculator) calculateVolumeDiscount(baseFee, volume uint64) uint64 {
	// Volume discount tiers (additional to tier discount)
	// 0-1000 tokens: 0%
	// 1000-5000 tokens: 5%
	// 5000-20000 tokens: 10%
	// 20000+ tokens: 15%
	var discountBps uint32
	switch {
	case volume >= 20000000000: // 20000 tokens
		discountBps = 1500 // 15%
	case volume >= 5000000000: // 5000 tokens
		discountBps = 1000 // 10%
	case volume >= 1000000000: // 1000 tokens
		discountBps = 500 // 5%
	default:
		discountBps = 0
	}

	return (baseFee * uint64(discountBps)) / 10000
}

// GetTierForVolume returns the appropriate fee tier for a given 30-day volume
func GetTierForVolume(volume uint64) FeeTier {
	if volume >= FeeTierDiamond.VolumeThreshold() {
		return FeeTierDiamond
	}
	if volume >= FeeTierPlatinum.VolumeThreshold() {
		return FeeTierPlatinum
	}
	if volume >= FeeTierGold.VolumeThreshold() {
		return FeeTierGold
	}
	if volume >= FeeTierSilver.VolumeThreshold() {
		return FeeTierSilver
	}
	if volume >= FeeTierBronze.VolumeThreshold() {
		return FeeTierBronze
	}
	return FeeTierStandard
}

// UserFeeState tracks a user's fee-related state
type UserFeeState struct {
	// Address is the user's address
	Address string `json:"address"`

	// Tier is the user's current fee tier
	Tier FeeTier `json:"tier"`

	// Volume30Day is the user's 30-day trading volume
	Volume30Day uint64 `json:"volume_30day"`

	// TotalVolume is the user's all-time trading volume
	TotalVolume uint64 `json:"total_volume"`

	// TotalFeesPaid is the total fees paid by the user
	TotalFeesPaid uint64 `json:"total_fees_paid"`

	// TotalRebatesEarned is the total rebates earned by the user
	TotalRebatesEarned uint64 `json:"total_rebates_earned"`

	// IsEarlyAdopter indicates if the user is an early adopter
	IsEarlyAdopter bool `json:"is_early_adopter"`

	// FirstOrderAt is when the user placed their first order
	FirstOrderAt *time.Time `json:"first_order_at,omitempty"`

	// LastOrderAt is when the user placed their last order
	LastOrderAt *time.Time `json:"last_order_at,omitempty"`

	// LastTierUpdate is when the tier was last recalculated
	LastTierUpdate time.Time `json:"last_tier_update"`
}

// NewUserFeeState creates a new user fee state
func NewUserFeeState(address string) *UserFeeState {
	return &UserFeeState{
		Address: address,
		Tier:    FeeTierStandard,
	}
}

// UpdateTier recalculates and updates the user's tier based on volume
func (s *UserFeeState) UpdateTier(now time.Time) {
	s.Tier = GetTierForVolume(s.Volume30Day)
	s.LastTierUpdate = now
}

// RecordOrder updates the user's fee state after an order
func (s *UserFeeState) RecordOrder(orderValue, feePaid, rebateEarned uint64, now time.Time) {
	if s.FirstOrderAt == nil {
		s.FirstOrderAt = &now
	}
	s.LastOrderAt = &now
	s.Volume30Day += orderValue
	s.TotalVolume += orderValue
	s.TotalFeesPaid += feePaid
	s.TotalRebatesEarned += rebateEarned
	s.UpdateTier(now)
}

// EconomicsParams holds the economic parameters for the marketplace
type EconomicsParams struct {
	// FeeSchedule is the fee schedule configuration
	FeeSchedule FeeSchedule `json:"fee_schedule"`

	// FeeCollectorAddress is the address that collects fees
	FeeCollectorAddress string `json:"fee_collector_address"`

	// TreasuryShareBps is the percentage of fees that go to treasury
	TreasuryShareBps uint32 `json:"treasury_share_bps"`

	// StakingRewardsShareBps is the percentage of fees distributed to stakers
	StakingRewardsShareBps uint32 `json:"staking_rewards_share_bps"`

	// BurnShareBps is the percentage of fees that are burned
	BurnShareBps uint32 `json:"burn_share_bps"`

	// ProviderRewardsShareBps is the percentage of fees distributed to providers
	ProviderRewardsShareBps uint32 `json:"provider_rewards_share_bps"`

	// FeeUpdateIntervalBlocks is how often fees can be updated
	FeeUpdateIntervalBlocks int64 `json:"fee_update_interval_blocks"`

	// MinOrderValueForDiscount is the minimum order value to qualify for discounts
	MinOrderValueForDiscount uint64 `json:"min_order_value_for_discount"`
}

// DefaultEconomicsParams returns the default economics parameters
func DefaultEconomicsParams() EconomicsParams {
	return EconomicsParams{
		FeeSchedule:              DefaultFeeSchedule(),
		TreasuryShareBps:         2000, // 20%
		StakingRewardsShareBps:   5000, // 50%
		BurnShareBps:             1000, // 10%
		ProviderRewardsShareBps:  2000, // 20%
		FeeUpdateIntervalBlocks:  100,
		MinOrderValueForDiscount: 1000000, // 1 token minimum
	}
}

// Validate validates the economics parameters
func (p *EconomicsParams) Validate() error {
	if err := p.FeeSchedule.Validate(); err != nil {
		return fmt.Errorf("invalid fee schedule: %w", err)
	}

	totalShare := p.TreasuryShareBps + p.StakingRewardsShareBps + p.BurnShareBps + p.ProviderRewardsShareBps
	if totalShare != 10000 {
		return fmt.Errorf("fee distribution shares must sum to 10000 (100%%), got %d", totalShare)
	}

	if p.FeeUpdateIntervalBlocks <= 0 {
		return fmt.Errorf("fee_update_interval_blocks must be positive")
	}

	return nil
}

// FeeDistribution represents how a collected fee is distributed
type FeeDistribution struct {
	// TotalFee is the total fee collected
	TotalFee uint64 `json:"total_fee"`

	// TreasuryAmount is the amount going to treasury
	TreasuryAmount uint64 `json:"treasury_amount"`

	// StakingRewardsAmount is the amount for staking rewards
	StakingRewardsAmount uint64 `json:"staking_rewards_amount"`

	// BurnAmount is the amount to be burned
	BurnAmount uint64 `json:"burn_amount"`

	// ProviderRewardsAmount is the amount for provider rewards
	ProviderRewardsAmount uint64 `json:"provider_rewards_amount"`
}

// CalculateFeeDistribution computes how a fee should be distributed
func CalculateFeeDistribution(fee uint64, params EconomicsParams) FeeDistribution {
	return FeeDistribution{
		TotalFee:              fee,
		TreasuryAmount:        (fee * uint64(params.TreasuryShareBps)) / 10000,
		StakingRewardsAmount:  (fee * uint64(params.StakingRewardsShareBps)) / 10000,
		BurnAmount:            (fee * uint64(params.BurnShareBps)) / 10000,
		ProviderRewardsAmount: (fee * uint64(params.ProviderRewardsShareBps)) / 10000,
	}
}
