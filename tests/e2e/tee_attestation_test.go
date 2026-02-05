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
	mockBackend     *hardware.MockBackend
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

	// Create mock backend for deterministic testing
	s.mockBackend = hardware.NewMockBackend()
	s.mockBackend.SetAvailable(true)
	s.mockBackend.SetInitialized(true)

	// Create hardware manager with simulation mode
	s.hardwareManager = hardware.NewHardwareManager(hardware.Config{
		Mode:              hardware.ModeSimulate,
		EnableHealthCheck: false,
	})
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
	detector := hardware.GetDetector()
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
	t.Logf("  Preferred Platform: %s", caps.PreferredPlatform)

	// On non-Linux or non-TEE hardware, should fall back to simulated
	if caps.PreferredPlatform == hardware.PlatformSimulated {
		t.Log("No hardware TEE detected, simulation mode active")
	}
}

// TestHardwareDetectionCaching tests that detection results are properly cached.
func (s *TEEAttestationE2ETestSuite) TestHardwareDetectionCaching() {
	t := s.T()

	detector := hardware.GetDetector()

	// First detection
	caps1, err := detector.Detect()
	require.NoError(t, err)

	// Second detection should use cache
	caps2, err := detector.Detect()
	require.NoError(t, err)

	// Results should be identical
	require.Equal(t, caps1.PreferredPlatform, caps2.PreferredPlatform)
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

	// Create SGX enclave in simulation mode
	enclave := sgx.NewEnclave()
	require.NotNil(t, enclave)

	// Generate report data with random nonce
	var reportData [64]byte
	_, err := rand.Read(reportData[:32])
	require.NoError(t, err)

	// Generate quote
	quote, err := enclave.GenerateQuote(reportData)
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

	enclave := sgx.NewEnclave()

	var reportData [64]byte
	copy(reportData[:], []byte("test-attestation-data"))

	// Generate quote
	quote, err := enclave.GenerateQuote(reportData)
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

	// Create DCAP client with simulation mode
	client := sgx.NewDCAPClient(sgx.DCAPConfig{
		PCSURL:     "https://api.trustedservices.intel.com/sgx/certification/v4",
		Timeout:    30 * time.Second,
		CacheSize:  100,
		Simulation: true,
	})
	require.NotNil(t, client)

	// Generate test quote
	enclave := sgx.NewEnclave()
	var reportData [64]byte
	copy(reportData[:], []byte("verification-test"))

	quote, err := enclave.GenerateQuote(reportData)
	require.NoError(t, err)

	serialized := sgx.SerializeQuote(quote)

	// Verify quote (in simulation mode)
	result, err := client.VerifyQuote(s.ctx, serialized, nil)
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
	serialized := sev.SerializeReport(report)
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
	key1, err := guest.DeriveKey(&sev.KeyRequest{
		RootKey:     sev.KeyRootVCEK,
		FieldSelect: sev.KeyFieldGuest,
		Context:     ctx,
	})
	require.NoError(t, err)
	require.Len(t, key1, 32)

	// Same context should produce same key
	key2, err := guest.DeriveKey(&sev.KeyRequest{
		RootKey:     sev.KeyRootVCEK,
		FieldSelect: sev.KeyFieldGuest,
		Context:     ctx,
	})
	require.NoError(t, err)
	require.Equal(t, key1, key2, "Same context should produce same key")

	// Different context should produce different key
	key3, err := guest.DeriveKey(&sev.KeyRequest{
		RootKey:     sev.KeyRootVCEK,
		FieldSelect: sev.KeyFieldGuest,
		Context:     []byte("different-context"),
	})
	require.NoError(t, err)
	require.NotEqual(t, key1, key3, "Different context should produce different key")
}

