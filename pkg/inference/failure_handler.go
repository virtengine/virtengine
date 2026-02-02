// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0
//
// Package inference provides failure handling for consensus-critical inference.
// These handlers ensure that inference failures do not break blockchain consensus.
//
// VE-219: Deterministic identity verification runtime

package inference

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Failure Handling Strategy
// ============================================================================

// FailureStrategy defines how to handle inference failures.
type FailureStrategy int

const (
	// FailureStrategyRetry retries the inference operation.
	FailureStrategyRetry FailureStrategy = iota

	// FailureStrategyFallback returns a fallback score.
	FailureStrategyFallback

	// FailureStrategyDefer defers verification to the next block.
	FailureStrategyDefer

	// FailureStrategyReject rejects the verification request.
	FailureStrategyReject
)

// String returns the string representation of the failure strategy.
func (fs FailureStrategy) String() string {
	switch fs {
	case FailureStrategyRetry:
		return "retry"
	case FailureStrategyFallback:
		return "fallback"
	case FailureStrategyDefer:
		return "defer"
	case FailureStrategyReject:
		return "reject"
	default:
		return "unknown"
	}
}

// ============================================================================
// Failure Handler
// ============================================================================

// FailureHandlerConfig configures the failure handler.
type FailureHandlerConfig struct {
	// DefaultStrategy is the default failure strategy.
	DefaultStrategy FailureStrategy

	// MaxRetries is the maximum number of retries for FailureStrategyRetry.
	MaxRetries int

	// RetryDelay is the delay between retries.
	RetryDelay time.Duration

	// FallbackScore is the score returned for FailureStrategyFallback.
	FallbackScore uint32

	// FallbackConfidence is the confidence returned for fallback.
	FallbackConfidence float32

	// DeferTimeout is how long to defer before failing.
	DeferTimeout time.Duration

	// TimeoutStrategy is the strategy for timeout failures.
	TimeoutStrategy FailureStrategy

	// ConnectionStrategy is the strategy for connection failures.
	ConnectionStrategy FailureStrategy

	// ValidationStrategy is the strategy for validation failures.
	ValidationStrategy FailureStrategy
}

// DefaultFailureHandlerConfig returns the default failure handler config.
func DefaultFailureHandlerConfig() FailureHandlerConfig {
	return FailureHandlerConfig{
		DefaultStrategy:    FailureStrategyFallback,
		MaxRetries:         2,
		RetryDelay:         100 * time.Millisecond,
		FallbackScore:      0,
		FallbackConfidence: 0.0,
		DeferTimeout:       5 * time.Second,
		TimeoutStrategy:    FailureStrategyFallback,
		ConnectionStrategy: FailureStrategyRetry,
		ValidationStrategy: FailureStrategyReject,
	}
}

// FailureHandler handles inference failures in a consensus-safe manner.
type FailureHandler struct {
	config  FailureHandlerConfig
	metrics *MetricsCollector

	mu            sync.RWMutex
	failureCount  map[string]int
	lastFailure   map[string]time.Time
	deferredQueue []DeferredRequest
}

// DeferredRequest represents a deferred inference request.
type DeferredRequest struct {
	Inputs      *ScoreInputs
	RequestedAt time.Time
	ExpiresAt   time.Time
	Attempts    int
}

// NewFailureHandler creates a new failure handler.
func NewFailureHandler(config FailureHandlerConfig) *FailureHandler {
	return &FailureHandler{
		config:       config,
		metrics:      GetGlobalMetricsCollector(),
		failureCount: make(map[string]int),
		lastFailure:  make(map[string]time.Time),
	}
}

