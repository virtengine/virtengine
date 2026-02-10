package keeper_test

import (
	"context"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/dex"
	"github.com/virtengine/virtengine/pkg/payments/offramp"
	"github.com/virtengine/virtengine/x/settlement/keeper"
	"github.com/virtengine/virtengine/x/settlement/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

type capturingSwapExecutor struct {
	quote       dex.SwapQuote
	result      dex.SwapResult
	quoteErr    error
	execErr     error
	lastRequest dex.SwapRequest
}

func (c *capturingSwapExecutor) GetQuote(ctx context.Context, request dex.SwapRequest) (dex.SwapQuote, error) {
	c.lastRequest = request
	if c.quoteErr != nil {
		return dex.SwapQuote{}, c.quoteErr
	}
	return c.quote, nil
}

func (c *capturingSwapExecutor) ExecuteSwap(ctx context.Context, quote dex.SwapQuote, signedTx []byte) (dex.SwapResult, error) {
	if c.execErr != nil {
		return dex.SwapResult{}, c.execErr
	}
	return c.result, nil
}

type capturingOffRampBridge struct {
	quote        offramp.Quote
	result       offramp.PayoutResult
	quoteErr     error
	initErr      error
	lastQuoteReq offramp.QuoteRequest
	lastTxRef    string
}

func (c *capturingOffRampBridge) GetQuote(ctx context.Context, req offramp.QuoteRequest) (offramp.Quote, error) {
	c.lastQuoteReq = req
	if c.quoteErr != nil {
		return offramp.Quote{}, c.quoteErr
	}
	c.quote.Request = req
	return c.quote, nil
}

func (c *capturingOffRampBridge) InitiatePayout(ctx context.Context, quote offramp.Quote, cryptoTxRef string, destination string, metadata map[string]string) (offramp.PayoutResult, error) {
	c.lastTxRef = cryptoTxRef
	if c.initErr != nil {
		return offramp.PayoutResult{}, c.initErr
	}
	c.result.QuoteID = quote.ID
	return c.result, nil
}

func (c *capturingOffRampBridge) GetStatus(ctx context.Context, payoutID string) (offramp.PayoutResult, error) {
	return c.result, nil
}

func (c *capturingOffRampBridge) Cancel(ctx context.Context, payoutID string) error {
	return nil
}

func (s *KeeperTestSuite) configureFiatConversionDeps(t *testing.T, swapExec keeper.DexSwapExecutor, bridge keeper.OffRampBridge) {
	params := s.keeper.GetParams(s.ctx)
	params.FiatConversionEnabled = true
	params.FiatConversionMinAmount = "1"
	params.FiatConversionMaxAmount = "1000000000"
	params.FiatConversionDailyLimit = "10000000000"
	params.FiatConversionStableDenom = "uusdc"
	params.FiatConversionStableSymbol = "USDC"
	params.FiatConversionStableDecimals = 6
	params.FiatConversionMaxSlippage = rate005
	params.FiatConversionMinComplianceStatus = "CLEARED"
	require.NoError(t, s.keeper.SetParams(s.ctx, params))

	record := veidtypes.NewComplianceRecord(s.provider.String(), s.ctx.BlockTime())
	record.Status = veidtypes.ComplianceStatusCleared
	record.RiskScore = 5
	record.ExpiresAt = s.ctx.BlockTime().Add(24 * time.Hour).Unix()

	s.keeper.SetDexSwapExecutor(swapExec)
	s.keeper.SetOffRampBridge(bridge)
	s.keeper.SetComplianceKeeper(mockComplianceKeeper{record: record})
}

func (s *KeeperTestSuite) TestFiatConversionMultiHopAndAuditTrail() {
	t := s.T()

	swapQuote := dex.SwapQuote{
		ID: "quote-multi",
		Route: dex.SwapRoute{
			Hops: []dex.SwapHop{
				{AmountOut: sdkmath.NewInt(800)},
				{AmountOut: sdkmath.NewInt(900)},
			},
		},
		ExpiresAt: s.ctx.BlockTime().Add(5 * time.Minute),
	}
	swapExec := &capturingSwapExecutor{
		quote: swapQuote,
		result: dex.SwapResult{
			QuoteID:      swapQuote.ID,
			TxHash:       "swap-multi",
			InputAmount:  sdkmath.NewInt(1000),
			OutputAmount: sdkmath.NewInt(900),
			ExecutedAt:   s.ctx.BlockTime(),
		},
	}

	bridge := &capturingOffRampBridge{
		quote: offramp.Quote{
			ID:         "off-quote",
			FiatAmount: sdkmath.LegacyNewDec(100),
			CreatedAt:  s.ctx.BlockTime(),
			ExpiresAt:  s.ctx.BlockTime().Add(5 * time.Minute),
		},
		result: offramp.PayoutResult{
			ID:           "off-payout",
			Status:       offramp.StatusCompleted,
			Provider:     "mock",
			FiatAmount:   sdkmath.LegacyNewDec(100),
			CryptoAmount: sdkmath.NewInt(900),
			Reference:    "ref-1",
			InitiatedAt:  s.ctx.BlockTime(),
		},
	}

	s.configureFiatConversionDeps(t, swapExec, bridge)

	settlement := s.buildSettlement(t, "fiat-multi")
	require.NoError(t, s.keeper.SetSettlement(s.ctx, settlement))

	request := types.FiatConversionRequest{
		InvoiceID:         "inv-multi",
		SettlementID:      settlement.SettlementID,
		Provider:          settlement.Provider,
		Customer:          settlement.Customer,
		RequestedBy:       settlement.Provider,
		CryptoAmount:      sdk.NewCoin("uve", sdkmath.NewInt(1000)),
		FiatCurrency:      "USD",
		PaymentMethod:     "bank_transfer",
		DestinationHash:   types.HashDestination("acct-token"),
		SlippageTolerance: 0.2,
		CryptoToken:       types.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken:       types.TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6},
		EncryptedPayload:  makeEncryptedSettlementPayload(t, []string{"provider-key", "customer-key"}),
	}

	_, err := s.keeper.RequestFiatConversion(s.ctx, request)
	require.NoError(t, err)

	payout, err := s.keeper.ExecutePayout(s.ctx, "inv-multi", settlement.SettlementID)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStateCompleted, payout.State)

	conversion, found := s.keeper.GetFiatConversionByInvoice(s.ctx, "inv-multi")
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateCompleted, conversion.State)
	require.Equal(t, sdkmath.NewInt(900), conversion.StableAmount.Amount)

	require.InEpsilon(t, 0.05, swapExec.lastRequest.SlippageTolerance, 0.0001)

	actions := make(map[string]bool)
	for _, entry := range conversion.AuditTrail {
		actions[entry.Action] = true
	}
	require.True(t, actions["conversion_requested"])
	require.True(t, actions["swap_requested"])
	require.True(t, actions["swap_executed"])
	require.True(t, actions["offramp_initiated"])
}

