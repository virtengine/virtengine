// Package enclave_runtime provides TEE enclave implementations.
//
// This file contains tests for the EnclaveManager which orchestrates multiple TEE backends.
//
// Task Reference: VE-2027 - TEE Orchestrator/Manager
package enclave_runtime

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// Test Helpers
// =============================================================================

// mockEnclaveService is a configurable mock for testing.
type mockEnclaveService struct {
	mu           sync.RWMutex
	initialized  bool
	available    bool
	scoreFn      func(ctx context.Context, request *ScoringRequest) (*ScoringResult, error)
	scoreLatency time.Duration
	shouldFail   bool
	failCount    int32
	callCount    int32
}

func newMockEnclaveService() *mockEnclaveService {
	return &mockEnclaveService{
		available: true,
	}
}

func (m *mockEnclaveService) Initialize(config RuntimeConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.initialized = true
	return nil
}

func (m *mockEnclaveService) Score(ctx context.Context, request *ScoringRequest) (*ScoringResult, error) {
	atomic.AddInt32(&m.callCount, 1)

	if m.scoreLatency > 0 {
		select {
		case <-time.After(m.scoreLatency):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.shouldFail {
		atomic.AddInt32(&m.failCount, 1)
		return nil, errors.New("mock failure")
	}

	if m.scoreFn != nil {
		return m.scoreFn(ctx, request)
	}

	return &ScoringResult{
		RequestID: request.RequestID,
		Score:     75,
		Status:    "verified",
	}, nil
}

func (m *mockEnclaveService) GetMeasurement() ([]byte, error) {
	return []byte("mock-measurement"), nil
}

func (m *mockEnclaveService) GetEncryptionPubKey() ([]byte, error) {
	return []byte("mock-enc-key"), nil
}

func (m *mockEnclaveService) GetSigningPubKey() ([]byte, error) {
	return []byte("mock-sign-key"), nil
}

func (m *mockEnclaveService) GenerateAttestation(reportData []byte) ([]byte, error) {
	return append([]byte("mock-attestation-"), reportData...), nil
}

func (m *mockEnclaveService) RotateKeys() error {
	return nil
}

func (m *mockEnclaveService) GetStatus() EnclaveStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return EnclaveStatus{
		Initialized: m.initialized,
		Available:   m.available,
	}
}

func (m *mockEnclaveService) Shutdown() error {
	return nil
}

func (m *mockEnclaveService) setAvailable(available bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.available = available
}

func (m *mockEnclaveService) getCallCount() int32 {
	return atomic.LoadInt32(&m.callCount)
}

func createTestRequest(id string) *ScoringRequest {
	return &ScoringRequest{
		RequestID:      id,
		Ciphertext:     []byte("test-ciphertext"),
		WrappedKey:     []byte("test-wrapped-key"),
		Nonce:          []byte("test-nonce"),
		ScopeID:        "scope-1",
		AccountAddress: "virtengine1test",
	}
}

func createTestBackend(id string, priority int) *EnclaveBackend {
	svc := newMockEnclaveService()
	svc.initialized = true

	backend := NewEnclaveBackend(id, AttestationTypeSimulated, svc)
	backend.Priority = priority
	backend.Health = HealthHealthy
	return backend
}

func createTestManager(t *testing.T) *EnclaveManager {
	t.Helper()
	config := DefaultEnclaveManagerConfig()
	config.HealthCheckInterval = 1 * time.Second // Minimum allowed
	config.UnhealthyThreshold = 2
	config.RecoveryThreshold = 1
	config.RequestTimeout = 1 * time.Second

	manager, err := NewEnclaveManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	return manager
}

// =============================================================================
// Configuration Tests
// =============================================================================

func TestEnclaveManagerConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultEnclaveManagerConfig()

		if config.SelectionStrategy != StrategyPriority {
			t.Errorf("expected StrategyPriority, got %v", config.SelectionStrategy)
		}
		if config.HealthCheckInterval != 30*time.Second {
			t.Errorf("expected 30s health check interval, got %v", config.HealthCheckInterval)
		}
		if config.UnhealthyThreshold != 3 {
			t.Errorf("expected unhealthy threshold 3, got %d", config.UnhealthyThreshold)
		}
		if config.RecoveryThreshold != 2 {
			t.Errorf("expected recovery threshold 2, got %d", config.RecoveryThreshold)
		}
		if config.MaxRetries != 3 {
			t.Errorf("expected max retries 3, got %d", config.MaxRetries)
		}
		if !config.EnableFailover {
			t.Error("expected failover to be enabled")
		}
	})

	t.Run("ValidateConfig", func(t *testing.T) {
		tests := []struct {
			name      string
			modify    func(*EnclaveManagerConfig)
			expectErr bool
		}{
			{
				name:      "ValidDefault",
				modify:    func(c *EnclaveManagerConfig) {},
				expectErr: false,
			},
			{
				name: "InvalidHealthCheckInterval",
				modify: func(c *EnclaveManagerConfig) {
					c.HealthCheckInterval = 100 * time.Millisecond
				},
				expectErr: true,
			},
			{
				name: "InvalidUnhealthyThreshold",
				modify: func(c *EnclaveManagerConfig) {
					c.UnhealthyThreshold = 0
				},
				expectErr: true,
			},
			{
				name: "InvalidRecoveryThreshold",
				modify: func(c *EnclaveManagerConfig) {
					c.RecoveryThreshold = 0
				},
				expectErr: true,
			},
			{
				name: "InvalidRequestTimeout",
				modify: func(c *EnclaveManagerConfig) {
					c.RequestTimeout = 10 * time.Millisecond
				},
				expectErr: true,
			},
			{
				name: "InvalidMaxRetries",
				modify: func(c *EnclaveManagerConfig) {
					c.MaxRetries = -1
				},
				expectErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				config := DefaultEnclaveManagerConfig()
				tt.modify(&config)
				err := config.Validate()
				if tt.expectErr && err == nil {
					t.Error("expected error, got nil")
				}
				if !tt.expectErr && err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			})
		}
	})
}

