package types

import (
	errorsmod "cosmossdk.io/errors"
)

// ============================================================================
// Cryptographic Signature Verification Errors (VE-3022)
// Error codes 1130-1139 reserved for signature verification
// ============================================================================

var (
	// ErrInvalidPublicKeyLength is returned when a public key has invalid length
	ErrInvalidPublicKeyLength = errorsmod.Register(ModuleName, 1130, "invalid public key length")

	// ErrInvalidSignatureLength is returned when a signature has invalid length
	ErrInvalidSignatureLength = errorsmod.Register(ModuleName, 1131, "invalid signature length")

	// ErrSignatureVerificationFailed is returned when signature verification fails
	ErrSignatureVerificationFailed = errorsmod.Register(ModuleName, 1132, "signature verification failed")

	// ErrClientKeyMismatch is returned when client key doesn't match expected key
	ErrClientKeyMismatch = errorsmod.Register(ModuleName, 1133, "client key mismatch")

	// ErrSaltBindingInvalid is returned when salt binding verification fails
	ErrSaltBindingInvalid = errorsmod.Register(ModuleName, 1134, "salt binding verification failed")

	// ErrUnsupportedSignatureAlgorithm is returned when an unsupported algorithm is specified
	ErrUnsupportedSignatureAlgorithm = errorsmod.Register(ModuleName, 1135, "unsupported signature algorithm")

	// ErrPublicKeyMismatch is returned when the public key doesn't derive to expected address
	ErrPublicKeyMismatch = errorsmod.Register(ModuleName, 1136, "public key does not match address")

	// ErrInvalidSaltBindingPayload is returned when salt binding payload is malformed
	ErrInvalidSaltBindingPayload = errorsmod.Register(ModuleName, 1137, "invalid salt binding payload")

	// ErrTimestampOutOfRange is returned when salt binding timestamp is too old or in future
	ErrTimestampOutOfRange = errorsmod.Register(ModuleName, 1138, "timestamp out of acceptable range")
)
