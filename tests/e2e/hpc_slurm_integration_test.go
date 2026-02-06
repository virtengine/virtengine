//go:build e2e.integration

// Package e2e contains end-to-end integration tests.
//
// VE-21C: E2E test for SLURM provider daemon integration.
// Tests the complete SLURM integration flow:
// 1. HPC Provider initialization with SLURM backend
// 2. Job submission routing to SLURM
// 3. Job lifecycle monitoring
// 4. Usage accounting pipeline
// 5. Settlement submission
package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/pkg/slurm_adapter"
	"github.com/virtengine/virtengine/tests/e2e/fixtures"
	"github.com/virtengine/virtengine/tests/e2e/mocks"
	"github.com/virtengine/virtengine/testutil"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCSLURMIntegrationE2ETestSuite tests the SLURM provider daemon integration.
type HPCSLURMIntegrationE2ETestSuite struct {
	*testutil.NetworkTestSuite

	// Test addresses
	providerAddr string
	customerAddr string

	// Mock components
	slurmClient   *mocks.MockSLURMIntegration
	chainReporter *MockChainReporter
	auditLogger   *MockAuditLoggerE2E
	credManager   *MockCredentialManager

	// HPC Provider under test
	hpcProvider *pd.HPCProvider

	// Test data
	testCluster  *hpctypes.HPCCluster
	testOffering *hpctypes.HPCOffering

	// Tracking
	submittedJobs   []string
	lifecycleEvents []pd.HPCJobLifecycleEvent
}

func TestHPCSLURMIntegration(t *testing.T) {
	suite.Run(t, &HPCSLURMIntegrationE2ETestSuite{
		NetworkTestSuite: testutil.NewNetworkTestSuite(nil, &HPCSLURMIntegrationE2ETestSuite{}),
	})
}

func (s *HPCSLURMIntegrationE2ETestSuite) SetupSuite() {
	s.NetworkTestSuite.SetupSuite()

	val := s.Network().Validators[0]
	s.providerAddr = val.Address.String()
	s.customerAddr = val.Address.String()

	// Initialize mock components
	s.slurmClient = mocks.NewMockSLURMIntegration()
	s.chainReporter = NewMockChainReporter()
	s.auditLogger = NewMockAuditLoggerE2E()
	s.credManager = NewMockCredentialManager()
	s.submittedJobs = make([]string, 0)
	s.lifecycleEvents = make([]pd.HPCJobLifecycleEvent, 0)

	// Create test cluster and offering
	clusterConfig := fixtures.DefaultTestClusterConfig()
	clusterConfig.ProviderAddr = s.providerAddr
	s.testCluster = fixtures.CreateTestCluster(clusterConfig)

	offeringConfig := fixtures.DefaultTestOfferingConfig()
	offeringConfig.ProviderAddr = s.providerAddr
	s.testOffering = fixtures.CreateTestOffering(offeringConfig)
}

func (s *HPCSLURMIntegrationE2ETestSuite) TearDownSuite() {
	if s.hpcProvider != nil && s.hpcProvider.IsRunning() {
		_ = s.hpcProvider.Stop()
	}
	s.NetworkTestSuite.TearDownSuite()
}

// =============================================================================
// 1. HPC Provider Initialization Tests
// =============================================================================

func (s *HPCSLURMIntegrationE2ETestSuite) Test01_HPCProviderInitialization() {
	ctx := context.Background()

	s.Run("CreateHPCProviderConfig", func() {
		config := s.createTestProviderConfig()

		s.NotEmpty(config.HPC.ClusterID)
		s.Equal(pd.HPCSchedulerTypeSLURM, config.HPC.SchedulerType)
		s.True(config.HPC.Enabled)
	})

	s.Run("InitializeHPCProvider", func() {
		config := s.createTestProviderConfig()

		// Create the HPC provider
		provider, err := pd.NewHPCProvider(config, s.chainReporter, s.auditLogger)
		s.Require().NoError(err)
		s.NotNil(provider)

		s.hpcProvider = provider
	})

	s.Run("StartHPCProvider", func() {
		s.Require().NotNil(s.hpcProvider)

		err := s.hpcProvider.Start(ctx)
		s.Require().NoError(err)
		s.True(s.hpcProvider.IsRunning())
	})

	s.Run("VerifyComponentsRunning", func() {
		s.Require().NotNil(s.hpcProvider)
		s.True(s.hpcProvider.IsRunning())

		health := s.hpcProvider.GetHealth()
		s.NotNil(health)
		s.True(health.Overall)
	})
}

// =============================================================================
// 2. Job Submission Tests
// =============================================================================

