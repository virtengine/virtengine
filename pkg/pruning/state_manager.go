// Package pruning provides state pruning optimization for VirtEngine blockchain.
package pruning

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"cosmossdk.io/log"
)

// StateManager handles state pruning operations.
type StateManager struct {
	config     Config
	logger     log.Logger
	mu         sync.RWMutex
	pruning    atomic.Bool
	lastPruned int64
	metrics    *Metrics
	hooks      []PruningHook
}

// PruningHook allows external components to be notified of pruning events.
type PruningHook interface {
	// OnPruneStart is called when pruning begins.
	OnPruneStart(ctx context.Context, fromHeight, toHeight int64)

	// OnPruneComplete is called when pruning completes.
	OnPruneComplete(ctx context.Context, fromHeight, toHeight int64, prunedCount int64, duration time.Duration)

	// OnPruneError is called when pruning fails.
	OnPruneError(ctx context.Context, err error)
}

// HeightInfo contains information about state at a specific height.
type HeightInfo struct {
	Height      int64
	Timestamp   time.Time
	StateSize   int64
	IsPruned    bool
	IsSnapshot  bool
	IsSampled   bool
	RetainUntil time.Time
}

// PruneResult contains the result of a pruning operation.
type PruneResult struct {
	FromHeight   int64
	ToHeight     int64
	PrunedCount  int64
	BytesFreed   int64
	Duration     time.Duration
	Errors       []error
	SkippedCount int64
}

// NewStateManager creates a new state manager.
func NewStateManager(config Config, logger log.Logger) (*StateManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &StateManager{
		config:  config,
		logger:  logger.With("module", "pruning"),
		metrics: NewMetrics(),
	}, nil
}

// Config returns the current configuration.
func (sm *StateManager) Config() Config {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.config
}

// UpdateConfig updates the pruning configuration.
func (sm *StateManager) UpdateConfig(config Config) error {
	if err := config.Validate(); err != nil {
		return err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.config = config
	return nil
}

// RegisterHook registers a pruning hook.
func (sm *StateManager) RegisterHook(hook PruningHook) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.hooks = append(sm.hooks, hook)
}

// ShouldPruneHeight determines if a height should be pruned based on current config.
func (sm *StateManager) ShouldPruneHeight(currentHeight, targetHeight int64) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Never prune if strategy is nothing
	if sm.config.Strategy == StrategyNothing {
		return false
	}

	// Calculate the cutoff height
	keepRecent, _ := sm.config.PruningOptions()
	cutoffHeight := currentHeight - int64(keepRecent)

	// Don't prune if within keep-recent window
	if targetHeight > cutoffHeight {
		return false
	}

	// For tiered strategy, check if this height should be sampled
	if sm.config.Strategy == StrategyTiered && sm.config.Tiered.Enabled {
		return sm.shouldPruneTiered(currentHeight, targetHeight)
	}

	return true
}

// shouldPruneTiered determines if a height should be pruned under tiered strategy.
func (sm *StateManager) shouldPruneTiered(currentHeight, targetHeight int64) bool {
	age := currentHeight - targetHeight

	// Tier 1: Keep all (full resolution)
	if age <= int64(sm.config.Tiered.Tier1Blocks) {
		return false
	}

	// Tier 2: Sample every N blocks
	if age <= int64(sm.config.Tiered.Tier2Blocks) {
		return targetHeight%int64(sm.config.Tiered.Tier2SamplingRate) != 0
	}

	// Tier 3: Sample at lower rate
	return targetHeight%int64(sm.config.Tiered.Tier3SamplingRate) != 0
}

// ShouldRetainForSnapshot determines if a height should be retained for snapshots.
func (sm *StateManager) ShouldRetainForSnapshot(height int64) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if !sm.config.Snapshot.Enabled {
		return false
	}

	// Retain heights that are snapshot boundaries
	return height%int64(sm.config.Snapshot.Interval) == 0
}

// CalculatePruneRange calculates the range of heights to prune.
func (sm *StateManager) CalculatePruneRange(currentHeight, lastPruned int64) (fromHeight, toHeight int64, shouldPrune bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Check pruning interval
	_, interval := sm.config.PruningOptions()
	if interval == 0 {
		return 0, 0, false
	}

	if currentHeight%int64(interval) != 0 {
		return 0, 0, false
	}

	keepRecent, _ := sm.config.PruningOptions()
	toHeight = currentHeight - int64(keepRecent)

	// Start from last pruned height + 1
	fromHeight = lastPruned + 1
	if fromHeight < 1 {
		fromHeight = 1
	}

	// Nothing to prune
	if fromHeight >= toHeight {
		return 0, 0, false
	}

	return fromHeight, toHeight, true
}

// StartPruning marks pruning as in progress.
func (sm *StateManager) StartPruning() error {
	if !sm.pruning.CompareAndSwap(false, true) {
		return ErrPruningInProgress
	}
	return nil
}

// EndPruning marks pruning as complete.
func (sm *StateManager) EndPruning(toHeight int64) {
	sm.lastPruned = toHeight
	sm.pruning.Store(false)
}

// IsPruning returns whether pruning is in progress.
func (sm *StateManager) IsPruning() bool {
	return sm.pruning.Load()
}

// LastPrunedHeight returns the last pruned height.
func (sm *StateManager) LastPrunedHeight() int64 {
	return sm.lastPruned
}

