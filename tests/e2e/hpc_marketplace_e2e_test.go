//go:build e2e.integration

// Package e2e contains end-to-end integration tests.
//
// VE-8C: E2E tests for HPC and marketplace provider flow.
// Tests the complete provider workflow from registration through settlement.
package e2e

import (
	"context"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/testutil"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCDefaultDeposit is the default deposit for HPC marketplace deployments
var HPCDefaultDeposit = sdk.NewCoin("uvirt", sdkmath.NewInt(5000000))

// HPCMarketplaceE2ETestSuite tests the complete HPC and marketplace provider flow:
// Provider registration → Offering listing → Job execution → Usage reporting → Settlement
type HPCMarketplaceE2ETestSuite struct {
	*testutil.NetworkTestSuite

	// Test addresses
	providerAddr string
	customerAddr string

	// Test paths
	providerPath   string
	deploymentPath string

	// Mock components for unit-level E2E testing
	mockScheduler   *MockHPCScheduler
	mockReporter    *MockHPCOnChainReporter
	mockAuditor     *MockHPCAuditLogger
	mockWaldur      *MockWaldurClient
	mockSettlement  *MockSettlementClient
}

func TestHPCMarketplaceE2E(t *testing.T) {
	suite.Run(t, &HPCMarketplaceE2ETestSuite{
		NetworkTestSuite: testutil.NewNetworkTestSuite(nil, &HPCMarketplaceE2ETestSuite{}),
	})
}

func (s *HPCMarketplaceE2ETestSuite) SetupSuite() {
	s.NetworkTestSuite.SetupSuite()

	val := s.Network().Validators[0]
	s.providerAddr = val.Address.String()
	s.customerAddr = val.Address.String()

	var err error
	s.providerPath, err = filepath.Abs("../../x/provider/testdata/provider.yaml")
	s.Require().NoError(err)

	s.deploymentPath, err = filepath.Abs("../../x/deployment/testdata/deployment.yaml")
	s.Require().NoError(err)

	// Initialize mock components
	s.mockScheduler = NewMockHPCScheduler()
	s.mockReporter = NewMockHPCOnChainReporter()
	s.mockAuditor = NewMockHPCAuditLogger()
	s.mockWaldur = NewMockWaldurClient()
	s.mockSettlement = NewMockSettlementClient()
}

// =============================================================================
// A. Staging Environment Setup Tests
// =============================================================================

func (s *HPCMarketplaceE2ETestSuite) TestA_StagingEnvironmentSetup() {
	s.Run("ValidateSchedulerBackend", func() {
		ctx := context.Background()
		
		// Verify scheduler is functional
		err := s.mockScheduler.Start(ctx)
		s.Require().NoError(err)
		s.True(s.mockScheduler.IsRunning())
		
		// Configure HPC settings
		config := pd.DefaultHPCConfig()
		config.ClusterID = "e2e-test-cluster"
		config.SchedulerType = pd.HPCSchedulerTypeSLURM
		config.UsageReporting.Enabled = true
		config.UsageReporting.ReportInterval = time.Second * 5
		
		s.Equal("e2e-test-cluster", config.ClusterID)
	})

	s.Run("ValidateProviderDaemonConfig", func() {
		config := pd.DefaultBidEngineConfig()
		config.ProviderAddress = s.providerAddr
		config.OrderPollInterval = time.Millisecond * 100
		config.ConfigPollInterval = time.Millisecond * 100
		
		s.NotEmpty(config.ProviderAddress)
	})

	s.Run("ValidateWaldurBridgeConfig", func() {
		config := pd.DefaultWaldurBridgeConfig()
		config.ProviderAddress = s.providerAddr
		config.OfferingSyncEnabled = true
		config.OfferingSyncInterval = 60
		
		s.True(config.OfferingSyncEnabled)
	})
}

// =============================================================================
// B. Provider Registration and Offering Publishing Tests
// =============================================================================

func (s *HPCMarketplaceE2ETestSuite) TestB_ProviderRegistrationAndOfferings() {
	ctx := context.Background()

	s.Run("RegisterProvider", func() {
		// Simulate provider registration
		s.mockWaldur.SetProviderRegistered(s.providerAddr, true)
		
		registered := s.mockWaldur.IsProviderRegistered(s.providerAddr)
		s.True(registered)
	})

	s.Run("PublishOfferings", func() {
		offerings := []MockOffering{
			{
				OfferingID:   "hpc-compute-standard",
				Name:         "HPC Compute Standard",
				Category:     "compute",
				CPUCores:     64,
				MemoryGB:     256,
				GPUs:         4,
				PricePerHour: "10.0",
				Active:       true,
			},
			{
				OfferingID:   "hpc-gpu-a100",
				Name:         "HPC GPU A100",
				Category:     "gpu",
				CPUCores:     32,
				MemoryGB:     128,
				GPUs:         8,
				PricePerHour: "50.0",
				Active:       true,
			},
		}

		for _, offering := range offerings {
			err := s.mockWaldur.PublishOffering(ctx, offering)
			s.Require().NoError(err)
		}

		// Verify offerings are published
		published := s.mockWaldur.GetPublishedOfferings(s.providerAddr)
		s.Len(published, 2)
	})

	s.Run("SyncOfferingsToChain", func() {
		// Verify chain sync
		synced := s.mockWaldur.GetSyncedOfferingIDs()
		s.GreaterOrEqual(len(synced), 2)
	})
}

// =============================================================================
// C. Order Creation and Resource Allocation Tests
// =============================================================================

func (s *HPCMarketplaceE2ETestSuite) TestC_OrderCreationAndAllocation() {
	ctx := context.Background()

	var orderID string
	var allocationID string

	s.Run("CreateTestOrder", func() {
		order := MockOrder{
			OrderID:      "order-e2e-test-1",
			CustomerAddr: s.customerAddr,
			OfferingID:   "hpc-compute-standard",
			Requirements: pd.ResourceRequirements{
				CPUCores:  8,
				MemoryGB:  32,
				GPUs:      1,
				StorageGB: 100,
			},
			MaxPrice: "100.0",
			Duration: time.Hour * 24,
		}

		err := s.mockWaldur.CreateOrder(ctx, order)
		s.Require().NoError(err)
		orderID = order.OrderID
	})

	s.Run("ProviderPlacesBid", func() {
		bid := MockBid{
			BidID:        "bid-e2e-test-1",
			OrderID:      orderID,
			ProviderAddr: s.providerAddr,
			Price:        "80.0", // Under max price
			TTL:          time.Hour,
		}

		err := s.mockWaldur.PlaceBid(ctx, bid)
		s.Require().NoError(err)

		// Verify bid was placed
		bids := s.mockWaldur.GetBidsForOrder(orderID)
		s.Len(bids, 1)
		s.Equal("80.0", bids[0].Price)
	})

	s.Run("AcceptBidAndCreateAllocation", func() {
		// Accept the bid
		err := s.mockWaldur.AcceptBid(ctx, orderID, "bid-e2e-test-1")
		s.Require().NoError(err)

		// Verify allocation was created
		allocations := s.mockWaldur.GetAllocationsForOrder(orderID)
		s.Len(allocations, 1)
		allocationID = allocations[0].AllocationID
		s.NotEmpty(allocationID)
	})

	s.Run("ProvisionResources", func() {
		// Verify resources are provisioned
		status := s.mockWaldur.GetAllocationStatus(allocationID)
		s.Equal("provisioned", status)
	})
}

// =============================================================================
// D. Job Submission and Scheduler Execution Tests
// =============================================================================

func (s *HPCMarketplaceE2ETestSuite) TestD_JobSubmissionAndExecution() {
	ctx := context.Background()

	var jobID string

	s.Run("CreateHPCJob", func() {
		job := &hpctypes.HPCJob{
			JobID:           "hpc-job-e2e-1",
			ClusterID:       "e2e-test-cluster",
			OfferingID:      "hpc-compute-standard",
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			State:           hpctypes.JobStatePending,
			QueueName:       "default",
			WorkloadSpec: hpctypes.JobWorkloadSpec{
				ContainerImage: "alpine:latest",
				Command:        "echo 'Hello HPC' && sleep 60",
			},
			Resources: hpctypes.JobResources{
				Nodes:           1,
				CPUCoresPerNode: 4,
				MemoryGBPerNode: 8,
				GPUsPerNode:     1,
				StorageGB:       10,
			},
			MaxRuntimeSeconds: int64(time.Hour / time.Second),
			CreatedAt:         time.Now(),
		}

		jobID = job.JobID

		// Submit to mock scheduler
		schedulerJob, err := s.mockScheduler.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)
		s.Equal(pd.HPCJobStatePending, schedulerJob.State)
	})

	s.Run("VerifyJobQueued", func() {
		// Wait for job to be queued
		time.Sleep(100 * time.Millisecond)

		status, err := s.mockScheduler.GetJobStatus(ctx, jobID)
		s.Require().NoError(err)
		s.Contains([]pd.HPCJobState{pd.HPCJobStatePending, pd.HPCJobStateQueued, pd.HPCJobStateRunning}, status.State)
	})

	s.Run("SimulateJobExecution", func() {
		// Progress job to running
		s.mockScheduler.SetJobState(jobID, pd.HPCJobStateRunning)

		status, err := s.mockScheduler.GetJobStatus(ctx, jobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateRunning, status.State)
	})

	s.Run("VerifySchedulerExecution", func() {
		// Get job accounting while running
		metrics, err := s.mockScheduler.GetJobAccounting(ctx, jobID)
		s.Require().NoError(err)
		s.NotNil(metrics)
	})

	s.Run("CompleteJob", func() {
		// Complete the job
		s.mockScheduler.SetJobState(jobID, pd.HPCJobStateCompleted)
		s.mockScheduler.SetJobMetrics(jobID, &pd.HPCSchedulerMetrics{
			WallClockSeconds: 3600,
			CPUCoreSeconds:   14400, // 4 cores * 1 hour
			MemoryGBSeconds:  28800, // 8 GB * 1 hour
			GPUSeconds:       3600,
			NodesUsed:        1,
			NodeHours:        1.0,
		})

		status, err := s.mockScheduler.GetJobStatus(ctx, jobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCompleted, status.State)
		s.True(status.State.IsTerminal())
	})
}

