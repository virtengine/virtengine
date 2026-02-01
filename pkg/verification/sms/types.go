// Package sms provides SMS OTP verification service for VEID identity attestations.
//
// This package implements SMS OTP verification with:
// - Primary/secondary SMS gateway failover
// - Secure OTP generation and hashing (never store plaintext)
// - VoIP detection and phone number validation
// - Rate limiting, velocity checks, and device/IP heuristics
// - Signed verification attestations for on-chain storage
// - Delivery status tracking and monitoring
// - Localized templates and region-based rate limits
//
// Task Reference: VE-4C - SMS Verification Delivery + Anti-Fraud
package sms

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/virtengine/virtengine/pkg/errors"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Configuration Constants
// ============================================================================

const (
	// DefaultOTPLength is the default length of SMS OTP codes
	DefaultOTPLength = 6

	// DefaultOTPTTLSeconds is the default OTP time-to-live (5 minutes)
	DefaultOTPTTLSeconds int64 = 300

	// DefaultMaxOTPAttempts is the maximum OTP verification attempts
	DefaultMaxOTPAttempts uint32 = 3

	// DefaultMaxResends is the maximum OTP resends per verification session
	DefaultMaxResends uint32 = 3

	// DefaultResendCooldownSeconds is the cooldown between resends
	DefaultResendCooldownSeconds int64 = 60

	// DefaultVerificationValidityDays is how long a verification remains valid
	DefaultVerificationValidityDays = 365

	// OTPHashSaltLength is the length of the salt for OTP hashing
	OTPHashSaltLength = 16

	// DefaultVelocityWindowMinutes is the window for velocity checks
	DefaultVelocityWindowMinutes = 60

	// DefaultMaxRequestsPerPhonePerHour is max requests per phone per hour
	DefaultMaxRequestsPerPhonePerHour = 3

	// DefaultMaxRequestsPerIPPerHour is max requests per IP per hour
	DefaultMaxRequestsPerIPPerHour = 10

	// DefaultMaxRequestsPerAccountPerDay is max requests per account per day
	DefaultMaxRequestsPerAccountPerDay = 10
)

// ============================================================================
// Phone Number Types
// ============================================================================

// CarrierType represents the type of phone carrier
type CarrierType string

const (
	// CarrierTypeMobile indicates a mobile carrier
	CarrierTypeMobile CarrierType = "mobile"

	// CarrierTypeLandline indicates a landline carrier
	CarrierTypeLandline CarrierType = "landline"

	// CarrierTypeVoIP indicates a VoIP carrier
	CarrierTypeVoIP CarrierType = "voip"

	// CarrierTypeUnknown indicates an unknown carrier type
	CarrierTypeUnknown CarrierType = "unknown"
)

// PhoneInfo contains validated phone number information
type PhoneInfo struct {
	// E164 is the phone number in E.164 format
	E164 string `json:"e164"`

	// CountryCode is the ISO 3166-1 alpha-2 country code
	CountryCode string `json:"country_code"`

	// CountryCallingCode is the numeric calling code (e.g., "1" for US)
	CountryCallingCode string `json:"country_calling_code"`

	// NationalNumber is the national significant number
	NationalNumber string `json:"national_number"`

	// CarrierType is the type of carrier
	CarrierType CarrierType `json:"carrier_type"`

	// CarrierName is the name of the carrier
	CarrierName string `json:"carrier_name,omitempty"`

	// IsVoIP indicates if this is a VoIP number
	IsVoIP bool `json:"is_voip"`

	// IsMobile indicates if this is a mobile number
	IsMobile bool `json:"is_mobile"`

	// IsValid indicates if the phone number is valid
	IsValid bool `json:"is_valid"`

	// RiskScore is a risk score (0-100, higher = riskier)
	RiskScore uint32 `json:"risk_score"`

	// RiskFactors lists detected risk factors
	RiskFactors []string `json:"risk_factors,omitempty"`
}

// HashPhoneNumber creates a SHA256 HMAC hash of a phone number
func HashPhoneNumber(phoneNumber string) string {
	hash := sha256.Sum256([]byte(phoneNumber))
	return hex.EncodeToString(hash[:])
}

