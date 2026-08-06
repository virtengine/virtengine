// Package main implements the VirtEngine HPC Node Agent.
//
// VE-500: Metrics collection for node agent heartbeats.
// VE-7A: Command injection prevention and input sanitization
package main

import (
	"bufio"
	"fmt"
	"math"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/security"
)

const (
	osLinux          = "linux"
	osUnknown        = "unknown"
	maxTrackedJobIDs = 25
	maxInt32Value    = int32(^uint32(0) >> 1)
)

type readFileFunc func(string) ([]byte, error)
type resolveExecutableFunc func(string, string) (string, error)
type execCommandFunc func(string, ...string) ([]byte, error)
type hostnameFunc func() (string, error)

type cpuSample struct {
	total uint64
	idle  uint64
}

type diskSample struct {
	collectedAt time.Time
	busyMillis  uint64
}

type networkSample struct {
	collectedAt    time.Time
	totalBytes     uint64
	linkBitsPerSec float64
}

type gpuSnapshot struct {
	total          int32
	available      int32
	allocated      int32
	gpuType        string
	utilPercent    int32
	memUtilPercent int32
	memoryUsedMiB  int32
	temperatureC   int32
}

// MetricsCollector collects system metrics for heartbeats.
type MetricsCollector struct {
	startTime time.Time
	osName    string

	readFile          readFileFunc
	resolveExecutable resolveExecutableFunc
	execCommand       execCommandFunc
	hostname          hostnameFunc
	now               func() time.Time

	sampleMu          sync.Mutex
	lastCPUSample     *cpuSample
	lastDiskSample    *diskSample
	lastNetworkSample *networkSample
}

// NewMetricsCollector creates a new metrics collector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime:         time.Now(),
		osName:            runtime.GOOS,
		readFile:          os.ReadFile,
		resolveExecutable: security.ResolveAndValidateExecutable,
		execCommand:       defaultExecCommand,
		hostname:          os.Hostname,
		now:               time.Now,
	}
}

func defaultExecCommand(path string, args ...string) ([]byte, error) {
	//nolint:gosec // G204: callers provide validated executable paths and fixed/sanitized args
	return exec.Command(path, args...).Output()
}

// CollectCapacity collects node capacity metrics.
//
//nolint:unparam // result 1 (error) reserved for hardware query failures
func (m *MetricsCollector) CollectCapacity() (*NodeCapacity, error) {
	capacity := &NodeCapacity{}

	// CPU capacity is derived from the real scheduler pressure signal exposed by load average.
	//nolint:gosec // G115: NumCPU returns small positive int, safe for int32
	capacity.CPUCoresTotal = int32(runtime.NumCPU())
	capacity.CPUCoresAvailable, capacity.CPUCoresAllocated = availableCoresFromLoad(
		capacity.CPUCoresTotal,
		m.getLoadAverage1m(),
	)

	memTotal, memAvailable := m.getMemoryInfo()
	//nolint:gosec // G115: memory in GB is bounded well under int32 max
	capacity.MemoryGBTotal = int32(memTotal / (1024 * 1024 * 1024))
	//nolint:gosec // G115: memory in GB is bounded well under int32 max
	capacity.MemoryGBAvailable = int32(memAvailable / (1024 * 1024 * 1024))
	capacity.MemoryGBAllocated = clampInt32(capacity.MemoryGBTotal-capacity.MemoryGBAvailable, capacity.MemoryGBTotal)

	gpu := m.getGPUSnapshot()
	capacity.GPUsTotal = gpu.total
	capacity.GPUsAvailable = gpu.available
	capacity.GPUsAllocated = gpu.allocated
	capacity.GPUType = gpu.gpuType

	storageTotal, storageAvailable := m.getStorageInfo("/")
	//nolint:gosec // G115: storage in GB is bounded well under int32 max
	capacity.StorageGBTotal = int32(storageTotal / (1024 * 1024 * 1024))
	//nolint:gosec // G115: storage in GB is bounded well under int32 max
	capacity.StorageGBAvailable = int32(storageAvailable / (1024 * 1024 * 1024))
	capacity.StorageGBAllocated = clampInt32(capacity.StorageGBTotal-capacity.StorageGBAvailable, capacity.StorageGBTotal)

	return capacity, nil
}

