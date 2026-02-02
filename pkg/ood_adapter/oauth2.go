// Package ood_adapter implements the Open OnDemand integration adapter for VirtEngine.
//
// VE-918: Open OnDemand using Waldur - OAuth2/OIDC token management.
package ood_adapter

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// OAuth2Config contains OAuth2/OIDC configuration.
type OAuth2Config struct {
	// IssuerURL is the OIDC issuer URL
	IssuerURL string `json:"issuer_url"`

	// ClientID is the OAuth2 client ID
	ClientID string `json:"client_id"`

	// ClientSecret is the OAuth2 client secret (never log)
	ClientSecret string `json:"-"`

	// RedirectURL is the OAuth2 redirect URL
	RedirectURL string `json:"redirect_url"`

	// Scopes are the OAuth2 scopes to request
	Scopes []string `json:"scopes"`

	// UsePKCE enables PKCE flow
	UsePKCE bool `json:"use_pkce"`

	// TokenRefreshThreshold is when to refresh tokens before expiry
	TokenRefreshThreshold time.Duration `json:"token_refresh_threshold"`
}

// DefaultOAuth2Config returns default OAuth2 configuration.
func DefaultOAuth2Config() *OAuth2Config {
	return &OAuth2Config{
		Scopes:                []string{"openid", "profile", "email", "veid"},
		UsePKCE:               true,
		TokenRefreshThreshold: 5 * time.Minute,
	}
}

// OAuth2TokenManager manages OAuth2/OIDC tokens with automatic refresh.
type OAuth2TokenManager struct {
	config     *OAuth2Config
	httpClient *http.Client
	endpoints  *oidcEndpoints
	mu         sync.RWMutex
	tokens     map[string]*ManagedToken
	refreshing map[string]bool
	stopCh     chan struct{}
	running    bool
}

// ManagedToken represents a managed OAuth2 token with refresh capability.
type ManagedToken struct {
	// VEIDToken is the underlying token
	*VEIDToken

	// PKCEVerifier is the PKCE verifier (never log)
	PKCEVerifier string `json:"-"`

	// LastRefresh is when the token was last refreshed
	LastRefresh time.Time `json:"last_refresh"`

	// RefreshCount is how many times the token has been refreshed
	RefreshCount int `json:"refresh_count"`

	// OnRefresh is called when token is refreshed
	OnRefresh func(*ManagedToken) `json:"-"`
}

// NewOAuth2TokenManager creates a new OAuth2 token manager.
func NewOAuth2TokenManager(config *OAuth2Config) *OAuth2TokenManager {
	return &OAuth2TokenManager{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		tokens:     make(map[string]*ManagedToken),
		refreshing: make(map[string]bool),
		stopCh:     make(chan struct{}),
	}
}

// Initialize fetches OIDC discovery configuration.
func (m *OAuth2TokenManager) Initialize(ctx context.Context) error {
	wellKnownURL := strings.TrimSuffix(m.config.IssuerURL, "/") + "/.well-known/openid-configuration"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnownURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create discovery request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch OIDC discovery: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OIDC discovery failed with status: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&m.endpoints); err != nil {
		return fmt.Errorf("failed to decode OIDC discovery: %w", err)
	}

	return nil
}

// Start starts the automatic token refresh background process.
func (m *OAuth2TokenManager) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.stopCh = make(chan struct{})
	m.mu.Unlock()

	go m.refreshLoop()
}

// Stop stops the token refresh background process.
func (m *OAuth2TokenManager) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	m.running = false
	close(m.stopCh)
	m.mu.Unlock()
}

// refreshLoop periodically refreshes tokens.
func (m *OAuth2TokenManager) refreshLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.refreshExpiring()
		}
	}
}

// refreshExpiring refreshes tokens that are about to expire.
func (m *OAuth2TokenManager) refreshExpiring() {
	m.mu.RLock()
	var toRefresh []string
	threshold := time.Now().Add(m.config.TokenRefreshThreshold)
	for veidAddr, token := range m.tokens {
		if token.ExpiresAt.Before(threshold) && token.RefreshToken != "" {
			if !m.refreshing[veidAddr] {
				toRefresh = append(toRefresh, veidAddr)
			}
		}
	}
	m.mu.RUnlock()

	for _, veidAddr := range toRefresh {
		go func(addr string) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_ = m.RefreshToken(ctx, addr)
		}(veidAddr)
	}
}

