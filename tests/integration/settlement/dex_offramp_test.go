//go:build e2e.integration

package settlement_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/dex"
	"github.com/virtengine/virtengine/pkg/payments/offramp"
	"github.com/virtengine/virtengine/testutil/state"
	"github.com/virtengine/virtengine/x/settlement/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

type swapExecutor struct {
	quote  dex.SwapQuote
	result dex.SwapResult
}

func (s swapExecutor) GetQuote(ctx context.Context, request dex.SwapRequest) (dex.SwapQuote, error) {
	return s.quote, nil
}

func (s swapExecutor) ExecuteSwap(ctx context.Context, quote dex.SwapQuote, signedTx []byte) (dex.SwapResult, error) {
	return s.result, nil
}

type pendingOffRamp struct {
	status offramp.PayoutResult
}

func (p *pendingOffRamp) Name() string { return "pending-mock" }
func (p *pendingOffRamp) GetQuote(ctx context.Context, req offramp.QuoteRequest) (offramp.Quote, error) {
	return offramp.Quote{
		ID:           "quote-pending",
		Request:      req,
		FiatAmount:   sdkmath.LegacyNewDec(100),
		ExchangeRate: sdkmath.LegacyNewDec(1),
		Fee:          sdkmath.NewInt(1),
		Provider:     p.Name(),
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
	}, nil
}
func (p *pendingOffRamp) InitiatePayout(ctx context.Context, req offramp.PayoutRequest) (offramp.PayoutResult, error) {
	p.status = offramp.PayoutResult{
		ID:           "pending-payout",
		QuoteID:      req.Quote.ID,
		Status:       offramp.StatusProcessing,
		Provider:     p.Name(),
		FiatAmount:   req.Quote.FiatAmount,
		CryptoAmount: req.Quote.Request.CryptoAmount,
		Fee:          req.Quote.Fee,
		Reference:    "pending-ref",
		InitiatedAt:  time.Now().UTC(),
	}
	return p.status, nil
}
func (p *pendingOffRamp) GetStatus(ctx context.Context, payoutID string) (offramp.PayoutResult, error) {
	completedAt := time.Now().UTC()
	p.status.Status = offramp.StatusCompleted
	p.status.CompletedAt = &completedAt
	return p.status, nil
}
func (p *pendingOffRamp) Cancel(ctx context.Context, payoutID string) error { return nil }
func (p *pendingOffRamp) SupportsCurrency(currency string) bool             { return currency == "USD" }
func (p *pendingOffRamp) SupportsMethod(method string) bool                 { return method == "bank_transfer" }
func (p *pendingOffRamp) IsHealthy(ctx context.Context) bool                { return true }

func setupConversionDeps(t *testing.T, suite *state.TestSuite) *swapExecutor {
	ctx := suite.Context()
	app := suite.App()

	swapQuote := dex.SwapQuote{
		ID: "quote-int",
		Route: dex.SwapRoute{
			Hops: []dex.SwapHop{
				{AmountOut: sdkmath.NewInt(900)},
			},
		},
		ExpiresAt: ctx.BlockTime().Add(5 * time.Minute),
	}
	swapExec := &swapExecutor{
		quote: swapQuote,
		result: dex.SwapResult{
			QuoteID:      swapQuote.ID,
			TxHash:       "swap-int",
			InputAmount:  sdkmath.NewInt(1000),
			OutputAmount: sdkmath.NewInt(900),
			ExecutedAt:   ctx.BlockTime(),
		},
	}

	bridge := offramp.NewBridge()
	require.NoError(t, bridge.RegisterAdapter(offramp.NewMockProvider("mock", []string{"USD"}, []string{"bank_transfer"})))

	keeper := &app.Keepers.VirtEngine.Settlement
	keeper.SetDexSwapExecutor(swapExec)
	keeper.SetOffRampBridge(bridge)
	keeper.SetComplianceKeeper(app.Keepers.VirtEngine.VEID)

	params := keeper.GetParams(ctx)
	params.FiatConversionEnabled = true
	params.FiatConversionMinAmount = "1"
	params.FiatConversionMaxAmount = "1000000000"
	params.FiatConversionDailyLimit = "10000000000"
	params.FiatConversionStableDenom = "uusdc"
	params.FiatConversionStableSymbol = "USDC"
	params.FiatConversionStableDecimals = 6
	params.FiatConversionMaxSlippage = "0.05"
	params.FiatConversionMinComplianceStatus = "CLEARED"
	require.NoError(t, keeper.SetParams(ctx, params))

	return swapExec
}

