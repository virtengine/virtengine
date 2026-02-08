package identity_scopes

import (
	"context"
	"fmt"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// SSOProofRequest contains the data required to submit an SSO proof.
type SSOProofRequest struct {
	AccountAddress         string
	LinkageID              string
	Attestation            *veidtypes.SSOAttestation
	EvidenceStorageRef     string
	EvidenceStorageBackend string
	EvidenceMetadata       map[string]string
}

// SSOAdapter submits SSO verification proofs to the chain.
type SSOAdapter interface {
	Provider() veidtypes.SSOProviderType
	SubmitProof(ctx context.Context, req SSOProofRequest) (*ProofResult, error)
}

type ssoAdapter struct {
	provider    veidtypes.SSOProviderType
	chainClient WebScopeChainClient
}

// NewGoogleSSOAdapter creates an adapter for Google SSO proofs.
func NewGoogleSSOAdapter(chainClient WebScopeChainClient) SSOAdapter {
	return &ssoAdapter{provider: veidtypes.SSOProviderGoogle, chainClient: chainClient}
}

// NewMicrosoftSSOAdapter creates an adapter for Microsoft SSO proofs.
func NewMicrosoftSSOAdapter(chainClient WebScopeChainClient) SSOAdapter {
	return &ssoAdapter{provider: veidtypes.SSOProviderMicrosoft, chainClient: chainClient}
}

func (a *ssoAdapter) Provider() veidtypes.SSOProviderType {
	return a.provider
}

func (a *ssoAdapter) SubmitProof(ctx context.Context, req SSOProofRequest) (*ProofResult, error) {
	if req.Attestation == nil {
		return nil, fmt.Errorf("attestation is required")
	}
	if req.AccountAddress == "" {
		return nil, fmt.Errorf("account address is required")
	}
	if req.LinkageID == "" {
		return nil, fmt.Errorf("linkage id is required")
	}
	if req.Attestation.ProviderType != a.provider {
		return nil, fmt.Errorf("attestation provider mismatch: %s", req.Attestation.ProviderType)
	}
	if a.chainClient == nil {
		return nil, fmt.Errorf("chain client not configured")
	}

	attestationBytes, err := req.Attestation.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal attestation: %w", err)
	}

	msg := &veidtypes.MsgSubmitSSOVerificationProof{
		AccountAddress:         req.AccountAddress,
		LinkageId:              req.LinkageID,
		AttestationData:        attestationBytes,
		EvidenceStorageRef:     req.EvidenceStorageRef,
		EvidenceStorageBackend: req.EvidenceStorageBackend,
		EvidenceMetadata:       req.EvidenceMetadata,
	}

	if err := a.chainClient.SubmitSSOVerificationProof(ctx, msg); err != nil {
		return nil, err
	}

	return &ProofResult{
		LinkageID: req.LinkageID,
		Status:    string(veidtypes.SSOStatusVerified),
	}, nil
}
