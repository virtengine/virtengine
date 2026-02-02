package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

const (
	fieldUsername = "username"
)

func TestCodedError(t *testing.T) {
	t.Run("basic error", func(t *testing.T) {
		err := NewCodedError("test", 100, "test error", CategoryValidation)
		if err.Error() != "test:100: test error" {
			t.Errorf("unexpected error message: %s", err.Error())
		}
	})

	t.Run("with context", func(t *testing.T) {
		err := NewCodedError("test", 100, "test error", CategoryValidation)
		_ = err.WithContext("field", fieldUsername)
		_ = err.WithContext("value", "invalid")

		if err.Context["field"] != fieldUsername {
			t.Error("context not set")
		}
	})

	t.Run("error comparison", func(t *testing.T) {
		err1 := NewCodedError("test", 100, "test error", CategoryValidation)
		err2 := NewCodedError("test", 100, "different message", CategoryValidation)
		err3 := NewCodedError("test", 101, "test error", CategoryValidation)

		if !errors.Is(err1, err2) {
			t.Error("errors with same module and code should match")
		}

		if errors.Is(err1, err3) {
			t.Error("errors with different codes should not match")
		}
	})
}

func TestValidationError(t *testing.T) {
	err := NewValidationError("test", 100, fieldUsername, "must be alphanumeric")

	if !strings.Contains(err.Error(), fieldUsername) {
		t.Error("error should contain field name")
	}

	if err.Field != fieldUsername {
		t.Error("field not set")
	}

	if err.Category != CategoryValidation {
		t.Error("wrong category")
	}
}

func TestNotFoundError(t *testing.T) {
	err := NewNotFoundError("test", 110, "user", "user123")

	if !strings.Contains(err.Error(), "user123") {
		t.Error("error should contain resource ID")
	}

	if err.ResourceType != "user" {
		t.Error("resource type not set")
	}

	if err.Category != CategoryNotFound {
		t.Error("wrong category")
	}
}

func TestConflictError(t *testing.T) {
	err := NewConflictError("test", 120, "user", "user123", "")

	if !strings.Contains(err.Error(), "already exists") {
		t.Error("error should contain 'already exists'")
	}

	if err.Category != CategoryConflict {
		t.Error("wrong category")
	}
}

func TestUnauthorizedError(t *testing.T) {
	err := NewUnauthorizedError("test", 130, "delete_resource", "insufficient permissions")

	if !strings.Contains(err.Error(), "unauthorized") {
		t.Error("error should contain 'unauthorized'")
	}

	if err.Action != "delete_resource" {
		t.Error("action not set")
	}

	if err.Category != CategoryUnauthorized {
		t.Error("wrong category")
	}
}

func TestTimeoutError(t *testing.T) {
	err := NewTimeoutError("test", 151, "database query", "30s")

	if !strings.Contains(err.Error(), "timeout") {
		t.Error("error should contain 'timeout'")
	}

	if !err.Retryable {
		t.Error("timeout errors should be retryable")
	}

	if err.Category != CategoryTimeout {
		t.Error("wrong category")
	}
}

func TestInternalError(t *testing.T) {
	err := NewInternalError("test", 160, "database", "connection pool exhausted")

	if err.Severity != SeverityCritical {
		t.Error("internal errors should have critical severity")
	}

	if err.Category != CategoryInternal {
		t.Error("wrong category")
	}
}

func TestExternalError(t *testing.T) {
	err := NewExternalError("test", 150, "stripe", "create_payment", "API unavailable")

	if !err.Retryable {
		t.Error("external errors should be retryable")
	}

	if err.Service != "stripe" {
		t.Error("service not set")
	}

	if err.Category != CategoryExternal {
		t.Error("wrong category")
	}
}

func TestRateLimitError(t *testing.T) {
	err := NewRateLimitError("test", 180, 100, "2024-01-01T00:00:00Z")

	if !strings.Contains(err.Error(), "rate limit") {
		t.Error("error should contain 'rate limit'")
	}

	if !err.Retryable {
		t.Error("rate limit errors should be retryable")
	}

	if err.Limit != 100 {
		t.Error("limit not set")
	}

	if err.Category != CategoryRateLimit {
		t.Error("wrong category")
	}
}

func TestGetCategory(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorCategory
	}{
		{"validation error", NewValidationError("test", 100, "field", "message"), CategoryValidation},
		{"not found error", NewNotFoundError("test", 110, "user", "123"), CategoryNotFound},
		{"standard error", errors.New("standard error"), CategoryInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat := GetCategory(tt.err)
			if cat != tt.expected {
				t.Errorf("expected category %s, got %s", tt.expected, cat)
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"timeout error", NewTimeoutError("test", 151, "op", "30s"), true},
		{"external error", NewExternalError("test", 150, "api", "op", "msg"), true},
		{"validation error", NewValidationError("test", 100, "field", "msg"), false},
		{"standard error", errors.New("standard error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retryable := IsRetryable(tt.err)
			if retryable != tt.expected {
				t.Errorf("expected retryable %v, got %v", tt.expected, retryable)
			}
		})
	}
}

