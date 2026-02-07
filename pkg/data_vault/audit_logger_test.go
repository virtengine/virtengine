package data_vault

import (
	"context"
	"testing"
)

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
