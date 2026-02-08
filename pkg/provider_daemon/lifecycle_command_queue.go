// Package provider_daemon implements provider-side services for VirtEngine.
//
// VE-34E: Durable lifecycle command queue for provider daemon operations.
package provider_daemon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// LifecycleCommandStatus represents the lifecycle command queue state.
type LifecycleCommandStatus string

const (
	// LifecycleCommandStatusPending indicates command is ready to be processed.
	LifecycleCommandStatusPending LifecycleCommandStatus = "pending"
	// LifecycleCommandStatusExecuting indicates command is being executed.
	LifecycleCommandStatusExecuting LifecycleCommandStatus = "executing"
	// LifecycleCommandStatusSucceeded indicates command completed successfully.
	LifecycleCommandStatusSucceeded LifecycleCommandStatus = "succeeded"
	// LifecycleCommandStatusFailed indicates command failed permanently.
	LifecycleCommandStatusFailed LifecycleCommandStatus = "failed"
	// LifecycleCommandStatusDeadLettered indicates command exceeded max retries.
	LifecycleCommandStatusDeadLettered LifecycleCommandStatus = "dead_lettered"
)

// LifecycleCommand represents a durable lifecycle action request.
type LifecycleCommand struct {
	ID             string                          `json:"id"`
	IdempotencyKey string                          `json:"idempotency_key"`
	AllocationID   string                          `json:"allocation_id"`
	OrderID        string                          `json:"order_id,omitempty"`
	OperationID    string                          `json:"operation_id,omitempty"`
	Provider       string                          `json:"provider_address"`
	ResourceUUID   string                          `json:"resource_uuid,omitempty"`
	Action         marketplace.LifecycleActionType `json:"action"`
	TargetState    marketplace.AllocationState     `json:"target_state"`
	RequestedBy    string                          `json:"requested_by"`
	Parameters     map[string]interface{}          `json:"parameters,omitempty"`
	RollbackPolicy marketplace.RollbackPolicy      `json:"rollback_policy"`
	Status         LifecycleCommandStatus          `json:"status"`
	AttemptCount   int                             `json:"attempt_count"`
	MaxAttempts    int                             `json:"max_attempts"`
	LastAttemptAt  *time.Time                      `json:"last_attempt_at,omitempty"`
	NextAttemptAt  *time.Time                      `json:"next_attempt_at,omitempty"`
	CreatedAt      time.Time                       `json:"created_at"`
	UpdatedAt      time.Time                       `json:"updated_at"`
	CompletedAt    *time.Time                      `json:"completed_at,omitempty"`
	LastError      string                          `json:"last_error,omitempty"`
	WaldurOpID     string                          `json:"waldur_operation_id,omitempty"`
	EventID        string                          `json:"event_id,omitempty"`
	BlockHeight    int64                           `json:"block_height,omitempty"`
	Sequence       uint64                          `json:"sequence,omitempty"`
	EventTimestamp time.Time                       `json:"event_timestamp"`
	Reconcile      bool                            `json:"reconcile"`
}

// LifecycleDesiredState tracks desired state for reconciliation.
type LifecycleDesiredState struct {
	AllocationID    string                          `json:"allocation_id"`
	ResourceUUID    string                          `json:"resource_uuid,omitempty"`
	DesiredState    marketplace.AllocationState     `json:"desired_state"`
	LastAction      marketplace.LifecycleActionType `json:"last_action"`
	LastCommandID   string                          `json:"last_command_id"`
	LastObserved    marketplace.AllocationState     `json:"last_observed_state,omitempty"`
	UpdatedAt       time.Time                       `json:"updated_at"`
	LastReconciled  *time.Time                      `json:"last_reconciled_at,omitempty"`
	LastObservation *time.Time                      `json:"last_observation_at,omitempty"`
}

// LifecycleCommandQueueConfig configures the durable command queue.
type LifecycleCommandQueueConfig struct {
	Enabled           bool
	Backend           string
	Path              string
	InMemory          bool
	WorkerCount       int
	MaxRetries        int
	RetryBackoff      time.Duration
	MaxBackoff        time.Duration
	PollInterval      time.Duration
	ReconcileInterval time.Duration
	ReconcileOnStart  bool
	StaleAfter        time.Duration
	ExecuteTimeout    time.Duration
	CallbackOnSuccess bool
	CallbackOnFailure bool
}

