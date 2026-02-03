// Package signer provides the verification attestation signing service.
package signer

import (
	"cosmossdk.io/errors"
)

// Error codes for the signer package
var (
	// ErrInvalidConfig indicates invalid signer configuration
	ErrInvalidConfig = errors.Register("verification/signer", 1, "invalid configuration")

	// ErrNoActiveKey indicates no active key is available
	ErrNoActiveKey = errors.Register("verification/signer", 2, "no active key available")

	// ErrKeyNotFound indicates the requested key was not found
	ErrKeyNotFound = errors.Register("verification/signer", 3, "key not found")

	// ErrKeyRevoked indicates the key has been revoked
	ErrKeyRevoked = errors.Register("verification/signer", 4, "key has been revoked")

	// ErrKeyExpired indicates the key has expired
	ErrKeyExpired = errors.Register("verification/signer", 5, "key has expired")

	// ErrSigningFailed indicates a signing operation failed
	ErrSigningFailed = errors.Register("verification/signer", 6, "signing failed")

	// ErrVerificationFailed indicates signature verification failed
	ErrVerificationFailed = errors.Register("verification/signer", 7, "verification failed")

	// ErrRotationInProgress indicates a rotation is already in progress
	ErrRotationInProgress = errors.Register("verification/signer", 8, "key rotation already in progress")

	// ErrRotationNotFound indicates the rotation was not found
	ErrRotationNotFound = errors.Register("verification/signer", 9, "rotation not found")

	// ErrKeyStorageError indicates a key storage operation failed
	ErrKeyStorageError = errors.Register("verification/signer", 10, "key storage error")

	// ErrInvalidAttestation indicates the attestation is invalid
	ErrInvalidAttestation = errors.Register("verification/signer", 11, "invalid attestation")

	// ErrKeyGenerationFailed indicates key generation failed
	ErrKeyGenerationFailed = errors.Register("verification/signer", 12, "key generation failed")

	// ErrInvalidSignature indicates the signature is invalid
	ErrInvalidSignature = errors.Register("verification/signer", 13, "invalid signature")

	// ErrServiceUnavailable indicates the service is unavailable
	ErrServiceUnavailable = errors.Register("verification/signer", 14, "service unavailable")
)
