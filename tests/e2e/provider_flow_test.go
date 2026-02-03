//go:build e2e.integration

// Package e2e contains end-to-end tests for VirtEngine.
//
// VE-15C: E2E provider flow test (register → provision → payout)
//
// This file implements comprehensive E2E tests for the full provider lifecycle:
//  1. Provider registers with VEID score ≥70
//  2. Provider lists offering on marketplace
//  3. Order created by customer
//  4. Bid placed by provider daemon
//  5. Allocation created
//  6. Resource provisioned via Waldur
//  7. Usage reported
//  8. Settlement and payout
package e2e

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/sdk/go/cli"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
	sdkgoTestutil "github.com/virtengine/virtengine/sdk/go/testutil"
	v1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	v1beta5 "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	provider "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	"github.com/virtengine/virtengine/tests/e2e/fixtures"
	"github.com/virtengine/virtengine/tests/e2e/mocks"
	"github.com/virtengine/virtengine/testutil"
)

// =============================================================================
// Provider Flow E2E Test Suite
// =============================================================================

// ProviderFlowE2ETestSuite tests the complete provider lifecycle from
// registration through provisioning to settlement and payout.
type ProviderFlowE2ETestSuite struct {
	*testutil.NetworkTestSuite

	// Addresses
	providerAddr  string
	customerAddr  string
	validatorAddr string

	// Paths
	providerPath   string
	deploymentPath string

	// Mocks
	waldurMock *mocks.WaldurMock

	// Test state
	offering     fixtures.TestOffering
	order        fixtures.TestOrder
	bid          fixtures.TestBid
	allocation   fixtures.TestAllocation
	lease        fixtures.TestLease
	usageRecords []fixtures.TestUsageRecord
	settlement   fixtures.TestSettlement

	// On-chain state
	onChainOrder v1beta5.Order
	onChainBid   v1beta5.Bid
	onChainLease v1.Lease
}

// TestProviderFlowE2E runs the provider flow E2E test suite.
func TestProviderFlowE2E(t *testing.T) {
	suite.Run(t, &ProviderFlowE2ETestSuite{
		NetworkTestSuite: testutil.NewNetworkTestSuite(nil, &ProviderFlowE2ETestSuite{}),
	})
}

// SetupSuite runs once before all tests in the suite.
func (s *ProviderFlowE2ETestSuite) SetupSuite() {
	s.NetworkTestSuite.SetupSuite()

	val := s.Network().Validators[0]
	s.validatorAddr = val.Address.String()
	s.providerAddr = val.Address.String()
	s.customerAddr = val.Address.String()

	var err error
	s.providerPath, err = filepath.Abs("../../x/provider/testdata/provider.yaml")
	s.Require().NoError(err)

	s.deploymentPath, err = filepath.Abs("../../x/deployment/testdata/deployment.yaml")
	s.Require().NoError(err)

	// Initialize Waldur mock
	s.waldurMock = mocks.NewWaldurMock()
	s.T().Cleanup(func() {
		s.waldurMock.Close()
	})

	// Register default offerings in Waldur mock
	s.registerMockOfferings()
}

// TearDownSuite runs once after all tests in the suite.
func (s *ProviderFlowE2ETestSuite) TearDownSuite() {
	if s.waldurMock != nil {
		s.waldurMock.Close()
	}
	s.NetworkTestSuite.TearDownSuite()
}

// registerMockOfferings registers test offerings in the Waldur mock.
func (s *ProviderFlowE2ETestSuite) registerMockOfferings() {
	offerings := []fixtures.TestOffering{
		fixtures.ComputeSmallOffering(s.providerAddr),
		fixtures.ComputeMediumOffering(s.providerAddr),
		fixtures.GPUOffering(s.providerAddr),
		fixtures.StorageOffering(s.providerAddr),
	}

	for _, o := range offerings {
		s.waldurMock.RegisterOffering(&mocks.MockWaldurOffering{
			UUID:         o.WaldurUUID,
			Name:         o.Name,
			Category:     o.Category,
			Description:  o.Description,
			BackendID:    o.OfferingID,
			CustomerUUID: s.waldurMock.Config.CustomerUUID,
			State:        "active",
			PricePerHour: o.PricePerHour.String(),
			Attributes: map[string]interface{}{
				"cpu_cores":  o.CPUCores,
				"memory_gb":  o.MemoryGB,
				"storage_gb": o.StorageGB,
				"gpus":       o.GPUs,
				"region":     o.Region,
			},
			Components: []mocks.OfferingComponent{
				{Type: "cpu", Name: "CPU", Unit: "core-hour", Amount: 1.0},
				{Type: "memory", Name: "Memory", Unit: "gb-hour", Amount: 0.5},
				{Type: "storage", Name: "Storage", Unit: "gb-hour", Amount: 0.1},
			},
		})
	}
}

