package ood_adapter_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ood "github.com/virtengine/virtengine/pkg/ood_adapter"
)

const (
	// wellKnownConfigPath is the OIDC discovery endpoint path.
	wellKnownConfigPath = "/.well-known/openid-configuration"
	// tokenEndpointPath is the OAuth2 token endpoint path.
	tokenEndpointPath = "/token"
	// testJWTTokenVEID1User is a test JWT token for veid1user with identity score 0.95.
	//nolint:gosec // G101: test JWT token, not a real credential
	testJWTTokenVEID1User = "eyJhbGciOiJSUzI1NiJ9.eyJ2ZWlkX2FkZHJlc3MiOiJ2ZWlkMXVzZXIiLCJpZGVudGl0eV9zY29yZSI6MC45NX0.sig"
	// testJWTTokenVEID1UserNoScore is a test JWT token for veid1user without identity score.
	//nolint:gosec // G101: test JWT token, not a real credential
	testJWTTokenVEID1UserNoScore = "eyJhbGciOiJSUzI1NiJ9.eyJ2ZWlkX2FkZHJlc3MiOiJ2ZWlkMXVzZXIifQ.sig"
)

// TestOAuth2ConfigDefaults tests default OAuth2 configuration.
func TestOAuth2ConfigDefaults(t *testing.T) {
	config := ood.DefaultOAuth2Config()

	require.NotNil(t, config)
	require.Contains(t, config.Scopes, "openid")
	require.Contains(t, config.Scopes, "veid")
	require.True(t, config.UsePKCE)
	require.Equal(t, 5*time.Minute, config.TokenRefreshThreshold)
}

// TestOAuth2TokenManagerCreation tests token manager creation.
func TestOAuth2TokenManagerCreation(t *testing.T) {
	config := &ood.OAuth2Config{
		IssuerURL:    "https://auth.example.com",
		ClientID:     "client-123",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost/callback",
		Scopes:       []string{"openid", "profile"},
	}

	manager := ood.NewOAuth2TokenManager(config)
	require.NotNil(t, manager)
}

// TestOAuth2TokenManagerInitialize tests OIDC discovery.
func TestOAuth2TokenManagerInitialize(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == wellKnownConfigPath {
				w.Header().Set("Content-Type", "application/json")
				baseURL := "http://" + r.Host
				_, _ = w.Write([]byte(`{
					"issuer":"` + baseURL + `",
					"authorization_endpoint":"` + baseURL + `/authorize",
					"token_endpoint":"` + baseURL + `/token",
					"userinfo_endpoint":"` + baseURL + `/userinfo"
				}`))
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		config := &ood.OAuth2Config{
			IssuerURL: server.URL,
			ClientID:  "client",
		}

		manager := ood.NewOAuth2TokenManager(config)
		ctx := context.Background()

		err := manager.Initialize(ctx)
		require.NoError(t, err)
	})

	t.Run("discovery failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		config := &ood.OAuth2Config{
			IssuerURL: server.URL,
		}

		manager := ood.NewOAuth2TokenManager(config)
		ctx := context.Background()

		err := manager.Initialize(ctx)
		require.Error(t, err)
	})
}

// TestOAuth2TokenManagerGenerateAuthURL tests authorization URL generation.
func TestOAuth2TokenManagerGenerateAuthURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		baseURL := "http://" + r.Host
		_, _ = w.Write([]byte(`{
			"authorization_endpoint":"` + baseURL + `/authorize",
			"token_endpoint":"` + baseURL + `/token"
		}`))
	}))
	defer server.Close()

	config := &ood.OAuth2Config{
		IssuerURL:   server.URL,
		ClientID:    "client-123",
		RedirectURL: "http://localhost/callback",
		Scopes:      []string{"openid", "profile", "veid"},
		UsePKCE:     true,
	}

	manager := ood.NewOAuth2TokenManager(config)
	ctx := context.Background()

	err := manager.Initialize(ctx)
	require.NoError(t, err)

	authURL, verifier, err := manager.GenerateAuthorizationURL("test-state")
	require.NoError(t, err)
	require.NotEmpty(t, authURL)
	require.NotEmpty(t, verifier)

	require.Contains(t, authURL, "response_type=code")
	require.Contains(t, authURL, "client_id=client-123")
	require.Contains(t, authURL, "redirect_uri=")
	require.Contains(t, authURL, "state=test-state")
	require.Contains(t, authURL, "code_challenge=")
	require.Contains(t, authURL, "code_challenge_method=S256")
}

