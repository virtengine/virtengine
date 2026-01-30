// Package enclave_runtime provides TEE enclave implementations.
//
// This file implements the EnclaveManager which orchestrates multiple TEE backends.
// The manager provides:
// - Automatic backend selection based on availability and health
// - Load balancing across multiple enclave instances
// - Failover when an enclave becomes unavailable
// - Unified interface for all enclave operations
// - Health monitoring and metrics collection
//
// Task Reference: VE-2027 - TEE Orchestrator/Manager
package enclave_runtime

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// =============================================================================
// Error Definitions
// =============================================================================

var (
	// ErrNoBackendsAvailable is returned when no healthy backends are available.
	ErrNoBackendsAvailable = errors.New("no enclave backends available")

	// ErrBackendNotFound is returned when the requested backend doesn't exist.
	ErrBackendNotFound = errors.New("enclave backend not found")

	// ErrBackendAlreadyRegistered is returned when attempting to register a duplicate backend.
	ErrBackendAlreadyRegistered = errors.New("enclave backend already registered")

	// ErrManagerNotRunning is returned when operations are attempted on a stopped manager.
	ErrManagerNotRunning = errors.New("enclave manager not running")

	// ErrAllRetriesFailed is returned when all retry attempts have been exhausted.
	ErrAllRetriesFailed = errors.New("all retry attempts failed")

	// ErrCircuitOpen is returned when the circuit breaker is open for a backend.
	ErrCircuitOpen = errors.New("circuit breaker is open")

	// ErrDuplicateRequest is returned when a duplicate request is detected.
	ErrDuplicateRequest = errors.New("duplicate request detected")
)

// =============================================================================
// Types and Constants
// =============================================================================

// HealthStatus represents enclave health state.
type HealthStatus int

const (
	// HealthUnknown indicates the health status is not yet known.
	HealthUnknown HealthStatus = iota
	// HealthHealthy indicates the backend is healthy and available.
	HealthHealthy
	// HealthDegraded indicates the backend is experiencing issues but still functional.
	HealthDegraded
	// HealthUnhealthy indicates the backend is not available for requests.
	HealthUnhealthy
)

// String returns the string representation of the health status.
func (h HealthStatus) String() string {
	switch h {
	case HealthHealthy:
		return "HEALTHY"
	case HealthDegraded:
		return "DEGRADED"
	case HealthUnhealthy:
		return "UNHEALTHY"
	default:
		return "UNKNOWN"
	}
}

// SelectionStrategy defines how backends are selected.
type SelectionStrategy int

const (
	// StrategyPriority uses the highest priority available backend.
	StrategyPriority SelectionStrategy = iota
	// StrategyRoundRobin rotates through healthy backends.
	StrategyRoundRobin
	// StrategyLeastLoaded uses the backend with fewest active requests.
	StrategyLeastLoaded
	// StrategyWeighted performs weighted random selection.
	StrategyWeighted
	// StrategyLatency uses the backend with lowest average latency.
	StrategyLatency
)

// String returns the string representation of the selection strategy.
func (s SelectionStrategy) String() string {
	switch s {
	case StrategyPriority:
		return "PRIORITY"
	case StrategyRoundRobin:
		return "ROUND_ROBIN"
	case StrategyLeastLoaded:
		return "LEAST_LOADED"
	case StrategyWeighted:
		return "WEIGHTED"
	case StrategyLatency:
		return "LATENCY"
	default:
		return "UNKNOWN"
	}
}

// BackendMetrics tracks enclave performance.
type BackendMetrics struct {
	TotalRequests      uint64
	SuccessfulRequests uint64
	FailedRequests     uint64
	TotalLatencyMs     uint64
	AverageLatencyMs   float64
	LastRequestTime    time.Time
}

// Clone returns a copy of the metrics.
func (m *BackendMetrics) Clone() BackendMetrics {
	return BackendMetrics{
		TotalRequests:      m.TotalRequests,
		SuccessfulRequests: m.SuccessfulRequests,
		FailedRequests:     m.FailedRequests,
		TotalLatencyMs:     m.TotalLatencyMs,
		AverageLatencyMs:   m.AverageLatencyMs,
		LastRequestTime:    m.LastRequestTime,
	}
}

// CircuitBreakerState represents the state of a circuit breaker.
type CircuitBreakerState int