// DefaultLifecycleCommandQueueConfig returns defaults.
func DefaultLifecycleCommandQueueConfig() LifecycleCommandQueueConfig {
	return LifecycleCommandQueueConfig{
		Enabled:           true,
		Backend:           "badger",
		Path:              "data/lifecycle_queue",
		WorkerCount:       2,
		MaxRetries:        5,
		RetryBackoff:      10 * time.Second,
		MaxBackoff:        5 * time.Minute,
		PollInterval:      2 * time.Second,
		ReconcileInterval: 5 * time.Minute,
		ReconcileOnStart:  true,
		StaleAfter:        20 * time.Minute,
		ExecuteTimeout:    2 * time.Minute,
		CallbackOnSuccess: true,
		CallbackOnFailure: true,
	}
}

// LifecycleCommandExecutionResult captures executor output.
type LifecycleCommandExecutionResult struct {
	WaldurOperationID string
	ResourceState     string
}

// LifecycleCommandExecutor executes lifecycle commands against the backend.
type LifecycleCommandExecutor interface {
	Execute(ctx context.Context, cmd *LifecycleCommand) (*LifecycleCommandExecutionResult, error)
	GetResourceState(ctx context.Context, resourceUUID string) (waldur.ResourceState, error)
}

// LifecycleCommandCallback publishes command completion callbacks.
type LifecycleCommandCallback func(ctx context.Context, cmd *LifecycleCommand, err error) error

// LifecycleResourceResolver resolves allocation IDs to resource UUIDs.
type LifecycleResourceResolver func(ctx context.Context, allocationID string) (string, error)

// LifecycleCommandStore persists lifecycle commands and desired state.
type LifecycleCommandStore interface {
	Enqueue(ctx context.Context, cmd *LifecycleCommand) (*LifecycleCommand, bool, error)
	Get(ctx context.Context, id string) (*LifecycleCommand, error)
	Update(ctx context.Context, cmd *LifecycleCommand) error
	ClaimNextReady(ctx context.Context, now time.Time, workerID string) (*LifecycleCommand, error)
	ListByStatus(ctx context.Context, statuses ...LifecycleCommandStatus) ([]*LifecycleCommand, error)
	ListByAllocation(ctx context.Context, allocationID string, statuses ...LifecycleCommandStatus) ([]*LifecycleCommand, error)
	ListDesiredStates(ctx context.Context) ([]*LifecycleDesiredState, error)
	GetDesiredState(ctx context.Context, allocationID string) (*LifecycleDesiredState, error)
	SetDesiredState(ctx context.Context, state *LifecycleDesiredState) error
	Close() error
}

// LifecycleCommandQueue manages durable lifecycle command processing.
type LifecycleCommandQueue struct {
	cfg      LifecycleCommandQueueConfig
	store    LifecycleCommandStore
	executor LifecycleCommandExecutor
	resolver LifecycleResourceResolver
	callback LifecycleCommandCallback
	metrics  *LifecycleQueueMetrics

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewLifecycleCommandQueue creates a new command queue.
func NewLifecycleCommandQueue(
	cfg LifecycleCommandQueueConfig,
	store LifecycleCommandStore,
	executor LifecycleCommandExecutor,
	resolver LifecycleResourceResolver,
	callback LifecycleCommandCallback,
	metrics *LifecycleQueueMetrics,
) (*LifecycleCommandQueue, error) {
	if store == nil {
		return nil, errors.New("lifecycle command store is required")
	}
	if executor == nil {
		return nil, errors.New("lifecycle command executor is required")
	}
	if metrics == nil {
		metrics = NewLifecycleQueueMetrics()
	}

	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 1
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = 10 * time.Second
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = 5 * time.Minute
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 2 * time.Second
	}
	if cfg.ReconcileInterval <= 0 {
		cfg.ReconcileInterval = 5 * time.Minute
	}
	if cfg.StaleAfter <= 0 {
		cfg.StaleAfter = 20 * time.Minute
	}
	if cfg.ExecuteTimeout <= 0 {
		cfg.ExecuteTimeout = 2 * time.Minute
	}

	return &LifecycleCommandQueue{
		cfg:      cfg,
		store:    store,
		executor: executor,
		resolver: resolver,
		callback: callback,
		metrics:  metrics,
		stopCh:   make(chan struct{}),
	}, nil
}

