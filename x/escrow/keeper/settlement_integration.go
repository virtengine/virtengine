// Package keeper provides the escrow module keeper with settlement integration capabilities.
//
// This file implements the settlement integration with the invoicing system,
// wiring invoice settlement to escrow release and treasury accounting.
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

// SettlementIntegrationKeeper defines the interface for settlement integration
type SettlementIntegrationKeeper interface {
	// SettleInvoice settles an invoice and releases escrow funds
	SettleInvoice(ctx sdk.Context, invoiceID string, initiator string) (*billing.SettlementRecord, error)

	// ProcessBatchSettlement processes multiple invoices in batch
	ProcessBatchSettlement(ctx sdk.Context, invoiceIDs []string, initiator string) ([]*billing.SettlementRecord, []error)

	// GetSettlement retrieves a settlement record by ID
	GetSettlement(ctx sdk.Context, settlementID string) (*billing.SettlementRecord, error)

	// GetSettlementByInvoice retrieves settlement for an invoice
	GetSettlementByInvoice(ctx sdk.Context, invoiceID string) (*billing.SettlementRecord, error)

	// GetSettlementsByProvider retrieves settlements by provider
	GetSettlementsByProvider(ctx sdk.Context, provider string, pagination *query.PageRequest) ([]*billing.SettlementRecord, *query.PageResponse, error)

	// GetSettlementsByStatus retrieves settlements by status
	GetSettlementsByStatus(ctx sdk.Context, status billing.SettlementStatus, pagination *query.PageRequest) ([]*billing.SettlementRecord, *query.PageResponse, error)

	// ValidateSettlementPrerequisites validates prerequisites for settlement
	ValidateSettlementPrerequisites(ctx sdk.Context, invoiceID string) error

	// HoldbackForDispute holds back settlement funds for a dispute
	HoldbackForDispute(ctx sdk.Context, settlementID string, amount sdk.Coins, reason string) error

	// ReleaseHoldback releases a dispute holdback
	ReleaseHoldback(ctx sdk.Context, settlementID string) error

	// GetFeeConfig retrieves the current fee configuration
	GetFeeConfig(ctx sdk.Context) (billing.FeeConfig, error)

	// SaveFeeConfig saves the fee configuration
	SaveFeeConfig(ctx sdk.Context, config *billing.FeeConfig) error

	// CreateTreasuryAllocations creates treasury allocations for a settlement
	CreateTreasuryAllocations(ctx sdk.Context, settlement *billing.SettlementRecord) ([]billing.TreasuryAllocation, error)

	// GetTreasuryAllocation retrieves a treasury allocation by ID
	GetTreasuryAllocation(ctx sdk.Context, allocationID string) (*billing.TreasuryAllocation, error)

	// GetAllocationsBySettlement retrieves allocations for a settlement
	GetAllocationsBySettlement(ctx sdk.Context, settlementID string) ([]*billing.TreasuryAllocation, error)

	// GetTreasurySummary generates a treasury summary for a period
	GetTreasurySummary(ctx sdk.Context, periodStart, periodEnd time.Time) (*billing.TreasurySummary, error)

	// GetSettlementSequence gets the current settlement sequence number
	GetSettlementSequence(ctx sdk.Context) uint64

	// SetSettlementSequence sets the settlement sequence number
	SetSettlementSequence(ctx sdk.Context, sequence uint64)

	// WithSettlements iterates over all settlements
	WithSettlements(ctx sdk.Context, fn func(*billing.SettlementRecord) bool)

	// SaveSettlement saves a settlement record
	SaveSettlement(ctx sdk.Context, settlement *billing.SettlementRecord) error
}

// settlementIntegrationKeeper implements SettlementIntegrationKeeper
type settlementIntegrationKeeper struct {
	k *keeper
}

// NewSettlementIntegrationKeeper creates a new settlement integration keeper
func (k *keeper) NewSettlementIntegrationKeeper() SettlementIntegrationKeeper {
	return &settlementIntegrationKeeper{k: k}
}

