// Package email provides link-based verification handlers.
//
// This file implements magic link verification endpoints with:
// - Secure token validation
// - Expiration checks
// - Anti-replay protection
// - Redirect handling
//
// Task Reference: VE-3F - Email Verification Delivery + Attestation
package email

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/virtengine/virtengine/pkg/errors"
	"github.com/virtengine/virtengine/pkg/verification/audit"
)

// ============================================================================
// Link Verification Configuration
// ============================================================================

// LinkVerificationConfig contains configuration for link-based verification.
type LinkVerificationConfig struct {
	// BaseURL is the base URL for verification links
	BaseURL string `json:"base_url"`

	// VerifyPath is the path for verification endpoint
	VerifyPath string `json:"verify_path"`

	// SuccessRedirectURL is where to redirect on successful verification
	SuccessRedirectURL string `json:"success_redirect_url"`

	// FailureRedirectURL is where to redirect on failed verification
	FailureRedirectURL string `json:"failure_redirect_url"`

	// TokenTTLSeconds is the token time-to-live in seconds
	TokenTTLSeconds int64 `json:"token_ttl_seconds"`

	// MaxVerificationAttempts is the maximum verification attempts
	MaxVerificationAttempts uint32 `json:"max_verification_attempts"`

	// EnableCSRFProtection enables CSRF protection for web handlers
	EnableCSRFProtection bool `json:"enable_csrf_protection"`

	// AllowedOrigins for CORS
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
}

// DefaultLinkVerificationConfig returns the default configuration.
func DefaultLinkVerificationConfig() LinkVerificationConfig {
	return LinkVerificationConfig{
		VerifyPath:              "/verify",
		TokenTTLSeconds:         DefaultLinkTTLSeconds,
		MaxVerificationAttempts: DefaultMaxOTPAttempts,
		EnableCSRFProtection:    true,
	}
}

// Validate validates the configuration.
func (c *LinkVerificationConfig) Validate() error {
	if c.BaseURL == "" {
		return errors.Wrap(ErrInvalidConfig, "base_url is required")
	}
	if c.VerifyPath == "" {
		return errors.Wrap(ErrInvalidConfig, "verify_path is required")
	}
	if c.TokenTTLSeconds <= 0 {
		c.TokenTTLSeconds = DefaultLinkTTLSeconds
	}
	return nil
}

// ============================================================================
// Link Verification Service
// ============================================================================

// LinkVerificationService handles magic link verification.
type LinkVerificationService struct {
	config  LinkVerificationConfig
	service EmailVerificationService
	auditor audit.AuditLogger
	logger  zerolog.Logger
}

// NewLinkVerificationService creates a new link verification service.
func NewLinkVerificationService(
	config LinkVerificationConfig,
	service EmailVerificationService,
	auditor audit.AuditLogger,
	logger zerolog.Logger,
) (*LinkVerificationService, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &LinkVerificationService{
		config:  config,
		service: service,
		auditor: auditor,
		logger:  logger.With().Str("component", "link_verification").Logger(),
	}, nil
}

// ============================================================================
// Link Generation
// ============================================================================

// LinkGenerationRequest contains parameters for generating a verification link.
type LinkGenerationRequest struct {
	// ChallengeID is the challenge this link is for
	ChallengeID string `json:"challenge_id"`

	// Token is the verification token
	Token string `json:"token"`

	// AccountAddress is the account to verify
	AccountAddress string `json:"account_address"`

	// ExpiresAt is when the link expires
	ExpiresAt time.Time `json:"expires_at"`

	// Metadata contains additional data to encode in the link
	Metadata map[string]string `json:"metadata,omitempty"`
}

// LinkGenerationResult contains the generated verification link.
type LinkGenerationResult struct {
	// Link is the full verification URL
	Link string `json:"link"`

	// ShortLink is a shortened version (if available)
	ShortLink string `json:"short_link,omitempty"`

	// ExpiresAt is when the link expires
	ExpiresAt time.Time `json:"expires_at"`

	// QRCode is a base64-encoded QR code (if requested)
	QRCode string `json:"qr_code,omitempty"`
}