// Start begins worker and reconciliation loops.
func (q *LifecycleCommandQueue) Start(ctx context.Context) error {
	if !q.cfg.Enabled {
		return nil
	}

	if err := q.refreshMetrics(ctx); err != nil {
		log.Printf("[lifecycle-queue] metrics refresh failed: %v", err)
	}

	if q.cfg.ReconcileOnStart {
		q.ReconcileOnce(ctx)
	}

	q.wg.Add(q.cfg.WorkerCount)
	for i := 0; i < q.cfg.WorkerCount; i++ {
		workerID := fmt.Sprintf("worker-%d", i+1)
		go func(id string) {
			defer q.wg.Done()
			q.workerLoop(ctx, id)
		}(workerID)
	}

	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		q.reconcileLoop(ctx)
	}()

	return nil
}

// Stop stops all workers and closes the store.
func (q *LifecycleCommandQueue) Stop() {
	close(q.stopCh)
	q.wg.Wait()
	_ = q.store.Close()
}

// EnqueueFromEvent enqueues a lifecycle action requested event.
func (q *LifecycleCommandQueue) EnqueueFromEvent(ctx context.Context, event marketplace.LifecycleActionRequestedEvent, resourceUUID string) (*LifecycleCommand, error) {
	cmd := buildLifecycleCommandFromEvent(event, resourceUUID, q.cfg.MaxRetries)
	stored, existing, err := q.store.Enqueue(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if existing {
		return stored, nil
	}

	q.metrics.QueueDepth.WithLabelValues(string(LifecycleCommandStatusPending)).Inc()
	q.metrics.CommandsTotal.WithLabelValues(string(cmd.Action), "enqueued").Inc()

	desired := &LifecycleDesiredState{
		AllocationID:  cmd.AllocationID,
		ResourceUUID:  cmd.ResourceUUID,
		DesiredState:  cmd.TargetState,
		LastAction:    cmd.Action,
		LastCommandID: cmd.ID,
		UpdatedAt:     time.Now().UTC(),
	}
	if err := q.store.SetDesiredState(ctx, desired); err != nil {
		log.Printf("[lifecycle-queue] failed to persist desired state for %s: %v", cmd.AllocationID, err)
	}

	return stored, nil
}

// ReconcileOnce runs reconciliation for stale commands and drift.
func (q *LifecycleCommandQueue) ReconcileOnce(ctx context.Context) {
	if !q.cfg.Enabled {
		return
	}
	q.metrics.ReconcileRuns.WithLabelValues("started").Inc()

	if err := q.requeueStaleExecuting(ctx); err != nil {
		log.Printf("[lifecycle-queue] stale command reconcile failed: %v", err)
		q.metrics.ReconcileRuns.WithLabelValues("error").Inc()
		return
	}

	desiredStates, err := q.store.ListDesiredStates(ctx)
	if err != nil {
		log.Printf("[lifecycle-queue] desired state list failed: %v", err)
		q.metrics.ReconcileRuns.WithLabelValues("error").Inc()
		return
	}

	drifted := 0
	for _, desired := range desiredStates {
		if desired == nil || desired.AllocationID == "" {
			continue
		}
		if err := q.reconcileDesiredState(ctx, desired); err != nil {
			log.Printf("[lifecycle-queue] reconcile allocation %s failed: %v", desired.AllocationID, err)
			q.metrics.ReconcileCommands.WithLabelValues("unknown", "error").Inc()
			continue
		}
		if desired.LastObserved != desired.DesiredState {
			drifted++
		}
	}

	if drifted > 0 {
		q.metrics.ReconcileRuns.WithLabelValues("drift_detected").Inc()
	} else {
		q.metrics.ReconcileRuns.WithLabelValues("matched").Inc()
	}
}

func (q *LifecycleCommandQueue) reconcileLoop(ctx context.Context) {
	ticker := time.NewTicker(q.cfg.ReconcileInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-q.stopCh:
			return
		case <-ticker.C:
			q.ReconcileOnce(ctx)
		}
	}
}

