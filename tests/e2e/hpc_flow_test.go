//go:build e2e.integration

// Package e2e contains end-to-end integration tests.
//
// VE-15D: E2E test for complete HPC job flow from submission to settlement.
// Tests the complete provider workflow:
// 1. HPC provider registers with SLURM cluster
// 2. HPC offering listed
// 3. Customer submits job
// 4. Job routed to appropriate cluster
// 5. Job executed via SLURM (mocked)
// 6. Usage recorded
// 7. Billing/settlement completed
package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/tests/e2e/fixtures"
	"github.com/virtengine/virtengine/tests/e2e/mocks"
	"github.com/virtengine/virtengine/testutil"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCFlowE2ETestSuite tests the complete HPC job flow from submission to settlement.
type HPCFlowE2ETestSuite struct {
	*testutil.NetworkTestSuite

	// Test addresses
	providerAddr string
	customerAddr string

	// Mock components
	slurmMock      *mocks.MockSLURMIntegration
	usageReporter  *MockUsageReporterE2E
	settlementMock *MockSettlementE2E
	auditLogger    *MockAuditLoggerE2E

	// Test data
	testCluster  *hpctypes.HPCCluster
	testOffering *hpctypes.HPCOffering

	// Lifecycle tracking
	lifecycleEvents []LifecycleEvent
}

// LifecycleEvent tracks job lifecycle events for verification.
type LifecycleEvent struct {
	JobID     string
	FromState pd.HPCJobState
	ToState   pd.HPCJobState
	Timestamp time.Time
}

func TestHPCFlowE2E(t *testing.T) {
	suite.Run(t, &HPCFlowE2ETestSuite{
		NetworkTestSuite: testutil.NewNetworkTestSuite(nil, &HPCFlowE2ETestSuite{}),
	})
}

func (s *HPCFlowE2ETestSuite) SetupSuite() {
	s.NetworkTestSuite.SetupSuite()

	val := s.Network().Validators[0]
	s.providerAddr = val.Address.String()
	s.customerAddr = val.Address.String()

	// Initialize mock components
	s.slurmMock = mocks.NewMockSLURMIntegration()
	s.usageReporter = NewMockUsageReporterE2E()
	s.settlementMock = NewMockSettlementE2E()
	s.auditLogger = NewMockAuditLoggerE2E()
	s.lifecycleEvents = make([]LifecycleEvent, 0)

	// Register lifecycle callback
	s.slurmMock.RegisterLifecycleCallback(s.onJobLifecycleEvent)

	// Create test cluster and offering using fixtures
	clusterConfig := fixtures.DefaultTestClusterConfig()
	clusterConfig.ProviderAddr = s.providerAddr
	s.testCluster = fixtures.CreateTestCluster(clusterConfig)

	offeringConfig := fixtures.DefaultTestOfferingConfig()
	offeringConfig.ProviderAddr = s.providerAddr
	s.testOffering = fixtures.CreateTestOffering(offeringConfig)
}

func (s *HPCFlowE2ETestSuite) TearDownSuite() {
	if s.slurmMock != nil && s.slurmMock.IsRunning() {
		_ = s.slurmMock.Stop()
	}
	s.NetworkTestSuite.TearDownSuite()
}

func (s *HPCFlowE2ETestSuite) onJobLifecycleEvent(job *pd.HPCSchedulerJob, event pd.HPCJobLifecycleEvent, prevState pd.HPCJobState) {
	s.lifecycleEvents = append(s.lifecycleEvents, LifecycleEvent{
		JobID:     job.VirtEngineJobID,
		FromState: prevState,
		ToState:   job.State,
		Timestamp: time.Now(),
	})
}

// =============================================================================
// 1. HPC Provider Registration Tests
// =============================================================================

