// Package marketplace provides types for the marketplace on-chain module.
//
// ECON-002: Marketplace Economics Optimization
// This file defines market maker and liquidity provision incentives.
package marketplace

import (
	"fmt"
	"time"
)

// LiquidityTier represents the liquidity provider tier
type LiquidityTier uint8

const (
	// LiquidityTierNone indicates no liquidity tier
	LiquidityTierNone LiquidityTier = 0

	// LiquidityTierBronze is the entry-level liquidity tier
	LiquidityTierBronze LiquidityTier = 1

	// LiquidityTierSilver is the intermediate liquidity tier
	LiquidityTierSilver LiquidityTier = 2

	// LiquidityTierGold is the advanced liquidity tier
	LiquidityTierGold LiquidityTier = 3

	// LiquidityTierPlatinum is the expert liquidity tier
	LiquidityTierPlatinum LiquidityTier = 4

	// LiquidityTierDiamond is the elite liquidity tier
	LiquidityTierDiamond LiquidityTier = 5
)

// LiquidityTierNames maps liquidity tiers to human-readable names
var LiquidityTierNames = map[LiquidityTier]string{
	LiquidityTierNone:     "none",
	LiquidityTierBronze:   "bronze",
	LiquidityTierSilver:   "silver",
	LiquidityTierGold:     "gold",
	LiquidityTierPlatinum: "platinum",
	LiquidityTierDiamond:  "diamond",
}

// LiquidityTierRewardMultipliers maps tiers to reward multipliers (100 = 1x)
var LiquidityTierRewardMultipliers = map[LiquidityTier]uint32{
	LiquidityTierNone:     100, // 1x
	LiquidityTierBronze:   120, // 1.2x
	LiquidityTierSilver:   150, // 1.5x
	LiquidityTierGold:     200, // 2x
	LiquidityTierPlatinum: 300, // 3x
	LiquidityTierDiamond:  500, // 5x
}

