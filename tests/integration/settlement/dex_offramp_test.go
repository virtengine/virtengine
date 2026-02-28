//go:build e2e.integration

package settlement_test

import (
	"context"
	"encoding/json"
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
	return setupConversionDepsWithBridge(t, suite, nil)
}

func setupConversionDepsWithBridge(t *testing.T, suite *state.TestSuite, bridge offramp.Bridge) *swapExecutor {
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

	if bridge == nil {
		bridge = offramp.NewBridge()
		require.NoError(t, bridge.RegisterAdapter(offramp.NewMockProvider("mock", []string{"USD"}, []string{"bank_transfer"})))
	}

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

type financeEvidenceRecord struct {
	Provider                 string   `json:"provider"`
	InvoiceID                string   `json:"invoice_id"`
	SettlementID             string   `json:"settlement_id"`
	PayoutID                 string   `json:"payout_id"`
	PayoutState              string   `json:"payout_state"`
	PayoutTxHash             string   `json:"payout_tx_hash"`
	PayoutIdempotencyKey     string   `json:"payout_idempotency_key"`
	PayoutLedgerEntryTypes   []string `json:"payout_ledger_entry_types"`
	TreasuryBalance          string   `json:"treasury_balance"`
	ExpectedTreasuryBalance  string   `json:"expected_treasury_balance"`
	ConversionID             string   `json:"conversion_id"`
	ConversionState          string   `json:"conversion_state"`
	ConversionIdempotencyKey string   `json:"conversion_idempotency_key"`
	OffRampProvider          string   `json:"off_ramp_provider"`
	OffRampQuoteID           string   `json:"off_ramp_quote_id"`
	OffRampID                string   `json:"off_ramp_id"`
	OffRampStatus            string   `json:"off_ramp_status"`
	OffRampReference         string   `json:"off_ramp_reference"`
	BridgeStatus             string   `json:"bridge_status"`
	BridgeReference          string   `json:"bridge_reference"`
	BridgeQuoteID            string   `json:"bridge_quote_id"`
	ConversionAuditActions   []string `json:"conversion_audit_actions"`
	TransitionCount          int      `json:"transition_count"`
}

func financeReconciliationMetadata(conversion types.FiatConversionRecord, payout types.PayoutRecord) map[string]string {
	metadata := map[string]string{
		"conversion_id": conversion.ConversionID,
	}
	if conversion.IdempotencyKey != "" {
		metadata["idempotency_key"] = conversion.IdempotencyKey
	}
	if conversion.InvoiceID != "" {
		metadata["invoice_id"] = conversion.InvoiceID
	}
	if payout.PayoutID != "" {
		metadata["payout_id"] = payout.PayoutID
	}
	if conversion.SettlementID != "" {
		metadata["settlement_id"] = conversion.SettlementID
	}
	return metadata
}

func payoutLedgerEntryTypes(entries []types.PayoutLedgerEntry) []string {
	typesList := make([]string, 0, len(entries))
	for _, entry := range entries {
		typesList = append(typesList, entry.EntryType.String())
	}
	return typesList
}

func conversionAuditActions(entries []types.FiatConversionAuditEntry) []string {
	actions := make([]string, 0, len(entries))
	for _, entry := range entries {
		actions = append(actions, entry.Action)
	}
	return actions
}

func emitFinanceEvidence(
	t *testing.T,
	payout types.PayoutRecord,
	conversion types.FiatConversionRecord,
	bridgeResult offramp.PayoutResult,
	ledger []types.PayoutLedgerEntry,
	treasury sdk.Coins,
	expectedTreasury sdk.Coins,
) {
	t.Helper()

	record := financeEvidenceRecord{
		Provider:                 payout.Provider,
		InvoiceID:                payout.InvoiceID,
		SettlementID:             payout.SettlementID,
		PayoutID:                 payout.PayoutID,
		PayoutState:              string(payout.State),
		PayoutTxHash:             payout.TxHash,
		PayoutIdempotencyKey:     payout.IdempotencyKey,
		PayoutLedgerEntryTypes:   payoutLedgerEntryTypes(ledger),
		TreasuryBalance:          treasury.String(),
		ExpectedTreasuryBalance:  expectedTreasury.String(),
		ConversionID:             conversion.ConversionID,
		ConversionState:          string(conversion.State),
		ConversionIdempotencyKey: conversion.IdempotencyKey,
		OffRampProvider:          conversion.OffRampProvider,
		OffRampQuoteID:           conversion.OffRampQuoteID,
		OffRampID:                conversion.OffRampID,
		OffRampStatus:            conversion.OffRampStatus,
		OffRampReference:         conversion.OffRampReference,
		BridgeStatus:             string(bridgeResult.Status),
		BridgeReference:          bridgeResult.Reference,
		BridgeQuoteID:            bridgeResult.QuoteID,
		ConversionAuditActions:   conversionAuditActions(conversion.AuditTrail),
		TransitionCount:          len(conversion.TransitionHistory),
	}

	payload, err := json.Marshal(record)
	require.NoError(t, err)
	t.Logf("finance-evidence=%s", payload)
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
	suite := state.SetupTestSuiteWithoutModuleServices(t)
	ctx := suite.Context()
	app := suite.App()
	keeper := &app.Keepers.VirtEngine.Settlement

	bridge := offramp.NewBridge()
	require.NoError(t, bridge.RegisterAdapter(offramp.NewMockProvider("mock", []string{"USD"}, []string{"bank_transfer"})))
	setupConversionDepsWithBridge(t, suite, bridge)

	depositor := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	provider := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	seedComplianceRecord(t, suite, provider)
	fundAccount(t, suite, depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100000))))

	pref := makeFiatPayoutPreference(t, suite, ctx.BlockTime(), provider, "uve", "acct-token")
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
	require.NotEmpty(t, payout.IdempotencyKey)

	conversion, found := keeper.GetFiatConversionBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateCompleted, conversion.State)
	require.NotEmpty(t, conversion.OffRampID)
	require.NotEmpty(t, conversion.OffRampProvider)
	require.NotEmpty(t, conversion.OffRampQuoteID)
	require.Equal(t, string(offramp.StatusCompleted), conversion.OffRampStatus)
	require.NotEmpty(t, conversion.OffRampReference)
	require.NotEmpty(t, conversion.IdempotencyKey)
	require.NotEmpty(t, conversion.TransitionHistory)
	require.Equal(t, types.FiatConversionStatePayoutCompleted, conversion.TransitionHistory[len(conversion.TransitionHistory)-1].To)
	require.Equal(t, fmt.Sprintf("fiat-%s", conversion.ConversionID), payout.TxHash)

	metadata := financeReconciliationMetadata(conversion, payout)
	bridgeResult, err := bridge.FindPayoutByMetadata(ctx, conversion.OffRampProvider, metadata)
	require.NoError(t, err)
	require.Equal(t, conversion.OffRampID, bridgeResult.ID)
	require.Equal(t, conversion.OffRampQuoteID, bridgeResult.QuoteID)
	require.Equal(t, conversion.OffRampReference, bridgeResult.Reference)
	require.Equal(t, offramp.StatusCompleted, bridgeResult.Status)
	require.Equal(t, metadata, bridgeResult.Metadata)

	ledger := keeper.GetPayoutLedgerEntries(ctx, payout.PayoutID)
	require.NotEmpty(t, ledger)
	require.Contains(t, payoutLedgerEntryTypes(ledger), types.PayoutLedgerEntryCompleted.String())

	expectedTreasury := payout.PlatformFee.Add(payout.ValidatorFee...)
	treasury := keeper.GetTreasuryBalance(ctx)
	require.True(t, treasury.Equal(expectedTreasury))

	emitFinanceEvidence(t, payout, conversion, bridgeResult, ledger, treasury, expectedTreasury)
}

