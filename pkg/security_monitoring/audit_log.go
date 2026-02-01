package security_monitoring

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// AuditLog provides structured security event audit logging
type AuditLog struct {
	logger    zerolog.Logger
	logFile   *os.File
	logPath   string
	mu        sync.Mutex
}

// SecurityEvent represents a security event for logging and processing
type SecurityEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    SecurityEventSeverity  `json:"severity"`
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`

	// Classification
	Category    string                 `json:"category,omitempty"`
	Subcategory string                 `json:"subcategory,omitempty"`

	// Attribution
	AccountAddress  string             `json:"account_address,omitempty"`
	ProviderID      string             `json:"provider_id,omitempty"`
	ValidatorID     string             `json:"validator_id,omitempty"`
	SourceIP        string             `json:"source_ip,omitempty"`

	// Context
	BlockHeight     int64              `json:"block_height,omitempty"`
	TransactionHash string             `json:"transaction_hash,omitempty"`
	RequestID       string             `json:"request_id,omitempty"`

	// Response tracking
	ActionsTaken    []string           `json:"actions_taken,omitempty"`
	PlaybookID      string             `json:"playbook_id,omitempty"`
}

// SecurityAlert represents a security alert sent to operators
type SecurityAlert struct {
	ID          string                 `json:"id"`
	EventID     string                 `json:"event_id"`
	Type        string                 `json:"type"`
	Severity    SecurityEventSeverity  `json:"severity"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Source      string                 `json:"source"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Acknowledged bool                  `json:"acknowledged"`
	AcknowledgedAt *time.Time          `json:"acknowledged_at,omitempty"`
	AcknowledgedBy string              `json:"acknowledged_by,omitempty"`
}

// RateLimitBreachData represents rate limit breach data
type RateLimitBreachData struct {
	LimitType     string                `json:"limit_type"`
	SourceIP      string                `json:"source_ip"`
	CurrentCount  int64                 `json:"current_count"`
	Limit         int64                 `json:"limit"`
	BypassAttempt bool                  `json:"bypass_attempt"`
	Description   string                `json:"description"`
	Severity      SecurityEventSeverity `json:"severity"`
}

// AuditLogEntry represents a structured audit log entry
type AuditLogEntry struct {
	Timestamp     string                 `json:"@timestamp"`
	Level         string                 `json:"level"`
	LogType       string                 `json:"log_type"`
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	Severity      string                 `json:"severity"`
	Source        string                 `json:"source"`
	Description   string                 `json:"description"`
	Account       string                 `json:"account,omitempty"`
	Provider      string                 `json:"provider,omitempty"`
	SourceIP      string                 `json:"source_ip,omitempty"`
	BlockHeight   int64                  `json:"block_height,omitempty"`
	TxHash        string                 `json:"tx_hash,omitempty"`
	RequestID     string                 `json:"request_id,omitempty"`
	Actions       []string               `json:"actions,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// NewAuditLog creates a new audit log
func NewAuditLog(logPath string, logger zerolog.Logger) (*AuditLog, error) {
	audit := &AuditLog{
		logger:  logger.With().Str("component", "audit-log").Logger(),
		logPath: logPath,
	}

	// Open log file if path provided
	if logPath != "" {
		var err error
		audit.logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}
	}

	return audit, nil
}

// LogEvent logs a security event to the audit log
func (a *AuditLog) LogEvent(event *SecurityEvent) {
	entry := AuditLogEntry{
		Timestamp:   event.Timestamp.Format(time.RFC3339Nano),
		Level:       severityToLogLevel(event.Severity),
		LogType:     "security_audit",
		EventID:     event.ID,
		EventType:   event.Type,
		Severity:    string(event.Severity),
		Source:      event.Source,
		Description: event.Description,
		Account:     event.AccountAddress,
		Provider:    event.ProviderID,
		SourceIP:    event.SourceIP,
		BlockHeight: event.BlockHeight,
		TxHash:      event.TransactionHash,
		RequestID:   event.RequestID,
		Actions:     event.ActionsTaken,
		Metadata:    event.Metadata,
	}

	// Log to structured logger
	a.logger.WithLevel(severityToZerologLevel(event.Severity)).
		Str("event_id", event.ID).
		Str("event_type", event.Type).
		Str("severity", string(event.Severity)).
		Str("source", event.Source).
		Str("description", event.Description).
		Msg("SECURITY_AUDIT")

	// Write to file if configured
	if a.logFile != nil {
		a.writeToFile(entry)
	}
}

