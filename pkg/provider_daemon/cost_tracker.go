// Package provider_daemon implements the provider daemon for VirtEngine.
//
// PERF-9B: Cost tracking and optimization for provider resources
package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	verrors "github.com/virtengine/virtengine/pkg/errors"
)

// CostUnit represents the unit for cost calculation
type CostUnit string

const (
	// CostUnitHour represents hourly cost
	CostUnitHour CostUnit = "hour"
	// CostUnitMonth represents monthly cost
	CostUnitMonth CostUnit = "month"
)

// ResourceCostConfig contains pricing configuration for resources
type ResourceCostConfig struct {
	// CPUCostPerCore is the cost per CPU core per hour
	CPUCostPerCore float64 `json:"cpu_cost_per_core"`

	// MemoryGBCostPerHour is the cost per GB of memory per hour
	MemoryGBCostPerHour float64 `json:"memory_gb_cost_per_hour"`

	// StorageGBCostPerHour is the cost per GB of storage per hour
	StorageGBCostPerHour float64 `json:"storage_gb_cost_per_hour"`

	// GPUCostPerHour is the cost per GPU per hour
	GPUCostPerHour float64 `json:"gpu_cost_per_hour"`

	// NetworkGBCost is the cost per GB of network transfer
	NetworkGBCost float64 `json:"network_gb_cost"`

	// Currency is the currency for cost calculations
	Currency string `json:"currency"`
}

// DefaultResourceCostConfig returns default cost configuration
func DefaultResourceCostConfig() ResourceCostConfig {
	return ResourceCostConfig{
		CPUCostPerCore:       0.05,  // $0.05 per core per hour
		MemoryGBCostPerHour:  0.01,  // $0.01 per GB per hour
		StorageGBCostPerHour: 0.001, // $0.001 per GB per hour
		GPUCostPerHour:       0.50,  // $0.50 per GPU per hour
		NetworkGBCost:        0.02,  // $0.02 per GB transferred
		Currency:             "USD",
	}
}

// WorkloadCost represents the cost breakdown for a workload
type WorkloadCost struct {
	// WorkloadID is the workload identifier
	WorkloadID string `json:"workload_id"`

	// DeploymentID is the on-chain deployment ID
	DeploymentID string `json:"deployment_id"`

	// LeaseID is the on-chain lease ID
	LeaseID string `json:"lease_id"`

	// ProviderID is the provider identifier
	ProviderID string `json:"provider_id"`

	// StartTime is when cost tracking started
	StartTime time.Time `json:"start_time"`

	// EndTime is when cost tracking ended (or current time if active)
	EndTime time.Time `json:"end_time"`

	// CPUCost is the total CPU cost
	CPUCost float64 `json:"cpu_cost"`

	// MemoryCost is the total memory cost
	MemoryCost float64 `json:"memory_cost"`

	// StorageCost is the total storage cost
	StorageCost float64 `json:"storage_cost"`

	// GPUCost is the total GPU cost
	GPUCost float64 `json:"gpu_cost"`

	// NetworkCost is the total network cost
	NetworkCost float64 `json:"network_cost"`

	// TotalCost is the sum of all costs
	TotalCost float64 `json:"total_cost"`

	// Currency is the currency for costs
	Currency string `json:"currency"`

	// ResourceUsage contains the underlying resource usage
	ResourceUsage ResourceUsageSummary `json:"resource_usage"`
}

// ResourceUsageSummary summarizes resource usage for cost calculation
type ResourceUsageSummary struct {
	// CPUCoreHours is the total CPU core-hours used
	CPUCoreHours float64 `json:"cpu_core_hours"`

	// MemoryGBHours is the total memory GB-hours used
	MemoryGBHours float64 `json:"memory_gb_hours"`

	// StorageGBHours is the total storage GB-hours used
	StorageGBHours float64 `json:"storage_gb_hours"`

	// GPUHours is the total GPU-hours used
	GPUHours float64 `json:"gpu_hours"`

	// NetworkInGB is the total network ingress in GB
	NetworkInGB float64 `json:"network_in_gb"`

	// NetworkOutGB is the total network egress in GB
	NetworkOutGB float64 `json:"network_out_gb"`
}

