package errors

import (
	"errors"
	"fmt"
)

// ErrorCategory defines the category of an error.
type ErrorCategory string

const (
	// CategoryValidation indicates input validation errors.
	CategoryValidation ErrorCategory = "validation"

	// CategoryNotFound indicates resource not found errors.
	CategoryNotFound ErrorCategory = "not_found"

	// CategoryConflict indicates resource conflict errors.
	CategoryConflict ErrorCategory = "conflict"

	// CategoryUnauthorized indicates authorization errors.
	CategoryUnauthorized ErrorCategory = "unauthorized"

	// CategoryState indicates invalid state or lifecycle errors.
	CategoryState ErrorCategory = "state"

	// CategoryExternal indicates external service errors.
	CategoryExternal ErrorCategory = "external"

	// CategoryInternal indicates internal system errors.
	CategoryInternal ErrorCategory = "internal"

	// CategoryTimeout indicates timeout errors.
	CategoryTimeout ErrorCategory = "timeout"

	// CategoryRateLimit indicates rate limiting errors.
	CategoryRateLimit ErrorCategory = "rate_limit"
)

// ErrorSeverity defines the severity of an error.
type ErrorSeverity string

const (
	// SeverityInfo indicates informational errors (e.g., not found).
	SeverityInfo ErrorSeverity = "info"

	// SeverityWarning indicates warning-level errors (e.g., deprecation).
	SeverityWarning ErrorSeverity = "warning"

	// SeverityError indicates standard errors.
	SeverityError ErrorSeverity = "error"

	// SeverityCritical indicates critical system errors.
	SeverityCritical ErrorSeverity = "critical"
)

// CodedError is an error with an error code, module, and metadata.
type CodedError struct {
	Module    string
	Code      uint32
	Message   string
	Category  ErrorCategory
	Severity  ErrorSeverity
	Retryable bool
	Cause     error
	Context   map[string]interface{}
}

// Error implements the error interface.
func (e *CodedError) Error() string {
	if e.Module != "" {
		return fmt.Sprintf("%s:%d: %s", e.Module, e.Code, e.Message)
	}
	return e.Message
}

// Unwrap returns the underlying cause.
func (e *CodedError) Unwrap() error {
	return e.Cause
}

// Is implements error comparison for errors.Is().
func (e *CodedError) Is(target error) bool {
	t, ok := target.(*CodedError)
	if !ok {
		return false
	}
	return e.Module == t.Module && e.Code == t.Code
}

// WithContext adds context key-value pairs to the error.
func (e *CodedError) WithContext(key string, value interface{}) *CodedError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// NewCodedError creates a new coded error.
func NewCodedError(module string, code uint32, message string, category ErrorCategory) *CodedError {
	return &CodedError{
		Module:    module,
		Code:      code,
		Message:   message,
		Category:  category,
		Severity:  SeverityError,
		Retryable: false,
		Context:   make(map[string]interface{}),
	}
}

// ValidationError represents an input validation error.
type ValidationError struct {
	*CodedError
	Field string
	Value interface{}
}

// NewValidationError creates a new validation error.
func NewValidationError(module string, code uint32, field string, message string) *ValidationError {
	return &ValidationError{
		CodedError: NewCodedError(module, code, message, CategoryValidation),
		Field:      field,
	}
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s:%d: validation error on field '%s': %s", e.Module, e.Code, e.Field, e.Message)
	}
	return e.CodedError.Error()
}

// NotFoundError represents a resource not found error.
type NotFoundError struct {
	*CodedError
	ResourceType string
	ResourceID   string
}

// NewNotFoundError creates a new not found error.
func NewNotFoundError(module string, code uint32, resourceType string, resourceID string) *NotFoundError {
	message := fmt.Sprintf("%s not found: %s", resourceType, resourceID)
	return &NotFoundError{
		CodedError:   NewCodedError(module, code, message, CategoryNotFound),
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}
}

// ConflictError represents a resource conflict error.
type ConflictError struct {
	*CodedError
	ResourceType string
	ResourceID   string
}

// NewConflictError creates a new conflict error.
func NewConflictError(module string, code uint32, resourceType string, resourceID string, message string) *ConflictError {
	if message == "" {
		message = fmt.Sprintf("%s already exists: %s", resourceType, resourceID)
	}
	return &ConflictError{
		CodedError:   NewCodedError(module, code, message, CategoryConflict),
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}
}

// UnauthorizedError represents an authorization error.
type UnauthorizedError struct {
	*CodedError
	Action string
	Reason string
}

// NewUnauthorizedError creates a new unauthorized error.
func NewUnauthorizedError(module string, code uint32, action string, reason string) *UnauthorizedError {
	message := fmt.Sprintf("unauthorized: %s", reason)
	return &UnauthorizedError{
		CodedError: NewCodedError(module, code, message, CategoryUnauthorized),
		Action:     action,
		Reason:     reason,
	}
}

// TimeoutError represents a timeout error.
type TimeoutError struct {
	*CodedError
	Operation string
	Duration  string
}

// NewTimeoutError creates a new timeout error.
func NewTimeoutError(module string, code uint32, operation string, duration string) *TimeoutError {
	message := fmt.Sprintf("timeout after %s: %s", duration, operation)
	err := NewCodedError(module, code, message, CategoryTimeout)
	err.Retryable = true
	return &TimeoutError{
		CodedError: err,
		Operation:  operation,
		Duration:   duration,
	}
}

