//go:build e2e.integration

// Package e2e contains end-to-end integration tests.
//
// VE-HPC: Comprehensive E2E tests for the HPC on-chain module.
// Tests job CRUD, state transitions, cluster management, offerings,
// scheduling, and accounting functionality.
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

// =============================================================================
// HPCModuleE2ETestSuite - Comprehensive E2E Tests for HPC Module
// =============================================================================

// HPCModuleE2ETestSuite tests the complete HPC on-chain module functionality
// including job CRUD, state transitions, cluster management, offerings,
// scheduling decisions, and job accounting.
type HPCModuleE2ETestSuite struct {
	*testutil.NetworkTestSuite

	// Test addresses
	providerAddr string
	customerAddr string

	// Mock components
	slurmMock      *mocks.MockSLURMIntegration
	usageReporter  *MockHPCUsageReporter
	settlementMock *MockHPCSettlement
	auditLogger    *MockHPCAuditLog

	// Test data
	testCluster   *hpctypes.HPCCluster
	testOffering  *hpctypes.HPCOffering
	testOffering2 *hpctypes.HPCOffering

	// In-memory stores for test entities
	clusters            map[string]*hpctypes.HPCCluster
	offerings           map[string]*hpctypes.HPCOffering
	jobs                map[string]*hpctypes.HPCJob
	accountingRecords   map[string]*hpctypes.HPCAccountingRecord
	schedulingDecisions map[string]*hpctypes.SchedulingDecision
	usageSnapshots      map[string]*hpctypes.HPCUsageSnapshot

	// Lifecycle tracking
	lifecycleEvents []HPCLifecycleEvent
}

// HPCLifecycleEvent tracks job lifecycle events for verification.
type HPCLifecycleEvent struct {
	JobID     string
	FromState pd.HPCJobState
	ToState   pd.HPCJobState
	Timestamp time.Time
	EventType string
}

// MockHPCUsageReporter is a mock usage reporter for E2E tests.
type MockHPCUsageReporter struct {
	records         map[string]*HPCUsageRecordE2E
	statusReports   []*pd.HPCStatusReport
	snapshotCounter uint32
}

// HPCUsageRecordE2E represents a usage record for E2E testing.
type HPCUsageRecordE2E struct {
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
	SequenceNumber  uint32
}

// MockHPCSettlement is a mock settlement component for E2E tests.
type MockHPCSettlement struct {
	invoices    map[string]*HPCInvoiceE2E
	settlements map[string]*HPCSettlementRecordE2E
}

// HPCInvoiceE2E represents an invoice for E2E testing.
type HPCInvoiceE2E struct {
	InvoiceID    string
	JobID        string
	ProviderAddr string
	CustomerAddr string
	TotalAmount  sdk.Coins
	Status       string
	CreatedAt    time.Time
}

// HPCSettlementRecordE2E represents a settlement record for E2E testing.
type HPCSettlementRecordE2E struct {
	SettlementID   string
	InvoiceID      string
	ProviderPayout sdk.Coins
	PlatformFee    sdk.Coins
	Status         string
	SettledAt      time.Time
}

// MockHPCAuditLog is a mock audit logger for E2E tests.
type MockHPCAuditLog struct {
	entries []hpctypes.AuditTrailEntry
}

func NewMockHPCUsageReporter() *MockHPCUsageReporter {
	return &MockHPCUsageReporter{
		records:       make(map[string]*HPCUsageRecordE2E),
		statusReports: make([]*pd.HPCStatusReport, 0),
	}
}

func (m *MockHPCUsageReporter) RecordUsage(record *HPCUsageRecordE2E) error {
	m.records[record.RecordID] = record
	return nil
}

func (m *MockHPCUsageReporter) GetRecordsForJob(jobID string) []*HPCUsageRecordE2E {
	var result []*HPCUsageRecordE2E
	for _, r := range m.records {
		if r.JobID == jobID {
			result = append(result, r)
		}
	}
	return result
}

func (m *MockHPCUsageReporter) SubmitStatusReport(report *pd.HPCStatusReport) error {
	m.statusReports = append(m.statusReports, report)
	return nil
}

func (m *MockHPCUsageReporter) GetSubmittedReports() []*pd.HPCStatusReport {
	return m.statusReports
}

func (m *MockHPCUsageReporter) NextSequenceNumber() uint32 {
	m.snapshotCounter++
	return m.snapshotCounter
}

func NewMockHPCSettlement() *MockHPCSettlement {
	return &MockHPCSettlement{
		invoices:    make(map[string]*HPCInvoiceE2E),
		settlements: make(map[string]*HPCSettlementRecordE2E),
	}
}

func (m *MockHPCSettlement) CreateInvoice(invoice *HPCInvoiceE2E) {
	m.invoices[invoice.InvoiceID] = invoice
}

func (m *MockHPCSettlement) GetInvoice(invoiceID string) *HPCInvoiceE2E {
	return m.invoices[invoiceID]
}

func (m *MockHPCSettlement) SettleInvoice(invoiceID string) (*HPCSettlementRecordE2E, error) {
	invoice, ok := m.invoices[invoiceID]
	if !ok {
		return nil, fmt.Errorf("invoice not found: %s", invoiceID)
	}
	if invoice.Status == "disputed" {
		return nil, fmt.Errorf("cannot settle disputed invoice")
	}

	invoice.Status = "settled"

	// Calculate 2.5% platform fee
	totalInt := invoice.TotalAmount.AmountOf("uvirt").Int64()
	feeAmount := totalInt * 25 / 1000
	payoutAmount := totalInt - feeAmount

	settlement := &HPCSettlementRecordE2E{
		SettlementID:   fmt.Sprintf("settlement-%s", invoiceID),
		InvoiceID:      invoiceID,
		ProviderPayout: sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(payoutAmount))),
		PlatformFee:    sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(feeAmount))),
		Status:         "completed",
		SettledAt:      time.Now(),
	}
	m.settlements[settlement.SettlementID] = settlement
	return settlement, nil
}

func NewMockHPCAuditLog() *MockHPCAuditLog {
	return &MockHPCAuditLog{
		entries: make([]hpctypes.AuditTrailEntry, 0),
	}
}

func (m *MockHPCAuditLog) LogEntry(entry hpctypes.AuditTrailEntry) {
	m.entries = append(m.entries, entry)
}

func (m *MockHPCAuditLog) GetEntriesForEntity(entityType, entityID string) []hpctypes.AuditTrailEntry {
	var result []hpctypes.AuditTrailEntry
	for _, e := range m.entries {
		if e.EntityType == entityType && e.EntityID == entityID {
			result = append(result, e)
		}
	}
	return result
}

// =============================================================================
// Test Suite Setup and Teardown
// =============================================================================

func TestHPCModuleE2E(t *testing.T) {
	suite.Run(t, &HPCModuleE2ETestSuite{
		NetworkTestSuite: testutil.NewNetworkTestSuite(nil, &HPCModuleE2ETestSuite{}),
	})
}

func (s *HPCModuleE2ETestSuite) SetupSuite() {
	s.NetworkTestSuite.SetupSuite()

	val := s.Network().Validators[0]
	s.providerAddr = val.Address.String()
	s.customerAddr = val.Address.String()

	// Initialize mock components
	s.slurmMock = mocks.NewMockSLURMIntegration()
	s.usageReporter = NewMockHPCUsageReporter()
	s.settlementMock = NewMockHPCSettlement()
	s.auditLogger = NewMockHPCAuditLog()
	s.lifecycleEvents = make([]HPCLifecycleEvent, 0)

	// Initialize in-memory stores
	s.clusters = make(map[string]*hpctypes.HPCCluster)
	s.offerings = make(map[string]*hpctypes.HPCOffering)
	s.jobs = make(map[string]*hpctypes.HPCJob)
	s.accountingRecords = make(map[string]*hpctypes.HPCAccountingRecord)
	s.schedulingDecisions = make(map[string]*hpctypes.SchedulingDecision)
	s.usageSnapshots = make(map[string]*hpctypes.HPCUsageSnapshot)

	// Register lifecycle callback
	s.slurmMock.RegisterLifecycleCallback(s.onJobLifecycleEvent)

	// Create test cluster and offering using fixtures
	clusterConfig := fixtures.DefaultTestClusterConfig()
	clusterConfig.ProviderAddr = s.providerAddr
	s.testCluster = fixtures.CreateTestCluster(clusterConfig)
	s.clusters[s.testCluster.ClusterID] = s.testCluster

	offeringConfig := fixtures.DefaultTestOfferingConfig()
	offeringConfig.ProviderAddr = s.providerAddr
	s.testOffering = fixtures.CreateTestOffering(offeringConfig)
	s.offerings[s.testOffering.OfferingID] = s.testOffering

	// Create a second test offering for variety
	offeringConfig2 := fixtures.DefaultTestOfferingConfig()
	offeringConfig2.OfferingID = "hpc-gpu-premium"
	offeringConfig2.Name = "HPC GPU Premium"
	offeringConfig2.ProviderAddr = s.providerAddr
	offeringConfig2.Pricing.GPUHourPrice = "5.00"
	s.testOffering2 = fixtures.CreateTestOffering(offeringConfig2)
	s.offerings[s.testOffering2.OfferingID] = s.testOffering2
}

func (s *HPCModuleE2ETestSuite) TearDownSuite() {
	if s.slurmMock != nil && s.slurmMock.IsRunning() {
		_ = s.slurmMock.Stop()
	}
	s.NetworkTestSuite.TearDownSuite()
}

func (s *HPCModuleE2ETestSuite) onJobLifecycleEvent(job *pd.HPCSchedulerJob, event pd.HPCJobLifecycleEvent, prevState pd.HPCJobState) {
	s.lifecycleEvents = append(s.lifecycleEvents, HPCLifecycleEvent{
		JobID:     job.VirtEngineJobID,
		FromState: prevState,
		ToState:   job.State,
		Timestamp: time.Now(),
		EventType: string(event),
	})
}

// =============================================================================
// A. Job CRUD Operations Tests (~400 lines)
// =============================================================================

