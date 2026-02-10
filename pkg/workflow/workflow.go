// Package workflow provides idempotent handlers, checkpoints, and state machines
// for operational hardening.
//
// VE-709: Operational Hardening
package workflow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// State represents a workflow state
type State string

// Transition represents a state transition
type Transition struct {
	From   State
	To     State
	Event  string
	Guard  func(ctx context.Context, data interface{}) bool
	Action func(ctx context.Context, data interface{}) error
}

// StateMachine manages state transitions
type StateMachine struct {
	mu          sync.RWMutex
	name        string
	states      map[State]bool
	initial     State
	current     State
	transitions []Transition
	data        interface{}
	history     []StateChange
}

// StateChange records a state change
type StateChange struct {
	From      State
	To        State
	Event     string
	Timestamp time.Time
}

// NewStateMachine creates a new state machine
func NewStateMachine(name string, initial State) *StateMachine {
	return &StateMachine{
		name:    name,
		states:  make(map[State]bool),
		initial: initial,
		current: initial,
		history: make([]StateChange, 0),
	}
}

// AddState adds a valid state
func (sm *StateMachine) AddState(state State) *StateMachine {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.states[state] = true
	return sm
}

// AddStates adds multiple valid states
func (sm *StateMachine) AddStates(states ...State) *StateMachine {
	for _, state := range states {
		sm.AddState(state)
	}
	return sm
}

// AddTransition adds a state transition
func (sm *StateMachine) AddTransition(t Transition) *StateMachine {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.transitions = append(sm.transitions, t)
	return sm
}

// SetData sets the workflow data
func (sm *StateMachine) SetData(data interface{}) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.data = data
}

// Current returns the current state
func (sm *StateMachine) Current() State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current
}

// Can checks if an event can trigger a transition
func (sm *StateMachine) Can(event string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, t := range sm.transitions {
		if t.From == sm.current && t.Event == event {
			if t.Guard == nil || t.Guard(context.Background(), sm.data) {
				return true
			}
		}
	}
	return false
}

// Fire triggers an event
func (sm *StateMachine) Fire(ctx context.Context, event string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, t := range sm.transitions {
		if t.From == sm.current && t.Event == event {
			if t.Guard != nil && !t.Guard(ctx, sm.data) {
				continue
			}

			// Execute action if present
			if t.Action != nil {
				if err := t.Action(ctx, sm.data); err != nil {
					return fmt.Errorf("transition action failed: %w", err)
				}
			}

			// Record state change
			sm.history = append(sm.history, StateChange{
				From:      sm.current,
				To:        t.To,
				Event:     event,
				Timestamp: time.Now(),
			})

			sm.current = t.To
			return nil
		}
	}

	return fmt.Errorf("no valid transition from state %s with event %s", sm.current, event)
}

// History returns the state change history
func (sm *StateMachine) History() []StateChange {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	result := make([]StateChange, len(sm.history))
	copy(result, sm.history)
	return result
}

// Reset resets the state machine to initial state
func (sm *StateMachine) Reset() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.current = sm.initial
	sm.history = make([]StateChange, 0)
}

// IdempotencyKey generates an idempotency key
type IdempotencyKey string

// GenerateIdempotencyKey generates a key from input data
func GenerateIdempotencyKey(data ...interface{}) IdempotencyKey {
	hasher := sha256.New()
	for _, d := range data {
		_, _ = fmt.Fprintf(hasher, "%v", d)
	}
	return IdempotencyKey(hex.EncodeToString(hasher.Sum(nil))[:32])
}

// IdempotentHandler wraps a handler with idempotency checking
type IdempotentHandler struct {
	mu       sync.RWMutex
	store    IdempotencyStore
	ttl      time.Duration
	onReplay func(key IdempotencyKey, result interface{})
}

// IdempotencyStore stores idempotency records
type IdempotencyStore interface {
	// Get retrieves a stored result
	Get(ctx context.Context, key IdempotencyKey) (interface{}, bool, error)

	// Set stores a result
	Set(ctx context.Context, key IdempotencyKey, result interface{}, ttl time.Duration) error

	// Delete removes a stored result
	Delete(ctx context.Context, key IdempotencyKey) error

	// SetPending marks a key as pending (to prevent concurrent processing)
	SetPending(ctx context.Context, key IdempotencyKey, ttl time.Duration) (bool, error)
}

// NewIdempotentHandler creates a new idempotent handler
func NewIdempotentHandler(store IdempotencyStore, ttl time.Duration) *IdempotentHandler {
	return &IdempotentHandler{
		store: store,
		ttl:   ttl,
	}
}

// OnReplay sets a callback for replayed results
func (h *IdempotentHandler) OnReplay(fn func(key IdempotencyKey, result interface{})) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onReplay = fn
}

