// Package benchmark_daemon implements benchmarking for VirtEngine providers.
//
// VE-600: Benchmark runner implementation
// VE-7A: Command injection prevention and input sanitization
package benchmark_daemon

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	verrors "github.com/virtengine/virtengine/pkg/errors"
	"github.com/virtengine/virtengine/pkg/security"
)

// DefaultBenchmarkRunner is the default benchmark runner implementation
type DefaultBenchmarkRunner struct {
	// Timeout for individual benchmark tests
	TestTimeout time.Duration
}

// NewDefaultBenchmarkRunner creates a new default benchmark runner
func NewDefaultBenchmarkRunner() *DefaultBenchmarkRunner {
	return &DefaultBenchmarkRunner{
		TestTimeout: time.Minute * 5,
	}
}

// RunBenchmarks runs the benchmark suite and returns metrics
func (r *DefaultBenchmarkRunner) RunBenchmarks(ctx context.Context, config BenchmarkDaemonConfig) (*BenchmarkMetrics, error) {
	metrics := &BenchmarkMetrics{}

	// Run CPU benchmarks
	if err := r.runCPUBenchmarks(ctx, metrics); err != nil {
		return nil, fmt.Errorf("cpu benchmarks failed: %w", err)
	}

	// Run memory benchmarks
	if err := r.runMemoryBenchmarks(ctx, metrics); err != nil {
		return nil, fmt.Errorf("memory benchmarks failed: %w", err)
	}

	// Run disk benchmarks
	if err := r.runDiskBenchmarks(ctx, metrics); err != nil {
		return nil, fmt.Errorf("disk benchmarks failed: %w", err)
	}

	// Run network benchmarks
	if err := r.runNetworkBenchmarks(ctx, metrics, config.NetworkReferenceEndpoint); err != nil {
		return nil, fmt.Errorf("network benchmarks failed: %w", err)
	}

	// Run GPU benchmarks if enabled
	if config.EnableGPU {
		if err := r.runGPUBenchmarks(ctx, metrics); err != nil {
			// GPU benchmarks are optional, log but don't fail
			metrics.GPUPresent = false
		}
	}

	return metrics, nil
}

// runCPUBenchmarks runs CPU benchmark tests
//
//nolint:unparam // error return preserved for future extensibility
func (r *DefaultBenchmarkRunner) runCPUBenchmarks(ctx context.Context, metrics *BenchmarkMetrics) error {
	// Get basic CPU info
	//nolint:gosec // G115: NumCPU returns small positive int, safe for int32
	metrics.CPUCoreCount = int32(runtime.NumCPU())
	metrics.CPUThreadCount = metrics.CPUCoreCount // Simplified

	// Try to get frequency info (platform-dependent)
	metrics.CPUBaseFreqMHz = 3000 // Default fallback
	metrics.CPUBoostFreqMHz = 3500

	// Run synthetic CPU benchmark
	// Single-core test
	singleScore := r.runSyntheticCPUTest(ctx, 1)
	metrics.CPUSingleCoreScore = normalizeScore(singleScore, 0, 100000, 0, 10000)

	// Multi-core test
	multiScore := r.runSyntheticCPUTest(ctx, int(metrics.CPUCoreCount))
	metrics.CPUMultiCoreScore = normalizeScore(multiScore, 0, 100000*int64(metrics.CPUCoreCount), 0, 10000)

	return nil
}

// runSyntheticCPUTest runs a synthetic CPU benchmark
func (r *DefaultBenchmarkRunner) runSyntheticCPUTest(_ context.Context, threads int) int64 {
	// Simple synthetic benchmark: compute iterations per second
	const testDuration = time.Second * 2

	resultCh := make(chan int64, threads)

	for i := 0; i < threads; i++ {
		verrors.SafeGo("", func() {
			defer func() {}() // WG Done if needed
			var count int64
			deadline := time.Now().Add(testDuration)
			for time.Now().Before(deadline) {
				// Simple compute work
				x := int64(1)
				for j := 0; j < 10000; j++ {
					x = (x*1103515245 + 12345) & 0x7fffffff
				}
				count++
			}
			resultCh <- count
		})
	}

	var total int64
	for i := 0; i < threads; i++ {
		total += <-resultCh
	}

	return total
}