// CollectHealth collects node health metrics.
//
//nolint:unparam // result 1 (error) reserved for health check failures
func (m *MetricsCollector) CollectHealth() (*NodeHealth, error) {
	now := m.now()
	health := &NodeHealth{
		Status:        "healthy",
		UptimeSeconds: int64(now.Sub(m.startTime).Seconds()),
	}

	load1, load5, load15 := m.getLoadAverage()
	health.LoadAverage1m = fmt.Sprintf("%.6f", load1)
	health.LoadAverage5m = fmt.Sprintf("%.6f", load5)
	health.LoadAverage15m = fmt.Sprintf("%.6f", load15)

	if cpuUtil, ok := m.sampleCPUUtilization(); ok {
		health.CPUUtilizationPercent = cpuUtil
	}

	memTotal, memAvailable := m.getMemoryInfo()
	if memTotal > 0 {
		health.MemoryUtilizationPercent = percentFromFraction(memTotal-memAvailable, memTotal)
	}

	gpu := m.getGPUSnapshot()
	health.GPUUtilizationPercent = gpu.utilPercent
	health.GPUMemoryUtilizationPercent = gpu.memUtilPercent
	health.GPUTemperatureCelsius = gpu.temperatureC

	health.DiskIOUtilizationPercent = m.getDiskIOUtilization(now)
	health.NetworkUtilizationPercent = m.getNetworkUtilization(now)
	health.SLURMState = m.getSLURMNodeState()
	health.Status = determineHealthStatus(health, runtime.NumCPU())

	return health, nil
}

// CollectHardware collects node hardware details.
//
//nolint:unparam // Signature reserved for future hardware probes that may return errors.
func (m *MetricsCollector) CollectHardware() (*NodeHardware, error) {
	hardware := &NodeHardware{
		CPUArch: runtime.GOARCH,
	}

	if m.osName == osLinux {
		model, vendor := m.getCPUInfo()
		hardware.CPUModel = model
		hardware.CPUVendor = vendor
	}

	hardware.GPUModel = m.getGPUSnapshot().gpuType

	return hardware, nil
}

// CollectLatency collects latency measurements.
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

	if len(latency.Measurements) > 0 {
		var total int64
		for _, probe := range latency.Measurements {
			total += probe.LatencyUs
		}
		latency.AvgClusterLatency = total / int64(len(latency.Measurements))
	}

	return latency
}

// CollectJobs collects job information.
func (m *MetricsCollector) CollectJobs() *NodeJobs {
	jobs := &NodeJobs{}

	args, err := security.SLURMSqueueArgs("%i,%T", "", "")
	if err != nil {
		return jobs
	}

	output, err := m.runValidatedCommand("slurm", "squeue", args...)
	if err != nil {
		return jobs
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ",", 2)
		if len(parts) != 2 {
			continue
		}

		jobID := strings.TrimSpace(parts[0])
		switch classifySLURMJobState(parts[1]) {
		case "running":
			jobs.RunningCount++
			if jobID != "" && len(jobs.ActiveJobIDs) < maxTrackedJobIDs {
				jobs.ActiveJobIDs = append(jobs.ActiveJobIDs, jobID)
			}
		case "pending":
			jobs.PendingCount++
		}
	}

	return jobs
}

// CollectServices collects service status.
func (m *MetricsCollector) CollectServices() *NodeServices {
	services := &NodeServices{}

	if _, err := m.runValidatedCommand("system", "pgrep", "-x", "slurmd"); err == nil {
		services.SLURMDRunning = true
	}

	if output, err := m.runValidatedCommand("slurm", "slurmd", "--version"); err == nil {
		services.SLURMDVersion = strings.TrimSpace(string(output))
	}

	if _, err := m.runValidatedCommand("system", "pgrep", "-x", "munged"); err == nil {
		services.MungeRunning = true
	}

	if output, err := m.runValidatedCommand("system", "singularity", "--version"); err == nil {
		services.ContainerRuntime = "singularity"
		services.ContainerRuntimeVersion = strings.TrimSpace(string(output))
	} else if output, err := m.runValidatedCommand("system", "docker", "--version"); err == nil {
		services.ContainerRuntime = "docker"
		services.ContainerRuntimeVersion = strings.TrimSpace(string(output))
	}

	return services
}

