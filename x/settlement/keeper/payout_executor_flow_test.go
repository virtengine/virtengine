package keeper_test

import (
	"context"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/dex"
	"github.com/virtengine/virtengine/pkg/payments/offramp"
	"github.com/virtengine/virtengine/x/settlement/keeper"
	"github.com/virtengine/virtengine/x/settlement/types"
)

type pendingOffRampBridge struct {
	quote   offramp.Quote
	result  offramp.PayoutResult
	status  offramp.PayoutResult
	getErr  error
	initErr error
}

func (p *pendingOffRampBridge) GetQuote(ctx context.Context, req offramp.QuoteRequest) (offramp.Quote, error) {
	if p.getErr != nil {
		return offramp.Quote{}, p.getErr
	}
	p.quote.Request = req
	return p.quote, nil
}

func (p *pendingOffRampBridge) InitiatePayout(ctx context.Context, quote offramp.Quote, cryptoTxRef string, destination string, metadata map[string]string) (offramp.PayoutResult, error) {
	if p.initErr != nil {
		return offramp.PayoutResult{}, p.initErr
	}
	p.result.QuoteID = quote.ID
	return p.result, nil
}

func (p *pendingOffRampBridge) GetStatus(ctx context.Context, payoutID string) (offramp.PayoutResult, error) {
	return p.status, p.getErr
}

func (p *pendingOffRampBridge) Cancel(ctx context.Context, payoutID string) error {
	return nil
}

func (s *KeeperTestSuite) TestExecutePayoutSingleSettlement() {
	t := s.T()

	params := s.keeper.GetParams(s.ctx)
	params.PayoutHoldbackRate = "0.10"
	require.NoError(t, s.keeper.SetParams(s.ctx, params))

	settlement := s.buildSettlement(t, "payout-single")
	settlement.TotalAmount = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000)))
	settlement.PlatformFee = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(50)))
	settlement.ValidatorFee = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))
	settlement.ProviderShare = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(940)))

	require.NoError(t, s.keeper.SetSettlement(s.ctx, settlement))

	payout, err := s.keeper.ExecutePayout(s.ctx, "inv-payout-1", settlement.SettlementID)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStateCompleted, payout.State)
	require.Equal(t, sdkmath.NewInt(840), payout.NetAmount.AmountOf("uve"))

	treasury := s.keeper.GetTreasuryBalance(s.ctx)
	require.Equal(t, sdkmath.NewInt(160), treasury.AmountOf("uve"))

	providerBalance := s.bankKeeper.GetBalance(s.ctx, s.provider, "uve")
	require.Equal(t, sdkmath.NewInt(840), providerBalance.Amount)

	entries := s.keeper.GetPayoutLedgerEntries(s.ctx, payout.PayoutID)
	require.NotEmpty(t, entries)
}

func (s *KeeperTestSuite) TestProcessPendingPayoutsBatch() {
	t := s.T()

	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000)))
	payout1 := types.NewPayoutRecord(
		"payout-batch-1",
		"inv-batch-1",
		"settle-batch-1",
		"escrow-batch-1",
		"order-batch-1",
		"lease-batch-1",
		s.provider.String(),
		s.depositor.String(),
		amount,
		sdk.NewCoins(),
		sdk.NewCoins(),
		sdk.NewCoins(),
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
	)
	payout2 := types.NewPayoutRecord(
		"payout-batch-2",
		"inv-batch-2",
		"settle-batch-2",
		"escrow-batch-2",
		"order-batch-2",
		"lease-batch-2",
		s.provider.String(),
		s.depositor.String(),
		amount,
		sdk.NewCoins(),
		sdk.NewCoins(),
		sdk.NewCoins(),
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
	)

	require.NoError(t, s.keeper.SetPayout(s.ctx, *payout1))
	require.NoError(t, s.keeper.SetPayout(s.ctx, *payout2))

	require.NoError(t, s.keeper.ProcessPendingPayouts(s.ctx))

	updated1, found := s.keeper.GetPayout(s.ctx, payout1.PayoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateCompleted, updated1.State)

	updated2, found := s.keeper.GetPayout(s.ctx, payout2.PayoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateCompleted, updated2.State)
}

func (s *KeeperTestSuite) TestRetryFailedPayouts() {
	t := s.T()

	params := s.keeper.GetParams(s.ctx)
	params.MaxPayoutRetries = 2
	require.NoError(t, s.keeper.SetParams(s.ctx, params))

	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000)))
	retryPayout := types.NewPayoutRecord(
		"payout-retry-1",
		"inv-retry-1",
		"settle-retry-1",
		"escrow-retry-1",
		"order-retry-1",
		"lease-retry-1",
		s.provider.String(),
		s.depositor.String(),
		amount,
		sdk.NewCoins(),
		sdk.NewCoins(),
		sdk.NewCoins(),
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
	)
	retryPayout.State = types.PayoutStateFailed
	retryPayout.ExecutionAttempts = 1

	maxedPayout := types.NewPayoutRecord(
		"payout-retry-2",
		"inv-retry-2",
		"settle-retry-2",
		"escrow-retry-2",
		"order-retry-2",
		"lease-retry-2",
		s.provider.String(),
		s.depositor.String(),
		amount,
		sdk.NewCoins(),
		sdk.NewCoins(),
		sdk.NewCoins(),
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
	)
	maxedPayout.State = types.PayoutStateFailed
	maxedPayout.ExecutionAttempts = 2

	require.NoError(t, s.keeper.SetPayout(s.ctx, *retryPayout))
	require.NoError(t, s.keeper.SetPayout(s.ctx, *maxedPayout))

	require.NoError(t, s.keeper.RetryFailedPayouts(s.ctx))

	updatedRetry, found := s.keeper.GetPayout(s.ctx, retryPayout.PayoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateCompleted, updatedRetry.State)

	updatedMaxed, found := s.keeper.GetPayout(s.ctx, maxedPayout.PayoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateFailed, updatedMaxed.State)
}