func seedComplianceRecord(t *testing.T, suite *state.TestSuite, provider sdk.AccAddress) {
	ctx := suite.Context()
	veid := suite.App().Keepers.VirtEngine.VEID
	record := veidtypes.NewComplianceRecord(provider.String(), ctx.BlockTime())
	record.Status = veidtypes.ComplianceStatusCleared
	record.RiskScore = 5
	record.ExpiresAt = ctx.BlockTime().Add(24 * time.Hour).Unix()
	require.NoError(t, veid.SetComplianceRecord(ctx, record))
}

func fundAccount(t *testing.T, suite *state.TestSuite, addr sdk.AccAddress, coins sdk.Coins) {
	ctx := suite.Context()
	bank := suite.App().Keepers.Cosmos.Bank
	require.NoError(t, bank.MintCoins(ctx, minttypes.ModuleName, coins))
	require.NoError(t, bank.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, coins))
}

func TestFiatConversionPipelineSuccess(t *testing.T) {
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

	escrowID, err := keeper.CreateEscrow(ctx, "order-int-1", depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))), 24*time.Hour, nil)
	require.NoError(t, err)
	require.NoError(t, keeper.ActivateEscrow(ctx, escrowID, "lease-int-1", provider))

	usage := &types.UsageRecord{
		OrderID:           "order-int-1",
		LeaseID:           "lease-int-1",
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

	settlement, err := keeper.SettleOrder(ctx, "order-int-1", []string{usage.UsageID}, false)
	require.NoError(t, err)

	payout, found := keeper.GetPayoutBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateCompleted, payout.State)

	conversion, found := keeper.GetFiatConversionBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateCompleted, conversion.State)
	require.NotEmpty(t, conversion.OffRampID)
}

func TestFiatConversionReconciliation(t *testing.T) {
	suite := state.SetupTestSuite(t)
	ctx := suite.Context()
	app := suite.App()
	keeper := &app.Keepers.VirtEngine.Settlement

	swapExec := setupConversionDeps(t, suite)

	// Override bridge with pending adapter
	bridge := offramp.NewBridge()
	pending := &pendingOffRamp{}
	require.NoError(t, bridge.RegisterAdapter(pending))
	keeper.SetOffRampBridge(bridge)
	keeper.SetDexSwapExecutor(swapExec)
	keeper.SetComplianceKeeper(app.Keepers.VirtEngine.VEID)

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

	escrowID, err := keeper.CreateEscrow(ctx, "order-int-2", depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))), 24*time.Hour, nil)
	require.NoError(t, err)
	require.NoError(t, keeper.ActivateEscrow(ctx, escrowID, "lease-int-2", provider))

	usage := &types.UsageRecord{
		OrderID:           "order-int-2",
		LeaseID:           "lease-int-2",
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

	settlement, err := keeper.SettleOrder(ctx, "order-int-2", []string{usage.UsageID}, false)
	require.NoError(t, err)

	conversion, found := keeper.GetFiatConversionBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateOffRampPending, conversion.State)

	reconciled, err := keeper.ReconcileFiatConversion(ctx, conversion.ConversionID)
	require.NoError(t, err)
	require.Equal(t, types.FiatConversionStateCompleted, reconciled.State)

	payout, found := keeper.GetPayoutBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateCompleted, payout.State)
	require.Equal(t, fmt.Sprintf("fiat-%s", conversion.ConversionID), payout.TxHash)
}
