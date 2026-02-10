package keeper_test

import (
	"context"
	"errors"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/dex"
	"github.com/virtengine/virtengine/pkg/payments/offramp"
	"github.com/virtengine/virtengine/x/settlement/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

type mockSwapExecutor struct {
	quote    dex.SwapQuote
	result   dex.SwapResult
	quoteErr error
	execErr  error
}

func (m *mockSwapExecutor) GetQuote(ctx context.Context, request dex.SwapRequest) (dex.SwapQuote, error) {
	if m.quoteErr != nil {
		return dex.SwapQuote{}, m.quoteErr
	}
	return m.quote, nil
}

func (m *mockSwapExecutor) ExecuteSwap(ctx context.Context, quote dex.SwapQuote, signedTx []byte) (dex.SwapResult, error) {
	if m.execErr != nil {
		return dex.SwapResult{}, m.execErr
	}
	return m.result, nil
}

type mockComplianceKeeper struct {
	record *veidtypes.ComplianceRecord
}

func (m mockComplianceKeeper) GetComplianceRecord(ctx sdk.Context, address string) (*veidtypes.ComplianceRecord, bool) {
	if m.record == nil {
		return nil, false
	}
	return m.record, true
}

func (s *KeeperTestSuite) configureFiatConversion(t *testing.T, swapExec *mockSwapExecutor) {
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

	bridge := offramp.NewBridge()
	require.NoError(t, bridge.RegisterAdapter(offramp.NewMockProvider("mock", []string{"USD"}, []string{"bank_transfer"})))
	s.keeper.SetOffRampBridge(bridge)
	s.keeper.SetComplianceKeeper(mockComplianceKeeper{record: record})
}

func (s *KeeperTestSuite) buildSettlement(t *testing.T, settlementID string) types.SettlementRecord {
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "order-"+settlementID, s.depositor, amount, 24*time.Hour, nil)
	require.NoError(t, err)
	require.NoError(t, s.keeper.ActivateEscrow(s.ctx, escrowID, "lease-"+settlementID, s.provider))

	return *types.NewSettlementRecord(
		settlementID,
		escrowID,
		"order-"+settlementID,
		"lease-"+settlementID,
		s.provider.String(),
		s.depositor.String(),
		sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
		sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
		sdk.NewCoins(),
		sdk.NewCoins(),
		nil,
		0,
		s.ctx.BlockTime().Add(-time.Hour),
		s.ctx.BlockTime(),
		types.SettlementTypeUsageBased,
		false,
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
	)
}

func (s *KeeperTestSuite) TestFiatConversionPayoutSuccess() {
	t := s.T()

	swapQuote := dex.SwapQuote{
		ID: "quote-1",
		Route: dex.SwapRoute{
			Hops: []dex.SwapHop{
				{
					AmountOut: sdkmath.NewInt(900),
				},
			},
		},
		ExpiresAt: s.ctx.BlockTime().Add(5 * time.Minute),
	}

	swapExec := &mockSwapExecutor{
		quote: swapQuote,
		result: dex.SwapResult{
			QuoteID:      swapQuote.ID,
			TxHash:       "swap-tx",
			InputAmount:  sdkmath.NewInt(1000),
			OutputAmount: sdkmath.NewInt(900),
			ExecutedAt:   s.ctx.BlockTime(),
		},
	}

	s.configureFiatConversion(t, swapExec)

	settlement := s.buildSettlement(t, "settle-1")
	require.NoError(t, s.keeper.SetSettlement(s.ctx, settlement))

	request := types.FiatConversionRequest{
		InvoiceID:         "inv-1",
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

	payout, err := s.keeper.ExecutePayout(s.ctx, "inv-1", settlement.SettlementID)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStateCompleted, payout.State)
	require.NotEmpty(t, payout.FiatConversionID)

	conversion, found := s.keeper.GetFiatConversionByPayout(s.ctx, payout.PayoutID)
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateCompleted, conversion.State)
	require.NotEmpty(t, conversion.OffRampID)
	require.NotEmpty(t, conversion.SwapTxHash)
}

func (s *KeeperTestSuite) TestFiatConversionPayoutSwapFailure() {
	t := s.T()

	swapExec := &mockSwapExecutor{
		quote: dex.SwapQuote{
			ID: "quote-2",
			Route: dex.SwapRoute{
				Hops: []dex.SwapHop{
					{
						AmountOut: sdkmath.NewInt(900),
					},
				},
			},
			ExpiresAt: s.ctx.BlockTime().Add(5 * time.Minute),
		},
		execErr: errors.New("swap failed"),
	}

	s.configureFiatConversion(t, swapExec)

	settlement := s.buildSettlement(t, "settle-2")
	require.NoError(t, s.keeper.SetSettlement(s.ctx, settlement))

	request := types.FiatConversionRequest{
		InvoiceID:         "inv-2",
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

	payout, err := s.keeper.ExecutePayout(s.ctx, "inv-2", settlement.SettlementID)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStateFailed, payout.State)

	conversion, found := s.keeper.GetFiatConversionByInvoice(s.ctx, "inv-2")
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateFailed, conversion.State)
}
