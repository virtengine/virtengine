package hpc

import (
	"time"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
)

const gib = int64(1024 * 1024 * 1024)

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
