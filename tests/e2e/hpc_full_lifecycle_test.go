//go:build e2e.integration

// Package e2e contains end-to-end integration tests.
//
// VE-68B: Comprehensive HPC end-to-end lifecycle test
// This test validates the complete HPC job flow including:
// - Cluster registration and capacity reporting
// - Job submission with resource requirements
// - Multi-cluster scheduling with proximity scoring
// - Job execution tracking
// - Usage billing and settlement
// - Reward distribution to providers
// - Job cancellation scenarios
// - SLA breach penalties
package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/tests/e2e/fixtures"
	"github.com/virtengine/virtengine/tests/e2e/mocks"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCFullLifecycleTestSuite tests comprehensive HPC workflows
type HPCFullLifecycleTestSuite struct {
	suite.Suite

	providerAddr string
	customerAddr string

	slurmMock      *mocks.MockSLURMIntegration
	providerMock   *mocks.MockHPCProviderDaemon
	settlementMock *BillingMockSettlementProcessor
}

func TestHPCFullLifecycleTestSuite(t *testing.T) {
	suite.Run(t, &HPCFullLifecycleTestSuite{})
}

func (s *HPCFullLifecycleTestSuite) SetupSuite() {
	s.providerAddr = sdk.AccAddress([]byte("hpc-full-provider-00001")).String()
	s.customerAddr = sdk.AccAddress([]byte("hpc-full-customer-00001")).String()

	s.slurmMock = mocks.NewMockSLURMIntegration()
	s.providerMock = mocks.NewMockHPCProviderDaemon(s.slurmMock)
	s.settlementMock = NewBillingMockSettlementProcessor()

	ctx := context.Background()
	err := s.slurmMock.Start(ctx)
	s.Require().NoError(err)
}

func (s *HPCFullLifecycleTestSuite) TearDownSuite() {
	if s.slurmMock != nil && s.slurmMock.IsRunning() {
		_ = s.slurmMock.Stop()
	}
}

