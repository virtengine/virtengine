// Package workflow provides workflow engine with persistence and recovery
package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"

	verrors "github.com/virtengine/virtengine/pkg/errors"
)

// Engine manages workflow execution with persistence and recovery
type Engine struct {
	mu           sync.RWMutex
	store        WorkflowStore
	definitions  map[string]*WorkflowDefinition
	runners      map[string]*activeWorkflow
	logger       zerolog.Logger
	config       EngineConfig
	recoveryDone chan struct{}
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

// EngineConfig configures the workflow engine
type EngineConfig struct {
	// Store is the configuration for the workflow store
	Store WorkflowStoreConfig `json:"store" yaml:"store"`

	// RecoveryOnStart enables automatic workflow recovery on engine start
	RecoveryOnStart bool `json:"recovery_on_start" yaml:"recovery_on_start"`

	// RecoveryTimeout is the maximum time to wait for recovery to complete
	RecoveryTimeout time.Duration `json:"recovery_timeout" yaml:"recovery_timeout"`

	// MaxConcurrentWorkflows limits concurrent workflow executions (0 = unlimited)
	MaxConcurrentWorkflows int `json:"max_concurrent_workflows" yaml:"max_concurrent_workflows"`

	// DefaultStepTimeout is the default timeout for workflow steps
	DefaultStepTimeout time.Duration `json:"default_step_timeout" yaml:"default_step_timeout"`
}

// DefaultEngineConfig returns default engine configuration
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		Store:                  DefaultWorkflowStoreConfig(),
		RecoveryOnStart:        true,
		RecoveryTimeout:        5 * time.Minute,
		MaxConcurrentWorkflows: 100,
		DefaultStepTimeout:     30 * time.Second,
	}
}

// WorkflowDefinition defines a workflow's structure
type WorkflowDefinition struct {
	Name        string
	Steps       []WorkflowStep
	OnError     func(ctx context.Context, step string, err error) error
	OnComplete  func(ctx context.Context, state *WorkflowState) error
	MaxRetries  int
	StepTimeout time.Duration
}

// activeWorkflow tracks an in-progress workflow
type activeWorkflow struct {
	state      *WorkflowState
	definition *WorkflowDefinition
	cancelFn   context.CancelFunc
	done       chan struct{}
}

// NewEngine creates a new workflow engine
func NewEngine(ctx context.Context, config EngineConfig, logger zerolog.Logger) (*Engine, error) {
	store, err := NewWorkflowStore(ctx, config.Store)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow store: %w", err)
	}

	e := &Engine{
		store:        store,
		definitions:  make(map[string]*WorkflowDefinition),
		runners:      make(map[string]*activeWorkflow),
		logger:       logger.With().Str("component", "workflow-engine").Logger(),
		config:       config,
		recoveryDone: make(chan struct{}),
		stopCh:       make(chan struct{}),
	}

	if config.RecoveryOnStart {
		verrors.SafeGo("workflow:recovery", func() {
			e.runRecovery(ctx)
		})
	} else {
		close(e.recoveryDone)
	}

	return e, nil
}

// RegisterWorkflow registers a workflow definition
func (e *Engine) RegisterWorkflow(def *WorkflowDefinition) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if def.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(def.Steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}

	e.definitions[def.Name] = def
	e.logger.Info().Str("workflow", def.Name).Int("steps", len(def.Steps)).Msg("registered workflow definition")
	return nil
}

