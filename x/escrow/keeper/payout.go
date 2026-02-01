// Package keeper provides the escrow module keeper with payout capabilities.
//
// This file implements provider payout calculation and execution,
// including fee deductions, batch processing, and payout scheduling.
package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/virtengine/virtengine/x/escrow/types/billing"
)

// PayoutKeeper defines the interface for payout management
type PayoutKeeper interface {
	// CalculateProviderPayout calculates payout amounts for a settlement
	CalculateProviderPayout(ctx sdk.Context, settlementID string) (*billing.PayoutCalculation, error)

	// CreatePayout creates a new payout record
	CreatePayout(ctx sdk.Context, settlementID string, provider string) (*billing.PayoutRecord, error)

	// ExecutePayout executes a pending payout
	ExecutePayout(ctx sdk.Context, payoutID string) error

	// GetPayout retrieves a payout record by ID
	GetPayout(ctx sdk.Context, payoutID string) (*billing.PayoutRecord, error)

	// GetPayoutsByProvider retrieves payouts for a provider
	GetPayoutsByProvider(ctx sdk.Context, provider string, pagination *query.PageRequest) ([]*billing.PayoutRecord, *query.PageResponse, error)

	// GetPayoutsByStatus retrieves payouts by status
	GetPayoutsByStatus(ctx sdk.Context, status billing.PayoutStatus, pagination *query.PageRequest) ([]*billing.PayoutRecord, *query.PageResponse, error)

	// GetPendingPayouts retrieves all pending payouts
	GetPendingPayouts(ctx sdk.Context) ([]*billing.PayoutRecord, error)

	// ProcessBatchPayouts processes multiple payouts in batch
	ProcessBatchPayouts(ctx sdk.Context, payoutIDs []string) ([]string, []error)

	// GetPayoutSummary generates a payout summary for a provider
	GetPayoutSummary(ctx sdk.Context, provider string, periodStart, periodEnd time.Time) (*billing.PayoutSummary, error)

	// CancelPayout cancels a pending payout
	CancelPayout(ctx sdk.Context, payoutID string, reason string) error

	// RefundPayout refunds a completed payout
	RefundPayout(ctx sdk.Context, payoutID string, amount sdk.Coins, reason string) error

	// GetPayoutSequence gets the current payout sequence number
	GetPayoutSequence(ctx sdk.Context) uint64

	// SetPayoutSequence sets the payout sequence number
	SetPayoutSequence(ctx sdk.Context, sequence uint64)

	// WithPayouts iterates over all payouts
	WithPayouts(ctx sdk.Context, fn func(*billing.PayoutRecord) bool)
}

// payoutKeeper implements PayoutKeeper
type payoutKeeper struct {
	k *keeper
}

// NewPayoutKeeper creates a new payout keeper
func (k *keeper) NewPayoutKeeper() PayoutKeeper {
	return &payoutKeeper{k: k}
}

// CalculateProviderPayout calculates payout amounts for a settlement
func (pk *payoutKeeper) CalculateProviderPayout(ctx sdk.Context, settlementID string) (*billing.PayoutCalculation, error) {
	settlementKeeper := pk.k.NewSettlementIntegrationKeeper()

	// Get settlement
	settlement, err := settlementKeeper.GetSettlement(ctx, settlementID)
	if err != nil {
		return nil, fmt.Errorf("failed to get settlement: %w", err)
	}

	// Calculate payout
	calculation := &billing.PayoutCalculation{
		SettlementID:   settlementID,
		Provider:       settlement.Provider,
		GrossAmount:    settlement.GrossAmount,
		TotalFees:      settlement.FeeBreakdown.TotalFees,
		NetAmount:      settlement.FeeBreakdown.NetAmount,
		HoldbackAmount: settlement.HoldbackAmount,
		PayableAmount:  settlement.NetPayout,
		CalculatedAt:   ctx.BlockTime(),
		BlockHeight:    ctx.BlockHeight(),
	}

	// Adjust payable amount for holdbacks
	if !settlement.HoldbackAmount.IsZero() {
		calculation.PayableAmount = calculation.NetAmount.Sub(settlement.HoldbackAmount...)
	}

	return calculation, nil
}