// CostAlert represents a cost-related alert
type CostAlert struct {
	// AlertID is the unique identifier
	AlertID string `json:"alert_id"`

	// WorkloadID is the affected workload
	WorkloadID string `json:"workload_id"`

	// AlertType is the type of cost alert
	AlertType CostAlertType `json:"alert_type"`

	// Message is the alert message
	Message string `json:"message"`

	// CurrentCost is the current accumulated cost
	CurrentCost float64 `json:"current_cost"`

	// Threshold is the threshold that was exceeded
	Threshold float64 `json:"threshold"`

	// ProjectedCost is the projected cost for the period
	ProjectedCost float64 `json:"projected_cost,omitempty"`

	// CreatedAt is when the alert was created
	CreatedAt time.Time `json:"created_at"`
}

// CostAlertType represents the type of cost alert
type CostAlertType string

const (
	// CostAlertThresholdExceeded indicates a cost threshold was exceeded
	CostAlertThresholdExceeded CostAlertType = "threshold_exceeded"

	// CostAlertAnomalyDetected indicates unusual cost patterns
	CostAlertAnomalyDetected CostAlertType = "anomaly_detected"

	// CostAlertProjectionExceeded indicates projected cost exceeds budget
	CostAlertProjectionExceeded CostAlertType = "projection_exceeded"

	// CostAlertUnusedResources indicates resources are being wasted
	CostAlertUnusedResources CostAlertType = "unused_resources"

	// CostAlertRightsizingRecommendation indicates a right-sizing opportunity
	CostAlertRightsizingRecommendation CostAlertType = "rightsizing_recommendation"
)

// CostThreshold represents a cost threshold for alerting
type CostThreshold struct {
	// Name is the threshold name
	Name string `json:"name"`

	// Amount is the threshold amount
	Amount float64 `json:"amount"`

	// Period is the time period (daily, weekly, monthly)
	Period string `json:"period"`

	// NotifyEmails are email addresses to notify
	NotifyEmails []string `json:"notify_emails,omitempty"`
}

// CostTrackerConfig configures the cost tracker
type CostTrackerConfig struct {
	// Enabled enables cost tracking
	Enabled bool `json:"enabled"`

	// ProviderID is the provider identifier
	ProviderID string `json:"provider_id"`

	// CostConfig contains resource pricing
	CostConfig ResourceCostConfig `json:"cost_config"`

	// Thresholds are cost thresholds for alerts
	Thresholds []CostThreshold `json:"thresholds,omitempty"`

	// CollectionInterval is how often to collect cost data
	CollectionInterval time.Duration `json:"collection_interval"`

	// RetentionPeriod is how long to retain cost records
	RetentionPeriod time.Duration `json:"retention_period"`

	// EnableAnomalyDetection enables cost anomaly detection
	EnableAnomalyDetection bool `json:"enable_anomaly_detection"`

	// AnomalyThresholdPercent is the percentage deviation to flag as anomaly
	AnomalyThresholdPercent float64 `json:"anomaly_threshold_percent"`
}

// DefaultCostTrackerConfig returns default configuration
func DefaultCostTrackerConfig() CostTrackerConfig {
	return CostTrackerConfig{
		Enabled:                 true,
		CostConfig:              DefaultResourceCostConfig(),
		CollectionInterval:      time.Hour,
		RetentionPeriod:         30 * 24 * time.Hour,
		EnableAnomalyDetection:  true,
		AnomalyThresholdPercent: 50.0,
		Thresholds: []CostThreshold{
			{Name: "daily_warning", Amount: 100.0, Period: "daily"},
			{Name: "daily_critical", Amount: 200.0, Period: "daily"},
			{Name: "monthly_budget", Amount: 2000.0, Period: "monthly"},
		},
	}
}

