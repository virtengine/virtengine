// Package waldur provides Waldur lifecycle control API methods
//
// VE-4E: Resource lifecycle control via Waldur
package waldur

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// LifecycleAction represents the type of lifecycle action
type LifecycleAction string

const (
	// LifecycleActionStart starts a stopped resource
	LifecycleActionStart LifecycleAction = "start"

	// LifecycleActionStop stops a running resource
	LifecycleActionStop LifecycleAction = "stop"

	// LifecycleActionRestart restarts a resource
	LifecycleActionRestart LifecycleAction = "restart"

	// LifecycleActionSuspend suspends a resource
	LifecycleActionSuspend LifecycleAction = "suspend"

	// LifecycleActionResume resumes a suspended resource
	LifecycleActionResume LifecycleAction = "resume"

	// LifecycleActionResize resizes resource allocation
	LifecycleActionResize LifecycleAction = "resize"

	// LifecycleActionTerminate terminates a resource
	LifecycleActionTerminate LifecycleAction = "terminate"
)

// Operation state constants for polling
const (
	opStateDone     = "done"
	opStateCanceled = "canceled"
)

// LifecycleRequest contains parameters for lifecycle operations
type LifecycleRequest struct {
	// ResourceUUID is the Waldur resource UUID
	ResourceUUID string `json:"resource_uuid"`

	// Action is the lifecycle action to perform
	Action LifecycleAction `json:"action"`

	// IdempotencyKey prevents duplicate operations
	IdempotencyKey string `json:"idempotency_key,omitempty"`

	// CallbackURL is the URL for operation completion callbacks
	CallbackURL string `json:"callback_url,omitempty"`

	// Parameters contains action-specific parameters
	Parameters map[string]interface{} `json:"parameters,omitempty"`

	// Timeout is the operation timeout in seconds
	Timeout int `json:"timeout,omitempty"`

	// Immediate requests immediate execution (for stop/terminate)
	Immediate bool `json:"immediate,omitempty"`
}

// ResizeRequest contains resize-specific parameters
type ResizeRequest struct {
	LifecycleRequest

	// CPUCores is the new CPU core count
	CPUCores *int `json:"cpu_cores,omitempty"`

	// MemoryMB is the new memory in MB
	MemoryMB *int `json:"memory_mb,omitempty"`

	// DiskGB is the new disk size in GB
	DiskGB *int `json:"disk_gb,omitempty"`

	// Flavor is the new instance flavor (for OpenStack/Azure)
	Flavor string `json:"flavor,omitempty"`

	// InstanceType is the new instance type (for AWS)
	InstanceType string `json:"instance_type,omitempty"`
}

// LifecycleResponse contains the response from a lifecycle operation
type LifecycleResponse struct {
	// OperationID is the Waldur operation identifier
	OperationID string `json:"operation_id,omitempty"`

	// UUID is the operation UUID
	UUID string `json:"uuid,omitempty"`

	// State is the current operation state
	State string `json:"state"`

	// ResourceUUID is the resource being operated on
	ResourceUUID string `json:"resource_uuid,omitempty"`

	// ResourceState is the current resource state
	ResourceState string `json:"resource_state,omitempty"`

	// CreatedAt is when the operation was created
	CreatedAt time.Time `json:"created,omitempty"`

	// CompletedAt is when the operation completed
	CompletedAt *time.Time `json:"completed,omitempty"`

	// Error contains error details if failed
	Error string `json:"error,omitempty"`

	// ErrorCode is the error code if failed
	ErrorCode string `json:"error_code,omitempty"`
}

// ResourceState represents the state of a Waldur resource
type ResourceState string

const (
	ResourceStateCreating    ResourceState = "Creating"
	ResourceStateOK          ResourceState = "OK"
	ResourceStateErred       ResourceState = "Erred"
	ResourceStateUpdating    ResourceState = "Updating"
	ResourceStateTerminating ResourceState = "Terminating"
	ResourceStateTerminated  ResourceState = "Terminated"
	ResourceStateStopped     ResourceState = "Stopped"
	ResourceStatePaused      ResourceState = "Paused"
)

// LifecycleClient provides lifecycle operations for Waldur resources
type LifecycleClient struct {
	marketplace *MarketplaceClient
}

// NewLifecycleClient creates a new lifecycle client
func NewLifecycleClient(m *MarketplaceClient) *LifecycleClient {
	return &LifecycleClient{marketplace: m}
}

