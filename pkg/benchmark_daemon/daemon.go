// Package benchmark_daemon implements the benchmarking daemon for VirtEngine providers.
//
// VE-600: Benchmarking daemon - provider performance metrics collection
package benchmark_daemon

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrDaemonNotRunning is returned when the daemon is not running
var ErrDaemonNotRunning = errors.New("benchmark daemon is not running")

// ErrBenchmarkInProgress is returned when a benchmark is already in progress
var ErrBenchmarkInProgress = errors.New("benchmark already in progress")

// ErrInvalidConfig is returned when the configuration is invalid
var ErrInvalidConfig = errors.New("invalid configuration")

// ErrSubmissionFailed is returned when benchmark submission fails
var ErrSubmissionFailed = errors.New("benchmark submission failed")

// ErrRateLimited is returned when submissions are rate limited
var ErrRateLimited = errors.New("benchmark submission rate limited")

// BenchmarkDaemonConfig configures the benchmark daemon
type BenchmarkDaemonConfig struct {
	// ProviderAddress is the provider's blockchain address
	ProviderAddress string `json:"provider_address"`

	// ClusterID is the cluster being benchmarked
	ClusterID string `json:"cluster_id"`

	// Region is the geographic region
	Region string `json:"region"`

	// ScheduleInterval is the interval between scheduled benchmarks
	ScheduleInterval time.Duration `json:"schedule_interval"`

	// ChallengeCheckInterval is the interval to check for challenges
	ChallengeCheckInterval time.Duration `json:"challenge_check_interval"`

	// ChainEndpoint is the blockchain RPC endpoint
	ChainEndpoint string `json:"chain_endpoint"`

	// MaxRetries is the maximum number of submission retries
	MaxRetries int `json:"max_retries"`

	// RetryDelay is the delay between retries
	RetryDelay time.Duration `json:"retry_delay"`

	// SuiteVersion is the benchmark suite version
	SuiteVersion string `json:"suite_version"`

	// EnableGPU enables GPU benchmarking
	EnableGPU bool `json:"enable_gpu"`

	// NetworkReferenceEndpoint is the endpoint for network benchmarks
	NetworkReferenceEndpoint string `json:"network_reference_endpoint"`
}

// DefaultBenchmarkDaemonConfig returns the default configuration
func DefaultBenchmarkDaemonConfig() BenchmarkDaemonConfig {
	return BenchmarkDaemonConfig{
		ScheduleInterval:         time.Hour * 6,
		ChallengeCheckInterval:   time.Minute * 5,
		MaxRetries:               3,
		RetryDelay:               time.Second * 30,
		SuiteVersion:             "1.0.0",
		EnableGPU:                false,
		NetworkReferenceEndpoint: "benchmark.virtengine.com",
	}
}

// Validate validates the configuration
func (c *BenchmarkDaemonConfig) Validate() error {
	if c.ProviderAddress == "" {
		return fmt.Errorf("%w: provider_address is required", ErrInvalidConfig)
	}
	if c.ClusterID == "" {
		return fmt.Errorf("%w: cluster_id is required", ErrInvalidConfig)
	}
	if c.ScheduleInterval <= 0 {
		return fmt.Errorf("%w: schedule_interval must be positive", ErrInvalidConfig)
	}
	if c.ChainEndpoint == "" {
		return fmt.Errorf("%w: chain_endpoint is required", ErrInvalidConfig)
	}
	return nil
}

// BenchmarkResult contains the result of a benchmark run
type BenchmarkResult struct {
	// ReportID is the unique identifier for this report
	ReportID string `json:"report_id"`

	// Success indicates if the benchmark completed successfully
	Success bool `json:"success"`

	// Error contains any error message
	Error string `json:"error,omitempty"`

	// Metrics contains the benchmark metrics
	Metrics *BenchmarkMetrics `json:"metrics,omitempty"`

	// SummaryScore is the normalized summary score
	SummaryScore int64 `json:"summary_score,omitempty"`

	// Timestamp is when the benchmark completed
	Timestamp time.Time `json:"timestamp"`

	// Duration is how long the benchmark took
	Duration time.Duration `json:"duration"`

	// Submitted indicates if the report was submitted to chain
	Submitted bool `json:"submitted"`
}

