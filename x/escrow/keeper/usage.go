// Package keeper provides the escrow module keeper with usage report handling
// and the usage→invoice→settlement pipeline.
package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/escrow/types/billing"
)

// UsagePipelineKeeper defines the interface for the usage→invoice→settlement pipeline
type UsagePipelineKeeper interface {
	// SubmitUsageReport validates and stores a usage report from a provider
	SubmitUsageReport(ctx sdk.Context, report *UsageReport) (*billing.UsageRecord, error)

	// GenerateInvoiceFromUsage creates an invoice from accumulated usage records
	GenerateInvoiceFromUsage(ctx sdk.Context, leaseID string, periodEnd time.Time) (*billing.InvoiceLedgerRecord, error)

	// ProcessUsageSettlement runs the full usage→invoice→settlement pipeline for a lease
	ProcessUsageSettlement(ctx sdk.Context, leaseID string) (*UsageSettlementResult, error)

	// GetPendingUsageRecords returns all pending usage records for a lease
	GetPendingUsageRecords(ctx sdk.Context, leaseID string) ([]*billing.UsageRecord, error)

	// ApproveInvoice approves an invoice for settlement
	ApproveInvoice(ctx sdk.Context, invoiceID string, approver string) error

	// DisputeInvoice initiates a dispute on an invoice
	DisputeInvoice(ctx sdk.Context, invoiceID string, disputant string, reason string) error
}

// UsageReport represents a usage report submitted by a provider daemon
type UsageReport struct {
	// Provider is the address of the provider submitting the report
	Provider string `json:"provider"`

	// LeaseID links to the marketplace lease
	LeaseID string `json:"lease_id"`

	// Customer is the customer address
	Customer string `json:"customer"`

	// EscrowID is the escrow account ID
	EscrowID string `json:"escrow_id"`

	// Resources contains the usage metrics for each resource type
	Resources []ResourceUsage `json:"resources"`

	// PeriodStart is when the usage period started
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is when the usage period ended
	PeriodEnd time.Time `json:"period_end"`

	// Signature is the provider's signature on the report data
	Signature string `json:"signature,omitempty"`
}

// ResourceUsage describes usage for a single resource type
type ResourceUsage struct {
	// Type is the usage type (CPU, Memory, Storage, etc.)
	Type billing.UsageType `json:"type"`

	// Quantity is the measured usage
	Quantity sdkmath.LegacyDec `json:"quantity"`

	// Unit is the measurement unit
	Unit string `json:"unit"`

	// UnitPrice is the per-unit price
	UnitPrice sdk.DecCoin `json:"unit_price"`
}

// UsageSettlementResult contains the outcome of a usage settlement pipeline run
type UsageSettlementResult struct {
	// UsageRecords is the list of usage records consumed
	UsageRecordIDs []string `json:"usage_record_ids"`

	// InvoiceID is the invoice generated from the usage
	InvoiceID string `json:"invoice_id"`

	// SettlementID is the settlement record ID (if settled)
	SettlementID string `json:"settlement_id,omitempty"`

	// TotalAmount is the total settled amount
	TotalAmount sdk.Coins `json:"total_amount"`

	// Status describes the pipeline outcome
	Status string `json:"status"`
}

// Validate validates the usage report
func (r *UsageReport) Validate() error {
	if r.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if _, err := sdk.AccAddressFromBech32(r.Provider); err != nil {
		return fmt.Errorf("invalid provider address: %w", err)
	}
	if r.LeaseID == "" {
		return fmt.Errorf("lease_id is required")
	}
	if r.Customer == "" {
		return fmt.Errorf("customer is required")
	}
	if _, err := sdk.AccAddressFromBech32(r.Customer); err != nil {
		return fmt.Errorf("invalid customer address: %w", err)
	}
	if r.EscrowID == "" {
		return fmt.Errorf("escrow_id is required")
	}
	if len(r.Resources) == 0 {
		return fmt.Errorf("at least one resource usage entry is required")
	}
	if r.PeriodEnd.Before(r.PeriodStart) {
		return fmt.Errorf("period_end must be after period_start")
	}
	if r.PeriodEnd.Equal(r.PeriodStart) {
		return fmt.Errorf("period_end must be after period_start")
	}
	for i, res := range r.Resources {
		if res.Quantity.IsNegative() {
			return fmt.Errorf("resource[%d] quantity cannot be negative", i)
		}
		if res.Unit == "" {
			return fmt.Errorf("resource[%d] unit is required", i)
		}
		if res.UnitPrice.IsNegative() {
			return fmt.Errorf("resource[%d] unit_price cannot be negative", i)
		}
	}
	return nil
}

