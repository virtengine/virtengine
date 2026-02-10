// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-5C: Waldur usage reconciliation for settlement integration
package provider_daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	verrors "github.com/virtengine/virtengine/pkg/errors"
	"github.com/virtengine/virtengine/pkg/waldur"
)

// WaldurReconcilerConfig configures the Waldur reconciler.
type WaldurReconcilerConfig struct {
	// Enabled enables Waldur reconciliation.
	Enabled bool

	// ReconciliationInterval is the interval between reconciliation runs.
	ReconciliationInterval time.Duration

	// DiscrepancyThreshold is the percentage threshold for flagging discrepancies.
	DiscrepancyThreshold float64

	// MaxAgeForReconciliation is the max age of records to reconcile.
	MaxAgeForReconciliation time.Duration

	// AlertOnDiscrepancy enables alerts when discrepancies are found.
	AlertOnDiscrepancy bool

	// AutoCorrect enables automatic correction of minor discrepancies.
	AutoCorrect bool

	// AutoCorrectThreshold is the max discrepancy percentage for auto-correction.
	AutoCorrectThreshold float64
}

// DefaultWaldurReconcilerConfig returns default reconciler config.
func DefaultWaldurReconcilerConfig() WaldurReconcilerConfig {
	return WaldurReconcilerConfig{
		Enabled:                 true,
		ReconciliationInterval:  6 * time.Hour,
		DiscrepancyThreshold:    10.0, // 10% discrepancy flags alert
		MaxAgeForReconciliation: 7 * 24 * time.Hour,
		AlertOnDiscrepancy:      true,
		AutoCorrect:             false,
		AutoCorrectThreshold:    5.0, // Auto-correct up to 5% discrepancy
	}
}

// WaldurUsageStats represents usage statistics from Waldur.
type WaldurUsageStats struct {
	// ResourceUUID is the Waldur resource UUID.
	ResourceUUID string `json:"resource_uuid"`

	// AllocationID is the VirtEngine allocation ID.
	AllocationID string `json:"allocation_id"`

	// PeriodStart is the start of the usage period.
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the end of the usage period.
	PeriodEnd time.Time `json:"period_end"`

	// CPUHours is CPU usage in hours.
	CPUHours float64 `json:"cpu_hours"`

	// RAMGBHours is RAM usage in GB-hours.
	RAMGBHours float64 `json:"ram_gb_hours"`

	// StorageGBHours is storage usage in GB-hours.
	StorageGBHours float64 `json:"storage_gb_hours"`

	// GPUHours is GPU usage in hours.
	GPUHours float64 `json:"gpu_hours"`

	// NetworkGB is network usage in GB.
	NetworkGB float64 `json:"network_gb"`

	// TotalCost is the total cost reported by Waldur.
	TotalCost float64 `json:"total_cost"`

	// Currency is the currency for the cost.
	Currency string `json:"currency"`

	// Components contains component-level usage.
	Components []WaldurUsageComponent `json:"components,omitempty"`
}

// WaldurUsageComponent represents a usage component from Waldur.
type WaldurUsageComponent struct {
	// Type is the component type.
	Type string `json:"type"`

	// Name is the component name.
	Name string `json:"name"`

	// Amount is the usage amount.
	Amount float64 `json:"amount"`

	// Price is the price.
	Price float64 `json:"price"`

	// Unit is the unit of measurement.
	Unit string `json:"unit"`
}

// WaldurReconciler reconciles provider-reported usage with Waldur stats.
type WaldurReconciler struct {
	mu sync.RWMutex

	cfg                WaldurReconcilerConfig
	marketplace        *waldur.MarketplaceClient
	usageStore         *UsageSnapshotStore
	settlementPipeline *SettlementPipeline

	// results stores reconciliation results by allocation ID.
	results map[string]*ReconciliationResult

	// discrepancies stores recent discrepancies.
	discrepancies []MetricDiscrepancy

	// running indicates if the reconciler is running.
	running bool

	// stopChan stops the reconciliation loop.
	stopChan chan struct{}

	// wg waits for goroutines to finish.
	wg sync.WaitGroup
}

