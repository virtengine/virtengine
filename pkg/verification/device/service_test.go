package device

import (
	"context"
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
		RequestedAt:        time.Now().UTC(),
		RequireAttestation: true,
	}

	result, err := service.VerifyAttestation(context.Background(), req)
	require.NoError(t, err)
	require.True(t, result.Verified)
	require.Equal(t, AttestationStatusVerified, result.Status)

	record := BuildDeviceAttestationRecord(result, "vault://device/attestation", []byte("hash"))
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