// GenerateAuthorizationURL generates an authorization URL for OIDC flow.
func (m *OAuth2TokenManager) GenerateAuthorizationURL(state string) (string, string, error) {
	if m.endpoints == nil {
		return "", "", errors.New("OAuth2 manager not initialized")
	}

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", m.config.ClientID)
	params.Set("redirect_uri", m.config.RedirectURL)
	params.Set("scope", strings.Join(m.config.Scopes, " "))
	params.Set("state", state)

	var verifier string
	if m.config.UsePKCE {
		var err error
		verifier, err = generateCodeVerifier()
		if err != nil {
			return "", "", fmt.Errorf("failed to generate PKCE verifier: %w", err)
		}

		challenge := generateCodeChallenge(verifier)
		params.Set("code_challenge", challenge)
		params.Set("code_challenge_method", "S256")
	}

	return m.endpoints.Authorization + "?" + params.Encode(), verifier, nil
}

// ExchangeCode exchanges an authorization code for tokens.
func (m *OAuth2TokenManager) ExchangeCode(ctx context.Context, code, verifier string) (*ManagedToken, error) {
	if m.endpoints == nil {
		return nil, errors.New("OAuth2 manager not initialized")
	}

	// Note: code and verifier are sensitive, never log them
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", m.config.RedirectURL)
	data.Set("client_id", m.config.ClientID)

	if m.config.ClientSecret != "" {
		data.Set("client_secret", m.config.ClientSecret)
	}

	if m.config.UsePKCE && verifier != "" {
		data.Set("code_verifier", verifier)
	}

	token, err := m.doTokenRequest(ctx, data)
	if err != nil {
		return nil, err
	}

	// Store token
	m.mu.Lock()
	m.tokens[token.VEIDAddress] = token
	m.mu.Unlock()

	return token, nil
}

// RefreshToken refreshes the token for a user.
func (m *OAuth2TokenManager) RefreshToken(ctx context.Context, veidAddress string) error {
	m.mu.Lock()
	if m.refreshing[veidAddress] {
		m.mu.Unlock()
		return nil // Already refreshing
	}
	token, exists := m.tokens[veidAddress]
	if !exists || token.RefreshToken == "" {
		m.mu.Unlock()
		return ErrInvalidToken
	}
	m.refreshing[veidAddress] = true
	refreshToken := token.RefreshToken // Store before unlock
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.refreshing, veidAddress)
		m.mu.Unlock()
	}()

	// Note: refreshToken is sensitive, never log it
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", m.config.ClientID)

	if m.config.ClientSecret != "" {
		data.Set("client_secret", m.config.ClientSecret)
	}

	newToken, err := m.doTokenRequest(ctx, data)
	if err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	}

	// Update stored token
	m.mu.Lock()
	oldToken := m.tokens[veidAddress]
	newToken.RefreshCount = oldToken.RefreshCount + 1
	newToken.LastRefresh = time.Now()
	newToken.OnRefresh = oldToken.OnRefresh
	m.tokens[veidAddress] = newToken
	m.mu.Unlock()

	// Call refresh callback
	if newToken.OnRefresh != nil {
		newToken.OnRefresh(newToken)
	}

	return nil
}

