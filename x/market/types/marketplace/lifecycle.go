// Package marketplace provides types for the marketplace on-chain module.
//
// VE-4E: Resource lifecycle control via Waldur
// This file defines lifecycle action types, state transitions, and callback validation.
package marketplace

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// Lifecycle-related errors
var (
	// ErrInvalidLifecycleAction is returned for invalid lifecycle actions
	ErrInvalidLifecycleAction = errors.New("invalid lifecycle action")

	// ErrLifecycleCallbackExpired is returned when a callback has expired
	ErrLifecycleCallbackExpired = errors.New("lifecycle callback expired")

	// ErrLifecycleCallbackInvalidSignature is returned for invalid signatures
	ErrLifecycleCallbackInvalidSignature = errors.New("invalid callback signature")

	// ErrLifecycleIdempotencyConflict is returned for idempotency key conflicts
	ErrLifecycleIdempotencyConflict = errors.New("idempotency key conflict")

	// ErrLifecycleOperationInProgress is returned when an operation is already in progress
	ErrLifecycleOperationInProgress = errors.New("lifecycle operation already in progress")

	// ErrLifecycleRollbackFailed is returned when rollback fails
	ErrLifecycleRollbackFailed = errors.New("lifecycle rollback failed")
)

// LifecycleActionType represents types of lifecycle actions
type LifecycleActionType string

const (
	// LifecycleActionStart starts a stopped resource
	LifecycleActionStart LifecycleActionType = "start"

	// LifecycleActionStop stops a running resource
	LifecycleActionStop LifecycleActionType = "stop"

	// LifecycleActionRestart restarts a resource
	LifecycleActionRestart LifecycleActionType = "restart"

	// LifecycleActionSuspend suspends a resource (preserving state)
	LifecycleActionSuspend LifecycleActionType = "suspend"

	// LifecycleActionResume resumes a suspended resource
	LifecycleActionResume LifecycleActionType = "resume"

	// LifecycleActionResize resizes resource allocation
	LifecycleActionResize LifecycleActionType = "resize"

	// LifecycleActionTerminate terminates a resource permanently
	LifecycleActionTerminate LifecycleActionType = "terminate"

	// LifecycleActionProvision provisions a new resource
	LifecycleActionProvision LifecycleActionType = "provision"
)

// AllLifecycleActionTypes returns all lifecycle action types
func AllLifecycleActionTypes() []LifecycleActionType {
	return []LifecycleActionType{
		LifecycleActionStart,
		LifecycleActionStop,
		LifecycleActionRestart,
		LifecycleActionSuspend,
		LifecycleActionResume,
		LifecycleActionResize,
		LifecycleActionTerminate,
		LifecycleActionProvision,
	}
}

// IsValid checks if the action type is valid
func (a LifecycleActionType) IsValid() bool {
	switch a {
	case LifecycleActionStart, LifecycleActionStop, LifecycleActionRestart,
		LifecycleActionSuspend, LifecycleActionResume, LifecycleActionResize,
		LifecycleActionTerminate, LifecycleActionProvision:
		return true
	default:
		return false
	}
}

// String returns string representation
func (a LifecycleActionType) String() string {
	return string(a)
}

// LifecycleOperationState represents the state of a lifecycle operation
type LifecycleOperationState string

const (
	// LifecycleOpStatePending operation is pending execution
	LifecycleOpStatePending LifecycleOperationState = "pending"

	// LifecycleOpStateExecuting operation is being executed
	LifecycleOpStateExecuting LifecycleOperationState = "executing"

	// LifecycleOpStateAwaitingCallback waiting for signed callback
	LifecycleOpStateAwaitingCallback LifecycleOperationState = "awaiting_callback"

	// LifecycleOpStateCompleted operation completed successfully
	LifecycleOpStateCompleted LifecycleOperationState = "completed"

	// LifecycleOpStateFailed operation failed
	LifecycleOpStateFailed LifecycleOperationState = "failed"

	// LifecycleOpStateRolledBack operation was rolled back
	LifecycleOpStateRolledBack LifecycleOperationState = "rolled_back"

	// LifecycleOpStateCancelled operation was cancelled
	LifecycleOpStateCancelled LifecycleOperationState = "cancelled"
)