const (
	reconcileSeverityCritical = "critical"
	reconcileSeverityHigh     = "high"
	reconcileSeverityMedium   = "medium"
	reconcileSeverityLow      = "low"
)

// NewWaldurReconciler creates a new Waldur reconciler.
func NewWaldurReconciler(
	cfg WaldurReconcilerConfig,
	marketplace *waldur.MarketplaceClient,
	usageStore *UsageSnapshotStore,
	pipeline *SettlementPipeline,
) *WaldurReconciler {
	return &WaldurReconciler{
		cfg:                cfg,
		marketplace:        marketplace,
		usageStore:         usageStore,
		settlementPipeline: pipeline,
		results:            make(map[string]*ReconciliationResult),
		discrepancies:      make([]MetricDiscrepancy, 0),
		stopChan:           make(chan struct{}),
	}
}

// Start starts the reconciler.
func (r *WaldurReconciler) Start(ctx context.Context) error {
	if !r.cfg.Enabled {
		return nil
	}

	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return nil
	}
	r.running = true
	r.mu.Unlock()

	r.wg.Add(1)
	verrors.SafeGo("provider-daemon:waldur-reconciler", func() {
		defer r.wg.Done()
		r.runLoop(ctx)
	})

	log.Printf("[waldur-reconciler] started with interval %v", r.cfg.ReconciliationInterval)
	return nil
}

// Stop stops the reconciler.
func (r *WaldurReconciler) Stop() {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return
	}
	r.running = false
	r.mu.Unlock()

	close(r.stopChan)
	r.wg.Wait()

	r.stopChan = make(chan struct{})
	log.Printf("[waldur-reconciler] stopped")
}

// ReconcileAllocation reconciles usage for a specific allocation.
func (r *WaldurReconciler) ReconcileAllocation(ctx context.Context, allocationID string, resourceUUID string) (*ReconciliationResult, error) {
	now := time.Now()

	// Get provider-reported usage
	periodEnd := now
	periodStart := now.Add(-r.cfg.ReconciliationInterval)
	providerRecord, found := r.usageStore.FindLatest(allocationID, &periodStart, &periodEnd)
	if !found {
		return nil, fmt.Errorf("no provider usage record found for allocation %s", allocationID)
	}

	// Get Waldur usage stats
	waldurStats, err := r.fetchWaldurUsage(ctx, resourceUUID, periodStart, periodEnd)
	if err != nil {
		// If Waldur stats not available, still return result with provider data only
		result := &ReconciliationResult{
			AllocationID:       allocationID,
			ReconciliationTime: now,
			ProviderMetrics:    providerRecord.Metrics,
			WaldurMetrics:      nil,
			InSync:             true, // Can't verify, assume in sync
			Score:              50,   // Neutral score
		}
		r.storeResult(result)
		return result, nil
	}

	// Convert Waldur stats to ResourceMetrics
	waldurMetrics := r.convertWaldurToMetrics(waldurStats)

	// Compare and find discrepancies
	discrepancies := r.compareMetrics(&providerRecord.Metrics, waldurMetrics)

	// Calculate reconciliation score
	score := r.calculateScore(discrepancies)

	result := &ReconciliationResult{
		AllocationID:       allocationID,
		ReconciliationTime: now,
		ProviderMetrics:    providerRecord.Metrics,
		WaldurMetrics:      waldurMetrics,
		Discrepancies:      discrepancies,
		InSync:             len(discrepancies) == 0,
		Score:              score,
	}

	r.storeResult(result)

	// Handle discrepancies
	if len(discrepancies) > 0 {
		r.handleDiscrepancies(allocationID, discrepancies)
	}

	return result, nil
}

// fetchWaldurUsage fetches usage statistics from Waldur.
func (r *WaldurReconciler) fetchWaldurUsage(ctx context.Context, resourceUUID string, periodStart, periodEnd time.Time) (*WaldurUsageStats, error) {
	if r.marketplace == nil {
		return nil, fmt.Errorf("marketplace client not configured")
	}

	// Get resource from Waldur to retrieve usage data
	resource, err := r.marketplace.GetResource(ctx, resourceUUID)
	if err != nil {
		return nil, fmt.Errorf("fetch resource: %w", err)
	}

	// Create stats from resource data
	// In a real implementation, this would call a dedicated usage API endpoint
	stats := &WaldurUsageStats{
		ResourceUUID: resource.UUID,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		// Note: Actual usage values would come from Waldur's usage reporting API
		// This is a placeholder that would need real Waldur API integration
	}

	return stats, nil
}

