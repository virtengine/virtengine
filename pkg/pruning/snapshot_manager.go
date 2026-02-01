// Package pruning provides state pruning optimization for VirtEngine blockchain.
package pruning

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"cosmossdk.io/log"
)

// SnapshotManager handles snapshot lifecycle operations.
type SnapshotManager struct {
	config    SnapshotConfig
	logger    log.Logger
	mu        sync.RWMutex
	creating  atomic.Bool
	snapshots []SnapshotInfo
	baseDir   string
}

// SnapshotInfo contains metadata about a snapshot.
type SnapshotInfo struct {
	Height      uint64    `json:"height"`
	Format      uint32    `json:"format"`
	Chunks      uint32    `json:"chunks"`
	Hash        []byte    `json:"hash"`
	CreatedAt   time.Time `json:"created_at"`
	Size        int64     `json:"size"`
	Compressed  bool      `json:"compressed"`
	MetadataURL string    `json:"metadata_url,omitempty"`
}

// SnapshotCreateResult contains the result of snapshot creation.
type SnapshotCreateResult struct {
	Info     SnapshotInfo
	Duration time.Duration
	Success  bool
	Error    error
}

// NewSnapshotManager creates a new snapshot manager.
func NewSnapshotManager(config SnapshotConfig, baseDir string, logger log.Logger) (*SnapshotManager, error) {
	snapshotDir := config.Directory
	if snapshotDir == "" {
		snapshotDir = filepath.Join(baseDir, "data", "snapshots")
	}

	// Ensure directory exists
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	sm := &SnapshotManager{
		config:  config,
		logger:  logger.With("module", "snapshot-manager"),
		baseDir: snapshotDir,
	}

	// Load existing snapshots
	if err := sm.loadSnapshots(); err != nil {
		logger.Warn("failed to load existing snapshots", "error", err)
	}

	return sm, nil
}

// loadSnapshots loads snapshot metadata from disk.
func (sm *SnapshotManager) loadSnapshots() error {
	metadataFile := filepath.Join(sm.baseDir, "metadata.json")
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var snapshots []SnapshotInfo
	if err := json.Unmarshal(data, &snapshots); err != nil {
		return err
	}

	sm.mu.Lock()
	sm.snapshots = snapshots
	sm.mu.Unlock()

	return nil
}

// saveSnapshotsLocked saves snapshots to disk. Caller must hold the lock.
func (sm *SnapshotManager) saveSnapshotsLocked() error {
	data, err := json.MarshalIndent(sm.snapshots, "", "  ")
	if err != nil {
		return err
	}

	metadataFile := filepath.Join(sm.baseDir, "metadata.json")
	return os.WriteFile(metadataFile, data, 0644)
}

// ShouldCreateSnapshot determines if a snapshot should be created at this height.
func (sm *SnapshotManager) ShouldCreateSnapshot(height uint64) bool {
	if !sm.config.Enabled {
		return false
	}

	if sm.config.Interval == 0 {
		return false
	}

	return height%sm.config.Interval == 0
}

// StartCreating marks snapshot creation as in progress.
func (sm *SnapshotManager) StartCreating() error {
	if !sm.creating.CompareAndSwap(false, true) {
		return ErrSnapshotInProgress
	}
	return nil
}

// EndCreating marks snapshot creation as complete.
func (sm *SnapshotManager) EndCreating() {
	sm.creating.Store(false)
}

// IsCreating returns whether snapshot creation is in progress.
func (sm *SnapshotManager) IsCreating() bool {
	return sm.creating.Load()
}

// RegisterSnapshot registers a new snapshot.
func (sm *SnapshotManager) RegisterSnapshot(info SnapshotInfo) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check for duplicate
	for _, s := range sm.snapshots {
		if s.Height == info.Height {
			return fmt.Errorf("snapshot at height %d already exists", info.Height)
		}
	}

	sm.snapshots = append(sm.snapshots, info)

	// Sort by height descending
	sort.Slice(sm.snapshots, func(i, j int) bool {
		return sm.snapshots[i].Height > sm.snapshots[j].Height
	})

	// Cleanup old snapshots
	if err := sm.cleanupOldSnapshots(); err != nil {
		sm.logger.Warn("failed to cleanup old snapshots", "error", err)
	}

	return sm.saveSnapshotsLocked()
}

// cleanupOldSnapshots removes snapshots beyond the retention limit.
//
//nolint:unparam // result 0 (error) reserved for future cleanup failures
func (sm *SnapshotManager) cleanupOldSnapshots() error {
	keepRecent := int(sm.config.KeepRecent)
	if keepRecent <= 0 {
		keepRecent = int(DefaultSnapshotKeepRecent)
	}

	if len(sm.snapshots) <= keepRecent {
		return nil
	}

	// Remove excess snapshots
	toRemove := sm.snapshots[keepRecent:]
	sm.snapshots = sm.snapshots[:keepRecent]

	// Delete snapshot files
	for _, s := range toRemove {
		snapshotPath := sm.getSnapshotPath(s.Height)
		if err := os.RemoveAll(snapshotPath); err != nil {
			sm.logger.Warn("failed to delete snapshot", "height", s.Height, "error", err)
		} else {
			sm.logger.Info("deleted old snapshot", "height", s.Height)
		}
	}

	return nil
}

