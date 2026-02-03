// Package provider_daemon implements the provider daemon for VirtEngine.
//
// Tests for cost_tracker.go
package provider_daemon

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

const (
	testProviderID = "test-provider"
	testWorkloadID = "workload-123"
)

func TestDefaultResourceCostConfig(t *testing.T) {
	cfg := DefaultResourceCostConfig()

	if cfg.CPUCostPerCore <= 0 {
		t.Error("CPUCostPerCore should be positive")
	}
	if cfg.MemoryGBCostPerHour <= 0 {
		t.Error("MemoryGBCostPerHour should be positive")
	}
	if cfg.StorageGBCostPerHour <= 0 {
		t.Error("StorageGBCostPerHour should be positive")
	}
	if cfg.Currency == "" {
		t.Error("Currency should not be empty")
	}
}

func TestDefaultCostTrackerConfig(t *testing.T) {
	cfg := DefaultCostTrackerConfig()

	if !cfg.Enabled {
		t.Error("Default config should be enabled")
	}
	if cfg.CollectionInterval <= 0 {
		t.Error("CollectionInterval should be positive")
	}
	if cfg.RetentionPeriod <= 0 {
		t.Error("RetentionPeriod should be positive")
	}
	if len(cfg.Thresholds) == 0 {
		t.Error("Default config should have thresholds")
	}
}

func TestNewCostTracker(t *testing.T) {
	cfg := DefaultCostTrackerConfig()
	cfg.ProviderID = testProviderID

	tracker := NewCostTracker(cfg)

	if tracker == nil {
		t.Fatal("NewCostTracker should not return nil")
	}
	if tracker.cfg.ProviderID != testProviderID {
		t.Error("Provider ID not set correctly")
	}
}

func TestCostTrackerStartWorkloadTracking(t *testing.T) {
	cfg := DefaultCostTrackerConfig()
	cfg.ProviderID = testProviderID
	tracker := NewCostTracker(cfg)

	workloadID := testWorkloadID
	deploymentID := "deployment-456"
	leaseID := "lease-789"

	tracker.StartWorkloadTracking(workloadID, deploymentID, leaseID)

	cost, ok := tracker.GetWorkloadCost(workloadID)
	if !ok {
		t.Fatal("Workload should be tracked")
	}

	if cost.WorkloadID != workloadID {
		t.Errorf("Expected workload ID %s, got %s", workloadID, cost.WorkloadID)
	}
	if cost.DeploymentID != deploymentID {
		t.Errorf("Expected deployment ID %s, got %s", deploymentID, cost.DeploymentID)
	}
	if cost.LeaseID != leaseID {
		t.Errorf("Expected lease ID %s, got %s", leaseID, cost.LeaseID)
	}
	if cost.TotalCost != 0 {
		t.Error("Initial cost should be 0")
	}
}

func TestCostTrackerUpdateWorkloadCost(t *testing.T) {
	cfg := DefaultCostTrackerConfig()
	cfg.ProviderID = testProviderID
	tracker := NewCostTracker(cfg)

	workloadID := testWorkloadID
	tracker.StartWorkloadTracking(workloadID, "deployment-456", "lease-789")

	usage := ResourceMetrics{
		CPUMilliSeconds:    3600000,                        // 1 CPU-hour
		MemoryByteSeconds:  3600 * 1024 * 1024 * 1024,      // 1 GB-hour
		StorageByteSeconds: 3600 * 10 * 1024 * 1024 * 1024, // 10 GB-hour
		GPUSeconds:         3600,                           // 1 GPU-hour
		NetworkBytesIn:     1024 * 1024 * 1024,             // 1 GB
		NetworkBytesOut:    1024 * 1024 * 1024,             // 1 GB
	}

	err := tracker.UpdateWorkloadCost(workloadID, usage, time.Hour)
	if err != nil {
		t.Fatalf("UpdateWorkloadCost failed: %v", err)
	}

	cost, ok := tracker.GetWorkloadCost(workloadID)
	if !ok {
		t.Fatal("Workload should still be tracked")
	}

	if cost.TotalCost <= 0 {
		t.Error("Total cost should be positive after update")
	}
	if cost.CPUCost <= 0 {
		t.Error("CPU cost should be positive")
	}
	if cost.MemoryCost <= 0 {
		t.Error("Memory cost should be positive")
	}
	if cost.NetworkCost <= 0 {
		t.Error("Network cost should be positive")
	}
}