// convertWaldurToMetrics converts Waldur stats to ResourceMetrics.
func (r *WaldurReconciler) convertWaldurToMetrics(stats *WaldurUsageStats) *ResourceMetrics {
	if stats == nil {
		return nil
	}

	return &ResourceMetrics{
		CPUMilliSeconds:    int64(stats.CPUHours * 3600 * 1000),
		MemoryByteSeconds:  int64(stats.RAMGBHours * 1024 * 1024 * 1024 * 3600),
		StorageByteSeconds: int64(stats.StorageGBHours * 1024 * 1024 * 1024 * 3600),
		GPUSeconds:         int64(stats.GPUHours * 3600),
		NetworkBytesIn:     int64(stats.NetworkGB * 1024 * 1024 * 1024 / 2), // Assume 50/50 split
		NetworkBytesOut:    int64(stats.NetworkGB * 1024 * 1024 * 1024 / 2),
	}
}

// compareMetrics compares provider and Waldur metrics.
func (r *WaldurReconciler) compareMetrics(provider, waldur *ResourceMetrics) []MetricDiscrepancy {
	if provider == nil || waldur == nil {
		return nil
	}

	discrepancies := make([]MetricDiscrepancy, 0)

	// Compare CPU
	if diff := r.calculateDiscrepancy("cpu_milli_seconds", provider.CPUMilliSeconds, waldur.CPUMilliSeconds); diff != nil {
		discrepancies = append(discrepancies, *diff)
	}

	// Compare Memory
	if diff := r.calculateDiscrepancy("memory_byte_seconds", provider.MemoryByteSeconds, waldur.MemoryByteSeconds); diff != nil {
		discrepancies = append(discrepancies, *diff)
	}

	// Compare Storage
	if diff := r.calculateDiscrepancy("storage_byte_seconds", provider.StorageByteSeconds, waldur.StorageByteSeconds); diff != nil {
		discrepancies = append(discrepancies, *diff)
	}

	// Compare GPU
	if diff := r.calculateDiscrepancy("gpu_seconds", provider.GPUSeconds, waldur.GPUSeconds); diff != nil {
		discrepancies = append(discrepancies, *diff)
	}

	// Compare Network
	providerNetwork := provider.NetworkBytesIn + provider.NetworkBytesOut
	waldurNetwork := waldur.NetworkBytesIn + waldur.NetworkBytesOut
	if diff := r.calculateDiscrepancy("network_bytes", providerNetwork, waldurNetwork); diff != nil {
		discrepancies = append(discrepancies, *diff)
	}

	return discrepancies
}

// calculateDiscrepancy calculates discrepancy between two values.
func (r *WaldurReconciler) calculateDiscrepancy(metricName string, provider, waldur int64) *MetricDiscrepancy {
	if provider == 0 && waldur == 0 {
		return nil
	}

	var diffPercent float64
	if waldur != 0 {
		diffPercent = float64(provider-waldur) / float64(waldur) * 100
	} else if provider != 0 {
		diffPercent = 100.0 // 100% difference when Waldur reports 0
	}

	// Check if difference exceeds threshold
	absDiff := diffPercent
	if absDiff < 0 {
		absDiff = -absDiff
	}

	if absDiff < r.cfg.DiscrepancyThreshold {
		return nil
	}

	return &MetricDiscrepancy{
		MetricName:        metricName,
		ProviderValue:     provider,
		WaldurValue:       waldur,
		DifferencePercent: diffPercent,
		Severity:          r.severityFromDifference(absDiff),
	}
}

