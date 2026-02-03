package sso

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/virtengine/virtengine/pkg/verification/audit"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ChainClient defines the minimal interface for on-chain submissions.
type ChainClient interface {
	SubmitSSOLinkage(ctx context.Context, msg *veidtypes.MsgCreateSSOLinkage) error
}

// SubmitSSOLinkage submits the SSO linkage on-chain if a client is configured.
func SubmitSSOLinkage(ctx context.Context, chainClient ChainClient, attestation *veidtypes.SSOAttestation, linkageID string, auditor audit.AuditLogger) error {
	if attestation == nil {
		return fmt.Errorf("%w: nil attestation", ErrOnChainSubmissionFailed)
	}

	if linkageID == "" {
		return fmt.Errorf("%w: missing linkage id", ErrOnChainSubmissionFailed)
	}

	attestationBytes, err := json.Marshal(attestation)
	if err != nil {
		return fmt.Errorf("%w: marshal attestation: %v", ErrOnChainSubmissionFailed, err)
	}

	attestationSignature, err := attestation.GetProofBytes()
	if err != nil {
		return fmt.Errorf("%w: attestation signature: %v", ErrOnChainSubmissionFailed, err)
	}

	msg := &veidtypes.MsgCreateSSOLinkage{
		AccountAddress:       attestation.LinkedAccountAddress,
		LinkageID:            linkageID,
		Provider:             attestation.ProviderType,
		Issuer:               attestation.OIDCIssuer,
		SubjectHash:          attestation.SubjectHash,
		Nonce:                attestation.OIDCNonce,
		AttestationData:      attestationBytes,
		AttestationSignature: attestationSignature,
		SignerFingerprint:    attestation.Issuer.KeyFingerprint,
		AccountSignature:     attestation.LinkageSignature,
	}

	if !attestation.ExpiresAt.IsZero() {
		msg.ExpiresAt = &attestation.ExpiresAt
	}
	if attestation.EmailDomainHash != "" {
		msg.EmailDomainHash = attestation.EmailDomainHash
	}
	if attestation.TenantIDHash != "" {
		msg.TenantIDHash = attestation.TenantIDHash
	}

	if err := msg.ValidateBasic(); err != nil {
		return fmt.Errorf("%w: %v", ErrOnChainSubmissionFailed, err)
	}

	if auditor != nil {
		auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeVerificationStarted,
			Timestamp: time.Now(),
			Actor:     msg.AccountAddress,
			Resource:  string(msg.Provider),
			Action:    "submit_sso_linkage",
			Outcome:   audit.OutcomePending,
			Details: map[string]interface{}{
				"issuer":       msg.Issuer,
				"subject_hash": msg.SubjectHash,
				"linkage_id":   msg.LinkageID,
			},
		})
	}

	if chainClient == nil {
		if auditor != nil {
			auditor.Log(ctx, audit.Event{
				Type:      audit.EventTypeError,
				Timestamp: time.Now(),
				Actor:     msg.AccountAddress,
				Resource:  string(msg.Provider),
				Action:    "submit_sso_linkage",
				Outcome:   audit.OutcomeFailure,
				Details: map[string]interface{}{
					"error": ErrChainClientNotConfigured.Error(),
				},
			})
		}
		return ErrChainClientNotConfigured
	}

	if err := chainClient.SubmitSSOLinkage(ctx, msg); err != nil {
		if auditor != nil {
			auditor.Log(ctx, audit.Event{
				Type:      audit.EventTypeVerificationFailed,
				Timestamp: time.Now(),
				Actor:     msg.AccountAddress,
				Resource:  string(msg.Provider),
				Action:    "submit_sso_linkage",
				Outcome:   audit.OutcomeFailure,
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			})
		}
		return fmt.Errorf("%w: %v", ErrOnChainSubmissionFailed, err)
	}

	if auditor != nil {
		auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeVerificationCompleted,
			Timestamp: time.Now(),
			Actor:     msg.AccountAddress,
			Resource:  string(msg.Provider),
			Action:    "submit_sso_linkage",
			Outcome:   audit.OutcomeSuccess,
			Details: map[string]interface{}{
				"linkage_id": msg.LinkageID,
			},
		})
	}

	return nil
}
