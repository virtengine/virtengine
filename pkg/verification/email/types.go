// Package email provides email verification service for VEID identity attestations.
//
// This package implements email OTP/link verification with:
// - Secure OTP generation and hashing (never store plaintext)
// - Transactional email provider integration (SES, SendGrid)
// - Rate limiting and abuse detection
// - Signed verification attestations for on-chain storage
// - Delivery status tracking with webhooks
//
// Task Reference: VE-3F - Email Verification Delivery + Attestation
package email

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/virtengine/virtengine/pkg/errors"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Configuration Constants
// ============================================================================

const (
	// DefaultOTPLength is the default length of email OTP codes
	DefaultOTPLength = 6

	// DefaultOTPTTLSeconds is the default OTP time-to-live (10 minutes)
	DefaultOTPTTLSeconds int64 = 600

	// DefaultLinkTTLSeconds is the default magic link time-to-live (24 hours)
	DefaultLinkTTLSeconds int64 = 86400

	// DefaultMaxOTPAttempts is the maximum OTP verification attempts
	DefaultMaxOTPAttempts uint32 = 5

	// DefaultMaxResends is the maximum OTP resends per verification session
	DefaultMaxResends uint32 = 3

	// DefaultResendCooldownSeconds is the cooldown between resends
	DefaultResendCooldownSeconds int64 = 60

	// DefaultVerificationValidityDays is how long a verification remains valid
	DefaultVerificationValidityDays = 365

	// OTPHashSaltLength is the length of the salt for OTP hashing
	OTPHashSaltLength = 16

	// MagicLinkTokenLength is the length of magic link tokens
	MagicLinkTokenLength = 32
)

// ============================================================================
// Verification Types
// ============================================================================

// VerificationMethod defines the verification method used
type VerificationMethod string

const (
	// MethodOTP uses a one-time password sent via email
	MethodOTP VerificationMethod = "otp"

	// MethodMagicLink uses a verification link sent via email
	MethodMagicLink VerificationMethod = "magic_link"
)

// ChallengeStatus represents the status of an email verification challenge
type ChallengeStatus string

const (
	// StatusPending indicates the challenge is awaiting verification
	StatusPending ChallengeStatus = "pending"

	// StatusDelivered indicates the email was delivered successfully
	StatusDelivered ChallengeStatus = "delivered"

	// StatusVerified indicates successful verification
	StatusVerified ChallengeStatus = "verified"

	// StatusFailed indicates verification failed (max attempts exceeded)
	StatusFailed ChallengeStatus = "failed"

	// StatusExpired indicates the challenge has expired
	StatusExpired ChallengeStatus = "expired"

	// StatusCancelled indicates the challenge was cancelled
	StatusCancelled ChallengeStatus = "cancelled"

	// StatusBounced indicates the email bounced
	StatusBounced ChallengeStatus = "bounced"
)

// DeliveryStatus represents the email delivery status
type DeliveryStatus string

const (
	// DeliveryPending indicates delivery is pending
	DeliveryPending DeliveryStatus = "pending"

	// DeliverySent indicates email was sent to provider
	DeliverySent DeliveryStatus = "sent"

	// DeliveryDelivered indicates email was delivered
	DeliveryDelivered DeliveryStatus = "delivered"

	// DeliveryBounced indicates email bounced
	DeliveryBounced DeliveryStatus = "bounced"

	// DeliveryFailed indicates delivery failed
	DeliveryFailed DeliveryStatus = "failed"

	// DeliveryComplaint indicates a spam complaint
	DeliveryComplaint DeliveryStatus = "complaint"
)

// ============================================================================
// Email Challenge Types
// ============================================================================

