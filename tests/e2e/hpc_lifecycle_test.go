//go:build e2e.integration

// Package e2e contains end-to-end integration tests.
//
// HPC-E2E-001: E2E lifecycle coverage from submission to settlement.
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

// LifecycleStateMachine validates lifecycle transitions.
type LifecycleStateMachine struct {
	expected []mocks.LifecyclePhase
}

func NewLifecycleStateMachine(expected []mocks.LifecyclePhase) *LifecycleStateMachine {
	return &LifecycleStateMachine{expected: expected}
}

func (m *LifecycleStateMachine) Validate(phases []mocks.LifecyclePhase) bool {
	if len(phases) < len(m.expected) {
		return false
	}
	idx := 0
	for _, phase := range phases {
		if phase == m.expected[idx] {
			idx++
			if idx == len(m.expected) {
				return true
			}
		}
	}
	return false
}

// HPCJobLifecycleE2ETestSuite covers full lifecycle states.
type HPCJobLifecycleE2ETestSuite struct {
	*testutil.NetworkTestSuite

	providerAddr string
	customerAddr string

	slurmMock    *mocks.MockSLURMIntegration
	providerMock *mocks.MockHPCProviderDaemon
	settlement   *BillingMockSettlementProcessor
}

func TestHPCJobLifecycleE2E(t *testing.T) {
	suite.Run(t, &HPCJobLifecycleE2ETestSuite{
		NetworkTestSuite: testutil.NewNetworkTestSuite(nil, &HPCJobLifecycleE2ETestSuite{}),
	})
}

func (s *HPCJobLifecycleE2ETestSuite) SetupSuite() {
	s.NetworkTestSuite.SetupSuite()

	val := s.Network().Validators[0]
	s.providerAddr = val.Address.String()
	s.customerAddr = val.Address.String()

	s.slurmMock = mocks.NewMockSLURMIntegration()
	s.providerMock = mocks.NewMockHPCProviderDaemon(s.slurmMock)
	s.settlement = NewBillingMockSettlementProcessor()

	ctx := context.Background()
	err := s.slurmMock.Start(ctx)
	s.Require().NoError(err)

	cluster := mocks.DefaultTestCluster()
	s.slurmMock.RegisterCluster(cluster)

	s.providerMock.AddCluster(mocks.ProviderCluster{
		ClusterID:        cluster.ClusterID,
		ProviderID:       "provider-1",
		Region:           cluster.Region,
		AvailableCPU:     cluster.TotalCPU,
		AvailableMemory:  cluster.TotalMemoryGB,
		AvailableGPUs:    cluster.TotalGPUs,
		GPUType:          "nvidia-a100",
		LatencyScore:     0.9,
		PriceScore:       0.8,
		IdentityTier:     90,
		SupportsGPUTypes: []string{"nvidia-a100", "nvidia-v100"},
	})
}

func (s *HPCJobLifecycleE2ETestSuite) TearDownSuite() {
	if s.slurmMock != nil && s.slurmMock.IsRunning() {
		_ = s.slurmMock.Stop()
	}
	s.NetworkTestSuite.TearDownSuite()
}

func (s *HPCJobLifecycleE2ETestSuite) TestHappyPathLifecycle() {
	ctx := context.Background()

	job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
	job.JobID = "lifecycle-happy-1"

	err := s.providerMock.EnqueueJob(job, mocks.JobQueueOptions{
		Priority:       90,
		CustomerTier:   85,
		RequiredTier:   70,
		RequiredRegion: "us-east",
	})
	s.Require().NoError(err)

	decision, err := s.providerMock.ScheduleNext(ctx)
	s.Require().NoError(err)
	s.Equal(job.ClusterID, decision.SelectedClusterID)

	schedulerJob, err := s.providerMock.StartJob(ctx, job.JobID)
	s.Require().NoError(err)
	s.NotNil(schedulerJob)

	// Simulate execution
	metrics := fixtures.StandardJobMetrics(3600)
	s.slurmMock.SetJobMetrics(job.JobID, metrics)
	s.slurmMock.SetJobExitCode(job.JobID, 0)
	s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

	s.providerMock.MarkCompleted(job.JobID, metrics)

	// Create accounting record and settle
	record := &hpctypes.HPCAccountingRecord{
		RecordID:        fmt.Sprintf("record-%s", job.JobID),
		JobID:           job.JobID,
		ClusterID:       job.ClusterID,
		ProviderAddress: s.providerAddr,
		CustomerAddress: s.customerAddr,
		OfferingID:      job.OfferingID,
		SchedulerType:   "SLURM",
		UsageMetrics: hpctypes.HPCDetailedMetrics{
			WallClockSeconds: metrics.WallClockSeconds,
			CPUCoreSeconds:   metrics.CPUCoreSeconds,
			MemoryGBSeconds:  metrics.MemoryGBSeconds,
			GPUSeconds:       metrics.GPUSeconds,
			StorageGBHours:   metrics.StorageGBHours,
			NetworkBytesIn:   metrics.NetworkBytesIn,
			NetworkBytesOut:  metrics.NetworkBytesOut,
			NodeHours:        sdkmath.LegacyNewDec(int64(metrics.NodeHours)),
			NodesUsed:        metrics.NodesUsed,
		},
		BillableAmount: sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1000000))),
		Status:         hpctypes.AccountingStatusFinalized,
		PeriodStart:    time.Now().Add(-time.Hour),
		PeriodEnd:      time.Now(),
		FormulaVersion: hpctypes.CurrentBillingFormulaVersion,
		CreatedAt:      time.Now(),
	}

	result := s.settlement.ProcessSettlement(record, time.Now())
	s.True(result.Success)
	s.providerMock.MarkSettled(job.JobID)

	phases := s.providerMock.GetLifecycle(job.JobID)
	machine := NewLifecycleStateMachine([]mocks.LifecyclePhase{
		mocks.LifecycleSubmitted,
		mocks.LifecycleQueued,
		mocks.LifecycleScheduled,
		mocks.LifecycleRunning,
		mocks.LifecycleCompleted,
		mocks.LifecycleSettled,
	})
	s.True(machine.Validate(phases))
}

