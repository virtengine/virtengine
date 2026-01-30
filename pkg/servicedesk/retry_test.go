package servicedesk

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/log"
)

func TestRetryQueue(t *testing.T) {
	config := DefaultRetryConfig()
	queue := NewRetryQueue(config, log.NewNopLogger())

	// Add an event
	event := &SyncEvent{
		ID:         "test-1",
		TicketID:   "TKT-00000001",
		MaxRetries: 3,
		RetryCount: 0,
		Status:     SyncStatusFailed,
	}

	queue.Add(event)

	if queue.PendingCount() != 1 {
		t.Errorf("expected 1 pending event, got %d", queue.PendingCount())
	}

	// Check that retry time was set
	if event.NextRetryAt == nil {
		t.Error("expected NextRetryAt to be set")
	}
}

func TestRetryQueueProcessing(t *testing.T) {
	config := RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}
	queue := NewRetryQueue(config, log.NewNopLogger())

	// Track processing
	processed := make(chan string, 10)
	processor := func(ctx context.Context, event *SyncEvent) error {
		processed <- event.ID
		return nil
	}

	// Add an event with immediate retry
	event := &SyncEvent{
		ID:         "test-1",
		TicketID:   "TKT-00000001",
		MaxRetries: 3,
		RetryCount: 0,
		Status:     SyncStatusFailed,
	}
	now := time.Now().Add(-1 * time.Second) // Already due
	event.NextRetryAt = &now
	
	queue.mu.Lock()
	queue.queue = append(queue.queue, event)
	queue.mu.Unlock()

	// Process ready events
	ctx := context.Background()
	queue.processReadyEvents(ctx, processor)

	// Check it was processed
	select {
	case id := <-processed:
		if id != "test-1" {
			t.Errorf("expected test-1, got %s", id)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected event to be processed")
	}

	// Should no longer be pending
	if queue.PendingCount() != 0 {
		t.Errorf("expected 0 pending after processing, got %d", queue.PendingCount())
	}
}

func TestRetryQueueRequeue(t *testing.T) {
	config := RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}
	queue := NewRetryQueue(config, log.NewNopLogger())

	// Processor that always fails
	failCount := 0
	processor := func(ctx context.Context, event *SyncEvent) error {
		failCount++
		return ErrSyncFailed.Wrap("test error")
	}

	// Add an event with immediate retry
	event := &SyncEvent{
		ID:         "test-1",
		TicketID:   "TKT-00000001",
		MaxRetries: 3,
		RetryCount: 0,
		Status:     SyncStatusPending,
	}
	now := time.Now().Add(-1 * time.Second)
	event.NextRetryAt = &now

	queue.mu.Lock()
	queue.queue = append(queue.queue, event)
	queue.mu.Unlock()

	// Process - should fail and requeue
	ctx := context.Background()
	queue.processReadyEvents(ctx, processor)

	if failCount != 1 {
		t.Errorf("expected 1 fail, got %d", failCount)
	}

	// Should be requeued
	if queue.PendingCount() != 1 {
		t.Errorf("expected 1 pending after failed processing, got %d", queue.PendingCount())
	}

	// Check retry count was incremented
	queue.mu.Lock()
	if len(queue.queue) > 0 && queue.queue[0].RetryCount != 1 {
		t.Errorf("expected RetryCount 1, got %d", queue.queue[0].RetryCount)
	}
	queue.mu.Unlock()
}

func TestRetryQueuePermanentFailure(t *testing.T) {
	config := RetryConfig{
		MaxRetries:        2,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}
	queue := NewRetryQueue(config, log.NewNopLogger())

	// Processor that always fails
	processor := func(ctx context.Context, event *SyncEvent) error {
		return ErrSyncFailed.Wrap("test error")
	}

	// Add an event that has already exhausted retries
	event := &SyncEvent{
		ID:         "test-1",
		TicketID:   "TKT-00000001",
		MaxRetries: 2,
		RetryCount: 2, // Already at max
		Status:     SyncStatusFailed,
	}
	now := time.Now().Add(-1 * time.Second)
	event.NextRetryAt = &now

	queue.mu.Lock()
	queue.queue = append(queue.queue, event)
	queue.mu.Unlock()

	// Process - should fail permanently
	ctx := context.Background()
	queue.processReadyEvents(ctx, processor)

	// Should be in failed list, not pending
	if queue.PendingCount() != 0 {
		t.Errorf("expected 0 pending, got %d", queue.PendingCount())
	}
	if queue.FailedCount() != 1 {
		t.Errorf("expected 1 failed, got %d", queue.FailedCount())
	}

	// Check failed event
	failed := queue.GetFailed()
	if len(failed) != 1 {
		t.Fatalf("expected 1 failed event, got %d", len(failed))
	}
	if failed[0].Status != SyncStatusFailed {
		t.Errorf("expected status Failed, got %s", failed[0].Status)
	}
}

func TestRetryQueueClearFailed(t *testing.T) {
	config := DefaultRetryConfig()
	queue := NewRetryQueue(config, log.NewNopLogger())

	// Add some failed events
	queue.mu.Lock()
	queue.failed = append(queue.failed, &SyncEvent{ID: "failed-1"})
	queue.failed = append(queue.failed, &SyncEvent{ID: "failed-2"})
	queue.mu.Unlock()

	if queue.FailedCount() != 2 {
		t.Errorf("expected 2 failed, got %d", queue.FailedCount())
	}

	queue.ClearFailed()

	if queue.FailedCount() != 0 {
		t.Errorf("expected 0 failed after clear, got %d", queue.FailedCount())
	}
}
