// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-5C: Settlement pipeline for usage reporting to settlement integration
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

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	verrors "github.com/virtengine/virtengine/pkg/errors"
)

// SettlementConfig configures the settlement pipeline.
type SettlementConfig struct {
	// ProviderAddress is the provider's on-chain address.
	ProviderAddress string

	// SettlementInterval is the interval for periodic settlements.
	SettlementInterval time.Duration

	// DisputeWindow is the time window for disputes after usage is reported.
	DisputeWindow time.Duration

	// ReconciliationInterval is the interval for Waldur reconciliation.
	ReconciliationInterval time.Duration

	// AnomalyThresholds contains thresholds for anomaly detection.
	AnomalyThresholds AnomalyThresholds

	// MaxPendingRecords is the max number of pending records before forcing settlement.
	MaxPendingRecords int

	// RetryAttempts is the number of retry attempts for failed submissions.
	RetryAttempts int

	// RetryBackoff is the initial backoff duration for retries.
	RetryBackoff time.Duration
}

// DefaultSettlementConfig returns default settlement configuration.
func DefaultSettlementConfig() SettlementConfig {
	return SettlementConfig{
		SettlementInterval:     time.Hour,
		DisputeWindow:          24 * time.Hour,
		ReconciliationInterval: 6 * time.Hour,
		AnomalyThresholds:      DefaultAnomalyThresholds(),
		MaxPendingRecords:      100,
		RetryAttempts:          3,
		RetryBackoff:           time.Second * 5,
	}
}

// AnomalyThresholds contains thresholds for anomaly detection.
type AnomalyThresholds struct {
	// MaxCPUVariance is the max variance allowed in CPU usage (percentage).
	MaxCPUVariance float64

	// MaxMemoryVariance is the max variance allowed in memory usage (percentage).
	MaxMemoryVariance float64

	// MaxCostVariance is the max variance allowed in cost calculations (percentage).
	MaxCostVariance float64

	// MinRecordDuration is the minimum duration for a valid record.
	MinRecordDuration time.Duration

	// MaxRecordDuration is the maximum duration for a valid record.
	MaxRecordDuration time.Duration
}

// DefaultAnomalyThresholds returns default anomaly thresholds.
func DefaultAnomalyThresholds() AnomalyThresholds {
	return AnomalyThresholds{
		MaxCPUVariance:    50.0,
		MaxMemoryVariance: 50.0,
		MaxCostVariance:   25.0,
		MinRecordDuration: time.Minute,
		MaxRecordDuration: 25 * time.Hour,
	}
}

// BillableLineItem represents a line item for invoicing.
type BillableLineItem struct {
	// LineItemID is the unique identifier for this line item.
	LineItemID string `json:"line_item_id"`

	// OrderID is the linked marketplace order.
	OrderID string `json:"order_id"`

	// LeaseID is the linked lease.
	LeaseID string `json:"lease_id"`

	// UsageRecordID is the linked usage record.
	UsageRecordID string `json:"usage_record_id"`

	// ResourceType is the type of resource (cpu, memory, storage, gpu, network).
	ResourceType string `json:"resource_type"`

	// Quantity is the quantity of resource consumed.
	Quantity sdkmath.LegacyDec `json:"quantity"`

	// Unit is the unit of measurement.
	Unit string `json:"unit"`

	// UnitPrice is the price per unit.
	UnitPrice sdk.DecCoin `json:"unit_price"`

	// TotalCost is the total cost for this line item.
	TotalCost sdk.Coin `json:"total_cost"`

	// PeriodStart is the start of the billing period.
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the end of the billing period.
	PeriodEnd time.Time `json:"period_end"`

	// CreatedAt is when the line item was created.
	CreatedAt time.Time `json:"created_at"`

	// Disputed indicates if this line item is disputed.
	Disputed bool `json:"disputed"`

	// DisputeReason is the reason for the dispute.
	DisputeReason string `json:"dispute_reason,omitempty"`
}