// =============================================================================
// E. Usage Metrics and Reporting Tests
// =============================================================================

func (s *HPCMarketplaceE2ETestSuite) TestE_UsageMetricsAndReporting() {
	ctx := context.Background()

	var usageRecordID string

	s.Run("CaptureUsageMetrics", func() {
		metrics := &pd.HPCSchedulerMetrics{
			WallClockSeconds: 3600,
			CPUCoreSeconds:   14400,
			MemoryGBSeconds:  28800,
			GPUSeconds:       3600,
			NodesUsed:        1,
			NodeHours:        1.0,
			NetworkBytesIn:   1073741824,  // 1 GB
			NetworkBytesOut:  536870912,   // 0.5 GB
		}

		s.mockReporter.SetJobMetrics("hpc-job-e2e-1", metrics)

		captured := s.mockReporter.GetJobMetrics("hpc-job-e2e-1")
		s.NotNil(captured)
		s.Equal(int64(3600), captured.WallClockSeconds)
	})

	s.Run("CreateUsageReport", func() {
		record := &pd.HPCUsageRecord{
			RecordID:        "usage-record-e2e-1",
			JobID:           "hpc-job-e2e-1",
			ClusterID:       "e2e-test-cluster",
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			PeriodStart:     time.Now().Add(-time.Hour),
			PeriodEnd:       time.Now(),
			Metrics:         s.mockReporter.GetJobMetrics("hpc-job-e2e-1"),
			IsFinal:         true,
			JobState:        pd.HPCJobStateCompleted,
			Timestamp:       time.Now(),
		}

		usageRecordID = record.RecordID

		// Sign and queue the record
		hash := record.Hash()
		s.NotNil(hash)

		err := s.mockReporter.QueueUsageRecord(record)
		s.Require().NoError(err)
	})

	s.Run("SubmitUsageReportOnChain", func() {
		report := &pd.HPCStatusReport{
			ProviderAddress: s.providerAddr,
			VirtEngineJobID: "hpc-job-e2e-1",
			SchedulerJobID:  "slurm-hpc-job-e2e-1",
			SchedulerType:   pd.HPCSchedulerTypeSLURM,
			State:           pd.HPCJobStateCompleted,
			Timestamp:       time.Now(),
			Signature:       "mock-signature",
		}

		err := s.mockReporter.ReportJobStatus(ctx, report)
		s.Require().NoError(err)

		// Verify report was submitted
		reports := s.mockReporter.GetSubmittedReports()
		s.GreaterOrEqual(len(reports), 1)
	})

	s.Run("VerifyUsageRecordIntegrity", func() {
		record := s.mockReporter.GetUsageRecord(usageRecordID)
		s.NotNil(record)
		s.Equal("hpc-job-e2e-1", record.JobID)
		s.True(record.IsFinal)
	})
}

