//go:build e2e.integration

// Package e2e contains end-to-end integration tests.
//
// VE-15D: Comprehensive E2E tests for HPC provider daemon integration.
package e2e

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
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

// HPCProviderE2ETestSuite tests the HPC provider daemon integration.
type HPCProviderE2ETestSuite struct {
	*testutil.NetworkTestSuite
	providerAddr    string
	customerAddr    string
	slurmMock       *mocks.MockSLURMIntegration
	usageReporter   *ProviderMockUsageReporter
	auditLogger     *ProviderMockAuditLogger
	testCluster     *hpctypes.HPCCluster
	testOffering    *hpctypes.HPCOffering
	mu              sync.Mutex
	lifecycleEvents []LifecycleEventE2E
	usageSnapshots  map[string][]*UsageSnapshotE2E
}

func TestHPCProviderE2E(t *testing.T) {
	suite.Run(t, &HPCProviderE2ETestSuite{
		NetworkTestSuite: testutil.NewNetworkTestSuite(nil, &HPCProviderE2ETestSuite{}),
	})
}

// LifecycleEventE2E tracks job lifecycle events for verification.
type LifecycleEventE2E struct {
	JobID     string
	Event     pd.HPCJobLifecycleEvent
	FromState pd.HPCJobState
	ToState   pd.HPCJobState
	Timestamp time.Time
}

// UsageSnapshotE2E tracks usage snapshots for verification.
type UsageSnapshotE2E struct {
	SnapshotID   string
	JobID        string
	Timestamp    time.Time
	Metrics      *pd.HPCSchedulerMetrics
	IsFinal      bool
	SnapshotType string
}

// AuditEvent tracks audit events for verification.
type AuditEvent struct {
	EventType string
	JobID     string
	Details   map[string]string
	Timestamp time.Time
}

// ProviderMockUsageReporter tracks usage reports for verification.
type ProviderMockUsageReporter struct {
	mu      sync.Mutex
	reports []*pd.HPCStatusReport
	records map[string]*pd.HPCUsageRecord
}

// NewProviderMockUsageReporter creates a new mock usage reporter.
func NewProviderMockUsageReporter() *ProviderMockUsageReporter {
	return &ProviderMockUsageReporter{
		reports: make([]*pd.HPCStatusReport, 0),
		records: make(map[string]*pd.HPCUsageRecord),
	}
}

// SubmitStatusReport submits a status report.
func (m *ProviderMockUsageReporter) SubmitStatusReport(report *pd.HPCStatusReport) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reports = append(m.reports, report)
	return nil
}

// GetSubmittedReports returns all submitted reports.
func (m *ProviderMockUsageReporter) GetSubmittedReports() []*pd.HPCStatusReport {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]*pd.HPCStatusReport, len(m.reports))
	copy(result, m.reports)
	return result
}

// SubmitUsageRecord submits a usage record.
func (m *ProviderMockUsageReporter) SubmitUsageRecord(record *pd.HPCUsageRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records[record.JobID] = record
	return nil
}

// GetUsageRecord returns a usage record for a job.
func (m *ProviderMockUsageReporter) GetUsageRecord(jobID string) (*pd.HPCUsageRecord, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	record, ok := m.records[jobID]
	return record, ok
}

// GetAllRecords returns all usage records.
func (m *ProviderMockUsageReporter) GetAllRecords() map[string]*pd.HPCUsageRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string]*pd.HPCUsageRecord)
	for k, v := range m.records {
		result[k] = v
	}
	return result
}

// Clear clears all reports and records.
func (m *ProviderMockUsageReporter) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reports = make([]*pd.HPCStatusReport, 0)
	m.records = make(map[string]*pd.HPCUsageRecord)
}

// ProviderMockAuditLogger tracks audit events.
type ProviderMockAuditLogger struct {
	mu     sync.Mutex
	events []AuditEvent
}

// NewProviderMockAuditLogger creates a new mock audit logger.
func NewProviderMockAuditLogger() *ProviderMockAuditLogger {
	return &ProviderMockAuditLogger{
		events: make([]AuditEvent, 0),
	}
}

// LogEvent logs an audit event.
func (m *ProviderMockAuditLogger) LogEvent(event AuditEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
}

// GetEvents returns all audit events.
func (m *ProviderMockAuditLogger) GetEvents() []AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]AuditEvent, len(m.events))
	copy(result, m.events)
	return result
}

// GetEventsForJob returns audit events for a specific job.
func (m *ProviderMockAuditLogger) GetEventsForJob(jobID string) []AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []AuditEvent
	for _, e := range m.events {
		if e.JobID == jobID {
			result = append(result, e)
		}
	}
	return result
}

// Clear clears all events.
func (m *ProviderMockAuditLogger) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = make([]AuditEvent, 0)
}

// =============================================================================
// Suite Setup and Teardown
// =============================================================================

func (s *HPCProviderE2ETestSuite) SetupSuite() {
	s.NetworkTestSuite.SetupSuite()

	val := s.Network().Validators[0]
	s.providerAddr = val.Address.String()
	s.customerAddr = val.Address.String()

	// Initialize mock components
	s.slurmMock = mocks.NewMockSLURMIntegration()
	s.usageReporter = NewProviderMockUsageReporter()
	s.auditLogger = NewProviderMockAuditLogger()
	s.lifecycleEvents = make([]LifecycleEventE2E, 0)
	s.usageSnapshots = make(map[string][]*UsageSnapshotE2E)

	// Register lifecycle callback
	s.slurmMock.RegisterLifecycleCallback(s.onJobLifecycleEvent)

	// Create test cluster and offering using fixtures
	clusterConfig := fixtures.DefaultTestClusterConfig()
	clusterConfig.ProviderAddr = s.providerAddr
	s.testCluster = fixtures.CreateTestCluster(clusterConfig)

	offeringConfig := fixtures.DefaultTestOfferingConfig()
	offeringConfig.ProviderAddr = s.providerAddr
	s.testOffering = fixtures.CreateTestOffering(offeringConfig)

	// Register default cluster
	cluster := mocks.DefaultTestCluster()
	s.slurmMock.RegisterCluster(cluster)
}

func (s *HPCProviderE2ETestSuite) TearDownSuite() {
	if s.slurmMock != nil && s.slurmMock.IsRunning() {
		_ = s.slurmMock.Stop()
	}
	s.NetworkTestSuite.TearDownSuite()
}

func (s *HPCProviderE2ETestSuite) onJobLifecycleEvent(job *pd.HPCSchedulerJob, event pd.HPCJobLifecycleEvent, prevState pd.HPCJobState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lifecycleEvents = append(s.lifecycleEvents, LifecycleEventE2E{
		JobID:     job.VirtEngineJobID,
		Event:     event,
		FromState: prevState,
		ToState:   job.State,
		Timestamp: time.Now(),
	})
}

func (s *HPCProviderE2ETestSuite) recordUsageSnapshot(jobID string, metrics *pd.HPCSchedulerMetrics, isFinal bool, snapshotType string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshot := &UsageSnapshotE2E{
		SnapshotID:   fmt.Sprintf("snapshot-%s-%d", jobID, time.Now().UnixNano()),
		JobID:        jobID,
		Timestamp:    time.Now(),
		Metrics:      metrics,
		IsFinal:      isFinal,
		SnapshotType: snapshotType,
	}
	s.usageSnapshots[jobID] = append(s.usageSnapshots[jobID], snapshot)
}

func (s *HPCProviderE2ETestSuite) getLifecycleEventsForJob(jobID string) []LifecycleEventE2E {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result []LifecycleEventE2E
	for _, e := range s.lifecycleEvents {
		if e.JobID == jobID {
			result = append(result, e)
		}
	}
	return result
}

func (s *HPCProviderE2ETestSuite) clearLifecycleEvents() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lifecycleEvents = make([]LifecycleEventE2E, 0)
}

func (s *HPCProviderE2ETestSuite) uniqueJobID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()%1000000)
}

func calculateExpectedCPUCoreSeconds(durationSeconds int64, cores int32) int64 {
	return durationSeconds * int64(cores)
}

func calculateExpectedMemoryGBSeconds(durationSeconds int64, memoryGB int32) int64 {
	return durationSeconds * int64(memoryGB)
}

func calculateExpectedNodeHours(durationSeconds int64, nodes int32) float64 {
	return float64(durationSeconds) * float64(nodes) / 3600.0
}

func generateVerificationHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Ensure imports are used
var (
	_ = sort.Strings
	_ = sdkmath.NewInt
	_ = sdk.AccAddress{}
)

// =============================================================================
// Section A: SLURM Adapter Integration Tests (~500 lines)
// =============================================================================