// NormalizePhoneNumber normalizes a phone number to E.164 format
// This is a simplified implementation - production should use libphonenumber
func NormalizePhoneNumber(phone string, defaultCountryCode string) (string, error) {
	// Remove all non-digit characters except leading +
	hasPlus := strings.HasPrefix(phone, "+")
	digits := regexp.MustCompile(`[^0-9]`).ReplaceAllString(phone, "")

	if len(digits) == 0 {
		return "", errors.Wrap(ErrInvalidPhoneNumber, "no digits in phone number")
	}

	// If already has +, it's already E.164-ish
	if hasPlus {
		return "+" + digits, nil
	}

	// If starts with country code, add +
	if len(digits) >= 10 {
		// Common international prefixes
		if strings.HasPrefix(digits, "1") && len(digits) == 11 { // US/Canada
			return "+" + digits, nil
		}
		if strings.HasPrefix(digits, "44") && len(digits) >= 12 { // UK
			return "+" + digits, nil
		}
		if strings.HasPrefix(digits, "91") && len(digits) >= 12 { // India
			return "+" + digits, nil
		}
	}

	// Apply default country code
	if defaultCountryCode != "" {
		code := strings.TrimPrefix(defaultCountryCode, "+")
		return "+" + code + digits, nil
	}

	return "", errors.Wrap(ErrInvalidPhoneNumber, "cannot determine country code")
}

// MaskPhoneNumber masks a phone number for display
// Example: +14155551234 -> +1***...1234
func MaskPhoneNumber(phoneNumber string) string {
	if len(phoneNumber) < 6 {
		return "****"
	}
	prefix := phoneNumber[:3]
	suffix := phoneNumber[len(phoneNumber)-4:]
	return fmt.Sprintf("%s***...%s", prefix, suffix)
}

// ============================================================================
// Challenge Types
// ============================================================================

// ChallengeStatus represents the status of an SMS verification challenge
type ChallengeStatus string

const (
	// StatusPending indicates the challenge is awaiting verification
	StatusPending ChallengeStatus = "pending"

	// StatusSent indicates the SMS was sent successfully
	StatusSent ChallengeStatus = "sent"

	// StatusDelivered indicates the SMS was delivered
	StatusDelivered ChallengeStatus = "delivered"

	// StatusVerified indicates successful verification
	StatusVerified ChallengeStatus = "verified"

	// StatusFailed indicates verification failed (max attempts exceeded)
	StatusFailed ChallengeStatus = "failed"

	// StatusExpired indicates the challenge has expired
	StatusExpired ChallengeStatus = "expired"

	// StatusCancelled indicates the challenge was cancelled
	StatusCancelled ChallengeStatus = "cancelled"

	// StatusBlocked indicates the phone was blocked (VoIP/fraud)
	StatusBlocked ChallengeStatus = "blocked"

	// StatusRevoked indicates the verification was revoked
	StatusRevoked ChallengeStatus = "revoked"
)

// DeliveryStatus represents the SMS delivery status
type DeliveryStatus string

const (
	// DeliveryPending indicates delivery is pending
	DeliveryPending DeliveryStatus = "pending"

	// DeliverySent indicates SMS was sent to provider
	DeliverySent DeliveryStatus = "sent"

	// DeliveryDelivered indicates SMS was delivered
	DeliveryDelivered DeliveryStatus = "delivered"

	// DeliveryFailed indicates delivery failed
	DeliveryFailed DeliveryStatus = "failed"

	// DeliveryUndelivered indicates SMS was not delivered
	DeliveryUndelivered DeliveryStatus = "undelivered"
)