// HandleFailure handles an inference failure and returns the appropriate response.
func (fh *FailureHandler) HandleFailure(
	ctx context.Context,
	err error,
	inputs *ScoreInputs,
	scorer Scorer,
) (*ScoreResult, error) {
	failureType := classifyError(err)
	strategy := fh.getStrategy(failureType)

	fh.recordFailure(inputs.Metadata.RequestID, failureType)

	switch strategy {
	case FailureStrategyRetry:
		return fh.handleRetry(ctx, err, inputs, scorer)

	case FailureStrategyFallback:
		return fh.handleFallback(inputs, err)

	case FailureStrategyDefer:
		return fh.handleDefer(inputs, err)

	case FailureStrategyReject:
		return nil, fmt.Errorf("inference rejected: %w", err)

	default:
		return fh.handleFallback(inputs, err)
	}
}

// handleRetry retries the inference operation.
func (fh *FailureHandler) handleRetry(
	ctx context.Context,
	originalErr error,
	inputs *ScoreInputs,
	scorer Scorer,
) (*ScoreResult, error) {
	lastErr := originalErr

	for attempt := 1; attempt <= fh.config.MaxRetries; attempt++ {
		// Wait before retry
		select {
		case <-ctx.Done():
			return fh.handleFallback(inputs, ctx.Err())
		case <-time.After(fh.config.RetryDelay):
		}

		// Retry inference
		result, err := scorer.ComputeScore(inputs)
		if err == nil {
			return result, nil
		}

		lastErr = err
	}

	// All retries failed, use fallback
	return fh.handleFallback(inputs, lastErr)
}

// handleFallback returns a fallback score.
func (fh *FailureHandler) handleFallback(inputs *ScoreInputs, originalErr error) (*ScoreResult, error) {
	dc := NewDeterminismController(42, true)

	result := &ScoreResult{
		Score:        fh.config.FallbackScore,
		RawScore:     float32(fh.config.FallbackScore),
		Confidence:   fh.config.FallbackConfidence,
		ModelVersion: "fallback",
		ModelHash:    "",
		InputHash:    dc.ComputeInputHash(inputs),
		OutputHash:   dc.ComputeOutputHash([]float32{float32(fh.config.FallbackScore)}),
		ReasonCodes:  []string{ReasonCodeInferenceError},
	}

	if originalErr != nil {
		result.ReasonCodes = append(result.ReasonCodes, "FALLBACK_USED")
	}

	return result, nil
}

// handleDefer defers the request for later processing.
func (fh *FailureHandler) handleDefer(inputs *ScoreInputs, _ error) (*ScoreResult, error) {
	fh.mu.Lock()
	defer fh.mu.Unlock()

	deferred := DeferredRequest{
		Inputs:      inputs,
		RequestedAt: time.Now(),
		ExpiresAt:   time.Now().Add(fh.config.DeferTimeout),
		Attempts:    1,
	}

	fh.deferredQueue = append(fh.deferredQueue, deferred)

	// Return a pending result
	dc := NewDeterminismController(42, true)
	return &ScoreResult{
		Score:        0,
		RawScore:     0,
		Confidence:   0,
		ModelVersion: "deferred",
		InputHash:    dc.ComputeInputHash(inputs),
		ReasonCodes:  []string{"DEFERRED"},
	}, nil
}

// getStrategy returns the appropriate strategy for a failure type.
func (fh *FailureHandler) getStrategy(failureType FailureType) FailureStrategy {
	switch failureType {
	case FailureTypeTimeout:
		return fh.config.TimeoutStrategy
	case FailureTypeConnection:
		return fh.config.ConnectionStrategy
	case FailureTypeValidation:
		return fh.config.ValidationStrategy
	default:
		return fh.config.DefaultStrategy
	}
}

// recordFailure records a failure for tracking.
//
//nolint:unparam // requestID kept for future per-request failure tracking
func (fh *FailureHandler) recordFailure(_ string, failureType FailureType) {
	fh.mu.Lock()
	defer fh.mu.Unlock()

	key := string(failureType)
	fh.failureCount[key]++
	fh.lastFailure[key] = time.Now()
}