// InternalError represents an internal system error.
type InternalError struct {
	*CodedError
	Component string
}

// NewInternalError creates a new internal error.
func NewInternalError(module string, code uint32, component string, message string) *InternalError {
	err := NewCodedError(module, code, message, CategoryInternal)
	err.Severity = SeverityCritical
	return &InternalError{
		CodedError: err,
		Component:  component,
	}
}

// ExternalError represents an external service error.
type ExternalError struct {
	*CodedError
	Service   string
	Operation string
}

// NewExternalError creates a new external error.
func NewExternalError(module string, code uint32, service string, operation string, message string) *ExternalError {
	err := NewCodedError(module, code, message, CategoryExternal)
	err.Retryable = true
	return &ExternalError{
		CodedError: err,
		Service:    service,
		Operation:  operation,
	}
}

// RateLimitError represents a rate limit error.
type RateLimitError struct {
	*CodedError
	Limit     int
	Remaining int
	ResetAt   string
}

// NewRateLimitError creates a new rate limit error.
func NewRateLimitError(module string, code uint32, limit int, resetAt string) *RateLimitError {
	message := fmt.Sprintf("rate limit exceeded: %d requests allowed, resets at %s", limit, resetAt)
	err := NewCodedError(module, code, message, CategoryRateLimit)
	err.Retryable = true
	return &RateLimitError{
		CodedError: err,
		Limit:      limit,
		Remaining:  0,
		ResetAt:    resetAt,
	}
}

// GetCategory returns the category of an error.
func GetCategory(err error) ErrorCategory {
	// Check specialized types first (they embed CodedError)
	var valErr *ValidationError
	if errors.As(err, &valErr) {
		return valErr.Category
	}

	var nfErr *NotFoundError
	if errors.As(err, &nfErr) {
		return nfErr.Category
	}

	var conflictErr *ConflictError
	if errors.As(err, &conflictErr) {
		return conflictErr.Category
	}

	var unauthErr *UnauthorizedError
	if errors.As(err, &unauthErr) {
		return unauthErr.Category
	}

	var timeoutErr *TimeoutError
	if errors.As(err, &timeoutErr) {
		return timeoutErr.Category
	}

	var intErr *InternalError
	if errors.As(err, &intErr) {
		return intErr.Category
	}

	var extErr *ExternalError
	if errors.As(err, &extErr) {
		return extErr.Category
	}

	var rlErr *RateLimitError
	if errors.As(err, &rlErr) {
		return rlErr.Category
	}

	// Check base CodedError
	var codedErr *CodedError
	if errors.As(err, &codedErr) {
		return codedErr.Category
	}
	return CategoryInternal
}

// IsRetryable returns true if the error is retryable.
func IsRetryable(err error) bool {
	// Check specialized types first
	var valErr *ValidationError
	if errors.As(err, &valErr) {
		return valErr.Retryable
	}

	var nfErr *NotFoundError
	if errors.As(err, &nfErr) {
		return nfErr.Retryable
	}

	var conflictErr *ConflictError
	if errors.As(err, &conflictErr) {
		return conflictErr.Retryable
	}

	var unauthErr *UnauthorizedError
	if errors.As(err, &unauthErr) {
		return unauthErr.Retryable
	}

	var timeoutErr *TimeoutError
	if errors.As(err, &timeoutErr) {
		return timeoutErr.Retryable
	}

	var intErr *InternalError
	if errors.As(err, &intErr) {
		return intErr.Retryable
	}

	var extErr *ExternalError
	if errors.As(err, &extErr) {
		return extErr.Retryable
	}

	var rlErr *RateLimitError
	if errors.As(err, &rlErr) {
		return rlErr.Retryable
	}

	// Check base CodedError
	var codedErr *CodedError
	if errors.As(err, &codedErr) {
		return codedErr.Retryable
	}
	return false
}

// GetCode returns the error code of an error.
func GetCode(err error) (string, uint32) {
	// Check specialized types first
	var valErr *ValidationError
	if errors.As(err, &valErr) {
		return valErr.Module, valErr.Code
	}

	var nfErr *NotFoundError
	if errors.As(err, &nfErr) {
		return nfErr.Module, nfErr.Code
	}

	var conflictErr *ConflictError
	if errors.As(err, &conflictErr) {
		return conflictErr.Module, conflictErr.Code
	}

	var unauthErr *UnauthorizedError
	if errors.As(err, &unauthErr) {
		return unauthErr.Module, unauthErr.Code
	}

	var timeoutErr *TimeoutError
	if errors.As(err, &timeoutErr) {
		return timeoutErr.Module, timeoutErr.Code
	}

	var intErr *InternalError
	if errors.As(err, &intErr) {
		return intErr.Module, intErr.Code
	}

	var extErr *ExternalError
	if errors.As(err, &extErr) {
		return extErr.Module, extErr.Code
	}

	var rlErr *RateLimitError
	if errors.As(err, &rlErr) {
		return rlErr.Module, rlErr.Code
	}

	// Check base CodedError
	var codedErr *CodedError
	if errors.As(err, &codedErr) {
		return codedErr.Module, codedErr.Code
	}
	return "", 0
}