// IsTerminal returns true if state is terminal
func (s LifecycleOperationState) IsTerminal() bool {
	return s == LifecycleOpStateCompleted ||
		s == LifecycleOpStateFailed ||
		s == LifecycleOpStateRolledBack ||
		s == LifecycleOpStateCancelled
}

// AllocationLifecycleTransitions defines valid state transitions for allocations based on lifecycle actions
var AllocationLifecycleTransitions = map[AllocationState]map[LifecycleActionType]AllocationState{
	AllocationStatePending: {
		LifecycleActionProvision: AllocationStateProvisioning,
		LifecycleActionTerminate: AllocationStateTerminated,
	},
	AllocationStateAccepted: {
		LifecycleActionProvision: AllocationStateProvisioning,
		LifecycleActionTerminate: AllocationStateTerminated,
	},
	AllocationStateProvisioning: {
		LifecycleActionTerminate: AllocationStateTerminating,
	},
	AllocationStateActive: {
		LifecycleActionStop:      AllocationStateSuspended,
		LifecycleActionSuspend:   AllocationStateSuspended,
		LifecycleActionRestart:   AllocationStateActive, // Stays active after restart
		LifecycleActionResize:    AllocationStateActive, // Stays active after resize
		LifecycleActionTerminate: AllocationStateTerminating,
	},
	AllocationStateSuspended: {
		LifecycleActionStart:     AllocationStateActive,
		LifecycleActionResume:    AllocationStateActive,
		LifecycleActionTerminate: AllocationStateTerminating,
	},
	AllocationStateTerminating: {
		// Terminal state, only terminate can complete
	},
}

// ValidateLifecycleTransition validates if a lifecycle action is valid for current state
func ValidateLifecycleTransition(currentState AllocationState, action LifecycleActionType) (AllocationState, error) {
	transitions, ok := AllocationLifecycleTransitions[currentState]
	if !ok {
		return currentState, fmt.Errorf("%w: no transitions from state %s", ErrInvalidStateTransition, currentState)
	}

	newState, valid := transitions[action]
	if !valid {
		return currentState, fmt.Errorf("%w: action %s not allowed in state %s", ErrInvalidStateTransition, action, currentState)
	}

	return newState, nil
}

// RollbackPolicy defines how to handle failures during lifecycle operations
type RollbackPolicy string

const (
	// RollbackPolicyNone no automatic rollback
	RollbackPolicyNone RollbackPolicy = "none"

	// RollbackPolicyAutomatic automatically rollback on failure
	RollbackPolicyAutomatic RollbackPolicy = "automatic"

	// RollbackPolicyManual require manual intervention for rollback
	RollbackPolicyManual RollbackPolicy = "manual"

	// RollbackPolicyRetry retry the operation before rolling back
	RollbackPolicyRetry RollbackPolicy = "retry"
)

// LifecycleOperation represents a lifecycle operation request
type LifecycleOperation struct {
	// ID is the unique operation identifier
	ID string `json:"id"`

	// IdempotencyKey is used to prevent duplicate operations
	IdempotencyKey string `json:"idempotency_key"`

	// AllocationID is the allocation being operated on
	AllocationID string `json:"allocation_id"`

	// Action is the lifecycle action type
	Action LifecycleActionType `json:"action"`

	// State is the current operation state
	State LifecycleOperationState `json:"state"`

	// PreviousAllocationState is the allocation state before operation
	PreviousAllocationState AllocationState `json:"previous_allocation_state"`

	// TargetAllocationState is the expected allocation state after operation
	TargetAllocationState AllocationState `json:"target_allocation_state"`

	// RequestedBy is who requested the operation
	RequestedBy string `json:"requested_by"`

	// ProviderAddress is the provider handling the operation
	ProviderAddress string `json:"provider_address"`

	// Parameters contains action-specific parameters
	Parameters map[string]string `json:"parameters,omitempty"`

	// ResizeSpec contains resize specifications (if action is resize)
	ResizeSpec *ResizeSpecification `json:"resize_spec,omitempty"`

	// RollbackPolicy defines the rollback behavior
	RollbackPolicy RollbackPolicy `json:"rollback_policy"`

	// MaxRetries is the maximum number of retries
	MaxRetries uint32 `json:"max_retries"`

	// RetryCount is the current retry count
	RetryCount uint32 `json:"retry_count"`

	// CreatedAt is when the operation was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the operation was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// StartedAt is when execution started
	StartedAt *time.Time `json:"started_at,omitempty"`

	// CompletedAt is when the operation completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// ExpiresAt is when the operation expires
	ExpiresAt time.Time `json:"expires_at"`

	// CallbackID is the expected callback ID
	CallbackID string `json:"callback_id,omitempty"`

	// Error contains any error message
	Error string `json:"error,omitempty"`

	// WaldurOperationID is the Waldur-side operation ID
	WaldurOperationID string `json:"waldur_operation_id,omitempty"`
}

