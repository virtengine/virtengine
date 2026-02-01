package security_monitoring

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// SecurityMonitorConfig configures the security monitor
type SecurityMonitorConfig struct {
	// Enable/disable individual detectors
	EnableTransactionDetector bool `json:"enable_transaction_detector"`
	EnableFraudDetector       bool `json:"enable_fraud_detector"`
	EnableCryptoAnomalyDetector bool `json:"enable_crypto_anomaly_detector"`
	EnableProviderSecurity    bool `json:"enable_provider_security"`

	// Alert configuration
	AlertWebhookURL    string `json:"alert_webhook_url,omitempty"`
	AlertCooldownSecs  int    `json:"alert_cooldown_secs"`
	MaxAlertsPerMinute int    `json:"max_alerts_per_minute"`

	// Audit log configuration
	AuditLogPath        string `json:"audit_log_path,omitempty"`
	AuditLogMaxSizeMB   int    `json:"audit_log_max_size_mb"`
	AuditLogMaxAgeDays  int    `json:"audit_log_max_age_days"`
	AuditLogCompression bool   `json:"audit_log_compression"`

	// Incident response configuration
	EnableAutoResponse bool `json:"enable_auto_response"`
	PlaybookPath       string `json:"playbook_path,omitempty"`

	// Thresholds
	ThreatLevelHighThreshold     int `json:"threat_level_high_threshold"`
	ThreatLevelCriticalThreshold int `json:"threat_level_critical_threshold"`

	// Monitoring intervals
	MetricsIntervalSecs int `json:"metrics_interval_secs"`
	CleanupIntervalSecs int `json:"cleanup_interval_secs"`
}

// DefaultSecurityMonitorConfig returns default configuration
func DefaultSecurityMonitorConfig() *SecurityMonitorConfig {
	return &SecurityMonitorConfig{
		EnableTransactionDetector:   true,
		EnableFraudDetector:         true,
		EnableCryptoAnomalyDetector: true,
		EnableProviderSecurity:      true,

		AlertCooldownSecs:  60,
		MaxAlertsPerMinute: 100,

		AuditLogMaxSizeMB:   100,
		AuditLogMaxAgeDays:  90,
		AuditLogCompression: true,

		EnableAutoResponse: true,

		ThreatLevelHighThreshold:     10,
		ThreatLevelCriticalThreshold: 25,

		MetricsIntervalSecs: 30,
		CleanupIntervalSecs: 300,
	}
}

// SecurityMonitor is the main security monitoring orchestrator
type SecurityMonitor struct {
	config  *SecurityMonitorConfig
	logger  zerolog.Logger
	metrics *SecurityMetrics

	// Detectors
	txDetector     *TransactionDetector
	fraudDetector  *FraudDetector
	cryptoDetector *CryptoAnomalyDetector
	providerSec    *ProviderSecurityMonitor

	// Audit and response
	auditLog          *AuditLog
	incidentResponder *IncidentResponder

	// State
	activeIncidents   map[string]*SecurityIncident
	recentAlertCounts map[string]int
	lastAlertTimes    map[string]time.Time

	// Channels
	eventChan  chan *SecurityEvent
	alertChan  chan *SecurityAlert

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex
}

// SecurityIncident represents an active security incident
type SecurityIncident struct {
	ID            string                `json:"id"`
	Type          string                `json:"type"`
	Severity      SecurityEventSeverity `json:"severity"`
	StartedAt     time.Time             `json:"started_at"`
	LastEventAt   time.Time             `json:"last_event_at"`
	EventCount    int                   `json:"event_count"`
	Description   string                `json:"description"`
	AffectedAssets []string             `json:"affected_assets,omitempty"`
	Status        string                `json:"status"` // open, investigating, mitigated, resolved
	AssignedTo    string                `json:"assigned_to,omitempty"`
	PlaybookID    string                `json:"playbook_id,omitempty"`
}