func TestCostTrackerUpdateWorkloadCostNotTracked(t *testing.T) {
	cfg := DefaultCostTrackerConfig()
	tracker := NewCostTracker(cfg)

	usage := ResourceMetrics{}
	err := tracker.UpdateWorkloadCost("non-existent", usage, time.Hour)

	if err == nil {
		t.Error("Should return error for non-existent workload")
	}
}

func TestCostTrackerStopWorkloadTracking(t *testing.T) {
	cfg := DefaultCostTrackerConfig()
	tracker := NewCostTracker(cfg)

	workloadID := testWorkloadID
	tracker.StartWorkloadTracking(workloadID, "deployment-456", "lease-789")

	cost := tracker.StopWorkloadTracking(workloadID)

	if cost == nil {
		t.Fatal("Should return cost on stop")
	}
	if cost.WorkloadID != workloadID {
		t.Error("Returned cost should have correct workload ID")
	}
	if cost.EndTime.IsZero() {
		t.Error("End time should be set")
	}

	// Verify workload is no longer tracked
	_, ok := tracker.GetWorkloadCost(workloadID)
	if ok {
		t.Error("Workload should not be tracked after stop")
	}
}

func TestCostTrackerStopWorkloadTrackingNotTracked(t *testing.T) {
	cfg := DefaultCostTrackerConfig()
	tracker := NewCostTracker(cfg)

	cost := tracker.StopWorkloadTracking("non-existent")
	if cost != nil {
		t.Error("Should return nil for non-existent workload")
	}
}

func TestCostTrackerGetAllWorkloadCosts(t *testing.T) {
	cfg := DefaultCostTrackerConfig()
	tracker := NewCostTracker(cfg)

	tracker.StartWorkloadTracking("workload-1", "deployment-1", "lease-1")
	tracker.StartWorkloadTracking("workload-2", "deployment-2", "lease-2")
	tracker.StartWorkloadTracking("workload-3", "deployment-3", "lease-3")

	costs := tracker.GetAllWorkloadCosts()

	if len(costs) != 3 {
		t.Errorf("Expected 3 workloads, got %d", len(costs))
	}
}

func TestCostTrackerGetTotalCost(t *testing.T) {
	cfg := DefaultCostTrackerConfig()
	tracker := NewCostTracker(cfg)

	tracker.StartWorkloadTracking("workload-1", "deployment-1", "lease-1")
	tracker.StartWorkloadTracking("workload-2", "deployment-2", "lease-2")

	usage := ResourceMetrics{
		CPUMilliSeconds: 3600000,
	}

	_ = tracker.UpdateWorkloadCost("workload-1", usage, time.Hour)
	_ = tracker.UpdateWorkloadCost("workload-2", usage, time.Hour)

	total := tracker.GetTotalCost()
	if total <= 0 {
		t.Error("Total cost should be positive")
	}
}

func TestCostTrackerGetCostSummary(t *testing.T) {
	cfg := DefaultCostTrackerConfig()
	cfg.ProviderID = testProviderID
	tracker := NewCostTracker(cfg)

	tracker.StartWorkloadTracking("workload-1", "deployment-1", "lease-1")

	usage := ResourceMetrics{
		CPUMilliSeconds:    3600000,
		MemoryByteSeconds:  3600 * 1024 * 1024 * 1024,
		StorageByteSeconds: 3600 * 1024 * 1024 * 1024,
		GPUSeconds:         3600,
		NetworkBytesIn:     1024 * 1024 * 1024,
		NetworkBytesOut:    1024 * 1024 * 1024,
	}

	_ = tracker.UpdateWorkloadCost("workload-1", usage, time.Hour)

	summary := tracker.GetCostSummary()

	if summary.WorkloadCount != 1 {
		t.Errorf("Expected 1 workload, got %d", summary.WorkloadCount)
	}
	if summary.TotalCost <= 0 {
		t.Error("Total cost should be positive")
	}
	if summary.Currency == "" {
		t.Error("Currency should be set")
	}
	if summary.TotalCPUCost <= 0 {
		t.Error("CPU cost should be positive")
	}
}

func TestCostTrackerStartStop(t *testing.T) {
	cfg := DefaultCostTrackerConfig()
	cfg.Enabled = true
	cfg.CollectionInterval = 100 * time.Millisecond
	tracker := NewCostTracker(cfg)

	ctx := context.Background()
	err := tracker.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Let it run briefly
	time.Sleep(50 * time.Millisecond)

	tracker.Stop()
	// Should be able to stop without hanging
}

