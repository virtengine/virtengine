// Package hardware provides a comprehensive hardware abstraction layer for TEE integration.
//
// This file implements a mock backend for testing. The MockBackend implements the
// Backend interface and allows test code to configure its behavior, record method
// calls, and simulate various failure conditions.
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package hardware

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// Mock Backend
// =============================================================================

// MockBackend implements the Backend interface for testing purposes.
// It provides configurable behavior and records method calls for assertions.
type MockBackend struct {
	mu sync.RWMutex

	// Configuration
	config MockConfig

	// State
	initialized bool
	sealingKey  []byte

	// Call recording
	calls []MethodCall

	// Simulated hardware info
	hardwareInfo *HardwareInfo
	capabilities *HardwareCapabilities
	platform     Platform
}

// MockConfig configures the behavior of the mock backend.
type MockConfig struct {
	// Platform to simulate
	Platform Platform

	// Failure simulation
	FailInitialize    bool
	FailShutdown      bool
	FailAttestation   bool
	FailKeyDerivation bool
	FailSeal          bool
	FailUnseal        bool
	FailHealthCheck   bool

	// Error to return on failure
	InitializeError    error
	ShutdownError      error
	AttestationError   error
	KeyDerivationError error
	SealError          error
	UnsealError        error
	HealthCheckError   error

	// Delays to simulate
	InitializeDelay    time.Duration
	AttestationDelay   time.Duration
	KeyDerivationDelay time.Duration
	SealDelay          time.Duration
	UnsealDelay        time.Duration
	HealthCheckDelay   time.Duration

	// Custom attestation generator
	AttestationGenerator func(nonce []byte) ([]byte, error)

	// Custom key derivation function
	KeyDerivationFunc func(context []byte, size int) ([]byte, error)

	// Pre-configured sealing key (if nil, generates random)
	SealingKey []byte
}

// DefaultMockConfig returns a MockConfig with sensible defaults.
func DefaultMockConfig() MockConfig {
	return MockConfig{
		Platform: PlatformSimulated,
	}
}

// MethodCall records a call to a mock method.
type MethodCall struct {
	Method    string
	Args      []interface{}
	Timestamp time.Time
	Duration  time.Duration
	Error     error
}

// NewMockBackend creates a new mock backend with the given configuration.
func NewMockBackend(config MockConfig) *MockBackend {
	if config.Platform == PlatformUnknown {
		config.Platform = PlatformSimulated
	}

	return &MockBackend{
		config:   config,
		platform: config.Platform,
		calls:    make([]MethodCall, 0),
	}
}

// NewMockBackendWithDefaults creates a new mock backend with default configuration.
func NewMockBackendWithDefaults() *MockBackend {
	return NewMockBackend(DefaultMockConfig())
}

// =============================================================================
// Backend Interface Implementation
// =============================================================================

// Platform returns the TEE platform type for this backend.
func (b *MockBackend) Platform() Platform {
	return b.platform
}

// IsAvailable returns true if this hardware is currently available.
func (b *MockBackend) IsAvailable() bool {
	b.recordCall("IsAvailable", nil)
	return true
}

// Initialize sets up the mock backend.
func (b *MockBackend) Initialize() error {
	startTime := time.Now()

	if b.config.InitializeDelay > 0 {
		time.Sleep(b.config.InitializeDelay)
	}

	var err error
	if b.config.FailInitialize {
		err = b.config.InitializeError
		if err == nil {
			err = ErrInitializationFailed
		}
	}

	if err == nil {
		b.mu.Lock()
		b.initialized = true

		// Set up sealing key
		if b.config.SealingKey != nil {
			b.sealingKey = make([]byte, len(b.config.SealingKey))
			copy(b.sealingKey, b.config.SealingKey)
		} else {
			b.sealingKey = make([]byte, 32)
			_, _ = rand.Read(b.sealingKey)
		}
		b.mu.Unlock()
	}

	b.recordCallWithDuration("Initialize", nil, time.Since(startTime), err)
	return err
}

// Shutdown cleanly shuts down the mock backend.
func (b *MockBackend) Shutdown() error {
	startTime := time.Now()

	var err error
	if b.config.FailShutdown {
		err = b.config.ShutdownError
		if err == nil {
			err = fmt.Errorf("mock shutdown failure")
		}
	}

	if err == nil {
		b.mu.Lock()
		b.initialized = false
		b.sealingKey = nil
		b.mu.Unlock()
	}

	b.recordCallWithDuration("Shutdown", nil, time.Since(startTime), err)
	return err
}

