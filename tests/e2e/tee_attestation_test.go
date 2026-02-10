//go:build e2e.integration

// Package e2e contains end-to-end tests for VirtEngine.
//
// This file implements comprehensive E2E tests for TEE (Trusted Execution Environment)
// attestation flows including:
// - Hardware detection and initialization
// - SGX quote generation and verification
// - SEV-SNP report generation and verification
// - Nitro attestation document handling
// - Remote attestation protocol
// - Cross-platform attestation verification
//
// Task Reference: 26E - TEE Hardware Integration and Attestation Testing
package e2e

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/pkg/enclave_runtime"
	"github.com/virtengine/virtengine/pkg/enclave_runtime/hardware"
	"github.com/virtengine/virtengine/pkg/enclave_runtime/nitro"
	"github.com/virtengine/virtengine/pkg/enclave_runtime/sev"
	"github.com/virtengine/virtengine/pkg/enclave_runtime/sgx"
)

// ============================================================================
// TEE Attestation E2E Test Suite
// ============================================================================

// TEEAttestationE2ETestSuite tests the complete TEE attestation flows.
//
// Test Coverage:
//  1. Hardware detection on all platforms
//  2. SGX quote generation and DCAP verification
//  3. SEV-SNP report generation and KDS verification
//  4. Nitro attestation document verification
//  5. Remote attestation protocol (validator-to-validator)
//  6. Measurement allowlist validation
//  7. Error handling for invalid attestations
type TEEAttestationE2ETestSuite struct {
	suite.Suite

	hardwareManager *hardware.HardwareManager
	ctx             context.Context
	cancel          context.CancelFunc
}

// TestTEEAttestationE2E runs the TEE attestation E2E test suite.
func TestTEEAttestationE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	suite.Run(t, new(TEEAttestationE2ETestSuite))
}

// SetupSuite runs once before all tests in the suite.
func (s *TEEAttestationE2ETestSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 5*time.Minute)

	// Create hardware manager with simulation mode
	config := hardware.DefaultConfig()
	config.AllowSimulation = true
	config.RequireHardware = false
	config.HealthCheckInterval = 30 * time.Second
	manager, err := hardware.NewHardwareManager(config)
	s.Require().NoError(err)

	s.hardwareManager = manager
	s.Require().NoError(s.hardwareManager.Initialize(s.ctx))
}

// TearDownSuite runs once after all tests in the suite.
func (s *TEEAttestationE2ETestSuite) TearDownSuite() {
	s.cancel()
	if s.hardwareManager != nil {
		_ = s.hardwareManager.Shutdown()
	}
}

// ============================================================================
// Test: Hardware Detection
// ============================================================================

// TestHardwareDetection tests unified hardware detection across platforms.
func (s *TEEAttestationE2ETestSuite) TestHardwareDetection() {
	t := s.T()

	// Get unified detector
	detector := hardware.NewUnifiedDetector()
	require.NotNil(t, detector)

	// Perform detection
	caps, err := detector.Detect()
	require.NoError(t, err)

	// Verify capabilities structure
	require.NotNil(t, caps)

	// Log detected capabilities
	t.Logf("Hardware detection results:")
	t.Logf("  SGX Available: %v (version: %d, FLC: %v)",
		caps.SGX.Available, caps.SGX.Version, caps.SGX.FLCSupported)
	t.Logf("  SEV-SNP Available: %v (version: %s)",
		caps.SEVSNP.Available, caps.SEVSNP.Version)
	t.Logf("  Nitro Available: %v (version: %s)",
		caps.Nitro.Available, caps.Nitro.Version)
	t.Logf("  Preferred Platform: %s", caps.Platform)

	// On non-Linux or non-TEE hardware, should fall back to simulated
	if caps.Platform == hardware.PlatformSimulated {
		t.Log("No hardware TEE detected, simulation mode active")
	}
}

// TestHardwareDetectionCaching tests that detection results are properly cached.
func (s *TEEAttestationE2ETestSuite) TestHardwareDetectionCaching() {
	t := s.T()

	detector := hardware.NewUnifiedDetector()

	// First detection
	caps1, err := detector.Detect()
	require.NoError(t, err)

	// Second detection should use cache
	caps2, err := detector.Detect()
	require.NoError(t, err)

	// Results should be identical
	require.Equal(t, caps1.Platform, caps2.Platform)
	require.Equal(t, caps1.SGX.Available, caps2.SGX.Available)
	require.Equal(t, caps1.SEVSNP.Available, caps2.SEVSNP.Available)
	require.Equal(t, caps1.Nitro.Available, caps2.Nitro.Available)
}