func (m *MetricsCollector) runValidatedCommand(category string, name string, args ...string) ([]byte, error) {
	path, err := m.resolveExecutable(category, name)
	if err != nil {
		return nil, err
	}
	return m.execCommand(path, args...)
}

func (m *MetricsCollector) getMemoryInfo() (uint64, uint64) {
	if m.osName != osLinux {
		return 16 * 1024 * 1024 * 1024, 8 * 1024 * 1024 * 1024
	}

	data, err := m.readFile("/proc/meminfo")
	if err != nil {
		return 0, 0
	}

	var memTotal uint64
	var memAvailable uint64
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}

		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		switch fields[0] {
		case "MemTotal:":
			memTotal = value * 1024
		case "MemAvailable:":
			memAvailable = value * 1024
		}
	}

	if memAvailable == 0 {
		memAvailable = memTotal
	}

	return memTotal, memAvailable
}

func (m *MetricsCollector) getCPUInfo() (string, string) {
	if m.osName != osLinux {
		return "", ""
	}

	data, err := m.readFile("/proc/cpuinfo")
	if err != nil {
		return "", ""
	}

	var model string
	var vendor string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
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
	if m.osName != osLinux {
		return 0, 0, 0
	}

	data, err := m.readFile("/proc/loadavg")
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

func (m *MetricsCollector) getLoadAverage1m() float64 {
	load1, _, _ := m.getLoadAverage()
	return load1
}

func (m *MetricsCollector) sampleCPUUtilization() (int32, bool) {
	if m.osName != osLinux {
		return 0, false
	}

	data, err := m.readFile("/proc/stat")
	if err != nil {
		return 0, false
	}

	current, err := parseCPUSample(data)
	if err != nil {
		return 0, false
	}

	m.sampleMu.Lock()
	defer m.sampleMu.Unlock()

	if m.lastCPUSample == nil {
		m.lastCPUSample = &current
		return 0, false
	}

	prev := *m.lastCPUSample
	m.lastCPUSample = &current

	if current.total <= prev.total || current.idle < prev.idle {
		return 0, false
	}

	deltaTotal := current.total - prev.total
	deltaIdle := current.idle - prev.idle
	if deltaTotal == 0 || deltaIdle > deltaTotal {
		return 0, false
	}

	return percentFromFraction(deltaTotal-deltaIdle, deltaTotal), true
}

func (m *MetricsCollector) getGPUSnapshot() gpuSnapshot {
	output, err := m.runValidatedCommand(
		"system",
		"nvidia-smi",
		"--query-gpu=name,utilization.gpu,utilization.memory,memory.total,memory.used,temperature.gpu",
		"--format=csv,noheader,nounits",
	)
	if err != nil {
		return gpuSnapshot{}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && strings.TrimSpace(lines[0]) == "") {
		return gpuSnapshot{}
	}

	var total int32
	var available int32
	var totalUtil int32
	var totalMemUtil int32
	var totalTemp int32
	gpuType := ""
	mixedTypes := false

	for _, line := range lines {
		parsed, ok := parseNvidiaSMILine(line)
		if !ok {
			continue
		}

		total++
		totalUtil += parsed.utilPercent
		totalMemUtil += parsed.memUtilPercent
		totalTemp += parsed.temperatureC

		if gpuType == "" {
			gpuType = parsed.gpuType
		} else if gpuType != parsed.gpuType {
			mixedTypes = true
		}

		if parsed.utilPercent <= 10 && parsed.memUtilPercent <= 10 && parsed.memoryUsedMiB <= 512 {
			available++
		}
	}

	if total == 0 {
		return gpuSnapshot{}
	}

	if mixedTypes {
		gpuType = "mixed"
	}

	return gpuSnapshot{
		total:          total,
		available:      available,
		allocated:      clampInt32(total-available, total),
		gpuType:        gpuType,
		utilPercent:    clampInt32(totalUtil/total, 100),
		memUtilPercent: clampInt32(totalMemUtil/total, 100),
		temperatureC:   clampInt32(totalTemp/total, maxInt32Value),
	}
}

func parseNvidiaSMILine(line string) (gpuSnapshot, bool) {
	parts := strings.Split(line, ",")
	if len(parts) < 6 {
		return gpuSnapshot{}, false
	}

	utilPercent, err := parseInt32(parts[1])
	if err != nil {
		return gpuSnapshot{}, false
	}
	memUtilPercent, err := parseInt32(parts[2])
	if err != nil {
		return gpuSnapshot{}, false
	}
	_, err = parseInt32(parts[3])
	if err != nil {
		return gpuSnapshot{}, false
	}
	memoryUsedMiB, err := parseInt32(parts[4])
	if err != nil {
		return gpuSnapshot{}, false
	}
	temperatureC, err := parseInt32(parts[5])
	if err != nil {
		return gpuSnapshot{}, false
	}

	return gpuSnapshot{
		gpuType:        strings.TrimSpace(parts[0]),
		utilPercent:    clampInt32(utilPercent, 100),
		memUtilPercent: clampInt32(memUtilPercent, 100),
		memoryUsedMiB:  clampInt32(memoryUsedMiB, maxInt32Value),
		temperatureC:   clampInt32(temperatureC, maxInt32Value),
	}, true
}

func (m *MetricsCollector) getStorageInfo(path string) (uint64, uint64) {
	if m.osName != osLinux {
		return 1000 * 1024 * 1024 * 1024, 500 * 1024 * 1024 * 1024
	}

	args, err := security.DfArgs(path)
	if err != nil {
		return 0, 0
	}

	output, err := m.runValidatedCommand("system", "df", args...)
	if err != nil {
		return 0, 0
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return 0, 0
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return 0, 0
	}

	total, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0, 0
	}
	available, err := strconv.ParseUint(fields[3], 10, 64)
	if err != nil {
		return 0, 0
	}

	return total, available
}

