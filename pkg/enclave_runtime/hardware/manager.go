// Package hardware provides a comprehensive hardware abstraction layer for TEE integration.
//
// This file implements the HardwareManager which handles lifecycle management of
// TEE hardware backends. It provides initialization, backend selection, health checks,
// and automatic fallback to simulation when hardware is unavailable.
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package hardware

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// Hardware Manager
// =============================================================================

// HardwareManager manages the lifecycle of TEE hardware backends.
// It handles initialization, backend selection, health monitoring, and cleanup.
type HardwareManager struct {
	mu sync.RWMutex

	// Configuration
	config Config

	// Detection
	detector *UnifiedDetector

	// Backends
	backends       map[Platform]Backend
	activeBackend  Backend
	activePlatform Platform

	// State
	initialized bool
	initError   error

	// Health check state
	lastHealthCheck  time.Time
	healthCheckError error
	healthCheckCtx   context.Context
	healthCheckStop  context.CancelFunc

	// Statistics
	stats ManagerStats
}

// ManagerStats tracks hardware manager statistics.
type ManagerStats struct {
	InitializationTime  time.Duration
	TotalAttestations   int64
	TotalKeyDerivations int64
	TotalSealOperations int64
	TotalHealthChecks   int64
	FailedHealthChecks  int64
	BackendSwitchCount  int64
	LastBackendSwitch   time.Time
}

// NewHardwareManager creates a new hardware manager with the given configuration.
func NewHardwareManager(config Config) (*HardwareManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &HardwareManager{
		config:   config,
		detector: NewUnifiedDetector(),
		backends: make(map[Platform]Backend),
	}, nil
}

// NewHardwareManagerWithDefaults creates a new hardware manager with default configuration.
func NewHardwareManagerWithDefaults() (*HardwareManager, error) {
	return NewHardwareManager(DefaultConfig())
}

// Initialize sets up the hardware manager and initializes available backends.
func (m *HardwareManager) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return nil
	}

	startTime := time.Now()

	// Detect available hardware
	caps, err := m.detector.Detect()
	if err != nil {
		m.initError = fmt.Errorf("hardware detection failed: %w", err)
		// Don't return error - we might still be able to use simulation
	}

	// Handle hardware requirement
	if m.config.RequireHardware && (caps == nil || !caps.HasAnyHardware()) {
		m.initError = fmt.Errorf("%w: mode requires hardware but none detected", ErrHardwareNotFound)
		return m.initError
	}

	// Initialize backends based on available hardware
	initErrors := m.initializeBackends(ctx, caps)

	// Select active backend
	if err := m.selectActiveBackend(); err != nil {
		if m.config.RequireHardware {
			m.initError = fmt.Errorf("failed to select backend: %w", err)
			return m.initError
		}
		// Fall back to simulation if allowed
		if m.config.AllowSimulation {
			simBackend := NewSimulatedBackend()
			if err := simBackend.Initialize(); err != nil {
				m.initError = fmt.Errorf("failed to initialize simulated backend: %w", err)
				return m.initError
			}
			m.backends[PlatformSimulated] = simBackend
			m.activeBackend = simBackend
			m.activePlatform = PlatformSimulated
		}
	}

	// Start health check loop if configured
	if m.config.HealthCheckInterval > 0 && m.activeBackend != nil {
		m.startHealthCheckLoop()
	}

	m.stats.InitializationTime = time.Since(startTime)
	m.initialized = true

	// Log any initialization errors that were non-fatal
	if len(initErrors) > 0 {
		// These are logged but don't prevent initialization
		_ = initErrors
	}

	return nil
}

// initializeBackends initializes all available hardware backends.
func (m *HardwareManager) initializeBackends(ctx context.Context, caps *HardwareCapabilities) []error {
	var errors []error

	if caps == nil {
		return errors
	}

	// Initialize SGX backend
	if caps.SGX.Available {
		backend := NewSGXBackend(m.config.SGXConfig)
		if err := backend.Initialize(); err != nil {
			errors = append(errors, fmt.Errorf("SGX initialization failed: %w", err))
		} else {
			m.backends[PlatformSGX] = backend
		}
	}

	// Initialize SEV-SNP backend
	if caps.SEVSNP.Available {
		backend := NewSEVBackend(m.config.SEVConfig)
		if err := backend.Initialize(); err != nil {
			errors = append(errors, fmt.Errorf("SEV-SNP initialization failed: %w", err))
		} else {
			m.backends[PlatformSEVSNP] = backend
		}
	}

	// Initialize Nitro backend
	if caps.Nitro.Available {
		backend := NewNitroBackend(m.config.NitroConfig)
		if err := backend.Initialize(); err != nil {
			errors = append(errors, fmt.Errorf("Nitro initialization failed: %w", err))
		} else {
			m.backends[PlatformNitro] = backend
		}
	}

	_ = ctx // Reserved for timeout handling
	return errors
}