// EmailChallenge represents a pending email verification challenge
type EmailChallenge struct {
	// ChallengeID is a unique identifier for this challenge
	ChallengeID string `json:"challenge_id"`

	// AccountAddress is the account requesting email verification
	AccountAddress string `json:"account_address"`

	// EmailHash is a SHA256 hash of the normalized email address
	EmailHash string `json:"email_hash"`

	// DomainHash is a SHA256 hash of the email domain
	DomainHash string `json:"domain_hash"`

	// Method is the verification method (OTP or magic link)
	Method VerificationMethod `json:"method"`

	// OTPHash is a hash of the OTP (only for OTP method)
	OTPHash string `json:"otp_hash,omitempty"`

	// TokenHash is a hash of the magic link token (only for magic link method)
	TokenHash string `json:"token_hash,omitempty"`

	// Nonce is a unique nonce for replay protection
	Nonce string `json:"nonce"`

	// Status is the current challenge status
	Status ChallengeStatus `json:"status"`

	// CreatedAt is when this challenge was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when this challenge expires
	ExpiresAt time.Time `json:"expires_at"`

	// Attempts is the number of verification attempts
	Attempts uint32 `json:"attempts"`

	// MaxAttempts is the maximum allowed attempts
	MaxAttempts uint32 `json:"max_attempts"`

	// ResendCount tracks how many times OTP was resent
	ResendCount uint32 `json:"resend_count"`

	// MaxResends is the maximum allowed resends
	MaxResends uint32 `json:"max_resends"`

	// LastAttemptAt is when the last attempt was made
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`

	// LastResendAt is when the OTP was last resent
	LastResendAt *time.Time `json:"last_resend_at,omitempty"`

	// DeliveryStatus tracks email delivery status
	DeliveryStatus DeliveryStatus `json:"delivery_status"`

	// ProviderMessageID is the email provider's message ID
	ProviderMessageID string `json:"provider_message_id,omitempty"`

	// MaskedEmail is the masked email for display (e.g., t***@example.com)
	MaskedEmail string `json:"masked_email"`

	// IsOrganizational indicates if this is an organizational email domain
	IsOrganizational bool `json:"is_organizational"`

	// IPAddress is the IP address of the requester
	IPAddress string `json:"ip_address,omitempty"`

	// UserAgent is the user agent of the requester
	UserAgent string `json:"user_agent,omitempty"`

	// VerifiedAt is when verification succeeded
	VerifiedAt *time.Time `json:"verified_at,omitempty"`

	// IsConsumed indicates if the challenge has been used
	IsConsumed bool `json:"is_consumed"`
}

// ChallengeConfig contains configuration for creating an email challenge
type ChallengeConfig struct {
	ChallengeID    string
	AccountAddress string
	Email          string // Will be hashed, not stored
	Method         VerificationMethod
	CreatedAt      time.Time
	TTLSeconds     int64
	MaxAttempts    uint32
	MaxResends     uint32
	IPAddress      string
	UserAgent      string
}

// NewEmailChallenge creates a new email verification challenge
func NewEmailChallenge(cfg ChallengeConfig) (*EmailChallenge, string, error) {
	if cfg.Email == "" {
		return nil, "", errors.Wrap(ErrInvalidEmail, "email cannot be empty")
	}
	if cfg.AccountAddress == "" {
		return nil, "", errors.Wrap(ErrInvalidRequest, "account address cannot be empty")
	}

	// Hash email and domain
	emailHash, domainHash := veidtypes.HashEmail(cfg.Email)

	// Generate nonce
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, "", errors.Wrapf(ErrChallengeCreation, "failed to generate nonce: %v", err)
	}
	nonce := hex.EncodeToString(nonceBytes)

	// Generate secret based on method
	var secret, secretHash string
	var err error

	switch cfg.Method {
	case MethodOTP:
		secret, secretHash, err = GenerateOTP(DefaultOTPLength)
	case MethodMagicLink:
		secret, secretHash, err = GenerateMagicLinkToken()
	default:
		return nil, "", errors.Wrapf(ErrInvalidRequest, "unsupported verification method: %s", cfg.Method)
	}
	if err != nil {
		return nil, "", err
	}

	// Calculate expiry
	ttl := cfg.TTLSeconds
	if ttl <= 0 {
		if cfg.Method == MethodOTP {
			ttl = DefaultOTPTTLSeconds
		} else {
			ttl = DefaultLinkTTLSeconds
		}
	}
	expiresAt := cfg.CreatedAt.Add(time.Duration(ttl) * time.Second)

	// Set defaults
	maxAttempts := cfg.MaxAttempts
	if maxAttempts == 0 {
		maxAttempts = DefaultMaxOTPAttempts
	}
	maxResends := cfg.MaxResends
	if maxResends == 0 {
		maxResends = DefaultMaxResends
	}

	challenge := &EmailChallenge{
		ChallengeID:      cfg.ChallengeID,
		AccountAddress:   cfg.AccountAddress,
		EmailHash:        emailHash,
		DomainHash:       domainHash,
		Method:           cfg.Method,
		Nonce:            nonce,
		Status:           StatusPending,
		CreatedAt:        cfg.CreatedAt,
		ExpiresAt:        expiresAt,
		MaxAttempts:      maxAttempts,
		MaxResends:       maxResends,
		DeliveryStatus:   DeliveryPending,
		MaskedEmail:      MaskEmail(cfg.Email),
		IsOrganizational: IsOrganizationalDomain(cfg.Email),
		IPAddress:        cfg.IPAddress,
		UserAgent:        cfg.UserAgent,
	}

	// Set secret hash based on method
	if cfg.Method == MethodOTP {
		challenge.OTPHash = secretHash
	} else {
		challenge.TokenHash = secretHash
	}

	return challenge, secret, nil
}

// Validate validates the email challenge
func (c *EmailChallenge) Validate() error {
	if c.ChallengeID == "" {
		return errors.Wrap(ErrInvalidRequest, "challenge_id cannot be empty")
	}
	if c.AccountAddress == "" {
		return errors.Wrap(ErrInvalidRequest, "account address cannot be empty")
	}
	if c.EmailHash == "" {
		return errors.Wrap(ErrInvalidEmail, "email_hash cannot be empty")
	}
	if len(c.EmailHash) != 64 { // SHA256 hex
		return errors.Wrap(ErrInvalidEmail, "email_hash must be a valid SHA256 hex string")
	}
	if c.Nonce == "" {
		return errors.Wrap(ErrInvalidRequest, "nonce cannot be empty")
	}
	if c.CreatedAt.IsZero() {
		return errors.Wrap(ErrInvalidRequest, "created_at cannot be zero")
	}
	if c.ExpiresAt.IsZero() {
		return errors.Wrap(ErrInvalidRequest, "expires_at cannot be zero")
	}
	if c.MaxAttempts == 0 {
		return errors.Wrap(ErrInvalidRequest, "max_attempts must be positive")
	}
	return nil
}

// IsExpired returns true if the challenge has expired
func (c *EmailChallenge) IsExpired(now time.Time) bool {
	return now.After(c.ExpiresAt)
}

// CanAttempt returns true if another verification attempt is allowed
func (c *EmailChallenge) CanAttempt() bool {
	return !c.IsConsumed && c.Attempts < c.MaxAttempts && c.Status != StatusFailed
}

// CanResend returns true if the secret can be resent
func (c *EmailChallenge) CanResend(now time.Time) bool {
	if c.IsConsumed || c.ResendCount >= c.MaxResends {
		return false
	}
	// Check cooldown
	if c.LastResendAt != nil {
		cooldownEnd := c.LastResendAt.Add(time.Duration(DefaultResendCooldownSeconds) * time.Second)
		if now.Before(cooldownEnd) {
			return false
		}
	}
	return true
}

// RecordAttempt records a verification attempt
func (c *EmailChallenge) RecordAttempt(attemptTime time.Time, success bool) {
	c.Attempts++
	c.LastAttemptAt = &attemptTime
	if success {
		c.IsConsumed = true
		c.Status = StatusVerified
		c.VerifiedAt = &attemptTime
	} else if c.Attempts >= c.MaxAttempts {
		c.Status = StatusFailed
	}
}

// RecordResend records an OTP/link resend
func (c *EmailChallenge) RecordResend(resendTime time.Time, newSecretHash string, newExpiresAt time.Time) error {
	if !c.CanResend(resendTime) {
		return ErrResendLimitExceeded
	}
	c.ResendCount++
	c.LastResendAt = &resendTime

	if c.Method == MethodOTP {
		c.OTPHash = newSecretHash
	} else {
		c.TokenHash = newSecretHash
	}
	c.ExpiresAt = newExpiresAt
	c.Attempts = 0 // Reset attempts on resend
	c.DeliveryStatus = DeliveryPending
	return nil
}

// VerifySecret checks if the provided secret matches
func (c *EmailChallenge) VerifySecret(providedSecret string) bool {
	if c.IsConsumed {
		return false
	}
	providedHash := HashSecret(providedSecret)
	if c.Method == MethodOTP {
		return providedHash == c.OTPHash
	}
	return providedHash == c.TokenHash
}

// UpdateDeliveryStatus updates the delivery status
func (c *EmailChallenge) UpdateDeliveryStatus(status DeliveryStatus, messageID string) {
	c.DeliveryStatus = status
	if messageID != "" {
		c.ProviderMessageID = messageID
	}
	if status == DeliveryDelivered {
		c.Status = StatusDelivered
	} else if status == DeliveryBounced || status == DeliveryFailed {
		c.Status = StatusBounced
	}
}

// ============================================================================
// OTP and Token Generation
// ============================================================================

// GenerateOTP generates a cryptographically secure OTP code
// Returns the OTP (to send via email) and its hash (to store)
func GenerateOTP(length int) (otp string, otpHash string, err error) {
	if length < 4 || length > 10 {
		length = DefaultOTPLength
	}

	// Generate random bytes for OTP
	otpBytes := make([]byte, length)
	if _, err := rand.Read(otpBytes); err != nil {
		return "", "", errors.Wrapf(ErrChallengeCreation, "failed to generate OTP: %v", err)
	}

	// Convert to numeric OTP
	digits := make([]byte, length)
	for i := range otpBytes {
		digits[i] = '0' + (otpBytes[i] % 10)
	}
	otp = string(digits)

	// Hash the OTP for storage
	otpHash = HashSecret(otp)

	return otp, otpHash, nil
}

// GenerateMagicLinkToken generates a cryptographically secure magic link token
func GenerateMagicLinkToken() (token string, tokenHash string, err error) {
	tokenBytes := make([]byte, MagicLinkTokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", errors.Wrapf(ErrChallengeCreation, "failed to generate token: %v", err)
	}

	// Use URL-safe base64 encoding
	token = base64.URLEncoding.EncodeToString(tokenBytes)
	tokenHash = HashSecret(token)

	return token, tokenHash, nil
}

// HashSecret creates a SHA256 hash of a secret (OTP or token)
func HashSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}

// ============================================================================
// Email Utilities
// ============================================================================

// MaskEmail masks an email address for display
// Example: test@example.com -> t***@example.com
func MaskEmail(email string) string {
	if len(email) < 3 {
		return "***"
	}

	atIndex := -1
	for i := range email {
		if email[i] == '@' {
			atIndex = i
			break
		}
	}

	if atIndex <= 0 {
		return "***"
	}

	// Keep first char and domain
	firstChar := string(email[0])
	domain := email[atIndex:]

	if atIndex == 1 {
		return firstChar + "***" + domain
	}

	return firstChar + "***" + domain
}

// IsOrganizationalDomain checks if an email is from a known organizational domain
// This is a simplified check - in production, use a more comprehensive database
func IsOrganizationalDomain(email string) bool {
	// Find domain
	atIndex := -1
	for i := range email {
		if email[i] == '@' {
			atIndex = i
			break
		}
	}
	if atIndex < 0 || atIndex >= len(email)-1 {
		return false
	}
	domain := email[atIndex+1:]

	// Check against common personal email domains
	personalDomains := map[string]bool{
		"gmail.com":      true,
		"yahoo.com":      true,
		"hotmail.com":    true,
		"outlook.com":    true,
		"aol.com":        true,
		"icloud.com":     true,
		"me.com":         true,
		"mail.com":       true,
		"proton.me":      true,
		"protonmail.com": true,
	}

	return !personalDomains[domain]
}

// ============================================================================
// Delivery Result
// ============================================================================

// DeliveryResult represents the result of an email delivery attempt
type DeliveryResult struct {
	// ChallengeID is the challenge this delivery is for
	ChallengeID string `json:"challenge_id"`

	// Success indicates if the email was sent successfully
	Success bool `json:"success"`

	// ProviderMessageID is the provider's message ID for tracking
	ProviderMessageID string `json:"provider_message_id,omitempty"`

	// DeliveryStatus is the current delivery status
	DeliveryStatus DeliveryStatus `json:"delivery_status"`

	// SentAt is when the email was sent
	SentAt time.Time `json:"sent_at"`

	// ErrorCode is the error code if delivery failed
	ErrorCode string `json:"error_code,omitempty"`

	// ErrorMessage is the error message if delivery failed
	ErrorMessage string `json:"error_message,omitempty"`

	// Provider is which email provider was used
	Provider string `json:"provider"`
}

// ============================================================================
// Webhook Event Types
// ============================================================================

// WebhookEventType represents the type of webhook event
type WebhookEventType string

const (
	// WebhookEventDelivered indicates email was delivered
	WebhookEventDelivered WebhookEventType = "delivered"

	// WebhookEventBounced indicates email bounced
	WebhookEventBounced WebhookEventType = "bounced"

	// WebhookEventComplaint indicates spam complaint
	WebhookEventComplaint WebhookEventType = "complaint"

	// WebhookEventOpened indicates email was opened
	WebhookEventOpened WebhookEventType = "opened"

	// WebhookEventClicked indicates link was clicked
	WebhookEventClicked WebhookEventType = "clicked"
)

// WebhookEvent represents a delivery status webhook event
type WebhookEvent struct {
	// EventType is the type of event
	EventType WebhookEventType `json:"event_type"`

	// MessageID is the provider's message ID
	MessageID string `json:"message_id"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// Recipient is the recipient email (hashed)
	RecipientHash string `json:"recipient_hash"`

	// BounceType is the type of bounce (if applicable)
	BounceType string `json:"bounce_type,omitempty"`

	// BounceSubtype is the subtype of bounce
	BounceSubtype string `json:"bounce_subtype,omitempty"`

	// ComplaintType is the type of complaint (if applicable)
	ComplaintType string `json:"complaint_type,omitempty"`

	// Raw is the raw webhook payload (for debugging)
	Raw map[string]interface{} `json:"raw,omitempty"`
}