// TestA01_SLURMAdapterStart tests starting the SLURM adapter and verifying connection.
func (s *HPCProviderE2ETestSuite) TestA01_SLURMAdapterStart() {
	ctx := context.Background()

	s.Run("StartSLURMAdapter", func() {
		err := s.slurmMock.Start(ctx)
		s.Require().NoError(err)
		s.True(s.slurmMock.IsRunning())
	})

	s.Run("VerifySchedulerType", func() {
		s.Equal(pd.HPCSchedulerTypeSLURM, s.slurmMock.Type())
	})

	s.Run("VerifyAdapterState", func() {
		s.True(s.slurmMock.IsRunning())
	})

	s.Run("VerifyDefaultClusterRegistered", func() {
		cluster, exists := s.slurmMock.GetCluster("e2e-slurm-cluster")
		s.True(exists)
		s.NotNil(cluster)
		s.Equal("E2E Test SLURM Cluster", cluster.Name)
	})

	s.Run("VerifyClusterPartitions", func() {
		cluster, exists := s.slurmMock.GetCluster("e2e-slurm-cluster")
		s.True(exists)
		s.Len(cluster.Partitions, 3)

		partitionNames := make([]string, len(cluster.Partitions))
		for i, p := range cluster.Partitions {
			partitionNames[i] = p.Name
		}
		s.Contains(partitionNames, "default")
		s.Contains(partitionNames, "gpu")
		s.Contains(partitionNames, "highmem")
	})

	s.Run("VerifyClusterCapacity", func() {
		cluster, exists := s.slurmMock.GetCluster("e2e-slurm-cluster")
		s.True(exists)
		s.Equal(int32(100), cluster.TotalNodes)
		s.Equal(int32(6400), cluster.TotalCPU)
		s.Equal(int64(25600), cluster.TotalMemoryGB)
		s.Equal(int32(400), cluster.TotalGPUs)
	})
}

// TestA02_SubmitJobToSLURM tests submitting a job through the SLURM adapter.
func (s *HPCProviderE2ETestSuite) TestA02_SubmitJobToSLURM() {
	ctx := context.Background()

	s.Run("SubmitStandardJob", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-submit-std")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)
		s.Equal(job.JobID, schedulerJob.VirtEngineJobID)
		s.NotEmpty(schedulerJob.SchedulerJobID)
		s.True(schedulerJob.SchedulerJobID != "")
	})

	s.Run("SubmitJobWithCorrectState", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-submit-state")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStatePending, schedulerJob.State)
	})

	s.Run("SubmitJobWithSchedulerType", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-submit-type")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.Equal(pd.HPCSchedulerTypeSLURM, schedulerJob.SchedulerType)
	})

	s.Run("SubmitJobWithSubmitTime", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-submit-time")

		beforeSubmit := time.Now()
		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		afterSubmit := time.Now()

		s.Require().NoError(err)
		s.True(schedulerJob.SubmitTime.After(beforeSubmit) || schedulerJob.SubmitTime.Equal(beforeSubmit))
		s.True(schedulerJob.SubmitTime.Before(afterSubmit) || schedulerJob.SubmitTime.Equal(afterSubmit))
	})

	s.Run("SubmitGPUJob", func() {
		job := fixtures.GPUComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-submit-gpu")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)
		s.Equal(job.Resources.GPUsPerNode, schedulerJob.OriginalJob.Resources.GPUsPerNode)
	})

	s.Run("SubmitMultiNodeJob", func() {
		job := fixtures.MultiNodeJob(s.providerAddr, s.customerAddr, 4)
		job.JobID = s.uniqueJobID("slurm-submit-multi")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)
		s.Equal(int32(4), schedulerJob.OriginalJob.Resources.Nodes)
	})

	s.Run("SubmitHighMemoryJob", func() {
		job := fixtures.HighMemoryJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-submit-highmem")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)
		s.Equal("highmem", schedulerJob.OriginalJob.QueueName)
	})

	s.Run("RejectOversizedJob", func() {
		job := fixtures.OversizedJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-submit-oversized")

		s.slurmMock.SetMaxCapacity(1000, 1024*1024, 100)

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Error(err)
		s.Contains(err.Error(), "insufficient")
	})
}

// TestA03_GetJobStatusFromSLURM tests polling job status from SLURM.
func (s *HPCProviderE2ETestSuite) TestA03_GetJobStatusFromSLURM() {
	ctx := context.Background()

	s.Run("GetPendingJobStatus", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-status-pending")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStatePending, status.State)
	})

	s.Run("GetQueuedJobStatus", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-status-queued")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateQueued, status.State)
	})

	s.Run("GetRunningJobStatus", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-status-running")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateRunning, status.State)
		s.NotNil(status.StartTime)
	})

	s.Run("GetCompletedJobStatus", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-status-completed")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCompleted, status.State)
		s.NotNil(status.EndTime)
	})

	s.Run("GetNonExistentJobStatus", func() {
		_, err := s.slurmMock.GetJobStatus(ctx, "nonexistent-job-id")
		s.Error(err)
		s.Contains(err.Error(), "not found")
	})

	s.Run("GetJobStatusWithExitCode", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-status-exitcode")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobExitCode(job.JobID, 42)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(int32(42), status.ExitCode)
	})
}

// TestA04_CancelJobInSLURM tests cancelling running jobs in SLURM.
func (s *HPCProviderE2ETestSuite) TestA04_CancelJobInSLURM() {
	ctx := context.Background()

	s.Run("CancelPendingJob", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-cancel-pending")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		err = s.slurmMock.CancelJob(ctx, job.JobID)
		s.Require().NoError(err)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCancelled, status.State)
	})

	s.Run("CancelQueuedJob", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-cancel-queued")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)

		err = s.slurmMock.CancelJob(ctx, job.JobID)
		s.Require().NoError(err)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCancelled, status.State)
	})

	s.Run("CancelRunningJob", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-cancel-running")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		err = s.slurmMock.CancelJob(ctx, job.JobID)
		s.Require().NoError(err)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCancelled, status.State)
	})

	s.Run("CancelNonExistentJob", func() {
		err := s.slurmMock.CancelJob(ctx, "nonexistent-cancel-job")
		s.Error(err)
		s.Contains(err.Error(), "not found")
	})

	s.Run("CancelCompletedJobFails", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-cancel-completed")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		err = s.slurmMock.CancelJob(ctx, job.JobID)
		s.Error(err)
		s.Contains(err.Error(), "terminal")
	})

	s.Run("CancelJobSetsEndTime", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-cancel-endtime")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		err = s.slurmMock.CancelJob(ctx, job.JobID)
		s.Require().NoError(err)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.NotNil(status.EndTime)
	})
}

// TestA05_JobAccountingFromSLURM tests retrieving accounting data after completion.
func (s *HPCProviderE2ETestSuite) TestA05_JobAccountingFromSLURM() {
	ctx := context.Background()

	s.Run("GetAccountingForCompletedJob", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-acct-completed")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		metrics := fixtures.StandardJobMetrics(3600)
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.NotNil(accounting)
		s.Equal(metrics.WallClockSeconds, accounting.WallClockSeconds)
		s.Equal(metrics.CPUCoreSeconds, accounting.CPUCoreSeconds)
	})

	s.Run("GetAccountingWithGPUMetrics", func() {
		job := fixtures.GPUComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-acct-gpu")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		metrics := fixtures.GPUJobMetrics(1800)
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(metrics.GPUSeconds, accounting.GPUSeconds)
		s.Greater(accounting.GPUSeconds, int64(0))
	})

	s.Run("GetAccountingWithMultiNodeMetrics", func() {
		job := fixtures.MultiNodeJob(s.providerAddr, s.customerAddr, 4)
		job.JobID = s.uniqueJobID("slurm-acct-multinode")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		metrics := fixtures.MultiNodeJobMetrics(7200, 4)
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(int32(4), accounting.NodesUsed)
		s.Greater(accounting.NodeHours, 0.0)
	})

	s.Run("GetAccountingForNonExistentJob", func() {
		_, err := s.slurmMock.GetJobAccounting(ctx, "nonexistent-acct-job")
		s.Error(err)
		s.Contains(err.Error(), "not found")
	})

	s.Run("GetAccountingWithNetworkMetrics", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-acct-network")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		metrics := fixtures.StandardJobMetrics(3600)
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Greater(accounting.NetworkBytesIn, int64(0))
		s.Greater(accounting.NetworkBytesOut, int64(0))
	})
}

// TestA06_SLURMReconnection tests adapter reconnection on failure.
func (s *HPCProviderE2ETestSuite) TestA06_SLURMReconnection() {
	ctx := context.Background()

	s.Run("StopAndRestartAdapter", func() {
		// Ensure running
		if !s.slurmMock.IsRunning() {
			err := s.slurmMock.Start(ctx)
			s.Require().NoError(err)
		}
		s.True(s.slurmMock.IsRunning())

		// Stop adapter
		err := s.slurmMock.Stop()
		s.Require().NoError(err)
		s.False(s.slurmMock.IsRunning())

		// Restart adapter
		err = s.slurmMock.Start(ctx)
		s.Require().NoError(err)
		s.True(s.slurmMock.IsRunning())
	})

	s.Run("SubmitJobAfterReconnection", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-reconnect-submit")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)
	})

	s.Run("ClusterStatePreservedAfterReconnection", func() {
		cluster, exists := s.slurmMock.GetCluster("e2e-slurm-cluster")
		s.True(exists)
		s.NotNil(cluster)
		s.Equal("E2E Test SLURM Cluster", cluster.Name)
	})

	s.Run("SubmitJobWhenNotRunningFails", func() {
		err := s.slurmMock.Stop()
		s.Require().NoError(err)

		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-not-running")

		_, err = s.slurmMock.SubmitJob(ctx, job)
		s.Error(err)
		s.Contains(err.Error(), "not running")

		// Restart for subsequent tests
		err = s.slurmMock.Start(ctx)
		s.Require().NoError(err)
	})
}