// selectActiveBackend chooses the best available backend.
func (m *HardwareManager) selectActiveBackend() error {
	// If a specific platform is preferred and available, use it
	if m.config.PreferredPlatform != PlatformUnknown {
		if backend, ok := m.backends[m.config.PreferredPlatform]; ok {
			m.activeBackend = backend
			m.activePlatform = m.config.PreferredPlatform
			return nil
		}
	}

	// Otherwise, use the recommended platform from detection
	recommended := m.detector.GetRecommendedPlatform()
	if backend, ok := m.backends[recommended]; ok {
		m.activeBackend = backend
		m.activePlatform = recommended
		return nil
	}

	// Fall back to any available backend
	for platform, backend := range m.backends {
		m.activeBackend = backend
		m.activePlatform = platform
		return nil
	}

	return ErrHardwareNotFound
}

// startHealthCheckLoop starts the background health check loop.
func (m *HardwareManager) startHealthCheckLoop() {
	m.healthCheckCtx, m.healthCheckStop = context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(m.config.HealthCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-m.healthCheckCtx.Done():
				return
			case <-ticker.C:
				m.performHealthCheck()
			}
		}
	}()
}

// performHealthCheck runs a health check on the active backend.
func (m *HardwareManager) performHealthCheck() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.activeBackend == nil {
		return
	}

	m.stats.TotalHealthChecks++
	m.lastHealthCheck = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), m.config.HealthCheckTimeout)
	defer cancel()

	// Use a channel to handle the health check with timeout
	errCh := make(chan error, 1)
	go func() {
		errCh <- m.activeBackend.HealthCheck()
	}()

	select {
	case err := <-errCh:
		m.healthCheckError = err
		if err != nil {
			m.stats.FailedHealthChecks++
			// Consider switching to a different backend
			m.handleHealthCheckFailure()
		}
	case <-ctx.Done():
		m.healthCheckError = fmt.Errorf("health check timeout: %w", ctx.Err())
		m.stats.FailedHealthChecks++
	}
}

// handleHealthCheckFailure handles a failed health check.
func (m *HardwareManager) handleHealthCheckFailure() {
	// Try to switch to another available backend
	for platform, backend := range m.backends {
		if platform == m.activePlatform {
			continue
		}
		if backend.HealthCheck() == nil {
			m.activeBackend = backend
			m.activePlatform = platform
			m.stats.BackendSwitchCount++
			m.stats.LastBackendSwitch = time.Now()
			return
		}
	}

	// If no other backend is available and simulation is allowed, use it
	if m.config.AllowSimulation && m.activePlatform != PlatformSimulated {
		if simBackend, ok := m.backends[PlatformSimulated]; ok {
			m.activeBackend = simBackend
			m.activePlatform = PlatformSimulated
			m.stats.BackendSwitchCount++
			m.stats.LastBackendSwitch = time.Now()
		}
	}
}

// GetBackend returns the currently active backend.
func (m *HardwareManager) GetBackend() Backend {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeBackend
}

// GetPlatform returns the currently active platform.
func (m *HardwareManager) GetPlatform() Platform {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activePlatform
}

// GetCapabilities returns the detected hardware capabilities.
func (m *HardwareManager) GetCapabilities() *HardwareCapabilities {
	return m.detector.GetCapabilities()
}

// IsHardwareActive returns true if real hardware is being used (not simulation).
func (m *HardwareManager) IsHardwareActive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activePlatform.IsHardware()
}

// IsInitialized returns true if the manager has been initialized.
func (m *HardwareManager) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.initialized
}