// doTokenRequest performs a token endpoint request.
func (m *OAuth2TokenManager) doTokenRequest(ctx context.Context, data url.Values) (*ManagedToken, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.endpoints.Token, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := m.httpClient.Do(req)
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

	token := &ManagedToken{
		VEIDToken: &VEIDToken{
			AccessToken:  tokenResp.AccessToken,
			TokenType:    tokenResp.TokenType,
			ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
			RefreshToken: tokenResp.RefreshToken,
			Scope:        tokenResp.Scope,
		},
		LastRefresh: time.Now(),
	}

	// Parse ID token to get VEID info
	if tokenResp.IDToken != "" {
		if claims, err := m.parseIDToken(tokenResp.IDToken); err == nil {
			token.VEIDAddress = claims.VEIDAddress
			token.IdentityScore = claims.IdentityScore
		}
	}

	// If no VEID from ID token, get from userinfo
	if token.VEIDAddress == "" {
		if userInfo, err := m.getUserInfo(ctx, token.AccessToken); err == nil {
			token.VEIDAddress = userInfo.VEIDAddress
			token.IdentityScore = userInfo.IdentityScore
		}
	}

	return token, nil
}

// idTokenClaims represents ID token claims.
type idTokenClaims struct {
	Subject       string  `json:"sub"`
	VEIDAddress   string  `json:"veid_address"`
	IdentityScore float64 `json:"identity_score"`
	Email         string  `json:"email"`
	Name          string  `json:"name"`
}

// parseIDToken parses claims from ID token (simplified, should verify signature in production).
func (m *OAuth2TokenManager) parseIDToken(idToken string) (*idTokenClaims, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid ID token format")
	}

	// Decode payload (middle part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode ID token payload: %w", err)
	}

	var claims idTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ID token claims: %w", err)
	}

	return &claims, nil
}

// userInfoResponse represents userinfo endpoint response.
type userInfoResponse struct {
	Subject       string  `json:"sub"`
	VEIDAddress   string  `json:"veid_address"`
	IdentityScore float64 `json:"identity_score"`
	Email         string  `json:"email"`
	Name          string  `json:"name"`
}

// getUserInfo fetches user info from the userinfo endpoint.
func (m *OAuth2TokenManager) getUserInfo(ctx context.Context, accessToken string) (*userInfoResponse, error) {
	if m.endpoints == nil || m.endpoints.UserInfo == "" {
		return nil, errors.New("userinfo endpoint not available")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.endpoints.UserInfo, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed: %d", resp.StatusCode)
	}

	var userInfo userInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// GetToken retrieves a managed token for a user.
func (m *OAuth2TokenManager) GetToken(veidAddress string) *ManagedToken {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tokens[veidAddress]
}

// GetValidToken retrieves a valid (non-expired) token, refreshing if needed.
func (m *OAuth2TokenManager) GetValidToken(ctx context.Context, veidAddress string) (*ManagedToken, error) {
	m.mu.RLock()
	token, exists := m.tokens[veidAddress]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrInvalidToken
	}

	// Check if token needs refresh
	if token.ExpiresAt.Before(time.Now().Add(m.config.TokenRefreshThreshold)) {
		if err := m.RefreshToken(ctx, veidAddress); err != nil {
			// If refresh fails but token is still valid, return it
			if token.IsValid() {
				return token, nil
			}
			return nil, err
		}

		// Get refreshed token
		m.mu.RLock()
		token = m.tokens[veidAddress]
		m.mu.RUnlock()
	}

	if !token.IsValid() {
		return nil, ErrInvalidToken
	}

	return token, nil
}

// SetToken stores a token for a user.
func (m *OAuth2TokenManager) SetToken(veidAddress string, token *ManagedToken) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[veidAddress] = token
}

// RemoveToken removes a token for a user.
func (m *OAuth2TokenManager) RemoveToken(veidAddress string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tokens, veidAddress)
}

// RevokeToken revokes a token.
func (m *OAuth2TokenManager) RevokeToken(ctx context.Context, veidAddress string) error {
	m.mu.RLock()
	token, exists := m.tokens[veidAddress]
	m.mu.RUnlock()

	if !exists {
		return nil
	}

	// Remove from storage
	m.mu.Lock()
	delete(m.tokens, veidAddress)
	m.mu.Unlock()

	// Revoke at IdP if endpoint available
	if m.endpoints != nil && m.endpoints.Revocation != "" {
		// Note: tokens are sensitive, never log them
		data := url.Values{}
		data.Set("token", token.AccessToken)
		data.Set("client_id", m.config.ClientID)
		if m.config.ClientSecret != "" {
			data.Set("client_secret", m.config.ClientSecret)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.endpoints.Revocation, strings.NewReader(data.Encode()))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := m.httpClient.Do(req)
		if err != nil {
			return err
		}
		_ = resp.Body.Close()
	}

	return nil
}

