// Package oidc provides OIDC token verification for the VEID SSO verification service.
package oidc

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Test Fixtures
// ============================================================================

type testOIDCServer struct {
	server           *httptest.Server
	privateKey       *rsa.PrivateKey
	keyID            string
	issuer           string
	discoveryHandler http.HandlerFunc
	jwksHandler      http.HandlerFunc
	tokenHandler     http.HandlerFunc
}

func newTestOIDCServer(t *testing.T) *testOIDCServer {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyID := "test-key-1"

	ts := &testOIDCServer{
		privateKey: privateKey,
		keyID:      keyID,
	}

	// Set up HTTP handlers
	mux := http.NewServeMux()

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		if ts.discoveryHandler != nil {
			ts.discoveryHandler(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"issuer":                 ts.issuer,
			"authorization_endpoint": ts.issuer + "/authorize",
			"token_endpoint":         ts.issuer + "/token",
			"userinfo_endpoint":      ts.issuer + "/userinfo",
			"jwks_uri":               ts.issuer + "/.well-known/jwks.json",
			"response_types_supported": []string{"code", "token", "id_token"},
			"subject_types_supported":  []string{"public"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})

	mux.HandleFunc("/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		if ts.jwksHandler != nil {
			ts.jwksHandler(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ts.getJWKS())
	})

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if ts.tokenHandler != nil {
			ts.tokenHandler(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "test-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"id_token":     ts.generateIDToken(t, map[string]interface{}{
				"sub":            "user-123",
				"email":          "test@example.com",
				"email_verified": true,
				"nonce":          "test-nonce",
			}),
		})
	})

	server := httptest.NewServer(mux)
	ts.server = server
	ts.issuer = server.URL

	return ts
}

func (ts *testOIDCServer) Close() {
	ts.server.Close()
}

func (ts *testOIDCServer) getJWKS() map[string]interface{} {
	return map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"kid": ts.keyID,
				"n":   base64.RawURLEncoding.EncodeToString(ts.privateKey.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(ts.privateKey.E)).Bytes()),
			},
		},
	}
}

func (ts *testOIDCServer) generateIDToken(t *testing.T, claims map[string]interface{}) string {
	// Add standard claims
	now := time.Now()
	if _, ok := claims["iss"]; !ok {
		claims["iss"] = ts.issuer
	}
	if _, ok := claims["aud"]; !ok {
		claims["aud"] = "test-client-id"
	}
	if _, ok := claims["exp"]; !ok {
		claims["exp"] = now.Add(time.Hour).Unix()
	}
	if _, ok := claims["iat"]; !ok {
		claims["iat"] = now.Unix()
	}

	// Encode header
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
		"kid": ts.keyID,
	}
	headerBytes, _ := json.Marshal(header)
	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)

	// Encode payload
	payloadBytes, _ := json.Marshal(claims)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// Sign
	signedContent := headerB64 + "." + payloadB64
	signature := ts.sign(t, signedContent)
	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	return signedContent + "." + signatureB64
}

func (ts *testOIDCServer) sign(t *testing.T, content string) []byte {
	hashed := sha256.Sum256([]byte(content))
	signature, err := rsa.SignPKCS1v15(rand.Reader, ts.privateKey, crypto.SHA256, hashed[:])
	require.NoError(t, err)
	return signature
}

// ============================================================================
// JWKSManager Tests
// ============================================================================

func TestJWKSManager_GetKey(t *testing.T) {
	ts := newTestOIDCServer(t)
	defer ts.Close()

	logger := zerolog.New(zerolog.NewConsoleWriter())
	manager := NewJWKSManager(nil, logger)
	defer manager.Close()

	ctx := context.Background()

	// Test successful key retrieval
	key, err := manager.GetKey(ctx, ts.issuer, ts.issuer+"/.well-known/jwks.json", ts.keyID)
	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, ts.keyID, key.KID)
	assert.Equal(t, "RSA", key.KTY)

	// Test cache hit
	key2, err := manager.GetKey(ctx, ts.issuer, ts.issuer+"/.well-known/jwks.json", ts.keyID)
	require.NoError(t, err)
	assert.Equal(t, key, key2)
}

func TestJWKSManager_GetKey_NotFound(t *testing.T) {
	ts := newTestOIDCServer(t)
	defer ts.Close()

	logger := zerolog.New(zerolog.NewConsoleWriter())
	manager := NewJWKSManager(nil, logger)
	defer manager.Close()

	ctx := context.Background()

	_, err := manager.GetKey(ctx, ts.issuer, ts.issuer+"/.well-known/jwks.json", "nonexistent-key")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrKeyNotFound)
}

func TestJWKSManager_GetKey_FetchError(t *testing.T) {
	logger := zerolog.New(zerolog.NewConsoleWriter())
	config := &JWKSManagerConfig{
		DefaultCacheTTL: time.Hour,
		RefreshInterval: 30 * time.Minute,
		HTTPTimeout:     time.Second,
		MaxRetries:      1,
		RetryBackoff:    10 * time.Millisecond,
	}
	manager := NewJWKSManager(config, logger)
	defer manager.Close()

	ctx := context.Background()

	// Invalid URL should fail
	_, err := manager.GetKey(ctx, "http://invalid.local", "http://invalid.local/.well-known/jwks.json", "key-1")
	assert.Error(t, err)
}

// ============================================================================
// DefaultVerifier Tests
// ============================================================================

