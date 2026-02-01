package pruning

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/log"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name:    "valid nothing config",
			config:  NothingConfig(),
			wantErr: false,
		},
		{
			name: "invalid config",
			config: Config{
				Strategy: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewManager(tt.config, tmpDir, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && manager == nil {
				t.Error("NewManager() returned nil manager")
			}
		})
	}
}

func TestManagerStartStop(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()
	cfg := DefaultConfig()
	cfg.DiskMonitor.Enabled = false // Disable to avoid background goroutines

	manager, err := NewManager(cfg, tmpDir, logger)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	ctx := context.Background()

	// Start
	if err := manager.Start(ctx); err != nil {
		t.Errorf("Start() error = %v", err)
	}

	// Double start should fail
	if err := manager.Start(ctx); err == nil {
		t.Error("Double Start() should fail")
	}

	// Stop
	manager.Stop()

	// Double stop should not panic
	manager.Stop()
}

func TestManagerGetStatus(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()
	cfg := DefaultConfig()
	cfg.DiskMonitor.Enabled = false

	manager, err := NewManager(cfg, tmpDir, logger)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	status, err := manager.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if status.Strategy != StrategyDefault {
		t.Errorf("Strategy = %s, want %s", status.Strategy, StrategyDefault)
	}

	if status.KeepRecent != DefaultKeepRecent {
		t.Errorf("KeepRecent = %d, want %d", status.KeepRecent, DefaultKeepRecent)
	}

	if !status.SnapshotEnabled {
		t.Error("SnapshotEnabled should be true")
	}
}

func TestManagerEstimateStorageSavings(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()
	cfg := DefaultConfig()

	manager, err := NewManager(cfg, tmpDir, logger)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	estimate := manager.EstimateStorageSavings(1000000, 1024)

	if estimate.TotalBlocks != 1000000 {
		t.Errorf("TotalBlocks = %d, want 1000000", estimate.TotalBlocks)
	}

	if estimate.SavingsAmount <= 0 {
		t.Error("Expected positive savings for default strategy")
	}
}

func TestManagerUpdateConfig(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()
	cfg := DefaultConfig()

	manager, err := NewManager(cfg, tmpDir, logger)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Update to custom config
	newCfg := Config{
		Strategy:   StrategyCustom,
		KeepRecent: 50000,
		Interval:   20,
		Snapshot: SnapshotConfig{
			Enabled:    true,
			Interval:   2000,
			KeepRecent: 3,
		},
	}

	if err := manager.UpdateConfig(newCfg); err != nil {
		t.Errorf("UpdateConfig() error = %v", err)
	}

	// Verify update
	updatedCfg := manager.GetConfig()
	if updatedCfg.Strategy != StrategyCustom {
		t.Errorf("Strategy = %s, want %s", updatedCfg.Strategy, StrategyCustom)
	}
	if updatedCfg.KeepRecent != 50000 {
		t.Errorf("KeepRecent = %d, want 50000", updatedCfg.KeepRecent)
	}
}

func TestManagerValidateStateSyncCompatibility(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "default config - compatible",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name:    "nothing config - compatible",
			config:  NothingConfig(),
			wantErr: false,
		},
		{
			name:    "everything config - incompatible",
			config:  EverythingConfig(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewManager(tt.config, tmpDir, logger)
			if err != nil {
				t.Fatalf("NewManager() error = %v", err)
			}

			err = manager.ValidateStateSyncCompatibility()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStateSyncCompatibility() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManagerRecordMetrics(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()
	cfg := DefaultConfig()

	manager, err := NewManager(cfg, tmpDir, logger)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Record operations
	manager.RecordPruningOperation(100, 1024, 5*time.Second, nil)
	manager.RecordSnapshotOperation(true, 1024*1024, 30*time.Second, nil)

	// Verify metrics
	metrics := manager.Metrics().GetMetrics()
	if metrics.TotalPruningOperations != 1 {
		t.Errorf("TotalPruningOperations = %d, want 1", metrics.TotalPruningOperations)
	}
	if metrics.TotalSnapshotsCreated != 1 {
		t.Errorf("TotalSnapshotsCreated = %d, want 1", metrics.TotalSnapshotsCreated)
	}
}

func TestManagerGetSubManagers(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()
	cfg := DefaultConfig()

	manager, err := NewManager(cfg, tmpDir, logger)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if manager.StateManager() == nil {
		t.Error("StateManager() returned nil")
	}

	if manager.SnapshotManager() == nil {
		t.Error("SnapshotManager() returned nil")
	}

	if manager.DiskMonitor() == nil {
		t.Error("DiskMonitor() returned nil")
	}

	if manager.Metrics() == nil {
		t.Error("Metrics() returned nil")
	}
}

