package device

import (
	"context"
	"errors"
	"time"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// Service verifies device attestations using configured providers.
type Service struct {
	providers map[veidtypes.DeviceAttestationProvider]AttestationVerifier
}

// NewService creates a new device attestation service.
func NewService(providers map[veidtypes.DeviceAttestationProvider]AttestationVerifier) *Service {
	return &Service{providers: providers}
}

// VerifyAttestation verifies device attestation and returns the verification result.
func (s *Service) VerifyAttestation(ctx context.Context, req AttestationRequest) (AttestationResult, error) {
	provider := req.Provider
	if provider == "" {
		provider = defaultProviderForPlatform(req.Platform)
	}

	verifier, ok := s.providers[provider]
	if !ok {
		if req.RequireAttestation {
			return AttestationResult{}, errors.New("device attestation provider not configured")
		}
		return unsupportedResult(req, provider, "provider_not_configured"), nil
	}

	result, err := verifier.Verify(ctx, req)
	if err != nil {
		if req.AllowFallback && !req.RequireAttestation {
			return unsupportedResult(req, provider, err.Error()), nil
		}
		return AttestationResult{}, err
	}

	if result.AttestedAt.IsZero() {
		result.AttestedAt = time.Now().UTC()
	}
	if result.Nonce == "" {
		result.Nonce = req.Nonce
	}

	return result, nil
}

func defaultProviderForPlatform(platform veidtypes.DevicePlatform) veidtypes.DeviceAttestationProvider {
	switch platform {
	case veidtypes.DevicePlatformAndroid:
		return veidtypes.DeviceAttestationProviderPlayIntegrity
	case veidtypes.DevicePlatformIOS:
		return veidtypes.DeviceAttestationProviderAppAttest
	default:
		return ""
	}
}

func unsupportedResult(req AttestationRequest, provider veidtypes.DeviceAttestationProvider, reason string) AttestationResult {
	return AttestationResult{
		Status:         AttestationStatusUnsupported,
		Verified:       false,
		IntegrityLevel: veidtypes.DeviceIntegrityUnsupported,
		IntegrityScore: 5000,
		FailureReason:  reason,
		AttestedAt:     time.Now().UTC(),
		Provider:       provider,
		Platform:       req.Platform,
		DeviceModel:    req.DeviceModel,
		OSVersion:      req.OSVersion,
		AppVersion:     req.AppVersion,
		AppID:          req.AppID,
		Nonce:          req.Nonce,
	}
}
