package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	FindPayoutByMetadata(ctx context.Context, provider string, metadata map[string]string) (offramp.PayoutResult, error)
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
	if pref.EncryptedPayload != nil {
		pref.EncryptedPayload.EnsureEnvelopeHash()
		if k.encryptionKeeper != nil && pref.EncryptedPayload.Envelope != nil {
			missing, err := k.encryptionKeeper.ValidateEnvelopeRecipients(ctx, pref.EncryptedPayload.Envelope)
			if err != nil {
				return types.ErrInvalidParams.Wrapf("recipient validation failed: %v", err)
			}
			if len(missing) > 0 {
				return types.ErrInvalidParams.Wrapf("unregistered recipients: %v", missing)
			}
		}
	}
	if err := pref.Validate(); err != nil {
		return err
	}

	if pref.EncryptedPayload != nil && pref.EncryptedPayload.EnvelopeRef != "" {
		pref.DestinationRef = pref.EncryptedPayload.EnvelopeRef
	}

	if pref.DestinationHash == "" && pref.DestinationRef != "" && pref.EncryptedPayload == nil {
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
	if conversion.EncryptedPayload != nil {
		conversion.EncryptedPayload.EnsureEnvelopeHash()
		if k.encryptionKeeper != nil && conversion.EncryptedPayload.Envelope != nil {
			missing, err := k.encryptionKeeper.ValidateEnvelopeRecipients(ctx, conversion.EncryptedPayload.Envelope)
			if err != nil {
				return types.ErrInvalidParams.Wrapf("recipient validation failed: %v", err)
			}
			if len(missing) > 0 {
				return types.ErrInvalidParams.Wrapf("unregistered recipients: %v", missing)
			}
		}
	}
	if conversion.EncryptedPayload != nil && conversion.EncryptedPayload.EnvelopeRef != "" {
		conversion.DestinationRef = conversion.EncryptedPayload.EnvelopeRef
	}
	if conversion.IdempotencyKey == "" {
		conversion.IdempotencyKey = conversion.DefaultIdempotencyKey()
	}
	if err := conversion.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	existing, found := k.GetFiatConversion(ctx, conversion.ConversionID)
	if found {
		if existing.State != conversion.State {
			k.updateFiatConversionState(ctx, conversion, existing.State)
		}
		if existing.InvoiceID != "" && existing.InvoiceID != conversion.InvoiceID {
			store.Delete(types.FiatConversionByInvoiceKey(existing.InvoiceID))
		}
		if existing.SettlementID != "" && existing.SettlementID != conversion.SettlementID {
			store.Delete(types.FiatConversionBySettlementKey(existing.SettlementID))
		}
		if existing.PayoutID != "" && existing.PayoutID != conversion.PayoutID {
			store.Delete(types.FiatConversionByPayoutKey(existing.PayoutID))
		}
		if existing.Provider != "" && existing.Provider != conversion.Provider {
			store.Delete(types.FiatConversionByProviderKey(existing.Provider, existing.ConversionID))
		}
		if existing.IdempotencyKey != "" && existing.IdempotencyKey != conversion.IdempotencyKey {
			store.Delete(types.FiatConversionIdempotencyKey(existing.IdempotencyKey))
		}
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
	if conversion.Provider != "" {
		store.Set(types.FiatConversionByProviderKey(conversion.Provider, conversion.ConversionID), []byte{})
	}
	store.Set(types.FiatConversionByStateKey(conversion.State, conversion.ConversionID), []byte{})
	store.Set(types.FiatConversionIdempotencyKey(conversion.IdempotencyKey), []byte(conversion.ConversionID))

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

func (k Keeper) getFiatConversionByIdempotencyKey(ctx sdk.Context, idempotencyKey string) (types.FiatConversionRecord, bool) {
	if idempotencyKey == "" {
		return types.FiatConversionRecord{}, false
	}

	store := ctx.KVStore(k.skey)
	bz := store.Get(types.FiatConversionIdempotencyKey(idempotencyKey))
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
	if request.EncryptedPayload != nil {
		request.EncryptedPayload.EnsureEnvelopeHash()
		if k.encryptionKeeper != nil && request.EncryptedPayload.Envelope != nil {
			missing, err := k.encryptionKeeper.ValidateEnvelopeRecipients(ctx, request.EncryptedPayload.Envelope)
			if err != nil {
				return nil, types.ErrInvalidParams.Wrapf("recipient validation failed: %v", err)
			}
			if len(missing) > 0 {
				return nil, types.ErrInvalidParams.Wrapf("unregistered recipients: %v", missing)
			}
		}
	}

	params := k.GetParams(ctx)
	if !params.FiatConversionEnabled {
		return nil, types.ErrFiatConversionNotAllowed.Wrap("fiat conversion disabled")
	}

	idempotencyKey := fmt.Sprintf("fiatconv:%s:%s:%s:%s", request.InvoiceID, request.SettlementID, request.PayoutID, request.Provider)
	if existing, found := k.getFiatConversionByIdempotencyKey(ctx, idempotencyKey); found {
		return &existing, nil
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

	complianceStatus, complianceRisk, err := k.validateFiatConversionCompliance(ctx, request)
	if err != nil {
		return nil, err
	}

	seq := k.incrementFiatConversionSequence(ctx)
	conversionID := generateIDWithTimestamp("conv", seq, ctx.BlockTime().Unix())
	conversion := types.NewFiatConversionRecord(conversionID, request, request.CryptoAmount, ctx.BlockTime())
	conversion.IdempotencyKey = idempotencyKey
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

// ReconcileFiatConversion rejects direct provider reconciliation. An off-chain
// worker must submit an authenticated on-chain observation before state can
// advance; carrier version 0 has no such message yet.
func (k Keeper) ReconcileFiatConversion(ctx sdk.Context, conversionID string) (*types.FiatConversionRecord, error) {
	conversion, found := k.GetFiatConversion(ctx, conversionID)
	if !found {
		return nil, types.ErrFiatConversionNotFound.Wrapf("conversion %s not found", conversionID)
	}
	return &conversion, ensureNoConsensusExternalIO()
}

// ProcessInFlightFiatConversions intentionally leaves all non-terminal records
// pending. It performs no callback and reports no false completion.
func (k Keeper) ProcessInFlightFiatConversions(ctx sdk.Context) error {
	_ = ctx
	return nil
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
		DestinationHash:   pref.DestinationHash,
		DestinationRegion: pref.DestinationRegion,
		PreferredDEX:      pref.PreferredDEX,
		PreferredOffRamp:  pref.PreferredOffRamp,
		SlippageTolerance: pref.SlippageTolerance,
		CryptoToken:       pref.CryptoToken,
		StableToken:       pref.StableToken,
		EncryptedPayload:  pref.EncryptedPayload,
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

func ensureNoConsensusExternalIO() error {
	return types.ErrExternalIOForbidden.Wrap("submit an authenticated on-chain observation from an off-chain worker")
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

func (k Keeper) executeFiatConversion(_ sdk.Context, _ *types.PayoutRecord, _ *types.FiatConversionRecord) error {
	return ensureNoConsensusExternalIO()
}
