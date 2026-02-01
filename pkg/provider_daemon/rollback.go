// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-14E: Resource lifecycle control via Waldur
// This file implements rollback mechanisms for failed lifecycle operations.
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

// Rollback errors
var (
	// ErrRollbackNotSupported is returned when rollback is not supported
	ErrRollbackNotSupported = errors.New("rollback not supported for this action")

	// ErrRollbackFailed is returned when rollback fails
	ErrRollbackFailed = errors.New("rollback operation failed")

	// ErrRollbackAlreadyInProgress is returned when rollback is already in progress
	ErrRollbackAlreadyInProgress = errors.New("rollback already in progress")

	// ErrRollbackNotRequired is returned when rollback is not required
	ErrRollbackNotRequired = errors.New("rollback not required")

	// ErrRollbackTimeout is returned when rollback times out
	ErrRollbackTimeout = errors.New("rollback operation timed out")
)

// RollbackConfig configures the rollback manager
type RollbackConfig struct {
	// Enabled enables automatic rollback
	Enabled bool `json:"enabled"`

	// MaxRollbackAttempts is the maximum rollback attempts
	MaxRollbackAttempts int `json:"max_rollback_attempts"`

	// RollbackTimeout is the timeout for rollback operations
	RollbackTimeout time.Duration `json:"rollback_timeout"`

	// RetryInterval is the interval between rollback retries
	RetryInterval time.Duration `json:"retry_interval"`

	// EnableAuditLogging enables audit logging for rollback
	EnableAuditLogging bool `json:"enable_audit_logging"`

	// PreserveState preserves state on rollback failure
	PreserveState bool `json:"preserve_state"`
}

// DefaultRollbackConfig returns default configuration
func DefaultRollbackConfig() RollbackConfig {
	return RollbackConfig{
		Enabled:             true,
		MaxRollbackAttempts: 3,
		RollbackTimeout:     10 * time.Minute,
		RetryInterval:       30 * time.Second,
		EnableAuditLogging:  true,
		PreserveState:       true,
	}
}

// RollbackActionMap maps original actions to their rollback actions
var RollbackActionMap = map[marketplace.LifecycleActionType]marketplace.LifecycleActionType{
	marketplace.LifecycleActionStart:   marketplace.LifecycleActionStop,
	marketplace.LifecycleActionStop:    marketplace.LifecycleActionStart,
	marketplace.LifecycleActionSuspend: marketplace.LifecycleActionResume,
	marketplace.LifecycleActionResume:  marketplace.LifecycleActionSuspend,
	// Resize has special handling - restore previous spec
	// Terminate cannot be rolled back
	// Provision cannot be rolled back
}

