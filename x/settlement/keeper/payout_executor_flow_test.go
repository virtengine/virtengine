package keeper_test

import (
	"context"
	"errors"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/dex"
	"github.com/virtengine/virtengine/pkg/payments/offramp"
	"github.com/virtengine/virtengine/x/escrow/types/billing"
	"github.com/virtengine/virtengine/x/settlement/keeper"
	"github.com/virtengine/virtengine/x/settlement/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

type pendingOffRampBridge struct {
	quote       offramp.Quote
	result      offramp.PayoutResult
	status      offramp.PayoutResult
	lookup      offramp.PayoutResult
	getErr      error
	initErr     error
	lookupErr   error
	initCalls   int
	statusCalls int
	lookupCalls int
}

func (p *pendingOffRampBridge) GetQuote(ctx context.Context, req offramp.QuoteRequest) (offramp.Quote, error) {
	if p.getErr != nil {
		return offramp.Quote{}, p.getErr
	}
	p.quote.Request = req
	return p.quote, nil
}

func (p *pendingOffRampBridge) InitiatePayout(ctx context.Context, quote offramp.Quote, cryptoTxRef string, destination string, metadata map[string]string) (offramp.PayoutResult, error) {
	p.initCalls++
	if p.initErr != nil {
		return offramp.PayoutResult{}, p.initErr
	}
	p.result.QuoteID = quote.ID
	return p.result, nil
}

func (p *pendingOffRampBridge) GetStatus(ctx context.Context, payoutID string) (offramp.PayoutResult, error) {
	p.statusCalls++
	return p.status, p.getErr
}

func (p *pendingOffRampBridge) FindPayoutByMetadata(ctx context.Context, provider string, metadata map[string]string) (offramp.PayoutResult, error) {
	p.lookupCalls++
	if p.lookupErr != nil {
		return offramp.PayoutResult{}, p.lookupErr
	}
	return p.lookup, nil
}

func (p *pendingOffRampBridge) Cancel(ctx context.Context, payoutID string) error {
	return nil
}

type retryableOffRampBridge struct {
	quote              offramp.Quote
	firstInitErr       error
	pendingResult      offramp.PayoutResult
	completedStatus    offramp.PayoutResult
	lookupResult       offramp.PayoutResult
	invokeCount        int
	statusRequestCount int
	lookupCount        int
}

func (b *retryableOffRampBridge) GetQuote(ctx context.Context, req offramp.QuoteRequest) (offramp.Quote, error) {
	b.quote.Request = req
	return b.quote, nil
}

func (b *retryableOffRampBridge) InitiatePayout(ctx context.Context, quote offramp.Quote, cryptoTxRef string, destination string, metadata map[string]string) (offramp.PayoutResult, error) {
	b.invokeCount++
	b.pendingResult.QuoteID = quote.ID
	if b.invokeCount == 1 && b.firstInitErr != nil {
		return offramp.PayoutResult{}, b.firstInitErr
	}
	return b.pendingResult, nil
}

func (b *retryableOffRampBridge) GetStatus(ctx context.Context, payoutID string) (offramp.PayoutResult, error) {
	b.statusRequestCount++
	return b.completedStatus, nil
}

func (b *retryableOffRampBridge) FindPayoutByMetadata(ctx context.Context, provider string, metadata map[string]string) (offramp.PayoutResult, error) {
	b.lookupCount++
	if b.lookupResult.ID == "" {
		return offramp.PayoutResult{}, errors.New("not found")
	}
	return b.lookupResult, nil
}

func (b *retryableOffRampBridge) Cancel(ctx context.Context, payoutID string) error {
	return nil
}

func (s *KeeperTestSuite) TestExecuteFiatConversionRecoversAmbiguousOffRampSubmission() {
	t := s.T()

	params := s.keeper.GetParams(s.ctx)
	configureCertifiedFiatProfiles(&params)
	params.FiatConversionMinAmount = "1"
	params.FiatConversionMaxAmount = "1000000000"
	params.FiatConversionDailyLimit = "10000000000"
	params.FiatConversionStableDenom = testStableDenom
	params.FiatConversionStableSymbol = testStableSymbol
	params.FiatConversionStableDecimals = 6
	params.FiatConversionMaxSlippage = rate005
	params.FiatConversionMinComplianceStatus = testComplianceCleared
	require.NoError(t, s.keeper.SetParams(s.ctx, params))

	record := veidtypes.NewComplianceRecord(s.provider.String(), s.ctx.BlockTime())
	record.Status = veidtypes.ComplianceStatusCleared
	record.RiskScore = 5
	record.ExpiresAt = s.ctx.BlockTime().Add(24 * time.Hour).Unix()
	s.keeper.SetComplianceKeeper(mockComplianceKeeper{record: record})

	swapExec := &mockSwapExecutor{
		quote: dex.SwapQuote{
			ID: "swap-quote-ambiguous",
			Route: dex.SwapRoute{
				Hops: []dex.SwapHop{{AmountOut: sdkmath.NewInt(900)}},
			},
			ExpiresAt: s.ctx.BlockTime().Add(5 * time.Minute),
		},
		result: dex.SwapResult{
			QuoteID:      "swap-quote-ambiguous",
			TxHash:       "swap-tx-ambiguous",
			InputAmount:  sdkmath.NewInt(1000),
			OutputAmount: sdkmath.NewInt(900),
		},
	}
	s.keeper.SetDexSwapExecutor(swapExec)

	bridge := &retryableOffRampBridge{
		quote: offramp.Quote{
			ID:           "off-quote-ambiguous",
			Provider:     "mock",
			FiatAmount:   sdkmath.LegacyNewDec(100),
			ExchangeRate: sdkmath.LegacyOneDec(),
			CreatedAt:    s.ctx.BlockTime(),
			ExpiresAt:    s.ctx.BlockTime().Add(5 * time.Minute),
		},
		firstInitErr: errors.New("temporary partner timeout"),
		lookupResult: offramp.PayoutResult{
			ID:           "off-payout-ambiguous",
			Status:       offramp.StatusProcessing,
			Provider:     "mock",
			FiatAmount:   sdkmath.LegacyNewDec(100),
			CryptoAmount: sdkmath.NewInt(900),
			Reference:    "ref-ambiguous",
			InitiatedAt:  s.ctx.BlockTime(),
		},
		completedStatus: offramp.PayoutResult{
			ID:           "off-payout-ambiguous",
			Status:       offramp.StatusCompleted,
			Provider:     "mock",
			FiatAmount:   sdkmath.LegacyNewDec(100),
			CryptoAmount: sdkmath.NewInt(900),
			Reference:    "ref-ambiguous",
		},
	}
	s.keeper.SetOffRampBridge(bridge)

	settlement := s.buildSettlement(t, "ambiguous-offramp")
	require.NoError(t, s.keeper.SetSettlement(s.ctx, settlement))

	request := types.FiatConversionRequest{
		InvoiceID:         "inv-ambiguous-offramp",
		SettlementID:      settlement.SettlementID,
		Provider:          settlement.Provider,
		Customer:          settlement.Customer,
		RequestedBy:       settlement.Provider,
		CryptoAmount:      sdk.NewCoin("uve", sdkmath.NewInt(1000)),
		FiatCurrency:      "USD",
		PaymentMethod:     "bank_transfer",
		DestinationHash:   types.HashDestination("acct-token"),
		SlippageTolerance: 0.01,
		CryptoToken:       types.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken:       types.TokenSpec{Symbol: testStableSymbol, Denom: testStableDenom, Decimals: 6},
		EncryptedPayload:  makeEncryptedSettlementPayload(t, []string{"provider-key", "customer-key"}),
	}
	_, err := s.keeper.RequestFiatConversion(s.ctx, request)
	require.NoError(t, err)

	payout, err := s.keeper.ExecutePayout(s.ctx, request.InvoiceID, settlement.SettlementID)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStatePending, payout.State)
	require.Zero(t, bridge.invokeCount)
	require.Zero(t, bridge.lookupCount)
	require.Zero(t, swapExec.quoteCalls)
	require.Zero(t, swapExec.execCalls)

	conversion, found := s.keeper.GetFiatConversionByInvoice(s.ctx, request.InvoiceID)
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateCreated, conversion.State)
	require.Empty(t, conversion.OffRampID)
	require.Empty(t, conversion.OffRampReference)
}

