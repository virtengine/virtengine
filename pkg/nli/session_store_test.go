package nli

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestInMemorySessionStore_BasicOperations(t *testing.T) {
	config := DefaultSessionStoreConfig()
	config.Backend = "memory"
	config.MaxSessions = 100
	config.SessionTTL = time.Minute

	store := NewInMemorySessionStore(config, zerolog.Nop())
	defer store.Close()

	ctx := context.Background()

	// Test Set and Get
	session := &Session{
		ID:          "test-session-1",
		UserAddress: "virtengine1abc...",
		History: []ChatMessage{
			{Role: "user", Content: "Hello", Timestamp: time.Now()},
		},
		CreatedAt: time.Now(),
	}

	err := store.Set(ctx, session)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	retrieved, err := store.Get(ctx, "test-session-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.ID != session.ID {
		t.Errorf("Expected ID %s, got %s", session.ID, retrieved.ID)
	}

	if retrieved.UserAddress != session.UserAddress {
		t.Errorf("Expected UserAddress %s, got %s", session.UserAddress, retrieved.UserAddress)
	}

	if len(retrieved.History) != 1 {
		t.Errorf("Expected 1 history item, got %d", len(retrieved.History))
	}
}

func TestInMemorySessionStore_SessionNotFound(t *testing.T) {
	config := DefaultSessionStoreConfig()
	config.Backend = "memory"

	store := NewInMemorySessionStore(config, zerolog.Nop())
	defer store.Close()

	ctx := context.Background()

	_, err := store.Get(ctx, "nonexistent-session")
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}
}

func TestInMemorySessionStore_Count(t *testing.T) {
	config := DefaultSessionStoreConfig()
	config.Backend = "memory"

	store := NewInMemorySessionStore(config, zerolog.Nop())
	defer store.Close()

	ctx := context.Background()

	// Initially empty
	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 sessions, got %d", count)
	}

	// Add sessions
	for i := 0; i < 5; i++ {
		session := &Session{
			ID:        "session-" + string(rune('a'+i)),
			CreatedAt: time.Now(),
		}
		if err := store.Set(ctx, session); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	count, err = store.Count(ctx)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 5 {
		t.Errorf("Expected 5 sessions, got %d", count)
	}
}

func TestInMemorySessionStore_Delete(t *testing.T) {
	config := DefaultSessionStoreConfig()
	config.Backend = "memory"

	store := NewInMemorySessionStore(config, zerolog.Nop())
	defer store.Close()

	ctx := context.Background()

	// Add a session
	session := &Session{
		ID:        "delete-test",
		CreatedAt: time.Now(),
	}
	if err := store.Set(ctx, session); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Delete it
	if err := store.Delete(ctx, "delete-test"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	_, err := store.Get(ctx, "delete-test")
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}
}

func TestInMemorySessionStore_Touch(t *testing.T) {
	config := DefaultSessionStoreConfig()
	config.Backend = "memory"
	config.SessionTTL = time.Hour

	store := NewInMemorySessionStore(config, zerolog.Nop())
	defer store.Close()

	ctx := context.Background()

	// Add a session
	session := &Session{
		ID:           "touch-test",
		LastActivity: time.Now().Add(-time.Hour),
		CreatedAt:    time.Now().Add(-time.Hour),
	}
	if err := store.Set(ctx, session); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Touch it
	if err := store.Touch(ctx, "touch-test"); err != nil {
		t.Fatalf("Touch failed: %v", err)
	}

	// Verify last activity is updated
	retrieved, err := store.Get(ctx, "touch-test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if time.Since(retrieved.LastActivity) > time.Second {
		t.Errorf("LastActivity not updated, got %v", retrieved.LastActivity)
	}
}