// SMSChallenge represents a pending SMS verification challenge
type SMSChallenge struct {
	// ChallengeID is a unique identifier for this challenge
	ChallengeID string `json:"challenge_id"`

	// VerificationID links to the parent verification record
	VerificationID string `json:"verification_id,omitempty"`

	// AccountAddress is the account requesting SMS verification
	AccountAddress string `json:"account_address"`

	// PhoneHash is a SHA256 hash of the E.164 phone number
	PhoneHash string `json:"phone_hash"`

	// CountryCode is the ISO country code
	CountryCode string `json:"country_code"`

	// OTPHash is a hash of the OTP (never store plaintext)
	OTPHash string `json:"otp_hash"`

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

	// DeliveryStatus tracks SMS delivery status
	DeliveryStatus DeliveryStatus `json:"delivery_status"`

	// ProviderMessageID is the SMS provider's message ID
	ProviderMessageID string `json:"provider_message_id,omitempty"`

	// Provider is which SMS provider was used
	Provider string `json:"provider,omitempty"`

	// MaskedPhone is the masked phone for display (e.g., +1***...1234)
	MaskedPhone string `json:"masked_phone"`

	// CarrierType is the detected carrier type
	CarrierType CarrierType `json:"carrier_type"`

	// IsVoIP indicates if VoIP was detected
	IsVoIP bool `json:"is_voip"`

	// IPAddress is the IP address of the requester
	IPAddress string `json:"ip_address,omitempty"`

	// IPHash is a hash of the IP address
	IPHash string `json:"ip_hash,omitempty"`

	// DeviceFingerprint is a hash of device identifiers
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`

	// UserAgent is the user agent of the requester
	UserAgent string `json:"user_agent,omitempty"`

	// VerifiedAt is when verification succeeded
	VerifiedAt *time.Time `json:"verified_at,omitempty"`

	// IsConsumed indicates if the challenge has been used
	IsConsumed bool `json:"is_consumed"`

	// RiskScore is the anti-fraud risk score (0-100)
	RiskScore uint32 `json:"risk_score"`

	// RiskFactors lists detected risk factors
	RiskFactors []string `json:"risk_factors,omitempty"`

	// Locale is the locale for message templates
	Locale string `json:"locale,omitempty"`

	// FailoverUsed indicates if failover provider was used
	FailoverUsed bool `json:"failover_used"`
}

// ChallengeConfig contains configuration for creating an SMS challenge
type ChallengeConfig struct {
	ChallengeID       string
	VerificationID    string
	AccountAddress    string
	PhoneNumber       string // Will be hashed, not stored
	CountryCode       string
	CreatedAt         time.Time
	TTLSeconds        int64
	MaxAttempts       uint32
	MaxResends        uint32
	IPAddress         string
	DeviceFingerprint string
	UserAgent         string
	Locale            string
	PhoneInfo         *PhoneInfo
}

// NewSMSChallenge creates a new SMS verification challenge
func NewSMSChallenge(cfg ChallengeConfig) (*SMSChallenge, string, error) {
	if cfg.PhoneNumber == "" {
		return nil, "", errors.Wrap(ErrInvalidPhoneNumber, "phone number cannot be empty")
	}
	if cfg.AccountAddress == "" {
		return nil, "", errors.Wrap(ErrInvalidRequest, "account address cannot be empty")
	}

	// Hash phone number
	phoneHash := HashPhoneNumber(cfg.PhoneNumber)

	// Generate nonce
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, "", errors.Wrapf(ErrChallengeCreation, "failed to generate nonce: %v", err)
	}
	nonce := hex.EncodeToString(nonceBytes)

	// Generate OTP
	otp, otpHash, err := GenerateOTP(DefaultOTPLength)
	if err != nil {
		return nil, "", err
	}

	// Calculate expiry
	ttl := cfg.TTLSeconds
	if ttl <= 0 {
		ttl = DefaultOTPTTLSeconds
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

	// Hash IP address
	ipHash := ""
	if cfg.IPAddress != "" {
		ipHash = HashPhoneNumber(cfg.IPAddress) // Reuse hash function
	}

	challenge := &SMSChallenge{
		ChallengeID:       cfg.ChallengeID,
		VerificationID:    cfg.VerificationID,
		AccountAddress:    cfg.AccountAddress,
		PhoneHash:         phoneHash,
		CountryCode:       cfg.CountryCode,
		OTPHash:           otpHash,
		Nonce:             nonce,
		Status:            StatusPending,
		CreatedAt:         cfg.CreatedAt,
		ExpiresAt:         expiresAt,
		MaxAttempts:       maxAttempts,
		MaxResends:        maxResends,
		DeliveryStatus:    DeliveryPending,
		MaskedPhone:       MaskPhoneNumber(cfg.PhoneNumber),
		IPAddress:         cfg.IPAddress,
		IPHash:            ipHash,
		DeviceFingerprint: cfg.DeviceFingerprint,
		UserAgent:         cfg.UserAgent,
		Locale:            cfg.Locale,
	}

	// Apply phone info if available
	if cfg.PhoneInfo != nil {
		challenge.CarrierType = cfg.PhoneInfo.CarrierType
		challenge.IsVoIP = cfg.PhoneInfo.IsVoIP
		challenge.RiskScore = cfg.PhoneInfo.RiskScore
		challenge.RiskFactors = cfg.PhoneInfo.RiskFactors
	}

	return challenge, otp, nil
}

// Validate validates the SMS challenge
func (c *SMSChallenge) Validate() error {
	if c.ChallengeID == "" {
		return errors.Wrap(ErrInvalidRequest, "challenge_id cannot be empty")
	}
	if c.AccountAddress == "" {
		return errors.Wrap(ErrInvalidRequest, "account address cannot be empty")
	}
	if c.PhoneHash == "" {
		return errors.Wrap(ErrInvalidPhoneNumber, "phone_hash cannot be empty")
	}
	if len(c.PhoneHash) != 64 { // SHA256 hex
		return errors.Wrap(ErrInvalidPhoneNumber, "phone_hash must be a valid SHA256 hex string")
	}
	if c.OTPHash == "" {
		return errors.Wrap(ErrInvalidRequest, "otp_hash cannot be empty")
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
func (c *SMSChallenge) IsExpired(now time.Time) bool {
	return now.After(c.ExpiresAt)
}

// CanAttempt returns true if another verification attempt is allowed
func (c *SMSChallenge) CanAttempt() bool {
	return !c.IsConsumed && c.Attempts < c.MaxAttempts && c.Status != StatusFailed && c.Status != StatusBlocked
}

// CanResend returns true if the OTP can be resent
func (c *SMSChallenge) CanResend(now time.Time) bool {
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
func (c *SMSChallenge) RecordAttempt(attemptTime time.Time, success bool) {
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

// RecordResend records an OTP resend
func (c *SMSChallenge) RecordResend(resendTime time.Time, newOTPHash string, newExpiresAt time.Time) error {
	if !c.CanResend(resendTime) {
		return ErrResendLimitExceeded
	}
	c.ResendCount++
	c.LastResendAt = &resendTime
	c.OTPHash = newOTPHash
	c.ExpiresAt = newExpiresAt
	c.Attempts = 0 // Reset attempts on resend
	c.DeliveryStatus = DeliveryPending
	return nil
}

// VerifyOTP checks if the provided OTP matches
func (c *SMSChallenge) VerifyOTP(providedOTP string) bool {
	if c.IsConsumed {
		return false
	}
	providedHash := HashOTP(providedOTP)
	return providedHash == c.OTPHash
}

// UpdateDeliveryStatus updates the delivery status
func (c *SMSChallenge) UpdateDeliveryStatus(status DeliveryStatus, messageID string, provider string) {
	c.DeliveryStatus = status
	if messageID != "" {
		c.ProviderMessageID = messageID
	}
	if provider != "" {
		c.Provider = provider
	}
	if status == DeliveryDelivered {
		c.Status = StatusDelivered
	} else if status == DeliverySent {
		c.Status = StatusSent
	}
}

// MarkBlocked marks the challenge as blocked
func (c *SMSChallenge) MarkBlocked(reason string) {
	c.Status = StatusBlocked
	c.IsConsumed = true
	if reason != "" {
		c.RiskFactors = append(c.RiskFactors, reason)
	}
}

// String returns a string representation of the SMS challenge
func (c *SMSChallenge) String() string {
	return fmt.Sprintf("SMSChallenge{ID: %s, Status: %s, Attempts: %d/%d, IsVoIP: %t}",
		c.ChallengeID, c.Status, c.Attempts, c.MaxAttempts, c.IsVoIP)
}

// ============================================================================
// OTP Generation
// ============================================================================

// GenerateOTP generates a cryptographically secure OTP code
// Returns the OTP (to send via SMS) and its hash (to store)
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
	otpHash = HashOTP(otp)

	return otp, otpHash, nil
}

// HashOTP creates a SHA256 hash of an OTP
func HashOTP(otp string) string {
	hash := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(hash[:])
}

// ============================================================================
// Request/Response Types
// ============================================================================

// InitiateRequest is the request to initiate SMS verification
type InitiateRequest struct {
	// AccountAddress is the account requesting verification
	AccountAddress string `json:"account_address"`

	// PhoneNumber is the phone number to verify (E.164 format preferred)
	PhoneNumber string `json:"phone_number"`

	// CountryCode is the ISO 3166-1 alpha-2 country code (e.g., "US")
	CountryCode string `json:"country_code,omitempty"`

	// RequestID is an optional request identifier for correlation
	RequestID string `json:"request_id,omitempty"`

	// IPAddress is the client IP address
	IPAddress string `json:"ip_address,omitempty"`

	// DeviceFingerprint is a hash of device identifiers
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`

	// UserAgent is the client user agent
	UserAgent string `json:"user_agent,omitempty"`

	// Locale is the preferred locale for the SMS
	Locale string `json:"locale,omitempty"`

	// BypassVoIPCheck bypasses VoIP detection (for testing only)
	BypassVoIPCheck bool `json:"bypass_voip_check,omitempty"`
}