func (s *KeeperTestSuite) TestExecutePayoutSingleSettlement() {
	t := s.T()

	params := s.keeper.GetParams(s.ctx)
	params.PayoutHoldbackRate = testPayoutHoldbackRateTen
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

func (s *KeeperTestSuite) TestDisputePartialRefundReleasesPayout() {
	t := s.T()

	gross := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000)))
	holdback := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(200)))

	payout := types.NewPayoutRecord(
		"payout-partial-1",
		"inv-partial-1",
		"settle-partial-1",
		"escrow-partial-1",
		"order-partial-1",
		"lease-partial-1",
		s.provider.String(),
		s.depositor.String(),
		gross,
		sdk.NewCoins(),
		sdk.NewCoins(),
		holdback,
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
	)

	require.NoError(t, s.keeper.SetPayout(s.ctx, *payout))
	require.NoError(t, s.keeper.HoldPayout(s.ctx, payout.PayoutID, "dispute-partial-1", "partial refund test"))

	customerBefore := s.bankKeeper.GetBalance(s.ctx, s.depositor, "uve").Amount
	providerBefore := s.bankKeeper.GetBalance(s.ctx, s.provider, "uve").Amount

	require.NoError(t, s.keeper.OnDisputeResolved(s.ctx, payout.InvoiceID, billing.DisputeResolutionPartialRefund))

	updated, found := s.keeper.GetPayout(s.ctx, payout.PayoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateCompleted, updated.State)
	require.True(t, updated.HoldbackAmount.IsZero())

	customerAfter := s.bankKeeper.GetBalance(s.ctx, s.depositor, "uve").Amount
	providerAfter := s.bankKeeper.GetBalance(s.ctx, s.provider, "uve").Amount
	treasuryAfter := s.keeper.GetTreasuryBalance(s.ctx)

	require.Equal(t, customerBefore.Add(holdback.AmountOf("uve")), customerAfter)
	require.Equal(t, providerBefore.Add(updated.NetAmount.AmountOf("uve")), providerAfter)
	require.True(t, treasuryAfter.IsZero())
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