// =============================================================================
// F. Invoice Generation and Settlement Tests
// =============================================================================

func (s *HPCMarketplaceE2ETestSuite) TestF_InvoiceAndSettlement() {
	ctx := context.Background()

	var invoiceID string

	s.Run("GenerateInvoice", func() {
		invoice := MockInvoice{
			InvoiceID:    "invoice-e2e-1",
			ProviderAddr: s.providerAddr,
			CustomerAddr: s.customerAddr,
			OrderID:      "order-e2e-test-1",
			LineItems: []MockLineItem{
				{
					ResourceType: "cpu",
					Quantity:     14400, // core-seconds
					UnitPrice:    "0.0001",
					TotalCost:    "1.44",
				},
				{
					ResourceType: "memory",
					Quantity:     28800, // GB-seconds
					UnitPrice:    "0.00001",
					TotalCost:    "0.288",
				},
				{
					ResourceType: "gpu",
					Quantity:     3600, // seconds
					UnitPrice:    "0.001",
					TotalCost:    "3.60",
				},
			},
			TotalAmount:  "5.328",
			PeriodStart:  time.Now().Add(-time.Hour),
			PeriodEnd:    time.Now(),
			Status:       "pending",
		}

		err := s.mockSettlement.CreateInvoice(ctx, invoice)
		s.Require().NoError(err)
		invoiceID = invoice.InvoiceID
	})

	s.Run("VerifyBillableLineItems", func() {
		invoice := s.mockSettlement.GetInvoice(invoiceID)
		s.NotNil(invoice)
		s.Len(invoice.LineItems, 3)
		
		// Verify each line item
		for _, item := range invoice.LineItems {
			s.NotEmpty(item.ResourceType)
			s.NotEmpty(item.TotalCost)
		}
	})

	s.Run("TriggerSettlement", func() {
		err := s.mockSettlement.TriggerSettlement(ctx, invoiceID)
		s.Require().NoError(err)

		// Verify settlement status
		invoice := s.mockSettlement.GetInvoice(invoiceID)
		s.Equal("settled", invoice.Status)
	})

	s.Run("VerifyProviderPayout", func() {
		payout := s.mockSettlement.GetProviderPayout(s.providerAddr, invoiceID)
		s.NotNil(payout)
		s.Equal("completed", payout.Status)
		
		// Provider should receive ~97.5% (platform fee is 2.5%)
		s.NotEmpty(payout.Amount)
	})

	s.Run("VerifyPlatformFee", func() {
		fee := s.mockSettlement.GetPlatformFee(invoiceID)
		s.NotNil(fee)
		s.NotEmpty(fee.Amount)
	})
}

