//go:build e2e.integration

package settlement_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/testutil/state"
	"github.com/virtengine/virtengine/x/escrow/types/billing"
	"github.com/virtengine/virtengine/x/settlement/types"
)

func TestSettlementPayoutOffRampFlow(t *testing.T) {
	suite := state.SetupTestSuite(t)
	ctx := suite.Context()
	app := suite.App()
	keeper := &app.Keepers.VirtEngine.Settlement

	setupConversionDeps(t, suite)

	depositor := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	provider := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	seedComplianceRecord(t, suite, provider)
	fundAccount(t, suite, depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100000))))

	pref := types.FiatPayoutPreference{
		Provider:          provider.String(),
		Enabled:           true,
		FiatCurrency:      "USD",
		PaymentMethod:     "bank_transfer",
		DestinationRef:    "acct-token",
		SlippageTolerance: 0.01,
		CryptoToken:       types.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken:       types.TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6},
		CreatedAt:         ctx.BlockTime(),
		UpdatedAt:         ctx.BlockTime(),
	}
	require.NoError(t, keeper.SetFiatPayoutPreference(ctx, pref))

	escrowID, err := keeper.CreateEscrow(ctx, "order-flow-1", depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))), 24*time.Hour, nil)
	require.NoError(t, err)
	require.NoError(t, keeper.ActivateEscrow(ctx, escrowID, "lease-flow-1", provider))

	usage := &types.UsageRecord{
		OrderID:           "order-flow-1",
		LeaseID:           "lease-flow-1",
		Provider:          provider.String(),
		Customer:          depositor.String(),
		UsageUnits:        1,
		UsageType:         "compute",
		TotalCost:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
		PeriodStart:       ctx.BlockTime().Add(-time.Hour),
		PeriodEnd:         ctx.BlockTime(),
		SubmittedAt:       ctx.BlockTime(),
		ProviderSignature: []byte("sig"),
	}
	require.NoError(t, keeper.RecordUsage(ctx, usage))

	settlement, err := keeper.SettleOrder(ctx, "order-flow-1", []string{usage.UsageID}, false)
	require.NoError(t, err)

	payout, found := keeper.GetPayoutBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateCompleted, payout.State)

	conversion, found := keeper.GetFiatConversionBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateCompleted, conversion.State)
	require.NotEmpty(t, conversion.OffRampID)

	ledger := keeper.GetPayoutLedgerEntries(ctx, payout.PayoutID)
	require.NotEmpty(t, ledger)
}

func TestDisputeArbitrationRefundFlow(t *testing.T) {
	suite := state.SetupTestSuite(t)
	ctx := suite.Context()
	app := suite.App()
	keeper := &app.Keepers.VirtEngine.Settlement
	bank := app.Keepers.Cosmos.Bank

	depositor := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	provider := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	fundAccount(t, suite, depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100000))))
	balanceBefore := bank.GetBalance(ctx, depositor, "uve")

	escrowID, err := keeper.CreateEscrow(ctx, "order-dispute-1", depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))), 24*time.Hour, nil)
	require.NoError(t, err)
	require.NoError(t, keeper.ActivateEscrow(ctx, escrowID, "lease-dispute-1", provider))

	balanceAfterEscrow := bank.GetBalance(ctx, depositor, "uve")
	require.True(t, balanceAfterEscrow.Amount.LT(balanceBefore.Amount))

	payout := types.NewPayoutRecord(
		"payout-dispute-1",
		"inv-dispute-1",
		"settle-dispute-1",
		escrowID,
		"order-dispute-1",
		"lease-dispute-1",
		provider.String(),
		depositor.String(),
		sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
		sdk.NewCoins(),
		sdk.NewCoins(),
		sdk.NewCoins(),
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)
	require.NoError(t, keeper.SetPayout(ctx, *payout))

	require.NoError(t, keeper.OnDisputeOpened(ctx, "inv-dispute-1", "dispute-1", "customer complaint"))
	held, found := keeper.GetPayout(ctx, payout.PayoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateHeld, held.State)

	require.NoError(t, keeper.OnDisputeResolved(ctx, "inv-dispute-1", billing.DisputeResolutionCustomerWin))
	refunded, found := keeper.GetPayout(ctx, payout.PayoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateRefunded, refunded.State)

	balanceAfterRefund := bank.GetBalance(ctx, depositor, "uve")
	require.Equal(t, balanceBefore.Amount, balanceAfterRefund.Amount)
}

func TestAutoSettleEdgeCases(t *testing.T) {
	suite := state.SetupTestSuite(t)
	ctx := suite.Context()
	app := suite.App()
	keeper := &app.Keepers.VirtEngine.Settlement

	params := keeper.GetParams(ctx)
	params.SettlementPeriod = 3600
	require.NoError(t, keeper.SetParams(ctx, params))

	depositor := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	provider := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	fundAccount(t, suite, depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100000))))

	escrowID, err := keeper.CreateEscrow(ctx, "order-auto-1", depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(2000))), 24*time.Hour, nil)
	require.NoError(t, err)
	require.NoError(t, keeper.ActivateEscrow(ctx, escrowID, "lease-auto-1", provider))

	usage := &types.UsageRecord{
		OrderID:           "order-auto-1",
		LeaseID:           "lease-auto-1",
		Provider:          provider.String(),
		Customer:          depositor.String(),
		UsageUnits:        1,
		UsageType:         "compute",
		TotalCost:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
		PeriodStart:       ctx.BlockTime().Add(-time.Hour),
		PeriodEnd:         ctx.BlockTime(),
		SubmittedAt:       ctx.BlockTime(),
		ProviderSignature: []byte("sig"),
	}
	require.NoError(t, keeper.RecordUsage(ctx, usage))

	earlyCtx := ctx.WithBlockTime(ctx.BlockTime().Add(30 * time.Minute))
	require.NoError(t, keeper.AutoSettle(earlyCtx))
	require.Empty(t, keeper.GetSettlementsByOrder(earlyCtx, "order-auto-1"))

	lateCtx := ctx.WithBlockTime(ctx.BlockTime().Add(2 * time.Hour))
	require.NoError(t, keeper.AutoSettle(lateCtx))
	require.NotEmpty(t, keeper.GetSettlementsByOrder(lateCtx, "order-auto-1"))

	expiringEscrowID, err := keeper.CreateEscrow(ctx, "order-auto-expired", depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500))), time.Hour, nil)
	require.NoError(t, err)
	require.NoError(t, keeper.ActivateEscrow(ctx, expiringEscrowID, "lease-auto-expired", provider))

	expiryUsage := &types.UsageRecord{
		OrderID:           "order-auto-expired",
		LeaseID:           "lease-auto-expired",
		Provider:          provider.String(),
		Customer:          depositor.String(),
		UsageUnits:        1,
		UsageType:         "compute",
		TotalCost:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500))),
		PeriodStart:       ctx.BlockTime().Add(-time.Hour),
		PeriodEnd:         ctx.BlockTime(),
		SubmittedAt:       ctx.BlockTime(),
		ProviderSignature: []byte("sig"),
	}
	require.NoError(t, keeper.RecordUsage(ctx, expiryUsage))

	expiredCtx := ctx.WithBlockTime(ctx.BlockTime().Add(3 * time.Hour))
	require.NoError(t, keeper.AutoSettle(expiredCtx))
	escrow, found := keeper.GetEscrow(expiredCtx, expiringEscrowID)
	require.True(t, found)
	require.Equal(t, types.EscrowStateExpired, escrow.State)
}