// GetAttestation generates a mock attestation.
func (b *MockBackend) GetAttestation(nonce []byte) ([]byte, error) {
	startTime := time.Now()

	if b.config.AttestationDelay > 0 {
		time.Sleep(b.config.AttestationDelay)
	}

	b.mu.RLock()
	initialized := b.initialized
	b.mu.RUnlock()

	if !initialized {
		err := ErrNotInitialized
		b.recordCallWithDuration("GetAttestation", []interface{}{nonce}, time.Since(startTime), err)
		return nil, err
	}

	var attestation []byte
	var err error

	if b.config.FailAttestation {
		err = b.config.AttestationError
		if err == nil {
			err = ErrAttestationFailed
		}
	} else if b.config.AttestationGenerator != nil {
		attestation, err = b.config.AttestationGenerator(nonce)
	} else {
		attestation = b.generateMockAttestation(nonce)
	}

	b.recordCallWithDuration("GetAttestation", []interface{}{nonce}, time.Since(startTime), err)
	return attestation, err
}

// generateMockAttestation generates a realistic mock attestation.
func (b *MockBackend) generateMockAttestation(nonce []byte) []byte {
	// Create a mock attestation structure
	attestation := make([]byte, 0, 512)

	// Header: magic bytes + version + platform
	attestation = append(attestation, []byte("MOCK")...)
	attestation = append(attestation, 0x01) // Version 1
	attestation = append(attestation, byte(b.platform))

	// Timestamp (8 bytes)
	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, uint64(time.Now().UnixNano()))
	attestation = append(attestation, ts...)

	// Nonce hash (32 bytes)
	nonceHash := sha256.Sum256(nonce)
	attestation = append(attestation, nonceHash[:]...)

	// Measurement (simulated - 48 bytes)
	measurement := sha256.Sum256([]byte("mock-enclave-measurement"))
	attestation = append(attestation, measurement[:]...)
	attestation = append(attestation, measurement[:16]...) // Extend to 48 bytes

	// Signer ID (simulated - 32 bytes)
	signerID := sha256.Sum256([]byte("mock-enclave-signer"))
	attestation = append(attestation, signerID[:]...)

	// Report data (64 bytes) - includes user data and padding
	reportData := make([]byte, 64)
	copy(reportData, nonce)
	attestation = append(attestation, reportData...)

	// Simulated signature (64 bytes)
	b.mu.RLock()
	key := b.sealingKey
	b.mu.RUnlock()

	signatureData := append(attestation, key...)
	signature := sha256.Sum256(signatureData)
	attestation = append(attestation, signature[:]...)
	attestation = append(attestation, signature[:]...) // 64 bytes total

	// Padding to 512 bytes
	for len(attestation) < 512 {
		attestation = append(attestation, 0x00)
	}

	return attestation
}

// DeriveKey derives a key from the mock root of trust.
func (b *MockBackend) DeriveKey(context []byte, keySize int) ([]byte, error) {
	startTime := time.Now()

	if b.config.KeyDerivationDelay > 0 {
		time.Sleep(b.config.KeyDerivationDelay)
	}

	b.mu.RLock()
	initialized := b.initialized
	sealingKey := b.sealingKey
	b.mu.RUnlock()

	if !initialized {
		err := ErrNotInitialized
		b.recordCallWithDuration("DeriveKey", []interface{}{context, keySize}, time.Since(startTime), err)
		return nil, err
	}

	var derivedKey []byte
	var err error

	if b.config.FailKeyDerivation {
		err = b.config.KeyDerivationError
		if err == nil {
			err = ErrKeyDerivationFailed
		}
	} else if b.config.KeyDerivationFunc != nil {
		derivedKey, err = b.config.KeyDerivationFunc(context, keySize)
	} else {
		// Simple deterministic key derivation
		derivedKey = make([]byte, keySize)
		h := sha256.New()
		h.Write(sealingKey)
		h.Write(context)
		seed := h.Sum(nil)

		for i := 0; i < keySize; i++ {
			derivedKey[i] = seed[i%len(seed)]
		}
	}

	b.recordCallWithDuration("DeriveKey", []interface{}{context, keySize}, time.Since(startTime), err)
	return derivedKey, err
}

