package servicedesk

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/log"
)

func TestAuditLogger(t *testing.T) {
	config := AuditConfig{
		Enabled:       true,
		LogLevel:      "info",
		RetentionDays: 30,
		LogSensitive:  false,
	}

	logger := NewAuditLogger(config, log.NewNopLogger())

	ctx := context.Background()

	// Log an event
	logger.LogEvent(ctx, AuditEventTicketCreate, map[string]interface{}{
		"ticket_id": "TKT-00000001",
		"customer":  "cosmos1abc...",
		"category":  "technical",
	})

	// Check entry was recorded
	entries := logger.GetEntriesByTicket("TKT-00000001")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.EventType != AuditEventTicketCreate {
		t.Errorf("expected event type %s, got %s", AuditEventTicketCreate, entry.EventType)
	}
	if entry.TicketID != "TKT-00000001" {
		t.Errorf("expected ticket ID TKT-00000001, got %s", entry.TicketID)
	}
	if entry.Status != "success" {
		t.Errorf("expected status success, got %s", entry.Status)
	}
}

func TestAuditLoggerError(t *testing.T) {
	config := DefaultAuditConfig()
	logger := NewAuditLogger(config, log.NewNopLogger())

	ctx := context.Background()

	// Log an error event
	logger.LogEvent(ctx, AuditEventSyncFailed, map[string]interface{}{
		"ticket_id": "TKT-00000002",
		"error":     "connection timeout",
	})

	entries := logger.GetEntriesByTicket("TKT-00000002")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Status != "failed" {
		t.Errorf("expected status failed, got %s", entry.Status)
	}
	if entry.Error != "connection timeout" {
		t.Errorf("expected error 'connection timeout', got %s", entry.Error)
	}
}

func TestAuditLoggerDisabled(t *testing.T) {
	config := AuditConfig{
		Enabled: false,
	}
	logger := NewAuditLogger(config, log.NewNopLogger())

	ctx := context.Background()

	logger.LogEvent(ctx, AuditEventTicketCreate, map[string]interface{}{
		"ticket_id": "TKT-00000001",
	})

	// Should not record when disabled
	entries := logger.GetEntriesByTicket("TKT-00000001")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries when disabled, got %d", len(entries))
	}
}

func TestAuditFilter(t *testing.T) {
	config := DefaultAuditConfig()
	logger := NewAuditLogger(config, log.NewNopLogger())

	ctx := context.Background()

	// Log multiple events
	logger.LogEvent(ctx, AuditEventTicketCreate, map[string]interface{}{
		"ticket_id":    "TKT-00000001",
		"service_desk": ServiceDeskJira,
	})
	logger.LogEvent(ctx, AuditEventSyncSuccess, map[string]interface{}{
		"ticket_id":    "TKT-00000001",
		"service_desk": ServiceDeskJira,
	})
	logger.LogEvent(ctx, AuditEventTicketCreate, map[string]interface{}{
		"ticket_id":    "TKT-00000002",
		"service_desk": ServiceDeskWaldur,
	})

	// Filter by ticket ID
	entries := logger.GetEntries(AuditFilter{TicketID: "TKT-00000001"})
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for TKT-00000001, got %d", len(entries))
	}

	// Filter by event type
	entries = logger.GetEntries(AuditFilter{EventType: AuditEventTicketCreate})
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for ticket_create, got %d", len(entries))
	}

	// Filter by service desk
	entries = logger.GetEntries(AuditFilter{ServiceDesk: ServiceDeskJira})
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for jira, got %d", len(entries))
	}

	// Filter by service desk waldur
	entries = logger.GetEntries(AuditFilter{ServiceDesk: ServiceDeskWaldur})
	if len(entries) != 1 {
		t.Errorf("expected 1 entry for waldur, got %d", len(entries))
	}
}

func TestAuditPurge(t *testing.T) {
	config := AuditConfig{
		Enabled:       true,
		RetentionDays: 1, // 1 day retention
	}
	logger := NewAuditLogger(config, log.NewNopLogger())

	// Manually add an old entry
	logger.mu.Lock()
	logger.entries = append(logger.entries, AuditEntry{
		ID:        "old-1",
		Timestamp: time.Now().AddDate(0, 0, -5), // 5 days old
		EventType: AuditEventTicketCreate,
		TicketID:  "TKT-OLD",
		Status:    "success",
	})
	logger.mu.Unlock()

	ctx := context.Background()
	logger.LogEvent(ctx, AuditEventTicketCreate, map[string]interface{}{
		"ticket_id": "TKT-NEW",
	})

	// Should have 2 entries before purge
	if logger.EntryCount() != 2 {
		t.Errorf("expected 2 entries before purge, got %d", logger.EntryCount())
	}

	// Purge old entries
	purged := logger.PurgeOldEntries()
	if purged != 1 {
		t.Errorf("expected to purge 1 entry, purged %d", purged)
	}

	// Should have 1 entry after purge
	if logger.EntryCount() != 1 {
		t.Errorf("expected 1 entry after purge, got %d", logger.EntryCount())
	}

	// Old entry should be gone
	entries := logger.GetEntriesByTicket("TKT-OLD")
	if len(entries) != 0 {
		t.Error("expected old entry to be purged")
	}

	// New entry should remain
	entries = logger.GetEntriesByTicket("TKT-NEW")
	if len(entries) != 1 {
		t.Error("expected new entry to remain")
	}
}

func TestAuditExport(t *testing.T) {
	config := DefaultAuditConfig()
	logger := NewAuditLogger(config, log.NewNopLogger())

	ctx := context.Background()
	logger.LogEvent(ctx, AuditEventTicketCreate, map[string]interface{}{
		"ticket_id": "TKT-00000001",
	})

	data, err := logger.ExportAuditLog(AuditFilter{})
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty export")
	}

	// Should be valid JSON
	if data[0] != '[' {
		t.Error("expected JSON array")
	}
}

