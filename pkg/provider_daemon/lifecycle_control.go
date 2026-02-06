// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-14E: Resource lifecycle control via Waldur
// This file implements enhanced lifecycle control for start/stop/resize/terminate
// operations with audit trail and integration with the LifecycleController.
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

// Lifecycle control errors
var (
	// ErrResourceNotFound is returned when a resource is not found
	ErrResourceNotFound = errors.New("resource not found")

	// ErrInvalidResourceState is returned for invalid resource states
	ErrInvalidResourceState = errors.New("invalid resource state for operation")

	// ErrResizeNotSupported is returned when resize is not supported
	ErrResizeNotSupported = errors.New("resize not supported for this resource type")

	// ErrTerminateInProgress is returned when termination is already in progress
	ErrTerminateInProgress = errors.New("termination already in progress")

	// ErrCleanupFailed is returned when resource cleanup fails
	ErrCleanupFailed = errors.New("resource cleanup failed")

	// ErrResourceBusy is returned when resource is busy with another operation
	ErrResourceBusy = errors.New("resource is busy with another operation")
)

const paramValueTrue = "true"

// ResourceLifecycleConfig configures resource lifecycle operations
type ResourceLifecycleConfig struct {
	// StartTimeout is the timeout for start operations
	StartTimeout time.Duration `json:"start_timeout"`

	// StopTimeout is the timeout for stop operations
	StopTimeout time.Duration `json:"stop_timeout"`

	// ResizeTimeout is the timeout for resize operations
	ResizeTimeout time.Duration `json:"resize_timeout"`

	// TerminateTimeout is the timeout for terminate operations
	TerminateTimeout time.Duration `json:"terminate_timeout"`

	// CleanupRetries is the number of cleanup retries on failure
	CleanupRetries int `json:"cleanup_retries"`

	// CleanupRetryInterval is the interval between cleanup retries
	CleanupRetryInterval time.Duration `json:"cleanup_retry_interval"`

	// ForceTerminateAfter is when to force terminate after graceful timeout
	ForceTerminateAfter time.Duration `json:"force_terminate_after"`

	// EnablePreflightChecks enables preflight checks before operations
	EnablePreflightChecks bool `json:"enable_preflight_checks"`

	// EnablePostflightValidation enables validation after operations
	EnablePostflightValidation bool `json:"enable_postflight_validation"`
}

// DefaultResourceLifecycleConfig returns default configuration
func DefaultResourceLifecycleConfig() ResourceLifecycleConfig {
	return ResourceLifecycleConfig{
		StartTimeout:               5 * time.Minute,
		StopTimeout:                5 * time.Minute,
		ResizeTimeout:              10 * time.Minute,
		TerminateTimeout:           10 * time.Minute,
		CleanupRetries:             3,
		CleanupRetryInterval:       30 * time.Second,
		ForceTerminateAfter:        15 * time.Minute,
		EnablePreflightChecks:      true,
		EnablePostflightValidation: true,
	}
}

