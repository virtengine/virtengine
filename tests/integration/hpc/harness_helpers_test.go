package hpc

import (
	"time"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
)

const gib = int64(1024 * 1024 * 1024)

func realNodeCapacityFixture() *hpcv1.NodeCapacity {
	return &hpcv1.NodeCapacity{
		CpuCoresTotal:      64,
		CpuCoresAvailable:  44,
		CpuCoresAllocated:  20,
		MemoryGbTotal:      512,
		MemoryGbAvailable:  384,
		MemoryGbAllocated:  128,
		GpusTotal:          8,
		GpusAvailable:      6,
		GpusAllocated:      2,
		GpuType:            "NVIDIA H100",
		StorageGbTotal:     4000,
		StorageGbAvailable: 2750,
		StorageGbAllocated: 1250,
	}
}

func realNodeHealthFixture(status hpcv1.HealthStatus, slurmState string) *hpcv1.NodeHealth {
	return &hpcv1.NodeHealth{
		Status:                      status,
		UptimeSeconds:               7200,
		LoadAverage_1M:              "19.250000",
		LoadAverage_5M:              "17.500000",
		LoadAverage_15M:             "16.000000",
		CpuUtilizationPercent:       68,
		MemoryUtilizationPercent:    75,
		GpuUtilizationPercent:       81,
		GpuMemoryUtilizationPercent: 77,
		DiskIoUtilizationPercent:    43,
		NetworkUtilizationPercent:   36,
		GpuTemperatureCelsius:       72,
		SlurmState:                  slurmState,
	}
}

func realNodeHardwareFixture() *hpcv1.NodeHardware {
	return &hpcv1.NodeHardware{
		CpuModel:    "AMD EPYC 9654",
		CpuVendor:   "AMD",
		CpuArch:     "x86_64",
		GpuModel:    "NVIDIA H100",
		GpuMemoryGb: 80,
		StorageType: "NVMe",
		Features:    []string{"infiniband", "rdma", "gpu"},
	}
}

func realNodeLocalityFixture() *hpcv1.NodeLocality {
	return &hpcv1.NodeLocality{
		Region:     "us-east-1",
		Datacenter: "dc1",
		Zone:       "a",
		Rack:       "rack-7",
		Row:        "row-2",
		Position:   "u14",
	}
}

func realSchedulerMetricsFixture(
	duration time.Duration,
	nodesUsed int32,
	cpuCoresPerNode int32,
	memoryGBPerNode int32,
	gpusPerNode int32,
	storageGB int64,
) *pd.HPCSchedulerMetrics {
	wallSeconds := int64(duration.Seconds())
	if wallSeconds < 1 {
		wallSeconds = 1
	}

	totalCPUCores := int64(nodesUsed) * int64(cpuCoresPerNode)
	totalMemoryGB := int64(nodesUsed) * int64(memoryGBPerNode)
	totalGPUs := int64(nodesUsed) * int64(gpusPerNode)
	nodeHours := float64(wallSeconds) * float64(nodesUsed) / 3600.0

	return &pd.HPCSchedulerMetrics{
		WallClockSeconds: wallSeconds,
		CPUTimeSeconds:   wallSeconds * totalCPUCores * 8 / 10,
		CPUCoreSeconds:   wallSeconds * totalCPUCores,
		MemoryBytesMax:   totalMemoryGB * gib,
		MemoryGBSeconds:  wallSeconds * totalMemoryGB,
		GPUSeconds:       wallSeconds * totalGPUs,
		NodesUsed:        nodesUsed,
		NodeHours:        nodeHours,
		StorageGBHours:   int64(nodeHours * float64(storageGB)),
		NetworkBytesIn:   totalMemoryGB * gib / 2,
		NetworkBytesOut:  totalMemoryGB * gib / 4,
		EnergyJoules:     wallSeconds * max64(totalCPUCores*45, 90),
		SchedulerSpecific: map[string]interface{}{
			"average_cpu_utilization_percent":        int32(68),
			"average_memory_utilization_percent":     int32(75),
			"average_gpu_utilization_percent":        int32(81),
			"average_gpu_memory_utilization_percent": int32(77),
			"disk_io_utilization_percent":            int32(43),
			"network_utilization_percent":            int32(36),
			"slurm_state":                            "mixed",
			"running_jobs":                           int32(3),
			"pending_jobs":                           int32(1),
			"active_job_ids":                         []string{"job-a", "job-b", "job-c"},
		},
	}
}

func max64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
