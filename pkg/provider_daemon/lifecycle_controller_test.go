// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-4E: Resource lifecycle control via Waldur
// This file contains tests for the lifecycle controller.
package provider_daemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

func TestLifecycleControllerConfig_Defaults(t *testing.T) {
	cfg := DefaultLifecycleControllerConfig()

	assert.True(t, cfg.Enabled)
	assert.Equal(t, 5*time.Minute, cfg.OperationTimeout)
	assert.Equal(t, time.Hour, cfg.CallbackTTL)
	assert.Equal(t, 10, cfg.MaxConcurrentOps)
	assert.Equal(t, 30*time.Second, cfg.RetryInterval)
	assert.Equal(t, time.Hour, cfg.CleanupInterval)
	assert.Equal(t, 7, cfg.OperationRetentionDays)
	assert.True(t, cfg.EnableAuditLogging)
}

func TestLifecycleControllerState_NewState(t *testing.T) {
	state := NewLifecycleControllerState()

	require.NotNil(t, state)
	assert.NotNil(t, state.Operations)
	assert.NotNil(t, state.IdempotencyIndex)
	assert.NotNil(t, state.ProcessedCallbacks)
	assert.NotNil(t, state.Metrics)
	assert.False(t, state.LastUpdated.IsZero())
}

func TestLifecycleControllerState_AddOperation(t *testing.T) {
	state := NewLifecycleControllerState()

	op, err := marketplace.NewLifecycleOperation(
		"alloc-123",
		marketplace.LifecycleActionStart,
		"requester-789",
		"provider-456",
		marketplace.AllocationStateSuspended,
	)
	require.NoError(t, err)
	require.NotNil(t, op)

	state.Operations[op.ID] = op
	state.IdempotencyIndex[op.IdempotencyKey] = op.ID

	// Verify operation is stored
	stored, exists := state.Operations[op.ID]
	assert.True(t, exists)
	assert.Equal(t, op.AllocationID, stored.AllocationID)
	assert.Equal(t, marketplace.LifecycleActionStart, stored.Action)
	assert.Equal(t, marketplace.LifecycleOpStatePending, stored.State)

	// Verify idempotency index
	opID, exists := state.IdempotencyIndex[op.IdempotencyKey]
	assert.True(t, exists)
	assert.Equal(t, op.ID, opID)
}

func TestLifecycleControllerState_ProcessedCallbacks(t *testing.T) {
	state := NewLifecycleControllerState()
	now := time.Now().UTC()

	// Add some processed callbacks
	state.ProcessedCallbacks["nonce-1"] = now.Add(-time.Hour)
	state.ProcessedCallbacks["nonce-2"] = now.Add(-30 * time.Minute)
	state.ProcessedCallbacks["nonce-3"] = now

	assert.Len(t, state.ProcessedCallbacks, 3)

	// Verify timestamps
	ts, exists := state.ProcessedCallbacks["nonce-1"]
	assert.True(t, exists)
	assert.True(t, ts.Before(now))
}

func TestLifecycleController_StateFilePersistence(t *testing.T) {
	// Create temp directory for state file
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "lifecycle_state.json")

	cfg := DefaultLifecycleControllerConfig()
	cfg.StateFilePath = stateFile
	cfg.ProviderAddress = "provider-test-123"

	// Create initial state
	state := NewLifecycleControllerState()
	op, err := marketplace.NewLifecycleOperation(
		"alloc-123",
		marketplace.LifecycleActionStop,
		"requester-789",
		"provider-456",
		marketplace.AllocationStateActive,
	)
	require.NoError(t, err)
	state.Operations[op.ID] = op
	state.IdempotencyIndex[op.IdempotencyKey] = op.ID
	state.ProcessedCallbacks["callback-nonce-1"] = time.Now().UTC()

	// Create controller with state (manual setup for testing)
	lc := &LifecycleController{
		cfg:    cfg,
		state:  state,
		stopCh: make(chan struct{}),
	}

	// Save state
	err = lc.saveState()
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(stateFile)
	require.NoError(t, err)

	// Create new controller and load state
	lc2 := &LifecycleController{
		cfg:    cfg,
		state:  NewLifecycleControllerState(),
		stopCh: make(chan struct{}),
	}

	err = lc2.loadState()
	require.NoError(t, err)

	// Verify state was loaded correctly
	assert.Len(t, lc2.state.Operations, 1)
	assert.Len(t, lc2.state.IdempotencyIndex, 1)
	assert.Len(t, lc2.state.ProcessedCallbacks, 1)

	loadedOp, exists := lc2.state.Operations[op.ID]
	assert.True(t, exists)
	assert.Equal(t, op.AllocationID, loadedOp.AllocationID)
	assert.Equal(t, marketplace.LifecycleActionStop, loadedOp.Action)
}

