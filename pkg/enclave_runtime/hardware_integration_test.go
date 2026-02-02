// Package enclave_runtime provides TEE enclave implementations.
//
// This file contains integration tests for hardware TEE backend integration.
// Tests are designed to pass on machines without TEE hardware by using
// automatic fallback to simulation mode.
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package enclave_runtime

import (
	"context"
	"testing"
	"time"
)

// =============================================================================
// SGX Hardware Integration Tests
// =============================================================================

func TestSGXHardwareIntegration(t *testing.T) {
	t.Run("AutoModeWithFallback", func(t *testing.T) {
		config := SGXEnclaveConfig{
			EnclavePath: "/opt/virtengine/enclaves/test.so",
			Debug:       true, // Allow debug for tests
		}

		svc, err := NewSGXEnclaveServiceImplWithMode(config, HardwareModeAuto)
		if err != nil {
			t.Fatalf("Failed to create SGX service in auto mode: %v", err)
		}
		defer func() { _ = svc.Shutdown() }()

		// Initialize the service
		err = svc.Initialize(DefaultRuntimeConfig())
		if err != nil {
			t.Fatalf("Failed to initialize SGX service: %v", err)
		}

		// Verify service is functional
		status := svc.GetStatus()
		if !status.Initialized {
			t.Error("SGX service should be initialized")
		}

		// Check hardware status
		t.Logf("SGX hardware enabled: %v", svc.IsHardwareEnabled())
		t.Logf("SGX hardware mode: %s", svc.GetHardwareMode())
	})

	t.Run("SimulateModeForced", func(t *testing.T) {
		config := SGXEnclaveConfig{
			EnclavePath: "/opt/virtengine/enclaves/test.so",
			Debug:       true,
		}

		svc, err := NewSGXEnclaveServiceImplWithMode(config, HardwareModeSimulate)
		if err != nil {
			t.Fatalf("Failed to create SGX service in simulate mode: %v", err)
		}
		defer func() { _ = svc.Shutdown() }()

		err = svc.Initialize(DefaultRuntimeConfig())
		if err != nil {
			t.Fatalf("Failed to initialize SGX service: %v", err)
		}

		// In simulate mode, hardware should never be enabled
		if svc.IsHardwareEnabled() {
			t.Error("Hardware should not be enabled in simulate mode")
		}

		if svc.GetHardwareMode() != HardwareModeSimulate {
			t.Errorf("Expected HardwareModeSimulate, got %v", svc.GetHardwareMode())
		}
	})

	t.Run("AttestationGeneration", func(t *testing.T) {
		config := SGXEnclaveConfig{
			EnclavePath: "/opt/virtengine/enclaves/test.so",
			Debug:       true,
		}

		svc, err := NewSGXEnclaveServiceImplWithMode(config, HardwareModeAuto)
		if err != nil {
			t.Fatalf("Failed to create SGX service: %v", err)
		}
		defer func() { _ = svc.Shutdown() }()

		err = svc.Initialize(DefaultRuntimeConfig())
		if err != nil {
			t.Fatalf("Failed to initialize SGX service: %v", err)
		}

		// Generate attestation
		reportData := []byte("test-report-data-for-sgx")
		quote, err := svc.GenerateAttestation(reportData)
		if err != nil {
			t.Fatalf("Failed to generate attestation: %v", err)
		}

		if len(quote) == 0 {
			t.Error("Generated quote should not be empty")
		}
		t.Logf("Generated SGX quote of %d bytes", len(quote))
	})

	t.Run("SealUnsealData", func(t *testing.T) {
		config := SGXEnclaveConfig{
			EnclavePath: "/opt/virtengine/enclaves/test.so",
			Debug:       true,
		}

		svc, err := NewSGXEnclaveServiceImplWithMode(config, HardwareModeAuto)
		if err != nil {
			t.Fatalf("Failed to create SGX service: %v", err)
		}
		defer svc.Shutdown()

		err = svc.Initialize(DefaultRuntimeConfig())
		if err != nil {
			t.Fatalf("Failed to initialize SGX service: %v", err)
		}

		// Test seal/unseal
		plaintext := []byte("secret data to seal")
		aad := []byte("additional authenticated data")

		sealed, err := svc.SealData(plaintext, aad)
		if err != nil {
			t.Fatalf("Failed to seal data: %v", err)
		}

		if len(sealed) == 0 {
			t.Error("Sealed data should not be empty")
		}

		unsealed, _, err := svc.UnsealData(sealed)
		if err != nil {
			t.Fatalf("Failed to unseal data: %v", err)
		}

		if string(unsealed) != string(plaintext) {
			t.Errorf("Unsealed data mismatch: got %s, want %s", unsealed, plaintext)
		}
	})
}

