package data_vault

import (
	"context"
	"errors"
	"testing"
)

type failOnceAuditStore struct {
	events      []*AuditEvent
	fail        bool
	failAt      int
	appendCalls int
}

func (s *failOnceAuditStore) Append(_ context.Context, event *AuditEvent) error {
	s.appendCalls++
	if s.fail || (s.failAt > 0 && s.appendCalls == s.failAt) {
		s.fail = false
		return errors.New("append failed")
	}
	s.events = append(s.events, cloneAuditEvent(event))
	return nil
}

func (s *failOnceAuditStore) Query(context.Context, AuditFilter) ([]*AuditEvent, error) {
	result := make([]*AuditEvent, 0, len(s.events))
	for _, event := range s.events {
		result = append(result, cloneAuditEvent(event))
	}
	return result, nil
}

func TestAuditLogger_Chaining(t *testing.T) {
	store := NewMemoryAuditStore()
	logger := NewAuditLogger(DefaultAuditLogConfig(), store)

	event1 := &AuditEvent{
		EventType: "read",
		BlobID:    "blob-1",
		Scope:     ScopeSupport,
		Requester: "user1",
		OrgID:     "org-1",
		Success:   true,
	}
	if err := logger.LogEvent(context.Background(), event1); err != nil {
		t.Fatalf("log event1: %v", err)
	}
	if event1.Hash == "" {
		t.Fatalf("expected hash for event1")
	}

	event2 := &AuditEvent{
		EventType: "read",
		BlobID:    "blob-2",
		Scope:     ScopeSupport,
		Requester: "user1",
		OrgID:     "org-1",
		Success:   true,
	}
	if err := logger.LogEvent(context.Background(), event2); err != nil {
		t.Fatalf("log event2: %v", err)
	}
	if event2.PreviousHash != event1.Hash {
		t.Fatalf("expected previous hash to chain")
	}
	if event2.Hash == "" {
		t.Fatalf("expected hash for event2")
	}
}

func TestAuditLogger_Query(t *testing.T) {
	store := NewMemoryAuditStore()
	logger := NewAuditLogger(DefaultAuditLogConfig(), store)

	_ = logger.LogEvent(context.Background(), &AuditEvent{
		EventType: "read",
		BlobID:    "blob-1",
		Scope:     ScopeSupport,
		Requester: "user1",
		OrgID:     "org-1",
		Success:   true,
	})
	_ = logger.LogEvent(context.Background(), &AuditEvent{
		EventType: "read",
		BlobID:    "blob-2",
		Scope:     ScopeMarket,
		Requester: "user2",
		OrgID:     "org-2",
		Success:   true,
	})

	events, err := logger.QueryEvents(context.Background(), AuditFilter{
		Scope:     ScopeSupport,
		Requester: "user1",
	})
	if err != nil {
		t.Fatalf("query events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestAuditLoggerAppendFailureDoesNotAdvanceChain(t *testing.T) {
	store := &failOnceAuditStore{fail: true}
	logger := NewAuditLogger(DefaultAuditLogConfig(), store)
	failed := &AuditEvent{EventType: "upload", Requester: "owner", Metadata: map[string]string{"value": "original"}}
	if err := logger.LogEvent(context.Background(), failed); err == nil {
		t.Fatal("expected append failure")
	}
	failed.Metadata["value"] = "mutated"
	succeeded := &AuditEvent{EventType: "upload", Requester: "owner"}
	if err := logger.LogEvent(context.Background(), succeeded); err != nil {
		t.Fatalf("second append: %v", err)
	}
	if succeeded.PreviousHash != "" {
		t.Fatalf("failed append advanced predecessor to %q", succeeded.PreviousHash)
	}
	queried, err := logger.QueryEvents(context.Background(), AuditFilter{})
	if err != nil {
		t.Fatal(err)
	}
	queried[0].Metadata = map[string]string{"mutated": "true"}
	again, err := logger.QueryEvents(context.Background(), AuditFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := again[0].Metadata["mutated"]; exists {
		t.Fatal("query returned store-owned audit metadata")
	}
}
