// Package offramp provides fiat off-ramp integration for token-to-fiat payouts.
//
// VE-5E: Fiat off-ramp integration for PayPal/ACH payouts
package offramp

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Metrics Collector
// ============================================================================

// MetricsCollector collects and exposes off-ramp metrics.
type MetricsCollector struct {
	mu sync.RWMutex

	// Payout metrics
	PayoutsTotal        int64            `json:"payouts_total"`
	PayoutsSucceeded    int64            `json:"payouts_succeeded"`
	PayoutsFailed       int64            `json:"payouts_failed"`
	PayoutsCanceled     int64            `json:"payouts_canceled"`
	PayoutsReversed     int64            `json:"payouts_reversed"`
	PayoutsByProvider   map[ProviderType]int64 `json:"payouts_by_provider"`
	PayoutsByCurrency   map[string]int64 `json:"payouts_by_currency"`

	// Amount metrics
	TotalAmountPaidOut  int64            `json:"total_amount_paid_out"`
	TotalFees           int64            `json:"total_fees"`
	AmountByProvider    map[ProviderType]int64 `json:"amount_by_provider"`

	// Timing metrics
	AvgProcessingTimeMs int64            `json:"avg_processing_time_ms"`
	MaxProcessingTimeMs int64            `json:"max_processing_time_ms"`
	MinProcessingTimeMs int64            `json:"min_processing_time_ms"`

	// Compliance metrics
	KYCRejections       int64            `json:"kyc_rejections"`
	AMLRejections       int64            `json:"aml_rejections"`
	AMLFlagged          int64            `json:"aml_flagged"`

	// Webhook metrics
	WebhooksReceived    int64            `json:"webhooks_received"`
	WebhooksProcessed   int64            `json:"webhooks_processed"`
	WebhooksFailed      int64            `json:"webhooks_failed"`

	// Reconciliation metrics
	ReconciliationsRun  int64            `json:"reconciliations_run"`
	ReconciliationMatches int64          `json:"reconciliation_matches"`
	ReconciliationMismatches int64       `json:"reconciliation_mismatches"`

	// Error metrics
	ProviderErrors      map[ProviderType]int64 `json:"provider_errors"`
	LastError           string           `json:"last_error,omitempty"`
	LastErrorTime       *time.Time       `json:"last_error_time,omitempty"`
}

// NewMetricsCollector creates a new metrics collector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		PayoutsByProvider: make(map[ProviderType]int64),
		PayoutsByCurrency: make(map[string]int64),
		AmountByProvider:  make(map[ProviderType]int64),
		ProviderErrors:    make(map[ProviderType]int64),
	}
}

// RecordPayout records a payout.
func (m *MetricsCollector) RecordPayout(provider ProviderType, currency string, amount int64, fee int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.PayoutsTotal++
	m.PayoutsByProvider[provider]++
	m.PayoutsByCurrency[currency]++
	m.TotalAmountPaidOut += amount
	m.TotalFees += fee
	m.AmountByProvider[provider] += amount
}

// RecordPayoutResult records a payout result.
func (m *MetricsCollector) RecordPayoutResult(status PayoutStatus, processingTimeMs int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch status {
	case PayoutStatusSucceeded:
		m.PayoutsSucceeded++
	case PayoutStatusFailed:
		m.PayoutsFailed++
	case PayoutStatusCanceled:
		m.PayoutsCanceled++
	case PayoutStatusReversed:
		m.PayoutsReversed++
	}

	// Update processing time metrics
	if m.MinProcessingTimeMs == 0 || processingTimeMs < m.MinProcessingTimeMs {
		m.MinProcessingTimeMs = processingTimeMs
	}
	if processingTimeMs > m.MaxProcessingTimeMs {
		m.MaxProcessingTimeMs = processingTimeMs
	}

	// Update average (simple running average)
	total := m.PayoutsSucceeded + m.PayoutsFailed
	if total > 0 {
		m.AvgProcessingTimeMs = (m.AvgProcessingTimeMs*(total-1) + processingTimeMs) / total
	}
}

// RecordKYCRejection records a KYC rejection.
func (m *MetricsCollector) RecordKYCRejection() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.KYCRejections++
}