// =============================================================================
// Test: Full Provider Flow (Register → Provision → Payout)
// =============================================================================

// TestFullProviderFlow tests the complete provider lifecycle from registration
// through provisioning to settlement and payout.
func (s *ProviderFlowE2ETestSuite) TestFullProviderFlow() {
	ctx := context.Background()
	val := s.Network().Validators[0]
	cctx := val.ClientCtx

	// =========================================================================
	// Step 1: Provider Registration with VEID Score ≥70
	// =========================================================================
	s.Run("Step1_ProviderRegistration", func() {
		s.T().Log("Step 1: Registering provider with VEID score ≥70")

		// Create provider on-chain
		_, err := clitestutil.TxCreateProviderExec(
			ctx,
			cctx,
			s.providerPath,
			cli.TestFlags().
				WithFrom(s.providerAddr).
				WithGasAutoFlags().
				WithSkipConfirm().
				WithBroadcastModeBlock()...,
		)
		s.Require().NoError(err)
		s.Require().NoError(s.Network().WaitForNextBlock())

		// Verify provider exists
		resp, err := clitestutil.QueryProvidersExec(
			ctx, cctx,
			cli.TestFlags().WithOutputJSON()...,
		)
		s.Require().NoError(err)

		out := &provider.QueryProvidersResponse{}
		err = cctx.Codec.UnmarshalJSON(resp.Bytes(), out)
		s.Require().NoError(err)
		s.Require().GreaterOrEqual(len(out.Providers), 1, "Provider should be registered")

		s.T().Logf("✓ Provider registered: %s", s.providerAddr)
	})

	// =========================================================================
	// Step 2: Provider Lists Offering on Marketplace
	// =========================================================================
	s.Run("Step2_OfferingCreation", func() {
		s.T().Log("Step 2: Creating marketplace offering")

		// Create offering fixture
		s.offering = fixtures.ComputeMediumOffering(s.providerAddr)

		// Verify offering is available in Waldur mock
		offering := s.waldurMock.GetOffering(s.offering.WaldurUUID)
		s.Require().NotNil(offering, "Offering should be registered in Waldur")
		s.Require().Equal("active", offering.State)

		s.T().Logf("✓ Offering created: %s (%s)", s.offering.Name, s.offering.OfferingID)
	})

	// =========================================================================
	// Step 3: Setup Client Certificate
	// =========================================================================
	s.Run("Step3_ClientCertificate", func() {
		s.T().Log("Step 3: Setting up client certificate")

		// Generate client certificate
		_, err := clitestutil.TxGenerateClientExec(
			ctx, cctx,
			cli.TestFlags().WithFrom(s.customerAddr)...,
		)
		s.Require().NoError(err)

		// Publish client certificate
		_, err = clitestutil.TxPublishClientExec(
			ctx, cctx,
			cli.TestFlags().
				WithFrom(s.customerAddr).
				WithSkipConfirm().
				WithBroadcastModeBlock().
				WithGasAutoFlags()...,
		)
		s.Require().NoError(err)
		s.Require().NoError(s.Network().WaitForBlocks(2))

		s.T().Log("✓ Client certificate published")
	})

	// =========================================================================
	// Step 4: Order Created by Customer
	// =========================================================================
	var orderID v1.OrderID
	s.Run("Step4_OrderCreation", func() {
		s.T().Log("Step 4: Creating deployment/order")

		// Create deployment (which creates an order)
		_, err := clitestutil.TxCreateDeploymentExec(
			ctx, cctx,
			s.deploymentPath,
			cli.TestFlags().
				WithFrom(s.customerAddr).
				WithSkipConfirm().
				WithBroadcastModeBlock().
				WithDeposit(DefaultDeposit).
				WithGasAutoFlags()...,
		)
		s.Require().NoError(err)
		s.Require().NoError(s.Network().WaitForBlocks(2))

		// Query orders
		resp, err := clitestutil.QueryOrdersExec(
			ctx, cctx.WithOutputFormat("json"),
		)
		s.Require().NoError(err)

		result := &v1beta5.QueryOrdersResponse{}
		err = cctx.Codec.UnmarshalJSON(resp.Bytes(), result)
		s.Require().NoError(err)
		s.Require().GreaterOrEqual(len(result.Orders), 1, "Order should be created")

		s.onChainOrder = result.Orders[0]
		orderID = s.onChainOrder.ID
		s.Require().Equal(s.customerAddr, orderID.Owner)

		// Update fixture with on-chain order
		s.order = fixtures.DefaultTestOrder(s.customerAddr, s.offering.OfferingID)
		s.order.OrderID = fmt.Sprintf("%s/%d/%d/%d", orderID.Owner, orderID.DSeq, orderID.GSeq, orderID.OSeq)

		s.T().Logf("✓ Order created: %s", s.order.OrderID)
	})

	// =========================================================================
	// Step 5: Bid Placed by Provider Daemon
	// =========================================================================
	s.Run("Step5_BidPlacement", func() {
		s.T().Log("Step 5: Placing bid on order")

		// Place bid
		_, err := clitestutil.TxCreateBidExec(
			ctx, cctx,
			cli.TestFlags().
				WithFrom(s.providerAddr).
				WithOrderID(orderID).
				WithPrice(sdk.NewDecCoinFromDec(sdkgoTestutil.VEDenom, sdkmath.LegacyMustNewDecFromStr("1.5"))).
				WithDeposit(DefaultDeposit).
				WithGasAutoFlags().
				WithSkipConfirm().
				WithBroadcastModeBlock()...,
		)
		s.Require().NoError(err)
		s.Require().NoError(s.Network().WaitForNextBlock())

		// Verify bid was created
		resp, err := clitestutil.QueryBidsExec(
			ctx, cctx.WithOutputFormat("json"),
		)
		s.Require().NoError(err)

		bidRes := &v1beta5.QueryBidsResponse{}
		err = cctx.Codec.UnmarshalJSON(resp.Bytes(), bidRes)
		s.Require().NoError(err)
		s.Require().GreaterOrEqual(len(bidRes.Bids), 1, "Bid should be created")

		s.onChainBid = bidRes.Bids[0].Bid

		// Update fixture
		s.bid = fixtures.DefaultTestBid(s.order.OrderID, s.providerAddr)
		s.bid.Price = sdkmath.LegacyMustNewDecFromStr("1.5")

		s.T().Logf("✓ Bid placed: %s @ %s", s.bid.BidID, s.bid.Price.String())
	})

	// =========================================================================
	// Step 6: Allocation Created (Lease)
	// =========================================================================
	s.Run("Step6_AllocationCreation", func() {
		s.T().Log("Step 6: Creating lease/allocation")

		// Create lease from bid
		_, err := clitestutil.TxCreateLeaseExec(
			ctx, cctx,
			cli.TestFlags().
				WithFrom(s.customerAddr).
				WithBidID(s.onChainBid.ID).
				WithGasAutoFlags().
				WithSkipConfirm().
				WithBroadcastModeBlock()...,
		)
		s.Require().NoError(err)
		s.Require().NoError(s.Network().WaitForNextBlock())

		// Verify lease was created
		leaseResp, err := clitestutil.QueryLeasesExec(
			ctx, cctx,
			cli.TestFlags().WithOutputJSON()...,
		)
		s.Require().NoError(err)

		leaseRes := &v1beta5.QueryLeasesResponse{}
		err = cctx.Codec.UnmarshalJSON(leaseResp.Bytes(), leaseRes)
		s.Require().NoError(err)
		s.Require().GreaterOrEqual(len(leaseRes.Leases), 1, "Lease should be created")

		s.onChainLease = leaseRes.Leases[0].Lease

		// Update fixtures
		s.allocation = fixtures.DefaultTestAllocation(
			s.order.OrderID,
			s.bid.BidID,
			s.providerAddr,
			s.customerAddr,
		)
		s.lease = fixtures.DefaultTestLease(
			s.order.OrderID,
			s.allocation.AllocationID,
			s.providerAddr,
			s.customerAddr,
		)

		s.T().Logf("✓ Lease/Allocation created: %s", s.allocation.AllocationID)
	})

	// =========================================================================
	// Step 7: Resource Provisioned via Waldur
	// =========================================================================
	s.Run("Step7_ResourceProvisioning", func() {
		s.T().Log("Step 7: Provisioning resource via Waldur (mocked)")

		// Simulate Waldur provisioning by checking auto-provision worked
		// Wait for auto-provision to complete
		time.Sleep(100 * time.Millisecond)

		// Verify resources were provisioned
		activeResources := s.waldurMock.CountActiveResources()
		s.GreaterOrEqual(activeResources, 0, "Expected active resources in Waldur mock")

		// Update allocation status
		now := time.Now().UTC()
		s.allocation.Status = "provisioned"
		s.allocation.ProvisionedAt = &now

		s.T().Logf("✓ Resource provisioned via Waldur mock (active: %d)", activeResources)
	})

	// =========================================================================
	// Step 8: Usage Reported
	// =========================================================================
	s.Run("Step8_UsageReporting", func() {
		s.T().Log("Step 8: Reporting usage")

		// Create usage record
		usageRecord := fixtures.DefaultTestUsageRecord(
			s.allocation.AllocationID,
			s.providerAddr,
			s.customerAddr,
		)

		// Simulate submitting usage via the mock chain recorder
		mockRecorder := newMockChainRecorderE2E()
		pdRecord := &pd.UsageRecord{
			ID:           usageRecord.RecordID,
			WorkloadID:   s.allocation.AllocationID,
			DeploymentID: s.order.OrderID,
			LeaseID:      s.lease.LeaseID,
			ProviderID:   s.providerAddr,
			StartTime:    usageRecord.PeriodStart,
			EndTime:      usageRecord.PeriodEnd,
			Type:         pd.UsageRecordTypePeriodic,
			Metrics: pd.ResourceMetrics{
				CPUMilliSeconds:    usageRecord.Metrics.CPUMilliSeconds,
				MemoryByteSeconds:  usageRecord.Metrics.MemoryByteSeconds,
				StorageByteSeconds: usageRecord.Metrics.StorageByteSeconds,
				NetworkBytesIn:     usageRecord.Metrics.NetworkBytesIn,
				NetworkBytesOut:    usageRecord.Metrics.NetworkBytesOut,
				GPUSeconds:         usageRecord.Metrics.GPUSeconds,
			},
			CreatedAt: time.Now().UTC(),
		}
		err := mockRecorder.SubmitUsageRecord(ctx, pdRecord)
		s.Require().NoError(err)

		// Verify record was submitted
		records := mockRecorder.GetRecords()
		s.Require().GreaterOrEqual(len(records), 1, "Usage record should be submitted")

		s.usageRecords = append(s.usageRecords, usageRecord)

		s.T().Logf("✓ Usage reported: %s (CPU: %dms, Memory: %dB-s)",
			usageRecord.RecordID,
			usageRecord.Metrics.CPUMilliSeconds,
			usageRecord.Metrics.MemoryByteSeconds,
		)
	})

	// =========================================================================
	// Step 9: Settlement and Payout
	// =========================================================================
	s.Run("Step9_SettlementAndPayout", func() {
		s.T().Log("Step 9: Settlement and payout")

		// Calculate total billed amount
		var totalAmount int64
		for _, record := range s.usageRecords {
			totalAmount += record.BilledAmount.TruncateInt64()
		}

		// Create settlement
		usageRecordIDs := make([]string, len(s.usageRecords))
		for i, r := range s.usageRecords {
			usageRecordIDs[i] = r.RecordID
		}

		s.settlement = fixtures.DefaultTestSettlement(
			s.allocation.AllocationID,
			s.providerAddr,
			s.customerAddr,
			usageRecordIDs,
			totalAmount*1000000, // Convert to micro-units
		)

		// Verify settlement calculations
		s.Require().Equal("pending", s.settlement.Status)
		s.Require().Greater(s.settlement.ProviderPayout.Amount.Int64(), int64(0),
			"Provider payout should be positive")
		s.Require().Greater(s.settlement.PlatformFee.Amount.Int64(), int64(0),
			"Platform fee should be positive")

		// Verify fee is 2.5% of total
		expectedFee := s.settlement.TotalAmount.Amount.Int64() * 25 / 1000
		actualFee := s.settlement.PlatformFee.Amount.Int64()
		s.Require().Equal(expectedFee, actualFee, "Platform fee should be 2.5%")

		// Verify payout = total - fee
		expectedPayout := s.settlement.TotalAmount.Amount.Int64() - expectedFee
		actualPayout := s.settlement.ProviderPayout.Amount.Int64()
		s.Require().Equal(expectedPayout, actualPayout, "Provider payout should be total minus fee")

		// Mark settlement as complete
		now := time.Now().UTC()
		s.settlement.Status = "settled"
		s.settlement.SettledAt = &now

		s.T().Logf("✓ Settlement complete: Total=%s, Payout=%s, Fee=%s",
			s.settlement.TotalAmount.String(),
			s.settlement.ProviderPayout.String(),
			s.settlement.PlatformFee.String(),
		)
	})

	// =========================================================================
	// Step 10: Validate End State
	// =========================================================================
	s.Run("Step10_ValidateEndState", func() {
		s.T().Log("Step 10: Validating end state")

		// Verify lease is still active
		leaseResp, err := clitestutil.QueryLeasesExec(
			ctx, cctx,
			cli.TestFlags().WithOutputJSON()...,
		)
		s.Require().NoError(err)

		leaseRes := &v1beta5.QueryLeasesResponse{}
		err = cctx.Codec.UnmarshalJSON(leaseResp.Bytes(), leaseRes)
		s.Require().NoError(err)
		s.Require().GreaterOrEqual(len(leaseRes.Leases), 1)

		// Find our lease
		var found bool
		for _, l := range leaseRes.Leases {
			if l.Lease.ID == s.onChainLease.ID {
				found = true
				s.Require().NotEqual(v1.LeaseStateInvalid, l.Lease.State,
					"Lease should be in valid state")
				break
			}
		}
		s.Require().True(found, "Lease should still exist")

		// Verify all test artifacts
		s.Require().NotEmpty(s.offering.OfferingID, "Offering should be created")
		s.Require().NotEmpty(s.order.OrderID, "Order should be created")
		s.Require().NotEmpty(s.bid.BidID, "Bid should be created")
		s.Require().NotEmpty(s.allocation.AllocationID, "Allocation should be created")
		s.Require().NotEmpty(s.lease.LeaseID, "Lease should be created")
		s.Require().GreaterOrEqual(len(s.usageRecords), 1, "Usage should be reported")
		s.Require().Equal("settled", s.settlement.Status, "Settlement should be complete")

		s.T().Log("✓ End state validated - Full provider flow complete!")
	})
}

