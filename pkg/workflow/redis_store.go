// Package workflow provides Redis-backed workflow state storage
package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisWorkflowStore implements WorkflowStore using Redis as the backend
type RedisWorkflowStore struct {
	client *redis.Client
	prefix string
	config WorkflowStoreConfig
}

// Redis key patterns
const (
	stateKeyPattern      = "%s:state:%s"              // prefix:state:workflowID
	stateIndexKey        = "%s:states"                // prefix:states (ZSET for listing)
	checkpointKeyPattern = "%s:checkpoint:%s:%s"      // prefix:checkpoint:workflowID:step
	checkpointIndexKey   = "%s:checkpoints:%s"        // prefix:checkpoints:workflowID (SET)
	historyKeyPattern    = "%s:history:%s"            // prefix:history:workflowID (LIST)
)

// NewRedisWorkflowStore creates a new Redis-backed workflow store
func NewRedisWorkflowStore(ctx context.Context, config WorkflowStoreConfig) (*RedisWorkflowStore, error) {
	opts, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	prefix := config.RedisPrefix
	if prefix == "" {
		prefix = "workflow"
	}

	return &RedisWorkflowStore{
		client: client,
		prefix: prefix,
		config: config,
	}, nil
}

// SaveState saves or updates a workflow state
func (s *RedisWorkflowStore) SaveState(ctx context.Context, id string, state *WorkflowState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	key := fmt.Sprintf(stateKeyPattern, s.prefix, id)
	indexKey := fmt.Sprintf(stateIndexKey, s.prefix)

	// Use pipeline for atomic operation
	pipe := s.client.Pipeline()

	// Set state with optional TTL
	if s.config.StateTTL > 0 && isTerminalStatus(state.Status) {
		pipe.Set(ctx, key, data, s.config.StateTTL)
	} else {
		pipe.Set(ctx, key, data, 0)
	}

	// Add to index ZSET with score = started timestamp
	score := float64(state.StartedAt.UnixNano())
	pipe.ZAdd(ctx, indexKey, redis.Z{Score: score, Member: id})

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// LoadState loads a workflow state by ID
func (s *RedisWorkflowStore) LoadState(ctx context.Context, id string) (*WorkflowState, error) {
	key := fmt.Sprintf(stateKeyPattern, s.prefix, id)

	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	var state WorkflowState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// DeleteState deletes a workflow state
func (s *RedisWorkflowStore) DeleteState(ctx context.Context, id string) error {
	key := fmt.Sprintf(stateKeyPattern, s.prefix, id)
	indexKey := fmt.Sprintf(stateIndexKey, s.prefix)

	pipe := s.client.Pipeline()
	pipe.Del(ctx, key)
	pipe.ZRem(ctx, indexKey, id)
	_, err := pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete state: %w", err)
	}

	return nil
}

// ListStates lists workflow states matching the filter
func (s *RedisWorkflowStore) ListStates(ctx context.Context, filter StateFilter) ([]*WorkflowState, error) {
	indexKey := fmt.Sprintf(stateIndexKey, s.prefix)

	// Get all workflow IDs from the index
	// Use ZREVRANGE for newest-first ordering
	start := int64(0)
	stop := int64(-1) // Get all

	if filter.Limit > 0 {
		// Apply offset and limit at Redis level for efficiency
		start = int64(filter.Offset)
		stop = int64(filter.Offset + filter.Limit - 1)
	}

	ids, err := s.client.ZRevRange(ctx, indexKey, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow IDs: %w", err)
	}

	if len(ids) == 0 {
		return []*WorkflowState{}, nil
	}

	// Fetch states using pipeline
	pipe := s.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(ids))
	for i, id := range ids {
		key := fmt.Sprintf(stateKeyPattern, s.prefix, id)
		cmds[i] = pipe.Get(ctx, key)
	}
	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to fetch states: %w", err)
	}

	// Parse results and apply filters
	var results []*WorkflowState
	for _, cmd := range cmds {
		data, err := cmd.Bytes()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			continue // Skip invalid entries
		}

		var state WorkflowState
		if err := json.Unmarshal(data, &state); err != nil {
			continue
		}

		// Apply filters
		if !matchesStateFilter(&state, filter) {
			continue
		}

		results = append(results, &state)
	}

	return results, nil
}