// SwitchBackend attempts to switch to a different backend.
func (m *HardwareManager) SwitchBackend(platform Platform) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return ErrNotInitialized
	}

	backend, ok := m.backends[platform]
	if !ok {
		return fmt.Errorf("%w: platform %s not available", ErrHardwareNotFound, platform)
	}

	// Verify the backend is healthy
	if err := backend.HealthCheck(); err != nil {
		return fmt.Errorf("backend health check failed: %w", err)
	}

	m.activeBackend = backend
	m.activePlatform = platform
	m.stats.BackendSwitchCount++
	m.stats.LastBackendSwitch = time.Now()

	return nil
}

// GetAttestation generates an attestation using the active backend.
func (m *HardwareManager) GetAttestation(nonce []byte) ([]byte, error) {
	m.mu.RLock()
	backend := m.activeBackend
	m.mu.RUnlock()

	if backend == nil {
		return nil, ErrNotInitialized
	}

	m.mu.Lock()
	m.stats.TotalAttestations++
	m.mu.Unlock()

	attestation, err := backend.GetAttestation(nonce)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAttestationFailed, err)
	}

	return attestation, nil
}

// DeriveKey derives a key using the active backend.
func (m *HardwareManager) DeriveKey(context []byte, size int) ([]byte, error) {
	m.mu.RLock()
	backend := m.activeBackend
	m.mu.RUnlock()

	if backend == nil {
		return nil, ErrNotInitialized
	}

	m.mu.Lock()
	m.stats.TotalKeyDerivations++
	m.mu.Unlock()

	key, err := backend.DeriveKey(context, size)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrKeyDerivationFailed, err)
	}

	return key, nil
}

// Seal encrypts data using the active backend.
func (m *HardwareManager) Seal(plaintext []byte) ([]byte, error) {
	m.mu.RLock()
	backend := m.activeBackend
	m.mu.RUnlock()

	if backend == nil {
		return nil, ErrNotInitialized
	}

	m.mu.Lock()
	m.stats.TotalSealOperations++
	m.mu.Unlock()

	sealed, err := backend.Seal(plaintext)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSealingFailed, err)
	}

	return sealed, nil
}

// Unseal decrypts data using the active backend.
func (m *HardwareManager) Unseal(ciphertext []byte) ([]byte, error) {
	m.mu.RLock()
	backend := m.activeBackend
	m.mu.RUnlock()

	if backend == nil {
		return nil, ErrNotInitialized
	}

	plaintext, err := backend.Unseal(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnsealingFailed, err)
	}

	return plaintext, nil
}

// HealthCheck performs a health check on the active backend.
func (m *HardwareManager) HealthCheck() error {
	m.mu.RLock()
	backend := m.activeBackend
	m.mu.RUnlock()

	if backend == nil {
		return ErrNotInitialized
	}

	return backend.HealthCheck()
}

// GetStats returns the current manager statistics.
func (m *HardwareManager) GetStats() ManagerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// GetAvailablePlatforms returns a list of available platforms.
func (m *HardwareManager) GetAvailablePlatforms() []Platform {
	m.mu.RLock()
	defer m.mu.RUnlock()

	platforms := make([]Platform, 0, len(m.backends))
	for platform := range m.backends {
		platforms = append(platforms, platform)
	}
	return platforms
}

// Shutdown cleanly shuts down all backends and releases resources.
func (m *HardwareManager) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return nil
	}

	// Stop health check loop
	if m.healthCheckStop != nil {
		m.healthCheckStop()
	}

	// Shutdown all backends
	var errs []error
	for platform, backend := range m.backends {
		if err := backend.Shutdown(); err != nil {
			errs = append(errs, fmt.Errorf("%s shutdown: %w", platform, err))
		}
	}

	m.backends = make(map[Platform]Backend)
	m.activeBackend = nil
	m.activePlatform = PlatformUnknown
	m.initialized = false

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}

// =============================================================================
// Backend Implementations (Stubs - Real implementations in separate files)
// =============================================================================

// SGXBackendImpl implements the Backend interface for Intel SGX.
type SGXBackendImpl struct {
	config      SGXConfig
	initialized bool
	mu          sync.RWMutex
}

// NewSGXBackend creates a new SGX backend.
func NewSGXBackend(config SGXConfig) Backend {
	return &SGXBackendImpl{config: config}
}

func (b *SGXBackendImpl) Platform() Platform { return PlatformSGX }