func (s *KeeperTestSuite) TestFiatConversionLimitExceeded() {
	t := s.T()

	swapQuote := dex.SwapQuote{
		ID: "quote-limit",
		Route: dex.SwapRoute{
			Hops: []dex.SwapHop{
				{AmountOut: sdkmath.NewInt(900)},
			},
		},
		ExpiresAt: s.ctx.BlockTime().Add(5 * time.Minute),
	}
	swapExec := &capturingSwapExecutor{
		quote: swapQuote,
		result: dex.SwapResult{
			QuoteID:      swapQuote.ID,
			TxHash:       "swap-limit",
			InputAmount:  sdkmath.NewInt(1000),
			OutputAmount: sdkmath.NewInt(900),
			ExecutedAt:   s.ctx.BlockTime(),
		},
	}
	bridge := &capturingOffRampBridge{}

	s.configureFiatConversionDeps(t, swapExec, bridge)

	params := s.keeper.GetParams(s.ctx)
	params.FiatConversionMaxAmount = "800"
	require.NoError(t, s.keeper.SetParams(s.ctx, params))

	settlement := s.buildSettlement(t, "fiat-limit")
	require.NoError(t, s.keeper.SetSettlement(s.ctx, settlement))

	request := types.FiatConversionRequest{
		InvoiceID:         "inv-limit",
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
		StableToken:       types.TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6},
		EncryptedPayload:  makeEncryptedSettlementPayload(t, []string{"provider-key", "customer-key"}),
	}

	_, err := s.keeper.RequestFiatConversion(s.ctx, request)
	require.NoError(t, err)

	payout, err := s.keeper.ExecutePayout(s.ctx, "inv-limit", settlement.SettlementID)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStateFailed, payout.State)

	conversion, found := s.keeper.GetFiatConversionByInvoice(s.ctx, "inv-limit")
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateFailed, conversion.State)
}

