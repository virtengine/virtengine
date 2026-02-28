package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCollectCapacityDerivesAllocatedResources(t *testing.T) {
	t.Parallel()

	totalCores := int32(runtime.NumCPU())
	load1 := float64(totalCores)
	if totalCores > 2 {
		load1 = float64(totalCores - 2)
	}

	collector := newTestCollector(
		map[string]string{
			"/proc/loadavg": fmt.Sprintf("%.2f 0.50 0.25 1/100 123\n", load1),
			"/proc/meminfo": "MemTotal:       131072000 kB\nMemAvailable:    98304000 kB\n",
		},
		map[string]string{
			commandKey("nvidia-smi",
				"--query-gpu=name,utilization.gpu,utilization.memory,memory.total,memory.used,temperature.gpu",
				"--format=csv,noheader,nounits"): "NVIDIA H100, 80, 60, 81920, 65536, 72\nNVIDIA H100, 0, 0, 81920, 0, 45\n",
			commandKey("df", "-B1", "/"): "Filesystem 1B-blocks Used Available Use% Mounted on\n/dev/sda1 1099511627776 329853488332 769658139444 30% /\n",
		},
	)

	capacity, err := collector.CollectCapacity()
	if err != nil {
		t.Fatalf("CollectCapacity() error = %v", err)
	}

	expectedAvailableCores, expectedAllocatedCores := availableCoresFromLoad(totalCores, load1)
	if capacity.CPUCoresAvailable != expectedAvailableCores {
		t.Fatalf("CPUCoresAvailable = %d, want %d", capacity.CPUCoresAvailable, expectedAvailableCores)
	}
	if capacity.CPUCoresAllocated != expectedAllocatedCores {
		t.Fatalf("CPUCoresAllocated = %d, want %d", capacity.CPUCoresAllocated, expectedAllocatedCores)
	}
	if capacity.MemoryGBAllocated != capacity.MemoryGBTotal-capacity.MemoryGBAvailable {
		t.Fatalf("MemoryGBAllocated = %d, want %d", capacity.MemoryGBAllocated, capacity.MemoryGBTotal-capacity.MemoryGBAvailable)
	}
	if capacity.GPUsTotal != 2 || capacity.GPUsAvailable != 1 || capacity.GPUsAllocated != 1 {
		t.Fatalf("unexpected GPU capacity: %+v", capacity)
	}
	if capacity.StorageGBAllocated != capacity.StorageGBTotal-capacity.StorageGBAvailable {
		t.Fatalf("StorageGBAllocated = %d, want %d", capacity.StorageGBAllocated, capacity.StorageGBTotal-capacity.StorageGBAvailable)
	}
}

func TestCollectHealthAggregatesBoundedRealMetrics(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0)
	collector := newTestCollector(
		map[string]string{
			"/proc/loadavg":             "9.75 8.25 7.50 1/100 123\n",
			"/proc/meminfo":             "MemTotal:       104857600 kB\nMemAvailable:    8388608 kB\n",
			"/proc/stat":                "cpu  300 0 160 700 40 0 0 0 0 0\n",
			"/proc/diskstats":           "8 0 sda 100 0 0 0 200 0 0 0 0 4000 0\n7 0 loop0 0 0 0 0 0 0 0 0 0 0 0\n",
			"/proc/net/dev":             "Inter-|   Receive                                                |  Transmit\n face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed\n  eth0: 126000000 0 0 0 0 0 0 0 125000000 0 0 0 0 0 0 0\n    lo: 1000 0 0 0 0 0 0 0 1000 0 0 0 0 0 0 0\n",
			"/sys/class/net/eth0/speed": "1000\n",
		},
		map[string]string{
			commandKey("nvidia-smi",
				"--query-gpu=name,utilization.gpu,utilization.memory,memory.total,memory.used,temperature.gpu",
				"--format=csv,noheader,nounits"): "NVIDIA H100, 90, 80, 81920, 70000, 75\nNVIDIA H100, 70, 60, 81920, 60000, 70\n",
			commandKey("sinfo", "-h", "-o", "%T", "-N", "-n", "node-a"): "MIXED+PLANNED\n",
		},
	)
	collector.now = func() time.Time { return now }
	collector.startTime = now.Add(-2 * time.Hour)
	collector.hostname = func() (string, error) { return "node-a", nil }
	collector.lastCPUSample = &cpuSample{total: 1000, idle: 700}
	collector.lastDiskSample = &diskSample{collectedAt: now.Add(-10 * time.Second), busyMillis: 1000}
	collector.lastNetworkSample = &networkSample{
		collectedAt:    now.Add(-10 * time.Second),
		totalBytes:     1_000_000,
		linkBitsPerSec: 1_000_000_000,
	}

	health, err := collector.CollectHealth()
	if err != nil {
		t.Fatalf("CollectHealth() error = %v", err)
	}

	if health.CPUUtilizationPercent != 80 {
		t.Fatalf("CPUUtilizationPercent = %d, want 80", health.CPUUtilizationPercent)
	}
	if health.MemoryUtilizationPercent != 92 {
		t.Fatalf("MemoryUtilizationPercent = %d, want 92", health.MemoryUtilizationPercent)
	}
	if health.GPUUtilizationPercent != 80 || health.GPUMemoryUtilizationPercent != 70 {
		t.Fatalf("unexpected GPU metrics: %+v", health)
	}
	if health.GPUTemperatureCelsius != 72 {
		t.Fatalf("GPUTemperatureCelsius = %d, want 72", health.GPUTemperatureCelsius)
	}
	if health.DiskIOUtilizationPercent != 30 {
		t.Fatalf("DiskIOUtilizationPercent = %d, want 30", health.DiskIOUtilizationPercent)
	}
	if health.NetworkUtilizationPercent != 20 {
		t.Fatalf("NetworkUtilizationPercent = %d, want 20", health.NetworkUtilizationPercent)
	}
	if health.SLURMState != "mixed" {
		t.Fatalf("SLURMState = %q, want mixed", health.SLURMState)
	}
	if health.Status != "degraded" {
		t.Fatalf("Status = %q, want degraded", health.Status)
	}
}