// RollbackRecord tracks a rollback operation
type RollbackRecord struct {
	// ID is the unique rollback ID
	ID string `json:"id"`

	// OriginalOperationID is the original failed operation
	OriginalOperationID string `json:"original_operation_id"`

	// RollbackOperationID is the rollback operation ID
	RollbackOperationID string `json:"rollback_operation_id"`

	// AllocationID is the allocation being rolled back
	AllocationID string `json:"allocation_id"`

	// OriginalAction is the original action that failed
	OriginalAction marketplace.LifecycleActionType `json:"original_action"`

	// RollbackAction is the rollback action
	RollbackAction marketplace.LifecycleActionType `json:"rollback_action"`

	// State is the current rollback state
	State RollbackState `json:"state"`

	// AttemptCount is the number of rollback attempts
	AttemptCount int `json:"attempt_count"`

	// OriginalState is the state before the original operation
	OriginalState marketplace.AllocationState `json:"original_state"`

	// RestoredState is the state after rollback
	RestoredState marketplace.AllocationState `json:"restored_state,omitempty"`

	// OriginalResizeSpec is the original resize spec (for resize rollback)
	OriginalResizeSpec *marketplace.ResizeSpecification `json:"original_resize_spec,omitempty"`

	// Error contains any error message
	Error string `json:"error,omitempty"`

	// CreatedAt is when the rollback was initiated
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the rollback was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// CompletedAt is when the rollback completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// RollbackState represents the state of a rollback operation
type RollbackState string

const (
	// RollbackStatePending rollback is pending
	RollbackStatePending RollbackState = "pending"

	// RollbackStateExecuting rollback is executing
	RollbackStateExecuting RollbackState = "executing"

	// RollbackStateCompleted rollback completed successfully
	RollbackStateCompleted RollbackState = "completed"

	// RollbackStateFailed rollback failed
	RollbackStateFailed RollbackState = "failed"

	// RollbackStateSkipped rollback was skipped
	RollbackStateSkipped RollbackState = "skipped"
)

// IsTerminal returns true if the state is terminal
func (s RollbackState) IsTerminal() bool {
	return s == RollbackStateCompleted || s == RollbackStateFailed || s == RollbackStateSkipped
}

// RollbackManager manages rollback operations
type RollbackManager struct {
	cfg         RollbackConfig
	controller  *LifecycleController
	lifecycle   *waldur.LifecycleClient
	auditLogger *AuditLogger
	records     map[string]*RollbackRecord
	activeRollbacks map[string]string // allocationID -> rollbackID
	mu          sync.RWMutex
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// NewRollbackManager creates a new rollback manager
func NewRollbackManager(
	cfg RollbackConfig,
	controller *LifecycleController,
	lifecycle *waldur.LifecycleClient,
	auditLogger *AuditLogger,
) *RollbackManager {
	return &RollbackManager{
		cfg:             cfg,
		controller:      controller,
		lifecycle:       lifecycle,
		auditLogger:     auditLogger,
		records:         make(map[string]*RollbackRecord),
		activeRollbacks: make(map[string]string),
		stopCh:          make(chan struct{}),
	}
}

// Start starts the rollback manager
func (m *RollbackManager) Start(ctx context.Context) error {
	if !m.cfg.Enabled {
		return nil
	}

	log.Printf("[rollback-manager] starting rollback manager")

	// Start retry worker
	m.wg.Add(1)
	go m.retryWorker(ctx)

	return nil
}

// Stop stops the rollback manager
func (m *RollbackManager) Stop() error {
	close(m.stopCh)
	m.wg.Wait()
	return nil
}

// RollbackOperation initiates a rollback for a failed operation
func (m *RollbackManager) RollbackOperation(ctx context.Context, op *marketplace.LifecycleOperation) error {
	if !m.cfg.Enabled {
		return ErrRollbackNotSupported
	}

	// Check if rollback is supported for this action
	rollbackAction, supported := RollbackActionMap[op.Action]
	if !supported {
		return fmt.Errorf("%w: %s", ErrRollbackNotSupported, op.Action)
	}

	// Check if rollback is already in progress
	m.mu.Lock()
	if existingID, exists := m.activeRollbacks[op.AllocationID]; exists {
		m.mu.Unlock()
		return fmt.Errorf("%w: existing rollback %s", ErrRollbackAlreadyInProgress, existingID)
	}

	// Create rollback record
	now := time.Now().UTC()
	record := &RollbackRecord{
		ID:                  fmt.Sprintf("rb_%s_%d", op.ID, now.UnixNano()),
		OriginalOperationID: op.ID,
		AllocationID:        op.AllocationID,
		OriginalAction:      op.Action,
		RollbackAction:      rollbackAction,
		State:               RollbackStatePending,
		AttemptCount:        0,
		OriginalState:       op.PreviousAllocationState,
		OriginalResizeSpec:  op.ResizeSpec,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	m.records[record.ID] = record
	m.activeRollbacks[op.AllocationID] = record.ID
	m.mu.Unlock()

	// Execute rollback
	go m.executeRollback(ctx, record, op)

	log.Printf("[rollback-manager] initiated rollback %s for operation %s (%s -> %s)",
		record.ID, op.ID, op.Action, rollbackAction)

	return nil
}

// executeRollback executes a rollback operation
func (m *RollbackManager) executeRollback(ctx context.Context, record *RollbackRecord, op *marketplace.LifecycleOperation) {
	defer func() {
		m.mu.Lock()
		delete(m.activeRollbacks, record.AllocationID)
		m.mu.Unlock()
	}()

	// Create timeout context
	rollbackCtx, cancel := context.WithTimeout(ctx, m.cfg.RollbackTimeout)
	defer cancel()

	// Update state to executing
	m.updateRecordState(record, RollbackStateExecuting, "")

	// Build rollback parameters
	params := make(map[string]string)
	params["rollback_for"] = op.ID
	params["original_action"] = string(op.Action)

	// Handle resize rollback specially
	if op.Action == marketplace.LifecycleActionResize && op.ResizeSpec != nil {
		// Would restore original resize spec
		params["restore_spec"] = "true"
	}

	// Get resource UUID from operation
	waldurResourceUUID := op.WaldurOperationID
	if waldurResourceUUID == "" {
		// Try to look up from allocation
		waldurResourceUUID = m.lookupWaldurResourceUUID(op.AllocationID)
	}

	if waldurResourceUUID == "" {
		m.updateRecordState(record, RollbackStateFailed, "cannot determine Waldur resource UUID")
		m.logRollbackEvent(record, false)
		return
	}

	// Execute the rollback action via controller
	rollbackOp, err := m.controller.ExecuteLifecycleAction(
		rollbackCtx,
		op.AllocationID,
		record.RollbackAction,
		op.TargetAllocationState, // Current expected state
		waldurResourceUUID,
		"rollback-manager",
		params,
	)

	if err != nil {
		m.handleRollbackFailure(ctx, record, err)
		return
	}

	record.RollbackOperationID = rollbackOp.ID
	record.AttemptCount++
	m.updateRecordState(record, RollbackStateExecuting, "")

	// Wait for rollback to complete
	m.waitForRollbackCompletion(rollbackCtx, record)
}

// waitForRollbackCompletion waits for a rollback operation to complete
func (m *RollbackManager) waitForRollbackCompletion(ctx context.Context, record *RollbackRecord) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.updateRecordState(record, RollbackStateFailed, ErrRollbackTimeout.Error())
			m.logRollbackEvent(record, false)
			return

		case <-m.stopCh:
			return

		case <-ticker.C:
			op, found := m.controller.GetOperation(record.RollbackOperationID)
			if !found {
				continue
			}

			if op.State.IsTerminal() {
				if op.State == marketplace.LifecycleOpStateCompleted {
					record.RestoredState = record.OriginalState
					m.updateRecordState(record, RollbackStateCompleted, "")
					m.logRollbackEvent(record, true)
					log.Printf("[rollback-manager] rollback %s completed successfully", record.ID)
				} else {
					m.handleRollbackFailure(context.Background(), record, errors.New(op.Error))
				}
				return
			}
		}
	}
}

