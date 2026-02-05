// Package hardware provides a comprehensive hardware abstraction layer for TEE integration.
//
// This file contains tests for the hardware abstraction layer.
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package hardware

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

// =============================================================================
// Platform Tests
// =============================================================================

func TestPlatform_String(t *testing.T) {
	tests := []struct {
		platform Platform
		expected string
	}{
		{PlatformUnknown, "unknown"},
		{PlatformSGX, "sgx"},
		{PlatformSEVSNP, "sev-snp"},
		{PlatformNitro, "nitro"},
		{PlatformSimulated, "simulated"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.platform.String(); got != tt.expected {
				t.Errorf("Platform.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParsePlatform(t *testing.T) {
	tests := []struct {
		input    string
		expected Platform
	}{
		{"sgx", PlatformSGX},
		{"SGX", PlatformSGX},
		{"intel-sgx", PlatformSGX},
		{"sev-snp", PlatformSEVSNP},
		{"SEV-SNP", PlatformSEVSNP},
		{"sev", PlatformSEVSNP},
		{"nitro", PlatformNitro},
		{"Nitro", PlatformNitro},
		{"aws-nitro", PlatformNitro},
		{"simulated", PlatformSimulated},
		{"sim", PlatformSimulated},
		{"unknown-platform", PlatformUnknown},
		{"", PlatformUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ParsePlatform(tt.input); got != tt.expected {
				t.Errorf("ParsePlatform(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPlatform_IsHardware(t *testing.T) {
	tests := []struct {
		platform Platform
		expected bool
	}{
		{PlatformUnknown, false},
		{PlatformSGX, true},
		{PlatformSEVSNP, true},
		{PlatformNitro, true},
		{PlatformSimulated, false},
	}

	for _, tt := range tests {
		t.Run(tt.platform.String(), func(t *testing.T) {
			if got := tt.platform.IsHardware(); got != tt.expected {
				t.Errorf("Platform.IsHardware() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// =============================================================================
// TCB Status Tests
// =============================================================================

func TestTCBStatus_String(t *testing.T) {
	tests := []struct {
		status   TCBStatus
		expected string
	}{
		{TCBStatusUnknown, "Unknown"},
		{TCBStatusUpToDate, "UpToDate"},
		{TCBStatusOutOfDate, "OutOfDate"},
		{TCBStatusConfigurationNeeded, "ConfigurationNeeded"},
		{TCBStatusRevoked, "Revoked"},
		{TCBStatusSWHardeningNeeded, "SWHardeningNeeded"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("TCBStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTCBStatus_IsSecure(t *testing.T) {
	tests := []struct {
		status   TCBStatus
		expected bool
	}{
		{TCBStatusUnknown, false},
		{TCBStatusUpToDate, true},
		{TCBStatusOutOfDate, false},
		{TCBStatusConfigurationNeeded, false},
		{TCBStatusRevoked, false},
		{TCBStatusSWHardeningNeeded, true},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			if got := tt.status.IsSecure(); got != tt.expected {
				t.Errorf("TCBStatus.IsSecure() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// =============================================================================
// KeyPolicy Tests
// =============================================================================

func TestKeyPolicy_String(t *testing.T) {
	tests := []struct {
		policy   KeyPolicy
		expected string
	}{
		{KeyPolicyEnclave, "enclave"},
		{KeyPolicySigner, "signer"},
		{KeyPolicyPlatform, "platform"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.policy.String(); got != tt.expected {
				t.Errorf("KeyPolicy.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Config Tests
// =============================================================================

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.PreferredPlatform != PlatformUnknown {
		t.Errorf("DefaultConfig().PreferredPlatform = %v, want %v", config.PreferredPlatform, PlatformUnknown)
	}

	if config.RequireHardware {
		t.Error("DefaultConfig().RequireHardware should be false")
	}

	if !config.AllowSimulation {
		t.Error("DefaultConfig().AllowSimulation should be true")
	}

	if config.HealthCheckInterval != 30*time.Second {
		t.Errorf("DefaultConfig().HealthCheckInterval = %v, want %v", config.HealthCheckInterval, 30*time.Second)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "require hardware with simulation",
			config: Config{
				RequireHardware:     true,
				AllowSimulation:     true,
				HealthCheckInterval: 30 * time.Second,
				HealthCheckTimeout:  5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "health check interval too short",
			config: Config{
				HealthCheckInterval: 100 * time.Millisecond,
				HealthCheckTimeout:  5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "health check timeout too short",
			config: Config{
				HealthCheckInterval: 30 * time.Second,
				HealthCheckTimeout:  50 * time.Millisecond,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// Hardware Capabilities Tests
// =============================================================================

func TestHardwareCapabilities_HasAnyHardware(t *testing.T) {
	tests := []struct {
		name     string
		caps     HardwareCapabilities
		expected bool
	}{
		{
			name:     "no hardware",
			caps:     HardwareCapabilities{},
			expected: false,
		},
		{
			name: "SGX available",
			caps: HardwareCapabilities{
				SGX: SGXCapabilities{Available: true},
			},
			expected: true,
		},
		{
			name: "SEV-SNP available",
			caps: HardwareCapabilities{
				SEVSNP: SEVSNPCapabilities{Available: true},
			},
			expected: true,
		},
		{
			name: "Nitro available",
			caps: HardwareCapabilities{
				Nitro: NitroCapabilities{Available: true},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.caps.HasAnyHardware(); got != tt.expected {
				t.Errorf("HardwareCapabilities.HasAnyHardware() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHardwareCapabilities_GetRecommendedPlatform(t *testing.T) {
	tests := []struct {
		name     string
		caps     HardwareCapabilities
		expected Platform
	}{
		{
			name:     "no hardware returns simulated",
			caps:     HardwareCapabilities{},
			expected: PlatformSimulated,
		},
		{
			name: "SGX with FLC and DCAP is preferred",
			caps: HardwareCapabilities{
				SGX: SGXCapabilities{
					Available:     true,
					FLCSupported:  true,
					DCAPAvailable: true,
				},
				SEVSNP: SEVSNPCapabilities{Available: true},
			},
			expected: PlatformSGX,
		},
		{
			name: "SEV-SNP preferred when SGX lacks FLC",
			caps: HardwareCapabilities{
				SGX:    SGXCapabilities{Available: true},
				SEVSNP: SEVSNPCapabilities{Available: true},
			},
			expected: PlatformSEVSNP,
		},
		{
			name: "Nitro when only option",
			caps: HardwareCapabilities{
				Nitro: NitroCapabilities{Available: true},
			},
			expected: PlatformNitro,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.caps.GetRecommendedPlatform(); got != tt.expected {
				t.Errorf("HardwareCapabilities.GetRecommendedPlatform() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Hardware Error Tests
// =============================================================================

func TestHardwareError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *HardwareError
		contains []string
	}{
		{
			name: "basic error",
			err: &HardwareError{
				Platform:   PlatformSGX,
				Operation:  "initialize",
				Underlying: errors.New("device not found"),
			},
			contains: []string{"sgx", "initialize", "device not found"},
		},
		{
			name: "error with device path",
			err: &HardwareError{
				Platform:   PlatformSEVSNP,
				Operation:  "attestation",
				DevicePath: "/dev/sev-guest",
				Underlying: errors.New("permission denied"),
			},
			contains: []string{"sev-snp", "attestation", "/dev/sev-guest", "permission denied"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			for _, s := range tt.contains {
				if !bytes.Contains([]byte(errMsg), []byte(s)) {
					t.Errorf("HardwareError.Error() = %q, should contain %q", errMsg, s)
				}
			}
		})
	}
}

func TestHardwareError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &HardwareError{
		Platform:   PlatformSGX,
		Operation:  "test",
		Underlying: underlying,
	}

	if !errors.Is(err, underlying) {
		t.Error("HardwareError should unwrap to underlying error")
	}
}

// =============================================================================
// Mock Backend Tests
// =============================================================================

func TestMockBackend_Initialize(t *testing.T) {
	mock := NewMockBackendWithDefaults()

	err := mock.Initialize()
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if !mock.WasMethodCalled("Initialize") {
		t.Error("Initialize should be recorded")
	}
}

func TestMockBackend_Initialize_Failure(t *testing.T) {
	expectedErr := errors.New("init failed")
	mock := NewMockBackend(MockConfig{
		FailInitialize:  true,
		InitializeError: expectedErr,
	})

	err := mock.Initialize()
	if err != expectedErr {
		t.Errorf("Initialize() error = %v, want %v", err, expectedErr)
	}
}

func TestMockBackend_GetAttestation(t *testing.T) {
	mock := NewMockBackendWithDefaults()
	if err := mock.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	nonce := []byte("test-nonce-12345")
	attestation, err := mock.GetAttestation(nonce)
	if err != nil {
		t.Fatalf("GetAttestation() error = %v", err)
	}

	if len(attestation) == 0 {
		t.Error("GetAttestation() returned empty attestation")
	}

	// Verify call was recorded
	calls := mock.GetCallsForMethod("GetAttestation")
	if len(calls) != 1 {
		t.Errorf("Expected 1 GetAttestation call, got %d", len(calls))
	}
}

func TestMockBackend_GetAttestation_NotInitialized(t *testing.T) {
	mock := NewMockBackendWithDefaults()

	_, err := mock.GetAttestation([]byte("nonce"))
	if !errors.Is(err, ErrNotInitialized) {
		t.Errorf("GetAttestation() error = %v, want ErrNotInitialized", err)
	}
}

func TestMockBackend_SealUnseal(t *testing.T) {
	mock := NewMockBackendWithDefaults()
	if err := mock.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	plaintext := []byte("sensitive data to seal")

	sealed, err := mock.Seal(plaintext)
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}

	unsealed, err := mock.Unseal(sealed)
	if err != nil {
		t.Fatalf("Unseal() error = %v", err)
	}

	if !bytes.Equal(plaintext, unsealed) {
		t.Errorf("Unseal() = %v, want %v", unsealed, plaintext)
	}
}

func TestMockBackend_DeriveKey(t *testing.T) {
	mock := NewMockBackendWithDefaults()
	if err := mock.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	context := []byte("key derivation context")
	keySize := 32

	key, err := mock.DeriveKey(context, keySize)
	if err != nil {
		t.Fatalf("DeriveKey() error = %v", err)
	}

	if len(key) != keySize {
		t.Errorf("DeriveKey() returned key of length %d, want %d", len(key), keySize)
	}

	// Verify same context produces same key
	key2, err := mock.DeriveKey(context, keySize)
	if err != nil {
		t.Fatalf("DeriveKey() error = %v", err)
	}

	if !bytes.Equal(key, key2) {
		t.Error("DeriveKey() should produce deterministic keys for same context")
	}

	// Different context should produce different key
	key3, err := mock.DeriveKey([]byte("different context"), keySize)
	if err != nil {
		t.Fatalf("DeriveKey() error = %v", err)
	}

	if bytes.Equal(key, key3) {
		t.Error("DeriveKey() should produce different keys for different contexts")
	}
}

func TestMockBackend_HealthCheck(t *testing.T) {
	mock := NewMockBackendWithDefaults()

	// Not initialized should fail
	if err := mock.HealthCheck(); err == nil {
		t.Error("HealthCheck() should fail when not initialized")
	}

	// After initialization should succeed
	if err := mock.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if err := mock.HealthCheck(); err != nil {
		t.Errorf("HealthCheck() error = %v", err)
	}
}

func TestMockBackend_ConfigureFailure(t *testing.T) {
	mock := NewMockBackendWithDefaults()
	if err := mock.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	expectedErr := errors.New("custom error")
	mock.ConfigureFailure("GetAttestation", expectedErr)

	_, err := mock.GetAttestation([]byte("nonce"))
	if err != expectedErr {
		t.Errorf("GetAttestation() error = %v, want %v", err, expectedErr)
	}

	mock.ClearFailure("GetAttestation")

	_, err = mock.GetAttestation([]byte("nonce"))
	if err != nil {
		t.Errorf("GetAttestation() after ClearFailure error = %v", err)
	}
}

func TestMockBackend_CallRecording(t *testing.T) {
	mock := NewMockBackendWithDefaults()
	if err := mock.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Make some calls
	_, _ = mock.GetAttestation([]byte("nonce1"))
	_, _ = mock.GetAttestation([]byte("nonce2"))
	_ = mock.HealthCheck()

	// Check call counts
	if count := mock.GetCallCount("GetAttestation"); count != 2 {
		t.Errorf("GetCallCount(GetAttestation) = %d, want 2", count)
	}

	if count := mock.GetCallCount("HealthCheck"); count != 1 {
		t.Errorf("GetCallCount(HealthCheck) = %d, want 1", count)
	}

	// Check assertions
	if err := mock.AssertCalled("GetAttestation"); err != nil {
		t.Errorf("AssertCalled(GetAttestation) = %v", err)
	}

	if err := mock.AssertNotCalled("DeriveKey"); err != nil {
		t.Errorf("AssertNotCalled(DeriveKey) = %v", err)
	}

	if err := mock.AssertCallCount("GetAttestation", 2); err != nil {
		t.Errorf("AssertCallCount(GetAttestation, 2) = %v", err)
	}

	// Check last call
	lastCall := mock.GetLastCall()
	if lastCall == nil || lastCall.Method != "HealthCheck" {
		t.Error("GetLastCall() should return HealthCheck")
	}

	// Clear calls
	mock.ClearCalls()
	if count := mock.GetTotalCallCount(); count != 0 {
		t.Errorf("GetTotalCallCount() after ClearCalls = %d, want 0", count)
	}
}

func TestMockBackend_Reset(t *testing.T) {
	mock := NewMockBackendWithDefaults()
	if err := mock.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	_, _ = mock.GetAttestation([]byte("nonce"))
	mock.ConfigureFailure("HealthCheck", errors.New("fail"))

	mock.Reset()

	// Calls should be cleared
	if count := mock.GetTotalCallCount(); count != 0 {
		t.Errorf("After Reset(), call count = %d, want 0", count)
	}

	// Should no longer be initialized (this call will be recorded after count check)
	if err := mock.HealthCheck(); !errors.Is(err, ErrNotInitialized) {
		t.Error("After Reset(), mock should not be initialized")
	}
}

// =============================================================================
// Hardware Manager Tests
// =============================================================================

func TestNewHardwareManager(t *testing.T) {
	manager, err := NewHardwareManagerWithDefaults()
	if err != nil {
		t.Fatalf("NewHardwareManagerWithDefaults() error = %v", err)
	}

	if manager == nil {
		t.Error("NewHardwareManagerWithDefaults() returned nil")
	}
}

func TestHardwareManager_InvalidConfig(t *testing.T) {
	config := Config{
		RequireHardware: true,
		AllowSimulation: true,
	}

	_, err := NewHardwareManager(config)
	if err == nil {
		t.Error("NewHardwareManager() with invalid config should return error")
	}
}

func TestHardwareManager_Initialize(t *testing.T) {
	manager, err := NewHardwareManagerWithDefaults()
	if err != nil {
		t.Fatalf("NewHardwareManagerWithDefaults() error = %v", err)
	}

	ctx := context.Background()
	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if !manager.IsInitialized() {
		t.Error("Manager should be initialized")
	}

	// Cleanup
	if err := manager.Shutdown(); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

func TestHardwareManager_DoubleInitialize(t *testing.T) {
	manager, err := NewHardwareManagerWithDefaults()
	if err != nil {
		t.Fatalf("NewHardwareManagerWithDefaults() error = %v", err)
	}

	ctx := context.Background()
	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Second initialize should be a no-op
	if err := manager.Initialize(ctx); err != nil {
		t.Errorf("Second Initialize() error = %v", err)
	}

	// Cleanup
	_ = manager.Shutdown()
}

func TestSimulatedBackend_Operations(t *testing.T) {
	backend := NewSimulatedBackend()

	// Initialize
	if err := backend.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Platform
	if backend.Platform() != PlatformSimulated {
		t.Errorf("Platform() = %v, want %v", backend.Platform(), PlatformSimulated)
	}

	// IsAvailable
	if !backend.IsAvailable() {
		t.Error("Simulated backend should always be available")
	}

	// Attestation
	attestation, err := backend.GetAttestation([]byte("nonce"))
	if err != nil {
		t.Fatalf("GetAttestation() error = %v", err)
	}
	if len(attestation) == 0 {
		t.Error("GetAttestation() returned empty attestation")
	}

	// Key derivation
	key, err := backend.DeriveKey([]byte("context"), 32)
	if err != nil {
		t.Fatalf("DeriveKey() error = %v", err)
	}
	if len(key) != 32 {
		t.Errorf("DeriveKey() returned key of length %d, want 32", len(key))
	}

	// Seal/Unseal
	plaintext := []byte("test data")
	sealed, err := backend.Seal(plaintext)
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}

	unsealed, err := backend.Unseal(sealed)
	if err != nil {
		t.Fatalf("Unseal() error = %v", err)
	}
	if !bytes.Equal(plaintext, unsealed) {
		t.Errorf("Unseal() = %v, want %v", unsealed, plaintext)
	}

	// Health check
	if err := backend.HealthCheck(); err != nil {
		t.Errorf("HealthCheck() error = %v", err)
	}

	// Shutdown
	if err := backend.Shutdown(); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

// =============================================================================
// Detector Tests
// =============================================================================

func TestUnifiedDetector_Detect(t *testing.T) {
	detector := NewUnifiedDetector()

	caps, err := detector.Detect()
	if err != nil {
		// Some errors are expected on non-Linux platforms
		t.Logf("Detect() returned error (may be expected): %v", err)
	}

	if caps == nil {
		t.Fatal("Detect() returned nil capabilities")
	}

	// Verify detection was cached
	caps2, _ := detector.Detect()
	if caps2 != caps {
		t.Error("Detect() should return cached results")
	}
}

func TestGlobalDetector(t *testing.T) {
	detector1 := GetGlobalDetector()
	detector2 := GetGlobalDetector()

	if detector1 != detector2 {
		t.Error("GetGlobalDetector() should return the same instance")
	}
}