// =============================================================================
// Backend Registration Tests
// =============================================================================

func TestBackendRegistration(t *testing.T) {
	t.Run("RegisterBackend", func(t *testing.T) {
		manager := createTestManager(t)
		backend := createTestBackend("backend-1", 10)

		err := manager.RegisterBackend(backend)
		if err != nil {
			t.Fatalf("failed to register backend: %v", err)
		}

		backends := manager.ListBackends()
		if len(backends) != 1 {
			t.Errorf("expected 1 backend, got %d", len(backends))
		}
	})

	t.Run("RegisterDuplicateBackend", func(t *testing.T) {
		manager := createTestManager(t)
		backend1 := createTestBackend("backend-1", 10)
		backend2 := createTestBackend("backend-1", 20)

		if err := manager.RegisterBackend(backend1); err != nil {
			t.Fatalf("failed to register first backend: %v", err)
		}

		err := manager.RegisterBackend(backend2)
		if !errors.Is(err, ErrBackendAlreadyRegistered) {
			t.Errorf("expected ErrBackendAlreadyRegistered, got %v", err)
		}
	})

	t.Run("RegisterNilBackend", func(t *testing.T) {
		manager := createTestManager(t)
		err := manager.RegisterBackend(nil)
		if err == nil {
			t.Error("expected error for nil backend")
		}
	})

	t.Run("RegisterBackendWithoutID", func(t *testing.T) {
		manager := createTestManager(t)
		backend := createTestBackend("", 10)
		backend.ID = ""

		err := manager.RegisterBackend(backend)
		if err == nil {
			t.Error("expected error for backend without ID")
		}
	})

	t.Run("RegisterBackendWithoutService", func(t *testing.T) {
		manager := createTestManager(t)
		backend := &EnclaveBackend{
			ID:   "test",
			Type: AttestationTypeSimulated,
		}

		err := manager.RegisterBackend(backend)
		if err == nil {
			t.Error("expected error for backend without service")
		}
	})

	t.Run("UnregisterBackend", func(t *testing.T) {
		manager := createTestManager(t)
		backend := createTestBackend("backend-1", 10)

		if err := manager.RegisterBackend(backend); err != nil {
			t.Fatalf("failed to register backend: %v", err)
		}

		if err := manager.UnregisterBackend("backend-1"); err != nil {
			t.Errorf("failed to unregister backend: %v", err)
		}

		backends := manager.ListBackends()
		if len(backends) != 0 {
			t.Errorf("expected 0 backends, got %d", len(backends))
		}
	})

	t.Run("UnregisterNonexistentBackend", func(t *testing.T) {
		manager := createTestManager(t)

		err := manager.UnregisterBackend("nonexistent")
		if !errors.Is(err, ErrBackendNotFound) {
			t.Errorf("expected ErrBackendNotFound, got %v", err)
		}
	})

	t.Run("GetBackend", func(t *testing.T) {
		manager := createTestManager(t)
		backend := createTestBackend("backend-1", 10)

		if err := manager.RegisterBackend(backend); err != nil {
			t.Fatalf("failed to register backend: %v", err)
		}

		retrieved, err := manager.GetBackend("backend-1")
		if err != nil {
			t.Errorf("failed to get backend: %v", err)
		}
		if retrieved.ID != "backend-1" {
			t.Errorf("expected ID 'backend-1', got '%s'", retrieved.ID)
		}
	})

	t.Run("GetNonexistentBackend", func(t *testing.T) {
		manager := createTestManager(t)

		_, err := manager.GetBackend("nonexistent")
		if !errors.Is(err, ErrBackendNotFound) {
			t.Errorf("expected ErrBackendNotFound, got %v", err)
		}
	})
}

