//go:build e2e.integration

package e2e

import (
	"testing"
	"time"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
)

const telemetryGiB = int64(1024 * 1024 * 1024)

func newRealSchedulerMetrics(
	duration time.Duration,
	nodes int32,
	cpuCoresPerNode int32,
	memoryGBPerNode int32,
	gpusPerNode int32,
	storageGB int64,
) *pd.HPCSchedulerMetrics {
	wallSeconds := int64(duration.Seconds())
	if wallSeconds < 1 {
		wallSeconds = 1
	}

	totalCPUCores := int64(nodes) * int64(cpuCoresPerNode)
	totalMemoryGB := int64(nodes) * int64(memoryGBPerNode)
	totalGPUs := int64(nodes) * int64(gpusPerNode)
	nodeHours := float64(wallSeconds) * float64(nodes) / 3600.0

	return &pd.HPCSchedulerMetrics{
		WallClockSeconds: wallSeconds,
		CPUTimeSeconds:   wallSeconds * totalCPUCores * 8 / 10,
		CPUCoreSeconds:   wallSeconds * totalCPUCores,
		MemoryBytesMax:   totalMemoryGB * telemetryGiB,
		MemoryGBSeconds:  wallSeconds * totalMemoryGB,
		GPUSeconds:       wallSeconds * totalGPUs,
		NodesUsed:        nodes,
		NodeHours:        nodeHours,
		StorageGBHours:   int64(nodeHours * float64(storageGB)),
		NetworkBytesIn:   totalMemoryGB * telemetryGiB / 2,
		NetworkBytesOut:  totalMemoryGB * telemetryGiB / 4,
		EnergyJoules:     wallSeconds * maxTelemetry64(totalCPUCores*45, 90),
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

func requireRealSchedulerTelemetry(t *testing.T, metrics *pd.HPCSchedulerMetrics) {
	t.Helper()

	if metrics == nil {
		t.Fatal("expected scheduler metrics")
	}
	if metrics.StorageGBHours <= 0 {
		t.Fatalf("expected storage hours > 0, got %d", metrics.StorageGBHours)
	}
	if metrics.NetworkBytesIn <= 0 || metrics.NetworkBytesOut <= 0 {
		t.Fatalf("expected network bytes > 0, got in=%d out=%d", metrics.NetworkBytesIn, metrics.NetworkBytesOut)
	}
	if metrics.EnergyJoules <= 0 {
		t.Fatalf("expected energy joules > 0, got %d", metrics.EnergyJoules)
	}

	requiredSignals := []string{
		"average_cpu_utilization_percent",
		"average_memory_utilization_percent",
		"disk_io_utilization_percent",
		"network_utilization_percent",
		"slurm_state",
	}
	for _, key := range requiredSignals {
		if _, ok := metrics.SchedulerSpecific[key]; !ok {
			t.Fatalf("expected scheduler_specific[%q] to be present", key)
		}
	}
}

func maxTelemetry64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
