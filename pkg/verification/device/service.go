package device

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	if ctx == nil {
		return AttestationResult{}, errors.New("device attestation context is required")
	}
	if err := ctx.Err(); err != nil {
		return AttestationResult{}, err
	}
	provider := req.Provider
	if provider == "" {
		provider = defaultProviderForPlatform(req.Platform)
	}
	req.Provider = provider
	if err := validateAttestationRequest(req); err != nil {
		return AttestationResult{}, err
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

	if err := validateAttestationResult(result, req); err != nil {
		return AttestationResult{}, err
	}

	return result, nil
}

func validateAttestationRequest(request AttestationRequest) error {
	if !veidtypes.IsValidDevicePlatform(request.Platform) {
		return fmt.Errorf("invalid attestation request platform: %s", request.Platform)
	}
	if !veidtypes.IsValidDeviceAttestationProvider(request.Provider) || !attestationProviderMatchesPlatform(request.Provider, request.Platform) {
		return fmt.Errorf("attestation request provider is incompatible with platform")
	}
	if strings.TrimSpace(request.AccountAddress) == "" || strings.TrimSpace(request.AppID) == "" || strings.TrimSpace(request.AppVersion) == "" || strings.TrimSpace(request.DeviceModel) == "" || strings.TrimSpace(request.OSVersion) == "" || strings.TrimSpace(request.Nonce) == "" {
		return errors.New("attestation request identity bindings are required")
	}
	if request.RequestedAt.IsZero() {
		return errors.New("attestation request time is required")
	}
	if request.AllowFallback && request.RequireAttestation {
		return errors.New("required attestation cannot allow fallback")
	}
	if request.RequireAttestation && len(request.Attestation) == 0 {
		return errors.New("required attestation payload is missing")
	}
	return nil
}

func validateAttestationResult(result AttestationResult, request AttestationRequest) error {
	switch result.Status {
	case AttestationStatusVerified:
		if !result.Verified {
			return errors.New("verified attestation status requires a verified result")
		}
		if result.IntegrityLevel == veidtypes.DeviceIntegrityUnknown || result.IntegrityLevel == veidtypes.DeviceIntegrityUnsupported || !veidtypes.IsValidDeviceIntegrityLevel(result.IntegrityLevel) {
			return errors.New("verified attestation requires a concrete integrity level")
		}
		if result.IntegrityLevel == veidtypes.DeviceIntegrityHardwareBacked && !result.HardwareBacked {
			return errors.New("hardware-backed integrity requires hardware-backed evidence")
		}
		if strings.TrimSpace(result.FailureReason) != "" {
			return errors.New("verified attestation cannot include a failure reason")
		}
	case AttestationStatusFailed, AttestationStatusUnsupported:
		if result.Verified {
			return fmt.Errorf("%s attestation status cannot be verified", result.Status)
		}
		if strings.TrimSpace(result.FailureReason) == "" {
			return fmt.Errorf("%s attestation status requires a failure reason", result.Status)
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
	if result.IntegrityScore > 10000 {
		return errors.New("attestation result integrity score is out of range")
	}
	return nil
}

func attestationProviderMatchesPlatform(provider veidtypes.DeviceAttestationProvider, platform veidtypes.DevicePlatform) bool {
	switch platform {
	case veidtypes.DevicePlatformAndroid:
		return provider == veidtypes.DeviceAttestationProviderPlayIntegrity || provider == veidtypes.DeviceAttestationProviderSafetyNet
	case veidtypes.DevicePlatformIOS:
		return provider == veidtypes.DeviceAttestationProviderDeviceCheck || provider == veidtypes.DeviceAttestationProviderAppAttest
	default:
		return false
	}
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