// =============================================================================
// Selection Strategy Tests
// =============================================================================

func TestPrioritySelection(t *testing.T) {
	config := DefaultEnclaveManagerConfig()
	config.SelectionStrategy = StrategyPriority
	config.HealthCheckInterval = 1 * time.Second

	manager, err := NewEnclaveManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Add backends with different priorities
	backends := []*EnclaveBackend{
		createTestBackend("low-priority", 100),
		createTestBackend("high-priority", 10),
		createTestBackend("medium-priority", 50),
	}

	for _, b := range backends {
		if err := manager.RegisterBackend(b); err != nil {
			t.Fatalf("failed to register backend: %v", err)
		}
	}

	// Select should always return highest priority (lowest number)
	for i := 0; i < 10; i++ {
		selected, err := manager.SelectBackend()
		if err != nil {
			t.Fatalf("failed to select backend: %v", err)
		}
		if selected.ID != "high-priority" {
			t.Errorf("expected 'high-priority', got '%s'", selected.ID)
		}
	}
}

func TestRoundRobinSelection(t *testing.T) {
	config := DefaultEnclaveManagerConfig()
	config.SelectionStrategy = StrategyRoundRobin
	config.HealthCheckInterval = 1 * time.Second

	manager, err := NewEnclaveManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Add backends
	backends := []*EnclaveBackend{
		createTestBackend("backend-a", 10),
		createTestBackend("backend-b", 10),
		createTestBackend("backend-c", 10),
	}

	for _, b := range backends {
		if err := manager.RegisterBackend(b); err != nil {
			t.Fatalf("failed to register backend: %v", err)
		}
	}

	// Track which backends are selected
	selected := make(map[string]int)

	for i := 0; i < 9; i++ {
		b, err := manager.SelectBackend()
		if err != nil {
			t.Fatalf("failed to select backend: %v", err)
		}
		selected[b.ID]++
	}

	// Each backend should be selected 3 times
	for _, b := range backends {
		if selected[b.ID] != 3 {
			t.Errorf("expected backend '%s' to be selected 3 times, got %d", b.ID, selected[b.ID])
		}
	}
}

func TestLeastLoadedSelection(t *testing.T) {
	config := DefaultEnclaveManagerConfig()
	config.SelectionStrategy = StrategyLeastLoaded
	config.HealthCheckInterval = 1 * time.Second

	manager, err := NewEnclaveManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	b1 := createTestBackend("backend-1", 10)
	b2 := createTestBackend("backend-2", 10)
	b3 := createTestBackend("backend-3", 10)

	// Set different loads
	atomic.StoreInt32(&b1.activeRequests, 5)
	atomic.StoreInt32(&b2.activeRequests, 2)
	atomic.StoreInt32(&b3.activeRequests, 8)

	for _, b := range []*EnclaveBackend{b1, b2, b3} {
		if err := manager.RegisterBackend(b); err != nil {
			t.Fatalf("failed to register backend: %v", err)
		}
	}

	// Should select backend-2 (least loaded)
	selected, err := manager.SelectBackend()
	if err != nil {
		t.Fatalf("failed to select backend: %v", err)
	}
	if selected.ID != "backend-2" {
		t.Errorf("expected 'backend-2', got '%s'", selected.ID)
	}
}

