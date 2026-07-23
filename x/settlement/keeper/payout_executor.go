package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/escrow/types/billing"
	"github.com/virtengine/virtengine/x/settlement/types"
)

// PayoutExecutor handles payout execution logic
type PayoutExecutor interface {
	// ExecutePayout executes a payout for a settlement/invoice
	ExecutePayout(ctx sdk.Context, settlementID string, invoiceID string) (*types.PayoutRecord, error)

	// ExecutePayoutByID executes a payout by its ID
	ExecutePayoutByID(ctx sdk.Context, payoutID string) error

	// GetPayout retrieves a payout record
	GetPayout(ctx sdk.Context, payoutID string) (types.PayoutRecord, bool)

	// GetPayoutByInvoice retrieves payout by invoice ID
	GetPayoutByInvoice(ctx sdk.Context, invoiceID string) (types.PayoutRecord, bool)

	// GetPayoutBySettlement retrieves payout by settlement ID
	GetPayoutBySettlement(ctx sdk.Context, settlementID string) (types.PayoutRecord, bool)

	// GetPayoutsByProvider retrieves payouts for a provider
	GetPayoutsByProvider(ctx sdk.Context, provider string) []types.PayoutRecord

	// GetPayoutsByState retrieves payouts in a specific state
	GetPayoutsByState(ctx sdk.Context, state types.PayoutState) []types.PayoutRecord

	// HoldPayout places a hold on a payout due to dispute
	HoldPayout(ctx sdk.Context, payoutID string, disputeID string, reason string) error

	// ReleasePayoutHold releases a hold on a payout
	ReleasePayoutHold(ctx sdk.Context, payoutID string) error

	// RefundPayout refunds a held payout to the customer
	RefundPayout(ctx sdk.Context, payoutID string, reason string) error

	// ProcessPendingPayouts processes all pending payouts
	ProcessPendingPayouts(ctx sdk.Context) error

	// RetryFailedPayouts retries failed payouts
	RetryFailedPayouts(ctx sdk.Context) error

	// SetPayout saves a payout record
	SetPayout(ctx sdk.Context, payout types.PayoutRecord) error

	// WithPayouts iterates over all payouts
	WithPayouts(ctx sdk.Context, fn func(types.PayoutRecord) bool)
}

// ============================================================================
// Payout Sequence Management
// ============================================================================

func (k Keeper) getNextPayoutSequence(ctx sdk.Context) uint64 {
	return k.getNextSequence(ctx, types.PayoutSequenceKey())
}

func (k Keeper) incrementPayoutSequence(ctx sdk.Context) uint64 {
	seq := k.getNextPayoutSequence(ctx)
	k.setNextSequence(ctx, types.PayoutSequenceKey(), seq+1)
	return seq
}

// SetNextPayoutSequence sets the next payout sequence
func (k Keeper) SetNextPayoutSequence(ctx sdk.Context, seq uint64) {
	k.setNextSequence(ctx, types.PayoutSequenceKey(), seq)
}

// GetPayoutSequence returns the next persisted payout sequence for genesis export.
func (k Keeper) GetPayoutSequence(ctx sdk.Context) uint64 { return k.getNextPayoutSequence(ctx) }

// ============================================================================
// Payout Storage
// ============================================================================

// SetPayout saves a payout record to the store
func (k Keeper) SetPayout(ctx sdk.Context, payout types.PayoutRecord) error {
	if err := payout.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	for _, binding := range []struct {
		name, value string
		key         []byte
	}{{"invoice", payout.InvoiceID, types.PayoutByInvoiceKey(payout.InvoiceID)}, {"settlement", payout.SettlementID, types.PayoutBySettlementKey(payout.SettlementID)}, {"idempotency", payout.IdempotencyKey, types.PayoutIdempotencyKey(payout.IdempotencyKey)}} {
		if binding.value == "" {
			continue
		}
		if owner := store.Get(binding.key); owner != nil && string(owner) != payout.PayoutID {
			return types.ErrInvalidPayout.Wrapf("%s index already belongs to payout %s", binding.name, string(owner))
		}
	}
	existing, found := k.GetPayout(ctx, payout.PayoutID)
	if found {
		if existing.State != payout.State {
			k.updatePayoutState(ctx, payout, existing.State)
		}
		if existing.InvoiceID != "" && existing.InvoiceID != payout.InvoiceID {
			store.Delete(types.PayoutByInvoiceKey(existing.InvoiceID))
		}
		if existing.SettlementID != "" && existing.SettlementID != payout.SettlementID {
			store.Delete(types.PayoutBySettlementKey(existing.SettlementID))
		}
		if existing.Provider != "" && existing.Provider != payout.Provider {
			store.Delete(types.PayoutByProviderKey(existing.Provider, payout.PayoutID))
		}
		if existing.IdempotencyKey != "" && existing.IdempotencyKey != payout.IdempotencyKey {
			store.Delete(types.PayoutIdempotencyKey(existing.IdempotencyKey))
		}
	}

	bz, err := json.Marshal(&payout)
	if err != nil {
		return err
	}

	// Store by payout ID
	store.Set(types.PayoutKey(payout.PayoutID), bz)

	// Store by invoice ID
	if payout.InvoiceID != "" {
		store.Set(types.PayoutByInvoiceKey(payout.InvoiceID), []byte(payout.PayoutID))
	}

	// Store by settlement ID
	if payout.SettlementID != "" {
		store.Set(types.PayoutBySettlementKey(payout.SettlementID), []byte(payout.PayoutID))
	}

	// Store by provider
	store.Set(types.PayoutByProviderKey(payout.Provider, payout.PayoutID), []byte{})

	// Store by state
	store.Set(types.PayoutByStateKey(payout.State, payout.PayoutID), []byte{})

	// Store idempotency key
	store.Set(types.PayoutIdempotencyKey(payout.IdempotencyKey), []byte(payout.PayoutID))

	return nil
}

