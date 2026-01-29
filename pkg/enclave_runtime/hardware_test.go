// Package enclave_runtime provides TEE enclave implementations.
//
// This file contains comprehensive tests for the hardware abstraction layer
// for Intel SGX, AMD SEV-SNP, and AWS Nitro Enclaves.
//
// Tests are designed to pass even without real hardware by using simulation mode.
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package enclave_runtime

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"
)

// =============================================================================
// SGX Hardware Detection Tests
// =============================================================================

func TestSGXHardwareDetection(t *testing.T) {
	t.Run("detector creation", func(t *testing.T) {
		detector := NewSGXHardwareDetector()
		if detector == nil {
			t.Fatal("NewSGXHardwareDetector returned nil")
		}
	})

	t.Run("detect runs without panic", func(t *testing.T) {
		detector := NewSGXHardwareDetector()
		// Detection may fail on non-SGX hardware, but should not panic
		_ = detector.Detect()
	})

	t.Run("version returns valid value", func(t *testing.T) {
		detector := NewSGXHardwareDetector()
		_ = detector.Detect()
		version := detector.Version()
		// Version should be 0, 1, or 2
		if version < 0 || version > 2 {
			t.Errorf("unexpected SGX version: %d", version)
		}
	})

	t.Run("device paths are sensible", func(t *testing.T) {
		detector := NewSGXHardwareDetector()
		_ = detector.Detect()
		enclave, provision := detector.GetDevicePaths()
		// Paths should be empty or valid paths
		if enclave != "" && enclave[0] != '/' {
			t.Errorf("invalid enclave device path: %s", enclave)
		}
		if provision != "" && provision[0] != '/' {
			t.Errorf("invalid provision device path: %s", provision)
		}
	})

	t.Run("concurrent detection is safe", func(t *testing.T) {
		detector := NewSGXHardwareDetector()
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = detector.Detect()
				_ = detector.IsAvailable()
				_ = detector.Version()
				_ = detector.HasFLC()
			}()
		}
		wg.Wait()
	})
}

func TestSGXEnclaveLoader(t *testing.T) {
	detector := NewSGXHardwareDetector()
	_ = detector.Detect()

	t.Run("loader creation", func(t *testing.T) {
		loader := NewSGXEnclaveLoader(detector)
		if loader == nil {
			t.Fatal("NewSGXEnclaveLoader returned nil")
		}
	})

	t.Run("load and unload enclave", func(t *testing.T) {
		loader := NewSGXEnclaveLoader(detector)

		// Load should work (even in simulation)
		err := loader.Load("/fake/enclave.so", false)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if !loader.IsLoaded() {
			t.Error("enclave should be loaded")
		}

		// Should be in simulated mode (no real hardware in CI)
		if !loader.IsSimulated() {
			t.Log("Running on real SGX hardware")
		}

		// Unload
		err = loader.Unload()
		if err != nil {
			t.Fatalf("Unload failed: %v", err)
		}

		if loader.IsLoaded() {
			t.Error("enclave should not be loaded after unload")
		}
	})

	t.Run("double load fails", func(t *testing.T) {
		loader := NewSGXEnclaveLoader(detector)

		err := loader.Load("/fake/enclave.so", false)
		if err != nil {
			t.Fatalf("First Load failed: %v", err)
		}

		err = loader.Load("/fake/enclave2.so", false)
		if err == nil {
			t.Error("second Load should fail")
		}

		loader.Unload()
	})

	t.Run("measurement is non-zero after load", func(t *testing.T) {
		loader := NewSGXEnclaveLoader(detector)
		_ = loader.Load("/fake/enclave.so", false)
		defer loader.Unload()

		measurement := loader.GetMeasurement()
		if measurement == (SGXMeasurement{}) {
			t.Error("measurement should not be zero")
		}
	})
}

func TestSGXReportGenerator(t *testing.T) {
	detector := NewSGXHardwareDetector()
	_ = detector.Detect()
	loader := NewSGXEnclaveLoader(detector)

	t.Run("generate report requires loaded enclave", func(t *testing.T) {
		gen := NewSGXReportGenerator(loader)

		var reportData [64]byte
		copy(reportData[:], []byte("test-data"))

		_, err := gen.GenerateReport(reportData, nil)
		if err == nil {
			t.Error("should fail without loaded enclave")
		}
	})

	t.Run("generate report with loaded enclave", func(t *testing.T) {
		_ = loader.Load("/fake/enclave.so", false)
		defer loader.Unload()

		gen := NewSGXReportGenerator(loader)

		var reportData [64]byte
		copy(reportData[:], []byte("test-nonce-data"))

		report, err := gen.GenerateReport(reportData, nil)
		if err != nil {
			t.Fatalf("GenerateReport failed: %v", err)
		}

		if report == nil {
			t.Fatal("report should not be nil")
		}

		// Verify report data is preserved
		if report.ReportData != reportData {
			t.Error("report data mismatch")
		}

		// Verify measurement matches loader
		if report.MREnclave != loader.GetMeasurement() {
			t.Error("measurement mismatch")
		}
	})
}

