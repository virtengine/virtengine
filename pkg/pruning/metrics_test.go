package pruning

import (
	"testing"
	"time"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()
	if m == nil {
		t.Fatal("NewMetrics() returned nil")
	}
}

func TestMetricsRecordPruningOperation(t *testing.T) {
	m := NewMetrics()

	// Record successful operation
	m.RecordPruningOperation(100, 1024*1024, 5*time.Second, nil)

	metrics := m.GetMetrics()

	if metrics.TotalPruningOperations != 1 {
		t.Errorf("TotalPruningOperations = %d, want 1", metrics.TotalPruningOperations)
	}

	if metrics.TotalHeightsPruned != 100 {
		t.Errorf("TotalHeightsPruned = %d, want 100", metrics.TotalHeightsPruned)
	}

	if metrics.TotalBytesPruned != 1024*1024 {
		t.Errorf("TotalBytesPruned = %d, want %d", metrics.TotalBytesPruned, 1024*1024)
	}

	if metrics.LastPruningDuration != 5*time.Second {
		t.Errorf("LastPruningDuration = %v, want 5s", metrics.LastPruningDuration)
	}

	if metrics.PruningErrorCount != 0 {
		t.Errorf("PruningErrorCount = %d, want 0", metrics.PruningErrorCount)
	}
}

func TestMetricsRecordPruningOperationWithError(t *testing.T) {
	m := NewMetrics()

	// Record failed operation
	m.RecordPruningOperation(0, 0, 1*time.Second, ErrPruningInProgress)

	metrics := m.GetMetrics()

	if metrics.PruningErrorCount != 1 {
		t.Errorf("PruningErrorCount = %d, want 1", metrics.PruningErrorCount)
	}
}

func TestMetricsRecordSnapshotOperation(t *testing.T) {
	m := NewMetrics()

	// Record created snapshot
	m.RecordSnapshotOperation(true, 1024*1024*100, 30*time.Second, nil)

	metrics := m.GetMetrics()

	if metrics.TotalSnapshotsCreated != 1 {
		t.Errorf("TotalSnapshotsCreated = %d, want 1", metrics.TotalSnapshotsCreated)
	}

	if metrics.LastSnapshotSize != 1024*1024*100 {
		t.Errorf("LastSnapshotSize = %d, want %d", metrics.LastSnapshotSize, 1024*1024*100)
	}

	if metrics.LastSnapshotDuration != 30*time.Second {
		t.Errorf("LastSnapshotDuration = %v, want 30s", metrics.LastSnapshotDuration)
	}

	// Record deleted snapshot
	m.RecordSnapshotOperation(false, 0, 1*time.Second, nil)

	metrics = m.GetMetrics()

	if metrics.TotalSnapshotsDeleted != 1 {
		t.Errorf("TotalSnapshotsDeleted = %d, want 1", metrics.TotalSnapshotsDeleted)
	}
}

func TestMetricsUpdateDiskMetrics(t *testing.T) {
	m := NewMetrics()

	m.UpdateDiskMetrics(500*1024*1024*1024, 75.5, 1024*1024*1024, 90)

	metrics := m.GetMetrics()

	if metrics.CurrentDiskUsage != 500*1024*1024*1024 {
		t.Errorf("CurrentDiskUsage = %d, want %d", metrics.CurrentDiskUsage, 500*1024*1024*1024)
	}

	if metrics.CurrentDiskUsagePercent != 75.5 {
		t.Errorf("CurrentDiskUsagePercent = %f, want 75.5", metrics.CurrentDiskUsagePercent)
	}

	if metrics.DailyGrowthRate != 1024*1024*1024 {
		t.Errorf("DailyGrowthRate = %d, want %d", metrics.DailyGrowthRate, 1024*1024*1024)
	}

	if metrics.DaysUntilFull != 90 {
		t.Errorf("DaysUntilFull = %d, want 90", metrics.DaysUntilFull)
	}
}

func TestMetricsUpdateStateMetrics(t *testing.T) {
	m := NewMetrics()

	m.UpdateStateMetrics(10000, 5000, 5000, 5)

	metrics := m.GetMetrics()

	if metrics.CurrentHeight != 10000 {
		t.Errorf("CurrentHeight = %d, want 10000", metrics.CurrentHeight)
	}

	if metrics.OldestRetainedHeight != 5000 {
		t.Errorf("OldestRetainedHeight = %d, want 5000", metrics.OldestRetainedHeight)
	}

	if metrics.RetainedHeightsCount != 5000 {
		t.Errorf("RetainedHeightsCount = %d, want 5000", metrics.RetainedHeightsCount)
	}

	if metrics.SnapshotsCount != 5 {
		t.Errorf("SnapshotsCount = %d, want 5", metrics.SnapshotsCount)
	}
}