func TestWeightedSelection(t *testing.T) {
	config := DefaultEnclaveManagerConfig()
	config.SelectionStrategy = StrategyWeighted
	config.HealthCheckInterval = 1 * time.Second

	manager, err := NewEnclaveManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create backends with different weights
	b1 := createTestBackend("backend-1", 10)
	b1.Weight = 1
	b2 := createTestBackend("backend-2", 10)
	b2.Weight = 9 // Should be selected ~90% of the time

	for _, b := range []*EnclaveBackend{b1, b2} {
		if err := manager.RegisterBackend(b); err != nil {
			t.Fatalf("failed to register backend: %v", err)
		}
	}

	// Track selections
	selected := make(map[string]int)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		b, err := manager.SelectBackend()
		if err != nil {
			t.Fatalf("failed to select backend: %v", err)
		}
		selected[b.ID]++
	}

	// backend-2 should be selected significantly more than backend-1
	if selected["backend-2"] < selected["backend-1"]*3 {
		t.Errorf("weighted selection not working as expected: backend-1=%d, backend-2=%d",
			selected["backend-1"], selected["backend-2"])
	}
}

func TestLatencySelection(t *testing.T) {
	config := DefaultEnclaveManagerConfig()
	config.SelectionStrategy = StrategyLatency
	config.HealthCheckInterval = 1 * time.Second

	manager, err := NewEnclaveManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	b1 := createTestBackend("backend-1", 10)
	b1.Metrics.TotalRequests = 100
	b1.Metrics.AverageLatencyMs = 50.0

	b2 := createTestBackend("backend-2", 10)
	b2.Metrics.TotalRequests = 100
	b2.Metrics.AverageLatencyMs = 20.0 // Lowest latency

	b3 := createTestBackend("backend-3", 10)
	b3.Metrics.TotalRequests = 100
	b3.Metrics.AverageLatencyMs = 100.0

	for _, b := range []*EnclaveBackend{b1, b2, b3} {
		if err := manager.RegisterBackend(b); err != nil {
			t.Fatalf("failed to register backend: %v", err)
		}
	}

	selected, err := manager.SelectBackend()
	if err != nil {
		t.Fatalf("failed to select backend: %v", err)
	}
	if selected.ID != "backend-2" {
		t.Errorf("expected 'backend-2' (lowest latency), got '%s'", selected.ID)
	}
}

// =============================================================================
// Health Monitoring Tests
// =============================================================================

func TestHealthMonitoring(t *testing.T) {
	config := DefaultEnclaveManagerConfig()
	config.HealthCheckInterval = 1 * time.Second
	config.UnhealthyThreshold = 2
	config.RecoveryThreshold = 1

	manager, err := NewEnclaveManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	svc := newMockEnclaveService()
	svc.initialized = true

	backend := NewEnclaveBackend("test-backend", AttestationTypeSimulated, svc)
	backend.Health = HealthHealthy

	if err := manager.RegisterBackend(backend); err != nil {
		t.Fatalf("failed to register backend: %v", err)
	}

	if err := manager.Start(); err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	defer manager.Stop()

	// Manual health check instead of waiting
	manager.performHealthChecks()

	status := manager.GetStatus()
	if status.HealthyBackends != 1 {
		t.Errorf("expected 1 healthy backend, got %d", status.HealthyBackends)
	}

	// Make backend unavailable
	svc.setAvailable(false)

	// Manually trigger health checks to detect unhealthy state
	for i := 0; i < config.UnhealthyThreshold; i++ {
		manager.performHealthChecks()
	}

	status = manager.GetStatus()
	if status.UnhealthyBackends != 1 {
		t.Errorf("expected 1 unhealthy backend, got %d (healthy=%d, degraded=%d)",
			status.UnhealthyBackends, status.HealthyBackends, status.DegradedBackends)
	}

	// Restore availability
	svc.setAvailable(true)

	// Manually trigger recovery
	for i := 0; i < config.RecoveryThreshold; i++ {
		manager.performHealthChecks()
	}

	status = manager.GetStatus()
	if status.HealthyBackends != 1 {
		t.Errorf("expected 1 healthy backend after recovery, got %d", status.HealthyBackends)
	}
}

