//go:build e2e.integration

// Package fixtures provides test fixtures for E2E tests.
//
// VE-15D: Test job fixtures for HPC E2E flow tests.
package fixtures

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// TestJobConfig configures a test job fixture.
type TestJobConfig struct {
	// JobID is an optional custom job ID
	JobID string

	// ClusterID is the target cluster
	ClusterID string

	// OfferingID is the HPC offering to use
	OfferingID string

	// ProviderAddress is the provider address
	ProviderAddress string

	// CustomerAddress is the customer address
	CustomerAddress string

	// QueueName is the target queue/partition
	QueueName string

	// Resources defines resource requirements
	Resources JobResourceSpec

	// Workload defines the workload specification
	Workload WorkloadSpec

	// MaxRuntimeSeconds is the maximum runtime
	MaxRuntimeSeconds int64

	// AgreedPrice is the agreed price for the job
	AgreedPrice sdk.Coins
}

// JobResourceSpec defines resource requirements for test jobs.
type JobResourceSpec struct {
	Nodes           int32
	CPUCoresPerNode int32
	MemoryGBPerNode int32
	GPUsPerNode     int32
	StorageGB       int32
	GPUType         string
}

// WorkloadSpec defines workload details for test jobs.
type WorkloadSpec struct {
	ContainerImage string
	Command        string
	Arguments      []string
	Environment    map[string]string
}

// DefaultTestJobConfig returns a default test job configuration.
func DefaultTestJobConfig() TestJobConfig {
	return TestJobConfig{
		ClusterID:         "e2e-slurm-cluster",
		OfferingID:        "hpc-compute-standard",
		ProviderAddress:   sdk.AccAddress([]byte("provider-e2e-test-00001")).String(),
		CustomerAddress:   sdk.AccAddress([]byte("customer-e2e-test-00001")).String(),
		QueueName:         "default",
		MaxRuntimeSeconds: 3600,
		Resources: JobResourceSpec{
			Nodes:           1,
			CPUCoresPerNode: 4,
			MemoryGBPerNode: 8,
			StorageGB:       10,
		},
		Workload: WorkloadSpec{
			ContainerImage: "alpine:latest",
			Command:        "echo 'Hello HPC'",
		},
		AgreedPrice: sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1000000))),
	}
}

// CreateTestJob creates an HPCJob from the config.
func CreateTestJob(config TestJobConfig) *hpctypes.HPCJob {
	jobID := config.JobID
	if jobID == "" {
		jobID = fmt.Sprintf("hpc-job-e2e-%d", time.Now().UnixNano()%1000000)
	}

	return &hpctypes.HPCJob{
		JobID:           jobID,
		ClusterID:       config.ClusterID,
		OfferingID:      config.OfferingID,
		ProviderAddress: config.ProviderAddress,
		CustomerAddress: config.CustomerAddress,
		State:           hpctypes.JobStatePending,
		QueueName:       config.QueueName,
		WorkloadSpec: hpctypes.JobWorkloadSpec{
			ContainerImage: config.Workload.ContainerImage,
			Command:        config.Workload.Command,
			Arguments:      config.Workload.Arguments,
			Environment:    config.Workload.Environment,
		},
		Resources: hpctypes.JobResources{
			Nodes:           config.Resources.Nodes,
			CPUCoresPerNode: config.Resources.CPUCoresPerNode,
			MemoryGBPerNode: config.Resources.MemoryGBPerNode,
			GPUsPerNode:     config.Resources.GPUsPerNode,
			StorageGB:       config.Resources.StorageGB,
			GPUType:         config.Resources.GPUType,
		},
		MaxRuntimeSeconds: config.MaxRuntimeSeconds,
		AgreedPrice:       config.AgreedPrice,
		CreatedAt:         time.Now(),
	}
}

// StandardComputeJob creates a standard compute job for testing.
func StandardComputeJob(provider, customer string) *hpctypes.HPCJob {
	config := DefaultTestJobConfig()
	config.JobID = fmt.Sprintf("hpc-job-compute-%d", time.Now().UnixNano()%1000000)
	config.ProviderAddress = provider
	config.CustomerAddress = customer
	config.Resources = JobResourceSpec{
		Nodes:           1,
		CPUCoresPerNode: 8,
		MemoryGBPerNode: 16,
		StorageGB:       20,
	}
	config.Workload = WorkloadSpec{
		ContainerImage: "alpine:latest",
		Command:        "stress --cpu 4 --timeout 60",
	}
	return CreateTestJob(config)
}