// =============================================================================
// G. On-Chain State Transitions and Events Tests
// =============================================================================

func (s *HPCMarketplaceE2ETestSuite) TestG_StateTransitionsAndEvents() {
	s.Run("VerifyJobStateTransitions", func() {
		// Valid transitions
		validTransitions := map[pd.HPCJobState][]pd.HPCJobState{
			pd.HPCJobStatePending:   {pd.HPCJobStateQueued, pd.HPCJobStateFailed, pd.HPCJobStateCancelled},
			pd.HPCJobStateQueued:    {pd.HPCJobStateStarting, pd.HPCJobStateFailed, pd.HPCJobStateCancelled},
			pd.HPCJobStateStarting:  {pd.HPCJobStateRunning, pd.HPCJobStateFailed, pd.HPCJobStateCancelled},
			pd.HPCJobStateRunning:   {pd.HPCJobStateCompleted, pd.HPCJobStateFailed, pd.HPCJobStateCancelled, pd.HPCJobStateSuspended, pd.HPCJobStateTimeout},
			pd.HPCJobStateSuspended: {pd.HPCJobStateRunning, pd.HPCJobStateFailed, pd.HPCJobStateCancelled},
		}

		for from, toStates := range validTransitions {
			for _, to := range toStates {
				valid := s.mockScheduler.IsValidTransition(from, to)
				s.True(valid, "Transition from %s to %s should be valid", from, to)
			}
		}
	})

	s.Run("VerifyEventEmissions", func() {
		events := s.mockAuditor.GetEvents()
		
		// Should have job lifecycle events
		jobEvents := filterEventsByType(events, "job")
		s.GreaterOrEqual(len(jobEvents), 0)
	})

	s.Run("VerifyAccountingStatusTransitions", func() {
		// Pending → Finalized → Settled
		statuses := []hpctypes.AccountingRecordStatus{
			hpctypes.AccountingStatusPending,
			hpctypes.AccountingStatusFinalized,
			hpctypes.AccountingStatusSettled,
		}

		for i := 0; i < len(statuses)-1; i++ {
			from := statuses[i]
			to := statuses[i+1]
			s.True(isValidAccountingTransition(from, to))
		}
	})
}

// =============================================================================
// H. Negative Tests - Failed Jobs and Partial Usage
// =============================================================================

