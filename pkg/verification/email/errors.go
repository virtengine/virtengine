// Package email provides email verification service errors.
package email

import (
	"github.com/virtengine/virtengine/pkg/errors"
)

const emailVerificationModule = "email_verification"

// Error codes for email verification service
var (
	// ErrInvalidEmail indicates an invalid email address
	ErrInvalidEmail = errors.NewCodedError(emailVerificationModule, 1, "invalid email address", errors.CategoryValidation)

	// ErrInvalidRequest indicates an invalid request
	ErrInvalidRequest = errors.NewCodedError(emailVerificationModule, 2, "invalid request", errors.CategoryValidation)

	// ErrInvalidConfig indicates invalid configuration
	ErrInvalidConfig = errors.NewCodedError(emailVerificationModule, 3, "invalid configuration", errors.CategoryValidation)

	// ErrChallengeCreation indicates failure to create a challenge
	ErrChallengeCreation = errors.NewCodedError(emailVerificationModule, 4, "failed to create challenge", errors.CategoryInternal)

	// ErrChallengeNotFound indicates the challenge was not found
	ErrChallengeNotFound = errors.NewCodedError(emailVerificationModule, 5, "challenge not found", errors.CategoryNotFound)

	// ErrChallengeExpired indicates the challenge has expired
	ErrChallengeExpired = errors.NewCodedError(emailVerificationModule, 6, "challenge expired", errors.CategoryState)

	// ErrMaxAttemptsExceeded indicates max verification attempts exceeded
	ErrMaxAttemptsExceeded = errors.NewCodedError(emailVerificationModule, 7, "max verification attempts exceeded", errors.CategoryState)

	// ErrResendLimitExceeded indicates max resends exceeded
	ErrResendLimitExceeded = errors.NewCodedError(emailVerificationModule, 8, "resend limit exceeded", errors.CategoryRateLimit)

	// ErrResendCooldown indicates resend cooldown is active
	ErrResendCooldown = errors.NewCodedError(emailVerificationModule, 9, "resend cooldown active", errors.CategoryRateLimit)

	// ErrInvalidSecret indicates the provided secret is invalid
	ErrInvalidSecret = errors.NewCodedError(emailVerificationModule, 10, "invalid verification secret", errors.CategoryValidation)

	// ErrAccountMismatch indicates the account does not match the challenge
	ErrAccountMismatch = errors.NewCodedError(emailVerificationModule, 11, "account address mismatch", errors.CategoryUnauthorized)

	// ErrDeliveryFailed indicates email delivery failed
	ErrDeliveryFailed = errors.NewCodedError(emailVerificationModule, 12, "email delivery failed", errors.CategoryExternal)

	// ErrProviderError indicates an email provider error
	ErrProviderError = errors.NewCodedError(emailVerificationModule, 13, "email provider error", errors.CategoryExternal)

	// ErrRateLimited indicates the request is rate limited
	ErrRateLimited = errors.NewCodedError(emailVerificationModule, 14, "rate limit exceeded", errors.CategoryRateLimit)

	// ErrServiceUnavailable indicates the service is unavailable
	ErrServiceUnavailable = errors.NewCodedError(emailVerificationModule, 15, "service unavailable", errors.CategoryExternal)

	// ErrCacheError indicates a cache error
	ErrCacheError = errors.NewCodedError(emailVerificationModule, 16, "cache error", errors.CategoryInternal)

	// ErrTemplateError indicates an email template error
	ErrTemplateError = errors.NewCodedError(emailVerificationModule, 17, "template rendering error", errors.CategoryInternal)

	// ErrWebhookInvalid indicates an invalid webhook request
	ErrWebhookInvalid = errors.NewCodedError(emailVerificationModule, 18, "invalid webhook", errors.CategoryValidation)

	// ErrAttestationFailed indicates attestation creation failed
	ErrAttestationFailed = errors.NewCodedError(emailVerificationModule, 19, "attestation creation failed", errors.CategoryInternal)

	// ErrAlreadyVerified indicates the email is already verified
	ErrAlreadyVerified = errors.NewCodedError(emailVerificationModule, 20, "email already verified", errors.CategoryConflict)

	// ErrChallengeConsumed indicates the challenge has already been used
	ErrChallengeConsumed = errors.NewCodedError(emailVerificationModule, 21, "challenge already consumed", errors.CategoryState)

	// ErrDisposableEmail indicates a disposable email address
	ErrDisposableEmail = errors.NewCodedError(emailVerificationModule, 22, "disposable email not allowed", errors.CategoryValidation)

	// ErrBlockedDomain indicates a blocked email domain
	ErrBlockedDomain = errors.NewCodedError(emailVerificationModule, 23, "email domain blocked", errors.CategoryValidation)

	// ErrSigningFailed indicates attestation signing failed
	ErrSigningFailed = errors.NewCodedError(emailVerificationModule, 24, "attestation signing failed", errors.CategoryInternal)
)

// Helper functions for wrapping errors with additional context

// WrapInvalidEmail wraps an error with invalid email context
func WrapInvalidEmail(msg string) error {
	return errors.Wrap(ErrInvalidEmail, msg)
}

// WrapChallengeNotFound wraps an error with challenge not found context
func WrapChallengeNotFound(challengeID string) error {
	return errors.Wrapf(ErrChallengeNotFound, "challenge ID: %s", challengeID)
}
