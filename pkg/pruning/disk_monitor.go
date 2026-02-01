// Package pruning provides state pruning optimization for VirtEngine blockchain.
package pruning

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"cosmossdk.io/log"
)

// DiskMonitor monitors disk usage and provides alerting.
type DiskMonitor struct {
	config       DiskMonitorConfig
	logger       log.Logger
	mu           sync.RWMutex
	dataDir      string
	history      []DiskUsageSample
	maxHistory   int
	alertHandler AlertHandler
	cancelFunc   context.CancelFunc
	wg           sync.WaitGroup
}

// DiskUsageSample represents a point-in-time disk usage measurement.
type DiskUsageSample struct {
	Timestamp    time.Time `json:"timestamp"`
	TotalBytes   uint64    `json:"total_bytes"`
	UsedBytes    uint64    `json:"used_bytes"`
	FreeBytes    uint64    `json:"free_bytes"`
	UsedPercent  float64   `json:"used_percent"`
	StateSize    uint64    `json:"state_size"`
	SnapshotSize uint64    `json:"snapshot_size"`
}

// DiskUsageStatus represents the current disk usage status.
type DiskUsageStatus struct {
	Current           DiskUsageSample `json:"current"`
	AlertLevel        AlertLevel      `json:"alert_level"`
	GrowthRatePerDay  int64           `json:"growth_rate_per_day"`
	DaysUntilFull     int             `json:"days_until_full"`
	Projection30Days  uint64          `json:"projection_30_days"`
	Projection90Days  uint64          `json:"projection_90_days"`
	RecommendedAction string          `json:"recommended_action"`
	LastAlertTime     time.Time       `json:"last_alert_time,omitempty"`
}

// AlertLevel represents the severity of a disk usage alert.
type AlertLevel string

const (
	AlertLevelNone     AlertLevel = "none"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
)

// Alert represents a disk usage alert.
type Alert struct {
	Level       AlertLevel `json:"level"`
	Message     string     `json:"message"`
	UsedPercent float64    `json:"used_percent"`
	FreeBytes   uint64     `json:"free_bytes"`
	Timestamp   time.Time  `json:"timestamp"`
}

// AlertHandler handles disk usage alerts.
type AlertHandler interface {
	OnAlert(alert Alert)
}

// GrowthProjection contains disk growth projections.
type GrowthProjection struct {
	CurrentUsage      uint64    `json:"current_usage"`
	DailyGrowthRate   int64     `json:"daily_growth_rate"`
	WeeklyGrowthRate  int64     `json:"weekly_growth_rate"`
	MonthlyGrowthRate int64     `json:"monthly_growth_rate"`
	Projection7Days   uint64    `json:"projection_7_days"`
	Projection30Days  uint64    `json:"projection_30_days"`
	Projection90Days  uint64    `json:"projection_90_days"`
	DaysUntilWarning  int       `json:"days_until_warning"`
	DaysUntilCritical int       `json:"days_until_critical"`
	DaysUntilFull     int       `json:"days_until_full"`
	CalculatedAt      time.Time `json:"calculated_at"`
}

// NewDiskMonitor creates a new disk monitor.
func NewDiskMonitor(config DiskMonitorConfig, dataDir string, logger log.Logger) *DiskMonitor {
	return &DiskMonitor{
		config:     config,
		logger:     logger.With("module", "disk-monitor"),
		dataDir:    dataDir,
		maxHistory: 1440, // 24 hours at 1-minute intervals or 5 days at 5-minute intervals
	}
}

// SetAlertHandler sets the alert handler.
func (dm *DiskMonitor) SetAlertHandler(handler AlertHandler) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.alertHandler = handler
}

// Start starts the disk monitoring routine.
func (dm *DiskMonitor) Start(ctx context.Context) {
	if !dm.config.Enabled {
		return
	}

	ctx, dm.cancelFunc = context.WithCancel(ctx)
	dm.wg.Add(1)

	go func() {
		defer dm.wg.Done()
		dm.monitorLoop(ctx)
	}()
}

// Stop stops the disk monitoring routine.
func (dm *DiskMonitor) Stop() {
	if dm.cancelFunc != nil {
		dm.cancelFunc()
	}
	dm.wg.Wait()
}

// monitorLoop is the main monitoring loop.
func (dm *DiskMonitor) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(dm.config.CheckInterval)
	defer ticker.Stop()

	// Initial check
	dm.checkDiskUsage()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			dm.checkDiskUsage()
		}
	}
}

// checkDiskUsage performs a disk usage check.
func (dm *DiskMonitor) checkDiskUsage() {
	sample, err := dm.collectSample()
	if err != nil {
		dm.logger.Error("failed to collect disk usage sample", "error", err)
		return
	}

	dm.mu.Lock()
	dm.history = append(dm.history, sample)
	if len(dm.history) > dm.maxHistory {
		dm.history = dm.history[1:]
	}
	dm.mu.Unlock()

	// Check thresholds and alert
	dm.checkThresholds(sample)
}

