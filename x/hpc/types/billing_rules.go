// Package types contains types for the HPC module.
//
// VE-5A: Billing rules for deterministic cost calculation
package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CurrentBillingFormulaVersion is the current version of the billing formula
const CurrentBillingFormulaVersion = "v1.0.0"

// HPCBillingRules defines the billing rules for HPC resources
type HPCBillingRules struct {
	// FormulaVersion is the billing formula version
	FormulaVersion string `json:"formula_version"`

	// ResourceRates contains rates per resource type
	ResourceRates HPCResourceRates `json:"resource_rates"`

	// DiscountRules defines discount policies
	DiscountRules []HPCDiscountRule `json:"discount_rules,omitempty"`

	// BillingCaps defines spending caps
	BillingCaps []HPCBillingCap `json:"billing_caps,omitempty"`

	// PlatformFeeRateBps is the platform fee in basis points
	PlatformFeeRateBps uint32 `json:"platform_fee_rate_bps"`

	// ProviderRewardRateBps is the provider reward rate in basis points (of billable)
	ProviderRewardRateBps uint32 `json:"provider_reward_rate_bps"`

	// MinimumCharge is the minimum charge per job
	MinimumCharge sdk.Coin `json:"minimum_charge"`

	// BillingGranularitySeconds is the minimum billing granularity
	BillingGranularitySeconds int64 `json:"billing_granularity_seconds"`

	// QueueTimePenaltyEnabled enables queue time penalties/credits
	QueueTimePenaltyEnabled bool `json:"queue_time_penalty_enabled"`

	// QueueTimePenaltyThresholdSeconds is the threshold for queue time penalty
	QueueTimePenaltyThresholdSeconds int64 `json:"queue_time_penalty_threshold_seconds"`

	// QueueTimePenaltyRateBps is the penalty rate per minute over threshold
	QueueTimePenaltyRateBps uint32 `json:"queue_time_penalty_rate_bps"`
}

// HPCResourceRates defines rates for each resource type
type HPCResourceRates struct {
	// CPUCoreHourRate is the rate per CPU core-hour
	CPUCoreHourRate sdk.DecCoin `json:"cpu_core_hour_rate"`

	// GPUHourRate is the rate per GPU-hour (base rate)
	GPUHourRate sdk.DecCoin `json:"gpu_hour_rate"`

	// GPUTypeRates are rates for specific GPU types
	GPUTypeRates map[string]sdk.DecCoin `json:"gpu_type_rates,omitempty"`

	// MemoryGBHourRate is the rate per GB-hour of memory
	MemoryGBHourRate sdk.DecCoin `json:"memory_gb_hour_rate"`

	// NodeHourRate is the rate per node-hour
	NodeHourRate sdk.DecCoin `json:"node_hour_rate"`

	// StorageGBHourRate is the rate per GB-hour of storage
	StorageGBHourRate sdk.DecCoin `json:"storage_gb_hour_rate"`

	// NetworkGBRate is the rate per GB of network transfer
	NetworkGBRate sdk.DecCoin `json:"network_gb_rate"`

	// EnergyJouleRate is the rate per joule of energy (optional)
	EnergyJouleRate sdk.DecCoin `json:"energy_joule_rate,omitempty"`
}

// HPCDiscountRule defines a discount rule
type HPCDiscountRule struct {
	// RuleID is the unique identifier
	RuleID string `json:"rule_id"`

	// RuleName is the human-readable name
	RuleName string `json:"rule_name"`

	// DiscountType is the discount type
	DiscountType HPCDiscountType `json:"discount_type"`

	// DiscountBps is the discount in basis points
	DiscountBps uint32 `json:"discount_bps"`

	// ThresholdValue is the threshold for volume-based discounts
	ThresholdValue sdkmath.LegacyDec `json:"threshold_value,omitempty"`

	// ThresholdUnit is the unit for the threshold
	ThresholdUnit string `json:"threshold_unit,omitempty"`

	// AppliesTo specifies what resources this applies to
	AppliesTo []string `json:"applies_to,omitempty"`

	// RequiresCode indicates if a promo code is required
	RequiresCode bool `json:"requires_code"`

	// PromoCode is the promo code (if required)
	PromoCode string `json:"promo_code,omitempty"`

	// ValidFrom is when the rule becomes active
	ValidFrom int64 `json:"valid_from"`

	// ValidUntil is when the rule expires (0 = never)
	ValidUntil int64 `json:"valid_until"`

	// MaxUsesPerCustomer is the max uses per customer (0 = unlimited)
	MaxUsesPerCustomer int32 `json:"max_uses_per_customer"`

	// Active indicates if the rule is active
	Active bool `json:"active"`
}

