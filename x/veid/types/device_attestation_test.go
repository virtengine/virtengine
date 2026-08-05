package types

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDeviceAttestationRecordValidate(t *testing.T) {
	payloadHash := sha256.Sum256([]byte("attestation-payload"))
	valid := DeviceAttestationRecord{
		AttestationID:  "attest-001",
		Platform:       DevicePlatformAndroid,
		Provider:       DeviceAttestationProviderPlayIntegrity,
		Nonce:          "nonce-123",
		AttestedAt:     time.Now().UTC(),
		IntegrityLevel: DeviceIntegrityStrong,
		DeviceModel:    "Pixel 9 Pro",
		OSVersion:      "Android 16",
		AppVersion:     "1.2.3",
		AppID:          "com.virtengine.veid",
		HardwareBacked: true,
		Supported:      true,
		Verified:       true,
		PayloadHash:    payloadHash[:],
		VaultRef:       "vault://attestations/attest-001",
	}

	require.NoError(t, valid.Validate())

	invalidProvider := valid
	invalidProvider.Provider = DeviceAttestationProvider("invalid")
	require.Error(t, invalidProvider.Validate())

	incompatibleProvider := valid
	incompatibleProvider.Platform = DevicePlatformIOS
	require.ErrorContains(t, incompatibleProvider.Validate(), "incompatible")

	missingPayloadHash := valid
	missingPayloadHash.PayloadHash = nil
	require.ErrorContains(t, missingPayloadHash.Validate(), "payload hash")

	missingVaultRef := valid
	missingVaultRef.VaultRef = ""
	require.ErrorContains(t, missingVaultRef.Validate(), "vault reference")

	whitespaceVaultRef := valid
	whitespaceVaultRef.VaultRef = " vault://attestations/attest-001"
	require.ErrorContains(t, whitespaceVaultRef.Validate(), "vault reference")

	unsupported := DeviceAttestationRecord{
		AttestationID: "attest-unsupported",
		Supported:     false,
		FailureReason: "unsupported_device",
	}
	require.NoError(t, unsupported.Validate())
}

func TestBiometricHardwareAttestationValidate(t *testing.T) {
	valid := BiometricHardwareAttestation{
		AttestationID:   "bio-attest-001",
		Modality:        BiometricModalityFingerprint,
		SensorType:      BiometricSensorUltrasonic,
		SecurityLevel:   BiometricSecurityHardwareBacked,
		AttestedAt:      time.Now().UTC(),
		LivenessScore:   8200,
		AntiSpoofScore:  9000,
		HardwareModel:   "Ultrasonic X2",
		FirmwareVersion: "5.4.1",
	}

	require.NoError(t, valid.Validate())

	invalid := valid
	invalid.Modality = BiometricModality("unknown")
	require.Error(t, invalid.Validate())
}