// ============================================================================
// Test: SGX Quote Generation and Verification
// ============================================================================

// TestSGXQuoteGeneration tests SGX quote generation in simulation mode.
func (s *TEEAttestationE2ETestSuite) TestSGXQuoteGeneration() {
	t := s.T()

	// Create SGX quote generator in simulation mode
	quoteGenerator := sgx.NewQuoteGenerator(nil)
	require.NotNil(t, quoteGenerator)

	// Generate report data with random nonce
	var reportData [64]byte
	_, err := rand.Read(reportData[:32])
	require.NoError(t, err)

	// Generate quote
	quote, err := quoteGenerator.GenerateQuote(reportData)
	require.NoError(t, err)
	require.NotNil(t, quote)

	// Verify quote structure
	require.Equal(t, uint16(3), quote.Header.Version, "DCAP quote version should be 3")
	require.NotEmpty(t, quote.ReportBody.MREnclave[:])
	require.NotEmpty(t, quote.ReportBody.MRSigner[:])
	require.Equal(t, reportData, quote.ReportBody.ReportData)

	t.Logf("SGX Quote generated:")
	t.Logf("  MRENCLAVE: %s", hex.EncodeToString(quote.ReportBody.MREnclave[:]))
	t.Logf("  MRSIGNER: %s", hex.EncodeToString(quote.ReportBody.MRSigner[:]))
}

// TestSGXQuoteSerialization tests SGX quote serialization and parsing.
func (s *TEEAttestationE2ETestSuite) TestSGXQuoteSerialization() {
	t := s.T()

	quoteGenerator := sgx.NewQuoteGenerator(nil)

	var reportData [64]byte
	copy(reportData[:], []byte("test-attestation-data"))

	// Generate quote
	quote, err := quoteGenerator.GenerateQuote(reportData)
	require.NoError(t, err)

	// Serialize
	serialized := sgx.SerializeQuote(quote)
	require.NotEmpty(t, serialized)

	// Parse back
	parsed, err := sgx.ParseQuote(serialized)
	require.NoError(t, err)
	require.NotNil(t, parsed)

	// Verify round-trip
	require.Equal(t, quote.Header.Version, parsed.Header.Version)
	require.Equal(t, quote.ReportBody.MREnclave, parsed.ReportBody.MREnclave)
	require.Equal(t, quote.ReportBody.MRSigner, parsed.ReportBody.MRSigner)
}

// TestSGXDCAPVerification tests SGX DCAP quote verification workflow.
func (s *TEEAttestationE2ETestSuite) TestSGXDCAPVerification() {
	t := s.T()

	if os.Getenv("VEID_E2E_DCAP") != "true" {
		t.Skip("VEID_E2E_DCAP not set; skipping DCAP verification")
	}

	// Create DCAP client with simulation mode
	cfg := sgx.DefaultDCAPClientConfig()
	cfg.PCSBaseURL = "https://api.trustedservices.intel.com/sgx/certification/v4"
	cfg.Timeout = 30 * time.Second
	cfg.AllowOutOfDateTCB = true
	cfg.SkipCRLCheck = true
	client := sgx.NewDCAPClient(cfg)
	require.NotNil(t, client)

	// Generate test quote
	quoteGenerator := sgx.NewQuoteGenerator(nil)
	var reportData [64]byte
	copy(reportData[:], []byte("verification-test"))

	quote, err := quoteGenerator.GenerateQuote(reportData)
	require.NoError(t, err)

	serialized := sgx.SerializeQuote(quote)

	// Fetch collateral and verify quote
	collateral, err := client.GetCollateralWithContext(s.ctx, serialized)
	require.NoError(t, err)

	result, err := client.VerifyQuote(serialized, collateral)
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Logf("SGX DCAP verification result:")
	t.Logf("  Valid: %v", result.Valid)
	t.Logf("  TCB Status: %s", result.TCBStatus)
}

// ============================================================================
// Test: SEV-SNP Report Generation and Verification
// ============================================================================

