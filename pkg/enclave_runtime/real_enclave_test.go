package enclave_runtime

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Platform Type Tests
// =============================================================================

func TestPlatformType_String(t *testing.T) {
	tests := []struct {
		platform PlatformType
		expected string
	}{
		{PlatformSimulated, "simulated"},
		{PlatformSGX, "sgx"},
		{PlatformSEVSNP, "sev-snp"},
		{PlatformNitro, "nitro"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.platform.String())
		})
	}
}

func TestPlatformType_IsSecure(t *testing.T) {
	tests := []struct {
		platform PlatformType
		secure   bool
	}{
		{PlatformSimulated, false},
		{PlatformSGX, true},
		{PlatformSEVSNP, true},
		{PlatformNitro, true},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.platform), func(t *testing.T) {
			assert.Equal(t, tt.secure, tt.platform.IsSecure())
		})
	}
}

// =============================================================================
// Attestation Report Tests
// =============================================================================

func TestAttestationReport_Validate(t *testing.T) {
	tests := []struct {
		name    string
		report  AttestationReport
		wantErr bool
	}{
		{
			name: "valid report",
			report: AttestationReport{
				Platform:    PlatformSGX,
				Measurement: []byte("measurement"),
				Signature:   []byte("signature"),
				Timestamp:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing platform",
			report: AttestationReport{
				Measurement: []byte("measurement"),
				Signature:   []byte("signature"),
			},
			wantErr: true,
		},
		{
			name: "missing measurement",
			report: AttestationReport{
				Platform:  PlatformSGX,
				Signature: []byte("signature"),
			},
			wantErr: true,
		},
		{
			name: "missing signature",
			report: AttestationReport{
				Platform:    PlatformSGX,
				Measurement: []byte("measurement"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.report.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// =============================================================================
// Enclave Service Factory Tests
// =============================================================================

func TestCreateEnclaveService_Simulated(t *testing.T) {
	config := EnclaveConfig{
		Platform:      PlatformSimulated,
		RuntimeConfig: DefaultRuntimeConfig(),
	}

	svc, err := CreateEnclaveService(config)
	require.NoError(t, err)
	require.NotNil(t, svc)

	// Verify it's a simulated service
	status := svc.GetStatus()
	assert.True(t, status.Initialized)
}

func TestCreateEnclaveService_DefaultIsSimulated(t *testing.T) {
	config := EnclaveConfig{
		Platform:      "", // Empty defaults to simulated
		RuntimeConfig: DefaultRuntimeConfig(),
	}

	svc, err := CreateEnclaveService(config)
	require.NoError(t, err)
	require.NotNil(t, svc)
}

func TestCreateEnclaveService_SGX(t *testing.T) {
	config := EnclaveConfig{
		Platform: PlatformSGX,
		SGXConfig: &SGXEnclaveConfig{
			EnclavePath: "/path/to/enclave.signed.so",
			DCAPEnabled: true,
		},
	}

	svc, err := CreateEnclaveService(config)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	// Verify it's the right type
	_, ok := svc.(*SGXEnclaveServiceImpl)
	assert.True(t, ok, "service should be SGXEnclaveServiceImpl")
}

func TestCreateEnclaveService_SEVSNP(t *testing.T) {
	config := EnclaveConfig{
		Platform: PlatformSEVSNP,
		SEVSNPConfig: &SEVSNPConfig{
			Endpoint: "unix:///var/run/enclave.sock",
		},
	}

	svc, err := CreateEnclaveService(config)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	// Verify it's the right type
	_, ok := svc.(*SEVSNPEnclaveServiceImpl)
	assert.True(t, ok, "service should be SEVSNPEnclaveServiceImpl")
}

func TestCreateEnclaveService_NitroNotImplemented(t *testing.T) {
	config := EnclaveConfig{
		Platform: PlatformNitro,
		NitroConfig: &NitroConfig{
			EnclaveImageURI: "enclave.eif",
			CPUCount:        2,
			MemoryMB:        512,
		},
	}

	svc, err := CreateEnclaveService(config)
	assert.Error(t, err)
	assert.Nil(t, svc)
	assert.Contains(t, err.Error(), "not yet implemented")
}

func TestCreateEnclaveService_MissingConfig(t *testing.T) {
	tests := []struct {
		name     string
		platform PlatformType
	}{
		{"SGX without config", PlatformSGX},
		{"SEV-SNP without config", PlatformSEVSNP},
		{"Nitro without config", PlatformNitro},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := EnclaveConfig{
				Platform: tt.platform,
			}

			svc, err := CreateEnclaveService(config)
			assert.Error(t, err)
			assert.Nil(t, svc)
			assert.Contains(t, err.Error(), "configuration required")
		})
	}
}

func TestCreateEnclaveService_UnknownPlatform(t *testing.T) {
	config := EnclaveConfig{
		Platform: "unknown-platform",
	}

	svc, err := CreateEnclaveService(config)
	assert.Error(t, err)
	assert.Nil(t, svc)
	assert.Contains(t, err.Error(), "unknown platform")
}

// =============================================================================
// Attestation Verifier Tests
// =============================================================================

func TestSimpleAttestationVerifier_AddAndCheck(t *testing.T) {
	verifier := NewSimpleAttestationVerifier()

	measurement := []byte("test-measurement-hash")

	// Initially not allowed
	assert.False(t, verifier.IsMeasurementAllowed(measurement))

	// Add to allowlist
	verifier.AddAllowedMeasurement(measurement)

	// Now allowed
	assert.True(t, verifier.IsMeasurementAllowed(measurement))
}

func TestSimpleAttestationVerifier_VerifyReport(t *testing.T) {
	verifier := NewSimpleAttestationVerifier()
	ctx := context.Background()

	measurement := []byte("valid-measurement")
	verifier.AddAllowedMeasurement(measurement)

	t.Run("valid report", func(t *testing.T) {
		report := &AttestationReport{
			Platform:    PlatformSGX,
			Measurement: measurement,
			Signature:   []byte("signature"),
			Timestamp:   time.Now(),
		}

		err := verifier.VerifyReport(ctx, report)
		assert.NoError(t, err)
	})

	t.Run("unknown measurement", func(t *testing.T) {
		report := &AttestationReport{
			Platform:    PlatformSGX,
			Measurement: []byte("unknown-measurement"),
			Signature:   []byte("signature"),
			Timestamp:   time.Now(),
		}

		err := verifier.VerifyReport(ctx, report)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in allowlist")
	})

	t.Run("invalid report", func(t *testing.T) {
		report := &AttestationReport{
			// Missing required fields
		}

		err := verifier.VerifyReport(ctx, report)
		assert.Error(t, err)
	})
}

// =============================================================================
// Platform Stub Tests
// =============================================================================

func TestSGXEnclaveService_GetPlatformType(t *testing.T) {
	svc := &SGXEnclaveService{}
	assert.Equal(t, PlatformSGX, svc.GetPlatformType())
	assert.True(t, svc.IsPlatformSecure())
}

func TestSEVSNPEnclaveService_GetPlatformType(t *testing.T) {
	svc := &SEVSNPEnclaveService{}
	assert.Equal(t, PlatformSEVSNP, svc.GetPlatformType())
	assert.True(t, svc.IsPlatformSecure())
}

func TestNitroEnclaveService_GetPlatformType(t *testing.T) {
	svc := &NitroEnclaveService{}
	assert.Equal(t, PlatformNitro, svc.GetPlatformType())
	assert.True(t, svc.IsPlatformSecure())
}
