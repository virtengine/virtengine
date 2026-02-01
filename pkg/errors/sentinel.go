package errors

import (
	"errors"
)

// Common sentinel errors that can be used across modules.
// These should be used with errors.Is() for comparison.

var (
	// ErrNotFound is returned when a resource is not found.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists is returned when a resource already exists.
	ErrAlreadyExists = errors.New("already exists")

	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("invalid input")

	// ErrInvalidAddress is returned when an address is invalid.
	ErrInvalidAddress = errors.New("invalid address")

	// ErrUnauthorized is returned when authorization fails.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when an action is forbidden.
	ErrForbidden = errors.New("forbidden")

	// ErrTimeout is returned when an operation times out.
	ErrTimeout = errors.New("timeout")

	// ErrUnavailable is returned when a service is unavailable.
	ErrUnavailable = errors.New("service unavailable")

	// ErrInternal is returned for internal system errors.
	ErrInternal = errors.New("internal error")

	// ErrExpired is returned when a resource has expired.
	ErrExpired = errors.New("expired")

	// ErrRevoked is returned when a resource has been revoked.
	ErrRevoked = errors.New("revoked")

	// ErrLocked is returned when a resource is locked.
	ErrLocked = errors.New("locked")

	// ErrInvalidState is returned when a resource is in an invalid state.
	ErrInvalidState = errors.New("invalid state")

	// ErrInvalidSignature is returned when a signature is invalid.
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrVerificationFailed is returned when verification fails.
	ErrVerificationFailed = errors.New("verification failed")

	// ErrEncryptionFailed is returned when encryption fails.
	ErrEncryptionFailed = errors.New("encryption failed")

	// ErrDecryptionFailed is returned when decryption fails.
	ErrDecryptionFailed = errors.New("decryption failed")

	// ErrRateLimitExceeded is returned when rate limit is exceeded.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrQuotaExceeded is returned when quota is exceeded.
	ErrQuotaExceeded = errors.New("quota exceeded")

	// ErrMaxRetriesExceeded is returned when max retries are exceeded.
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")

	// ErrCanceled is returned when an operation is canceled.
	ErrCanceled = errors.New("canceled")

	// ErrDeadlineExceeded is returned when a deadline is exceeded.
	ErrDeadlineExceeded = errors.New("deadline exceeded")

	// ErrNotImplemented is returned when a feature is not implemented.
	ErrNotImplemented = errors.New("not implemented")

	// ErrDeprecated is returned when a feature is deprecated.
	ErrDeprecated = errors.New("deprecated")

	// ErrConflict is returned when there is a resource conflict.
	ErrConflict = errors.New("conflict")

	// ErrPreconditionFailed is returned when a precondition is not met.
	ErrPreconditionFailed = errors.New("precondition failed")

	// ErrUnsupported is returned when an operation is unsupported.
	ErrUnsupported = errors.New("unsupported")
)

// Validation errors
var (
	// ErrInvalidParams is returned when parameters are invalid.
	ErrInvalidParams = errors.New("invalid parameters")

	// ErrMissingRequired is returned when a required field is missing.
	ErrMissingRequired = errors.New("missing required field")

	// ErrOutOfRange is returned when a value is out of range.
	ErrOutOfRange = errors.New("value out of range")

	// ErrInvalidFormat is returned when format is invalid.
	ErrInvalidFormat = errors.New("invalid format")

	// ErrInvalidLength is returned when length is invalid.
	ErrInvalidLength = errors.New("invalid length")
)

// External service errors
var (
	// ErrExternalService is returned when an external service fails.
	ErrExternalService = errors.New("external service error")

	// ErrNetworkError is returned for network errors.
	ErrNetworkError = errors.New("network error")

	// ErrConnectionFailed is returned when connection fails.
	ErrConnectionFailed = errors.New("connection failed")

	// ErrRequestFailed is returned when a request fails.
	ErrRequestFailed = errors.New("request failed")
)

// Consensus and blockchain errors
var (
	// ErrConsensus is returned for consensus errors.
	ErrConsensus = errors.New("consensus error")

	// ErrInvalidTransaction is returned when a transaction is invalid.
	ErrInvalidTransaction = errors.New("invalid transaction")

	// ErrInsufficientFunds is returned when funds are insufficient.
	ErrInsufficientFunds = errors.New("insufficient funds")

	// ErrNonceInvalid is returned when nonce is invalid.
	ErrNonceInvalid = errors.New("invalid nonce")
)

// ML and verification errors
var (
	// ErrInferenceFailed is returned when ML inference fails.
	ErrInferenceFailed = errors.New("inference failed")

	// ErrModelNotFound is returned when ML model is not found.
	ErrModelNotFound = errors.New("model not found")

	// ErrModelVersionMismatch is returned when model versions don't match.
	ErrModelVersionMismatch = errors.New("model version mismatch")

	// ErrScoreToleranceExceeded is returned when score tolerance is exceeded.
	ErrScoreToleranceExceeded = errors.New("score tolerance exceeded")
)

