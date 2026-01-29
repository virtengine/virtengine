package enclave_runtime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// SEV-SNP Enclave Service Implementation Tests
// =============================================================================

func TestSEVSNPEnclaveServiceImpl_NewService(t *testing.T) {
	tests := []struct {
		name    string
		config  SEVSNPConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: SEVSNPConfig{
				Endpoint:         "unix:///var/run/veid-enclave.sock",
				CertChainPath:    "/opt/certs/amd_chain.pem",
				MinTCBVersion:    "2.0.8.115",
				AllowDebugPolicy: false,
			},
			wantErr: false,
		},
		{
			name: "debug policy allowed for testing",
			config: SEVSNPConfig{
				Endpoint:         "unix:///var/run/veid-enclave.sock",
				AllowDebugPolicy: true,
			},
			wantErr: false,
		},
		{
			name: "missing endpoint",
			config: SEVSNPConfig{
				CertChainPath: "/opt/certs/amd_chain.pem",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewSEVSNPEnclaveServiceImpl(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, svc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, svc)
			}
		})
	}
}

func TestSEVSNPEnclaveServiceImpl_Initialize(t *testing.T) {
	config := SEVSNPConfig{
		Endpoint: "unix:///var/run/veid-enclave.sock",
	}

	svc, err := NewSEVSNPEnclaveServiceImpl(config)
	require.NoError(t, err)

	// Initialize
	err = svc.Initialize(DefaultRuntimeConfig())
	assert.NoError(t, err)

	// Check status
	status := svc.GetStatus()
	assert.True(t, status.Initialized)
	assert.True(t, status.Available)
	assert.Equal(t, uint64(1), status.CurrentEpoch)

	// Double initialization should fail
	err = svc.Initialize(DefaultRuntimeConfig())
	assert.Error(t, err)
}

func TestSEVSNPEnclaveServiceImpl_Score(t *testing.T) {
	config := SEVSNPConfig{
		Endpoint: "unix:///var/run/veid-enclave.sock",
	}

	svc, err := NewSEVSNPEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	request := &ScoringRequest{
		RequestID:      "snp-test-request-1",
		Ciphertext:     []byte("snp_encrypted_identity_data"),
		WrappedKey:     []byte("snp_wrapped_key"),
		Nonce:          []byte("snp_nonce_12345"),
		ScopeID:        "snp-scope-123",
		AccountAddress: "virtengine1snptest",
		BlockHeight:    12345,
	}

	result, err := svc.Score(context.Background(), request)
	require.NoError(t, err)
	require.True(t, result.IsSuccess())

	// Verify result fields
	assert.Equal(t, request.RequestID, result.RequestID)
	assert.LessOrEqual(t, result.Score, uint32(100))
	assert.NotEmpty(t, result.Status)
	assert.NotEmpty(t, result.EnclaveSignature)
	assert.NotEmpty(t, result.MeasurementHash)
	assert.NotEmpty(t, result.InputHash)

	// Verify SEV-SNP-specific fields
	assert.Contains(t, result.ReasonCodes[0], "snp_")
	assert.Len(t, result.MeasurementHash, SNPLaunchDigestSize)
}

func TestSEVSNPEnclaveServiceImpl_GetMeasurement(t *testing.T) {
	config := SEVSNPConfig{
		Endpoint: "unix:///var/run/veid-enclave.sock",
	}

	svc, err := NewSEVSNPEnclaveServiceImpl(config)
	require.NoError(t, err)

	// Should fail before init
	_, err = svc.GetMeasurement()
	assert.Error(t, err)

	// Initialize
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Should succeed after init (launch digest)
	launchDigest, err := svc.GetMeasurement()
	assert.NoError(t, err)
	assert.Len(t, launchDigest, SNPLaunchDigestSize)
}

func TestSEVSNPEnclaveServiceImpl_GenerateAttestation(t *testing.T) {
	config := SEVSNPConfig{
		Endpoint: "unix:///var/run/veid-enclave.sock",
	}

	svc, err := NewSEVSNPEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Generate attestation with nonce
	reportData := []byte("test_nonce_for_snp_attestation")
	report, err := svc.GenerateAttestation(reportData)
	assert.NoError(t, err)
	assert.NotEmpty(t, report)

	// Report should contain proper structure
	assert.Greater(t, len(report), 100)

	// Report data too large should fail
	largeData := make([]byte, SNPReportDataSize+1)
	_, err = svc.GenerateAttestation(largeData)
	assert.Error(t, err)
}

func TestSEVSNPEnclaveServiceImpl_KeyRotation(t *testing.T) {
	config := SEVSNPConfig{
		Endpoint: "unix:///var/run/veid-enclave.sock",
	}

	svc, err := NewSEVSNPEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Get initial keys
	encPub1, _ := svc.GetEncryptionPubKey()
	sigPub1, _ := svc.GetSigningPubKey()

	// Rotate keys
	err = svc.RotateKeys()
	assert.NoError(t, err)

	// Get new keys
	encPub2, _ := svc.GetEncryptionPubKey()
	sigPub2, _ := svc.GetSigningPubKey()

	// Keys should be different after rotation
	assert.NotEqual(t, encPub1, encPub2)
	assert.NotEqual(t, sigPub1, sigPub2)

	// Epoch should increase
	status := svc.GetStatus()
	assert.Equal(t, uint64(2), status.CurrentEpoch)
}