// GPUComputeJob creates a GPU compute job for testing.
func GPUComputeJob(provider, customer string) *hpctypes.HPCJob {
	config := DefaultTestJobConfig()
	config.JobID = fmt.Sprintf("hpc-job-gpu-%d", time.Now().UnixNano()%1000000)
	config.ProviderAddress = provider
	config.CustomerAddress = customer
	config.QueueName = "gpu"
	config.OfferingID = "hpc-gpu-a100"
	config.Resources = JobResourceSpec{
		Nodes:           1,
		CPUCoresPerNode: 16,
		MemoryGBPerNode: 64,
		GPUsPerNode:     2,
		StorageGB:       100,
		GPUType:         "nvidia-a100",
	}
	config.Workload = WorkloadSpec{
		ContainerImage: "nvidia/cuda:12.0-base",
		Command:        "nvidia-smi && echo 'GPU test complete'",
	}
	return CreateTestJob(config)
}

// MultiNodeJob creates a multi-node HPC job for testing.
func MultiNodeJob(provider, customer string, nodes int32) *hpctypes.HPCJob {
	config := DefaultTestJobConfig()
	config.JobID = fmt.Sprintf("hpc-job-multinode-%d", time.Now().UnixNano()%1000000)
	config.ProviderAddress = provider
	config.CustomerAddress = customer
	config.QueueName = "default"
	config.MaxRuntimeSeconds = 7200
	config.Resources = JobResourceSpec{
		Nodes:           nodes,
		CPUCoresPerNode: 32,
		MemoryGBPerNode: 64,
		StorageGB:       50,
	}
	config.Workload = WorkloadSpec{
		ContainerImage: "openmpi/openmpi:latest",
		Command:        fmt.Sprintf("mpirun -np %d hostname", nodes*32),
	}
	return CreateTestJob(config)
}

// HighMemoryJob creates a high-memory job for testing.
func HighMemoryJob(provider, customer string) *hpctypes.HPCJob {
	config := DefaultTestJobConfig()
	config.JobID = fmt.Sprintf("hpc-job-highmem-%d", time.Now().UnixNano()%1000000)
	config.ProviderAddress = provider
	config.CustomerAddress = customer
	config.QueueName = "highmem"
	config.OfferingID = "hpc-highmem-1tb"
	config.Resources = JobResourceSpec{
		Nodes:           1,
		CPUCoresPerNode: 64,
		MemoryGBPerNode: 512,
		StorageGB:       200,
	}
	config.Workload = WorkloadSpec{
		ContainerImage: "python:3.11",
		Command:        "python -c 'import numpy as np; a = np.zeros((1024*1024*1024,), dtype=np.float64)'",
	}
	return CreateTestJob(config)
}

// MLTrainingJob creates an ML training job for testing.
func MLTrainingJob(provider, customer string) *hpctypes.HPCJob {
	config := DefaultTestJobConfig()
	config.JobID = fmt.Sprintf("hpc-job-ml-%d", time.Now().UnixNano()%1000000)
	config.ProviderAddress = provider
	config.CustomerAddress = customer
	config.QueueName = "gpu"
	config.OfferingID = "hpc-gpu-a100"
	config.MaxRuntimeSeconds = 86400 // 24 hours
	config.Resources = JobResourceSpec{
		Nodes:           4,
		CPUCoresPerNode: 32,
		MemoryGBPerNode: 256,
		GPUsPerNode:     8,
		StorageGB:       500,
		GPUType:         "nvidia-a100",
	}
	config.Workload = WorkloadSpec{
		ContainerImage: "pytorch/pytorch:2.0-cuda12.0-runtime",
		Command:        "python train.py --epochs 100 --distributed",
		Environment: map[string]string{
			"NCCL_DEBUG":  "INFO",
			"WORLD_SIZE":  "32",
			"MASTER_PORT": "29500",
		},
	}
	return CreateTestJob(config)
}

// SimulationJob creates a simulation job for testing.
func SimulationJob(provider, customer string) *hpctypes.HPCJob {
	config := DefaultTestJobConfig()
	config.JobID = fmt.Sprintf("hpc-job-sim-%d", time.Now().UnixNano()%1000000)
	config.ProviderAddress = provider
	config.CustomerAddress = customer
	config.MaxRuntimeSeconds = 43200 // 12 hours
	config.Resources = JobResourceSpec{
		Nodes:           8,
		CPUCoresPerNode: 64,
		MemoryGBPerNode: 128,
		StorageGB:       1000,
	}
	config.Workload = WorkloadSpec{
		ContainerImage: "openfoam/openfoam:latest",
		Command:        "mpirun -np 512 simpleFoam -parallel",
	}
	return CreateTestJob(config)
}