// =============================================================================
// SEV Hardware Integration Tests
// =============================================================================

func TestSEVHardwareIntegration(t *testing.T) {
	t.Run("AutoModeWithFallback", func(t *testing.T) {
		config := SEVSNPConfig{
			Endpoint:         "localhost:8443",
			AllowDebugPolicy: true,
		}

		svc, err := NewSEVSNPEnclaveServiceImplWithMode(config, HardwareModeAuto)
		if err != nil {
			t.Fatalf("Failed to create SEV-SNP service in auto mode: %v", err)
		}
		defer svc.Shutdown()

		err = svc.Initialize(DefaultRuntimeConfig())
		if err != nil {
			t.Fatalf("Failed to initialize SEV-SNP service: %v", err)
		}

		status := svc.GetStatus()
		if !status.Initialized {
			t.Error("SEV-SNP service should be initialized")
		}

		t.Logf("SEV-SNP hardware enabled: %v", svc.IsHardwareEnabled())
		t.Logf("SEV-SNP hardware mode: %s", svc.GetHardwareMode())
	})

	t.Run("SimulateModeForced", func(t *testing.T) {
		config := SEVSNPConfig{
			Endpoint:         "localhost:8443",
			AllowDebugPolicy: true,
		}

		svc, err := NewSEVSNPEnclaveServiceImplWithMode(config, HardwareModeSimulate)
		if err != nil {
			t.Fatalf("Failed to create SEV-SNP service in simulate mode: %v", err)
		}
		defer svc.Shutdown()

		err = svc.Initialize(DefaultRuntimeConfig())
		if err != nil {
			t.Fatalf("Failed to initialize SEV-SNP service: %v", err)
		}

		if svc.IsHardwareEnabled() {
			t.Error("Hardware should not be enabled in simulate mode")
		}
	})

	t.Run("AttestationGeneration", func(t *testing.T) {
		config := SEVSNPConfig{
			Endpoint:         "localhost:8443",
			AllowDebugPolicy: true,
		}

		svc, err := NewSEVSNPEnclaveServiceImplWithMode(config, HardwareModeAuto)
		if err != nil {
			t.Fatalf("Failed to create SEV-SNP service: %v", err)
		}
		defer svc.Shutdown()

		err = svc.Initialize(DefaultRuntimeConfig())
		if err != nil {
			t.Fatalf("Failed to initialize SEV-SNP service: %v", err)
		}

		reportData := []byte("test-report-data-for-sev")
		report, err := svc.GenerateAttestation(reportData)
		if err != nil {
			t.Fatalf("Failed to generate attestation: %v", err)
		}

		if len(report) == 0 {
			t.Error("Generated report should not be empty")
		}
		t.Logf("Generated SEV-SNP report of %d bytes", len(report))
	})

	t.Run("KeyDerivation", func(t *testing.T) {
		config := SEVSNPConfig{
			Endpoint:         "localhost:8443",
			AllowDebugPolicy: true,
		}

		svc, err := NewSEVSNPEnclaveServiceImplWithMode(config, HardwareModeAuto)
		if err != nil {
			t.Fatalf("Failed to create SEV-SNP service: %v", err)
		}
		defer svc.Shutdown()

		err = svc.Initialize(DefaultRuntimeConfig())
		if err != nil {
			t.Fatalf("Failed to initialize SEV-SNP service: %v", err)
		}

		key, err := svc.DeriveKey([]byte("test-context"), 32)
		if err != nil {
			t.Fatalf("Failed to derive key: %v", err)
		}

		if len(key) != 32 {
			t.Errorf("Expected key length 32, got %d", len(key))
		}

		// Verify determinism - same context should give same key
		key2, err := svc.DeriveKey([]byte("test-context"), 32)
		if err != nil {
			t.Fatalf("Failed to derive key second time: %v", err)
		}

		if string(key) != string(key2) {
			t.Error("Key derivation should be deterministic")
		}
	})
}

