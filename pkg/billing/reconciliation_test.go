// Copyright 2026 VirtEngine Authors. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the LICENSE file.

package billing

import (
	"context"
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/escrow/types/billing"
)

type ReconciliationTestSuite struct {
	suite.Suite
	svc *ReconciliationService
}

func TestReconciliationTestSuite(t *testing.T) {
	suite.Run(t, new(ReconciliationTestSuite))
}

func (s *ReconciliationTestSuite) SetupTest() {
	s.svc = NewReconciliationService(DefaultReconciliationServiceConfig())
}

// TestReconcileEmptyInput tests reconciliation with no records
func (s *ReconciliationTestSuite) TestReconcileEmptyInput() {
	input := ReconciliationInput{
		PeriodStart: time.Now().Add(-24 * time.Hour),
		PeriodEnd:   time.Now(),
	}

	output, err := s.svc.Reconcile(context.Background(), input)
	s.Require().NoError(err)
	s.Require().Equal("reconciled", output.Status)
	s.Require().Empty(output.Discrepancies)
	s.Require().Equal(0, output.Summary.UsageRecordCount)
	s.Require().Equal(0, output.Summary.InvoiceCount)
	s.Require().Equal(0, output.Summary.PayoutCount)
}

// TestReconcileInvalidPeriod tests that invalid period is rejected
func (s *ReconciliationTestSuite) TestReconcileInvalidPeriod() {
	input := ReconciliationInput{
		PeriodStart: time.Now(),
		PeriodEnd:   time.Now().Add(-24 * time.Hour),
	}

	_, err := s.svc.Reconcile(context.Background(), input)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "period_end must be after period_start")
}

// TestReconcileMatchedRecords tests reconciliation with fully matched records
func (s *ReconciliationTestSuite) TestReconcileMatchedRecords() {
	now := time.Now()
	totalAmount := sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000))

	input := ReconciliationInput{
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		UsageRecords: []billing.UsageRecord{
			{
				RecordID:    "usage-001",
				LeaseID:     "lease-001",
				Provider:    "provider1",
				Customer:    "customer1",
				Status:      billing.UsageRecordStatusInvoiced,
				InvoiceID:   "inv-001",
				TotalAmount: totalAmount,
			},
		},
		Invoices: []billing.InvoiceLedgerRecord{
			{
				InvoiceID: "inv-001",
				Provider:  "provider1",
				Customer:  "customer1",
				Status:    billing.InvoiceStatusPaid,
				Subtotal:  totalAmount,
				Total:     totalAmount,
			},
		},
		Payouts: []billing.PayoutRecord{
			{
				PayoutID:    "payout-001",
				Provider:    "provider1",
				GrossAmount: totalAmount,
				NetAmount:   sdk.NewCoins(sdk.NewInt64Coin("uvirt", 920)),
				InvoiceIDs:  []string{"inv-001"},
				Status:      billing.PayoutStatusCompleted,
			},
		},
	}

	output, err := s.svc.Reconcile(context.Background(), input)
	s.Require().NoError(err)
	s.Require().Equal("reconciled", output.Status)
	s.Require().Empty(output.Discrepancies)

	// Verify summary
	s.Require().Equal(1, output.Summary.UsageRecordCount)
	s.Require().Equal(1, output.Summary.InvoiceCount)
	s.Require().Equal(1, output.Summary.PayoutCount)
	s.Require().Equal(1, output.Summary.InvoicedUsageCount)
	s.Require().Equal(1, output.Summary.PaidInvoiceCount)
}

// TestReconcileMissingInvoice tests detection of usage records without invoices
func (s *ReconciliationTestSuite) TestReconcileMissingInvoice() {
	now := time.Now()

	input := ReconciliationInput{
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		UsageRecords: []billing.UsageRecord{
			{
				RecordID:    "usage-001",
				LeaseID:     "lease-001",
				Provider:    "provider1",
				Customer:    "customer1",
				Status:      billing.UsageRecordStatusInvoiced,
				InvoiceID:   "inv-999", // Non-existent invoice
				TotalAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
			},
		},
	}

	output, err := s.svc.Reconcile(context.Background(), input)
	s.Require().NoError(err)
	s.Require().Equal("discrepancies_found", output.Status)
	s.Require().Len(output.Discrepancies, 1)
	s.Require().Equal(DiscrepancyTypeMissingInvoice, output.Discrepancies[0].Type)
	s.Require().Equal(DiscrepancySeverityHigh, output.Discrepancies[0].Severity)
}

