package device

import (
	"context"
	"time"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// MockVerifier is a mock attestation verifier for tests and CI.
type MockVerifier struct {
	ShouldFail     bool
	IntegrityScore uint32
	IntegrityLevel veidtypes.DeviceIntegrityLevel
	HardwareBacked bool
	FailureReason  string
}

// Verify returns a deterministic mock result.
func (m MockVerifier) Verify(_ context.Context, req AttestationRequest) (AttestationResult, error) {
	if m.ShouldFail {
		return AttestationResult{
			Status:         AttestationStatusFailed,
			Verified:       false,
			IntegrityLevel: veidtypes.DeviceIntegrityUnknown,
			IntegrityScore: 0,
			FailureReason:  m.failureReason(),
			AttestedAt:     time.Now().UTC(),
			Provider:       req.Provider,
			Platform:       req.Platform,
			DeviceModel:    req.DeviceModel,
			OSVersion:      req.OSVersion,
			AppVersion:     req.AppVersion,
			AppID:          req.AppID,
			Nonce:          req.Nonce,
		}, nil
	}

	score := m.IntegrityScore
	if score == 0 {
		score = 8500
	}
	level := m.IntegrityLevel
	if level == "" {
		level = veidtypes.DeviceIntegrityStrong
	}

	return AttestationResult{
		Status:         AttestationStatusVerified,
		Verified:       true,
		IntegrityLevel: level,
		IntegrityScore: score,
		HardwareBacked: m.HardwareBacked,
		AttestedAt:     time.Now().UTC(),
		Provider:       req.Provider,
		Platform:       req.Platform,
		DeviceModel:    req.DeviceModel,
		OSVersion:      req.OSVersion,
		AppVersion:     req.AppVersion,
		AppID:          req.AppID,
		Nonce:          req.Nonce,
	}, nil
}

func (m MockVerifier) failureReason() string {
	if m.FailureReason != "" {
		return m.FailureReason
	}
	return "mock_attestation_failed"
}