// ResourceInfo contains information about a managed resource
type ResourceInfo struct {
	// AllocationID is the VirtEngine allocation ID
	AllocationID string `json:"allocation_id"`

	// WaldurResourceUUID is the Waldur resource UUID
	WaldurResourceUUID string `json:"waldur_resource_uuid"`

	// ResourceType is the type of resource (vm, container, hpc_allocation, etc.)
	ResourceType string `json:"resource_type"`

	// CurrentState is the current resource state
	CurrentState marketplace.AllocationState `json:"current_state"`

	// WaldurState is the Waldur-side resource state
	WaldurState waldur.ResourceState `json:"waldur_state"`

	// ProviderAddress is the provider managing this resource
	ProviderAddress string `json:"provider_address"`

	// LastUpdated is when the resource info was last updated
	LastUpdated time.Time `json:"last_updated"`

	// Metadata contains additional resource metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// LifecycleActionRequest represents a lifecycle action request
type LifecycleActionRequest struct {
	// AllocationID is the allocation to operate on
	AllocationID string

	// Action is the lifecycle action to perform
	Action marketplace.LifecycleActionType

	// ResourceInfo is the resource information
	ResourceInfo *ResourceInfo

	// ResizeSpec contains resize specifications (for resize action)
	ResizeSpec *marketplace.ResizeSpecification

	// RequestedBy is who requested the action
	RequestedBy string

	// Immediate forces immediate execution (for stop/terminate)
	Immediate bool

	// Parameters contains action-specific parameters
	Parameters map[string]string

	// Timeout is the operation timeout (overrides default)
	Timeout time.Duration
}

// LifecycleActionResult contains the result of a lifecycle action
type LifecycleActionResult struct {
	// OperationID is the operation identifier
	OperationID string

	// Success indicates if the action was initiated successfully
	Success bool

	// State is the operation state
	State marketplace.LifecycleOperationState

	// NewResourceState is the new resource state (if completed)
	NewResourceState marketplace.AllocationState

	// WaldurOperationID is the Waldur-side operation ID
	WaldurOperationID string

	// Message contains any additional message
	Message string

	// Error contains any error message
	Error string

	// StartedAt is when the operation started
	StartedAt time.Time

	// CompletedAt is when the operation completed (if completed)
	CompletedAt *time.Time
}

// ResourceLifecycleManager manages resource lifecycle operations
type ResourceLifecycleManager struct {
	cfg         ResourceLifecycleConfig
	controller  *LifecycleController
	lifecycle   *waldur.LifecycleClient
	auditLogger *AuditLogger
	resources   map[string]*ResourceInfo
	activeOps   map[string]string // allocationID -> operationID
	rollbackMgr *RollbackManager
	mu          sync.RWMutex
}

// NewResourceLifecycleManager creates a new lifecycle manager
func NewResourceLifecycleManager(
	cfg ResourceLifecycleConfig,
	controller *LifecycleController,
	lifecycle *waldur.LifecycleClient,
	auditLogger *AuditLogger,
) *ResourceLifecycleManager {
	return &ResourceLifecycleManager{
		cfg:         cfg,
		controller:  controller,
		lifecycle:   lifecycle,
		auditLogger: auditLogger,
		resources:   make(map[string]*ResourceInfo),
		activeOps:   make(map[string]string),
	}
}

// SetRollbackManager sets the rollback manager
func (m *ResourceLifecycleManager) SetRollbackManager(rm *RollbackManager) {
	m.rollbackMgr = rm
}

// RegisterResource registers a resource for lifecycle management
func (m *ResourceLifecycleManager) RegisterResource(info *ResourceInfo) error {
	if info == nil {
		return errors.New("resource info is nil")
	}
	if info.AllocationID == "" {
		return errors.New("allocation ID is required")
	}
	if info.WaldurResourceUUID == "" {
		return errors.New("waldur resource UUID is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	info.LastUpdated = time.Now().UTC()
	m.resources[info.AllocationID] = info

	log.Printf("[lifecycle-manager] registered resource %s (Waldur: %s)",
		info.AllocationID, info.WaldurResourceUUID)

	return nil
}

// UnregisterResource removes a resource from lifecycle management
func (m *ResourceLifecycleManager) UnregisterResource(allocationID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.resources, allocationID)
	delete(m.activeOps, allocationID)
}

// GetResource retrieves resource information
func (m *ResourceLifecycleManager) GetResource(allocationID string) (*ResourceInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	info, ok := m.resources[allocationID]
	return info, ok
}

// Start starts a stopped resource
func (m *ResourceLifecycleManager) Start(ctx context.Context, req *LifecycleActionRequest) (*LifecycleActionResult, error) {
	req.Action = marketplace.LifecycleActionStart
	return m.executeAction(ctx, req)
}

// Stop stops a running resource
func (m *ResourceLifecycleManager) Stop(ctx context.Context, req *LifecycleActionRequest) (*LifecycleActionResult, error) {
	req.Action = marketplace.LifecycleActionStop
	return m.executeAction(ctx, req)
}

// Restart restarts a resource
func (m *ResourceLifecycleManager) Restart(ctx context.Context, req *LifecycleActionRequest) (*LifecycleActionResult, error) {
	req.Action = marketplace.LifecycleActionRestart
	return m.executeAction(ctx, req)
}

// Suspend suspends a resource (preserves state)
func (m *ResourceLifecycleManager) Suspend(ctx context.Context, req *LifecycleActionRequest) (*LifecycleActionResult, error) {
	req.Action = marketplace.LifecycleActionSuspend
	return m.executeAction(ctx, req)
}

// Resume resumes a suspended resource
func (m *ResourceLifecycleManager) Resume(ctx context.Context, req *LifecycleActionRequest) (*LifecycleActionResult, error) {
	req.Action = marketplace.LifecycleActionResume
	return m.executeAction(ctx, req)
}

// Resize resizes a resource
func (m *ResourceLifecycleManager) Resize(ctx context.Context, req *LifecycleActionRequest) (*LifecycleActionResult, error) {
	if req.ResizeSpec == nil {
		return nil, errors.New("resize specification is required")
	}
	req.Action = marketplace.LifecycleActionResize
	return m.executeAction(ctx, req)
}

// Terminate terminates a resource with cleanup
func (m *ResourceLifecycleManager) Terminate(ctx context.Context, req *LifecycleActionRequest) (*LifecycleActionResult, error) {
	req.Action = marketplace.LifecycleActionTerminate
	return m.executeAction(ctx, req)
}

// executeAction executes a lifecycle action
func (m *ResourceLifecycleManager) executeAction(ctx context.Context, req *LifecycleActionRequest) (*LifecycleActionResult, error) {
	if req.AllocationID == "" {
		return nil, errors.New("allocation ID is required")
	}

	// Get resource info
	m.mu.RLock()
	info, exists := m.resources[req.AllocationID]
	activeOpID := m.activeOps[req.AllocationID]
	m.mu.RUnlock()

	if !exists {
		if req.ResourceInfo != nil {
			info = req.ResourceInfo
		} else {
			return nil, ErrResourceNotFound
		}
	}

	// Check for active operation
	if activeOpID != "" {
		if m.controller != nil {
			if op, found := m.controller.GetOperation(activeOpID); found {
				if op.AllocationID == req.AllocationID && op.Action == req.Action {
					return &LifecycleActionResult{
						OperationID:       op.ID,
						Success:           true,
						State:             op.State,
						NewResourceState:  info.CurrentState,
						WaldurOperationID: op.WaldurOperationID,
						Message:           "idempotent request: returning existing operation",
						StartedAt:         op.CreatedAt,
					}, nil
				}
			}
		}
		return nil, ErrResourceBusy
	}

	// Preflight checks
	if m.cfg.EnablePreflightChecks {
		if err := m.preflightCheck(ctx, info, req.Action); err != nil {
			return nil, fmt.Errorf("preflight check failed: %w", err)
		}
	}

	// Get timeout for this action
	timeout := m.getTimeout(req)

	// Create parameters map
	params := make(map[string]string)
	for k, v := range req.Parameters {
		params[k] = v
	}

	// Add resize parameters
	if req.ResizeSpec != nil {
		m.addResizeParams(params, req.ResizeSpec)
	}

	// Add immediate flag
	if req.Immediate {
		params["immediate"] = paramValueTrue
	}

	// Execute via lifecycle controller
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	op, err := m.controller.ExecuteLifecycleAction(
		ctxWithTimeout,
		req.AllocationID,
		req.Action,
		info.CurrentState,
		info.WaldurResourceUUID,
		req.RequestedBy,
		params,
	)
	if err != nil {
		m.logAuditEvent(req, nil, err)
		return nil, err
	}

	// Track active operation
	m.mu.Lock()
	m.activeOps[req.AllocationID] = op.ID
	m.mu.Unlock()

	result := &LifecycleActionResult{
		OperationID:       op.ID,
		Success:           true,
		State:             op.State,
		WaldurOperationID: op.WaldurOperationID,
		StartedAt:         time.Now().UTC(),
	}

	// Log audit event
	m.logAuditEvent(req, result, nil)

	log.Printf("[lifecycle-manager] %s action initiated for %s (op: %s)",
		req.Action, req.AllocationID, op.ID)

	return result, nil
}

// preflightCheck performs preflight checks before an action
func (m *ResourceLifecycleManager) preflightCheck(ctx context.Context, info *ResourceInfo, action marketplace.LifecycleActionType) error {
	// Validate state transition is allowed
	_, err := marketplace.ValidateLifecycleTransition(info.CurrentState, action)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidResourceState, err)
	}

	// Check Waldur resource state if lifecycle client is available
	if m.lifecycle != nil && info.WaldurResourceUUID != "" {
		waldurState, err := m.lifecycle.GetResourceState(ctx, info.WaldurResourceUUID)
		if err != nil {
			log.Printf("[lifecycle-manager] warning: could not get Waldur state: %v", err)
		} else {
			if err := waldur.ValidateLifecycleAction(waldurState, m.mapToWaldurAction(action)); err != nil {
				return fmt.Errorf("%w: %v", ErrInvalidResourceState, err)
			}
		}
	}

	return nil
}

