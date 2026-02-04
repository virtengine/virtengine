//go:build e2e.integration

// Package e2e contains end-to-end integration tests.
//
// HPC-E2E-001: Scheduling algorithm tests.
package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/tests/e2e/fixtures"
	"github.com/virtengine/virtengine/tests/e2e/mocks"
	"github.com/virtengine/virtengine/testutil"
)

// HPCSchedulingE2ETestSuite validates scheduling decisions.
type HPCSchedulingE2ETestSuite struct {
	*testutil.NetworkTestSuite

	providerAddr string
	customerAddr string

	slurmMock    *mocks.MockSLURMIntegration
	providerMock *mocks.MockHPCProviderDaemon
}

func TestHPCSchedulingE2E(t *testing.T) {
	suite.Run(t, &HPCSchedulingE2ETestSuite{
		NetworkTestSuite: testutil.NewNetworkTestSuite(nil, &HPCSchedulingE2ETestSuite{}),
	})
}

func (s *HPCSchedulingE2ETestSuite) SetupSuite() {
	s.NetworkTestSuite.SetupSuite()
	val := s.Network().Validators[0]
	s.providerAddr = val.Address.String()
	s.customerAddr = val.Address.String()

	s.slurmMock = mocks.NewMockSLURMIntegration()
	s.providerMock = mocks.NewMockHPCProviderDaemon(s.slurmMock)

	s.providerMock.AddCluster(mocks.ProviderCluster{
		ClusterID:        "cluster-us-east",
		ProviderID:       "provider-us",
		Region:           "us-east",
		AvailableCPU:     8000,
		AvailableMemory:  20000,
		AvailableGPUs:    100,
		GPUType:          "nvidia-a100",
		LatencyScore:     0.95,
		PriceScore:       0.7,
		IdentityTier:     90,
		SupportsGPUTypes: []string{"nvidia-a100", "nvidia-v100"},
	})

	s.providerMock.AddCluster(mocks.ProviderCluster{
		ClusterID:        "cluster-eu-west",
		ProviderID:       "provider-eu",
		Region:           "eu-west",
		AvailableCPU:     6000,
		AvailableMemory:  15000,
		AvailableGPUs:    80,
		GPUType:          "nvidia-v100",
		LatencyScore:     0.85,
		PriceScore:       0.9,
		IdentityTier:     80,
		SupportsGPUTypes: []string{"nvidia-v100"},
	})

	s.providerMock.AddCluster(mocks.ProviderCluster{
		ClusterID:        "cluster-apac",
		ProviderID:       "provider-apac",
		Region:           "ap-south",
		AvailableCPU:     4000,
		AvailableMemory:  10000,
		AvailableGPUs:    20,
		GPUType:          "nvidia-t4",
		LatencyScore:     0.6,
		PriceScore:       0.95,
		IdentityTier:     70,
		SupportsGPUTypes: []string{"nvidia-t4"},
	})
}

func (s *HPCSchedulingE2ETestSuite) TearDownSuite() {
	s.NetworkTestSuite.TearDownSuite()
}

func (s *HPCSchedulingE2ETestSuite) TestProviderSelectionAlgorithm() {
	ctx := context.Background()

	job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
	job.JobID = "sched-select-1"

	s.Require().NoError(s.providerMock.EnqueueJob(job, mocks.JobQueueOptions{Priority: 70, CustomerTier: 85}))
	decision, err := s.providerMock.ScheduleNext(ctx)
	s.Require().NoError(err)

	s.NotEmpty(decision.SelectedClusterID)
	_, ok := decision.Scores[decision.SelectedClusterID]
	s.True(ok)
}

func (s *HPCSchedulingE2ETestSuite) TestGeographicConstraints() {
	ctx := context.Background()

	job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
	job.JobID = "sched-region-1"

	s.Require().NoError(s.providerMock.EnqueueJob(job, mocks.JobQueueOptions{
		Priority:       80,
		CustomerTier:   85,
		RequiredRegion: "eu-west",
	}))

	decision, err := s.providerMock.ScheduleNext(ctx)
	s.Require().NoError(err)
	s.Equal("cluster-eu-west", decision.SelectedClusterID)
}

func (s *HPCSchedulingE2ETestSuite) TestResourceAvailabilityChecks() {
	ctx := context.Background()

	job := fixtures.GPUComputeJob(s.providerAddr, s.customerAddr)
	job.JobID = "sched-resource-1"
	job.Resources.GPUsPerNode = 200

	s.Require().NoError(s.providerMock.EnqueueJob(job, mocks.JobQueueOptions{Priority: 90, CustomerTier: 90}))
	_, err := s.providerMock.ScheduleNext(ctx)
	s.Error(err)
}

func (s *HPCSchedulingE2ETestSuite) TestVEIDTierRequirements() {
	ctx := context.Background()

	job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
	job.JobID = "sched-tier-1"

	s.Require().NoError(s.providerMock.EnqueueJob(job, mocks.JobQueueOptions{
		Priority:     60,
		CustomerTier: 65,
		RequiredTier: 80,
	}))

	_, err := s.providerMock.ScheduleNext(ctx)
	s.Error(err)
}

func (s *HPCSchedulingE2ETestSuite) TestAllowedRegionsFallback() {
	ctx := context.Background()

	job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
	job.JobID = "sched-region-fallback"

	s.Require().NoError(s.providerMock.EnqueueJob(job, mocks.JobQueueOptions{
		Priority:       70,
		CustomerTier:   85,
		RequiredRegion: "us-west",
		AllowedRegions: []string{"eu-west", "ap-south"},
	}))

	decision, err := s.providerMock.ScheduleNext(ctx)
	s.Require().NoError(err)
	s.Contains([]string{"cluster-eu-west", "cluster-apac"}, decision.SelectedClusterID)
}
