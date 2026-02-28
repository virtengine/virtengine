package sso

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/verification/oidc"
	signerpkg "github.com/virtengine/virtengine/pkg/verification/signer"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

type stubOIDCVerifier struct {
	policy *oidc.IssuerPolicy
	claims *oidc.VerifiedClaims
}

func (s *stubOIDCVerifier) VerifyToken(ctx context.Context, token string, req *oidc.VerificationRequest) (*oidc.VerifiedClaims, error) {
	return s.claims, nil
}

func (s *stubOIDCVerifier) GetAuthorizationURL(ctx context.Context, req *oidc.AuthorizationRequest) (string, error) {
	return "https://example.com/auth", nil
}

func (s *stubOIDCVerifier) ExchangeCode(ctx context.Context, code string, req *oidc.CodeExchangeRequest) (*oidc.TokenResponse, error) {
	return &oidc.TokenResponse{VerifiedClaims: s.claims}, nil
}

func (s *stubOIDCVerifier) RefreshJWKS(ctx context.Context, issuer string) error {
	return nil
}

func (s *stubOIDCVerifier) GetIssuerPolicy(ctx context.Context, issuer string) (*oidc.IssuerPolicy, error) {
	return s.policy, nil
}

func (s *stubOIDCVerifier) IsIssuerAllowed(ctx context.Context, issuer string) bool {
	return true
}

func (s *stubOIDCVerifier) HealthCheck(ctx context.Context) (*oidc.HealthStatus, error) {
	return &oidc.HealthStatus{Healthy: true, Status: "healthy"}, nil
}

func (s *stubOIDCVerifier) Close() error {
	return nil
}

type stubSigner struct {
	activeKey *veidtypes.SignerKeyInfo
	signed    bool
}

func (s *stubSigner) SignAttestation(ctx context.Context, attestation *veidtypes.VerificationAttestation) error {
	s.signed = true
	return nil
}

func (s *stubSigner) VerifyAttestation(ctx context.Context, attestation *veidtypes.VerificationAttestation) (bool, error) {
	return true, nil
}

func (s *stubSigner) GetActiveKey(ctx context.Context) (*veidtypes.SignerKeyInfo, error) {
	return s.activeKey, nil
}

func (s *stubSigner) GetKeyByID(ctx context.Context, keyID string) (*veidtypes.SignerKeyInfo, error) {
	return s.activeKey, nil
}

func (s *stubSigner) GetKeyByFingerprint(ctx context.Context, fingerprint string) (*veidtypes.SignerKeyInfo, error) {
	return s.activeKey, nil
}

func (s *stubSigner) ListKeys(ctx context.Context) ([]*veidtypes.SignerKeyInfo, error) {
	return []*veidtypes.SignerKeyInfo{s.activeKey}, nil
}

func (s *stubSigner) RotateKey(ctx context.Context, req *signerpkg.KeyRotationRequest) (*veidtypes.KeyRotationRecord, error) {
	return nil, nil
}

func (s *stubSigner) CompleteRotation(ctx context.Context, rotationID string) error {
	return nil
}

func (s *stubSigner) RevokeKey(ctx context.Context, keyID string, reason veidtypes.KeyRevocationReason) error {
	return nil
}

func (s *stubSigner) GetRotationStatus(ctx context.Context, rotationID string) (*veidtypes.KeyRotationRecord, error) {
	return nil, nil
}

func (s *stubSigner) GetSignerInfo(ctx context.Context) (*veidtypes.SignerRegistryEntry, error) {
	return nil, nil
}

func (s *stubSigner) HealthCheck(ctx context.Context) (*signerpkg.HealthStatus, error) {
	return &signerpkg.HealthStatus{Healthy: true, Status: "healthy"}, nil
}

func (s *stubSigner) Close() error {
	return nil
}

type stubChainClient struct {
	walletBinding *WalletBinding
	linkages      map[string]*veidtypes.SSOLinkageMetadata
	submittedMsg  *veidtypes.MsgSubmitSSOVerificationProof
	revokedMsg    *veidtypes.MsgRevokeSSOLinkage
	queries       []LinkageQuery
}

func (s *stubChainClient) SubmitSSOVerificationProof(ctx context.Context, msg *veidtypes.MsgSubmitSSOVerificationProof) error {
	s.submittedMsg = msg
	return nil
}

