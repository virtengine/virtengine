// Package audit provides audit logging for verification services.
package audit

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// MemoryLogger implements AuditLogger using in-memory storage.
// Primarily for testing; not suitable for production.
type MemoryLogger struct {
	mu      sync.RWMutex
	events  []Event
	maxSize int
	config  Config
	logger  zerolog.Logger
	closed  bool
}

// NewMemoryLogger creates a new in-memory audit logger.
func NewMemoryLogger(config Config, logger zerolog.Logger) *MemoryLogger {
	maxSize := 10000
	if config.BufferSize > 0 {
		maxSize = config.BufferSize
	}

	return &MemoryLogger{
		events:  make([]Event, 0, maxSize),
		maxSize: maxSize,
		config:  config,
		logger:  logger.With().Str("component", "audit").Logger(),
	}
}

// Log writes an audit event.
func (m *MemoryLogger) Log(ctx context.Context, event Event) {
	m.LogWithSeverity(ctx, event, EventSeverity(event.Type))
}

// LogWithSeverity writes an audit event with explicit severity.
func (m *MemoryLogger) LogWithSeverity(ctx context.Context, event Event, severity Severity) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed || !m.config.Enabled {
		return
	}

	// Fill in defaults
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.Severity == "" {
		event.Severity = severity
	}
	if event.Outcome == "" {
		event.Outcome = OutcomeSuccess
	}
	if event.Source == nil {
		event.Source = &m.config.Source
	}

	// Add to buffer
	if len(m.events) >= m.maxSize {
		// Remove oldest event
		m.events = m.events[1:]
	}
	m.events = append(m.events, event)

	// Log to zerolog as well
	m.logEvent(event)
}

// Query retrieves audit events matching the filter.
func (m *MemoryLogger) Query(ctx context.Context, filter QueryFilter) ([]Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []Event
	for _, event := range m.events {
		if m.matchesFilter(event, filter) {
			results = append(results, event)
		}
	}

	// Apply pagination
	if filter.Offset > 0 && filter.Offset < len(results) {
		results = results[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(results) {
		results = results[:filter.Limit]
	}

	return results, nil
}

// Count returns the number of events matching the filter.
func (m *MemoryLogger) Count(ctx context.Context, filter QueryFilter) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var count int64
	for _, event := range m.events {
		if m.matchesFilter(event, filter) {
			count++
		}
	}

	return count, nil
}

// Flush is a no-op for memory logger.
func (m *MemoryLogger) Flush(ctx context.Context) error {
	return nil
}

// Close closes the audit logger.
func (m *MemoryLogger) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	m.events = nil
	return nil
}

// GetEvents returns all events (for testing).
func (m *MemoryLogger) GetEvents() []Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Event, len(m.events))
	copy(result, m.events)
	return result
}

// Clear removes all events (for testing).
func (m *MemoryLogger) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.events = m.events[:0]
}

func (m *MemoryLogger) matchesFilter(event Event, filter QueryFilter) bool {
	// Check event types
	if len(filter.EventTypes) > 0 {
		found := false
		for _, t := range filter.EventTypes {
			if event.Type == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check actors
	if len(filter.Actors) > 0 {
		found := false
		for _, a := range filter.Actors {
			if event.Actor == a {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check resources
	if len(filter.Resources) > 0 {
		found := false
		for _, r := range filter.Resources {
			if event.Resource == r {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check severities
	if len(filter.Severities) > 0 {
		found := false
		for _, s := range filter.Severities {
			if event.Severity == s {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check outcomes
	if len(filter.Outcomes) > 0 {
		found := false
		for _, o := range filter.Outcomes {
			if event.Outcome == o {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check time range
	if filter.StartTime != nil && event.Timestamp.Before(*filter.StartTime) {
		return false
	}
	if filter.EndTime != nil && event.Timestamp.After(*filter.EndTime) {
		return false
	}

	return true
}

func (m *MemoryLogger) logEvent(event Event) {
	var logEvent *zerolog.Event

	switch event.Severity {
	case SeverityCritical:
		logEvent = m.logger.Error()
	case SeverityError:
		logEvent = m.logger.Error()
	case SeverityWarning:
		logEvent = m.logger.Warn()
	case SeverityDebug:
		logEvent = m.logger.Debug()
	default:
		logEvent = m.logger.Info()
	}

	logEvent.
		Str("event_id", event.ID).
		Str("type", string(event.Type)).
		Str("actor", event.Actor).
		Str("resource", event.Resource).
		Str("action", event.Action).
		Str("outcome", string(event.Outcome)).
		Interface("details", event.Details).
		Msg("audit event")
}

// Ensure MemoryLogger implements AuditLogger
var _ AuditLogger = (*MemoryLogger)(nil)

// ============================================================================
// File Logger
// ============================================================================

// FileLogger implements AuditLogger using append-only files.
type FileLogger struct {
	mu      sync.Mutex
	config  Config
	file    *os.File
	encoder *json.Encoder
	logger  zerolog.Logger
	closed  bool
}

// NewFileLogger creates a new file-based audit logger.
func NewFileLogger(config Config, logger zerolog.Logger) (*FileLogger, error) {
	if config.File == nil {
		return nil, ErrInvalidConfig.Wrap("file configuration is required")
	}

	// Create directory if needed
	if err := os.MkdirAll(config.File.Directory, 0700); err != nil {
		return nil, ErrStorageError.Wrapf("failed to create directory: %v", err)
	}

	// Open file in append mode
	filepath := config.File.Directory + "/" + config.File.Filename
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, ErrStorageError.Wrapf("failed to open file: %v", err)
	}

	return &FileLogger{
		config:  config,
		file:    file,
		encoder: json.NewEncoder(file),
		logger:  logger.With().Str("component", "audit").Logger(),
	}, nil
}

// Log writes an audit event.
func (f *FileLogger) Log(ctx context.Context, event Event) {
	f.LogWithSeverity(ctx, event, EventSeverity(event.Type))
}

// LogWithSeverity writes an audit event with explicit severity.
func (f *FileLogger) LogWithSeverity(ctx context.Context, event Event, severity Severity) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed || !f.config.Enabled {
		return
	}

	// Fill in defaults
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.Severity == "" {
		event.Severity = severity
	}
	if event.Outcome == "" {
		event.Outcome = OutcomeSuccess
	}
	if event.Source == nil {
		event.Source = &f.config.Source
	}

	// Write to file
	if err := f.encoder.Encode(event); err != nil {
		f.logger.Error().Err(err).Msg("failed to write audit event")
	}
}

// Query is not efficiently supported by file logger.
func (f *FileLogger) Query(ctx context.Context, filter QueryFilter) ([]Event, error) {
	return nil, ErrUnsupportedOperation.Wrap("query not supported by file logger")
}

// Count is not supported by file logger.
func (f *FileLogger) Count(ctx context.Context, filter QueryFilter) (int64, error) {
	return 0, ErrUnsupportedOperation.Wrap("count not supported by file logger")
}

// Flush syncs the file to disk.
func (f *FileLogger) Flush(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return nil
	}

	return f.file.Sync()
}

// Close closes the file logger.
func (f *FileLogger) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return nil
	}

	f.closed = true
	return f.file.Close()
}

// Ensure FileLogger implements AuditLogger
var _ AuditLogger = (*FileLogger)(nil)