func (s *KeeperTestSuite) TestFiatConversionStaleQuoteFailure() {
	t := s.T()

	swapQuote := dex.SwapQuote{
		ID: "quote-stale",
		Route: dex.SwapRoute{
			Hops: []dex.SwapHop{
				{AmountOut: sdkmath.NewInt(900)},
			},
		},
		ExpiresAt: s.ctx.BlockTime().Add(-1 * time.Minute),
	}
	swapExec := &capturingSwapExecutor{
		quote: swapQuote,
		result: dex.SwapResult{
			QuoteID:      swapQuote.ID,
			TxHash:       "swap-stale",
			InputAmount:  sdkmath.NewInt(1000),
			OutputAmount: sdkmath.NewInt(900),
			ExecutedAt:   s.ctx.BlockTime(),
		},
	}
	bridge := &capturingOffRampBridge{}

	s.configureFiatConversionDeps(t, swapExec, bridge)

	settlement := s.buildSettlement(t, "fiat-stale")
	require.NoError(t, s.keeper.SetSettlement(s.ctx, settlement))

	request := types.FiatConversionRequest{
		InvoiceID:         "inv-stale",
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
		StableToken:       types.TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6},
		EncryptedPayload:  makeEncryptedSettlementPayload(t, []string{"provider-key", "customer-key"}),
	}

	_, err := s.keeper.RequestFiatConversion(s.ctx, request)
	require.NoError(t, err)

	payout, err := s.keeper.ExecutePayout(s.ctx, "inv-stale", settlement.SettlementID)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStateFailed, payout.State)

	conversion, found := s.keeper.GetFiatConversionByInvoice(s.ctx, "inv-stale")
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateFailed, conversion.State)
	require.True(t, strings.Contains(conversion.FailureReason, "swap quote expired"))
}

func (s *KeeperTestSuite) TestFiatConversionFailureRecovery() {
	t := s.T()

	swapQuote := dex.SwapQuote{
		ID: "quote-failure",
		Route: dex.SwapRoute{
			Hops: []dex.SwapHop{
				{AmountOut: sdkmath.NewInt(900)},
			},
		},
		ExpiresAt: s.ctx.BlockTime().Add(5 * time.Minute),
	}
	swapExec := &capturingSwapExecutor{
		quote: swapQuote,
		result: dex.SwapResult{
			QuoteID:      swapQuote.ID,
			TxHash:       "swap-failed",
			InputAmount:  sdkmath.NewInt(1000),
			OutputAmount: sdkmath.NewInt(900),
			ExecutedAt:   s.ctx.BlockTime(),
		},
	}

	bridge := &capturingOffRampBridge{
		result: offramp.PayoutResult{
			ID:     "off-failed",
			Status: offramp.StatusFailed,
		},
	}

	s.configureFiatConversionDeps(t, swapExec, bridge)

	settlement := s.buildSettlement(t, "fiat-failure")
	require.NoError(t, s.keeper.SetSettlement(s.ctx, settlement))

	request := types.FiatConversionRequest{
		InvoiceID:         "inv-failure",
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
		StableToken:       types.TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6},
		EncryptedPayload:  makeEncryptedSettlementPayload(t, []string{"provider-key", "customer-key"}),
	}

	_, err := s.keeper.RequestFiatConversion(s.ctx, request)
	require.NoError(t, err)

	payout, err := s.keeper.ExecutePayout(s.ctx, "inv-failure", settlement.SettlementID)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStateFailed, payout.State)

	conversion, found := s.keeper.GetFiatConversionByInvoice(s.ctx, "inv-failure")
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateFailed, conversion.State)
}
