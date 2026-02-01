// Package pruning provides state pruning optimization for VirtEngine blockchain.
//
// DATA-002: State Pruning Optimization
// This package implements optimized pruning strategies for storage efficiency
// while maintaining historical state query support and state sync compatibility.
package pruning

import (
	"errors"
	"time"
)

// Strategy defines the pruning strategy type.
type Strategy string

const (
	// StrategyDefault keeps the last 362880 states, pruning at 10 block intervals.
	// This is suitable for most validator nodes.
	StrategyDefault Strategy = "default"

	// StrategyNothing keeps all historic states (archiving node).
	// Use this for archive nodes that need full history.
	StrategyNothing Strategy = "nothing"

	// StrategyEverything keeps only 2 latest states, pruning at 10 block intervals.
	// Use this for light nodes or nodes that don't need history.
	StrategyEverything Strategy = "everything"

	// StrategyCustom allows manual specification of pruning options.
	StrategyCustom Strategy = "custom"

	// StrategyTiered implements tiered retention based on data age.
	// Recent data is kept at full resolution, older data is sampled.
	StrategyTiered Strategy = "tiered"
)

// Default configuration values.
const (
	// DefaultKeepRecent is the default number of recent heights to keep.
	DefaultKeepRecent uint64 = 362880

	// DefaultInterval is the default pruning interval in blocks.
	DefaultInterval uint64 = 10

	// DefaultMinRetainBlocks is the minimum blocks to retain for state sync.
	DefaultMinRetainBlocks uint64 = 0

	// DefaultSnapshotInterval is the default snapshot creation interval.
	DefaultSnapshotInterval uint64 = 1000

	// DefaultSnapshotKeepRecent is the default number of snapshots to keep.
	DefaultSnapshotKeepRecent uint32 = 2

	// DefaultDiskWarningThresholdPercent triggers a warning alert.
	DefaultDiskWarningThresholdPercent float64 = 80.0

	// DefaultDiskCriticalThresholdPercent triggers a critical alert.
	DefaultDiskCriticalThresholdPercent float64 = 90.0

	// DefaultGrowthProjectionDays is the number of days for growth projection.
	DefaultGrowthProjectionDays int = 30
)

// Tiered retention tier configuration.
const (
	// Tier1RetentionBlocks - full resolution for recent blocks.
	Tier1RetentionBlocks uint64 = 100000 // ~1 week at 6s blocks

	// Tier2RetentionBlocks - sampled every 100 blocks.
	Tier2RetentionBlocks uint64 = 500000 // ~1 month

	// Tier3SamplingRate - beyond tier 2, sample every N blocks.
	Tier3SamplingRate uint64 = 1000
)

// Errors returned by pruning operations.
var (
	ErrInvalidStrategy       = errors.New("invalid pruning strategy")
	ErrInvalidKeepRecent     = errors.New("keep-recent must be greater than 0")
	ErrInvalidInterval       = errors.New("pruning interval must be greater than 0")
	ErrSnapshotInProgress    = errors.New("snapshot operation already in progress")
	ErrSnapshotNotFound      = errors.New("snapshot not found")
	ErrPruningInProgress     = errors.New("pruning operation already in progress")
	ErrStateSyncIncompatible = errors.New("pruning configuration incompatible with state sync")
	ErrDiskSpaceCritical     = errors.New("disk space critically low")
)

// Config holds the complete pruning configuration.
type Config struct {
	// Strategy is the pruning strategy to use.
	Strategy Strategy `mapstructure:"strategy" json:"strategy"`

	// KeepRecent is the number of recent heights to keep on disk.
	KeepRecent uint64 `mapstructure:"keep_recent" json:"keep_recent"`

	// Interval is the height interval at which pruned heights are removed.
	Interval uint64 `mapstructure:"interval" json:"interval"`

	// MinRetainBlocks is the minimum block height offset for CometBFT block pruning.
	MinRetainBlocks uint64 `mapstructure:"min_retain_blocks" json:"min_retain_blocks"`

	// SnapshotConfig contains snapshot-related configuration.
	Snapshot SnapshotConfig `mapstructure:"snapshot" json:"snapshot"`

	// TieredConfig contains tiered retention configuration (for tiered strategy).
	Tiered TieredConfig `mapstructure:"tiered" json:"tiered"`

	// DiskMonitor contains disk monitoring configuration.
	DiskMonitor DiskMonitorConfig `mapstructure:"disk_monitor" json:"disk_monitor"`

	// HistoricalQueries enables support for historical state queries.
	HistoricalQueries HistoricalQueryConfig `mapstructure:"historical_queries" json:"historical_queries"`
}

