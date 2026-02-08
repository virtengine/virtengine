// Package types contains types for the HPC module.
//
// Agent messaging types for inter-agent handoff and task distribution
package types

import (
	"errors"
	"time"
)

// MessageType represents the type of agent message
type MessageType string

const (
	// MessageTypeHandoffRequest is a request to hand off a task
	MessageTypeHandoffRequest MessageType = "handoff_request"

	// MessageTypeHandoffResponse is a response to a handoff request
	MessageTypeHandoffResponse MessageType = "handoff_response"

	// MessageTypeNeedMoreRequest is a request for more tasks
	MessageTypeNeedMoreRequest MessageType = "needmore_request"

	// MessageTypeNeedMoreResponse is a response with available tasks
	MessageTypeNeedMoreResponse MessageType = "needmore_response"
)

// IsValidMessageType checks if the message type is valid
func IsValidMessageType(t MessageType) bool {
	switch t {
	case MessageTypeHandoffRequest, MessageTypeHandoffResponse,
		MessageTypeNeedMoreRequest, MessageTypeNeedMoreResponse:
		return true
	default:
		return false
	}
}

// MessagePriority represents the priority of a message
type MessagePriority int32

const (
	// MessagePriorityLow is for low priority messages
	MessagePriorityLow MessagePriority = 1

	// MessagePriorityNormal is for normal priority messages
	MessagePriorityNormal MessagePriority = 5

	// MessagePriorityHigh is for high priority messages
	MessagePriorityHigh MessagePriority = 10

	// MessagePriorityCritical is for critical messages
	MessagePriorityCritical MessagePriority = 20
)

// IsValidMessagePriority checks if the priority is valid
func IsValidMessagePriority(p MessagePriority) bool {
	return p >= MessagePriorityLow && p <= MessagePriorityCritical
}

// AgentMessage is the base message type for inter-agent communication
type AgentMessage struct {
	// MessageID is the unique identifier for this message
	MessageID string `json:"message_id"`

	// Type is the message type
	Type MessageType `json:"type"`

	// FromNodeID is the sender node ID
	FromNodeID string `json:"from_node_id"`

	// ToNodeID is the recipient node ID (empty for broadcast)
	ToNodeID string `json:"to_node_id,omitempty"`

	// ClusterID is the cluster context
	ClusterID string `json:"cluster_id"`

	// Priority is the message priority
	Priority MessagePriority `json:"priority"`

	// CreatedAt is when the message was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the message expires
	ExpiresAt time.Time `json:"expires_at"`

	// Payload contains the type-specific data
	Payload interface{} `json:"payload"`
}

// Validate validates an agent message
func (m *AgentMessage) Validate() error {
	if m.MessageID == "" {
		return errors.New("message_id required")
	}
	if !IsValidMessageType(m.Type) {
		return errors.New("invalid message type")
	}
	if m.FromNodeID == "" {
		return errors.New("from_node_id required")
	}
	if m.ClusterID == "" {
		return errors.New("cluster_id required")
	}
	if !IsValidMessagePriority(m.Priority) {
		return errors.New("invalid priority")
	}
	if m.CreatedAt.IsZero() {
		return errors.New("created_at required")
	}
	if m.ExpiresAt.Before(m.CreatedAt) {
		return errors.New("expires_at must be after created_at")
	}
	return nil
}

// IsExpired checks if the message has expired
func (m *AgentMessage) IsExpired() bool {
	return time.Now().After(m.ExpiresAt)
}

// HandoffRequest represents a request to hand off a task to another agent
type HandoffRequest struct {
	// TaskID is the ID of the task to hand off
	TaskID string `json:"task_id"`

	// JobID is the parent job ID
	JobID string `json:"job_id"`

	// Summary is a brief description of the task
	Summary string `json:"summary"`

	// Priority is the task priority
	Priority MessagePriority `json:"priority"`

	// Reason is why the handoff is requested
	Reason string `json:"reason"`

	// RequiredCapabilities lists required agent capabilities
	RequiredCapabilities AgentCapabilities `json:"required_capabilities"`

	// EstimatedRuntimeSeconds is the estimated time to complete
	EstimatedRuntimeSeconds int64 `json:"estimated_runtime_seconds"`

	// CheckpointData contains task state for resume
	CheckpointData string `json:"checkpoint_data,omitempty"`

	// DataReferences are input data locations
	DataReferences []DataReference `json:"data_references,omitempty"`
}

