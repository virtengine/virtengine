// Package pruning provides state pruning optimization for VirtEngine blockchain.
package pruning

import (
	"sync"
	"time"
)

// MetricsSnapshot is a copy-safe snapshot of metrics without the mutex.
// Use this for returning metrics from GetMetrics() to avoid copying sync.RWMutex.
type MetricsSnapshot struct {
	// Pruning metrics
	TotalPruningOperations int64         `json:"total_pruning_operations"`
	TotalHeightsPruned     int64         `json:"total_heights_pruned"`
	TotalBytesPruned       int64         `json:"total_bytes_pruned"`
	LastPruningDuration    time.Duration `json:"last_pruning_duration"`
	LastPruningTime        time.Time     `json:"last_pruning_time"`
	AveragePruningDuration time.Duration `json:"average_pruning_duration"`
	PruningErrorCount      int64         `json:"pruning_error_count"`

	// Snapshot metrics
	TotalSnapshotsCreated   int64         `json:"total_snapshots_created"`
	TotalSnapshotsDeleted   int64         `json:"total_snapshots_deleted"`
	LastSnapshotDuration    time.Duration `json:"last_snapshot_duration"`
	LastSnapshotTime        time.Time     `json:"last_snapshot_time"`
	LastSnapshotSize        int64         `json:"last_snapshot_size"`
	AverageSnapshotDuration time.Duration `json:"average_snapshot_duration"`
	SnapshotErrorCount      int64         `json:"snapshot_error_count"`

	// Disk metrics
	CurrentDiskUsage        uint64  `json:"current_disk_usage"`
	CurrentDiskUsagePercent float64 `json:"current_disk_usage_percent"`
	DailyGrowthRate         int64   `json:"daily_growth_rate"`
	DaysUntilFull           int     `json:"days_until_full"`

	// State metrics
	CurrentHeight        int64 `json:"current_height"`
	OldestRetainedHeight int64 `json:"oldest_retained_height"`
	RetainedHeightsCount int64 `json:"retained_heights_count"`
	SnapshotsCount       int   `json:"snapshots_count"`
}

// Metrics tracks pruning-related metrics for monitoring and telemetry.
type Metrics struct {
	mu sync.RWMutex

	// Pruning metrics
	TotalPruningOperations int64
	TotalHeightsPruned     int64
	TotalBytesPruned       int64
	LastPruningDuration    time.Duration
	LastPruningTime        time.Time
	AveragePruningDuration time.Duration
	PruningErrorCount      int64

	// Snapshot metrics
	TotalSnapshotsCreated   int64
	TotalSnapshotsDeleted   int64
	LastSnapshotDuration    time.Duration
	LastSnapshotTime        time.Time
	LastSnapshotSize        int64
	AverageSnapshotDuration time.Duration
	SnapshotErrorCount      int64

	// Disk metrics
	CurrentDiskUsage        uint64
	CurrentDiskUsagePercent float64
	DailyGrowthRate         int64
	DaysUntilFull           int

	// State metrics
	CurrentHeight        int64
	OldestRetainedHeight int64
	RetainedHeightsCount int64
	SnapshotsCount       int
}

// NewMetrics creates a new Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{}
}

// RecordPruningOperation records a pruning operation.
func (m *Metrics) RecordPruningOperation(heightsPruned int64, bytesPruned int64, duration time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalPruningOperations++
	m.TotalHeightsPruned += heightsPruned
	m.TotalBytesPruned += bytesPruned
	m.LastPruningDuration = duration
	m.LastPruningTime = time.Now()

	if err != nil {
		m.PruningErrorCount++
	}

	// Update average (simple moving average)
	if m.TotalPruningOperations == 1 {
		m.AveragePruningDuration = duration
	} else {
		m.AveragePruningDuration = time.Duration(
			(int64(m.AveragePruningDuration)*(m.TotalPruningOperations-1) + int64(duration)) / m.TotalPruningOperations,
		)
	}
}

// RecordSnapshotOperation records a snapshot operation.
func (m *Metrics) RecordSnapshotOperation(created bool, size int64, duration time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if created {
		m.TotalSnapshotsCreated++
		m.LastSnapshotSize = size
	} else {
		m.TotalSnapshotsDeleted++
	}

	m.LastSnapshotDuration = duration
	m.LastSnapshotTime = time.Now()

	if err != nil {
		m.SnapshotErrorCount++
	}

	// Update average
	if m.TotalSnapshotsCreated == 1 {
		m.AverageSnapshotDuration = duration
	} else {
		m.AverageSnapshotDuration = time.Duration(
			(int64(m.AverageSnapshotDuration)*(m.TotalSnapshotsCreated-1) + int64(duration)) / m.TotalSnapshotsCreated,
		)
	}
}

