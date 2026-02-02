package pruning

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"cosmossdk.io/log"
)

func TestNewSnapshotManager(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()

	cfg := SnapshotConfig{
		Enabled:    true,
		Interval:   1000,
		KeepRecent: 2,
	}

	sm, err := NewSnapshotManager(cfg, tmpDir, logger)
	if err != nil {
		t.Fatalf("NewSnapshotManager() error = %v", err)
	}

	if sm == nil {
		t.Fatal("NewSnapshotManager() returned nil")
	}

	// Verify directory was created
	snapshotDir := filepath.Join(tmpDir, "data", "snapshots")
	if _, err := os.Stat(snapshotDir); os.IsNotExist(err) {
		t.Error("Snapshot directory was not created")
	}
}

func TestSnapshotManagerShouldCreateSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()

	tests := []struct {
		name     string
		config   SnapshotConfig
		height   uint64
		expected bool
	}{
		{
			name: "disabled",
			config: SnapshotConfig{
				Enabled:  false,
				Interval: 1000,
			},
			height:   1000,
			expected: false,
		},
		{
			name: "enabled - on interval",
			config: SnapshotConfig{
				Enabled:    true,
				Interval:   1000,
				KeepRecent: 2,
			},
			height:   1000,
			expected: true,
		},
		{
			name: "enabled - not on interval",
			config: SnapshotConfig{
				Enabled:    true,
				Interval:   1000,
				KeepRecent: 2,
			},
			height:   1001,
			expected: false,
		},
		{
			name: "enabled - zero height",
			config: SnapshotConfig{
				Enabled:    true,
				Interval:   1000,
				KeepRecent: 2,
			},
			height:   0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm, err := NewSnapshotManager(tt.config, tmpDir, logger)
			if err != nil {
				t.Fatalf("NewSnapshotManager() error = %v", err)
			}

			got := sm.ShouldCreateSnapshot(tt.height)
			if got != tt.expected {
				t.Errorf("ShouldCreateSnapshot(%d) = %v, want %v", tt.height, got, tt.expected)
			}
		})
	}
}

func TestSnapshotManagerRegisterSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()

	cfg := SnapshotConfig{
		Enabled:    true,
		Interval:   1000,
		KeepRecent: 2,
	}

	sm, err := NewSnapshotManager(cfg, tmpDir, logger)
	if err != nil {
		t.Fatalf("NewSnapshotManager() error = %v", err)
	}

	// Register first snapshot
	info1 := SnapshotInfo{
		Height:    1000,
		Format:    1,
		Chunks:    10,
		CreatedAt: time.Now(),
		Size:      1024 * 1024,
	}

	if err := sm.RegisterSnapshot(info1); err != nil {
		t.Errorf("RegisterSnapshot() error = %v", err)
	}

	// Verify it was registered
	snapshots := sm.GetSnapshots()
	if len(snapshots) != 1 {
		t.Errorf("Expected 1 snapshot, got %d", len(snapshots))
	}

	// Register second snapshot
	info2 := SnapshotInfo{
		Height:    2000,
		Format:    1,
		Chunks:    10,
		CreatedAt: time.Now(),
		Size:      1024 * 1024,
	}

	if err := sm.RegisterSnapshot(info2); err != nil {
		t.Errorf("RegisterSnapshot() error = %v", err)
	}

	snapshots = sm.GetSnapshots()
	if len(snapshots) != 2 {
		t.Errorf("Expected 2 snapshots, got %d", len(snapshots))
	}

	// Verify sorting (newest first)
	if snapshots[0].Height != 2000 {
		t.Error("Snapshots not sorted correctly")
	}
}

func TestSnapshotManagerDuplicateSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()

	cfg := SnapshotConfig{
		Enabled:    true,
		Interval:   1000,
		KeepRecent: 2,
	}

	sm, err := NewSnapshotManager(cfg, tmpDir, logger)
	if err != nil {
		t.Fatalf("NewSnapshotManager() error = %v", err)
	}

	info := SnapshotInfo{
		Height:    1000,
		Format:    1,
		CreatedAt: time.Now(),
	}

	if err := sm.RegisterSnapshot(info); err != nil {
		t.Errorf("First RegisterSnapshot() error = %v", err)
	}

	// Try to register duplicate
	if err := sm.RegisterSnapshot(info); err == nil {
		t.Error("Expected error for duplicate snapshot")
	}
}

func TestSnapshotManagerGetSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()

	cfg := SnapshotConfig{
		Enabled:    true,
		Interval:   1000,
		KeepRecent: 3,
	}

	sm, err := NewSnapshotManager(cfg, tmpDir, logger)
	if err != nil {
		t.Fatalf("NewSnapshotManager() error = %v", err)
	}

	// Register snapshots
	for _, height := range []uint64{1000, 2000, 3000} {
		info := SnapshotInfo{
			Height:    height,
			Format:    1,
			CreatedAt: time.Now(),
		}
		if err := sm.RegisterSnapshot(info); err != nil {
			t.Fatalf("RegisterSnapshot() error = %v", err)
		}
	}

	// Get existing snapshot
	snapshot, found := sm.GetSnapshot(2000)
	if !found {
		t.Error("Expected to find snapshot at height 2000")
	}
	if snapshot.Height != 2000 {
		t.Errorf("GetSnapshot() height = %d, want 2000", snapshot.Height)
	}

	// Get non-existing snapshot
	_, found = sm.GetSnapshot(999)
	if found {
		t.Error("Expected not to find snapshot at height 999")
	}
}

func TestSnapshotManagerGetLatestSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()

	cfg := SnapshotConfig{
		Enabled:    true,
		Interval:   1000,
		KeepRecent: 3,
	}

	sm, err := NewSnapshotManager(cfg, tmpDir, logger)
	if err != nil {
		t.Fatalf("NewSnapshotManager() error = %v", err)
	}

	// No snapshots
	_, found := sm.GetLatestSnapshot()
	if found {
		t.Error("Expected not to find latest snapshot when none exist")
	}

	// Register snapshots
	for _, height := range []uint64{1000, 2000, 3000} {
		info := SnapshotInfo{
			Height:    height,
			Format:    1,
			CreatedAt: time.Now(),
		}
		if err := sm.RegisterSnapshot(info); err != nil {
			t.Fatalf("RegisterSnapshot() error = %v", err)
		}
	}

	// Get latest
	latest, found := sm.GetLatestSnapshot()
	if !found {
		t.Fatal("Expected to find latest snapshot")
	}
	if latest.Height != 3000 {
		t.Errorf("GetLatestSnapshot() height = %d, want 3000", latest.Height)
	}
}

func TestSnapshotManagerDeleteSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()

	cfg := SnapshotConfig{
		Enabled:    true,
		Interval:   1000,
		KeepRecent: 3,
	}

	sm, err := NewSnapshotManager(cfg, tmpDir, logger)
	if err != nil {
		t.Fatalf("NewSnapshotManager() error = %v", err)
	}

	// Register snapshot
	info := SnapshotInfo{
		Height:    1000,
		Format:    1,
		CreatedAt: time.Now(),
	}
	if err := sm.RegisterSnapshot(info); err != nil {
		t.Fatalf("RegisterSnapshot() error = %v", err)
	}

	// Delete snapshot
	if err := sm.DeleteSnapshot(1000); err != nil {
		t.Errorf("DeleteSnapshot() error = %v", err)
	}

	// Verify deleted
	_, found := sm.GetSnapshot(1000)
	if found {
		t.Error("Snapshot should have been deleted")
	}

	// Delete non-existing
	if err := sm.DeleteSnapshot(999); err != ErrSnapshotNotFound {
		t.Errorf("Expected ErrSnapshotNotFound, got %v", err)
	}
}