// usagePipelineKeeper implements UsagePipelineKeeper
type usagePipelineKeeper struct {
	k *keeper
}

// NewUsagePipelineKeeper creates a new usage pipeline keeper from the base keeper
func (k *keeper) NewUsagePipelineKeeper() UsagePipelineKeeper {
	return &usagePipelineKeeper{k: k}
}

// SubmitUsageReport validates and stores a usage report from a provider
func (uk *usagePipelineKeeper) SubmitUsageReport(ctx sdk.Context, report *UsageReport) (*billing.UsageRecord, error) {
	if report == nil {
		return nil, fmt.Errorf("usage report is nil")
	}

	if err := report.Validate(); err != nil {
		return nil, fmt.Errorf("invalid usage report: %w", err)
	}

	now := ctx.BlockTime()
	height := ctx.BlockHeight()

	// Calculate total amount from resources
	totalAmount := sdk.NewCoins()
	for _, res := range report.Resources {
		if res.UnitPrice.Amount.IsZero() {
			continue
		}
		amount := res.Quantity.Mul(res.UnitPrice.Amount).TruncateInt()
		if amount.IsPositive() {
			totalAmount = totalAmount.Add(sdk.NewCoin(res.UnitPrice.Denom, amount))
		}
	}

	// Generate record ID from lease + block height
	recordID := fmt.Sprintf("usage-%s-%d", report.LeaseID, height)

	// Use the first resource type as primary type
	resourceType := billing.UsageTypeCPU
	if len(report.Resources) > 0 {
		resourceType = report.Resources[0].Type
	}

	// Build aggregate quantity and unit price
	aggregateQuantity := sdkmath.LegacyZeroDec()
	var primaryUnitPrice sdk.DecCoin
	for _, res := range report.Resources {
		aggregateQuantity = aggregateQuantity.Add(res.Quantity)
		if primaryUnitPrice.IsZero() {
			primaryUnitPrice = res.UnitPrice
		}
	}

	record := &billing.UsageRecord{
		RecordID:     recordID,
		LeaseID:      report.LeaseID,
		Provider:     report.Provider,
		Customer:     report.Customer,
		StartTime:    report.PeriodStart,
		EndTime:      report.PeriodEnd,
		ResourceType: resourceType,
		UsageAmount:  aggregateQuantity,
		UnitPrice:    primaryUnitPrice,
		TotalAmount:  totalAmount,
		Status:       billing.UsageRecordStatusPending,
		BlockHeight:  height,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := record.Validate(); err != nil {
		return nil, fmt.Errorf("invalid usage record: %w", err)
	}

	// Persist via the reconciliation keeper
	rk := uk.k.NewReconciliationKeeper()
	if err := rk.SaveUsageRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to save usage record: %w", err)
	}

	// Also store the detailed resource breakdown
	uk.saveResourceBreakdown(ctx, recordID, report.Resources)

	return record, nil
}