const (
	// CircuitClosed allows requests to pass through.
	CircuitClosed CircuitBreakerState = iota
	// CircuitOpen blocks all requests.
	CircuitOpen
	// CircuitHalfOpen allows limited requests to test recovery.
	CircuitHalfOpen
)

// String returns the string representation of the circuit breaker state.
func (c CircuitBreakerState) String() string {
	switch c {
	case CircuitClosed:
		return "CLOSED"
	case CircuitOpen:
		return "OPEN"
	case CircuitHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// EnclaveBackend represents a registered enclave instance.
type EnclaveBackend struct {
	ID              string
	Type            AttestationType
	Service         EnclaveService
	Priority        int // Lower = higher priority
	Weight          int // For weighted load balancing
	MaxConcurrent   int
	Health          HealthStatus
	LastHealthCheck time.Time
	Metrics         BackendMetrics

	// Internal state
	mu                   sync.RWMutex
	activeRequests       int32
	consecutiveFailures  int
	consecutiveSuccesses int
	circuitState         CircuitBreakerState
	circuitOpenTime      time.Time
}

// NewEnclaveBackend creates a new enclave backend with default settings.
func NewEnclaveBackend(id string, backendType AttestationType, service EnclaveService) *EnclaveBackend {
	return &EnclaveBackend{
		ID:            id,
		Type:          backendType,
		Service:       service,
		Priority:      100,
		Weight:        1,
		MaxConcurrent: 10,
		Health:        HealthUnknown,
		circuitState:  CircuitClosed,
	}
}

// GetActiveRequests returns the current number of active requests.
func (b *EnclaveBackend) GetActiveRequests() int {
	return int(atomic.LoadInt32(&b.activeRequests))
}

// IncrementActive increments the active request count.
func (b *EnclaveBackend) IncrementActive() {
	atomic.AddInt32(&b.activeRequests, 1)
}

// DecrementActive decrements the active request count.
func (b *EnclaveBackend) DecrementActive() {
	atomic.AddInt32(&b.activeRequests, -1)
}

// IsAvailable returns true if the backend can accept requests.
func (b *EnclaveBackend) IsAvailable() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.Health == HealthUnhealthy {
		return false
	}

	if b.circuitState == CircuitOpen {
		return false
	}

	if int(atomic.LoadInt32(&b.activeRequests)) >= b.MaxConcurrent {
		return false
	}

	return true
}

// RecordSuccess records a successful request.
func (b *EnclaveBackend) RecordSuccess(latencyMs int64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.Metrics.TotalRequests++
	b.Metrics.SuccessfulRequests++
	b.Metrics.TotalLatencyMs += uint64(latencyMs)
	b.Metrics.LastRequestTime = time.Now()

	if b.Metrics.TotalRequests > 0 {
		b.Metrics.AverageLatencyMs = float64(b.Metrics.TotalLatencyMs) / float64(b.Metrics.TotalRequests)
	}

	b.consecutiveSuccesses++
	b.consecutiveFailures = 0

	// Reset circuit breaker on success in half-open state
	if b.circuitState == CircuitHalfOpen {
		b.circuitState = CircuitClosed
	}
}

// RecordFailure records a failed request.
func (b *EnclaveBackend) RecordFailure(latencyMs int64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.Metrics.TotalRequests++
	b.Metrics.FailedRequests++
	b.Metrics.TotalLatencyMs += uint64(latencyMs)
	b.Metrics.LastRequestTime = time.Now()

	if b.Metrics.TotalRequests > 0 {
		b.Metrics.AverageLatencyMs = float64(b.Metrics.TotalLatencyMs) / float64(b.Metrics.TotalRequests)
	}

	b.consecutiveFailures++
	b.consecutiveSuccesses = 0
}

// EnclaveManagerConfig configures the manager.
type EnclaveManagerConfig struct {
	SelectionStrategy   SelectionStrategy
	HealthCheckInterval time.Duration
	UnhealthyThreshold  int // Consecutive failures before marking unhealthy
	RecoveryThreshold   int // Consecutive successes before marking healthy
	RequestTimeout      time.Duration
	EnableFailover      bool
	MaxRetries          int
	RetryBackoff        time.Duration
	CircuitBreakerReset time.Duration // Time before attempting to close circuit
}

// DefaultEnclaveManagerConfig returns the default manager configuration.
func DefaultEnclaveManagerConfig() EnclaveManagerConfig {
	return EnclaveManagerConfig{
		SelectionStrategy:   StrategyPriority,
		HealthCheckInterval: 30 * time.Second,
		UnhealthyThreshold:  3,
		RecoveryThreshold:   2,
		RequestTimeout:      5 * time.Second,
		EnableFailover:      true,
		MaxRetries:          3,
		RetryBackoff:        100 * time.Millisecond,
		CircuitBreakerReset: 30 * time.Second,
	}
}

// Validate validates the configuration.
func (c EnclaveManagerConfig) Validate() error {
	if c.HealthCheckInterval < time.Second {
		return errors.New("health check interval must be at least 1 second")
	}
	if c.UnhealthyThreshold < 1 {
		return errors.New("unhealthy threshold must be at least 1")
	}
	if c.RecoveryThreshold < 1 {
		return errors.New("recovery threshold must be at least 1")
	}
	if c.RequestTimeout < 100*time.Millisecond {
		return errors.New("request timeout must be at least 100ms")
	}
	if c.MaxRetries < 0 {
		return errors.New("max retries cannot be negative")
	}
	return nil
}

// ManagerStatus represents the aggregate status of all backends.
type ManagerStatus struct {
	Running           bool
	TotalBackends     int
	HealthyBackends   int
	DegradedBackends  int
	UnhealthyBackends int
	TotalRequests     uint64
	SuccessRate       float64
	Backends          []BackendStatus
}

// BackendStatus represents the status of a single backend.
type BackendStatus struct {
	ID               string
	Type             AttestationType
	Health           HealthStatus
	CircuitState     CircuitBreakerState
	ActiveRequests   int
	TotalRequests    uint64
	SuccessRate      float64
	AverageLatencyMs float64
	LastHealthCheck  time.Time
}

// =============================================================================
// EnclaveManager Implementation
// =============================================================================

// EnclaveManager orchestrates multiple TEE backends with failover and load balancing.
type EnclaveManager struct {
	mu       sync.RWMutex
	config   EnclaveManagerConfig
	backends map[string]*EnclaveBackend
	running  bool

	// Hardware capabilities
	hardwareCapabilities *HardwareCapabilities

	// Round-robin state
	rrIndex int
	rrMu    sync.Mutex

	// Request deduplication
	pendingRequests map[string]struct{}
	pendingMu       sync.Mutex

	// Health monitoring
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup

	// Random source for weighted selection
	rng *rand.Rand
}

// NewEnclaveManager creates a new enclave manager.
func NewEnclaveManager(config EnclaveManagerConfig) (*EnclaveManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Detect hardware capabilities
	caps := DetectHardware()

	return &EnclaveManager{
		config:               config,
		backends:             make(map[string]*EnclaveBackend),
		pendingRequests:      make(map[string]struct{}),
		stopCh:               make(chan struct{}),
		rng:                  rand.New(rand.NewSource(time.Now().UnixNano())),
		hardwareCapabilities: &caps,
	}, nil
}

// RegisterBackend adds an enclave backend to the manager.
func (m *EnclaveManager) RegisterBackend(backend *EnclaveBackend) error {
	if backend == nil {
		return errors.New("backend cannot be nil")
	}
	if backend.ID == "" {
		return errors.New("backend ID required")
	}
	if backend.Service == nil {
		return errors.New("backend service required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.backends[backend.ID]; exists {
		return ErrBackendAlreadyRegistered
	}

	// Set defaults if not provided
	if backend.MaxConcurrent <= 0 {
		backend.MaxConcurrent = 10
	}
	if backend.Weight <= 0 {
		backend.Weight = 1
	}

	m.backends[backend.ID] = backend
	return nil
}

// UnregisterBackend removes a backend from the manager.
func (m *EnclaveManager) UnregisterBackend(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.backends[id]; !exists {
		return ErrBackendNotFound
	}

	delete(m.backends, id)
	return nil
}

// GetBackend returns a specific backend by ID.
func (m *EnclaveManager) GetBackend(id string) (*EnclaveBackend, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	backend, exists := m.backends[id]
	if !exists {
		return nil, ErrBackendNotFound
	}

	return backend, nil
}

// ListBackends returns all registered backends.
func (m *EnclaveManager) ListBackends() []*EnclaveBackend {
	m.mu.RLock()
	defer m.mu.RUnlock()

	backends := make([]*EnclaveBackend, 0, len(m.backends))
	for _, b := range m.backends {
		backends = append(backends, b)
	}
	return backends
}

// SelectBackend selects the best available backend based on the configured strategy.
// Backends with real hardware are preferred over simulation backends.
func (m *EnclaveManager) SelectBackend() (*EnclaveBackend, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	available := m.getAvailableBackends()
	if len(available) == 0 {
		return nil, ErrNoBackendsAvailable
	}

	// Separate hardware-enabled and simulation backends
	var hardwareBackends, simulationBackends []*EnclaveBackend
	for _, b := range available {
		if m.isBackendHardwareEnabled(b) {
			hardwareBackends = append(hardwareBackends, b)
		} else {
			simulationBackends = append(simulationBackends, b)
		}
	}

	// Prefer hardware backends
	selectionPool := hardwareBackends
	if len(selectionPool) == 0 {
		selectionPool = simulationBackends
	}

	switch m.config.SelectionStrategy {
	case StrategyPriority:
		return m.selectByPriority(selectionPool), nil
	case StrategyRoundRobin:
		return m.selectRoundRobin(selectionPool), nil
	case StrategyLeastLoaded:
		return m.selectLeastLoaded(selectionPool), nil
	case StrategyWeighted:
		return m.selectWeighted(selectionPool), nil
	case StrategyLatency:
		return m.selectByLatency(selectionPool), nil
	default:
		return m.selectByPriority(selectionPool), nil
	}
}

// isBackendHardwareEnabled checks if a backend is using real hardware
func (m *EnclaveManager) isBackendHardwareEnabled(backend *EnclaveBackend) bool {
	// Check if the service implements IsHardwareEnabled
	type hardwareChecker interface {
		IsHardwareEnabled() bool
	}
	if hc, ok := backend.Service.(hardwareChecker); ok {
		return hc.IsHardwareEnabled()
	}
	// Non-simulated types without the interface are assumed to be hardware-enabled
	return backend.Type != AttestationTypeSimulated
}

// GetHardwareCapabilities returns the detected hardware capabilities
func (m *EnclaveManager) GetHardwareCapabilities() HardwareCapabilities {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.hardwareCapabilities == nil {
		return HardwareCapabilities{}
	}
	return *m.hardwareCapabilities
}

// RefreshHardwareCapabilities re-detects hardware capabilities
func (m *EnclaveManager) RefreshHardwareCapabilities() HardwareCapabilities {
	caps := RefreshHardwareDetection()
	m.mu.Lock()
	m.hardwareCapabilities = &caps
	m.mu.Unlock()
	return caps
}

// GetPreferredBackendType returns the preferred backend type based on hardware detection
func (m *EnclaveManager) GetPreferredBackendType() AttestationType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.hardwareCapabilities == nil {
		return AttestationTypeSimulated
	}
	return m.hardwareCapabilities.PreferredBackend
}

// LogHardwareStatus logs the hardware status of all backends
func (m *EnclaveManager) LogHardwareStatus() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fmt.Printf("=== Enclave Manager Hardware Status ===\n")
	if m.hardwareCapabilities != nil {
		fmt.Printf("Hardware Detection: %s\n", m.hardwareCapabilities.String())
		fmt.Printf("Preferred Backend: %s\n", m.hardwareCapabilities.PreferredBackend)
	}
	fmt.Printf("Registered Backends:\n")
	for id, b := range m.backends {
		hwEnabled := m.isBackendHardwareEnabled(b)
		mode := "simulation"
		if hwEnabled {
			mode = "hardware"
		}
		fmt.Printf("  - %s (%s): %s mode, health=%s\n", id, b.Type, mode, b.Health)
	}
	fmt.Printf("========================================\n")
}