// handleRollbackFailure handles a failed rollback attempt
func (m *RollbackManager) handleRollbackFailure(ctx context.Context, record *RollbackRecord, err error) {
	record.AttemptCount++
	errMsg := err.Error()

	if record.AttemptCount < m.cfg.MaxRollbackAttempts {
		// Schedule retry
		m.updateRecordState(record, RollbackStatePending, errMsg)
		log.Printf("[rollback-manager] rollback %s failed (attempt %d/%d): %v",
			record.ID, record.AttemptCount, m.cfg.MaxRollbackAttempts, err)
	} else {
		// Max attempts reached
		m.updateRecordState(record, RollbackStateFailed, errMsg)
		m.logRollbackEvent(record, false)
		log.Printf("[rollback-manager] rollback %s failed permanently after %d attempts: %v",
			record.ID, record.AttemptCount, err)
	}
}

// updateRecordState updates a rollback record's state
func (m *RollbackManager) updateRecordState(record *RollbackRecord, state RollbackState, errMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record.State = state
	record.Error = errMsg
	record.UpdatedAt = time.Now().UTC()

	if state.IsTerminal() {
		completedAt := time.Now().UTC()
		record.CompletedAt = &completedAt
	}
}

// lookupWaldurResourceUUID looks up Waldur resource UUID for an allocation
func (m *RollbackManager) lookupWaldurResourceUUID(allocationID string) string {
	// In a real implementation, this would look up the resource mapping
	// For now, return empty and let the caller handle it
	return ""
}

// retryWorker periodically retries pending rollbacks
func (m *RollbackManager) retryWorker(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(m.cfg.RetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.processPendingRollbacks(ctx)
		}
	}
}

