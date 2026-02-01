// Package keeper provides the escrow module keeper with invoice management capabilities.
package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/virtengine/virtengine/x/escrow/types/billing"
)

// InvoiceKeeper defines the interface for invoice management
type InvoiceKeeper interface {
	// CreateInvoice creates a new invoice and its ledger record
	CreateInvoice(ctx sdk.Context, invoice *billing.Invoice, artifactCID string) (*billing.InvoiceLedgerRecord, error)

	// GetInvoice retrieves an invoice ledger record by ID
	GetInvoice(ctx sdk.Context, invoiceID string) (*billing.InvoiceLedgerRecord, error)

	// GetInvoicesByProvider retrieves invoices by provider
	GetInvoicesByProvider(ctx sdk.Context, provider string, pagination *query.PageRequest) ([]*billing.InvoiceLedgerRecord, *query.PageResponse, error)

	// GetInvoicesByCustomer retrieves invoices by customer
	GetInvoicesByCustomer(ctx sdk.Context, customer string, pagination *query.PageRequest) ([]*billing.InvoiceLedgerRecord, *query.PageResponse, error)

	// GetInvoicesByStatus retrieves invoices by status
	GetInvoicesByStatus(ctx sdk.Context, status billing.InvoiceStatus, pagination *query.PageRequest) ([]*billing.InvoiceLedgerRecord, *query.PageResponse, error)

	// GetInvoicesByEscrow retrieves invoices by escrow ID
	GetInvoicesByEscrow(ctx sdk.Context, escrowID string, pagination *query.PageRequest) ([]*billing.InvoiceLedgerRecord, *query.PageResponse, error)

	// UpdateInvoiceStatus updates an invoice status with state machine enforcement
	UpdateInvoiceStatus(ctx sdk.Context, invoiceID string, newStatus billing.InvoiceStatus, initiator string) (*billing.InvoiceLedgerEntry, error)

	// RecordPayment records a payment against an invoice
	RecordPayment(ctx sdk.Context, invoiceID string, amount sdk.Coins, initiator string) (*billing.InvoiceLedgerEntry, error)

	// GetInvoiceLedgerEntries retrieves ledger entries for an invoice
	GetInvoiceLedgerEntries(ctx sdk.Context, invoiceID string) ([]*billing.InvoiceLedgerEntry, error)

	// WithInvoices iterates over all invoices
	WithInvoices(ctx sdk.Context, fn func(*billing.InvoiceLedgerRecord) bool)

	// GetInvoiceSequence gets the current invoice sequence number
	GetInvoiceSequence(ctx sdk.Context) uint64

	// SetInvoiceSequence sets the invoice sequence number
	SetInvoiceSequence(ctx sdk.Context, sequence uint64)

	// SaveReconciliationReport saves a reconciliation report
	SaveReconciliationReport(ctx sdk.Context, report *billing.ReconciliationReport) error

	// GetReconciliationReport retrieves a reconciliation report
	GetReconciliationReport(ctx sdk.Context, reportID string) (*billing.ReconciliationReport, error)
}

// invoiceKeeper implements InvoiceKeeper
type invoiceKeeper struct {
	k *keeper
}

// NewInvoiceKeeper creates a new invoice keeper from the base keeper
func (k *keeper) NewInvoiceKeeper() InvoiceKeeper {
	return &invoiceKeeper{k: k}
}

// CreateInvoice creates a new invoice and its ledger record
func (ik *invoiceKeeper) CreateInvoice(
	ctx sdk.Context,
	invoice *billing.Invoice,
	artifactCID string,
) (*billing.InvoiceLedgerRecord, error) {
	store := ctx.KVStore(ik.k.skey)

	// Check if invoice already exists
	key := billing.BuildInvoiceLedgerRecordKey(invoice.InvoiceID)
	if store.Has(key) {
		return nil, fmt.Errorf("invoice already exists: %s", invoice.InvoiceID)
	}

	// Create ledger record
	record, err := billing.NewInvoiceLedgerRecord(
		invoice,
		artifactCID,
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ledger record: %w", err)
	}

	// Validate record
	if err := record.Validate(); err != nil {
		return nil, fmt.Errorf("invalid ledger record: %w", err)
	}

	// Marshal and store
	bz, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ledger record: %w", err)
	}
	store.Set(key, bz)

	// Create indexes
	ik.setInvoiceIndexes(store, record)

	// Create initial ledger entry
	entry := billing.NewInvoiceLedgerEntry(
		fmt.Sprintf("%s-created", invoice.InvoiceID),
		invoice.InvoiceID,
		billing.LedgerEntryTypeCreated,
		billing.InvoiceStatusDraft,
		invoice.Status,
		sdk.NewCoins(),
		"invoice created",
		"system",
		"",
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)
	ik.saveLedgerEntry(store, entry)

	// Increment sequence
	seq := ik.GetInvoiceSequence(ctx)
	ik.SetInvoiceSequence(ctx, seq+1)

	return record, nil
}

