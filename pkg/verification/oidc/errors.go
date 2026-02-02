// Package oidc provides OIDC token verification for the VEID SSO verification service.
package oidc

import (
	"errors"
	"fmt"
)

// ============================================================================
// Error Definitions
// ============================================================================

var (
	// ErrInvalidToken indicates an invalid OIDC token.
	ErrInvalidToken = errors.New("invalid OIDC token")

	// ErrTokenExpired indicates an expired token.
	ErrTokenExpired = errors.New("OIDC token has expired")

	// ErrInvalidIssuer indicates an invalid or unallowed issuer.
	ErrInvalidIssuer = errors.New("invalid or unallowed OIDC issuer")

	// ErrInvalidAudience indicates an invalid audience claim.
	ErrInvalidAudience = errors.New("invalid OIDC audience")

	// ErrInvalidNonce indicates a nonce mismatch (replay attempt).
	ErrInvalidNonce = errors.New("invalid or mismatched nonce")

	// ErrMissingClaim indicates a required claim is missing.
	ErrMissingClaim = errors.New("required claim is missing")

	// ErrInvalidSignature indicates the token signature is invalid.
	ErrInvalidSignature = errors.New("invalid token signature")

	// ErrJWKSFetchFailed indicates JWKS fetch failed.
	ErrJWKSFetchFailed = errors.New("failed to fetch JWKS")

	// ErrKeyNotFound indicates the signing key was not found in JWKS.
	ErrKeyNotFound = errors.New("signing key not found in JWKS")

	// ErrDiscoveryFailed indicates OIDC discovery failed.
	ErrDiscoveryFailed = errors.New("OIDC discovery failed")

	// ErrInvalidIssuerPolicy indicates an invalid issuer policy configuration.
	ErrInvalidIssuerPolicy = errors.New("invalid issuer policy")

	// ErrIssuerNotAllowed indicates the issuer is not in the allowlist.
	ErrIssuerNotAllowed = errors.New("issuer not in allowlist")

	// ErrEmailNotVerified indicates the email was not verified.
	ErrEmailNotVerified = errors.New("email not verified by provider")

	// ErrEmailDomainNotAllowed indicates the email domain is not allowed.
	ErrEmailDomainNotAllowed = errors.New("email domain not allowed")

	// ErrTenantNotAllowed indicates the tenant is not allowed.
	ErrTenantNotAllowed = errors.New("tenant not allowed")

	// ErrAuthTooOld indicates the authentication is too old.
	ErrAuthTooOld = errors.New("authentication is too old")

	// ErrCodeExchangeFailed indicates the code exchange failed.
	ErrCodeExchangeFailed = errors.New("authorization code exchange failed")

	// ErrServiceUnavailable indicates the service is unavailable.
	ErrServiceUnavailable = errors.New("OIDC verification service unavailable")

	// ErrInvalidACR indicates an invalid or unallowed ACR value.
	ErrInvalidACR = errors.New("invalid or unallowed authentication context class")

	// ErrConfigurationError indicates a configuration error.
	ErrConfigurationError = errors.New("OIDC configuration error")
)

// WrapError wraps an error with additional context.
func WrapError(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}