// StartWorkflow starts a new workflow execution
func (e *Engine) StartWorkflow(ctx context.Context, workflowName, workflowID string, data map[string]interface{}) (*WorkflowState, error) {
	e.mu.RLock()
	def, ok := e.definitions[workflowName]
	e.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("workflow '%s' not registered", workflowName)
	}

	// Check concurrency limit
	e.mu.RLock()
	running := len(e.runners)
	e.mu.RUnlock()

	if e.config.MaxConcurrentWorkflows > 0 && running >= e.config.MaxConcurrentWorkflows {
		return nil, fmt.Errorf("max concurrent workflows (%d) reached", e.config.MaxConcurrentWorkflows)
	}

	// Create initial state
	state := &WorkflowState{
		ID:             workflowID,
		Name:           workflowName,
		Status:         WorkflowStatusPending,
		Data:           data,
		StartedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CompletedSteps: []string{},
		Metadata:       make(map[string]string),
	}

	// Save initial state
	if err := e.store.SaveState(ctx, workflowID, state); err != nil {
		return nil, fmt.Errorf("failed to save initial state: %w", err)
	}

	// Record workflow started event
	e.appendHistory(ctx, workflowID, &HistoryEvent{
		ID:          fmt.Sprintf("%s-started-%d", workflowID, time.Now().UnixNano()),
		WorkflowID:  workflowID,
		EventType:   HistoryEventWorkflowStarted,
		Status:      WorkflowStatusPending,
		Data:        data,
		Timestamp:   time.Now(),
		Description: fmt.Sprintf("Workflow '%s' started", workflowName),
	})

	// Start execution in background
	e.wg.Add(1)
	go e.runWorkflow(ctx, state, def)

	return state, nil
}

// GetWorkflowState retrieves the current state of a workflow
func (e *Engine) GetWorkflowState(ctx context.Context, workflowID string) (*WorkflowState, error) {
	return e.store.LoadState(ctx, workflowID)
}

// GetWorkflowHistory retrieves the history of a workflow
func (e *Engine) GetWorkflowHistory(ctx context.Context, workflowID string) ([]*HistoryEvent, error) {
	return e.store.GetHistory(ctx, workflowID)
}

// CancelWorkflow cancels a running workflow
func (e *Engine) CancelWorkflow(ctx context.Context, workflowID string) error {
	e.mu.Lock()
	active, ok := e.runners[workflowID]
	e.mu.Unlock()

	if !ok {
		// Check if workflow exists in store
		state, err := e.store.LoadState(ctx, workflowID)
		if err != nil {
			return err
		}
		if state == nil {
			return fmt.Errorf("workflow '%s' not found", workflowID)
		}
		if state.Status == WorkflowStatusCompleted || state.Status == WorkflowStatusFailed || state.Status == WorkflowStatusCancelled {
			return fmt.Errorf("workflow '%s' is already in terminal state: %s", workflowID, state.Status)
		}
		// Workflow exists but not actively running - mark as cancelled
		state.Status = WorkflowStatusCancelled
		now := time.Now()
		state.CompletedAt = &now
		state.UpdatedAt = now
		return e.store.SaveState(ctx, workflowID, state)
	}

	// Cancel active workflow
	active.cancelFn()
	<-active.done // Wait for it to finish

	return nil
}

// PauseWorkflow pauses a running workflow
func (e *Engine) PauseWorkflow(ctx context.Context, workflowID string) error {
	state, err := e.store.LoadState(ctx, workflowID)
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("workflow '%s' not found", workflowID)
	}
	if state.Status != WorkflowStatusRunning {
		return fmt.Errorf("can only pause running workflows (current: %s)", state.Status)
	}

	state.Status = WorkflowStatusPaused
	state.UpdatedAt = time.Now()

	if err := e.store.SaveState(ctx, workflowID, state); err != nil {
		return err
	}

	e.appendHistory(ctx, workflowID, &HistoryEvent{
		ID:          fmt.Sprintf("%s-paused-%d", workflowID, time.Now().UnixNano()),
		WorkflowID:  workflowID,
		EventType:   HistoryEventWorkflowPaused,
		Status:      WorkflowStatusPaused,
		Timestamp:   time.Now(),
		Description: "Workflow paused",
	})

	return nil
}