// GetInvoice retrieves an invoice ledger record by ID
func (ik *invoiceKeeper) GetInvoice(ctx sdk.Context, invoiceID string) (*billing.InvoiceLedgerRecord, error) {
	store := ctx.KVStore(ik.k.skey)
	key := billing.BuildInvoiceLedgerRecordKey(invoiceID)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("invoice not found: %s", invoiceID)
	}

	var record billing.InvoiceLedgerRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ledger record: %w", err)
	}

	return &record, nil
}

// GetInvoicesByProvider retrieves invoices by provider
func (ik *invoiceKeeper) GetInvoicesByProvider(
	ctx sdk.Context,
	provider string,
	pagination *query.PageRequest,
) ([]*billing.InvoiceLedgerRecord, *query.PageResponse, error) {
	store := ctx.KVStore(ik.k.skey)
	prefix := billing.BuildInvoiceByProviderPrefix(provider)

	return ik.paginateInvoiceIndex(store, prefix, pagination)
}

// GetInvoicesByCustomer retrieves invoices by customer
func (ik *invoiceKeeper) GetInvoicesByCustomer(
	ctx sdk.Context,
	customer string,
	pagination *query.PageRequest,
) ([]*billing.InvoiceLedgerRecord, *query.PageResponse, error) {
	store := ctx.KVStore(ik.k.skey)
	prefix := billing.BuildInvoiceByCustomerPrefix(customer)

	return ik.paginateInvoiceIndex(store, prefix, pagination)
}

// GetInvoicesByStatus retrieves invoices by status
func (ik *invoiceKeeper) GetInvoicesByStatus(
	ctx sdk.Context,
	status billing.InvoiceStatus,
	pagination *query.PageRequest,
) ([]*billing.InvoiceLedgerRecord, *query.PageResponse, error) {
	store := ctx.KVStore(ik.k.skey)
	prefix := billing.BuildInvoiceByStatusPrefix(status)

	return ik.paginateInvoiceIndex(store, prefix, pagination)
}

// GetInvoicesByEscrow retrieves invoices by escrow ID
func (ik *invoiceKeeper) GetInvoicesByEscrow(
	ctx sdk.Context,
	escrowID string,
	pagination *query.PageRequest,
) ([]*billing.InvoiceLedgerRecord, *query.PageResponse, error) {
	store := ctx.KVStore(ik.k.skey)
	prefix := billing.BuildInvoiceByEscrowPrefix(escrowID)

	return ik.paginateInvoiceIndex(store, prefix, pagination)
}

