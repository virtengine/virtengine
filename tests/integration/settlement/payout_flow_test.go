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
	suite := state.SetupTestSuiteWithoutModuleServices(t)
	ctx := suite.Context()
	app := suite.App()
	keeper := &app.Keepers.VirtEngine.Settlement

	configureFiatTestParams(t, suite)
	keeper.SetComplianceKeeper(app.Keepers.VirtEngine.VEID)

	depositor := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	provider := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	seedComplianceRecord(t, suite, provider)
	fundAccount(t, suite, depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100000))))
	usageAuth := newAuthenticatedUsageFixture(t, suite, provider)

	pref := makeFiatPayoutPreference(t, suite, ctx.BlockTime(), provider, "uve", "acct-token")
	require.NoError(t, keeper.SetFiatPayoutPreference(ctx, pref))

	escrowID, err := keeper.CreateEscrow(ctx, "order-flow-1", depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))), 24*time.Hour, nil)
	require.NoError(t, err)
	require.NoError(t, keeper.ActivateEscrow(ctx, escrowID, "lease-flow-1", provider))

	usage := usageAuth.newUsage(t, ctx, "order-flow-1", "lease-flow-1", provider, depositor,
		"compute", 1, 1000, ctx.BlockTime().Add(-time.Hour), ctx.BlockTime())
	require.NoError(t, keeper.RecordUsage(ctx, usage))

	settlement, err := keeper.SettleOrder(ctx, "order-flow-1", []string{usage.UsageID}, false)
	require.NoError(t, err)

	payout, found := keeper.GetPayoutBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)
	require.Equal(t, types.PayoutStatePending, payout.State)

	conversion, found := keeper.GetFiatConversionBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateCreated, conversion.State)
	require.Equal(t, conversion.ConversionID, payout.FiatConversionID)
	require.Empty(t, payout.TxHash)

	ledger := keeper.GetPayoutLedgerEntries(ctx, payout.PayoutID)
	require.NotEmpty(t, ledger)
}

func TestDisputeArbitrationRefundFlow(t *testing.T) {
	suite := state.SetupTestSuiteWithoutModuleServices(t)
	ctx := suite.Context()
	app := suite.App()
	keeper := &app.Keepers.VirtEngine.Settlement
	bank := app.Keepers.Cosmos.Bank

	configureFiatTestParams(t, suite)
	keeper.SetComplianceKeeper(app.Keepers.VirtEngine.VEID)

	depositor := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	provider := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	seedComplianceRecord(t, suite, provider)
	fundAccount(t, suite, depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100000))))
	usageAuth := newAuthenticatedUsageFixture(t, suite, provider)
	balanceBefore := bank.GetBalance(ctx, depositor, "uve")

	pref := makeFiatPayoutPreference(t, suite, ctx.BlockTime(), provider, "uve", "acct-dispute-refund")
	require.NoError(t, keeper.SetFiatPayoutPreference(ctx, pref))

	escrowID, err := keeper.CreateEscrow(ctx, "order-dispute-1", depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))), 24*time.Hour, nil)
	require.NoError(t, err)
	require.NoError(t, keeper.ActivateEscrow(ctx, escrowID, "lease-dispute-1", provider))

	balanceAfterEscrow := bank.GetBalance(ctx, depositor, "uve")
	require.True(t, balanceAfterEscrow.Amount.LT(balanceBefore.Amount))

	usage := usageAuth.newUsage(t, ctx, "order-dispute-1", "lease-dispute-1", provider, depositor,
		"compute", 1, 1000, ctx.BlockTime().Add(-time.Hour), ctx.BlockTime())
	require.NoError(t, keeper.RecordUsage(ctx, usage))

	settlement, err := keeper.SettleOrder(ctx, "order-dispute-1", []string{usage.UsageID}, false)
	require.NoError(t, err)

	payout, found := keeper.GetPayoutBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)
	require.Equal(t, types.PayoutStatePending, payout.State)

	openErr := keeper.OnDisputeOpened(ctx, payout.InvoiceID, "dispute-1", "customer complaint")
	held, found := keeper.GetPayout(ctx, payout.PayoutID)
	require.True(t, found)
	if openErr == nil {
		require.Equal(t, types.PayoutStateHeld, held.State)
		require.ErrorIs(t, keeper.OnDisputeResolved(ctx, payout.InvoiceID, billing.DisputeResolutionCustomerWin), types.ErrLegacyFinancialMutationFenced)
	} else {
		require.ErrorIs(t, openErr, types.ErrLegacyFinancialMutationFenced)
		require.Equal(t, types.PayoutStatePending, held.State)
	}
	require.Equal(t, balanceAfterEscrow.Amount, bank.GetBalance(ctx, depositor, "uve").Amount)
	require.True(t, keeper.GetTreasuryBalance(ctx).IsZero())

	conversion, found := keeper.GetFiatConversionBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)
	require.Contains(t, []types.FiatConversionState{types.FiatConversionStateCreated, types.FiatConversionStateCancelled}, conversion.State)

	ledger := keeper.GetPayoutLedgerEntries(ctx, payout.PayoutID)
	require.NotEmpty(t, ledger)
}

func TestAutoSettleEdgeCases(t *testing.T) {
	suite := state.SetupTestSuiteWithoutModuleServices(t)
	ctx := suite.Context()
	app := suite.App()
	keeper := &app.Keepers.VirtEngine.Settlement

	params := keeper.GetParams(ctx)
	params.SettlementPeriod = 3600
	require.NoError(t, keeper.SetParams(ctx, params))

	depositor := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	provider := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	fundAccount(t, suite, depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100000))))
	usageAuth := newAuthenticatedUsageFixture(t, suite, provider)

	escrowID, err := keeper.CreateEscrow(ctx, "order-auto-1", depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(2000))), 24*time.Hour, nil)
	require.NoError(t, err)
	require.NoError(t, keeper.ActivateEscrow(ctx, escrowID, "lease-auto-1", provider))

	usage := usageAuth.newUsage(t, ctx, "order-auto-1", "lease-auto-1", provider, depositor,
		"compute", 1, 1000, ctx.BlockTime().Add(-time.Hour), ctx.BlockTime())
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

	expiryUsage := usageAuth.newUsage(t, ctx, "order-auto-expired", "lease-auto-expired", provider, depositor,
		"compute", 1, 500, ctx.BlockTime().Add(-time.Hour), ctx.BlockTime())
	require.NoError(t, keeper.RecordUsage(ctx, expiryUsage))

	expiredCtx := ctx.WithBlockTime(ctx.BlockTime().Add(3 * time.Hour))
	require.NoError(t, keeper.AutoSettle(expiredCtx))
	require.NoError(t, keeper.ProcessExpiredEscrows(expiredCtx))
	escrow, found := keeper.GetEscrow(expiredCtx, expiringEscrowID)
	require.True(t, found)
	require.Equal(t, types.EscrowStateExpired, escrow.State)
}
