// Package sms provides SMS verification service errors.
//
// Task Reference: VE-4C - SMS Verification Delivery + Anti-Fraud
package sms

import (
	"github.com/virtengine/virtengine/pkg/errors"
)

const smsVerificationModule = "sms_verification"

// Error codes for SMS verification service
var (
	// ErrInvalidPhoneNumber indicates an invalid phone number
	ErrInvalidPhoneNumber = errors.NewCodedError(smsVerificationModule, 1, "invalid phone number", errors.CategoryValidation)

	// ErrInvalidRequest indicates an invalid request
	ErrInvalidRequest = errors.NewCodedError(smsVerificationModule, 2, "invalid request", errors.CategoryValidation)

	// ErrInvalidConfig indicates invalid configuration
	ErrInvalidConfig = errors.NewCodedError(smsVerificationModule, 3, "invalid configuration", errors.CategoryValidation)

	// ErrChallengeCreation indicates failure to create a challenge
	ErrChallengeCreation = errors.NewCodedError(smsVerificationModule, 4, "failed to create challenge", errors.CategoryInternal)

	// ErrChallengeNotFound indicates the challenge was not found
	ErrChallengeNotFound = errors.NewCodedError(smsVerificationModule, 5, "challenge not found", errors.CategoryNotFound)

	// ErrChallengeExpired indicates the challenge has expired
	ErrChallengeExpired = errors.NewCodedError(smsVerificationModule, 6, "challenge expired", errors.CategoryState)

	// ErrMaxAttemptsExceeded indicates max verification attempts exceeded
	ErrMaxAttemptsExceeded = errors.NewCodedError(smsVerificationModule, 7, "max verification attempts exceeded", errors.CategoryState)

	// ErrResendLimitExceeded indicates max resends exceeded
	ErrResendLimitExceeded = errors.NewCodedError(smsVerificationModule, 8, "resend limit exceeded", errors.CategoryRateLimit)

	// ErrResendCooldown indicates resend cooldown is active
	ErrResendCooldown = errors.NewCodedError(smsVerificationModule, 9, "resend cooldown active", errors.CategoryRateLimit)

	// ErrInvalidOTP indicates the provided OTP is invalid
	ErrInvalidOTP = errors.NewCodedError(smsVerificationModule, 10, "invalid OTP", errors.CategoryValidation)

	// ErrAccountMismatch indicates the account does not match the challenge
	ErrAccountMismatch = errors.NewCodedError(smsVerificationModule, 11, "account address mismatch", errors.CategoryUnauthorized)

	// ErrDeliveryFailed indicates SMS delivery failed
	ErrDeliveryFailed = errors.NewCodedError(smsVerificationModule, 12, "SMS delivery failed", errors.CategoryExternal)

	// ErrProviderError indicates an SMS provider error
	ErrProviderError = errors.NewCodedError(smsVerificationModule, 13, "SMS provider error", errors.CategoryExternal)

	// ErrRateLimited indicates the request is rate limited
	ErrRateLimited = errors.NewCodedError(smsVerificationModule, 14, "rate limit exceeded", errors.CategoryRateLimit)

	// ErrServiceUnavailable indicates the service is unavailable
	ErrServiceUnavailable = errors.NewCodedError(smsVerificationModule, 15, "service unavailable", errors.CategoryExternal)

	// ErrCacheError indicates a cache error
	ErrCacheError = errors.NewCodedError(smsVerificationModule, 16, "cache error", errors.CategoryInternal)

	// ErrAttestationFailed indicates attestation creation failed
	ErrAttestationFailed = errors.NewCodedError(smsVerificationModule, 17, "attestation creation failed", errors.CategoryInternal)

	// ErrAlreadyVerified indicates the phone is already verified
	ErrAlreadyVerified = errors.NewCodedError(smsVerificationModule, 18, "phone already verified", errors.CategoryConflict)

	// ErrChallengeConsumed indicates the challenge has already been used
	ErrChallengeConsumed = errors.NewCodedError(smsVerificationModule, 19, "challenge already consumed", errors.CategoryState)

	// ErrVoIPDetected indicates a VoIP number was detected
	ErrVoIPDetected = errors.NewCodedError(smsVerificationModule, 20, "VoIP numbers not allowed", errors.CategoryValidation)

	// ErrPhoneBlocked indicates the phone number is blocked
	ErrPhoneBlocked = errors.NewCodedError(smsVerificationModule, 21, "phone number blocked", errors.CategoryValidation)

	// ErrSigningFailed indicates attestation signing failed
	ErrSigningFailed = errors.NewCodedError(smsVerificationModule, 22, "attestation signing failed", errors.CategoryInternal)

	// ErrCarrierLookupFailed indicates carrier lookup failed
	ErrCarrierLookupFailed = errors.NewCodedError(smsVerificationModule, 23, "carrier lookup failed", errors.CategoryExternal)

	// ErrSuspiciousActivity indicates suspicious activity was detected
	ErrSuspiciousActivity = errors.NewCodedError(smsVerificationModule, 24, "suspicious activity detected", errors.CategoryRateLimit)

	// ErrVelocityExceeded indicates velocity limits were exceeded
	ErrVelocityExceeded = errors.NewCodedError(smsVerificationModule, 25, "velocity limit exceeded", errors.CategoryRateLimit)

	// ErrDeviceFingerprintMismatch indicates device fingerprint doesn't match
	ErrDeviceFingerprintMismatch = errors.NewCodedError(smsVerificationModule, 26, "device fingerprint mismatch", errors.CategoryUnauthorized)

	// ErrPrimaryProviderFailed indicates the primary SMS provider failed
	ErrPrimaryProviderFailed = errors.NewCodedError(smsVerificationModule, 27, "primary SMS provider failed", errors.CategoryExternal)

	// ErrAllProvidersFailed indicates all SMS providers failed
	ErrAllProvidersFailed = errors.NewCodedError(smsVerificationModule, 28, "all SMS providers failed", errors.CategoryExternal)

	// ErrRegionNotSupported indicates the phone region is not supported
	ErrRegionNotSupported = errors.NewCodedError(smsVerificationModule, 29, "phone region not supported", errors.CategoryValidation)

	// ErrInvalidCountryCode indicates an invalid country code
	ErrInvalidCountryCode = errors.NewCodedError(smsVerificationModule, 30, "invalid country code", errors.CategoryValidation)

	// ErrVoIPNotAllowed indicates VoIP numbers are not allowed
	ErrVoIPNotAllowed = errors.NewCodedError(smsVerificationModule, 31, "VoIP numbers not allowed", errors.CategoryValidation)

	// ErrCountryBlocked indicates the country is blocked
	ErrCountryBlocked = errors.NewCodedError(smsVerificationModule, 32, "country is blocked", errors.CategoryValidation)

	// ErrPhoneMismatch indicates the phone number doesn't match the challenge
	ErrPhoneMismatch = errors.NewCodedError(smsVerificationModule, 33, "phone number mismatch", errors.CategoryValidation)

	// ErrTemplateError indicates a template rendering error
	ErrTemplateError = errors.NewCodedError(smsVerificationModule, 34, "template rendering error", errors.CategoryInternal)

	// ErrWebhookInvalid indicates an invalid webhook payload
	ErrWebhookInvalid = errors.NewCodedError(smsVerificationModule, 35, "invalid webhook payload", errors.CategoryValidation)
)

// Helper functions for wrapping errors with additional context

// WrapInvalidPhone wraps an error with invalid phone context
func WrapInvalidPhone(msg string) error {
	return errors.Wrap(ErrInvalidPhoneNumber, msg)
}

// WrapChallengeNotFound wraps an error with challenge not found context
func WrapChallengeNotFound(challengeID string) error {
	return errors.Wrapf(ErrChallengeNotFound, "challenge ID: %s", challengeID)
}

// WrapVoIPDetected wraps an error indicating VoIP was detected
func WrapVoIPDetected(reason string) error {
	return errors.Wrap(ErrVoIPDetected, reason)
}

// WrapProviderError wraps a provider error with additional context
func WrapProviderError(provider string, err error) error {
	return errors.Wrapf(ErrProviderError, "provider %s: %v", provider, err)
}
