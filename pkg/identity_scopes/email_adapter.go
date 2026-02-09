package identity_scopes

import (
	"context"
	"fmt"
	"time"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// EmailOTPProofRequest contains the data required to submit an email OTP proof.
type EmailOTPProofRequest struct {
	AccountAddress         string
	VerificationID         string
	EmailHash              string
	DomainHash             string
	Nonce                  string
	IsOrganizational       bool
	Attestation            *veidtypes.VerificationAttestation
	AccountSignature       []byte
	VerifiedAt             *time.Time
	ExpiresAt              *time.Time
	EvidenceStorageRef     string
	EvidenceStorageBackend string
	EvidenceMetadata       map[string]string
}

// EmailOTPAdapter submits email OTP proofs to the chain.
type EmailOTPAdapter interface {
	SubmitProof(ctx context.Context, req EmailOTPProofRequest) (*ProofResult, error)
}

type emailAdapter struct {
	chainClient WebScopeChainClient
}

// NewEmailOTPAdapter creates a new email OTP adapter.
func NewEmailOTPAdapter(chainClient WebScopeChainClient) EmailOTPAdapter {
	return &emailAdapter{chainClient: chainClient}
}

func (a *emailAdapter) SubmitProof(ctx context.Context, req EmailOTPProofRequest) (*ProofResult, error) {
	if req.Attestation == nil {
		return nil, fmt.Errorf("attestation is required")
	}
	if req.AccountAddress == "" {
		return nil, fmt.Errorf("account address is required")
	}
	if req.VerificationID == "" {
		return nil, fmt.Errorf("verification id is required")
	}
	if req.EmailHash == "" {
		return nil, fmt.Errorf("email hash is required")
	}
	if req.Nonce == "" {
		return nil, fmt.Errorf("nonce is required")
	}
	if a.chainClient == nil {
		return nil, fmt.Errorf("chain client not configured")
	}

	if req.Attestation.Type != veidtypes.AttestationTypeEmailVerification {
		return nil, fmt.Errorf("invalid attestation type: %s", req.Attestation.Type)
	}

	attestationBytes, err := req.Attestation.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal attestation: %w", err)
	}

	msg := &veidtypes.MsgSubmitEmailVerificationProof{
		AccountAddress:         req.AccountAddress,
		VerificationId:         req.VerificationID,
		EmailHash:              req.EmailHash,
		DomainHash:             req.DomainHash,
		Nonce:                  req.Nonce,
		IsOrganizational:       req.IsOrganizational,
		AttestationData:        attestationBytes,
		AccountSignature:       req.AccountSignature,
		EvidenceStorageRef:     req.EvidenceStorageRef,
		EvidenceStorageBackend: req.EvidenceStorageBackend,
		EvidenceMetadata:       req.EvidenceMetadata,
	}

	if req.VerifiedAt != nil {
		msg.VerifiedAt = req.VerifiedAt.Unix()
	}
	if req.ExpiresAt != nil {
		msg.ExpiresAt = req.ExpiresAt.Unix()
	}

	if err := a.chainClient.SubmitEmailVerificationProof(ctx, msg); err != nil {
		return nil, err
	}

	return &ProofResult{
		VerificationID: req.VerificationID,
		Status:         string(veidtypes.EmailStatusVerified),
	}, nil
}
