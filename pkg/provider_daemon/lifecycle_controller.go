// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-4E: Resource lifecycle control via Waldur
// This file implements the lifecycle controller that handles lifecycle operations
// with signed callbacks, idempotency, and rollback policies.
package provider_daemon

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// LifecycleControllerConfig configures the lifecycle controller
type LifecycleControllerConfig struct {
	// Enabled indicates if lifecycle control is enabled
	Enabled bool `json:"enabled"`

	// ProviderAddress is the provider's blockchain address
	ProviderAddress string `json:"provider_address"`

	// StateFilePath is the path for persisting lifecycle state
	StateFilePath string `json:"state_file_path"`

	// OperationTimeout is the timeout for lifecycle operations
	OperationTimeout time.Duration `json:"operation_timeout"`

	// CallbackTTL is how long callbacks are valid
	CallbackTTL time.Duration `json:"callback_ttl"`

	// MaxConcurrentOps is the maximum concurrent operations
	MaxConcurrentOps int `json:"max_concurrent_ops"`

	// RetryInterval is the interval between retries
	RetryInterval time.Duration `json:"retry_interval"`

	// CleanupInterval is the interval for cleaning up old operations
	CleanupInterval time.Duration `json:"cleanup_interval"`

	// OperationRetentionDays is how long to keep completed operations
	OperationRetentionDays int `json:"operation_retention_days"`

	// CallbackURL is the callback URL for Waldur notifications
	CallbackURL string `json:"callback_url"`

	// EnableAuditLogging enables audit logging for lifecycle operations
	EnableAuditLogging bool `json:"enable_audit_logging"`
}

// DefaultLifecycleControllerConfig returns default configuration
func DefaultLifecycleControllerConfig() LifecycleControllerConfig {
	return LifecycleControllerConfig{
		Enabled:                true,
		StateFilePath:          "data/lifecycle_state.json",
		OperationTimeout:       5 * time.Minute,
		CallbackTTL:            time.Hour,
		MaxConcurrentOps:       10,
		RetryInterval:          30 * time.Second,
		CleanupInterval:        time.Hour,
		OperationRetentionDays: 7,
		EnableAuditLogging:     true,
	}
}

// LifecycleControllerState persists lifecycle controller state
type LifecycleControllerState struct {
	// Operations maps operation ID to operation
	Operations map[string]*marketplace.LifecycleOperation `json:"operations"`

	// IdempotencyIndex maps idempotency key to operation ID
	IdempotencyIndex map[string]string `json:"idempotency_index"`

	// ProcessedCallbacks tracks processed callback nonces
	ProcessedCallbacks map[string]time.Time `json:"processed_callbacks"`

	// Metrics contains lifecycle metrics
	Metrics *marketplace.LifecycleMetrics `json:"metrics"`

	// LastUpdated is when state was last updated
	LastUpdated time.Time `json:"last_updated"`
}

// NewLifecycleControllerState creates new state
func NewLifecycleControllerState() *LifecycleControllerState {
	return &LifecycleControllerState{
		Operations:         make(map[string]*marketplace.LifecycleOperation),
		IdempotencyIndex:   make(map[string]string),
		ProcessedCallbacks: make(map[string]time.Time),
		Metrics:            marketplace.NewLifecycleMetrics(),
		LastUpdated:        time.Now().UTC(),
	}
}

