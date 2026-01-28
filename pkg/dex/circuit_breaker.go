// Package dex provides DEX integration for crypto-to-fiat conversions.
//
// VE-905: DEX integration for crypto-to-fiat conversions
package dex

import (
	"sync"
	"sync/atomic"
	"time"
)

// circuitBreaker implements safety circuit breaker functionality
type circuitBreaker struct {
	cfg       CircuitBreakerConfig
	tripped   atomic.Bool
	failures  int64
	successes int64
	lastReset time.Time
	mu        sync.Mutex
}

// newCircuitBreaker creates a new circuit breaker
func newCircuitBreaker(cfg CircuitBreakerConfig) *circuitBreaker {
	return &circuitBreaker{
		cfg:       cfg,
		lastReset: time.Now(),
	}
}

// IsTripped checks if the circuit breaker is tripped
func (cb *circuitBreaker) IsTripped() bool {
	if !cb.cfg.Enabled {
		return false
	}
	return cb.tripped.Load()
}

// RecordFailure records a failure
func (cb *circuitBreaker) RecordFailure() {
	if !cb.cfg.Enabled {
		return
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Reset counters if a minute has passed
	cb.maybeReset()

	cb.failures++
	cb.successes = 0

	if cb.failures >= int64(cb.cfg.MaxFailuresPerMinute) {
		cb.trip()
	}
}

// RecordSuccess records a success
func (cb *circuitBreaker) RecordSuccess() {
	if !cb.cfg.Enabled {
		return
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.maybeReset()

	cb.successes++

	// Check if we should reset after recovery
	if cb.tripped.Load() && cb.successes >= int64(cb.cfg.RecoveryThreshold) {
		cb.reset()
	}
}

// trip trips the circuit breaker
func (cb *circuitBreaker) trip() {
	cb.tripped.Store(true)

	// Schedule automatic reset after cooldown
	go func() {
		time.Sleep(cb.cfg.CooldownPeriod)
		cb.mu.Lock()
		defer cb.mu.Unlock()
		cb.reset()
	}()
}

// reset resets the circuit breaker
func (cb *circuitBreaker) reset() {
	cb.tripped.Store(false)
	cb.failures = 0
	cb.successes = 0
	cb.lastReset = time.Now()
}

// maybeReset resets counters if a minute has passed
func (cb *circuitBreaker) maybeReset() {
	if time.Since(cb.lastReset) > time.Minute {
		cb.failures = 0
		cb.lastReset = time.Now()
	}
}

// CheckPriceDeviation checks if price has deviated too much from TWAP
func (cb *circuitBreaker) CheckPriceDeviation(currentPrice, twapPrice float64) bool {
	if !cb.cfg.Enabled || twapPrice == 0 {
		return false
	}

	deviation := (currentPrice - twapPrice) / twapPrice
	if deviation < 0 {
		deviation = -deviation
	}

	if deviation > cb.cfg.PriceDeviationThreshold {
		cb.mu.Lock()
		cb.trip()
		cb.mu.Unlock()
		return true
	}

	return false
}

// CheckVolumeSpike checks if volume has spiked abnormally
func (cb *circuitBreaker) CheckVolumeSpike(currentVolume, averageVolume float64) bool {
	if !cb.cfg.Enabled || averageVolume == 0 {
		return false
	}

	if currentVolume > averageVolume*cb.cfg.VolumeSpikeFactor {
		cb.mu.Lock()
		cb.trip()
		cb.mu.Unlock()
		return true
	}

	return false
}
