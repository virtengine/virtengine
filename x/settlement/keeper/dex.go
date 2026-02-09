package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/pkg/dex"
	"github.com/virtengine/virtengine/pkg/payments/offramp"
	"github.com/virtengine/virtengine/x/settlement/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// DexSwapExecutor is the subset of DEX swap executor used for conversions.
type DexSwapExecutor interface {
	GetQuote(ctx context.Context, request dex.SwapRequest) (dex.SwapQuote, error)
	ExecuteSwap(ctx context.Context, quote dex.SwapQuote, signedTx []byte) (dex.SwapResult, error)
}

// OffRampBridge is the subset of offramp bridge used for conversions.
type OffRampBridge interface {
	GetQuote(ctx context.Context, req offramp.QuoteRequest) (offramp.Quote, error)
	InitiatePayout(ctx context.Context, quote offramp.Quote, cryptoTxRef string, destination string, metadata map[string]string) (offramp.PayoutResult, error)
	GetStatus(ctx context.Context, payoutID string) (offramp.PayoutResult, error)
	Cancel(ctx context.Context, payoutID string) error
}

// ComplianceKeeper provides compliance records for conversion checks.
type ComplianceKeeper interface {
	GetComplianceRecord(ctx sdk.Context, address string) (*veidtypes.ComplianceRecord, bool)
}

const complianceStatusUnknown = "UNKNOWN"

// ======================================================================
// Sequence management
// ======================================================================

func (k Keeper) getNextFiatConversionSequence(ctx sdk.Context) uint64 {
	return k.getNextSequence(ctx, types.FiatConversionSequenceKey())
}

func (k Keeper) incrementFiatConversionSequence(ctx sdk.Context) uint64 {
	seq := k.getNextFiatConversionSequence(ctx)
	k.setNextSequence(ctx, types.FiatConversionSequenceKey(), seq+1)
	return seq
}

// SetNextFiatConversionSequence sets the next fiat conversion sequence.
func (k Keeper) SetNextFiatConversionSequence(ctx sdk.Context, seq uint64) {
	k.setNextSequence(ctx, types.FiatConversionSequenceKey(), seq)
}

// ======================================================================
// Preference storage
// ======================================================================

// SetFiatPayoutPreference stores provider fiat payout preferences.
func (k Keeper) SetFiatPayoutPreference(ctx sdk.Context, pref types.FiatPayoutPreference) error {
	if err := pref.Validate(); err != nil {
		return err
	}

	if pref.DestinationHash == "" && pref.DestinationRef != "" {
		pref.DestinationHash = types.HashDestination(pref.DestinationRef)
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&pref)
	if err != nil {
		return err
	}
	store.Set(types.FiatPayoutPreferenceKey(pref.Provider), bz)
	return nil
}

// GetFiatPayoutPreference retrieves provider fiat payout preferences.
func (k Keeper) GetFiatPayoutPreference(ctx sdk.Context, provider string) (types.FiatPayoutPreference, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.FiatPayoutPreferenceKey(provider))
	if bz == nil {
		return types.FiatPayoutPreference{}, false
	}

	var pref types.FiatPayoutPreference
	if err := json.Unmarshal(bz, &pref); err != nil {
		return types.FiatPayoutPreference{}, false
	}
	return pref, true
}

// DeleteFiatPayoutPreference removes preferences for a provider.
func (k Keeper) DeleteFiatPayoutPreference(ctx sdk.Context, provider string) error {
	store := ctx.KVStore(k.skey)
	store.Delete(types.FiatPayoutPreferenceKey(provider))
	return nil
}

// WithFiatPayoutPreferences iterates over all preferences.
func (k Keeper) WithFiatPayoutPreferences(ctx sdk.Context, fn func(types.FiatPayoutPreference) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixFiatPayoutPreference)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var pref types.FiatPayoutPreference
		if err := json.Unmarshal(iter.Value(), &pref); err != nil {
			continue
		}
		if fn(pref) {
			break
		}
	}
}

// ======================================================================
// Conversion storage
// ======================================================================