func (s *HPCFlowE2ETestSuite) Test01_HPCProviderRegistration() {
	ctx := context.Background()

	s.Run("StartSLURMScheduler", func() {
		err := s.slurmMock.Start(ctx)
		s.Require().NoError(err)
		s.True(s.slurmMock.IsRunning())
	})

	s.Run("RegisterSLURMCluster", func() {
		// Register the default test cluster
		cluster := mocks.DefaultTestCluster()
		s.slurmMock.RegisterCluster(cluster)

		// Verify cluster is registered
		registered, exists := s.slurmMock.GetCluster(cluster.ClusterID)
		s.True(exists)
		s.Equal(cluster.ClusterID, registered.ClusterID)
		s.Equal(cluster.Name, registered.Name)
		s.Equal(cluster.Region, registered.Region)
	})

	s.Run("VerifyClusterCapabilities", func() {
		cluster, exists := s.slurmMock.GetCluster("e2e-slurm-cluster")
		s.True(exists)

		// Verify partitions
		s.Len(cluster.Partitions, 3)

		// Verify default partition
		var defaultPartition *mocks.SLURMPartition
		for i := range cluster.Partitions {
			if cluster.Partitions[i].Name == "default" {
				defaultPartition = &cluster.Partitions[i]
				break
			}
		}
		s.NotNil(defaultPartition)
		s.Equal(int32(50), defaultPartition.Nodes)
		s.Equal("up", defaultPartition.State)

		// Verify GPU partition
		var gpuPartition *mocks.SLURMPartition
		for i := range cluster.Partitions {
			if cluster.Partitions[i].Name == "gpu" {
				gpuPartition = &cluster.Partitions[i]
				break
			}
		}
		s.NotNil(gpuPartition)
		s.Contains(gpuPartition.Features, "a100")
	})

	s.Run("VerifyMultipleClusterSupport", func() {
		// Register additional clusters
		s.slurmMock.RegisterCluster(&mocks.SLURMCluster{
			ClusterID:    "e2e-cluster-eu",
			Name:         "E2E EU Cluster",
			Region:       "eu-west",
			SLURMVersion: "23.02.4",
			TotalNodes:   50,
			Partitions: []mocks.SLURMPartition{
				{Name: "default", Nodes: 50, State: "up"},
			},
		})

		clusters := s.slurmMock.GetClusters()
		s.GreaterOrEqual(len(clusters), 2)
	})
}

// =============================================================================
// 2. HPC Offering Publication Tests
// =============================================================================

func (s *HPCFlowE2ETestSuite) Test02_HPCOfferingPublication() {
	s.Run("VerifyOfferingConfiguration", func() {
		s.NotNil(s.testOffering)
		s.Equal("hpc-compute-standard", s.testOffering.OfferingID)
		s.Equal("e2e-slurm-cluster", s.testOffering.ClusterID)
		s.True(s.testOffering.Active)
		s.True(s.testOffering.SupportsCustomWorkloads)
	})

	s.Run("VerifyOfferingPricing", func() {
		pricing := s.testOffering.Pricing
		s.NotEmpty(pricing.BaseNodeHourPrice)
		s.NotEmpty(pricing.CPUCoreHourPrice)
		s.NotEmpty(pricing.Currency)
		s.Equal("uvirt", pricing.Currency)
	})

	s.Run("VerifyQueueOptions", func() {
		s.GreaterOrEqual(len(s.testOffering.QueueOptions), 2)

		// Verify default queue
		var defaultQueue *hpctypes.QueueOption
		for i := range s.testOffering.QueueOptions {
			if s.testOffering.QueueOptions[i].PartitionName == "default" {
				defaultQueue = &s.testOffering.QueueOptions[i]
				break
			}
		}
		s.NotNil(defaultQueue)
		s.Equal("Standard Compute", defaultQueue.DisplayName)
	})

	s.Run("VerifyIdentityThreshold", func() {
		s.Equal(int32(70), s.testOffering.RequiredIdentityThreshold)
	})
}

// =============================================================================
// 3. Job Submission Tests
// =============================================================================

func (s *HPCFlowE2ETestSuite) Test03_JobSubmission() {
	ctx := context.Background()

	s.Run("SubmitStandardComputeJob", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)

		s.Equal(job.JobID, schedulerJob.VirtEngineJobID)
		s.NotEmpty(schedulerJob.SchedulerJobID)
		s.Equal(pd.HPCSchedulerTypeSLURM, schedulerJob.SchedulerType)
		s.Equal(pd.HPCJobStatePending, schedulerJob.State)
	})

	s.Run("SubmitGPUJob", func() {
		job := fixtures.GPUComputeJob(s.providerAddr, s.customerAddr)

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)
		s.Equal(pd.HPCJobStatePending, schedulerJob.State)
	})

	s.Run("SubmitMultiNodeJob", func() {
		job := fixtures.MultiNodeJob(s.providerAddr, s.customerAddr, 4)

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)
	})

	s.Run("RejectOversizedJob", func() {
		job := fixtures.OversizedJob(s.providerAddr, s.customerAddr)

		// Set capacity limits
		s.slurmMock.SetMaxCapacity(1000, 1024*1024, 100)

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Error(err)
		s.Contains(err.Error(), "insufficient")
	})
}