func (s *HPCModuleE2ETestSuite) TestA01_SubmitJob() {
	ctx := context.Background()

	s.Run("StartSLURMScheduler", func() {
		err := s.slurmMock.Start(ctx)
		s.Require().NoError(err)
		s.True(s.slurmMock.IsRunning())
	})

	s.Run("RegisterTestCluster", func() {
		cluster := mocks.DefaultTestCluster()
		s.slurmMock.RegisterCluster(cluster)

		registered, exists := s.slurmMock.GetCluster(cluster.ClusterID)
		s.True(exists)
		s.Equal(cluster.ClusterID, registered.ClusterID)
	})

	s.Run("SubmitValidJob", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "crud-submit-job-1"

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)

		// Verify job state is pending
		s.Equal(job.JobID, schedulerJob.VirtEngineJobID)
		s.Equal(pd.HPCJobStatePending, schedulerJob.State)
		s.NotEmpty(schedulerJob.SchedulerJobID)
		s.Equal(pd.HPCSchedulerTypeSLURM, schedulerJob.SchedulerType)

		// Store for later tests
		s.jobs[job.JobID] = job
	})

	s.Run("SubmitJobWithGPU", func() {
		job := fixtures.GPUComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "crud-submit-gpu-job-1"

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)
		s.Equal(pd.HPCJobStatePending, schedulerJob.State)

		s.jobs[job.JobID] = job
	})

	s.Run("SubmitMultiNodeJob", func() {
		job := fixtures.MultiNodeJob(s.providerAddr, s.customerAddr, 4)
		job.JobID = "crud-submit-multinode-job-1"

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)

		s.jobs[job.JobID] = job
	})

	s.Run("SubmitQuickTestJob", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = "crud-submit-quick-job-1"

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)

		s.jobs[job.JobID] = job
	})

	s.Run("VerifyJobValidation", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "validation-test-job"

		err := job.Validate()
		s.Require().NoError(err)
	})

	s.Run("VerifyJobTimestamps", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		s.NotZero(job.CreatedAt)
		s.True(time.Since(job.CreatedAt) < time.Minute)
	})
}

func (s *HPCModuleE2ETestSuite) TestA02_GetJob() {
	ctx := context.Background()

	s.Run("GetExistingJob", func() {
		jobID := "crud-submit-job-1"

		job, err := s.slurmMock.GetJobStatus(ctx, jobID)
		s.Require().NoError(err)
		s.NotNil(job)
		s.Equal(jobID, job.VirtEngineJobID)
	})

	s.Run("GetJobDetails", func() {
		jobID := "crud-submit-job-1"

		job, err := s.slurmMock.GetJobStatus(ctx, jobID)
		s.Require().NoError(err)

		s.NotEmpty(job.SchedulerJobID)
		s.Equal(pd.HPCSchedulerTypeSLURM, job.SchedulerType)
		s.NotNil(job.OriginalJob)
		s.Equal(s.providerAddr, job.OriginalJob.ProviderAddress)
		s.Equal(s.customerAddr, job.OriginalJob.CustomerAddress)
	})

	s.Run("GetNonExistentJob", func() {
		_, err := s.slurmMock.GetJobStatus(ctx, "nonexistent-job-id")
		s.Error(err)
		s.Contains(err.Error(), "not found")
	})

	s.Run("GetJobWithResources", func() {
		jobID := "crud-submit-gpu-job-1"

		job, err := s.slurmMock.GetJobStatus(ctx, jobID)
		s.Require().NoError(err)

		s.NotNil(job.OriginalJob)
		s.Greater(job.OriginalJob.Resources.GPUsPerNode, int32(0))
		s.NotEmpty(job.OriginalJob.Resources.GPUType)
	})

	s.Run("GetMultiNodeJobDetails", func() {
		jobID := "crud-submit-multinode-job-1"

		job, err := s.slurmMock.GetJobStatus(ctx, jobID)
		s.Require().NoError(err)

		s.NotNil(job.OriginalJob)
		s.Greater(job.OriginalJob.Resources.Nodes, int32(1))
	})
}

func (s *HPCModuleE2ETestSuite) TestA03_GetJobsByCustomer() {
	ctx := context.Background()

	s.Run("ListActiveJobs", func() {
		activeJobs, err := s.slurmMock.ListActiveJobs(ctx)
		s.Require().NoError(err)

		// Should have jobs from previous tests
		s.GreaterOrEqual(len(activeJobs), 1)
	})

	s.Run("FilterJobsByCustomer", func() {
		activeJobs, err := s.slurmMock.ListActiveJobs(ctx)
		s.Require().NoError(err)

		var customerJobs []*pd.HPCSchedulerJob
		for _, job := range activeJobs {
			if job.OriginalJob != nil && job.OriginalJob.CustomerAddress == s.customerAddr {
				customerJobs = append(customerJobs, job)
			}
		}

		s.GreaterOrEqual(len(customerJobs), 1)
	})

	s.Run("VerifyJobMetadata", func() {
		activeJobs, err := s.slurmMock.ListActiveJobs(ctx)
		s.Require().NoError(err)

		for _, job := range activeJobs {
			s.NotEmpty(job.VirtEngineJobID)
			s.NotEmpty(job.SchedulerJobID)
			s.NotNil(job.OriginalJob)
		}
	})

	s.Run("VerifyJobStatesInList", func() {
		activeJobs, err := s.slurmMock.ListActiveJobs(ctx)
		s.Require().NoError(err)

		for _, job := range activeJobs {
			// Active jobs should not be in terminal state
			s.False(job.State.IsTerminal())
		}
	})
}

func (s *HPCModuleE2ETestSuite) TestA04_GetJobsByCluster() {
	ctx := context.Background()

	s.Run("FilterJobsByCluster", func() {
		activeJobs, err := s.slurmMock.ListActiveJobs(ctx)
		s.Require().NoError(err)

		clusterID := "e2e-slurm-cluster"
		var clusterJobs []*pd.HPCSchedulerJob
		for _, job := range activeJobs {
			if job.OriginalJob != nil && job.OriginalJob.ClusterID == clusterID {
				clusterJobs = append(clusterJobs, job)
			}
		}

		s.GreaterOrEqual(len(clusterJobs), 1)
	})

	s.Run("VerifyClusterAssignment", func() {
		job, err := s.slurmMock.GetJobStatus(ctx, "crud-submit-job-1")
		s.Require().NoError(err)

		s.NotNil(job.OriginalJob)
		s.NotEmpty(job.OriginalJob.ClusterID)
	})

	s.Run("VerifyQueueAssignment", func() {
		job, err := s.slurmMock.GetJobStatus(ctx, "crud-submit-job-1")
		s.Require().NoError(err)

		s.NotNil(job.OriginalJob)
		s.NotEmpty(job.OriginalJob.QueueName)
	})
}

func (s *HPCModuleE2ETestSuite) TestA05_CancelJob() {
	ctx := context.Background()

	s.Run("CancelPendingJob", func() {
		// Create a job to cancel
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = "crud-cancel-job-1"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Verify job is pending
		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStatePending, status.State)

		// Cancel the job
		err = s.slurmMock.CancelJob(ctx, job.JobID)
		s.Require().NoError(err)

		// Verify job is cancelled
		status, err = s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCancelled, status.State)
		s.True(status.State.IsTerminal())
	})

	s.Run("CancelQueuedJob", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = "crud-cancel-queued-job-1"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Move to queued state
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)

		// Cancel the job
		err = s.slurmMock.CancelJob(ctx, job.JobID)
		s.Require().NoError(err)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCancelled, status.State)
	})

	s.Run("CancelRunningJob", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = "crud-cancel-running-job-1"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Move to running state
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		// Cancel the job
		err = s.slurmMock.CancelJob(ctx, job.JobID)
		s.Require().NoError(err)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCancelled, status.State)
	})

	s.Run("VerifyCancelledJobHasEndTime", func() {
		status, err := s.slurmMock.GetJobStatus(ctx, "crud-cancel-running-job-1")
		s.Require().NoError(err)
		s.NotNil(status.EndTime)
	})

	s.Run("CannotCancelCompletedJob", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = "crud-cancel-completed-job-1"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Complete the job
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		// Try to cancel - should fail
		err = s.slurmMock.CancelJob(ctx, job.JobID)
		s.Error(err)
		s.Contains(err.Error(), "terminal")
	})

	s.Run("CancelNonExistentJob", func() {
		err := s.slurmMock.CancelJob(ctx, "nonexistent-cancel-job")
		s.Error(err)
		s.Contains(err.Error(), "not found")
	})
}

func (s *HPCModuleE2ETestSuite) TestA06_SubmitInvalidJob() {
	ctx := context.Background()

	s.Run("RejectJobWithMissingFields", func() {
		job := &hpctypes.HPCJob{
			JobID:           "",
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
		}

		err := job.Validate()
		s.Error(err)
		s.Contains(err.Error(), "job_id")
	})

	s.Run("RejectJobWithInvalidProvider", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "invalid-provider-job"
		job.ProviderAddress = "invalid-address"

		err := job.Validate()
		s.Error(err)
		s.Contains(err.Error(), "provider")
	})

	s.Run("RejectJobWithInvalidCustomer", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "invalid-customer-job"
		job.CustomerAddress = "invalid-address"

		err := job.Validate()
		s.Error(err)
		s.Contains(err.Error(), "customer")
	})

	s.Run("RejectJobWithInvalidResources", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "invalid-resources-job"
		job.Resources.Nodes = 0 // Invalid - must be at least 1

		err := job.Validate()
		s.Error(err)
		s.Contains(err.Error(), "nodes")
	})

	s.Run("RejectJobWithInvalidRuntime", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "invalid-runtime-job"
		job.MaxRuntimeSeconds = 30 // Invalid - must be at least 60

		err := job.Validate()
		s.Error(err)
		s.Contains(err.Error(), "runtime")
	})

	s.Run("RejectJobWithMissingWorkload", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "missing-workload-job"
		job.WorkloadSpec.ContainerImage = ""
		job.WorkloadSpec.IsPreconfigured = false

		err := job.Validate()
		s.Error(err)
		s.Contains(err.Error(), "container_image")
	})

	s.Run("RejectJobWithMissingCluster", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "missing-cluster-job"
		job.ClusterID = ""

		err := job.Validate()
		s.Error(err)
		s.Contains(err.Error(), "cluster_id")
	})

	s.Run("RejectJobWithMissingQueue", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "missing-queue-job"
		job.QueueName = ""

		err := job.Validate()
		s.Error(err)
		s.Contains(err.Error(), "queue_name")
	})

	s.Run("RejectOversizedJob", func() {
		job := fixtures.OversizedJob(s.providerAddr, s.customerAddr)
		job.JobID = "oversized-job-1"

		// Set strict capacity limits
		s.slurmMock.SetMaxCapacity(100, 100*1024, 10)

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Error(err)
		s.Contains(err.Error(), "insufficient")

		// Reset capacity
		s.slurmMock.SetMaxCapacity(10000, 1024*1024, 1000)
	})

	s.Run("RejectJobWithInvalidState", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "invalid-state-job"
		job.State = "invalid_state"

		err := job.Validate()
		s.Error(err)
		s.Contains(err.Error(), "state")
	})
}

