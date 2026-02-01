package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

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

// ============================================================================
// Payout Storage
// ============================================================================

// SetPayout saves a payout record to the store
func (k Keeper) SetPayout(ctx sdk.Context, payout types.PayoutRecord) error {
	if err := payout.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

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
	holdbackAmount := sdk.NewCoins()
	params := k.GetParams(ctx)
	if params.PayoutHoldbackRate != "" {
		holdbackRate, err := sdkmath.LegacyNewDecFromStr(params.PayoutHoldbackRate)
		if err == nil && holdbackRate.IsPositive() {
			for _, coin := range settlement.TotalAmount {
				holdbackCoin := sdk.NewCoin(coin.Denom, holdbackRate.MulInt(coin.Amount).TruncateInt())
				holdbackAmount = holdbackAmount.Add(holdbackCoin)
			}
		}
	}

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

	// Execute the payout immediately
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
	k.recordTreasuryEntry(ctx, payout, types.TreasuryRecordPlatformFee, payout.PlatformFee)
	k.recordTreasuryEntry(ctx, payout, types.TreasuryRecordValidatorFee, payout.ValidatorFee)
	if !payout.HoldbackAmount.IsZero() {
		k.recordTreasuryEntry(ctx, payout, types.TreasuryRecordHoldback, payout.HoldbackAmount)
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

	return k.executePayoutTransfer(ctx, &payout)
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

// RefundPayout refunds a held payout to the customer
func (k Keeper) RefundPayout(ctx sdk.Context, payoutID string, reason string) error {
	payout, found := k.GetPayout(ctx, payoutID)
	if !found {
		return types.ErrPayoutNotFound.Wrapf("payout %s not found", payoutID)
	}

	if payout.State != types.PayoutStateHeld {
		return types.ErrInvalidStateTransition.Wrap("can only refund held payouts")
	}

	oldState := payout.State

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

	// Record treasury refund
	k.recordTreasuryEntry(ctx, &payout, types.TreasuryRecordRefund, payout.GrossAmount)

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
		if payout.ExecutionAttempts >= maxRetries {
			continue // Max retries exceeded
		}

		// Reset to pending for retry
		payout.State = types.PayoutStatePending
		if err := k.SetPayout(ctx, payout); err != nil {
			continue
		}

		k.updatePayoutState(ctx, payout, types.PayoutStateFailed)

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
) {
	if amount.IsZero() {
		return
	}

	store := ctx.KVStore(k.skey)

	// Get current treasury balance
	balance := k.getTreasuryBalance(ctx)

	// Update balance based on record type
	var balanceAfter sdk.Coins
	switch recordType {
	case types.TreasuryRecordPlatformFee, types.TreasuryRecordValidatorFee, types.TreasuryRecordHoldback:
		balanceAfter = balance.Add(amount...)
	case types.TreasuryRecordRefund, types.TreasuryRecordWithdrawal:
		balanceAfter = balance.Sub(amount...)
	default:
		balanceAfter = balance
	}

	// Create treasury record
	seq := k.incrementTreasurySequence(ctx)
	recordID := fmt.Sprintf("treasury-%d-%d", ctx.BlockTime().Unix(), seq)

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
		return // silently skip if marshal fails
	}
	store.Set(types.TreasuryRecordKey(recordID), bz)

	// Update treasury balance
	k.setTreasuryBalance(ctx, balanceAfter)
}

func (k Keeper) getTreasuryBalance(ctx sdk.Context) sdk.Coins {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.PrefixTreasuryBalance)
	if bz == nil {
		return sdk.NewCoins()
	}

	var balance sdk.Coins
	if err := json.Unmarshal(bz, &balance); err != nil {
		return sdk.NewCoins()
	}
	return balance
}

func (k Keeper) setTreasuryBalance(ctx sdk.Context, balance sdk.Coins) {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&balance)
	if err != nil {
		return // silently skip if marshal fails
	}
	store.Set(types.PrefixTreasuryBalance, bz)
}

func (k Keeper) incrementTreasurySequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	key := append(types.PrefixTreasuryRecord, []byte("_seq")...)
	bz := store.Get(key)
	var seq uint64
	if bz != nil {
		seq = binary.BigEndian.Uint64(bz)
	}
	seq++
	newBz := make([]byte, 8)
	binary.BigEndian.PutUint64(newBz, seq)
	store.Set(key, newBz)
	return seq
}

// GetTreasuryBalance returns the current treasury balance
func (k Keeper) GetTreasuryBalance(ctx sdk.Context) sdk.Coins {
	return k.getTreasuryBalance(ctx)
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
	_, err := k.ExecutePayout(ctx, matchingSettlement.SettlementID, invoiceRecord.InvoiceID)
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
		// TODO: Implement partial refund logic
		return k.ReleasePayoutHold(ctx, payout.PayoutID)

	case billing.DisputeResolutionMutualAgreement:
		// Release payout (agreement reached)
		return k.ReleasePayoutHold(ctx, payout.PayoutID)

	default:
		return k.ReleasePayoutHold(ctx, payout.PayoutID)
	}
}