func (s *stubChainClient) GetWalletBinding(ctx context.Context, accountAddress string) (*WalletBinding, error) {
	if s.walletBinding == nil || s.walletBinding.AccountAddress != accountAddress {
		return nil, ErrLinkageNotFound
	}
	return s.walletBinding, nil
}

func (s *stubChainClient) RevokeSSOLinkage(ctx context.Context, msg *veidtypes.MsgRevokeSSOLinkage) error {
	s.revokedMsg = msg
	return nil
}

func (s *stubChainClient) QuerySSOLinkage(ctx context.Context, query LinkageQuery) (*veidtypes.SSOLinkageMetadata, error) {
	s.queries = append(s.queries, query)
	key := linkageQueryKey(query.AccountAddress, query.Provider, query.LinkageID)
	if linkage, ok := s.linkages[key]; ok {
		return linkage, nil
	}
	return nil, ErrLinkageNotFound
}

func linkageQueryKey(account string, provider veidtypes.SSOProviderType, linkageID string) string {
	return fmt.Sprintf("%s|%s|%s", account, provider, linkageID)
}

func testAccountAddress() string {
	return sdk.AccAddress(bytes.Repeat([]byte{0x1}, 20)).String()
}

func testChallenge(account string) *Challenge {
	return &Challenge{
		ChallengeID:    "challenge-1",
		AccountAddress: account,
		ProviderType:   veidtypes.SSOProviderGoogle,
		OIDCIssuer:     "https://accounts.google.com",
		State:          "state-1",
		Nonce:          "nonce-1",
		LinkageMessage: "I authorize linking my SSO identity to this wallet",
		RedirectURI:    "https://portal.virtengine.example/callback",
		Status:         ChallengeStatusPending,
		CreatedAt:      time.Now().Add(-time.Minute),
		ExpiresAt:      time.Now().Add(time.Hour),
	}
}

func signChallengeMessage(privateKey ed25519.PrivateKey, message string) []byte {
	return ed25519.Sign(privateKey, hashLinkageMessage([]byte(message)))
}

