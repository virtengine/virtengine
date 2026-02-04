//go:build e2e.integration

// Package waldur provides E2E tests for Waldur integration.
//
// VE-21D: Waldur marketplace E2E integration tests
// This file tests the complete purchase → provision → control → usage → terminate flow.
package waldur

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/tests/e2e/mocks"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// ============================================================================
// Marketplace Lifecycle E2E Test Suite
// ============================================================================

// MarketplaceLifecycleTestSuite tests the complete E2E flow:
// 1. Offering creation and sync
// 2. Purchase flow (order creation)
// 3. Provisioning flow
// 4. Resource lifecycle (start/stop/resize)
// 5. Usage reporting
// 6. Termination
type MarketplaceLifecycleTestSuite struct {
	suite.Suite

	// Mock Waldur server
	waldurMock *mocks.WaldurMock

	// Waldur clients
	waldurClient      *waldur.Client
	marketplaceClient *waldur.MarketplaceClient
	lifecycleClient   *waldur.LifecycleClient
	usageClient       *waldur.UsageClient

	// Test data
	testProjectUUID  string
	testCustomerUUID string
	testProviderAddr string
	testOfferingUUID string
	testOrderUUID    string
	testResourceUUID string
	testAllocationID string

	// Provider daemon components
	usageReporter *provider_daemon.WaldurUsageReporter
}

func TestMarketplaceLifecycle(t *testing.T) {
	suite.Run(t, new(MarketplaceLifecycleTestSuite))
}

func (s *MarketplaceLifecycleTestSuite) SetupSuite() {
	// Create mock Waldur server
	config := mocks.DefaultWaldurMockConfig()
	config.AutoApproveOrders = false
	config.AutoProvision = false
	s.waldurMock = mocks.NewWaldurMockWithConfig(config)

	s.testProjectUUID = config.ProjectUUID
	s.testCustomerUUID = config.CustomerUUID
	s.testProviderAddr = "ve1provider" + uuid.New().String()[:8]
	s.testAllocationID = "alloc-" + uuid.New().String()[:8]

	// Create Waldur clients
	waldurCfg := waldur.DefaultConfig()
	waldurCfg.BaseURL = s.waldurMock.BaseURL()
	waldurCfg.Token = "test-token"

	var err error
	s.waldurClient, err = waldur.NewClient(waldurCfg)
	s.Require().NoError(err)

	s.marketplaceClient = waldur.NewMarketplaceClient(s.waldurClient)
	s.lifecycleClient = waldur.NewLifecycleClient(s.marketplaceClient)
	s.usageClient = waldur.NewUsageClient(s.marketplaceClient)

	// Create usage reporter
	reporterCfg := provider_daemon.DefaultWaldurUsageReporterConfig()
	reporterCfg.ProviderAddress = s.testProviderAddr
	reporterCfg.StateFilePath = "" // In-memory only for tests

	s.usageReporter, err = provider_daemon.NewWaldurUsageReporter(
		reporterCfg,
		s.marketplaceClient,
		provider_daemon.NewUsageSnapshotStore(),
		nil, // No bridge state needed for tests
		nil, // No audit logger for tests
	)
	s.Require().NoError(err)
}

func (s *MarketplaceLifecycleTestSuite) TearDownSuite() {
	if s.usageReporter != nil {
		_ = s.usageReporter.Stop()
	}
	if s.waldurMock != nil {
		s.waldurMock.Close()
	}
}

// ============================================================================
// Helper Methods
// ============================================================================