// BenchmarkMetrics contains collected benchmark metrics
type BenchmarkMetrics struct {
	// CPU metrics
	CPUSingleCoreScore int64 `json:"cpu_single_core_score"`
	CPUMultiCoreScore  int64 `json:"cpu_multi_core_score"`
	CPUCoreCount       int32 `json:"cpu_core_count"`
	CPUThreadCount     int32 `json:"cpu_thread_count"`
	CPUBaseFreqMHz     int64 `json:"cpu_base_freq_mhz"`
	CPUBoostFreqMHz    int64 `json:"cpu_boost_freq_mhz"`

	// Memory metrics
	MemoryTotalGB       int64 `json:"memory_total_gb"`
	MemoryBandwidthMBps int64 `json:"memory_bandwidth_mbps"`
	MemoryLatencyNs     int64 `json:"memory_latency_ns"`
	MemoryScore         int64 `json:"memory_score"`

	// Disk metrics
	DiskReadIOPS            int64 `json:"disk_read_iops"`
	DiskWriteIOPS           int64 `json:"disk_write_iops"`
	DiskReadThroughputMBps  int64 `json:"disk_read_throughput_mbps"`
	DiskWriteThroughputMBps int64 `json:"disk_write_throughput_mbps"`
	DiskTotalStorageGB      int64 `json:"disk_total_storage_gb"`
	DiskScore               int64 `json:"disk_score"`

	// Network metrics
	NetworkThroughputMbps int64  `json:"network_throughput_mbps"`
	NetworkLatencyMs      int64  `json:"network_latency_ms"`
	NetworkPacketLossRate int64  `json:"network_packet_loss_rate"`
	NetworkEndpoint       string `json:"network_endpoint"`
	NetworkScore          int64  `json:"network_score"`

	// GPU metrics (optional)
	GPUPresent             bool   `json:"gpu_present"`
	GPUDeviceCount         int32  `json:"gpu_device_count,omitempty"`
	GPUDeviceType          string `json:"gpu_device_type,omitempty"`
	GPUTotalMemoryGB       int64  `json:"gpu_total_memory_gb,omitempty"`
	GPUComputeScore        int64  `json:"gpu_compute_score,omitempty"`
	GPUMemoryBandwidthGBps int64  `json:"gpu_memory_bandwidth_gbps,omitempty"`
}

// ChainClient is an interface for chain interactions
type ChainClient interface {
	// SubmitBenchmarks submits benchmark reports to the chain
	SubmitBenchmarks(ctx context.Context, reports []json.RawMessage) error

	// GetPendingChallenges gets pending challenges for a provider
	GetPendingChallenges(ctx context.Context, providerAddr string) ([]Challenge, error)

	// RespondToChallenge responds to a benchmark challenge
	RespondToChallenge(ctx context.Context, challengeID string, report json.RawMessage, explanationRef string) error
}

// Challenge represents a benchmark challenge from the chain
type Challenge struct {
	ChallengeID          string    `json:"challenge_id"`
	ProviderAddress      string    `json:"provider_address"`
	ClusterID            string    `json:"cluster_id"`
	RequiredSuiteVersion string    `json:"required_suite_version"`
	SuiteHash            string    `json:"suite_hash"`
	Deadline             time.Time `json:"deadline"`
}

// BenchmarkRunner is an interface for running benchmarks
type BenchmarkRunner interface {
	// RunBenchmarks runs the benchmark suite and returns metrics
	RunBenchmarks(ctx context.Context, config BenchmarkDaemonConfig) (*BenchmarkMetrics, error)
}

// KeySigner is an interface for signing benchmark reports
type KeySigner interface {
	// Sign signs data and returns the signature
	Sign(data []byte) ([]byte, error)

	// PublicKey returns the public key
	PublicKey() []byte
}

