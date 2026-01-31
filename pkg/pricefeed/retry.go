package pricefeed

import (
	"context"
	"math/rand"
	"strings"
	"time"
)

// ============================================================================
// Retry Logic with Exponential Backoff
// ============================================================================

// Retryer provides retry logic with exponential backoff
type Retryer struct {
	maxRetries    int
	initialDelay  time.Duration
	maxDelay      time.Duration
	backoffFactor float64
	retryableErrs []string
}

// NewRetryer creates a new Retryer with the given config
func NewRetryer(cfg RetryConfig) *Retryer {
	return &Retryer{
		maxRetries:    cfg.MaxRetries,
		initialDelay:  cfg.InitialDelay,
		maxDelay:      cfg.MaxDelay,
		backoffFactor: cfg.BackoffFactor,
		retryableErrs: cfg.RetryableErrors,
	}
}

// RetryFunc is a function that can be retried
type RetryFunc func(ctx context.Context) error

// RetryFuncWithResult is a function that returns a result and can be retried
type RetryFuncWithResult[T any] func(ctx context.Context) (T, error)

// Do executes the function with retries
func (r *Retryer) Do(ctx context.Context, fn RetryFunc) error {
	var lastErr error

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := fn(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !r.isRetryable(err) {
			return err
		}

		// Don't sleep after the last attempt
		if attempt < r.maxRetries {
			delay := r.calculateDelay(attempt)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return lastErr
}

// DoWithResult executes the function with retries and returns a result
func DoWithResult[T any](ctx context.Context, r *Retryer, fn RetryFuncWithResult[T]) (T, error) {
	var zero T
	var lastErr error
	var result T

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}

		var err error
		result, err = fn(ctx)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !r.isRetryable(err) {
			return zero, err
		}

		// Don't sleep after the last attempt
		if attempt < r.maxRetries {
			delay := r.calculateDelay(attempt)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return zero, ctx.Err()
			}
		}
	}

	return zero, lastErr
}

// isRetryable checks if an error should be retried
func (r *Retryer) isRetryable(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	for _, retryable := range r.retryableErrs {
		if strings.Contains(errStr, strings.ToLower(retryable)) {
			return true
		}
	}

	return false
}

// calculateDelay calculates the delay for a given attempt with jitter
func (r *Retryer) calculateDelay(attempt int) time.Duration {
	// Exponential backoff
	delay := float64(r.initialDelay) * pow(r.backoffFactor, float64(attempt))

	// Cap at max delay
	if delay > float64(r.maxDelay) {
		delay = float64(r.maxDelay)
	}

	// Add jitter (±25%)
	jitter := delay * 0.25 * (rand.Float64()*2 - 1)
	delay += jitter

	return time.Duration(delay)
}

// pow calculates x^y for float64
func pow(x, y float64) float64 {
	result := 1.0
	for i := 0; i < int(y); i++ {
		result *= x
	}
	return result
}

// ============================================================================
// Circuit Breaker
// ============================================================================

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// CircuitClosed means requests are allowed
	CircuitClosed CircuitState = iota

	// CircuitOpen means requests are blocked
	CircuitOpen

	// CircuitHalfOpen means limited requests are allowed to test recovery
	CircuitHalfOpen
)

// String returns the string representation
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	state            CircuitState
	failureCount     int
	successCount     int
	lastFailure      time.Time
	failureThreshold int
	successThreshold int
	timeout          time.Duration
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            CircuitClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
}

// Allow checks if a request is allowed
func (cb *CircuitBreaker) Allow() bool {
	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// Check if timeout has passed
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.state = CircuitHalfOpen
			cb.successCount = 0
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	switch cb.state {
	case CircuitHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = CircuitClosed
			cb.failureCount = 0
		}
	case CircuitClosed:
		cb.failureCount = 0
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.failureCount++
	cb.lastFailure = time.Now()

	switch cb.state {
	case CircuitHalfOpen:
		cb.state = CircuitOpen
	case CircuitClosed:
		if cb.failureCount >= cb.failureThreshold {
			cb.state = CircuitOpen
		}
	}
}

// State returns the current circuit state
func (cb *CircuitBreaker) State() CircuitState {
	// Check if we should transition from open to half-open
	if cb.state == CircuitOpen && time.Since(cb.lastFailure) > cb.timeout {
		cb.state = CircuitHalfOpen
		cb.successCount = 0
	}
	return cb.state
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.state = CircuitClosed
	cb.failureCount = 0
	cb.successCount = 0
}