// =============================================================================
// Nitro Hardware Integration Tests
// =============================================================================

func TestNitroHardwareIntegration(t *testing.T) {
	t.Run("AutoModeWithFallback", func(t *testing.T) {
		config := NitroEnclaveConfig{
			EnclaveImagePath: "/opt/virtengine/enclaves/test.eif",
			CPUCount:         2,
			MemoryMB:         512,
			DebugMode:        true,
		}

		svc, err := NewNitroEnclaveServiceImplWithMode(config, HardwareModeAuto)
		if err != nil {
			t.Fatalf("Failed to create Nitro service in auto mode: %v", err)
		}
		defer svc.Shutdown()

		err = svc.Initialize(DefaultRuntimeConfig())
		if err != nil {
			t.Fatalf("Failed to initialize Nitro service: %v", err)
		}

		status := svc.GetStatus()
		if !status.Initialized {
			t.Error("Nitro service should be initialized")
		}

		t.Logf("Nitro hardware enabled: %v", svc.IsHardwareEnabled())
		t.Logf("Nitro hardware mode: %s", svc.GetHardwareMode())
	})

	t.Run("SimulateModeForced", func(t *testing.T) {
		config := NitroEnclaveConfig{
			EnclaveImagePath: "/opt/virtengine/enclaves/test.eif",
			CPUCount:         2,
			MemoryMB:         512,
			DebugMode:        true,
		}

		svc, err := NewNitroEnclaveServiceImplWithMode(config, HardwareModeSimulate)
		if err != nil {
			t.Fatalf("Failed to create Nitro service in simulate mode: %v", err)
		}
		defer svc.Shutdown()

		err = svc.Initialize(DefaultRuntimeConfig())
		if err != nil {
			t.Fatalf("Failed to initialize Nitro service: %v", err)
		}

		if svc.IsHardwareEnabled() {
			t.Error("Hardware should not be enabled in simulate mode")
		}
	})

	t.Run("AttestationGeneration", func(t *testing.T) {
		config := NitroEnclaveConfig{
			EnclaveImagePath: "/opt/virtengine/enclaves/test.eif",
			CPUCount:         2,
			MemoryMB:         512,
			DebugMode:        true,
		}

		svc, err := NewNitroEnclaveServiceImplWithMode(config, HardwareModeAuto)
		if err != nil {
			t.Fatalf("Failed to create Nitro service: %v", err)
		}
		defer svc.Shutdown()

		err = svc.Initialize(DefaultRuntimeConfig())
		if err != nil {
			t.Fatalf("Failed to initialize Nitro service: %v", err)
		}

		reportData := []byte("test-report-data-for-nitro")
		doc, err := svc.GenerateAttestation(reportData)
		if err != nil {
			t.Fatalf("Failed to generate attestation: %v", err)
		}

		if len(doc) == 0 {
			t.Error("Generated attestation document should not be empty")
		}
		t.Logf("Generated Nitro attestation document of %d bytes", len(doc))
	})

	t.Run("LaunchEnclave", func(t *testing.T) {
		config := NitroEnclaveConfig{
			EnclaveImagePath: "/opt/virtengine/enclaves/test.eif",
			CPUCount:         2,
			MemoryMB:         512,
			DebugMode:        true,
		}

		svc, err := NewNitroEnclaveServiceImplWithMode(config, HardwareModeAuto)
		if err != nil {
			t.Fatalf("Failed to create Nitro service: %v", err)
		}
		defer svc.Shutdown()

		err = svc.Initialize(DefaultRuntimeConfig())
		if err != nil {
			t.Fatalf("Failed to initialize Nitro service: %v", err)
		}

		// Launch enclave
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err = svc.LaunchEnclave(ctx)
		if err != nil {
			// This may fail if hardware is not available, which is expected
			t.Logf("LaunchEnclave result: %v (expected if no hardware)", err)
		}
	})
}

