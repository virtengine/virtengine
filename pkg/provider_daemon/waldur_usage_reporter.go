// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-21D: Waldur usage reporter for automated marketplace integration
// This file implements the usage reporter that submits usage data to Waldur.
package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/virtengine/virtengine/pkg/waldur"
)

// WaldurUsageReporterConfig configures the Waldur usage reporter
type WaldurUsageReporterConfig struct {
	// Enabled indicates if usage reporting is enabled
	Enabled bool `json:"enabled"`

	// ProviderAddress is the provider's blockchain address
	ProviderAddress string `json:"provider_address"`

	// ReportIntervalSeconds is how often to submit usage reports
	ReportIntervalSeconds int64 `json:"report_interval_seconds"`

	// BatchSize is the maximum number of reports to submit in one batch
	BatchSize int `json:"batch_size"`

	// MaxRetries is the maximum number of retries for failed reports
	MaxRetries int `json:"max_retries"`

	// RetryBackoffSeconds is the base backoff duration between retries
	RetryBackoffSeconds int64 `json:"retry_backoff_seconds"`

	// StateFilePath is the path to persist reporter state
	StateFilePath string `json:"state_file_path"`

	// OperationTimeout is the timeout for Waldur API calls
	OperationTimeout time.Duration `json:"operation_timeout"`

	// EnableAuditLogging enables audit logging for usage reports
	EnableAuditLogging bool `json:"enable_audit_logging"`
}

// DefaultWaldurUsageReporterConfig returns default configuration
func DefaultWaldurUsageReporterConfig() WaldurUsageReporterConfig {
	return WaldurUsageReporterConfig{
		Enabled:               true,
		ReportIntervalSeconds: 3600, // 1 hour
		BatchSize:             50,
		MaxRetries:            3,
		RetryBackoffSeconds:   60,
		StateFilePath:         "data/waldur_usage_reporter_state.json",
		OperationTimeout:      60 * time.Second,
		EnableAuditLogging:    true,
	}
}

// WaldurUsageReporterState persists reporter state
type WaldurUsageReporterState struct {
	// LastReportTime is when the last report was submitted
	LastReportTime time.Time `json:"last_report_time"`

	// PendingReports contains reports that need to be submitted
	PendingReports map[string]*PendingUsageReport `json:"pending_reports"`

	// FailedReports contains reports that failed to submit
	FailedReports map[string]*FailedUsageReport `json:"failed_reports"`

	// ReportedPeriods tracks which periods have been reported for each resource
	ReportedPeriods map[string][]ReportedPeriod `json:"reported_periods"`

	// Metrics contains usage reporting metrics
	Metrics *WaldurUsageReportingMetrics `json:"metrics"`

	// LastUpdated is when state was last updated
	LastUpdated time.Time `json:"last_updated"`
}

// PendingUsageReport represents a report waiting to be submitted
type PendingUsageReport struct {
	ID            string                      `json:"id"`
	AllocationID  string                      `json:"allocation_id"`
	ResourceUUID  string                      `json:"resource_uuid"`
	Report        *waldur.ResourceUsageReport `json:"report"`
	CreatedAt     time.Time                   `json:"created_at"`
	ScheduledAt   time.Time                   `json:"scheduled_at"`
	AttemptCount  int                         `json:"attempt_count"`
	LastAttemptAt *time.Time                  `json:"last_attempt_at,omitempty"`
	LastError     string                      `json:"last_error,omitempty"`
}

// FailedUsageReport represents a report that failed to submit
type FailedUsageReport struct {
	ID            string                      `json:"id"`
	AllocationID  string                      `json:"allocation_id"`
	ResourceUUID  string                      `json:"resource_uuid"`
	Report        *waldur.ResourceUsageReport `json:"report"`
	FailedAt      time.Time                   `json:"failed_at"`
	TotalAttempts int                         `json:"total_attempts"`
	LastError     string                      `json:"last_error"`
	DeadLettered  bool                        `json:"dead_lettered"`
}

// ReportedPeriod tracks a period that has been reported
type ReportedPeriod struct {
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	ReportedAt  time.Time `json:"reported_at"`
	WaldurUUID  string    `json:"waldur_uuid,omitempty"`
}