// SaveCheckpoint saves a workflow checkpoint
func (s *RedisWorkflowStore) SaveCheckpoint(ctx context.Context, workflowID string, checkpoint *Checkpoint) error {
	data, err := json.Marshal(checkpoint)
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	key := fmt.Sprintf(checkpointKeyPattern, s.prefix, workflowID, checkpoint.Step)
	indexKey := fmt.Sprintf(checkpointIndexKey, s.prefix, workflowID)

	pipe := s.client.Pipeline()
	pipe.Set(ctx, key, data, 0) // Checkpoints don't expire while workflow is active
	pipe.SAdd(ctx, indexKey, checkpoint.Step)
	_, err = pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to save checkpoint: %w", err)
	}

	return nil
}

// LoadLatestCheckpoint loads the most recent checkpoint for a workflow
func (s *RedisWorkflowStore) LoadLatestCheckpoint(ctx context.Context, workflowID string) (*Checkpoint, error) {
	checkpoints, err := s.ListCheckpoints(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	if len(checkpoints) == 0 {
		return nil, nil
	}

	// Sort by UpdatedAt and return the most recent
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i].UpdatedAt.After(checkpoints[j].UpdatedAt)
	})

	return checkpoints[0], nil
}

// LoadCheckpoint loads a specific checkpoint by workflow and step
func (s *RedisWorkflowStore) LoadCheckpoint(ctx context.Context, workflowID, step string) (*Checkpoint, error) {
	key := fmt.Sprintf(checkpointKeyPattern, s.prefix, workflowID, step)

	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load checkpoint: %w", err)
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// ListCheckpoints lists all checkpoints for a workflow
func (s *RedisWorkflowStore) ListCheckpoints(ctx context.Context, workflowID string) ([]*Checkpoint, error) {
	indexKey := fmt.Sprintf(checkpointIndexKey, s.prefix, workflowID)

	steps, err := s.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list checkpoint steps: %w", err)
	}

	if len(steps) == 0 {
		return []*Checkpoint{}, nil
	}

	// Fetch all checkpoints using pipeline
	pipe := s.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(steps))
	for i, step := range steps {
		key := fmt.Sprintf(checkpointKeyPattern, s.prefix, workflowID, step)
		cmds[i] = pipe.Get(ctx, key)
	}
	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to fetch checkpoints: %w", err)
	}

	var results []*Checkpoint
	for _, cmd := range cmds {
		data, err := cmd.Bytes()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			continue
		}

		var checkpoint Checkpoint
		if err := json.Unmarshal(data, &checkpoint); err != nil {
			continue
		}

		results = append(results, &checkpoint)
	}

	// Sort by creation time
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.Before(results[j].CreatedAt)
	})

	return results, nil
}

// DeleteCheckpoints deletes all checkpoints for a workflow
func (s *RedisWorkflowStore) DeleteCheckpoints(ctx context.Context, workflowID string) error {
	indexKey := fmt.Sprintf(checkpointIndexKey, s.prefix, workflowID)

	steps, err := s.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return fmt.Errorf("failed to list checkpoint steps: %w", err)
	}

	if len(steps) == 0 {
		return nil
	}

	// Delete all checkpoint keys and the index
	pipe := s.client.Pipeline()
	for _, step := range steps {
		key := fmt.Sprintf(checkpointKeyPattern, s.prefix, workflowID, step)
		pipe.Del(ctx, key)
	}
	pipe.Del(ctx, indexKey)
	_, err = pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete checkpoints: %w", err)
	}

	return nil
}

// AppendHistory appends an event to workflow history
func (s *RedisWorkflowStore) AppendHistory(ctx context.Context, workflowID string, event *HistoryEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal history event: %w", err)
	}

	key := fmt.Sprintf(historyKeyPattern, s.prefix, workflowID)

	pipe := s.client.Pipeline()
	pipe.RPush(ctx, key, data)

	// Enforce max history limit using LTRIM
	if s.config.MaxHistoryPerWorkflow > 0 {
		pipe.LTrim(ctx, key, int64(-s.config.MaxHistoryPerWorkflow), -1)
	}

	// Set TTL on history if configured
	if s.config.HistoryTTL > 0 {
		pipe.Expire(ctx, key, s.config.HistoryTTL)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to append history: %w", err)
	}

	return nil
}