// ResizeSpecification contains resource resize parameters
type ResizeSpecification struct {
	// CPUCores is the new CPU core count
	CPUCores *uint32 `json:"cpu_cores,omitempty"`

	// MemoryMB is the new memory in MB
	MemoryMB *uint64 `json:"memory_mb,omitempty"`

	// StorageGB is the new storage in GB
	StorageGB *uint64 `json:"storage_gb,omitempty"`

	// GPUCount is the new GPU count
	GPUCount *uint32 `json:"gpu_count,omitempty"`

	// CustomLimits contains custom resource limits
	CustomLimits map[string]int64 `json:"custom_limits,omitempty"`
}

// NewLifecycleOperation creates a new lifecycle operation
func NewLifecycleOperation(
	allocationID string,
	action LifecycleActionType,
	requestedBy string,
	providerAddress string,
	currentState AllocationState,
) (*LifecycleOperation, error) {
	return NewLifecycleOperationAt(allocationID, action, requestedBy, providerAddress, currentState, time.Now())
}

// NewLifecycleOperationAt creates a new lifecycle operation at a specific time
func NewLifecycleOperationAt(
	allocationID string,
	action LifecycleActionType,
	requestedBy string,
	providerAddress string,
	currentState AllocationState,
	now time.Time,
) (*LifecycleOperation, error) {
	if !action.IsValid() {
		return nil, fmt.Errorf("%w: %s", ErrInvalidLifecycleAction, action)
	}

	targetState, err := ValidateLifecycleTransition(currentState, action)
	if err != nil {
		return nil, err
	}

	timestamp := now.UTC()
	idempotencyKey := GenerateIdempotencyKey(allocationID, action, timestamp)
	operationID := GenerateOperationID(allocationID, action, timestamp)

	return &LifecycleOperation{
		ID:                      operationID,
		IdempotencyKey:          idempotencyKey,
		AllocationID:            allocationID,
		Action:                  action,
		State:                   LifecycleOpStatePending,
		PreviousAllocationState: currentState,
		TargetAllocationState:   targetState,
		RequestedBy:             requestedBy,
		ProviderAddress:         providerAddress,
		Parameters:              make(map[string]string),
		RollbackPolicy:          RollbackPolicyAutomatic,
		MaxRetries:              3,
		CreatedAt:               timestamp,
		UpdatedAt:               timestamp,
		ExpiresAt:               timestamp.Add(time.Hour),
	}, nil
}