// =============================================================================
// 4. Job Routing Verification Tests
// =============================================================================

func (s *HPCFlowE2ETestSuite) Test04_JobRoutingVerification() {
	ctx := context.Background()

	s.Run("VerifyRoutingDecisionCreated", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = "routing-test-job-1"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Verify routing decision was recorded
		decision, exists := s.slurmMock.GetRoutingDecisionForJob(job.JobID)
		s.True(exists)
		s.NotNil(decision)
		s.Equal(job.JobID, decision.JobID)
		s.NotEmpty(decision.SelectedCluster)
	})

	s.Run("VerifyRoutingDecisionAuditability", func() {
		decisions := s.slurmMock.GetRoutingDecisions()
		s.GreaterOrEqual(len(decisions), 1)

		for _, decision := range decisions {
			// Each decision should have an audit hash
			s.NotEmpty(decision.DecisionHash)

			// Each decision should have scoring factors
			s.NotEmpty(decision.ScoringFactors)

			// Each decision should have a reason
			s.NotEmpty(decision.Reason)

			// Each decision should have candidate clusters
			s.GreaterOrEqual(len(decision.CandidateClusters), 1)
		}
	})

	s.Run("VerifyScoringFactors", func() {
		decision, exists := s.slurmMock.GetRoutingDecisionForJob("routing-test-job-1")
		s.True(exists)

		// Verify expected scoring factors
		expectedFactors := []string{
			"resource_availability",
			"queue_depth",
			"geographic_proximity",
			"price_competitiveness",
		}

		for _, factor := range expectedFactors {
			_, hasKey := decision.ScoringFactors[factor]
			s.True(hasKey, "Expected scoring factor: %s", factor)
		}
	})
}

// =============================================================================
// 5. Job Execution Tests (SLURM Integration)
// =============================================================================

func (s *HPCFlowE2ETestSuite) Test05_JobExecution() {
	ctx := context.Background()

	s.Run("ExecuteJobLifecycle", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = "execution-test-job-1"

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStatePending, schedulerJob.State)

		// Progress through states
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)
		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateQueued, status.State)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		status, err = s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateRunning, status.State)
		s.NotNil(status.StartTime)

		// Complete the job
		metrics := fixtures.StandardJobMetrics(3600)
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobExitCode(job.JobID, 0)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		status, err = s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCompleted, status.State)
		s.True(status.State.IsTerminal())
		s.NotNil(status.EndTime)
	})

	s.Run("VerifyExecutionRecord", func() {
		record, exists := s.slurmMock.GetExecutionRecord("execution-test-job-1")
		s.True(exists)
		s.NotNil(record)

		s.Equal("execution-test-job-1", record.JobID)
		s.NotEmpty(record.ClusterID)
		s.NotEmpty(record.SLURMJobID)
		s.NotNil(record.StartTime)
		s.NotNil(record.EndTime)
		s.Equal(int32(0), record.ExitCode)
	})

	s.Run("VerifyLifecycleEvents", func() {
		// Filter events for our test job
		var jobEvents []LifecycleEvent
		for _, e := range s.lifecycleEvents {
			if e.JobID == "execution-test-job-1" {
				jobEvents = append(jobEvents, e)
			}
		}

		// Should have events for each state transition
		s.GreaterOrEqual(len(jobEvents), 3)
	})

	s.Run("SimulateCompleteExecution", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = "simulated-execution-job"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		metrics := fixtures.StandardJobMetrics(60)
		err = s.slurmMock.SimulateJobExecution(ctx, job.JobID, 100, true, metrics)
		s.Require().NoError(err)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCompleted, status.State)
	})
}

// =============================================================================
// 6. Usage Recording Tests
// =============================================================================

