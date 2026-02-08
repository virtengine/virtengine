//go:build e2e.integration

package settlement_test

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/testutil"
	escrowkeeper "github.com/virtengine/virtengine/x/escrow/keeper"
	"github.com/virtengine/virtengine/x/escrow/types/billing"
)

// UsagePipelineTestSuite tests the usage→invoice→settlement pipeline
type UsagePipelineTestSuite struct {
	suite.Suite
	ctx           sdk.Context
	escrowKeeper  escrowkeeper.Keeper
	usagePipeline escrowkeeper.UsagePipelineKeeper
	provider      sdk.AccAddress
	customer      sdk.AccAddress
	currency      string
}

func TestUsagePipelineTestSuite(t *testing.T) {
	suite.Run(t, new(UsagePipelineTestSuite))
}

func (suite *UsagePipelineTestSuite) SetupTest() {
	suite.ctx, suite.escrowKeeper = testutil.SetupEscrowKeeper(suite.T())
	suite.usagePipeline = suite.escrowKeeper.NewUsagePipelineKeeper()

	suite.provider = testutil.AccAddress(suite.T())
	suite.customer = testutil.AccAddress(suite.T())
	suite.currency = "uakt"
}

// TestSubmitUsageReport tests usage report submission
func (suite *UsagePipelineTestSuite) TestSubmitUsageReport() {
	t := suite.T()
	ctx := suite.ctx

	now := ctx.BlockTime()
	periodStart := now.Add(-1 * time.Hour)
	periodEnd := now

	report := &escrowkeeper.UsageReport{
		Provider: suite.provider.String(),
		LeaseID:  "lease-test-1",
		Customer: suite.customer.String(),
		EscrowID: "escrow-test-1",
		Resources: []escrowkeeper.ResourceUsage{
			{
				Type:      billing.UsageTypeCPU,
				Quantity:  sdkmath.LegacyNewDec(100),
				Unit:      "cpu-hours",
				UnitPrice: sdk.NewDecCoin(suite.currency, sdkmath.NewInt(10)),
			},
			{
				Type:      billing.UsageTypeMemory,
				Quantity:  sdkmath.LegacyNewDec(200),
				Unit:      "gb-hours",
				UnitPrice: sdk.NewDecCoin(suite.currency, sdkmath.NewInt(5)),
			},
		},
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}

	record, err := suite.usagePipeline.SubmitUsageReport(ctx, report)
	require.NoError(t, err)
	require.NotNil(t, record)

	// Verify record fields
	require.Equal(t, report.LeaseID, record.LeaseID)
	require.Equal(t, report.Provider, record.Provider)
	require.Equal(t, report.Customer, record.Customer)
	require.Equal(t, billing.UsageRecordStatusPending, record.Status)
	require.False(t, record.TotalAmount.IsZero())
}

// TestGenerateInvoiceFromUsage tests invoice generation from usage
func (suite *UsagePipelineTestSuite) TestGenerateInvoiceFromUsage() {
	t := suite.T()
	ctx := suite.ctx

	leaseID := "lease-invoice-gen"
	now := ctx.BlockTime()

	// Submit multiple usage reports
	for i := 0; i < 3; i++ {
		report := suite.createUsageReport(leaseID, i)
		_, err := suite.usagePipeline.SubmitUsageReport(ctx, report)
		require.NoError(t, err)
	}

	// Generate invoice from accumulated usage
	invoiceRecord, err := suite.usagePipeline.GenerateInvoiceFromUsage(ctx, leaseID, now)
	require.NoError(t, err)
	require.NotNil(t, invoiceRecord)

	// Verify invoice created
	require.Equal(t, leaseID, invoiceRecord.LeaseID)
	require.Equal(t, suite.provider.String(), invoiceRecord.Provider)
	require.Equal(t, suite.customer.String(), invoiceRecord.Customer)
	require.Equal(t, billing.InvoiceStatusPending, invoiceRecord.Status)
	require.False(t, invoiceRecord.Total.IsZero())

	// Verify usage records marked as invoiced
	pending, err := suite.usagePipeline.GetPendingUsageRecords(ctx, leaseID)
	require.NoError(t, err)
	require.Empty(t, pending) // All should be invoiced
}

// TestFullUsageSettlementPipeline tests the complete pipeline
func (suite *UsagePipelineTestSuite) TestFullUsageSettlementPipeline() {
	t := suite.T()
	ctx := suite.ctx

	leaseID := "lease-full-pipeline"

	// Step 1: Submit usage
	report := suite.createUsageReport(leaseID, 0)
	usageRecord, err := suite.usagePipeline.SubmitUsageReport(ctx, report)
	require.NoError(t, err)
	require.Equal(t, billing.UsageRecordStatusPending, usageRecord.Status)

	// Step 2: Process full settlement pipeline
	result, err := suite.usagePipeline.ProcessUsageSettlement(ctx, leaseID)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify result
	require.NotEmpty(t, result.InvoiceID)
	require.Contains(t, result.UsageRecordIDs, usageRecord.RecordID)
	require.NotEmpty(t, result.Status)
	require.False(t, result.TotalAmount.IsZero())
}

