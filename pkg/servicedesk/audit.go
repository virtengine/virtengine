package servicedesk

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"cosmossdk.io/log"
)

// AuditEventType represents types of audit events
type AuditEventType string

const (
	// Ticket lifecycle events
	AuditEventTicketCreate AuditEventType = "ticket_create"
	AuditEventTicketUpdate AuditEventType = "ticket_update"
	AuditEventTicketClose  AuditEventType = "ticket_close"

	// Sync events
	AuditEventSyncSuccess       AuditEventType = "sync_success"
	AuditEventSyncFailed        AuditEventType = "sync_failed"
	AuditEventConflictDetected  AuditEventType = "conflict_detected"
	AuditEventConflictResolved  AuditEventType = "conflict_resolved"

	// External events
	AuditEventExternalCallback  AuditEventType = "external_callback"
	AuditEventAttachmentSync    AuditEventType = "attachment_sync"

	// Admin events
	AuditEventManualSync        AuditEventType = "manual_sync"
	AuditEventConfigChange      AuditEventType = "config_change"
)

// AuditEntry represents an audit log entry
type AuditEntry struct {
	// ID is the unique entry ID
	ID string `json:"id"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// EventType is the type of event
	EventType AuditEventType `json:"event_type"`

	// TicketID is the related ticket ID (if applicable)
	TicketID string `json:"ticket_id,omitempty"`

	// Actor is who performed the action
	Actor string `json:"actor,omitempty"`

	// ServiceDesk is the external service desk (if applicable)
	ServiceDesk ServiceDeskType `json:"service_desk,omitempty"`

	// ExternalID is the external ticket ID (if applicable)
	ExternalID string `json:"external_id,omitempty"`

	// Direction is the sync direction
	Direction SyncDirection `json:"direction,omitempty"`

	// Status indicates success or failure
	Status string `json:"status"`

	// Details contains additional event details
	Details map[string]interface{} `json:"details,omitempty"`

	// Error contains error information (if failed)
	Error string `json:"error,omitempty"`

	// BlockHeight is the on-chain block height (if applicable)
	BlockHeight int64 `json:"block_height,omitempty"`

	// TxHash is the transaction hash (if applicable)
	TxHash string `json:"tx_hash,omitempty"`
}

// AuditLogger handles audit logging for the service desk bridge
type AuditLogger struct {
	config AuditConfig
	logger log.Logger

	// In-memory storage (would use persistent storage in production)
	mu      sync.RWMutex
	entries []AuditEntry
	seq     int64
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(config AuditConfig, logger log.Logger) *AuditLogger {
	return &AuditLogger{
		config:  config,
		logger:  logger.With("component", "audit"),
		entries: make([]AuditEntry, 0),
	}
}

// LogEvent logs an audit event
func (a *AuditLogger) LogEvent(ctx context.Context, eventType AuditEventType, details map[string]interface{}) {
	if !a.config.Enabled {
		return
	}

	entry := AuditEntry{
		ID:        a.generateID(),
		Timestamp: time.Now().UTC(),
		EventType: eventType,
		Status:    "success",
		Details:   details,
	}

	// Extract common fields from details
	if ticketID, ok := details["ticket_id"].(string); ok {
		entry.TicketID = ticketID
	}
	if actor, ok := details["actor"].(string); ok {
		entry.Actor = actor
	}
	if service, ok := details["service_desk"].(ServiceDeskType); ok {
		entry.ServiceDesk = service
	}
	if externalID, ok := details["external_id"].(string); ok {
		entry.ExternalID = externalID
	}
	if direction, ok := details["direction"].(SyncDirection); ok {
		entry.Direction = direction
	}
	if blockHeight, ok := details["block_height"].(int64); ok {
		entry.BlockHeight = blockHeight
	}
	if txHash, ok := details["tx_hash"].(string); ok {
		entry.TxHash = txHash
	}
	if errStr, ok := details["error"].(string); ok {
		entry.Error = errStr
		entry.Status = "failed"
	}

	a.store(entry)
	a.logEntry(entry)
}

// LogError logs an error audit event
func (a *AuditLogger) LogError(ctx context.Context, eventType AuditEventType, err error, details map[string]interface{}) {
	if !a.config.Enabled {
		return
	}

	if details == nil {
		details = make(map[string]interface{})
	}
	details["error"] = err.Error()

	a.LogEvent(ctx, eventType, details)
}

// GetEntries returns audit entries matching the filter
func (a *AuditLogger) GetEntries(filter AuditFilter) []AuditEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var result []AuditEntry
	for _, entry := range a.entries {
		if filter.Matches(entry) {
			result = append(result, entry)
		}
	}
	return result
}

// GetEntriesByTicket returns audit entries for a specific ticket
func (a *AuditLogger) GetEntriesByTicket(ticketID string) []AuditEntry {
	return a.GetEntries(AuditFilter{TicketID: ticketID})
}

// GetEntriesByTimeRange returns audit entries within a time range
func (a *AuditLogger) GetEntriesByTimeRange(start, end time.Time) []AuditEntry {
	return a.GetEntries(AuditFilter{StartTime: start, EndTime: end})
}

// PurgeOldEntries removes entries older than retention period
func (a *AuditLogger) PurgeOldEntries() int {
	a.mu.Lock()
	defer a.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -a.config.RetentionDays)
	var remaining []AuditEntry
	purged := 0

	for _, entry := range a.entries {
		if entry.Timestamp.After(cutoff) {
			remaining = append(remaining, entry)
		} else {
			purged++
		}
	}

	a.entries = remaining
	return purged
}

// store stores an audit entry
func (a *AuditLogger) store(entry AuditEntry) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.entries = append(a.entries, entry)

	// Trim if too many entries (keep last 10000)
	if len(a.entries) > 10000 {
		a.entries = a.entries[len(a.entries)-10000:]
	}
}

// generateID generates a unique audit entry ID
func (a *AuditLogger) generateID() string {
	a.mu.Lock()
	a.seq++
	seq := a.seq
	a.mu.Unlock()

	return time.Now().Format("20060102150405") + "-" + string(rune(seq))
}

// logEntry logs the entry to the configured logger
func (a *AuditLogger) logEntry(entry AuditEntry) {
	// Determine log level
	logFn := a.logger.Info
	switch a.config.LogLevel {
	case "debug":
		logFn = a.logger.Debug
	case "warn":
		if entry.Status == "failed" {
			logFn = a.logger.Warn
		}
	case "error":
		if entry.Status == "failed" {
			logFn = a.logger.Error
		} else {
			return // Don't log non-errors at error level
		}
	}

	// Build log message
	args := []interface{}{
		"event_type", entry.EventType,
		"status", entry.Status,
	}
	if entry.TicketID != "" {
		args = append(args, "ticket_id", entry.TicketID)
	}
	if entry.ServiceDesk != "" {
		args = append(args, "service_desk", entry.ServiceDesk)
	}
	if entry.ExternalID != "" {
		args = append(args, "external_id", entry.ExternalID)
	}
	if entry.Error != "" {
		args = append(args, "error", entry.Error)
	}

	// Add details if sensitive logging is enabled
	if a.config.LogSensitive && entry.Details != nil {
		detailsJSON, _ := json.Marshal(entry.Details)
		args = append(args, "details", string(detailsJSON))
	}

	logFn("audit", args...)
}

// AuditFilter defines criteria for filtering audit entries
type AuditFilter struct {
	// TicketID filters by ticket ID
	TicketID string

	// EventType filters by event type
	EventType AuditEventType

	// ServiceDesk filters by service desk
	ServiceDesk ServiceDeskType

	// Status filters by status
	Status string

	// StartTime filters entries after this time
	StartTime time.Time

	// EndTime filters entries before this time
	EndTime time.Time

	// Actor filters by actor
	Actor string
}

// Matches checks if an entry matches the filter
func (f AuditFilter) Matches(entry AuditEntry) bool {
	if f.TicketID != "" && entry.TicketID != f.TicketID {
		return false
	}
	if f.EventType != "" && entry.EventType != f.EventType {
		return false
	}
	if f.ServiceDesk != "" && entry.ServiceDesk != f.ServiceDesk {
		return false
	}
	if f.Status != "" && entry.Status != f.Status {
		return false
	}
	if !f.StartTime.IsZero() && entry.Timestamp.Before(f.StartTime) {
		return false
	}
	if !f.EndTime.IsZero() && entry.Timestamp.After(f.EndTime) {
		return false
	}
	if f.Actor != "" && entry.Actor != f.Actor {
		return false
	}
	return true
}

// ExportAuditLog exports the audit log as JSON
func (a *AuditLogger) ExportAuditLog(filter AuditFilter) ([]byte, error) {
	entries := a.GetEntries(filter)
	return json.MarshalIndent(entries, "", "  ")
}

// EntryCount returns the total number of entries
func (a *AuditLogger) EntryCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.entries)
}
