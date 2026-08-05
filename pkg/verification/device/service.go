package device

import (
	"context"
	"errors"
	"fmt"
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
	req.Provider = provider

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

	if err := validateAttestationResult(result, req); err != nil {
		return AttestationResult{}, err
	}

	return result, nil
}

func validateAttestationResult(result AttestationResult, request AttestationRequest) error {
	switch result.Status {
	case AttestationStatusVerified:
		if !result.Verified {
			return errors.New("verified attestation status requires a verified result")
		}
	case AttestationStatusFailed, AttestationStatusUnsupported:
		if result.Verified {
			return fmt.Errorf("%s attestation status cannot be verified", result.Status)
		}
	default:
		return fmt.Errorf("invalid device attestation status: %s", result.Status)
	}
	if result.Provider != request.Provider {
		return errors.New("attestation result provider does not match request")
	}
	if result.Platform != request.Platform {
		return errors.New("attestation result platform does not match request")
	}
	if result.AppID != request.AppID || result.AppVersion != request.AppVersion {
		return errors.New("attestation result app binding does not match request")
	}
	if result.DeviceModel != request.DeviceModel || result.OSVersion != request.OSVersion {
		return errors.New("attestation result device binding does not match request")
	}
	if result.Nonce != request.Nonce {
		return errors.New("attestation result nonce does not match request")
	}
	if result.AttestedAt.IsZero() {
		return errors.New("attestation result time is required")
	}
	return nil
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