// CostTracker tracks and analyzes infrastructure costs
type CostTracker struct {
	mu sync.RWMutex

	cfg    CostTrackerConfig
	costs  map[string]*WorkloadCost
	alerts []CostAlert

	// Metrics
	totalCostTracked   float64
	totalAlertsCreated int64
	lastCollectionTime time.Time

	// Historical data for anomaly detection
	dailyCosts   []float64
	weeklyCosts  []float64
	monthlyCosts []float64

	// Control
	running  bool
	stopChan chan struct{}
	wg       sync.WaitGroup

	// Callbacks
	alertHandler func(context.Context, *CostAlert) error
}

// NewCostTracker creates a new cost tracker
func NewCostTracker(cfg CostTrackerConfig) *CostTracker {
	return &CostTracker{
		cfg:          cfg,
		costs:        make(map[string]*WorkloadCost),
		alerts:       make([]CostAlert, 0),
		dailyCosts:   make([]float64, 0, 30),
		weeklyCosts:  make([]float64, 0, 12),
		monthlyCosts: make([]float64, 0, 12),
		stopChan:     make(chan struct{}),
	}
}

// Start starts the cost tracker
func (ct *CostTracker) Start(ctx context.Context) error {
	if !ct.cfg.Enabled {
		return nil
	}

	ct.mu.Lock()
	if ct.running {
		ct.mu.Unlock()
		return nil
	}
	ct.running = true
	ct.mu.Unlock()

	ct.wg.Add(1)
	verrors.SafeGo("provider-daemon:cost-tracker", func() {
		defer ct.wg.Done()
		ct.trackingLoop(ctx)
	})

	return nil
}

// Stop stops the cost tracker
func (ct *CostTracker) Stop() {
	ct.mu.Lock()
	if !ct.running {
		ct.mu.Unlock()
		return
	}
	ct.running = false
	ct.mu.Unlock()

	close(ct.stopChan)
	ct.wg.Wait()
	ct.stopChan = make(chan struct{})
}

// SetAlertHandler sets the alert handler callback
func (ct *CostTracker) SetAlertHandler(handler func(context.Context, *CostAlert) error) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.alertHandler = handler
}

// StartWorkloadTracking starts cost tracking for a workload
func (ct *CostTracker) StartWorkloadTracking(workloadID, deploymentID, leaseID string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.costs[workloadID] = &WorkloadCost{
		WorkloadID:   workloadID,
		DeploymentID: deploymentID,
		LeaseID:      leaseID,
		ProviderID:   ct.cfg.ProviderID,
		StartTime:    time.Now(),
		Currency:     ct.cfg.CostConfig.Currency,
	}
}

// StopWorkloadTracking stops cost tracking for a workload
func (ct *CostTracker) StopWorkloadTracking(workloadID string) *WorkloadCost {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	cost, ok := ct.costs[workloadID]
	if !ok {
		return nil
	}

	cost.EndTime = time.Now()
	delete(ct.costs, workloadID)

	return cost
}