// Validate validates a handoff request
func (h *HandoffRequest) Validate() error {
	if h.TaskID == "" {
		return errors.New("task_id required")
	}
	if h.JobID == "" {
		return errors.New("job_id required")
	}
	if h.Summary == "" {
		return errors.New("summary required")
	}
	if len(h.Summary) > 500 {
		return errors.New("summary exceeds 500 characters")
	}
	if !IsValidMessagePriority(h.Priority) {
		return errors.New("invalid priority")
	}
	if h.Reason == "" {
		return errors.New("reason required")
	}
	if len(h.Reason) > 200 {
		return errors.New("reason exceeds 200 characters")
	}
	if h.EstimatedRuntimeSeconds <= 0 {
		return errors.New("estimated_runtime_seconds must be positive")
	}
	return h.RequiredCapabilities.Validate()
}

// HandoffResponse represents a response to a handoff request
type HandoffResponse struct {
	// RequestMessageID is the ID of the request being responded to
	RequestMessageID string `json:"request_message_id"`

	// Accepted indicates if the handoff was accepted
	Accepted bool `json:"accepted"`

	// Reason explains acceptance or rejection
	Reason string `json:"reason"`

	// RejectionCode categorizes rejection reason
	RejectionCode RejectionCode `json:"rejection_code,omitempty"`

	// AlternativeSuggestions lists other agents that might accept
	AlternativeSuggestions []string `json:"alternative_suggestions,omitempty"`

	// EstimatedStartTime is when the agent can start (if accepted)
	EstimatedStartTime *time.Time `json:"estimated_start_time,omitempty"`
}

// Validate validates a handoff response
func (h *HandoffResponse) Validate() error {
	if h.RequestMessageID == "" {
		return errors.New("request_message_id required")
	}
	if h.Reason == "" {
		return errors.New("reason required")
	}
	if len(h.Reason) > 500 {
		return errors.New("reason exceeds 500 characters")
	}
	if !h.Accepted && !IsValidRejectionCode(h.RejectionCode) {
		return errors.New("rejection_code required when not accepted")
	}
	return nil
}

// RejectionCode categorizes handoff rejection reasons
type RejectionCode string

const (
	// RejectionCodeOverloaded indicates the agent is at capacity
	RejectionCodeOverloaded RejectionCode = "overloaded"

	// RejectionCodeIncompatible indicates capability mismatch
	RejectionCodeIncompatible RejectionCode = "incompatible"

	// RejectionCodeBusy indicates the agent is busy with critical tasks
	RejectionCodeBusy RejectionCode = "busy"

	// RejectionCodeDraining indicates the agent is draining
	RejectionCodeDraining RejectionCode = "draining"

	// RejectionCodeUnhealthy indicates the agent is unhealthy
	RejectionCodeUnhealthy RejectionCode = "unhealthy"

	// RejectionCodeLowPriority indicates the task priority is too low
	RejectionCodeLowPriority RejectionCode = "low_priority"

	// RejectionCodeBlacklisted indicates the requester is blacklisted
	RejectionCodeBlacklisted RejectionCode = "blacklisted"
)

// IsValidRejectionCode checks if the rejection code is valid
func IsValidRejectionCode(c RejectionCode) bool {
	switch c {
	case RejectionCodeOverloaded, RejectionCodeIncompatible, RejectionCodeBusy,
		RejectionCodeDraining, RejectionCodeUnhealthy, RejectionCodeLowPriority,
		RejectionCodeBlacklisted:
		return true
	default:
		return false
	}
}

// NeedMoreRequest represents a request for more tasks
type NeedMoreRequest struct {
	// AvailableCapacity describes available resources
	AvailableCapacity NodeCapacity `json:"available_capacity"`

	// Capabilities lists agent capabilities
	Capabilities AgentCapabilities `json:"capabilities"`

	// MaxTasks is the maximum number of tasks requested
	MaxTasks int32 `json:"max_tasks"`

	// PreferredPriority is the preferred task priority
	PreferredPriority MessagePriority `json:"preferred_priority"`
}

// Validate validates a needmore request
func (n *NeedMoreRequest) Validate() error {
	if err := n.AvailableCapacity.Validate(); err != nil {
		return err
	}
	if err := n.Capabilities.Validate(); err != nil {
		return err
	}
	if n.MaxTasks <= 0 || n.MaxTasks > 100 {
		return errors.New("max_tasks must be 1-100")
	}
	if !IsValidMessagePriority(n.PreferredPriority) {
		return errors.New("invalid preferred_priority")
	}
	return nil
}