func (s *HPCJobLifecycleE2ETestSuite) TestFailureLifecycle() {
	ctx := context.Background()

	job := fixtures.FailingJob(s.providerAddr, s.customerAddr)
	job.JobID = "lifecycle-fail-1"

	err := s.providerMock.EnqueueJob(job, mocks.JobQueueOptions{
		Priority:       80,
		CustomerTier:   85,
		RequiredTier:   70,
		RequiredRegion: "us-east",
	})
	s.Require().NoError(err)

	_, err = s.providerMock.ScheduleNext(ctx)
	s.Require().NoError(err)

	_, err = s.providerMock.StartJob(ctx, job.JobID)
	s.Require().NoError(err)

	metrics := fixtures.PartialJobMetrics(600)
	s.slurmMock.SetJobMetrics(job.JobID, metrics)
	s.slurmMock.SetJobExitCode(job.JobID, 1)
	s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateFailed)

	s.providerMock.MarkFailed(job.JobID, metrics)

	phases := s.providerMock.GetLifecycle(job.JobID)
	machine := NewLifecycleStateMachine([]mocks.LifecyclePhase{
		mocks.LifecycleSubmitted,
		mocks.LifecycleQueued,
		mocks.LifecycleScheduled,
		mocks.LifecycleRunning,
		mocks.LifecycleFailed,
	})
	s.True(machine.Validate(phases))
}

func (s *HPCJobLifecycleE2ETestSuite) TestTimeoutLifecycle() {
	ctx := context.Background()

	job := fixtures.TimeoutJob(s.providerAddr, s.customerAddr)
	job.JobID = "lifecycle-timeout-1"

	err := s.providerMock.EnqueueJob(job, mocks.JobQueueOptions{
		Priority:       70,
		CustomerTier:   85,
		RequiredTier:   70,
		RequiredRegion: "us-east",
	})
	s.Require().NoError(err)

	_, err = s.providerMock.ScheduleNext(ctx)
	s.Require().NoError(err)

	_, err = s.providerMock.StartJob(ctx, job.JobID)
	s.Require().NoError(err)

	s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateTimeout)
	s.providerMock.MarkTimeout(job.JobID)

	phases := s.providerMock.GetLifecycle(job.JobID)
	machine := NewLifecycleStateMachine([]mocks.LifecyclePhase{
		mocks.LifecycleSubmitted,
		mocks.LifecycleQueued,
		mocks.LifecycleScheduled,
		mocks.LifecycleRunning,
		mocks.LifecycleTimeout,
	})
	s.True(machine.Validate(phases))
}

func (s *HPCJobLifecycleE2ETestSuite) TestCancellationLifecycle() {
	ctx := context.Background()

	job := fixtures.QuickTestJob(s.providerAddr, s.customerAddr)
	job.JobID = "lifecycle-cancel-1"

	err := s.providerMock.EnqueueJob(job, mocks.JobQueueOptions{
		Priority:       75,
		CustomerTier:   85,
		RequiredTier:   70,
		RequiredRegion: "us-east",
	})
	s.Require().NoError(err)

	_, err = s.providerMock.ScheduleNext(ctx)
	s.Require().NoError(err)

	_, err = s.providerMock.StartJob(ctx, job.JobID)
	s.Require().NoError(err)

	err = s.slurmMock.CancelJob(ctx, job.JobID)
	s.Require().NoError(err)
	s.providerMock.MarkCancelled(job.JobID)

	phases := s.providerMock.GetLifecycle(job.JobID)
	machine := NewLifecycleStateMachine([]mocks.LifecyclePhase{
		mocks.LifecycleSubmitted,
		mocks.LifecycleQueued,
		mocks.LifecycleScheduled,
		mocks.LifecycleRunning,
		mocks.LifecycleCancelled,
	})
	s.True(machine.Validate(phases))
}