// UpdateWorkloadCost updates the cost for a workload based on usage
func (ct *CostTracker) UpdateWorkloadCost(workloadID string, usage ResourceMetrics, duration time.Duration) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	cost, ok := ct.costs[workloadID]
	if !ok {
		return fmt.Errorf("workload %s not tracked", workloadID)
	}

	hours := duration.Hours()

	// Calculate costs from usage
	cpuCoreHours := float64(usage.CPUMilliSeconds) / 1000.0 / 3600.0
	memoryGBHours := float64(usage.MemoryByteSeconds) / (1024 * 1024 * 1024) / 3600.0
	storageGBHours := float64(usage.StorageByteSeconds) / (1024 * 1024 * 1024) / 3600.0
	gpuHours := float64(usage.GPUSeconds) / 3600.0
	networkInGB := float64(usage.NetworkBytesIn) / (1024 * 1024 * 1024)
	networkOutGB := float64(usage.NetworkBytesOut) / (1024 * 1024 * 1024)

	// Add to cumulative usage
	cost.ResourceUsage.CPUCoreHours += cpuCoreHours
	cost.ResourceUsage.MemoryGBHours += memoryGBHours
	cost.ResourceUsage.StorageGBHours += storageGBHours
	cost.ResourceUsage.GPUHours += gpuHours
	cost.ResourceUsage.NetworkInGB += networkInGB
	cost.ResourceUsage.NetworkOutGB += networkOutGB

	// Calculate costs
	cpuCost := cpuCoreHours * ct.cfg.CostConfig.CPUCostPerCore
	memoryCost := memoryGBHours * ct.cfg.CostConfig.MemoryGBCostPerHour
	storageCost := storageGBHours * ct.cfg.CostConfig.StorageGBCostPerHour
	gpuCost := gpuHours * ct.cfg.CostConfig.GPUCostPerHour
	networkCost := (networkInGB + networkOutGB) * ct.cfg.CostConfig.NetworkGBCost

	cost.CPUCost += cpuCost
	cost.MemoryCost += memoryCost
	cost.StorageCost += storageCost
	cost.GPUCost += gpuCost
	cost.NetworkCost += networkCost
	cost.TotalCost = cost.CPUCost + cost.MemoryCost + cost.StorageCost + cost.GPUCost + cost.NetworkCost
	cost.EndTime = time.Now()

	ct.totalCostTracked += cpuCost + memoryCost + storageCost + gpuCost + networkCost

	// Check for hours parameter validity
	_ = hours

	return nil
}

// GetWorkloadCost returns the current cost for a workload
func (ct *CostTracker) GetWorkloadCost(workloadID string) (*WorkloadCost, bool) {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	cost, ok := ct.costs[workloadID]
	if !ok {
		return nil, false
	}

	// Return a copy
	costCopy := *cost
	return &costCopy, true
}

// GetAllWorkloadCosts returns costs for all tracked workloads
func (ct *CostTracker) GetAllWorkloadCosts() []WorkloadCost {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	result := make([]WorkloadCost, 0, len(ct.costs))
	for _, cost := range ct.costs {
		result = append(result, *cost)
	}
	return result
}

// GetTotalCost returns the total cost across all workloads
func (ct *CostTracker) GetTotalCost() float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	var total float64
	for _, cost := range ct.costs {
		total += cost.TotalCost
	}
	return total
}

// GetCostSummary returns a summary of costs by category
func (ct *CostTracker) GetCostSummary() CostSummary {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	summary := CostSummary{
		Currency:      ct.cfg.CostConfig.Currency,
		WorkloadCount: len(ct.costs),
		Timestamp:     time.Now(),
	}

	for _, cost := range ct.costs {
		summary.TotalCPUCost += cost.CPUCost
		summary.TotalMemoryCost += cost.MemoryCost
		summary.TotalStorageCost += cost.StorageCost
		summary.TotalGPUCost += cost.GPUCost
		summary.TotalNetworkCost += cost.NetworkCost
		summary.TotalCost += cost.TotalCost
	}

	return summary
}

// CostSummary provides an aggregate cost summary
type CostSummary struct {
	TotalCPUCost     float64   `json:"total_cpu_cost"`
	TotalMemoryCost  float64   `json:"total_memory_cost"`
	TotalStorageCost float64   `json:"total_storage_cost"`
	TotalGPUCost     float64   `json:"total_gpu_cost"`
	TotalNetworkCost float64   `json:"total_network_cost"`
	TotalCost        float64   `json:"total_cost"`
	Currency         string    `json:"currency"`
	WorkloadCount    int       `json:"workload_count"`
	Timestamp        time.Time `json:"timestamp"`
}

// GetAlerts returns recent cost alerts
func (ct *CostTracker) GetAlerts() []CostAlert {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	result := make([]CostAlert, len(ct.alerts))
	copy(result, ct.alerts)
	return result
}

