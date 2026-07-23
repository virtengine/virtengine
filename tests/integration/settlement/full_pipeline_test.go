//go:build e2e.integration

// Package settlement_test contains integration tests for the settlement pipeline.
//
// VE-68B: Settlement pipeline integration - full usage→invoice→settlement→payout flow
// These tests validate the complete financial settlement pipeline including:
// - Usage record processing
// - Invoice generation
// - Settlement calculation with platform fees
// - Dispute handling and resolution
// - Escrow balance reconciliation
package settlement_test

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/testutil/state"
	escrowkeeper "github.com/virtengine/virtengine/x/escrow/keeper"
	"github.com/virtengine/virtengine/x/escrow/types/billing"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

type FullPipelineTestSuite struct {
	suite.Suite
	testSuite     *state.TestSuite
	ctx           sdk.Context
	invoiceKeeper escrowkeeper.InvoiceKeeper
	provider      sdk.AccAddress
	customer      sdk.AccAddress
	currency      string
	usage         *authenticatedUsageFixture
}

func TestFullPipelineTestSuite(t *testing.T) {
	suite.Run(t, new(FullPipelineTestSuite))
}

func (suite *FullPipelineTestSuite) SetupTest() {
	suite.testSuite = state.SetupTestSuiteWithoutModuleServices(suite.T())
	suite.ctx = suite.testSuite.Context()
	suite.invoiceKeeper = requireInvoiceKeeper(suite.T(), suite.testSuite.App().Keepers.VirtEngine.Escrow)
	suite.provider = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	suite.customer = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	suite.currency = "uve"
	suite.usage = newAuthenticatedUsageFixture(suite.T(), suite.testSuite, suite.provider)

	fundAccount(suite.T(), suite.testSuite, suite.customer, sdk.NewCoins(sdk.NewCoin(suite.currency, sdkmath.NewInt(500000))))
}

func (suite *FullPipelineTestSuite) TestUsageToInvoiceToPayout() {
	t := suite.T()
	ctx := suite.ctx

	orderID := "order-pipeline-001"
	leaseID := "lease-pipeline-001"
	initialEscrow := int64(100000)
	escrowID := suite.createAndActivateEscrow(ctx, orderID, leaseID, initialEscrow, 72*time.Hour)
	providerBalanceBefore := suite.testSuite.App().Keepers.Cosmos.Bank.GetBalance(ctx, suite.provider, suite.currency)

	usageIDs := []string{
		suite.recordUsage(ctx, orderID, leaseID, "cpu", 384, 38400, ctx.BlockTime().Add(-24*time.Hour), ctx.BlockTime()),
		suite.recordUsage(ctx, orderID, leaseID, "memory", 1536, 9600, ctx.BlockTime().Add(-24*time.Hour), ctx.BlockTime()),
		suite.recordUsage(ctx, orderID, leaseID, "storage", 24, 2000, ctx.BlockTime().Add(-24*time.Hour), ctx.BlockTime()),
	}

	keeper := &suite.testSuite.App().Keepers.VirtEngine.Settlement
	settlement, err := keeper.SettleOrder(ctx, orderID, usageIDs, false)
	require.NoError(t, err)
	require.NotNil(t, settlement)

	payout, found := keeper.GetPayoutBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)
	require.Equal(t, settlementtypes.PayoutStateCompleted, payout.State)

	invoice, err := suite.invoiceKeeper.GetInvoice(ctx, payout.InvoiceID)
	require.NoError(t, err)
	require.Equal(t, billing.InvoiceStatusPaid, invoice.Status)
	require.Equal(t, orderID, invoice.OrderID)
	require.Equal(t, leaseID, invoice.LeaseID)
	require.True(t, invoice.Total.Equal(settlement.TotalAmount))
	require.Equal(t, settlement.SettlementID, invoice.SettlementID)

	ledgerEntries, err := suite.invoiceKeeper.GetInvoiceLedgerEntries(ctx, invoice.InvoiceID)
	require.NoError(t, err)
	require.NotEmpty(t, ledgerEntries)

	ledgerChain, err := suite.invoiceKeeper.GetInvoiceLedgerChain(ctx, invoice.InvoiceID)
	require.NoError(t, err)
	require.NoError(t, ledgerChain.Validate())

	providerBalanceAfter := suite.testSuite.App().Keepers.Cosmos.Bank.GetBalance(ctx, suite.provider, suite.currency)
	require.Equal(t, providerBalanceBefore.Amount.Add(payout.NetAmount.AmountOf(suite.currency)), providerBalanceAfter.Amount)

	treasury := keeper.GetTreasuryBalance(ctx)
	require.True(t, treasury.Equal(payout.PlatformFee.Add(payout.ValidatorFee...)))

	escrow, found := keeper.GetEscrow(ctx, escrowID)
	require.True(t, found)
	require.Equal(t, sdkmath.NewInt(initialEscrow).Sub(settlement.TotalAmount.AmountOf(suite.currency)), escrow.Balance.AmountOf(suite.currency))

	suite.verifySettlementInvariant(t, settlement)
}