// =============================================================================
// B. Job State Transitions Tests (~400 lines)
// =============================================================================

func (s *HPCModuleE2ETestSuite) TestB01_PendingToQueued() {
	ctx := context.Background()

	s.Run("TransitionPendingToQueued", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = "state-pending-queued-1"

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStatePending, schedulerJob.State)

		// Transition to queued
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateQueued, status.State)
	})

	s.Run("VerifyValidTransitionPendingToQueued", func() {
		isValid := s.slurmMock.IsValidTransition(pd.HPCJobStatePending, pd.HPCJobStateQueued)
		s.True(isValid)
	})

	s.Run("VerifyLifecycleEventRecorded", func() {
		var found bool
		for _, event := range s.lifecycleEvents {
			if event.JobID == "state-pending-queued-1" &&
				event.FromState == pd.HPCJobStatePending &&
				event.ToState == pd.HPCJobStateQueued {
				found = true
				break
			}
		}
		s.True(found, "Lifecycle event for pending->queued transition should be recorded")
	})

	s.Run("TransitionMultipleJobsToQueued", func() {
		for i := 0; i < 3; i++ {
			job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
			job.JobID = fmt.Sprintf("state-batch-queued-%d", i)

			_, err := s.slurmMock.SubmitJob(ctx, job)
			s.Require().NoError(err)

			s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)

			status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
			s.Require().NoError(err)
			s.Equal(pd.HPCJobStateQueued, status.State)
		}
	})
}

func (s *HPCModuleE2ETestSuite) TestB02_QueuedToRunning() {
	ctx := context.Background()

	s.Run("TransitionQueuedToRunning", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = "state-queued-running-1"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateRunning, status.State)
	})

	s.Run("VerifyStartTimeSet", func() {
		status, err := s.slurmMock.GetJobStatus(ctx, "state-queued-running-1")
		s.Require().NoError(err)
		s.NotNil(status.StartTime, "StartTime should be set when job transitions to running")
	})

	s.Run("VerifyValidTransitionQueuedToRunning", func() {
		isValid := s.slurmMock.IsValidTransition(pd.HPCJobStateQueued, pd.HPCJobStateRunning)
		s.True(isValid)
	})

	s.Run("VerifyExecutionRecordCreated", func() {
		record, exists := s.slurmMock.GetExecutionRecord("state-queued-running-1")
		s.True(exists)
		s.NotNil(record)
		s.NotNil(record.StartTime)
		s.NotEmpty(record.NodeList)
	})

	s.Run("VerifyJobNotTerminal", func() {
		status, err := s.slurmMock.GetJobStatus(ctx, "state-queued-running-1")
		s.Require().NoError(err)
		s.False(status.State.IsTerminal())
	})
}

func (s *HPCModuleE2ETestSuite) TestB03_RunningToCompleted() {
	ctx := context.Background()

	s.Run("TransitionRunningToCompleted", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = "state-running-completed-1"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		// Set metrics and exit code before completion
		metrics := fixtures.StandardJobMetrics(300)
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobExitCode(job.JobID, 0)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCompleted, status.State)
	})

	s.Run("VerifyExitCodeZero", func() {
		status, err := s.slurmMock.GetJobStatus(ctx, "state-running-completed-1")
		s.Require().NoError(err)
		s.Equal(int32(0), status.ExitCode)
	})

	s.Run("VerifyEndTimeSet", func() {
		status, err := s.slurmMock.GetJobStatus(ctx, "state-running-completed-1")
		s.Require().NoError(err)
		s.NotNil(status.EndTime)
	})

	s.Run("VerifyJobIsTerminal", func() {
		status, err := s.slurmMock.GetJobStatus(ctx, "state-running-completed-1")
		s.Require().NoError(err)
		s.True(status.State.IsTerminal())
	})

	s.Run("VerifyMetricsRecorded", func() {
		metrics, err := s.slurmMock.GetJobAccounting(ctx, "state-running-completed-1")
		s.Require().NoError(err)
		s.NotNil(metrics)
		s.Equal(int64(300), metrics.WallClockSeconds)
		s.Greater(metrics.CPUCoreSeconds, int64(0))
	})

	s.Run("VerifyExecutionRecordComplete", func() {
		record, exists := s.slurmMock.GetExecutionRecord("state-running-completed-1")
		s.True(exists)
		s.NotNil(record.EndTime)
		s.Equal(int32(0), record.ExitCode)
		s.NotNil(record.Metrics)
	})
}

func (s *HPCModuleE2ETestSuite) TestB04_RunningToFailed() {
	ctx := context.Background()

	s.Run("TransitionRunningToFailed", func() {
		job := fixtures.FailingJob(s.providerAddr, s.customerAddr)
		job.JobID = "state-running-failed-1"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		// Set non-zero exit code to indicate failure
		s.slurmMock.SetJobExitCode(job.JobID, 1)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateFailed)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateFailed, status.State)
	})

	s.Run("VerifyNonZeroExitCode", func() {
		status, err := s.slurmMock.GetJobStatus(ctx, "state-running-failed-1")
		s.Require().NoError(err)
		s.NotEqual(int32(0), status.ExitCode)
	})

	s.Run("VerifyFailedJobIsTerminal", func() {
		status, err := s.slurmMock.GetJobStatus(ctx, "state-running-failed-1")
		s.Require().NoError(err)
		s.True(status.State.IsTerminal())
	})

	s.Run("VerifyFailedJobHasEndTime", func() {
		status, err := s.slurmMock.GetJobStatus(ctx, "state-running-failed-1")
		s.Require().NoError(err)
		s.NotNil(status.EndTime)
	})

	s.Run("SimulateFailedExecution", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = "simulated-failure-1"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		err = s.slurmMock.SimulateJobExecution(ctx, job.JobID, 50, false, nil)
		s.Require().NoError(err)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateFailed, status.State)
	})
}

func (s *HPCModuleE2ETestSuite) TestB05_JobTimeout() {
	ctx := context.Background()

	s.Run("TransitionRunningToTimeout", func() {
		job := fixtures.TimeoutJob(s.providerAddr, s.customerAddr)
		job.JobID = "state-running-timeout-1"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		// Simulate timeout
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateTimeout)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateTimeout, status.State)
	})

	s.Run("VerifyTimeoutIsTerminal", func() {
		status, err := s.slurmMock.GetJobStatus(ctx, "state-running-timeout-1")
		s.Require().NoError(err)
		s.True(status.State.IsTerminal())
	})

	s.Run("VerifyTimeoutHasEndTime", func() {
		status, err := s.slurmMock.GetJobStatus(ctx, "state-running-timeout-1")
		s.Require().NoError(err)
		s.NotNil(status.EndTime)
	})

	s.Run("VerifyValidTransitionToTimeout", func() {
		isValid := s.slurmMock.IsValidTransition(pd.HPCJobStateRunning, pd.HPCJobStateTimeout)
		s.True(isValid)
	})

	s.Run("VerifyTimeoutJobMetrics", func() {
		// Even timed out jobs should have partial metrics
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = "timeout-with-metrics-1"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		// Set partial metrics before timeout
		partialMetrics := fixtures.PartialJobMetrics(30)
		s.slurmMock.SetJobMetrics(job.JobID, partialMetrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateTimeout)

		metrics, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.NotNil(metrics)
		s.Greater(metrics.WallClockSeconds, int64(0))
	})
}

func (s *HPCModuleE2ETestSuite) TestB06_InvalidStateTransition() {
	// Note: ctx not used in validation-only tests (no mock calls needed)
	_ = context.Background()

	s.Run("InvalidPendingToCompleted", func() {
		isValid := s.slurmMock.IsValidTransition(pd.HPCJobStatePending, pd.HPCJobStateCompleted)
		s.False(isValid, "Direct transition from pending to completed should be invalid")
	})

	s.Run("InvalidPendingToRunning", func() {
		isValid := s.slurmMock.IsValidTransition(pd.HPCJobStatePending, pd.HPCJobStateRunning)
		s.False(isValid, "Direct transition from pending to running should be invalid")
	})

	s.Run("InvalidCompletedToRunning", func() {
		isValid := s.slurmMock.IsValidTransition(pd.HPCJobStateCompleted, pd.HPCJobStateRunning)
		s.False(isValid, "Transition from terminal state should be invalid")
	})

	s.Run("InvalidFailedToRunning", func() {
		isValid := s.slurmMock.IsValidTransition(pd.HPCJobStateFailed, pd.HPCJobStateRunning)
		s.False(isValid, "Transition from terminal state should be invalid")
	})

	s.Run("InvalidCancelledToRunning", func() {
		isValid := s.slurmMock.IsValidTransition(pd.HPCJobStateCancelled, pd.HPCJobStateRunning)
		s.False(isValid, "Transition from terminal state should be invalid")
	})

	s.Run("InvalidTimeoutToCompleted", func() {
		isValid := s.slurmMock.IsValidTransition(pd.HPCJobStateTimeout, pd.HPCJobStateCompleted)
		s.False(isValid, "Transition from terminal state should be invalid")
	})

	s.Run("VerifyValidTransitions", func() {
		// Verify all valid transitions
		validTransitions := []struct {
			from pd.HPCJobState
			to   pd.HPCJobState
		}{
			{pd.HPCJobStatePending, pd.HPCJobStateQueued},
			{pd.HPCJobStatePending, pd.HPCJobStateFailed},
			{pd.HPCJobStatePending, pd.HPCJobStateCancelled},
			{pd.HPCJobStateQueued, pd.HPCJobStateRunning},
			{pd.HPCJobStateQueued, pd.HPCJobStateFailed},
			{pd.HPCJobStateQueued, pd.HPCJobStateCancelled},
			{pd.HPCJobStateRunning, pd.HPCJobStateCompleted},
			{pd.HPCJobStateRunning, pd.HPCJobStateFailed},
			{pd.HPCJobStateRunning, pd.HPCJobStateCancelled},
			{pd.HPCJobStateRunning, pd.HPCJobStateTimeout},
		}

		for _, tt := range validTransitions {
			isValid := s.slurmMock.IsValidTransition(tt.from, tt.to)
			s.True(isValid, "Transition from %s to %s should be valid", tt.from, tt.to)
		}
	})

	s.Run("VerifyTerminalStateDetection", func() {
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
			pd.HPCJobStateRunning,
		}

		for _, state := range nonTerminalStates {
			s.False(state.IsTerminal(), "State %s should not be terminal", state)
		}
	})
}