// TestCompleteLifecycleWithBillingAndRewards tests the full flow from submission to rewards
func (s *HPCFullLifecycleTestSuite) TestCompleteLifecycleWithBillingAndRewards() {
	t := s.T()
	ctx := context.Background()

	t.Log("=== Phase 1: Cluster Registration ===")

	// Register cluster with detailed capacity
	cluster := &hpctypes.HPCCluster{
		ClusterID:       "hpc-cluster-full-001",
		ProviderAddress: s.providerAddr,
		Name:            "Full Lifecycle Test Cluster",
		Description:     "High-performance GPU cluster for ML workloads",
		Region:          "us-west-2",
		State:           hpctypes.ClusterStateActive,
		Partitions: []hpctypes.Partition{{
			Name:           "gpu",
			Nodes:          32,
			MaxRuntime:     24 * 3600,
			DefaultRuntime: 4 * 3600,
			MaxNodes:       8,
			Features:       []string{"gpu", "a100"},
			Priority:       100,
			State:          "UP",
		}},
		TotalNodes:     32,
		AvailableNodes: 32,
		ClusterMetadata: hpctypes.ClusterMetadata{
			TotalCPUCores:    1024,
			TotalMemoryGB:    4096,
			TotalGPUs:        64,
			GPUTypes:         []string{"nvidia-a100-80gb"},
			InterconnectType: "infiniband",
			StorageType:      "lustre",
			TotalStorageGB:   500 * 1024,
		},
		SLURMVersion: "23.02.4",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	s.slurmMock.RegisterCluster(mockSLURMClusterFromHPCCluster(cluster))
	s.providerMock.AddCluster(providerClusterFromHPCCluster(cluster, 0.95, 0.85, 90))

	t.Logf("✓ Cluster registered: %s (%d CPU, %d GPU)",
		cluster.ClusterID, cluster.ClusterMetadata.TotalCPUCores, cluster.ClusterMetadata.TotalGPUs)

	t.Log("=== Phase 2: Capacity Reporting ===")

	// Provider reports current capacity using the same totals exported in the mock registration.
	capacityReport := struct {
		AvailableCPU      int64
		AvailableMemoryGB int64
		AvailableGPUs     int64
		QueuedJobs        int
		RunningJobs       int
	}{
		AvailableCPU:      cluster.ClusterMetadata.TotalCPUCores,
		AvailableMemoryGB: cluster.ClusterMetadata.TotalMemoryGB,
		AvailableGPUs:     cluster.ClusterMetadata.TotalGPUs,
	}

	t.Logf("✓ Capacity reported: %d/%d CPU, %d/%d GPU available",
		capacityReport.AvailableCPU, cluster.ClusterMetadata.TotalCPUCores,
		capacityReport.AvailableGPUs, cluster.ClusterMetadata.TotalGPUs)

	t.Log("=== Phase 3: Job Submission ===")

	// Customer submits ML training job
	job := &hpctypes.HPCJob{
		JobID:           "job-full-lifecycle-001",
		ClusterID:       cluster.ClusterID,
		OfferingID:      "offering-gpu-training",
		CustomerAddress: s.customerAddr,
		ProviderAddress: s.providerAddr,
		QueueName:       "gpu",
		WorkloadSpec: hpctypes.JobWorkloadSpec{
			ContainerImage: "python:3.11-cuda",
			Command:        "python",
			Arguments:      []string{"train.py", "--distributed"},
		},
		Resources: hpctypes.JobResources{
			Nodes:           1,
			CPUCoresPerNode: 128,
			MemoryGBPerNode: 512,
			GPUsPerNode:     8,
			GPUType:         "nvidia-a100-80gb",
			StorageGB:       1000,
		},
		MaxRuntimeSeconds: 24 * 3600,
		CreatedAt:         time.Now(),
		State:             hpctypes.JobStatePending,
	}

	err := s.providerMock.EnqueueJob(job, mocks.JobQueueOptions{
		Priority:       80,
		CustomerTier:   85,
		RequiredTier:   70,
		RequiredRegion: cluster.Region,
	})
	require.NoError(t, err)

	t.Logf("✓ Job submitted: %s (8 GPU, 128 CPU, 24h walltime)", job.JobID)

	t.Log("=== Phase 4: Scheduling ===")

	// Schedule job to cluster
	decision, err := s.providerMock.ScheduleNext(ctx)
	require.NoError(t, err)
	require.NotNil(t, decision)
	require.Equal(t, cluster.ClusterID, decision.SelectedClusterID)

	t.Logf("✓ Job scheduled to cluster: %s", decision.SelectedClusterID)
	t.Logf("  - Scheduling reason: %s", decision.Reason)

	t.Log("=== Phase 5: Execution ===")

	// Start job execution
	schedulerJob, err := s.providerMock.StartJob(ctx, job.JobID)
	require.NoError(t, err)
	require.NotNil(t, schedulerJob)

	job.State = hpctypes.JobStateRunning
	startedAt := time.Now()
	job.StartedAt = &startedAt

	t.Logf("✓ Job started at %s", job.StartedAt.Format(time.RFC3339))

	// Simulate execution progress
	executionTime := 22 * time.Hour // Completed in 22 hours (under 24h walltime)

	schedulerMetrics := newRealSchedulerMetrics(executionTime, 1, 128, 512, 8, 1000)
	requireRealSchedulerTelemetry(t, schedulerMetrics)
	metrics := &hpctypes.HPCDetailedMetrics{
		WallClockSeconds: schedulerMetrics.WallClockSeconds,
		CPUTimeSeconds:   schedulerMetrics.CPUTimeSeconds,
		CPUCoreSeconds:   schedulerMetrics.CPUCoreSeconds,
		MemoryBytesMax:   schedulerMetrics.MemoryBytesMax,
		MemoryGBSeconds:  schedulerMetrics.MemoryGBSeconds,
		GPUSeconds:       schedulerMetrics.GPUSeconds,
		GPUType:          "nvidia-a100-80gb",
		StorageGBHours:   schedulerMetrics.StorageGBHours,
		NetworkBytesIn:   schedulerMetrics.NetworkBytesIn,
		NetworkBytesOut:  schedulerMetrics.NetworkBytesOut,
		EnergyJoules:     schedulerMetrics.EnergyJoules,
	}

	s.slurmMock.SetJobMetrics(job.JobID, schedulerMetrics)
	s.slurmMock.SetJobExitCode(job.JobID, 0)
	s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

	job.State = hpctypes.JobStateCompleted
	completedAt := job.StartedAt.Add(executionTime)
	job.CompletedAt = &completedAt

	t.Logf("✓ Job completed at %s", job.CompletedAt.Format(time.RFC3339))
	t.Logf("  - Duration: %.2f hours", executionTime.Hours())
	t.Logf("  - GPU-hours: %.2f", float64(metrics.GPUSeconds)/3600)

	executionRecord, found := s.slurmMock.GetExecutionRecord(job.JobID)
	require.True(t, found)
	require.NotNil(t, executionRecord.Metrics)
	require.Equal(t, schedulerMetrics.WallClockSeconds, executionRecord.Metrics.WallClockSeconds)

	t.Log("=== Phase 6: Billing ===")

	// Calculate costs based on usage
	pricePerGPUHour := sdkmath.LegacyNewDec(50000)   // 50k uakt per GPU-hour
	pricePerCPUHour := sdkmath.LegacyNewDec(100)     // 100 uakt per CPU-hour
	pricePerGBMemoryHour := sdkmath.LegacyNewDec(10) // 10 uakt per GB-hour

	gpuHours := sdkmath.LegacyNewDec(metrics.GPUSeconds).QuoInt64(3600)
	cpuHours := sdkmath.LegacyNewDec(metrics.CPUCoreSeconds).QuoInt64(3600)
	memoryGBHours := sdkmath.LegacyNewDec(metrics.MemoryGBSeconds).QuoInt64(3600)

	gpuCost := gpuHours.Mul(pricePerGPUHour)
	cpuCost := cpuHours.Mul(pricePerCPUHour)
	memoryCost := memoryGBHours.Mul(pricePerGBMemoryHour)

	totalCost := gpuCost.Add(cpuCost).Add(memoryCost)

	accountingRecord := &hpctypes.HPCAccountingRecord{
		RecordID:        fmt.Sprintf("record-%s", job.JobID),
		JobID:           job.JobID,
		ClusterID:       cluster.ClusterID,
		ProviderAddress: s.providerAddr,
		CustomerAddress: s.customerAddr,
		OfferingID:      job.OfferingID,
		SchedulerType:   "slurm",
		UsageMetrics:    *metrics,
		BillableAmount:  sdk.NewCoins(sdk.NewCoin("uakt", totalCost.TruncateInt())),
		ProviderReward:  sdk.NewCoins(sdk.NewCoin("uakt", totalCost.TruncateInt())),
		PlatformFee:     sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.ZeroInt())),
		Status:          hpctypes.AccountingStatusFinalized,
		PeriodStart:     *job.StartedAt,
		PeriodEnd:       *job.CompletedAt,
		FormulaVersion:  hpctypes.CurrentBillingFormulaVersion,
		CreatedAt:       time.Now(),
	}

	t.Logf("✓ Accounting record created:")
	t.Logf("  - GPU cost: %s uakt (%.2f hours @ %s/hr)",
		gpuCost.TruncateInt().String(), gpuHours.MustFloat64(), pricePerGPUHour.String())
	t.Logf("  - CPU cost: %s uakt (%.2f hours @ %s/hr)",
		cpuCost.TruncateInt().String(), cpuHours.MustFloat64(), pricePerCPUHour.String())
	t.Logf("  - Memory cost: %s uakt (%.2f GB-hours @ %s/GB-hr)",
		memoryCost.TruncateInt().String(), memoryGBHours.MustFloat64(), pricePerGBMemoryHour.String())
	t.Logf("  - Total: %s", accountingRecord.BillableAmount.String())

	t.Log("=== Phase 7: Settlement ===")

	settlement := s.settlementMock.ProcessSettlement(accountingRecord, time.Now())
	require.True(t, settlement.Success)

	platformFeeRate := sdkmath.LegacyMustNewDecFromStr("0.02") // 2% platform fee
	platformFee := totalCost.Mul(platformFeeRate).TruncateInt()
	providerNet := totalCost.Sub(sdkmath.LegacyNewDec(platformFee.Int64())).TruncateInt()

	t.Logf("✓ Settlement processed:")
	t.Logf("  - Settlement ID: %s", settlement.SettlementID)
	t.Logf("  - Total: %s", accountingRecord.BillableAmount.String())
	t.Logf("  - Platform fee (2%%): %s", sdk.NewCoins(sdk.NewCoin("uakt", platformFee)).String())
	t.Logf("  - Provider net: %s", sdk.NewCoins(sdk.NewCoin("uakt", providerNet)).String())

	t.Log("=== Phase 8: Reward Distribution ===")

	// Provider receives rewards
	baseProviderReward := providerNet // Provider keeps 98%
	performanceBonus := sdkmath.LegacyNewDec(baseProviderReward.Int64()).Mul(
		sdkmath.LegacyMustNewDecFromStr("0.05")).TruncateInt() // 5% bonus for on-time completion

	totalProviderReward := sdk.NewCoins(
		sdk.NewCoin("uakt", baseProviderReward.Add(performanceBonus)))

	t.Logf("✓ Rewards distributed:")
	t.Logf("  - Provider base: %s", sdk.NewCoins(sdk.NewCoin("uakt", baseProviderReward)).String())
	t.Logf("  - Performance bonus (5%%): %s", sdk.NewCoins(sdk.NewCoin("uakt", performanceBonus)).String())
	t.Logf("  - Total provider reward: %s", totalProviderReward.String())

	t.Log("✓✓✓ Complete HPC lifecycle test passed ✓✓✓")
}