// Validate validates the initiate request
func (r *InitiateRequest) Validate() error {
	if r.AccountAddress == "" {
		return errors.Wrap(ErrInvalidRequest, "account_address is required")
	}
	if r.PhoneNumber == "" {
		return errors.Wrap(ErrInvalidPhoneNumber, "phone_number is required")
	}
	return nil
}

// InitiateResponse is the response from initiating SMS verification
type InitiateResponse struct {
	// ChallengeID is the unique identifier for the verification challenge
	ChallengeID string `json:"challenge_id"`

	// MaskedPhone is the masked phone number
	MaskedPhone string `json:"masked_phone"`

	// ExpiresAt is when the challenge expires
	ExpiresAt time.Time `json:"expires_at"`

	// ResendCooldownSeconds is how long to wait before resending
	ResendCooldownSeconds int64 `json:"resend_cooldown_seconds"`

	// CarrierType is the detected carrier type
	CarrierType CarrierType `json:"carrier_type,omitempty"`

	// IsVoIP indicates if VoIP was detected (may be blocked)
	IsVoIP bool `json:"is_voip,omitempty"`
}

// VerifyRequest is the request to verify an SMS challenge
type VerifyRequest struct {
	// ChallengeID is the challenge to verify
	ChallengeID string `json:"challenge_id"`

	// OTP is the OTP code from the SMS
	OTP string `json:"otp"`

	// AccountAddress is the account verifying (must match challenge)
	AccountAddress string `json:"account_address"`

	// RequestID is an optional request identifier
	RequestID string `json:"request_id,omitempty"`

	// IPAddress is the client IP address
	IPAddress string `json:"ip_address,omitempty"`

	// DeviceFingerprint is a hash of device identifiers
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`
}

// Validate validates the verify request
func (r *VerifyRequest) Validate() error {
	if r.ChallengeID == "" {
		return errors.Wrap(ErrInvalidRequest, "challenge_id is required")
	}
	if r.OTP == "" {
		return errors.Wrap(ErrInvalidOTP, "otp is required")
	}
	if r.AccountAddress == "" {
		return errors.Wrap(ErrInvalidRequest, "account_address is required")
	}
	return nil
}

// VerifyResponse is the response from verifying an SMS challenge
type VerifyResponse struct {
	// Success indicates if verification succeeded
	Success bool `json:"success"`

	// Verified indicates if the phone is now verified
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

// ResendRequest is the request to resend verification SMS
type ResendRequest struct {
	// ChallengeID is the challenge to resend for
	ChallengeID string `json:"challenge_id"`

	// AccountAddress is the requesting account
	AccountAddress string `json:"account_address"`

	// PhoneNumber is required for resend (we don't store plaintext)
	PhoneNumber string `json:"phone_number"`

	// RequestID is an optional request identifier
	RequestID string `json:"request_id,omitempty"`
}

// ResendResponse is the response from resending verification SMS
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
// Delivery Result
// ============================================================================

// DeliveryResult represents the result of an SMS delivery attempt
type DeliveryResult struct {
	// ChallengeID is the challenge this delivery is for
	ChallengeID string `json:"challenge_id"`

	// Success indicates if the SMS was sent successfully
	Success bool `json:"success"`

	// ProviderMessageID is the provider's message ID for tracking
	ProviderMessageID string `json:"provider_message_id,omitempty"`

	// DeliveryStatus is the current delivery status
	DeliveryStatus DeliveryStatus `json:"delivery_status"`

	// SentAt is when the SMS was sent
	SentAt time.Time `json:"sent_at"`

	// ErrorCode is the error code if delivery failed
	ErrorCode string `json:"error_code,omitempty"`

	// ErrorMessage is the error message if delivery failed
	ErrorMessage string `json:"error_message,omitempty"`

	// Provider is which SMS provider was used
	Provider string `json:"provider"`

	// FailoverUsed indicates if failover provider was used
	FailoverUsed bool `json:"failover_used"`
}

// ============================================================================
// Webhook Event Types
// ============================================================================

// WebhookEventType represents the type of webhook event
type WebhookEventType string

const (
	// WebhookEventDelivered indicates SMS was delivered
	WebhookEventDelivered WebhookEventType = "delivered"

	// WebhookEventFailed indicates SMS delivery failed
	WebhookEventFailed WebhookEventType = "failed"

	// WebhookEventUndelivered indicates SMS was not delivered
	WebhookEventUndelivered WebhookEventType = "undelivered"

	// WebhookEventSent indicates SMS was sent
	WebhookEventSent WebhookEventType = "sent"
)

// WebhookEvent represents a delivery status webhook event
type WebhookEvent struct {
	// EventType is the type of event
	EventType WebhookEventType `json:"event_type"`

	// MessageID is the provider's message ID
	MessageID string `json:"message_id"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// PhoneHash is the hashed phone number
	PhoneHash string `json:"phone_hash"`

	// ErrorCode is the error code (if applicable)
	ErrorCode string `json:"error_code,omitempty"`

	// ErrorMessage is the error message (if applicable)
	ErrorMessage string `json:"error_message,omitempty"`

	// Provider is the SMS provider
	Provider string `json:"provider"`

	// Raw is the raw webhook payload (for debugging)
	Raw map[string]interface{} `json:"raw,omitempty"`
}

