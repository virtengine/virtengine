// Package ratelimit provides verification-specific rate limiting and abuse scoring.
//
// This package extends the base rate limiting infrastructure with verification-specific
// controls, abuse detection, and risk scoring.
//
// Task Reference: VE-2B - Verification Shared Infrastructure
package ratelimit

import (
	"context"
	"time"

	baseratelimit "github.com/virtengine/virtengine/pkg/ratelimit"
)

// ============================================================================
// Verification-Specific Limit Types
// ============================================================================

// VerificationLimitType represents verification-specific limit categories.
type VerificationLimitType string

const (
	// LimitTypeEmailVerification for email verification requests
	LimitTypeEmailVerification VerificationLimitType = "email_verification"

	// LimitTypeSMSVerification for SMS verification requests
	LimitTypeSMSVerification VerificationLimitType = "sms_verification"

	// LimitTypeFacialVerification for facial verification requests
	LimitTypeFacialVerification VerificationLimitType = "facial_verification"

	// LimitTypeLivenessCheck for liveness check requests
	LimitTypeLivenessCheck VerificationLimitType = "liveness_check"

	// LimitTypeDocumentVerification for document verification requests
	LimitTypeDocumentVerification VerificationLimitType = "document_verification"

	// LimitTypeSSOVerification for SSO verification requests
	LimitTypeSSOVerification VerificationLimitType = "sso_verification"

	// LimitTypeAttestationSigning for attestation signing requests
	LimitTypeAttestationSigning VerificationLimitType = "attestation_signing"

	// LimitTypeNonceGeneration for nonce generation requests
	LimitTypeNonceGeneration VerificationLimitType = "nonce_generation"
)

// ============================================================================
// Abuse Scoring
// ============================================================================

// AbuseScore represents the risk score for an identity.
type AbuseScore struct {
	// Identifier is the scored entity (IP, account, etc.)
	Identifier string `json:"identifier"`

	// Score is the abuse score (0-100, higher = more risky)
	Score int `json:"score"`

	// Factors contributing to the score
	Factors []AbuseFactor `json:"factors"`

	// CalculatedAt is when the score was last calculated
	CalculatedAt time.Time `json:"calculated_at"`

	// ExpiresAt is when the score expires
	ExpiresAt time.Time `json:"expires_at"`
}

// AbuseFactor represents a factor contributing to the abuse score.
type AbuseFactor struct {
	// Type identifies the factor type
	Type AbuseFactorType `json:"type"`

	// Weight is the contribution to the total score
	Weight int `json:"weight"`

	// Description explains the factor
	Description string `json:"description"`

	// ObservedAt is when the factor was observed
	ObservedAt time.Time `json:"observed_at"`

	// Details contains additional information
	Details map[string]interface{} `json:"details,omitempty"`
}

// AbuseFactorType identifies types of abuse factors.
type AbuseFactorType string

const (
	// FactorHighRequestRate indicates unusually high request volume
	FactorHighRequestRate AbuseFactorType = "high_request_rate"

	// FactorRepeatedFailures indicates repeated verification failures
	FactorRepeatedFailures AbuseFactorType = "repeated_failures"

	// FactorSuspiciousPattern indicates suspicious behavior patterns
	FactorSuspiciousPattern AbuseFactorType = "suspicious_pattern"

	// FactorKnownAbuser indicates known abuse history
	FactorKnownAbuser AbuseFactorType = "known_abuser"

	// FactorVPNDetected indicates VPN/proxy usage
	FactorVPNDetected AbuseFactorType = "vpn_detected"

	// FactorTorExitNode indicates Tor exit node
	FactorTorExitNode AbuseFactorType = "tor_exit_node"

	// FactorGeographicAnomaly indicates unusual geographic patterns
	FactorGeographicAnomaly AbuseFactorType = "geographic_anomaly"

	// FactorTemporalAnomaly indicates unusual timing patterns
	FactorTemporalAnomaly AbuseFactorType = "temporal_anomaly"

	// FactorBotBehavior indicates bot-like behavior
	FactorBotBehavior AbuseFactorType = "bot_behavior"

	// FactorMultipleIdentities indicates multiple identity attempts
	FactorMultipleIdentities AbuseFactorType = "multiple_identities"

	// FactorVOIPPhone indicates VOIP phone number
	FactorVOIPPhone AbuseFactorType = "voip_phone"

	// FactorDisposableEmail indicates disposable email address
	FactorDisposableEmail AbuseFactorType = "disposable_email"
)

// RiskLevel represents the overall risk level based on abuse score.
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// GetRiskLevel returns the risk level for an abuse score.
func GetRiskLevel(score int) RiskLevel {
	switch {
	case score >= 80:
		return RiskLevelCritical
	case score >= 60:
		return RiskLevelHigh
	case score >= 40:
		return RiskLevelMedium
	default:
		return RiskLevelLow
	}
}