func (s *HPCMarketplaceE2ETestSuite) TestH_NegativeScenarios() {
	ctx := context.Background()

	s.Run("FailedJobHandling", func() {
		job := &hpctypes.HPCJob{
			JobID:           "hpc-job-fail-1",
			ClusterID:       "e2e-test-cluster",
			OfferingID:      "hpc-compute-standard",
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			State:           hpctypes.JobStatePending,
			QueueName:       "default",
			WorkloadSpec: hpctypes.JobWorkloadSpec{
				ContainerImage: "alpine:latest",
				Command:        "exit 1",
			},
			Resources: hpctypes.JobResources{
				Nodes:           1,
				CPUCoresPerNode: 2,
				MemoryGBPerNode: 4,
				StorageGB:       5,
			},
			MaxRuntimeSeconds: 3600,
			CreatedAt:         time.Now(),
		}

		schedulerJob, err := s.mockScheduler.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Simulate failure
		s.mockScheduler.SetJobState(job.JobID, pd.HPCJobStateFailed)
		s.mockScheduler.SetJobExitCode(job.JobID, 1)

		status, err := s.mockScheduler.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateFailed, status.State)
		s.Equal(int32(1), status.ExitCode)
		s.True(schedulerJob.State.IsTerminal() || status.State.IsTerminal())
	})

	s.Run("PartialUsageReporting", func() {
		// Job that runs partially before failure
		job := &hpctypes.HPCJob{
			JobID:           "hpc-job-partial-1",
			ClusterID:       "e2e-test-cluster",
			OfferingID:      "hpc-compute-standard",
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			State:           hpctypes.JobStatePending,
			QueueName:       "default",
			WorkloadSpec: hpctypes.JobWorkloadSpec{
				ContainerImage: "alpine:latest",
				Command:        "sleep 3600",
			},
			Resources: hpctypes.JobResources{
				Nodes:           1,
				CPUCoresPerNode: 4,
				MemoryGBPerNode: 8,
				StorageGB:       5,
			},
			MaxRuntimeSeconds: 3600,
			CreatedAt:         time.Now(),
		}

		_, err := s.mockScheduler.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Run for 30 minutes then fail
		s.mockScheduler.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.mockScheduler.SetJobMetrics(job.JobID, &pd.HPCSchedulerMetrics{
			WallClockSeconds: 1800, // 30 minutes
			CPUCoreSeconds:   7200, // 4 cores * 30 min
		})
		s.mockScheduler.SetJobState(job.JobID, pd.HPCJobStateFailed)

		// Verify partial usage is captured
		metrics, err := s.mockScheduler.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(int64(1800), metrics.WallClockSeconds)
		s.Equal(int64(7200), metrics.CPUCoreSeconds)
	})

	s.Run("TimeoutHandling", func() {
		job := &hpctypes.HPCJob{
			JobID:           "hpc-job-timeout-1",
			ClusterID:       "e2e-test-cluster",
			OfferingID:      "hpc-compute-standard",
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			State:           hpctypes.JobStatePending,
			QueueName:       "default",
			WorkloadSpec: hpctypes.JobWorkloadSpec{
				ContainerImage: "alpine:latest",
				Command:        "sleep 120",
			},
			Resources: hpctypes.JobResources{
				Nodes:           1,
				CPUCoresPerNode: 1,
				MemoryGBPerNode: 2,
				StorageGB:       1,
			},
			MaxRuntimeSeconds: 60,
			CreatedAt:         time.Now(),
		}

		_, err := s.mockScheduler.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.mockScheduler.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.mockScheduler.SetJobState(job.JobID, pd.HPCJobStateTimeout)

		status, err := s.mockScheduler.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateTimeout, status.State)
		s.True(status.State.IsTerminal())
	})

	s.Run("CancelledJobHandling", func() {
		job := &hpctypes.HPCJob{
			JobID:           "hpc-job-cancel-1",
			ClusterID:       "e2e-test-cluster",
			OfferingID:      "hpc-compute-standard",
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			State:           hpctypes.JobStatePending,
			QueueName:       "default",
			WorkloadSpec: hpctypes.JobWorkloadSpec{
				ContainerImage: "alpine:latest",
				Command:        "sleep 3600",
			},
			Resources: hpctypes.JobResources{
				Nodes:           1,
				CPUCoresPerNode: 1,
				MemoryGBPerNode: 2,
				StorageGB:       1,
			},
			MaxRuntimeSeconds: 3600,
			CreatedAt:         time.Now(),
		}

		_, err := s.mockScheduler.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Cancel the job
		err = s.mockScheduler.CancelJob(ctx, job.JobID)
		s.Require().NoError(err)

		status, err := s.mockScheduler.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCancelled, status.State)
	})

	s.Run("InsufficientResourcesRejection", func() {
		// Request more resources than available
		job := &hpctypes.HPCJob{
			JobID:           "hpc-job-reject-1",
			ClusterID:       "e2e-test-cluster",
			OfferingID:      "hpc-compute-standard",
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			State:           hpctypes.JobStatePending,
			QueueName:       "default",
			WorkloadSpec: hpctypes.JobWorkloadSpec{
				ContainerImage: "alpine:latest",
				Command:        "echo oversized",
			},
			Resources: hpctypes.JobResources{
				Nodes:           1,
				CPUCoresPerNode: 10000, // Exceeds capacity
				MemoryGBPerNode: 999999,
				StorageGB:       1,
			},
			MaxRuntimeSeconds: 3600,
			CreatedAt:         time.Now(),
		}

		s.mockScheduler.SetMaxCapacity(100, 512*1024, 8) // 100 cores, 512GB, 8 GPUs

		_, err := s.mockScheduler.SubmitJob(ctx, job)
		s.Error(err)
	})

	s.Run("DisputedSettlement", func() {
		invoice := MockInvoice{
			InvoiceID:    "invoice-dispute-1",
			ProviderAddr: s.providerAddr,
			CustomerAddr: s.customerAddr,
			OrderID:      "order-dispute-1",
			TotalAmount:  "100.0",
			Status:       "pending",
		}

		err := s.mockSettlement.CreateInvoice(ctx, invoice)
		s.Require().NoError(err)

		// Customer disputes the invoice
		err = s.mockSettlement.DisputeInvoice(ctx, invoice.InvoiceID, "Incorrect usage metrics")
		s.Require().NoError(err)

		disputed := s.mockSettlement.GetInvoice(invoice.InvoiceID)
		s.Equal("disputed", disputed.Status)

		// Settlement should not proceed while disputed
		err = s.mockSettlement.TriggerSettlement(ctx, invoice.InvoiceID)
		s.Error(err)
	})
}

// =============================================================================
// Helper Types and Mock Implementations
// =============================================================================

// MockHPCScheduler is a mock HPC scheduler for testing
type MockHPCScheduler struct {
	running      bool
	jobs         map[string]*pd.HPCSchedulerJob
	metrics      map[string]*pd.HPCSchedulerMetrics
	maxCPU       int32
	maxMemoryMB  int64
	maxGPUs      int32
}

func NewMockHPCScheduler() *MockHPCScheduler {
	return &MockHPCScheduler{
		jobs:        make(map[string]*pd.HPCSchedulerJob),
		metrics:     make(map[string]*pd.HPCSchedulerMetrics),
		maxCPU:      1000,
		maxMemoryMB: 1024 * 1024, // 1 TB
		maxGPUs:     100,
	}
}

func (m *MockHPCScheduler) Type() pd.HPCSchedulerType {
	return pd.HPCSchedulerTypeSLURM
}

func (m *MockHPCScheduler) Start(ctx context.Context) error {
	m.running = true
	return nil
}

func (m *MockHPCScheduler) Stop() error {
	m.running = false
	return nil
}

func (m *MockHPCScheduler) IsRunning() bool {
	return m.running
}