// Start starts a stopped resource
func (l *LifecycleClient) Start(ctx context.Context, req LifecycleRequest) (*LifecycleResponse, error) {
	req.Action = LifecycleActionStart
	return l.executeAction(ctx, req, "start")
}

// Stop stops a running resource
func (l *LifecycleClient) Stop(ctx context.Context, req LifecycleRequest) (*LifecycleResponse, error) {
	req.Action = LifecycleActionStop
	return l.executeAction(ctx, req, "stop")
}

// Restart restarts a resource
func (l *LifecycleClient) Restart(ctx context.Context, req LifecycleRequest) (*LifecycleResponse, error) {
	req.Action = LifecycleActionRestart
	return l.executeAction(ctx, req, "restart")
}

// Suspend suspends a resource (preserving state)
func (l *LifecycleClient) Suspend(ctx context.Context, req LifecycleRequest) (*LifecycleResponse, error) {
	req.Action = LifecycleActionSuspend
	return l.executeAction(ctx, req, "suspend")
}

// Resume resumes a suspended resource
func (l *LifecycleClient) Resume(ctx context.Context, req LifecycleRequest) (*LifecycleResponse, error) {
	req.Action = LifecycleActionResume
	return l.executeAction(ctx, req, "resume")
}

// Resize resizes a resource
func (l *LifecycleClient) Resize(ctx context.Context, req ResizeRequest) (*LifecycleResponse, error) {
	req.Action = LifecycleActionResize

	// Build resize parameters
	params := make(map[string]interface{})
	if req.CPUCores != nil {
		params["cpu_cores"] = *req.CPUCores
	}
	if req.MemoryMB != nil {
		params["memory_mb"] = *req.MemoryMB
	}
	if req.DiskGB != nil {
		params["disk_gb"] = *req.DiskGB
	}
	if req.Flavor != "" {
		params["flavor"] = req.Flavor
	}
	if req.InstanceType != "" {
		params["instance_type"] = req.InstanceType
	}
	for k, v := range req.Parameters {
		params[k] = v
	}
	req.Parameters = params

	return l.executeAction(ctx, req.LifecycleRequest, "resize")
}

// Terminate terminates a resource
func (l *LifecycleClient) Terminate(ctx context.Context, req LifecycleRequest) (*LifecycleResponse, error) {
	req.Action = LifecycleActionTerminate
	return l.executeAction(ctx, req, "terminate")
}

// GetOperationStatus gets the status of a lifecycle operation
func (l *LifecycleClient) GetOperationStatus(ctx context.Context, resourceUUID, operationID string) (*LifecycleResponse, error) {
	if resourceUUID == "" || operationID == "" {
		return nil, fmt.Errorf("resource UUID and operation ID are required")
	}

	var response *LifecycleResponse

	err := l.marketplace.client.doWithRetry(ctx, func() error {
		path := fmt.Sprintf("/marketplace-resources/%s/actions/%s/", resourceUUID, operationID)
		respBody, statusCode, err := l.marketplace.client.doRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return err
		}

		if statusCode != http.StatusOK {
			return mapHTTPError(statusCode, respBody)
		}

		if err := json.Unmarshal(respBody, &response); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}

		return nil
	})

	return response, err
}

// GetResourceState gets the current state of a resource
func (l *LifecycleClient) GetResourceState(ctx context.Context, resourceUUID string) (ResourceState, error) {
	if resourceUUID == "" {
		return "", fmt.Errorf("resource UUID is required")
	}

	resource, err := l.marketplace.GetResource(ctx, resourceUUID)
	if err != nil {
		return "", err
	}

	return ResourceState(resource.State), nil
}

// WaitForOperation waits for a lifecycle operation to complete
func (l *LifecycleClient) WaitForOperation(ctx context.Context, resourceUUID, operationID string, pollInterval time.Duration) (*LifecycleResponse, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			status, err := l.GetOperationStatus(ctx, resourceUUID, operationID)
			if err != nil {
				// Continue polling on transient errors
				continue
			}

			switch status.State {
			case opStateDone, "completed", "OK":
				return status, nil
			case "erred", "failed", "error":
				return status, fmt.Errorf("operation failed: %s", status.Error)
			case opStateCanceled, "cancelled":
				return status, fmt.Errorf("operation was cancelled")
			}
		}
	}
}