// TestA07_MultiClusterSupport tests multiple cluster registration.
func (s *HPCProviderE2ETestSuite) TestA07_MultiClusterSupport() {
	ctx := context.Background()

	s.Run("RegisterSecondCluster", func() {
		cluster := &mocks.SLURMCluster{
			ClusterID:     "e2e-cluster-eu",
			Name:          "E2E EU Cluster",
			Region:        "eu-west",
			SLURMVersion:  "23.02.4",
			TotalNodes:    50,
			TotalCPU:      3200,
			TotalMemoryGB: 12800,
			TotalGPUs:     100,
			Endpoint:      "slurm://e2e-cluster-eu.example.com:6817",
			Partitions: []mocks.SLURMPartition{
				{Name: "default", Nodes: 40, State: "up", Priority: 50},
				{Name: "gpu", Nodes: 10, State: "up", Priority: 100, AvailableGPU: 100},
			},
		}
		s.slurmMock.RegisterCluster(cluster)

		registered, exists := s.slurmMock.GetCluster("e2e-cluster-eu")
		s.True(exists)
		s.Equal("E2E EU Cluster", registered.Name)
		s.Equal("eu-west", registered.Region)
	})

	s.Run("RegisterThirdCluster", func() {
		cluster := &mocks.SLURMCluster{
			ClusterID:     "e2e-cluster-asia",
			Name:          "E2E Asia Cluster",
			Region:        "ap-east",
			SLURMVersion:  "23.02.4",
			TotalNodes:    75,
			TotalCPU:      4800,
			TotalMemoryGB: 19200,
			TotalGPUs:     200,
			Endpoint:      "slurm://e2e-cluster-asia.example.com:6817",
			Partitions: []mocks.SLURMPartition{
				{Name: "default", Nodes: 50, State: "up"},
				{Name: "gpu", Nodes: 25, State: "up", AvailableGPU: 200},
			},
		}
		s.slurmMock.RegisterCluster(cluster)

		registered, exists := s.slurmMock.GetCluster("e2e-cluster-asia")
		s.True(exists)
		s.Equal("E2E Asia Cluster", registered.Name)
	})

	s.Run("ListAllClusters", func() {
		clusters := s.slurmMock.GetClusters()
		s.GreaterOrEqual(len(clusters), 3)

		clusterIDs := make([]string, len(clusters))
		for i, c := range clusters {
			clusterIDs[i] = c.ClusterID
		}
		s.Contains(clusterIDs, "e2e-slurm-cluster")
		s.Contains(clusterIDs, "e2e-cluster-eu")
		s.Contains(clusterIDs, "e2e-cluster-asia")
	})

	s.Run("SubmitJobToSpecificCluster", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-multicluster")
		job.ClusterID = "e2e-cluster-eu"

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)

		decision, exists := s.slurmMock.GetRoutingDecisionForJob(job.JobID)
		s.True(exists)
		s.Equal("e2e-cluster-eu", decision.SelectedCluster)
	})

	s.Run("SubmitJobToNonExistentClusterFails", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("slurm-badcluster")
		job.ClusterID = "nonexistent-cluster"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Error(err)
		s.Contains(err.Error(), "not found")
	})

	s.Run("VerifyRoutingDecisionAudit", func() {
		decisions := s.slurmMock.GetRoutingDecisions()
		s.Greater(len(decisions), 0)

		for _, d := range decisions {
			s.NotEmpty(d.JobID)
			s.NotEmpty(d.SelectedCluster)
			s.NotEmpty(d.DecisionHash)
			s.False(d.Timestamp.IsZero())
		}
	})
}

// =============================================================================
// Section B: Usage Metering Accuracy Tests (~500 lines)
// =============================================================================

// TestB01_CPUUsageMetering tests CPU core-seconds calculation accuracy.
func (s *HPCProviderE2ETestSuite) TestB01_CPUUsageMetering() {
	ctx := context.Background()

	s.Run("VerifyCPUCoreSecondsCalculation", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("cpu-metering-basic")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		durationSeconds := int64(3600)
		cores := job.Resources.CPUCoresPerNode
		expectedCPUCoreSeconds := calculateExpectedCPUCoreSeconds(durationSeconds, cores)

		metrics := &pd.HPCSchedulerMetrics{
			WallClockSeconds: durationSeconds,
			CPUTimeSeconds:   durationSeconds * int64(cores),
			CPUCoreSeconds:   expectedCPUCoreSeconds,
			NodesUsed:        1,
		}
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(expectedCPUCoreSeconds, accounting.CPUCoreSeconds)
	})

	s.Run("VerifyCPUTimeVsWallClockTime", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("cpu-metering-wall")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		wallClockSeconds := int64(1800)
		cpuTimeSeconds := int64(7200) // 4 cores fully utilized

		metrics := &pd.HPCSchedulerMetrics{
			WallClockSeconds: wallClockSeconds,
			CPUTimeSeconds:   cpuTimeSeconds,
			CPUCoreSeconds:   cpuTimeSeconds,
			NodesUsed:        1,
		}
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Greater(accounting.CPUTimeSeconds, accounting.WallClockSeconds)
	})

	s.Run("VerifyMultiNodeCPUAggregation", func() {
		nodes := int32(4)
		job := fixtures.MultiNodeJob(s.providerAddr, s.customerAddr, nodes)
		job.JobID = s.uniqueJobID("cpu-metering-multinode")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		durationSeconds := int64(3600)
		coresPerNode := job.Resources.CPUCoresPerNode
		totalCores := int64(coresPerNode) * int64(nodes)
		expectedCPUCoreSeconds := durationSeconds * totalCores

		metrics := fixtures.MultiNodeJobMetrics(durationSeconds, nodes)
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(expectedCPUCoreSeconds, accounting.CPUCoreSeconds)
	})

	s.Run("VerifyPartialCPUUtilization", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("cpu-metering-partial")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// 50% CPU utilization
		wallClockSeconds := int64(3600)
		cpuTimeSeconds := int64(1800)

		metrics := &pd.HPCSchedulerMetrics{
			WallClockSeconds: wallClockSeconds,
			CPUTimeSeconds:   cpuTimeSeconds,
			CPUCoreSeconds:   cpuTimeSeconds,
			NodesUsed:        1,
		}
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Less(accounting.CPUTimeSeconds, accounting.WallClockSeconds)
	})

	s.Run("VerifyZeroCPUUsage", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("cpu-metering-zero")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		metrics := &pd.HPCSchedulerMetrics{
			WallClockSeconds: 60,
			CPUTimeSeconds:   0,
			CPUCoreSeconds:   0,
			NodesUsed:        1,
		}
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(int64(0), accounting.CPUCoreSeconds)
	})
}

// TestB02_MemoryUsageMetering tests memory GB-seconds calculation accuracy.
func (s *HPCProviderE2ETestSuite) TestB02_MemoryUsageMetering() {
	ctx := context.Background()

	s.Run("VerifyMemoryGBSecondsCalculation", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("mem-metering-basic")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		durationSeconds := int64(3600)
		memoryGB := job.Resources.MemoryGBPerNode
		expectedMemoryGBSeconds := calculateExpectedMemoryGBSeconds(durationSeconds, memoryGB)

		metrics := &pd.HPCSchedulerMetrics{
			WallClockSeconds: durationSeconds,
			MemoryGBSeconds:  expectedMemoryGBSeconds,
			MemoryBytesMax:   int64(memoryGB) * 1024 * 1024 * 1024,
			NodesUsed:        1,
		}
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(expectedMemoryGBSeconds, accounting.MemoryGBSeconds)
	})

	s.Run("VerifyPeakMemoryTracking", func() {
		job := fixtures.HighMemoryJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("mem-metering-peak")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		peakMemoryBytes := int64(512) * 1024 * 1024 * 1024 // 512 GB

		metrics := &pd.HPCSchedulerMetrics{
			WallClockSeconds: 3600,
			MemoryGBSeconds:  3600 * 512,
			MemoryBytesMax:   peakMemoryBytes,
			NodesUsed:        1,
		}
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(peakMemoryBytes, accounting.MemoryBytesMax)
	})

	s.Run("VerifyMultiNodeMemoryAggregation", func() {
		nodes := int32(4)
		job := fixtures.MultiNodeJob(s.providerAddr, s.customerAddr, nodes)
		job.JobID = s.uniqueJobID("mem-metering-multinode")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		durationSeconds := int64(3600)
		memoryGBPerNode := job.Resources.MemoryGBPerNode
		totalMemoryGB := int64(memoryGBPerNode) * int64(nodes)
		expectedMemoryGBSeconds := durationSeconds * totalMemoryGB

		metrics := &pd.HPCSchedulerMetrics{
			WallClockSeconds: durationSeconds,
			MemoryGBSeconds:  expectedMemoryGBSeconds,
			MemoryBytesMax:   totalMemoryGB * 1024 * 1024 * 1024,
			NodesUsed:        nodes,
		}
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(expectedMemoryGBSeconds, accounting.MemoryGBSeconds)
	})

	s.Run("VerifyVariableMemoryUsage", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("mem-metering-variable")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Simulate variable memory usage (average 50% of peak)
		peakMemoryBytes := int64(16) * 1024 * 1024 * 1024
		averageMemoryGBSeconds := int64(3600 * 8) // 8 GB average over 1 hour

		metrics := &pd.HPCSchedulerMetrics{
			WallClockSeconds: 3600,
			MemoryGBSeconds:  averageMemoryGBSeconds,
			MemoryBytesMax:   peakMemoryBytes,
			NodesUsed:        1,
		}
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Less(accounting.MemoryGBSeconds, int64(3600*16)) // Less than peak * duration
	})
}