func TestDefaultVerifier_IsIssuerAllowed(t *testing.T) {
	ts := newTestOIDCServer(t)
	defer ts.Close()

	logger := zerolog.New(zerolog.NewConsoleWriter())
	config := Config{
		Enabled: true,
		IssuerPolicies: map[string]*IssuerPolicy{
			ts.issuer: {
				Issuer:       ts.issuer,
				ProviderType: veidtypes.SSOProviderGoogle,
				ClientID:     "test-client-id",
				Enabled:      true,
			},
		},
	}

	verifier, err := NewDefaultVerifier(context.Background(), config, nil, logger)
	require.NoError(t, err)
	defer verifier.Close()

	ctx := context.Background()

	// Allowed issuer
	assert.True(t, verifier.IsIssuerAllowed(ctx, ts.issuer))

	// Disallowed issuer
	assert.False(t, verifier.IsIssuerAllowed(ctx, "https://unknown.issuer.com"))
}

func TestDefaultVerifier_GetIssuerPolicy(t *testing.T) {
	ts := newTestOIDCServer(t)
	defer ts.Close()

	logger := zerolog.New(zerolog.NewConsoleWriter())
	policy := &IssuerPolicy{
		Issuer:       ts.issuer,
		ProviderType: veidtypes.SSOProviderGoogle,
		ClientID:     "test-client-id",
		Enabled:      true,
		ScoreWeight:  250,
	}
	config := Config{
		Enabled: true,
		IssuerPolicies: map[string]*IssuerPolicy{
			ts.issuer: policy,
		},
	}

	verifier, err := NewDefaultVerifier(context.Background(), config, nil, logger)
	require.NoError(t, err)
	defer verifier.Close()

	ctx := context.Background()

	// Get existing policy
	got, err := verifier.GetIssuerPolicy(ctx, ts.issuer)
	require.NoError(t, err)
	assert.Equal(t, policy.ClientID, got.ClientID)
	assert.Equal(t, policy.ScoreWeight, got.ScoreWeight)

	// Get non-existent policy
	_, err = verifier.GetIssuerPolicy(ctx, "https://unknown.issuer.com")
	assert.Error(t, err)
}

// ============================================================================
// IssuerPolicy Tests
// ============================================================================

func TestIssuerPolicy_Validate(t *testing.T) {
	tests := []struct {
		name    string
		policy  IssuerPolicy
		wantErr bool
	}{
		{
			name: "valid policy",
			policy: IssuerPolicy{
				Issuer:       "https://accounts.google.com",
				ProviderType: veidtypes.SSOProviderGoogle,
				ClientID:     "client-123",
				ScoreWeight:  250,
			},
			wantErr: false,
		},
		{
			name: "missing issuer",
			policy: IssuerPolicy{
				ProviderType: veidtypes.SSOProviderGoogle,
				ClientID:     "client-123",
			},
			wantErr: true,
		},
		{
			name: "invalid provider type",
			policy: IssuerPolicy{
				Issuer:       "https://accounts.google.com",
				ProviderType: "invalid",
				ClientID:     "client-123",
			},
			wantErr: true,
		},
		{
			name: "missing client ID",
			policy: IssuerPolicy{
				Issuer:       "https://accounts.google.com",
				ProviderType: veidtypes.SSOProviderGoogle,
			},
			wantErr: true,
		},
		{
			name: "score weight too high",
			policy: IssuerPolicy{
				Issuer:       "https://accounts.google.com",
				ProviderType: veidtypes.SSOProviderGoogle,
				ClientID:     "client-123",
				ScoreWeight:  15000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIssuerPolicy_IsEmailDomainAllowed(t *testing.T) {
	policy := IssuerPolicy{
		AllowedEmailDomains: []string{"example.com", "test.org"},
	}

	assert.True(t, policy.IsEmailDomainAllowed("example.com"))
	assert.True(t, policy.IsEmailDomainAllowed("test.org"))
	assert.False(t, policy.IsEmailDomainAllowed("other.com"))

	// Empty allowlist allows all
	emptyPolicy := IssuerPolicy{}
	assert.True(t, emptyPolicy.IsEmailDomainAllowed("any.com"))
}

func TestIssuerPolicy_IsTenantAllowed(t *testing.T) {
	policy := IssuerPolicy{
		AllowedTenants: []string{"tenant-1", "tenant-2"},
	}

	assert.True(t, policy.IsTenantAllowed("tenant-1"))
	assert.True(t, policy.IsTenantAllowed("tenant-2"))
	assert.False(t, policy.IsTenantAllowed("tenant-3"))

	// Empty allowlist allows all
	emptyPolicy := IssuerPolicy{}
	assert.True(t, emptyPolicy.IsTenantAllowed("any-tenant"))
}

// ============================================================================
// VerifiedClaims Tests
// ============================================================================

func TestVerifiedClaims_GetEmailDomain(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "valid email",
			email:    "user@example.com",
			expected: "example.com",
		},
		{
			name:     "subdomain email",
			email:    "user@sub.example.com",
			expected: "sub.example.com",
		},
		{
			name:     "empty email",
			email:    "",
			expected: "",
		},
		{
			name:     "invalid email",
			email:    "not-an-email",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &VerifiedClaims{Email: tt.email}
			assert.Equal(t, tt.expected, claims.GetEmailDomain())
		})
	}
}

// ============================================================================
// Well-Known Issuers Tests
// ============================================================================

func TestWellKnownIssuers(t *testing.T) {
	// Verify well-known issuers are defined
	assert.NotEmpty(t, WellKnownIssuers)

	// Check Google
	googleIssuer, ok := WellKnownIssuers[veidtypes.SSOProviderGoogle]
	assert.True(t, ok)
	assert.Equal(t, "https://accounts.google.com", googleIssuer)

	// Check Microsoft
	microsoftIssuer, ok := WellKnownIssuers[veidtypes.SSOProviderMicrosoft]
	assert.True(t, ok)
	assert.Contains(t, microsoftIssuer, "login.microsoftonline.com")
}
