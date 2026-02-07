// Copyright 2026 VirtEngine Authors. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the LICENSE file.

// Package billing provides billing reconciliation services for verifying
// usage reports against invoices and settlement payouts.
package billing

import (
	"context"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/escrow/types/billing"
)

// ReconciliationService provides off-chain reconciliation logic for comparing
// usage records, invoices, and payouts to detect discrepancies.
type ReconciliationService struct {
	config ReconciliationServiceConfig
}

// ReconciliationServiceConfig holds configuration for the reconciliation service
type ReconciliationServiceConfig struct {
	// VarianceThresholdPercent is the percentage threshold for flagging discrepancies (e.g. 1.0 for 1%)
	VarianceThresholdPercent float64

	// AutoResolveThresholdPercent is the max variance that can be auto-resolved (e.g. 0.1 for 0.1%)
	AutoResolveThresholdPercent float64

	// MaxDiscrepanciesBeforeFail is the maximum number of discrepancies before failing
	MaxDiscrepanciesBeforeFail int
}

// DefaultReconciliationServiceConfig returns the default configuration
func DefaultReconciliationServiceConfig() ReconciliationServiceConfig {
	return ReconciliationServiceConfig{
		VarianceThresholdPercent:    1.0,
		AutoResolveThresholdPercent: 0.1,
		MaxDiscrepanciesBeforeFail:  100,
	}
}

// NewReconciliationService creates a new reconciliation service
func NewReconciliationService(config ReconciliationServiceConfig) *ReconciliationService {
	return &ReconciliationService{config: config}
}

// ReconciliationInput contains all data needed for a reconciliation run
type ReconciliationInput struct {
	// UsageRecords are the provider-reported usage records
	UsageRecords []billing.UsageRecord

	// Invoices are the generated invoices
	Invoices []billing.InvoiceLedgerRecord

	// Payouts are the settlement payout records
	Payouts []billing.PayoutRecord

	// PeriodStart is the start of the reconciliation period
	PeriodStart time.Time

	// PeriodEnd is the end of the reconciliation period
	PeriodEnd time.Time

	// Provider filters to a specific provider (optional)
	Provider string
}

// ReconciliationOutput contains the result of a reconciliation run
type ReconciliationOutput struct {
	// Summary contains aggregated totals
	Summary ReconciliationSummary

	// Discrepancies contains all identified discrepancies
	Discrepancies []Discrepancy

	// Status is the overall reconciliation status
	Status string

	// GeneratedAt is when the report was generated
	GeneratedAt time.Time
}

// ReconciliationSummary contains aggregate totals for reconciliation
type ReconciliationSummary struct {
	TotalUsageAmount     sdk.Coins
	TotalInvoiceAmount   sdk.Coins
	TotalPayoutAmount    sdk.Coins
	UsageRecordCount     int
	InvoiceCount         int
	PayoutCount          int
	PendingUsageCount    int
	InvoicedUsageCount   int
	SettledUsageCount    int
	PendingInvoiceCount  int
	PaidInvoiceCount     int
	DisputedInvoiceCount int
	DiscrepancyCount     int
}

// Discrepancy represents a mismatch found during reconciliation
type Discrepancy struct {
	Type        DiscrepancyType
	Severity    DiscrepancySeverity
	Description string
	InvoiceID   string
	UsageID     string
	PayoutID    string
	Expected    sdk.Coins
	Actual      sdk.Coins
	Difference  sdk.Coins
}

// DiscrepancyType identifies the category of discrepancy
type DiscrepancyType string

const (
	DiscrepancyTypeUsageInvoiceMismatch  DiscrepancyType = "usage_invoice_mismatch"
	DiscrepancyTypeInvoicePayoutMismatch DiscrepancyType = "invoice_payout_mismatch"
	DiscrepancyTypeMissingInvoice        DiscrepancyType = "missing_invoice"
	DiscrepancyTypeMissingPayout         DiscrepancyType = "missing_payout"
	DiscrepancyTypeDuplicateUsage        DiscrepancyType = "duplicate_usage"
	DiscrepancyTypeOverpayment           DiscrepancyType = "overpayment"
	DiscrepancyTypeUnderpayment          DiscrepancyType = "underpayment"
)

// DiscrepancySeverity indicates the impact level
type DiscrepancySeverity string

const (
	DiscrepancySeverityLow      DiscrepancySeverity = "low"
	DiscrepancySeverityMedium   DiscrepancySeverity = "medium"
	DiscrepancySeverityHigh     DiscrepancySeverity = "high"
	DiscrepancySeverityCritical DiscrepancySeverity = "critical"
)