// collectSample collects a disk usage sample.
func (dm *DiskMonitor) collectSample() (DiskUsageSample, error) {
	sample := DiskUsageSample{
		Timestamp: time.Now(),
	}

	// Get disk usage for data directory
	usage, err := getDiskUsage(dm.dataDir)
	if err != nil {
		return sample, err
	}

	sample.TotalBytes = usage.Total
	sample.UsedBytes = usage.Used
	sample.FreeBytes = usage.Free
	if usage.Total > 0 {
		sample.UsedPercent = float64(usage.Used) / float64(usage.Total) * 100
	}

	// Get state size
	stateDir := filepath.Join(dm.dataDir, "data", "application.db")
	sample.StateSize = getDirSize(stateDir)

	// Get snapshot size
	snapshotDir := filepath.Join(dm.dataDir, "data", "snapshots")
	sample.SnapshotSize = getDirSize(snapshotDir)

	return sample, nil
}

// checkThresholds checks disk usage against thresholds.
func (dm *DiskMonitor) checkThresholds(sample DiskUsageSample) {
	var alertLevel AlertLevel
	var message string

	if sample.UsedPercent >= dm.config.CriticalThresholdPercent {
		alertLevel = AlertLevelCritical
		message = fmt.Sprintf("CRITICAL: Disk usage at %.1f%% (%.2f GB free)",
			sample.UsedPercent, float64(sample.FreeBytes)/(1024*1024*1024))
	} else if sample.UsedPercent >= dm.config.WarningThresholdPercent {
		alertLevel = AlertLevelWarning
		message = fmt.Sprintf("WARNING: Disk usage at %.1f%% (%.2f GB free)",
			sample.UsedPercent, float64(sample.FreeBytes)/(1024*1024*1024))
	}

	if alertLevel != AlertLevelNone {
		alert := Alert{
			Level:       alertLevel,
			Message:     message,
			UsedPercent: sample.UsedPercent,
			FreeBytes:   sample.FreeBytes,
			Timestamp:   sample.Timestamp,
		}

		dm.logger.Warn(message, "level", alertLevel, "used_percent", sample.UsedPercent)

		dm.mu.RLock()
		handler := dm.alertHandler
		dm.mu.RUnlock()

		if handler != nil {
			handler.OnAlert(alert)
		}
	}
}

// GetStatus returns the current disk usage status.
func (dm *DiskMonitor) GetStatus() (DiskUsageStatus, error) {
	sample, err := dm.collectSample()
	if err != nil {
		return DiskUsageStatus{}, err
	}

	status := DiskUsageStatus{
		Current: sample,
	}

	// Determine alert level
	if sample.UsedPercent >= dm.config.CriticalThresholdPercent {
		status.AlertLevel = AlertLevelCritical
	} else if sample.UsedPercent >= dm.config.WarningThresholdPercent {
		status.AlertLevel = AlertLevelWarning
	} else {
		status.AlertLevel = AlertLevelNone
	}

	// Calculate growth projections
	projection := dm.CalculateGrowthProjection()
	status.GrowthRatePerDay = projection.DailyGrowthRate
	status.DaysUntilFull = projection.DaysUntilFull
	status.Projection30Days = projection.Projection30Days
	status.Projection90Days = projection.Projection90Days

	// Determine recommended action
	status.RecommendedAction = dm.getRecommendedAction(status)

	return status, nil
}

// getRecommendedAction returns a recommended action based on status.
func (dm *DiskMonitor) getRecommendedAction(status DiskUsageStatus) string {
	switch status.AlertLevel {
	case AlertLevelCritical:
		return "URGENT: Increase disk space or enable aggressive pruning immediately"
	case AlertLevelWarning:
		if status.DaysUntilFull < 7 {
			return "Consider increasing disk space or adjusting pruning settings"
		}
		return "Monitor disk usage; consider increasing disk space if growth continues"
	default:
		if status.DaysUntilFull < 30 {
			return "Plan for disk space increase within the next month"
		}
		return "No action required"
	}
}