// SettleInvoice settles an invoice and releases escrow funds
func (sik *settlementIntegrationKeeper) SettleInvoice(
	ctx sdk.Context,
	invoiceID string,
	initiator string,
) (*billing.SettlementRecord, error) {
	// Validate prerequisites
	if err := sik.ValidateSettlementPrerequisites(ctx, invoiceID); err != nil {
		return nil, fmt.Errorf("settlement prerequisites not met: %w", err)
	}

	// Get invoice
	invoiceKeeper := sik.k.NewInvoiceKeeper()
	invoice, err := invoiceKeeper.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Get fee configuration
	feeConfig, err := sik.GetFeeConfig(ctx)
	if err != nil {
		// Use defaults if not configured
		feeConfig = billing.DefaultFeeConfig()
	}

	// Calculate fee breakdown
	var feeBreakdown billing.FeeBreakdown
	if feeConfig.IsExempt(invoice.Provider) {
		// Provider is fee-exempt
		feeBreakdown = billing.FeeBreakdown{
			GrossAmount:  invoice.Total,
			PlatformFee:  sdk.NewCoins(),
			NetworkFee:   sdk.NewCoins(),
			CommunityFee: sdk.NewCoins(),
			TakeFee:      sdk.NewCoins(),
			TotalFees:    sdk.NewCoins(),
			NetAmount:    invoice.Total,
		}
	} else {
		feeBreakdown = feeConfig.CalculateFees(invoice.Total)
	}

	// Generate settlement ID
	seq := sik.GetSettlementSequence(ctx)
	settlementID := billing.NextSettlementID(seq, "VE")

	// Create settlement record
	settlement := billing.NewSettlementRecord(
		settlementID,
		invoiceID,
		invoice.EscrowID,
		invoice.Provider,
		invoice.Customer,
		invoice.Total,
		feeBreakdown,
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)

	// Check for active disputes that require holdback
	disputeKeeper := sik.k.NewDisputeKeeper()
	disputes, err := disputeKeeper.GetDisputesByInvoice(ctx, invoiceID)
	if err == nil && len(disputes) > 0 {
		// Check for unresolved disputes
		for _, dispute := range disputes {
			if dispute.Status != billing.DisputeStatusResolved &&
				dispute.Status != billing.DisputeStatusClosed &&
				dispute.Status != billing.DisputeStatusExpired {
				// Set holdback for disputed amount
				if err := settlement.SetHoldback(dispute.DisputedAmount, fmt.Sprintf("dispute_%s", dispute.DisputeID)); err != nil {
					return nil, fmt.Errorf("failed to set holdback for dispute: %w", err)
				}
				break // Only apply first active dispute
			}
		}
	}

	// Mark settlement as processing
	settlement.Status = billing.SettlementStatusProcessing
	settlement.UpdatedAt = ctx.BlockTime()

	// Save settlement
	if err := sik.SaveSettlement(ctx, settlement); err != nil {
		return nil, fmt.Errorf("failed to save settlement: %w", err)
	}

	// Create treasury allocations
	allocations, err := sik.CreateTreasuryAllocations(ctx, settlement)
	if err != nil {
		// Mark as failed if allocation creation fails
		settlement.Status = billing.SettlementStatusFailed
		_ = sik.SaveSettlement(ctx, settlement)
		return nil, fmt.Errorf("failed to create treasury allocations: %w", err)
	}

	// Add allocations to settlement
	for _, alloc := range allocations {
		settlement.AddAllocation(alloc)
	}

	// Update invoice status to paid
	_, err = invoiceKeeper.UpdateInvoiceStatus(ctx, invoiceID, billing.InvoiceStatusPaid, initiator)
	if err != nil {
		// Mark as failed if invoice update fails
		settlement.Status = billing.SettlementStatusFailed
		_ = sik.SaveSettlement(ctx, settlement)
		return nil, fmt.Errorf("failed to update invoice status: %w", err)
	}

	// Complete settlement (if no holdback)
	if settlement.Status != billing.SettlementStatusHeldBack {
		if err := settlement.Complete(ctx.BlockTime()); err != nil {
			return nil, fmt.Errorf("failed to complete settlement: %w", err)
		}
	}

	// Save final settlement state
	if err := sik.SaveSettlement(ctx, settlement); err != nil {
		return nil, fmt.Errorf("failed to save completed settlement: %w", err)
	}

	// Increment sequence
	sik.SetSettlementSequence(ctx, seq+1)

	return settlement, nil
}