// =============================================================================
// Failover Tests
// =============================================================================

func TestFailover(t *testing.T) {
	config := DefaultEnclaveManagerConfig()
	config.SelectionStrategy = StrategyPriority
	config.EnableFailover = true
	config.MaxRetries = 2
	config.RetryBackoff = 10 * time.Millisecond
	config.RequestTimeout = 1 * time.Second
	config.HealthCheckInterval = 1 * time.Second

	manager, err := NewEnclaveManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create failing primary backend
	svc1 := newMockEnclaveService()
	svc1.initialized = true
	svc1.shouldFail = true
	b1 := NewEnclaveBackend("primary", AttestationTypeSimulated, svc1)
	b1.Priority = 10
	b1.Health = HealthHealthy

	// Create working secondary backend
	svc2 := newMockEnclaveService()
	svc2.initialized = true
	b2 := NewEnclaveBackend("secondary", AttestationTypeSimulated, svc2)
	b2.Priority = 20
	b2.Health = HealthHealthy

	for _, b := range []*EnclaveBackend{b1, b2} {
		if err := manager.RegisterBackend(b); err != nil {
			t.Fatalf("failed to register backend: %v", err)
		}
	}

	if err := manager.Start(); err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	defer manager.Stop()

	// Request should fail on primary, then succeed on secondary
	request := createTestRequest("failover-test-1")
	result, err := manager.Score(context.Background(), request)

	if err != nil {
		t.Fatalf("expected success after failover, got error: %v", err)
	}
	if !result.IsSuccess() {
		t.Errorf("expected successful result, got error: %s", result.Error)
	}

	// Verify primary was called first
	if svc1.getCallCount() != 1 {
		t.Errorf("expected primary to be called once, got %d", svc1.getCallCount())
	}
	if svc2.getCallCount() != 1 {
		t.Errorf("expected secondary to be called once, got %d", svc2.getCallCount())
	}
}

func TestFailoverDisabled(t *testing.T) {
	config := DefaultEnclaveManagerConfig()
	config.EnableFailover = false
	config.HealthCheckInterval = 1 * time.Second

	manager, err := NewEnclaveManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	svc := newMockEnclaveService()
	svc.initialized = true
	svc.shouldFail = true

	backend := NewEnclaveBackend("failing", AttestationTypeSimulated, svc)
	backend.Health = HealthHealthy

	if err := manager.RegisterBackend(backend); err != nil {
		t.Fatalf("failed to register backend: %v", err)
	}

	if err := manager.Start(); err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	defer manager.Stop()

	request := createTestRequest("no-failover-test")
	_, err = manager.Score(context.Background(), request)

	if err == nil {
		t.Error("expected error when failover is disabled")
	}
}

// =============================================================================
// Circuit Breaker Tests
// =============================================================================

func TestCircuitBreaker(t *testing.T) {
	config := DefaultEnclaveManagerConfig()
	config.UnhealthyThreshold = 2
	config.CircuitBreakerReset = 100 * time.Millisecond
	config.HealthCheckInterval = 1 * time.Second
	config.EnableFailover = false
	config.RequestTimeout = 100 * time.Millisecond

	manager, err := NewEnclaveManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	svc := newMockEnclaveService()
	svc.initialized = true
	svc.shouldFail = true

	backend := NewEnclaveBackend("circuit-test", AttestationTypeSimulated, svc)
	backend.Health = HealthHealthy

	if err := manager.RegisterBackend(backend); err != nil {
		t.Fatalf("failed to register backend: %v", err)
	}

	if err := manager.Start(); err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	defer manager.Stop()

	// Cause failures to trip the circuit breaker
	for i := 0; i < 3; i++ {
		request := createTestRequest("circuit-test-" + string(rune('a'+i)))
		_, _ = manager.Score(context.Background(), request)
	}

	// Check that circuit is open
	backend.mu.RLock()
	circuitState := backend.circuitState
	backend.mu.RUnlock()

	if circuitState != CircuitOpen {
		t.Errorf("expected circuit to be OPEN, got %v", circuitState)
	}

	// Wait for circuit breaker reset time
	time.Sleep(150 * time.Millisecond)

	// Manually trigger circuit breaker reset check
	manager.resetCircuitBreakers()

	// Circuit should be half-open now
	backend.mu.RLock()
	circuitState = backend.circuitState
	backend.mu.RUnlock()

	if circuitState != CircuitHalfOpen {
		t.Errorf("expected circuit to be HALF_OPEN, got %v", circuitState)
	}

	// For half-open state to work, we need to also reset health to at least degraded
	// (in real scenario, the health check would detect service is available again)
	backend.mu.Lock()
	backend.Health = HealthDegraded
	backend.mu.Unlock()

	// Successful request should close the circuit
	svc.shouldFail = false
	request := createTestRequest("circuit-recovery")
	_, err = manager.Score(context.Background(), request)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	backend.mu.RLock()
	circuitState = backend.circuitState
	backend.mu.RUnlock()

	if circuitState != CircuitClosed {
		t.Errorf("expected circuit to be CLOSED after success, got %v", circuitState)
	}
}