// WaldurUsageReportingMetrics tracks usage reporting metrics for Waldur submissions
type WaldurUsageReportingMetrics struct {
	TotalReports        int64     `json:"total_reports"`
	SuccessfulReports   int64     `json:"successful_reports"`
	FailedReports       int64     `json:"failed_reports"`
	RetryAttempts       int64     `json:"retry_attempts"`
	DeadLetteredReports int64     `json:"dead_lettered_reports"`
	LastSuccessTime     time.Time `json:"last_success_time"`
	LastFailureTime     time.Time `json:"last_failure_time"`
	AverageLatencyMs    int64     `json:"average_latency_ms"`
}

// WaldurUsageReporter submits usage data to Waldur
type WaldurUsageReporter struct {
	cfg         WaldurUsageReporterConfig
	usageClient *waldur.UsageClient
	usageStore  *UsageSnapshotStore
	bridgeState *WaldurBridgeStateStore
	state       *WaldurUsageReporterState
	auditLogger *AuditLogger
	mu          sync.RWMutex
	stopCh      chan struct{}
	doneCh      chan struct{}
	running     bool

	// Prometheus-compatible metrics
	promMetrics *UsageReporterPrometheusMetrics
}

// UsageReporterPrometheusMetrics provides Prometheus-compatible metrics
type UsageReporterPrometheusMetrics struct {
	ReportsTotal        atomic.Int64
	ReportsSuccessful   atomic.Int64
	ReportsFailed       atomic.Int64
	ReportsDeadLettered atomic.Int64
	RetryAttempts       atomic.Int64
	PendingReports      atomic.Int64
	ReportDurationSum   atomic.Int64
	ReportDurationCount atomic.Int64
}

// NewWaldurUsageReporter creates a new usage reporter
func NewWaldurUsageReporter(
	cfg WaldurUsageReporterConfig,
	marketplace *waldur.MarketplaceClient,
	usageStore *UsageSnapshotStore,
	bridgeState *WaldurBridgeStateStore,
	auditLogger *AuditLogger,
) (*WaldurUsageReporter, error) {
	if marketplace == nil {
		return nil, fmt.Errorf("marketplace client is required")
	}

	reporter := &WaldurUsageReporter{
		cfg:         cfg,
		usageClient: waldur.NewUsageClient(marketplace),
		usageStore:  usageStore,
		bridgeState: bridgeState,
		auditLogger: auditLogger,
		stopCh:      make(chan struct{}),
		doneCh:      make(chan struct{}),
		promMetrics: &UsageReporterPrometheusMetrics{},
		state: &WaldurUsageReporterState{
			PendingReports:  make(map[string]*PendingUsageReport),
			FailedReports:   make(map[string]*FailedUsageReport),
			ReportedPeriods: make(map[string][]ReportedPeriod),
			Metrics:         &WaldurUsageReportingMetrics{},
		},
	}

	// Load persisted state
	if err := reporter.loadState(); err != nil {
		log.Printf("[waldur-usage-reporter] failed to load state: %v", err)
	}

	return reporter, nil
}

// Start starts the usage reporter
func (r *WaldurUsageReporter) Start(ctx context.Context) error {
	if !r.cfg.Enabled {
		return nil
	}

	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return fmt.Errorf("reporter already running")
	}
	r.running = true
	r.mu.Unlock()

	log.Printf("[waldur-usage-reporter] started for provider %s", r.cfg.ProviderAddress)

	// Start background workers
	go r.reportLoop(ctx)
	go r.retryLoop(ctx)

	return nil
}

// Stop stops the usage reporter
func (r *WaldurUsageReporter) Stop() error {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return nil
	}
	r.running = false
	r.mu.Unlock()

	close(r.stopCh)

	// Wait for workers to finish
	select {
	case <-r.doneCh:
	case <-time.After(10 * time.Second):
		log.Printf("[waldur-usage-reporter] shutdown timeout")
	}

	// Save state
	if err := r.saveState(); err != nil {
		log.Printf("[waldur-usage-reporter] failed to save state: %v", err)
	}

	log.Printf("[waldur-usage-reporter] stopped")
	return nil
}

