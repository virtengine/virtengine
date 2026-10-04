package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestHPCBillingCalculator_AppliesQueueTimePenaltyCredit(t *testing.T) {
	denom := "uvirt"
	zeroRate := sdk.NewDecCoinFromDec(denom, sdkmath.LegacyZeroDec())

	rules := DefaultHPCBillingRules(denom)
	rules.ResourceRates = HPCResourceRates{
		CPUCoreHourRate:   sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(100)),
		GPUHourRate:       zeroRate,
		MemoryGBHourRate:  zeroRate,
		NodeHourRate:      zeroRate,
		StorageGBHourRate: zeroRate,
		NetworkGBRate:     zeroRate,
	}
	rules.MinimumCharge = sdk.NewCoin(denom, sdkmath.ZeroInt())
	rules.DiscountRules = nil
	rules.BillingCaps = nil
	rules.QueueTimePenaltyEnabled = true
	rules.QueueTimePenaltyThresholdSeconds = 3600
	rules.QueueTimePenaltyRateBps = 10

	calculator := NewHPCBillingCalculator(rules)
	breakdown, billable, err := calculator.CalculateBillableAmount(&HPCDetailedMetrics{
		CPUCoreSeconds:   3600,
		QueueTimeSeconds: 7200,
		NodeHours:        sdkmath.LegacyZeroDec(),
	}, nil, nil)
	require.NoError(t, err)

	require.Equal(t, int64(100), breakdown.CPUCost.Amount.Int64())
	require.Equal(t, int64(100), breakdown.Subtotal.AmountOf(denom).Int64())
	require.Equal(t, int64(6), breakdown.QueuePenalty.Amount.Int64())
	require.Equal(t, int64(94), billable.AmountOf(denom).Int64())
}

// regressionRules returns rules whose rates are large enough that every cost
// component truncates to a non-zero integer amount, which is what makes the
// duplicate-denomination defect in CalculateBillableAmount reachable.
func regressionRules(denom string) HPCBillingRules {
	rules := DefaultHPCBillingRules(denom)
	rules.ResourceRates = HPCResourceRates{
		CPUCoreHourRate:   sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(10000)),
		MemoryGBHourRate:  sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(1000)),
		GPUHourRate:       sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(2000)),
		NodeHourRate:      sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(50000)),
		StorageGBHourRate: sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(50)),
		NetworkGBRate:     sdk.NewDecCoinFromDec(denom, sdkmath.LegacyNewDec(100)),
		GPUTypeRates:      map[string]sdk.DecCoin{},
	}
	rules.DiscountRules = nil
	rules.BillingCaps = nil
	rules.MinimumCharge = sdk.NewCoin(denom, sdkmath.ZeroInt())

	return rules
}

// TestCalculateBillableAmountSubtotalDoesNotPanicOnSharedDenom pins the
// regression where CalculateBillableAmount passed six same-denom coins to
// sdk.NewCoins, which panics with "duplicate denomination" as soon as two or
// more cost components are non-zero.
func TestCalculateBillableAmountSubtotalDoesNotPanicOnSharedDenom(t *testing.T) {
	const denom = "uvirt"

	calculator := NewHPCBillingCalculator(regressionRules(denom))

	// 4 core-hours * 10000 = 40000
	// 16 GB-hours * 1000 = 16000
	// 2 GPU-hours * 2000 = 4000
	// 1 node-hour * 50000 = 50000
	// 10 GB-hours * 50 = 500
	// 1.5 GB * 100 = 150 (truncated)
	metrics := &HPCDetailedMetrics{
		WallClockSeconds: 3600,
		CPUCoreSeconds:   3600 * 4,
		MemoryGBSeconds:  3600 * 16,
		GPUSeconds:       3600 * 2,
		StorageGBHours:   10,
		NetworkBytesIn:   1024 * 1024 * 1024,
		NetworkBytesOut:  512 * 1024 * 1024,
		NodeHours:        sdkmath.LegacyNewDec(1),
		NodesUsed:        1,
	}

	var (
		breakdown *BillableBreakdown
		billable  sdk.Coins
		err       error
	)
	require.NotPanics(t, func() {
		breakdown, billable, err = calculator.CalculateBillableAmount(metrics, nil, nil)
	}, "CalculateBillableAmount must not panic on multiple non-zero components")
	require.NoError(t, err)
	require.NotNil(t, breakdown)

	components := breakdown.CPUCost.Amount.
		Add(breakdown.MemoryCost.Amount).
		Add(breakdown.GPUCost.Amount).
		Add(breakdown.NodeCost.Amount).
		Add(breakdown.StorageCost.Amount).
		Add(breakdown.NetworkCost.Amount)

	// Per-component breakdown must remain available to callers
	// (keeper/billing.go and keeper/settlement.go read these fields).
	require.Equal(t, int64(40000), breakdown.CPUCost.Amount.Int64())
	require.Equal(t, int64(16000), breakdown.MemoryCost.Amount.Int64())
	require.Equal(t, int64(4000), breakdown.GPUCost.Amount.Int64())
	require.Equal(t, int64(50000), breakdown.NodeCost.Amount.Int64())
	require.Equal(t, int64(500), breakdown.StorageCost.Amount.Int64())
	require.Equal(t, int64(150), breakdown.NetworkCost.Amount.Int64())

	// The subtotal is the sum of the components, under a single denom.
	require.Equal(t, sdk.NewCoins(sdk.NewCoin(denom, components)), breakdown.Subtotal)
	require.Equal(t, int64(110650), components.Int64())
	require.Equal(t, components, breakdown.Subtotal.AmountOf(denom))

	// With no discounts, caps or queue penalty the final amount is the subtotal.
	require.Equal(t, breakdown.Subtotal, billable)
}