func TestMetricsReset(t *testing.T) {
	m := NewMetrics()

	// Record some data
	m.RecordPruningOperation(100, 1024, 1*time.Second, nil)
	m.RecordSnapshotOperation(true, 1024, 1*time.Second, nil)

	// Reset
	m.Reset()

	metrics := m.GetMetrics()

	if metrics.TotalPruningOperations != 0 {
		t.Error("Metrics not reset properly")
	}

	if metrics.TotalSnapshotsCreated != 0 {
		t.Error("Metrics not reset properly")
	}
}

func TestMetricsAverageCalculation(t *testing.T) {
	m := NewMetrics()

	// Record multiple operations
	m.RecordPruningOperation(100, 1024, 2*time.Second, nil)
	m.RecordPruningOperation(100, 1024, 4*time.Second, nil)
	m.RecordPruningOperation(100, 1024, 6*time.Second, nil)

	metrics := m.GetMetrics()

	// Average should be calculated
	if metrics.AveragePruningDuration < 3*time.Second || metrics.AveragePruningDuration > 5*time.Second {
		t.Errorf("AveragePruningDuration = %v, expected around 4s", metrics.AveragePruningDuration)
	}
}

func TestMetricsTelemetryLabels(t *testing.T) {
	m := NewMetrics()

	m.UpdateStateMetrics(10000, 5000, 5000, 5)
	m.UpdateDiskMetrics(500*1024*1024*1024, 75.5, 1024*1024, 90)

	labels := m.TelemetryLabels()

	if labels == nil {
		t.Fatal("TelemetryLabels() returned nil")
	}

	// Check that labels exist
	requiredLabels := []string{
		"current_height",
		"oldest_height",
		"retained_count",
		"snapshots",
		"disk_usage_percent",
		"daily_growth_rate",
		"days_until_full",
	}

	for _, label := range requiredLabels {
		if _, ok := labels[label]; !ok {
			t.Errorf("Missing required label: %s", label)
		}
	}
}

func TestNewBenchmark(t *testing.T) {
	cfg := DefaultConfig()
	b := NewBenchmark(cfg)

	if b == nil {
		t.Fatal("NewBenchmark() returned nil")
	}
}

func TestBenchmarkRunSimulation(t *testing.T) {
	tests := []struct {
		name           string
		config         Config
		totalBlocks    int64
		avgBlockSize   int64
		wantSavings    bool
		wantSavingsPct float64
	}{
		{
			name:           "nothing strategy",
			config:         NothingConfig(),
			totalBlocks:    1000000,
			avgBlockSize:   1024,
			wantSavings:    false,
			wantSavingsPct: 0,
		},
		{
			name:           "everything strategy",
			config:         EverythingConfig(),
			totalBlocks:    1000000,
			avgBlockSize:   1024,
			wantSavings:    true,
			wantSavingsPct: 99.9, // Almost everything pruned
		},
		{
			name:           "default strategy",
			config:         DefaultConfig(),
			totalBlocks:    1000000,
			avgBlockSize:   1024,
			wantSavings:    true,
			wantSavingsPct: 60, // Significant savings
		},
		{
			name:           "tiered strategy",
			config:         TieredRetentionConfig(),
			totalBlocks:    1000000,
			avgBlockSize:   1024,
			wantSavings:    true,
			wantSavingsPct: 50, // Moderate savings with sampling
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBenchmark(tt.config)
			result := b.RunSimulation(tt.totalBlocks, tt.avgBlockSize)

			if result.Strategy != tt.config.Strategy {
				t.Errorf("Strategy = %s, want %s", result.Strategy, tt.config.Strategy)
			}

			if result.TotalBlocks != tt.totalBlocks {
				t.Errorf("TotalBlocks = %d, want %d", result.TotalBlocks, tt.totalBlocks)
			}

			hasSavings := result.EstimatedSavings > 0
			if hasSavings != tt.wantSavings {
				t.Errorf("hasSavings = %v, want %v", hasSavings, tt.wantSavings)
			}

			if tt.wantSavings && result.SavingsPercent < tt.wantSavingsPct {
				t.Errorf("SavingsPercent = %.1f, want >= %.1f", result.SavingsPercent, tt.wantSavingsPct)
			}
		})
	}
}

func TestCompareBenchmarks(t *testing.T) {
	results := CompareBenchmarks(1000000, 1024)

	expectedStrategies := []Strategy{
		StrategyDefault,
		StrategyNothing,
		StrategyEverything,
		StrategyTiered,
	}

	for _, strategy := range expectedStrategies {
		if _, ok := results[strategy]; !ok {
			t.Errorf("Missing benchmark result for strategy: %s", strategy)
		}
	}

	// Verify relative savings
	nothingResult := results[StrategyNothing]
	everythingResult := results[StrategyEverything]

	if nothingResult.SavingsPercent != 0 {
		t.Error("Nothing strategy should have 0% savings")
	}

	if everythingResult.SavingsPercent < 99 {
		t.Error("Everything strategy should have > 99% savings")
	}
}