// SetFiatConversion saves a fiat conversion record.
func (k Keeper) SetFiatConversion(ctx sdk.Context, conversion types.FiatConversionRecord) error {
	if err := conversion.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	existing, found := k.GetFiatConversion(ctx, conversion.ConversionID)
	if found && existing.State != conversion.State {
		k.updateFiatConversionState(ctx, conversion, existing.State)
	}

	if conversion.InvoiceID != "" {
		store.Set(types.FiatConversionByInvoiceKey(conversion.InvoiceID), []byte(conversion.ConversionID))
	}
	if conversion.SettlementID != "" {
		store.Set(types.FiatConversionBySettlementKey(conversion.SettlementID), []byte(conversion.ConversionID))
	}
	if conversion.PayoutID != "" {
		store.Set(types.FiatConversionByPayoutKey(conversion.PayoutID), []byte(conversion.ConversionID))
	}
	store.Set(types.FiatConversionByProviderKey(conversion.Provider, conversion.ConversionID), []byte{})
	store.Set(types.FiatConversionByStateKey(conversion.State, conversion.ConversionID), []byte{})

	bz, err := json.Marshal(&conversion)
	if err != nil {
		return err
	}
	store.Set(types.FiatConversionKey(conversion.ConversionID), bz)
	return nil
}

// GetFiatConversion retrieves a fiat conversion record.
func (k Keeper) GetFiatConversion(ctx sdk.Context, conversionID string) (types.FiatConversionRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.FiatConversionKey(conversionID))
	if bz == nil {
		return types.FiatConversionRecord{}, false
	}
	var conversion types.FiatConversionRecord
	if err := json.Unmarshal(bz, &conversion); err != nil {
		return types.FiatConversionRecord{}, false
	}
	return conversion, true
}

// GetFiatConversionByInvoice retrieves conversion by invoice.
func (k Keeper) GetFiatConversionByInvoice(ctx sdk.Context, invoiceID string) (types.FiatConversionRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.FiatConversionByInvoiceKey(invoiceID))
	if bz == nil {
		return types.FiatConversionRecord{}, false
	}
	return k.GetFiatConversion(ctx, string(bz))
}

// GetFiatConversionBySettlement retrieves conversion by settlement.
func (k Keeper) GetFiatConversionBySettlement(ctx sdk.Context, settlementID string) (types.FiatConversionRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.FiatConversionBySettlementKey(settlementID))
	if bz == nil {
		return types.FiatConversionRecord{}, false
	}
	return k.GetFiatConversion(ctx, string(bz))
}

// GetFiatConversionByPayout retrieves conversion by payout.
func (k Keeper) GetFiatConversionByPayout(ctx sdk.Context, payoutID string) (types.FiatConversionRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.FiatConversionByPayoutKey(payoutID))
	if bz == nil {
		return types.FiatConversionRecord{}, false
	}
	return k.GetFiatConversion(ctx, string(bz))
}

// WithFiatConversions iterates over conversions.
func (k Keeper) WithFiatConversions(ctx sdk.Context, fn func(types.FiatConversionRecord) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixFiatConversion)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var conversion types.FiatConversionRecord
		if err := json.Unmarshal(iter.Value(), &conversion); err != nil {
			continue
		}
		if fn(conversion) {
			break
		}
	}
}

// WithFiatConversionsByState iterates over conversions by state.
func (k Keeper) WithFiatConversionsByState(ctx sdk.Context, state types.FiatConversionState, fn func(types.FiatConversionRecord) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.FiatConversionByStatePrefixKey(state))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		statePrefix := types.FiatConversionByStatePrefixKey(state)
		if len(key) <= len(statePrefix) {
			continue
		}
		conversionID := string(key[len(statePrefix):])
		conversion, found := k.GetFiatConversion(ctx, conversionID)
		if !found {
			continue
		}
		if fn(conversion) {
			break
		}
	}
}

func (k Keeper) updateFiatConversionState(ctx sdk.Context, conversion types.FiatConversionRecord, oldState types.FiatConversionState) {
	store := ctx.KVStore(k.skey)
	store.Delete(types.FiatConversionByStateKey(oldState, conversion.ConversionID))
	store.Set(types.FiatConversionByStateKey(conversion.State, conversion.ConversionID), []byte{})
}