// ============================================================================
// Service Configuration
// ============================================================================

// Config contains configuration for the SMS verification service
type Config struct {
	// PrimaryProvider is the primary SMS provider
	PrimaryProvider string `json:"primary_provider"`

	// SecondaryProvider is the failover SMS provider
	SecondaryProvider string `json:"secondary_provider,omitempty"`

	// ProviderConfigs contains provider-specific configurations
	ProviderConfigs map[string]ProviderConfig `json:"provider_configs"`

	// OTPLength is the length of OTP codes
	OTPLength int `json:"otp_length"`

	// OTPTTLSeconds is the OTP time-to-live in seconds
	OTPTTLSeconds int64 `json:"otp_ttl_seconds"`

	// MaxAttempts is the maximum verification attempts
	MaxAttempts uint32 `json:"max_attempts"`

	// MaxResends is the maximum resends per challenge
	MaxResends uint32 `json:"max_resends"`

	// ResendCooldownSeconds is the cooldown between resends
	ResendCooldownSeconds int64 `json:"resend_cooldown_seconds"`

	// VerificationValidityDays is how long a verification remains valid
	VerificationValidityDays int `json:"verification_validity_days"`

	// EnableVoIPBlocking blocks VoIP numbers
	EnableVoIPBlocking bool `json:"enable_voip_blocking"`

	// EnableCarrierLookup enables carrier lookup
	EnableCarrierLookup bool `json:"enable_carrier_lookup"`

	// EnableVelocityChecks enables velocity-based anti-fraud
	EnableVelocityChecks bool `json:"enable_velocity_checks"`

	// MaxRequestsPerPhonePerHour is the max requests per phone per hour
	MaxRequestsPerPhonePerHour int `json:"max_requests_per_phone_per_hour"`

	// MaxRequestsPerIPPerHour is the max requests per IP per hour
	MaxRequestsPerIPPerHour int `json:"max_requests_per_ip_per_hour"`

	// MaxRequestsPerAccountPerDay is the max requests per account per day
	MaxRequestsPerAccountPerDay int `json:"max_requests_per_account_per_day"`

	// BlockedCountryCodes is a list of blocked country codes
	BlockedCountryCodes []string `json:"blocked_country_codes,omitempty"`

	// AllowedCountryCodes is a list of allowed country codes (if set, only these are allowed)
	AllowedCountryCodes []string `json:"allowed_country_codes,omitempty"`

	// DefaultLocale is the default locale for SMS templates
	DefaultLocale string `json:"default_locale"`

	// RateLimitEnabled enables rate limiting
	RateLimitEnabled bool `json:"rate_limit_enabled"`

	// AuditLogEnabled enables audit logging
	AuditLogEnabled bool `json:"audit_log_enabled"`

	// MetricsEnabled enables Prometheus metrics
	MetricsEnabled bool `json:"metrics_enabled"`

	// FailoverEnabled enables automatic failover to secondary provider
	FailoverEnabled bool `json:"failover_enabled"`

	// WebhookSecret is the secret for validating webhooks
	WebhookSecret string `json:"webhook_secret,omitempty"`
}