// GetPayout retrieves a payout record by ID
func (k Keeper) GetPayout(ctx sdk.Context, payoutID string) (types.PayoutRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.PayoutKey(payoutID))
	if bz == nil {
		return types.PayoutRecord{}, false
	}

	var payout types.PayoutRecord
	if err := json.Unmarshal(bz, &payout); err != nil {
		return types.PayoutRecord{}, false
	}

	return payout, true
}

// GetPayoutByInvoice retrieves payout by invoice ID
func (k Keeper) GetPayoutByInvoice(ctx sdk.Context, invoiceID string) (types.PayoutRecord, bool) {
	store := ctx.KVStore(k.skey)
	payoutID := store.Get(types.PayoutByInvoiceKey(invoiceID))
	if payoutID == nil {
		return types.PayoutRecord{}, false
	}

	return k.GetPayout(ctx, string(payoutID))
}

// GetPayoutBySettlement retrieves payout by settlement ID
func (k Keeper) GetPayoutBySettlement(ctx sdk.Context, settlementID string) (types.PayoutRecord, bool) {
	store := ctx.KVStore(k.skey)
	payoutID := store.Get(types.PayoutBySettlementKey(settlementID))
	if payoutID == nil {
		return types.PayoutRecord{}, false
	}

	return k.GetPayout(ctx, string(payoutID))
}

// GetPayoutsByProvider retrieves payouts for a provider
func (k Keeper) GetPayoutsByProvider(ctx sdk.Context, provider string) []types.PayoutRecord {
	store := ctx.KVStore(k.skey)
	prefix := types.PayoutByProviderPrefixKey(provider)
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var payouts []types.PayoutRecord
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		payoutID := string(key[len(prefix):])
		if payout, found := k.GetPayout(ctx, payoutID); found {
			payouts = append(payouts, payout)
		}
	}

	return payouts
}

// GetPayoutsByState retrieves payouts in a specific state
func (k Keeper) GetPayoutsByState(ctx sdk.Context, state types.PayoutState) []types.PayoutRecord {
	store := ctx.KVStore(k.skey)
	prefix := types.PayoutByStatePrefixKey(state)
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var payouts []types.PayoutRecord
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		payoutID := string(key[len(prefix):])
		if payout, found := k.GetPayout(ctx, payoutID); found {
			payouts = append(payouts, payout)
		}
	}

	return payouts
}

// WithPayouts iterates over all payouts
func (k Keeper) WithPayouts(ctx sdk.Context, fn func(types.PayoutRecord) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixPayout)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var payout types.PayoutRecord
		if err := json.Unmarshal(iter.Value(), &payout); err != nil {
			continue
		}
		if fn(payout) {
			break
		}
	}
}

// WithPayoutsByState iterates over payouts filtered by state
func (k Keeper) WithPayoutsByState(ctx sdk.Context, state types.PayoutState, fn func(types.PayoutRecord) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PayoutByStatePrefixKey(state))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// The key contains the payout ID after the state prefix
		key := iter.Key()
		statePrefix := types.PayoutByStatePrefixKey(state)
		if len(key) <= len(statePrefix) {
			continue
		}
		payoutID := string(key[len(statePrefix):])
		payout, found := k.GetPayout(ctx, payoutID)
		if !found {
			continue
		}
		if fn(payout) {
			break
		}
	}
}

// updatePayoutState updates the state index for a payout
func (k Keeper) updatePayoutState(ctx sdk.Context, payout types.PayoutRecord, oldState types.PayoutState) {
	store := ctx.KVStore(k.skey)

	// Remove old state index
	store.Delete(types.PayoutByStateKey(oldState, payout.PayoutID))

	// Add new state index
	store.Set(types.PayoutByStateKey(payout.State, payout.PayoutID), []byte{})
}

// ============================================================================
// Payout Execution
// ============================================================================

// ExecutePayout executes a payout for a settlement/invoice
func (k Keeper) ExecutePayout(ctx sdk.Context, invoiceID string, settlementID string) (*types.PayoutRecord, error) {
	cacheCtx, write := ctx.CacheContext()
	payout, err := k.executePayout(cacheCtx, invoiceID, settlementID)
	if err != nil {
		return nil, err
	}
	write()
	return payout, nil
}