func TestCostTrackerStartStopDisabled(t *testing.T) {
	cfg := DefaultCostTrackerConfig()
	cfg.Enabled = false
	tracker := NewCostTracker(cfg)

	ctx := context.Background()
	err := tracker.Start(ctx)
	if err != nil {
		t.Fatalf("Start should succeed even when disabled: %v", err)
	}

	tracker.Stop()
}

func TestCostTrackerGetMetrics(t *testing.T) {
	cfg := DefaultCostTrackerConfig()
	tracker := NewCostTracker(cfg)

	tracker.StartWorkloadTracking("workload-1", "deployment-1", "lease-1")

	metrics := tracker.GetMetrics()

	if metrics.ActiveWorkloads != 1 {
		t.Errorf("Expected 1 active workload, got %d", metrics.ActiveWorkloads)
	}
}

func TestCostTrackerExportCostReport(t *testing.T) {
	cfg := DefaultCostTrackerConfig()
	cfg.ProviderID = testProviderID
	tracker := NewCostTracker(cfg)

	tracker.StartWorkloadTracking("workload-1", "deployment-1", "lease-1")

	usage := ResourceMetrics{
		CPUMilliSeconds: 3600000,
	}
	_ = tracker.UpdateWorkloadCost("workload-1", usage, time.Hour)

	report, err := tracker.ExportCostReport()
	if err != nil {
		t.Fatalf("ExportCostReport failed: %v", err)
	}

	if len(report) == 0 {
		t.Error("Report should not be empty")
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(report, &parsed); err != nil {
		t.Errorf("Report should be valid JSON: %v", err)
	}

	if parsed["provider_id"] != testProviderID {
		t.Error("Report should contain provider ID")
	}
}

func TestCostTrackerConcurrency(t *testing.T) {
	cfg := DefaultCostTrackerConfig()
	tracker := NewCostTracker(cfg)

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Start multiple goroutines doing various operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				workloadID := "workload-" + string(rune('a'+id))
				tracker.StartWorkloadTracking(workloadID, "deployment", "lease")

				usage := ResourceMetrics{CPUMilliSeconds: 1000}
				_ = tracker.UpdateWorkloadCost(workloadID, usage, time.Minute)

				_, _ = tracker.GetWorkloadCost(workloadID)
				_ = tracker.GetTotalCost()
				_ = tracker.GetCostSummary()
				_ = tracker.GetAllWorkloadCosts()

				tracker.StopWorkloadTracking(workloadID)
			}
		}(i)
	}

	wg.Wait()
	// Test passes if no race conditions or panics occurred
}

func TestCostTrackerAlertHandler(t *testing.T) {
	cfg := DefaultCostTrackerConfig()
	cfg.Thresholds = []CostThreshold{
		{Name: "test", Amount: 0.01, Period: "daily"},
	}
	tracker := NewCostTracker(cfg)

	alertReceived := make(chan *CostAlert, 1)
	tracker.SetAlertHandler(func(ctx context.Context, alert *CostAlert) error {
		alertReceived <- alert
		return nil
	})

	// Create workload with cost exceeding threshold
	tracker.StartWorkloadTracking("workload-1", "deployment-1", "lease-1")
	usage := ResourceMetrics{
		CPUMilliSeconds: 3600000, // Should exceed $0.01 threshold
	}
	_ = tracker.UpdateWorkloadCost("workload-1", usage, time.Hour)

	// Check thresholds
	ctx := context.Background()
	tracker.CheckThresholds(ctx)

	// Wait for alert
	select {
	case alert := <-alertReceived:
		if alert.AlertType != CostAlertThresholdExceeded {
			t.Errorf("Expected threshold exceeded alert, got %s", alert.AlertType)
		}
	case <-time.After(time.Second):
		t.Error("Expected to receive alert")
	}
}

func TestCostTrackerGetAlerts(t *testing.T) {
	cfg := DefaultCostTrackerConfig()
	cfg.Thresholds = []CostThreshold{
		{Name: "test", Amount: 0.001, Period: "daily"},
	}
	tracker := NewCostTracker(cfg)

	tracker.StartWorkloadTracking("workload-1", "deployment-1", "lease-1")
	usage := ResourceMetrics{CPUMilliSeconds: 3600000}
	_ = tracker.UpdateWorkloadCost("workload-1", usage, time.Hour)

	ctx := context.Background()
	tracker.CheckThresholds(ctx)

	alerts := tracker.GetAlerts()
	if len(alerts) == 0 {
		t.Error("Expected at least one alert")
	}
}