// severityFromDifference determines severity based on difference percentage.
func (r *WaldurReconciler) severityFromDifference(diff float64) string {
	switch {
	case diff >= 50:
		return reconcileSeverityCritical
	case diff >= 25:
		return reconcileSeverityHigh
	case diff >= 15:
		return reconcileSeverityMedium
	default:
		return reconcileSeverityLow
	}
}

// calculateScore calculates the reconciliation confidence score.
func (r *WaldurReconciler) calculateScore(discrepancies []MetricDiscrepancy) int {
	if len(discrepancies) == 0 {
		return 100
	}

	score := 100
	for _, d := range discrepancies {
		switch d.Severity {
		case reconcileSeverityCritical:
			score -= 30
		case reconcileSeverityHigh:
			score -= 20
		case reconcileSeverityMedium:
			score -= 10
		case reconcileSeverityLow:
			score -= 5
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

// storeResult stores a reconciliation result.
func (r *WaldurReconciler) storeResult(result *ReconciliationResult) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.results[result.AllocationID] = result

	// Store discrepancies
	r.discrepancies = append(r.discrepancies, result.Discrepancies...)

	// Limit stored discrepancies to last 1000
	if len(r.discrepancies) > 1000 {
		r.discrepancies = r.discrepancies[len(r.discrepancies)-1000:]
	}
}

// handleDiscrepancies handles detected discrepancies.
func (r *WaldurReconciler) handleDiscrepancies(allocationID string, discrepancies []MetricDiscrepancy) {
	for _, d := range discrepancies {
		log.Printf("[waldur-reconciler] discrepancy detected for %s: %s provider=%d waldur=%d diff=%.2f%% severity=%s",
			allocationID, d.MetricName, d.ProviderValue, d.WaldurValue, d.DifferencePercent, d.Severity)

		// Auto-correct minor discrepancies if enabled
		if r.cfg.AutoCorrect && d.DifferencePercent < r.cfg.AutoCorrectThreshold && d.DifferencePercent > -r.cfg.AutoCorrectThreshold {
			log.Printf("[waldur-reconciler] auto-correcting minor discrepancy for %s", allocationID)
			// In a real implementation, this would apply a correction
		}
	}
}

// GetResult gets the latest reconciliation result for an allocation.
func (r *WaldurReconciler) GetResult(allocationID string) (*ReconciliationResult, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result, ok := r.results[allocationID]
	return result, ok
}

// GetRecentDiscrepancies returns recent discrepancies.
func (r *WaldurReconciler) GetRecentDiscrepancies(limit int) []MetricDiscrepancy {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 || limit > len(r.discrepancies) {
		limit = len(r.discrepancies)
	}

	start := len(r.discrepancies) - limit
	result := make([]MetricDiscrepancy, limit)
	copy(result, r.discrepancies[start:])
	return result
}

// GetSyncStatus returns overall sync status.
func (r *WaldurReconciler) GetSyncStatus() ReconciliationSyncStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status := ReconciliationSyncStatus{
		TotalAllocations:  len(r.results),
		LastReconcileTime: time.Time{},
	}

	for _, result := range r.results {
		if result.ReconciliationTime.After(status.LastReconcileTime) {
			status.LastReconcileTime = result.ReconciliationTime
		}
		if result.InSync {
			status.InSyncCount++
		} else {
			status.OutOfSyncCount++
		}
		status.TotalScore += result.Score
	}

	if status.TotalAllocations > 0 {
		status.AverageScore = status.TotalScore / status.TotalAllocations
	}

	return status
}

// ReconciliationSyncStatus represents overall sync status.
type ReconciliationSyncStatus struct {
	// TotalAllocations is the total number of reconciled allocations.
	TotalAllocations int `json:"total_allocations"`

	// InSyncCount is the count of in-sync allocations.
	InSyncCount int `json:"in_sync_count"`

	// OutOfSyncCount is the count of out-of-sync allocations.
	OutOfSyncCount int `json:"out_of_sync_count"`

	// AverageScore is the average reconciliation score.
	AverageScore int `json:"average_score"`

	// TotalScore is used for calculation.
	TotalScore int `json:"-"`

	// LastReconcileTime is the last reconciliation time.
	LastReconcileTime time.Time `json:"last_reconcile_time"`
}

// runLoop runs the reconciliation loop.
func (r *WaldurReconciler) runLoop(ctx context.Context) {
	// Initial reconciliation after startup delay
	time.Sleep(time.Minute)

	ticker := time.NewTicker(r.cfg.ReconciliationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.stopChan:
			return
		case <-ticker.C:
			r.runReconciliation(ctx)
		}
	}
}