// getAvailableBackends returns all backends that can accept requests.
func (m *EnclaveManager) getAvailableBackends() []*EnclaveBackend {
	available := make([]*EnclaveBackend, 0)
	for _, b := range m.backends {
		if b.IsAvailable() {
			available = append(available, b)
		}
	}
	return available
}

// selectByPriority returns the backend with the lowest priority number.
func (m *EnclaveManager) selectByPriority(backends []*EnclaveBackend) *EnclaveBackend {
	sort.Slice(backends, func(i, j int) bool {
		return backends[i].Priority < backends[j].Priority
	})
	return backends[0]
}

// selectRoundRobin returns the next backend in rotation.
func (m *EnclaveManager) selectRoundRobin(backends []*EnclaveBackend) *EnclaveBackend {
	m.rrMu.Lock()
	defer m.rrMu.Unlock()

	// Sort by ID for consistent ordering
	sort.Slice(backends, func(i, j int) bool {
		return backends[i].ID < backends[j].ID
	})

	if m.rrIndex >= len(backends) {
		m.rrIndex = 0
	}

	backend := backends[m.rrIndex]
	m.rrIndex = (m.rrIndex + 1) % len(backends)
	return backend
}

// selectLeastLoaded returns the backend with fewest active requests.
func (m *EnclaveManager) selectLeastLoaded(backends []*EnclaveBackend) *EnclaveBackend {
	var selected *EnclaveBackend
	minLoad := int(^uint(0) >> 1) // Max int

	for _, b := range backends {
		active := b.GetActiveRequests()
		if active < minLoad {
			minLoad = active
			selected = b
		}
	}

	return selected
}

