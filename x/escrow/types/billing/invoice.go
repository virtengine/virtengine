// Package billing provides billing and invoice types for the escrow module.
//
// This package defines invoice schemas, line-item models, pricing rules,
// tax handling, and settlement hooks for VirtEngine marketplace billing.
package billing

import (
	"encoding/json"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InvoiceStatus defines the status of an invoice
type InvoiceStatus uint8

const (
	// InvoiceStatusDraft is an invoice being prepared
	InvoiceStatusDraft InvoiceStatus = 0

	// InvoiceStatusPending is awaiting payment
	InvoiceStatusPending InvoiceStatus = 1

	// InvoiceStatusPaid is fully paid
	InvoiceStatusPaid InvoiceStatus = 2

	// InvoiceStatusPartiallyPaid has partial payment
	InvoiceStatusPartiallyPaid InvoiceStatus = 3

	// InvoiceStatusOverdue is past due date
	InvoiceStatusOverdue InvoiceStatus = 4

	// InvoiceStatusDisputed is under dispute
	InvoiceStatusDisputed InvoiceStatus = 5

	// InvoiceStatusCancelled is cancelled/voided
	InvoiceStatusCancelled InvoiceStatus = 6

	// InvoiceStatusRefunded is refunded
	InvoiceStatusRefunded InvoiceStatus = 7
)

// InvoiceStatusNames maps status to human-readable names
var InvoiceStatusNames = map[InvoiceStatus]string{
	InvoiceStatusDraft:         "draft",
	InvoiceStatusPending:       "pending",
	InvoiceStatusPaid:          "paid",
	InvoiceStatusPartiallyPaid: "partially_paid",
	InvoiceStatusOverdue:       "overdue",
	InvoiceStatusDisputed:      "disputed",
	InvoiceStatusCancelled:     "cancelled",
	InvoiceStatusRefunded:      "refunded",
}

// String returns string representation
func (s InvoiceStatus) String() string {
	if name, ok := InvoiceStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// IsTerminal returns true if the status is final
func (s InvoiceStatus) IsTerminal() bool {
	return s == InvoiceStatusPaid || s == InvoiceStatusCancelled || s == InvoiceStatusRefunded
}

// Invoice represents a billing invoice for marketplace services
type Invoice struct {
	// InvoiceID is the unique identifier
	InvoiceID string `json:"invoice_id"`

	// InvoiceNumber is the human-readable invoice number
	InvoiceNumber string `json:"invoice_number"`

	// EscrowID links to the escrow account
	EscrowID string `json:"escrow_id"`

	// OrderID links to the marketplace order
	OrderID string `json:"order_id"`

	// LeaseID links to the marketplace lease
	LeaseID string `json:"lease_id"`

	// Provider is the service provider address
	Provider string `json:"provider"`

	// Customer is the customer address
	Customer string `json:"customer"`

	// Status is the current invoice status
	Status InvoiceStatus `json:"status"`

	// BillingPeriod defines the billing period
	BillingPeriod BillingPeriod `json:"billing_period"`

	// LineItems are the individual charges
	LineItems []LineItem `json:"line_items"`

	// Subtotal is the sum of line items before adjustments
	Subtotal sdk.Coins `json:"subtotal"`

	// Discounts applied to this invoice
	Discounts []AppliedDiscount `json:"discounts,omitempty"`

	// DiscountTotal is the total discount amount
	DiscountTotal sdk.Coins `json:"discount_total"`

	// TaxDetails contains tax calculations
	TaxDetails *TaxDetails `json:"tax_details,omitempty"`

	// TaxTotal is the total tax amount
	TaxTotal sdk.Coins `json:"tax_total"`

	// Total is the final amount due (subtotal - discounts + tax)
	Total sdk.Coins `json:"total"`

	// Currency is the primary currency/denomination
	Currency string `json:"currency"`

	// AmountPaid is the amount already paid
	AmountPaid sdk.Coins `json:"amount_paid"`

	// AmountDue is the remaining amount due
	AmountDue sdk.Coins `json:"amount_due"`

	// DueDate is when payment is due
	DueDate time.Time `json:"due_date"`

	// IssuedAt is when the invoice was issued
	IssuedAt time.Time `json:"issued_at"`

	// PaidAt is when fully paid (if paid)
	PaidAt *time.Time `json:"paid_at,omitempty"`

	// SettlementID links to the settlement record (if settled)
	SettlementID string `json:"settlement_id,omitempty"`

	// DisputeWindow defines the dispute period
	DisputeWindow *DisputeWindow `json:"dispute_window,omitempty"`

	// Metadata contains additional invoice details
	Metadata map[string]string `json:"metadata,omitempty"`

	// BlockHeight is when the invoice was created
	BlockHeight int64 `json:"block_height"`

	// CreatedAt is when the record was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the record was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// BillingPeriod defines a billing period
type BillingPeriod struct {
	// StartTime is the start of the billing period
	StartTime time.Time `json:"start_time"`

	// EndTime is the end of the billing period
	EndTime time.Time `json:"end_time"`

	// DurationSeconds is the billing period duration
	DurationSeconds int64 `json:"duration_seconds"`

	// PeriodType is the type of billing period
	PeriodType BillingPeriodType `json:"period_type"`

	// SequenceNumber is the period sequence (e.g., month 1, 2, 3...)
	SequenceNumber uint32 `json:"sequence_number"`
}

// BillingPeriodType defines the type of billing period
type BillingPeriodType uint8

const (
	// BillingPeriodTypeHourly for hourly billing
	BillingPeriodTypeHourly BillingPeriodType = 0

	// BillingPeriodTypeDaily for daily billing
	BillingPeriodTypeDaily BillingPeriodType = 1

	// BillingPeriodTypeWeekly for weekly billing
	BillingPeriodTypeWeekly BillingPeriodType = 2

	// BillingPeriodTypeMonthly for monthly billing
	BillingPeriodTypeMonthly BillingPeriodType = 3

	// BillingPeriodTypeUsageBased for usage-based billing
	BillingPeriodTypeUsageBased BillingPeriodType = 4

	// BillingPeriodTypeFinal for final settlement
	BillingPeriodTypeFinal BillingPeriodType = 5
)

// BillingPeriodTypeNames maps types to names
var BillingPeriodTypeNames = map[BillingPeriodType]string{
	BillingPeriodTypeHourly:     "hourly",
	BillingPeriodTypeDaily:      "daily",
	BillingPeriodTypeWeekly:     "weekly",
	BillingPeriodTypeMonthly:    "monthly",
	BillingPeriodTypeUsageBased: "usage_based",
	BillingPeriodTypeFinal:      "final",
}

// String returns string representation
func (t BillingPeriodType) String() string {
	if name, ok := BillingPeriodTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// LineItem represents a single charge on an invoice
type LineItem struct {
	// LineItemID is the unique identifier
	LineItemID string `json:"line_item_id"`

	// Description describes the charge
	Description string `json:"description"`

	// UsageType is the type of resource usage
	UsageType UsageType `json:"usage_type"`

	// Quantity is the quantity consumed
	Quantity sdkmath.LegacyDec `json:"quantity"`

	// Unit is the measurement unit
	Unit string `json:"unit"`

	// UnitPrice is the price per unit
	UnitPrice sdk.DecCoin `json:"unit_price"`

	// Amount is the line total (quantity * unit_price)
	Amount sdk.Coins `json:"amount"`

	// UsageRecordIDs links to usage records
	UsageRecordIDs []string `json:"usage_record_ids,omitempty"`

	// PricingTier is the pricing tier applied
	PricingTier string `json:"pricing_tier,omitempty"`

	// DiscountApplied is any line-level discount
	DiscountApplied sdk.Coins `json:"discount_applied,omitempty"`

	// Metadata contains additional line item details
	Metadata map[string]string `json:"metadata,omitempty"`
}

// UsageType defines the type of resource usage
type UsageType uint8

const (
	// UsageTypeCPU for CPU usage
	UsageTypeCPU UsageType = 0

	// UsageTypeMemory for memory usage
	UsageTypeMemory UsageType = 1

	// UsageTypeStorage for storage usage
	UsageTypeStorage UsageType = 2

	// UsageTypeNetwork for network bandwidth
	UsageTypeNetwork UsageType = 3

	// UsageTypeGPU for GPU usage
	UsageTypeGPU UsageType = 4

	// UsageTypeFixed for fixed/flat charges
	UsageTypeFixed UsageType = 5

	// UsageTypeSetup for one-time setup fees
	UsageTypeSetup UsageType = 6

	// UsageTypeOther for other charges
	UsageTypeOther UsageType = 7
)

// UsageTypeNames maps types to names
var UsageTypeNames = map[UsageType]string{
	UsageTypeCPU:     "cpu",
	UsageTypeMemory:  "memory",
	UsageTypeStorage: "storage",
	UsageTypeNetwork: "network",
	UsageTypeGPU:     "gpu",
	UsageTypeFixed:   "fixed",
	UsageTypeSetup:   "setup",
	UsageTypeOther:   "other",
}

// String returns string representation
func (t UsageType) String() string {
	if name, ok := UsageTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// UnitForUsageType returns the default unit for a usage type
func UnitForUsageType(t UsageType) string {
	switch t {
	case UsageTypeCPU:
		return "core-hour"
	case UsageTypeMemory:
		return "gb-hour"
	case UsageTypeStorage:
		return "gb-month"
	case UsageTypeNetwork:
		return "gb"
	case UsageTypeGPU:
		return "gpu-hour"
	case UsageTypeFixed:
		return "unit"
	case UsageTypeSetup:
		return "unit"
	default:
		return "unit"
	}
}

// NewInvoice creates a new invoice
func NewInvoice(
	invoiceID string,
	invoiceNumber string,
	escrowID string,
	orderID string,
	leaseID string,
	provider string,
	customer string,
	currency string,
	billingPeriod BillingPeriod,
	dueDate time.Time,
	blockHeight int64,
	now time.Time,
) *Invoice {
	return &Invoice{
		InvoiceID:     invoiceID,
		InvoiceNumber: invoiceNumber,
		EscrowID:      escrowID,
		OrderID:       orderID,
		LeaseID:       leaseID,
		Provider:      provider,
		Customer:      customer,
		Status:        InvoiceStatusDraft,
		BillingPeriod: billingPeriod,
		LineItems:     make([]LineItem, 0),
		Subtotal:      sdk.NewCoins(),
		Discounts:     make([]AppliedDiscount, 0),
		DiscountTotal: sdk.NewCoins(),
		TaxTotal:      sdk.NewCoins(),
		Total:         sdk.NewCoins(),
		Currency:      currency,
		AmountPaid:    sdk.NewCoins(),
		AmountDue:     sdk.NewCoins(),
		DueDate:       dueDate,
		IssuedAt:      now,
		BlockHeight:   blockHeight,
		CreatedAt:     now,
		UpdatedAt:     now,
		Metadata:      make(map[string]string),
	}
}

// AddLineItem adds a line item to the invoice
func (inv *Invoice) AddLineItem(item LineItem) {
	inv.LineItems = append(inv.LineItems, item)
	inv.recalculateTotals()
}

// recalculateTotals recalculates invoice totals
func (inv *Invoice) recalculateTotals() {
	// Calculate subtotal from line items
	subtotal := sdk.NewCoins()
	for _, item := range inv.LineItems {
		subtotal = subtotal.Add(item.Amount...)
	}
	inv.Subtotal = subtotal

	// Calculate discount total
	discountTotal := sdk.NewCoins()
	for _, d := range inv.Discounts {
		discountTotal = discountTotal.Add(d.Amount...)
	}
	inv.DiscountTotal = discountTotal

	// Calculate tax total
	taxTotal := sdk.NewCoins()
	if inv.TaxDetails != nil {
		taxTotal = inv.TaxDetails.TotalTax
	}
	inv.TaxTotal = taxTotal

	// Calculate total: subtotal - discounts + tax
	total := subtotal.Sub(discountTotal...).Add(taxTotal...)
	inv.Total = total

	// Calculate amount due
	inv.AmountDue = total.Sub(inv.AmountPaid...)
}

// Finalize transitions invoice from draft to pending
func (inv *Invoice) Finalize(now time.Time) error {
	if inv.Status != InvoiceStatusDraft {
		return fmt.Errorf("can only finalize draft invoices, current status: %s", inv.Status)
	}

	if len(inv.LineItems) == 0 {
		return fmt.Errorf("cannot finalize invoice with no line items")
	}

	inv.recalculateTotals()
	inv.Status = InvoiceStatusPending
	inv.UpdatedAt = now
	return nil
}

// RecordPayment records a payment against the invoice
func (inv *Invoice) RecordPayment(amount sdk.Coins, now time.Time) error {
	if inv.Status.IsTerminal() {
		return fmt.Errorf("cannot record payment on %s invoice", inv.Status)
	}

	inv.AmountPaid = inv.AmountPaid.Add(amount...)
	inv.AmountDue = inv.Total.Sub(inv.AmountPaid...)

	// Check if fully paid
	if inv.AmountDue.IsZero() || inv.AmountPaid.IsAllGTE(inv.Total) {
		inv.Status = InvoiceStatusPaid
		inv.PaidAt = &now
	} else if inv.AmountPaid.IsAllPositive() {
		inv.Status = InvoiceStatusPartiallyPaid
	}

	inv.UpdatedAt = now
	return nil
}

// MarkDisputed marks the invoice as disputed
func (inv *Invoice) MarkDisputed(disputeWindow *DisputeWindow, now time.Time) error {
	if inv.Status.IsTerminal() {
		return fmt.Errorf("cannot dispute %s invoice", inv.Status)
	}

	inv.Status = InvoiceStatusDisputed
	inv.DisputeWindow = disputeWindow
	inv.UpdatedAt = now
	return nil
}

// Cancel cancels the invoice
func (inv *Invoice) Cancel(now time.Time) error {
	if inv.Status == InvoiceStatusPaid || inv.Status == InvoiceStatusRefunded {
		return fmt.Errorf("cannot cancel %s invoice", inv.Status)
	}

	inv.Status = InvoiceStatusCancelled
	inv.UpdatedAt = now
	return nil
}

// Validate validates the invoice
func (inv *Invoice) Validate() error {
	if inv.InvoiceID == "" {
		return fmt.Errorf("invoice_id is required")
	}

	if len(inv.InvoiceID) > 64 {
		return fmt.Errorf("invoice_id exceeds maximum length of 64")
	}

	if inv.EscrowID == "" {
		return fmt.Errorf("escrow_id is required")
	}

	if _, err := sdk.AccAddressFromBech32(inv.Provider); err != nil {
		return fmt.Errorf("invalid provider address: %w", err)
	}

	if _, err := sdk.AccAddressFromBech32(inv.Customer); err != nil {
		return fmt.Errorf("invalid customer address: %w", err)
	}

	if inv.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	if inv.BillingPeriod.EndTime.Before(inv.BillingPeriod.StartTime) {
		return fmt.Errorf("billing period end_time must be after start_time")
	}

	// Validate line items
	for i, item := range inv.LineItems {
		if item.LineItemID == "" {
			return fmt.Errorf("line_item[%d]: line_item_id is required", i)
		}
		if item.Description == "" {
			return fmt.Errorf("line_item[%d]: description is required", i)
		}
		if item.Quantity.IsNegative() {
			return fmt.Errorf("line_item[%d]: quantity cannot be negative", i)
		}
	}

	// Validate totals consistency
	if !inv.Total.IsValid() {
		return fmt.Errorf("total must be valid coins")
	}

	return nil
}

// MarshalJSON implements json.Marshaler
func (inv Invoice) MarshalJSON() ([]byte, error) {
	type Alias Invoice
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(inv),
	})
}

// InvoiceSummary provides a summary view of an invoice
type InvoiceSummary struct {
	InvoiceID     string        `json:"invoice_id"`
	InvoiceNumber string        `json:"invoice_number"`
	Provider      string        `json:"provider"`
	Customer      string        `json:"customer"`
	Status        InvoiceStatus `json:"status"`
	Total         sdk.Coins     `json:"total"`
	AmountPaid    sdk.Coins     `json:"amount_paid"`
	AmountDue     sdk.Coins     `json:"amount_due"`
	DueDate       time.Time     `json:"due_date"`
	IssuedAt      time.Time     `json:"issued_at"`
}

// ToSummary creates a summary from the invoice
func (inv *Invoice) ToSummary() InvoiceSummary {
	return InvoiceSummary{
		InvoiceID:     inv.InvoiceID,
		InvoiceNumber: inv.InvoiceNumber,
		Provider:      inv.Provider,
		Customer:      inv.Customer,
		Status:        inv.Status,
		Total:         inv.Total,
		AmountPaid:    inv.AmountPaid,
		AmountDue:     inv.AmountDue,
		DueDate:       inv.DueDate,
		IssuedAt:      inv.IssuedAt,
	}
}