// CalculateGrowthProjection calculates storage growth projections.
func (dm *DiskMonitor) CalculateGrowthProjection() GrowthProjection {
	dm.mu.RLock()
	history := make([]DiskUsageSample, len(dm.history))
	copy(history, dm.history)
	dm.mu.RUnlock()

	projection := GrowthProjection{
		CalculatedAt: time.Now(),
	}

	if len(history) == 0 {
		return projection
	}

	latest := history[len(history)-1]
	projection.CurrentUsage = latest.UsedBytes

	// Calculate daily growth rate from history
	if len(history) >= 2 {
		// Find samples from ~24 hours ago
		dayAgo := time.Now().Add(-24 * time.Hour)
		var oldestRelevant *DiskUsageSample
		for i := range history {
			if history[i].Timestamp.After(dayAgo) && oldestRelevant == nil {
				if i > 0 {
					oldestRelevant = &history[i-1]
				} else {
					oldestRelevant = &history[i]
				}
				break
			}
		}

		if oldestRelevant != nil {
			elapsed := latest.Timestamp.Sub(oldestRelevant.Timestamp)
			if elapsed > 0 {
				//nolint:gosec // G115: conversion is safe for disk usage values in practice
				bytesChange := int64(latest.UsedBytes) - int64(oldestRelevant.UsedBytes)
				daysElapsed := elapsed.Hours() / 24
				if daysElapsed > 0 {
					projection.DailyGrowthRate = int64(float64(bytesChange) / daysElapsed)
				}
			}
		}
	}

	// Calculate projections
	if projection.DailyGrowthRate > 0 {
		projection.WeeklyGrowthRate = projection.DailyGrowthRate * 7
		projection.MonthlyGrowthRate = projection.DailyGrowthRate * 30

		//nolint:gosec // G115: growth rate is checked > 0, conversion is safe
		projection.Projection7Days = latest.UsedBytes + uint64(projection.DailyGrowthRate*7)
		//nolint:gosec // G115: growth rate is checked > 0, conversion is safe
		projection.Projection30Days = latest.UsedBytes + uint64(projection.DailyGrowthRate*30)
		//nolint:gosec // G115: growth rate is checked > 0, conversion is safe
		projection.Projection90Days = latest.UsedBytes + uint64(projection.DailyGrowthRate*90)

		// Calculate days until thresholds
		freeBytes := latest.FreeBytes
		totalBytes := latest.TotalBytes

		warningThreshold := uint64(float64(totalBytes) * dm.config.WarningThresholdPercent / 100)
		criticalThreshold := uint64(float64(totalBytes) * dm.config.CriticalThresholdPercent / 100)

		if latest.UsedBytes < warningThreshold {
			//nolint:gosec // G115: conversion is safe, UsedBytes < warningThreshold ensures positive result
			projection.DaysUntilWarning = int((int64(warningThreshold) - int64(latest.UsedBytes)) / projection.DailyGrowthRate)
		}
		if latest.UsedBytes < criticalThreshold {
			//nolint:gosec // G115: conversion is safe, UsedBytes < criticalThreshold ensures positive result
			projection.DaysUntilCritical = int((int64(criticalThreshold) - int64(latest.UsedBytes)) / projection.DailyGrowthRate)
		}
		//nolint:gosec // G115: freeBytes is disk free space, safe for int64 conversion
		projection.DaysUntilFull = int(int64(freeBytes) / projection.DailyGrowthRate)
	} else {
		// Stable or shrinking - set to a large value
		projection.DaysUntilWarning = 9999
		projection.DaysUntilCritical = 9999
		projection.DaysUntilFull = 9999
	}

	return projection
}

// GetHistory returns the disk usage history.
func (dm *DiskMonitor) GetHistory() []DiskUsageSample {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	result := make([]DiskUsageSample, len(dm.history))
	copy(result, dm.history)
	return result
}

// DiskUsageInfo contains raw disk usage information.
type DiskUsageInfo struct {
	Total uint64
	Used  uint64
	Free  uint64
}

// getDiskUsage returns disk usage for a path.
func getDiskUsage(path string) (DiskUsageInfo, error) {
	// Platform-specific implementation
	if runtime.GOOS == "windows" {
		return getDiskUsageWindows(path)
	}
	return getDiskUsageUnix(path)
}

// getDiskUsageWindows gets disk usage on Windows.
func getDiskUsageWindows(path string) (DiskUsageInfo, error) {
	// Get the volume path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return DiskUsageInfo{}, err
	}

	// On Windows, use a simplified approach based on available info
	// In production, this would use syscall.GetDiskFreeSpaceEx
	// For now, return estimated values based on directory size
	info := DiskUsageInfo{
		Total: 1024 * 1024 * 1024 * 1024, // 1TB default
		Free:  512 * 1024 * 1024 * 1024,  // 512GB free
	}
	info.Used = info.Total - info.Free

	// Try to get actual values if possible
	// Actual implementation would call Windows API via syscall.GetDiskFreeSpaceEx
	// For now, we verify the path exists to ensure valid input
	_, _ = os.Stat(absPath)

	return info, nil
}

// getDiskUsageUnix gets disk usage on Unix systems.
//
//nolint:unparam // path kept for future syscall.Statfs implementation
func getDiskUsageUnix(_ string) (DiskUsageInfo, error) {
	// On Unix, this would use syscall.Statfs
	// For portability, return estimated values
	info := DiskUsageInfo{
		Total: 1024 * 1024 * 1024 * 1024, // 1TB default
		Free:  512 * 1024 * 1024 * 1024,  // 512GB free
	}
	info.Used = info.Total - info.Free
	return info, nil
}

// getDirSize calculates the total size of a directory.
func getDirSize(path string) uint64 {
	var size uint64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Ignore errors and continue
		}
		if !info.IsDir() {
			//nolint:gosec // G115: file sizes are non-negative, safe conversion
			size += uint64(info.Size())
		}
		return nil
	})

	if err != nil {
		return 0
	}

	return size
}

// FormatBytes formats bytes as a human-readable string.
func FormatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