func (s *HPCFlowE2ETestSuite) Test06_UsageRecording() {
	ctx := context.Background()

	s.Run("RecordJobUsageMetrics", func() {
		jobID := "execution-test-job-1"

		metrics, err := s.slurmMock.GetJobAccounting(ctx, jobID)
		s.Require().NoError(err)
		s.NotNil(metrics)

		s.Equal(int64(3600), metrics.WallClockSeconds)
		s.Greater(metrics.CPUCoreSeconds, int64(0))
		s.Greater(metrics.MemoryGBSeconds, int64(0))
	})

	s.Run("CreateUsageRecord", func() {
		jobID := "execution-test-job-1"

		metrics, _ := s.slurmMock.GetJobAccounting(ctx, jobID)

		record := &UsageRecordE2E{
			RecordID:        fmt.Sprintf("usage-%s-%d", jobID, time.Now().UnixNano()),
			JobID:           jobID,
			ClusterID:       "e2e-slurm-cluster",
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			PeriodStart:     time.Now().Add(-time.Hour),
			PeriodEnd:       time.Now(),
			Metrics:         metrics,
			IsFinal:         true,
			JobState:        pd.HPCJobStateCompleted,
		}

		err := s.usageReporter.RecordUsage(record)
		s.Require().NoError(err)

		// Verify record was stored
		storedRecords := s.usageReporter.GetRecordsForJob(jobID)
		s.GreaterOrEqual(len(storedRecords), 1)
	})

	s.Run("VerifyUsageRecordIntegrity", func() {
		records := s.usageReporter.GetRecordsForJob("execution-test-job-1")
		s.GreaterOrEqual(len(records), 1)

		record := records[0]
		s.NotEmpty(record.RecordID)
		s.NotNil(record.Metrics)
		s.True(record.IsFinal)
		s.Equal(pd.HPCJobStateCompleted, record.JobState)
	})

	s.Run("SubmitUsageReportOnChain", func() {
		job, _ := s.slurmMock.GetJobStatus(ctx, "execution-test-job-1")

		report, err := s.slurmMock.CreateStatusReport(job)
		s.Require().NoError(err)
		s.NotNil(report)

		s.NotEmpty(report.ProviderAddress)
		s.NotEmpty(report.VirtEngineJobID)
		s.NotEmpty(report.Signature)
		s.Equal(pd.HPCJobStateCompleted, report.State)

		// Submit to mock reporter
		err = s.usageReporter.SubmitStatusReport(report)
		s.Require().NoError(err)

		reports := s.usageReporter.GetSubmittedReports()
		s.GreaterOrEqual(len(reports), 1)
	})
}

// =============================================================================
// 7. Billing and Settlement Tests
// =============================================================================