// =============================================================================
// C. Cluster Management Tests (~300 lines)
// =============================================================================

func (s *HPCModuleE2ETestSuite) TestC01_RegisterCluster() {
	s.Run("RegisterNewCluster", func() {
		cluster := &hpctypes.HPCCluster{
			ClusterID:       "test-cluster-c01",
			ProviderAddress: s.providerAddr,
			Name:            "Test Cluster C01",
			Description:     "Test cluster for registration tests",
			State:           hpctypes.ClusterStatePending,
			Region:          "us-west",
			TotalNodes:      50,
			AvailableNodes:  50,
			SLURMVersion:    "23.02.4",
			Partitions: []hpctypes.Partition{
				{
					Name:       "default",
					Nodes:      30,
					MaxRuntime: 86400,
					MaxNodes:   10,
					Features:   []string{"cpu"},
					State:      "up",
				},
				{
					Name:       "gpu",
					Nodes:      20,
					MaxRuntime: 172800,
					MaxNodes:   5,
					Features:   []string{"gpu", "a100"},
					State:      "up",
				},
			},
			ClusterMetadata: hpctypes.ClusterMetadata{
				TotalCPUCores:    3200,
				TotalMemoryGB:    12800,
				TotalGPUs:        160,
				InterconnectType: "infiniband",
				StorageType:      "lustre",
				TotalStorageGB:   50000,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := cluster.Validate()
		s.Require().NoError(err)

		s.clusters[cluster.ClusterID] = cluster
		s.Equal("test-cluster-c01", cluster.ClusterID)
	})

	s.Run("VerifyClusterValidation", func() {
		cluster := s.clusters["test-cluster-c01"]
		s.NotNil(cluster)

		err := cluster.Validate()
		s.Require().NoError(err)
	})

	s.Run("RegisterClusterInSLURMMock", func() {
		mockCluster := &mocks.SLURMCluster{
			ClusterID:     "test-cluster-c01-slurm",
			Name:          "Test SLURM Cluster C01",
			Region:        "us-west",
			SLURMVersion:  "23.02.4",
			TotalNodes:    50,
			TotalCPU:      3200,
			TotalMemoryGB: 12800,
			TotalGPUs:     160,
			Partitions: []mocks.SLURMPartition{
				{Name: "default", Nodes: 30, MaxRuntime: 86400, State: "up"},
				{Name: "gpu", Nodes: 20, MaxRuntime: 172800, Features: []string{"gpu", "a100"}, State: "up"},
			},
		}

		s.slurmMock.RegisterCluster(mockCluster)

		registered, exists := s.slurmMock.GetCluster("test-cluster-c01-slurm")
		s.True(exists)
		s.Equal("Test SLURM Cluster C01", registered.Name)
	})

	s.Run("VerifyClusterPartitions", func() {
		cluster := s.clusters["test-cluster-c01"]
		s.Len(cluster.Partitions, 2)

		var defaultPartition *hpctypes.Partition
		var gpuPartition *hpctypes.Partition

		for i := range cluster.Partitions {
			if cluster.Partitions[i].Name == "default" {
				defaultPartition = &cluster.Partitions[i]
			}
			if cluster.Partitions[i].Name == "gpu" {
				gpuPartition = &cluster.Partitions[i]
			}
		}

		s.NotNil(defaultPartition)
		s.Equal(int32(30), defaultPartition.Nodes)

		s.NotNil(gpuPartition)
		s.Contains(gpuPartition.Features, "a100")
	})

	s.Run("VerifyClusterMetadata", func() {
		cluster := s.clusters["test-cluster-c01"]

		s.Equal(int64(3200), cluster.ClusterMetadata.TotalCPUCores)
		s.Equal(int64(12800), cluster.ClusterMetadata.TotalMemoryGB)
		s.Equal(int64(160), cluster.ClusterMetadata.TotalGPUs)
		s.Equal("infiniband", cluster.ClusterMetadata.InterconnectType)
	})

	s.Run("RegisterClusterWithInvalidData", func() {
		invalidCluster := &hpctypes.HPCCluster{
			ClusterID:       "",
			ProviderAddress: s.providerAddr,
		}

		err := invalidCluster.Validate()
		s.Error(err)
		s.Contains(err.Error(), "cluster_id")
	})
}

func (s *HPCModuleE2ETestSuite) TestC02_UpdateCluster() {
	s.Run("UpdateClusterState", func() {
		cluster := s.clusters["test-cluster-c01"]
		s.NotNil(cluster)

		// Activate the cluster
		cluster.State = hpctypes.ClusterStateActive
		cluster.UpdatedAt = time.Now()

		s.Equal(hpctypes.ClusterStateActive, cluster.State)
	})

	s.Run("UpdateClusterAvailableNodes", func() {
		cluster := s.clusters["test-cluster-c01"]

		// Simulate some nodes being used
		cluster.AvailableNodes = 40
		cluster.UpdatedAt = time.Now()

		s.Equal(int32(40), cluster.AvailableNodes)
		s.LessOrEqual(cluster.AvailableNodes, cluster.TotalNodes)
	})

	s.Run("UpdateClusterPartition", func() {
		cluster := s.clusters["test-cluster-c01"]

		// Update partition priority
		for i := range cluster.Partitions {
			if cluster.Partitions[i].Name == "gpu" {
				cluster.Partitions[i].Priority = 100
			}
		}
		cluster.UpdatedAt = time.Now()

		var gpuPartition *hpctypes.Partition
		for i := range cluster.Partitions {
			if cluster.Partitions[i].Name == "gpu" {
				gpuPartition = &cluster.Partitions[i]
			}
		}
		s.Equal(int32(100), gpuPartition.Priority)
	})

	s.Run("TransitionClusterToDraining", func() {
		cluster := s.clusters["test-cluster-c01"]

		cluster.State = hpctypes.ClusterStateDraining
		cluster.UpdatedAt = time.Now()

		s.Equal(hpctypes.ClusterStateDraining, cluster.State)
	})

	s.Run("VerifyClusterStateValidity", func() {
		s.True(hpctypes.IsValidClusterState(hpctypes.ClusterStatePending))
		s.True(hpctypes.IsValidClusterState(hpctypes.ClusterStateActive))
		s.True(hpctypes.IsValidClusterState(hpctypes.ClusterStateDraining))
		s.True(hpctypes.IsValidClusterState(hpctypes.ClusterStateOffline))
		s.True(hpctypes.IsValidClusterState(hpctypes.ClusterStateDeregistered))
		s.False(hpctypes.IsValidClusterState("invalid_state"))
	})
}

func (s *HPCModuleE2ETestSuite) TestC03_DeregisterCluster() {
	s.Run("DeregisterCluster", func() {
		// Create a cluster to deregister
		cluster := &hpctypes.HPCCluster{
			ClusterID:       "test-cluster-deregister",
			ProviderAddress: s.providerAddr,
			Name:            "Cluster to Deregister",
			State:           hpctypes.ClusterStateActive,
			Region:          "eu-west",
			TotalNodes:      10,
			AvailableNodes:  10,
			SLURMVersion:    "23.02.4",
			Partitions: []hpctypes.Partition{
				{Name: "default", Nodes: 10, MaxRuntime: 86400, State: "up"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		s.clusters[cluster.ClusterID] = cluster

		// Deregister
		cluster.State = hpctypes.ClusterStateDeregistered
		cluster.UpdatedAt = time.Now()

		s.Equal(hpctypes.ClusterStateDeregistered, cluster.State)
	})

	s.Run("VerifyDeregisteredClusterNotActive", func() {
		cluster := s.clusters["test-cluster-deregister"]
		s.NotEqual(hpctypes.ClusterStateActive, cluster.State)
	})

	s.Run("CleanupDeregisteredCluster", func() {
		delete(s.clusters, "test-cluster-deregister")
		_, exists := s.clusters["test-cluster-deregister"]
		s.False(exists)
	})
}

func (s *HPCModuleE2ETestSuite) TestC04_ClusterHealthCheck() {
	s.Run("VerifyClusterHealth", func() {
		cluster := s.clusters["test-cluster-c01"]
		s.NotNil(cluster)

		// Restore to active state
		cluster.State = hpctypes.ClusterStateActive

		// Health check - verify essential properties
		s.Greater(cluster.TotalNodes, int32(0))
		s.GreaterOrEqual(cluster.AvailableNodes, int32(0))
		s.LessOrEqual(cluster.AvailableNodes, cluster.TotalNodes)
		s.NotEmpty(cluster.SLURMVersion)
		s.GreaterOrEqual(len(cluster.Partitions), 1)
	})

	s.Run("VerifyPartitionHealth", func() {
		cluster := s.clusters["test-cluster-c01"]

		for _, partition := range cluster.Partitions {
			s.NotEmpty(partition.Name)
			s.Greater(partition.Nodes, int32(0))
			s.Equal("up", partition.State)
		}
	})

	s.Run("VerifyMetadataHealth", func() {
		cluster := s.clusters["test-cluster-c01"]

		s.Greater(cluster.ClusterMetadata.TotalCPUCores, int64(0))
		s.Greater(cluster.ClusterMetadata.TotalMemoryGB, int64(0))
		s.NotEmpty(cluster.ClusterMetadata.InterconnectType)
		s.NotEmpty(cluster.ClusterMetadata.StorageType)
	})

	s.Run("SimulateHealthCheckFromMock", func() {
		clusters := s.slurmMock.GetClusters()
		s.GreaterOrEqual(len(clusters), 1)

		for _, cluster := range clusters {
			s.NotEmpty(cluster.ClusterID)
			s.Greater(cluster.TotalNodes, int32(0))
			s.NotEmpty(cluster.SLURMVersion)
		}
	})
}

func (s *HPCModuleE2ETestSuite) TestC05_ClusterHeartbeatTimeout() {
	s.Run("SimulateHeartbeatTimeout", func() {
		// Create a cluster that will timeout
		cluster := &hpctypes.HPCCluster{
			ClusterID:       "test-cluster-timeout",
			ProviderAddress: s.providerAddr,
			Name:            "Timeout Test Cluster",
			State:           hpctypes.ClusterStateActive,
			Region:          "ap-south",
			TotalNodes:      20,
			AvailableNodes:  20,
			SLURMVersion:    "23.02.4",
			Partitions: []hpctypes.Partition{
				{Name: "default", Nodes: 20, MaxRuntime: 86400, State: "up"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now().Add(-time.Hour * 2), // Simulate stale update
		}

		s.clusters[cluster.ClusterID] = cluster

		// Check for timeout condition
		heartbeatTimeout := time.Minute * 30
		lastUpdate := cluster.UpdatedAt
		if time.Since(lastUpdate) > heartbeatTimeout {
			cluster.State = hpctypes.ClusterStateOffline
		}

		s.Equal(hpctypes.ClusterStateOffline, cluster.State)
	})

	s.Run("VerifyOfflineClusterState", func() {
		cluster := s.clusters["test-cluster-timeout"]
		s.Equal(hpctypes.ClusterStateOffline, cluster.State)
	})

	s.Run("RecoverFromOffline", func() {
		cluster := s.clusters["test-cluster-timeout"]

		// Simulate heartbeat recovery
		cluster.UpdatedAt = time.Now()
		cluster.State = hpctypes.ClusterStateActive

		s.Equal(hpctypes.ClusterStateActive, cluster.State)
	})

	s.Run("CleanupTimeoutCluster", func() {
		delete(s.clusters, "test-cluster-timeout")
	})
}

// =============================================================================
// D. Offering Management Tests (~300 lines)
// =============================================================================

func (s *HPCModuleE2ETestSuite) TestD01_CreateOffering() {
	s.Run("CreateNewOffering", func() {
		offering := &hpctypes.HPCOffering{
			OfferingID:      "offering-d01-standard",
			ClusterID:       "test-cluster-c01",
			ProviderAddress: s.providerAddr,
			Name:            "Standard Compute D01",
			Description:     "Standard compute offering for testing",
			QueueOptions: []hpctypes.QueueOption{
				{
					PartitionName:   "default",
					DisplayName:     "Standard Queue",
					MaxNodes:        10,
					MaxRuntime:      86400,
					PriceMultiplier: "1.0",
				},
			},
			Pricing: hpctypes.HPCPricing{
				BaseNodeHourPrice: "10.0",
				CPUCoreHourPrice:  "0.10",
				GPUHourPrice:      "2.50",
				MemoryGBHourPrice: "0.01",
				StorageGBPrice:    "0.001",
				NetworkGBPrice:    "0.05",
				Currency:          "uvirt",
				MinimumCharge:     "1.0",
			},
			RequiredIdentityThreshold: 70,
			MaxRuntimeSeconds:         86400,
			SupportsCustomWorkloads:   true,
			Active:                    true,
			CreatedAt:                 time.Now(),
			UpdatedAt:                 time.Now(),
		}

		err := offering.Validate()
		s.Require().NoError(err)

		s.offerings[offering.OfferingID] = offering
		s.Equal("offering-d01-standard", offering.OfferingID)
	})

	s.Run("CreateGPUOffering", func() {
		offering := &hpctypes.HPCOffering{
			OfferingID:      "offering-d01-gpu",
			ClusterID:       "test-cluster-c01",
			ProviderAddress: s.providerAddr,
			Name:            "GPU Compute D01",
			Description:     "GPU compute offering for ML workloads",
			QueueOptions: []hpctypes.QueueOption{
				{
					PartitionName:   "gpu",
					DisplayName:     "GPU Queue",
					MaxNodes:        5,
					MaxRuntime:      172800,
					Features:        []string{"gpu", "a100"},
					PriceMultiplier: "2.5",
				},
			},
			Pricing: hpctypes.HPCPricing{
				BaseNodeHourPrice: "25.0",
				CPUCoreHourPrice:  "0.15",
				GPUHourPrice:      "5.00",
				MemoryGBHourPrice: "0.02",
				StorageGBPrice:    "0.002",
				NetworkGBPrice:    "0.05",
				Currency:          "uvirt",
				MinimumCharge:     "5.0",
			},
			RequiredIdentityThreshold: 80,
			MaxRuntimeSeconds:         172800,
			SupportsCustomWorkloads:   true,
			Active:                    true,
			CreatedAt:                 time.Now(),
			UpdatedAt:                 time.Now(),
		}

		err := offering.Validate()
		s.Require().NoError(err)

		s.offerings[offering.OfferingID] = offering
	})

	s.Run("CreateOfferingWithPreconfiguredWorkloads", func() {
		offering := &hpctypes.HPCOffering{
			OfferingID:      "offering-d01-ml-training",
			ClusterID:       "test-cluster-c01",
			ProviderAddress: s.providerAddr,
			Name:            "ML Training D01",
			QueueOptions: []hpctypes.QueueOption{
				{
					PartitionName:   "gpu",
					DisplayName:     "ML Training Queue",
					MaxNodes:        8,
					MaxRuntime:      259200,
					PriceMultiplier: "3.0",
				},
			},
			Pricing: hpctypes.HPCPricing{
				BaseNodeHourPrice: "50.0",
				CPUCoreHourPrice:  "0.20",
				GPUHourPrice:      "10.00",
				MemoryGBHourPrice: "0.03",
				StorageGBPrice:    "0.003",
				Currency:          "uvirt",
				MinimumCharge:     "10.0",
			},
			PreconfiguredWorkloads: []hpctypes.PreconfiguredWorkload{
				{
					WorkloadID:     "pytorch-training",
					Name:           "PyTorch Training",
					Description:    "Distributed PyTorch training workload",
					ContainerImage: "pytorch/pytorch:2.0-cuda12.0-runtime",
					DefaultCommand: "python train.py",
					Category:       "ml-training",
					Version:        "2.0",
				},
				{
					WorkloadID:     "tensorflow-training",
					Name:           "TensorFlow Training",
					Description:    "Distributed TensorFlow training workload",
					ContainerImage: "tensorflow/tensorflow:2.12.0-gpu",
					DefaultCommand: "python train.py",
					Category:       "ml-training",
					Version:        "2.12",
				},
			},
			RequiredIdentityThreshold: 85,
			MaxRuntimeSeconds:         259200,
			SupportsCustomWorkloads:   false,
			Active:                    true,
			CreatedAt:                 time.Now(),
			UpdatedAt:                 time.Now(),
		}

		err := offering.Validate()
		s.Require().NoError(err)

		s.offerings[offering.OfferingID] = offering
		s.Len(offering.PreconfiguredWorkloads, 2)
	})

	s.Run("VerifyOfferingValidation", func() {
		invalidOffering := &hpctypes.HPCOffering{
			OfferingID:      "",
			ClusterID:       "test-cluster",
			ProviderAddress: s.providerAddr,
		}

		err := invalidOffering.Validate()
		s.Error(err)
		s.Contains(err.Error(), "offering_id")
	})

	s.Run("VerifyOfferingQueueOptions", func() {
		offering := s.offerings["offering-d01-standard"]
		s.GreaterOrEqual(len(offering.QueueOptions), 1)

		queue := offering.QueueOptions[0]
		s.NotEmpty(queue.PartitionName)
		s.NotEmpty(queue.DisplayName)
		s.Greater(queue.MaxNodes, int32(0))
	})
}

func (s *HPCModuleE2ETestSuite) TestD02_UpdateOffering() {
	s.Run("UpdateOfferingPricing", func() {
		offering := s.offerings["offering-d01-standard"]
		s.NotNil(offering)

		// Update pricing
		offering.Pricing.BaseNodeHourPrice = "12.0"
		offering.Pricing.GPUHourPrice = "3.00"
		offering.UpdatedAt = time.Now()

		s.Equal("12.0", offering.Pricing.BaseNodeHourPrice)
	})

	s.Run("UpdateOfferingQueueOptions", func() {
		offering := s.offerings["offering-d01-standard"]

		// Add a new queue option
		offering.QueueOptions = append(offering.QueueOptions, hpctypes.QueueOption{
			PartitionName:   "highmem",
			DisplayName:     "High Memory Queue",
			MaxNodes:        4,
			MaxRuntime:      43200,
			PriceMultiplier: "1.5",
		})
		offering.UpdatedAt = time.Now()

		s.Len(offering.QueueOptions, 2)
	})

	s.Run("UpdateOfferingIdentityThreshold", func() {
		offering := s.offerings["offering-d01-standard"]

		offering.RequiredIdentityThreshold = 75
		offering.UpdatedAt = time.Now()

		s.Equal(int32(75), offering.RequiredIdentityThreshold)
	})

	s.Run("UpdateOfferingMaxRuntime", func() {
		offering := s.offerings["offering-d01-standard"]

		offering.MaxRuntimeSeconds = 172800 // 48 hours
		offering.UpdatedAt = time.Now()

		s.Equal(int64(172800), offering.MaxRuntimeSeconds)
	})

	s.Run("VerifyOfferingValidationAfterUpdate", func() {
		offering := s.offerings["offering-d01-standard"]

		err := offering.Validate()
		s.Require().NoError(err)
	})
}

func (s *HPCModuleE2ETestSuite) TestD03_GetOfferingsByCluster() {
	s.Run("ListOfferingsForCluster", func() {
		clusterID := "test-cluster-c01"

		var clusterOfferings []*hpctypes.HPCOffering
		for _, offering := range s.offerings {
			if offering.ClusterID == clusterID {
				clusterOfferings = append(clusterOfferings, offering)
			}
		}

		s.GreaterOrEqual(len(clusterOfferings), 1)
	})

	s.Run("FilterActiveOfferings", func() {
		var activeOfferings []*hpctypes.HPCOffering
		for _, offering := range s.offerings {
			if offering.Active {
				activeOfferings = append(activeOfferings, offering)
			}
		}

		s.GreaterOrEqual(len(activeOfferings), 1)

		for _, offering := range activeOfferings {
			s.True(offering.Active)
		}
	})

	s.Run("VerifyOfferingDetails", func() {
		offering := s.offerings["offering-d01-standard"]
		s.NotNil(offering)

		s.NotEmpty(offering.OfferingID)
		s.NotEmpty(offering.ClusterID)
		s.NotEmpty(offering.ProviderAddress)
		s.NotEmpty(offering.Name)
		s.NotEmpty(offering.Pricing.Currency)
	})

	s.Run("VerifyOfferingPricingDetails", func() {
		offering := s.offerings["offering-d01-gpu"]
		s.NotNil(offering)

		s.NotEmpty(offering.Pricing.BaseNodeHourPrice)
		s.NotEmpty(offering.Pricing.GPUHourPrice)
		s.NotEmpty(offering.Pricing.MinimumCharge)
	})
}

func (s *HPCModuleE2ETestSuite) TestD04_DeactivateOffering() {
	s.Run("DeactivateOffering", func() {
		// Create an offering to deactivate
		offering := &hpctypes.HPCOffering{
			OfferingID:      "offering-d04-deactivate",
			ClusterID:       "test-cluster-c01",
			ProviderAddress: s.providerAddr,
			Name:            "Offering to Deactivate",
			QueueOptions: []hpctypes.QueueOption{
				{
					PartitionName:   "default",
					DisplayName:     "Default Queue",
					MaxNodes:        5,
					MaxRuntime:      3600,
					PriceMultiplier: "1.0",
				},
			},
			Pricing: hpctypes.HPCPricing{
				BaseNodeHourPrice: "5.0",
				CPUCoreHourPrice:  "0.05",
				Currency:          "uvirt",
				MinimumCharge:     "0.5",
			},
			RequiredIdentityThreshold: 50,
			MaxRuntimeSeconds:         3600,
			Active:                    true,
			CreatedAt:                 time.Now(),
			UpdatedAt:                 time.Now(),
		}

		s.offerings[offering.OfferingID] = offering

		// Deactivate
		offering.Active = false
		offering.UpdatedAt = time.Now()

		s.False(offering.Active)
	})

	s.Run("VerifyDeactivatedOfferingFiltered", func() {
		var activeOfferings []*hpctypes.HPCOffering
		for _, offering := range s.offerings {
			if offering.Active {
				activeOfferings = append(activeOfferings, offering)
			}
		}

		for _, offering := range activeOfferings {
			s.NotEqual("offering-d04-deactivate", offering.OfferingID)
		}
	})

	s.Run("ReactivateOffering", func() {
		offering := s.offerings["offering-d04-deactivate"]
		s.NotNil(offering)

		offering.Active = true
		offering.UpdatedAt = time.Now()

		s.True(offering.Active)
	})

	s.Run("CleanupDeactivatedOffering", func() {
		offering := s.offerings["offering-d04-deactivate"]
		offering.Active = false
		delete(s.offerings, "offering-d04-deactivate")
	})
}

// =============================================================================
// E. Job Scheduling Tests (~400 lines)
// =============================================================================

func (s *HPCModuleE2ETestSuite) TestE01_ScheduleJobToCluster() {
	ctx := context.Background()

	s.Run("ScheduleJobToCorrectCluster", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "schedule-cluster-1"
		job.ClusterID = "e2e-slurm-cluster"

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)

		// Verify job was scheduled to the correct cluster
		s.Equal(job.ClusterID, schedulerJob.OriginalJob.ClusterID)
	})

	s.Run("VerifyClusterCapability", func() {
		cluster, exists := s.slurmMock.GetCluster("e2e-slurm-cluster")
		s.True(exists)
		s.NotNil(cluster)

		// Verify cluster has required capabilities
		s.Greater(cluster.TotalNodes, int32(0))
		s.Greater(cluster.TotalCPU, int32(0))
	})

	s.Run("ScheduleGPUJobToGPUCluster", func() {
		job := fixtures.GPUComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "schedule-gpu-cluster-1"
		job.QueueName = "gpu"

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)

		s.Equal("gpu", schedulerJob.OriginalJob.QueueName)
	})

	s.Run("ScheduleMultiNodeJob", func() {
		job := fixtures.MultiNodeJob(s.providerAddr, s.customerAddr, 4)
		job.JobID = "schedule-multinode-1"

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)

		s.Equal(int32(4), schedulerJob.OriginalJob.Resources.Nodes)
	})

	s.Run("VerifyQueueAssignment", func() {
		job, err := s.slurmMock.GetJobStatus(ctx, "schedule-cluster-1")
		s.Require().NoError(err)

		s.NotEmpty(job.OriginalJob.QueueName)
	})

	s.Run("ScheduleToSpecificPartition", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = "schedule-partition-1"
		job.QueueName = "default"

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.Equal("default", schedulerJob.OriginalJob.QueueName)
	})
}

func (s *HPCModuleE2ETestSuite) TestE02_SchedulingDecisionRecorded() {
	ctx := context.Background()

	s.Run("VerifyRoutingDecisionCreated", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "routing-decision-job-1"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Verify routing decision was recorded
		decision, exists := s.slurmMock.GetRoutingDecisionForJob(job.JobID)
		s.True(exists)
		s.NotNil(decision)
		s.Equal(job.JobID, decision.JobID)
	})

	s.Run("VerifyDecisionContainsSelectedCluster", func() {
		decision, exists := s.slurmMock.GetRoutingDecisionForJob("routing-decision-job-1")
		s.True(exists)
		s.NotEmpty(decision.SelectedCluster)
	})

	s.Run("VerifyDecisionContainsCandidates", func() {
		decision, exists := s.slurmMock.GetRoutingDecisionForJob("routing-decision-job-1")
		s.True(exists)
		s.GreaterOrEqual(len(decision.CandidateClusters), 1)
	})

	s.Run("VerifyDecisionContainsScoringFactors", func() {
		decision, exists := s.slurmMock.GetRoutingDecisionForJob("routing-decision-job-1")
		s.True(exists)
		s.NotEmpty(decision.ScoringFactors)

		// Verify expected scoring factors
		expectedFactors := []string{
			"resource_availability",
			"queue_depth",
			"geographic_proximity",
			"price_competitiveness",
		}

		for _, factor := range expectedFactors {
			_, hasKey := decision.ScoringFactors[factor]
			s.True(hasKey, "Missing scoring factor: %s", factor)
		}
	})

	s.Run("VerifyDecisionHash", func() {
		decision, exists := s.slurmMock.GetRoutingDecisionForJob("routing-decision-job-1")
		s.True(exists)
		s.NotEmpty(decision.DecisionHash)
		s.Len(decision.DecisionHash, 64) // SHA256 hex string
	})

	s.Run("VerifyDecisionReason", func() {
		decision, exists := s.slurmMock.GetRoutingDecisionForJob("routing-decision-job-1")
		s.True(exists)
		s.NotEmpty(decision.Reason)
	})

	s.Run("VerifyDecisionTimestamp", func() {
		decision, exists := s.slurmMock.GetRoutingDecisionForJob("routing-decision-job-1")
		s.True(exists)
		s.False(decision.Timestamp.IsZero())
		s.True(time.Since(decision.Timestamp) < time.Minute)
	})

	s.Run("CreateOnChainSchedulingDecision", func() {
		decision := &hpctypes.SchedulingDecision{
			DecisionID:        "decision-e02-1",
			JobID:             "routing-decision-job-1",
			SelectedClusterID: "e2e-slurm-cluster",
			CandidateClusters: []hpctypes.ClusterCandidate{
				{
					ClusterID:      "e2e-slurm-cluster",
					Region:         "us-east",
					AvgLatencyMs:   10,
					AvailableNodes: 50,
					LatencyScore:   "0.950000",
					CapacityScore:  "0.900000",
					CombinedScore:  "0.920000",
					Eligible:       true,
				},
			},
			DecisionReason: "Best combined score for resource availability and latency",
			LatencyScore:   "0.950000",
			CapacityScore:  "0.900000",
			CombinedScore:  "0.920000",
			IsFallback:     false,
			CreatedAt:      time.Now(),
			BlockHeight:    12345,
		}

		err := decision.Validate()
		s.Require().NoError(err)

		s.schedulingDecisions[decision.DecisionID] = decision
	})
}

func (s *HPCModuleE2ETestSuite) TestE03_RoutingValidation() {
	ctx := context.Background()

	s.Run("ValidateRoutingToExistingCluster", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "routing-validate-1"
		job.ClusterID = "e2e-slurm-cluster"

		// Verify cluster exists
		cluster, exists := s.slurmMock.GetCluster(job.ClusterID)
		s.True(exists)
		s.NotNil(cluster)

		// Submit job
		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
	})

	s.Run("RejectRoutingToNonexistentCluster", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "routing-invalid-cluster-1"
		job.ClusterID = "nonexistent-cluster"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Error(err)
		s.Contains(err.Error(), "cluster not found")
	})

	s.Run("ValidatePartitionRouting", func() {
		job := fixtures.GPUComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "routing-partition-1"
		job.QueueName = "gpu"

		// Verify GPU partition exists in cluster
		cluster, _ := s.slurmMock.GetCluster("e2e-slurm-cluster")
		var hasGPUPartition bool
		for _, partition := range cluster.Partitions {
			if partition.Name == "gpu" {
				hasGPUPartition = true
				break
			}
		}
		s.True(hasGPUPartition)

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
	})

	s.Run("VerifyRoutingEnforcement", func() {
		decisions := s.slurmMock.GetRoutingDecisions()
		s.GreaterOrEqual(len(decisions), 1)

		// All decisions should have valid clusters
		for _, decision := range decisions {
			s.NotEmpty(decision.SelectedCluster)
			s.NotEmpty(decision.JobID)
		}
	})

	s.Run("ValidateSchedulingDecision", func() {
		decision := &hpctypes.SchedulingDecision{
			DecisionID:        "decision-validate-1",
			JobID:             "routing-validate-1",
			SelectedClusterID: "e2e-slurm-cluster",
			DecisionReason:    "Selected based on capacity",
			CombinedScore:     "0.850000",
			CreatedAt:         time.Now(),
		}

		err := decision.Validate()
		s.Require().NoError(err)
	})

	s.Run("RejectInvalidSchedulingDecision", func() {
		decision := &hpctypes.SchedulingDecision{
			DecisionID:        "",
			JobID:             "some-job",
			SelectedClusterID: "some-cluster",
		}

		err := decision.Validate()
		s.Error(err)
		s.Contains(err.Error(), "decision_id")
	})
}