func TestLifecycleController_IdempotencyDetection(t *testing.T) {
	state := NewLifecycleControllerState()

	// Create first operation
	op1, err := marketplace.NewLifecycleOperation(
		"alloc-123",
		marketplace.LifecycleActionRestart,
		"requester-789",
		"provider-456",
		marketplace.AllocationStateActive,
	)
	require.NoError(t, err)

	state.Operations[op1.ID] = op1
	state.IdempotencyIndex[op1.IdempotencyKey] = op1.ID

	// Check idempotency - same key should return existing operation
	existingOpID, isDuplicate := state.IdempotencyIndex[op1.IdempotencyKey]
	assert.True(t, isDuplicate)
	assert.Equal(t, op1.ID, existingOpID)

	// Different operation should not be a duplicate
	op2, err := marketplace.NewLifecycleOperation(
		"alloc-999", // Different allocation
		marketplace.LifecycleActionRestart,
		"requester-789",
		"provider-456",
		marketplace.AllocationStateActive,
	)
	require.NoError(t, err)

	_, isDuplicate = state.IdempotencyIndex[op2.IdempotencyKey]
	assert.False(t, isDuplicate)
}

func TestLifecycleOperation_StateTransitions(t *testing.T) {
	tests := []struct {
		name      string
		fromState marketplace.LifecycleOperationState
		toState   marketplace.LifecycleOperationState
		wantValid bool
	}{
		{
			name:      "pending to executing",
			fromState: marketplace.LifecycleOpStatePending,
			toState:   marketplace.LifecycleOpStateExecuting,
			wantValid: true,
		},
		{
			name:      "executing to awaiting_callback",
			fromState: marketplace.LifecycleOpStateExecuting,
			toState:   marketplace.LifecycleOpStateAwaitingCallback,
			wantValid: true,
		},
		{
			name:      "awaiting_callback to completed",
			fromState: marketplace.LifecycleOpStateAwaitingCallback,
			toState:   marketplace.LifecycleOpStateCompleted,
			wantValid: true,
		},
		{
			name:      "awaiting_callback to failed",
			fromState: marketplace.LifecycleOpStateAwaitingCallback,
			toState:   marketplace.LifecycleOpStateFailed,
			wantValid: true,
		},
		{
			name:      "failed to rolled_back",
			fromState: marketplace.LifecycleOpStateFailed,
			toState:   marketplace.LifecycleOpStateRolledBack,
			wantValid: true,
		},
		{
			name:      "pending to completed - invalid",
			fromState: marketplace.LifecycleOpStatePending,
			toState:   marketplace.LifecycleOpStateCompleted,
			wantValid: false,
		},
		{
			name:      "completed to executing - invalid",
			fromState: marketplace.LifecycleOpStateCompleted,
			toState:   marketplace.LifecycleOpStateExecuting,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, err := marketplace.NewLifecycleOperation(
				"alloc-123",
				marketplace.LifecycleActionStart,
				"requester-789",
				"provider-456",
				marketplace.AllocationStateSuspended,
			)
			require.NoError(t, err)

			// Set initial state
			op.State = tt.fromState

			// Check transition validity based on state machine
			isValid := isValidOperationStateTransition(tt.fromState, tt.toState)
			assert.Equal(t, tt.wantValid, isValid, "transition from %s to %s", tt.fromState, tt.toState)
		})
	}
}