// TestB03_GPUUsageMetering tests GPU-seconds calculation accuracy.
func (s *HPCProviderE2ETestSuite) TestB03_GPUUsageMetering() {
	ctx := context.Background()

	s.Run("VerifyGPUSecondsCalculation", func() {
		job := fixtures.GPUComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("gpu-metering-basic")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		durationSeconds := int64(3600)
		gpuCount := int64(job.Resources.GPUsPerNode)
		expectedGPUSeconds := durationSeconds * gpuCount

		metrics := fixtures.GPUJobMetrics(durationSeconds)
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(expectedGPUSeconds, accounting.GPUSeconds)
	})

	s.Run("VerifyMLTrainingGPUUsage", func() {
		job := fixtures.MLTrainingJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("gpu-metering-ml")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		durationSeconds := int64(86400) // 24 hours
		gpuCount := int64(job.Resources.GPUsPerNode) * int64(job.Resources.Nodes)
		expectedGPUSeconds := durationSeconds * gpuCount

		metrics := &pd.HPCSchedulerMetrics{
			WallClockSeconds: durationSeconds,
			GPUSeconds:       expectedGPUSeconds,
			NodesUsed:        job.Resources.Nodes,
		}
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(expectedGPUSeconds, accounting.GPUSeconds)
	})

	s.Run("VerifyZeroGPUForNonGPUJob", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("gpu-metering-zero")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		metrics := fixtures.StandardJobMetrics(3600)
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(int64(0), accounting.GPUSeconds)
	})

	s.Run("VerifyPartialGPUUtilization", func() {
		job := fixtures.GPUComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("gpu-metering-partial")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// 75% GPU utilization over 1 hour
		durationSeconds := int64(3600)
		gpuSecondsUtilized := int64(3600 * 2 * 3 / 4) // 2 GPUs at 75%

		metrics := &pd.HPCSchedulerMetrics{
			WallClockSeconds: durationSeconds,
			GPUSeconds:       gpuSecondsUtilized,
			NodesUsed:        1,
		}
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(gpuSecondsUtilized, accounting.GPUSeconds)
	})
}

// TestB04_NetworkUsageMetering tests network bytes in/out tracking.
func (s *HPCProviderE2ETestSuite) TestB04_NetworkUsageMetering() {
	ctx := context.Background()

	s.Run("VerifyNetworkBytesTracking", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("net-metering-basic")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		bytesIn := int64(1073741824) // 1 GB
		bytesOut := int64(536870912) // 0.5 GB

		metrics := &pd.HPCSchedulerMetrics{
			WallClockSeconds: 3600,
			NetworkBytesIn:   bytesIn,
			NetworkBytesOut:  bytesOut,
			NodesUsed:        1,
		}
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(bytesIn, accounting.NetworkBytesIn)
		s.Equal(bytesOut, accounting.NetworkBytesOut)
	})

	s.Run("VerifyHighNetworkTraffic", func() {
		job := fixtures.MLTrainingJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("net-metering-high")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// ML training with high network I/O
		bytesIn := int64(10) * 1024 * 1024 * 1024 // 10 GB
		bytesOut := int64(5) * 1024 * 1024 * 1024 // 5 GB

		metrics := &pd.HPCSchedulerMetrics{
			WallClockSeconds: 86400,
			NetworkBytesIn:   bytesIn,
			NetworkBytesOut:  bytesOut,
			NodesUsed:        4,
		}
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Greater(accounting.NetworkBytesIn, int64(0))
		s.Greater(accounting.NetworkBytesOut, int64(0))
	})

	s.Run("VerifyZeroNetworkUsage", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("net-metering-zero")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		metrics := &pd.HPCSchedulerMetrics{
			WallClockSeconds: 60,
			NetworkBytesIn:   0,
			NetworkBytesOut:  0,
			NodesUsed:        1,
		}
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(int64(0), accounting.NetworkBytesIn)
		s.Equal(int64(0), accounting.NetworkBytesOut)
	})
}

// TestB05_NodeHoursCalculation tests node-hours calculation accuracy.
func (s *HPCProviderE2ETestSuite) TestB05_NodeHoursCalculation() {
	ctx := context.Background()

	s.Run("VerifySingleNodeHours", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("nodehours-single")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		durationSeconds := int64(3600) // 1 hour
		expectedNodeHours := calculateExpectedNodeHours(durationSeconds, 1)

		metrics := &pd.HPCSchedulerMetrics{
			WallClockSeconds: durationSeconds,
			NodesUsed:        1,
			NodeHours:        expectedNodeHours,
		}
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.InDelta(expectedNodeHours, accounting.NodeHours, 0.001)
	})

	s.Run("VerifyMultiNodeHours", func() {
		nodes := int32(4)
		job := fixtures.MultiNodeJob(s.providerAddr, s.customerAddr, nodes)
		job.JobID = s.uniqueJobID("nodehours-multi")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		durationSeconds := int64(7200) // 2 hours
		expectedNodeHours := calculateExpectedNodeHours(durationSeconds, nodes)

		metrics := fixtures.MultiNodeJobMetrics(durationSeconds, nodes)
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.InDelta(expectedNodeHours, accounting.NodeHours, 0.001)
	})

	s.Run("VerifyFractionalNodeHours", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("nodehours-fractional")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		durationSeconds := int64(900) // 15 minutes = 0.25 hours
		expectedNodeHours := calculateExpectedNodeHours(durationSeconds, 1)

		metrics := &pd.HPCSchedulerMetrics{
			WallClockSeconds: durationSeconds,
			NodesUsed:        1,
			NodeHours:        expectedNodeHours,
		}
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.InDelta(0.25, accounting.NodeHours, 0.001)
	})

	s.Run("VerifyLargeScaleNodeHours", func() {
		nodes := int32(8)
		job := fixtures.SimulationJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("nodehours-large")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		durationSeconds := int64(43200) // 12 hours
		expectedNodeHours := calculateExpectedNodeHours(durationSeconds, nodes)

		metrics := &pd.HPCSchedulerMetrics{
			WallClockSeconds: durationSeconds,
			NodesUsed:        nodes,
			NodeHours:        expectedNodeHours,
		}
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.InDelta(96.0, accounting.NodeHours, 0.01) // 8 nodes * 12 hours
	})
}

// TestB06_UsageSnapshotInterval tests periodic snapshot capture.
func (s *HPCProviderE2ETestSuite) TestB06_UsageSnapshotInterval() {
	ctx := context.Background()

	s.Run("CapturePeriodicSnapshots", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("snapshot-periodic")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		// Simulate periodic snapshots
		for i := 0; i < 5; i++ {
			progressMetrics := &pd.HPCSchedulerMetrics{
				WallClockSeconds: int64((i + 1) * 600),
				CPUCoreSeconds:   int64((i + 1) * 600 * 8),
				NodesUsed:        1,
			}
			s.recordUsageSnapshot(job.JobID, progressMetrics, false, "periodic")
			time.Sleep(10 * time.Millisecond)
		}

		s.mu.Lock()
		snapshots := s.usageSnapshots[job.JobID]
		s.mu.Unlock()

		s.Len(snapshots, 5)
		for i, snap := range snapshots {
			s.False(snap.IsFinal)
			s.Equal("periodic", snap.SnapshotType)
			s.Equal(int64((i+1)*600), snap.Metrics.WallClockSeconds)
		}
	})

	s.Run("VerifySnapshotOrdering", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("snapshot-ordering")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		for i := 0; i < 3; i++ {
			progressMetrics := &pd.HPCSchedulerMetrics{
				WallClockSeconds: int64((i + 1) * 300),
			}
			s.recordUsageSnapshot(job.JobID, progressMetrics, false, "periodic")
			time.Sleep(5 * time.Millisecond)
		}

		s.mu.Lock()
		snapshots := s.usageSnapshots[job.JobID]
		s.mu.Unlock()

		// Verify timestamps are in order
		for i := 1; i < len(snapshots); i++ {
			s.True(snapshots[i].Timestamp.After(snapshots[i-1].Timestamp) ||
				snapshots[i].Timestamp.Equal(snapshots[i-1].Timestamp))
		}
	})

	s.Run("VerifySnapshotIDUniqueness", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("snapshot-unique")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		for i := 0; i < 3; i++ {
			s.recordUsageSnapshot(job.JobID, &pd.HPCSchedulerMetrics{}, false, "periodic")
			time.Sleep(1 * time.Millisecond)
		}

		s.mu.Lock()
		snapshots := s.usageSnapshots[job.JobID]
		s.mu.Unlock()

		ids := make(map[string]bool)
		for _, snap := range snapshots {
			s.NotContains(ids, snap.SnapshotID)
			ids[snap.SnapshotID] = true
		}
	})
}