func (s *HPCFlowE2ETestSuite) Test07_BillingAndSettlement() {
	ctx := context.Background()

	var invoiceID string

	s.Run("CalculateBillableAmount", func() {
		records := s.usageReporter.GetRecordsForJob("execution-test-job-1")
		s.Require().GreaterOrEqual(len(records), 1)

		record := records[0]
		metrics := record.Metrics

		// Calculate billable amount based on pricing
		// BaseNodeHourPrice: 10.0, CPUCoreHourPrice: 0.10, MemoryGBHourPrice: 0.01
		nodeHours := float64(metrics.WallClockSeconds) / 3600.0
		cpuCoreHours := float64(metrics.CPUCoreSeconds) / 3600.0
		memoryGBHours := float64(metrics.MemoryGBSeconds) / 3600.0

		nodeCost := nodeHours * 10.0
		cpuCost := cpuCoreHours * 0.10
		memoryCost := memoryGBHours * 0.01

		totalCost := nodeCost + cpuCost + memoryCost
		s.Greater(totalCost, 0.0)
	})

	s.Run("GenerateInvoice", func() {
		invoiceID = fmt.Sprintf("invoice-e2e-%d", time.Now().UnixNano())

		invoice := &InvoiceE2E{
			InvoiceID:    invoiceID,
			ProviderAddr: s.providerAddr,
			CustomerAddr: s.customerAddr,
			JobID:        "execution-test-job-1",
			LineItems: []LineItemE2E{
				{
					ResourceType: "node-hours",
					Quantity:     sdkmath.LegacyMustNewDecFromStr("1.0"),
					UnitPrice:    "10.0",
					TotalCost:    "10.0",
				},
				{
					ResourceType: "cpu-core-hours",
					Quantity:     sdkmath.LegacyMustNewDecFromStr("28800.0"),
					UnitPrice:    "0.00002778",
					TotalCost:    "0.80",
				},
				{
					ResourceType: "memory-gb-hours",
					Quantity:     sdkmath.LegacyMustNewDecFromStr("57600.0"),
					UnitPrice:    "0.00000278",
					TotalCost:    "0.16",
				},
			},
			TotalAmount: "10.96",
			PeriodStart: time.Now().Add(-time.Hour),
			PeriodEnd:   time.Now(),
			Status:      "pending",
		}

		err := s.settlementMock.CreateInvoice(ctx, invoice)
		s.Require().NoError(err)

		// Verify invoice was created
		stored := s.settlementMock.GetInvoice(invoiceID)
		s.NotNil(stored)
		s.Equal("pending", stored.Status)
	})

	s.Run("VerifyBillableLineItems", func() {
		invoice := s.settlementMock.GetInvoice(invoiceID)
		s.NotNil(invoice)
		s.Len(invoice.LineItems, 3)

		for _, item := range invoice.LineItems {
			s.NotEmpty(item.ResourceType)
			s.NotEmpty(item.TotalCost)
		}
	})

	s.Run("TriggerSettlement", func() {
		err := s.settlementMock.TriggerSettlement(ctx, invoiceID)
		s.Require().NoError(err)

		invoice := s.settlementMock.GetInvoice(invoiceID)
		s.Equal("settled", invoice.Status)
	})

	s.Run("VerifyProviderPayout", func() {
		payout := s.settlementMock.GetProviderPayout(s.providerAddr, invoiceID)
		s.NotNil(payout)
		s.Equal("completed", payout.Status)
		s.NotEmpty(payout.Amount)
	})

	s.Run("VerifyPlatformFee", func() {
		fee := s.settlementMock.GetPlatformFee(invoiceID)
		s.NotNil(fee)
		s.NotEmpty(fee.Amount)
	})

	s.Run("VerifySettlementAuditTrail", func() {
		auditRecords := s.settlementMock.GetAuditTrail(invoiceID)
		s.GreaterOrEqual(len(auditRecords), 2) // Created + Settled
	})
}

// =============================================================================
// 8. Negative Scenario Tests
// =============================================================================

func (s *HPCFlowE2ETestSuite) Test08_NegativeScenarios() {
	ctx := context.Background()

	s.Run("FailedJobHandling", func() {
		job := fixtures.FailingJob(s.providerAddr, s.customerAddr)

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.slurmMock.SetJobExitCode(job.JobID, 1)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateFailed)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateFailed, status.State)
		s.Equal(int32(1), status.ExitCode)
		s.True(status.State.IsTerminal())
	})

	s.Run("TimeoutHandling", func() {
		job := fixtures.TimeoutJob(s.providerAddr, s.customerAddr)

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateTimeout)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateTimeout, status.State)
		s.True(status.State.IsTerminal())
	})

	s.Run("CancelledJobHandling", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = "cancel-test-job"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		err = s.slurmMock.CancelJob(ctx, job.JobID)
		s.Require().NoError(err)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCancelled, status.State)
	})

	s.Run("PartialUsageReporting", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = "partial-usage-job"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		// Set partial metrics before failure
		partialMetrics := fixtures.PartialJobMetrics(1800)
		s.slurmMock.SetJobMetrics(job.JobID, partialMetrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateFailed)

		metrics, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(int64(1800), metrics.WallClockSeconds)
	})

	s.Run("DisputedSettlement", func() {
		invoiceID := fmt.Sprintf("invoice-dispute-%d", time.Now().UnixNano())

		invoice := &InvoiceE2E{
			InvoiceID:    invoiceID,
			ProviderAddr: s.providerAddr,
			CustomerAddr: s.customerAddr,
			JobID:        "dispute-test-job",
			TotalAmount:  "100.0",
			Status:       "pending",
		}

		err := s.settlementMock.CreateInvoice(ctx, invoice)
		s.Require().NoError(err)

		// Dispute the invoice
		err = s.settlementMock.DisputeInvoice(ctx, invoiceID, "Incorrect usage metrics")
		s.Require().NoError(err)

		disputed := s.settlementMock.GetInvoice(invoiceID)
		s.Equal("disputed", disputed.Status)

		// Settlement should fail while disputed
		err = s.settlementMock.TriggerSettlement(ctx, invoiceID)
		s.Error(err)
	})
}

