package provider_daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// LifecycleExecutor defines lifecycle operations used by the portal API.
type LifecycleExecutor interface {
	Start(ctx context.Context, req *LifecycleActionRequest) (*LifecycleActionResult, error)
	Stop(ctx context.Context, req *LifecycleActionRequest) (*LifecycleActionResult, error)
	Restart(ctx context.Context, req *LifecycleActionRequest) (*LifecycleActionResult, error)
	Resize(ctx context.Context, req *LifecycleActionRequest) (*LifecycleActionResult, error)
}

// DeploymentActionRequest is the payload for deployment lifecycle actions.
type DeploymentActionRequest struct {
	Action     string                 `json:"action"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// DeploymentActionResponse is the response for lifecycle actions.
type DeploymentActionResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message,omitempty"`
	OperationID string `json:"operation_id,omitempty"`
	State       string `json:"state,omitempty"`
}

func (s *PortalAPIServer) handleDeploymentAction(w http.ResponseWriter, r *http.Request) {
	deploymentID := mux.Vars(r)["deploymentId"]
	if deploymentID == "" {
		writeJSONError(w, http.StatusBadRequest, "deployment id required")
		return
	}

	if s.lifecycleExec == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "lifecycle service unavailable")
		return
	}

	var reqBody DeploymentActionRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	action := strings.ToLower(strings.TrimSpace(reqBody.Action))
	if action == "" {
		writeJSONError(w, http.StatusBadRequest, "action required")
		return
	}

	authCtx := authFromContext(r.Context())
	if authCtx.Address == "" {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if err := s.authorizeLifecycleAction(r.Context(), authCtx.Address, action); err != nil {
		writeJSONError(w, http.StatusForbidden, err.Error())
		return
	}

	actionReq := &LifecycleActionRequest{
		AllocationID: deploymentID,
		RequestedBy:  authCtx.Address,
		Parameters:   mapStringInterfaceToString(reqBody.Parameters),
	}

	var (
		result *LifecycleActionResult
		err    error
	)

	switch action {
	case "start":
		result, err = s.lifecycleExec.Start(r.Context(), actionReq)
	case "stop":
		result, err = s.lifecycleExec.Stop(r.Context(), actionReq)
	case "restart":
		result, err = s.lifecycleExec.Restart(r.Context(), actionReq)
	case "resize":
		resizeSpec, specErr := parseResizeSpec(reqBody.Parameters)
		if specErr != nil {
			writeJSONError(w, http.StatusBadRequest, specErr.Error())
			return
		}
		actionReq.ResizeSpec = resizeSpec
		result, err = s.lifecycleExec.Resize(r.Context(), actionReq)
	default:
		writeJSONError(w, http.StatusBadRequest, "unsupported action")
		return
	}

	if err != nil {
		switch {
		case errors.Is(err, ErrResourceNotFound):
			writeJSONError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrResourceBusy):
			writeJSONError(w, http.StatusConflict, err.Error())
		case errors.Is(err, ErrInvalidResourceState), errors.Is(err, ErrResizeNotSupported):
			writeJSONError(w, http.StatusBadRequest, err.Error())
		default:
			writeJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	resp := DeploymentActionResponse{
		Success: result != nil && result.Success,
	}
	if result != nil {
		resp.Message = result.Message
		resp.OperationID = result.OperationID
		resp.State = string(result.State)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *PortalAPIServer) authorizeLifecycleAction(ctx context.Context, address, action string) error {
	if address == "" {
		return errors.New("authentication required")
	}

	if s.cfg.LifecycleRequireConsent {
		if s.cfg.LifecycleConsentScope == "" {
			return errors.New("consent scope not configured")
		}
		hasConsent, err := s.chainQuery.HasConsent(ctx, address, s.cfg.LifecycleConsentScope)
		if err != nil {
			return err
		}
		if !hasConsent {
			return fmt.Errorf("consent not granted for lifecycle action %s", action)
		}
	}

	allowedRoles := s.cfg.LifecycleAllowedRoles
	if len(allowedRoles) == 0 {
		return nil
	}

	for _, role := range allowedRoles {
		ok, err := s.chainQuery.HasRole(ctx, address, role)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}

	return fmt.Errorf("insufficient role for lifecycle action %s", action)
}

func parseResizeSpec(params map[string]interface{}) (*marketplace.ResizeSpecification, error) {
	if len(params) == 0 {
		return nil, errors.New("resize parameters required")
	}

	var spec marketplace.ResizeSpecification
	set := false

	if value, ok := readIntParam(params, "cpu_cores"); ok {
		v, err := toUint32(value, "cpu_cores")
		if err != nil {
			return nil, err
		}
		spec.CPUCores = &v
		set = true
	}
	if value, ok := readIntParam(params, "memory_mb"); ok {
		v, err := toUint64(value, "memory_mb")
		if err != nil {
			return nil, err
		}
		spec.MemoryMB = &v
		set = true
	}
	if value, ok := readIntParam(params, "storage_gb"); ok {
		v, err := toUint64(value, "storage_gb")
		if err != nil {
			return nil, err
		}
		spec.StorageGB = &v
		set = true
	}
	if value, ok := readIntParam(params, "gpu_count"); ok {
		v, err := toUint32(value, "gpu_count")
		if err != nil {
			return nil, err
		}
		spec.GPUCount = &v
		set = true
	}
	if value, ok := params["custom_limits"]; ok {
		limits, err := parseCustomLimits(value)
		if err != nil {
			return nil, err
		}
		spec.CustomLimits = limits
		set = true
	}

	if !set {
		return nil, errors.New("resize parameters missing")
	}

	return &spec, nil
}

func readIntParam(params map[string]interface{}, key string) (int64, bool) {
	value, ok := params[key]
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		return int64(v), true
	case json.Number:
		parsed, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return parsed, true
	case string:
		parsed, err := json.Number(v).Int64()
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func parseCustomLimits(value interface{}) (map[string]int64, error) {
	limits := make(map[string]int64)
	switch v := value.(type) {
	case map[string]interface{}:
		for key := range v {
			parsed, ok := readIntParam(v, key)
			if !ok {
				return nil, errors.New("invalid custom_limits value")
			}
			limits[key] = parsed
		}
	case map[string]int64:
		for key, raw := range v {
			limits[key] = raw
		}
	default:
		return nil, errors.New("invalid custom_limits value")
	}
	return limits, nil
}

func toUint32(value int64, field string) (uint32, error) {
	const maxUint32 = int64(^uint32(0))
	if value < 0 || value > maxUint32 {
		return 0, fmt.Errorf("%s out of range", field)
	}
	return uint32(value), nil
}

func toUint64(value int64, field string) (uint64, error) {
	if value < 0 {
		return 0, fmt.Errorf("%s must be non-negative", field)
	}
	return uint64(value), nil
}

func mapStringInterfaceToString(params map[string]interface{}) map[string]string {
	if len(params) == 0 {
		return nil
	}
	result := make(map[string]string, len(params))
	for key, value := range params {
		switch v := value.(type) {
		case string:
			result[key] = v
		case json.Number:
			result[key] = v.String()
		case float64:
			result[key] = json.Number(fmt.Sprintf("%v", v)).String()
		case int, int32, int64, uint, uint32, uint64:
			result[key] = fmt.Sprintf("%v", v)
		case bool:
			result[key] = fmt.Sprintf("%t", v)
		}
	}
	return result
}