func (suite *FullPipelineTestSuite) TestDisputeResolutionAdjustedPayout() {
	t := suite.T()
	ctx := suite.ctx

	keeper := &suite.testSuite.App().Keepers.VirtEngine.Settlement
	configureFiatTestParams(t, suite.testSuite)
	keeper.SetComplianceKeeper(suite.testSuite.App().Keepers.VirtEngine.VEID)

	params := keeper.GetParams(ctx)
	params.PayoutHoldbackRate = "0.20"
	require.NoError(t, keeper.SetParams(ctx, params))

	seedComplianceRecord(t, suite.testSuite, suite.provider)
	require.NoError(t, keeper.SetFiatPayoutPreference(
		ctx,
		makeFiatPayoutPreference(t, suite.testSuite, ctx.BlockTime(), suite.provider, suite.currency, "acct-adjustment"),
	))

	orderID := "order-dispute-adjusted"
	leaseID := "lease-dispute-adjusted"
	escrowID := suite.createAndActivateEscrow(ctx, orderID, leaseID, 1000, 72*time.Hour)
	customerBalanceBeforeResolution := suite.testSuite.App().Keepers.Cosmos.Bank.GetBalance(ctx, suite.customer, suite.currency)
	usageID := suite.recordUsage(ctx, orderID, leaseID, "compute", 1, 1000, ctx.BlockTime().Add(-time.Hour), ctx.BlockTime())

	settlement, err := keeper.SettleOrder(ctx, orderID, []string{usageID}, false)
	require.NoError(t, err)

	payout, found := keeper.GetPayoutBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)
	require.Equal(t, settlementtypes.PayoutStatePending, payout.State)
	require.False(t, payout.HoldbackAmount.IsZero())

	openErr := keeper.OnDisputeOpened(ctx, payout.InvoiceID, "dispute-adjusted-1", "service degradation")
	held, found := keeper.GetPayout(ctx, payout.PayoutID)
	require.True(t, found)
	if openErr == nil {
		require.Equal(t, settlementtypes.PayoutStateHeld, held.State)
		require.ErrorIs(t, keeper.OnDisputeResolved(ctx, payout.InvoiceID, billing.DisputeResolutionPartialRefund), settlementtypes.ErrLegacyFinancialMutationFenced)
	} else {
		require.ErrorIs(t, openErr, settlementtypes.ErrLegacyFinancialMutationFenced)
		require.Equal(t, settlementtypes.PayoutStatePending, held.State)
	}

	conversion, found := keeper.GetFiatConversionBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)
	require.Contains(t, []settlementtypes.FiatConversionState{settlementtypes.FiatConversionStateCreated, settlementtypes.FiatConversionStateCancelled}, conversion.State)

	customerBalanceAfterResolution := suite.testSuite.App().Keepers.Cosmos.Bank.GetBalance(ctx, suite.customer, suite.currency)
	require.Equal(t, customerBalanceBeforeResolution.Amount, customerBalanceAfterResolution.Amount)
	require.True(t, keeper.GetTreasuryBalance(ctx).IsZero())

	ledger := keeper.GetPayoutLedgerEntries(ctx, payout.PayoutID)
	require.NotEmpty(t, ledger)

	escrow, found := keeper.GetEscrow(ctx, escrowID)
	require.True(t, found)
	require.Equal(t, sdkmath.ZeroInt(), escrow.Balance.AmountOf(suite.currency))
}