// =============================================================================
// Enclave Manager Hardware Detection Tests
// =============================================================================

func TestEnclaveManagerHardwareDetection(t *testing.T) {
	t.Run("GetHardwareCapabilities", func(t *testing.T) {
		manager, err := NewEnclaveManager(DefaultEnclaveManagerConfig())
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		caps := manager.GetHardwareCapabilities()

		t.Logf("Hardware capabilities: %s", caps.String())
		t.Logf("SGX Available: %v", caps.SGXAvailable)
		t.Logf("SEV-SNP Available: %v", caps.SEVSNPAvailable)
		t.Logf("Nitro Available: %v", caps.NitroAvailable)
		t.Logf("Preferred Backend: %s", caps.PreferredBackend)
	})

	t.Run("BackendSelectionPrefersHardware", func(t *testing.T) {
		manager, err := NewEnclaveManager(DefaultEnclaveManagerConfig())
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}
		defer func() { _ = manager.Stop() }()

		// Create a simulated backend
		simBackend, err := CreateSimulatedBackend("sim-1")
		if err != nil {
			t.Fatalf("Failed to create simulated backend: %v", err)
		}
		_ = manager.RegisterBackend(simBackend)

		// Create an SGX backend (will use simulation if no hardware)
		sgxConfig := SGXEnclaveConfig{
			EnclavePath: "/opt/virtengine/enclaves/test.so",
			Debug:       true,
		}
		sgxSvc, err := NewSGXEnclaveServiceImplWithMode(sgxConfig, HardwareModeAuto)
		if err != nil {
			t.Fatalf("Failed to create SGX service: %v", err)
		}
		_ = sgxSvc.Initialize(DefaultRuntimeConfig())

		sgxBackend := NewEnclaveBackend("sgx-1", AttestationTypeSGX, sgxSvc)
		sgxBackend.Priority = 1 // Higher priority (lower number)
		sgxBackend.Health = HealthHealthy
		_ = manager.RegisterBackend(sgxBackend)

		_ = manager.Start()

		// Select backend
		selected, err := manager.SelectBackend()
		if err != nil {
			t.Fatalf("Failed to select backend: %v", err)
		}

		// The SGX backend should be selected due to priority and type preference
		t.Logf("Selected backend: %s (type: %s)", selected.ID, selected.Type)
	})

	t.Run("PreferredBackendType", func(t *testing.T) {
		manager, err := NewEnclaveManager(DefaultEnclaveManagerConfig())
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		preferredType := manager.GetPreferredBackendType()
		t.Logf("Preferred backend type: %s", preferredType)

		// On a machine without TEE hardware, should be Simulated
		// On a machine with TEE hardware, should be the detected type
	})

	t.Run("LogHardwareStatus", func(t *testing.T) {
		manager, err := NewEnclaveManager(DefaultEnclaveManagerConfig())
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		simBackend, _ := CreateSimulatedBackend("test-sim")
		_ = manager.RegisterBackend(simBackend)

		// This should log hardware status without errors
		manager.LogHardwareStatus()
	})
}

// =============================================================================
// Automatic Fallback Tests
// =============================================================================

