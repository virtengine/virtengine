// Package sdk provides shared interfaces for verification provider integrations.
//
// This package defines common interfaces and types for email, SMS, and SSO
// verification service integrations.
//
// Task Reference: VE-2B - Verification Shared Infrastructure
package sdk

import (
	"context"
	"time"
)

// ============================================================================
// Common Types
// ============================================================================

// ProviderType identifies the type of verification provider.
type ProviderType string

const (
	ProviderTypeEmail ProviderType = "email"
	ProviderTypeSMS   ProviderType = "sms"
	ProviderTypeSSO   ProviderType = "sso"
	ProviderTypeMFA   ProviderType = "mfa"
)

// VerificationStatus represents the status of a verification.
type VerificationStatus string

const (
	StatusPending   VerificationStatus = "pending"
	StatusVerified  VerificationStatus = "verified"
	StatusFailed    VerificationStatus = "failed"
	StatusExpired   VerificationStatus = "expired"
	StatusCancelled VerificationStatus = "cancelled"
)

// ============================================================================
// Base Provider Interface
// ============================================================================

// Provider is the base interface for all verification providers.
type Provider interface {
	// Name returns the provider name.
	Name() string

	// Type returns the provider type.
	Type() ProviderType

	// HealthCheck verifies the provider is accessible.
	HealthCheck(ctx context.Context) error

	// Close closes the provider.
	Close() error
}

// ============================================================================
// Email Provider
// ============================================================================

// EmailProvider defines the interface for email verification providers.
type EmailProvider interface {
	Provider

	// SendVerificationEmail sends a verification email.
	SendVerificationEmail(ctx context.Context, req SendEmailRequest) (*SendEmailResponse, error)

	// VerifyEmailCode verifies an email verification code.
	VerifyEmailCode(ctx context.Context, req VerifyEmailRequest) (*VerifyEmailResponse, error)

	// GetVerificationStatus gets the status of an email verification.
	GetVerificationStatus(ctx context.Context, verificationID string) (*EmailVerificationStatus, error)

	// ResendVerificationEmail resends a verification email.
	ResendVerificationEmail(ctx context.Context, verificationID string) error

	// CancelVerification cancels a pending verification.
	CancelVerification(ctx context.Context, verificationID string) error
}