// =============================================================================
// 9. State Transition Validation Tests
// =============================================================================

func (s *HPCFlowE2ETestSuite) Test09_StateTransitionValidation() {
	s.Run("ValidJobStateTransitions", func() {
		validTransitions := map[pd.HPCJobState][]pd.HPCJobState{
			pd.HPCJobStatePending:  {pd.HPCJobStateQueued, pd.HPCJobStateFailed, pd.HPCJobStateCancelled},
			pd.HPCJobStateQueued:   {pd.HPCJobStateStarting, pd.HPCJobStateRunning, pd.HPCJobStateFailed, pd.HPCJobStateCancelled},
			pd.HPCJobStateStarting: {pd.HPCJobStateRunning, pd.HPCJobStateFailed, pd.HPCJobStateCancelled},
			pd.HPCJobStateRunning:  {pd.HPCJobStateCompleted, pd.HPCJobStateFailed, pd.HPCJobStateCancelled, pd.HPCJobStateSuspended, pd.HPCJobStateTimeout},
		}

		for from, toStates := range validTransitions {
			for _, to := range toStates {
				valid := s.slurmMock.IsValidTransition(from, to)
				s.True(valid, "Transition from %s to %s should be valid", from, to)
			}
		}
	})

	s.Run("InvalidJobStateTransitions", func() {
		invalidTransitions := []struct {
			from pd.HPCJobState
			to   pd.HPCJobState
		}{
			{pd.HPCJobStateCompleted, pd.HPCJobStateRunning},
			{pd.HPCJobStateFailed, pd.HPCJobStateCompleted},
			{pd.HPCJobStateCancelled, pd.HPCJobStateRunning},
			{pd.HPCJobStatePending, pd.HPCJobStateCompleted},
		}

		for _, t := range invalidTransitions {
			valid := s.slurmMock.IsValidTransition(t.from, t.to)
			s.False(valid, "Transition from %s to %s should be invalid", t.from, t.to)
		}
	})

	s.Run("TerminalStateVerification", func() {
		terminalStates := []pd.HPCJobState{
			pd.HPCJobStateCompleted,
			pd.HPCJobStateFailed,
			pd.HPCJobStateCancelled,
			pd.HPCJobStateTimeout,
		}

		for _, state := range terminalStates {
			s.True(state.IsTerminal(), "State %s should be terminal", state)
		}

		nonTerminalStates := []pd.HPCJobState{
			pd.HPCJobStatePending,
			pd.HPCJobStateQueued,
			pd.HPCJobStateStarting,
			pd.HPCJobStateRunning,
			pd.HPCJobStateSuspended,
		}

		for _, state := range nonTerminalStates {
			s.False(state.IsTerminal(), "State %s should not be terminal", state)
		}
	})
}

// =============================================================================
// 10. Complete Flow Integration Test
// =============================================================================

