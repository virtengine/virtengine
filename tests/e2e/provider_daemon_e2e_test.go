//go:build e2e.integration

package e2e

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/sdk/go/cli"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
	v1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	v1beta5 "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	provider "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	"github.com/virtengine/virtengine/testutil"
)

// providerDaemonE2ETestSuite tests the full provider daemon workflow:
// Provider register → Order created → Bid placed → Allocation → Provision → Usage report
type providerDaemonE2ETestSuite struct {
	*testutil.NetworkTestSuite

	providerAddr  string
	deployerAddr  string
	providerPath  string
	deploymentPath string
}

func TestProviderDaemonE2E(t *testing.T) {
	suite.Run(t, &providerDaemonE2ETestSuite{
		NetworkTestSuite: testutil.NewNetworkTestSuite(nil, &providerDaemonE2ETestSuite{}),
	})
}

func (s *providerDaemonE2ETestSuite) SetupSuite() {
	s.NetworkTestSuite.SetupSuite()

	val := s.Network().Validators[0]
	s.providerAddr = val.Address.String()
	s.deployerAddr = val.Address.String()

	var err error
	s.providerPath, err = filepath.Abs("../../x/provider/testdata/provider.yaml")
	s.Require().NoError(err)

	s.deploymentPath, err = filepath.Abs("../../x/deployment/testdata/deployment.yaml")
	s.Require().NoError(err)
}

