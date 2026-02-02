// Package sso provides the SSO/OIDC verification service for VEID.
package sso

import "errors"

// ============================================================================
// Error Definitions
// ============================================================================

var (
	// ErrChallengeNotFound indicates the challenge was not found.
	ErrChallengeNotFound = errors.New("SSO challenge not found")

	// ErrChallengeExpired indicates the challenge has expired.
	ErrChallengeExpired = errors.New("SSO challenge has expired")

	// ErrChallengeAlreadyCompleted indicates the challenge was already completed.
	ErrChallengeAlreadyCompleted = errors.New("SSO challenge already completed")

	// ErrStateMismatch indicates the OAuth state parameter doesn't match.
	ErrStateMismatch = errors.New("OAuth state mismatch")

	// ErrNonceMismatch indicates the OIDC nonce doesn't match.
	ErrNonceMismatch = errors.New("OIDC nonce mismatch")

	// ErrInvalidLinkageSignature indicates the linkage signature is invalid.
	ErrInvalidLinkageSignature = errors.New("invalid linkage signature")

	// ErrAccountMismatch indicates the account address doesn't match.
	ErrAccountMismatch = errors.New("account address mismatch")

	// ErrLinkageAlreadyExists indicates an SSO linkage already exists.
	ErrLinkageAlreadyExists = errors.New("SSO linkage already exists for this account")

	// ErrLinkageNotFound indicates the SSO linkage was not found.
	ErrLinkageNotFound = errors.New("SSO linkage not found")

	// ErrRateLimitExceeded indicates the rate limit was exceeded.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrServiceUnavailable indicates the service is unavailable.
	ErrServiceUnavailable = errors.New("SSO verification service unavailable")

	// ErrAttestationFailed indicates attestation creation failed.
	ErrAttestationFailed = errors.New("failed to create SSO attestation")

	// ErrSigningFailed indicates attestation signing failed.
	ErrSigningFailed = errors.New("failed to sign SSO attestation")

	// ErrOnChainSubmissionFailed indicates on-chain submission failed.
	ErrOnChainSubmissionFailed = errors.New("failed to submit linkage on-chain")

	// ErrMaxChallengesExceeded indicates too many pending challenges.
	ErrMaxChallengesExceeded = errors.New("maximum pending challenges exceeded")

	// ErrInvalidConfig indicates invalid configuration.
	ErrInvalidConfig = errors.New("invalid SSO service configuration")

	// ErrProviderNotConfigured indicates the provider is not configured.
	ErrProviderNotConfigured = errors.New("SSO provider not configured")

	// ErrRevocationFailed indicates revocation failed.
	ErrRevocationFailed = errors.New("SSO linkage revocation failed")

	// ErrUnauthorizedRevocation indicates unauthorized revocation attempt.
	ErrUnauthorizedRevocation = errors.New("unauthorized to revoke SSO linkage")
)
