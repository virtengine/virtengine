// Package oidc provides OIDC token verification for the VEID SSO verification service.
//
// This package implements OIDC discovery, JWKS rotation, and token validation
// with support for Google, Microsoft, GitHub, and generic OIDC providers.
//
// Task Reference: VE-4B - SSO/OIDC Verification Service
package oidc

import (
	"context"
	"fmt"
	"time"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// OIDC Verifier Interface
// ============================================================================

// OIDCVerifier defines the interface for OIDC token verification.
type OIDCVerifier interface {
	// VerifyToken verifies an OIDC ID token and returns the claims.
	VerifyToken(ctx context.Context, token string, req *VerificationRequest) (*VerifiedClaims, error)

	// GetAuthorizationURL returns the authorization URL for initiating OIDC flow.
	GetAuthorizationURL(ctx context.Context, req *AuthorizationRequest) (string, error)

	// ExchangeCode exchanges an authorization code for tokens.
	ExchangeCode(ctx context.Context, code string, req *CodeExchangeRequest) (*TokenResponse, error)

	// RefreshJWKS forces a refresh of the JWKS for an issuer.
	RefreshJWKS(ctx context.Context, issuer string) error

	// GetIssuerPolicy returns the policy for an issuer.
	GetIssuerPolicy(ctx context.Context, issuer string) (*IssuerPolicy, error)

	// IsIssuerAllowed checks if an issuer is in the allowlist.
	IsIssuerAllowed(ctx context.Context, issuer string) bool

	// HealthCheck returns the health status of the verifier.
	HealthCheck(ctx context.Context) (*HealthStatus, error)

	// Close releases resources.
	Close() error
}

// ============================================================================
// Verification Request/Response
// ============================================================================

// VerificationRequest contains parameters for token verification.
type VerificationRequest struct {
	// ExpectedAudience is the expected audience claim
	ExpectedAudience string `json:"expected_audience"`

	// ExpectedIssuer is the expected issuer (optional, validated if set)
	ExpectedIssuer string `json:"expected_issuer,omitempty"`

	// ExpectedNonce is the expected nonce claim (for replay protection)
	ExpectedNonce string `json:"expected_nonce"`

	// RequiredClaims are claims that must be present
	RequiredClaims []string `json:"required_claims,omitempty"`

	// MaxAge is the maximum age of the authentication (seconds)
	MaxAge int64 `json:"max_age,omitempty"`

	// AllowedACRValues are the allowed authentication context class values
	AllowedACRValues []string `json:"allowed_acr_values,omitempty"`

	// RequireEmailVerified requires the email_verified claim to be true
	RequireEmailVerified bool `json:"require_email_verified"`

	// ProviderType is the expected provider type
	ProviderType veidtypes.SSOProviderType `json:"provider_type,omitempty"`
}

// VerifiedClaims contains the validated claims from an ID token.
type VerifiedClaims struct {
	// Issuer is the token issuer
	Issuer string `json:"iss"`

	// Subject is the unique subject identifier
	Subject string `json:"sub"`

	// Audience is the intended audience
	Audience []string `json:"aud"`

	// ExpiresAt is when the token expires
	ExpiresAt time.Time `json:"exp"`

	// IssuedAt is when the token was issued
	IssuedAt time.Time `json:"iat"`

	// AuthTime is when the user authenticated (optional)
	AuthTime *time.Time `json:"auth_time,omitempty"`

	// Nonce is the nonce used in the authentication request
	Nonce string `json:"nonce,omitempty"`

	// ACR is the authentication context class reference
	ACR string `json:"acr,omitempty"`

	// AMR is the authentication methods references
	AMR []string `json:"amr,omitempty"`

	// AZP is the authorized party
	AZP string `json:"azp,omitempty"`

	// Email is the user's email address (optional)
	Email string `json:"email,omitempty"`

	// EmailVerified indicates if the email was verified
	EmailVerified bool `json:"email_verified,omitempty"`

	// Name is the user's display name (optional)
	Name string `json:"name,omitempty"`

	// Picture is the user's profile picture URL (optional)
	Picture string `json:"picture,omitempty"`

	// Locale is the user's locale (optional)
	Locale string `json:"locale,omitempty"`

	// TenantID is the organization tenant ID (Microsoft-specific)
	TenantID string `json:"tid,omitempty"`

	// Groups are the user's group memberships (optional)
	Groups []string `json:"groups,omitempty"`

	// RawClaims contains all claims as received
	RawClaims map[string]interface{} `json:"raw_claims,omitempty"`

	// ProviderType is the detected provider type
	ProviderType veidtypes.SSOProviderType `json:"provider_type"`

	// KeyID is the key ID that signed the token
	KeyID string `json:"kid,omitempty"`
}

// GetEmailDomain extracts the domain from the email address.
func (c *VerifiedClaims) GetEmailDomain() string {
	if c.Email == "" {
		return ""
	}
	for i := len(c.Email) - 1; i >= 0; i-- {
		if c.Email[i] == '@' {
			return c.Email[i+1:]
		}
	}
	return ""
}

// ============================================================================
// Authorization Request
// ============================================================================

// AuthorizationRequest contains parameters for generating an authorization URL.
type AuthorizationRequest struct {
	// ProviderType is the SSO provider to use
	ProviderType veidtypes.SSOProviderType `json:"provider_type"`

	// Issuer is the specific OIDC issuer (for generic OIDC)
	Issuer string `json:"issuer,omitempty"`

	// ClientID is the OAuth client ID
	ClientID string `json:"client_id"`

	// RedirectURI is where to redirect after authentication
	RedirectURI string `json:"redirect_uri"`

	// State is the OAuth2 state parameter (CSRF protection)
	State string `json:"state"`

	// Nonce is the OIDC nonce (replay protection)
	Nonce string `json:"nonce"`

	// Scopes are the OAuth scopes to request
	Scopes []string `json:"scopes,omitempty"`

	// Prompt specifies the user interaction mode
	Prompt string `json:"prompt,omitempty"`

	// MaxAge is the maximum authentication age (seconds)
	MaxAge *int64 `json:"max_age,omitempty"`

	// LoginHint is a hint about the user's identity
	LoginHint string `json:"login_hint,omitempty"`

	// ACRValues are the requested authentication context class values
	ACRValues []string `json:"acr_values,omitempty"`

	// TenantHint is the tenant hint (for multi-tenant scenarios)
	TenantHint string `json:"tenant_hint,omitempty"`
}

// ============================================================================
// Code Exchange
// ============================================================================

// CodeExchangeRequest contains parameters for exchanging an authorization code.
type CodeExchangeRequest struct {
	// ProviderType is the SSO provider
	ProviderType veidtypes.SSOProviderType `json:"provider_type"`

	// Issuer is the specific OIDC issuer (for generic OIDC)
	Issuer string `json:"issuer,omitempty"`

	// ClientID is the OAuth client ID
	ClientID string `json:"client_id"`

	// ClientSecret is the OAuth client secret
	ClientSecret string `json:"client_secret"`

	// RedirectURI is the redirect URI used in the authorization request
	RedirectURI string `json:"redirect_uri"`

	// Code is the authorization code
	Code string `json:"code"`

	// CodeVerifier is the PKCE code verifier (if PKCE was used)
	CodeVerifier string `json:"code_verifier,omitempty"`

	// ExpectedNonce is the expected nonce in the ID token
	ExpectedNonce string `json:"expected_nonce"`
}

// TokenResponse contains the response from a token exchange.
type TokenResponse struct {
	// AccessToken is the OAuth access token
	AccessToken string `json:"access_token"`

	// TokenType is the token type (usually "Bearer")
	TokenType string `json:"token_type"`

	// ExpiresIn is the access token lifetime in seconds
	ExpiresIn int64 `json:"expires_in"`

	// RefreshToken is the refresh token (if issued)
	RefreshToken string `json:"refresh_token,omitempty"`

	// Scope is the granted scopes
	Scope string `json:"scope,omitempty"`

	// IDToken is the OIDC ID token
	IDToken string `json:"id_token"`

	// VerifiedClaims are the validated claims from the ID token
	VerifiedClaims *VerifiedClaims `json:"verified_claims,omitempty"`
}

// ============================================================================
// Issuer Policy
// ============================================================================

// IssuerPolicy defines the policy for an OIDC issuer.
type IssuerPolicy struct {
	// Issuer is the OIDC issuer URL
	Issuer string `json:"issuer"`

	// ProviderType is the provider category
	ProviderType veidtypes.SSOProviderType `json:"provider_type"`

	// Enabled indicates if this issuer is enabled
	Enabled bool `json:"enabled"`

	// ClientID is the OAuth client ID for this issuer
	ClientID string `json:"client_id"`

	// ClientSecretRef is a reference to the client secret (e.g., vault path)
	ClientSecretRef string `json:"client_secret_ref,omitempty"`

	// AllowedAudiences are the allowed audience values
	AllowedAudiences []string `json:"allowed_audiences,omitempty"`

	// RequiredScopes are scopes required for verification
	RequiredScopes []string `json:"required_scopes,omitempty"`

	// RequiredClaims are claims that must be present
	RequiredClaims []string `json:"required_claims,omitempty"`

	// RequireEmailVerified requires email_verified to be true
	RequireEmailVerified bool `json:"require_email_verified"`

	// AllowedEmailDomains restricts to specific email domains (empty = all)
	AllowedEmailDomains []string `json:"allowed_email_domains,omitempty"`

	// AllowedTenants restricts to specific tenant IDs (empty = all)
	AllowedTenants []string `json:"allowed_tenants,omitempty"`

	// MaxAuthAgeSeconds is the maximum authentication age
	MaxAuthAgeSeconds int64 `json:"max_auth_age_seconds,omitempty"`

	// AllowedACRValues are the allowed authentication context class values
	AllowedACRValues []string `json:"allowed_acr_values,omitempty"`

	// ScoreWeight is the score weight for this issuer (basis points, max 10000)
	ScoreWeight uint32 `json:"score_weight"`

	// RateLimitPerHour is the rate limit for this issuer
	RateLimitPerHour int `json:"rate_limit_per_hour"`

	// JWKSCacheTTLSeconds is the JWKS cache TTL
	JWKSCacheTTLSeconds int64 `json:"jwks_cache_ttl_seconds"`

	// DiscoveryURL overrides the default discovery URL
	DiscoveryURL string `json:"discovery_url,omitempty"`

	// Notes contains administrative notes
	Notes string `json:"notes,omitempty"`

	// CreatedAt is when the policy was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the policy was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates the issuer policy.
func (p *IssuerPolicy) Validate() error {
	if p.Issuer == "" {
		return fmt.Errorf("%w: issuer is required", ErrInvalidIssuerPolicy)
	}
	if !veidtypes.IsValidSSOProviderType(p.ProviderType) {
		return fmt.Errorf("%w: invalid provider_type: %s", ErrInvalidIssuerPolicy, p.ProviderType)
	}
	if p.ClientID == "" {
		return fmt.Errorf("%w: client_id is required", ErrInvalidIssuerPolicy)
	}
	if p.ScoreWeight > 10000 {
		return fmt.Errorf("%w: score_weight cannot exceed 10000", ErrInvalidIssuerPolicy)
	}
	return nil
}

// IsEmailDomainAllowed checks if an email domain is allowed by this policy.
func (p *IssuerPolicy) IsEmailDomainAllowed(domain string) bool {
	if len(p.AllowedEmailDomains) == 0 {
		return true
	}
	for _, allowed := range p.AllowedEmailDomains {
		if allowed == domain {
			return true
		}
	}
	return false
}

// IsTenantAllowed checks if a tenant is allowed by this policy.
func (p *IssuerPolicy) IsTenantAllowed(tenantID string) bool {
	if len(p.AllowedTenants) == 0 {
		return true
	}
	for _, allowed := range p.AllowedTenants {
		if allowed == tenantID {
			return true
		}
	}
	return false
}

// ============================================================================
// Configuration
// ============================================================================

// Config contains configuration for the OIDC verifier.
type Config struct {
	// Enabled determines if OIDC verification is active
	Enabled bool `json:"enabled"`

	// IssuerPolicies contains policies for each allowed issuer
	IssuerPolicies map[string]*IssuerPolicy `json:"issuer_policies"`

	// DefaultScopes are default OAuth scopes to request
	DefaultScopes []string `json:"default_scopes"`

	// JWKSCacheTTLSeconds is the default JWKS cache TTL
	JWKSCacheTTLSeconds int64 `json:"jwks_cache_ttl_seconds"`

	// JWKSRefreshIntervalSeconds is how often to proactively refresh JWKS
	JWKSRefreshIntervalSeconds int64 `json:"jwks_refresh_interval_seconds"`

	// MaxClockSkewSeconds is the maximum allowed clock skew
	MaxClockSkewSeconds int64 `json:"max_clock_skew_seconds"`

	// TokenValidityWindowSeconds is the token validity window
	TokenValidityWindowSeconds int64 `json:"token_validity_window_seconds"`

	// HTTPClientTimeoutSeconds is the HTTP client timeout
	HTTPClientTimeoutSeconds int `json:"http_client_timeout_seconds"`

	// AllowInsecureHTTP allows insecure HTTP for development
	AllowInsecureHTTP bool `json:"allow_insecure_http"`

	// MetricsEnabled enables Prometheus metrics
	MetricsEnabled bool `json:"metrics_enabled"`

	// AuditEnabled enables audit logging
	AuditEnabled bool `json:"audit_enabled"`
}

// DefaultConfig returns the default OIDC verifier configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:                    true,
		IssuerPolicies:             make(map[string]*IssuerPolicy),
		DefaultScopes:              []string{"openid", "email", "profile"},
		JWKSCacheTTLSeconds:        3600, // 1 hour
		JWKSRefreshIntervalSeconds: 1800, // 30 minutes
		MaxClockSkewSeconds:        300,  // 5 minutes
		TokenValidityWindowSeconds: 3600, // 1 hour
		HTTPClientTimeoutSeconds:   30,
		AllowInsecureHTTP:          false,
		MetricsEnabled:             true,
		AuditEnabled:               true,
	}
}