// writeToFile writes an audit entry to the log file
func (a *AuditLog) writeToFile(entry AuditLogEntry) {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		a.logger.Error().Err(err).Msg("failed to marshal audit log entry")
		return
	}

	_, _ = a.logFile.Write(data)
	_, _ = a.logFile.WriteString("\n")
}

// LogAlert logs a security alert
func (a *AuditLog) LogAlert(alert *SecurityAlert) {
	a.logger.WithLevel(severityToZerologLevel(alert.Severity)).
		Str("alert_id", alert.ID).
		Str("event_id", alert.EventID).
		Str("type", alert.Type).
		Str("severity", string(alert.Severity)).
		Str("title", alert.Title).
		Msg("SECURITY_ALERT")

	if a.logFile != nil {
		entry := map[string]interface{}{
			"@timestamp":  alert.Timestamp.Format(time.RFC3339Nano),
			"log_type":    "security_alert",
			"alert_id":    alert.ID,
			"event_id":    alert.EventID,
			"type":        alert.Type,
			"severity":    string(alert.Severity),
			"title":       alert.Title,
			"description": alert.Description,
			"source":      alert.Source,
			"metadata":    alert.Metadata,
		}

		a.mu.Lock()
		data, err := json.Marshal(entry)
		if err == nil {
			_, _ = a.logFile.Write(data)
			_, _ = a.logFile.WriteString("\n")
		}
		a.mu.Unlock()
	}
}

// LogIncidentAction logs an incident response action
func (a *AuditLog) LogIncidentAction(incidentID, action, performedBy, result string) {
	a.logger.Info().
		Str("incident_id", incidentID).
		Str("action", action).
		Str("performed_by", performedBy).
		Str("result", result).
		Msg("INCIDENT_ACTION")

	if a.logFile != nil {
		entry := map[string]interface{}{
			"@timestamp":   time.Now().Format(time.RFC3339Nano),
			"log_type":     "incident_action",
			"incident_id":  incidentID,
			"action":       action,
			"performed_by": performedBy,
			"result":       result,
		}

		a.mu.Lock()
		data, err := json.Marshal(entry)
		if err == nil {
			_, _ = a.logFile.Write(data)
			_, _ = a.logFile.WriteString("\n")
		}
		a.mu.Unlock()
	}
}

// LogPlaybookExecution logs playbook execution
func (a *AuditLog) LogPlaybookExecution(playbookID, incidentID, status string, steps []string, duration time.Duration) {
	a.logger.Info().
		Str("playbook_id", playbookID).
		Str("incident_id", incidentID).
		Str("status", status).
		Strs("steps", steps).
		Dur("duration", duration).
		Msg("PLAYBOOK_EXECUTION")

	if a.logFile != nil {
		entry := map[string]interface{}{
			"@timestamp":   time.Now().Format(time.RFC3339Nano),
			"log_type":     "playbook_execution",
			"playbook_id":  playbookID,
			"incident_id":  incidentID,
			"status":       status,
			"steps":        steps,
			"duration_ms":  duration.Milliseconds(),
		}

		a.mu.Lock()
		data, err := json.Marshal(entry)
		if err == nil {
			_, _ = a.logFile.Write(data)
			_, _ = a.logFile.WriteString("\n")
		}
		a.mu.Unlock()
	}
}

// Close closes the audit log
func (a *AuditLog) Close() error {
	if a.logFile != nil {
		return a.logFile.Close()
	}
	return nil
}

// Helper functions
func severityToLogLevel(severity SecurityEventSeverity) string {
	switch severity {
	case SeverityCritical:
		return "error"
	case SeverityHigh:
		return "warn"
	case SeverityMedium:
		return "warn"
	case SeverityLow:
		return "info"
	default:
		return "info"
	}
}

func severityToZerologLevel(severity SecurityEventSeverity) zerolog.Level {
	switch severity {
	case SeverityCritical:
		return zerolog.ErrorLevel
	case SeverityHigh:
		return zerolog.WarnLevel
	case SeverityMedium:
		return zerolog.WarnLevel
	case SeverityLow:
		return zerolog.InfoLevel
	default:
		return zerolog.InfoLevel
	}
}