// NewSecurityMonitor creates a new security monitor
func NewSecurityMonitor(config *SecurityMonitorConfig, logger zerolog.Logger) (*SecurityMonitor, error) {
	if config == nil {
		config = DefaultSecurityMonitorConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	sm := &SecurityMonitor{
		config:            config,
		logger:            logger.With().Str("component", "security-monitor").Logger(),
		metrics:           GetSecurityMetrics(),
		activeIncidents:   make(map[string]*SecurityIncident),
		recentAlertCounts: make(map[string]int),
		lastAlertTimes:    make(map[string]time.Time),
		eventChan:         make(chan *SecurityEvent, 1000),
		alertChan:         make(chan *SecurityAlert, 100),
		ctx:               ctx,
		cancel:            cancel,
	}

	// Initialize detectors
	if config.EnableTransactionDetector {
		sm.txDetector = NewTransactionDetector(DefaultTransactionDetectorConfig(), logger)
	}
	if config.EnableFraudDetector {
		sm.fraudDetector = NewFraudDetector(DefaultFraudDetectorConfig(), logger)
	}
	if config.EnableCryptoAnomalyDetector {
		sm.cryptoDetector = NewCryptoAnomalyDetector(DefaultCryptoAnomalyConfig(), logger)
	}
	if config.EnableProviderSecurity {
		sm.providerSec = NewProviderSecurityMonitor(DefaultProviderSecurityConfig(), logger)
	}

	// Initialize audit log
	var err error
	sm.auditLog, err = NewAuditLog(config.AuditLogPath, logger)
	if err != nil {
		cancel()
		return nil, err
	}

	// Initialize incident responder
	if config.EnableAutoResponse {
		sm.incidentResponder, err = NewIncidentResponder(config.PlaybookPath, logger)
		if err != nil {
			cancel()
			return nil, err
		}
	}

	return sm, nil
}

// Start starts the security monitor
func (sm *SecurityMonitor) Start() error {
	sm.logger.Info().Msg("starting security monitor")

	// Start background processors
	sm.wg.Add(3)
	go sm.eventProcessor()
	go sm.alertProcessor()
	go sm.periodicTasks()

	// Start detectors
	if sm.txDetector != nil {
		sm.txDetector.Start(sm.ctx, sm.eventChan)
	}
	if sm.fraudDetector != nil {
		sm.fraudDetector.Start(sm.ctx, sm.eventChan)
	}
	if sm.cryptoDetector != nil {
		sm.cryptoDetector.Start(sm.ctx, sm.eventChan)
	}
	if sm.providerSec != nil {
		sm.providerSec.Start(sm.ctx, sm.eventChan)
	}

	// Initialize security score
	sm.metrics.SecurityScore.Set(100)
	sm.metrics.ThreatLevel.Set(0)

	return nil
}

// Stop stops the security monitor
func (sm *SecurityMonitor) Stop() {
	sm.logger.Info().Msg("stopping security monitor")
	sm.cancel()
	sm.wg.Wait()

	if sm.auditLog != nil {
		sm.auditLog.Close()
	}
}

// eventProcessor processes incoming security events
func (sm *SecurityMonitor) eventProcessor() {
	defer sm.wg.Done()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case event := <-sm.eventChan:
			sm.handleEvent(event)
		}
	}
}

// handleEvent processes a single security event
func (sm *SecurityMonitor) handleEvent(event *SecurityEvent) {
	// Update metrics
	sm.metrics.AuditEventsTotal.WithLabelValues(event.Type).Inc()
	sm.metrics.AuditEventsBySeverity.WithLabelValues(string(event.Severity)).Inc()

	// Log to audit
	if sm.auditLog != nil {
		sm.auditLog.LogEvent(event)
	}

	// Check if this should create/update an incident
	sm.updateIncidents(event)

	// Check if alert should be sent
	if sm.shouldAlert(event) {
		sm.sendAlert(event)
	}

	// Log event
	sm.logger.Debug().
		Str("event_id", event.ID).
		Str("type", event.Type).
		Str("severity", string(event.Severity)).
		Msg("security event processed")
}

// updateIncidents updates active incidents based on the event
func (sm *SecurityMonitor) updateIncidents(event *SecurityEvent) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Find or create incident
	incidentKey := event.Type + ":" + event.Source
	incident, exists := sm.activeIncidents[incidentKey]

	if !exists && event.Severity >= SeverityMedium {
		// Create new incident
		incident = &SecurityIncident{
			ID:          generateIncidentID(),
			Type:        event.Type,
			Severity:    event.Severity,
			StartedAt:   event.Timestamp,
			LastEventAt: event.Timestamp,
			EventCount:  1,
			Description: event.Description,
			Status:      "open",
		}
		sm.activeIncidents[incidentKey] = incident
		sm.metrics.SecurityIncidentsActive.Inc()

		sm.logger.Warn().
			Str("incident_id", incident.ID).
			Str("type", incident.Type).
			Str("severity", string(incident.Severity)).
			Msg("new security incident created")

		// Trigger automated response if configured
		if sm.incidentResponder != nil && sm.config.EnableAutoResponse {
			go sm.incidentResponder.HandleIncident(sm.ctx, incident)
		}
	} else if exists {
		// Update existing incident
		incident.LastEventAt = event.Timestamp
		incident.EventCount++
		if event.Severity > incident.Severity {
			incident.Severity = event.Severity
		}
	}

	// Update threat level
	sm.updateThreatLevel()
}

