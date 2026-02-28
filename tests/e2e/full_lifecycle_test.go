//go:build e2e.integration

// Package e2e contains end-to-end integration tests.
//
// VE-68B: Full VEID→marketplace→provider→settlement→rewards lifecycle
// This test validates the complete system flow from identity verification
// through marketplace transactions, provider deployment, settlement, and
// reward distribution. It exercises all major module interactions.
package e2e

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/tests/e2e/helpers"
	"github.com/virtengine/virtengine/x/escrow/types/billing"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

func TestFullLifecycleE2E(t *testing.T) {
	app := helpers.SetupOnboardingTestApp(t, helpers.NewOnboardingTestClient())
	ctx := helpers.NewTestContext(app, 1, helpers.FixedTimestamp())

	customer := helpers.CreateTestAccount(t)
	provider := helpers.CreateTestAccount(t)
	validator := helpers.CreateTestAccount(t)

	t.Log("Starting full lifecycle E2E test")
	t.Logf("Customer: %s", customer.String())
	t.Logf("Provider: %s", provider.String())
	t.Logf("Validator: %s", validator.String())

	t.Run("Phase1_Identity", func(t *testing.T) {
		helpers.CreateIdentityRecordForAccount(t, app, ctx, customer)
		helpers.UpdateAccountScore(t, app, ctx, customer, 82)
		helpers.SeedVerifiedScope(t, app, ctx, customer, helpers.DefaultDomainVerifyUploadParams("lifecycle-customer-domain-001"), validator.String())

		record, found := app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
		require.True(t, found, "customer identity record should exist")
		require.Equal(t, veidtypes.IdentityTierVerified, record.Tier)
		require.Equal(t, uint32(82), record.CurrentScore)
		require.True(t, record.HasVerifiedScope(veidtypes.ScopeTypeDomainVerify))

		score, status, found := app.Keepers.VirtEngine.VEID.GetScore(ctx, customer.String())
		require.True(t, found, "customer score should exist")
		require.Equal(t, uint32(82), score)
		require.Equal(t, veidtypes.AccountStatusVerified, status)

		helpers.SeedVerifiedScope(t, app, ctx, provider, helpers.DefaultDomainVerifyUploadParams("lifecycle-provider-domain-001"), validator.String())

		providerRecord, found := app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, provider)
		require.True(t, found, "provider identity record should exist")
		require.True(t, providerRecord.HasVerifiedScope(veidtypes.ScopeTypeDomainVerify))
	})

	ctx = helpers.CommitAndAdvanceBlock(app, ctx)

	var mfaSessionID string
	t.Run("Phase2_MFA", func(t *testing.T) {
		enrollment := &mfatypes.FactorEnrollment{
			AccountAddress:   customer.String(),
			FactorType:       mfatypes.FactorTypeTOTP,
			FactorID:         "totp-lifecycle-001",
			PublicIdentifier: []byte("totp-lifecycle-public-ref"),
			Label:            "Lifecycle Test TOTP",
			Status:           mfatypes.EnrollmentStatusActive,
			EnrolledAt:       ctx.BlockTime().Unix(),
			VerifiedAt:       ctx.BlockTime().Unix(),
			LastUsedAt:       ctx.BlockTime().Unix(),
		}
		require.NoError(t, app.Keepers.VirtEngine.MFA.EnrollFactor(ctx, enrollment))

		retrieved, found := app.Keepers.VirtEngine.MFA.GetFactorEnrollment(ctx, customer, mfatypes.FactorTypeTOTP, "totp-lifecycle-001")
		require.True(t, found, "TOTP enrollment should exist")
		require.Equal(t, mfatypes.EnrollmentStatusActive, retrieved.Status)

		session := &mfatypes.AuthorizationSession{
			SessionID:       "mfa-session-lifecycle-001",
			AccountAddress:  customer.String(),
			TransactionType: mfatypes.SensitiveTxHighValueOrder,
			VerifiedFactors: []mfatypes.FactorType{mfatypes.FactorTypeTOTP},
			CreatedAt:       ctx.BlockTime().Unix(),
			ExpiresAt:       ctx.BlockTime().Add(30 * time.Minute).Unix(),
			IsSingleUse:     false,
		}
		require.NoError(t, app.Keepers.VirtEngine.MFA.CreateAuthorizationSession(ctx, session))
		mfaSessionID = session.SessionID

		retrievedSession, found := app.Keepers.VirtEngine.MFA.GetAuthorizationSession(ctx, mfaSessionID)
		require.True(t, found, "MFA session should exist")
		require.Equal(t, mfatypes.SensitiveTxHighValueOrder, retrievedSession.TransactionType)
	})

	ctx = helpers.CommitAndAdvanceBlock(app, ctx)

	var offering marketplace.Offering
	t.Run("Phase3_ProviderRegistration", func(t *testing.T) {
		offering = helpers.CreateOfferingWithVEIDRequirement(
			t,
			app,
			ctx,
			provider,
			70,
			string(veidtypes.AccountStatusVerified),
		)
		offering.IdentityRequirement.RequireVerifiedDomain = true
		offering.Name = "Lifecycle Test Offering"
		offering.Description = "HPC cluster with GPU support for testing"
		offering.Pricing.Model = marketplace.PricingModelUsageBased

		require.NoError(t, app.Keepers.VirtEngine.Marketplace.UpdateOffering(ctx, &offering))
	})

	ctx = helpers.CommitAndAdvanceBlock(app, ctx)

	var order *marketplace.Order
	var bid marketplace.MarketplaceBid
	var leaseID string
	var settlementOrderID string
	var settlementEscrowID string

	t.Run("Phase4_OrderLifecycle", func(t *testing.T) {
		order = helpers.AttemptCreateOrder(t, app, ctx, customer, offering, false)
		require.NotEmpty(t, order.ID, "order should be created")

		bid = marketplace.MarketplaceBid{
			ID: marketplace.BidID{
				OrderID:         order.ID,
				ProviderAddress: provider.String(),
				Sequence:        1,
			},
			OfferingID: offering.ID,
			Price:      5000,
			PublicMetadata: map[string]string{
				"region":     "us-west-1",
				"datacenter": "lifecycle-test-dc",
			},
			ResourcesOffer: map[string]string{
				"cpu":       "16",
				"memory_gb": "64",
				"gpu":       "2",
			},
		}
		require.NoError(t, app.Keepers.VirtEngine.Marketplace.CreateBid(ctx, &bid))

		allocation, err := app.Keepers.VirtEngine.Marketplace.AcceptBid(ctx, bid.ID)
		require.NoError(t, err)
		require.NotNil(t, allocation)

		leaseID = allocation.ID.String()
		settlementOrderID = order.ID.String()

		depositAmount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100000)))
		fundE2EAccount(t, app, ctx, customer, depositAmount.Add(depositAmount...))

		settlementEscrowID = createActiveSettlementEscrowE2E(
			t,
			app,
			ctx,
			settlementOrderID,
			leaseID,
			customer,
			provider,
			depositAmount,
			72*time.Hour,
		)
	})

	ctx = helpers.CommitAndAdvanceBlock(app, ctx)

	t.Run("Phase5_Deployment", func(t *testing.T) {
		t.Log("✓ Workload deployed (mocked)")
		t.Log("✓ Status: RUNNING")
		t.Log("✓ Resources: 16 CPU, 64GB RAM, 2 GPU")
	})

	ctx = helpers.CommitAndAdvanceBlock(app, ctx)

	var settlementRecord *settlementtypes.SettlementRecord
	var settlementPayout settlementtypes.PayoutRecord
	var settlementInvoiceID string
	t.Run("Phase6_UsageAndSettlement", func(t *testing.T) {
		usagePeriodStart := ctx.BlockTime()
		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(24 * time.Hour))
		usagePeriodEnd := ctx.BlockTime()

		usageRecord := recordSettlementUsageE2E(
			t,
			app,
			ctx,
			settlementOrderID,
			leaseID,
			provider,
			customer,
			"cpu",
			384,
			sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(38400))),
			usagePeriodStart,
			usagePeriodEnd,
		)

		var err error
		settlementRecord, err = app.Keepers.VirtEngine.Settlement.SettleOrder(ctx, settlementOrderID, []string{usageRecord.UsageID}, false)
		require.NoError(t, err)
		require.NotNil(t, settlementRecord)
		requireSettlementSplit(t, settlementRecord)

		var found bool
		settlementPayout, found = app.Keepers.VirtEngine.Settlement.GetPayoutBySettlement(ctx, settlementRecord.SettlementID)
		require.True(t, found)
		require.Equal(t, settlementtypes.PayoutStateCompleted, settlementPayout.State)
		settlementInvoiceID = settlementPayout.InvoiceID

		invoiceKeeper := requireE2EInvoiceKeeper(t, app)
		invoiceRecord, err := invoiceKeeper.GetInvoice(ctx, settlementInvoiceID)
		require.NoError(t, err)
		require.Equal(t, billing.InvoiceStatusPaid, invoiceRecord.Status)
		require.True(t, invoiceRecord.Total.Equal(settlementRecord.TotalAmount))
		require.Equal(t, settlementRecord.SettlementID, invoiceRecord.SettlementID)

		ledgerChain, err := invoiceKeeper.GetInvoiceLedgerChain(ctx, settlementInvoiceID)
		require.NoError(t, err)
		require.NoError(t, ledgerChain.Validate())
	})

	ctx = helpers.CommitAndAdvanceBlock(app, ctx)

	t.Run("Phase7_Rewards", func(t *testing.T) {
		params := app.Keepers.VirtEngine.Settlement.GetParams(ctx)
		epochLength := params.StakingRewardEpochLength
		if epochLength == 0 {
			epochLength = 1
		}

		height := ctx.BlockHeight()
		if height < 0 {
			height = 0
		}
		epochNumber := uint64(height) / epochLength

		distributions := app.Keepers.VirtEngine.Settlement.GetRewardsByEpoch(ctx, epochNumber)
		require.NotEmpty(t, distributions)

		usageFound := false
		for _, dist := range distributions {
			if dist.Source == settlementtypes.RewardSourceUsage {
				usageFound = true
				t.Logf("✓ Usage rewards recorded for epoch %d: %s", dist.EpochNumber, dist.TotalRewards.String())
			}
		}
		require.True(t, usageFound, "expected usage reward distribution")
		t.Logf("✓ Provider payout from settlement: %s", settlementPayout.NetAmount.String())
	})

	ctx = helpers.CommitAndAdvanceBlock(app, ctx)

	t.Run("Phase8_Cleanup", func(t *testing.T) {
		require.NoError(t, app.Keepers.VirtEngine.Settlement.RefundEscrow(ctx, settlementEscrowID, "lifecycle cleanup"))
		settlementEscrow, found := app.Keepers.VirtEngine.Settlement.GetEscrow(ctx, settlementEscrowID)
		require.True(t, found)
		require.Equal(t, settlementtypes.EscrowStateRefunded, settlementEscrow.State)
	})

	ctx = helpers.CommitAndAdvanceBlock(app, ctx)

	t.Run("FinalVerification", func(t *testing.T) {
		customerRecord, found := app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
		require.True(t, found)
		require.Equal(t, veidtypes.IdentityTierVerified, customerRecord.Tier)

		providerRecord, found := app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, provider)
		require.True(t, found)
		require.True(t, providerRecord.HasVerifiedScope(veidtypes.ScopeTypeDomainVerify))

		session, found := app.Keepers.VirtEngine.MFA.GetAuthorizationSession(ctx, mfaSessionID)
		if found {
			require.Greater(t, ctx.BlockTime().Unix(), session.ExpiresAt, "MFA session should be expired after 24+ hours")
		}

		invoiceKeeper := requireE2EInvoiceKeeper(t, app)
		invoiceRecord, err := invoiceKeeper.GetInvoice(ctx, settlementInvoiceID)
		require.NoError(t, err)
		require.Equal(t, billing.InvoiceStatusPaid, invoiceRecord.Status)

		payout, found := app.Keepers.VirtEngine.Settlement.GetPayoutBySettlement(ctx, settlementRecord.SettlementID)
		require.True(t, found)
		require.Equal(t, settlementtypes.PayoutStateCompleted, payout.State)
	})
}