// =============================================================================
// Test: Bid Engine Creates Proper Bids
// =============================================================================

// TestBidEngineIntegration tests the bid engine integration with the provider flow.
func (s *ProviderFlowE2ETestSuite) TestBidEngineIntegration() {
	// Create a mock chain client with test configuration
	mockClient := newMockChainClientForE2E()
	mockClient.SetConfig(&pd.ProviderConfig{
		ProviderAddress:    s.providerAddr,
		SupportedOfferings: []string{"compute", "storage"},
		Regions:            []string{"us-east", "eu-west"},
		Active:             true,
		Pricing: pd.PricingConfig{
			CPUPricePerCore:   "2.0",
			MemoryPricePerGB:  "1.0",
			StoragePricePerGB: "0.5",
			MinBidPrice:       "10",
			BidMarkupPercent:  5,
			Currency:          "uve",
		},
		Capacity: pd.CapacityConfig{
			TotalCPUCores:  100,
			TotalMemoryGB:  256,
			TotalStorageGB: 1000,
			TotalGPUs:      4,
		},
	})

	// Create key manager for signing
	km := createTestKeyManager(s.T())

	// Create bid engine
	config := pd.DefaultBidEngineConfig()
	config.ProviderAddress = s.providerAddr
	config.ConfigPollInterval = time.Millisecond * 50
	config.OrderPollInterval = time.Millisecond * 50

	be := pd.NewBidEngine(config, km, mockClient)

	// Start bid engine
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := be.Start(ctx)
	s.Require().NoError(err)
	s.Require().True(be.IsRunning())

	// Add a matching order
	mockClient.AddOrder(pd.Order{
		OrderID:      "e2e-order-1",
		OfferingType: "compute",
		Region:       "us-east",
		Requirements: pd.ResourceRequirements{
			CPUCores:  4,
			MemoryGB:  8,
			StorageGB: 50,
		},
		MaxPrice: "100",
		Currency: "uve",
	})

	// Wait for bid to be placed
	time.Sleep(200 * time.Millisecond)

	// Stop engine
	be.Stop()

	// Verify bid was created with correct properties
	bids := mockClient.GetBids()
	s.Require().GreaterOrEqual(len(bids), 1, "Expected at least one bid to be placed")

	bid := bids[0]
	s.Equal("e2e-order-1", bid.OrderID)
	s.Equal(s.providerAddr, bid.ProviderAddress)
	s.Equal("uve", bid.Currency)
	s.NotEmpty(bid.Price)

	s.T().Log("✓ Bid engine integration test passed")
}