func TestCompleteVerificationVerifiesLinkageSignature(t *testing.T) {
	t.Parallel()

	account := testAccountAddress()
	challenge := testChallenge(account)
	pubKey, privateKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	service := &DefaultService{
		config: Config{
			RequireLinkageSignature: true,
			AttestationValidityDays: 365,
		},
		oidcVerifier: &stubOIDCVerifier{
			policy: &oidc.IssuerPolicy{ClientID: "client-id"},
			claims: &oidc.VerifiedClaims{
				Issuer:        challenge.OIDCIssuer,
				Subject:       "subject-1",
				Email:         "user@example.com",
				EmailVerified: true,
				ProviderType:  challenge.ProviderType,
			},
		},
		signer: &stubSigner{
			activeKey: &veidtypes.SignerKeyInfo{Fingerprint: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
		},
		chainClient: &stubChainClient{
			walletBinding: &WalletBinding{
				AccountAddress: account,
				WalletID:       "wallet-1",
				BindingPubKey:  pubKey,
			},
			linkages: map[string]*veidtypes.SSOLinkageMetadata{},
		},
		challenges: map[string]*Challenge{challenge.ChallengeID: challenge},
		byAccount:  map[string][]string{},
	}

	resp, err := service.CompleteVerification(context.Background(), &CompleteRequest{
		ChallengeID:      challenge.ChallengeID,
		IDToken:          "token",
		LinkageSignature: signChallengeMessage(privateKey, challenge.LinkageMessage),
		State:            challenge.State,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Attestation)
	assert.Equal(t, ChallengeStatusCompleted, service.challenges[challenge.ChallengeID].Status)
	assert.True(t, service.signer.(*stubSigner).signed)
}

func TestCompleteVerificationRejectsInvalidLinkageSignature(t *testing.T) {
	t.Parallel()

	account := testAccountAddress()
	challenge := testChallenge(account)
	pubKey, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	_, wrongPrivateKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	service := &DefaultService{
		config: Config{
			RequireLinkageSignature: true,
			AttestationValidityDays: 365,
		},
		oidcVerifier: &stubOIDCVerifier{
			policy: &oidc.IssuerPolicy{ClientID: "client-id"},
			claims: &oidc.VerifiedClaims{
				Issuer:        challenge.OIDCIssuer,
				Subject:       "subject-1",
				Email:         "user@example.com",
				EmailVerified: true,
				ProviderType:  challenge.ProviderType,
			},
		},
		signer: &stubSigner{
			activeKey: &veidtypes.SignerKeyInfo{Fingerprint: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
		},
		chainClient: &stubChainClient{
			walletBinding: &WalletBinding{
				AccountAddress: account,
				WalletID:       "wallet-1",
				BindingPubKey:  pubKey,
			},
			linkages: map[string]*veidtypes.SSOLinkageMetadata{},
		},
		challenges: map[string]*Challenge{challenge.ChallengeID: challenge},
		byAccount:  map[string][]string{},
	}

	resp, err := service.CompleteVerification(context.Background(), &CompleteRequest{
		ChallengeID:      challenge.ChallengeID,
		IDToken:          "token",
		LinkageSignature: signChallengeMessage(wrongPrivateKey, challenge.LinkageMessage),
		State:            challenge.State,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
	require.NotNil(t, resp.Error)
	assert.Equal(t, "invalid_linkage_signature", resp.Error.Code)
	assert.Equal(t, ChallengeStatusFailed, service.challenges[challenge.ChallengeID].Status)
}

func TestRevokeVerificationSubmitsChainMessage(t *testing.T) {
	t.Parallel()

	account := testAccountAddress()
	linkage := &veidtypes.SSOLinkageMetadata{
		Version:     veidtypes.SSOVerificationVersion,
		LinkageID:   "linkage-1",
		Provider:    veidtypes.SSOProviderGoogle,
		Issuer:      "https://accounts.google.com",
		SubjectHash: veidtypes.HashSubjectID("subject-1"),
		Nonce:       "nonce-1",
		VerifiedAt:  time.Now().Add(-time.Hour),
		Status:      veidtypes.SSOStatusVerified,
	}

	chainClient := &stubChainClient{
		linkages: map[string]*veidtypes.SSOLinkageMetadata{
			linkageQueryKey(account, "", linkage.LinkageID): linkage,
		},
	}

	service := &DefaultService{
		chainClient: chainClient,
		config:      Config{},
	}

	err := service.RevokeVerification(context.Background(), &RevokeRequest{
		LinkageID:      linkage.LinkageID,
		AccountAddress: account,
		Reason:         "user_requested",
		Signature:      []byte("signed"),
	})
	require.NoError(t, err)
	require.NotNil(t, chainClient.revokedMsg)
	assert.Equal(t, linkage.LinkageID, chainClient.revokedMsg.LinkageID)
	assert.Equal(t, account, chainClient.revokedMsg.AccountAddress)
}

func TestGetLinkageStatusUsesChainData(t *testing.T) {
	t.Parallel()

	account := testAccountAddress()
	verifiedAt := time.Now().Add(-2 * time.Hour).UTC()
	expiresAt := time.Now().Add(24 * time.Hour).UTC()
	linkage := &veidtypes.SSOLinkageMetadata{
		Version:     veidtypes.SSOVerificationVersion,
		LinkageID:   "linkage-1",
		Provider:    veidtypes.SSOProviderGoogle,
		Issuer:      "https://accounts.google.com",
		SubjectHash: veidtypes.HashSubjectID("subject-1"),
		Nonce:       "nonce-1",
		VerifiedAt:  verifiedAt,
		ExpiresAt:   &expiresAt,
		Status:      veidtypes.SSOStatusVerified,
	}

	chainClient := &stubChainClient{
		linkages: map[string]*veidtypes.SSOLinkageMetadata{
			linkageQueryKey(account, veidtypes.SSOProviderGoogle, ""): linkage,
		},
	}

	service := &DefaultService{
		chainClient: chainClient,
		config:      Config{},
	}

	status, err := service.GetLinkageStatus(context.Background(), account)
	require.NoError(t, err)
	require.NotNil(t, status)
	assert.True(t, status.Exists)
	assert.Equal(t, linkage.LinkageID, status.LinkageID)
	assert.Equal(t, veidtypes.SSOProviderGoogle, status.ProviderType)
	assert.Equal(t, veidtypes.GetSSOScoringWeight(veidtypes.SSOProviderGoogle), status.ScoreContribution)
}
