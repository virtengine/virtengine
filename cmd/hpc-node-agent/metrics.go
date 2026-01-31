// Package main implements the VirtEngine HPC Node Agent.
//
// VE-500: Metrics collection for node agent heartbeats.
// VE-7A: Command injection prevention and input sanitization
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/virtengine/virtengine/pkg/security"
)

// MetricsCollector collects system metrics for heartbeats
type MetricsCollector struct {
	startTime time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime: time.Now(),
	}
}

// CollectCapacity collects node capacity metrics
func (m *MetricsCollector) CollectCapacity() (*NodeCapacity, error) {
	capacity := &NodeCapacity{}

	// CPU cores
	capacity.CPUCoresTotal = int32(runtime.NumCPU())
	capacity.CPUCoresAvailable = capacity.CPUCoresTotal // Simplified; would check SLURM allocation

	// Memory
	memTotal, memAvailable := m.getMemoryInfo()
	capacity.MemoryGBTotal = int32(memTotal / (1024 * 1024 * 1024))
	capacity.MemoryGBAvailable = int32(memAvailable / (1024 * 1024 * 1024))

	// GPU (placeholder - would use nvidia-smi or similar)
	capacity.GPUsTotal, capacity.GPUsAvailable, capacity.GPUType = m.getGPUInfo()

	// Storage
	storageTotal, storageAvailable := m.getStorageInfo("/")
	capacity.StorageGBTotal = int32(storageTotal / (1024 * 1024 * 1024))
	capacity.StorageGBAvailable = int32(storageAvailable / (1024 * 1024 * 1024))

	return capacity, nil
}

// CollectHealth collects node health metrics
func (m *MetricsCollector) CollectHealth() (*NodeHealth, error) {
	health := &NodeHealth{
		Status:        "healthy",
		UptimeSeconds: int64(time.Since(m.startTime).Seconds()),
	}

	// Load average
	load1, load5, load15 := m.getLoadAverage()
	health.LoadAverage1m = fmt.Sprintf("%.6f", load1)
	health.LoadAverage5m = fmt.Sprintf("%.6f", load5)
	health.LoadAverage15m = fmt.Sprintf("%.6f", load15)

	// CPU utilization (simplified)
	health.CPUUtilizationPercent = m.getCPUUtilization()

	// Memory utilization
	memTotal, memAvailable := m.getMemoryInfo()
	if memTotal > 0 {
		health.MemoryUtilizationPercent = int32(100 * (memTotal - memAvailable) / memTotal)
	}

	// Disk I/O utilization (placeholder)
	health.DiskIOUtilizationPercent = 0

	// Network utilization (placeholder)
	health.NetworkUtilizationPercent = 0

	// SLURM state
	health.SLURMState = m.getSLURMNodeState()

	// Determine overall health status
	if health.CPUUtilizationPercent > 90 || health.MemoryUtilizationPercent > 90 {
		health.Status = "degraded"
	}
	if health.SLURMState == "down" || health.SLURMState == "drain" {
		health.Status = "draining"
	}

	return health, nil
}

// CollectLatency collects latency measurements
func (m *MetricsCollector) CollectLatency(targets []string) *NodeLatency {
	latency := &NodeLatency{
		Measurements: make([]LatencyProbe, 0, len(targets)),
	}

	for _, target := range targets {
		probe := m.measureLatency(target)
		if probe != nil {
			latency.Measurements = append(latency.Measurements, *probe)
		}
	}

	// Calculate average
	if len(latency.Measurements) > 0 {
		var total int64
		for _, probe := range latency.Measurements {
			total += probe.LatencyUs
		}
		latency.AvgClusterLatency = total / int64(len(latency.Measurements))
	}

	return latency
}

// CollectJobs collects job information
func (m *MetricsCollector) CollectJobs() *NodeJobs {
	jobs := &NodeJobs{}

	// Build validated arguments for squeue command
	args, err := security.SLURMSqueueArgs("%T", "", "")
	if err != nil {
		// Log error and return empty jobs
		return jobs
	}

	// Try to get SLURM job counts using validated arguments
	if output, err := exec.Command("squeue", args...).Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			switch strings.TrimSpace(line) {
			case "RUNNING":
				jobs.RunningCount++
			case "PENDING":
				jobs.PendingCount++
			}
		}
	}

	return jobs
}

// CollectServices collects service status
func (m *MetricsCollector) CollectServices() *NodeServices {
	services := &NodeServices{}

	// Check slurmd
	if _, err := exec.Command("pgrep", "-x", "slurmd").Output(); err == nil {
		services.SLURMDRunning = true
	}

	// Get slurmd version
	if output, err := exec.Command("slurmd", "--version").Output(); err == nil {
		services.SLURMDVersion = strings.TrimSpace(string(output))
	}

	// Check munge
	if _, err := exec.Command("pgrep", "-x", "munged").Output(); err == nil {
		services.MungeRunning = true
	}

	// Check container runtime
	if _, err := exec.LookPath("singularity"); err == nil {
		services.ContainerRuntime = "singularity"
		if output, err := exec.Command("singularity", "--version").Output(); err == nil {
			services.ContainerRuntimeVersion = strings.TrimSpace(string(output))
		}
	} else if _, err := exec.LookPath("docker"); err == nil {
		services.ContainerRuntime = "docker"
		if output, err := exec.Command("docker", "--version").Output(); err == nil {
			services.ContainerRuntimeVersion = strings.TrimSpace(string(output))
		}
	}

	return services
}