// TestProviderDaemonFullFlow tests the complete provider daemon workflow
func (s *providerDaemonE2ETestSuite) TestProviderDaemonFullFlow() {
	ctx := context.Background()
	val := s.Network().Validators[0]
	cctx := val.ClientCtx

	// Step 1: Register provider on-chain
	s.Run("RegisterProvider", func() {
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
		s.Require().GreaterOrEqual(len(out.Providers), 1)
	})

	// Step 2: Generate and publish client certificate
	s.Run("SetupClientCertificate", func() {
		_, err := clitestutil.TxGenerateClientExec(
			ctx, cctx,
			cli.TestFlags().WithFrom(s.deployerAddr)...,
		)
		s.Require().NoError(err)

		_, err = clitestutil.TxPublishClientExec(
			ctx, cctx,
			cli.TestFlags().
				WithFrom(s.deployerAddr).
				WithSkipConfirm().
				WithBroadcastModeBlock().
				WithGasAutoFlags()...,
		)
		s.Require().NoError(err)
		s.Require().NoError(s.Network().WaitForBlocks(2))
	})

	// Step 3: Create deployment (generates order)
	var orderID v1beta5.OrderID
	s.Run("CreateDeployment", func() {
		_, err := clitestutil.TxCreateDeploymentExec(
			ctx, cctx,
			s.deploymentPath,
			cli.TestFlags().
				WithFrom(s.deployerAddr).
				WithSkipConfirm().
				WithBroadcastModeBlock().
				WithDeposit(DefaultDeposit).
				WithGasAutoFlags()...,
		)
		s.Require().NoError(err)
		s.Require().NoError(s.Network().WaitForBlocks(2))

		// Verify order was created
		resp, err := clitestutil.QueryOrdersExec(
			ctx, cctx.WithOutputFormat("json"),
		)
		s.Require().NoError(err)

		result := &v1beta5.QueryOrdersResponse{}
		err = cctx.Codec.UnmarshalJSON(resp.Bytes(), result)
		s.Require().NoError(err)
		s.Require().GreaterOrEqual(len(result.Orders), 1)

		orderID = result.Orders[0].ID
		s.Require().Equal(s.deployerAddr, orderID.Owner)
	})

	// Step 4: Place bid on order (simulating bid engine behavior)
	s.Run("PlaceBid", func() {
		_, err := clitestutil.TxCreateBidExec(
			ctx, cctx,
			cli.TestFlags().
				WithFrom(s.providerAddr).
				WithOrderID(orderID).
				WithPrice(sdk.NewDecCoinFromDec(testutil.CoinDenom, sdk.MustNewDecFromStr("1.5"))).
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
		s.Require().GreaterOrEqual(len(bidRes.Bids), 1)
	})

	// Step 5: Accept bid and create lease (allocation)
	var lease v1.Lease
	s.Run("CreateLease", func() {
		// Get the bid
		resp, err := clitestutil.QueryBidsExec(
			ctx, cctx.WithOutputFormat("json"),
		)
		s.Require().NoError(err)

		bidRes := &v1beta5.QueryBidsResponse{}
		err = cctx.Codec.UnmarshalJSON(resp.Bytes(), bidRes)
		s.Require().NoError(err)
		s.Require().GreaterOrEqual(len(bidRes.Bids), 1)

		bid := bidRes.Bids[0].Bid

		// Create lease from bid
		_, err = clitestutil.TxCreateLeaseExec(
			ctx, cctx,
			cli.TestFlags().
				WithFrom(s.deployerAddr).
				WithBidID(bid.ID).
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
		s.Require().GreaterOrEqual(len(leaseRes.Leases), 1)

		lease = leaseRes.Leases[0].Lease
	})

	// Step 6: Validate lease is active (provisioning state)
	s.Run("ValidateLeaseActive", func() {
		leaseResp, err := clitestutil.QueryLeasesExec(
			ctx, cctx,
			cli.TestFlags().WithOutputJSON()...,
		)
		s.Require().NoError(err)

		leaseRes := &v1beta5.QueryLeasesResponse{}
		err = cctx.Codec.UnmarshalJSON(leaseResp.Bytes(), leaseRes)
		s.Require().NoError(err)
		s.Require().GreaterOrEqual(len(leaseRes.Leases), 1)

		// Verify lease is in active state
		for _, l := range leaseRes.Leases {
			if l.Lease.ID == lease.ID {
				s.Require().NotEqual(v1.LeaseStateInvalid, l.Lease.State)
				break
			}
		}
	})
}

// TestBidEngineCreatesProperBids tests the bid engine bid creation logic
func (s *providerDaemonE2ETestSuite) TestBidEngineCreatesProperBids() {
	// Create a mock chain client with test configuration
	mockClient := newMockChainClientForE2E()
	mockClient.SetConfig(&pd.ProviderConfig{
		ProviderAddress:    "test-provider",
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
	config.ProviderAddress = "test-provider"
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
		OrderID:      "test-order-1",
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
	s.Equal("test-order-1", bid.OrderID)
	s.Equal("test-provider", bid.ProviderAddress)
	s.Equal("uve", bid.Currency)
	s.Equal("open", bid.State)
	s.NotEmpty(bid.Price)
}

// TestBidEngineRejectsNonMatchingOrders tests order matching logic
func (s *providerDaemonE2ETestSuite) TestBidEngineRejectsNonMatchingOrders() {
	mockClient := newMockChainClientForE2E()
	mockClient.SetConfig(&pd.ProviderConfig{
		ProviderAddress:    "test-provider",
		SupportedOfferings: []string{"compute"},
		Regions:            []string{"us-east"},
		Active:             true,
		Pricing: pd.PricingConfig{
			CPUPricePerCore:   "1.0",
			MemoryPricePerGB:  "1.0",
			StoragePricePerGB: "1.0",
			MinBidPrice:       "10",
			Currency:          "uve",
		},
		Capacity: pd.CapacityConfig{
			TotalCPUCores:  10, // Limited capacity
			TotalMemoryGB:  32,
			TotalStorageGB: 100,
		},
	})

	km := createTestKeyManager(s.T())
	config := pd.DefaultBidEngineConfig()
	config.ProviderAddress = "test-provider"
	config.OrderPollInterval = time.Millisecond * 50

	be := pd.NewBidEngine(config, km, mockClient)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := be.Start(ctx)
	s.Require().NoError(err)

	// Add orders that should NOT match
	mockClient.AddOrder(pd.Order{
		OrderID:      "wrong-region",
		OfferingType: "compute",
		Region:       "ap-south", // Wrong region
		Requirements: pd.ResourceRequirements{CPUCores: 2},
	})
	mockClient.AddOrder(pd.Order{
		OrderID:      "wrong-offering",
		OfferingType: "ml", // Unsupported offering
		Region:       "us-east",
		Requirements: pd.ResourceRequirements{CPUCores: 2},
	})
	mockClient.AddOrder(pd.Order{
		OrderID:      "exceeds-capacity",
		OfferingType: "compute",
		Region:       "us-east",
		Requirements: pd.ResourceRequirements{CPUCores: 100}, // Exceeds capacity
	})

	// Wait for processing
	time.Sleep(200 * time.Millisecond)
	be.Stop()

	// Verify NO bids were placed (all orders should be rejected)
	bids := mockClient.GetBids()
	s.Equal(0, len(bids), "Expected no bids for non-matching orders")
}

// TestUsageMeteringWorkflow tests the usage metering workflow
func (s *providerDaemonE2ETestSuite) TestUsageMeteringWorkflow() {
	collector := newMockMetricsCollectorE2E()
	recorder := newMockChainRecorderE2E()

	// Set up expected metrics
	collector.SetMetrics("workload-1", &pd.ResourceMetrics{
		CPUMilliSeconds:    5000,
		MemoryByteSeconds:  2 * 1024 * 1024,
		StorageByteSeconds: 1024 * 1024,
		NetworkBytesIn:     10000,
		NetworkBytesOut:    5000,
		GPUSeconds:         0,
	})

	km := createTestKeyManager(s.T())
	recordChan := make(chan *pd.UsageRecord, 10)

	meter := pd.NewUsageMeter(pd.UsageMeterConfig{
		ProviderID:       "provider-e2e-test",
		Interval:         pd.MeteringIntervalMinute,
		MetricsCollector: collector,
		ChainRecorder:    recorder,
		KeyManager:       km,
		RecordChan:       recordChan,
	})

	ctx := context.Background()

	// Step 1: Start metering for a workload (simulates provisioning complete)
	err := meter.StartMetering(
		"workload-1",
		"deployment-1",
		"lease-1",
		pd.PricingInputs{
			AgreedCPURate:    "0.001",
			AgreedMemoryRate: "0.0001",
		},
	)
	s.Require().NoError(err)

	// Step 2: Verify metering state
	state, err := meter.GetMeteringState("workload-1")
	s.Require().NoError(err)
	s.Equal("workload-1", state.WorkloadID)
	s.Equal("deployment-1", state.DeploymentID)
	s.Equal("lease-1", state.LeaseID)
	s.True(state.Active)

	// Step 3: Force collect metrics (simulates periodic collection)
	record, err := meter.ForceCollect(ctx, "workload-1")
	s.Require().NoError(err)
	s.Require().NotNil(record)

	// Verify record properties
	s.Equal(pd.UsageRecordTypePeriodic, record.Type)
	s.Equal("workload-1", record.WorkloadID)
	s.Equal("provider-e2e-test", record.ProviderID)
	s.Equal(int64(5000), record.Metrics.CPUMilliSeconds)
	s.NotEmpty(record.Signature, "Record should be signed")

	// Step 4: Verify record was submitted to chain
	records := recorder.GetRecords()
	s.GreaterOrEqual(len(records), 1)

	// Step 5: Stop metering (simulates workload termination)
	finalRecord, err := meter.StopMetering(ctx, "workload-1")
	s.Require().NoError(err)
	s.Require().NotNil(finalRecord)
	s.Equal(pd.UsageRecordTypeFinal, finalRecord.Type)

	// Verify final settlement was submitted
	finalRecords := recorder.GetFinalRecords()
	s.GreaterOrEqual(len(finalRecords), 1)

	// Step 6: Verify workload is no longer being metered
	_, err = meter.GetMeteringState("workload-1")
	s.Error(err)
	s.Equal(pd.ErrWorkloadNotMetered, err)
}

// TestUsageMeteringCumulativeMetrics tests cumulative metric accumulation
func (s *providerDaemonE2ETestSuite) TestUsageMeteringCumulativeMetrics() {
	collector := newMockMetricsCollectorE2E()
	collector.SetMetrics("workload-cumulative", &pd.ResourceMetrics{
		CPUMilliSeconds:   1000,
		MemoryByteSeconds: 1024,
	})

	meter := pd.NewUsageMeter(pd.UsageMeterConfig{
		ProviderID:       "provider-cumulative",
		MetricsCollector: collector,
	})

	ctx := context.Background()
	err := meter.StartMetering("workload-cumulative", "deploy-1", "lease-1", pd.PricingInputs{})
	s.Require().NoError(err)

	// Collect multiple times
	for i := 0; i < 5; i++ {
		_, err := meter.ForceCollect(ctx, "workload-cumulative")
		s.Require().NoError(err)
	}

	// Verify cumulative metrics
	state, err := meter.GetMeteringState("workload-cumulative")
	s.Require().NoError(err)
	s.Equal(int64(5000), state.CumulativeMetrics.CPUMilliSeconds)
	s.Equal(int64(5120), state.CumulativeMetrics.MemoryByteSeconds)
}

// TestFraudDetection tests the fraud detection logic
func (s *providerDaemonE2ETestSuite) TestFraudDetection() {
	checker := pd.NewFraudChecker()

	s.Run("ValidRecord", func() {
		record := &pd.UsageRecord{
			StartTime: time.Now().Add(-time.Hour),
			EndTime:   time.Now(),
			Metrics: pd.ResourceMetrics{
				CPUMilliSeconds:    3600000,
				MemoryByteSeconds:  1024 * 1024 * 3600,
				StorageByteSeconds: 1024 * 1024 * 3600,
			},
		}
		allocated := &pd.ResourceMetrics{
			CPUMilliSeconds:   1000000,
			MemoryByteSeconds: 1024 * 1024,
		}

		result := checker.CheckRecord(record, allocated)
		s.True(result.Valid)
		s.Equal(0, result.Score)
	})

	s.Run("FutureTimestamp", func() {
		record := &pd.UsageRecord{
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Hour), // Future
			Metrics:   pd.ResourceMetrics{},
		}

		result := checker.CheckRecord(record, nil)
		s.False(result.Valid)
		s.Contains(result.Flags, "FUTURE_TIMESTAMP")
	})

	s.Run("NegativeMetrics", func() {
		record := &pd.UsageRecord{
			StartTime: time.Now().Add(-time.Hour),
			EndTime:   time.Now(),
			Metrics: pd.ResourceMetrics{
				CPUMilliSeconds: -1000,
			},
		}

		result := checker.CheckRecord(record, nil)
		s.False(result.Valid)
		s.Contains(result.Flags, "NEGATIVE_METRICS")
	})

	s.Run("DurationTooLong", func() {
		record := &pd.UsageRecord{
			StartTime: time.Now().Add(-48 * time.Hour),
			EndTime:   time.Now(),
			Metrics:   pd.ResourceMetrics{},
		}

		result := checker.CheckRecord(record, nil)
		s.False(result.Valid)
		s.Contains(result.Flags, "DURATION_TOO_LONG")
	})
}

// Helper functions

func createTestKeyManager(t *testing.T) *pd.KeyManager {
	t.Helper()
	config := pd.DefaultKeyManagerConfig()
	config.StorageType = pd.KeyStorageTypeMemory
	km, err := pd.NewKeyManager(config)
	if err != nil {
		t.Fatalf("Failed to create key manager: %v", err)
	}

	if err := km.Unlock(""); err != nil {
		t.Fatalf("Failed to unlock key manager: %v", err)
	}

	if _, err := km.GenerateKey("test-provider"); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	return km
}

// mockChainClientE2E is a mock implementation for E2E testing
type mockChainClientE2E struct {
	config      *pd.ProviderConfig
	orders      []pd.Order
	bids        []pd.Bid
	placeBidErr error
}

func newMockChainClientForE2E() *mockChainClientE2E {
	return &mockChainClientE2E{
		orders: make([]pd.Order, 0),
		bids:   make([]pd.Bid, 0),
	}
}

func (m *mockChainClientE2E) GetProviderConfig(ctx context.Context, address string) (*pd.ProviderConfig, error) {
	if m.config == nil {
		return nil, fmt.Errorf("config not found")
	}
	return m.config, nil
}

func (m *mockChainClientE2E) GetOpenOrders(ctx context.Context, offeringTypes []string, regions []string) ([]pd.Order, error) {
	return m.orders, nil
}

func (m *mockChainClientE2E) PlaceBid(ctx context.Context, bid *pd.Bid, signature *pd.Signature) error {
	if m.placeBidErr != nil {
		return m.placeBidErr
	}
	m.bids = append(m.bids, *bid)
	return nil
}

func (m *mockChainClientE2E) GetProviderBids(ctx context.Context, address string) ([]pd.Bid, error) {
	return m.bids, nil
}

func (m *mockChainClientE2E) SetConfig(config *pd.ProviderConfig) {
	m.config = config
}

func (m *mockChainClientE2E) AddOrder(order pd.Order) {
	m.orders = append(m.orders, order)
}

func (m *mockChainClientE2E) GetBids() []pd.Bid {
	return m.bids
}

// mockMetricsCollectorE2E is a mock metrics collector for E2E testing
type mockMetricsCollectorE2E struct {
	metrics map[string]*pd.ResourceMetrics
}

func newMockMetricsCollectorE2E() *mockMetricsCollectorE2E {
	return &mockMetricsCollectorE2E{
		metrics: make(map[string]*pd.ResourceMetrics),
	}
}

func (m *mockMetricsCollectorE2E) CollectMetrics(ctx context.Context, workloadID string) (*pd.ResourceMetrics, error) {
	if metrics, ok := m.metrics[workloadID]; ok {
		return metrics, nil
	}
	return &pd.ResourceMetrics{
		CPUMilliSeconds:    1000,
		MemoryByteSeconds:  1024 * 1024,
		StorageByteSeconds: 1024 * 1024,
	}, nil
}

func (m *mockMetricsCollectorE2E) SetMetrics(workloadID string, metrics *pd.ResourceMetrics) {
	m.metrics[workloadID] = metrics
}

// mockChainRecorderE2E is a mock chain recorder for E2E testing
type mockChainRecorderE2E struct {
	records      []*pd.UsageRecord
	finalRecords []*pd.UsageRecord
}

func newMockChainRecorderE2E() *mockChainRecorderE2E {
	return &mockChainRecorderE2E{
		records:      make([]*pd.UsageRecord, 0),
		finalRecords: make([]*pd.UsageRecord, 0),
	}
}

func (m *mockChainRecorderE2E) SubmitUsageRecord(ctx context.Context, record *pd.UsageRecord) error {
	m.records = append(m.records, record)
	return nil
}

func (m *mockChainRecorderE2E) SubmitFinalSettlement(ctx context.Context, record *pd.UsageRecord) error {
	m.finalRecords = append(m.finalRecords, record)
	return nil
}

func (m *mockChainRecorderE2E) GetRecords() []*pd.UsageRecord {
	return m.records
}

func (m *mockChainRecorderE2E) GetFinalRecords() []*pd.UsageRecord {
	return m.finalRecords
}