// ResumeWorkflow resumes a paused workflow
func (e *Engine) ResumeWorkflow(ctx context.Context, workflowID string) error {
	state, err := e.store.LoadState(ctx, workflowID)
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("workflow '%s' not found", workflowID)
	}
	if state.Status != WorkflowStatusPaused {
		return fmt.Errorf("can only resume paused workflows (current: %s)", state.Status)
	}

	e.mu.RLock()
	def, ok := e.definitions[state.Name]
	e.mu.RUnlock()

	if !ok {
		return fmt.Errorf("workflow definition '%s' not registered", state.Name)
	}

	state.Status = WorkflowStatusRunning
	state.UpdatedAt = time.Now()

	if err := e.store.SaveState(ctx, workflowID, state); err != nil {
		return err
	}

	e.appendHistory(ctx, workflowID, &HistoryEvent{
		ID:          fmt.Sprintf("%s-resumed-%d", workflowID, time.Now().UnixNano()),
		WorkflowID:  workflowID,
		EventType:   HistoryEventWorkflowResumed,
		Status:      WorkflowStatusRunning,
		Timestamp:   time.Now(),
		Description: "Workflow resumed",
	})

	// Resume execution
	e.wg.Add(1)
	go e.runWorkflow(ctx, state, def)

	return nil
}

// RecoverWorkflows recovers all interrupted workflows from persistent storage
func (e *Engine) RecoverWorkflows(ctx context.Context) error {
	e.logger.Info().Msg("starting workflow recovery")
	startTime := time.Now()

	// Find all running/paused workflows
	runningStates, err := e.store.ListStates(ctx, StateFilter{Status: WorkflowStatusRunning})
	if err != nil {
		return fmt.Errorf("failed to list running workflows: %w", err)
	}

	pausedStates, err := e.store.ListStates(ctx, StateFilter{Status: WorkflowStatusPaused})
	if err != nil {
		return fmt.Errorf("failed to list paused workflows: %w", err)
	}

	// Also recover pending workflows that never started
	pendingStates, err := e.store.ListStates(ctx, StateFilter{Status: WorkflowStatusPending})
	if err != nil {
		return fmt.Errorf("failed to list pending workflows: %w", err)
	}

	// Combine running and pending - these need immediate recovery
	toRecover := append(runningStates, pendingStates...)

	recovered := 0
	failed := 0

	for _, state := range toRecover {
		e.mu.RLock()
		def, ok := e.definitions[state.Name]
		e.mu.RUnlock()

		if !ok {
			e.logger.Warn().
				Str("workflow_id", state.ID).
				Str("workflow_name", state.Name).
				Msg("cannot recover workflow: definition not registered")
			failed++
			continue
		}

		// Load checkpoint for this workflow
		checkpoint, err := e.store.LoadLatestCheckpoint(ctx, state.ID)
		if err != nil {
			e.logger.Warn().
				Err(err).
				Str("workflow_id", state.ID).
				Msg("failed to load checkpoint, starting from last known state")
		}

		e.appendHistory(ctx, state.ID, &HistoryEvent{
			ID:          fmt.Sprintf("%s-recovery-%d", state.ID, time.Now().UnixNano()),
			WorkflowID:  state.ID,
			EventType:   HistoryEventRecoveryStarted,
			Step:        state.CurrentStep,
			Status:      state.Status,
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Recovery started from step '%s'", state.CurrentStep),
		})

		if checkpoint != nil {
			e.logger.Info().
				Str("workflow_id", state.ID).
				Str("step", checkpoint.Step).
				Time("checkpoint_time", checkpoint.UpdatedAt).
				Msg("resuming workflow from checkpoint")
		}

		// Resume the workflow
		e.wg.Add(1)
		go e.runWorkflow(ctx, state, def)
		recovered++
	}

	e.logger.Info().
		Int("running_found", len(runningStates)).
		Int("pending_found", len(pendingStates)).
		Int("paused_found", len(pausedStates)).
		Int("recovered", recovered).
		Int("failed", failed).
		Dur("duration", time.Since(startTime)).
		Msg("workflow recovery complete")

	return nil
}

// WaitForRecovery waits for initial recovery to complete
func (e *Engine) WaitForRecovery() {
	<-e.recoveryDone
}