func TestGetCode(t *testing.T) {
	err := NewCodedError("test", 100, "message", CategoryValidation)
	module, code := GetCode(err)

	if module != "test" {
		t.Errorf("expected module 'test', got '%s'", module)
	}

	if code != 100 {
		t.Errorf("expected code 100, got %d", code)
	}

	// Test with standard error
	module, code = GetCode(errors.New("standard error"))
	if module != "" || code != 0 {
		t.Error("standard error should return empty module and 0 code")
	}
}

func TestWrap(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrap(original, "additional context")

	if !errors.Is(wrapped, original) {
		t.Error("wrapped error should unwrap to original")
	}

	if !strings.Contains(wrapped.Error(), "additional context") {
		t.Error("wrapped error should contain context")
	}

	if !strings.Contains(wrapped.Error(), "original error") {
		t.Error("wrapped error should contain original message")
	}
}

func TestWrapf(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrapf(original, "context: %s", "formatted")

	if !errors.Is(wrapped, original) {
		t.Error("wrapped error should unwrap to original")
	}

	if !strings.Contains(wrapped.Error(), "formatted") {
		t.Error("wrapped error should contain formatted context")
	}
}

func TestWrapCoded(t *testing.T) {
	original := errors.New("original error")
	wrapped := WrapCoded(original, "test", 100, "coded message", CategoryInternal)

	var codedErr *CodedError
	if !errors.As(wrapped, &codedErr) {
		t.Fatal("wrapped error should be CodedError")
	}

	if codedErr.Module != "test" || codedErr.Code != 100 {
		t.Error("coded error attributes not set")
	}

	if !errors.Is(wrapped, original) {
		t.Error("wrapped error should preserve original as cause")
	}
}

func TestCause(t *testing.T) {
	root := errors.New("root cause")
	wrapped1 := fmt.Errorf("wrap1: %w", root)
	wrapped2 := fmt.Errorf("wrap2: %w", wrapped1)

	cause := Cause(wrapped2)
	if cause != root {
		t.Error("Cause should return root error")
	}
}

func TestWithField(t *testing.T) {
	err := NewCodedError("test", 100, "message", CategoryValidation)
	err = WithField(err, fieldUsername, "testuser").(*CodedError)

	if err.Context["field"] != fieldUsername {
		t.Error("field not added to context")
	}

	if err.Context["value"] != "testuser" {
		t.Error("value not added to context")
	}
}

func TestWithOperation(t *testing.T) {
	err := NewCodedError("test", 100, "message", CategoryInternal)
	err = WithOperation(err, "database.query").(*CodedError)

	if err.Context["operation"] != "database.query" {
		t.Error("operation not added to context")
	}
}

func TestWithResource(t *testing.T) {
	err := NewCodedError("test", 100, "message", CategoryNotFound)
	err = WithResource(err, "user", "user123").(*CodedError)

	if err.Context["resource_type"] != "user" {
		t.Error("resource_type not added to context")
	}

	if err.Context["resource_id"] != "user123" {
		t.Error("resource_id not added to context")
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name     string
		sentinel error
	}{
		{"ErrNotFound", ErrNotFound},
		{"ErrAlreadyExists", ErrAlreadyExists},
		{"ErrInvalidInput", ErrInvalidInput},
		{"ErrUnauthorized", ErrUnauthorized},
		{"ErrTimeout", ErrTimeout},
		{"ErrInternal", ErrInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.sentinel == nil {
				t.Error("sentinel error is nil")
			}

			// Test that wrapped sentinel can be identified
			wrapped := fmt.Errorf("context: %w", tt.sentinel)
			if !errors.Is(wrapped, tt.sentinel) {
				t.Error("wrapped sentinel not identified")
			}
		})
	}
}

func TestModuleCodeRanges(t *testing.T) {
	// Test that all modules have valid ranges
	for _, r := range AllModuleRanges {
		if r.StartCode >= r.EndCode {
			t.Errorf("invalid range for %s: %d-%d", r.Module, r.StartCode, r.EndCode)
		}

		if r.EndCode-r.StartCode != 99 {
			t.Errorf("module %s does not have 100 codes allocated", r.Module)
		}
	}

	// Test GetModuleRange
	r, ok := GetModuleRange("veid")
	if !ok {
		t.Error("veid module not found")
	}
	if r.StartCode != 1000 || r.EndCode != 1099 {
		t.Errorf("unexpected range for veid: %d-%d", r.StartCode, r.EndCode)
	}

	// Test non-existent module
	_, ok = GetModuleRange("nonexistent")
	if ok {
		t.Error("nonexistent module should not be found")
	}
}

func TestValidateCode(t *testing.T) {
	tests := []struct {
		name     string
		module   string
		code     uint32
		expected bool
	}{
		{"valid veid code", "veid", 1050, true},
		{"veid start boundary", "veid", 1000, true},
		{"veid end boundary", "veid", 1099, true},
		{"veid out of range low", "veid", 999, false},
		{"veid out of range high", "veid", 1100, false},
		{"valid mfa code", "mfa", 1250, true},
		{"invalid module", "nonexistent", 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := ValidateCode(tt.module, tt.code)
			if valid != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, valid)
			}
		})
	}
}