// UpdateInvoiceStatus updates an invoice status with state machine enforcement
func (ik *invoiceKeeper) UpdateInvoiceStatus(
	ctx sdk.Context,
	invoiceID string,
	newStatus billing.InvoiceStatus,
	initiator string,
) (*billing.InvoiceLedgerEntry, error) {
	store := ctx.KVStore(ik.k.skey)

	// Get current record
	record, err := ik.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	oldStatus := record.Status

	// Validate transition
	if !billing.IsValidTransition(oldStatus, newStatus) {
		return nil, fmt.Errorf("invalid transition from %s to %s", oldStatus, newStatus)
	}

	// Remove old status index
	oldStatusKey := billing.BuildInvoiceByStatusKey(oldStatus, invoiceID)
	store.Delete(oldStatusKey)

	// Update record
	record.Status = newStatus
	record.UpdatedAt = ctx.BlockTime()

	if newStatus == billing.InvoiceStatusPaid {
		now := ctx.BlockTime()
		record.PaidAt = &now
		record.AmountDue = sdk.NewCoins()
	}

	// Save updated record
	key := billing.BuildInvoiceLedgerRecordKey(invoiceID)
	bz, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated record: %w", err)
	}
	store.Set(key, bz)

	// Add new status index
	newStatusKey := billing.BuildInvoiceByStatusKey(newStatus, invoiceID)
	store.Set(newStatusKey, []byte(invoiceID))

	// Create ledger entry
	transition, _ := billing.GetTransition(oldStatus, newStatus)
	entry := billing.NewInvoiceLedgerEntry(
		fmt.Sprintf("%s-status-%d", invoiceID, ctx.BlockHeight()),
		invoiceID,
		ik.getEntryTypeForStatus(newStatus),
		oldStatus,
		newStatus,
		sdk.NewCoins(),
		transition.Description,
		initiator,
		"",
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)
	ik.saveLedgerEntry(store, entry)

	return entry, nil
}

// RecordPayment records a payment against an invoice
func (ik *invoiceKeeper) RecordPayment(
	ctx sdk.Context,
	invoiceID string,
	amount sdk.Coins,
	initiator string,
) (*billing.InvoiceLedgerEntry, error) {
	store := ctx.KVStore(ik.k.skey)

	// Get current record
	record, err := ik.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	oldStatus := record.Status

	// Check if payment is allowed
	if record.Status.IsTerminal() {
		return nil, fmt.Errorf("cannot record payment on %s invoice", record.Status)
	}

	// Update payment amounts
	record.AmountPaid = record.AmountPaid.Add(amount...)
	record.AmountDue = record.Total.Sub(record.AmountPaid...)

	// Determine new status
	newStatus := record.Status
	if record.AmountDue.IsZero() || record.AmountPaid.IsAllGTE(record.Total) {
		newStatus = billing.InvoiceStatusPaid
		now := ctx.BlockTime()
		record.PaidAt = &now
	} else if record.AmountPaid.IsAllPositive() && oldStatus != billing.InvoiceStatusPartiallyPaid {
		newStatus = billing.InvoiceStatusPartiallyPaid
	}

	// Remove old status index if changed
	if oldStatus != newStatus {
		oldStatusKey := billing.BuildInvoiceByStatusKey(oldStatus, invoiceID)
		store.Delete(oldStatusKey)

		// Add new status index
		newStatusKey := billing.BuildInvoiceByStatusKey(newStatus, invoiceID)
		store.Set(newStatusKey, []byte(invoiceID))
	}

	record.Status = newStatus
	record.UpdatedAt = ctx.BlockTime()

	// Save updated record
	key := billing.BuildInvoiceLedgerRecordKey(invoiceID)
	bz, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated record: %w", err)
	}
	store.Set(key, bz)

	// Create ledger entry
	entry := billing.NewInvoiceLedgerEntry(
		fmt.Sprintf("%s-payment-%d", invoiceID, ctx.BlockHeight()),
		invoiceID,
		billing.LedgerEntryTypePayment,
		oldStatus,
		newStatus,
		amount,
		fmt.Sprintf("payment recorded: %s", amount.String()),
		initiator,
		"",
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)
	ik.saveLedgerEntry(store, entry)

	return entry, nil
}

// GetInvoiceLedgerEntries retrieves ledger entries for an invoice
func (ik *invoiceKeeper) GetInvoiceLedgerEntries(ctx sdk.Context, invoiceID string) ([]*billing.InvoiceLedgerEntry, error) {
	store := ctx.KVStore(ik.k.skey)
	prefix := billing.BuildInvoiceLedgerEntryByInvoicePrefix(invoiceID)

	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var entries []*billing.InvoiceLedgerEntry
	for ; iter.Valid(); iter.Next() {
		entryID := string(iter.Value())
		entryKey := billing.BuildInvoiceLedgerEntryKey(entryID)
		bz := store.Get(entryKey)
		if bz == nil {
			continue
		}

		var entry billing.InvoiceLedgerEntry
		if err := json.Unmarshal(bz, &entry); err != nil {
			continue
		}
		entries = append(entries, &entry)
	}

	return entries, nil
}