// Stop stops the engine and waits for all workflows to finish
func (e *Engine) Stop(ctx context.Context) error {
	close(e.stopCh)

	// Cancel all running workflows
	e.mu.Lock()
	for _, active := range e.runners {
		active.cancelFn()
	}
	e.mu.Unlock()

	// Wait for all workflows to complete with panic recovery
	done := make(chan struct{})
	verrors.SafeGo("workflow:shutdown-wait", func() {
		e.wg.Wait()
		close(done)
	})

	select {
	case <-done:
		// All workflows completed
	case <-ctx.Done():
		return ctx.Err()
	}

	// Close the store
	return e.store.Close()
}

// ListActiveWorkflows returns IDs of currently executing workflows
func (e *Engine) ListActiveWorkflows() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ids := make([]string, 0, len(e.runners))
	for id := range e.runners {
		ids = append(ids, id)
	}
	return ids
}

// runRecovery runs the recovery process on startup
func (e *Engine) runRecovery(ctx context.Context) {
	defer close(e.recoveryDone)

	recoveryCtx, cancel := context.WithTimeout(ctx, e.config.RecoveryTimeout)
	defer cancel()

	if err := e.RecoverWorkflows(recoveryCtx); err != nil {
		e.logger.Error().Err(err).Msg("workflow recovery failed")
	}
}