// CleanupExpired removes all expired tokens.
func (m *OAuth2TokenManager) CleanupExpired() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for addr, token := range m.tokens {
		if token.IsExpired() && token.RefreshToken == "" {
			delete(m.tokens, addr)
			count++
		}
	}
	return count
}

// generateCodeVerifier generates a PKCE code verifier.
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateCodeChallenge generates a PKCE code challenge from verifier.
func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// OAuth2ProviderAdapter wraps OAuth2TokenManager to implement VEIDAuthProvider.
type OAuth2ProviderAdapter struct {
	manager *OAuth2TokenManager
}

// NewOAuth2ProviderAdapter creates a new OAuth2 provider adapter.
func NewOAuth2ProviderAdapter(manager *OAuth2TokenManager) *OAuth2ProviderAdapter {
	return &OAuth2ProviderAdapter{manager: manager}
}

// ExchangeCodeForToken exchanges an authorization code for tokens.
func (a *OAuth2ProviderAdapter) ExchangeCodeForToken(ctx context.Context, code string, redirectURI string) (*VEIDToken, error) {
	// Note: code is sensitive, never log it
	// For the adapter, we assume PKCE verifier is stored separately
	token, err := a.manager.ExchangeCode(ctx, code, "")
	if err != nil {
		return nil, err
	}
	return token.VEIDToken, nil
}

// RefreshToken refreshes an expired token.
func (a *OAuth2ProviderAdapter) RefreshToken(ctx context.Context, refreshToken string) (*VEIDToken, error) {
	// Note: refreshToken is sensitive, never log it
	// Find the user with this refresh token
	a.manager.mu.RLock()
	var veidAddress string
	for addr, token := range a.manager.tokens {
		if token.RefreshToken == refreshToken {
			veidAddress = addr
			break
		}
	}
	a.manager.mu.RUnlock()

	if veidAddress == "" {
		return nil, ErrInvalidToken
	}

	if err := a.manager.RefreshToken(ctx, veidAddress); err != nil {
		return nil, err
	}

	token := a.manager.GetToken(veidAddress)
	if token == nil {
		return nil, ErrInvalidToken
	}

	return token.VEIDToken, nil
}

// ValidateToken validates a token and returns user info.
func (a *OAuth2ProviderAdapter) ValidateToken(ctx context.Context, accessToken string) (*VEIDToken, error) {
	// Note: accessToken is sensitive, never log it
	userInfo, err := a.manager.getUserInfo(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	return &VEIDToken{
		AccessToken:   accessToken,
		TokenType:     "Bearer",
		VEIDAddress:   userInfo.VEIDAddress,
		IdentityScore: userInfo.IdentityScore,
		ExpiresAt:     time.Now().Add(1 * time.Hour), // Assume 1 hour validity
	}, nil
}

// GetAuthorizationURL gets the authorization URL for OIDC flow.
func (a *OAuth2ProviderAdapter) GetAuthorizationURL(state string, redirectURI string) string {
	url, _, _ := a.manager.GenerateAuthorizationURL(state)
	return url
}

// RevokeToken revokes a token.
func (a *OAuth2ProviderAdapter) RevokeToken(ctx context.Context, token string) error {
	// Note: token is sensitive, never log it
	// Find user with this token
	a.manager.mu.RLock()
	var veidAddress string
	for addr, managedToken := range a.manager.tokens {
		if managedToken.AccessToken == token {
			veidAddress = addr
			break
		}
	}
	a.manager.mu.RUnlock()

	if veidAddress != "" {
		return a.manager.RevokeToken(ctx, veidAddress)
	}
	return nil
}

// Ensure OAuth2ProviderAdapter implements VEIDAuthProvider.
var _ VEIDAuthProvider = (*OAuth2ProviderAdapter)(nil)