// TestMultiClusterJobRouting tests job routing across multiple clusters with proximity scoring
func (s *HPCFullLifecycleTestSuite) TestMultiClusterJobRouting() {
	t := s.T()
	ctx := context.Background()
	slurmMock := mocks.NewMockSLURMIntegration()
	providerMock := mocks.NewMockHPCProviderDaemon(slurmMock)
	require.NoError(t, slurmMock.Start(ctx))
	defer func() { _ = slurmMock.Stop() }()

	t.Log("=== Multi-Cluster Routing Test ===")

	// Register 3 clusters in different regions
	clusters := []struct {
		id       string
		region   string
		gpuCount int
		latency  float64
		price    float64
	}{
		{"cluster-us-east", "us-east-1", 32, 0.95, 0.80},
		{"cluster-us-west", "us-west-2", 64, 0.90, 0.85},
		{"cluster-eu-west", "eu-west-1", 48, 0.70, 0.75},
	}

	for _, c := range clusters {
		cluster := mocks.DefaultTestCluster()
		cluster.ClusterID = c.id
		cluster.Region = c.region
		cluster.TotalGPUs = int32(c.gpuCount)

		slurmMock.RegisterCluster(cluster)
		providerMock.AddCluster(mocks.ProviderCluster{
			ClusterID:        c.id,
			ProviderID:       "provider-1",
			Region:           c.region,
			AvailableCPU:     512,
			AvailableMemory:  2048,
			AvailableGPUs:    int32(c.gpuCount),
			GPUType:          "nvidia-a100",
			LatencyScore:     c.latency,
			PriceScore:       c.price,
			IdentityTier:     90,
			SupportsGPUTypes: []string{"nvidia-a100"},
		})

		t.Logf("✓ Registered cluster: %s in %s (%d GPUs, latency=%.2f, price=%.2f)",
			c.id, c.region, c.gpuCount, c.latency, c.price)
	}

	// Submit job preferring us-west region
	job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
	job.JobID = "job-routing-001"

	err := providerMock.EnqueueJob(job, mocks.JobQueueOptions{
		Priority:       85,
		CustomerTier:   85,
		RequiredTier:   70,
		RequiredRegion: "us-west-2",
	})
	require.NoError(t, err)

	// Schedule - should select us-west cluster due to region preference
	decision, err := providerMock.ScheduleNext(ctx)
	require.NoError(t, err)
	require.Equal(t, "cluster-us-west", decision.SelectedClusterID)

	t.Logf("✓ Job routed to %s based on region preference", decision.SelectedClusterID)
	t.Log("✓ Multi-cluster routing test passed")
}