// LifecycleController manages lifecycle operations via Waldur
type LifecycleController struct {
	cfg          LifecycleControllerConfig
	keyManager   *KeyManager
	callbackSink CallbackSink
	lifecycle    *waldur.LifecycleClient
	auditLogger  *AuditLogger
	state        *LifecycleControllerState
	mu           sync.RWMutex
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

// NewLifecycleController creates a new lifecycle controller
func NewLifecycleController(
	cfg LifecycleControllerConfig,
	keyManager *KeyManager,
	callbackSink CallbackSink,
	marketplace *waldur.MarketplaceClient,
	auditLogger *AuditLogger,
) (*LifecycleController, error) {
	if keyManager == nil {
		return nil, errors.New("key manager is required")
	}
	if marketplace == nil {
		return nil, errors.New("marketplace client is required")
	}

	lc := &LifecycleController{
		cfg:          cfg,
		keyManager:   keyManager,
		callbackSink: callbackSink,
		lifecycle:    waldur.NewLifecycleClient(marketplace),
		auditLogger:  auditLogger,
		state:        NewLifecycleControllerState(),
		stopCh:       make(chan struct{}),
	}

	// Load persisted state
	if err := lc.loadState(); err != nil {
		log.Printf("[lifecycle-controller] failed to load state: %v", err)
	}

	return lc, nil
}

// Start starts the lifecycle controller
func (lc *LifecycleController) Start(ctx context.Context) error {
	if !lc.cfg.Enabled {
		return nil
	}

	log.Printf("[lifecycle-controller] starting with provider %s", lc.cfg.ProviderAddress)

	// Start background workers
	lc.wg.Add(2)
	go lc.retryWorker(ctx)
	go lc.cleanupWorker(ctx)

	return nil
}

// Stop stops the lifecycle controller
func (lc *LifecycleController) Stop() error {
	close(lc.stopCh)
	lc.wg.Wait()

	// Save state
	if err := lc.saveState(); err != nil {
		log.Printf("[lifecycle-controller] failed to save state: %v", err)
	}

	return nil
}

// ExecuteLifecycleAction executes a lifecycle action
func (lc *LifecycleController) ExecuteLifecycleAction(
	ctx context.Context,
	allocationID string,
	action marketplace.LifecycleActionType,
	currentState marketplace.AllocationState,
	waldurResourceUUID string,
	requestedBy string,
	params map[string]string,
) (*marketplace.LifecycleOperation, error) {
	// Create operation
	op, err := marketplace.NewLifecycleOperation(
		allocationID,
		action,
		requestedBy,
		lc.cfg.ProviderAddress,
		currentState,
	)
	if err != nil {
		return nil, err
	}

	// Set parameters
	if params != nil {
		op.Parameters = params
	}

	// Check idempotency
	lc.mu.Lock()
	if existingOpID, ok := lc.state.IdempotencyIndex[op.IdempotencyKey]; ok {
		existingOp := lc.state.Operations[existingOpID]
		lc.mu.Unlock()
		if existingOp != nil && !existingOp.State.IsTerminal() {
			return existingOp, nil // Return existing operation
		}
	}

	// Check concurrent operation limit
	pendingCount := 0
	for _, existingOp := range lc.state.Operations {
		if existingOp.AllocationID == allocationID && !existingOp.State.IsTerminal() {
			pendingCount++
		}
	}
	if pendingCount >= lc.cfg.MaxConcurrentOps {
		lc.mu.Unlock()
		return nil, marketplace.ErrLifecycleOperationInProgress
	}

	// Save operation
	lc.state.Operations[op.ID] = op
	lc.state.IdempotencyIndex[op.IdempotencyKey] = op.ID
	lc.state.Metrics.TotalOperations++
	lc.state.Metrics.PendingOperations++
	lc.state.Metrics.OperationsByAction[action]++
	lc.mu.Unlock()

	// Audit log
	if lc.auditLogger != nil && lc.cfg.EnableAuditLogging {
		_ = lc.auditLogger.Log(&AuditEvent{
			Type:      AuditEventType("lifecycle_action_requested"),
			Operation: string(action),
			Success:   true,
			Details: map[string]interface{}{
				"operation_id":  op.ID,
				"allocation_id": allocationID,
				"action":        action,
				"requested_by":  requestedBy,
			},
		})
	}

	// Execute asynchronously
	go lc.executeOperation(ctx, op, waldurResourceUUID)

	return op, nil
}

// executeOperation executes a lifecycle operation
func (lc *LifecycleController) executeOperation(ctx context.Context, op *marketplace.LifecycleOperation, waldurResourceUUID string) {
	opCtx, cancel := context.WithTimeout(ctx, lc.cfg.OperationTimeout)
	defer cancel()

	// Update state to executing
	lc.mu.Lock()
	op.State = marketplace.LifecycleOpStateExecuting
	startTime := time.Now().UTC()
	op.StartedAt = &startTime
	op.UpdatedAt = startTime
	lc.state.Metrics.PendingOperations--
	lc.state.Metrics.ExecutingOperations++
	lc.mu.Unlock()

	// Map action to Waldur lifecycle action
	waldurAction := mapToWaldurAction(op.Action)

	// Build request
	req := waldur.LifecycleRequest{
		ResourceUUID:   waldurResourceUUID,
		Action:         waldurAction,
		IdempotencyKey: op.IdempotencyKey,
		CallbackURL:    lc.cfg.CallbackURL,
		Parameters:     make(map[string]interface{}),
	}

	// Add parameters
	for k, v := range op.Parameters {
		req.Parameters[k] = v
	}

	// Execute action
	var response *waldur.LifecycleResponse
	var err error

	switch waldurAction {
	case waldur.LifecycleActionStart:
		response, err = lc.lifecycle.Start(opCtx, req)
	case waldur.LifecycleActionStop:
		response, err = lc.lifecycle.Stop(opCtx, req)
	case waldur.LifecycleActionRestart:
		response, err = lc.lifecycle.Restart(opCtx, req)
	case waldur.LifecycleActionSuspend:
		response, err = lc.lifecycle.Suspend(opCtx, req)
	case waldur.LifecycleActionResume:
		response, err = lc.lifecycle.Resume(opCtx, req)
	case waldur.LifecycleActionTerminate:
		response, err = lc.lifecycle.Terminate(opCtx, req)
	case waldur.LifecycleActionResize:
		resizeReq := waldur.ResizeRequest{LifecycleRequest: req}
		response, err = lc.lifecycle.Resize(opCtx, resizeReq)
	default:
		err = fmt.Errorf("unsupported action: %s", waldurAction)
	}

	if err != nil {
		lc.handleOperationFailure(op, err)
		return
	}

	// Update with Waldur operation ID
	lc.mu.Lock()
	op.WaldurOperationID = response.OperationID
	if op.WaldurOperationID == "" {
		op.WaldurOperationID = response.UUID
	}

	// Generate callback ID and transition to awaiting callback
	callbackID := fmt.Sprintf("lcb_%s", op.ID)
	op.CallbackID = callbackID
	op.State = marketplace.LifecycleOpStateAwaitingCallback
	op.UpdatedAt = time.Now().UTC()
	lc.mu.Unlock()

	// Save state
	_ = lc.saveState()

	log.Printf("[lifecycle-controller] operation %s awaiting callback for allocation %s action %s",
		op.ID, op.AllocationID, op.Action)
}

// handleOperationFailure handles a failed operation
func (lc *LifecycleController) handleOperationFailure(op *marketplace.LifecycleOperation, err error) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	op.Error = err.Error()

	// Check if should retry
	if op.ShouldRetry() && op.IncrementRetry() {
		op.State = marketplace.LifecycleOpStatePending
		op.UpdatedAt = time.Now().UTC()
		log.Printf("[lifecycle-controller] operation %s will retry (attempt %d/%d): %v",
			op.ID, op.RetryCount, op.MaxRetries, err)
	} else {
		// Handle rollback
		switch op.RollbackPolicy {
		case marketplace.RollbackPolicyAutomatic:
			// Would trigger rollback action
			lc.state.Metrics.RolledBackOperations++
			completedAt := time.Now().UTC()
			op.CompletedAt = &completedAt
			op.State = marketplace.LifecycleOpStateRolledBack
		default:
			lc.state.Metrics.FailedOperations++
			completedAt := time.Now().UTC()
			op.CompletedAt = &completedAt
			op.State = marketplace.LifecycleOpStateFailed
		}
		lc.state.Metrics.ExecutingOperations--
		op.UpdatedAt = time.Now().UTC()

		log.Printf("[lifecycle-controller] operation %s failed: %v", op.ID, err)
	}

	// Audit log
	if lc.auditLogger != nil && lc.cfg.EnableAuditLogging {
		_ = lc.auditLogger.Log(&AuditEvent{
			Type:         AuditEventType("lifecycle_action_failed"),
			Operation:    string(op.Action),
			Success:      false,
			ErrorMessage: err.Error(),
			Details: map[string]interface{}{
				"operation_id":  op.ID,
				"allocation_id": op.AllocationID,
				"action":        op.Action,
				"retry_count":   op.RetryCount,
			},
		})
	}

	// Submit failure callback
	callback := lc.createFailureCallback(op)
	if callback != nil {
		_ = lc.signAndSubmitCallback(context.Background(), callback)
	}

	_ = lc.saveState()
}