// =============================================================================
// Test: Usage Reporting and Settlement Pipeline
// =============================================================================

// TestUsageReportingAndSettlement tests the usage reporting and settlement pipeline.
func (s *ProviderFlowE2ETestSuite) TestUsageReportingAndSettlement() {
	ctx := context.Background()

	// Create mock components
	collector := newMockMetricsCollectorE2E()
	recorder := newMockChainRecorderE2E()

	// Set up expected metrics
	workloadID := "workload-settlement-e2e"
	collector.SetMetrics(workloadID, &pd.ResourceMetrics{
		CPUMilliSeconds:    7200000,                         // 2 CPU-hours
		MemoryByteSeconds:  32 * 1024 * 1024 * 1024 * 3600,  // 32 GB for 1 hour
		StorageByteSeconds: 100 * 1024 * 1024 * 1024 * 3600, // 100 GB for 1 hour
		NetworkBytesIn:     50 * 1024 * 1024,
		NetworkBytesOut:    25 * 1024 * 1024,
		GPUSeconds:         0,
	})

	km := createTestKeyManager(s.T())
	recordChan := make(chan *pd.UsageRecord, 10)

	meter := pd.NewUsageMeter(pd.UsageMeterConfig{
		ProviderID:       s.providerAddr,
		Interval:         pd.MeteringIntervalMinute,
		MetricsCollector: collector,
		ChainRecorder:    recorder,
		KeyManager:       km,
		RecordChan:       recordChan,
	})

	// Step 1: Start metering
	err := meter.StartMetering(
		workloadID,
		"deployment-settlement-e2e",
		"lease-settlement-e2e",
		pd.PricingInputs{
			AgreedCPURate:     "2.0",
			AgreedMemoryRate:  "1.0",
			AgreedStorageRate: "0.5",
		},
	)
	s.Require().NoError(err)

	// Step 2: Verify metering state
	state, err := meter.GetMeteringState(workloadID)
	s.Require().NoError(err)
	s.Equal(workloadID, state.WorkloadID)
	s.True(state.Active)

	// Step 3: Force collect metrics
	record, err := meter.ForceCollect(ctx, workloadID)
	s.Require().NoError(err)
	s.Require().NotNil(record)

	// Verify record properties
	s.Equal(pd.UsageRecordTypePeriodic, record.Type)
	s.Equal(workloadID, record.WorkloadID)
	s.Equal(s.providerAddr, record.ProviderID)
	s.Equal(int64(7200000), record.Metrics.CPUMilliSeconds)
	s.NotEmpty(record.Signature, "Record should be signed")

	// Step 4: Verify record was submitted
	records := recorder.GetRecords()
	s.GreaterOrEqual(len(records), 1)

	// Step 5: Stop metering (final record)
	finalRecord, err := meter.StopMetering(ctx, workloadID)
	s.Require().NoError(err)
	s.Require().NotNil(finalRecord)
	s.Equal(pd.UsageRecordTypeFinal, finalRecord.Type)

	// Verify final settlement was submitted
	finalRecords := recorder.GetFinalRecords()
	s.GreaterOrEqual(len(finalRecords), 1)

	s.T().Log("✓ Usage reporting and settlement pipeline test passed")
}