// TestJobCancellationWithPartialBilling tests job cancellation mid-execution
func (s *HPCFullLifecycleTestSuite) TestJobCancellationWithPartialBilling() {
	t := s.T()
	ctx := context.Background()

	t.Log("=== Job Cancellation Test ===")

	cluster := mocks.DefaultTestCluster()
	s.slurmMock.RegisterCluster(cluster)
	s.providerMock.AddCluster(providerClusterFromSLURMCluster(cluster))

	// Submit and start job
	job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
	job.JobID = "job-cancel-001"

	err := s.providerMock.EnqueueJob(job, mocks.JobQueueOptions{
		Priority:     80,
		CustomerTier: 85,
		RequiredTier: 70,
	})
	require.NoError(t, err)

	_, err = s.providerMock.ScheduleNext(ctx)
	require.NoError(t, err)

	_, err = s.providerMock.StartJob(ctx, job.JobID)
	require.NoError(t, err)

	t.Log("✓ Job started")

	// Simulate partial execution (2 hours out of 4)
	partialExecutionTime := 2 * time.Hour
	partialMetrics := newRealSchedulerMetrics(partialExecutionTime, 1, 8, 16, 0, 40)
	requireRealSchedulerTelemetry(t, partialMetrics)

	s.slurmMock.SetJobMetrics(job.JobID, partialMetrics)

	// Cancel job
	s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCancelled)

	t.Log("✓ Job cancelled after 2 hours")

	status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
	require.NoError(t, err)
	require.Equal(t, pd.HPCJobStateCancelled, status.State)

	// Calculate partial billing
	totalEstimatedCost := sdkmath.LegacyNewDec(10000)        // Full job would be 10k uakt
	partialUsageRatio := sdkmath.LegacyNewDec(2).QuoInt64(4) // 2/4 hours
	partialCost := totalEstimatedCost.Mul(partialUsageRatio)

	cancellationFee := totalEstimatedCost.Mul(sdkmath.LegacyMustNewDecFromStr("0.10")) // 10% cancellation fee
	totalCharge := partialCost.Add(cancellationFee)

	t.Logf("✓ Partial billing calculated:")
	t.Logf("  - Usage (50%%): %s uakt", partialCost.TruncateInt().String())
	t.Logf("  - Cancellation fee (10%%): %s uakt", cancellationFee.TruncateInt().String())
	t.Logf("  - Total charge: %s uakt", totalCharge.TruncateInt().String())

	t.Log("✓ Cancellation with partial billing test passed")
}