// =============================================================================
// Concurrent Request Tests
// =============================================================================

func TestConcurrentRequests(t *testing.T) {
	config := DefaultEnclaveManagerConfig()
	config.HealthCheckInterval = 1 * time.Second

	manager, err := NewEnclaveManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create multiple backends
	for i := 0; i < 3; i++ {
		svc := newMockEnclaveService()
		svc.initialized = true
		svc.scoreLatency = 10 * time.Millisecond

		backend := NewEnclaveBackend(
			"backend-"+string(rune('a'+i)),
			AttestationTypeSimulated,
			svc,
		)
		backend.Priority = i * 10
		backend.Health = HealthHealthy
		backend.MaxConcurrent = 10

		if err := manager.RegisterBackend(backend); err != nil {
			t.Fatalf("failed to register backend: %v", err)
		}
	}

	if err := manager.Start(); err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	defer manager.Stop()

	// Run concurrent requests
	var wg sync.WaitGroup
	errCount := int32(0)
	successCount := int32(0)
	numRequests := 50

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			request := createTestRequest("concurrent-" + string(rune('a'+idx%26)) + "-" + time.Now().String())
			result, err := manager.Score(context.Background(), request)

			if err != nil {
				atomic.AddInt32(&errCount, 1)
				return
			}
			if !result.IsSuccess() {
				atomic.AddInt32(&errCount, 1)
				return
			}
			atomic.AddInt32(&successCount, 1)
		}(i)
	}

	wg.Wait()

	//nolint:gosec // G115: numRequests/2 is small positive test value
	if successCount < int32(numRequests/2) {
		t.Errorf("too many failures in concurrent test: success=%d, errors=%d",
			successCount, errCount)
	}
}

// =============================================================================
// Metrics Collection Tests
// =============================================================================