func (m *MetricsCollector) getMemoryInfo() (uint64, uint64) {
	if runtime.GOOS == "linux" {
		file, err := os.Open("/proc/meminfo")
		if err != nil {
			return 0, 0
		}
		defer file.Close()

		var memTotal, memAvailable uint64
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}

			value, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}

			switch fields[0] {
			case "MemTotal:":
				memTotal = value * 1024 // Convert KB to bytes
			case "MemAvailable:":
				memAvailable = value * 1024
			}
		}
		return memTotal, memAvailable
	}

	// Fallback for non-Linux systems
	return 16 * 1024 * 1024 * 1024, 8 * 1024 * 1024 * 1024 // 16GB total, 8GB available
}

func (m *MetricsCollector) getLoadAverage() (float64, float64, float64) {
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/proc/loadavg")
		if err != nil {
			return 0, 0, 0
		}

		fields := strings.Fields(string(data))
		if len(fields) < 3 {
			return 0, 0, 0
		}

		load1, _ := strconv.ParseFloat(fields[0], 64)
		load5, _ := strconv.ParseFloat(fields[1], 64)
		load15, _ := strconv.ParseFloat(fields[2], 64)
		return load1, load5, load15
	}

	return 0, 0, 0
}

func (m *MetricsCollector) getCPUUtilization() int32 {
	if runtime.GOOS == "linux" {
		// Read CPU stats
		data, err := os.ReadFile("/proc/stat")
		if err != nil {
			return 0
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "cpu ") {
				fields := strings.Fields(line)
				if len(fields) < 5 {
					return 0
				}

				user, _ := strconv.ParseUint(fields[1], 10, 64)
				nice, _ := strconv.ParseUint(fields[2], 10, 64)
				system, _ := strconv.ParseUint(fields[3], 10, 64)
				idle, _ := strconv.ParseUint(fields[4], 10, 64)

				total := user + nice + system + idle
				if total > 0 {
					used := user + nice + system
					return int32(100 * used / total)
				}
				break
			}
		}
	}

	return 0
}

func (m *MetricsCollector) getGPUInfo() (int32, int32, string) {
	// Try nvidia-smi
	output, err := exec.Command("nvidia-smi", "--query-gpu=count,name", "--format=csv,noheader").Output()
	if err != nil {
		return 0, 0, ""
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return 0, 0, ""
	}

	count := int32(len(lines))
	gpuType := ""
	if parts := strings.SplitN(lines[0], ",", 2); len(parts) == 2 {
		gpuType = strings.TrimSpace(parts[1])
	}

	return count, count, gpuType // Simplified: all GPUs available
}

func (m *MetricsCollector) getStorageInfo(path string) (uint64, uint64) {
	if runtime.GOOS == "linux" {
		// Validate and sanitize the path argument
		args, err := security.DfArgs(path)
		if err != nil {
			return 0, 0
		}

		output, err := exec.Command("df", args...).Output()
		if err != nil {
			return 0, 0
		}

		lines := strings.Split(string(output), "\n")
		if len(lines) < 2 {
			return 0, 0
		}

		fields := strings.Fields(lines[1])
		if len(fields) < 4 {
			return 0, 0
		}

		total, _ := strconv.ParseUint(fields[1], 10, 64)
		available, _ := strconv.ParseUint(fields[3], 10, 64)
		return total, available
	}

	return 1000 * 1024 * 1024 * 1024, 500 * 1024 * 1024 * 1024 // 1TB total, 500GB available
}

func (m *MetricsCollector) getSLURMNodeState() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}

	// Validate hostname before using in command
	if err := security.ValidateHostname(hostname); err != nil {
		return "unknown"
	}

	// Build validated arguments for sinfo command
	args, err := security.SLURMSinfoArgs("%T", hostname)
	if err != nil {
		return "unknown"
	}

	output, err := exec.Command("sinfo", args...).Output()
	if err != nil {
		return "unknown"
	}

	return strings.TrimSpace(string(output))
}

func (m *MetricsCollector) measureLatency(target string) *LatencyProbe {
	start := time.Now()

	// Validate target before using in any command
	if err := security.ValidatePingTarget(target); err != nil {
		return nil
	}

	// Try TCP connection as a proxy for latency
	conn, err := net.DialTimeout("tcp", target+":22", 5*time.Second)
	if err != nil {
		// Build validated ping arguments
		args, pingErr := security.PingArgs(target, 1)
		if pingErr != nil {
			return nil
		}

		// Try ICMP ping if available
		_, execErr := exec.Command("ping", args...).Output()
		if execErr != nil {
			return nil
		}
	} else {
		conn.Close()
	}

	latency := time.Since(start)
	return &LatencyProbe{
		TargetNodeID:      target,
		LatencyUs:         latency.Microseconds(),
		PacketLossPercent: 0,
		MeasuredAt:        time.Now(),
	}
}
