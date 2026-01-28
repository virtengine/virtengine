// Package workflow provides in-memory workflow state storage for testing
package workflow

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// MemoryWorkflowStore is an in-memory implementation of WorkflowStore for testing
type MemoryWorkflowStore struct {
	mu          sync.RWMutex
	states      map[string]*WorkflowState
	checkpoints map[string]map[string]*Checkpoint // workflowID -> step -> checkpoint
	history     map[string][]*HistoryEvent        // workflowID -> events
	config      WorkflowStoreConfig
	closed      bool
}

// NewMemoryWorkflowStore creates a new in-memory workflow store
func NewMemoryWorkflowStore(config WorkflowStoreConfig) *MemoryWorkflowStore {
	store := &MemoryWorkflowStore{
		states:      make(map[string]*WorkflowState),
		checkpoints: make(map[string]map[string]*Checkpoint),
		history:     make(map[string][]*HistoryEvent),
		config:      config,
	}

	// Start cleanup goroutine for TTL expiration
	if config.StateTTL > 0 || config.HistoryTTL > 0 {
		go store.cleanup()
	}

	return store
}

// SaveState saves or updates a workflow state
func (s *MemoryWorkflowStore) SaveState(ctx context.Context, id string, state *WorkflowState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("store is closed")
	}

	// Deep copy to prevent external mutation
	stateCopy := *state
	stateCopy.Data = copyMap(state.Data)
	stateCopy.Metadata = copyStringMap(state.Metadata)
	stateCopy.CompletedSteps = copyStringSlice(state.CompletedSteps)

	s.states[id] = &stateCopy
	return nil
}

// LoadState loads a workflow state by ID
func (s *MemoryWorkflowStore) LoadState(ctx context.Context, id string) (*WorkflowState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, fmt.Errorf("store is closed")
	}

	state, ok := s.states[id]
	if !ok {
		return nil, nil
	}

	// Deep copy to prevent external mutation
	stateCopy := *state
	stateCopy.Data = copyMap(state.Data)
	stateCopy.Metadata = copyStringMap(state.Metadata)
	stateCopy.CompletedSteps = copyStringSlice(state.CompletedSteps)

	return &stateCopy, nil
}

// DeleteState deletes a workflow state
func (s *MemoryWorkflowStore) DeleteState(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("store is closed")
	}

	delete(s.states, id)
	return nil
}

// ListStates lists workflow states matching the filter
func (s *MemoryWorkflowStore) ListStates(ctx context.Context, filter StateFilter) ([]*WorkflowState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, fmt.Errorf("store is closed")
	}

	var results []*WorkflowState
	for _, state := range s.states {
		if matchesFilter(state, filter) {
			// Deep copy
			stateCopy := *state
			stateCopy.Data = copyMap(state.Data)
			stateCopy.Metadata = copyStringMap(state.Metadata)
			stateCopy.CompletedSteps = copyStringSlice(state.CompletedSteps)
			results = append(results, &stateCopy)
		}
	}

	// Sort by started time (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].StartedAt.After(results[j].StartedAt)
	})

	// Apply pagination
	if filter.Offset > 0 {
		if filter.Offset >= len(results) {
			return []*WorkflowState{}, nil
		}
		results = results[filter.Offset:]
	}
	if filter.Limit > 0 && len(results) > filter.Limit {
		results = results[:filter.Limit]
	}

	return results, nil
}

// SaveCheckpoint saves a workflow checkpoint
func (s *MemoryWorkflowStore) SaveCheckpoint(ctx context.Context, workflowID string, checkpoint *Checkpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("store is closed")
	}

	if s.checkpoints[workflowID] == nil {
		s.checkpoints[workflowID] = make(map[string]*Checkpoint)
	}

	// Deep copy
	cpCopy := *checkpoint
	s.checkpoints[workflowID][checkpoint.Step] = &cpCopy

	return nil
}

// LoadLatestCheckpoint loads the most recent checkpoint for a workflow
func (s *MemoryWorkflowStore) LoadLatestCheckpoint(ctx context.Context, workflowID string) (*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, fmt.Errorf("store is closed")
	}

	steps := s.checkpoints[workflowID]
	if len(steps) == 0 {
		return nil, nil
	}

	// Find the most recently updated checkpoint
	var latest *Checkpoint
	for _, cp := range steps {
		if latest == nil || cp.UpdatedAt.After(latest.UpdatedAt) {
			latest = cp
		}
	}

	if latest != nil {
		cpCopy := *latest
		return &cpCopy, nil
	}

	return nil, nil
}