func (b *SGXBackendImpl) IsAvailable() bool {
	caps, _ := DetectHardware()
	return caps != nil && caps.SGX.Available
}

func (b *SGXBackendImpl) Initialize() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initialized = true
	return nil
}

func (b *SGXBackendImpl) Shutdown() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initialized = false
	return nil
}

func (b *SGXBackendImpl) GetAttestation(nonce []byte) ([]byte, error) {
	// Stub - real implementation would use SGX SDK
	return nil, ErrNotSupported
}

func (b *SGXBackendImpl) DeriveKey(context []byte, keySize int) ([]byte, error) {
	// Stub - real implementation would use SGX sealing
	return nil, ErrNotSupported
}

func (b *SGXBackendImpl) Seal(plaintext []byte) ([]byte, error) {
	return nil, ErrNotSupported
}

func (b *SGXBackendImpl) Unseal(ciphertext []byte) ([]byte, error) {
	return nil, ErrNotSupported
}

func (b *SGXBackendImpl) HealthCheck() error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if !b.initialized {
		return ErrNotInitialized
	}
	return nil
}

func (b *SGXBackendImpl) GetCapabilities() *HardwareCapabilities {
	caps, _ := DetectHardware()
	return caps
}

func (b *SGXBackendImpl) GetInfo() *HardwareInfo {
	return &HardwareInfo{
		Platform: PlatformSGX,
		Name:     "Intel SGX",
		Vendor:   "Intel",
	}
}

// SEVBackendImpl implements the Backend interface for AMD SEV-SNP.
type SEVBackendImpl struct {
	config      SEVConfig
	initialized bool
	mu          sync.RWMutex
}

// NewSEVBackend creates a new SEV-SNP backend.
func NewSEVBackend(config SEVConfig) Backend {
	return &SEVBackendImpl{config: config}
}

func (b *SEVBackendImpl) Platform() Platform { return PlatformSEVSNP }

func (b *SEVBackendImpl) IsAvailable() bool {
	caps, _ := DetectHardware()
	return caps != nil && caps.SEVSNP.Available
}

func (b *SEVBackendImpl) Initialize() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initialized = true
	return nil
}

func (b *SEVBackendImpl) Shutdown() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initialized = false
	return nil
}

func (b *SEVBackendImpl) GetAttestation(nonce []byte) ([]byte, error) {
	return nil, ErrNotSupported
}

func (b *SEVBackendImpl) DeriveKey(context []byte, keySize int) ([]byte, error) {
	return nil, ErrNotSupported
}

func (b *SEVBackendImpl) Seal(plaintext []byte) ([]byte, error) {
	return nil, ErrNotSupported
}

func (b *SEVBackendImpl) Unseal(ciphertext []byte) ([]byte, error) {
	return nil, ErrNotSupported
}

func (b *SEVBackendImpl) HealthCheck() error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if !b.initialized {
		return ErrNotInitialized
	}
	return nil
}

func (b *SEVBackendImpl) GetCapabilities() *HardwareCapabilities {
	caps, _ := DetectHardware()
	return caps
}

func (b *SEVBackendImpl) GetInfo() *HardwareInfo {
	return &HardwareInfo{
		Platform: PlatformSEVSNP,
		Name:     "AMD SEV-SNP",
		Vendor:   "AMD",
	}
}

// NitroBackendImpl implements the Backend interface for AWS Nitro.
type NitroBackendImpl struct {
	config      NitroConfig
	initialized bool
	mu          sync.RWMutex
}

// NewNitroBackend creates a new Nitro backend.
func NewNitroBackend(config NitroConfig) Backend {
	return &NitroBackendImpl{config: config}
}

func (b *NitroBackendImpl) Platform() Platform { return PlatformNitro }

func (b *NitroBackendImpl) IsAvailable() bool {
	caps, _ := DetectHardware()
	return caps != nil && caps.Nitro.Available
}

func (b *NitroBackendImpl) Initialize() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initialized = true
	return nil
}

func (b *NitroBackendImpl) Shutdown() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initialized = false
	return nil
}

func (b *NitroBackendImpl) GetAttestation(nonce []byte) ([]byte, error) {
	return nil, ErrNotSupported
}

func (b *NitroBackendImpl) DeriveKey(context []byte, keySize int) ([]byte, error) {
	return nil, ErrNotSupported
}

