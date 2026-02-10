// Package audit provides audit logging for verification services.
//
// This package implements append-only audit logging for all verification
// actions including signing, verification, key rotation, and access events.
//
// Task Reference: VE-2B - Verification Shared Infrastructure
package audit

import (
	"context"
	"time"
)

// ============================================================================
// Event Types
// ============================================================================

// EventType identifies the type of audit event.
type EventType string

const (
	// Key Management Events
	EventTypeKeyGenerated         EventType = "key_generated"
	EventTypeKeyActivated         EventType = "key_activated"
	EventTypeKeyRotated           EventType = "key_rotated"
	EventTypeKeyRotationCompleted EventType = "key_rotation_completed"
	EventTypeKeyRevoked           EventType = "key_revoked"
	EventTypeKeyExpired           EventType = "key_expired"
	EventTypeKeyAccessed          EventType = "key_accessed"

	// Attestation Events
	EventTypeAttestationSigned   EventType = "attestation_signed"
	EventTypeAttestationVerified EventType = "attestation_verified"
	EventTypeAttestationRejected EventType = "attestation_rejected"

	// Nonce Events
	EventTypeNonceCreated  EventType = "nonce_created"
	EventTypeNonceUsed     EventType = "nonce_used"
	EventTypeNonceExpired  EventType = "nonce_expired"
	EventTypeNonceRejected EventType = "nonce_rejected"

	// Verification Events
	EventTypeVerificationStarted   EventType = "verification_started"
	EventTypeVerificationCompleted EventType = "verification_completed"
	EventTypeVerificationFailed    EventType = "verification_failed"
	EventTypeVerificationAborted   EventType = "verification_aborted"
	EventTypeVerificationInitiated EventType = "verification_initiated"
	EventTypeVerificationCancelled EventType = "verification_cancelled"
	EventTypeResendRequested       EventType = "resend_requested"
	EventTypeWebhookReceived       EventType = "webhook_received"

	// Rate Limiting Events
	EventTypeRateLimitExceeded EventType = "rate_limit_exceeded"
	EventTypeAbuseDetected     EventType = "abuse_detected"
	EventTypeIPBanned          EventType = "ip_banned"
	EventTypeUserBanned        EventType = "user_banned"

	// Access Events
	EventTypeAccessGranted EventType = "access_granted"
	EventTypeAccessDenied  EventType = "access_denied"
	EventTypeLoginAttempt  EventType = "login_attempt"
	EventTypeLogout        EventType = "logout"

	// Administrative Events
	EventTypeConfigChanged  EventType = "config_changed"
	EventTypeServiceStarted EventType = "service_started"
	EventTypeServiceStopped EventType = "service_stopped"
	EventTypeHealthCheck    EventType = "health_check"

	// Error Events
	EventTypeError         EventType = "error"
	EventTypeSecurityAlert EventType = "security_alert"
)

// ============================================================================
// Severity Levels
// ============================================================================

// Severity represents the severity level of an audit event.
type Severity string