// GenerateInvoiceFromUsage creates an invoice from accumulated usage records for a lease
func (uk *usagePipelineKeeper) GenerateInvoiceFromUsage(ctx sdk.Context, leaseID string, periodEnd time.Time) (*billing.InvoiceLedgerRecord, error) {
	if leaseID == "" {
		return nil, fmt.Errorf("lease_id is required")
	}

	rk := uk.k.NewReconciliationKeeper()
	records, err := rk.GetUsageRecordsByLease(ctx, leaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage records: %w", err)
	}

	// Filter to pending records within the period
	var pendingRecords []*billing.UsageRecord
	for _, r := range records {
		if r.Status == billing.UsageRecordStatusPending && !r.EndTime.After(periodEnd) {
			pendingRecords = append(pendingRecords, r)
		}
	}

	if len(pendingRecords) == 0 {
		return nil, fmt.Errorf("no pending usage records found for lease %s", leaseID)
	}

	// Aggregate usage into line items
	first := pendingRecords[0]
	now := ctx.BlockTime()

	// Calculate period from records
	periodStart := first.StartTime
	for _, r := range pendingRecords {
		if r.StartTime.Before(periodStart) {
			periodStart = r.StartTime
		}
	}

	// Build usage inputs from records
	var usageInputs []billing.UsageInput
	totalAmount := sdk.NewCoins()

	for _, r := range pendingRecords {
		usageInputs = append(usageInputs, billing.UsageInput{
			UsageRecordID: r.RecordID,
			UsageType:     r.ResourceType,
			Quantity:      r.UsageAmount,
			Unit:          billing.UnitForUsageType(r.ResourceType),
			UnitPrice:     r.UnitPrice,
			Description:   fmt.Sprintf("%s usage for %s", r.ResourceType.String(), leaseID),
			PeriodStart:   r.StartTime,
			PeriodEnd:     r.EndTime,
		})
		totalAmount = totalAmount.Add(r.TotalAmount...)
	}

	// Determine currency from first record
	currency := billing.DefaultCurrency
	if len(first.TotalAmount) > 0 {
		currency = first.TotalAmount[0].Denom
	}

	// Get next invoice sequence
	ik := uk.k.NewInvoiceKeeper()
	seq := ik.GetInvoiceSequence(ctx)
	invoiceNumber := fmt.Sprintf("VE-USG-%d", seq+1)

	// Generate invoice
	durationSeconds := int64(periodEnd.Sub(periodStart).Seconds())
	generator := billing.NewInvoiceGenerator(billing.DefaultInvoiceGeneratorConfig())

	req := billing.InvoiceGenerationRequest{
		EscrowID:    first.Provider, // Provider as escrow key when no explicit escrow
		OrderID:     "",
		LeaseID:     leaseID,
		Provider:    first.Provider,
		Customer:    first.Customer,
		UsageInputs: usageInputs,
		BillingPeriod: billing.BillingPeriod{
			StartTime:       periodStart,
			EndTime:         periodEnd,
			DurationSeconds: durationSeconds,
			PeriodType:      billing.BillingPeriodTypeMonthly,
		},
		Currency: currency,
	}

	invoice, err := generator.GenerateInvoice(req, ctx.BlockHeight(), now)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice: %w", err)
	}

	invoice.InvoiceNumber = invoiceNumber
	invoice.Status = billing.InvoiceStatusPending
	invoice.AmountDue = totalAmount

	artifactCID := fmt.Sprintf("invoice-%s", invoice.InvoiceID)

	record, err := ik.CreateInvoice(ctx, invoice, artifactCID)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Mark usage records as invoiced
	for _, r := range pendingRecords {
		r.Status = billing.UsageRecordStatusInvoiced
		r.InvoiceID = invoice.InvoiceID
		r.UpdatedAt = now
		if err := rk.SaveUsageRecord(ctx, r); err != nil {
			// Log but don't fail - invoice is already created
			continue
		}
	}

	return record, nil
}

// ProcessUsageSettlement runs the full usage→invoice→settlement pipeline for a lease
func (uk *usagePipelineKeeper) ProcessUsageSettlement(ctx sdk.Context, leaseID string) (*UsageSettlementResult, error) {
	if leaseID == "" {
		return nil, fmt.Errorf("lease_id is required")
	}

	now := ctx.BlockTime()

	// Step 1: Generate invoice from pending usage
	invoiceRecord, err := uk.GenerateInvoiceFromUsage(ctx, leaseID, now)
	if err != nil {
		return nil, fmt.Errorf("invoice generation failed: %w", err)
	}

	// Collect usage record IDs that were consumed
	rk := uk.k.NewReconciliationKeeper()
	records, err := rk.GetUsageRecordsByLease(ctx, leaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage records: %w", err)
	}

	var usageRecordIDs []string
	for _, r := range records {
		if r.InvoiceID == invoiceRecord.InvoiceID {
			usageRecordIDs = append(usageRecordIDs, r.RecordID)
		}
	}

	// Step 2: Settle the invoice via the settlement integration keeper
	sik := uk.k.NewSettlementIntegrationKeeper()
	settlement, err := sik.SettleInvoice(ctx, invoiceRecord.InvoiceID, "usage_pipeline")
	if err != nil {
		// Invoice created but settlement failed - return partial result
		return &UsageSettlementResult{
			UsageRecordIDs: usageRecordIDs,
			InvoiceID:      invoiceRecord.InvoiceID,
			TotalAmount:    invoiceRecord.Total,
			Status:         "invoice_created",
		}, nil
	}

	return &UsageSettlementResult{
		UsageRecordIDs: usageRecordIDs,
		InvoiceID:      invoiceRecord.InvoiceID,
		SettlementID:   settlement.SettlementID,
		TotalAmount:    invoiceRecord.Total,
		Status:         "settled",
	}, nil
}

// GetPendingUsageRecords returns all pending usage records for a lease
func (uk *usagePipelineKeeper) GetPendingUsageRecords(ctx sdk.Context, leaseID string) ([]*billing.UsageRecord, error) {
	rk := uk.k.NewReconciliationKeeper()
	records, err := rk.GetUsageRecordsByLease(ctx, leaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage records: %w", err)
	}

	var pending []*billing.UsageRecord
	for _, r := range records {
		if r.Status == billing.UsageRecordStatusPending {
			pending = append(pending, r)
		}
	}

	return pending, nil
}

