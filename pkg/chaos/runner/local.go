// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

// Package runner provides experiment execution backends for chaos engineering.
package runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/chaos"
)

// LocalRunner implements the ExperimentRunner interface for local testing.
// It simulates chaos injection without actually affecting infrastructure,
// making it suitable for unit tests and dry-run experiments.
type LocalRunner struct {
	mu sync.RWMutex

	// experiments tracks running experiments
	experiments map[string]*chaos.ExperimentSpec

	// simulatedDelay adds delay to simulate real execution
	simulatedDelay time.Duration

	// failureRate simulates random failures (0.0-1.0)
	failureRate float64

	// logger for debug output
	logger chaos.Logger
}

// LocalRunnerOption configures the LocalRunner.
type LocalRunnerOption func(*LocalRunner)

// WithSimulatedDelay sets the simulated execution delay.
func WithSimulatedDelay(delay time.Duration) LocalRunnerOption {
	return func(r *LocalRunner) {
		r.simulatedDelay = delay
	}
}

// WithFailureRate sets the simulated failure rate.
func WithFailureRate(rate float64) LocalRunnerOption {
	return func(r *LocalRunner) {
		r.failureRate = rate
	}
}

// WithLocalLogger sets the logger.
func WithLocalLogger(logger chaos.Logger) LocalRunnerOption {
	return func(r *LocalRunner) {
		r.logger = logger
	}
}

// NewLocalRunner creates a new local runner for testing.
func NewLocalRunner(opts ...LocalRunnerOption) *LocalRunner {
	r := &LocalRunner{
		experiments:    make(map[string]*chaos.ExperimentSpec),
		simulatedDelay: 100 * time.Millisecond,
		failureRate:    0,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Start begins a chaos experiment simulation.
func (r *LocalRunner) Start(ctx context.Context, experiment *chaos.ExperimentSpec) error {
	if experiment == nil {
		return fmt.Errorf("experiment is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.experiments[experiment.ID]; exists {
		return fmt.Errorf("experiment %s already running", experiment.ID)
	}

	// Simulate startup delay
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(r.simulatedDelay):
	}

	experiment.State = chaos.ExperimentStateRunning
	experiment.StartTime = time.Now()
	r.experiments[experiment.ID] = experiment

	if r.logger != nil {
		r.logger.Info("started experiment",
			chaos.LogField{Key: "experiment_id", Value: experiment.ID},
			chaos.LogField{Key: "type", Value: string(experiment.Type)},
		)
	}

	return nil
}

// Stop terminates a running experiment.
func (r *LocalRunner) Stop(ctx context.Context, experiment *chaos.ExperimentSpec) error {
	if experiment == nil {
		return fmt.Errorf("experiment is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	running, exists := r.experiments[experiment.ID]
	if !exists {
		return fmt.Errorf("experiment %s not found", experiment.ID)
	}

	// Simulate stop delay
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(r.simulatedDelay):
	}

	running.State = chaos.ExperimentStateCompleted
	running.EndTime = time.Now()
	delete(r.experiments, experiment.ID)

	if r.logger != nil {
		r.logger.Info("stopped experiment",
			chaos.LogField{Key: "experiment_id", Value: experiment.ID},
		)
	}

	return nil
}

// Status returns the current state of an experiment.
func (r *LocalRunner) Status(_ context.Context, experimentID string) (*chaos.ExperimentSpec, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	experiment, exists := r.experiments[experimentID]
	if !exists {
		return nil, fmt.Errorf("experiment %s not found", experimentID)
	}

	return experiment, nil
}

// Rollback reverts changes made by an experiment.
func (r *LocalRunner) Rollback(ctx context.Context, experiment *chaos.ExperimentSpec) error {
	if experiment == nil {
		return fmt.Errorf("experiment is nil")
	}

	// Simulate rollback delay
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(r.simulatedDelay):
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if running, exists := r.experiments[experiment.ID]; exists {
		running.State = chaos.ExperimentStateAborted
		delete(r.experiments, experiment.ID)
	}

	if r.logger != nil {
		r.logger.Info("rolled back experiment",
			chaos.LogField{Key: "experiment_id", Value: experiment.ID},
		)
	}

	return nil
}

// ExecuteAction executes a single experiment action.
func (r *LocalRunner) ExecuteAction(ctx context.Context, action chaos.ExperimentAction) error {
	// Simulate action execution
	duration := action.Duration
	if duration == 0 {
		duration = r.simulatedDelay
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(duration):
	}

	if r.logger != nil {
		r.logger.Info("executed action",
			chaos.LogField{Key: "action", Value: action.Name},
			chaos.LogField{Key: "type", Value: action.Type},
		)
	}

	return nil
}

// ExecuteRollback executes a rollback action.
func (r *LocalRunner) ExecuteRollback(ctx context.Context, action chaos.RollbackAction) error {
	// Simulate rollback execution using action timeout if specified
	delay := action.Timeout
	if delay == 0 {
		delay = r.simulatedDelay
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
	}

	if r.logger != nil {
		r.logger.Info("executed rollback",
			chaos.LogField{Key: "rollback_type", Value: action.Type},
		)
	}

	return nil
}

// Validate validates an action before execution.
func (r *LocalRunner) Validate(action chaos.ExperimentAction) error {
	if action.Name == "" {
		return fmt.Errorf("action name is required")
	}
	return nil
}

// ListRunning returns all running experiments.
func (r *LocalRunner) ListRunning() []*chaos.ExperimentSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*chaos.ExperimentSpec, 0, len(r.experiments))
	for _, exp := range r.experiments {
		result = append(result, exp)
	}
	return result
}

// Reset clears all running experiments (for testing).
func (r *LocalRunner) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.experiments = make(map[string]*chaos.ExperimentSpec)
}