// GetDeferredRequests returns pending deferred requests.
func (fh *FailureHandler) GetDeferredRequests() []DeferredRequest {
	fh.mu.RLock()
	defer fh.mu.RUnlock()

	result := make([]DeferredRequest, len(fh.deferredQueue))
	copy(result, fh.deferredQueue)
	return result
}

// ClearExpiredDeferred removes expired deferred requests.
func (fh *FailureHandler) ClearExpiredDeferred() int {
	fh.mu.Lock()
	defer fh.mu.Unlock()

	now := time.Now()
	remaining := make([]DeferredRequest, 0)
	expired := 0

	for _, req := range fh.deferredQueue {
		if req.ExpiresAt.After(now) {
			remaining = append(remaining, req)
		} else {
			expired++
		}
	}

	fh.deferredQueue = remaining
	return expired
}

// ============================================================================
// Error Classification
// ============================================================================

// FailureType classifies the type of failure.
type FailureType string

const (
	FailureTypeTimeout    FailureType = "timeout"
	FailureTypeConnection FailureType = "connection"
	FailureTypeValidation FailureType = "validation"
	FailureTypeModel      FailureType = "model"
	FailureTypeUnknown    FailureType = "unknown"
)

// classifyError classifies an error into a failure type.
func classifyError(err error) FailureType {
	if err == nil {
		return FailureTypeUnknown
	}

	errStr := err.Error()

	// Check for timeout errors
	if containsAny(errStr, "timeout", "deadline exceeded", "context deadline") {
		return FailureTypeTimeout
	}

	// Check for connection errors
	if containsAny(errStr, "connection", "refused", "unavailable", "dial", "network") {
		return FailureTypeConnection
	}

	// Check for validation errors
	if containsAny(errStr, "validation", "invalid", "dimension", "mismatch") {
		return FailureTypeValidation
	}

	// Check for model errors
	if containsAny(errStr, "model", "inference", "tensorflow", "load") {
		return FailureTypeModel
	}

	return FailureTypeUnknown
}

// containsAny checks if s contains any of the substrings.
func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

// ============================================================================
// Consensus-Safe Wrapper
// ============================================================================

// ConsensusSafeScorer wraps a scorer with consensus-safe failure handling.
type ConsensusSafeScorer struct {
	scorer         Scorer
	failureHandler *FailureHandler
}

// NewConsensusSafeScorer creates a new consensus-safe scorer.
func NewConsensusSafeScorer(scorer Scorer, config FailureHandlerConfig) *ConsensusSafeScorer {
	return &ConsensusSafeScorer{
		scorer:         scorer,
		failureHandler: NewFailureHandler(config),
	}
}

// ComputeScore computes a score with consensus-safe failure handling.
func (css *ConsensusSafeScorer) ComputeScore(inputs *ScoreInputs) (*ScoreResult, error) {
	return css.ComputeScoreWithContext(context.Background(), inputs)
}

// ComputeScoreWithContext computes a score with context and failure handling.
func (css *ConsensusSafeScorer) ComputeScoreWithContext(ctx context.Context, inputs *ScoreInputs) (*ScoreResult, error) {
	result, err := css.scorer.ComputeScore(inputs)
	if err != nil {
		return css.failureHandler.HandleFailure(ctx, err, inputs, css.scorer)
	}
	return result, nil
}

// GetModelVersion returns the model version from the underlying scorer.
func (css *ConsensusSafeScorer) GetModelVersion() string {
	return css.scorer.GetModelVersion()
}

// GetModelHash returns the model hash from the underlying scorer.
func (css *ConsensusSafeScorer) GetModelHash() string {
	return css.scorer.GetModelHash()
}

// IsHealthy returns the health status of the underlying scorer.
func (css *ConsensusSafeScorer) IsHealthy() bool {
	return css.scorer.IsHealthy()
}

// Close closes the underlying scorer.
func (css *ConsensusSafeScorer) Close() error {
	return css.scorer.Close()
}