// processPendingRollbacks processes pending rollback retries
func (m *RollbackManager) processPendingRollbacks(ctx context.Context) {
	m.mu.RLock()
	var pending []*RollbackRecord
	for _, record := range m.records {
		if record.State == RollbackStatePending && record.AttemptCount > 0 {
			pending = append(pending, record)
		}
	}
	m.mu.RUnlock()

	for _, record := range pending {
		// Re-execute rollback
		op, found := m.controller.GetOperation(record.OriginalOperationID)
		if !found {
			m.updateRecordState(record, RollbackStateFailed, "original operation not found")
			continue
		}

		go m.executeRollback(ctx, record, op)
	}
}

// GetRollbackRecord retrieves a rollback record
func (m *RollbackManager) GetRollbackRecord(id string) (*RollbackRecord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	record, ok := m.records[id]
	return record, ok
}

// GetRollbacksForAllocation retrieves rollback records for an allocation
func (m *RollbackManager) GetRollbacksForAllocation(allocationID string) []*RollbackRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*RollbackRecord
	for _, record := range m.records {
		if record.AllocationID == allocationID {
			result = append(result, record)
		}
	}
	return result
}

// GetActiveRollback retrieves the active rollback for an allocation
func (m *RollbackManager) GetActiveRollback(allocationID string) (*RollbackRecord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rbID, exists := m.activeRollbacks[allocationID]
	if !exists {
		return nil, false
	}

	record, ok := m.records[rbID]
	return record, ok
}

// CancelRollback cancels a pending rollback
func (m *RollbackManager) CancelRollback(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, ok := m.records[id]
	if !ok {
		return errors.New("rollback not found")
	}

	if record.State.IsTerminal() {
		return errors.New("rollback already completed")
	}

	record.State = RollbackStateSkipped
	record.Error = "cancelled by user"
	completedAt := time.Now().UTC()
	record.CompletedAt = &completedAt
	record.UpdatedAt = completedAt

	delete(m.activeRollbacks, record.AllocationID)

	log.Printf("[rollback-manager] rollback %s cancelled", id)

	return nil
}

// GetRollbackMetrics returns rollback metrics
func (m *RollbackManager) GetRollbackMetrics() *RollbackMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := &RollbackMetrics{
		TotalRollbacks: int64(len(m.records)),
		ByState:        make(map[RollbackState]int64),
		ByAction:       make(map[marketplace.LifecycleActionType]int64),
	}

	for _, record := range m.records {
		metrics.ByState[record.State]++
		metrics.ByAction[record.OriginalAction]++

		if record.State == RollbackStateCompleted {
			metrics.SuccessfulRollbacks++
		} else if record.State == RollbackStateFailed {
			metrics.FailedRollbacks++
		}
	}

	metrics.ActiveRollbacks = int64(len(m.activeRollbacks))
	metrics.LastUpdated = time.Now().UTC()

	return metrics
}

// RollbackMetrics contains rollback statistics
type RollbackMetrics struct {
	// TotalRollbacks is the total number of rollbacks
	TotalRollbacks int64 `json:"total_rollbacks"`

	// SuccessfulRollbacks is the number of successful rollbacks
	SuccessfulRollbacks int64 `json:"successful_rollbacks"`

	// FailedRollbacks is the number of failed rollbacks
	FailedRollbacks int64 `json:"failed_rollbacks"`

	// ActiveRollbacks is the number of active rollbacks
	ActiveRollbacks int64 `json:"active_rollbacks"`

	// ByState contains counts by state
	ByState map[RollbackState]int64 `json:"by_state"`

	// ByAction contains counts by original action
	ByAction map[marketplace.LifecycleActionType]int64 `json:"by_action"`

	// LastUpdated is when metrics were last updated
	LastUpdated time.Time `json:"last_updated"`
}