func (suite *FullPipelineTestSuite) TestPartialRefund() {
	t := suite.T()
	ctx := suite.ctx

	initialBalance := int64(100000)
	orderID := "order-refund-001"
	leaseID := "lease-refund-001"
	escrowID := suite.createAndActivateEscrow(ctx, orderID, leaseID, initialBalance, 72*time.Hour)
	customerBalanceBefore := suite.testSuite.App().Keepers.Cosmos.Bank.GetBalance(ctx, suite.customer, suite.currency)
	usageID := suite.recordUsage(ctx, orderID, leaseID, "compute", 12, 24000, ctx.BlockTime().Add(-12*time.Hour), ctx.BlockTime())

	keeper := &suite.testSuite.App().Keepers.VirtEngine.Settlement
	_, err := keeper.SettleOrder(ctx, orderID, []string{usageID}, false)
	require.NoError(t, err)

	escrowBeforeRefund, found := keeper.GetEscrow(ctx, escrowID)
	require.True(t, found)
	refundAmount := escrowBeforeRefund.Balance.AmountOf(suite.currency)
	require.True(t, refundAmount.IsPositive())

	require.NoError(t, keeper.RefundEscrow(ctx, escrowID, "lease terminated early"))

	refundedEscrow, found := keeper.GetEscrow(ctx, escrowID)
	require.True(t, found)
	require.Equal(t, settlementtypes.EscrowStateRefunded, refundedEscrow.State)
	require.Equal(t, sdkmath.ZeroInt(), refundedEscrow.Balance.AmountOf(suite.currency))

	customerBalanceAfter := suite.testSuite.App().Keepers.Cosmos.Bank.GetBalance(ctx, suite.customer, suite.currency)
	require.Equal(t, customerBalanceBefore.Amount.Add(refundAmount), customerBalanceAfter.Amount)

	settlements := keeper.GetSettlementsByOrder(ctx, orderID)
	require.Len(t, settlements, 2)
	require.Equal(t, refundAmount.Int64(), settlements[1].TotalAmount.AmountOf(suite.currency).Int64())
}

func (suite *FullPipelineTestSuite) TestMultiOrderSettlementBatch() {
	t := suite.T()
	ctx := suite.ctx

	const numOrders = 4
	providerBalanceBefore := suite.testSuite.App().Keepers.Cosmos.Bank.GetBalance(ctx, suite.provider, suite.currency)
	totalGross := sdkmath.ZeroInt()
	totalNet := sdkmath.ZeroInt()
	totalFees := sdkmath.ZeroInt()

	keeper := &suite.testSuite.App().Keepers.VirtEngine.Settlement
	for i := 0; i < numOrders; i++ {
		orderID := fmt.Sprintf("order-batch-%03d", i+1)
		leaseID := fmt.Sprintf("lease-batch-%03d", i+1)
		usageCost := int64(7000 + (i * 2500))

		suite.createAndActivateEscrow(ctx, orderID, leaseID, 50000, 72*time.Hour)
		usageID := suite.recordUsage(ctx, orderID, leaseID, "compute", uint64(6+i*2), usageCost, ctx.BlockTime().Add(-time.Duration(6+i*2)*time.Hour), ctx.BlockTime())

		settlement, err := keeper.SettleOrder(ctx, orderID, []string{usageID}, false)
		require.NoError(t, err)

		payout, found := keeper.GetPayoutBySettlement(ctx, settlement.SettlementID)
		require.True(t, found)
		require.Equal(t, settlementtypes.PayoutStateCompleted, payout.State)

		totalGross = totalGross.Add(settlement.TotalAmount.AmountOf(suite.currency))
		totalNet = totalNet.Add(payout.NetAmount.AmountOf(suite.currency))
		totalFees = totalFees.Add(payout.PlatformFee.AmountOf(suite.currency)).Add(payout.ValidatorFee.AmountOf(suite.currency))
	}

	providerBalanceAfter := suite.testSuite.App().Keepers.Cosmos.Bank.GetBalance(ctx, suite.provider, suite.currency)
	require.Equal(t, providerBalanceBefore.Amount.Add(totalNet), providerBalanceAfter.Amount)

	treasury := keeper.GetTreasuryBalance(ctx)
	require.Equal(t, totalFees, treasury.AmountOf(suite.currency))
	require.True(t, totalGross.Equal(totalNet.Add(totalFees)))
}

func (suite *FullPipelineTestSuite) TestEscrowExpiryAutoSettlement() {
	t := suite.T()
	ctx := suite.ctx

	keeper := &suite.testSuite.App().Keepers.VirtEngine.Settlement
	params := keeper.GetParams(ctx)
	params.SettlementPeriod = 3600
	require.NoError(t, keeper.SetParams(ctx, params))

	orderID := "order-expiry-001"
	leaseID := "lease-expiry-001"
	escrowID := suite.createAndActivateEscrow(ctx, orderID, leaseID, 50000, 4*time.Hour)
	usageID := suite.recordUsage(ctx, orderID, leaseID, "compute", 12, 12000, ctx.BlockTime().Add(-2*time.Hour), ctx.BlockTime())

	beforeExpiry := ctx.WithBlockTime(ctx.BlockTime().Add(2 * time.Hour))
	require.NoError(t, keeper.AutoSettle(beforeExpiry))

	settlements := keeper.GetSettlementsByOrder(beforeExpiry, orderID)
	require.Len(t, settlements, 1)
	require.Contains(t, settlements[0].UsageRecordIDs, usageID)

	afterExpiry := beforeExpiry.WithBlockTime(beforeExpiry.BlockTime().Add(3 * time.Hour))
	require.NoError(t, keeper.ProcessExpiredEscrows(afterExpiry))

	escrow, found := keeper.GetEscrow(afterExpiry, escrowID)
	require.True(t, found)
	require.Equal(t, settlementtypes.EscrowStateExpired, escrow.State)
	require.Equal(t, sdkmath.ZeroInt(), escrow.Balance.AmountOf(suite.currency))
}

