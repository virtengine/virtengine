// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-5C: Waldur usage reconciliation for settlement integration
package provider_daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	verrors "github.com/virtengine/virtengine/pkg/errors"
	"github.com/virtengine/virtengine/pkg/waldur"
)

var ErrSettlementReconciliationHold = errors.New("settlement held pending matched reconciliation")

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

	// EvidenceTime is the newest timestamp among the independent records.
	EvidenceTime time.Time `json:"evidence_time"`
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
	usageClient        *waldur.UsageClient
	usageStore         *UsageSnapshotStore
	settlementPipeline *SettlementPipeline
	stateStore         *WaldurBridgeStateStore
	jobStore           ReconciliationJobStore
	metrics            *ReconciliationMetrics

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

// SetJobStore installs the required durable reconciliation store.
func (r *WaldurReconciler) SetJobStore(store ReconciliationJobStore) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobStore = store
}

// SetMetrics installs optional bounded reconciliation metrics.
func (r *WaldurReconciler) SetMetrics(metrics *ReconciliationMetrics) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics = metrics
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
	stateStore *WaldurBridgeStateStore,
) *WaldurReconciler {
	var usageClient *waldur.UsageClient
	if marketplace != nil {
		usageClient = waldur.NewUsageClient(marketplace)
	}

	return &WaldurReconciler{
		cfg:                cfg,
		marketplace:        marketplace,
		usageClient:        usageClient,
		usageStore:         usageStore,
		settlementPipeline: pipeline,
		stateStore:         stateStore,
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
	r.mu.RLock()
	alreadyRunning := r.running
	r.mu.RUnlock()
	if alreadyRunning {
		return nil
	}
	if r.jobStore == nil {
		return ErrReconciliationUnavailable
	}
	if err := r.jobStore.Open(ctx); err != nil {
		return fmt.Errorf("open reconciliation store: %w", err)
	}
	projection, err := r.jobStore.LoadProjection(ctx)
	if err != nil {
		_ = r.jobStore.Close()
		return fmt.Errorf("load reconciliation projection: %w", err)
	}
	r.metrics.ObserveProjection(projection)
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return nil
	}
	r.results = make(map[string]*ReconciliationResult, len(projection.Results))
	r.discrepancies = r.discrepancies[:0]
	for _, durable := range projection.Results {
		result := durable.Result
		r.results[result.AllocationID] = &result
		r.discrepancies = append(r.discrepancies, result.Discrepancies...)
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
	if r.jobStore != nil {
		_ = r.jobStore.Close()
	}

	r.stopChan = make(chan struct{})
	log.Printf("[waldur-reconciler] stopped")
}

// ReconcileAllocation reconciles usage for a specific allocation.
func (r *WaldurReconciler) ReconcileAllocation(ctx context.Context, allocationID string, resourceUUID string) (*ReconciliationResult, error) {
	now := time.Now()
	periodEnd := now
	periodStart := now.Add(-r.cfg.ReconciliationInterval)
	return r.reconcileAllocation(ctx, allocationID, resourceUUID, now, periodStart, periodEnd, true)
}

func (r *WaldurReconciler) reconcileAllocation(ctx context.Context, allocationID, resourceUUID string, now, periodStart, periodEnd time.Time, publish bool) (*ReconciliationResult, error) {
	if r.usageStore == nil {
		return nil, fmt.Errorf("usage store not configured")
	}
	publishResult := func(result *ReconciliationResult) {
		if publish {
			r.storeResult(result)
		}
	}

	// Get provider-reported usage
	providerRecord, found := r.usageStore.FindLatest(allocationID, &periodStart, &periodEnd)
	if !found {
		result := r.newClassifiedResult(allocationID, now, ResourceMetrics{}, ReconciliationStateUnavailable, ReconciliationReasonProviderEvidenceUnavailable)
		publishResult(result)
		return result, nil
	}
	if providerRecord.EndTime.IsZero() || providerRecord.EndTime.Before(providerRecord.StartTime) {
		result := r.newClassifiedResult(allocationID, now, providerRecord.Metrics, ReconciliationStateUnresolved, ReconciliationReasonMalformedEvidence)
		publishResult(result)
		return result, nil
	}
	if r.evidenceStale(now, providerRecord.EndTime) {
		result := r.newClassifiedResult(allocationID, now, providerRecord.Metrics, ReconciliationStateStale, ReconciliationReasonProviderEvidenceStale)
		publishResult(result)
		return result, nil
	}

	// Get Waldur usage stats
	waldurStats, err := r.fetchWaldurUsage(ctx, resourceUUID, periodStart, periodEnd)
	if err != nil {
		state, reason := ReconciliationStateUnavailable, ReconciliationReasonIndependentEvidenceUnavailable
		if validationErr, ok := err.(*reconciliationEvidenceError); ok {
			state, reason = ReconciliationStateUnresolved, validationErr.reason
		}
		result := r.newClassifiedResult(allocationID, now, providerRecord.Metrics, state, reason)
		publishResult(result)
		return result, nil
	}
	if r.evidenceStale(now, waldurStats.EvidenceTime) {
		result := r.newClassifiedResult(allocationID, now, providerRecord.Metrics, ReconciliationStateStale, ReconciliationReasonIndependentEvidenceStale)
		publishResult(result)
		return result, nil
	}
	if !hasCompleteIndependentEvidence(providerRecord.Metrics, waldurStats.Components) {
		result := r.newClassifiedResult(allocationID, now, providerRecord.Metrics, ReconciliationStateUnresolved, ReconciliationReasonPartialEvidence)
		publishResult(result)
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
		State:              ReconciliationStateMatched,
		ReasonCode:         ReconciliationReasonExactMatch,
		Score:              score,
	}
	if len(discrepancies) > 0 {
		result.State = ReconciliationStateMismatched
		result.ReasonCode = ReconciliationReasonMetricThresholdExceeded
	} else if providerRecord.Metrics != *waldurMetrics {
		result.ReasonCode = ReconciliationReasonWithinTolerance
	}

	publishResult(result)

	// Handle discrepancies
	if len(discrepancies) > 0 {
		r.handleDiscrepancies(allocationID, discrepancies)
	}

	return result, nil
}

type reconciliationEvidenceError struct {
	reason ReconciliationReasonCode
	msg    string
}

func (e *reconciliationEvidenceError) Error() string { return e.msg }

func (r *WaldurReconciler) newClassifiedResult(allocationID string, now time.Time, provider ResourceMetrics, state ReconciliationState, reason ReconciliationReasonCode) *ReconciliationResult {
	return &ReconciliationResult{AllocationID: allocationID, ReconciliationTime: now, ProviderMetrics: provider, State: state, ReasonCode: reason}
}

func (r *WaldurReconciler) evidenceStale(now, evidenceTime time.Time) bool {
	return evidenceTime.IsZero() || r.cfg.MaxAgeForReconciliation <= 0 || now.Sub(evidenceTime) > r.cfg.MaxAgeForReconciliation
}

// fetchWaldurUsage fetches usage statistics from Waldur.
func (r *WaldurReconciler) fetchWaldurUsage(ctx context.Context, resourceUUID string, periodStart, periodEnd time.Time) (*WaldurUsageStats, error) {
	if r.usageClient == nil {
		return nil, fmt.Errorf("waldur usage client not configured")
	}

	records, err := r.usageClient.GetResourceUsage(ctx, resourceUUID, &periodStart, &periodEnd)
	if err != nil {
		return nil, fmt.Errorf("fetch usage: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("independent usage evidence unavailable")
	}

	stats := &WaldurUsageStats{
		ResourceUUID: resourceUUID,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		Components:   make([]WaldurUsageComponent, 0, len(records)),
	}
	for _, record := range records {
		if err := validateWaldurUsageRecord(record, resourceUUID); err != nil {
			return nil, err
		}
		evidenceTime := record.Date
		if record.Created.After(evidenceTime) {
			evidenceTime = record.Created
		}
		if evidenceTime.After(stats.EvidenceTime) {
			stats.EvidenceTime = evidenceTime
		}
		r.applyWaldurUsageRecord(stats, record)
	}

	return stats, nil
}

func validateWaldurUsageRecord(record waldur.UsageRecord, resourceUUID string) error {
	componentType := strings.ToLower(strings.TrimSpace(record.ComponentType))
	if record.ResourceUUID != resourceUUID || componentType == "" || usageUnitForComponentType(componentType) == "unit" ||
		math.IsNaN(record.Usage) || math.IsInf(record.Usage, 0) || record.Usage < 0 || (record.Date.IsZero() && record.Created.IsZero()) {
		return &reconciliationEvidenceError{reason: ReconciliationReasonMalformedEvidence, msg: "malformed independent usage evidence"}
	}
	return nil
}

func hasCompleteIndependentEvidence(provider ResourceMetrics, components []WaldurUsageComponent) bool {
	present := make(map[string]bool)
	for _, component := range components {
		componentType := strings.ToLower(strings.TrimSpace(component.Type))
		switch {
		case strings.Contains(componentType, "cpu"):
			present["cpu"] = true
		case strings.Contains(componentType, "mem"), strings.Contains(componentType, "ram"):
			present["memory"] = true
		case strings.Contains(componentType, "storage"):
			present["storage"] = true
		case strings.Contains(componentType, "gpu"):
			present["gpu"] = true
		case strings.Contains(componentType, "network"), strings.Contains(componentType, "bandwidth"):
			present["network"] = true
		}
	}
	return (provider.CPUMilliSeconds == 0 || present["cpu"]) &&
		(provider.MemoryByteSeconds == 0 || present["memory"]) &&
		(provider.StorageByteSeconds == 0 || present["storage"]) &&
		(provider.GPUSeconds == 0 || present["gpu"]) &&
		(provider.NetworkBytesIn+provider.NetworkBytesOut == 0 || present["network"])
}

func (r *WaldurReconciler) applyWaldurUsageRecord(stats *WaldurUsageStats, record waldur.UsageRecord) {
	if stats == nil {
		return
	}

	componentType := strings.ToLower(strings.TrimSpace(record.ComponentType))
	stats.Components = append(stats.Components, WaldurUsageComponent{
		Type:   record.ComponentType,
		Name:   record.ComponentType,
		Amount: record.Usage,
		Unit:   usageUnitForComponentType(componentType),
	})

	switch {
	case strings.Contains(componentType, "cpu"):
		stats.CPUHours += record.Usage
	case strings.Contains(componentType, "mem"), strings.Contains(componentType, "ram"):
		stats.RAMGBHours += record.Usage
	case strings.Contains(componentType, "storage") && strings.Contains(componentType, "month"):
		stats.StorageGBHours += record.Usage * 24 * 30
	case strings.Contains(componentType, "storage"):
		stats.StorageGBHours += record.Usage
	case strings.Contains(componentType, "gpu"):
		stats.GPUHours += record.Usage
	case strings.Contains(componentType, "network"), strings.Contains(componentType, "bandwidth"):
		stats.NetworkGB += record.Usage
	}
}

func usageUnitForComponentType(componentType string) string {
	switch {
	case strings.Contains(componentType, "cpu"):
		return "cpu-hour"
	case strings.Contains(componentType, "mem"), strings.Contains(componentType, "ram"):
		return "gb-hour"
	case strings.Contains(componentType, "storage") && strings.Contains(componentType, "month"):
		return "gb-month"
	case strings.Contains(componentType, "storage"):
		return "gb-hour"
	case strings.Contains(componentType, "gpu"):
		return "gpu-hour"
	case strings.Contains(componentType, "network"), strings.Contains(componentType, "bandwidth"):
		return "gb"
	default:
		return "unit"
	}
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

// SettlementEligibility fails closed unless a durably published result is matched.
func (r *WaldurReconciler) SettlementEligibility(allocationID string) error {
	if r == nil || strings.TrimSpace(allocationID) == "" {
		return ErrSettlementReconciliationHold
	}
	result, found := r.GetResult(allocationID)
	if !found || result.State != ReconciliationStateMatched {
		return ErrSettlementReconciliationHold
	}
	return nil
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
		switch result.State {
		case ReconciliationStateMatched:
			status.MatchedCount++
		case ReconciliationStateMismatched:
			status.MismatchedCount++
		case ReconciliationStateUnavailable:
			status.UnavailableCount++
		case ReconciliationStateStale:
			status.StaleCount++
		case ReconciliationStateUnresolved:
			status.UnresolvedCount++
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

	MatchedCount     int `json:"matched_count"`
	MismatchedCount  int `json:"mismatched_count"`
	UnavailableCount int `json:"unavailable_count"`
	StaleCount       int `json:"stale_count"`
	UnresolvedCount  int `json:"unresolved_count"`

	// AverageScore is the average reconciliation score.
	AverageScore int `json:"average_score"`

	// TotalScore is used for calculation.
	TotalScore int `json:"-"`

	// LastReconcileTime is the last reconciliation time.
	LastReconcileTime time.Time `json:"last_reconcile_time"`
}

// runLoop runs the reconciliation loop.
func (r *WaldurReconciler) runLoop(ctx context.Context) {
	startDelay := time.NewTimer(time.Minute)
	defer startDelay.Stop()
	select {
	case <-ctx.Done():
		return
	case <-r.stopChan:
		return
	case <-startDelay.C:
	}

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
func (r *WaldurReconciler) runReconciliation(ctx context.Context) {
	log.Printf("[waldur-reconciler] starting reconciliation cycle")
	if r.stateStore == nil {
		log.Printf("[waldur-reconciler] skipped: no Waldur bridge state store configured")
		return
	}

	state, err := r.stateStore.Load()
	if err != nil {
		log.Printf("[waldur-reconciler] failed to load bridge state: %v", err)
		return
	}

	if r.jobStore == nil {
		log.Printf("[waldur-reconciler] skipped: durable reconciliation store unavailable")
		return
	}

	periodEnd := time.Now().UTC()
	periodStart := periodEnd.Add(-r.cfg.ReconciliationInterval)
	allocationIDs := make([]string, 0, len(state.Mappings))
	for allocationID := range state.Mappings {
		allocationIDs = append(allocationIDs, allocationID)
	}
	sort.Strings(allocationIDs)
	for _, allocationID := range allocationIDs {
		mapping := state.Mappings[allocationID]
		if mapping == nil || mapping.ResourceUUID == "" {
			continue
		}
		if mapping.AllocationID != "" {
			allocationID = mapping.AllocationID
		}
		job := newReconciliationJob(allocationID, mapping.ResourceUUID, periodStart, periodEnd)
		if _, _, err := r.jobStore.PutJobIfAbsent(ctx, job); err != nil {
			log.Printf("[waldur-reconciler] failed to persist job %s: %v", job.ID, err)
		}
	}
	r.refreshMetrics(ctx)

	pending, err := r.jobStore.PendingJobs(ctx)
	if err != nil {
		log.Printf("[waldur-reconciler] failed to load pending jobs: %v", err)
		return
	}
	processed := 0
	skipped := 0
	for _, job := range pending {
		attempt, err := r.jobStore.BeginAttempt(ctx, job.ID)
		if err != nil {
			log.Printf("[waldur-reconciler] failed to begin job %s: %v", job.ID, err)
			continue
		}
		result, err := r.reconcileAllocation(ctx, job.AllocationID, job.ResourceUUID, job.PeriodEnd, job.PeriodStart, job.PeriodEnd, false)
		if err != nil {
			_ = r.jobStore.FailAttempt(ctx, job.ID, attempt.Number, "reconciliation_error")
			r.refreshMetrics(ctx)
			log.Printf("[waldur-reconciler] failed to reconcile %s: %v", job.AllocationID, err)
			continue
		}
		durable, intents, cursor, err := buildDurableReconciliationCompletion(job, attempt, *result)
		if err != nil {
			_ = r.jobStore.FailAttempt(ctx, job.ID, attempt.Number, "evidence_digest_error")
			r.refreshMetrics(ctx)
			continue
		}
		if err := r.jobStore.CompleteAttempt(ctx, durable, intents, cursor); err != nil {
			log.Printf("[waldur-reconciler] failed to commit job %s: %v", job.ID, err)
			continue
		}
		r.refreshMetrics(ctx)
		r.storeResult(result)
		processed++
	}

	status := r.GetSyncStatus()
	log.Printf("[waldur-reconciler] reconciliation complete: %d processed, %d skipped, %d allocations, %d matched, %d mismatched, %d unavailable, %d stale, %d unresolved, avg score %d",
		processed, skipped, status.TotalAllocations, status.MatchedCount, status.MismatchedCount, status.UnavailableCount, status.StaleCount, status.UnresolvedCount, status.AverageScore)
}

func (r *WaldurReconciler) refreshMetrics(ctx context.Context) {
	if r.metrics == nil || r.jobStore == nil {
		return
	}
	projection, err := r.jobStore.LoadProjection(ctx)
	if err != nil {
		return
	}
	r.metrics.ObserveProjection(projection)
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
			allocationID := record.AllocationID
			if allocationID == "" {
				allocationID = record.DeploymentID
			}
			if err := c.reconciler.SettlementEligibility(allocationID); err != nil {
				log.Printf("[scheduled-collector] settlement held for %s: %v", workloadID, err)
				continue
			}
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