// ======================================================================
// Conversion execution
// ======================================================================

// RequestFiatConversion creates a conversion record after compliance checks.
func (k Keeper) RequestFiatConversion(ctx sdk.Context, request types.FiatConversionRequest) (*types.FiatConversionRecord, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	params := k.GetParams(ctx)
	if !params.FiatConversionEnabled {
		return nil, types.ErrFiatConversionNotAllowed.Wrap("fiat conversion disabled")
	}

	if request.InvoiceID != "" {
		if existing, found := k.GetFiatConversionByInvoice(ctx, request.InvoiceID); found {
			return &existing, nil
		}
	}
	if request.SettlementID != "" {
		if existing, found := k.GetFiatConversionBySettlement(ctx, request.SettlementID); found {
			return &existing, nil
		}
	}
	if request.PayoutID != "" {
		if existing, found := k.GetFiatConversionByPayout(ctx, request.PayoutID); found {
			return &existing, nil
		}
	}

	if err := k.ensureFiatConversionDependencies(); err != nil {
		return nil, err
	}

	complianceStatus, complianceRisk, err := k.validateFiatConversionCompliance(ctx, request)
	if err != nil {
		return nil, err
	}

	seq := k.incrementFiatConversionSequence(ctx)
	conversionID := generateIDWithTimestamp("conv", seq, ctx.BlockTime().Unix())
	conversion := types.NewFiatConversionRecord(conversionID, request, request.CryptoAmount, ctx.BlockTime())
	conversion.ComplianceStatus = complianceStatus
	conversion.ComplianceRiskScore = complianceRisk
	conversion.ComplianceCheckedAt = ctx.BlockTime().Unix()
	conversion.AddAuditEntry("conversion_requested", request.RequestedBy, "", map[string]string{
		"invoice_id":    request.InvoiceID,
		"settlement_id": request.SettlementID,
	}, ctx.BlockTime())

	if err := k.SetFiatConversion(ctx, *conversion); err != nil {
		return nil, err
	}

	_ = ctx.EventManager().EmitTypedEvent(&types.EventFiatConversionRequested{
		ConversionID:  conversionID,
		InvoiceID:     request.InvoiceID,
		SettlementID:  request.SettlementID,
		Provider:      request.Provider,
		FiatCurrency:  request.FiatCurrency,
		PaymentMethod: request.PaymentMethod,
		RequestedAt:   ctx.BlockTime().Unix(),
	})

	return conversion, nil
}

// ReconcileFiatConversion refreshes conversion status from the off-ramp.
func (k Keeper) ReconcileFiatConversion(ctx sdk.Context, conversionID string) (*types.FiatConversionRecord, error) {
	conversion, found := k.GetFiatConversion(ctx, conversionID)
	if !found {
		return nil, types.ErrFiatConversionNotFound.Wrapf("conversion %s not found", conversionID)
	}

	if conversion.OffRampID == "" || k.offRampBridge == nil {
		return &conversion, nil
	}

	status, err := k.offRampBridge.GetStatus(ctx, conversion.OffRampID)
	if err != nil {
		return &conversion, err
	}

	conversion.OffRampStatus = string(status.Status)
	conversion.OffRampReference = status.Reference
	conversion.AddAuditEntry("offramp_reconciled", "system", "", map[string]string{
		"status": string(status.Status),
	}, ctx.BlockTime())

	switch status.Status {
	case offramp.StatusCompleted:
		_ = conversion.MarkCompleted(ctx.BlockTime())
	case offramp.StatusFailed:
		_ = conversion.MarkFailed("off-ramp failed", ctx.BlockTime())
	}

	if err := k.SetFiatConversion(ctx, conversion); err != nil {
		return nil, err
	}

	if conversion.PayoutID != "" {
		payout, found := k.GetPayout(ctx, conversion.PayoutID)
		if found {
			if conversion.State == types.FiatConversionStateCompleted && payout.State != types.PayoutStateCompleted {
				oldState := payout.State
				_ = payout.MarkCompleted(fmt.Sprintf("fiat-%s", conversion.ConversionID), ctx.BlockTime())
				k.updatePayoutState(ctx, payout, oldState)
				_ = k.SetPayout(ctx, payout)
			} else if conversion.State == types.FiatConversionStateFailed && payout.State != types.PayoutStateFailed {
				oldState := payout.State
				_ = payout.MarkFailed("fiat conversion failed", ctx.BlockTime())
				k.updatePayoutState(ctx, payout, oldState)
				_ = k.SetPayout(ctx, payout)
			}
		}
	}

	_ = ctx.EventManager().EmitTypedEvent(&types.EventFiatConversionReconciled{
		ConversionID: conversion.ConversionID,
		Provider:     conversion.Provider,
		State:        string(conversion.State),
		ReconciledAt: ctx.BlockTime().Unix(),
	})

	return &conversion, nil
}

