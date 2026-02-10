package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/escrow/types/billing"
)

// SaveUsageRecord persists a billing usage record on-chain via the reconciliation keeper.
func (k *keeper) SaveUsageRecord(ctx sdk.Context, record *billing.UsageRecord) error {
	return k.NewReconciliationKeeper().SaveUsageRecord(ctx, record)
}

// GetUsageRecord retrieves a billing usage record by ID.
func (k *keeper) GetUsageRecord(ctx sdk.Context, recordID string) (*billing.UsageRecord, error) {
	return k.NewReconciliationKeeper().GetUsageRecord(ctx, recordID)
}

// CreateInvoice stores an invoice ledger record.
func (k *keeper) CreateInvoice(ctx sdk.Context, invoice *billing.Invoice, artifactCID string) (*billing.InvoiceLedgerRecord, error) {
	return k.NewInvoiceKeeper().CreateInvoice(ctx, invoice, artifactCID)
}

// UpdateInvoiceStatus updates invoice status and ledger entry.
func (k *keeper) UpdateInvoiceStatus(ctx sdk.Context, invoiceID string, newStatus billing.InvoiceStatus, initiator string) (*billing.InvoiceLedgerEntry, error) {
	return k.NewInvoiceKeeper().UpdateInvoiceStatus(ctx, invoiceID, newStatus, initiator)
}

// RecordPayment records a payment against the invoice.
func (k *keeper) RecordPayment(ctx sdk.Context, invoiceID string, amount sdk.Coins, initiator string) (*billing.InvoiceLedgerEntry, error) {
	return k.NewInvoiceKeeper().RecordPayment(ctx, invoiceID, amount, initiator)
}

// GetInvoiceSequence returns the current invoice sequence.
func (k *keeper) GetInvoiceSequence(ctx sdk.Context) uint64 {
	return k.NewInvoiceKeeper().GetInvoiceSequence(ctx)
}
