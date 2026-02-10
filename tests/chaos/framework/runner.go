// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"context"
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ChaosRunner manages chaos experiment execution
type ChaosRunner struct {
	kubeClient client.Client
	namespace  string
	validators []HealthChecker
}

// HealthChecker defines health check interface
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) error
}

// ChaosExperiment defines a chaos engineering experiment
type ChaosExperiment struct {
	Name        string
	ChaosObject client.Object
	Duration    time.Duration

	// Expected behavior
	MaxDowntime      time.Duration
	ExpectedRecovery time.Duration
	AllowDataLoss    bool

	// Validation
	PreConditions  []Condition
	PostConditions []Condition
}

// Condition is a validation function
type Condition func(ctx context.Context) error

// NewChaosRunner creates a new chaos runner
func NewChaosRunner(kubeClient client.Client, namespace string) *ChaosRunner {
	return &ChaosRunner{
		kubeClient: kubeClient,
		namespace:  namespace,
		validators: []HealthChecker{},
	}
}

// AddValidator adds a health checker
func (r *ChaosRunner) AddValidator(checker HealthChecker) {
	r.validators = append(r.validators, checker)
}

// RunExperiment executes a chaos experiment
func (r *ChaosRunner) RunExperiment(ctx context.Context, exp ChaosExperiment) (*ExperimentResult, error) {
	result := &ExperimentResult{
		Name:      exp.Name,
		StartTime: time.Now(),
	}

	for _, cond := range exp.PreConditions {
		if err := cond(ctx); err != nil {
			return nil, fmt.Errorf("pre-condition failed: %w", err)
		}
	}

	if err := r.kubeClient.Create(ctx, exp.ChaosObject); err != nil {
		return nil, fmt.Errorf("create chaos: %w", err)
	}
	defer r.cleanup(ctx, exp.ChaosObject)

	healthTicker := time.NewTicker(5 * time.Second)
	defer healthTicker.Stop()

	var downtimeStart time.Time
	var totalDowntime time.Duration

	timeout := time.After(exp.Duration + exp.MaxDowntime + exp.ExpectedRecovery)

	for {
		select {
		case <-timeout:
			result.EndTime = time.Now()
			result.TotalDowntime = totalDowntime

			for _, cond := range exp.PostConditions {
				if err := cond(ctx); err != nil {
					result.PostConditionsFailed = append(result.PostConditionsFailed, err.Error())
				}
			}

			result.Success = len(result.PostConditionsFailed) == 0 &&
				totalDowntime <= exp.MaxDowntime

			return result, nil

		case <-healthTicker.C:
			healthy := true
			for _, checker := range r.validators {
				if err := checker.Check(ctx); err != nil {
					healthy = false
					result.HealthErrors = append(result.HealthErrors, HealthError{
						Time:    time.Now(),
						Checker: checker.Name(),
						Error:   err.Error(),
					})
					break
				}
			}

			if !healthy && downtimeStart.IsZero() {
				downtimeStart = time.Now()
			} else if healthy && !downtimeStart.IsZero() {
				totalDowntime += time.Since(downtimeStart)
				result.RecoveryTime = time.Since(downtimeStart)
				downtimeStart = time.Time{}
			}
		}
	}
}

func (r *ChaosRunner) cleanup(ctx context.Context, obj client.Object) {
	if err := r.kubeClient.Delete(ctx, obj); err != nil {
		fmt.Printf("Failed to cleanup chaos object: %v\n", err)
	}
}

// ExperimentResult contains the result of a chaos experiment
type ExperimentResult struct {
	Name                 string
	StartTime            time.Time
	EndTime              time.Time
	Success              bool
	TotalDowntime        time.Duration
	RecoveryTime         time.Duration
	HealthErrors         []HealthError
	PostConditionsFailed []string
}

// HealthError records a health check failure
type HealthError struct {
	Time    time.Time
	Checker string
	Error   string
}