// TestSEVSNPReportGeneration tests SEV-SNP report generation.
func (s *TEEAttestationE2ETestSuite) TestSEVSNPReportGeneration() {
	t := s.T()

	// Create SEV guest in simulation mode
	guest := sev.NewSEVGuest()
	require.NotNil(t, guest)

	err := guest.Initialize()
	require.NoError(t, err)
	defer guest.Close()

	// Generate attestation with user data
	var userData [64]byte
	_, err = rand.Read(userData[:])
	require.NoError(t, err)

	attestation, err := guest.GenerateAttestation(userData)
	require.NoError(t, err)
	require.NotEmpty(t, attestation)

	t.Logf("SEV-SNP attestation generated: %d bytes", len(attestation))
}

// TestSEVSNPReportParsing tests SEV-SNP report parsing and serialization.
func (s *TEEAttestationE2ETestSuite) TestSEVSNPReportParsing() {
	t := s.T()

	guest := sev.NewSEVGuest()
	require.NoError(t, guest.Initialize())
	defer guest.Close()

	var userData [64]byte
	copy(userData[:], []byte("snp-test-data"))

	attestation, err := guest.GenerateAttestation(userData)
	require.NoError(t, err)

	// Parse report
	report, err := sev.ParseReport(attestation)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Verify report fields
	require.Equal(t, uint32(2), report.Version, "SNP report version should be 2")
	require.NotEmpty(t, report.LaunchDigest[:])
	require.Equal(t, userData, report.ReportData)

	// Validate report structure
	err = sev.ValidateReport(report)
	require.NoError(t, err)

	// Serialize and re-parse
	serialized, err := sev.SerializeReport(report)
	require.NoError(t, err)
	reParsed, err := sev.ParseReport(serialized)
	require.NoError(t, err)
	require.Equal(t, report.LaunchDigest, reParsed.LaunchDigest)

	t.Logf("SEV-SNP Report parsed:")
	t.Logf("  Launch Digest: %s", hex.EncodeToString(report.LaunchDigest[:]))
	t.Logf("  Guest Policy: %+v", report.Policy)
}

// TestSEVSNPKeyDerivation tests hardware-bound key derivation.
func (s *TEEAttestationE2ETestSuite) TestSEVSNPKeyDerivation() {
	t := s.T()

	guest := sev.NewSEVGuest()
	require.NoError(t, guest.Initialize())
	defer guest.Close()

	// Derive key with specific context
	ctx := []byte("test-key-context-v1")
	key1, err := guest.DeriveKeyWithContext(&sev.KeyRequest{
		RootKeySelect:    sev.KeyRootVCEK,
		GuestFieldSelect: sev.KeyFieldGuest,
	}, ctx, 32)
	require.NoError(t, err)
	require.Len(t, key1, 32)

	// Same context should produce same key
	key2, err := guest.DeriveKeyWithContext(&sev.KeyRequest{
		RootKeySelect:    sev.KeyRootVCEK,
		GuestFieldSelect: sev.KeyFieldGuest,
	}, ctx, 32)
	require.NoError(t, err)
	require.Equal(t, key1, key2, "Same context should produce same key")

	// Different context should produce different key
	key3, err := guest.DeriveKeyWithContext(&sev.KeyRequest{
		RootKeySelect:    sev.KeyRootVCEK,
		GuestFieldSelect: sev.KeyFieldGuest,
	}, []byte("different-context"), 32)
	require.NoError(t, err)
	require.NotEqual(t, key1, key3, "Different context should produce different key")
}

// TestSEVSNPKDSClient tests AMD Key Distribution Server client.
func (s *TEEAttestationE2ETestSuite) TestSEVSNPKDSClient() {
	t := s.T()

	if os.Getenv("VEID_E2E_KDS") != "true" {
		t.Skip("VEID_E2E_KDS not set; skipping KDS verification")
	}

	// Create KDS client
	client := sev.NewKDSClient(sev.MilanConfig())
	require.NotNil(t, client)

	// Get simulated certificate chain
	chain, err := client.GetCertificateChainWithContext(s.ctx, make([]byte, 64), sev.TCBVersion{
		BootLoader: 3,
		TEE:        0,
		SNP:        14,
		Microcode:  209,
	})
	require.NoError(t, err)
	require.NotNil(t, chain)
	require.NotNil(t, chain.VCEK)
	require.NotNil(t, chain.ASK)
	require.NotNil(t, chain.ARK)

	t.Log("SEV-SNP KDS client working (simulation mode)")
}