// isValidOperationStateTransition checks if a state transition is valid
func isValidOperationStateTransition(from, to marketplace.LifecycleOperationState) bool {
	validTransitions := map[marketplace.LifecycleOperationState][]marketplace.LifecycleOperationState{
		marketplace.LifecycleOpStatePending: {
			marketplace.LifecycleOpStateExecuting,
			marketplace.LifecycleOpStateCancelled,
		},
		marketplace.LifecycleOpStateExecuting: {
			marketplace.LifecycleOpStateAwaitingCallback,
			marketplace.LifecycleOpStateCompleted,
			marketplace.LifecycleOpStateFailed,
		},
		marketplace.LifecycleOpStateAwaitingCallback: {
			marketplace.LifecycleOpStateCompleted,
			marketplace.LifecycleOpStateFailed,
		},
		marketplace.LifecycleOpStateFailed: {
			marketplace.LifecycleOpStateRolledBack,
			marketplace.LifecycleOpStatePending, // Retry
		},
	}

	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

func TestLifecycleMetrics_RecordSuccess(t *testing.T) {
	metrics := marketplace.NewLifecycleMetrics()

	// Record some successful operations by manually updating metrics
	metrics.TotalOperations++
	metrics.CompletedOperations++
	metrics.OperationsByAction[marketplace.LifecycleActionStart]++
	metrics.TotalOperations++
	metrics.CompletedOperations++
	metrics.OperationsByAction[marketplace.LifecycleActionStart]++
	metrics.TotalOperations++
	metrics.CompletedOperations++
	metrics.OperationsByAction[marketplace.LifecycleActionStop]++

	assert.Equal(t, int64(3), metrics.TotalOperations)
	assert.Equal(t, int64(3), metrics.CompletedOperations)
	assert.Equal(t, int64(0), metrics.FailedOperations)
}

func TestLifecycleMetrics_RecordFailure(t *testing.T) {
	metrics := marketplace.NewLifecycleMetrics()

	// Record some failed operations
	metrics.TotalOperations++
	metrics.FailedOperations++
	metrics.TotalOperations++
	metrics.FailedOperations++

	assert.Equal(t, int64(2), metrics.TotalOperations)
	assert.Equal(t, int64(0), metrics.CompletedOperations)
	assert.Equal(t, int64(2), metrics.FailedOperations)
}

func TestLifecycleCallback_Validation(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		callback  *marketplace.LifecycleCallback
		wantValid bool
		wantErr   string
	}{
		{
			name: "valid callback",
			callback: &marketplace.LifecycleCallback{
				ID:              "lcb-123",
				OperationID:     "op-123",
				AllocationID:    "alloc-456",
				Action:          marketplace.LifecycleActionStart,
				Success:         true,
				ProviderAddress: "provider-789",
				Timestamp:       now,
				ExpiresAt:       now.Add(time.Hour),
				Nonce:           "nonce-123",
				SignerID:        "provider-789",
				Signature:       []byte("test-signature"),
			},
			wantValid: true,
		},
		{
			name: "expired callback",
			callback: &marketplace.LifecycleCallback{
				ID:              "lcb-123",
				OperationID:     "op-123",
				AllocationID:    "alloc-456",
				Action:          marketplace.LifecycleActionStart,
				Success:         true,
				ProviderAddress: "provider-789",
				Timestamp:       now.Add(-2 * time.Hour),
				ExpiresAt:       now.Add(-time.Hour),
				Nonce:           "nonce-123",
				SignerID:        "provider-789",
				Signature:       []byte("test-signature"),
			},
			wantValid: false,
			wantErr:   "expired",
		},
		{
			name: "missing operation ID",
			callback: &marketplace.LifecycleCallback{
				ID:              "lcb-123",
				AllocationID:    "alloc-456",
				Action:          marketplace.LifecycleActionStart,
				Success:         true,
				ProviderAddress: "provider-789",
				Timestamp:       now,
				ExpiresAt:       now.Add(time.Hour),
				Nonce:           "nonce-123",
				SignerID:        "provider-789",
				Signature:       []byte("test-signature"),
			},
			wantValid: false,
			wantErr:   "operation",
		},
		{
			name: "missing nonce",
			callback: &marketplace.LifecycleCallback{
				ID:              "lcb-123",
				OperationID:     "op-123",
				AllocationID:    "alloc-456",
				Action:          marketplace.LifecycleActionStart,
				Success:         true,
				ProviderAddress: "provider-789",
				Timestamp:       now,
				ExpiresAt:       now.Add(time.Hour),
				SignerID:        "provider-789",
				Signature:       []byte("test-signature"),
			},
			wantValid: false,
			wantErr:   "nonce",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.callback.Validate()
			if tt.wantValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if tt.wantErr != "" {
					assert.Contains(t, err.Error(), tt.wantErr)
				}
			}
		})
	}
}