// NeedMoreResponse represents a response with available tasks
type NeedMoreResponse struct {
	// RequestMessageID is the ID of the request being responded to
	RequestMessageID string `json:"request_message_id"`

	// TaskIDs lists available task IDs
	TaskIDs []string `json:"task_ids"`

	// JobIDs maps task IDs to job IDs
	JobIDs map[string]string `json:"job_ids,omitempty"`

	// EstimatedRuntimes maps task IDs to estimated runtimes
	EstimatedRuntimes map[string]int64 `json:"estimated_runtimes,omitempty"`

	// Priorities maps task IDs to priorities
	Priorities map[string]MessagePriority `json:"priorities,omitempty"`

	// NoTasksAvailable indicates no tasks match criteria
	NoTasksAvailable bool `json:"no_tasks_available"`

	// RetryAfterSeconds suggests when to retry
	RetryAfterSeconds int32 `json:"retry_after_seconds,omitempty"`
}

// Validate validates a needmore response
func (n *NeedMoreResponse) Validate() error {
	if n.RequestMessageID == "" {
		return errors.New("request_message_id required")
	}
	if !n.NoTasksAvailable && len(n.TaskIDs) == 0 {
		return errors.New("task_ids required when tasks available")
	}
	if n.NoTasksAvailable && len(n.TaskIDs) > 0 {
		return errors.New("cannot have both no_tasks_available and task_ids")
	}
	if len(n.TaskIDs) > 100 {
		return errors.New("task_ids exceeds maximum of 100")
	}
	return nil
}

// AgentCapabilities describes agent capabilities for matching
type AgentCapabilities struct {
	// GPUTypes lists supported GPU types (empty = no GPU)
	GPUTypes []string `json:"gpu_types,omitempty"`

	// MinMemoryGB is minimum available memory
	MinMemoryGB int32 `json:"min_memory_gb"`

	// MinCPUCores is minimum available CPU cores
	MinCPUCores int32 `json:"min_cpu_cores"`

	// MinGPUs is minimum available GPUs
	MinGPUs int32 `json:"min_gpus"`

	// SupportedContainerRuntimes lists supported runtimes
	SupportedContainerRuntimes []string `json:"supported_container_runtimes"`

	// Features lists special features (mpi, cuda, etc.)
	Features []string `json:"features,omitempty"`

	// MaxTaskDurationSeconds is the maximum task duration
	MaxTaskDurationSeconds int64 `json:"max_task_duration_seconds"`
}

// Validate validates agent capabilities
func (c *AgentCapabilities) Validate() error {
	if c.MinMemoryGB < 0 {
		return errors.New("min_memory_gb cannot be negative")
	}
	if c.MinCPUCores < 0 {
		return errors.New("min_cpu_cores cannot be negative")
	}
	if c.MinGPUs < 0 {
		return errors.New("min_gpus cannot be negative")
	}
	if c.MaxTaskDurationSeconds <= 0 {
		return errors.New("max_task_duration_seconds must be positive")
	}
	if len(c.SupportedContainerRuntimes) == 0 {
		return errors.New("supported_container_runtimes required")
	}
	return nil
}

// Matches checks if capabilities match requirements
func (c *AgentCapabilities) Matches(required AgentCapabilities) bool {
	// Check memory
	if c.MinMemoryGB < required.MinMemoryGB {
		return false
	}

	// Check CPU
	if c.MinCPUCores < required.MinCPUCores {
		return false
	}

	// Check GPU
	if c.MinGPUs < required.MinGPUs {
		return false
	}

	// Check GPU types if required
	if len(required.GPUTypes) > 0 {
		hasMatch := false
		for _, reqType := range required.GPUTypes {
			for _, availType := range c.GPUTypes {
				if reqType == availType {
					hasMatch = true
					break
				}
			}
			if hasMatch {
				break
			}
		}
		if !hasMatch {
			return false
		}
	}

	// Check container runtime
	hasRuntime := false
	for _, reqRuntime := range required.SupportedContainerRuntimes {
		for _, availRuntime := range c.SupportedContainerRuntimes {
			if reqRuntime == availRuntime {
				hasRuntime = true
				break
			}
		}
		if hasRuntime {
			break
		}
	}
	if !hasRuntime {
		return false
	}

	// Check features
	for _, reqFeature := range required.Features {
		found := false
		for _, availFeature := range c.Features {
			if reqFeature == availFeature {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check duration
	if c.MaxTaskDurationSeconds < required.MaxTaskDurationSeconds {
		return false
	}

	return true
}
