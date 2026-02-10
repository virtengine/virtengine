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

const (
	osLinux   = "linux"
	osUnknown = "unknown"
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
//
//nolint:unparam // result 1 (error) reserved for hardware query failures
func (m *MetricsCollector) CollectCapacity() (*NodeCapacity, error) {
	capacity := &NodeCapacity{}

	// CPU cores
	//nolint:gosec // G115: NumCPU returns small positive int, safe for int32
	capacity.CPUCoresTotal = int32(runtime.NumCPU())
	capacity.CPUCoresAvailable = capacity.CPUCoresTotal // Simplified; would check SLURM allocation

	// Memory
	memTotal, memAvailable := m.getMemoryInfo()
	//nolint:gosec // G115: memory in GB is bounded well under int32 max
	capacity.MemoryGBTotal = int32(memTotal / (1024 * 1024 * 1024))
	//nolint:gosec // G115: memory in GB is bounded well under int32 max
	capacity.MemoryGBAvailable = int32(memAvailable / (1024 * 1024 * 1024))

	// GPU (placeholder - would use nvidia-smi or similar)
	capacity.GPUsTotal, capacity.GPUsAvailable, capacity.GPUType = m.getGPUInfo()

	// Storage
	storageTotal, storageAvailable := m.getStorageInfo("/")
	//nolint:gosec // G115: storage in GB is bounded well under int32 max
	capacity.StorageGBTotal = int32(storageTotal / (1024 * 1024 * 1024))
	//nolint:gosec // G115: storage in GB is bounded well under int32 max
	capacity.StorageGBAvailable = int32(storageAvailable / (1024 * 1024 * 1024))

	return capacity, nil
}

// CollectHealth collects node health metrics
//
//nolint:unparam // result 1 (error) reserved for health check failures
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
		//nolint:gosec // G115: percentage is 0-100, safe for int32
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

// CollectHardware collects node hardware details.
//
//nolint:unparam // Signature reserved for future hardware probes that may return errors.
func (m *MetricsCollector) CollectHardware() (*NodeHardware, error) {
	hardware := &NodeHardware{
		CPUArch: runtime.GOARCH,
	}

	if runtime.GOOS == osLinux {
		model, vendor := m.getCPUInfo()
		hardware.CPUModel = model
		hardware.CPUVendor = vendor
	}

	_, _, gpuType := m.getGPUInfo()
	hardware.GPUModel = gpuType

	return hardware, nil
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

	// Resolve and validate squeue executable
	squeuePath, err := security.ResolveAndValidateExecutable("slurm", "squeue")
	if err != nil {
		// SLURM not available, return empty jobs
		return jobs
	}

	// Execute with validated path and arguments
	//nolint:gosec // G204: Executable path and arguments validated by security package
	if output, err := exec.Command(squeuePath, args...).Output(); err == nil {
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

	// Check slurmd using pgrep
	if pgrepPath, err := security.ResolveAndValidateExecutable("system", "pgrep"); err == nil {
		//nolint:gosec // G204: Executable path validated, args are fixed literals
		if _, err := exec.Command(pgrepPath, "-x", "slurmd").Output(); err == nil {
			services.SLURMDRunning = true
		}
	}

	// Get slurmd version
	if slurmdPath, err := security.ResolveAndValidateExecutable("slurm", "slurmd"); err == nil {
		//nolint:gosec // G204: Executable path validated, args are fixed literals
		if output, err := exec.Command(slurmdPath, "--version").Output(); err == nil {
			services.SLURMDVersion = strings.TrimSpace(string(output))
		}
	}

	// Check munge
	if pgrepPath, err := security.ResolveAndValidateExecutable("system", "pgrep"); err == nil {
		//nolint:gosec // G204: Executable path validated, args are fixed literals
		if _, err := exec.Command(pgrepPath, "-x", "munged").Output(); err == nil {
			services.MungeRunning = true
		}
	}

	// Check container runtime
	if singularityPath, err := security.ResolveAndValidateExecutable("system", "singularity"); err == nil {
		services.ContainerRuntime = "singularity"
		//nolint:gosec // G204: Executable path validated, args are fixed literals
		if output, err := exec.Command(singularityPath, "--version").Output(); err == nil {
			services.ContainerRuntimeVersion = strings.TrimSpace(string(output))
		}
	} else if dockerPath, err := security.ResolveAndValidateExecutable("system", "docker"); err == nil {
		services.ContainerRuntime = "docker"
		//nolint:gosec // G204: Executable path validated, args are fixed literals
		if output, err := exec.Command(dockerPath, "--version").Output(); err == nil {
			services.ContainerRuntimeVersion = strings.TrimSpace(string(output))
		}
	}

	return services
}

func (m *MetricsCollector) getMemoryInfo() (uint64, uint64) {
	if runtime.GOOS == osLinux {
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

func (m *MetricsCollector) getCPUInfo() (string, string) {
	if runtime.GOOS != osLinux {
		return "", ""
	}

	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return "", ""
	}
	defer file.Close()

	var model, vendor string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "model name") && model == "" {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				model = strings.TrimSpace(parts[1])
			}
			continue
		}
		if strings.HasPrefix(line, "vendor_id") && vendor == "" {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				vendor = strings.TrimSpace(parts[1])
			}
		}
		if model != "" && vendor != "" {
			break
		}
	}

	return model, vendor
}