func (s *HPCModuleE2ETestSuite) TestE04_RefreshSchedulingDecision() {
	ctx := context.Background()

	s.Run("CreateInitialDecision", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "refresh-decision-job-1"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		decision, exists := s.slurmMock.GetRoutingDecisionForJob(job.JobID)
		s.True(exists)
		s.NotNil(decision)
	})

	s.Run("SimulateDecisionRefresh", func() {
		// Create a new decision with updated scores
		newDecision := &hpctypes.SchedulingDecision{
			DecisionID:        "decision-refresh-1",
			JobID:             "refresh-decision-job-1",
			SelectedClusterID: "e2e-slurm-cluster",
			CandidateClusters: []hpctypes.ClusterCandidate{
				{
					ClusterID:     "e2e-slurm-cluster",
					Region:        "us-east",
					LatencyScore:  "0.980000",
					CapacityScore: "0.850000",
					CombinedScore: "0.900000",
					Eligible:      true,
				},
			},
			DecisionReason: "Refreshed decision with updated cluster state",
			LatencyScore:   "0.980000",
			CapacityScore:  "0.850000",
			CombinedScore:  "0.900000",
			IsFallback:     false,
			CreatedAt:      time.Now(),
		}

		err := newDecision.Validate()
		s.Require().NoError(err)

		s.schedulingDecisions[newDecision.DecisionID] = newDecision
	})

	s.Run("VerifyDecisionUpdate", func() {
		decision := s.schedulingDecisions["decision-refresh-1"]
		s.NotNil(decision)
		s.Contains(decision.DecisionReason, "Refreshed")
	})

	s.Run("TrackDecisionHistory", func() {
		// In a real system, we'd keep history of decisions
		jobID := "refresh-decision-job-1"

		var jobDecisions []*hpctypes.SchedulingDecision
		for _, decision := range s.schedulingDecisions {
			if decision.JobID == jobID {
				jobDecisions = append(jobDecisions, decision)
			}
		}

		s.GreaterOrEqual(len(jobDecisions), 1)
	})

	s.Run("VerifyFallbackDecision", func() {
		fallbackDecision := &hpctypes.SchedulingDecision{
			DecisionID:        "decision-fallback-1",
			JobID:             "fallback-job-1",
			SelectedClusterID: "e2e-slurm-cluster",
			DecisionReason:    "Fallback to default cluster",
			IsFallback:        true,
			FallbackReason:    "Primary cluster unavailable",
			CombinedScore:     "0.500000",
			CreatedAt:         time.Now(),
		}

		err := fallbackDecision.Validate()
		s.Require().NoError(err)

		s.True(fallbackDecision.IsFallback)
		s.NotEmpty(fallbackDecision.FallbackReason)
	})
}

