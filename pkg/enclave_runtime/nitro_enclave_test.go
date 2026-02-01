package enclave_runtime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Nitro Enclave Service Implementation Tests
// =============================================================================

func TestNitroEnclaveConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  NitroEnclaveConfig
		wantErr bool
	}{
		{
			name: "valid config with all fields",
			config: NitroEnclaveConfig{
				EnclaveImagePath: "/opt/enclaves/veid.eif",
				CPUCount:         2,
				MemoryMB:         2048,
				CID:              16,
				VsockPort:        5000,
			},
			wantErr: false,
		},
		{
			name: "valid config with defaults",
			config: NitroEnclaveConfig{
				EnclaveImagePath: "/opt/enclaves/veid.eif",
			},
			wantErr: false,
		},
		{
			name: "missing enclave path",
			config: NitroEnclaveConfig{
				CPUCount: 2,
				MemoryMB: 2048,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Check defaults are applied
				assert.Greater(t, tt.config.CPUCount, 0)
				assert.Greater(t, tt.config.MemoryMB, 0)
				assert.NotZero(t, tt.config.CID)
				assert.NotZero(t, tt.config.VsockPort)
			}
		})
	}
}

func TestNitroEnclaveServiceImpl_NewService(t *testing.T) {
	tests := []struct {
		name    string
		config  NitroEnclaveConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: NitroEnclaveConfig{
				EnclaveImagePath: "/opt/enclaves/veid.eif",
				CPUCount:         2,
				MemoryMB:         2048,
				DebugMode:        false,
			},
			wantErr: false,
		},
		{
			name: "debug mode allowed for testing",
			config: NitroEnclaveConfig{
				EnclaveImagePath: "/opt/enclaves/veid.eif",
				DebugMode:        true,
			},
			wantErr: false,
		},
		{
			name: "missing enclave image path",
			config: NitroEnclaveConfig{
				CPUCount: 2,
				MemoryMB: 2048,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewNitroEnclaveServiceImpl(tt.config)
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

func TestNitroEnclaveServiceImpl_Initialize(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
		CPUCount:         2,
		MemoryMB:         2048,
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
	require.NoError(t, err)

	// Initialize
	err = svc.Initialize(DefaultRuntimeConfig())
	assert.NoError(t, err)

	// Check status
	status := svc.GetStatus()
	assert.True(t, status.Initialized)
	assert.True(t, status.Available)
	assert.Equal(t, uint64(1), status.CurrentEpoch)

	// Check enclave state
	assert.Equal(t, NitroEnclaveStateRunning, svc.GetEnclaveState())

	// Double initialization should fail
	err = svc.Initialize(DefaultRuntimeConfig())
	assert.Error(t, err)
}

func TestNitroEnclaveServiceImpl_Score(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
		CPUCount:         2,
		MemoryMB:         2048,
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	request := &ScoringRequest{
		RequestID:      "nitro-test-request-1",
		Ciphertext:     []byte("nitro_encrypted_identity_data"),
		WrappedKey:     []byte("nitro_wrapped_key"),
		Nonce:          []byte("nitro_nonce_12345"),
		ScopeID:        "nitro-scope-123",
		AccountAddress: "virtengine1nitrotest",
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

	// Verify Nitro-specific fields
	assert.Contains(t, result.ReasonCodes[0], "nitro_")

	// Measurement should be SHA-384 size (PCR digest)
	assert.Len(t, result.MeasurementHash, NitroPCRDigestSize)
}

func TestNitroEnclaveServiceImpl_ScoreNotInitialized(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
	require.NoError(t, err)

	// Attempt to score without initialization
	request := &ScoringRequest{
		RequestID:      "test",
		Ciphertext:     []byte("data"),
		WrappedKey:     []byte("key"),
		Nonce:          []byte("nonce"),
		ScopeID:        "scope",
		AccountAddress: "addr",
	}

	_, err = svc.Score(context.Background(), request)
	assert.Equal(t, ErrEnclaveNotInitialized, err)
}

func TestNitroEnclaveServiceImpl_ScoreValidation(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	tests := []struct {
		name    string
		request *ScoringRequest
		wantErr bool
	}{
		{
			name: "missing request ID",
			request: &ScoringRequest{
				Ciphertext:     []byte("data"),
				WrappedKey:     []byte("key"),
				Nonce:          []byte("nonce"),
				ScopeID:        "scope",
				AccountAddress: "addr",
			},
			wantErr: true,
		},
		{
			name: "missing ciphertext",
			request: &ScoringRequest{
				RequestID:      "req1",
				WrappedKey:     []byte("key"),
				Nonce:          []byte("nonce"),
				ScopeID:        "scope",
				AccountAddress: "addr",
			},
			wantErr: true,
		},
		{
			name: "missing scope ID",
			request: &ScoringRequest{
				RequestID:      "req1",
				Ciphertext:     []byte("data"),
				WrappedKey:     []byte("key"),
				Nonce:          []byte("nonce"),
				AccountAddress: "addr",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.Score(context.Background(), tt.request)
			require.NoError(t, err) // Score returns error in result, not as error
			if tt.wantErr {
				assert.False(t, result.IsSuccess())
				assert.NotEmpty(t, result.Error)
			}
		})
	}
}

func TestNitroEnclaveServiceImpl_GetMeasurement(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
	require.NoError(t, err)

	// Should fail before init
	_, err = svc.GetMeasurement()
	assert.Error(t, err)

	// Initialize
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Should succeed after init (PCR digest - SHA-384)
	measurement, err := svc.GetMeasurement()
	assert.NoError(t, err)
	assert.Len(t, measurement, NitroPCRDigestSize)
}

func TestNitroEnclaveServiceImpl_GetPCR(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
	require.NoError(t, err)

	// Should fail before init
	_, err = svc.GetPCR(NitroPCR0EIF)
	assert.Error(t, err)

	// Initialize
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Get PCR0 (EIF measurement)
	pcr0, err := svc.GetPCR(NitroPCR0EIF)
	assert.NoError(t, err)
	assert.False(t, pcr0.IsZero())

	// Get PCR1 (kernel)
	pcr1, err := svc.GetPCR(NitroPCR1Kernel)
	assert.NoError(t, err)
	assert.False(t, pcr1.IsZero())

	// Get PCR2 (application)
	pcr2, err := svc.GetPCR(NitroPCR2App)
	assert.NoError(t, err)
	assert.False(t, pcr2.IsZero())

	// Get PCR4 (instance ID)
	pcr4, err := svc.GetPCR(NitroPCR4InstanceID)
	assert.NoError(t, err)
	assert.False(t, pcr4.IsZero())

	// Invalid PCR index
	_, err = svc.GetPCR(NitroPCRIndex(20))
	assert.Error(t, err)
}

func TestNitroEnclaveServiceImpl_GetPCRSet(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	pcrSet, err := svc.GetPCRSet()
	assert.NoError(t, err)
	assert.NotNil(t, pcrSet)

	// Core PCRs should be set
	assert.False(t, pcrSet.Get(NitroPCR0EIF).IsZero())
	assert.False(t, pcrSet.Get(NitroPCR1Kernel).IsZero())
	assert.False(t, pcrSet.Get(NitroPCR2App).IsZero())

	// PCR3 (IAM role) should be zero in simulation
	assert.True(t, pcrSet.Get(NitroPCR3IAMRole).IsZero())
}

func TestNitroAttestationDocument_Validate(t *testing.T) {
	tests := []struct {
		name    string
		doc     NitroAttestationDocument
		wantErr bool
	}{
		{
			name: "valid document",
			doc: NitroAttestationDocument{
				ModuleID:    "veid-nitro-test",
				Timestamp:   1706543210000,
				Digest:      "SHA384",
				Certificate: []byte("cert_data"),
				PCRs: func() NitroPCRSet {
					var pcrs NitroPCRSet
					pcrs.PCRs[0] = NitroPCR{1, 2, 3} // Non-zero PCR0
					return pcrs
				}(),
			},
			wantErr: false,
		},
		{
			name: "missing module ID",
			doc: NitroAttestationDocument{
				Timestamp:   1706543210000,
				Digest:      "SHA384",
				Certificate: []byte("cert_data"),
			},
			wantErr: true,
		},
		{
			name: "missing timestamp",
			doc: NitroAttestationDocument{
				ModuleID:    "test",
				Digest:      "SHA384",
				Certificate: []byte("cert_data"),
			},
			wantErr: true,
		},
		{
			name: "wrong digest algorithm",
			doc: NitroAttestationDocument{
				ModuleID:    "test",
				Timestamp:   1706543210000,
				Digest:      "SHA256",
				Certificate: []byte("cert_data"),
			},
			wantErr: true,
		},
		{
			name: "missing certificate",
			doc: NitroAttestationDocument{
				ModuleID:  "test",
				Timestamp: 1706543210000,
				Digest:    "SHA384",
			},
			wantErr: true,
		},
		{
			name: "user data too large",
			doc: NitroAttestationDocument{
				ModuleID:    "test",
				Timestamp:   1706543210000,
				Digest:      "SHA384",
				Certificate: []byte("cert"),
				UserData:    make([]byte, NitroMaxUserData+1),
				PCRs: func() NitroPCRSet {
					var pcrs NitroPCRSet
					pcrs.PCRs[0] = NitroPCR{1}
					return pcrs
				}(),
			},
			wantErr: true,
		},
		{
			name: "zero PCR0",
			doc: NitroAttestationDocument{
				ModuleID:    "test",
				Timestamp:   1706543210000,
				Digest:      "SHA384",
				Certificate: []byte("cert"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.doc.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNitroEnclaveServiceImpl_GenerateAttestation(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Generate attestation with user data
	userData := []byte("test_nonce_for_nitro_attestation")
	attestation, err := svc.GenerateAttestation(userData)
	assert.NoError(t, err)
	assert.NotEmpty(t, attestation)

	// Attestation should contain proper structure
	assert.Greater(t, len(attestation), 100)

	// Should start with magic
	assert.Equal(t, "NITRO_ATTEST_V1", string(attestation[:15]))

	// User data too large should fail
	largeData := make([]byte, NitroMaxUserData+1)
	_, err = svc.GenerateAttestation(largeData)
	assert.Error(t, err)
}

func TestNitroEnclaveServiceImpl_KeyRotation(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
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

func TestNitroEnclaveServiceImpl_KeyRotationNotInitialized(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
	require.NoError(t, err)

	err = svc.RotateKeys()
	assert.Equal(t, ErrEnclaveNotInitialized, err)
}

func TestNitroEnclaveServiceImpl_GetEnclaveID(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
	require.NoError(t, err)

	// Should fail before init
	_, err = svc.GetEnclaveID()
	assert.Error(t, err)

	// Initialize
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Should succeed after init
	enclaveID, err := svc.GetEnclaveID()
	assert.NoError(t, err)
	assert.NotEmpty(t, enclaveID)
	assert.Contains(t, enclaveID, "i-") // Simulated format
}

func TestNitroEnclaveServiceImpl_VerifyPCRs(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Get actual PCR0
	pcr0, _ := svc.GetPCR(NitroPCR0EIF)

	// Verification should pass without AllowedPCR values
	err = svc.VerifyPCRs()
	assert.NoError(t, err)

	// Create service with expected PCR0
	configWithPCR := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
		AllowedPCR0:      pcr0[:],
	}
	svc2, err := NewNitroEnclaveServiceImpl(configWithPCR)
	require.NoError(t, err)
	require.NoError(t, svc2.Initialize(DefaultRuntimeConfig()))

	// Should pass with matching PCR0
	err = svc2.VerifyPCRs()
	assert.NoError(t, err)

	// Create service with wrong PCR0
	configWrongPCR := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
		AllowedPCR0:      []byte("wrong_pcr_value_that_does_not_match"),
	}
	svc3, err := NewNitroEnclaveServiceImpl(configWrongPCR)
	require.NoError(t, err)
	require.NoError(t, svc3.Initialize(DefaultRuntimeConfig()))

	// Should fail with wrong PCR0
	err = svc3.VerifyPCRs()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PCR0 mismatch")
}

func TestNitroEnclaveServiceImpl_VerifyMeasurement(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Get measurement
	measurement, err := svc.GetMeasurement()
	require.NoError(t, err)

	// Should verify correctly
	assert.True(t, svc.VerifyMeasurement(measurement))

	// Should fail with wrong measurement
	wrongMeasurement := make([]byte, len(measurement))
	copy(wrongMeasurement, measurement)
	wrongMeasurement[0] ^= 0xFF
	assert.False(t, svc.VerifyMeasurement(wrongMeasurement))
}

func TestNitroEnclaveServiceImpl_Shutdown(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Verify initialized
	status := svc.GetStatus()
	assert.True(t, status.Initialized)

	// Shutdown
	err = svc.Shutdown()
	assert.NoError(t, err)

	// Verify shut down
	status = svc.GetStatus()
	assert.False(t, status.Initialized)
	assert.False(t, status.Available)
	assert.Equal(t, NitroEnclaveStateStopped, svc.GetEnclaveState())

	// Operations should fail after shutdown
	_, err = svc.GetMeasurement()
	assert.Equal(t, ErrEnclaveNotInitialized, err)

	// Shutdown again should be no-op
	err = svc.Shutdown()
	assert.NoError(t, err)
}

func TestNitroEnclaveServiceImpl_PlatformType(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
	require.NoError(t, err)

	assert.Equal(t, PlatformNitro, svc.GetPlatformType())
}

func TestNitroEnclaveServiceImpl_IsPlatformSecure(t *testing.T) {
	// Normal mode is secure
	configSecure := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
		DebugMode:        false,
	}
	svcSecure, err := NewNitroEnclaveServiceImpl(configSecure)
	require.NoError(t, err)
	assert.True(t, svcSecure.IsPlatformSecure())

	// Debug mode is not secure
	configDebug := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
		DebugMode:        true,
	}
	svcDebug, err := NewNitroEnclaveServiceImpl(configDebug)
	require.NoError(t, err)
	assert.False(t, svcDebug.IsPlatformSecure())
}

func TestNitroPCR_IsZero(t *testing.T) {
	var zeroPCR NitroPCR
	assert.True(t, zeroPCR.IsZero())

	nonZeroPCR := NitroPCR{1, 2, 3}
	assert.False(t, nonZeroPCR.IsZero())
}

func TestNitroPCR_String(t *testing.T) {
	pcr := NitroPCR{0xab, 0xcd, 0xef}
	str := pcr.String()
	assert.Contains(t, str, "abcdef")
}

func TestNitroPCRSet_GetSet(t *testing.T) {
	var pcrSet NitroPCRSet

	// Set a value
	testPCR := NitroPCR{1, 2, 3, 4, 5}
	pcrSet.Set(NitroPCR0EIF, testPCR)

	// Get the value
	retrieved := pcrSet.Get(NitroPCR0EIF)
	assert.Equal(t, testPCR, retrieved)

	// Get non-existent (should return zero)
	invalid := pcrSet.Get(NitroPCRIndex(99))
	assert.True(t, invalid.IsZero())

	// Set invalid index (should be no-op)
	pcrSet.Set(NitroPCRIndex(99), testPCR)
}

func TestNitroPCRSet_Digest(t *testing.T) {
	var pcrSet1 NitroPCRSet
	var pcrSet2 NitroPCRSet

	// Same sets should produce same digest
	assert.Equal(t, pcrSet1.Digest(), pcrSet2.Digest())

	// Different sets should produce different digest
	pcrSet1.Set(NitroPCR0EIF, NitroPCR{1, 2, 3})
	assert.NotEqual(t, pcrSet1.Digest(), pcrSet2.Digest())

	// Digest should be SHA-384 size
	digest := pcrSet1.Digest()
	assert.Len(t, digest, 48)
}

func TestNitroEnclaveState_String(t *testing.T) {
	tests := []struct {
		state    NitroEnclaveState
		expected string
	}{
		{NitroEnclaveStateStopped, "stopped"},
		{NitroEnclaveStateStarting, "starting"},
		{NitroEnclaveStateRunning, "running"},
		{NitroEnclaveStateStopping, "stopping"},
		{NitroEnclaveStateFailed, "failed"},
		{NitroEnclaveState(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

func TestVerifyNitroAttestation(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Generate attestation
	attestation, err := svc.GenerateAttestation([]byte("test"))
	require.NoError(t, err)

	// Verify without expected PCR0 (should pass)
	err = VerifyNitroAttestation(attestation, nil)
	assert.NoError(t, err)

	// Verify with wrong PCR0 (should fail)
	wrongPCR0 := make([]byte, NitroPCRDigestSize)
	err = VerifyNitroAttestation(attestation, wrongPCR0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PCR0 mismatch")

	// Invalid attestation format
	err = VerifyNitroAttestation([]byte("short"), nil)
	assert.Error(t, err)

	// Wrong magic
	wrongMagic := make([]byte, 100)
	err = VerifyNitroAttestation(wrongMagic, nil)
	assert.Error(t, err)
}

func TestExtractPCRsFromAttestation(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Generate attestation
	attestation, err := svc.GenerateAttestation([]byte("test"))
	require.NoError(t, err)

	// Extract PCRs
	pcrSet, err := ExtractPCRsFromAttestation(attestation)
	assert.NoError(t, err)
	assert.NotNil(t, pcrSet)

	// PCR0 should be non-zero
	assert.False(t, pcrSet.Get(NitroPCR0EIF).IsZero())

	// Invalid attestation
	_, err = ExtractPCRsFromAttestation([]byte("short"))
	assert.Error(t, err)
}

func TestBindNitroResultToReport(t *testing.T) {
	result := &ScoringResult{
		RequestID:        "test-request",
		InputHash:        []byte("input_hash"),
		EnclaveSignature: []byte("signature"),
	}

	nonce := []byte("test_nonce_12345")
	userData := BindNitroResultToReport(result, nonce)

	// User data should contain digest + nonce
	assert.NotEmpty(t, userData)
	assert.LessOrEqual(t, len(userData), 64)

	// Long nonce should be truncated
	longNonce := make([]byte, 100)
	userData2 := BindNitroResultToReport(result, longNonce)
	assert.Equal(t, 64, len(userData2))
}

func TestNitroEnclaveServiceImpl_ConcurrentScoring(t *testing.T) {
	config := NitroEnclaveConfig{
		EnclaveImagePath: "/opt/enclaves/veid.eif",
	}

	svc, err := NewNitroEnclaveServiceImpl(config)
	require.NoError(t, err)

	runtimeConfig := DefaultRuntimeConfig()
	runtimeConfig.MaxConcurrentRequests = 4
	require.NoError(t, svc.Initialize(runtimeConfig))

	// Run multiple scoring requests concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			request := &ScoringRequest{
				RequestID:      "concurrent-" + string(rune('0'+idx)),
				Ciphertext:     []byte("data"),
				WrappedKey:     []byte("key"),
				Nonce:          []byte("nonce"),
				ScopeID:        "scope",
				AccountAddress: "addr",
			}
			_, _ = svc.Score(context.Background(), request)
			done <- true
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check total processed
	status := svc.GetStatus()
	assert.Greater(t, status.TotalProcessed, uint64(0))
}

func TestNitroConstants(t *testing.T) {
	// Verify constants are set correctly
	assert.Equal(t, uint32(16), uint32(NitroDefaultCID))
	assert.Equal(t, uint32(5000), uint32(NitroDefaultVsockPort))
	assert.Equal(t, 16, NitroPCRCount)
	assert.Equal(t, 48, NitroPCRDigestSize)
	assert.Equal(t, 1024, NitroMaxUserData)
	assert.Equal(t, 64, NitroModuleIDSize)
}