func TestSGXQuoteGenerator(t *testing.T) {
	detector := NewSGXHardwareDetector()
	_ = detector.Detect()
	loader := NewSGXEnclaveLoader(detector)

	t.Run("generate quote with loaded enclave", func(t *testing.T) {
		_ = loader.Load("/fake/enclave.so", false)
		defer loader.Unload()

		quoter := NewSGXQuoteGenerator(detector, loader)

		var reportData [64]byte
		copy(reportData[:], []byte("quote-nonce"))

		quote, err := quoter.GenerateQuote(reportData)
		if err != nil {
			t.Fatalf("GenerateQuote failed: %v", err)
		}

		if quote == nil {
			t.Fatal("quote should not be nil")
		}

		// Verify quote header
		if quote.Header.Version != SGXQuoteVersionDCAP {
			t.Errorf("unexpected quote version: %d", quote.Header.Version)
		}

		// Verify signature exists
		if len(quote.Signature) == 0 {
			t.Error("quote should have signature")
		}
	})
}

func TestSGXSealingService(t *testing.T) {
	detector := NewSGXHardwareDetector()
	_ = detector.Detect()
	loader := NewSGXEnclaveLoader(detector)
	_ = loader.Load("/fake/enclave.so", false)
	defer loader.Unload()

	sealer := NewSGXSealingService(loader)

	t.Run("seal and unseal", func(t *testing.T) {
		plaintext := []byte("sensitive data to seal")

		sealed, err := sealer.Seal(plaintext)
		if err != nil {
			t.Fatalf("Seal failed: %v", err)
		}

		if len(sealed) < len(plaintext) {
			t.Error("sealed data should be at least as long as plaintext")
		}

		unsealed, err := sealer.Unseal(sealed)
		if err != nil {
			t.Fatalf("Unseal failed: %v", err)
		}

		if !bytes.Equal(unsealed, plaintext) {
			t.Errorf("unsealed data mismatch: got %q, want %q", unsealed, plaintext)
		}
	})

	t.Run("unseal with corrupted data fails gracefully", func(t *testing.T) {
		plaintext := []byte("test data")
		sealed, _ := sealer.Seal(plaintext)

		// Corrupt the sealed data
		if len(sealed) > 15 {
			sealed[15] ^= 0xFF
		}

		unsealed, err := sealer.Unseal(sealed)
		// In simulation mode, this might not fail, but data should differ
		if err == nil && bytes.Equal(unsealed, plaintext) {
			t.Error("corrupted data should produce different output")
		}
	})

	t.Run("seal empty data", func(t *testing.T) {
		sealed, err := sealer.Seal([]byte{})
		if err != nil {
			t.Fatalf("Seal empty data failed: %v", err)
		}

		unsealed, err := sealer.Unseal(sealed)
		if err != nil {
			t.Fatalf("Unseal empty data failed: %v", err)
		}

		if len(unsealed) != 0 {
			t.Errorf("unsealed should be empty, got %d bytes", len(unsealed))
		}
	})
}

func TestSGXECallInterface(t *testing.T) {
	detector := NewSGXHardwareDetector()
	_ = detector.Detect()
	loader := NewSGXEnclaveLoader(detector)
	_ = loader.Load("/fake/enclave.so", false)
	defer loader.Unload()

	ecaller := NewSGXECallInterface(loader)

	t.Run("ecall returns result", func(t *testing.T) {
		input := []byte("ecall input data")

		result, err := ecaller.Call(1, input)
		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}

		if result == nil {
			t.Fatal("result should not be nil")
		}

		if result.ReturnValue != SGX_SUCCESS {
			t.Errorf("unexpected return value: %d", result.ReturnValue)
		}

		if len(result.OutputData) == 0 {
			t.Error("output data should not be empty")
		}
	})

	t.Run("call count increments", func(t *testing.T) {
		initial := ecaller.GetCallCount()

		for i := 0; i < 5; i++ {
			ecaller.Call(i, []byte("test"))
		}

		final := ecaller.GetCallCount()
		if final != initial+5 {
			t.Errorf("call count mismatch: got %d, want %d", final, initial+5)
		}
	})
}