// Reconcile performs reconciliation between usage records, invoices, and payouts
func (s *ReconciliationService) Reconcile(_ context.Context, input ReconciliationInput) (*ReconciliationOutput, error) {
	if err := s.validateInput(input); err != nil {
		return nil, fmt.Errorf("invalid reconciliation input: %w", err)
	}

	output := &ReconciliationOutput{
		GeneratedAt: time.Now().UTC(),
	}

	// Build indexes for fast lookups
	invoiceByID := make(map[string]*billing.InvoiceLedgerRecord)
	for i := range input.Invoices {
		invoiceByID[input.Invoices[i].InvoiceID] = &input.Invoices[i]
	}

	payoutByInvoice := make(map[string]*billing.PayoutRecord)
	for i := range input.Payouts {
		for _, invID := range input.Payouts[i].InvoiceIDs {
			payoutByInvoice[invID] = &input.Payouts[i]
		}
	}

	// Compute summary totals
	summary := s.computeSummary(input)

	// Step 1: Check usage records have corresponding invoices
	usageDiscrepancies := s.reconcileUsageToInvoice(input.UsageRecords, invoiceByID)
	output.Discrepancies = append(output.Discrepancies, usageDiscrepancies...)

	// Step 2: Check invoices have corresponding payouts
	invoiceDiscrepancies := s.reconcileInvoiceToPayout(input.Invoices, payoutByInvoice)
	output.Discrepancies = append(output.Discrepancies, invoiceDiscrepancies...)

	// Step 3: Check for amount mismatches between usage totals and invoice totals
	amountDiscrepancies := s.reconcileAmounts(input.UsageRecords, input.Invoices)
	output.Discrepancies = append(output.Discrepancies, amountDiscrepancies...)

	summary.DiscrepancyCount = len(output.Discrepancies)
	output.Summary = summary

	// Determine overall status
	if len(output.Discrepancies) == 0 {
		output.Status = "reconciled"
	} else if len(output.Discrepancies) > s.config.MaxDiscrepanciesBeforeFail {
		output.Status = "failed"
	} else {
		output.Status = "discrepancies_found"
	}

	return output, nil
}

func (s *ReconciliationService) validateInput(input ReconciliationInput) error {
	if input.PeriodEnd.Before(input.PeriodStart) {
		return fmt.Errorf("period_end must be after period_start")
	}
	return nil
}

func (s *ReconciliationService) computeSummary(input ReconciliationInput) ReconciliationSummary {
	summary := ReconciliationSummary{
		TotalUsageAmount:   sdk.NewCoins(),
		TotalInvoiceAmount: sdk.NewCoins(),
		TotalPayoutAmount:  sdk.NewCoins(),
		UsageRecordCount:   len(input.UsageRecords),
		InvoiceCount:       len(input.Invoices),
		PayoutCount:        len(input.Payouts),
	}

	for _, u := range input.UsageRecords {
		summary.TotalUsageAmount = summary.TotalUsageAmount.Add(u.TotalAmount...)
		switch u.Status {
		case billing.UsageRecordStatusPending:
			summary.PendingUsageCount++
		case billing.UsageRecordStatusInvoiced:
			summary.InvoicedUsageCount++
		case billing.UsageRecordStatusSettled:
			summary.SettledUsageCount++
		}
	}

	for _, inv := range input.Invoices {
		summary.TotalInvoiceAmount = summary.TotalInvoiceAmount.Add(inv.Total...)
		switch inv.Status {
		case billing.InvoiceStatusPending:
			summary.PendingInvoiceCount++
		case billing.InvoiceStatusPaid:
			summary.PaidInvoiceCount++
		case billing.InvoiceStatusDisputed:
			summary.DisputedInvoiceCount++
		}
	}

	for _, p := range input.Payouts {
		summary.TotalPayoutAmount = summary.TotalPayoutAmount.Add(p.NetAmount...)
	}

	return summary
}

func (s *ReconciliationService) reconcileUsageToInvoice(
	usageRecords []billing.UsageRecord,
	invoiceByID map[string]*billing.InvoiceLedgerRecord,
) []Discrepancy {
	var discrepancies []Discrepancy

	for _, usage := range usageRecords {
		if usage.Status == billing.UsageRecordStatusPending {
			continue // Pending records not yet invoiced
		}

		if usage.InvoiceID == "" {
			discrepancies = append(discrepancies, Discrepancy{
				Type:        DiscrepancyTypeMissingInvoice,
				Severity:    DiscrepancySeverityHigh,
				Description: fmt.Sprintf("usage record %s (status: %s) has no linked invoice", usage.RecordID, usage.Status),
				UsageID:     usage.RecordID,
				Expected:    usage.TotalAmount,
				Actual:      sdk.NewCoins(),
				Difference:  usage.TotalAmount,
			})
			continue
		}

		inv, exists := invoiceByID[usage.InvoiceID]
		if !exists {
			discrepancies = append(discrepancies, Discrepancy{
				Type:        DiscrepancyTypeMissingInvoice,
				Severity:    DiscrepancySeverityHigh,
				Description: fmt.Sprintf("usage record %s links to invoice %s which does not exist", usage.RecordID, usage.InvoiceID),
				UsageID:     usage.RecordID,
				InvoiceID:   usage.InvoiceID,
				Expected:    usage.TotalAmount,
				Actual:      sdk.NewCoins(),
				Difference:  usage.TotalAmount,
			})
			continue
		}

		// Check that usage provider matches invoice provider
		if usage.Provider != inv.Provider {
			discrepancies = append(discrepancies, Discrepancy{
				Type:        DiscrepancyTypeUsageInvoiceMismatch,
				Severity:    DiscrepancySeverityCritical,
				Description: fmt.Sprintf("usage record %s provider (%s) does not match invoice %s provider (%s)", usage.RecordID, usage.Provider, inv.InvoiceID, inv.Provider),
				UsageID:     usage.RecordID,
				InvoiceID:   inv.InvoiceID,
			})
		}
	}

	return discrepancies
}