// WithInvoices iterates over all invoices
func (ik *invoiceKeeper) WithInvoices(ctx sdk.Context, fn func(*billing.InvoiceLedgerRecord) bool) {
	store := ctx.KVStore(ik.k.skey)
	iter := storetypes.KVStorePrefixIterator(store, billing.InvoiceLedgerRecordPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var record billing.InvoiceLedgerRecord
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}

		if stop := fn(&record); stop {
			break
		}
	}
}

// GetInvoiceSequence gets the current invoice sequence number
func (ik *invoiceKeeper) GetInvoiceSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(ik.k.skey)
	bz := store.Get(billing.InvoiceSequenceKey)
	if bz == nil {
		return 0
	}

	var seq uint64
	if err := json.Unmarshal(bz, &seq); err != nil {
		return 0
	}
	return seq
}

// SetInvoiceSequence sets the invoice sequence number
func (ik *invoiceKeeper) SetInvoiceSequence(ctx sdk.Context, sequence uint64) {
	store := ctx.KVStore(ik.k.skey)
	bz, _ := json.Marshal(sequence)
	store.Set(billing.InvoiceSequenceKey, bz)
}

// SaveReconciliationReport saves a reconciliation report
func (ik *invoiceKeeper) SaveReconciliationReport(ctx sdk.Context, report *billing.ReconciliationReport) error {
	store := ctx.KVStore(ik.k.skey)

	if err := report.Validate(); err != nil {
		return fmt.Errorf("invalid report: %w", err)
	}

	key := billing.BuildReconciliationReportKey(report.ReportID)
	bz, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	store.Set(key, bz)

	// Create indexes
	if report.Provider != "" {
		providerKey := billing.BuildReconciliationReportByProviderKey(report.Provider, report.ReportID)
		store.Set(providerKey, []byte(report.ReportID))
	}

	if report.Customer != "" {
		customerKey := billing.BuildReconciliationReportByCustomerKey(report.Customer, report.ReportID)
		store.Set(customerKey, []byte(report.ReportID))
	}

	return nil
}

// GetReconciliationReport retrieves a reconciliation report
func (ik *invoiceKeeper) GetReconciliationReport(ctx sdk.Context, reportID string) (*billing.ReconciliationReport, error) {
	store := ctx.KVStore(ik.k.skey)
	key := billing.BuildReconciliationReportKey(reportID)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("reconciliation report not found: %s", reportID)
	}

	var report billing.ReconciliationReport
	if err := json.Unmarshal(bz, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal report: %w", err)
	}

	return &report, nil
}

// Helper methods

func (ik *invoiceKeeper) setInvoiceIndexes(store storetypes.KVStore, record *billing.InvoiceLedgerRecord) {
	// Provider index
	providerKey := billing.BuildInvoiceByProviderKey(record.Provider, record.InvoiceID)
	store.Set(providerKey, []byte(record.InvoiceID))

	// Customer index
	customerKey := billing.BuildInvoiceByCustomerKey(record.Customer, record.InvoiceID)
	store.Set(customerKey, []byte(record.InvoiceID))

	// Status index
	statusKey := billing.BuildInvoiceByStatusKey(record.Status, record.InvoiceID)
	store.Set(statusKey, []byte(record.InvoiceID))

	// Escrow index
	escrowKey := billing.BuildInvoiceByEscrowKey(record.EscrowID, record.InvoiceID)
	store.Set(escrowKey, []byte(record.InvoiceID))
}

func (ik *invoiceKeeper) saveLedgerEntry(store storetypes.KVStore, entry *billing.InvoiceLedgerEntry) {
	// Save entry
	entryKey := billing.BuildInvoiceLedgerEntryKey(entry.EntryID)
	bz, _ := json.Marshal(entry)
	store.Set(entryKey, bz)

	// Create invoice index
	indexKey := billing.BuildInvoiceLedgerEntryByInvoiceKey(entry.InvoiceID, entry.EntryID)
	store.Set(indexKey, []byte(entry.EntryID))
}