// updateThreatLevel recalculates the current threat level
func (sm *SecurityMonitor) updateThreatLevel() {
	var highCount, criticalCount int
	for _, incident := range sm.activeIncidents {
		switch incident.Severity {
		case SeverityHigh:
			highCount++
		case SeverityCritical:
			criticalCount++
		}
	}

	totalSeverityScore := highCount + (criticalCount * 3)

	if totalSeverityScore >= sm.config.ThreatLevelCriticalThreshold || criticalCount > 0 {
		sm.metrics.ThreatLevel.Set(ThreatLevelValue(SeverityCritical))
	} else if totalSeverityScore >= sm.config.ThreatLevelHighThreshold {
		sm.metrics.ThreatLevel.Set(ThreatLevelValue(SeverityHigh))
	} else if len(sm.activeIncidents) > 0 {
		sm.metrics.ThreatLevel.Set(ThreatLevelValue(SeverityMedium))
	} else {
		sm.metrics.ThreatLevel.Set(ThreatLevelValue(SeverityLow))
	}

	// Update security score (inverse of threat level, 0-100)
	incidentPenalty := len(sm.activeIncidents) * 5
	highPenalty := highCount * 10
	criticalPenalty := criticalCount * 25
	score := 100 - incidentPenalty - highPenalty - criticalPenalty
	if score < 0 {
		score = 0
	}
	sm.metrics.SecurityScore.Set(float64(score))
}

// shouldAlert checks if an alert should be sent for this event
func (sm *SecurityMonitor) shouldAlert(event *SecurityEvent) bool {
	// Only alert on medium severity or higher
	if event.Severity < SeverityMedium {
		return false
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	alertKey := event.Type + ":" + event.Source

	// Check cooldown
	if lastTime, exists := sm.lastAlertTimes[alertKey]; exists {
		cooldown := time.Duration(sm.config.AlertCooldownSecs) * time.Second
		if time.Since(lastTime) < cooldown {
			return false
		}
	}

	// Check rate limit
	if sm.recentAlertCounts[alertKey] >= sm.config.MaxAlertsPerMinute {
		return false
	}

	// Update tracking
	sm.lastAlertTimes[alertKey] = time.Now()
	sm.recentAlertCounts[alertKey]++

	return true
}

// sendAlert sends a security alert
func (sm *SecurityMonitor) sendAlert(event *SecurityEvent) {
	alert := &SecurityAlert{
		ID:          generateAlertID(),
		EventID:     event.ID,
		Type:        event.Type,
		Severity:    event.Severity,
		Title:       "Security Alert: " + event.Type,
		Description: event.Description,
		Source:      event.Source,
		Timestamp:   time.Now(),
		Metadata:    event.Metadata,
	}

	select {
	case sm.alertChan <- alert:
		sm.metrics.AlertsTriggered.WithLabelValues(event.Type, string(event.Severity)).Inc()
	default:
		sm.logger.Warn().Str("event_id", event.ID).Msg("alert channel full, dropping alert")
	}
}

// alertProcessor processes security alerts
func (sm *SecurityMonitor) alertProcessor() {
	defer sm.wg.Done()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case alert := <-sm.alertChan:
			sm.handleAlert(alert)
		}
	}
}

// handleAlert processes a single alert
func (sm *SecurityMonitor) handleAlert(alert *SecurityAlert) {
	// Log the alert
	logLevel := sm.getSeverityLogLevel(alert.Severity)
	sm.logger.WithLevel(logLevel).
		Str("alert_id", alert.ID).
		Str("type", alert.Type).
		Str("severity", string(alert.Severity)).
		Str("source", alert.Source).
		Msg(alert.Title)

	// Send to webhook if configured
	if sm.config.AlertWebhookURL != "" {
		go sm.sendAlertWebhook(alert)
	}
}

// getSeverityLogLevel maps severity to zerolog level
func (sm *SecurityMonitor) getSeverityLogLevel(severity SecurityEventSeverity) zerolog.Level {
	switch severity {
	case SeverityCritical:
		return zerolog.ErrorLevel
	case SeverityHigh:
		return zerolog.WarnLevel
	case SeverityMedium:
		return zerolog.WarnLevel
	default:
		return zerolog.InfoLevel
	}
}

// sendAlertWebhook sends alert to configured webhook
func (sm *SecurityMonitor) sendAlertWebhook(alert *SecurityAlert) {
	// Implementation would send HTTP POST to webhook URL
	_ = alert
}