// ============================================================================
// Test: Nitro Attestation
// ============================================================================

// TestNitroAttestationGeneration tests Nitro attestation document generation.
func (s *TEEAttestationE2ETestSuite) TestNitroAttestationGeneration() {
	t := s.T()

	device := nitro.NewNSMDevice()
	require.NoError(t, device.Open())
	defer device.Close()

	// Generate attestation document
	userData := []byte("nitro-test-user-data")
	nonce := make([]byte, 32)
	_, err := rand.Read(nonce)
	require.NoError(t, err)

	attestation, err := device.GetAttestation(userData, nonce, nil)
	require.NoError(t, err)
	require.NotEmpty(t, attestation)

	t.Logf("Nitro attestation generated: %d bytes", len(attestation))
}

// TestNitroDocumentParsing tests Nitro attestation document parsing.
func (s *TEEAttestationE2ETestSuite) TestNitroDocumentParsing() {
	t := s.T()

	device := nitro.NewNSMDevice()
	require.NoError(t, device.Open())
	defer device.Close()

	userData := []byte("parse-test")
	nonce := []byte("test-nonce-12345678901234567890")

	attestation, err := device.GetAttestation(userData, nonce, nil)
	require.NoError(t, err)

	// Parse document
	doc, err := nitro.ParseDocument(attestation)
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Validate document structure
	err = nitro.ValidateDocument(doc)
	require.NoError(t, err)

	// Verify nonce is embedded
	require.Equal(t, nonce, doc.Payload.Nonce)
	require.Equal(t, userData, doc.Payload.UserData)

	t.Logf("Nitro document parsed:")
	t.Logf("  Module ID: %s", doc.Payload.ModuleID)
	t.Logf("  PCR0: %s", hex.EncodeToString(doc.Payload.PCRs[0]))
}

// TestNitroDocumentVerification tests Nitro attestation verification.
func (s *TEEAttestationE2ETestSuite) TestNitroDocumentVerification() {
	t := s.T()

	// Create verifier
	config := nitro.DefaultVerifierConfig()
	config.MaxDocumentAge = 24 * time.Hour
	config.AllowSimulated = true
	config.SkipSignatureVerification = true
	config.SkipCertificateChainVerification = true
	verifier := nitro.NewVerifierWithConfig(config)
	require.NotNil(t, verifier)

	// Generate attestation
	device := nitro.NewNSMDevice()
	require.NoError(t, device.Open())
	defer device.Close()

	attestation, err := device.GetAttestation([]byte("verify-test"), nil, nil)
	require.NoError(t, err)

	doc, err := nitro.ParseDocument(attestation)
	require.NoError(t, err)

	// Verify
	result, err := verifier.VerifyDocument(doc)
	require.NoError(t, err)
	require.True(t, result.Valid)

	t.Log("Nitro document verification passed")
}

// ============================================================================
// Test: Remote Attestation Protocol
// ============================================================================

// TestRemoteAttestationChallenge tests the attestation challenge-response protocol.
func (s *TEEAttestationE2ETestSuite) TestRemoteAttestationChallenge() {
	t := s.T()

	// Create enclave service
	enclaveService := enclave_runtime.NewSimulatedEnclaveService()
	require.NoError(t, enclaveService.Initialize(enclave_runtime.DefaultRuntimeConfig()))
	defer enclaveService.Shutdown()

	// Create verifier
	verifier := enclave_runtime.NewSimpleAttestationVerifier()

	// Create remote attestation protocol
	protocol := enclave_runtime.NewRemoteAttestationProtocol(
		enclaveService,
		verifier,
		enclave_runtime.RemoteAttestationConfig{
			NonceSize:            32,
			ChallengeTimeout:     30 * time.Second,
			MaxPendingChallenges: 100,
			AllowSimulated:       true,
		},
	)
	require.NotNil(t, protocol)

	// Generate challenge for peer
	peerID := "validator-peer-001"
	request, err := protocol.GenerateChallenge(peerID)
	require.NoError(t, err)
	require.NotNil(t, request)
	require.NotEmpty(t, request.ChallengeID)
	require.Len(t, request.Nonce, 32)

	// Simulate peer handling the challenge
	response, err := protocol.HandleChallengeRequest(s.ctx, request, peerID)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.True(t, response.IsSuccess())

	// Verify the response
	verifier.AddAllowedMeasurement(response.Measurement)
	result, err := protocol.VerifyResponse(response)
	require.NoError(t, err)
	require.True(t, result.Valid)
	require.Equal(t, peerID, result.PeerID)
	require.NotEmpty(t, result.Measurement)

	t.Logf("Remote attestation successful:")
	t.Logf("  Peer ID: %s", result.PeerID)
	t.Logf("  Platform: %s", result.Platform)
	t.Logf("  Measurement: %s", result.MeasurementHex)
}