func (k Keeper) executePayout(ctx sdk.Context, invoiceID string, settlementID string) (*types.PayoutRecord, error) {
	if caseID, held := k.HasActiveFinancialCase(ctx, "invoice", invoiceID); held {
		return nil, types.ErrDisputeActive.Wrapf("invoice held by canonical case %s", caseID)
	}
	// Check idempotency
	idempotencyKey := fmt.Sprintf("payout-%s-%s", invoiceID, settlementID)
	if existingPayoutID := k.checkPayoutIdempotency(ctx, idempotencyKey); existingPayoutID != "" {
		payout, found := k.GetPayout(ctx, existingPayoutID)
		if found {
			return &payout, nil // Already processed
		}
	}

	// Get settlement record
	settlement, found := k.GetSettlement(ctx, settlementID)
	if !found {
		return nil, types.ErrSettlementNotFound.Wrapf("settlement %s not found", settlementID)
	}

	// Check if escrow is in valid state for payout
	escrow, found := k.GetEscrow(ctx, settlement.EscrowID)
	if !found {
		return nil, types.ErrEscrowNotFound.Wrapf("escrow %s not found", settlement.EscrowID)
	}

	// Check for active disputes
	if escrow.State == types.EscrowStateDisputed {
		return nil, types.ErrDisputeActive.Wrap("escrow is under dispute")
	}

	// Calculate holdback (if any)
	holdbackAmount := k.calculateHoldbackAmount(ctx, settlement.TotalAmount)

	// Generate payout ID
	seq := k.incrementPayoutSequence(ctx)
	payoutID := generateIDWithTimestamp("payout", seq, ctx.BlockTime().Unix())

	// Create payout record
	payout := types.NewPayoutRecord(
		payoutID,
		invoiceID,
		settlementID,
		settlement.EscrowID,
		settlement.OrderID,
		settlement.LeaseID,
		settlement.Provider,
		settlement.Customer,
		settlement.TotalAmount,
		settlement.PlatformFee,
		settlement.ValidatorFee,
		holdbackAmount,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)

	// Save payout record
	if err := k.SetPayout(ctx, *payout); err != nil {
		return nil, err
	}

	// Create ledger entry
	k.savePayoutLedgerEntry(ctx, payout.PayoutID, types.PayoutLedgerEntryCreated,
		types.PayoutStatePending, types.PayoutStatePending,
		payout.NetAmount, "payout created", "system")

	// Check if fiat conversion is requested for this payout
	conversion, hasConversion := k.GetFiatConversionByInvoice(ctx, invoiceID)
	if !hasConversion {
		conversion, hasConversion = k.GetFiatConversionBySettlement(ctx, settlementID)
	}

	if hasConversion {
		if len(payout.NetAmount) != 1 {
			err := types.ErrInvalidAmount.Wrap("fiat conversion requires single denom payout")
			_ = payout.MarkFailed(err.Error(), ctx.BlockTime())
			_ = k.SetPayout(ctx, *payout)
			k.savePayoutLedgerEntry(ctx, payout.PayoutID, types.PayoutLedgerEntryFailed,
				types.PayoutStatePending, types.PayoutStateFailed,
				sdk.NewCoins(), fmt.Sprintf("fiat conversion failed: %s", err.Error()), "system")
			return payout, nil
		}

		conversion.PayoutID = payout.PayoutID
		conversion.EscrowID = payout.EscrowID
		conversion.OrderID = payout.OrderID
		conversion.LeaseID = payout.LeaseID
		conversion.CryptoAmount = payout.NetAmount[0]
		conversion.AddAuditEntry("payout_linked", "system", "", map[string]string{
			"payout_id": payout.PayoutID,
		}, ctx.BlockTime())

		if err := k.SetFiatConversion(ctx, conversion); err != nil {
			return nil, err
		}

		if payout.FiatConversionID != "" && payout.FiatConversionID != conversion.ConversionID {
			return nil, types.ErrInvalidPayout.Wrap("payout already claimed by another fiat conversion")
		}
		payout.FiatConversionID = conversion.ConversionID
		if err := k.SetPayout(ctx, *payout); err != nil {
			return nil, err
		}
		// External execution is deferred. An off-chain worker observes this
		// committed request and must submit an authenticated result in a future
		// schema version; consensus never calls the endpoint directly.
		return payout, nil
	}

	// An enabled fiat preference reserves this newly created payout as the
	// authoritative value hold. Settlement then creates a conversion that must
	// reference this exact payout before any external work can begin.
	if pref, ok := k.GetFiatPayoutPreference(ctx, payout.Provider); ok && pref.Enabled {
		return payout, nil
	}

	// Execute the payout immediately (crypto path)
	if err := k.executePayoutTransfer(ctx, payout); err != nil {
		// Mark as failed
		_ = payout.MarkFailed(err.Error(), ctx.BlockTime())
		_ = k.SetPayout(ctx, *payout)
		k.savePayoutLedgerEntry(ctx, payout.PayoutID, types.PayoutLedgerEntryFailed,
			types.PayoutStatePending, types.PayoutStateFailed,
			sdk.NewCoins(), fmt.Sprintf("payout failed: %s", err.Error()), "system")
		return payout, nil
	}

	return payout, nil
}