// ============================================================================
// Verification Request/Response
// ============================================================================

// InitiateRequest is the request to initiate email verification
type InitiateRequest struct {
	// AccountAddress is the account requesting verification
	AccountAddress string `json:"account_address"`

	// Email is the email address to verify
	Email string `json:"email"`

	// Method is the verification method (OTP or magic link)
	Method VerificationMethod `json:"method"`

	// RequestID is an optional request identifier for correlation
	RequestID string `json:"request_id,omitempty"`

	// IPAddress is the client IP address
	IPAddress string `json:"ip_address,omitempty"`

	// UserAgent is the client user agent
	UserAgent string `json:"user_agent,omitempty"`

	// Locale is the preferred locale for the email
	Locale string `json:"locale,omitempty"`
}

// Validate validates the initiate request
func (r *InitiateRequest) Validate() error {
	if r.AccountAddress == "" {
		return errors.Wrap(ErrInvalidRequest, "account_address is required")
	}
	if r.Email == "" {
		return errors.Wrap(ErrInvalidEmail, "email is required")
	}
	if r.Method == "" {
		r.Method = MethodOTP // Default to OTP
	}
	if r.Method != MethodOTP && r.Method != MethodMagicLink {
		return errors.Wrapf(ErrInvalidRequest, "invalid verification method: %s", r.Method)
	}
	return nil
}