// ============================================================================
// Well-Known Provider Configurations
// ============================================================================

// WellKnownIssuers contains issuer URLs for well-known providers.
var WellKnownIssuers = map[veidtypes.SSOProviderType]string{
	veidtypes.SSOProviderGoogle:    "https://accounts.google.com",
	veidtypes.SSOProviderMicrosoft: "https://login.microsoftonline.com/common/v2.0",
	veidtypes.SSOProviderGitHub:    "https://token.actions.githubusercontent.com", // GitHub Actions OIDC
}

// WellKnownDiscoveryURLs contains discovery URLs for well-known providers.
var WellKnownDiscoveryURLs = map[veidtypes.SSOProviderType]string{
	veidtypes.SSOProviderGoogle:    "https://accounts.google.com/.well-known/openid-configuration",
	veidtypes.SSOProviderMicrosoft: "https://login.microsoftonline.com/common/v2.0/.well-known/openid-configuration",
}

// DefaultIssuerPolicy returns the default policy for a provider type.
func DefaultIssuerPolicy(providerType veidtypes.SSOProviderType) *IssuerPolicy {
	issuer, ok := WellKnownIssuers[providerType]
	if !ok {
		issuer = ""
	}

	now := time.Now()
	return &IssuerPolicy{
		Issuer:               issuer,
		ProviderType:         providerType,
		Enabled:              true,
		RequiredScopes:       []string{"openid", "email"},
		RequiredClaims:       []string{"sub", "email"},
		RequireEmailVerified: true,
		MaxAuthAgeSeconds:    86400, // 24 hours
		ScoreWeight:          veidtypes.GetSSOScoringWeight(providerType),
		RateLimitPerHour:     100,
		JWKSCacheTTLSeconds:  3600,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// ============================================================================
// Health Status
// ============================================================================

// HealthStatus represents the health status of the OIDC verifier.
type HealthStatus struct {
	// Healthy indicates if the verifier is healthy
	Healthy bool `json:"healthy"`

	// Status is a human-readable status message
	Status string `json:"status"`

	// Timestamp is when the health check was performed
	Timestamp time.Time `json:"timestamp"`

	// IssuerStatuses contains per-issuer status
	IssuerStatuses map[string]*IssuerHealthStatus `json:"issuer_statuses,omitempty"`

	// Details contains additional details
	Details map[string]interface{} `json:"details,omitempty"`

	// Warnings contains any warnings
	Warnings []string `json:"warnings,omitempty"`
}

// IssuerHealthStatus represents the health status for a specific issuer.
type IssuerHealthStatus struct {
	// Issuer is the issuer URL
	Issuer string `json:"issuer"`

	// Healthy indicates if the issuer is healthy
	Healthy bool `json:"healthy"`

	// JWKSCached indicates if JWKS is cached
	JWKSCached bool `json:"jwks_cached"`

	// JWKSLastRefresh is when JWKS was last refreshed
	JWKSLastRefresh *time.Time `json:"jwks_last_refresh,omitempty"`

	// JWKSExpiresAt is when the JWKS cache expires
	JWKSExpiresAt *time.Time `json:"jwks_expires_at,omitempty"`

	// KeyCount is the number of keys in the JWKS
	KeyCount int `json:"key_count"`

	// LastError contains the last error (if any)
	LastError string `json:"last_error,omitempty"`

	// LastErrorAt is when the last error occurred
	LastErrorAt *time.Time `json:"last_error_at,omitempty"`
}
