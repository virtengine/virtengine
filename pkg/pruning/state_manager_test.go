package pruning

import (
	"testing"

	"cosmossdk.io/log"
)

func TestNewStateManager(t *testing.T) {
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
			logger := log.NewNopLogger()
			sm, err := NewStateManager(tt.config, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStateManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && sm == nil {
				t.Error("NewStateManager() returned nil state manager")
			}
		})
	}
}

func TestStateManagerShouldPruneHeight(t *testing.T) {
	logger := log.NewNopLogger()

	tests := []struct {
		name          string
		config        Config
		currentHeight int64
		targetHeight  int64
		want          bool
	}{
		{
			name:          "nothing strategy - never prune",
			config:        NothingConfig(),
			currentHeight: 1000,
			targetHeight:  500,
			want:          false,
		},
		{
			name: "custom strategy - height within keep-recent",
			config: Config{
				Strategy:   StrategyCustom,
				KeepRecent: 100,
				Interval:   10,
			},
			currentHeight: 1000,
			targetHeight:  950,
			want:          false,
		},
		{
			name: "custom strategy - height beyond keep-recent",
			config: Config{
				Strategy:   StrategyCustom,
				KeepRecent: 100,
				Interval:   10,
			},
			currentHeight: 1000,
			targetHeight:  800,
			want:          true,
		},
		{
			name:          "everything strategy - old height",
			config:        EverythingConfig(),
			currentHeight: 1000,
			targetHeight:  100,
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm, err := NewStateManager(tt.config, logger)
			if err != nil {
				t.Fatalf("NewStateManager() error = %v", err)
			}

			got := sm.ShouldPruneHeight(tt.currentHeight, tt.targetHeight)
			if got != tt.want {
				t.Errorf("ShouldPruneHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStateManagerTieredPruning(t *testing.T) {
	logger := log.NewNopLogger()

	cfg := TieredRetentionConfig()
	cfg.Tiered.Tier1Blocks = 1000
	cfg.Tiered.Tier2Blocks = 5000
	cfg.Tiered.Tier2SamplingRate = 100
	cfg.Tiered.Tier3SamplingRate = 1000
	cfg.KeepRecent = 1000
	cfg.Interval = 10

	sm, err := NewStateManager(cfg, logger)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	tests := []struct {
		name          string
		currentHeight int64
		targetHeight  int64
		wantPrune     bool
	}{
		{
			name:          "tier 1 - full retention",
			currentHeight: 10000,
			targetHeight:  9500,
			wantPrune:     false,
		},
		{
			name:          "tier 2 - sampled (not on boundary)",
			currentHeight: 10000,
			targetHeight:  7050,
			wantPrune:     true,
		},
		{
			name:          "tier 2 - sampled (on boundary)",
			currentHeight: 10000,
			targetHeight:  7000,
			wantPrune:     false,
		},
		{
			name:          "tier 3 - sampled (not on boundary)",
			currentHeight: 10000,
			targetHeight:  2050,
			wantPrune:     true,
		},
		{
			name:          "tier 3 - sampled (on boundary)",
			currentHeight: 10000,
			targetHeight:  2000,
			wantPrune:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sm.ShouldPruneHeight(tt.currentHeight, tt.targetHeight)
			if got != tt.wantPrune {
				t.Errorf("ShouldPruneHeight() = %v, want %v", got, tt.wantPrune)
			}
		})
	}
}

func TestStateManagerCalculatePruneRange(t *testing.T) {
	logger := log.NewNopLogger()

	tests := []struct {
		name          string
		config        Config
		currentHeight int64
		lastPruned    int64
		wantFrom      int64
		wantTo        int64
		wantPrune     bool
	}{
		{
			name: "should prune - on interval",
			config: Config{
				Strategy:   StrategyCustom,
				KeepRecent: 100,
				Interval:   10,
			},
			currentHeight: 200,
			lastPruned:    50,
			wantFrom:      51,
			wantTo:        100,
			wantPrune:     true,
		},
		{
			name: "should not prune - not on interval",
			config: Config{
				Strategy:   StrategyCustom,
				KeepRecent: 100,
				Interval:   10,
			},
			currentHeight: 205,
			lastPruned:    50,
			wantFrom:      0,
			wantTo:        0,
			wantPrune:     false,
		},
		{
			name:          "nothing strategy - never prune",
			config:        NothingConfig(),
			currentHeight: 1000,
			lastPruned:    0,
			wantFrom:      0,
			wantTo:        0,
			wantPrune:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm, err := NewStateManager(tt.config, logger)
			if err != nil {
				t.Fatalf("NewStateManager() error = %v", err)
			}

			from, to, shouldPrune := sm.CalculatePruneRange(tt.currentHeight, tt.lastPruned)
			if shouldPrune != tt.wantPrune {
				t.Errorf("CalculatePruneRange() shouldPrune = %v, want %v", shouldPrune, tt.wantPrune)
			}
			if shouldPrune {
				if from != tt.wantFrom {
					t.Errorf("CalculatePruneRange() from = %d, want %d", from, tt.wantFrom)
				}
				if to != tt.wantTo {
					t.Errorf("CalculatePruneRange() to = %d, want %d", to, tt.wantTo)
				}
			}
		})
	}
}

func TestStateManagerPruningInProgress(t *testing.T) {
	logger := log.NewNopLogger()
	cfg := DefaultConfig()

	sm, err := NewStateManager(cfg, logger)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	// Initially not pruning
	if sm.IsPruning() {
		t.Error("IsPruning() should be false initially")
	}

	// Start pruning
	if err := sm.StartPruning(); err != nil {
		t.Errorf("StartPruning() error = %v", err)
	}

	if !sm.IsPruning() {
		t.Error("IsPruning() should be true after StartPruning()")
	}

	// Try to start again - should fail
	if err := sm.StartPruning(); err != ErrPruningInProgress {
		t.Errorf("StartPruning() should return ErrPruningInProgress, got %v", err)
	}

	// End pruning
	sm.EndPruning(1000)

	if sm.IsPruning() {
		t.Error("IsPruning() should be false after EndPruning()")
	}

	if sm.LastPrunedHeight() != 1000 {
		t.Errorf("LastPrunedHeight() = %d, want 1000", sm.LastPrunedHeight())
	}
}

func TestStateManagerGetRetentionPolicy(t *testing.T) {
	logger := log.NewNopLogger()

	tests := []struct {
		name   string
		config Config
		want   string
	}{
		{
			name:   "nothing",
			config: NothingConfig(),
			want:   "Archive mode: all states are retained",
		},
		{
			name:   "everything",
			config: EverythingConfig(),
			want:   "Minimal mode: only 2 most recent states are kept",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm, err := NewStateManager(tt.config, logger)
			if err != nil {
				t.Fatalf("NewStateManager() error = %v", err)
			}

			got := sm.GetRetentionPolicy()
			if got != tt.want {
				t.Errorf("GetRetentionPolicy() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStateManagerEstimateStorageSavings(t *testing.T) {
	logger := log.NewNopLogger()

	tests := []struct {
		name         string
		config       Config
		totalBlocks  int64
		avgBlockSize int64
		wantSavings  bool
	}{
		{
			name:         "nothing - no savings",
			config:       NothingConfig(),
			totalBlocks:  1000000,
			avgBlockSize: 1024,
			wantSavings:  false,
		},
		{
			name:         "everything - max savings",
			config:       EverythingConfig(),
			totalBlocks:  1000000,
			avgBlockSize: 1024,
			wantSavings:  true,
		},
		{
			name:         "default - some savings",
			config:       DefaultConfig(),
			totalBlocks:  1000000,
			avgBlockSize: 1024,
			wantSavings:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm, err := NewStateManager(tt.config, logger)
			if err != nil {
				t.Fatalf("NewStateManager() error = %v", err)
			}

			estimate := sm.EstimateStorageSavings(tt.totalBlocks, tt.avgBlockSize)

			if tt.wantSavings && estimate.SavingsAmount == 0 {
				t.Error("Expected savings but got 0")
			}
			if !tt.wantSavings && estimate.SavingsAmount != 0 {
				t.Errorf("Expected no savings but got %d", estimate.SavingsAmount)
			}
		})
	}
}

func TestStateManagerGetHeightStatus(t *testing.T) {
	logger := log.NewNopLogger()
	cfg := Config{
		Strategy:   StrategyCustom,
		KeepRecent: 100,
		Interval:   10,
	}

	sm, err := NewStateManager(cfg, logger)
	if err != nil {
		t.Fatalf("NewStateManager() error = %v", err)
	}

	// Height within keep-recent
	status := sm.GetHeightStatus(1000, 950)
	if status.IsPruned {
		t.Error("Height 950 should not be pruned")
	}

	// Height beyond keep-recent
	status = sm.GetHeightStatus(1000, 800)
	if !status.IsPruned {
		t.Error("Height 800 should be pruned")
	}
}