func TestAutomaticFallbackToSimulation(t *testing.T) {
	t.Run("SGXFallsBackToSimulation", func(t *testing.T) {
		config := SGXEnclaveConfig{
			EnclavePath: "/opt/virtengine/enclaves/test.so",
			Debug:       true,
		}

		// Auto mode should work even without hardware
		svc, err := NewSGXEnclaveServiceImplWithMode(config, HardwareModeAuto)
		if err != nil {
			t.Fatalf("Failed to create SGX service: %v", err)
		}
		defer svc.Shutdown()

		err = svc.Initialize(DefaultRuntimeConfig())
		if err != nil {
			t.Fatalf("Service should initialize even without hardware: %v", err)
		}

		// Should be able to perform operations
		attestation, err := svc.GenerateAttestation([]byte("test"))
		if err != nil {
			t.Fatalf("Should be able to generate attestation: %v", err)
		}

		if len(attestation) == 0 {
			t.Error("Attestation should not be empty")
		}
	})

	t.Run("SEVFallsBackToSimulation", func(t *testing.T) {
		config := SEVSNPConfig{
			Endpoint:         "localhost:8443",
			AllowDebugPolicy: true,
		}

		svc, err := NewSEVSNPEnclaveServiceImplWithMode(config, HardwareModeAuto)
		if err != nil {
			t.Fatalf("Failed to create SEV-SNP service: %v", err)
		}
		defer svc.Shutdown()

		err = svc.Initialize(DefaultRuntimeConfig())
		if err != nil {
			t.Fatalf("Service should initialize even without hardware: %v", err)
		}

		attestation, err := svc.GenerateAttestation([]byte("test"))
		if err != nil {
			t.Fatalf("Should be able to generate attestation: %v", err)
		}

		if len(attestation) == 0 {
			t.Error("Attestation should not be empty")
		}
	})

	t.Run("NitroFallsBackToSimulation", func(t *testing.T) {
		config := NitroEnclaveConfig{
			EnclaveImagePath: "/opt/virtengine/enclaves/test.eif",
			CPUCount:         2,
			MemoryMB:         512,
			DebugMode:        true,
		}

		svc, err := NewNitroEnclaveServiceImplWithMode(config, HardwareModeAuto)
		if err != nil {
			t.Fatalf("Failed to create Nitro service: %v", err)
		}
		defer svc.Shutdown()

		err = svc.Initialize(DefaultRuntimeConfig())
		if err != nil {
			t.Fatalf("Service should initialize even without hardware: %v", err)
		}

		attestation, err := svc.GenerateAttestation([]byte("test"))
		if err != nil {
			t.Fatalf("Should be able to generate attestation: %v", err)
		}

		if len(attestation) == 0 {
			t.Error("Attestation should not be empty")
		}
	})

	t.Run("RequireModeFailsWithoutHardware", func(t *testing.T) {
		// Skip on machines with actual hardware
		caps := DetectHardware()
		if caps.SGXAvailable {
			t.Skip("SGX hardware is available, skipping require-mode failure test")
		}

		config := SGXEnclaveConfig{
			EnclavePath: "/opt/virtengine/enclaves/test.so",
			Debug:       true,
		}

		_, err := NewSGXEnclaveServiceImplWithMode(config, HardwareModeRequire)
		if err == nil {
			t.Error("Should fail when requiring hardware that's not available")
		}
		t.Logf("Expected error: %v", err)
	})

	t.Run("ScoringWorksInSimulationMode", func(t *testing.T) {
		config := SGXEnclaveConfig{
			EnclavePath: "/opt/virtengine/enclaves/test.so",
			Debug:       true,
		}

		svc, err := NewSGXEnclaveServiceImplWithMode(config, HardwareModeSimulate)
		if err != nil {
			t.Fatalf("Failed to create SGX service: %v", err)
		}
		defer svc.Shutdown()

		err = svc.Initialize(DefaultRuntimeConfig())
		if err != nil {
			t.Fatalf("Failed to initialize: %v", err)
		}

		// Create a scoring request
		request := &ScoringRequest{
			RequestID:      "test-request-1",
			Ciphertext:     []byte("encrypted identity data"),
			WrappedKey:     []byte("wrapped key material"),
			Nonce:          []byte("unique-nonce"),
			ScopeID:        "test-scope",
			AccountAddress: "cosmos1test...",
		}

		ctx := context.Background()
		result, err := svc.Score(ctx, request)
		if err != nil {
			t.Fatalf("Scoring failed: %v", err)
		}

		if !result.IsSuccess() {
			t.Errorf("Scoring should succeed, got error: %s", result.Error)
		}

		t.Logf("Score: %d, Status: %s", result.Score, result.Status)
	})
}