// ProviderConfig contains configuration for a specific SMS provider
type ProviderConfig struct {
	// Type is the provider type (twilio, aws_sns, vonage, etc.)
	Type string `json:"type"`

	// AccountSID is the account identifier (Twilio)
	AccountSID string `json:"account_sid,omitempty"`

	// AuthToken is the authentication token
	AuthToken string `json:"auth_token,omitempty"`

	// APIKey is the API key
	APIKey string `json:"api_key,omitempty"`

	// APISecret is the API secret
	APISecret string `json:"api_secret,omitempty"`

	// FromNumber is the sender phone number
	FromNumber string `json:"from_number"`

	// MessagingServiceSID is the messaging service ID (Twilio)
	MessagingServiceSID string `json:"messaging_service_sid,omitempty"`

	// Region is the AWS region (SNS)
	Region string `json:"region,omitempty"`

	// SenderID is the alphanumeric sender ID
	SenderID string `json:"sender_id,omitempty"`

	// SupportedRegions is the list of supported country codes
	SupportedRegions []string `json:"supported_regions,omitempty"`

	// WebhookURL is the webhook URL for delivery status
	WebhookURL string `json:"webhook_url,omitempty"`

	// WebhookSecret is the secret for webhook signature validation
	WebhookSecret string `json:"webhook_secret,omitempty"`

	// Timeout is the HTTP request timeout in seconds
	Timeout int `json:"timeout,omitempty"`

	// Enabled indicates if this provider is enabled
	Enabled bool `json:"enabled"`
}