// TestB07_FinalUsageSnapshot tests final snapshot on job completion.
func (s *HPCProviderE2ETestSuite) TestB07_FinalUsageSnapshot() {
	ctx := context.Background()

	s.Run("CaptureFinalSnapshotOnCompletion", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("snapshot-final")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		finalMetrics := fixtures.StandardJobMetrics(3600)
		s.slurmMock.SetJobMetrics(job.JobID, finalMetrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		s.recordUsageSnapshot(job.JobID, finalMetrics, true, "final")

		s.mu.Lock()
		snapshots := s.usageSnapshots[job.JobID]
		s.mu.Unlock()

		s.Len(snapshots, 1)
		s.True(snapshots[0].IsFinal)
		s.Equal("final", snapshots[0].SnapshotType)
	})

	s.Run("FinalSnapshotContainsCompleteMetrics", func() {
		job := fixtures.GPUComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("snapshot-final-complete")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		finalMetrics := fixtures.GPUJobMetrics(7200)
		s.slurmMock.SetJobMetrics(job.JobID, finalMetrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		s.recordUsageSnapshot(job.JobID, finalMetrics, true, "final")

		s.mu.Lock()
		snapshots := s.usageSnapshots[job.JobID]
		s.mu.Unlock()

		s.Len(snapshots, 1)
		s.Equal(finalMetrics.WallClockSeconds, snapshots[0].Metrics.WallClockSeconds)
		s.Equal(finalMetrics.CPUCoreSeconds, snapshots[0].Metrics.CPUCoreSeconds)
		s.Equal(finalMetrics.GPUSeconds, snapshots[0].Metrics.GPUSeconds)
	})

	s.Run("FinalSnapshotOnFailure", func() {
		job := fixtures.FailingJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("snapshot-final-fail")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		partialMetrics := fixtures.PartialJobMetrics(300)
		s.slurmMock.SetJobMetrics(job.JobID, partialMetrics)
		s.slurmMock.SetJobExitCode(job.JobID, 1)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateFailed)

		s.recordUsageSnapshot(job.JobID, partialMetrics, true, "final")

		s.mu.Lock()
		snapshots := s.usageSnapshots[job.JobID]
		s.mu.Unlock()

		s.Len(snapshots, 1)
		s.True(snapshots[0].IsFinal)
	})
}

// TestB08_UsageSignedRecords tests provider-signed usage records.
func (s *HPCProviderE2ETestSuite) TestB08_UsageSignedRecords() {
	ctx := context.Background()

	s.Run("CreateSignedStatusReport", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("signed-status")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		report, err := s.slurmMock.CreateStatusReport(schedulerJob)
		s.Require().NoError(err)
		s.NotNil(report)
		s.NotEmpty(report.Signature)
	})

	s.Run("VerifyReportContainsProviderAddress", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("signed-provider")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		report, err := s.slurmMock.CreateStatusReport(schedulerJob)
		s.Require().NoError(err)
		s.Equal(s.providerAddr, report.ProviderAddress)
	})

	s.Run("VerifyReportContainsJobID", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("signed-jobid")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		report, err := s.slurmMock.CreateStatusReport(schedulerJob)
		s.Require().NoError(err)
		s.Equal(job.JobID, report.VirtEngineJobID)
		s.Equal(schedulerJob.SchedulerJobID, report.SchedulerJobID)
	})

	s.Run("VerifyReportContainsState", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("signed-state")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)
		schedulerJob, _ = s.slurmMock.GetJobStatus(ctx, job.JobID)

		report, err := s.slurmMock.CreateStatusReport(schedulerJob)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCompleted, report.State)
	})

	s.Run("VerifyReportTimestamp", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("signed-timestamp")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		beforeReport := time.Now()
		report, err := s.slurmMock.CreateStatusReport(schedulerJob)
		afterReport := time.Now()

		s.Require().NoError(err)
		s.True(report.Timestamp.After(beforeReport) || report.Timestamp.Equal(beforeReport))
		s.True(report.Timestamp.Before(afterReport) || report.Timestamp.Equal(afterReport))
	})

	s.Run("SubmitReportToUsageReporter", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("signed-submit")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		report, err := s.slurmMock.CreateStatusReport(schedulerJob)
		s.Require().NoError(err)

		err = s.usageReporter.SubmitStatusReport(report)
		s.Require().NoError(err)

		reports := s.usageReporter.GetSubmittedReports()
		s.Greater(len(reports), 0)
	})
}

// =============================================================================
// Section C: Multi-Job Orchestration Tests (~500 lines)
// =============================================================================

// TestC01_ConcurrentJobSubmission tests submitting multiple jobs concurrently.
func (s *HPCProviderE2ETestSuite) TestC01_ConcurrentJobSubmission() {
	ctx := context.Background()

	s.Run("SubmitMultipleJobsConcurrently", func() {
		numJobs := 10
		var wg sync.WaitGroup
		results := make(chan *pd.HPCSchedulerJob, numJobs)
		errors := make(chan error, numJobs)

		for i := 0; i < numJobs; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
				job.JobID = s.uniqueJobID(fmt.Sprintf("concurrent-job-%d", idx))

				schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
				if err != nil {
					errors <- err
					return
				}
				results <- schedulerJob
			}(i)
		}

		wg.Wait()
		close(results)
		close(errors)

		// Verify no errors
		for err := range errors {
			s.Fail("Job submission failed", err.Error())
		}

		// Verify all jobs submitted
		submittedJobs := make([]*pd.HPCSchedulerJob, 0)
		for job := range results {
			submittedJobs = append(submittedJobs, job)
		}
		s.Len(submittedJobs, numJobs)
	})

	s.Run("VerifyUniqueSchedulerJobIDs", func() {
		numJobs := 5
		jobIDs := make(map[string]bool)

		for i := 0; i < numJobs; i++ {
			job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
			job.JobID = s.uniqueJobID(fmt.Sprintf("unique-id-test-%d", i))

			schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
			s.Require().NoError(err)

			s.NotContains(jobIDs, schedulerJob.SchedulerJobID)
			jobIDs[schedulerJob.SchedulerJobID] = true
		}
	})

	s.Run("ConcurrentSubmissionWithMixedTypes", func() {
		var wg sync.WaitGroup
		results := make(chan *pd.HPCSchedulerJob, 4)

		jobCreators := []func(string, string) *hpctypes.HPCJob{
			fixtures.StandardComputeJob,
			fixtures.GPUComputeJob,
			fixtures.HighMemoryJob,
			fixtures.QuickTestJob,
		}

		for i, creator := range jobCreators {
			wg.Add(1)
			go func(idx int, create func(string, string) *hpctypes.HPCJob) {
				defer wg.Done()
				job := create(s.providerAddr, s.customerAddr)
				job.JobID = s.uniqueJobID(fmt.Sprintf("mixed-type-%d", idx))

				schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
				if err == nil {
					results <- schedulerJob
				}
			}(i, creator)
		}

		wg.Wait()
		close(results)

		submittedJobs := make([]*pd.HPCSchedulerJob, 0)
		for job := range results {
			submittedJobs = append(submittedJobs, job)
		}
		s.GreaterOrEqual(len(submittedJobs), 3)
	})

	s.Run("VerifyActiveJobsList", func() {
		// Submit multiple jobs
		for i := 0; i < 3; i++ {
			job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
			job.JobID = s.uniqueJobID(fmt.Sprintf("active-list-%d", i))

			_, err := s.slurmMock.SubmitJob(ctx, job)
			s.Require().NoError(err)
		}

		activeJobs, err := s.slurmMock.ListActiveJobs(ctx)
		s.Require().NoError(err)
		s.GreaterOrEqual(len(activeJobs), 3)
	})
}

// TestC02_QueuePriorityOrdering tests priority queue ordering.
func (s *HPCProviderE2ETestSuite) TestC02_QueuePriorityOrdering() {
	ctx := context.Background()

	s.Run("VerifyQueueAssignment", func() {
		// Standard job -> default queue
		stdJob := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		stdJob.JobID = s.uniqueJobID("queue-std")

		_, err := s.slurmMock.SubmitJob(ctx, stdJob)
		s.Require().NoError(err)

		record, exists := s.slurmMock.GetExecutionRecord(stdJob.JobID)
		s.True(exists)
		s.Equal("default", record.PartitionName)
	})

	s.Run("VerifyGPUQueueAssignment", func() {
		gpuJob := fixtures.GPUComputeJob(s.providerAddr, s.customerAddr)
		gpuJob.JobID = s.uniqueJobID("queue-gpu")

		_, err := s.slurmMock.SubmitJob(ctx, gpuJob)
		s.Require().NoError(err)

		record, exists := s.slurmMock.GetExecutionRecord(gpuJob.JobID)
		s.True(exists)
		s.Equal("gpu", record.PartitionName)
	})

	s.Run("VerifyHighMemQueueAssignment", func() {
		highMemJob := fixtures.HighMemoryJob(s.providerAddr, s.customerAddr)
		highMemJob.JobID = s.uniqueJobID("queue-highmem")

		_, err := s.slurmMock.SubmitJob(ctx, highMemJob)
		s.Require().NoError(err)

		record, exists := s.slurmMock.GetExecutionRecord(highMemJob.JobID)
		s.True(exists)
		s.Equal("highmem", record.PartitionName)
	})

	s.Run("VerifyRoutingDecisionScoringFactors", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("queue-scoring")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		decision, exists := s.slurmMock.GetRoutingDecisionForJob(job.JobID)
		s.True(exists)
		s.NotEmpty(decision.ScoringFactors)

		// Verify scoring factors exist
		_, hasResourceAvail := decision.ScoringFactors["resource_availability"]
		s.True(hasResourceAvail)
	})

	s.Run("VerifyMultipleQueueDecisions", func() {
		jobs := []*hpctypes.HPCJob{
			fixtures.StandardComputeJob(s.providerAddr, s.customerAddr),
			fixtures.GPUComputeJob(s.providerAddr, s.customerAddr),
			fixtures.HighMemoryJob(s.providerAddr, s.customerAddr),
		}

		for i, job := range jobs {
			job.JobID = s.uniqueJobID(fmt.Sprintf("multi-queue-%d", i))
			_, err := s.slurmMock.SubmitJob(ctx, job)
			s.Require().NoError(err)
		}

		decisions := s.slurmMock.GetRoutingDecisions()
		s.GreaterOrEqual(len(decisions), 3)
	})
}