// selectWeighted performs weighted random selection.
func (m *EnclaveManager) selectWeighted(backends []*EnclaveBackend) *EnclaveBackend {
	totalWeight := 0
	for _, b := range backends {
		totalWeight += b.Weight
	}

	if totalWeight == 0 {
		return backends[0]
	}

	r := m.rng.Intn(totalWeight)
	cumulative := 0

	for _, b := range backends {
		cumulative += b.Weight
		if r < cumulative {
			return b
		}
	}

	return backends[len(backends)-1]
}

// selectByLatency returns the backend with the lowest average latency.
func (m *EnclaveManager) selectByLatency(backends []*EnclaveBackend) *EnclaveBackend {
	var selected *EnclaveBackend
	minLatency := float64(^uint64(0) >> 1) // Max float

	for _, b := range backends {
		b.mu.RLock()
		latency := b.Metrics.AverageLatencyMs
		// For backends with no metrics, use a high default
		if b.Metrics.TotalRequests == 0 {
			latency = 1000.0
		}
		b.mu.RUnlock()

		if latency < minLatency {
			minLatency = latency
			selected = b
		}
	}

	return selected
}

// Score routes a scoring request to the best available backend with failover.
func (m *EnclaveManager) Score(ctx context.Context, request *ScoringRequest) (*ScoringResult, error) {
	if !m.isRunning() {
		return nil, ErrManagerNotRunning
	}

	// Request deduplication
	requestHash := m.hashRequest(request)
	if !m.acquireRequest(requestHash) {
		return nil, ErrDuplicateRequest
	}
	defer m.releaseRequest(requestHash)

	var lastErr error
	retriesLeft := m.config.MaxRetries
	backoff := m.config.RetryBackoff

	// Track tried backends to avoid retrying the same one
	triedBackends := make(map[string]bool)

	for retriesLeft >= 0 {
		backend, err := m.selectBackendExcluding(triedBackends)
		if err != nil {
			if lastErr != nil {
				return nil, fmt.Errorf("%w: last error: %v", ErrAllRetriesFailed, lastErr)
			}
			return nil, err
		}

		triedBackends[backend.ID] = true
		backend.IncrementActive()

		// Create timeout context
		timeoutCtx, cancel := context.WithTimeout(ctx, m.config.RequestTimeout)

		startTime := time.Now()
		result, err := backend.Service.Score(timeoutCtx, request)
		latency := time.Since(startTime).Milliseconds()

		cancel()
		backend.DecrementActive()

		if err == nil && result.IsSuccess() {
			backend.RecordSuccess(latency)
			m.updateHealth(backend)
			return result, nil
		}

		// Record failure
		if err != nil {
			lastErr = err
		} else if result.Error != "" {
			lastErr = errors.New(result.Error)
		}

		backend.RecordFailure(latency)
		m.updateHealth(backend)
		m.checkCircuitBreaker(backend)

		// Check if failover is enabled
		if !m.config.EnableFailover {
			return result, err
		}

		retriesLeft--
		if retriesLeft >= 0 {
			// Exponential backoff
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
			}
		}
	}

	return nil, fmt.Errorf("%w: %v", ErrAllRetriesFailed, lastErr)
}

