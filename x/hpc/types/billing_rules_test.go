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