// periodicTasks runs periodic maintenance tasks
func (sm *SecurityMonitor) periodicTasks() {
	defer sm.wg.Done()

	metricsTimer := time.NewTicker(time.Duration(sm.config.MetricsIntervalSecs) * time.Second)
	cleanupTimer := time.NewTicker(time.Duration(sm.config.CleanupIntervalSecs) * time.Second)

	defer metricsTimer.Stop()
	defer cleanupTimer.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-metricsTimer.C:
			sm.updateMetrics()
		case <-cleanupTimer.C:
			sm.cleanup()
		}
	}
}

// updateMetrics updates periodic metrics
func (sm *SecurityMonitor) updateMetrics() {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sm.metrics.SecurityIncidentsActive.Set(float64(len(sm.activeIncidents)))
}

// cleanup cleans up old state
func (sm *SecurityMonitor) cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Reset alert rate counters
	sm.recentAlertCounts = make(map[string]int)

	// Clean up old resolved incidents (older than 1 hour)
	cutoff := time.Now().Add(-1 * time.Hour)
	for key, incident := range sm.activeIncidents {
		if incident.Status == "resolved" && incident.LastEventAt.Before(cutoff) {
			delete(sm.activeIncidents, key)
		}
	}
}

// RecordTransactionEvent records a transaction for security analysis
func (sm *SecurityMonitor) RecordTransactionEvent(tx *TransactionData) {
	if sm.txDetector != nil {
		sm.txDetector.Analyze(tx)
	}
}

// RecordVEIDVerification records a VEID verification for fraud analysis
func (sm *SecurityMonitor) RecordVEIDVerification(verification *VEIDVerificationData) {
	if sm.fraudDetector != nil {
		sm.fraudDetector.Analyze(verification)
	}
}

// RecordCryptoOperation records a cryptographic operation for anomaly detection
func (sm *SecurityMonitor) RecordCryptoOperation(op *CryptoOperationData) {
	if sm.cryptoDetector != nil {
		sm.cryptoDetector.Analyze(op)
	}
}

// RecordProviderActivity records provider daemon activity for security monitoring
func (sm *SecurityMonitor) RecordProviderActivity(activity *ProviderActivityData) {
	if sm.providerSec != nil {
		sm.providerSec.Analyze(activity)
	}
}

// RecordRateLimitBreach records a rate limit breach
func (sm *SecurityMonitor) RecordRateLimitBreach(breach *RateLimitBreachData) {
	event := &SecurityEvent{
		ID:          generateEventID(),
		Type:        "rate_limit_breach",
		Severity:    breach.Severity,
		Timestamp:   time.Now(),
		Source:      breach.SourceIP,
		Description: breach.Description,
		Metadata: map[string]interface{}{
			"limit_type":     breach.LimitType,
			"current_count":  breach.CurrentCount,
			"limit":          breach.Limit,
			"bypass_attempt": breach.BypassAttempt,
		},
	}

	sm.metrics.RateLimitBreaches.WithLabelValues(breach.LimitType, string(breach.Severity)).Inc()
	if breach.BypassAttempt {
		sm.metrics.RateLimitBypassAttempts.Inc()
	}

	select {
	case sm.eventChan <- event:
	default:
		sm.logger.Warn().Msg("event channel full, dropping rate limit breach event")
	}
}

// RecordAuthFailure records an authentication failure
func (sm *SecurityMonitor) RecordAuthFailure(reason, source string) {
	sm.metrics.AuthFailures.WithLabelValues(reason, source).Inc()

	event := &SecurityEvent{
		ID:          generateEventID(),
		Type:        "auth_failure",
		Severity:    SeverityMedium,
		Timestamp:   time.Now(),
		Source:      source,
		Description: "Authentication failure: " + reason,
		Metadata: map[string]interface{}{
			"reason": reason,
		},
	}

	select {
	case sm.eventChan <- event:
	default:
	}
}

// GetActiveIncidents returns current active incidents
func (sm *SecurityMonitor) GetActiveIncidents() []*SecurityIncident {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	incidents := make([]*SecurityIncident, 0, len(sm.activeIncidents))
	for _, incident := range sm.activeIncidents {
		incidents = append(incidents, incident)
	}
	return incidents
}

// ResolveIncident marks an incident as resolved
func (sm *SecurityMonitor) ResolveIncident(incidentID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for key, incident := range sm.activeIncidents {
		if incident.ID == incidentID {
			incident.Status = "resolved"
			sm.metrics.SecurityIncidentsActive.Dec()
			sm.logger.Info().
				Str("incident_id", incidentID).
				Msg("security incident resolved")

			// Clean up after short delay
			go func(k string) {
				time.Sleep(5 * time.Minute)
				sm.mu.Lock()
				delete(sm.activeIncidents, k)
				sm.mu.Unlock()
			}(key)
			return nil
		}
	}
	return nil
}