// TestRemoteAttestationChallengeExpiry tests challenge expiration.
func (s *TEEAttestationE2ETestSuite) TestRemoteAttestationChallengeExpiry() {
	t := s.T()

	enclaveService := enclave_runtime.NewSimulatedEnclaveService()
	require.NoError(t, enclaveService.Initialize(enclave_runtime.DefaultRuntimeConfig()))
	defer enclaveService.Shutdown()

	protocol := enclave_runtime.NewRemoteAttestationProtocol(
		enclaveService,
		nil,
		enclave_runtime.RemoteAttestationConfig{
			NonceSize:            32,
			ChallengeTimeout:     100 * time.Millisecond, // Very short timeout
			MaxPendingChallenges: 10,
			AllowSimulated:       true,
		},
	)

	// Generate challenge
	request, err := protocol.GenerateChallenge("peer-expiry-test")
	require.NoError(t, err)

	// Wait for expiry
	time.Sleep(200 * time.Millisecond)

	// Create fake response with expired challenge ID
	response := &enclave_runtime.AttestationResponse{
		ChallengeID:     request.ChallengeID,
		ResponderID:     "peer-expiry-test",
		Platform:        enclave_runtime.AttestationTypeSimulated,
		AttestationData: []byte("simulated"),
		Measurement:     make([]byte, 32),
		Timestamp:       time.Now(),
	}

	// Should fail due to expiry
	result, err := protocol.VerifyResponse(response)
	require.NoError(t, err)
	require.False(t, result.Valid)
	require.Contains(t, result.Errors, "challenge expired")
}

// ============================================================================
// Test: Hardware Manager
// ============================================================================

// TestHardwareManagerLifecycle tests hardware manager initialization and shutdown.
func (s *TEEAttestationE2ETestSuite) TestHardwareManagerLifecycle() {
	t := s.T()

	config := hardware.DefaultConfig()
	config.AllowSimulation = true
	config.RequireHardware = false
	manager, err := hardware.NewHardwareManager(config)
	require.NoError(t, err)
	require.NotNil(t, manager)

	// Initialize
	err = manager.Initialize(context.Background())
	require.NoError(t, err)

	// Verify state
	require.True(t, manager.IsInitialized())

	// Get active backend
	backend := manager.GetBackend()
	require.NotNil(t, backend)
	t.Logf("Hardware manager backend: %s", backend.Platform())

	// Shutdown
	err = manager.Shutdown()
	require.NoError(t, err)
	require.False(t, manager.IsInitialized())
}

// TestHardwareManagerAttestation tests attestation via hardware manager.
func (s *TEEAttestationE2ETestSuite) TestHardwareManagerAttestation() {
	t := s.T()

	config := hardware.DefaultConfig()
	config.AllowSimulation = true
	config.RequireHardware = false
	manager, err := hardware.NewHardwareManager(config)
	require.NoError(t, err)
	require.NoError(t, manager.Initialize(context.Background()))
	defer manager.Shutdown()

	// Generate nonce
	nonce := make([]byte, 32)
	_, err = rand.Read(nonce)
	require.NoError(t, err)

	// Get attestation
	attestation, err := manager.GetAttestation(nonce)
	require.NoError(t, err)
	require.NotEmpty(t, attestation)

	t.Logf("Hardware manager attestation: %d bytes", len(attestation))
}