// TestOAuth2TokenManagerExchangeCode tests code exchange.
func TestOAuth2TokenManagerExchangeCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == wellKnownConfigPath {
			w.Header().Set("Content-Type", "application/json")
			baseURL := "http://" + r.Host
			_, _ = w.Write([]byte(`{
				"authorization_endpoint":"` + baseURL + `/authorize",
				"token_endpoint":"` + baseURL + `/token",
				"userinfo_endpoint":"` + baseURL + `/userinfo"
			}`))
			return
		}

		if r.URL.Path == tokenEndpointPath && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			//nolint:gosec // G101: test JWT token, not a real credential
			idToken := testJWTTokenVEID1User
			_, _ = w.Write([]byte(`{
				"access_token":"access-token-123",
				"token_type":"Bearer",
				"expires_in":3600,
				"refresh_token":"refresh-token-123",
				"id_token":"` + idToken + `"
			}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	config := &ood.OAuth2Config{
		IssuerURL:   server.URL,
		ClientID:    "client",
		RedirectURL: "http://localhost/callback",
		UsePKCE:     true,
	}

	manager := ood.NewOAuth2TokenManager(config)
	ctx := context.Background()

	err := manager.Initialize(ctx)
	require.NoError(t, err)

	token, err := manager.ExchangeCode(ctx, "auth-code", "verifier")
	require.NoError(t, err)
	require.NotNil(t, token)
	require.NotEmpty(t, token.AccessToken)
	require.NotEmpty(t, token.RefreshToken)
	require.Equal(t, "veid1user", token.VEIDAddress)
}

// TestOAuth2TokenManagerRefreshToken tests token refresh.
func TestOAuth2TokenManagerRefreshToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == wellKnownConfigPath {
			w.Header().Set("Content-Type", "application/json")
			baseURL := "http://" + r.Host
			_, _ = w.Write([]byte(`{
				"authorization_endpoint":"` + baseURL + `/authorize",
				"token_endpoint":"` + baseURL + `/token",
				"userinfo_endpoint":"` + baseURL + `/userinfo"
			}`))
			return
		}

		if r.URL.Path == tokenEndpointPath {
			w.Header().Set("Content-Type", "application/json")
			//nolint:gosec // G101: test JWT token, not a real credential
			idToken := testJWTTokenVEID1User
			_, _ = w.Write([]byte(`{
				"access_token":"new-access-token",
				"token_type":"Bearer",
				"expires_in":3600,
				"refresh_token":"new-refresh-token",
				"id_token":"` + idToken + `"
			}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	config := &ood.OAuth2Config{
		IssuerURL:   server.URL,
		ClientID:    "client",
		RedirectURL: "http://localhost/callback",
	}

	manager := ood.NewOAuth2TokenManager(config)
	ctx := context.Background()

	err := manager.Initialize(ctx)
	require.NoError(t, err)

	// Exchange code first to get initial token
	token, err := manager.ExchangeCode(ctx, "code", "")
	require.NoError(t, err)

	// Refresh token
	err = manager.RefreshToken(ctx, token.VEIDAddress)
	require.NoError(t, err)

	refreshed := manager.GetToken(token.VEIDAddress)
	require.NotNil(t, refreshed)
	require.Equal(t, 1, refreshed.RefreshCount)
}

// TestOAuth2TokenManagerGetValidToken tests getting valid tokens.
func TestOAuth2TokenManagerGetValidToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == wellKnownConfigPath {
			w.Header().Set("Content-Type", "application/json")
			baseURL := "http://" + r.Host
			_, _ = w.Write([]byte(`{
				"authorization_endpoint":"` + baseURL + `/authorize",
				"token_endpoint":"` + baseURL + `/token",
				"userinfo_endpoint":"` + baseURL + `/userinfo"
			}`))
			return
		}

		if r.URL.Path == tokenEndpointPath {
			w.Header().Set("Content-Type", "application/json")
			//nolint:gosec // G101: test JWT token, not a real credential
			idToken := testJWTTokenVEID1UserNoScore
			_, _ = w.Write([]byte(`{
				"access_token":"token",
				"expires_in":3600,
				"refresh_token":"refresh",
				"id_token":"` + idToken + `"
			}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	config := &ood.OAuth2Config{
		IssuerURL:             server.URL,
		ClientID:              "client",
		TokenRefreshThreshold: 5 * time.Minute,
	}

	manager := ood.NewOAuth2TokenManager(config)
	ctx := context.Background()

	err := manager.Initialize(ctx)
	require.NoError(t, err)

	// Exchange code first
	_, err = manager.ExchangeCode(ctx, "code", "")
	require.NoError(t, err)

	// Get valid token
	token, err := manager.GetValidToken(ctx, "veid1user")
	require.NoError(t, err)
	require.NotNil(t, token)

	// Get token for unknown user
	_, err = manager.GetValidToken(ctx, "unknown")
	require.ErrorIs(t, err, ood.ErrInvalidToken)
}

// TestOAuth2TokenManagerTokenStorage tests token storage operations.
func TestOAuth2TokenManagerTokenStorage(t *testing.T) {
	config := &ood.OAuth2Config{
		IssuerURL: "https://auth.example.com",
		ClientID:  "client",
	}

	manager := ood.NewOAuth2TokenManager(config)

	// Create managed token
	token := &ood.ManagedToken{
		VEIDToken: &ood.VEIDToken{
			AccessToken:   "access",
			RefreshToken:  "refresh",
			ExpiresAt:     time.Now().Add(1 * time.Hour),
			VEIDAddress:   "veid1user",
			IdentityScore: 0.95,
		},
		LastRefresh: time.Now(),
	}

	// Store token
	manager.SetToken("veid1user", token)

	// Get token
	retrieved := manager.GetToken("veid1user")
	require.NotNil(t, retrieved)
	require.Equal(t, "access", retrieved.AccessToken)

	// Remove token
	manager.RemoveToken("veid1user")

	removed := manager.GetToken("veid1user")
	require.Nil(t, removed)
}

// TestOAuth2TokenManagerRevokeToken tests token revocation.
func TestOAuth2TokenManagerRevokeToken(t *testing.T) {
	revokedTokens := make([]string, 0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == wellKnownConfigPath {
			w.Header().Set("Content-Type", "application/json")
			baseURL := "http://" + r.Host
			_, _ = w.Write([]byte(`{
				"authorization_endpoint":"` + baseURL + `/authorize",
				"token_endpoint":"` + baseURL + `/token",
				"revocation_endpoint":"` + baseURL + `/revoke"
			}`))
			return
		}

		if r.URL.Path == tokenEndpointPath {
			w.Header().Set("Content-Type", "application/json")
			//nolint:gosec // G101: test JWT token, not a real credential
			idToken := testJWTTokenVEID1UserNoScore
			_, _ = w.Write([]byte(`{
				"access_token":"token-to-revoke",
				"expires_in":3600,
				"refresh_token":"refresh",
				"id_token":"` + idToken + `"
			}`))
			return
		}

		if r.URL.Path == "/revoke" && r.Method == http.MethodPost {
			_ = r.ParseForm()
			revokedTokens = append(revokedTokens, r.FormValue("token"))
			w.WriteHeader(http.StatusOK)
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	config := &ood.OAuth2Config{
		IssuerURL: server.URL,
		ClientID:  "client",
	}

	manager := ood.NewOAuth2TokenManager(config)
	ctx := context.Background()

	err := manager.Initialize(ctx)
	require.NoError(t, err)

	// Exchange code
	token, err := manager.ExchangeCode(ctx, "code", "")
	require.NoError(t, err)

	// Revoke token
	err = manager.RevokeToken(ctx, token.VEIDAddress)
	require.NoError(t, err)

	// Token should be removed
	retrieved := manager.GetToken(token.VEIDAddress)
	require.Nil(t, retrieved)

	// Revocation endpoint should have been called
	require.NotEmpty(t, revokedTokens)
}

// TestOAuth2TokenManagerCleanupExpired tests expired token cleanup.
func TestOAuth2TokenManagerCleanupExpired(t *testing.T) {
	config := &ood.OAuth2Config{
		IssuerURL: "https://auth.example.com",
		ClientID:  "client",
	}

	manager := ood.NewOAuth2TokenManager(config)

	// Add valid token
	manager.SetToken("valid", &ood.ManagedToken{
		VEIDToken: &ood.VEIDToken{
			AccessToken: "valid",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		},
	})

	// Add expired token without refresh
	manager.SetToken("expired", &ood.ManagedToken{
		VEIDToken: &ood.VEIDToken{
			AccessToken: "expired",
			ExpiresAt:   time.Now().Add(-1 * time.Hour),
		},
	})

	// Add expired token with refresh (should not be cleaned)
	manager.SetToken("expired-with-refresh", &ood.ManagedToken{
		VEIDToken: &ood.VEIDToken{
			AccessToken:  "expired",
			RefreshToken: "refresh",
			ExpiresAt:    time.Now().Add(-1 * time.Hour),
		},
	})

	count := manager.CleanupExpired()
	require.Equal(t, 1, count)

	require.NotNil(t, manager.GetToken("valid"))
	require.Nil(t, manager.GetToken("expired"))
	require.NotNil(t, manager.GetToken("expired-with-refresh"))
}

// TestOAuth2TokenManagerStartStop tests background refresh.
func TestOAuth2TokenManagerStartStop(t *testing.T) {
	config := &ood.OAuth2Config{
		IssuerURL:             "https://auth.example.com",
		ClientID:              "client",
		TokenRefreshThreshold: time.Millisecond,
	}

	manager := ood.NewOAuth2TokenManager(config)

	// Should not panic on multiple starts
	manager.Start()
	manager.Start()

	// Should not panic on stop
	manager.Stop()
	manager.Stop()
}

// TestOAuth2ProviderAdapter tests the VEIDAuthProvider adapter.
func TestOAuth2ProviderAdapter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == wellKnownConfigPath {
			w.Header().Set("Content-Type", "application/json")
			baseURL := "http://" + r.Host
			_, _ = w.Write([]byte(`{
				"authorization_endpoint":"` + baseURL + `/authorize",
				"token_endpoint":"` + baseURL + `/token",
				"userinfo_endpoint":"` + baseURL + `/userinfo",
				"revocation_endpoint":"` + baseURL + `/revoke"
			}`))
			return
		}

		if r.URL.Path == tokenEndpointPath {
			w.Header().Set("Content-Type", "application/json")
			//nolint:gosec // G101: test JWT token, not a real credential
			idToken := testJWTTokenVEID1User
			_, _ = w.Write([]byte(`{
				"access_token":"access-token",
				"expires_in":3600,
				"refresh_token":"refresh-token",
				"id_token":"` + idToken + `"
			}`))
			return
		}

		if r.URL.Path == "/userinfo" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"veid_address":"veid1user",
				"identity_score":0.95
			}`))
			return
		}

		if r.URL.Path == "/revoke" {
			w.WriteHeader(http.StatusOK)
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	config := &ood.OAuth2Config{
		IssuerURL:   server.URL,
		ClientID:    "client",
		RedirectURL: "http://localhost/callback",
	}

	manager := ood.NewOAuth2TokenManager(config)
	ctx := context.Background()

	err := manager.Initialize(ctx)
	require.NoError(t, err)

	adapter := ood.NewOAuth2ProviderAdapter(manager)

	t.Run("GetAuthorizationURL", func(t *testing.T) {
		url := adapter.GetAuthorizationURL("state", "http://localhost/callback")
		require.NotEmpty(t, url)
		require.Contains(t, url, "state=state")
	})

	t.Run("ExchangeCodeForToken", func(t *testing.T) {
		token, err := adapter.ExchangeCodeForToken(ctx, "code", "http://localhost/callback")
		require.NoError(t, err)
		require.NotNil(t, token)
		require.NotEmpty(t, token.AccessToken)
	})

	t.Run("ValidateToken", func(t *testing.T) {
		token, err := adapter.ValidateToken(ctx, "some-token")
		require.NoError(t, err)
		require.NotNil(t, token)
		require.Equal(t, "veid1user", token.VEIDAddress)
	})

	t.Run("RefreshToken", func(t *testing.T) {
		// First exchange code to have a token
		_, err := adapter.ExchangeCodeForToken(ctx, "code", "callback")
		require.NoError(t, err)

		token, err := adapter.RefreshToken(ctx, "refresh-token")
		require.NoError(t, err)
		require.NotNil(t, token)
	})

	t.Run("RevokeToken", func(t *testing.T) {
		err := adapter.RevokeToken(ctx, "access-token")
		require.NoError(t, err)
	})
}

// TestManagedTokenOnRefreshCallback tests refresh callback.
func TestManagedTokenOnRefreshCallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == wellKnownConfigPath {
			w.Header().Set("Content-Type", "application/json")
			baseURL := "http://" + r.Host
			_, _ = w.Write([]byte(`{
				"authorization_endpoint":"` + baseURL + `/authorize",
				"token_endpoint":"` + baseURL + `/token"
			}`))
			return
		}

		if r.URL.Path == tokenEndpointPath {
			w.Header().Set("Content-Type", "application/json")
			//nolint:gosec // G101: test JWT token, not a real credential
			idToken := testJWTTokenVEID1UserNoScore
			_, _ = w.Write([]byte(`{
				"access_token":"token",
				"expires_in":3600,
				"refresh_token":"refresh",
				"id_token":"` + idToken + `"
			}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	config := &ood.OAuth2Config{
		IssuerURL: server.URL,
		ClientID:  "client",
	}

	manager := ood.NewOAuth2TokenManager(config)
	ctx := context.Background()

	err := manager.Initialize(ctx)
	require.NoError(t, err)

	// Exchange code
	token, err := manager.ExchangeCode(ctx, "code", "")
	require.NoError(t, err)

	// Set callback
	callbackCalled := false
	token.OnRefresh = func(t *ood.ManagedToken) {
		callbackCalled = true
	}
	manager.SetToken(token.VEIDAddress, token)

	// Refresh
	err = manager.RefreshToken(ctx, token.VEIDAddress)
	require.NoError(t, err)
	require.True(t, callbackCalled)
}

// TestIDTokenParsing tests ID token claims parsing.
func TestIDTokenParsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == wellKnownConfigPath {
			w.Header().Set("Content-Type", "application/json")
			baseURL := "http://" + r.Host
			_, _ = w.Write([]byte(`{
				"authorization_endpoint":"` + baseURL + `/authorize",
				"token_endpoint":"` + baseURL + `/token"
			}`))
			return
		}

		if r.URL.Path == tokenEndpointPath {
			w.Header().Set("Content-Type", "application/json")
			// Valid ID token with claims
			//nolint:gosec // G101: test JWT token, not a real credential
			idToken := "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ1c2VyMTIzIiwidmVpZF9hZGRyZXNzIjoidmVpZDFjdXN0b211c2VyIiwiaWRlbnRpdHlfc2NvcmUiOjAuOTksImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsIm5hbWUiOiJUZXN0IFVzZXIifQ.signature"
			_, _ = w.Write([]byte(`{
				"access_token":"access",
				"expires_in":3600,
				"id_token":"` + idToken + `"
			}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	config := &ood.OAuth2Config{
		IssuerURL: server.URL,
		ClientID:  "client",
	}

	manager := ood.NewOAuth2TokenManager(config)
	ctx := context.Background()

	err := manager.Initialize(ctx)
	require.NoError(t, err)

	token, err := manager.ExchangeCode(ctx, "code", "")
	require.NoError(t, err)

	require.Equal(t, "veid1customuser", token.VEIDAddress)
	require.Equal(t, 0.99, token.IdentityScore)
}