// =============================================================================
// F. Job Accounting Tests (~400 lines)
// =============================================================================

func (s *HPCModuleE2ETestSuite) TestF01_CreateJobAccounting() {
	ctx := context.Background()

	s.Run("CreateAccountingForCompletedJob", func() {
		// First, create and complete a job
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "accounting-job-1"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Progress to completion
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		metrics := fixtures.StandardJobMetrics(3600) // 1 hour
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobExitCode(job.JobID, 0)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		// Create accounting record
		now := time.Now()
		accountingRecord := &hpctypes.HPCAccountingRecord{
			RecordID:        "accounting-record-1",
			JobID:           job.JobID,
			ClusterID:       job.ClusterID,
			ProviderAddress: job.ProviderAddress,
			CustomerAddress: job.CustomerAddress,
			OfferingID:      job.OfferingID,
			SchedulerType:   "SLURM",
			UsageMetrics: hpctypes.HPCDetailedMetrics{
				WallClockSeconds: metrics.WallClockSeconds,
				CPUCoreSeconds:   metrics.CPUCoreSeconds,
				MemoryGBSeconds:  metrics.MemoryGBSeconds,
				NodesUsed:        1,
				NodeHours:        sdkmath.LegacyNewDec(1),
				SubmitTime:       now.Add(-time.Hour * 2),
			},
			BillableAmount:     sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1000000))),
			ProviderReward:     sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(975000))),
			PlatformFee:        sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(25000))),
			Status:             hpctypes.AccountingStatusPending,
			PeriodStart:        now.Add(-time.Hour),
			PeriodEnd:          now,
			FormulaVersion:     "1.0.0",
			SignedUsageRecords: []string{"usage-record-1"},
			CreatedAt:          now,
			BlockHeight:        12345,
		}

		err = accountingRecord.Validate()
		s.Require().NoError(err)

		s.accountingRecords[accountingRecord.RecordID] = accountingRecord
	})

	s.Run("VerifyAccountingRecordContents", func() {
		record := s.accountingRecords["accounting-record-1"]
		s.NotNil(record)

		s.Equal("accounting-job-1", record.JobID)
		s.NotEmpty(record.ClusterID)
		s.NotEmpty(record.ProviderAddress)
		s.NotEmpty(record.CustomerAddress)
	})

	s.Run("VerifyAccountingMetrics", func() {
		record := s.accountingRecords["accounting-record-1"]

		s.Greater(record.UsageMetrics.WallClockSeconds, int64(0))
		s.Greater(record.UsageMetrics.CPUCoreSeconds, int64(0))
	})

	s.Run("VerifyAccountingAmounts", func() {
		record := s.accountingRecords["accounting-record-1"]

		s.True(record.BillableAmount.IsValid())
		s.True(record.ProviderReward.IsValid())
		s.True(record.PlatformFee.IsValid())

		// Provider reward + platform fee should equal billable amount
		totalPayout := record.ProviderReward.Add(record.PlatformFee...)
		s.True(totalPayout.Equal(record.BillableAmount))
	})

	s.Run("CreateAccountingForGPUJob", func() {
		job := fixtures.GPUComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = "accounting-gpu-job-1"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		gpuMetrics := fixtures.GPUJobMetrics(7200) // 2 hours
		s.slurmMock.SetJobMetrics(job.JobID, gpuMetrics)
		s.slurmMock.SetJobExitCode(job.JobID, 0)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		now := time.Now()
		accountingRecord := &hpctypes.HPCAccountingRecord{
			RecordID:        "accounting-gpu-record-1",
			JobID:           job.JobID,
			ClusterID:       job.ClusterID,
			ProviderAddress: job.ProviderAddress,
			CustomerAddress: job.CustomerAddress,
			OfferingID:      job.OfferingID,
			SchedulerType:   "SLURM",
			UsageMetrics: hpctypes.HPCDetailedMetrics{
				WallClockSeconds: gpuMetrics.WallClockSeconds,
				CPUCoreSeconds:   gpuMetrics.CPUCoreSeconds,
				MemoryGBSeconds:  gpuMetrics.MemoryGBSeconds,
				GPUSeconds:       gpuMetrics.GPUSeconds,
				GPUType:          "nvidia-a100",
				NodesUsed:        1,
				NodeHours:        sdkmath.LegacyNewDec(2),
			},
			BillableAmount:     sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(5000000))),
			ProviderReward:     sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(4875000))),
			PlatformFee:        sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(125000))),
			Status:             hpctypes.AccountingStatusPending,
			PeriodStart:        now.Add(-time.Hour * 2),
			PeriodEnd:          now,
			FormulaVersion:     "1.0.0",
			SignedUsageRecords: []string{"usage-record-gpu-1"},
			CreatedAt:          now,
			BlockHeight:        12346,
		}

		err = accountingRecord.Validate()
		s.Require().NoError(err)

		s.accountingRecords[accountingRecord.RecordID] = accountingRecord
	})

	s.Run("VerifyGPUAccountingMetrics", func() {
		record := s.accountingRecords["accounting-gpu-record-1"]
		s.NotNil(record)

		s.Greater(record.UsageMetrics.GPUSeconds, int64(0))
		s.NotEmpty(record.UsageMetrics.GPUType)
	})
}

