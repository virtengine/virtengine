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
	SubmitSSOVerificationProof(ctx context.Context, msg *veidtypes.MsgSubmitSSOVerificationProof) error
}

// SubmitSSOVerificationProof submits the SSO verification proof on-chain if a client is configured.
func SubmitSSOVerificationProof(ctx context.Context, chainClient ChainClient, attestation *veidtypes.SSOAttestation, linkageID string, auditor audit.AuditLogger) error {
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

	metadata := attestation.Metadata
	if metadata == nil {
		metadata = map[string]string{}
	}

	msg := &veidtypes.MsgSubmitSSOVerificationProof{
		AccountAddress:         attestation.LinkedAccountAddress,
		LinkageId:              linkageID,
		AttestationData:        attestationBytes,
		EvidenceStorageRef:     metadata["evidence_storage_ref"],
		EvidenceStorageBackend: metadata["evidence_storage_backend"],
		EvidenceMetadata:       metadata,
	}

	if err := msg.ValidateBasic(); err != nil {
		return fmt.Errorf("%w: %v", ErrOnChainSubmissionFailed, err)
	}

	if auditor != nil {
		auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeVerificationStarted,
			Timestamp: time.Now(),
			Actor:     msg.AccountAddress,
			Resource:  string(attestation.ProviderType),
			Action:    "submit_sso_linkage",
			Outcome:   audit.OutcomePending,
			Details: map[string]interface{}{
				"issuer":       attestation.OIDCIssuer,
				"subject_hash": attestation.SubjectHash,
				"linkage_id":   linkageID,
			},
		})
	}

	if chainClient == nil {
		if auditor != nil {
			auditor.Log(ctx, audit.Event{
				Type:      audit.EventTypeError,
				Timestamp: time.Now(),
				Actor:     msg.AccountAddress,
				Resource:  string(attestation.ProviderType),
				Action:    "submit_sso_linkage",
				Outcome:   audit.OutcomeFailure,
				Details: map[string]interface{}{
					"error": ErrChainClientNotConfigured.Error(),
				},
			})
		}
		return ErrChainClientNotConfigured
	}

	if err := chainClient.SubmitSSOVerificationProof(ctx, msg); err != nil {
		if auditor != nil {
			auditor.Log(ctx, audit.Event{
				Type:      audit.EventTypeVerificationFailed,
				Timestamp: time.Now(),
				Actor:     msg.AccountAddress,
				Resource:  string(attestation.ProviderType),
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
			Resource:  string(attestation.ProviderType),
			Action:    "submit_sso_linkage",
			Outcome:   audit.OutcomeSuccess,
			Details: map[string]interface{}{
				"linkage_id": linkageID,
			},
		})
	}

	return nil
}