// executePayoutTransfer performs the actual fund transfer
func (k Keeper) executePayoutTransfer(ctx sdk.Context, payout *types.PayoutRecord) error {
	if payout == nil {
		return types.ErrInvalidPayout.Wrap("payout required")
	}
	if payout.FiatConversionID != "" {
		return types.ErrPayoutHeld.Wrap("fiat conversion owns payout value; legacy chain transfer forbidden")
	}
	if conversion, found := k.GetFiatConversionByPayout(ctx, payout.PayoutID); found {
		if conversion.State != types.FiatConversionStateCancelled || conversion.TerminalPolicy != terminalPolicyCancelNoSwap || conversion.DailyQuotaReserved {
			return types.ErrPayoutHeld.Wrapf("fiat conversion %s owns payout value", conversion.ConversionID)
		}
	}
	if caseID, held := k.HasActiveFinancialCase(ctx, "invoice", payout.InvoiceID); held {
		return types.ErrPayoutHeld.Wrapf("canonical case %s is active", caseID)
	}
	oldState := payout.State

	// Mark as processing
	if err := payout.MarkProcessing(ctx.BlockTime()); err != nil {
		return err
	}
	k.updatePayoutState(ctx, *payout, oldState)

	// Save processing state
	if err := k.SetPayout(ctx, *payout); err != nil {
		return err
	}

	k.savePayoutLedgerEntry(ctx, payout.PayoutID, types.PayoutLedgerEntryProcessing,
		oldState, types.PayoutStateProcessing,
		sdk.NewCoins(), "payout processing", "system")

	// Get provider address
	provider, err := sdk.AccAddressFromBech32(payout.Provider)
	if err != nil {
		return types.ErrInvalidAddress.Wrap("invalid provider address")
	}

	// Transfer net amount to provider
	if !payout.NetAmount.IsZero() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleAccountName,
			provider,
			payout.NetAmount,
		); err != nil {
			return types.ErrPayoutExecutionFailed.Wrap(err.Error())
		}
	}

	// Mark as completed
	txHash := fmt.Sprintf("payout-%s-%d", payout.PayoutID, ctx.BlockHeight())
	if err := payout.MarkCompleted(txHash, ctx.BlockTime()); err != nil {
		return err
	}
	effectHash := sha256.Sum256([]byte(strings.Join([]string{
		"virtengine/settlement/native-payout/v1", payout.PayoutID, payout.Provider, payout.NetAmount.String(), txHash,
	}, "\x00")))
	payout.ValueMovementApplied = true
	payout.ValueMovementEffectHash = effectHash[:]

	// Update state index
	k.updatePayoutState(ctx, *payout, types.PayoutStateProcessing)

	// Save completed state
	if err := k.SetPayout(ctx, *payout); err != nil {
		return err
	}

	k.savePayoutLedgerEntry(ctx, payout.PayoutID, types.PayoutLedgerEntryCompleted,
		types.PayoutStateProcessing, types.PayoutStateCompleted,
		payout.NetAmount, "payout completed", "system")

	// Record treasury entries for fees
	if err := k.recordPayoutRetainedTreasuryEntries(ctx, payout); err != nil {
		return err
	}

	// Emit event
	if err := ctx.EventManager().EmitTypedEvent(&types.EventPayoutCompleted{
		PayoutID:     payout.PayoutID,
		SettlementID: payout.SettlementID,
		InvoiceID:    payout.InvoiceID,
		Provider:     payout.Provider,
		NetAmount:    payout.NetAmount.String(),
		PlatformFee:  payout.PlatformFee.String(),
		CompletedAt:  ctx.BlockTime().Unix(),
	}); err != nil {
		k.Logger(ctx).Error("failed to emit payout completed event", "error", err)
	}

	k.Logger(ctx).Info("payout completed",
		"payout_id", payout.PayoutID,
		"provider", payout.Provider,
		"net_amount", payout.NetAmount.String(),
	)

	return nil
}

// ExecutePayoutByID executes a payout by its ID
func (k Keeper) ExecutePayoutByID(ctx sdk.Context, payoutID string) error {
	cacheCtx, write := ctx.CacheContext()
	if err := k.executePayoutByID(cacheCtx, payoutID); err != nil {
		return err
	}
	write()
	return nil
}

func (k Keeper) executePayoutByID(ctx sdk.Context, payoutID string) error {
	payout, found := k.GetPayout(ctx, payoutID)
	if !found {
		return types.ErrPayoutNotFound.Wrapf("payout %s not found", payoutID)
	}

	if payout.State.IsTerminal() {
		return nil // Already completed
	}

	if payout.State == types.PayoutStateHeld {
		return types.ErrPayoutHeld.Wrap("payout is on hold")
	}
	if caseID, held := k.HasActiveFinancialCase(ctx, "invoice", payout.InvoiceID); held {
		return types.ErrPayoutHeld.Wrapf("canonical case %s is active", caseID)
	}

	if payout.FiatConversionID != "" {
		conversion, found := k.GetFiatConversion(ctx, payout.FiatConversionID)
		if !found {
			return types.ErrFiatConversionNotFound.Wrapf("conversion %s not found for payout %s", payout.FiatConversionID, payout.PayoutID)
		}
		if conversion.State == types.FiatConversionStatePayoutCompleted {
			return nil
		}
		if conversion.State == types.FiatConversionStateCancelled && conversion.TerminalPolicy == terminalPolicyCancelNoSwap && !conversion.DailyQuotaReserved {
			payout.FiatConversionID = ""
			if err := k.SetPayout(ctx, payout); err != nil {
				return err
			}
			return k.executePayoutTransfer(ctx, &payout)
		}
		if conversion.State == types.FiatConversionStateFailed || conversion.State == types.FiatConversionStateCancelled {
			return types.ErrFiatConversionFailed.Wrap(conversion.TerminalPolicy)
		}
		return types.ErrPayoutHeld.Wrap("fiat payout remains held pending authenticated terminal observation")
	}

	return k.executePayoutTransfer(ctx, &payout)
}

// calculateHoldbackAmount computes holdback coins based on params.
func (k Keeper) calculateHoldbackAmount(ctx sdk.Context, gross sdk.Coins) sdk.Coins {
	holdbackAmount := sdk.NewCoins()
	params := k.GetParams(ctx)
	if params.PayoutHoldbackRate == "" {
		return holdbackAmount
	}

	holdbackRate, err := sdkmath.LegacyNewDecFromStr(params.PayoutHoldbackRate)
	if err != nil || !holdbackRate.IsPositive() {
		return holdbackAmount
	}

	for _, coin := range gross {
		holdbackCoin := sdk.NewCoin(coin.Denom, holdbackRate.MulInt(coin.Amount).TruncateInt())
		holdbackAmount = holdbackAmount.Add(holdbackCoin)
	}

	return holdbackAmount
}

// checkPayoutIdempotency checks if a payout has already been processed
func (k Keeper) checkPayoutIdempotency(ctx sdk.Context, idempotencyKey string) string {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.PayoutIdempotencyKey(idempotencyKey))
	if bz == nil {
		return ""
	}
	return string(bz)
}

// ============================================================================
// Dispute Integration
// ============================================================================

