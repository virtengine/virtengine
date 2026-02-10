// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultCurrency is the default currency denomination
const DefaultCurrency = "uvirt"

// DecimalPrecision is the number of decimal places for calculations
const DecimalPrecision = 18

// CurrencyPrecisionMap defines decimal precision for known currencies
var CurrencyPrecisionMap = map[string]int{
	"uvirt": 6,  // micro-virt (1 virt = 1,000,000 uvirt)
	"nvirt": 9,  // nano-virt
	"avirt": 18, // atto-virt
	"uusd":  6,  // stablecoin
	"ibc/*": 6,  // IBC tokens default
}

// RoundingMode defines how fractional amounts are rounded
type RoundingMode uint8

const (
	// RoundingModeHalfEven is banker's rounding (round half to even)
	RoundingModeHalfEven RoundingMode = 0

	// RoundingModeHalfUp rounds half up (standard)
	RoundingModeHalfUp RoundingMode = 1

	// RoundingModeDown truncates (floor)
	RoundingModeDown RoundingMode = 2

	// RoundingModeUp rounds up (ceiling)
	RoundingModeUp RoundingMode = 3
)

// RoundingModeNames maps modes to names
var RoundingModeNames = map[RoundingMode]string{
	RoundingModeHalfEven: "half_even",
	RoundingModeHalfUp:   "half_up",
	RoundingModeDown:     "down",
	RoundingModeUp:       "up",
}

// String returns string representation
func (m RoundingMode) String() string {
	if name, ok := RoundingModeNames[m]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", m)
}

// PricingConfig defines pricing configuration
type PricingConfig struct {
	// RoundingMode is the rounding mode for calculations
	RoundingMode RoundingMode `json:"rounding_mode"`

	// MinimumCharge is the minimum charge per invoice
	MinimumCharge sdk.Coin `json:"minimum_charge"`

	// GracePeriodSeconds is the grace period before overdue
	GracePeriodSeconds int64 `json:"grace_period_seconds"`

	// DefaultPaymentTermDays is the default payment term
	DefaultPaymentTermDays int32 `json:"default_payment_term_days"`

	// LateFeePercentBps is late fee as basis points
	LateFeePercentBps uint32 `json:"late_fee_percent_bps"`

	// MaxLateFeePercent is maximum late fee percentage
	MaxLateFeePercentBps uint32 `json:"max_late_fee_percent_bps"`
}

// DefaultPricingConfig returns default pricing configuration
func DefaultPricingConfig() PricingConfig {
	return PricingConfig{
		RoundingMode:           RoundingModeHalfEven,
		MinimumCharge:          sdk.NewCoin(DefaultCurrency, sdkmath.NewInt(1000)), // 0.001 virt
		GracePeriodSeconds:     86400,                                              // 1 day
		DefaultPaymentTermDays: 7,
		LateFeePercentBps:      100, // 1%
		MaxLateFeePercentBps:   500, // 5%
	}
}

// Validate validates the pricing configuration
func (c *PricingConfig) Validate() error {
	if c.DefaultPaymentTermDays < 0 {
		return fmt.Errorf("default_payment_term_days cannot be negative")
	}
	if c.LateFeePercentBps > 10000 {
		return fmt.Errorf("late_fee_percent_bps cannot exceed 10000")
	}
	if c.MaxLateFeePercentBps > 10000 {
		return fmt.Errorf("max_late_fee_percent_bps cannot exceed 10000")
	}
	if c.LateFeePercentBps > c.MaxLateFeePercentBps {
		return fmt.Errorf("late_fee_percent_bps cannot exceed max_late_fee_percent_bps")
	}
	return nil
}

// ResourcePricing defines pricing for a resource type
type ResourcePricing struct {
	// ResourceType is the type of resource
	ResourceType UsageType `json:"resource_type"`

	// BaseRate is the base rate per unit
	BaseRate sdk.DecCoin `json:"base_rate"`

	// Unit is the pricing unit
	Unit string `json:"unit"`

	// MinQuantity is the minimum billable quantity
	MinQuantity sdkmath.LegacyDec `json:"min_quantity"`

	// GranularitySeconds is the billing granularity
	GranularitySeconds int64 `json:"granularity_seconds"`

	// TierPricing defines volume-based tiers
	TierPricing []PricingTier `json:"tier_pricing,omitempty"`
}

// PricingTier defines a volume-based pricing tier
type PricingTier struct {
	// TierID is the tier identifier
	TierID string `json:"tier_id"`

	// TierName is the human-readable tier name
	TierName string `json:"tier_name"`

	// MinQuantity is the minimum quantity for this tier
	MinQuantity sdkmath.LegacyDec `json:"min_quantity"`

	// MaxQuantity is the maximum quantity for this tier (0 = unlimited)
	MaxQuantity sdkmath.LegacyDec `json:"max_quantity"`

	// Rate is the rate for this tier
	Rate sdk.DecCoin `json:"rate"`

	// DiscountBps is the discount in basis points
	DiscountBps uint32 `json:"discount_bps"`
}