// RecordAMLRejection records an AML rejection.
func (m *MetricsCollector) RecordAMLRejection() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AMLRejections++
}

// RecordAMLFlagged records an AML flagged payout.
func (m *MetricsCollector) RecordAMLFlagged() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AMLFlagged++
}

// RecordWebhook records a webhook event.
func (m *MetricsCollector) RecordWebhook(processed bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.WebhooksReceived++
	if processed {
		m.WebhooksProcessed++
	} else {
		m.WebhooksFailed++
	}
}

// RecordReconciliation records a reconciliation result.
func (m *MetricsCollector) RecordReconciliation(result *ReconciliationResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ReconciliationsRun++
	m.ReconciliationMatches += int64(result.Matched)
	m.ReconciliationMismatches += int64(result.Mismatched)
}

// RecordProviderError records a provider error.
func (m *MetricsCollector) RecordProviderError(provider ProviderType, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ProviderErrors[provider]++
	m.LastError = err.Error()
	now := time.Now()
	m.LastErrorTime = &now
}

// GetSnapshot returns a snapshot of the metrics.
func (m *MetricsCollector) GetSnapshot() *MetricsCollector {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy
	snapshot := &MetricsCollector{
		PayoutsTotal:        m.PayoutsTotal,
		PayoutsSucceeded:    m.PayoutsSucceeded,
		PayoutsFailed:       m.PayoutsFailed,
		PayoutsCanceled:     m.PayoutsCanceled,
		PayoutsReversed:     m.PayoutsReversed,
		PayoutsByProvider:   make(map[ProviderType]int64),
		PayoutsByCurrency:   make(map[string]int64),
		TotalAmountPaidOut:  m.TotalAmountPaidOut,
		TotalFees:           m.TotalFees,
		AmountByProvider:    make(map[ProviderType]int64),
		AvgProcessingTimeMs: m.AvgProcessingTimeMs,
		MaxProcessingTimeMs: m.MaxProcessingTimeMs,
		MinProcessingTimeMs: m.MinProcessingTimeMs,
		KYCRejections:       m.KYCRejections,
		AMLRejections:       m.AMLRejections,
		AMLFlagged:          m.AMLFlagged,
		WebhooksReceived:    m.WebhooksReceived,
		WebhooksProcessed:   m.WebhooksProcessed,
		WebhooksFailed:      m.WebhooksFailed,
		ReconciliationsRun:  m.ReconciliationsRun,
		ReconciliationMatches: m.ReconciliationMatches,
		ReconciliationMismatches: m.ReconciliationMismatches,
		ProviderErrors:      make(map[ProviderType]int64),
		LastError:           m.LastError,
		LastErrorTime:       m.LastErrorTime,
	}

	for k, v := range m.PayoutsByProvider {
		snapshot.PayoutsByProvider[k] = v
	}
	for k, v := range m.PayoutsByCurrency {
		snapshot.PayoutsByCurrency[k] = v
	}
	for k, v := range m.AmountByProvider {
		snapshot.AmountByProvider[k] = v
	}
	for k, v := range m.ProviderErrors {
		snapshot.ProviderErrors[k] = v
	}

	return snapshot
}

// ============================================================================
// Alert Manager
// ============================================================================

// AlertLevel represents the severity of an alert.
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelError    AlertLevel = "error"
	AlertLevelCritical AlertLevel = "critical"
)