func (m *MetricsCollector) getDiskIOUtilization(now time.Time) int32 {
	if m.osName != osLinux {
		return 0
	}

	data, err := m.readFile("/proc/diskstats")
	if err != nil {
		return 0
	}

	busyMillis := parseDiskBusyMillis(data)

	m.sampleMu.Lock()
	defer m.sampleMu.Unlock()

	if m.lastDiskSample == nil {
		m.lastDiskSample = &diskSample{collectedAt: now, busyMillis: busyMillis}
		return 0
	}

	prev := *m.lastDiskSample
	m.lastDiskSample = &diskSample{collectedAt: now, busyMillis: busyMillis}
	if now.Before(prev.collectedAt) || busyMillis < prev.busyMillis {
		return 0
	}

	elapsedMillis := now.Sub(prev.collectedAt).Milliseconds()
	if elapsedMillis <= 0 {
		return 0
	}

	deltaBusyMillis := busyMillis - prev.busyMillis
	return clampPercentFloat64((float64(deltaBusyMillis) / float64(elapsedMillis)) * 100)
}

func (m *MetricsCollector) getNetworkUtilization(now time.Time) int32 {
	if m.osName != osLinux {
		return 0
	}

	totalBytes, linkBitsPerSec, err := m.readNetworkSnapshot()
	if err != nil || linkBitsPerSec <= 0 {
		return 0
	}

	m.sampleMu.Lock()
	defer m.sampleMu.Unlock()

	if m.lastNetworkSample == nil {
		m.lastNetworkSample = &networkSample{
			collectedAt:    now,
			totalBytes:     totalBytes,
			linkBitsPerSec: linkBitsPerSec,
		}
		return 0
	}

	prev := *m.lastNetworkSample
	m.lastNetworkSample = &networkSample{
		collectedAt:    now,
		totalBytes:     totalBytes,
		linkBitsPerSec: linkBitsPerSec,
	}

	if now.Before(prev.collectedAt) || totalBytes < prev.totalBytes {
		return 0
	}

	elapsedSeconds := now.Sub(prev.collectedAt).Seconds()
	if elapsedSeconds <= 0 {
		return 0
	}

	deltaBytes := totalBytes - prev.totalBytes
	bitsPerSecond := float64(deltaBytes) * 8 / elapsedSeconds
	return clampPercentFloat64((bitsPerSecond / linkBitsPerSec) * 100)
}

