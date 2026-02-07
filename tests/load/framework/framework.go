// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ProfileType defines the type of load profile
type ProfileType string

const (
	// ProfileConstant maintains a constant load rate
	ProfileConstant ProfileType = "constant"
	// ProfileStep increases load in discrete steps
	ProfileStep ProfileType = "step"
	// ProfileLinear increases load linearly over time
	ProfileLinear ProfileType = "linear"
	// ProfileSpike creates sudden load spikes
	ProfileSpike ProfileType = "spike"
)

// LoadProfile defines the load pattern for a test
type LoadProfile struct {
	Type      ProfileType
	Duration  time.Duration
	StartRate float64 // requests per second
	EndRate   float64
	StepSize  float64
	StepTime  time.Duration
}

// Scenario defines the interface for load test scenarios
type Scenario interface {
	Name() string
	Setup(ctx context.Context) error
	Execute(ctx context.Context) (*ExecutionResult, error)
	Teardown(ctx context.Context) error
}

// ExecutionResult contains the result of a single scenario execution
type ExecutionResult struct {
	Success    bool
	Duration   time.Duration
	StatusCode int
	Error      error
	Metadata   map[string]interface{}
}

// LoadTest manages the execution of a load test scenario
type LoadTest struct {
	name     string
	scenario Scenario
	profile  LoadProfile
	metrics  *MetricsCollector
	workers  int
}

// NewLoadTest creates a new load test instance
func NewLoadTest(name string, scenario Scenario, profile LoadProfile) *LoadTest {
	return &LoadTest{
		name:     name,
		scenario: scenario,
		profile:  profile,
		metrics:  NewMetricsCollector(),
		workers:  100,
	}
}

// WithWorkers sets the number of concurrent workers
func (lt *LoadTest) WithWorkers(n int) *LoadTest {
	lt.workers = n
	return lt
}

// Run executes the load test
func (lt *LoadTest) Run(ctx context.Context) (*TestReport, error) {
	if err := lt.scenario.Setup(ctx); err != nil {
		return nil, fmt.Errorf("setup failed: %w", err)
	}
	defer func() {
		if err := lt.scenario.Teardown(ctx); err != nil {
			fmt.Printf("teardown error: %v\n", err)
		}
	}()

	jobs := make(chan struct{}, lt.workers*10)
	results := make(chan *ExecutionResult, lt.workers*10)

	var wg sync.WaitGroup
	for i := 0; i < lt.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				result, err := lt.scenario.Execute(ctx)
				if err != nil {
					result = &ExecutionResult{Success: false, Error: err}
				}
				results <- result
			}
		}()
	}

	go func() {
		for result := range results {
			lt.metrics.Record(result)
		}
	}()

	startTime := time.Now()
	ticker := lt.createTicker()
	defer ticker.Stop()

	executionLoop := func() error {
		for {
			select {
			case <-ctx.Done():
				close(jobs)
				wg.Wait()
				close(results)
				return ctx.Err()
			case <-ticker.C:
				if time.Since(startTime) > lt.profile.Duration {
					close(jobs)
					wg.Wait()
					close(results)
					return nil
				}
				select {
				case jobs <- struct{}{}:
				default:
				}
			}
		}
	}

	if err := executionLoop(); err != nil && err != context.DeadlineExceeded {
		return lt.generateReport(startTime), err
	}

	return lt.generateReport(startTime), nil
}

// createTicker creates a ticker based on the load profile
func (lt *LoadTest) createTicker() *time.Ticker {
	rate := lt.profile.StartRate
	if rate <= 0 {
		rate = 1
	}
	interval := time.Duration(float64(time.Second) / rate)
	return time.NewTicker(interval)
}

// generateReport creates a test report from collected metrics
func (lt *LoadTest) generateReport(startTime time.Time) *TestReport {
	duration := time.Since(startTime)
	return lt.metrics.GenerateReport(lt.name, duration)
}

// TestReport contains aggregated test results
type TestReport struct {
	Name           string
	Duration       time.Duration
	TotalRequests  int64
	SuccessCount   int64
	FailureCount   int64
	AvgLatency     time.Duration
	P50Latency     time.Duration
	P95Latency     time.Duration
	P99Latency     time.Duration
	MaxLatency     time.Duration
	RequestsPerSec float64
	ErrorRate      float64
	Errors         map[string]int
}