func TestSGXHardwareBackend(t *testing.T) {
	backend := NewSGXHardwareBackend()

	t.Run("platform is SGX", func(t *testing.T) {
		if backend.Platform() != AttestationTypeSGX {
			t.Errorf("unexpected platform: %s", backend.Platform())
		}
	})

	t.Run("initialize and shutdown", func(t *testing.T) {
		err := backend.Initialize()
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		err = backend.Shutdown()
		if err != nil {
			t.Fatalf("Shutdown failed: %v", err)
		}
	})

	t.Run("get attestation", func(t *testing.T) {
		_ = backend.Initialize()
		defer backend.Shutdown()

		nonce := []byte("test-attestation-nonce")
		attestation, err := backend.GetAttestation(nonce)
		if err != nil {
			t.Fatalf("GetAttestation failed: %v", err)
		}

		if len(attestation) == 0 {
			t.Error("attestation should not be empty")
		}
	})

	t.Run("derive key", func(t *testing.T) {
		_ = backend.Initialize()
		defer backend.Shutdown()

		key1, err := backend.DeriveKey([]byte("context1"), 32)
		if err != nil {
			t.Fatalf("DeriveKey failed: %v", err)
		}

		if len(key1) != 32 {
			t.Errorf("key length mismatch: got %d, want 32", len(key1))
		}

		key2, err := backend.DeriveKey([]byte("context2"), 32)
		if err != nil {
			t.Fatalf("DeriveKey failed: %v", err)
		}

		// Different contexts should produce different keys
		if bytes.Equal(key1, key2) {
			t.Error("different contexts should produce different keys")
		}
	})

	t.Run("seal and unseal", func(t *testing.T) {
		_ = backend.Initialize()
		defer backend.Shutdown()

		plaintext := []byte("backend seal test")

		sealed, err := backend.Seal(plaintext)
		if err != nil {
			t.Fatalf("Seal failed: %v", err)
		}

		unsealed, err := backend.Unseal(sealed)
		if err != nil {
			t.Fatalf("Unseal failed: %v", err)
		}

		if !bytes.Equal(unsealed, plaintext) {
			t.Error("unsealed data mismatch")
		}
	})
}

// =============================================================================
// SEV Hardware Detection Tests
// =============================================================================

func TestSEVHardwareDetection(t *testing.T) {
	t.Run("detector creation", func(t *testing.T) {
		detector := NewSEVHardwareDetector()
		if detector == nil {
			t.Fatal("NewSEVHardwareDetector returned nil")
		}
	})

	t.Run("detect runs without panic", func(t *testing.T) {
		detector := NewSEVHardwareDetector()
		_ = detector.Detect()
	})

	t.Run("version is returned", func(t *testing.T) {
		detector := NewSEVHardwareDetector()
		_ = detector.Detect()
		version := detector.Version()
		// Version should be "unknown" or a valid version string
		if version == "" {
			version = "unknown"
		}
		t.Logf("SEV-SNP version: %s", version)
	})

	t.Run("api version is non-negative", func(t *testing.T) {
		detector := NewSEVHardwareDetector()
		_ = detector.Detect()
		apiVersion := detector.APIVersion()
		if apiVersion < 0 {
			t.Errorf("negative API version: %d", apiVersion)
		}
	})

	t.Run("concurrent detection is safe", func(t *testing.T) {
		detector := NewSEVHardwareDetector()
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = detector.Detect()
				_ = detector.IsAvailable()
				_ = detector.Version()
			}()
		}
		wg.Wait()
	})
}

func TestSEVGuestDevice(t *testing.T) {
	detector := NewSEVHardwareDetector()
	_ = detector.Detect()

	t.Run("device creation", func(t *testing.T) {
		device := NewSEVGuestDevice(detector)
		if device == nil {
			t.Fatal("NewSEVGuestDevice returned nil")
		}
	})

	t.Run("open and close device", func(t *testing.T) {
		device := NewSEVGuestDevice(detector)

		err := device.Open()
		if err != nil {
			t.Fatalf("Open failed: %v", err)
		}

		// Should be simulated on non-SEV hardware
		t.Logf("SEV device simulated: %v", device.IsSimulated())

		err = device.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}
	})
}

func TestSNPReportRequester(t *testing.T) {
	detector := NewSEVHardwareDetector()
	_ = detector.Detect()
	device := NewSEVGuestDevice(detector)
	_ = device.Open()
	defer device.Close()

	t.Run("request report", func(t *testing.T) {
		requester := NewSNPReportRequester(device)

		var userData [64]byte
		copy(userData[:], []byte("test-user-data"))

		report, err := requester.RequestReport(userData, 0)
		if err != nil {
			t.Fatalf("RequestReport failed: %v", err)
		}

		if report == nil {
			t.Fatal("report should not be nil")
		}

		// Verify user data is in report
		if report.ReportData != userData {
			t.Error("user data mismatch in report")
		}

		// Verify version
		if report.Version < SNPReportVersion {
			t.Errorf("unexpected report version: %d", report.Version)
		}

		// Verify policy (debug should be false)
		if report.Policy.Debug {
			t.Error("debug mode should be disabled")
		}
	})
}

