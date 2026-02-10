// Package provider_daemon implements provider-side services for VirtEngine.
//
// VE-34E: Lifecycle command executor for Waldur lifecycle APIs.
package provider_daemon

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// WaldurLifecycleCommandExecutor executes lifecycle commands using Waldur APIs.
type WaldurLifecycleCommandExecutor struct {
	lifecycle *waldur.LifecycleClient
}

// NewWaldurLifecycleCommandExecutor creates a new executor.
func NewWaldurLifecycleCommandExecutor(lifecycle *waldur.LifecycleClient) *WaldurLifecycleCommandExecutor {
	return &WaldurLifecycleCommandExecutor{lifecycle: lifecycle}
}

// Execute executes the lifecycle command.
func (e *WaldurLifecycleCommandExecutor) Execute(ctx context.Context, cmd *LifecycleCommand) (*LifecycleCommandExecutionResult, error) {
	if cmd == nil {
		return nil, fmt.Errorf("command is nil")
	}
	if cmd.ResourceUUID == "" {
		return nil, fmt.Errorf("resource UUID is required")
	}
	if e.lifecycle == nil {
		return nil, fmt.Errorf("waldur lifecycle client is not configured")
	}

	params := make(map[string]interface{})
	for k, v := range cmd.Parameters {
		params[k] = v
	}

	req := waldur.LifecycleRequest{
		ResourceUUID:   cmd.ResourceUUID,
		IdempotencyKey: cmd.IdempotencyKey,
		Parameters:     params,
	}

	var resp *waldur.LifecycleResponse
	var err error

	switch cmd.Action {
	case marketplace.LifecycleActionStart:
		resp, err = e.lifecycle.Start(ctx, req)
	case marketplace.LifecycleActionStop:
		resp, err = e.lifecycle.Stop(ctx, req)
	case marketplace.LifecycleActionRestart:
		resp, err = e.lifecycle.Restart(ctx, req)
	case marketplace.LifecycleActionSuspend:
		resp, err = e.lifecycle.Suspend(ctx, req)
	case marketplace.LifecycleActionResume:
		resp, err = e.lifecycle.Resume(ctx, req)
	case marketplace.LifecycleActionTerminate:
		resp, err = e.lifecycle.Terminate(ctx, req)
	case marketplace.LifecycleActionResize:
		resizeReq := waldur.ResizeRequest{LifecycleRequest: req}
		applyResizeParams(&resizeReq, params)
		resp, err = e.lifecycle.Resize(ctx, resizeReq)
	case marketplace.LifecycleActionProvision:
		return nil, fmt.Errorf("provision action is handled via allocation provisioning flow")
	default:
		return nil, fmt.Errorf("unsupported lifecycle action: %s", cmd.Action)
	}

	if err != nil {
		return nil, err
	}

	result := &LifecycleCommandExecutionResult{}
	if resp != nil {
		result.WaldurOperationID = resp.OperationID
		if result.WaldurOperationID == "" {
			result.WaldurOperationID = resp.UUID
		}
		result.ResourceState = resp.ResourceState
	}

	return result, nil
}

// GetResourceState returns the resource state via Waldur API.
func (e *WaldurLifecycleCommandExecutor) GetResourceState(ctx context.Context, resourceUUID string) (waldur.ResourceState, error) {
	if e.lifecycle == nil {
		return "", fmt.Errorf("waldur lifecycle client is not configured")
	}
	return e.lifecycle.GetResourceState(ctx, resourceUUID)
}

func applyResizeParams(req *waldur.ResizeRequest, params map[string]interface{}) {
	if params == nil {
		return
	}
	if cpu := intFromParam(params["cpu_cores"]); cpu > 0 {
		req.CPUCores = &cpu
	}
	if mem := intFromParam(params["memory_mb"]); mem > 0 {
		req.MemoryMB = &mem
	}
	if disk := intFromParam(params["disk_gb"]); disk > 0 {
		req.DiskGB = &disk
	}
	if disk := intFromParam(params["storage_gb"]); disk > 0 && req.DiskGB == nil {
		req.DiskGB = &disk
	}
	if flavor := stringFromParam(params["flavor"]); flavor != "" {
		req.Flavor = flavor
	}
	if it := stringFromParam(params["instance_type"]); it != "" {
		req.InstanceType = it
	}
}

func intFromParam(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		if v > math.MaxInt || v < math.MinInt {
			return 0
		}
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		if v > uint64(math.MaxInt) {
			return 0
		}
		return int(v)
	case float64:
		return int(v)
	case string:
		parsed, err := strconv.Atoi(v)
		if err == nil {
			return parsed
		}
	}
	return 0
}

func stringFromParam(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return ""
	}
}
