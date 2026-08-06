package device

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

func TestServiceVerifyAttestation(t *testing.T) {
	service := NewService(map[veidtypes.DeviceAttestationProvider]AttestationVerifier{
		veidtypes.DeviceAttestationProviderPlayIntegrity: MockVerifier{},
	})

	req := AttestationRequest{
		AccountAddress:     "virt1device",
		Platform:           veidtypes.DevicePlatformAndroid,
		Provider:           veidtypes.DeviceAttestationProviderPlayIntegrity,
		AppID:              "com.virtengine.veid",
		AppVersion:         "1.0.0",
		DeviceModel:        "Pixel",
		OSVersion:          "Android 16",
		Nonce:              "nonce",
		Attestation:        []byte("attestation"),
		RequestedAt:        time.Now().UTC(),
		RequireAttestation: true,
	}

	result, err := service.VerifyAttestation(context.Background(), req)
	require.NoError(t, err)
	require.True(t, result.Verified)
	require.Equal(t, AttestationStatusVerified, result.Status)

	payloadHash := sha256.Sum256([]byte("attestation-payload"))
	record := BuildDeviceAttestationRecord(result, "vault://device/attestation", payloadHash[:])
	require.NoError(t, record.Validate())
}

func TestServiceVerifyAttestationFallback(t *testing.T) {
	service := NewService(map[veidtypes.DeviceAttestationProvider]AttestationVerifier{})

	req := AttestationRequest{
		AccountAddress:     "virt1device",
		Platform:           veidtypes.DevicePlatformIOS,
		Provider:           veidtypes.DeviceAttestationProviderAppAttest,
		AppID:              "com.virtengine.veid",
		AppVersion:         "1.0.0",
		DeviceModel:        "iPhone",
		OSVersion:          "iOS 19",
		Nonce:              "nonce",
		RequestedAt:        time.Now().UTC(),
		AllowFallback:      true,
		RequireAttestation: false,
	}

	result, err := service.VerifyAttestation(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, AttestationStatusUnsupported, result.Status)
	require.False(t, result.Verified)
}

func TestServiceRejectsUnboundAttestationResult(t *testing.T) {
	req := AttestationRequest{
		AccountAddress: "virt1device",
		Platform:       veidtypes.DevicePlatformAndroid,
		Provider:       veidtypes.DeviceAttestationProviderPlayIntegrity,
		AppID:          "com.virtengine.veid",
		AppVersion:     "1.0.0",
		DeviceModel:    "Pixel",
		OSVersion:      "Android 16",
		Nonce:          "expected-nonce",
		RequestedAt:    time.Now().UTC(),
	}
	valid, err := (MockVerifier{}).Verify(context.Background(), req)
	require.NoError(t, err)

	for name, mutate := range map[string]func(*AttestationResult){
		"nonce":     func(result *AttestationResult) { result.Nonce = "other-nonce" },
		"provider":  func(result *AttestationResult) { result.Provider = veidtypes.DeviceAttestationProviderAppAttest },
		"status":    func(result *AttestationResult) { result.Status = AttestationStatusFailed; result.Verified = true },
		"timestamp": func(result *AttestationResult) { result.AttestedAt = time.Time{} },
	} {
		t.Run(name, func(t *testing.T) {
			result := valid
			mutate(&result)
			service := NewService(map[veidtypes.DeviceAttestationProvider]AttestationVerifier{
				veidtypes.DeviceAttestationProviderPlayIntegrity: attestationVerifierFunc(func(context.Context, AttestationRequest) (AttestationResult, error) {
					return result, nil
				}),
			})

			_, err := service.VerifyAttestation(context.Background(), req)
			require.Error(t, err)
		})
	}
}

func TestServiceBindsDefaultProviderBeforeVerification(t *testing.T) {
	service := NewService(map[veidtypes.DeviceAttestationProvider]AttestationVerifier{
		veidtypes.DeviceAttestationProviderPlayIntegrity: MockVerifier{},
	})
	req := AttestationRequest{
		AccountAddress: "virt1device",
		Platform:       veidtypes.DevicePlatformAndroid,
		AppID:          "com.virtengine.veid",
		AppVersion:     "1.0.0",
		DeviceModel:    "Pixel",
		OSVersion:      "Android 16",
		Nonce:          "nonce",
		RequestedAt:    time.Now().UTC(),
	}

	result, err := service.VerifyAttestation(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, veidtypes.DeviceAttestationProviderPlayIntegrity, result.Provider)
}