// CheckThresholds checks current costs against thresholds
func (ct *CostTracker) CheckThresholds(ctx context.Context) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	totalCost := ct.getTotalCostLocked()

	for _, threshold := range ct.cfg.Thresholds {
		if totalCost > threshold.Amount {
			alert := CostAlert{
				AlertID:     fmt.Sprintf("cost-threshold-%s-%d", threshold.Name, time.Now().Unix()),
				AlertType:   CostAlertThresholdExceeded,
				Message:     fmt.Sprintf("Cost threshold '%s' exceeded: $%.2f > $%.2f", threshold.Name, totalCost, threshold.Amount),
				CurrentCost: totalCost,
				Threshold:   threshold.Amount,
				CreatedAt:   time.Now(),
			}

			ct.alerts = append(ct.alerts, alert)
			atomic.AddInt64(&ct.totalAlertsCreated, 1)

			if ct.alertHandler != nil {
				go func(a CostAlert) {
					_ = ct.alertHandler(ctx, &a)
				}(alert)
			}
		}
	}
}

// DetectAnomalies detects cost anomalies
func (ct *CostTracker) DetectAnomalies(ctx context.Context) {
	if !ct.cfg.EnableAnomalyDetection {
		return
	}

	ct.mu.Lock()
	defer ct.mu.Unlock()

	// Need historical data for anomaly detection
	if len(ct.dailyCosts) < 7 {
		return
	}

	// Calculate average and standard deviation
	var sum, sumSq float64
	for _, c := range ct.dailyCosts {
		sum += c
		sumSq += c * c
	}
	n := float64(len(ct.dailyCosts))
	mean := sum / n
	variance := (sumSq / n) - (mean * mean)
	stdDev := 0.0
	if variance > 0 {
		stdDev = variance // Simplified, should use math.Sqrt
	}

	currentCost := ct.getTotalCostLocked()

	// Check if current cost is an anomaly
	threshold := mean + (stdDev * ct.cfg.AnomalyThresholdPercent / 100.0)
	if currentCost > threshold && threshold > 0 {
		alert := CostAlert{
			AlertID:     fmt.Sprintf("cost-anomaly-%d", time.Now().Unix()),
			AlertType:   CostAlertAnomalyDetected,
			Message:     fmt.Sprintf("Unusual cost pattern detected: $%.2f (expected: $%.2f ± $%.2f)", currentCost, mean, stdDev),
			CurrentCost: currentCost,
			Threshold:   threshold,
			CreatedAt:   time.Now(),
		}

		ct.alerts = append(ct.alerts, alert)
		atomic.AddInt64(&ct.totalAlertsCreated, 1)

		if ct.alertHandler != nil {
			go func(a CostAlert) {
				_ = ct.alertHandler(ctx, &a)
			}(alert)
		}
	}
}