// LoadCheckpoint loads a specific checkpoint by workflow and step
func (s *MemoryWorkflowStore) LoadCheckpoint(ctx context.Context, workflowID, step string) (*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, fmt.Errorf("store is closed")
	}

	steps := s.checkpoints[workflowID]
	if steps == nil {
		return nil, nil
	}

	cp := steps[step]
	if cp == nil {
		return nil, nil
	}

	cpCopy := *cp
	return &cpCopy, nil
}

// ListCheckpoints lists all checkpoints for a workflow
func (s *MemoryWorkflowStore) ListCheckpoints(ctx context.Context, workflowID string) ([]*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, fmt.Errorf("store is closed")
	}

	steps := s.checkpoints[workflowID]
	if steps == nil {
		return []*Checkpoint{}, nil
	}

	var results []*Checkpoint
	for _, cp := range steps {
		cpCopy := *cp
		results = append(results, &cpCopy)
	}

	// Sort by creation time
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.Before(results[j].CreatedAt)
	})

	return results, nil
}

// DeleteCheckpoints deletes all checkpoints for a workflow
func (s *MemoryWorkflowStore) DeleteCheckpoints(ctx context.Context, workflowID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("store is closed")
	}

	delete(s.checkpoints, workflowID)
	return nil
}

// AppendHistory appends an event to workflow history
func (s *MemoryWorkflowStore) AppendHistory(ctx context.Context, workflowID string, event *HistoryEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("store is closed")
	}

	// Deep copy
	eventCopy := *event

	s.history[workflowID] = append(s.history[workflowID], &eventCopy)

	// Enforce max history limit
	if s.config.MaxHistoryPerWorkflow > 0 && len(s.history[workflowID]) > s.config.MaxHistoryPerWorkflow {
		// Trim oldest events
		excess := len(s.history[workflowID]) - s.config.MaxHistoryPerWorkflow
		s.history[workflowID] = s.history[workflowID][excess:]
	}

	return nil
}

// GetHistory retrieves the complete history for a workflow
func (s *MemoryWorkflowStore) GetHistory(ctx context.Context, workflowID string) ([]*HistoryEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil, fmt.Errorf("store is closed")
	}

	events := s.history[workflowID]
	if events == nil {
		return []*HistoryEvent{}, nil
	}

	// Deep copy
	results := make([]*HistoryEvent, len(events))
	for i, event := range events {
		eventCopy := *event
		results[i] = &eventCopy
	}

	return results, nil
}

// DeleteHistory deletes all history for a workflow
func (s *MemoryWorkflowStore) DeleteHistory(ctx context.Context, workflowID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("store is closed")
	}

	delete(s.history, workflowID)
	return nil
}

// Close closes the store
func (s *MemoryWorkflowStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true
	return nil
}

// cleanup runs periodically to remove expired entries
func (s *MemoryWorkflowStore) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			return
		}

		now := time.Now()

		// Cleanup expired states
		if s.config.StateTTL > 0 {
			for id, state := range s.states {
				// Only delete completed/failed/cancelled workflows past TTL
				if state.Status == WorkflowStatusCompleted ||
					state.Status == WorkflowStatusFailed ||
					state.Status == WorkflowStatusCancelled {
					if state.CompletedAt != nil && now.Sub(*state.CompletedAt) > s.config.StateTTL {
						delete(s.states, id)
						delete(s.checkpoints, id)
						delete(s.history, id)
					}
				}
			}
		}

		// Cleanup expired history events
		if s.config.HistoryTTL > 0 {
			for workflowID, events := range s.history {
				var filtered []*HistoryEvent
				for _, event := range events {
					if now.Sub(event.Timestamp) <= s.config.HistoryTTL {
						filtered = append(filtered, event)
					}
				}
				if len(filtered) == 0 {
					delete(s.history, workflowID)
				} else {
					s.history[workflowID] = filtered
				}
			}
		}

		s.mu.Unlock()
	}
}

// matchesFilter checks if a workflow state matches the given filter
func matchesFilter(state *WorkflowState, filter StateFilter) bool {
	if filter.Status != "" && state.Status != filter.Status {
		return false
	}
	if filter.Name != "" && state.Name != filter.Name {
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

// Helper functions for deep copying
func copyMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

func copyStringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

func copyStringSlice(s []string) []string {
	if s == nil {
		return nil
	}
	result := make([]string, len(s))
	copy(result, s)
	return result
}