// ProcessBatchSettlement processes multiple invoices in batch
func (sik *settlementIntegrationKeeper) ProcessBatchSettlement(
	ctx sdk.Context,
	invoiceIDs []string,
	initiator string,
) ([]*billing.SettlementRecord, []error) {
	settlements := make([]*billing.SettlementRecord, 0, len(invoiceIDs))
	errors := make([]error, 0)

	for _, invoiceID := range invoiceIDs {
		settlement, err := sik.SettleInvoice(ctx, invoiceID, initiator)
		if err != nil {
			errors = append(errors, fmt.Errorf("invoice %s: %w", invoiceID, err))
			continue
		}
		settlements = append(settlements, settlement)
	}

	return settlements, errors
}

// GetSettlement retrieves a settlement record by ID
func (sik *settlementIntegrationKeeper) GetSettlement(ctx sdk.Context, settlementID string) (*billing.SettlementRecord, error) {
	store := ctx.KVStore(sik.k.skey)
	key := billing.BuildSettlementRecordKey(settlementID)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("settlement not found: %s", settlementID)
	}

	var settlement billing.SettlementRecord
	if err := json.Unmarshal(bz, &settlement); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settlement: %w", err)
	}

	return &settlement, nil
}

// GetSettlementByInvoice retrieves settlement for an invoice
func (sik *settlementIntegrationKeeper) GetSettlementByInvoice(ctx sdk.Context, invoiceID string) (*billing.SettlementRecord, error) {
	store := ctx.KVStore(sik.k.skey)
	prefix := billing.BuildSettlementByInvoicePrefix(invoiceID)

	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	// Return the first settlement found (typically one per invoice)
	if iter.Valid() {
		settlementID := string(iter.Value())
		return sik.GetSettlement(ctx, settlementID)
	}

	return nil, fmt.Errorf("no settlement found for invoice: %s", invoiceID)
}

// GetSettlementsByProvider retrieves settlements by provider
func (sik *settlementIntegrationKeeper) GetSettlementsByProvider(
	ctx sdk.Context,
	provider string,
	pagination *query.PageRequest,
) ([]*billing.SettlementRecord, *query.PageResponse, error) {
	store := ctx.KVStore(sik.k.skey)
	prefix := billing.BuildSettlementByProviderPrefix(provider)

	return sik.paginateSettlementIndex(store, prefix, pagination)
}

// GetSettlementsByStatus retrieves settlements by status
func (sik *settlementIntegrationKeeper) GetSettlementsByStatus(
	ctx sdk.Context,
	status billing.SettlementStatus,
	pagination *query.PageRequest,
) ([]*billing.SettlementRecord, *query.PageResponse, error) {
	store := ctx.KVStore(sik.k.skey)
	prefix := billing.BuildSettlementByStatusPrefix(status)

	return sik.paginateSettlementIndex(store, prefix, pagination)
}