func TestSnapshotManagerCreatingState(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()

	cfg := SnapshotConfig{
		Enabled:    true,
		Interval:   1000,
		KeepRecent: 2,
	}

	sm, err := NewSnapshotManager(cfg, tmpDir, logger)
	if err != nil {
		t.Fatalf("NewSnapshotManager() error = %v", err)
	}

	// Initially not creating
	if sm.IsCreating() {
		t.Error("IsCreating() should be false initially")
	}

	// Start creating
	if err := sm.StartCreating(); err != nil {
		t.Errorf("StartCreating() error = %v", err)
	}

	if !sm.IsCreating() {
		t.Error("IsCreating() should be true after StartCreating()")
	}

	// Try to start again
	if err := sm.StartCreating(); err != ErrSnapshotInProgress {
		t.Errorf("StartCreating() should return ErrSnapshotInProgress, got %v", err)
	}

	// End creating
	sm.EndCreating()

	if sm.IsCreating() {
		t.Error("IsCreating() should be false after EndCreating()")
	}
}

func TestSnapshotManagerGetTotalSize(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()

	cfg := SnapshotConfig{
		Enabled:    true,
		Interval:   1000,
		KeepRecent: 5,
	}

	sm, err := NewSnapshotManager(cfg, tmpDir, logger)
	if err != nil {
		t.Fatalf("NewSnapshotManager() error = %v", err)
	}

	// Register snapshots with sizes
	sizes := []int64{1024, 2048, 4096}
	for i, size := range sizes {
		info := SnapshotInfo{
			//nolint:gosec // G115: i is a small loop index in test code
			Height:    uint64((i + 1) * 1000),
			Format:    1,
			Size:      size,
			CreatedAt: time.Now(),
		}
		if err := sm.RegisterSnapshot(info); err != nil {
			t.Fatalf("RegisterSnapshot() error = %v", err)
		}
	}

	expectedTotal := int64(1024 + 2048 + 4096)
	total := sm.GetTotalSize()
	if total != expectedTotal {
		t.Errorf("GetTotalSize() = %d, want %d", total, expectedTotal)
	}
}

func TestSnapshotManagerStateSyncInfo(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()

	cfg := SnapshotConfig{
		Enabled:    true,
		Interval:   1000,
		KeepRecent: 2,
	}

	sm, err := NewSnapshotManager(cfg, tmpDir, logger)
	if err != nil {
		t.Fatalf("NewSnapshotManager() error = %v", err)
	}

	// Register snapshot
	info := SnapshotInfo{
		Height:    5000,
		Format:    1,
		CreatedAt: time.Now(),
	}
	if err := sm.RegisterSnapshot(info); err != nil {
		t.Fatalf("RegisterSnapshot() error = %v", err)
	}

	// Get state sync info
	syncInfo := sm.GetStateSyncInfo()

	if !syncInfo.Enabled {
		t.Error("Expected state sync to be enabled")
	}

	if syncInfo.LatestSnapshot == nil {
		t.Fatal("Expected latest snapshot to be set")
	}

	if syncInfo.LatestSnapshot.Height != 5000 {
		t.Errorf("Expected trust height 5000, got %d", syncInfo.TrustHeight)
	}

	if syncInfo.SnapshotInterval != 1000 {
		t.Errorf("Expected snapshot interval 1000, got %d", syncInfo.SnapshotInterval)
	}
}

func TestSnapshotManagerValidateStateSyncCompatibility(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()

	tests := []struct {
		name            string
		config          SnapshotConfig
		minRetainBlocks uint64
		keepRecent      uint64
		wantErr         bool
	}{
		{
			name: "valid configuration",
			config: SnapshotConfig{
				Enabled:    true,
				Interval:   1000,
				KeepRecent: 2,
			},
			minRetainBlocks: 2000,
			keepRecent:      10000,
			wantErr:         false,
		},
		{
			name: "disabled snapshots",
			config: SnapshotConfig{
				Enabled:    false,
				Interval:   1000,
				KeepRecent: 2,
			},
			wantErr: true,
		},
		{
			name: "too few snapshots",
			config: SnapshotConfig{
				Enabled:    true,
				Interval:   1000,
				KeepRecent: 1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm, err := NewSnapshotManager(tt.config, tmpDir, logger)
			if err != nil {
				t.Fatalf("NewSnapshotManager() error = %v", err)
			}

			err = sm.ValidateStateSyncCompatibility(tt.minRetainBlocks, tt.keepRecent)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStateSyncCompatibility() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