// ============================================================================
// Verification Rate Limiter Interface
// ============================================================================

// VerificationLimiter extends RateLimiter with verification-specific features.
type VerificationLimiter interface {
	// Embed base rate limiter
	baseratelimit.RateLimiter

	// AllowVerification checks if a verification request is allowed.
	AllowVerification(ctx context.Context, identifier string, verificationType VerificationLimitType) (bool, *VerificationLimitResult, error)

	// GetAbuseScore returns the abuse score for an identifier.
	GetAbuseScore(ctx context.Context, identifier string) (*AbuseScore, error)

	// RecordVerificationAttempt records a verification attempt for abuse scoring.
	RecordVerificationAttempt(ctx context.Context, attempt VerificationAttempt) error

	// RecordVerificationResult records the result of a verification for abuse scoring.
	RecordVerificationResult(ctx context.Context, result VerificationResult) error

	// IsBlocked checks if an identifier is blocked due to abuse.
	IsBlocked(ctx context.Context, identifier string) (bool, string, error)

	// BlockIdentifier blocks an identifier due to abuse.
	BlockIdentifier(ctx context.Context, identifier string, duration time.Duration, reason string) error

	// UnblockIdentifier removes a block on an identifier.
	UnblockIdentifier(ctx context.Context, identifier string) error
}

// VerificationLimitResult contains the result of a verification rate limit check.
type VerificationLimitResult struct {
	// Allowed indicates if the request is allowed
	Allowed bool `json:"allowed"`

	// Reason explains why the request was denied (if not allowed)
	Reason string `json:"reason,omitempty"`

	// Remaining is the number of requests remaining
	Remaining int `json:"remaining"`

	// ResetAt is when the limit resets
	ResetAt time.Time `json:"reset_at"`

	// AbuseScore is the current abuse score
	AbuseScore *AbuseScore `json:"abuse_score,omitempty"`

	// RiskLevel is the current risk level
	RiskLevel RiskLevel `json:"risk_level"`

	// RequiresCaptcha indicates if CAPTCHA verification is required
	RequiresCaptcha bool `json:"requires_captcha"`

	// RequiresMFA indicates if MFA is required
	RequiresMFA bool `json:"requires_mfa"`

	// RetryAfter is when the client can retry (seconds)
	RetryAfter int `json:"retry_after,omitempty"`
}

// VerificationAttempt records a verification attempt.
type VerificationAttempt struct {
	// Identifier is the entity making the attempt
	Identifier string `json:"identifier"`

	// Type is the verification type
	Type VerificationLimitType `json:"type"`

	// Timestamp is when the attempt occurred
	Timestamp time.Time `json:"timestamp"`

	// IPAddress is the client IP
	IPAddress string `json:"ip_address"`

	// UserAgent is the client user agent
	UserAgent string `json:"user_agent,omitempty"`

	// AccountAddress is the target account
	AccountAddress string `json:"account_address,omitempty"`

	// RequestID is the request identifier
	RequestID string `json:"request_id,omitempty"`

	// Metadata contains additional context
	Metadata map[string]string `json:"metadata,omitempty"`
}

// VerificationResult records the result of a verification.
type VerificationResult struct {
	// Identifier is the entity that made the attempt
	Identifier string `json:"identifier"`

	// Type is the verification type
	Type VerificationLimitType `json:"type"`

	// Success indicates if verification succeeded
	Success bool `json:"success"`

	// Timestamp is when the result occurred
	Timestamp time.Time `json:"timestamp"`

	// Score is the verification score (if applicable)
	Score int `json:"score,omitempty"`

	// ErrorCode is the error code if failed
	ErrorCode string `json:"error_code,omitempty"`

	// ErrorMessage is the error message if failed
	ErrorMessage string `json:"error_message,omitempty"`

	// Duration is how long the verification took
	Duration time.Duration `json:"duration,omitempty"`

	// RequestID is the request identifier
	RequestID string `json:"request_id,omitempty"`
}

// ============================================================================
// Configuration
// ============================================================================

// VerificationLimiterConfig contains configuration for the verification limiter.
type VerificationLimiterConfig struct {
	// BaseConfig is the base rate limiter configuration
	BaseConfig baseratelimit.RateLimitConfig `json:"base_config"`

	// VerificationLimits contains per-verification-type limits
	VerificationLimits map[VerificationLimitType]VerificationLimits `json:"verification_limits"`

	// AbuseScoring contains abuse scoring configuration
	AbuseScoring AbuseScoringConfig `json:"abuse_scoring"`

	// CaptchaConfig contains CAPTCHA configuration
	CaptchaConfig CaptchaConfig `json:"captcha_config"`
}