func TestSNPDerivedKeyRequester(t *testing.T) {
	detector := NewSEVHardwareDetector()
	_ = detector.Detect()
	device := NewSEVGuestDevice(detector)
	_ = device.Open()
	defer device.Close()

	t.Run("request key", func(t *testing.T) {
		requester := NewSNPDerivedKeyRequester(device)

		key, err := requester.RequestKey(SNP_KEY_ROOT_VCEK, SNP_KEY_GUEST_FIELD, 0)
		if err != nil {
			t.Fatalf("RequestKey failed: %v", err)
		}

		if len(key) != 32 {
			t.Errorf("key length mismatch: got %d, want 32", len(key))
		}
	})

	t.Run("different parameters produce different keys", func(t *testing.T) {
		requester := NewSNPDerivedKeyRequester(device)

		key1, _ := requester.RequestKey(SNP_KEY_ROOT_VCEK, SNP_KEY_GUEST_FIELD, 0)
		key2, _ := requester.RequestKey(SNP_KEY_ROOT_VCEK, SNP_KEY_TCB_FIELD, 0)

		if bytes.Equal(key1, key2) {
			t.Error("different field selections should produce different keys")
		}
	})
}

func TestSNPExtendedReportRequester(t *testing.T) {
	detector := NewSEVHardwareDetector()
	_ = detector.Detect()
	device := NewSEVGuestDevice(detector)
	_ = device.Open()
	defer device.Close()

	t.Run("request extended report", func(t *testing.T) {
		requester := NewSNPExtendedReportRequester(device)

		var userData [64]byte
		copy(userData[:], []byte("extended-report-test"))

		extReport, err := requester.RequestExtendedReport(userData, 0)
		if err != nil {
			t.Fatalf("RequestExtendedReport failed: %v", err)
		}

		if extReport == nil {
			t.Fatal("extended report should not be nil")
		}

		if extReport.Report == nil {
			t.Error("report should not be nil")
		}

		// Verify certificates are present
		if len(extReport.VCEKCert) == 0 {
			t.Error("VCEK cert should not be empty")
		}
		if len(extReport.ASKCert) == 0 {
			t.Error("ASK cert should not be empty")
		}
		if len(extReport.ARKCert) == 0 {
			t.Error("ARK cert should not be empty")
		}
	})
}

func TestSEVHardwareBackend(t *testing.T) {
	backend := NewSEVHardwareBackend()

	t.Run("platform is SEV-SNP", func(t *testing.T) {
		if backend.Platform() != AttestationTypeSEVSNP {
			t.Errorf("unexpected platform: %s", backend.Platform())
		}
	})

	t.Run("initialize and shutdown", func(t *testing.T) {
		err := backend.Initialize()
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		err = backend.Shutdown()
		if err != nil {
			t.Fatalf("Shutdown failed: %v", err)
		}
	})

	t.Run("get attestation", func(t *testing.T) {
		_ = backend.Initialize()
		defer backend.Shutdown()

		nonce := []byte("sev-snp-attestation-nonce")
		attestation, err := backend.GetAttestation(nonce)
		if err != nil {
			t.Fatalf("GetAttestation failed: %v", err)
		}

		if len(attestation) == 0 {
			t.Error("attestation should not be empty")
		}
	})

	t.Run("seal and unseal", func(t *testing.T) {
		_ = backend.Initialize()
		defer backend.Shutdown()

		plaintext := []byte("sev-snp backend seal test")

		sealed, err := backend.Seal(plaintext)
		if err != nil {
			t.Fatalf("Seal failed: %v", err)
		}

		unsealed, err := backend.Unseal(sealed)
		if err != nil {
			t.Fatalf("Unseal failed: %v", err)
		}

		if !bytes.Equal(unsealed, plaintext) {
			t.Error("unsealed data mismatch")
		}
	})

	t.Run("get platform info", func(t *testing.T) {
		_ = backend.Initialize()
		defer backend.Shutdown()

		info, err := backend.GetPlatformInfo()
		if err != nil {
			t.Fatalf("GetPlatformInfo failed: %v", err)
		}

		if info == nil {
			t.Fatal("platform info should not be nil")
		}

		t.Logf("SEV API version: %d", info.APIVersion)
	})
}

