package device

import (
	"context"
	"time"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// AttestationStatus represents the device attestation verification state.
type AttestationStatus string

const (
	AttestationStatusVerified    AttestationStatus = "verified"
	AttestationStatusFailed      AttestationStatus = "failed"
	AttestationStatusUnsupported AttestationStatus = "unsupported"
)

// AttestationRequest contains the data required to verify device attestation.
type AttestationRequest struct {
	AccountAddress string                              `json:"account_address"`
	Platform       veidtypes.DevicePlatform            `json:"platform"`
	Provider       veidtypes.DeviceAttestationProvider `json:"provider"`
	AppID          string                              `json:"app_id"`
	AppVersion     string                              `json:"app_version"`
	DeviceModel    string                              `json:"device_model"`
	OSVersion      string                              `json:"os_version"`
	Nonce          string                              `json:"nonce"`
	Attestation    []byte                              `json:"attestation"`
	RequestedAt    time.Time                           `json:"requested_at"`

	RequireAttestation bool `json:"require_attestation"`
	AllowFallback      bool `json:"allow_fallback"`
}

// AttestationResult contains the verification outcome and scoring signals.
type AttestationResult struct {
	Status         AttestationStatus                   `json:"status"`
	Verified       bool                                `json:"verified"`
	IntegrityLevel veidtypes.DeviceIntegrityLevel      `json:"integrity_level"`
	IntegrityScore uint32                              `json:"integrity_score"`
	FailureReason  string                              `json:"failure_reason,omitempty"`
	AttestedAt     time.Time                           `json:"attested_at"`
	HardwareBacked bool                                `json:"hardware_backed"`
	Provider       veidtypes.DeviceAttestationProvider `json:"provider"`
	Platform       veidtypes.DevicePlatform            `json:"platform"`
	DeviceModel    string                              `json:"device_model"`
	OSVersion      string                              `json:"os_version"`
	AppVersion     string                              `json:"app_version"`
	AppID          string                              `json:"app_id"`
	Nonce          string                              `json:"nonce"`
	RawAttestation []byte                              `json:"raw_attestation,omitempty"`
	Metadata       map[string]string                   `json:"metadata,omitempty"`
}

// AttestationVerifier verifies a device attestation document.
type AttestationVerifier interface {
	Verify(ctx context.Context, req AttestationRequest) (AttestationResult, error)
}

// BuildDeviceAttestationRecord converts a verification result to an on-chain record.
func BuildDeviceAttestationRecord(result AttestationResult, vaultRef string, payloadHash []byte) veidtypes.DeviceAttestationRecord {
	return veidtypes.DeviceAttestationRecord{
		AttestationID:  result.AttestedAt.Format("20060102T150405Z") + ":" + result.DeviceModel,
		Platform:       result.Platform,
		Provider:       result.Provider,
		Nonce:          result.Nonce,
		AttestedAt:     result.AttestedAt,
		IntegrityLevel: result.IntegrityLevel,
		DeviceModel:    result.DeviceModel,
		OSVersion:      result.OSVersion,
		AppVersion:     result.AppVersion,
		AppID:          result.AppID,
		HardwareBacked: result.HardwareBacked,
		Supported:      result.Status != AttestationStatusUnsupported,
		Verified:       result.Verified,
		FailureReason:  result.FailureReason,
		PayloadHash:    payloadHash,
		VaultRef:       vaultRef,
		Metadata:       result.Metadata,
	}
}