// getTimeout returns the timeout for an action
func (m *ResourceLifecycleManager) getTimeout(req *LifecycleActionRequest) time.Duration {
	if req.Timeout > 0 {
		return req.Timeout
	}

	switch req.Action {
	case marketplace.LifecycleActionStart:
		return m.cfg.StartTimeout
	case marketplace.LifecycleActionStop:
		return m.cfg.StopTimeout
	case marketplace.LifecycleActionResize:
		return m.cfg.ResizeTimeout
	case marketplace.LifecycleActionTerminate:
		return m.cfg.TerminateTimeout
	default:
		return m.cfg.StartTimeout
	}
}

// addResizeParams adds resize specification to parameters
func (m *ResourceLifecycleManager) addResizeParams(params map[string]string, spec *marketplace.ResizeSpecification) {
	if spec.CPUCores != nil {
		params["cpu_cores"] = fmt.Sprintf("%d", *spec.CPUCores)
	}
	if spec.MemoryMB != nil {
		params["memory_mb"] = fmt.Sprintf("%d", *spec.MemoryMB)
	}
	if spec.StorageGB != nil {
		params["storage_gb"] = fmt.Sprintf("%d", *spec.StorageGB)
	}
	if spec.GPUCount != nil {
		params["gpu_count"] = fmt.Sprintf("%d", *spec.GPUCount)
	}
	for k, v := range spec.CustomLimits {
		params[fmt.Sprintf("custom_%s", k)] = fmt.Sprintf("%d", v)
	}
}

