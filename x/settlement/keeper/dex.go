package keeper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/pkg/dex"
	"github.com/virtengine/virtengine/pkg/payments/offramp"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
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

// GetFiatConversionSequence returns the next persisted conversion sequence for genesis export.
func (k Keeper) GetFiatConversionSequence(ctx sdk.Context) uint64 {
	return k.getNextFiatConversionSequence(ctx)
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
	return k.setFiatConversion(ctx, conversion, false)
}

// ImportFiatConversion restores a genesis/migration record after GenesisState
// validation has established quarantine and payout-lineage invariants.
func (k Keeper) ImportFiatConversion(ctx sdk.Context, conversion types.FiatConversionRecord) error {
	return k.setFiatConversion(ctx, conversion, true)
}

func (k Keeper) setFiatConversion(ctx sdk.Context, conversion types.FiatConversionRecord, migration bool) error {
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
	for _, binding := range []struct {
		name, value string
		key         []byte
	}{
		{"invoice", conversion.InvoiceID, types.FiatConversionByInvoiceKey(conversion.InvoiceID)},
		{"settlement", conversion.SettlementID, types.FiatConversionBySettlementKey(conversion.SettlementID)},
		{"payout", conversion.PayoutID, types.FiatConversionByPayoutKey(conversion.PayoutID)},
		{"idempotency", conversion.IdempotencyKey, types.FiatConversionIdempotencyKey(conversion.IdempotencyKey)},
	} {
		if binding.value == "" {
			continue
		}
		if owner := store.Get(binding.key); owner != nil && string(owner) != conversion.ConversionID {
			return types.ErrFiatConversionIdempotencyConflict.Wrapf("%s index already belongs to conversion %s", binding.name, string(owner))
		}
	}

	existing, found := k.GetFiatConversion(ctx, conversion.ConversionID)
	if found {
		if !migration && existing.ProtocolVersion > 0 && conversion.ProtocolVersion > 0 &&
			(existing.Provider != conversion.Provider || existing.Customer != conversion.Customer ||
				existing.InvoiceID != conversion.InvoiceID || existing.SettlementID != conversion.SettlementID ||
				existing.IdempotencyKey != conversion.IdempotencyKey || !bytes.Equal(existing.RequestDigest, conversion.RequestDigest)) {
			return types.ErrFiatConversionIdempotencyConflict.Wrap("immutable conversion binding changed")
		}
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
	cacheCtx, write := ctx.CacheContext()
	conversion, err := k.requestFiatConversion(cacheCtx, request)
	if err != nil {
		return nil, err
	}
	write()
	return conversion, nil
}

func (k Keeper) requestFiatConversion(ctx sdk.Context, request types.FiatConversionRequest) (*types.FiatConversionRecord, error) {
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
	certified := settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_CERTIFIED_ENABLED
	if params.FiatConversionDEXProfileState != certified || params.FiatConversionPayoutProfileState != certified ||
		len(params.FiatConversionDEXProfileDigest) != 32 || len(params.FiatConversionPayoutProfileDigest) != 32 {
		return nil, types.ErrFiatProfileCommitment.Wrap("certified DEX and payout profile commitments required")
	}
	if request.RequestedBy != request.Provider {
		return nil, types.ErrUnauthorized.Wrap("fiat conversion must be requested by provider")
	}
	request.PreferredDEX = params.FiatConversionDEXProfileID
	request.PreferredOffRamp = params.FiatConversionPayoutProfileID
	if request.StableToken.Denom != params.FiatConversionStableDenom || request.StableToken.Symbol != params.FiatConversionStableSymbol || uint32(request.StableToken.Decimals) != params.FiatConversionStableDecimals {
		return nil, types.ErrInvalidAmount.Wrap("stable token differs from governed conversion token")
	}
	request.SlippageToleranceExact = request.CanonicalSlippageTolerance()
	var err error
	request, err = k.bindFiatConversionPayout(ctx, request)
	if err != nil {
		return nil, err
	}
	if err := validateFiatConversionLimits(params, request.CryptoAmount.Amount); err != nil {
		return nil, err
	}
	bucket := dailyFiatBucket(ctx.BlockTime())
	dailyKey := types.FiatDailyTotalKey(request.Provider, bucket)
	store := ctx.KVStore(k.skey)
	dailyTotal, err := decodeDailyTotal(store.Get(dailyKey))
	if err != nil {
		return nil, err
	}
	dailyLimit, _ := sdkmath.NewIntFromString(params.FiatConversionDailyLimit)
	if dailyLimit.IsPositive() && dailyTotal.Add(request.CryptoAmount.Amount).GT(dailyLimit) {
		return nil, types.ErrFiatLimitExceeded.Wrap("provider daily conversion limit exceeded")
	}
	if err := k.validateFiatConversionLineage(ctx, request); err != nil {
		return nil, err
	}
	for _, entry := range []struct{ kind, value string }{{"invoice", request.InvoiceID}, {"settlement", request.SettlementID}} {
		if caseID, held := k.HasActiveFinancialCase(ctx, entry.kind, entry.value); held {
			return nil, types.ErrDisputeActive.Wrapf("%s held by canonical case %s", entry.kind, caseID)
		}
	}

	complianceStatus, complianceRisk, complianceDigest, err := k.validateFiatConversionComplianceDecision(ctx, request)
	if err != nil {
		return nil, err
	}
	requestDigest, err := canonicalFiatRequestDigest(request, params, complianceDigest)
	if err != nil {
		return nil, err
	}

	idempotencyKey := fmt.Sprintf("fiatconv:%s:%s:%s:%s", request.InvoiceID, request.SettlementID, request.PayoutID, request.Provider)
	if existing, found := k.getFiatConversionByIdempotencyKey(ctx, idempotencyKey); found {
		if bytes.Equal(existing.RequestDigest, requestDigest) {
			return &existing, nil
		}
		return nil, types.ErrFiatConversionIdempotencyConflict
	}

	if request.InvoiceID != "" {
		if existing, found := k.GetFiatConversionByInvoice(ctx, request.InvoiceID); found {
			if bytes.Equal(existing.RequestDigest, requestDigest) {
				return &existing, nil
			}
			return nil, types.ErrFiatConversionIdempotencyConflict.Wrap("invoice already bound to different conversion payload")
		}
	}
	if request.SettlementID != "" {
		if existing, found := k.GetFiatConversionBySettlement(ctx, request.SettlementID); found {
			if bytes.Equal(existing.RequestDigest, requestDigest) {
				return &existing, nil
			}
			return nil, types.ErrFiatConversionIdempotencyConflict.Wrap("settlement already bound to different conversion payload")
		}
	}
	if request.PayoutID != "" {
		if existing, found := k.GetFiatConversionByPayout(ctx, request.PayoutID); found {
			if bytes.Equal(existing.RequestDigest, requestDigest) {
				return &existing, nil
			}
			return nil, types.ErrFiatConversionIdempotencyConflict.Wrap("payout already bound to different conversion payload")
		}
	}
	seq := k.incrementFiatConversionSequence(ctx)
	conversionID := generateIDWithTimestamp("conv", seq, ctx.BlockTime().Unix())
	conversion := types.NewFiatConversionRecord(conversionID, request, request.CryptoAmount, ctx.BlockTime())
	settlement, _ := k.GetSettlement(ctx, request.SettlementID)
	conversion.EscrowID = settlement.EscrowID
	conversion.OrderID = settlement.OrderID
	conversion.LeaseID = settlement.LeaseID
	conversion.IdempotencyKey = idempotencyKey
	conversion.ComplianceStatus = complianceStatus
	conversion.ComplianceRiskScore = complianceRisk
	conversion.ComplianceCheckedAt = ctx.BlockTime().Unix()
	conversion.ProtocolVersion = fiatConversionProtocolVersion
	conversion.DEXProfileID = params.FiatConversionDEXProfileID
	conversion.DEXProfileDigest = append([]byte(nil), params.FiatConversionDEXProfileDigest...)
	conversion.PayoutProfileID = params.FiatConversionPayoutProfileID
	conversion.PayoutProfileDigest = append([]byte(nil), params.FiatConversionPayoutProfileDigest...)
	conversion.ComplianceDecisionHash = append([]byte(nil), complianceDigest...)
	conversion.RequestDigest = append([]byte(nil), requestDigest...)
	conversion.DailyBucket = bucket
	conversion.DailyQuotaReserved = true
	conversion.DexAdapter = params.FiatConversionDEXProfileID
	conversion.OffRampProvider = params.FiatConversionPayoutProfileID
	conversion.AddAuditEntry("conversion_requested", request.RequestedBy, "", map[string]string{
		"invoice_id":    request.InvoiceID,
		"settlement_id": request.SettlementID,
	}, ctx.BlockTime())

	payout, found := k.GetPayout(ctx, request.PayoutID)
	if !found || payout.State != types.PayoutStatePending || payout.FiatConversionID != "" {
		return nil, types.ErrPayoutHeld.Wrap("pending payout was claimed before conversion commit")
	}
	payout.FiatConversionID = conversionID
	if err := k.SetPayout(ctx, payout); err != nil {
		return nil, err
	}
	if err := k.SetFiatConversion(ctx, *conversion); err != nil {
		return nil, err
	}
	store.Set(types.FiatConversionRequestDigestKey(idempotencyKey), requestDigest)
	store.Set(dailyKey, encodeDailyTotal(dailyTotal.Add(request.CryptoAmount.Amount)))

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

func (k Keeper) bindFiatConversionPayout(ctx sdk.Context, request types.FiatConversionRequest) (types.FiatConversionRequest, error) {
	if strings.TrimSpace(request.SettlementID) == "" {
		return request, types.ErrInvalidSettlement.Wrap("settlement_id required for fiat conversion")
	}
	if request.PayoutID != "" {
		return request, nil
	}
	if request.InvoiceID != "" {
		if payout, found := k.GetPayoutByInvoice(ctx, request.InvoiceID); found {
			request.PayoutID = payout.PayoutID
			return request, nil
		}
	}
	if payout, found := k.GetPayoutBySettlement(ctx, request.SettlementID); found {
		request.PayoutID = payout.PayoutID
		return request, nil
	}
	settlement, found := k.GetSettlement(ctx, request.SettlementID)
	if !found {
		return request, types.ErrSettlementNotFound.Wrapf("settlement %s not found", request.SettlementID)
	}
	holdback := k.calculateHoldbackAmount(ctx, settlement.TotalAmount)
	sequence := k.incrementPayoutSequence(ctx)
	payout := types.NewPayoutRecord(
		generateIDWithTimestamp("payout", sequence, ctx.BlockTime().Unix()), request.InvoiceID, settlement.SettlementID,
		settlement.EscrowID, settlement.OrderID, settlement.LeaseID, settlement.Provider, settlement.Customer,
		settlement.TotalAmount, settlement.PlatformFee, settlement.ValidatorFee, holdback, ctx.BlockTime(), ctx.BlockHeight(),
	)
	if err := k.SetPayout(ctx, *payout); err != nil {
		return request, err
	}
	k.savePayoutLedgerEntry(ctx, payout.PayoutID, types.PayoutLedgerEntryCreated, types.PayoutStatePending, types.PayoutStatePending, payout.NetAmount, "payout created and held for fiat conversion", "system")
	request.PayoutID = payout.PayoutID
	return request, nil
}

func (k Keeper) validateFiatConversionLineage(ctx sdk.Context, request types.FiatConversionRequest) error {
	if request.SettlementID == "" {
		return types.ErrInvalidSettlement.Wrap("settlement_id required for fiat conversion")
	}
	settlement, found := k.GetSettlement(ctx, request.SettlementID)
	if !found {
		return types.ErrSettlementNotFound.Wrapf("settlement %s not found", request.SettlementID)
	}
	if settlement.Provider != request.Provider || settlement.Customer != request.Customer {
		return types.ErrUnauthorized.Wrap("settlement provider/customer lineage mismatch")
	}
	if request.PayoutID == "" {
		return types.ErrPayoutNotFound.Wrap("payout_id required for fiat conversion")
	}
	payout, found := k.GetPayout(ctx, request.PayoutID)
	if !found {
		return types.ErrPayoutNotFound.Wrapf("payout %s not found", request.PayoutID)
	}
	if payout.State != types.PayoutStatePending {
		return types.ErrPayoutHeld.Wrap("payout is not an unclaimed pending value hold")
	}
	if payout.FiatConversionID != "" {
		claimed, found := k.GetFiatConversion(ctx, payout.FiatConversionID)
		if !found || claimed.PayoutID != payout.PayoutID || claimed.InvoiceID != request.InvoiceID || claimed.SettlementID != request.SettlementID || claimed.Provider != request.Provider || claimed.Customer != request.Customer {
			return types.ErrPayoutHeld.Wrap("payout is claimed by another fiat conversion")
		}
	}
	if payout.SettlementID != request.SettlementID || payout.InvoiceID != request.InvoiceID || payout.Provider != request.Provider || payout.Customer != request.Customer ||
		payout.EscrowID != settlement.EscrowID || payout.OrderID != settlement.OrderID || payout.LeaseID != settlement.LeaseID {
		return types.ErrInvalidPayout.Wrap("payout lineage mismatch")
	}
	if len(payout.NetAmount) != 1 || !payout.NetAmount[0].IsEqual(request.CryptoAmount) {
		return types.ErrInvalidAmount.Wrap("conversion amount must equal pending payout net amount")
	}
	return nil
}

func validateFiatConversionLimits(params types.Params, amount sdkmath.Int) error {
	minimum, ok := sdkmath.NewIntFromString(params.FiatConversionMinAmount)
	if !ok || amount.LT(minimum) {
		return types.ErrFiatLimitExceeded.Wrap("conversion amount below minimum")
	}
	maximum, ok := sdkmath.NewIntFromString(params.FiatConversionMaxAmount)
	if !ok || (maximum.IsPositive() && amount.GT(maximum)) {
		return types.ErrFiatLimitExceeded.Wrap("conversion amount above maximum")
	}
	return nil
}

func (k Keeper) validateFiatConversionComplianceDecision(ctx sdk.Context, request types.FiatConversionRequest) (string, int32, []byte, error) {
	status, risk, err := k.validateFiatConversionCompliance(ctx, request)
	if err != nil {
		return status, risk, nil, err
	}
	record, found := k.complianceKeeper.GetComplianceRecord(ctx, request.Provider)
	if !found || record == nil {
		return status, risk, nil, types.ErrComplianceRequired.Wrap("compliance record missing")
	}
	digest, err := complianceDecisionDigest(record)
	if err != nil {
		return status, risk, nil, types.ErrComplianceRequired.Wrap("compliance decision digest invalid")
	}
	return status, risk, digest, nil
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

	payout, found := k.GetPayoutBySettlement(ctx, settlement.SettlementID)
	if !found || payout.State != types.PayoutStatePending {
		return nil, types.ErrPayoutNotFound.Wrap("pending payout must be created before fiat conversion")
	}
	if len(payout.NetAmount) != 1 {
		return nil, types.ErrInvalidAmount.Wrap("net payout must be single denom for fiat conversion")
	}
	netAmount := payout.NetAmount[0]
	if netAmount.Denom != pref.CryptoToken.Denom {
		return nil, types.ErrInvalidAmount.Wrap("payout denom does not match preference crypto token")
	}

	request := types.FiatConversionRequest{
		InvoiceID:              invoiceID,
		SettlementID:           settlement.SettlementID,
		PayoutID:               payout.PayoutID,
		Provider:               settlement.Provider,
		Customer:               settlement.Customer,
		RequestedBy:            settlement.Provider,
		CryptoAmount:           netAmount,
		FiatCurrency:           pref.FiatCurrency,
		PaymentMethod:          pref.PaymentMethod,
		DestinationHash:        pref.DestinationHash,
		DestinationRegion:      pref.DestinationRegion,
		PreferredDEX:           pref.PreferredDEX,
		PreferredOffRamp:       pref.PreferredOffRamp,
		SlippageTolerance:      pref.SlippageTolerance,
		SlippageToleranceExact: pref.SlippageToleranceExact,
		CryptoToken:            pref.CryptoToken,
		StableToken:            pref.StableToken,
		EncryptedPayload:       pref.EncryptedPayload,
	}

	return k.RequestFiatConversion(ctx, request)
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
