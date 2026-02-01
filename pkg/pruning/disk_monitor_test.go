package pruning

import (
	"testing"
	"time"

	"cosmossdk.io/log"
)

func TestNewDiskMonitor(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()

	cfg := DiskMonitorConfig{
		Enabled:                  true,
		CheckInterval:            5 * time.Minute,
		WarningThresholdPercent:  80.0,
		CriticalThresholdPercent: 90.0,
	}

	dm := NewDiskMonitor(cfg, tmpDir, logger)
	if dm == nil {
		t.Fatal("NewDiskMonitor() returned nil")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes uint64
		want  string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.00 KB"},
		{1024 * 1024, "1.00 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
		{1024 * 1024 * 1024 * 1024, "1.00 TB"},
		{1536 * 1024 * 1024, "1.50 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatBytes(tt.bytes)
			if got != tt.want {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestDiskMonitorGrowthProjection(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()

	cfg := DiskMonitorConfig{
		Enabled:                  true,
		CheckInterval:            5 * time.Minute,
		WarningThresholdPercent:  80.0,
		CriticalThresholdPercent: 90.0,
		GrowthProjectionDays:     30,
	}

	dm := NewDiskMonitor(cfg, tmpDir, logger)

	// Initially no history, projection should have zero growth
	projection := dm.CalculateGrowthProjection()

	if projection.CalculatedAt.IsZero() {
		t.Error("CalculatedAt should not be zero")
	}

	// With no history, days until thresholds should be very high
	if projection.DaysUntilFull != 0 && projection.DaysUntilFull < 9999 && projection.DailyGrowthRate > 0 {
		t.Log("Growth rate calculated from empty history")
	}
}

func TestDiskMonitorGetHistory(t *testing.T) {
	tmpDir := t.TempDir()
	logger := log.NewNopLogger()

	cfg := DiskMonitorConfig{
		Enabled:       true,
		CheckInterval: 1 * time.Second,
	}

	dm := NewDiskMonitor(cfg, tmpDir, logger)

	// Initially empty
	history := dm.GetHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d entries", len(history))
	}
}

func TestAlertLevel(t *testing.T) {
	tests := []struct {
		level AlertLevel
		want  string
	}{
		{AlertLevelNone, "none"},
		{AlertLevelWarning, "warning"},
		{AlertLevelCritical, "critical"},
	}

	for _, tt := range tests {
		if string(tt.level) != tt.want {
			t.Errorf("AlertLevel = %q, want %q", tt.level, tt.want)
		}
	}
}