// UpdateDiskMetrics updates disk-related metrics.
func (m *Metrics) UpdateDiskMetrics(usage uint64, usagePercent float64, growthRate int64, daysUntilFull int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CurrentDiskUsage = usage
	m.CurrentDiskUsagePercent = usagePercent
	m.DailyGrowthRate = growthRate
	m.DaysUntilFull = daysUntilFull
}

// UpdateStateMetrics updates state-related metrics.
func (m *Metrics) UpdateStateMetrics(currentHeight, oldestHeight, retainedCount int64, snapshotCount int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CurrentHeight = currentHeight
	m.OldestRetainedHeight = oldestHeight
	m.RetainedHeightsCount = retainedCount
	m.SnapshotsCount = snapshotCount
}

// GetMetrics returns a copy of all metrics as a snapshot (safe to copy).
func (m *Metrics) GetMetrics() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return MetricsSnapshot{
		TotalPruningOperations:  m.TotalPruningOperations,
		TotalHeightsPruned:      m.TotalHeightsPruned,
		TotalBytesPruned:        m.TotalBytesPruned,
		LastPruningDuration:     m.LastPruningDuration,
		LastPruningTime:         m.LastPruningTime,
		AveragePruningDuration:  m.AveragePruningDuration,
		PruningErrorCount:       m.PruningErrorCount,
		TotalSnapshotsCreated:   m.TotalSnapshotsCreated,
		TotalSnapshotsDeleted:   m.TotalSnapshotsDeleted,
		LastSnapshotDuration:    m.LastSnapshotDuration,
		LastSnapshotTime:        m.LastSnapshotTime,
		LastSnapshotSize:        m.LastSnapshotSize,
		AverageSnapshotDuration: m.AverageSnapshotDuration,
		SnapshotErrorCount:      m.SnapshotErrorCount,
		CurrentDiskUsage:        m.CurrentDiskUsage,
		CurrentDiskUsagePercent: m.CurrentDiskUsagePercent,
		DailyGrowthRate:         m.DailyGrowthRate,
		DaysUntilFull:           m.DaysUntilFull,
		CurrentHeight:           m.CurrentHeight,
		OldestRetainedHeight:    m.OldestRetainedHeight,
		RetainedHeightsCount:    m.RetainedHeightsCount,
		SnapshotsCount:          m.SnapshotsCount,
	}
}

// Reset resets all metrics.
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Reset all fields except the mutex
	m.TotalPruningOperations = 0
	m.TotalHeightsPruned = 0
	m.TotalBytesPruned = 0
	m.LastPruningDuration = 0
	m.LastPruningTime = time.Time{}
	m.AveragePruningDuration = 0
	m.PruningErrorCount = 0
	m.TotalSnapshotsCreated = 0
	m.TotalSnapshotsDeleted = 0
	m.LastSnapshotDuration = 0
	m.LastSnapshotTime = time.Time{}
	m.LastSnapshotSize = 0
	m.AverageSnapshotDuration = 0
	m.SnapshotErrorCount = 0
	m.CurrentDiskUsage = 0
	m.CurrentDiskUsagePercent = 0
	m.DailyGrowthRate = 0
	m.DaysUntilFull = 0
	m.CurrentHeight = 0
	m.OldestRetainedHeight = 0
	m.RetainedHeightsCount = 0
	m.SnapshotsCount = 0
}

// BenchmarkResult contains results from a pruning benchmark.
type BenchmarkResult struct {
	Strategy         Strategy      `json:"strategy"`
	TotalBlocks      int64         `json:"total_blocks"`
	BlocksToProcess  int64         `json:"blocks_to_process"`
	BlocksProcessed  int64         `json:"blocks_processed"`
	BlocksPruned     int64         `json:"blocks_pruned"`
	BlocksRetained   int64         `json:"blocks_retained"`
	Duration         time.Duration `json:"duration"`
	BlocksPerSecond  float64       `json:"blocks_per_second"`
	AvgTimePerBlock  time.Duration `json:"avg_time_per_block"`
	MemoryUsedMB     float64       `json:"memory_used_mb"`
	EstimatedSavings int64         `json:"estimated_savings_bytes"`
	SavingsPercent   float64       `json:"savings_percent"`
}

// Benchmark runs a pruning benchmark.
type Benchmark struct {
	config Config
	logger interface{ Info(string, ...interface{}) } //nolint:unused // Reserved for future benchmark logging
}

// NewBenchmark creates a new benchmark runner.
func NewBenchmark(config Config) *Benchmark {
	return &Benchmark{
		config: config,
	}
}