// SnapshotConfig holds snapshot-related configuration.
type SnapshotConfig struct {
	// Enabled indicates whether snapshots are enabled.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Interval is the block interval for snapshot creation.
	Interval uint64 `mapstructure:"interval" json:"interval"`

	// KeepRecent is the number of recent snapshots to keep.
	KeepRecent uint32 `mapstructure:"keep_recent" json:"keep_recent"`

	// Directory is the directory for storing snapshots.
	Directory string `mapstructure:"directory" json:"directory"`

	// Compression enables compression for snapshots.
	Compression bool `mapstructure:"compression" json:"compression"`

	// CompressionLevel is the compression level (1-9, where 9 is highest).
	CompressionLevel int `mapstructure:"compression_level" json:"compression_level"`
}

// TieredConfig holds tiered retention configuration.
type TieredConfig struct {
	// Enabled indicates whether tiered retention is active.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Tier1Blocks is the number of blocks to keep at full resolution.
	Tier1Blocks uint64 `mapstructure:"tier1_blocks" json:"tier1_blocks"`

	// Tier2Blocks is the number of blocks for tier 2 (sampled).
	Tier2Blocks uint64 `mapstructure:"tier2_blocks" json:"tier2_blocks"`

	// Tier2SamplingRate is the sampling rate for tier 2 (keep every N blocks).
	Tier2SamplingRate uint64 `mapstructure:"tier2_sampling_rate" json:"tier2_sampling_rate"`

	// Tier3SamplingRate is the sampling rate for tier 3 (archive tier).
	Tier3SamplingRate uint64 `mapstructure:"tier3_sampling_rate" json:"tier3_sampling_rate"`
}

// DiskMonitorConfig holds disk monitoring configuration.
type DiskMonitorConfig struct {
	// Enabled indicates whether disk monitoring is active.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// CheckInterval is the interval between disk usage checks.
	CheckInterval time.Duration `mapstructure:"check_interval" json:"check_interval"`

	// WarningThresholdPercent is the disk usage percentage that triggers a warning.
	WarningThresholdPercent float64 `mapstructure:"warning_threshold_percent" json:"warning_threshold_percent"`

	// CriticalThresholdPercent is the disk usage percentage that triggers a critical alert.
	CriticalThresholdPercent float64 `mapstructure:"critical_threshold_percent" json:"critical_threshold_percent"`

	// AutoPruneOnCritical enables automatic aggressive pruning when critical threshold is reached.
	AutoPruneOnCritical bool `mapstructure:"auto_prune_on_critical" json:"auto_prune_on_critical"`

	// GrowthProjectionDays is the number of days for growth rate calculation.
	GrowthProjectionDays int `mapstructure:"growth_projection_days" json:"growth_projection_days"`
}

// HistoricalQueryConfig holds configuration for historical state queries.
type HistoricalQueryConfig struct {
	// Enabled indicates whether historical queries are supported.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// MinQueryableHeight is the minimum height that can be queried.
	// Set to 0 to allow queries back to the oldest retained state.
	MinQueryableHeight int64 `mapstructure:"min_queryable_height" json:"min_queryable_height"`

	// CacheSize is the LRU cache size for historical query results.
	CacheSize int `mapstructure:"cache_size" json:"cache_size"`
}

// DefaultConfig returns the default pruning configuration.
func DefaultConfig() Config {
	return Config{
		Strategy:        StrategyDefault,
		KeepRecent:      DefaultKeepRecent,
		Interval:        DefaultInterval,
		MinRetainBlocks: DefaultMinRetainBlocks,
		Snapshot: SnapshotConfig{
			Enabled:          true,
			Interval:         DefaultSnapshotInterval,
			KeepRecent:       DefaultSnapshotKeepRecent,
			Directory:        "data/snapshots",
			Compression:      true,
			CompressionLevel: 6,
		},
		Tiered: TieredConfig{
			Enabled:           false,
			Tier1Blocks:       Tier1RetentionBlocks,
			Tier2Blocks:       Tier2RetentionBlocks,
			Tier2SamplingRate: 100,
			Tier3SamplingRate: Tier3SamplingRate,
		},
		DiskMonitor: DiskMonitorConfig{
			Enabled:                  true,
			CheckInterval:            5 * time.Minute,
			WarningThresholdPercent:  DefaultDiskWarningThresholdPercent,
			CriticalThresholdPercent: DefaultDiskCriticalThresholdPercent,
			AutoPruneOnCritical:      false,
			GrowthProjectionDays:     DefaultGrowthProjectionDays,
		},
		HistoricalQueries: HistoricalQueryConfig{
			Enabled:            true,
			MinQueryableHeight: 0,
			CacheSize:          1000,
		},
	}
}

