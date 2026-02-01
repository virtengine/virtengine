// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-14E: Resource lifecycle control via Waldur
// This file contains tests for the lifecycle control, waldur callbacks, and rollback functionality.
package provider_daemon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// TestResourceLifecycleConfig_Defaults tests default configuration
func TestResourceLifecycleConfig_Defaults(t *testing.T) {
	cfg := DefaultResourceLifecycleConfig()

	assert.Equal(t, 5*time.Minute, cfg.StartTimeout)
	assert.Equal(t, 5*time.Minute, cfg.StopTimeout)
	assert.Equal(t, 10*time.Minute, cfg.ResizeTimeout)
	assert.Equal(t, 10*time.Minute, cfg.TerminateTimeout)
	assert.Equal(t, 3, cfg.CleanupRetries)
	assert.Equal(t, 30*time.Second, cfg.CleanupRetryInterval)
	assert.True(t, cfg.EnablePreflightChecks)
	assert.True(t, cfg.EnablePostflightValidation)
}

// TestResourceInfo_Creation tests resource info creation
func TestResourceInfo_Creation(t *testing.T) {
	info := &ResourceInfo{
		AllocationID:       "alloc-123",
		WaldurResourceUUID: "uuid-456",
		ResourceType:       "vm",
		CurrentState:       marketplace.AllocationStateActive,
		ProviderAddress:    "provider-789",
		LastUpdated:        time.Now().UTC(),
		Metadata: map[string]string{
			"region": "us-east",
		},
	}

	assert.Equal(t, "alloc-123", info.AllocationID)
	assert.Equal(t, "uuid-456", info.WaldurResourceUUID)
	assert.Equal(t, marketplace.AllocationStateActive, info.CurrentState)
	assert.Equal(t, "us-east", info.Metadata["region"])
}

// TestResourceLifecycleManager_RegisterResource tests resource registration
func TestResourceLifecycleManager_RegisterResource(t *testing.T) {
	cfg := DefaultResourceLifecycleConfig()
	mgr := NewResourceLifecycleManager(cfg, nil, nil, nil)

	info := &ResourceInfo{
		AllocationID:       "alloc-123",
		WaldurResourceUUID: "uuid-456",
		ResourceType:       "vm",
		CurrentState:       marketplace.AllocationStateActive,
		ProviderAddress:    "provider-789",
	}

	err := mgr.RegisterResource(info)
	require.NoError(t, err)

	// Verify registration
	retrieved, found := mgr.GetResource("alloc-123")
	assert.True(t, found)
	assert.Equal(t, "uuid-456", retrieved.WaldurResourceUUID)
	assert.False(t, retrieved.LastUpdated.IsZero())

	// Test unregistration
	mgr.UnregisterResource("alloc-123")
	_, found = mgr.GetResource("alloc-123")
	assert.False(t, found)
}

