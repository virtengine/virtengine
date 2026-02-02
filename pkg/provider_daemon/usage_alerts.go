// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-5C: Usage anomaly detection metrics and alerting
package provider_daemon

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	verrors "github.com/virtengine/virtengine/pkg/errors"
)

// AlertSeverity represents the severity of an alert.
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertType represents the type of alert.
type AlertType string

const (
	AlertTypeUsageAnomaly        AlertType = "usage_anomaly"
	AlertTypeReconciliationError AlertType = "reconciliation_error"
	AlertTypeSubmissionFailure   AlertType = "submission_failure"
	AlertTypeDisputeCreated      AlertType = "dispute_created"
	AlertTypeSettlementFailure   AlertType = "settlement_failure"
	AlertTypeFraudSuspected      AlertType = "fraud_suspected"
	AlertTypeThresholdExceeded   AlertType = "threshold_exceeded"
	AlertTypeServiceDegraded     AlertType = "service_degraded"
)

// Alert represents a usage reporting alert.
type Alert struct {
	// AlertID is the unique identifier for this alert.
	AlertID string `json:"alert_id"`

	// Type is the alert type.
	Type AlertType `json:"type"`

	// Severity is the alert severity.
	Severity AlertSeverity `json:"severity"`

	// Message is the alert message.
	Message string `json:"message"`

	// Details contains additional alert details.
	Details map[string]string `json:"details,omitempty"`

	// AllocationID is the affected allocation (if any).
	AllocationID string `json:"allocation_id,omitempty"`

	// OrderID is the affected order (if any).
	OrderID string `json:"order_id,omitempty"`

	// CreatedAt is when the alert was created.
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the alert expires.
	ExpiresAt time.Time `json:"expires_at"`

	// Acknowledged indicates if the alert was acknowledged.
	Acknowledged bool `json:"acknowledged"`

	// AcknowledgedAt is when the alert was acknowledged.
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`

	// AcknowledgedBy is who acknowledged the alert.
	AcknowledgedBy string `json:"acknowledged_by,omitempty"`
}

// AlertHandler handles alert notifications.
type AlertHandler interface {
	// HandleAlert handles an alert notification.
	HandleAlert(ctx context.Context, alert *Alert) error
}

// LogAlertHandler logs alerts.
type LogAlertHandler struct{}

// HandleAlert logs the alert.
func (h *LogAlertHandler) HandleAlert(_ context.Context, alert *Alert) error {
	log.Printf("[ALERT] [%s] [%s] %s (allocation=%s, order=%s)",
		alert.Severity, alert.Type, alert.Message, alert.AllocationID, alert.OrderID)
	return nil
}

// AlertManagerConfig configures the alert manager.
type AlertManagerConfig struct {
	// Enabled enables alerting.
	Enabled bool

	// DefaultAlertTTL is the default TTL for alerts.
	DefaultAlertTTL time.Duration

	// MaxAlerts is the maximum number of alerts to store.
	MaxAlerts int

	// DeduplicationWindow is the window for deduplicating alerts.
	DeduplicationWindow time.Duration

	// ThrottleInterval is the minimum interval between duplicate alerts.
	ThrottleInterval time.Duration
}

// DefaultAlertManagerConfig returns default alert manager config.
func DefaultAlertManagerConfig() AlertManagerConfig {
	return AlertManagerConfig{
		Enabled:             true,
		DefaultAlertTTL:     24 * time.Hour,
		MaxAlerts:           10000,
		DeduplicationWindow: 5 * time.Minute,
		ThrottleInterval:    time.Minute,
	}
}

// UsageAlertManager manages usage-related alerts.
type UsageAlertManager struct {
	mu sync.RWMutex

	cfg      AlertManagerConfig
	handlers []AlertHandler

	// alerts contains all alerts indexed by ID.
	alerts map[string]*Alert

	// alertsByType contains alert IDs indexed by type.
	alertsByType map[AlertType][]string

	// recentAlertKeys tracks recent alert keys for deduplication.
	recentAlertKeys map[string]time.Time

	// counters for metrics.
	totalCreated      int64
	totalAcknowledged int64
	totalExpired      int64

	// running indicates if cleanup is running.
	running  bool
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewUsageAlertManager creates a new alert manager.
func NewUsageAlertManager(cfg AlertManagerConfig) *UsageAlertManager {
	return &UsageAlertManager{
		cfg:             cfg,
		handlers:        []AlertHandler{&LogAlertHandler{}},
		alerts:          make(map[string]*Alert),
		alertsByType:    make(map[AlertType][]string),
		recentAlertKeys: make(map[string]time.Time),
		stopChan:        make(chan struct{}),
	}
}

// Start starts the alert manager cleanup loop.
func (m *UsageAlertManager) Start(ctx context.Context) error {
	if !m.cfg.Enabled {
		return nil
	}

	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = true
	m.mu.Unlock()

	m.wg.Add(1)
	verrors.SafeGo("provider-daemon:alert-manager", func() {
		defer m.wg.Done()
		m.cleanupLoop(ctx)
	})

	log.Printf("[alert-manager] started")
	return nil
}

// Stop stops the alert manager.
func (m *UsageAlertManager) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	m.running = false
	m.mu.Unlock()

	close(m.stopChan)
	m.wg.Wait()

	m.stopChan = make(chan struct{})
	log.Printf("[alert-manager] stopped")
}

// AddHandler adds an alert handler.
func (m *UsageAlertManager) AddHandler(handler AlertHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, handler)
}

// CreateAlert creates a new alert.
func (m *UsageAlertManager) CreateAlert(ctx context.Context, alertType AlertType, severity AlertSeverity, message string, allocationID, orderID string, details map[string]string) (*Alert, error) {
	if !m.cfg.Enabled {
		return nil, nil
	}

	// Check for duplicate
	alertKey := m.generateAlertKey(alertType, allocationID, orderID)
	if m.isDuplicate(alertKey) {
		return nil, nil
	}

	now := time.Now()
	alertID := m.generateAlertID(alertType, now)

	alert := &Alert{
		AlertID:      alertID,
		Type:         alertType,
		Severity:     severity,
		Message:      message,
		Details:      details,
		AllocationID: allocationID,
		OrderID:      orderID,
		CreatedAt:    now,
		ExpiresAt:    now.Add(m.cfg.DefaultAlertTTL),
	}

	m.mu.Lock()
	m.alerts[alertID] = alert
	m.alertsByType[alertType] = append(m.alertsByType[alertType], alertID)
	m.recentAlertKeys[alertKey] = now
	atomic.AddInt64(&m.totalCreated, 1)

	// Enforce max alerts
	if len(m.alerts) > m.cfg.MaxAlerts {
		m.pruneOldestAlerts()
	}
	m.mu.Unlock()

	// Notify handlers
	for _, handler := range m.handlers {
		if err := handler.HandleAlert(ctx, alert); err != nil {
			log.Printf("[alert-manager] handler error: %v", err)
		}
	}

	return alert, nil
}

// CreateUsageAnomalyAlert creates an alert for a usage anomaly.
func (m *UsageAlertManager) CreateUsageAnomalyAlert(ctx context.Context, anomaly *UsageAnomaly) (*Alert, error) {
	severity := m.mapAnomalySeverity(anomaly.Severity)
	details := map[string]string{
		"anomaly_id":     anomaly.AnomalyID,
		"anomaly_type":   anomaly.AnomalyType,
		"value":          formatFloatAlert(anomaly.Value),
		"expected_range": anomaly.ExpectedRange,
	}

	return m.CreateAlert(ctx, AlertTypeUsageAnomaly, severity, anomaly.Description, "", anomaly.OrderID, details)
}

// CreateReconciliationAlert creates an alert for reconciliation issues.
func (m *UsageAlertManager) CreateReconciliationAlert(ctx context.Context, result *ReconciliationResult) (*Alert, error) {
	if result.InSync || len(result.Discrepancies) == 0 {
		return nil, nil
	}

	// Determine severity based on score
	var severity AlertSeverity
	switch {
	case result.Score < 30:
		severity = AlertSeverityCritical
	case result.Score < 50:
		severity = AlertSeverityError
	case result.Score < 70:
		severity = AlertSeverityWarning
	default:
		severity = AlertSeverityInfo
	}

	message := "Reconciliation discrepancies detected"
	details := map[string]string{
		"score":             formatIntAlert(result.Score),
		"discrepancy_count": formatIntAlert(len(result.Discrepancies)),
	}

	// Add top discrepancies to details
	for i, d := range result.Discrepancies {
		if i >= 3 {
			break
		}
		key := "discrepancy_" + formatIntAlert(i+1)
		details[key] = d.MetricName + ": " + formatFloatAlert(d.DifferencePercent) + "%"
	}

	return m.CreateAlert(ctx, AlertTypeReconciliationError, severity, message, result.AllocationID, "", details)
}

// CreateDisputeAlert creates an alert for a new dispute.
func (m *UsageAlertManager) CreateDisputeAlert(ctx context.Context, dispute *UsageDispute) (*Alert, error) {
	message := "Usage dispute created: " + dispute.Reason
	details := map[string]string{
		"dispute_id":      dispute.DisputeID,
		"usage_record_id": dispute.UsageRecordID,
		"initiator":       dispute.Initiator,
		"expires_at":      dispute.ExpiresAt.Format(time.RFC3339),
	}

	return m.CreateAlert(ctx, AlertTypeDisputeCreated, AlertSeverityWarning, message, "", dispute.OrderID, details)
}

// CreateSubmissionFailureAlert creates an alert for submission failures.
func (m *UsageAlertManager) CreateSubmissionFailureAlert(ctx context.Context, orderID string, errorMsg string) (*Alert, error) {
	message := "Failed to submit usage to chain: " + errorMsg
	details := map[string]string{
		"error": errorMsg,
	}

	return m.CreateAlert(ctx, AlertTypeSubmissionFailure, AlertSeverityError, message, "", orderID, details)
}

// CreateFraudAlert creates an alert for suspected fraud.
func (m *UsageAlertManager) CreateFraudAlert(ctx context.Context, record *UsageRecord, fraudResult *FraudCheckResult) (*Alert, error) {
	message := "Fraud suspected in usage record"
	details := map[string]string{
		"usage_record_id": record.ID,
		"fraud_score":     formatIntAlert(fraudResult.Score),
		"flags":           formatFlags(fraudResult.Flags),
	}

	severity := AlertSeverityWarning
	if fraudResult.Score >= 80 {
		severity = AlertSeverityCritical
	} else if fraudResult.Score >= 50 {
		severity = AlertSeverityError
	}

	return m.CreateAlert(ctx, AlertTypeFraudSuspected, severity, message, "", record.DeploymentID, details)
}

// AcknowledgeAlert acknowledges an alert.
func (m *UsageAlertManager) AcknowledgeAlert(alertID string, acknowledgedBy string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	alert, ok := m.alerts[alertID]
	if !ok {
		return ErrAlertNotFound
	}

	if alert.Acknowledged {
		return nil
	}

	now := time.Now()
	alert.Acknowledged = true
	alert.AcknowledgedAt = &now
	alert.AcknowledgedBy = acknowledgedBy
	atomic.AddInt64(&m.totalAcknowledged, 1)

	return nil
}

// GetAlert gets an alert by ID.
func (m *UsageAlertManager) GetAlert(alertID string) (*Alert, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	alert, ok := m.alerts[alertID]
	return alert, ok
}

// GetAlertsByType gets alerts by type.
func (m *UsageAlertManager) GetAlertsByType(alertType AlertType) []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := m.alertsByType[alertType]
	alerts := make([]*Alert, 0, len(ids))
	for _, id := range ids {
		if alert, ok := m.alerts[id]; ok {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

// GetActiveAlerts gets all active (unacknowledged) alerts.
func (m *UsageAlertManager) GetActiveAlerts() []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	active := make([]*Alert, 0)
	for _, alert := range m.alerts {
		if !alert.Acknowledged && now.Before(alert.ExpiresAt) {
			active = append(active, alert)
		}
	}
	return active
}

// GetAlertCounts gets alert counts by severity.
func (m *UsageAlertManager) GetAlertCounts() map[AlertSeverity]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	counts := make(map[AlertSeverity]int)
	for _, alert := range m.alerts {
		if !alert.Acknowledged && now.Before(alert.ExpiresAt) {
			counts[alert.Severity]++
		}
	}
	return counts
}

// GetMetrics returns alert metrics.
func (m *UsageAlertManager) GetMetrics() AlertMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	metrics := AlertMetrics{
		TotalCreated:      atomic.LoadInt64(&m.totalCreated),
		TotalAcknowledged: atomic.LoadInt64(&m.totalAcknowledged),
		TotalExpired:      atomic.LoadInt64(&m.totalExpired),
	}

	for _, alert := range m.alerts {
		if !alert.Acknowledged && now.Before(alert.ExpiresAt) {
			metrics.ActiveCount++
			switch alert.Severity {
			case AlertSeverityCritical:
				metrics.CriticalCount++
			case AlertSeverityError:
				metrics.ErrorCount++
			case AlertSeverityWarning:
				metrics.WarningCount++
			case AlertSeverityInfo:
				metrics.InfoCount++
			}
		}
	}

	return metrics
}

// AlertMetrics contains alert metrics.
type AlertMetrics struct {
	// TotalCreated is the total number of alerts created.
	TotalCreated int64 `json:"total_created"`

	// TotalAcknowledged is the total number of acknowledged alerts.
	TotalAcknowledged int64 `json:"total_acknowledged"`

	// TotalExpired is the total number of expired alerts.
	TotalExpired int64 `json:"total_expired"`

	// ActiveCount is the number of active alerts.
	ActiveCount int `json:"active_count"`

	// CriticalCount is the number of critical alerts.
	CriticalCount int `json:"critical_count"`

	// ErrorCount is the number of error alerts.
	ErrorCount int `json:"error_count"`

	// WarningCount is the number of warning alerts.
	WarningCount int `json:"warning_count"`

	// InfoCount is the number of info alerts.
	InfoCount int `json:"info_count"`
}

// generateAlertID generates a unique alert ID.
func (m *UsageAlertManager) generateAlertID(alertType AlertType, timestamp time.Time) string {
	return string(alertType) + "-" + timestamp.Format("20060102150405.000000000")
}

// generateAlertKey generates a key for deduplication.
func (m *UsageAlertManager) generateAlertKey(alertType AlertType, allocationID, orderID string) string {
	return string(alertType) + ":" + allocationID + ":" + orderID
}

// isDuplicate checks if an alert is a duplicate.
func (m *UsageAlertManager) isDuplicate(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lastTime, ok := m.recentAlertKeys[key]
	if !ok {
		return false
	}

	return time.Since(lastTime) < m.cfg.ThrottleInterval
}

// mapAnomalySeverity maps anomaly severity to alert severity.
func (m *UsageAlertManager) mapAnomalySeverity(severity string) AlertSeverity {
	switch severity {
	case "critical":
		return AlertSeverityCritical
	case "high":
		return AlertSeverityError
	case "medium":
		return AlertSeverityWarning
	default:
		return AlertSeverityInfo
	}
}

// pruneOldestAlerts removes oldest alerts when max is exceeded.
func (m *UsageAlertManager) pruneOldestAlerts() {
	// Find oldest alerts
	var oldest *Alert
	for _, alert := range m.alerts {
		if oldest == nil || alert.CreatedAt.Before(oldest.CreatedAt) {
			oldest = alert
		}
	}

	if oldest != nil {
		delete(m.alerts, oldest.AlertID)
		atomic.AddInt64(&m.totalExpired, 1)
	}
}

// cleanupLoop runs the cleanup loop.
func (m *UsageAlertManager) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.cleanup()
		}
	}
}

// cleanup removes expired alerts and old dedup keys.
func (m *UsageAlertManager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// Remove expired alerts
	for id, alert := range m.alerts {
		if now.After(alert.ExpiresAt) {
			delete(m.alerts, id)
			atomic.AddInt64(&m.totalExpired, 1)
		}
	}

	// Remove old dedup keys
	for key, timestamp := range m.recentAlertKeys {
		if now.Sub(timestamp) > m.cfg.DeduplicationWindow {
			delete(m.recentAlertKeys, key)
		}
	}

	// Rebuild alertsByType index
	m.alertsByType = make(map[AlertType][]string)
	for id, alert := range m.alerts {
		m.alertsByType[alert.Type] = append(m.alertsByType[alert.Type], id)
	}
}

// ErrAlertNotFound is returned when an alert is not found.
var ErrAlertNotFound = verrors.ErrNotFound

// Helper functions for formatting.
func formatFloatAlert(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func formatIntAlert(v int) string {
	return fmt.Sprintf("%d", v)
}

func formatFlags(flags []string) string {
	if len(flags) == 0 {
		return ""
	}
	result := flags[0]
	for i := 1; i < len(flags); i++ {
		result += ", " + flags[i]
	}
	return result
}

// UsageMetricsCollector collects usage reporting metrics.
type UsageMetricsCollector struct {
	mu sync.RWMutex

	// Counters
	recordsCollected   int64
	recordsSubmitted   int64
	submissionFailures int64
	settlementsSuccess int64
	settlementsFailure int64
	disputesCreated    int64
	disputesResolved   int64
	anomaliesDetected  int64
	correctionsApplied int64

	// Gauges
	pendingRecords      int64
	activeDisputes      int64
	activeAnomalies     int64
	reconciliationScore int64

	// Histograms (simplified as averages)
	avgCollectionTime float64
	avgSubmissionTime float64
	avgSettlementTime float64

	// Timestamps
	lastCollection time.Time
	lastSubmission time.Time
	lastSettlement time.Time
}

// NewUsageMetricsCollector creates a new metrics collector.
func NewUsageMetricsCollector() *UsageMetricsCollector {
	return &UsageMetricsCollector{}
}

// RecordCollection records a usage collection.
func (c *UsageMetricsCollector) RecordCollection(duration time.Duration) {
	atomic.AddInt64(&c.recordsCollected, 1)
	c.mu.Lock()
	c.lastCollection = time.Now()
	c.avgCollectionTime = (c.avgCollectionTime + duration.Seconds()) / 2
	c.mu.Unlock()
}

// RecordSubmission records a chain submission.
func (c *UsageMetricsCollector) RecordSubmission(success bool, duration time.Duration) {
	if success {
		atomic.AddInt64(&c.recordsSubmitted, 1)
	} else {
		atomic.AddInt64(&c.submissionFailures, 1)
	}
	c.mu.Lock()
	c.lastSubmission = time.Now()
	c.avgSubmissionTime = (c.avgSubmissionTime + duration.Seconds()) / 2
	c.mu.Unlock()
}

// RecordSettlement records a settlement.
func (c *UsageMetricsCollector) RecordSettlement(success bool, duration time.Duration) {
	if success {
		atomic.AddInt64(&c.settlementsSuccess, 1)
	} else {
		atomic.AddInt64(&c.settlementsFailure, 1)
	}
	c.mu.Lock()
	c.lastSettlement = time.Now()
	c.avgSettlementTime = (c.avgSettlementTime + duration.Seconds()) / 2
	c.mu.Unlock()
}

// RecordDispute records a dispute.
func (c *UsageMetricsCollector) RecordDispute(created bool) {
	if created {
		atomic.AddInt64(&c.disputesCreated, 1)
	} else {
		atomic.AddInt64(&c.disputesResolved, 1)
	}
}

// RecordAnomaly records an anomaly detection.
func (c *UsageMetricsCollector) RecordAnomaly() {
	atomic.AddInt64(&c.anomaliesDetected, 1)
}

// RecordCorrection records a correction.
func (c *UsageMetricsCollector) RecordCorrection() {
	atomic.AddInt64(&c.correctionsApplied, 1)
}

// SetPendingRecords sets the pending records gauge.
func (c *UsageMetricsCollector) SetPendingRecords(count int64) {
	atomic.StoreInt64(&c.pendingRecords, count)
}

// SetActiveDisputes sets the active disputes gauge.
func (c *UsageMetricsCollector) SetActiveDisputes(count int64) {
	atomic.StoreInt64(&c.activeDisputes, count)
}

// SetActiveAnomalies sets the active anomalies gauge.
func (c *UsageMetricsCollector) SetActiveAnomalies(count int64) {
	atomic.StoreInt64(&c.activeAnomalies, count)
}

// SetReconciliationScore sets the reconciliation score gauge.
func (c *UsageMetricsCollector) SetReconciliationScore(score int64) {
	atomic.StoreInt64(&c.reconciliationScore, score)
}

// GetMetrics returns current metrics.
func (c *UsageMetricsCollector) GetMetrics() UsageReportingMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return UsageReportingMetrics{
		TotalRecordsCollected:      atomic.LoadInt64(&c.recordsCollected),
		TotalRecordsSubmitted:      atomic.LoadInt64(&c.recordsSubmitted),
		TotalSettlementsProcessed:  atomic.LoadInt64(&c.settlementsSuccess),
		TotalDisputesCreated:       atomic.LoadInt64(&c.disputesCreated),
		TotalDisputesResolved:      atomic.LoadInt64(&c.disputesResolved),
		TotalAnomaliesDetected:     atomic.LoadInt64(&c.anomaliesDetected),
		TotalCorrectionsApplied:    atomic.LoadInt64(&c.correctionsApplied),
		LastCollectionTime:         c.lastCollection,
		LastSubmissionTime:         c.lastSubmission,
		LastSettlementTime:         c.lastSettlement,
		AverageReconciliationScore: int(atomic.LoadInt64(&c.reconciliationScore)),
	}
}

// GetCounters returns counter values.
func (c *UsageMetricsCollector) GetCounters() map[string]int64 {
	return map[string]int64{
		"records_collected":   atomic.LoadInt64(&c.recordsCollected),
		"records_submitted":   atomic.LoadInt64(&c.recordsSubmitted),
		"submission_failures": atomic.LoadInt64(&c.submissionFailures),
		"settlements_success": atomic.LoadInt64(&c.settlementsSuccess),
		"settlements_failure": atomic.LoadInt64(&c.settlementsFailure),
		"disputes_created":    atomic.LoadInt64(&c.disputesCreated),
		"disputes_resolved":   atomic.LoadInt64(&c.disputesResolved),
		"anomalies_detected":  atomic.LoadInt64(&c.anomaliesDetected),
		"corrections_applied": atomic.LoadInt64(&c.correctionsApplied),
	}
}

// GetGauges returns gauge values.
func (c *UsageMetricsCollector) GetGauges() map[string]int64 {
	return map[string]int64{
		"pending_records":      atomic.LoadInt64(&c.pendingRecords),
		"active_disputes":      atomic.LoadInt64(&c.activeDisputes),
		"active_anomalies":     atomic.LoadInt64(&c.activeAnomalies),
		"reconciliation_score": atomic.LoadInt64(&c.reconciliationScore),
	}
}

// GetAverages returns average values.
func (c *UsageMetricsCollector) GetAverages() map[string]float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]float64{
		"avg_collection_time_seconds": c.avgCollectionTime,
		"avg_submission_time_seconds": c.avgSubmissionTime,
		"avg_settlement_time_seconds": c.avgSettlementTime,
	}
}