// GetEffectiveRate returns the effective rate after discount
func (t *PricingTier) GetEffectiveRate() sdk.DecCoin {
	if t.DiscountBps == 0 {
		return t.Rate
	}

	// Calculate discount multiplier: (10000 - discount) / 10000
	discountMultiplier := sdkmath.LegacyNewDec(int64(10000 - t.DiscountBps)).Quo(sdkmath.LegacyNewDec(10000))
	discountedAmount := t.Rate.Amount.Mul(discountMultiplier)
	return sdk.NewDecCoinFromDec(t.Rate.Denom, discountedAmount)
}

// CalculateAmount calculates the amount for a given quantity using tier pricing
func (rp *ResourcePricing) CalculateAmount(quantity sdkmath.LegacyDec, mode RoundingMode) sdk.Coin {
	if quantity.LT(rp.MinQuantity) {
		quantity = rp.MinQuantity
	}

	var totalAmount sdkmath.LegacyDec

	// If no tiers, use base rate
	if len(rp.TierPricing) == 0 {
		totalAmount = quantity.Mul(rp.BaseRate.Amount)
	} else {
		// Apply tiered pricing
		remainingQty := quantity
		totalAmount = sdkmath.LegacyZeroDec()

		for _, tier := range rp.TierPricing {
			if remainingQty.IsZero() || remainingQty.IsNegative() {
				break
			}

			tierQty := remainingQty
			if !tier.MaxQuantity.IsZero() && remainingQty.GT(tier.MaxQuantity.Sub(tier.MinQuantity)) {
				tierQty = tier.MaxQuantity.Sub(tier.MinQuantity)
			}

			effectiveRate := tier.GetEffectiveRate()
			tierAmount := tierQty.Mul(effectiveRate.Amount)
			totalAmount = totalAmount.Add(tierAmount)
			remainingQty = remainingQty.Sub(tierQty)
		}
	}

	// Apply rounding
	roundedAmount := applyRounding(totalAmount, mode)
	return sdk.NewCoin(rp.BaseRate.Denom, roundedAmount)
}

// applyRounding applies the specified rounding mode
func applyRounding(amount sdkmath.LegacyDec, mode RoundingMode) sdkmath.Int {
	switch mode {
	case RoundingModeHalfEven:
		return roundHalfEven(amount)
	case RoundingModeHalfUp:
		return amount.RoundInt()
	case RoundingModeDown:
		return amount.TruncateInt()
	case RoundingModeUp:
		return amount.Ceil().TruncateInt()
	default:
		return amount.TruncateInt()
	}
}

// roundHalfEven implements banker's rounding
func roundHalfEven(amount sdkmath.LegacyDec) sdkmath.Int {
	truncated := amount.TruncateInt()
	fraction := amount.Sub(sdkmath.LegacyNewDecFromInt(truncated))

	half := sdkmath.LegacyNewDecWithPrec(5, 1) // 0.5

	if fraction.GT(half) {
		return truncated.Add(sdkmath.OneInt())
	} else if fraction.LT(half) {
		return truncated
	}

	// Exactly 0.5: round to even
	if truncated.Mod(sdkmath.NewInt(2)).IsZero() {
		return truncated
	}
	return truncated.Add(sdkmath.OneInt())
}

// CurrencyConversion tracks a currency conversion
type CurrencyConversion struct {
	// FromCurrency is the source currency
	FromCurrency string `json:"from_currency"`

	// ToCurrency is the target currency
	ToCurrency string `json:"to_currency"`

	// FromAmount is the source amount
	FromAmount sdk.Coin `json:"from_amount"`

	// ToAmount is the converted amount
	ToAmount sdk.Coin `json:"to_amount"`

	// ExchangeRate is the rate used
	ExchangeRate sdkmath.LegacyDec `json:"exchange_rate"`

	// ConversionFee is any fee charged
	ConversionFee sdk.Coin `json:"conversion_fee,omitempty"`

	// RateSource is the source of the exchange rate
	RateSource string `json:"rate_source"`

	// RateTimestamp is when the rate was obtained
	RateTimestamp int64 `json:"rate_timestamp"`
}

// Validate validates the currency conversion
func (c *CurrencyConversion) Validate() error {
	if c.FromCurrency == "" {
		return fmt.Errorf("from_currency is required")
	}
	if c.ToCurrency == "" {
		return fmt.Errorf("to_currency is required")
	}
	if !c.FromAmount.IsPositive() {
		return fmt.Errorf("from_amount must be positive")
	}
	if !c.ToAmount.IsPositive() {
		return fmt.Errorf("to_amount must be positive")
	}
	if !c.ExchangeRate.IsPositive() {
		return fmt.Errorf("exchange_rate must be positive")
	}
	return nil
}