// HPCDiscountType defines the type of discount
type HPCDiscountType string

const (
	// HPCDiscountTypeVolume is for volume-based discounts
	HPCDiscountTypeVolume HPCDiscountType = "volume"

	// HPCDiscountTypeLoyalty is for loyalty discounts
	HPCDiscountTypeLoyalty HPCDiscountType = "loyalty"

	// HPCDiscountTypePromo is for promotional discounts
	HPCDiscountTypePromo HPCDiscountType = "promo"

	// HPCDiscountTypeBundle is for bundle discounts
	HPCDiscountTypeBundle HPCDiscountType = "bundle"

	// HPCDiscountTypePartner is for partner discounts
	HPCDiscountTypePartner HPCDiscountType = "partner"
)

// IsValidHPCDiscountType checks if the discount type is valid
func IsValidHPCDiscountType(t HPCDiscountType) bool {
	switch t {
	case HPCDiscountTypeVolume, HPCDiscountTypeLoyalty, HPCDiscountTypePromo,
		HPCDiscountTypeBundle, HPCDiscountTypePartner:
		return true
	default:
		return false
	}
}

// HPCBillingCap defines a billing cap
type HPCBillingCap struct {
	// CapID is the unique identifier
	CapID string `json:"cap_id"`

	// CapName is the human-readable name
	CapName string `json:"cap_name"`

	// CapType is the cap type
	CapType HPCBillingCapType `json:"cap_type"`

	// CapAmount is the cap amount
	CapAmount sdk.Coin `json:"cap_amount"`

	// PeriodSeconds is the cap period in seconds (for periodic caps)
	PeriodSeconds int64 `json:"period_seconds,omitempty"`

	// AppliesTo specifies what resources this applies to
	AppliesTo []string `json:"applies_to,omitempty"`

	// PerCluster indicates if cap is per-cluster
	PerCluster bool `json:"per_cluster"`

	// Active indicates if the cap is active
	Active bool `json:"active"`
}

// HPCBillingCapType defines the type of billing cap
type HPCBillingCapType string

const (
	// HPCBillingCapTypePerJob is a per-job cap
	HPCBillingCapTypePerJob HPCBillingCapType = "per_job"

	// HPCBillingCapTypeDaily is a daily cap
	HPCBillingCapTypeDaily HPCBillingCapType = "daily"

	// HPCBillingCapTypeWeekly is a weekly cap
	HPCBillingCapTypeWeekly HPCBillingCapType = "weekly"

	// HPCBillingCapTypeMonthly is a monthly cap
	HPCBillingCapTypeMonthly HPCBillingCapType = "monthly"

	// HPCBillingCapTypeTotal is a total/lifetime cap
	HPCBillingCapTypeTotal HPCBillingCapType = "total"
)

// IsValidHPCBillingCapType checks if the cap type is valid
func IsValidHPCBillingCapType(t HPCBillingCapType) bool {
	switch t {
	case HPCBillingCapTypePerJob, HPCBillingCapTypeDaily, HPCBillingCapTypeWeekly,
		HPCBillingCapTypeMonthly, HPCBillingCapTypeTotal:
		return true
	default:
		return false
	}
}

// DefaultHPCBillingRules returns default billing rules
func DefaultHPCBillingRules(denom string) HPCBillingRules {
	return HPCBillingRules{
		FormulaVersion: CurrentBillingFormulaVersion,
		ResourceRates: HPCResourceRates{
			CPUCoreHourRate:   sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDecWithPrec(50, 3)),    // 0.05
			GPUHourRate:       sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(1)),                // 1.00
			MemoryGBHourRate:  sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDecWithPrec(10, 3)),    // 0.01
			NodeHourRate:      sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDecWithPrec(500, 3)),   // 0.50
			StorageGBHourRate: sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDecWithPrec(1, 3)),     // 0.001
			NetworkGBRate:     sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDecWithPrec(100, 3)),   // 0.10
			GPUTypeRates: map[string]sdk.DecCoin{
				"nvidia-a100": sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(3)),       // 3.00
				"nvidia-v100": sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(2)),       // 2.00
				"nvidia-t4":   sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDecWithPrec(5, 1)), // 0.50
			},
		},
		DiscountRules: []HPCDiscountRule{
			{
				RuleID:         "volume-100h",
				RuleName:       "Volume Discount - 100+ core-hours",
				DiscountType:   HPCDiscountTypeVolume,
				DiscountBps:    500, // 5%
				ThresholdValue: sdkmath.LegacyNewDec(100),
				ThresholdUnit:  "cpu_core_hours",
				Active:         true,
			},
			{
				RuleID:         "volume-1000h",
				RuleName:       "Volume Discount - 1000+ core-hours",
				DiscountType:   HPCDiscountTypeVolume,
				DiscountBps:    1000, // 10%
				ThresholdValue: sdkmath.LegacyNewDec(1000),
				ThresholdUnit:  "cpu_core_hours",
				Active:         true,
			},
		},
		BillingCaps: []HPCBillingCap{
			{
				CapID:    "monthly-cap",
				CapName:  "Monthly Spending Cap",
				CapType:  HPCBillingCapTypeMonthly,
				CapAmount: sdk.NewCoin(denom, sdkmath.NewInt(10000000000)), // 10000 virt
				Active:   true,
			},
		},
		PlatformFeeRateBps:               250, // 2.5%
		ProviderRewardRateBps:           9750, // 97.5% (100% - platform fee)
		MinimumCharge:                    sdk.NewCoin(denom, sdkmath.NewInt(1000)), // 0.001 virt
		BillingGranularitySeconds:        60, // 1 minute
		QueueTimePenaltyEnabled:          true,
		QueueTimePenaltyThresholdSeconds: 3600, // 1 hour
		QueueTimePenaltyRateBps:          10,   // 0.1% per minute
	}
}