// TestApproveInvoice tests invoice approval
func (suite *UsagePipelineTestSuite) TestApproveInvoice() {
	t := suite.T()
	ctx := suite.ctx

	leaseID := "lease-approval-test"

	// Submit usage and generate invoice
	report := suite.createUsageReport(leaseID, 0)
	_, err := suite.usagePipeline.SubmitUsageReport(ctx, report)
	require.NoError(t, err)

	invoiceRecord, err := suite.usagePipeline.GenerateInvoiceFromUsage(ctx, leaseID, ctx.BlockTime())
	require.NoError(t, err)
	require.Equal(t, billing.InvoiceStatusPending, invoiceRecord.Status)

	// Approve invoice
	err = suite.usagePipeline.ApproveInvoice(ctx, invoiceRecord.InvoiceID, suite.customer.String())
	require.NoError(t, err)

	// Verify status changed to paid
	ik := suite.escrowKeeper.NewInvoiceKeeper()
	updated, err := ik.GetInvoice(ctx, invoiceRecord.InvoiceID)
	require.NoError(t, err)
	require.Equal(t, billing.InvoiceStatusPaid, updated.Status)
}

// TestDisputeInvoice tests invoice dispute
func (suite *UsagePipelineTestSuite) TestDisputeInvoice() {
	t := suite.T()
	ctx := suite.ctx

	leaseID := "lease-dispute-test"

	// Submit usage and generate invoice
	report := suite.createUsageReport(leaseID, 0)
	_, err := suite.usagePipeline.SubmitUsageReport(ctx, report)
	require.NoError(t, err)

	invoiceRecord, err := suite.usagePipeline.GenerateInvoiceFromUsage(ctx, leaseID, ctx.BlockTime())
	require.NoError(t, err)

	// Dispute invoice
	err = suite.usagePipeline.DisputeInvoice(ctx, invoiceRecord.InvoiceID, suite.customer.String(), "Incorrect usage calculation")
	require.NoError(t, err)

	// Verify status changed to disputed
	ik := suite.escrowKeeper.NewInvoiceKeeper()
	updated, err := ik.GetInvoice(ctx, invoiceRecord.InvoiceID)
	require.NoError(t, err)
	require.Equal(t, billing.InvoiceStatusDisputed, updated.Status)
}

// TestUsageReportValidation tests usage report validation
func (suite *UsagePipelineTestSuite) TestUsageReportValidation() {
	t := suite.T()
	ctx := suite.ctx

	// Test: Empty provider
	invalidReport := &escrowkeeper.UsageReport{
		Provider: "",
		LeaseID:  "lease-1",
		Customer: suite.customer.String(),
		EscrowID: "escrow-1",
		Resources: []escrowkeeper.ResourceUsage{
			{
				Type:      billing.UsageTypeCPU,
				Quantity:  sdkmath.LegacyNewDec(100),
				Unit:      "hours",
				UnitPrice: sdk.NewDecCoin(suite.currency, sdkmath.NewInt(10)),
			},
		},
		PeriodStart: ctx.BlockTime().Add(-1 * time.Hour),
		PeriodEnd:   ctx.BlockTime(),
	}
	_, err := suite.usagePipeline.SubmitUsageReport(ctx, invalidReport)
	require.Error(t, err)
	require.Contains(t, err.Error(), "provider")

	// Test: Invalid period (end before start)
	invalidPeriod := &escrowkeeper.UsageReport{
		Provider:    suite.provider.String(),
		LeaseID:     "lease-1",
		Customer:    suite.customer.String(),
		EscrowID:    "escrow-1",
		Resources:   []escrowkeeper.ResourceUsage{{Type: billing.UsageTypeCPU, Quantity: sdkmath.LegacyNewDec(1), Unit: "h", UnitPrice: sdk.NewDecCoin(suite.currency, sdkmath.NewInt(1))}},
		PeriodStart: ctx.BlockTime(),
		PeriodEnd:   ctx.BlockTime().Add(-1 * time.Hour),
	}
	_, err = suite.usagePipeline.SubmitUsageReport(ctx, invalidPeriod)
	require.Error(t, err)

	// Test: Negative quantity
	negativeQuantity := &escrowkeeper.UsageReport{
		Provider: suite.provider.String(),
		LeaseID:  "lease-1",
		Customer: suite.customer.String(),
		EscrowID: "escrow-1",
		Resources: []escrowkeeper.ResourceUsage{
			{
				Type:      billing.UsageTypeCPU,
				Quantity:  sdkmath.LegacyNewDec(-100),
				Unit:      "hours",
				UnitPrice: sdk.NewDecCoin(suite.currency, sdkmath.NewInt(10)),
			},
		},
		PeriodStart: ctx.BlockTime().Add(-1 * time.Hour),
		PeriodEnd:   ctx.BlockTime(),
	}
	_, err = suite.usagePipeline.SubmitUsageReport(ctx, negativeQuantity)
	require.Error(t, err)
	require.Contains(t, err.Error(), "negative")
}

// Helper methods

func (suite *UsagePipelineTestSuite) createUsageReport(leaseID string, index int) *escrowkeeper.UsageReport {
	now := suite.ctx.BlockTime()
	periodStart := now.Add(time.Duration(-1*(index+1)) * time.Hour)
	periodEnd := now.Add(time.Duration(-index) * time.Hour)

	return &escrowkeeper.UsageReport{
		Provider: suite.provider.String(),
		LeaseID:  leaseID,
		Customer: suite.customer.String(),
		EscrowID: fmt.Sprintf("escrow-%s", leaseID),
		Resources: []escrowkeeper.ResourceUsage{
			{
				Type:      billing.UsageTypeCPU,
				Quantity:  sdkmath.LegacyNewDec(100 * int64(index+1)),
				Unit:      "cpu-hours",
				UnitPrice: sdk.NewDecCoin(suite.currency, sdkmath.NewInt(10)),
			},
		},
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}
}