func (s *HPCFlowE2ETestSuite) Test10_CompleteFlowIntegration() {
	ctx := context.Background()

	s.Run("EndToEndJobFlow", func() {
		// Step 1: Create and submit job
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "e2e-complete-flow-job"

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStatePending, schedulerJob.State)

		// Step 2: Verify routing decision
		decision, exists := s.slurmMock.GetRoutingDecisionForJob(job.JobID)
		s.True(exists)
		s.NotEmpty(decision.DecisionHash)

		// Step 3: Progress through execution states
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		// Step 4: Set metrics and complete
		metrics := fixtures.StandardJobMetrics(3600)
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobExitCode(job.JobID, 0)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		// Step 5: Verify final state
		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCompleted, status.State)

		// Step 6: Record usage
		usageMetrics, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)

		usageRecord := &UsageRecordE2E{
			RecordID:        fmt.Sprintf("usage-%s", job.JobID),
			JobID:           job.JobID,
			ClusterID:       job.ClusterID,
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			PeriodStart:     time.Now().Add(-time.Hour),
			PeriodEnd:       time.Now(),
			Metrics:         usageMetrics,
			IsFinal:         true,
			JobState:        pd.HPCJobStateCompleted,
		}
		err = s.usageReporter.RecordUsage(usageRecord)
		s.Require().NoError(err)

		// Step 7: Generate and settle invoice
		invoiceID := fmt.Sprintf("invoice-%s", job.JobID)
		invoice := &InvoiceE2E{
			InvoiceID:    invoiceID,
			ProviderAddr: s.providerAddr,
			CustomerAddr: s.customerAddr,
			JobID:        job.JobID,
			LineItems: []LineItemE2E{
				{ResourceType: "compute", Quantity: sdkmath.LegacyMustNewDecFromStr("1.0"), TotalCost: "10.0"},
			},
			TotalAmount: "10.0",
			Status:      "pending",
		}
		err = s.settlementMock.CreateInvoice(ctx, invoice)
		s.Require().NoError(err)

		err = s.settlementMock.TriggerSettlement(ctx, invoiceID)
		s.Require().NoError(err)

		// Step 8: Verify settlement complete
		settledInvoice := s.settlementMock.GetInvoice(invoiceID)
		s.Equal("settled", settledInvoice.Status)

		payout := s.settlementMock.GetProviderPayout(s.providerAddr, invoiceID)
		s.NotNil(payout)
		s.Equal("completed", payout.Status)
	})
}

// =============================================================================
// Mock Implementations for E2E Tests
// =============================================================================

// UsageRecordE2E represents a usage record for E2E testing.
type UsageRecordE2E struct {
	RecordID        string
	JobID           string
	ClusterID       string
	ProviderAddress string
	CustomerAddress string
	PeriodStart     time.Time
	PeriodEnd       time.Time
	Metrics         *pd.HPCSchedulerMetrics
	IsFinal         bool
	JobState        pd.HPCJobState
}

// MockUsageReporterE2E is a mock usage reporter for E2E testing.
type MockUsageReporterE2E struct {
	records map[string]*UsageRecordE2E
	reports []*pd.HPCStatusReport
}

// NewMockUsageReporterE2E creates a new mock usage reporter.
func NewMockUsageReporterE2E() *MockUsageReporterE2E {
	return &MockUsageReporterE2E{
		records: make(map[string]*UsageRecordE2E),
		reports: make([]*pd.HPCStatusReport, 0),
	}
}

func (m *MockUsageReporterE2E) RecordUsage(record *UsageRecordE2E) error {
	m.records[record.RecordID] = record
	return nil
}

func (m *MockUsageReporterE2E) GetRecordsForJob(jobID string) []*UsageRecordE2E {
	var result []*UsageRecordE2E
	for _, r := range m.records {
		if r.JobID == jobID {
			result = append(result, r)
		}
	}
	return result
}

func (m *MockUsageReporterE2E) SubmitStatusReport(report *pd.HPCStatusReport) error {
	m.reports = append(m.reports, report)
	return nil
}

func (m *MockUsageReporterE2E) GetSubmittedReports() []*pd.HPCStatusReport {
	return m.reports
}

// InvoiceE2E represents an invoice for E2E testing.
type InvoiceE2E struct {
	InvoiceID    string
	ProviderAddr string
	CustomerAddr string
	JobID        string
	LineItems    []LineItemE2E
	TotalAmount  string
	PeriodStart  time.Time
	PeriodEnd    time.Time
	Status       string
	SettledAt    *time.Time
}

// LineItemE2E represents a line item for E2E testing.
type LineItemE2E struct {
	ResourceType string
	Quantity     sdkmath.LegacyDec
	UnitPrice    string
	TotalCost    string
}

// PayoutE2E represents a payout for E2E testing.
type PayoutE2E struct {
	PayoutID  string
	InvoiceID string
	Provider  string
	Amount    string
	Status    string
}

// PlatformFeeE2E represents a platform fee for E2E testing.
type PlatformFeeE2E struct {
	FeeID     string
	InvoiceID string
	Amount    string
}

// AuditRecordE2E represents an audit record for E2E testing.
type AuditRecordE2E struct {
	RecordID  string
	InvoiceID string
	Action    string
	Timestamp time.Time
	Details   string
}