// Validate validates the billing rules
func (r *HPCBillingRules) Validate() error {
	if r.FormulaVersion == "" {
		return fmt.Errorf("formula_version cannot be empty")
	}

	if r.PlatformFeeRateBps > 10000 {
		return fmt.Errorf("platform_fee_rate_bps cannot exceed 10000")
	}

	if r.ProviderRewardRateBps > 10000 {
		return fmt.Errorf("provider_reward_rate_bps cannot exceed 10000")
	}

	if r.PlatformFeeRateBps+r.ProviderRewardRateBps > 10000 {
		return fmt.Errorf("platform_fee + provider_reward cannot exceed 100%%")
	}

	if r.BillingGranularitySeconds < 1 {
		return fmt.Errorf("billing_granularity_seconds must be at least 1")
	}

	// Validate discount rules
	for i, rule := range r.DiscountRules {
		if rule.RuleID == "" {
			return fmt.Errorf("discount_rule[%d]: rule_id cannot be empty", i)
		}
		if !IsValidHPCDiscountType(rule.DiscountType) {
			return fmt.Errorf("discount_rule[%d]: invalid discount_type", i)
		}
		if rule.DiscountBps > 10000 {
			return fmt.Errorf("discount_rule[%d]: discount_bps cannot exceed 10000", i)
		}
	}

	// Validate billing caps
	for i, cap := range r.BillingCaps {
		if cap.CapID == "" {
			return fmt.Errorf("billing_cap[%d]: cap_id cannot be empty", i)
		}
		if !IsValidHPCBillingCapType(cap.CapType) {
			return fmt.Errorf("billing_cap[%d]: invalid cap_type", i)
		}
	}

	return nil
}

// HPCBillingCalculator calculates billable amounts
type HPCBillingCalculator struct {
	Rules HPCBillingRules
}

// NewHPCBillingCalculator creates a new billing calculator
func NewHPCBillingCalculator(rules HPCBillingRules) *HPCBillingCalculator {
	return &HPCBillingCalculator{Rules: rules}
}