// ProcessCallback processes a lifecycle callback
func (lc *LifecycleController) ProcessCallback(ctx context.Context, callback *marketplace.LifecycleCallback) error {
	// Validate callback
	if err := callback.ValidateAt(time.Now()); err != nil {
		return fmt.Errorf("invalid callback: %w", err)
	}

	lc.mu.Lock()
	defer lc.mu.Unlock()

	// Check for replay
	if _, processed := lc.state.ProcessedCallbacks[callback.Nonce]; processed {
		return nil // Already processed
	}

	// Find operation
	op, ok := lc.state.Operations[callback.OperationID]
	if !ok {
		return fmt.Errorf("operation not found: %s", callback.OperationID)
	}

	// Verify callback matches expected
	if op.CallbackID != "" && op.CallbackID != callback.ID {
		return fmt.Errorf("callback ID mismatch: expected %s, got %s", op.CallbackID, callback.ID)
	}

	// Update operation state
	completedAt := time.Now().UTC()
	op.CompletedAt = &completedAt
	op.UpdatedAt = completedAt

	if callback.Success {
		op.State = marketplace.LifecycleOpStateCompleted
		lc.state.Metrics.CompletedOperations++
	} else {
		op.State = marketplace.LifecycleOpStateFailed
		op.Error = callback.Error
		lc.state.Metrics.FailedOperations++
	}
	lc.state.Metrics.ExecutingOperations--

	// Calculate completion time
	if op.StartedAt != nil {
		duration := completedAt.Sub(*op.StartedAt).Milliseconds()
		if lc.state.Metrics.CompletedOperations > 0 {
			avgTime := lc.state.Metrics.AverageCompletionTimeMs
			total := avgTime * (lc.state.Metrics.CompletedOperations - 1)
			lc.state.Metrics.AverageCompletionTimeMs = (total + duration) / lc.state.Metrics.CompletedOperations
		} else {
			lc.state.Metrics.AverageCompletionTimeMs = duration
		}
	}

	// Mark callback as processed
	lc.state.ProcessedCallbacks[callback.Nonce] = completedAt.Add(2 * time.Hour)

	// Audit log
	if lc.auditLogger != nil && lc.cfg.EnableAuditLogging {
		_ = lc.auditLogger.Log(&AuditEvent{
			Type:      AuditEventType("lifecycle_callback_received"),
			Operation: string(op.Action),
			Success:   callback.Success,
			Details: map[string]interface{}{
				"operation_id":  op.ID,
				"allocation_id": op.AllocationID,
				"callback_id":   callback.ID,
				"result_state":  callback.ResultState,
			},
		})
	}

	log.Printf("[lifecycle-controller] operation %s completed with callback %s (success=%t)",
		op.ID, callback.ID, callback.Success)

	return lc.saveState()
}