// GetHistory retrieves the complete history for a workflow
func (s *RedisWorkflowStore) GetHistory(ctx context.Context, workflowID string) ([]*HistoryEvent, error) {
	key := fmt.Sprintf(historyKeyPattern, s.prefix, workflowID)

	data, err := s.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}

	var results []*HistoryEvent
	for _, item := range data {
		var event HistoryEvent
		if err := json.Unmarshal([]byte(item), &event); err != nil {
			continue
		}
		results = append(results, &event)
	}

	return results, nil
}

// DeleteHistory deletes all history for a workflow
func (s *RedisWorkflowStore) DeleteHistory(ctx context.Context, workflowID string) error {
	key := fmt.Sprintf(historyKeyPattern, s.prefix, workflowID)

	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete history: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (s *RedisWorkflowStore) Close() error {
	return s.client.Close()
}

// Ping checks if the Redis connection is alive
func (s *RedisWorkflowStore) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

// GetRunningWorkflows returns all workflows currently in running status
// This is useful for recovery on restart
func (s *RedisWorkflowStore) GetRunningWorkflows(ctx context.Context) ([]*WorkflowState, error) {
	return s.ListStates(ctx, StateFilter{Status: WorkflowStatusRunning})
}

// GetPausedWorkflows returns all workflows currently in paused status
func (s *RedisWorkflowStore) GetPausedWorkflows(ctx context.Context) ([]*WorkflowState, error) {
	return s.ListStates(ctx, StateFilter{Status: WorkflowStatusPaused})
}

// CleanupCompletedWorkflows removes completed workflows older than the configured TTL
func (s *RedisWorkflowStore) CleanupCompletedWorkflows(ctx context.Context) (int, error) {
	if s.config.StateTTL == 0 {
		return 0, nil // No TTL configured
	}

	indexKey := fmt.Sprintf(stateIndexKey, s.prefix)
	cutoff := time.Now().Add(-s.config.StateTTL).UnixNano()

	// Get old workflow IDs
	ids, err := s.client.ZRangeByScore(ctx, indexKey, &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%d", cutoff),
	}).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to find old workflows: %w", err)
	}

	deleted := 0
	for _, id := range ids {
		state, err := s.LoadState(ctx, id)
		if err != nil || state == nil {
			continue
		}

		// Only delete terminal workflows
		if isTerminalStatus(state.Status) {
			if err := s.deleteWorkflowFully(ctx, id); err == nil {
				deleted++
			}
		}
	}

	return deleted, nil
}

// deleteWorkflowFully removes a workflow and all its associated data
func (s *RedisWorkflowStore) deleteWorkflowFully(ctx context.Context, id string) error {
	pipe := s.client.Pipeline()

	// Delete state
	stateKey := fmt.Sprintf(stateKeyPattern, s.prefix, id)
	pipe.Del(ctx, stateKey)

	// Remove from index
	indexKey := fmt.Sprintf(stateIndexKey, s.prefix)
	pipe.ZRem(ctx, indexKey, id)

	// Delete history
	historyKey := fmt.Sprintf(historyKeyPattern, s.prefix, id)
	pipe.Del(ctx, historyKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	// Delete checkpoints separately (needs to fetch step list first)
	return s.DeleteCheckpoints(ctx, id)
}

// Helper functions

func isTerminalStatus(status WorkflowStatus) bool {
	return status == WorkflowStatusCompleted ||
		status == WorkflowStatusFailed ||
		status == WorkflowStatusCancelled
}

func matchesStateFilter(state *WorkflowState, filter StateFilter) bool {
	if filter.Status != "" && state.Status != filter.Status {
		return false
	}
	if filter.Name != "" && !strings.Contains(state.Name, filter.Name) {
		return false
	}
	if filter.StartedAfter != nil && state.StartedAt.Before(*filter.StartedAfter) {
		return false
	}
	if filter.StartedBefore != nil && state.StartedAt.After(*filter.StartedBefore) {
		return false
	}
	return true
}