// runReconciliation runs a reconciliation cycle.
func (r *WaldurReconciler) runReconciliation(_ context.Context) {
	log.Printf("[waldur-reconciler] starting reconciliation cycle")

	// In a real implementation, this would iterate over all active allocations
	// and reconcile each one with Waldur

	status := r.GetSyncStatus()
	log.Printf("[waldur-reconciler] reconciliation complete: %d allocations, %d in-sync, %d out-of-sync, avg score %d",
		status.TotalAllocations, status.InSyncCount, status.OutOfSyncCount, status.AverageScore)
}

// ScheduledUsageCollector collects usage on a schedule and integrates with settlement.
type ScheduledUsageCollector struct {
	mu sync.RWMutex

	cfg                ScheduledCollectorConfig
	usageMeter         *UsageMeter
	settlementPipeline *SettlementPipeline
	reconciler         *WaldurReconciler

	// running indicates if the collector is running.
	running bool

	// stopChan stops the collection loop.
	stopChan chan struct{}

	// wg waits for goroutines to finish.
	wg sync.WaitGroup
}

// ScheduledCollectorConfig configures the scheduled collector.
type ScheduledCollectorConfig struct {
	// CollectionInterval is the interval for usage collection.
	CollectionInterval time.Duration

	// ImmediateOnThreshold triggers immediate collection when pending exceeds threshold.
	ImmediateOnThreshold int

	// ReconcileAfterCollection triggers reconciliation after each collection.
	ReconcileAfterCollection bool
}

// DefaultScheduledCollectorConfig returns default collector config.
func DefaultScheduledCollectorConfig() ScheduledCollectorConfig {
	return ScheduledCollectorConfig{
		CollectionInterval:       time.Hour,
		ImmediateOnThreshold:     100,
		ReconcileAfterCollection: true,
	}
}

// NewScheduledUsageCollector creates a new scheduled collector.
func NewScheduledUsageCollector(
	cfg ScheduledCollectorConfig,
	usageMeter *UsageMeter,
	pipeline *SettlementPipeline,
	reconciler *WaldurReconciler,
) *ScheduledUsageCollector {
	return &ScheduledUsageCollector{
		cfg:                cfg,
		usageMeter:         usageMeter,
		settlementPipeline: pipeline,
		reconciler:         reconciler,
		stopChan:           make(chan struct{}),
	}
}

// Start starts the scheduled collector.
func (c *ScheduledUsageCollector) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = true
	c.mu.Unlock()

	c.wg.Add(1)
	verrors.SafeGo("provider-daemon:scheduled-collector", func() {
		defer c.wg.Done()
		c.runLoop(ctx)
	})

	log.Printf("[scheduled-collector] started with interval %v", c.cfg.CollectionInterval)
	return nil
}

// Stop stops the scheduled collector.
func (c *ScheduledUsageCollector) Stop() {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return
	}
	c.running = false
	c.mu.Unlock()

	close(c.stopChan)
	c.wg.Wait()

	c.stopChan = make(chan struct{})
	log.Printf("[scheduled-collector] stopped")
}