// TestReconcileMissingPayout tests detection of paid invoices without payouts
func (s *ReconciliationTestSuite) TestReconcileMissingPayout() {
	now := time.Now()
	totalAmount := sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000))

	input := ReconciliationInput{
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		Invoices: []billing.InvoiceLedgerRecord{
			{
				InvoiceID: "inv-001",
				Provider:  "provider1",
				Customer:  "customer1",
				Status:    billing.InvoiceStatusPaid,
				Total:     totalAmount,
			},
		},
		// No payouts
	}

	output, err := s.svc.Reconcile(context.Background(), input)
	s.Require().NoError(err)
	s.Require().Equal("discrepancies_found", output.Status)
	s.Require().Len(output.Discrepancies, 1)
	s.Require().Equal(DiscrepancyTypeMissingPayout, output.Discrepancies[0].Type)
}

// TestReconcileOverpayment tests detection of overpayment
func (s *ReconciliationTestSuite) TestReconcileOverpayment() {
	now := time.Now()

	input := ReconciliationInput{
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		Invoices: []billing.InvoiceLedgerRecord{
			{
				InvoiceID: "inv-001",
				Provider:  "provider1",
				Customer:  "customer1",
				Status:    billing.InvoiceStatusPaid,
				Total:     sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
			},
		},
		Payouts: []billing.PayoutRecord{
			{
				PayoutID:    "payout-001",
				Provider:    "provider1",
				GrossAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 2000)), // More than invoice total
				NetAmount:   sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1840)),
				InvoiceIDs:  []string{"inv-001"},
				Status:      billing.PayoutStatusCompleted,
			},
		},
	}

	output, err := s.svc.Reconcile(context.Background(), input)
	s.Require().NoError(err)
	s.Require().Equal("discrepancies_found", output.Status)

	var overpaymentFound bool
	for _, d := range output.Discrepancies {
		if d.Type == DiscrepancyTypeOverpayment {
			overpaymentFound = true
			s.Require().Equal(DiscrepancySeverityHigh, d.Severity)
		}
	}
	s.Require().True(overpaymentFound, "expected overpayment discrepancy")
}

// TestReconcileProviderMismatch tests detection of provider mismatch between usage and invoice
func (s *ReconciliationTestSuite) TestReconcileProviderMismatch() {
	now := time.Now()

	input := ReconciliationInput{
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		UsageRecords: []billing.UsageRecord{
			{
				RecordID:    "usage-001",
				LeaseID:     "lease-001",
				Provider:    "provider-A",
				Customer:    "customer1",
				Status:      billing.UsageRecordStatusInvoiced,
				InvoiceID:   "inv-001",
				TotalAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
			},
		},
		Invoices: []billing.InvoiceLedgerRecord{
			{
				InvoiceID: "inv-001",
				Provider:  "provider-B", // Different provider
				Customer:  "customer1",
				Status:    billing.InvoiceStatusPaid,
				Total:     sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
			},
		},
	}

	output, err := s.svc.Reconcile(context.Background(), input)
	s.Require().NoError(err)

	var mismatchFound bool
	for _, d := range output.Discrepancies {
		if d.Type == DiscrepancyTypeUsageInvoiceMismatch {
			mismatchFound = true
			s.Require().Equal(DiscrepancySeverityCritical, d.Severity)
		}
	}
	s.Require().True(mismatchFound, "expected provider mismatch discrepancy")
}

// TestReconcilePendingUsageSkipped tests that pending usage records are skipped
func (s *ReconciliationTestSuite) TestReconcilePendingUsageSkipped() {
	now := time.Now()

	input := ReconciliationInput{
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		UsageRecords: []billing.UsageRecord{
			{
				RecordID:    "usage-001",
				LeaseID:     "lease-001",
				Provider:    "provider1",
				Customer:    "customer1",
				Status:      billing.UsageRecordStatusPending,
				TotalAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
			},
		},
	}

	output, err := s.svc.Reconcile(context.Background(), input)
	s.Require().NoError(err)
	s.Require().Equal("reconciled", output.Status)
	s.Require().Empty(output.Discrepancies)
	s.Require().Equal(1, output.Summary.PendingUsageCount)
}

// TestReconcilePendingInvoiceSkipped tests that pending invoices are skipped for payout check
func (s *ReconciliationTestSuite) TestReconcilePendingInvoiceSkipped() {
	now := time.Now()

	input := ReconciliationInput{
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		Invoices: []billing.InvoiceLedgerRecord{
			{
				InvoiceID: "inv-001",
				Provider:  "provider1",
				Customer:  "customer1",
				Status:    billing.InvoiceStatusPending,
				Total:     sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
			},
		},
	}

	output, err := s.svc.Reconcile(context.Background(), input)
	s.Require().NoError(err)
	s.Require().Equal("reconciled", output.Status)
	s.Require().Empty(output.Discrepancies)
	s.Require().Equal(1, output.Summary.PendingInvoiceCount)
}