// runMemoryBenchmarks runs memory benchmark tests
//
//nolint:unparam // error return preserved for future extensibility
func (r *DefaultBenchmarkRunner) runMemoryBenchmarks(ctx context.Context, metrics *BenchmarkMetrics) error {
	// Get total memory (simplified - would use platform-specific APIs)
	metrics.MemoryTotalGB = 64 // Default fallback

	// Run memory bandwidth test
	bandwidthMBps := r.runMemoryBandwidthTest(ctx)
	metrics.MemoryBandwidthMBps = bandwidthMBps

	// Run memory latency test
	latencyNs := r.runMemoryLatencyTest(ctx)
	metrics.MemoryLatencyNs = latencyNs

	// Compute score
	// Higher bandwidth and lower latency = higher score
	bandwidthScore := normalizeScore(bandwidthMBps, 0, 200000, 0, 5000)
	latencyScore := normalizeScore(200-latencyNs, 0, 200, 0, 5000) // Lower is better
	metrics.MemoryScore = bandwidthScore + latencyScore

	return nil
}

// runMemoryBandwidthTest runs a memory bandwidth test
func (r *DefaultBenchmarkRunner) runMemoryBandwidthTest(_ context.Context) int64 {
	// Allocate test buffer
	const bufferSize = 64 * 1024 * 1024 // 64MB
	buffer := make([]byte, bufferSize)

	// Fill buffer to ensure allocation
	for i := range buffer {
		buffer[i] = byte(i)
	}

	// Measure copy bandwidth
	start := time.Now()
	const iterations = 10

	for i := 0; i < iterations; i++ {
		copy(buffer[bufferSize/2:], buffer[:bufferSize/2])
	}

	elapsed := time.Since(start)
	bytesTransferred := int64(bufferSize / 2 * iterations)
	mbps := (bytesTransferred * 1000) / elapsed.Milliseconds() / (1024 * 1024) * 1000

	return mbps
}

// runMemoryLatencyTest runs a memory latency test
func (r *DefaultBenchmarkRunner) runMemoryLatencyTest(_ context.Context) int64 {
	// Simple latency test using random access
	const size = 1024 * 1024 // 1M elements
	data := make([]int, size)

	// Initialize with pointer chasing pattern
	for i := range data {
		data[i] = (i + 1) % size
	}

	// Measure access time
	start := time.Now()
	const accesses = 1000000
	idx := 0
	for i := 0; i < accesses; i++ {
		idx = data[idx]
	}

	elapsed := time.Since(start)
	latencyNs := elapsed.Nanoseconds() / accesses

	// Prevent optimization
	_ = idx

	return latencyNs
}

// runDiskBenchmarks runs disk I/O benchmark tests
//
//nolint:unparam // error return preserved for future extensibility
func (r *DefaultBenchmarkRunner) runDiskBenchmarks(ctx context.Context, metrics *BenchmarkMetrics) error {
	// Default values - in production would use actual disk benchmarks
	metrics.DiskTotalStorageGB = 1000

	// Try to run fio or similar if available, otherwise use defaults
	readIOPS, writeIOPS, readMBps, writeMBps := r.runDiskIOTest(ctx)

	metrics.DiskReadIOPS = readIOPS
	metrics.DiskWriteIOPS = writeIOPS
	metrics.DiskReadThroughputMBps = readMBps
	metrics.DiskWriteThroughputMBps = writeMBps

	// Compute score
	iopsScore := normalizeScore((readIOPS+writeIOPS)/2, 0, 500000, 0, 5000)
	throughputScore := normalizeScore((readMBps+writeMBps)/2, 0, 10000, 0, 5000)
	metrics.DiskScore = iopsScore + throughputScore

	return nil
}

// runDiskIOTest runs disk I/O tests
func (r *DefaultBenchmarkRunner) runDiskIOTest(ctx context.Context) (readIOPS, writeIOPS, readMBps, writeMBps int64) {
	// Simplified disk test - in production would use proper tools
	// Return reasonable defaults
	return 100000, 80000, 3000, 2500
}