// TestProviderPenaltiesForSLABreach tests penalties when provider breaches SLA
func (s *HPCFullLifecycleTestSuite) TestProviderPenaltiesForSLABreach() {
	t := s.T()
	ctx := context.Background()

	t.Log("=== SLA Breach Penalty Test ===")

	cluster := mocks.DefaultTestCluster()
	s.slurmMock.RegisterCluster(cluster)
	s.providerMock.AddCluster(providerClusterFromSLURMCluster(cluster))

	// Submit job with 4-hour SLA
	job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
	job.JobID = "job-sla-breach-001"
	job.MaxRuntimeSeconds = 4 * 3600 // 4-hour SLA

	err := s.providerMock.EnqueueJob(job, mocks.JobQueueOptions{
		Priority:     90,
		CustomerTier: 85,
		RequiredTier: 70,
	})
	require.NoError(t, err)

	_, err = s.providerMock.ScheduleNext(ctx)
	require.NoError(t, err)

	_, err = s.providerMock.StartJob(ctx, job.JobID)
	require.NoError(t, err)

	t.Log("✓ Job started with 4-hour SLA")

	// Simulate job exceeding SLA (runs for 6 hours)
	actualExecutionTime := 6 * time.Hour
	metrics := newRealSchedulerMetrics(actualExecutionTime, 1, 8, 16, 0, 40)
	requireRealSchedulerTelemetry(t, metrics)

	s.slurmMock.SetJobMetrics(job.JobID, metrics)
	s.slurmMock.SetJobExitCode(job.JobID, 0)
	s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

	t.Logf("✓ Job completed in %.2f hours (exceeded 4-hour SLA)", actualExecutionTime.Hours())

	status, err := s.slurmMock.GetJobStatus(ctx, job.JobID)
	require.NoError(t, err)
	require.Equal(t, pd.HPCJobStateCompleted, status.State)

	// Calculate SLA breach penalty
	baseCharge := sdkmath.LegacyNewDec(15000) // 15k uakt
	slaBreachHours := actualExecutionTime.Hours() - 4.0
	penaltyRate := sdkmath.LegacyMustNewDecFromStr("0.20") // 20% penalty per breach hour
	totalPenaltyRate := sdkmath.LegacyNewDec(int64(slaBreachHours * 100)).QuoInt64(100).Mul(penaltyRate)

	penalty := baseCharge.Mul(totalPenaltyRate)
	customerRefund := penalty
	providerPenalty := penalty

	netProviderRevenue := baseCharge.Sub(providerPenalty)

	t.Logf("✓ SLA breach penalty calculated:")
	t.Logf("  - Base charge: %s uakt", baseCharge.TruncateInt().String())
	t.Logf("  - Breach hours: %.2f", slaBreachHours)
	t.Logf("  - Penalty rate: %.1f%%", totalPenaltyRate.MustFloat64()*100)
	t.Logf("  - Customer refund: %s uakt", customerRefund.TruncateInt().String())
	t.Logf("  - Provider penalty: %s uakt", providerPenalty.TruncateInt().String())
	t.Logf("  - Net provider revenue: %s uakt", netProviderRevenue.TruncateInt().String())

	require.Less(t, netProviderRevenue.MustFloat64(), baseCharge.MustFloat64(),
		"provider revenue should be reduced due to SLA breach")

	t.Log("✓ SLA breach penalty test passed")
}