// VerificationLimits contains limits for a verification type.
type VerificationLimits struct {
	// RequestsPerHour is the max requests per hour
	RequestsPerHour int `json:"requests_per_hour"`

	// RequestsPerDay is the max requests per day
	RequestsPerDay int `json:"requests_per_day"`

	// MaxFailuresPerHour is the max failures before blocking
	MaxFailuresPerHour int `json:"max_failures_per_hour"`

	// CooldownMinutes is the cooldown after max failures
	CooldownMinutes int `json:"cooldown_minutes"`
}

// AbuseScoringConfig contains abuse scoring configuration.
type AbuseScoringConfig struct {
	// Enabled determines if abuse scoring is active
	Enabled bool `json:"enabled"`

	// ScoreThresholdForCaptcha is the score above which CAPTCHA is required
	ScoreThresholdForCaptcha int `json:"score_threshold_for_captcha"`

	// ScoreThresholdForBlock is the score above which blocking occurs
	ScoreThresholdForBlock int `json:"score_threshold_for_block"`

	// ScoreDecayPerHour is how much the score decays per hour
	ScoreDecayPerHour int `json:"score_decay_per_hour"`

	// FactorWeights contains weights for each abuse factor
	FactorWeights map[AbuseFactorType]int `json:"factor_weights"`
}

// CaptchaConfig contains CAPTCHA configuration.
type CaptchaConfig struct {
	// Enabled determines if CAPTCHA is active
	Enabled bool `json:"enabled"`

	// Provider is the CAPTCHA provider (hcaptcha, recaptcha, turnstile)
	Provider string `json:"provider"`

	// SiteKey is the CAPTCHA site key
	SiteKey string `json:"site_key"`

	// SecretKey is the CAPTCHA secret key
	SecretKey string `json:"secret_key"`
}

// DefaultVerificationLimiterConfig returns the default configuration.
func DefaultVerificationLimiterConfig() VerificationLimiterConfig {
	return VerificationLimiterConfig{
		BaseConfig: baseratelimit.DefaultConfig(),
		VerificationLimits: map[VerificationLimitType]VerificationLimits{
			LimitTypeEmailVerification: {
				RequestsPerHour:    10,
				RequestsPerDay:     50,
				MaxFailuresPerHour: 5,
				CooldownMinutes:    60,
			},
			LimitTypeSMSVerification: {
				RequestsPerHour:    5,
				RequestsPerDay:     20,
				MaxFailuresPerHour: 3,
				CooldownMinutes:    120,
			},
			LimitTypeFacialVerification: {
				RequestsPerHour:    20,
				RequestsPerDay:     100,
				MaxFailuresPerHour: 10,
				CooldownMinutes:    30,
			},
			LimitTypeLivenessCheck: {
				RequestsPerHour:    30,
				RequestsPerDay:     200,
				MaxFailuresPerHour: 15,
				CooldownMinutes:    30,
			},
			LimitTypeDocumentVerification: {
				RequestsPerHour:    10,
				RequestsPerDay:     30,
				MaxFailuresPerHour: 5,
				CooldownMinutes:    60,
			},
			LimitTypeSSOVerification: {
				RequestsPerHour:    20,
				RequestsPerDay:     100,
				MaxFailuresPerHour: 10,
				CooldownMinutes:    30,
			},
			LimitTypeAttestationSigning: {
				RequestsPerHour:    100,
				RequestsPerDay:     1000,
				MaxFailuresPerHour: 20,
				CooldownMinutes:    15,
			},
			LimitTypeNonceGeneration: {
				RequestsPerHour:    200,
				RequestsPerDay:     2000,
				MaxFailuresPerHour: 50,
				CooldownMinutes:    10,
			},
		},
		AbuseScoring: AbuseScoringConfig{
			Enabled:                  true,
			ScoreThresholdForCaptcha: 40,
			ScoreThresholdForBlock:   80,
			ScoreDecayPerHour:        5,
			FactorWeights: map[AbuseFactorType]int{
				FactorHighRequestRate:    20,
				FactorRepeatedFailures:   30,
				FactorSuspiciousPattern:  25,
				FactorKnownAbuser:        50,
				FactorVPNDetected:        10,
				FactorTorExitNode:        40,
				FactorGeographicAnomaly:  15,
				FactorTemporalAnomaly:    15,
				FactorBotBehavior:        35,
				FactorMultipleIdentities: 40,
				FactorVOIPPhone:          20,
				FactorDisposableEmail:    25,
			},
		},
		CaptchaConfig: CaptchaConfig{
			Enabled:  false,
			Provider: "hcaptcha",
		},
	}
}