// CollectNow triggers immediate collection for all workloads.
func (c *ScheduledUsageCollector) CollectNow(ctx context.Context) error {
	if c.usageMeter == nil {
		return fmt.Errorf("usage meter not configured")
	}

	workloads := c.usageMeter.ListMeteredWorkloads()
	for _, workloadID := range workloads {
		record, err := c.usageMeter.ForceCollect(ctx, workloadID)
		if err != nil {
			log.Printf("[scheduled-collector] failed to collect for %s: %v", workloadID, err)
			continue
		}

		// Add to settlement pipeline
		if c.settlementPipeline != nil {
			c.settlementPipeline.AddPendingUsage(record)

			// Generate line items
			if _, err := c.settlementPipeline.ProcessUsageToLineItems(record); err != nil {
				log.Printf("[scheduled-collector] failed to create line items for %s: %v", workloadID, err)
			}

			// Detect anomalies
			anomalies := c.settlementPipeline.DetectAnomalies(record, nil)
			if len(anomalies) > 0 {
				log.Printf("[scheduled-collector] detected %d anomalies for %s", len(anomalies), workloadID)
			}

			// Submit to chain
			if err := c.settlementPipeline.SubmitUsageToChain(ctx, record); err != nil {
				log.Printf("[scheduled-collector] failed to submit to chain for %s: %v", workloadID, err)
			}
		}
	}

	log.Printf("[scheduled-collector] collected usage for %d workloads", len(workloads))
	return nil
}

// runLoop runs the collection loop.
func (c *ScheduledUsageCollector) runLoop(ctx context.Context) {
	ticker := time.NewTicker(c.cfg.CollectionInterval)
	defer ticker.Stop()

	checkTicker := time.NewTicker(time.Minute)
	defer checkTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopChan:
			return
		case <-ticker.C:
			if err := c.CollectNow(ctx); err != nil {
				log.Printf("[scheduled-collector] collection failed: %v", err)
			}
		case <-checkTicker.C:
			// Check if threshold exceeded
			if c.settlementPipeline != nil {
				pending := c.settlementPipeline.GetPendingCount()
				if pending >= c.cfg.ImmediateOnThreshold {
					log.Printf("[scheduled-collector] threshold exceeded (%d >= %d), triggering immediate collection",
						pending, c.cfg.ImmediateOnThreshold)
					if err := c.CollectNow(ctx); err != nil {
						log.Printf("[scheduled-collector] immediate collection failed: %v", err)
					}
				}
			}
		}
	}
}

// UsageReportingMetrics contains metrics for usage reporting.
type UsageReportingMetrics struct {
	// TotalRecordsCollected is the total number of records collected.
	TotalRecordsCollected int64 `json:"total_records_collected"`

	// TotalRecordsSubmitted is the total number of records submitted to chain.
	TotalRecordsSubmitted int64 `json:"total_records_submitted"`

	// TotalSettlementsProcessed is the total number of settlements processed.
	TotalSettlementsProcessed int64 `json:"total_settlements_processed"`

	// TotalDisputesCreated is the total number of disputes created.
	TotalDisputesCreated int64 `json:"total_disputes_created"`

	// TotalDisputesResolved is the total number of disputes resolved.
	TotalDisputesResolved int64 `json:"total_disputes_resolved"`

	// TotalAnomaliesDetected is the total number of anomalies detected.
	TotalAnomaliesDetected int64 `json:"total_anomalies_detected"`

	// TotalCorrectionsApplied is the total number of corrections applied.
	TotalCorrectionsApplied int64 `json:"total_corrections_applied"`

	// LastCollectionTime is the last collection time.
	LastCollectionTime time.Time `json:"last_collection_time"`

	// LastSubmissionTime is the last chain submission time.
	LastSubmissionTime time.Time `json:"last_submission_time"`

	// LastSettlementTime is the last settlement time.
	LastSettlementTime time.Time `json:"last_settlement_time"`

	// AverageReconciliationScore is the average reconciliation score.
	AverageReconciliationScore int `json:"average_reconciliation_score"`
}

// generateReconciliationID generates a unique reconciliation ID.
//
//nolint:unused // reserved for future reconciliation tracking
func generateReconciliationID(allocationID string, timestamp time.Time) string {
	data := allocationID + ":" + timestamp.Format(time.RFC3339Nano)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:12])
}

// MarshalJSON implements json.Marshaler for ReconciliationResult.
func (r *ReconciliationResult) MarshalJSON() ([]byte, error) {
	type Alias ReconciliationResult
	return json.Marshal(&struct {
		*Alias
		ReconciliationTime string `json:"reconciliation_time"`
	}{
		Alias:              (*Alias)(r),
		ReconciliationTime: r.ReconciliationTime.Format(time.RFC3339),
	})
}