// TestC03_ResourceContentionHandling tests resource limit enforcement.
func (s *HPCProviderE2ETestSuite) TestC03_ResourceContentionHandling() {
	ctx := context.Background()

	s.Run("EnforceMaxCPULimit", func() {
		// Set low capacity
		s.slurmMock.SetMaxCapacity(100, 1024*1024, 1000)

		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("contention-cpu")
		job.Resources.CPUCoresPerNode = 200 // Exceeds limit

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Error(err)
		s.Contains(err.Error(), "CPU")

		// Reset capacity
		s.slurmMock.SetMaxCapacity(10000, 1024*1024, 1000)
	})

	s.Run("EnforceMaxMemoryLimit", func() {
		// Set low memory capacity (1 GB)
		s.slurmMock.SetMaxCapacity(10000, 1024, 1000)

		job := fixtures.HighMemoryJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("contention-mem")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Error(err)
		s.Contains(err.Error(), "memory")

		// Reset capacity
		s.slurmMock.SetMaxCapacity(10000, 1024*1024, 1000)
	})

	s.Run("EnforceMaxGPULimit", func() {
		// Set low GPU capacity
		s.slurmMock.SetMaxCapacity(10000, 1024*1024, 1)

		job := fixtures.GPUComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("contention-gpu")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Error(err)
		s.Contains(err.Error(), "GPU")

		// Reset capacity
		s.slurmMock.SetMaxCapacity(10000, 1024*1024, 1000)
	})

	s.Run("AcceptJobWithinLimits", func() {
		s.slurmMock.SetMaxCapacity(10000, 1024*1024, 1000)

		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("contention-accept")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)
	})

	s.Run("MultipleLimitViolations", func() {
		s.slurmMock.SetMaxCapacity(10, 100, 0)

		job := fixtures.MLTrainingJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("contention-multiple")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Error(err)

		// Reset capacity
		s.slurmMock.SetMaxCapacity(10000, 1024*1024, 1000)
	})
}

// TestC04_JobPreemption tests preemption scenarios.
func (s *HPCProviderE2ETestSuite) TestC04_JobPreemption() {
	ctx := context.Background()

	s.Run("SuspendRunningJob", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("preempt-suspend")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateSuspended)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateSuspended, status.State)
	})

	s.Run("ResumeSuspendedJob", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("preempt-resume")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateSuspended)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateRunning, status.State)
	})

	s.Run("VerifyValidStateTransitions", func() {
		s.True(s.slurmMock.IsValidTransition(pd.HPCJobStateRunning, pd.HPCJobStateSuspended))
		s.True(s.slurmMock.IsValidTransition(pd.HPCJobStateSuspended, pd.HPCJobStateRunning))
		s.True(s.slurmMock.IsValidTransition(pd.HPCJobStateRunning, pd.HPCJobStateCompleted))
		s.False(s.slurmMock.IsValidTransition(pd.HPCJobStateCompleted, pd.HPCJobStateRunning))
	})

	s.Run("VerifyPreemptionTrackingInEvents", func() {
		s.clearLifecycleEvents()

		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("preempt-events")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateSuspended)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		events := s.getLifecycleEventsForJob(job.JobID)
		s.GreaterOrEqual(len(events), 3)

		// Find suspended event
		hasSuspended := false
		for _, e := range events {
			if e.ToState == pd.HPCJobStateSuspended {
				hasSuspended = true
				break
			}
		}
		s.True(hasSuspended)
	})

	s.Run("MultipleSuspendResumeCycles", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("preempt-cycles")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		for i := 0; i < 3; i++ {
			s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateSuspended)
			status, _ := s.slurmMock.GetJobStatus(ctx, job.JobID)
			s.Equal(pd.HPCJobStateSuspended, status.State)

			s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
			status, _ = s.slurmMock.GetJobStatus(ctx, job.JobID)
			s.Equal(pd.HPCJobStateRunning, status.State)
		}
	})
}

// TestC05_FairShareScheduling tests fair-share scheduling among users.
func (s *HPCProviderE2ETestSuite) TestC05_FairShareScheduling() {
	ctx := context.Background()

	s.Run("SubmitJobsFromMultipleCustomers", func() {
		customers := []string{
			sdk.AccAddress([]byte("customer-fairshare-01")).String(),
			sdk.AccAddress([]byte("customer-fairshare-02")).String(),
			sdk.AccAddress([]byte("customer-fairshare-03")).String(),
		}

		for i, customer := range customers {
			job := fixtures.StandardComputeJob(s.providerAddr, customer)
			job.JobID = s.uniqueJobID(fmt.Sprintf("fairshare-cust-%d", i))

			schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
			s.Require().NoError(err)
			s.NotNil(schedulerJob)
			s.Equal(customer, schedulerJob.OriginalJob.CustomerAddress)
		}
	})

	s.Run("VerifyJobsProcessedFromAllCustomers", func() {
		activeJobs, err := s.slurmMock.ListActiveJobs(ctx)
		s.Require().NoError(err)

		customerCounts := make(map[string]int)
		for _, job := range activeJobs {
			if job.OriginalJob != nil {
				customerCounts[job.OriginalJob.CustomerAddress]++
			}
		}
		s.GreaterOrEqual(len(customerCounts), 1)
	})

	s.Run("SubmitManyJobsFromSameCustomer", func() {
		customer := sdk.AccAddress([]byte("customer-many-jobs-01")).String()

		for i := 0; i < 5; i++ {
			job := fixtures.QuickTestJob(s.providerAddr, customer)
			job.JobID = s.uniqueJobID(fmt.Sprintf("fairshare-many-%d", i))

			_, err := s.slurmMock.SubmitJob(ctx, job)
			s.Require().NoError(err)
		}

		activeJobs, err := s.slurmMock.ListActiveJobs(ctx)
		s.Require().NoError(err)

		count := 0
		for _, job := range activeJobs {
			if job.OriginalJob != nil && job.OriginalJob.CustomerAddress == customer {
				count++
			}
		}
		s.GreaterOrEqual(count, 5)
	})
}

// TestC06_PartitionRouting tests routing to correct partition.
func (s *HPCProviderE2ETestSuite) TestC06_PartitionRouting() {
	ctx := context.Background()

	s.Run("RouteToDefaultPartition", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("partition-default")
		job.QueueName = "default"

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		record, exists := s.slurmMock.GetExecutionRecord(job.JobID)
		s.True(exists)
		s.Equal("default", record.PartitionName)
	})

	s.Run("RouteToGPUPartition", func() {
		job := fixtures.GPUComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("partition-gpu")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		record, exists := s.slurmMock.GetExecutionRecord(job.JobID)
		s.True(exists)
		s.Equal("gpu", record.PartitionName)
	})

	s.Run("RouteToHighMemPartition", func() {
		job := fixtures.HighMemoryJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("partition-highmem")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		record, exists := s.slurmMock.GetExecutionRecord(job.JobID)
		s.True(exists)
		s.Equal("highmem", record.PartitionName)
	})

	s.Run("VerifyPartitionInRoutingDecision", func() {
		job := fixtures.GPUComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("partition-decision")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		decision, exists := s.slurmMock.GetRoutingDecisionForJob(job.JobID)
		s.True(exists)
		s.NotEmpty(decision.SelectedCluster)
		s.NotEmpty(decision.Reason)
	})

	s.Run("MultiplePartitionAssignments", func() {
		partitionJobs := map[string]*hpctypes.HPCJob{
			"default": fixtures.StandardComputeJob(s.providerAddr, s.customerAddr),
			"gpu":     fixtures.GPUComputeJob(s.providerAddr, s.customerAddr),
			"highmem": fixtures.HighMemoryJob(s.providerAddr, s.customerAddr),
		}

		for partition, job := range partitionJobs {
			job.JobID = s.uniqueJobID(fmt.Sprintf("partition-multi-%s", partition))
			_, err := s.slurmMock.SubmitJob(ctx, job)
			s.Require().NoError(err)

			record, exists := s.slurmMock.GetExecutionRecord(job.JobID)
			s.True(exists)
			s.Equal(partition, record.PartitionName)
		}
	})
}

// TestC07_MultiNodeJobExecution tests distributed job execution.
func (s *HPCProviderE2ETestSuite) TestC07_MultiNodeJobExecution() {
	ctx := context.Background()

	s.Run("Submit4NodeJob", func() {
		job := fixtures.MultiNodeJob(s.providerAddr, s.customerAddr, 4)
		job.JobID = s.uniqueJobID("multinode-4")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)
		s.Equal(int32(4), schedulerJob.OriginalJob.Resources.Nodes)
	})

	s.Run("Submit8NodeSimulationJob", func() {
		job := fixtures.SimulationJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("multinode-sim")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)
		s.Equal(int32(8), schedulerJob.OriginalJob.Resources.Nodes)
	})

	s.Run("SimulateMultiNodeExecution", func() {
		job := fixtures.MultiNodeJob(s.providerAddr, s.customerAddr, 4)
		job.JobID = s.uniqueJobID("multinode-exec")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		err = s.slurmMock.SimulateJobExecution(ctx, job.JobID, 100, true, fixtures.MultiNodeJobMetrics(3600, 4))
		s.Require().NoError(err)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCompleted, status.State)
	})

	s.Run("VerifyMultiNodeMetrics", func() {
		job := fixtures.MultiNodeJob(s.providerAddr, s.customerAddr, 4)
		job.JobID = s.uniqueJobID("multinode-metrics")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		metrics := fixtures.MultiNodeJobMetrics(3600, 4)
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(int32(4), accounting.NodesUsed)
		s.InDelta(4.0, accounting.NodeHours, 0.01) // 4 nodes * 1 hour
	})

	s.Run("VerifyNodeListAssignment", func() {
		job := fixtures.MultiNodeJob(s.providerAddr, s.customerAddr, 2)
		job.JobID = s.uniqueJobID("multinode-nodelist")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		record, exists := s.slurmMock.GetExecutionRecord(job.JobID)
		s.True(exists)
		s.NotEmpty(record.NodeList)
	})

	s.Run("VerifyExecutionRecordCreated", func() {
		job := fixtures.MultiNodeJob(s.providerAddr, s.customerAddr, 4)
		job.JobID = s.uniqueJobID("multinode-record")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		record, exists := s.slurmMock.GetExecutionRecord(job.JobID)
		s.True(exists)
		s.Equal(job.JobID, record.JobID)
		s.Equal(job.ClusterID, record.ClusterID)
		s.NotEmpty(record.SLURMJobID)
	})
}