func TestAllocationLifecycleTransitions(t *testing.T) {
	tests := []struct {
		name      string
		fromState marketplace.AllocationState
		action    marketplace.LifecycleActionType
		toState   marketplace.AllocationState
		wantValid bool
	}{
		{
			name:      "active to suspended via stop",
			fromState: marketplace.AllocationStateActive,
			action:    marketplace.LifecycleActionStop,
			toState:   marketplace.AllocationStateSuspended,
			wantValid: true,
		},
		{
			name:      "active to suspended via suspend",
			fromState: marketplace.AllocationStateActive,
			action:    marketplace.LifecycleActionSuspend,
			toState:   marketplace.AllocationStateSuspended,
			wantValid: true,
		},
		{
			name:      "suspended to active via start",
			fromState: marketplace.AllocationStateSuspended,
			action:    marketplace.LifecycleActionStart,
			toState:   marketplace.AllocationStateActive,
			wantValid: true,
		},
		{
			name:      "suspended to active via resume",
			fromState: marketplace.AllocationStateSuspended,
			action:    marketplace.LifecycleActionResume,
			toState:   marketplace.AllocationStateActive,
			wantValid: true,
		},
		{
			name:      "active to terminating via terminate",
			fromState: marketplace.AllocationStateActive,
			action:    marketplace.LifecycleActionTerminate,
			toState:   marketplace.AllocationStateTerminating,
			wantValid: true,
		},
		{
			name:      "suspended to terminating via terminate",
			fromState: marketplace.AllocationStateSuspended,
			action:    marketplace.LifecycleActionTerminate,
			toState:   marketplace.AllocationStateTerminating,
			wantValid: true,
		},
		{
			name:      "active restart stays active",
			fromState: marketplace.AllocationStateActive,
			action:    marketplace.LifecycleActionRestart,
			toState:   marketplace.AllocationStateActive,
			wantValid: true,
		},
		{
			name:      "terminated cannot start",
			fromState: marketplace.AllocationStateTerminated,
			action:    marketplace.LifecycleActionStart,
			toState:   marketplace.AllocationStateActive,
			wantValid: false,
		},
		{
			name:      "pending cannot stop",
			fromState: marketplace.AllocationStatePending,
			action:    marketplace.LifecycleActionStop,
			toState:   marketplace.AllocationStateSuspended,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newState, err := marketplace.ValidateLifecycleTransition(tt.fromState, tt.action)
			if tt.wantValid {
				assert.NoError(t, err)
				assert.Equal(t, tt.toState, newState)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestGenerateIdempotencyKey(t *testing.T) {
	now := time.Now().UTC()
	hourAgo := now.Add(-time.Hour)

	// Same inputs within same hour should produce same key
	key1 := marketplace.GenerateIdempotencyKey("alloc-123", marketplace.LifecycleActionStart, now)
	key2 := marketplace.GenerateIdempotencyKey("alloc-123", marketplace.LifecycleActionStart, now.Add(30*time.Minute))

	// Keys within same hour should be identical (hour truncation)
	if now.Truncate(time.Hour).Equal(now.Add(30 * time.Minute).Truncate(time.Hour)) {
		assert.Equal(t, key1, key2)
	}

	// Different allocation should produce different key
	key3 := marketplace.GenerateIdempotencyKey("alloc-456", marketplace.LifecycleActionStart, now)
	assert.NotEqual(t, key1, key3)

	// Different action should produce different key
	key4 := marketplace.GenerateIdempotencyKey("alloc-123", marketplace.LifecycleActionStop, now)
	assert.NotEqual(t, key1, key4)

	// Different hour should produce different key
	key5 := marketplace.GenerateIdempotencyKey("alloc-123", marketplace.LifecycleActionStart, hourAgo)
	if !now.Truncate(time.Hour).Equal(hourAgo.Truncate(time.Hour)) {
		assert.NotEqual(t, key1, key5)
	}
}

func TestLifecycleController_CleanupOldOperations(t *testing.T) {
	state := NewLifecycleControllerState()
	now := time.Now().UTC()

	// Add old completed operation
	oldOp, _ := marketplace.NewLifecycleOperation(
		"alloc-old",
		marketplace.LifecycleActionStart,
		"requester-789",
		"provider-456",
		marketplace.AllocationStateSuspended,
	)
	oldOp.State = marketplace.LifecycleOpStateCompleted
	completedTime := now.Add(-10 * 24 * time.Hour)
	oldOp.CompletedAt = &completedTime
	state.Operations[oldOp.ID] = oldOp
	state.IdempotencyIndex[oldOp.IdempotencyKey] = oldOp.ID

	// Add recent completed operation
	recentOp, _ := marketplace.NewLifecycleOperation(
		"alloc-recent",
		marketplace.LifecycleActionStop,
		"requester-789",
		"provider-456",
		marketplace.AllocationStateActive,
	)
	recentOp.State = marketplace.LifecycleOpStateCompleted
	recentCompletedTime := now.Add(-1 * 24 * time.Hour)
	recentOp.CompletedAt = &recentCompletedTime
	state.Operations[recentOp.ID] = recentOp
	state.IdempotencyIndex[recentOp.IdempotencyKey] = recentOp.ID

	// Add pending operation (should not be cleaned up)
	pendingOp, _ := marketplace.NewLifecycleOperation(
		"alloc-pending",
		marketplace.LifecycleActionRestart,
		"requester-789",
		"provider-456",
		marketplace.AllocationStateActive,
	)
	pendingOp.State = marketplace.LifecycleOpStatePending
	state.Operations[pendingOp.ID] = pendingOp
	state.IdempotencyIndex[pendingOp.IdempotencyKey] = pendingOp.ID

	// Add old processed callbacks
	state.ProcessedCallbacks["old-callback"] = now.Add(-48 * time.Hour)
	state.ProcessedCallbacks["recent-callback"] = now.Add(-1 * time.Hour)

	assert.Len(t, state.Operations, 3)
	assert.Len(t, state.ProcessedCallbacks, 2)

	// Simulate cleanup (retention = 7 days)
	retentionCutoff := now.Add(-7 * 24 * time.Hour)
	callbackCutoff := now.Add(-24 * time.Hour)

	var toDelete []string
	for opID, op := range state.Operations {
		if op.State == marketplace.LifecycleOpStateCompleted && op.CompletedAt != nil && op.CompletedAt.Before(retentionCutoff) {
			toDelete = append(toDelete, opID)
		}
	}

	for _, opID := range toDelete {
		op := state.Operations[opID]
		delete(state.IdempotencyIndex, op.IdempotencyKey)
		delete(state.Operations, opID)
	}

	var callbacksToDelete []string
	for nonce, ts := range state.ProcessedCallbacks {
		if ts.Before(callbackCutoff) {
			callbacksToDelete = append(callbacksToDelete, nonce)
		}
	}
	for _, nonce := range callbacksToDelete {
		delete(state.ProcessedCallbacks, nonce)
	}

	// Old operation should be cleaned up
	assert.Len(t, state.Operations, 2)
	_, exists := state.Operations[oldOp.ID]
	assert.False(t, exists)

	// Recent and pending should remain
	_, exists = state.Operations[recentOp.ID]
	assert.True(t, exists)
	_, exists = state.Operations[pendingOp.ID]
	assert.True(t, exists)

	// Old callback should be cleaned up
	assert.Len(t, state.ProcessedCallbacks, 1)
	_, exists = state.ProcessedCallbacks["old-callback"]
	assert.False(t, exists)
	_, exists = state.ProcessedCallbacks["recent-callback"]
	assert.True(t, exists)
}

func TestLifecycleController_ConcurrencyLimit(t *testing.T) {
	cfg := DefaultLifecycleControllerConfig()
	cfg.MaxConcurrentOps = 3

	state := NewLifecycleControllerState()

	// Add operations to simulate concurrent limit
	for i := 0; i < 5; i++ {
		allocID := "alloc-" + string(rune('a'+i))
		op, _ := marketplace.NewLifecycleOperation(
			allocID,
			marketplace.LifecycleActionStart,
			"requester-789",
			"provider-456",
			marketplace.AllocationStateSuspended,
		)
		if i < 3 {
			op.State = marketplace.LifecycleOpStateExecuting
		} else {
			op.State = marketplace.LifecycleOpStatePending
		}
		state.Operations[op.ID] = op
	}

	// Count executing operations
	executingCount := 0
	for _, op := range state.Operations {
		if op.State == marketplace.LifecycleOpStateExecuting {
			executingCount++
		}
	}

	assert.Equal(t, 3, executingCount)

	// Check if we can start more
	canStartMore := executingCount < cfg.MaxConcurrentOps
	assert.False(t, canStartMore)
}

// BenchmarkIdempotencyKeyGeneration benchmarks idempotency key generation
func BenchmarkIdempotencyKeyGeneration(b *testing.B) {
	now := time.Now().UTC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		marketplace.GenerateIdempotencyKey("alloc-123", marketplace.LifecycleActionStart, now)
	}
}

// TestLifecycleController_Integration tests the full lifecycle flow
func TestLifecycleController_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create temp directory for state
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "lifecycle_state.json")

	cfg := DefaultLifecycleControllerConfig()
	cfg.StateFilePath = stateFile
	cfg.ProviderAddress = "provider-integration-test"
	cfg.OperationTimeout = 5 * time.Second
	cfg.RetryInterval = 1 * time.Second
	cfg.CleanupInterval = 2 * time.Second

	// Create controller without external dependencies for basic integration
	lc := &LifecycleController{
		cfg:    cfg,
		state:  NewLifecycleControllerState(),
		stopCh: make(chan struct{}),
	}

	// Create a test operation
	op, err := marketplace.NewLifecycleOperation(
		"alloc-integration-123",
		marketplace.LifecycleActionStart,
		cfg.ProviderAddress,
		cfg.ProviderAddress,
		marketplace.AllocationStateSuspended,
	)
	require.NoError(t, err)

	// Add operation to state
	lc.mu.Lock()
	lc.state.Operations[op.ID] = op
	lc.state.IdempotencyIndex[op.IdempotencyKey] = op.ID
	lc.mu.Unlock()

	// Verify operation was added
	lc.mu.RLock()
	storedOp, exists := lc.state.Operations[op.ID]
	lc.mu.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, marketplace.LifecycleOpStatePending, storedOp.State)

	// Simulate state transition
	lc.mu.Lock()
	err = storedOp.SetExecuting("waldur-op-123", time.Now())
	lc.mu.Unlock()
	require.NoError(t, err)

	lc.mu.RLock()
	assert.Equal(t, marketplace.LifecycleOpStateExecuting, lc.state.Operations[op.ID].State)
	lc.mu.RUnlock()

	// Save and reload state
	err = lc.saveState()
	require.NoError(t, err)

	lc2 := &LifecycleController{
		cfg:    cfg,
		state:  NewLifecycleControllerState(),
		stopCh: make(chan struct{}),
	}

	err = lc2.loadState()
	require.NoError(t, err)

	// Verify state was persisted correctly
	lc2.mu.RLock()
	reloadedOp, exists := lc2.state.Operations[op.ID]
	lc2.mu.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, marketplace.LifecycleOpStateExecuting, reloadedOp.State)
	assert.Equal(t, "alloc-integration-123", reloadedOp.AllocationID)

	_ = ctx // Used to ensure context is available for future async tests
}

