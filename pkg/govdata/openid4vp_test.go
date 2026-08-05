package govdata

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestVerifyOpenID4VPVerifiesSignedNonceBoundSDJWT(t *testing.T) {
	now := time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC)
	request := validDigitalIDRequest(now)
	request.ProviderID = "openid4vp-test"
	public, private, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	disclosure := encodeDisclosure(t, "salt", "given_name", "Ada")
	digest := sha256.Sum256([]byte(disclosure))
	token := signedVPToken(t, private, request, now, []string{base64.RawURLEncoding.EncodeToString(digest[:])})
	config := testVerifierConfig(request)
	identity, err := VerifyOpenID4VP(context.Background(), request, OpenID4VPAuthorizationResponse{State: request.State, VPToken: token, Disclosures: []string{disclosure}}, config, staticIssuerKey{public}, activeCredentialStatus{}, testAssuranceMapper{}, now)
	require.NoError(t, err)
	require.Equal(t, "Ada", identity.Claims["given_name"])
	_, err = VerifyOpenID4VP(context.Background(), request, OpenID4VPAuthorizationResponse{State: request.State, VPToken: token, Disclosures: []string{disclosure}}, config, staticIssuerKey{public}, activeCredentialStatus{}, testAssuranceMapper{}, now)
	require.ErrorIs(t, err, ErrOpenID4VPVerification)
}

func TestVerifyOpenID4VPFailsClosedForUnsignedOrWrongNonceToken(t *testing.T) {
	now := time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC)
	request := validDigitalIDRequest(now)
	public, private, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	disclosure := encodeDisclosure(t, "salt", "given_name", "Ada")
	digest := sha256.Sum256([]byte(disclosure))
	token := signedVPToken(t, private, request, now, []string{base64.RawURLEncoding.EncodeToString(digest[:])})
	_, err = VerifyOpenID4VP(context.Background(), request, OpenID4VPAuthorizationResponse{State: request.State, VPToken: token, Disclosures: []string{disclosure}}, testVerifierConfig(request), staticIssuerKey{public}, activeCredentialStatus{}, testAssuranceMapper{}, now)
	require.NoError(t, err)
	request.Nonce = "changed-nonce-which-is-long-enough"
	_, err = VerifyOpenID4VP(context.Background(), request, OpenID4VPAuthorizationResponse{State: request.State, VPToken: token, Disclosures: []string{disclosure}}, testVerifierConfig(request), staticIssuerKey{public}, activeCredentialStatus{}, testAssuranceMapper{}, now)
	require.ErrorIs(t, err, ErrOpenID4VPVerification)
	_, err = VerifyOpenID4VP(context.Background(), request, OpenID4VPAuthorizationResponse{State: request.State, VPToken: "eyJhbGciOiJub2luIn0.eyJzdWIiOiJ4In0.", Disclosures: []string{disclosure}}, testVerifierConfig(request), staticIssuerKey{public}, activeCredentialStatus{}, testAssuranceMapper{}, now)
	require.ErrorIs(t, err, ErrOpenID4VPVerification)
}

func TestVerifyOpenID4VPRejectsUncommittedDisclosure(t *testing.T) {
	now := time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC)
	request := validDigitalIDRequest(now)
	public, private, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	disclosure := encodeDisclosure(t, "salt", "given_name", "Mallory")
	digest := sha256.Sum256([]byte("different"))
	token := signedVPToken(t, private, request, now, []string{base64.RawURLEncoding.EncodeToString(digest[:])})
	_, err = VerifyOpenID4VP(context.Background(), request, OpenID4VPAuthorizationResponse{State: request.State, VPToken: token, Disclosures: []string{disclosure}}, testVerifierConfig(request), staticIssuerKey{public}, activeCredentialStatus{}, testAssuranceMapper{}, now)
	require.ErrorIs(t, err, ErrOpenID4VPVerification)
}

func signedVPToken(t *testing.T, private ed25519.PrivateKey, request DigitalIDAuthorizationRequest, now time.Time, digests []string) string {
	t.Helper()
	claims := jwt.MapClaims{"iss": "https://issuer.example", "sub": "pairwise-subject", "aud": request.ClientID, "nonce": request.Nonce, "iat": now.Unix(), "exp": now.Add(time.Minute).Unix(), "assurance": "high", "credential_status": map[string]string{"id": "https://issuer.example/status/1"}, "_sd": digests, "_sd_alg": "sha-256"}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["kid"] = "key-1"
	signed, err := token.SignedString(private)
	require.NoError(t, err)
	return signed
}

func encodeDisclosure(t *testing.T, salt, name, value string) string {
	t.Helper()
	raw, err := json.Marshal([]string{salt, name, value})
	require.NoError(t, err)
	return base64.RawURLEncoding.EncodeToString(raw)
}

type staticIssuerKey struct{ key crypto.PublicKey }

func (s staticIssuerKey) ResolveIssuerKey(context.Context, string, string) (crypto.PublicKey, error) {
	return s.key, nil
}

type activeCredentialStatus struct{}

func (activeCredentialStatus) ResolveCredentialStatus(context.Context, string, string, json.RawMessage) (CredentialStatus, error) {
	return CredentialStatusActive, nil
}

type testAssuranceMapper struct{}

func (testAssuranceMapper) MapAssurance(value string) (AssuranceLevel, bool) {
	return AssuranceLevel(value), AssuranceLevel(value).Valid()
}

type testReplayGuard struct {
	mu     sync.Mutex
	states map[string]struct{}
}

func (g *testReplayGuard) ConsumeAuthorizationState(_ context.Context, state string, _ time.Time) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, exists := g.states[state]; exists {
		return errors.New("replayed")
	}
	g.states[state] = struct{}{}
	return nil
}
func testVerifierConfig(request DigitalIDAuthorizationRequest) OpenID4VPVerifierConfig {
	return OpenID4VPVerifierConfig{ProviderID: request.ProviderID, Issuer: "https://issuer.example", ReplayGuard: &testReplayGuard{states: map[string]struct{}{}}}
}