// String returns the string representation of a LiquidityTier
func (t LiquidityTier) String() string {
	if name, ok := LiquidityTierNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// RewardMultiplier returns the reward multiplier for this tier
func (t LiquidityTier) RewardMultiplier() uint32 {
	if mult, ok := LiquidityTierRewardMultipliers[t]; ok {
		return mult
	}
	return 100
}

// MarketMakerStatus represents the market maker's status
type MarketMakerStatus uint8

const (
	// MarketMakerStatusInactive indicates an inactive market maker
	MarketMakerStatusInactive MarketMakerStatus = 0

	// MarketMakerStatusActive indicates an active market maker
	MarketMakerStatusActive MarketMakerStatus = 1

	// MarketMakerStatusSuspended indicates a suspended market maker
	MarketMakerStatusSuspended MarketMakerStatus = 2

	// MarketMakerStatusPenalized indicates a penalized market maker
	MarketMakerStatusPenalized MarketMakerStatus = 3
)

// MarketMakerStatusNames maps statuses to human-readable names
var MarketMakerStatusNames = map[MarketMakerStatus]string{
	MarketMakerStatusInactive:  "inactive",
	MarketMakerStatusActive:    "active",
	MarketMakerStatusSuspended: "suspended",
	MarketMakerStatusPenalized: "penalized",
}

// String returns the string representation of a MarketMakerStatus
func (s MarketMakerStatus) String() string {
	if name, ok := MarketMakerStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// SpreadTierConfig defines spread-based reward tiers
type SpreadTierConfig struct {
	// MaxSpreadBps is the maximum spread to qualify for this tier
	MaxSpreadBps uint32 `json:"max_spread_bps"`

	// RewardBonusBps is the reward bonus for this spread tier
	RewardBonusBps uint32 `json:"reward_bonus_bps"`
}

// MarketMakerConfig defines market maker incentive configuration
type MarketMakerConfig struct {
	// Enabled indicates if market maker rewards are enabled
	Enabled bool `json:"enabled"`

	// MinQuoteDuration is the minimum duration quotes must be maintained
	MinQuoteDuration time.Duration `json:"min_quote_duration"`

	// MinQuoteSize is the minimum quote size to qualify
	MinQuoteSize uint64 `json:"min_quote_size"`

	// MaxSpreadBps is the maximum spread to qualify for any rewards
	MaxSpreadBps uint32 `json:"max_spread_bps"`

	// BaseRewardPerBlock is the base reward per block for maintaining quotes
	BaseRewardPerBlock uint64 `json:"base_reward_per_block"`

	// SpreadTiers define spread-based reward tiers
	SpreadTiers []SpreadTierConfig `json:"spread_tiers"`

	// UptimeMinimumPct is the minimum uptime percentage for rewards
	UptimeMinimumPct uint32 `json:"uptime_minimum_pct"`

	// MaxRewardPerEpoch caps rewards per epoch per market maker
	MaxRewardPerEpoch uint64 `json:"max_reward_per_epoch"`

	// PenaltyForWithdrawal is the penalty for early quote withdrawal
	PenaltyForWithdrawalBps uint32 `json:"penalty_for_withdrawal_bps"`
}

// DefaultMarketMakerConfig returns the default market maker configuration
func DefaultMarketMakerConfig() MarketMakerConfig {
	return MarketMakerConfig{
		Enabled:            true,
		MinQuoteDuration:   time.Hour,
		MinQuoteSize:       1000000, // 1 token
		MaxSpreadBps:       500,     // 5% max spread
		BaseRewardPerBlock: 1000,
		SpreadTiers: []SpreadTierConfig{
			{MaxSpreadBps: 50, RewardBonusBps: 5000},  // <0.5% spread: 50% bonus
			{MaxSpreadBps: 100, RewardBonusBps: 2500}, // <1% spread: 25% bonus
			{MaxSpreadBps: 200, RewardBonusBps: 1000}, // <2% spread: 10% bonus
			{MaxSpreadBps: 500, RewardBonusBps: 0},    // <5% spread: no bonus
		},
		UptimeMinimumPct:        80,
		MaxRewardPerEpoch:       100000000, // 100 tokens per epoch
		PenaltyForWithdrawalBps: 1000,      // 10% penalty
	}
}

// Validate validates the market maker configuration
func (c *MarketMakerConfig) Validate() error {
	if c.MinQuoteDuration < 0 {
		return fmt.Errorf("min_quote_duration cannot be negative")
	}
	if c.MaxSpreadBps > 10000 {
		return fmt.Errorf("max_spread_bps cannot exceed 10000")
	}
	if c.UptimeMinimumPct > 100 {
		return fmt.Errorf("uptime_minimum_pct cannot exceed 100")
	}
	if c.PenaltyForWithdrawalBps > 10000 {
		return fmt.Errorf("penalty_for_withdrawal_bps cannot exceed 10000")
	}
	return nil
}

// LiquidityMiningConfig defines liquidity mining reward configuration
type LiquidityMiningConfig struct {
	// Enabled indicates if liquidity mining is enabled
	Enabled bool `json:"enabled"`

	// RewardTokenDenom is the token used for rewards
	RewardTokenDenom string `json:"reward_token_denom"`

	// RewardPerBlock is the reward distributed per block
	RewardPerBlock uint64 `json:"reward_per_block"`

	// MinLiquidityAmount is the minimum liquidity to qualify
	MinLiquidityAmount uint64 `json:"min_liquidity_amount"`

	// LockupPeriodBlocks is the minimum lockup period for bonus
	LockupPeriodBlocks int64 `json:"lockup_period_blocks"`

	// LockupBonusBps is the bonus for locking liquidity
	LockupBonusBps uint32 `json:"lockup_bonus_bps"`

	// EarlyUnlockPenaltyBps is the penalty for early unlock
	EarlyUnlockPenaltyBps uint32 `json:"early_unlock_penalty_bps"`

	// MaxTotalRewardsPerEpoch caps total rewards per epoch
	MaxTotalRewardsPerEpoch uint64 `json:"max_total_rewards_per_epoch"`

	// BoostMultiplierForStakers gives stakers bonus rewards (100 = 1x)
	BoostMultiplierForStakers uint32 `json:"boost_multiplier_for_stakers"`
}

// DefaultLiquidityMiningConfig returns the default liquidity mining configuration
func DefaultLiquidityMiningConfig() LiquidityMiningConfig {
	return LiquidityMiningConfig{
		Enabled:                   true,
		RewardTokenDenom:          "uakt",
		RewardPerBlock:            10000,       // 10000 tokens per block
		MinLiquidityAmount:        10000000,    // 10 tokens minimum
		LockupPeriodBlocks:        604800,      // ~42 days at 6s blocks
		LockupBonusBps:            2000,        // 20% bonus for lockup
		EarlyUnlockPenaltyBps:     5000,        // 50% penalty for early unlock
		MaxTotalRewardsPerEpoch:   10000000000, // 10000 tokens per epoch
		BoostMultiplierForStakers: 150,         // 1.5x for stakers
	}
}

// Validate validates the liquidity mining configuration
func (c *LiquidityMiningConfig) Validate() error {
	if c.RewardTokenDenom == "" {
		return fmt.Errorf("reward_token_denom is required")
	}
	if c.LockupBonusBps > 10000 {
		return fmt.Errorf("lockup_bonus_bps cannot exceed 10000")
	}
	if c.EarlyUnlockPenaltyBps > 10000 {
		return fmt.Errorf("early_unlock_penalty_bps cannot exceed 10000")
	}
	if c.BoostMultiplierForStakers == 0 {
		return fmt.Errorf("boost_multiplier_for_stakers must be positive")
	}
	return nil
}

// LiquidityPosition represents a liquidity provider's position
type LiquidityPosition struct {
	// ProviderAddress is the provider's address
	ProviderAddress string `json:"provider_address"`

	// Amount is the liquidity amount
	Amount uint64 `json:"amount"`

	// LockedUntilBlock is when the liquidity can be withdrawn
	LockedUntilBlock int64 `json:"locked_until_block,omitempty"`

	// IsLocked indicates if the position is locked
	IsLocked bool `json:"is_locked"`

	// CreatedAt is when the position was created
	CreatedAt time.Time `json:"created_at"`

	// LastRewardBlock is when rewards were last claimed
	LastRewardBlock int64 `json:"last_reward_block"`

	// AccumulatedRewards is the accumulated unclaimed rewards
	AccumulatedRewards uint64 `json:"accumulated_rewards"`

	// TotalRewardsClaimed is the total rewards claimed
	TotalRewardsClaimed uint64 `json:"total_rewards_claimed"`

	// Tier is the liquidity tier
	Tier LiquidityTier `json:"tier"`
}

// NewLiquidityPosition creates a new liquidity position
func NewLiquidityPosition(provider string, amount uint64, lockupBlocks int64, currentBlock int64, now time.Time) *LiquidityPosition {
	pos := &LiquidityPosition{
		ProviderAddress: provider,
		Amount:          amount,
		CreatedAt:       now,
		LastRewardBlock: currentBlock,
		Tier:            LiquidityTierNone,
	}

	if lockupBlocks > 0 {
		pos.IsLocked = true
		pos.LockedUntilBlock = currentBlock + lockupBlocks
	}

	pos.UpdateTier()
	return pos
}

// UpdateTier updates the liquidity tier based on amount
func (p *LiquidityPosition) UpdateTier() {
	switch {
	case p.Amount >= 1000000000000: // 1M tokens
		p.Tier = LiquidityTierDiamond
	case p.Amount >= 100000000000: // 100K tokens
		p.Tier = LiquidityTierPlatinum
	case p.Amount >= 10000000000: // 10K tokens
		p.Tier = LiquidityTierGold
	case p.Amount >= 1000000000: // 1K tokens
		p.Tier = LiquidityTierSilver
	case p.Amount >= 100000000: // 100 tokens
		p.Tier = LiquidityTierBronze
	default:
		p.Tier = LiquidityTierNone
	}
}

// CanWithdraw checks if the position can be withdrawn
func (p *LiquidityPosition) CanWithdraw(currentBlock int64) bool {
	if !p.IsLocked {
		return true
	}
	return currentBlock >= p.LockedUntilBlock
}

// MarketMaker represents a market maker in the system
type MarketMaker struct {
	// Address is the market maker's address
	Address string `json:"address"`

	// Status is the current status
	Status MarketMakerStatus `json:"status"`

	// Tier is the liquidity tier
	Tier LiquidityTier `json:"tier"`

	// TotalLiquidity is the total liquidity provided
	TotalLiquidity uint64 `json:"total_liquidity"`

	// ActiveQuotes is the number of active quotes
	ActiveQuotes uint32 `json:"active_quotes"`

	// TotalQuotesProvided is the total number of quotes provided
	TotalQuotesProvided uint64 `json:"total_quotes_provided"`

	// AverageSpreadBps is the average spread in basis points
	AverageSpreadBps uint32 `json:"average_spread_bps"`

	// UptimePercentage is the quote uptime percentage
	UptimePercentage uint32 `json:"uptime_percentage"`

	// TotalVolumeFacilitated is the total volume facilitated
	TotalVolumeFacilitated uint64 `json:"total_volume_facilitated"`

	// TotalRewardsEarned is the total rewards earned
	TotalRewardsEarned uint64 `json:"total_rewards_earned"`

	// PendingRewards is the pending unclaimed rewards
	PendingRewards uint64 `json:"pending_rewards"`

	// PenaltiesIncurred is the total penalties incurred
	PenaltiesIncurred uint64 `json:"penalties_incurred"`

	// RegisteredAt is when the market maker registered
	RegisteredAt time.Time `json:"registered_at"`

	// LastActiveAt is when the market maker was last active
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`

	// LastRewardCalculation is when rewards were last calculated
	LastRewardCalculation int64 `json:"last_reward_calculation"`

	// IsStaker indicates if the market maker is also a staker
	IsStaker bool `json:"is_staker"`

	// StakedAmount is the amount staked
	StakedAmount uint64 `json:"staked_amount"`
}

// NewMarketMaker creates a new market maker
func NewMarketMaker(address string, now time.Time) *MarketMaker {
	return &MarketMaker{
		Address:      address,
		Status:       MarketMakerStatusActive,
		Tier:         LiquidityTierNone,
		RegisteredAt: now,
	}
}

// QualifiesForRewards checks if the market maker qualifies for rewards
func (m *MarketMaker) QualifiesForRewards(config MarketMakerConfig) bool {
	if m.Status != MarketMakerStatusActive {
		return false
	}
	if m.UptimePercentage < config.UptimeMinimumPct {
		return false
	}
	if m.AverageSpreadBps > config.MaxSpreadBps {
		return false
	}
	if m.TotalLiquidity < config.MinQuoteSize {
		return false
	}
	return true
}

// Quote represents a market maker's quote
type Quote struct {
	// ID is the unique quote identifier
	ID string `json:"id"`

	// MarketMakerAddress is the market maker's address
	MarketMakerAddress string `json:"market_maker_address"`

	// OfferingID is the offering this quote is for
	OfferingID OfferingID `json:"offering_id"`

	// BidPrice is the bid price (buy)
	BidPrice uint64 `json:"bid_price"`

	// AskPrice is the ask price (sell)
	AskPrice uint64 `json:"ask_price"`

	// Size is the quote size
	Size uint64 `json:"size"`

	// SpreadBps is the spread in basis points
	SpreadBps uint32 `json:"spread_bps"`

	// CreatedAt is when the quote was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the quote expires
	ExpiresAt time.Time `json:"expires_at"`

	// IsActive indicates if the quote is active
	IsActive bool `json:"is_active"`

	// FilledAmount is the amount filled
	FilledAmount uint64 `json:"filled_amount"`
}

// NewQuote creates a new quote
func NewQuote(id, makerAddr string, offeringID OfferingID, bidPrice, askPrice, size uint64, duration time.Duration, now time.Time) *Quote {
	var spreadBps uint32
	if askPrice > 0 && bidPrice > 0 {
		spreadBps = uint32(((askPrice - bidPrice) * 10000) / askPrice)
	}

	return &Quote{
		ID:                 id,
		MarketMakerAddress: makerAddr,
		OfferingID:         offeringID,
		BidPrice:           bidPrice,
		AskPrice:           askPrice,
		Size:               size,
		SpreadBps:          spreadBps,
		CreatedAt:          now,
		ExpiresAt:          now.Add(duration),
		IsActive:           true,
	}
}

// IsValid checks if the quote is still valid
func (q *Quote) IsValid(now time.Time) bool {
	return q.IsActive && now.Before(q.ExpiresAt)
}

// RemainingSize returns the remaining unfilled size
func (q *Quote) RemainingSize() uint64 {
	if q.FilledAmount >= q.Size {
		return 0
	}
	return q.Size - q.FilledAmount
}

// LiquidityIncentiveCalculator calculates liquidity rewards
type LiquidityIncentiveCalculator struct {
	MarketMakerConfig     MarketMakerConfig     `json:"market_maker_config"`
	LiquidityMiningConfig LiquidityMiningConfig `json:"liquidity_mining_config"`
}

// NewLiquidityIncentiveCalculator creates a new calculator
func NewLiquidityIncentiveCalculator(mmConfig MarketMakerConfig, lmConfig LiquidityMiningConfig) *LiquidityIncentiveCalculator {
	return &LiquidityIncentiveCalculator{
		MarketMakerConfig:     mmConfig,
		LiquidityMiningConfig: lmConfig,
	}
}

// MarketMakerRewardCalculation represents a market maker reward calculation
type MarketMakerRewardCalculation struct {
	// Address is the market maker's address
	Address string `json:"address"`

	// BaseReward is the base reward amount
	BaseReward uint64 `json:"base_reward"`

	// SpreadBonus is the bonus for tight spreads
	SpreadBonus uint64 `json:"spread_bonus"`

	// TierMultiplier is the tier multiplier applied
	TierMultiplier uint32 `json:"tier_multiplier"`

	// StakerBoost is the staker boost applied
	StakerBoost uint64 `json:"staker_boost"`

	// FinalReward is the final reward after all adjustments
	FinalReward uint64 `json:"final_reward"`

	// BlockHeight is the block height
	BlockHeight int64 `json:"block_height"`

	// Reason describes the reward calculation
	Reason string `json:"reason"`
}

// CalculateMarketMakerReward calculates rewards for a market maker
func (c *LiquidityIncentiveCalculator) CalculateMarketMakerReward(mm *MarketMaker, blockHeight int64) *MarketMakerRewardCalculation {
	if !c.MarketMakerConfig.Enabled {
		return nil
	}

	if !mm.QualifiesForRewards(c.MarketMakerConfig) {
		return nil
	}

	result := &MarketMakerRewardCalculation{
		Address:     mm.Address,
		BlockHeight: blockHeight,
	}

	// Base reward
	result.BaseReward = c.MarketMakerConfig.BaseRewardPerBlock

	// Spread bonus
	for _, tier := range c.MarketMakerConfig.SpreadTiers {
		if mm.AverageSpreadBps <= tier.MaxSpreadBps {
			result.SpreadBonus = (result.BaseReward * uint64(tier.RewardBonusBps)) / 10000
			break
		}
	}

	// Tier multiplier
	result.TierMultiplier = mm.Tier.RewardMultiplier()
	baseWithBonus := result.BaseReward + result.SpreadBonus
	adjustedReward := (baseWithBonus * uint64(result.TierMultiplier)) / 100

	// Staker boost
	if mm.IsStaker && c.LiquidityMiningConfig.BoostMultiplierForStakers > 100 {
		result.StakerBoost = (adjustedReward * uint64(c.LiquidityMiningConfig.BoostMultiplierForStakers-100)) / 100
	}

	result.FinalReward = adjustedReward + result.StakerBoost
	result.Reason = fmt.Sprintf("spread_%dbps_tier_%s", mm.AverageSpreadBps, mm.Tier.String())

	return result
}

// LiquidityMiningRewardCalculation represents a liquidity mining reward calculation
type LiquidityMiningRewardCalculation struct {
	// ProviderAddress is the provider's address
	ProviderAddress string `json:"provider_address"`

	// BaseReward is the base reward
	BaseReward uint64 `json:"base_reward"`

	// LockupBonus is the lockup bonus
	LockupBonus uint64 `json:"lockup_bonus"`

	// TierMultiplier is the tier multiplier
	TierMultiplier uint32 `json:"tier_multiplier"`

	// FinalReward is the final reward
	FinalReward uint64 `json:"final_reward"`

	// BlocksElapsed is the number of blocks since last reward
	BlocksElapsed int64 `json:"blocks_elapsed"`

	// BlockHeight is the current block height
	BlockHeight int64 `json:"block_height"`
}

// CalculateLiquidityMiningReward calculates liquidity mining rewards
func (c *LiquidityIncentiveCalculator) CalculateLiquidityMiningReward(pos *LiquidityPosition, totalLiquidity uint64, currentBlock int64) *LiquidityMiningRewardCalculation {
	if !c.LiquidityMiningConfig.Enabled {
		return nil
	}

	if pos.Amount < c.LiquidityMiningConfig.MinLiquidityAmount {
		return nil
	}

	if totalLiquidity == 0 {
		return nil
	}

	result := &LiquidityMiningRewardCalculation{
		ProviderAddress: pos.ProviderAddress,
		BlockHeight:     currentBlock,
	}

	// Calculate blocks elapsed
	result.BlocksElapsed = currentBlock - pos.LastRewardBlock
	if result.BlocksElapsed <= 0 {
		return nil
	}

	// Pro-rata share of rewards
	sharePerBlock := (c.LiquidityMiningConfig.RewardPerBlock * pos.Amount) / totalLiquidity
	result.BaseReward = sharePerBlock * uint64(result.BlocksElapsed)

	// Lockup bonus
	if pos.IsLocked {
		result.LockupBonus = (result.BaseReward * uint64(c.LiquidityMiningConfig.LockupBonusBps)) / 10000
	}

	// Tier multiplier
	result.TierMultiplier = pos.Tier.RewardMultiplier()
	baseWithBonus := result.BaseReward + result.LockupBonus
	result.FinalReward = (baseWithBonus * uint64(result.TierMultiplier)) / 100

	return result
}

// LiquidityPoolStats tracks aggregate liquidity pool statistics
type LiquidityPoolStats struct {
	// TotalLiquidity is the total liquidity in the pool
	TotalLiquidity uint64 `json:"total_liquidity"`

	// TotalProviders is the number of liquidity providers
	TotalProviders uint64 `json:"total_providers"`

	// TotalMarketMakers is the number of active market makers
	TotalMarketMakers uint64 `json:"total_market_makers"`

	// TotalRewardsDistributed is the total rewards distributed
	TotalRewardsDistributed uint64 `json:"total_rewards_distributed"`

	// AverageAPY is the average APY in basis points
	AverageAPY uint32 `json:"average_apy_bps"`

	// TotalVolume24h is the 24-hour volume
	TotalVolume24h uint64 `json:"total_volume_24h"`

	// LockedLiquidity is the amount of locked liquidity
	LockedLiquidity uint64 `json:"locked_liquidity"`

	// UnlockedLiquidity is the amount of unlocked liquidity
	UnlockedLiquidity uint64 `json:"unlocked_liquidity"`

	// LastUpdated is when stats were last updated
	LastUpdated time.Time `json:"last_updated"`
}

// LiquidityIncentiveParams holds all liquidity incentive parameters
type LiquidityIncentiveParams struct {
	// MarketMakerConfig is the market maker configuration
	MarketMakerConfig MarketMakerConfig `json:"market_maker_config"`

	// LiquidityMiningConfig is the liquidity mining configuration
	LiquidityMiningConfig LiquidityMiningConfig `json:"liquidity_mining_config"`

	// RewardPoolAddress is the address holding reward funds
	RewardPoolAddress string `json:"reward_pool_address"`

	// EmissionRateDecayBps is the emission decay rate per epoch
	EmissionRateDecayBps uint32 `json:"emission_rate_decay_bps"`

	// MinTotalLiquidityForRewards is the minimum pool liquidity for rewards
	MinTotalLiquidityForRewards uint64 `json:"min_total_liquidity_for_rewards"`
}

// DefaultLiquidityIncentiveParams returns default liquidity incentive parameters
func DefaultLiquidityIncentiveParams() LiquidityIncentiveParams {
	return LiquidityIncentiveParams{
		MarketMakerConfig:           DefaultMarketMakerConfig(),
		LiquidityMiningConfig:       DefaultLiquidityMiningConfig(),
		EmissionRateDecayBps:        100,       // 1% decay per epoch
		MinTotalLiquidityForRewards: 100000000, // 100 tokens minimum
	}
}

// Validate validates the liquidity incentive parameters
func (p *LiquidityIncentiveParams) Validate() error {
	if err := p.MarketMakerConfig.Validate(); err != nil {
		return fmt.Errorf("invalid market maker config: %w", err)
	}
	if err := p.LiquidityMiningConfig.Validate(); err != nil {
		return fmt.Errorf("invalid liquidity mining config: %w", err)
	}
	if p.EmissionRateDecayBps > 10000 {
		return fmt.Errorf("emission_rate_decay_bps cannot exceed 10000")
	}
	return nil
}