func (s *KeeperTestSuite) TestSetPayoutReindexesStateOnUpdate() {
	t := s.T()

	payout := types.NewPayoutRecord(
		"payout-reindex-1",
		"inv-reindex-1",
		"settle-reindex-1",
		"escrow-reindex-1",
		"order-reindex-1",
		"lease-reindex-1",
		s.provider.String(),
		s.depositor.String(),
		sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
		sdk.NewCoins(),
		sdk.NewCoins(),
		sdk.NewCoins(),
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
	)

	require.NoError(t, s.keeper.SetPayout(s.ctx, *payout))
	require.Len(t, s.keeper.GetPayoutsByState(s.ctx, types.PayoutStatePending), 1)

	require.NoError(t, payout.MarkProcessing(s.ctx.BlockTime().Add(time.Minute)))
	require.NoError(t, s.keeper.SetPayout(s.ctx, *payout))

	require.Empty(t, s.keeper.GetPayoutsByState(s.ctx, types.PayoutStatePending))
	require.Len(t, s.keeper.GetPayoutsByState(s.ctx, types.PayoutStateProcessing), 1)
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
		DestinationHash:   types.HashDestination("acct-token"),
		SlippageTolerance: 0.01,
		CryptoToken:       types.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken:       types.TokenSpec{Symbol: testStableSymbol, Denom: testStableDenom, Decimals: 6},
		EncryptedPayload:  makeEncryptedSettlementPayload(t, []string{"provider-key", "customer-key"}),
	}

	_, err := s.keeper.RequestFiatConversion(s.ctx, request)
	require.NoError(t, err)

	payout, err := s.keeper.ExecutePayout(s.ctx, "inv-reconcile", settlement.SettlementID)
	require.NoError(t, err)

	conversion, found := s.keeper.GetFiatConversionByInvoice(s.ctx, "inv-reconcile")
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateCreated, conversion.State)
	require.Equal(t, types.PayoutStatePending, payout.State)
	require.Zero(t, swapExec.quoteCalls)
	require.Zero(t, swapExec.execCalls)
	require.Zero(t, bridge.statusCalls)

	restarted := keeper.NewKeeper(s.cdc, s.keeper.StoreKey(), s.bankKeeper, s.escrow, "authority", mockEncryptionKeeper{})
	restarted.SetDexSwapExecutor(swapExec)
	restarted.SetOffRampBridge(bridge)

	reconciled, err := restarted.ReconcileFiatConversion(s.ctx, conversion.ConversionID)
	require.ErrorIs(t, err, types.ErrExternalIOForbidden)
	require.Equal(t, types.FiatConversionStateCreated, reconciled.State)
	require.Zero(t, bridge.statusCalls)
}

