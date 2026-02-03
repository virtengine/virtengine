// Package provider_daemon implements the provider daemon for VirtEngine.
//
// SECURITY-007: Key Compromise Detection
// This file provides key compromise detection and response mechanisms.
package provider_daemon

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	verrors "github.com/virtengine/virtengine/pkg/errors"
)

// Sentinel errors for key compromise
var (
	// ErrKeyCompromised is returned when a key is determined to be compromised
	ErrKeyCompromised = verrors.ErrRevoked

	// ErrCompromiseAlreadyReported is returned when a compromise was already reported
	ErrCompromiseAlreadyReported = verrors.ErrConflict
)

// CompromiseIndicator represents a type of compromise indicator
type CompromiseIndicator string

const (
	// IndicatorUnauthorizedUsage indicates unauthorized key usage detected
	IndicatorUnauthorizedUsage CompromiseIndicator = "unauthorized_usage"

	// IndicatorAnomalousLocation indicates key used from unexpected location
	IndicatorAnomalousLocation CompromiseIndicator = "anomalous_location"

	// IndicatorAnomalousTime indicates key used at unexpected time
	IndicatorAnomalousTime CompromiseIndicator = "anomalous_time"

	// IndicatorRapidUsage indicates abnormally rapid key usage
	IndicatorRapidUsage CompromiseIndicator = "rapid_usage"

	// IndicatorFailedVerification indicates signature verification failures
	IndicatorFailedVerification CompromiseIndicator = "failed_verification"

	// IndicatorDuplicateSignature indicates duplicate signatures detected
	IndicatorDuplicateSignature CompromiseIndicator = "duplicate_signature"

	// IndicatorExternalReport indicates external compromise report
	IndicatorExternalReport CompromiseIndicator = "external_report"

	// IndicatorKeyLeakage indicates potential key leakage detected
	IndicatorKeyLeakage CompromiseIndicator = "key_leakage"

	// IndicatorBruteForceAttempt indicates brute force attempts detected
	IndicatorBruteForceAttempt CompromiseIndicator = "brute_force_attempt"

	// IndicatorWeakEntropy indicates weak entropy in key generation
	IndicatorWeakEntropy CompromiseIndicator = "weak_entropy"
)

// CompromiseSeverity represents the severity of a compromise
type CompromiseSeverity string

const (
	// SeverityLow indicates low severity (investigation recommended)
	SeverityLow CompromiseSeverity = "low"

	// SeverityMedium indicates medium severity (action recommended)
	SeverityMedium CompromiseSeverity = "medium"

	// SeverityHigh indicates high severity (immediate action required)
	SeverityHigh CompromiseSeverity = "high"

	// SeverityCritical indicates critical severity (emergency response)
	SeverityCritical CompromiseSeverity = "critical"
)

// CompromiseDetectorConfig configures the compromise detector
type CompromiseDetectorConfig struct {
	// UsageThresholdPerMinute is the max signatures per minute before flagging
	UsageThresholdPerMinute int `json:"usage_threshold_per_minute"`

	// UsageThresholdPerHour is the max signatures per hour before flagging
	UsageThresholdPerHour int `json:"usage_threshold_per_hour"`

	// FailedVerificationThreshold is max verification failures before flagging
	FailedVerificationThreshold int `json:"failed_verification_threshold"`

	// AnomalousTimeWindowStart is the start of expected usage window (hour)
	AnomalousTimeWindowStart int `json:"anomalous_time_window_start"`

	// AnomalousTimeWindowEnd is the end of expected usage window (hour)
	AnomalousTimeWindowEnd int `json:"anomalous_time_window_end"`

	// EnableLocationTracking enables location-based anomaly detection
	EnableLocationTracking bool `json:"enable_location_tracking"`

	// AllowedLocations is the list of allowed IP ranges/locations
	AllowedLocations []string `json:"allowed_locations,omitempty"`

	// AutoRevokeOnCritical automatically revokes keys on critical severity
	AutoRevokeOnCritical bool `json:"auto_revoke_on_critical"`

	// AlertWebhookURL is a webhook for compromise alerts
	AlertWebhookURL string `json:"alert_webhook_url,omitempty"`

	// RetentionDays is how long to keep compromise events
	RetentionDays int `json:"retention_days"`
}