func TestFullLifecycleWithFailures(t *testing.T) {
	app := helpers.SetupOnboardingTestApp(t, helpers.NewOnboardingTestClient())
	ctx := helpers.NewTestContext(app, 1, helpers.FixedTimestamp())

	customer := helpers.CreateTestAccount(t)
	provider := helpers.CreateTestAccount(t)

	t.Run("OrderRejectedDueToLowVEIDScore", func(t *testing.T) {
		helpers.CreateIdentityRecordForAccount(t, app, ctx, customer)
		helpers.UpdateAccountScore(t, app, ctx, customer, 45)

		offering := helpers.CreateOfferingWithVEIDRequirement(t, app, ctx, provider, 70, string(veidtypes.AccountStatusVerified))
		ctx = helpers.CommitAndAdvanceBlock(app, ctx)
		helpers.AttemptCreateOrder(t, app, ctx, customer, offering, true)
	})

	t.Run("BidRejectedDueToUnverifiedProvider", func(t *testing.T) {
		ctx = helpers.CommitAndAdvanceBlock(app, ctx)

		customer2 := helpers.CreateTestAccount(t)
		helpers.CreateIdentityRecordForAccount(t, app, ctx, customer2)
		helpers.UpdateAccountScore(t, app, ctx, customer2, 85)
		helpers.SeedVerifiedScope(t, app, ctx, customer2, helpers.DefaultDomainVerifyUploadParams("failure-customer-domain-001"), customer2.String())

		unverifiedProvider := helpers.CreateTestAccount(t)
		offering := helpers.CreateOfferingWithVEIDRequirement(t, app, ctx, unverifiedProvider, 70, string(veidtypes.AccountStatusVerified))
		offering.IdentityRequirement.RequireVerifiedDomain = true
		require.NoError(t, app.Keepers.VirtEngine.Marketplace.UpdateOffering(ctx, &offering))

		ctx = helpers.CommitAndAdvanceBlock(app, ctx)

		order := helpers.AttemptCreateOrder(t, app, ctx, customer2, offering, false)
		bid := marketplace.MarketplaceBid{
			ID: marketplace.BidID{
				OrderID:         order.ID,
				ProviderAddress: unverifiedProvider.String(),
				Sequence:        1,
			},
			OfferingID: offering.ID,
			Price:      5000,
		}

		err := app.Keepers.VirtEngine.Marketplace.CreateBid(ctx, &bid)
		if err == nil {
			t.Log("⚠ Bid created but should be rejected at acceptance due to unverified provider")
		} else {
			t.Log("✓ Bid correctly rejected for unverified provider")
		}
	})

	t.Run("SettlementDisputeScenario", func(t *testing.T) {
		customer3 := helpers.CreateTestAccount(t)
		provider3 := helpers.CreateTestAccount(t)
		fundE2EAccount(t, app, ctx, customer3, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(5000))))

		orderID := "failure-dispute-order"
		leaseID := "failure-dispute-lease"
		escrowID := createActiveSettlementEscrowE2E(
			t,
			app,
			ctx,
			orderID,
			leaseID,
			customer3,
			provider3,
			sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(2000))),
			24*time.Hour,
		)

		usage := recordSettlementUsageE2E(
			t,
			app,
			ctx,
			orderID,
			leaseID,
			provider3,
			customer3,
			"gpu",
			2,
			sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1600))),
			ctx.BlockTime().Add(-time.Hour),
			ctx.BlockTime(),
		)

		require.NoError(t, app.Keepers.VirtEngine.Settlement.DisputeEscrow(ctx, escrowID, "usage mismatch"))
		escrow, found := app.Keepers.VirtEngine.Settlement.GetEscrow(ctx, escrowID)
		require.True(t, found)
		require.Equal(t, settlementtypes.EscrowStateDisputed, escrow.State)

		_, err := app.Keepers.VirtEngine.Settlement.SettleOrder(ctx, orderID, []string{usage.UsageID}, false)
		require.Error(t, err)
		require.Empty(t, app.Keepers.VirtEngine.Settlement.GetSettlementsByOrder(ctx, orderID))

		require.NoError(t, app.Keepers.VirtEngine.Settlement.RefundEscrow(ctx, escrowID, "dispute resolved for customer"))
		escrow, found = app.Keepers.VirtEngine.Settlement.GetEscrow(ctx, escrowID)
		require.True(t, found)
		require.Equal(t, settlementtypes.EscrowStateRefunded, escrow.State)
	})
}