// QueueUsageReport queues a usage report for submission
func (r *WaldurUsageReporter) QueueUsageReport(
	allocationID string,
	resourceUUID string,
	periodStart, periodEnd time.Time,
	metrics *ResourceMetrics,
) error {
	if allocationID == "" || resourceUUID == "" {
		return fmt.Errorf("allocation ID and resource UUID are required")
	}
	if metrics == nil {
		return fmt.Errorf("metrics are required")
	}

	// Check if this period has already been reported
	r.mu.RLock()
	periods := r.state.ReportedPeriods[allocationID]
	for _, period := range periods {
		if period.PeriodStart.Equal(periodStart) && period.PeriodEnd.Equal(periodEnd) {
			r.mu.RUnlock()
			log.Printf("[waldur-usage-reporter] period already reported for %s", allocationID)
			return nil
		}
	}
	r.mu.RUnlock()

	// Create usage report
	report := r.buildUsageReport(resourceUUID, periodStart, periodEnd, allocationID, metrics)

	// Create pending report
	reportID := fmt.Sprintf("%s_%d_%d", allocationID, periodStart.Unix(), periodEnd.Unix())
	pending := &PendingUsageReport{
		ID:           reportID,
		AllocationID: allocationID,
		ResourceUUID: resourceUUID,
		Report:       report,
		CreatedAt:    time.Now().UTC(),
		ScheduledAt:  time.Now().UTC(),
	}

	r.mu.Lock()
	r.state.PendingReports[reportID] = pending
	r.mu.Unlock()

	r.promMetrics.PendingReports.Add(1)

	log.Printf("[waldur-usage-reporter] queued usage report for %s (period: %s to %s)",
		allocationID, periodStart.Format(time.RFC3339), periodEnd.Format(time.RFC3339))

	return nil
}