// HoldPayout places a hold on a payout due to dispute
func (k Keeper) HoldPayout(ctx sdk.Context, payoutID string, disputeID string, reason string) error {
	payout, found := k.GetPayout(ctx, payoutID)
	if !found {
		return types.ErrPayoutNotFound.Wrapf("payout %s not found", payoutID)
	}

	oldState := payout.State
	if payout.State == types.PayoutStateHeld {
		if payout.DisputeID == disputeID {
			return nil
		}
		return types.ErrFinancialCaseHold.Wrapf("payout already held by %s", payout.DisputeID)
	}
	if err := payout.Hold(disputeID, reason, ctx.BlockTime()); err != nil {
		return err
	}

	k.updatePayoutState(ctx, payout, oldState)

	if err := k.SetPayout(ctx, payout); err != nil {
		return err
	}

	k.savePayoutLedgerEntry(ctx, payout.PayoutID, types.PayoutLedgerEntryHeld,
		oldState, types.PayoutStateHeld,
		sdk.NewCoins(), fmt.Sprintf("payout held: %s", reason), "dispute")

	// Emit event
	if err := ctx.EventManager().EmitTypedEvent(&types.EventPayoutHeld{
		PayoutID:  payout.PayoutID,
		DisputeID: disputeID,
		Reason:    reason,
		HeldAt:    ctx.BlockTime().Unix(),
	}); err != nil {
		k.Logger(ctx).Error("failed to emit payout held event", "error", err)
	}

	k.Logger(ctx).Info("payout held",
		"payout_id", payout.PayoutID,
		"dispute_id", disputeID,
		"reason", reason,
	)

	return nil
}

// ReleasePayoutHold releases a hold on a payout
func (k Keeper) ReleasePayoutHold(ctx sdk.Context, payoutID string) error {
	payout, found := k.GetPayout(ctx, payoutID)
	if !found {
		return types.ErrPayoutNotFound.Wrapf("payout %s not found", payoutID)
	}

	if payout.State != types.PayoutStateHeld {
		return types.ErrInvalidStateTransition.Wrap("payout is not on hold")
	}
	if k.IsFinancialCasesActive(ctx) && strings.HasPrefix(payout.DisputeID, "financial-case/") {
		return types.ErrLegacyFinancialMutationFenced
	}

	oldState := payout.State
	if err := payout.ReleaseHold(); err != nil {
		return err
	}

	k.updatePayoutState(ctx, payout, oldState)

	if err := k.SetPayout(ctx, payout); err != nil {
		return err
	}

	k.savePayoutLedgerEntry(ctx, payout.PayoutID, types.PayoutLedgerEntryReleased,
		oldState, types.PayoutStatePending,
		sdk.NewCoins(), "payout hold released", "dispute_resolution")

	// Emit event
	if err := ctx.EventManager().EmitTypedEvent(&types.EventPayoutReleased{
		PayoutID:   payout.PayoutID,
		ReleasedAt: ctx.BlockTime().Unix(),
	}); err != nil {
		k.Logger(ctx).Error("failed to emit payout released event", "error", err)
	}

	k.Logger(ctx).Info("payout hold released",
		"payout_id", payout.PayoutID,
	)

	// Execute the payout now that hold is released
	return k.ExecutePayoutByID(ctx, payoutID)
}

func (k Keeper) cancelPayoutFiatConversion(ctx sdk.Context, payout types.PayoutRecord, reason string) error {
	if payout.FiatConversionID == "" {
		return nil
	}

	conversion, found := k.GetFiatConversion(ctx, payout.FiatConversionID)
	if !found || conversion.State.IsTerminal() {
		return nil
	}

	if err := conversion.MarkCancelled(reason, ctx.BlockTime()); err != nil {
		return err
	}
	conversion.AddAuditEntry("payout_cancelled", "system", reason, map[string]string{
		"payout_id": payout.PayoutID,
	}, ctx.BlockTime())

	return k.SetFiatConversion(ctx, conversion)
}

// RefundPayout refunds a held payout to the customer
func (k Keeper) RefundPayout(ctx sdk.Context, payoutID string, reason string) error {
	payout, found := k.GetPayout(ctx, payoutID)
	if !found {
		return types.ErrPayoutNotFound.Wrapf("payout %s not found", payoutID)
	}

	if payout.State != types.PayoutStateHeld {
		return types.ErrInvalidStateTransition.Wrap("can only refund held payouts")
	}
	if k.IsFinancialCasesActive(ctx) && strings.HasPrefix(payout.DisputeID, "financial-case/") {
		return types.ErrLegacyFinancialMutationFenced
	}

	oldState := payout.State

	if err := k.cancelPayoutFiatConversion(ctx, payout, reason); err != nil {
		return types.ErrPayoutExecutionFailed.Wrap(err.Error())
	}

	// Get customer address
	customer, err := sdk.AccAddressFromBech32(payout.Customer)
	if err != nil {
		return types.ErrInvalidAddress.Wrap("invalid customer address")
	}

	// Transfer back to customer
	if !payout.GrossAmount.IsZero() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleAccountName,
			customer,
			payout.GrossAmount,
		); err != nil {
			return types.ErrPayoutExecutionFailed.Wrap(err.Error())
		}
	}

	if err := payout.Refund(reason, ctx.BlockTime()); err != nil {
		return err
	}

	k.updatePayoutState(ctx, payout, oldState)

	if err := k.SetPayout(ctx, payout); err != nil {
		return err
	}

	k.savePayoutLedgerEntry(ctx, payout.PayoutID, types.PayoutLedgerEntryRefunded,
		oldState, types.PayoutStateRefunded,
		payout.GrossAmount, fmt.Sprintf("payout refunded: %s", reason), "dispute_resolution")

	// Emit event
	if err := ctx.EventManager().EmitTypedEvent(&types.EventPayoutRefunded{
		PayoutID:   payout.PayoutID,
		Customer:   payout.Customer,
		Amount:     payout.GrossAmount.String(),
		Reason:     reason,
		RefundedAt: ctx.BlockTime().Unix(),
	}); err != nil {
		k.Logger(ctx).Error("failed to emit payout refunded event", "error", err)
	}

	k.Logger(ctx).Info("payout refunded",
		"payout_id", payout.PayoutID,
		"customer", payout.Customer,
		"amount", payout.GrossAmount.String(),
	)

	return nil
}