func TestMetricsCollection(t *testing.T) {
	config := DefaultEnclaveManagerConfig()
	config.HealthCheckInterval = 1 * time.Second

	manager, err := NewEnclaveManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	svc := newMockEnclaveService()
	svc.initialized = true
	svc.scoreLatency = 10 * time.Millisecond

	backend := NewEnclaveBackend("metrics-test", AttestationTypeSimulated, svc)
	backend.Health = HealthHealthy

	if err := manager.RegisterBackend(backend); err != nil {
		t.Fatalf("failed to register backend: %v", err)
	}

	if err := manager.Start(); err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	defer manager.Stop()

	// Run some successful requests
	for i := 0; i < 5; i++ {
		request := createTestRequest("metrics-success-" + string(rune('a'+i)))
		_, err := manager.Score(context.Background(), request)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// Run some failing requests
	svc.shouldFail = true
	for i := 0; i < 3; i++ {
		request := createTestRequest("metrics-fail-" + string(rune('a'+i)))
		_, _ = manager.Score(context.Background(), request)
	}

	status := manager.GetStatus()
	if len(status.Backends) != 1 {
		t.Fatalf("expected 1 backend status, got %d", len(status.Backends))
	}

	bs := status.Backends[0]
	if bs.TotalRequests != 8 {
		t.Errorf("expected 8 total requests, got %d", bs.TotalRequests)
	}

	// Success rate should be ~62.5% (5/8)
	expectedRate := float64(5) / float64(8) * 100
	if bs.SuccessRate < expectedRate-1 || bs.SuccessRate > expectedRate+1 {
		t.Errorf("expected success rate ~%.1f%%, got %.1f%%", expectedRate, bs.SuccessRate)
	}

	if bs.AverageLatencyMs <= 0 {
		t.Error("expected positive average latency")
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestNoBackendsAvailable(t *testing.T) {
	manager := createTestManager(t)

	_, err := manager.SelectBackend()
	if !errors.Is(err, ErrNoBackendsAvailable) {
		t.Errorf("expected ErrNoBackendsAvailable, got %v", err)
	}
}

func TestManagerNotRunning(t *testing.T) {
	manager := createTestManager(t)

	backend := createTestBackend("test", 10)
	if err := manager.RegisterBackend(backend); err != nil {
		t.Fatalf("failed to register backend: %v", err)
	}

	request := createTestRequest("not-running-test")
	_, err := manager.Score(context.Background(), request)
	if !errors.Is(err, ErrManagerNotRunning) {
		t.Errorf("expected ErrManagerNotRunning, got %v", err)
	}
}

func TestDuplicateRequest(t *testing.T) {
	config := DefaultEnclaveManagerConfig()
	config.HealthCheckInterval = 1 * time.Second

	manager, err := NewEnclaveManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	svc := newMockEnclaveService()
	svc.initialized = true
	svc.scoreLatency = 100 * time.Millisecond // Slow to ensure overlap

	backend := NewEnclaveBackend("dup-test", AttestationTypeSimulated, svc)
	backend.Health = HealthHealthy

	if err := manager.RegisterBackend(backend); err != nil {
		t.Fatalf("failed to register backend: %v", err)
	}

	if err := manager.Start(); err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	defer manager.Stop()

	// Start first request
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		request := &ScoringRequest{
			RequestID:      "dup-request",
			Ciphertext:     []byte("test"),
			WrappedKey:     []byte("key"),
			Nonce:          []byte("nonce"),
			ScopeID:        "scope",
			AccountAddress: "addr",
		}
		_, _ = manager.Score(context.Background(), request)
	}()

	// Give it time to acquire the lock
	time.Sleep(10 * time.Millisecond)

	// Try duplicate request
	request := &ScoringRequest{
		RequestID:      "dup-request",
		Ciphertext:     []byte("test"),
		WrappedKey:     []byte("key"),
		Nonce:          []byte("nonce"),
		ScopeID:        "scope",
		AccountAddress: "addr",
	}
	_, err = manager.Score(context.Background(), request)
	if !errors.Is(err, ErrDuplicateRequest) {
		t.Errorf("expected ErrDuplicateRequest, got %v", err)
	}

	wg.Wait()
}

func TestGenerateAttestation(t *testing.T) {
	manager := createTestManager(t)

	backend := createTestBackend("attestation-test", 10)
	if err := manager.RegisterBackend(backend); err != nil {
		t.Fatalf("failed to register backend: %v", err)
	}

	if err := manager.Start(); err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	defer manager.Stop()

	reportData := []byte("test-report-data")
	attestation, err := manager.GenerateAttestation(reportData)
	if err != nil {
		t.Fatalf("failed to generate attestation: %v", err)
	}

	if len(attestation) == 0 {
		t.Error("expected non-empty attestation")
	}
}

func TestManagerStartStop(t *testing.T) {
	manager := createTestManager(t)

	// Start
	if err := manager.Start(); err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}

	status := manager.GetStatus()
	if !status.Running {
		t.Error("expected manager to be running")
	}

	// Double start should fail
	if err := manager.Start(); err == nil {
		t.Error("expected error on double start")
	}

	// Stop
	if err := manager.Stop(); err != nil {
		t.Fatalf("failed to stop manager: %v", err)
	}

	status = manager.GetStatus()
	if status.Running {
		t.Error("expected manager to be stopped")
	}

	// Double stop should be idempotent
	if err := manager.Stop(); err != nil {
		t.Errorf("double stop should not error: %v", err)
	}
}

func TestBackendAvailability(t *testing.T) {
	t.Run("HealthyAvailable", func(t *testing.T) {
		backend := createTestBackend("test", 10)
		backend.Health = HealthHealthy

		if !backend.IsAvailable() {
			t.Error("healthy backend should be available")
		}
	})

	t.Run("UnhealthyUnavailable", func(t *testing.T) {
		backend := createTestBackend("test", 10)
		backend.Health = HealthUnhealthy

		if backend.IsAvailable() {
			t.Error("unhealthy backend should not be available")
		}
	})

	t.Run("CircuitOpenUnavailable", func(t *testing.T) {
		backend := createTestBackend("test", 10)
		backend.Health = HealthHealthy
		backend.circuitState = CircuitOpen

		if backend.IsAvailable() {
			t.Error("backend with open circuit should not be available")
		}
	})

	t.Run("MaxConcurrentReached", func(t *testing.T) {
		backend := createTestBackend("test", 10)
		backend.Health = HealthHealthy
		backend.MaxConcurrent = 5
		atomic.StoreInt32(&backend.activeRequests, 5)

		if backend.IsAvailable() {
			t.Error("backend at max concurrent should not be available")
		}
	})
}

func TestCreateSimulatedBackend(t *testing.T) {
	backend, err := CreateSimulatedBackend("sim-test")
	if err != nil {
		t.Fatalf("failed to create simulated backend: %v", err)
	}

	if backend.ID != "sim-test" {
		t.Errorf("expected ID 'sim-test', got '%s'", backend.ID)
	}
	if backend.Type != AttestationTypeSimulated {
		t.Errorf("expected AttestationTypeSimulated, got %v", backend.Type)
	}
	if backend.Health != HealthHealthy {
		t.Errorf("expected HealthHealthy, got %v", backend.Health)
	}
}

func TestCreateDefaultManager(t *testing.T) {
	manager, err := CreateDefaultManager()
	if err != nil {
		t.Fatalf("failed to create default manager: %v", err)
	}

	backends := manager.ListBackends()
	if len(backends) != 1 {
		t.Errorf("expected 1 backend, got %d", len(backends))
	}

	if backends[0].ID != "simulated-default" {
		t.Errorf("expected 'simulated-default', got '%s'", backends[0].ID)
	}
}

func TestHealthStatusString(t *testing.T) {
	tests := []struct {
		status   HealthStatus
		expected string
	}{
		{HealthUnknown, "UNKNOWN"},
		{HealthHealthy, "HEALTHY"},
		{HealthDegraded, "DEGRADED"},
		{HealthUnhealthy, "UNHEALTHY"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, got)
		}
	}
}