// runWorkflow executes a workflow
func (e *Engine) runWorkflow(ctx context.Context, state *WorkflowState, def *WorkflowDefinition) {
	defer e.wg.Done()

	workflowCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	active := &activeWorkflow{
		state:      state,
		definition: def,
		cancelFn:   cancel,
		done:       make(chan struct{}),
	}

	e.mu.Lock()
	e.runners[state.ID] = active
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		delete(e.runners, state.ID)
		e.mu.Unlock()
		close(active.done)
	}()

	// Update status to running
	state.Status = WorkflowStatusRunning
	state.UpdatedAt = time.Now()
	if err := e.store.SaveState(ctx, state.ID, state); err != nil {
		e.logger.Error().Err(err).Str("workflow_id", state.ID).Msg("failed to update workflow state")
		return
	}

	// Build set of completed steps for quick lookup
	completedSet := make(map[string]bool)
	for _, step := range state.CompletedSteps {
		completedSet[step] = true
	}

	// Execute remaining steps
	for _, step := range def.Steps {
		// Check if already completed
		if completedSet[step.Name] {
			continue
		}

		// Check context cancellation
		select {
		case <-workflowCtx.Done():
			e.handleWorkflowCancellation(ctx, state)
			return
		case <-e.stopCh:
			e.handleWorkflowCancellation(ctx, state)
			return
		default:
		}

		// Check for paused status
		currentState, err := e.store.LoadState(ctx, state.ID)
		if err == nil && currentState != nil && currentState.Status == WorkflowStatusPaused {
			e.logger.Info().Str("workflow_id", state.ID).Msg("workflow paused, stopping execution")
			return
		}

		// Update current step
		state.CurrentStep = step.Name
		state.UpdatedAt = time.Now()
		if err := e.store.SaveState(ctx, state.ID, state); err != nil {
			e.logger.Error().Err(err).Str("workflow_id", state.ID).Msg("failed to save state before step")
		}

		// Create checkpoint
		checkpoint := &Checkpoint{
			ID:         fmt.Sprintf("%s-%s", state.ID, step.Name),
			WorkflowID: state.ID,
			Step:       step.Name,
			Status:     CheckpointPending,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := e.store.SaveCheckpoint(ctx, state.ID, checkpoint); err != nil {
			e.logger.Warn().Err(err).Str("workflow_id", state.ID).Str("step", step.Name).Msg("failed to save checkpoint")
		}

		e.appendHistory(ctx, state.ID, &HistoryEvent{
			ID:          fmt.Sprintf("%s-%s-started-%d", state.ID, step.Name, time.Now().UnixNano()),
			WorkflowID:  state.ID,
			EventType:   HistoryEventStepStarted,
			Step:        step.Name,
			Timestamp:   time.Now(),
			Description: fmt.Sprintf("Step '%s' started", step.Name),
		})

		// Execute step with timeout
		stepTimeout := e.config.DefaultStepTimeout
		if def.StepTimeout > 0 {
			stepTimeout = def.StepTimeout
		}

		stepStart := time.Now()
		result, err := e.executeStepWithRetry(workflowCtx, step, state.Data, def.MaxRetries, stepTimeout)
		stepDuration := time.Since(stepStart)

		if err != nil {
			e.handleStepError(ctx, state, step, checkpoint, def, err, stepDuration)
			return
		}

		// Update checkpoint and state
		checkpoint.Status = CheckpointCompleted
		checkpoint.Data = result
		checkpoint.UpdatedAt = time.Now()
		if err := e.store.SaveCheckpoint(ctx, state.ID, checkpoint); err != nil {
			e.logger.Warn().Err(err).Msg("failed to update checkpoint")
		}

		state.CompletedSteps = append(state.CompletedSteps, step.Name)
		state.UpdatedAt = time.Now()
		if err := e.store.SaveState(ctx, state.ID, state); err != nil {
			e.logger.Error().Err(err).Msg("failed to save state after step")
		}

		e.appendHistory(ctx, state.ID, &HistoryEvent{
			ID:          fmt.Sprintf("%s-%s-completed-%d", state.ID, step.Name, time.Now().UnixNano()),
			WorkflowID:  state.ID,
			EventType:   HistoryEventStepCompleted,
			Step:        step.Name,
			Timestamp:   time.Now(),
			DurationMs:  stepDuration.Milliseconds(),
			Description: fmt.Sprintf("Step '%s' completed in %v", step.Name, stepDuration),
		})
	}

	// Workflow completed successfully
	state.Status = WorkflowStatusCompleted
	now := time.Now()
	state.CompletedAt = &now
	state.UpdatedAt = now
	if err := e.store.SaveState(ctx, state.ID, state); err != nil {
		e.logger.Error().Err(err).Msg("failed to save final state")
	}

	e.appendHistory(ctx, state.ID, &HistoryEvent{
		ID:          fmt.Sprintf("%s-completed-%d", state.ID, time.Now().UnixNano()),
		WorkflowID:  state.ID,
		EventType:   HistoryEventWorkflowCompleted,
		Status:      WorkflowStatusCompleted,
		Timestamp:   time.Now(),
		DurationMs:  time.Since(state.StartedAt).Milliseconds(),
		Description: fmt.Sprintf("Workflow completed in %v", time.Since(state.StartedAt)),
	})

	if def.OnComplete != nil {
		if err := def.OnComplete(ctx, state); err != nil {
			e.logger.Warn().Err(err).Str("workflow_id", state.ID).Msg("OnComplete callback failed")
		}
	}

	e.logger.Info().
		Str("workflow_id", state.ID).
		Dur("duration", time.Since(state.StartedAt)).
		Int("steps", len(state.CompletedSteps)).
		Msg("workflow completed successfully")
}

// executeStepWithRetry executes a step with retry logic
func (e *Engine) executeStepWithRetry(ctx context.Context, step WorkflowStep, data interface{}, maxRetries int, timeout time.Duration) (interface{}, error) {
	var lastErr error
	attempts := 0

	if step.MaxRetries > 0 {
		maxRetries = step.MaxRetries
	}

	for attempts <= maxRetries {
		attempts++

		// Create timeout context for step
		stepCtx, cancel := context.WithTimeout(ctx, timeout)

		resultCh := make(chan interface{}, 1)
		errCh := make(chan error, 1)

		verrors.SafeGoWithError("workflow:step-execute", errCh, func() error {
			result, err := step.Execute(stepCtx, data)
			if err != nil {
				return err
			}
			resultCh <- result
			return nil
		})

		select {
		case result := <-resultCh:
			cancel()
			return result, nil
		case err := <-errCh:
			cancel()
			lastErr = err
			if !step.CanRetry || attempts > maxRetries {
				return nil, lastErr
			}
			// Exponential backoff
			//nolint:gosec // G115: attempts is small retry counter
			backoff := time.Duration(1<<uint(attempts-1)) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		case <-stepCtx.Done():
			cancel()
			lastErr = stepCtx.Err()
			if !step.CanRetry || attempts > maxRetries {
				return nil, lastErr
			}
		}
	}

	return nil, lastErr
}