func (q *LifecycleCommandQueue) workerLoop(ctx context.Context, workerID string) {
	ticker := time.NewTicker(q.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-q.stopCh:
			return
		case <-ticker.C:
			cmd, err := q.store.ClaimNextReady(ctx, time.Now().UTC(), workerID)
			if err != nil {
				log.Printf("[lifecycle-queue] claim failed: %v", err)
				continue
			}
			if cmd == nil {
				continue
			}
			q.metrics.QueueDepth.WithLabelValues(string(LifecycleCommandStatusPending)).Dec()
			q.metrics.QueueDepth.WithLabelValues(string(LifecycleCommandStatusExecuting)).Inc()

			q.executeCommand(ctx, cmd)

			q.metrics.QueueDepth.WithLabelValues(string(LifecycleCommandStatusExecuting)).Dec()
		}
	}
}

func (q *LifecycleCommandQueue) executeCommand(ctx context.Context, cmd *LifecycleCommand) {
	if cmd == nil {
		return
	}

	if cmd.ResourceUUID == "" && q.resolver != nil {
		resourceUUID, err := q.resolver(ctx, cmd.AllocationID)
		if err != nil {
			q.handleCommandFailure(ctx, cmd, err)
			return
		}
		cmd.ResourceUUID = resourceUUID
		_ = q.store.Update(ctx, cmd)
	}

	execCtx, cancel := context.WithTimeout(ctx, q.cfg.ExecuteTimeout)
	defer cancel()

	result, err := q.executor.Execute(execCtx, cmd)
	if err != nil {
		q.handleCommandFailure(ctx, cmd, err)
		return
	}

	if result != nil {
		cmd.WaldurOpID = result.WaldurOperationID
	}

	now := time.Now().UTC()
	cmd.Status = LifecycleCommandStatusSucceeded
	cmd.CompletedAt = &now
	cmd.UpdatedAt = now
	cmd.LastError = ""

	if err := q.store.Update(ctx, cmd); err != nil {
		log.Printf("[lifecycle-queue] update success failed: %v", err)
	}

	q.metrics.CommandsTotal.WithLabelValues(string(cmd.Action), "succeeded").Inc()
	q.metrics.QueueDepth.WithLabelValues(string(LifecycleCommandStatusSucceeded)).Inc()

	if q.callback != nil && q.cfg.CallbackOnSuccess {
		if err := q.callback(ctx, cmd, nil); err != nil {
			log.Printf("[lifecycle-queue] callback failed for %s: %v", cmd.ID, err)
		}
	}
}

func (q *LifecycleCommandQueue) handleCommandFailure(ctx context.Context, cmd *LifecycleCommand, err error) {
	if cmd == nil {
		return
	}

	now := time.Now().UTC()
	cmd.LastError = err.Error()
	cmd.UpdatedAt = now

	if cmd.AttemptCount >= cmd.MaxAttempts {
		cmd.Status = LifecycleCommandStatusDeadLettered
		cmd.CompletedAt = &now
		if err := q.store.Update(ctx, cmd); err != nil {
			log.Printf("[lifecycle-queue] dead-letter update failed: %v", err)
		}
		q.metrics.CommandsTotal.WithLabelValues(string(cmd.Action), "dead_lettered").Inc()
		q.metrics.QueueDepth.WithLabelValues(string(LifecycleCommandStatusDeadLettered)).Inc()
		if q.callback != nil && q.cfg.CallbackOnFailure {
			if cbErr := q.callback(ctx, cmd, err); cbErr != nil {
				log.Printf("[lifecycle-queue] failure callback failed for %s: %v", cmd.ID, cbErr)
			}
		}
		return
	}

	delay := retryDelay(cmd.AttemptCount, q.cfg.RetryBackoff, q.cfg.MaxBackoff)
	nextAttempt := now.Add(delay)
	cmd.Status = LifecycleCommandStatusPending
	cmd.NextAttemptAt = &nextAttempt

	if updateErr := q.store.Update(ctx, cmd); updateErr != nil {
		log.Printf("[lifecycle-queue] retry update failed: %v", updateErr)
	}

	q.metrics.RetriesTotal.WithLabelValues(string(cmd.Action)).Inc()
	q.metrics.QueueDepth.WithLabelValues(string(LifecycleCommandStatusPending)).Inc()
}