func (m *MetricsCollector) getLoadAverage() (float64, float64, float64) {
	if runtime.GOOS == osLinux {
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
	if runtime.GOOS == osLinux {
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
					//nolint:gosec // G115: percentage is 0-100, safe for int32
					return int32(100 * used / total)
				}
				break
			}
		}
	}

	return 0
}

func (m *MetricsCollector) getGPUInfo() (int32, int32, string) {
	// Resolve and validate nvidia-smi
	nvidiaSmiPath, err := security.ResolveAndValidateExecutable("system", "nvidia-smi")
	if err != nil {
		return 0, 0, ""
	}

	// Execute with fixed arguments
	//nolint:gosec // G204: Executable path validated, args are fixed literals
	output, err := exec.Command(nvidiaSmiPath, "--query-gpu=count,name", "--format=csv,noheader").Output()
	if err != nil {
		return 0, 0, ""
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return 0, 0, ""
	}

	//nolint:gosec // G115: GPU count is bounded (realistically < 100)
	count := int32(len(lines))
	gpuType := ""
	if parts := strings.SplitN(lines[0], ",", 2); len(parts) == 2 {
		gpuType = strings.TrimSpace(parts[1])
	}

	return count, count, gpuType // Simplified: all GPUs available
}

func (m *MetricsCollector) getStorageInfo(path string) (uint64, uint64) {
	if runtime.GOOS == osLinux {
		// Validate and sanitize the path argument
		args, err := security.DfArgs(path)
		if err != nil {
			return 0, 0
		}

		// Resolve and validate df executable
		dfPath, err := security.ResolveAndValidateExecutable("system", "df")
		if err != nil {
			return 0, 0
		}

		// Execute with validated path and arguments
		//nolint:gosec // G204: Executable path and arguments validated by security package
		output, err := exec.Command(dfPath, args...).Output()
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
		return osUnknown
	}

	// Validate hostname before using in command
	if err := security.ValidateHostname(hostname); err != nil {
		return osUnknown
	}

	// Build validated arguments for sinfo command
	args, err := security.SLURMSinfoArgs("%T", hostname)
	if err != nil {
		return osUnknown
	}

	// Resolve and validate sinfo executable
	sinfoPath, err := security.ResolveAndValidateExecutable("slurm", "sinfo")
	if err != nil {
		return osUnknown
	}

	// Execute with validated path and arguments
	//nolint:gosec // G204: Executable path and arguments validated by security package
	output, err := exec.Command(sinfoPath, args...).Output()
	if err != nil {
		return osUnknown
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

		// Resolve and validate ping executable
		pingPath, pingErr := security.ResolveAndValidateExecutable("system", "ping")
		if pingErr != nil {
			return nil
		}

		// Try ICMP ping if available
		//nolint:gosec // G204: Executable path and arguments validated by security package
		_, execErr := exec.Command(pingPath, args...).Output()
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
