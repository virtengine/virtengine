// Copyright 2024 VirtEngine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chaos

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Controller Error Definitions
// ============================================================================

var (
	// ErrExperimentNotFound is returned when an experiment ID is not found
	ErrExperimentNotFound = errors.New("experiment not found")

	// ErrExperimentAlreadyRunning is returned when attempting to run an already running experiment
	ErrExperimentAlreadyRunning = errors.New("experiment already running")

	// ErrExperimentNotRunning is returned when attempting to pause/cancel a non-running experiment
	ErrExperimentNotRunning = errors.New("experiment not running")

	// ErrExperimentNotPaused is returned when attempting to resume a non-paused experiment
	ErrExperimentNotPaused = errors.New("experiment not paused")

	// ErrInvalidSchedule is returned when the schedule format is invalid
	ErrInvalidSchedule = errors.New("invalid schedule format")

	// ErrControllerStopped is returned when operations are attempted on a stopped controller
	ErrControllerStopped = errors.New("controller has been stopped")

	// ErrSteadyStateViolation is returned when steady state hypothesis verification fails
	ErrSteadyStateViolation = errors.New("steady state hypothesis violated")

	// ErrRollbackFailed is returned when rollback execution fails
	ErrRollbackFailed = errors.New("rollback execution failed")
)

// ============================================================================
// Controller-specific Type Definitions
// ============================================================================