// BenchmarkDaemon is the main benchmark daemon
type BenchmarkDaemon struct {
	config BenchmarkDaemonConfig
	client ChainClient
	runner BenchmarkRunner
	signer KeySigner

	running      bool
	runningMu    sync.RWMutex
	stopCh       chan struct{}
	inProgress   bool
	inProgressMu sync.Mutex

	// Tracking for rate limiting
	lastSubmit    time.Time
	submitCount   int
	submitCountMu sync.Mutex

	// Results history
	results   []BenchmarkResult
	resultsMu sync.RWMutex
}

// NewBenchmarkDaemon creates a new benchmark daemon
func NewBenchmarkDaemon(config BenchmarkDaemonConfig, client ChainClient, runner BenchmarkRunner, signer KeySigner) (*BenchmarkDaemon, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &BenchmarkDaemon{
		config:  config,
		client:  client,
		runner:  runner,
		signer:  signer,
		stopCh:  make(chan struct{}),
		results: make([]BenchmarkResult, 0),
	}, nil
}

// Start starts the benchmark daemon
func (d *BenchmarkDaemon) Start(ctx context.Context) error {
	d.runningMu.Lock()
	if d.running {
		d.runningMu.Unlock()
		return nil
	}
	d.running = true
	d.stopCh = make(chan struct{})
	d.runningMu.Unlock()

	// Start scheduled benchmark goroutine
	go d.scheduleLoop(ctx)

	// Start challenge check goroutine
	go d.challengeLoop(ctx)

	return nil
}

// Stop stops the benchmark daemon
func (d *BenchmarkDaemon) Stop() error {
	d.runningMu.Lock()
	defer d.runningMu.Unlock()

	if !d.running {
		return nil
	}

	close(d.stopCh)
	d.running = false
	return nil
}

// IsRunning returns whether the daemon is running
func (d *BenchmarkDaemon) IsRunning() bool {
	d.runningMu.RLock()
	defer d.runningMu.RUnlock()
	return d.running
}

// RunBenchmark runs a benchmark immediately
func (d *BenchmarkDaemon) RunBenchmark(ctx context.Context) (*BenchmarkResult, error) {
	if !d.IsRunning() {
		return nil, ErrDaemonNotRunning
	}

	d.inProgressMu.Lock()
	if d.inProgress {
		d.inProgressMu.Unlock()
		return nil, ErrBenchmarkInProgress
	}
	d.inProgress = true
	d.inProgressMu.Unlock()

	defer func() {
		d.inProgressMu.Lock()
		d.inProgress = false
		d.inProgressMu.Unlock()
	}()

	return d.runAndSubmitBenchmark(ctx, "")
}

// scheduleLoop runs scheduled benchmarks
func (d *BenchmarkDaemon) scheduleLoop(ctx context.Context) {
	ticker := time.NewTicker(d.config.ScheduleInterval)
	defer ticker.Stop()

	// Run initial benchmark
	_, _ = d.runAndSubmitBenchmark(ctx, "")

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		case <-ticker.C:
			_, _ = d.runAndSubmitBenchmark(ctx, "")
		}
	}
}

// challengeLoop checks for and responds to challenges
func (d *BenchmarkDaemon) challengeLoop(ctx context.Context) {
	ticker := time.NewTicker(d.config.ChallengeCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.handlePendingChallenges(ctx)
		}
	}
}

// handlePendingChallenges checks for and handles pending challenges
func (d *BenchmarkDaemon) handlePendingChallenges(ctx context.Context) {
	if d.client == nil {
		return
	}

	challenges, err := d.client.GetPendingChallenges(ctx, d.config.ProviderAddress)
	if err != nil {
		// Log error but don't fail - never log secrets
		return
	}

	for _, challenge := range challenges {
		// Check deadline
		if time.Now().After(challenge.Deadline) {
			continue
		}

		// Run benchmark for challenge
		result, err := d.runAndSubmitBenchmark(ctx, challenge.ChallengeID)
		if err != nil {
			continue
		}

		if result.Success {
			// Submit challenge response
			report := d.createReportJSON(result)
			if report != nil {
				_ = d.client.RespondToChallenge(ctx, challenge.ChallengeID, report, "")
			}
		}
	}
}