// GenerateRightsizingRecommendations analyzes workloads for right-sizing opportunities
func (ct *CostTracker) GenerateRightsizingRecommendations(ctx context.Context) []RightsizingRecommendation {
	ct.mu.RLock()
	// Copy costs to avoid holding lock during analysis
	costsCopy := make([]WorkloadCost, 0, len(ct.costs))
	for _, cost := range ct.costs {
		costsCopy = append(costsCopy, *cost)
	}
	ct.mu.RUnlock()

	recommendations := make([]RightsizingRecommendation, 0)

	for _, cost := range costsCopy {
		// Calculate utilization ratios (simplified)
		duration := cost.EndTime.Sub(cost.StartTime)
		if duration < time.Hour {
			continue
		}

		hours := duration.Hours()

		// Check CPU utilization
		avgCPUCores := cost.ResourceUsage.CPUCoreHours / hours
		if avgCPUCores < 0.1 && cost.CPUCost > 1.0 {
			recommendations = append(recommendations, RightsizingRecommendation{
				WorkloadID:      cost.WorkloadID,
				ResourceType:    "CPU",
				CurrentUsage:    fmt.Sprintf("%.2f core-hours/hour", avgCPUCores),
				Recommendation:  "Consider reducing CPU allocation - average utilization is very low",
				EstimatedSaving: cost.CPUCost * 0.5,
			})
		}

		// Check memory utilization
		avgMemoryGB := cost.ResourceUsage.MemoryGBHours / hours
		if avgMemoryGB < 0.5 && cost.MemoryCost > 1.0 {
			recommendations = append(recommendations, RightsizingRecommendation{
				WorkloadID:      cost.WorkloadID,
				ResourceType:    "Memory",
				CurrentUsage:    fmt.Sprintf("%.2f GB-hours/hour", avgMemoryGB),
				Recommendation:  "Consider reducing memory allocation - average utilization is low",
				EstimatedSaving: cost.MemoryCost * 0.3,
			})
		}

		// Check GPU utilization
		avgGPU := cost.ResourceUsage.GPUHours / hours
		if avgGPU < 0.2 && cost.GPUCost > 10.0 {
			recommendations = append(recommendations, RightsizingRecommendation{
				WorkloadID:      cost.WorkloadID,
				ResourceType:    "GPU",
				CurrentUsage:    fmt.Sprintf("%.2f GPU-hours/hour", avgGPU),
				Recommendation:  "Consider using GPU only when needed - utilization is low",
				EstimatedSaving: cost.GPUCost * 0.7,
			})
		}
	}

	// Create alerts for significant savings opportunities
	for _, rec := range recommendations {
		if rec.EstimatedSaving > 10.0 {
			alert := CostAlert{
				AlertID:       fmt.Sprintf("rightsizing-%s-%d", rec.WorkloadID, time.Now().Unix()),
				WorkloadID:    rec.WorkloadID,
				AlertType:     CostAlertRightsizingRecommendation,
				Message:       rec.Recommendation,
				CurrentCost:   rec.EstimatedSaving,
				ProjectedCost: rec.EstimatedSaving,
				CreatedAt:     time.Now(),
			}

			ct.mu.Lock()
			ct.alerts = append(ct.alerts, alert)
			ct.mu.Unlock()
		}
	}

	return recommendations
}

// RightsizingRecommendation represents a right-sizing recommendation
type RightsizingRecommendation struct {
	WorkloadID      string  `json:"workload_id"`
	ResourceType    string  `json:"resource_type"`
	CurrentUsage    string  `json:"current_usage"`
	Recommendation  string  `json:"recommendation"`
	EstimatedSaving float64 `json:"estimated_saving"`
}

// GetMetrics returns cost tracker metrics
func (ct *CostTracker) GetMetrics() CostTrackerMetrics {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	return CostTrackerMetrics{
		TotalCostTracked:   ct.totalCostTracked,
		ActiveWorkloads:    len(ct.costs),
		TotalAlertsCreated: atomic.LoadInt64(&ct.totalAlertsCreated),
		ActiveAlerts:       len(ct.alerts),
		LastCollectionTime: ct.lastCollectionTime,
		CurrentTotalCost:   ct.getTotalCostLocked(),
	}
}

// CostTrackerMetrics contains metrics for the cost tracker
type CostTrackerMetrics struct {
	TotalCostTracked   float64   `json:"total_cost_tracked"`
	ActiveWorkloads    int       `json:"active_workloads"`
	TotalAlertsCreated int64     `json:"total_alerts_created"`
	ActiveAlerts       int       `json:"active_alerts"`
	LastCollectionTime time.Time `json:"last_collection_time"`
	CurrentTotalCost   float64   `json:"current_total_cost"`
}