// selectBackendExcluding selects a backend excluding already-tried ones.
func (m *EnclaveManager) selectBackendExcluding(excluded map[string]bool) (*EnclaveBackend, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	available := make([]*EnclaveBackend, 0)
	for _, b := range m.backends {
		if !excluded[b.ID] && b.IsAvailable() {
			available = append(available, b)
		}
	}

	if len(available) == 0 {
		return nil, ErrNoBackendsAvailable
	}

	switch m.config.SelectionStrategy {
	case StrategyPriority:
		return m.selectByPriority(available), nil
	case StrategyRoundRobin:
		return m.selectRoundRobin(available), nil
	case StrategyLeastLoaded:
		return m.selectLeastLoaded(available), nil
	case StrategyWeighted:
		return m.selectWeighted(available), nil
	case StrategyLatency:
		return m.selectByLatency(available), nil
	default:
		return m.selectByPriority(available), nil
	}
}

// GenerateAttestation generates an attestation from any available backend.
func (m *EnclaveManager) GenerateAttestation(reportData []byte) ([]byte, error) {
	if !m.isRunning() {
		return nil, ErrManagerNotRunning
	}

	backend, err := m.SelectBackend()
	if err != nil {
		return nil, err
	}

	return backend.Service.GenerateAttestation(reportData)
}