func TestFiatConversionReconciliation(t *testing.T) {
	suite := state.SetupTestSuiteWithoutModuleServices(t)
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

	pref := makeFiatPayoutPreference(t, suite, ctx.BlockTime(), provider, "uve", "acct-token")
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
	require.Equal(t, types.FiatConversionStatePayoutSubmitted, conversion.State)
	require.Equal(t, string(offramp.StatusProcessing), conversion.OffRampStatus)
	require.Equal(t, "pending-ref", conversion.OffRampReference)
	require.NotEmpty(t, conversion.OffRampID)
	require.NotEmpty(t, conversion.OffRampQuoteID)
	require.NotEmpty(t, conversion.IdempotencyKey)

	payout, found := keeper.GetPayoutBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)

	metadata := financeReconciliationMetadata(conversion, payout)
	bridgePending, err := bridge.FindPayoutByMetadata(ctx, pending.Name(), metadata)
	require.NoError(t, err)
	require.Equal(t, conversion.OffRampID, bridgePending.ID)
	require.Equal(t, offramp.StatusProcessing, bridgePending.Status)
	require.Equal(t, "pending-ref", bridgePending.Reference)
	require.Equal(t, metadata, bridgePending.Metadata)

	reconciled, err := keeper.ReconcileFiatConversion(ctx, conversion.ConversionID)
	require.NoError(t, err)
	require.Equal(t, types.FiatConversionStateCompleted, reconciled.State)
	require.Equal(t, string(offramp.StatusCompleted), reconciled.OffRampStatus)
	require.Equal(t, "pending-ref", reconciled.OffRampReference)
	require.NotEmpty(t, reconciled.TransitionHistory)
	require.Equal(t, types.FiatConversionStatePayoutCompleted, reconciled.TransitionHistory[len(reconciled.TransitionHistory)-1].To)

	payout, found = keeper.GetPayoutBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateCompleted, payout.State)
	require.Equal(t, fmt.Sprintf("fiat-%s", conversion.ConversionID), payout.TxHash)

	bridgeCompleted, err := bridge.GetStatus(ctx, reconciled.OffRampID)
	require.NoError(t, err)
	require.Equal(t, offramp.StatusCompleted, bridgeCompleted.Status)
	require.Equal(t, reconciled.OffRampReference, bridgeCompleted.Reference)
	require.Equal(t, metadata, bridgeCompleted.Metadata)

	ledger := keeper.GetPayoutLedgerEntries(ctx, payout.PayoutID)
	require.NotEmpty(t, ledger)
	require.Contains(t, payoutLedgerEntryTypes(ledger), types.PayoutLedgerEntryCompleted.String())

	expectedTreasury := payout.PlatformFee.Add(payout.ValidatorFee...)
	treasury := keeper.GetTreasuryBalance(ctx)
	require.True(t, treasury.Equal(expectedTreasury))

	emitFinanceEvidence(t, payout, *reconciled, bridgeCompleted, ledger, treasury, expectedTreasury)
}