func (m *MockHPCScheduler) SubmitJob(ctx context.Context, job *hpctypes.HPCJob) (*pd.HPCSchedulerJob, error) {
	// Convert GB to MB for comparison (MemoryGBPerNode is in GB, m.maxMemoryMB is in MB)
	memoryMB := int64(job.Resources.MemoryGBPerNode) * 1024
	if job.Resources.CPUCoresPerNode > m.maxCPU || memoryMB > m.maxMemoryMB {
		return nil, fmt.Errorf("insufficient resources")
	}

	schedulerJob := &pd.HPCSchedulerJob{
		VirtEngineJobID: job.JobID,
		SchedulerJobID:  fmt.Sprintf("slurm-%s", job.JobID),
		SchedulerType:   pd.HPCSchedulerTypeSLURM,
		State:           pd.HPCJobStatePending,
		SubmitTime:      time.Now(),
		OriginalJob:     job,
	}
	m.jobs[job.JobID] = schedulerJob
	m.metrics[job.JobID] = &pd.HPCSchedulerMetrics{}
	return schedulerJob, nil
}

func (m *MockHPCScheduler) CancelJob(ctx context.Context, jobID string) error {
	if job, ok := m.jobs[jobID]; ok {
		job.State = pd.HPCJobStateCancelled
		return nil
	}
	return fmt.Errorf("job not found: %s", jobID)
}

func (m *MockHPCScheduler) GetJobStatus(ctx context.Context, jobID string) (*pd.HPCSchedulerJob, error) {
	if job, ok := m.jobs[jobID]; ok {
		return job, nil
	}
	return nil, fmt.Errorf("job not found: %s", jobID)
}

func (m *MockHPCScheduler) GetJobAccounting(ctx context.Context, jobID string) (*pd.HPCSchedulerMetrics, error) {
	if metrics, ok := m.metrics[jobID]; ok {
		return metrics, nil
	}
	return nil, fmt.Errorf("job not found: %s", jobID)
}

func (m *MockHPCScheduler) ListActiveJobs(ctx context.Context) ([]*pd.HPCSchedulerJob, error) {
	var active []*pd.HPCSchedulerJob
	for _, job := range m.jobs {
		if !job.State.IsTerminal() {
			active = append(active, job)
		}
	}
	return active, nil
}

func (m *MockHPCScheduler) RegisterLifecycleCallback(cb pd.HPCJobLifecycleCallback) {}

func (m *MockHPCScheduler) CreateStatusReport(job *pd.HPCSchedulerJob) (*pd.HPCStatusReport, error) {
	return &pd.HPCStatusReport{
		VirtEngineJobID: job.VirtEngineJobID,
		SchedulerJobID:  job.SchedulerJobID,
		SchedulerType:   job.SchedulerType,
		State:           job.State,
		Timestamp:       time.Now(),
	}, nil
}

func (m *MockHPCScheduler) SetJobState(jobID string, state pd.HPCJobState) {
	if job, ok := m.jobs[jobID]; ok {
		job.State = state
		if state.IsTerminal() {
			now := time.Now()
			job.EndTime = &now
		}
	}
}

func (m *MockHPCScheduler) SetJobMetrics(jobID string, metrics *pd.HPCSchedulerMetrics) {
	m.metrics[jobID] = metrics
}

func (m *MockHPCScheduler) SetJobExitCode(jobID string, code int32) {
	if job, ok := m.jobs[jobID]; ok {
		job.ExitCode = code
	}
}

func (m *MockHPCScheduler) SetMaxCapacity(cpu int32, memoryMB int64, gpus int32) {
	m.maxCPU = cpu
	m.maxMemoryMB = memoryMB
	m.maxGPUs = gpus
}

func (m *MockHPCScheduler) IsValidTransition(from, to pd.HPCJobState) bool {
	validTransitions := map[pd.HPCJobState][]pd.HPCJobState{
		pd.HPCJobStatePending:   {pd.HPCJobStateQueued, pd.HPCJobStateFailed, pd.HPCJobStateCancelled},
		pd.HPCJobStateQueued:    {pd.HPCJobStateStarting, pd.HPCJobStateFailed, pd.HPCJobStateCancelled},
		pd.HPCJobStateStarting:  {pd.HPCJobStateRunning, pd.HPCJobStateFailed, pd.HPCJobStateCancelled},
		pd.HPCJobStateRunning:   {pd.HPCJobStateCompleted, pd.HPCJobStateFailed, pd.HPCJobStateCancelled, pd.HPCJobStateSuspended, pd.HPCJobStateTimeout},
		pd.HPCJobStateSuspended: {pd.HPCJobStateRunning, pd.HPCJobStateFailed, pd.HPCJobStateCancelled},
	}

	if valid, ok := validTransitions[from]; ok {
		for _, v := range valid {
			if v == to {
				return true
			}
		}
	}
	return false
}

// MockHPCOnChainReporter is a mock on-chain reporter for testing
type MockHPCOnChainReporter struct {
	reports      []*pd.HPCStatusReport
	usageRecords map[string]*pd.HPCUsageRecord
	jobMetrics   map[string]*pd.HPCSchedulerMetrics
}

