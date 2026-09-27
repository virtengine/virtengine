package keeper_test

import (
	"bytes"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/testutil/state"
	"github.com/virtengine/virtengine/x/escrow/types/billing"
)

// TestDisputeResolutionIsOrderScoped is DONE-WHEN criterion 2: resolving a dispute
// changes only that order/invoice/escrow, and the accounts involved are left untouched.
//
// A dispute about one order is a statement about that order. It must not switch the
// customer, provider or resolver off across every other service they use, so the test
// asserts the invoice moved and the global account state did not.
func TestDisputeResolutionIsOrderScoped(t *testing.T) {
	// Use the suite variant that pins a real block time: dispute workflow audit entries
	// require a non-zero timestamp, and the plain suite leaves BlockTime at the zero value.
	suite := state.SetupTestSuiteWithoutModuleServices(t)
	ctx := suite.Context()

	escrowKeeper := suite.EscrowKeeper()
	invoiceKeeper := escrowKeeper.NewInvoiceKeeper()
	disputeKeeper := escrowKeeper.NewDisputeKeeper()
	rolesKeeper := suite.App().Keepers.VirtEngine.Roles

	customer := sdk.AccAddress(bytes.Repeat([]byte{11}, 20))
	provider := sdk.AccAddress(bytes.Repeat([]byte{12}, 20))
	resolver := sdk.AccAddress(bytes.Repeat([]byte{13}, 20))

	// --- Baseline: none of these accounts has any global state, and all are operational.
	for _, addr := range []sdk.AccAddress{customer, provider, resolver} {
		_, found := rolesKeeper.GetAccountState(ctx, addr)
		require.False(t, found, "fixture should start with no global account state")
		require.True(t, rolesKeeper.IsAccountOperational(ctx, addr))
	}

	// --- Create an invoice for one order and issue it (Draft -> Pending).
	invoice := billing.NewInvoice(
		"inv-scope-1",
		"VE-INV-00000001",
		"escrow-scope-1",
		"order-scope-1",
		"lease-scope-1",
		provider.String(),
		customer.String(),
		"uvirt",
		billing.BillingPeriod{
			StartTime:       ctx.BlockTime().Add(-24 * time.Hour),
			EndTime:         ctx.BlockTime(),
			DurationSeconds: 86400,
			PeriodType:      billing.BillingPeriodTypeDaily,
		},
		ctx.BlockTime().Add(7*24*time.Hour),
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)
	invoice.AddLineItem(billing.LineItem{
		LineItemID:  "li-1",
		Description: "compute usage",
		UsageType:   billing.UsageTypeCPU,
		Quantity:    sdkmath.LegacyNewDec(10),
		Unit:        "core-hour",
		UnitPrice:   sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyNewDec(500000)),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("uvirt", 5000000)),
	})

	_, err := invoiceKeeper.CreateInvoice(ctx, invoice, "")
	require.NoError(t, err)

	_, err = invoiceKeeper.UpdateInvoiceStatus(ctx, invoice.InvoiceID, billing.InvoiceStatusPending, provider.String())
	require.NoError(t, err)

	issued, err := invoiceKeeper.GetInvoice(ctx, invoice.InvoiceID)
	require.NoError(t, err)
	require.Equal(t, billing.InvoiceStatusPending, issued.Status)

	// --- Open a dispute against that one order's invoice.
	disputedAmount := sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000000))
	workflow, err := disputeKeeper.InitiateDispute(
		ctx,
		invoice.InvoiceID,
		billing.DisputeCategoryUsageMismatch,
		"usage mismatch",
		"billed for more cores than were delivered",
		disputedAmount,
		customer.String(),
	)
	require.NoError(t, err)
	require.NotNil(t, workflow)

	disputedInvoice, err := invoiceKeeper.GetInvoice(ctx, invoice.InvoiceID)
	require.NoError(t, err)
	require.Equal(t, billing.InvoiceStatusDisputed, disputedInvoice.Status,
		"opening the dispute should move the invoice, not the account")

	// The dispute must not have touched global account state either.
	for _, addr := range []sdk.AccAddress{customer, provider, resolver} {
		_, found := rolesKeeper.GetAccountState(ctx, addr)
		require.False(t, found, "opening a dispute must not write global account state")
		require.True(t, rolesKeeper.IsAccountOperational(ctx, addr),
			"opening a dispute must not suspend an account")
	}

	// --- Resolve the dispute in the customer's favour with a refund.
	require.NoError(t, disputeKeeper.ResolveDispute(
		ctx,
		workflow.DisputeID,
		billing.DisputeResolutionCustomerWin,
		"provider conceded the usage mismatch",
		resolver.String(),
		disputedAmount,
	))

	// The order-scoped effect landed: the invoice is refunded for that order.
	resolvedInvoice, err := invoiceKeeper.GetInvoice(ctx, invoice.InvoiceID)
	require.NoError(t, err)
	require.Equal(t, billing.InvoiceStatusRefunded, resolvedInvoice.Status,
		"the resolution must land on the disputed invoice")

	resolvedWorkflow, err := disputeKeeper.GetDisputeWorkflow(ctx, workflow.DisputeID)
	require.NoError(t, err)
	require.Equal(t, billing.DisputeStatusResolved, resolvedWorkflow.Status)

	// And the account-wide effect did NOT: resolving one order's dispute left every
	// involved account exactly as it was.
	for _, addr := range []sdk.AccAddress{customer, provider, resolver} {
		_, found := rolesKeeper.GetAccountState(ctx, addr)
		require.False(t, found,
			"resolving a dispute must not create global account state for the account")
		require.True(t, rolesKeeper.IsAccountOperational(ctx, addr),
			"resolving a dispute must not leave the account non-operational")
	}
}