// TestResourceLifecycleManager_RegisterResource_Errors tests registration errors
func TestResourceLifecycleManager_RegisterResource_Errors(t *testing.T) {
	cfg := DefaultResourceLifecycleConfig()
	mgr := NewResourceLifecycleManager(cfg, nil, nil, nil)

	tests := []struct {
		name    string
		info    *ResourceInfo
		wantErr bool
	}{
		{
			name:    "nil resource",
			info:    nil,
			wantErr: true,
		},
		{
			name:    "missing allocation ID",
			info:    &ResourceInfo{WaldurResourceUUID: "uuid-123"},
			wantErr: true,
		},
		{
			name:    "missing Waldur UUID",
			info:    &ResourceInfo{AllocationID: "alloc-123"},
			wantErr: true,
		},
		{
			name: "valid resource",
			info: &ResourceInfo{
				AllocationID:       "alloc-123",
				WaldurResourceUUID: "uuid-456",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.RegisterResource(tt.info)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestResourceLifecycleManager_UpdateResourceState tests state updates
func TestResourceLifecycleManager_UpdateResourceState(t *testing.T) {
	cfg := DefaultResourceLifecycleConfig()
	mgr := NewResourceLifecycleManager(cfg, nil, nil, nil)

	info := &ResourceInfo{
		AllocationID:       "alloc-123",
		WaldurResourceUUID: "uuid-456",
		CurrentState:       marketplace.AllocationStateActive,
	}
	_ = mgr.RegisterResource(info)

	// Update state
	err := mgr.UpdateResourceState("alloc-123", marketplace.AllocationStateSuspended)
	require.NoError(t, err)

	// Verify update
	retrieved, found := mgr.GetResource("alloc-123")
	assert.True(t, found)
	assert.Equal(t, marketplace.AllocationStateSuspended, retrieved.CurrentState)

	// Test update for non-existent resource
	err = mgr.UpdateResourceState("non-existent", marketplace.AllocationStateActive)
	assert.Equal(t, ErrResourceNotFound, err)
}

// TestWaldurCallbackConfig_Defaults tests default callback configuration
func TestWaldurCallbackConfig_Defaults(t *testing.T) {
	cfg := DefaultWaldurCallbackConfig()

	assert.Equal(t, ":8443", cfg.ListenAddr)
	assert.Equal(t, "/v1/callbacks/waldur", cfg.CallbackPath)
	assert.True(t, cfg.SignatureRequired)
	assert.Equal(t, 3600, cfg.NonceWindowSeconds)
	assert.Equal(t, int64(1<<20), cfg.MaxPayloadBytes)
	assert.True(t, cfg.EnableAuditLogging)
}

// TestNonceTracker tests nonce tracking functionality
func TestNonceTracker(t *testing.T) {
	tracker := NewNonceTracker(time.Hour)

	// Test initial state
	assert.False(t, tracker.IsProcessed("nonce-1"))

	// Mark as processed
	tracker.MarkProcessed("nonce-1")
	assert.True(t, tracker.IsProcessed("nonce-1"))

	// Test different nonce
	assert.False(t, tracker.IsProcessed("nonce-2"))

	// Mark second nonce
	tracker.MarkProcessed("nonce-2")
	assert.True(t, tracker.IsProcessed("nonce-2"))
}

// TestNonceTracker_Cleanup tests nonce cleanup
func TestNonceTracker_Cleanup(t *testing.T) {
	tracker := NewNonceTracker(100 * time.Millisecond)

	// Add nonces
	tracker.MarkProcessed("nonce-1")
	tracker.MarkProcessed("nonce-2")
	tracker.MarkProcessed("nonce-3")

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)

	// Cleanup
	count := tracker.Cleanup()
	assert.Equal(t, 3, count)

	// Verify nonces are cleaned
	assert.False(t, tracker.IsProcessed("nonce-1"))
	assert.False(t, tracker.IsProcessed("nonce-2"))
	assert.False(t, tracker.IsProcessed("nonce-3"))
}

// TestMapWaldurStateToAllocationState tests state mapping
func TestMapWaldurStateToAllocationState(t *testing.T) {
	tests := []struct {
		waldurState string
		expected    marketplace.AllocationState
	}{
		{"OK", marketplace.AllocationStateActive},
		{"done", marketplace.AllocationStateActive},
		{"completed", marketplace.AllocationStateActive},
		{"active", marketplace.AllocationStateActive},
		{"stopped", marketplace.AllocationStateSuspended},
		{"Stopped", marketplace.AllocationStateSuspended},
		{"paused", marketplace.AllocationStateSuspended},
		{"Paused", marketplace.AllocationStateSuspended},
		{"suspended", marketplace.AllocationStateSuspended},
		{"terminated", marketplace.AllocationStateTerminated},
		{"Terminated", marketplace.AllocationStateTerminated},
		{"deleted", marketplace.AllocationStateTerminated},
		{"creating", marketplace.AllocationStateProvisioning},
		{"Creating", marketplace.AllocationStateProvisioning},
		{"provisioning", marketplace.AllocationStateProvisioning},
		{"terminating", marketplace.AllocationStateTerminating},
		{"Terminating", marketplace.AllocationStateTerminating},
		{"deleting", marketplace.AllocationStateTerminating},
		{"erred", marketplace.AllocationStateFailed},
		{"Erred", marketplace.AllocationStateFailed},
		{"error", marketplace.AllocationStateFailed},
		{"failed", marketplace.AllocationStateFailed},
		{"unknown", marketplace.AllocationStateActive}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.waldurState, func(t *testing.T) {
			result := mapWaldurStateToAllocationState(tt.waldurState)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRollbackConfig_Defaults tests default rollback configuration
func TestRollbackConfig_Defaults(t *testing.T) {
	cfg := DefaultRollbackConfig()

	assert.True(t, cfg.Enabled)
	assert.Equal(t, 3, cfg.MaxRollbackAttempts)
	assert.Equal(t, 10*time.Minute, cfg.RollbackTimeout)
	assert.Equal(t, 30*time.Second, cfg.RetryInterval)
	assert.True(t, cfg.EnableAuditLogging)
	assert.True(t, cfg.PreserveState)
}

// TestRollbackActionMap tests rollback action mapping
func TestRollbackActionMap(t *testing.T) {
	tests := []struct {
		action         marketplace.LifecycleActionType
		rollbackAction marketplace.LifecycleActionType
		supported      bool
	}{
		{marketplace.LifecycleActionStart, marketplace.LifecycleActionStop, true},
		{marketplace.LifecycleActionStop, marketplace.LifecycleActionStart, true},
		{marketplace.LifecycleActionSuspend, marketplace.LifecycleActionResume, true},
		{marketplace.LifecycleActionResume, marketplace.LifecycleActionSuspend, true},
		{marketplace.LifecycleActionTerminate, "", false},
		{marketplace.LifecycleActionProvision, "", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			rollback, ok := RollbackActionMap[tt.action]
			assert.Equal(t, tt.supported, ok)
			if tt.supported {
				assert.Equal(t, tt.rollbackAction, rollback)
			}
		})
	}
}

// TestCanRollback tests rollback capability check
func TestCanRollback(t *testing.T) {
	assert.True(t, CanRollback(marketplace.LifecycleActionStart))
	assert.True(t, CanRollback(marketplace.LifecycleActionStop))
	assert.True(t, CanRollback(marketplace.LifecycleActionSuspend))
	assert.True(t, CanRollback(marketplace.LifecycleActionResume))
	assert.False(t, CanRollback(marketplace.LifecycleActionTerminate))
	assert.False(t, CanRollback(marketplace.LifecycleActionProvision))
}

// TestGetRollbackAction tests getting rollback actions
func TestGetRollbackAction(t *testing.T) {
	action, ok := GetRollbackAction(marketplace.LifecycleActionStart)
	assert.True(t, ok)
	assert.Equal(t, marketplace.LifecycleActionStop, action)

	action, ok = GetRollbackAction(marketplace.LifecycleActionTerminate)
	assert.False(t, ok)
}

// TestRollbackState_IsTerminal tests terminal state check
func TestRollbackState_IsTerminal(t *testing.T) {
	tests := []struct {
		state    RollbackState
		terminal bool
	}{
		{RollbackStatePending, false},
		{RollbackStateExecuting, false},
		{RollbackStateCompleted, true},
		{RollbackStateFailed, true},
		{RollbackStateSkipped, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			assert.Equal(t, tt.terminal, tt.state.IsTerminal())
		})
	}
}

// TestRollbackRecord_Creation tests rollback record creation
func TestRollbackRecord_Creation(t *testing.T) {
	now := time.Now().UTC()
	record := &RollbackRecord{
		ID:                  "rb-123",
		OriginalOperationID: "op-456",
		AllocationID:        "alloc-789",
		OriginalAction:      marketplace.LifecycleActionStart,
		RollbackAction:      marketplace.LifecycleActionStop,
		State:               RollbackStatePending,
		AttemptCount:        0,
		OriginalState:       marketplace.AllocationStateSuspended,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	assert.Equal(t, "rb-123", record.ID)
	assert.Equal(t, marketplace.LifecycleActionStart, record.OriginalAction)
	assert.Equal(t, marketplace.LifecycleActionStop, record.RollbackAction)
	assert.Equal(t, RollbackStatePending, record.State)
	assert.False(t, record.State.IsTerminal())
}

// TestRollbackPlanGenerator tests rollback plan generation
func TestRollbackPlanGenerator(t *testing.T) {
	generator := NewRollbackPlanGenerator()

	// Create test operation
	op, err := marketplace.NewLifecycleOperation(
		"alloc-123",
		marketplace.LifecycleActionStart,
		"requester",
		"provider",
		marketplace.AllocationStateSuspended,
	)
	require.NoError(t, err)

	// Generate plan
	plan, err := generator.GeneratePlan(op)
	require.NoError(t, err)

	assert.Equal(t, marketplace.LifecycleActionStop, plan.RollbackAction)
	assert.Equal(t, marketplace.AllocationStateSuspended, plan.TargetState)
	assert.Len(t, plan.Steps, 3)
	assert.Equal(t, "validate_state", plan.Steps[0].Name)
	assert.Equal(t, "execute_rollback", plan.Steps[1].Name)
	assert.Equal(t, "verify_state", plan.Steps[2].Name)
}

// TestRollbackPlanGenerator_UnsupportedAction tests plan generation for unsupported actions
func TestRollbackPlanGenerator_UnsupportedAction(t *testing.T) {
	generator := NewRollbackPlanGenerator()

	op, err := marketplace.NewLifecycleOperation(
		"alloc-123",
		marketplace.LifecycleActionTerminate,
		"requester",
		"provider",
		marketplace.AllocationStateActive,
	)
	require.NoError(t, err)

	_, err = generator.GeneratePlan(op)
	assert.ErrorIs(t, err, ErrRollbackNotSupported)
}

// TestCallbackBatcher tests callback batching
func TestCallbackBatcher(t *testing.T) {
	batcher := NewCallbackBatcher(3)

	// Add callbacks
	cb1 := &marketplace.WaldurCallback{ID: "cb-1"}
	cb2 := &marketplace.WaldurCallback{ID: "cb-2"}
	cb3 := &marketplace.WaldurCallback{ID: "cb-3"}
	cb4 := &marketplace.WaldurCallback{ID: "cb-4"}

	assert.True(t, batcher.Add(cb1))
	assert.True(t, batcher.Add(cb2))
	assert.True(t, batcher.Add(cb3))
	assert.False(t, batcher.Add(cb4)) // Full

	assert.Equal(t, 3, batcher.Size())

	// Flush
	batch := batcher.Flush()
	assert.Len(t, batch, 3)
	assert.Equal(t, 0, batcher.Size())

	// Can add again after flush
	assert.True(t, batcher.Add(cb4))
}

// TestCallbackSignatureVerifier tests signature verification setup
func TestCallbackSignatureVerifier(t *testing.T) {
	verifier := NewCallbackSignatureVerifier()

	// Verify empty verifier returns error
	callback := &marketplace.WaldurCallback{
		ID:        "cb-123",
		SignerID:  "unknown-signer",
		Signature: make([]byte, 64),
	}

	err := verifier.Verify(callback)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "public key not found")
}

// TestComputeCallbackHash tests callback hash computation
func TestComputeCallbackHash(t *testing.T) {
	callback := &marketplace.WaldurCallback{
		ID:            "cb-123",
		WaldurID:      "waldur-456",
		ChainEntityID: "chain-789",
		Nonce:         "nonce-abc",
	}

	hash := ComputeCallbackHash(callback)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64) // SHA256 hex string

	// Same callback should produce same hash
	hash2 := ComputeCallbackHash(callback)
	assert.Equal(t, hash, hash2)

	// Different callback should produce different hash
	callback2 := &marketplace.WaldurCallback{
		ID:            "cb-different",
		WaldurID:      "waldur-456",
		ChainEntityID: "chain-789",
		Nonce:         "nonce-abc",
	}
	hash3 := ComputeCallbackHash(callback2)
	assert.NotEqual(t, hash, hash3)
}

// TestWaldurCallbackHandler_HealthEndpoint tests health endpoint
func TestWaldurCallbackHandler_HealthEndpoint(t *testing.T) {
	cfg := DefaultWaldurCallbackConfig()
	handler := NewWaldurCallbackHandler(cfg, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.handleHealth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

// TestWaldurCallbackHandler_CallbackEndpoint_MethodNotAllowed tests method validation
func TestWaldurCallbackHandler_CallbackEndpoint_MethodNotAllowed(t *testing.T) {
	cfg := DefaultWaldurCallbackConfig()
	handler := NewWaldurCallbackHandler(cfg, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/callbacks/waldur", nil)
	w := httptest.NewRecorder()

	handler.handleCallback(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestWaldurCallbackHandler_CallbackEndpoint_InvalidJSON tests JSON validation
func TestWaldurCallbackHandler_CallbackEndpoint_InvalidJSON(t *testing.T) {
	cfg := DefaultWaldurCallbackConfig()
	cfg.SignatureRequired = false // Disable for testing
	handler := NewWaldurCallbackHandler(cfg, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/v1/callbacks/waldur",
		strings.NewReader("invalid json"))
	w := httptest.NewRecorder()

	handler.handleCallback(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid JSON")
}

// TestRollbackManager_GetMetrics tests metrics retrieval
func TestRollbackManager_GetMetrics(t *testing.T) {
	cfg := DefaultRollbackConfig()
	mgr := NewRollbackManager(cfg, nil, nil, nil)

	// Add some test records
	mgr.records["rb-1"] = &RollbackRecord{
		ID:             "rb-1",
		State:          RollbackStateCompleted,
		OriginalAction: marketplace.LifecycleActionStart,
	}
	mgr.records["rb-2"] = &RollbackRecord{
		ID:             "rb-2",
		State:          RollbackStateFailed,
		OriginalAction: marketplace.LifecycleActionStop,
	}
	mgr.records["rb-3"] = &RollbackRecord{
		ID:             "rb-3",
		State:          RollbackStatePending,
		OriginalAction: marketplace.LifecycleActionStart,
	}

	metrics := mgr.GetRollbackMetrics()

	assert.Equal(t, int64(3), metrics.TotalRollbacks)
	assert.Equal(t, int64(1), metrics.SuccessfulRollbacks)
	assert.Equal(t, int64(1), metrics.FailedRollbacks)
	assert.Equal(t, int64(0), metrics.ActiveRollbacks)
	assert.Equal(t, int64(1), metrics.ByState[RollbackStateCompleted])
	assert.Equal(t, int64(1), metrics.ByState[RollbackStateFailed])
	assert.Equal(t, int64(1), metrics.ByState[RollbackStatePending])
	assert.Equal(t, int64(2), metrics.ByAction[marketplace.LifecycleActionStart])
	assert.Equal(t, int64(1), metrics.ByAction[marketplace.LifecycleActionStop])
}

// TestRollbackManager_CancelRollback tests rollback cancellation
func TestRollbackManager_CancelRollback(t *testing.T) {
	cfg := DefaultRollbackConfig()
	mgr := NewRollbackManager(cfg, nil, nil, nil)

	// Add pending rollback
	mgr.records["rb-1"] = &RollbackRecord{
		ID:           "rb-1",
		AllocationID: "alloc-123",
		State:        RollbackStatePending,
	}
	mgr.activeRollbacks["alloc-123"] = "rb-1"

	// Cancel it
	err := mgr.CancelRollback("rb-1")
	require.NoError(t, err)

	// Verify state
	record, found := mgr.GetRollbackRecord("rb-1")
	assert.True(t, found)
	assert.Equal(t, RollbackStateSkipped, record.State)
	assert.NotNil(t, record.CompletedAt)
	assert.Contains(t, record.Error, "cancelled")

	// Verify removed from active
	_, found = mgr.GetActiveRollback("alloc-123")
	assert.False(t, found)
}

// TestRollbackManager_CancelRollback_Errors tests cancellation errors
func TestRollbackManager_CancelRollback_Errors(t *testing.T) {
	cfg := DefaultRollbackConfig()
	mgr := NewRollbackManager(cfg, nil, nil, nil)

	// Cancel non-existent
	err := mgr.CancelRollback("non-existent")
	assert.Error(t, err)

	// Add completed rollback
	completedAt := time.Now().UTC()
	mgr.records["rb-1"] = &RollbackRecord{
		ID:          "rb-1",
		State:       RollbackStateCompleted,
		CompletedAt: &completedAt,
	}

	// Try to cancel completed
	err = mgr.CancelRollback("rb-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already completed")
}

// TestRollbackManager_CleanupOldRecords tests record cleanup
func TestRollbackManager_CleanupOldRecords(t *testing.T) {
	cfg := DefaultRollbackConfig()
	mgr := NewRollbackManager(cfg, nil, nil, nil)

	now := time.Now().UTC()
	oldTime := now.Add(-10 * 24 * time.Hour)
	recentTime := now.Add(-1 * 24 * time.Hour)

	// Add old completed record
	mgr.records["rb-old"] = &RollbackRecord{
		ID:          "rb-old",
		State:       RollbackStateCompleted,
		CompletedAt: &oldTime,
	}

	// Add recent completed record
	mgr.records["rb-recent"] = &RollbackRecord{
		ID:          "rb-recent",
		State:       RollbackStateCompleted,
		CompletedAt: &recentTime,
	}

	// Add pending record (should not be cleaned)
	mgr.records["rb-pending"] = &RollbackRecord{
		ID:    "rb-pending",
		State: RollbackStatePending,
	}

	// Cleanup with 7 day retention
	count := mgr.CleanupOldRecords(7)
	assert.Equal(t, 1, count)

	// Verify only old record was cleaned
	_, found := mgr.GetRollbackRecord("rb-old")
	assert.False(t, found)

	_, found = mgr.GetRollbackRecord("rb-recent")
	assert.True(t, found)

	_, found = mgr.GetRollbackRecord("rb-pending")
	assert.True(t, found)
}

// TestLifecycleActionRequest tests action request creation
func TestLifecycleActionRequest(t *testing.T) {
	req := &LifecycleActionRequest{
		AllocationID: "alloc-123",
		Action:       marketplace.LifecycleActionStart,
		RequestedBy:  "user-456",
		Immediate:    true,
		Parameters: map[string]string{
			"force": "true",
		},
		Timeout: 5 * time.Minute,
	}

	assert.Equal(t, "alloc-123", req.AllocationID)
	assert.Equal(t, marketplace.LifecycleActionStart, req.Action)
	assert.True(t, req.Immediate)
	assert.Equal(t, "true", req.Parameters["force"])
	assert.Equal(t, 5*time.Minute, req.Timeout)
}

// TestLifecycleActionRequest_WithResizeSpec tests resize specification
func TestLifecycleActionRequest_WithResizeSpec(t *testing.T) {
	cpuCores := uint32(8)
	memoryMB := uint64(16384)
	storageGB := uint64(500)

	req := &LifecycleActionRequest{
		AllocationID: "alloc-123",
		Action:       marketplace.LifecycleActionResize,
		RequestedBy:  "user-456",
		ResizeSpec: &marketplace.ResizeSpecification{
			CPUCores:  &cpuCores,
			MemoryMB:  &memoryMB,
			StorageGB: &storageGB,
		},
	}

	assert.NotNil(t, req.ResizeSpec)
	assert.Equal(t, uint32(8), *req.ResizeSpec.CPUCores)
	assert.Equal(t, uint64(16384), *req.ResizeSpec.MemoryMB)
	assert.Equal(t, uint64(500), *req.ResizeSpec.StorageGB)
}

// TestLifecycleActionResult tests action result creation
func TestLifecycleActionResult(t *testing.T) {
	now := time.Now().UTC()
	completedAt := now.Add(30 * time.Second)

	result := &LifecycleActionResult{
		OperationID:       "op-123",
		Success:           true,
		State:             marketplace.LifecycleOpStateCompleted,
		NewResourceState:  marketplace.AllocationStateActive,
		WaldurOperationID: "waldur-456",
		Message:           "Operation completed successfully",
		StartedAt:         now,
		CompletedAt:       &completedAt,
	}

	assert.Equal(t, "op-123", result.OperationID)
	assert.True(t, result.Success)
	assert.Equal(t, marketplace.LifecycleOpStateCompleted, result.State)
	assert.Equal(t, marketplace.AllocationStateActive, result.NewResourceState)
	assert.Equal(t, "waldur-456", result.WaldurOperationID)
	assert.NotNil(t, result.CompletedAt)
}

// TestSerializeCallbackForSigning tests callback serialization
func TestSerializeCallbackForSigning(t *testing.T) {
	timestamp := time.Unix(1700000000, 0).UTC()
	callback := &marketplace.WaldurCallback{
		ID:              "cb-123",
		WaldurID:        "waldur-456",
		ChainEntityID:   "chain-789",
		ChainEntityType: marketplace.SyncTypeAllocation,
		ActionType:      marketplace.ActionTypeStatusUpdate,
		SignerID:        "signer-abc",
		Nonce:           "nonce-def",
		Timestamp:       timestamp,
	}

	serialized := SerializeCallbackForSigning(callback)
	assert.NotEmpty(t, serialized)

	// Verify it contains expected parts
	serializedStr := string(serialized)
	assert.Contains(t, serializedStr, "cb-123")
	assert.Contains(t, serializedStr, "waldur-456")
	assert.Contains(t, serializedStr, "chain-789")
	assert.Contains(t, serializedStr, "signer-abc")
	assert.Contains(t, serializedStr, "nonce-def")
}

// TestAuditLogger_LifecycleEvents tests lifecycle event logging
func TestAuditLogger_LifecycleEvents(t *testing.T) {
	cfg := DefaultAuditLogConfig()
	cfg.LogFile = "" // Disable file logging for test

	logger, err := NewAuditLogger(cfg)
	require.NoError(t, err)
	defer logger.Close()

	// Log lifecycle action
	err = logger.LogLifecycleAction(
		AuditEventLifecycleActionCompleted,
		"alloc-123",
		"start",
		"op-456",
		true,
		"",
		nil,
	)
	require.NoError(t, err)

	// Log state change
	err = logger.LogLifecycleStateChange(
		"alloc-123",
		"suspended",
		"active",
		"user_request",
	)
	require.NoError(t, err)

	// Log callback
	err = logger.LogLifecycleCallback(
		"cb-789",
		"op-456",
		"alloc-123",
		true,
		"",
	)
	require.NoError(t, err)

	// Log rollback
	err = logger.LogRollback(
		"rb-111",
		"op-456",
		"alloc-123",
		"start",
		"stop",
		true,
		"",
	)
	require.NoError(t, err)

	// Verify events were logged
	since := time.Now().Add(-time.Hour)
	events := logger.GetLifecycleEvents(since)
	assert.GreaterOrEqual(t, len(events), 4)

	// Verify by allocation
	allocEvents := logger.GetEventsByAllocation("alloc-123", since)
	assert.GreaterOrEqual(t, len(allocEvents), 4)
}

// TestAuditLogger_WaldurCallback tests Waldur callback logging
func TestAuditLogger_WaldurCallback(t *testing.T) {
	cfg := DefaultAuditLogConfig()
	cfg.LogFile = "" // Disable file logging for test

	logger, err := NewAuditLogger(cfg)
	require.NoError(t, err)
	defer logger.Close()

	// Log successful callback
	err = logger.LogWaldurCallback(
		"cb-123",
		"waldur-456",
		"chain-789",
		"status_update",
		true,
		"",
	)
	require.NoError(t, err)

	// Log failed callback
	err = logger.LogWaldurCallback(
		"cb-456",
		"waldur-789",
		"chain-abc",
		"create",
		false,
		"signature verification failed",
	)
	require.NoError(t, err)

	// Verify events
	since := time.Now().Add(-time.Hour)
	events := logger.GetLifecycleEvents(since)
	assert.GreaterOrEqual(t, len(events), 2)
}

// MockWaldurServer creates a mock Waldur server for testing
type MockWaldurServer struct {
	*httptest.Server
	Responses   map[string]MockResponse
	ReceivedRequests []MockRequest
	mu          sync.Mutex
}

// MockResponse represents a mock response
type MockResponse struct {
	StatusCode int
	Body       interface{}
}

// MockRequest represents a received request
type MockRequest struct {
	Method string
	Path   string
	Body   json.RawMessage
}

// NewMockWaldurServer creates a new mock Waldur server
func NewMockWaldurServer() *MockWaldurServer {
	mock := &MockWaldurServer{
		Responses: make(map[string]MockResponse),
		ReceivedRequests: make([]MockRequest, 0),
	}

	mock.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mock.mu.Lock()
		body, _ := json.Marshal(r.Body)
		mock.ReceivedRequests = append(mock.ReceivedRequests, MockRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Body:   body,
		})
		mock.mu.Unlock()

		key := r.Method + " " + r.URL.Path
		resp, ok := mock.Responses[key]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(resp.StatusCode)
		if resp.Body != nil {
			_ = json.NewEncoder(w).Encode(resp.Body)
		}
	}))

	return mock
}

// AddResponse adds a mock response
func (m *MockWaldurServer) AddResponse(method, path string, status int, body interface{}) {
	m.Responses[method+" "+path] = MockResponse{
		StatusCode: status,
		Body:       body,
	}
}

// TestMockWaldurServer tests the mock server works
func TestMockWaldurServer(t *testing.T) {
	server := NewMockWaldurServer()
	defer server.Close()

	server.AddResponse("GET", "/api/v1/health", http.StatusOK, map[string]string{"status": "ok"})

	resp, err := http.Get(server.URL + "/api/v1/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// BenchmarkNonceTracker benchmarks nonce operations
func BenchmarkNonceTracker(b *testing.B) {
	tracker := NewNonceTracker(time.Hour)

	b.Run("MarkProcessed", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tracker.MarkProcessed("nonce-" + string(rune(i)))
		}
	})

	// Pre-populate
	for i := 0; i < 10000; i++ {
		tracker.MarkProcessed("nonce-" + string(rune(i)))
	}

	b.Run("IsProcessed", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tracker.IsProcessed("nonce-" + string(rune(i%10000)))
		}
	})
}

// BenchmarkCallbackHash benchmarks callback hash computation
func BenchmarkCallbackHash(b *testing.B) {
	callback := &marketplace.WaldurCallback{
		ID:            "cb-123",
		WaldurID:      "waldur-456",
		ChainEntityID: "chain-789",
		Nonce:         "nonce-abc",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeCallbackHash(callback)
	}
}