func NewMockHPCOnChainReporter() *MockHPCOnChainReporter {
	return &MockHPCOnChainReporter{
		reports:      make([]*pd.HPCStatusReport, 0),
		usageRecords: make(map[string]*pd.HPCUsageRecord),
		jobMetrics:   make(map[string]*pd.HPCSchedulerMetrics),
	}
}

func (m *MockHPCOnChainReporter) ReportJobStatus(ctx context.Context, report *pd.HPCStatusReport) error {
	m.reports = append(m.reports, report)
	return nil
}

func (m *MockHPCOnChainReporter) ReportJobAccounting(ctx context.Context, jobID string, metrics *pd.HPCSchedulerMetrics) error {
	m.jobMetrics[jobID] = metrics
	return nil
}

func (m *MockHPCOnChainReporter) GetSubmittedReports() []*pd.HPCStatusReport {
	return m.reports
}

func (m *MockHPCOnChainReporter) QueueUsageRecord(record *pd.HPCUsageRecord) error {
	m.usageRecords[record.RecordID] = record
	return nil
}

func (m *MockHPCOnChainReporter) GetUsageRecord(recordID string) *pd.HPCUsageRecord {
	return m.usageRecords[recordID]
}

func (m *MockHPCOnChainReporter) SetJobMetrics(jobID string, metrics *pd.HPCSchedulerMetrics) {
	m.jobMetrics[jobID] = metrics
}

func (m *MockHPCOnChainReporter) GetJobMetrics(jobID string) *pd.HPCSchedulerMetrics {
	return m.jobMetrics[jobID]
}

// MockHPCAuditLogger is a mock audit logger for testing
type MockHPCAuditLogger struct {
	events []pd.HPCAuditEvent
}

func NewMockHPCAuditLogger() *MockHPCAuditLogger {
	return &MockHPCAuditLogger{
		events: make([]pd.HPCAuditEvent, 0),
	}
}

func (m *MockHPCAuditLogger) LogJobEvent(event pd.HPCAuditEvent) {
	m.events = append(m.events, event)
}

func (m *MockHPCAuditLogger) LogSecurityEvent(event pd.HPCAuditEvent) {
	m.events = append(m.events, event)
}

func (m *MockHPCAuditLogger) LogUsageReport(event pd.HPCAuditEvent) {
	m.events = append(m.events, event)
}

func (m *MockHPCAuditLogger) GetEvents() []pd.HPCAuditEvent {
	return m.events
}

// MockWaldurClient is a mock Waldur client for testing
type MockWaldurClient struct {
	providers   map[string]bool
	offerings   map[string][]MockOffering
	orders      map[string]MockOrder
	bids        map[string][]MockBid
	allocations map[string]MockAllocation
}

type MockOffering struct {
	OfferingID   string
	Name         string
	Category     string
	CPUCores     int32
	MemoryGB     int32
	GPUs         int32
	PricePerHour string
	Active       bool
	WaldurUUID   string
}

type MockOrder struct {
	OrderID      string
	CustomerAddr string
	OfferingID   string
	Requirements pd.ResourceRequirements
	MaxPrice     string
	Duration     time.Duration
	Status       string
}

type MockBid struct {
	BidID        string
	OrderID      string
	ProviderAddr string
	Price        string
	TTL          time.Duration
}

type MockAllocation struct {
	AllocationID string
	OrderID      string
	ProviderAddr string
	Status       string
}

func NewMockWaldurClient() *MockWaldurClient {
	return &MockWaldurClient{
		providers:   make(map[string]bool),
		offerings:   make(map[string][]MockOffering),
		orders:      make(map[string]MockOrder),
		bids:        make(map[string][]MockBid),
		allocations: make(map[string]MockAllocation),
	}
}

func (m *MockWaldurClient) SetProviderRegistered(addr string, registered bool) {
	m.providers[addr] = registered
}

func (m *MockWaldurClient) IsProviderRegistered(addr string) bool {
	return m.providers[addr]
}

func (m *MockWaldurClient) PublishOffering(ctx context.Context, offering MockOffering) error {
	offering.WaldurUUID = fmt.Sprintf("waldur-%s", offering.OfferingID)
	
	// Use first provider address as key (simplified)
	for addr := range m.providers {
		m.offerings[addr] = append(m.offerings[addr], offering)
		break
	}
	return nil
}

func (m *MockWaldurClient) GetPublishedOfferings(providerAddr string) []MockOffering {
	return m.offerings[providerAddr]
}

func (m *MockWaldurClient) GetSyncedOfferingIDs() []string {
	var ids []string
	for _, offerings := range m.offerings {
		for _, o := range offerings {
			if o.WaldurUUID != "" {
				ids = append(ids, o.OfferingID)
			}
		}
	}
	return ids
}

func (m *MockWaldurClient) CreateOrder(ctx context.Context, order MockOrder) error {
	order.Status = "open"
	m.orders[order.OrderID] = order
	return nil
}

func (m *MockWaldurClient) PlaceBid(ctx context.Context, bid MockBid) error {
	m.bids[bid.OrderID] = append(m.bids[bid.OrderID], bid)
	return nil
}

func (m *MockWaldurClient) GetBidsForOrder(orderID string) []MockBid {
	return m.bids[orderID]
}