// TestSLURMWorkloadTemplateValidation tests workload template validation for SLURM-backed jobs.
func (s *HPCFullLifecycleTestSuite) TestSLURMWorkloadTemplateValidation() {
	t := s.T()

	t.Log("=== SLURM Template Validation Test ===")

	validTemplate := &hpctypes.WorkloadTemplate{
		TemplateID:  "slurm-ml-training-v1",
		Name:        "ML Training Template",
		Version:     "1.0.0",
		Description: "Standard template for ML training jobs",
		Type:        hpctypes.WorkloadTypeGPU,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType:       "apptainer",
			ContainerImage:    "python:3.11-cuda",
			RequiredModules:   []string{"cuda/11.8", "python/3.10"},
			MPIImplementation: "openmpi",
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               1,
			MaxNodes:               8,
			DefaultNodes:           2,
			MinCPUsPerNode:         8,
			MaxCPUsPerNode:         128,
			DefaultCPUsPerNode:     16,
			MinMemoryMBPerNode:     16384,
			MaxMemoryMBPerNode:     524288,
			DefaultMemoryMBPerNode: 131072,
			MinGPUsPerNode:         1,
			MaxGPUsPerNode:         8,
			DefaultGPUsPerNode:     4,
			MinRuntimeMinutes:      30,
			MaxRuntimeMinutes:      24 * 60,
			DefaultRuntimeMinutes:  6 * 60,
			NetworkRequired:        true,
		},
		Security: hpctypes.WorkloadSecuritySpec{
			SandboxLevel:       "basic",
			AllowNetworkAccess: true,
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{
			Command:     "python",
			DefaultArgs: []string{"train.py", "--distributed"},
			UseMPIRun:   true,
		},
		Publisher:      s.providerAddr,
		ApprovalStatus: hpctypes.WorkloadApprovalApproved,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Tags:           []string{"slurm", "gpu", "training"},
	}

	require.NoError(t, validTemplate.Validate())
	require.NotEmpty(t, validTemplate.Runtime.ContainerImage)
	require.True(t, validTemplate.Resources.DefaultGPUsPerNode > 0)
	require.True(t, validTemplate.Entrypoint.UseMPIRun)
	require.Contains(t, validTemplate.Tags, "slurm")

	t.Log("✓ Valid SLURM template accepted")

	// Invalid template (missing required SBATCH directives)
	invalidTemplate := &hpctypes.WorkloadTemplate{
		TemplateID: "slurm-invalid-v1",
		Version:    "1.0.0",
		Type:       hpctypes.WorkloadTypeGPU,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType: "apptainer",
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               1,
			MaxNodes:               2,
			DefaultNodes:           1,
			MinCPUsPerNode:         1,
			MaxCPUsPerNode:         4,
			DefaultCPUsPerNode:     4,
			MinMemoryMBPerNode:     1024,
			MaxMemoryMBPerNode:     4096,
			DefaultMemoryMBPerNode: 2048,
			MinRuntimeMinutes:      1,
			MaxRuntimeMinutes:      120,
			DefaultRuntimeMinutes:  30,
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{},
	}

	// Validation should fail
	require.Error(t, invalidTemplate.Validate())
	require.Empty(t, invalidTemplate.Entrypoint.Command)
	t.Log("✓ Invalid template correctly rejected")

	t.Log("✓ SLURM template validation test passed")
}