// DefaultCompromiseDetectorConfig returns the default configuration
func DefaultCompromiseDetectorConfig() *CompromiseDetectorConfig {
	return &CompromiseDetectorConfig{
		UsageThresholdPerMinute:     10,
		UsageThresholdPerHour:       100,
		FailedVerificationThreshold: 5,
		AnomalousTimeWindowStart:    6,  // 6 AM
		AnomalousTimeWindowEnd:      22, // 10 PM
		EnableLocationTracking:      true,
		AutoRevokeOnCritical:        true,
		RetentionDays:               90,
	}
}

// CompromiseEvent represents a detected compromise indicator
type CompromiseEvent struct {
	// ID is the unique event identifier
	ID string `json:"id"`

	// KeyID is the affected key
	KeyID string `json:"key_id"`

	// KeyFingerprint is the key's public key fingerprint
	KeyFingerprint string `json:"key_fingerprint"`

	// Indicator is the type of compromise indicator
	Indicator CompromiseIndicator `json:"indicator"`

	// Severity is the severity level
	Severity CompromiseSeverity `json:"severity"`

	// Description is a human-readable description
	Description string `json:"description"`

	// Evidence contains supporting evidence
	Evidence *CompromiseEvidence `json:"evidence,omitempty"`

	// DetectedAt is when the compromise was detected
	DetectedAt time.Time `json:"detected_at"`

	// ReportedBy indicates the source of the report
	ReportedBy string `json:"reported_by"`

	// Acknowledged indicates if the event was acknowledged
	Acknowledged bool `json:"acknowledged"`

	// AcknowledgedBy is who acknowledged the event
	AcknowledgedBy string `json:"acknowledged_by,omitempty"`

	// AcknowledgedAt is when the event was acknowledged
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`

	// ResponseActions lists actions taken in response
	ResponseActions []ResponseAction `json:"response_actions,omitempty"`
}

// CompromiseEvidence contains evidence for a compromise event
type CompromiseEvidence struct {
	// SourceIP is the source IP address (if applicable)
	SourceIP string `json:"source_ip,omitempty"`

	// Location is the geographic location (if available)
	Location string `json:"location,omitempty"`

	// Timestamp is when the suspicious activity occurred
	Timestamp time.Time `json:"timestamp"`

	// UsageCount is the count of usages in the detection window
	UsageCount int `json:"usage_count,omitempty"`

	// ExpectedBehavior describes expected behavior
	ExpectedBehavior string `json:"expected_behavior,omitempty"`

	// ObservedBehavior describes observed behavior
	ObservedBehavior string `json:"observed_behavior,omitempty"`

	// RawData contains raw evidence data
	RawData map[string]interface{} `json:"raw_data,omitempty"`
}

// ResponseAction represents an action taken in response to compromise
type ResponseAction struct {
	// Action is the action type
	Action string `json:"action"`

	// PerformedAt is when the action was performed
	PerformedAt time.Time `json:"performed_at"`

	// PerformedBy is who performed the action
	PerformedBy string `json:"performed_by"`

	// Success indicates if the action succeeded
	Success bool `json:"success"`

	// Details contains action details
	Details string `json:"details,omitempty"`
}

// CompromiseDetector detects and responds to key compromises
type CompromiseDetector struct {
	config       *CompromiseDetectorConfig
	keyManager   *KeyManager
	events       map[string]*CompromiseEvent
	usageHistory map[string][]time.Time // key ID -> usage timestamps
	mu           sync.RWMutex
}

// NewCompromiseDetector creates a new compromise detector
func NewCompromiseDetector(config *CompromiseDetectorConfig, keyManager *KeyManager) *CompromiseDetector {
	if config == nil {
		config = DefaultCompromiseDetectorConfig()
	}
	return &CompromiseDetector{
		config:       config,
		keyManager:   keyManager,
		events:       make(map[string]*CompromiseEvent),
		usageHistory: make(map[string][]time.Time),
	}
}

// RecordKeyUsage records a key usage event for anomaly detection
func (d *CompromiseDetector) RecordKeyUsage(keyID string, sourceIP string, timestamp time.Time) []CompromiseIndicator {
	d.mu.Lock()
	defer d.mu.Unlock()

	indicators := make([]CompromiseIndicator, 0)

	// Add to usage history
	if _, exists := d.usageHistory[keyID]; !exists {
		d.usageHistory[keyID] = make([]time.Time, 0)
	}
	d.usageHistory[keyID] = append(d.usageHistory[keyID], timestamp)

	// Prune old entries (keep last hour)
	cutoff := timestamp.Add(-1 * time.Hour)
	newHistory := make([]time.Time, 0)
	for _, t := range d.usageHistory[keyID] {
		if t.After(cutoff) {
			newHistory = append(newHistory, t)
		}
	}
	d.usageHistory[keyID] = newHistory

	// Check for rapid usage
	usageInLastMinute := 0
	minuteCutoff := timestamp.Add(-1 * time.Minute)
	for _, t := range newHistory {
		if t.After(minuteCutoff) {
			usageInLastMinute++
		}
	}

	if usageInLastMinute > d.config.UsageThresholdPerMinute {
		indicators = append(indicators, IndicatorRapidUsage)
		d.createEvent(keyID, IndicatorRapidUsage, SeverityMedium,
			fmt.Sprintf("Rapid key usage detected: %d uses in last minute (threshold: %d)",
				usageInLastMinute, d.config.UsageThresholdPerMinute),
			&CompromiseEvidence{
				SourceIP:   sourceIP,
				Timestamp:  timestamp,
				UsageCount: usageInLastMinute,
			})
	}

	// Check for hourly threshold
	if len(newHistory) > d.config.UsageThresholdPerHour {
		indicators = append(indicators, IndicatorRapidUsage)
	}

	// Check for anomalous time
	hour := timestamp.Hour()
	if hour < d.config.AnomalousTimeWindowStart || hour > d.config.AnomalousTimeWindowEnd {
		indicators = append(indicators, IndicatorAnomalousTime)
		d.createEvent(keyID, IndicatorAnomalousTime, SeverityLow,
			fmt.Sprintf("Key used outside normal hours: %02d:00 (expected: %02d:00-%02d:00)",
				hour, d.config.AnomalousTimeWindowStart, d.config.AnomalousTimeWindowEnd),
			&CompromiseEvidence{
				SourceIP:         sourceIP,
				Timestamp:        timestamp,
				ExpectedBehavior: fmt.Sprintf("Usage between %02d:00 and %02d:00", d.config.AnomalousTimeWindowStart, d.config.AnomalousTimeWindowEnd),
				ObservedBehavior: fmt.Sprintf("Usage at %02d:00", hour),
			})
	}

	// Check for location anomaly (if enabled)
	if d.config.EnableLocationTracking && len(d.config.AllowedLocations) > 0 {
		allowed := false
		for _, loc := range d.config.AllowedLocations {
			if matchesLocation(sourceIP, loc) {
				allowed = true
				break
			}
		}
		if !allowed {
			indicators = append(indicators, IndicatorAnomalousLocation)
			d.createEvent(keyID, IndicatorAnomalousLocation, SeverityMedium,
				fmt.Sprintf("Key used from unauthorized location: %s", sourceIP),
				&CompromiseEvidence{
					SourceIP:  sourceIP,
					Timestamp: timestamp,
				})
		}
	}

	return indicators
}

// RecordVerificationFailure records a signature verification failure
func (d *CompromiseDetector) RecordVerificationFailure(keyID string, message []byte, signature []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Count recent failures for this key
	failureCount := 0
	now := time.Now().UTC()
	hourAgo := now.Add(-1 * time.Hour)

	for _, event := range d.events {
		if event.KeyID == keyID && event.Indicator == IndicatorFailedVerification &&
			event.DetectedAt.After(hourAgo) {
			failureCount++
		}
	}

	if failureCount >= d.config.FailedVerificationThreshold {
		d.createEvent(keyID, IndicatorFailedVerification, SeverityHigh,
			fmt.Sprintf("Multiple signature verification failures: %d in last hour",
				failureCount+1),
			&CompromiseEvidence{
				Timestamp:  now,
				UsageCount: failureCount + 1,
			})

		// Auto-revoke if configured
		if d.config.AutoRevokeOnCritical && failureCount >= d.config.FailedVerificationThreshold*2 {
			d.handleCriticalCompromise(keyID)
		}
	}
}

// ReportCompromise reports an external compromise indicator
func (d *CompromiseDetector) ReportCompromise(keyID string, indicator CompromiseIndicator, severity CompromiseSeverity, description string, reportedBy string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	event := d.createEvent(keyID, indicator, severity, description, &CompromiseEvidence{
		Timestamp: time.Now().UTC(),
	})
	event.ReportedBy = reportedBy

	// Handle critical severity
	if severity == SeverityCritical && d.config.AutoRevokeOnCritical {
		d.handleCriticalCompromise(keyID)
	}

	return nil
}

// createEvent creates a new compromise event
func (d *CompromiseDetector) createEvent(keyID string, indicator CompromiseIndicator, severity CompromiseSeverity, description string, evidence *CompromiseEvidence) *CompromiseEvent {
	eventID := generateCompromiseEventID(keyID, indicator)

	event := &CompromiseEvent{
		ID:             eventID,
		KeyID:          keyID,
		KeyFingerprint: computeKeyFingerprint(keyID),
		Indicator:      indicator,
		Severity:       severity,
		Description:    description,
		Evidence:       evidence,
		DetectedAt:     time.Now().UTC(),
		ReportedBy:     "compromise_detector",
	}

	d.events[eventID] = event

	// Send alert if webhook configured with panic recovery
	if d.config.AlertWebhookURL != "" {
		verrors.SafeGo("provider-daemon:compromise-alert", func() {
			d.sendAlert(event)
		})
	}

	return event
}

// handleCriticalCompromise handles a critical compromise by revoking the key
func (d *CompromiseDetector) handleCriticalCompromise(keyID string) {
	if d.keyManager == nil {
		return
	}

	// Attempt to revoke the key
	err := d.keyManager.RevokeKey(keyID)

	action := ResponseAction{
		Action:      "auto_revoke",
		PerformedAt: time.Now().UTC(),
		PerformedBy: "compromise_detector",
		Success:     err == nil,
	}
	if err != nil {
		action.Details = err.Error()
	} else {
		action.Details = "Key automatically revoked due to critical compromise"
	}

	// Update all critical events for this key with the action
	for _, event := range d.events {
		if event.KeyID == keyID && event.Severity == SeverityCritical {
			event.ResponseActions = append(event.ResponseActions, action)
		}
	}
}

// sendAlert sends an alert webhook (placeholder)
func (d *CompromiseDetector) sendAlert(event *CompromiseEvent) {
	// In production, would send HTTP POST to webhook URL
	_ = event
}

// GetEvents retrieves all compromise events
func (d *CompromiseDetector) GetEvents() []*CompromiseEvent {
	d.mu.RLock()
	defer d.mu.RUnlock()

	events := make([]*CompromiseEvent, 0, len(d.events))
	for _, event := range d.events {
		events = append(events, event)
	}
	return events
}

// GetEventsByKey retrieves events for a specific key
func (d *CompromiseDetector) GetEventsByKey(keyID string) []*CompromiseEvent {
	d.mu.RLock()
	defer d.mu.RUnlock()

	events := make([]*CompromiseEvent, 0)
	for _, event := range d.events {
		if event.KeyID == keyID {
			events = append(events, event)
		}
	}
	return events
}

// GetEventsBySeverity retrieves events by severity
func (d *CompromiseDetector) GetEventsBySeverity(severity CompromiseSeverity) []*CompromiseEvent {
	d.mu.RLock()
	defer d.mu.RUnlock()

	events := make([]*CompromiseEvent, 0)
	for _, event := range d.events {
		if event.Severity == severity {
			events = append(events, event)
		}
	}
	return events
}

// AcknowledgeEvent marks an event as acknowledged
func (d *CompromiseDetector) AcknowledgeEvent(eventID, acknowledgedBy string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	event, exists := d.events[eventID]
	if !exists {
		return fmt.Errorf("event not found: %s", eventID)
	}

	now := time.Now().UTC()
	event.Acknowledged = true
	event.AcknowledgedBy = acknowledgedBy
	event.AcknowledgedAt = &now

	return nil
}

// AddResponseAction adds a response action to an event
func (d *CompromiseDetector) AddResponseAction(eventID string, action ResponseAction) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	event, exists := d.events[eventID]
	if !exists {
		return fmt.Errorf("event not found: %s", eventID)
	}

	event.ResponseActions = append(event.ResponseActions, action)
	return nil
}

// IsKeyCompromised checks if a key has any unacknowledged critical/high severity events
func (d *CompromiseDetector) IsKeyCompromised(keyID string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, event := range d.events {
		if event.KeyID == keyID && !event.Acknowledged {
			if event.Severity == SeverityCritical || event.Severity == SeverityHigh {
				return true
			}
		}
	}
	return false
}

// Cleanup removes old events based on retention policy
func (d *CompromiseDetector) Cleanup() int {
	d.mu.Lock()
	defer d.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -d.config.RetentionDays)
	count := 0

	for id, event := range d.events {
		if event.DetectedAt.Before(cutoff) {
			delete(d.events, id)
			count++
		}
	}

	return count
}

// Helper functions

func generateCompromiseEventID(keyID string, indicator CompromiseIndicator) string {
	data := fmt.Sprintf("%s-%s-%d", keyID, indicator, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:12])
}

func computeKeyFingerprint(keyID string) string {
	hash := sha256.Sum256([]byte(keyID))
	return hex.EncodeToString(hash[:])
}

func matchesLocation(sourceIP, allowedLocation string) bool {
	// Simplified location matching - would use proper IP range checking
	// or GeoIP database in production
	return sourceIP == allowedLocation
}

// CompromiseAlert represents an alert sent for a compromise
type CompromiseAlert struct {
	EventID     string              `json:"event_id"`
	KeyID       string              `json:"key_id"`
	Indicator   CompromiseIndicator `json:"indicator"`
	Severity    CompromiseSeverity  `json:"severity"`
	Description string              `json:"description"`
	DetectedAt  time.Time           `json:"detected_at"`
	AlertSentAt time.Time           `json:"alert_sent_at"`
}

// CompromiseReport generates a report of compromise events
type CompromiseReport struct {
	GeneratedAt    time.Time      `json:"generated_at"`
	ReportPeriod   string         `json:"report_period"`
	TotalEvents    int            `json:"total_events"`
	BySeverity     map[string]int `json:"by_severity"`
	ByIndicator    map[string]int `json:"by_indicator"`
	AffectedKeys   []string       `json:"affected_keys"`
	CriticalCount  int            `json:"critical_count"`
	Acknowledged   int            `json:"acknowledged"`
	Unacknowledged int            `json:"unacknowledged"`
}

// GenerateReport generates a compromise report
func (d *CompromiseDetector) GenerateReport(since time.Time) *CompromiseReport {
	d.mu.RLock()
	defer d.mu.RUnlock()

	report := &CompromiseReport{
		GeneratedAt:  time.Now().UTC(),
		ReportPeriod: fmt.Sprintf("Since %s", since.Format(time.RFC3339)),
		BySeverity:   make(map[string]int),
		ByIndicator:  make(map[string]int),
		AffectedKeys: make([]string, 0),
	}

	affectedKeysMap := make(map[string]bool)

	for _, event := range d.events {
		if event.DetectedAt.Before(since) {
			continue
		}

		report.TotalEvents++
		report.BySeverity[string(event.Severity)]++
		report.ByIndicator[string(event.Indicator)]++

		if event.Severity == SeverityCritical {
			report.CriticalCount++
		}

		if event.Acknowledged {
			report.Acknowledged++
		} else {
			report.Unacknowledged++
		}

		affectedKeysMap[event.KeyID] = true
	}

	for keyID := range affectedKeysMap {
		report.AffectedKeys = append(report.AffectedKeys, keyID)
	}

	return report
}