// GetStatus returns the aggregate status of all backends.
func (m *EnclaveManager) GetStatus() ManagerStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := ManagerStatus{
		Running:       m.running,
		TotalBackends: len(m.backends),
		Backends:      make([]BackendStatus, 0, len(m.backends)),
	}

	var totalRequests, successfulRequests uint64

	for _, b := range m.backends {
		b.mu.RLock()

		switch b.Health {
		case HealthHealthy:
			status.HealthyBackends++
		case HealthDegraded:
			status.DegradedBackends++
		case HealthUnhealthy:
			status.UnhealthyBackends++
		}

		totalRequests += b.Metrics.TotalRequests
		successfulRequests += b.Metrics.SuccessfulRequests

		var successRate float64
		if b.Metrics.TotalRequests > 0 {
			successRate = float64(b.Metrics.SuccessfulRequests) / float64(b.Metrics.TotalRequests) * 100
		}

		status.Backends = append(status.Backends, BackendStatus{
			ID:               b.ID,
			Type:             b.Type,
			Health:           b.Health,
			CircuitState:     b.circuitState,
			ActiveRequests:   int(atomic.LoadInt32(&b.activeRequests)),
			TotalRequests:    b.Metrics.TotalRequests,
			SuccessRate:      successRate,
			AverageLatencyMs: b.Metrics.AverageLatencyMs,
			LastHealthCheck:  b.LastHealthCheck,
		})

		b.mu.RUnlock()
	}

	status.TotalRequests = totalRequests
	if totalRequests > 0 {
		status.SuccessRate = float64(successfulRequests) / float64(totalRequests) * 100
	}

	return status
}

// Start starts the health monitoring background goroutine.
func (m *EnclaveManager) Start() error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return errors.New("manager already running")
	}
	m.running = true
	m.mu.Unlock()

	// Do initial health check
	m.performHealthChecks()

	// Start health monitoring goroutine
	m.wg.Add(1)
	go m.healthMonitorLoop()

	return nil
}

// Stop stops the manager and cleans up resources.
func (m *EnclaveManager) Stop() error {
	m.stopOnce.Do(func() {
		m.mu.Lock()
		m.running = false
		m.mu.Unlock()

		close(m.stopCh)
		m.wg.Wait()
	})
	return nil
}

// isRunning returns true if the manager is running.
func (m *EnclaveManager) isRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// healthMonitorLoop is the background health monitoring goroutine.
func (m *EnclaveManager) healthMonitorLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.performHealthChecks()
			m.resetCircuitBreakers()
		}
	}
}

// performHealthChecks checks the health of all backends.
func (m *EnclaveManager) performHealthChecks() {
	m.mu.RLock()
	backends := make([]*EnclaveBackend, 0, len(m.backends))
	for _, b := range m.backends {
		backends = append(backends, b)
	}
	m.mu.RUnlock()

	for _, backend := range backends {
		m.checkBackendHealth(backend)
	}
}

// checkBackendHealth checks a single backend's health.
func (m *EnclaveManager) checkBackendHealth(backend *EnclaveBackend) {
	status := backend.Service.GetStatus()

	backend.mu.Lock()
	defer backend.mu.Unlock()

	backend.LastHealthCheck = time.Now()

	if !status.Available {
		backend.consecutiveFailures++
		backend.consecutiveSuccesses = 0
	} else {
		backend.consecutiveSuccesses++
		backend.consecutiveFailures = 0
	}

	// Update health based on thresholds
	if backend.consecutiveFailures >= m.config.UnhealthyThreshold {
		backend.Health = HealthUnhealthy
	} else if backend.consecutiveSuccesses >= m.config.RecoveryThreshold {
		backend.Health = HealthHealthy
	} else if backend.consecutiveFailures > 0 {
		backend.Health = HealthDegraded
	} else if status.Available {
		if backend.Health == HealthUnknown {
			backend.Health = HealthHealthy
		}
	}
}