// DefaultConfig returns the default SMS verification configuration
func DefaultConfig() Config {
	return Config{
		PrimaryProvider:             "twilio",
		OTPLength:                   DefaultOTPLength,
		OTPTTLSeconds:               DefaultOTPTTLSeconds,
		MaxAttempts:                 DefaultMaxOTPAttempts,
		MaxResends:                  DefaultMaxResends,
		ResendCooldownSeconds:       DefaultResendCooldownSeconds,
		VerificationValidityDays:    DefaultVerificationValidityDays,
		EnableVoIPBlocking:          true,
		EnableCarrierLookup:         true,
		EnableVelocityChecks:        true,
		MaxRequestsPerPhonePerHour:  DefaultMaxRequestsPerPhonePerHour,
		MaxRequestsPerIPPerHour:     DefaultMaxRequestsPerIPPerHour,
		MaxRequestsPerAccountPerDay: DefaultMaxRequestsPerAccountPerDay,
		DefaultLocale:               "en",
		RateLimitEnabled:            true,
		AuditLogEnabled:             true,
		MetricsEnabled:              true,
		FailoverEnabled:             true,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.PrimaryProvider == "" {
		return errors.Wrap(ErrInvalidConfig, "primary_provider is required")
	}
	if c.OTPLength < 4 || c.OTPLength > 10 {
		return errors.Wrap(ErrInvalidConfig, "otp_length must be between 4 and 10")
	}
	if c.OTPTTLSeconds < 60 || c.OTPTTLSeconds > 600 {
		return errors.Wrap(ErrInvalidConfig, "otp_ttl_seconds must be between 60 and 600")
	}
	if c.MaxAttempts < 1 || c.MaxAttempts > 10 {
		return errors.Wrap(ErrInvalidConfig, "max_attempts must be between 1 and 10")
	}
	return nil
}

// IsCountryAllowed checks if a country code is allowed
func (c *Config) IsCountryAllowed(countryCode string) bool {
	// Check blocked list
	for _, blocked := range c.BlockedCountryCodes {
		if strings.EqualFold(blocked, countryCode) {
			return false
		}
	}
	// If allowed list is set, check it
	if len(c.AllowedCountryCodes) > 0 {
		for _, allowed := range c.AllowedCountryCodes {
			if strings.EqualFold(allowed, countryCode) {
				return true
			}
		}
		return false
	}
	return true
}

// ============================================================================
// Service Interface
// ============================================================================

// SMSVerificationService defines the interface for the SMS verification service
type SMSVerificationService interface {
	// InitiateVerification starts a new SMS verification challenge
	InitiateVerification(ctx context.Context, req *InitiateRequest) (*InitiateResponse, error)

	// VerifyChallenge verifies an SMS challenge with the provided OTP
	VerifyChallenge(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error)

	// ResendVerification resends the verification SMS
	ResendVerification(ctx context.Context, req *ResendRequest) (*ResendResponse, error)

	// GetChallenge retrieves a challenge by ID
	GetChallenge(ctx context.Context, challengeID string) (*SMSChallenge, error)

	// CancelChallenge cancels an active challenge
	CancelChallenge(ctx context.Context, challengeID string, accountAddress string) error

	// ProcessWebhook processes a delivery status webhook
	ProcessWebhook(ctx context.Context, provider string, payload []byte) error

	// GetDeliveryStatus returns the delivery status for a challenge
	GetDeliveryStatus(ctx context.Context, challengeID string) (*DeliveryResult, error)

	// LookupPhoneInfo performs carrier/VoIP lookup on a phone number
	LookupPhoneInfo(ctx context.Context, phoneNumber string) (*PhoneInfo, error)

	// HealthCheck returns the health status of the service
	HealthCheck(ctx context.Context) (*HealthStatus, error)

	// Close closes the service and releases resources
	Close() error
}

// HealthStatus represents the health status of the SMS verification service
type HealthStatus struct {
	// Healthy indicates if the service is healthy
	Healthy bool `json:"healthy"`

	// Status is a human-readable status message
	Status string `json:"status"`

	// PrimaryProviderHealthy indicates if the primary provider is healthy
	PrimaryProviderHealthy bool `json:"primary_provider_healthy"`

	// SecondaryProviderHealthy indicates if the secondary provider is healthy
	SecondaryProviderHealthy bool `json:"secondary_provider_healthy,omitempty"`

	// CacheHealthy indicates if the cache is healthy
	CacheHealthy bool `json:"cache_healthy"`

	// Details contains detailed health information
	Details map[string]interface{} `json:"details,omitempty"`

	// Timestamp is when the health check was performed
	Timestamp time.Time `json:"timestamp"`
}

// ============================================================================
// Attestation Types
// ============================================================================

// SMSVerificationAttestation extends VerificationAttestation for SMS
type SMSVerificationAttestation struct {
	*veidtypes.VerificationAttestation

	// PhoneHash is the hash of the verified phone number
	PhoneHash string `json:"phone_hash"`

	// CountryCode is the country code
	CountryCode string `json:"country_code"`

	// CarrierType is the carrier type
	CarrierType CarrierType `json:"carrier_type"`

	// IsVoIP indicates if VoIP was detected
	IsVoIP bool `json:"is_voip"`
}