// Execute executes a handler with idempotency
func (h *IdempotentHandler) Execute(
	ctx context.Context,
	key IdempotencyKey,
	handler func(ctx context.Context) (interface{}, error),
) (interface{}, error) {
	// Check for existing result
	if result, found, err := h.store.Get(ctx, key); err != nil {
		return nil, fmt.Errorf("failed to check idempotency store: %w", err)
	} else if found {
		h.mu.RLock()
		onReplay := h.onReplay
		h.mu.RUnlock()
		if onReplay != nil {
			onReplay(key, result)
		}
		return result, nil
	}

	// Try to set pending
	acquired, err := h.store.SetPending(ctx, key, h.ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to set pending: %w", err)
	}
	if !acquired {
		// Another process is handling this key
		return nil, ErrProcessingInProgress
	}

	// Execute the handler
	result, err := handler(ctx)
	if err != nil {
		// Delete pending marker on error
		_ = h.store.Delete(ctx, key)
		return nil, err
	}

	// Store the result
	if storeErr := h.store.Set(ctx, key, result, h.ttl); storeErr != nil {
		// Log but don't fail - the operation succeeded
		return result, nil
	}

	return result, nil
}

// ErrProcessingInProgress indicates another process is handling the key
var ErrProcessingInProgress = fmt.Errorf("request is being processed by another handler")

// InMemoryIdempotencyStore is an in-memory implementation for testing
type InMemoryIdempotencyStore struct {
	mu      sync.RWMutex
	results map[IdempotencyKey]storeEntry
	pending map[IdempotencyKey]time.Time
}

type storeEntry struct {
	result    interface{}
	expiresAt time.Time
}

// NewInMemoryIdempotencyStore creates a new in-memory store
func NewInMemoryIdempotencyStore() *InMemoryIdempotencyStore {
	store := &InMemoryIdempotencyStore{
		results: make(map[IdempotencyKey]storeEntry),
		pending: make(map[IdempotencyKey]time.Time),
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

func (s *InMemoryIdempotencyStore) Get(ctx context.Context, key IdempotencyKey) (interface{}, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, found := s.results[key]
	if !found {
		return nil, false, nil
	}

	if time.Now().After(entry.expiresAt) {
		return nil, false, nil
	}

	return entry.result, true, nil
}

func (s *InMemoryIdempotencyStore) Set(ctx context.Context, key IdempotencyKey, result interface{}, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.results[key] = storeEntry{
		result:    result,
		expiresAt: time.Now().Add(ttl),
	}

	delete(s.pending, key)
	return nil
}

func (s *InMemoryIdempotencyStore) Delete(ctx context.Context, key IdempotencyKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.results, key)
	delete(s.pending, key)
	return nil
}

func (s *InMemoryIdempotencyStore) SetPending(ctx context.Context, key IdempotencyKey, ttl time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if expires, found := s.pending[key]; found {
		if time.Now().Before(expires) {
			return false, nil
		}
	}

	s.pending[key] = time.Now().Add(ttl)
	return true, nil
}

func (s *InMemoryIdempotencyStore) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()

		for key, entry := range s.results {
			if now.After(entry.expiresAt) {
				delete(s.results, key)
			}
		}

		for key, expires := range s.pending {
			if now.After(expires) {
				delete(s.pending, key)
			}
		}
		s.mu.Unlock()
	}
}