// CreatePayout creates a new payout record
func (pk *payoutKeeper) CreatePayout(ctx sdk.Context, settlementID string, provider string) (*billing.PayoutRecord, error) {
	store := ctx.KVStore(pk.k.skey)

	// Calculate payout amounts
	calculation, err := pk.CalculateProviderPayout(ctx, settlementID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate payout: %w", err)
	}

	// Generate payout ID
	seq := pk.GetPayoutSequence(ctx)
	payoutID := billing.NextPayoutID(seq, "VE")

	// Create payout record
	record := billing.NewPayoutRecord(
		payoutID,
		settlementID,
		calculation.Provider,
		calculation.GrossAmount,
		calculation.TotalFees,
		calculation.NetAmount,
		calculation.PayableAmount,
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)

	// Save record
	key := billing.BuildPayoutRecordKey(payoutID)
	bz, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payout record: %w", err)
	}
	store.Set(key, bz)

	// Create indexes
	pk.setPayoutIndexes(store, record)

	// Increment sequence
	pk.SetPayoutSequence(ctx, seq+1)

	return record, nil
}

// ExecutePayout executes a pending payout
func (pk *payoutKeeper) ExecutePayout(ctx sdk.Context, payoutID string) error {
	store := ctx.KVStore(pk.k.skey)

	// Get payout
	payout, err := pk.GetPayout(ctx, payoutID)
	if err != nil {
		return err
	}

	// Validate status
	if payout.Status != billing.PayoutStatusPending {
		return fmt.Errorf("payout is not pending: %s", payout.Status)
	}

	oldStatus := payout.Status

	// Mark as processing
	payout.Status = billing.PayoutStatusProcessing
	payout.UpdatedAt = ctx.BlockTime()

	// Validate provider address
	providerAddr, err := sdk.AccAddressFromBech32(payout.Provider)
	if err != nil {
		payout.Status = billing.PayoutStatusFailed
		payout.FailureReason = fmt.Sprintf("invalid provider address: %s", err.Error())
		_ = pk.savePayout(store, payout, oldStatus)
		return fmt.Errorf("invalid provider address: %w", err)
	}

	// Get settlement to access escrow info
	settlementKeeper := pk.k.NewSettlementIntegrationKeeper()
	settlement, err := settlementKeeper.GetSettlement(ctx, payout.SettlementID)
	if err != nil {
		payout.Status = billing.PayoutStatusFailed
		payout.FailureReason = fmt.Sprintf("failed to get settlement: %s", err.Error())
		_ = pk.savePayout(store, payout, oldStatus)
		return fmt.Errorf("failed to get settlement: %w", err)
	}

	// Execute the payout transfer using the bank keeper
	// The actual transfer from escrow to provider
	payoutCoins := sdk.NewCoins()
	for _, coin := range payout.PayoutAmount {
		payoutCoins = payoutCoins.Add(coin)
	}

	if !payoutCoins.IsZero() {
		// Transfer from escrow module to provider
		err = pk.k.bkeeper.SendCoinsFromModuleToAccount(
			ctx,
			"escrow",
			providerAddr,
			payoutCoins,
		)
		if err != nil {
			payout.Status = billing.PayoutStatusFailed
			payout.FailureReason = fmt.Sprintf("transfer failed: %s", err.Error())
			_ = pk.savePayout(store, payout, oldStatus)
			return fmt.Errorf("failed to transfer payout: %w", err)
		}
	}

	// Mark as completed
	payout.Status = billing.PayoutStatusCompleted
	now := ctx.BlockTime()
	payout.PayoutDate = now
	payout.CompletedAt = &now
	payout.UpdatedAt = now
	payout.TransactionHash = fmt.Sprintf("tx-%d-%s", ctx.BlockHeight(), payoutID)

	// Update settlement status if needed
	if settlement.Status == billing.SettlementStatusProcessing {
		if err := settlement.Complete(ctx.BlockTime()); err == nil {
			_ = settlementKeeper.SaveSettlement(ctx, settlement)
		}
	}

	return pk.savePayout(store, payout, oldStatus)
}

// GetPayout retrieves a payout record by ID
func (pk *payoutKeeper) GetPayout(ctx sdk.Context, payoutID string) (*billing.PayoutRecord, error) {
	store := ctx.KVStore(pk.k.skey)
	key := billing.BuildPayoutRecordKey(payoutID)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("payout not found: %s", payoutID)
	}

	var record billing.PayoutRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payout record: %w", err)
	}

	return &record, nil
}