func (s *KeeperTestSuite) TestExecutePayoutIdempotentRequests() {
	t := s.T()

	settlement := s.buildSettlement(t, "payout-idempotent")
	require.NoError(t, s.keeper.SetSettlement(s.ctx, settlement))

	first, err := s.keeper.ExecutePayout(s.ctx, "inv-idem-1", settlement.SettlementID)
	require.NoError(t, err)

	second, err := s.keeper.ExecutePayout(s.ctx, "inv-idem-1", settlement.SettlementID)
	require.NoError(t, err)
	require.Equal(t, first.PayoutID, second.PayoutID)

	loaded, found := s.keeper.GetPayoutBySettlement(s.ctx, settlement.SettlementID)
	require.True(t, found)
	require.Equal(t, first.PayoutID, loaded.PayoutID)
}

func (s *KeeperTestSuite) TestReconcilePayoutAfterRestart() {
	t := s.T()

	swapQuote := dex.SwapQuote{
		ID: "quote-reconcile",
		Route: dex.SwapRoute{
			Hops: []dex.SwapHop{
				{AmountOut: sdkmath.NewInt(900)},
			},
		},
		ExpiresAt: s.ctx.BlockTime().Add(5 * time.Minute),
	}
	swapExec := &mockSwapExecutor{
		quote: swapQuote,
		result: dex.SwapResult{
			QuoteID:      swapQuote.ID,
			TxHash:       "swap-reconcile",
			InputAmount:  sdkmath.NewInt(1000),
			OutputAmount: sdkmath.NewInt(900),
			ExecutedAt:   s.ctx.BlockTime(),
		},
	}

	s.configureFiatConversion(t, swapExec)

	bridge := &pendingOffRampBridge{
		quote: offramp.Quote{
			ID:         "off-quote",
			FiatAmount: sdkmath.LegacyNewDec(100),
			CreatedAt:  s.ctx.BlockTime(),
			ExpiresAt:  s.ctx.BlockTime().Add(5 * time.Minute),
		},
		result: offramp.PayoutResult{
			ID:           "off-payout",
			Status:       offramp.StatusProcessing,
			Provider:     "mock",
			FiatAmount:   sdkmath.LegacyNewDec(100),
			CryptoAmount: sdkmath.NewInt(900),
			Reference:    "ref-1",
			InitiatedAt:  s.ctx.BlockTime(),
		},
		status: offramp.PayoutResult{
			ID:         "off-payout",
			Status:     offramp.StatusCompleted,
			Provider:   "mock",
			FiatAmount: sdkmath.LegacyNewDec(100),
			Reference:  "ref-1",
		},
	}
	s.keeper.SetOffRampBridge(bridge)

	settlement := s.buildSettlement(t, "payout-reconcile")
	require.NoError(t, s.keeper.SetSettlement(s.ctx, settlement))

	request := types.FiatConversionRequest{
		InvoiceID:         "inv-reconcile",
		SettlementID:      settlement.SettlementID,
		Provider:          settlement.Provider,
		Customer:          settlement.Customer,
		RequestedBy:       settlement.Provider,
		CryptoAmount:      sdk.NewCoin("uve", sdkmath.NewInt(1000)),
		FiatCurrency:      "USD",
		PaymentMethod:     "bank_transfer",
		Destination:       "acct-token",
		SlippageTolerance: 0.01,
		CryptoToken:       types.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken:       types.TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6},
	}

	_, err := s.keeper.RequestFiatConversion(s.ctx, request)
	require.NoError(t, err)

	payout, err := s.keeper.ExecutePayout(s.ctx, "inv-reconcile", settlement.SettlementID)
	require.NoError(t, err)

	conversion, found := s.keeper.GetFiatConversionByInvoice(s.ctx, "inv-reconcile")
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateOffRampPending, conversion.State)
	require.Equal(t, types.PayoutStateProcessing, payout.State)

	restarted := keeper.NewKeeper(s.cdc, s.keeper.StoreKey(), s.bankKeeper, "authority")
	restarted.SetDexSwapExecutor(swapExec)
	restarted.SetOffRampBridge(bridge)

	reconciled, err := restarted.ReconcileFiatConversion(s.ctx, conversion.ConversionID)
	require.NoError(t, err)
	require.Equal(t, types.FiatConversionStateCompleted, reconciled.State)

	updated, found := restarted.GetPayout(s.ctx, payout.PayoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateCompleted, updated.State)
}