func TestCostTrackerRightsizingRecommendations(t *testing.T) {
	cfg := DefaultCostTrackerConfig()
	tracker := NewCostTracker(cfg)

	tracker.StartWorkloadTracking("workload-1", "deployment-1", "lease-1")

	// Simulate usage that results in low utilization but high cost
	// Use actual UpdateWorkloadCost to set up state properly
	usage := ResourceMetrics{
		CPUMilliSeconds: 360000, // 0.1 CPU-hour worth
		GPUSeconds:      360,    // 0.1 GPU-hour worth
	}
	_ = tracker.UpdateWorkloadCost("workload-1", usage, time.Hour)

	// Manually adjust the start time to make duration long enough
	tracker.mu.Lock()
	if cost, ok := tracker.costs["workload-1"]; ok {
		cost.StartTime = time.Now().Add(-2 * time.Hour)
		cost.EndTime = time.Now()
		// Set high costs for low utilization scenario
		cost.CPUCost = 10.0
		cost.GPUCost = 50.0
	}
	tracker.mu.Unlock()

	ctx := context.Background()
	recommendations := tracker.GenerateRightsizingRecommendations(ctx)

	if len(recommendations) == 0 {
		t.Error("Expected rightsizing recommendations")
	}

	// Check that recommendations have required fields
	for _, rec := range recommendations {
		if rec.WorkloadID == "" {
			t.Error("Recommendation should have workload ID")
		}
		if rec.ResourceType == "" {
			t.Error("Recommendation should have resource type")
		}
		if rec.EstimatedSaving <= 0 {
			t.Error("Recommendation should have positive estimated saving")
		}
	}
}

func TestRightsizingRecommendation(t *testing.T) {
	rec := RightsizingRecommendation{
		WorkloadID:      "test-workload",
		ResourceType:    "CPU",
		CurrentUsage:    "0.05 core-hours/hour",
		Recommendation:  "Reduce CPU allocation",
		EstimatedSaving: 5.50,
	}

	if rec.WorkloadID != "test-workload" {
		t.Error("WorkloadID not set correctly")
	}
	if rec.EstimatedSaving != 5.50 {
		t.Error("EstimatedSaving not set correctly")
	}
}

func TestCostAlertTypes(t *testing.T) {
	types := []CostAlertType{
		CostAlertThresholdExceeded,
		CostAlertAnomalyDetected,
		CostAlertProjectionExceeded,
		CostAlertUnusedResources,
		CostAlertRightsizingRecommendation,
	}

	for _, at := range types {
		if at == "" {
			t.Error("Alert type should not be empty")
		}
	}
}

func TestWorkloadCostFields(t *testing.T) {
	cost := WorkloadCost{
		WorkloadID:   "wl-123",
		DeploymentID: "dp-456",
		LeaseID:      "ls-789",
		ProviderID:   "prov-abc",
		StartTime:    time.Now().Add(-time.Hour),
		EndTime:      time.Now(),
		CPUCost:      10.0,
		MemoryCost:   5.0,
		StorageCost:  2.0,
		GPUCost:      20.0,
		NetworkCost:  3.0,
		TotalCost:    40.0,
		Currency:     "USD",
	}

	if cost.TotalCost != 40.0 {
		t.Error("TotalCost not set correctly")
	}
	if cost.Currency != "USD" {
		t.Error("Currency not set correctly")
	}
}

func TestResourceUsageSummary(t *testing.T) {
	usage := ResourceUsageSummary{
		CPUCoreHours:   10.5,
		MemoryGBHours:  20.0,
		StorageGBHours: 100.0,
		GPUHours:       5.0,
		NetworkInGB:    50.0,
		NetworkOutGB:   30.0,
	}

	if usage.CPUCoreHours != 10.5 {
		t.Error("CPUCoreHours not set correctly")
	}
	if usage.NetworkInGB+usage.NetworkOutGB != 80.0 {
		t.Error("Network usage calculation incorrect")
	}
}

func TestCostThreshold(t *testing.T) {
	threshold := CostThreshold{
		Name:         "monthly_budget",
		Amount:       1000.0,
		Period:       "monthly",
		NotifyEmails: []string{"admin@example.com"},
	}

	if threshold.Amount != 1000.0 {
		t.Error("Amount not set correctly")
	}
	if len(threshold.NotifyEmails) != 1 {
		t.Error("NotifyEmails not set correctly")
	}
}