const (
	SeverityDebug    Severity = "debug"
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// EventSeverity returns the default severity for an event type.
func EventSeverity(eventType EventType) Severity {
	switch eventType {
	case EventTypeKeyRevoked, EventTypeAbuseDetected, EventTypeSecurityAlert:
		return SeverityCritical
	case EventTypeRateLimitExceeded, EventTypeAccessDenied, EventTypeVerificationFailed,
		EventTypeIPBanned, EventTypeUserBanned, EventTypeNonceRejected:
		return SeverityWarning
	case EventTypeError:
		return SeverityError
	case EventTypeKeyAccessed:
		return SeverityDebug
	default:
		return SeverityInfo
	}
}

// ============================================================================
// Audit Event
// ============================================================================

// Event represents a single audit log entry.
type Event struct {
	// ID is the unique identifier for this event (generated if not set)
	ID string `json:"id,omitempty"`

	// Type identifies the event type
	Type EventType `json:"type"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// Actor is the identity that performed the action
	Actor string `json:"actor"`

	// Resource is the resource being acted upon
	Resource string `json:"resource"`

	// Action is the specific action performed
	Action string `json:"action"`

	// Outcome is the result of the action (success, failure, etc.)
	Outcome EventOutcome `json:"outcome"`

	// Severity is the severity level of the event
	Severity Severity `json:"severity,omitempty"`

	// Details contains additional event-specific information
	Details map[string]interface{} `json:"details,omitempty"`

	// Source identifies where the event originated
	Source *EventSource `json:"source,omitempty"`

	// Request contains request context
	Request *RequestContext `json:"request,omitempty"`

	// Duration is how long the action took
	Duration time.Duration `json:"duration,omitempty"`

	// TraceID is the distributed trace ID
	TraceID string `json:"trace_id,omitempty"`

	// SpanID is the span ID within the trace
	SpanID string `json:"span_id,omitempty"`
}

// EventOutcome represents the outcome of an action.
type EventOutcome string

const (
	OutcomeSuccess EventOutcome = "success"
	OutcomeFailure EventOutcome = "failure"
	OutcomePending EventOutcome = "pending"
	OutcomeUnknown EventOutcome = "unknown"
)

// EventSource identifies where an event originated.
type EventSource struct {
	// Service is the service name
	Service string `json:"service"`

	// Component is the component within the service
	Component string `json:"component,omitempty"`

	// Instance is the service instance ID
	Instance string `json:"instance,omitempty"`

	// Version is the service version
	Version string `json:"version,omitempty"`

	// Host is the hostname
	Host string `json:"host,omitempty"`
}

// RequestContext contains context about the originating request.
type RequestContext struct {
	// RequestID is the unique request identifier
	RequestID string `json:"request_id,omitempty"`

	// IPAddress is the client IP address
	IPAddress string `json:"ip_address,omitempty"`

	// UserAgent is the client user agent
	UserAgent string `json:"user_agent,omitempty"`

	// Endpoint is the API endpoint
	Endpoint string `json:"endpoint,omitempty"`

	// Method is the HTTP method
	Method string `json:"method,omitempty"`

	// AccountAddress is the blockchain account
	AccountAddress string `json:"account_address,omitempty"`
}

// ============================================================================
// Audit Logger Interface
// ============================================================================

// AuditLogger defines the interface for audit logging.
type AuditLogger interface {
	// Log writes an audit event.
	Log(ctx context.Context, event Event)

	// LogWithSeverity writes an audit event with explicit severity.
	LogWithSeverity(ctx context.Context, event Event, severity Severity)

	// Query retrieves audit events matching the filter.
	Query(ctx context.Context, filter QueryFilter) ([]Event, error)

	// Count returns the number of events matching the filter.
	Count(ctx context.Context, filter QueryFilter) (int64, error)

	// Flush ensures all buffered events are written.
	Flush(ctx context.Context) error

	// Close closes the audit logger.
	Close() error
}

// ============================================================================
// Query Filter
// ============================================================================

// QueryFilter defines criteria for querying audit events.
type QueryFilter struct {
	// EventTypes filters by event type
	EventTypes []EventType `json:"event_types,omitempty"`

	// Actors filters by actor
	Actors []string `json:"actors,omitempty"`

	// Resources filters by resource
	Resources []string `json:"resources,omitempty"`

	// Severities filters by severity
	Severities []Severity `json:"severities,omitempty"`

	// Outcomes filters by outcome
	Outcomes []EventOutcome `json:"outcomes,omitempty"`

	// StartTime filters events after this time
	StartTime *time.Time `json:"start_time,omitempty"`

	// EndTime filters events before this time
	EndTime *time.Time `json:"end_time,omitempty"`

	// Limit is the maximum number of results
	Limit int `json:"limit,omitempty"`

	// Offset is the number of results to skip
	Offset int `json:"offset,omitempty"`

	// OrderBy specifies the sort field
	OrderBy string `json:"order_by,omitempty"`

	// OrderDesc sorts in descending order
	OrderDesc bool `json:"order_desc,omitempty"`
}

// ============================================================================
// Configuration
// ============================================================================

// Config contains configuration for the audit logger.
type Config struct {
	// Enabled determines if audit logging is active
	Enabled bool `json:"enabled"`

	// Backend specifies the storage backend
	Backend BackendType `json:"backend"`

	// BufferSize is the size of the event buffer
	BufferSize int `json:"buffer_size"`

	// FlushInterval is how often to flush the buffer
	FlushInterval time.Duration `json:"flush_interval"`

	// RetentionDays is how long to retain events
	RetentionDays int `json:"retention_days"`

	// Source is the event source information
	Source EventSource `json:"source"`

	// File contains file backend configuration
	File *FileBackendConfig `json:"file,omitempty"`

	// Redis contains Redis backend configuration
	Redis *RedisBackendConfig `json:"redis,omitempty"`

	// Elasticsearch contains Elasticsearch backend configuration
	Elasticsearch *ElasticsearchConfig `json:"elasticsearch,omitempty"`
}

// BackendType identifies the audit log storage backend.
type BackendType string

const (
	BackendTypeMemory        BackendType = "memory"
	BackendTypeFile          BackendType = "file"
	BackendTypeRedis         BackendType = "redis"
	BackendTypeElasticsearch BackendType = "elasticsearch"
)

// FileBackendConfig contains configuration for file-based audit logging.
type FileBackendConfig struct {
	// Directory is the directory for log files
	Directory string `json:"directory"`

	// Filename is the base filename
	Filename string `json:"filename"`

	// MaxSizeMB is the maximum file size before rotation
	MaxSizeMB int `json:"max_size_mb"`

	// MaxBackups is the maximum number of backup files
	MaxBackups int `json:"max_backups"`

	// Compress enables compression for rotated files
	Compress bool `json:"compress"`
}

// RedisBackendConfig contains configuration for Redis-based audit logging.
type RedisBackendConfig struct {
	// URL is the Redis connection URL
	URL string `json:"url"`

	// Prefix is the key prefix for audit events
	Prefix string `json:"prefix"`

	// MaxEntries is the maximum number of entries to store
	MaxEntries int64 `json:"max_entries"`
}

// ElasticsearchConfig contains configuration for Elasticsearch-based audit logging.
type ElasticsearchConfig struct {
	// Addresses is the list of Elasticsearch addresses
	Addresses []string `json:"addresses"`

	// Index is the index name
	Index string `json:"index"`

	// Username is the authentication username
	Username string `json:"username"`

	// Password is the authentication password
	Password string `json:"password"`
}

// DefaultConfig returns the default audit configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:       true,
		Backend:       BackendTypeMemory,
		BufferSize:    1000,
		FlushInterval: 5 * time.Second,
		RetentionDays: 90,
		Source: EventSource{
			Service:   "verification-signer",
			Component: "audit",
		},
	}
}