// mapToWaldurAction maps marketplace action to Waldur action
func (m *ResourceLifecycleManager) mapToWaldurAction(action marketplace.LifecycleActionType) waldur.LifecycleAction {
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
		return waldur.LifecycleAction(action)
	}
}

// HandleOperationComplete handles operation completion
func (m *ResourceLifecycleManager) HandleOperationComplete(
	allocationID string,
	operationID string,
	success bool,
	newState marketplace.AllocationState,
	errMsg string,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear active operation
	if m.activeOps[allocationID] == operationID {
		delete(m.activeOps, allocationID)
	}

	// Update resource state
	if info, ok := m.resources[allocationID]; ok {
		if success {
			info.CurrentState = newState
		}
		info.LastUpdated = time.Now().UTC()
	}

	// Handle failure with rollback if configured
	if !success && m.rollbackMgr != nil {
		op, found := m.controller.GetOperation(operationID)
		if found && op.RollbackPolicy == marketplace.RollbackPolicyAutomatic {
			// Schedule rollback
			go func() {
				ctx := context.Background()
				if err := m.rollbackMgr.RollbackOperation(ctx, op); err != nil {
					log.Printf("[lifecycle-manager] rollback failed for operation %s: %v", operationID, err)
				}
			}()
		}
	}

	log.Printf("[lifecycle-manager] operation %s completed (success=%t, newState=%s)",
		operationID, success, newState)
}