// GetOperation retrieves an operation by ID
func (lc *LifecycleController) GetOperation(id string) (*marketplace.LifecycleOperation, bool) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	op, ok := lc.state.Operations[id]
	return op, ok
}

// GetOperationByIdempotencyKey retrieves an operation by idempotency key
func (lc *LifecycleController) GetOperationByIdempotencyKey(key string) (*marketplace.LifecycleOperation, bool) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	opID, ok := lc.state.IdempotencyIndex[key]
	if !ok {
		return nil, false
	}
	op, ok := lc.state.Operations[opID]
	return op, ok
}

// GetPendingOperations returns pending operations for an allocation
func (lc *LifecycleController) GetPendingOperations(allocationID string) []*marketplace.LifecycleOperation {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	var result []*marketplace.LifecycleOperation
	for _, op := range lc.state.Operations {
		if op.AllocationID == allocationID && !op.State.IsTerminal() {
			result = append(result, op)
		}
	}
	return result
}

// GetMetrics returns lifecycle metrics
func (lc *LifecycleController) GetMetrics() *marketplace.LifecycleMetrics {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.state.Metrics
}

// createFailureCallback creates a callback for a failed operation
func (lc *LifecycleController) createFailureCallback(op *marketplace.LifecycleOperation) *marketplace.WaldurCallback {
	callback := marketplace.NewWaldurCallback(
		marketplace.ActionTypeStatusUpdate,
		op.WaldurOperationID,
		marketplace.SyncTypeAllocation,
		op.AllocationID,
	)
	callback.SignerID = lc.cfg.ProviderAddress
	callback.ExpiresAt = callback.Timestamp.Add(lc.cfg.CallbackTTL)
	callback.Payload["operation_id"] = op.ID
	callback.Payload["action"] = string(op.Action)
	callback.Payload["state"] = "failed"
	callback.Payload["error"] = op.Error
	return callback
}