// QuickTestJob creates a quick-running job for fast tests.
func QuickTestJob(provider, customer string) *hpctypes.HPCJob {
	config := DefaultTestJobConfig()
	config.JobID = fmt.Sprintf("hpc-job-quick-%d", time.Now().UnixNano()%1000000)
	config.ProviderAddress = provider
	config.CustomerAddress = customer
	config.MaxRuntimeSeconds = 60
	config.Resources = JobResourceSpec{
		Nodes:           1,
		CPUCoresPerNode: 1,
		MemoryGBPerNode: 1,
		StorageGB:       1,
	}
	config.Workload = WorkloadSpec{
		ContainerImage: "alpine:latest",
		Command:        "echo 'Quick test' && sleep 1",
	}
	return CreateTestJob(config)
}

// FailingJob creates a job designed to fail for testing error handling.
func FailingJob(provider, customer string) *hpctypes.HPCJob {
	config := DefaultTestJobConfig()
	config.JobID = fmt.Sprintf("hpc-job-fail-%d", time.Now().UnixNano()%1000000)
	config.ProviderAddress = provider
	config.CustomerAddress = customer
	config.MaxRuntimeSeconds = 300
	config.Resources = JobResourceSpec{
		Nodes:           1,
		CPUCoresPerNode: 2,
		MemoryGBPerNode: 4,
		StorageGB:       5,
	}
	config.Workload = WorkloadSpec{
		ContainerImage: "alpine:latest",
		Command:        "exit 1",
	}
	return CreateTestJob(config)
}

// TimeoutJob creates a job designed to timeout for testing timeout handling.
func TimeoutJob(provider, customer string) *hpctypes.HPCJob {
	config := DefaultTestJobConfig()
	config.JobID = fmt.Sprintf("hpc-job-timeout-%d", time.Now().UnixNano()%1000000)
	config.ProviderAddress = provider
	config.CustomerAddress = customer
	config.MaxRuntimeSeconds = 60 // Short timeout
	config.Resources = JobResourceSpec{
		Nodes:           1,
		CPUCoresPerNode: 1,
		MemoryGBPerNode: 2,
		StorageGB:       1,
	}
	config.Workload = WorkloadSpec{
		ContainerImage: "alpine:latest",
		Command:        "sleep 120", // Longer than timeout
	}
	return CreateTestJob(config)
}

// OversizedJob creates a job that exceeds resource limits for testing rejection.
func OversizedJob(provider, customer string) *hpctypes.HPCJob {
	config := DefaultTestJobConfig()
	config.JobID = fmt.Sprintf("hpc-job-oversized-%d", time.Now().UnixNano()%1000000)
	config.ProviderAddress = provider
	config.CustomerAddress = customer
	config.Resources = JobResourceSpec{
		Nodes:           1,
		CPUCoresPerNode: 100000, // Exceeds any reasonable capacity
		MemoryGBPerNode: 999999,
		StorageGB:       1,
	}
	config.Workload = WorkloadSpec{
		ContainerImage: "alpine:latest",
		Command:        "echo 'This should not run'",
	}
	return CreateTestJob(config)
}

// =============================================================================
// Test Metrics Fixtures
// =============================================================================

// StandardJobMetrics returns typical metrics for a completed compute job.
func StandardJobMetrics(durationSeconds int64) *pd.HPCSchedulerMetrics {
	return &pd.HPCSchedulerMetrics{
		WallClockSeconds: durationSeconds,
		CPUTimeSeconds:   durationSeconds * 4,     // 4 cores utilized
		CPUCoreSeconds:   durationSeconds * 8,     // 8 core-seconds
		MemoryBytesMax:   16 * 1024 * 1024 * 1024, // 16 GB
		MemoryGBSeconds:  durationSeconds * 16,    // 16 GB * seconds
		NodesUsed:        1,
		NodeHours:        float64(durationSeconds) / 3600.0,
		NetworkBytesIn:   1073741824, // 1 GB
		NetworkBytesOut:  536870912,  // 0.5 GB
	}
}