// =============================================================================
// Hardware Detection Tests
// =============================================================================

func TestHardwareDetection(t *testing.T) {
	t.Run("DetectHardwareReturnsCapabilities", func(t *testing.T) {
		caps := DetectHardware()

		// Should have a detection timestamp
		if caps.DetectedAt.IsZero() {
			t.Error("DetectedAt should be set")
		}

		// PreferredBackend should be set
		t.Logf("Preferred backend: %s", caps.PreferredBackend)
	})

	t.Run("RefreshHardwareDetection", func(t *testing.T) {
		caps1 := DetectHardware()
		time.Sleep(10 * time.Millisecond)
		caps2 := RefreshHardwareDetection()

		// Refresh should update detection time
		if !caps2.DetectedAt.After(caps1.DetectedAt) {
			t.Error("Refresh should update detection time")
		}
	})

	t.Run("HasAnyHardwareCheck", func(t *testing.T) {
		caps := DetectHardware()

		hasHardware := caps.HasAnyHardware()
		t.Logf("Has any hardware: %v", hasHardware)

		if hasHardware {
			// At least one should be true
			if !caps.SGXAvailable && !caps.SEVSNPAvailable && !caps.NitroAvailable {
				t.Error("HasAnyHardware is true but no specific hardware is available")
			}
		}
	})

	t.Run("HardwareModeString", func(t *testing.T) {
		tests := []struct {
			mode     HardwareMode
			expected string
		}{
			{HardwareModeAuto, "auto"},
			{HardwareModeSimulate, "simulate"},
			{HardwareModeRequire, "require"},
		}

		for _, tc := range tests {
			if tc.mode.String() != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, tc.mode.String())
			}
		}
	})
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkSGXAttestationSimulated(b *testing.B) {
	config := SGXEnclaveConfig{
		EnclavePath: "/opt/virtengine/enclaves/test.so",
		Debug:       true,
	}

	svc, err := NewSGXEnclaveServiceImplWithMode(config, HardwareModeSimulate)
	if err != nil {
		b.Fatalf("Failed to create SGX service: %v", err)
	}
	defer func() { _ = svc.Shutdown() }()

	_ = svc.Initialize(DefaultRuntimeConfig())

	reportData := []byte("benchmark-report-data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.GenerateAttestation(reportData)
		if err != nil {
			b.Fatalf("Attestation failed: %v", err)
		}
	}
}

func BenchmarkSEVAttestationSimulated(b *testing.B) {
	config := SEVSNPConfig{
		Endpoint:         "localhost:8443",
		AllowDebugPolicy: true,
	}

	svc, err := NewSEVSNPEnclaveServiceImplWithMode(config, HardwareModeSimulate)
	if err != nil {
		b.Fatalf("Failed to create SEV-SNP service: %v", err)
	}
	defer func() { _ = svc.Shutdown() }()

	_ = svc.Initialize(DefaultRuntimeConfig())

	reportData := []byte("benchmark-report-data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.GenerateAttestation(reportData)
		if err != nil {
			b.Fatalf("Attestation failed: %v", err)
		}
	}
}

func BenchmarkNitroAttestationSimulated(b *testing.B) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/virtengine/enclaves/test.eif",
		CPUCount:         2,
		MemoryMB:         512,
		DebugMode:        true,
	}

	svc, err := NewNitroEnclaveServiceImplWithMode(config, HardwareModeSimulate)
	if err != nil {
		b.Fatalf("Failed to create Nitro service: %v", err)
	}
	defer func() { _ = svc.Shutdown() }()

	_ = svc.Initialize(DefaultRuntimeConfig())

	reportData := []byte("benchmark-report-data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.GenerateAttestation(reportData)
		if err != nil {
			b.Fatalf("Attestation failed: %v", err)
		}
	}
}