func TestInMemorySessionStore_MaxSessions(t *testing.T) {
	config := DefaultSessionStoreConfig()
	config.Backend = "memory"
	config.MaxSessions = 10

	store := NewInMemorySessionStore(config, zerolog.Nop())
	defer store.Close()

	ctx := context.Background()

	// Add more sessions than the limit
	for i := 0; i < 15; i++ {
		session := &Session{
			ID:        "session-" + string(rune('0'+i)),
			CreatedAt: time.Now(),
		}
		// Sleep briefly to ensure different timestamps
		time.Sleep(time.Millisecond)
		if err := store.Set(ctx, session); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	// Count should not exceed max
	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count > int64(config.MaxSessions) {
		t.Errorf("Expected at most %d sessions, got %d", config.MaxSessions, count)
	}
}

func TestInMemorySessionStore_TTLExpiry(t *testing.T) {
	config := DefaultSessionStoreConfig()
	config.Backend = "memory"
	config.SessionTTL = 50 * time.Millisecond

	store := NewInMemorySessionStore(config, zerolog.Nop())
	defer store.Close()

	ctx := context.Background()

	// Add a session
	session := &Session{
		ID:        "expiry-test",
		CreatedAt: time.Now(),
	}
	if err := store.Set(ctx, session); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Verify it exists
	_, err := store.Get(ctx, "expiry-test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Wait for expiry
	time.Sleep(100 * time.Millisecond)

	// Should be expired now
	_, err = store.Get(ctx, "expiry-test")
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound after TTL, got %v", err)
	}
}

func TestInMemorySessionStore_HistoryTrimming(t *testing.T) {
	config := DefaultSessionStoreConfig()
	config.Backend = "memory"
	config.MaxHistoryLength = 5

	store := NewInMemorySessionStore(config, zerolog.Nop())
	defer store.Close()

	ctx := context.Background()

	// Create session with too many history items
	history := make([]ChatMessage, 20)
	for i := 0; i < 20; i++ {
		history[i] = ChatMessage{
			Role:      "user",
			Content:   "message",
			Timestamp: time.Now(),
		}
	}

	session := &Session{
		ID:        "history-test",
		History:   history,
		CreatedAt: time.Now(),
	}
	if err := store.Set(ctx, session); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Retrieve and check trimmed
	retrieved, err := store.Get(ctx, "history-test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// History should be trimmed to MaxHistoryLength*2
	if len(retrieved.History) > config.MaxHistoryLength*2 {
		t.Errorf("Expected at most %d history items, got %d", config.MaxHistoryLength*2, len(retrieved.History))
	}
}

func TestInMemorySessionStore_Metrics(t *testing.T) {
	config := DefaultSessionStoreConfig()
	config.Backend = "memory"

	store := NewInMemorySessionStore(config, zerolog.Nop())
	defer store.Close()

	ctx := context.Background()

	// Perform operations
	session := &Session{ID: "metrics-test", CreatedAt: time.Now()}
	_ = store.Set(ctx, session)
	_, _ = store.Get(ctx, "metrics-test")
	_, _ = store.Get(ctx, "nonexistent") // Miss
	_ = store.Delete(ctx, "metrics-test")

	metrics := store.GetMetrics()

	if metrics.Sets != 1 {
		t.Errorf("Expected 1 set, got %d", metrics.Sets)
	}
	if metrics.Gets != 2 {
		t.Errorf("Expected 2 gets, got %d", metrics.Gets)
	}
	if metrics.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", metrics.Hits)
	}
	if metrics.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", metrics.Misses)
	}
	if metrics.Deletes != 1 {
		t.Errorf("Expected 1 delete, got %d", metrics.Deletes)
	}
}

func TestNewSessionStore_Memory(t *testing.T) {
	config := DefaultSessionStoreConfig()
	config.Backend = "memory"

	store, err := NewSessionStore(context.Background(), config, zerolog.Nop())
	if err != nil {
		t.Fatalf("NewSessionStore failed: %v", err)
	}
	defer store.Close()

	// Verify it works
	ctx := context.Background()
	session := &Session{ID: "factory-test", CreatedAt: time.Now()}
	if err := store.Set(ctx, session); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	retrieved, err := store.Get(ctx, "factory-test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved.ID != "factory-test" {
		t.Errorf("Expected ID 'factory-test', got %s", retrieved.ID)
	}
}

func TestNewSessionStore_UnknownBackend(t *testing.T) {
	config := DefaultSessionStoreConfig()
	config.Backend = "unknown"

	_, err := NewSessionStore(context.Background(), config, zerolog.Nop())
	if err == nil {
		t.Error("Expected error for unknown backend")
	}
}

func TestDefaultSessionStoreConfig(t *testing.T) {
	config := DefaultSessionStoreConfig()

	if config.Backend != "memory" {
		t.Errorf("Expected backend 'memory', got %s", config.Backend)
	}
	if config.SessionTTL != 30*time.Minute {
		t.Errorf("Expected TTL 30m, got %v", config.SessionTTL)
	}
	if config.MaxSessions != 10000 {
		t.Errorf("Expected MaxSessions 10000, got %d", config.MaxSessions)
	}
	if config.RedisPrefix != "virtengine:nli:session:" {
		t.Errorf("Expected prefix 'virtengine:nli:session:', got %s", config.RedisPrefix)
	}
}