// Seal encrypts data using the mock sealing key.
func (b *MockBackend) Seal(plaintext []byte) ([]byte, error) {
	startTime := time.Now()

	if b.config.SealDelay > 0 {
		time.Sleep(b.config.SealDelay)
	}

	b.mu.RLock()
	initialized := b.initialized
	sealingKey := b.sealingKey
	b.mu.RUnlock()

	if !initialized {
		err := ErrNotInitialized
		b.recordCallWithDuration("Seal", []interface{}{plaintext}, time.Since(startTime), err)
		return nil, err
	}

	var sealed []byte
	var err error

	if b.config.FailSeal {
		err = b.config.SealError
		if err == nil {
			err = ErrSealingFailed
		}
	} else {
		// Generate nonce
		nonce := make([]byte, 12)
		_, _ = rand.Read(nonce)

		// Simple XOR encryption (NOT SECURE - for testing only)
		ciphertext := make([]byte, len(plaintext))
		keyStream := b.expandKeyStream(sealingKey, nonce, len(plaintext))
		for i, p := range plaintext {
			ciphertext[i] = p ^ keyStream[i]
		}

		// Create authenticated sealed blob
		// Format: version(1) + platform(1) + nonce(12) + ciphertext + mac(32)
		sealed = make([]byte, 0, 1+1+12+len(ciphertext)+32)
		sealed = append(sealed, 0x01)             // Version
		sealed = append(sealed, byte(b.platform)) // Platform
		sealed = append(sealed, nonce...)
		sealed = append(sealed, ciphertext...)

		// Compute MAC
		h := sha256.New()
		h.Write(sealingKey)
		h.Write(sealed)
		mac := h.Sum(nil)
		sealed = append(sealed, mac...)
	}

	b.recordCallWithDuration("Seal", []interface{}{plaintext}, time.Since(startTime), err)
	return sealed, err
}

// Unseal decrypts data that was previously sealed.
func (b *MockBackend) Unseal(ciphertext []byte) ([]byte, error) {
	startTime := time.Now()

	if b.config.UnsealDelay > 0 {
		time.Sleep(b.config.UnsealDelay)
	}

	b.mu.RLock()
	initialized := b.initialized
	sealingKey := b.sealingKey
	b.mu.RUnlock()

	if !initialized {
		err := ErrNotInitialized
		b.recordCallWithDuration("Unseal", []interface{}{ciphertext}, time.Since(startTime), err)
		return nil, err
	}

	var plaintext []byte
	var err error

	if b.config.FailUnseal {
		err = b.config.UnsealError
		if err == nil {
			err = ErrUnsealingFailed
		}
	} else {
		// Parse sealed blob
		// Format: version(1) + platform(1) + nonce(12) + ciphertext + mac(32)
		minLen := 1 + 1 + 12 + 32 // Minimum: header + nonce + mac
		if len(ciphertext) < minLen {
			err = fmt.Errorf("%w: sealed data too short", ErrUnsealingFailed)
		} else {
			version := ciphertext[0]
			if version != 0x01 {
				err = fmt.Errorf("%w: unsupported version %d", ErrUnsealingFailed, version)
			} else {
				nonce := ciphertext[2:14]
				encryptedData := ciphertext[14 : len(ciphertext)-32]
				mac := ciphertext[len(ciphertext)-32:]

				// Verify MAC
				h := sha256.New()
				h.Write(sealingKey)
				h.Write(ciphertext[:len(ciphertext)-32])
				expectedMAC := h.Sum(nil)

				if !constantTimeEqual(mac, expectedMAC) {
					err = fmt.Errorf("%w: MAC verification failed", ErrUnsealingFailed)
				} else {
					// Decrypt
					plaintext = make([]byte, len(encryptedData))
					keyStream := b.expandKeyStream(sealingKey, nonce, len(encryptedData))
					for i, c := range encryptedData {
						plaintext[i] = c ^ keyStream[i]
					}
				}
			}
		}
	}

	b.recordCallWithDuration("Unseal", []interface{}{ciphertext}, time.Since(startTime), err)
	return plaintext, err
}

// expandKeyStream expands a key and nonce into a keystream.
func (b *MockBackend) expandKeyStream(key, nonce []byte, length int) []byte {
	keyStream := make([]byte, length)
	h := sha256.New()
	counter := uint32(0)

	for i := 0; i < length; i += 32 {
		h.Reset()
		h.Write(key)
		h.Write(nonce)
		counterBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(counterBytes, counter)
		h.Write(counterBytes)
		block := h.Sum(nil)
		copy(keyStream[i:], block)
		counter++
	}

	return keyStream
}