// CalculateBillableAmount calculates the billable amount for given metrics
func (c *HPCBillingCalculator) CalculateBillableAmount(
	metrics *HPCDetailedMetrics,
	appliedDiscounts []AppliedDiscount,
	appliedCaps []AppliedCap,
) (*BillableBreakdown, sdk.Coins, error) {
	if metrics == nil {
		return nil, sdk.NewCoins(), nil
	}

	denom := c.Rules.ResourceRates.CPUCoreHourRate.Denom
	breakdown := BillableBreakdown{}

	// Calculate CPU cost (core-seconds -> core-hours)
	cpuCoreHours := sdkmath.LegacyNewDec(metrics.CPUCoreSeconds).Quo(sdkmath.LegacyNewDec(3600))
	cpuCost := cpuCoreHours.Mul(c.Rules.ResourceRates.CPUCoreHourRate.Amount)
	breakdown.CPUCost = sdk.NewCoin(denom, cpuCost.TruncateInt())

	// Calculate GPU cost
	gpuHours := sdkmath.LegacyNewDec(metrics.GPUSeconds).Quo(sdkmath.LegacyNewDec(3600))
	gpuRate := c.Rules.ResourceRates.GPUHourRate.Amount
	if metrics.GPUType != "" {
		if typeRate, ok := c.Rules.ResourceRates.GPUTypeRates[metrics.GPUType]; ok {
			gpuRate = typeRate.Amount
		}
	}
	gpuCost := gpuHours.Mul(gpuRate)
	breakdown.GPUCost = sdk.NewCoin(denom, gpuCost.TruncateInt())

	// Calculate memory cost (GB-seconds -> GB-hours)
	memGBHours := sdkmath.LegacyNewDec(metrics.MemoryGBSeconds).Quo(sdkmath.LegacyNewDec(3600))
	memCost := memGBHours.Mul(c.Rules.ResourceRates.MemoryGBHourRate.Amount)
	breakdown.MemoryCost = sdk.NewCoin(denom, memCost.TruncateInt())

	// Calculate node cost (handle uninitialized NodeHours)
	nodeHours := metrics.NodeHours
	if nodeHours.IsNil() {
		nodeHours = sdkmath.LegacyZeroDec()
	}
	nodeCost := nodeHours.Mul(c.Rules.ResourceRates.NodeHourRate.Amount)
	breakdown.NodeCost = sdk.NewCoin(denom, nodeCost.TruncateInt())

	// Calculate storage cost
	storageCost := sdkmath.LegacyNewDec(metrics.StorageGBHours).Mul(c.Rules.ResourceRates.StorageGBHourRate.Amount)
	breakdown.StorageCost = sdk.NewCoin(denom, storageCost.TruncateInt())

	// Calculate network cost (bytes -> GB)
	networkGB := sdkmath.LegacyNewDec(metrics.NetworkBytesIn + metrics.NetworkBytesOut).Quo(
		sdkmath.LegacyNewDec(1024 * 1024 * 1024))
	networkCost := networkGB.Mul(c.Rules.ResourceRates.NetworkGBRate.Amount)
	breakdown.NetworkCost = sdk.NewCoin(denom, networkCost.TruncateInt())

	// Calculate queue time penalty if enabled
	if c.Rules.QueueTimePenaltyEnabled && metrics.QueueTimeSeconds > c.Rules.QueueTimePenaltyThresholdSeconds {
		excessMinutes := (metrics.QueueTimeSeconds - c.Rules.QueueTimePenaltyThresholdSeconds) / 60
		// Penalty is applied as credit (negative cost) to customer
		penaltyBps := int64(c.Rules.QueueTimePenaltyRateBps) * excessMinutes
		if penaltyBps > 5000 { // Cap at 50%
			penaltyBps = 5000
		}
		// Queue penalty is a credit, so it's tracked but applied differently
		breakdown.QueuePenalty = sdk.NewCoin(denom, sdkmath.NewInt(0))
	}

	// Calculate subtotal
	subtotal := sdk.NewCoins(
		breakdown.CPUCost,
		breakdown.MemoryCost,
		breakdown.GPUCost,
		breakdown.NodeCost,
		breakdown.StorageCost,
		breakdown.NetworkCost,
	)

	// Consolidate coins with same denom
	consolidatedSubtotal := sdk.NewCoins()
	for _, coin := range subtotal {
		if coin.IsPositive() {
			consolidatedSubtotal = consolidatedSubtotal.Add(coin)
		}
	}
	breakdown.Subtotal = consolidatedSubtotal

	// Apply discounts
	totalDiscount := sdk.NewCoins()
	for _, discount := range appliedDiscounts {
		totalDiscount = totalDiscount.Add(discount.DiscountAmount...)
	}

	// Apply caps
	totalCapped := sdk.NewCoins()
	for _, cap := range appliedCaps {
		totalCapped = totalCapped.Add(cap.CappedAmount...)
	}

	// Calculate final amount: subtotal - discounts - caps
	finalAmount := consolidatedSubtotal
	if !totalDiscount.IsZero() {
		finalAmount = finalAmount.Sub(totalDiscount...)
	}
	if !totalCapped.IsZero() {
		finalAmount = finalAmount.Sub(totalCapped...)
	}

	// Apply minimum charge
	if finalAmount.IsZero() || finalAmount.AmountOf(denom).LT(c.Rules.MinimumCharge.Amount) {
		finalAmount = sdk.NewCoins(c.Rules.MinimumCharge)
	}

	return &breakdown, finalAmount, nil
}

