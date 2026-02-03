package sso

import (
	"context"
	"fmt"
	"time"

	"github.com/virtengine/virtengine/pkg/verification/audit"
	"github.com/virtengine/virtengine/pkg/verification/signer"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// SignAndSubmitAttestation signs the attestation and optionally submits linkage on-chain.
func SignAndSubmitAttestation(ctx context.Context, attestation *veidtypes.SSOAttestation, signerSvc signer.SignerService, chainClient ChainClient, linkageID string, auditor audit.AuditLogger) error {
	if attestation == nil {
		return fmt.Errorf("invalid attestation: missing required fields")
	}
	if attestation.LinkedAccountAddress == "" || attestation.OIDCIssuer == "" {
		return fmt.Errorf("invalid attestation: missing required fields")
	}
	if signerSvc == nil {
		return fmt.Errorf("signer service not available")
	}

	if auditor != nil {
		auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeAttestationSigned,
			Timestamp: time.Now(),
			Actor:     attestation.LinkedAccountAddress,
			Resource:  string(attestation.ProviderType),
			Action:    "sign_sso_attestation",
			Outcome:   audit.OutcomePending,
			Details: map[string]interface{}{
				"oidc_issuer":  attestation.OIDCIssuer,
				"subject_hash": attestation.SubjectHash,
			},
		})
	}

	if err := signerSvc.SignAttestation(ctx, &attestation.VerificationAttestation); err != nil {
		if auditor != nil {
			auditor.Log(ctx, audit.Event{
				Type:      audit.EventTypeError,
				Timestamp: time.Now(),
				Actor:     attestation.LinkedAccountAddress,
				Resource:  string(attestation.ProviderType),
				Action:    "sign_sso_attestation",
				Outcome:   audit.OutcomeFailure,
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			})
		}
		return fmt.Errorf("failed to sign attestation: %w", err)
	}

	if auditor != nil {
		auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeAttestationSigned,
			Timestamp: time.Now(),
			Actor:     attestation.LinkedAccountAddress,
			Resource:  string(attestation.ProviderType),
			Action:    "sign_sso_attestation",
			Outcome:   audit.OutcomeSuccess,
			Details: map[string]interface{}{
				"attestation_id": attestation.ID,
			},
		})
	}

	return SubmitSSOLinkage(ctx, chainClient, attestation, linkageID, auditor)
}
