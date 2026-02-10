// Package sso provides the SSO/OIDC verification service for VEID.
//
// This package implements the complete SSO verification flow including:
// - OIDC token verification
// - Attestation signing
// - On-chain linkage creation
// - Audit logging and rate limiting
//
// Task Reference: VE-4B - SSO/OIDC Verification Service
package sso

import (
	"context"
	"time"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// SSO Verification Service Interface
// ============================================================================

// VerificationService defines the interface for SSO/OIDC verification.
type VerificationService interface {
	// InitiateVerification starts an SSO verification flow.
	InitiateVerification(ctx context.Context, req *InitiateRequest) (*InitiateResponse, error)

	// CompleteVerification completes an SSO verification with OIDC token.
	CompleteVerification(ctx context.Context, req *CompleteRequest) (*CompleteResponse, error)

	// ExchangeCodeAndComplete exchanges an authorization code and completes verification.
	ExchangeCodeAndComplete(ctx context.Context, req *CodeExchangeCompleteRequest) (*CompleteResponse, error)

	// GetChallenge retrieves a pending verification challenge.
	GetChallenge(ctx context.Context, challengeID string) (*Challenge, error)

	// RevokeVerification revokes an existing SSO verification.
	RevokeVerification(ctx context.Context, req *RevokeRequest) error

	// GetLinkageStatus returns the status of an SSO linkage.
	GetLinkageStatus(ctx context.Context, accountAddress string) (*LinkageStatus, error)

	// HealthCheck returns the service health status.
	HealthCheck(ctx context.Context) (*HealthStatus, error)

	// Close releases resources.
	Close() error
}

// ============================================================================
// Request/Response Types
// ============================================================================

// InitiateRequest contains parameters for initiating SSO verification.
type InitiateRequest struct {
	// AccountAddress is the blockchain account to link
	AccountAddress string `json:"account_address"`

	// ProviderType is the SSO provider to use
	ProviderType veidtypes.SSOProviderType `json:"provider_type"`

	// OIDCIssuer is the specific OIDC issuer (for generic OIDC)
	OIDCIssuer string `json:"oidc_issuer,omitempty"`

	// SAMLInstitutionID is the EduGAIN institution entity ID (for SAML)
	SAMLInstitutionID string `json:"saml_institution_id,omitempty"`

	// RedirectURI is where to redirect after OIDC flow
	RedirectURI string `json:"redirect_uri"`

	// RequestedScopes specifies additional OAuth scopes to request
	RequestedScopes []string `json:"requested_scopes,omitempty"`

	// LinkageMessage is the message for the user to sign (optional, generated if empty)
	LinkageMessage string `json:"linkage_message,omitempty"`

	// ClientIP is the client's IP address (for rate limiting)
	ClientIP string `json:"client_ip,omitempty"`

	// RequestID is an optional request identifier for correlation
	RequestID string `json:"request_id,omitempty"`
}

// Validate validates the initiate request.
func (r *InitiateRequest) Validate() error {
	if r.AccountAddress == "" {
		return veidtypes.ErrInvalidAddress.Wrap("account_address is required")
	}
	if !veidtypes.IsValidSSOProviderType(r.ProviderType) {
		return veidtypes.ErrInvalidSSO.Wrapf("invalid provider_type: %s", r.ProviderType)
	}
	if r.RedirectURI == "" {
		return veidtypes.ErrInvalidSSO.Wrap("redirect_uri is required")
	}
	if r.ProviderType == veidtypes.SSOProviderOIDC && r.OIDCIssuer == "" {
		return veidtypes.ErrInvalidSSO.Wrap("oidc_issuer required for generic OIDC provider")
	}
	if r.ProviderType == veidtypes.SSOProviderEduGAIN && r.SAMLInstitutionID == "" {
		return veidtypes.ErrInvalidSSO.Wrap("saml_institution_id required for EduGAIN provider")
	}
	return nil
}

// InitiateResponse contains the result of initiating SSO verification.
type InitiateResponse struct {
	// ChallengeID is the unique identifier for this verification attempt
	ChallengeID string `json:"challenge_id"`

	// AuthorizationURL is the URL to redirect the user to
	AuthorizationURL string `json:"authorization_url"`

	// State is the OAuth2 state parameter (for CSRF verification)
	State string `json:"state"`

	// Nonce is the OIDC nonce (for replay protection)
	Nonce string `json:"nonce"`

	// LinkageMessage is the message the user should sign
	LinkageMessage string `json:"linkage_message"`

	// ExpiresAt is when this challenge expires
	ExpiresAt time.Time `json:"expires_at"`

	// SAMLRequest is the encoded SAML AuthnRequest (optional)
	SAMLRequest string `json:"saml_request,omitempty"`

	// RelayState is the SAML relay state (optional)
	RelayState string `json:"relay_state,omitempty"`

	// Binding is the SAML binding used (optional)
	Binding string `json:"binding,omitempty"`

	// PostFormHTML contains the POST form for SAML (optional)
	PostFormHTML string `json:"post_form_html,omitempty"`
}

// CompleteRequest contains parameters for completing SSO verification.
type CompleteRequest struct {
	// ChallengeID is the verification challenge ID
	ChallengeID string `json:"challenge_id"`

	// IDToken is the OIDC ID token received from the provider
	IDToken string `json:"id_token"`

	// SAMLResponse is the base64-encoded SAML response (EduGAIN)
	SAMLResponse string `json:"saml_response,omitempty"`

	// LinkageSignature is the user's signature of the linkage message
	LinkageSignature []byte `json:"linkage_signature"`

	// State is the OAuth2 state parameter (for CSRF verification)
	State string `json:"state"`

	// ClientIP is the client's IP address (for rate limiting)
	ClientIP string `json:"client_ip,omitempty"`
}

// Validate validates the complete request.
func (r *CompleteRequest) Validate() error {
	if r.ChallengeID == "" {
		return veidtypes.ErrInvalidSSO.Wrap("challenge_id is required")
	}
	if r.IDToken == "" && r.SAMLResponse == "" {
		return veidtypes.ErrInvalidSSO.Wrap("id_token or saml_response is required")
	}
	if len(r.LinkageSignature) == 0 {
		return veidtypes.ErrInvalidBindingSignature.Wrap("linkage_signature is required")
	}
	if r.State == "" {
		return veidtypes.ErrInvalidSSO.Wrap("state is required")
	}
	return nil
}

// CodeExchangeCompleteRequest contains parameters for code exchange and completion.
type CodeExchangeCompleteRequest struct {
	// ChallengeID is the verification challenge ID
	ChallengeID string `json:"challenge_id"`

	// AuthorizationCode is the code received from the provider
	AuthorizationCode string `json:"authorization_code"`

	// SAMLResponse is the base64-encoded SAML response (EduGAIN)
	SAMLResponse string `json:"saml_response,omitempty"`

	// LinkageSignature is the user's signature of the linkage message
	LinkageSignature []byte `json:"linkage_signature"`

	// State is the OAuth2 state parameter
	State string `json:"state"`

	// ClientIP is the client's IP address
	ClientIP string `json:"client_ip,omitempty"`
}

// Validate validates the code exchange request.
func (r *CodeExchangeCompleteRequest) Validate() error {
	if r.ChallengeID == "" {
		return veidtypes.ErrInvalidSSO.Wrap("challenge_id is required")
	}
	if r.AuthorizationCode == "" && r.SAMLResponse == "" {
		return veidtypes.ErrInvalidSSO.Wrap("authorization_code or saml_response is required")
	}
	if len(r.LinkageSignature) == 0 {
		return veidtypes.ErrInvalidBindingSignature.Wrap("linkage_signature is required")
	}
	if r.State == "" {
		return veidtypes.ErrInvalidSSO.Wrap("state is required")
	}
	return nil
}

// CompleteResponse contains the result of completing SSO verification.
type CompleteResponse struct {
	// Success indicates if verification was successful
	Success bool `json:"success"`

	// Attestation is the signed SSO attestation (if successful)
	Attestation *veidtypes.SSOAttestation `json:"attestation,omitempty"`

	// LinkageID is the on-chain linkage ID (if created)
	LinkageID string `json:"linkage_id,omitempty"`

	// Error contains error details (if failed)
	Error *VerificationError `json:"error,omitempty"`

	// ProcessedAt is when the response was generated
	ProcessedAt time.Time `json:"processed_at"`
}

// VerificationError contains error details for failed verification.
type VerificationError struct {
	// Code is the error code
	Code string `json:"code"`

	// Message is the error message
	Message string `json:"message"`

	// Details contains additional error context
	Details map[string]string `json:"details,omitempty"`
}

// NewCompleteSuccess creates a successful complete response.
func NewCompleteSuccess(attestation *veidtypes.SSOAttestation, linkageID string) *CompleteResponse {
	return &CompleteResponse{
		Success:     true,
		Attestation: attestation,
		LinkageID:   linkageID,
		ProcessedAt: time.Now(),
	}
}

// NewCompleteError creates an error complete response.
func NewCompleteError(code, message string) *CompleteResponse {
	return &CompleteResponse{
		Success: false,
		Error: &VerificationError{
			Code:    code,
			Message: message,
		},
		ProcessedAt: time.Now(),
	}
}

// RevokeRequest contains parameters for revoking an SSO verification.
type RevokeRequest struct {
	// LinkageID is the linkage to revoke
	LinkageID string `json:"linkage_id"`

	// AccountAddress is the account that owns the linkage
	AccountAddress string `json:"account_address"`

	// Reason is the revocation reason
	Reason string `json:"reason"`

	// Signature is the account's signature authorizing revocation
	Signature []byte `json:"signature"`

	// ClientIP is the client's IP address
	ClientIP string `json:"client_ip,omitempty"`
}

// Validate validates the revoke request.
func (r *RevokeRequest) Validate() error {
	if r.LinkageID == "" {
		return veidtypes.ErrInvalidSSO.Wrap("linkage_id is required")
	}
	if r.AccountAddress == "" {
		return veidtypes.ErrInvalidAddress.Wrap("account_address is required")
	}
	if len(r.Signature) == 0 {
		return veidtypes.ErrInvalidBindingSignature.Wrap("signature is required")
	}
	return nil
}

// LinkageStatus contains the status of an SSO linkage.
type LinkageStatus struct {
	// Exists indicates if a linkage exists
	Exists bool `json:"exists"`

	// LinkageID is the linkage identifier
	LinkageID string `json:"linkage_id,omitempty"`

	// AccountAddress is the linked account
	AccountAddress string `json:"account_address,omitempty"`

	// ProviderType is the SSO provider
	ProviderType veidtypes.SSOProviderType `json:"provider_type,omitempty"`

	// Issuer is the OIDC issuer
	Issuer string `json:"issuer,omitempty"`

	// Status is the current status
	Status veidtypes.SSOVerificationStatus `json:"status,omitempty"`

	// VerifiedAt is when the linkage was verified
	VerifiedAt *time.Time `json:"verified_at,omitempty"`

	// ExpiresAt is when the linkage expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// ScoreContribution is the score contribution from this linkage
	ScoreContribution uint32 `json:"score_contribution,omitempty"`
}

// ============================================================================
// Challenge Types
// ============================================================================

// Challenge represents a pending SSO verification challenge.
type Challenge struct {
	// ChallengeID is the unique identifier
	ChallengeID string `json:"challenge_id"`

	// AccountAddress is the account requesting verification
	AccountAddress string `json:"account_address"`

	// ProviderType is the SSO provider
	ProviderType veidtypes.SSOProviderType `json:"provider_type"`

	// OIDCIssuer is the OIDC issuer
	OIDCIssuer string `json:"oidc_issuer"`

	// SAMLInstitutionID is the EduGAIN institution entity ID (SAML)
	SAMLInstitutionID string `json:"saml_institution_id,omitempty"`

	// State is the OAuth2 state parameter
	State string `json:"state"`

	// Nonce is the OIDC nonce
	Nonce string `json:"nonce"`

	// LinkageMessage is the message for the user to sign
	LinkageMessage string `json:"linkage_message"`

	// RedirectURI is the callback URI
	RedirectURI string `json:"redirect_uri"`

	// Status is the challenge status
	Status ChallengeStatus `json:"status"`

	// CreatedAt is when the challenge was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the challenge expires
	ExpiresAt time.Time `json:"expires_at"`

	// CompletedAt is when the challenge was completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// ClientIP is the initiating client's IP
	ClientIP string `json:"client_ip,omitempty"`
}

// ChallengeStatus represents the status of a challenge.
type ChallengeStatus string

const (
	// ChallengeStatusPending indicates the challenge is pending
	ChallengeStatusPending ChallengeStatus = "pending"

	// ChallengeStatusCompleted indicates the challenge was completed successfully
	ChallengeStatusCompleted ChallengeStatus = "completed"

	// ChallengeStatusFailed indicates the challenge failed
	ChallengeStatusFailed ChallengeStatus = "failed"

	// ChallengeStatusExpired indicates the challenge expired
	ChallengeStatusExpired ChallengeStatus = "expired"

	// ChallengeStatusCancelled indicates the challenge was cancelled
	ChallengeStatusCancelled ChallengeStatus = "cancelled"
)

// IsExpired checks if the challenge has expired.
func (c *Challenge) IsExpired(now time.Time) bool {
	return now.After(c.ExpiresAt)
}

// CanComplete checks if the challenge can be completed.
func (c *Challenge) CanComplete(now time.Time) bool {
	return c.Status == ChallengeStatusPending && !c.IsExpired(now)
}

// ============================================================================
// Health Status
// ============================================================================

// HealthStatus represents the health status of the SSO service.
type HealthStatus struct {
	// Healthy indicates if the service is healthy
	Healthy bool `json:"healthy"`

	// Status is a human-readable status message
	Status string `json:"status"`

	// Timestamp is when the health check was performed
	Timestamp time.Time `json:"timestamp"`

	// Components contains component health statuses
	Components map[string]*ComponentHealth `json:"components,omitempty"`

	// Details contains additional details
	Details map[string]interface{} `json:"details,omitempty"`

	// Warnings contains any warnings
	Warnings []string `json:"warnings,omitempty"`
}

// ComponentHealth represents the health of a component.
type ComponentHealth struct {
	// Name is the component name
	Name string `json:"name"`

	// Healthy indicates if the component is healthy
	Healthy bool `json:"healthy"`

	// Status is a status message
	Status string `json:"status,omitempty"`

	// LastError is the last error (if any)
	LastError string `json:"last_error,omitempty"`

	// Details contains additional details
	Details map[string]interface{} `json:"details,omitempty"`
}

// ============================================================================
// Configuration
// ============================================================================

// Config contains configuration for the SSO verification service.
type Config struct {
	// Enabled determines if SSO verification is active
	Enabled bool `json:"enabled"`

	// ChallengeTTLSeconds is the challenge validity period
	ChallengeTTLSeconds int64 `json:"challenge_ttl_seconds"`

	// AttestationValidityDays is how long attestations are valid
	AttestationValidityDays int `json:"attestation_validity_days"`

	// MaxChallengesPerAccount is the max pending challenges per account
	MaxChallengesPerAccount int `json:"max_challenges_per_account"`

	// RequireLinkageSignature requires user signature for linkage
	RequireLinkageSignature bool `json:"require_linkage_signature"`

	// LinkageMessageTemplate is the template for linkage messages
	LinkageMessageTemplate string `json:"linkage_message_template"`

	// EnableOnChainSubmission enables on-chain linkage submission
	EnableOnChainSubmission bool `json:"enable_on_chain_submission"`

	// AuditEnabled enables audit logging
	AuditEnabled bool `json:"audit_enabled"`

	// MetricsEnabled enables Prometheus metrics
	MetricsEnabled bool `json:"metrics_enabled"`

	// RateLimitEnabled enables rate limiting
	RateLimitEnabled bool `json:"rate_limit_enabled"`
}

// DefaultConfig returns the default SSO service configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:                 true,
		ChallengeTTLSeconds:     600, // 10 minutes
		AttestationValidityDays: 365,
		MaxChallengesPerAccount: 5,
		RequireLinkageSignature: true,
		LinkageMessageTemplate:  "I authorize linking my SSO identity to VirtEngine account %s at %s. Nonce: %s",
		EnableOnChainSubmission: true,
		AuditEnabled:            true,
		MetricsEnabled:          true,
		RateLimitEnabled:        true,
	}
}
