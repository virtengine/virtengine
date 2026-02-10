// Package workflow provides persistent workflow state storage
package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// WorkflowState represents the complete state of a workflow execution
type WorkflowState struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Status         WorkflowStatus         `json:"status"`
	CurrentStep    string                 `json:"current_step"`
	Data           map[string]interface{} `json:"data"`
	Error          string                 `json:"error,omitempty"`
	RetryCount     int                    `json:"retry_count"`
	MaxRetries     int                    `json:"max_retries"`
	StartedAt      time.Time              `json:"started_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	CompletedSteps []string               `json:"completed_steps"`
	Metadata       map[string]string      `json:"metadata,omitempty"`
}

// WorkflowStatus represents the status of a workflow
type WorkflowStatus string

const (
	// WorkflowStatusPending indicates the workflow is queued but not started
	WorkflowStatusPending WorkflowStatus = "pending"
	// WorkflowStatusRunning indicates the workflow is actively executing
	WorkflowStatusRunning WorkflowStatus = "running"
	// WorkflowStatusPaused indicates the workflow is paused and can be resumed
	WorkflowStatusPaused WorkflowStatus = "paused"
	// WorkflowStatusCompleted indicates the workflow finished successfully
	WorkflowStatusCompleted WorkflowStatus = "completed"
	// WorkflowStatusFailed indicates the workflow failed with an error
	WorkflowStatusFailed WorkflowStatus = "failed"
	// WorkflowStatusCancelled indicates the workflow was manually cancelled
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
)

// StateFilter defines filters for listing workflow states
type StateFilter struct {
	// Status filters by workflow status
	Status WorkflowStatus `json:"status,omitempty"`
	// Name filters by workflow name
	Name string `json:"name,omitempty"`
	// StartedAfter filters workflows started after this time
	StartedAfter *time.Time `json:"started_after,omitempty"`
	// StartedBefore filters workflows started before this time
	StartedBefore *time.Time `json:"started_before,omitempty"`
	// Limit limits the number of results
	Limit int `json:"limit,omitempty"`
	// Offset for pagination
	Offset int `json:"offset,omitempty"`
}

// HistoryEvent represents an event in workflow history
type HistoryEvent struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflow_id"`
	EventType   HistoryEventType       `json:"event_type"`
	Step        string                 `json:"step,omitempty"`
	Status      WorkflowStatus         `json:"status,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	DurationMs  int64                  `json:"duration_ms,omitempty"`
	ActorID     string                 `json:"actor_id,omitempty"`
	Description string                 `json:"description,omitempty"`
}

// HistoryEventType defines types of workflow history events
type HistoryEventType string

const (
	// HistoryEventWorkflowStarted indicates workflow started
	HistoryEventWorkflowStarted HistoryEventType = "workflow_started"
	// HistoryEventStepStarted indicates a step started
	HistoryEventStepStarted HistoryEventType = "step_started"
	// HistoryEventStepCompleted indicates a step completed
	HistoryEventStepCompleted HistoryEventType = "step_completed"
	// HistoryEventStepFailed indicates a step failed
	HistoryEventStepFailed HistoryEventType = "step_failed"
	// HistoryEventStepRetried indicates a step was retried
	HistoryEventStepRetried HistoryEventType = "step_retried"
	// HistoryEventStepSkipped indicates a step was skipped
	HistoryEventStepSkipped HistoryEventType = "step_skipped"
	// HistoryEventWorkflowPaused indicates workflow was paused
	HistoryEventWorkflowPaused HistoryEventType = "workflow_paused"
	// HistoryEventWorkflowResumed indicates workflow was resumed
	HistoryEventWorkflowResumed HistoryEventType = "workflow_resumed"
	// HistoryEventWorkflowCompleted indicates workflow completed
	HistoryEventWorkflowCompleted HistoryEventType = "workflow_completed"
	// HistoryEventWorkflowFailed indicates workflow failed
	HistoryEventWorkflowFailed HistoryEventType = "workflow_failed"
	// HistoryEventWorkflowCancelled indicates workflow was cancelled
	HistoryEventWorkflowCancelled HistoryEventType = "workflow_cancelled"
	// HistoryEventCheckpointSaved indicates a checkpoint was saved
	HistoryEventCheckpointSaved HistoryEventType = "checkpoint_saved"
	// HistoryEventRecoveryStarted indicates recovery from checkpoint started
	HistoryEventRecoveryStarted HistoryEventType = "recovery_started"
)

// WorkflowStore defines the interface for persistent workflow state storage
type WorkflowStore interface {
	// SaveState saves or updates a workflow state
	SaveState(ctx context.Context, id string, state *WorkflowState) error

	// LoadState loads a workflow state by ID
	LoadState(ctx context.Context, id string) (*WorkflowState, error)

	// DeleteState deletes a workflow state
	DeleteState(ctx context.Context, id string) error

	// ListStates lists workflow states matching the filter
	ListStates(ctx context.Context, filter StateFilter) ([]*WorkflowState, error)

	// SaveCheckpoint saves a workflow checkpoint
	SaveCheckpoint(ctx context.Context, workflowID string, checkpoint *Checkpoint) error

	// LoadLatestCheckpoint loads the most recent checkpoint for a workflow
	LoadLatestCheckpoint(ctx context.Context, workflowID string) (*Checkpoint, error)

	// LoadCheckpoint loads a specific checkpoint by workflow and step
	LoadCheckpoint(ctx context.Context, workflowID, step string) (*Checkpoint, error)

	// ListCheckpoints lists all checkpoints for a workflow
	ListCheckpoints(ctx context.Context, workflowID string) ([]*Checkpoint, error)

	// DeleteCheckpoints deletes all checkpoints for a workflow
	DeleteCheckpoints(ctx context.Context, workflowID string) error

	// AppendHistory appends an event to workflow history
	AppendHistory(ctx context.Context, workflowID string, event *HistoryEvent) error

	// GetHistory retrieves the complete history for a workflow
	GetHistory(ctx context.Context, workflowID string) ([]*HistoryEvent, error)

	// DeleteHistory deletes all history for a workflow
	DeleteHistory(ctx context.Context, workflowID string) error

	// Close closes the store and releases resources
	Close() error
}

// WorkflowStoreConfig configures the workflow store backend
type WorkflowStoreConfig struct {
	// Type specifies the storage backend: "memory", "redis", or "postgres"
	Type string `json:"type" yaml:"type"`

	// RedisURL is the Redis connection URL (redis://user:pass@host:port/db)
	RedisURL string `json:"redis_url,omitempty" yaml:"redis_url,omitempty"`

	// RedisPrefix is the key prefix for Redis keys (default: "workflow")
	RedisPrefix string `json:"redis_prefix,omitempty" yaml:"redis_prefix,omitempty"`

	// PostgresURL is the PostgreSQL connection URL
	PostgresURL string `json:"postgres_url,omitempty" yaml:"postgres_url,omitempty"`

	// StateTTL is how long to keep completed workflow states (0 = forever)
	StateTTL time.Duration `json:"state_ttl,omitempty" yaml:"state_ttl,omitempty"`

	// HistoryTTL is how long to keep workflow history (0 = forever)
	HistoryTTL time.Duration `json:"history_ttl,omitempty" yaml:"history_ttl,omitempty"`

	// MaxHistoryPerWorkflow limits history events per workflow (0 = unlimited)
	MaxHistoryPerWorkflow int `json:"max_history_per_workflow,omitempty" yaml:"max_history_per_workflow,omitempty"`
}

// DefaultWorkflowStoreConfig returns a default configuration using in-memory storage
func DefaultWorkflowStoreConfig() WorkflowStoreConfig {
	return WorkflowStoreConfig{
		Type:                  "memory",
		RedisPrefix:           "workflow",
		StateTTL:              24 * time.Hour * 7,  // 7 days
		HistoryTTL:            24 * time.Hour * 30, // 30 days
		MaxHistoryPerWorkflow: 10000,
	}
}

// Validate validates the configuration
func (c WorkflowStoreConfig) Validate() error {
	switch c.Type {
	case "memory":
		// No additional validation needed
	case "redis":
		if c.RedisURL == "" {
			return fmt.Errorf("redis_url is required when type is 'redis'")
		}
	case "postgres":
		if c.PostgresURL == "" {
			return fmt.Errorf("postgres_url is required when type is 'postgres'")
		}
	default:
		return fmt.Errorf("unsupported store type: %s (supported: memory, redis, postgres)", c.Type)
	}
	return nil
}

// NewWorkflowStore creates a new workflow store based on configuration
func NewWorkflowStore(ctx context.Context, config WorkflowStoreConfig) (WorkflowStore, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	switch config.Type {
	case "memory":
		return NewMemoryWorkflowStore(config), nil
	case "redis":
		return NewRedisWorkflowStore(ctx, config)
	case "postgres":
		return nil, fmt.Errorf("postgres backend not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported store type: %s", config.Type)
	}
}

// MarshalWorkflowState marshals workflow state to JSON
func MarshalWorkflowState(state *WorkflowState) ([]byte, error) {
	return json.Marshal(state)
}

// UnmarshalWorkflowState unmarshals workflow state from JSON
func UnmarshalWorkflowState(data []byte) (*WorkflowState, error) {
	var state WorkflowState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// MarshalCheckpoint marshals checkpoint to JSON
func MarshalCheckpoint(cp *Checkpoint) ([]byte, error) {
	return json.Marshal(cp)
}

// UnmarshalCheckpoint unmarshals checkpoint from JSON
func UnmarshalCheckpoint(data []byte) (*Checkpoint, error) {
	var cp Checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, err
	}
	return &cp, nil
}

// MarshalHistoryEvent marshals history event to JSON
func MarshalHistoryEvent(event *HistoryEvent) ([]byte, error) {
	return json.Marshal(event)
}

// UnmarshalHistoryEvent unmarshals history event from JSON
func UnmarshalHistoryEvent(data []byte) (*HistoryEvent, error) {
	var event HistoryEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}