// GPUJobMetrics returns metrics for a GPU job.
func GPUJobMetrics(durationSeconds int64) *pd.HPCSchedulerMetrics {
	return &pd.HPCSchedulerMetrics{
		WallClockSeconds: durationSeconds,
		CPUTimeSeconds:   durationSeconds * 8,
		CPUCoreSeconds:   durationSeconds * 16,
		MemoryBytesMax:   64 * 1024 * 1024 * 1024,
		MemoryGBSeconds:  durationSeconds * 64,
		GPUSeconds:       durationSeconds * 2, // 2 GPUs
		NodesUsed:        1,
		NodeHours:        float64(durationSeconds) / 3600.0,
		NetworkBytesIn:   10737418240, // 10 GB
		NetworkBytesOut:  5368709120,  // 5 GB
	}
}

// MultiNodeJobMetrics returns metrics for a multi-node job.
func MultiNodeJobMetrics(durationSeconds int64, nodes int32) *pd.HPCSchedulerMetrics {
	return &pd.HPCSchedulerMetrics{
		WallClockSeconds: durationSeconds,
		CPUTimeSeconds:   durationSeconds * int64(nodes) * 32,
		CPUCoreSeconds:   durationSeconds * int64(nodes) * 32,
		MemoryBytesMax:   int64(nodes) * 64 * 1024 * 1024 * 1024,
		MemoryGBSeconds:  durationSeconds * int64(nodes) * 64,
		NodesUsed:        nodes,
		NodeHours:        float64(durationSeconds) * float64(nodes) / 3600.0,
		NetworkBytesIn:   int64(nodes) * 5368709120,
		NetworkBytesOut:  int64(nodes) * 2684354560,
	}
}

// PartialJobMetrics returns metrics for a partially completed job.
func PartialJobMetrics(completedSeconds int64) *pd.HPCSchedulerMetrics {
	return &pd.HPCSchedulerMetrics{
		WallClockSeconds: completedSeconds,
		CPUTimeSeconds:   completedSeconds * 2,
		CPUCoreSeconds:   completedSeconds * 4,
		MemoryBytesMax:   8 * 1024 * 1024 * 1024,
		MemoryGBSeconds:  completedSeconds * 8,
		NodesUsed:        1,
		NodeHours:        float64(completedSeconds) / 3600.0,
	}
}

// =============================================================================
// Test Cluster Configurations
// =============================================================================

// TestClusterConfig holds configuration for a test cluster.
type TestClusterConfig struct {
	ClusterID     string
	Name          string
	Region        string
	ProviderAddr  string
	TotalNodes    int32
	TotalCPU      int64
	TotalMemoryGB int64
	TotalGPUs     int64
	Partitions    []TestPartitionConfig
}

// TestPartitionConfig holds configuration for a test partition.
type TestPartitionConfig struct {
	Name       string
	Nodes      int32
	MaxRuntime int64
	MaxNodes   int32
	Features   []string
	GPUs       int32
}

// DefaultTestClusterConfig returns a default cluster configuration.
func DefaultTestClusterConfig() TestClusterConfig {
	return TestClusterConfig{
		ClusterID:     "e2e-slurm-cluster",
		Name:          "E2E Test SLURM Cluster",
		Region:        "us-east",
		ProviderAddr:  sdk.AccAddress([]byte("provider-e2e-test-00001")).String(),
		TotalNodes:    100,
		TotalCPU:      6400,
		TotalMemoryGB: 25600,
		TotalGPUs:     400,
		Partitions: []TestPartitionConfig{
			{
				Name:       "default",
				Nodes:      50,
				MaxRuntime: 86400,
				MaxNodes:   10,
				Features:   []string{"cpu"},
			},
			{
				Name:       "gpu",
				Nodes:      30,
				MaxRuntime: 172800,
				MaxNodes:   5,
				Features:   []string{"gpu", "a100"},
				GPUs:       240,
			},
			{
				Name:       "highmem",
				Nodes:      20,
				MaxRuntime: 86400,
				MaxNodes:   4,
				Features:   []string{"highmem", "1tb"},
			},
		},
	}
}

