package errors

import (
	"errors"
	"fmt"
)

// Wrap wraps an error with additional context message.
// It preserves the original error for unwrapping with errors.Unwrap().
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf wraps an error with a formatted context message.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// WrapCoded wraps an error with a CodedError, preserving the original error as cause.
func WrapCoded(err error, module string, code uint32, message string, category ErrorCategory) error {
	if err == nil {
		return nil
	}
	return &CodedError{
		Module:   module,
		Code:     code,
		Message:  message,
		Category: category,
		Severity: SeverityError,
		Cause:    err,
		Context:  make(map[string]interface{}),
	}
}

// WrapWithContext wraps an error with a CodedError and additional context.
func WrapWithContext(err error, module string, code uint32, message string, category ErrorCategory, context map[string]interface{}) error {
	if err == nil {
		return nil
	}
	return &CodedError{
		Module:   module,
		Code:     code,
		Message:  message,
		Category: category,
		Severity: SeverityError,
		Cause:    err,
		Context:  context,
	}
}

// Join combines multiple errors into a single error.
// Uses Go 1.20+ errors.Join if available.
func Join(errs ...error) error {
	return errors.Join(errs...)
}

// Cause returns the root cause of an error by unwrapping recursively.
func Cause(err error) error {
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

// AddContext adds context to a CodedError. If the error is not a CodedError,
// it wraps it in one.
func AddContext(err error, module string, code uint32, key string, value interface{}) error {
	if err == nil {
		return nil
	}

	var codedErr *CodedError
	if errors.As(err, &codedErr) {
		codedErr.WithContext(key, value)
		return err
	}

	// Wrap in CodedError
	newErr := &CodedError{
		Module:   module,
		Code:     code,
		Message:  err.Error(),
		Category: CategoryInternal,
		Severity: SeverityError,
		Cause:    err,
		Context:  map[string]interface{}{key: value},
	}
	return newErr
}

// WithField is a convenience function for adding a field context to an error.
func WithField(err error, field string, value interface{}) error {
	if err == nil {
		return nil
	}

	var codedErr *CodedError
	if errors.As(err, &codedErr) {
		codedErr.WithContext("field", field)
		if value != nil {
			codedErr.WithContext("value", value)
		}
		return err
	}

	return err
}

// WithOperation adds operation context to an error.
func WithOperation(err error, operation string) error {
	if err == nil {
		return nil
	}

	var codedErr *CodedError
	if errors.As(err, &codedErr) {
		codedErr.WithContext("operation", operation)
		return err
	}

	return Wrapf(err, "operation: %s", operation)
}

// WithResource adds resource context to an error.
func WithResource(err error, resourceType string, resourceID string) error {
	if err == nil {
		return nil
	}

	var codedErr *CodedError
	if errors.As(err, &codedErr) {
		codedErr.WithContext("resource_type", resourceType)
		codedErr.WithContext("resource_id", resourceID)
		return err
	}

	return Wrapf(err, "resource: %s/%s", resourceType, resourceID)
}

// EnsureStack ensures an error has stack trace information.
// This is a placeholder for potential stack trace integration.
func EnsureStack(err error) error {
	// Future: integrate with pkg/errors or similar for stack traces
	return err
}