// constantTimeEqual compares two byte slices in constant time.
func constantTimeEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// HealthCheck performs a health check on the mock backend.
func (b *MockBackend) HealthCheck() error {
	startTime := time.Now()

	if b.config.HealthCheckDelay > 0 {
		time.Sleep(b.config.HealthCheckDelay)
	}

	b.mu.RLock()
	initialized := b.initialized
	b.mu.RUnlock()

	var err error
	if b.config.FailHealthCheck {
		err = b.config.HealthCheckError
		if err == nil {
			err = ErrHealthCheckFailed
		}
	} else if !initialized {
		err = ErrNotInitialized
	}

	b.recordCallWithDuration("HealthCheck", nil, time.Since(startTime), err)
	return err
}

// GetCapabilities returns mock hardware capabilities.
func (b *MockBackend) GetCapabilities() *HardwareCapabilities {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.capabilities != nil {
		return b.capabilities
	}

	// Return default mock capabilities
	return &HardwareCapabilities{
		Platform:   b.platform,
		Available:  true,
		DetectedAt: time.Now(),
		TCBStatus:  TCBStatusUpToDate,
	}
}

// GetInfo returns mock hardware information.
func (b *MockBackend) GetInfo() *HardwareInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.hardwareInfo != nil {
		return b.hardwareInfo
	}

	// Return default mock info
	return &HardwareInfo{
		Platform:    b.platform,
		Name:        fmt.Sprintf("Mock %s Backend", b.platform),
		Vendor:      "VirtEngine Test",
		Version:     "1.0.0-mock",
		HardwareID:  []byte("mock-hardware-id"),
		TCBStatus:   TCBStatusUpToDate,
		LastUpdated: time.Now(),
		Features: map[string]bool{
			"attestation":    true,
			"sealing":        true,
			"key_derivation": true,
		},
	}
}

// =============================================================================
// Mock Control Methods
// =============================================================================

// SetCapabilities sets custom capabilities for the mock.
func (b *MockBackend) SetCapabilities(caps *HardwareCapabilities) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.capabilities = caps
}

// SetHardwareInfo sets custom hardware info for the mock.
func (b *MockBackend) SetHardwareInfo(info *HardwareInfo) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.hardwareInfo = info
}

// SetSealingKey sets the sealing key for the mock.
func (b *MockBackend) SetSealingKey(key []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sealingKey = make([]byte, len(key))
	copy(b.sealingKey, key)
}

// SetInitialized sets the initialization state.
func (b *MockBackend) SetInitialized(initialized bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initialized = initialized
}

// ConfigureFailure configures the mock to fail on a specific operation.
func (b *MockBackend) ConfigureFailure(operation string, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch operation {
	case "Initialize":
		b.config.FailInitialize = true
		b.config.InitializeError = err
	case "Shutdown":
		b.config.FailShutdown = true
		b.config.ShutdownError = err
	case "GetAttestation":
		b.config.FailAttestation = true
		b.config.AttestationError = err
	case "DeriveKey":
		b.config.FailKeyDerivation = true
		b.config.KeyDerivationError = err
	case "Seal":
		b.config.FailSeal = true
		b.config.SealError = err
	case "Unseal":
		b.config.FailUnseal = true
		b.config.UnsealError = err
	case "HealthCheck":
		b.config.FailHealthCheck = true
		b.config.HealthCheckError = err
	}
}

// ClearFailure clears a configured failure.
func (b *MockBackend) ClearFailure(operation string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch operation {
	case "Initialize":
		b.config.FailInitialize = false
		b.config.InitializeError = nil
	case "Shutdown":
		b.config.FailShutdown = false
		b.config.ShutdownError = nil
	case "GetAttestation":
		b.config.FailAttestation = false
		b.config.AttestationError = nil
	case "DeriveKey":
		b.config.FailKeyDerivation = false
		b.config.KeyDerivationError = nil
	case "Seal":
		b.config.FailSeal = false
		b.config.SealError = nil
	case "Unseal":
		b.config.FailUnseal = false
		b.config.UnsealError = nil
	case "HealthCheck":
		b.config.FailHealthCheck = false
		b.config.HealthCheckError = nil
	}
}

