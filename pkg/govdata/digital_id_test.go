package govdata

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDigitalIDAuthorizationRequestRequiresStrongCallbackBindings(t *testing.T) {
	now := time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC)
	req := validDigitalIDRequest(now)
	require.NoError(t, req.Validate(now))
	req.CodeChallengeMethod = "plain"
	require.ErrorContains(t, req.Validate(now), "PKCE S256")
}

func TestValidateDigitalIDIdentityFailsClosedForRevocationAndAssurance(t *testing.T) {
	now := time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC)
	req := validDigitalIDRequest(now)
	identity := DigitalIDIdentity{ProviderID: req.ProviderID, Subject: "pairwise-subject", Assurance: AssuranceLevelHigh, Claims: map[string]string{"given_name": "Ada"}, IssuedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute), Status: CredentialStatusActive}
	require.NoError(t, ValidateIdentity(req, identity, now))
	identity.Status = CredentialStatusRevoked
	require.ErrorContains(t, ValidateIdentity(req, identity, now), "not active")
	identity.Status = CredentialStatusActive
	identity.Assurance = AssuranceLevelLow
	require.ErrorContains(t, ValidateIdentity(req, identity, now), "assurance")
}

func TestUnavailableDigitalIDProviderNeverSimulatesSuccess(t *testing.T) {
	p := UnavailableDigitalIDProvider{ID: "eidas"}
	_, err := p.BeginAuthorization(context.Background(), validDigitalIDRequest(time.Now().UTC()))
	require.ErrorIs(t, err, ErrDigitalIDProtocolUnavailable)
}

func TestLegacyDocumentAdaptersExposeOnlyFailClosedDigitalIDBridge(t *testing.T) {
	var dvs DigitalIDCapableAdapter = (*dvsDMVAdapter)(nil)
	var eidas DigitalIDCapableAdapter = (*eidasAdapter)(nil)
	for _, adapter := range []DigitalIDCapableAdapter{dvs, eidas} {
		_, err := adapter.DigitalIDProvider().BeginAuthorization(context.Background(), validDigitalIDRequest(time.Now().UTC()))
		require.ErrorIs(t, err, ErrDigitalIDProtocolUnavailable)
	}
}

func validDigitalIDRequest(now time.Time) DigitalIDAuthorizationRequest {
	return DigitalIDAuthorizationRequest{ProviderID: "eidas", ClientID: "client", RedirectURI: "https://wallet.example/callback", State: "1234567890123456", Nonce: "abcdefghijklmnopqrstuvwxyz", CodeChallenge: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~", CodeChallengeMethod: PKCEMethodS256, ConsentID: "consent-1", RequestedClaims: []string{"given_name"}, MinimumAssurance: AssuranceLevelSubstantial, ExpiresAt: now.Add(5 * time.Minute)}
}