// ============================================================================
// Batch Processing
// ============================================================================

// ProcessPendingPayouts processes all pending payouts
func (k Keeper) ProcessPendingPayouts(ctx sdk.Context) error {
	pendingPayouts := k.GetPayoutsByState(ctx, types.PayoutStatePending)

	for _, payout := range pendingPayouts {
		if payout.FiatConversionID != "" {
			continue
		}
		if _, held := k.HasActiveFinancialCase(ctx, "invoice", payout.InvoiceID); held {
			continue
		}
		if err := k.ExecutePayoutByID(ctx, payout.PayoutID); err != nil {
			k.Logger(ctx).Error("failed to process pending payout",
				"payout_id", payout.PayoutID,
				"error", err,
			)
		}
	}

	return nil
}

// RetryFailedPayouts retries failed payouts
func (k Keeper) RetryFailedPayouts(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	maxRetries := params.MaxPayoutRetries
	if maxRetries == 0 {
		maxRetries = 3 // Default
	}

	failedPayouts := k.GetPayoutsByState(ctx, types.PayoutStateFailed)

	for _, payout := range failedPayouts {
		if payout.FiatConversionID != "" {
			continue
		}
		if _, held := k.HasActiveFinancialCase(ctx, "invoice", payout.InvoiceID); held {
			continue
		}
		if payout.ExecutionAttempts >= maxRetries {
			continue // Max retries exceeded
		}

		// Reset to pending for retry and keep state index monotonic.
		payout.State = types.PayoutStatePending
		k.updatePayoutState(ctx, payout, types.PayoutStateFailed)
		if err := k.SetPayout(ctx, payout); err != nil {
			continue
		}

		if err := k.ExecutePayoutByID(ctx, payout.PayoutID); err != nil {
			k.Logger(ctx).Error("failed to retry payout",
				"payout_id", payout.PayoutID,
				"attempt", payout.ExecutionAttempts,
				"error", err,
			)
		}
	}

	return nil
}

// ============================================================================
// Ledger Entries
// ============================================================================

func (k Keeper) savePayoutLedgerEntry(
	ctx sdk.Context,
	payoutID string,
	entryType types.PayoutLedgerEntryType,
	prevState types.PayoutState,
	newState types.PayoutState,
	amount sdk.Coins,
	description string,
	initiator string,
) {
	store := ctx.KVStore(k.skey)

	entryID := fmt.Sprintf("%s-%d-%d", payoutID, ctx.BlockHeight(), ctx.BlockTime().UnixNano())
	entry := types.NewPayoutLedgerEntry(
		entryID,
		payoutID,
		entryType,
		prevState,
		newState,
		amount,
		description,
		initiator,
		"",
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)

	bz, err := json.Marshal(entry)
	if err != nil {
		return // silently skip if marshal fails
	}
	store.Set(types.PayoutLedgerEntryKey(entryID), bz)
	store.Set(types.PayoutLedgerByPayoutKey(payoutID, entryID), []byte(entryID))
}