// InitiateResponse is the response from initiating email verification
type InitiateResponse struct {
	// ChallengeID is the unique identifier for the verification challenge
	ChallengeID string `json:"challenge_id"`

	// MaskedEmail is the masked email address
	MaskedEmail string `json:"masked_email"`

	// Method is the verification method used
	Method VerificationMethod `json:"method"`

	// ExpiresAt is when the challenge expires
	ExpiresAt time.Time `json:"expires_at"`

	// ResendCooldownSeconds is how long to wait before resending
	ResendCooldownSeconds int64 `json:"resend_cooldown_seconds"`
}

// VerifyRequest is the request to verify an email challenge
type VerifyRequest struct {
	// ChallengeID is the challenge to verify
	ChallengeID string `json:"challenge_id"`

	// Secret is the OTP or magic link token
	Secret string `json:"secret"`

	// AccountAddress is the account verifying (must match challenge)
	AccountAddress string `json:"account_address"`

	// RequestID is an optional request identifier
	RequestID string `json:"request_id,omitempty"`

	// IPAddress is the client IP address
	IPAddress string `json:"ip_address,omitempty"`
}

// Validate validates the verify request
func (r *VerifyRequest) Validate() error {
	if r.ChallengeID == "" {
		return errors.Wrap(ErrInvalidRequest, "challenge_id is required")
	}
	if r.Secret == "" {
		return errors.Wrap(ErrInvalidRequest, "secret is required")
	}
	if r.AccountAddress == "" {
		return errors.Wrap(ErrInvalidRequest, "account_address is required")
	}
	return nil
}