func (q *LifecycleCommandQueue) refreshMetrics(ctx context.Context) error {
	commands, err := q.store.ListByStatus(ctx,
		LifecycleCommandStatusPending,
		LifecycleCommandStatusExecuting,
		LifecycleCommandStatusSucceeded,
		LifecycleCommandStatusFailed,
		LifecycleCommandStatusDeadLettered,
	)
	if err != nil {
		return err
	}

	counts := map[LifecycleCommandStatus]int{}
	for _, cmd := range commands {
		if cmd == nil {
			continue
		}
		counts[cmd.Status]++
	}

	for _, status := range []LifecycleCommandStatus{
		LifecycleCommandStatusPending,
		LifecycleCommandStatusExecuting,
		LifecycleCommandStatusSucceeded,
		LifecycleCommandStatusFailed,
		LifecycleCommandStatusDeadLettered,
	} {
		q.metrics.QueueDepth.WithLabelValues(string(status)).Set(0)
	}

	for status, count := range counts {
		q.metrics.QueueDepth.WithLabelValues(string(status)).Set(float64(count))
	}

	return nil
}

func (q *LifecycleCommandQueue) requeueStaleExecuting(ctx context.Context) error {
	commands, err := q.store.ListByStatus(ctx, LifecycleCommandStatusExecuting)
	if err != nil {
		return err
	}

	cutoff := time.Now().UTC().Add(-q.cfg.StaleAfter)
	for _, cmd := range commands {
		if cmd == nil || cmd.LastAttemptAt == nil {
			continue
		}
		if cmd.LastAttemptAt.After(cutoff) {
			continue
		}
		cmd.Status = LifecycleCommandStatusPending
		now := time.Now().UTC()
		cmd.NextAttemptAt = &now
		cmd.UpdatedAt = now
		if err := q.store.Update(ctx, cmd); err != nil {
			log.Printf("[lifecycle-queue] failed to requeue stale command %s: %v", cmd.ID, err)
			continue
		}
		q.metrics.RetriesTotal.WithLabelValues(string(cmd.Action)).Inc()
		q.metrics.ReconcileCommands.WithLabelValues(string(cmd.Action), "stale_requeued").Inc()
	}

	return nil
}

func (q *LifecycleCommandQueue) reconcileDesiredState(ctx context.Context, desired *LifecycleDesiredState) error {
	resourceUUID := desired.ResourceUUID
	if resourceUUID == "" && q.resolver != nil {
		resolved, err := q.resolver(ctx, desired.AllocationID)
		if err != nil {
			return err
		}
		resourceUUID = resolved
		desired.ResourceUUID = resourceUUID
	}

	if resourceUUID == "" {
		return fmt.Errorf("resource UUID unavailable for %s", desired.AllocationID)
	}

	state, err := q.executor.GetResourceState(ctx, resourceUUID)
	if err != nil {
		return err
	}

	observed := mapWaldurStateToAllocationState(string(state))
	now := time.Now().UTC()
	desired.LastObserved = observed
	desired.LastObservation = &now
	desired.LastReconciled = &now
	if err := q.store.SetDesiredState(ctx, desired); err != nil {
		log.Printf("[lifecycle-queue] desired state update failed for %s: %v", desired.AllocationID, err)
	}

	if observed == desired.DesiredState {
		q.metrics.ReconcileCommands.WithLabelValues("none", "matched").Inc()
		return nil
	}

	inFlight, err := q.store.ListByAllocation(ctx, desired.AllocationID, LifecycleCommandStatusPending, LifecycleCommandStatusExecuting)
	if err != nil {
		return err
	}
	if len(inFlight) > 0 {
		q.metrics.ReconcileCommands.WithLabelValues("none", "in_flight").Inc()
		return nil
	}

	action, ok := reconcileActionForDesired(desired.DesiredState, state)
	if !ok {
		q.metrics.ReconcileCommands.WithLabelValues("none", "unsupported").Inc()
		return nil
	}

	cmd := &LifecycleCommand{
		ID:             fmt.Sprintf("reconcile-%s-%d", desired.AllocationID, time.Now().UnixNano()),
		IdempotencyKey: marketplace.GenerateIdempotencyKey(desired.AllocationID, action, time.Now().UTC()),
		AllocationID:   desired.AllocationID,
		Provider:       "reconciler",
		ResourceUUID:   resourceUUID,
		Action:         action,
		TargetState:    desired.DesiredState,
		RequestedBy:    "reconciler",
		Parameters: map[string]interface{}{
			"reconcile":      "true",
			"expected_state": desired.DesiredState.String(),
			"observed_state": observed.String(),
		},
		Status:         LifecycleCommandStatusPending,
		AttemptCount:   0,
		MaxAttempts:    q.cfg.MaxRetries,
		CreatedAt:      now,
		UpdatedAt:      now,
		EventTimestamp: now,
		Reconcile:      true,
	}

	_, _, err = q.store.Enqueue(ctx, cmd)
	if err != nil {
		q.metrics.ReconcileCommands.WithLabelValues(string(action), "enqueue_failed").Inc()
		return err
	}

	q.metrics.QueueDepth.WithLabelValues(string(LifecycleCommandStatusPending)).Inc()
	q.metrics.ReconcileCommands.WithLabelValues(string(action), "issued").Inc()
	return nil
}

