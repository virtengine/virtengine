package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/escrow/types/billing"
)

// BillingKeeper defines the subset of escrow billing functionality used by HPC.
type BillingKeeper interface {
	SaveUsageRecord(ctx sdk.Context, record *billing.UsageRecord) error
	CreateInvoice(ctx sdk.Context, invoice *billing.Invoice, artifactCID string) (*billing.InvoiceLedgerRecord, error)
	UpdateInvoiceStatus(ctx sdk.Context, invoiceID string, newStatus billing.InvoiceStatus, initiator string) (*billing.InvoiceLedgerEntry, error)
	RecordPayment(ctx sdk.Context, invoiceID string, amount sdk.Coins, initiator string) (*billing.InvoiceLedgerEntry, error)
	GetInvoiceSequence(ctx sdk.Context) uint64
}

// SetBillingKeeper configures the billing integration keeper.
func (k *Keeper) SetBillingKeeper(billingKeeper BillingKeeper) {
	k.billingKeeper = billingKeeper
}

func (k Keeper) billingEnabled() bool {
	return k.billingKeeper != nil
}
