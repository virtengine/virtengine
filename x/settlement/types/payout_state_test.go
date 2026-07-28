package types

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestPayoutRetryableFailureCanResume(t *testing.T) {
	now := time.Now().UTC()
	payout := NewPayoutRecord(
		"payout-1",
		"inv-1",
		"set-1",
		"escrow-1",
		"order-1",
		"lease-1",
		sdk.AccAddress("provider-1").String(),
		sdk.AccAddress("customer-1").String(),
		sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100))),
		sdk.NewCoins(),
		sdk.NewCoins(),
		sdk.NewCoins(),
		now,
		1,
	)

	require.NoError(t, payout.MarkProcessing(now))
	require.NoError(t, payout.MarkFailedWithRetryability("temporary outage", true, now.Add(time.Minute)))
	require.True(t, payout.CanRetry())

	retryAt := now.Add(2 * time.Minute)
	require.NoError(t, payout.MarkProcessing(retryAt))
	require.Equal(t, PayoutStateProcessing, payout.State)
	require.False(t, payout.CanRetry())
	require.False(t, payout.LastErrorRetryable)
	require.Equal(t, uint32(2), payout.ExecutionAttempts)
}

func TestPayoutTerminalFailureBlocksRetry(t *testing.T) {
	now := time.Now().UTC()
	payout := NewPayoutRecord(
		"payout-2",
		"inv-2",
		"set-2",
		"escrow-2",
		"order-2",
		"lease-2",
		sdk.AccAddress("provider-1").String(),
		sdk.AccAddress("customer-1").String(),
		sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100))),
		sdk.NewCoins(),
		sdk.NewCoins(),
		sdk.NewCoins(),
		now,
		1,
	)

	require.NoError(t, payout.MarkProcessing(now))
	require.NoError(t, payout.MarkFailedWithRetryability("compliance rejected", false, now.Add(time.Minute)))
	require.False(t, payout.CanRetry())

	err := payout.MarkProcessing(now.Add(2 * time.Minute))
	require.Error(t, err)
	require.ErrorContains(t, err, "cannot retry terminal payout")
}