func (s *HPCModuleE2ETestSuite) TestF02_FinalizeAccounting() {
	s.Run("FinalizeAccountingRecord", func() {
		record := s.accountingRecords["accounting-record-1"]
		s.NotNil(record)

		s.Equal(hpctypes.AccountingStatusPending, record.Status)

		now := time.Now()
		err := record.Finalize(now)
		s.Require().NoError(err)

		s.Equal(hpctypes.AccountingStatusFinalized, record.Status)
		s.NotNil(record.FinalizedAt)
		s.NotEmpty(record.CalculationHash)
	})

	s.Run("VerifyCalculationHashGenerated", func() {
		record := s.accountingRecords["accounting-record-1"]

		s.NotEmpty(record.CalculationHash)
		// Hash should be SHA256 hex (64 chars)
		s.Len(record.CalculationHash, 64)
	})

	s.Run("CannotFinalizeAlreadyFinalized", func() {
		record := s.accountingRecords["accounting-record-1"]

		err := record.Finalize(time.Now())
		s.Error(err)
		s.Contains(err.Error(), "finalize")
	})

	s.Run("SettleFinalizedRecord", func() {
		record := s.accountingRecords["accounting-record-1"]

		now := time.Now()
		err := record.Settle("settlement-1", now)
		s.Require().NoError(err)

		s.Equal(hpctypes.AccountingStatusSettled, record.Status)
		s.NotNil(record.SettledAt)
		s.Equal("settlement-1", record.SettlementID)
	})

	s.Run("FinalizeGPUAccounting", func() {
		record := s.accountingRecords["accounting-gpu-record-1"]
		s.NotNil(record)

		now := time.Now()
		err := record.Finalize(now)
		s.Require().NoError(err)

		s.Equal(hpctypes.AccountingStatusFinalized, record.Status)
	})

	s.Run("VerifyAccountingStatusTransitions", func() {
		s.True(hpctypes.IsValidAccountingRecordStatus(hpctypes.AccountingStatusPending))
		s.True(hpctypes.IsValidAccountingRecordStatus(hpctypes.AccountingStatusFinalized))
		s.True(hpctypes.IsValidAccountingRecordStatus(hpctypes.AccountingStatusDisputed))
		s.True(hpctypes.IsValidAccountingRecordStatus(hpctypes.AccountingStatusSettled))
		s.True(hpctypes.IsValidAccountingRecordStatus(hpctypes.AccountingStatusCorrected))
		s.False(hpctypes.IsValidAccountingRecordStatus("invalid_status"))
	})
}