// runAndSubmitBenchmark runs a benchmark and submits it to the chain
func (d *BenchmarkDaemon) runAndSubmitBenchmark(ctx context.Context, _ string) (*BenchmarkResult, error) {
	startTime := time.Now()

	result := &BenchmarkResult{
		ReportID:  generateReportID(d.config.ProviderAddress, startTime),
		Timestamp: startTime,
	}

	// Run benchmarks
	metrics, err := d.runner.RunBenchmarks(ctx, d.config)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Duration = time.Since(startTime)
		d.addResult(*result)
		return result, err
	}

	result.Metrics = metrics
	result.SummaryScore = computeSummaryScore(metrics)
	result.Success = true
	result.Duration = time.Since(startTime)

	// Submit to chain with retry logic
	if d.client != nil {
		submitted := false
		var lastErr error

		for i := 0; i < d.config.MaxRetries; i++ {
			if err := d.checkRateLimit(); err != nil {
				lastErr = err
				time.Sleep(d.config.RetryDelay)
				continue
			}

			report := d.createReportJSON(result)
			if report == nil {
				lastErr = errors.New("failed to create report JSON")
				break
			}

			if err := d.client.SubmitBenchmarks(ctx, []json.RawMessage{report}); err != nil {
				lastErr = err
				time.Sleep(d.config.RetryDelay)
				continue
			}

			submitted = true
			d.recordSubmission()
			break
		}

		result.Submitted = submitted
		if !submitted && lastErr != nil {
			result.Error = lastErr.Error()
		}
	}

	d.addResult(*result)
	return result, nil
}

// createReportJSON creates a signed JSON report
func (d *BenchmarkDaemon) createReportJSON(result *BenchmarkResult) json.RawMessage {
	if result.Metrics == nil {
		return nil
	}

	report := map[string]interface{}{
		"report_id":        result.ReportID,
		"provider_address": d.config.ProviderAddress,
		"cluster_id":       d.config.ClusterID,
		"node_metadata": map[string]interface{}{
			"node_id": "node-1",
			"region":  d.config.Region,
			"os_type": "linux",
		},
		"suite_version": d.config.SuiteVersion,
		"suite_hash":    computeSuiteHash(d.config.SuiteVersion),
		"metrics": map[string]interface{}{
			"schema_version": "1.0.0",
			"cpu": map[string]interface{}{
				"single_core_score":   result.Metrics.CPUSingleCoreScore,
				"multi_core_score":    result.Metrics.CPUMultiCoreScore,
				"core_count":          result.Metrics.CPUCoreCount,
				"thread_count":        result.Metrics.CPUThreadCount,
				"base_frequency_mhz":  result.Metrics.CPUBaseFreqMHz,
				"boost_frequency_mhz": result.Metrics.CPUBoostFreqMHz,
			},
			"memory": map[string]interface{}{
				"total_gb":       result.Metrics.MemoryTotalGB,
				"bandwidth_mbps": result.Metrics.MemoryBandwidthMBps,
				"latency_ns":     result.Metrics.MemoryLatencyNs,
				"score":          result.Metrics.MemoryScore,
			},
			"disk": map[string]interface{}{
				"read_iops":             result.Metrics.DiskReadIOPS,
				"write_iops":            result.Metrics.DiskWriteIOPS,
				"read_throughput_mbps":  result.Metrics.DiskReadThroughputMBps,
				"write_throughput_mbps": result.Metrics.DiskWriteThroughputMBps,
				"total_storage_gb":      result.Metrics.DiskTotalStorageGB,
				"score":                 result.Metrics.DiskScore,
			},
			"network": map[string]interface{}{
				"throughput_mbps":    result.Metrics.NetworkThroughputMbps,
				"latency_ms":         result.Metrics.NetworkLatencyMs,
				"packet_loss_rate":   result.Metrics.NetworkPacketLossRate,
				"reference_endpoint": result.Metrics.NetworkEndpoint,
				"score":              result.Metrics.NetworkScore,
			},
			"gpu": map[string]interface{}{
				"present": result.Metrics.GPUPresent,
			},
		},
		"summary_score": result.SummaryScore,
		"timestamp":     result.Timestamp,
		"public_key":    hex.EncodeToString(d.signer.PublicKey()),
	}

	// Sign the report
	reportBytes, err := json.Marshal(report)
	if err != nil {
		return nil
	}

	hash := sha256.Sum256(reportBytes)
	sig, err := d.signer.Sign(hash[:])
	if err != nil {
		return nil
	}

	report["signature"] = hex.EncodeToString(sig)

	finalBytes, err := json.Marshal(report)
	if err != nil {
		return nil
	}

	return finalBytes
}