// Checkpoint represents a workflow checkpoint
type Checkpoint struct {
	ID         string
	WorkflowID string
	Step       string
	Data       interface{}
	Status     CheckpointStatus
	Error      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// CheckpointStatus represents checkpoint status
type CheckpointStatus string

const (
	CheckpointPending   CheckpointStatus = "pending"
	CheckpointCompleted CheckpointStatus = "completed"
	CheckpointFailed    CheckpointStatus = "failed"
	CheckpointSkipped   CheckpointStatus = "skipped"
)

// CheckpointStore stores checkpoints
type CheckpointStore interface {
	// Save saves a checkpoint
	Save(ctx context.Context, cp *Checkpoint) error

	// Get retrieves a checkpoint
	Get(ctx context.Context, workflowID, step string) (*Checkpoint, error)

	// List lists checkpoints for a workflow
	List(ctx context.Context, workflowID string) ([]*Checkpoint, error)

	// Delete deletes checkpoints for a workflow
	Delete(ctx context.Context, workflowID string) error
}

// WorkflowRunner runs a workflow with checkpoints
type WorkflowRunner struct {
	store   CheckpointStore
	steps   []WorkflowStep
	onError func(ctx context.Context, step string, err error) error
}

// WorkflowStep represents a step in a workflow
type WorkflowStep struct {
	Name       string
	Execute    func(ctx context.Context, data interface{}) (interface{}, error)
	Compensate func(ctx context.Context, data interface{}) error
	CanRetry   bool
	MaxRetries int
}

// NewWorkflowRunner creates a new workflow runner
func NewWorkflowRunner(store CheckpointStore) *WorkflowRunner {
	return &WorkflowRunner{
		store: store,
		steps: make([]WorkflowStep, 0),
	}
}

// AddStep adds a step to the workflow
func (r *WorkflowRunner) AddStep(step WorkflowStep) *WorkflowRunner {
	r.steps = append(r.steps, step)
	return r
}

// OnError sets the error handler
func (r *WorkflowRunner) OnError(fn func(ctx context.Context, step string, err error) error) *WorkflowRunner {
	r.onError = fn
	return r
}

// Run executes the workflow
func (r *WorkflowRunner) Run(ctx context.Context, workflowID string, initialData interface{}) error {
	data := initialData
	completedSteps := make([]string, 0)

	for _, step := range r.steps {
		// Check for existing checkpoint
		cp, err := r.store.Get(ctx, workflowID, step.Name)
		if err == nil && cp != nil && cp.Status == CheckpointCompleted {
			// Step already completed, use stored data
			data = cp.Data
			completedSteps = append(completedSteps, step.Name)
			continue
		}

		// Create pending checkpoint
		cp = &Checkpoint{
			ID:         fmt.Sprintf("%s-%s", workflowID, step.Name),
			WorkflowID: workflowID,
			Step:       step.Name,
			Status:     CheckpointPending,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := r.store.Save(ctx, cp); err != nil {
			return fmt.Errorf("failed to save checkpoint: %w", err)
		}

		// Execute step
		result, err := r.executeStep(ctx, step, data)
		if err != nil {
			// Mark checkpoint as failed
			cp.Status = CheckpointFailed
			cp.Error = err.Error()
			cp.UpdatedAt = time.Now()
			_ = r.store.Save(ctx, cp)

			// Handle error
			if r.onError != nil {
				if handleErr := r.onError(ctx, step.Name, err); handleErr != nil {
					// Compensate completed steps in reverse order
					for i := len(completedSteps) - 1; i >= 0; i-- {
						stepName := completedSteps[i]
						for _, s := range r.steps {
							if s.Name == stepName && s.Compensate != nil {
								_ = s.Compensate(ctx, data)
							}
						}
					}
					return handleErr
				}
			} else {
				return err
			}
		}

		// Update checkpoint with result
		cp.Status = CheckpointCompleted
		cp.Data = result
		cp.UpdatedAt = time.Now()
		if err := r.store.Save(ctx, cp); err != nil {
			return fmt.Errorf("failed to update checkpoint: %w", err)
		}

		data = result
		completedSteps = append(completedSteps, step.Name)
	}

	return nil
}

func (r *WorkflowRunner) executeStep(ctx context.Context, step WorkflowStep, data interface{}) (interface{}, error) {
	var lastErr error
	maxAttempts := 1
	if step.CanRetry && step.MaxRetries > 0 {
		maxAttempts = step.MaxRetries + 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		result, err := step.Execute(ctx, data)
		if err == nil {
			return result, nil
		}

		lastErr = err

		if !step.CanRetry {
			break
		}

		// Exponential backoff
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		//nolint:gosec // G115: attempt is bounded by maxAttempts configuration
		case <-time.After(time.Duration(1<<uint(attempt)) * 100 * time.Millisecond):
		}
	}

	return nil, lastErr
}

// InMemoryCheckpointStore is an in-memory implementation for testing
type InMemoryCheckpointStore struct {
	mu          sync.RWMutex
	checkpoints map[string]*Checkpoint
}

// NewInMemoryCheckpointStore creates a new in-memory checkpoint store
func NewInMemoryCheckpointStore() *InMemoryCheckpointStore {
	return &InMemoryCheckpointStore{
		checkpoints: make(map[string]*Checkpoint),
	}
}

func (s *InMemoryCheckpointStore) Save(ctx context.Context, cp *Checkpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkpoints[cp.ID] = cp
	return nil
}

func (s *InMemoryCheckpointStore) Get(ctx context.Context, workflowID, step string) (*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id := fmt.Sprintf("%s-%s", workflowID, step)
	return s.checkpoints[id], nil
}

func (s *InMemoryCheckpointStore) List(ctx context.Context, workflowID string) ([]*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Checkpoint
	for _, cp := range s.checkpoints {
		if cp.WorkflowID == workflowID {
			result = append(result, cp)
		}
	}
	return result, nil
}

func (s *InMemoryCheckpointStore) Delete(ctx context.Context, workflowID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, cp := range s.checkpoints {
		if cp.WorkflowID == workflowID {
			delete(s.checkpoints, id)
		}
	}
	return nil
}