func (suite *FullPipelineTestSuite) TestSettlementEscrowInvariant() {
	t := suite.T()
	ctx := suite.ctx

	initialBalance := int64(100000)
	orderID := "order-invariant-001"
	leaseID := "lease-invariant-001"
	escrowID := suite.createAndActivateEscrow(ctx, orderID, leaseID, initialBalance, 72*time.Hour)

	keeper := &suite.testSuite.App().Keepers.VirtEngine.Settlement
	firstUsage := suite.recordUsage(ctx, orderID, leaseID, "cpu", 8, 20000, ctx.BlockTime().Add(-2*time.Hour), ctx.BlockTime().Add(-time.Hour))
	firstSettlement, err := keeper.SettleOrder(ctx, orderID, []string{firstUsage}, false)
	require.NoError(t, err)

	secondUsage := suite.recordUsage(ctx, orderID, leaseID, "gpu", 4, 15000, ctx.BlockTime().Add(-time.Hour), ctx.BlockTime())
	secondSettlement, err := keeper.SettleOrder(ctx, orderID, []string{secondUsage}, false)
	require.NoError(t, err)

	escrow, found := keeper.GetEscrow(ctx, escrowID)
	require.True(t, found)
	refundAmount := escrow.Balance.AmountOf(suite.currency)
	require.True(t, refundAmount.IsPositive())

	require.NoError(t, keeper.RefundEscrow(ctx, escrowID, "invariant refund"))

	settlements := keeper.GetSettlementsByOrder(ctx, orderID)
	require.Len(t, settlements, 3)

	total := sdkmath.ZeroInt()
	for _, settlement := range settlements {
		total = total.Add(settlement.TotalAmount.AmountOf(suite.currency))
	}

	require.Equal(t, sdkmath.NewInt(initialBalance), total)
	suite.verifySettlementInvariant(t, firstSettlement)
	suite.verifySettlementInvariant(t, secondSettlement)
	require.Equal(t, refundAmount.Int64(), settlements[2].TotalAmount.AmountOf(suite.currency).Int64())
}

func (suite *FullPipelineTestSuite) createAndActivateEscrow(ctx sdk.Context, orderID, leaseID string, amount int64, expiry time.Duration) string {
	t := suite.T()

	escrowID, err := suite.testSuite.App().Keepers.VirtEngine.Settlement.CreateEscrow(
		ctx,
		orderID,
		suite.customer,
		sdk.NewCoins(sdk.NewCoin(suite.currency, sdkmath.NewInt(amount))),
		expiry,
		nil,
	)
	require.NoError(t, err)
	require.NoError(t, suite.testSuite.App().Keepers.VirtEngine.Settlement.ActivateEscrow(ctx, escrowID, leaseID, suite.provider))

	return escrowID
}

func (suite *FullPipelineTestSuite) recordUsage(
	ctx sdk.Context,
	orderID string,
	leaseID string,
	usageType string,
	usageUnits uint64,
	totalCost int64,
	periodStart time.Time,
	periodEnd time.Time,
) string {
	t := suite.T()

	record := suite.usage.newUsage(
		t, ctx, orderID, leaseID, suite.provider, suite.customer,
		usageType, usageUnits, totalCost, periodStart, periodEnd,
	)
	require.NoError(t, suite.testSuite.App().Keepers.VirtEngine.Settlement.RecordUsage(ctx, record))
	require.True(t, record.IsAuthenticated())

	return record.UsageID
}

func (suite *FullPipelineTestSuite) verifySettlementInvariant(t *testing.T, settlement *settlementtypes.SettlementRecord) {
	t.Helper()

	require.True(
		t,
		settlement.ProviderShare.Add(settlement.PlatformFee...).Add(settlement.ValidatorFee...).Equal(settlement.TotalAmount),
		"expected settlement split %s + %s + %s to equal %s",
		settlement.ProviderShare.String(),
		settlement.PlatformFee.String(),
		settlement.ValidatorFee.String(),
		settlement.TotalAmount.String(),
	)
}