func (s *KeeperTestSuite) TestFiatConversionRetryDoesNotReexecuteSwap() {
	t := s.T()

	swapQuote := dex.SwapQuote{
		ID: "quote-retry-no-dup-swap",
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
			TxHash:       "swap-no-dup",
			InputAmount:  sdkmath.NewInt(1000),
			OutputAmount: sdkmath.NewInt(900),
			ExecutedAt:   s.ctx.BlockTime(),
		},
	}
	s.configureFiatConversion(t, swapExec)

	bridge := &retryableOffRampBridge{
		quote: offramp.Quote{
			ID:         "off-quote-retry",
			FiatAmount: sdkmath.LegacyNewDec(100),
			CreatedAt:  s.ctx.BlockTime(),
			ExpiresAt:  s.ctx.BlockTime().Add(5 * time.Minute),
		},
		firstInitErr: errors.New("temporary partner timeout"),
		pendingResult: offramp.PayoutResult{
			ID:           "off-payout-retry",
			Status:       offramp.StatusProcessing,
			Provider:     "mock",
			FiatAmount:   sdkmath.LegacyNewDec(100),
			CryptoAmount: sdkmath.NewInt(900),
			Reference:    "ref-retry",
			InitiatedAt:  s.ctx.BlockTime(),
		},
		completedStatus: offramp.PayoutResult{
			ID:         "off-payout-retry",
			Status:     offramp.StatusCompleted,
			Provider:   "mock",
			FiatAmount: sdkmath.LegacyNewDec(100),
			Reference:  "ref-retry",
		},
	}
	s.keeper.SetOffRampBridge(bridge)

	settlement := s.buildSettlement(t, "payout-retry-no-dup-swap")
	require.NoError(t, s.keeper.SetSettlement(s.ctx, settlement))

	request := types.FiatConversionRequest{
		InvoiceID:         "inv-retry-no-dup-swap",
		SettlementID:      settlement.SettlementID,
		Provider:          settlement.Provider,
		Customer:          settlement.Customer,
		RequestedBy:       settlement.Provider,
		CryptoAmount:      sdk.NewCoin("uve", sdkmath.NewInt(1000)),
		FiatCurrency:      "USD",
		PaymentMethod:     "bank_transfer",
		DestinationHash:   types.HashDestination("acct-token"),
		SlippageTolerance: 0.01,
		CryptoToken:       types.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken:       types.TokenSpec{Symbol: testStableSymbol, Denom: testStableDenom, Decimals: 6},
		EncryptedPayload:  makeEncryptedSettlementPayload(t, []string{"provider-key", "customer-key"}),
	}
	_, err := s.keeper.RequestFiatConversion(s.ctx, request)
	require.NoError(t, err)

	firstAttempt, err := s.keeper.ExecutePayout(s.ctx, "inv-retry-no-dup-swap", settlement.SettlementID)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStatePending, firstAttempt.State)
	require.Zero(t, swapExec.quoteCalls)
	require.Zero(t, swapExec.execCalls)
	require.Zero(t, bridge.invokeCount)

	require.NoError(t, s.keeper.RetryFailedPayouts(s.ctx))

	secondAttempt, found := s.keeper.GetPayout(s.ctx, firstAttempt.PayoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStatePending, secondAttempt.State)
	require.Zero(t, swapExec.execCalls)
	require.Zero(t, bridge.invokeCount)

	conv, found := s.keeper.GetFiatConversionByInvoice(s.ctx, "inv-retry-no-dup-swap")
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateCreated, conv.State)
	require.Empty(t, conv.SwapTxHash)
}