// RunSimulation runs a simulation to estimate pruning performance.
func (b *Benchmark) RunSimulation(totalBlocks, avgBlockSize int64) BenchmarkResult {
	start := time.Now()
	result := BenchmarkResult{
		Strategy:    b.config.Strategy,
		TotalBlocks: totalBlocks,
	}

	keepRecent, _ := b.config.PruningOptions()

	// Calculate blocks to prune
	if b.config.Strategy == StrategyNothing {
		result.BlocksRetained = totalBlocks
		result.BlocksPruned = 0
	} else if b.config.Strategy == StrategyEverything {
		result.BlocksRetained = 2
		result.BlocksPruned = totalBlocks - 2
	} else if b.config.Strategy == StrategyTiered && b.config.Tiered.Enabled {
		result = b.simulateTieredPruning(totalBlocks, avgBlockSize)
	} else {
		//nolint:gosec // G115: keepRecent is a configuration value bounded by practical limits
		retained := int64(keepRecent)
		if retained > totalBlocks {
			retained = totalBlocks
		}
		result.BlocksRetained = retained
		result.BlocksPruned = totalBlocks - retained
	}

	result.BlocksToProcess = result.BlocksPruned
	result.BlocksProcessed = result.BlocksPruned
	result.Duration = time.Since(start)

	if result.Duration.Seconds() > 0 {
		result.BlocksPerSecond = float64(result.BlocksProcessed) / result.Duration.Seconds()
	}
	if result.BlocksProcessed > 0 {
		result.AvgTimePerBlock = result.Duration / time.Duration(result.BlocksProcessed)
	}

	result.EstimatedSavings = result.BlocksPruned * avgBlockSize
	if totalBlocks > 0 {
		result.SavingsPercent = float64(result.BlocksPruned) / float64(totalBlocks) * 100
	}

	return result
}

// simulateTieredPruning simulates tiered pruning strategy.
func (b *Benchmark) simulateTieredPruning(totalBlocks, avgBlockSize int64) BenchmarkResult {
	result := BenchmarkResult{
		Strategy:    StrategyTiered,
		TotalBlocks: totalBlocks,
	}

	// Tier 1: Full retention
	//nolint:gosec // G115: Tier1Blocks is a configuration value bounded by practical limits
	tier1 := int64(b.config.Tiered.Tier1Blocks)
	if tier1 > totalBlocks {
		tier1 = totalBlocks
	}
	result.BlocksRetained += tier1

	remaining := totalBlocks - tier1
	if remaining <= 0 {
		return result
	}

	// Tier 2: Sample every N blocks
	//nolint:gosec // G115: Tier2Blocks is a configuration value bounded by practical limits
	tier2 := int64(b.config.Tiered.Tier2Blocks) - tier1
	if tier2 > remaining {
		tier2 = remaining
	}
	tier2Retained := tier2 / int64(b.config.Tiered.Tier2SamplingRate)
	result.BlocksRetained += tier2Retained

	remaining -= tier2
	if remaining <= 0 {
		result.BlocksPruned = totalBlocks - result.BlocksRetained
		return result
	}

	// Tier 3: Lower sample rate
	tier3Retained := remaining / int64(b.config.Tiered.Tier3SamplingRate)
	result.BlocksRetained += tier3Retained

	result.BlocksPruned = totalBlocks - result.BlocksRetained
	result.EstimatedSavings = result.BlocksPruned * avgBlockSize
	if totalBlocks > 0 {
		result.SavingsPercent = float64(result.BlocksPruned) / float64(totalBlocks) * 100
	}

	return result
}

// CompareBenchmarks compares different pruning strategies.
func CompareBenchmarks(totalBlocks, avgBlockSize int64) map[Strategy]BenchmarkResult {
	results := make(map[Strategy]BenchmarkResult)

	strategies := []Config{
		DefaultConfig(),
		NothingConfig(),
		EverythingConfig(),
		TieredRetentionConfig(),
	}

	for _, cfg := range strategies {
		bench := NewBenchmark(cfg)
		results[cfg.Strategy] = bench.RunSimulation(totalBlocks, avgBlockSize)
	}

	return results
}

// TelemetryLabels returns labels for telemetry/metrics systems.
func (m *Metrics) TelemetryLabels() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]string{
		"current_height":       formatInt64(m.CurrentHeight),
		"oldest_height":        formatInt64(m.OldestRetainedHeight),
		"retained_count":       formatInt64(m.RetainedHeightsCount),
		"snapshots":            formatInt(m.SnapshotsCount),
		"disk_usage_percent":   formatFloat64(m.CurrentDiskUsagePercent),
		"daily_growth_rate":    formatInt64(m.DailyGrowthRate),
		"days_until_full":      formatInt(m.DaysUntilFull),
		"total_pruned_heights": formatInt64(m.TotalHeightsPruned),
		"pruning_errors":       formatInt64(m.PruningErrorCount),
		"snapshot_errors":      formatInt64(m.SnapshotErrorCount),
	}
}

func formatInt64(v int64) string {
	return string(rune(v))
}

func formatInt(v int) string {
	return string(rune(v))
}

func formatFloat64(v float64) string {
	return string(rune(int(v)))
}
