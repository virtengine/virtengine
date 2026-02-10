package identity_scopes

import (
	"context"
	"fmt"
	"time"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// SocialMediaScopeRequest contains the data required to submit a social media scope.
type SocialMediaScopeRequest struct {
	AccountAddress         string
	ScopeID                string
	Provider               veidtypes.SocialMediaProviderType
	ProfileName            string
	ProfileNameHash        string
	Email                  string
	EmailHash              string
	Username               string
	UsernameHash           string
	Org                    string
	OrgHash                string
	AccountCreatedAt       *time.Time
	AccountAgeDays         uint32
	IsVerified             bool
	FriendCountRange       string
	Attestation            *veidtypes.VerificationAttestation
	AccountSignature       []byte
	EncryptedPayload       veidtypes.EncryptedPayloadEnvelope
	EvidenceStorageRef     string
	EvidenceStorageBackend string
	EvidenceMetadata       map[string]string
}

// SocialMediaAdapter submits social media scopes to the chain.
type SocialMediaAdapter interface {
	Provider() veidtypes.SocialMediaProviderType
	SubmitScope(ctx context.Context, req SocialMediaScopeRequest) (*ProofResult, error)
}

type socialMediaAdapter struct {
	provider    veidtypes.SocialMediaProviderType
	chainClient WebScopeChainClient
}

// NewGoogleSocialMediaAdapter creates an adapter for Google social media scopes.
func NewGoogleSocialMediaAdapter(chainClient WebScopeChainClient) SocialMediaAdapter {
	return &socialMediaAdapter{provider: veidtypes.SocialMediaProviderGoogle, chainClient: chainClient}
}

// NewFacebookSocialMediaAdapter creates an adapter for Facebook social media scopes.
func NewFacebookSocialMediaAdapter(chainClient WebScopeChainClient) SocialMediaAdapter {
	return &socialMediaAdapter{provider: veidtypes.SocialMediaProviderFacebook, chainClient: chainClient}
}

// NewMicrosoftSocialMediaAdapter creates an adapter for Microsoft social media scopes.
func NewMicrosoftSocialMediaAdapter(chainClient WebScopeChainClient) SocialMediaAdapter {
	return &socialMediaAdapter{provider: veidtypes.SocialMediaProviderMicrosoft, chainClient: chainClient}
}

func (a *socialMediaAdapter) Provider() veidtypes.SocialMediaProviderType {
	return a.provider
}

func (a *socialMediaAdapter) SubmitScope(ctx context.Context, req SocialMediaScopeRequest) (*ProofResult, error) {
	if req.Attestation == nil {
		return nil, fmt.Errorf("attestation is required")
	}
	if req.AccountAddress == "" {
		return nil, fmt.Errorf("account address is required")
	}
	if req.ScopeID == "" {
		return nil, fmt.Errorf("scope id is required")
	}
	if req.Provider != a.provider {
		return nil, fmt.Errorf("provider mismatch: %s", req.Provider)
	}
	if a.chainClient == nil {
		return nil, fmt.Errorf("chain client not configured")
	}
	if len(req.AccountSignature) == 0 {
		return nil, fmt.Errorf("account signature is required")
	}
	if err := req.EncryptedPayload.Validate(); err != nil {
		return nil, fmt.Errorf("invalid encrypted payload: %w", err)
	}

	attestationBytes, err := req.Attestation.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal attestation: %w", err)
	}

	profileNameHash := req.ProfileNameHash
	if profileNameHash == "" && req.ProfileName != "" {
		profileNameHash = veidtypes.HashSocialMediaField(req.ProfileName)
	}
	if profileNameHash == "" {
		return nil, fmt.Errorf("profile name hash is required")
	}

	emailHash := req.EmailHash
	if emailHash == "" && req.Email != "" {
		emailHash = veidtypes.HashSocialMediaField(req.Email)
	}

	usernameHash := req.UsernameHash
	if usernameHash == "" && req.Username != "" {
		usernameHash = veidtypes.HashSocialMediaField(req.Username)
	}

	orgHash := req.OrgHash
	if orgHash == "" && req.Org != "" {
		orgHash = veidtypes.HashSocialMediaField(req.Org)
	}

	var accountCreatedAt int64
	if req.AccountCreatedAt != nil {
		accountCreatedAt = req.AccountCreatedAt.Unix()
	}

	msg := &veidtypes.MsgSubmitSocialMediaScope{
		AccountAddress:         req.AccountAddress,
		ScopeId:                req.ScopeID,
		Provider:               veidtypes.SocialMediaProviderToProto(req.Provider),
		ProfileNameHash:        profileNameHash,
		EmailHash:              emailHash,
		UsernameHash:           usernameHash,
		OrgHash:                orgHash,
		AccountCreatedAt:       accountCreatedAt,
		AccountAgeDays:         req.AccountAgeDays,
		IsVerified:             req.IsVerified,
		FriendCountRange:       req.FriendCountRange,
		AttestationData:        attestationBytes,
		AccountSignature:       req.AccountSignature,
		EncryptedPayload:       req.EncryptedPayload,
		EvidenceStorageRef:     req.EvidenceStorageRef,
		EvidenceStorageBackend: req.EvidenceStorageBackend,
		EvidenceMetadata:       req.EvidenceMetadata,
	}

	if err := a.chainClient.SubmitSocialMediaScope(ctx, msg); err != nil {
		return nil, err
	}

	return &ProofResult{
		VerificationID: req.ScopeID,
		Status:         string(veidtypes.SocialMediaStatusVerified),
	}, nil
}