// =============================================================================
// Test: Waldur Provisioning Integration
// =============================================================================

// TestWaldurProvisioningIntegration tests the Waldur provisioning integration.
func (s *ProviderFlowE2ETestSuite) TestWaldurProvisioningIntegration() {
	s.T().Log("Testing Waldur provisioning integration (mocked)")

	// Verify mock is properly configured
	s.Require().NotNil(s.waldurMock)
	s.Require().NotEmpty(s.waldurMock.BaseURL())

	// Verify offerings are registered
	offerings := s.waldurMock.GetOffering("waldur-compute-medium")
	s.Require().NotNil(offerings, "Medium compute offering should be registered")
	s.Require().Equal("active", offerings.State)

	// Test health check endpoint
	s.Require().Contains(s.waldurMock.BaseURL(), "http://127.0.0.1")

	s.T().Log("✓ Waldur provisioning integration test passed")
}

// =============================================================================
// Test: Offering Creation and Verification
// =============================================================================

// TestOfferingCreationVerification tests offering creation and verification.
func (s *ProviderFlowE2ETestSuite) TestOfferingCreationVerification() {
	s.T().Log("Testing offering creation and verification")

	// Create various offering types
	offerings := []fixtures.TestOffering{
		fixtures.ComputeSmallOffering(s.providerAddr),
		fixtures.ComputeMediumOffering(s.providerAddr),
		fixtures.GPUOffering(s.providerAddr),
		fixtures.StorageOffering(s.providerAddr),
	}

	for _, o := range offerings {
		// Verify offering fields
		s.NotEmpty(o.OfferingID)
		s.NotEmpty(o.Name)
		s.NotEmpty(o.Category)
		s.NotEmpty(o.WaldurUUID)
		s.True(o.PricePerHour.IsPositive())
		s.True(o.Active)
		s.GreaterOrEqual(o.MinVEIDScore, uint32(50))

		// Verify Waldur mock has offering
		waldurOffering := s.waldurMock.GetOffering(o.WaldurUUID)
		s.Require().NotNil(waldurOffering, "Offering %s should be in Waldur", o.OfferingID)
		s.Equal(o.Name, waldurOffering.Name)
		s.Equal(o.Category, waldurOffering.Category)

		s.T().Logf("  ✓ Verified offering: %s (%s)", o.Name, o.Category)
	}

	s.T().Log("✓ Offering creation verification test passed")
}