// TestHardwareManagerKeyDerivation tests key derivation via hardware manager.
func (s *TEEAttestationE2ETestSuite) TestHardwareManagerKeyDerivation() {
	t := s.T()

	config := hardware.DefaultConfig()
	config.AllowSimulation = true
	config.RequireHardware = false
	manager, err := hardware.NewHardwareManager(config)
	require.NoError(t, err)
	require.NoError(t, manager.Initialize(context.Background()))
	defer manager.Shutdown()

	// Derive key
	keyCtx := []byte("manager-key-test")
	key, err := manager.DeriveKey(keyCtx, 32)
	require.NoError(t, err)
	require.Len(t, key, 32)

	// Same context should give same key
	key2, err := manager.DeriveKey(keyCtx, 32)
	require.NoError(t, err)
	require.Equal(t, key, key2)
}

// TestHardwareManagerSealUnseal tests sealing and unsealing via hardware manager.
func (s *TEEAttestationE2ETestSuite) TestHardwareManagerSealUnseal() {
	t := s.T()

	config := hardware.DefaultConfig()
	config.AllowSimulation = true
	config.RequireHardware = false
	manager, err := hardware.NewHardwareManager(config)
	require.NoError(t, err)
	require.NoError(t, manager.Initialize(context.Background()))
	defer manager.Shutdown()

	// Test data
	plaintext := []byte("secret data to seal for TEE protection")

	// Seal
	sealed, err := manager.Seal(plaintext)
	require.NoError(t, err)
	require.NotEmpty(t, sealed)
	require.False(t, bytes.Equal(plaintext, sealed))

	// Unseal
	unsealed, err := manager.Unseal(sealed)
	require.NoError(t, err)
	require.Equal(t, plaintext, unsealed)
}

// ============================================================================
// Test: Mock Backend
// ============================================================================

// TestMockBackendBasicOperations tests mock backend for testing.
func (s *TEEAttestationE2ETestSuite) TestMockBackendBasicOperations() {
	t := s.T()

	mock := hardware.NewMockBackendWithDefaults()
	require.NoError(t, mock.Initialize())

	// Test attestation
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)

	attestation, err := mock.GetAttestation(nonce)
	require.NoError(t, err)
	require.NotEmpty(t, attestation)

	// Verify call was recorded
	require.True(t, mock.WasMethodCalled(hardware.MethodGetAttestation))
	require.Equal(t, 1, mock.GetCallCount(hardware.MethodGetAttestation))
}

// TestMockBackendConfigurableFailure tests mock failure injection.
func (s *TEEAttestationE2ETestSuite) TestMockBackendConfigurableFailure() {
	t := s.T()

	mock := hardware.NewMockBackendWithDefaults()
	require.NoError(t, mock.Initialize())

	// Configure failure
	testError := errors.New("simulated hardware failure")
	mock.ConfigureFailure(hardware.MethodGetAttestation, testError)

	// Should fail
	_, err := mock.GetAttestation([]byte("nonce"))
	require.Error(t, err)
	require.Equal(t, testError, err)
}

// TestMockBackendCallRecording tests call recording for assertions.
func (s *TEEAttestationE2ETestSuite) TestMockBackendCallRecording() {
	t := s.T()

	mock := hardware.NewMockBackendWithDefaults()
	require.NoError(t, mock.Initialize())

	// Make several calls
	_, _ = mock.GetAttestation([]byte("nonce1"))
	_, _ = mock.GetAttestation([]byte("nonce2"))
	_, _ = mock.DeriveKey([]byte("context"), 32)
	_, _ = mock.Seal([]byte("plaintext"))
	_, _ = mock.Unseal(make([]byte, 44)) // nonce + ciphertext

	// Verify recordings
	require.Equal(t, 2, mock.GetCallCount(hardware.MethodGetAttestation))
	require.Equal(t, 1, mock.GetCallCount(hardware.MethodDeriveKey))
	require.Equal(t, 1, mock.GetCallCount(hardware.MethodSeal))
	require.Equal(t, 1, mock.GetCallCount(hardware.MethodUnseal))

	// Reset
	mock.Reset()
	require.Equal(t, 0, mock.GetCallCount(hardware.MethodGetAttestation))
}

// ============================================================================
// Test: Cross-Platform Verification
// ============================================================================