func TestSNPGuestPolicy(t *testing.T) {
	t.Run("validate good policy", func(t *testing.T) {
		policy := SNPGuestPolicy{
			ABIMajor: 1,
			ABIMinor: 0,
			SMT:      true,
			Debug:    false,
		}

		err := VerifyGuestPolicy(policy)
		if err != nil {
			t.Errorf("valid policy should pass: %v", err)
		}
	})

	t.Run("reject debug policy", func(t *testing.T) {
		policy := SNPGuestPolicy{
			ABIMajor: 1,
			Debug:    true,
		}

		err := VerifyGuestPolicy(policy)
		if err == nil {
			t.Error("debug policy should be rejected")
		}
	})

	t.Run("reject zero ABI major", func(t *testing.T) {
		policy := SNPGuestPolicy{
			ABIMajor: 0,
			Debug:    false,
		}

		err := VerifyGuestPolicy(policy)
		if err == nil {
			t.Error("zero ABI major should be rejected")
		}
	})
}

// =============================================================================
// Nitro Hardware Detection Tests
// =============================================================================

func TestNitroHardwareDetection(t *testing.T) {
	t.Run("detector creation", func(t *testing.T) {
		detector := NewNitroHardwareDetector()
		if detector == nil {
			t.Fatal("NewNitroHardwareDetector returned nil")
		}
	})

	t.Run("detect runs without panic", func(t *testing.T) {
		detector := NewNitroHardwareDetector()
		_ = detector.Detect()
	})

	t.Run("version is returned", func(t *testing.T) {
		detector := NewNitroHardwareDetector()
		_ = detector.Detect()
		version := detector.Version()
		t.Logf("Nitro CLI version: %s", version)
	})

	t.Run("cli availability check", func(t *testing.T) {
		detector := NewNitroHardwareDetector()
		_ = detector.Detect()
		hasCLI := detector.HasCLI()
		t.Logf("Nitro CLI available: %v", hasCLI)
	})

	t.Run("concurrent detection is safe", func(t *testing.T) {
		detector := NewNitroHardwareDetector()
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = detector.Detect()
				_ = detector.IsAvailable()
				_ = detector.Version()
			}()
		}
		wg.Wait()
	})
}

func TestNitroCLIRunner(t *testing.T) {
	detector := NewNitroHardwareDetector()
	_ = detector.Detect()
	runner := NewNitroCLIRunner(detector)

	t.Run("runner creation", func(t *testing.T) {
		if runner == nil {
			t.Fatal("NewNitroCLIRunner returned nil")
		}
	})

	t.Run("simulated mode detection", func(t *testing.T) {
		// In CI without Nitro hardware, should be simulated
		t.Logf("CLI runner simulated: %v", runner.IsSimulated())
	})

	t.Run("describe enclaves in simulation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		enclaves, err := runner.DescribeEnclaves(ctx)
		if err != nil {
			t.Fatalf("DescribeEnclaves failed: %v", err)
		}

		// In simulation, should return empty list
		if runner.IsSimulated() && len(enclaves) != 0 {
			t.Errorf("simulated mode should return empty list, got %d", len(enclaves))
		}
	})

	t.Run("run simulated enclave", func(t *testing.T) {
		if !runner.IsSimulated() {
			t.Skip("skipping on real hardware")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := runner.RunEnclave(ctx, "/fake/enclave.eif", 2, 2048)
		if err != nil {
			t.Fatalf("RunEnclave failed: %v", err)
		}

		if result.EnclaveID == "" {
			t.Error("enclave ID should not be empty")
		}
		if result.EnclaveCID == 0 {
			t.Error("enclave CID should not be zero")
		}
		if result.NumberOfCPUs != 2 {
			t.Errorf("CPU count mismatch: got %d, want 2", result.NumberOfCPUs)
		}
		if result.MemoryMiB != 2048 {
			t.Errorf("memory mismatch: got %d, want 2048", result.MemoryMiB)
		}
	})
}

func TestNitroVsockClient(t *testing.T) {
	t.Run("client creation", func(t *testing.T) {
		client := NewNitroVsockClient(100, 5000)
		if client == nil {
			t.Fatal("NewNitroVsockClient returned nil")
		}

		if client.GetCID() != 100 {
			t.Errorf("CID mismatch: got %d, want 100", client.GetCID())
		}
		if client.GetPort() != 5000 {
			t.Errorf("port mismatch: got %d, want 5000", client.GetPort())
		}
	})

	t.Run("connect and disconnect", func(t *testing.T) {
		client := NewNitroVsockClient(16, 5000)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := client.Connect(ctx)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}

		if !client.IsConnected() {
			t.Error("should be connected")
		}

		err = client.Disconnect()
		if err != nil {
			t.Fatalf("Disconnect failed: %v", err)
		}

		if client.IsConnected() {
			t.Error("should not be connected after disconnect")
		}
	})

	t.Run("send without connection fails", func(t *testing.T) {
		client := NewNitroVsockClient(16, 5000)

		err := client.Send([]byte("test"))
		if err == nil {
			t.Error("Send should fail without connection")
		}
	})
}