// buildUsageReport builds a Waldur usage report from metrics
func (r *WaldurUsageReporter) buildUsageReport(
	resourceUUID string,
	periodStart, periodEnd time.Time,
	backendID string,
	metrics *ResourceMetrics,
) *waldur.ResourceUsageReport {
	components := make([]waldur.ComponentUsage, 0)

	// Convert metrics to Waldur component usages
	cpuHours := hoursFromMilliSeconds(metrics.CPUMilliSeconds)
	if cpuHours > 0 {
		components = append(components, waldur.ComponentUsage{
			Type:        "cpu_hours",
			Amount:      cpuHours,
			Description: "CPU usage in core-hours",
		})
	}

	gpuHours := hoursFromSeconds(metrics.GPUSeconds)
	if gpuHours > 0 {
		components = append(components, waldur.ComponentUsage{
			Type:        "gpu_hours",
			Amount:      gpuHours,
			Description: "GPU usage in GPU-hours",
		})
	}

	ramGBHours := gbHoursFromByteSeconds(metrics.MemoryByteSeconds)
	if ramGBHours > 0 {
		components = append(components, waldur.ComponentUsage{
			Type:        "ram_gb_hours",
			Amount:      ramGBHours,
			Description: "Memory usage in GB-hours",
		})
	}

	storageGBHours := gbHoursFromByteSeconds(metrics.StorageByteSeconds)
	if storageGBHours > 0 {
		components = append(components, waldur.ComponentUsage{
			Type:        "storage_gb_hours",
			Amount:      storageGBHours,
			Description: "Storage usage in GB-hours",
		})
	}

	networkGB := gbFromBytes(metrics.NetworkBytesIn + metrics.NetworkBytesOut)
	if networkGB > 0 {
		components = append(components, waldur.ComponentUsage{
			Type:        "network_gb",
			Amount:      networkGB,
			Description: "Network transfer in GB",
		})
	}

	return &waldur.ResourceUsageReport{
		ResourceUUID: resourceUUID,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		Components:   components,
		BackendID:    backendID,
		Metadata: map[string]string{
			"provider":        r.cfg.ProviderAddress,
			"allocation_id":   backendID,
			"collection_time": time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// reportLoop periodically submits pending usage reports
func (r *WaldurUsageReporter) reportLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(r.cfg.ReportIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			close(r.doneCh)
			return
		case <-r.stopCh:
			close(r.doneCh)
			return
		case <-ticker.C:
			r.processPendingReports(ctx)
		}
	}
}

// processPendingReports submits pending usage reports
func (r *WaldurUsageReporter) processPendingReports(ctx context.Context) {
	r.mu.RLock()
	pending := make([]*PendingUsageReport, 0, len(r.state.PendingReports))
	for _, report := range r.state.PendingReports {
		if time.Now().After(report.ScheduledAt) {
			pending = append(pending, report)
		}
	}
	r.mu.RUnlock()

	if len(pending) == 0 {
		return
	}

	log.Printf("[waldur-usage-reporter] processing %d pending reports", len(pending))

	submitted := 0
	for i, report := range pending {
		if i >= r.cfg.BatchSize {
			break
		}

		if err := r.submitReport(ctx, report); err != nil {
			log.Printf("[waldur-usage-reporter] failed to submit report %s: %v", report.ID, err)
			continue
		}
		submitted++
	}

	log.Printf("[waldur-usage-reporter] submitted %d/%d reports", submitted, len(pending))
}

// submitReport submits a single usage report to Waldur
func (r *WaldurUsageReporter) submitReport(ctx context.Context, pending *PendingUsageReport) error {
	startTime := time.Now()

	r.promMetrics.ReportsTotal.Add(1)
	r.state.Metrics.TotalReports++

	// Update attempt tracking
	r.mu.Lock()
	pending.AttemptCount++
	now := time.Now().UTC()
	pending.LastAttemptAt = &now
	r.mu.Unlock()

	opCtx, cancel := context.WithTimeout(ctx, r.cfg.OperationTimeout)
	defer cancel()

	// Submit to Waldur
	response, err := r.usageClient.SubmitUsageReport(opCtx, pending.Report)

	duration := time.Since(startTime)
	r.promMetrics.ReportDurationSum.Add(int64(duration))
	r.promMetrics.ReportDurationCount.Add(1)

	if err != nil {
		r.handleSubmitError(pending, err)
		return err
	}

	// Mark as successful
	r.mu.Lock()
	delete(r.state.PendingReports, pending.ID)

	// Track reported period
	r.state.ReportedPeriods[pending.AllocationID] = append(
		r.state.ReportedPeriods[pending.AllocationID],
		ReportedPeriod{
			PeriodStart: pending.Report.PeriodStart,
			PeriodEnd:   pending.Report.PeriodEnd,
			ReportedAt:  time.Now().UTC(),
			WaldurUUID:  response.UUID,
		},
	)

	r.state.Metrics.SuccessfulReports++
	r.state.Metrics.LastSuccessTime = time.Now().UTC()
	r.state.LastReportTime = time.Now().UTC()
	r.mu.Unlock()

	r.promMetrics.ReportsSuccessful.Add(1)
	r.promMetrics.PendingReports.Add(-1)

	// Audit log
	if r.auditLogger != nil && r.cfg.EnableAuditLogging {
		_ = r.auditLogger.Log(&AuditEvent{
			Type:      AuditEventType("usage_report_submitted"),
			Operation: "submit_usage",
			Success:   true,
			Details: map[string]interface{}{
				"allocation_id": pending.AllocationID,
				"resource_uuid": pending.ResourceUUID,
				"period_start":  pending.Report.PeriodStart,
				"period_end":    pending.Report.PeriodEnd,
				"waldur_uuid":   response.UUID,
				"duration_ms":   duration.Milliseconds(),
			},
		})
	}

	log.Printf("[waldur-usage-reporter] submitted usage for %s (response: %s)",
		pending.AllocationID, response.UUID)

	return r.saveState()
}

// handleSubmitError handles a failed submission
func (r *WaldurUsageReporter) handleSubmitError(pending *PendingUsageReport, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pending.LastError = err.Error()
	r.state.Metrics.FailedReports++
	r.state.Metrics.LastFailureTime = time.Now().UTC()

	r.promMetrics.ReportsFailed.Add(1)

	if pending.AttemptCount >= r.cfg.MaxRetries {
		// Move to failed reports
		delete(r.state.PendingReports, pending.ID)
		r.state.FailedReports[pending.ID] = &FailedUsageReport{
			ID:            pending.ID,
			AllocationID:  pending.AllocationID,
			ResourceUUID:  pending.ResourceUUID,
			Report:        pending.Report,
			FailedAt:      time.Now().UTC(),
			TotalAttempts: pending.AttemptCount,
			LastError:     err.Error(),
			DeadLettered:  true,
		}

		r.state.Metrics.DeadLetteredReports++
		r.promMetrics.ReportsDeadLettered.Add(1)
		r.promMetrics.PendingReports.Add(-1)

		log.Printf("[waldur-usage-reporter] dead-lettered report %s after %d attempts",
			pending.ID, pending.AttemptCount)
	} else {
		// Schedule retry with exponential backoff
		// Cap the shift amount to prevent overflow (max 10 retries = 1024x backoff)
		shiftAmount := pending.AttemptCount - 1
		if shiftAmount < 0 {
			shiftAmount = 0
		}
		if shiftAmount > 10 {
			shiftAmount = 10
		}
		backoff := time.Duration(r.cfg.RetryBackoffSeconds) * time.Second * time.Duration(1<<shiftAmount)
		pending.ScheduledAt = time.Now().Add(backoff)
		r.promMetrics.RetryAttempts.Add(1)
		r.state.Metrics.RetryAttempts++

		log.Printf("[waldur-usage-reporter] scheduling retry for %s in %v", pending.ID, backoff)
	}

	_ = r.saveState()
}

// retryLoop handles retrying failed reports
func (r *WaldurUsageReporter) retryLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(r.cfg.RetryBackoffSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.stopCh:
			return
		case <-ticker.C:
			// Retries are handled in processPendingReports based on ScheduledAt
		}
	}
}