// MockSettlementE2E is a mock settlement pipeline for E2E testing.
type MockSettlementE2E struct {
	invoices     map[string]*InvoiceE2E
	payouts      map[string]*PayoutE2E
	fees         map[string]*PlatformFeeE2E
	auditRecords map[string][]*AuditRecordE2E
}

// NewMockSettlementE2E creates a new mock settlement pipeline.
func NewMockSettlementE2E() *MockSettlementE2E {
	return &MockSettlementE2E{
		invoices:     make(map[string]*InvoiceE2E),
		payouts:      make(map[string]*PayoutE2E),
		fees:         make(map[string]*PlatformFeeE2E),
		auditRecords: make(map[string][]*AuditRecordE2E),
	}
}

func (m *MockSettlementE2E) CreateInvoice(ctx context.Context, invoice *InvoiceE2E) error {
	m.invoices[invoice.InvoiceID] = invoice
	m.addAuditRecord(invoice.InvoiceID, "created", "Invoice created")
	return nil
}

func (m *MockSettlementE2E) GetInvoice(invoiceID string) *InvoiceE2E {
	return m.invoices[invoiceID]
}

func (m *MockSettlementE2E) TriggerSettlement(ctx context.Context, invoiceID string) error {
	invoice, ok := m.invoices[invoiceID]
	if !ok {
		return fmt.Errorf("invoice not found: %s", invoiceID)
	}

	if invoice.Status == "disputed" {
		return fmt.Errorf("cannot settle disputed invoice")
	}

	now := time.Now()
	invoice.Status = "settled"
	invoice.SettledAt = &now

	// Create payout (97.5% to provider)
	m.payouts[invoiceID] = &PayoutE2E{
		PayoutID:  fmt.Sprintf("payout-%s", invoiceID),
		InvoiceID: invoiceID,
		Provider:  invoice.ProviderAddr,
		Amount:    invoice.TotalAmount,
		Status:    "completed",
	}

	// Create platform fee (2.5%)
	m.fees[invoiceID] = &PlatformFeeE2E{
		FeeID:     fmt.Sprintf("fee-%s", invoiceID),
		InvoiceID: invoiceID,
		Amount:    "0.025",
	}

	m.addAuditRecord(invoiceID, "settled", "Invoice settled and payout completed")

	return nil
}

func (m *MockSettlementE2E) DisputeInvoice(ctx context.Context, invoiceID, reason string) error {
	invoice, ok := m.invoices[invoiceID]
	if !ok {
		return fmt.Errorf("invoice not found: %s", invoiceID)
	}

	invoice.Status = "disputed"
	m.addAuditRecord(invoiceID, "disputed", fmt.Sprintf("Invoice disputed: %s", reason))
	return nil
}

func (m *MockSettlementE2E) GetProviderPayout(providerAddr, invoiceID string) *PayoutE2E {
	payout := m.payouts[invoiceID]
	if payout != nil && payout.Provider == providerAddr {
		return payout
	}
	return nil
}

func (m *MockSettlementE2E) GetPlatformFee(invoiceID string) *PlatformFeeE2E {
	return m.fees[invoiceID]
}

func (m *MockSettlementE2E) GetAuditTrail(invoiceID string) []*AuditRecordE2E {
	return m.auditRecords[invoiceID]
}

func (m *MockSettlementE2E) addAuditRecord(invoiceID, action, details string) {
	record := &AuditRecordE2E{
		RecordID:  fmt.Sprintf("audit-%s-%d", invoiceID, time.Now().UnixNano()),
		InvoiceID: invoiceID,
		Action:    action,
		Timestamp: time.Now(),
		Details:   details,
	}
	m.auditRecords[invoiceID] = append(m.auditRecords[invoiceID], record)
}

// MockAuditLoggerE2E is a mock audit logger for E2E testing.
type MockAuditLoggerE2E struct {
	events []pd.HPCAuditEvent
}

// NewMockAuditLoggerE2E creates a new mock audit logger.
func NewMockAuditLoggerE2E() *MockAuditLoggerE2E {
	return &MockAuditLoggerE2E{
		events: make([]pd.HPCAuditEvent, 0),
	}
}

func (m *MockAuditLoggerE2E) LogEvent(event pd.HPCAuditEvent) {
	m.events = append(m.events, event)
}

func (m *MockAuditLoggerE2E) GetEvents() []pd.HPCAuditEvent {
	return m.events
}