// GenerateLink creates a verification link.
func (s *LinkVerificationService) GenerateLink(req *LinkGenerationRequest) (*LinkGenerationResult, error) {
	if req.ChallengeID == "" {
		return nil, errors.Wrap(ErrInvalidRequest, "challenge_id is required")
	}
	if req.Token == "" {
		return nil, errors.Wrap(ErrInvalidRequest, "token is required")
	}

	// Build verification URL
	verifyURL, err := url.Parse(s.config.BaseURL)
	if err != nil {
		return nil, errors.Wrap(ErrInvalidConfig, "invalid base_url")
	}

	verifyURL.Path = s.config.VerifyPath

	// Add query parameters
	query := verifyURL.Query()
	query.Set("token", req.Token)
	query.Set("challenge", req.ChallengeID)
	verifyURL.RawQuery = query.Encode()

	return &LinkGenerationResult{
		Link:      verifyURL.String(),
		ExpiresAt: req.ExpiresAt,
	}, nil
}

// ============================================================================
// Link Verification Handlers
// ============================================================================

// LinkVerificationRequest represents a verification request from a magic link.
type LinkVerificationRequest struct {
	// Token is the verification token from the URL
	Token string `json:"token"`

	// ChallengeID is the challenge ID from the URL
	ChallengeID string `json:"challenge_id"`

	// AccountAddress is the account to verify (optional, for extra validation)
	AccountAddress string `json:"account_address,omitempty"`

	// IPAddress is the client IP
	IPAddress string `json:"ip_address,omitempty"`

	// UserAgent is the client user agent
	UserAgent string `json:"user_agent,omitempty"`
}

// Validate validates the link verification request.
func (r *LinkVerificationRequest) Validate() error {
	if r.Token == "" {
		return errors.Wrap(ErrInvalidRequest, "token is required")
	}
	if r.ChallengeID == "" {
		return errors.Wrap(ErrInvalidRequest, "challenge_id is required")
	}
	return nil
}

// LinkVerificationResponse represents the result of link verification.
type LinkVerificationResponse struct {
	// Success indicates if verification succeeded
	Success bool `json:"success"`

	// Verified indicates if the email is now verified
	Verified bool `json:"verified"`

	// AttestationID is the ID of the created attestation
	AttestationID string `json:"attestation_id,omitempty"`

	// AccountAddress is the verified account
	AccountAddress string `json:"account_address,omitempty"`

	// RedirectURL is where to redirect the user
	RedirectURL string `json:"redirect_url,omitempty"`

	// ErrorCode is the error code if verification failed
	ErrorCode string `json:"error_code,omitempty"`

	// ErrorMessage is the error message if verification failed
	ErrorMessage string `json:"error_message,omitempty"`
}

// VerifyLink processes a verification link.
func (s *LinkVerificationService) VerifyLink(ctx context.Context, req *LinkVerificationRequest) (*LinkVerificationResponse, error) {
	if err := req.Validate(); err != nil {
		return &LinkVerificationResponse{
			Success:      false,
			ErrorCode:    "INVALID_REQUEST",
			ErrorMessage: err.Error(),
			RedirectURL:  s.config.FailureRedirectURL,
		}, err
	}

	// Get the challenge to find the account address
	challenge, err := s.service.GetChallenge(ctx, req.ChallengeID)
	if err != nil {
		s.logger.Warn().
			Str("challenge_id", req.ChallengeID).
			Err(err).
			Msg("challenge not found for link verification")

		return &LinkVerificationResponse{
			Success:      false,
			ErrorCode:    "CHALLENGE_NOT_FOUND",
			ErrorMessage: "Verification link is invalid or expired",
			RedirectURL:  s.config.FailureRedirectURL,
		}, ErrChallengeNotFound
	}

	// Verify using the main service
	verifyReq := &VerifyRequest{
		ChallengeID:    req.ChallengeID,
		Secret:         req.Token,
		AccountAddress: challenge.AccountAddress,
		RequestID:      fmt.Sprintf("link-%s", req.ChallengeID),
		IPAddress:      req.IPAddress,
	}

	verifyResp, err := s.service.VerifyChallenge(ctx, verifyReq)
	if err != nil {
		s.logger.Warn().
			Str("challenge_id", req.ChallengeID).
			Err(err).
			Msg("link verification failed")

		return &LinkVerificationResponse{
			Success:      false,
			ErrorCode:    verifyResp.ErrorCode,
			ErrorMessage: verifyResp.ErrorMessage,
			RedirectURL:  s.buildFailureRedirect(verifyResp.ErrorCode),
		}, err
	}

	// Audit log
	if s.auditor != nil {
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeVerificationCompleted,
			Timestamp: time.Now(),
			Actor:     challenge.AccountAddress,
			Resource:  req.ChallengeID,
			Action:    "link_verification",
			Details: map[string]interface{}{
				"ip_address":     req.IPAddress,
				"user_agent":     req.UserAgent,
				"attestation_id": verifyResp.AttestationID,
			},
		})
	}

	s.logger.Info().
		Str("challenge_id", req.ChallengeID).
		Str("account", challenge.AccountAddress).
		Msg("link verification successful")

	return &LinkVerificationResponse{
		Success:        true,
		Verified:       true,
		AttestationID:  verifyResp.AttestationID,
		AccountAddress: challenge.AccountAddress,
		RedirectURL:    s.buildSuccessRedirect(verifyResp.AttestationID),
	}, nil
}