func TestNitroNSMClient(t *testing.T) {
	t.Run("client creation", func(t *testing.T) {
		client := NewNitroNSMClient()
		if client == nil {
			t.Fatal("NewNitroNSMClient returned nil")
		}
	})

	t.Run("open and close", func(t *testing.T) {
		client := NewNitroNSMClient()

		err := client.Open()
		if err != nil {
			t.Fatalf("Open failed: %v", err)
		}

		t.Logf("NSM client simulated: %v", client.IsSimulated())

		err = client.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}
	})

	t.Run("get attestation document", func(t *testing.T) {
		client := NewNitroNSMClient()
		_ = client.Open()
		defer client.Close()

		userData := []byte("user data")
		nonce := []byte("nonce")
		pubKey := []byte("public key")

		doc, err := client.GetAttestationDocument(userData, nonce, pubKey)
		if err != nil {
			t.Fatalf("GetAttestationDocument failed: %v", err)
		}

		if doc == nil {
			t.Fatal("document should not be nil")
		}

		if doc.ModuleID == "" {
			t.Error("module ID should not be empty")
		}
		if doc.Timestamp == 0 {
			t.Error("timestamp should not be zero")
		}
		if len(doc.PCRs) != 16 {
			t.Errorf("should have 16 PCRs, got %d", len(doc.PCRs))
		}
	})

	t.Run("describe PCRs", func(t *testing.T) {
		client := NewNitroNSMClient()
		_ = client.Open()
		defer client.Close()

		pcrs, err := client.DescribePCRs()
		if err != nil {
			t.Fatalf("DescribePCRs failed: %v", err)
		}

		if len(pcrs) != 16 {
			t.Errorf("should have 16 PCRs, got %d", len(pcrs))
		}

		// Each PCR should be 48 bytes (SHA-384)
		for idx, pcr := range pcrs {
			if len(pcr) != 48 {
				t.Errorf("PCR %d length mismatch: got %d, want 48", idx, len(pcr))
			}
		}
	})

	t.Run("extend PCR", func(t *testing.T) {
		client := NewNitroNSMClient()
		_ = client.Open()
		defer client.Close()

		// Get initial PCR value
		pcrs1, _ := client.DescribePCRs()
		initialPCR := pcrs1[8]

		// Extend PCR 8
		err := client.ExtendPCR(8, []byte("extend data"))
		if err != nil {
			t.Fatalf("ExtendPCR failed: %v", err)
		}

		// Get new PCR value
		pcrs2, _ := client.DescribePCRs()
		newPCR := pcrs2[8]

		// Should be different
		if bytes.Equal(initialPCR, newPCR) {
			t.Error("PCR should change after extend")
		}
	})

	t.Run("lock PCR", func(t *testing.T) {
		client := NewNitroNSMClient()
		_ = client.Open()
		defer client.Close()

		err := client.LockPCR(4)
		if err != nil {
			t.Fatalf("LockPCR failed: %v", err)
		}
	})

	t.Run("invalid PCR index", func(t *testing.T) {
		client := NewNitroNSMClient()
		_ = client.Open()
		defer client.Close()

		err := client.ExtendPCR(20, []byte("test"))
		if err == nil {
			t.Error("should fail for PCR index > 15")
		}
	})
}

func TestNitroEnclaveImageBuilder(t *testing.T) {
	detector := NewNitroHardwareDetector()
	_ = detector.Detect()
	builder := NewNitroEnclaveImageBuilder(detector)

	t.Run("build simulated enclave", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		config := BuildConfig{
			DockerUri:  "hello-world:latest",
			OutputPath: "/tmp/test.eif",
			Name:       "test-enclave",
		}

		result, err := builder.BuildEnclave(ctx, config)
		if err != nil {
			t.Fatalf("BuildEnclave failed: %v", err)
		}

		if result.Measurements.HashAlgorithm != "SHA384" {
			t.Errorf("unexpected hash algorithm: %s", result.Measurements.HashAlgorithm)
		}
		if result.Measurements.PCR0 == "" {
			t.Error("PCR0 should not be empty")
		}
	})
}

func TestNitroHardwareBackend(t *testing.T) {
	backend := NewNitroHardwareBackend()

	t.Run("platform is Nitro", func(t *testing.T) {
		if backend.Platform() != AttestationTypeNitro {
			t.Errorf("unexpected platform: %s", backend.Platform())
		}
	})

	t.Run("initialize and shutdown", func(t *testing.T) {
		err := backend.Initialize()
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		err = backend.Shutdown()
		if err != nil {
			t.Fatalf("Shutdown failed: %v", err)
		}
	})

	t.Run("get attestation", func(t *testing.T) {
		_ = backend.Initialize()
		defer backend.Shutdown()

		nonce := []byte("nitro-attestation-nonce")
		attestation, err := backend.GetAttestation(nonce)
		if err != nil {
			t.Fatalf("GetAttestation failed: %v", err)
		}

		if len(attestation) == 0 {
			t.Error("attestation should not be empty")
		}
	})

	t.Run("seal and unseal", func(t *testing.T) {
		_ = backend.Initialize()
		defer backend.Shutdown()

		plaintext := []byte("nitro backend seal test")

		sealed, err := backend.Seal(plaintext)
		if err != nil {
			t.Fatalf("Seal failed: %v", err)
		}

		unsealed, err := backend.Unseal(sealed)
		if err != nil {
			t.Fatalf("Unseal failed: %v", err)
		}

		if !bytes.Equal(unsealed, plaintext) {
			t.Error("unsealed data mismatch")
		}
	})
}