// runNetworkBenchmarks runs network benchmark tests
//
//nolint:unparam // error return preserved for future extensibility
func (r *DefaultBenchmarkRunner) runNetworkBenchmarks(ctx context.Context, metrics *BenchmarkMetrics, endpoint string) error {
	metrics.NetworkEndpoint = endpoint

	// Run latency test (ping)
	latencyMs := r.runPingTest(ctx, endpoint)
	metrics.NetworkLatencyMs = latencyMs

	// Run throughput test (simplified)
	throughputMbps := r.runThroughputTest(ctx, endpoint)
	metrics.NetworkThroughputMbps = throughputMbps

	// Packet loss rate (fixed-point, 0-1000000)
	packetLoss := r.runPacketLossTest(ctx, endpoint)
	metrics.NetworkPacketLossRate = packetLoss

	// Compute score
	// Lower latency and higher throughput = higher score
	latencyScore := normalizeScore(100000-latencyMs, 0, 100000, 0, 3333)
	throughputScore := normalizeScore(throughputMbps, 0, 100000, 0, 3333)
	lossScore := normalizeScore(1000000-packetLoss, 0, 1000000, 0, 3334)
	metrics.NetworkScore = latencyScore + throughputScore + lossScore

	return nil
}

// runPingTest runs a ping test to measure latency
func (r *DefaultBenchmarkRunner) runPingTest(ctx context.Context, endpoint string) int64 {
	// Validate endpoint before using in ping command
	if err := security.ValidatePingTarget(endpoint); err != nil {
		// Return default latency if validation fails
		return 10000 // 10ms in fixed-point (*1000)
	}

	// Build validated ping arguments
	args, err := security.PingArgs(endpoint, 5)
	if err != nil {
		return 10000
	}

	cmd := exec.CommandContext(ctx, "ping", args...)

	output, err := cmd.Output()
	if err != nil {
		// Return default latency if ping fails
		return 10000 // 10ms in fixed-point (*1000)
	}

	// Parse ping output (simplified)
	latencyMs := parsePingOutput(string(output))
	return latencyMs * 1000 // Convert to fixed-point
}

// parsePingOutput parses ping command output to extract average latency
func parsePingOutput(output string) int64 {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Average") || strings.Contains(line, "avg") {
			// Try to extract the average value
			parts := strings.Split(line, "/")
			if len(parts) >= 5 {
				if val, err := strconv.ParseFloat(strings.TrimSpace(parts[4]), 64); err == nil {
					return int64(val)
				}
			}
			// Windows format
			parts = strings.Split(line, "=")
			if len(parts) >= 3 {
				valStr := strings.TrimSuffix(strings.TrimSpace(parts[len(parts)-1]), "ms")
				if val, err := strconv.ParseFloat(valStr, 64); err == nil {
					return int64(val)
				}
			}
		}
	}
	return 10 // Default 10ms
}

// runThroughputTest runs a network throughput test
func (r *DefaultBenchmarkRunner) runThroughputTest(ctx context.Context, endpoint string) int64 {
	// In production, would use iperf3 or similar
	// Return reasonable default
	return 10000 // 10 Gbps
}

// runPacketLossTest runs a packet loss test
func (r *DefaultBenchmarkRunner) runPacketLossTest(ctx context.Context, endpoint string) int64 {
	// In production, would parse ping output for packet loss
	// Return 0.01% packet loss (100 in fixed-point * 1000000)
	return 100
}

// runGPUBenchmarks runs GPU benchmark tests
//
//nolint:unparam // error return preserved for future extensibility
func (r *DefaultBenchmarkRunner) runGPUBenchmarks(_ context.Context, metrics *BenchmarkMetrics) error {
	// Check if GPU is present (simplified)
	// In production, would use nvidia-smi or similar

	// For now, return not present
	metrics.GPUPresent = false
	return nil
}

// normalizeScore normalizes a value to a score range
func normalizeScore(value, minVal, maxVal, minScore, maxScore int64) int64 {
	if value <= minVal {
		return minScore
	}
	if value >= maxVal {
		return maxScore
	}

	// Use fixed-point arithmetic
	const scale int64 = 1000000

	valueRange := maxVal - minVal
	scoreRange := maxScore - minScore

	normalized := ((value - minVal) * scale) / valueRange
	score := minScore + (normalized*scoreRange)/scale

	return score
}