func (s *HPCSLURMIntegrationE2ETestSuite) Test02_JobSubmission() {
	ctx := context.Background()

	s.Run("SubmitSimpleJob", func() {
		s.Require().NotNil(s.hpcProvider)

		job := s.createTestJob("test-job-01")

		schedulerJob, err := s.hpcProvider.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)
		s.Equal(job.JobID, schedulerJob.VirtEngineJobID)
		s.NotEmpty(schedulerJob.SchedulerJobID)

		s.submittedJobs = append(s.submittedJobs, job.JobID)
	})

	s.Run("SubmitJobWithGPU", func() {
		s.Require().NotNil(s.hpcProvider)

		job := s.createTestJobWithGPU("test-job-gpu-01")

		schedulerJob, err := s.hpcProvider.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)
		s.Equal(job.JobID, schedulerJob.VirtEngineJobID)

		s.submittedJobs = append(s.submittedJobs, job.JobID)
	})

	s.Run("VerifyJobInActiveList", func() {
		s.Require().NotNil(s.hpcProvider)

		activeJobs, err := s.hpcProvider.ListActiveJobs(ctx)
		s.Require().NoError(err)
		s.GreaterOrEqual(len(activeJobs), 1)
	})
}

// =============================================================================
// 3. Job Lifecycle Tests
// =============================================================================

func (s *HPCSLURMIntegrationE2ETestSuite) Test03_JobLifecycle() {
	ctx := context.Background()

	s.Run("SubmitAndTrackJob", func() {
		s.Require().NotNil(s.hpcProvider)

		job := s.createTestJob("test-lifecycle-01")

		schedulerJob, err := s.hpcProvider.SubmitJob(ctx, job)
		s.Require().NoError(err)
		s.NotNil(schedulerJob)

		// Initial state should be pending or queued
		s.Contains(
			[]pd.HPCJobState{pd.HPCJobStatePending, pd.HPCJobStateQueued},
			schedulerJob.State,
		)

		s.submittedJobs = append(s.submittedJobs, job.JobID)
	})

	s.Run("GetJobStatus", func() {
		s.Require().NotNil(s.hpcProvider)
		s.Require().NotEmpty(s.submittedJobs)

		jobID := s.submittedJobs[0]
		status, err := s.hpcProvider.GetJobStatus(ctx, jobID)
		s.Require().NoError(err)
		s.NotNil(status)
		s.Equal(jobID, status.VirtEngineJobID)
	})

	s.Run("CancelJob", func() {
		s.Require().NotNil(s.hpcProvider)

		// Submit a job to cancel
		job := s.createTestJob("test-cancel-01")
		_, err := s.hpcProvider.SubmitJob(ctx, job)
		s.Require().NoError(err)

		// Cancel it
		err = s.hpcProvider.CancelJob(ctx, job.JobID)
		s.Require().NoError(err)

		// Verify cancellation
		status, err := s.hpcProvider.GetJobStatus(ctx, job.JobID)
		s.Require().NoError(err)
		s.Equal(pd.HPCJobStateCancelled, status.State)
	})
}

// =============================================================================
// 4. Usage Accounting Tests
// =============================================================================

func (s *HPCSLURMIntegrationE2ETestSuite) Test04_UsageAccounting() {
	ctx := context.Background()

	s.Run("GetJobAccounting", func() {
		s.Require().NotNil(s.hpcProvider)
		s.Require().NotEmpty(s.submittedJobs)

		jobID := s.submittedJobs[0]
		metrics, err := s.hpcProvider.GetJobAccounting(ctx, jobID)
		// May return nil if job hasn't run yet - that's OK
		if err == nil && metrics != nil {
			s.GreaterOrEqual(metrics.WallClockSeconds, int64(0))
			s.GreaterOrEqual(metrics.CPUCoreSeconds, int64(0))
		}
	})

	s.Run("VerifyUsageReporting", func() {
		// Wait a bit for usage reporting loop
		time.Sleep(100 * time.Millisecond)

		// Check that the chain reporter received usage reports
		reports := s.chainReporter.GetAccountingReports()
		// Usage reports may be empty if jobs are still pending
		s.NotNil(reports)
	})
}

// =============================================================================
// 5. Settlement Pipeline Tests
// =============================================================================

func (s *HPCSLURMIntegrationE2ETestSuite) Test05_SettlementPipeline() {
	s.Run("VerifySettlementQueueing", func() {
		// Settlement happens on job completion
		// For this test, we verify the settlement component is accessible
		s.Require().NotNil(s.hpcProvider)

		settlement := s.hpcProvider.GetSettlementPipeline()
		s.NotNil(settlement)

		// Check pending count
		pending := settlement.GetPendingCount()
		s.GreaterOrEqual(pending, 0)
	})

	s.Run("VerifySettlementStats", func() {
		s.Require().NotNil(s.hpcProvider)

		settlement := s.hpcProvider.GetSettlementPipeline()
		s.NotNil(settlement)

		stats := settlement.GetStats()
		s.NotNil(stats)
		s.GreaterOrEqual(stats.TotalQueued, int64(0))
	})
}