// ExperimentAction defines an action to execute during the experiment.
type ExperimentAction struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Provider string                 `json:"provider"`
	Duration time.Duration          `json:"duration,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

// ControllerExperiment extends ExperimentSpec with controller-specific runtime state.
// This wraps the base ExperimentSpec with additional fields needed for execution control.
type ControllerExperiment struct {
	*ExperimentSpec

	// Actions are the chaos actions to execute
	Actions []ExperimentAction `json:"actions,omitempty"`

	// RollbackActions are actions to execute on rollback
	RollbackActions []RollbackAction `json:"rollback_actions,omitempty"`

	// Tags for categorization
	Tags []string `json:"tags,omitempty"`

	// Timeline records events during execution
	Timeline []Event `json:"timeline"`

	// Internal state for controller orchestration
	cancelFunc context.CancelFunc `json:"-"`
	pauseCh    chan struct{}      `json:"-"`
	resumeCh   chan struct{}      `json:"-"`
}

// ControllerExperimentResults extends ExperimentResults with controller-specific fields.
type ControllerExperimentResults struct {
	ExperimentID          string         `json:"experiment_id"`
	Success               bool           `json:"success"`
	Error                 string         `json:"error,omitempty"`
	StartTime             time.Time      `json:"start_time"`
	EndTime               time.Time      `json:"end_time"`
	Duration              time.Duration  `json:"duration"`
	Timeline              []Event        `json:"timeline"`
	SteadyStateVerified   bool           `json:"steady_state_verified"`
	SteadyStateViolations []Violation    `json:"steady_state_violations,omitempty"`
	ActionsExecuted       int            `json:"actions_executed"`
	RollbackExecuted      bool           `json:"rollback_executed"`
	MetricsData           map[string]float64 `json:"metrics,omitempty"`
}

// ============================================================================
// Controller Interface Definitions
// ============================================================================

// ControllerExperimentRunner executes chaos experiment actions for the controller.
type ControllerExperimentRunner interface {
	// ExecuteAction executes a single experiment action
	ExecuteAction(ctx context.Context, action ExperimentAction) error

	// ExecuteRollback executes a rollback action
	ExecuteRollback(ctx context.Context, action RollbackAction) error

	// Validate validates an action before execution
	Validate(action ExperimentAction) error
}

// ControllerSLOVerifier verifies service level objectives for the controller.
type ControllerSLOVerifier interface {
	// VerifySteadyState verifies the steady state hypothesis
	VerifySteadyState(ctx context.Context, hypothesis *SteadyStateHypothesis) ([]SteadyStateViolation, error)

	// CheckProbe checks a single steady state probe
	CheckProbe(ctx context.Context, probe Probe) (float64, error)
}

// Logger provides structured logging for the chaos controller.
type Logger interface {
	Debug(msg string, fields ...LogField)
	Info(msg string, fields ...LogField)
	Warn(msg string, fields ...LogField)
	Error(msg string, fields ...LogField)
}

// LogField represents a structured log field.
type LogField struct {
	Key   string
	Value interface{}
}

// ============================================================================
// Controller Metrics
// ============================================================================

// ControllerMetrics tracks simple counters for controller operations.
// This is separate from the Prometheus-based Metrics type in metrics.go.
type ControllerMetrics struct {
	ExperimentsStarted   int64
	ExperimentsCompleted int64
	ExperimentsFailed    int64
	ExperimentsCanceled  int64
	ActionsExecuted      int64
	ActionsFailed        int64
	RollbacksExecuted    int64
	RollbacksFailed      int64
	SLOViolations        int64
	mu                   sync.RWMutex
}

// NewControllerMetrics creates a new ControllerMetrics instance.
func NewControllerMetrics() *ControllerMetrics {
	return &ControllerMetrics{}
}

// IncrementExperimentsStarted increments the experiments started counter.
func (m *ControllerMetrics) IncrementExperimentsStarted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ExperimentsStarted++
}

// IncrementExperimentsCompleted increments the experiments completed counter.
func (m *ControllerMetrics) IncrementExperimentsCompleted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ExperimentsCompleted++
}

// IncrementExperimentsFailed increments the experiments failed counter.
func (m *ControllerMetrics) IncrementExperimentsFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ExperimentsFailed++
}

// IncrementExperimentsCanceled increments the experiments canceled counter.
func (m *ControllerMetrics) IncrementExperimentsCanceled() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ExperimentsCanceled++
}

// IncrementActionsExecuted increments the actions executed counter.
func (m *ControllerMetrics) IncrementActionsExecuted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActionsExecuted++
}

// IncrementActionsFailed increments the actions failed counter.
func (m *ControllerMetrics) IncrementActionsFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActionsFailed++
}

// IncrementRollbacksExecuted increments the rollbacks executed counter.
func (m *ControllerMetrics) IncrementRollbacksExecuted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RollbacksExecuted++
}

// IncrementRollbacksFailed increments the rollbacks failed counter.
func (m *ControllerMetrics) IncrementRollbacksFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RollbacksFailed++
}

// IncrementSLOViolations increments the SLO violations counter.
func (m *ControllerMetrics) IncrementSLOViolations() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SLOViolations++
}

// Snapshot returns a copy of the current metrics.
func (m *ControllerMetrics) Snapshot() ControllerMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return ControllerMetrics{
		ExperimentsStarted:   m.ExperimentsStarted,
		ExperimentsCompleted: m.ExperimentsCompleted,
		ExperimentsFailed:    m.ExperimentsFailed,
		ExperimentsCanceled:  m.ExperimentsCanceled,
		ActionsExecuted:      m.ActionsExecuted,
		ActionsFailed:        m.ActionsFailed,
		RollbacksExecuted:    m.RollbacksExecuted,
		RollbacksFailed:      m.RollbacksFailed,
		SLOViolations:        m.SLOViolations,
	}
}

// ============================================================================
// Scheduler
// ============================================================================

// ScheduledExperiment represents a scheduled experiment.
type ScheduledExperiment struct {
	Experiment   *ExperimentSpec `json:"experiment"`
	Schedule     string          `json:"schedule"`
	NextRun      time.Time       `json:"next_run"`
	LastRun      *time.Time      `json:"last_run,omitempty"`
	RunCount     int             `json:"run_count"`
	Enabled      bool            `json:"enabled"`
}

// ExperimentScheduler schedules experiments for recurring execution.
type ExperimentScheduler struct {
	cronSchedule    string
	enabled         bool
	experimentQueue []*ExperimentSpec
	scheduled       map[string]*ScheduledExperiment
	mu              sync.RWMutex
	stopCh          chan struct{}
	controller      *Controller
}

// NewExperimentScheduler creates a new experiment scheduler.
func NewExperimentScheduler(controller *Controller) *ExperimentScheduler {
	return &ExperimentScheduler{
		enabled:         true,
		experimentQueue: make([]*ExperimentSpec, 0),
		scheduled:       make(map[string]*ScheduledExperiment),
		stopCh:          make(chan struct{}),
		controller:      controller,
	}
}

// Schedule adds an experiment to the schedule.
func (s *ExperimentScheduler) Schedule(exp *ExperimentSpec, schedule string) error {
	if schedule == "" {
		return ErrInvalidSchedule
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Parse next run time from schedule
	nextRun, err := parseSchedule(schedule)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSchedule, err)
	}

	s.scheduled[exp.ID] = &ScheduledExperiment{
		Experiment: exp,
		Schedule:   schedule,
		NextRun:    nextRun,
		Enabled:    true,
	}

	return nil
}

// Unschedule removes an experiment from the schedule.
func (s *ExperimentScheduler) Unschedule(experimentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.scheduled[experimentID]; !ok {
		return ErrExperimentNotFound
	}

	delete(s.scheduled, experimentID)
	return nil
}

// GetScheduled returns all scheduled experiments.
func (s *ExperimentScheduler) GetScheduled() []*ScheduledExperiment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*ScheduledExperiment, 0, len(s.scheduled))
	for _, exp := range s.scheduled {
		result = append(result, exp)
	}
	return result
}

// Enable enables the scheduler.
func (s *ExperimentScheduler) Enable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = true
}

// Disable disables the scheduler.
func (s *ExperimentScheduler) Disable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = false
}

// IsEnabled returns whether the scheduler is enabled.
func (s *ExperimentScheduler) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// Start starts the scheduler loop.
func (s *ExperimentScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.processSchedule(ctx)
		}
	}
}

// Stop stops the scheduler.
func (s *ExperimentScheduler) Stop() {
	close(s.stopCh)
}

// processSchedule checks and executes due experiments.
func (s *ExperimentScheduler) processSchedule(ctx context.Context) {
	s.mu.Lock()
	if !s.enabled {
		s.mu.Unlock()
		return
	}

	now := time.Now()
	toRun := make([]*ScheduledExperiment, 0)

	for _, scheduled := range s.scheduled {
		if scheduled.Enabled && now.After(scheduled.NextRun) {
			toRun = append(toRun, scheduled)
		}
	}
	s.mu.Unlock()

	for _, scheduled := range toRun {
		if s.controller != nil {
			// Create a copy of the experiment for this run
			expCopy := *scheduled.Experiment
			expCopy.ID = fmt.Sprintf("%s-%d", scheduled.Experiment.ID, time.Now().UnixNano())
			expCopy.State = ExperimentStatePending
			expCopy.Results = nil

			go func(exp *ExperimentSpec, sched *ScheduledExperiment) {
				_, _ = s.controller.Execute(ctx, exp)

				s.mu.Lock()
				defer s.mu.Unlock()
				now := time.Now()
				sched.LastRun = &now
				sched.RunCount++
				nextRun, err := parseSchedule(sched.Schedule)
				if err == nil {
					sched.NextRun = nextRun
				}
			}(&expCopy, scheduled)
		}
	}
}

// parseSchedule parses a schedule string and returns the next run time
func parseSchedule(schedule string) (time.Time, error) {
	// Simple implementation supporting duration-based scheduling
	// Format: "every 1h", "every 30m", "every 24h"
	var duration time.Duration
	_, err := fmt.Sscanf(schedule, "every %v", &duration)
	if err != nil {
		return time.Time{}, fmt.Errorf("unsupported schedule format: %s", schedule)
	}
	return time.Now().Add(duration), nil
}

// ============================================================================
// No-op Logger
// ============================================================================

// noopLogger is a logger that does nothing
type noopLogger struct{}

func (noopLogger) Debug(_ string, _ ...LogField) {}
func (noopLogger) Info(_ string, _ ...LogField)  {}
func (noopLogger) Warn(_ string, _ ...LogField)  {}
func (noopLogger) Error(_ string, _ ...LogField) {}

// ============================================================================
// Controller
// ============================================================================

// Controller orchestrates chaos experiments.
type Controller struct {
	runner      ControllerExperimentRunner
	sloVerifier ControllerSLOVerifier
	metrics     *ControllerMetrics
	experiments map[string]*ExperimentSpec
	mu          sync.RWMutex
	scheduler   *ExperimentScheduler
	logger      Logger
	stopped     bool
	stopCh      chan struct{}
	wg          sync.WaitGroup

	// Internal experiment state tracking
	cancelFuncs map[string]context.CancelFunc
	pauseChans  map[string]chan struct{}
	resumeChans map[string]chan struct{}
}

// ControllerOption configures the Controller.
type ControllerOption func(*Controller)

// WithRunner sets the experiment runner.
func WithRunner(runner ControllerExperimentRunner) ControllerOption {
	return func(c *Controller) {
		c.runner = runner
	}
}

// WithSLOVerifier sets the SLO verifier.
func WithSLOVerifier(verifier ControllerSLOVerifier) ControllerOption {
	return func(c *Controller) {
		c.sloVerifier = verifier
	}
}

// WithMetrics sets the metrics collector.
func WithMetrics(metrics *ControllerMetrics) ControllerOption {
	return func(c *Controller) {
		c.metrics = metrics
	}
}

// WithLogger sets the logger.
func WithLogger(logger Logger) ControllerOption {
	return func(c *Controller) {
		c.logger = logger
	}
}

// NewController creates a new chaos experiment controller.
func NewController(opts ...ControllerOption) *Controller {
	c := &Controller{
		experiments: make(map[string]*ExperimentSpec),
		metrics:     NewControllerMetrics(),
		logger:      noopLogger{},
		stopCh:      make(chan struct{}),
		cancelFuncs: make(map[string]context.CancelFunc),
		pauseChans:  make(map[string]chan struct{}),
		resumeChans: make(map[string]chan struct{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	c.scheduler = NewExperimentScheduler(c)

	return c
}

// Start starts the controller and its scheduler.
func (c *Controller) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return ErrControllerStopped
	}
	c.mu.Unlock()

	c.logger.Info("chaos controller starting")

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.scheduler.Start(ctx)
	}()

	return nil
}

// Stop gracefully stops the controller.
func (c *Controller) Stop() error {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return nil
	}
	c.stopped = true
	c.mu.Unlock()

	c.logger.Info("chaos controller stopping")

	close(c.stopCh)
	c.scheduler.Stop()

	// Cancel all running experiments
	c.mu.RLock()
	for id, exp := range c.experiments {
		if exp.State == ExperimentStateRunning || exp.State == ExperimentStatePaused {
			if cancelFunc, ok := c.cancelFuncs[id]; ok {
				cancelFunc()
			}
		}
	}
	c.mu.RUnlock()

	c.wg.Wait()

	c.logger.Info("chaos controller stopped")
	return nil
}

// Execute runs a chaos experiment and returns the results.
func (c *Controller) Execute(ctx context.Context, exp *ExperimentSpec) (*ExperimentResults, error) {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return nil, ErrControllerStopped
	}

	if existing, ok := c.experiments[exp.ID]; ok {
		if !existing.State.IsTerminal() {
			c.mu.Unlock()
			return nil, ErrExperimentAlreadyRunning
		}
	}

	// Initialize experiment
	exp.State = ExperimentStatePending
	if exp.Results == nil {
		exp.Results = &ExperimentResults{
			Timeline: make([]Event, 0),
			Metrics:  make(map[string]float64),
		}
	}

	expCtx, cancel := context.WithCancel(ctx)
	c.cancelFuncs[exp.ID] = cancel
	c.pauseChans[exp.ID] = make(chan struct{})
	c.resumeChans[exp.ID] = make(chan struct{})

	c.experiments[exp.ID] = exp
	c.mu.Unlock()

	c.logger.Info("executing experiment",
		LogField{Key: "experiment_id", Value: exp.ID},
		LogField{Key: "experiment_name", Value: exp.Name},
	)

	results := c.runExperiment(expCtx, exp)

	return results, nil
}

// Schedule schedules an experiment for recurring execution.
func (c *Controller) Schedule(exp *ExperimentSpec, schedule string) error {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return ErrControllerStopped
	}
	c.mu.Unlock()

	c.logger.Info("scheduling experiment",
		LogField{Key: "experiment_id", Value: exp.ID},
		LogField{Key: "schedule", Value: schedule},
	)

	return c.scheduler.Schedule(exp, schedule)
}

// List returns all experiments.
func (c *Controller) List() []*ExperimentSpec {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*ExperimentSpec, 0, len(c.experiments))
	for _, exp := range c.experiments {
		result = append(result, exp)
	}
	return result
}

// Get retrieves an experiment by ID.
func (c *Controller) Get(id string) (*ExperimentSpec, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	exp, ok := c.experiments[id]
	return exp, ok
}

// Cancel cancels a running experiment.
func (c *Controller) Cancel(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	exp, ok := c.experiments[id]
	if !ok {
		return ErrExperimentNotFound
	}

	if exp.State.IsTerminal() {
		return ErrExperimentNotRunning
	}

	c.logger.Info("canceling experiment",
		LogField{Key: "experiment_id", Value: id},
	)

	if cancelFunc, ok := c.cancelFuncs[id]; ok {
		cancelFunc()
	}

	c.updateState(exp, ExperimentStateAborted)
	c.recordTimelineEvent(exp, Event{
		Timestamp: time.Now(),
		Type:      EventTypeExperimentAborted,
		Message:   "experiment canceled by user",
	})

	if c.metrics != nil {
		c.metrics.IncrementExperimentsCanceled()
	}

	return nil
}

// PauseExperiment pauses a running experiment.
func (c *Controller) PauseExperiment(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	exp, ok := c.experiments[id]
	if !ok {
		return ErrExperimentNotFound
	}

	if exp.State != ExperimentStateRunning {
		return ErrExperimentNotRunning
	}

	c.logger.Info("pausing experiment",
		LogField{Key: "experiment_id", Value: id},
	)

	c.updateState(exp, ExperimentStatePaused)
	c.recordTimelineEvent(exp, Event{
		Timestamp: time.Now(),
		Type:      EventTypeChaosRemoved,
		Message:   "experiment paused",
	})

	// Signal pause
	if ch, ok := c.pauseChans[id]; ok {
		close(ch)
		c.pauseChans[id] = make(chan struct{})
	}

	return nil
}

// ResumeExperiment resumes a paused experiment.
func (c *Controller) ResumeExperiment(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	exp, ok := c.experiments[id]
	if !ok {
		return ErrExperimentNotFound
	}

	if exp.State != ExperimentStatePaused {
		return ErrExperimentNotPaused
	}

	c.logger.Info("resuming experiment",
		LogField{Key: "experiment_id", Value: id},
	)

	c.updateState(exp, ExperimentStateRunning)
	c.recordTimelineEvent(exp, Event{
		Timestamp: time.Now(),
		Type:      EventTypeChaosInjected,
		Message:   "experiment resumed",
	})

	// Signal resume
	if ch, ok := c.resumeChans[id]; ok {
		close(ch)
		c.resumeChans[id] = make(chan struct{})
	}

	return nil
}

// GetMetrics returns the current metrics snapshot.
func (c *Controller) GetMetrics() ControllerMetrics {
	if c.metrics == nil {
		return ControllerMetrics{}
	}
	return c.metrics.Snapshot()
}

// GetScheduler returns the experiment scheduler.
func (c *Controller) GetScheduler() *ExperimentScheduler {
	return c.scheduler
}

// ============================================================================
// Internal Methods
// ============================================================================

// runExperiment executes the experiment workflow.
func (c *Controller) runExperiment(ctx context.Context, exp *ExperimentSpec) *ExperimentResults {
	startTime := time.Now()
	exp.StartTime = startTime

	c.updateState(exp, ExperimentStateRunning)
	c.recordTimelineEvent(exp, Event{
		Timestamp: startTime,
		Type:      EventTypeExperimentStarted,
		Message:   fmt.Sprintf("experiment '%s' started", exp.Name),
	})

	if c.metrics != nil {
		c.metrics.IncrementExperimentsStarted()
	}

	results := &ExperimentResults{
		Success:  false,
		Duration: 0,
		Metrics:  make(map[string]float64),
		Timeline: make([]Event, 0),
	}

	// Verify initial steady state
	if exp.SteadyStateHypothesis != nil {
		violations, err := c.verifySteadyState(ctx, exp)
		if err != nil {
			c.handleExperimentError(exp, results, fmt.Errorf("initial steady state verification failed: %w", err))
			return results
		}
		if len(violations) > 0 {
			results.SteadyStateViolations = violations
			c.handleExperimentError(exp, results, ErrSteadyStateViolation)
			return results
		}
	}

	// Note: ExperimentSpec does not have Actions field directly.
	// The runner handles execution through Start/Stop interface from types.go.
	// For now, we simulate action execution by calling runner.Start().
	if c.runner != nil {
		// Create action from experiment parameters
		action := ExperimentAction{
			Name:     exp.Name,
			Type:     string(exp.Type),
			Duration: exp.Duration,
			Config:   exp.Parameters,
		}

		// Check for cancellation
		select {
		case <-ctx.Done():
			c.recordTimelineEvent(exp, Event{
				Timestamp: time.Now(),
				Type:      EventTypeExperimentAborted,
				Message:   "experiment canceled",
			})
			results.Success = false
			results.Error = "experiment canceled"
			return c.finalizeResults(exp, results)
		default:
		}

		// Check for pause
		c.mu.RLock()
		isPaused := exp.State == ExperimentStatePaused
		resumeCh := c.resumeChans[exp.ID]
		c.mu.RUnlock()

		if isPaused {
			c.logger.Debug("experiment paused, waiting for resume",
				LogField{Key: "experiment_id", Value: exp.ID},
			)
			select {
			case <-ctx.Done():
				return c.finalizeResults(exp, results)
			case <-resumeCh:
				// Continue
			}
		}

		c.recordTimelineEvent(exp, Event{
			Timestamp: time.Now(),
			Type:      EventTypeChaosInjected,
			Message:   fmt.Sprintf("executing action: %s", action.Name),
			Details: map[string]interface{}{
				"action_type": action.Type,
			},
		})

		if err := c.runner.ExecuteAction(ctx, action); err != nil {
			if c.metrics != nil {
				c.metrics.IncrementActionsFailed()
			}
			c.logger.Error("action execution failed",
				LogField{Key: "experiment_id", Value: exp.ID},
				LogField{Key: "action", Value: action.Name},
				LogField{Key: "error", Value: err.Error()},
			)
			c.executeRollback(ctx, exp)
			c.handleExperimentError(exp, results, fmt.Errorf("action '%s' failed: %w", action.Name, err))
			return results
		}

		if c.metrics != nil {
			c.metrics.IncrementActionsExecuted()
		}

		// Verify steady state after action
		if exp.SteadyStateHypothesis != nil && c.sloVerifier != nil {
			violations, err := c.verifySteadyState(ctx, exp)
			if err != nil {
				c.logger.Warn("steady state verification failed during experiment",
					LogField{Key: "experiment_id", Value: exp.ID},
					LogField{Key: "error", Value: err.Error()},
				)
			} else if len(violations) > 0 {
				results.SteadyStateViolations = append(results.SteadyStateViolations, violations...)
				c.recordTimelineEvent(exp, Event{
					Timestamp: time.Now(),
					Type:      EventTypeSteadyStateViolation,
					Message:   fmt.Sprintf("SLO violation detected: %d violations", len(violations)),
				})
			}
		}
	}

	// Final steady state verification
	if exp.SteadyStateHypothesis != nil && c.sloVerifier != nil {
		violations, err := c.verifySteadyState(ctx, exp)
		if err != nil {
			c.logger.Warn("final steady state verification failed",
				LogField{Key: "experiment_id", Value: exp.ID},
				LogField{Key: "error", Value: err.Error()},
			)
		} else if len(violations) > 0 {
			results.SteadyStateViolations = append(results.SteadyStateViolations, violations...)
		}
	}

	results.Success = true
	c.updateState(exp, ExperimentStateCompleted)
	c.recordTimelineEvent(exp, Event{
		Timestamp: time.Now(),
		Type:      EventTypeExperimentCompleted,
		Message:   "experiment completed successfully",
	})

	if c.metrics != nil {
		c.metrics.IncrementExperimentsCompleted()
	}

	c.logger.Info("experiment completed successfully",
		LogField{Key: "experiment_id", Value: exp.ID},
	)

	return c.finalizeResults(exp, results)
}

// verifySteadyState verifies the steady state hypothesis.
func (c *Controller) verifySteadyState(ctx context.Context, exp *ExperimentSpec) ([]SteadyStateViolation, error) {
	if c.sloVerifier == nil {
		return nil, nil
	}

	c.recordTimelineEvent(exp, Event{
		Timestamp: time.Now(),
		Type:      EventTypeSteadyStateCheck,
		Message:   "verifying steady state hypothesis",
	})

	violations, err := c.sloVerifier.VerifySteadyState(ctx, exp.SteadyStateHypothesis)
	if err != nil {
		return nil, err
	}

	if len(violations) > 0 && c.metrics != nil {
		for range violations {
			c.metrics.IncrementSLOViolations()
		}
	}

	return violations, nil
}

// executeRollback executes rollback actions.
func (c *Controller) executeRollback(ctx context.Context, exp *ExperimentSpec) {
	if exp.Rollback == nil {
		return
	}

	c.logger.Info("executing rollback",
		LogField{Key: "experiment_id", Value: exp.ID},
	)

	// Record state change
	prevState := exp.State
	exp.State = ExperimentStatePending // Temporarily mark as pending during rollback

	c.recordTimelineEvent(exp, Event{
		Timestamp: time.Now(),
		Type:      EventTypeRollbackStarted,
		Message:   "starting rollback",
	})

	if c.runner != nil {
		if err := c.runner.ExecuteRollback(ctx, *exp.Rollback); err != nil {
			if c.metrics != nil {
				c.metrics.IncrementRollbacksFailed()
			}
			c.logger.Error("rollback action failed",
				LogField{Key: "experiment_id", Value: exp.ID},
				LogField{Key: "error", Value: err.Error()},
			)
			c.recordTimelineEvent(exp, Event{
				Timestamp: time.Now(),
				Type:      EventTypeRollbackFailed,
				Message:   fmt.Sprintf("rollback failed: %v", err),
			})
			exp.State = prevState
			return
		}

		if c.metrics != nil {
			c.metrics.IncrementRollbacksExecuted()
		}
	}

	c.recordTimelineEvent(exp, Event{
		Timestamp: time.Now(),
		Type:      EventTypeRollbackCompleted,
		Message:   "rollback completed",
	})

	exp.State = prevState
}

// recordTimelineEvent records a timeline event.
func (c *Controller) recordTimelineEvent(exp *ExperimentSpec, event Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if exp.Results == nil {
		exp.Results = &ExperimentResults{
			Timeline: make([]Event, 0),
			Metrics:  make(map[string]float64),
		}
	}
	exp.Results.Timeline = append(exp.Results.Timeline, event)
}

// updateState updates the experiment state.
func (c *Controller) updateState(exp *ExperimentSpec, state ExperimentState) {
	exp.State = state
	exp.UpdatedAt = time.Now()
	if state.IsTerminal() {
		exp.EndTime = time.Now()
	}
}

// handleExperimentError handles experiment errors.
func (c *Controller) handleExperimentError(exp *ExperimentSpec, results *ExperimentResults, err error) {
	c.updateState(exp, ExperimentStateFailed)
	results.Success = false
	results.Error = err.Error()

	c.recordTimelineEvent(exp, Event{
		Timestamp: time.Now(),
		Type:      EventTypeExperimentFailed,
		Message:   fmt.Sprintf("experiment failed: %v", err),
	})

	if c.metrics != nil {
		c.metrics.IncrementExperimentsFailed()
	}

	c.logger.Error("experiment failed",
		LogField{Key: "experiment_id", Value: exp.ID},
		LogField{Key: "error", Value: err.Error()},
	)

	c.finalizeResults(exp, results)
}

// finalizeResults finalizes the experiment results.
func (c *Controller) finalizeResults(exp *ExperimentSpec, results *ExperimentResults) *ExperimentResults {
	results.Duration = time.Since(exp.StartTime)

	c.mu.RLock()
	if exp.Results != nil && len(exp.Results.Timeline) > 0 {
		results.Timeline = make([]Event, len(exp.Results.Timeline))
		copy(results.Timeline, exp.Results.Timeline)
	}
	c.mu.RUnlock()

	// Store results in experiment
	exp.Results = results

	return results
}
