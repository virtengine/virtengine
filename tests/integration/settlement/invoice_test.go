// Package settlement provides integration tests for the billing settlement system.
// These tests verify the end-to-end usage→invoice→settlement→payment pipeline.
//
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

// InvoiceIntegrationTestSuite provides integration tests for invoice lifecycle and ledger
type InvoiceIntegrationTestSuite struct {
	suite.Suite
	ctx           sdk.Context
	escrowKeeper  escrowkeeper.Keeper
	invoiceKeeper escrowkeeper.InvoiceKeeper
	provider      sdk.AccAddress
	customer      sdk.AccAddress
	currency      string
}

func TestInvoiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(InvoiceIntegrationTestSuite))
}

func (suite *InvoiceIntegrationTestSuite) SetupTest() {
	// Setup test environment using testutil helpers
	suite.ctx, suite.escrowKeeper = testutil.SetupEscrowKeeper(suite.T())
	suite.invoiceKeeper = suite.escrowKeeper.NewInvoiceKeeper()

	suite.provider = testutil.AccAddress(suite.T())
	suite.customer = testutil.AccAddress(suite.T())
	suite.currency = "uakt"
}

// TestInvoiceCreation verifies invoice creation with ledger record
func (suite *InvoiceIntegrationTestSuite) TestInvoiceCreation() {
	t := suite.T()
	ctx := suite.ctx

	now := ctx.BlockTime()
	periodStart := now.Add(-24 * time.Hour)
	periodEnd := now

	// Create invoice
	invoice := &billing.Invoice{
		InvoiceID:     fmt.Sprintf("inv-%d", ctx.BlockHeight()),
		InvoiceNumber: "VE-TEST-001",
		EscrowID:      "escrow-test-1",
		OrderID:       "order-test-1",
		LeaseID:       "lease-test-1",
		Provider:      suite.provider.String(),
		Customer:      suite.customer.String(),
		Status:        billing.InvoiceStatusPending,
		LineItems: []billing.LineItem{
			{
				Description: "CPU usage",
				Quantity:    sdkmath.LegacyNewDec(100),
				UnitPrice:   sdk.NewDecCoin(suite.currency, sdkmath.NewInt(10)),
				Amount:      sdk.NewCoins(sdk.NewCoin(suite.currency, sdkmath.NewInt(1000))),
			},
		},
		Total:      sdk.NewCoins(sdk.NewCoin(suite.currency, sdkmath.NewInt(1000))),
		AmountDue:  sdk.NewCoins(sdk.NewCoin(suite.currency, sdkmath.NewInt(1000))),
		AmountPaid: sdk.NewCoins(),
		Currency:   suite.currency,
		BillingPeriod: billing.BillingPeriod{
			StartTime:       periodStart,
			EndTime:         periodEnd,
			DurationSeconds: 86400,
			PeriodType:      billing.BillingPeriodTypeMonthly,
		},
		IssuedAt:  &now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	artifactCID := fmt.Sprintf("ipfs-%s", invoice.InvoiceID)

	// Create invoice and ledger record
	record, err := suite.invoiceKeeper.CreateInvoice(ctx, invoice, artifactCID)
	require.NoError(t, err)
	require.NotNil(t, record)

	// Verify ledger record fields
	require.Equal(t, invoice.InvoiceID, record.InvoiceID)
	require.Equal(t, invoice.Provider, record.Provider)
	require.Equal(t, invoice.Customer, record.Customer)
	require.Equal(t, billing.InvoiceStatusPending, record.Status)
	require.True(t, record.Total.IsEqual(invoice.Total))
	require.NotEmpty(t, record.ContentHash)
	require.Equal(t, artifactCID, record.ArtifactCID)

	// Verify retrieval
	retrieved, err := suite.invoiceKeeper.GetInvoice(ctx, invoice.InvoiceID)
	require.NoError(t, err)
	require.Equal(t, record.InvoiceID, retrieved.InvoiceID)
	require.Equal(t, record.ContentHash, retrieved.ContentHash)
}

// TestInvoiceStatusTransitions tests the invoice state machine
func (suite *InvoiceIntegrationTestSuite) TestInvoiceStatusTransitions() {
	t := suite.T()
	ctx := suite.ctx

	// Create invoice
	invoice := suite.createTestInvoice("transition-test", 1000)
	record, err := suite.invoiceKeeper.CreateInvoice(ctx, invoice, "artifact-1")
	require.NoError(t, err)
	require.Equal(t, billing.InvoiceStatusPending, record.Status)

	// Test valid transition: Pending → Paid
	entry, err := suite.invoiceKeeper.UpdateInvoiceStatus(
		ctx,
		invoice.InvoiceID,
		billing.InvoiceStatusPaid,
		"test-initiator",
	)
	require.NoError(t, err)
	require.NotNil(t, entry)
	require.Equal(t, billing.InvoiceStatusPaid, entry.NewStatus)
	require.Equal(t, billing.InvoiceStatusPending, entry.OldStatus)

	// Verify record updated
	updated, err := suite.invoiceKeeper.GetInvoice(ctx, invoice.InvoiceID)
	require.NoError(t, err)
	require.Equal(t, billing.InvoiceStatusPaid, updated.Status)
	require.NotNil(t, updated.PaidAt)

	// Test invalid transition: Paid → Pending (should fail)
	_, err = suite.invoiceKeeper.UpdateInvoiceStatus(
		ctx,
		invoice.InvoiceID,
		billing.InvoiceStatusPending,
		"test-initiator",
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid transition")
}

// TestInvoicePaymentRecording tests payment recording and partial payments
func (suite *InvoiceIntegrationTestSuite) TestInvoicePaymentRecording() {
	t := suite.T()
	ctx := suite.ctx

	totalAmount := sdkmath.NewInt(1000)
	invoice := suite.createTestInvoice("payment-test", totalAmount.Int64())
	_, err := suite.invoiceKeeper.CreateInvoice(ctx, invoice, "artifact-pay")
	require.NoError(t, err)

	// Record partial payment (50%)
	partialAmount := sdk.NewCoins(sdk.NewCoin(suite.currency, totalAmount.QuoRaw(2)))
	entry, err := suite.invoiceKeeper.RecordPayment(
		ctx,
		invoice.InvoiceID,
		partialAmount,
		"payment-processor",
	)
	require.NoError(t, err)
	require.NotNil(t, entry)
	require.Equal(t, billing.LedgerEntryTypePayment, entry.EntryType)
	require.True(t, entry.Amount.IsEqual(partialAmount))

	// Verify partial payment status
	updated, err := suite.invoiceKeeper.GetInvoice(ctx, invoice.InvoiceID)
	require.NoError(t, err)
	require.Equal(t, billing.InvoiceStatusPartiallyPaid, updated.Status)
	require.True(t, updated.AmountPaid.IsEqual(partialAmount))
	require.False(t, updated.AmountDue.IsZero())

	// Record remaining payment (50%)
	remainingAmount := sdk.NewCoins(sdk.NewCoin(suite.currency, totalAmount.QuoRaw(2)))
	entry2, err := suite.invoiceKeeper.RecordPayment(
		ctx,
		invoice.InvoiceID,
		remainingAmount,
		"payment-processor",
	)
	require.NoError(t, err)
	require.NotNil(t, entry2)

	// Verify fully paid status
	final, err := suite.invoiceKeeper.GetInvoice(ctx, invoice.InvoiceID)
	require.NoError(t, err)
	require.Equal(t, billing.InvoiceStatusPaid, final.Status)
	require.True(t, final.AmountPaid.IsAllGTE(invoice.Total))
	require.NotNil(t, final.PaidAt)
}

// TestLedgerChainIntegrity tests the hash-chained ledger entries
func (suite *InvoiceIntegrationTestSuite) TestLedgerChainIntegrity() {
	t := suite.T()
	ctx := suite.ctx

	invoice := suite.createTestInvoice("ledger-chain-test", 1000)
	_, err := suite.invoiceKeeper.CreateInvoice(ctx, invoice, "artifact-ledger")
	require.NoError(t, err)

	// Perform multiple state transitions
	_, err = suite.invoiceKeeper.UpdateInvoiceStatus(ctx, invoice.InvoiceID, billing.InvoiceStatusPaid, "sys")
	require.NoError(t, err)

	payment := sdk.NewCoins(sdk.NewCoin(suite.currency, sdkmath.NewInt(500)))
	_, err = suite.invoiceKeeper.RecordPayment(ctx, invoice.InvoiceID, payment, "sys")
	require.NoError(t, err)

	// Get and verify ledger chain
	chain, err := suite.invoiceKeeper.GetInvoiceLedgerChain(ctx, invoice.InvoiceID)
	require.NoError(t, err)
	require.NotNil(t, chain)
	require.True(t, len(chain.Entries) >= 3) // Genesis + status change + payment

	// Verify hash chain integrity
	err = chain.Validate()
	require.NoError(t, err)

	// Verify using keeper method
	err = suite.invoiceKeeper.VerifyLedgerChain(ctx, invoice.InvoiceID)
	require.NoError(t, err)
}

// Helper methods

func (suite *InvoiceIntegrationTestSuite) createTestInvoice(idSuffix string, totalAmount int64) *billing.Invoice {
	now := suite.ctx.BlockTime()
	periodStart := now.Add(-24 * time.Hour)
	periodEnd := now

	invoiceID := fmt.Sprintf("inv-%s-%d", idSuffix, suite.ctx.BlockHeight())

	return &billing.Invoice{
		InvoiceID:     invoiceID,
		InvoiceNumber: fmt.Sprintf("VE-%s", idSuffix),
		EscrowID:      fmt.Sprintf("escrow-%s", idSuffix),
		OrderID:       fmt.Sprintf("order-%s", idSuffix),
		LeaseID:       fmt.Sprintf("lease-%s", idSuffix),
		Provider:      suite.provider.String(),
		Customer:      suite.customer.String(),
		Status:        billing.InvoiceStatusPending,
		LineItems: []billing.LineItem{
			{
				Description: "Test usage",
				Quantity:    sdkmath.LegacyNewDec(100),
				UnitPrice:   sdk.NewDecCoin(suite.currency, sdkmath.NewInt(totalAmount/100)),
				Amount:      sdk.NewCoins(sdk.NewCoin(suite.currency, sdkmath.NewInt(totalAmount))),
			},
		},
		Total:      sdk.NewCoins(sdk.NewCoin(suite.currency, sdkmath.NewInt(totalAmount))),
		AmountDue:  sdk.NewCoins(sdk.NewCoin(suite.currency, sdkmath.NewInt(totalAmount))),
		AmountPaid: sdk.NewCoins(),
		Currency:   suite.currency,
		BillingPeriod: billing.BillingPeriod{
			StartTime:       periodStart,
			EndTime:         periodEnd,
			DurationSeconds: 86400,
			PeriodType:      billing.BillingPeriodTypeMonthly,
		},
		IssuedAt:  &now,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