func (m *MetricsCollector) readNetworkSnapshot() (uint64, float64, error) {
	data, err := m.readFile("/proc/net/dev")
	if err != nil {
		return 0, 0, err
	}

	var totalBytes uint64
	var totalLinkBitsPerSec float64
	lines := strings.Split(string(data), "\n")
	for _, line := range lines[2:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		iface := strings.TrimSpace(parts[0])
		if !isTrackedNetworkInterface(iface) {
			continue
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 16 {
			continue
		}

		rxBytes, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			continue
		}
		txBytes, err := strconv.ParseUint(fields[8], 10, 64)
		if err != nil {
			continue
		}

		speedData, err := m.readFile(fmt.Sprintf("/sys/class/net/%s/speed", iface))
		if err != nil {
			continue
		}

		speedMbps, err := strconv.ParseFloat(strings.TrimSpace(string(speedData)), 64)
		if err != nil || speedMbps <= 0 {
			continue
		}

		totalBytes += rxBytes + txBytes
		totalLinkBitsPerSec += speedMbps * 1_000_000
	}

	if totalLinkBitsPerSec <= 0 {
		return 0, 0, fmt.Errorf("no usable network interfaces")
	}

	return totalBytes, totalLinkBitsPerSec, nil
}

func isTrackedNetworkInterface(name string) bool {
	switch {
	case name == "" || name == "lo":
		return false
	case strings.HasPrefix(name, "docker"),
		strings.HasPrefix(name, "veth"),
		strings.HasPrefix(name, "cni"),
		strings.HasPrefix(name, "flannel"),
		strings.HasPrefix(name, "br-"),
		strings.HasPrefix(name, "virbr"),
		strings.HasPrefix(name, "tun"),
		strings.HasPrefix(name, "tap"),
		strings.HasPrefix(name, "wg"):
		return false
	default:
		return true
	}
}

func parseDiskBusyMillis(data []byte) uint64 {
	var totalBusy uint64
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 14 {
			continue
		}
		if !isTrackedBlockDevice(fields[2]) {
			continue
		}

		busyMillis, err := strconv.ParseUint(fields[12], 10, 64)
		if err != nil {
			continue
		}
		totalBusy += busyMillis
	}

	return totalBusy
}

func isTrackedBlockDevice(name string) bool {
	switch {
	case strings.HasPrefix(name, "loop"), strings.HasPrefix(name, "ram"), strings.HasPrefix(name, "fd"):
		return false
	case strings.HasPrefix(name, "sd"):
		return len(name) == 3
	case strings.HasPrefix(name, "vd"):
		return len(name) == 3
	case strings.HasPrefix(name, "xvd"):
		return len(name) == 4
	case strings.HasPrefix(name, "nvme"):
		return !strings.Contains(name, "p")
	case strings.HasPrefix(name, "mmcblk"):
		return !strings.Contains(name, "p")
	case strings.HasPrefix(name, "dm-"), strings.HasPrefix(name, "md"):
		return true
	default:
		return false
	}
}

func (m *MetricsCollector) getSLURMNodeState() string {
	hostname, err := m.hostname()
	if err != nil {
		return osUnknown
	}

	if err := security.ValidateHostname(hostname); err != nil {
		return osUnknown
	}

	args, err := security.SLURMSinfoArgs("%T", hostname)
	if err != nil {
		return osUnknown
	}

	output, err := m.runValidatedCommand("slurm", "sinfo", args...)
	if err != nil {
		return osUnknown
	}

	return normalizeSLURMState(string(output))
}

func normalizeSLURMState(raw string) string {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	switch {
	case normalized == "":
		return osUnknown
	case strings.Contains(normalized, "DRAIN"):
		return nodeCommandDrain
	case strings.Contains(normalized, "DOWN"),
		strings.Contains(normalized, "FAIL"),
		strings.Contains(normalized, "NOT_RESPONDING"),
		strings.Contains(normalized, "POWER_DOWN"),
		strings.Contains(normalized, "POWERED_DOWN"),
		strings.Contains(normalized, "POWERING_DOWN"),
		strings.Contains(normalized, "MAINT"):
		return "down"
	case strings.Contains(normalized, "MIXED"):
		return "mixed"
	case strings.Contains(normalized, "ALLOC"):
		return "allocated"
	case strings.Contains(normalized, "IDLE"):
		return "idle"
	default:
		return osUnknown
	}
}

