package nli

import (
	"errors"
)

// Service errors
var (
	// ErrEmptyMessage is returned when the message is empty
	ErrEmptyMessage = errors.New("nli: message cannot be empty")

	// ErrMissingSessionID is returned when session ID is missing
	ErrMissingSessionID = errors.New("nli: session ID is required")

	// ErrServiceClosed is returned when the service has been closed
	ErrServiceClosed = errors.New("nli: service has been closed")

	// ErrInvalidConfig is returned when configuration is invalid
	ErrInvalidConfig = errors.New("nli: invalid configuration")

	// ErrContextDeadlineExceeded is returned when the context deadline is exceeded
	ErrContextDeadlineExceeded = errors.New("nli: context deadline exceeded")

	// ErrRateLimited is returned when rate limit is exceeded
	ErrRateLimited = errors.New("nli: rate limit exceeded")
)

// LLM backend errors
var (
	// ErrLLMBackendNotConfigured is returned when LLM backend is not configured
	ErrLLMBackendNotConfigured = errors.New("nli: LLM backend not configured")

	// ErrLLMBackendUnavailable is returned when LLM backend is unavailable
	ErrLLMBackendUnavailable = errors.New("nli: LLM backend unavailable")

	// ErrLLMCompletionFailed is returned when LLM completion fails
	ErrLLMCompletionFailed = errors.New("nli: LLM completion failed")

	// ErrInvalidLLMResponse is returned when LLM response is invalid
	ErrInvalidLLMResponse = errors.New("nli: invalid LLM response")
)

// Query errors
var (
	// ErrQueryFailed is returned when a blockchain query fails
	ErrQueryFailed = errors.New("nli: query failed")

	// ErrAddressNotProvided is returned when address is required but not provided
	ErrAddressNotProvided = errors.New("nli: user address not provided")

	// ErrInvalidAddress is returned when the provided address is invalid
	ErrInvalidAddress = errors.New("nli: invalid address format")

	// ErrQueryExecutorNotConfigured is returned when query executor is not set
	ErrQueryExecutorNotConfigured = errors.New("nli: query executor not configured")
)

// Classification errors
var (
	// ErrClassificationFailed is returned when intent classification fails
	ErrClassificationFailed = errors.New("nli: intent classification failed")

	// ErrLowConfidence is returned when classification confidence is too low
	ErrLowConfidence = errors.New("nli: classification confidence too low")
)

// Session errors
var (
	// ErrSessionNotFound is returned when a session is not found
	ErrSessionNotFound = errors.New("nli: session not found")

	// ErrSessionExpired is returned when a session has expired
	ErrSessionExpired = errors.New("nli: session expired")

	// ErrSessionStoreClosed is returned when the session store is closed
	ErrSessionStoreClosed = errors.New("nli: session store closed")
)