func TestSEVSNPEnclaveServiceImpl_GetChipID(t *testing.T) {
	config := SEVSNPConfig{
		Endpoint: "unix:///var/run/veid-enclave.sock",
	}

	svc, err := NewSEVSNPEnclaveServiceImpl(config)
	require.NoError(t, err)

	// Should fail before init
	_, err = svc.GetChipID()
	assert.Error(t, err)

	// Initialize
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Should succeed after init
	chipID, err := svc.GetChipID()
	assert.NoError(t, err)
	assert.Len(t, chipID, SNPChipIDSize)
}

func TestSEVSNPEnclaveServiceImpl_GetTCBVersion(t *testing.T) {
	config := SEVSNPConfig{
		Endpoint: "unix:///var/run/veid-enclave.sock",
	}

	svc, err := NewSEVSNPEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	tcb, err := svc.GetTCBVersion()
	assert.NoError(t, err)
	assert.NotNil(t, tcb)

	// Verify TCB fields are set
	assert.Greater(t, tcb.SNP, uint8(0))
	assert.Greater(t, tcb.Microcode, uint8(0))
}

func TestSEVSNPEnclaveServiceImpl_GetGuestPolicy(t *testing.T) {
	config := SEVSNPConfig{
		Endpoint:         "unix:///var/run/veid-enclave.sock",
		AllowDebugPolicy: false,
	}

	svc, err := NewSEVSNPEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	policy, err := svc.GetGuestPolicy()
	assert.NoError(t, err)
	assert.NotNil(t, policy)

	// Debug should be disabled for production
	assert.False(t, policy.Debug)
}

func TestSEVSNPEnclaveServiceImpl_VerifyMemoryEncryption(t *testing.T) {
	config := SEVSNPConfig{
		Endpoint: "unix:///var/run/veid-enclave.sock",
	}

	svc, err := NewSEVSNPEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Should succeed in POC (always returns nil)
	err = svc.VerifyMemoryEncryption()
	assert.NoError(t, err)
}

func TestSEVSNPEnclaveServiceImpl_VerifyLaunchMeasurement(t *testing.T) {
	config := SEVSNPConfig{
		Endpoint: "unix:///var/run/veid-enclave.sock",
	}

	svc, err := NewSEVSNPEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Get actual measurement
	measurement, err := svc.GetMeasurement()
	require.NoError(t, err)

	// Should pass with correct measurement
	err = svc.VerifyLaunchMeasurement(measurement)
	assert.NoError(t, err)

	// Should fail with wrong measurement
	wrongMeasurement := make([]byte, SNPLaunchDigestSize)
	err = svc.VerifyLaunchMeasurement(wrongMeasurement)
	assert.Error(t, err)

	// Should fail with wrong size
	err = svc.VerifyLaunchMeasurement([]byte("too short"))
	assert.Error(t, err)
}

func TestSEVSNPEnclaveServiceImpl_FetchVCEKCertificate(t *testing.T) {
	config := SEVSNPConfig{
		Endpoint: "unix:///var/run/veid-enclave.sock",
	}

	svc, err := NewSEVSNPEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Fetch VCEK (simulated)
	vcek, err := svc.FetchVCEKCertificate()
	assert.NoError(t, err)
	assert.NotEmpty(t, vcek)
}

func TestSEVSNPEnclaveServiceImpl_GenerateExtendedReport(t *testing.T) {
	config := SEVSNPConfig{
		Endpoint: "unix:///var/run/veid-enclave.sock",
	}

	svc, err := NewSEVSNPEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	reportData := []byte("extended_report_test_data")
	report, certChain, err := svc.GenerateExtendedReport(reportData)
	assert.NoError(t, err)
	assert.NotEmpty(t, report)
	assert.NotEmpty(t, certChain)

	// Should have VCEK -> ASK -> ARK chain
	assert.Len(t, certChain, 3)
}

func TestSEVSNPEnclaveServiceImpl_VerifyReport(t *testing.T) {
	config := SEVSNPConfig{
		Endpoint: "unix:///var/run/veid-enclave.sock",
	}

	svc, err := NewSEVSNPEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Generate a report
	report, err := svc.GenerateAttestation([]byte("test"))
	require.NoError(t, err)

	// Verify the report
	err = svc.VerifyReport(report)
	assert.NoError(t, err)

	// Invalid report should fail
	err = svc.VerifyReport([]byte("too short"))
	assert.Error(t, err)
}

func TestSEVSNPEnclaveServiceImpl_Shutdown(t *testing.T) {
	config := SEVSNPConfig{
		Endpoint: "unix:///var/run/veid-enclave.sock",
	}

	svc, err := NewSEVSNPEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Shutdown
	err = svc.Shutdown()
	assert.NoError(t, err)

	// Status should show not initialized
	status := svc.GetStatus()
	assert.False(t, status.Initialized)

	// Operations should fail after shutdown
	_, err = svc.GetMeasurement()
	assert.Error(t, err)
}