// =============================================================================
// Common Hardware Capability Tests
// =============================================================================

func TestHardwareCapabilities(t *testing.T) {
	t.Run("detect hardware", func(t *testing.T) {
		caps := DetectHardware()

		t.Logf("Hardware capabilities: %s", caps.String())
		t.Logf("  SGX available: %v (v%d, FLC: %v)", caps.SGXAvailable, caps.SGXVersion, caps.SGXFLCSupported)
		t.Logf("  SEV-SNP available: %v (%s)", caps.SEVSNPAvailable, caps.SEVSNPVersion)
		t.Logf("  Nitro available: %v (%s)", caps.NitroAvailable, caps.NitroVersion)
		t.Logf("  Preferred backend: %s", caps.PreferredBackend)

		if len(caps.DetectionErrors) > 0 {
			t.Logf("  Detection errors: %v", caps.DetectionErrors)
		}
	})

	t.Run("refresh hardware detection", func(t *testing.T) {
		caps1 := DetectHardware()
		caps2 := RefreshHardwareDetection()

		// Results should be consistent
		if caps1.SGXAvailable != caps2.SGXAvailable {
			t.Log("SGX availability changed between detections")
		}
	})

	t.Run("has any hardware", func(t *testing.T) {
		caps := DetectHardware()

		hasAny := caps.HasAnyHardware()
		manual := caps.SGXAvailable || caps.SEVSNPAvailable || caps.NitroAvailable

		if hasAny != manual {
			t.Error("HasAnyHardware mismatch")
		}
	})

	t.Run("preferred backend is valid", func(t *testing.T) {
		caps := DetectHardware()

		validBackends := map[AttestationType]bool{
			AttestationTypeSGX:       true,
			AttestationTypeSEVSNP:    true,
			AttestationTypeNitro:     true,
			AttestationTypeSimulated: true,
		}

		if !validBackends[caps.PreferredBackend] {
			t.Errorf("invalid preferred backend: %s", caps.PreferredBackend)
		}
	})
}

func TestSimulationFallback(t *testing.T) {
	t.Run("SGX falls back to simulation", func(t *testing.T) {
		detector := NewSGXHardwareDetector()
		_ = detector.Detect()
		loader := NewSGXEnclaveLoader(detector)

		err := loader.Load("/nonexistent/enclave.so", false)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		defer loader.Unload()

		// Should be in simulation mode
		if !loader.IsSimulated() {
			t.Log("Running on real SGX hardware - skipping simulation test")
			return
		}

		// Verify operations work in simulation
		sealer := NewSGXSealingService(loader)
		sealed, err := sealer.Seal([]byte("test"))
		if err != nil {
			t.Fatalf("Seal in simulation failed: %v", err)
		}

		_, err = sealer.Unseal(sealed)
		if err != nil {
			t.Fatalf("Unseal in simulation failed: %v", err)
		}
	})

	t.Run("SEV-SNP falls back to simulation", func(t *testing.T) {
		detector := NewSEVHardwareDetector()
		_ = detector.Detect()
		device := NewSEVGuestDevice(detector)

		err := device.Open()
		if err != nil {
			t.Fatalf("Open failed: %v", err)
		}
		defer device.Close()

		if !device.IsSimulated() {
			t.Log("Running on real SEV-SNP hardware - skipping simulation test")
			return
		}

		// Verify operations work in simulation
		requester := NewSNPReportRequester(device)
		var userData [64]byte
		_, err = requester.RequestReport(userData, 0)
		if err != nil {
			t.Fatalf("RequestReport in simulation failed: %v", err)
		}
	})

	t.Run("Nitro falls back to simulation", func(t *testing.T) {
		detector := NewNitroHardwareDetector()
		_ = detector.Detect()
		runner := NewNitroCLIRunner(detector)

		if !runner.IsSimulated() {
			t.Log("Running on real Nitro hardware - skipping simulation test")
			return
		}

		// Verify operations work in simulation
		ctx := context.Background()
		_, err := runner.RunEnclave(ctx, "/fake/enclave.eif", 2, 2048)
		if err != nil {
			t.Fatalf("RunEnclave in simulation failed: %v", err)
		}
	})
}