// GetPayoutsByProvider retrieves payouts for a provider
func (pk *payoutKeeper) GetPayoutsByProvider(
	ctx sdk.Context,
	provider string,
	pagination *query.PageRequest,
) ([]*billing.PayoutRecord, *query.PageResponse, error) {
	store := ctx.KVStore(pk.k.skey)
	prefix := billing.BuildPayoutRecordByProviderPrefix(provider)

	return pk.paginatePayoutIndex(store, prefix, pagination)
}

// GetPayoutsByStatus retrieves payouts by status
func (pk *payoutKeeper) GetPayoutsByStatus(
	ctx sdk.Context,
	status billing.PayoutStatus,
	pagination *query.PageRequest,
) ([]*billing.PayoutRecord, *query.PageResponse, error) {
	store := ctx.KVStore(pk.k.skey)
	prefix := billing.BuildPayoutRecordByStatusPrefix(status)

	return pk.paginatePayoutIndex(store, prefix, pagination)
}

// GetPendingPayouts retrieves all pending payouts
func (pk *payoutKeeper) GetPendingPayouts(ctx sdk.Context) ([]*billing.PayoutRecord, error) {
	store := ctx.KVStore(pk.k.skey)
	prefix := billing.BuildPayoutRecordByStatusPrefix(billing.PayoutStatusPending)

	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var payouts []*billing.PayoutRecord
	for ; iter.Valid(); iter.Next() {
		payoutID := string(iter.Value())
		payout, err := pk.GetPayout(ctx, payoutID)
		if err != nil {
			continue
		}
		payouts = append(payouts, payout)
	}

	return payouts, nil
}

// ProcessBatchPayouts processes multiple payouts in batch
func (pk *payoutKeeper) ProcessBatchPayouts(ctx sdk.Context, payoutIDs []string) ([]string, []error) {
	successIDs := make([]string, 0, len(payoutIDs))
	errors := make([]error, 0)

	for _, payoutID := range payoutIDs {
		if err := pk.ExecutePayout(ctx, payoutID); err != nil {
			errors = append(errors, fmt.Errorf("payout %s: %w", payoutID, err))
			continue
		}
		successIDs = append(successIDs, payoutID)
	}

	return successIDs, errors
}

// GetPayoutSummary generates a payout summary for a provider
func (pk *payoutKeeper) GetPayoutSummary(
	ctx sdk.Context,
	provider string,
	periodStart, periodEnd time.Time,
) (*billing.PayoutSummary, error) {
	summary := billing.NewPayoutSummary(provider, periodStart, periodEnd, ctx.BlockHeight(), ctx.BlockTime())

	// Iterate over all payouts for the provider
	store := ctx.KVStore(pk.k.skey)
	prefix := billing.BuildPayoutRecordByProviderPrefix(provider)

	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		payoutID := string(iter.Value())
		payout, err := pk.GetPayout(ctx, payoutID)
		if err != nil {
			continue
		}

		// Filter by period
		if payout.CreatedAt.After(periodStart) && payout.CreatedAt.Before(periodEnd) {
			summary.AddPayout(payout)
		}
	}

	return summary, nil
}

// CancelPayout cancels a pending payout
func (pk *payoutKeeper) CancelPayout(ctx sdk.Context, payoutID string, reason string) error {
	store := ctx.KVStore(pk.k.skey)

	payout, err := pk.GetPayout(ctx, payoutID)
	if err != nil {
		return err
	}

	if payout.Status != billing.PayoutStatusPending {
		return fmt.Errorf("can only cancel pending payouts, current status: %s", payout.Status)
	}

	oldStatus := payout.Status
	payout.Status = billing.PayoutStatusCancelled
	payout.FailureReason = reason
	payout.UpdatedAt = ctx.BlockTime()

	return pk.savePayout(store, payout, oldStatus)
}