// VerifyResponse is the response from verifying an email challenge
type VerifyResponse struct {
	// Success indicates if verification succeeded
	Success bool `json:"success"`

	// Verified indicates if the email is now verified
	Verified bool `json:"verified"`

	// AttestationID is the ID of the created attestation (if verified)
	AttestationID string `json:"attestation_id,omitempty"`

	// RemainingAttempts is how many attempts remain
	RemainingAttempts uint32 `json:"remaining_attempts"`

	// ErrorCode is the error code if verification failed
	ErrorCode string `json:"error_code,omitempty"`

	// ErrorMessage is the error message if verification failed
	ErrorMessage string `json:"error_message,omitempty"`
}

// ResendRequest is the request to resend verification email
type ResendRequest struct {
	// ChallengeID is the challenge to resend for
	ChallengeID string `json:"challenge_id"`

	// AccountAddress is the requesting account
	AccountAddress string `json:"account_address"`

	// RequestID is an optional request identifier
	RequestID string `json:"request_id,omitempty"`
}

// ResendResponse is the response from resending verification email
type ResendResponse struct {
	// Success indicates if resend was successful
	Success bool `json:"success"`

	// ExpiresAt is the new expiry time
	ExpiresAt time.Time `json:"expires_at"`

	// ResendsRemaining is how many resends remain
	ResendsRemaining uint32 `json:"resends_remaining"`

	// NextResendAt is when the next resend is allowed
	NextResendAt time.Time `json:"next_resend_at"`
}

