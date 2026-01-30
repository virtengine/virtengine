package servicedesk

import (
	"context"
	"sync"
	"time"

	"cosmossdk.io/log"
)

// RetryQueue manages failed sync events for retry
type RetryQueue struct {
	config RetryConfig
	logger log.Logger

	mu      sync.Mutex
	queue   []*SyncEvent
	failed  []*SyncEvent
	running bool
	stopCh  chan struct{}
}

// NewRetryQueue creates a new retry queue
func NewRetryQueue(config RetryConfig, logger log.Logger) *RetryQueue {
	return &RetryQueue{
		config: config,
		logger: logger.With("component", "retry_queue"),
		queue:  make([]*SyncEvent, 0),
		failed: make([]*SyncEvent, 0),
		stopCh: make(chan struct{}),
	}
}

// Add adds an event to the retry queue
func (q *RetryQueue) Add(event *SyncEvent) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Calculate next retry time
	backoff := q.config.CalculateBackoff(event.RetryCount)
	nextRetry := time.Now().Add(backoff)
	event.NextRetryAt = &nextRetry

	q.queue = append(q.queue, event)
	q.logger.Debug("event added to retry queue",
		"event_id", event.ID,
		"retry_count", event.RetryCount,
		"next_retry", nextRetry,
	)
}

// Start starts the retry queue processor
func (q *RetryQueue) Start(ctx context.Context, processor func(context.Context, *SyncEvent) error) {
	q.mu.Lock()
	if q.running {
		q.mu.Unlock()
		return
	}
	q.running = true
	q.mu.Unlock()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-q.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			q.processReadyEvents(ctx, processor)
		}
	}
}

// Stop stops the retry queue
func (q *RetryQueue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.running {
		close(q.stopCh)
		q.running = false
	}
}

// processReadyEvents processes events that are ready for retry
func (q *RetryQueue) processReadyEvents(ctx context.Context, processor func(context.Context, *SyncEvent) error) {
	q.mu.Lock()
	now := time.Now()
	ready := make([]*SyncEvent, 0)
	remaining := make([]*SyncEvent, 0)

	for _, event := range q.queue {
		if event.NextRetryAt != nil && now.After(*event.NextRetryAt) {
			ready = append(ready, event)
		} else {
			remaining = append(remaining, event)
		}
	}
	q.queue = remaining
	q.mu.Unlock()

	// Process ready events
	for _, event := range ready {
		if err := processor(ctx, event); err != nil {
			q.logger.Error("retry failed",
				"event_id", event.ID,
				"retry_count", event.RetryCount,
				"error", err,
			)

			event.RetryCount++
			event.Error = err.Error()

			if event.CanRetry() {
				q.Add(event)
			} else {
				q.mu.Lock()
				event.Status = SyncStatusFailed
				q.failed = append(q.failed, event)
				q.mu.Unlock()
				q.logger.Warn("event failed permanently",
					"event_id", event.ID,
					"ticket_id", event.TicketID,
				)
			}
		} else {
			event.Status = SyncStatusSynced
			q.logger.Info("retry succeeded",
				"event_id", event.ID,
				"retry_count", event.RetryCount,
			)
		}
	}
}

// FailedCount returns the number of permanently failed events
func (q *RetryQueue) FailedCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.failed)
}

// GetFailed returns the failed events
func (q *RetryQueue) GetFailed() []*SyncEvent {
	q.mu.Lock()
	defer q.mu.Unlock()

	result := make([]*SyncEvent, len(q.failed))
	copy(result, q.failed)
	return result
}

// ClearFailed clears the failed events list
func (q *RetryQueue) ClearFailed() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.failed = make([]*SyncEvent, 0)
}

// PendingCount returns the number of pending retry events
func (q *RetryQueue) PendingCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.queue)
}