// TestReconcileAmountMismatch tests detection of amount variance between usage and invoice
func (s *ReconciliationTestSuite) TestReconcileAmountMismatch() {
	now := time.Now()

	input := ReconciliationInput{
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		UsageRecords: []billing.UsageRecord{
			{
				RecordID:    "usage-001",
				LeaseID:     "lease-001",
				Provider:    "provider1",
				Customer:    "customer1",
				Status:      billing.UsageRecordStatusInvoiced,
				InvoiceID:   "inv-001",
				TotalAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
			},
		},
		Invoices: []billing.InvoiceLedgerRecord{
			{
				InvoiceID: "inv-001",
				Provider:  "provider1",
				Customer:  "customer1",
				Status:    billing.InvoiceStatusPaid,
				Subtotal:  sdk.NewCoins(sdk.NewInt64Coin("uvirt", 500)), // Only 50% of usage amount
				Total:     sdk.NewCoins(sdk.NewInt64Coin("uvirt", 500)),
			},
		},
	}

	output, err := s.svc.Reconcile(context.Background(), input)
	s.Require().NoError(err)

	// Should find the amount mismatch (50% variance > 1% threshold)
	var amountMismatchFound bool
	for _, d := range output.Discrepancies {
		if d.Type == DiscrepancyTypeUsageInvoiceMismatch && d.InvoiceID == "inv-001" {
			amountMismatchFound = true
		}
	}
	s.Require().True(amountMismatchFound, "expected amount mismatch discrepancy")
}

// TestReconcileAmountWithinThreshold tests that small variances are not flagged
func (s *ReconciliationTestSuite) TestReconcileAmountWithinThreshold() {
	now := time.Now()

	input := ReconciliationInput{
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		UsageRecords: []billing.UsageRecord{
			{
				RecordID:    "usage-001",
				LeaseID:     "lease-001",
				Provider:    "provider1",
				Customer:    "customer1",
				Status:      billing.UsageRecordStatusInvoiced,
				InvoiceID:   "inv-001",
				TotalAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 10000)),
			},
		},
		Invoices: []billing.InvoiceLedgerRecord{
			{
				InvoiceID: "inv-001",
				Provider:  "provider1",
				Customer:  "customer1",
				Status:    billing.InvoiceStatusPaid,
				Subtotal:  sdk.NewCoins(sdk.NewInt64Coin("uvirt", 10000)), // Exact match
				Total:     sdk.NewCoins(sdk.NewInt64Coin("uvirt", 10000)),
			},
		},
		Payouts: []billing.PayoutRecord{
			{
				PayoutID:    "payout-001",
				GrossAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 10000)),
				NetAmount:   sdk.NewCoins(sdk.NewInt64Coin("uvirt", 9200)),
				InvoiceIDs:  []string{"inv-001"},
				Status:      billing.PayoutStatusCompleted,
			},
		},
	}

	output, err := s.svc.Reconcile(context.Background(), input)
	s.Require().NoError(err)
	s.Require().Equal("reconciled", output.Status)
	s.Require().Empty(output.Discrepancies)
}