// CreateTestCluster creates an HPCCluster from the config.
func CreateTestCluster(config TestClusterConfig) *hpctypes.HPCCluster {
	partitions := make([]hpctypes.Partition, len(config.Partitions))
	for i, p := range config.Partitions {
		partitions[i] = hpctypes.Partition{
			Name:       p.Name,
			Nodes:      p.Nodes,
			MaxRuntime: p.MaxRuntime,
			MaxNodes:   p.MaxNodes,
			Features:   p.Features,
			State:      "up",
		}
	}

	return &hpctypes.HPCCluster{
		ClusterID:       config.ClusterID,
		ProviderAddress: config.ProviderAddr,
		Name:            config.Name,
		Region:          config.Region,
		State:           hpctypes.ClusterStateActive,
		TotalNodes:      config.TotalNodes,
		AvailableNodes:  config.TotalNodes,
		Partitions:      partitions,
		SLURMVersion:    "23.02.4",
		ClusterMetadata: hpctypes.ClusterMetadata{
			TotalCPUCores:    config.TotalCPU,
			TotalMemoryGB:    config.TotalMemoryGB,
			TotalGPUs:        config.TotalGPUs,
			InterconnectType: "infiniband",
			StorageType:      "lustre",
			TotalStorageGB:   100000,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// =============================================================================
// Test Offering Configurations
// =============================================================================

// TestOfferingConfig holds configuration for a test offering.
type TestOfferingConfig struct {
	OfferingID                string
	ClusterID                 string
	ProviderAddr              string
	Name                      string
	MaxRuntimeSeconds         int64
	SupportsCustomWorkloads   bool
	RequiredIdentityThreshold int32
	Pricing                   TestPricingConfig
	QueueOptions              []TestQueueOption
}

// TestPricingConfig holds pricing configuration.
type TestPricingConfig struct {
	BaseNodeHourPrice string
	CPUCoreHourPrice  string
	GPUHourPrice      string
	MemoryGBHourPrice string
	StorageGBPrice    string
	Currency          string
	MinimumCharge     string
}

// TestQueueOption holds queue option configuration.
type TestQueueOption struct {
	PartitionName   string
	DisplayName     string
	MaxNodes        int32
	MaxRuntime      int64
	PriceMultiplier string
}

// DefaultTestOfferingConfig returns a default offering configuration.
func DefaultTestOfferingConfig() TestOfferingConfig {
	return TestOfferingConfig{
		OfferingID:                "hpc-compute-standard",
		ClusterID:                 "e2e-slurm-cluster",
		ProviderAddr:              sdk.AccAddress([]byte("provider-e2e-test-00001")).String(),
		Name:                      "HPC Compute Standard",
		MaxRuntimeSeconds:         86400,
		SupportsCustomWorkloads:   true,
		RequiredIdentityThreshold: 70,
		Pricing: TestPricingConfig{
			BaseNodeHourPrice: "10.0",
			CPUCoreHourPrice:  "0.10",
			GPUHourPrice:      "2.50",
			MemoryGBHourPrice: "0.01",
			StorageGBPrice:    "0.001",
			Currency:          "uvirt",
			MinimumCharge:     "1.0",
		},
		QueueOptions: []TestQueueOption{
			{
				PartitionName:   "default",
				DisplayName:     "Standard Compute",
				MaxNodes:        10,
				MaxRuntime:      86400,
				PriceMultiplier: "1.0",
			},
			{
				PartitionName:   "gpu",
				DisplayName:     "GPU Compute",
				MaxNodes:        5,
				MaxRuntime:      172800,
				PriceMultiplier: "2.5",
			},
		},
	}
}

// CreateTestOffering creates an HPCOffering from the config.
func CreateTestOffering(config TestOfferingConfig) *hpctypes.HPCOffering {
	queueOptions := make([]hpctypes.QueueOption, len(config.QueueOptions))
	for i, q := range config.QueueOptions {
		queueOptions[i] = hpctypes.QueueOption{
			PartitionName:   q.PartitionName,
			DisplayName:     q.DisplayName,
			MaxNodes:        q.MaxNodes,
			MaxRuntime:      q.MaxRuntime,
			PriceMultiplier: q.PriceMultiplier,
		}
	}

	return &hpctypes.HPCOffering{
		OfferingID:      config.OfferingID,
		ClusterID:       config.ClusterID,
		ProviderAddress: config.ProviderAddr,
		Name:            config.Name,
		QueueOptions:    queueOptions,
		Pricing: hpctypes.HPCPricing{
			BaseNodeHourPrice: config.Pricing.BaseNodeHourPrice,
			CPUCoreHourPrice:  config.Pricing.CPUCoreHourPrice,
			GPUHourPrice:      config.Pricing.GPUHourPrice,
			MemoryGBHourPrice: config.Pricing.MemoryGBHourPrice,
			StorageGBPrice:    config.Pricing.StorageGBPrice,
			Currency:          config.Pricing.Currency,
			MinimumCharge:     config.Pricing.MinimumCharge,
		},
		RequiredIdentityThreshold: config.RequiredIdentityThreshold,
		MaxRuntimeSeconds:         config.MaxRuntimeSeconds,
		SupportsCustomWorkloads:   config.SupportsCustomWorkloads,
		Active:                    true,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}
}