// createConversionFromPreference builds a conversion request from preferences.
func (k Keeper) createConversionFromPreference(ctx sdk.Context, settlement types.SettlementRecord, invoiceID string, pref types.FiatPayoutPreference) (*types.FiatConversionRecord, error) {
	if !pref.Enabled {
		return nil, nil
	}

	netAmount, err := k.calculateNetPayoutAmount(ctx, settlement)
	if err != nil {
		return nil, err
	}
	if netAmount.Denom != pref.CryptoToken.Denom {
		return nil, types.ErrInvalidAmount.Wrap("payout denom does not match preference crypto token")
	}

	request := types.FiatConversionRequest{
		InvoiceID:         invoiceID,
		SettlementID:      settlement.SettlementID,
		Provider:          settlement.Provider,
		Customer:          settlement.Customer,
		RequestedBy:       settlement.Provider,
		CryptoAmount:      netAmount,
		FiatCurrency:      pref.FiatCurrency,
		PaymentMethod:     pref.PaymentMethod,
		Destination:       pref.DestinationRef,
		DestinationRegion: pref.DestinationRegion,
		PreferredDEX:      pref.PreferredDEX,
		PreferredOffRamp:  pref.PreferredOffRamp,
		SlippageTolerance: pref.SlippageTolerance,
		CryptoToken:       pref.CryptoToken,
		StableToken:       pref.StableToken,
	}

	conversion, err := k.RequestFiatConversion(ctx, request)
	if err != nil {
		return nil, err
	}
	conversion.OrderID = settlement.OrderID
	conversion.EscrowID = settlement.EscrowID
	conversion.LeaseID = settlement.LeaseID
	if err := k.SetFiatConversion(ctx, *conversion); err != nil {
		return nil, err
	}
	return conversion, nil
}

func (k Keeper) ensureFiatConversionDependencies() error {
	if k.dexSwap == nil {
		return types.ErrDexUnavailable.Wrap("dex swap executor missing")
	}
	if k.offRampBridge == nil {
		return types.ErrOffRampUnavailable.Wrap("off-ramp bridge missing")
	}
	return nil
}