func TestSEVSNPEnclaveServiceImpl_PlatformInfo(t *testing.T) {
	config := SEVSNPConfig{
		Endpoint:         "unix:///var/run/veid-enclave.sock",
		AllowDebugPolicy: false,
	}

	svc, err := NewSEVSNPEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Platform type
	assert.Equal(t, PlatformSEVSNP, svc.GetPlatformType())

	// Secure if debug policy not allowed
	assert.True(t, svc.IsPlatformSecure())
}

// =============================================================================
// SNP Type Tests
// =============================================================================

func TestSNPGuestPolicy_ToUint64(t *testing.T) {
	// Production policy (no debug)
	prodPolicy := SNPGuestPolicy{
		ABIMinor: 0,
		ABIMajor: 1,
		SMT:      true,
		Debug:    false,
	}
	prodUint := prodPolicy.ToUint64()
	assert.NotEqual(t, uint64(0), prodUint&SNPPolicyNoDebug) // NoDebug flag set

	// Debug policy
	debugPolicy := SNPGuestPolicy{
		ABIMinor: 0,
		ABIMajor: 1,
		Debug:    true,
	}
	debugUint := debugPolicy.ToUint64()
	assert.Equal(t, uint64(0), debugUint&SNPPolicyNoDebug) // NoDebug flag NOT set
}

func TestSNPTCBVersion_ToUint64(t *testing.T) {
	tcb := SNPTCBVersion{
		BootLoader: 2,
		TEE:        0,
		SNP:        8,
		Microcode:  115,
	}

	uint64Val := tcb.ToUint64()
	assert.NotEqual(t, uint64(0), uint64Val)

	// Verify individual byte positions
	assert.Equal(t, uint8(2), uint8(uint64Val&0xFF))       // BootLoader at byte 0
	assert.Equal(t, uint8(0), uint8((uint64Val>>8)&0xFF))  // TEE at byte 1
}

func TestSNPLaunchDigest_String(t *testing.T) {
	var d SNPLaunchDigest
	d[0] = 0xAB
	d[1] = 0xCD

	str := d.String()
	assert.Contains(t, str, "abcd")
}

func TestSNPAttestationReport_Validate(t *testing.T) {
	// Valid report
	validReport := &SNPAttestationReport{
		Version: SNPReportVersion,
		Policy:  SNPGuestPolicy{Debug: false},
	}
	assert.NoError(t, validReport.Validate())

	// Old version
	oldReport := &SNPAttestationReport{
		Version: 1,
		Policy:  SNPGuestPolicy{Debug: false},
	}
	assert.Error(t, oldReport.Validate())

	// Debug enabled
	debugReport := &SNPAttestationReport{
		Version: SNPReportVersion,
		Policy:  SNPGuestPolicy{Debug: true},
	}
	assert.Error(t, debugReport.Validate())
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestBindResultToReport(t *testing.T) {
	result := &ScoringResult{
		RequestID:        "test-request",
		InputHash:        []byte("input_hash"),
		EnclaveSignature: []byte("signature"),
	}
	nonce := []byte("test_nonce_1234567890123456")

	reportData := BindResultToReport(result, nonce)
	assert.Len(t, reportData, SNPReportDataSize)

	// Nonce should be in second half
	extractedNonce := ExtractNonceFromReport(reportData)
	assert.NotEmpty(t, extractedNonce)
}

func TestExtractNonceFromReport(t *testing.T) {
	// Valid report data
	reportData := make([]byte, SNPReportDataSize)
	copy(reportData[32:], []byte("test_nonce_12345678901234"))

	nonce := ExtractNonceFromReport(reportData)
	assert.Len(t, nonce, 32)

	// Too short
	shortData := make([]byte, 10)
	assert.Nil(t, ExtractNonceFromReport(shortData))
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestSEVSNPEnclaveServiceImpl_ConcurrentScoring(t *testing.T) {
	config := SEVSNPConfig{
		Endpoint: "unix:///var/run/veid-enclave.sock",
	}

	svc, err := NewSEVSNPEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Run multiple concurrent scoring requests
	const numRequests = 10
	results := make(chan *ScoringResult, numRequests)
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(idx int) {
			request := &ScoringRequest{
				RequestID:      string(rune('a' + idx)),
				Ciphertext:     []byte("concurrent_snp_test_data"),
				WrappedKey:     []byte("wrapped_key"),
				Nonce:          []byte("nonce"),
				ScopeID:        "scope",
				AccountAddress: "addr",
			}
			result, err := svc.Score(context.Background(), request)
			if err != nil {
				errors <- err
			} else {
				results <- result
			}
		}(i)
	}

	// Collect results (some may fail due to concurrency limits)
	successCount := 0
	for i := 0; i < numRequests; i++ {
		select {
		case <-results:
			successCount++
		case <-errors:
			// Expected for requests exceeding concurrent limit
		}
	}

	// At least max concurrent should succeed
	assert.GreaterOrEqual(t, successCount, 1)
}
