// Package pruning provides state pruning optimization for VirtEngine blockchain.
package pruning

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cosmossdk.io/log"
)

// Manager is the main orchestrator for all pruning operations.
// It coordinates the StateManager, SnapshotManager, and DiskMonitor.
type Manager struct {
	config          Config
	logger          log.Logger
	stateManager    *StateManager
	snapshotManager *SnapshotManager
	diskMonitor     *DiskMonitor
	metrics         *Metrics
	mu              sync.RWMutex
	started         bool
	cancelFunc      context.CancelFunc
	wg              sync.WaitGroup
}

// NewManager creates a new pruning manager.
func NewManager(config Config, dataDir string, logger log.Logger) (*Manager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	stateManager, err := NewStateManager(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create state manager: %w", err)
	}

	snapshotManager, err := NewSnapshotManager(config.Snapshot, dataDir, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot manager: %w", err)
	}

	diskMonitor := NewDiskMonitor(config.DiskMonitor, dataDir, logger)

	return &Manager{
		config:          config,
		logger:          logger.With("module", "pruning-manager"),
		stateManager:    stateManager,
		snapshotManager: snapshotManager,
		diskMonitor:     diskMonitor,
		metrics:         NewMetrics(),
	}, nil
}

// Start starts all pruning-related background processes.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return fmt.Errorf("manager already started")
	}

	ctx, m.cancelFunc = context.WithCancel(ctx)

	// Start disk monitor
	if m.config.DiskMonitor.Enabled {
		m.diskMonitor.SetAlertHandler(m)
		m.diskMonitor.Start(ctx)
	}

	m.started = true
	m.logger.Info("pruning manager started",
		"strategy", m.config.Strategy,
		"keep_recent", m.config.KeepRecent,
		"snapshot_enabled", m.config.Snapshot.Enabled,
		"disk_monitor_enabled", m.config.DiskMonitor.Enabled,
	)

	return nil
}

// Stop stops all pruning-related background processes.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return
	}

	if m.cancelFunc != nil {
		m.cancelFunc()
	}

	m.diskMonitor.Stop()
	m.wg.Wait()
	m.started = false

	m.logger.Info("pruning manager stopped")
}

// OnAlert implements AlertHandler for disk alerts.
func (m *Manager) OnAlert(alert Alert) {
	m.logger.Warn("disk usage alert",
		"level", alert.Level,
		"used_percent", alert.UsedPercent,
		"free_bytes", FormatBytes(alert.FreeBytes),
	)

	// If auto-prune is enabled and critical threshold reached, trigger aggressive pruning
	if m.config.DiskMonitor.AutoPruneOnCritical && alert.Level == AlertLevelCritical {
		m.logger.Warn("critical disk usage - auto-pruning enabled but requires manual action")
	}
}

// GetStatus returns the current status of all pruning components.
func (m *Manager) GetStatus() (Status, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := Status{
		Strategy:           m.config.Strategy,
		KeepRecent:         m.config.KeepRecent,
		PruningInterval:    m.config.Interval,
		SnapshotEnabled:    m.config.Snapshot.Enabled,
		SnapshotInterval:   m.config.Snapshot.Interval,
		DiskMonitorEnabled: m.config.DiskMonitor.Enabled,
		Started:            m.started,
		RetentionPolicy:    m.stateManager.GetRetentionPolicy(),
		LastPrunedHeight:   m.stateManager.LastPrunedHeight(),
		IsPruning:          m.stateManager.IsPruning(),
		IsCreatingSnapshot: m.snapshotManager.IsCreating(),
	}

	// Get disk status
	if m.config.DiskMonitor.Enabled {
		diskStatus, err := m.diskMonitor.GetStatus()
		if err == nil {
			status.DiskStatus = &diskStatus
		}
	}

	// Get snapshot info
	snapshots := m.snapshotManager.GetSnapshots()
	status.SnapshotCount = len(snapshots)
	if len(snapshots) > 0 {
		status.LatestSnapshot = &snapshots[0]
	}

	// Get state sync info
	status.StateSyncInfo = m.snapshotManager.GetStateSyncInfo()

	// Get metrics
	status.Metrics = m.metrics.GetMetrics()

	return status, nil
}