// ValidateSettlementPrerequisites validates prerequisites for settlement
func (sik *settlementIntegrationKeeper) ValidateSettlementPrerequisites(ctx sdk.Context, invoiceID string) error {
	invoiceKeeper := sik.k.NewInvoiceKeeper()

	// Get invoice
	invoice, err := invoiceKeeper.GetInvoice(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	// Check invoice status
	switch invoice.Status {
	case billing.InvoiceStatusPaid:
		return fmt.Errorf("invoice already paid")
	case billing.InvoiceStatusCancelled:
		return fmt.Errorf("invoice is cancelled")
	case billing.InvoiceStatusRefunded:
		return fmt.Errorf("invoice is refunded")
	case billing.InvoiceStatusDraft:
		return fmt.Errorf("invoice is still in draft status")
	case billing.InvoiceStatusDisputed:
		// Disputed invoices can be settled with holdback
	}

	// Check if settlement already exists
	_, err = sik.GetSettlementByInvoice(ctx, invoiceID)
	if err == nil {
		return fmt.Errorf("settlement already exists for invoice")
	}

	// Validate provider and customer addresses
	if _, err := sdk.AccAddressFromBech32(invoice.Provider); err != nil {
		return fmt.Errorf("invalid provider address: %w", err)
	}

	if _, err := sdk.AccAddressFromBech32(invoice.Customer); err != nil {
		return fmt.Errorf("invalid customer address: %w", err)
	}

	// Validate amounts
	if !invoice.Total.IsValid() || invoice.Total.IsZero() {
		return fmt.Errorf("invoice total is invalid or zero")
	}

	return nil
}

// HoldbackForDispute holds back settlement funds for a dispute
func (sik *settlementIntegrationKeeper) HoldbackForDispute(
	ctx sdk.Context,
	settlementID string,
	amount sdk.Coins,
	reason string,
) error {
	settlement, err := sik.GetSettlement(ctx, settlementID)
	if err != nil {
		return err
	}

	if err := settlement.SetHoldback(amount, reason); err != nil {
		return fmt.Errorf("failed to set holdback: %w", err)
	}

	return sik.SaveSettlement(ctx, settlement)
}

// ReleaseHoldback releases a dispute holdback
func (sik *settlementIntegrationKeeper) ReleaseHoldback(ctx sdk.Context, settlementID string) error {
	settlement, err := sik.GetSettlement(ctx, settlementID)
	if err != nil {
		return err
	}

	if err := settlement.ReleaseHoldback(); err != nil {
		return fmt.Errorf("failed to release holdback: %w", err)
	}

	// If settlement was previously held back, complete it now
	if err := settlement.Complete(ctx.BlockTime()); err != nil {
		return fmt.Errorf("failed to complete settlement: %w", err)
	}

	return sik.SaveSettlement(ctx, settlement)
}

// GetFeeConfig retrieves the current fee configuration
func (sik *settlementIntegrationKeeper) GetFeeConfig(ctx sdk.Context) (billing.FeeConfig, error) {
	store := ctx.KVStore(sik.k.skey)
	bz := store.Get(billing.FeeConfigKey)

	if bz == nil {
		return billing.FeeConfig{}, fmt.Errorf("fee config not found")
	}

	var config billing.FeeConfig
	if err := json.Unmarshal(bz, &config); err != nil {
		return billing.FeeConfig{}, fmt.Errorf("failed to unmarshal fee config: %w", err)
	}

	return config, nil
}

// SaveFeeConfig saves the fee configuration
func (sik *settlementIntegrationKeeper) SaveFeeConfig(ctx sdk.Context, config *billing.FeeConfig) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid fee config: %w", err)
	}

	store := ctx.KVStore(sik.k.skey)
	bz, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal fee config: %w", err)
	}

	store.Set(billing.FeeConfigKey, bz)
	return nil
}

// CreateTreasuryAllocations creates treasury allocations for a settlement
func (sik *settlementIntegrationKeeper) CreateTreasuryAllocations(
	ctx sdk.Context,
	settlement *billing.SettlementRecord,
) ([]billing.TreasuryAllocation, error) {
	store := ctx.KVStore(sik.k.skey)
	allocations := make([]billing.TreasuryAllocation, 0)
	fb := settlement.FeeBreakdown

	// Platform fee allocation
	if !fb.PlatformFee.IsZero() {
		alloc := billing.TreasuryAllocation{
			AllocationID: billing.NextTreasuryAllocationID(settlement.SettlementID, billing.FeeTypePlatform),
			FeeType:      billing.FeeTypePlatform,
			InvoiceID:    settlement.InvoiceID,
			SettlementID: settlement.SettlementID,
			Amount:       fb.PlatformFee,
			Destination:  "platform_revenue",
			Description:  "Platform service fee",
			BlockHeight:  ctx.BlockHeight(),
			Timestamp:    ctx.BlockTime(),
			Status:       billing.TreasuryAllocationStatusCompleted,
		}
		if err := sik.saveTreasuryAllocation(store, &alloc); err != nil {
			return nil, err
		}
		allocations = append(allocations, alloc)
	}

	// Network fee allocation
	if !fb.NetworkFee.IsZero() {
		alloc := billing.TreasuryAllocation{
			AllocationID: billing.NextTreasuryAllocationID(settlement.SettlementID, billing.FeeTypeNetwork),
			FeeType:      billing.FeeTypeNetwork,
			InvoiceID:    settlement.InvoiceID,
			SettlementID: settlement.SettlementID,
			Amount:       fb.NetworkFee,
			Destination:  "network_fees",
			Description:  "Network usage fee",
			BlockHeight:  ctx.BlockHeight(),
			Timestamp:    ctx.BlockTime(),
			Status:       billing.TreasuryAllocationStatusCompleted,
		}
		if err := sik.saveTreasuryAllocation(store, &alloc); err != nil {
			return nil, err
		}
		allocations = append(allocations, alloc)
	}

	// Community pool allocation
	if !fb.CommunityFee.IsZero() {
		alloc := billing.TreasuryAllocation{
			AllocationID: billing.NextTreasuryAllocationID(settlement.SettlementID, billing.FeeTypeCommunity),
			FeeType:      billing.FeeTypeCommunity,
			InvoiceID:    settlement.InvoiceID,
			SettlementID: settlement.SettlementID,
			Amount:       fb.CommunityFee,
			Destination:  "distribution",
			Description:  "Community pool contribution",
			BlockHeight:  ctx.BlockHeight(),
			Timestamp:    ctx.BlockTime(),
			Status:       billing.TreasuryAllocationStatusCompleted,
		}
		if err := sik.saveTreasuryAllocation(store, &alloc); err != nil {
			return nil, err
		}
		allocations = append(allocations, alloc)
	}

	// Take fee allocation (to validator/community pool via x/take)
	if !fb.TakeFee.IsZero() {
		alloc := billing.TreasuryAllocation{
			AllocationID: billing.NextTreasuryAllocationID(settlement.SettlementID, billing.FeeTypeTake),
			FeeType:      billing.FeeTypeTake,
			InvoiceID:    settlement.InvoiceID,
			SettlementID: settlement.SettlementID,
			Amount:       fb.TakeFee,
			Destination:  "distribution",
			Description:  "Take rate fee",
			BlockHeight:  ctx.BlockHeight(),
			Timestamp:    ctx.BlockTime(),
			Status:       billing.TreasuryAllocationStatusCompleted,
		}
		if err := sik.saveTreasuryAllocation(store, &alloc); err != nil {
			return nil, err
		}
		allocations = append(allocations, alloc)
	}

	return allocations, nil
}