// ============================================================================
// Service Configuration
// ============================================================================

// Config contains configuration for the email verification service
type Config struct {
	// Provider is the email provider to use (ses, sendgrid)
	Provider string `json:"provider"`

	// ProviderConfig contains provider-specific configuration
	ProviderConfig map[string]string `json:"provider_config"`

	// OTPLength is the length of OTP codes
	OTPLength int `json:"otp_length"`

	// OTPTTLSeconds is the OTP time-to-live in seconds
	OTPTTLSeconds int64 `json:"otp_ttl_seconds"`

	// LinkTTLSeconds is the magic link time-to-live in seconds
	LinkTTLSeconds int64 `json:"link_ttl_seconds"`

	// MaxAttempts is the maximum verification attempts
	MaxAttempts uint32 `json:"max_attempts"`

	// MaxResends is the maximum resends per challenge
	MaxResends uint32 `json:"max_resends"`

	// ResendCooldownSeconds is the cooldown between resends
	ResendCooldownSeconds int64 `json:"resend_cooldown_seconds"`

	// VerificationValidityDays is how long a verification remains valid
	VerificationValidityDays int `json:"verification_validity_days"`

	// FromAddress is the sender email address
	FromAddress string `json:"from_address"`

	// FromName is the sender display name
	FromName string `json:"from_name"`

	// BaseURL is the base URL for magic links
	BaseURL string `json:"base_url"`

	// TemplateDir is the directory containing email templates
	TemplateDir string `json:"template_dir,omitempty"`

	// WebhookSecret is the secret for validating webhooks
	WebhookSecret string `json:"webhook_secret,omitempty"`

	// RateLimitEnabled enables rate limiting
	RateLimitEnabled bool `json:"rate_limit_enabled"`

	// AuditLogEnabled enables audit logging
	AuditLogEnabled bool `json:"audit_log_enabled"`

	// MetricsEnabled enables Prometheus metrics
	MetricsEnabled bool `json:"metrics_enabled"`
}