// TestCrossPlatformAttestationVerification tests verification across platforms.
func (s *TEEAttestationE2ETestSuite) TestCrossPlatformAttestationVerification() {
	t := s.T()

	// Generate attestations from each platform
	platforms := []struct {
		name     string
		generate func() ([]byte, error)
	}{
		{
			name: "SGX",
			generate: func() ([]byte, error) {
				quoteGenerator := sgx.NewQuoteGenerator(nil)
				var rd [64]byte
				_, _ = rand.Read(rd[:])
				q, err := quoteGenerator.GenerateQuote(rd)
				if err != nil {
					return nil, err
				}
				return sgx.SerializeQuote(q), nil
			},
		},
		{
			name: "SEV-SNP",
			generate: func() ([]byte, error) {
				guest := sev.NewSEVGuest()
				if err := guest.Initialize(); err != nil {
					return nil, err
				}
				defer guest.Close()
				var ud [64]byte
				_, _ = rand.Read(ud[:])
				return guest.GenerateAttestation(ud)
			},
		},
		{
			name: "Nitro",
			generate: func() ([]byte, error) {
				device := nitro.NewNSMDevice()
				if err := device.Open(); err != nil {
					return nil, err
				}
				defer device.Close()
				return device.GetAttestation([]byte("test"), nil, nil)
			},
		},
	}

	for _, p := range platforms {
		t.Run(p.name, func(t *testing.T) {
			attestation, err := p.generate()
			require.NoError(t, err)
			require.NotEmpty(t, attestation)

			// Detect attestation type
			detectedType := enclave_runtime.DetectAttestationType(attestation)
			t.Logf("%s detected as: %s", p.name, detectedType)
		})
	}
}

// ============================================================================
// Test: Measurement Allowlist
// ============================================================================

// TestMeasurementAllowlist tests measurement allowlist validation.
func (s *TEEAttestationE2ETestSuite) TestMeasurementAllowlist() {
	t := s.T()

	// Create verifier with specific allowlist
	allowedMeasurement := sha256.Sum256([]byte("approved-enclave-v1"))
	disallowedMeasurement := sha256.Sum256([]byte("unapproved-enclave"))

	allowlist := enclave_runtime.NewMeasurementAllowlist()

	// Add to allowlist
	require.NoError(t, allowlist.AddMeasurement(enclave_runtime.AttestationTypeSimulated, allowedMeasurement[:], "approved"))

	// Should pass
	require.True(t, allowlist.IsTrusted(enclave_runtime.AttestationTypeSimulated, allowedMeasurement[:]))

	// Should fail
	require.False(t, allowlist.IsTrusted(enclave_runtime.AttestationTypeSimulated, disallowedMeasurement[:]))

	// Remove from allowlist
	require.NoError(t, allowlist.RemoveMeasurement(enclave_runtime.AttestationTypeSimulated, allowedMeasurement[:]))
	require.False(t, allowlist.IsTrusted(enclave_runtime.AttestationTypeSimulated, allowedMeasurement[:]))
}

// ============================================================================
// Test: Error Handling
// ============================================================================

// TestInvalidAttestationHandling tests handling of invalid attestations.
func (s *TEEAttestationE2ETestSuite) TestInvalidAttestationHandling() {
	t := s.T()

	// Invalid SGX quote
	_, err := sgx.ParseQuote([]byte("invalid-quote-data"))
	require.Error(t, err)

	// Invalid SEV-SNP report
	_, err = sev.ParseReport([]byte("invalid-report-data"))
	require.Error(t, err)

	// Invalid Nitro document
	_, err = nitro.ParseDocument([]byte("invalid-document-data"))
	require.Error(t, err)
}

// TestHardwareNotAvailable tests handling when hardware is not available.
func (s *TEEAttestationE2ETestSuite) TestHardwareNotAvailable() {
	t := s.T()

	// Create manager in require mode (should fail on non-TEE system)
	config := hardware.DefaultConfig()
	config.RequireHardware = true
	config.AllowSimulation = false
	config.PreferredPlatform = hardware.PlatformSGX
	manager, err := hardware.NewHardwareManager(config)
	require.NoError(t, err)

	// This may fail on non-TEE hardware
	err = manager.Initialize(context.Background())
	if err != nil {
		// Expected on non-TEE systems
		require.ErrorIs(t, err, hardware.ErrHardwareNotFound)
		t.Log("Hardware not available (expected on non-TEE systems)")
	} else {
		// TEE hardware actually available
		t.Log("TEE hardware available!")
		_ = manager.Shutdown()
	}
}

