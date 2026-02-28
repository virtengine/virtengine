//go:build e2e.integration

// Package e2e contains end-to-end integration tests.
//
// VE-36I: Identity -> order -> provision -> payout flow with failure scenarios.
// This test validates the full patent path including web-scope verification,
// provisioning callbacks, usage reporting, and settlement outcomes.
package e2e

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/tests/e2e/helpers"
	"github.com/virtengine/virtengine/x/escrow/types/billing"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

func TestIdentityOrderProvisionPayoutE2E(t *testing.T) {
	app := helpers.SetupOnboardingTestApp(t, helpers.NewOnboardingTestClient())
	ctx := helpers.NewTestContext(app, 1, helpers.FixedTimestamp())

	customer := helpers.CreateTestAccount(t)
	provider := helpers.CreateTestAccount(t)

	helpers.CreateIdentityRecordForAccount(t, app, ctx, customer)
	helpers.UpdateAccountScore(t, app, ctx, customer, 82)

	offering := helpers.CreateOfferingWithVEIDRequirement(
		t,
		app,
		ctx,
		provider,
		70,
		string(veidtypes.AccountStatusVerified),
	)
	offering.IdentityRequirement.RequireVerifiedDomain = true
	require.NoError(t, app.Keepers.VirtEngine.Marketplace.UpdateOffering(ctx, &offering))

	helpers.AttemptCreateOrder(t, app, ctx, customer, offering, true)

	helpers.SeedVerifiedScope(t, app, ctx, customer, helpers.DefaultDomainVerifyUploadParams("scope-e2e-domain-verify-001"), provider.String())

	ctx = helpers.CommitAndAdvanceBlock(app, ctx)

	order := helpers.AttemptCreateOrder(t, app, ctx, customer, offering, false)

	bid := marketplace.MarketplaceBid{
		ID: marketplace.BidID{
			OrderID:         order.ID,
			ProviderAddress: provider.String(),
			Sequence:        1,
		},
		OfferingID: offering.ID,
		Price:      4000,
		PublicMetadata: map[string]string{
			"region": "us-east",
		},
		ResourcesOffer: map[string]string{
			"cpu":       "8",
			"memory_gb": "32",
		},
	}
	require.NoError(t, app.Keepers.VirtEngine.Marketplace.CreateBid(ctx, &bid))

	allocation, err := app.Keepers.VirtEngine.Marketplace.AcceptBid(ctx, bid.ID)
	require.NoError(t, err)

	provisionCallback := marketplace.NewWaldurCallbackAt(
		marketplace.ActionTypeProvision,
		"waldur-alloc-1",
		marketplace.SyncTypeAllocation,
		allocation.ID.String(),
		ctx.BlockTime(),
	)
	provisionCallback.Payload["reason"] = "provisioning started"
	provisionCallback.Payload["encrypted_config_ref"] = "config-ref-1"
	provisionCallback.Signature = []byte("e2e-signature")
	provisionCallback.SignerID = "waldur-test"
	require.NoError(t, app.Keepers.VirtEngine.Marketplace.ProcessWaldurCallback(ctx, provisionCallback))

	updatedOrder, found := app.Keepers.VirtEngine.Marketplace.GetOrder(ctx, order.ID)
	require.True(t, found)
	require.Equal(t, marketplace.OrderStateProvisioning, updatedOrder.State)

	updatedAllocation, found := app.Keepers.VirtEngine.Marketplace.GetAllocation(ctx, allocation.ID)
	require.True(t, found)
	require.Equal(t, marketplace.AllocationStateProvisioning, updatedAllocation.State)

	ctx = helpers.CommitAndAdvanceBlock(app, ctx)

	statusCallback := marketplace.NewWaldurCallbackAt(
		marketplace.ActionTypeStatusUpdate,
		"waldur-order-1",
		marketplace.SyncTypeOrder,
		order.ID.String(),
		ctx.BlockTime(),
	)
	statusCallback.Payload["state"] = "active"
	statusCallback.Signature = []byte("e2e-signature")
	statusCallback.SignerID = "waldur-test"
	require.NoError(t, app.Keepers.VirtEngine.Marketplace.ProcessWaldurCallback(ctx, statusCallback))

	updatedOrder, found = app.Keepers.VirtEngine.Marketplace.GetOrder(ctx, order.ID)
	require.True(t, found)
	require.Equal(t, marketplace.OrderStateActive, updatedOrder.State)

	updatedAllocation, found = app.Keepers.VirtEngine.Marketplace.GetAllocation(ctx, allocation.ID)
	require.True(t, found)
	require.NoError(t, updatedAllocation.SetStateAt(marketplace.AllocationStateActive, "provisioned", ctx.BlockTime()))
	require.NoError(t, app.Keepers.VirtEngine.Marketplace.UpdateAllocation(ctx, updatedAllocation))

	usageReporter := NewMockUsageReporterE2E()
	auditLogger := NewMockAuditLoggerE2E()

	fundE2EAccount(t, app, ctx, customer, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(200000))))
	settlementEscrowID := createActiveSettlementEscrowE2E(
		t,
		app,
		ctx,
		order.ID.String(),
		allocation.ID.String(),
		customer,
		provider,
		sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100000))),
		72*time.Hour,
	)

	usageRecord := &UsageRecordE2E{
		RecordID:        "usage-e2e-001",
		JobID:           "job-e2e-success-001",
		ClusterID:       "e2e-cluster-01",
		ProviderAddress: provider.String(),
		CustomerAddress: customer.String(),
		PeriodStart:     ctx.BlockTime().Add(-time.Hour),
		PeriodEnd:       ctx.BlockTime(),
		Metrics: &pd.HPCSchedulerMetrics{
			WallClockSeconds: 3600,
			CPUCoreSeconds:   14400,
			MemoryGBSeconds:  28800,
			NodesUsed:        1,
			NodeHours:        1.0,
		},
		IsFinal:  true,
		JobState: pd.HPCJobStateCompleted,
	}
	require.NoError(t, usageReporter.RecordUsage(usageRecord))
	auditLogger.LogUsageReport(pd.HPCAuditEvent{
		Timestamp: ctx.BlockTime(),
		EventType: "usage_reported",
		JobID:     usageRecord.JobID,
		ClusterID: usageRecord.ClusterID,
		Success:   true,
	})

	settlementUsage := recordSettlementUsageE2E(
		t,
		app,
		ctx,
		order.ID.String(),
		allocation.ID.String(),
		provider,
		customer,
		"cpu",
		4,
		sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10000))),
		usageRecord.PeriodStart,
		usageRecord.PeriodEnd,
	)

	settlementRecord, err := app.Keepers.VirtEngine.Settlement.SettleOrder(ctx, order.ID.String(), []string{settlementUsage.UsageID}, false)
	require.NoError(t, err)
	requireSettlementSplit(t, settlementRecord)

	payout, found := app.Keepers.VirtEngine.Settlement.GetPayoutBySettlement(ctx, settlementRecord.SettlementID)
	require.True(t, found)
	require.Equal(t, settlementtypes.PayoutStateCompleted, payout.State)

	invoiceKeeper := requireE2EInvoiceKeeper(t, app)
	invoiceRecord, err := invoiceKeeper.GetInvoice(ctx, payout.InvoiceID)
	require.NoError(t, err)
	require.Equal(t, billing.InvoiceStatusPaid, invoiceRecord.Status)

	chain, err := invoiceKeeper.GetInvoiceLedgerChain(ctx, payout.InvoiceID)
	require.NoError(t, err)
	require.NoError(t, chain.Validate())

	ctx = helpers.CommitAndAdvanceBlock(app, ctx)
	orderTimeout := helpers.AttemptCreateOrder(t, app, ctx, customer, offering, false)

	timeoutBid := marketplace.MarketplaceBid{
		ID: marketplace.BidID{
			OrderID:         orderTimeout.ID,
			ProviderAddress: provider.String(),
			Sequence:        1,
		},
		OfferingID: offering.ID,
		Price:      3500,
	}
	require.NoError(t, app.Keepers.VirtEngine.Marketplace.CreateBid(ctx, &timeoutBid))
	_, err = app.Keepers.VirtEngine.Marketplace.AcceptBid(ctx, timeoutBid.ID)
	require.NoError(t, err)

	timeoutCallback := marketplace.NewWaldurCallbackAt(
		marketplace.ActionTypeStatusUpdate,
		"waldur-order-timeout",
		marketplace.SyncTypeOrder,
		orderTimeout.ID.String(),
		ctx.BlockTime(),
	)
	timeoutCallback.Payload["state"] = "failed"
	timeoutCallback.Payload["reason"] = "provider timeout"
	timeoutCallback.Signature = []byte("e2e-signature")
	timeoutCallback.SignerID = "waldur-test"
	require.NoError(t, app.Keepers.VirtEngine.Marketplace.ProcessWaldurCallback(ctx, timeoutCallback))

	failedOrder, found := app.Keepers.VirtEngine.Marketplace.GetOrder(ctx, orderTimeout.ID)
	require.True(t, found)
	require.Equal(t, marketplace.OrderStateFailed, failedOrder.State)

	auditLogger.LogJobEvent(pd.HPCAuditEvent{
		Timestamp: ctx.BlockTime(),
		EventType: "provision_timeout",
		JobID:     "job-timeout-001",
		ClusterID: "e2e-cluster-01",
		Success:   false,
		ErrorMsg:  "provider timeout",
	})
	require.True(t, hasAuditEvent(auditLogger.GetEvents(), "provision_timeout"))

	partialUsage := &UsageRecordE2E{
		RecordID:        "usage-e2e-partial-001",
		JobID:           "job-e2e-partial-001",
		ClusterID:       "e2e-cluster-01",
		ProviderAddress: provider.String(),
		CustomerAddress: customer.String(),
		PeriodStart:     ctx.BlockTime().Add(-30 * time.Minute),
		PeriodEnd:       ctx.BlockTime(),
		Metrics: &pd.HPCSchedulerMetrics{
			WallClockSeconds: 900,
			CPUCoreSeconds:   1800,
			MemoryGBSeconds:  3600,
			NodesUsed:        1,
			NodeHours:        0.25,
		},
		IsFinal:  false,
		JobState: pd.HPCJobStateRunning,
	}
	require.NoError(t, usageReporter.RecordUsage(partialUsage))
	auditLogger.LogUsageReport(pd.HPCAuditEvent{
		Timestamp: ctx.BlockTime(),
		EventType: "usage_partial",
		JobID:     partialUsage.JobID,
		ClusterID: partialUsage.ClusterID,
		Success:   true,
	})

	partialOrderID := "identity-partial-order"
	partialLeaseID := "identity-partial-lease"
	createActiveSettlementEscrowE2E(
		t,
		app,
		ctx,
		partialOrderID,
		partialLeaseID,
		customer,
		provider,
		sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(5000))),
		24*time.Hour,
	)

	partialSettlementUsage := recordSettlementUsageE2E(
		t,
		app,
		ctx,
		partialOrderID,
		partialLeaseID,
		provider,
		customer,
		"cpu",
		1,
		sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(2500))),
		partialUsage.PeriodStart,
		partialUsage.PeriodEnd,
	)

	storedPartialUsage, found := app.Keepers.VirtEngine.Settlement.GetUsageRecord(ctx, partialSettlementUsage.UsageID)
	require.True(t, found)
	require.False(t, storedPartialUsage.Settled)
	require.Empty(t, app.Keepers.VirtEngine.Settlement.GetSettlementsByOrder(ctx, partialOrderID))

	disputeOrderID := "identity-dispute-order"
	disputeLeaseID := "identity-dispute-lease"
	disputeEscrowID := createActiveSettlementEscrowE2E(
		t,
		app,
		ctx,
		disputeOrderID,
		disputeLeaseID,
		customer,
		provider,
		sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(16000))),
		24*time.Hour,
	)

	disputeSettlementUsage := recordSettlementUsageE2E(
		t,
		app,
		ctx,
		disputeOrderID,
		disputeLeaseID,
		provider,
		customer,
		"gpu",
		2,
		sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(16000))),
		ctx.BlockTime().Add(-time.Hour),
		ctx.BlockTime(),
	)

	require.NoError(t, app.Keepers.VirtEngine.Settlement.DisputeEscrow(ctx, disputeEscrowID, "usage mismatch"))

	disputedEscrow, found := app.Keepers.VirtEngine.Settlement.GetEscrow(ctx, disputeEscrowID)
	require.True(t, found)
	require.Equal(t, settlementtypes.EscrowStateDisputed, disputedEscrow.State)

	_, err = app.Keepers.VirtEngine.Settlement.SettleOrder(ctx, disputeOrderID, []string{disputeSettlementUsage.UsageID}, false)
	require.Error(t, err)
	require.Empty(t, app.Keepers.VirtEngine.Settlement.GetSettlementsByOrder(ctx, disputeOrderID))

	require.NoError(t, app.Keepers.VirtEngine.Settlement.RefundEscrow(ctx, disputeEscrowID, "customer dispute upheld"))
	disputedEscrow, found = app.Keepers.VirtEngine.Settlement.GetEscrow(ctx, disputeEscrowID)
	require.True(t, found)
	require.Equal(t, settlementtypes.EscrowStateRefunded, disputedEscrow.State)

	activeEscrow, found := app.Keepers.VirtEngine.Settlement.GetEscrow(ctx, settlementEscrowID)
	require.True(t, found)
	require.Equal(t, settlementtypes.EscrowStateActive, activeEscrow.State)
}

func hasAuditEvent(events []pd.HPCAuditEvent, eventType string) bool {
	for _, event := range events {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}