func (s *MarketplaceLifecycleTestSuite) httpGet(path string) ([]byte, int, error) {
	resp, err := http.Get(s.waldurMock.BaseURL() + path)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

func (s *MarketplaceLifecycleTestSuite) httpPost(path string, payload interface{}) ([]byte, int, error) {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	resp, err := http.Post(s.waldurMock.BaseURL()+path, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

// ============================================================================
// Phase 1: Offering Creation and Sync
// ============================================================================

func (s *MarketplaceLifecycleTestSuite) TestPhase1_OfferingSync() {
	s.Run("CreateOfferingInWaldur", func() {
		ctx := context.Background()

		// Create offering via API
		offering, err := s.marketplaceClient.CreateOffering(ctx, waldur.CreateOfferingRequest{
			Name:         "VirtEngine Compute - Test",
			Description:  "Test compute offering for E2E",
			Type:         "VirtEngine.Compute",
			CustomerUUID: s.testCustomerUUID,
			Shared:       true,
			Billable:     true,
			BackendID:    s.testProviderAddr + "/compute-001",
			Components: []waldur.PricingComponent{
				{
					Type:         "usage",
					Name:         "cpu_hours",
					MeasuredUnit: "hours",
					BillingType:  "usage",
					Price:        "0.10",
				},
				{
					Type:         "usage",
					Name:         "ram_gb_hours",
					MeasuredUnit: "GB-hours",
					BillingType:  "usage",
					Price:        "0.05",
				},
			},
		})
		s.Require().NoError(err)
		s.NotEmpty(offering.UUID)
		s.testOfferingUUID = offering.UUID
	})

	s.Run("VerifyOfferingExists", func() {
		ctx := context.Background()

		offering, err := s.marketplaceClient.GetOffering(ctx, s.testOfferingUUID)
		s.Require().NoError(err)
		s.Equal("VirtEngine Compute - Test", offering.Name)
	})

	s.Run("FindOfferingByBackendID", func() {
		ctx := context.Background()

		backendID := s.testProviderAddr + "/compute-001"
		offering, err := s.marketplaceClient.GetOfferingByBackendID(ctx, backendID)
		s.Require().NoError(err)
		s.Equal(s.testOfferingUUID, offering.UUID)
	})
}

// ============================================================================
// Phase 2: Purchase Flow
// ============================================================================

func (s *MarketplaceLifecycleTestSuite) TestPhase2_PurchaseFlow() {
	// Ensure offering exists
	if s.testOfferingUUID == "" {
		s.T().Skip("Offering not created, skipping purchase tests")
	}

	s.Run("CreateOrder", func() {
		ctx := context.Background()

		order, err := s.marketplaceClient.CreateOrder(ctx, waldur.CreateOrderRequest{
			OfferingUUID:   s.testOfferingUUID,
			ProjectUUID:    s.testProjectUUID,
			Name:           "test-allocation",
			Description:    "E2E test allocation",
			RequestComment: "Automated E2E test",
			Attributes: map[string]interface{}{
				"allocation_id":    s.testAllocationID,
				"customer_address": "ve1customer123",
				"provider_address": s.testProviderAddr,
			},
		})
		s.Require().NoError(err)
		s.NotEmpty(order.UUID)
		s.testOrderUUID = order.UUID
	})

	s.Run("ApproveOrderByProvider", func() {
		ctx := context.Background()

		err := s.marketplaceClient.ApproveOrderByProvider(ctx, s.testOrderUUID)
		s.Require().NoError(err)

		// Verify order state
		order, err := s.marketplaceClient.GetOrder(ctx, s.testOrderUUID)
		s.Require().NoError(err)
		s.Contains([]string{"approved", "executing", "done"}, order.State)
	})

	s.Run("SetOrderBackendID", func() {
		ctx := context.Background()

		err := s.marketplaceClient.SetOrderBackendID(ctx, s.testOrderUUID, s.testAllocationID)
		s.Require().NoError(err)
	})
}

// ============================================================================
// Phase 3: Provisioning Flow
// ============================================================================

func (s *MarketplaceLifecycleTestSuite) TestPhase3_Provisioning() {
	// Skip if order not created
	if s.testOrderUUID == "" {
		s.T().Skip("Order not created, skipping provisioning tests")
	}

	s.Run("ProvisionResource", func() {
		// Simulate resource provisioning by creating resource directly in mock
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "test-allocation-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			OrderUUID:    s.testOrderUUID,
			BackendID:    s.testAllocationID,
			State:        "OK",
			Attributes: map[string]interface{}{
				"allocation_id":    s.testAllocationID,
				"provider_address": s.testProviderAddr,
				"external_ip":      "203.0.113.100",
			},
		})
		s.testResourceUUID = resourceUUID
	})

	s.Run("VerifyResourceState", func() {
		ctx := context.Background()

		resource, err := s.marketplaceClient.GetResource(ctx, s.testResourceUUID)
		s.Require().NoError(err)
		s.Equal("OK", resource.State)
	})

	s.Run("GetResourceState", func() {
		ctx := context.Background()

		state, err := s.lifecycleClient.GetResourceState(ctx, s.testResourceUUID)
		s.Require().NoError(err)
		s.Equal(waldur.ResourceStateOK, state)
	})
}

// ============================================================================
// Phase 4: Resource Lifecycle Control
// ============================================================================

func (s *MarketplaceLifecycleTestSuite) TestPhase4_LifecycleControl() {
	// Skip if resource not created
	if s.testResourceUUID == "" {
		s.T().Skip("Resource not created, skipping lifecycle tests")
	}

	s.Run("StopResource", func() {
		ctx := context.Background()

		req := waldur.LifecycleRequest{
			ResourceUUID:   s.testResourceUUID,
			IdempotencyKey: fmt.Sprintf("stop-%d", time.Now().Unix()),
		}

		resp, err := s.lifecycleClient.Stop(ctx, req)
		s.Require().NoError(err)
		s.Contains([]string{"accepted", "executing", "done"}, resp.State)

		// Update mock resource state
		s.waldurMock.SetResourceState(s.testResourceUUID, "Stopped")
	})

	s.Run("StartResource", func() {
		ctx := context.Background()

		req := waldur.LifecycleRequest{
			ResourceUUID:   s.testResourceUUID,
			IdempotencyKey: fmt.Sprintf("start-%d", time.Now().Unix()),
		}

		resp, err := s.lifecycleClient.Start(ctx, req)
		s.Require().NoError(err)
		s.Contains([]string{"accepted", "executing", "done"}, resp.State)

		// Update mock resource state
		s.waldurMock.SetResourceState(s.testResourceUUID, "OK")
	})

	s.Run("RestartResource", func() {
		ctx := context.Background()

		req := waldur.LifecycleRequest{
			ResourceUUID:   s.testResourceUUID,
			IdempotencyKey: fmt.Sprintf("restart-%d", time.Now().Unix()),
		}

		resp, err := s.lifecycleClient.Restart(ctx, req)
		s.Require().NoError(err)
		s.Contains([]string{"accepted", "executing", "done"}, resp.State)
	})

	s.Run("ResizeResource", func() {
		ctx := context.Background()

		cpuCores := 4
		memoryMB := 8192
		req := waldur.ResizeRequest{
			LifecycleRequest: waldur.LifecycleRequest{
				ResourceUUID:   s.testResourceUUID,
				IdempotencyKey: fmt.Sprintf("resize-%d", time.Now().Unix()),
			},
			CPUCores: &cpuCores,
			MemoryMB: &memoryMB,
		}

		resp, err := s.lifecycleClient.Resize(ctx, req)
		s.Require().NoError(err)
		s.Contains([]string{"accepted", "executing", "done"}, resp.State)
	})

	s.Run("ValidateLifecycleTransitions", func() {
		// Valid: OK -> Stop
		err := waldur.ValidateLifecycleAction(waldur.ResourceStateOK, waldur.LifecycleActionStop)
		s.NoError(err)

		// Valid: Stopped -> Start
		err = waldur.ValidateLifecycleAction(waldur.ResourceStateStopped, waldur.LifecycleActionStart)
		s.NoError(err)

		// Invalid: Terminated -> Start
		err = waldur.ValidateLifecycleAction(waldur.ResourceStateTerminated, waldur.LifecycleActionStart)
		s.Error(err)
	})
}

// ============================================================================
// Phase 5: Usage Reporting
// ============================================================================

func (s *MarketplaceLifecycleTestSuite) TestPhase5_UsageReporting() {
	// Skip if resource not created
	if s.testResourceUUID == "" {
		s.T().Skip("Resource not created, skipping usage tests")
	}

	s.Run("SubmitComponentUsage", func() {
		ctx := context.Background()

		periodStart := time.Now().Add(-1 * time.Hour)
		periodEnd := time.Now()

		resp, err := s.usageClient.SubmitComponentUsage(
			ctx,
			s.testResourceUUID,
			periodStart,
			periodEnd,
			2.5,  // cpuHours
			0.5,  // gpuHours
			8.0,  // ramGBHours
			50.0, // storageGBHours
			1.2,  // networkGB
		)
		s.Require().NoError(err)
		s.Equal("submitted", resp.State)
	})

	s.Run("SubmitDetailedUsageReport", func() {
		ctx := context.Background()

		periodStart := time.Now().Add(-2 * time.Hour)
		periodEnd := time.Now().Add(-1 * time.Hour)

		report := &waldur.ResourceUsageReport{
			ResourceUUID: s.testResourceUUID,
			PeriodStart:  periodStart,
			PeriodEnd:    periodEnd,
			Components: []waldur.ComponentUsage{
				{Type: "cpu_hours", Amount: 1.5, Description: "CPU usage"},
				{Type: "ram_gb_hours", Amount: 4.0, Description: "Memory usage"},
				{Type: "storage_gb_hours", Amount: 25.0, Description: "Storage usage"},
			},
			BackendID: s.testAllocationID,
			Metadata: map[string]string{
				"provider": s.testProviderAddr,
			},
		}

		resp, err := s.usageClient.SubmitUsageReport(ctx, report)
		s.Require().NoError(err)
		s.Equal("submitted", resp.State)
	})

	s.Run("QueueUsageViaReporter", func() {
		periodStart := time.Now().Add(-3 * time.Hour)
		periodEnd := time.Now().Add(-2 * time.Hour)

		metrics := &provider_daemon.ResourceMetrics{
			CPUMilliSeconds:    3600000,            // 1 hour
			MemoryByteSeconds:  1073741824 * 3600,  // 1GB for 1 hour
			StorageByteSeconds: 10737418240 * 3600, // 10GB for 1 hour
		}

		err := s.usageReporter.QueueUsageReport(
			s.testAllocationID,
			s.testResourceUUID,
			periodStart,
			periodEnd,
			metrics,
		)
		s.Require().NoError(err)
		s.Equal(1, s.usageReporter.GetPendingCount())
	})
}

// ============================================================================
// Phase 6: Termination
// ============================================================================

func (s *MarketplaceLifecycleTestSuite) TestPhase6_Termination() {
	// Skip if resource not created
	if s.testResourceUUID == "" {
		s.T().Skip("Resource not created, skipping termination tests")
	}

	s.Run("TerminateResource", func() {
		ctx := context.Background()

		req := waldur.LifecycleRequest{
			ResourceUUID:   s.testResourceUUID,
			IdempotencyKey: fmt.Sprintf("terminate-%d", time.Now().Unix()),
			Immediate:      true,
		}

		resp, err := s.lifecycleClient.Terminate(ctx, req)
		s.Require().NoError(err)
		s.Contains([]string{"accepted", "executing", "done"}, resp.State)

		// Update mock resource state
		s.waldurMock.SetResourceState(s.testResourceUUID, "Terminated")
	})

	s.Run("VerifyResourceTerminated", func() {
		ctx := context.Background()

		state, err := s.lifecycleClient.GetResourceState(ctx, s.testResourceUUID)
		s.Require().NoError(err)
		s.Equal(waldur.ResourceStateTerminated, state)
	})

	s.Run("CannotStartTerminatedResource", func() {
		// Validate that terminated resources cannot be started
		err := waldur.ValidateLifecycleAction(waldur.ResourceStateTerminated, waldur.LifecycleActionStart)
		s.Error(err)
	})
}

// ============================================================================
// Integration with Marketplace Types
// ============================================================================

func (s *MarketplaceLifecycleTestSuite) TestMarketplaceTypeIntegration() {
	s.Run("OfferingSyncChecksum", func() {
		offering := &marketplace.Offering{
			Name:        "Test Offering",
			Description: "Test description",
			Category:    marketplace.OfferingCategoryCompute,
			State:       marketplace.OfferingStateActive,
		}

		checksum1 := offering.SyncChecksum()
		s.NotEmpty(checksum1)

		// Same offering should have same checksum
		checksum2 := offering.SyncChecksum()
		s.Equal(checksum1, checksum2)

		// Changed offering should have different checksum
		offering.Name = "Changed Name"
		checksum3 := offering.SyncChecksum()
		s.NotEqual(checksum1, checksum3)
	})

	s.Run("LifecycleTransitionValidation", func() {
		// Valid: Active -> Suspended (stop)
		newState, err := marketplace.ValidateLifecycleTransition(
			marketplace.AllocationStateActive,
			marketplace.LifecycleActionStop,
		)
		s.NoError(err)
		s.Equal(marketplace.AllocationStateSuspended, newState)

		// Valid: Suspended -> Active (start)
		newState, err = marketplace.ValidateLifecycleTransition(
			marketplace.AllocationStateSuspended,
			marketplace.LifecycleActionStart,
		)
		s.NoError(err)
		s.Equal(marketplace.AllocationStateActive, newState)

		// Invalid: Terminated -> Start
		_, err = marketplace.ValidateLifecycleTransition(
			marketplace.AllocationStateTerminated,
			marketplace.LifecycleActionStart,
		)
		s.Error(err)
	})

	s.Run("LifecycleOperationCreation", func() {
		op, err := marketplace.NewLifecycleOperation(
			s.testAllocationID,
			marketplace.LifecycleActionStart,
			"ve1requestor123",
			s.testProviderAddr,
			marketplace.AllocationStateSuspended,
		)
		s.Require().NoError(err)
		s.NotEmpty(op.ID)
		s.NotEmpty(op.IdempotencyKey)
		s.Equal(marketplace.LifecycleOpStatePending, op.State)
		s.Equal(marketplace.AllocationStateActive, op.TargetAllocationState)
	})

	s.Run("IdempotencyKeyGeneration", func() {
		now := time.Date(2025, 1, 1, 10, 15, 0, 0, time.UTC)

		key1 := marketplace.GenerateIdempotencyKey(
			s.testAllocationID,
			marketplace.LifecycleActionStart,
			now,
		)
		s.NotEmpty(key1)

		// Same params within same hour should give same key
		key2 := marketplace.GenerateIdempotencyKey(
			s.testAllocationID,
			marketplace.LifecycleActionStart,
			now.Add(30*time.Minute),
		)
		s.Equal(key1, key2)

		// Different action should give different key
		key3 := marketplace.GenerateIdempotencyKey(
			s.testAllocationID,
			marketplace.LifecycleActionStop,
			now,
		)
		s.NotEqual(key1, key3)
	})
}

// ============================================================================
// Error Scenarios
// ============================================================================

func (s *MarketplaceLifecycleTestSuite) TestErrorScenarios() {
	s.Run("OrderForNonExistentOffering", func() {
		ctx := context.Background()

		_, err := s.marketplaceClient.CreateOrder(ctx, waldur.CreateOrderRequest{
			OfferingUUID: "00000000-0000-0000-0000-000000000099",
			ProjectUUID:  s.testProjectUUID,
			Name:         "invalid-order",
		})
		s.Error(err)
	})

	s.Run("LifecycleActionOnNonExistentResource", func() {
		ctx := context.Background()

		req := waldur.LifecycleRequest{
			ResourceUUID: "00000000-0000-0000-0000-000000000099",
		}

		_, err := s.lifecycleClient.Start(ctx, req)
		s.Error(err)
	})

	s.Run("UsageReportForNonExistentResource", func() {
		ctx := context.Background()

		report := &waldur.ResourceUsageReport{
			ResourceUUID: "00000000-0000-0000-0000-000000000099",
			PeriodStart:  time.Now().Add(-1 * time.Hour),
			PeriodEnd:    time.Now(),
			Components: []waldur.ComponentUsage{
				{Type: "cpu_hours", Amount: 1.0},
			},
		}

		_, err := s.usageClient.SubmitUsageReport(ctx, report)
		s.Error(err)
	})

	s.Run("EmptyUsageReport", func() {
		ctx := context.Background()

		_, err := s.usageClient.SubmitUsageReport(ctx, nil)
		s.Error(err)

		_, err = s.usageClient.SubmitUsageReport(ctx, &waldur.ResourceUsageReport{
			ResourceUUID: s.testResourceUUID,
			PeriodStart:  time.Now(),
			PeriodEnd:    time.Now(),
			Components:   []waldur.ComponentUsage{}, // Empty
		})
		s.Error(err)
	})
}