// getSnapshotPath returns the path to a snapshot directory.
func (sm *SnapshotManager) getSnapshotPath(height uint64) string {
	return filepath.Join(sm.baseDir, fmt.Sprintf("snapshot-%d", height))
}

// GetSnapshots returns all registered snapshots.
func (sm *SnapshotManager) GetSnapshots() []SnapshotInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]SnapshotInfo, len(sm.snapshots))
	copy(result, sm.snapshots)
	return result
}

// GetSnapshot returns a specific snapshot by height.
func (sm *SnapshotManager) GetSnapshot(height uint64) (SnapshotInfo, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, s := range sm.snapshots {
		if s.Height == height {
			return s, true
		}
	}
	return SnapshotInfo{}, false
}

// GetLatestSnapshot returns the most recent snapshot.
func (sm *SnapshotManager) GetLatestSnapshot() (SnapshotInfo, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if len(sm.snapshots) == 0 {
		return SnapshotInfo{}, false
	}

	return sm.snapshots[0], true
}

// DeleteSnapshot deletes a snapshot by height.
func (sm *SnapshotManager) DeleteSnapshot(height uint64) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	idx := -1
	for i, s := range sm.snapshots {
		if s.Height == height {
			idx = i
			break
		}
	}

	if idx == -1 {
		return ErrSnapshotNotFound
	}

	// Delete files
	snapshotPath := sm.getSnapshotPath(height)
	if err := os.RemoveAll(snapshotPath); err != nil {
		return fmt.Errorf("failed to delete snapshot files: %w", err)
	}

	// Remove from list
	sm.snapshots = append(sm.snapshots[:idx], sm.snapshots[idx+1:]...)

	return sm.saveSnapshotsLocked()
}

// GetTotalSize returns the total size of all snapshots.
func (sm *SnapshotManager) GetTotalSize() int64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var total int64
	for _, s := range sm.snapshots {
		total += s.Size
	}
	return total
}

// StateSyncInfo contains information for state sync.
type StateSyncInfo struct {
	Enabled           bool           `json:"enabled"`
	LatestSnapshot    *SnapshotInfo  `json:"latest_snapshot,omitempty"`
	AvailableHeights  []uint64       `json:"available_heights"`
	TrustHeight       int64          `json:"trust_height"`
	TrustHash         string         `json:"trust_hash"`
	SnapshotInterval  uint64         `json:"snapshot_interval"`
	SnapshotKeepCount uint32         `json:"snapshot_keep_count"`
}

// GetStateSyncInfo returns state sync compatibility information.
func (sm *SnapshotManager) GetStateSyncInfo() StateSyncInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	info := StateSyncInfo{
		Enabled:           sm.config.Enabled,
		SnapshotInterval:  sm.config.Interval,
		SnapshotKeepCount: sm.config.KeepRecent,
	}

	if len(sm.snapshots) > 0 {
		latest := sm.snapshots[0]
		info.LatestSnapshot = &latest
		info.TrustHeight = int64(latest.Height)
	}

	for _, s := range sm.snapshots {
		info.AvailableHeights = append(info.AvailableHeights, s.Height)
	}

	return info
}

// ValidateStateSyncCompatibility validates if the current configuration
// is compatible with state sync requirements.
func (sm *SnapshotManager) ValidateStateSyncCompatibility(minRetainBlocks uint64, keepRecent uint64) error {
	if !sm.config.Enabled {
		return fmt.Errorf("snapshots must be enabled for state sync")
	}

	if sm.config.Interval == 0 {
		return fmt.Errorf("snapshot interval must be greater than 0")
	}

	if sm.config.KeepRecent < 2 {
		return fmt.Errorf("at least 2 snapshots must be kept for state sync")
	}

	// Ensure we retain enough blocks for state sync
	if minRetainBlocks > 0 && minRetainBlocks < sm.config.Interval {
		return fmt.Errorf("min-retain-blocks (%d) must be >= snapshot-interval (%d)",
			minRetainBlocks, sm.config.Interval)
	}

	// Ensure keep-recent is compatible
	if keepRecent > 0 && keepRecent < sm.config.Interval*uint64(sm.config.KeepRecent) {
		return fmt.Errorf("keep-recent (%d) should be >= snapshot-interval * snapshot-keep-recent (%d)",
			keepRecent, sm.config.Interval*uint64(sm.config.KeepRecent))
	}

	return nil
}

// SnapshotScheduler handles automatic snapshot scheduling.
type SnapshotScheduler struct {
	manager    *SnapshotManager
	logger     log.Logger
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
}

// NewSnapshotScheduler creates a new snapshot scheduler.
func NewSnapshotScheduler(manager *SnapshotManager, logger log.Logger) *SnapshotScheduler {
	return &SnapshotScheduler{
		manager: manager,
		logger:  logger.With("module", "snapshot-scheduler"),
	}
}

// Stop stops the scheduler.
func (ss *SnapshotScheduler) Stop() {
	if ss.cancelFunc != nil {
		ss.cancelFunc()
	}
	ss.wg.Wait()
}