// CalculateProviderReward calculates the provider reward from billable amount
func (c *HPCBillingCalculator) CalculateProviderReward(billableAmount sdk.Coins) sdk.Coins {
	denom := c.Rules.ResourceRates.CPUCoreHourRate.Denom
	total := billableAmount.AmountOf(denom)

	// Provider gets (100% - platform fee) of billable
	rewardMultiplier := sdkmath.LegacyNewDec(int64(c.Rules.ProviderRewardRateBps)).Quo(sdkmath.LegacyNewDec(10000))
	reward := sdkmath.LegacyNewDecFromInt(total).Mul(rewardMultiplier).TruncateInt()

	return sdk.NewCoins(sdk.NewCoin(denom, reward))
}

// CalculatePlatformFee calculates the platform fee from billable amount
func (c *HPCBillingCalculator) CalculatePlatformFee(billableAmount sdk.Coins) sdk.Coins {
	denom := c.Rules.ResourceRates.CPUCoreHourRate.Denom
	total := billableAmount.AmountOf(denom)

	feeMultiplier := sdkmath.LegacyNewDec(int64(c.Rules.PlatformFeeRateBps)).Quo(sdkmath.LegacyNewDec(10000))
	fee := sdkmath.LegacyNewDecFromInt(total).Mul(feeMultiplier).TruncateInt()

	return sdk.NewCoins(sdk.NewCoin(denom, fee))
}

// EvaluateDiscounts evaluates which discounts apply to given usage
func (c *HPCBillingCalculator) EvaluateDiscounts(
	metrics *HPCDetailedMetrics,
	customerAddress string,
	historicalUsage *AccountingAggregation,
) []AppliedDiscount {
	var appliedDiscounts []AppliedDiscount
	denom := c.Rules.ResourceRates.CPUCoreHourRate.Denom

	for _, rule := range c.Rules.DiscountRules {
		if !rule.Active {
			continue
		}

		// Check volume-based discounts
		if rule.DiscountType == HPCDiscountTypeVolume && historicalUsage != nil {
			var usageValue sdkmath.LegacyDec
			switch rule.ThresholdUnit {
			case "cpu_core_hours":
				usageValue = historicalUsage.TotalCPUCoreHours
			case "gpu_hours":
				usageValue = historicalUsage.TotalGPUHours
			case "node_hours":
				usageValue = historicalUsage.TotalNodeHours
			default:
				continue
			}

			if usageValue.GTE(rule.ThresholdValue) {
				// Calculate discount amount from current job
				cpuCoreHours := sdkmath.LegacyNewDec(metrics.CPUCoreSeconds).Quo(sdkmath.LegacyNewDec(3600))
				baseCost := cpuCoreHours.Mul(c.Rules.ResourceRates.CPUCoreHourRate.Amount)
				discountAmount := baseCost.Mul(sdkmath.LegacyNewDec(int64(rule.DiscountBps))).Quo(sdkmath.LegacyNewDec(10000))

				appliedDiscounts = append(appliedDiscounts, AppliedDiscount{
					DiscountID:     rule.RuleID,
					DiscountType:   string(rule.DiscountType),
					Description:    rule.RuleName,
					DiscountBps:    rule.DiscountBps,
					DiscountAmount: sdk.NewCoins(sdk.NewCoin(denom, discountAmount.TruncateInt())),
					AppliedTo:      "cpu",
				})
			}
		}
	}

	return appliedDiscounts
}

// EvaluateCaps evaluates which caps apply to given billing
func (c *HPCBillingCalculator) EvaluateCaps(
	billableAmount sdk.Coins,
	customerAddress string,
	periodSpending sdk.Coins,
) []AppliedCap {
	var appliedCaps []AppliedCap

	for _, cap := range c.Rules.BillingCaps {
		if !cap.Active {
			continue
		}

		denom := cap.CapAmount.Denom
		currentSpending := periodSpending.AmountOf(denom)
		newSpending := billableAmount.AmountOf(denom)
		totalSpending := currentSpending.Add(newSpending)

		if totalSpending.GT(cap.CapAmount.Amount) {
			// Cap exceeded - calculate capped amount
			allowedSpending := cap.CapAmount.Amount.Sub(currentSpending)
			if allowedSpending.IsNegative() {
				allowedSpending = sdkmath.ZeroInt()
			}
			cappedAmount := newSpending.Sub(allowedSpending)

			if cappedAmount.IsPositive() {
				appliedCaps = append(appliedCaps, AppliedCap{
					CapID:          cap.CapID,
					CapType:        string(cap.CapType),
					Description:    cap.CapName,
					CapAmount:      sdk.NewCoins(cap.CapAmount),
					OriginalAmount: sdk.NewCoins(sdk.NewCoin(denom, newSpending)),
					CappedAmount:   sdk.NewCoins(sdk.NewCoin(denom, cappedAmount)),
				})
			}
		}
	}

	return appliedCaps
}