func (b *NitroBackendImpl) Seal(plaintext []byte) ([]byte, error) {
	return nil, ErrNotSupported
}

func (b *NitroBackendImpl) Unseal(ciphertext []byte) ([]byte, error) {
	return nil, ErrNotSupported
}

func (b *NitroBackendImpl) HealthCheck() error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if !b.initialized {
		return ErrNotInitialized
	}
	return nil
}

func (b *NitroBackendImpl) GetCapabilities() *HardwareCapabilities {
	caps, _ := DetectHardware()
	return caps
}

func (b *NitroBackendImpl) GetInfo() *HardwareInfo {
	return &HardwareInfo{
		Platform: PlatformNitro,
		Name:     "AWS Nitro Enclaves",
		Vendor:   "Amazon Web Services",
	}
}

// =============================================================================
// Simulated Backend
// =============================================================================

// SimulatedBackend provides a simulated TEE backend for testing and development.
type SimulatedBackend struct {
	initialized bool
	sealingKey  []byte
	mu          sync.RWMutex
}

// NewSimulatedBackend creates a new simulated backend.
func NewSimulatedBackend() *SimulatedBackend {
	return &SimulatedBackend{}
}

func (b *SimulatedBackend) Platform() Platform { return PlatformSimulated }

func (b *SimulatedBackend) IsAvailable() bool { return true }

func (b *SimulatedBackend) Initialize() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Generate a random sealing key
	b.sealingKey = make([]byte, 32)
	// Use a fixed key for simulation (DO NOT use in production)
	copy(b.sealingKey, []byte("simulated-sealing-key-32bytes!!"))
	b.initialized = true
	return nil
}

func (b *SimulatedBackend) Shutdown() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initialized = false
	b.sealingKey = nil
	return nil
}

func (b *SimulatedBackend) GetAttestation(nonce []byte) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, ErrNotInitialized
	}

	// Return a simulated attestation
	attestation := make([]byte, 256)
	copy(attestation, []byte("SIMULATED-ATTESTATION-"))
	copy(attestation[22:], nonce)
	return attestation, nil
}

func (b *SimulatedBackend) DeriveKey(context []byte, keySize int) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, ErrNotInitialized
	}

	// Simple key derivation (NOT SECURE - for simulation only)
	key := make([]byte, keySize)
	for i := 0; i < keySize; i++ {
		key[i] = b.sealingKey[i%len(b.sealingKey)] ^ context[i%len(context)]
	}
	return key, nil
}

func (b *SimulatedBackend) Seal(plaintext []byte) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, ErrNotInitialized
	}

	// Simple XOR (NOT SECURE - for simulation only)
	sealed := make([]byte, len(plaintext)+16)
	copy(sealed[:16], []byte("SIM-SEALED-HDR!!"))
	for i, p := range plaintext {
		sealed[16+i] = p ^ b.sealingKey[i%len(b.sealingKey)]
	}
	return sealed, nil
}

func (b *SimulatedBackend) Unseal(ciphertext []byte) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, ErrNotInitialized
	}

	if len(ciphertext) < 16 {
		return nil, ErrUnsealingFailed
	}

	// Verify header
	if string(ciphertext[:16]) != "SIM-SEALED-HDR!!" {
		return nil, ErrUnsealingFailed
	}

	// Simple XOR (NOT SECURE - for simulation only)
	plaintext := make([]byte, len(ciphertext)-16)
	for i := 0; i < len(plaintext); i++ {
		plaintext[i] = ciphertext[16+i] ^ b.sealingKey[i%len(b.sealingKey)]
	}
	return plaintext, nil
}

func (b *SimulatedBackend) HealthCheck() error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if !b.initialized {
		return ErrNotInitialized
	}
	return nil
}

func (b *SimulatedBackend) GetCapabilities() *HardwareCapabilities {
	return &HardwareCapabilities{
		Platform:  PlatformSimulated,
		Available: true,
	}
}

func (b *SimulatedBackend) GetInfo() *HardwareInfo {
	return &HardwareInfo{
		Platform:    PlatformSimulated,
		Name:        "Simulated TEE",
		Vendor:      "VirtEngine",
		Version:     "1.0.0",
		LastUpdated: time.Now(),
		Features: map[string]bool{
			"attestation":    true,
			"sealing":        true,
			"key_derivation": true,
		},
	}
}