func classifySLURMJobState(raw string) string {
	state := strings.ToUpper(strings.TrimSpace(raw))
	switch {
	case strings.Contains(state, "RUNNING"),
		strings.Contains(state, "COMPLETING"),
		strings.Contains(state, "CONFIGURING"),
		strings.Contains(state, "SUSPENDED"):
		return "running"
	case strings.Contains(state, "PENDING"),
		strings.Contains(state, "RESIZING"):
		return "pending"
	default:
		return "other"
	}
}

func determineHealthStatus(health *NodeHealth, cpuCores int) string {
	switch health.SLURMState {
	case "down":
		return "offline"
	case nodeCommandDrain:
		return "draining"
	}

	maxUtilization := maxInt32(
		health.CPUUtilizationPercent,
		health.MemoryUtilizationPercent,
		health.GPUUtilizationPercent,
		health.GPUMemoryUtilizationPercent,
		health.DiskIOUtilizationPercent,
		health.NetworkUtilizationPercent,
	)

	if maxUtilization >= 95 {
		return "unhealthy"
	}
	if maxUtilization >= 85 {
		return "degraded"
	}

	if cpuCores > 0 {
		load1, err := strconv.ParseFloat(health.LoadAverage1m, 64)
		if err == nil && load1 > float64(cpuCores)*1.25 {
			return "degraded"
		}
	}

	return "healthy"
}

func availableCoresFromLoad(totalCores int32, load1 float64) (int32, int32) {
	if totalCores <= 0 {
		return 0, 0
	}

	allocated := clampInt32(safeInt32FromFloat64Round(load1), totalCores)
	return totalCores - allocated, allocated
}

func parseCPUSample(data []byte) (cpuSample, error) {
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 8 {
			break
		}

		var values []uint64
		for _, field := range fields[1:] {
			value, err := strconv.ParseUint(field, 10, 64)
			if err != nil {
				return cpuSample{}, err
			}
			values = append(values, value)
		}

		var total uint64
		for _, value := range values {
			total += value
		}

		idle := values[3]
		if len(values) > 4 {
			idle += values[4]
		}

		return cpuSample{total: total, idle: idle}, nil
	}

	return cpuSample{}, fmt.Errorf("cpu line not found")
}

func percentFromFraction(numerator uint64, denominator uint64) int32 {
	if denominator == 0 {
		return 0
	}
	return clampPercentFloat64((float64(numerator) / float64(denominator)) * 100)
}

func clampPercentFloat64(v float64) int32 {
	if math.IsNaN(v) || v <= 0 {
		return 0
	}
	if v >= 100 {
		return 100
	}
	return int32(math.Round(v))
}

func clampInt32(v int32, max int32) int32 {
	if v < 0 {
		return 0
	}
	if v > max {
		return max
	}
	return v
}

func safeInt32FromInt(value int) int32 {
	if value < 0 {
		return 0
	}
	if value > int(maxInt32Value) {
		return maxInt32Value
	}
	return int32(value)
}

func safeInt32FromFloat64Round(value float64) int32 {
	if math.IsNaN(value) || value <= 0 {
		return 0
	}
	if value >= float64(maxInt32Value) {
		return maxInt32Value
	}
	return int32(math.Round(value))
}

func maxInt32(values ...int32) int32 {
	var maxValue int32
	for i, value := range values {
		if i == 0 || value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}

func parseInt32(raw string) (int32, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(value), nil
}

func (m *MetricsCollector) measureLatency(target string) *LatencyProbe {
	start := m.now()

	if err := security.ValidatePingTarget(target); err != nil {
		return nil
	}

	conn, err := net.DialTimeout("tcp", target+":22", 5*time.Second)
	if err != nil {
		args, pingErr := security.PingArgs(target, 1)
		if pingErr != nil {
			return nil
		}

		if _, pingErr = m.runValidatedCommand("system", "ping", args...); pingErr != nil {
			return nil
		}
	} else {
		_ = conn.Close()
	}

	latency := m.now().Sub(start)
	return &LatencyProbe{
		TargetNodeID:      target,
		LatencyUs:         latency.Microseconds(),
		PacketLossPercent: 0,
		MeasuredAt:        m.now(),
	}
}