// TerminateWithCleanup terminates a resource with full cleanup
func (m *ResourceLifecycleManager) TerminateWithCleanup(ctx context.Context, req *LifecycleActionRequest) (*LifecycleActionResult, error) {
	req.Action = marketplace.LifecycleActionTerminate

	// Execute terminate action
	result, err := m.executeAction(ctx, req)
	if err != nil {
		return nil, err
	}

	// Schedule cleanup tasks asynchronously
	go m.performCleanup(ctx, req.AllocationID, result.OperationID)

	return result, nil
}

// performCleanup performs post-termination cleanup
func (m *ResourceLifecycleManager) performCleanup(ctx context.Context, allocationID, operationID string) {
	// Wait for termination to complete
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeout := time.After(m.cfg.TerminateTimeout)

	for {
		select {
		case <-ctx.Done():
			return
		case <-timeout:
			log.Printf("[lifecycle-manager] cleanup timeout for %s", allocationID)
			return
		case <-ticker.C:
			op, found := m.controller.GetOperation(operationID)
			if !found {
				return
			}
			if op.State.IsTerminal() {
				// Perform cleanup tasks
				m.executeCleanupTasks(ctx, allocationID, op.State == marketplace.LifecycleOpStateCompleted)
				return
			}
		}
	}
}

// executeCleanupTasks executes cleanup tasks after termination
func (m *ResourceLifecycleManager) executeCleanupTasks(_ context.Context, allocationID string, terminated bool) {
	if !terminated {
		log.Printf("[lifecycle-manager] skipping cleanup for %s (termination failed)", allocationID)
		return
	}

	log.Printf("[lifecycle-manager] performing cleanup for %s", allocationID)

	// Unregister resource
	m.UnregisterResource(allocationID)

	// Additional cleanup tasks would go here:
	// - Release IP addresses
	// - Delete persistent volumes
	// - Remove DNS records
	// - Archive logs

	log.Printf("[lifecycle-manager] cleanup completed for %s", allocationID)
}

// logAuditEvent logs a lifecycle audit event
func (m *ResourceLifecycleManager) logAuditEvent(req *LifecycleActionRequest, result *LifecycleActionResult, err error) {
	if m.auditLogger == nil {
		return
	}

	eventType := AuditEventType("lifecycle_action_initiated")
	if err != nil {
		eventType = AuditEventType("lifecycle_action_failed")
	}

	details := map[string]interface{}{
		"allocation_id": req.AllocationID,
		"action":        req.Action,
		"requested_by":  req.RequestedBy,
	}

	if result != nil {
		details["operation_id"] = result.OperationID
		details["waldur_operation_id"] = result.WaldurOperationID
	}

	if req.ResizeSpec != nil {
		details["resize_spec"] = req.ResizeSpec
	}

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	_ = m.auditLogger.Log(&AuditEvent{
		Type:         eventType,
		Operation:    string(req.Action),
		Success:      err == nil,
		ErrorMessage: errMsg,
		Details:      details,
	})
}

// GetActiveOperations returns active operations
func (m *ResourceLifecycleManager) GetActiveOperations() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range m.activeOps {
		result[k] = v
	}
	return result
}

// GetManagedResources returns managed resources
func (m *ResourceLifecycleManager) GetManagedResources() []*ResourceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*ResourceInfo, 0, len(m.resources))
	for _, info := range m.resources {
		result = append(result, info)
	}
	return result
}

// UpdateResourceState updates a resource's state
func (m *ResourceLifecycleManager) UpdateResourceState(allocationID string, state marketplace.AllocationState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, ok := m.resources[allocationID]
	if !ok {
		return ErrResourceNotFound
	}

	info.CurrentState = state
	info.LastUpdated = time.Now().UTC()
	return nil
}