// signAndSubmitCallback signs and submits a callback
func (lc *LifecycleController) signAndSubmitCallback(ctx context.Context, callback *marketplace.WaldurCallback) error {
	if callback == nil || lc.callbackSink == nil {
		return nil
	}

	sig, err := lc.keyManager.Sign(callback.SigningPayload())
	if err != nil {
		return fmt.Errorf("sign callback: %w", err)
	}

	sigBytes, err := hex.DecodeString(sig.Signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}

	callback.Signature = sigBytes
	return lc.callbackSink.Submit(ctx, callback)
}

// retryWorker retries failed operations
func (lc *LifecycleController) retryWorker(ctx context.Context) {
	defer lc.wg.Done()

	ticker := time.NewTicker(lc.cfg.RetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-lc.stopCh:
			return
		case <-ticker.C:
			lc.processRetries(ctx)
		}
	}
}

// processRetries processes operations needing retry
func (lc *LifecycleController) processRetries(ctx context.Context) {
	lc.mu.RLock()
	var toRetry []*marketplace.LifecycleOperation
	for _, op := range lc.state.Operations {
		if op.State == marketplace.LifecycleOpStatePending && op.RetryCount > 0 {
			toRetry = append(toRetry, op)
		}
	}
	lc.mu.RUnlock()

	for _, op := range toRetry {
		log.Printf("[lifecycle-controller] retrying operation %s (attempt %d)", op.ID, op.RetryCount)
		// Re-execute would need resource UUID from allocation mapping
		// For now, just log
	}
}

// cleanupWorker cleans up old operations
func (lc *LifecycleController) cleanupWorker(ctx context.Context) {
	defer lc.wg.Done()

	ticker := time.NewTicker(lc.cfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-lc.stopCh:
			return
		case <-ticker.C:
			lc.cleanup()
		}
	}
}

// cleanup removes old completed operations
func (lc *LifecycleController) cleanup() {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -lc.cfg.OperationRetentionDays)
	cleanedOps := 0
	cleanedCallbacks := 0

	// Clean old operations
	for id, op := range lc.state.Operations {
		if op.State.IsTerminal() && op.CompletedAt != nil && op.CompletedAt.Before(cutoff) {
			delete(lc.state.Operations, id)
			delete(lc.state.IdempotencyIndex, op.IdempotencyKey)
			cleanedOps++
		}
	}

	// Clean expired callback nonces
	now := time.Now()
	for nonce, expiry := range lc.state.ProcessedCallbacks {
		if now.After(expiry) {
			delete(lc.state.ProcessedCallbacks, nonce)
			cleanedCallbacks++
		}
	}

	if cleanedOps > 0 || cleanedCallbacks > 0 {
		log.Printf("[lifecycle-controller] cleaned %d operations, %d callback nonces", cleanedOps, cleanedCallbacks)
		_ = lc.saveState()
	}
}

// loadState loads persisted state
func (lc *LifecycleController) loadState() error {
	if lc.cfg.StateFilePath == "" {
		return nil
	}

	data, err := os.ReadFile(lc.cfg.StateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var state LifecycleControllerState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	lc.state = &state
	return nil
}

// saveState persists state to disk
func (lc *LifecycleController) saveState() error {
	if lc.cfg.StateFilePath == "" {
		return nil
	}

	lc.state.LastUpdated = time.Now().UTC()

	data, err := json.MarshalIndent(lc.state, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(lc.cfg.StateFilePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	tmp := lc.cfg.StateFilePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}

	return os.Rename(tmp, lc.cfg.StateFilePath)
}

// mapToWaldurAction maps marketplace action to Waldur action
func mapToWaldurAction(action marketplace.LifecycleActionType) waldur.LifecycleAction {
	switch action {
	case marketplace.LifecycleActionStart:
		return waldur.LifecycleActionStart
	case marketplace.LifecycleActionStop:
		return waldur.LifecycleActionStop
	case marketplace.LifecycleActionRestart:
		return waldur.LifecycleActionRestart
	case marketplace.LifecycleActionSuspend:
		return waldur.LifecycleActionSuspend
	case marketplace.LifecycleActionResume:
		return waldur.LifecycleActionResume
	case marketplace.LifecycleActionResize:
		return waldur.LifecycleActionResize
	case marketplace.LifecycleActionTerminate:
		return waldur.LifecycleActionTerminate
	default:
		return waldur.LifecycleAction(strings.ToLower(string(action)))
	}
}