func retryDelay(attempt int, base, max time.Duration) time.Duration {
	if attempt <= 0 {
		return base
	}
	delay := base
	for i := 1; i < attempt; i++ {
		delay *= 2
		if delay >= max {
			return max
		}
	}
	if delay > max {
		return max
	}
	return delay
}

func buildLifecycleCommandFromEvent(event marketplace.LifecycleActionRequestedEvent, resourceUUID string, maxRetries int) *LifecycleCommand {
	now := time.Now().UTC()
	cmdID := event.OperationID
	if cmdID == "" {
		cmdID = event.EventID
	}
	if cmdID == "" {
		cmdID = fmt.Sprintf("cmd-%s-%d", event.AllocationID, time.Now().UnixNano())
	}

	idempotencyKey := event.OperationID
	if idempotencyKey == "" {
		idempotencyKey = event.EventID
	}
	if idempotencyKey == "" {
		idempotencyKey = marketplace.GenerateIdempotencyKey(event.AllocationID, event.Action, event.Timestamp)
	}

	if maxRetries <= 0 {
		maxRetries = 3
	}

	return &LifecycleCommand{
		ID:             cmdID,
		IdempotencyKey: idempotencyKey,
		AllocationID:   event.AllocationID,
		OrderID:        event.OrderID,
		OperationID:    event.OperationID,
		Provider:       event.ProviderAddress,
		ResourceUUID:   resourceUUID,
		Action:         event.Action,
		TargetState:    event.TargetState,
		RequestedBy:    event.RequestedBy,
		Parameters:     event.Parameters,
		RollbackPolicy: event.RollbackPolicy,
		Status:         LifecycleCommandStatusPending,
		AttemptCount:   0,
		MaxAttempts:    maxRetries,
		CreatedAt:      now,
		UpdatedAt:      now,
		NextAttemptAt:  &now,
		EventID:        event.EventID,
		BlockHeight:    event.BlockHeight,
		Sequence:       event.Sequence,
		EventTimestamp: event.Timestamp,
	}
}

func reconcileActionForDesired(desired marketplace.AllocationState, waldurState waldur.ResourceState) (marketplace.LifecycleActionType, bool) {
	switch desired {
	case marketplace.AllocationStateActive:
		switch waldurState {
		case waldur.ResourceStateStopped:
			return marketplace.LifecycleActionStart, true
		case waldur.ResourceStatePaused:
			return marketplace.LifecycleActionResume, true
		default:
			return "", false
		}
	case marketplace.AllocationStateSuspended:
		if waldurState == waldur.ResourceStateOK {
			return marketplace.LifecycleActionStop, true
		}
	case marketplace.AllocationStateTerminating, marketplace.AllocationStateTerminated:
		if waldurState != waldur.ResourceStateTerminated && waldurState != waldur.ResourceStateTerminating {
			return marketplace.LifecycleActionTerminate, true
		}
	}
	return "", false
}
