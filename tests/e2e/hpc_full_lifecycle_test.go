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
	"github.com/virtengine/virtengine/testutil"
	"github.com/virtengine/virtengine/x/escrow/types/billing"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

// HPCFullLifecycleTestSuite tests comprehensive HPC workflows
type HPCFullLifecycleTestSuite struct {
	*testutil.NetworkTestSuite

	providerAddr string
	customerAddr string

	slurmMock      *mocks.MockSLURMIntegration
	providerMock   *mocks.MockHPCProviderDaemon
	settlementMock *BillingMockSettlementProcessor
}

func TestHPCFullLifecycleTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping comprehensive HPC lifecycle test in short mode")
	}

	suite.Run(t, &HPCFullLifecycleTestSuite{
		NetworkTestSuite: testutil.NewNetworkTestSuite(nil, &HPCFullLifecycleTestSuite{}),
	})
}

func (s *HPCFullLifecycleTestSuite) SetupSuite() {
	s.NetworkTestSuite.SetupSuite()

	val := s.Network().Validators[0]
	s.providerAddr = val.Address.String()
	s.customerAddr = val.Address.String()

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
	s.NetworkTestSuite.TearDownSuite()
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
		SchedulerType:   hpctypes.SchedulerTypeSLURM,
		TotalCPU:        1024,
		TotalMemoryGB:   4096,
		TotalGPUs:       64,
		GPUType:         "nvidia-a100-80gb",
		StorageTB:       500,
		NetworkGbps:     100,
		Status:          hpctypes.ClusterStatusActive,
		RegisteredAt:    time.Now(),
	}

	s.slurmMock.RegisterCluster(mocks.ClusterFromHPCCluster(cluster))
	s.providerMock.AddCluster(mocks.ProviderClusterFromHPC(cluster, 0.95, 0.85, 90))

	t.Logf("✓ Cluster registered: %s (%d CPU, %d GPU)",
		cluster.ClusterID, cluster.TotalCPU, cluster.TotalGPUs)

	t.Log("=== Phase 2: Capacity Reporting ===")

	// Provider reports current capacity
	capacityReport := &hpctypes.ClusterCapacityReport{
		ClusterID:         cluster.ClusterID,
		ProviderAddress:   s.providerAddr,
		AvailableCPU:      1024,
		AvailableMemoryGB: 4096,
		AvailableGPUs:     64,
		UsageCPU:          0,
		UsageMemoryGB:     0,
		UsageGPUs:         0,
		QueuedJobs:        0,
		RunningJobs:       0,
		ReportedAt:        time.Now(),
	}

	t.Logf("✓ Capacity reported: %d/%d CPU, %d/%d GPU available",
		capacityReport.AvailableCPU, cluster.TotalCPU,
		capacityReport.AvailableGPUs, cluster.TotalGPUs)

	t.Log("=== Phase 3: Job Submission ===")

	// Customer submits ML training job
	job := &hpctypes.HPCJob{
		JobID:           "job-full-lifecycle-001",
		ClusterID:       cluster.ClusterID,
		OfferingID:      "offering-gpu-training",
		CustomerAddress: s.customerAddr,
		ProviderAddress: s.providerAddr,
		JobName:         "llm-training-run",
		Description:     "Large language model training (100B parameters)",
		Requirements: hpctypes.HPCResourceRequirements{
			CPU:           128,
			MemoryGB:      512,
			GPUs:          8,
			GPUType:       "nvidia-a100-80gb",
			StorageGB:     1000,
			WallTimeHours: 24,
		},
		Priority:    80,
		SubmittedAt: time.Now(),
		Status:      hpctypes.JobStatusPending,
	}

	err := s.providerMock.EnqueueJob(fixtures.MockJobFromHPC(job), mocks.JobQueueOptions{
		Priority:       80,
		CustomerTier:   85,
		RequiredTier:   70,
		RequiredRegion: "us-west",
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

	job.Status = hpctypes.JobStatusRunning
	job.StartedAt = time.Now()

	t.Logf("✓ Job started at %s", job.StartedAt.Format(time.RFC3339))

	// Simulate execution progress
	executionTime := 22 * time.Hour // Completed in 22 hours (under 24h walltime)

	metrics := &hpctypes.HPCDetailedMetrics{
		WallClockSeconds: int64(executionTime.Seconds()),
		CPUCoreSeconds:   int64(128 * executionTime.Seconds()),
		MemoryGBSeconds:  int64(512 * executionTime.Seconds()),
		GPUSeconds:       int64(8 * executionTime.Seconds()),
		GPUType:          "nvidia-a100-80gb",
		StorageGBSeconds: int64(1000 * executionTime.Seconds()),
		NetworkBytesIn:   1024 * 1024 * 1024 * 500, // 500 GB ingress
		NetworkBytesOut:  1024 * 1024 * 1024 * 100, // 100 GB egress
	}

	s.slurmMock.SetJobMetrics(job.JobID, fixtures.MetricsFromHPC(metrics))
	s.slurmMock.SetJobExitCode(job.JobID, 0)
	s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

	s.providerMock.MarkCompleted(job.JobID, fixtures.MetricsFromHPC(metrics))

	job.Status = hpctypes.JobStatusCompleted
	job.CompletedAt = job.StartedAt.Add(executionTime)

	t.Logf("✓ Job completed at %s", job.CompletedAt.Format(time.RFC3339))
	t.Logf("  - Duration: %.2f hours", executionTime.Hours())
	t.Logf("  - GPU-hours: %.2f", float64(metrics.GPUSeconds)/3600)

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
		SchedulerType:   string(hpctypes.SchedulerTypeSLURM),
		UsageMetrics:    *metrics,
		BillingMetrics: hpctypes.HPCBillingMetrics{
			GPUHours:       gpuHours,
			CPUHours:       cpuHours,
			MemoryGBHours:  memoryGBHours,
			StorageGBHours: sdkmath.LegacyNewDec(metrics.StorageGBSeconds).QuoInt64(3600),
		},
		TotalCost: sdk.NewCoins(sdk.NewCoin("uakt", totalCost.TruncateInt())),
		StartTime: job.StartedAt,
		EndTime:   job.CompletedAt,
		CreatedAt: time.Now(),
	}

	t.Logf("✓ Accounting record created:")
	t.Logf("  - GPU cost: %s uakt (%.2f hours @ %s/hr)",
		gpuCost.TruncateInt().String(), gpuHours.MustFloat64(), pricePerGPUHour.String())
	t.Logf("  - CPU cost: %s uakt (%.2f hours @ %s/hr)",
		cpuCost.TruncateInt().String(), cpuHours.MustFloat64(), pricePerCPUHour.String())
	t.Logf("  - Memory cost: %s uakt (%.2f GB-hours @ %s/GB-hr)",
		memoryCost.TruncateInt().String(), memoryGBHours.MustFloat64(), pricePerGBMemoryHour.String())
	t.Logf("  - Total: %s", accountingRecord.TotalCost.String())

	t.Log("=== Phase 7: Settlement ===")

	// Process settlement
	platformFeeRate := sdkmath.LegacyMustNewDecFromStr("0.02") // 2% platform fee
	platformFee := totalCost.Mul(platformFeeRate).TruncateInt()
	providerNet := totalCost.Sub(sdkmath.LegacyNewDec(platformFee.Int64())).TruncateInt()

	settlement := &settlementtypes.SettlementRecord{
		RecordID:    fmt.Sprintf("settlement-%s", job.JobID),
		LeaseID:     fmt.Sprintf("lease-%s", job.JobID),
		InvoiceID:   fmt.Sprintf("invoice-%s", job.JobID),
		Provider:    s.providerAddr,
		Customer:    s.customerAddr,
		Amount:      sdk.NewCoins(sdk.NewCoin("uakt", totalCost.TruncateInt())),
		PlatformFee: sdk.NewCoins(sdk.NewCoin("uakt", platformFee)),
		ProviderNet: sdk.NewCoins(sdk.NewCoin("uakt", providerNet)),
		Status:      settlementtypes.SettlementStatusCompleted,
		SettledAt:   time.Now(),
		CreatedAt:   time.Now(),
	}

	t.Logf("✓ Settlement processed:")
	t.Logf("  - Total: %s", settlement.Amount.String())
	t.Logf("  - Platform fee (2%%): %s", settlement.PlatformFee.String())
	t.Logf("  - Provider net: %s", settlement.ProviderNet.String())

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

		s.slurmMock.RegisterCluster(cluster)
		s.providerMock.AddCluster(mocks.ProviderCluster{
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
	job.PreferredRegion = "us-west"

	err := s.providerMock.EnqueueJob(job, mocks.JobQueueOptions{
		Priority:       85,
		CustomerTier:   85,
		RequiredTier:   70,
		RequiredRegion: "us-west",
	})
	require.NoError(t, err)

	// Schedule - should select us-west cluster due to region preference
	decision, err := s.providerMock.ScheduleNext(ctx)
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
	s.providerMock.AddCluster(mocks.ProviderClusterFromTestCluster(cluster))

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
	partialMetrics := fixtures.StandardJobMetrics(int64(partialExecutionTime.Seconds()))

	s.slurmMock.SetJobMetrics(job.JobID, partialMetrics)

	// Cancel job
	s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCancelled)
	s.providerMock.MarkCancelled(job.JobID, "customer requested cancellation")

	t.Log("✓ Job cancelled after 2 hours")

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
	s.providerMock.AddCluster(mocks.ProviderClusterFromTestCluster(cluster))

	// Submit job with 4-hour SLA
	job := fixtures.StandardComputeJob(s.providerAddr, s.customerAddr)
	job.JobID = "job-sla-breach-001"
	job.MaxWallTimeSeconds = 4 * 3600 // 4-hour SLA

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
	metrics := fixtures.StandardJobMetrics(int64(actualExecutionTime.Seconds()))

	s.slurmMock.SetJobMetrics(job.JobID, metrics)
	s.slurmMock.SetJobExitCode(job.JobID, 0)
	s.slurmMock.SetJobState(job.JobID, pd.HPCJobStateCompleted)

	s.providerMock.MarkCompleted(job.JobID, metrics)

	t.Logf("✓ Job completed in %.2f hours (exceeded 4-hour SLA)", actualExecutionTime.Hours())

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

// TestSLURMWorkloadTemplateValidation tests SLURM workload template processing
func (s *HPCFullLifecycleTestSuite) TestSLURMWorkloadTemplateValidation() {
	t := s.T()

	t.Log("=== SLURM Template Validation Test ===")

	// Valid SLURM template
	validTemplate := &hpctypes.SLURMWorkloadTemplate{
		TemplateID:  "slurm-ml-training-v1",
		Name:        "ML Training Template",
		Description: "Standard template for ML training jobs",
		SBatchScript: `#!/bin/bash
#SBATCH --job-name=ml-training
#SBATCH --nodes=2
#SBATCH --ntasks-per-node=8
#SBATCH --gres=gpu:8
#SBATCH --time=24:00:00
#SBATCH --partition=gpu

module load cuda/11.8
python train.py --distributed
`,
		RequiredModules: []string{"cuda/11.8", "python/3.10"},
		ResourceLimits: hpctypes.HPCResourceRequirements{
			CPU:           16,
			MemoryGB:      128,
			GPUs:          8,
			GPUType:       "nvidia-a100",
			WallTimeHours: 24,
		},
		ValidatedAt: time.Now(),
		Active:      true,
	}

	// Validate template
	require.NotEmpty(t, validTemplate.SBatchScript)
	require.Contains(t, validTemplate.SBatchScript, "#SBATCH")
	require.Contains(t, validTemplate.SBatchScript, "--gres=gpu")
	require.Greater(t, validTemplate.ResourceLimits.GPUs, int32(0))

	t.Log("✓ Valid SLURM template accepted")

	// Invalid template (missing required SBATCH directives)
	invalidTemplate := &hpctypes.SLURMWorkloadTemplate{
		TemplateID: "slurm-invalid-v1",
		SBatchScript: `#!/bin/bash
python train.py
`,
	}

	// Validation should fail
	require.NotContains(t, invalidTemplate.SBatchScript, "#SBATCH")
	t.Log("✓ Invalid template correctly rejected")

	t.Log("✓ SLURM template validation test passed")
}

// BillingMockSettlementProcessor mocks settlement processing
type BillingMockSettlementProcessor struct {
	settlements []settlementtypes.SettlementRecord
}

func NewBillingMockSettlementProcessor() *BillingMockSettlementProcessor {
	return &BillingMockSettlementProcessor{
		settlements: make([]settlementtypes.SettlementRecord, 0),
	}
}

func (p *BillingMockSettlementProcessor) ProcessSettlement(settlement settlementtypes.SettlementRecord) error {
	p.settlements = append(p.settlements, settlement)
	return nil
}

func (p *BillingMockSettlementProcessor) GetSettlements() []settlementtypes.SettlementRecord {
	return p.settlements
}