func TestServiceRejectsInvalidAttestationRequests(t *testing.T) {
	service := NewService(map[veidtypes.DeviceAttestationProvider]AttestationVerifier{
		veidtypes.DeviceAttestationProviderPlayIntegrity: MockVerifier{},
	})
	valid := validAttestationRequest()

	for name, mutate := range map[string]func(*AttestationRequest){
		"missing context":          func(request *AttestationRequest) {},
		"cancelled context":        func(request *AttestationRequest) {},
		"invalid platform":         func(request *AttestationRequest) { request.Platform = "unknown" },
		"incompatible provider":    func(request *AttestationRequest) { request.Provider = veidtypes.DeviceAttestationProviderAppAttest },
		"missing binding":          func(request *AttestationRequest) { request.AppID = "" },
		"missing request time":     func(request *AttestationRequest) { request.RequestedAt = time.Time{} },
		"missing required payload": func(request *AttestationRequest) { request.Attestation = nil },
		"contradictory fallback":   func(request *AttestationRequest) { request.AllowFallback = true },
	} {
		t.Run(name, func(t *testing.T) {
			req := valid
			mutate(&req)
			ctx := context.Background()
			if name == "missing context" {
				ctx = nil
			}
			if name == "cancelled context" {
				cancelled, cancel := context.WithCancel(ctx)
				cancel()
				ctx = cancelled
			}
			_, err := service.VerifyAttestation(ctx, req)
			require.Error(t, err)
		})
	}
}

func TestServiceRejectsInconsistentAttestationResult(t *testing.T) {
	req := validAttestationRequest()
	valid, err := (MockVerifier{}).Verify(context.Background(), req)
	require.NoError(t, err)

	for name, mutate := range map[string]func(*AttestationResult){
		"unknown integrity": func(result *AttestationResult) { result.IntegrityLevel = veidtypes.DeviceIntegrityUnknown },
		"hardware mismatch": func(result *AttestationResult) {
			result.IntegrityLevel = veidtypes.DeviceIntegrityHardwareBacked
			result.HardwareBacked = false
		},
		"failure on success": func(result *AttestationResult) { result.FailureReason = "unexpected" },
		"score out of range": func(result *AttestationResult) { result.IntegrityScore = 10001 },
		"failed without reason": func(result *AttestationResult) {
			result.Status = AttestationStatusFailed
			result.Verified = false
			result.FailureReason = ""
		},
	} {
		t.Run(name, func(t *testing.T) {
			result := valid
			mutate(&result)
			service := NewService(map[veidtypes.DeviceAttestationProvider]AttestationVerifier{
				veidtypes.DeviceAttestationProviderPlayIntegrity: attestationVerifierFunc(func(context.Context, AttestationRequest) (AttestationResult, error) {
					return result, nil
				}),
			})
			_, err := service.VerifyAttestation(context.Background(), req)
			require.Error(t, err)
		})
	}
}

func TestBuildDeviceAttestationRecordCopiesEvidence(t *testing.T) {
	result, err := (MockVerifier{}).Verify(context.Background(), validAttestationRequest())
	require.NoError(t, err)
	result.Metadata = map[string]string{"provider": "play"}
	payloadHash := sha256.Sum256([]byte("attestation-payload"))
	record := BuildDeviceAttestationRecord(result, "vault://device/attestation", payloadHash[:])
	payloadHash[0] ^= 0xff
	result.Metadata["provider"] = "mutated"

	require.NotEqual(t, payloadHash[:], record.PayloadHash)
	require.Equal(t, "play", record.Metadata["provider"])
}

func validAttestationRequest() AttestationRequest {
	return AttestationRequest{
		AccountAddress:     "virt1device",
		Platform:           veidtypes.DevicePlatformAndroid,
		Provider:           veidtypes.DeviceAttestationProviderPlayIntegrity,
		AppID:              "com.virtengine.veid",
		AppVersion:         "1.0.0",
		DeviceModel:        "Pixel",
		OSVersion:          "Android 16",
		Nonce:              "nonce",
		Attestation:        []byte("attestation"),
		RequestedAt:        time.Now().UTC(),
		RequireAttestation: true,
	}
}

type attestationVerifierFunc func(context.Context, AttestationRequest) (AttestationResult, error)

func (f attestationVerifierFunc) Verify(ctx context.Context, req AttestationRequest) (AttestationResult, error) {
	return f(ctx, req)
}