// handleStepError handles a step failure
func (e *Engine) handleStepError(ctx context.Context, state *WorkflowState, step WorkflowStep, checkpoint *Checkpoint, def *WorkflowDefinition, err error, duration time.Duration) {
	e.logger.Error().
		Err(err).
		Str("workflow_id", state.ID).
		Str("step", step.Name).
		Msg("step failed")

	// Update checkpoint
	checkpoint.Status = CheckpointFailed
	checkpoint.Error = err.Error()
	checkpoint.UpdatedAt = time.Now()
	if saveErr := e.store.SaveCheckpoint(ctx, state.ID, checkpoint); saveErr != nil {
		e.logger.Warn().Err(saveErr).Msg("failed to update failed checkpoint")
	}

	e.appendHistory(ctx, state.ID, &HistoryEvent{
		ID:          fmt.Sprintf("%s-%s-failed-%d", state.ID, step.Name, time.Now().UnixNano()),
		WorkflowID:  state.ID,
		EventType:   HistoryEventStepFailed,
		Step:        step.Name,
		Error:       err.Error(),
		Timestamp:   time.Now(),
		DurationMs:  duration.Milliseconds(),
		Description: fmt.Sprintf("Step '%s' failed: %v", step.Name, err),
	})

	// Try error handler
	if def.OnError != nil {
		if handleErr := def.OnError(ctx, step.Name, err); handleErr == nil {
			// Error was handled, don't fail the workflow
			e.logger.Info().Str("workflow_id", state.ID).Str("step", step.Name).Msg("error was handled by OnError callback")
			return
		}
	}

	// Mark workflow as failed
	state.Status = WorkflowStatusFailed
	state.Error = err.Error()
	now := time.Now()
	state.CompletedAt = &now
	state.UpdatedAt = now
	if saveErr := e.store.SaveState(ctx, state.ID, state); saveErr != nil {
		e.logger.Error().Err(saveErr).Msg("failed to save failed state")
	}

	e.appendHistory(ctx, state.ID, &HistoryEvent{
		ID:          fmt.Sprintf("%s-failed-%d", state.ID, time.Now().UnixNano()),
		WorkflowID:  state.ID,
		EventType:   HistoryEventWorkflowFailed,
		Status:      WorkflowStatusFailed,
		Error:       err.Error(),
		Timestamp:   time.Now(),
		Description: fmt.Sprintf("Workflow failed at step '%s': %v", step.Name, err),
	})
}

// handleWorkflowCancellation handles workflow cancellation
func (e *Engine) handleWorkflowCancellation(ctx context.Context, state *WorkflowState) {
	e.logger.Info().Str("workflow_id", state.ID).Msg("workflow cancelled")

	state.Status = WorkflowStatusCancelled
	now := time.Now()
	state.CompletedAt = &now
	state.UpdatedAt = now
	if err := e.store.SaveState(ctx, state.ID, state); err != nil {
		e.logger.Error().Err(err).Msg("failed to save cancelled state")
	}

	e.appendHistory(ctx, state.ID, &HistoryEvent{
		ID:          fmt.Sprintf("%s-cancelled-%d", state.ID, time.Now().UnixNano()),
		WorkflowID:  state.ID,
		EventType:   HistoryEventWorkflowCancelled,
		Status:      WorkflowStatusCancelled,
		Timestamp:   time.Now(),
		Description: "Workflow was cancelled",
	})
}

// appendHistory is a helper to append history events with error logging
func (e *Engine) appendHistory(ctx context.Context, workflowID string, event *HistoryEvent) {
	if err := e.store.AppendHistory(ctx, workflowID, event); err != nil {
		e.logger.Warn().Err(err).Str("workflow_id", workflowID).Msg("failed to append history event")
	}
}