// Hash generates a hash of the billable line item.
func (b *BillableLineItem) Hash() []byte {
	data := struct {
		OrderID      string `json:"order_id"`
		LeaseID      string `json:"lease_id"`
		ResourceType string `json:"resource_type"`
		Quantity     string `json:"quantity"`
		PeriodStart  int64  `json:"period_start"`
		PeriodEnd    int64  `json:"period_end"`
	}{
		OrderID:      b.OrderID,
		LeaseID:      b.LeaseID,
		ResourceType: b.ResourceType,
		Quantity:     b.Quantity.String(),
		PeriodStart:  b.PeriodStart.Unix(),
		PeriodEnd:    b.PeriodEnd.Unix(),
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	hash := sha256.Sum256(bytes)
	return hash[:]
}

// UsageDispute represents a dispute on a usage record.
type UsageDispute struct {
	// DisputeID is the unique identifier for this dispute.
	DisputeID string `json:"dispute_id"`

	// UsageRecordID is the disputed usage record.
	UsageRecordID string `json:"usage_record_id"`

	// OrderID is the linked order.
	OrderID string `json:"order_id"`

	// Initiator is the address of the dispute initiator.
	Initiator string `json:"initiator"`

	// Reason is the reason for the dispute.
	Reason string `json:"reason"`

	// Evidence is additional evidence for the dispute.
	Evidence string `json:"evidence,omitempty"`

	// ExpectedUsage is the expected usage value.
	ExpectedUsage *ResourceMetrics `json:"expected_usage,omitempty"`

	// ReportedUsage is the reported usage value.
	ReportedUsage *ResourceMetrics `json:"reported_usage,omitempty"`

	// Status is the dispute status.
	Status DisputeStatus `json:"status"`

	// Resolution is the dispute resolution.
	Resolution string `json:"resolution,omitempty"`

	// CreatedAt is when the dispute was created.
	CreatedAt time.Time `json:"created_at"`

	// ResolvedAt is when the dispute was resolved.
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`

	// ExpiresAt is when the dispute window expires.
	ExpiresAt time.Time `json:"expires_at"`
}

// DisputeStatus represents the status of a dispute.
type DisputeStatus string

const (
	DisputeStatusPending   DisputeStatus = "pending"
	DisputeStatusReviewing DisputeStatus = "reviewing"
	DisputeStatusResolved  DisputeStatus = "resolved"
	DisputeStatusRejected  DisputeStatus = "rejected"
	DisputeStatusExpired   DisputeStatus = "expired"
)

// UsageCorrection represents a correction to a usage record.
type UsageCorrection struct {
	// CorrectionID is the unique identifier for this correction.
	CorrectionID string `json:"correction_id"`

	// OriginalUsageID is the original usage record ID.
	OriginalUsageID string `json:"original_usage_id"`

	// DisputeID is the linked dispute (if any).
	DisputeID string `json:"dispute_id,omitempty"`

	// OriginalMetrics is the original metrics.
	OriginalMetrics ResourceMetrics `json:"original_metrics"`

	// CorrectedMetrics is the corrected metrics.
	CorrectedMetrics ResourceMetrics `json:"corrected_metrics"`

	// Reason is the reason for the correction.
	Reason string `json:"reason"`

	// AppliedAt is when the correction was applied.
	AppliedAt time.Time `json:"applied_at"`

	// Signature is the signature on the correction.
	Signature string `json:"signature"`
}

// ReconciliationResult contains the result of a reconciliation check.
type ReconciliationResult struct {
	// AllocationID is the allocation being reconciled.
	AllocationID string `json:"allocation_id"`

	// ReconciliationTime is when the reconciliation occurred.
	ReconciliationTime time.Time `json:"reconciliation_time"`

	// ProviderMetrics is the provider-reported metrics.
	ProviderMetrics ResourceMetrics `json:"provider_metrics"`

	// WaldurMetrics is the Waldur-reported metrics (if available).
	WaldurMetrics *ResourceMetrics `json:"waldur_metrics,omitempty"`

	// Discrepancies contains any discrepancies found.
	Discrepancies []MetricDiscrepancy `json:"discrepancies,omitempty"`

	// InSync indicates if provider and Waldur data are in sync.
	InSync bool `json:"in_sync"`

	// Score is the reconciliation confidence score (0-100).
	Score int `json:"score"`
}

// MetricDiscrepancy represents a discrepancy between metrics sources.
type MetricDiscrepancy struct {
	// MetricName is the name of the metric.
	MetricName string `json:"metric_name"`

	// ProviderValue is the provider-reported value.
	ProviderValue int64 `json:"provider_value"`

	// WaldurValue is the Waldur-reported value.
	WaldurValue int64 `json:"waldur_value"`

	// DifferencePercent is the difference as a percentage.
	DifferencePercent float64 `json:"difference_percent"`

	// Severity is the severity of the discrepancy.
	Severity string `json:"severity"`
}

// UsageAnomaly represents a detected usage anomaly.
type UsageAnomaly struct {
	// AnomalyID is the unique identifier for this anomaly.
	AnomalyID string `json:"anomaly_id"`

	// UsageRecordID is the linked usage record.
	UsageRecordID string `json:"usage_record_id"`

	// OrderID is the linked order.
	OrderID string `json:"order_id"`

	// AnomalyType is the type of anomaly detected.
	AnomalyType string `json:"anomaly_type"`

	// Description is the anomaly description.
	Description string `json:"description"`

	// Severity is the severity (low, medium, high, critical).
	Severity string `json:"severity"`

	// Value is the anomalous value.
	Value float64 `json:"value"`

	// ExpectedRange is the expected range.
	ExpectedRange string `json:"expected_range"`

	// DetectedAt is when the anomaly was detected.
	DetectedAt time.Time `json:"detected_at"`

	// Acknowledged indicates if the anomaly was acknowledged.
	Acknowledged bool `json:"acknowledged"`
}

// ChainUsageSubmitter submits usage records to the chain.
type ChainUsageSubmitter interface {
	// SubmitUsageReport submits a usage report to the chain.
	SubmitUsageReport(ctx context.Context, report *ChainUsageReport) error

	// SubmitSettlementRequest submits a settlement request.
	SubmitSettlementRequest(ctx context.Context, orderID string, usageRecordIDs []string, isFinal bool) error
}

// ChainUsageReport represents a usage report for chain submission.
type ChainUsageReport struct {
	// OrderID is the marketplace order ID.
	OrderID string `json:"order_id"`

	// LeaseID is the lease ID.
	LeaseID string `json:"lease_id"`

	// UsageUnits is the number of usage units.
	UsageUnits uint64 `json:"usage_units"`

	// UsageType is the type of usage.
	UsageType string `json:"usage_type"`

	// PeriodStart is the start of the period.
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the end of the period.
	PeriodEnd time.Time `json:"period_end"`

	// UnitPrice is the price per unit.
	UnitPrice sdk.DecCoin `json:"unit_price"`

	// Signature is the provider's signature.
	Signature []byte `json:"signature"`
}

// SettlementPipeline coordinates usage reporting to settlement.
type SettlementPipeline struct {
	mu sync.RWMutex

	cfg           SettlementConfig
	keyManager    *KeyManager
	usageMeter    *UsageMeter
	usageStore    *UsageSnapshotStore
	chainSubmit   ChainUsageSubmitter

	// pending contains usage records pending settlement.
	pending map[string]*UsageRecord

	// disputes contains active disputes.
	disputes map[string]*UsageDispute

	// corrections contains applied corrections.
	corrections map[string]*UsageCorrection

	// anomalies contains detected anomalies.
	anomalies map[string]*UsageAnomaly

	// reconciliations contains recent reconciliation results.
	reconciliations map[string]*ReconciliationResult

	// lineItems contains generated line items.
	lineItems map[string]*BillableLineItem

	// running indicates if the pipeline is running.
	running bool

	// stopChan stops the pipeline loop.
	stopChan chan struct{}

	// wg waits for goroutines to finish.
	wg sync.WaitGroup
}

// NewSettlementPipeline creates a new settlement pipeline.
func NewSettlementPipeline(
	cfg SettlementConfig,
	keyManager *KeyManager,
	usageMeter *UsageMeter,
	usageStore *UsageSnapshotStore,
	chainSubmit ChainUsageSubmitter,
) *SettlementPipeline {
	return &SettlementPipeline{
		cfg:             cfg,
		keyManager:      keyManager,
		usageMeter:      usageMeter,
		usageStore:      usageStore,
		chainSubmit:     chainSubmit,
		pending:         make(map[string]*UsageRecord),
		disputes:        make(map[string]*UsageDispute),
		corrections:     make(map[string]*UsageCorrection),
		anomalies:       make(map[string]*UsageAnomaly),
		reconciliations: make(map[string]*ReconciliationResult),
		lineItems:       make(map[string]*BillableLineItem),
		stopChan:        make(chan struct{}),
	}
}

// Start starts the settlement pipeline.
func (p *SettlementPipeline) Start(ctx context.Context) error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return nil
	}
	p.running = true
	p.mu.Unlock()

	p.wg.Add(1)
	verrors.SafeGo("provider-daemon:settlement-pipeline", func() {
		defer p.wg.Done()
		p.runLoop(ctx)
	})

	return nil
}

// Stop stops the settlement pipeline.
func (p *SettlementPipeline) Stop() {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return
	}
	p.running = false
	p.mu.Unlock()

	close(p.stopChan)
	p.wg.Wait()

	p.stopChan = make(chan struct{})
}

// AddPendingUsage adds a usage record to the pending queue.
func (p *SettlementPipeline) AddPendingUsage(record *UsageRecord) {
	if record == nil || record.ID == "" {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.pending[record.ID] = record
	p.usageStore.Track(record)
}

// GetPendingCount returns the number of pending usage records.
func (p *SettlementPipeline) GetPendingCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.pending)
}

// CreateDispute creates a new usage dispute.
func (p *SettlementPipeline) CreateDispute(
	usageRecordID string,
	orderID string,
	initiator string,
	reason string,
	evidence string,
	expectedUsage *ResourceMetrics,
) (*UsageDispute, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	record, ok := p.pending[usageRecordID]
	if !ok {
		return nil, fmt.Errorf("usage record not found: %s", usageRecordID)
	}

	now := time.Now()
	disputeID := p.generateID("dispute", now)

	dispute := &UsageDispute{
		DisputeID:     disputeID,
		UsageRecordID: usageRecordID,
		OrderID:       orderID,
		Initiator:     initiator,
		Reason:        reason,
		Evidence:      evidence,
		ExpectedUsage: expectedUsage,
		ReportedUsage: &record.Metrics,
		Status:        DisputeStatusPending,
		CreatedAt:     now,
		ExpiresAt:     now.Add(p.cfg.DisputeWindow),
	}

	p.disputes[disputeID] = dispute
	return dispute, nil
}

// ResolveDispute resolves a dispute.
func (p *SettlementPipeline) ResolveDispute(disputeID string, resolution string, accept bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	dispute, ok := p.disputes[disputeID]
	if !ok {
		return fmt.Errorf("dispute not found: %s", disputeID)
	}

	if dispute.Status != DisputeStatusPending && dispute.Status != DisputeStatusReviewing {
		return fmt.Errorf("dispute already resolved: %s", disputeID)
	}

	now := time.Now()
	dispute.Resolution = resolution
	dispute.ResolvedAt = &now

	if accept {
		dispute.Status = DisputeStatusResolved

		// Apply correction if expected usage was provided
		if dispute.ExpectedUsage != nil {
			if err := p.applyCorrection(dispute); err != nil {
				log.Printf("[settlement-pipeline] failed to apply correction: %v", err)
			}
		}
	} else {
		dispute.Status = DisputeStatusRejected
	}

	return nil
}

// applyCorrection applies a usage correction from a dispute.
func (p *SettlementPipeline) applyCorrection(dispute *UsageDispute) error {
	record, ok := p.pending[dispute.UsageRecordID]
	if !ok {
		return fmt.Errorf("usage record not found for correction")
	}

	now := time.Now()
	correctionID := p.generateID("correction", now)

	correction := &UsageCorrection{
		CorrectionID:     correctionID,
		OriginalUsageID:  dispute.UsageRecordID,
		DisputeID:        dispute.DisputeID,
		OriginalMetrics:  record.Metrics,
		CorrectedMetrics: *dispute.ExpectedUsage,
		Reason:           dispute.Resolution,
		AppliedAt:        now,
	}

	// Sign the correction
	if p.keyManager != nil {
		hash := sha256.Sum256([]byte(correctionID + dispute.UsageRecordID))
		sig, err := p.keyManager.Sign(hash[:])
		if err == nil {
			correction.Signature = sig.Signature
		}
	}

	p.corrections[correctionID] = correction

	// Update the usage record with corrected metrics
	record.Metrics = *dispute.ExpectedUsage

	return nil
}

// ProcessUsageToLineItems converts usage records to billable line items.
func (p *SettlementPipeline) ProcessUsageToLineItems(record *UsageRecord) ([]*BillableLineItem, error) {
	if record == nil {
		return nil, fmt.Errorf("usage record is nil")
	}

	now := time.Now()
	items := make([]*BillableLineItem, 0)

	// Convert CPU usage
	if record.Metrics.CPUMilliSeconds > 0 {
		cpuHours := float64(record.Metrics.CPUMilliSeconds) / (1000.0 * 3600.0)
		item := p.createLineItem(record, "cpu", cpuHours, "cpu-hours", record.PricingInputs.AgreedCPURate, now)
		if item != nil {
			items = append(items, item)
		}
	}

	// Convert Memory usage
	if record.Metrics.MemoryByteSeconds > 0 {
		memGBHours := float64(record.Metrics.MemoryByteSeconds) / (1024.0 * 1024.0 * 1024.0 * 3600.0)
		item := p.createLineItem(record, "memory", memGBHours, "gb-hours", record.PricingInputs.AgreedMemoryRate, now)
		if item != nil {
			items = append(items, item)
		}
	}

	// Convert Storage usage
	if record.Metrics.StorageByteSeconds > 0 {
		storageGBHours := float64(record.Metrics.StorageByteSeconds) / (1024.0 * 1024.0 * 1024.0 * 3600.0)
		item := p.createLineItem(record, "storage", storageGBHours, "gb-hours", record.PricingInputs.AgreedStorageRate, now)
		if item != nil {
			items = append(items, item)
		}
	}

	// Convert GPU usage
	if record.Metrics.GPUSeconds > 0 {
		gpuHours := float64(record.Metrics.GPUSeconds) / 3600.0
		item := p.createLineItem(record, "gpu", gpuHours, "gpu-hours", record.PricingInputs.AgreedGPURate, now)
		if item != nil {
			items = append(items, item)
		}
	}

	// Convert Network usage
	networkBytes := record.Metrics.NetworkBytesIn + record.Metrics.NetworkBytesOut
	if networkBytes > 0 {
		networkGB := float64(networkBytes) / (1024.0 * 1024.0 * 1024.0)
		item := p.createLineItem(record, "network", networkGB, "gb", record.PricingInputs.AgreedNetworkRate, now)
		if item != nil {
			items = append(items, item)
		}
	}

	// Store line items
	p.mu.Lock()
	for _, item := range items {
		p.lineItems[item.LineItemID] = item
	}
	p.mu.Unlock()

	return items, nil
}

// createLineItem creates a billable line item for a resource type.
func (p *SettlementPipeline) createLineItem(
	record *UsageRecord,
	resourceType string,
	quantity float64,
	unit string,
	rateStr string,
	now time.Time,
) *BillableLineItem {
	if rateStr == "" {
		return nil
	}

	rate, err := sdkmath.LegacyNewDecFromStr(rateStr)
	if err != nil {
		return nil
	}

	lineItemID := p.generateID("line-"+resourceType, now)
	quantityDec := sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(int64(quantity * 1000000))).QuoInt64(1000000)
	// Rate is in virt per unit, convert to uvirt (1 virt = 1,000,000 uvirt)
	uvirtMultiplier := sdkmath.LegacyNewDec(1000000)
	totalAmount := rate.Mul(quantityDec).Mul(uvirtMultiplier).TruncateInt()

	return &BillableLineItem{
		LineItemID:    lineItemID,
		OrderID:       record.DeploymentID,
		LeaseID:       record.LeaseID,
		UsageRecordID: record.ID,
		ResourceType:  resourceType,
		Quantity:      quantityDec,
		Unit:          unit,
		UnitPrice:     sdk.NewDecCoinFromDec("uvirt", rate),
		TotalCost:     sdk.NewCoin("uvirt", totalAmount),
		PeriodStart:   record.StartTime,
		PeriodEnd:     record.EndTime,
		CreatedAt:     now,
	}
}

// DetectAnomalies detects anomalies in a usage record.
func (p *SettlementPipeline) DetectAnomalies(record *UsageRecord, allocatedResources *ResourceMetrics) []*UsageAnomaly {
	anomalies := make([]*UsageAnomaly, 0)
	now := time.Now()

	// Check duration anomalies
	duration := record.EndTime.Sub(record.StartTime)
	if duration < p.cfg.AnomalyThresholds.MinRecordDuration {
		anomalies = append(anomalies, &UsageAnomaly{
			AnomalyID:     p.generateID("anomaly", now),
			UsageRecordID: record.ID,
			OrderID:       record.DeploymentID,
			AnomalyType:   "duration_too_short",
			Description:   fmt.Sprintf("Record duration %v is below minimum %v", duration, p.cfg.AnomalyThresholds.MinRecordDuration),
			Severity:      "medium",
			Value:         duration.Seconds(),
			ExpectedRange: fmt.Sprintf(">= %v", p.cfg.AnomalyThresholds.MinRecordDuration),
			DetectedAt:    now,
		})
	}

	if duration > p.cfg.AnomalyThresholds.MaxRecordDuration {
		anomalies = append(anomalies, &UsageAnomaly{
			AnomalyID:     p.generateID("anomaly", now),
			UsageRecordID: record.ID,
			OrderID:       record.DeploymentID,
			AnomalyType:   "duration_too_long",
			Description:   fmt.Sprintf("Record duration %v exceeds maximum %v", duration, p.cfg.AnomalyThresholds.MaxRecordDuration),
			Severity:      "high",
			Value:         duration.Seconds(),
			ExpectedRange: fmt.Sprintf("<= %v", p.cfg.AnomalyThresholds.MaxRecordDuration),
			DetectedAt:    now,
		})
	}

	// Check resource usage anomalies if allocated resources provided
	if allocatedResources != nil {
		durationHours := duration.Hours()
		if durationHours > 0 {
			// CPU variance check
			expectedCPU := float64(allocatedResources.CPUMilliSeconds) * durationHours
			if expectedCPU > 0 {
				variance := (float64(record.Metrics.CPUMilliSeconds) - expectedCPU) / expectedCPU * 100
				if variance > p.cfg.AnomalyThresholds.MaxCPUVariance || variance < -p.cfg.AnomalyThresholds.MaxCPUVariance {
					anomalies = append(anomalies, &UsageAnomaly{
						AnomalyID:     p.generateID("anomaly", now),
						UsageRecordID: record.ID,
						OrderID:       record.DeploymentID,
						AnomalyType:   "cpu_variance",
						Description:   fmt.Sprintf("CPU usage variance %.2f%% exceeds threshold", variance),
						Severity:      p.severityFromVariance(variance),
						Value:         variance,
						ExpectedRange: fmt.Sprintf("+/- %.0f%%", p.cfg.AnomalyThresholds.MaxCPUVariance),
						DetectedAt:    now,
					})
				}
			}

			// Memory variance check
			expectedMem := float64(allocatedResources.MemoryByteSeconds) * durationHours * 3600
			if expectedMem > 0 {
				variance := (float64(record.Metrics.MemoryByteSeconds) - expectedMem) / expectedMem * 100
				if variance > p.cfg.AnomalyThresholds.MaxMemoryVariance || variance < -p.cfg.AnomalyThresholds.MaxMemoryVariance {
					anomalies = append(anomalies, &UsageAnomaly{
						AnomalyID:     p.generateID("anomaly", now),
						UsageRecordID: record.ID,
						OrderID:       record.DeploymentID,
						AnomalyType:   "memory_variance",
						Description:   fmt.Sprintf("Memory usage variance %.2f%% exceeds threshold", variance),
						Severity:      p.severityFromVariance(variance),
						Value:         variance,
						ExpectedRange: fmt.Sprintf("+/- %.0f%%", p.cfg.AnomalyThresholds.MaxMemoryVariance),
						DetectedAt:    now,
					})
				}
			}
		}
	}

	// Check for negative values
	if record.Metrics.CPUMilliSeconds < 0 || record.Metrics.MemoryByteSeconds < 0 ||
		record.Metrics.StorageByteSeconds < 0 || record.Metrics.GPUSeconds < 0 ||
		record.Metrics.NetworkBytesIn < 0 || record.Metrics.NetworkBytesOut < 0 {
		anomalies = append(anomalies, &UsageAnomaly{
			AnomalyID:     p.generateID("anomaly", now),
			UsageRecordID: record.ID,
			OrderID:       record.DeploymentID,
			AnomalyType:   "negative_values",
			Description:   "Usage record contains negative values",
			Severity:      "critical",
			DetectedAt:    now,
		})
	}

	// Store detected anomalies
	p.mu.Lock()
	for _, a := range anomalies {
		p.anomalies[a.AnomalyID] = a
	}
	p.mu.Unlock()

	return anomalies
}

// severityFromVariance determines severity based on variance.
func (p *SettlementPipeline) severityFromVariance(variance float64) string {
	absVariance := variance
	if absVariance < 0 {
		absVariance = -absVariance
	}

	switch {
	case absVariance > 100:
		return "critical"
	case absVariance > 75:
		return "high"
	case absVariance > 50:
		return "medium"
	default:
		return "low"
	}
}

// GetActiveDisputes returns all active disputes.
func (p *SettlementPipeline) GetActiveDisputes() []*UsageDispute {
	p.mu.RLock()
	defer p.mu.RUnlock()

	active := make([]*UsageDispute, 0)
	for _, d := range p.disputes {
		if d.Status == DisputeStatusPending || d.Status == DisputeStatusReviewing {
			active = append(active, d)
		}
	}
	return active
}

// GetUnacknowledgedAnomalies returns all unacknowledged anomalies.
func (p *SettlementPipeline) GetUnacknowledgedAnomalies() []*UsageAnomaly {
	p.mu.RLock()
	defer p.mu.RUnlock()

	unacked := make([]*UsageAnomaly, 0)
	for _, a := range p.anomalies {
		if !a.Acknowledged {
			unacked = append(unacked, a)
		}
	}
	return unacked
}

// AcknowledgeAnomaly acknowledges an anomaly.
func (p *SettlementPipeline) AcknowledgeAnomaly(anomalyID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	anomaly, ok := p.anomalies[anomalyID]
	if !ok {
		return fmt.Errorf("anomaly not found: %s", anomalyID)
	}

	anomaly.Acknowledged = true
	return nil
}

// SubmitUsageToChain submits a usage record to the chain.
func (p *SettlementPipeline) SubmitUsageToChain(ctx context.Context, record *UsageRecord) error {
	if p.chainSubmit == nil {
		return fmt.Errorf("chain submitter not configured")
	}

	// Calculate usage units (sum of normalized resource usage)
	usageUnits := p.calculateUsageUnits(record)

	// Determine primary usage type
	usageType := p.determineUsageType(record)

	// Get the primary rate for unit price
	unitPrice := p.getPrimaryRate(record, usageType)

	report := &ChainUsageReport{
		OrderID:     record.DeploymentID,
		LeaseID:     record.LeaseID,
		UsageUnits:  usageUnits,
		UsageType:   usageType,
		PeriodStart: record.StartTime,
		PeriodEnd:   record.EndTime,
		UnitPrice:   unitPrice,
	}

	// Sign the report
	if p.keyManager != nil {
		hash := record.Hash()
		sig, err := p.keyManager.Sign(hash)
		if err == nil {
			sigBytes, _ := hex.DecodeString(sig.Signature)
			report.Signature = sigBytes
		}
	}

	return p.chainSubmit.SubmitUsageReport(ctx, report)
}

// calculateUsageUnits calculates usage units from a record.
func (p *SettlementPipeline) calculateUsageUnits(record *UsageRecord) uint64 {
	var units uint64

	// Convert each resource type to normalized units
	// CPU: 1 unit = 1 CPU-hour
	units += uint64(record.Metrics.CPUMilliSeconds / (1000 * 3600))

	// Memory: 1 unit = 1 GB-hour
	units += uint64(record.Metrics.MemoryByteSeconds / (1024 * 1024 * 1024 * 3600))

	// Storage: 1 unit = 1 GB-hour
	units += uint64(record.Metrics.StorageByteSeconds / (1024 * 1024 * 1024 * 3600))

	// GPU: 1 unit = 1 GPU-hour
	units += uint64(record.Metrics.GPUSeconds / 3600)

	// Network: 1 unit = 1 GB
	networkBytes := record.Metrics.NetworkBytesIn + record.Metrics.NetworkBytesOut
	units += uint64(networkBytes / (1024 * 1024 * 1024))

	if units == 0 {
		units = 1 // Minimum 1 unit
	}

	return units
}

// determineUsageType determines the primary usage type.
func (p *SettlementPipeline) determineUsageType(record *UsageRecord) string {
	// Return the type with highest usage
	maxUsage := record.Metrics.CPUMilliSeconds
	usageType := "compute"

	if record.Metrics.GPUSeconds*1000 > maxUsage {
		maxUsage = record.Metrics.GPUSeconds * 1000
		usageType = "gpu"
	}
	if record.Metrics.StorageByteSeconds/(1024*1024*1024) > maxUsage {
		maxUsage = record.Metrics.StorageByteSeconds / (1024 * 1024 * 1024)
		usageType = "storage"
	}
	networkBytes := record.Metrics.NetworkBytesIn + record.Metrics.NetworkBytesOut
	if networkBytes/(1024*1024) > maxUsage {
		usageType = "network"
	}

	return usageType
}

// getPrimaryRate gets the unit price for the primary usage type.
func (p *SettlementPipeline) getPrimaryRate(record *UsageRecord, usageType string) sdk.DecCoin {
	var rateStr string

	switch usageType {
	case "compute":
		rateStr = record.PricingInputs.AgreedCPURate
	case "gpu":
		rateStr = record.PricingInputs.AgreedGPURate
	case "storage":
		rateStr = record.PricingInputs.AgreedStorageRate
	case "network":
		rateStr = record.PricingInputs.AgreedNetworkRate
	default:
		rateStr = record.PricingInputs.AgreedCPURate
	}

	if rateStr == "" {
		return sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyZeroDec())
	}

	rate, err := sdkmath.LegacyNewDecFromStr(rateStr)
	if err != nil {
		return sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyZeroDec())
	}

	return sdk.NewDecCoinFromDec("uvirt", rate)
}

// runLoop runs the settlement pipeline loop.
func (p *SettlementPipeline) runLoop(ctx context.Context) {
	settleTicker := time.NewTicker(p.cfg.SettlementInterval)
	defer settleTicker.Stop()

	disputeTicker := time.NewTicker(time.Hour)
	defer disputeTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopChan:
			return
		case <-settleTicker.C:
			p.processSettlements(ctx)
		case <-disputeTicker.C:
			p.processExpiredDisputes()
		}
	}
}

// processSettlements processes pending settlements.
func (p *SettlementPipeline) processSettlements(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.pending) == 0 {
		return
	}

	// Group pending records by order
	byOrder := make(map[string][]*UsageRecord)
	for _, record := range p.pending {
		orderID := record.DeploymentID
		byOrder[orderID] = append(byOrder[orderID], record)
	}

	// Process each order
	for orderID, records := range byOrder {
		// Skip if any records are disputed
		hasDispute := false
		for _, record := range records {
			for _, dispute := range p.disputes {
				if dispute.UsageRecordID == record.ID && 
					(dispute.Status == DisputeStatusPending || dispute.Status == DisputeStatusReviewing) {
					hasDispute = true
					break
				}
			}
			if hasDispute {
				break
			}
		}

		if hasDispute {
			log.Printf("[settlement-pipeline] skipping order %s with pending dispute", orderID)
			continue
		}

		// Submit settlement request
		usageIDs := make([]string, len(records))
		for i, r := range records {
			usageIDs[i] = r.ID
		}

		if p.chainSubmit != nil {
			if err := p.chainSubmit.SubmitSettlementRequest(ctx, orderID, usageIDs, false); err != nil {
				log.Printf("[settlement-pipeline] failed to submit settlement for order %s: %v", orderID, err)
				continue
			}
		}

		// Remove settled records from pending
		for _, r := range records {
			delete(p.pending, r.ID)
		}

		log.Printf("[settlement-pipeline] settlement submitted for order %s with %d records", orderID, len(records))
	}
}

// processExpiredDisputes handles expired disputes.
func (p *SettlementPipeline) processExpiredDisputes() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for _, dispute := range p.disputes {
		if dispute.Status == DisputeStatusPending && now.After(dispute.ExpiresAt) {
			dispute.Status = DisputeStatusExpired
			dispute.Resolution = "Dispute window expired"
			dispute.ResolvedAt = &now
		}
	}
}

// generateID generates a unique ID with prefix.
func (p *SettlementPipeline) generateID(prefix string, timestamp time.Time) string {
	data := prefix + ":" + timestamp.Format(time.RFC3339Nano)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}