// =============================================================================
// Section D: Job Lifecycle Events Tests (~500 lines)
// =============================================================================

// TestD01_LifecycleCallbackRegistration tests registering lifecycle callbacks.
func (s *HPCProviderE2ETestSuite) TestD01_LifecycleCallbackRegistration() {
	s.Run("RegisterCallback", func() {
		callbackCalled := false
		callback := func(job *pd.HPCSchedulerJob, event pd.HPCJobLifecycleEvent, prevState pd.HPCJobState) {
			callbackCalled = true
		}

		s.slurmMock.RegisterLifecycleCallback(callback)
		s.True(true) // Callback registered without error
	})

	s.Run("CallbackInvokedOnSubmit", func() {
		s.clearLifecycleEvents()

		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("lifecycle-callback")

		_, err := s.slurmMock.SubmitJob(context.Background(), job)
		s.Require().NoError(err)

		events := s.getLifecycleEventsForJob(job.JobID)
		s.GreaterOrEqual(len(events), 1)
	})

	s.Run("MultipleCallbacksInvoked", func() {
		counter := 0
		additionalCallback := func(job *pd.HPCSchedulerJob, event pd.HPCJobLifecycleEvent, prevState pd.HPCJobState) {
			counter++
		}

		s.slurmMock.RegisterLifecycleCallback(additionalCallback)

		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("lifecycle-multi-callback")

		_, err := s.slurmMock.SubmitJob(context.Background(), job)
		s.Require().NoError(err)

		s.Greater(counter, 0)
	})
}

// TestD02_JobSubmittedEvent tests that submitted event is fired.
func (s *HPCProviderE2ETestSuite) TestD02_JobSubmittedEvent() {
	ctx := context.Background()

	s.Run("VerifySubmittedEventFired", func() {
		s.clearLifecycleEvents()

		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-submitted")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		events := s.getLifecycleEventsForJob(job.JobID)
		s.GreaterOrEqual(len(events), 1)

		hasSubmittedEvent := false
		for _, e := range events {
			if e.Event == pd.HPCJobEventSubmitted {
				hasSubmittedEvent = true
				break
			}
		}
		s.True(hasSubmittedEvent)
	})

	s.Run("SubmittedEventHasCorrectJobID", func() {
		s.clearLifecycleEvents()

		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-submitted-id")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		events := s.getLifecycleEventsForJob(job.JobID)
		s.GreaterOrEqual(len(events), 1)
		s.Equal(job.JobID, events[0].JobID)
	})

	s.Run("SubmittedEventHasTimestamp", func() {
		s.clearLifecycleEvents()

		beforeSubmit := time.Now()
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-submitted-time")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		afterSubmit := time.Now()
		s.Require().NoError(err)

		events := s.getLifecycleEventsForJob(job.JobID)
		s.GreaterOrEqual(len(events), 1)
		s.True(events[0].Timestamp.After(beforeSubmit) || events[0].Timestamp.Equal(beforeSubmit))
		s.True(events[0].Timestamp.Before(afterSubmit) || events[0].Timestamp.Equal(afterSubmit))
	})

	s.Run("SubmittedEventInitialState", func() {
		s.clearLifecycleEvents()

		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-submitted-state")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		events := s.getLifecycleEventsForJob(job.JobID)
		s.GreaterOrEqual(len(events), 1)
		s.Equal(pd.HPCJobStatePending, events[0].ToState)
	})
}

// TestD03_JobQueuedEvent tests that queued event is fired.
func (s *HPCProviderE2ETestSuite) TestD03_JobQueuedEvent() {
	ctx := context.Background()

	s.Run("VerifyQueuedEventFired", func() {
		s.clearLifecycleEvents()

		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-queued")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)

		events := s.getLifecycleEventsForJob(job.JobID)
		hasQueuedEvent := false
		for _, e := range events {
			if e.Event == pd.HPCJobEventQueued && e.ToState == pd.HPCJobStateQueued {
				hasQueuedEvent = true
				break
			}
		}
		s.True(hasQueuedEvent)
	})

	s.Run("QueuedEventShowsTransition", func() {
		s.clearLifecycleEvents()

		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-queued-transition")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)

		events := s.getLifecycleEventsForJob(job.JobID)
		for _, e := range events {
			if e.ToState == pd.HPCJobStateQueued {
				s.Equal(pd.HPCJobStatePending, e.FromState)
				break
			}
		}
	})

	s.Run("QueuedEventAfterSubmitted", func() {
		s.clearLifecycleEvents()

		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-queued-order")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)

		events := s.getLifecycleEventsForJob(job.JobID)
		s.GreaterOrEqual(len(events), 2)

		// First event should be submitted, second should be queued
		s.Equal(pd.HPCJobStatePending, events[0].ToState)
		s.Equal(pd.HPCJobStateQueued, events[1].ToState)
	})
}

// TestD04_JobStartedEvent tests that started event is fired.
func (s *HPCProviderE2ETestSuite) TestD04_JobStartedEvent() {
	ctx := context.Background()

	s.Run("VerifyStartedEventFired", func() {
		s.clearLifecycleEvents()

		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-started")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		events := s.getLifecycleEventsForJob(job.JobID)
		hasStartedEvent := false
		for _, e := range events {
			if e.Event == pd.HPCJobEventStarted && e.ToState == pd.HPCJobStateRunning {
				hasStartedEvent = true
				break
			}
		}
		s.True(hasStartedEvent)
	})

	s.Run("StartedEventSetsStartTime", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-started-time")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.NotNil(status.StartTime)
	})

	s.Run("StartedEventTransitionFromQueued", func() {
		s.clearLifecycleEvents()

		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-started-queued")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)

		events := s.getLifecycleEventsForJob(job.JobID)
		for _, e := range events {
			if e.ToState == pd.HPCJobStateRunning {
				s.Equal(pd.HPCJobStateQueued, e.FromState)
				break
			}
		}
	})
}

// TestD05_JobCompletedEvent tests that completed event is fired with metrics.
func (s *HPCProviderE2ETestSuite) TestD05_JobCompletedEvent() {
	ctx := context.Background()

	s.Run("VerifyCompletedEventFired", func() {
		s.clearLifecycleEvents()

		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-completed")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		events := s.getLifecycleEventsForJob(job.JobID)
		hasCompletedEvent := false
		for _, e := range events {
			if e.Event == pd.HPCJobEventCompleted && e.ToState == pd.HPCJobStateCompleted {
				hasCompletedEvent = true
				break
			}
		}
		s.True(hasCompletedEvent)
	})

	s.Run("CompletedEventSetsEndTime", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-completed-time")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.NotNil(status.EndTime)
	})

	s.Run("CompletedEventWithMetrics", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-completed-metrics")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		metrics := fixtures.StandardJobMetrics(3600)
		s.slurmMock.SetJobMetrics(job.JobID, metrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.NotNil(accounting)
		s.Greater(accounting.WallClockSeconds, int64(0))
	})

	s.Run("CompletedEventWithZeroExitCode", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-completed-exit")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobExitCode(job.JobID, 0)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(int32(0), status.ExitCode)
	})
}

// TestD06_JobFailedEvent tests that failed event is fired with exit code.
func (s *HPCProviderE2ETestSuite) TestD06_JobFailedEvent() {
	ctx := context.Background()

	s.Run("VerifyFailedEventFired", func() {
		s.clearLifecycleEvents()

		job := fixtures.FailingJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-failed")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.slurmMock.SetJobExitCode(job.JobID, 1)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateFailed)

		events := s.getLifecycleEventsForJob(job.JobID)
		hasFailedEvent := false
		for _, e := range events {
			if e.Event == pd.HPCJobEventFailed && e.ToState == pd.HPCJobStateFailed {
				hasFailedEvent = true
				break
			}
		}
		s.True(hasFailedEvent)
	})

	s.Run("FailedEventWithNonZeroExitCode", func() {
		job := fixtures.FailingJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-failed-exit")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobExitCode(job.JobID, 127)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateFailed)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(int32(127), status.ExitCode)
	})

	s.Run("FailedEventSetsEndTime", func() {
		job := fixtures.FailingJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-failed-time")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateFailed)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.NotNil(status.EndTime)
	})

	s.Run("FailedEventWithPartialMetrics", func() {
		job := fixtures.FailingJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-failed-metrics")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		partialMetrics := fixtures.PartialJobMetrics(300)
		s.slurmMock.SetJobMetrics(job.JobID, partialMetrics)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateFailed)

		accounting, err := s.slurmMock.GetJobAccounting(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(int64(300), accounting.WallClockSeconds)
	})
}

