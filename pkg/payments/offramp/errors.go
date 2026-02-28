package offramp

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrAdapterUnavailable   = errors.New("offramp adapter unavailable")
	ErrQuoteExpired         = errors.New("offramp quote expired")
	ErrInvalidRequest       = errors.New("offramp request invalid")
	ErrPayoutNotFound       = errors.New("offramp payout not found")
	ErrPayoutNotCancellable = errors.New("offramp payout not cancellable")
	ErrProviderRejected     = errors.New("offramp provider rejected request")
	ErrProviderTemporary    = errors.New("offramp provider temporary failure")
	ErrProviderAmbiguous    = errors.New("offramp provider result ambiguous")
)

// ErrorKind classifies normalized provider failures.
type ErrorKind string

const (
	ErrorKindUnknown        ErrorKind = "unknown"
	ErrorKindInvalidRequest ErrorKind = "invalid_request"
	ErrorKindNotFound       ErrorKind = "not_found"
	ErrorKindUnavailable    ErrorKind = "unavailable"
	ErrorKindTemporary      ErrorKind = "temporary"
	ErrorKindTimeout        ErrorKind = "timeout"
	ErrorKindRateLimited    ErrorKind = "rate_limited"
	ErrorKindRejected       ErrorKind = "rejected"
	ErrorKindAmbiguous      ErrorKind = "ambiguous"
	ErrorKindConflict       ErrorKind = "conflict"
)

// ErrorNormalizer optionally lets adapters normalize partner-specific failures.
type ErrorNormalizer interface {
	NormalizeError(operation string, err error) error
}

// ProviderError describes a normalized partner failure.
type ProviderError struct {
	Provider  string
	Operation string
	Code      string
	Kind      ErrorKind
	Retryable bool
	Ambiguous bool
	err       error
}

func (e *ProviderError) Error() string {
	parts := make([]string, 0, 4)
	if e.Provider != "" {
		parts = append(parts, e.Provider)
	}
	if e.Operation != "" {
		parts = append(parts, e.Operation)
	}
	if e.Kind != "" {
		parts = append(parts, string(e.Kind))
	}
	if e.Code != "" {
		parts = append(parts, e.Code)
	}
	prefix := strings.Join(parts, " ")
	if prefix == "" {
		prefix = "offramp provider error"
	}
	if e.err == nil {
		return prefix
	}
	return fmt.Sprintf("%s: %v", prefix, e.err)
}

func (e *ProviderError) Unwrap() error {
	return e.err
}

// NormalizeError maps adapter or transport failures into a single bridge contract.
func NormalizeError(provider string, operation string, err error) error {
	if err == nil {
		return nil
	}

	var providerErr *ProviderError
	if errors.As(err, &providerErr) {
		normalized := *providerErr
		if normalized.Provider == "" {
			normalized.Provider = provider
		}
		if normalized.Operation == "" {
			normalized.Operation = operation
		}
		if normalized.Kind == "" {
			normalized.Kind = ErrorKindUnknown
		}
		if normalized.err == nil {
			normalized.err = baseError(normalized.Kind, normalized.Ambiguous)
		}
		return &normalized
	}

	kind, retryable, ambiguous := classifyError(operation, err)
	return &ProviderError{
		Provider:  provider,
		Operation: operation,
		Kind:      kind,
		Retryable: retryable,
		Ambiguous: ambiguous,
		err:       wrapBaseError(baseError(kind, ambiguous), err),
	}
}

// IsRetryable returns true when a normalized provider failure can be retried.
func IsRetryable(err error) bool {
	var providerErr *ProviderError
	return errors.As(err, &providerErr) && providerErr.Retryable
}

// IsAmbiguous returns true when the provider outcome is unknown and must be reconciled.
func IsAmbiguous(err error) bool {
	var providerErr *ProviderError
	return errors.As(err, &providerErr) && providerErr.Ambiguous
}

// IsNotFound returns true when the normalized provider failure means the payout does not exist.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrPayoutNotFound)
}

// CanFailover returns true when the bridge may safely attempt a different provider.
func CanFailover(err error) bool {
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		return false
	}
	if providerErr.Ambiguous || !providerErr.Retryable {
		return false
	}
	switch providerErr.Kind {
	case ErrorKindInvalidRequest, ErrorKindRejected, ErrorKindConflict:
		return false
	default:
		return true
	}
}

func classifyError(operation string, err error) (ErrorKind, bool, bool) {
	switch {
	case errors.Is(err, ErrInvalidRequest):
		return ErrorKindInvalidRequest, false, false
	case errors.Is(err, ErrPayoutNotFound):
		return ErrorKindNotFound, false, false
	case errors.Is(err, ErrAdapterUnavailable):
		return ErrorKindUnavailable, true, false
	case errors.Is(err, ErrProviderRejected):
		return ErrorKindRejected, false, false
	case errors.Is(err, ErrProviderTemporary):
		return ErrorKindTemporary, true, false
	case errors.Is(err, ErrProviderAmbiguous):
		return ErrorKindAmbiguous, true, true
	case errors.Is(err, context.DeadlineExceeded):
		return ErrorKindTimeout, true, operation == operationInitiatePayout || operation == operationCancel
	}

	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "deadline exceeded"),
		strings.Contains(message, "timeout"),
		strings.Contains(message, "timed out"):
		return ErrorKindTimeout, true, operation == operationInitiatePayout || operation == operationCancel
	case strings.Contains(message, "rate limit"),
		strings.Contains(message, "too many requests"),
		strings.Contains(message, "429"):
		return ErrorKindRateLimited, true, false
	case strings.Contains(message, "unavailable"),
		strings.Contains(message, "connection reset"),
		strings.Contains(message, "connection refused"),
		strings.Contains(message, "transport"),
		strings.Contains(message, "temporary"),
		strings.Contains(message, "eof"),
		strings.Contains(message, "network"):
		return ErrorKindUnavailable, true, false
	case strings.Contains(message, "idempot"),
		strings.Contains(message, "already exists"),
		strings.Contains(message, "conflict"),
		strings.Contains(message, "duplicate"):
		return ErrorKindConflict, false, true
	case strings.Contains(message, "not found"),
		strings.Contains(message, "unknown payout"),
		strings.Contains(message, "missing payout"):
		return ErrorKindNotFound, false, false
	case strings.Contains(message, "invalid"),
		strings.Contains(message, "unsupported"),
		strings.Contains(message, "malformed"),
		strings.Contains(message, "missing"):
		return ErrorKindInvalidRequest, false, false
	case strings.Contains(message, "reject"),
		strings.Contains(message, "declin"),
		strings.Contains(message, "denied"),
		strings.Contains(message, "forbidden"),
		strings.Contains(message, "compliance"):
		return ErrorKindRejected, false, false
	default:
		return ErrorKindUnknown, false, false
	}
}

func baseError(kind ErrorKind, ambiguous bool) error {
	if ambiguous {
		return ErrProviderAmbiguous
	}
	switch kind {
	case ErrorKindInvalidRequest:
		return ErrInvalidRequest
	case ErrorKindNotFound:
		return ErrPayoutNotFound
	case ErrorKindUnavailable, ErrorKindTemporary, ErrorKindTimeout, ErrorKindRateLimited:
		return ErrProviderTemporary
	case ErrorKindRejected:
		return ErrProviderRejected
	default:
		return ErrAdapterUnavailable
	}
}

func wrapBaseError(base error, err error) error {
	if err == nil {
		return base
	}
	if base == nil || errors.Is(err, base) {
		return err
	}
	return fmt.Errorf("%w: %v", base, err)
}