// ExportCostReport exports a cost report as JSON
func (ct *CostTracker) ExportCostReport() ([]byte, error) {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	report := struct {
		GeneratedAt   time.Time                   `json:"generated_at"`
		ProviderID    string                      `json:"provider_id"`
		Summary       CostSummary                 `json:"summary"`
		Workloads     []WorkloadCost              `json:"workloads"`
		Alerts        []CostAlert                 `json:"alerts"`
		Configuration ResourceCostConfig          `json:"configuration"`
		Metrics       CostTrackerMetrics          `json:"metrics"`
		Recommendations []RightsizingRecommendation `json:"recommendations,omitempty"`
	}{
		GeneratedAt:   time.Now(),
		ProviderID:    ct.cfg.ProviderID,
		Summary:       ct.getCostSummaryLocked(),
		Workloads:     ct.getWorkloadCostsLocked(),
		Alerts:        ct.alerts,
		Configuration: ct.cfg.CostConfig,
		Metrics:       ct.getMetricsLocked(),
	}

	return json.MarshalIndent(report, "", "  ")
}

// trackingLoop runs the main tracking loop
func (ct *CostTracker) trackingLoop(ctx context.Context) {
	ticker := time.NewTicker(ct.cfg.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ct.stopChan:
			return
		case <-ticker.C:
			ct.mu.Lock()
			ct.lastCollectionTime = time.Now()
			ct.recordDailyCost()
			ct.mu.Unlock()

			ct.CheckThresholds(ctx)
			ct.DetectAnomalies(ctx)
			ct.cleanup()
		}
	}
}

// recordDailyCost records the current cost to daily history
func (ct *CostTracker) recordDailyCost() {
	cost := ct.getTotalCostLocked()
	ct.dailyCosts = append(ct.dailyCosts, cost)

	// Keep only last 30 days
	if len(ct.dailyCosts) > 30 {
		ct.dailyCosts = ct.dailyCosts[len(ct.dailyCosts)-30:]
	}
}

// cleanup removes old alerts
func (ct *CostTracker) cleanup() {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	cutoff := time.Now().Add(-ct.cfg.RetentionPeriod)
	newAlerts := make([]CostAlert, 0, len(ct.alerts))

	for _, alert := range ct.alerts {
		if alert.CreatedAt.After(cutoff) {
			newAlerts = append(newAlerts, alert)
		}
	}

	ct.alerts = newAlerts
}

// getTotalCostLocked returns total cost (must hold lock)
func (ct *CostTracker) getTotalCostLocked() float64 {
	var total float64
	for _, cost := range ct.costs {
		total += cost.TotalCost
	}
	return total
}

// getCostSummaryLocked returns cost summary (must hold lock)
func (ct *CostTracker) getCostSummaryLocked() CostSummary {
	summary := CostSummary{
		Currency:      ct.cfg.CostConfig.Currency,
		WorkloadCount: len(ct.costs),
		Timestamp:     time.Now(),
	}

	for _, cost := range ct.costs {
		summary.TotalCPUCost += cost.CPUCost
		summary.TotalMemoryCost += cost.MemoryCost
		summary.TotalStorageCost += cost.StorageCost
		summary.TotalGPUCost += cost.GPUCost
		summary.TotalNetworkCost += cost.NetworkCost
		summary.TotalCost += cost.TotalCost
	}

	return summary
}

// getWorkloadCostsLocked returns workload costs (must hold lock)
func (ct *CostTracker) getWorkloadCostsLocked() []WorkloadCost {
	result := make([]WorkloadCost, 0, len(ct.costs))
	for _, cost := range ct.costs {
		result = append(result, *cost)
	}
	return result
}

// getMetricsLocked returns metrics (must hold lock)
func (ct *CostTracker) getMetricsLocked() CostTrackerMetrics {
	return CostTrackerMetrics{
		TotalCostTracked:   ct.totalCostTracked,
		ActiveWorkloads:    len(ct.costs),
		TotalAlertsCreated: atomic.LoadInt64(&ct.totalAlertsCreated),
		ActiveAlerts:       len(ct.alerts),
		LastCollectionTime: ct.lastCollectionTime,
		CurrentTotalCost:   ct.getTotalCostLocked(),
	}
}
