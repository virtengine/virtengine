package types

import "time"

// BiometricModality identifies biometric modality.
type BiometricModality string

const (
	BiometricModalityFingerprint BiometricModality = "fingerprint"
	BiometricModalityIris        BiometricModality = "iris"
)

// BiometricSensorType identifies the sensor technology.
type BiometricSensorType string

const (
	BiometricSensorOptical     BiometricSensorType = "optical"
	BiometricSensorCapacitive  BiometricSensorType = "capacitive"
	BiometricSensorUltrasonic  BiometricSensorType = "ultrasonic"
	BiometricSensorIris        BiometricSensorType = "iris"
	BiometricSensorUnspecified BiometricSensorType = "unspecified"
)

// BiometricSecurityLevel describes the security tier for the sensor.
type BiometricSecurityLevel string

const (
	BiometricSecurityUnknown        BiometricSecurityLevel = "unknown"
	BiometricSecurityBasic          BiometricSecurityLevel = "basic"
	BiometricSecurityStrong         BiometricSecurityLevel = "strong"
	BiometricSecurityHardwareBacked BiometricSecurityLevel = "hardware_backed"
)

// BiometricHardwareAttestation represents a biometric hardware attestation record.
type BiometricHardwareAttestation struct {
	AttestationID string                 `json:"attestation_id"`
	Modality      BiometricModality      `json:"modality"`
	SensorType    BiometricSensorType    `json:"sensor_type"`
	SecurityLevel BiometricSecurityLevel `json:"security_level"`
	AttestedAt    time.Time              `json:"attested_at"`

	LivenessScore   uint32 `json:"liveness_score"`
	AntiSpoofScore  uint32 `json:"anti_spoof_score"`
	HardwareModel   string `json:"hardware_model"`
	FirmwareVersion string `json:"firmware_version"`

	PayloadHash []byte            `json:"payload_hash,omitempty"`
	VaultRef    string            `json:"vault_ref,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Validate validates the biometric hardware attestation.
func (b BiometricHardwareAttestation) Validate() error {
	if b.AttestationID == "" {
		return ErrInvalidScope.Wrap("biometric hardware attestation ID is required")
	}
	if !IsValidBiometricModality(b.Modality) {
		return ErrInvalidScope.Wrap("invalid biometric modality")
	}
	if !IsValidBiometricSensorType(b.SensorType) {
		return ErrInvalidScope.Wrap("invalid biometric sensor type")
	}
	if !IsValidBiometricSecurityLevel(b.SecurityLevel) {
		return ErrInvalidScope.Wrap("invalid biometric security level")
	}
	if b.AttestedAt.IsZero() {
		return ErrInvalidScope.Wrap("attested_at is required")
	}
	if b.HardwareModel == "" {
		return ErrInvalidScope.Wrap("hardware_model is required")
	}
	if b.FirmwareVersion == "" {
		return ErrInvalidScope.Wrap("firmware_version is required")
	}
	return nil
}

// IsValidBiometricModality checks if modality is supported.
func IsValidBiometricModality(modality BiometricModality) bool {
	switch modality {
	case BiometricModalityFingerprint, BiometricModalityIris:
		return true
	default:
		return false
	}
}

// IsValidBiometricSensorType checks if sensor type is supported.
func IsValidBiometricSensorType(sensor BiometricSensorType) bool {
	switch sensor {
	case BiometricSensorOptical,
		BiometricSensorCapacitive,
		BiometricSensorUltrasonic,
		BiometricSensorIris,
		BiometricSensorUnspecified:
		return true
	default:
		return false
	}
}

// IsValidBiometricSecurityLevel checks if security level is supported.
func IsValidBiometricSecurityLevel(level BiometricSecurityLevel) bool {
	switch level {
	case BiometricSecurityUnknown, BiometricSecurityBasic, BiometricSecurityStrong, BiometricSecurityHardwareBacked:
		return true
	default:
		return false
	}
}