// ApproveInvoice approves a pending invoice for settlement
func (uk *usagePipelineKeeper) ApproveInvoice(ctx sdk.Context, invoiceID string, approver string) error {
	if invoiceID == "" {
		return fmt.Errorf("invoice_id is required")
	}
	if approver == "" {
		return fmt.Errorf("approver is required")
	}

	ik := uk.k.NewInvoiceKeeper()

	// Validate invoice exists and is in pending status
	record, err := ik.GetInvoice(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	if record.Status != billing.InvoiceStatusPending {
		return fmt.Errorf("invoice %s is not in pending status (current: %s)", invoiceID, record.Status.String())
	}

	// Update status to paid via the invoice keeper
	_, err = ik.UpdateInvoiceStatus(ctx, invoiceID, billing.InvoiceStatusPaid, approver)
	if err != nil {
		return fmt.Errorf("failed to approve invoice: %w", err)
	}

	return nil
}

// DisputeInvoice initiates a dispute on an invoice
func (uk *usagePipelineKeeper) DisputeInvoice(ctx sdk.Context, invoiceID string, disputant string, reason string) error {
	if invoiceID == "" {
		return fmt.Errorf("invoice_id is required")
	}
	if disputant == "" {
		return fmt.Errorf("disputant is required")
	}
	if reason == "" {
		return fmt.Errorf("reason is required")
	}

	ik := uk.k.NewInvoiceKeeper()

	// Validate invoice exists and can be disputed
	record, err := ik.GetInvoice(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	if record.Status != billing.InvoiceStatusPending && record.Status != billing.InvoiceStatusPartiallyPaid {
		return fmt.Errorf("invoice %s cannot be disputed (current: %s)", invoiceID, record.Status.String())
	}

	// Update status to disputed
	_, err = ik.UpdateInvoiceStatus(ctx, invoiceID, billing.InvoiceStatusDisputed, disputant)
	if err != nil {
		return fmt.Errorf("failed to dispute invoice: %w", err)
	}

	// If a settlement already exists, apply holdback
	sik := uk.k.NewSettlementIntegrationKeeper()
	settlement, err := sik.GetSettlementByInvoice(ctx, invoiceID)
	if err == nil && settlement != nil {
		if err := sik.HoldbackForDispute(ctx, settlement.SettlementID, record.Total, reason); err != nil {
			// Log holdback failure but don't fail the dispute itself
			_ = err
		}
	}

	return nil
}

// resourceBreakdownKey builds a KV store key for resource breakdowns
var resourceBreakdownPrefix = []byte{0x90}

func buildResourceBreakdownKey(recordID string) []byte {
	key := make([]byte, 0, len(resourceBreakdownPrefix)+len(recordID))
	key = append(key, resourceBreakdownPrefix...)
	return append(key, []byte(recordID)...)
}

// saveResourceBreakdown stores the detailed resource breakdown for a usage record
func (uk *usagePipelineKeeper) saveResourceBreakdown(ctx sdk.Context, recordID string, resources []ResourceUsage) {
	store := ctx.KVStore(uk.k.skey)
	key := buildResourceBreakdownKey(recordID)

	bz, err := json.Marshal(resources)
	if err != nil {
		return
	}
	store.Set(key, bz)
}

// GetResourceBreakdown retrieves the detailed resource breakdown for a usage record
func (uk *usagePipelineKeeper) GetResourceBreakdown(ctx sdk.Context, recordID string) ([]ResourceUsage, error) {
	store := ctx.KVStore(uk.k.skey)
	key := buildResourceBreakdownKey(recordID)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("resource breakdown not found for record: %s", recordID)
	}

	var resources []ResourceUsage
	if err := json.Unmarshal(bz, &resources); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resource breakdown: %w", err)
	}

	return resources, nil
}

// SubmitUsageReport is a convenience method on the keeper for submitting usage reports
func (k *keeper) SubmitUsageReport(ctx sdk.Context, report *UsageReport) (*billing.UsageRecord, error) {
	return k.NewUsagePipelineKeeper().SubmitUsageReport(ctx, report)
}

// GenerateInvoiceFromUsage is a convenience method for generating invoices from usage
func (k *keeper) GenerateInvoiceFromUsage(ctx sdk.Context, leaseID string, periodEnd time.Time) (*billing.InvoiceLedgerRecord, error) {
	return k.NewUsagePipelineKeeper().GenerateInvoiceFromUsage(ctx, leaseID, periodEnd)
}

// ProcessUsageSettlement is a convenience method for the full pipeline
func (k *keeper) ProcessUsageSettlement(ctx sdk.Context, leaseID string) (*UsageSettlementResult, error) {
	return k.NewUsagePipelineKeeper().ProcessUsageSettlement(ctx, leaseID)
}