func (k Keeper) validateFiatConversionCompliance(ctx sdk.Context, request types.FiatConversionRequest) (string, int32, error) {
	if k.complianceKeeper == nil {
		return complianceStatusUnknown, 0, types.ErrComplianceRequired.Wrap("compliance keeper not configured")
	}

	record, found := k.complianceKeeper.GetComplianceRecord(ctx, request.Provider)
	if !found || record == nil {
		return complianceStatusUnknown, 0, types.ErrComplianceRequired.Wrap("compliance record missing")
	}

	if record.IsExpired(ctx.BlockTime().Unix()) {
		return record.Status.String(), record.RiskScore, types.ErrComplianceRequired.Wrap("compliance record expired")
	}

	params := k.GetParams(ctx)
	if record.RiskScore > params.FiatConversionRiskScoreThreshold {
		return record.Status.String(), record.RiskScore, types.ErrComplianceRequired.Wrap("risk score exceeds threshold")
	}

	if request.DestinationRegion != "" {
		for _, region := range record.RestrictedRegions {
			if region == request.DestinationRegion {
				return record.Status.String(), record.RiskScore, types.ErrComplianceRequired.Wrap("destination region restricted")
			}
		}
	}

	minStatus := strings.ToUpper(params.FiatConversionMinComplianceStatus)
	requiredStatus := veidtypes.ComplianceStatusCleared
	switch minStatus {
	case complianceStatusUnknown:
		requiredStatus = veidtypes.ComplianceStatusUnknown
	case "PENDING":
		requiredStatus = veidtypes.ComplianceStatusPending
	case "CLEARED":
		requiredStatus = veidtypes.ComplianceStatusCleared
	case "FLAGGED":
		requiredStatus = veidtypes.ComplianceStatusFlagged
	case "BLOCKED":
		requiredStatus = veidtypes.ComplianceStatusBlocked
	case "EXPIRED":
		requiredStatus = veidtypes.ComplianceStatusExpired
	}

	if record.Status != requiredStatus {
		return record.Status.String(), record.RiskScore, types.ErrComplianceRequired.Wrap("compliance status not sufficient")
	}

	return record.Status.String(), record.RiskScore, nil
}

func (k Keeper) calculateNetPayoutAmount(ctx sdk.Context, settlement types.SettlementRecord) (sdk.Coin, error) {
	holdback := k.calculateHoldbackAmount(ctx, settlement.TotalAmount)
	netAmount := settlement.ProviderShare.Sub(holdback...)
	if len(netAmount) != 1 {
		return sdk.Coin{}, types.ErrInvalidAmount.Wrap("net payout must be single denom for fiat conversion")
	}
	return netAmount[0], nil
}

func (k Keeper) validateConversionLimits(ctx sdk.Context, provider string, amount sdkmath.Int) error {
	params := k.GetParams(ctx)

	minAmount, _ := sdkmath.NewIntFromString(params.FiatConversionMinAmount)
	maxAmount, _ := sdkmath.NewIntFromString(params.FiatConversionMaxAmount)
	dailyLimit, _ := sdkmath.NewIntFromString(params.FiatConversionDailyLimit)

	if minAmount.IsPositive() && amount.LT(minAmount) {
		return types.ErrFiatLimitExceeded.Wrap("amount below minimum")
	}
	if maxAmount.IsPositive() && amount.GT(maxAmount) {
		return types.ErrFiatLimitExceeded.Wrap("amount exceeds maximum")
	}

	if dailyLimit.IsPositive() {
		day := ctx.BlockTime().UTC().Format("20060102")
		total := k.getFiatDailyTotal(ctx, provider, day)
		if total.Add(amount).GT(dailyLimit) {
			return types.ErrFiatLimitExceeded.Wrap("daily limit exceeded")
		}
	}
	return nil
}

func (k Keeper) getFiatDailyTotal(ctx sdk.Context, provider string, day string) sdkmath.Int {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.FiatDailyTotalKey(provider, day))
	if bz == nil {
		return sdkmath.ZeroInt()
	}
	total, ok := sdkmath.NewIntFromString(string(bz))
	if !ok {
		return sdkmath.ZeroInt()
	}
	return total
}

func (k Keeper) setFiatDailyTotal(ctx sdk.Context, provider string, day string, total sdkmath.Int) {
	store := ctx.KVStore(k.skey)
	store.Set(types.FiatDailyTotalKey(provider, day), []byte(total.String()))
}

func tokenSpecToDexToken(spec types.TokenSpec) dex.Token {
	return dex.Token{
		Symbol:   spec.Symbol,
		Denom:    spec.Denom,
		Decimals: spec.Decimals,
		ChainID:  spec.ChainID,
	}
}

func swapQuoteOutputAmount(quote dex.SwapQuote) (sdkmath.Int, error) {
	if len(quote.Route.Hops) == 0 {
		return sdkmath.ZeroInt(), fmt.Errorf("swap quote missing route")
	}
	lastHop := quote.Route.Hops[len(quote.Route.Hops)-1]
	return lastHop.AmountOut, nil
}