// PricingPolicy defines the pricing policy for a provider or deployment
type PricingPolicy struct {
	// PolicyID is the unique identifier
	PolicyID string `json:"policy_id"`

	// Provider is the provider address
	Provider string `json:"provider"`

	// ResourcePricing maps resource types to pricing
	ResourcePricing map[UsageType]ResourcePricing `json:"resource_pricing"`

	// PricingConfig is the pricing configuration
	PricingConfig PricingConfig `json:"pricing_config"`

	// SupportedCurrencies lists accepted currencies
	SupportedCurrencies []string `json:"supported_currencies"`

	// DefaultCurrency is the default billing currency
	DefaultCurrency string `json:"default_currency"`

	// EffectiveFrom is when this policy takes effect
	EffectiveFrom int64 `json:"effective_from"`

	// EffectiveUntil is when this policy expires (0 = no expiry)
	EffectiveUntil int64 `json:"effective_until"`
}

// Validate validates the pricing policy
func (p *PricingPolicy) Validate() error {
	if p.PolicyID == "" {
		return fmt.Errorf("policy_id is required")
	}

	if _, err := sdk.AccAddressFromBech32(p.Provider); err != nil {
		return fmt.Errorf("invalid provider address: %w", err)
	}

	if len(p.SupportedCurrencies) == 0 {
		return fmt.Errorf("at least one supported currency is required")
	}

	if p.DefaultCurrency == "" {
		return fmt.Errorf("default_currency is required")
	}

	// Validate default currency is in supported list
	found := false
	for _, c := range p.SupportedCurrencies {
		if c == p.DefaultCurrency {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("default_currency must be in supported_currencies")
	}

	if err := p.PricingConfig.Validate(); err != nil {
		return fmt.Errorf("pricing_config: %w", err)
	}

	return nil
}

// IsActive returns true if the policy is currently active
func (p *PricingPolicy) IsActive(blockHeight int64) bool {
	if blockHeight < p.EffectiveFrom {
		return false
	}
	if p.EffectiveUntil > 0 && blockHeight > p.EffectiveUntil {
		return false
	}
	return true
}

// GetResourcePricing returns pricing for a resource type
func (p *PricingPolicy) GetResourcePricing(t UsageType) (ResourcePricing, bool) {
	rp, ok := p.ResourcePricing[t]
	return rp, ok
}

// DefaultResourcePricing returns default pricing for common resource types
func DefaultResourcePricing(currency string) map[UsageType]ResourcePricing {
	return map[UsageType]ResourcePricing{
		UsageTypeCPU: {
			ResourceType:       UsageTypeCPU,
			BaseRate:           sdk.NewDecCoinFromDec(currency, sdkmath.LegacyNewDecWithPrec(10, 3)), // 0.01 per core-hour
			Unit:               "core-hour",
			MinQuantity:        sdkmath.LegacyNewDecWithPrec(1, 1), // 0.1 core-hour minimum
			GranularitySeconds: 3600,                               // 1 hour
		},
		UsageTypeMemory: {
			ResourceType:       UsageTypeMemory,
			BaseRate:           sdk.NewDecCoinFromDec(currency, sdkmath.LegacyNewDecWithPrec(5, 3)), // 0.005 per GB-hour
			Unit:               "gb-hour",
			MinQuantity:        sdkmath.LegacyNewDecWithPrec(1, 1), // 0.1 GB-hour minimum
			GranularitySeconds: 3600,
		},
		UsageTypeStorage: {
			ResourceType:       UsageTypeStorage,
			BaseRate:           sdk.NewDecCoinFromDec(currency, sdkmath.LegacyNewDecWithPrec(1, 3)), // 0.001 per GB-month
			Unit:               "gb-month",
			MinQuantity:        sdkmath.LegacyOneDec(),
			GranularitySeconds: 86400, // 1 day
		},
		UsageTypeNetwork: {
			ResourceType:       UsageTypeNetwork,
			BaseRate:           sdk.NewDecCoinFromDec(currency, sdkmath.LegacyNewDecWithPrec(1, 4)), // 0.0001 per GB
			Unit:               "gb",
			MinQuantity:        sdkmath.LegacyOneDec(),
			GranularitySeconds: 0, // usage-based
		},
		UsageTypeGPU: {
			ResourceType:       UsageTypeGPU,
			BaseRate:           sdk.NewDecCoinFromDec(currency, sdkmath.LegacyNewDecWithPrec(100, 3)), // 0.1 per GPU-hour
			Unit:               "gpu-hour",
			MinQuantity:        sdkmath.LegacyOneDec(),
			GranularitySeconds: 3600,
		},
	}
}