// SendEmailRequest contains parameters for sending a verification email.
type SendEmailRequest struct {
	// Email is the email address to verify
	Email string `json:"email"`

	// AccountAddress is the associated blockchain account
	AccountAddress string `json:"account_address"`

	// Subject is the email subject (optional, uses default if empty)
	Subject string `json:"subject,omitempty"`

	// Template is the email template to use (optional)
	Template string `json:"template,omitempty"`

	// ExpirationMinutes is how long the code is valid (optional)
	ExpirationMinutes int `json:"expiration_minutes,omitempty"`

	// Locale is the user's locale for localization
	Locale string `json:"locale,omitempty"`

	// Metadata contains additional context
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SendEmailResponse contains the result of sending a verification email.
type SendEmailResponse struct {
	// VerificationID is the unique identifier for this verification
	VerificationID string `json:"verification_id"`

	// SentAt is when the email was sent
	SentAt time.Time `json:"sent_at"`

	// ExpiresAt is when the verification expires
	ExpiresAt time.Time `json:"expires_at"`

	// MaskedEmail is the masked email address (e.g., j***@example.com)
	MaskedEmail string `json:"masked_email"`
}

// VerifyEmailRequest contains parameters for verifying an email code.
type VerifyEmailRequest struct {
	// VerificationID is the verification identifier
	VerificationID string `json:"verification_id"`

	// Code is the verification code entered by the user
	Code string `json:"code"`

	// AccountAddress is the expected account address
	AccountAddress string `json:"account_address"`
}

// VerifyEmailResponse contains the result of email verification.
type VerifyEmailResponse struct {
	// Verified indicates if the email was verified
	Verified bool `json:"verified"`

	// Email is the verified email address
	Email string `json:"email"`

	// VerifiedAt is when the email was verified
	VerifiedAt time.Time `json:"verified_at"`

	// ErrorCode is the error code if verification failed
	ErrorCode string `json:"error_code,omitempty"`

	// ErrorMessage is the error message if verification failed
	ErrorMessage string `json:"error_message,omitempty"`

	// AttemptsRemaining is the number of attempts remaining
	AttemptsRemaining int `json:"attempts_remaining,omitempty"`
}

// EmailVerificationStatus contains the current status of an email verification.
type EmailVerificationStatus struct {
	// VerificationID is the verification identifier
	VerificationID string `json:"verification_id"`

	// Status is the current status
	Status VerificationStatus `json:"status"`

	// Email is the email being verified
	Email string `json:"email"`

	// CreatedAt is when the verification was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the verification expires
	ExpiresAt time.Time `json:"expires_at"`

	// VerifiedAt is when it was verified (if verified)
	VerifiedAt *time.Time `json:"verified_at,omitempty"`

	// AttemptCount is the number of verification attempts
	AttemptCount int `json:"attempt_count"`

	// MaxAttempts is the maximum allowed attempts
	MaxAttempts int `json:"max_attempts"`
}

// ============================================================================
// SMS Provider
// ============================================================================

// SMSProvider defines the interface for SMS verification providers.
type SMSProvider interface {
	Provider

	// SendVerificationSMS sends a verification SMS.
	SendVerificationSMS(ctx context.Context, req SendSMSRequest) (*SendSMSResponse, error)

	// VerifySMSCode verifies an SMS verification code.
	VerifySMSCode(ctx context.Context, req VerifySMSRequest) (*VerifySMSResponse, error)

	// GetVerificationStatus gets the status of an SMS verification.
	GetVerificationStatus(ctx context.Context, verificationID string) (*SMSVerificationStatus, error)

	// CheckPhoneType checks if a phone number is VOIP, mobile, or landline.
	CheckPhoneType(ctx context.Context, phoneNumber string) (*PhoneTypeResult, error)

	// FormatPhoneNumber formats a phone number to E.164 format.
	FormatPhoneNumber(ctx context.Context, phoneNumber string, countryCode string) (string, error)
}

// SendSMSRequest contains parameters for sending a verification SMS.
type SendSMSRequest struct {
	// PhoneNumber is the phone number to verify (E.164 format)
	PhoneNumber string `json:"phone_number"`

	// AccountAddress is the associated blockchain account
	AccountAddress string `json:"account_address"`

	// CountryCode is the country code (for formatting, optional if E.164)
	CountryCode string `json:"country_code,omitempty"`

	// Template is the SMS template to use (optional)
	Template string `json:"template,omitempty"`

	// ExpirationMinutes is how long the code is valid (optional)
	ExpirationMinutes int `json:"expiration_minutes,omitempty"`

	// Locale is the user's locale for localization
	Locale string `json:"locale,omitempty"`

	// Metadata contains additional context
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SendSMSResponse contains the result of sending a verification SMS.
type SendSMSResponse struct {
	// VerificationID is the unique identifier for this verification
	VerificationID string `json:"verification_id"`

	// SentAt is when the SMS was sent
	SentAt time.Time `json:"sent_at"`

	// ExpiresAt is when the verification expires
	ExpiresAt time.Time `json:"expires_at"`

	// MaskedPhone is the masked phone number (e.g., +1***5678)
	MaskedPhone string `json:"masked_phone"`

	// PhoneType is the type of phone (mobile, voip, landline)
	PhoneType string `json:"phone_type,omitempty"`
}

// VerifySMSRequest contains parameters for verifying an SMS code.
type VerifySMSRequest struct {
	// VerificationID is the verification identifier
	VerificationID string `json:"verification_id"`

	// Code is the verification code entered by the user
	Code string `json:"code"`

	// AccountAddress is the expected account address
	AccountAddress string `json:"account_address"`
}

// VerifySMSResponse contains the result of SMS verification.
type VerifySMSResponse struct {
	// Verified indicates if the phone was verified
	Verified bool `json:"verified"`

	// PhoneNumber is the verified phone number
	PhoneNumber string `json:"phone_number"`

	// VerifiedAt is when the phone was verified
	VerifiedAt time.Time `json:"verified_at"`

	// ErrorCode is the error code if verification failed
	ErrorCode string `json:"error_code,omitempty"`

	// ErrorMessage is the error message if verification failed
	ErrorMessage string `json:"error_message,omitempty"`

	// AttemptsRemaining is the number of attempts remaining
	AttemptsRemaining int `json:"attempts_remaining,omitempty"`
}

// SMSVerificationStatus contains the current status of an SMS verification.
type SMSVerificationStatus struct {
	// VerificationID is the verification identifier
	VerificationID string `json:"verification_id"`

	// Status is the current status
	Status VerificationStatus `json:"status"`

	// PhoneNumber is the phone number being verified
	PhoneNumber string `json:"phone_number"`

	// CreatedAt is when the verification was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the verification expires
	ExpiresAt time.Time `json:"expires_at"`

	// VerifiedAt is when it was verified (if verified)
	VerifiedAt *time.Time `json:"verified_at,omitempty"`

	// AttemptCount is the number of verification attempts
	AttemptCount int `json:"attempt_count"`

	// MaxAttempts is the maximum allowed attempts
	MaxAttempts int `json:"max_attempts"`
}

// PhoneTypeResult contains the result of a phone type check.
type PhoneTypeResult struct {
	// PhoneNumber is the phone number checked
	PhoneNumber string `json:"phone_number"`

	// Type is the phone type (mobile, voip, landline, unknown)
	Type string `json:"type"`

	// Carrier is the carrier name (if available)
	Carrier string `json:"carrier,omitempty"`

	// CountryCode is the country code
	CountryCode string `json:"country_code"`

	// IsVOIP indicates if the number is a VOIP number
	IsVOIP bool `json:"is_voip"`

	// RiskLevel is the risk level for this phone type
	RiskLevel string `json:"risk_level"`
}

// ============================================================================
// SSO Provider
// ============================================================================

// SSOProvider defines the interface for SSO verification providers.
type SSOProvider interface {
	Provider

	// GetAuthorizationURL returns the URL to redirect for SSO authentication.
	GetAuthorizationURL(ctx context.Context, req SSOAuthRequest) (*SSOAuthURLResponse, error)

	// HandleCallback handles the SSO callback.
	HandleCallback(ctx context.Context, req SSOCallbackRequest) (*SSOCallbackResponse, error)

	// GetUserInfo retrieves user information from the SSO provider.
	GetUserInfo(ctx context.Context, accessToken string) (*SSOUserInfo, error)

	// RefreshToken refreshes an access token.
	RefreshToken(ctx context.Context, refreshToken string) (*SSOTokenResponse, error)

	// RevokeToken revokes an access token.
	RevokeToken(ctx context.Context, token string) error
}

// SSOAuthRequest contains parameters for initiating SSO authentication.
type SSOAuthRequest struct {
	// AccountAddress is the associated blockchain account
	AccountAddress string `json:"account_address"`

	// RedirectURL is the URL to redirect after authentication
	RedirectURL string `json:"redirect_url"`

	// State is the CSRF protection state parameter
	State string `json:"state"`

	// Nonce is the nonce for ID token validation
	Nonce string `json:"nonce,omitempty"`

	// Scopes is the list of requested scopes
	Scopes []string `json:"scopes,omitempty"`

	// Metadata contains additional context
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SSOAuthURLResponse contains the authorization URL.
type SSOAuthURLResponse struct {
	// AuthorizationURL is the URL to redirect the user to
	AuthorizationURL string `json:"authorization_url"`

	// State is the state parameter to verify on callback
	State string `json:"state"`

	// ExpiresAt is when the authorization attempt expires
	ExpiresAt time.Time `json:"expires_at"`
}

// SSOCallbackRequest contains parameters from the SSO callback.
type SSOCallbackRequest struct {
	// Code is the authorization code
	Code string `json:"code"`

	// State is the state parameter for CSRF verification
	State string `json:"state"`

	// AccountAddress is the expected account address
	AccountAddress string `json:"account_address"`

	// ExpectedNonce is the expected nonce value
	ExpectedNonce string `json:"expected_nonce,omitempty"`
}

// SSOCallbackResponse contains the result of processing the callback.
type SSOCallbackResponse struct {
	// Verified indicates if SSO verification succeeded
	Verified bool `json:"verified"`

	// UserInfo contains the user information
	UserInfo *SSOUserInfo `json:"user_info,omitempty"`

	// Tokens contains the OAuth tokens
	Tokens *SSOTokenResponse `json:"tokens,omitempty"`

	// VerifiedAt is when the verification completed
	VerifiedAt time.Time `json:"verified_at"`

	// ErrorCode is the error code if verification failed
	ErrorCode string `json:"error_code,omitempty"`

	// ErrorMessage is the error message if verification failed
	ErrorMessage string `json:"error_message,omitempty"`
}

// SSOUserInfo contains user information from the SSO provider.
type SSOUserInfo struct {
	// SubjectID is the unique user identifier at the provider
	SubjectID string `json:"subject_id"`

	// Email is the user's email address
	Email string `json:"email,omitempty"`

	// EmailVerified indicates if the email is verified
	EmailVerified bool `json:"email_verified,omitempty"`

	// Name is the user's full name
	Name string `json:"name,omitempty"`

	// GivenName is the user's first name
	GivenName string `json:"given_name,omitempty"`

	// FamilyName is the user's last name
	FamilyName string `json:"family_name,omitempty"`

	// Picture is the URL to the user's profile picture
	Picture string `json:"picture,omitempty"`

	// Locale is the user's locale
	Locale string `json:"locale,omitempty"`

	// Provider is the SSO provider name
	Provider string `json:"provider"`

	// Raw contains the raw claims from the provider
	Raw map[string]interface{} `json:"raw,omitempty"`
}

// SSOTokenResponse contains OAuth tokens.
type SSOTokenResponse struct {
	// AccessToken is the access token
	AccessToken string `json:"access_token"`

	// TokenType is the token type (usually "Bearer")
	TokenType string `json:"token_type"`

	// ExpiresIn is the token expiration time in seconds
	ExpiresIn int `json:"expires_in"`

	// RefreshToken is the refresh token (if available)
	RefreshToken string `json:"refresh_token,omitempty"`

	// IDToken is the ID token (for OIDC providers)
	IDToken string `json:"id_token,omitempty"`

	// Scope is the granted scope
	Scope string `json:"scope,omitempty"`
}

// ============================================================================
// Provider Registry
// ============================================================================

// ProviderRegistry manages verification providers.
type ProviderRegistry interface {
	// RegisterEmailProvider registers an email provider.
	RegisterEmailProvider(name string, provider EmailProvider) error

	// RegisterSMSProvider registers an SMS provider.
	RegisterSMSProvider(name string, provider SMSProvider) error

	// RegisterSSOProvider registers an SSO provider.
	RegisterSSOProvider(name string, provider SSOProvider) error

	// GetEmailProvider returns an email provider by name.
	GetEmailProvider(name string) (EmailProvider, error)

	// GetSMSProvider returns an SMS provider by name.
	GetSMSProvider(name string) (SMSProvider, error)

	// GetSSOProvider returns an SSO provider by name.
	GetSSOProvider(name string) (SSOProvider, error)

	// ListProviders returns all registered providers.
	ListProviders() map[ProviderType][]string

	// Close closes all providers.
	Close() error
}