func (k Keeper) executeFiatConversion(ctx sdk.Context, payout *types.PayoutRecord, conversion *types.FiatConversionRecord) error {
	if err := k.ensureFiatConversionDependencies(); err != nil {
		return err
	}

	oldState := payout.State
	if err := payout.MarkProcessing(ctx.BlockTime()); err != nil {
		return err
	}
	k.updatePayoutState(ctx, *payout, oldState)
	if err := k.SetPayout(ctx, *payout); err != nil {
		return err
	}
	k.savePayoutLedgerEntry(ctx, payout.PayoutID, types.PayoutLedgerEntryProcessing,
		oldState, types.PayoutStateProcessing, sdk.NewCoins(), "fiat conversion processing", "system")

	if err := conversion.MarkSwapping(ctx.BlockTime()); err != nil {
		return err
	}
	conversion.AddAuditEntry("swap_requested", "system", "", nil, ctx.BlockTime())
	if err := k.SetFiatConversion(ctx, *conversion); err != nil {
		return err
	}

	params := k.GetParams(ctx)
	slippage := conversion.SlippageTolerance
	if slippage <= 0 {
		slippage = 0.01
	}
	if params.FiatConversionMaxSlippage != "" {
		if maxSlippage, err := sdkmath.LegacyNewDecFromStr(params.FiatConversionMaxSlippage); err == nil {
			if slippageDec, err := sdkmath.LegacyNewDecFromStr(fmt.Sprintf("%f", slippage)); err == nil {
				if slippageDec.GT(maxSlippage) {
					slippage, _ = maxSlippage.Float64()
				}
			}
		}
	}

	swapReq := dex.SwapRequest{
		FromToken:         tokenSpecToDexToken(conversion.CryptoToken),
		ToToken:           tokenSpecToDexToken(conversion.StableToken),
		Amount:            conversion.CryptoAmount.Amount,
		Type:              dex.SwapTypeExactIn,
		SlippageTolerance: slippage,
		Deadline:          ctx.BlockTime().Add(2 * time.Minute),
		Sender:            payout.Provider,
		PreferredDEX:      conversion.DexAdapter,
	}

	swapQuote, err := k.dexSwap.GetQuote(ctx, swapReq)
	if err != nil {
		return types.ErrFiatConversionFailed.Wrapf("swap quote failed: %s", err)
	}
	if !swapQuote.ExpiresAt.IsZero() && ctx.BlockTime().After(swapQuote.ExpiresAt) {
		return types.ErrFiatConversionFailed.Wrap("swap quote expired")
	}
	expectedOut, err := swapQuoteOutputAmount(swapQuote)
	if err != nil {
		return types.ErrFiatConversionFailed.Wrap(err.Error())
	}
	if !expectedOut.IsPositive() {
		return types.ErrFiatConversionFailed.Wrap("swap quote output must be positive")
	}
	if err := k.validateConversionLimits(ctx, conversion.Provider, expectedOut); err != nil {
		return err
	}

	swapResult, err := k.dexSwap.ExecuteSwap(ctx, swapQuote, []byte("settlement"))
	if err != nil {
		return types.ErrFiatConversionFailed.Wrapf("swap execution failed: %s", err)
	}

	conversion.SwapQuoteID = swapResult.QuoteID
	conversion.SwapTxHash = swapResult.TxHash
	conversion.SwapStatus = "executed"
	conversion.StableAmount = sdk.NewCoin(conversion.StableToken.Denom, swapResult.OutputAmount)
	conversion.AddAuditEntry("swap_executed", "system", "", map[string]string{
		"tx_hash": swapResult.TxHash,
	}, ctx.BlockTime())

	if err := conversion.MarkOffRampPending(ctx.BlockTime()); err != nil {
		return err
	}
	if err := k.SetFiatConversion(ctx, *conversion); err != nil {
		return err
	}

	offQuote, err := k.offRampBridge.GetQuote(ctx, offramp.QuoteRequest{
		CryptoSymbol:  conversion.StableToken.Symbol,
		CryptoDenom:   conversion.StableToken.Denom,
		CryptoAmount:  conversion.StableAmount.Amount,
		FiatCurrency:  conversion.FiatCurrency,
		PaymentMethod: conversion.PaymentMethod,
		Sender:        payout.Provider,
		Destination:   conversion.DestinationRef,
	})
	if err != nil {
		return types.ErrFiatConversionFailed.Wrapf("off-ramp quote failed: %s", err)
	}

	offResult, err := k.offRampBridge.InitiatePayout(ctx, offQuote, swapResult.TxHash, conversion.DestinationRef, map[string]string{
		"conversion_id": conversion.ConversionID,
	})
	if err != nil {
		return types.ErrFiatConversionFailed.Wrapf("off-ramp initiation failed: %s", err)
	}

	conversion.OffRampProvider = offResult.Provider
	conversion.OffRampQuoteID = offResult.QuoteID
	conversion.OffRampID = offResult.ID
	conversion.OffRampStatus = string(offResult.Status)
	conversion.OffRampReference = offResult.Reference
	conversion.FiatAmount = offResult.FiatAmount.String()
	conversion.AddAuditEntry("offramp_initiated", "system", "", map[string]string{
		"offramp_id": offResult.ID,
	}, ctx.BlockTime())

	switch offResult.Status {
	case offramp.StatusCompleted:
		_ = conversion.MarkCompleted(ctx.BlockTime())
		_ = payout.MarkCompleted(fmt.Sprintf("fiat-%s", conversion.ConversionID), ctx.BlockTime())
		k.updatePayoutState(ctx, *payout, types.PayoutStateProcessing)
		if err := k.SetPayout(ctx, *payout); err != nil {
			return err
		}
		k.savePayoutLedgerEntry(ctx, payout.PayoutID, types.PayoutLedgerEntryCompleted,
			types.PayoutStateProcessing, types.PayoutStateCompleted, payout.NetAmount, "fiat conversion completed", "system")

		k.recordTreasuryEntry(ctx, payout, types.TreasuryRecordPlatformFee, payout.PlatformFee)
		k.recordTreasuryEntry(ctx, payout, types.TreasuryRecordValidatorFee, payout.ValidatorFee)
		if !payout.HoldbackAmount.IsZero() {
			k.recordTreasuryEntry(ctx, payout, types.TreasuryRecordHoldback, payout.HoldbackAmount)
		}

		day := ctx.BlockTime().UTC().Format("20060102")
		total := k.getFiatDailyTotal(ctx, conversion.Provider, day)
		k.setFiatDailyTotal(ctx, conversion.Provider, day, total.Add(offResult.CryptoAmount))
		_ = ctx.EventManager().EmitTypedEvent(&types.EventFiatConversionCompleted{
			ConversionID: conversion.ConversionID,
			Provider:     conversion.Provider,
			FiatCurrency: conversion.FiatCurrency,
			FiatAmount:   conversion.FiatAmount,
			CompletedAt:  ctx.BlockTime().Unix(),
		})
	case offramp.StatusFailed:
		_ = conversion.MarkFailed("off-ramp failed", ctx.BlockTime())
		_ = payout.MarkFailed("fiat conversion failed", ctx.BlockTime())
		k.updatePayoutState(ctx, *payout, types.PayoutStateProcessing)
		if err := k.SetPayout(ctx, *payout); err != nil {
			return err
		}
		k.savePayoutLedgerEntry(ctx, payout.PayoutID, types.PayoutLedgerEntryFailed,
			types.PayoutStateProcessing, types.PayoutStateFailed, sdk.NewCoins(), "fiat conversion failed", "system")
		_ = ctx.EventManager().EmitTypedEvent(&types.EventFiatConversionFailed{
			ConversionID: conversion.ConversionID,
			Provider:     conversion.Provider,
			Reason:       "off-ramp failed",
			FailedAt:     ctx.BlockTime().Unix(),
		})
	}

	if err := k.SetFiatConversion(ctx, *conversion); err != nil {
		return err
	}

	return nil
}