// NothingConfig returns a configuration for archiving nodes.
func NothingConfig() Config {
	cfg := DefaultConfig()
	cfg.Strategy = StrategyNothing
	cfg.KeepRecent = 0
	cfg.Interval = 0
	cfg.DiskMonitor.WarningThresholdPercent = 70.0
	cfg.DiskMonitor.CriticalThresholdPercent = 85.0
	return cfg
}

// EverythingConfig returns a configuration for minimal storage.
func EverythingConfig() Config {
	cfg := DefaultConfig()
	cfg.Strategy = StrategyEverything
	cfg.KeepRecent = 2
	cfg.Interval = 10
	cfg.HistoricalQueries.Enabled = false
	return cfg
}

// TieredConfig returns a configuration for tiered retention.
func TieredRetentionConfig() Config {
	cfg := DefaultConfig()
	cfg.Strategy = StrategyTiered
	cfg.Tiered.Enabled = true
	return cfg
}

// Validate validates the pruning configuration.
func (c *Config) Validate() error {
	// Validate strategy
	switch c.Strategy {
	case StrategyDefault, StrategyNothing, StrategyEverything, StrategyCustom, StrategyTiered:
		// valid
	default:
		return ErrInvalidStrategy
	}

	// Validate keep-recent for non-nothing strategies
	if c.Strategy != StrategyNothing && c.KeepRecent == 0 {
		return ErrInvalidKeepRecent
	}

	// Validate interval for non-nothing strategies
	if c.Strategy != StrategyNothing && c.Interval == 0 {
		return ErrInvalidInterval
	}

	// Validate snapshot config
	if c.Snapshot.Enabled {
		if c.Snapshot.Interval == 0 {
			c.Snapshot.Interval = DefaultSnapshotInterval
		}
		if c.Snapshot.KeepRecent == 0 {
			c.Snapshot.KeepRecent = DefaultSnapshotKeepRecent
		}
	}

	// Validate tiered config
	if c.Strategy == StrategyTiered && c.Tiered.Enabled {
		if c.Tiered.Tier1Blocks == 0 {
			c.Tiered.Tier1Blocks = Tier1RetentionBlocks
		}
		if c.Tiered.Tier2SamplingRate == 0 {
			c.Tiered.Tier2SamplingRate = 100
		}
		if c.Tiered.Tier3SamplingRate == 0 {
			c.Tiered.Tier3SamplingRate = Tier3SamplingRate
		}
	}

	// Validate disk monitor config
	if c.DiskMonitor.Enabled {
		if c.DiskMonitor.WarningThresholdPercent <= 0 || c.DiskMonitor.WarningThresholdPercent > 100 {
			c.DiskMonitor.WarningThresholdPercent = DefaultDiskWarningThresholdPercent
		}
		if c.DiskMonitor.CriticalThresholdPercent <= 0 || c.DiskMonitor.CriticalThresholdPercent > 100 {
			c.DiskMonitor.CriticalThresholdPercent = DefaultDiskCriticalThresholdPercent
		}
		if c.DiskMonitor.CriticalThresholdPercent <= c.DiskMonitor.WarningThresholdPercent {
			c.DiskMonitor.CriticalThresholdPercent = c.DiskMonitor.WarningThresholdPercent + 10
		}
	}

	return nil
}

// IsStateSyncCompatible checks if the configuration is compatible with state sync.
func (c *Config) IsStateSyncCompatible() bool {
	// State sync requires snapshots
	if !c.Snapshot.Enabled {
		return false
	}

	// State sync requires sufficient state retention
	if c.Strategy == StrategyEverything {
		return false
	}

	// Minimum retain blocks should be set for state sync
	if c.MinRetainBlocks > 0 && c.MinRetainBlocks < c.Snapshot.Interval {
		return false
	}

	return true
}

// PruningOptions returns the SDK-compatible pruning options.
func (c *Config) PruningOptions() (keepRecent uint64, interval uint64) {
	switch c.Strategy {
	case StrategyNothing:
		return 0, 0
	case StrategyEverything:
		return 2, 10
	case StrategyDefault:
		return DefaultKeepRecent, DefaultInterval
	case StrategyCustom, StrategyTiered:
		return c.KeepRecent, c.Interval
	default:
		return DefaultKeepRecent, DefaultInterval
	}
}

