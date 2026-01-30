// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DiscountType defines the type of discount
type DiscountType uint8

const (
	// DiscountTypePercentage is a percentage discount
	DiscountTypePercentage DiscountType = 0

	// DiscountTypeFixed is a fixed amount discount
	DiscountTypeFixed DiscountType = 1

	// DiscountTypeVolume is a volume-based discount
	DiscountTypeVolume DiscountType = 2

	// DiscountTypeCoupon is a coupon code discount
	DiscountTypeCoupon DiscountType = 3

	// DiscountTypeReferral is a referral discount
	DiscountTypeReferral DiscountType = 4

	// DiscountTypeLoyalty is a loyalty/rewards discount
	DiscountTypeLoyalty DiscountType = 5
)

// DiscountTypeNames maps types to names
var DiscountTypeNames = map[DiscountType]string{
	DiscountTypePercentage: "percentage",
	DiscountTypeFixed:      "fixed",
	DiscountTypeVolume:     "volume",
	DiscountTypeCoupon:     "coupon",
	DiscountTypeReferral:   "referral",
	DiscountTypeLoyalty:    "loyalty",
}

// String returns string representation
func (t DiscountType) String() string {
	if name, ok := DiscountTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// DiscountPolicy defines a discount policy
type DiscountPolicy struct {
	// PolicyID is the unique identifier
	PolicyID string `json:"policy_id"`

	// Name is the policy name
	Name string `json:"name"`

	// Description describes the discount
	Description string `json:"description"`

	// Type is the discount type
	Type DiscountType `json:"type"`

	// PercentageBps is the discount percentage in basis points (for percentage type)
	PercentageBps uint32 `json:"percentage_bps,omitempty"`

	// FixedAmount is the fixed discount amount (for fixed type)
	FixedAmount sdk.Coin `json:"fixed_amount,omitempty"`

	// VolumeThresholds are thresholds for volume discounts
	VolumeThresholds []VolumeThreshold `json:"volume_thresholds,omitempty"`

	// ApplicableUsageTypes limits discount to specific usage types
	ApplicableUsageTypes []UsageType `json:"applicable_usage_types,omitempty"`

	// MinOrderAmount is minimum order amount for discount to apply
	MinOrderAmount sdk.Coin `json:"min_order_amount,omitempty"`

	// MaxDiscountAmount is maximum discount amount
	MaxDiscountAmount sdk.Coin `json:"max_discount_amount,omitempty"`

	// EffectiveFrom is when the policy starts
	EffectiveFrom time.Time `json:"effective_from"`

	// EffectiveUntil is when the policy expires
	EffectiveUntil time.Time `json:"effective_until"`

	// MaxUsageCount is maximum times the discount can be used (0 = unlimited)
	MaxUsageCount uint32 `json:"max_usage_count"`

	// CurrentUsageCount is current usage count
	CurrentUsageCount uint32 `json:"current_usage_count"`

	// StackableWith lists policies this can be combined with
	StackableWith []string `json:"stackable_with,omitempty"`

	// Provider limits the discount to a specific provider (empty = all)
	Provider string `json:"provider,omitempty"`

	// IsActive indicates if the policy is active
	IsActive bool `json:"is_active"`
}

// VolumeThreshold defines a volume-based discount threshold
type VolumeThreshold struct {
	// ThresholdID is the threshold identifier
	ThresholdID string `json:"threshold_id"`

	// MinVolume is the minimum volume for this threshold
	MinVolume sdkmath.LegacyDec `json:"min_volume"`

	// MaxVolume is the maximum volume for this threshold (0 = unlimited)
	MaxVolume sdkmath.LegacyDec `json:"max_volume"`

	// DiscountBps is the discount percentage in basis points
	DiscountBps uint32 `json:"discount_bps"`
}

// Validate validates the discount policy
func (d *DiscountPolicy) Validate() error {
	if d.PolicyID == "" {
		return fmt.Errorf("policy_id is required")
	}

	if d.Name == "" {
		return fmt.Errorf("name is required")
	}

	switch d.Type {
	case DiscountTypePercentage:
		if d.PercentageBps == 0 || d.PercentageBps > 10000 {
			return fmt.Errorf("percentage_bps must be between 1 and 10000")
		}
	case DiscountTypeFixed:
		if !d.FixedAmount.IsPositive() {
			return fmt.Errorf("fixed_amount must be positive")
		}
	case DiscountTypeVolume:
		if len(d.VolumeThresholds) == 0 {
			return fmt.Errorf("volume discounts require at least one threshold")
		}
	}

	if !d.EffectiveUntil.IsZero() && d.EffectiveUntil.Before(d.EffectiveFrom) {
		return fmt.Errorf("effective_until must be after effective_from")
	}

	return nil
}

// IsValidAt checks if the policy is valid at the given time
func (d *DiscountPolicy) IsValidAt(t time.Time) bool {
	if !d.IsActive {
		return false
	}

	if t.Before(d.EffectiveFrom) {
		return false
	}

	if !d.EffectiveUntil.IsZero() && t.After(d.EffectiveUntil) {
		return false
	}

	if d.MaxUsageCount > 0 && d.CurrentUsageCount >= d.MaxUsageCount {
		return false
	}

	return true
}

// CalculateDiscount calculates the discount amount for a given subtotal
func (d *DiscountPolicy) CalculateDiscount(subtotal sdk.Coins, volume sdkmath.LegacyDec) sdk.Coins {
	if !d.IsActive {
		return sdk.NewCoins()
	}

	// Check minimum order amount
	if d.MinOrderAmount.IsPositive() && !subtotal.IsAllGTE(sdk.NewCoins(d.MinOrderAmount)) {
		return sdk.NewCoins()
	}

	var discountAmount sdk.Coins

	switch d.Type {
	case DiscountTypePercentage:
		// Calculate percentage discount
		for _, coin := range subtotal {
			discountAmt := coin.Amount.Mul(sdkmath.NewInt(int64(d.PercentageBps))).Quo(sdkmath.NewInt(10000))
			discountAmount = discountAmount.Add(sdk.NewCoin(coin.Denom, discountAmt))
		}

	case DiscountTypeFixed:
		discountAmount = sdk.NewCoins(d.FixedAmount)

	case DiscountTypeVolume:
		// Find applicable threshold
		var applicableBps uint32
		for _, threshold := range d.VolumeThresholds {
			if volume.GTE(threshold.MinVolume) {
				if threshold.MaxVolume.IsZero() || volume.LT(threshold.MaxVolume) {
					applicableBps = threshold.DiscountBps
				}
			}
		}

		if applicableBps > 0 {
			for _, coin := range subtotal {
				discountAmt := coin.Amount.Mul(sdkmath.NewInt(int64(applicableBps))).Quo(sdkmath.NewInt(10000))
				discountAmount = discountAmount.Add(sdk.NewCoin(coin.Denom, discountAmt))
			}
		}

	default:
		return sdk.NewCoins()
	}

	// Apply maximum discount cap
	if d.MaxDiscountAmount.IsPositive() {
		capped := sdk.NewCoins()
		for _, coin := range discountAmount {
			if coin.Denom == d.MaxDiscountAmount.Denom && coin.Amount.GT(d.MaxDiscountAmount.Amount) {
				capped = capped.Add(d.MaxDiscountAmount)
			} else {
				capped = capped.Add(coin)
			}
		}
		discountAmount = capped
	}

	// Ensure discount doesn't exceed subtotal
	if discountAmount.IsAllGT(subtotal) {
		return subtotal
	}

	return discountAmount
}

// CouponCode represents a redeemable coupon code
type CouponCode struct {
	// Code is the coupon code
	Code string `json:"code"`

	// DiscountPolicyID links to the discount policy
	DiscountPolicyID string `json:"discount_policy_id"`

	// MaxRedemptions is maximum redemption count (0 = unlimited)
	MaxRedemptions uint32 `json:"max_redemptions"`

	// CurrentRedemptions is current redemption count
	CurrentRedemptions uint32 `json:"current_redemptions"`

	// PerCustomerLimit is redemptions per customer (0 = unlimited)
	PerCustomerLimit uint32 `json:"per_customer_limit"`

	// CustomerRedemptions tracks per-customer usage
	CustomerRedemptions map[string]uint32 `json:"customer_redemptions"`

	// ValidFrom is when the coupon becomes valid
	ValidFrom time.Time `json:"valid_from"`

	// ValidUntil is when the coupon expires
	ValidUntil time.Time `json:"valid_until"`

	// IsActive indicates if the coupon is active
	IsActive bool `json:"is_active"`

	// CreatedBy is the creator address
	CreatedBy string `json:"created_by"`

	// CreatedAt is when the coupon was created
	CreatedAt time.Time `json:"created_at"`
}

// Validate validates the coupon code
func (c *CouponCode) Validate() error {
	if c.Code == "" {
		return fmt.Errorf("code is required")
	}

	if len(c.Code) < 4 || len(c.Code) > 32 {
		return fmt.Errorf("code must be between 4 and 32 characters")
	}

	if c.DiscountPolicyID == "" {
		return fmt.Errorf("discount_policy_id is required")
	}

	if !c.ValidUntil.IsZero() && c.ValidUntil.Before(c.ValidFrom) {
		return fmt.Errorf("valid_until must be after valid_from")
	}

	return nil
}

// CanRedeem checks if the coupon can be redeemed by a customer
func (c *CouponCode) CanRedeem(customer string, t time.Time) error {
	if !c.IsActive {
		return fmt.Errorf("coupon is not active")
	}

	if t.Before(c.ValidFrom) {
		return fmt.Errorf("coupon is not yet valid")
	}

	if !c.ValidUntil.IsZero() && t.After(c.ValidUntil) {
		return fmt.Errorf("coupon has expired")
	}

	if c.MaxRedemptions > 0 && c.CurrentRedemptions >= c.MaxRedemptions {
		return fmt.Errorf("coupon has reached maximum redemptions")
	}

	if c.PerCustomerLimit > 0 {
		if count, ok := c.CustomerRedemptions[customer]; ok && count >= c.PerCustomerLimit {
			return fmt.Errorf("customer has reached redemption limit for this coupon")
		}
	}

	return nil
}

// RecordRedemption records a coupon redemption
func (c *CouponCode) RecordRedemption(customer string) {
	c.CurrentRedemptions++

	if c.CustomerRedemptions == nil {
		c.CustomerRedemptions = make(map[string]uint32)
	}
	c.CustomerRedemptions[customer]++
}

// AppliedDiscount records a discount applied to an invoice
type AppliedDiscount struct {
	// DiscountID is the unique identifier for this application
	DiscountID string `json:"discount_id"`

	// PolicyID is the discount policy applied
	PolicyID string `json:"policy_id"`

	// CouponCode is the coupon code used (if any)
	CouponCode string `json:"coupon_code,omitempty"`

	// Type is the discount type
	Type DiscountType `json:"type"`

	// Description describes the discount
	Description string `json:"description"`

	// Amount is the discount amount
	Amount sdk.Coins `json:"amount"`

	// AppliedAt is when the discount was applied
	AppliedAt time.Time `json:"applied_at"`

	// AppliedBy is the address that applied the discount
	AppliedBy string `json:"applied_by"`
}

// Validate validates the applied discount
func (a *AppliedDiscount) Validate() error {
	if a.DiscountID == "" {
		return fmt.Errorf("discount_id is required")
	}

	if a.PolicyID == "" {
		return fmt.Errorf("policy_id is required")
	}

	if !a.Amount.IsValid() {
		return fmt.Errorf("amount must be valid")
	}

	return nil
}

// LoyaltyProgram defines a loyalty/rewards program
type LoyaltyProgram struct {
	// ProgramID is the unique identifier
	ProgramID string `json:"program_id"`

	// Name is the program name
	Name string `json:"name"`

	// Description describes the program
	Description string `json:"description"`

	// PointsPerUnit is points earned per currency unit spent
	PointsPerUnit sdkmath.LegacyDec `json:"points_per_unit"`

	// RedemptionRate is currency value per point redeemed
	RedemptionRate sdkmath.LegacyDec `json:"redemption_rate"`

	// MinRedemptionPoints is minimum points for redemption
	MinRedemptionPoints uint64 `json:"min_redemption_points"`

	// MaxRedemptionPercentBps is max percentage of invoice payable with points
	MaxRedemptionPercentBps uint32 `json:"max_redemption_percent_bps"`

	// Tiers define loyalty tiers
	Tiers []LoyaltyTier `json:"tiers"`

	// IsActive indicates if the program is active
	IsActive bool `json:"is_active"`
}

// LoyaltyTier defines a loyalty tier
type LoyaltyTier struct {
	// TierID is the tier identifier
	TierID string `json:"tier_id"`

	// TierName is the tier name
	TierName string `json:"tier_name"`

	// MinPoints is minimum points for this tier
	MinPoints uint64 `json:"min_points"`

	// BonusMultiplierBps is bonus points multiplier in basis points (10000 = 1x)
	BonusMultiplierBps uint32 `json:"bonus_multiplier_bps"`

	// ExtraDiscountBps is additional discount in basis points
	ExtraDiscountBps uint32 `json:"extra_discount_bps"`
}

// CustomerLoyalty tracks a customer's loyalty status
type CustomerLoyalty struct {
	// Customer is the customer address
	Customer string `json:"customer"`

	// ProgramID is the loyalty program
	ProgramID string `json:"program_id"`

	// TotalPointsEarned is all-time points earned
	TotalPointsEarned uint64 `json:"total_points_earned"`

	// AvailablePoints is current redeemable points
	AvailablePoints uint64 `json:"available_points"`

	// RedeemedPoints is points already redeemed
	RedeemedPoints uint64 `json:"redeemed_points"`

	// CurrentTierID is the current loyalty tier
	CurrentTierID string `json:"current_tier_id"`

	// TierAchievedAt is when current tier was reached
	TierAchievedAt time.Time `json:"tier_achieved_at"`

	// TotalSpent is total amount spent
	TotalSpent sdk.Coins `json:"total_spent"`

	// JoinedAt is when customer joined the program
	JoinedAt time.Time `json:"joined_at"`

	// LastActivityAt is last activity timestamp
	LastActivityAt time.Time `json:"last_activity_at"`
}

// EarnPoints adds points to the customer's balance
func (cl *CustomerLoyalty) EarnPoints(points uint64, now time.Time) {
	cl.TotalPointsEarned += points
	cl.AvailablePoints += points
	cl.LastActivityAt = now
}

// RedeemPoints redeems points from the customer's balance
func (cl *CustomerLoyalty) RedeemPoints(points uint64, now time.Time) error {
	if points > cl.AvailablePoints {
		return fmt.Errorf("insufficient points: have %d, need %d", cl.AvailablePoints, points)
	}

	cl.AvailablePoints -= points
	cl.RedeemedPoints += points
	cl.LastActivityAt = now
	return nil
}

// UpdateTier updates the customer's tier based on points
func (cl *CustomerLoyalty) UpdateTier(program *LoyaltyProgram, now time.Time) {
	var newTierID string
	for _, tier := range program.Tiers {
		if cl.TotalPointsEarned >= tier.MinPoints {
			newTierID = tier.TierID
		}
	}

	if newTierID != cl.CurrentTierID {
		cl.CurrentTierID = newTierID
		cl.TierAchievedAt = now
	}
}
