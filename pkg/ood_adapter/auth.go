// Package ood_adapter implements the Open OnDemand integration adapter for VirtEngine.
//
// VE-918: Open OnDemand using Waldur - VEID SSO authentication support.
package ood_adapter

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Error messages
const errOIDCNotInitialized = "OIDC endpoints not initialized"

// VEIDAuthClient implements VEID SSO authentication via OIDC
type VEIDAuthClient struct {
	issuer       string
	clientID     string
	clientSecret string
	httpClient   *http.Client
	endpoints    *oidcEndpoints
}

// oidcEndpoints contains OIDC endpoint URLs
type oidcEndpoints struct {
	Authorization string `json:"authorization_endpoint"`
	Token         string `json:"token_endpoint"`
	UserInfo      string `json:"userinfo_endpoint"`
	Revocation    string `json:"revocation_endpoint"`
	JWKSURI       string `json:"jwks_uri"`
}

// NewVEIDAuthClient creates a new VEID auth client
func NewVEIDAuthClient(issuer, clientID, clientSecret string) *VEIDAuthClient {
	return &VEIDAuthClient{
		issuer:       issuer,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Initialize fetches OIDC configuration from the issuer
func (c *VEIDAuthClient) Initialize(ctx context.Context) error {
	wellKnownURL := strings.TrimSuffix(c.issuer, "/") + "/.well-known/openid-configuration"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnownURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch OIDC configuration: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OIDC configuration request failed with status: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&c.endpoints); err != nil {
		return fmt.Errorf("failed to decode OIDC configuration: %w", err)
	}

	return nil
}

// GetAuthorizationURL gets the authorization URL for OIDC flow
func (c *VEIDAuthClient) GetAuthorizationURL(state string, redirectURI string) string {
	if c.endpoints == nil {
		return ""
	}

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", c.clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", "openid profile email veid")
	params.Set("state", state)

	// Generate PKCE challenge
	nonce, _ := generateRandomString(32)
	params.Set("nonce", nonce)

	return c.endpoints.Authorization + "?" + params.Encode()
}

// ExchangeCodeForToken exchanges an authorization code for tokens
func (c *VEIDAuthClient) ExchangeCodeForToken(ctx context.Context, code string, redirectURI string) (*VEIDToken, error) {
	if c.endpoints == nil {
		return nil, errors.New(errOIDCNotInitialized)
	}

	// Note: code is sensitive, never log it
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)

	return c.doTokenRequest(ctx, data)
}

// RefreshToken refreshes an expired token
func (c *VEIDAuthClient) RefreshToken(ctx context.Context, refreshToken string) (*VEIDToken, error) {
	if c.endpoints == nil {
		return nil, errors.New(errOIDCNotInitialized)
	}

	// Note: refreshToken is sensitive, never log it
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)

	return c.doTokenRequest(ctx, data)
}

// ValidateToken validates a token and returns user info
func (c *VEIDAuthClient) ValidateToken(ctx context.Context, accessToken string) (*VEIDToken, error) {
	if c.endpoints == nil {
		return nil, errors.New(errOIDCNotInitialized)
	}

	// Note: accessToken is sensitive, never log it
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoints.UserInfo, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed with status: %d", resp.StatusCode)
	}

	var userInfo struct {
		Sub           string  `json:"sub"`
		VEIDAddress   string  `json:"veid_address"`
		IdentityScore float64 `json:"identity_score"`
		Email         string  `json:"email"`
		Name          string  `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo: %w", err)
	}

	return &VEIDToken{
		AccessToken:   accessToken,
		TokenType:     "Bearer",
		VEIDAddress:   userInfo.VEIDAddress,
		IdentityScore: userInfo.IdentityScore,
		ExpiresAt:     time.Now().Add(1 * time.Hour), // Assume 1 hour validity
	}, nil
}

// RevokeToken revokes a token
func (c *VEIDAuthClient) RevokeToken(ctx context.Context, token string) error {
	if c.endpoints == nil || c.endpoints.Revocation == "" {
		return nil // Revocation not supported
	}

	// Note: token is sensitive, never log it
	data := url.Values{}
	data.Set("token", token)
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoints.Revocation, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revocation request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("revocation request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("revocation failed with status: %d", resp.StatusCode)
	}

	return nil
}

// doTokenRequest performs a token request
func (c *VEIDAuthClient) doTokenRequest(ctx context.Context, data url.Values) (*VEIDToken, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoints.Token, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
		IDToken      string `json:"id_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	token := &VEIDToken{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		RefreshToken: tokenResp.RefreshToken,
		Scope:        tokenResp.Scope,
	}

	// Extract VEID info from token (would normally parse ID token)
	if validatedToken, err := c.ValidateToken(ctx, token.AccessToken); err == nil {
		token.VEIDAddress = validatedToken.VEIDAddress
		token.IdentityScore = validatedToken.IdentityScore
	}

	return token, nil
}

// generateRandomString generates a random string of the specified length
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes)[:length], nil
}

// SessionTokenManager manages session tokens securely
type SessionTokenManager struct {
	tokens map[string]*VEIDToken
}

// NewSessionTokenManager creates a new session token manager
func NewSessionTokenManager() *SessionTokenManager {
	return &SessionTokenManager{
		tokens: make(map[string]*VEIDToken),
	}
}

// StoreToken stores a token (never logs the token)
func (m *SessionTokenManager) StoreToken(veidAddress string, token *VEIDToken) {
	// Note: Never log tokens or credentials
	m.tokens[veidAddress] = token
}

// GetToken retrieves a token (never logs the token)
func (m *SessionTokenManager) GetToken(veidAddress string) *VEIDToken {
	return m.tokens[veidAddress]
}

// RemoveToken removes a token
func (m *SessionTokenManager) RemoveToken(veidAddress string) {
	delete(m.tokens, veidAddress)
}

// IsTokenValid checks if a token is valid
func (m *SessionTokenManager) IsTokenValid(veidAddress string) bool {
	token := m.tokens[veidAddress]
	return token != nil && token.IsValid()
}

// CleanupExpiredTokens removes expired tokens
func (m *SessionTokenManager) CleanupExpiredTokens() {
	for addr, token := range m.tokens {
		if token.IsExpired() {
			delete(m.tokens, addr)
		}
	}
}
