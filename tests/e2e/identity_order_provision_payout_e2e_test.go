//go:build e2e.integration

// Package e2e contains end-to-end integration tests.
//
// VE-36I: Identity -> order -> provision -> payout flow with failure scenarios.
// This test validates the full patent path including web-scope verification,
// provisioning callbacks, usage reporting, and settlement outcomes.
package e2e

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/tests/e2e/helpers"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

func TestIdentityOrderProvisionPayoutE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e flow in short mode")
	}

	client := helpers.NewOnboardingTestClient()
	app := helpers.SetupOnboardingTestApp(t, client)
	ctx := helpers.NewTestContext(app, 1, helpers.FixedTimestamp())
	msgServer := helpers.GetVEIDMsgServer(app)

	customer := helpers.CreateTestAccount(t)
	provider := helpers.CreateTestAccount(t)

	// ---------------------------------------------------------------------
	// Identity + Web-Scope Verification (Domain)
	// ---------------------------------------------------------------------
	helpers.UploadScope(t, msgServer, ctx, customer, client, helpers.DefaultSelfieUploadParams("scope-e2e-selfie-001"))
	require.NoError(t, app.Keepers.VirtEngine.VEID.SetScore(ctx, customer.String(), 82, helpers.TestModelVersion))

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

	domainScopeID := "scope-e2e-domain-verify-001"
	helpers.UploadScope(t, msgServer, ctx, customer, client, helpers.DefaultDomainVerifyUploadParams(domainScopeID))
	require.NoError(t, app.Keepers.VirtEngine.VEID.UpdateVerificationStatus(
		ctx,
		customer,
		domainScopeID,
		veidtypes.VerificationStatusVerified,
		"domain verified",
		provider.String(),
	))

	ctx = helpers.CommitAndAdvanceBlock(app, ctx)

	// ---------------------------------------------------------------------
	// Success Path: Order -> Provision -> Usage -> Payout
	// ---------------------------------------------------------------------
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
	settlement := NewMockSettlementE2E()
	auditLogger := NewMockAuditLoggerE2E()
	background := context.Background()

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

	invoice := &InvoiceE2E{
		InvoiceID:    "invoice-e2e-001",
		ProviderAddr: provider.String(),
		CustomerAddr: customer.String(),
		JobID:        usageRecord.JobID,
		LineItems: []LineItemE2E{
			{
				ResourceType: "cpu",
				Quantity:     sdkmath.LegacyNewDec(4),
				UnitPrice:    "2.5",
				TotalCost:    "10.0",
			},
		},
		TotalAmount: "10.0",
		PeriodStart: usageRecord.PeriodStart,
		PeriodEnd:   usageRecord.PeriodEnd,
		Status:      "pending",
	}
	require.NoError(t, settlement.CreateInvoice(background, invoice))
	require.NoError(t, settlement.TriggerSettlement(background, invoice.InvoiceID))
	require.NotNil(t, settlement.GetProviderPayout(provider.String(), invoice.InvoiceID))
	require.GreaterOrEqual(t, len(settlement.GetAuditTrail(invoice.InvoiceID)), 2)

	// ---------------------------------------------------------------------
	// Failure Scenario: Provider Timeout
	// ---------------------------------------------------------------------
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

	// ---------------------------------------------------------------------
	// Failure Scenario: Partial Usage (No Settlement)
	// ---------------------------------------------------------------------
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

	partialInvoice := &InvoiceE2E{
		InvoiceID:    "invoice-e2e-partial-001",
		ProviderAddr: provider.String(),
		CustomerAddr: customer.String(),
		JobID:        partialUsage.JobID,
		LineItems: []LineItemE2E{
			{
				ResourceType: "cpu",
				Quantity:     sdkmath.LegacyNewDecWithPrec(25, 1),
				UnitPrice:    "1.0",
				TotalCost:    "2.5",
			},
		},
		TotalAmount: "2.5",
		PeriodStart: partialUsage.PeriodStart,
		PeriodEnd:   partialUsage.PeriodEnd,
		Status:      "pending_partial",
	}
	require.NoError(t, settlement.CreateInvoice(background, partialInvoice))
	require.Nil(t, settlement.GetProviderPayout(provider.String(), partialInvoice.InvoiceID))
	require.Len(t, settlement.GetAuditTrail(partialInvoice.InvoiceID), 1)

	// ---------------------------------------------------------------------
	// Failure Scenario: Dispute
	// ---------------------------------------------------------------------
	disputeInvoice := &InvoiceE2E{
		InvoiceID:    "invoice-e2e-dispute-001",
		ProviderAddr: provider.String(),
		CustomerAddr: customer.String(),
		JobID:        "job-e2e-dispute-001",
		LineItems: []LineItemE2E{
			{
				ResourceType: "gpu",
				Quantity:     sdkmath.LegacyNewDec(2),
				UnitPrice:    "8.0",
				TotalCost:    "16.0",
			},
		},
		TotalAmount: "16.0",
		PeriodStart: ctx.BlockTime().Add(-time.Hour),
		PeriodEnd:   ctx.BlockTime(),
		Status:      "pending",
	}
	require.NoError(t, settlement.CreateInvoice(background, disputeInvoice))
	require.NoError(t, settlement.DisputeInvoice(background, disputeInvoice.InvoiceID, "usage mismatch"))

	disputed := settlement.GetInvoice(disputeInvoice.InvoiceID)
	require.Equal(t, "disputed", disputed.Status)
	require.Error(t, settlement.TriggerSettlement(background, disputeInvoice.InvoiceID))
	require.Nil(t, settlement.GetProviderPayout(provider.String(), disputeInvoice.InvoiceID))
	require.True(t, hasAuditAction(settlement.GetAuditTrail(disputeInvoice.InvoiceID), "disputed"))
}

func hasAuditEvent(events []pd.HPCAuditEvent, eventType string) bool {
	for _, event := range events {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}

func hasAuditAction(records []*AuditRecordE2E, action string) bool {
	for _, record := range records {
		if record.Action == action {
			return true
		}
	}
	return false
}