// TestReconcileSummaryTotals tests that summary totals are calculated correctly
func (s *ReconciliationTestSuite) TestReconcileSummaryTotals() {
	now := time.Now()

	input := ReconciliationInput{
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		UsageRecords: []billing.UsageRecord{
			{
				RecordID:    "usage-001",
				Status:      billing.UsageRecordStatusPending,
				TotalAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 500)),
			},
			{
				RecordID:    "usage-002",
				Status:      billing.UsageRecordStatusInvoiced,
				InvoiceID:   "inv-001",
				TotalAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
			},
			{
				RecordID:    "usage-003",
				Status:      billing.UsageRecordStatusSettled,
				InvoiceID:   "inv-002",
				TotalAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 2000)),
			},
		},
		Invoices: []billing.InvoiceLedgerRecord{
			{
				InvoiceID: "inv-001",
				Status:    billing.InvoiceStatusPending,
				Subtotal:  sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				Total:     sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
			},
			{
				InvoiceID: "inv-002",
				Status:    billing.InvoiceStatusPaid,
				Subtotal:  sdk.NewCoins(sdk.NewInt64Coin("uvirt", 2000)),
				Total:     sdk.NewCoins(sdk.NewInt64Coin("uvirt", 2000)),
			},
		},
		Payouts: []billing.PayoutRecord{
			{
				PayoutID:    "payout-001",
				GrossAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 2000)),
				NetAmount:   sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1840)),
				InvoiceIDs:  []string{"inv-002"},
			},
		},
	}

	output, err := s.svc.Reconcile(context.Background(), input)
	s.Require().NoError(err)

	// Total usage: 500 + 1000 + 2000 = 3500
	s.Require().True(output.Summary.TotalUsageAmount.AmountOf("uvirt").Equal(sdkmath.NewInt(3500)))

	// Total invoices: 1000 + 2000 = 3000
	s.Require().True(output.Summary.TotalInvoiceAmount.AmountOf("uvirt").Equal(sdkmath.NewInt(3000)))

	// Total payouts: 1840
	s.Require().True(output.Summary.TotalPayoutAmount.AmountOf("uvirt").Equal(sdkmath.NewInt(1840)))

	// Counts
	s.Require().Equal(3, output.Summary.UsageRecordCount)
	s.Require().Equal(2, output.Summary.InvoiceCount)
	s.Require().Equal(1, output.Summary.PayoutCount)
	s.Require().Equal(1, output.Summary.PendingUsageCount)
	s.Require().Equal(1, output.Summary.InvoicedUsageCount)
	s.Require().Equal(1, output.Summary.SettledUsageCount)
	s.Require().Equal(1, output.Summary.PendingInvoiceCount)
	s.Require().Equal(1, output.Summary.PaidInvoiceCount)
}

// TestReconcileTooManyDiscrepancies tests the failure threshold
func (s *ReconciliationTestSuite) TestReconcileTooManyDiscrepancies() {
	config := DefaultReconciliationServiceConfig()
	config.MaxDiscrepanciesBeforeFail = 2
	svc := NewReconciliationService(config)

	now := time.Now()

	// Create many invoiced usage records without matching invoices
	var usageRecords []billing.UsageRecord
	for i := 0; i < 5; i++ {
		usageRecords = append(usageRecords, billing.UsageRecord{
			RecordID:    fmt.Sprintf("usage-%03d", i),
			Status:      billing.UsageRecordStatusInvoiced,
			InvoiceID:   fmt.Sprintf("inv-%03d", i),
			TotalAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 100)),
		})
	}

	input := ReconciliationInput{
		PeriodStart:  now.Add(-24 * time.Hour),
		PeriodEnd:    now,
		UsageRecords: usageRecords,
		// No invoices - all usage records will generate discrepancies
	}

	output, err := svc.Reconcile(context.Background(), input)
	s.Require().NoError(err)
	s.Require().Equal("failed", output.Status)
	s.Require().Greater(len(output.Discrepancies), 2)
}

// TestReconcileDisputedInvoices tests that disputed invoices are counted in summary
func (s *ReconciliationTestSuite) TestReconcileDisputedInvoices() {
	now := time.Now()

	input := ReconciliationInput{
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		Invoices: []billing.InvoiceLedgerRecord{
			{
				InvoiceID: "inv-001",
				Status:    billing.InvoiceStatusDisputed,
				Total:     sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
			},
		},
	}

	output, err := s.svc.Reconcile(context.Background(), input)
	s.Require().NoError(err)
	s.Require().Equal(1, output.Summary.DisputedInvoiceCount)
}

// TestDefaultReconciliationServiceConfig tests default config values
func (s *ReconciliationTestSuite) TestDefaultReconciliationServiceConfig() {
	config := DefaultReconciliationServiceConfig()
	s.Require().Equal(1.0, config.VarianceThresholdPercent)
	s.Require().Equal(0.1, config.AutoResolveThresholdPercent)
	s.Require().Equal(100, config.MaxDiscrepanciesBeforeFail)
}

// TestReconcileUsageNoInvoiceID tests usage record with invoiced status but empty invoice ID
func (s *ReconciliationTestSuite) TestReconcileUsageNoInvoiceID() {
	now := time.Now()

	input := ReconciliationInput{
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		UsageRecords: []billing.UsageRecord{
			{
				RecordID:    "usage-001",
				Status:      billing.UsageRecordStatusInvoiced,
				InvoiceID:   "", // No invoice linked
				TotalAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
			},
		},
	}

	output, err := s.svc.Reconcile(context.Background(), input)
	s.Require().NoError(err)
	s.Require().Len(output.Discrepancies, 1)
	s.Require().Equal(DiscrepancyTypeMissingInvoice, output.Discrepancies[0].Type)
}