// TestSEVSNPKDSClient tests AMD Key Distribution Server client.
func (s *TEEAttestationE2ETestSuite) TestSEVSNPKDSClient() {
	t := s.T()

	// Create KDS client with simulation
	client := sev.NewKDSClient(sev.KDSConfig{
		ProcessorFamily: sev.ProcessorMilan,
		Timeout:         30 * time.Second,
		CacheSize:       100,
		Simulation:      true,
	})
	require.NotNil(t, client)

	// Get simulated certificate chain
	chain, err := client.GetCertificateChain(s.ctx, make([]byte, 64), sev.TCBVersion{
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

	// Create Nitro enclave in simulation mode
	enc := nitro.NewNitroEnclave()
	require.NotNil(t, enc)

	err := enc.Initialize()
	require.NoError(t, err)
	defer enc.Close()

	// Generate attestation document
	userData := []byte("nitro-test-user-data")
	nonce := make([]byte, 32)
	_, err = rand.Read(nonce)
	require.NoError(t, err)

	attestation, err := enc.GetAttestation(userData, nonce, nil)
	require.NoError(t, err)
	require.NotEmpty(t, attestation)

	t.Logf("Nitro attestation generated: %d bytes", len(attestation))
}

// TestNitroDocumentParsing tests Nitro attestation document parsing.
func (s *TEEAttestationE2ETestSuite) TestNitroDocumentParsing() {
	t := s.T()

	enc := nitro.NewNitroEnclave()
	require.NoError(t, enc.Initialize())
	defer enc.Close()

	userData := []byte("parse-test")
	nonce := []byte("test-nonce-12345678901234567890")

	attestation, err := enc.GetAttestation(userData, nonce, nil)
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
	verifier := nitro.NewVerifier(nitro.VerifierConfig{
		MaxAge:     24 * time.Hour,
		Simulation: true,
	})
	require.NotNil(t, verifier)

	// Generate attestation
	enc := nitro.NewNitroEnclave()
	require.NoError(t, enc.Initialize())
	defer enc.Close()

	attestation, err := enc.GetAttestation([]byte("verify-test"), nil, nil)
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
	enclaveService, err := enclave_runtime.NewSimulatedEnclaveService(
		enclave_runtime.SimulatedEnclaveServiceOptions{
			EnableAttestation: true,
		},
	)
	require.NoError(t, err)
	defer enclaveService.Shutdown()

	// Create verifier
	verifier := enclave_runtime.NewAttestationVerifier(
		enclave_runtime.PermissiveVerificationPolicy(),
	)

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

	enclaveService, err := enclave_runtime.NewSimulatedEnclaveService(
		enclave_runtime.SimulatedEnclaveServiceOptions{},
	)
	require.NoError(t, err)
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

	manager := hardware.NewHardwareManager(hardware.Config{
		Mode:              hardware.ModeSimulate,
		EnableHealthCheck: false,
	})
	require.NotNil(t, manager)

	// Initialize
	err := manager.Initialize()
	require.NoError(t, err)

	// Verify state
	require.True(t, manager.IsInitialized())

	// Get active backend
	backend := manager.GetBackend()
	require.NotNil(t, backend)
	require.Equal(t, hardware.PlatformSimulated, backend.Platform())

	// Shutdown
	err = manager.Shutdown()
	require.NoError(t, err)
	require.False(t, manager.IsInitialized())
}

// TestHardwareManagerAttestation tests attestation via hardware manager.
func (s *TEEAttestationE2ETestSuite) TestHardwareManagerAttestation() {
	t := s.T()

	manager := hardware.NewHardwareManager(hardware.Config{
		Mode: hardware.ModeSimulate,
	})
	require.NoError(t, manager.Initialize())
	defer manager.Shutdown()

	// Generate nonce
	nonce := make([]byte, 32)
	_, err := rand.Read(nonce)
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

	manager := hardware.NewHardwareManager(hardware.Config{
		Mode: hardware.ModeSimulate,
	})
	require.NoError(t, manager.Initialize())
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

	manager := hardware.NewHardwareManager(hardware.Config{
		Mode: hardware.ModeSimulate,
	})
	require.NoError(t, manager.Initialize())
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

	mock := hardware.NewMockBackend()
	mock.SetAvailable(true)
	mock.SetInitialized(true)

	// Test attestation
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)

	attestation, err := mock.GetAttestation(nonce)
	require.NoError(t, err)
	require.NotEmpty(t, attestation)

	// Verify call was recorded
	require.True(t, mock.WasCalled(hardware.MockMethodGetAttestation))
	require.Equal(t, 1, mock.GetCallCount(hardware.MockMethodGetAttestation))
}

// TestMockBackendConfigurableFailure tests mock failure injection.
func (s *TEEAttestationE2ETestSuite) TestMockBackendConfigurableFailure() {
	t := s.T()

	mock := hardware.NewMockBackend()
	mock.SetAvailable(true)
	mock.SetInitialized(true)

	// Configure failure
	testError := errors.New("simulated hardware failure")
	mock.SetError(hardware.MockMethodGetAttestation, testError)

	// Should fail
	_, err := mock.GetAttestation([]byte("nonce"))
	require.Error(t, err)
	require.Equal(t, testError, err)
}

// TestMockBackendCallRecording tests call recording for assertions.
func (s *TEEAttestationE2ETestSuite) TestMockBackendCallRecording() {
	t := s.T()

	mock := hardware.NewMockBackend()
	mock.SetAvailable(true)
	mock.SetInitialized(true)

	// Make several calls
	_, _ = mock.GetAttestation([]byte("nonce1"))
	_, _ = mock.GetAttestation([]byte("nonce2"))
	_, _ = mock.DeriveKey([]byte("context"), 32)
	_, _ = mock.Seal([]byte("plaintext"))
	_, _ = mock.Unseal(make([]byte, 44)) // nonce + ciphertext

	// Verify recordings
	require.Equal(t, 2, mock.GetCallCount(hardware.MockMethodGetAttestation))
	require.Equal(t, 1, mock.GetCallCount(hardware.MockMethodDeriveKey))
	require.Equal(t, 1, mock.GetCallCount(hardware.MockMethodSeal))
	require.Equal(t, 1, mock.GetCallCount(hardware.MockMethodUnseal))

	// Reset
	mock.Reset()
	require.Equal(t, 0, mock.GetCallCount(hardware.MockMethodGetAttestation))
}

// ============================================================================
// Test: Cross-Platform Verification
// ============================================================================

// TestCrossPlatformAttestationVerification tests verification across platforms.
func (s *TEEAttestationE2ETestSuite) TestCrossPlatformAttestationVerification() {
	t := s.T()

	verifier := enclave_runtime.NewAttestationVerifier(
		enclave_runtime.VerificationPolicy{
			AllowDebugMode:   true,
			RequireLatestTCB: false,
			AllowedPlatforms: []enclave_runtime.AttestationType{
				enclave_runtime.AttestationTypeSGX,
				enclave_runtime.AttestationTypeSEVSNP,
				enclave_runtime.AttestationTypeNitro,
				enclave_runtime.AttestationTypeSimulated,
			},
			MinimumSecurityLevel: 1,
			MaxAttestationAge:    24 * time.Hour,
			RequireNonce:         false,
		},
	)
	require.NotNil(t, verifier)

	// Generate attestations from each platform
	platforms := []struct {
		name     string
		generate func() ([]byte, error)
	}{
		{
			name: "SGX",
			generate: func() ([]byte, error) {
				enc := sgx.NewEnclave()
				var rd [64]byte
				_, _ = rand.Read(rd[:])
				q, err := enc.GenerateQuote(rd)
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
				enc := nitro.NewNitroEnclave()
				if err := enc.Initialize(); err != nil {
					return nil, err
				}
				defer enc.Close()
				return enc.GetAttestation([]byte("test"), nil, nil)
			},
		},
	}

	for _, p := range platforms {
		t.Run(p.name, func(t *testing.T) {
			attestation, err := p.generate()
			require.NoError(t, err)
			require.NotEmpty(t, attestation)

			// Detect attestation type
			detectedType := verifier.DetectAttestationType(attestation)
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

	verifier := enclave_runtime.NewAttestationVerifier(
		enclave_runtime.VerificationPolicy{
			AllowDebugMode:       true,
			AllowedPlatforms:     []enclave_runtime.AttestationType{enclave_runtime.AttestationTypeSimulated},
			MinimumSecurityLevel: 1,
		},
	)

	// Add to allowlist
	verifier.AddMeasurementToAllowlist(allowedMeasurement[:])

	// Should pass
	require.True(t, verifier.IsMeasurementAllowed(allowedMeasurement[:]))

	// Should fail
	require.False(t, verifier.IsMeasurementAllowed(disallowedMeasurement[:]))

	// Remove from allowlist
	verifier.RemoveMeasurementFromAllowlist(allowedMeasurement[:])
	require.False(t, verifier.IsMeasurementAllowed(allowedMeasurement[:]))
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
	manager := hardware.NewHardwareManager(hardware.Config{
		Mode:              hardware.ModeRequire,
		PreferredPlatform: hardware.PlatformSGX, // Require specific platform
	})

	// This may fail on non-TEE hardware
	err := manager.Initialize()
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

	manager := hardware.NewHardwareManager(hardware.Config{
		Mode: hardware.ModeSimulate,
	})
	require.NoError(t, manager.Initialize())
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

	manager := hardware.NewHardwareManager(hardware.Config{
		Mode: hardware.ModeSimulate,
	})
	require.NoError(t, manager.Initialize())
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
	manager := hardware.NewHardwareManager(hardware.Config{
		Mode: hardware.ModeSimulate,
	})
	if err := manager.Initialize(); err != nil {
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
	manager := hardware.NewHardwareManager(hardware.Config{
		Mode: hardware.ModeSimulate,
	})
	if err := manager.Initialize(); err != nil {
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
	manager := hardware.NewHardwareManager(hardware.Config{
		Mode: hardware.ModeSimulate,
	})
	if err := manager.Initialize(); err != nil {
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