// logRollbackEvent logs a rollback audit event
func (m *RollbackManager) logRollbackEvent(record *RollbackRecord, success bool) {
	if m.auditLogger == nil || !m.cfg.EnableAuditLogging {
		return
	}

	eventType := AuditEventType("rollback_completed")
	if !success {
		eventType = AuditEventType("rollback_failed")
	}

	details := map[string]interface{}{
		"rollback_id":          record.ID,
		"original_operation_id": record.OriginalOperationID,
		"allocation_id":        record.AllocationID,
		"original_action":      record.OriginalAction,
		"rollback_action":      record.RollbackAction,
		"attempt_count":        record.AttemptCount,
		"original_state":       record.OriginalState,
	}

	if success {
		details["restored_state"] = record.RestoredState
	}

	_ = m.auditLogger.Log(&AuditEvent{
		Type:         eventType,
		Operation:    "rollback",
		Success:      success,
		ErrorMessage: record.Error,
		Details:      details,
	})
}

// CleanupOldRecords removes old completed rollback records
func (m *RollbackManager) CleanupOldRecords(retentionDays int) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	count := 0

	for id, record := range m.records {
		if record.State.IsTerminal() && record.CompletedAt != nil && record.CompletedAt.Before(cutoff) {
			delete(m.records, id)
			count++
		}
	}

	if count > 0 {
		log.Printf("[rollback-manager] cleaned up %d old rollback records", count)
	}

	return count
}

// CanRollback checks if an operation can be rolled back
func CanRollback(action marketplace.LifecycleActionType) bool {
	_, supported := RollbackActionMap[action]
	return supported
}

// GetRollbackAction returns the rollback action for an action
func GetRollbackAction(action marketplace.LifecycleActionType) (marketplace.LifecycleActionType, bool) {
	rollback, ok := RollbackActionMap[action]
	return rollback, ok
}

// RollbackPlanGenerator generates rollback plans
type RollbackPlanGenerator struct {
	// CustomHandlers contains custom rollback handlers by action
	CustomHandlers map[marketplace.LifecycleActionType]RollbackHandler
}

// RollbackHandler is a custom rollback handler function
type RollbackHandler func(ctx context.Context, op *marketplace.LifecycleOperation) error

// NewRollbackPlanGenerator creates a new rollback plan generator
func NewRollbackPlanGenerator() *RollbackPlanGenerator {
	return &RollbackPlanGenerator{
		CustomHandlers: make(map[marketplace.LifecycleActionType]RollbackHandler),
	}
}

// RegisterHandler registers a custom rollback handler
func (g *RollbackPlanGenerator) RegisterHandler(action marketplace.LifecycleActionType, handler RollbackHandler) {
	g.CustomHandlers[action] = handler
}

// GeneratePlan generates a rollback plan for an operation
func (g *RollbackPlanGenerator) GeneratePlan(op *marketplace.LifecycleOperation) (*RollbackPlan, error) {
	rollbackAction, supported := RollbackActionMap[op.Action]
	if !supported {
		return nil, fmt.Errorf("%w: %s", ErrRollbackNotSupported, op.Action)
	}

	plan := &RollbackPlan{
		OriginalOperation: op,
		RollbackAction:    rollbackAction,
		TargetState:       op.PreviousAllocationState,
		Steps:             make([]RollbackStep, 0),
	}

	// Add pre-rollback steps
	plan.Steps = append(plan.Steps, RollbackStep{
		Name:        "validate_state",
		Description: "Validate current resource state",
		Required:    true,
	})

	// Add main rollback step
	plan.Steps = append(plan.Steps, RollbackStep{
		Name:        "execute_rollback",
		Description: fmt.Sprintf("Execute %s action", rollbackAction),
		Required:    true,
	})

	// Add post-rollback steps
	plan.Steps = append(plan.Steps, RollbackStep{
		Name:        "verify_state",
		Description: "Verify resource returned to original state",
		Required:    true,
	})

	return plan, nil
}

// RollbackPlan describes a rollback operation plan
type RollbackPlan struct {
	// OriginalOperation is the original failed operation
	OriginalOperation *marketplace.LifecycleOperation

	// RollbackAction is the action to perform for rollback
	RollbackAction marketplace.LifecycleActionType

	// TargetState is the target state after rollback
	TargetState marketplace.AllocationState

	// Steps are the rollback steps to execute
	Steps []RollbackStep
}

// RollbackStep is a step in a rollback plan
type RollbackStep struct {
	// Name is the step name
	Name string

	// Description is the step description
	Description string

	// Required indicates if the step is required
	Required bool

	// Handler is an optional custom handler
	Handler RollbackHandler
}

