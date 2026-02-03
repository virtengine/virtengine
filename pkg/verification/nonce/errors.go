// Package nonce provides replay protection storage for verification attestations.
package nonce

import (
	"cosmossdk.io/errors"
)

// Error codes for the nonce package
var (
	// ErrNonceNotFound indicates the nonce was not found
	ErrNonceNotFound = errors.Register("verification/nonce", 1, "nonce not found")

	// ErrNonceAlreadyExists indicates the nonce already exists
	ErrNonceAlreadyExists = errors.Register("verification/nonce", 2, "nonce already exists")

	// ErrNonceAlreadyUsed indicates the nonce was already used
	ErrNonceAlreadyUsed = errors.Register("verification/nonce", 3, "nonce already used")

	// ErrNonceExpired indicates the nonce has expired
	ErrNonceExpired = errors.Register("verification/nonce", 4, "nonce expired")

	// ErrStoreClosed indicates the store is closed
	ErrStoreClosed = errors.Register("verification/nonce", 5, "store is closed")

	// ErrStoreFull indicates the store is full
	ErrStoreFull = errors.Register("verification/nonce", 6, "store is full")

	// ErrTooManyNonces indicates too many nonces for an issuer
	ErrTooManyNonces = errors.Register("verification/nonce", 7, "too many nonces for issuer")

	// ErrNonceGeneration indicates nonce generation failed
	ErrNonceGeneration = errors.Register("verification/nonce", 8, "nonce generation failed")

	// ErrInvalidNonce indicates the nonce is invalid
	ErrInvalidNonce = errors.Register("verification/nonce", 9, "invalid nonce")

	// ErrConnectionError indicates a connection error
	ErrConnectionError = errors.Register("verification/nonce", 10, "connection error")
)