// buildSuccessRedirect builds the success redirect URL.
func (s *LinkVerificationService) buildSuccessRedirect(attestationID string) string {
	if s.config.SuccessRedirectURL == "" {
		return ""
	}

	redirectURL, err := url.Parse(s.config.SuccessRedirectURL)
	if err != nil {
		return s.config.SuccessRedirectURL
	}

	query := redirectURL.Query()
	query.Set("verified", "true")
	if attestationID != "" {
		query.Set("attestation_id", attestationID)
	}
	redirectURL.RawQuery = query.Encode()

	return redirectURL.String()
}

// buildFailureRedirect builds the failure redirect URL.
func (s *LinkVerificationService) buildFailureRedirect(errorCode string) string {
	if s.config.FailureRedirectURL == "" {
		return ""
	}

	redirectURL, err := url.Parse(s.config.FailureRedirectURL)
	if err != nil {
		return s.config.FailureRedirectURL
	}

	query := redirectURL.Query()
	query.Set("verified", "false")
	if errorCode != "" {
		query.Set("error", errorCode)
	}
	redirectURL.RawQuery = query.Encode()

	return redirectURL.String()
}

// ============================================================================
// HTTP Handlers
// ============================================================================

// HTTPHandler returns an HTTP handler for link verification.
func (s *LinkVerificationService) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Parse request
		token := r.URL.Query().Get("token")
		challengeID := r.URL.Query().Get("challenge")

		req := &LinkVerificationRequest{
			Token:       token,
			ChallengeID: challengeID,
			IPAddress:   getClientIP(r),
			UserAgent:   r.UserAgent(),
		}

		// Verify
		resp, err := s.VerifyLink(ctx, req)

		// Handle redirect
		if resp.RedirectURL != "" {
			http.Redirect(w, r, resp.RedirectURL, http.StatusFound)
			return
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"success":false,"error":"%s","message":"%s"}`,
				resp.ErrorCode, resp.ErrorMessage)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"success":true,"verified":true,"attestation_id":"%s"}`,
			resp.AttestationID)
	})
}

// getClientIP extracts the client IP from the request.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the chain
		if idx := strings.Index(forwarded, ","); idx != -1 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return strings.TrimSpace(forwarded)
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

// ============================================================================
// CSRF Protection
// ============================================================================

// CSRFToken generates a CSRF token for form submissions.
func (s *LinkVerificationService) CSRFToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", errors.Wrap(ErrChallengeCreation, "failed to generate CSRF token")
	}
	return base64.URLEncoding.EncodeToString(token), nil
}

// ValidateCSRFToken validates a CSRF token.
func (s *LinkVerificationService) ValidateCSRFToken(token, expected string) bool {
	if !s.config.EnableCSRFProtection {
		return true
	}
	return token != "" && token == expected
}

// ============================================================================
// One-Click Verification
// ============================================================================

// OneClickVerificationRequest is for one-click email verification (no redirect).
type OneClickVerificationRequest struct {
	// Token is the verification token
	Token string `json:"token"`

	// ChallengeID is the challenge ID
	ChallengeID string `json:"challenge_id"`

	// Signature is an optional signature for enhanced security
	Signature string `json:"signature,omitempty"`
}

// OneClickVerify performs one-click verification (typically from email client).
// This is useful for email clients that support List-Unsubscribe-Post style headers.
func (s *LinkVerificationService) OneClickVerify(ctx context.Context, req *OneClickVerificationRequest) error {
	linkReq := &LinkVerificationRequest{
		Token:       req.Token,
		ChallengeID: req.ChallengeID,
	}

	_, err := s.VerifyLink(ctx, linkReq)
	return err
}