// GetTreasuryAllocation retrieves a treasury allocation by ID
func (sik *settlementIntegrationKeeper) GetTreasuryAllocation(ctx sdk.Context, allocationID string) (*billing.TreasuryAllocation, error) {
	store := ctx.KVStore(sik.k.skey)
	key := billing.BuildTreasuryAllocationKey(allocationID)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("treasury allocation not found: %s", allocationID)
	}

	var alloc billing.TreasuryAllocation
	if err := json.Unmarshal(bz, &alloc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal treasury allocation: %w", err)
	}

	return &alloc, nil
}

// GetAllocationsBySettlement retrieves allocations for a settlement
func (sik *settlementIntegrationKeeper) GetAllocationsBySettlement(ctx sdk.Context, settlementID string) ([]*billing.TreasuryAllocation, error) {
	store := ctx.KVStore(sik.k.skey)
	prefix := billing.BuildTreasuryAllocationBySettlementPrefix(settlementID)

	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var allocations []*billing.TreasuryAllocation
	for ; iter.Valid(); iter.Next() {
		allocationID := string(iter.Value())
		alloc, err := sik.GetTreasuryAllocation(ctx, allocationID)
		if err != nil {
			continue
		}
		allocations = append(allocations, alloc)
	}

	return allocations, nil
}

// GetTreasurySummary generates a treasury summary for a period
func (sik *settlementIntegrationKeeper) GetTreasurySummary(
	ctx sdk.Context,
	periodStart, periodEnd time.Time,
) (*billing.TreasurySummary, error) {
	summary := billing.NewTreasurySummary(periodStart, periodEnd, ctx.BlockHeight(), ctx.BlockTime())

	// Iterate over all settlements and filter by period
	sik.WithSettlements(ctx, func(settlement *billing.SettlementRecord) bool {
		// Check if settlement falls within period
		if settlement.CreatedAt.After(periodStart) && settlement.CreatedAt.Before(periodEnd) {
			summary.AddSettlement(settlement)
		}
		return false // Continue iterating
	})

	return summary, nil
}

// GetSettlementSequence gets the current settlement sequence number
func (sik *settlementIntegrationKeeper) GetSettlementSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(sik.k.skey)
	bz := store.Get(billing.SettlementSequenceKey)
	if bz == nil {
		return 0
	}

	var seq uint64
	if err := json.Unmarshal(bz, &seq); err != nil {
		return 0
	}
	return seq
}