// GetPayoutLedgerEntries retrieves ledger entries for a payout
func (k Keeper) GetPayoutLedgerEntries(ctx sdk.Context, payoutID string) []types.PayoutLedgerEntry {
	store := ctx.KVStore(k.skey)
	prefix := append([]byte(nil), types.PrefixPayoutLedgerByPayout...)
	prefix = append(prefix, []byte(payoutID)...)
	prefix = append(prefix, byte('/'))
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var entries []types.PayoutLedgerEntry
	for ; iter.Valid(); iter.Next() {
		entryID := string(iter.Value())
		bz := store.Get(types.PayoutLedgerEntryKey(entryID))
		if bz == nil {
			continue
		}

		var entry types.PayoutLedgerEntry
		if err := json.Unmarshal(bz, &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	return entries
}

// ============================================================================
// Treasury Accounting
// ============================================================================

func (k Keeper) recordTreasuryEntry(
	ctx sdk.Context,
	payout *types.PayoutRecord,
	recordType types.TreasuryRecordType,
	amount sdk.Coins,
) error {
	if amount.IsZero() {
		return nil
	}

	store := ctx.KVStore(k.skey)
	recordID := treasuryPayoutRecordID(payout.PayoutID, recordType)
	key := types.TreasuryRecordKey(recordID)
	if existing := store.Get(key); existing != nil {
		var record types.TreasuryRecord
		if err := json.Unmarshal(existing, &record); err != nil || record.PayoutID != payout.PayoutID || record.SettlementID != payout.SettlementID || record.RecordType != recordType || !record.Amount.Equal(amount) {
			return types.ErrInvalidSettlement.Wrapf("treasury effect conflict for payout %s type %s", payout.PayoutID, recordType.String())
		}
		return nil
	}

	// Get current treasury balance
	balance := k.getTreasuryBalance(ctx)

	// Update balance based on record type
	var balanceAfter sdk.Coins
	switch recordType {
	case types.TreasuryRecordPlatformFee, types.TreasuryRecordValidatorFee, types.TreasuryRecordHoldback:
		balanceAfter = balance.Add(amount...)
	case types.TreasuryRecordRefund, types.TreasuryRecordWithdrawal:
		if !balance.IsAllGTE(amount) {
			return types.ErrInvalidSettlement.Wrapf("treasury balance underflow for payout %s type %s", payout.PayoutID, recordType.String())
		}
		balanceAfter = balance.Sub(amount...)
	default:
		return types.ErrInvalidSettlement.Wrapf("unknown treasury record type %d", recordType)
	}

	record := types.TreasuryRecord{
		RecordID:     recordID,
		RecordType:   recordType,
		PayoutID:     payout.PayoutID,
		SettlementID: payout.SettlementID,
		Amount:       amount,
		BalanceAfter: balanceAfter,
		Description:  fmt.Sprintf("%s for payout %s", recordType.String(), payout.PayoutID),
		BlockHeight:  ctx.BlockHeight(),
		Timestamp:    ctx.BlockTime(),
	}

	bz, err := json.Marshal(&record)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	// Update treasury balance
	return k.setTreasuryBalance(ctx, balanceAfter)
}

func treasuryPayoutRecordID(payoutID string, recordType types.TreasuryRecordType) string {
	return "payout/" + payoutID + "/" + recordType.String()
}

func (k Keeper) validateRetainedTreasuryEntries(ctx sdk.Context, payout *types.PayoutRecord) error {
	if payout == nil {
		return types.ErrInvalidPayout.Wrap("payout required for treasury accounting")
	}
	for _, entry := range []struct {
		recordType types.TreasuryRecordType
		amount     sdk.Coins
	}{
		{types.TreasuryRecordPlatformFee, payout.PlatformFee},
		{types.TreasuryRecordValidatorFee, payout.ValidatorFee},
		{types.TreasuryRecordHoldback, payout.HoldbackAmount},
	} {
		if entry.amount.IsZero() {
			continue
		}
		if existing := ctx.KVStore(k.skey).Get(types.TreasuryRecordKey(treasuryPayoutRecordID(payout.PayoutID, entry.recordType))); existing != nil {
			var record types.TreasuryRecord
			if err := json.Unmarshal(existing, &record); err != nil || record.PayoutID != payout.PayoutID || record.RecordType != entry.recordType || !record.Amount.Equal(entry.amount) {
				return types.ErrInvalidSettlement.Wrapf("treasury effect conflict for payout %s type %s", payout.PayoutID, entry.recordType.String())
			}
		}
	}
	return nil
}

func (k Keeper) recordPayoutRetainedTreasuryEntries(ctx sdk.Context, payout *types.PayoutRecord) error {
	if err := k.validateRetainedTreasuryEntries(ctx, payout); err != nil {
		return err
	}
	for _, entry := range []struct {
		recordType types.TreasuryRecordType
		amount     sdk.Coins
	}{
		{types.TreasuryRecordPlatformFee, payout.PlatformFee},
		{types.TreasuryRecordValidatorFee, payout.ValidatorFee},
		{types.TreasuryRecordHoldback, payout.HoldbackAmount},
	} {
		if err := k.recordTreasuryEntry(ctx, payout, entry.recordType, entry.amount); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) loadTreasuryBalance(ctx sdk.Context) (sdk.Coins, error) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.PrefixTreasuryBalance)
	if bz == nil {
		return sdk.NewCoins(), nil
	}

	var balance sdk.Coins
	if err := json.Unmarshal(bz, &balance); err != nil {
		return nil, types.ErrInvalidSettlement.Wrap("malformed treasury balance")
	}
	if !balance.IsValid() {
		return nil, types.ErrInvalidSettlement.Wrap("invalid treasury balance")
	}
	return balance, nil
}

func (k Keeper) getTreasuryBalance(ctx sdk.Context) sdk.Coins {
	balance, err := k.loadTreasuryBalance(ctx)
	if err != nil {
		panic(err)
	}
	return balance
}

func (k Keeper) setTreasuryBalance(ctx sdk.Context, balance sdk.Coins) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&balance)
	if err != nil {
		return err
	}
	store.Set(types.PrefixTreasuryBalance, bz)
	return nil
}

// GetTreasuryBalance returns the current treasury balance
func (k Keeper) GetTreasuryBalance(ctx sdk.Context) sdk.Coins {
	return k.getTreasuryBalance(ctx)
}

func (k Keeper) ImportTreasuryAccounting(ctx sdk.Context, records []types.TreasuryRecord, balance sdk.Coins) error {
	if !balance.IsValid() {
		return types.ErrInvalidSettlement.Wrap("invalid treasury genesis balance")
	}
	cacheCtx, write := ctx.CacheContext()
	credits := sdk.NewCoins()
	debits := sdk.NewCoins()
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		if record.RecordID == "" || record.PayoutID == "" || record.SettlementID == "" || !record.Amount.IsValid() || record.Amount.IsZero() || !record.BalanceAfter.IsValid() {
			return types.ErrInvalidSettlement.Wrap("invalid treasury genesis record")
		}
		if _, duplicate := seen[record.RecordID]; duplicate {
			return types.ErrInvalidSettlement.Wrap("duplicate treasury genesis record")
		}
		seen[record.RecordID] = struct{}{}
		switch record.RecordType {
		case types.TreasuryRecordPlatformFee, types.TreasuryRecordValidatorFee, types.TreasuryRecordHoldback:
			credits = credits.Add(record.Amount...)
		case types.TreasuryRecordRefund, types.TreasuryRecordWithdrawal:
			debits = debits.Add(record.Amount...)
		default:
			return types.ErrInvalidSettlement.Wrap("unknown treasury genesis record type")
		}
		payout, found := k.GetPayout(cacheCtx, record.PayoutID)
		if !found || payout.SettlementID != record.SettlementID {
			return types.ErrInvalidSettlement.Wrapf("treasury genesis record %s has invalid payout linkage", record.RecordID)
		}
		key := types.TreasuryRecordKey(record.RecordID)
		if cacheCtx.KVStore(k.skey).Has(key) {
			return types.ErrInvalidSettlement.Wrapf("treasury genesis record %s conflicts with stored accounting", record.RecordID)
		}
		raw, err := json.Marshal(record)
		if err != nil {
			return err
		}
		cacheCtx.KVStore(k.skey).Set(key, raw)
	}
	if !credits.IsAllGTE(debits) {
		return types.ErrInvalidSettlement.Wrap("treasury genesis accounting underflow")
	}
	expected := credits.Sub(debits...)
	if !expected.Equal(balance) {
		return types.ErrInvalidSettlement.Wrapf("treasury genesis balance mismatch: expected %s got %s", expected, balance)
	}
	if err := k.setTreasuryBalance(cacheCtx, balance); err != nil {
		return err
	}
	write()
	return nil
}