func (s *KeeperTestSuite) TestProcessInFlightFiatConversionsRecoveryLoop() {
	t := s.T()

	swapQuote := dex.SwapQuote{
		ID: "quote-recovery-loop",
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
			TxHash:       "swap-recovery-loop",
			InputAmount:  sdkmath.NewInt(1000),
			OutputAmount: sdkmath.NewInt(900),
			ExecutedAt:   s.ctx.BlockTime(),
		},
	}

	s.configureFiatConversion(t, swapExec)

	bridge := &pendingOffRampBridge{
		quote: offramp.Quote{
			ID:         "off-quote-recovery-loop",
			FiatAmount: sdkmath.LegacyNewDec(100),
			CreatedAt:  s.ctx.BlockTime(),
			ExpiresAt:  s.ctx.BlockTime().Add(5 * time.Minute),
		},
		result: offramp.PayoutResult{
			ID:           "off-payout-recovery-loop",
			Status:       offramp.StatusProcessing,
			Provider:     "mock",
			FiatAmount:   sdkmath.LegacyNewDec(100),
			CryptoAmount: sdkmath.NewInt(900),
			Reference:    "ref-recovery-loop",
			InitiatedAt:  s.ctx.BlockTime(),
		},
		status: offramp.PayoutResult{
			ID:         "off-payout-recovery-loop",
			Status:     offramp.StatusCompleted,
			Provider:   "mock",
			FiatAmount: sdkmath.LegacyNewDec(100),
			Reference:  "ref-recovery-loop",
		},
	}
	s.keeper.SetOffRampBridge(bridge)

	settlement := s.buildSettlement(t, "payout-recovery-loop")
	require.NoError(t, s.keeper.SetSettlement(s.ctx, settlement))

	request := types.FiatConversionRequest{
		InvoiceID:         "inv-recovery-loop",
		SettlementID:      settlement.SettlementID,
		Provider:          settlement.Provider,
		Customer:          settlement.Customer,
		RequestedBy:       settlement.Provider,
		CryptoAmount:      sdk.NewCoin("uve", sdkmath.NewInt(1000)),
		FiatCurrency:      "USD",
		PaymentMethod:     "bank_transfer",
		DestinationHash:   types.HashDestination("acct-token"),
		SlippageTolerance: 0.01,
		CryptoToken:       types.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken:       types.TokenSpec{Symbol: testStableSymbol, Denom: testStableDenom, Decimals: 6},
		EncryptedPayload:  makeEncryptedSettlementPayload(t, []string{"provider-key", "customer-key"}),
	}
	_, err := s.keeper.RequestFiatConversion(s.ctx, request)
	require.NoError(t, err)

	initialPayout, err := s.keeper.ExecutePayout(s.ctx, "inv-recovery-loop", settlement.SettlementID)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStatePending, initialPayout.State)

	initialConversion, found := s.keeper.GetFiatConversionByInvoice(s.ctx, "inv-recovery-loop")
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateCreated, initialConversion.State)
	require.Zero(t, bridge.initCalls)
	require.Zero(t, swapExec.quoteCalls)
	require.Zero(t, swapExec.execCalls)

	restarted := keeper.NewKeeper(s.cdc, s.keeper.StoreKey(), s.bankKeeper, s.escrow, "authority", mockEncryptionKeeper{})
	restarted.SetDexSwapExecutor(swapExec)
	restarted.SetOffRampBridge(bridge)

	require.NoError(t, restarted.ProcessInFlightFiatConversions(s.ctx))

	recoveredConversion, found := restarted.GetFiatConversion(s.ctx, initialConversion.ConversionID)
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateCreated, recoveredConversion.State)

	recoveredPayout, found := restarted.GetPayout(s.ctx, initialPayout.PayoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStatePending, recoveredPayout.State)

	require.NoError(t, restarted.ProcessInFlightFiatConversions(s.ctx))
	postDuplicate, found := restarted.GetPayout(s.ctx, initialPayout.PayoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStatePending, postDuplicate.State)
	require.Zero(t, bridge.initCalls)
	require.Zero(t, bridge.statusCalls)
}