// DefaultConfig returns the default email verification configuration
func DefaultConfig() Config {
	return Config{
		Provider:                 "ses",
		OTPLength:                DefaultOTPLength,
		OTPTTLSeconds:            DefaultOTPTTLSeconds,
		LinkTTLSeconds:           DefaultLinkTTLSeconds,
		MaxAttempts:              DefaultMaxOTPAttempts,
		MaxResends:               DefaultMaxResends,
		ResendCooldownSeconds:    DefaultResendCooldownSeconds,
		VerificationValidityDays: DefaultVerificationValidityDays,
		FromName:                 "VirtEngine Identity",
		RateLimitEnabled:         true,
		AuditLogEnabled:          true,
		MetricsEnabled:           true,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Provider == "" {
		return errors.Wrap(ErrInvalidConfig, "provider is required")
	}
	if c.FromAddress == "" {
		return errors.Wrap(ErrInvalidConfig, "from_address is required")
	}
	if c.OTPLength < 4 || c.OTPLength > 10 {
		return errors.Wrap(ErrInvalidConfig, "otp_length must be between 4 and 10")
	}
	if c.OTPTTLSeconds < 60 || c.OTPTTLSeconds > 3600 {
		return errors.Wrap(ErrInvalidConfig, "otp_ttl_seconds must be between 60 and 3600")
	}
	if c.MaxAttempts < 1 || c.MaxAttempts > 10 {
		return errors.Wrap(ErrInvalidConfig, "max_attempts must be between 1 and 10")
	}
	return nil
}

// ============================================================================
// Service Interface
// ============================================================================

// EmailVerificationService defines the interface for the email verification service
type EmailVerificationService interface {
	// InitiateVerification starts a new email verification challenge
	InitiateVerification(ctx context.Context, req *InitiateRequest) (*InitiateResponse, error)

	// VerifyChallenge verifies an email challenge with the provided secret
	VerifyChallenge(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error)

	// ResendVerification resends the verification email
	ResendVerification(ctx context.Context, req *ResendRequest) (*ResendResponse, error)

	// GetChallenge retrieves a challenge by ID
	GetChallenge(ctx context.Context, challengeID string) (*EmailChallenge, error)

	// CancelChallenge cancels an active challenge
	CancelChallenge(ctx context.Context, challengeID string, accountAddress string) error

	// ProcessWebhook processes a delivery status webhook
	ProcessWebhook(ctx context.Context, provider string, payload []byte) error

	// GetDeliveryStatus returns the delivery status for a challenge
	GetDeliveryStatus(ctx context.Context, challengeID string) (*DeliveryResult, error)

	// HealthCheck returns the health status of the service
	HealthCheck(ctx context.Context) (*HealthStatus, error)

	// Close closes the service and releases resources
	Close() error
}

// HealthStatus represents the health status of the email verification service
type HealthStatus struct {
	// Healthy indicates if the service is healthy
	Healthy bool `json:"healthy"`

	// Status is a human-readable status message
	Status string `json:"status"`

	// ProviderHealthy indicates if the email provider is healthy
	ProviderHealthy bool `json:"provider_healthy"`

	// CacheHealthy indicates if the cache is healthy
	CacheHealthy bool `json:"cache_healthy"`

	// Details contains detailed health information
	Details map[string]interface{} `json:"details,omitempty"`

	// Timestamp is when the health check was performed
	Timestamp time.Time `json:"timestamp"`
}

// String returns a string representation of the email challenge
func (c *EmailChallenge) String() string {
	return fmt.Sprintf("EmailChallenge{ID: %s, Status: %s, Method: %s, Attempts: %d/%d}",
		c.ChallengeID, c.Status, c.Method, c.Attempts, c.MaxAttempts)
}