// updateHealth updates a backend's health based on recent request results.
func (m *EnclaveManager) updateHealth(backend *EnclaveBackend) {
	backend.mu.Lock()
	defer backend.mu.Unlock()

	if backend.consecutiveFailures >= m.config.UnhealthyThreshold {
		backend.Health = HealthUnhealthy
	} else if backend.consecutiveSuccesses >= m.config.RecoveryThreshold {
		backend.Health = HealthHealthy
	} else if backend.consecutiveFailures > 0 && backend.Health == HealthHealthy {
		backend.Health = HealthDegraded
	}
}

// checkCircuitBreaker updates the circuit breaker state based on failures.
func (m *EnclaveManager) checkCircuitBreaker(backend *EnclaveBackend) {
	backend.mu.Lock()
	defer backend.mu.Unlock()

	if backend.consecutiveFailures >= m.config.UnhealthyThreshold {
		if backend.circuitState != CircuitOpen {
			backend.circuitState = CircuitOpen
			backend.circuitOpenTime = time.Now()
		}
	}
}

// resetCircuitBreakers checks if any open circuits should be reset to half-open.
func (m *EnclaveManager) resetCircuitBreakers() {
	m.mu.RLock()
	backends := make([]*EnclaveBackend, 0, len(m.backends))
	for _, b := range m.backends {
		backends = append(backends, b)
	}
	m.mu.RUnlock()

	now := time.Now()

	for _, backend := range backends {
		backend.mu.Lock()
		if backend.circuitState == CircuitOpen {
			if now.Sub(backend.circuitOpenTime) >= m.config.CircuitBreakerReset {
				backend.circuitState = CircuitHalfOpen
				backend.consecutiveFailures = 0
			}
		}
		backend.mu.Unlock()
	}
}

// hashRequest creates a hash for request deduplication.
func (m *EnclaveManager) hashRequest(request *ScoringRequest) string {
	h := sha256.New()
	h.Write([]byte(request.RequestID))
	h.Write([]byte(request.ScopeID))
	h.Write([]byte(request.AccountAddress))
	return fmt.Sprintf("%x", h.Sum(nil)[:16])
}

// acquireRequest attempts to acquire a request slot (for deduplication).
func (m *EnclaveManager) acquireRequest(hash string) bool {
	m.pendingMu.Lock()
	defer m.pendingMu.Unlock()

	if _, exists := m.pendingRequests[hash]; exists {
		return false
	}

	m.pendingRequests[hash] = struct{}{}
	return true
}

// releaseRequest releases a request slot.
func (m *EnclaveManager) releaseRequest(hash string) {
	m.pendingMu.Lock()
	defer m.pendingMu.Unlock()
	delete(m.pendingRequests, hash)
}

// =============================================================================
// Convenience Functions
// =============================================================================

// CreateSimulatedBackend creates a backend using the simulated enclave service.
// This is useful for testing and development.
func CreateSimulatedBackend(id string) (*EnclaveBackend, error) {
	service := NewSimulatedEnclaveService()
	if err := service.Initialize(DefaultRuntimeConfig()); err != nil {
		return nil, fmt.Errorf("failed to initialize simulated service: %w", err)
	}

	backend := NewEnclaveBackend(id, AttestationTypeSimulated, service)
	backend.Priority = 1000 // Low priority for simulation
	backend.Health = HealthHealthy

	return backend, nil
}

// CreateDefaultManager creates a manager with default configuration
// and a simulated backend for testing.
func CreateDefaultManager() (*EnclaveManager, error) {
	manager, err := NewEnclaveManager(DefaultEnclaveManagerConfig())
	if err != nil {
		return nil, err
	}

	// Add a simulated backend for basic functionality
	simBackend, err := CreateSimulatedBackend("simulated-default")
	if err != nil {
		return nil, err
	}

	if err := manager.RegisterBackend(simBackend); err != nil {
		return nil, err
	}

	return manager, nil
}