// ClearAllFailures clears all configured failures.
func (b *MockBackend) ClearAllFailures() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.config.FailInitialize = false
	b.config.FailShutdown = false
	b.config.FailAttestation = false
	b.config.FailKeyDerivation = false
	b.config.FailSeal = false
	b.config.FailUnseal = false
	b.config.FailHealthCheck = false

	b.config.InitializeError = nil
	b.config.ShutdownError = nil
	b.config.AttestationError = nil
	b.config.KeyDerivationError = nil
	b.config.SealError = nil
	b.config.UnsealError = nil
	b.config.HealthCheckError = nil
}

// ConfigureDelay configures a delay for a specific operation.
func (b *MockBackend) ConfigureDelay(operation string, delay time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch operation {
	case "Initialize":
		b.config.InitializeDelay = delay
	case "GetAttestation":
		b.config.AttestationDelay = delay
	case "DeriveKey":
		b.config.KeyDerivationDelay = delay
	case "Seal":
		b.config.SealDelay = delay
	case "Unseal":
		b.config.UnsealDelay = delay
	case "HealthCheck":
		b.config.HealthCheckDelay = delay
	}
}

// =============================================================================
// Call Recording Methods
// =============================================================================

// GetCalls returns all recorded method calls.
func (b *MockBackend) GetCalls() []MethodCall {
	b.mu.RLock()
	defer b.mu.RUnlock()

	calls := make([]MethodCall, len(b.calls))
	copy(calls, b.calls)
	return calls
}

// GetCallsForMethod returns recorded calls for a specific method.
func (b *MockBackend) GetCallsForMethod(method string) []MethodCall {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var calls []MethodCall
	for _, call := range b.calls {
		if call.Method == method {
			calls = append(calls, call)
		}
	}
	return calls
}

// GetCallCount returns the number of calls for a specific method.
func (b *MockBackend) GetCallCount(method string) int {
	return len(b.GetCallsForMethod(method))
}

// GetTotalCallCount returns the total number of recorded calls.
func (b *MockBackend) GetTotalCallCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.calls)
}

// ClearCalls clears all recorded calls.
func (b *MockBackend) ClearCalls() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.calls = make([]MethodCall, 0)
}

// WasMethodCalled returns true if the specified method was called.
func (b *MockBackend) WasMethodCalled(method string) bool {
	return b.GetCallCount(method) > 0
}

// GetLastCall returns the most recent call, or nil if no calls recorded.
func (b *MockBackend) GetLastCall() *MethodCall {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.calls) == 0 {
		return nil
	}
	call := b.calls[len(b.calls)-1]
	return &call
}

// GetLastCallForMethod returns the most recent call for a method, or nil.
func (b *MockBackend) GetLastCallForMethod(method string) *MethodCall {
	calls := b.GetCallsForMethod(method)
	if len(calls) == 0 {
		return nil
	}
	return &calls[len(calls)-1]
}

// recordCall records a method call.
func (b *MockBackend) recordCall(method string, args []interface{}) {
	b.recordCallWithDuration(method, args, 0, nil)
}

// recordCallWithDuration records a method call with duration and error.
func (b *MockBackend) recordCallWithDuration(method string, args []interface{}, duration time.Duration, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.calls = append(b.calls, MethodCall{
		Method:    method,
		Args:      args,
		Timestamp: time.Now(),
		Duration:  duration,
		Error:     err,
	})
}

// =============================================================================
// Assertion Helpers
// =============================================================================

// AssertCalled asserts that a method was called at least once.
func (b *MockBackend) AssertCalled(method string) error {
	if !b.WasMethodCalled(method) {
		return fmt.Errorf("method %s was not called", method)
	}
	return nil
}

// AssertNotCalled asserts that a method was not called.
func (b *MockBackend) AssertNotCalled(method string) error {
	if b.WasMethodCalled(method) {
		return fmt.Errorf("method %s was called %d times", method, b.GetCallCount(method))
	}
	return nil
}

// AssertCallCount asserts that a method was called exactly n times.
func (b *MockBackend) AssertCallCount(method string, count int) error {
	actual := b.GetCallCount(method)
	if actual != count {
		return fmt.Errorf("method %s was called %d times, expected %d", method, actual, count)
	}
	return nil
}

// Reset resets the mock to its initial state.
func (b *MockBackend) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initialized = false
	b.sealingKey = nil
	b.calls = make([]MethodCall, 0)
	b.hardwareInfo = nil
	b.capabilities = nil

	// Reset configuration to defaults but keep platform
	platform := b.config.Platform
	b.config = DefaultMockConfig()
	b.config.Platform = platform
}