func (k Keeper) ExportTreasuryAccounting(ctx sdk.Context) ([]types.TreasuryRecord, sdk.Coins, error) {
	records := make([]types.TreasuryRecord, 0)
	iterator := storetypes.KVStorePrefixIterator(ctx.KVStore(k.skey), types.PrefixTreasuryRecord)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var record types.TreasuryRecord
		if err := json.Unmarshal(iterator.Value(), &record); err != nil {
			return nil, nil, types.ErrInvalidSettlement.Wrapf("malformed treasury record %x", iterator.Key())
		}
		if !bytes.Equal(iterator.Key(), types.TreasuryRecordKey(record.RecordID)) {
			return nil, nil, types.ErrInvalidSettlement.Wrapf("mis-keyed treasury record %x", iterator.Key())
		}
		records = append(records, record)
	}
	balance, err := k.loadTreasuryBalance(ctx)
	if err != nil {
		return nil, nil, err
	}
	return records, balance, nil
}

// ============================================================================
// Invoice Settlement Hooks
// ============================================================================

// OnInvoicePaid is called when an invoice is marked as paid
func (k Keeper) OnInvoicePaid(ctx sdk.Context, invoiceRecord *billing.InvoiceLedgerRecord) error {
	// Check if settlement already exists
	settlements := k.GetSettlementsByOrder(ctx, invoiceRecord.OrderID)
	var matchingSettlement *types.SettlementRecord
	for i, s := range settlements {
		if s.IsFinal || s.LeaseID == invoiceRecord.LeaseID {
			matchingSettlement = &settlements[i]
			break
		}
	}

	if matchingSettlement == nil {
		k.Logger(ctx).Debug("no matching settlement for paid invoice",
			"invoice_id", invoiceRecord.InvoiceID,
			"order_id", invoiceRecord.OrderID,
		)
		return nil
	}

	// Execute payout
	_, err := k.ExecutePayout(ctx, invoiceRecord.InvoiceID, matchingSettlement.SettlementID)
	if err != nil {
		return err
	}

	return nil
}

// OnDisputeOpened is called when a dispute is opened
func (k Keeper) OnDisputeOpened(ctx sdk.Context, invoiceID string, disputeID string, reason string) error {
	payout, found := k.GetPayoutByInvoice(ctx, invoiceID)
	if !found {
		return nil // No payout to hold
	}

	if payout.State.IsTerminal() {
		return nil // Already completed
	}

	return k.HoldPayout(ctx, payout.PayoutID, disputeID, reason)
}

// OnDisputeResolved is called when a dispute is resolved
func (k Keeper) OnDisputeResolved(ctx sdk.Context, invoiceID string, resolution billing.DisputeResolutionType) error {
	if k.IsFinancialCasesActive(ctx) {
		return types.ErrLegacyFinancialMutationFenced
	}
	payout, found := k.GetPayoutByInvoice(ctx, invoiceID)
	if !found {
		return nil
	}

	if payout.State != types.PayoutStateHeld {
		return nil
	}

	switch resolution {
	case billing.DisputeResolutionProviderWin:
		// Release payout to provider
		return k.ReleasePayoutHold(ctx, payout.PayoutID)

	case billing.DisputeResolutionCustomerWin:
		// Refund to customer
		return k.RefundPayout(ctx, payout.PayoutID, "dispute resolved in customer's favor")

	case billing.DisputeResolutionPartialRefund:
		refundAmount := payout.HoldbackAmount
		if refundAmount.IsZero() {
			return k.ReleasePayoutHold(ctx, payout.PayoutID)
		}

		customer, err := sdk.AccAddressFromBech32(payout.Customer)
		if err != nil {
			return types.ErrInvalidAddress.Wrap("invalid customer address")
		}

		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, customer, refundAmount); err != nil {
			return types.ErrPayoutExecutionFailed.Wrap(err.Error())
		}

		payout.HoldbackAmount = sdk.NewCoins()
		if err := k.SetPayout(ctx, payout); err != nil {
			return err
		}

		k.savePayoutLedgerEntry(ctx, payout.PayoutID, types.PayoutLedgerEntryRefunded,
			payout.State, payout.State,
			refundAmount, "partial refund issued", "dispute_resolution")

		if err := ctx.EventManager().EmitTypedEvent(&types.EventPayoutRefunded{
			PayoutID:   payout.PayoutID,
			Customer:   payout.Customer,
			Amount:     refundAmount.String(),
			Reason:     "partial refund",
			RefundedAt: ctx.BlockTime().Unix(),
		}); err != nil {
			k.Logger(ctx).Error("failed to emit payout partial refund event", "error", err)
		}

		k.Logger(ctx).Info("payout partial refund issued",
			"payout_id", payout.PayoutID,
			"amount", refundAmount.String(),
		)

		return k.ReleasePayoutHold(ctx, payout.PayoutID)

	case billing.DisputeResolutionMutualAgreement:
		// Release payout (agreement reached)
		return k.ReleasePayoutHold(ctx, payout.PayoutID)

	default:
		return k.ReleasePayoutHold(ctx, payout.PayoutID)
	}
}