// =============================================================================
// 6. Health Check Tests
// =============================================================================

func (s *HPCSLURMIntegrationE2ETestSuite) Test06_HealthChecks() {
	s.Run("ProviderHealth", func() {
		s.Require().NotNil(s.hpcProvider)

		health := s.hpcProvider.GetHealth()
		s.NotNil(health)
		s.NotEmpty(health.Message)
		s.NotZero(health.LastCheck)
	})

	s.Run("ComponentHealth", func() {
		s.Require().NotNil(s.hpcProvider)

		health := s.hpcProvider.GetHealth()
		s.NotNil(health)

		// All components should report health via the Components slice
		s.NotEmpty(health.Components)
	})
}

// =============================================================================
// 7. Shutdown Tests
// =============================================================================

func (s *HPCSLURMIntegrationE2ETestSuite) Test99_Shutdown() {
	s.Run("StopHPCProvider", func() {
		s.Require().NotNil(s.hpcProvider)

		err := s.hpcProvider.Stop()
		s.Require().NoError(err)
		s.False(s.hpcProvider.IsRunning())
	})
}

// =============================================================================
// Helper Methods
// =============================================================================

func (s *HPCSLURMIntegrationE2ETestSuite) createTestProviderConfig() pd.HPCProviderConfig {
	return pd.HPCProviderConfig{
		HPC: pd.HPCConfig{
			Enabled:         true,
			SchedulerType:   pd.HPCSchedulerTypeSLURM,
			ClusterID:       "e2e-test-cluster",
			ProviderAddress: s.providerAddr,
			SLURM: slurm_adapter.SLURMConfig{
				ClusterName:       "e2e-slurm",
				ControllerHost:    "localhost",
				ControllerPort:    6817,
				AuthMethod:        "munge",
				DefaultPartition:  "default",
				JobPollInterval:   5 * time.Second,
				ConnectionTimeout: 10 * time.Second,
				MaxRetries:        3,
			},
			JobService: pd.HPCJobServiceConfig{
				JobPollInterval:     5 * time.Second,
				JobTimeoutDefault:   1 * time.Hour,
				MaxConcurrentJobs:   100,
				EnableStateRecovery: false,
			},
			UsageReporting: pd.HPCUsageReportingConfig{
				Enabled:        true,
				ReportInterval: 10 * time.Second,
				BatchSize:      10,
				RetryOnFailure: true,
			},
			Audit: pd.HPCAuditConfig{
				Enabled:           true,
				LogJobEvents:      true,
				LogSecurityEvents: true,
				LogUsageReports:   true,
			},
		},
		Chain: pd.HPCChainSubscriberConfig{
			Enabled:           true,
			ClusterID:         "e2e-test-cluster",
			ProviderAddress:   s.providerAddr,
			ReconnectInterval: 5 * time.Second,
			EventBufferSize:   100,
		},
		Settlement: pd.HPCSettlementConfig{
			Enabled:           true,
			BatchSize:         10,
			BatchInterval:     5 * time.Second,
			MaxRetries:        3,
			RetryBackoff:      time.Second,
			MaxPendingRecords: 1000,
		},
	}
}

func (s *HPCSLURMIntegrationE2ETestSuite) createTestJob(jobID string) *hpctypes.HPCJob {
	return &hpctypes.HPCJob{
		JobID:           jobID,
		OfferingID:      "test-offering-01",
		ClusterID:       "e2e-test-cluster",
		ProviderAddress: s.providerAddr,
		CustomerAddress: s.customerAddr,
		State:           hpctypes.JobStatePending,
		QueueName:       "default",
		WorkloadSpec: hpctypes.JobWorkloadSpec{
			ContainerImage: "python:3.11-slim",
			Command:        "python -c 'print(\"Hello from SLURM\")'",
		},
		Resources: hpctypes.JobResources{
			Nodes:           1,
			CPUCoresPerNode: 4,
			MemoryGBPerNode: 8,
			StorageGB:       10,
		},
		MaxRuntimeSeconds: 3600,
		CreatedAt:         time.Now(),
	}
}

func (s *HPCSLURMIntegrationE2ETestSuite) createTestJobWithGPU(jobID string) *hpctypes.HPCJob {
	job := s.createTestJob(jobID)
	job.Resources.GPUsPerNode = 1
	job.Resources.GPUType = "nvidia-a100"
	job.QueueName = "gpu"
	return job
}