// Alert represents an alert.
type Alert struct {
	// ID is the unique alert identifier
	ID string `json:"id"`

	// Level is the alert severity
	Level AlertLevel `json:"level"`

	// Type identifies the alert type
	Type string `json:"type"`

	// Title is a short title
	Title string `json:"title"`

	// Message is the detailed message
	Message string `json:"message"`

	// Provider is the related provider if any
	Provider ProviderType `json:"provider,omitempty"`

	// PayoutID is the related payout ID if any
	PayoutID string `json:"payout_id,omitempty"`

	// Metadata contains additional context
	Metadata map[string]string `json:"metadata,omitempty"`

	// CreatedAt is when the alert was created
	CreatedAt time.Time `json:"created_at"`

	// Acknowledged indicates if the alert was acknowledged
	Acknowledged bool `json:"acknowledged"`

	// AcknowledgedAt is when the alert was acknowledged
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`

	// AcknowledgedBy is who acknowledged the alert
	AcknowledgedBy string `json:"acknowledged_by,omitempty"`
}

// AlertHandler is a function that handles alerts.
type AlertHandler func(ctx context.Context, alert *Alert) error

// AlertManager manages alerts.
type AlertManager struct {
	mu       sync.RWMutex
	alerts   []*Alert
	handlers []AlertHandler
	config   AlertConfig
}

// AlertConfig contains alert configuration.
type AlertConfig struct {
	// Enabled enables alerting
	Enabled bool `json:"enabled"`

	// MaxAlerts is the maximum number of alerts to keep
	MaxAlerts int `json:"max_alerts"`

	// Thresholds for automatic alerts
	FailureRateThreshold float64 `json:"failure_rate_threshold"` // 0-1
	ProviderErrorThreshold int64 `json:"provider_error_threshold"`
	ReconciliationMismatchThreshold int64 `json:"reconciliation_mismatch_threshold"`
}

// DefaultAlertConfig returns the default alert configuration.
func DefaultAlertConfig() AlertConfig {
	return AlertConfig{
		Enabled:                        true,
		MaxAlerts:                      1000,
		FailureRateThreshold:           0.1,  // 10% failure rate
		ProviderErrorThreshold:         10,   // 10 provider errors
		ReconciliationMismatchThreshold: 5,   // 5 mismatches
	}
}

// NewAlertManager creates a new alert manager.
func NewAlertManager(config AlertConfig) *AlertManager {
	return &AlertManager{
		alerts:   make([]*Alert, 0),
		handlers: make([]AlertHandler, 0),
		config:   config,
	}
}

// AddHandler adds an alert handler.
func (m *AlertManager) AddHandler(handler AlertHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, handler)
}

// CreateAlert creates a new alert.
func (m *AlertManager) CreateAlert(ctx context.Context, level AlertLevel, alertType, title, message string) *Alert {
	alert := &Alert{
		ID:        fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		Level:     level,
		Type:      alertType,
		Title:     title,
		Message:   message,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]string),
	}

	m.mu.Lock()
	m.alerts = append(m.alerts, alert)
	
	// Trim if over max
	if len(m.alerts) > m.config.MaxAlerts {
		m.alerts = m.alerts[len(m.alerts)-m.config.MaxAlerts:]
	}
	m.mu.Unlock()

	// Call handlers
	for _, handler := range m.handlers {
		go handler(ctx, alert)
	}

	return alert
}

// AlertPayoutFailure creates an alert for a payout failure.
func (m *AlertManager) AlertPayoutFailure(ctx context.Context, payout *PayoutIntent) {
	m.CreateAlert(ctx, AlertLevelError, "payout_failure",
		fmt.Sprintf("Payout Failed: %s", payout.ID),
		fmt.Sprintf("Payout %s failed with code: %s - %s", payout.ID, payout.FailureCode, payout.FailureMessage),
	)
}

// AlertProviderError creates an alert for a provider error.
func (m *AlertManager) AlertProviderError(ctx context.Context, provider ProviderType, err error) {
	alert := m.CreateAlert(ctx, AlertLevelWarning, "provider_error",
		fmt.Sprintf("Provider Error: %s", provider),
		fmt.Sprintf("Provider %s reported error: %v", provider, err),
	)
	alert.Provider = provider
}

// AlertReconciliationMismatch creates an alert for a reconciliation mismatch.
func (m *AlertManager) AlertReconciliationMismatch(ctx context.Context, record *ReconciliationRecord) {
	alert := m.CreateAlert(ctx, AlertLevelWarning, "reconciliation_mismatch",
		fmt.Sprintf("Reconciliation Mismatch: %s", record.PayoutID),
		fmt.Sprintf("Payout %s has a discrepancy of %d", record.PayoutID, record.Discrepancy),
	)
	alert.PayoutID = record.PayoutID
	alert.Metadata["discrepancy"] = fmt.Sprintf("%d", record.Discrepancy)
}

// AlertHighFailureRate creates an alert for high failure rate.
func (m *AlertManager) AlertHighFailureRate(ctx context.Context, failureRate float64) {
	m.CreateAlert(ctx, AlertLevelCritical, "high_failure_rate",
		"High Payout Failure Rate",
		fmt.Sprintf("Payout failure rate is %.1f%%, exceeding threshold of %.1f%%",
			failureRate*100, m.config.FailureRateThreshold*100),
	)
}

// GetAlerts returns all alerts.
func (m *AlertManager) GetAlerts(limit int) []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.alerts) {
		limit = len(m.alerts)
	}

	result := make([]*Alert, limit)
	copy(result, m.alerts[len(m.alerts)-limit:])

	// Reverse to get newest first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// GetUnacknowledgedAlerts returns unacknowledged alerts.
func (m *AlertManager) GetUnacknowledgedAlerts() []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Alert, 0)
	for _, alert := range m.alerts {
		if !alert.Acknowledged {
			result = append(result, alert)
		}
	}

	return result
}

// AcknowledgeAlert acknowledges an alert.
func (m *AlertManager) AcknowledgeAlert(alertID, acknowledgedBy string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, alert := range m.alerts {
		if alert.ID == alertID {
			alert.Acknowledged = true
			now := time.Now()
			alert.AcknowledgedAt = &now
			alert.AcknowledgedBy = acknowledgedBy
			return nil
		}
	}

	return fmt.Errorf("alert not found: %s", alertID)
}

// ============================================================================
// Monitoring Service
// ============================================================================

// MonitoringService provides monitoring and alerting for off-ramp operations.
type MonitoringService struct {
	metrics *MetricsCollector
	alerts  *AlertManager
	service Service

	// Check intervals
	checkInterval time.Duration
	stopCh        chan struct{}
	doneCh        chan struct{}
}

// NewMonitoringService creates a new monitoring service.
func NewMonitoringService(service Service, alertConfig AlertConfig) *MonitoringService {
	return &MonitoringService{
		metrics:       NewMetricsCollector(),
		alerts:        NewAlertManager(alertConfig),
		service:       service,
		checkInterval: 1 * time.Minute,
		stopCh:        make(chan struct{}),
		doneCh:        make(chan struct{}),
	}
}

// Start starts the monitoring service.
func (m *MonitoringService) Start(ctx context.Context) {
	go m.run(ctx)
}

// Stop stops the monitoring service.
func (m *MonitoringService) Stop() {
	close(m.stopCh)
	<-m.doneCh
}

// run is the monitoring loop.
func (m *MonitoringService) run(ctx context.Context) {
	defer close(m.doneCh)

	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.runChecks(ctx)
		case <-m.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// runChecks runs monitoring checks.
func (m *MonitoringService) runChecks(ctx context.Context) {
	snapshot := m.metrics.GetSnapshot()

	// Check failure rate
	if snapshot.PayoutsTotal > 0 {
		failureRate := float64(snapshot.PayoutsFailed) / float64(snapshot.PayoutsTotal)
		if failureRate > m.alerts.config.FailureRateThreshold {
			m.alerts.AlertHighFailureRate(ctx, failureRate)
		}
	}

	// Check provider errors
	for provider, errorCount := range snapshot.ProviderErrors {
		if errorCount > m.alerts.config.ProviderErrorThreshold {
			m.alerts.CreateAlert(ctx, AlertLevelWarning, "provider_error_threshold",
				fmt.Sprintf("High Provider Errors: %s", provider),
				fmt.Sprintf("Provider %s has %d errors, exceeding threshold of %d",
					provider, errorCount, m.alerts.config.ProviderErrorThreshold),
			)
		}
	}

	// Check reconciliation mismatches
	if snapshot.ReconciliationMismatches > m.alerts.config.ReconciliationMismatchThreshold {
		m.alerts.CreateAlert(ctx, AlertLevelWarning, "reconciliation_mismatch_threshold",
			"High Reconciliation Mismatches",
			fmt.Sprintf("%d reconciliation mismatches, exceeding threshold of %d",
				snapshot.ReconciliationMismatches, m.alerts.config.ReconciliationMismatchThreshold),
		)
	}
}

// GetMetrics returns the metrics collector.
func (m *MonitoringService) GetMetrics() *MetricsCollector {
	return m.metrics
}

// GetAlertManager returns the alert manager.
func (m *MonitoringService) GetAlertManager() *AlertManager {
	return m.alerts
}