// SetSettlementSequence sets the settlement sequence number
func (sik *settlementIntegrationKeeper) SetSettlementSequence(ctx sdk.Context, sequence uint64) {
	store := ctx.KVStore(sik.k.skey)
	bz, _ := json.Marshal(sequence)
	store.Set(billing.SettlementSequenceKey, bz)
}

// WithSettlements iterates over all settlements
func (sik *settlementIntegrationKeeper) WithSettlements(ctx sdk.Context, fn func(*billing.SettlementRecord) bool) {
	store := ctx.KVStore(sik.k.skey)
	iter := storetypes.KVStorePrefixIterator(store, billing.SettlementRecordPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var settlement billing.SettlementRecord
		if err := json.Unmarshal(iter.Value(), &settlement); err != nil {
			continue
		}

		if stop := fn(&settlement); stop {
			break
		}
	}
}

// SaveSettlement saves a settlement record
func (sik *settlementIntegrationKeeper) SaveSettlement(ctx sdk.Context, settlement *billing.SettlementRecord) error {
	store := ctx.KVStore(sik.k.skey)

	// Validate
	if err := settlement.Validate(); err != nil {
		return fmt.Errorf("invalid settlement: %w", err)
	}

	// Update timestamp
	settlement.UpdatedAt = ctx.BlockTime()

	// Marshal and store
	key := billing.BuildSettlementRecordKey(settlement.SettlementID)
	bz, err := json.Marshal(settlement)
	if err != nil {
		return fmt.Errorf("failed to marshal settlement: %w", err)
	}
	store.Set(key, bz)

	// Create indexes
	sik.setSettlementIndexes(store, settlement)

	return nil
}

// Helper methods

func (sik *settlementIntegrationKeeper) setSettlementIndexes(store storetypes.KVStore, settlement *billing.SettlementRecord) {
	// Invoice index
	invoiceKey := billing.BuildSettlementByInvoiceKey(settlement.InvoiceID, settlement.SettlementID)
	store.Set(invoiceKey, []byte(settlement.SettlementID))

	// Provider index
	providerKey := billing.BuildSettlementByProviderKey(settlement.Provider, settlement.SettlementID)
	store.Set(providerKey, []byte(settlement.SettlementID))

	// Status index
	statusKey := billing.BuildSettlementByStatusKey(settlement.Status, settlement.SettlementID)
	store.Set(statusKey, []byte(settlement.SettlementID))

	// Escrow index
	escrowKey := billing.BuildSettlementByEscrowKey(settlement.EscrowID, settlement.SettlementID)
	store.Set(escrowKey, []byte(settlement.SettlementID))
}

func (sik *settlementIntegrationKeeper) saveTreasuryAllocation(store storetypes.KVStore, alloc *billing.TreasuryAllocation) error {
	// Save allocation
	key := billing.BuildTreasuryAllocationKey(alloc.AllocationID)
	bz, err := json.Marshal(alloc)
	if err != nil {
		return fmt.Errorf("failed to marshal allocation: %w", err)
	}
	store.Set(key, bz)

	// Create settlement index
	settlementKey := billing.BuildTreasuryAllocationBySettlementKey(alloc.SettlementID, alloc.AllocationID)
	store.Set(settlementKey, []byte(alloc.AllocationID))

	// Create fee type index
	feeTypeKey := billing.BuildTreasuryAllocationByFeeTypeKey(alloc.FeeType, alloc.AllocationID)
	store.Set(feeTypeKey, []byte(alloc.AllocationID))

	return nil
}

//nolint:unparam // prefix kept for future index-specific pagination
func (sik *settlementIntegrationKeeper) paginateSettlementIndex(
	store storetypes.KVStore,
	_ []byte,
	pagination *query.PageRequest,
) ([]*billing.SettlementRecord, *query.PageResponse, error) {
	var settlements []*billing.SettlementRecord

	pageRes, err := query.Paginate(store, pagination, func(key []byte, value []byte) error {
		settlementID := string(value)
		settlementKey := billing.BuildSettlementRecordKey(settlementID)
		bz := store.Get(settlementKey)
		if bz == nil {
			return nil
		}

		var settlement billing.SettlementRecord
		if err := json.Unmarshal(bz, &settlement); err != nil {
			return nil
		}

		settlements = append(settlements, &settlement)
		return nil
	})

	return settlements, pageRes, err
}