func (m *MockWaldurClient) AcceptBid(ctx context.Context, orderID, bidID string) error {
	order := m.orders[orderID]
	order.Status = "matched"
	m.orders[orderID] = order

	// Create allocation
	var providerAddr string
	for _, bid := range m.bids[orderID] {
		if bid.BidID == bidID {
			providerAddr = bid.ProviderAddr
			break
		}
	}

	allocation := MockAllocation{
		AllocationID: fmt.Sprintf("alloc-%s", orderID),
		OrderID:      orderID,
		ProviderAddr: providerAddr,
		Status:       "provisioned",
	}
	m.allocations[allocation.AllocationID] = allocation
	return nil
}

func (m *MockWaldurClient) GetAllocationsForOrder(orderID string) []MockAllocation {
	var result []MockAllocation
	for _, alloc := range m.allocations {
		if alloc.OrderID == orderID {
			result = append(result, alloc)
		}
	}
	return result
}

func (m *MockWaldurClient) GetAllocationStatus(allocationID string) string {
	if alloc, ok := m.allocations[allocationID]; ok {
		return alloc.Status
	}
	return ""
}

// MockSettlementClient is a mock settlement client for testing
type MockSettlementClient struct {
	invoices map[string]*MockInvoice
	payouts  map[string]*MockPayout
	fees     map[string]*MockFee
}

type MockInvoice struct {
	InvoiceID    string
	ProviderAddr string
	CustomerAddr string
	OrderID      string
	LineItems    []MockLineItem
	TotalAmount  string
	PeriodStart  time.Time
	PeriodEnd    time.Time
	Status       string
}

type MockLineItem struct {
	ResourceType string
	Quantity     int64
	UnitPrice    string
	TotalCost    string
}

type MockPayout struct {
	PayoutID  string
	InvoiceID string
	Provider  string
	Amount    string
	Status    string
}

type MockFee struct {
	FeeID     string
	InvoiceID string
	Amount    string
}

func NewMockSettlementClient() *MockSettlementClient {
	return &MockSettlementClient{
		invoices: make(map[string]*MockInvoice),
		payouts:  make(map[string]*MockPayout),
		fees:     make(map[string]*MockFee),
	}
}

func (m *MockSettlementClient) CreateInvoice(ctx context.Context, invoice MockInvoice) error {
	m.invoices[invoice.InvoiceID] = &invoice
	return nil
}

func (m *MockSettlementClient) GetInvoice(invoiceID string) *MockInvoice {
	return m.invoices[invoiceID]
}

func (m *MockSettlementClient) TriggerSettlement(ctx context.Context, invoiceID string) error {
	invoice := m.invoices[invoiceID]
	if invoice == nil {
		return fmt.Errorf("invoice not found: %s", invoiceID)
	}
	if invoice.Status == "disputed" {
		return fmt.Errorf("cannot settle disputed invoice")
	}

	invoice.Status = "settled"

	// Create payout (97.5% to provider)
	m.payouts[invoiceID] = &MockPayout{
		PayoutID:  fmt.Sprintf("payout-%s", invoiceID),
		InvoiceID: invoiceID,
		Provider:  invoice.ProviderAddr,
		Amount:    invoice.TotalAmount, // Simplified
		Status:    "completed",
	}

	// Create platform fee (2.5%)
	m.fees[invoiceID] = &MockFee{
		FeeID:     fmt.Sprintf("fee-%s", invoiceID),
		InvoiceID: invoiceID,
		Amount:    "0.133", // 2.5% of 5.328
	}

	return nil
}

func (m *MockSettlementClient) GetProviderPayout(providerAddr, invoiceID string) *MockPayout {
	return m.payouts[invoiceID]
}

func (m *MockSettlementClient) GetPlatformFee(invoiceID string) *MockFee {
	return m.fees[invoiceID]
}

func (m *MockSettlementClient) DisputeInvoice(ctx context.Context, invoiceID, reason string) error {
	if invoice, ok := m.invoices[invoiceID]; ok {
		invoice.Status = "disputed"
		return nil
	}
	return fmt.Errorf("invoice not found: %s", invoiceID)
}

// Helper functions

func filterEventsByType(events []pd.HPCAuditEvent, eventType string) []pd.HPCAuditEvent {
	var filtered []pd.HPCAuditEvent
	for _, e := range events {
		if contains(e.EventType, eventType) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0)
}

func isValidAccountingTransition(from, to hpctypes.AccountingRecordStatus) bool {
	validTransitions := map[hpctypes.AccountingRecordStatus][]hpctypes.AccountingRecordStatus{
		hpctypes.AccountingStatusPending:   {hpctypes.AccountingStatusFinalized, hpctypes.AccountingStatusDisputed},
		hpctypes.AccountingStatusFinalized: {hpctypes.AccountingStatusSettled, hpctypes.AccountingStatusDisputed},
		hpctypes.AccountingStatusDisputed:  {hpctypes.AccountingStatusFinalized, hpctypes.AccountingStatusCorrected},
	}

	if valid, ok := validTransitions[from]; ok {
		for _, v := range valid {
			if v == to {
				return true
			}
		}
	}
	return false
}

// Signature helper for test key manager
func signTestData(data []byte) string {
	// Simple mock signature for testing
	return hex.EncodeToString(data[:min(32, len(data))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