func (s *ReconciliationService) reconcileInvoiceToPayout(
	invoices []billing.InvoiceLedgerRecord,
	payoutByInvoice map[string]*billing.PayoutRecord,
) []Discrepancy {
	var discrepancies []Discrepancy

	for _, inv := range invoices {
		if inv.Status != billing.InvoiceStatusPaid {
			continue // Only check paid invoices for payout
		}

		payout, exists := payoutByInvoice[inv.InvoiceID]
		if !exists {
			discrepancies = append(discrepancies, Discrepancy{
				Type:        DiscrepancyTypeMissingPayout,
				Severity:    DiscrepancySeverityMedium,
				Description: fmt.Sprintf("paid invoice %s has no matching payout record", inv.InvoiceID),
				InvoiceID:   inv.InvoiceID,
				Expected:    inv.Total,
				Actual:      sdk.NewCoins(),
				Difference:  inv.Total,
			})
			continue
		}

		// Check payout amount vs invoice amount (accounting for fees)
		if payout.GrossAmount.IsAllGT(inv.Total) {
			discrepancies = append(discrepancies, Discrepancy{
				Type:        DiscrepancyTypeOverpayment,
				Severity:    DiscrepancySeverityHigh,
				Description: fmt.Sprintf("payout %s gross amount exceeds invoice %s total", payout.PayoutID, inv.InvoiceID),
				InvoiceID:   inv.InvoiceID,
				PayoutID:    payout.PayoutID,
				Expected:    inv.Total,
				Actual:      payout.GrossAmount,
			})
		}
	}

	return discrepancies
}

func (s *ReconciliationService) reconcileAmounts(
	usageRecords []billing.UsageRecord,
	invoices []billing.InvoiceLedgerRecord,
) []Discrepancy {
	var discrepancies []Discrepancy

	// Group usage amounts by invoice
	usageByInvoice := make(map[string]sdk.Coins)
	for _, u := range usageRecords {
		if u.InvoiceID != "" {
			existing := usageByInvoice[u.InvoiceID]
			usageByInvoice[u.InvoiceID] = existing.Add(u.TotalAmount...)
		}
	}

	// Compare aggregated usage amounts to invoice amounts
	threshold := sdkmath.LegacyNewDecWithPrec(int64(s.config.VarianceThresholdPercent*10), 3)

	for _, inv := range invoices {
		usageTotal, exists := usageByInvoice[inv.InvoiceID]
		if !exists {
			continue // No usage records linked to this invoice
		}

		// Check if the usage total matches the invoice subtotal within threshold
		for _, coin := range inv.Subtotal {
			usageCoin := sdk.NewCoin(coin.Denom, sdkmath.ZeroInt())
			for _, uc := range usageTotal {
				if uc.Denom == coin.Denom {
					usageCoin = uc
					break
				}
			}

			if coin.Amount.IsZero() {
				continue
			}

			diff := coin.Amount.Sub(usageCoin.Amount).Abs()
			variance := sdkmath.LegacyNewDecFromInt(diff).Quo(sdkmath.LegacyNewDecFromInt(coin.Amount))

			if variance.GT(threshold) {
				severity := DiscrepancySeverityLow
				if variance.GT(sdkmath.LegacyNewDecWithPrec(50, 3)) {
					severity = DiscrepancySeverityHigh
				} else if variance.GT(sdkmath.LegacyNewDecWithPrec(10, 3)) {
					severity = DiscrepancySeverityMedium
				}

				discrepancies = append(discrepancies, Discrepancy{
					Type:     DiscrepancyTypeUsageInvoiceMismatch,
					Severity: severity,
					Description: fmt.Sprintf(
						"invoice %s subtotal (%s %s) differs from usage total (%s %s) by %.2f%%",
						inv.InvoiceID, coin.Amount.String(), coin.Denom,
						usageCoin.Amount.String(), usageCoin.Denom,
						variance.MustFloat64()*100,
					),
					InvoiceID:  inv.InvoiceID,
					Expected:   sdk.NewCoins(coin),
					Actual:     sdk.NewCoins(usageCoin),
					Difference: sdk.NewCoins(sdk.NewCoin(coin.Denom, diff)),
				})
			}
		}
	}

	return discrepancies
}