// ReprocessFailedReport attempts to reprocess a failed report
func (r *WaldurUsageReporter) ReprocessFailedReport(reportID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	failed, ok := r.state.FailedReports[reportID]
	if !ok {
		return fmt.Errorf("report %s not found in failed reports", reportID)
	}

	// Move back to pending
	pending := &PendingUsageReport{
		ID:           failed.ID,
		AllocationID: failed.AllocationID,
		ResourceUUID: failed.ResourceUUID,
		Report:       failed.Report,
		CreatedAt:    failed.FailedAt,
		ScheduledAt:  time.Now().UTC(),
		AttemptCount: 0, // Reset attempt count
	}

	r.state.PendingReports[reportID] = pending
	delete(r.state.FailedReports, reportID)

	r.promMetrics.ReportsDeadLettered.Add(-1)
	r.promMetrics.PendingReports.Add(1)

	log.Printf("[waldur-usage-reporter] reprocessing failed report %s", reportID)

	return r.saveState()
}

// SubmitFromUsageRecord submits usage from a usage record
func (r *WaldurUsageReporter) SubmitFromUsageRecord(
	ctx context.Context,
	allocationID string,
	resourceUUID string,
	record *UsageRecord,
) error {
	if record == nil {
		return fmt.Errorf("usage record is required")
	}

	return r.QueueUsageReport(
		allocationID,
		resourceUUID,
		record.StartTime,
		record.EndTime,
		&record.Metrics,
	)
}

// GetMetrics returns current metrics
func (r *WaldurUsageReporter) GetMetrics() *WaldurUsageReportingMetrics {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.Metrics
}

// GetPendingCount returns the number of pending reports
func (r *WaldurUsageReporter) GetPendingCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.state.PendingReports)
}

// GetFailedCount returns the number of failed reports
func (r *WaldurUsageReporter) GetFailedCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.state.FailedReports)
}

// PrometheusMetrics returns the prometheus-compatible metrics
func (r *WaldurUsageReporter) PrometheusMetrics() *UsageReporterPrometheusMetrics {
	return r.promMetrics
}

// loadState loads persisted state
func (r *WaldurUsageReporter) loadState() error {
	if r.cfg.StateFilePath == "" {
		return nil
	}

	data, err := os.ReadFile(r.cfg.StateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var state WaldurUsageReporterState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	r.state = &state
	return nil
}

// saveState persists state to disk
func (r *WaldurUsageReporter) saveState() error {
	if r.cfg.StateFilePath == "" {
		return nil
	}

	r.state.LastUpdated = time.Now().UTC()

	data, err := json.MarshalIndent(r.state, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(r.cfg.StateFilePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	tmp := r.cfg.StateFilePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}

	return os.Rename(tmp, r.cfg.StateFilePath)
}