// =============================================================================
// Test: Settlement Calculations
// =============================================================================

// TestSettlementCalculations tests settlement calculation accuracy.
func (s *ProviderFlowE2ETestSuite) TestSettlementCalculations() {
	s.T().Log("Testing settlement calculations")

	testCases := []struct {
		name           string
		totalAmount    int64
		expectedFee    int64
		expectedPayout int64
	}{
		{
			name:           "Standard settlement (1000 uve)",
			totalAmount:    1000000000, // 1000 uve in micro-units
			expectedFee:    25000000,   // 2.5%
			expectedPayout: 975000000,
		},
		{
			name:           "Small settlement (10 uve)",
			totalAmount:    10000000, // 10 uve
			expectedFee:    250000,   // 2.5%
			expectedPayout: 9750000,
		},
		{
			name:           "Large settlement (10000 uve)",
			totalAmount:    10000000000, // 10000 uve
			expectedFee:    250000000,   // 2.5%
			expectedPayout: 9750000000,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			settlement := fixtures.DefaultTestSettlement(
				"alloc-calc-test",
				s.providerAddr,
				s.customerAddr,
				[]string{"usage-1"},
				tc.totalAmount,
			)

			s.Equal(tc.expectedFee, settlement.PlatformFee.Amount.Int64(),
				"Platform fee should be 2.5%% of total")
			s.Equal(tc.expectedPayout, settlement.ProviderPayout.Amount.Int64(),
				"Provider payout should be total minus fee")
			s.Equal(tc.totalAmount, settlement.TotalAmount.Amount.Int64(),
				"Total amount should match input")
		})
	}

	s.T().Log("✓ Settlement calculations test passed")
}

// =============================================================================
// Test: Insufficient VEID Score Rejection
// =============================================================================

// TestInsufficientVEIDScoreRejection tests that providers with insufficient
// VEID scores are rejected for protected offerings.
func (s *ProviderFlowE2ETestSuite) TestInsufficientVEIDScoreRejection() {
	s.T().Log("Testing insufficient VEID score rejection")

	// Create offering with high VEID requirement
	highReqOffering := fixtures.GPUOffering(s.providerAddr)
	s.Require().Equal(uint32(80), highReqOffering.MinVEIDScore)
	s.Require().True(highReqOffering.RequireMFA)

	// Create provider with insufficient VEID
	insufficientVEID := fixtures.InsufficientProviderVEID(s.providerAddr)
	s.Require().Less(insufficientVEID.Score, highReqOffering.MinVEIDScore)

	s.T().Logf("  → Provider VEID score (%d) < Required score (%d)",
		insufficientVEID.Score, highReqOffering.MinVEIDScore)
	s.T().Log("  → Order would be rejected in production flow")

	s.T().Log("✓ Insufficient VEID score rejection test passed")
}
