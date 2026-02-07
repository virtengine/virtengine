package types

import (
	"fmt"
	"time"
)

// DevicePlatform identifies the mobile platform that generated the attestation.
type DevicePlatform string

const (
	DevicePlatformAndroid DevicePlatform = "android"
	DevicePlatformIOS     DevicePlatform = "ios"
)

// DeviceAttestationProvider identifies the attestation provider.
type DeviceAttestationProvider string

const (
	DeviceAttestationProviderPlayIntegrity DeviceAttestationProvider = "android_play_integrity"
	DeviceAttestationProviderSafetyNet     DeviceAttestationProvider = "android_safetynet"
	DeviceAttestationProviderDeviceCheck   DeviceAttestationProvider = "ios_devicecheck"
	DeviceAttestationProviderAppAttest     DeviceAttestationProvider = "ios_app_attest"
)

// DeviceIntegrityLevel represents the assessed integrity level of a device.
type DeviceIntegrityLevel string

const (
	DeviceIntegrityUnknown        DeviceIntegrityLevel = "unknown"
	DeviceIntegrityBasic          DeviceIntegrityLevel = "basic"
	DeviceIntegrityStrong         DeviceIntegrityLevel = "strong"
	DeviceIntegrityHardwareBacked DeviceIntegrityLevel = "hardware_backed"
	DeviceIntegrityUnsupported    DeviceIntegrityLevel = "unsupported"
)

// DeviceAttestationRecord represents a device integrity attestation payload stored on-chain.
// The encrypted payload (vault reference) is stored separately in the identity scope envelope.
type DeviceAttestationRecord struct {
	AttestationID string                    `json:"attestation_id"`
	Platform      DevicePlatform            `json:"platform"`
	Provider      DeviceAttestationProvider `json:"provider"`
	Nonce         string                    `json:"nonce"`
	AttestedAt    time.Time                 `json:"attested_at"`

	IntegrityLevel DeviceIntegrityLevel `json:"integrity_level"`
	DeviceModel    string               `json:"device_model"`
	OSVersion      string               `json:"os_version"`
	AppVersion     string               `json:"app_version"`
	AppID          string               `json:"app_id"`
	HardwareBacked bool                 `json:"hardware_backed"`

	Supported     bool   `json:"supported"`
	Verified      bool   `json:"verified"`
	FailureReason string `json:"failure_reason,omitempty"`

	PayloadHash []byte            `json:"payload_hash,omitempty"`
	VaultRef    string            `json:"vault_ref,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Validate validates the device attestation record.
func (r DeviceAttestationRecord) Validate() error {
	if r.AttestationID == "" {
		return ErrInvalidScope.Wrap("device attestation ID is required")
	}

	if !r.Supported {
		if r.FailureReason == "" {
			return ErrInvalidScope.Wrap("unsupported device attestation requires failure reason")
		}
		return nil
	}

	if !IsValidDevicePlatform(r.Platform) {
		return ErrInvalidScope.Wrapf("invalid device platform: %s", r.Platform)
	}
	if !IsValidDeviceAttestationProvider(r.Provider) {
		return ErrInvalidScope.Wrapf("invalid device attestation provider: %s", r.Provider)
	}
	if r.Nonce == "" {
		return ErrInvalidScope.Wrap("attestation nonce is required")
	}
	if r.AttestedAt.IsZero() {
		return ErrInvalidScope.Wrap("attested_at is required")
	}
	if !IsValidDeviceIntegrityLevel(r.IntegrityLevel) {
		return ErrInvalidScope.Wrapf("invalid device integrity level: %s", r.IntegrityLevel)
	}
	if r.DeviceModel == "" {
		return ErrInvalidScope.Wrap("device_model is required")
	}
	if r.OSVersion == "" {
		return ErrInvalidScope.Wrap("os_version is required")
	}
	if r.AppVersion == "" {
		return ErrInvalidScope.Wrap("app_version is required")
	}
	if r.AppID == "" {
		return ErrInvalidScope.Wrap("app_id is required")
	}

	return nil
}

// IsValidDevicePlatform checks if a platform is supported.
func IsValidDevicePlatform(platform DevicePlatform) bool {
	switch platform {
	case DevicePlatformAndroid, DevicePlatformIOS:
		return true
	default:
		return false
	}
}

// IsValidDeviceAttestationProvider checks if the attestation provider is supported.
func IsValidDeviceAttestationProvider(provider DeviceAttestationProvider) bool {
	switch provider {
	case DeviceAttestationProviderPlayIntegrity,
		DeviceAttestationProviderSafetyNet,
		DeviceAttestationProviderDeviceCheck,
		DeviceAttestationProviderAppAttest:
		return true
	default:
		return false
	}
}

// IsValidDeviceIntegrityLevel checks if integrity level is supported.
func IsValidDeviceIntegrityLevel(level DeviceIntegrityLevel) bool {
	switch level {
	case DeviceIntegrityUnknown,
		DeviceIntegrityBasic,
		DeviceIntegrityStrong,
		DeviceIntegrityHardwareBacked,
		DeviceIntegrityUnsupported:
		return true
	default:
		return false
	}
}

// DeviceAttestationFingerprint builds a short deterministic fingerprint for audit logs.
func DeviceAttestationFingerprint(record DeviceAttestationRecord) string {
	if record.AttestationID == "" || record.Nonce == "" {
		return ""
	}
	return fmt.Sprintf("%s:%s:%s", record.AttestationID, record.Platform, record.Nonce[:min(12, len(record.Nonce))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
