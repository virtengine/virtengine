//go:build e2e.integration

// Package e2e contains end-to-end integration tests.
//
// HPC-E2E-001: Queue management tests.
package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/tests/e2e/fixtures"
	"github.com/virtengine/virtengine/tests/e2e/mocks"
	"github.com/virtengine/virtengine/testutil"
)

// HPCQueueE2ETestSuite validates queue ordering and limits.
type HPCQueueE2ETestSuite struct {
	*testutil.NetworkTestSuite

	providerAddr string
	customer1    string
	customer2    string

	slurmMock    *mocks.MockSLURMIntegration
	providerMock *mocks.MockHPCProviderDaemon
}

func TestHPCQueueE2E(t *testing.T) {
	suite.Run(t, &HPCQueueE2ETestSuite{
		NetworkTestSuite: testutil.NewNetworkTestSuite(nil, &HPCQueueE2ETestSuite{}),
	})
}

func (s *HPCQueueE2ETestSuite) SetupSuite() {
	s.NetworkTestSuite.SetupSuite()
	val := s.Network().Validators[0]
	s.providerAddr = val.Address.String()
	s.customer1 = val.Address.String()
	s.customer2 = "customer-queue-2"

	s.slurmMock = mocks.NewMockSLURMIntegration()
	s.providerMock = mocks.NewMockHPCProviderDaemon(s.slurmMock)

	cluster := mocks.DefaultTestCluster()
	s.slurmMock.RegisterCluster(cluster)
	s.providerMock.AddCluster(mocks.ProviderCluster{
		ClusterID:        cluster.ClusterID,
		ProviderID:       "provider-queue",
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

func (s *HPCQueueE2ETestSuite) TearDownSuite() {
	s.NetworkTestSuite.TearDownSuite()
}

func (s *HPCQueueE2ETestSuite) TestPriorityOrdering() {
	ctx := context.Background()

	jobLow := fixtures.QuickTestJob(s.providerAddr, s.customer1)
	jobLow.JobID = "queue-priority-low"
	jobMid := fixtures.QuickTestJob(s.providerAddr, s.customer1)
	jobMid.JobID = "queue-priority-mid"
	jobHigh := fixtures.QuickTestJob(s.providerAddr, s.customer1)
	jobHigh.JobID = "queue-priority-high"

	s.Require().NoError(s.providerMock.EnqueueJob(jobLow, mocks.JobQueueOptions{Priority: 10, CustomerTier: 70}))
	s.Require().NoError(s.providerMock.EnqueueJob(jobMid, mocks.JobQueueOptions{Priority: 50, CustomerTier: 70}))
	s.Require().NoError(s.providerMock.EnqueueJob(jobHigh, mocks.JobQueueOptions{Priority: 90, CustomerTier: 70}))

	decision, err := s.providerMock.ScheduleNext(ctx)
	s.Require().NoError(err)
	s.Equal(jobHigh.JobID, decision.JobID)

	decision, err = s.providerMock.ScheduleNext(ctx)
	s.Require().NoError(err)
	s.Equal(jobMid.JobID, decision.JobID)

	decision, err = s.providerMock.ScheduleNext(ctx)
	s.Require().NoError(err)
	s.Equal(jobLow.JobID, decision.JobID)
}

func (s *HPCQueueE2ETestSuite) TestFairShareScheduling() {
	ctx := context.Background()

	s.Require().NoError(s.slurmMock.Start(ctx))
	s.providerMock.EnableFairShare(true)

	jobA1 := fixtures.QuickTestJob(s.providerAddr, s.customer1)
	jobA1.JobID = "fairshare-a1"
	jobA2 := fixtures.QuickTestJob(s.providerAddr, s.customer1)
	jobA2.JobID = "fairshare-a2"
	jobB1 := fixtures.QuickTestJob(s.providerAddr, s.customer2)
	jobB1.JobID = "fairshare-b1"

	s.Require().NoError(s.providerMock.EnqueueJob(jobA1, mocks.JobQueueOptions{Priority: 50, CustomerTier: 80}))
	s.Require().NoError(s.providerMock.EnqueueJob(jobA2, mocks.JobQueueOptions{Priority: 50, CustomerTier: 80}))
	s.Require().NoError(s.providerMock.EnqueueJob(jobB1, mocks.JobQueueOptions{Priority: 50, CustomerTier: 80}))

	decision1, err := s.providerMock.ScheduleNext(ctx)
	s.Require().NoError(err)
	_, err = s.providerMock.StartJob(ctx, decision1.JobID)
	s.Require().NoError(err)
	decision2, err := s.providerMock.ScheduleNext(ctx)
	s.Require().NoError(err)

	s.NotEqual(decision1.JobID, decision2.JobID)
}

func (s *HPCQueueE2ETestSuite) TestResourceMatchingAndQueueDepth() {
	ctx := context.Background()

	s.providerMock.SetQueueDepthLimit(2)

	job1 := fixtures.StandardComputeJob(s.providerAddr, s.customer1)
	job1.JobID = "queue-depth-1"
	job2 := fixtures.StandardComputeJob(s.providerAddr, s.customer1)
	job2.JobID = "queue-depth-2"
	job3 := fixtures.StandardComputeJob(s.providerAddr, s.customer1)
	job3.JobID = "queue-depth-3"

	s.Require().NoError(s.providerMock.EnqueueJob(job1, mocks.JobQueueOptions{Priority: 50, CustomerTier: 70}))
	s.Require().NoError(s.providerMock.EnqueueJob(job2, mocks.JobQueueOptions{Priority: 50, CustomerTier: 70}))
	s.Error(s.providerMock.EnqueueJob(job3, mocks.JobQueueOptions{Priority: 50, CustomerTier: 70}))

	decision, err := s.providerMock.ScheduleNext(ctx)
	s.Require().NoError(err)
	s.Equal(job1.JobID, decision.JobID)
}

func (s *HPCQueueE2ETestSuite) TestGPUResourceMatching() {
	ctx := context.Background()

	jobGPU := fixtures.GPUComputeJob(s.providerAddr, s.customer1)
	jobGPU.JobID = "queue-gpu-1"

	s.Require().NoError(s.providerMock.EnqueueJob(jobGPU, mocks.JobQueueOptions{Priority: 80, CustomerTier: 90}))
	decision, err := s.providerMock.ScheduleNext(ctx)
	s.Require().NoError(err)
	s.Equal(jobGPU.JobID, decision.JobID)
	s.NotEmpty(decision.SelectedClusterID)
}