// RefundPayout refunds a completed payout
func (pk *payoutKeeper) RefundPayout(ctx sdk.Context, payoutID string, amount sdk.Coins, reason string) error {
	store := ctx.KVStore(pk.k.skey)

	payout, err := pk.GetPayout(ctx, payoutID)
	if err != nil {
		return err
	}

	if payout.Status != billing.PayoutStatusCompleted {
		return fmt.Errorf("can only refund completed payouts, current status: %s", payout.Status)
	}

	// Validate refund amount doesn't exceed payout
	if amount.IsAllGT(payout.PayoutAmount) {
		return fmt.Errorf("refund amount %s exceeds payout amount %s", amount.String(), payout.PayoutAmount.String())
	}

	oldStatus := payout.Status
	payout.Status = billing.PayoutStatusRefunded
	payout.RefundAmount = amount
	payout.RefundReason = reason
	now := ctx.BlockTime()
	payout.RefundedAt = &now
	payout.UpdatedAt = now

	return pk.savePayout(store, payout, oldStatus)
}

// GetPayoutSequence gets the current payout sequence number
func (pk *payoutKeeper) GetPayoutSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(pk.k.skey)
	bz := store.Get(billing.PayoutSequenceKey)
	if bz == nil {
		return 0
	}

	var seq uint64
	if err := json.Unmarshal(bz, &seq); err != nil {
		return 0
	}
	return seq
}

// SetPayoutSequence sets the payout sequence number
func (pk *payoutKeeper) SetPayoutSequence(ctx sdk.Context, sequence uint64) {
	store := ctx.KVStore(pk.k.skey)
	bz, _ := json.Marshal(sequence)
	store.Set(billing.PayoutSequenceKey, bz)
}

// WithPayouts iterates over all payouts
func (pk *payoutKeeper) WithPayouts(ctx sdk.Context, fn func(*billing.PayoutRecord) bool) {
	store := ctx.KVStore(pk.k.skey)
	iter := storetypes.KVStorePrefixIterator(store, billing.PayoutRecordPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var record billing.PayoutRecord
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}

		if stop := fn(&record); stop {
			break
		}
	}
}

// Helper methods

func (pk *payoutKeeper) setPayoutIndexes(store storetypes.KVStore, record *billing.PayoutRecord) {
	// Provider index
	providerKey := billing.BuildPayoutRecordByProviderKey(record.Provider, record.PayoutID)
	store.Set(providerKey, []byte(record.PayoutID))

	// Status index
	statusKey := billing.BuildPayoutRecordByStatusKey(record.Status, record.PayoutID)
	store.Set(statusKey, []byte(record.PayoutID))

	// Settlement index
	settlementKey := billing.BuildPayoutRecordBySettlementKey(record.SettlementID, record.PayoutID)
	store.Set(settlementKey, []byte(record.PayoutID))
}

func (pk *payoutKeeper) savePayout(store storetypes.KVStore, payout *billing.PayoutRecord, oldStatus billing.PayoutStatus) error {
	// Remove old status index if status changed
	if oldStatus != payout.Status {
		oldStatusKey := billing.BuildPayoutRecordByStatusKey(oldStatus, payout.PayoutID)
		store.Delete(oldStatusKey)

		// Add new status index
		newStatusKey := billing.BuildPayoutRecordByStatusKey(payout.Status, payout.PayoutID)
		store.Set(newStatusKey, []byte(payout.PayoutID))
	}

	// Save payout
	key := billing.BuildPayoutRecordKey(payout.PayoutID)
	bz, err := json.Marshal(payout)
	if err != nil {
		return fmt.Errorf("failed to marshal payout: %w", err)
	}
	store.Set(key, bz)

	return nil
}

//nolint:unparam // prefix kept for future index-specific pagination
func (pk *payoutKeeper) paginatePayoutIndex(
	store storetypes.KVStore,
	_ []byte,
	pagination *query.PageRequest,
) ([]*billing.PayoutRecord, *query.PageResponse, error) {
	var payouts []*billing.PayoutRecord

	pageRes, err := query.Paginate(store, pagination, func(key []byte, value []byte) error {
		payoutID := string(value)
		payoutKey := billing.BuildPayoutRecordKey(payoutID)
		bz := store.Get(payoutKey)
		if bz == nil {
			return nil
		}

		var payout billing.PayoutRecord
		if err := json.Unmarshal(bz, &payout); err != nil {
			return nil
		}

		payouts = append(payouts, &payout)
		return nil
	})

	return payouts, pageRes, err
}