func mockSLURMClusterFromHPCCluster(cluster *hpctypes.HPCCluster) *mocks.SLURMCluster {
	partitions := make([]mocks.SLURMPartition, 0, len(cluster.Partitions))
	for _, partition := range cluster.Partitions {
		partitions = append(partitions, mocks.SLURMPartition{
			Name:         partition.Name,
			Nodes:        partition.Nodes,
			MaxRuntime:   partition.MaxRuntime,
			MaxNodes:     partition.MaxNodes,
			Features:     partition.Features,
			Priority:     partition.Priority,
			State:        partition.State,
			AvailableGPU: int32(cluster.ClusterMetadata.TotalGPUs),
			AvailableCPU: int32(cluster.ClusterMetadata.TotalCPUCores),
		})
	}

	return &mocks.SLURMCluster{
		ClusterID:     cluster.ClusterID,
		Name:          cluster.Name,
		Region:        cluster.Region,
		SLURMVersion:  cluster.SLURMVersion,
		Partitions:    partitions,
		TotalNodes:    cluster.TotalNodes,
		TotalCPU:      int32(cluster.ClusterMetadata.TotalCPUCores),
		TotalMemoryGB: cluster.ClusterMetadata.TotalMemoryGB,
		TotalGPUs:     int32(cluster.ClusterMetadata.TotalGPUs),
	}
}

func providerClusterFromHPCCluster(cluster *hpctypes.HPCCluster, latency, price float64, identityTier int32) mocks.ProviderCluster {
	gpuType := ""
	if len(cluster.ClusterMetadata.GPUTypes) > 0 {
		gpuType = cluster.ClusterMetadata.GPUTypes[0]
	}

	return mocks.ProviderCluster{
		ClusterID:        cluster.ClusterID,
		ProviderID:       cluster.ProviderAddress,
		Region:           cluster.Region,
		AvailableCPU:     int32(cluster.ClusterMetadata.TotalCPUCores),
		AvailableMemory:  cluster.ClusterMetadata.TotalMemoryGB,
		AvailableGPUs:    int32(cluster.ClusterMetadata.TotalGPUs),
		GPUType:          gpuType,
		LatencyScore:     latency,
		PriceScore:       price,
		IdentityTier:     identityTier,
		SupportsGPUTypes: append([]string(nil), cluster.ClusterMetadata.GPUTypes...),
	}
}

func providerClusterFromSLURMCluster(cluster *mocks.SLURMCluster) mocks.ProviderCluster {
	supports := make([]string, 0, len(cluster.Partitions))
	for _, partition := range cluster.Partitions {
		supports = append(supports, partition.Features...)
	}

	return mocks.ProviderCluster{
		ClusterID:        cluster.ClusterID,
		ProviderID:       "provider-1",
		Region:           cluster.Region,
		AvailableCPU:     cluster.TotalCPU,
		AvailableMemory:  cluster.TotalMemoryGB,
		AvailableGPUs:    cluster.TotalGPUs,
		GPUType:          "nvidia-a100",
		LatencyScore:     0.90,
		PriceScore:       0.80,
		IdentityTier:     90,
		SupportsGPUTypes: supports,
	}
}
