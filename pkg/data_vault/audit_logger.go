package data_vault

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	verrors "github.com/virtengine/virtengine/pkg/errors"
)

// AuditLogConfig controls audit logging behavior.
type AuditLogConfig struct {
	Enabled        bool
	EnableChaining bool
}

// DefaultAuditLogConfig returns default audit log configuration.
func DefaultAuditLogConfig() AuditLogConfig {
	return AuditLogConfig{
		Enabled:        true,
		EnableChaining: true,
	}
}

// AuditExporter receives audit events for external sinks.
type AuditExporter interface {
	Export(ctx context.Context, event *AuditEvent) error
}

// AuditStore stores audit events.
type AuditStore interface {
	Append(ctx context.Context, event *AuditEvent) error
	Query(ctx context.Context, filter AuditFilter) ([]*AuditEvent, error)
}

// MemoryAuditStore keeps audit events in memory.
type MemoryAuditStore struct {
	mu     sync.RWMutex
	events []*AuditEvent
}

// NewMemoryAuditStore creates an in-memory audit store.
func NewMemoryAuditStore() *MemoryAuditStore {
	return &MemoryAuditStore{
		events: make([]*AuditEvent, 0),
	}
}

// Append stores a new event.
func (s *MemoryAuditStore) Append(_ context.Context, event *AuditEvent) error {
	if event == nil {
		return errors.New("audit event required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, cloneAuditEvent(event))
	return nil
}

// Query returns events matching the filter.
func (s *MemoryAuditStore) Query(_ context.Context, filter AuditFilter) ([]*AuditEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*AuditEvent, 0, len(s.events))
	for _, event := range s.events {
		if filter.BlobID != "" && event.BlobID != filter.BlobID {
			continue
		}
		if filter.Scope != "" && event.Scope != filter.Scope {
			continue
		}
		if filter.Requester != "" && event.Requester != filter.Requester {
			continue
		}
		if filter.OrgID != "" && event.OrgID != filter.OrgID {
			continue
		}
		if filter.StartTime != nil && event.Timestamp.Unix() < *filter.StartTime {
			continue
		}
		if filter.EndTime != nil && event.Timestamp.Unix() > *filter.EndTime {
			continue
		}
		result = append(result, cloneAuditEvent(event))
		if filter.Limit > 0 && len(result) >= filter.Limit {
			break
		}
	}

	return result, nil
}

// AuditLogger appends immutable audit events with hash chaining.
type AuditLogger struct {
	cfg       AuditLogConfig
	store     AuditStore
	exporters []AuditExporter

	mu       sync.Mutex
	lastHash string
	initErr  error
}

// NewAuditLogger creates a new audit logger with a memory store by default.
func NewAuditLogger(cfg AuditLogConfig, store AuditStore) *AuditLogger {
	if store == nil {
		store = NewMemoryAuditStore()
	}
	if cfg == (AuditLogConfig{}) {
		cfg = DefaultAuditLogConfig()
	}
	logger := &AuditLogger{
		cfg:   cfg,
		store: store,
	}
	if events, err := store.Query(context.Background(), AuditFilter{}); err != nil {
		logger.initErr = fmt.Errorf("load audit chain: %w", err)
	} else if err := validateAuditEvents(events); err != nil {
		logger.initErr = err
	} else if len(events) > 0 {
		logger.lastHash = events[len(events)-1].Hash
	}
	return logger
}

// RegisterExporter registers an exporter for audit events.
func (l *AuditLogger) RegisterExporter(exporter AuditExporter) {
	if exporter == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.exporters = append(l.exporters, exporter)
}

// LogEvent appends an audit event to the store and exporters.
func (l *AuditLogger) LogEvent(ctx context.Context, event *AuditEvent) error {
	if event == nil {
		return NewVaultError("AuditLog", ErrInvalidRequest, "event required")
	}
	if !l.cfg.Enabled {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if l.initErr != nil {
		return l.initErr
	}

	stored := cloneAuditEvent(event)

	if stored.Timestamp.IsZero() {
		stored.Timestamp = time.Now().UTC()
	}
	if stored.ID == "" {
		stored.ID = generateAuditEventID(stored)
	}

	if l.cfg.EnableChaining {
		stored.PreviousHash = l.lastHash
		stored.Hash = computeAuditHash(stored)
	} else {
		stored.Hash = computeAuditHash(stored)
	}

	if err := l.store.Append(ctx, stored); err != nil {
		return err
	}
	if l.cfg.EnableChaining {
		l.lastHash = stored.Hash
	}
	*event = *cloneAuditEvent(stored)

	for _, exporter := range l.exporters {
		exp := exporter
		exported := cloneAuditEvent(stored)
		verrors.SafeGo("data-vault:audit-export", func() {
			_ = exp.Export(ctx, exported)
		})
	}

	return nil
}

func cloneAuditEvent(event *AuditEvent) *AuditEvent {
	if event == nil {
		return nil
	}
	cloned := *event
	cloned.Metadata = cloneStringMap(event.Metadata)
	return &cloned
}

func validateAuditEvents(events []*AuditEvent) error {
	previous := ""
	for index, event := range events {
		if event == nil {
			return fmt.Errorf("audit chain event %d is nil", index)
		}
		if event.PreviousHash != previous {
			return fmt.Errorf("audit chain event %d predecessor mismatch", index)
		}
		if event.Hash == "" || event.Hash != computeAuditHash(event) {
			return fmt.Errorf("audit chain event %d hash mismatch", index)
		}
		previous = event.Hash
	}
	return nil
}

func cloneStringMap(source map[string]string) map[string]string {
	if source == nil {
		return nil
	}
	cloned := make(map[string]string, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

// QueryEvents returns audit events matching the filter.
func (l *AuditLogger) QueryEvents(ctx context.Context, filter AuditFilter) ([]*AuditEvent, error) {
	return l.store.Query(ctx, filter)
}

// Close releases resources held by a durable audit store.
func (l *AuditLogger) Close() error {
	if closer, ok := l.store.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

func generateAuditEventID(event *AuditEvent) string {
	hash := sha256.Sum256([]byte(event.EventType + ":" + string(event.BlobID) + ":" + event.Requester + ":" + event.Timestamp.UTC().String()))
	return hex.EncodeToString(hash[:8])
}

type auditMetadataEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type auditHashPayload struct {
	ID           string               `json:"id"`
	EventType    string               `json:"event_type"`
	BlobID       BlobID               `json:"blob_id"`
	Scope        Scope                `json:"scope"`
	Requester    string               `json:"requester"`
	OrgID        string               `json:"org_id,omitempty"`
	Success      bool                 `json:"success"`
	Error        string               `json:"error,omitempty"`
	Timestamp    time.Time            `json:"timestamp"`
	PreviousHash string               `json:"previous_hash,omitempty"`
	Metadata     []auditMetadataEntry `json:"metadata,omitempty"`
}

func computeAuditHash(event *AuditEvent) string {
	entries := make([]auditMetadataEntry, 0, len(event.Metadata))
	for k, v := range event.Metadata {
		entries = append(entries, auditMetadataEntry{Key: k, Value: v})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Key < entries[j].Key
	})

	payload := auditHashPayload{
		ID:           event.ID,
		EventType:    event.EventType,
		BlobID:       event.BlobID,
		Scope:        event.Scope,
		Requester:    event.Requester,
		OrgID:        event.OrgID,
		Success:      event.Success,
		Error:        event.Error,
		Timestamp:    event.Timestamp.UTC(),
		PreviousHash: event.PreviousHash,
		Metadata:     entries,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