// executeAction executes a lifecycle action on a resource
func (l *LifecycleClient) executeAction(ctx context.Context, req LifecycleRequest, action string) (*LifecycleResponse, error) {
	if req.ResourceUUID == "" {
		return nil, fmt.Errorf("resource UUID is required")
	}

	var response *LifecycleResponse

	err := l.marketplace.client.doWithRetry(ctx, func() error {
		body := make(map[string]interface{})

		if req.IdempotencyKey != "" {
			body["idempotency_key"] = req.IdempotencyKey
		}
		if req.CallbackURL != "" {
			body["callback_url"] = req.CallbackURL
		}
		if req.Timeout > 0 {
			body["timeout"] = req.Timeout
		}
		if req.Immediate {
			body["immediate"] = req.Immediate
		}
		if len(req.Parameters) > 0 {
			for k, v := range req.Parameters {
				body[k] = v
			}
		}

		var bodyBytes []byte
		var err error
		if len(body) > 0 {
			bodyBytes, err = json.Marshal(body)
			if err != nil {
				return fmt.Errorf("marshal request: %w", err)
			}
		}

		path := fmt.Sprintf("/marketplace-resources/%s/%s/", req.ResourceUUID, action)
		respBody, statusCode, err := l.marketplace.client.doRequest(ctx, http.MethodPost, path, bodyBytes)
		if err != nil {
			return err
		}

		// Accept 200, 201, 202 for async operations
		if statusCode != http.StatusOK && statusCode != http.StatusCreated && statusCode != http.StatusAccepted {
			return mapHTTPError(statusCode, respBody)
		}

		// Parse response
		response = &LifecycleResponse{}
		if len(respBody) > 0 {
			if err := json.Unmarshal(respBody, response); err != nil {
				// Some actions return minimal response
				response.State = "accepted"
			}
		} else {
			response.State = "accepted"
		}

		response.ResourceUUID = req.ResourceUUID
		return nil
	})

	return response, err
}

// ValidateLifecycleAction validates if a lifecycle action can be performed
func ValidateLifecycleAction(currentState ResourceState, action LifecycleAction) error {
	validTransitions := map[ResourceState]map[LifecycleAction]bool{
		ResourceStateOK: {
			LifecycleActionStop:      true,
			LifecycleActionRestart:   true,
			LifecycleActionSuspend:   true,
			LifecycleActionResize:    true,
			LifecycleActionTerminate: true,
		},
		ResourceStateStopped: {
			LifecycleActionStart:     true,
			LifecycleActionResize:    true,
			LifecycleActionTerminate: true,
		},
		ResourceStatePaused: {
			LifecycleActionResume:    true,
			LifecycleActionTerminate: true,
		},
		ResourceStateCreating: {
			LifecycleActionTerminate: true,
		},
		ResourceStateUpdating: {
			LifecycleActionTerminate: true,
		},
	}

	transitions, ok := validTransitions[currentState]
	if !ok {
		return fmt.Errorf("no valid transitions from state %s", currentState)
	}

	if !transitions[action] {
		return fmt.Errorf("action %s not allowed in state %s", action, currentState)
	}

	return nil
}

// LifecycleCallbackPayload represents the callback payload from Waldur
type LifecycleCallbackPayload struct {
	// ResourceUUID is the resource identifier
	ResourceUUID string `json:"resource_uuid"`

	// OperationID is the operation identifier
	OperationID string `json:"operation_id"`

	// Action is the lifecycle action
	Action LifecycleAction `json:"action"`

	// State is the resulting state
	State string `json:"state"`

	// Success indicates if the action succeeded
	Success bool `json:"success"`

	// Error contains error details
	Error string `json:"error,omitempty"`

	// ErrorCode is the error code
	ErrorCode string `json:"error_code,omitempty"`

	// Timestamp is when the callback was generated
	Timestamp time.Time `json:"timestamp"`

	// BackendID is the VirtEngine allocation ID
	BackendID string `json:"backend_id,omitempty"`

	// IdempotencyKey is the original idempotency key
	IdempotencyKey string `json:"idempotency_key,omitempty"`

	// Metadata contains additional metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ParseLifecycleCallback parses a lifecycle callback payload
func ParseLifecycleCallback(data []byte) (*LifecycleCallbackPayload, error) {
	var payload LifecycleCallbackPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parse callback: %w", err)
	}

	if payload.ResourceUUID == "" {
		return nil, fmt.Errorf("resource UUID is required in callback")
	}

	return &payload, nil
}
