package identity_scopes

import (
	"context"
	"fmt"
	"time"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// SMSOTPProofRequest contains the data required to submit an SMS OTP proof.
type SMSOTPProofRequest struct {
	AccountAddress         string
	VerificationID         string
	PhoneHash              string
	PhoneHashSalt          string
	CountryCodeHash        string
	CarrierType            string
	IsVoIP                 bool
	ValidatorAddress       string
	Attestation            *veidtypes.VerificationAttestation
	AccountSignature       []byte
	VerifiedAt             *time.Time
	ExpiresAt              *time.Time
	EvidenceStorageRef     string
	EvidenceStorageBackend string
	EvidenceMetadata       map[string]string
}

// SMSOTPAdapter submits SMS OTP proofs to the chain.
type SMSOTPAdapter interface {
	SubmitProof(ctx context.Context, req SMSOTPProofRequest) (*ProofResult, error)
}

type smsAdapter struct {
	chainClient WebScopeChainClient
}

// NewSMSOTPAdapter creates a new SMS OTP adapter.
func NewSMSOTPAdapter(chainClient WebScopeChainClient) SMSOTPAdapter {
	return &smsAdapter{chainClient: chainClient}
}

func (a *smsAdapter) SubmitProof(ctx context.Context, req SMSOTPProofRequest) (*ProofResult, error) {
	if req.Attestation == nil {
		return nil, fmt.Errorf("attestation is required")
	}
	if req.AccountAddress == "" {
		return nil, fmt.Errorf("account address is required")
	}
	if req.VerificationID == "" {
		return nil, fmt.Errorf("verification id is required")
	}
	if req.PhoneHash == "" || req.PhoneHashSalt == "" {
		return nil, fmt.Errorf("phone hash and salt are required")
	}
	if a.chainClient == nil {
		return nil, fmt.Errorf("chain client not configured")
	}

	if req.Attestation.Type != veidtypes.AttestationTypeSMSVerification {
		return nil, fmt.Errorf("invalid attestation type: %s", req.Attestation.Type)
	}

	attestationBytes, err := req.Attestation.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal attestation: %w", err)
	}

	msg := &veidtypes.MsgSubmitSMSVerificationProof{
		AccountAddress:         req.AccountAddress,
		VerificationId:         req.VerificationID,
		PhoneHash:              req.PhoneHash,
		PhoneHashSalt:          req.PhoneHashSalt,
		CountryCodeHash:        req.CountryCodeHash,
		CarrierType:            req.CarrierType,
		IsVoip:                 req.IsVoIP,
		ValidatorAddress:       req.ValidatorAddress,
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

	if err := a.chainClient.SubmitSMSVerificationProof(ctx, msg); err != nil {
		return nil, err
	}

	return &ProofResult{
		VerificationID: req.VerificationID,
		Status:         string(veidtypes.SMSStatusVerified),
	}, nil
}