func TestSelectionStrategyString(t *testing.T) {
	tests := []struct {
		strategy SelectionStrategy
		expected string
	}{
		{StrategyPriority, "PRIORITY"},
		{StrategyRoundRobin, "ROUND_ROBIN"},
		{StrategyLeastLoaded, "LEAST_LOADED"},
		{StrategyWeighted, "WEIGHTED"},
		{StrategyLatency, "LATENCY"},
	}

	for _, tt := range tests {
		if got := tt.strategy.String(); got != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, got)
		}
	}
}

func TestCircuitBreakerStateString(t *testing.T) {
	tests := []struct {
		state    CircuitBreakerState
		expected string
	}{
		{CircuitClosed, "CLOSED"},
		{CircuitOpen, "OPEN"},
		{CircuitHalfOpen, "HALF_OPEN"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, got)
		}
	}
}

func TestBackendMetricsClone(t *testing.T) {
	metrics := BackendMetrics{
		TotalRequests:      100,
		SuccessfulRequests: 90,
		FailedRequests:     10,
		TotalLatencyMs:     5000,
		AverageLatencyMs:   50.0,
		LastRequestTime:    time.Now(),
	}

	clone := metrics.Clone()

	if clone.TotalRequests != metrics.TotalRequests {
		t.Error("clone TotalRequests mismatch")
	}
	if clone.SuccessfulRequests != metrics.SuccessfulRequests {
		t.Error("clone SuccessfulRequests mismatch")
	}
	if clone.AverageLatencyMs != metrics.AverageLatencyMs {
		t.Error("clone AverageLatencyMs mismatch")
	}
}