// Status contains the complete status of the pruning system.
type Status struct {
	// Configuration
	Strategy           Strategy `json:"strategy"`
	KeepRecent         uint64   `json:"keep_recent"`
	PruningInterval    uint64   `json:"pruning_interval"`
	SnapshotEnabled    bool     `json:"snapshot_enabled"`
	SnapshotInterval   uint64   `json:"snapshot_interval"`
	DiskMonitorEnabled bool     `json:"disk_monitor_enabled"`

	// State
	Started            bool   `json:"started"`
	RetentionPolicy    string `json:"retention_policy"`
	LastPrunedHeight   int64  `json:"last_pruned_height"`
	IsPruning          bool   `json:"is_pruning"`
	IsCreatingSnapshot bool   `json:"is_creating_snapshot"`

	// Disk
	DiskStatus *DiskUsageStatus `json:"disk_status,omitempty"`

	// Snapshots
	SnapshotCount  int           `json:"snapshot_count"`
	LatestSnapshot *SnapshotInfo `json:"latest_snapshot,omitempty"`
	StateSyncInfo  StateSyncInfo `json:"state_sync_info"`

	// Metrics
	Metrics MetricsSnapshot `json:"metrics"`
}

// ShouldPruneHeight delegates to the state manager.
func (m *Manager) ShouldPruneHeight(currentHeight, targetHeight int64) bool {
	return m.stateManager.ShouldPruneHeight(currentHeight, targetHeight)
}

// ShouldCreateSnapshot delegates to the snapshot manager.
func (m *Manager) ShouldCreateSnapshot(height uint64) bool {
	return m.snapshotManager.ShouldCreateSnapshot(height)
}

// CalculatePruneRange delegates to the state manager.
func (m *Manager) CalculatePruneRange(currentHeight, lastPruned int64) (fromHeight, toHeight int64, shouldPrune bool) {
	return m.stateManager.CalculatePruneRange(currentHeight, lastPruned)
}

// RegisterSnapshot registers a new snapshot.
func (m *Manager) RegisterSnapshot(info SnapshotInfo) error {
	return m.snapshotManager.RegisterSnapshot(info)
}

// GetSnapshots returns all registered snapshots.
func (m *Manager) GetSnapshots() []SnapshotInfo {
	return m.snapshotManager.GetSnapshots()
}

// EstimateStorageSavings estimates storage savings.
func (m *Manager) EstimateStorageSavings(totalBlocks, avgBlockSize int64) StorageSavingsEstimate {
	return m.stateManager.EstimateStorageSavings(totalBlocks, avgBlockSize)
}

// GetDiskProjection returns disk growth projections.
func (m *Manager) GetDiskProjection() GrowthProjection {
	return m.diskMonitor.CalculateGrowthProjection()
}

// GetConfig returns the current configuration.
func (m *Manager) GetConfig() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// UpdateConfig updates the pruning configuration.
func (m *Manager) UpdateConfig(config Config) error {
	if err := config.Validate(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = config

	if err := m.stateManager.UpdateConfig(config); err != nil {
		return err
	}

	m.logger.Info("pruning configuration updated",
		"strategy", config.Strategy,
		"keep_recent", config.KeepRecent,
	)

	return nil
}

// RecordPruningOperation records a pruning operation for metrics.
func (m *Manager) RecordPruningOperation(heightsPruned, bytesPruned int64, duration time.Duration, err error) {
	m.metrics.RecordPruningOperation(heightsPruned, bytesPruned, duration, err)
}

// RecordSnapshotOperation records a snapshot operation for metrics.
func (m *Manager) RecordSnapshotOperation(created bool, size int64, duration time.Duration, err error) {
	m.metrics.RecordSnapshotOperation(created, size, duration, err)
}

// StateManager returns the state manager for advanced operations.
func (m *Manager) StateManager() *StateManager {
	return m.stateManager
}

// SnapshotManager returns the snapshot manager for advanced operations.
func (m *Manager) SnapshotManager() *SnapshotManager {
	return m.snapshotManager
}

// DiskMonitor returns the disk monitor for advanced operations.
func (m *Manager) DiskMonitor() *DiskMonitor {
	return m.diskMonitor
}

// Metrics returns the metrics collector.
func (m *Manager) Metrics() *Metrics {
	return m.metrics
}

// ValidateStateSyncCompatibility validates state sync compatibility.
func (m *Manager) ValidateStateSyncCompatibility() error {
	if !m.config.IsStateSyncCompatible() {
		return ErrStateSyncIncompatible
	}

	return m.snapshotManager.ValidateStateSyncCompatibility(
		m.config.MinRetainBlocks,
		m.config.KeepRecent,
	)
}