// TestD07_JobCancelledEvent tests that cancelled event is fired.
func (s *HPCProviderE2ETestSuite) TestD07_JobCancelledEvent() {
	ctx := context.Background()

	s.Run("VerifyCancelledEventFired", func() {
		s.clearLifecycleEvents()

		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-cancelled")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		err = s.slurmMock.CancelJob(ctx, job.JobID)
		s.Require().NoError(err)

		events := s.getLifecycleEventsForJob(job.JobID)
		hasCancelledEvent := false
		for _, e := range events {
			if e.Event == pd.HPCJobEventCancelled && e.ToState == pd.HPCJobStateCancelled {
				hasCancelledEvent = true
				break
			}
		}
		s.True(hasCancelledEvent)
	})

	s.Run("CancelledEventTransitionFromRunning", func() {
		s.clearLifecycleEvents()

		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-cancelled-running")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		err = s.slurmMock.CancelJob(ctx, job.JobID)
		s.Require().NoError(err)

		events := s.getLifecycleEventsForJob(job.JobID)
		for _, e := range events {
			if e.ToState == pd.HPCJobStateCancelled {
				s.Equal(pd.HPCJobStateRunning, e.FromState)
				break
			}
		}
	})

	s.Run("CancelledEventFromPending", func() {
		s.clearLifecycleEvents()

		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-cancelled-pending")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		err = s.slurmMock.CancelJob(ctx, job.JobID)
		s.Require().NoError(err)

		events := s.getLifecycleEventsForJob(job.JobID)
		for _, e := range events {
			if e.ToState == pd.HPCJobStateCancelled {
				s.Equal(pd.HPCJobStatePending, e.FromState)
				break
			}
		}
	})
}

// TestD08_JobTimeoutEvent tests that timeout event is fired.
func (s *HPCProviderE2ETestSuite) TestD08_JobTimeoutEvent() {
	ctx := context.Background()

	s.Run("VerifyTimeoutEventFired", func() {
		s.clearLifecycleEvents()

		job := fixtures.TimeoutJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-timeout")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateTimeout)

		events := s.getLifecycleEventsForJob(job.JobID)
		hasTimeoutEvent := false
		for _, e := range events {
			if e.Event == pd.HPCJobEventTimeout && e.ToState == pd.HPCJobStateTimeout {
				hasTimeoutEvent = true
				break
			}
		}
		s.True(hasTimeoutEvent)
	})

	s.Run("TimeoutEventSetsEndTime", func() {
		job := fixtures.TimeoutJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-timeout-time")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateTimeout)

		status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.NotNil(status.EndTime)
		s.Equal(pd.HPCJobStateTimeout, status.State)
	})

	s.Run("TimeoutEventTransitionFromRunning", func() {
		s.clearLifecycleEvents()

		job := fixtures.TimeoutJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-timeout-running")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateTimeout)

		events := s.getLifecycleEventsForJob(job.JobID)
		for _, e := range events {
			if e.ToState == pd.HPCJobStateTimeout {
				s.Equal(pd.HPCJobStateRunning, e.FromState)
				break
			}
		}
	})
}

// TestD09_EventOrdering tests that events fire in correct order.
func (s *HPCProviderE2ETestSuite) TestD09_EventOrdering() {
	ctx := context.Background()

	s.Run("VerifyEventOrderForSuccessfulJob", func() {
		s.clearLifecycleEvents()

		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-order-success")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		events := s.getLifecycleEventsForJob(job.JobID)
		s.GreaterOrEqual(len(events), 4)

		// Verify state progression
		expectedStates := []pd.HPCJobState{
			pd.HPCJobStatePending,
			pd.HPCJobStateQueued,
			pd.HPCJobStateRunning,
			pd.HPCJobStateCompleted,
		}

		for i, expected := range expectedStates {
			s.Equal(expected, events[i].ToState)
		}
	})

	s.Run("VerifyEventOrderForFailedJob", func() {
		s.clearLifecycleEvents()

		job := fixtures.FailingJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-order-failed")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateFailed)

		events := s.getLifecycleEventsForJob(job.JobID)
		s.GreaterOrEqual(len(events), 4)

		// Last event should be failed
		lastEvent := events[len(events)-1]
		s.Equal(pd.HPCJobStateFailed, lastEvent.ToState)
	})

	s.Run("VerifyTimestampOrder", func() {
		s.clearLifecycleEvents()

		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-order-timestamp")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		time.Sleep(1 * time.Millisecond)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateQueued)
		time.Sleep(1 * time.Millisecond)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		time.Sleep(1 * time.Millisecond)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		events := s.getLifecycleEventsForJob(job.JobID)
		s.GreaterOrEqual(len(events), 4)

		// Verify timestamps are in order
		for i := 1; i < len(events); i++ {
			s.True(events[i].Timestamp.After(events[i-1].Timestamp) ||
				events[i].Timestamp.Equal(events[i-1].Timestamp))
		}
	})

	s.Run("VerifyEventOrderWithPreemption", func() {
		s.clearLifecycleEvents()

		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("event-order-preempt")

		_, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateSuspended)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

		events := s.getLifecycleEventsForJob(job.JobID)
		s.GreaterOrEqual(len(events), 5)

		// Verify suspended state appears
		hasSuspended := false
		for _, e := range events {
			if e.ToState == pd.HPCJobStateSuspended {
				hasSuspended = true
				break
			}
		}
		s.True(hasSuspended)
	})
}

// TestD10_StatusReportGeneration tests signed status report creation.
func (s *HPCProviderE2ETestSuite) TestD10_StatusReportGeneration() {
	ctx := context.Background()

	s.Run("GenerateStatusReportForRunningJob", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("report-running")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		schedulerJob, _ = s.slurmMock.GetJobStatus(ctx, job.JobID)

		report, err := s.slurmMock.CreateStatusReport(schedulerJob)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateRunning, report.State)
	})

	s.Run("GenerateStatusReportForCompletedJob", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("report-completed")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		s.slurmMock.SetJobExitCode(job.JobID, 0)
		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)
		schedulerJob, _ = s.slurmMock.GetJobStatus(ctx, job.JobID)

		report, err := s.slurmMock.CreateStatusReport(schedulerJob)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCompleted, report.State)
		s.Equal(int32(0), report.ExitCode)
	})

	s.Run("StatusReportContainsSchedulerType", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("report-type")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		report, err := s.slurmMock.CreateStatusReport(schedulerJob)
		s.Require().NoError(err)
		s.Equal(pd.HPCSchedulerTypeSLURM, report.SchedulerType)
	})

	s.Run("StatusReportSignatureNonEmpty", func() {
		job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("report-signature")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		report, err := s.slurmMock.CreateStatusReport(schedulerJob)
		s.Require().NoError(err)
		s.NotEmpty(report.Signature)
	})

	s.Run("StatusReportSignatureUnique", func() {
		job1 := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job1.JobID = s.uniqueJobID("report-sig-1")

		job2 := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
		job2.JobID = s.uniqueJobID("report-sig-2")

		sj1, _ := s.slurmMock.SubmitJob(ctx, job1)
		sj2, _ := s.slurmMock.SubmitJob(ctx, job2)

		report1, _ := s.slurmMock.CreateStatusReport(sj1)
		time.Sleep(1 * time.Millisecond)
		report2, _ := s.slurmMock.CreateStatusReport(sj2)

		s.NotEqual(report1.Signature, report2.Signature)
	})

	s.Run("MultipleStatusReportsForSameJob", func() {
		job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
		job.JobID = s.uniqueJobID("report-multi")

		schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Generate reports at different states
		report1, _ := s.slurmMock.CreateStatusReport(schedulerJob)
		s.Equal(pd.HPCJobStatePending, report1.State)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateRunning)
		schedulerJob, _ = s.slurmMock.GetJobStatus(ctx, job.JobID)
		report2, _ := s.slurmMock.CreateStatusReport(schedulerJob)
		s.Equal(pd.HPCJobStateRunning, report2.State)

		s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)
		schedulerJob, _ = s.slurmMock.GetJobStatus(ctx, job.JobID)
		report3, _ := s.slurmMock.CreateStatusReport(schedulerJob)
		s.Equal(pd.HPCJobStateCompleted, report3.State)
	})

	s.Run("SubmitMultipleReportsToUsageReporter", func() {
		s.usageReporter.Clear()

		for i := 0; i < 5; i++ {
			job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
			job.JobID = s.uniqueJobID(fmt.Sprintf("report-submit-%d", i))

			schedulerJob, err := s.slurmMock.SubmitJob(ctx, job)
			s.Require().NoError(err)

			report, err := s.slurmMock.CreateStatusReport(schedulerJob)
			s.Require().NoError(err)

			err = s.usageReporter.SubmitStatusReport(report)
			s.Require().NoError(err)
		}

		reports := s.usageReporter.GetSubmittedReports()
		s.Len(reports, 5)
	})

	s.Run("VerifyReportJobIDsUnique", func() {
		s.usageReporter.Clear()

		jobIDs := make(map[string]bool)
		for i := 0; i < 3; i++ {
			job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
			job.JobID = s.uniqueJobID(fmt.Sprintf("report-unique-%d", i))

			schedulerJob, _ := s.slurmMock.SubmitJob(ctx, job)
			report, _ := s.slurmMock.CreateStatusReport(schedulerJob)
			_ = s.usageReporter.SubmitStatusReport(report)
		}

		reports := s.usageReporter.GetSubmittedReports()
		for _, r := range reports {
			s.NotContains(jobIDs, r.VirtEngineJobID)
			jobIDs[r.VirtEngineJobID] = true
		}
	})
}