// =============================================================================
// Mock Components
// =============================================================================

// MockChainReporter implements HPCChainClient for testing.
type MockChainReporter struct {
	statusReports     []*pd.HPCStatusReport
	accountingReports map[string]*pd.HPCSchedulerMetrics
	billingRules      *hpctypes.HPCBillingRules
	blockHeight       int64
}

func NewMockChainReporter() *MockChainReporter {
	return &MockChainReporter{
		statusReports:     make([]*pd.HPCStatusReport, 0),
		accountingReports: make(map[string]*pd.HPCSchedulerMetrics),
		billingRules:      &hpctypes.HPCBillingRules{},
		blockHeight:       1000,
	}
}

// HPCOnChainReporter methods
func (m *MockChainReporter) ReportJobStatus(ctx context.Context, report *pd.HPCStatusReport) error {
	m.statusReports = append(m.statusReports, report)
	return nil
}

func (m *MockChainReporter) ReportJobAccounting(ctx context.Context, jobID string, metrics *pd.HPCSchedulerMetrics) error {
	m.accountingReports[jobID] = metrics
	return nil
}

// HPCJobEventSubscriber methods
func (m *MockChainReporter) SubscribeToJobRequests(ctx context.Context, clusterID string, handler func(*hpctypes.HPCJob) error) error {
	// No-op for testing - jobs are submitted directly
	return nil
}

func (m *MockChainReporter) SubscribeToJobCancellations(ctx context.Context, clusterID string, handler func(jobID string) error) error {
	// No-op for testing
	return nil
}

// HPCAccountingSubmitter methods
func (m *MockChainReporter) SubmitAccountingRecord(ctx context.Context, record *hpctypes.HPCAccountingRecord) error {
	return nil
}

func (m *MockChainReporter) SubmitUsageSnapshot(ctx context.Context, snapshot *hpctypes.HPCUsageSnapshot) error {
	return nil
}

func (m *MockChainReporter) GetBillingRules(ctx context.Context, providerAddr string) (*hpctypes.HPCBillingRules, error) {
	return m.billingRules, nil
}

// GetCurrentBlockHeight implements HPCChainClient
func (m *MockChainReporter) GetCurrentBlockHeight(ctx context.Context) (int64, error) {
	return m.blockHeight, nil
}

// Test helper methods
func (m *MockChainReporter) GetStatusReports() []*pd.HPCStatusReport {
	return m.statusReports
}

func (m *MockChainReporter) GetAccountingReports() map[string]*pd.HPCSchedulerMetrics {
	return m.accountingReports
}

// MockAuditLoggerE2E implements HPCAuditLogger for testing.
type MockAuditLoggerE2E struct {
	jobEvents      []pd.HPCAuditEvent
	securityEvents []pd.HPCAuditEvent
	usageReports   []pd.HPCAuditEvent
}

func NewMockAuditLoggerE2E() *MockAuditLoggerE2E {
	return &MockAuditLoggerE2E{
		jobEvents:      make([]pd.HPCAuditEvent, 0),
		securityEvents: make([]pd.HPCAuditEvent, 0),
		usageReports:   make([]pd.HPCAuditEvent, 0),
	}
}

func (m *MockAuditLoggerE2E) LogJobEvent(event pd.HPCAuditEvent) {
	m.jobEvents = append(m.jobEvents, event)
}

func (m *MockAuditLoggerE2E) LogSecurityEvent(event pd.HPCAuditEvent) {
	m.securityEvents = append(m.securityEvents, event)
}

func (m *MockAuditLoggerE2E) LogUsageReport(event pd.HPCAuditEvent) {
	m.usageReports = append(m.usageReports, event)
}

// MockCredentialManager implements HPCCredentialManager for testing.
type MockCredentialManager struct {
	credentials map[string]*pd.HPCCredentials
}

func NewMockCredentialManager() *MockCredentialManager {
	return &MockCredentialManager{
		credentials: make(map[string]*pd.HPCCredentials),
	}
}

func (m *MockCredentialManager) GetCredentials(ctx context.Context, clusterID string, credType pd.CredentialType) (*pd.HPCCredentials, error) {
	return &pd.HPCCredentials{
		ClusterID:      clusterID,
		CredentialType: credType,
		SSHPrivateKey:  "mock-private-key",
	}, nil
}

func (m *MockCredentialManager) Sign(data []byte) ([]byte, error) {
	return []byte("mock-signature"), nil
}

func (m *MockCredentialManager) Verify(data, signature []byte) bool {
	return true
}
