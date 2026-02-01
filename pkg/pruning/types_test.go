package pruning

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Strategy != StrategyDefault {
		t.Errorf("expected strategy %s, got %s", StrategyDefault, cfg.Strategy)
	}

	if cfg.KeepRecent != DefaultKeepRecent {
		t.Errorf("expected keep_recent %d, got %d", DefaultKeepRecent, cfg.KeepRecent)
	}

	if cfg.Interval != DefaultInterval {
		t.Errorf("expected interval %d, got %d", DefaultInterval, cfg.Interval)
	}

	if !cfg.Snapshot.Enabled {
		t.Error("expected snapshots to be enabled by default")
	}

	if !cfg.DiskMonitor.Enabled {
		t.Error("expected disk monitor to be enabled by default")
	}

	if !cfg.HistoricalQueries.Enabled {
		t.Error("expected historical queries to be enabled by default")
	}
}

func TestNothingConfig(t *testing.T) {
	cfg := NothingConfig()

	if cfg.Strategy != StrategyNothing {
		t.Errorf("expected strategy %s, got %s", StrategyNothing, cfg.Strategy)
	}

	if cfg.KeepRecent != 0 {
		t.Errorf("expected keep_recent 0, got %d", cfg.KeepRecent)
	}

	if cfg.Interval != 0 {
		t.Errorf("expected interval 0, got %d", cfg.Interval)
	}
}

func TestEverythingConfig(t *testing.T) {
	cfg := EverythingConfig()

	if cfg.Strategy != StrategyEverything {
		t.Errorf("expected strategy %s, got %s", StrategyEverything, cfg.Strategy)
	}

	if cfg.KeepRecent != 2 {
		t.Errorf("expected keep_recent 2, got %d", cfg.KeepRecent)
	}

	if cfg.HistoricalQueries.Enabled {
		t.Error("expected historical queries to be disabled for everything strategy")
	}
}

func TestTieredRetentionConfig(t *testing.T) {
	cfg := TieredRetentionConfig()

	if cfg.Strategy != StrategyTiered {
		t.Errorf("expected strategy %s, got %s", StrategyTiered, cfg.Strategy)
	}

	if !cfg.Tiered.Enabled {
		t.Error("expected tiered retention to be enabled")
	}
}

func TestConfigValidation(t *testing.T) {
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
			name:    "valid everything config",
			config:  EverythingConfig(),
			wantErr: false,
		},
		{
			name: "invalid strategy",
			config: Config{
				Strategy:   "invalid",
				KeepRecent: 100,
				Interval:   10,
			},
			wantErr: true,
		},
		{
			name: "zero keep_recent for custom",
			config: Config{
				Strategy:   StrategyCustom,
				KeepRecent: 0,
				Interval:   10,
			},
			wantErr: true,
		},
		{
			name: "zero interval for custom",
			config: Config{
				Strategy:   StrategyCustom,
				KeepRecent: 100,
				Interval:   0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigPruningOptions(t *testing.T) {
	tests := []struct {
		name         string
		strategy     Strategy
		wantRecent   uint64
		wantInterval uint64
	}{
		{
			name:         "nothing",
			strategy:     StrategyNothing,
			wantRecent:   0,
			wantInterval: 0,
		},
		{
			name:         "everything",
			strategy:     StrategyEverything,
			wantRecent:   2,
			wantInterval: 10,
		},
		{
			name:         "default",
			strategy:     StrategyDefault,
			wantRecent:   DefaultKeepRecent,
			wantInterval: DefaultInterval,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Strategy: tt.strategy, KeepRecent: 100, Interval: 5}
			recent, interval := cfg.PruningOptions()
			if recent != tt.wantRecent {
				t.Errorf("PruningOptions() recent = %d, want %d", recent, tt.wantRecent)
			}
			if interval != tt.wantInterval {
				t.Errorf("PruningOptions() interval = %d, want %d", interval, tt.wantInterval)
			}
		})
	}
}

func TestConfigIsStateSyncCompatible(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   bool
	}{
		{
			name:   "default config",
			config: DefaultConfig(),
			want:   true,
		},
		{
			name:   "nothing config",
			config: NothingConfig(),
			want:   true,
		},
		{
			name:   "everything config",
			config: EverythingConfig(),
			want:   false,
		},
		{
			name: "snapshots disabled",
			config: Config{
				Strategy:   StrategyDefault,
				KeepRecent: 1000,
				Interval:   10,
				Snapshot:   SnapshotConfig{Enabled: false},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.IsStateSyncCompatible(); got != tt.want {
				t.Errorf("IsStateSyncCompatible() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiskMonitorConfigDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DiskMonitor.WarningThresholdPercent != DefaultDiskWarningThresholdPercent {
		t.Errorf("expected warning threshold %f, got %f",
			DefaultDiskWarningThresholdPercent, cfg.DiskMonitor.WarningThresholdPercent)
	}

	if cfg.DiskMonitor.CriticalThresholdPercent != DefaultDiskCriticalThresholdPercent {
		t.Errorf("expected critical threshold %f, got %f",
			DefaultDiskCriticalThresholdPercent, cfg.DiskMonitor.CriticalThresholdPercent)
	}

	if cfg.DiskMonitor.CheckInterval != 5*time.Minute {
		t.Errorf("expected check interval 5m, got %v", cfg.DiskMonitor.CheckInterval)
	}
}

func TestSnapshotConfigDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Snapshot.Interval != DefaultSnapshotInterval {
		t.Errorf("expected snapshot interval %d, got %d",
			DefaultSnapshotInterval, cfg.Snapshot.Interval)
	}

	if cfg.Snapshot.KeepRecent != DefaultSnapshotKeepRecent {
		t.Errorf("expected snapshot keep recent %d, got %d",
			DefaultSnapshotKeepRecent, cfg.Snapshot.KeepRecent)
	}

	if !cfg.Snapshot.Compression {
		t.Error("expected compression to be enabled by default")
	}
}

func TestTieredConfigDefaults(t *testing.T) {
	cfg := TieredRetentionConfig()

	if cfg.Tiered.Tier1Blocks != Tier1RetentionBlocks {
		t.Errorf("expected tier1 blocks %d, got %d",
			Tier1RetentionBlocks, cfg.Tiered.Tier1Blocks)
	}

	if cfg.Tiered.Tier2Blocks != Tier2RetentionBlocks {
		t.Errorf("expected tier2 blocks %d, got %d",
			Tier2RetentionBlocks, cfg.Tiered.Tier2Blocks)
	}

	if cfg.Tiered.Tier3SamplingRate != Tier3SamplingRate {
		t.Errorf("expected tier3 sampling rate %d, got %d",
			Tier3SamplingRate, cfg.Tiered.Tier3SamplingRate)
	}
}