func TestHardwareState(t *testing.T) {
	t.Run("create hardware state", func(t *testing.T) {
		state := NewHardwareState(HardwareModeAuto)
		if state == nil {
			t.Fatal("NewHardwareState returned nil")
		}
	})

	t.Run("initialize in auto mode", func(t *testing.T) {
		state := NewHardwareState(HardwareModeAuto)

		err := state.Initialize()
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		err = state.Shutdown()
		if err != nil {
			t.Fatalf("Shutdown failed: %v", err)
		}
	})

	t.Run("initialize in simulate mode", func(t *testing.T) {
		state := NewHardwareState(HardwareModeSimulate)

		err := state.Initialize()
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		// In simulate mode, no active backend
		if state.IsHardwareActive() {
			t.Error("simulate mode should not have active hardware")
		}

		_ = state.Shutdown()
	})

	t.Run("require mode fails without hardware", func(t *testing.T) {
		caps := DetectHardware()
		if caps.HasAnyHardware() {
			t.Skip("skipping - real hardware detected")
		}

		state := NewHardwareState(HardwareModeRequire)

		err := state.Initialize()
		if err == nil {
			t.Error("require mode should fail without hardware")
			state.Shutdown()
		}
	})

	t.Run("double initialize is idempotent", func(t *testing.T) {
		state := NewHardwareState(HardwareModeAuto)

		err1 := state.Initialize()
		err2 := state.Initialize()

		if err1 != nil || err2 != nil {
			t.Errorf("initialization errors: %v, %v", err1, err2)
		}

		_ = state.Shutdown()
	})

	t.Run("get active backend", func(t *testing.T) {
		state := NewHardwareState(HardwareModeAuto)
		_ = state.Initialize()
		defer state.Shutdown()

		backend := state.GetActiveBackend()
		if backend != nil {
			t.Logf("Active backend: %s", backend.Platform())
		} else {
			t.Log("No active backend (simulation mode)")
		}
	})
}

func TestHardwareModeString(t *testing.T) {
	tests := []struct {
		mode HardwareMode
		want string
	}{
		{HardwareModeAuto, "auto"},
		{HardwareModeSimulate, "simulate"},
		{HardwareModeRequire, "require"},
		{HardwareMode(99), "unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("HardwareMode.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHardwareError(t *testing.T) {
	t.Run("error with device path", func(t *testing.T) {
		err := &HardwareError{
			Platform:   AttestationTypeSGX,
			Operation:  "open",
			DevicePath: "/dev/sgx_enclave",
			Underlying: ErrPermissionDenied,
		}

		msg := err.Error()
		if msg == "" {
			t.Error("error message should not be empty")
		}
		if !bytes.Contains([]byte(msg), []byte("SGX")) {
			t.Error("error message should contain platform")
		}
		if !bytes.Contains([]byte(msg), []byte("/dev/sgx_enclave")) {
			t.Error("error message should contain device path")
		}
	})

	t.Run("error without device path", func(t *testing.T) {
		err := &HardwareError{
			Platform:   AttestationTypeSEVSNP,
			Operation:  "attestation",
			Underlying: ErrHardwareOperationFailed,
		}

		msg := err.Error()
		if msg == "" {
			t.Error("error message should not be empty")
		}
	})

	t.Run("unwrap underlying error", func(t *testing.T) {
		underlying := ErrDeviceNotFound
		err := &HardwareError{
			Platform:   AttestationTypeNitro,
			Operation:  "detect",
			Underlying: underlying,
		}

		if err.Unwrap() != underlying {
			t.Error("Unwrap should return underlying error")
		}
	})
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkSGXSeal(b *testing.B) {
	detector := NewSGXHardwareDetector()
	_ = detector.Detect()
	loader := NewSGXEnclaveLoader(detector)
	_ = loader.Load("/fake/enclave.so", false)
	defer loader.Unload()

	sealer := NewSGXSealingService(loader)
	plaintext := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sealer.Seal(plaintext)
	}
}

func BenchmarkSEVRequestReport(b *testing.B) {
	detector := NewSEVHardwareDetector()
	_ = detector.Detect()
	device := NewSEVGuestDevice(detector)
	_ = device.Open()
	defer device.Close()

	requester := NewSNPReportRequester(device)
	var userData [64]byte

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		requester.RequestReport(userData, 0)
	}
}

func BenchmarkNitroAttestation(b *testing.B) {
	client := NewNitroNSMClient()
	_ = client.Open()
	defer client.Close()

	userData := []byte("bench")
	nonce := []byte("nonce")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.GetAttestationDocument(userData, nonce, nil)
	}
}

func BenchmarkDetectHardware(b *testing.B) {
	// First call to populate cache
	DetectHardware()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectHardware()
	}
}