// GetHeightStatus returns the status of a specific height.
func (sm *StateManager) GetHeightStatus(currentHeight, targetHeight int64) HeightInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	info := HeightInfo{
		Height: targetHeight,
	}

	keepRecent, _ := sm.config.PruningOptions()
	cutoffHeight := currentHeight - int64(keepRecent)

	// Determine if pruned
	if targetHeight <= cutoffHeight && sm.config.Strategy != StrategyNothing {
		// Check if it's a sampled height
		if sm.config.Strategy == StrategyTiered && sm.config.Tiered.Enabled {
			age := currentHeight - targetHeight
			if age > int64(sm.config.Tiered.Tier2Blocks) {
				info.IsSampled = targetHeight%int64(sm.config.Tiered.Tier3SamplingRate) == 0
			} else if age > int64(sm.config.Tiered.Tier1Blocks) {
				info.IsSampled = targetHeight%int64(sm.config.Tiered.Tier2SamplingRate) == 0
			}
			info.IsPruned = !info.IsSampled
		} else {
			info.IsPruned = true
		}
	}

	// Check if snapshot
	if sm.config.Snapshot.Enabled && targetHeight%int64(sm.config.Snapshot.Interval) == 0 {
		info.IsSnapshot = true
	}

	return info
}

// GetRetentionPolicy returns a human-readable description of the retention policy.
func (sm *StateManager) GetRetentionPolicy() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	switch sm.config.Strategy {
	case StrategyNothing:
		return "Archive mode: all states are retained"
	case StrategyEverything:
		return "Minimal mode: only 2 most recent states are kept"
	case StrategyDefault:
		return fmt.Sprintf("Default mode: %d recent states kept, pruning every %d blocks",
			DefaultKeepRecent, DefaultInterval)
	case StrategyTiered:
		return fmt.Sprintf("Tiered mode: full resolution for %d blocks, sampled beyond",
			sm.config.Tiered.Tier1Blocks)
	case StrategyCustom:
		return fmt.Sprintf("Custom mode: %d recent states kept, pruning every %d blocks",
			sm.config.KeepRecent, sm.config.Interval)
	default:
		return "Unknown strategy"
	}
}

// EstimateStorageSavings estimates storage savings for the current configuration.
func (sm *StateManager) EstimateStorageSavings(totalBlocks int64, avgBlockSize int64) StorageSavingsEstimate {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	estimate := StorageSavingsEstimate{
		TotalBlocks:   totalBlocks,
		AvgBlockSize:  avgBlockSize,
		TotalSize:     totalBlocks * avgBlockSize,
		RetainedSize:  0,
		SavingsAmount: 0,
	}

	keepRecent, _ := sm.config.PruningOptions()

	switch sm.config.Strategy {
	case StrategyNothing:
		estimate.RetainedSize = estimate.TotalSize
	case StrategyEverything:
		estimate.RetainedSize = 2 * avgBlockSize
	case StrategyDefault, StrategyCustom:
		retained := int64(keepRecent)
		if retained > totalBlocks {
			retained = totalBlocks
		}
		estimate.RetainedSize = retained * avgBlockSize
	case StrategyTiered:
		estimate.RetainedSize = sm.estimateTieredRetention(totalBlocks, avgBlockSize)
	}

	estimate.SavingsAmount = estimate.TotalSize - estimate.RetainedSize
	if estimate.TotalSize > 0 {
		estimate.SavingsPercent = float64(estimate.SavingsAmount) / float64(estimate.TotalSize) * 100
	}

	return estimate
}

// estimateTieredRetention estimates retained storage for tiered strategy.
func (sm *StateManager) estimateTieredRetention(totalBlocks, avgBlockSize int64) int64 {
	var retained int64

	// Tier 1: Full resolution
	tier1 := int64(sm.config.Tiered.Tier1Blocks)
	if tier1 > totalBlocks {
		tier1 = totalBlocks
	}
	retained += tier1 * avgBlockSize

	remaining := totalBlocks - tier1
	if remaining <= 0 {
		return retained
	}

	// Tier 2: Sampled
	tier2 := int64(sm.config.Tiered.Tier2Blocks) - tier1
	if tier2 > remaining {
		tier2 = remaining
	}
	tier2Sampled := tier2 / int64(sm.config.Tiered.Tier2SamplingRate)
	retained += tier2Sampled * avgBlockSize

	remaining -= tier2
	if remaining <= 0 {
		return retained
	}

	// Tier 3: Lower sample rate
	tier3Sampled := remaining / int64(sm.config.Tiered.Tier3SamplingRate)
	retained += tier3Sampled * avgBlockSize

	return retained
}

// StorageSavingsEstimate contains storage savings estimates.
type StorageSavingsEstimate struct {
	TotalBlocks    int64
	AvgBlockSize   int64
	TotalSize      int64
	RetainedSize   int64
	SavingsAmount  int64
	SavingsPercent float64
}

// notifyHooks notifies all registered hooks of a pruning event.
func (sm *StateManager) notifyHooks(event string, ctx context.Context, fromHeight, toHeight int64, prunedCount int64, duration time.Duration, err error) {
	sm.mu.RLock()
	hooks := sm.hooks
	sm.mu.RUnlock()

	for _, hook := range hooks {
		switch event {
		case "start":
			hook.OnPruneStart(ctx, fromHeight, toHeight)
		case "complete":
			hook.OnPruneComplete(ctx, fromHeight, toHeight, prunedCount, duration)
		case "error":
			hook.OnPruneError(ctx, err)
		}
	}
}

// Metrics returns the pruning metrics.
func (sm *StateManager) Metrics() *Metrics {
	return sm.metrics
}