// GenerateIdempotencyKey generates an idempotency key
func GenerateIdempotencyKey(allocationID string, action LifecycleActionType, timestamp time.Time) string {
	// Include hour precision for idempotency (allow retry within same hour window)
	data := fmt.Sprintf("%s:%s:%d", allocationID, action, timestamp.Truncate(time.Hour).Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

// GenerateOperationID generates a unique operation ID
func GenerateOperationID(allocationID string, action LifecycleActionType, timestamp time.Time) string {
	data := fmt.Sprintf("%s:%s:%d", allocationID, action, timestamp.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("lco_%s", hex.EncodeToString(hash[:12]))
}

// SetExecuting transitions operation to executing state
func (op *LifecycleOperation) SetExecuting(waldurOpID string, now time.Time) error {
	if op.State != LifecycleOpStatePending {
		return fmt.Errorf("cannot execute from state %s", op.State)
	}
	timestamp := now.UTC()
	op.State = LifecycleOpStateExecuting
	op.StartedAt = &timestamp
	op.UpdatedAt = timestamp
	op.WaldurOperationID = waldurOpID
	return nil
}

// SetAwaitingCallback transitions operation to awaiting callback state
func (op *LifecycleOperation) SetAwaitingCallback(callbackID string, now time.Time) error {
	if op.State != LifecycleOpStateExecuting {
		return fmt.Errorf("cannot await callback from state %s", op.State)
	}
	op.State = LifecycleOpStateAwaitingCallback
	op.CallbackID = callbackID
	op.UpdatedAt = now.UTC()
	return nil
}

// SetCompleted marks the operation as completed
func (op *LifecycleOperation) SetCompleted(now time.Time) {
	timestamp := now.UTC()
	op.State = LifecycleOpStateCompleted
	op.CompletedAt = &timestamp
	op.UpdatedAt = timestamp
}

// SetFailed marks the operation as failed
func (op *LifecycleOperation) SetFailed(errMsg string, now time.Time) {
	timestamp := now.UTC()
	op.State = LifecycleOpStateFailed
	op.Error = errMsg
	op.CompletedAt = &timestamp
	op.UpdatedAt = timestamp
}

// SetRolledBack marks the operation as rolled back
func (op *LifecycleOperation) SetRolledBack(now time.Time) {
	timestamp := now.UTC()
	op.State = LifecycleOpStateRolledBack
	op.CompletedAt = &timestamp
	op.UpdatedAt = timestamp
}

// IncrementRetry increments the retry count
func (op *LifecycleOperation) IncrementRetry() bool {
	op.RetryCount++
	return op.RetryCount <= op.MaxRetries
}

// IsExpired checks if the operation has expired
func (op *LifecycleOperation) IsExpired(now time.Time) bool {
	return now.After(op.ExpiresAt)
}

// ShouldRetry returns true if the operation should be retried
func (op *LifecycleOperation) ShouldRetry() bool {
	if op.RollbackPolicy != RollbackPolicyRetry {
		return false
	}
	return op.RetryCount < op.MaxRetries
}

// LifecycleCallback represents a signed callback for lifecycle state confirmation
type LifecycleCallback struct {
	// ID is the unique callback ID
	ID string `json:"id"`

	// OperationID is the operation this callback is for
	OperationID string `json:"operation_id"`

	// AllocationID is the allocation ID
	AllocationID string `json:"allocation_id"`

	// Action is the lifecycle action that was performed
	Action LifecycleActionType `json:"action"`

	// Success indicates if the action succeeded
	Success bool `json:"success"`

	// ResultState is the resulting allocation state
	ResultState AllocationState `json:"result_state"`

	// WaldurResourceID is the Waldur resource identifier
	WaldurResourceID string `json:"waldur_resource_id,omitempty"`

	// ProviderAddress is the provider that executed the action
	ProviderAddress string `json:"provider_address"`

	// Payload contains action-specific result data
	Payload map[string]string `json:"payload,omitempty"`

	// Error contains error details if failed
	Error string `json:"error,omitempty"`

	// ErrorCode is a machine-readable error code
	ErrorCode string `json:"error_code,omitempty"`

	// Signature is the cryptographic signature
	Signature []byte `json:"signature"`

	// SignerID identifies the signer (provider address)
	SignerID string `json:"signer_id"`

	// Nonce is for replay protection
	Nonce string `json:"nonce"`

	// Timestamp is when the callback was created
	Timestamp time.Time `json:"timestamp"`

	// ExpiresAt is when the callback expires
	ExpiresAt time.Time `json:"expires_at"`

	// IdempotencyKey links to the original operation
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// NewLifecycleCallback creates a new lifecycle callback
func NewLifecycleCallback(
	operationID string,
	allocationID string,
	action LifecycleActionType,
	success bool,
	resultState AllocationState,
	providerAddress string,
) *LifecycleCallback {
	return NewLifecycleCallbackAt(operationID, allocationID, action, success, resultState, providerAddress, time.Now())
}

// NewLifecycleCallbackAt creates a new lifecycle callback at a specific time
func NewLifecycleCallbackAt(
	operationID string,
	allocationID string,
	action LifecycleActionType,
	success bool,
	resultState AllocationState,
	providerAddress string,
	now time.Time,
) *LifecycleCallback {
	timestamp := now.UTC()
	nonce := generateNonceAt(timestamp)
	return &LifecycleCallback{
		ID:              fmt.Sprintf("lcb_%s_%s", operationID, nonce[:8]),
		OperationID:     operationID,
		AllocationID:    allocationID,
		Action:          action,
		Success:         success,
		ResultState:     resultState,
		ProviderAddress: providerAddress,
		Payload:         make(map[string]string),
		SignerID:        providerAddress,
		Nonce:           nonce,
		Timestamp:       timestamp,
		ExpiresAt:       timestamp.Add(time.Hour),
	}
}

// SigningPayload returns the payload to be signed
func (c *LifecycleCallback) SigningPayload() []byte {
	h := sha256.New()
	h.Write([]byte(c.ID))
	h.Write([]byte(c.OperationID))
	h.Write([]byte(c.AllocationID))
	h.Write([]byte(c.Action))
	if c.Success {
		h.Write([]byte("success"))
	} else {
		h.Write([]byte("failure"))
	}
	h.Write([]byte(c.ResultState.String()))
	h.Write([]byte(c.ProviderAddress))
	h.Write([]byte(c.Nonce))
	h.Write([]byte(fmt.Sprintf("%d", c.Timestamp.Unix())))
	return h.Sum(nil)
}

// Validate validates the callback
func (c *LifecycleCallback) Validate() error {
	return c.ValidateAt(time.Now())
}

// ValidateAt validates the callback at a specific time
func (c *LifecycleCallback) ValidateAt(now time.Time) error {
	if c.ID == "" {
		return fmt.Errorf("callback ID is required")
	}
	if c.OperationID == "" {
		return fmt.Errorf("operation ID is required")
	}
	if c.AllocationID == "" {
		return fmt.Errorf("allocation ID is required")
	}
	if !c.Action.IsValid() {
		return fmt.Errorf("invalid action: %s", c.Action)
	}
	if c.ProviderAddress == "" {
		return fmt.Errorf("provider address is required")
	}
	if c.Nonce == "" {
		return fmt.Errorf("nonce is required")
	}
	if len(c.Signature) == 0 {
		return fmt.Errorf("signature is required")
	}
	if now.After(c.ExpiresAt) {
		return ErrLifecycleCallbackExpired
	}
	return nil
}

// LifecycleOperationStore provides storage for lifecycle operations
type LifecycleOperationStore interface {
	// SaveOperation saves an operation
	SaveOperation(op *LifecycleOperation) error

	// GetOperation retrieves an operation by ID
	GetOperation(id string) (*LifecycleOperation, error)

	// GetOperationByIdempotencyKey retrieves an operation by idempotency key
	GetOperationByIdempotencyKey(key string) (*LifecycleOperation, error)

	// GetPendingOperations retrieves pending operations for an allocation
	GetPendingOperations(allocationID string) ([]*LifecycleOperation, error)

	// GetOperationsByState retrieves operations in a given state
	GetOperationsByState(state LifecycleOperationState) ([]*LifecycleOperation, error)
}

// Note: Lifecycle event types (EventLifecycleActionRequested, EventLifecycleActionStarted,
// EventLifecycleActionCompleted, EventLifecycleActionFailed, EventLifecycleCallbackReceived)
// are defined in events.go along with the event structs.

// LifecycleMetrics contains metrics for lifecycle operations
type LifecycleMetrics struct {
	// TotalOperations is total operations count
	TotalOperations int64 `json:"total_operations"`

	// PendingOperations is pending operations count
	PendingOperations int64 `json:"pending_operations"`

	// ExecutingOperations is executing operations count
	ExecutingOperations int64 `json:"executing_operations"`

	// CompletedOperations is completed operations count
	CompletedOperations int64 `json:"completed_operations"`

	// FailedOperations is failed operations count
	FailedOperations int64 `json:"failed_operations"`

	// RolledBackOperations is rolled back operations count
	RolledBackOperations int64 `json:"rolled_back_operations"`

	// OperationsByAction contains counts by action type
	OperationsByAction map[LifecycleActionType]int64 `json:"operations_by_action"`

	// AverageCompletionTimeMs is average completion time in milliseconds
	AverageCompletionTimeMs int64 `json:"average_completion_time_ms"`

	// LastUpdated is when metrics were last updated
	LastUpdated time.Time `json:"last_updated"`
}

// NewLifecycleMetrics creates new lifecycle metrics
func NewLifecycleMetrics() *LifecycleMetrics {
	return &LifecycleMetrics{
		OperationsByAction: make(map[LifecycleActionType]int64),
		LastUpdated:        time.Now().UTC(),
	}
}
