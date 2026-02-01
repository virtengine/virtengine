// Package oidc provides OIDC token verification for the VEID SSO verification service.
package oidc

import (
	"context"
	"crypto"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/virtengine/virtengine/pkg/verification/audit"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Default OIDC Verifier
// ============================================================================

// DefaultVerifier implements OIDCVerifier with JWKS rotation support.
type DefaultVerifier struct {
	config     Config
	jwksManager *JWKSManager
	httpClient *http.Client
	auditor    audit.AuditLogger
	logger     zerolog.Logger

	mu             sync.RWMutex
	discoveryCache map[string]*OIDCDiscoveryDocument // issuer -> discovery doc
}

// OIDCDiscoveryDocument contains the OIDC discovery metadata.
type OIDCDiscoveryDocument struct {
	Issuer                           string   `json:"issuer"`
	AuthorizationEndpoint            string   `json:"authorization_endpoint"`
	TokenEndpoint                    string   `json:"token_endpoint"`
	UserinfoEndpoint                 string   `json:"userinfo_endpoint"`
	JwksURI                          string   `json:"jwks_uri"`
	RegistrationEndpoint             string   `json:"registration_endpoint,omitempty"`
	ScopesSupported                  []string `json:"scopes_supported,omitempty"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	ResponseModesSupported           []string `json:"response_modes_supported,omitempty"`
	GrantTypesSupported              []string `json:"grant_types_supported,omitempty"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
	ClaimsSupported                  []string `json:"claims_supported,omitempty"`
	ACRValuesSupported               []string `json:"acr_values_supported,omitempty"`

	// Cache metadata
	FetchedAt time.Time `json:"-"`
}

// NewDefaultVerifier creates a new DefaultVerifier.
func NewDefaultVerifier(
	ctx context.Context,
	config Config,
	auditor audit.AuditLogger,
	logger zerolog.Logger,
) (*DefaultVerifier, error) {
	jwksConfig := &JWKSManagerConfig{
		DefaultCacheTTL: time.Duration(config.JWKSCacheTTLSeconds) * time.Second,
		RefreshInterval: time.Duration(config.JWKSRefreshIntervalSeconds) * time.Second,
		HTTPTimeout:     time.Duration(config.HTTPClientTimeoutSeconds) * time.Second,
		MaxRetries:      3,
		RetryBackoff:    time.Second,
	}

	v := &DefaultVerifier{
		config:         config,
		jwksManager:    NewJWKSManager(jwksConfig, logger),
		httpClient:     &http.Client{Timeout: time.Duration(config.HTTPClientTimeoutSeconds) * time.Second},
		auditor:        auditor,
		logger:         logger.With().Str("component", "oidc_verifier").Logger(),
		discoveryCache: make(map[string]*OIDCDiscoveryDocument),
	}

	// Pre-fetch discovery documents for configured issuers
	for issuer := range config.IssuerPolicies {
		if _, err := v.getDiscoveryDocument(ctx, issuer); err != nil {
			v.logger.Warn().
				Str("issuer", issuer).
				Err(err).
				Msg("failed to pre-fetch discovery document")
		}
	}

	return v, nil
}

// VerifyToken verifies an OIDC ID token and returns the claims.
func (v *DefaultVerifier) VerifyToken(ctx context.Context, token string, req *VerificationRequest) (*VerifiedClaims, error) {
	// Parse the JWT (without verification first to get header)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("%w: token must have 3 parts", ErrInvalidToken)
	}

	// Decode header
	headerBytes, err := base64URLDecode(parts[0])
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode header: %v", ErrInvalidToken, err)
	}

	var header struct {
		ALG string `json:"alg"`
		KID string `json:"kid"`
		TYP string `json:"typ"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("%w: failed to parse header: %v", ErrInvalidToken, err)
	}

	// Decode payload
	payloadBytes, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode payload: %v", ErrInvalidToken, err)
	}

	var rawClaims map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &rawClaims); err != nil {
		return nil, fmt.Errorf("%w: failed to parse payload: %v", ErrInvalidToken, err)
	}

	// Extract issuer from claims
	issuerClaim, ok := rawClaims["iss"].(string)
	if !ok || issuerClaim == "" {
		return nil, fmt.Errorf("%w: missing or invalid issuer claim", ErrInvalidToken)
	}

	// Check if issuer is allowed
	if !v.IsIssuerAllowed(ctx, issuerClaim) {
		v.logAudit(ctx, audit.EventTypeAccessDenied, issuerClaim, "issuer_not_allowed", nil)
		return nil, fmt.Errorf("%w: issuer: %s", ErrIssuerNotAllowed, issuerClaim)
	}

	// Validate expected issuer if provided
	if req.ExpectedIssuer != "" && issuerClaim != req.ExpectedIssuer {
		return nil, fmt.Errorf("%w: expected %s, got %s", ErrInvalidIssuer, req.ExpectedIssuer, issuerClaim)
	}

	// Get discovery document to find JWKS URI
	discovery, err := v.getDiscoveryDocument(ctx, issuerClaim)
	if err != nil {
		return nil, err
	}

	// Get signing key
	key, err := v.jwksManager.GetKey(ctx, issuerClaim, discovery.JwksURI, header.KID)
	if err != nil {
		return nil, err
	}

	// Verify signature
	if err := v.verifySignature(token, parts, header.ALG, key); err != nil {
		return nil, err
	}

	// Parse claims into structured format
	claims, err := v.parseClaims(rawClaims, issuerClaim)
	if err != nil {
		return nil, err
	}
	claims.KeyID = header.KID

	// Validate claims
	if err := v.validateClaims(claims, req); err != nil {
		return nil, err
	}

	// Apply issuer policy
	policy, _ := v.GetIssuerPolicy(ctx, issuerClaim)
	if policy != nil {
		if err := v.applyPolicy(claims, policy); err != nil {
			return nil, err
		}
	}

	v.logAudit(ctx, audit.EventTypeVerificationCompleted, issuerClaim, "token_verified", map[string]interface{}{
		"subject_hash": veidtypes.HashSubjectID(claims.Subject),
		"email_domain": claims.GetEmailDomain(),
	})

	return claims, nil
}

// verifySignature verifies the JWT signature.
//
//nolint:unparam // token kept for future full token verification
func (v *DefaultVerifier) verifySignature(_ string, parts []string, alg string, key *JSONWebKey) error {
	signedContent := parts[0] + "." + parts[1]
	signatureBytes, err := base64URLDecode(parts[2])
	if err != nil {
		return fmt.Errorf("%w: failed to decode signature: %v", ErrInvalidSignature, err)
	}

	switch alg {
	case "RS256":
		return v.verifyRS256(signedContent, signatureBytes, key)
	case "RS384":
		return v.verifyRS384(signedContent, signatureBytes, key)
	case "RS512":
		return v.verifyRS512(signedContent, signatureBytes, key)
	default:
		return fmt.Errorf("%w: unsupported algorithm: %s", ErrInvalidToken, alg)
	}
}

func (v *DefaultVerifier) verifyRS256(signedContent string, signature []byte, key *JSONWebKey) error {
	return v.verifyRSA(signedContent, signature, key, crypto.SHA256)
}

func (v *DefaultVerifier) verifyRS384(signedContent string, signature []byte, key *JSONWebKey) error {
	return v.verifyRSA(signedContent, signature, key, crypto.SHA384)
}

func (v *DefaultVerifier) verifyRS512(signedContent string, signature []byte, key *JSONWebKey) error {
	return v.verifyRSA(signedContent, signature, key, crypto.SHA512)
}

func (v *DefaultVerifier) verifyRSA(signedContent string, signature []byte, key *JSONWebKey, hash crypto.Hash) error {
	pubKey, ok := key.GetPublicKey().(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("%w: key is not an RSA public key", ErrInvalidSignature)
	}

	hasher := hash.New()
	hasher.Write([]byte(signedContent))
	hashed := hasher.Sum(nil)

	if err := rsa.VerifyPKCS1v15(pubKey, hash, hashed, signature); err != nil {
		return fmt.Errorf("%w: signature verification failed: %v", ErrInvalidSignature, err)
	}

	return nil
}

// parseClaims parses raw claims into VerifiedClaims.
func (v *DefaultVerifier) parseClaims(raw map[string]interface{}, issuer string) (*VerifiedClaims, error) {
	claims := &VerifiedClaims{
		Issuer:       issuer,
		RawClaims:    raw,
		ProviderType: v.detectProviderType(issuer),
	}

	// Required claims
	if sub, ok := raw["sub"].(string); ok {
		claims.Subject = sub
	} else {
		return nil, fmt.Errorf("%w: sub", ErrMissingClaim)
	}

	// Audience (can be string or array)
	switch aud := raw["aud"].(type) {
	case string:
		claims.Audience = []string{aud}
	case []interface{}:
		claims.Audience = make([]string, 0, len(aud))
		for _, a := range aud {
			if s, ok := a.(string); ok {
				claims.Audience = append(claims.Audience, s)
			}
		}
	}

	// Timestamps
	if exp, ok := raw["exp"].(float64); ok {
		claims.ExpiresAt = time.Unix(int64(exp), 0)
	}
	if iat, ok := raw["iat"].(float64); ok {
		claims.IssuedAt = time.Unix(int64(iat), 0)
	}
	if authTime, ok := raw["auth_time"].(float64); ok {
		t := time.Unix(int64(authTime), 0)
		claims.AuthTime = &t
	}

	// Optional claims
	if nonce, ok := raw["nonce"].(string); ok {
		claims.Nonce = nonce
	}
	if acr, ok := raw["acr"].(string); ok {
		claims.ACR = acr
	}
	if amr, ok := raw["amr"].([]interface{}); ok {
		claims.AMR = make([]string, 0, len(amr))
		for _, a := range amr {
			if s, ok := a.(string); ok {
				claims.AMR = append(claims.AMR, s)
			}
		}
	}
	if azp, ok := raw["azp"].(string); ok {
		claims.AZP = azp
	}

	// Profile claims
	if email, ok := raw["email"].(string); ok {
		claims.Email = email
	}
	if emailVerified, ok := raw["email_verified"].(bool); ok {
		claims.EmailVerified = emailVerified
	}
	if name, ok := raw["name"].(string); ok {
		claims.Name = name
	}
	if picture, ok := raw["picture"].(string); ok {
		claims.Picture = picture
	}
	if locale, ok := raw["locale"].(string); ok {
		claims.Locale = locale
	}

	// Provider-specific claims
	if tid, ok := raw["tid"].(string); ok {
		claims.TenantID = tid
	}
	if groups, ok := raw["groups"].([]interface{}); ok {
		claims.Groups = make([]string, 0, len(groups))
		for _, g := range groups {
			if s, ok := g.(string); ok {
				claims.Groups = append(claims.Groups, s)
			}
		}
	}

	return claims, nil
}

// validateClaims validates the parsed claims.
func (v *DefaultVerifier) validateClaims(claims *VerifiedClaims, req *VerificationRequest) error {
	now := time.Now()
	maxSkew := time.Duration(v.config.MaxClockSkewSeconds) * time.Second

	// Check expiration
	if claims.ExpiresAt.Add(maxSkew).Before(now) {
		return fmt.Errorf("%w: expired at %s", ErrTokenExpired, claims.ExpiresAt)
	}

	// Check issued at
	if claims.IssuedAt.Add(-maxSkew).After(now) {
		return fmt.Errorf("%w: token issued in the future", ErrInvalidToken)
	}

	// Check audience
	if req.ExpectedAudience != "" {
		found := false
		for _, aud := range claims.Audience {
			if aud == req.ExpectedAudience {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%w: expected %s, got %v", ErrInvalidAudience, req.ExpectedAudience, claims.Audience)
		}
	}

	// Check nonce
	if req.ExpectedNonce != "" && claims.Nonce != req.ExpectedNonce {
		return fmt.Errorf("%w: expected %s, got %s", ErrInvalidNonce, req.ExpectedNonce, claims.Nonce)
	}

	// Check auth_time (max age)
	if req.MaxAge > 0 && claims.AuthTime != nil {
		maxAge := time.Duration(req.MaxAge) * time.Second
		if claims.AuthTime.Add(maxAge).Before(now) {
			return fmt.Errorf("%w: auth_time: %s, max_age: %s", ErrAuthTooOld, claims.AuthTime, maxAge)
		}
	}

	// Check required claims
	for _, claim := range req.RequiredClaims {
		if _, ok := claims.RawClaims[claim]; !ok {
			return fmt.Errorf("%w: required claim: %s", ErrMissingClaim, claim)
		}
	}

	// Check email verified
	if req.RequireEmailVerified && !claims.EmailVerified {
		return ErrEmailNotVerified
	}

	// Check ACR values
	if len(req.AllowedACRValues) > 0 && claims.ACR != "" {
		found := false
		for _, acr := range req.AllowedACRValues {
			if acr == claims.ACR {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%w: acr: %s, allowed: %v", ErrInvalidACR, claims.ACR, req.AllowedACRValues)
		}
	}

	return nil
}

// applyPolicy applies issuer-specific policy to the claims.
func (v *DefaultVerifier) applyPolicy(claims *VerifiedClaims, policy *IssuerPolicy) error {
	// Check email domain
	if len(policy.AllowedEmailDomains) > 0 {
		domain := claims.GetEmailDomain()
		if !policy.IsEmailDomainAllowed(domain) {
			return fmt.Errorf("%w: domain: %s", ErrEmailDomainNotAllowed, domain)
		}
	}

	// Check tenant
	if len(policy.AllowedTenants) > 0 {
		if !policy.IsTenantAllowed(claims.TenantID) {
			return fmt.Errorf("%w: tenant: %s", ErrTenantNotAllowed, claims.TenantID)
		}
	}

	// Check max auth age from policy
	if policy.MaxAuthAgeSeconds > 0 && claims.AuthTime != nil {
		maxAge := time.Duration(policy.MaxAuthAgeSeconds) * time.Second
		if claims.AuthTime.Add(maxAge).Before(time.Now()) {
			return fmt.Errorf("%w: auth_time: %s, policy max_age: %s", ErrAuthTooOld, claims.AuthTime, maxAge)
		}
	}

	// Check email verified from policy
	if policy.RequireEmailVerified && !claims.EmailVerified {
		return ErrEmailNotVerified
	}

	// Check ACR from policy
	if len(policy.AllowedACRValues) > 0 && claims.ACR != "" {
		found := false
		for _, acr := range policy.AllowedACRValues {
			if acr == claims.ACR {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%w: acr: %s, policy allowed: %v", ErrInvalidACR, claims.ACR, policy.AllowedACRValues)
		}
	}

	return nil
}

// detectProviderType detects the provider type from the issuer.
func (v *DefaultVerifier) detectProviderType(issuer string) veidtypes.SSOProviderType {
	for providerType, knownIssuer := range WellKnownIssuers {
		if issuer == knownIssuer || strings.HasPrefix(issuer, knownIssuer) {
			return providerType
		}
	}

	// Check for Microsoft with specific tenant
	if strings.Contains(issuer, "login.microsoftonline.com") {
		return veidtypes.SSOProviderMicrosoft
	}

	return veidtypes.SSOProviderOIDC
}

// GetAuthorizationURL returns the authorization URL for initiating OIDC flow.
func (v *DefaultVerifier) GetAuthorizationURL(ctx context.Context, req *AuthorizationRequest) (string, error) {
	// Determine issuer
	issuer := req.Issuer
	if issuer == "" {
		if knownIssuer, ok := WellKnownIssuers[req.ProviderType]; ok {
			issuer = knownIssuer
		} else {
			return "", fmt.Errorf("%w: issuer required for generic OIDC", ErrInvalidIssuer)
		}
	}

	// Get discovery document
	discovery, err := v.getDiscoveryDocument(ctx, issuer)
	if err != nil {
		return "", err
	}

	// Build authorization URL
	authURL, err := url.Parse(discovery.AuthorizationEndpoint)
	if err != nil {
		return "", fmt.Errorf("%w: invalid authorization endpoint: %v", ErrConfigurationError, err)
	}

	scopes := req.Scopes
	if len(scopes) == 0 {
		scopes = v.config.DefaultScopes
	}

	params := url.Values{
		"client_id":     {req.ClientID},
		"redirect_uri":  {req.RedirectURI},
		"response_type": {"code"},
		"scope":         {strings.Join(scopes, " ")},
		"state":         {req.State},
		"nonce":         {req.Nonce},
	}

	if req.Prompt != "" {
		params.Set("prompt", req.Prompt)
	}
	if req.MaxAge != nil {
		params.Set("max_age", fmt.Sprintf("%d", *req.MaxAge))
	}
	if req.LoginHint != "" {
		params.Set("login_hint", req.LoginHint)
	}
	if len(req.ACRValues) > 0 {
		params.Set("acr_values", strings.Join(req.ACRValues, " "))
	}

	authURL.RawQuery = params.Encode()
	return authURL.String(), nil
}

// ExchangeCode exchanges an authorization code for tokens.
func (v *DefaultVerifier) ExchangeCode(ctx context.Context, code string, req *CodeExchangeRequest) (*TokenResponse, error) {
	// Determine issuer
	issuer := req.Issuer
	if issuer == "" {
		if knownIssuer, ok := WellKnownIssuers[req.ProviderType]; ok {
			issuer = knownIssuer
		} else {
			return nil, fmt.Errorf("%w: issuer required for generic OIDC", ErrInvalidIssuer)
		}
	}

	// Get discovery document
	discovery, err := v.getDiscoveryDocument(ctx, issuer)
	if err != nil {
		return nil, err
	}

	// Build token request
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {req.RedirectURI},
		"client_id":     {req.ClientID},
		"client_secret": {req.ClientSecret},
	}

	if req.CodeVerifier != "" {
		data.Set("code_verifier", req.CodeVerifier)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, discovery.TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request: %v", ErrCodeExchangeFailed, err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: request failed: %v", ErrCodeExchangeFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: unexpected status: %d", ErrCodeExchangeFailed, resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("%w: failed to decode response: %v", ErrCodeExchangeFailed, err)
	}

	// Verify the ID token
	if tokenResp.IDToken != "" {
		policy, _ := v.GetIssuerPolicy(ctx, issuer)
		audience := req.ClientID
		if policy != nil && len(policy.AllowedAudiences) > 0 {
			audience = policy.AllowedAudiences[0]
		}

		claims, err := v.VerifyToken(ctx, tokenResp.IDToken, &VerificationRequest{
			ExpectedAudience: audience,
			ExpectedNonce:    req.ExpectedNonce,
			ExpectedIssuer:   issuer,
		})
		if err != nil {
			return nil, err
		}
		tokenResp.VerifiedClaims = claims
	}

	return &tokenResp, nil
}

// getDiscoveryDocument fetches and caches the OIDC discovery document.
func (v *DefaultVerifier) getDiscoveryDocument(ctx context.Context, issuer string) (*OIDCDiscoveryDocument, error) {
	v.mu.RLock()
	doc, ok := v.discoveryCache[issuer]
	v.mu.RUnlock()

	// Check if cache is valid (1 hour TTL)
	if ok && time.Since(doc.FetchedAt) < time.Hour {
		return doc, nil
	}

	// Determine discovery URL
	discoveryURL := issuer + "/.well-known/openid-configuration"
	if policy, exists := v.config.IssuerPolicies[issuer]; exists && policy.DiscoveryURL != "" {
		discoveryURL = policy.DiscoveryURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request: %v", ErrDiscoveryFailed, err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: request failed: %v", ErrDiscoveryFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: unexpected status: %d", ErrDiscoveryFailed, resp.StatusCode)
	}

	doc = &OIDCDiscoveryDocument{}
	if err := json.NewDecoder(resp.Body).Decode(doc); err != nil {
		return nil, fmt.Errorf("%w: failed to decode: %v", ErrDiscoveryFailed, err)
	}

	doc.FetchedAt = time.Now()

	// Cache the document
	v.mu.Lock()
	v.discoveryCache[issuer] = doc
	v.mu.Unlock()

	v.logger.Debug().
		Str("issuer", issuer).
		Str("jwks_uri", doc.JwksURI).
		Msg("fetched OIDC discovery document")

	return doc, nil
}

// RefreshJWKS forces a refresh of the JWKS for an issuer.
func (v *DefaultVerifier) RefreshJWKS(ctx context.Context, issuer string) error {
	discovery, err := v.getDiscoveryDocument(ctx, issuer)
	if err != nil {
		return err
	}
	return v.jwksManager.Refresh(ctx, issuer, discovery.JwksURI)
}

// GetIssuerPolicy returns the policy for an issuer.
func (v *DefaultVerifier) GetIssuerPolicy(ctx context.Context, issuer string) (*IssuerPolicy, error) {
	policy, ok := v.config.IssuerPolicies[issuer]
	if !ok {
		return nil, fmt.Errorf("%w: no policy for issuer: %s", ErrIssuerNotAllowed, issuer)
	}
	return policy, nil
}

// IsIssuerAllowed checks if an issuer is in the allowlist.
func (v *DefaultVerifier) IsIssuerAllowed(ctx context.Context, issuer string) bool {
	policy, ok := v.config.IssuerPolicies[issuer]
	return ok && policy.Enabled
}

// HealthCheck returns the health status of the verifier.
func (v *DefaultVerifier) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		Healthy:        true,
		Status:         "healthy",
		Timestamp:      time.Now(),
		IssuerStatuses: make(map[string]*IssuerHealthStatus),
		Details:        make(map[string]interface{}),
		Warnings:       make([]string, 0),
	}

	status.Details["enabled"] = v.config.Enabled
	status.Details["issuer_count"] = len(v.config.IssuerPolicies)

	for issuer, policy := range v.config.IssuerPolicies {
		issuerStatus := v.jwksManager.GetCacheStatus(issuer)
		issuerStatus.Issuer = issuer
		status.IssuerStatuses[issuer] = issuerStatus

		if !policy.Enabled {
			status.Warnings = append(status.Warnings, fmt.Sprintf("issuer %s is disabled", issuer))
		}

		if !issuerStatus.Healthy {
			status.Warnings = append(status.Warnings, fmt.Sprintf("issuer %s is unhealthy", issuer))
		}
	}

	return status, nil
}

// Close releases resources.
func (v *DefaultVerifier) Close() error {
	return v.jwksManager.Close()
}

// logAudit logs an audit event.
func (v *DefaultVerifier) logAudit(ctx context.Context, eventType audit.EventType, resource, action string, details map[string]interface{}) {
	if v.auditor == nil || !v.config.AuditEnabled {
		return
	}

	v.auditor.Log(ctx, audit.Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Actor:     "oidc_verifier",
		Resource:  resource,
		Action:    action,
		Details:   details,
	})
}