//nolint:unparam // prefix kept for future index-specific pagination
func (ik *invoiceKeeper) paginateInvoiceIndex(
	store storetypes.KVStore,
	_ []byte,
	pagination *query.PageRequest,
) ([]*billing.InvoiceLedgerRecord, *query.PageResponse, error) {
	var records []*billing.InvoiceLedgerRecord

	pageRes, err := query.Paginate(store, pagination, func(key []byte, value []byte) error {
		invoiceID := string(value)
		recordKey := billing.BuildInvoiceLedgerRecordKey(invoiceID)
		bz := store.Get(recordKey)
		if bz == nil {
			return nil
		}

		var record billing.InvoiceLedgerRecord
		if err := json.Unmarshal(bz, &record); err != nil {
			return nil
		}

		records = append(records, &record)
		return nil
	})

	return records, pageRes, err
}

func (ik *invoiceKeeper) getEntryTypeForStatus(status billing.InvoiceStatus) billing.InvoiceLedgerEntryType {
	switch status {
	case billing.InvoiceStatusPending:
		return billing.LedgerEntryTypeIssued
	case billing.InvoiceStatusPaid, billing.InvoiceStatusPartiallyPaid:
		return billing.LedgerEntryTypePayment
	case billing.InvoiceStatusDisputed:
		return billing.LedgerEntryTypeDisputed
	case billing.InvoiceStatusCancelled:
		return billing.LedgerEntryTypeCancelled
	case billing.InvoiceStatusRefunded:
		return billing.LedgerEntryTypeRefunded
	case billing.InvoiceStatusOverdue:
		return billing.LedgerEntryTypeOverdue
	default:
		return billing.LedgerEntryTypeCreated
	}
}

// InvoiceQueryServer defines the gRPC query server interface for invoices
type InvoiceQueryServer interface {
	// Invoice returns an invoice by ID
	Invoice(ctx context.Context, req *QueryInvoiceRequest) (*QueryInvoiceResponse, error)

	// InvoicesByProvider returns invoices for a provider
	InvoicesByProvider(ctx context.Context, req *QueryInvoicesByProviderRequest) (*QueryInvoicesByProviderResponse, error)

	// InvoicesByCustomer returns invoices for a customer
	InvoicesByCustomer(ctx context.Context, req *QueryInvoicesByCustomerRequest) (*QueryInvoicesByCustomerResponse, error)

	// InvoiceLedger returns the ledger entries for an invoice
	InvoiceLedger(ctx context.Context, req *QueryInvoiceLedgerRequest) (*QueryInvoiceLedgerResponse, error)
}

// Query request/response types

// QueryInvoiceRequest is the request for Invoice query
type QueryInvoiceRequest struct {
	InvoiceID string `json:"invoice_id"`
}

// QueryInvoiceResponse is the response for Invoice query
type QueryInvoiceResponse struct {
	Invoice *billing.InvoiceLedgerRecord `json:"invoice"`
}

// QueryInvoicesByProviderRequest is the request for InvoicesByProvider query
type QueryInvoicesByProviderRequest struct {
	Provider   string              `json:"provider"`
	Pagination *query.PageRequest  `json:"pagination,omitempty"`
}

// QueryInvoicesByProviderResponse is the response for InvoicesByProvider query
type QueryInvoicesByProviderResponse struct {
	Invoices   []*billing.InvoiceLedgerRecord `json:"invoices"`
	Pagination *query.PageResponse            `json:"pagination,omitempty"`
}

// QueryInvoicesByCustomerRequest is the request for InvoicesByCustomer query
type QueryInvoicesByCustomerRequest struct {
	Customer   string              `json:"customer"`
	Pagination *query.PageRequest  `json:"pagination,omitempty"`
}

// QueryInvoicesByCustomerResponse is the response for InvoicesByCustomer query
type QueryInvoicesByCustomerResponse struct {
	Invoices   []*billing.InvoiceLedgerRecord `json:"invoices"`
	Pagination *query.PageResponse            `json:"pagination,omitempty"`
}

// QueryInvoiceLedgerRequest is the request for InvoiceLedger query
type QueryInvoiceLedgerRequest struct {
	InvoiceID string `json:"invoice_id"`
}

// QueryInvoiceLedgerResponse is the response for InvoiceLedger query
type QueryInvoiceLedgerResponse struct {
	Entries []*billing.InvoiceLedgerEntry `json:"entries"`
}