// checkRateLimit checks if we're within rate limits
func (d *BenchmarkDaemon) checkRateLimit() error {
	d.submitCountMu.Lock()
	defer d.submitCountMu.Unlock()

	// Reset counter if it's been over an hour
	if time.Since(d.lastSubmit) > time.Hour {
		d.submitCount = 0
	}

	// Allow max 10 submissions per hour
	if d.submitCount >= 10 {
		return ErrRateLimited
	}

	return nil
}

// recordSubmission records a successful submission
func (d *BenchmarkDaemon) recordSubmission() {
	d.submitCountMu.Lock()
	defer d.submitCountMu.Unlock()

	d.lastSubmit = time.Now()
	d.submitCount++
}

// addResult adds a result to the history
func (d *BenchmarkDaemon) addResult(result BenchmarkResult) {
	d.resultsMu.Lock()
	defer d.resultsMu.Unlock()

	d.results = append(d.results, result)

	// Keep only last 100 results
	if len(d.results) > 100 {
		d.results = d.results[len(d.results)-100:]
	}
}

// GetResults returns the benchmark results history
func (d *BenchmarkDaemon) GetResults() []BenchmarkResult {
	d.resultsMu.RLock()
	defer d.resultsMu.RUnlock()

	results := make([]BenchmarkResult, len(d.results))
	copy(results, d.results)
	return results
}

// GetLatestResult returns the most recent benchmark result
func (d *BenchmarkDaemon) GetLatestResult() (*BenchmarkResult, bool) {
	d.resultsMu.RLock()
	defer d.resultsMu.RUnlock()

	if len(d.results) == 0 {
		return nil, false
	}

	return &d.results[len(d.results)-1], true
}

// generateReportID generates a unique report ID
func generateReportID(providerAddr string, timestamp time.Time) string {
	data := fmt.Sprintf("%s-%d", providerAddr, timestamp.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

// computeSuiteHash computes the hash of the benchmark suite
func computeSuiteHash(version string) string {
	hash := sha256.Sum256([]byte("benchmark-suite-" + version))
	return hex.EncodeToString(hash[:16])
}

// computeSummaryScore computes a summary score from metrics
func computeSummaryScore(metrics *BenchmarkMetrics) int64 {
	if metrics == nil {
		return 0
	}

	// Weighted average of component scores
	// Using fixed-point arithmetic (scale: 1000000)
	const scale int64 = 1000000

	cpuScore := (metrics.CPUSingleCoreScore + metrics.CPUMultiCoreScore) / 2
	weights := map[string]int64{
		"cpu":     250000, // 25%
		"memory":  250000, // 25%
		"disk":    250000, // 25%
		"network": 250000, // 25%
	}

	weightedSum := (cpuScore * weights["cpu"]) +
		(metrics.MemoryScore * weights["memory"]) +
		(metrics.DiskScore * weights["disk"]) +
		(metrics.NetworkScore * weights["network"])

	return weightedSum / scale
}