// ============================================================================
// Test: Concurrent Operations
// ============================================================================

// TestConcurrentAttestations tests concurrent attestation generation.
func (s *TEEAttestationE2ETestSuite) TestConcurrentAttestations() {
	t := s.T()

	config := hardware.DefaultConfig()
	config.AllowSimulation = true
	config.RequireHardware = false
	manager, err := hardware.NewHardwareManager(config)
	require.NoError(t, err)
	require.NoError(t, manager.Initialize(context.Background()))
	defer manager.Shutdown()

	// Run concurrent attestations
	const numGoroutines = 10
	results := make(chan []byte, numGoroutines)
	errs := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			nonce := make([]byte, 32)
			_, _ = rand.Read(nonce)
			attestation, err := manager.GetAttestation(nonce)
			if err != nil {
				errs <- err
				return
			}
			results <- attestation
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		select {
		case att := <-results:
			require.NotEmpty(t, att)
		case err := <-errs:
			t.Fatalf("Concurrent attestation failed: %v", err)
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for attestations")
		}
	}
}

// TestConcurrentSealUnseal tests concurrent seal/unseal operations.
func (s *TEEAttestationE2ETestSuite) TestConcurrentSealUnseal() {
	t := s.T()

	config := hardware.DefaultConfig()
	config.AllowSimulation = true
	config.RequireHardware = false
	manager, err := hardware.NewHardwareManager(config)
	require.NoError(t, err)
	require.NoError(t, manager.Initialize(context.Background()))
	defer manager.Shutdown()

	// Run concurrent seal/unseal
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			data := []byte(fmt.Sprintf("concurrent-data-%d", idx))
			sealed, err := manager.Seal(data)
			if err != nil {
				t.Errorf("Seal failed: %v", err)
				done <- false
				return
			}

			unsealed, err := manager.Unseal(sealed)
			if err != nil {
				t.Errorf("Unseal failed: %v", err)
				done <- false
				return
			}

			if !bytes.Equal(data, unsealed) {
				t.Errorf("Data mismatch: expected %s, got %s", data, unsealed)
				done <- false
				return
			}

			done <- true
		}(i)
	}

	// Wait for all
	successCount := 0
	for i := 0; i < numGoroutines; i++ {
		select {
		case success := <-done:
			if success {
				successCount++
			}
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for seal/unseal")
		}
	}

	require.Equal(t, numGoroutines, successCount, "All concurrent operations should succeed")
}

// ============================================================================
// Benchmark Tests
// ============================================================================

// BenchmarkAttestationGeneration benchmarks attestation generation.
func BenchmarkAttestationGeneration(b *testing.B) {
	config := hardware.DefaultConfig()
	config.AllowSimulation = true
	config.RequireHardware = false
	manager, err := hardware.NewHardwareManager(config)
	if err != nil {
		b.Fatal(err)
	}
	if err := manager.Initialize(context.Background()); err != nil {
		b.Fatal(err)
	}
	defer manager.Shutdown()

	nonce := make([]byte, 32)
	rand.Read(nonce)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetAttestation(nonce)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSealUnseal benchmarks seal/unseal operations.
func BenchmarkSealUnseal(b *testing.B) {
	config := hardware.DefaultConfig()
	config.AllowSimulation = true
	config.RequireHardware = false
	manager, err := hardware.NewHardwareManager(config)
	if err != nil {
		b.Fatal(err)
	}
	if err := manager.Initialize(context.Background()); err != nil {
		b.Fatal(err)
	}
	defer manager.Shutdown()

	data := make([]byte, 1024)
	rand.Read(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sealed, err := manager.Seal(data)
		if err != nil {
			b.Fatal(err)
		}
		_, err = manager.Unseal(sealed)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkKeyDerivation benchmarks key derivation.
func BenchmarkKeyDerivation(b *testing.B) {
	config := hardware.DefaultConfig()
	config.AllowSimulation = true
	config.RequireHardware = false
	manager, err := hardware.NewHardwareManager(config)
	if err != nil {
		b.Fatal(err)
	}
	if err := manager.Initialize(context.Background()); err != nil {
		b.Fatal(err)
	}
	defer manager.Shutdown()

	context := []byte("benchmark-context")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.DeriveKey(context, 32)
		if err != nil {
			b.Fatal(err)
		}
	}
}