// TestCalculateBillableAmountSubtotalSingleComponentUnchanged proves the fix
// does not change behaviour when only one component is non-zero: the subtotal
// is still exactly that component's amount.
func TestCalculateBillableAmountSubtotalSingleComponentUnchanged(t *testing.T) {
	const denom = "uvirt"

	calculator := NewHPCBillingCalculator(regressionRules(denom))

	// Only CPU is non-zero: 4 core-hours * 10000 = 40000.
	metrics := &HPCDetailedMetrics{
		WallClockSeconds: 3600,
		CPUCoreSeconds:   3600 * 4,
		NodeHours:        sdkmath.LegacyZeroDec(),
	}

	var breakdown *BillableBreakdown
	require.NotPanics(t, func() {
		var err error
		breakdown, _, err = calculator.CalculateBillableAmount(metrics, nil, nil)
		require.NoError(t, err)
	})

	require.Equal(t, int64(40000), breakdown.CPUCost.Amount.Int64())
	require.True(t, breakdown.MemoryCost.IsZero())
	require.True(t, breakdown.GPUCost.IsZero())
	require.True(t, breakdown.NodeCost.IsZero())
	require.True(t, breakdown.StorageCost.IsZero())
	require.True(t, breakdown.NetworkCost.IsZero())

	require.Equal(t, breakdown.CPUCost.Amount, breakdown.Subtotal.AmountOf(denom))
	require.Equal(t, sdk.NewCoins(breakdown.CPUCost), breakdown.Subtotal)
}

// TestCalculateBillableAmountSubtotalAllZeroStaysEmpty covers the boundary case
// where no component is non-zero: the subtotal must remain an empty Coin set
// rather than a zero-valued coin, since sdk.Coins drops zero coins.
func TestCalculateBillableAmountSubtotalAllZeroStaysEmpty(t *testing.T) {
	const denom = "uvirt"

	zeroRate := sdk.NewDecCoinFromDec(denom, sdkmath.LegacyZeroDec())
	rules := DefaultHPCBillingRules(denom)
	rules.ResourceRates = HPCResourceRates{
		CPUCoreHourRate:   zeroRate,
		MemoryGBHourRate:  zeroRate,
		GPUHourRate:       zeroRate,
		NodeHourRate:      zeroRate,
		StorageGBHourRate: zeroRate,
		NetworkGBRate:     zeroRate,
	}
	rules.DiscountRules = nil
	rules.BillingCaps = nil
	rules.MinimumCharge = sdk.NewCoin(denom, sdkmath.ZeroInt())

	calculator := NewHPCBillingCalculator(rules)

	var (
		breakdown *BillableBreakdown
		billable  sdk.Coins
	)
	require.NotPanics(t, func() {
		var err error
		breakdown, billable, err = calculator.CalculateBillableAmount(
			&HPCDetailedMetrics{NodeHours: sdkmath.LegacyZeroDec()}, nil, nil)
		require.NoError(t, err)
	})

	// Minimum charge is zero, so nothing is owed.
	require.Equal(t, sdk.NewCoins(), billable)
	require.True(t, breakdown.Subtotal.IsZero())
	require.Empty(t, breakdown.Subtotal)
}