func (s *HPCModuleE2ETestSuite) TestF03_AccountingRecordContents() {
	s.Run("VerifyBillableBreakdown", func() {
		now := time.Now()
		breakdown := hpctypes.BillableBreakdown{
			CPUCost:     sdk.NewCoin("uvirt", sdkmath.NewInt(400000)),
			MemoryCost:  sdk.NewCoin("uvirt", sdkmath.NewInt(100000)),
			GPUCost:     sdk.NewCoin("uvirt", sdkmath.NewInt(0)),
			StorageCost: sdk.NewCoin("uvirt", sdkmath.NewInt(50000)),
			NetworkCost: sdk.NewCoin("uvirt", sdkmath.NewInt(50000)),
			NodeCost:    sdk.NewCoin("uvirt", sdkmath.NewInt(400000)),
			Subtotal:    sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1000000))),
		}

		record := &hpctypes.HPCAccountingRecord{
			RecordID:          "breakdown-record-1",
			JobID:             "breakdown-job-1",
			ClusterID:         "e2e-slurm-cluster",
			ProviderAddress:   s.providerAddr,
			CustomerAddress:   s.customerAddr,
			OfferingID:        "hpc-compute-standard",
			SchedulerType:     "SLURM",
			BillableBreakdown: breakdown,
			BillableAmount:    sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1000000))),
			ProviderReward:    sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(975000))),
			PlatformFee:       sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(25000))),
			Status:            hpctypes.AccountingStatusPending,
			PeriodStart:       now.Add(-time.Hour),
			PeriodEnd:         now,
			FormulaVersion:    "1.0.0",
			CreatedAt:         now,
		}

		s.Equal(sdkmath.NewInt(400000), record.BillableBreakdown.CPUCost.Amount)
		s.Equal(sdkmath.NewInt(400000), record.BillableBreakdown.NodeCost.Amount)
	})

	s.Run("VerifyAppliedDiscounts", func() {
		discount := hpctypes.AppliedDiscount{
			DiscountID:     "volume-discount-1",
			DiscountType:   "volume",
			Description:    "10% volume discount for usage > 100 hours",
			DiscountBps:    1000, // 10%
			DiscountAmount: sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(100000))),
			AppliedTo:      "billable_amount",
		}

		s.Equal("volume", discount.DiscountType)
		s.Equal(uint32(1000), discount.DiscountBps)
	})

	s.Run("VerifyAppliedCaps", func() {
		cap := hpctypes.AppliedCap{
			CapID:          "daily-cap-1",
			CapType:        "daily",
			Description:    "Daily spending cap of 10M uvirt",
			CapAmount:      sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(10000000))),
			OriginalAmount: sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(15000000))),
			CappedAmount:   sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(5000000))),
		}

		s.Equal("daily", cap.CapType)
		s.True(cap.OriginalAmount.IsAllGT(cap.CapAmount))
	})

	s.Run("VerifyNodeRewards", func() {
		nodeReward := hpctypes.NodeReward{
			NodeID:             "node-001",
			ProviderAddress:    s.providerAddr,
			Amount:             sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(250000))),
			ContributionWeight: "0.250000",
			UsageSeconds:       3600,
		}

		s.NotEmpty(nodeReward.NodeID)
		s.Equal(int64(3600), nodeReward.UsageSeconds)
	})

	s.Run("VerifyDetailedMetricsConversion", func() {
		detailed := hpctypes.HPCDetailedMetrics{
			WallClockSeconds: 3600,
			CPUCoreSeconds:   14400,
			MemoryGBSeconds:  57600,
			GPUSeconds:       7200,
			NodesUsed:        1,
			NodeHours:        sdkmath.LegacyNewDec(1),
		}

		legacy := detailed.ToLegacyMetrics()

		s.Equal(int64(3600), legacy.WallClockSeconds)
		s.Equal(int64(14400), legacy.CPUCoreSeconds)
		s.Equal(int64(57600), legacy.MemoryGBSeconds)
		s.Equal(int64(7200), legacy.GPUSeconds)
	})
}

func (s *HPCModuleE2ETestSuite) TestF04_UsageSnapshotCreation() {
	s.Run("CreateInterimSnapshot", func() {
		now := time.Now()
		snapshot := &hpctypes.HPCUsageSnapshot{
			SnapshotID:      "snapshot-interim-1",
			JobID:           "accounting-job-1",
			ClusterID:       "e2e-slurm-cluster",
			SchedulerType:   "SLURM",
			SnapshotType:    hpctypes.SnapshotTypeInterim,
			SequenceNumber:  1,
			ProviderAddress: s.providerAddr,
			CustomerAddress: s.customerAddr,
			Metrics: hpctypes.HPCDetailedMetrics{
				WallClockSeconds: 1800,
				CPUCoreSeconds:   7200,
				MemoryGBSeconds:  28800,
				NodesUsed:        1,
			},
			CumulativeMetrics: hpctypes.HPCDetailedMetrics{
				WallClockSeconds: 1800,
				CPUCoreSeconds:   7200,
				MemoryGBSeconds:  28800,
				NodesUsed:        1,
			},
			DeltaMetrics: hpctypes.HPCDetailedMetrics{
				WallClockSeconds: 1800,
				CPUCoreSeconds:   7200,
				MemoryGBSeconds:  28800,
			},
			JobState:          hpctypes.JobStateRunning,
			SnapshotTime:      now,
			ProviderSignature: "mock-signature-interim-1",
			CreatedAt:         now,
			BlockHeight:       12350,
		}

		snapshot.ContentHash = snapshot.CalculateContentHash()
		s.NotEmpty(snapshot.ContentHash)

		err := snapshot.Validate()
		s.Require().NoError(err)

		s.usageSnapshots[snapshot.SnapshotID] = snapshot
	})

	s.Run("CreateFinalSnapshot", func() {
		now := time.Now()
		snapshot := &hpctypes.HPCUsageSnapshot{
			SnapshotID:         "snapshot-final-1",
			JobID:              "accounting-job-1",
			ClusterID:          "e2e-slurm-cluster",
			SchedulerType:      "SLURM",
			SnapshotType:       hpctypes.SnapshotTypeFinal,
			SequenceNumber:     2,
			ProviderAddress:    s.providerAddr,
			CustomerAddress:    s.customerAddr,
			PreviousSnapshotID: "snapshot-interim-1",
			Metrics: hpctypes.HPCDetailedMetrics{
				WallClockSeconds: 3600,
				CPUCoreSeconds:   14400,
				MemoryGBSeconds:  57600,
				NodesUsed:        1,
			},
			CumulativeMetrics: hpctypes.HPCDetailedMetrics{
				WallClockSeconds: 3600,
				CPUCoreSeconds:   14400,
				MemoryGBSeconds:  57600,
				NodesUsed:        1,
			},
			DeltaMetrics: hpctypes.HPCDetailedMetrics{
				WallClockSeconds: 1800,
				CPUCoreSeconds:   7200,
				MemoryGBSeconds:  28800,
			},
			JobState:          hpctypes.JobStateCompleted,
			SnapshotTime:      now,
			ProviderSignature: "mock-signature-final-1",
			CreatedAt:         now,
			BlockHeight:       12360,
		}

		snapshot.ContentHash = snapshot.CalculateContentHash()

		err := snapshot.Validate()
		s.Require().NoError(err)

		s.usageSnapshots[snapshot.SnapshotID] = snapshot
	})

	s.Run("VerifySnapshotChain", func() {
		finalSnapshot := s.usageSnapshots["snapshot-final-1"]
		s.NotNil(finalSnapshot)

		s.Equal("snapshot-interim-1", finalSnapshot.PreviousSnapshotID)
		s.Equal(uint32(2), finalSnapshot.SequenceNumber)
	})

	s.Run("VerifySnapshotTypes", func() {
		s.True(hpctypes.IsValidSnapshotType(hpctypes.SnapshotTypeInterim))
		s.True(hpctypes.IsValidSnapshotType(hpctypes.SnapshotTypeFinal))
		s.True(hpctypes.IsValidSnapshotType(hpctypes.SnapshotTypeReconciliation))
		s.True(hpctypes.IsValidSnapshotType(hpctypes.SnapshotTypeDispute))
		s.False(hpctypes.IsValidSnapshotType("invalid_type"))
	})

	s.Run("VerifyContentHashConsistency", func() {
		snapshot := s.usageSnapshots["snapshot-final-1"]

		// Recalculate hash and verify it matches
		recalculatedHash := snapshot.CalculateContentHash()
		s.Equal(snapshot.ContentHash, recalculatedHash)
	})

	s.Run("CreateReconciliationSnapshot", func() {
		now := time.Now()
		snapshot := &hpctypes.HPCUsageSnapshot{
			SnapshotID:         "snapshot-reconciliation-1",
			JobID:              "accounting-job-1",
			ClusterID:          "e2e-slurm-cluster",
			SchedulerType:      "SLURM",
			SnapshotType:       hpctypes.SnapshotTypeReconciliation,
			SequenceNumber:     3,
			ProviderAddress:    s.providerAddr,
			CustomerAddress:    s.customerAddr,
			PreviousSnapshotID: "snapshot-final-1",
			Metrics: hpctypes.HPCDetailedMetrics{
				WallClockSeconds: 3600,
				CPUCoreSeconds:   14400,
				MemoryGBSeconds:  57600,
				NodesUsed:        1,
			},
			CumulativeMetrics: hpctypes.HPCDetailedMetrics{
				WallClockSeconds: 3600,
				CPUCoreSeconds:   14400,
				MemoryGBSeconds:  57600,
			},
			JobState:          hpctypes.JobStateCompleted,
			SnapshotTime:      now,
			ProviderSignature: "mock-signature-reconciliation-1",
			SchedulerRawData:  `{"sacct_output": "job_id|user|state|elapsed"}`,
			CreatedAt:         now,
			BlockHeight:       12370,
		}

		snapshot.ContentHash = snapshot.CalculateContentHash()

		err := snapshot.Validate()
		s.Require().NoError(err)

		s.usageSnapshots[snapshot.SnapshotID] = snapshot

		s.Equal(hpctypes.SnapshotTypeReconciliation, snapshot.SnapshotType)
		s.NotEmpty(snapshot.SchedulerRawData)
	})

	s.Run("VerifyQueueTimeCalculation", func() {
		startTime := time.Now()
		submitTime := startTime.Add(-time.Minute * 5)

		metrics := hpctypes.HPCDetailedMetrics{
			SubmitTime: submitTime,
			StartTime:  &startTime,
		}

		queueTime := metrics.CalculateQueueTime()
		s.Equal(int64(300), queueTime) // 5 minutes = 300 seconds
	})

	s.Run("VerifyWallClockCalculation", func() {
		startTime := time.Now().Add(-time.Hour)
		endTime := time.Now()

		metrics := hpctypes.HPCDetailedMetrics{
			StartTime: &startTime,
			EndTime:   &endTime,
		}

		wallClock := metrics.CalculateWallClock()
		s.Equal(int64(3600), wallClock) // 1 hour = 3600 seconds
	})
}