func TestCollectHealthFailsClosedWithoutPriorSamples(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_500, 0)
	collector := newTestCollector(
		map[string]string{
			"/proc/loadavg":             "0.10 0.20 0.30 1/100 123\n",
			"/proc/meminfo":             "MemTotal:       1048576 kB\nMemAvailable:    524288 kB\n",
			"/proc/stat":                "cpu  10 0 5 100 0 0 0 0 0 0\n",
			"/proc/diskstats":           "8 0 sda 100 0 0 0 200 0 0 0 0 4000 0\n",
			"/proc/net/dev":             "Inter-|   Receive                                                |  Transmit\n face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed\n  eth0: 1000 0 0 0 0 0 0 0 2000 0 0 0 0 0 0 0\n",
			"/sys/class/net/eth0/speed": "1000\n",
		},
		map[string]string{
			commandKey("sinfo", "-h", "-o", "%T", "-N", "-n", "node-a"): "IDLE\n",
		},
	)
	collector.now = func() time.Time { return now }
	collector.hostname = func() (string, error) { return "node-a", nil }

	health, err := collector.CollectHealth()
	if err != nil {
		t.Fatalf("CollectHealth() error = %v", err)
	}

	if health.CPUUtilizationPercent != 0 {
		t.Fatalf("CPUUtilizationPercent = %d, want 0 on first sample", health.CPUUtilizationPercent)
	}
	if health.DiskIOUtilizationPercent != 0 {
		t.Fatalf("DiskIOUtilizationPercent = %d, want 0 on first sample", health.DiskIOUtilizationPercent)
	}
	if health.NetworkUtilizationPercent != 0 {
		t.Fatalf("NetworkUtilizationPercent = %d, want 0 on first sample", health.NetworkUtilizationPercent)
	}
	if health.Status != "healthy" {
		t.Fatalf("Status = %q, want healthy", health.Status)
	}
}

func TestCollectJobsParsesSchedulerSignalsAndBoundsIDs(t *testing.T) {
	t.Parallel()

	var lines []string
	for i := 0; i < 30; i++ {
		lines = append(lines, fmt.Sprintf("%d,RUNNING", 1000+i))
	}
	lines = append(lines, "2000,PENDING", "2001,PENDING", "3000,COMPLETED")

	collector := newTestCollector(nil, map[string]string{
		commandKey("squeue", "-h", "-o", "%i,%T"): strings.Join(lines, "\n"),
	})

	jobs := collector.CollectJobs()
	if jobs.RunningCount != 30 {
		t.Fatalf("RunningCount = %d, want 30", jobs.RunningCount)
	}
	if jobs.PendingCount != 2 {
		t.Fatalf("PendingCount = %d, want 2", jobs.PendingCount)
	}
	if len(jobs.ActiveJobIDs) != maxTrackedJobIDs {
		t.Fatalf("len(ActiveJobIDs) = %d, want %d", len(jobs.ActiveJobIDs), maxTrackedJobIDs)
	}
	if jobs.ActiveJobIDs[0] != "1000" || jobs.ActiveJobIDs[len(jobs.ActiveJobIDs)-1] != "1024" {
		t.Fatalf("unexpected bounded job IDs: %v", jobs.ActiveJobIDs)
	}
}

func TestNormalizeSLURMState(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"ALLOCATED+PLANNED": "allocated",
		"MIXED":             "mixed",
		"DRAINING":          "drain",
		"DOWN*":             "down",
		"IDLE":              "idle",
		"unknown-state":     osUnknown,
	}

	for raw, want := range tests {
		if got := normalizeSLURMState(raw); got != want {
			t.Fatalf("normalizeSLURMState(%q) = %q, want %q", raw, got, want)
		}
	}
}

func newTestCollector(files map[string]string, commands map[string]string) *MetricsCollector {
	fileMap := make(map[string][]byte, len(files))
	for path, data := range files {
		fileMap[path] = []byte(data)
	}

	return &MetricsCollector{
		startTime: time.Unix(0, 0),
		osName:    osLinux,
		readFile: func(path string) ([]byte, error) {
			if data, ok := fileMap[path]; ok {
				return data, nil
			}
			return nil, os.ErrNotExist
		},
		resolveExecutable: func(category string, name string) (string, error) {
			return name, nil
		},
		execCommand: func(path string, args ...string) ([]byte, error) {
			if output, ok := commands[commandKey(path, args...)]; ok {
				return []byte(output), nil
			}
			return nil, fmt.Errorf("unexpected command: %s", commandKey(path, args...))
		},
		hostname: os.Hostname,
		now:      time.Now,
	}
}

func commandKey(path string, args ...string) string {
	return path + " " + strings.Join(args, " ")
}
